package syntropy

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/KYVENetwork/syntropy-demo/types"
	"github.com/KYVENetwork/syntropy-demo/utils"
	"github.com/SyntropyNet/pubsub-go/pubsub"
	"github.com/goccy/go-json"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nats-io/nats.go"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func StartSyntropyWS(accessToken, natsUrl, streamUrl, dbPath string, debug bool) {
	jwt, _ := pubsub.CreateAppJwt(accessToken)
	if debug {
		println(jwt)
	}
	// Set user credentials and options for NATS connection.
	opts := []nats.Option{}
	opts = append(opts, nats.UserJWTAndSeed(jwt, accessToken))

	// Connect to the NATS server using the provided options.
	service := pubsub.MustConnect(
		pubsub.Config{
			URI:  natsUrl,
			Opts: opts,
		})
	log.Println("Connected to NATS server.")

	// Create a context with a cancel function to control the cancellation of ongoing operations.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up DB
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		panic(err)
	}

	// Create the required table if it doesn't exist
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS messages (
            uId TEXT PRIMARY KEY,
            data BLOB,
            created TIMESTAMP
        )
    `)
	if err != nil {
		log.Fatal(err)
	}

	// insert
	stmt, err := db.Prepare("INSERT INTO messages(uId, created, data) values(?,?,?)")
	if err != nil {
		panic(err)
	}

	// Add a handler function to process messages received on the exampleSubscribeSubject.
	service.AddHandler(streamUrl, func(data []byte) error {
		var message types.Message
		if err := json.Unmarshal(data, &message); err != nil {
			panic(fmt.Sprintf("failed to unmarshal message: %w", err))
		}

		if debug {
			log.Println(string(data))
		}
		var created = time.Now().UTC()
		uId := utils.CreateSha256Checksum(append([]byte(created.String()), data...))

		result, err := stmt.Exec(uId, created, data)
		if err != nil {
			panic(err)
		}

		// Check the number of rows affected to determine if the insert was successful
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			log.Println("No rows were affected, insert may have failed")
			os.Exit(1)
		}
		return nil
	})

	// Set up signal handling to gracefully shut down when receiving SIGINT or SIGTERM signals.
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signalChan
		cancel()
	}()

	// Start serving messages and processing them using the registered handler function.
	if err := service.Serve(ctx); err != nil {
		println("\nStopped syntryve")
		os.Exit(1)
	}
}
