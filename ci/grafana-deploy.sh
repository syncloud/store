#!/bin/bash
set -ex

DIR=$( cd "$( dirname "$0" )" && pwd )
KEYFILE=/tmp/_deploy_key
SSH="ssh -i $KEYFILE -o StrictHostKeyChecking=no"
SCP="scp -i $KEYFILE -o StrictHostKeyChecking=no"
REMOTE="${DEPLOY_USER}@${DEPLOY_HOST}"
GRAFANA_HOST="${GRAFANA_HOST:-127.0.0.1:3000}"

$SCP "$DIR/grafana/popularity.json" "${REMOTE}:/tmp/store-popularity-dashboard.json"

$SSH $REMOTE "sudo -n bash -s '$GRAFANA_HOST'" <<'REMOTE_SCRIPT'
set -e
GRAFANA_HOST="$1"
USER=admin
PASS=admin
INI=/etc/grafana/grafana.ini
if [ -f "$INI" ]; then
    USER=$(awk -F= '/^[[:space:]]*admin_user[[:space:]]*=/{gsub(/^[[:space:]]+|[[:space:]]+$/, "", $2); print $2}' "$INI" | head -1)
    PASS=$(awk -F= '/^[[:space:]]*admin_password[[:space:]]*=/{gsub(/^[[:space:]]+|[[:space:]]+$/, "", $2); print $2}' "$INI" | head -1)
fi

for i in $(seq 1 60); do
    curl -fsS "http://${GRAFANA_HOST}/api/health" 2>/dev/null | grep -q '"database": *"ok"' && break
    sleep 2
done

DS_UID=$(curl -fsS -u "${USER}:${PASS}" "http://${GRAFANA_HOST}/api/datasources" \
    | python3 -c "import json,sys; print(next(d['uid'] for d in json.load(sys.stdin) if d['type']=='prometheus'))")

python3 - "$DS_UID" <<'EOF' | curl -fsS -u "${USER}:${PASS}" -X POST -H 'Content-Type: application/json' --data @- "http://${GRAFANA_HOST}/api/dashboards/db"
import json, sys
ds_uid = sys.argv[1]
with open('/tmp/store-popularity-dashboard.json') as f:
    raw = f.read()
raw = raw.replace('${DS_PROMETHEUS}', ds_uid)
d = json.loads(raw)
d.pop('__inputs', None)
d.pop('id', None)
print(json.dumps({'dashboard': d, 'overwrite': True, 'folderId': 0, 'message': 'CI auto-deploy'}))
EOF
echo
REMOTE_SCRIPT
