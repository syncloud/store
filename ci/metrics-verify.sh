#!/bin/bash
set -ex

apt-get update
apt-get install -y curl python3

STORE_METRICS=http://api.store.test:9090/metrics
VM=http://vm:8428

curl -fsS "$STORE_METRICS" > /tmp/metrics.txt
head -20 /tmp/metrics.txt

curl -fsS -X POST --data-binary @/tmp/metrics.txt "$VM/api/v1/import/prometheus"

sleep 1

answer=$(curl -fsS "$VM/api/v1/query?query=sum(store_popularity_record_total)")
echo "$answer"
value=$(echo "$answer" | python3 -c 'import json,sys; d=json.load(sys.stdin); r=d["data"]["result"]; print(int(float(r[0]["value"][1])) if r else 0)')
echo "store_popularity_record_total sum: $value"
[ "$value" -gt 0 ]

testapp1=$(curl -fsS "$VM/api/v1/query?query=store_popularity_record_total%7Bsnap%3D%22testapp1%22%7D" \
  | python3 -c 'import json,sys; d=json.load(sys.stdin); r=d["data"]["result"]; print(int(float(r[0]["value"][1])) if r else 0)')
echo "store_popularity_record_total{snap=testapp1}: $testapp1"
[ "$testapp1" -gt 0 ]
