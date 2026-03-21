package cmd

import (
	"fmt"

	"github.com/ludusrusso/fycr/carousel"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init <name>",
	Short: "Initialize a new carousel",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		path, err := carousel.Init(".", name)
		if err != nil {
			return err
		}
		fmt.Printf("Carousel created: %s\n", path)
		return nil
	},
}
