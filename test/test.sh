#!/bin/bash -ex

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

apt update
apt install -y sshpass curl wget

SCP="sshpass -p syncloud scp -o StrictHostKeyChecking=no"
SSH="sshpass -p syncloud ssh -o StrictHostKeyChecking=no"
ARTIFACTS_DIR=${DIR}/artifacts
LOG_DIR=$ARTIFACTS_DIR/log
SNAP_ARCH=$(dpkg --print-architecture)
mkdir -p $LOG_DIR

cd $DIR
./wait-for-device.sh device

wget --progress=dot:giga https://github.com/syncloud/snapd/releases/download/syncloud-5/snapd-640-${SNAP_ARCH}.tar.gz -O snapd2.tar.gz
$SCP snapd2.tar.gz root@device:/
$SCP ${DIR}/install-snapd-v2.sh root@device:/
$SCP ${DIR}/testapp2/testapp2_1_$SNAP_ARCH.snap root@device:/testapp2_1.snap

code=0
set +e
${DIR}/test -test.failfast
code=$(($code+$?))
set -e

$SSH root@device snap changes > $LOG_DIR/snap.changes.log 2>&1 || true
$SSH root@device journalctl > $LOG_DIR/journalctl.device.log 2>&1 || true
$SSH api.store journalctl > $LOG_DIR/journalctl.store.log 2>&1 || true
$SSH api.store 'docker logs syncloud-store 2>&1' > $LOG_DIR/docker.store.log 2>&1 || true
$SSH api.store ls -la /var/www/store > $LOG_DIR/var.www.store.log 2>&1 || true
chmod -R a+r $ARTIFACTS_DIR || true

exit $code
