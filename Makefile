export

LOCAL_BIN:=$(CURDIR)/bin
PATH:=$(LOCAL_BIN):$(PATH)

# HELP =================================================================================================================
# This will output the help for each task
# thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help

help: ## Display this help screen
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

install: ## Ensure the go.mod file is clean and updated with the project dependencies.
	pip3 install pre-commit commitizen
	go mod tidy
.PHONY: install

test: ## Clear the test cache and then execute all project tests with coverage.
	@mkdir -p coverage
	@go clean -testcache
	go test -v -failfast -race -cover -covermode=atomic ./test/... -coverpkg=./pkg/... -coverprofile=coverage/coverage.out -shuffle=on
	@echo "🧪 Test Completed"
.PHONY: test

coverage: ## Generate and visualize a test coverage report in HTML format.
	@mkdir -p coverage
	@go clean -testcache
	@go test -v -failfast -race -cover -covermode=atomic ./test/... -coverpkg=./pkg/... -coverprofile=coverage/coverage.out -shuffle=on > /dev/null
	@go tool cover -func=coverage/coverage.out
	@go tool cover -html=coverage/coverage.out -o coverage/coverage.html
	@echo "🧪 Test coverage completed"
.PHONY: coverage

linter: ## Run the golangci-lint on the project source code to detect style issues or errors.
	golangci-lint run
.PHONY: linter