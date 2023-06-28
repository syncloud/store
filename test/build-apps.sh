#!/bin/bash -xe

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
$DIR/build-app.sh testapp1 1
$DIR/build-app.sh testapp1 2
$DIR/build-app.sh testapp1 3
$DIR/build-app.sh testapp2 1
$DIR/build-app.sh testapp2 2
