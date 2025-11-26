.PHONY: all build test vet fmt tidy clean lint

all: build

build:
	go build ./...

test:
	go test -race -cover ./...

vet:
	go vet ./...

fmt:
	gofmt -s -w .

tidy:
	go mod tidy

lint:
	golangci-lint run --timeout=5m || true

clean:
	rm -rf bin build dist
