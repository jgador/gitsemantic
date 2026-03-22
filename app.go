package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"
)

type App struct {
	stdout io.Writer
	stderr io.Writer
}

const maxQueryTopK = 50

func (a *App) Run(ctx context.Context, args []string) int {
	if len(args) == 0 {
		a.writeRootUsage(a.stdout)
		return 0
	}

	switch args[0] {
	case "help", "--help", "-help":
		if len(args) > 1 {
			return a.runHelp(args[1])
		}

		a.writeRootUsage(a.stdout)
		return 0
	case "version":
		return a.runVersion(ctx, args[1:])
	case "status":
		return a.runStatus(ctx, args[1:])
	case "repos":
		return a.runRepos(ctx, args[1:])
	case "query":
		return a.runQuery(ctx, args[1:])
	case "ingest":
		return a.runIngest(ctx, args[1:])
	default:
		fmt.Fprintf(a.stderr, "error: unknown command %q\n\n", args[0])
		a.writeRootUsage(a.stderr)
		return 2
	}
}

func (a *App) runHelp(command string) int {
	switch command {
	case "version":
		a.writeVersionUsage(a.stdout)
	case "status":
		a.writeStatusUsage(a.stdout)
	case "repos":
		a.writeReposUsage(a.stdout)
	case "query":
		a.writeQueryUsage(a.stdout)
	case "ingest":
		a.writeIngestUsage(a.stdout)
	default:
		fmt.Fprintf(a.stderr, "error: unknown command %q\n\n", command)
		a.writeRootUsage(a.stderr)
		return 2
	}

	return 0
}

func (a *App) runVersion(ctx context.Context, args []string) int {
	common := CommonOptionsRaw{}
	fs := newFlagSet("version", a.writeVersionUsage)
	bindCommonFlags(fs, &common)
	if code, ok := a.parseFlags(fs, args); ok {
		return code
	}

	settings, err := ResolveSettings(common, visitedFlags(fs))
	if err != nil {
		return a.renderError("version", resolveRequestedOutputMode(common.Output), nil, err)
	}

	client := NewClient(settings, "")
	serverVersion, err := client.GetVersion(ctx)
	if err != nil {
		return a.renderError("version", settings.Output, &settings, err)
	}

	output := VersionCommandOutput{
		Command:       "version",
		Server:        settings.ServerBaseURL,
		APIVersion:    chooseAPIVersion(serverVersion.APIVersion, settings.APIVersion),
		CLI:           currentLocalVersion(),
		ServerVersion: serverVersion,
	}

	return a.renderOutput(settings.Output, output, func() error {
		return writeVersionText(a.stdout, output)
	})
}

func (a *App) runRepos(ctx context.Context, args []string) int {
	common := CommonOptionsRaw{}
	fs := newFlagSet("repos", a.writeReposUsage)
	bindCommonFlags(fs, &common)
	if code, ok := a.parseFlags(fs, args); ok {
		return code
	}

	settings, err := ResolveSettings(common, visitedFlags(fs))
	if err != nil {
		return a.renderError("repos", resolveRequestedOutputMode(common.Output), nil, err)
	}

	token, err := settings.ResolveBearerToken(true)
	if err != nil {
		return a.renderError("repos", settings.Output, &settings, err)
	}

	client := NewClient(settings, token)
	repos, err := client.ListRepos(ctx)
	if err != nil {
		return a.renderError("repos", settings.Output, &settings, err)
	}

	sort.SliceStable(repos, func(i int, j int) bool {
		return strings.ToLower(repos[i].Name) < strings.ToLower(repos[j].Name)
	})

	output := ReposCommandOutput{
		Command:    "repos",
		Server:     settings.ServerBaseURL,
		APIVersion: settings.APIVersion,
		Repos:      repos,
	}

	return a.renderOutput(settings.Output, output, func() error {
		return writeReposText(a.stdout, output)
	})
}

