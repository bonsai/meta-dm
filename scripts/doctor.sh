#!/usr/bin/env bash
set -Eeuo pipefail

ROOT="$(git rev-parse --show-toplevel 2>/dev/null || true)"
if [[ -z "$ROOT" || ! -f "$ROOT/go.mod" ]]; then
  echo "ERROR: run this from a cloned bonsai/meta-dm repository."
  exit 1
fi
cd "$ROOT"

section() { printf '\n== %s ==\n' "$1"; }
need_cmd() {
  if command -v "$1" >/dev/null 2>&1; then
    printf 'OK   %-10s %s\n' "$1" "$("$1" --version 2>/dev/null | head -1 || true)"
  else
    printf 'MISS %-10s install required\n' "$1"
    return 1
  fi
}

section "environment"
printf 'repo=%s\n' "$ROOT"
printf 'branch=%s\n' "$(git branch --show-current)"
printf 'commit=%s\n' "$(git rev-parse --short HEAD)"
need_cmd git
need_cmd gh
need_cmd go
need_cmd curl

section "git"
git status --short --branch

section "go"
go version
go env GOPATH GOMOD
go test ./...
go build -o meta-dm ./cmd/meta-dm
./meta-dm

section "github auth"
gh auth status

section "workflows"
gh workflow list
gh run list --workflow dm-history.yml --limit 5 || true
gh run list --workflow postman.yml --limit 5 || true

section "repository secrets"
gh secret list || true
for name in META_ACCESS_TOKEN META_IG_USER_ID META_GRAPH_VERSION; do
  if gh secret list | awk '{print $1}' | grep -qx "$name"; then
    printf 'OK   secret %-24s configured\n' "$name"
  else
    printf 'MISS secret %-24s NOT configured\n' "$name"
  fi
done

section "local meta-dm state"
if [[ -f .meta-dm/token.json ]]; then
  echo "OK   .meta-dm/token.json exists"
  chmod 600 .meta-dm/token.json
else
  echo "INFO .meta-dm/token.json not found"
fi
for name in META_ACCESS_TOKEN META_IG_USER_ID META_GRAPH_VERSION; do
  if [[ -n "${!name:-}" ]]; then
    printf 'OK   env %-24s set\n' "$name"
  else
    printf 'INFO env %-24s not set\n' "$name"
  fi
done

section "workflow failure diagnostics"
RUN_ID="${RUN_ID:-}"
if [[ -n "$RUN_ID" ]]; then
  echo "run_id=$RUN_ID"
  gh run view "$RUN_ID" --log-failed || gh run view "$RUN_ID" --log
fi

section "optional dispatch"
CONVERSATION_ID="${CONVERSATION_ID:-}"
if [[ -n "$CONVERSATION_ID" ]]; then
  echo "dispatching Private DM History"
  args=(workflow run .github/workflows/dm-history.yml -f "conversation_id=$CONVERSATION_ID" -f all=true)
  if [[ -n "${BEFORE:-}" ]]; then
    args+=( -f "before=$BEFORE" )
  fi
  gh "${args[@]}"
  echo
  echo "latest runs:"
  gh run list --workflow dm-history.yml --limit 3
else
  echo "INFO set CONVERSATION_ID=... to dispatch a real DM history run."
fi

section "done"
echo "No DM contents or access-token values are printed by this diagnostic."
