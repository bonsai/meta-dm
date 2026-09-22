# meta-dm

CLI skill for retrieving and preserving the history of a specific Instagram DM conversation.

## Primary purpose

The main job is **conversation history retrieval**:

```text
conversation_id
  ↓
Meta Messaging API
  ↓
paginate backwards
  ↓
normalize
  ↓
dedupe by external message ID
  ↓
append JSONL
  ↓
history → latest → now → action
```

The skill is designed for an Instagram Professional Account controlled by the operator. It does not provide access to arbitrary users' private messages.

## CLI

```bash
meta-dm conversations
meta-dm history <conversation_id>
meta-dm history <conversation_id> --limit 1000
meta-dm history <conversation_id> --before 2026-01-01T00:00:00Z
meta-dm history <conversation_id> --all
meta-dm search <conversation_id> "keyword"
```

### History semantics

`history` retrieves the newest available messages and follows pagination toward older messages.

- `--limit` limits normalized messages written during the run.
- `--before` stops when messages reach the specified timestamp.
- `--all` continues until the API has no older page.
- Re-running is safe: existing Meta message IDs are deduplicated.
- The local JSONL event stream is the canonical historical record.

## Storage

```text
data/
  messages/
    <conversation_id>.jsonl
  sync/
    <conversation_id>.json
```

A message record keeps the Meta external ID, conversation ID, sender, timestamp, text when available, and source metadata.

Real DM contents should normally remain outside this public repository. Use a private repository or private storage for production data; public fixtures must be synthetic.

## Skill boundary

The skill does four things:

1. identify a conversation;
2. retrieve older messages by pagination;
3. normalize and persist them;
4. report sync state so the next run can continue safely.

Webhook ingestion, classification, reply generation, and dashboards are secondary extensions, not the core skill.

## Security

Never put access tokens, app secrets, webhook verification tokens, or real DM exports in Git history. Prefer environment variables or GitHub Actions secrets.

## Development status

The repository contains the skill contract and storage model. Meta API version, permissions, and exact endpoint parameters must be configured against the current official Meta documentation before production use.
