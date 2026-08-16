package actions

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// Coverage test for the shared-component write fence (bugs_open/281, /285):
// EVERY action that rewrites content_components.html_template must either ask
// sharedComponentWriteCheck or be listed here with the reason its fan-out is
// intended.
//
// WHY. The fence closes the write that fired twice (2026-08-05, 2026-08-14) —
// tool-improver's update_component_html — but a fence on ONE writer is exactly
// the shape bugs_closed/021 named ("durable write guard covers one path only")
// and the council's bug_historian seat asked about on 360ae540 ("is there a
// SECOND writer bypassing the fence?"). The answer that day was a hand census in
// component_write_guard.go's header. A header census is true on the day it is
// written; this test is true on every build. A seventh writer fails here until
// its author decides, in writing, which side of the line it is on.
//
// Same shape and same known weakness as page_component_writer_coverage_test.go:
// it reads SOURCE, so it proves the CALL EXISTS, not that it executes. Comment
// lines are stripped before matching so a quoted statement in a doc comment
// (component_write_guard.go quotes the census grep) cannot register as a
// writer, and naming the fence in a comment cannot satisfy it — the LANDMINES
// entry "a source-scan test makes your COMMENTS load-bearing" is why.

// updatesHTMLTemplate matches a statement that sets html_template on an
// EXISTING content_components row (`SET html_template = …`, including the
// append form `html_template = html_template || …`). INSERTs are the birth
// path, gated elsewhere (store_generated_component's own checks).
var updatesHTMLTemplate = regexp.MustCompile(`(?is)UPDATE\s+content_components\b[^;]{0,400}?\bSET\b[^;]{0,400}?\bhtml_template\s*=`)

// asksTheFence: the one decision helper. Callers do not have to REFUSE on its
// answer (tool forks warn+proceed) — but they have to ask.
var asksTheFence = regexp.MustCompile(`\bsharedComponentWriteCheck\s*\(`)

// fanOutIntendedWriters — the writer census of 2026-08-15 (component_write_guard.go
// header, corrected 2026-08-16 by the bugfix_285 lane), in code, with the reason
// each may write a shared template WITHOUT the fence. A file here is a DECISION.
// The test for membership: is the COMPONENT the subject of the finding (fan-out to
// every placement is the point), or is a PAGE the subject (fan-out is the bug)?
// Only page-scoped writers belong outside this map.
var fanOutIntendedWriters = map[string]string{
	// NOTE — update_component_html_action.go is deliberately NOT here. It is the
	// page-scoped writer (tool-improver: one page's finding) and it passes on its
	// own merits by calling the fence; exempting it would let that call be
	// unwired undetected, which is the hole this test exists to close.

	// Two writes, neither a per-page LLM rewrite. repair_template_slots: a
	// mechanical `<no value>x</no>` → `{{.x}}` repair keyed by spec.component_id —
	// on a shared template that repair is right for every placement.
	// chrome_overflow_fix: APPENDS a media query to a chrome component reached via
	// site_components (never page_components), records `shared_sites` in its
	// result, and says in its own comment that the shared reach is intended.
	// The 2026-08-15 census called this file "page-aware (takes page_component_id,
	// reads the page's rendered_html)"; that describes its align_slot_name /
	// repair_page_component_status METADATA paths, which never touch html_template.
	"fix_component_template_action.go": "component-scoped mechanical repair + chrome CSS append; no per-page LLM rewrite reaches html_template here",

	// Colour rewrites addressed BY COMPONENT: the finding is "this template
	// hardcodes colours", so every placement is meant to change.
	"fix_harcoded_colours_action.go":    "component is the subject; colour rewrite intended for every placement",
	"fix_forced_text_colours_action.go": "component is the subject; colour rewrite intended for every placement",

	// Nav-link template repair addressed by component (site_components-backed
	// chrome); the whole point is that every site on the template gets the fix.
	"fix_nav_link_templates_action.go": "component is the subject; nav-link template repair intended for every placement",

	// component-creator regeneration: the component IS the work item's subject
	// and the regen snapshots the pre-state (component_versions) and flips
	// placements pending by design. Its blast radius is the known hazard of a
	// regen (bugs_open/285: one regen of the ported-page wrapper drove 154
	// re-renders), but that is the regen contract, not a mis-scoped write.
	"store_generated_component_action.go": "regeneration; the component is the subject and fan-out is the contract",
}

