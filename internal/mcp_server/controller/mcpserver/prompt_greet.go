package mcpserver

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
)

func (h *Handler) registerGreetingPrompt() {
	p := mcp.NewPrompt(
		"greet",
		mcp.WithPromptDescription("Returns a greeting message."),
		mcp.WithArgument("name", mcp.ArgumentDescription("Who to greet")),
	)

	h.srv.AddPrompt(p, func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		name := ""
		if request.Params.Arguments != nil {
			name = request.Params.Arguments["name"]
		}

		msg := mcp.NewPromptMessage(
			mcp.RoleAssistant,
			mcp.NewTextContent(h.uc.Greet(name)),
		)

		return mcp.NewGetPromptResult(
			"Greeting",
			[]mcp.PromptMessage{msg},
		), nil
	})
}