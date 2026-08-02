// FILE: platform/orchestration/actions/component_fallback_guard.go
//
// Birth gate for RFC_009 option B (owner ruling 2026-08-03, "C now, B next").
//
// A component template must not SUBSTITUTE A BUSINESS FACT when the site's datum
// is absent. `contact-info` did exactly that from the library's birth until
// 2026-08-02 (bugs_open/140):
//
//	{{if .phone}}…{{else}}+1234567890{{end}}                        <- a tel: link
//	{{if .hours}}{{.hours}}{{else}}Monday – Friday, 9am – 6pm{{end}}
//	{{if .email}}…{{else}}info@example.com{{end}}
//
// Eight live commercial sites served the invented hours; one served the invented
// phone. Every render "succeeded", nothing errored, and the fabricated values were
// styled identically to the real ones — so no reader, human or machine, could tell
// them apart. On a platform whose whole claims apparatus exists to stop unsourced
// assertions reaching a page, the component library asserted unverifiable business
// facts by default.
//
// ---------------------------------------------------------------------------
// THE DISTINCTION, which is the entire design.
//
//	A LABEL default is legitimate and the library is full of them:
//	  {{else}}Read more{{end}} · {{else}}Get Started{{end}} · {{else}}Send Message{{end}}
//	It names a control or a section. A site that does not override it is not
//	thereby making a claim about itself.
//
//	A FACT default is a fabrication:
//	  a phone number, an email address, a postal address, a price, a bare domain,
//	  a set of opening hours. Nobody stated it and no evidence register holds it.
//
// This is NOT a distinction invented here. `input_schema` already draws it —
// `on_missing: "skip_field"` for facts, an explicit `"fallback": "<text>"` with
// `use_fallback` for labels. contact-info declared skip_field for exactly the three
// fields that fabricated. The contract was right; only the template ignored it, and
// nothing enforced the contract. RFC_009 is the standing question of whether the
// RENDERER should enforce it; this is the cheap half — refuse it at the door.
// ---------------------------------------------------------------------------
//
// CALIBRATION — run before this shipped, and re-run it before you change anything
// here. The instruction is the same one component_write_guard.go's header gives,
// for the same reason: a guard that refuses good work gets switched off, and then
// it protects nothing.
//
//	Corpus: every component_versions row + every content_components row,
//	347 writes in total, exported 2026-08-03.
//	Result: 0 findings. This guard would have refused NOTHING in the platform's
//	entire recorded write history.
//
// That number is only half a calibration, and the other half nearly went missing.
// Zero false positives is also what a guard that does nothing scores — and by the
// time it ran, the corpus no longer contained the defect, because migration 287 had
// already repaired it. **Your own fix silences your own detector.** So the
// motivating case was recovered from the before-image that migration took
// (migration_backups, 287_contact_info_obeys_its_own_schema.sql) and re-tested:
//
//	Result: 5 findings on the pre-fix contact-info template — both phone literals,
//	the hours literal, and both copies of the example.com email.
//
// Both halves are pinned by tests below. If you only ever check that a guard is
// quiet on the corpus, you cannot tell "correct" from "inert".
//
// ---------------------------------------------------------------------------
// PARITY WITH THE PYTHON LINT — deliberate, and pinned.
//
// The same rule runs daily as `component-fallback-check` (CGV-029,
// deployments/kustomize/services/component-fallback-check/), in Python, because it
// must read the LIVE library from the database and a Go job there would need a
// clone and a compile. So this rule now has two implementations, which is the exact
// drift class it exists to catch, and pretending otherwise would be dishonest.
//
// They are pinned by a SHARED FIXTURE: testdata/component_fallback_fixtures.json
// carries the cases and the expected verdict for each. This file's tests read it,
// and `check_placeholder_fallbacks.py --selftest` reads the same file. A change to
// either implementation that is not matched in the other fails on one side.
//
// The two are deliberately scoped differently and that is not drift:
//   - Python also reports the milder "declared skip_field but never gated" class
//     (renders a blank, asserts nothing untrue — 68 fleet-wide, reported, never
//     blocking). A birth gate must not refuse a write for that.
//   - This one runs only at the write, against one template, with no DB.
// ---------------------------------------------------------------------------

package actions

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// fallbackBlockRe captures {{if …}} IF_BRANCH {{else}} LITERAL {{end}} as a pair,
// because whether the literal is a SUBSTITUTE is decidable only against the branch
// it replaces — see sameTextBothBranches below.
var fallbackBlockRe = regexp.MustCompile(
	`(?s)\{\{-?\s*(?:if|else\s+if)\s+[^}]*?-?\}\}(.*?)\{\{-?\s*else\s*-?\}\}(.*?)\{\{-?\s*end\s*-?\}\}`)

var htmlTagRe = regexp.MustCompile(`<[^>]+>`)
var whitespaceRe = regexp.MustCompile(`\s+`)

