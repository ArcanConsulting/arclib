package db

import (
	"fmt"
	"strings"
)

// ConvertPlaceholders rewrites PostgreSQL-style placeholders ($1, $2, ...)
// to positional ? placeholders used by SQLite and MySQL.
func ConvertPlaceholders(query string) string {
	var b strings.Builder
	b.Grow(len(query))

	i := 0
	for i < len(query) {
		if query[i] == '$' && i+1 < len(query) && query[i+1] >= '1' && query[i+1] <= '9' {
			b.WriteByte('?')
			i++ // skip '$'
			// skip all digits
			for i < len(query) && query[i] >= '0' && query[i] <= '9' {
				i++
			}
		} else {
			b.WriteByte(query[i])
			i++
		}
	}
	return b.String()
}

// Placeholders generates a comma-separated list of n placeholder parameters.
// For PostgreSQL: "$1, $2, $3"
// For SQLite/MySQL: "?, ?, ?"
func Placeholders(n int, postgres bool) string {
	if n <= 0 {
		return ""
	}

	parts := make([]string, n)
	for i := range n {
		if postgres {
			parts[i] = fmt.Sprintf("$%d", i+1)
		} else {
			parts[i] = "?"
		}
	}
	return strings.Join(parts, ", ")
}

// SetClause generates a SET clause for UPDATE statements.
// For PostgreSQL: "col1 = $1, col2 = $2" (starting at offset)
// For SQLite/MySQL: "col1 = ?, col2 = ?"
func SetClause(columns []string, postgres bool, offset int) string {
	if len(columns) == 0 {
		return ""
	}

	parts := make([]string, len(columns))
	for i, col := range columns {
		if postgres {
			parts[i] = fmt.Sprintf("%s = $%d", col, offset+i+1)
		} else {
			parts[i] = fmt.Sprintf("%s = ?", col)
		}
	}
	return strings.Join(parts, ", ")
}

// QueryBuilder helps construct SQL queries with conditional clauses.
type QueryBuilder struct {
	base       string
	conditions []string
	orderBy    string
	limit      int
	offset     int
	postgres   bool
}

// NewQueryBuilder creates a new QueryBuilder with a base SELECT statement.
func NewQueryBuilder(base string, postgres bool) *QueryBuilder {
	return &QueryBuilder{
		base:     base,
		postgres: postgres,
	}
}

// Where adds a WHERE condition. Multiple calls are joined with AND.
func (qb *QueryBuilder) Where(condition string) *QueryBuilder {
	qb.conditions = append(qb.conditions, condition)
	return qb
}

// OrderBy sets the ORDER BY clause.
func (qb *QueryBuilder) OrderBy(clause string) *QueryBuilder {
	qb.orderBy = clause
	return qb
}

// Limit sets the LIMIT value.
func (qb *QueryBuilder) Limit(n int) *QueryBuilder {
	qb.limit = n
	return qb
}

// Offset sets the OFFSET value.
func (qb *QueryBuilder) Offset(n int) *QueryBuilder {
	qb.offset = n
	return qb
}

// Build returns the final SQL query string.
func (qb *QueryBuilder) Build() string {
	var b strings.Builder
	b.WriteString(qb.base)

	if len(qb.conditions) > 0 {
		b.WriteString(" WHERE ")
		b.WriteString(strings.Join(qb.conditions, " AND "))
	}

	if qb.orderBy != "" {
		b.WriteString(" ORDER BY ")
		b.WriteString(qb.orderBy)
	}

	if qb.limit > 0 {
		fmt.Fprintf(&b, " LIMIT %d", qb.limit)
	}

	if qb.offset > 0 {
		fmt.Fprintf(&b, " OFFSET %d", qb.offset)
	}

	return b.String()
}
