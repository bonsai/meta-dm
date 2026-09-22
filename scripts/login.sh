#!/usr/bin/env bash
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

: "${META_APP_ID:?META_APP_ID is required}"
: "${META_APP_SECRET:?META_APP_SECRET is required}"

export META_REDIRECT_URI="${META_REDIRECT_URI:-http://127.0.0.1:8787/callback}"
export META_SCOPE="${META_SCOPE:-instagram_business_basic,instagram_business_manage_messages}"

go build -o meta-dm ./cmd/meta-dm
./meta-dm login
