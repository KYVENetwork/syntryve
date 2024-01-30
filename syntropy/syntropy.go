package syntropy

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/KYVENetwork/syntryve/utils"
	"github.com/SyntropyNet/pubsub-go/pubsub"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nats-io/nats.go"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

const (
	exampleStreamName = "osmosis-stream"
	streamSize        = 20_000
	streamTTL         = 24 * time.Hour
)

var (
	mu        sync.Mutex
	logger    = utils.SyntryveLogger("syntropy")
	received  int
	startTime time.Time
)

func StartSyntropyWS(accessToken, natsUrl, streamUrl, consumerId, dbPath string, debug bool) {
	jwt, _ := pubsub.CreateAppJwt(accessToken)
	if debug {
		logger.Info().Msg(jwt)
	}

	nc, err := nats.Connect(natsUrl,
		nats.DrainTimeout(10*time.Second),
		nats.UserJWTAndSeed(jwt, accessToken),
	)
	if err != nil {
		logger.Error().Str("err", err.Error()).Msg("failed to connect to NATS")
		os.Exit(1)
	}
	defer nc.Close()

	logger.Info().Msg("Connected to NATS server.")

	js, err := nc.JetStream()
	if err != nil {
		logger.Error().Str("err", err.Error()).Msg("failed to add JetStream")
		os.Exit(1)
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
		logger.Error().Str("err", err.Error()).Str("example-stream-name", exampleStreamName).Msg("error during adding of JetStream")
		os.Exit(1)
	}
	if debug {
		logger.Info().Msg(fmt.Sprintf("JetStream %s added successfully.\n", exampleStreamName))
	}

	_, err = js.AddConsumer(exampleStreamName, &nats.ConsumerConfig{
		Durable: consumerId,
	})
	if err != nil {
		logger.Error().Msg(fmt.Sprintf("Error during adding consumer of JetStream %s: %s", exampleStreamName, err))
		os.Exit(1)
	}

	if debug {
		logger.Info().Msg(fmt.Sprintf("Consumer to stream %s added successfully.\n", exampleStreamName))
	}

	// Set up DB
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		panic(err)
	}

	// Create the required table if it doesn't exist
	mu.Lock()
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS messages (
            uId TEXT PRIMARY KEY,
            data BLOB,
            created TIMESTAMP
        )
    `)
	mu.Unlock()
	if err != nil {
		logger.Error().Str("err", err.Error()).Msg("failed to create table")
		os.Exit(1)
	}

	// insert
	stmt, err := db.Prepare("INSERT INTO messages(uId, created, data) values(?,?,?)")
	if err != nil {
		logger.Error().Str("err", err.Error()).Msg("failed to prepare query")
		os.Exit(1)
	}

	subscription, err := js.PullSubscribe(streamUrl, consumerId, nats.ManualAck(), nats.Bind(exampleStreamName, consumerId))
	if err != nil {
		logger.Error().Str("err", err.Error()).Msg("error during pull subscription")
	}

	startTime = time.Now()

	go func() {
		for {
			msgs, err := subscription.Fetch(1, nats.Context(ctx))
			if err != nil {
				if errors.Is(err, context.Canceled) {
					logger.Error().Str("err", err.Error()).Msg("error during fetching subscription")
					return
				}
				logger.Error().Str("err", err.Error()).Msg("error during pulling next message")
			}

			if len(msgs) == 1 {
				msg := msgs[0]

				timestamp, err := strconv.ParseInt(msg.Header.Get("timestamp"), 10, 64)
				if err != nil {
					logger.Error().Str("err", err.Error()).Msg("failed to parse header")
					os.Exit(1)
				}

				created := time.Unix(timestamp/1e9, 0)
				uId := utils.CreateSha256Checksum(append([]byte(created.String()), msg.Data...))

				if debug {
					logger.Info().Msg(fmt.Sprint("Received message ", created, " with size ", len(msg.Data)))
					logger.Info().Msg(string(msg.Data))
				}

				mu.Lock()
				result, err := stmt.Exec(uId, created, msg.Data)
				if err != nil {
					logger.Error().Str("err", err.Error()).Msg("failed during query execution")
					os.Exit(1)
				}

				// Check the number of rows affected to determine if the insert was successful
				rowsAffected, _ := result.RowsAffected()
				if rowsAffected == 0 {
					logger.Error().Msg("no rows were affected, insert may have failed")
					os.Exit(1)
				}
				mu.Unlock()

				if err = msg.Ack(); err != nil {
					logger.Error().Str("err", err.Error()).Msg("error during message ack")
				}

				received++
				if time.Since(startTime) > 60*time.Second {
					logger.Info().Msgf("Received %d messages in the last 60 seconds", received)
					received = 0
					startTime = time.Now()
				}
			}
		}
	}()

	// Set up signal handling to gracefully shut down when receiving SIGINT or SIGTERM signals.
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan
	if err := nc.Drain(); err != nil {
		logger.Error().Str("err", err.Error()).Msg("error during connection close")
	}

	logger.Info().Msg("subscription stopped")
	os.Exit(0)
}
