package config

import (
	"fmt"
	"os"

	clickhousepkg "github.com/erlitx/mcp_server/pkg/clickhouse"
	"github.com/erlitx/mcp_server/pkg/httpserver"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog/log"
)

type MCPServer struct {
	Name    string `envconfig:"APP_NAME" default:"mcp_server"`
	Version string `envconfig:"APP_VERSION" default:"v0.1.0"`
	Env     string `envconfig:"APP_ENV" default:"dev"`
}

type MCP struct {
	// Transport currently supported in this template:
	//   - http  (JSON-RPC over HTTP POST)
	Transport string `envconfig:"MCP_TRANSPORT" default:"http"`

	// HTTP server address and path.
	Addr string `envconfig:"MCP_ADDR" default:":8081"`
	Path string `envconfig:"MCP_PATH" default:"/mcp"`
}

type MCPClient struct {
	Name    string `envconfig:"MCP_CLIENT_NAME" default:"mcp_client"`
	Version string `envconfig:"MCP_CLIENT_VERSION" default:"v0.1.0"`
	Env     string `envconfig:"APP_ENV" default:"dev"`
}

type Claude struct {
	APIKey    string `envconfig:"CLAUDE_API_KEY" required:"true"`
	Model     string `envconfig:"CLAUDE_MODEL" default:"claude-opus-4-6"`
	MaxTokens int    `envconfig:"CLAUDE_MAX_TOKENS" default:"4096"`
}

type MCPServerConnection struct {
	Addr string `envconfig:"MCP_SERVER_ADDR" required:"true"`
}

type Config struct {
	MCPServer           MCPServer
	MCP                 MCP
	MCPClient           MCPClient
	Claude              Claude
	MCPServerConnection MCPServerConnection
	HTTP                httpserver.Config
	ClickHouse          clickhousepkg.Config
}

func New() (Config, error) {
	var config Config

	// Detect which environment to load
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev"
	}
	envFile := ".env." + env

	// Only load .env file if not inside Docker
	if os.Getenv("DOCKER") != "true" {
		log.Info().Msg(fmt.Sprintf("Loading from: %s", envFile))
		err := godotenv.Overload(envFile)
		if err != nil {
			return config, fmt.Errorf("godotenv.Load: %w", err)
		}
	}

	// Load config from environment variables
	err := envconfig.Process("", &config)
	if err != nil {
		return config, fmt.Errorf("envconfig.Process: %w", err)
	}

	return config, nil
}
