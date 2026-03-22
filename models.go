package main

import "time"

type OutputMode string

const (
	OutputText OutputMode = "text"
	OutputJSON OutputMode = "json"
)

type HealthResponse struct {
	Status                string    `json:"status"`
	APIVersion            string    `json:"apiVersion"`
	ServerTimeUTC         time.Time `json:"serverTimeUtc"`
	DatabaseReady         bool      `json:"databaseReady"`
	ModelAccessConfigured bool      `json:"modelAccessConfigured"`
	ModelAccessError      string    `json:"modelAccessError,omitempty"`
}

type VersionResponse struct {
	APIVersion           string   `json:"apiVersion"`
	SupportedAPIVersions []string `json:"supportedApiVersions"`
	ServerVersion        string   `json:"serverVersion"`
	EngineVersion        string   `json:"engineVersion"`
}

type RepoStageState struct {
	Stage                  string     `json:"stage"`
	LastSequence           int64      `json:"lastSequence"`
	LastCommitHash         string     `json:"lastCommitHash,omitempty"`
	LastCommitTimestampUTC *time.Time `json:"lastCommitTimestampUtc,omitempty"`
	LastIngestedAtUTC      *time.Time `json:"lastIngestedAtUtc,omitempty"`
}

type RepoSummary struct {
	RepoID             int64            `json:"repoId"`
	Name               string           `json:"name"`
	RepoURL            string           `json:"repoUrl"`
	DefaultBranch      string           `json:"defaultBranch,omitempty"`
	CommitCount        int              `json:"commitCount"`
	ChunkCount         int              `json:"chunkCount"`
	EmbeddedChunkCount int              `json:"embeddedChunkCount"`
	IssueCount         int              `json:"issueCount"`
	LastIngestedAtUTC  *time.Time       `json:"lastIngestedAtUtc,omitempty"`
	Stages             []RepoStageState `json:"stages"`
}

type QueryGraphExpansionRequest struct {
	Enabled       *bool `json:"enabled,omitempty"`
	Depth         *int  `json:"depth,omitempty"`
	MaxCandidates *int  `json:"maxCandidates,omitempty"`
	SeedLimit     *int  `json:"seedLimit,omitempty"`
}

type QueryRequest struct {
	QueryText      string                      `json:"queryText"`
	RepoID         *int64                      `json:"repoId,omitempty"`
	TopK           *int                        `json:"topK,omitempty"`
	ChunkTypes     []string                    `json:"chunkTypes,omitempty"`
	PathPrefix     string                      `json:"pathPrefix,omitempty"`
	Author         string                      `json:"author,omitempty"`
	Commit         string                      `json:"commit,omitempty"`
	Since          *time.Time                  `json:"since,omitempty"`
	Until          *time.Time                  `json:"until,omitempty"`
	GraphExpansion *QueryGraphExpansionRequest `json:"graphExpansion,omitempty"`
}

type QueryResult struct {
	ChunkID            int64      `json:"chunkId"`
	ChunkType          string     `json:"chunkType"`
	Path               string     `json:"path,omitempty"`
	CommitHash         string     `json:"commitHash"`
	CommitTimestampUTC *time.Time `json:"commitTimestampUtc,omitempty"`
	Author             string     `json:"author,omitempty"`
	Subject            string     `json:"subject,omitempty"`
	StartLine          *int       `json:"startLine,omitempty"`
	EndLine            *int       `json:"endLine,omitempty"`
	Snippet            string     `json:"snippet"`
	Score              float64    `json:"score"`
}

type QueryResponse struct {
	APIVersion  string        `json:"apiVersion"`
	RepoID      int64         `json:"repoId"`
	TopK        int           `json:"topK"`
	SortMode    string        `json:"sortMode"`
	ResultCount int           `json:"resultCount"`
	Results     []QueryResult `json:"results"`
}

type IngestRequest struct {
	RepoPath string `json:"repoPath"`
	RepoURL  string `json:"repoUrl,omitempty"`
	Branch   string `json:"branch,omitempty"`
	Mode     string `json:"mode,omitempty"`
}

type IngestAcceptedResponse struct {
	JobID     string `json:"jobId"`
	State     string `json:"state"`
	StatusURL string `json:"statusUrl"`
}

type IngestJobResponse struct {
	JobID          string     `json:"jobId"`
	State          string     `json:"state"`
	RepoPath       string     `json:"repoPath"`
	RepoURL        string     `json:"repoUrl"`
	Branch         string     `json:"branch,omitempty"`
	Mode           string     `json:"mode"`
	RepoID         *int64     `json:"repoId,omitempty"`
	CreatedAtUTC   time.Time  `json:"createdAtUtc"`
	StartedAtUTC   *time.Time `json:"startedAtUtc,omitempty"`
	CompletedAtUTC *time.Time `json:"completedAtUtc,omitempty"`
	Error          string     `json:"error,omitempty"`
}

type VersionCommandOutput struct {
	Command       string           `json:"command"`
	Server        string           `json:"server"`
	APIVersion    string           `json:"apiVersion"`
	CLI           LocalVersionInfo `json:"cli"`
	ServerVersion *VersionResponse `json:"serverVersion,omitempty"`
}

type RepoTotals struct {
	Count              int `json:"count"`
	CommitCount        int `json:"commitCount"`
	ChunkCount         int `json:"chunkCount"`
	EmbeddedChunkCount int `json:"embeddedChunkCount"`
	IssueCount         int `json:"issueCount"`
}

type StatusCommandOutput struct {
	Command    string             `json:"command"`
	Server     string             `json:"server"`
	APIVersion string             `json:"apiVersion"`
	CLI        LocalVersionInfo   `json:"cli"`
	Health     *HealthResponse    `json:"health,omitempty"`
	Repos      []RepoSummary      `json:"repos,omitempty"`
	RepoTotals *RepoTotals        `json:"repoTotals,omitempty"`
	Job        *IngestJobResponse `json:"job,omitempty"`
	Warnings   []string           `json:"warnings,omitempty"`
}

type IngestCommandOutput struct {
	Command           string                 `json:"command"`
	Server            string                 `json:"server"`
	APIVersion        string                 `json:"apiVersion"`
	RequestedRepoPath string                 `json:"requestedRepoPath"`
	ResolvedRepoPath  string                 `json:"resolvedRepoPath"`
	Response          IngestAcceptedResponse `json:"response"`
}

type ReposCommandOutput struct {
	Command    string        `json:"command"`
	Server     string        `json:"server"`
	APIVersion string        `json:"apiVersion"`
	Repos      []RepoSummary `json:"repos"`
}

type QueryCommandOutput struct {
	Command    string        `json:"command"`
	Server     string        `json:"server"`
	APIVersion string        `json:"apiVersion"`
	Response   QueryResponse `json:"response"`
}

type CommandErrorOutput struct {
	Command    string `json:"command"`
	Server     string `json:"server,omitempty"`
	APIVersion string `json:"apiVersion,omitempty"`
	StatusCode int    `json:"statusCode,omitempty"`
	Error      string `json:"error"`
}
