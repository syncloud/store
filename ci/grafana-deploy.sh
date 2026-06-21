#!/bin/bash
set -ex

DIR=$( cd "$( dirname "$0" )" && pwd )
KEYFILE=/tmp/_deploy_key
SSH="ssh -i $KEYFILE -o StrictHostKeyChecking=no"
SCP="scp -i $KEYFILE -o StrictHostKeyChecking=no"
REMOTE="${DEPLOY_USER}@${DEPLOY_HOST}"

$SCP "$DIR/grafana/popularity.json" "${REMOTE}:/tmp/store-popularity-dashboard.json"

$SSH $REMOTE 'sudo -n bash -s' <<'REMOTE_SCRIPT'
set -e
USER=$(awk -F= '/^[[:space:]]*admin_user[[:space:]]*=/{gsub(/^[[:space:]]+|[[:space:]]+$/, "", $2); print $2}' /etc/grafana/grafana.ini | head -1)
PASS=$(awk -F= '/^[[:space:]]*admin_password[[:space:]]*=/{gsub(/^[[:space:]]+|[[:space:]]+$/, "", $2); print $2}' /etc/grafana/grafana.ini | head -1)
DS_UID=$(curl -s -u "${USER}:${PASS}" http://127.0.0.1:3000/api/datasources \
  | python3 -c "import json,sys; print(next(d['uid'] for d in json.load(sys.stdin) if d['type']=='prometheus'))")

python3 - "${DS_UID}" <<'EOF' | curl -fsS -u "${USER}:${PASS}" -X POST -H 'Content-Type: application/json' --data @- http://127.0.0.1:3000/api/dashboards/db
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
