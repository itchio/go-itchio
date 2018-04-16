#!/bin/sh -xe

go version

export GOPATH=$PWD/gopath
rm -rf $GOPATH

export PKG=github.com/itchio/go-itchio
export PATH=$PATH:$GOPATH/bin

mkdir -p $GOPATH/src/$PKG
rsync -a --exclude 'gopath' . $GOPATH/src/$PKG

go get -v -d -t $PKG/...
go test -v -cover -coverprofile=coverage.txt -race ./...

curl -s https://codecov.io/bash | bash

