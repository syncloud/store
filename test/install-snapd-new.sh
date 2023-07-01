#!/bin/bash -xe

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

SNAPD=$1
cd ${DIR}
tar xzvf ${SNAPD}
sed -i 's#Environment=SNAPPY_FORCE_API_URL=.*#Environment=SNAPPY_FORCE_API_URL=http://api.store.syncloud.org#g' snapd/conf/snapd.service
./snapd/install.sh