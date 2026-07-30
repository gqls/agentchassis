//go:build ignore

// fence_check.go — the S1 gate for tool-ai-vendor-trust-checklist.
//
// Extracts the ```criteria fence from a PLAN and validates it against the
// platform's OWN capability tables, copied from
// platform/orchestration/actions/experience_criteria.go. Those tables are held
// in lockstep with the two checkers by
// TestExperienceCheckCapabilities_LockstepWithCheckers, which reads the real
// switch statements out of their source — so they are the authority, not a
// second opinion.
//
// WHY NOT CALL ValidateExperienceCriteria DIRECTLY. That validator is for
// EXPERIENCE REGISTER entries: its rules P3/P4/P5 require every selector to be
// a {{binding.*}} placeholder declared in a binding_schema, because a register
// entry is a fleet-wide template that each site forks. A tool PLAN fence is the
// opposite — one tool, one site, literal selectors, no binding_schema (see
// smart-contrast, the pilot that passed first complete run). Running it here
// would report a wall of spurious placeholder errors and teach the next author
// to ignore the tool. So this reuses its TABLES and drops its register-only
// rules, and says so rather than implying the fence went through the register
// validator.
//
// R7 IS NOT FROM THAT FILE. It is this build's own finding: interaction checks
// share one browser page and accumulate state, so a fence whose interactions do
// not each reset first is order-dependent and its later claims are not what
// they appear to be.
//
// Usage: go run fence_check.go PLAN_doc.md
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// ── tables mirrored from experience_criteria.go (2026-07-30) ────────────────

var checkTiers = map[string]int{
	"selector_exists": 2, "selector_count": 2, "interaction": 2,
	"asset_loads": 2, "page_status_ok": 2,
	"attribute_absent": 2, "attribute_matches": 2,
	"no_horizontal_overflow": 4, "no_console_errors": 4,
	"has_visible_area": 4,
}

var checkFields = map[string]bool{
	"id": true, "type": true, "selector": true, "path": true,
	"profiles": true, "steps": true, "expect": true, "container": true,
}

var checkTypeFields = map[string]map[string]bool{
	"attribute_absent":  {"attributes": true},
	"attribute_matches": {"attribute": true, "matches": true, "not_matches": true},
	"has_visible_area":  {"min_width": true, "min_height": true},
}

var stepActions = map[string]bool{"fill": true, "click": true, "select": true}
var expectFields = map[string]bool{"selector": true, "text_matches": true}
var stepFields = map[string]bool{"action": true, "selector": true, "value": true}

const resetSelector = "#vtc-reset"

