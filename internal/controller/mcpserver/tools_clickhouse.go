package mcpserver

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func (h *Handler) registerClickHouseTools() {
	// ch_ping() -> text
	pingTool := mcp.NewTool(
		"ch_ping",
		mcp.WithDescription("Ping ClickHouse using the configured adapter."),
	)
	
	h.srv.AddTool(pingTool, h.clickHousePing)


	// ch_query(query:string) -> json rows (limited)
	queryTool := mcp.NewTool(
		"ch_query",
		mcp.WithDescription("Execute a read-only ClickHouse query (SELECT/SHOW/DESCRIBE/EXPLAIN)."),
		mcp.WithString("query", mcp.Description("SQL query to execute"), mcp.Required()),
	)

	h.srv.AddTool(queryTool, h.clickHouseQuery)

}
