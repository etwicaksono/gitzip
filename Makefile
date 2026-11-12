# Binary output name
BINARY_NAME=gitzip

# Detect OS
ifeq ($(OS),Windows_NT)
	BINARY=$(BINARY_NAME).exe
else
	BINARY=$(BINARY_NAME)
endif

# Default diff arguments for 'make run'
FROM ?= HEAD~1
TO   ?= HEAD
OUT  ?= changes.zip

.PHONY: all help build run install tidy fmt vet test clean

# Default target
all: build

help:
	@echo "GitZip Makefile commands:"
	@echo "  make build               Build the binary ($(BINARY))"
	@echo "  make run [FROM=] [TO=] [OUT=]  Run gitzip (default: FROM=HEAD~1 TO=HEAD OUT=changes.zip)"
	@echo "  make install             Install gitzip via 'go install'"
	@echo "  make tidy                Run 'go mod tidy'"
	@echo "  make fmt                 Format Go source files"
	@echo "  make vet                 Run 'go vet'"
	@echo "  make test                Run tests"
	@echo "  make clean               Remove binary and generated zip/txt files"

build:
	go build -o $(BINARY) .

run: build
	./$(BINARY) $(FROM) $(TO) $(OUT)

install:
	go install .

tidy:
	go mod tidy

fmt:
	go fmt ./...

vet:
	go vet ./...

test:
	go test -v ./...

clean:
	go clean
	rm -f $(BINARY_NAME) $(BINARY_NAME).exe main.exe *.zip *_deleted.txt

