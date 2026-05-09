#!/bin/bash
# Drone CI step: scp the deploy bundle (deploy/ + config/<env>/) to the
# api.store.test service container. Mirrors the drone-scp step used for
# real UAT/prod hosts but uses sshpass since the test container uses
# password auth.

set -ex

if [ "$#" -ne 1 ]; then
    echo "usage: $0 <env: test>" >&2
    exit 1
fi

ENV=$1
TARGET=api.store.test
SSH="sshpass -p syncloud ssh -o StrictHostKeyChecking=no root@${TARGET}"
SCP="sshpass -p syncloud scp -r -o StrictHostKeyChecking=no"

apt-get update
apt-get install -y sshpass openssh-client

$SSH 'rm -rf /tmp/syncloud-store && mkdir -p /tmp/syncloud-store/config'
$SCP deploy "root@${TARGET}:/tmp/syncloud-store/"
$SCP "config/${ENV}" "root@${TARGET}:/tmp/syncloud-store/config/"
