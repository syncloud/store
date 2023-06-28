#!/bin/bash -ex

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

if [ "$#" -lt 1 ]; then
    echo "usage $0 device"
    exit 1
fi

DEVICE=$1
STORE_DIR=/var/www/html
SCP="sshpass -p syncloud scp -o StrictHostKeyChecking=no"
SSH="sshpass -p syncloud ssh -o StrictHostKeyChecking=no"
ARTIFACTS_DIR=${DIR}/../../artifacts
mkdir $ARTIFACTS_DIR
LOG_DIR=$ARTIFACTS_DIR/log/$DEVICE
SNAP_ARCH=$(dpkg --print-architecture)

apt update
apt install -y sshpass curl
cd $DIR

./wait-for-device.sh $DEVICE

mkdir -p $LOG_DIR

$SCP ${DIR}/install-store.sh root@$DEVICE:/install-store.sh
$SCP ${DIR}/../out/syncloud-store-*.tar.gz root@$DEVICE:/syncloud-store.tar.gz

$SCP ${DIR}/install-snapd.sh root@$DEVICE:/install-snapd.sh
$SCP ${DIR}/../../snapd-*.tar.gz root@$DEVICE:/snapd.tar.gz

$SSH root@$DEVICE /install-store.sh
$SSH root@$DEVICE /install-snapd.sh
$SCP ${DIR}/testapp2_1_$SNAP_ARCH.snap root@$DEVICE:/testapp2_1.snap
$SCP ${DIR}/test root@$DEVICE:/

code=0
set +e

${DIR}/test -test.run Outside
code=$(($code+$?))

$SSH root@$DEVICE /test -test.run Inside
code=$(($code+$?))

set -e

$SSH root@$DEVICE snap changes > $LOG_DIR/snap.changes.log || true
$SSH root@$DEVICE journalctl > $LOG_DIR/journalctl.device.log
$SSH apps.syncloud.org journalctl > $LOG_DIR/journalctl.store.log
$SCP -r apps.syncloud.org:$STORE_DIR $ARTIFACTS_DIR/store
$SCP apps.syncloud.org:/var/log/nginx/access.log $ARTIFACTS_DIR
chmod -R a+r $ARTIFACTS_DIR

exit $code
