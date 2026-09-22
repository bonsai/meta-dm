#!/usr/bin/env bash
set -Eeuo pipefail

ROOT="$(git rev-parse --show-toplevel 2>/dev/null || true)"
if [[ -z "$ROOT" || ! -f "$ROOT/go.mod" ]]; then
  echo "ERROR: run this from a cloned bonsai/meta-dm repository."
  exit 1
fi
cd "$ROOT"

: "${META_GRAPH_VERSION:=v23.0}"
API="https://graph.facebook.com/${META_GRAPH_VERSION}/me/accounts"
FIELDS="name,access_token,tasks,instagram_business_account"

echo "== Meta Page Access Token bootstrap =="
echo "This is the Postman 'Get Access Tokens of Pages You Manage' request."
echo "The User Access Token is read from stdin and is never printed."

if [[ -n "${META_USER_ACCESS_TOKEN:-}" ]]; then
  USER_TOKEN="$META_USER_ACCESS_TOKEN"
else
  read -rsp "User Access Token: " USER_TOKEN
  echo
fi

[[ -n "$USER_TOKEN" ]] || { echo "ERROR: User Access Token is required." >&2; exit 1; }

TMP="$(mktemp)"
trap 'rm -f "$TMP"' EXIT

curl --fail-with-body --silent --show-error --location --globoff   --get "$API"   --data-urlencode "fields=$FIELDS"   --data-urlencode "access_token=$USER_TOKEN"   -o "$TMP"

python3 - "$TMP" <<'PY'
import json, sys
p = json.load(open(sys.argv[1]))
rows = p.get("data", [])
if not rows:
    raise SystemExit("ERROR: no Facebook Pages were returned.")
for i, row in enumerate(rows):
    ig = row.get("instagram_business_account") or {}
    print(f"[{i}] {row.get('name','')} page_id={row.get('id','')} ig_user_id={ig.get('id','')}")
PY

read -rp "Page number to use [0]: " INDEX
INDEX="${INDEX:-0}"

python3 - "$TMP" "$INDEX" <<'PY' > .meta-dm/page-token.env
import json, sys, os
p = json.load(open(sys.argv[1]))
i = int(sys.argv[2])
rows = p.get("data", [])
if i < 0 or i >= len(rows):
    raise SystemExit("ERROR: invalid page number.")
row = rows[i]
ig = row.get("instagram_business_account") or {}
token = row.get("access_token", "")
ig_id = ig.get("id", "")
if not token or not ig_id:
    raise SystemExit("ERROR: selected Page has no Page Access Token or linked Instagram account.")
print("META_ACCESS_TOKEN=" + token)
print("META_IG_USER_ID=" + ig_id)
PY

chmod 600 .meta-dm/page-token.env

set -a
source .meta-dm/page-token.env
set +a

echo
echo "Selected:"
echo "  META_IG_USER_ID=$META_IG_USER_ID"
echo "  META_ACCESS_TOKEN=********"
echo

if command -v gh >/dev/null 2>&1 && gh auth status >/dev/null 2>&1; then
  read -rp "Set these as GitHub Actions secrets? [Y/n] " PUSH
  PUSH="${PUSH:-Y}"
  if [[ "$PUSH" =~ ^[Yy]$ ]]; then
    printf '%s' "$META_ACCESS_TOKEN" | gh secret set META_ACCESS_TOKEN
    printf '%s' "$META_IG_USER_ID" | gh secret set META_IG_USER_ID
    printf '%s' "$META_GRAPH_VERSION" | gh secret set META_GRAPH_VERSION
    echo "GitHub Actions secrets updated."
  fi
fi

echo
echo "Local credentials: .meta-dm/page-token.env"
echo "Run: bash scripts/doctor.sh"
