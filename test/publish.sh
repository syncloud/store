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

function wait_for_host() {
  local host=$1
  attempts=100
  attempt=0
  set +e
  sshpass -p syncloud ssh -o StrictHostKeyChecking=no root@$host date
  while test $? -gt 0
  do
    if [ $attempt -gt $attempts ]; then
      exit 1
    fi
    sleep 3
    echo "Waiting for SSH $attempt"
    attempt=$((attempt+1))
    sshpass -p syncloud ssh -o StrictHostKeyChecking=no root@$host date
  done
  set -e
}

wait_for_host apps.syncloud.org

mkdir -p $LOG_DIR
STORE_DIR=/var/www/html

$SSH root@apps.syncloud.org apt update
$SSH root@apps.syncloud.org apt install -y nginx tree
$SSH root@apps.syncloud.org mkdir -p $STORE_DIR/releases/master
$SSH root@apps.syncloud.org mkdir -p $STORE_DIR/releases/rc
$SSH root@apps.syncloud.org mkdir -p $STORE_DIR/releases/stable
$SSH root@apps.syncloud.org mkdir -p $STORE_DIR/apps
$SSH root@apps.syncloud.org mkdir -p $STORE_DIR/revisions
$SCP ${DIR}/../out/syncloud-release-$ARCH root@apps.syncloud.org:/syncloud-release
$SCP ${DIR}/testapp1_1_$SNAP_ARCH.snap root@apps.syncloud.org:/
$SCP ${DIR}/testapp1_2_$SNAP_ARCH.snap root@apps.syncloud.org:/
$SCP ${DIR}/testapp1_3_$SNAP_ARCH.snap root@apps.syncloud.org:/
$SCP ${DIR}/testapp2_1_$SNAP_ARCH.snap root@apps.syncloud.org:/
$SCP ${DIR}/testapp2_2_$SNAP_ARCH.snap root@apps.syncloud.org:/

$SSH root@apps.syncloud.org /syncloud-release publish -f /testapp1_1_$SNAP_ARCH.snap -b stable -t $STORE_DIR
$SSH root@apps.syncloud.org /syncloud-release promote -n testapp1 -a $SNAP_ARCH -t $STORE_DIR

$SSH root@apps.syncloud.org /syncloud-release publish -f /testapp2_1_$SNAP_ARCH.snap -b master -t $STORE_DIR
#$SSH root@apps.syncloud.org /syncloud-release promote -n testapp2 -a $SNAP_ARCH -t $STORE_DIR

$SCP ${DIR}/index-v2 root@apps.syncloud.org:$STORE_DIR/releases/master
$SCP ${DIR}/index-v2 root@apps.syncloud.org:$STORE_DIR/releases/rc
$SCP ${DIR}/index-v2 root@apps.syncloud.org:$STORE_DIR/releases/stable
$SSH root@apps.syncloud.org tree $STORE_DIR > $LOG_DIR/store.tree.log
$SSH root@apps.syncloud.org systemctl status nginx > $LOG_DIR/nginx.status.log
