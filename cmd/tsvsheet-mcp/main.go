// Command tsvsheet-mcp is the Model Context Protocol (MCP) server for the
// tsvsheet engine. It speaks MCP over stdio (JSON-RPC) so AI agents can compute
// and inspect .tsvt spreadsheets through four tools: tsvsheet_render,
// tsvsheet_check, tsvsheet_explain, and tsvsheet_eval. Every tool consumes the
// same github.com/tsvsheet/go-tsvsheet engine; nothing here re-implements parsing
// or evaluation.
package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/urfave/cli/v3"

	"github.com/tsvsheet/tsvsheet.mcp/internal/app"
)

const (
	name  = `tsvsheet-mcp`
	usage = `Model Context Protocol server exposing the tsvsheet engine over stdio.`

	description = `tsvsheet-mcp is the Model Context Protocol (MCP) server for tsvsheet.

It speaks MCP over stdio (JSON-RPC) so AI agents can compute and inspect .tsvt
spreadsheets. A .tsvt is a single TAB-separated grid whose cells are literal
values or =formulas addressing other cells in A1 notation.

Tools:
  tsvsheet_render   compute the sheet and render the grid (tsv, csv, html, markdown)
  tsvsheet_check    report static diagnostics for the sheet
  tsvsheet_explain  explain one cell: value, formula, and the inputs it reads
  tsvsheet_eval     compute a single formula and return its value`
)

// version is the application version. Set via ldflags: -X main.version=1.0.0
var version = "dev"

// The following are indirected so the run path stays testable without opening a
// real stdio session: osExit observes the process exit code, appCreator builds
// the CLI, and newServer/runServer/transport are the seams the root Action
// dials through.
var (
	osExit                   = os.Exit
	appCreator               = createApp
	newServer                = app.NewServer
	runServer                = app.Run
	transport  mcp.Transport = &mcp.StdioTransport{}
)

func main() { osExit(run(os.Args)) }

// run builds and executes the CLI, returning the process exit code. Keeping the
// exit code a return value (rather than calling os.Exit here) makes the whole
// run path testable.
func run(args []string) int {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	if err := appCreator().Run(ctx, args); err != nil {
		return 1
	}
	return 0
}

// serve is the root Action: it builds the MCP server and serves it over the
// configured transport until the context is cancelled or the client disconnects.
func serve(ctx context.Context, _ *cli.Command) error {
	return runServer(ctx, newServer(app.Version(version)), transport)
}

// createApp constructs the CLI: a single root command whose Action starts the
// stdio MCP server.
func createApp() *cli.Command {
	return &cli.Command{
		Name:                  name,
		Usage:                 usage,
		Description:           description,
		Version:               version,
		EnableShellCompletion: true,
		Action:                serve,
	}
}
