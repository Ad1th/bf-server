.PHONY: all build test test-race bench lint fmt run clean

BINARY_NAME=bf-server

all: test build

build:
	go build -o $(BINARY_NAME) ./cmd/bf-server

test:
	go test -v ./...

test-race:
	go test -v -race ./...

bench:
	go test -bench=. -benchmem ./pkg/interpreter

fmt:
	gofmt -s -w .

lint:
	go vet ./...

run: build
	./$(BINARY_NAME) --app ./examples/basic --dev

clean:
	rm -f $(BINARY_NAME) coverage.out coverage.html
