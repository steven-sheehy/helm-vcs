BINARY  := helmvcs
GOARCH  := amd64
IMAGE   := steven-sheehy/helm-vcs
MAIN    := cmd/helmvcs/main.go
PKGS    := $(shell go list ./... | grep -v vendor)
VERSION := v0.1.0

all: clean build test install

build: build-linux build-macos build-windows

build-linux:
	GOOS=linux go build -v -o bin/linux/$(BINARY) $(MAIN)

build-macos:
	GOOS=darwin go build -v -o bin/macos/$(BINARY) $(MAIN)

build-windows:
	GOOS=windows go build -v -o bin/windows/$(BINARY) $(MAIN)

clean:
	rm -rf bin/ *.out coverage.html

coverage: test
	go tool cover -html=c.out -o coverage.html

docker:
	docker build -t $(IMAGE) .

install:
	go install -v $(MAIN)

lint:
	go get -u github.com/golangci/golangci-lint/cmd/golangci-lint &>/dev/null
	golangci-lint run $(PKGS)

test: lint
	go test $(PKGS) -v -cover -coverprofile=c.ou $(MAIN)

