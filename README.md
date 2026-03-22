# GitSemantic CLI

GitSemantic CLI is a thin HTTP client for compatible GitSemantic server deployments. It is part of Goblin Board's semantic search work.

- Goblin Board: https://goblinboard.com
- Semantic Search: https://goblinboard.com/semantic-search

> Active development: GitSemantic CLI is not ready for use yet. We are aiming to make an early testing build available by the end of March 2026, if current work stays on track.

GitSemantic CLI helps AI coding agents work across the full history of a repository, not just the current checkout. It has been tested against `dotnet/runtime`, a codebase with more than 140,000 commits. You can try the experience at https://goblinboard.com/semantic-search.

GitSemantic CLI is part of the open-source effort around Goblin Board and gives developers a preview of the repository-history workflows behind Goblin Board integrations.

## Run GitSemantic Server with Docker

Run a compatible GitSemantic server with Docker and mount the repository you want to index into the server container.

Example `docker-compose.yml`:

```yaml
name: gitsemantic

services:
  gitsemantic-server:
    container_name: gitsemantic-server
    image: your-compatible-gitsemantic-server-image
    ports:
      - "7280:7280"
    environment:
      ASPNETCORE_URLS: http://+:7280
      GITSEMANTIC_API_PORT: 7280
      OPENAI_API_KEY: ${OPENAI_API_KEY:-}
      # Or use Azure OpenAI instead:
      # AZURE_OPENAI_ENDPOINT: ${AZURE_OPENAI_ENDPOINT:-}
      # AZURE_OPENAI_API_KEY: ${AZURE_OPENAI_API_KEY:-}
    volumes:
      - ./:/repo:ro
```

Start the named Compose project and server container:

```bash
docker compose --project-name gitsemantic up -d
```

## Connect with GitSemantic CLI

Point the CLI at the server and provide the bearer token from your server deployment:

```bash
export GITSEMANTIC_SERVER=http://127.0.0.1:7280
export GITSEMANTIC_TOKEN=<your-server-token>
export GITSEMANTIC_API_VERSION=1
```

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

Notes:

- Protected endpoints require `GITSEMANTIC_TOKEN` or `GITSEMANTIC_TOKEN_FILE`.
- `--repo` must be the path visible inside the server container. With the example above, use `/repo`.
- Set either `OPENAI_API_KEY` or both `AZURE_OPENAI_ENDPOINT` and `AZURE_OPENAI_API_KEY`.
- `Issues` ingestion is still under development.
