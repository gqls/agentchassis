// FILE: platform/orchestration/actions/experience_criteria.go
//
// Write-time validation for experience-register criteria templates
// (migration 218; design:
// docs/agent_docs/docs024_key_docs_latest/experience_register/design/criteria_template_schema_v1.md).
//
// The owner's ruling of 2026-07-24 was that approval is per experience and that
// "the acceptance-test side must be formalised". This file is that
// formalisation. It exists because of three live failure classes, all sighted
// before it was written:
//
//   1. Criteria are parsed at READ time only, so a template that asserts markup
//      the page no longer ships, or that carries an unfilled "-EDIT"
//      placeholder, looks green while asserting nothing.
//   2. A template can name a check the platform cannot execute. This is not
//      hypothetical: the council-APPROVED EXPERIENCE_PLAN for the vonc gauntlet
//      contains two interaction checks that assert 300 ms after a click while
//      the API they wait on takes 8–23 s. They parse perfectly. They would fail
//      a correct page, and the tempting repair — painting placeholder text —
//      would make them pass with the engine switched off.
//   3. A base entry that carries a concrete URL or a site-specific value is the
//      bugs_closed/045 class: a static fallback re-applied on every render,
//      which a per-site binding cannot override.
//
// So validation has three jobs, in this order: the document parses; every
// placeholder closes against the entry's binding_schema; and every check is
// either executable by a named tier or explicitly marked deferred. A check the
// platform cannot run is NOT an error — dropping it silently would be worse, as
// the clause simply disappears from the record — but it must be declared, and
// it is never counted as a pass.

package actions

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// ── what the platform can actually execute ─────────────────────────────────
//
// These tables mirror the two checkers and are held in lockstep with them by
// TestExperienceCheckCapabilities_LockstepWithCheckers, which reads the real
// switch statements out of their source. That test is the point: a capability
// table maintained by hand drifts from the checker within one change, and a
// validator that believes a stale table waves through exactly the checks it
// exists to catch.
//
//   Tier 2 — platform/orchestration/actions/discovery_checks/check_tool_acceptance.go
//            (static, against deployed HTML: anchors, asset references, status)
//   Tier 4 — internal/adapters/browserrunner/run_checks_action.go
//            (a real browser: the same plus overflow, console errors, and the
//            behaviour an interaction actually produces)

// experienceCheckTiers maps a check type to the LOWEST tier that executes it.
var experienceCheckTiers = map[string]int{
	"selector_exists":        2,
	"selector_count":         2,
	"interaction":            2,
	"asset_loads":            2,
	"page_status_ok":         2,
	"attribute_absent":       2,
	"attribute_matches":      2,
	"no_horizontal_overflow": 4,
	"no_console_errors":      4,
}

// experienceCheckFields are the keys either checker reads off ANY check.
// Anything else in a check object is inert — the runner never sees it — so a
// check that carries one is asserting less than it appears to.
var experienceCheckFields = map[string]bool{
	"id": true, "type": true, "selector": true, "path": true,
	"profiles": true, "steps": true, "expect": true, "container": true,
}

// experienceCheckTypeFields are keys read only for a PARTICULAR check type.
//
// They are kept separate from the table above rather than merged into it: an
// `attributes` list means something on `attribute_absent` and nothing at all on
// `selector_exists`, and folding the two together would stop P7 catching the
// second case — which is exactly the "field the runner never reads" defect the
// rule exists for. TestExperienceCheckTypeFields_LockstepWithRunnerStruct holds
// this table against the runner's own decode struct, so a field claimed here
// that the runner cannot decode fails the build.
var experienceCheckTypeFields = map[string]map[string]bool{
	"attribute_absent":  {"attributes": true},
	"attribute_matches": {"attribute": true, "matches": true, "not_matches": true},
}

// experienceStepActions are the interaction steps the browser runner performs.
var experienceStepActions = map[string]bool{"fill": true, "click": true, "select": true}

// experienceExpectFields are the assertion keys. `text_matches` is Tier 4 only:
// Tier 2's expect struct carries a selector and nothing else, so a text
// assertion is confirmed by anchor presence there and by the actual text only
// in a browser.
var experienceExpectFields = map[string]bool{"selector": true, "text_matches": true}

