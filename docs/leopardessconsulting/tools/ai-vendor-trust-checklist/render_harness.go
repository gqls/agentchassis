//go:build ignore

// render_harness.go — the S2 gate for tool-ai-vendor-trust-checklist.
//
// Renders template.html with sample_data.json through html/template exactly as
// the platform render path does, then asserts the properties that matter,
// BEFORE anything is written to the database.
//
// WHY A CROSS-FILE CHECK IS THE POINT OF THIS HARNESS.
// The platform already validates a tool template's id contract:
// deploy_tool_action calls datahelpers.OrphanElementRefs, which finds every
// getElementById/querySelector reference in the page and asserts the ids exist.
// For THIS tool it returns nil, and passes — because the JavaScript lives in a
// separate static asset, so the template contains no references at all. The
// validator is not wrong; it is structurally blind to a contract that spans two
// files. That blindness is exactly how llm-cost-calculator shipped pointing at
// another tool's JS filename and stayed broken. So checks J and F below assert
// the template/JS contract that nothing downstream can see.
//
// EVERY ASSERTION CLASS HAS A MUTANT (--selftest). A green check nobody has
// watched go red is not evidence, so the gate is: baseline all-green AND every
// mutant turns its own check red. Run --selftest, not just the baseline.
//
// Usage:
//   go run render_harness.go template.html sample_data.json tool-ai-vendor-trust-checklist.js
//   go run render_harness.go --selftest template.html sample_data.json tool-...js
package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"regexp"
	"sort"
	"strings"
)

const (
	wantComponent = "tool-ai-vendor-trust-checklist"
	wantScriptSrc = "/tools/assets/tool-ai-vendor-trust-checklist.js"
	wantItems     = 12
)

var (
	reItem        = regexp.MustCompile(`data-vtc-item="(\d+)"`)
	reIDPresent   = regexp.MustCompile(`id="([A-Za-z0-9_-]+)"`)
	reScriptSrc   = regexp.MustCompile(`<script src="([^"]+)"`)
	reDataComp    = regexp.MustCompile(`data-component="([^"]+)"`)
	reWhy         = regexp.MustCompile(`class="vtc-why"`)
	reTierAttr    = regexp.MustCompile(`data-(strong|mid|low)-(label|detail)="([^"]*)"`)
	reGroupTitle  = regexp.MustCompile(`class="vtc-group-title">([^<]+)<`)
	reVerdict     = regexp.MustCompile(`<p class="vtc-verdict" id="vtc-verdict">([^<]*)</p>`)
	reVerdictDet  = regexp.MustCompile(`<p class="vtc-verdict-detail" id="vtc-verdict-detail">([^<]*)</p>`)
	reScoreCount  = regexp.MustCompile(`id="vtc-score-count">([^<]*)<`)
	reScoreTotal  = regexp.MustCompile(`id="vtc-score-total">([^<]*)<`)
	reInitialTier = regexp.MustCompile(`id="vtc-verdict-box" data-tier="([a-z]+)"`)
	reNaBox       = regexp.MustCompile(`<input type="checkbox" id="vtc-na-sector"([^>]*)>`)

	reJSID   = regexp.MustCompile(`querySelector\('#([A-Za-z0-9_-]+)'\)`)
	reJSComp = regexp.MustCompile(`querySelector\('\[data-component="([^"]+)"\]'\)`)
)

// requiredIDs are the elements the tool cannot work without. Listed explicitly
// rather than derived, so deleting one from BOTH files still fails the gate.
var requiredIDs = []string{
	"vtc-score-count", "vtc-score-total", "vtc-verdict", "vtc-verdict-box",
	"vtc-verdict-detail", "vtc-meter-fill", "vtc-reset", "vtc-gaps",
	"vtc-na-sector",
}

type check struct {
	name string
	// run returns (pass, detail). detail ALWAYS reports the value measured, so
	// "measured nothing" is distinguishable from "measured the right thing".
	run func(markup, full, js string) (bool, string)
}

