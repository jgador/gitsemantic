package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultHTTPTimeout = 30 * time.Second

type APIError struct {
	StatusCode int
	Method     string
	URL        string
	Message    string
}

func (e *APIError) Error() string {
	statusText := http.StatusText(e.StatusCode)
	if statusText == "" {
		statusText = "HTTP error"
	}

	if e.Message == "" {
		return fmt.Sprintf("%s %s failed with %d %s", e.Method, e.URL, e.StatusCode, statusText)
	}

	return fmt.Sprintf("%s %s failed with %d %s: %s", e.Method, e.URL, e.StatusCode, statusText, e.Message)
}

type Client struct {
	httpClient *http.Client
	baseURL    string
	apiVersion string
	token      string
}

func NewClient(settings ResolvedSettings, token string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: defaultHTTPTimeout},
		baseURL:    settings.ServerBaseURL,
		apiVersion: settings.APIVersion,
		token:      strings.TrimSpace(token),
	}
}

func (c *Client) GetHealth(ctx context.Context) (*HealthResponse, error) {
	response, err := doRequest[HealthResponse](ctx, c, http.MethodGet, "/api/health", nil)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *Client) GetVersion(ctx context.Context) (*VersionResponse, error) {
	response, err := doRequest[VersionResponse](ctx, c, http.MethodGet, "/api/version", nil)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *Client) ListRepos(ctx context.Context) ([]RepoSummary, error) {
	response, err := doRequest[[]RepoSummary](ctx, c, http.MethodGet, "/api/repos", nil)
	if err != nil {
		return nil, err
	}

	return *response, nil
}

func (c *Client) Query(ctx context.Context, request QueryRequest) (*QueryResponse, error) {
	response, err := doRequest[QueryResponse](ctx, c, http.MethodPost, "/api/query", request)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *Client) Ingest(ctx context.Context, request IngestRequest) (*IngestAcceptedResponse, error) {
	response, err := doRequest[IngestAcceptedResponse](ctx, c, http.MethodPost, "/api/ingest", request)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *Client) GetIngestionStatus(ctx context.Context, jobID string) (*IngestJobResponse, error) {
	response, err := doRequest[IngestJobResponse](ctx, c, http.MethodGet, "/api/ingest/"+url.PathEscape(jobID), nil)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func doRequest[T any](ctx context.Context, client *Client, method string, endpointPath string, requestBody any) (*T, error) {
	endpointURL, err := buildEndpointURL(client.baseURL, endpointPath, client.apiVersion)
	if err != nil {
		return nil, err
	}

	var bodyReader io.Reader
	if requestBody != nil {
		payload, err := json.Marshal(requestBody)
		if err != nil {
			return nil, fmt.Errorf("failed to encode request for %s %s: %w", method, endpointURL, err)
		}

		bodyReader = bytes.NewReader(payload)
	}

	request, err := http.NewRequestWithContext(ctx, method, endpointURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for %s %s: %w", method, endpointURL, err)
	}

	if requestBody != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	request.Header.Set("Accept", "application/json")
	if client.token != "" {
		request.Header.Set("Authorization", "Bearer "+client.token)
	}

	response, err := client.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("%s %s failed: %w", method, endpointURL, err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response from %s %s: %w", method, endpointURL, err)
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, &APIError{
			StatusCode: response.StatusCode,
			Method:     method,
			URL:        endpointURL,
			Message:    decodeErrorMessage(responseBody),
		}
	}

	var decoded T
	if len(bytes.TrimSpace(responseBody)) == 0 {
		return &decoded, nil
	}

	if err := json.Unmarshal(responseBody, &decoded); err != nil {
		return nil, fmt.Errorf("failed to decode %s %s response: %w", method, endpointURL, err)
	}

	return &decoded, nil
}

func buildEndpointURL(baseURL string, endpointPath string, apiVersion string) (string, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base URL %q: %w", baseURL, err)
	}

	parsed.Path = strings.TrimRight(parsed.Path, "/") + endpointPath
	query := parsed.Query()
	query.Set("api-version", apiVersion)
	parsed.RawQuery = query.Encode()

	return parsed.String(), nil
}

func decodeErrorMessage(body []byte) string {
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return ""
	}

	var payload struct {
		Error            string `json:"error"`
		Detail           string `json:"detail"`
		Title            string `json:"title"`
		ModelAccessError string `json:"modelAccessError"`
	}

	if err := json.Unmarshal(body, &payload); err == nil {
		switch {
		case strings.TrimSpace(payload.Error) != "":
			return strings.TrimSpace(payload.Error)
		case strings.TrimSpace(payload.Title) != "" && strings.TrimSpace(payload.Detail) != "":
			return strings.TrimSpace(payload.Title) + ": " + strings.TrimSpace(payload.Detail)
		case strings.TrimSpace(payload.Detail) != "":
			return strings.TrimSpace(payload.Detail)
		case strings.TrimSpace(payload.Title) != "":
			return strings.TrimSpace(payload.Title)
		case strings.TrimSpace(payload.ModelAccessError) != "":
			return strings.TrimSpace(payload.ModelAccessError)
		}
	}

	return trimmed
}
