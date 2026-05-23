#!/bin/bash -ex

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

for i in $(seq 60); do
    mc alias set local http://minio test testtest 2>/dev/null && break
    sleep 1
done

mc mb -p local/test || true
mc anonymous set public local/test

for ch in master stable rc; do
    for app in testapp1 testapp2; do
        mc cp ${DIR}/${app}/meta/snap.yaml local/test/v2/apps/${ch}/${app}/snap.yaml
        mc cp ${DIR}/images/${app}.png    local/test/v2/apps/${ch}/${app}/icon.png
    done
done

for f in ${DIR}/testapp*.snap; do
    name=$(basename "$f")
    size=$(stat -c %s "$f")
    sha=$(openssl dgst -sha3-384 -binary "$f" | base64 | tr '+/' '-_' | tr -d '=')
    app=$(echo "$name" | sed 's/_.*//')
    ver=$(echo "$name" | sed 's/[^_]*_\([^_]*\)_.*/\1/')
    mc cp "$f" "local/test/apps/${name}"
    printf '%s' "$sha"  > /tmp/sha
    printf '%s' "$size" > /tmp/size
    mc cp /tmp/sha  "local/test/apps/${name}.sha384"
    mc cp /tmp/size "local/test/apps/${name}.size"
    printf '{"snap-revision":"%s","snap-id":"%s.%s","snap-size":"%s","snap-sha3-385":"%s"}' \
        "$ver" "$app" "$ver" "$size" "$sha" > /tmp/rev
    mc cp /tmp/rev "local/test/revisions/${sha}.revision"
done

for app in testapp1 testapp2; do
    for arch in amd64 arm64 armhf; do
        printf '1' > /tmp/ver
        mc cp /tmp/ver "local/test/releases/stable/${app}.${arch}.version"
    done
done

mc ls -r local/test | head -50
