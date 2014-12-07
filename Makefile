all: test build

test:
	godep go test ./...

build:
	godep go install github.com/3onyc/hipdate/hipdated

.PHONY: all test build
