package assert

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Expectation interface {
	Check(*testing.T, []map[string]any)
}

type RowsExpectation struct {
	Rows []map[string]any
}

func (e RowsExpectation) Check(t *testing.T, result []map[string]any) {
	assert.Len(t, result, len(e.Rows))
	assert.Equal(t, e.Rows, result)
}

type CountExpectation struct {
	Count int
}

func (e CountExpectation) Check(t *testing.T, result []map[string]any) {
	assert.Len(t, result, e.Count)
}

func RowsExist(t *testing.T, p *pgxpool.Pool, q string, e []Expectation) {
	ctx := context.Background()

	rows, err := p.Query(ctx, q)
	require.Nil(t, err, "query failed")

	defer rows.Close()

	var result []map[string]any

	for rows.Next() {
		values, err := rows.Values()
		require.Nil(t, err)

		fieldDescriptions := rows.FieldDescriptions()
		rowMap := make(map[string]any, len(values))

		for i, fd := range fieldDescriptions {
			colName := string(fd.Name)
			rowMap[colName] = values[i]
		}

		result = append(result, rowMap)
	}

	if !assert.NoError(t, rows.Err()) {
		return
	}

	for _, ex := range e {
		ex.Check(t, result)
	}
}