// Documented escape hatches an entry may carry on a check. `_unsupported` (or
// `deferred: true`) declares that the check is beyond the platform today and
// says why; `literal_ok` justifies a literal where a placeholder is expected.
const (
	experienceUnsupportedKey = "_unsupported"
	experienceDeferredKey    = "deferred"
	experienceLiteralOKKey   = "literal_ok"
	experienceInheritedKey   = "_inherited"
	experienceNoteKey        = "note"
	experienceTierKey        = "tier"
)

// experienceMetaFields are keys the register itself understands but the
// checkers do not — they are stripped before a template is handed to a runner.
var experienceMetaFields = map[string]bool{
	experienceUnsupportedKey: true, experienceDeferredKey: true,
	experienceLiteralOKKey: true, experienceInheritedKey: true,
	experienceNoteKey: true, experienceTierKey: true,
}

// ── placeholders ───────────────────────────────────────────────────────────

// experiencePlaceholderRE matches {{binding.name}} and {{binding.name.field}}.
var experiencePlaceholderRE = regexp.MustCompile(`\{\{\s*binding\.([a-zA-Z0-9_]+)(?:\.[a-zA-Z0-9_.]+)?\s*\}\}`)

// experienceLiteralURLRE catches an absolute URL anywhere in a base entry.
var experienceLiteralURLRE = regexp.MustCompile(`https?://`)

// fields whose value must be supplied per site, never fixed in the base entry.
var experienceBoundFields = map[string]bool{"selector": true, "path": true, "url": true}

// ── result ─────────────────────────────────────────────────────────────────

// ExperienceCriteriaIssue names one problem (or one deferral) and where it is.
type ExperienceCriteriaIssue struct {
	CheckID string `json:"check_id,omitempty"`
	Field   string `json:"field,omitempty"`
	Detail  string `json:"detail"`
}

// ExperienceCriteriaValidation is the outcome of validating one entry's
// criteria template against its binding schema.
//
// Executable means "the register's own consumer will actually run this",
// NOT "some tier implements the type". The distinction is the whole value of
// the number: it is what migration 230's approval CHECK rests on, and an entry
// approved on the strength of checks nothing will run is the vacuous-pass
// defect wearing a count. Tier 4 clauses are carried as deferrals with their
// reason, so nothing is lost — see experienceNeedsBrowserReason.
type ExperienceCriteriaValidation struct {
	Errors       []ExperienceCriteriaIssue `json:"errors"`
	Deferred     []ExperienceCriteriaIssue `json:"deferred"`
	Placeholders []string                  `json:"placeholders"`
	Executable   int                       `json:"executable_checks"`
}

// OK reports whether the template may be written.
func (v ExperienceCriteriaValidation) OK() bool { return len(v.Errors) == 0 }

// Summary renders the outcome for a log line or an action result.
func (v ExperienceCriteriaValidation) Summary() string {
	return fmt.Sprintf("%d executable, %d deferred, %d error(s)",
		v.Executable, len(v.Deferred), len(v.Errors))
}

func (v *ExperienceCriteriaValidation) errf(checkID, field, format string, args ...interface{}) {
	v.Errors = append(v.Errors, ExperienceCriteriaIssue{CheckID: checkID, Field: field, Detail: fmt.Sprintf(format, args...)})
}

// ── validation ─────────────────────────────────────────────────────────────

