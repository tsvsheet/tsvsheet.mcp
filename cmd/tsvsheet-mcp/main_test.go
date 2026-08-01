package main

import (
	"bytes"
	"context"
	"errors"
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

// TestInstructionsFlagPrintsAndServesNothing pins the flag's contract: it
// writes the server's advertised instructions to stdout and exits 0 without
// opening a transport, so it can be piped into an agent's configuration
// without a server ever starting.
func TestInstructionsFlagPrintsAndServesNothing(t *testing.T) {
	origRun, origOut := runServer, stdout
	t.Cleanup(func() { runServer, stdout = origRun, origOut })

	served := false
	runServer = func(context.Context, *mcp.Server, mcp.Transport) error {
		served = true
		return nil
	}
	var out bytes.Buffer
	stdout = &out

	assert.Equal(t, 0, run([]string{name, "--instructions"}))
	assert.Equal(t, string(app.Instructions()), out.String())
	assert.False(t, served, "the flag starts no server")
}

func TestWithoutTheFlagTheServerRuns(t *testing.T) {
	origRun := runServer
	t.Cleanup(func() { runServer = origRun })
	served := false
	runServer = func(context.Context, *mcp.Server, mcp.Transport) error {
		served = true
		return nil
	}
	assert.Equal(t, 0, run([]string{name}))
	assert.True(t, served)
}

// failingWriter fails every write, standing in for a full disk or a closed
// descriptor.
type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("no space left on device") }

// TestInstructionsWriteFailureIsReported pins the fix for a defect an
// adversarial pass found: the flag exists to be redirected into a
// configuration file, so a write that does not land must not exit 0. A
// truncated config reported as success is worse than a failure, because the
// caller cannot tell it happened.
func TestInstructionsWriteFailureIsReported(t *testing.T) {
	origOut := stdout
	t.Cleanup(func() { stdout = origOut })
	var diagnostics bytes.Buffer
	stdout = failingWriter{}
	prev := stderr
	stderr = &diagnostics
	t.Cleanup(func() { stderr = prev })

	assert.Equal(t, 1, run([]string{name, "--instructions"}))
	assert.Contains(t, diagnostics.String(), name, "the failure names the command")
	assert.Contains(t, diagnostics.String(), constants.ErrWriteInstructions.Error(),
		"and the specific sentinel, so a caller can tell a failed write from a transport fault")
}

// TestInstructionsWriteFailureCarriesItsSentinel pins the identity of the
// error, not merely that one occurred: a mutation swapping it for an unrelated
// sentinel previously survived.
func TestInstructionsWriteFailureCarriesItsSentinel(t *testing.T) {
	origOut := stdout
	t.Cleanup(func() { stdout = origOut })
	stdout = failingWriter{}
	err := serve(context.Background(), instructionsCommand())
	require.Error(t, err)
	assert.ErrorIs(t, err, constants.ErrWriteInstructions)
}

// instructionsCommand is a parsed command with --instructions set.
func instructionsCommand() *cli.Command {
	cmd := createApp()
	cmd.Writer = io.Discard
	_ = cmd.Set(flagInstructions, "true")
	return cmd
}
