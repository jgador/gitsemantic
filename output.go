package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"
)

func writeJSON(output io.Writer, value any) error {
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func writeVersionText(output io.Writer, value VersionCommandOutput) error {
	_, err := fmt.Fprintf(
		output,
		"CLI\t%s\nCommit\t%s\nBuilt\t%s\nServer\t%s\nAPI Version\t%s\nServer Version\t%s\nEngine Version\t%s\n",
		value.CLI.Version,
		value.CLI.Commit,
		value.CLI.Date,
		value.Server,
		value.APIVersion,
		formatVersionField(value.ServerVersion, func(v *VersionResponse) string { return v.ServerVersion }),
		formatVersionField(value.ServerVersion, func(v *VersionResponse) string { return v.EngineVersion }),
	)
	return err
}

func writeReposText(output io.Writer, value ReposCommandOutput) error {
	if _, err := fmt.Fprintf(output, "Server: %s\nAPI Version: %s\n", value.Server, value.APIVersion); err != nil {
		return err
	}

	if len(value.Repos) == 0 {
		_, err := fmt.Fprintln(output, "No indexed repositories found.")
		return err
	}

	writer := tabwriter.NewWriter(output, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(writer, "ID\tName\tDefault Branch\tCommits\tChunks\tEmbedded\tIssues\tLast Ingested"); err != nil {
		return err
	}

	for _, repo := range value.Repos {
		if _, err := fmt.Fprintf(
			writer,
			"%d\t%s\t%s\t%d\t%d\t%d\t%d\t%s\n",
			repo.RepoID,
			blankIfEmpty(repo.Name),
			blankIfEmpty(repo.DefaultBranch),
			repo.CommitCount,
			repo.ChunkCount,
			repo.EmbeddedChunkCount,
			repo.IssueCount,
			formatTime(repo.LastIngestedAtUTC),
		); err != nil {
			return err
		}
	}

	return writer.Flush()
}

func writeQueryText(output io.Writer, value QueryCommandOutput) error {
	if _, err := fmt.Fprintf(
		output,
		"Server: %s\nAPI Version: %s\nRepo ID: %d\nSort Mode: %s\nResults: %d of top %d\n",
		value.Server,
		value.APIVersion,
		value.Response.RepoID,
		blankIfEmpty(value.Response.SortMode),
		value.Response.ResultCount,
		value.Response.TopK,
	); err != nil {
		return err
	}

	if len(value.Response.Results) == 0 {
		_, err := fmt.Fprintln(output, "No matches found.")
		return err
	}

	for index, result := range value.Response.Results {
		if _, err := fmt.Fprintf(
			output,
			"\n%d. score=%.4f type=%s path=%s lines=%s commit=%s\n",
			index+1,
			result.Score,
			result.ChunkType,
			blankIfEmpty(result.Path),
			formatLineRange(result.StartLine, result.EndLine),
			shortCommitHash(result.CommitHash),
		); err != nil {
			return err
		}

		if result.Author != "" {
			if _, err := fmt.Fprintf(output, "   author: %s\n", result.Author); err != nil {
				return err
			}
		}

		if result.Subject != "" {
			if _, err := fmt.Fprintf(output, "   subject: %s\n", result.Subject); err != nil {
				return err
			}
		}

		if result.CommitTimestampUTC != nil {
			if _, err := fmt.Fprintf(output, "   commit timestamp: %s\n", formatTime(result.CommitTimestampUTC)); err != nil {
				return err
			}
		}

		if _, err := fmt.Fprintf(output, "   snippet:\n%s\n", indentBlock(result.Snippet, "     ")); err != nil {
			return err
		}
	}

	return nil
}

func writeIngestText(output io.Writer, value IngestCommandOutput) error {
	_, err := fmt.Fprintf(
		output,
		"Server: %s\nAPI Version: %s\nRequested Repo Path: %s\nRepo Path Sent: %s\nJob ID: %s\nState: %s\nStatus URL: %s\n",
		value.Server,
		value.APIVersion,
		value.RequestedRepoPath,
		value.ResolvedRepoPath,
		value.Response.JobID,
		value.Response.State,
		value.Response.StatusURL,
	)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(output, "Note: the current Docker dev server expects a container path such as /repo for ingestion.")
	return err
}

func writeStatusText(output io.Writer, value StatusCommandOutput) error {
	if _, err := fmt.Fprintf(
		output,
		"Server: %s\nAPI Version: %s\nCLI Version: %s\nHealth: %s\nDatabase Ready: %t\nModel Access Configured: %t\n",
		value.Server,
		value.APIVersion,
		value.CLI.Version,
		statusLabel(value.Health),
		value.Health != nil && value.Health.DatabaseReady,
		value.Health != nil && value.Health.ModelAccessConfigured,
	); err != nil {
		return err
	}

	if value.Health != nil && value.Health.ModelAccessError != "" {
		if _, err := fmt.Fprintf(output, "Model Access Error: %s\n", value.Health.ModelAccessError); err != nil {
			return err
		}
	}

	if value.RepoTotals != nil {
		if _, err := fmt.Fprintf(
			output,
			"Repo Count: %d\nTotal Commits: %d\nTotal Chunks: %d\nEmbedded Chunks: %d\nTotal Issues: %d\n",
			value.RepoTotals.Count,
			value.RepoTotals.CommitCount,
			value.RepoTotals.ChunkCount,
			value.RepoTotals.EmbeddedChunkCount,
			value.RepoTotals.IssueCount,
		); err != nil {
			return err
		}
	}

	if len(value.Warnings) > 0 {
		if _, err := fmt.Fprintln(output, "Warnings:"); err != nil {
			return err
		}

		for _, warning := range value.Warnings {
			if _, err := fmt.Fprintf(output, "- %s\n", warning); err != nil {
				return err
			}
		}
	}

	if value.Job != nil {
		if _, err := fmt.Fprintf(
			output,
			"Ingestion Job: %s\nJob State: %s\nJob Mode: %s\nJob Repo Path: %s\nCreated: %s\nStarted: %s\nCompleted: %s\n",
			value.Job.JobID,
			value.Job.State,
			value.Job.Mode,
			value.Job.RepoPath,
			formatTime(&value.Job.CreatedAtUTC),
			formatTime(value.Job.StartedAtUTC),
			formatTime(value.Job.CompletedAtUTC),
		); err != nil {
			return err
		}

		if value.Job.RepoID != nil {
			if _, err := fmt.Fprintf(output, "Indexed Repo ID: %d\n", *value.Job.RepoID); err != nil {
				return err
			}
		}

		if value.Job.Error != "" {
			if _, err := fmt.Fprintf(output, "Job Error: %s\n", value.Job.Error); err != nil {
				return err
			}
		}
	}

	if len(value.Repos) == 0 {
		return nil
	}

	if _, err := fmt.Fprintln(output, "\nRepositories:"); err != nil {
		return err
	}

	return writeReposTable(output, value.Repos)
}

func writeReposTable(output io.Writer, repos []RepoSummary) error {
	writer := tabwriter.NewWriter(output, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(writer, "ID\tName\tCommits\tChunks\tEmbedded\tIssues\tLast Ingested"); err != nil {
		return err
	}

	for _, repo := range repos {
		if _, err := fmt.Fprintf(
			writer,
			"%d\t%s\t%d\t%d\t%d\t%d\t%s\n",
			repo.RepoID,
			blankIfEmpty(repo.Name),
			repo.CommitCount,
			repo.ChunkCount,
			repo.EmbeddedChunkCount,
			repo.IssueCount,
			formatTime(repo.LastIngestedAtUTC),
		); err != nil {
			return err
		}
	}

	return writer.Flush()
}

func formatVersionField(value *VersionResponse, selector func(*VersionResponse) string) string {
	if value == nil {
		return "-"
	}

	selected := selector(value)
	if selected == "" {
		return "-"
	}

	return selected
}

func formatTime(value *time.Time) string {
	if value == nil || value.IsZero() {
		return "-"
	}

	return value.UTC().Format(time.RFC3339)
}

func formatLineRange(start *int, end *int) string {
	switch {
	case start == nil && end == nil:
		return "-"
	case start != nil && end != nil:
		return fmt.Sprintf("%d-%d", *start, *end)
	case start != nil:
		return fmt.Sprintf("%d", *start)
	default:
		return fmt.Sprintf("%d", *end)
	}
}

func indentBlock(value string, indent string) string {
	if value == "" {
		return indent + "-"
	}

	lines := strings.Split(strings.TrimRight(value, "\n"), "\n")
	for index, line := range lines {
		lines[index] = indent + line
	}

	return strings.Join(lines, "\n")
}

func shortCommitHash(hash string) string {
	if len(hash) <= 12 {
		return hash
	}

	return hash[:12]
}

func blankIfEmpty(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}

	return value
}

func statusLabel(health *HealthResponse) string {
	if health == nil || strings.TrimSpace(health.Status) == "" {
		return "unknown"
	}

	return health.Status
}
