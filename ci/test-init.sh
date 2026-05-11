#!/bin/bash
set -ex

apt-get update
apt-get install -y sshpass openssh-client

KEYFILE=/tmp/_deploy_key
ssh-keygen -t ed25519 -f $KEYFILE -N "" -q
PUB=$(cat ${KEYFILE}.pub)

TARGET=api.store.test
sshpass -p syncloud ssh -o StrictHostKeyChecking=no root@$TARGET \
    "mkdir -p /root/.ssh && echo '$PUB' >> /root/.ssh/authorized_keys && chmod 600 /root/.ssh/authorized_keys"

MOCK_IP=$(getent hosts apps.syncloud.org | awk '{print $1}' | head -1)
echo "mock apps.syncloud.org ip: $MOCK_IP"
ssh -i $KEYFILE -o StrictHostKeyChecking=no root@$TARGET \
    "grep -q 'apps.syncloud.org' /etc/hosts || echo $MOCK_IP apps.syncloud.org >> /etc/hosts"
