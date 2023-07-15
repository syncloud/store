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
./wait-for-device.sh api.store.syncloud.org
./wait-for-device.sh apps.syncloud.org

mkdir -p $LOG_DIR

$SCP ${DIR}/../bin/install.sh root@api.store.syncloud.org:/install.sh
$SCP ${DIR}/../out/store-*.tar.gz root@api.store.syncloud.org:/store.tar.gz

#VERSION_V1=$(curl http://apps.syncloud.org/releases/stable/snapd.version)
wget --progress=dot:giga https://github.com/syncloud/snapd/releases/download/1.2/snapd-361-$SNAP_ARCH.tar.gz -O snapd1.tar.gz
wget http://apps.syncloud.org/apps/"${VERSION_V1}" --progress=dot:giga
$SCP snapd1.tar.gz root@device:/

#VERSION_V2=$(curl http://apps.syncloud.org/releases/stable/snapd2.version)
wget --progress=dot:giga https://github.com/syncloud/snapd/releases/download/syncloud-2.0/snapd-543-$SNAP_ARCH.tar.gz -O snapd2.tar.gz
wget http://apps.syncloud.org/apps/"${VERSION_V2}" --progress=dot:giga
$SCP snapd2.tar.gz root@device:/

$SCP ${DIR}/install-snapd-v1.sh root@device:/
$SCP ${DIR}/install-snapd-v2.sh root@device:/
$SCP ${DIR}/upgrade-snapd.sh root@device:/

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
$SCP api.store.syncloud.org:/var/log/apache2/store-access.log $LOG_DIR
$SCP api.store.syncloud.org:/var/log/apache2/store-error.log $LOG_DIR
$SSH api.store.syncloud.org journalctl > $LOG_DIR/journalctl.store.log
$SSH api.store.syncloud.org ls -la /var/www/store > $LOG_DIR/var.www.store.log
$SCP -r apps.syncloud.org:$STORE_DIR $ARTIFACTS_DIR/store
$SCP apps.syncloud.org:/var/log/nginx/access.log $LOG_DIR/apps.nginx.access.log
chmod -R a+r $ARTIFACTS_DIR

exit $code
