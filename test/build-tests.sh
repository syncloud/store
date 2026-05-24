#!/bin/bash -xe

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd $DIR
go build -ldflags '-linkmode external -extldflags -static' -o seed ./cmd/seed