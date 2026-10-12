package fetcher

import "testing"

func TestDedupeAndSort(t *testing.T) {
	in := []Repository{
		{Name: "b/repo", Stars: 10},
		{Name: "a/repo", Stars: 50},
		{Name: "b/repo", Stars: 10},
		{Name: "c/repo", Stars: 50},
	}

	got := dedupeAndSort(in)
	if len(got) != 3 {
		t.Fatalf("expected 3 unique repos, got %d", len(got))
	}
	// 50-star repos first, ties broken by name.
	if got[0].Name != "a/repo" || got[1].Name != "c/repo" || got[2].Name != "b/repo" {
		t.Fatalf("unexpected order: %+v", got)
	}
}

func TestParseRetryAfter(t *testing.T) {
	if got := parseRetryAfter(""); got != 0 {
		t.Fatalf("expected 0 for empty header, got %v", got)
	}
	if got := parseRetryAfter("120"); got.Seconds() != 120 {
		t.Fatalf("expected 120s, got %v", got)
	}
}
