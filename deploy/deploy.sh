#!/bin/bash
set -ex

if [ "$#" -ne 2 ]; then
    echo "usage: $0 <docker-tag> <env: test|uat|prod>" >&2
    exit 1
fi

TAG=$1
ENV=$2
DIR=$( cd "$( dirname "$0" )" && pwd )
APACHE_SRC="$DIR/../config/$ENV/apache.conf"
if [ ! -f "$APACHE_SRC" ]; then
    echo "missing $APACHE_SRC" >&2
    exit 1
fi
CONTAINER=syncloud-store
STORE_DIR=/var/www/store
APACHE_SITE=/etc/apache2/sites-available/store.conf

if ! command -v docker >/dev/null 2>&1 || ! command -v apache2 >/dev/null 2>&1; then
    apt-get update
    apt-get install -y docker.io apache2
fi

if ! id -u store >/dev/null 2>&1; then
    adduser --disabled-password --gecos "" store
fi
STORE_UID=$(id -u store)
STORE_GID=$(id -g store)

mkdir -p "$STORE_DIR"
chown "$STORE_UID:$STORE_GID" "$STORE_DIR"

SECRET_SRC="$DIR/../config/$ENV/secret.yaml"
if [ -f "$SECRET_SRC" ] && [ ! -f "$STORE_DIR/secret.yaml" ]; then
    install -m 0640 -o "$STORE_UID" -g "$STORE_GID" "$SECRET_SRC" "$STORE_DIR/secret.yaml"
fi

docker pull "$TAG"
docker rm -f "$CONTAINER" 2>/dev/null || true
docker run -d \
    --name "$CONTAINER" \
    --restart=unless-stopped \
    --user "$STORE_UID:$STORE_GID" \
    --network host \
    -v "$STORE_DIR:$STORE_DIR" \
    -v /etc/hosts:/etc/hosts:ro \
    -e AWS_ACCESS_KEY_ID="${AWS_ACCESS_KEY_ID:-}" \
    -e AWS_SECRET_ACCESS_KEY="${AWS_SECRET_ACCESS_KEY:-}" \
    -e AWS_S3_ENDPOINT="${AWS_S3_ENDPOINT:-}" \
    -e AWS_REGION="${AWS_REGION:-us-west-2}" \
    "$TAG"

for i in $(seq 1 30); do
    if docker ps -q --filter name="$CONTAINER" --filter status=running | grep -q .; then
        break
    fi
    sleep 2
done
if ! docker ps -q --filter name="$CONTAINER" --filter status=running | grep -q .; then
    echo "container is not running:"
    docker ps -a --filter name="$CONTAINER"
    docker logs "$CONTAINER" 2>&1 | tail -40
    exit 1
fi

cp "$APACHE_SRC" "$APACHE_SITE"

if a2query -s 000-default >/dev/null 2>&1; then
    a2dissite 000-default
fi
a2ensite store
a2enmod proxy proxy_http rewrite ssl
apache2ctl configtest
systemctl reload apache2 || systemctl restart apache2

docker image prune -f
