#!/bin/bash
set -ex

if [ "$#" -ne 1 ]; then
    echo "usage: $0 <env>" >&2
    exit 1
fi
ENV=$1

if ! command -v curl >/dev/null; then
    apt-get update
    apt-get install -y curl python3
fi

KEYFILE=/tmp/_deploy_key
SSH="ssh -i $KEYFILE -o StrictHostKeyChecking=no"
REMOTE="${DEPLOY_USER}@${DEPLOY_HOST}"

for i in $(seq 1 60); do
    n=$(curl -fsS "${DEPLOY_URL}/api/ui/v1/apps" 2>/dev/null \
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
    $SSH $REMOTE sudo -n docker logs syncloud-store 2>&1 | tail -40
    exit 1
fi

curl -fsS "${DEPLOY_URL}/v2/snaps/find?architecture=amd64&channel=stable" \
    | python3 -c 'import json,sys; d=json.load(sys.stdin); print("api vhost results:", len(d.get("results",[])))'
