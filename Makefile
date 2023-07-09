.PHONY: format test clean lint tidy

format:
	go fmt ./...

lint:
	golangci-lint run

clean-test:
	go clean -testcache

test:
	$(MAKE) clean-test && go test -parallel 4 ./pkg/...

test-v:
	$(MAKE) clean-test && go test -v -cover ./pkg/...

test-race:
	$(MAKE) clean-test && go test -race ./pkg/...

clean:
	rm -f ./bin/

tidy:
	go mod tidy
	go mod vendor