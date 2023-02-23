package service

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"github.com/pariz/gountries"
	"github.com/vibeitco/accounting-service/model"
	emailSvc "github.com/vibeitco/email-service/model"
	"github.com/vibeitco/go-utils/log"
	"github.com/vibeitco/go-utils/server"
	"github.com/vibeitco/service-definitions/go/common"
)

type handler struct {
	config       Config
	emailService emailSvc.EmailServiceClient
}
type AccountingResponse struct {
	ProcessId string `json:"processId"`
	Success   bool   `json:"success"`
}

func NewHandler(conf Config, emailService emailSvc.EmailServiceClient) (*handler, error) {
	h := &handler{
		config:       conf,
		emailService: emailService,
	}

	return h, nil
}

func getPaymentType(document model.Document, token string) []model.Payment {

	params := url.Values{}

	params.Set("filter[where][documentId]", document.ID)
	query := params.Encode()

	url := fmt.Sprintf("https://api.spaceinvoices.com/v1/organizations/62da67215365227ab19d4fdd/payments?%s", query)

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", token)

	client := http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Error getting request", err)
		return nil
	}

	defer resp.Body.Close()

	var responseDocument []model.Payment
	json.NewDecoder(resp.Body).Decode(&responseDocument)

	return responseDocument
}

func getDocuments(skip int, from time.Time, to time.Time, token string) []model.Document {

	params := url.Values{}

	params.Set("filter[where][dateService][between][1]", to.Format(time.RFC3339))
	params.Set("filter[where][dateService][between][0]", from.Format(time.RFC3339))
	params.Set("filter[skip]", fmt.Sprint(skip))
	query := params.Encode()

	url := fmt.Sprintf("https://api.spaceinvoices.com/v1/organizations/62da67215365227ab19d4fdd/documents?%s", query)

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", token)

	client := http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Error getting request", err)
		return nil
	}

	defer resp.Body.Close()

	var responseDocument []model.Document
	err = json.NewDecoder(resp.Body).Decode(&responseDocument)

	fmt.Println("Current date and time is: ", responseDocument[0].Date)

	if len(responseDocument) == 100 {
		results := getDocuments(skip+100, from, to, token)
		responseDocument = append(responseDocument, results...)
	}

	return responseDocument
}

