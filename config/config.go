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

// Метаданные MCP-сервера
type MCPServer struct {
	Name    string `envconfig:"APP_NAME" default:"mcp_server"`
	Version string `envconfig:"APP_VERSION" default:"v0.1.0"`
	Env     string `envconfig:"APP_ENV" default:"dev"`
}

// Настройки транспорта MCP
type MCP struct {
	// Транспорт MCP (http — JSON-RPC по POST)
	Transport string `envconfig:"MCP_TRANSPORT" default:"http"`

	// Адрес и путь HTTP-сервера MCP
	Addr string `envconfig:"MCP_ADDR" default:":8081"`
	Path string `envconfig:"MCP_PATH" default:"/mcp"`
}

// Метаданные MCP-клиента
type MCPClient struct {
	Name    string `envconfig:"MCP_CLIENT_NAME" default:"mcp_client"`
	Version string `envconfig:"MCP_CLIENT_VERSION" default:"v0.1.0"`
	Env     string `envconfig:"APP_ENV" default:"dev"`
}

// Параметры API Claude
type Claude struct {
	APIKey    string `envconfig:"CLAUDE_API_KEY" required:"true"`
	Model     string `envconfig:"CLAUDE_MODEL" default:"claude-opus-4-6"`
	MaxTokens int    `envconfig:"CLAUDE_MAX_TOKENS" default:"4096"`
}

// Адрес подключения к MCP-серверу
type MCPServerConnection struct {
	Addr string `envconfig:"MCP_SERVER_ADDR" required:"true"`
}

// Учётные данные Superset
type Superset struct {
	URL      string `envconfig:"SUPERSET_URL" required:"true"`
	User     string `envconfig:"SUPERSET_USER" required:"true"`
	Password string `envconfig:"SUPERSET_PASSWORD" required:"true"`
}

// Параметры dbt
type DBT struct {
	ManifestPath string `envconfig:"DBT_MANIFEST_PATH" default:"./manifest.json"`
}

// Корневая конфигурация приложения
type Config struct {
	MCPServer           MCPServer
	MCP                 MCP
	MCPClient           MCPClient
	Claude              Claude
	MCPServerConnection MCPServerConnection
	Superset            Superset
	DBT                 DBT
	HTTP                httpserver.Config
	ClickHouse          clickhousepkg.Config
}

// Создаёт новый экземпляр
func New() (Config, error) {
	var config Config

	// Определяем окружение для загрузки
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev"
	}
	envFile := ".env." + env

	// Загружаем .env только вне Docker
	if os.Getenv("DOCKER") != "true" {
		log.Info().Msg(fmt.Sprintf("Loading from: %s", envFile))
		err := godotenv.Overload(envFile)
		if err != nil {
			return config, fmt.Errorf("godotenv.Load: %w", err)
		}
	}

	// Читаем конфиг из переменных окружения
	err := envconfig.Process("", &config)
	if err != nil {
		return config, fmt.Errorf("envconfig.Process: %w", err)
	}

	return config, nil
}
