# Data Model

The canonical internal representation is an event stream.

## Message event

```json
{
  "event_id": "stable-local-event-id",
  "event_type": "message_received",
  "source": "instagram",
  "external_id": "META_MESSAGE_ID",
  "conversation_id": "META_CONVERSATION_ID",
  "participant": {
    "external_id": "META_USER_ID"
  },
  "message": {
    "text": "..."
  },
  "observed_at": "2026-09-22T12:00:00Z"
}
```

## Storage

Use JSONL for append-only events:

```text
data/
  raw/
  events/
  derived/
  sync/
```

Do not publish real DM contents in a public repository. For a public demo, use synthetic fixtures.

## Idempotency

The Meta external message/event identifier is the primary deduplication key. Replayed webhooks must not create duplicate canonical events.

## Derived data

Examples:

- conversation summaries
- intent labels
- unanswered-message flags
- response drafts
- topic aggregates

Derived data always references the source event IDs.
