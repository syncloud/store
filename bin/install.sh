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

cp ${CURRENT}/config/$ENV/apache.conf /etc/apache2/sites-available/store.conf
if [ ! -f "$STORE_DIR/secret.yaml" ]; then
  cp ${CURRENT}/config/test/secret.yaml $STORE_DIR/secret.yaml
fi
cat > $STORE_DIR/aws.env <<EOF
AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID:-GK31c4cef60f8f78b1bf12cd71}
AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY:-b8a31bf6c5d4e7a9f2b3c1d8e7f6a5b4c3d2e1f0a9b8c7d6e5f4a3b2c1d0e9f8}
AWS_S3_ENDPOINT=${AWS_S3_ENDPOINT:-http://s3}
AWS_REGION=${AWS_REGION:-garage}
EOF
chown -R store:store $STORE_DIR

systemctl enable syncloud-store.service
systemctl start syncloud-store.service

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
