package mcpserver

import (
	"context"
	"fmt"
	"strings"

	"github.com/erlitx/mcp_server/internal/mcp_server/domain"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rs/zerolog/log"
)

// registerManifestTools initializes MCP TOOLS related to DBT manifest.
//
// IMPORTANT:
// - This runs once at startup
// - It uses cached manifest snapshot (NOT dynamic)
// - Registers:
//  1. list_dbt_models     -> returns all model names
//  2. get_dbt_model       -> returns full JSON for one model by name
//
// WHY TOOLS INSTEAD OF RESOURCES:
// - Claude Message API supports tools more directly
// - tools/call can return structured JSON/text payloads
func (h *Handler) registerManifestTools() {
	ctx := context.Background()

	manifest, err := h.uc.GetManifestCache(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to get manifest cache during tool registration")
		return
	}

	h.modelByName = buildModelIndex(manifest)

	h.registerListModelsTool(manifest)
	h.registerGetModelTool()

	log.Info().Int("count", len(manifest.Models)).Msg("registered DBT manifest tools")
}

func buildModelIndex(manifest *domain.DBTManifest) map[string]domain.DBTModel {
	modelByName := make(map[string]domain.DBTModel, len(manifest.Models))
	for _, model := range manifest.Models {
		modelByName[model.Name] = model
	}
	return modelByName
}

// registerListModelsTool registers a tool:
//
//	list_dbt_models()
//
// Returns:
//
//	{
//	  "models": ["model1", "model2"],
//	  "count": N
//	}
func (h *Handler) registerListModelsTool(manifest *domain.DBTManifest) {
	tool := mcp.NewTool(
		"list_dwh_models",
		mcp.WithDescription("Return all DWH models (tables) names and business descriptions"),
	)

	h.srv.AddTool(tool, h.makeListModelsToolHandler(manifest))

	log.Debug().
		Str("tool", "list_dwh_models").
		Msg("registered DBT list tool")
}

// makeListModelsToolHandler handles:
//
//	tools/call -> list_dbt_models
func (h *Handler) makeListModelsToolHandler(manifest *domain.DBTManifest) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		models := make([]domain.ModelListItem, 0, len(manifest.Models))

		for _, model := range manifest.Models {
			models = append(models, domain.ModelListItem{
				Name:        model.Name,
				Description: strings.TrimSpace(model.Description),
			})
		}

		payload := map[string]interface{}{
			"models": models,
			"count":  len(models),
		}

		return &mcp.CallToolResult{
			StructuredContent: payload,
		}, nil

		//return mcp.NewToolResultJSON(payload)
	}
}

// registerGetModelTool registers a tool:
//
//	get_dbt_model(model_name: string)
//
// Returns FULL model JSON for one DBT model.
func (h *Handler) registerGetModelTool() {
	tool := mcp.NewTool(
		"get_dwh_model",
		mcp.WithDescription("Return full DWH model metadata by model name from the cached manifest"),
		mcp.WithString(
			"model_name",
			mcp.Required(),
			mcp.Description("Name of the DWH model"),
		),
	)

	h.srv.AddTool(tool, h.handleGetModelTool)

	log.Debug().
		Str("tool", "get_dbt_model").
		Msg("registered DBT get-model tool")
}

// makeGetModelToolHandler handles:
//
//	tools/call -> get_dbt_model { "model_name": "..." }
func (h *Handler) handleGetModelTool(
	ctx context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	modelName, err := req.RequireString("model_name")
	if err != nil {
		return mcp.NewToolResultError(
			fmt.Sprintf("missing required argument 'model_name': %v", err),
		), nil
	}

	model, ok := h.modelByName[modelName]
	if !ok {
		return mcp.NewToolResultError(
			fmt.Sprintf("model '%s' not found", modelName),
		), nil
	}

	log.Debug().
		Str("model", model.Name).
		Int("columns", len(model.Columns)).
		Int("refs", len(model.Refs)).
		Msg("serving DBT model tool response")

	return &mcp.CallToolResult{
			StructuredContent: model,
		}, nil

	//return mcp.NewToolResultJSON(model)
}
