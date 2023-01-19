.PHONY: build test

.DEFAULT_GOAL := build

build:
	CGO_ENABLED=0 GOOS=linux go build -o handler cmd/main.go

test:
	go test -v ./...
