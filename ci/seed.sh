#!/bin/sh -xe

apk add --no-cache curl jq
wget -q -O /usr/local/bin/garage "https://garagehq.deuxfleurs.fr/_releases/v1.0.1/$GARAGE_TRIPLE/garage"
chmod +x /usr/local/bin/garage

for i in $(seq 120); do
    curl -fsS http://apps.s3:3903/health >/dev/null 2>&1 && break
    sleep 1
done

NODE=$(curl -fsS -H "Authorization: Bearer test-admin-token" http://apps.s3:3903/v1/status | jq -r .node)
export GARAGE_RPC_HOST="$NODE@apps.s3:3901"

garage layout assign -z dc1 -c 1G "$NODE"
garage layout apply --version 1
garage bucket create apps.s3
garage key import --yes -n test GK31c4cef60f8f78b1bf12cd71 b8a31bf6c5d4e7a9f2b3c1d8e7f6a5b4c3d2e1f0a9b8c7d6e5f4a3b2c1d0e9f8
garage bucket allow --read --write --owner apps.s3 --key test
garage bucket website --allow apps.s3

./test/seed