var checks = []check{
	{"B-no-unrendered-actions", func(_, full, _ string) (bool, string) {
		n := strings.Count(full, "{{")
		return n == 0, fmt.Sprintf("%d occurrences of '{{' in rendered output", n)
	}},
	{"C-twelve-items", func(markup, _, _ string) (bool, string) {
		ms := reItem.FindAllStringSubmatch(markup, -1)
		seen := map[string]bool{}
		for _, m := range ms {
			seen[m[1]] = true
		}
		missing := []string{}
		for i := 1; i <= wantItems; i++ {
			n := fmt.Sprintf("%d", i)
			if !seen[n] {
				missing = append(missing, n)
			}
			if !strings.Contains(markup, fmt.Sprintf(`id="vtc-c%d"`, i)) {
				missing = append(missing, fmt.Sprintf("id vtc-c%d", i))
			}
		}
		return len(ms) == wantItems && len(missing) == 0,
			fmt.Sprintf("%d data-vtc-item attributes, %d distinct, missing=%v", len(ms), len(seen), missing)
	}},
	{"D-na-control-is-not-an-item", func(markup, _, _ string) (bool, string) {
		ms := reNaBox.FindAllStringSubmatch(markup, -1)
		if len(ms) != 1 {
			return false, fmt.Sprintf("%d #vtc-na-sector inputs (want 1)", len(ms))
		}
		hasItemAttr := strings.Contains(ms[0][1], "data-vtc-item")
		return !hasItemAttr,
			fmt.Sprintf("1 #vtc-na-sector input; carries data-vtc-item = %v (want false)", hasItemAttr)
	}},
	{"E-one-why-note-per-item", func(markup, _, _ string) (bool, string) {
		n := len(reWhy.FindAllString(markup, -1))
		return n == wantItems, fmt.Sprintf("%d .vtc-why notes for %d items", n, wantItems)
	}},
	{"F-script-src-exact", func(_, full, _ string) (bool, string) {
		ms := reScriptSrc.FindAllStringSubmatch(full, -1)
		if len(ms) != 1 {
			return false, fmt.Sprintf("%d <script src> tags (want exactly 1)", len(ms))
		}
		return ms[0][1] == wantScriptSrc,
			fmt.Sprintf("script src = %q (want %q)", ms[0][1], wantScriptSrc)
	}},
	{"G-data-component-and-class", func(markup, _, _ string) (bool, string) {
		ms := reDataComp.FindAllStringSubmatch(markup, -1)
		if len(ms) != 1 {
			return false, fmt.Sprintf("%d data-component attributes (want 1)", len(ms))
		}
		wantClass := wantComponent + "-section"
		hasClass := strings.Contains(markup, `class="`+wantClass+`"`)
		return ms[0][1] == wantComponent && hasClass,
			fmt.Sprintf("data-component = %q (want %q); section class %q present = %v",
				ms[0][1], wantComponent, wantClass, hasClass)
	}},
	{"H-six-tier-attributes", func(markup, _, _ string) (bool, string) {
		ms := reTierAttr.FindAllStringSubmatch(markup, -1)
		got := map[string]bool{}
		empty := []string{}
		for _, m := range ms {
			k := m[1] + "-" + m[2]
			got[k] = true
			if strings.TrimSpace(m[3]) == "" {
				empty = append(empty, k)
			}
		}
		missing := []string{}
		for _, tier := range []string{"strong", "mid", "low"} {
			for _, part := range []string{"label", "detail"} {
				if !got[tier+"-"+part] {
					missing = append(missing, tier+"-"+part)
				}
			}
		}
		return len(missing) == 0 && len(empty) == 0,
			fmt.Sprintf("%d tier attributes; missing=%v empty=%v", len(ms), missing, empty)
	}},
	{"I-required-ids-present", func(markup, _, _ string) (bool, string) {
		present := map[string]bool{}
		for _, m := range reIDPresent.FindAllStringSubmatch(markup, -1) {
			present[m[1]] = true
		}
		missing := []string{}
		for _, id := range requiredIDs {
			if !present[id] {
				missing = append(missing, id)
			}
		}
		return len(missing) == 0,
			fmt.Sprintf("%d ids in markup; %d required; missing=%v", len(present), len(requiredIDs), missing)
	}},
	{"J-js-template-id-contract", func(markup, _, js string) (bool, string) {
		present := map[string]bool{}
		for _, m := range reIDPresent.FindAllStringSubmatch(markup, -1) {
			present[m[1]] = true
		}
		refs := map[string]bool{}
		for _, m := range reJSID.FindAllStringSubmatch(js, -1) {
			refs[m[1]] = true
		}
		orphans := []string{}
		for id := range refs {
			if !present[id] {
				orphans = append(orphans, id)
			}
		}
		sort.Strings(orphans)

		compOK, compDetail := false, "no data-component query in JS"
		if cm := reJSComp.FindStringSubmatch(js); cm != nil {
			compOK = cm[1] == wantComponent
			compDetail = fmt.Sprintf("JS queries data-component=%q", cm[1])
		}
		return len(refs) > 0 && len(orphans) == 0 && compOK,
			fmt.Sprintf("JS references %d ids, %d orphaned %v; %s", len(refs), len(orphans), orphans, compDetail)
	}},
	{"K-static-markup-is-the-js-zero-state", func(markup, _, _ string) (bool, string) {
		tiers := map[string]string{}
		for _, m := range reTierAttr.FindAllStringSubmatch(markup, -1) {
			tiers[m[1]+"-"+m[2]] = m[3]
		}
		get := func(re *regexp.Regexp) string {
			if m := re.FindStringSubmatch(markup); m != nil {
				return strings.TrimSpace(m[1])
			}
			return ""
		}
		count, total, tier := get(reScoreCount), get(reScoreTotal), get(reInitialTier)
		verdict, detail := get(reVerdict), get(reVerdictDet)
		okCount := count == "0"
		okTotal := total == fmt.Sprintf("%d", wantItems)
		okTier := tier == "low"
		okVerdict := verdict != "" && verdict == tiers["low-label"]
		okDetail := detail != "" && detail == tiers["low-detail"]
		return okCount && okTotal && okTier && okVerdict && okDetail,
			fmt.Sprintf("static count=%q total=%q tier=%q; verdict matches low-label=%v; detail matches low-detail=%v",
				count, total, tier, okVerdict, okDetail)
	}},
	{"L-no-em-dash", func(_, full, _ string) (bool, string) {
		n := strings.Count(full, "—") + strings.Count(full, "&mdash;")
		return n == 0, fmt.Sprintf("%d em-dashes (literal or entity) in rendered output", n)
	}},
	{"M-four-group-titles", func(markup, _, _ string) (bool, string) {
		ms := reGroupTitle.FindAllStringSubmatch(markup, -1)
		got := []string{}
		for _, m := range ms {
			got = append(got, strings.TrimSpace(m[1]))
		}
		want := []string{"Certifications", "Data handling", "Governance and oversight", "Transparency"}
		ok := len(got) == len(want)
		if ok {
			for i := range want {
				if got[i] != want[i] {
					ok = false
				}
			}
		}
		return ok, fmt.Sprintf("%d group titles %v (want %v)", len(got), got, want)
	}},
}

