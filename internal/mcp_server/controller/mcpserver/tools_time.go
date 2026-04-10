package mcpserver

import "github.com/mark3labs/mcp-go/mcp"

func (h *Handler) registerTimeTools() {
	// now() -> ISO timestamp (UTC)
	nowTool := mcp.NewTool(
		"now",
		mcp.WithDescription("Get current time as RFC3339Nano in UTC."),
	)
	h.srv.AddTool(nowTool, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		_ = args
		return mcp.NewToolResultText(h.uc.NowISO()), nil
	})
}
