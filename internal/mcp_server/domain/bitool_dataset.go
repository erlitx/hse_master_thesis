package domain

// Входные данные создания датасета
type CreateDatasetInput struct {
	DatabaseID int64
	Schema     string
	TableName  string
	JWTToken   string
}

// Созданный датасет
type Dataset struct {
	ID int64
}
