BINARY  := helm-vcs
IMAGE   := steven-sheehy/helm-vcs
MAIN    := cmd/main.go
PKGS    := $(shell go list ./... | grep -v vendor)
VERSION := v0.1.0

all: clean build test install

build:
	go build -v -o $(BINARY) $(MAIN)

clean:
	rm -rf ${BINARY} *.out coverage.html

coverage: test
	go tool cover -html=c.out -o coverage.html

install: build
	mkdir -p ~/.helm/plugins/helm-vcs
	cp ${BINARY} plugin.yaml LICENSE README.md ~/.helm/plugins/helm-vcs

lint:
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s
	./bin/golangci-lint run

test:
	go test $(PKGS) -v -cover -coverprofile=c.out $(MAIN)

