#!/bin/bash
set -ex

KEYFILE=/tmp/_deploy_key
SSH="ssh -i $KEYFILE -o StrictHostKeyChecking=no"
SCP="scp -i $KEYFILE -o StrictHostKeyChecking=no"
REMOTE="${DEPLOY_USER}@${DEPLOY_HOST}"

$SCP ci/test-caddy/Caddyfile "${REMOTE}:/tmp/test-caddy-Caddyfile"

$SSH $REMOTE sudo -n bash -s <<'REMOTE_SCRIPT'
set -ex
if ! command -v docker >/dev/null 2>&1; then
    apt-get update
    apt-get install -y docker.io
fi

install -d /etc/caddy /etc/caddy/conf.d
install -m 0644 /tmp/test-caddy-Caddyfile /etc/caddy/Caddyfile

docker rm -f caddy 2>/dev/null || true
docker run -d \
    --name caddy \
    --restart=unless-stopped \
    --network=host \
    -e STORE_DOMAIN=http://store \
    -e STORE_API_DOMAIN=http://api.store \
    -v /etc/caddy:/etc/caddy:ro \
    -v /var/www:/var/www \
    caddy:2.10
REMOTE_SCRIPT
