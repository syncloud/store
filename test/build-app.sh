#!/bin/bash -xe

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
APP=$1
VERSION=$2
ARCH=$(uname -m)
APP_DIR=$DIR/$APP
BUILD_DIR=$APP_DIR/build
rm -rf ${BUILD_DIR}
mkdir ${BUILD_DIR}

ARCH=$(dpkg --print-architecture)
cp -r $APP_DIR/meta ${BUILD_DIR}
cp -r $APP_DIR/bin ${BUILD_DIR}
echo "version: $VERSION" >> ${BUILD_DIR}/meta/snap.yaml
echo "architectures:" >> ${BUILD_DIR}/meta/snap.yaml
echo "- ${ARCH}" >> ${BUILD_DIR}/meta/snap.yaml

mksquashfs ${BUILD_DIR} ${DIR}/${APP}_${VERSION}_${ARCH}.snap -noappend -comp xz -no-xattrs -all-root
