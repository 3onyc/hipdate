all: deps build

deps:
	godep restore

build:
	go install github.com/3onyc/hipdate/hipdated
