all: test build

test:
	godep go test ./...

check:
	OUTPUT=$$(gofmt -e -l .); echo $$OUTPUT; [ $$(echo -n "$$OUTPUT" | wc -l) -eq 0 ] || false
	go tool vet --composites=false backends hipdated shared sources
	go tool vet --composites=false $(wildcard *.go)
	#golint ./...

build:
	godep go build -o hipdated-$(shell uname -m) github.com/3onyc/hipdate/hipdated

release:
	github-release upload -u 3onyc -r hipdate -t $(RELEASE_TAG) -n hipdated-x86_64 -f hipdated-x86_64

clean:
	rm hipdated-*

.PHONY: all test build check clean
