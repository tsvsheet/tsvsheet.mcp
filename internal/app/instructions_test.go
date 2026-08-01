package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/tsvsheet.mcp/internal/constants"
)

// TestInstructionsReachTheClient pins the contract that makes the
// --instructions flag trustworthy: what an agent is told at initialize is the
// same text the flag prints, so a user pasting the flag's output into a
// configuration reads exactly what the server advertises.
func TestInstructionsReachTheClient(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	go func() { _ = Run(ctx, NewServer(Version("test")), serverTransport) }()

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "test"}, nil)
	session, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer func() { _ = session.Close() }()

	assert.Equal(t, string(Instructions()), session.InitializeResult().Instructions)
}

func TestInstructionsNameEveryTool(t *testing.T) {
	text := string(Instructions())
	for _, tool := range []string{toolRender, toolCheck, toolExplain, toolEval} {
		assert.Contains(t, text, tool)
	}
}

func TestInstructionsAreSelfContainedProse(t *testing.T) {
	text := string(Instructions())
	assert.NotEmpty(t, text)
	assert.True(t, strings.HasSuffix(text, "\n"), "the text ends in a newline so it pipes cleanly")
	assert.NotContains(t, text, "\t", "tabs would be ambiguous in a document about tab-separated files")
}

// flowed is the instruction text with its reading wrap collapsed, so a phrase
// assertion is about the words rather than where the line broke.
func flowed() string { return strings.Join(strings.Fields(string(Instructions())), " ") }

// packageSource concatenates this package's non-test sources.
func packageSource(t *testing.T) string {
	t.Helper()
	entries, err := os.ReadDir(".")
	require.NoError(t, err)
	var all strings.Builder
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		content, readErr := os.ReadFile(filepath.Clean(entry.Name()))
		require.NoError(t, readErr)
		_, _ = all.Write(content)
	}
	return all.String()
}

