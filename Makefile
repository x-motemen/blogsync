VERSION = $(shell godzil show-version)
CURRENT_REVISION = $(shell git rev-parse --short HEAD)
BUILD_LDFLAGS = "-s -w -X main.revision=$(CURRENT_REVISION)"
u := $(if $(update),-u)

export GO111MODULE=on

.PHONY: deps
deps:
	go get ${u} -d -v
	go mod tidy

.PHONY: devel-deps
devel-deps: deps
	go install github.com/tcnksm/ghr@latest
	go install github.com/Songmu/godzil/cmd/godzil@latest

.PHONY: test
test: deps
	go test ./...

.PHONY: build
build: deps
	go build -ldflags=$(BUILD_LDFLAGS)

.PHONY: release
release: devel-deps
	godzil release

.PHONY: CREDITS
CREDITS: deps devel-deps go.sum
	godzil credits -w

.PHONY: crossbuild
crossbuild: devel-deps
	godzil crossbuild -pv=v$(VERSION) -build-ldflags=$(BUILD_LDFLAGS) \
	  -os=linux,darwin,windows -arch=amd64 -d=./dist/v$(VERSION)

.PHONY: upload
upload:
	ghr -body="$$(godzil changelog --latest -F markdown)" v$(VERSION) dist/v$(VERSION)