func (h *handler) GenerateAccountingExport(res http.ResponseWriter, req *http.Request) {

	go func() {
		var buf bytes.Buffer
		zipWriter := zip.NewWriter(&buf)

		layout := "2006-01-02T15:04:05.000Z"
		fromString := req.Header.Get("from")
		toString := req.Header.Get("to")
		email := req.Header.Get("email")

		from, err := time.Parse(layout, fromString)

		if err != nil {
			log.Error(req.Context(), err, nil, "Error parsing from date")
		}

		to, err := time.Parse(layout, toString)

		if err != nil {
			log.Error(req.Context(), err, nil, "Error parsing to date")
		}

		var responseDocument []model.Document = getDocuments(0, from, to, h.config.SpaceapiAuth)

		var xmlWg sync.WaitGroup
		var xmlMutex sync.Mutex
		xmlWg.Add(1)

		go func() {
			defer xmlWg.Done()
			knjizbe := createKnjizbe(responseDocument, h.config.SpaceapiAuth)
			xmlBytes := createXmlFile(knjizbe)

			xmlMutex.Lock()
			addXmlFileToZip(req.Context(), zipWriter, xmlBytes)
			defer xmlMutex.Unlock()
		}()

		sem := make(chan struct{}, 20)
		var wg sync.WaitGroup
		var mu sync.Mutex

		wg.Add(len(responseDocument))

		for _, document := range responseDocument {
			sem <- struct{}{}

			go func(document model.Document) {
				defer wg.Done()
				req, err := http.NewRequest("GET", document.Pdf, nil)

				if err != nil {
					log.Error(req.Context(), err, nil, "Error loading request")
					return
				}

				req.Header.Add("Authorization", h.config.SpaceapiAuth)

				client := http.Client{}
				resp, err := client.Do(req)

				if err != nil {
					fmt.Printf("Error downloading %s: %s\n", document.Number, err)
					return
				}
				defer resp.Body.Close()

				// Read the PDF file from the response body
				pdfData, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					fmt.Printf("Error reading %s: %s\n", document.Number, err)
					return
				}

				if len(pdfData) == 0 {
					fmt.Printf("Empty file detected for : %s: %s\n", document.Number, err)
				}

				zipFileHeader := &zip.FileHeader{
					Name:     document.Number + ".pdf",
					Method:   zip.Deflate,
					Modified: time.Now(),
				}

				mu.Lock()
				defer mu.Unlock()
				pdfFile, err := zipWriter.CreateHeader(zipFileHeader)

				if err != nil {
					fmt.Printf("Error creating PDF file in archive for %s: %s\n", document.Number, err)
					<-sem
					return
				}
				_, err = io.Copy(pdfFile, bytes.NewReader(pdfData))
				if err != nil {
					fmt.Printf("Error writing PDF file to archive for %s: %s\n", document.Number, err)
					<-sem
					return
				}
				<-sem
			}(document)

		}

		wg.Wait()
		xmlWg.Wait()
		// acquire the semaphore to ensure that all operations have completed
		for i := 0; i < cap(sem); i++ {
			sem <- struct{}{}
		}
		fmt.Println("All operations have completed")

		if responseDocument == nil {
			return
		}

		zipWriter.Close()
		zipData := buf.Bytes()

		fileName := newUniqueFileName(".zip")
		folderName := "vibeit-accounting"
		urlSlug := constructURLSlug(h.config, fileName)

		// upload to GCS
		ctx := req.Context()
		ctxGCS := context.Background()

		bw, err := newGCSBucketWriter(ctxGCS, folderName, urlSlug)
		if err != nil {
			log.Error(ctx, err, nil, "UploadFile bucketWriterError")
			return
		}
		defer bw.Close()
		if _, err := io.Copy(bw, bytes.NewReader(zipData)); err != nil {
			log.Error(ctx, err, nil, "UploadFile copyFileError")
			return
		}
		log.Info(ctxGCS, log.Data{"name": fileName}, "file uploaded to GCS")

		_, err = h.emailService.SendEmail(ctxGCS, &emailSvc.SendEmailRequest{
			Subject:  "Link to accounting export",
			Receiver: email,
			Sender:   "hi@vibeit.co",
			ReplyTo:  "support@vibeit.co",
			Language: common.Language_Slovene,
			Template: emailSvc.EmailTemplate_AccountingFile,
			Data: map[string]string{
				"FileUrl": "https://storage.googleapis.com/vibeit-accounting/" + urlSlug,
			},
		})

		if err != nil {
			log.Error(ctx, err, nil, "Mail sending failed")
			return
		}

		log.Info(ctxGCS, nil, "https://storage.googleapis.com/vibeit-accounting/"+urlSlug)
	}()

	payload := &AccountingResponse{
		ProcessId: "Test",
		Success:   true,
	}

	server.JSON(req.Context(), res, http.StatusOK, payload)
}

func createKnjizbe(documents []model.Document, auth string) []model.Knjizba {
	var knjizbe []model.Knjizba

	for _, document := range documents {
		payment := getPaymentType(document, auth)

		baseClaimsAccount := createBaseKnjizba(document)
		baseIncomeGoodsAccount := createBaseKnjizba(document)
		baseIncomeAndServicesAccount := createBaseKnjizba(document)
		baseOutgoingTaxAccountLower := createBaseKnjizba(document)
		baseOutgoingTaxAccountHigher := createBaseKnjizba(document)

		generateClaimsAccount(&baseClaimsAccount, document, false)
		generateIncomeGoodsAccount(&baseIncomeGoodsAccount, document)
		generateIncomeAndServicesAccount(&baseIncomeAndServicesAccount, document)

		knjizbe = append(knjizbe, baseClaimsAccount)
		knjizbe = append(knjizbe, baseIncomeGoodsAccount)
		knjizbe = append(knjizbe, baseIncomeAndServicesAccount)

		higherOutgoingTaxTotal, lowerOutgoingTaxTotal := calculateOutgoingTax(document)

		// Values for taxes and totals can be negative, so we only dont show those that are 0
		if higherOutgoingTaxTotal != 0 {
			generateHigherOutgoingTaxAccount(&baseOutgoingTaxAccountHigher, document, higherOutgoingTaxTotal)
			knjizbe = append(knjizbe, baseOutgoingTaxAccountHigher)
		}
		if lowerOutgoingTaxTotal != 0 {
			generateLowerOutgoingTaxAccount(&baseOutgoingTaxAccountLower, document, lowerOutgoingTaxTotal)
			knjizbe = append(knjizbe, baseOutgoingTaxAccountLower)
		}

		// Extra entry if paid with card or paypal, that does not contain taxes
		// Two entries are created, one almost idential copy to Claims without taxes, and one with a seperate value
		paymentType := payment[0].Type
		if paymentType == "card" || paymentType == "paypal" {
			baseClaimsCardAccount := createBaseKnjizba(document)
			generateClaimsAccount(&baseClaimsCardAccount, document, true)
			knjizbe = append(knjizbe, baseClaimsCardAccount)

			basePaymentAccount := createBaseKnjizba(document)
			if paymentType == "card" {
				createCardAccount(&basePaymentAccount, document)
			} else {
				createPaypalAccount(&basePaymentAccount, document)
			}
			knjizbe = append(knjizbe, basePaymentAccount)
		}
	}

	return knjizbe
}