func (a *App) runQuery(ctx context.Context, args []string) int {
	common := CommonOptionsRaw{}
	var (
		repoID             int64
		topK               int
		pathPrefix         string
		author             string
		commit             string
		since              string
		until              string
		expandGraph        bool
		graphDepth         int
		graphMaxCandidates int
		graphSeedLimit     int
		chunkTypes         stringListFlag
	)

	fs := newFlagSet("query", a.writeQueryUsage)
	bindCommonFlags(fs, &common)
	fs.Int64Var(&repoID, "repo-id", 0, "Target repository ID. Defaults to GITSEMANTIC_REPO_ID or the server's single indexed repo.")
	fs.IntVar(&topK, "topk", 0, "Maximum number of results to request.")
	fs.Var(&chunkTypes, "chunk-type", "Filter by chunk type. Repeat or provide a comma-separated list.")
	fs.StringVar(&pathPrefix, "path-prefix", "", "Restrict results to paths with this prefix.")
	fs.StringVar(&author, "author", "", "Restrict results to a commit author.")
	fs.StringVar(&commit, "commit", "", "Restrict results to a commit hash or prefix.")
	fs.StringVar(&since, "since", "", "Only include results on or after this timestamp. Accepts RFC3339 or YYYY-MM-DD.")
	fs.StringVar(&until, "until", "", "Only include results on or before this timestamp. Accepts RFC3339 or YYYY-MM-DD.")
	fs.BoolVar(&expandGraph, "expand-graph", false, "Enable graph expansion.")
	fs.IntVar(&graphDepth, "graph-depth", 0, "Graph expansion depth.")
	fs.IntVar(&graphMaxCandidates, "graph-max-candidates", 0, "Graph expansion max candidate count.")
	fs.IntVar(&graphSeedLimit, "graph-seed-limit", 0, "Graph expansion seed limit.")
	if code, ok := a.parseFlags(fs, args); ok {
		return code
	}

	settings, err := ResolveSettings(common, visitedFlags(fs))
	if err != nil {
		return a.renderError("query", resolveRequestedOutputMode(common.Output), nil, err)
	}

	queryText := strings.TrimSpace(strings.Join(fs.Args(), " "))
	if queryText == "" {
		return a.renderError("query", settings.Output, &settings, errors.New("query text is required"))
	}

	token, err := settings.ResolveBearerToken(true)
	if err != nil {
		return a.renderError("query", settings.Output, &settings, err)
	}

	request := QueryRequest{
		QueryText: queryText,
	}

	visited := visitedFlags(fs)
	if visited["repo-id"] {
		if repoID <= 0 {
			return a.renderError("query", settings.Output, &settings, errors.New("repo-id must be greater than zero"))
		}

		request.RepoID = &repoID
	} else if settings.DefaultRepoID != nil {
		request.RepoID = settings.DefaultRepoID
	}

	if visited["topk"] {
		if topK <= 0 || topK > maxQueryTopK {
			return a.renderError("query", settings.Output, &settings, fmt.Errorf("topk must be between 1 and %d", maxQueryTopK))
		}

		request.TopK = &topK
	}

	if visited["chunk-type"] {
		request.ChunkTypes = chunkTypes.Values()
	}

	if visited["path-prefix"] {
		request.PathPrefix = strings.TrimSpace(pathPrefix)
	}

	if visited["author"] {
		request.Author = strings.TrimSpace(author)
	}

	if visited["commit"] {
		request.Commit = strings.TrimSpace(commit)
	}

	if visited["since"] {
		parsed, err := parseTimestamp(since, false)
		if err != nil {
			return a.renderError("query", settings.Output, &settings, fmt.Errorf("invalid --since value: %w", err))
		}

		request.Since = parsed
	}

	if visited["until"] {
		parsed, err := parseTimestamp(until, true)
		if err != nil {
			return a.renderError("query", settings.Output, &settings, fmt.Errorf("invalid --until value: %w", err))
		}

		request.Until = parsed
	}

	var graphExpansion QueryGraphExpansionRequest
	graphConfigured := false
	if visited["expand-graph"] {
		graphConfigured = true
		graphExpansion.Enabled = boolPointer(expandGraph)
	}

	if visited["graph-depth"] {
		if graphDepth <= 0 {
			return a.renderError("query", settings.Output, &settings, errors.New("graph-depth must be greater than zero"))
		}

		graphConfigured = true
		graphExpansion.Depth = intPointer(graphDepth)
	}

	if visited["graph-max-candidates"] {
		if graphMaxCandidates <= 0 {
			return a.renderError("query", settings.Output, &settings, errors.New("graph-max-candidates must be greater than zero"))
		}

		graphConfigured = true
		graphExpansion.MaxCandidates = intPointer(graphMaxCandidates)
	}

	if visited["graph-seed-limit"] {
		if graphSeedLimit <= 0 {
			return a.renderError("query", settings.Output, &settings, errors.New("graph-seed-limit must be greater than zero"))
		}

		graphConfigured = true
		graphExpansion.SeedLimit = intPointer(graphSeedLimit)
	}

	if graphConfigured {
		request.GraphExpansion = &graphExpansion
	}

	client := NewClient(settings, token)
	response, err := client.Query(ctx, request)
	if err != nil {
		return a.renderError("query", settings.Output, &settings, err)
	}

	output := QueryCommandOutput{
		Command:    "query",
		Server:     settings.ServerBaseURL,
		APIVersion: chooseAPIVersion(response.APIVersion, settings.APIVersion),
		Response:   *response,
	}

	return a.renderOutput(settings.Output, output, func() error {
		return writeQueryText(a.stdout, output)
	})
}

