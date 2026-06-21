#!/bin/bash
set -ex

if [ "$#" -ne 1 ]; then
    echo "usage: $0 <env>" >&2
    exit 1
fi
ENV=$1

DIR=$( cd "$( dirname "$0" )" && pwd )
ROOT=$( cd "$DIR/.." && pwd )

ENVFILE="$DIR/grafana.deploy.${ENV}"
[ -f "$ENVFILE" ] || { echo "no grafana env file: $ENVFILE" >&2; exit 1; }
source "$ENVFILE"

KEYFILE=/tmp/_deploy_key
SSH="ssh -i $KEYFILE -o StrictHostKeyChecking=no"
SCP="scp -p -i $KEYFILE -o StrictHostKeyChecking=no"
REMOTE="${DEPLOY_USER}@${DEPLOY_HOST}"

$SCP "$ROOT/build/bin/grafana-deploy" "${REMOTE}:/tmp/store-grafana-deploy"
$SCP "$DIR/grafana/popularity.json" "${REMOTE}:/tmp/store-popularity-dashboard.json"
$SSH $REMOTE "sudo -n /tmp/store-grafana-deploy --host $GRAFANA_HOST --dashboard /tmp/store-popularity-dashboard.json"
