#!/bin/bash
set -ex

if [ "$#" -ne 1 ]; then
    echo "usage: $0 <env>" >&2
    exit 1
fi
ENV=$1

apt-get update
apt-get install -y curl python3 sshpass openssh-client

KEYFILE=/tmp/_deploy_key
if [ -f "$KEYFILE" ]; then
    SSH="ssh -i $KEYFILE -o StrictHostKeyChecking=no"
elif [ -n "${SSH_PASSWORD:-}" ]; then
    SSH="sshpass -p $SSH_PASSWORD ssh -o StrictHostKeyChecking=no"
else
    echo "no ssh credentials available (neither $KEYFILE nor SSH_PASSWORD)" >&2
    exit 1
fi
REMOTE="${DEPLOY_USER}@${DEPLOY_HOST}"

fail_with_logs() {
    echo "$1" >&2
    $SSH $REMOTE 'sudo -n docker ps -a' 2>&1 || true
    $SSH $REMOTE 'sudo -n docker logs syncloud-store 2>&1' || true
    exit 1
}

for i in $(seq 1 60); do
    ver_code=$(curl -k -s -o /dev/null -w "%{http_code}" "${DEPLOY_URL}/api/ui/v1/version" || echo 000)
    if [ "$ver_code" = "200" ]; then break; fi
    sleep 2
done
[ "$ver_code" = "200" ] || fail_with_logs "/api/ui/v1/version did not return 200: $ver_code"
echo "version endpoint OK ($ver_code)"

set +x
refresh_body=$(curl -k -s -w "\n%{http_code}" -X POST "${DEPLOY_URL}/syncloud/v1/cache/refresh" \
    -H "Content-Type: application/json" \
    --data "{\"token\":\"${SYNCLOUD_TOKEN}\"}" \
    --max-time 300)
set -x
refresh_code=$(echo "$refresh_body" | tail -n1)
if [ "$refresh_code" != "200" ]; then
    echo "cache refresh response:" >&2
    echo "$refresh_body" | head -n-1 >&2
    fail_with_logs "/syncloud/v1/cache/refresh returned $refresh_code (token or aws creds wrong?)"
fi
echo "cache refresh OK ($refresh_code) — token and aws creds validated"

apps_body=$(curl -fsS "${DEPLOY_URL}/api/ui/v1/apps?channel=stable" 2>/dev/null || echo "")
n=$(echo "$apps_body" | python3 -c 'import json,sys; print(len(json.load(sys.stdin)))' 2>/dev/null || echo 0)
if [ "${n:-0}" -eq 0 ]; then
    echo "last apps response body:" >&2
    echo "$apps_body" >&2
    fail_with_logs "store did not populate index"
fi
echo "apps OK ($n apps)"

find_results=$(curl -fsS "${DEPLOY_URL}/v2/snaps/find?architecture=amd64&channel=stable" \
    | python3 -c 'import json,sys; d=json.load(sys.stdin); print(len(d.get("results",[])))' || echo 0)
[ "$find_results" -gt 0 ] || fail_with_logs "/v2/snaps/find returned no results"
echo "snaps/find OK ($find_results results)"

web_code=$(curl -k -s -o /dev/null -w "%{http_code}" "${DEPLOY_URL}/")
[ "$web_code" = "200" ] || fail_with_logs "web UI / did not return 200: $web_code"
echo "web UI OK ($web_code)"
