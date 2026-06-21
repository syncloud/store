#!/bin/bash
set -e

if [ "$#" -ne 2 ]; then
    echo "usage: $0 <grafana-host> <dashboard-file>" >&2
    exit 1
fi
GRAFANA_HOST=$1
DASHBOARD=$2

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

python3 - "$DS_UID" "$DASHBOARD" <<'EOF' | curl -fsS -u "${USER}:${PASS}" -X POST -H 'Content-Type: application/json' --data @- "http://${GRAFANA_HOST}/api/dashboards/db"
import json, sys
ds_uid = sys.argv[1]
with open(sys.argv[2]) as f:
    raw = f.read()
raw = raw.replace('${DS_PROMETHEUS}', ds_uid)
d = json.loads(raw)
d.pop('__inputs', None)
d.pop('id', None)
print(json.dumps({'dashboard': d, 'overwrite': True, 'folderId': 0, 'message': 'CI auto-deploy'}))
EOF
echo
