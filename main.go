package main

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/ludusrusso/mkcr/cmd"
)

//go:embed llm.md
var llmDoc string

var (
	version = "dev"
	commit  = "dev"
	date    = "dev"
)

func main() {
	cmd.SetLLMDoc(llmDoc)
	cmd.SetVersionInfo(version, commit, date)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
