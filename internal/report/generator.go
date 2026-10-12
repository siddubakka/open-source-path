package report

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/siddubakka/open-source-path/internal/fetcher"
)

// Generate writes a dated markdown report and updates reports/latest.md.
func Generate(repos []fetcher.Repository) error {
	return GenerateToDir(repos, "reports", false)
}

// GenerateToDir writes a dated markdown report into dir and updates latest.md.
// When alsoJSON is true it additionally writes a dated .json snapshot.
func GenerateToDir(repos []fetcher.Repository, dir string, alsoJSON bool) error {
	if dir == "" {
		dir = "reports"
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}

	date := time.Now().UTC().Format("2006-01-02")
	content := build(sortedCopy(repos), date)

	dated := filepath.Join(dir, date+".md")
	if err := os.WriteFile(dated, []byte(content), 0644); err != nil {
		return fmt.Errorf("write %s: %w", dated, err)
	}

	latest := filepath.Join(dir, "latest.md")
	if err := os.WriteFile(latest, []byte(content), 0644); err != nil {
		return fmt.Errorf("write %s: %w", latest, err)
	}

	if alsoJSON {
		payload, err := json.MarshalIndent(sortedCopy(repos), "", "  ")
		if err != nil {
			return fmt.Errorf("marshal json: %w", err)
		}
		jpath := filepath.Join(dir, date+".json")
		if err := os.WriteFile(jpath, payload, 0644); err != nil {
			return fmt.Errorf("write %s: %w", jpath, err)
		}
	}

	return nil
}

func sortedCopy(repos []fetcher.Repository) []fetcher.Repository {
	out := make([]fetcher.Repository, len(repos))
	copy(out, repos)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Stars == out[j].Stars {
			return out[i].Name < out[j].Name
		}
		return out[i].Stars > out[j].Stars
	})
	return out
}

func build(repos []fetcher.Repository, date string) string {
	repos = sortedCopy(repos)
	var b strings.Builder

	b.WriteString(fmt.Sprintf("# cloud-native repository report — %s\n\n", date))
	b.WriteString(fmt.Sprintf("generated: %s UTC\n", time.Now().UTC().Format("2006-01-02 15:04:05")))
	b.WriteString(fmt.Sprintf("repositories: %d\n\n", len(repos)))

	if len(repos) == 0 {
		b.WriteString("no results returned from the GitHub API.\n")
		return b.String()
	}

	b.WriteString("## repositories\n\n")
	b.WriteString("| name | stars | forks | language | description |\n")
	b.WriteString("|------|-------|-------|----------|-------------|\n")

	for _, r := range repos {
		desc := truncate(r.Description, 80)
		lang := r.Language
		if lang == "" {
			lang = "-"
		}
		b.WriteString(fmt.Sprintf("| [%s](%s) | %d | %d | %s | %s |\n",
			escapeCell(r.Name), r.URL, r.Stars, r.Forks, escapeCell(lang), escapeCell(desc)))
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

	names := make([]string, 0, len(langs))
	for lang := range langs {
		names = append(names, lang)
	}
	sort.Strings(names)

	b.WriteString("- languages: ")
	if len(names) == 0 {
		b.WriteString("-\n")
		return b.String()
	}
	for i, lang := range names {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(fmt.Sprintf("%s (%d)", lang, langs[lang]))
	}
	b.WriteString("\n")

	return b.String()
}

// escapeCell neutralises markdown table separators and newlines.
func escapeCell(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return s
}

func truncate(s string, max int) string {
	if s == "" {
		return "-"
	}
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-3]) + "..."
}
