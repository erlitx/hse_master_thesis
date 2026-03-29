package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Config struct {
	AppName       string `envconfig:"APP_NAME" required:"true"`
	AppVersion    string `envconfig:"APP_VERSION" required:"true"`
	Level         string `envconfig:"LOGGER_LEVEL" default:"error"`
	PrettyConsole bool   `envconfig:"LOGGER_PRETTY_CONSOLE" default:"false"`
	Env           string `envconfig:"APP_ENV" default:"prod"`
}

func Init(c Config) {
	zerolog.TimeFieldFormat = time.RFC3339

	level, err := zerolog.ParseLevel(c.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	var output io.Writer = os.Stderr
	if c.PrettyConsole {
		output = zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"}
	}

	log.Logger = zerolog.New(output).
		Level(level).
		With().
		Timestamp().
		Caller().
		// Str("app_name", c.AppName).
		// Str("app_version", c.AppVersion).
		Logger()

	log.Info().Msg("Logger initialized")
}
