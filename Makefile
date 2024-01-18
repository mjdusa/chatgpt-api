GIT_REPO:=github.com/mjdusa/chatgpt-api
BRANCH:=$(shell git rev-parse --abbrev-ref HEAD)
COMMIT:=$(shell git log --pretty=format:'%H' -n 1)
BUILD_TS:=$(shell date -u "+%Y-%m-%dT%TZ")
BUILD_DIR:=dist
APP_NAME:=chatgpt-api
#APP_VERSION:=$(shell git describe --tags)
APP_VERSION:=$(shell cat .version)
GO_VERSION:=$(shell go version | sed -r 's/go version go(.*)\ .*/\1/')
GOBIN:=${GOPATH}/bin

GOFLAGS = -a
LDFLAGS = -s -w -X '$(GIT_REPO)/internal/version.AppVersion=$(APP_VERSION)' -X '$(GIT_REPO)/internal/version.Branch=$(BRANCH)' -X '$(GIT_REPO)/internal/version.BuildTime=$(BUILD_TS)' -X '$(GIT_REPO)/internal/version.Commit=$(COMMIT)' -X '$(GIT_REPO)/internal/version.GoVersion=$(GO_VERSION)'
#GOCMD = GOPRIVATE='' ; CGO_ENABLED='0' ; GO111MODULE='on' ; go
GOCMD = GOPRIVATE='' ; GO111MODULE='on' ; go

LINTER_REPORT = $(BUILD_DIR)/golangci-lint-$(BUILD_TS).out
COVERAGE_REPORT = $(BUILD_DIR)/unit-test-coverage-$(BUILD_TS)

.PHONY: clean
clean:
	@echo "clean"
	rm -rf $(BUILD_DIR)
	go clean --cache

.PHONY: $(BUILD_DIR)
$(BUILD_DIR):
	mkdir -p $@

.PHONY: installdep
installdep:
ifeq (,$(wildcard $(GOBIN)/golangci-lint))
	@echo "Installing golangci-lint..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
else
	@echo "$(GOBIN)/golangci-lint detected, skipping install."
endif
ifeq (,$(wildcard $(GOBIN)/gcov2lcov))
	@echo "Installing gcov2lcov..."
	go install github.com/jandelgado/gcov2lcov@latest
else
	@echo "$(GOBIN)/gcov2lcov detected, skipping install."
endif
#ifeq (,$(wildcard $(which pre-commit)))
#	@echo "Brew installing pr-commit"
#	@brew install pre-commit || true
#else
#	@echo "pre-commit detected, skipping install."
#endif

.PHONY: init
init: installdep
ifeq (,$(wildcard ./.git/hooks/pre-commit))
	@echo "Adding pre-commit hook to .git/hooks/pre-commit"
	ln -s $(shell pwd)/hooks/pre-commit $(shell pwd)/.git/hooks/pre-commit || true
endif

.PHONY: prebuild
prebuild: init $(BUILD_DIR)
	@echo "Running go mod tidy & vendor"
	go version
	go env
	$(GOCMD) mod tidy && $(GOCMD) mod vendor

.PHONY: golangcilint
golangcilint: init $(BUILD_DIR)
	echo "Running golangci-lint"
	${GOPATH}/bin/golangci-lint --version
	${GOPATH}/bin/golangci-lint run --verbose --config .github/linters/.golangci.yml \
	  --issues-exit-code 0 --out-format=checkstyle > "$(LINTER_REPORT)"
#	cat $(LINTER_REPORT)

.PHONY: lint
lint: init golangcilint

.PHONY: unittest
unittest: init $(BUILD_DIR)
	$(GOCMD) test -coverprofile="$(COVERAGE_REPORT).gcov" ./... && gcov2lcov -infile "$(COVERAGE_REPORT).gcov" -outfile "$(COVERAGE_REPORT).lcov"
	$(GOCMD) tool cover -func="$(COVERAGE_REPORT).gcov"
#	$(GOCMD) tool cover -html="$(COVERAGE_REPORT).gcov"
#	gcov2lcov -infile "$(COVERAGE_REPORT).gcov" -outfile "$(COVERAGE_REPORT).lcov"
#	cat "$(COVERAGE_REPORT).gcov"
#	cat "$(COVERAGE_REPORT).lcov"

.PHONY: racetest
racetest:
	$(GOCMD) test -race ./...

.PHONY: test
test: unittest racetest

.PHONY: build
build: prebuild
	$(GOCMD) build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME) cmd/$(APP_NAME)/main.go

.PHONY: debug
debug: GOFLAGS += -x -v
debug: clean build

.PHONY: release
release: clean build

.PHONY: all
all: clean lint test build debug release

.PHONY: pre-commit
pre-commit: init
	pre-commit run --all-files

.PHONY: usage
usage:
	@echo "usage:"
	@echo "  make [command]"
	@echo "available commands:"
	@echo "  all - clean lint test build debug release"
	@echo "  clean - clean up build artifacts"
	@echo "  debug - build debug version of binary"
	@echo "  help - show usage"
	@echo "  installdep - install latest build app dependancies (ie: golangci-lint, gcov2lcov)"
	@echo "  lint - run all linter checks"
	@echo "  release - build release version of binary"
	@echo "  test - run all tests"
	@echo "  usage - show this information"

.PHONY: help
help: usage
