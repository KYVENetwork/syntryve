package types

type Message struct {
	Nonce string `json:"nonce"`
}

type SettingsResponse struct {
	Pool struct {
		Data struct {
			StartKey       string `json:"start_key"`
			CurrentKey     string `json:"current_key"`
			UploadInterval string `json:"upload_interval"`
			MaxBundleSize  string `json:"max_bundle_size"`
		} `json:"data"`
	} `json:"pool"`
}
