package main

var (
	cliVersion = "dev"
	cliCommit  = "unknown"
	cliDate    = "unknown"
)

type LocalVersionInfo struct {
	Version string `json:"version"`
	Commit  string `json:"commit,omitempty"`
	Date    string `json:"date,omitempty"`
}

func currentLocalVersion() LocalVersionInfo {
	return LocalVersionInfo{
		Version: cliVersion,
		Commit:  cliCommit,
		Date:    cliDate,
	}
}
