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

$SSH $REMOTE \
    "sudo -n env \
        AWS_ACCESS_KEY_ID='${AWS_ACCESS_KEY_ID:-}' \
        AWS_SECRET_ACCESS_KEY='${AWS_SECRET_ACCESS_KEY:-}' \
        AWS_S3_ENDPOINT='${AWS_S3_ENDPOINT:-}' \
        bash /tmp/syncloud-store/deploy/deploy.sh $TAG $ENV"
