package commands

import (
	"fmt"
	"github.com/KYVENetwork/syntryve/pruner"
	"github.com/KYVENetwork/syntryve/syntropy"
	"github.com/KYVENetwork/syntryve/utils"
	"github.com/spf13/cobra"
	"os"
	"slices"
)

func init() {
	serveCmd.Flags().StringVar(&accessToken, "token", "", "Syntropy access token")
	if err := serveCmd.MarkFlagRequired("token"); err != nil {
		panic(fmt.Errorf("flag 'token' should be required: %w", err))
	}

	serveCmd.Flags().StringVar(&natsUrl, "nats-url", "", "url of Syntropy nats")
	if err := serveCmd.MarkFlagRequired("nats-url"); err != nil {
		panic(fmt.Errorf("flag 'nats-url' should be required: %w", err))
	}

	serveCmd.Flags().StringVar(&streamUrl, "stream-url", "", "url of Syntropy stream")
	if err := serveCmd.MarkFlagRequired("stream-url"); err != nil {
		panic(fmt.Errorf("flag 'stream-url' should be required: %w", err))
	}

	serveCmd.Flags().StringVar(&consumerId, "consumer-id", "", "url of Syntropy stream")
	if err := serveCmd.MarkFlagRequired("consumer-id"); err != nil {
		panic(fmt.Errorf("flag 'consumer-id' should be required: %w", err))
	}

	serveCmd.Flags().StringVar(&dbPath, "db-path", ".syntropy/syntropy.db", "path to SQLite DB")

	serveCmd.Flags().Int64Var(&port, "port", 4242, "server port")

	serveCmd.Flags().Int64Var(&pruningInterval, "pruning-interval", 30, "pruning interval in minutes")

	serveCmd.Flags().StringVar(&chainId, "chain-id", "kyve-1", "KYVE chain-id (required for pruning)")

	serveCmd.Flags().Int64Var(&poolId, "pool-id", 0, "KYVE pool id (required for pruning)")

	serveCmd.Flags().StringVar(&poolEndpoints, "endpoints", "", "overwrite endpoints to query latest KYVE pool key")

	serveCmd.Flags().BoolVar(&debug, "debug", false, "debug mode")

	rootCmd.AddCommand(serveCmd)
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve data items of Syntropy stream for validating and archiving in a KYVE pool",
	Run: func(cmd *cobra.Command, args []string) {
		if pruningInterval > 0 {
			supportedChains := []string{"kyve-1", "kaon-1", "korellia-2"}
			if !slices.Contains(supportedChains, chainId) {
				logger.Error().Str("chain-id", chainId).Msg("specified chain-id %v is not supported")
				os.Exit(1)
			}
			if pruningInterval < 10 {
				logger.Error().Msg("pruning interval needs to be bigger than 10 minutes")
				os.Exit(1)
			}
		}

		if err := utils.EnsureDBPathExists(dbPath); err != nil {
			panic(err)
		}

		go syntropy.StartSyntropyWS(accessToken, natsUrl, streamUrl, consumerId, dbPath, debug)

		if pruningInterval > 0 {
			go pruner.StartPruningScheduler(pruningInterval, poolId, dbPath, chainId, poolEndpoints, debug)
		}

		syntropy.StartApiServer(dbPath, debug, port)
	},
}
