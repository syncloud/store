#!/bin/bash -ex

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

SCP="sshpass -p syncloud scp -o StrictHostKeyChecking=no"
SSH="sshpass -p syncloud ssh -o StrictHostKeyChecking=no"
ARTIFACTS_DIR=${DIR}/artifacts
LOG_DIR=$ARTIFACTS_DIR/log
SNAP_ARCH=$(dpkg --print-architecture)

apt update
apt install -y sshpass curl wget
cd $DIR

./wait-for-device.sh device
./wait-for-device.sh api.store.test
for i in $(seq 60); do
    curl -s -o /dev/null --max-time 2 http://apps/ && break
    sleep 1
done

mkdir -p $LOG_DIR

$SSH root@api.store.test "rm -rf /tmp/syncloud-store && mkdir -p /tmp/syncloud-store/config"
$SCP -r ${DIR}/../deploy root@api.store.test:/tmp/syncloud-store/
$SCP -r ${DIR}/../config/test root@api.store.test:/tmp/syncloud-store/config/

wget --progress=dot:giga https://github.com/syncloud/snapd/releases/download/syncloud-5/snapd-640-${SNAP_ARCH}.tar.gz -O snapd2.tar.gz
$SCP snapd2.tar.gz root@device:/

$SCP ${DIR}/install-snapd-v2.sh root@device:/

$SCP ${DIR}/testapp2_1_$SNAP_ARCH.snap root@device:/testapp2_1.snap

code=0
set +e
${DIR}/test -test.failfast
code=$(($code+$?))
set -e

$SSH root@device snap changes > $LOG_DIR/snap.changes.log || true
$SSH root@device journalctl > $LOG_DIR/journalctl.device.log || true
$SSH api.store.test journalctl > $LOG_DIR/journalctl.store.log || true
$SSH api.store.test 'docker logs syncloud-store 2>&1' > $LOG_DIR/docker.store.log || true
$SSH api.store.test ls -la /var/www/store > $LOG_DIR/var.www.store.log || true
chmod -R a+r $ARTIFACTS_DIR || true

exit $code
