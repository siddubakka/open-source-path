package report

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/siddubakka/open-source-path/internal/fetcher"
)

func TestTruncateRuneAware(t *testing.T) {
	if got := truncate("", 5); got != "-" {
		t.Fatalf("empty should give -, got %q", got)
	}
	// Multi-byte runes must not be split mid-byte.
	s := strings.Repeat("é", 10)
	got := truncate(s, 5)
	if len([]rune(got)) != 5 {
		t.Fatalf("expected 5 runes, got %q", got)
	}
}

func TestEscapeCell(t *testing.T) {
	got := escapeCell("a|b\nc")
	if strings.Contains(got, "|") && !strings.Contains(got, "\\|") {
		t.Fatalf("pipe not escaped: %q", got)
	}
	if strings.Contains(got, "\n") {
		t.Fatalf("newline not removed: %q", got)
	}
}

func TestBuildSortsByStars(t *testing.T) {
	repos := []fetcher.Repository{
		{Name: "b/b", URL: "https://example.com/b", Stars: 1},
		{Name: "a/a", URL: "https://example.com/a", Stars: 99},
	}
	out := build(repos, "2026-01-01")
	ia := strings.Index(out, "a/a")
	ib := strings.Index(out, "b/b")
	if ia < 0 || ib < 0 || ia > ib {
		t.Fatalf("expected a/a before b/b:\n%s", out)
	}
}

func TestGenerateToDirWritesFiles(t *testing.T) {
	dir := t.TempDir()
	repos := []fetcher.Repository{
		{Name: "a/a", URL: "https://example.com/a", Stars: 5, Language: "Go"},
	}
	if err := GenerateToDir(repos, dir, true); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"latest.md"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("missing %s: %v", name, err)
		}
	}
	matches, _ := filepath.Glob(filepath.Join(dir, "*.json"))
	if len(matches) != 1 {
		t.Fatalf("expected 1 json snapshot, got %v", matches)
	}
}
