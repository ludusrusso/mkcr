package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/ludusrusso/mkcr/carousel"
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

		addr, cleanup, err := carousel.StartPreviewServer(name)
		if err != nil {
			return err
		}
		defer cleanup()

		fmt.Printf("Preview server running at %s\n", addr)
		fmt.Println("Press Ctrl+C to stop")

		if err := openBrowser(addr); err != nil {
			fmt.Printf("Could not open browser: %v\nOpen %s manually.\n", err, addr)
		}

		// Wait for interrupt signal
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		<-sig

		fmt.Println("\nShutting down...")
		return nil
	},
}

func openBrowser(url string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Run()
	case "linux":
		return exec.Command("xdg-open", url).Run()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Run()
	default:
		return fmt.Errorf("unsupported platform")
	}
}
