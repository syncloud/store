#!/bin/bash -xe

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

STORE=syncloud-store.tar.gz
systemctl stop syncloud-store.service || true
systemctl disable syncloud-store.service || true

rm -rf /usr/lib/syncloud-store
mkdir -p /usr/lib/syncloud-store
tar xzvf $DIR/$STORE -C /usr/lib/syncloud-store

cp /usr/lib/syncloud-store/config/syncloud-store.service /lib/systemd/system/

systemctl enable syncloud-store.service
systemctl start syncloud-store.service
