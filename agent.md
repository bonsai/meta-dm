# meta-dm Agent

## Type

`meta-dm`

## Role

A CLI skill that retrieves the history of one specified Instagram DM conversation and preserves it as a local canonical event stream.

## Primary concern

> 「このチャットを、可能なところまで過去へさかのぼって保存する」

## Operation

```text
identify
→ fetch
→ paginate
→ normalize
→ dedupe
→ persist
→ report
```

## Inputs

- Instagram conversation identifier
- authenticated Meta access token
- optional limit
- optional cutoff timestamp

## Outputs

- normalized message JSONL
- sync state
- retrieval statistics

## Invariants

1. Meta external message IDs remain unchanged.
2. Re-running a history operation is idempotent.
3. Older pages are traversed until a declared stop condition.
4. Raw source and normalized records remain distinguishable.
5. Incoming messages and generated replies are never conflated.
6. Credentials never enter persisted message data.

## Non-goals

- arbitrary users' private messages;
- GUI automation;
- automatic reply sending;
- classification as the primary operation;
- webhook-first architecture.

## Skill interface

```text
meta-dm conversations
meta-dm history <conversation_id> [--limit N] [--before ISO8601] [--all]
meta-dm search <conversation_id> <query>
```
