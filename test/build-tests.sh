#!/bin/bash -xe

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd $DIR
go test -ldflags '-linkmode external -extldflags -static' -c -o test