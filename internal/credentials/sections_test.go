package credentials

import "testing"

func TestComputeSectionNames_Defaults(t *testing.T) {
	names, err := ComputeSectionNames("default", "", "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if names.LongTerm != "default-long-term" {
		t.Fatalf("long-term: expected %q, got %q", "default-long-term", names.LongTerm)
	}
	if names.ShortTerm != "default" {
		t.Fatalf("short-term: expected %q, got %q", "default", names.ShortTerm)
	}
}

func TestComputeSectionNames_CustomSuffixes(t *testing.T) {
	names, err := ComputeSectionNames("myorg", "none", "production")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if names.LongTerm != "myorg" {
		t.Fatalf("long-term: expected %q, got %q", "myorg", names.LongTerm)
	}
	if names.ShortTerm != "myorg-production" {
		t.Fatalf("short-term: expected %q, got %q", "myorg-production", names.ShortTerm)
	}
}

func TestComputeSectionNames_EqualRejected(t *testing.T) {
	_, err := ComputeSectionNames("prod", "none", "none")
	if err == nil {
		t.Fatalf("expected error when long-term and short-term names match")
	}
}
