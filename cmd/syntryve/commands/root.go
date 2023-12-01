package commands

import (
	"fmt"
	"github.com/spf13/cobra"
)

var (
	accessToken string
	dbPath      string
	natsUrl     string
	streamUrl   string
	port        int64
)

// RootCmd is the root command for syntryve.
var rootCmd = &cobra.Command{
	Use:   "syntryve",
	Short: "Serve Syntropy stream data to validate and archive it with KYVE.",
}

func Execute() {
	serveCmd.Flags().SortFlags = false

	if err := rootCmd.Execute(); err != nil {
		panic(fmt.Errorf("failed to execute root command: %w", err))
	}
}
