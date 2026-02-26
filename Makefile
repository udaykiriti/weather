.PHONY: all cli web build run-cli run-web clean fmt vet

BINARY_CLI = weather-cli
BINARY_WEB = weather-web

all: cli web

## Build the CLI binary
cli:
	go build -o $(BINARY_CLI) ./cmd/cli/

## Build the web server binary
web:
	go build -o $(BINARY_WEB) .

## Build both binaries
build: cli web

## Run the CLI (default city: London)
run-cli: cli
	./$(BINARY_CLI) $(ARGS)

## Run the web server
run-web: web
	./$(BINARY_WEB)

## Format all Go source files
fmt:
	gofmt -w .

## Run go vet
vet:
	go vet ./...

## Remove built binaries
clean:
	rm -f $(BINARY_CLI) $(BINARY_WEB)