func (a *App) runIngest(ctx context.Context, args []string) int {
	common := CommonOptionsRaw{}
	var (
		repoPath string
		repoURL  string
		branch   string
		mode     string
	)

	fs := newFlagSet("ingest", a.writeIngestUsage)
	bindCommonFlags(fs, &common)
	fs.StringVar(&repoPath, "repo", "", "Repository path as seen by the server. For the local Docker dev container, use /repo.")
	fs.StringVar(&repoURL, "url", "", "Repository URL metadata to associate with the ingest request.")
	fs.StringVar(&branch, "branch", "", "Branch or ref to scope ingestion.")
	fs.StringVar(&mode, "mode", "", "Ingestion mode: Commits, Chunks, Embeddings, Issues, or All.")
	if code, ok := a.parseFlags(fs, args); ok {
		return code
	}

	settings, err := ResolveSettings(common, visitedFlags(fs))
	if err != nil {
		return a.renderError("ingest", resolveRequestedOutputMode(common.Output), nil, err)
	}

	resolvedRepoPath := strings.TrimSpace(repoPath)
	if resolvedRepoPath == "" {
		return a.renderError("ingest", settings.Output, &settings, errors.New("repo is required"))
	}

	token, err := settings.ResolveBearerToken(true)
	if err != nil {
		return a.renderError("ingest", settings.Output, &settings, err)
	}

	request := IngestRequest{
		RepoPath: resolvedRepoPath,
		RepoURL:  strings.TrimSpace(repoURL),
		Branch:   strings.TrimSpace(branch),
	}

	if strings.TrimSpace(mode) != "" {
		canonicalMode, err := canonicalizeIngestionMode(mode)
		if err != nil {
			return a.renderError("ingest", settings.Output, &settings, err)
		}

		request.Mode = canonicalMode
	}

	client := NewClient(settings, token)
	response, err := client.Ingest(ctx, request)
	if err != nil {
		return a.renderError("ingest", settings.Output, &settings, err)
	}

	output := IngestCommandOutput{
		Command:           "ingest",
		Server:            settings.ServerBaseURL,
		APIVersion:        settings.APIVersion,
		RequestedRepoPath: strings.TrimSpace(repoPath),
		ResolvedRepoPath:  resolvedRepoPath,
		Response:          *response,
	}

	return a.renderOutput(settings.Output, output, func() error {
		return writeIngestText(a.stdout, output)
	})
}

