package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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

	level := zerolog.DebugLevel // default for dev
	if strings.EqualFold(c.Env, "prod") || strings.EqualFold(c.Env, "staging") {
		level = zerolog.ErrorLevel
	}

	// Allow manual override
	if c.Level != "" {
		if parsed, err := zerolog.ParseLevel(c.Level); err == nil {
			level = parsed
		}
	}

	zerolog.SetGlobalLevel(level)

	// --- Console output configuration ---
	var output io.Writer = os.Stderr
	if c.PrettyConsole {
		output = zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: "2006-01-02 15:04:05.000",
			// ✅ remove color on Windows or when not a TTY
			NoColor: runtime.GOOS == "windows" || !isTerminal(os.Stderr.Fd()),
		}
	}

	// ✅ shorten long file paths like ..\..\Users\Mihail\...
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		return fmt.Sprintf("%s:%d", trimToProject(file, "file_sync_app_win"), line)
	}

	log.Logger = zerolog.New(output).
		Level(level).
		With().
		Timestamp().
		Caller().
		// Str("app", c.AppName).
		// Str("ver", c.AppVersion).
		Logger()

	log.Info().Msg("Logger initialized")
	log.Error().Msg("Logger test error message")
}

// --- Helper to trim long source paths ---
func trimToProject(file, project string) string {
	file = filepath.ToSlash(file)
	if idx := strings.Index(file, project+"/"); idx != -1 {
		return file[idx+len(project)+1:]
	}
	return filepath.Base(file)
}

// --- Helper to detect if output is a terminal ---
func isTerminal(fd uintptr) bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

//func Init(c Config) {
//	zerolog.TimeFieldFormat = time.RFC3339
//
//	level, err := zerolog.ParseLevel(c.Level)
//	if err != nil {
//		level = zerolog.InfoLevel
//	}
//	zerolog.SetGlobalLevel(level)
//
//	var output io.Writer = os.Stderr
//	if c.PrettyConsole {
//		output = zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"}
//	}
//
//	log.Logger = zerolog.New(output).
//		Level(level).
//		With().
//		Timestamp().
//		Caller().
//		// Str("app_name", c.AppName).
//		// Str("app_version", c.AppVersion).
//		Logger()
//
//	log.Info().Msg("Logger initialized")
//}
