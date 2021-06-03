APP          			:= bookstore
PKG_LIST     			:= $(shell go list ./... | grep -v /vendor/)
LD_FLAGS     			:= "-s -w"
GOLANGCI_LINT_VERSION 	:= "v1.40.1"


.PHONY: \
	bookstore \
	test 
bookstore:
	@ go build -ldflags=$(LD_FLAGS) -o $(APP) main.go

run:
	@ LOG_DATE_TIME=1 go run main.go

clean:
	@ rm -f $(APP)-* $(APP)

test:
	go test -cover -race -count=1 ./... -coverpkg="$(APP)/..."

lint:
	@if [ ! -f ./bin/golangci-lint ]; then \
		curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s $(GOLANGCI_LINT_VERSION); \
	fi;
	@echo "golangci-lint checking..."
	@./bin/golangci-lint run --deadline=30m --skip-dirs tests --enable=misspell --enable=gosec --enable=gofmt --enable=goimports --enable=revive --enable sqlclosecheck --enable whitespace ./...
	@go vet ./...


