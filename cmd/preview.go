package cmd

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/ludusrusso/fycr/carousel"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(previewCmd)
}

var previewCmd = &cobra.Command{
	Use:   "preview <name>",
	Short: "Open carousel slides in the browser",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		slides, err := carousel.ListSlides(name)
		if err != nil {
			return fmt.Errorf("failed to list slides: %w", err)
		}
		if len(slides) == 0 {
			return fmt.Errorf("no slides found in %s", name)
		}

		// Open the first slide in the default browser
		firstSlide := carousel.SlidePath(name, slides[0])
		absPath, _ := filepath.Abs(firstSlide)
		fileURL := "file://" + absPath

		fmt.Printf("Opening %d slide(s) in browser...\n", len(slides))

		// Open each slide
		for _, slideNum := range slides {
			slidePath := carousel.SlidePath(name, slideNum)
			absP, _ := filepath.Abs(slidePath)
			url := "file://" + absP
			if err := openBrowser(url); err != nil {
				fmt.Printf("Failed to open slide %d: %v\n", slideNum, err)
			}
		}

		_ = fileURL // used above in loop
		return nil
	},
}

func openBrowser(url string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Start()
	case "linux":
		return exec.Command("xdg-open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	default:
		return fmt.Errorf("unsupported platform")
	}
}
