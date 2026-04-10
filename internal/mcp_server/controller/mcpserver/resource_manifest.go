package mcpserver

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rs/zerolog/log"
)

func (h *Handler) DWHResources(req mcp.ReadResourceRequest) ([]interface{}, error) {
	_ = req

	ctx := context.Background()

	//manifest, err := h.uc.GetDBTManifest(ctx)
	manifest, err := h.uc.GetManifestCache(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to get DBT manifest")
		panic(err)
	}

	b, err := json.Marshal(manifest)
	if err != nil {
		return nil, err
	}

	content := mcp.TextResourceContents{
		ResourceContents: mcp.ResourceContents{
			URI:      "dwh://models",
			MIMEType: "application/json",
		},
		Text: string(b),
	}

	return []interface{}{content}, nil
}
