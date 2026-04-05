package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rs/zerolog/log"
)

func (h *Handler) registerManifestResources() {
	ctx := context.Background()

	// Fetch the cached manifest during startup
	manifest, err := h.uc.GetManifestCache(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to get manifest cache during resource registration")
		return
	}

	// Register a bulk resource that lists all models
	bulkRes := mcp.NewResource(
		"dwh://models",
		"All DBT Models",
		mcp.WithResourceDescription("Returns the list of all dbt model names available in the manifest."),
		mcp.WithMIMEType("application/json"),
	)
	h.srv.AddResource(bulkRes, func(req mcp.ReadResourceRequest) ([]interface{}, error) {
		modelNames := make([]string, len(manifest.Models))
		for i, model := range manifest.Models {
			modelNames[i] = model.Name
		}

		jsonData, err := json.MarshalIndent(map[string]interface{}{
			"models": modelNames,
			"count":  len(modelNames),
		}, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal model list: %w", err)
		}

		content := mcp.TextResourceContents{
			ResourceContents: mcp.ResourceContents{
				URI:      "dwh://models",
				MIMEType: "application/json",
			},
			Text: string(jsonData),
		}
		return []interface{}{content}, nil
	})

	// Register individual resources for each model
	for _, model := range manifest.Models {
		// Capture the model for the closure
		modelData := model

		// Debug: log what data we have during registration
		log.Debug().
			Str("model", modelData.Name).
			Int("columns_count", len(modelData.Columns)).
			Int("refs_count", len(modelData.Refs)).
			Str("description", modelData.Description).
			Msg("registering model resource with data")

		uri := fmt.Sprintf("dwh://models/%s", modelData.Name)
		res := mcp.NewResource(
			uri,
			fmt.Sprintf("DBT Model: %s", modelData.Name),
			mcp.WithResourceDescription(fmt.Sprintf("DBT model '%s': %s", modelData.Name, modelData.Description)),
			mcp.WithMIMEType("application/json"),
		)

		h.srv.AddResource(res, func(req mcp.ReadResourceRequest) ([]interface{}, error) {
			// Debug: log column count
			log.Debug().
				Str("model", modelData.Name).
				Int("columns_count", len(modelData.Columns)).
				Int("refs_count", len(modelData.Refs)).
				Msg("marshaling model data for resource")

			jsonData, err := json.MarshalIndent(modelData, "", "  ")
			if err != nil {
				return nil, fmt.Errorf("failed to marshal model data: %w", err)
			}

			// Debug: show first 500 chars of JSON
			preview := string(jsonData)
			if len(preview) > 500 {
				preview = preview[:500] + "..."
			}
			log.Debug().Str("json_preview", preview).Msg("marshaled JSON")

			content := mcp.TextResourceContents{
				ResourceContents: mcp.ResourceContents{
					URI:      uri,
					MIMEType: "application/json",
				},
				Text: string(jsonData),
			}
			return []interface{}{content}, nil
		})

		log.Debug().Str("uri", uri).Str("model", modelData.Name).Msg("registered DBT model resource")
	}

	log.Info().Int("count", len(manifest.Models)).Msg("registered DBT model resources")
}
