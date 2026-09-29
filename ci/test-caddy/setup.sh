#!/bin/bash
set -ex

# overlay2 cannot mount on the container's own overlay rootfs, and docker 20.10
# then falls back to devicemapper, whose pool creation depends on host
# device-mapper state and fails intermittently. vfs is the only driver that
# works here unconditionally, and the step runs two containers, so it is cheap.
install -d /etc/docker
cat > /etc/docker/daemon.json <<'JSON'
{
    "storage-driver": "vfs"
}
JSON

if ! command -v docker >/dev/null 2>&1; then
    apt-get update
    apt-get install -y docker.io
fi

systemctl start docker || true
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
