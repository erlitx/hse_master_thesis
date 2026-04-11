package mcpserver

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

func (h *Handler) registerMathTools() {
	// add(a:number, b:number) -> text result with sum
	addTool := mcp.NewTool(
		"add",
		mcp.WithDescription("Add two numbers (a + b)."),
		mcp.WithNumber("a", mcp.Description("First number"), mcp.Required()),
		mcp.WithNumber("b", mcp.Description("Second number"), mcp.Required()),
	)
	h.srv.AddTool(addTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		a, err := req.RequireFloat("a")
		if err != nil {
			return mcp.NewToolResultError("missing/invalid argument: a (number)"), nil
		}
		b, err := req.RequireFloat("b")
		if err != nil {
			return mcp.NewToolResultError("missing/invalid argument: b (number)"), nil
		}

		sum := h.uc.Add(a, b)
		return mcp.NewToolResultText(fmt.Sprintf("%g", sum)), nil
	})

	// echo(text:string) -> text
	echoTool := mcp.NewTool(
		"echo",
		mcp.WithDescription("Echo back a string."),
		mcp.WithString("text", mcp.Description("Text to echo"), mcp.Required()),
	)
	h.srv.AddTool(echoTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		v, err := req.RequireString("text")
		if err != nil {
			return mcp.NewToolResultError("missing/invalid argument: text (string)"), nil
		}
		return mcp.NewToolResultText(h.uc.Echo(v)), nil
	})
}