// factShape is one way a literal can assert something checkable about a business.
type factShape struct {
	name string
	re   *regexp.Regexp
}

// factShapes — every pattern here is RE2-safe (no lookaround), and every one is
// exercised by the shared fixture.
var factShapes = []factShape{
	// +1234567890 · +1 (234) 567-890 · (555) 123-4567 · 07934 524 911
	{"phone", regexp.MustCompile(
		`(?:\+\d[\d\s().-]{7,}\d)|(?:\(\d{3}\)\s*\d{3}[-\s]?\d{3,4})|(?:\b0\d{3,4}[\s-]?\d{3}[\s-]?\d{3,4}\b)`)},
	{"email", regexp.MustCompile(`\b[\w.+-]+@[\w-]+\.[A-Za-z]{2,}\b`)},
	// a bare hostname with a real TLD — one site's domain as every site's default
	{"domain", regexp.MustCompile(`\b(?:[\w-]+\.)+(?:com|co\.uk|uk|org|net|io|ai|dev)\b`)},
	{"address", regexp.MustCompile(
		`\b\d+[a-zA-Z]?\s+[A-Z][\w'-]*\s+(?:St|Street|Ave|Avenue|Rd|Road|Ln|Lane|Way|Close|Drive|Dr|Blvd)\b`)},
	{"price", regexp.MustCompile(`[£$€]\s?\d`)},
	// Two forms. A day name is NOT required to state opening hours, and relying on
	// one was the Python lint's first false NEGATIVE: the control "Weekdays 8am to
	// 5pm" — a plain fabrication — slipped a day-name-anchored pattern. Only a
	// control the corpus did not contain could have caught that.
	{"opening_hours", regexp.MustCompile(`(?i)` +
		// a time RANGE: "9am – 6pm", "8am to 5pm", "08:00-17:00"
		`(?:\b\d{1,2}(?::\d{2})?\s*(?:am|pm)\b.{0,20}?\b\d{1,2}(?::\d{2})?\s*(?:am|pm)\b)` +
		`|` +
		// or a day/period word carrying a time
		`(?:\b(?:mon|tue|wed|thu|fri|sat|sun|weekday|weekend|daily)[a-z]*\b` +
		`[^|]{0,40}?\b\d{1,2}\s*(?:am|pm|:\d{2}))`)},
}

// Shapes that are NOT facts even though they carry digits or dots.
var cssDeclRe = regexp.MustCompile(`(?:^|;)\s*(?:--)?[a-z-]+\s*:\s*\S`)
var bareDestRe = regexp.MustCompile(`^[#/]`)

var attrValues = map[string]bool{
	"lazy": true, "eager": true, "false": true, "true": true,
	"_blank": true, "_self": true, "none": true, "auto": true,
}

// flattenTemplateText strips tags and collapses whitespace, so a literal is judged
// on what a reader would SEE rather than on its markup.
func flattenTemplateText(s string) string {
	return strings.TrimSpace(whitespaceRe.ReplaceAllString(htmlTagRe.ReplaceAllString(s, " "), " "))
}

// classifyFallbackLiteral returns the fact-shape a literal asserts, or "" when it
// is a legitimate label.
//
// The exclusions each cost something to learn and are not arbitrary:
//   - a bare path or fragment is a DESTINATION, and belongs to
//     scripts/check_cta_gates.py's PLACEHOLDER class. Two checks reporting one
//     finding teaches people to ignore both.
//   - a CSS declaration is a default colour, not a claim about the business.
//   - "Contact us for pricing" is deliberately NOT a price finding: it states no
//     figure. An honest non-claim is the RIGHT thing for a component to say when it
//     has no datum, and refusing it would push authors back toward inventing one.
func classifyFallbackLiteral(literal string) string {
	text := flattenTemplateText(literal)

	if len(text) < 3 {
		return "" // '.', '#', punctuation
	}
	if attrValues[strings.ToLower(text)] {
		return ""
	}
	if bareDestRe.MatchString(text) {
		return ""
	}
	if cssDeclRe.MatchString(text) {
		return ""
	}
	for _, fs := range factShapes {
		if fs.re.MatchString(text) {
			return fs.name
		}
	}
	return ""
}

// FabricatedFallback is one literal that would be substituted for an absent datum.
type FabricatedFallback struct {
	Shape   string // phone / email / domain / address / price / opening_hours
	Literal string
}

