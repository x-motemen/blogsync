VERSION = $(shell godzil show-version)
CURRENT_REVISION = $(shell git rev-parse --short HEAD)
BUILD_LDFLAGS = "-s -w -X main.revision=$(CURRENT_REVISION)"
ifdef update
  u=-u
endif

export GO111MODULE=on

deps:
	go get ${u} -d -v

devel-deps: deps
	sh -c '\
      tmpdir=$$(mktemp -d); \
	  cd $$tmpdir; \
	  go get ${u} \
	    golang.org/x/lint/golint            \
	    github.com/Songmu/godzil/cmd/godzil \
	    github.com/tcnksm/ghr;              \
	  rm -rf $$tmpdir'

test: deps
	go test ./...

lint: devel-deps
	go vet ./...
	golint -set_exit_status ./...

cover: devel-deps
	goverage -v -race -covermode=atomic ./...

build: deps
	go build -ldflags=$(BUILD_LDFLAGS)

CREDITS: deps devel-deps go.sum
	godzil credits -w

crossbuild: devel-deps
	godzil crossbuild -pv=v$(VERSION) -build-ldflags=$(BUILD_LDFLAGS) \
	  -os=linux,darwin,windows -arch=amd64 -d=./dist/v$(VERSION)

bump: devel-deps
	godzil release

upload:
	ghr v$(VERSION) dist/v$(VERSION)

release: bump crossbuild upload

.PHONY: deps devel-deps test lint cover build crossbuild bump upload release
