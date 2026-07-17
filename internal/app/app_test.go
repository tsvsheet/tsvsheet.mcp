package app

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestServerRoundTrip wires NewServer to a client over in-memory transports and
// exercises Run end-to-end: it lists the registered tools, calls one for a
// happy result and one for an error result, then cancels the context and
// confirms Run returns.
func TestServerRoundTrip(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	server := NewServer(Version("test"))

	runErr := make(chan error, 1)
	go func() { runErr <- Run(ctx, server, serverTransport) }()

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "test"}, nil)
	session, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)

	tools, err := session.ListTools(ctx, nil)
	require.NoError(t, err)
	names := make([]string, len(tools.Tools))
	for i, tool := range tools.Tools {
		names[i] = tool.Name
	}
	assert.ElementsMatch(t, []string{toolRender, toolCheck, toolExplain, toolEval}, names)

	good, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      toolEval,
		Arguments: map[string]any{"expression": "SUM(1,2,3)"},
	})
	require.NoError(t, err)
	assert.False(t, good.IsError)
	assert.Equal(t, "6", textOf(t, good))

	bad, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      toolRender,
		Arguments: map[string]any{"source": "1", "format": "yaml"},
	})
	require.NoError(t, err)
	assert.True(t, bad.IsError)

	// Cancel while the session is still open so Run takes its context-cancelled
	// path deterministically (rather than racing a client-close session end).
	cancel()
	assert.ErrorIs(t, <-runErr, context.Canceled)
	_ = session.Close()
}
