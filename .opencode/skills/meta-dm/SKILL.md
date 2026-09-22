---
name: meta-dm
description: Retrieve and preserve a specific Instagram DM conversation. Use the CLI to dispatch the GitHub Actions history workflow, inspect runs, and download the private archive.
---

# meta-dm

Use this skill when the user wants to retrieve the history of a specific Instagram DM conversation.

## Primary flow

1. Dispatch the history workflow:
   meta-dm workflow run <conversation_id> --all
2. Check workflow runs:
   meta-dm workflow runs
3. After the run succeeds, download its artifact:
   meta-dm workflow download <run_id>
4. Keep retrieved DM data outside the public repository.

## Direct CLI

meta-dm history <conversation_id> --all

## Rules

- Never put Meta access tokens in arguments, files, prompts, or committed data.
- Never commit real DM contents to the public repository.
- Treat conversation_id as an internal identifier; avoid displaying it unless needed.
- Prefer the GitHub Actions workflow when Meta credentials are stored as GitHub Secrets.
- The workflow is the execution boundary; the CLI is the control surface.
- Later, a CRX can provide the conversation identifier without changing this skill interface.

## OpenCode behavior

When invoked from OpenCode, operate through the CLI first. Ask only for the missing identifier or required choice. Do not ask the user to manually construct a GitHub Actions URL or YAML.

## Commands

- meta-dm workflow run <conversation_id> --all
- meta-dm workflow run <conversation_id> --before <RFC3339>
- meta-dm workflow runs
- meta-dm workflow download <run_id> --dir private-archive
- meta-dm history <conversation_id> --all
