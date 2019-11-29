
lint:
		go mod tidy
		go mod verify
		golangci-lint run --enable-all --fix

build:
		go build -ldflags "-X main.version=$(shell git describe --tags --abbrev=0) -X main.commit=$(shell git rev-parse --short HEAD)"

all: lint build
