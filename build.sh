#!/bin/bash -ex

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
VERSION=$1
GO_ARCH=$2
ARCH=$(dpkg --print-architecture)
BUILD_DIR=${DIR}/build
rm -rf ${BUILD_DIR}
mkdir -p ${BUILD_DIR}/bin
cd $DIR
go test ./...
go build -ldflags '-linkmode external -extldflags -static' -o ${BUILD_DIR}/bin/store ./cmd/store
go build -ldflags '-linkmode external -extldflags -static' -o ${BUILD_DIR}/bin/cli ./cmd/cli
cp -r ${DIR}/config ${BUILD_DIR}

OUT_DIR=${DIR}/out
rm -rf ${OUT_DIR}
mkdir $OUT_DIR
go build -ldflags '-linkmode external -extldflags -static' -o $OUT_DIR/syncloud-release-$GO_ARCH ./cmd/release
tar cpzf $OUT_DIR/syncloud-store-${VERSION}-${ARCH}.tar.gz -C $BUILD_DIR .
