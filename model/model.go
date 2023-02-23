package model

import (
	"encoding/xml"
	"time"
)

type Prenos struct {
	Knjizba                   []Knjizba `xml:"Knjizba"`
	XsiType                   xml.Attr  `xml:"xsi type,attr"`
	NoNamespaceSchemaLocation string    `xml:"xsi:noNamespaceSchemaLocation,attr"`
}

type Payment struct {
	Amount         float64   `json:"amount"`
	Type           string    `json:"type"`
	Date           time.Time `json:"date"`
	ID             string    `json:"id"`
	DocumentID     string    `json:"documentId"`
	OrganizationID string    `json:"organizationId"`
	IsDeleted      bool      `json:"_isDeleted"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type Knjizba struct {
	XMLName            xml.Name          `xml:"Knjizba"`
	Simbol             string            `xml:"Simbol"`
	Dokument           string            `xml:"Dokument"`
	Veza               string            `xml:"Veza"`
	DatumKnjizneja     string            `xml:"Datum_knjizenja"`
	Obracunsko_obdobje ObracunskoObdobje `xml:"Obracunsko_obdobje"`
	DatumDokumenta     string            `xml:"Datum_dokumenta"`
	Opisdokumenta      string            `xml:"Opis_dokumenta"`
	Attachment         Attachment        `xml:"Priloga"`

	Otvoritev float64 `xml:"Otvoritev"`
	Debet     float64 `xml:"Debet"`
	Kredit    float64 `xml:"Kredit"`

	Konto   int     `xml:"Konto"`
	Partner Partner `xml:"Partner,omitempty"`
	Ir      Ir      `xml:"Ir,omitempty"`
}

type Ir struct {
	Ddv        []Ddv  `xml:"Ddv,omitempty"`
	DatumZaDdv string `xml:"Datum_za_ddv,omitempty"`
}

type Ddv struct {
	VrstaDdv string  `xml:"vrsta_ddv"`
	Osnova   float64 `xml:"Osnova"`
	Znesek   float64 `xml:"Znesek"`
}

type Partner struct {
	Sifra                     string `xml:"Sifra,omitempty"`
	Naziv1                    string `xml:"Naziv1,omitempty"`
	Naslov                    string `xml:"Naslov,omitempty"`
	Posta                     string `xml:"Posta,omitempty"`
	Drzava                    string `xml:"Drzava,omitempty"`
	Identifikacijska_stevilka string `xml:"Identifikacijska_stevilka,omitempty"`
	Davcna_stevilka           string `xml:"Davcna_stevilka,omitempty"`
	Maticna_stevilka          string `xml:"Maticna_stevilka,omitempty"`
}

type ObracunskoObdobje struct {
	Mesec int `xml:"Mesec"`
	Leto  int `xml:"Leto"`
}

type Attachment struct {
	XMLName xml.Name        `xml:"Priloga"`
	File    AttachementFile `xml:"Datoteka"`
}

type AttachementFile struct {
	XMLName xml.Name `xml:"Datoteka"`
	XsiType xml.Attr `xml:"xsi type,attr"`
	XsiXs   string   `xml:"xsi:xs,attr"`
	XsiXsi  string   `xml:"xsi:xsi,attr"`
	Link    string   `xml:",chardata"`
}

type Document struct {
	Number        string      `json:"number"`
	Draft         bool        `json:"draft"`
	Date          time.Time   `json:"date"`
	DateService   time.Time   `json:"dateService"`
	DateServiceTo interface{} `json:"dateServiceTo"`
	DateDue       time.Time   `json:"dateDue"`
	Reference     string      `json:"reference"`
	Total         float64     `json:"total"`
	TotalDiscount float64     `json:"totalDiscount"`
	TotalWithTax  float64     `json:"totalWithTax"`
	DecimalPlaces int         `json:"decimalPlaces"`
	Note          string      `json:"note"`
	TaxClause     string      `json:"taxClause"`
	Footer        string      `json:"footer"`
	Signature     string      `json:"signature"`
	Type          string      `json:"type"`
	Canceled      bool        `json:"canceled"`
	SentSnailMail bool        `json:"sentSnailMail"`
	TotalPaid     float64     `json:"totalPaid"`
	TotalDue      float64     `json:"totalDue"`
	PaidInFull    bool        `json:"paidInFull"`
	CurrencyID    string      `json:"currencyId"`
	DocumentTaxes []struct {
		Tax          float32 `json:"tax"`
		TaxID        string  `json:"taxId"`
		Abbreviation string  `json:"abbreviation"`
		Base         float64 `json:"base"`
		TotalTax     float64 `json:"totalTax"`
	} `json:"_documentTaxes"`
	DocumentReverseTaxes []interface{} `json:"_documentReverseTaxes"`
	HasUnit              bool          `json:"hasUnit"`
	ValidateEslog        bool          `json:"validateEslog"`
	IsValidEslog         bool          `json:"isValidEslog"`
	Incoming             bool          `json:"incoming"`
	IssuedAt             time.Time     `json:"issuedAt"`
	DateYear             int           `json:"dateYear"`
	ID                   string        `json:"id"`
	DocumentIds          []interface{} `json:"documentIds"`
	CreatedAt            time.Time     `json:"createdAt"`
	UpdatedAt            time.Time     `json:"updatedAt"`
	Custom               struct {
	} `json:"custom"`
	L10N struct {
		NoteEn string `json:"note_en"`
		NoteHr string `json:"note_hr"`
		NoteDe string `json:"note_de"`
		NoteIt string `json:"note_it"`
		NoteFr string `json:"note_fr"`
		NoteCs string `json:"note_cs"`
		NoteSk string `json:"note_sk"`
		NoteLv string `json:"note_lv"`
		NoteNl string `json:"note_nl"`
		NoteEs string `json:"note_es"`
		NoteRo string `json:"note_ro"`
		NoteSr string `json:"note_sr"`
	} `json:"l10n"`
	IsDeleted     bool `json:"_isDeleted"`
	DocumentItems []struct {
		ID               string      `json:"id"`
		Discount         int         `json:"discount"`
		DiscountIsAmount bool        `json:"discountIsAmount"`
		Quantity         int         `json:"quantity"`
		Total            float64     `json:"total"`
		TotalWithTax     float64     `json:"totalWithTax"`
		TotalTax         float64     `json:"totalTax"`
		TotalDiscount    float64     `json:"totalDiscount"`
		IsSeparator      bool        `json:"isSeparator"`
		Name             string      `json:"name"`
		Description      string      `json:"description"`
		Price            int         `json:"price"`
		Unit             string      `json:"unit"`
		TrackInventory   bool        `json:"trackInventory"`
		Custom           interface{} `json:"custom"`
		L10N             struct {
			NameEn        string `json:"name_en"`
			NameHr        string `json:"name_hr"`
			NameDe        string `json:"name_de"`
			NameIt        string `json:"name_it"`
			NameFr        string `json:"name_fr"`
			NameCs        string `json:"name_cs"`
			NameSk        string `json:"name_sk"`
			NameLv        string `json:"name_lv"`
			NameNl        string `json:"name_nl"`
			NameEs        string `json:"name_es"`
			NameRo        string `json:"name_ro"`
			NameSr        string `json:"name_sr"`
			DescriptionEn string `json:"description_en"`
			DescriptionHr string `json:"description_hr"`
			DescriptionDe string `json:"description_de"`
			DescriptionIt string `json:"description_it"`
			DescriptionFr string `json:"description_fr"`
			DescriptionCs string `json:"description_cs"`
			DescriptionSk string `json:"description_sk"`
			DescriptionLv string `json:"description_lv"`
			DescriptionNl string `json:"description_nl"`
			DescriptionEs string `json:"description_es"`
			DescriptionRo string `json:"description_ro"`
			DescriptionSr string `json:"description_sr"`
		} `json:"l10n"`
		DocumentItemTaxes []struct {
			ID             string `json:"id"`
			Rate           int    `json:"rate"`
			ReverseCharged bool   `json:"reverseCharged"`
			Name           string `json:"name"`
			Abbreviation   string `json:"abbreviation"`
			Recoverable    bool   `json:"recoverable"`
			Compound       bool   `json:"compound"`
			TaxID          string `json:"taxId"`
		} `json:"_documentItemTaxes"`
		Components []interface{} `json:"_components"`
		Data       []interface{} `json:"_data"`
	} `json:"_documentItems"`
	DocumentIssuer struct {
		ID                string    `json:"id"`
		Name              string    `json:"name"`
		Address           string    `json:"address"`
		Address2          string    `json:"address2"`
		City              string    `json:"city"`
		Zip               string    `json:"zip"`
		Country           string    `json:"country"`
		CountryAlpha2Code string    `json:"countryAlpha2Code"`
		TaxNumber         string    `json:"taxNumber"`
		TaxSubject        bool      `json:"taxSubject"`
		CompanyNumber     string    `json:"companyNumber"`
		Iban              string    `json:"IBAN"`
		Bank              string    `json:"bank"`
		Swift             string    `json:"SWIFT"`
		CreatedAt         time.Time `json:"createdAt"`
		UpdatedAt         time.Time `json:"updatedAt"`
		Custom            struct {
		} `json:"custom"`
		Data []struct {
			ID      string `json:"id"`
			Name    string `json:"name"`
			Value   string `json:"value"`
			Options struct {
				OnPDF bool `json:"onPDF"`
				Bold  bool `json:"bold"`
			} `json:"options"`
		} `json:"_data"`
	} `json:"_documentIssuer"`
	ClientID       string `json:"clientId"`
	DocumentClient struct {
		ID                string      `json:"id"`
		IsEndCustomer     bool        `json:"isEndCustomer"`
		Name              string      `json:"name"`
		Address           string      `json:"address"`
		Address2          interface{} `json:"address2"`
		City              string      `json:"city"`
		Zip               string      `json:"zip"`
		Country           string      `json:"country"`
		CountryAlpha2Code string      `json:"countryAlpha2Code"`
		TaxNumber         string      `json:"taxNumber"`
		TaxSubject        bool        `json:"taxSubject"`
		CompanyNumber     string      `json:"companyNumber"`
		Iban              interface{} `json:"IBAN"`
		Bank              interface{} `json:"bank"`
		Swift             interface{} `json:"SWIFT"`
		Email             string      `json:"email"`
		Phone             interface{} `json:"phone"`
		CreatedAt         time.Time   `json:"createdAt"`
		UpdatedAt         time.Time   `json:"updatedAt"`
		Custom            struct {
		} `json:"custom"`
		Data    []interface{} `json:"_data"`
		Contact string        `json:"contact"`
	} `json:"_documentClient"`
	OrganizationID         string        `json:"organizationId"`
	AccountID              string        `json:"accountId"`
	Comments               []interface{} `json:"_comments"`
	PriceListID            string        `json:"priceListId"`
	DestinationWarehouseID interface{}   `json:"destinationWarehouseId"`
	Data                   []interface{} `json:"_data"`
	Pdf                    string        `json:"pdf"`
}
