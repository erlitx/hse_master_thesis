package mcpserver

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
)

func (h *Handler) registerTimeTools() {
	// now() -> ISO timestamp (UTC)
	nowTool := mcp.NewTool(
		"now",
		mcp.WithDescription("Get current time as RFC3339Nano in UTC."),
	)
	h.srv.AddTool(nowTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText(h.uc.NowISO()), nil
	})
}
