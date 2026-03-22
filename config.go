package main

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	defaultServerURL  = "http://127.0.0.1:7280"
	defaultAPIVersion = "1"

	configPathEnvironmentVariable = "GITSEMANTIC_CONFIG"
	serverEnvironmentVariable     = "GITSEMANTIC_SERVER"
	tokenEnvironmentVariable      = "GITSEMANTIC_TOKEN"
	tokenFileEnvironmentVariable  = "GITSEMANTIC_TOKEN_FILE"
	repoIDEnvironmentVariable     = "GITSEMANTIC_REPO_ID"
	apiVersionEnvironmentVariable = "GITSEMANTIC_API_VERSION"
	outputEnvironmentVariable     = "GITSEMANTIC_OUTPUT"
)

type ResolvedSettings struct {
	ServerBaseURL string
	APIVersion    string
	Output        OutputMode
	Token         string
	TokenFile     string
	ConfigPath    string
	DefaultRepoID *int64
}

type sourceConfig struct {
	ConfigPath       string
	HasConfigPath    bool
	Server           string
	HasServer        bool
	Token            string
	HasToken         bool
	TokenFile        string
	HasTokenFile     bool
	DefaultRepoID    int64
	HasDefaultRepoID bool
	APIVersion       string
	HasAPIVersion    bool
	Output           string
	HasOutput        bool
}

func ResolveSettings(raw CommonOptionsRaw, visited map[string]bool) (ResolvedSettings, error) {
	settings := defaultResolvedSettings()
	envConfig, err := readEnvironmentConfig()
	if err != nil {
		return ResolvedSettings{}, err
	}

	applySourceConfig(&settings, envConfig)
	configPath, readConfigFile := resolveConfigPath(raw, visited, envConfig)
	if configPath != "" {
		settings.ConfigPath = configPath
	}

	if readConfigFile {
		fileConfig, err := readConfigFileConfig(configPath)
		if err != nil {
			return ResolvedSettings{}, err
		}

		applySourceConfig(&settings, fileConfig)
	}

	flagConfig, err := readFlagConfig(raw, visited)
	if err != nil {
		return ResolvedSettings{}, err
	}

	applySourceConfig(&settings, flagConfig)
	serverBaseURL, err := normalizeServerBaseURL(settings.ServerBaseURL)
	if err != nil {
		return ResolvedSettings{}, err
	}

	outputMode, err := normalizeOutputMode(string(settings.Output))
	if err != nil {
		return ResolvedSettings{}, err
	}

	settings.ServerBaseURL = serverBaseURL
	settings.APIVersion = normalizeAPIVersion(settings.APIVersion)
	settings.Output = outputMode
	settings.Token = strings.TrimSpace(settings.Token)
	settings.TokenFile = normalizePath(settings.TokenFile)

	return settings, nil
}

func defaultResolvedSettings() ResolvedSettings {
	return ResolvedSettings{
		ServerBaseURL: defaultServerURL,
		APIVersion:    defaultAPIVersion,
		Output:        OutputText,
		TokenFile:     defaultTokenFilePath(),
		ConfigPath:    defaultConfigPath(),
	}
}

func (s ResolvedSettings) ResolveBearerToken(required bool) (string, error) {
	if trimmed := strings.TrimSpace(s.Token); trimmed != "" {
		return trimmed, nil
	}

	tokenFile := strings.TrimSpace(s.TokenFile)
	if tokenFile == "" {
		if required {
			return "", fmt.Errorf("no bearer token configured; use --token, --token-file, %s, or %s", tokenEnvironmentVariable, tokenFileEnvironmentVariable)
		}

		return "", nil
	}

	contents, err := os.ReadFile(tokenFile)
	if err != nil {
		if !required && os.IsNotExist(err) {
			return "", nil
		}

		return "", fmt.Errorf("failed to read bearer token file %q: %w", tokenFile, err)
	}

	token := strings.TrimSpace(string(contents))
	if token == "" {
		if required {
			return "", fmt.Errorf("bearer token file %q is empty", tokenFile)
		}

		return "", nil
	}

	return token, nil
}

func resolveRequestedOutputMode(rawOutput string) OutputMode {
	switch strings.ToLower(strings.TrimSpace(rawOutput)) {
	case string(OutputJSON):
		return OutputJSON
	default:
		return OutputText
	}
}

