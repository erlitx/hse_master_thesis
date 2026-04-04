package dbt

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/erlitx/mcp_server/internal/domain"
)

// ManifestParser handles parsing of DBT manifest.json files
type ManifestParser struct {
	manifestPath string
}

// New creates a new ManifestParser with the specified manifest file path
func New(manifestPath string) *ManifestParser {
	return &ManifestParser{
		manifestPath: manifestPath,
	}
}

// manifestJSON represents the structure of the manifest.json file
type manifestJSON struct {
	Nodes map[string]nodeJSON `json:"nodes"`
}

// nodeJSON represents a node in the manifest
type nodeJSON struct {
	Name         string                 `json:"name"`
	RelationName string                 `json:"relation_name"`
	ResourceType string                 `json:"resource_type"`
	Description  string                 `json:"description"`
	Meta         map[string]any         `json:"meta"`
	Columns      map[string]columnJSON  `json:"columns"`
	Refs         []refJSON              `json:"refs"`
}

// columnJSON represents a column definition
type columnJSON struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	DataType    string `json:"data_type"`
}

// refJSON represents a model reference
type refJSON struct {
	Name    string  `json:"name"`
	Package *string `json:"package"`
	Version *string `json:"version"`
}

// ParseManifest reads and parses the manifest.json file
func (mp *ManifestParser) ParseManifest(ctx context.Context) (*domain.DBTManifest, error) {
	// Read the manifest file
	data, err := os.ReadFile(mp.manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file: %w", err)
	}

	// Parse JSON
	var manifest manifestJSON
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal manifest JSON: %w", err)
	}

	// Convert to domain objects
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
