all: test build

test:
	godep go test ./...

check:
	OUTPUT=$$(gofmt -e -l .); echo $$OUTPUT; [ $$(echo -n "$$OUTPUT" | wc -l) -eq 0 ] || false
	go tool vet --composites=false backends hipdated shared sources
	go tool vet --composites=false $(wildcard *.go)
	#golint ./...

build:
	godep go install github.com/3onyc/hipdate/hipdated

.PHONY: all test build check
