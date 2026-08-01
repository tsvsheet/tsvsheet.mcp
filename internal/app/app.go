// Package app builds the tsvsheet MCP server: it registers the four tsvsheet
// tools (render, check, explain, eval) on a github.com/modelcontextprotocol
// server and serves them over an injected transport. The tool handlers wrap the
// github.com/tsvsheet/go-tsvsheet engine — the single source of truth for
// spreadsheet semantics — and translate engine failures into MCP tool-error
// content so an agent sees them and can self-correct.
package app

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// serverName is the MCP implementation name advertised to clients.
const serverName = "tsvsheet-mcp"

// The registered tool names, shared by the tool definitions and their tests.
const (
	toolRender  = "tsvsheet_render"
	toolCheck   = "tsvsheet_check"
	toolExplain = "tsvsheet_explain"
	toolEval    = "tsvsheet_eval"
)

// Version is the server version an agent may read; the binary overrides it via
// NewServer at startup.
type Version string

// NewServer builds the MCP server with the four tsvsheet tools registered. Each
// tool infers its input schema from its typed argument struct and its 'jsonschema'
// field tags, so an agent sees documented parameters.
func NewServer(version Version) *mcp.Server {
	server := mcp.NewServer(
		&mcp.Implementation{Name: serverName, Version: string(version)},
		&mcp.ServerOptions{Instructions: string(Instructions())},
	)
	mcp.AddTool(server, renderToolDef, renderTool)
	mcp.AddTool(server, checkToolDef, checkTool)
	mcp.AddTool(server, explainToolDef, explainTool)
	mcp.AddTool(server, evalToolDef, evalTool)
	return server
}

// Run serves the MCP server over the transport until the context is cancelled
// or the client disconnects. The binary passes a stdio transport; tests pass an
// in-memory transport.
func Run(ctx context.Context, server *mcp.Server, transport mcp.Transport) error {
	return server.Run(ctx, transport)
}
