package src

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	Cookie  string   `mapstructure:"cookie" json:"cookie"`
	OutPath string   `mapstructure:"outPath" json:"outPath"`
	Limit   int      `mapstructure:"limit" json:"limit"`
	Urls    []string `mapstructure:"urls" json:"urls"`
}

func (c *Config) LoadConfig(path string) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&c); err != nil {
		log.Fatal(err)
	}

	// Limit max 200
	if c.Limit <= 0 || c.Limit > 200 {
		c.Limit = 200
	}
}
