.PHONY: all build test vet fmt tidy clean

all: build

build:
	go build ./...

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -s -w .

tidy:
	go mod tidy

clean:
	rm -rf bin build dist
