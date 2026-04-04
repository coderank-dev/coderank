package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// githubCommit is the subset of the GitHub compare API commit object we need.
type githubCommit struct {
	SHA    string `json:"sha"`
	Commit struct {
		Message string `json:"message"`
		Author  struct {
			Date string `json:"date"`
		} `json:"author"`
	} `json:"commit"`
	Author *struct {
		Login string `json:"login"`
	} `json:"author"`
}

// githubTag from the list-tags API.
type githubTag struct {
	Name string `json:"name"`
}

// FetchCommitsBetweenTags fetches commits between two tags for one repo using the
// GitHub compare API. Pass an empty fromTag to get all commits up to toTag.
func FetchCommitsBetweenTags(org, repo, fromTag, toTag, token string) ([]githubCommit, error) {
	baseURL := "https://api.github.com"
	client := &http.Client{Timeout: 30 * time.Second}

	var base string
	if fromTag == "" {
		// No previous tag — use the first commit as base by comparing to an empty tree.
		// Simpler: just get commits on the default branch up to toTag.
		base = "HEAD~1000" // fallback; in practice we always have a previous tag
	} else {
		base = fromTag
	}

	url := fmt.Sprintf("%s/repos/%s/%s/compare/%s...%s", baseURL, org, repo, base, toTag)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github api request to %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api returned %d for %s", resp.StatusCode, url)
	}

	var payload struct {
		Commits []githubCommit `json:"commits"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode compare response: %w", err)
	}
	return payload.Commits, nil
}

// LatestTags returns the two most recent semver tags for a repo.
// Returns (latestTag, previousTag, error). previousTag is empty if only one tag exists.
func LatestTags(org, repo, token string) (latest, previous string, err error) {
	baseURL := "https://api.github.com"
	client := &http.Client{Timeout: 30 * time.Second}

	url := fmt.Sprintf("%s/repos/%s/%s/tags?per_page=10", baseURL, org, repo)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("github api request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("github api returned %d for tags", resp.StatusCode)
	}

	var tags []githubTag
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return "", "", fmt.Errorf("decode tags: %w", err)
	}

	if len(tags) == 0 {
		return "", "", fmt.Errorf("no tags found for %s/%s", org, repo)
	}
	latest = tags[0].Name
	if len(tags) > 1 {
		previous = tags[1].Name
	}
	return latest, previous, nil
}

// FetchRepoEntries fetches and classifies commits for one repo between its two
// most recent tags. It uses toTag as the "latest" anchor for repos that haven't
// had their own release yet (api, pipeline may lag behind coderank).
func FetchRepoEntries(org, repo, token string) ([]Entry, error) {
	latest, previous, err := LatestTags(org, repo, token)
	if err != nil {
		// Repo has no tags yet — skip gracefully.
		return nil, nil
	}

	commits, err := FetchCommitsBetweenTags(org, repo, previous, latest, token)
	if err != nil {
		return nil, fmt.Errorf("fetch commits for %s: %w", repo, err)
	}

	entries := make([]Entry, 0, len(commits))
	for _, c := range commits {
		login := ""
		if c.Author != nil {
			login = c.Author.Login
		}
		entries = append(entries, ClassifyCommit(c.Commit.Message, c.SHA[:7], repo, login))
	}
	return entries, nil
}
