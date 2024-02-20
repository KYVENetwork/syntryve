package utils

const (
	BackoffMaxRetries       = 15
	Interval          int64 = 5
)

var (
	KaonEndpoints = []string{
		"https://api.kaon.kyve.network",
	}
	KorelliaEndpoints = []string{
		"https://api.korellia.kyve.network",
	}
	MainnetEndpoints = []string{
		"https://api.kyve.network",
	}
)