func (a *App) runStatus(ctx context.Context, args []string) int {
	common := CommonOptionsRaw{}
	var jobID string

	fs := newFlagSet("status", a.writeStatusUsage)
	bindCommonFlags(fs, &common)
	fs.StringVar(&jobID, "job", "", "Optional ingestion job ID to inspect.")
	if code, ok := a.parseFlags(fs, args); ok {
		return code
	}

	settings, err := ResolveSettings(common, visitedFlags(fs))
	if err != nil {
		return a.renderError("status", resolveRequestedOutputMode(common.Output), nil, err)
	}

	publicClient := NewClient(settings, "")
	health, err := publicClient.GetHealth(ctx)
	if err != nil {
		return a.renderError("status", settings.Output, &settings, err)
	}

	output := StatusCommandOutput{
		Command:    "status",
		Server:     settings.ServerBaseURL,
		APIVersion: chooseAPIVersion(health.APIVersion, settings.APIVersion),
		CLI:        currentLocalVersion(),
		Health:     health,
	}

	var warnings []string
	if !health.DatabaseReady || !strings.EqualFold(health.Status, "ok") {
		warnings = append(warnings, "database readiness is degraded")
	}

	if !health.ModelAccessConfigured {
		warnings = append(warnings, "model access is not configured; /api/query currently returns 503")
	}

	token, tokenErr := settings.ResolveBearerToken(false)
	if tokenErr != nil {
		warnings = append(warnings, tokenErr.Error())
	} else if token == "" {
		warnings = append(warnings, "no bearer token configured; protected repo and ingestion endpoints were skipped")
	}

	if token != "" {
		protectedClient := NewClient(settings, token)
		repos, err := protectedClient.ListRepos(ctx)
		if err != nil {
			return a.renderError("status", settings.Output, &settings, err)
		}

		sort.SliceStable(repos, func(i int, j int) bool {
			return strings.ToLower(repos[i].Name) < strings.ToLower(repos[j].Name)
		})

		output.Repos = repos
		output.RepoTotals = computeRepoTotals(repos)

		if strings.TrimSpace(jobID) != "" {
			job, err := protectedClient.GetIngestionStatus(ctx, strings.TrimSpace(jobID))
			if err != nil {
				return a.renderError("status", settings.Output, &settings, err)
			}

			output.Job = job
		}
	} else if strings.TrimSpace(jobID) != "" {
		warnings = append(warnings, "job status requires a bearer token; --job was skipped")
	}

	output.Warnings = warnings
	return a.renderOutput(settings.Output, output, func() error {
		return writeStatusText(a.stdout, output)
	})
}

func (a *App) parseFlags(fs *flag.FlagSet, args []string) (int, bool) {
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fs.Usage()
			return 0, true
		}

		fmt.Fprintf(a.stderr, "error: %v\n\n", err)
		fs.Usage()
		return 2, true
	}

	return 0, false
}

func (a *App) renderOutput(outputMode OutputMode, jsonValue any, textWriter func() error) int {
	var err error
	if outputMode == OutputJSON {
		err = writeJSON(a.stdout, jsonValue)
	} else {
		err = textWriter()
	}

	if err != nil {
		fmt.Fprintf(a.stderr, "error: failed to write output: %v\n", err)
		return 1
	}

	return 0
}

func (a *App) renderError(command string, outputMode OutputMode, settings *ResolvedSettings, err error) int {
	if outputMode == OutputJSON {
		payload := CommandErrorOutput{
			Command: command,
			Error:   err.Error(),
		}

		var apiError *APIError
		if errors.As(err, &apiError) {
			payload.StatusCode = apiError.StatusCode
		}

		if settings != nil {
			payload.Server = settings.ServerBaseURL
			payload.APIVersion = settings.APIVersion
		}

		if writeErr := writeJSON(a.stdout, payload); writeErr == nil {
			return 1
		}
	}

	fmt.Fprintf(a.stderr, "error: %v\n", err)
	return 1
}

