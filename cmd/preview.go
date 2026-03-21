package cmd

import (
	"fmt"
	"os/exec"
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

		addr, err := carousel.StartPreviewServer(name)
		if err != nil {
			return err
		}

		fmt.Printf("Preview server running at %s\n", addr)
		fmt.Println("Press Ctrl+C to stop")

		if err := openBrowser(addr); err != nil {
			fmt.Printf("Could not open browser: %v\nOpen %s manually.\n", err, addr)
		}

		// Block forever until Ctrl+C
		select {}
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
