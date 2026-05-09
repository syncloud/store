#!/bin/bash -ex

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

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
./wait-for-device.sh api.store.test
./wait-for-device.sh apps.syncloud.org

mkdir -p $LOG_DIR

$SCP ${DIR}/../bin/install.sh root@api.store.test:/install.sh
$SCP ${DIR}/../out/store-*.tar.gz root@api.store.test:/store.tar.gz

SNAPD2_VERSION=$(curl -fsS http://apps.syncloud.org/releases/stable/snapd2.version)
wget --progress=dot:giga http://apps.syncloud.org/apps/snapd-${SNAPD2_VERSION}-${SNAP_ARCH}.tar.gz -O snapd2.tar.gz
$SCP snapd2.tar.gz root@device:/

$SCP ${DIR}/install-snapd-v2.sh root@device:/

$SCP ${DIR}/testapp2_1_$SNAP_ARCH.snap root@device:/testapp2_1.snap
#$SCP ${DIR}/test root@$DEVICE:/

code=0
set +e
${DIR}/test -test.failfast
code=$(($code+$?))
#$SSH root@$DEVICE /test -test.run Inside
#code=$(($code+$?))
#
set -e

$SSH root@device snap changes > $LOG_DIR/snap.changes.log || true
$SSH root@device journalctl > $LOG_DIR/journalctl.device.log || true
$SSH api.store.test journalctl > $LOG_DIR/journalctl.store.log || true
$SSH api.store.test ls -la /var/www/store > $LOG_DIR/var.www.store.log || true
$SCP api.store.test:/var/log/apache2/store-access.log $LOG_DIR || true
$SCP api.store.test:/var/log/apache2/store-error.log $LOG_DIR || true
$SCP -r apps.syncloud.org:$STORE_DIR $ARTIFACTS_DIR/store || true
$SCP apps.syncloud.org:/var/log/nginx/access.log $LOG_DIR/apps.nginx.access.log || true
chmod -R a+r $ARTIFACTS_DIR || true

exit $code
