#!/bin/bash -ex

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

if [ "$#" -lt 1 ]; then
    echo "usage $0 device"
    exit 1
fi

STORE_DIR=/var/www/html
SCP="sshpass -p syncloud scp -o StrictHostKeyChecking=no"
SSH="sshpass -p syncloud ssh -o StrictHostKeyChecking=no"
ARTIFACTS_DIR=${DIR}/artifacts
mkdir $ARTIFACTS_DIR
LOG_DIR=$ARTIFACTS_DIR/log
SNAP_ARCH=$(dpkg --print-architecture)

apt update
apt install -y sshpass curl wget
cd $DIR

./wait-for-device.sh device
./wait-for-device.sh api.store.syncloud.org
./wait-for-device.sh apps.syncloud.org

mkdir -p $LOG_DIR

$SCP ${DIR}/install-store.sh root@api.store.syncloud.org:/install-store.sh
$SCP ${DIR}/../out/syncloud-store-*.tar.gz root@api.store.syncloud.org:/syncloud-store.tar.gz

wget --progress=dot:giga https://github.com/syncloud/snapd/releases/download/2/snapd-361-amd64.tar.gz -O snapd1.tar.gz
$SCP snapd1.tar.gz root@device:/
wget --progress=dot:giga https://github.com/syncloud/snapd/releases/download/3-rc/snapd-514-amd64.tar.gz -O snapd2.tar.gz
$SCP snapd2.tar.gz root@device:/

$SCP ${DIR}/install-snapd1.sh root@device:/
$SCP ${DIR}/install-snapd2.sh root@device:/
$SCP ${DIR}/upgrade-snapd2.sh root@device:/

$SSH root@api.store.syncloud.org /install-store.sh
$SCP ${DIR}/testapp2_1_$SNAP_ARCH.snap root@device:/testapp2_1.snap
#$SCP ${DIR}/test root@$DEVICE:/

code=0
set +e
${DIR}/test
code=$(($code+$?))
#$SSH root@$DEVICE /test -test.run Inside
#code=$(($code+$?))
#
set -e

$SSH root@device snap changes > $LOG_DIR/snap.changes.log || true
$SSH root@device journalctl > $LOG_DIR/journalctl.device.log
$SSH apps.syncloud.org journalctl > $LOG_DIR/journalctl.apps.log
$SSH api.store.syncloud.org journalctl > $LOG_DIR/journalctl.store.log
$SCP -r apps.syncloud.org:$STORE_DIR $ARTIFACTS_DIR/store
$SCP apps.syncloud.org:/var/log/nginx/access.log $LOG_DIR/apps.nginx.access.log
chmod -R a+r $ARTIFACTS_DIR

exit $code
