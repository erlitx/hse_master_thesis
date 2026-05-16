package logger

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Корневая конфигурация приложения
type Config struct {
	AppName       string `envconfig:"APP_NAME" required:"true"`
	AppVersion    string `envconfig:"APP_VERSION" required:"true"`
	Level         string `envconfig:"LOGGER_LEVEL" default:"debug"`
	PrettyConsole bool   `envconfig:"LOGGER_PRETTY_CONSOLE" default:"true"`
	Env           string `envconfig:"APP_ENV" default:"prod"`
}

// Инициализирует логгер
func Init(c Config) {
	// --- parse level ---
	level, err := zerolog.ParseLevel(strings.ToLower(c.Level))
	if err != nil {
		level = zerolog.InfoLevel
	}

	// --- time format ---
	zerolog.TimeFieldFormat = time.RFC3339

	// --- output ---
	var out io.Writer = os.Stdout
	if c.PrettyConsole {
		out = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "15:04:05",
		}
	}

	// --- build logger ---
	logger := zerolog.New(out).
		Level(level).
		With().
		Timestamp().
		Str("app", c.AppName).
		Str("version", c.AppVersion).
		Str("env", c.Env).
		Caller().
		Logger()

	log.Logger = logger
	zerolog.SetGlobalLevel(level)

	// --- force visibility ---
	log.Info().
		Str("level", level.String()).
		Msg("logger initialized")
}
