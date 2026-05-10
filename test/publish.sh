#!/bin/bash -ex

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

if [ "$#" -lt 1 ]; then
    echo "usage $0 arch"
    exit 1
fi

SCP="sshpass -p syncloud scp -o StrictHostKeyChecking=no"
SSH="sshpass -p syncloud ssh -o StrictHostKeyChecking=no"
ARCH=$1
SNAP_ARCH=$(dpkg --print-architecture)
LOG_DIR=${DIR}/../../log

apt update
apt install -y sshpass curl
cd $DIR

./wait-for-device.sh apps.syncloud.org

mkdir -p $LOG_DIR
STORE_DIR=/var/www/html

$SSH root@apps.syncloud.org apt update
$SSH root@apps.syncloud.org apt install -y nginx tree
$SSH root@apps.syncloud.org mkdir -p $STORE_DIR/releases/master/images
$SSH root@apps.syncloud.org mkdir -p $STORE_DIR/releases/rc/images
$SSH root@apps.syncloud.org mkdir -p $STORE_DIR/releases/stable/images
$SSH root@apps.syncloud.org mkdir -p $STORE_DIR/apps
$SSH root@apps.syncloud.org mkdir -p $STORE_DIR/revisions
$SCP ${DIR}/../out/syncloud-release-$ARCH root@apps.syncloud.org:/syncloud-release

$SCP ${DIR}/testapp*.snap root@apps.syncloud.org:/

$SCP ${DIR}/index-v2 root@apps.syncloud.org:$STORE_DIR/releases/master
$SCP ${DIR}/index-v2 root@apps.syncloud.org:$STORE_DIR/releases/rc
$SCP ${DIR}/index-v2 root@apps.syncloud.org:$STORE_DIR/releases/stable
$SCP ${DIR}/images/*.png root@apps.syncloud.org:$STORE_DIR/releases/master/images/
$SCP ${DIR}/images/*.png root@apps.syncloud.org:$STORE_DIR/releases/rc/images/
$SCP ${DIR}/images/*.png root@apps.syncloud.org:$STORE_DIR/releases/stable/images/
$SSH root@apps.syncloud.org tree $STORE_DIR > $LOG_DIR/store.tree.log
$SSH root@apps.syncloud.org systemctl status nginx > $LOG_DIR/nginx.status.log
