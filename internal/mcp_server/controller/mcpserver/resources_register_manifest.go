package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rs/zerolog/log"
	"github.com/erlitx/mcp_server/internal/mcp_server/domain"
)


// registerManifestResources initializes ALL MCP resources related to DBT manifest.
//
// IMPORTANT:
// - This runs once at startup
// - It uses cached manifest snapshot (NOT dynamic)
// - Registers:
//   1. Bulk resource: dwh://models
//   2. Per-model resources: dwh://models/{model_name}
//
// MCP DESIGN NOTE:
// - resources/list → returns metadata (name, description, uri)
// - resources/read → returns actual content (via handler below)
func (h *Handler) registerManifestResources() {
	ctx := context.Background()

	// Load manifest snapshot from usecase layer
	manifest, err := h.uc.GetManifestCache(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to get manifest cache during resource registration")
		return
	}

	// Root/bulk resource (optional convenience endpoint)
	// Allows clients to fetch structured list of models via resources/read
	res := mcp.NewResource(
		"dwh://models",
		"All DBT Models",
		mcp.WithResourceDescription("Returns all DBT model names"),
		mcp.WithMIMEType("application/json"),
	)

	// Attach handler (executed on resources/read)
	h.srv.AddResource(res, h.makeBulkModelsHandler(manifest))

	// Register individual model resources (core functionality)
	h.registerModelResources(manifest)

	log.Info().Int("count", len(manifest.Models)).Msg("registered DBT model resources")
}


// registerModelResources creates ONE MCP resource per DBT model.
//
// Each model becomes accessible via:
//   dwh://models/{model_name}
//
// Example:
//   dwh://models/locations
//   dwh://models/net_work_minutes
//
// NOTE:
// - modelData is copied to avoid closure bug in Go loops
// - each handler is bound to its own model snapshot
func (h *Handler) registerModelResources(manifest *domain.DBTManifest) {
	for _, model := range manifest.Models {
		modelData := model // IMPORTANT: capture loop variable

		uri := fmt.Sprintf("dwh://models/%s", modelData.Name)

		res := mcp.NewResource(
			uri,
			fmt.Sprintf("DBT Model: %s", modelData.Name),
			mcp.WithResourceDescription(modelData.Description),
			mcp.WithMIMEType("application/json"),
		)

		// Attach per-model handler
		h.srv.AddResource(res, h.makeModelHandler(uri, modelData))

		log.Debug().
			Str("model", modelData.Name).
			Int("columns", len(modelData.Columns)).
			Int("refs", len(modelData.Refs)).
			Msg("registered model resource")
	}
}


// makeBulkModelsHandler returns handler for:
//
//   resources/read → dwh://models
//
// Returns:
// {
//   "models": ["model1", "model2"],
//   "count": N
// }
//
// NOTE:
// - This is actual DATA (not metadata like resources/list)
// - Useful for clients / LLMs to consume structured list
func (h *Handler) makeBulkModelsHandler(manifest *domain.DBTManifest) func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	return func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {

		// Extract only model names (lightweight response)
		modelNames := extractModelNames(manifest.Models)

		payload := map[string]interface{}{
			"models": modelNames,
			"count":  len(modelNames),
		}

		// Build MCP-compatible text resource
		return buildJSONResource("dwh://models", payload)
	}
}


// makeModelHandler returns handler for:
//
//   resources/read → dwh://models/{model}
//
// Returns FULL model JSON (columns, refs, meta, etc)
//
// NOTE:
// - This is the main "data endpoint" for each DBT model
// - Uses TextResourceContents → JSON is embedded as string (MCP spec)
// - Client must parse JSON from "text" field
func (h *Handler) makeModelHandler(uri string, model domain.DBTModel) func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	return func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {

		log.Debug().
			Str("model", model.Name).
			Int("columns", len(model.Columns)).
			Int("refs", len(model.Refs)).
			Msg("serving model resource")

		return buildJSONResource(uri, model)
	}
}


// buildJSONResource converts ANY payload into MCP TextResourceContents.
//
// IMPORTANT MCP BEHAVIOR:
// - "text" field is ALWAYS string → JSON gets escaped (\n, \")
// - This is expected and correct (JSON inside JSON)
//
// Client must:
//   1. read "text"
//   2. parse JSON again
func buildJSONResource(uri string, payload interface{}) ([]mcp.ResourceContents, error) {
	jsonData, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal failed: %w", err)
	}

	content := mcp.TextResourceContents{
		URI:      uri,
		MIMEType: "application/json",
		Text:     string(jsonData),
	}

	return []mcp.ResourceContents{&content}, nil
}


