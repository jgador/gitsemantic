# GitSemantic CLI

GitSemantic CLI is a thin HTTP client for compatible GitSemantic server deployments. It is part of Goblin Board's semantic search work.

- Goblin Board: https://goblinboard.com
- Semantic Search: https://goblinboard.com/semantic-search

> A coding-agent skill for GitSemantic-first repository exploration is included in [`SKILL.md`](SKILL.md).

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

## Install on Windows

For a direct Windows install without waiting for WinGet publication, paste this one-liner into PowerShell. It downloads the existing release zip, extracts `gitsemantic.exe` into `%LOCALAPPDATA%\Programs\GitSemantic`, and adds that directory to the user `PATH`.

```powershell
$version="0.1.0";$installDir=Join-Path $env:LOCALAPPDATA "Programs\GitSemantic";$assetName="gitsemantic_${version}_windows_amd64.zip";$downloadUrl="https://github.com/jgador/gitsemantic/releases/download/v$version/$assetName";$tempDir=Join-Path $env:TEMP ("gitsemantic-install-"+[guid]::NewGuid().ToString("N"));$zipPath=Join-Path $tempDir $assetName;New-Item -ItemType Directory -Path $tempDir -Force|Out-Null;New-Item -ItemType Directory -Path $installDir -Force|Out-Null;try{Invoke-WebRequest -Uri $downloadUrl -OutFile $zipPath;Expand-Archive -LiteralPath $zipPath -DestinationPath $tempDir -Force;Copy-Item -LiteralPath (Join-Path $tempDir "gitsemantic.exe") -Destination (Join-Path $installDir "gitsemantic.exe") -Force;$u=@([Environment]::GetEnvironmentVariable("Path","User") -split ";"|Where-Object{$_});if($u -notcontains $installDir){[Environment]::SetEnvironmentVariable("Path",(($u+$installDir)-join ";"),"User")};$p=@($env:Path -split ";"|Where-Object{$_});if($p -notcontains $installDir){$env:Path=if([string]::IsNullOrWhiteSpace($env:Path)){$installDir}else{"$env:Path;$installDir"}}}finally{Remove-Item -LiteralPath $tempDir -Recurse -Force -ErrorAction SilentlyContinue};gitsemantic version
```

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
