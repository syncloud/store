#!/bin/bash -xe

SNAPD=$1

cd /tmp
rm -rf snapd

tar xzvf "${SNAPD}"
systemctl stop snapd.service snapd.socket || true
#systemctl disable snapd.service snapd.socket || true

# TODO: /usr/lib/snapd/snapd is still busy sometimes right after the stop
sleep 5

cp snapd/bin/snapd /usr/lib/snapd
cp snapd/bin/snap-exec /usr/lib/snapd
cp snapd/bin/snap-confine /usr/lib/snapd
cp snapd/bin/snap-seccomp /usr/lib/snapd
cp snapd/bin/snap-repair /usr/lib/snapd
cp snapd/bin/snap-update-ns /usr/lib/snapd
cp snapd/bin/snap-discard-ns /usr/lib/snapd
cp snapd/bin/snap /usr/bin
cp snapd/bin/snapctl /usr/bin
cp snapd/bin/mksquashfs /usr/bin
cp snapd/bin/unsquashfs /usr/bin
mkdir -p /usr/lib/snapd/lib
cp snapd/lib/* /usr/lib/snapd/lib

cp snapd/conf/snapd.service /lib/systemd/system/
cp snapd/conf/snapd.socket /lib/systemd/system/

#systemctl enable snapd.service
#systemctl enable snapd.socket
systemctl daemon-reload
systemctl start snapd.service snapd.socket