// ValidateExperienceCriteria checks one register entry's criteria template
// against its binding schema. `extra` carries any other jsonb documents whose
// placeholders must also close (the entry's contract, states, data_contract and
// so on) — a placeholder used in a contract clause but never declared is the
// same defect as one used in a check.
//
// Rules, each traceable to something that has gone wrong:
//
//	P1  the document has a non-empty `checks` array
//	P2  every check has a unique, non-empty id and a type
//	P3  every placeholder used is declared in binding_schema (closure)
//	P4  every declared placeholder is used, or is marked {"unused_ok": true}
//	P5  no absolute URL anywhere; selector/path/url values are placeholders
//	    unless the check carries `literal_ok` with a reason
//	P6  a check whose type no tier executes is DEFERRED if it says so, an
//	    error if it does not
//	P7  a check carrying a field no checker reads is DEFERRED if it says so, an
//	    error if it does not — this is what catches `expect_within_ms`,
//	    `retries`, `after`, `equals`, `min`, `visible`
//	P8  interaction steps use only executable actions; expect uses only
//	    executable keys (same deferral rule)
//	P9  a check id ending in -EDIT is rejected outright: Tier 2 silently SKIPS
//	    those, so they read as green while asserting nothing
//	P10 an attribute check names something to assert, and its patterns compile
func ValidateExperienceCriteria(template map[string]interface{}, bindingSchema map[string]interface{}, extra ...interface{}) ExperienceCriteriaValidation {
	var v ExperienceCriteriaValidation

	used := map[string]bool{}
	collectExperiencePlaceholders(template, used)
	for _, doc := range extra {
		collectExperiencePlaceholders(doc, used)
	}
	for name := range used {
		v.Placeholders = append(v.Placeholders, name)
	}
	sort.Strings(v.Placeholders)

	// P3 — closure against the declared schema.
	for name := range used {
		if _, declared := bindingSchema[name]; !declared {
			v.errf("", "binding_schema",
				"placeholder {{binding.%s}} is used but not declared in binding_schema — a fork would have nothing to bind", name)
		}
	}
	// P4 — declared but never used.
	//
	// The first live run of this validator (2026-07-28) refused five of the nine
	// harvested entries here, and the refusals were RIGHT in a way worth writing
	// down: every unused binding named a part of the experience the entry's
	// clauses describe but our checkers cannot assert today — an empty region, an
	// attribute, an ordering, a cross-page navigation, a wait longer than the
	// runner's 300 ms. So an unused binding is the SHADOW of a deferred check.
	//
	// That makes blanket tolerance wrong (the mismatch is real and worth seeing)
	// and a hard error wrong too (deleting the binding would delete the record of
	// what the clause will need the day the harness can run it). It is therefore
	// a DEFERRAL, on the same terms as a deferred check: declared, carried,
	// reported, never counted as a pass — and it must say WHY.
	//
	// The reason is required rather than optional because the previous version
	// accepted a bare `unused_ok: true` while its own error message promised a
	// reason. An escape hatch that does not demand its justification is just a
	// slower way of ignoring the rule.
	for name, decl := range bindingSchema {
		if used[name] {
			continue
		}
		if m, ok := decl.(map[string]interface{}); ok {
			if b, _ := m["unused_ok"].(bool); b {
				reason := strings.TrimSpace(experienceString(m["reason"]))
				if reason == "" {
					v.errf("", "binding_schema",
						"binding %q is marked unused_ok with no reason — say which clause it serves and what stops us asserting it, or the marker is just a way of switching the check off", name)
					continue
				}
				v.Deferred = append(v.Deferred, ExperienceCriteriaIssue{
					CheckID: "binding:" + name,
					Field:   "binding_schema",
					Detail:  reason,
				})
				continue
			}
		}
		v.errf("", "binding_schema",
			"binding %q is declared but never used — every fork would be asked for a value nothing reads (add \"unused_ok\": true with a reason if that is deliberate)", name)
	}

	// P5 — a base entry carries no absolute URLs, anywhere.
	if experienceLiteralURLRE.MatchString(experienceFlatten(template)) {
		v.errf("", "criteria_template",
			"an absolute URL appears in the template — base entries are site-agnostic; concrete addresses belong in the per-site binding (bugs_closed/045: a static fallback re-applies on every render and cannot be overridden)")
	}

	checks, _ := template["checks"].([]interface{})
	if len(checks) == 0 {
		v.errf("", "checks", "criteria template has no checks — an entry whose acceptance side is empty cannot become 'proven'")
		return v
	}

	seen := map[string]bool{}
	for i, raw := range checks {
		ch, ok := raw.(map[string]interface{})
		if !ok {
			v.errf("", fmt.Sprintf("checks[%d]", i), "check is not an object")
			continue
		}
		// A comment-only entry (an inherited-clause marker) is allowed and
		// carries no assertions.
		if _, isMarker := ch[experienceInheritedKey]; isMarker && ch["id"] == nil {
			continue
		}

		id, _ := ch["id"].(string)
		typ, _ := ch["type"].(string)
		if strings.TrimSpace(id) == "" {
			v.errf("", fmt.Sprintf("checks[%d].id", i), "check has no id")
			continue
		}
		if seen[id] {
			v.errf(id, "id", "duplicate check id")
		}
		seen[id] = true

		// P9 — the -EDIT trap.
		if strings.HasSuffix(id, "-EDIT") {
			v.errf(id, "id", "check id ends in -EDIT; Tier 2 SKIPS those silently, so the check would read as green while asserting nothing")
		}
		if strings.TrimSpace(typ) == "" {
			v.errf(id, "type", "check has no type")
			continue
		}

		deferral, deferred := experienceDeferralReason(ch)

		// P6 — is the type executable at all?
		if _, executable := experienceCheckTiers[typ]; !executable {
			if deferred {
				v.Deferred = append(v.Deferred, ExperienceCriteriaIssue{CheckID: id, Field: "type",
					Detail: fmt.Sprintf("check type %q is not executable by any tier: %s", typ, deferral)})
			} else {
				v.errf(id, "type",
					"check type %q is not executable by any tier and is not marked deferred — mark it with %q (say why) or use a type the platform runs: %s",
					typ, experienceUnsupportedKey, experienceKnownTypes())
			}
			continue
		}

		// P7 — fields no checker reads.
		var inert []string
		for key := range ch {
			if experienceCheckFields[key] || experienceMetaFields[key] || experienceCheckTypeFields[typ][key] {
				continue
			}
			inert = append(inert, key)
		}
		sort.Strings(inert)
		if len(inert) > 0 {
			if deferred {
				v.Deferred = append(v.Deferred, ExperienceCriteriaIssue{CheckID: id, Field: strings.Join(inert, ","),
					Detail: fmt.Sprintf("field(s) %s are not read by any checker: %s", strings.Join(inert, ", "), deferral)})
			} else {
				v.errf(id, strings.Join(inert, ","),
					"field(s) %s are not read by any checker, so the check asserts less than it appears to — mark it %q with the reason, or remove them",
					strings.Join(inert, ", "), experienceUnsupportedKey)
			}
			continue
		}

		// P8 — steps and expectations.
		badStep := false
		if steps, ok := ch["steps"].([]interface{}); ok {
			for si, rawStep := range steps {
				st, ok := rawStep.(map[string]interface{})
				if !ok {
					v.errf(id, fmt.Sprintf("steps[%d]", si), "step is not an object")
					badStep = true
					continue
				}
				action, _ := st["action"].(string)
				if !experienceStepActions[action] {
					if deferred {
						v.Deferred = append(v.Deferred, ExperienceCriteriaIssue{CheckID: id, Field: fmt.Sprintf("steps[%d].action", si),
							Detail: fmt.Sprintf("step action %q is not performed by the runner: %s", action, deferral)})
					} else {
						v.errf(id, fmt.Sprintf("steps[%d].action", si),
							"step action %q is not performed by the runner (it does %s) — mark the check %q with the reason if the pattern genuinely needs it",
							action, experienceKnownStepActions(), experienceUnsupportedKey)
					}
					badStep = true
				}
			}
		}
		if exp, ok := ch["expect"].(map[string]interface{}); ok {
			for key := range exp {
				if experienceExpectFields[key] {
					continue
				}
				if deferred {
					v.Deferred = append(v.Deferred, ExperienceCriteriaIssue{CheckID: id, Field: "expect." + key,
						Detail: fmt.Sprintf("expectation %q is not evaluated: %s", key, deferral)})
				} else {
					v.errf(id, "expect."+key,
						"expectation %q is not evaluated by any checker (it reads selector and text_matches) — mark the check %q with the reason",
						key, experienceUnsupportedKey)
				}
				badStep = true
			}
		}

		// P10 — an attribute check must actually assert something.
		//
		// These are ERRORS even when the check is marked deferred. Deferral
		// says the PLATFORM cannot run the check yet; it is not permission to
		// store one that would assert nothing on the day it can. An
		// `attribute_absent` with no attributes, or an `attribute_matches` with
		// no pattern, is the vacuous-pass class written into the register
		// itself — and the runner's honest response to it is a SKIP, which is
		// silent. Refusing it at write time is the only moment it is loud.
		if issue := experienceAttributeShapeIssue(typ, ch); issue != "" {
			v.errf(id, "attribute", "%s", issue)
			badStep = true
		}

		// P5 (per field) — site-specific values must arrive by binding.
		literalOK, _ := ch[experienceLiteralOKKey].(string)
		for field := range experienceBoundFields {
			val, _ := ch[field].(string)
			if val == "" || experiencePlaceholderRE.MatchString(val) || literalOK != "" {
				continue
			}
			v.errf(id, field,
				"%s is a literal (%q) with no binding placeholder — a base entry must not carry a site's own value; bind it, or set %q to the reason it is genuinely universal",
				field, val, experienceLiteralOKKey)
		}

		switch {
		case !deferred && !badStep && experienceHasDeferral(v.Deferred, id):
			// Already recorded as deferred for another reason; do not count it.
		case !deferred && !badStep && experienceNeedsBrowserReason(ch) != "":
			// EXECUTABLE MEANS "THE REGISTER'S CONSUMER WILL ACTUALLY RUN IT".
			//
			// This used to count any check whose type SOME tier implements, and
			// that was a straight disagreement with the runner:
			// splitExperienceChecksByTier holds back everything Tier 4, so the
			// count included precisely the checks that would never execute.
			// CC-001 therefore recorded executable_checks: 2 when only one check
			// could run — which the approval council's honesty seat spotted
			// independently ("only list_exists is unambiguously executable
			// today"), arriving at the same number from the other direction.
			//
			// Nothing runs a browser for the register today, so a Tier 4 check is
			// deferred on exactly the same terms as any other clause we cannot
			// assert: carried, reported, reasoned, never counted as a pass. The
			// day the browser runner is wired in, this stops firing on its own,
			// because both sides now read one function.
			v.Deferred = append(v.Deferred, ExperienceCriteriaIssue{CheckID: id, Field: experienceTierKey,
				Detail: "not executable by the register's consumer: " + experienceNeedsBrowserReason(ch) +
					", and verify_site_experience evaluates Tier 2 only — nothing in the register drives a browser today"})
		case deferred && !experienceHasDeferral(v.Deferred, id):
			// Marked deferred, but nothing about it was actually beyond the
			// platform. Record it as deferred anyway — the marker is the
			// author's claim and the run must honour it rather than quietly
			// counting the check as a pass — and say that it looks executable,
			// because a stray marker is how a real check goes uncounted.
			v.Deferred = append(v.Deferred, ExperienceCriteriaIssue{CheckID: id,
				Detail: "marked deferred (" + deferral + ") although every part of it is executable today — remove the marker if the platform has caught up"})
		case !deferred && !badStep:
			v.Executable++
		}
	}

	return v
}

