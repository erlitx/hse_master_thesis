package mcpserver

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
)

func (h *Handler) clickHouseQuery(args map[string]interface{}) (*mcp.CallToolResult, error) {
	q, ok := args["query"].(string)
	if !ok {
		return mcp.NewToolResultError("missing/invalid argument: query (string)"), nil
	}
	rows, err := h.uc.ClickHouseQueryReadOnly(context.Background(), q)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	b, err := json.Marshal(rows)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(b)), nil
}
