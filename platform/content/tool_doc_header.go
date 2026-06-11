// FILE: platform/content/tool_doc_header.go  (relocation note: place wherever
// compile_page_sections lives so the render path imports it without a cycle;
// the tool_health check and create_tool_component validation import the same
// package. Pure functions, no dependencies.)
//
// The tool-doc header contract (019 §Tool Doc Header): every NEW tool's
// html_template begins its <script> with ONE sentinel-delimited JS block
// comment carrying purpose + behavioural invariants. The header is read by
// the tool-auditor / tool-improver / the future JS parser (all from the DB
// html_template) and is STRIPPED from rendered_html before deploy, so it
// never ships to the public page.
//
// Why sentinel-strip and NOT general comment stripping: removing arbitrary JS
// comments safely requires a JS-aware minifier (a regex destroys "https://…"
// string literals and regex literals). Exact-sentinel string scanning cannot
// corrupt code that lacks the sentinel, is idempotent, and needs no parser.

package content

import "strings"

const (
	// ToolDocOpen / ToolDocClose are the EXACT sentinels. The generator emits
	// them verbatim; validation and tool_health test for them verbatim; the
	// stripper removes everything between and including them. Do not vary the
	// spelling — the contract is the exact byte sequence.
	ToolDocOpen  = "/* === tool-doc ==="
	ToolDocClose = "=== /tool-doc === */"
)

// HasToolDocHeader reports whether template contains a well-formed tool-doc
// header: both sentinels present, opener before closer. Used by
// create_tool_component validation and the tool_health `no_doc_header` check
// (severity: warning — old tools converge on the normal sweep, no retrofit
// campaign).
func HasToolDocHeader(template string) bool {
	open := strings.Index(template, ToolDocOpen)
	if open < 0 {
		return false
	}
	close := strings.Index(template[open+len(ToolDocOpen):], ToolDocClose)
	return close >= 0
}

// StripToolDocHeader returns template with every tool-doc header block
// removed (sentinels inclusive), plus the immediately following newline so no
// blank gap ships. Idempotent: input without sentinels is returned unchanged.
// Call it where rendered_html is produced (compile_page_sections), for every
// component — it is a cheap string scan and a no-op when the block is absent.
//
// A malformed block (opener without closer) is left UNTOUCHED rather than
// guessed at: shipping an inert comment is harmless; truncating someone's
// script is not. The tool_health check surfaces the malformation instead.
func StripToolDocHeader(template string) string {
	for {
		open := strings.Index(template, ToolDocOpen)
		if open < 0 {
			return template
		}
		rest := template[open+len(ToolDocOpen):]
		close := strings.Index(rest, ToolDocClose)
		if close < 0 {
			return template // malformed — leave for tool_health to flag
		}
		end := open + len(ToolDocOpen) + close + len(ToolDocClose)
		// Swallow one trailing newline (and its \r) so no blank line ships.
		if end < len(template) && template[end] == '\r' {
			end++
		}
		if end < len(template) && template[end] == '\n' {
			end++
		}
		template = template[:open] + template[end:]
	}
}
