.PHONY: format test clean lint tidy

format:
	go fmt ./...

lint:
	golangci-lint run

test:
	go test ./...

test-v:
	go test -v ./...

test-race:
	go test -race ./...

clean:
	rm -f ./bin/

tidy:
	go mod tidy
	go mod vendor