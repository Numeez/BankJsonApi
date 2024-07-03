build:
	@go build -o bin/BankJsonApi

run:build
	@./bin/BankJsonApi

test:
	@go test -v ./...
