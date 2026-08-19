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
	//
	// EVIDENCE PER ENTRY (council d8668e1f, editquality: an exemption map with
	// prose reasons is a second hand-written census — so each entry cites the
	// input contract and the write it is exempted for, read at HEAD 2026-08-16.
	// If you re-verify, re-cite; a line number that drifted is a stale claim).

	// fix_component_template_action.go — FixComponentTemplateInputSpec:
	// Required [site_id], Optional [fix_type, slot_name, page_component_id].
	// FIVE html_template writes (third added 2026-08-17, fourth and fifth 2026-08-19, all bugs_open/283):
	// fixRepairTemplateSlots (component_id from input_data.spec.component_id /
	// input_data.component_id; mechanical `<no value>x</no>` → `{{.x}}` after a
	// `<no value>` presence check; snapshots to component_versions
	// change_source='repair_template_slots'); fixChromeOverflow (component_id
	// resolved FROM site_components by (site_id, slot_name);
	// `html_template = html_template || $2` — an APPEND of a scoped media
	// query; counts and returns `shared_sites`; its own comment: "this template
	// is shared — the fix reaches every site that uses it"); and
	// fixScopeComponentInstance (component_id from spec.component_id; the
	// RFC_034 deterministic conversion — the component ROW is the subject and
	// fan-out to every placement is the POINT of the fix; gated by
	// GateConvertedTemplate before any write, snapshots to component_versions
	// change_source='scope_component_instance', and REFUSES the write entirely
	// when the script is unscoped rather than shipping ids-only); and
	// fixScopeComponentInstanceJudged (same spec.component_id subject; the
	// RFC_034 JUDGED half — an LLM-rewritten script gated by
	// JudgedConversionIssues + componentRegressionIssues before any write,
	// snapshots change_source='scope_component_instance_judged'. Fan-out to
	// every placement is again the POINT: 22 of the 25 judged rows are
	// component_level='section', one placed on two pages, which is exactly
	// the shape sharedComponentWriteCheck would refuse — and refusing it
	// would leave that component unconvertible. PLAN_2026-08-18_judged_pipeline §1.);
	// and fixRepairInstanceScopeBindings (spec.component_id; pass 5 applied to
	// a row the mechanical batch converted before pass 5 existed — 283 §14;
	// gated by UnprefixedBindings + GateConvertedTemplate + the comparative
	// write guard; snapshots change_source='repair_instance_scope_bindings').
	// No page-scoped input reaches any of the five. page_component_id is consumed ONLY by
	// fixAlignSlotName and fixPageComponentStatus, both page_components
	// METADATA writers. The 2026-08-15 header census that called this file
	// "page-aware … writes the component's template" conflated those two paths.
	"fix_component_template_action.go": "component-scoped writes only: mechanical slot repair (spec.component_id), chrome CSS APPEND via site_components with shared_sites recorded, the 283 instance-scope conversion (spec.component_id, gate-refused before write when the script is unscoped), its judged sibling (spec.component_id, LLM rewrite gated by JudgedConversionIssues + the comparative write guard; fan-out is the point — 22/25 judged rows are section-level), and the pass-5 binding repair (spec.component_id, gated by UnprefixedBindings + the two-instance gate + the write guard); page_component_id reaches only metadata fix types",

	// fix_harcoded_colours_action.go — FixHardcodedColorsInputSpec: Required
	// [site_id], Optional []. fixTemplateColors selects `SELECT DISTINCT cc.id …
	// WHERE cc.html_template ~ 'background(-color)?:\s*#…'` AND the component is
	// in use on THIS site (EXISTS site_components OR page_components→pages
	// site_id=$1), then `UPDATE content_components SET html_template = $1 …
	// WHERE id = $2` with checks.ReplaceHardcodedColors(template). No page_id,
	// no per-finding spec: the subject is "every template on this site that
	// hardcodes a background". A template SHARED with another site is rewritten
	// too — stated, and correct for a colour-declaration substitution
	// (colour_fixer_class_preservation_test.go measures it never touches
	// element attributes/structure).
	"fix_harcoded_colours_action.go": "site-scoped sweep over templates in use on the site (Required [site_id], no page input); colour-declaration substitution intended for every placement, cross-site included",

	// fix_forced_text_colours_action.go — FixForcedTextColorsInputSpec: Required
	// [site_id], Optional []; the same site-in-use selection shape (EXISTS
	// site_components OR page_components→pages site_id=$1) and the same
	// `UPDATE content_components SET html_template = $1 … WHERE id = $2` on a
	// declaration-level rewrite; no page_id, no per-finding spec.
	"fix_forced_text_colours_action.go": "site-scoped sweep over templates in use on the site (Required [site_id], no page input); colour-declaration substitution intended for every placement",

	// fix_nav_link_templates_action.go — FixNavLinkTemplatesInputSpec: Required
	// [site_id], Optional []; selects `sc.slot_name, sc.component_id,
	// cc.html_template FROM site_components sc JOIN content_components cc` for the
	// site's chrome slots and writes `UPDATE content_components SET html_template
	// = $1 … WHERE id = $2` per chrome component. Chrome templates are shared
	// across sites by design (site_components → content_components); no page
	// input exists on this action.
	"fix_nav_link_templates_action.go": "site chrome sweep via site_components (Required [site_id], no page input); nav-link template repair intended for every site on the template",

	// store_generated_component_action.go — StoreGeneratedComponentInputSpec:
	// Required [section_type], Optional [site_type, page_context, description,
	// design_direction, generated_template]. isRegeneration is set when an
	// existing component matches; it snapshots the CURRENT template to
	// component_versions ("Regenerated by component-creator") and UPDATEs by
	// existingID, then markPagesPendingRebuild flips every placement pending and
	// raises one needs_rerender per affected site. The component is the item's
	// subject (needs_component_regeneration); fan-out is the regen contract. Its
	// blast radius IS the known hazard (bugs_closed/285: one regen of the
	// ported-page wrapper drove 154 re-renders and 73 refused LLM rebuilds) —
	// that is a hazard of REGENERATING a shared component, not a mis-scoped
	// write, and it belongs to the regen path's own review, not to this fence.
	"store_generated_component_action.go": "regeneration keyed by the existing component id (Required [section_type], no page input); snapshot + UPDATE + pending fan-out is the regen contract",
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
		body := withoutLineComments(string(src))
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

	// Subpackages (discovery_checks/, queryresolve/, …) are NOT covered by the
	// map above by design — an html_template writer appearing there is a scope
	// decision this test must be told about, not a file it should silently miss
	// (council d8668e1f, editquality: Glob("*.go") is non-recursive; the 2026-08-16
	// census found none, and this keeps that a checked claim).
	subFiles, err := filepath.Glob("*/*.go")
	if err != nil {
		t.Fatalf("glob subpackages: %v", err)
	}
	var subWriters []string
	for _, f := range subFiles {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		src, readErr := os.ReadFile(f)
		if readErr != nil {
			t.Fatalf("read %s: %v", f, readErr)
		}
		if updatesHTMLTemplate.MatchString(withoutLineComments(string(src))) {
			subWriters = append(subWriters, f)
		}
	}
	if len(subWriters) > 0 {
		t.Fatalf("content_components.html_template is rewritten from a SUBPACKAGE, which this test's "+
			"census does not cover: %s — decide its scope (fence call or fan-out reason) and extend "+
			"fanOutIntendedWriters/this scan to include it", strings.Join(subWriters, ", "))
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
		if !updatesHTMLTemplate.MatchString(withoutLineComments(string(src))) {
			t.Errorf("%s is declared fan-out-intended but no longer writes html_template — remove the stale entry", f)
		}
	}
}

// withoutLineComments drops `//` line comments so a statement quoted in a doc
// comment is neither a writer nor a fence call. Deliberately naive (a `//`
// inside a string literal on the same line is also dropped): the cost is a
// false "not a writer" on a line that has BOTH SQL and a `//` before it, which
// gofmt'd Go SQL literals never do; the alternative (a Go tokenizer) is more
// than this test's altitude warrants.
func withoutLineComments(src string) string {
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
