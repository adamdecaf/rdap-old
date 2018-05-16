.PHONY: build check test

linux: linux_amd64
linux_amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/rdap-linux github.com/adamdecaf/rdap/cmd/rdap

osx: osx_amd64
osx_amd64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o bin/rdap-osx github.com/adamdecaf/rdap/cmd/rdap

dist: build linux osx

check:
	go vet ./...
	go fmt ./...

test: check
	go test ./... -v -count 1

build: check
	go build -o bin/rdap github.com/adamdecaf/rdap/cmd/rdap
	@chmod +x bin/rdap
