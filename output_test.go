package main

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestWriteQueryTextIncludesSortModeAndCommitTimestamp(t *testing.T) {
	commitTime := time.Date(2026, time.March, 21, 11, 59, 38, 0, time.UTC)
	output := QueryCommandOutput{
		Command:    "query",
		Server:     "http://127.0.0.1:7280",
		APIVersion: "1",
		Response: QueryResponse{
			APIVersion:  "1",
			RepoID:      7,
			TopK:        5,
			SortMode:    "recent_first",
			ResultCount: 1,
			Results: []QueryResult{
				{
					ChunkID:            42,
					ChunkType:          "hunk",
					Path:               "internal/query/search.go",
					CommitHash:         "abcdef1234567890",
					CommitTimestampUTC: &commitTime,
					Author:             "octocat",
					Subject:            "recent-first ranking",
					Snippet:            "return search results;",
					Score:              0.9876,
				},
			},
		},
	}

	var buffer bytes.Buffer
	if err := writeQueryText(&buffer, output); err != nil {
		t.Fatalf("writeQueryText returned error: %v", err)
	}

	text := buffer.String()
	assertContains(t, text, "Sort Mode: recent_first")
	assertContains(t, text, "score=0.9876")
	assertContains(t, text, "commit timestamp: 2026-03-21T11:59:38Z")
	assertContains(t, text, "subject: recent-first ranking")
}

func assertContains(t *testing.T, text string, expected string) {
	t.Helper()
	if !strings.Contains(text, expected) {
		t.Fatalf("expected output to contain %q\nfull output:\n%s", expected, text)
	}
}
