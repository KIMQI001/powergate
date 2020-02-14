package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(fpsCmd)
}

var fpsCmd = &cobra.Command{
	Use:   "fps",
	Short: "Provides commands to manage fps",
	Long:  `Provides commands to manage fps`,
}