func readEnvironmentConfig() (sourceConfig, error) {
	config := sourceConfig{}
	if value, ok := os.LookupEnv(configPathEnvironmentVariable); ok {
		config.ConfigPath = normalizePath(value)
		config.HasConfigPath = true
	}

	if value, ok := os.LookupEnv(serverEnvironmentVariable); ok {
		config.Server = value
		config.HasServer = true
	}

	if value, ok := os.LookupEnv(tokenEnvironmentVariable); ok {
		config.Token = value
		config.HasToken = true
	}

	if value, ok := os.LookupEnv(tokenFileEnvironmentVariable); ok {
		config.TokenFile = normalizePath(value)
		config.HasTokenFile = true
	}

	if value, ok := os.LookupEnv(repoIDEnvironmentVariable); ok {
		repoID, err := parseOptionalInt64(value)
		if err != nil {
			return sourceConfig{}, fmt.Errorf("invalid %s value: %w", repoIDEnvironmentVariable, err)
		}

		if repoID != nil {
			config.DefaultRepoID = *repoID
		}

		config.HasDefaultRepoID = true
	}

	if value, ok := os.LookupEnv(apiVersionEnvironmentVariable); ok {
		config.APIVersion = value
		config.HasAPIVersion = true
	}

	if value, ok := os.LookupEnv(outputEnvironmentVariable); ok {
		config.Output = value
		config.HasOutput = true
	}

	return config, nil
}

func resolveConfigPath(raw CommonOptionsRaw, visited map[string]bool, envConfig sourceConfig) (string, bool) {
	if visited["config"] {
		return normalizePath(raw.ConfigPath), strings.TrimSpace(raw.ConfigPath) != ""
	}

	if envConfig.HasConfigPath {
		return normalizePath(envConfig.ConfigPath), strings.TrimSpace(envConfig.ConfigPath) != ""
	}

	return defaultConfigPath(), true
}

func readConfigFileConfig(configPath string) (sourceConfig, error) {
	if strings.TrimSpace(configPath) == "" {
		return sourceConfig{}, nil
	}

	contents, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return sourceConfig{}, nil
		}

		return sourceConfig{}, fmt.Errorf("failed to read config file %q: %w", configPath, err)
	}

	config := sourceConfig{}
	configDirectory := filepath.Dir(configPath)
	scanner := bufio.NewScanner(strings.NewReader(string(contents)))
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || line == "---" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := splitConfigLine(line)
		if !ok {
			if strings.HasSuffix(line, ":") {
				continue
			}

			return sourceConfig{}, fmt.Errorf("invalid config line %d in %q", lineNumber, configPath)
		}

		switch normalizeConfigKey(key) {
		case "server":
			config.Server = value
			config.HasServer = true
		case "token":
			config.Token = value
			config.HasToken = true
		case "token_file":
			config.TokenFile = resolveRelativePath(configDirectory, value)
			config.HasTokenFile = true
		case "repo_id":
			repoID, err := parseOptionalInt64(value)
			if err != nil {
				return sourceConfig{}, fmt.Errorf("invalid repo_id on line %d in %q: %w", lineNumber, configPath, err)
			}

			if repoID != nil {
				config.DefaultRepoID = *repoID
			}

			config.HasDefaultRepoID = true
		case "api_version":
			config.APIVersion = value
			config.HasAPIVersion = true
		case "output":
			config.Output = value
			config.HasOutput = true
		}
	}

	if err := scanner.Err(); err != nil {
		return sourceConfig{}, fmt.Errorf("failed to scan config file %q: %w", configPath, err)
	}

	return config, nil
}

func readFlagConfig(raw CommonOptionsRaw, visited map[string]bool) (sourceConfig, error) {
	config := sourceConfig{}
	if visited["config"] {
		config.ConfigPath = normalizePath(raw.ConfigPath)
		config.HasConfigPath = true
	}

	if visited["server"] {
		config.Server = raw.Server
		config.HasServer = true
	}

	if visited["token"] {
		config.Token = raw.Token
		config.HasToken = true
	}

	if visited["token-file"] {
		config.TokenFile = normalizePath(raw.TokenFile)
		config.HasTokenFile = true
	}

	if visited["api-version"] {
		config.APIVersion = raw.APIVersion
		config.HasAPIVersion = true
	}

	if visited["output"] {
		config.Output = raw.Output
		config.HasOutput = true
	}

	return config, nil
}

