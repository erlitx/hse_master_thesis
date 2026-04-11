package mcpserver

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
)

func (h *Handler) registerClickHouseTools() {
	// ch_ping() -> text
	pingTool := mcp.NewTool(
		"ch_ping",
		mcp.WithDescription("Ping ClickHouse using the configured adapter."),
	)

	h.srv.AddTool(pingTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return h.clickHousePing(ctx, req)
	})

	// ch_query(query:string) -> json rows (limited)
	queryTool := mcp.NewTool(
		"ch_query",
		mcp.WithDescription("Execute a read-only ClickHouse query (SELECT/SHOW/DESCRIBE/EXPLAIN)."),
		mcp.WithString("query", mcp.Description("SQL query to execute"), mcp.Required()),
	)

	h.srv.AddTool(queryTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return h.clickHouseQuery(ctx, req)
	})
}
