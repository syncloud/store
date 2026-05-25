#!/bin/bash -xe

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd $DIR
export CGO_ENABLED=0
go test -c -o test
go build -o seed ./cmd/seed