package lib

import (
	"encoding/json"
	"os"
)

// Msg type for collectd-proxy binaries
type Msg []byte

// Config struct for the collectd-proxy binaries.
// The server will start listeners on these addresses,
// while the client will send requests there.
type Config struct {
	UDPAddress  string
	HTTPAddress string
}

// GetConfig returns a Config struct with the addresses to be used
func GetConfig(filename string) (Config, error) {
	// 256 bytes should be more than enough for the config json
	rawConfig := make([]byte, 256)
	config := Config{}

	file, err := os.Open(filename)
	if err != nil {
		return config, err
	}

	n, err := file.Read(rawConfig)
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(rawConfig[:n], &config)
	if err != nil {
		return config, err
	}

	return config, nil
}
