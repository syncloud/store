#!/bin/bash
set -ex

if [ "$#" -ne 1 ]; then
    echo "usage: $0 <grafana-host>" >&2
    exit 1
fi
GRAFANA_HOST=$1

DIR=$( cd "$( dirname "$0" )" && pwd )
ROOT=$( cd "$DIR/.." && pwd )
KEYFILE=/tmp/_deploy_key
SSH="ssh -i $KEYFILE -o StrictHostKeyChecking=no"
SCP="scp -p -i $KEYFILE -o StrictHostKeyChecking=no"
REMOTE="${DEPLOY_USER}@${DEPLOY_HOST}"

$SCP "$ROOT/build/bin/grafana-deploy" "${REMOTE}:/tmp/store-grafana-deploy"
$SCP "$DIR/grafana/popularity.json" "${REMOTE}:/tmp/store-popularity-dashboard.json"
$SSH $REMOTE "sudo -n /tmp/store-grafana-deploy --host $GRAFANA_HOST --dashboard /tmp/store-popularity-dashboard.json"