// mutants: name -> (target check, transform). Each must turn its target red.
type mutant struct {
	name   string
	target string
	apply  func(tpl string) string
}

var mutants = []mutant{
	{"literal-braces", "B-no-unrendered-actions", func(t string) string {
		return strings.Replace(t, `{{.badge_label}}`, `{{"{{.badge_label}}"}}`, 1)
	}},
	{"drop-item-12", "C-twelve-items", func(t string) string {
		return strings.Replace(t, ` data-vtc-item="12"`, ``, 1)
	}},
	{"na-becomes-an-item", "D-na-control-is-not-an-item", func(t string) string {
		return strings.Replace(t, `<input type="checkbox" id="vtc-na-sector">`,
			`<input type="checkbox" id="vtc-na-sector" data-vtc-item="13">`, 1)
	}},
	{"drop-a-why-note", "E-one-why-note-per-item", func(t string) string {
		i := strings.Index(t, `<p class="vtc-why">`)
		if i < 0 {
			return t
		}
		j := strings.Index(t[i:], `</p>`)
		if j < 0 {
			return t
		}
		return t[:i] + t[i+j+4:]
	}},
	{"wrong-script-filename", "F-script-src-exact", func(t string) string {
		return strings.Replace(t, wantScriptSrc, "/tools/assets/tool-ai-data-risk-checker.js", 1)
	}},
	{"wrong-data-component", "G-data-component-and-class", func(t string) string {
		return strings.Replace(t, `data-component="`+wantComponent+`"`,
			`data-component="tool-vendor-trust"`, 1)
	}},
	{"drop-a-tier-attribute", "H-six-tier-attributes", func(t string) string {
		re := regexp.MustCompile(` data-strong-detail="[^"]*"`)
		return re.ReplaceAllString(t, "")
	}},
	{"rename-reset-button", "I-required-ids-present", func(t string) string {
		return strings.Replace(t, `id="vtc-reset"`, `id="vtc-reset-button"`, 1)
	}},
	{"rename-score-count", "J-js-template-id-contract", func(t string) string {
		return strings.Replace(t, `id="vtc-score-count"`, `id="vtc-score-counter"`, 1)
	}},
	{"static-total-disagrees", "K-static-markup-is-the-js-zero-state", func(t string) string {
		return strings.Replace(t, `id="vtc-score-total">12<`, `id="vtc-score-total">11<`, 1)
	}},
	{"reintroduce-em-dash", "L-no-em-dash", func(t string) string {
		return strings.Replace(t, `>Transparency<`, `>Transparency &mdash; what they tell people<`, 1)
	}},
	{"drop-a-group-title", "M-four-group-titles", func(t string) string {
		return strings.Replace(t, `<h4 class="vtc-group-title">Transparency</h4>`, ``, 1)
	}},
}