// experienceNeedsBrowserReason reports why a check can only be run by the
// browser runner, or "" when Tier 2 can run it.
//
// THIS IS THE ONE DEFINITION, deliberately. It was written twice — once in the
// validator's counting and once in the consumer's tier split — and the two
// disagreed, so `executable_checks` counted checks the runner would never
// execute. Two voices can say a check needs a browser: the author, via `tier`,
// and the platform, via the capability table. Either is sufficient.
func experienceNeedsBrowserReason(ch map[string]interface{}) string {
	switch t := ch[experienceTierKey].(type) {
	case float64:
		if t > 2 {
			return fmt.Sprintf("the entry declares tier %d", int(t))
		}
	case string:
		if t == "4" {
			return "the entry declares tier 4"
		}
	}
	if typ := experienceString(ch["type"]); typ != "" {
		if tier, known := experienceCheckTiers[typ]; known && tier > 2 {
			return fmt.Sprintf("check type %q exists only at Tier %d", typ, tier)
		}
	}
	return ""
}

// experienceAttributeShapeIssue reports why an attribute check would assert
// nothing, or "" when it is well formed. The conditions mirror the runner's own
// skip guards in discovery_checks/static_attribute_checks.go one for one — the
// runner skips silently at run time, so this is where the same conditions get
// said out loud.
func experienceAttributeShapeIssue(typ string, ch map[string]interface{}) string {
	switch typ {
	case "attribute_absent":
		attrs, _ := ch["attributes"].([]interface{})
		named := 0
		for _, a := range attrs {
			if strings.TrimSpace(experienceString(a)) != "" {
				named++
			}
		}
		if named == 0 {
			return "attribute_absent names no attributes — the check would match elements and then assert nothing about them; list what must not be present"
		}
	case "attribute_matches":
		if strings.TrimSpace(experienceString(ch["attribute"])) == "" {
			return "attribute_matches names no attribute — say which one is under test"
		}
		matches := strings.TrimSpace(experienceString(ch["matches"]))
		notMatches := strings.TrimSpace(experienceString(ch["not_matches"]))
		if matches == "" && notMatches == "" {
			return "attribute_matches carries neither matches nor not_matches — the check reads the attribute and then asserts nothing about its value"
		}
		for field, pattern := range map[string]string{"matches": matches, "not_matches": notMatches} {
			if pattern == "" {
				continue
			}
			if _, err := regexp.Compile(pattern); err != nil {
				return fmt.Sprintf("%s pattern %q does not compile (%v) — the runner would skip the check, which is silent; a pattern that cannot run must not be stored", field, pattern, err)
			}
		}
	}
	return ""
}

