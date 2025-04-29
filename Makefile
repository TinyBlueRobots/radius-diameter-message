.PHONY: all build clean lint format test

tidy:
	rm go.sum	
	go clean -cache	
	go mod tidy

lint:
	go tool golangci-lint run

format:
	go fmt ./...

test:
	go test ./...
