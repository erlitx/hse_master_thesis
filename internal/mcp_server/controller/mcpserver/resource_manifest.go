package mcpserver

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rs/zerolog/log"
)

func (h *Handler) DWHResources(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	_ = req

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
		URI:      "dwh://models",
		MIMEType: "application/json",
		Text:     string(b),
	}

	return []mcp.ResourceContents{&content}, nil
}
