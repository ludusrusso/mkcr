package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/ludusrusso/fycr/carousel"
	"github.com/spf13/cobra"
)

var (
	addHTML     string
	addFile    string
	addPosition int
)

func init() {
	addCmd.Flags().StringVar(&addHTML, "html", "", "HTML content for the slide")
	addCmd.Flags().StringVar(&addFile, "file", "", "Path to file containing HTML content")
	addCmd.Flags().IntVar(&addPosition, "position", 0, "Slide position (0 = auto-increment)")
	rootCmd.AddCommand(addCmd)
}

var addCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add a slide to a carousel",
	Long: `Add a slide to a carousel. Content can be provided via:
  --html "<div>...</div>"     HTML string
  --file path/to/content.html File containing HTML
  stdin                       Pipe HTML content`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// Determine content source
		var content string
		switch {
		case addHTML != "":
			content = addHTML
		case addFile != "":
			data, err := os.ReadFile(addFile)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}
			content = string(data)
		default:
			// Try stdin
			stat, _ := os.Stdin.Stat()
			if (stat.Mode() & os.ModeCharDevice) == 0 {
				data, err := io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("failed to read stdin: %w", err)
				}
				content = string(data)
			} else {
				return fmt.Errorf("no content provided. Use --html, --file, or pipe via stdin")
			}
		}

		// Determine position
		position := addPosition
		if position == 0 {
			next, err := carousel.NextSlideNumber(name)
			if err != nil {
				return err
			}
			position = next
		}

		path, err := carousel.AddSlide(name, content, position)
		if err != nil {
			return err
		}

		fmt.Printf("Slide %d: %s\n", position, path)
		return nil
	},
}
