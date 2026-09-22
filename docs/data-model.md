# Data Model

## Canonical unit

A retrieved message becomes one normalized event.

```json
{
  "event_type": "message_received",
  "source": "instagram",
  "external_id": "META_MESSAGE_ID",
  "conversation_id": "META_CONVERSATION_ID",
  "sender": {
    "external_id": "META_USER_ID"
  },
  "message": {
    "text": "..."
  },
  "observed_at": "2026-09-22T12:00:00Z"
}
```

## File layout

```text
data/messages/<conversation_id>.jsonl
data/sync/<conversation_id>.json
```

Each conversation has one append-only JSONL stream. A sync state records the latest retrieval run and enough cursor/state metadata to continue.

## Deduplication

The primary deduplication key is the Meta external message/event identifier.

The local conversation file may therefore be safely rebuilt or extended from repeated API pages without creating duplicate messages.

## Ordering

Persisted records should be normalized to chronological order for human inspection, while sync logic may fetch pages newest-first.

## Privacy

Real DM text is sensitive. Production storage should be private. This public repository should contain only schemas, code, documentation, and synthetic fixtures.
