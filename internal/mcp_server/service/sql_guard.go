package service

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	allowedSQLPrefixes = []string{"SELECT", "SHOW", "DESCRIBE", "EXPLAIN"}
	forbiddenSQLTokens = regexp.MustCompile(`(?i)\b(INSERT|UPDATE|DELETE|DROP|ALTER|TRUNCATE|CREATE|GRANT|REVOKE|ATTACH|DETACH|KILL|SYSTEM|OPTIMIZE|RENAME|REPLACE|MERGE)\b`)
)

// ValidateReadOnlySQL проверяет, что запрос допустим в read-only режиме.
func ValidateReadOnlySQL(query string) error {
	q := strings.TrimSpace(query)
	if q == "" {
		return fmt.Errorf("query is empty")
	}

	upper := strings.ToUpper(q)
	if !hasAllowedPrefix(upper) {
		return fmt.Errorf("only read-only queries are allowed (SELECT/SHOW/DESCRIBE/EXPLAIN)")
	}

	if forbiddenSQLTokens.MatchString(q) {
		return fmt.Errorf("query contains forbidden SQL statement")
	}

	return nil
}

func hasAllowedPrefix(upper string) bool {
	for _, prefix := range allowedSQLPrefixes {
		if strings.HasPrefix(upper, prefix) {
			return true
		}
	}
	return false
}
