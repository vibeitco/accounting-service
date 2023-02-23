package main

import (
	"context"
	"time"

	"github.com/vibeitco/accounting-service/service"
	"github.com/vibeitco/go-utils/common"
	"github.com/vibeitco/go-utils/config"
	"github.com/vibeitco/go-utils/dns"
	"github.com/vibeitco/go-utils/log"
	"github.com/vibeitco/go-utils/server"
	"google.golang.org/grpc"

	emailSvc "github.com/vibeitco/email-service/model"
	commonSvc "github.com/vibeitco/service-definitions/go/common"
)

func main() {
	ctx := context.Background()
	// read config
	conf := &service.Config{}
	err := config.Populate(conf)
	if err != nil {
		log.Fatal(ctx, err, nil, "failed populating config")
	}
	log.Info(ctx, log.Data{"conf": conf}, "config")

	emailService, err := NewEmailClient(ctx, conf)
	if err != nil {
		log.Fatal(ctx, err, nil, "failed creating user service client")
	}

	// handler
	uri := "/v1/accounting"
	handler, err := service.NewHandler(*conf, emailService)
	if err != nil {
		log.Fatal(ctx, err, nil, "failed creating handler")
	}

	r, srv, err := server.NewREST(&conf.Core, uri)
	if err != nil {
		log.Fatal(ctx, err, nil, "failed creating REST server")
	}
	r.Get(uri+"/export", handler.GenerateAccountingExport)

	// serve REST
	server.Run(ctx, "rest",
		func() {
			srv.ListenAndServe()
		},
		func() {
			err := srv.Shutdown(ctx)
			if err != nil {
				log.Error(ctx, err, nil, common.EventServerFatal)
			}
		})
}

func NewEmailClient(ctx context.Context, conf *service.Config) (emailSvc.EmailServiceClient, error) {
	//address := dns.AddressFor(dns.ServiceUser, &conf.Core, conf.Core.PortGRPC)
	address := dns.AddressFor("email-service", &conf.Core, conf.Core.PortGRPC)
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithTimeout(time.Second))
	if err != nil {
		return nil, err
	}
	client := emailSvc.NewEmailServiceClient(conn)
	_, err = client.Status(ctx, &commonSvc.GetServiceStatusRequest{})
	if err != nil {
		return nil, err
	}
	return client, nil
}
