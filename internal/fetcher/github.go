package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"sync"
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

// Options tunes fetching behaviour.
type Options struct {
	// PerPage is the number of results requested per query (max 100).
	PerPage int
	// Timeout is the total budget for all queries.
	Timeout time.Duration
	// MaxConcurrency bounds how many queries run in parallel.
	MaxConcurrency int
}

// DefaultOptions returns sane defaults.
func DefaultOptions() Options {
	return Options{
		PerPage:        10,
		MaxConcurrency: 3,
		Timeout:        60 * time.Second,
	}
}

// FetchCloudNativeRepos runs all defined queries against the GitHub Search API
// and returns a deduplicated list of repositories sorted by stars.
func FetchCloudNativeRepos(token string) ([]Repository, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultOptions().Timeout)
	defer cancel()
	return FetchWithOptions(ctx, token, DefaultOptions())
}

// FetchWithOptions is FetchCloudNativeRepos with explicit context and options.
func FetchWithOptions(ctx context.Context, token string, opts Options) ([]Repository, error) {
	if opts.PerPage <= 0 || opts.PerPage > 100 {
		opts.PerPage = 10
	}
	if opts.MaxConcurrency <= 0 {
		opts.MaxConcurrency = 3
	}

	client := &http.Client{Timeout: 15 * time.Second}

	type result struct {
		label string
		repos []Repository
		err   error
	}

	sem := make(chan struct{}, opts.MaxConcurrency)
	results := make([]result, len(queries))
	var wg sync.WaitGroup

	for i, q := range queries {
		wg.Add(1)
		go func(i int, q query) {
			defer wg.Done()

			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				results[i] = result{label: q.label, err: ctx.Err()}
				return
			}

			repos, err := search(ctx, client, token, q.q, opts.PerPage)
			results[i] = result{label: q.label, repos: repos, err: err}
		}(i, q)
	}
	wg.Wait()

	var errs []string
	all := make([]Repository, 0, len(queries)*opts.PerPage)
	for _, r := range results {
		if r.err != nil {
			// Rate-limit errors abort everything; other query errors are skipped.
			if isRateLimit(r.err) {
				return nil, r.err
			}
			fmt.Fprintf(os.Stderr, "skipping query %q: %v\n", r.label, r.err)
			continue
		}
		all = append(all, r.repos...)
	}

	if len(all) == 0 && len(errs) > 0 {
		return nil, fmt.Errorf("all queries failed")
	}

	return dedupeAndSort(all), nil
}

// rateLimitError signals GitHub rate limiting so callers can back off.
type rateLimitError struct {
	retryAfter time.Duration
	msg        string
}

func (e *rateLimitError) Error() string { return e.msg }

func isRateLimit(err error) bool {
	_, ok := err.(*rateLimitError)
	return ok
}

func search(ctx context.Context, client *http.Client, token, q string, perPage int) ([]Repository, error) {
	endpoint := fmt.Sprintf(
		"https://api.github.com/search/repositories?q=%s&sort=stars&order=desc&per_page=%d",
		url.QueryEscape(q),
		perPage,
	)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
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

	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == 429 {
		retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))
		if retryAfter == 0 {
			retryAfter = time.Minute
		}
		return nil, &rateLimitError{
			retryAfter: retryAfter,
			msg:        fmt.Sprintf("github rate limited (status %d), retry after %s", resp.StatusCode, retryAfter),
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var result searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Items, nil
}

func parseRetryAfter(h string) time.Duration {
	if h == "" {
		return 0
	}
	if secs, err := strconv.Atoi(h); err == nil {
		return time.Duration(secs) * time.Second
	}
	return 0
}

// dedupeAndSort removes duplicate repos by name and sorts by stars descending.
func dedupeAndSort(repos []Repository) []Repository {
	seen := make(map[string]bool, len(repos))
	out := make([]Repository, 0, len(repos))
	for _, r := range repos {
		if !seen[r.Name] {
			seen[r.Name] = true
			out = append(out, r)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Stars == out[j].Stars {
			return out[i].Name < out[j].Name
		}
		return out[i].Stars > out[j].Stars
	})
	return out
}
