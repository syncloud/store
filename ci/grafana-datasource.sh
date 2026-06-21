#!/bin/bash
set -ex

DIR=$( cd "$( dirname "$0" )" && pwd )
GRAFANA_HOST="${GRAFANA_HOST:-grafana:3000}"

command -v curl >/dev/null || { apt-get update >/dev/null; apt-get install -y -qq curl; }

for i in $(seq 1 60); do
    curl -fsS "http://${GRAFANA_HOST}/api/health" 2>/dev/null | grep -q '"database": *"ok"' && break
    sleep 2
done

curl -fsS -X POST "http://admin:admin@${GRAFANA_HOST}/api/datasources" \
    -H 'Content-Type: application/json' \
    -d @"$DIR/grafana/datasource.json"
echo