func render(tpl string, data map[string]interface{}) (string, error) {
	t, err := template.New("c").Parse(tpl)
	if err != nil {
		return "", fmt.Errorf("PARSE FAILED: %w", err)
	}
	var out strings.Builder
	if err := t.Execute(&out, data); err != nil {
		return "", fmt.Errorf("EXECUTE FAILED: %w", err)
	}
	return out.String(), nil
}

// markupOf slices the <style> block away. Counting elements across the whole
// document over-counts by exactly the number of CSS rules naming the same
// class — a check that cannot tell a rule from an element measures the wrong
// thing. This trap has been hit twice in the brochure lane.
func markupOf(full string) string {
	if i := strings.Index(full, "</style>"); i >= 0 {
		return full[i+len("</style>"):]
	}
	return full
}

func runAll(tpl, js string, data map[string]interface{}) (map[string]bool, []string, error) {
	full, err := render(tpl, data)
	if err != nil {
		return nil, nil, err
	}
	markup := markupOf(full)
	res := map[string]bool{}
	var lines []string
	for _, c := range checks {
		pass, detail := c.run(markup, full, js)
		res[c.name] = pass
		mark := "PASS"
		if !pass {
			mark = "FAIL"
		}
		lines = append(lines, fmt.Sprintf("  [%s] %-38s %s", mark, c.name, detail))
	}
	return res, lines, nil
}

func main() {
	args := os.Args[1:]
	selftest := false
	if len(args) > 0 && args[0] == "--selftest" {
		selftest, args = true, args[1:]
	}
	if len(args) < 3 {
		fmt.Println("usage: render_harness [--selftest] <template.html> <data.json> <tool.js>")
		os.Exit(2)
	}
	tb, err := os.ReadFile(args[0])
	if err != nil {
		fmt.Println("read template:", err)
		os.Exit(1)
	}
	db, err := os.ReadFile(args[1])
	if err != nil {
		fmt.Println("read data:", err)
		os.Exit(1)
	}
	jb, err := os.ReadFile(args[2])
	if err != nil {
		fmt.Println("read js:", err)
		os.Exit(1)
	}
	var data map[string]interface{}
	if err := json.Unmarshal(db, &data); err != nil {
		fmt.Println("parse data:", err)
		os.Exit(1)
	}
	tpl, js := string(tb), string(jb)

	fmt.Printf("BASELINE — %d checks over %d bytes of template, %d bytes of JS\n",
		len(checks), len(tpl), len(js))
	res, lines, err := runAll(tpl, js, data)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for _, l := range lines {
		fmt.Println(l)
	}
	failed := []string{}
	for _, c := range checks {
		if !res[c.name] {
			failed = append(failed, c.name)
		}
	}
	if len(failed) > 0 {
		fmt.Printf("\nBASELINE RED: %d of %d checks failed: %v\n", len(failed), len(checks), failed)
		os.Exit(1)
	}
	fmt.Printf("BASELINE GREEN: %d of %d checks passed\n", len(checks), len(checks))

	if !selftest {
		fmt.Println("\nNOTE: --selftest NOT run. A green baseline alone is not the S2 gate;")
		fmt.Println("every assertion class needs a mutant that turns it red.")
		return
	}

	fmt.Printf("\nSELFTEST — %d mutants, each must turn its own check RED\n", len(mutants))
	covered := map[string]bool{}
	bad := 0
	for _, mu := range mutants {
		mutated := mu.apply(tpl)
		if mutated == tpl {
			fmt.Printf("  [ERROR] %-24s mutant did not change the template (stale pattern)\n", mu.name)
			bad++
			continue
		}
		mres, _, merr := runAll(mutated, js, data)
		if merr != nil {
			// A parse/execute failure is a legitimate red for its target check.
			fmt.Printf("  [RED  ] %-24s -> %-38s (render refused: %v)\n", mu.name, mu.target, merr)
			covered[mu.target] = true
			continue
		}
		if mres[mu.target] {
			fmt.Printf("  [ERROR] %-24s -> %-38s STILL GREEN — the check cannot detect this\n",
				mu.name, mu.target)
			bad++
			continue
		}
		fmt.Printf("  [RED  ] %-24s -> %s\n", mu.name, mu.target)
		covered[mu.target] = true
	}

	uncovered := []string{}
	for _, c := range checks {
		if !covered[c.name] {
			uncovered = append(uncovered, c.name)
		}
	}
	fmt.Printf("\n%d of %d checks have a mutant that turns them red\n", len(covered), len(checks))
	if len(uncovered) > 0 {
		fmt.Printf("UNMUTATED (not yet evidence): %v\n", uncovered)
	}
	if bad > 0 || len(uncovered) > 0 {
		fmt.Println("S2 GATE: NOT SATISFIED")
		os.Exit(1)
	}
	fmt.Println("S2 GATE: SATISFIED — baseline green and every check proven able to fail")
}
