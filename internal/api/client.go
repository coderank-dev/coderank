// Package api provides an HTTP client for the CodeRank API. All CLI commands
// that fetch data from the API use this client. It handles authentication
// (API key from credentials file), request construction, and response parsing.
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	// DefaultBaseURL is the production API endpoint.
	DefaultBaseURL = "https://api.coderank.ai"

	// requestTimeout is the maximum duration for any API call.
	// The slowest operation is query (embedding + search + fetch ≈ 100ms),
	// so 10 seconds is generous and handles network variability.
	requestTimeout = 10 * time.Second
)

// Client is an HTTP client for the CodeRank API. Create via NewClient.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// QueryRequest is the POST body for /v1/query.
type QueryRequest struct {
	Q         string       `json:"q"`
	MaxTokens int          `json:"max_tokens,omitempty"`
	Library   string       `json:"library,omitempty"`
	Config    *QueryConfig `json:"config,omitempty"`
}

// QueryConfig represents the .coderank.yml settings sent with queries.
type QueryConfig struct {
	Preferred        []string `json:"preferred,omitempty"`
	Blocked          []string `json:"blocked,omitempty"`
	PreferTypeScript bool     `json:"prefer_typescript,omitempty"`
}

// QueryResponse is the response from POST /v1/query.
type QueryResponse struct {
	Results     []DocResult `json:"results"`
	TotalTokens int         `json:"total_tokens"`
	QueryMs     int         `json:"query_ms"`
}

// DocResult is a single document in a query response.
type DocResult struct {
	Library string `json:"library"`
	Version string `json:"version"`
	Topic   string `json:"topic"`
	Type    string `json:"type"`
	Tokens  int    `json:"tokens"`
	Score   int    `json:"score"`
	Content string `json:"content"`
}

// TopicResponse is the response from GET /v1/topic/:library/:topic.
type TopicResponse struct {
	Library string `json:"library"`
	Version string `json:"version"`
	Topic   string `json:"topic"`
	Tokens  int    `json:"tokens"`
	Content string `json:"content"`
}

// TopicsResponse is the response from GET /v1/topics/:library.
type TopicsResponse struct {
	Library string   `json:"library"`
	Version string   `json:"version"`
	Topics  []string `json:"topics"`
}

// HealthResponse is the response from GET /v1/health/:library.
type HealthResponse struct {
	Library     string         `json:"library"`
	Repo        string         `json:"repo"`
	HealthScore int            `json:"health_score"`
	Breakdown   map[string]int `json:"breakdown"`
	LastIndexed string         `json:"last_indexed"`
}

// CompareResponse is the response from GET /v1/compare.
type CompareResponse struct {
	Category  string           `json:"category"`
	Libraries []HealthResponse `json:"libraries"`
}

// NewClient creates an API client. It reads the API key from
// ~/.coderank/credentials. If baseURL is empty, uses the production API.
// Returns an error if no credentials are found (user needs to run coderank auth).
func NewClient(baseURL string) (*Client, error) {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	apiKey, err := readAPIKey()
	if err != nil {
		return nil, fmt.Errorf("not authenticated: run 'coderank auth <api-key>' first: %w", err)
	}

	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: requestTimeout,
		},
	}, nil
}

// Query calls POST /v1/query and returns condensed documentation.
func (c *Client) Query(req QueryRequest) (*QueryResponse, error) {
	if req.Library != "" {
		req.Library = NormalizeLibraryName(req.Library)
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling query request: %w", err)
	}

	respBody, err := c.doRequest("POST", "/v1/query", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	var result QueryResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parsing query response: %w", err)
	}
	return &result, nil
}

// NormalizeLibraryName applies client-side normalization before sending to the API:
// lowercases and strips common structural suffixes (.js, .ts).
// This covers case variants and ".js"-suffixed names (e.g. "React.js" → "react").
// Semantic aliases (e.g. "reactjs" → "react") are resolved server-side via the
// library_aliases table — they can't be handled safely without a lookup.
func NormalizeLibraryName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ToLower(name)
	for _, suffix := range []string{".js", ".ts"} {
		name = strings.TrimSuffix(name, suffix)
	}
	return name
}

// Topic calls GET /v1/topic/:library/:topic and returns the full topic content.
func (c *Client) Topic(library, topic string) (*TopicResponse, error) {
	library = NormalizeLibraryName(library)
	path := "/v1/topic/" + url.PathEscape(library) + "/" + url.PathEscape(topic)
	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	var result TopicResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parsing topic response: %w", err)
	}
	return &result, nil
}

// Topics calls GET /v1/topics/:library and returns the list of available topics.
func (c *Client) Topics(library string) (*TopicsResponse, error) {
	library = NormalizeLibraryName(library)
	respBody, err := c.doRequest("GET", "/v1/topics/"+url.PathEscape(library), nil)
	if err != nil {
		return nil, err
	}
	var result TopicsResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parsing topics response: %w", err)
	}
	return &result, nil
}

// Surface calls GET /v1/surface/:library and returns the API surface file.
func (c *Client) Surface(library string) (*DocResult, error) {
	library = NormalizeLibraryName(library)
	respBody, err := c.doRequest("GET", "/v1/surface/"+url.PathEscape(library), nil)
	if err != nil {
		return nil, err
	}

	var result DocResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parsing surface response: %w", err)
	}
	return &result, nil
}

// Health calls GET /v1/health/:library and returns the health score.
func (c *Client) Health(library string) (*HealthResponse, error) {
	respBody, err := c.doRequest("GET", "/v1/health/"+url.PathEscape(library), nil)
	if err != nil {
		return nil, err
	}

	var result HealthResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parsing health response: %w", err)
	}
	return &result, nil
}

// Compare calls GET /v1/compare and returns ranked libraries in a category.
func (c *Client) Compare(category string, limit int) (*CompareResponse, error) {
	path := fmt.Sprintf("/v1/compare?category=%s&limit=%d",
		url.QueryEscape(category), limit)

	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result CompareResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parsing compare response: %w", err)
	}
	return &result, nil
}

// doRequest executes an HTTP request with authentication and error handling.
func (c *Client) doRequest(method, path string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, c.baseURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode == 429 {
		return nil, fmt.Errorf("rate limit exceeded — upgrade at https://coderank.ai/pricing")
	}

	if resp.StatusCode >= 400 {
		var apiErr struct {
			Error string `json:"error"`
		}
		json.Unmarshal(respBody, &apiErr)
		if apiErr.Error != "" {
			return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, apiErr.Error)
		}
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// readAPIKey reads the API key from ~/.coderank/credentials.
// The file format is a single line: the raw API key.
func readAPIKey() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(filepath.Join(home, ".coderank", "credentials"))
	if err != nil {
		return "", err
	}

	key := strings.TrimSpace(string(data))
	if key == "" {
		return "", fmt.Errorf("credentials file is empty")
	}
	return key, nil
}
