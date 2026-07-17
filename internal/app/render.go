package app

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	tsvsheet "github.com/tsvsheet/go-tsvsheet"

	"github.com/tsvsheet/tsvsheet.mcp/internal/constants"
)

// RenderInput is the tsvsheet_render argument set: the spreadsheet source and
// the desired serialization of its computed grid.
type RenderInput struct {
	Source string `json:"source"           jsonschema:"the .tsvt source: a TAB-separated grid of literals and =formulas"`
	Format string `json:"format,omitempty" jsonschema:"grid serialization: tsv (default), csv, html, or markdown"`
}

// RenderOutput is the structured tsvsheet_render result: the computed grid in
// the requested format, plus the format that was applied.
type RenderOutput struct {
	Format   string `json:"format"   jsonschema:"the format the grid was rendered in"`
	Rendered string `json:"rendered" jsonschema:"the computed grid serialized in the requested format"`
}

// renderToolDef describes tsvsheet_render to the agent.
var renderToolDef = &mcp.Tool{
	Name: toolRender,
	Description: "Parse a .tsvt spreadsheet and compute every formula in place, returning the " +
		"resulting value grid serialized as tsv (the default), csv, html, or markdown.",
}

// renderTool parses the source, computes the grid, and serializes it. A parse
// failure surfaces as constants.ErrParse; an unrecognized format as
// constants.ErrUnknownFormat — both become MCP tool-error content.
func renderTool(_ context.Context, _ *mcp.CallToolRequest, in RenderInput) (*mcp.CallToolResult, RenderOutput, error) {
	sheet, err := tsvsheet.Parse([]byte(in.Source))
	if err != nil {
		return nil, RenderOutput{}, constants.ErrParse.With(err)
	}
	format := in.Format
	if format == "" {
		format = string(formatTSV)
	}
	rendered, err := render(sheet.Compute(), Format(format))
	if err != nil {
		return nil, RenderOutput{}, err
	}
	result := &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: rendered}}}
	return result, RenderOutput{Format: format, Rendered: rendered}, nil
}
