#!/bin/sh
set -e

if [ -f package-lock.json ]; then
  npm ci --no-audit --no-fund --legacy-peer-deps
else
  npm install --no-audit --no-fund --legacy-peer-deps
fi

npm run dev