func applySourceConfig(settings *ResolvedSettings, source sourceConfig) {
	if source.HasConfigPath {
		settings.ConfigPath = normalizePath(source.ConfigPath)
	}

	if source.HasServer {
		settings.ServerBaseURL = source.Server
	}

	if source.HasToken {
		settings.Token = source.Token
	}

	if source.HasTokenFile {
		settings.TokenFile = normalizePath(source.TokenFile)
	}

	if source.HasDefaultRepoID {
		if source.DefaultRepoID > 0 {
			repoID := source.DefaultRepoID
			settings.DefaultRepoID = &repoID
		} else {
			settings.DefaultRepoID = nil
		}
	}

	if source.HasAPIVersion {
		settings.APIVersion = source.APIVersion
	}

	if source.HasOutput {
		settings.Output = OutputMode(source.Output)
	}
}

func normalizeServerBaseURL(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		value = defaultServerURL
	}

	parsed, err := url.Parse(value)
	if err != nil {
		return "", fmt.Errorf("invalid server URL %q: %w", value, err)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("server URL %q must use http or https", value)
	}

	if parsed.Host == "" {
		return "", fmt.Errorf("server URL %q must include a host", value)
	}

	parsed.RawQuery = ""
	parsed.Fragment = ""
	parsed.Path = strings.TrimRight(parsed.Path, "/")

	return strings.TrimRight(parsed.String(), "/"), nil
}

func normalizeAPIVersion(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return defaultAPIVersion
	}

	return trimmed
}

func normalizeOutputMode(raw string) (OutputMode, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", string(OutputText):
		return OutputText, nil
	case string(OutputJSON):
		return OutputJSON, nil
	default:
		return "", fmt.Errorf("unsupported output mode %q; expected text or json", raw)
	}
}

func splitConfigLine(line string) (string, string, bool) {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return "", "", false
	}

	key := strings.TrimSpace(parts[0])
	if key == "" {
		return "", "", false
	}

	value := trimConfigValue(parts[1])
	return key, value, true
}

func trimConfigValue(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}

	builder := strings.Builder{}
	inSingleQuotes := false
	inDoubleQuotes := false
	for _, r := range value {
		switch r {
		case '\'':
			if !inDoubleQuotes {
				inSingleQuotes = !inSingleQuotes
			}
		case '"':
			if !inSingleQuotes {
				inDoubleQuotes = !inDoubleQuotes
			}
		case '#':
			if !inSingleQuotes && !inDoubleQuotes {
				return stripMatchingQuotes(strings.TrimSpace(builder.String()))
			}
		}

		builder.WriteRune(r)
	}

	return stripMatchingQuotes(strings.TrimSpace(builder.String()))
}

func stripMatchingQuotes(value string) string {
	if len(value) >= 2 {
		if (value[0] == '\'' && value[len(value)-1] == '\'') || (value[0] == '"' && value[len(value)-1] == '"') {
			return value[1 : len(value)-1]
		}
	}

	return value
}

func normalizeConfigKey(key string) string {
	return strings.NewReplacer("-", "_", ".", "_").Replace(strings.ToLower(strings.TrimSpace(key)))
}

func parseOptionalInt64(raw string) (*int64, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, nil
	}

	value, err := strconv.ParseInt(trimmed, 10, 64)
	if err != nil {
		return nil, err
	}

	if value <= 0 {
		return nil, fmt.Errorf("value must be greater than zero")
	}

	return &value, nil
}

func defaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ".gitsemantic/config.yaml"
	}

	return filepath.Join(home, ".gitsemantic", "config.yaml")
}

func defaultTokenFilePath() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ".gitsemantic/token"
	}

	return filepath.Join(home, ".gitsemantic", "token")
}

func normalizePath(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}

	if strings.HasPrefix(trimmed, "~"+string(os.PathSeparator)) || strings.HasPrefix(trimmed, "~/") || strings.HasPrefix(trimmed, "~\\") {
		home, err := os.UserHomeDir()
		if err == nil && home != "" {
			suffix := strings.TrimPrefix(trimmed[1:], "/")
			suffix = strings.TrimPrefix(suffix, "\\")
			return filepath.Clean(filepath.Join(home, suffix))
		}
	}

	return filepath.Clean(trimmed)
}

func resolveRelativePath(baseDirectory string, raw string) string {
	value := normalizePath(raw)
	if value == "" {
		return ""
	}

	if filepath.IsAbs(value) || baseDirectory == "" {
		return value
	}

	return filepath.Clean(filepath.Join(baseDirectory, value))
}
