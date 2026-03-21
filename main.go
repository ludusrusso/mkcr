package main

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/ludusrusso/fycr/cmd"
)

//go:embed llm.md
var llmDoc string

func main() {
	cmd.SetLLMDoc(llmDoc)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
