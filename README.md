# GitSemantic CLI

GitSemantic CLI is a thin HTTP client for compatible GitSemantic server deployments. It is part of Goblin Board's semantic search work.

- Goblin Board: https://goblinboard.com
- Semantic Search: https://goblinboard.com/semantic-search

> GitSemantic CLI can already be used today. A sample `SKILL.md` for coding agents will be added later in `gitsemantic`, but you can use the CLI directly now.

GitSemantic CLI helps AI coding agents work across the full history of a repository, not just the current checkout. It has been tested against `dotnet/runtime`, a codebase with more than 140,000 commits. You can try the experience at https://goblinboard.com/semantic-search.

GitSemantic CLI is part of the open-source effort around Goblin Board and gives developers a preview of the repository-history workflows behind Goblin Board integrations.

## Download GitSemantic CLI

Prebuilt CLI archives are attached to each GitHub Release.

- Windows x64: `gitsemantic_<version>_windows_amd64.zip`
- Linux x64: `gitsemantic_<version>_linux_amd64.tar.gz`
- macOS Apple Silicon: `gitsemantic_<version>_darwin_arm64.tar.gz`
- macOS Intel: `gitsemantic_<version>_darwin_amd64.tar.gz`
- Checksums: `checksums.txt`

Maintainers publish these assets by pushing a version tag such as `v0.1.0`. The release workflow builds the archives and uploads them to the GitHub Release automatically.

## Windows Package Manager

A ready-to-submit WinGet manifest for `JGador.GitSemantic` version `0.1.0` lives under `packaging/winget/manifests/j/JGador/GitSemantic/0.1.0`.

If you want to test the package locally before it is merged into the WinGet community repository, run:

```powershell
winget settings --enable LocalManifestFiles
winget install --manifest .\packaging\winget\manifests\j\JGador\GitSemantic\0.1.0
```

Run the `winget settings --enable LocalManifestFiles` command once from an elevated PowerShell session.

After the package is accepted into `microsoft/winget-pkgs`, install it with:

```powershell
winget install --id JGador.GitSemantic
```

See `packaging/winget/README.md` for the release-to-WinGet workflow and the exact `v0.1.0` asset metadata used in the manifest.

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
