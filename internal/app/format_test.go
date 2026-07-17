package app

import (
	"testing"

	errs "github.com/gomatic/go-error"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tsvsheet "github.com/tsvsheet/go-tsvsheet"

	"github.com/tsvsheet/tsvsheet.mcp/internal/constants"
)

// errWrite is the failure a failWriter returns.
const errWrite errs.Const = "write failed"

// failWriter is an io.Writer that always fails, exercising the writeCSV error
// path a strings.Builder can never reach.
type failWriter struct{}

func (failWriter) Write([]byte) (int, error) { return 0, errWrite }

func TestRenderFormats(t *testing.T) {
	t.Parallel()

	grid := tsvsheet.Grid{{"a", "b"}, {"c|d", "e"}}

	tests := []struct {
		name   string
		format Format
		want   string
	}{
		{name: "tsv", format: formatTSV, want: "a\tb\nc|d\te\n"},
		{name: "csv", format: formatCSV, want: "a,b\nc|d,e\n"},
		{
			name:   "html",
			format: formatHTML,
			want:   `<table class="tsvsheet">` + "\n<tr><td>a</td><td>b</td></tr>\n<tr><td>c|d</td><td>e</td></tr>\n</table>\n",
		},
		{name: "markdown", format: formatMarkdown, want: "| a | b |\n| --- | --- |\n| c\\|d | e |\n"},
		{name: "md alias", format: formatMD, want: "| a | b |\n| --- | --- |\n| c\\|d | e |\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := render(grid, tt.format)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRenderMarkdownNewline(t *testing.T) {
	t.Parallel()
	// A newline in a cell (reachable via CHAR(10)) becomes <br> so it cannot
	// break the row into two table lines.
	got, err := render(tsvsheet.Grid{{"H1", "H2"}, {"a\nb", "c\r\nd"}}, formatMarkdown)
	require.NoError(t, err)
	assert.Equal(t, "| H1 | H2 |\n| --- | --- |\n| a<br>b | c<br>d |\n", got)
}

func TestRenderHTMLEscapes(t *testing.T) {
	t.Parallel()
	got := renderHTML(tsvsheet.Grid{{"<b>&"}})
	assert.Equal(t, `<table class="tsvsheet">`+"\n<tr><td>&lt;b&gt;&amp;</td></tr>\n</table>\n", got)
}

func TestRenderUnknownFormat(t *testing.T) {
	t.Parallel()
	got, err := render(tsvsheet.Grid{{"a"}}, Format("yaml"))
	assert.Empty(t, got)
	require.ErrorIs(t, err, constants.ErrUnknownFormat)
	assert.Contains(t, err.Error(), "yaml")
}

func TestRenderMarkdownEmptyGrid(t *testing.T) {
	t.Parallel()
	got, err := render(tsvsheet.Grid{}, formatMarkdown)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestWriteCSVError(t *testing.T) {
	t.Parallel()
	err := writeCSV(failWriter{}, tsvsheet.Grid{{"a"}})
	require.ErrorIs(t, err, constants.ErrParse)
	assert.ErrorIs(t, err, errWrite)
}

func TestCellValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want string
		grid tsvsheet.Grid
	}{
		{name: "top-left", grid: tsvsheet.Grid{{"6", "x"}}, want: "6"},
		{name: "empty grid", grid: tsvsheet.Grid{}, want: ""},
		{name: "empty row", grid: tsvsheet.Grid{{}}, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, cellValue(tt.grid))
		})
	}
}
