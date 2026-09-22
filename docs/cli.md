# CLI implementation

## Environment

```bash
export META_ACCESS_TOKEN="..."
export META_IG_USER_ID="..."
export META_GRAPH_VERSION="v23.0"
```

The graph version is configurable so the skill can follow Meta's supported version without changing the CLI.

## Commands

```bash
go run ./cmd/meta-dm conversations
go run ./cmd/meta-dm history CONVERSATION_ID --all
go run ./cmd/meta-dm history CONVERSATION_ID --before 2026-01-01T00:00:00Z
```

The history command follows `paging.next` when `--all` is supplied and deduplicates by Meta message ID.

## Production note

Verify the current Meta permissions, API version, and endpoint availability for the connected Instagram Professional Account before production use.
