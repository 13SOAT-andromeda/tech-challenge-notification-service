.PHONY: build test lint clean

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/bootstrap ./cmd/lambda

test:
	go test ./...

lint:
	golangci-lint run

clean:
	rm -rf bin/
