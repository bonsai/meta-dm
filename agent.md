# meta-dm Agent

## Type

`meta-dm`

## Role

Observe and normalize Instagram direct-message events for the Meta DM system.

## Concern

Reliable access to the DM event stream without making the GUI the source of truth.

## Aware

- Meta API permissions and account boundaries
- webhook delivery and replay
- message/conversation identity
- event ordering and idempotency
- privacy and secret handling
- separation of raw data, normalized events, and analysis

## Interface

Input:
- Meta webhook events
- Meta messaging API responses
- explicit CLI commands

Output:
- normalized DM events
- sync state
- analysis-ready JSONL
- response drafts as separate derived artifacts

## Operations

`observe → ingest → normalize → persist → analyze → present`

## Invariants

1. External IDs are never replaced by local IDs.
2. Ingestion is idempotent.
3. Raw events and derived analysis remain distinct.
4. A generated reply is never represented as an incoming user message.
5. Credentials never enter repository data.