// extractModelNames is a small helper to avoid repeating loops.
//
// Converts:
//   []DBTModel → []string
//
// NOTE:
// - used only for bulk resource
// - keeps payload small and clean
func extractModelNames(models []domain.DBTModel) []string {
	names := make([]string, len(models))
	for i, m := range models {
		names[i] = m.Name
	}
	return names
}



// package mcpserver

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"

// 	"github.com/mark3labs/mcp-go/mcp"
// 	"github.com/rs/zerolog/log"
// )

// func (h *Handler) registerManifestResources() {
// 	ctx := context.Background()

// 	// Fetch the cached manifest during startup
// 	manifest, err := h.uc.GetManifestCache(ctx)
// 	//log.Debug().Msgf("CACHE MANIFEST: %v", manifest)
// 	if err != nil {
// 		log.Error().Err(err).Msg("failed to get manifest cache during resource registration")
// 		return
// 	}

// 	// Register a bulk resource that lists all models
// 	bulkRes := mcp.NewResource(
// 		"dwh://models",
// 		"All DBT Models",
// 		mcp.WithResourceDescription("Returns the list of all dbt model names available in the manifest."),
// 		mcp.WithMIMEType("application/json"),
// 	)
// 	h.srv.AddResource(bulkRes, func(req mcp.ReadResourceRequest) ([]interface{}, error) {
// 		modelNames := make([]string, len(manifest.Models))
// 		for i, model := range manifest.Models {
// 			modelNames[i] = model.Name
// 		}

// 		jsonData, err := json.MarshalIndent(map[string]interface{}{
// 			"models": modelNames,
// 			"count":  len(modelNames),
// 		}, "", "  ")
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to marshal model list: %w", err)
// 		}

// 		content := mcp.TextResourceContents{
// 			ResourceContents: mcp.ResourceContents{
// 				URI:      "dwh://models",
// 				MIMEType: "application/json",
// 			},
// 			Text: string(jsonData),
// 		}
// 		return []interface{}{content}, nil
// 	})

// 	// Register individual resources for each model
// 	for _, model := range manifest.Models {
// 		// Capture the model for the closure
// 		modelData := model

// 		// Debug: log what data we have during registration
// 		log.Debug().
// 			Str("model", modelData.Name).
// 			Int("columns_count", len(modelData.Columns)).
// 			Int("refs_count", len(modelData.Refs)).
// 			Str("description", modelData.Description).
// 			Msg("registering model resource with data")

// 		uri := fmt.Sprintf("dwh://models/%s", modelData.Name)
// 		res := mcp.NewResource(
// 			uri,
// 			fmt.Sprintf("DBT Model: %s", modelData.Name),
// 			mcp.WithResourceDescription(fmt.Sprintf("DBT model '%s': %s", modelData.Name, modelData.Description)),
// 			mcp.WithMIMEType("application/json"),
// 		)

// 		h.srv.AddResource(res, func(req mcp.ReadResourceRequest) ([]interface{}, error) {
// 			// Debug: log column count
// 			log.Debug().
// 				Str("model", modelData.Name).
// 				Int("columns_count", len(modelData.Columns)).
// 				Int("refs_count", len(modelData.Refs)).
// 				Msg("marshaling model data for resource")

// 			jsonData, err := json.MarshalIndent(modelData, "", "  ")
// 			if err != nil {
// 				return nil, fmt.Errorf("failed to marshal model data: %w", err)
// 			}

// 			// Debug: show first 500 chars of JSON
// 			preview := string(jsonData)
// 			if len(preview) > 500 {
// 				preview = preview[:500] + "..."
// 			}
// 			log.Debug().Str("json_preview", preview).Msg("marshaled JSON")

// 			content := mcp.TextResourceContents{
// 				ResourceContents: mcp.ResourceContents{
// 					URI:      uri,
// 					MIMEType: "application/json",
// 				},
// 				Text: string(jsonData),
// 			}
// 			return []interface{}{content}, nil
// 		})

// 		log.Debug().Str("uri", uri).Str("model", modelData.Name).Msg("registered DBT model resource")
// 	}

// 	log.Info().Int("count", len(manifest.Models)).Msg("registered DBT model resources")
// }
