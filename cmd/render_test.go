package cmd

import (
	"strings"
	"testing"
)

func TestRenderCommand_InvalidFormat(t *testing.T) {
	rootCmd.SetArgs([]string{"render", "testdata", "--format", "xyz"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid format")
	}
	if !strings.Contains(err.Error(), `invalid format "xyz"`) {
		t.Errorf("error = %q, want it to mention invalid format", err.Error())
	}
}
