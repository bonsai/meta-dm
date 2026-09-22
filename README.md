# meta-dm

Instagram DM ingestion and analysis system for bons.ai.

## Purpose

Collect messages from an Instagram Professional Account through Meta's official messaging interfaces, normalize them into an append-only event stream, and make the resulting data available to GitHub Actions / `gh aw` for analysis, classification, search, and response drafting.

## Architecture

```text
Instagram DM
   │
   ├── Webhook ───────────────┐
   │                          ▼
   │                    ingest endpoint
   │                          │
   └── Messaging API ─────────┤
                              ▼
                         normalized DM
                              │
                              ▼
                    data/dm/*.jsonl
                              │
                         gh aw / Actions
                              │
                 ┌────────────┼────────────┐
                 ▼            ▼            ▼
             classify       search      reply draft
                 │            │            │
                 └────────────┼────────────┘
                              ▼
                         GitHub Pages
```

## CLI concept

```bash
meta-dm auth
meta-dm account
meta-dm sync
meta-dm conversations
meta-dm messages <conversation_id>
meta-dm search "keyword"
meta-dm export
```

The CLI is an interface over the canonical event data; it must not become the source of truth.

## Data principles

- Raw webhook/API payloads are retained separately from normalized records.
- Normalized messages are append-only JSONL events.
- IDs from Meta are preserved as external identifiers.
- Timestamps remain ISO 8601.
- Secrets and access tokens never enter Git history.
- Analysis is derived data and can be regenerated.
- Human messages and agent-generated drafts are different event types.

## Meta boundary

This project targets an Instagram Professional Account controlled by the operator. It does not attempt to access arbitrary users' private messages.

## Planned layers

- `ontology/` — concepts and relations
- `topology/` — system/world boundaries and interfaces
- `agent.md` — this repository's agent definition
- `docs/` — implementation and setup documentation
- `data/` — local/generated event data; sensitive data should not be committed by default
- `.github/workflows/` — scheduled or manually triggered synchronization and analysis

## Security

Use GitHub Actions secrets or an external secret manager for Meta credentials. Do not commit DM contents, access tokens, webhook secrets, or personally identifying exports to a public repository.

## Status

Scaffold: architecture and data contract first. API credentials and webhook deployment are configured separately.
