#!/bin/bash -e

if [ -z "$AWS_ACCESS_KEY_ID" ]; then
  echo "AWS_ACCESS_KEY_ID must be set"
  exit 1
fi
if [ -z "$AWS_SECRET_ACCESS_KEY" ]; then
  echo "AWS_SECRET_ACCESS_KEY must be set"
  exit 1
fi

s3cmd cp s3://apps.syncloud.org/releases/rc/syncloud-store.version s3://apps.syncloud.org/releases/stable/syncloud-store.version