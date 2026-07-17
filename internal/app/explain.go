package app

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	tsvsheet "github.com/tsvsheet/go-tsvsheet"

	"github.com/tsvsheet/tsvsheet.mcp/internal/constants"
)

// ExplainInput is the tsvsheet_explain argument set: the spreadsheet source and
// the A1 address of the cell to explain.
type ExplainInput struct {
	Source string `json:"source" jsonschema:"the .tsvt source: a TAB-separated grid of literals and =formulas"`
	Cell   string `json:"cell"   jsonschema:"the A1 address of the cell to explain (e.g. D5)"`
}

// TraceInputInfo is one reference a formula reads, with its resolved value.
type TraceInputInfo struct {
	Ref   string `json:"ref"   jsonschema:"the A1 reference the formula reads"`
	Value string `json:"value" jsonschema:"the resolved value of that reference"`
}

// ExplainOutput is the structured tsvsheet_explain result: the cell's computed
// value, its formula (empty for a literal), and each reference it reads.
type ExplainOutput struct {
	Cell    string           `json:"cell"              jsonschema:"the A1 address of the explained cell"`
	Value   string           `json:"value"             jsonschema:"the cell's computed value"`
	Formula string           `json:"formula,omitempty" jsonschema:"the cell's formula, empty when the cell is a literal"`
	Inputs  []TraceInputInfo `json:"inputs,omitempty"  jsonschema:"the references the formula reads, with resolved values"`
}

// explainToolDef describes tsvsheet_explain to the agent.
var explainToolDef = &mcp.Tool{
	Name: toolExplain,
	Description: "Parse and compute a .tsvt spreadsheet, then explain one cell: its computed value, its " +
		"formula (if any), and the resolved value of every cell that formula reads.",
}

// explainTool parses the source, resolves the cell address, and traces it. A
// parse failure is constants.ErrParse; a malformed address is
// constants.ErrInvalidCell; a cell outside the grid is constants.ErrUnknownCell
// — each becomes MCP tool-error content.
func explainTool(
	_ context.Context,
	_ *mcp.CallToolRequest,
	in ExplainInput,
) (*mcp.CallToolResult, ExplainOutput, error) {
	sheet, err := tsvsheet.Parse([]byte(in.Source))
	if err != nil {
		return nil, ExplainOutput{}, constants.ErrParse.With(err)
	}
	address, err := tsvsheet.ParseAddress(tsvsheet.AddressText(in.Cell))
	if err != nil {
		return nil, ExplainOutput{}, constants.ErrInvalidCell.With(err, "cell", in.Cell)
	}
	trace, err := tsvsheet.Explain(sheet, address)
	if err != nil {
		return nil, ExplainOutput{}, constants.ErrUnknownCell.With(err, "cell", in.Cell)
	}
	return nil, explainOutput(trace), nil
}

// explainOutput projects an engine Trace into the tool's DTO.
func explainOutput(trace tsvsheet.Trace) ExplainOutput {
	return ExplainOutput{
		Cell:    trace.Cell,
		Value:   trace.Value,
		Formula: trace.Formula,
		Inputs:  traceInputInfos(trace.Inputs),
	}
}

// traceInputInfos projects the engine trace inputs into the tool's DTO,
// preserving a nil slice so an input-free trace omits the field.
func traceInputInfos(inputs []tsvsheet.TraceInput) []TraceInputInfo {
	if len(inputs) == 0 {
		return nil
	}
	out := make([]TraceInputInfo, len(inputs))
	for i, in := range inputs {
		out[i] = TraceInputInfo{Ref: in.Ref, Value: in.Value}
	}
	return out
}
