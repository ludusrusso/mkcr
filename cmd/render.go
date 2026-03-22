package cmd

import (
	"fmt"

	"github.com/ludusrusso/mkcr/carousel"
	"github.com/spf13/cobra"
)

var renderOutput string

func init() {
	renderCmd.Flags().StringVar(&renderOutput, "output", "", "Output PDF path (default: <name>/<name>.pdf)")
	rootCmd.AddCommand(renderCmd)
}

var renderCmd = &cobra.Command{
	Use:   "render <name>",
	Short: "Render carousel slides to a single PDF",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		path, err := carousel.RenderPDF(name, renderOutput)
		if err != nil {
			return err
		}
		fmt.Println(path)
		return nil
	},
}