func createXmlFile(knjizbe []model.Knjizba) []byte {
	var prenos model.Prenos
	prenos.Knjizba = append(prenos.Knjizba, knjizbe...)
	prenos.XsiType = xml.Attr{Name: xml.Name{Local: "xmlns:xsi"}, Value: "http://www.w3.org/2001/XMLSchema-instance"}
	prenos.NoNamespaceSchemaLocation = "http://www.vasco.si/xml/vasco-kn-21.xsd"
	byte, _ := xml.Marshal(prenos)
	return byte
}

func createBaseKnjizba(document model.Document) model.Knjizba {
	var knjizba model.Knjizba = model.Knjizba{

		Simbol:         "1",
		Dokument:       document.Number,
		Veza:           document.Number,
		DatumKnjizneja: document.Date.Format("2006-01-02"),
		Obracunsko_obdobje: model.ObracunskoObdobje{
			Mesec: int(document.Date.Month()),
			Leto:  document.Date.Year(),
		},
		DatumDokumenta: document.Date.Format("2006-01-02"),
		Opisdokumenta:  document.DocumentClient.Name,
		Attachment: model.Attachment{
			File: model.AttachementFile{
				XsiType: xml.Attr{Name: xml.Name{Local: "xmlns:xsi"}, Value: "xs:string"},
				XsiXs:   "http://www.w3.org/2001/XMLSchema",
				XsiXsi:  "http://www.w3.org/2001/XMLSchema-instance",
				Link:    document.Number + ".pdf",
			},
		},
	}
	return knjizba
}

func findDocumentCountry(document model.Document) (gountries.Country, error) {
	var country gountries.Country
	var err error
	query := gountries.New()

	if document.DocumentClient.CountryAlpha2Code != "" {
		country, err = query.FindCountryByAlpha(document.DocumentClient.CountryAlpha2Code)
	} else {
		country, err = query.FindCountryByName(document.DocumentClient.Country)
	}
	return country, err
}

func generateClaimsAccount(baseAccount *model.Knjizba, document model.Document, isCard bool) {

	var country gountries.Country
	var err error

	country, err = findDocumentCountry(document)

	konto := -1
	if err != nil {
		konto = 0
	} else {
		if country.Name.Common == "Slovenia" {
			konto = 1200
		} else {
			if country.EuMember {
				konto = 1210
			} else {
				konto = 1211
			}
		}
	}

	if isEndCustomer(document) {
		partner := model.Partner{
			Sifra:  "14",
			Naziv1: "Koncni kupec",
		}
		baseAccount.Partner = partner
	} else {
		partner := model.Partner{
			Sifra:                     document.DocumentClient.TaxNumber,
			Naziv1:                    document.DocumentClient.Name,
			Naslov:                    document.DocumentClient.Address,
			Posta:                     document.DocumentClient.City + " " + document.DocumentClient.Zip,
			Drzava:                    country.CCN3,
			Identifikacijska_stevilka: document.DocumentClient.TaxNumber,
			Davcna_stevilka:           document.DocumentClient.TaxNumber,
			Maticna_stevilka:          document.DocumentClient.CompanyNumber,
		}
		baseAccount.Partner = partner
	}

	baseAccount.Konto = konto

	var taxes []model.Ddv
	taxType := ""
	// Card has no taxes
	if !isCard {
		for _, tax := range document.DocumentTaxes {
			if tax.Tax > 9.5 {
				taxType = "ddv_visja_promet_znotraj_SLO"
			} else {
				taxType = "ddv_nizja_promet_znotraj_SLO"
			}

			tax := model.Ddv{
				VrstaDdv: taxType,
				Osnova:   tax.Base,
				Znesek:   tax.TotalTax,
			}

			taxes = append(taxes, tax)
		}
		ir := model.Ir{
			Ddv:        taxes,
			DatumZaDdv: document.Date.Format("2006-01-02"),
		}
		baseAccount.Ir = ir
		baseAccount.Kredit = document.TotalWithTax
	} else {
		baseAccount.Debet = document.TotalWithTax
	}

	baseAccount.Otvoritev = 0

}

