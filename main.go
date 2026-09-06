package main

import (
	"fmt"
	"os"

	"github.com/siddubakka/open-source-path/internal/fetcher"
	"github.com/siddubakka/open-source-path/internal/report"
)

func main() {
	token := os.Getenv("GITHUB_TOKEN")

	repos, err := fetcher.FetchCloudNativeRepos(token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error fetching repos: %v\n", err)
		os.Exit(1)
	}

	if err := report.Generate(repos); err != nil {
		fmt.Fprintf(os.Stderr, "error generating report: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("done: %d repositories written to reports/\n", len(repos))
}
