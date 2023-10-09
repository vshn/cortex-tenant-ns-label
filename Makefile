NAME := cortex-tenant-ns-label
VERSION := $(shell cat VERSION)
GO_SRC := $(wildcard *.go)

GO ?= go

.PHONY: test
test:
	go test ./...

.PHONY: build
build: test $(NAME)

$(NAME): $(GO_SRC)
	CGO_ENABLED=0 \
	GOARCH=amd64 \
	GOOS=linux \
	$(GO) build -a -tags netgo -ldflags '-s -w -extldflags "-static" -X main.version=$(VERSION)' -o $@

.PHONY: clean
clean:
	rm -f $(NAME)
