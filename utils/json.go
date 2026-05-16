package utils

import (
	"encoding/json"
	"fmt"
)

// Форматирует значение в JSON для логов
func PrettyJSON(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("<<failed to marshal pretty json: %v>>", err)
	}
	return string(b)
}
