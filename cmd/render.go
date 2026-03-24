package cmd

import (
	"fmt"

	"github.com/ludusrusso/mkcr/carousel"
	"github.com/spf13/cobra"
)

var (
	renderOutput string
	renderFormat string
)

func init() {
	renderCmd.Flags().StringVar(&renderOutput, "output", "", "Output path (file for pdf, directory for png)")
	renderCmd.Flags().StringVar(&renderFormat, "format", "pdf", "Output format: pdf or png")
	rootCmd.AddCommand(renderCmd)
}

var renderCmd = &cobra.Command{
	Use:   "render <name>",
	Short: "Render carousel slides to PDF or PNG",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		switch renderFormat {
		case "pdf":
			path, err := carousel.RenderPDF(name, renderOutput)
			if err != nil {
				return err
			}
			fmt.Println(path)
		case "png":
			path, err := carousel.RenderPNG(name, renderOutput)
			if err != nil {
				return err
			}
			fmt.Println(path)
		default:
			return fmt.Errorf("invalid format %q: must be pdf or png", renderFormat)
		}

		return nil
	},
}
