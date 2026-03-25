package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeLibraryName(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"react", "react"},
		{"React", "react"},
		{"React.js", "react"},
		{"react.js", "react"},
		{"Vue.ts", "vue"},
		{"  NextJS  ", "nextjs"},    // spaces trimmed, lowercase — alias resolution is server-side
		{"tailwindcss", "tailwindcss"}, // canonical name unchanged
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			assert.Equal(t, tc.want, NormalizeLibraryName(tc.input),
				"NormalizeLibraryName(%q) should return canonical form", tc.input)
		})
	}
}

func TestNewClientFailsWithoutCredentials(t *testing.T) {
	// Temporarily point HOME to an empty dir so no credentials file exists
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", t.TempDir())
	defer os.Setenv("HOME", origHome)

	_, err := NewClient("")
	assert.ErrorContains(t, err, "not authenticated",
		"should tell the user to run coderank auth when no credentials exist")
}

func TestClientHandlesAPIErrors(t *testing.T) {
	// Spin up a test server that returns a 401 error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		w.Write([]byte(`{"error":"Invalid API key"}`))
	}))
	defer server.Close()

	home := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", home)
	defer os.Setenv("HOME", origHome)
	require.NoError(t, os.MkdirAll(filepath.Join(home, ".coderank"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(home, ".coderank", "credentials"), []byte("cr_sk_test"), 0600))

	client, err := NewClient(server.URL)
	require.NoError(t, err)

	_, err = client.Query(QueryRequest{Q: "test"})
	assert.ErrorContains(t, err, "Invalid API key",
		"should propagate the API error message to the user")
}

func TestClientParsesTopicResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/topic/react/hooks", r.URL.Path)

		w.Write([]byte(`{"library":"react","version":"19.1.0","topic":"hooks","tokens":1200,"content":"# Hooks\n\nuseState docs."}`))
	}))
	defer server.Close()

	home := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", home)
	defer os.Setenv("HOME", origHome)
	require.NoError(t, os.MkdirAll(filepath.Join(home, ".coderank"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(home, ".coderank", "credentials"), []byte("cr_sk_test"), 0600))

	client, err := NewClient(server.URL)
	require.NoError(t, err)

	resp, err := client.Topic("react", "hooks")
	require.NoError(t, err)
	assert.Equal(t, "react", resp.Library)
	assert.Equal(t, "hooks", resp.Topic)
	assert.Equal(t, 1200, resp.Tokens)
}

func TestClientParsesTopicsResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/topics/react", r.URL.Path)

		w.Write([]byte(`{"library":"react","version":"19.1.0","topics":["hooks","components","routing"]}`))
	}))
	defer server.Close()

	home := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", home)
	defer os.Setenv("HOME", origHome)
	require.NoError(t, os.MkdirAll(filepath.Join(home, ".coderank"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(home, ".coderank", "credentials"), []byte("cr_sk_test"), 0600))

	client, err := NewClient(server.URL)
	require.NoError(t, err)

	resp, err := client.Topics("react")
	require.NoError(t, err)
	assert.Equal(t, []string{"hooks", "components", "routing"}, resp.Topics)
}

func TestClientParsesQueryResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/query", r.URL.Path)
		assert.Contains(t, r.Header.Get("Authorization"), "Bearer cr_sk_test")

		w.Write([]byte(`{
			"results": [{"library":"react","version":"19.1.0","topic":"hooks","tokens":1500,"content":"# Hooks"}],
			"total_tokens": 1500,
			"query_ms": 65
		}`))
	}))
	defer server.Close()

	home := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", home)
	defer os.Setenv("HOME", origHome)
	require.NoError(t, os.MkdirAll(filepath.Join(home, ".coderank"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(home, ".coderank", "credentials"), []byte("cr_sk_test"), 0600))

	client, err := NewClient(server.URL)
	require.NoError(t, err)

	resp, err := client.Query(QueryRequest{Q: "react hooks", MaxTokens: 5000})
	require.NoError(t, err)
	assert.Equal(t, 1, len(resp.Results))
	assert.Equal(t, "react", resp.Results[0].Library)
	assert.Equal(t, 1500, resp.TotalTokens)
}