// claims are the load-bearing sentences of the instruction text, each paired
// with the behaviour that makes it true.
//
// They are asserted by EQUALITY of the whole sentence, not by substring. An
// adversarial pass showed why: every earlier assertion was a phrase match, and
// a text rewritten to say the opposite still contained the phrases, so the
// entire document could be replaced with a self-declaredly false one and the
// suite stayed green. A claim an agent acts on has to be pinned to the words
// that make it, and to a check that it is true.
var claims = []struct {
	holds    func(*testing.T)
	sentence string
}{
	{
		sentence: "A \"#.\" line (a comment or a view directive) and a \"#!\" first line are " +
			"metadata: they are skipped when the grid is read and they do not occupy a row, " +
			"so A1 addressing counts only data rows.",
		holds: func(t *testing.T) {
			_, out, err := explainTool(context.Background(), nil, ExplainInput{
				Source: "#. a note\n#. another\nQ1\tQ2\n", Cell: "A1",
			})
			require.NoError(t, err)
			assert.Equal(t, "Q1", out.Value, "the third physical line is row 1")
		},
	},
	{
		sentence: "This server opens no files and reaches no network: send the source you " +
			"want computed, and it returns the result.",
		holds: func(t *testing.T) {
			for _, banned := range []string{`"os"`, `"net"`, `"net/http"`, `"os/exec"`} {
				assert.NotContains(t, packageSource(t), banned)
			}
		},
	},
	{
		sentence: "computes every formula and returns the resulting VALUE grid as tsv " +
			"(default), csv, html, or markdown.",
		holds: func(t *testing.T) {
			for _, format := range []string{"", "tsv", "csv", "html", "markdown"} {
				_, _, err := renderTool(context.Background(), nil, RenderInput{Source: "1\n", Format: format})
				assert.NoError(t, err, format)
			}
			for _, invented := range []string{"xlsx", "json", "pdf", "ods", "xml"} {
				_, _, err := renderTool(context.Background(), nil, RenderInput{Source: "1\n", Format: invented})
				assert.Error(t, err, invented)
			}
		},
	},
	{
		sentence: "Never write a render result back over the source — that discards them.",
		holds: func(t *testing.T) {
			_, out, err := renderTool(context.Background(), nil, RenderInput{
				Source: "#!/usr/bin/env tsvsheet\n#. keep\nQ1\t=A1\n", Format: "tsv",
			})
			require.NoError(t, err)
			for _, discarded := range []string{"#!", "keep", "=A1"} {
				assert.NotContains(t, out.Rendered, discarded)
			}
		},
	},
	{
		sentence: "reports one diagnostic per unknown function call, without evaluating.",
		holds: func(t *testing.T) {
			_, out, err := checkTool(context.Background(), nil, CheckInput{Source: "=bogus(A1)+alsobogus(A2)\n"})
			require.NoError(t, err)
			assert.Len(t, out.Diagnostics, 2, "two unknown calls in one cell are two diagnostics")
		},
	},
	{
		sentence: "That is its whole scope: it does not validate references, and it cannot " +
			"tell you a reference will compute to #REF!.",
		holds: func(t *testing.T) {
			_, out, err := checkTool(context.Background(), nil, CheckInput{Source: "=A99999\n"})
			require.NoError(t, err)
			assert.Empty(t, out.Diagnostics)
			_, rendered, renderErr := renderTool(context.Background(), nil, RenderInput{Source: "=A99999\n"})
			require.NoError(t, renderErr)
			assert.Contains(t, rendered.Rendered, "#REF!", "which check said nothing about")
		},
	},
	{
		sentence: "The cell address is upper-case A1 form (D5, not d5); the $ markers of an " +
			"absolute reference are accepted and ignored, so an address this tool reports " +
			"can be fed straight back to it.",
		holds: func(t *testing.T) {
			_, out, err := explainTool(context.Background(), nil, ExplainInput{
				Source: "1\t2\n=$A$1+1\t\n", Cell: "A2",
			})
			require.NoError(t, err)
			require.NotEmpty(t, out.Inputs)
			reported := out.Inputs[0].Ref
			_, back, backErr := explainTool(context.Background(), nil, ExplainInput{
				Source: "1\t2\n=$A$1+1\t\n", Cell: reported,
			})
			require.NoError(t, backErr, "an address this tool reports must be one it accepts: %q", reported)
			assert.Equal(t, "1", back.Value)

			_, _, lower := explainTool(context.Background(), nil, ExplainInput{Source: "1\n", Cell: "a1"})
			assert.ErrorIs(t, lower, constants.ErrInvalidCell)
		},
	},
	{
		sentence: "evaluates a single formula, with or without a leading =, as a one-cell sheet.",
		holds: func(t *testing.T) {
			for _, expression := range []Expression{"SUM(1,2,3)", "=SUM(1,2,3)"} {
				_, out, err := evalTool(context.Background(), nil, EvalInput{Expression: expression})
				require.NoError(t, err, expression)
				assert.Equal(t, "6", out.Value)
			}
		},
	},
	{
		sentence: "The error values are #REF!, #VALUE!, #N/A, #NAME?, #DIV/0!, #NUM!, " +
			"#CIRC!, #SPILL!, and #IMPORT!.",
		holds: func(t *testing.T) {
			// Each named value is one the language defines; none is invented.
			for _, value := range []string{"#REF!", "#VALUE!", "#N/A", "#NAME?", "#DIV/0!", "#NUM!", "#CIRC!", "#SPILL!", "#IMPORT!"} {
				_, out, err := renderTool(context.Background(), nil, RenderInput{Source: value + "\n"})
				require.NoError(t, err, value)
				assert.Contains(t, out.Rendered, value, "%s must round-trip as an error value literal", value)
			}
			// And one the engine declares but no formula can produce is absent:
			// naming it would send an agent looking for something impossible.
			assert.NotContains(t, flowed(), "#NULL!")
		},
	},
}

// TestEveryClaimIsPresentVerbatimAndTrue is the test the instruction text
// exists for: each load-bearing sentence must appear exactly as written, and
// the behaviour it describes must actually hold.
func TestEveryClaimIsPresentVerbatimAndTrue(t *testing.T) {
	text := flowed()
	for _, claim := range claims {
		want := strings.Join(strings.Fields(claim.sentence), " ")
		assert.Contains(t, text, want, "the instructions must state this claim verbatim")
		claim.holds(t)
	}
}
