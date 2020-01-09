.PHONY: mockery-prepare
mockery-prepare:
	@echo "Installing mockery"
	@go get -u github.com/vektra/mockery

.PHONY: short-test
short-test:
	@go test -v --short

.PHONY: lint-prepare
lint-prepare:
	@echo "Preparing Linter"
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh -s latest

.PHONY: lint
lint:
	@echo "Applying linter"
	./bin/golangci-lint run ./...
