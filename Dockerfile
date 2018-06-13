FROM golang:1.10.3-stretch
MAINTAINER Harry Lee

WORKDIR /go/src/app

RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

RUN useradd --create-home --shell /bin/bash 1000
RUN chown -R 1000 /go

USER 1000

