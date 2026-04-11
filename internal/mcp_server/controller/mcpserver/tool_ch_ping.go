package mcpserver

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
)

func (h *Handler) clickHousePing(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	err := h.uc.ClickHousePing(ctx)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText("ok"), nil
}
