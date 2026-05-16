package superset

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/erlitx/mcp_server/internal/mcp_server/domain"
)

// Тело запроса создания датасета в Superset
type createDatasetRequest struct {
	Database int64  `json:"database"`
	Schema   string `json:"schema"`
	Table    string `json:"table_name"`
}

// Результат создания датасета
type createDatasetResult struct {
	ID int64 `json:"id"`
}

// Ответ API Superset на создание датасета
type createDatasetResponse struct {
	Result createDatasetResult `json:"result"`
}

// Создаёт датасет в BI-инструменте
func (a *BiTool) CreateDataset(ctx context.Context, input domain.CreateDatasetInput) (*domain.Dataset, error) {
	if a == nil {
		return nil, fmt.Errorf("superset adapter is nil")
	}
	if input.DatabaseID <= 0 {
		return nil, fmt.Errorf("database_id must be greater than zero")
	}
	if strings.TrimSpace(input.TableName) == "" {
		return nil, fmt.Errorf("table_name is required")
	}
	if strings.TrimSpace(input.Schema) == "" {
		return nil, fmt.Errorf("schema is required")
	}

	token, err := a.accessToken(ctx, input.JWTToken)
	if err != nil {
		return nil, fmt.Errorf("get superset access token: %w", err)
	}

	body, err := json.Marshal(createDatasetRequest{
		Database: input.DatabaseID,
		Schema:   strings.TrimSpace(input.Schema),
		Table:    strings.TrimSpace(input.TableName),
	})
	if err != nil {
		return nil, fmt.Errorf("json.Marshal create dataset request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.cfg.URL+"/api/v1/dataset/", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("http.NewRequestWithContext create dataset: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("superset create dataset request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("superset create dataset failed with status %d", resp.StatusCode)
	}

	var parsed createDatasetResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("decode create dataset response: %w", err)
	}
	if parsed.Result.ID == 0 {
		return nil, fmt.Errorf("superset create dataset returned empty dataset id")
	}

	return &domain.Dataset{ID: parsed.Result.ID}, nil
}
