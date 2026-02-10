package main

import (
	"bytes"
	"testing"
)

func TestVersionFlagPrintsDevByDefault(t *testing.T) {
	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if got := out.String(); got != "dev\n" {
		t.Fatalf("expected %q, got %q", "dev\n", got)
	}
}
