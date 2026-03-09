BINARY  = termf1
BUILD   = go build -o $(BINARY) .
RUN     = ./$(BINARY)

.PHONY: all build run install clean tidy

all: build

## build: compile the binary
build:
	$(BUILD)

## run: build and launch the dashboard
run: build
	$(RUN)

## install: install to $GOPATH/bin so you can run `termf1` anywhere
install:
	go install .

## tidy: tidy go modules
tidy:
	go mod tidy

## clean: remove the compiled binary
clean:
	rm -f $(BINARY)

## vet: run go vet
vet:
	go vet ./...
