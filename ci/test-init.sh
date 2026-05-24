#!/bin/bash
set -ex

apt-get update
apt-get install -y sshpass openssh-client curl

KEYFILE=/tmp/_deploy_key
ssh-keygen -t ed25519 -f $KEYFILE -N "" -q
PUB=$(cat ${KEYFILE}.pub)

TARGET=${DEPLOY_HOST}
sshpass -p syncloud ssh -o StrictHostKeyChecking=no root@$TARGET \
    "mkdir -p /root/.ssh && echo '$PUB' >> /root/.ssh/authorized_keys && chmod 600 /root/.ssh/authorized_keys"
