#!/bin/bash -xe

SNAPD=$1

cd /tmp
rm -rf snapd
tar xzvf "${SNAPD}"
sed -i 's#Environment=SNAPPY_FORCE_API_URL=.*#Environment=SNAPPY_FORCE_API_URL=http://api.store.syncloud.org#g' snapd/conf/snapd.service
./snapd/upgrade.sh