type issue struct {
	rule   string
	detail string
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: fence_check <PLAN.md>")
		os.Exit(2)
	}
	b, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Println("read plan:", err)
		os.Exit(1)
	}
	body := string(b)

	// Extract the fence.
	start := strings.Index(body, "```criteria")
	if start < 0 {
		fmt.Println("R0 FAIL: no ```criteria fence in the PLAN")
		os.Exit(1)
	}
	rest := body[start+len("```criteria"):]
	end := strings.Index(rest, "```")
	if end < 0 {
		fmt.Println("R0 FAIL: unterminated ```criteria fence")
		os.Exit(1)
	}
	raw := strings.TrimSpace(rest[:end])

	var doc map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &doc); err != nil {
		fmt.Printf("R0 FAIL: fence is not valid JSON: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("R0 PASS: fence extracted and parsed (%d bytes of JSON)\n", len(raw))

	var issues []issue
	bad := func(rule, format string, args ...interface{}) {
		issues = append(issues, issue{rule, fmt.Sprintf(format, args...)})
	}

	checksAny, _ := doc["checks"].([]interface{})
	if len(checksAny) == 0 {
		fmt.Println("R1 FAIL: fence has no non-empty checks array")
		os.Exit(1)
	}

	seen := map[string]bool{}
	tier2, tier4, interactions, resetFirst := 0, 0, 0, 0

	for i, ca := range checksAny {
		ch, ok := ca.(map[string]interface{})
		if !ok {
			bad("R1", "check %d is not an object", i)
			continue
		}
		id, _ := ch["id"].(string)
		typ, _ := ch["type"].(string)
		where := fmt.Sprintf("check %d (%q)", i, id)

		// R1 — unique, non-empty id and type.
		if strings.TrimSpace(id) == "" {
			bad("R1", "%s has an empty id", where)
		} else if seen[id] {
			bad("R1", "%s has a duplicate id", where)
		}
		seen[id] = true
		if strings.TrimSpace(typ) == "" {
			bad("R1", "%s has no type", where)
			continue
		}

		// R2 — the type must be one a checker executes. An unknown type is
		// SKIPPED by the runner, not failed, and an all-skipped result set
		// reads as a PASS plus a 7-day cooldown.
		tier, known := checkTiers[typ]
		if !known {
			bad("R2", "%s type %q is executed by NO tier: it would be SKIPPED, and a skip reads as a pass", where, typ)
		} else if tier == 4 {
			tier4++
		} else {
			tier2++
		}

		// R3 — no -EDIT ids: the runner skips them silently.
		if strings.HasSuffix(id, "-EDIT") {
			bad("R3", "%s id ends in -EDIT, which the runner SKIPS while it reads as green", where)
		}

		// R4 — only fields a checker reads.
		for k := range ch {
			if checkFields[k] {
				continue
			}
			if allowed, ok := checkTypeFields[typ]; ok && allowed[k] {
				continue
			}
			owner := ""
			for t, fs := range checkTypeFields {
				if fs[k] {
					owner = t
				}
			}
			if owner != "" {
				bad("R4", "%s carries %q, which the runner reads only on %s: inert here", where, k, owner)
			} else {
				bad("R4", "%s carries %q, which no checker reads: the check asserts less than it appears to", where, k)
			}
		}

		// R5/R6/R7 — interaction internals.
		if typ != "interaction" {
			if _, has := ch["steps"]; has {
				bad("R5", "%s is not an interaction but carries steps", where)
			}
			continue
		}
		interactions++
		steps, _ := ch["steps"].([]interface{})
		if len(steps) == 0 {
			bad("R5", "%s is an interaction with no steps", where)
		}
		for j, sa := range steps {
			st, ok := sa.(map[string]interface{})
			if !ok {
				bad("R5", "%s step %d is not an object", where, j)
				continue
			}
			for k := range st {
				if !stepFields[k] {
					bad("R5", "%s step %d carries %q, which the step struct cannot decode", where, j, k)
				}
			}
			act, _ := st["action"].(string)
			if !stepActions[act] {
				bad("R5", "%s step %d action %q is not one the runner performs (fill|click|select)", where, j, act)
			}
			if sel, _ := st["selector"].(string); strings.TrimSpace(sel) == "" {
				bad("R5", "%s step %d has no selector", where, j)
			}
		}

		// R6 — expect must assert a terminal value with keys the runner reads,
		// and any text_matches must compile as Go RE2.
		exp, _ := ch["expect"].(map[string]interface{})
		if len(exp) == 0 {
			bad("R6", "%s has no expect: steps completing cleanly is a waypoint, not the tool working", where)
		}
		for k, v := range exp {
			if !expectFields[k] {
				bad("R6", "%s expect carries %q, which the runner does not read", where, k)
				continue
			}
			if k == "text_matches" {
				pat, _ := v.(string)
				if _, err := regexp.Compile(pat); err != nil {
					bad("R6", "%s text_matches %q does not compile: %v", where, pat, err)
				}
			}
		}
		if sel, _ := exp["selector"].(string); strings.TrimSpace(sel) == "" {
			bad("R6", "%s expect has no selector", where)
		}

		// R7 — every interaction must reset first. THIS BUILD'S OWN RULE:
		// evaluateOnPage drives every check against ONE page per profile, so
		// state set by an earlier interaction is still present. Without a reset
		// the claim depends on fence order and does not mean what it says.
		firstIsReset := false
		if len(steps) > 0 {
			if st, ok := steps[0].(map[string]interface{}); ok {
				act, _ := st["action"].(string)
				sel, _ := st["selector"].(string)
				firstIsReset = act == "click" && sel == resetSelector
			}
		}
		if firstIsReset {
			resetFirst++
		} else {
			bad("R7", "%s does not begin by clicking %s: interaction checks share one page, so this claim is order-dependent", where, resetSelector)
		}
	}

	// Always report the counts measured, so "asserted nothing" is
	// distinguishable from "asserted the right things".
	fmt.Printf("MEASURED: %d checks — %d executable at tier 2, %d tier-4 only, %d interactions, %d of which reset first\n",
		len(checksAny), tier2, tier4, interactions, resetFirst)

	if len(issues) == 0 {
		fmt.Println("S1 GATE: SATISFIED — every check is executable, every interaction asserts a terminal value and resets first")
		return
	}
	fmt.Printf("\nS1 GATE: NOT SATISFIED — %d issue(s)\n", len(issues))
	for _, is := range issues {
		fmt.Printf("  [%s] %s\n", is.rule, is.detail)
	}
	os.Exit(1)
}
