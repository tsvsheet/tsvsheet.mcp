package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/tsvsheet.mcp/internal/constants"
)

func TestExplainToolFormula(t *testing.T) {
	t.Parallel()

	_, out, err := explainTool(context.Background(), nil, ExplainInput{Source: "5\t=A1*2", Cell: "B1"})
	require.NoError(t, err)
	assert.Equal(t, "B1", out.Cell)
	assert.Equal(t, "10", out.Value)
	assert.Equal(t, "A1 * 2", out.Formula)
	require.Len(t, out.Inputs, 1)
	assert.Equal(t, "A1", out.Inputs[0].Ref)
	assert.Equal(t, "5", out.Inputs[0].Value)
}

func TestExplainToolLiteral(t *testing.T) {
	t.Parallel()

	_, out, err := explainTool(context.Background(), nil, ExplainInput{Source: "5", Cell: "A1"})
	require.NoError(t, err)
	assert.Equal(t, "A1", out.Cell)
	assert.Equal(t, "5", out.Value)
	assert.Empty(t, out.Formula)
	assert.Nil(t, out.Inputs)
}

func TestExplainToolParseError(t *testing.T) {
	t.Parallel()

	result, _, err := explainTool(context.Background(), nil, ExplainInput{Source: "=SUM(", Cell: "A1"})
	assert.Nil(t, result)
	require.ErrorIs(t, err, constants.ErrParse)
}

func TestExplainToolInvalidCell(t *testing.T) {
	t.Parallel()

	result, _, err := explainTool(context.Background(), nil, ExplainInput{Source: "5", Cell: "not-a-cell"})
	assert.Nil(t, result)
	require.ErrorIs(t, err, constants.ErrInvalidCell)
	assert.Contains(t, err.Error(), "not-a-cell")
}

func TestExplainToolUnknownCell(t *testing.T) {
	t.Parallel()

	result, _, err := explainTool(context.Background(), nil, ExplainInput{Source: "5", Cell: "Z99"})
	assert.Nil(t, result)
	require.ErrorIs(t, err, constants.ErrUnknownCell)
	assert.Contains(t, err.Error(), "Z99")
}
