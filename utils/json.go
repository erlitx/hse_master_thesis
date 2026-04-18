package utils

import (
	"encoding/json"
	"fmt"
)

// PrettyJSON returns an indented JSON string for logs/debugging.
func PrettyJSON(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("<<failed to marshal pretty json: %v>>", err)
	}
	return string(b)
}
