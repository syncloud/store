#!/bin/bash -ex

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

apt update
apt install -y sshpass curl wget openssh-client

SCP="sshpass -p syncloud scp -o StrictHostKeyChecking=no"
SSH="sshpass -p syncloud ssh -o StrictHostKeyChecking=no"
ARTIFACTS_DIR=${DIR}/artifacts
LOG_DIR=$ARTIFACTS_DIR/log
SNAP_ARCH=$(dpkg --print-architecture)
mkdir -p $LOG_DIR

cd $DIR
./wait-for-device.sh device
./wait-for-device.sh api.store.test
for i in $(seq 60); do
    curl -s -o /dev/null --max-time 2 http://apps/ && break
    sleep 1
done

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

if [ $code -eq 0 ]; then
    DEPLOY_URL="http://api.store.test" DEPLOY_HOST=api.store.test DEPLOY_USER=root \
        SSH_PASSWORD=syncloud ${DIR}/../ci/deploy-verify.sh test
    code=$?
fi

$SSH root@device snap changes > $LOG_DIR/snap.changes.log 2>&1 || true
$SSH root@device journalctl > $LOG_DIR/journalctl.device.log 2>&1 || true
$SSH api.store.test journalctl > $LOG_DIR/journalctl.store.log 2>&1 || true
$SSH api.store.test 'docker logs syncloud-store 2>&1' > $LOG_DIR/docker.store.log 2>&1 || true
$SSH api.store.test ls -la /var/www/store > $LOG_DIR/var.www.store.log 2>&1 || true
chmod -R a+r $ARTIFACTS_DIR || true

exit $code
