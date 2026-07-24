#!/bin/bash
set -ex

KEYFILE=/tmp/_deploy_key
SSH="ssh -i $KEYFILE -o StrictHostKeyChecking=no"
SCP="scp -i $KEYFILE -o StrictHostKeyChecking=no"
REMOTE="${DEPLOY_USER}@${DEPLOY_HOST}"

$SCP ci/test-caddy/Caddyfile "${REMOTE}:/tmp/test-caddy-Caddyfile"
$SCP ci/test-caddy/setup.sh "${REMOTE}:/tmp/test-caddy-setup.sh"
$SSH $REMOTE "sudo -n bash /tmp/test-caddy-setup.sh"
