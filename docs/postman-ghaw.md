# Postman × GitHub Actions × gh aw

## Role

meta-dm uses three layers:

1. **Meta official Postman collections** — API contract/reference and request experiments.
2. **Postman CLI + GitHub Actions** — repeatable contract/integration checks.
3. **Go CLI / gh aw** — actual DM-history retrieval, persistence, and view updates.

Postman is therefore not the production storage layer. The repository remains the operational home for the normalized JSONL history.

## Official references

- Meta Instagram API workspace: https://www.postman.com/meta/instagram/overview
- Meta Instagram API collection: https://www.postman.com/meta/instagram/collection/6yqw8pt/instagram-api
- Meta Conversations API: https://www.postman.com/meta/messenger-platform-api/folder/22794852-255610cd-47f5-4f4d-b3fa-71aec360be9a
- Postman CLI GitHub Actions: https://learning.postman.com/docs/postman-cli/postman-cli-github-actions

The Conversations API collection is the relevant reference for this skill because it contains requests for listing conversations, finding a conversation by user, listing messages, and getting message details.

## Workflow

.github/workflows/postman.yml is a manual contract-test entry point.

Required repository secret:

- POSTMAN_API_KEY

Workflow input:

- collection_id
- optional environment_id

Do not put Meta access tokens or DM contents in the repository. Use GitHub Actions secrets or environment-level secrets.

## gh aw relationship

The intended automation path is:

    Issue / question
          ↓
        gh aw
          ↓
       meta-dm
          ↓
    Meta Conversations API
          ↓
    normalized message events
          ↓
    data/messages/<conversation_id>.jsonl
          ↓
    latest / now / view

Postman is used to validate the API request shape before the Go implementation is promoted into the normal meta-dm path.

## Security

Never commit:

- META_ACCESS_TOKEN
- POSTMAN_API_KEY
- real DM exports
- webhook secrets
- production Postman environments containing credentials

Synthetic fixtures are allowed in public tests.
