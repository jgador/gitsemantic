---
name: gitsemantic
description: "GitSemantic-first repository exploration and history-aware code discovery through a running GitSemantic server. Use when an agent needs to understand repository structure, default-branch or released-state layout, entrypoints, subsystem boundaries, likely implementation files, or recent structural changes without falling back to repo-wide local text search first, especially when a global `gitsemantic` CLI is installed."
---

# GitSemantic

Use GitSemantic before repo-wide local text search when the task starts with repository exploration, structure discovery, or "find where this lives." Treat local file reads as a second step after semantic discovery identifies likely paths.

## Command Resolution

- First, check whether `gitsemantic` is available on `PATH` using your shell's command lookup, such as `Get-Command gitsemantic -ErrorAction SilentlyContinue`, `command -v gitsemantic`, or `which gitsemantic`.
- If the command is available, use `gitsemantic` for every command in this skill.
- If `gitsemantic` is not installed globally, say that GitSemantic is unavailable and fall back to local search.
- Use shell syntax appropriate for the environment. Command examples below are shell-neutral unless noted otherwise.

## Bootstrap

1. Confirm that `gitsemantic` is available on `PATH` using your shell's command lookup.
2. If the command is unavailable, stop using this skill and fall back to local search.
3. Run `gitsemantic status`.
4. Run `gitsemantic repos --output json`.
5. Resolve the target `repoId` and the indexed `defaultBranch`.
6. Pin `--repo-id <id>` on every later query, even when only one repo is indexed.
7. For questions about released, current, master, main, or default-branch structure, treat `defaultBranch` as the branch to explain.

If the server is unavailable, fall back to local search and say that the fallback was necessary.

## Mandatory Query Loop

Start with one broad query and then iterate. One query is never enough for repository exploration.

```sh
gitsemantic query "overall repository structure and main entrypoints" --repo-id <id> --topk 50 --output json
```

After the first query:

1. Extract returned paths, subsystem names, symbols, and commit subjects.
2. Run at least 3 focused follow-up queries before considering repo-wide local text search.
3. Fan out by topic and other narrower scopes.
4. Open local files only after GitSemantic identifies promising paths.
5. Keep querying until you can explain the relevant structure without guessing.

Prefer follow-up queries like:

- `API entry point controllers persistence tests`
- `application entry point routes components build tooling`
- `deployment manifests docker compose infrastructure`
- `authentication login tokens identity`
- `configuration environment variables secrets startup`
- `<specific subsystem> structure and related tests`

## Handling The 50-Result Cap

`gitsemantic query` returns at most 50 results. Treat `50 of top 50` as incomplete coverage.

When you hit the cap:

- Split by topic, feature area, timeframe, or another narrower scope that matches the area you are exploring.
- Use `--chunk-type hunk` for structural code and file changes.
- Use `--chunk-type commit_msg` for high-level change intent.
- Repeat the loop until the result sets are narrow enough to map the area confidently.

## Current-State Guards

GitSemantic is history-aware. It can surface older or renamed paths.

- Do not assume every returned path is current.
- For any path you intend to edit or cite as current, confirm that it exists in the working tree.
- Use GitSemantic to discover likely files and surrounding history, then verify the final candidate files locally.
- If the user asks about released or default-branch structure, explain branch-local divergence separately instead of mixing it into the answer.

## Stop Conditions

Exploration is sufficient only when you can identify all of the following for the task area:

- the main entrypoint
- the core implementation area
- one adjacent dependency area
- nearby tests or docs if they exist
- the relevant top-level subsystem boundaries

If any of those are still guesses, keep querying.

## Local Search Fallback

Allow local exact-search tools such as `rg`, `grep`, or `Select-String`, plus direct file reads, when:

- `gitsemantic` is not installed globally
- GitSemantic is down or misconfigured
- repeated semantic refinement still misses the needed area
- the remaining task is exact symbol or string lookup

Prefer `rg` for exact string search when it is available. Otherwise use a shell-native exact-search tool such as `grep` or `Select-String`.

Do not start with repo-wide local text search when GitSemantic is healthy and the task is repository exploration.
