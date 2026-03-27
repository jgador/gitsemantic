# GitSemantic CLI

GitSemantic CLI is a thin HTTP client for compatible GitSemantic server deployments. It is part of Goblin Board's semantic search work.

- Goblin Board: https://goblinboard.com
- Semantic Search: https://goblinboard.com/semantic-search

> Active development: GitSemantic CLI is not ready for use yet. We are aiming to make an early testing build available by the end of March 2026, if current work stays on track.

GitSemantic CLI helps AI coding agents work across the full history of a repository, not just the current checkout. It has been tested against `dotnet/runtime`, a codebase with more than 140,000 commits. You can try the experience at https://goblinboard.com/semantic-search.

GitSemantic CLI is part of the open-source effort around Goblin Board and gives developers a preview of the repository-history workflows behind Goblin Board integrations.

## Run GitSemantic Server with Docker

An example Compose file is included at `docker-compose.example.yml`.

Replace the sample absolute path `/absolute/path/to/your/repo` in that file with the host repository you want to ingest, then start the local self-hosted server:

```bash
docker compose -f docker-compose.example.yml --project-name gitsemantic up -d
```

## Connect with GitSemantic CLI

### Local Self-Hosted Server

The CLI already defaults to `http://127.0.0.1:7280` with API version `1`, so you can run it directly against the local server.

Check that the server is ready:

```bash
gitsemantic status
gitsemantic version
gitsemantic repos
```

Start ingestion for the mounted repository:

```bash
gitsemantic ingest --repo /repo --mode All
```

The ingest command returns a job ID. Check that job until it completes:

```bash
gitsemantic status --job <job-id>
```

List indexed repositories and note the `repoId`:

```bash
gitsemantic repos
```

Run a query:

```bash
gitsemantic query "how does authentication work?" --repo-id <repo-id>
```

### Hosted Server

Hosted GitSemantic server mode is not part of this OSS workflow. That hosted experience will be supported through Goblin Board instead.

## Notes

- `--repo` must be the path visible inside the server container. With the example above, use `/repo`.
- Set either `OPENAI_API_KEY` or both `AZURE_OPENAI_ENDPOINT` and `AZURE_OPENAI_API_KEY`.
- `Issues` ingestion is still under development.
