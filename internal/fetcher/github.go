package fetcher

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

// Repository holds the fields we extract from the GitHub Search API response.
type Repository struct {
	Name        string   `json:"full_name"`
	Description string   `json:"description"`
	Stars       int      `json:"stargazers_count"`
	Forks       int      `json:"forks_count"`
	Language    string   `json:"language"`
	URL         string   `json:"html_url"`
	Topics      []string `json:"topics"`
	UpdatedAt   string   `json:"updated_at"`
}

type searchResponse struct {
	Items []Repository `json:"items"`
}

type query struct {
	label string
	q     string
}

// queries defines the GitHub Search API queries to run each day.
// Modify this slice to change which topics are tracked.
var queries = []query{
	{"kubernetes", "topic:kubernetes stars:>500 pushed:>2024-01-01"},
	{"cncf", "topic:cncf stars:>200"},
	{"go cloud-native", "language:go topic:cloud-native stars:>300"},
	{"service-mesh", "topic:service-mesh stars:>100"},
	{"meshery", "topic:meshery OR meshery in:name stars:>10"},
}

// FetchCloudNativeRepos runs all defined queries against the GitHub Search API
// and returns a deduplicated list of repositories.
func FetchCloudNativeRepos(token string) ([]Repository, error) {
	client := &http.Client{Timeout: 15 * time.Second}

	seen := make(map[string]bool)
	var results []Repository

	for _, q := range queries {
		repos, err := search(client, token, q.q)
		if err != nil {
			fmt.Fprintf(os.Stderr, "skipping query %q: %v\n", q.label, err)
			continue
		}

		for _, r := range repos {
			if !seen[r.Name] {
				seen[r.Name] = true
				results = append(results, r)
			}
		}

		// Avoid hitting GitHub rate limits between requests.
		time.Sleep(1 * time.Second)
	}

	return results, nil
}

func search(client *http.Client, token, q string) ([]Repository, error) {
	endpoint := fmt.Sprintf(
		"https://api.github.com/search/repositories?q=%s&sort=stars&order=desc&per_page=10",
		url.QueryEscape(q),
	)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "open-source-path/1.0")

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var result searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Items, nil
}
