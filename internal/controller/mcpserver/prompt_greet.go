package mcpserver

import "github.com/mark3labs/mcp-go/mcp"

func (h *Handler) registerGreetingPrompt() {
	p := mcp.NewPrompt(
		"greet",
		mcp.WithPromptDescription("Returns a greeting message."),
		mcp.WithArgument("name", mcp.ArgumentDescription("Who to greet")),
	)
	h.srv.AddPrompt(p, func(args map[string]string) (*mcp.GetPromptResult, error) {
		name := args["name"]
		msg := mcp.NewPromptMessage("assistant", mcp.NewTextContent(h.uc.Greet(name)))
		return mcp.NewGetPromptResult("Greeting", []mcp.PromptMessage{msg}), nil
	})
}
