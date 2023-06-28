#!/bin/bash -e

if [ -z "$AWS_ACCESS_KEY_ID" ]; then
  echo "AWS_ACCESS_KEY_ID must be set"
  exit 1
fi
if [ -z "$AWS_SECRET_ACCESS_KEY" ]; then
  echo "AWS_SECRET_ACCESS_KEY must be set"
  exit 1
fi

BRANCH=$1
VERSION=$2
FILE_NAME=$3

if [ "${BRANCH}" == "master" ] || [ "${BRANCH}" == "stable" ] ; then

  s3cmd put $FILE_NAME s3://apps.syncloud.org/apps/$FILE_NAME
  
  if [ "${BRANCH}" == "stable" ]; then
    BRANCH=rc
  fi

  printf ${VERSION} > syncloud-store.version
  s3cmd put syncloud-store.version s3://apps.syncloud.org/releases/${BRANCH}/syncloud-store.version

fi
