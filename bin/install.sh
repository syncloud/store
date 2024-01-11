#!/bin/bash -xe

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

STORE=$1
VERSION=$2
ENV=$3

STORE_DIR=/var/www/store
CURRENT=$STORE_DIR/current

systemctl stop syncloud-store.service || true
systemctl disable syncloud-store.service || true

rm -rf ${CURRENT}
mkdir -p $STORE_DIR/$VERSION
tar xzvf $STORE -C $STORE_DIR/$VERSION
ln -s $STORE_DIR/$VERSION ${CURRENT}
cp ${CURRENT}/config/syncloud-store.service /lib/systemd/system/

if  ! id -u store > /dev/null 2>&1; then
    adduser --disabled-password --gecos "" store
fi

systemctl enable syncloud-store.service
systemctl start syncloud-store.service

cp ${CURRENT}/config/$ENV/apache.conf /etc/apache2/sites-available/store.conf
if ! -f $STORE_DIR/secret.yaml; then
  cp ${CURRENT}/config/test/secret.yaml $STORE_DIR/secret.yaml
fi

chown -R store.store $STORE_DIR
if a2query -s 000-default; then
  a2dissite 000-default
fi
if ! a2query -s store; then
  a2ensite store
fi
a2enmod rewrite
a2enmod ssl
a2enmod proxy
a2enmod proxy_http

service apache2 restart
