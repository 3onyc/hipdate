FROM golang:1.3.3-wheezy
MAINTAINER 3onyc

RUN go get github.com/tools/godep
WORKDIR /go/src/github.com/3onyc/hipdate

COPY . /go/src/github.com/3onyc/hipdate

RUN godep restore && \
    go install github.com/3onyc/hipdate/hipdated

CMD ["hipdated"]
