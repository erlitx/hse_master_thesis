package clickhouse

import (
	"context"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/rs/zerolog/log"
)

// Корневая конфигурация приложения
type Config struct {
	User     string `envconfig:"CLICKHOUSE_USER"`
	Password string `envconfig:"CLICKHOUSE_PASSWORD"`
	Host     string `envconfig:"CLICKHOUSE_HOST"`
	Port     string `envconfig:"CLICKHOUSE_PORT"`
	DBName   string `envconfig:"CLICKHOUSE_DB"`
	Secure   bool   `envconfig:"CLICKHOUSE_SECURE" default:"false"`
}

// Пул соединений ClickHouse
type Pool struct {
	conn clickhouse.Conn
}

// Создаёт новый экземпляр
func New(ctx context.Context, c Config) (*Pool, error) {
	addr := fmt.Sprintf("%s:%s", c.Host, c.Port)
	log.Debug().Msgf("Connecting to ClickHouse at %s", addr)
	opts := &clickhouse.Options{
		Addr: []string{addr},
		Auth: clickhouse.Auth{
			Database: c.DBName,
			Username: c.User,
			Password: c.Password,
		},
		// optional secure connection
		TLS: nil,
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
	}

	conn, err := clickhouse.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("clickhouse.Open: %w", err)
	}

	return &Pool{conn: conn}, nil
}

// Возвращает соединение ClickHouse
func (p *Pool) Conn() clickhouse.Conn {
	return p.conn
}

// Закрывает соединение
func (p *Pool) Close() {
	if p.conn != nil {
		err := p.conn.Close()
		if err != nil {
			log.Error().Err(err).Msg("failed to close ClickHouse connection")
		} else {
			log.Warn().Msg("ClickHouse connection was already nil")
		}

		log.Info().Msg("ClickHouse connection closed")
	}
}
