ifndef $(GOPATH)
    GOPATH=$(shell go env GOPATH)
    export GOPATH
endif

.PHONY: mockery-prepare

# Install the mockery. This command will install the mockery in the GOPATH/bin folder
mockery-prepare:
	 @go get github.com/vektra/mockery/.../

# Use the mockery to generate mock interface
mockery-gen:
	@rm -rf ./mocks
	$(GOPATH)/bin/mockery --dir ./cache/ --name ICacheInteractor
	 

.PHONY: short-test
short-test:
	@go test -v --short ./...

.PHONY: lint-prepare
lint-prepare:
	@echo "Preparing Linter"
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh -s latest

.PHONY: lint
lint:
	@echo "Applying linter"
	./bin/golangci-lint run ./...
