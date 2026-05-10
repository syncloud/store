#!/bin/bash
set -ex

if [ "$#" -ne 1 ]; then
    echo "usage: $0 <env>" >&2
    exit 1
fi
ENV=$1

if ! command -v ssh >/dev/null; then
    apt-get update
    apt-get install -y openssh-client
fi

KEYFILE=/tmp/_deploy_key
if [ ! -f "$KEYFILE" ]; then
    set +x
    printf '%s\n' "$DEPLOY_KEY" > "$KEYFILE"
    set -x
    chmod 600 "$KEYFILE"
fi

SSH="ssh -i $KEYFILE -o StrictHostKeyChecking=no"
SCP="scp -i $KEYFILE -o StrictHostKeyChecking=no -r"
REMOTE="${DEPLOY_USER}@${DEPLOY_HOST}"

$SSH $REMOTE "rm -rf /tmp/syncloud-store && mkdir -p /tmp/syncloud-store/config"
$SCP deploy "${REMOTE}:/tmp/syncloud-store/"
$SCP "config/${ENV}" "${REMOTE}:/tmp/syncloud-store/config/"
