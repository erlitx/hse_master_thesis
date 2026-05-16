package dbt

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/erlitx/mcp_server/internal/mcp_server/domain"
	"github.com/erlitx/mcp_server/internal/mcp_server/dto"
)

// Парсит manifest DBT из файла
func (mp *DbtParser) ParseManifest(ctx context.Context) (*domain.DBTManifest, error) {

	// TODO: start dbt parser in background and return cached manifest until it's ready, then switch to new one
	data, err := os.ReadFile(mp.manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file: %w", err)
	}

	var manifest dto.ManifestJSON
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal manifest JSON: %w", err)
	}

	result := &domain.DBTManifest{
		Models: make([]domain.DBTModel, 0),
	}

	for _, node := range manifest.Nodes {
		// Only process models (skip sources, tests, etc.)
		if node.ResourceType != "model" {
			continue
		}

		if node.Meta == nil {
			continue
		}

		level, ok := node.Meta["level"].(string)
		if !ok || level != "bi_data_mart" {
			continue
		}

		// Convert columns from map to slice
		columns := make([]domain.DBTModelColumn, 0, len(node.Columns))
		for _, col := range node.Columns {
			columns = append(columns, domain.DBTModelColumn{
				Name:        col.Name,
				Description: col.Description,
				DataType:    col.DataType,
			})
		}

		// Convert refs
		refs := make([]domain.DBTModelRef, 0, len(node.Refs))
		for _, ref := range node.Refs {
			refs = append(refs, domain.DBTModelRef{
				Name:    ref.Name,
				Package: ref.Package,
				Version: ref.Version,
			})
		}

		model := domain.DBTModel{
			Name:         node.Name,
			RelationName: node.RelationName,
			Description:  node.Description,
			Meta:         node.Meta,
			Columns:      columns,
			Refs:         refs,
		}

		result.Models = append(result.Models, model)
	}

	return result, nil
}
