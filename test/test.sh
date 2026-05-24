#!/bin/bash -ex

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

apt update
apt install -y sshpass curl openssh-client

SCP="sshpass -p syncloud scp -o StrictHostKeyChecking=no"
SSH="sshpass -p syncloud ssh -o StrictHostKeyChecking=no"
ARTIFACTS_DIR=${DIR}/artifacts
LOG_DIR=$ARTIFACTS_DIR/log
mkdir -p $LOG_DIR

${DIR}/wait-for-device.sh api.store.test
for i in $(seq 60); do
    curl -s -o /dev/null --max-time 2 http://apps/ && break
    sleep 1
done

$SSH root@api.store.test "rm -rf /tmp/syncloud-store && mkdir -p /tmp/syncloud-store/config"
$SCP -r ${DIR}/../deploy root@api.store.test:/tmp/syncloud-store/
$SCP -r ${DIR}/../config/test root@api.store.test:/tmp/syncloud-store/config/

$SSH root@api.store.test "apt update && apt install -y docker.io apache2"
$SSH root@api.store.test "
    AWS_ACCESS_KEY_ID='${AWS_ACCESS_KEY_ID}' \
    AWS_SECRET_ACCESS_KEY='${AWS_SECRET_ACCESS_KEY}' \
    AWS_S3_ENDPOINT='${AWS_S3_ENDPOINT}' \
    AWS_REGION='${AWS_REGION}' \
    bash /tmp/syncloud-store/deploy/deploy.sh '${DOCKER_IMAGE}' test
"

code=0
set +e
DEPLOY_URL="http://api.store.test" DEPLOY_HOST=api.store.test DEPLOY_USER=root \
    SSH_PASSWORD=syncloud ${DIR}/../ci/deploy-verify.sh test
code=$?
set -e

$SSH api.store.test journalctl > $LOG_DIR/journalctl.store.log 2>&1 || true
$SSH api.store.test 'docker logs syncloud-store 2>&1' > $LOG_DIR/docker.store.log 2>&1 || true
$SSH api.store.test ls -la /var/www/store > $LOG_DIR/var.www.store.log 2>&1 || true
chmod -R a+r $ARTIFACTS_DIR || true

exit $code
