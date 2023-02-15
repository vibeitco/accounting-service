package main

import (
	"context"

	"github.com/vibeitco/accounting-service/service"
	"github.com/vibeitco/go-utils/common"
	"github.com/vibeitco/go-utils/config"
	"github.com/vibeitco/go-utils/log"
	"github.com/vibeitco/go-utils/server"
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

	// handler
	uri := "/v1/accounting/test"
	handler, err := service.NewHandler(*conf)
	if err != nil {
		log.Fatal(ctx, err, nil, "failed creating handler")
	}

	r, srv, err := server.NewREST(&conf.Core, uri)
	if err != nil {
		log.Fatal(ctx, err, nil, "failed creating REST server")
	}
	r.Get(uri, handler.Auth)

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
