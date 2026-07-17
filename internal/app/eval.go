package app

import (
	"context"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	tsvsheet "github.com/tsvsheet/go-tsvsheet"

	"github.com/tsvsheet/tsvsheet.mcp/internal/constants"
)

// Expression is a single spreadsheet formula, with or without a leading =.
type Expression string

// EvalInput is the tsvsheet_eval argument set: a single formula to compute.
type EvalInput struct {
	Expression Expression `json:"expression" jsonschema:"a single formula, with or without a leading = (e.g. SUM(1,2,3))"`
}

// EvalOutput is the structured tsvsheet_eval result: the computed value, which
// may be an error value such as #DIV/0!.
type EvalOutput struct {
	Value string `json:"value" jsonschema:"the computed value (an error value like #DIV/0! is a value, not a failure)"`
}

// evalToolDef describes tsvsheet_eval to the agent.
var evalToolDef = &mcp.Tool{
	Name: toolEval,
	Description: "Compute a single spreadsheet formula (with or without a leading =) as a one-cell sheet " +
		"and return its value — e.g. SUM(1,2,3) yields 6, and 1/0 yields #DIV/0!.",
}

// evalTool trims and computes a single expression. An empty expression is
// constants.ErrEmptyExpression; an expression carrying a TAB or newline (the
// grid's own separators) is constants.ErrMultiCellExpression rather than a
// silently truncated multi-cell parse; a malformed formula is constants.ErrParse
// — all become MCP tool-error content.
func evalTool(_ context.Context, _ *mcp.CallToolRequest, in EvalInput) (*mcp.CallToolResult, EvalOutput, error) {
	expression := normalizeExpression(in.Expression)
	if expression == "" {
		return nil, EvalOutput{}, constants.ErrEmptyExpression
	}
	if strings.ContainsAny(string(expression), "\t\n") {
		return nil, EvalOutput{}, constants.ErrMultiCellExpression
	}
	sheet, err := tsvsheet.Parse([]byte("=" + string(expression)))
	if err != nil {
		return nil, EvalOutput{}, constants.ErrParse.With(err)
	}
	value := cellValue(sheet.Compute())
	result := &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: value}}}
	return result, EvalOutput{Value: value}, nil
}

// normalizeExpression strips surrounding whitespace and a single leading = so a
// caller may pass either "=A1*2" or "A1*2".
func normalizeExpression(raw Expression) Expression {
	trimmed := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(string(raw)), "="))
	return Expression(trimmed)
}

// cellValue reads the top-left value of a computed grid, returning "" when the
// grid is empty (a defensive contract; a one-cell sheet always fills it).
func cellValue(grid tsvsheet.Grid) string {
	if len(grid) == 0 || len(grid[0]) == 0 {
		return ""
	}
	return grid[0][0]
}
