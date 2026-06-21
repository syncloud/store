#!/bin/bash
set -ex

if [ "$#" -ne 1 ]; then
    echo "usage: $0 <grafana-host>" >&2
    exit 1
fi
GRAFANA_HOST=$1

DIR=$( cd "$( dirname "$0" )" && pwd )
KEYFILE=/tmp/_deploy_key
SSH="ssh -i $KEYFILE -o StrictHostKeyChecking=no"
SCP="scp -i $KEYFILE -o StrictHostKeyChecking=no"
REMOTE="${DEPLOY_USER}@${DEPLOY_HOST}"

$SCP "$DIR/grafana/popularity.json" "${REMOTE}:/tmp/store-popularity-dashboard.json"
$SCP "$DIR/grafana-remote.sh" "${REMOTE}:/tmp/store-grafana-remote.sh"
$SSH $REMOTE "sudo -n bash /tmp/store-grafana-remote.sh $GRAFANA_HOST /tmp/store-popularity-dashboard.json"
