#!/bin/bash -xe

SNAPD=$1

cd /tmp
rm -rf snapd
tar xzvf "${SNAPD}"
./snapd/upgrade.sh