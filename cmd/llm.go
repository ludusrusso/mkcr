package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var llmDoc string

func SetLLMDoc(doc string) {
	llmDoc = doc
}

func init() {
	rootCmd.AddCommand(llmCmd)
}

var llmCmd = &cobra.Command{
	Use:   "llm",
	Short: "Print LLM-oriented documentation for this tool",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(llmDoc)
	},
}
