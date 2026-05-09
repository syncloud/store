#!/bin/bash
set -ex

cd "$( dirname "$0" )"
npm ci --no-audit --no-fund
npx playwright test --project=desktop
npx playwright test --project=mobile
