#!/bin/sh -xe

apk add --no-cache curl jq
wget -q -O /usr/local/bin/garage "https://garagehq.deuxfleurs.fr/_releases/v1.0.1/$GARAGE_TRIPLE/garage"
chmod +x /usr/local/bin/garage

for i in $(seq 120); do
    curl -fsS http://s3:3903/health >/dev/null 2>&1 && break
    sleep 1
done

NODE=$(curl -fsS -H "Authorization: Bearer test-admin-token" http://s3:3903/v1/status | jq -r .node)
export GARAGE_RPC_HOST="$NODE@s3:3901"

garage layout assign -z dc1 -c 1G "$NODE"
garage layout apply --version 1
garage bucket create test
garage key import --yes -n test GK31c4cef60f8f78b1bf12cd71 testtest
garage bucket allow --read --write --owner test --key test
garage bucket website --allow test
garage status