func generateIncomeGoodsAccount(baseAccount *model.Knjizba, document model.Document) {

	country, err := findDocumentCountry(document)

	konto := -1
	if err != nil {
		konto = 0
	} else {
		if country.Name.Common == "Slovenia" {
			konto = 7620
		} else {
			if country.EuMember {
				konto = 7630
			} else {
				konto = 7632
			}
		}
	}

	baseAccount.Konto = konto
	baseAccount.Otvoritev = 0
	baseAccount.Kredit = document.Total
}

func generateIncomeAndServicesAccount(baseAccount *model.Knjizba, document model.Document) {
	baseAccount.Konto = 7601
	baseAccount.Otvoritev = 0
	baseAccount.Kredit = 0
}

func generateHigherOutgoingTaxAccount(baseAccount *model.Knjizba, document model.Document, tax float64) {
	baseAccount.Konto = 26000
	baseAccount.Otvoritev = 0
	baseAccount.Kredit = tax
}
func generateLowerOutgoingTaxAccount(baseAccount *model.Knjizba, document model.Document, tax float64) {
	baseAccount.Konto = 26001
	baseAccount.Otvoritev = 0
	baseAccount.Kredit = tax
}

// 1631 - CArd -> 1651 Paypal
func createCardAccount(baseAccount *model.Knjizba, document model.Document) {
	baseAccount.Konto = 1631
	baseAccount.Otvoritev = 0
	baseAccount.Debet = document.TotalWithTax
}
func createPaypalAccount(baseAccount *model.Knjizba, document model.Document) {
	baseAccount.Konto = 1651
	baseAccount.Otvoritev = 0
	baseAccount.Debet = document.TotalWithTax
}

func calculateOutgoingTax(document model.Document) (float64, float64) {
	var totalTaxLower float64 = 0
	var totalTaxHigher float64 = 0

	for _, tax := range document.DocumentTaxes {
		if tax.Tax <= 9.5 {
			totalTaxLower += tax.TotalTax
		} else {
			totalTaxHigher += tax.TotalTax
		}
	}
	return totalTaxLower, totalTaxHigher
}

func isEndCustomer(document model.Document) bool {
	fromWebshop := false
	for _, data := range document.DocumentIssuer.Data {
		if data.Name == "Spletna trgovina" || data.Name == "Trgovina" {
			fromWebshop = true
		}
	}

	return fromWebshop || document.DocumentClient.IsEndCustomer
}

func addXmlFileToZip(ctx context.Context, zipWriter *zip.Writer, fileBytes []byte) {

	zipFileHeader := &zip.FileHeader{
		Name:     "export.xml",
		Method:   zip.Deflate,
		Modified: time.Now(),
	}
	xmlFile, err := zipWriter.CreateHeader(zipFileHeader)

	log.Info(ctx, nil, "Writing xml bytes "+strconv.Itoa(len(fileBytes)))

	if err != nil {
		log.Error(ctx, err, nil, "Error creating xml file for zip")
		return
	}
	_, err = io.Copy(xmlFile, bytes.NewReader(fileBytes))
	if err != nil {
		log.Error(ctx, err, nil, "Error writing xml to Zip")
		return
	}
}

func newUniqueFileName(ext string) string {
	var name string
	uid, err := uuid.NewUUID()
	if err != nil {
		name = fmt.Sprintf("%d", time.Now().UnixNano())
	} else {
		name = uid.String()
	}
	return fmt.Sprintf("%s%s", name, ext)
}

func constructURLSlug(config Config, fileName string) string {
	return fmt.Sprintf("%s/%s/%s", config.Env, "accounting-files", fileName)
}

func newGCSBucketWriter(ctx context.Context, bucket string, objID string) (*storage.Writer, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Error(ctx, err, log.Data{"client": client}, "client error")
		return nil, err
	}
	bh := client.Bucket(bucket)
	if _, err = bh.Attrs(ctx); err != nil {
		return nil, err
	}
	obj := bh.Object(objID)
	w := obj.NewWriter(ctx)

	return w, nil
}
