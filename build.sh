#!/bin/bash -ex

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
VERSION=$1
GO_ARCH=$2
ARCH=$(dpkg --print-architecture)
BUILD_DIR=${DIR}/build
rm -rf ${BUILD_DIR}
mkdir -p ${BUILD_DIR}/bin
cd $DIR

GIT_SHA=${DRONE_COMMIT_SHA:-unknown}
BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS="-linkmode external -extldflags -static \
    -X github.com/syncloud/store/internal/version.GitSha=${GIT_SHA} \
    -X github.com/syncloud/store/internal/version.BuildNumber=${VERSION} \
    -X github.com/syncloud/store/internal/version.BuildTime=${BUILD_TIME}"

go build -ldflags "${LDFLAGS}" -o ${BUILD_DIR}/bin/store ./cmd/store
go build -ldflags "${LDFLAGS}" -o ${BUILD_DIR}/bin/cli ./cmd/cli
cp -r ${DIR}/config ${BUILD_DIR}
mkdir ${BUILD_DIR}/www

OUT_DIR=${DIR}/out
rm -rf ${OUT_DIR}
mkdir $OUT_DIR
go build -ldflags '-linkmode external -extldflags -static' -o $OUT_DIR/syncloud-release-$GO_ARCH ./cmd/release
tar cpzf $OUT_DIR/store-${VERSION}-${ARCH}.tar.gz -C $BUILD_DIR .
