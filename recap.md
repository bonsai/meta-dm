# meta-dm Recap

## Goal

Instagramの特定DMチャットを指定し、過去方向へ取得してJSONLとして保存するCLI Skill。

## Current architecture

Instagram DM → CRX → Resolver → OpenCode Skill → meta-dm CLI → GitHub Actions → Meta Conversations API → private archive

現在はCRX/Resolverより先に、CLIで実DM取得を検証するPhase 1を優先する。

## Completed

- Go CLI: conversations / history / workflow
- GitHub Actions: DM history workflow
- CLIからGitHub Actionsを操作
- OpenCode Skill: .opencode/skills/meta-dm/SKILL.md
- Postman contract/integration workflow
- private archive / Artifact保存設計
- scripts/doctor.sh
- scripts/meta-page-token.sh
- .meta-dm/ credential exclusion

## Important authentication distinction

meta-dm login はInstagram Login OAuth系統。

DMのConversations APIで使うPage Access Token取得には、Meta公式の流れに沿って Facebook User Access Token → /me/accounts → Page Access Token + Instagram Business Account ID を使う。

## Current local state

- origin/main: 18776a2
- scripts/meta-page-token.sh はpull済み
- meta-dm はローカルbuild binaryとして未追跡
- 過去のローカル変更はstash済み

## DM取得手順

### 1. Page Access Token bootstrap

```sh
cd ~/meta-dm
bash scripts/meta-page-token.sh
```

User Access Token: にはMetaのUser Access Tokenだけを入力する。TokenをChatGPTやGitへ貼らない。

対象Pageを選択し、GitHub Secretsへの登録を聞かれたら Y。

### 2. Diagnose

```sh
bash scripts/doctor.sh
```

### 3. Conversation discovery

```sh
./meta-dm conversations
```

### 4. Small history test

```sh
./meta-dm history "CONVERSATION_ID" --limit 10
```

### 5. Full history

```sh
./meta-dm history "CONVERSATION_ID" --all
```

### 6. GitHub Actions execution

```sh
./meta-dm workflow run "CONVERSATION_ID" --all
./meta-dm workflow runs
./meta-dm workflow download "RUN_ID"
```

## Phase 1 completion criteria

- Meta authentication works
- Page Access Token obtained
- Instagram Business Account ID obtained
- conversation_id obtained
- 10 messages retrieved
- full history retrieval works
- JSONL archive is created
- GitHub Actions run succeeds
- Artifact can be downloaded

## Next work after Phase 1

1. Verify current Meta Conversations API version, permissions, endpoint and response format against official documentation.
2. Fix/strengthen workflow preflight (meta-dm help or a zero-exit version command rather than relying on no-argument execution).
3. Build Resolver to map the currently opened Instagram DM to conversation_id.
4. Build CRX so the user does not need to copy/paste conversation_id.
5. Keep Meta credentials out of CRX, prompts, public repositories, and DM archives committed to public Git.

## Security

- Rotate the App Secret if it was exposed previously.
- Never paste access tokens or secrets into ChatGPT.
- Never commit real DM contents to a public repository.
- Prefer GitHub Secrets for GitHub Actions credentials.

## Final UX target

User opens an Instagram DM and asks:

> このトークを過去まで取得して

CRX identifies the current conversation, Resolver resolves conversation_id, OpenCode invokes the meta-dm Skill, and the existing CLI/Workflow performs retrieval without making the user aware of the underlying operational steps.
