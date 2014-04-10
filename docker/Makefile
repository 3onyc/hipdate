#!/bin/bash
IMAGE_NAME=x3tech/hipdated
VERSION?=latest

all: build docker clean

build:
	make -C ../
	cp "$(GOPATH)/bin/hipdate" "$(CURDIR)/hipdate"

docker: build
	docker build -t "$(IMAGE_NAME):$(VERSION)" .

clean:
	[ ! -f "$(CURDIR)/hipdate" ] || rm "$(CURDIR)/hipdate"

.PHONY: build docker clean
