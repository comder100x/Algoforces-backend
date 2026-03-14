package utils

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Init(env string) {
	// Default to dev if empty
	if env == "" {
		env = "dev"
	}
	// 1) Global level
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if env == "dev" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// 2) Output format — use io.Writer as the common type
	var out io.Writer = os.Stdout
	if env == "dev" {
		out = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	// 3) Build logger (timestamp is commonly included)
	logger := zerolog.New(out).With().
		Timestamp().
		Str("service", "algoforces").
		Logger()

	// 4) Set as global logger (so log.Info() uses it)
	log.Logger = logger
}
