#!/bin/bash
IMAGE_NAME=x3tech/hipdated
VERSION?=latest

all: build docker clean

build:
	make -C ../
	cp "$(GOPATH)/bin/hipdated" "$(CURDIR)/hipdated"

docker: build
	docker build -t "$(IMAGE_NAME):$(VERSION)" .

clean:
	[ ! -f "$(CURDIR)/hipdated" ] || rm "$(CURDIR)/hipdated"

.PHONY: build docker clean
