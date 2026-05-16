package service

import "fmt"

// Категория MCP-инструмента.
type ToolCategory string

const (
	ToolCategoryClickHouse ToolCategory = "clickhouse"
	ToolCategoryMetadata   ToolCategory = "metadata"
)

var toolRoutes = map[string]ToolCategory{
	"ch_ping":         ToolCategoryClickHouse,
	"ch_query":        ToolCategoryClickHouse,
	"list_dwh_models": ToolCategoryMetadata,
	"get_dwh_model":   ToolCategoryMetadata,
}

// ToolRouter маршрутизирует вызовы MCP tools по категориям.
type ToolRouter struct{}

// NewToolRouter создаёт маршрутизатор инструментов.
func NewToolRouter() *ToolRouter {
	return &ToolRouter{}
}

// Route возвращает категорию зарегистрированного инструмента.
func (r *ToolRouter) Route(toolName string) (ToolCategory, error) {
	if r == nil {
		return "", fmt.Errorf("tool router is nil")
	}

	category, ok := toolRoutes[toolName]
	if !ok {
		return "", fmt.Errorf("unknown tool: %s", toolName)
	}
	return category, nil
}

// IsRegistered сообщает, зарегистрирован ли инструмент.
func (r *ToolRouter) IsRegistered(toolName string) bool {
	_, ok := toolRoutes[toolName]
	return ok
}

// RegisteredTools возвращает копию карты маршрутов (для тестов и диагностики).
func (r *ToolRouter) RegisteredTools() map[string]ToolCategory {
	out := make(map[string]ToolCategory, len(toolRoutes))
	for name, cat := range toolRoutes {
		out[name] = cat
	}
	return out
}
