#!/bin/bash
set -ex

cd "$( dirname "$0" )"
npm ci --prefer-offline --no-audit --no-fund
npm run build
