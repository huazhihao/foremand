.PHONY: build install clean test
VERSION=`egrep -o '[0-9]+\.[0-9a-z.\-]+' version.go`
GIT_SHA=`git rev-parse --short HEAD || echo`

build:
	@echo "Building foremand..."
	@mkdir -p bin
	@go build -ldflags "-X main.GitSHA=${GIT_SHA}" -o bin/foremand .

install:
	@echo "Installing foremand..."
	@install -c bin/foremand /usr/local/bin/foremand

clean:
	@rm -f bin/*

test:
	@echo "Running tests..."
	@go test -v ./...
