#!/bin/bash
set -ex

if [ "$#" -ne 1 ]; then
    echo "usage: $0 <docker-tag>" >&2
    exit 1
fi

TAG=$1
CONTAINER=syncloud-store
STORE_DIR=/var/www/store
SERVICE=syncloud-store.service

if ! command -v docker >/dev/null 2>&1; then
    apt-get update
    apt-get install -y docker.io
fi

if systemctl is-active --quiet "$SERVICE"; then
    systemctl stop "$SERVICE"
fi
if systemctl is-enabled --quiet "$SERVICE" 2>/dev/null; then
    systemctl disable "$SERVICE"
fi

if ! id -u store >/dev/null 2>&1; then
    adduser --disabled-password --gecos "" store
fi
STORE_UID=$(id -u store)
STORE_GID=$(id -g store)

mkdir -p "$STORE_DIR"
chown "$STORE_UID:$STORE_GID" "$STORE_DIR"

docker pull "$TAG"
docker rm -f "$CONTAINER" 2>/dev/null || true

APPS_IP=$(getent hosts apps.syncloud.org | awk '{print $1}' | head -1)
if [ -z "$APPS_IP" ]; then
    echo "could not resolve apps.syncloud.org from host" >&2
    exit 1
fi

docker run -d \
    --name "$CONTAINER" \
    --restart=unless-stopped \
    --add-host "apps.syncloud.org:$APPS_IP" \
    --user "$STORE_UID:$STORE_GID" \
    -v "$STORE_DIR:$STORE_DIR" \
    "$TAG"

sleep 3
if ! docker ps -q --filter name="$CONTAINER" --filter status=running | grep -q .; then
    echo "container is not running:"
    docker ps -a --filter name="$CONTAINER"
    docker logs "$CONTAINER" 2>&1 | tail -40
    exit 1
fi

docker image prune -f
