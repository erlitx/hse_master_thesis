package mcpserver

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
)

// Регистрирует MCP-промпт аналитика DWH (генерация SQL для ClickHouse).
func (h *Handler) registerGreetingPrompt() {
	p := mcp.NewPrompt(
		"greet",
		mcp.WithPromptDescription("Промпт аналитика DWH: аналитические запросы к ClickHouse, ответ — SQL."),
		mcp.WithArgument(
			"question",
			mcp.ArgumentDescription("Аналитический вопрос или описание требуемой выборки"),
		),
	)

	h.srv.AddPrompt(p, func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		question := ""
		if request.Params.Arguments != nil {
			question = request.Params.Arguments["question"]
		}

		systemMsg := mcp.NewPromptMessage(
			mcp.RoleUser,
			mcp.NewTextContent(h.uc.Greet(question)),
		)
		taskMsg := mcp.NewPromptMessage(
			mcp.RoleUser,
			mcp.NewTextContent(h.uc.GreetUserMessage(question)),
		)

		return mcp.NewGetPromptResult(
			"Аналитик DWH: генерация SQL",
			[]mcp.PromptMessage{systemMsg, taskMsg},
		), nil
	})
}