// experienceDeferralReason reports whether a check declares itself beyond the
// platform, and the stated reason.
func experienceDeferralReason(ch map[string]interface{}) (string, bool) {
	if s, ok := ch[experienceUnsupportedKey].(string); ok && strings.TrimSpace(s) != "" {
		return s, true
	}
	if b, ok := ch[experienceDeferredKey].(bool); ok && b {
		if s, ok := ch[experienceNoteKey].(string); ok && strings.TrimSpace(s) != "" {
			return s, true
		}
		return "declared deferred", true
	}
	return "", false
}

// collectExperiencePlaceholders walks any decoded JSON value and records every
// {{binding.name}} it finds.
func collectExperiencePlaceholders(v interface{}, into map[string]bool) {
	switch t := v.(type) {
	case string:
		for _, m := range experiencePlaceholderRE.FindAllStringSubmatch(t, -1) {
			into[m[1]] = true
		}
	case []interface{}:
		for _, e := range t {
			collectExperiencePlaceholders(e, into)
		}
	case map[string]interface{}:
		// Values only — an object's KEYS are field names, never placeholders.
		for _, e := range t {
			collectExperiencePlaceholders(e, into)
		}
	}
}

// experienceFlatten renders every string in a decoded JSON value, for the
// whole-document scans (absolute URLs).
// experienceString renders a value as a string when it is one, so a reason
// written as a number or a bool is treated as absent rather than accepted.
func experienceString(v interface{}) string {
	s, _ := v.(string)
	return s
}

func experienceFlatten(v interface{}) string {
	var b strings.Builder
	var walk func(interface{})
	walk = func(x interface{}) {
		switch t := x.(type) {
		case string:
			b.WriteString(t)
			b.WriteByte('\n')
		case []interface{}:
			for _, e := range t {
				walk(e)
			}
		case map[string]interface{}:
			for _, e := range t {
				walk(e)
			}
		}
	}
	walk(v)
	return b.String()
}

// experienceHasDeferral reports whether this check already contributed a
// deferral, so a check marked deferred for a real reason is not double-counted.
func experienceHasDeferral(issues []ExperienceCriteriaIssue, checkID string) bool {
	for _, is := range issues {
		if is.CheckID == checkID {
			return true
		}
	}
	return false
}

func experienceKnownTypes() string {
	types := make([]string, 0, len(experienceCheckTiers))
	for t := range experienceCheckTiers {
		types = append(types, t)
	}
	sort.Strings(types)
	return strings.Join(types, ", ")
}

func experienceKnownStepActions() string {
	acts := make([]string, 0, len(experienceStepActions))
	for a := range experienceStepActions {
		acts = append(acts, a)
	}
	sort.Strings(acts)
	return strings.Join(acts, ", ")
}