func TestEveryHTMLTemplateRewriterAsksTheSharedFenceOrIsDeclaredFanOut(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob: %v", err)
	}

	var unfenced []string
	var sawAnyWriter, sawTheFencedWriter bool

	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		src, readErr := os.ReadFile(f)
		if readErr != nil {
			t.Fatalf("read %s: %v", f, readErr)
		}
		body := stripLineComments(string(src))
		if !updatesHTMLTemplate.MatchString(body) {
			continue
		}
		sawAnyWriter = true
		if _, intended := fanOutIntendedWriters[f]; intended {
			continue
		}
		if asksTheFence.MatchString(body) {
			if f == "update_component_html_action.go" {
				sawTheFencedWriter = true
			}
			continue
		}
		unfenced = append(unfenced, f)
	}

	// The detector must have found SOMETHING, or a regex that silently stopped
	// matching would make this test pass by seeing no writers at all.
	if !sawAnyWriter {
		t.Fatal("found NO content_components html_template writers — the detector is broken, " +
			"not the codebase; this test cannot pass vacuously")
	}
	// And it must have seen the writer that fired twice DOING the right thing —
	// otherwise a rename of that file, or a regex drift that stops seeing it,
	// would pass by omission.
	if !sawTheFencedWriter {
		t.Fatal("update_component_html_action.go was not seen as a fenced html_template writer — " +
			"either the fence call was removed (bugs_open/285: the write that fired 2026-08-05 and 2026-08-14) " +
			"or the detector no longer sees it; both are failures")
	}

	if len(unfenced) > 0 {
		t.Fatalf("these actions rewrite content_components.html_template without asking "+
			"sharedComponentWriteCheck and are not declared fan-out-intended: %s\n"+
			"A write keyed by component_id reaches EVERY placement (the ported-page wrapper is one row "+
			"behind ~115 pages on 2 sites). If the finding is page-scoped, call sharedComponentWriteCheck "+
			"and refuse (or write the instance's page_components.rendered_html instead); if the component "+
			"is the subject, add the file to fanOutIntendedWriters WITH A REASON. bugs_open/285.",
			strings.Join(unfenced, ", "))
	}
}

// The exemption list must not rot into a place where files are parked: every
// entry has to still be a writer, or a stale entry would silently cover a
// future one that lands in that file.
func TestFanOutIntendedWritersAreAllStillHTMLTemplateWriters(t *testing.T) {
	for f, reason := range fanOutIntendedWriters {
		if strings.TrimSpace(reason) == "" {
			t.Errorf("%s is declared fan-out-intended with no reason", f)
		}
		src, err := os.ReadFile(f)
		if err != nil {
			t.Errorf("%s is declared fan-out-intended but cannot be read: %v", f, err)
			continue
		}
		if !updatesHTMLTemplate.MatchString(stripLineComments(string(src))) {
			t.Errorf("%s is declared fan-out-intended but no longer writes html_template — remove the stale entry", f)
		}
	}
}

// stripLineComments drops `//` line comments so a statement quoted in a doc
// comment is neither a writer nor a fence call. Deliberately naive (a `//`
// inside a string literal on the same line is also dropped): the cost is a
// false "not a writer" on a line that has BOTH SQL and a `//` before it, which
// gofmt'd Go SQL literals never do; the alternative (a Go tokenizer) is more
// than this test's altitude warrants.
func stripLineComments(src string) string {
	var b strings.Builder
	for _, line := range strings.Split(src, "\n") {
		if i := strings.Index(line, "//"); i >= 0 {
			line = line[:i]
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}
	return b.String()
}
