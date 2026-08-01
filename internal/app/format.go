package app

import (
	"encoding/csv"
	"html"
	"io"
	"strings"

	tsvsheet "github.com/tsvsheet/go-tsvsheet"

	"github.com/tsvsheet/tsvsheet.mcp/internal/constants"
)

// Format selects a serialization of the computed value grid: TSV (the default),
// CSV, an HTML <table>, or a GitHub-flavored Markdown pipe table. One engine,
// one grid, many renderings.
type Format string

// The render output formats. md is accepted as an alias for markdown; every
// other value is an ErrUnknownFormat.
const (
	formatTSV      Format = "tsv"
	formatCSV      Format = "csv"
	formatHTML     Format = "html"
	formatMarkdown Format = "markdown"
	formatMD       Format = "md"
)

// render serializes the computed grid in format f. An unrecognized format is
// constants.ErrUnknownFormat naming the offending value (matchable with
// errors.Is).
func render(grid tsvsheet.Grid, f Format) (string, error) {
	switch f {
	case formatTSV:
		return renderTSV(grid), nil
	case formatCSV:
		var b strings.Builder
		err := writeCSV(&b, grid)
		return b.String(), err
	case formatHTML:
		return renderHTML(grid), nil
	case formatMarkdown, formatMD:
		return renderMarkdown(grid), nil
	default:
		return "", constants.ErrUnknownFormat.With(nil, "format", string(f))
	}
}

// renderTSV joins each row with tabs and terminates every row with a newline,
// the same bytes the engine's canonical TSV serialization produces. See TestRenderFormats.
func renderTSV(grid tsvsheet.Grid) string {
	lines := make([]string, len(grid))
	for i, row := range grid {
		lines[i] = strings.Join(row, "\t") + "\n"
	}
	return strings.Join(lines, "")
}

// writeCSV writes the grid to w as RFC 4180 CSV, quoting any cell that contains
// a comma, quote, or newline. A write failure surfaces as
// constants.ErrParse wrapping the cause; a strings.Builder never fails, so the
// error path is exercised only by an injected failing writer.
func writeCSV(w io.Writer, grid tsvsheet.Grid) error {
	cw := csv.NewWriter(w)
	if err := cw.WriteAll(grid); err != nil {
		return constants.ErrParse.With(err)
	}
	return nil
}

// renderHTML serializes the grid as a plain, class-tagged HTML <table> — one
// <tr> per row, one HTML-escaped <td> per cell, no inline styles — so it
// composes with any stylesheet.
func renderHTML(grid tsvsheet.Grid) string {
	rows := make([]string, len(grid))
	for i, row := range grid {
		rows[i] = htmlRow(row)
	}
	return `<table class="tsvsheet">` + "\n" + strings.Join(rows, "") + "</table>\n"
}

// htmlRow renders one grid row as a <tr> of HTML-escaped <td> cells.
func htmlRow(row []string) string {
	cells := make([]string, len(row))
	for i, cell := range row {
		cells[i] = "<td>" + html.EscapeString(cell) + "</td>"
	}
	return "<tr>" + strings.Join(cells, "") + "</tr>\n"
}

// renderMarkdown serializes the grid as a GitHub-flavored pipe table: the first
// row is the header, then a --- separator row, then the remaining rows. A | in a
// cell is escaped as \|. An empty grid yields no output.
func renderMarkdown(grid tsvsheet.Grid) string {
	if len(grid) == 0 {
		return ""
	}
	rows := make([]string, 0, len(grid)+1)
	rows = append(rows, markdownRow(grid[0]), markdownRow(separatorRow(grid[0])))
	for _, row := range grid[1:] {
		rows = append(rows, markdownRow(row))
	}
	return strings.Join(rows, "")
}

// markdownRow renders one grid row as a pipe-delimited table row. Each cell is
// escaped so no value breaks the table: a | is backslash-escaped, so it does not
// starts a new column, and a newline — which a cell can hold via CHAR(10) —
// becomes a <br>, keeping the row on one table line (GFM renders
// <br> as an in-cell line break).
func markdownRow(row []string) string {
	cells := make([]string, len(row))
	for i, cell := range row {
		escaped := strings.ReplaceAll(cell, "|", `\|`)
		escaped = strings.ReplaceAll(escaped, "\r\n", "<br>")
		escaped = strings.ReplaceAll(escaped, "\n", "<br>")
		cells[i] = strings.ReplaceAll(escaped, "\r", "<br>")
	}
	return "| " + strings.Join(cells, " | ") + " |\n"
}

// separatorRow builds the --- header/body divider matching the header's column
// count.
func separatorRow(header []string) []string {
	cells := make([]string, len(header))
	for i := range cells {
		cells[i] = "---"
	}
	return cells
}
