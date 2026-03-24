package cmd

import (
	"bytes"
	"testing"
)

func TestVersionCommand_OutputsVersionInfo(t *testing.T) {
	SetVersionInfo("1.2.3", "abc123", "2026-01-01")

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"version"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()

	want := "mkcr version 1.2.3\ncommit: abc123\ndate: 2026-01-01\n"
	if out != want {
		t.Errorf("got:\n%s\nwant:\n%s", out, want)
	}
}

func TestVersionCommand_DefaultValues(t *testing.T) {
	SetVersionInfo("dev", "dev", "dev")

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"version"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()

	want := "mkcr version dev\ncommit: dev\ndate: dev\n"
	if out != want {
		t.Errorf("got:\n%s\nwant:\n%s", out, want)
	}
}