func (a *App) writeRootUsage(output io.Writer) {
	fmt.Fprintln(output, "GitSemantic CLI")
	fmt.Fprintln(output, "Part of GoblinBoard semantic search tooling.")
	fmt.Fprintln(output, "Website: https://goblinboard.com")
	fmt.Fprintln(output, "Semantic Search: https://goblinboard.com/semantic-search")
	fmt.Fprintln(output, "Status: active development")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Usage:")
	fmt.Fprintln(output, "  gitsemantic <command> [flags]")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Commands:")
	fmt.Fprintln(output, "  status   Check server health and summarize indexed repositories.")
	fmt.Fprintln(output, "  version  Show CLI, server, and engine versions.")
	fmt.Fprintln(output, "  repos    List indexed repositories.")
	fmt.Fprintln(output, "  ingest   Queue repository ingestion.")
	fmt.Fprintln(output, "  query    Run semantic search.")
	fmt.Fprintln(output, "  help     Show command help.")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Config resolution order: environment -> config file -> flags.")
	fmt.Fprintln(output, "Default config file: ~/.gitsemantic/config.yaml")
	fmt.Fprintln(output, "Default token file: ~/.gitsemantic/token")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Environment variables:")
	fmt.Fprintln(output, "  GITSEMANTIC_SERVER")
	fmt.Fprintln(output, "  GITSEMANTIC_TOKEN")
	fmt.Fprintln(output, "  GITSEMANTIC_TOKEN_FILE")
	fmt.Fprintln(output, "  GITSEMANTIC_REPO_ID")
	fmt.Fprintln(output, "  GITSEMANTIC_API_VERSION")
	fmt.Fprintln(output, "  GITSEMANTIC_OUTPUT")
	fmt.Fprintln(output, "  GITSEMANTIC_CONFIG")
}

func (a *App) writeVersionUsage(output io.Writer) {
	fmt.Fprintln(output, "Usage:")
	fmt.Fprintln(output, "  gitsemantic version [common flags]")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Show the local CLI version plus /api/version from the server.")
	fmt.Fprintln(output)
	a.writeCommonFlags(output)
}

func (a *App) writeStatusUsage(output io.Writer) {
	fmt.Fprintln(output, "Usage:")
	fmt.Fprintln(output, "  gitsemantic status [--job <id>] [common flags]")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Call /api/health and, when a bearer token is available, /api/repos and optionally /api/ingest/{jobId}.")
	fmt.Fprintln(output)
	a.writeCommonFlags(output)
	fmt.Fprintln(output, "Status flags:")
	fmt.Fprintln(output, "  --job string     Optional ingestion job ID to inspect.")
}

func (a *App) writeReposUsage(output io.Writer) {
	fmt.Fprintln(output, "Usage:")
	fmt.Fprintln(output, "  gitsemantic repos [common flags]")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "List indexed repositories from /api/repos.")
	fmt.Fprintln(output)
	a.writeCommonFlags(output)
}

func (a *App) writeQueryUsage(output io.Writer) {
	fmt.Fprintln(output, "Usage:")
	fmt.Fprintln(output, "  gitsemantic query [flags] <query text>")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Run semantic search against /api/query.")
	fmt.Fprintln(output)
	a.writeCommonFlags(output)
	fmt.Fprintln(output, "Query flags:")
	fmt.Fprintln(output, "  --repo-id int                 Target repository ID.")
	fmt.Fprintln(output, "  --topk int                    Maximum number of results (max 50).")
	fmt.Fprintln(output, "  --chunk-type value            Repeat or comma-separate chunk types.")
	fmt.Fprintln(output, "  --path-prefix string          Filter by canonical path prefix.")
	fmt.Fprintln(output, "  --author string               Filter by author.")
	fmt.Fprintln(output, "  --commit string               Filter by commit hash or prefix.")
	fmt.Fprintln(output, "  --since string                RFC3339 or YYYY-MM-DD lower bound.")
	fmt.Fprintln(output, "  --until string                RFC3339 or YYYY-MM-DD upper bound.")
	fmt.Fprintln(output, "  --expand-graph                Enable graph expansion.")
	fmt.Fprintln(output, "  --graph-depth int             Graph expansion depth.")
	fmt.Fprintln(output, "  --graph-max-candidates int    Graph expansion max candidate count.")
	fmt.Fprintln(output, "  --graph-seed-limit int        Graph expansion seed limit.")
}

