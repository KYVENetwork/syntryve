package syntropy

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/KYVENetwork/syntropy-demo/utils"
	"github.com/SyntropyNet/pubsub-go/pubsub"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nats-io/nats.go"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

const (
	exampleStreamName  = "osmosis-stream"
	uniqueConsumerName = "troykessler"
	streamSize         = 20_000
	streamTTL          = 24 * time.Hour
)

func StartSyntropyWS(accessToken, natsUrl, streamUrl, dbPath string, debug bool) {
	jwt, _ := pubsub.CreateAppJwt(accessToken)
	if debug {
		println(jwt)
	}

	nc, err := nats.Connect(natsUrl,
		nats.DrainTimeout(10*time.Second),
		nats.UserJWTAndSeed(jwt, accessToken),
	)
	if err != nil {
		fmt.Println("failed to connect ", err)
	}
	defer nc.Close()
	fmt.Println("Connected to NATS server.")

	js, err := nc.JetStream()
	if err != nil {
		fmt.Println(err.Error())
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	connCtx, connCancelFn := context.WithTimeout(ctx, 10*time.Second)
	defer connCancelFn()

	_, err = js.AddStream(&nats.StreamConfig{
		Name:     exampleStreamName,
		Subjects: []string{streamUrl},
		MaxMsgs:  streamSize,
		MaxAge:   streamTTL,
	}, nats.Context(connCtx))
	if err != nil {
		fmt.Printf("Error during adding of JetStream %s: %s", exampleStreamName, err)
	}
	fmt.Printf("JetStream %s added successfully.\n", exampleStreamName)

	_, err = js.AddConsumer(exampleStreamName, &nats.ConsumerConfig{
		Durable: uniqueConsumerName,
	})
	if err != nil {
		fmt.Printf("Error during adding consumer of JetStream %s: %s", exampleStreamName, err)
	}
	fmt.Printf("Consumer to stream %s added successfully.\n", exampleStreamName)

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

	subscription, err := js.PullSubscribe(streamUrl, uniqueConsumerName, nats.ManualAck(), nats.Bind(exampleStreamName, uniqueConsumerName))
	if err != nil {
		fmt.Print("Error during pull subscription: ", err)
	}

	go func() {
		for {
			msgs, err := subscription.Fetch(1, nats.Context(ctx))
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return
				}
				fmt.Print("Error during pulling next message: ", err)
			}
			msg := msgs[0]

			timestamp, err := strconv.ParseInt(msg.Header.Get("timestamp"), 10, 64)
			if err != nil {
				panic(err)
			}

			created := time.Unix(timestamp/1e9, 0)
			uId := utils.CreateSha256Checksum(append([]byte(created.String()), msg.Data...))

			fmt.Println("Received message ", created, " with size ", len(msg.Data))

			if debug {
				fmt.Println(string(msg.Data))
			}

			result, err := stmt.Exec(uId, created, msg.Data)
			if err != nil {
				panic(err)
			}

			// Check the number of rows affected to determine if the insert was successful
			rowsAffected, _ := result.RowsAffected()
			if rowsAffected == 0 {
				log.Println("No rows were affected, insert may have failed")
				os.Exit(1)
			}

			if err := msg.Ack(); err != nil {
				fmt.Println("Error during message ack: ", err)
			}
		}
	}()

	// Set up signal handling to gracefully shut down when receiving SIGINT or SIGTERM signals.
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan
	if err := nc.Drain(); err != nil {
		fmt.Println("Error during connection close: ", err)
	}

	fmt.Println("Subscription stopped")
}
