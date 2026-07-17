package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/tsvsheet.mcp/internal/constants"
)

func TestEvalTool(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		expression Expression
		want       string
	}{
		{name: "sum", expression: "SUM(1,2,3)", want: "6"},
		{name: "leading equals", expression: "=2^10", want: "1024"},
		{name: "surrounding whitespace", expression: "  = 4 * 5 ", want: "20"},
		{name: "division by zero is a value", expression: "1/0", want: "#DIV/0!"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, out, err := evalTool(context.Background(), nil, EvalInput{Expression: tt.expression})
			require.NoError(t, err)
			assert.Equal(t, tt.want, out.Value)
			assert.Equal(t, tt.want, textOf(t, result))
		})
	}
}

func TestEvalToolEmpty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		expression Expression
	}{
		{name: "blank", expression: ""},
		{name: "whitespace", expression: "   "},
		{name: "bare equals", expression: "="},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, _, err := evalTool(context.Background(), nil, EvalInput{Expression: tt.expression})
			assert.Nil(t, result)
			require.ErrorIs(t, err, constants.ErrEmptyExpression)
		})
	}
}

func TestEvalToolParseError(t *testing.T) {
	t.Parallel()

	result, _, err := evalTool(context.Background(), nil, EvalInput{Expression: "SUM("})
	assert.Nil(t, result)
	require.ErrorIs(t, err, constants.ErrParse)
}

func TestEvalToolRejectsMultiCell(t *testing.T) {
	t.Parallel()

	// A TAB or newline is the grid's own field/row separator: wrapped into a
	// sheet it would split into multiple cells and silently evaluate only the
	// first. The tool must reject it, not return a truncated result.
	tests := []struct {
		name       string
		expression Expression
	}{
		{name: "embedded tab", expression: "1+1\tSUM(9,9)"},
		{name: "embedded newline", expression: "1+1\n=99"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, _, err := evalTool(context.Background(), nil, EvalInput{Expression: tt.expression})
			assert.Nil(t, result)
			require.ErrorIs(t, err, constants.ErrMultiCellExpression)
		})
	}
}
