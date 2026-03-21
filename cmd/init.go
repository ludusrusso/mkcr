package cmd

import (
	"fmt"

	"github.com/ludusrusso/fycr/carousel"
	"github.com/spf13/cobra"
)

var (
	initFormat string
	initWidth  int
	initHeight int
)

func init() {
	initCmd.Flags().StringVar(&initFormat, "format", "square", "slide format: square (1080x1080) or vertical (1080x1350)")
	initCmd.Flags().IntVar(&initWidth, "width", 0, "custom slide width in pixels (overrides --format)")
	initCmd.Flags().IntVar(&initHeight, "height", 0, "custom slide height in pixels (overrides --format)")
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init <name>",
	Short: "Initialize a new carousel",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		width, height, err := resolveSize()
		if err != nil {
			return err
		}

		path, err := carousel.Init(".", name, width, height)
		if err != nil {
			return err
		}
		fmt.Printf("Carousel created: %s (%dx%d)\n", path, width, height)
		return nil
	},
}

func resolveSize() (int, int, error) {
	customW := initWidth != 0
	customH := initHeight != 0

	if customW || customH {
		w := initWidth
		h := initHeight
		if !customW {
			w = carousel.DefaultWidth
		}
		if !customH {
			h = carousel.DefaultHeight
		}
		return w, h, nil
	}

	f, ok := carousel.Formats[initFormat]
	if !ok {
		return 0, 0, fmt.Errorf("unknown format %q (use square or vertical)", initFormat)
	}
	return f.Width, f.Height, nil
}
