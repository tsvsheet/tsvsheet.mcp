package app

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/tsvsheet.mcp/internal/constants"
)

// textOf returns the concatenated text of a result's content blocks.
func textOf(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	require.NotNil(t, result)
	var text string
	for _, c := range result.Content {
		tc, ok := c.(*mcp.TextContent)
		require.True(t, ok)
		text += tc.Text
	}
	return text
}

func TestRenderTool(t *testing.T) {
	t.Parallel()

	result, out, err := renderTool(context.Background(), nil, RenderInput{Source: "1\t2\n=A1+B1\t=A1*B1"})
	require.NoError(t, err)
	assert.Equal(t, "tsv", out.Format)
	assert.Equal(t, "1\t2\n3\t2\n", out.Rendered)
	assert.Equal(t, out.Rendered, textOf(t, result))
}

func TestRenderToolExplicitFormat(t *testing.T) {
	t.Parallel()

	result, out, err := renderTool(context.Background(), nil, RenderInput{Source: "1\t2", Format: "csv"})
	require.NoError(t, err)
	assert.Equal(t, "csv", out.Format)
	assert.Equal(t, "1,2\n", out.Rendered)
	assert.Equal(t, "1,2\n", textOf(t, result))
}

func TestRenderToolParseError(t *testing.T) {
	t.Parallel()

	result, _, err := renderTool(context.Background(), nil, RenderInput{Source: "=SUM("})
	assert.Nil(t, result)
	require.ErrorIs(t, err, constants.ErrParse)
}

func TestRenderToolUnknownFormat(t *testing.T) {
	t.Parallel()

	result, _, err := renderTool(context.Background(), nil, RenderInput{Source: "1", Format: "yaml"})
	assert.Nil(t, result)
	require.ErrorIs(t, err, constants.ErrUnknownFormat)
}

// TestRenderToolNamedValues pins the 023 seam at the MCP boundary: an agent
// rendering a sheet that declares names with the |@ meta clause gets the
// computed values — the tool cannot silently lag the engine's grammar.
func TestRenderToolNamedValues(t *testing.T) {
	t.Parallel()

	_, out, err := renderTool(context.Background(), nil,
		RenderInput{Source: "=0.08 |@ named(Rate)\t=@Rate * 100"})
	require.NoError(t, err)
	assert.Equal(t, "0.08\t8\n", out.Rendered)
}
