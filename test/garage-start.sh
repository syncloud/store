#!/bin/sh

cat > /etc/garage.toml <<'CFG'
metadata_dir = "/tmp/meta"
data_dir = "/tmp/data"
db_engine = "lmdb"
replication_mode = "none"
rpc_bind_addr = "[::]:3901"
rpc_public_addr = "127.0.0.1:3901"
rpc_secret = "1799ff75e85715cd0bd91e09f2a9d70b1799ff75e85715cd0bd91e09f2a9d70b"
[s3_api]
s3_region = "garage"
api_bind_addr = "[::]:80"
root_domain = ".s3.garage"
[s3_web]
bind_addr = "[::]:3902"
root_domain = ".web.garage"
index = "index.html"
[admin]
api_bind_addr = "[::]:3903"
admin_token = "test-admin-token"
CFG

/garage server > /tmp/garage.log 2>&1 &
GARAGE_PID=$!

for i in $(seq 120); do
    if /garage status 2>/dev/null | grep -qE "(NO ROLE|HEALTHY)"; then
        echo "garage RPC ready after ${i}s"
        break
    fi
    sleep 1
done

NODE=$(/garage status 2>/dev/null | awk '/NO ROLE/ {print $1; exit}')
if [ -z "$NODE" ]; then
    echo "no NO ROLE node found, status:"
    /garage status
fi

/garage layout assign -z dc1 -c 1G "$NODE"   || echo "layout assign failed"
/garage layout apply --version 1             || echo "layout apply failed"
/garage bucket create test                   || echo "bucket create failed"
/garage key import --yes -n test test testtest || echo "key import failed"
/garage bucket allow --read --write --owner test --key test || echo "bucket allow failed"
/garage bucket website --allow test          || echo "bucket website failed"

echo "garage init complete; current status:"
/garage status

wait $GARAGE_PID
echo "garage exited unexpectedly"
sleep infinity
