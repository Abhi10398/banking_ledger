package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCommand = &cobra.Command{
	Use:   "awesome-service",
	Short: "A CLI tool to start awesome API server",
}

func init() {
	rootCommand.AddCommand(apiCommand)
}

func Execute() {
	err := rootCommand.Execute()
	if err != nil {
		os.Exit(1)
	}
}
