#!/bin/bash
set -ex

for _ in $(seq 1 60); do
    case "$(systemctl is-system-running 2>/dev/null)" in
        running | degraded) break ;;
    esac
    sleep 1
done
systemctl is-system-running 2>/dev/null | grep -qE '^(running|degraded)$'

if ! command -v docker >/dev/null 2>&1; then
    apt-get update
    apt-get install -y docker.io
fi

systemctl start docker

for _ in $(seq 1 60); do
    docker info >/dev/null 2>&1 && break
    sleep 1
done
docker info >/dev/null

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
