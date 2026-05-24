#!/bin/sh -xe

NET=$(docker inspect "$(hostname)" --format '{{range $k, $v := .NetworkSettings.Networks}}{{$k}}{{"\n"}}{{end}}' | grep -m1 '^drone-')
echo "using network=$NET PWD=$PWD"

docker pull "$PUBLISHER_IMAGE"

docker run --rm --network "$NET" --volumes-from "$(hostname)" -e SYNCLOUD_TOKEN -w "$PWD" \
    "$PUBLISHER_IMAGE" \
    snap -f "test/testapp1_3_${ARCH}.snap" -c stable -s http://api.store.test \
    -y test/testapp1/meta/snap.yaml -i test/images/testapp1.png

docker run --rm --network "$NET" curlimages/curl:8.10.1 \
    -fsS 'http://api.store.test/api/ui/v1/apps?channel=stable' | grep -q testapp1
