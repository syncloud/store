#!/bin/bash
set -ex

apt-get update
apt-get install -y sshpass openssh-client

KEYFILE=/tmp/_deploy_key
ssh-keygen -t ed25519 -f "$KEYFILE" -N "" -q

SSH="sshpass -p syncloud ssh -o StrictHostKeyChecking=no root@${DEPLOY_HOST}"
$SSH "mkdir -p /root/.ssh"
$SSH "cat >> /root/.ssh/authorized_keys" < "${KEYFILE}.pub"
$SSH "chmod 600 /root/.ssh/authorized_keys"