// fabricatedFallbacks returns every fact-shaped {{else}} literal in a template.
// Pure: no DB, no I/O, same input → same output, so the rule stays directly
// testable against real component rows.
func fabricatedFallbacks(htmlTemplate string) []FabricatedFallback {
	var out []FabricatedFallback
	seen := map[string]bool{}

	for _, m := range fallbackBlockRe.FindAllStringSubmatch(htmlTemplate, -1) {
		ifBranch, literal := m[1], m[2]
		if strings.Contains(literal, "{{") {
			continue // a nested action, not a literal fallback
		}
		flat := flattenTemplateText(literal)
		if flat == "" {
			continue
		}
		// A fallback rendering the SAME text as the branch it replaces invents
		// nothing — it is one constant rendered two ways, commonly "link it if we
		// have a URL, otherwise print it". about-commercial-block's builder
		// attribution is exactly this, and reporting it was the Python lint's first
		// false positive, caught by reading the template rather than believing the
		// tool.
		if strings.Contains(flattenTemplateText(ifBranch), flat) {
			continue
		}
		shape := classifyFallbackLiteral(literal)
		if shape == "" {
			continue
		}
		key := shape + "\x00" + flat
		if seen[key] {
			continue // the same literal twice (href and text) is one defect
		}
		seen[key] = true
		out = append(out, FabricatedFallback{Shape: shape, Literal: flat})
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Shape != out[j].Shape {
			return out[i].Shape < out[j].Shape
		}
		return out[i].Literal < out[j].Literal
	})
	return out
}

// fabricatedFallbackIssue returns a blocking-issue string when a template would
// substitute a business fact for an absent datum, or "" when it would not.
//
// Phrased so the refusal tells the author what to do instead: the fix is always to
// gate the element on its own datum and DELETE the literal, which is what the
// schema's on_missing:"skip_field" already asks for.
func fabricatedFallbackIssue(htmlTemplate string) string {
	found := fabricatedFallbacks(htmlTemplate)
	if len(found) == 0 {
		return ""
	}
	parts := make([]string, 0, len(found))
	for _, f := range found {
		parts = append(parts, fmt.Sprintf("%s %q", f.Shape, f.Literal))
	}
	return fmt.Sprintf(
		"template invents %d business fact(s) when the site's datum is absent — %s; "+
			"a component must not assert a contact detail, price or address nobody supplied. "+
			"Gate the element on its own datum ({{if .field}}…{{end}}) and delete the fallback, "+
			"which is what the schema's on_missing:\"skip_field\" already specifies (bugs_open/140)",
		len(found), strings.Join(parts, ", "))
}

// fabricatedFallbackRegression is the COMPARATIVE form, for the update/repair
// paths — and it is comparative for the same reason every check in
// component_write_guard.go is: an ABSOLUTE fabrication gate on a repair path would
// refuse a legitimate repair to exactly the components most likely to need one.
//
// A colour fix or a nav-link fix on a template that already fabricates would be
// blocked by the absolute form, trapping that component permanently — no rewrite
// could ever land to improve it. So this refuses only what the replacement
// INTRODUCES: a fabricated fact that the row it replaces did not already carry.
//
// Raised by the council's bug_historian and guardian seats on
// 19bee790-ea55-46eb-9f39-c985ecf8bd56, both asking the same checkable question —
// is store_generated_component really the only writer of html_template? It is not.
// The census, taken 2026-08-03 (grep for INSERT INTO / UPDATE content_components):
//
//	store_generated_component  INSERT + UPDATE  <- the absolute gate (birth)
//	update_component_html      UPDATE           <- this comparative gate
//	create_tool_component      INSERT           } no fabrication gate; tool
//	deploy_tool_action         INSERT           } components, tracked in RFC_009
//	fix_component_template     UPDATE x2        } as remaining coverage
//	fix_harcoded_colours       UPDATE           } (narrow, non-LLM repairs that
//	fix_forced_text_colours    UPDATE           } rewrite style, not content)
//	fix_nav_link_templates     UPDATE           }
//	core-manager admin handler UPDATE           <- human-driven, deliberately ungated
//
// So the door is NOT fully closed by the write path alone, and claiming otherwise
// would be false. The daily lint (CGV-029) is what covers the remainder — it reads
// the LIVE library, so it sees a fabrication whichever writer introduced it. Gate
// where it is sound, report everywhere.
func fabricatedFallbackRegression(currentHTML, newHTML string) string {
	introduced := fabricatedFallbacks(newHTML)
	if len(introduced) == 0 {
		return ""
	}
	existing := map[string]bool{}
	for _, f := range fabricatedFallbacks(currentHTML) {
		existing[f.Shape+"\x00"+f.Literal] = true
	}

	var novel []string
	for _, f := range introduced {
		if !existing[f.Shape+"\x00"+f.Literal] {
			novel = append(novel, fmt.Sprintf("%s %q", f.Shape, f.Literal))
		}
	}
	if len(novel) == 0 {
		return "" // already present before this write — not this write's doing
	}
	return fmt.Sprintf(
		"replacement INTRODUCES %d fabricated business fact(s) the current template does not have — %s; "+
			"a component must not assert a contact detail, price or address nobody supplied. "+
			"Gate the element on its own datum ({{if .field}}…{{end}}) instead of substituting a literal (bugs_open/140)",
		len(novel), strings.Join(novel, ", "))
}
