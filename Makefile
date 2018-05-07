CURRENT_REVISION = $(shell git rev-parse --short HEAD)
BUILD_LDFLAGS = "-X main.revision=$(CURRENT_REVISION)"
ifdef update
  u=-u
endif

deps:
	go get ${u} github.com/golang/dep/cmd/dep
	dep ensure

devel-deps: deps
	go get ${u} github.com/golang/lint/golint \
	  github.com/haya14busa/goverage          \
	  github.com/mattn/goveralls              \
	  github.com/motemen/gobump               \
	  github.com/Songmu/goxz/cmd/goxz         \
	  github.com/Songmu/ghch                  \
	  github.com/tcnksm/ghr

test: deps
	go test ./...

lint: devel-deps
	go vet ./...
	go list ./... | xargs golint -set_exit_status

cover: devel-deps
	goverage -v -race -covermode=atomic ./...

build: deps
	go build -ldflags=$(BUILD_LDFLAGS)

crossbuild: devel-deps
	$(eval ver = $(shell gobump show -r))
	goxz -pv=v$(ver) -build-ldflags=$(BUILD_LDFLAGS) \
	  -os=linux,darwin,windows -arch=amd64 -d=./dist/v$(ver)

release: devel-deps
	_tools/releng
	_tools/upload_artifacts

.PHONY: deps devel-deps test lint cover build crossbuild release
