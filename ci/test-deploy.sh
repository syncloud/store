#!/bin/bash
# Drone CI step: validate deploy.sh end-to-end against the api.store.test
# service container. Pre-routes apps.syncloud.org -> mock service container
# via /etc/hosts so the inner DinD store binary reaches the mock instead of
# real S3, then asserts both apache vhosts proxy through to the new container.

set -ex

if [ "$#" -ne 1 ]; then
    echo "usage: $0 <docker-tag>" >&2
    exit 1
fi

TAG=$1
TARGET=api.store.test
SSH="sshpass -p syncloud ssh -o StrictHostKeyChecking=no root@${TARGET}"

apt-get update
apt-get install -y sshpass openssh-client curl python3

MOCK_IP=$(getent hosts apps.syncloud.org | awk '{print $1}' | head -1)
echo "mock apps.syncloud.org ip: $MOCK_IP"
$SSH "grep -q 'apps.syncloud.org' /etc/hosts || echo $MOCK_IP apps.syncloud.org >> /etc/hosts"

$SSH bash /tmp/syncloud-store/deploy/deploy.sh "$TAG" test

for i in $(seq 1 60); do
    n=$(curl -fsS http://${TARGET}/api/ui/v1/apps 2>/dev/null \
        | python3 -c 'import json,sys; print(len(json.load(sys.stdin)))' 2>/dev/null \
        || echo 0)
    if [ "$n" -gt 0 ]; then
        echo "OK ($n apps)"
        break
    fi
    sleep 2
done
if [ "${n:-0}" -eq 0 ]; then
    echo "store did not populate index"
    $SSH docker logs syncloud-store 2>&1 | tail -40
    exit 1
fi

echo "verifying api vhost (api.store.test) serves snap protocol"
curl -fsS "http://${TARGET}/v2/snaps/find?architecture=amd64&channel=stable" \
    | python3 -c 'import json,sys; d=json.load(sys.stdin); print("api vhost results:", len(d.get("results",[])))'

echo "verifying web vhost (store.test) reachable through same backend"
curl -fsS -H 'Host: store.test' http://${TARGET}/api/ui/v1/apps \
    | python3 -c 'import json,sys; print("web vhost ui apps:", len(json.load(sys.stdin)))'
