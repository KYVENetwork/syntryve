package pool

import (
	"encoding/json"
	"fmt"
	"github.com/KYVENetwork/syntryve/types"
	"github.com/KYVENetwork/syntryve/utils"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	logger = utils.SyntryveLogger("pool")
)

// GetLatestPoolKey retrieves the current KYVE pool key.
func GetLatestPoolKey(chainId string, poolId int64, poolEndpoints string) (string, error) {
	var endpoints []string

	if poolEndpoints != "" {
		endpoints = strings.Split(poolEndpoints, ",")
	} else {
		if chainId == "korellia-2" {
			endpoints = utils.KorelliaEndpoints
		} else if chainId == "kaon-1" {
			endpoints = utils.KaonEndpoints
		} else if chainId == "kyve-1" {
			endpoints = utils.MainnetEndpoints
		} else {
			return "", fmt.Errorf("unknown chainId")
		}
	}

	for i := 0; i <= utils.BackoffMaxRetries; i++ {
		delay := time.Duration(math.Pow(2, float64(i))) * time.Second

		for _, endpoint := range endpoints {
			latestKey, err := requestLatestPoolKey(poolId, endpoint)
			if err == nil {
				return latestKey, nil
			} else {
				logger.Error().Str("endpoint", endpoint).Str("err", err.Error()).Msg("failed to request pool height")
			}
		}

		if i <= utils.BackoffMaxRetries {
			logger.Info().Msg(fmt.Sprintf("retrying to query pool again in %v\n", delay))
			time.Sleep(delay)
		}
	}

	return "", fmt.Errorf("failed to get pool height from all endpoints")
}

// requestLatestPoolKey retrieves the latest KYVE pool key by making an GET request to the given endpoint.
func requestLatestPoolKey(poolId int64, endpoint string) (string, error) {
	poolEndpoint := endpoint + "/kyve/query/v1beta1/pool/" + strconv.FormatInt(poolId, 10)

	response, err := http.Get(poolEndpoint)
	if err != nil {
		return "", fmt.Errorf("failed requesting KYVE endpoint: %s", err)
	}

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("failed reading KYVE endpoint response: %s", err)
	}

	var resp types.SettingsResponse
	err = json.Unmarshal(responseData, &resp)
	if err != nil {
		return "", fmt.Errorf("failed unmarshalling KYVE endpoint response: %s", err)
	}

	return resp.Pool.Data.CurrentKey, nil
}
