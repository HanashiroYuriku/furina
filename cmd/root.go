package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "ayaka",
	Short: "Ayaka Backend Service",
	Long: "Ayaka is a backend service for managing various functionalities.",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(serverCmd)
}