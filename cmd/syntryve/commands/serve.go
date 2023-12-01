package commands

import (
	"fmt"
	"github.com/KYVENetwork/syntropy-demo/server"
	"github.com/KYVENetwork/syntropy-demo/syntropy"
	"github.com/spf13/cobra"
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

	serveCmd.Flags().StringVar(&dbPath, "db-path", ".syntropy/syntropy.db", "path to SQLite DB")

	serveCmd.Flags().Int64Var(&port, "port", 4242, "server port")

	rootCmd.AddCommand(serveCmd)
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve data items of Syntropy stream for validating and archiving in a KYVE pool",
	Run: func(cmd *cobra.Command, args []string) {
		go syntropy.StartSyntropyWS(accessToken, natsUrl, streamUrl, dbPath)
		server.StartApiServer(dbPath, port)
	},
}