func (a *App) writeIngestUsage(output io.Writer) {
	fmt.Fprintln(output, "Usage:")
	fmt.Fprintln(output, "  gitsemantic ingest --repo <path> [flags]")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Queue ingestion through /api/ingest.")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Current dev-container caveat: repo path must match the container filesystem, usually /repo.")
	fmt.Fprintln(output)
	a.writeCommonFlags(output)
	fmt.Fprintln(output, "Ingest flags:")
	fmt.Fprintln(output, "  --repo string     Repository path as seen by the server.")
	fmt.Fprintln(output, "  --url string      Repository URL metadata.")
	fmt.Fprintln(output, "  --branch string   Branch or ref to ingest.")
	fmt.Fprintln(output, "  --mode string     Commits, Chunks, Embeddings, Issues, or All.")
}

func (a *App) writeCommonFlags(output io.Writer) {
	fmt.Fprintln(output, "Common flags:")
	fmt.Fprintln(output, "  --config string       Path to config file. Default: ~/.gitsemantic/config.yaml")
	fmt.Fprintln(output, "  --server string       Base URL of the GitSemantic server. Default: http://127.0.0.1:7280")
	fmt.Fprintln(output, "  --token string        Bearer token for protected endpoints.")
	fmt.Fprintln(output, "  --token-file string   Path to a bearer token file. Default: ~/.gitsemantic/token")
	fmt.Fprintln(output, "  --api-version string  API version to pin. Default: 1")
	fmt.Fprintln(output, "  --output string       text or json. Default: text")
	fmt.Fprintln(output)
}

func newFlagSet(name string, usage func(io.Writer)) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Usage = func() {
		usage(os.Stdout)
	}

	return fs
}

func parseTimestamp(raw string, inclusiveEndOfDay bool) (*time.Time, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil, nil
	}

	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		normalized := parsed.UTC()
		return &normalized, nil
	}

	date, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil, fmt.Errorf("expected RFC3339 or YYYY-MM-DD")
	}

	normalized := date.UTC()
	if inclusiveEndOfDay {
		normalized = normalized.Add((24 * time.Hour) - time.Nanosecond)
	}

	return &normalized, nil
}

func canonicalizeIngestionMode(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	switch strings.ToLower(trimmed) {
	case "commits":
		return "Commits", nil
	case "chunks":
		return "Chunks", nil
	case "embeddings":
		return "Embeddings", nil
	case "issues":
		return "Issues", nil
	case "all":
		return "All", nil
	default:
		return "", fmt.Errorf("unsupported ingestion mode %q; expected one of: Commits, Chunks, Embeddings, Issues, All", raw)
	}
}

func computeRepoTotals(repos []RepoSummary) *RepoTotals {
	if len(repos) == 0 {
		return &RepoTotals{}
	}

	totals := &RepoTotals{
		Count: len(repos),
	}

	for _, repo := range repos {
		totals.CommitCount += repo.CommitCount
		totals.ChunkCount += repo.ChunkCount
		totals.EmbeddedChunkCount += repo.EmbeddedChunkCount
		totals.IssueCount += repo.IssueCount
	}

	return totals
}

func chooseAPIVersion(responseVersion string, fallback string) string {
	if strings.TrimSpace(responseVersion) != "" {
		return responseVersion
	}

	return fallback
}

func boolPointer(value bool) *bool {
	return &value
}

func intPointer(value int) *int {
	return &value
}
