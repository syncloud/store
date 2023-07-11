#!/bin/bash -xe

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

$DIR/build-app.sh testapp1 1 amd64
$DIR/build-app.sh testapp1 1 arm64
$DIR/build-app.sh testapp1 1 armhf

$DIR/build-app.sh testapp1 2 amd64
$DIR/build-app.sh testapp1 2 arm64
$DIR/build-app.sh testapp1 2 armhf

$DIR/build-app.sh testapp1 3 amd64
$DIR/build-app.sh testapp1 3 arm64
$DIR/build-app.sh testapp1 3 armhf

$DIR/build-app.sh testapp2 1 amd64
$DIR/build-app.sh testapp2 1 arm64
$DIR/build-app.sh testapp2 1 armhf

$DIR/build-app.sh testapp2 2 amd64
$DIR/build-app.sh testapp2 2 arm64
$DIR/build-app.sh testapp2 2 armhf
