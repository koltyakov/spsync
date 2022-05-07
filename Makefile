ifneq (,$(wildcard ./.env))
	include .env
	export
endif

install:
	go get -u ./... && go mod tidy

format:
	gofmt -s -w .

run:
	go run ./cmd/sync
