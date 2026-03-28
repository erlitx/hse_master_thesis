package clickhouse

import (
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
)


type ClickHouse struct {
	conn     clickhouse.Conn
}

func New(ch clickhouse.Conn) *ClickHouse {
	return &ClickHouse{
		conn: ch,
	}
}

func (ch *ClickHouse) Conn() clickhouse.Conn {
	return ch.conn
}

func (ch *ClickHouse) Close() error {
	if ch == nil || ch.conn == nil {
		return nil
	}

	if err := ch.conn.Close(); err != nil {
		return fmt.Errorf("clickhouse - Close - ch.conn.Close: %w", err)
	}

	return nil
}