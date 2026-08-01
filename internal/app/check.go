package app

import (
	"context"
	"strconv"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	tsvsheet "github.com/tsvsheet/go-tsvsheet"

	"github.com/tsvsheet/tsvsheet.mcp/internal/constants"
)

// summaryNone is the summary reported when a sheet has no diagnostics.
const summaryNone = "no diagnostics"

// CheckInput is the tsvsheet_check argument set: the spreadsheet source to
// diagnose.
type CheckInput struct {
	Source string `json:"source" jsonschema:"the .tsvt source to diagnose: a TAB-separated grid of literals and =formulas"`
}

// DiagnosticInfo is one finding. A finding about a cell carries its address;
// a finding about a view directive carries the physical line instead, since a
// directive occupies no row and so has no address. Both are advisory: a sheet
// with diagnostics still computes, and a sheet the parser rejects outright
// comes back as a tool error rather than a finding — which is why there is no
// "fatal" flag here to mislead an agent into branching on one.
type DiagnosticInfo struct {
	Cell    string `json:"cell,omitempty" jsonschema:"the A1 address of the cell the finding is about, when it is about a cell"`
	Message string `json:"message"        jsonschema:"a human-readable description of the finding"`
	Line    int    `json:"line,omitempty" jsonschema:"the 1-based physical line the finding is about, when it is not about a cell"`
}

// CheckOutput is the structured tsvsheet_check result: a summary line and the
// list of diagnostics (empty when the sheet is clean).
type CheckOutput struct {
	Summary     string           `json:"summary"     jsonschema:"one-line summary: 'no diagnostics' or the finding count"`
	Diagnostics []DiagnosticInfo `json:"diagnostics" jsonschema:"the diagnostics found, one per unknown function call"`
}

// checkToolDef describes tsvsheet_check to the agent.
var checkToolDef = &mcp.Tool{
	Name: toolCheck,
	Description: "Parse a .tsvt spreadsheet and report its static diagnostics — each formula cell that " +
		"calls an unknown function — as structured content, or 'no diagnostics' when the sheet is clean.",
}

// checkTool parses the source and reports its diagnostics. A parse failure
// surfaces as constants.ErrParse and becomes MCP tool-error content.
func checkTool(_ context.Context, _ *mcp.CallToolRequest, in CheckInput) (*mcp.CallToolResult, CheckOutput, error) {
	sheet, err := tsvsheet.Parse([]byte(in.Source))
	if err != nil {
		return nil, CheckOutput{}, constants.ErrParse.With(err)
	}
	diagnostics := diagnosticInfos(tsvsheet.Check(sheet))
	return nil, CheckOutput{Summary: summarize(diagnostics), Diagnostics: diagnostics}, nil
}

// diagnosticInfos projects the engine diagnostics into the tool's DTO.
func diagnosticInfos(diags []tsvsheet.Diagnostic) []DiagnosticInfo {
	out := make([]DiagnosticInfo, len(diags))
	for i, d := range diags {
		out[i] = DiagnosticInfo{Cell: d.Cell, Message: d.Message}
	}
	return out
}

// summarize renders the one-line summary for a diagnostic set.
func summarize(diags []DiagnosticInfo) string {
	if len(diags) == 0 {
		return summaryNone
	}
	return strconv.Itoa(len(diags)) + " diagnostic(s)"
}
