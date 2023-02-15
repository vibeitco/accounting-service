.ONESHELL:
SHELL=/bin/bash
-include .env

init:
	cp -n .env.example .env
	go mod download

build:
	go build -o svc

build-target:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./svc .

build-docker: build-target
	docker build --tag vibeitco/${SERVICE}:${VERSION} .

run: build
	./svc

run-docker: build-docker
	docker run --rm \
		-e SERVICE=${SERVICE} \
		-e ENV=${ENV} \
		-e VERSION=${VERSION} \
		-e MONGODB_PASSWORD=${MONGODB_PASSWORD} \
		-p 2020:2020 \
		-p 8080:8080 \
		vibeitco/${SERVICE}:${VERSION}