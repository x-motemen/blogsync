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

test: test-deps
	go test ./...

lint: devel-deps
	go vet ./...
	golint -set_exit_status ./...

cover: devel-deps
	goverage -v -race -covermode=atomic ./...

build:
	go build

.PHONY: deps test-deps devel-deps test lint cover build
