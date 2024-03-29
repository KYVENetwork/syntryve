package pruner

import (
	"database/sql"
	"fmt"
	"github.com/KYVENetwork/syntryve/utils"
	"sync"
)

var (
	logger = utils.SyntryveLogger("pruner")
	mu     sync.Mutex
)

func PruneDB(until, dbPath string) error {
	logger.Info().Msg(fmt.Sprintf("Starting pruning until %v", until))

	// Set up DB
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open db: %v", err)
	}

	stmt, err := db.Prepare("DELETE FROM messages WHERE strftime('%s', created) < ?")
	if err != nil {
		return fmt.Errorf("failed to prepare table for pruning: %v", err)
	}

	mu.Lock()
	result, err := stmt.Exec(until)
	if err != nil {
		return fmt.Errorf("failed to prune db: %v", err)
	}

	rowsAffected, _ := result.RowsAffected()
	logger.Info().Msg(fmt.Sprintf("Deleted %v entries", rowsAffected))

	// VACUUM the database to reclaim unused space
	_, err = db.Exec("VACUUM")
	if err != nil {
		mu.Unlock()
		return fmt.Errorf("failed to vacuum db: %v", err)
	}

	mu.Unlock()

	return nil
}
