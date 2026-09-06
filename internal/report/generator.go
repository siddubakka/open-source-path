package report

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/siddubakka/open-source-path/internal/fetcher"
)

// Generate writes a dated markdown report and updates reports/latest.md.
func Generate(repos []fetcher.Repository) error {
	if err := os.MkdirAll("reports", 0755); err != nil {
		return fmt.Errorf("mkdir reports: %w", err)
	}

	date := time.Now().UTC().Format("2006-01-02")
	content := build(repos, date)

	dated := filepath.Join("reports", date+".md")
	if err := os.WriteFile(dated, []byte(content), 0644); err != nil {
		return fmt.Errorf("write %s: %w", dated, err)
	}

	latest := filepath.Join("reports", "latest.md")
	if err := os.WriteFile(latest, []byte(content), 0644); err != nil {
		return fmt.Errorf("write %s: %w", latest, err)
	}

	return nil
}

func build(repos []fetcher.Repository, date string) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("# cloud-native repository report — %s\n\n", date))
	b.WriteString(fmt.Sprintf("generated: %s UTC\n", time.Now().UTC().Format("2006-01-02 15:04:05")))
	b.WriteString(fmt.Sprintf("repositories: %d\n\n", len(repos)))

	if len(repos) == 0 {
		b.WriteString("no results returned from the GitHub API.\n")
		return b.String()
	}

	b.WriteString("## repositories\n\n")
	b.WriteString("| name | stars | language | description |\n")
	b.WriteString("|------|-------|----------|-------------|\n")

	for _, r := range repos {
		desc := truncate(r.Description, 80)
		lang := r.Language
		if lang == "" {
			lang = "-"
		}
		b.WriteString(fmt.Sprintf("| [%s](%s) | %d | %s | %s |\n",
			r.Name, r.URL, r.Stars, lang, desc))
	}

	b.WriteString("\n## stats\n\n")

	total := 0
	langs := make(map[string]int)
	for _, r := range repos {
		total += r.Stars
		if r.Language != "" {
			langs[r.Language]++
		}
	}

	b.WriteString(fmt.Sprintf("- total stars: %d\n", total))
	b.WriteString(fmt.Sprintf("- unique repos: %d\n", len(repos)))

	b.WriteString("- languages: ")
	first := true
	for lang, count := range langs {
		if !first {
			b.WriteString(", ")
		}
		b.WriteString(fmt.Sprintf("%s (%d)", lang, count))
		first = false
	}
	b.WriteString("\n")

	return b.String()
}

func truncate(s string, max int) string {
	if s == "" {
		return "-"
	}
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
