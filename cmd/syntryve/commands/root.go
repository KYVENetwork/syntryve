package commands

import (
	"fmt"
	"github.com/KYVENetwork/syntryve/utils"
	"github.com/spf13/cobra"
)

var (
	accessToken     string
	chainId         string
	consumerId      string
	dbPath          string
	debug           bool
	natsUrl         string
	port            int64
	poolId          int64
	poolEndpoints   string
	pruningInterval int64
	streamUrl       string
	until           string
)

var (
	logger = utils.SyntryveLogger("commands")
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
