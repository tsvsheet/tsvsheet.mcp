package main

import (
	"context"
	"io"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	"github.com/tsvsheet/tsvsheet.mcp/internal/app"
	"github.com/tsvsheet/tsvsheet.mcp/internal/constants"
)

// stubCommand builds a minimal cli.Command whose Action returns err. It skips
// flag parsing so the ambient test-binary args (-test.*) never reach the parser.
func stubCommand(err error) *cli.Command {
	return &cli.Command{
		Name:            "x",
		SkipFlagParsing: true,
		Writer:          io.Discard,
		Action:          func(context.Context, *cli.Command) error { return err },
	}
}

func TestVersion(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "dev", version)
}

func TestServe(t *testing.T) {
	origServer, origRun := newServer, runServer
	t.Cleanup(func() { newServer, runServer = origServer, origRun })

	sentinel := app.NewServer("stub")
	var gotServer *mcp.Server
	newServer = func(app.Version) *mcp.Server { return sentinel }
	runServer = func(_ context.Context, s *mcp.Server, _ mcp.Transport) error {
		gotServer = s
		return nil
	}

	require.NoError(t, serve(context.Background(), nil))
	assert.Same(t, sentinel, gotServer)
}

func TestCreateApp(t *testing.T) {
	origRun := runServer
	t.Cleanup(func() { runServer = origRun })
	runServer = func(context.Context, *mcp.Server, mcp.Transport) error { return nil }

	cmd := createApp()
	require.NotNil(t, cmd)
	assert.Equal(t, name, cmd.Name)
	assert.Equal(t, version, cmd.Version)

	// Running with just the program name takes the normal path: the root Action
	// (serve) runs, dialing the stubbed runServer.
	require.NoError(t, cmd.Run(context.Background(), []string{name}))
}

func TestRun(t *testing.T) {
	origCreator := appCreator
	t.Cleanup(func() { appCreator = origCreator })

	tests := []struct {
		actionErr error
		name      string
		want      int
	}{
		{name: "success returns zero", actionErr: nil, want: 0},
		{name: "action error returns one", actionErr: constants.ErrParse, want: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appCreator = func() *cli.Command { return stubCommand(tt.actionErr) }
			assert.Equal(t, tt.want, run([]string{"x"}))
		})
	}
}

func TestMainFunc(t *testing.T) {
	origCreator, origExit := appCreator, osExit
	t.Cleanup(func() { appCreator, osExit = origCreator, origExit })

	appCreator = func() *cli.Command { return stubCommand(nil) }

	var code int
	osExit = func(c int) { code = c }

	main()
	assert.Equal(t, 0, code)
}
