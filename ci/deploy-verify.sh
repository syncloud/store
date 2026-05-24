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
if [ -f "$KEYFILE" ]; then
    SSH="ssh -i $KEYFILE -o StrictHostKeyChecking=no"
elif [ -n "${SSH_PASSWORD:-}" ]; then
    if ! command -v sshpass >/dev/null; then
        apt-get update
        apt-get install -y sshpass openssh-client
    fi
    SSH="sshpass -p $SSH_PASSWORD ssh -o StrictHostKeyChecking=no"
else
    echo "no ssh credentials available (neither $KEYFILE nor SSH_PASSWORD)" >&2
    exit 1
fi
REMOTE="${DEPLOY_USER}@${DEPLOY_HOST}"

fail_with_logs() {
    echo "$1" >&2
    $SSH $REMOTE 'docker ps -a' 2>&1 || true
    $SSH $REMOTE 'docker logs syncloud-store 2>&1' || true
    exit 1
}

apps_body=""
for i in $(seq 1 60); do
    apps_body=$(curl -fsS "${DEPLOY_URL}/api/ui/v1/apps?channel=stable" 2>/dev/null || echo "")
    n=$(echo "$apps_body" | python3 -c 'import json,sys; print(len(json.load(sys.stdin)))' 2>/dev/null || echo 0)
    if [ "$n" -gt 0 ]; then
        echo "apps OK ($n apps)"
        break
    fi
    sleep 2
done
if [ "${n:-0}" -eq 0 ]; then
    echo "last apps response body:" >&2
    echo "$apps_body" >&2
    fail_with_logs "store did not populate index"
fi

find_results=$(curl -fsS "${DEPLOY_URL}/v2/snaps/find?architecture=amd64&channel=stable" \
    | python3 -c 'import json,sys; d=json.load(sys.stdin); print(len(d.get("results",[])))' || echo 0)
[ "$find_results" -gt 0 ] || fail_with_logs "/v2/snaps/find returned no results"
echo "snaps/find OK ($find_results results)"

web_code=$(curl -k -s -o /dev/null -w "%{http_code}" "${DEPLOY_URL}/")
[ "$web_code" = "200" ] || fail_with_logs "web UI / did not return 200: $web_code"
echo "web UI OK ($web_code)"

ver_code=$(curl -k -s -o /dev/null -w "%{http_code}" "${DEPLOY_URL}/api/ui/v1/version")
[ "$ver_code" = "200" ] || fail_with_logs "/api/ui/v1/version did not return 200: $ver_code"
echo "version endpoint OK ($ver_code)"
