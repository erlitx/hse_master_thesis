package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/erlitx/mcp_server/internal/mcp_server/domain"
)

// Создаёт датасет в BI-инструменте
func (uc *UseCase) CreateDataset(ctx context.Context, input domain.CreateDatasetInput) (*domain.Dataset, error) {
	if uc.bitool == nil {
		return nil, fmt.Errorf("bitool adapter is not configured")
	}
	if input.DatabaseID <= 0 {
		return nil, fmt.Errorf("database_id must be greater than zero")
	}
	if strings.TrimSpace(input.Schema) == "" {
		return nil, fmt.Errorf("schema is required")
	}
	if strings.TrimSpace(input.TableName) == "" {
		return nil, fmt.Errorf("table_name is required")
	}

	dataset, err := uc.bitool.CreateDataset(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("usecase - CreateDataset - uc.bitool.CreateDataset: %w", err)
	}

	return dataset, nil
}
