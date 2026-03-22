package main

import (
	"flag"
	"strings"
)

type CommonOptionsRaw struct {
	ConfigPath string
	Server     string
	Token      string
	TokenFile  string
	APIVersion string
	Output     string
}

func bindCommonFlags(fs *flag.FlagSet, common *CommonOptionsRaw) {
	fs.StringVar(&common.ConfigPath, "config", "", "Path to the CLI config file.")
	fs.StringVar(&common.Server, "server", "", "Base URL of the GitSemantic server.")
	fs.StringVar(&common.Token, "token", "", "Bearer token for protected API endpoints.")
	fs.StringVar(&common.TokenFile, "token-file", "", "Path to a file containing the bearer token.")
	fs.StringVar(&common.APIVersion, "api-version", "", "API version to pin on requests. Defaults to 1.")
	fs.StringVar(&common.Output, "output", "", "Output format: text or json.")
}

func visitedFlags(fs *flag.FlagSet) map[string]bool {
	visited := make(map[string]bool)
	fs.Visit(func(f *flag.Flag) {
		visited[f.Name] = true
	})

	return visited
}

type stringListFlag struct {
	values []string
}

func (f *stringListFlag) String() string {
	return strings.Join(f.values, ",")
}

func (f *stringListFlag) Set(value string) error {
	for _, item := range strings.Split(value, ",") {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}

		f.values = append(f.values, trimmed)
	}

	return nil
}

func (f *stringListFlag) Values() []string {
	if len(f.values) == 0 {
		return nil
	}

	values := make([]string, 0, len(f.values))
	seen := make(map[string]bool)
	for _, value := range f.values {
		key := strings.ToLower(value)
		if seen[key] {
			continue
		}

		seen[key] = true
		values = append(values, value)
	}

	return values
}
