package service_test

import (
	"testing"

	"github.com/erlitx/mcp_server/internal/mcp_server/service"
)

func TestToolRouter_Route_clickhouse(t *testing.T) {
	t.Parallel()

	router := service.NewToolRouter()
	for _, name := range []string{"ch_ping", "ch_query"} {
		cat, err := router.Route(name)
		if err != nil {
			t.Fatalf("Route(%q): %v", name, err)
		}
		if cat != service.ToolCategoryClickHouse {
			t.Fatalf("Route(%q) = %q, want clickhouse", name, cat)
		}
	}
}

func TestToolRouter_Route_metadata(t *testing.T) {
	t.Parallel()

	router := service.NewToolRouter()
	for _, name := range []string{"list_dwh_models", "get_dwh_model"} {
		cat, err := router.Route(name)
		if err != nil {
			t.Fatalf("Route(%q): %v", name, err)
		}
		if cat != service.ToolCategoryMetadata {
			t.Fatalf("Route(%q) = %q, want metadata", name, cat)
		}
	}
}

func TestToolRouter_Route_unknown(t *testing.T) {
	t.Parallel()

	router := service.NewToolRouter()
	_, err := router.Route("unknown_tool")
	if err == nil {
		t.Fatal("expected error for unknown tool")
	}
}

func TestToolRouter_IsRegistered(t *testing.T) {
	t.Parallel()

	router := service.NewToolRouter()
	if !router.IsRegistered("ch_query") {
		t.Fatal("ch_query should be registered")
	}
	if router.IsRegistered("drop_database") {
		t.Fatal("drop_database should not be registered")
	}
}

func TestToolRouter_RegisteredTools_coversAllRoutes(t *testing.T) {
	t.Parallel()

	router := service.NewToolRouter()
	tools := router.RegisteredTools()
	if len(tools) != 4 {
		t.Fatalf("registered tools count = %d, want 4", len(tools))
	}

	for name, wantCat := range map[string]service.ToolCategory{
		"ch_ping":         service.ToolCategoryClickHouse,
		"ch_query":        service.ToolCategoryClickHouse,
		"list_dwh_models": service.ToolCategoryMetadata,
		"get_dwh_model":   service.ToolCategoryMetadata,
	} {
		got, ok := tools[name]
		if !ok {
			t.Fatalf("missing tool %q in RegisteredTools", name)
		}
		if got != wantCat {
			t.Fatalf("tool %q category = %q, want %q", name, got, wantCat)
		}
	}
}
