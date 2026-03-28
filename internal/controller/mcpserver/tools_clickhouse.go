package mcpserver

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
)

func (h *Handler) registerClickHouseTools() {
	// ch_ping() -> text
	pingTool := mcp.NewTool(
		"ch_ping",
		mcp.WithDescription("Ping ClickHouse using the configured adapter."),
	)
	h.srv.AddTool(pingTool, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		_ = args
		if err := h.uc.ClickHousePing(context.Background()); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText("ok"), nil
	})

	// ch_query(query:string) -> json rows (limited)
	queryTool := mcp.NewTool(
		"ch_query",
		mcp.WithDescription("Execute a read-only ClickHouse query (SELECT/SHOW/DESCRIBE/EXPLAIN)."),
		mcp.WithString("query", mcp.Description("SQL query to execute"), mcp.Required()),
	)
	h.srv.AddTool(queryTool, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		q, ok := args["query"].(string)
		if !ok {
			return mcp.NewToolResultError("missing/invalid argument: query (string)"), nil
		}
		rows, err := h.uc.ClickHouseQueryReadOnly(context.Background(), q)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		b, _ := json.Marshal(rows)
		return mcp.NewToolResultText(string(b)), nil
	})
}
