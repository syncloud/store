#!/bin/bash
set -ex

if [ "$#" -ne 1 ]; then
    echo "usage: $0 <env>" >&2
    exit 1
fi
ENV=$1

if ! command -v ssh >/dev/null; then
    apt-get update
    apt-get install -y openssh-client ca-certificates
fi

KEYFILE=/tmp/_deploy_key
set +x
printf '%s\n' "$DEPLOY_KEY" > "$KEYFILE"
set -x
chmod 600 "$KEYFILE"

SSH="ssh -i $KEYFILE -o StrictHostKeyChecking=no"
SCP="scp -i $KEYFILE -o StrictHostKeyChecking=no -r"
REMOTE="${DEPLOY_USER}@${DEPLOY_HOST}"

STAGE=$(mktemp -d)
trap 'rm -rf "$STAGE"' EXIT
cp -r "config/${ENV}/." "$STAGE/"

[ -n "$SYNCLOUD_TOKEN" ] || { echo "SYNCLOUD_TOKEN required" >&2; exit 1; }
set +x
sed -i "s|@token@|${SYNCLOUD_TOKEN}|g" "$STAGE/secret.yaml"
set -x

[ -n "$AWS_ACCESS_KEY_ID" ] || { echo "AWS_ACCESS_KEY_ID required" >&2; exit 1; }
set +x
sed -i "s|@aws_access_key_id@|${AWS_ACCESS_KEY_ID}|g" "$STAGE/secret.yaml"
set -x

[ -n "$AWS_SECRET_ACCESS_KEY" ] || { echo "AWS_SECRET_ACCESS_KEY required" >&2; exit 1; }
set +x
sed -i "s|@aws_secret_access_key@|${AWS_SECRET_ACCESS_KEY}|g" "$STAGE/secret.yaml"
set -x

$SSH $REMOTE "sudo -n rm -rf /tmp/syncloud-store && mkdir -p /tmp/syncloud-store/config/${ENV}"
$SCP deploy "${REMOTE}:/tmp/syncloud-store/"
$SCP "$STAGE/." "${REMOTE}:/tmp/syncloud-store/config/${ENV}/"
