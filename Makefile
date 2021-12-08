.PHONY: lint

lint:
	golangci-lint run ./...

test:
	go test -cover -race ./...
	
unittest:
	go test -short ./...
