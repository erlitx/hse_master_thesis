package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/erlitx/mcp_server/internal/mcp_server/domain"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rs/zerolog/log"
)

// Регистрирует ресурсы manifest DWH
func (h *Handler) registerManifestResources() {
	ctx := context.Background()

	manifest, err := h.uc.GetManifestCache(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to get manifest cache during resource registration")
		return
	}

	// Root/bulk resource (optional convenience endpoint)
	res := mcp.NewResource(
		"dwh://models",
		"All DBT Models",
		mcp.WithResourceDescription("Returns all DBT model names"),
		mcp.WithMIMEType("application/json"),
	)

	h.srv.AddResource(res, h.makeBulkModelsHandler(manifest))

	h.registerModelResources(manifest)

	log.Info().Int("count", len(manifest.Models)).Msg("registered DBT model resources")
}

// Регистрирует ресурс на каждую модель
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

// Обработчик чтения dwh://models
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

// Обработчик чтения dwh://models/{name}
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

// Сериализует payload в TextResourceContents
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

// Извлекает имена моделей из среза
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
