CURRENT_REVISION = $(shell git rev-parse --short HEAD)
BUILD_LDFLAGS = "-X main.revision=$(CURRENT_REVISION)"

ifdef update
  u=-u
endif

deps:
	go get ${u} -d -v ./...

test-deps:
	go get ${u} -d -t -v ./...

devel-deps: test-deps
	go get ${u} github.com/golang/lint/golint
	go get ${u} github.com/haya14busa/goverage
	go get ${u} github.com/mattn/goveralls
	go get ${u} github.com/motemen/gobump
	go get ${u} github.com/laher/goxc

test: test-deps
	go test ./...

lint: devel-deps
	go vet ./...
	golint -set_exit_status ./...

cover: devel-deps
	goverage -v -race -covermode=atomic ./...

build:
	go build -ldflags=$(BUILD_LDFLAGS)

crossbuild:
	goxc -pv=v$(shell gobump show -r) -build-ldflags=$(BUILD_LDFLAGS) \
	  -os=linux,darwin,windows -arch=amd64 -d=./dist \
	  -tasks=clean-destination,xc,archive,rmbin

.PHONY: deps test-deps devel-deps test lint cover build crossbuild
