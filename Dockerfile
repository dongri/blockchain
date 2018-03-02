FROM golang:1.10-alpine
MAINTAINER Dongri Jin

ADD . /go/src/github.com/dongri/blockchain
WORKDIR /go/src/github.com/dongri/blockchain
RUN go install github.com/dongri/blockchain

# Launch a server instance
ENTRYPOINT /go/bin/blockchain

EXPOSE 5000
