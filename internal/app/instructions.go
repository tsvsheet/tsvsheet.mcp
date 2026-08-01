package app

// This file holds the server's usage instructions: what an agent is told at
// initialize, and what `tsvsheet-mcp --instructions` prints. One text, two
// uses — so what a user pastes into a configuration is exactly what the server
// advertises. Every claim here is a claim an agent will act on, so
// instructions_test.go checks them against what the tools actually do rather
// than only that the text is non-empty.

// InstructionText is the server's usage guidance for an agent.
type InstructionText string

// Instructions returns the guidance an agent receives at initialize.
func Instructions() InstructionText { return instructions }

// instructions is the single source of the text. It is deliberately prose an
// agent can act on rather than a schema dump: the tool schemas already
// describe parameters, so what is missing from them is when to reach for
// which tool — and what each tool does NOT do.
const instructions InstructionText = `tsvsheet-mcp computes plain-text spreadsheets.

A .tsvt file is a spreadsheet: a TAB-separated grid whose cells are either
literal values or =formulas addressing other cells in A1 notation (B2, D2:D5).

Lines are not all rows. A "#." line (a comment or a view directive) and a "#!"
first line are metadata: they are skipped when the grid is read and they do not
occupy a row, so A1 addressing counts only data rows. Derive an address by
counting data rows, never physical lines. A "# " line is the superseded
hash-space comment form: still read, so do not be surprised by it, but write
"#." when you author a sheet.

The sheet travels in the request. This server opens no files and reaches no
network: send the source you want computed, and it returns the result. One
consequence is worth knowing before you read a result: because no loader and no
fetcher are supplied, every cross-sheet reference ("file"!A1, sheet(...))
computes to #REF! and every import (importcell, importrange, ...) computes to
#IMPORT! here, whatever they would do in a host that supplies them.

Which tool answers which question:

  ` + toolRender + `   "what does this sheet evaluate to?" — computes every formula and
                    returns the resulting VALUE grid as tsv (default), csv, html,
                    or markdown. The result is a view of the computed values, not
                    a replacement for the file: formulas, comments, view
                    directives, and any shebang are absent from it. Never write a
                    render result back over the source — that discards them.

  ` + toolCheck + `    "does this sheet call anything that does not exist?" — reports
                    one diagnostic per unknown function call, without evaluating.
                    That is its whole scope: it does not validate references, and
                    it cannot tell you a reference will compute to #REF!. To learn
                    that, render the sheet or explain the cell. A sheet that fails
                    to parse at all comes back as a tool error, not a diagnostic.

  ` + toolExplain + `  "why is this cell that value?" — computes the sheet, then reports
                    one cell's value, its formula, and the inputs it read. Reach
                    for this when a single result is surprising; it shows the
                    dependencies rather than making you infer them. The cell
                    address is upper-case A1 form (D5, not d5); the $ markers of
                    an absolute reference are accepted and ignored, so an address
                    this tool reports can be fed straight back to it.

  ` + toolEval + `     "what does this one expression compute to?" — evaluates a single
                    formula, with or without a leading =, as a one-cell sheet.
                    Reach for this to check a function's behaviour without
                    building a sheet around it.

Errors come back as tool errors carrying the engine's own message, so a failed
call tells you what to fix rather than only that something failed.

A cell whose formula cannot produce a value computes to an error value in
place, and the surrounding grid still computes — so a render or explain result
containing one is a finding to report, not a failed call. The error values are
#REF!, #VALUE!, #N/A, #NAME?, #DIV/0!, #NUM!, #CIRC!, #SPILL!, and #IMPORT!.
`
