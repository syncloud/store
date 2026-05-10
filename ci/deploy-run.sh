#!/bin/bash
set -ex

if [ "$#" -ne 2 ]; then
    echo "usage: $0 <env> <docker-tag>" >&2
    exit 1
fi
ENV=$1
TAG=$2

KEYFILE=/tmp/_deploy_key
SSH="ssh -i $KEYFILE -o StrictHostKeyChecking=no"
REMOTE="${DEPLOY_USER}@${DEPLOY_HOST}"

$SSH $REMOTE "bash /tmp/syncloud-store/deploy/deploy.sh $TAG $ENV"
