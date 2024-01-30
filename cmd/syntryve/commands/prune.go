package commands

import (
	"fmt"
	"github.com/KYVENetwork/syntryve/pruner"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	pruneCmd.Flags().StringVar(&until, "until", "", "prune all messages until specified unix timestamp")
	if err := pruneCmd.MarkFlagRequired("until"); err != nil {
		panic(fmt.Errorf("flag 'until' should be required: %w", err))
	}

	pruneCmd.Flags().StringVar(&dbPath, "db-path", ".syntropy/syntropy.db", "path to SQLite DB")

	rootCmd.AddCommand(pruneCmd)
}

var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Prune data until a specific timestamp",
	Run: func(cmd *cobra.Command, args []string) {
		if err := pruner.PruneDB(until, dbPath); err != nil {
			logger.Error().Msg(fmt.Sprintf("failed to prune until %v: %v", until, err))
			os.Exit(1)
		}
		logger.Info().Msg(fmt.Sprintf("successfully pruned until %v", until))
	},
}
