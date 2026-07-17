// Package constants declares the error sentinels the MCP server emits. Each is
// an errs.Const so callers (and tests) match with errors.Is rather than string
// comparison; the tool handlers surface these to the agent as MCP error content.
package constants

import errs "github.com/gomatic/go-error"

// Keep these constants sorted alphabetically.
const (
	// ErrEmptyExpression is returned by tsvsheet_eval when the expression is
	// empty (or only leading `=` / whitespace), so there is nothing to compute.
	ErrEmptyExpression errs.Const = "expression must not be empty"
	// ErrInvalidCell is returned when a cell argument is not valid A1 notation
	// (column letters followed by a positive row, e.g. "D5").
	ErrInvalidCell errs.Const = "invalid cell reference"
	// ErrMultiCellExpression is returned by tsvsheet_eval when the expression
	// contains a TAB or newline — the grid's own field/row separators — which
	// would split it into multiple cells and silently evaluate only the first.
	ErrMultiCellExpression errs.Const = "expression must be a single cell: it may not contain tab or newline characters"
	// ErrParse is returned when the source cannot be parsed into a sheet — a
	// malformed formula names the offending cell.
	ErrParse errs.Const = "failed to parse spreadsheet source"
	// ErrUnknownCell is returned when an explained cell lies outside the grid.
	ErrUnknownCell errs.Const = "cell not found in grid"
	// ErrUnknownFormat is returned when the requested render format is not one
	// of tsv, csv, html, or markdown.
	ErrUnknownFormat errs.Const = "unknown output format"
)
