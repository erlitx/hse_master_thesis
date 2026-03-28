package clickhouse

import (
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

