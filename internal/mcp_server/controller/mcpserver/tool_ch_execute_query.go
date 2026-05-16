package mcpserver

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
)

// Обработчик read-only запроса ClickHouse
func (h *Handler) clickHouseQuery(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	q, err := req.RequireString("query")
	if err != nil {
		return mcp.NewToolResultError("missing/invalid argument: query (string)"), nil
	}
	result, err := h.uc.ClickHouseQueryReadOnly(ctx, q)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	b, err := json.Marshal(result)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(b)), nil
}
