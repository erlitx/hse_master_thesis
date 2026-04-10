package mcpserver

import "github.com/mark3labs/mcp-go/mcp"

func (h *Handler) registerTimeResources() {
	// A simple fixed resource:
	//   time://now
	res := mcp.NewResource(
		"time://now",
		"Current time",
		mcp.WithResourceDescription("Returns the current server time in RFC3339Nano (UTC)."),
		mcp.WithMIMEType("text/plain"),
	)
	h.srv.AddResource(res, func(req mcp.ReadResourceRequest) ([]interface{}, error) {
		_ = req
		content := mcp.TextResourceContents{
			ResourceContents: mcp.ResourceContents{URI: "time://now", MIMEType: "text/plain"},
			Text:             h.uc.NowISO(),
		}
		return []interface{}{content}, nil
	})
}
