package dto

// HTTP-запрос создания датасета
type CreateDatasetRequest struct {
	DatabaseID int64  `json:"database_id"`
	Schema     string `json:"schema"`
	TableName  string `json:"table_name"`
	JWTToken   string `json:"jwt_token,omitempty"`
}

// HTTP-ответ создания датасета
type CreateDatasetResponse struct {
	Status  string `json:"status"`
	ID      int64  `json:"id"`
	Message string `json:"message,omitempty"`
}
