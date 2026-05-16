package service_test

import (
	"strings"
	"testing"

	"github.com/erlitx/mcp_server/internal/mcp_server/service"
)

func TestValidateReadOnlySQL_allowed(t *testing.T) {
	t.Parallel()

	allowed := []string{
		"SELECT 1",
		"  select id from dds.orders limit 10",
		"SHOW TABLES",
		"DESCRIBE TABLE dds.orders",
		"EXPLAIN SELECT count() FROM dds.orders",
		"WITH cte AS (SELECT 1) SELECT * FROM cte",
	}

	for _, q := range allowed {
		if err := service.ValidateReadOnlySQL(q); err != nil {
			t.Errorf("query %q: unexpected error: %v", q, err)
		}
	}
}

func TestValidateReadOnlySQL_rejectsEmpty(t *testing.T) {
	t.Parallel()

	if err := service.ValidateReadOnlySQL("   "); err == nil {
		t.Fatal("expected error for empty query")
	}
}

func TestValidateReadOnlySQL_rejectsDisallowedPrefix(t *testing.T) {
	t.Parallel()

	blocked := []string{
		"INSERT INTO t VALUES (1)",
		"UPDATE t SET x = 1",
		"DELETE FROM t",
		"CALL some_proc()",
	}

	for _, q := range blocked {
		err := service.ValidateReadOnlySQL(q)
		if err == nil {
			t.Fatalf("query %q: expected error", q)
		}
		if !strings.Contains(err.Error(), "read-only") {
			t.Fatalf("query %q: error = %v, want read-only message", q, err)
		}
	}
}

func TestValidateReadOnlySQL_rejectsForbiddenInsideSelect(t *testing.T) {
	t.Parallel()

	blocked := []string{
		"SELECT 1; DROP TABLE t",
		"SELECT * FROM t WHERE id IN (DELETE FROM u RETURNING id)",
		"SELECT 1 UNION ALL INSERT INTO t SELECT 1",
		"SELECT 1 /* DROP TABLE t */",
		"WITH x AS (SELECT 1) DELETE FROM t",
	}

	for _, q := range blocked {
		if err := service.ValidateReadOnlySQL(q); err == nil {
			t.Fatalf("query %q: expected forbidden SQL error", q)
		}
	}
}
