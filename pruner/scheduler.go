package pruner

import (
	"fmt"
	"github.com/KYVENetwork/syntryve/pool"
	"github.com/KYVENetwork/syntryve/utils"
	"os"
	"time"
)

func StartPruningScheduler(pruningInterval int64, poolId int64, dbPath, chainId, poolEndpoints string, debug bool) {
	var pruningCount float64 = 0

	for {
		if debug {
			logger.Info().Msg(fmt.Sprintf("pruning count = %v; pruning interval = %v", pruningCount, pruningInterval))
		}
		latestKey, err := pool.GetLatestPoolKey(chainId, poolId, poolEndpoints)
		if err != nil {
			logger.Error().Msg(fmt.Sprintf("failed to get latest pool key: %v", err))
			os.Exit(1)
		}
		if pruningCount > float64(pruningInterval) {
			if err = PruneDB(latestKey, dbPath); err != nil {
				logger.Error().Msg(fmt.Sprintf("failed to prune db: %v", err))
				os.Exit(1)
			}
			pruningCount = 0
		}
		pruningCount = pruningCount + float64(utils.Interval)/60
		time.Sleep(time.Second * time.Duration(utils.Interval))
	}
}
