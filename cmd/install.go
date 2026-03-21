package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

var installDir string

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install fycr binary to a directory in your PATH",
	RunE: func(cmd *cobra.Command, args []string) error {
		src, err := os.Executable()
		if err != nil {
			return fmt.Errorf("cannot determine current executable path: %w", err)
		}

		// Resolve symlinks to get the real path
		src, err = filepath.EvalSymlinks(src)
		if err != nil {
			return fmt.Errorf("cannot resolve executable path: %w", err)
		}

		dest := filepath.Join(installDir, "fycr")

		// On Windows, append .exe
		if runtime.GOOS == "windows" {
			dest += ".exe"
		}

		fmt.Printf("Installing fycr to %s\n", dest)

		// Use cp to copy the binary preserving permissions
		cpCmd := exec.Command("cp", src, dest)
		cpCmd.Stdout = os.Stdout
		cpCmd.Stderr = os.Stderr
		if err := cpCmd.Run(); err != nil {
			return fmt.Errorf("failed to install binary (you may need sudo): %w", err)
		}

		// Ensure the binary is executable
		if err := os.Chmod(dest, 0755); err != nil {
			return fmt.Errorf("failed to set executable permissions: %w", err)
		}

		fmt.Println("fycr installed successfully!")
		return nil
	},
}

func init() {
	installCmd.Flags().StringVar(&installDir, "dir", "/usr/local/bin", "destination directory")
	rootCmd.AddCommand(installCmd)
}
