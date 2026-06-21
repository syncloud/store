#!/bin/bash
set -ex

GRAFANA=http://admin:admin@grafana:3000
DIR=$( cd "$( dirname "$0" )" && pwd )

apt-get update >/dev/null
apt-get install -y -qq curl python3

for i in $(seq 1 60); do
    if curl -fsS http://grafana:3000/api/health 2>/dev/null | grep -q '"database": *"ok"'; then
        echo grafana up
        break
    fi
    sleep 2
done

curl -fsS -X POST "$GRAFANA/api/datasources" \
    -H 'Content-Type: application/json' \
    -d @"$DIR/grafana/datasource.json"
echo

python3 -c "
import json
raw = open('$DIR/grafana/popularity.json').read().replace('\${DS_PROMETHEUS}', 'victoria-metrics')
d = json.loads(raw)
d['id'] = None
print(json.dumps({'dashboard': d, 'overwrite': True}))
" | curl -fsS -X POST "$GRAFANA/api/dashboards/db" \
    -H 'Content-Type: application/json' \
    --data-binary @-
echo
