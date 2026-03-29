package mcpserver

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
)

func (h *Handler) clickHousePing(args map[string]interface{}) (*mcp.CallToolResult, error) {
	_ = args
	err := h.uc.ClickHousePing(context.Background())
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText("ok"), nil
}