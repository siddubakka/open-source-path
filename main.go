package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/siddubakka/open-source-path/internal/fetcher"
	"github.com/siddubakka/open-source-path/internal/report"
)

func main() {
	outDir := flag.String("out", "reports", "directory to write reports into")
	perPage := flag.Int("per-page", 10, "results per query (1-100)")
	timeout := flag.Duration("timeout", 60*time.Second, "total fetch timeout")
	alsoJSON := flag.Bool("json", false, "also write a JSON snapshot")
	flag.Parse()

	opts := fetcher.DefaultOptions()
	opts.PerPage = *perPage
	opts.Timeout = *timeout

	ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
	defer cancel()

	token := os.Getenv("GITHUB_TOKEN")

	repos, err := fetcher.FetchWithOptions(ctx, token, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error fetching repos: %v\n", err)
		os.Exit(1)
	}

	if err := report.GenerateToDir(repos, *outDir, *alsoJSON); err != nil {
		fmt.Fprintf(os.Stderr, "error generating report: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("done: %d repositories written to %s/\n", len(repos), *outDir)
}
