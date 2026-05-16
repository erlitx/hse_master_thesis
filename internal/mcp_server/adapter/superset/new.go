package superset

import (
	"net/http"
	"strings"
	"time"

	"github.com/erlitx/mcp_server/config"
)

// Корневая конфигурация приложения
type Config struct {
	URL      string
	User     string
	Password string
	JWTToken string
}

// Клиент Superset (BI)
type BiTool struct {
	cfg        Config
	httpClient *http.Client
}

// Создаёт новый экземпляр
func New(cfg config.Superset) *BiTool {
	return &BiTool{
		cfg: Config{
			URL:      strings.TrimRight(cfg.URL, "/"),
			User:     cfg.User,
			Password: cfg.Password,
		},
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}
