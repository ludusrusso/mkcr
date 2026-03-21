package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "fycr",
	Short: "Farmaceutica Younger Carousel Generator",
	Long:  "A CLI tool to generate LinkedIn carousel PDFs from HTML+Tailwind CSS slides.",
}

func Execute() error {
	return rootCmd.Execute()
}
