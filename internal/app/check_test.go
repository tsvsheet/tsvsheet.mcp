package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/tsvsheet.mcp/internal/constants"
)

func TestCheckToolClean(t *testing.T) {
	t.Parallel()

	_, out, err := checkTool(context.Background(), nil, CheckInput{Source: "1\t2\n=A1+B1\t3"})
	require.NoError(t, err)
	assert.Equal(t, summaryNone, out.Summary)
	assert.Empty(t, out.Diagnostics)
}

func TestCheckToolDiagnostics(t *testing.T) {
	t.Parallel()

	_, out, err := checkTool(context.Background(), nil, CheckInput{Source: "=NOTAFUNC()"})
	require.NoError(t, err)
	assert.Equal(t, "1 diagnostic(s)", out.Summary)
	require.Len(t, out.Diagnostics, 1)
	assert.Equal(t, "A1", out.Diagnostics[0].Cell)
	assert.Contains(t, out.Diagnostics[0].Message, "NOTAFUNC")
	assert.False(t, out.Diagnostics[0].IsFatal)
}

func TestCheckToolParseError(t *testing.T) {
	t.Parallel()

	result, _, err := checkTool(context.Background(), nil, CheckInput{Source: "=SUM("})
	assert.Nil(t, result)
	require.ErrorIs(t, err, constants.ErrParse)
}
