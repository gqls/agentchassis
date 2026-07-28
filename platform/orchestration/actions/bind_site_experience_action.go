// FILE: platform/orchestration/actions/bind_site_experience_action.go
//
// bind_site_experience — fork a register entry onto one site's real pages and
// selectors. Requires migration 218 (site_experiences).
//
// THE POINT OF THE SPLIT. A base entry is site-agnostic by construction: it names
// a destination ROLE, never a URL, and a selector PLACEHOLDER, never a selector.
// That is not tidiness — a concrete value baked into a shared base re-applies on
// every render and cannot be overridden per site (bugs_closed/045). Everything
// concrete lives here, in the fork, which is data.
//
// So this action is where the promises stop being abstract. It is the last point
// at which "this control leads somewhere real" is still a claim about a template;
// after it, the claim is about a named page on a named site, and can be checked.
//
// WHAT IT REFUSES, AND WHY EACH ONE IS A DOOR RATHER THAN A WARNING
//
//  1. AN UNCLOSED BINDING. Every placeholder the entry uses must be bound, and
//     every binding supplied must be used. The first half is obvious. The second
//     is the one that rots: a binding nobody reads looks like configuration and
//     is actually a lie about what the fork asserts.
//
//  2. AN EMPTY VALUE. A bound-but-empty selector is worse than an unbound one,
//     because it satisfies closure. `{{binding.x}}` resolving to "" produces a
//     check against the empty selector, which is the shape of a check that
//     cannot fail.
//
//  3. A SELECTOR WITH NO ANCHOR. Tier 2 confirms a check by finding the
//     selector's leftmost token in the deployed HTML; a selector it cannot parse
//     an anchor out of is SKIPPED, and a skipped check reads as green. This is
//     the same family as the `-EDIT` ids the write path rejects. Refusing here
//     means the skip is impossible rather than merely unlikely.
//
//  4. A PAGE ROLE THAT RESOLVES TO NOTHING. A destination role bound to a page
//     that does not exist on this site is precisely the defect this register was
//     built to end (bugs_open/023, bugs_open/071, and the four dead carousel
//     destinations found by hand on 2026-07-26). Checked against `pages` at bind
//     time, so a dead promise cannot be recorded as a live one.
//
// WHAT IT DOES NOT DO. It does not run the criteria — that is the consumer's job,
// and only a green run may write `verified`. It does not approve anything. And it
// will not let a fork of a DRAFT entry claim to be bound: a draft is visible but
// unselectable, so binding one is a proposal, not a commitment.
//
// Registration (registry.go):
//
//	"bind_site_experience": {
//	    Handler:     BindSiteExperienceAction,
//	    Category:    "experience_register",
//	    Description: "Bind a register entry to one site's real pages and selectors",
//	    IsLocal:     true,
//	}
package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var BindSiteExperienceInputSpec = datahelpers.ActionInputSpec{
	Optional: []string{
		"pattern_name", "pattern_name_field", "bindings_field",
		"instance_key", "site_id_field", "created_by",
	},
	Defaults: map[string]interface{}{
		"bindings_field":     "experience_bindings",
		"pattern_name_field": "input_data.pattern_name",
		"site_id_field":      "input_data.site_id",
		"instance_key":       "default",
	},
}

func init() {
	datahelpers.RegisterActionInputSpec("bind_site_experience", BindSiteExperienceInputSpec)
}

// experienceAnchorRE mirrors `anchorRe` in
// discovery_checks/check_tool_acceptance.go — the leftmost simple selector
// token, which is the ONLY thing Tier 2 can confirm. Held in lockstep by
// TestExperienceAnchorRE_MatchesTheCheckersOwn: if the checker's idea of an
// anchor changes and this does not, we would accept selectors it then skips,
// and a skipped check reads as green.
var experienceAnchorRE = regexp.MustCompile(`^\s*([#.]?[A-Za-z][A-Za-z0-9_-]*)`)

// experienceBindingKinds are the declared types a binding_schema entry may
// carry, READ FROM THE HARVESTED ENTRIES rather than chosen — the corpus is the
// authority on its own vocabulary, and this list was wrong the first time
// because I picked plausible names instead of counting.
// TestExperienceBindingKinds_CoverTheHarvestedCorpus fails if an entry declares
// a type this map does not know.
//
//	selector   49   a CSS selector; must carry a Tier-2-anchorable leftmost token
//	value      17   a literal the check types or compares
//	page        3   must name a real page of the site (checked against `pages`)
//	asset_path  3   a path to a served asset (feed JSON, image); not a page
//	url_param   1   a query/fragment parameter name
var experienceBindingKinds = map[string]bool{
	"selector": true, "value": true, "page": true, "asset_path": true, "url_param": true,
}

func BindSiteExperienceAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "bind_site_experience"))
	if params.ExecutionContext != nil && params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	config := params.StepConfig.Config

	siteID := datahelpers.GetStringField(config, "site_id", "")
	if siteID == "" {
		siteID = datahelpers.ExtractNestedFieldString(params.CollectedData,
			datahelpers.GetStringField(config, "site_id_field", "input_data.site_id"))
	}
	if siteID == "" {
		return nil, fmt.Errorf("bind_site_experience: no site_id")
	}

	patternName := datahelpers.GetStringField(config, "pattern_name", "")
	if patternName == "" {
		patternName = datahelpers.ExtractNestedFieldString(params.CollectedData,
			datahelpers.GetStringField(config, "pattern_name_field", "input_data.pattern_name"))
	}
	if patternName == "" {
		return nil, fmt.Errorf("bind_site_experience: no pattern_name")
	}

	bindings := datahelpers.ExtractNestedFieldMap(params.CollectedData,
		datahelpers.GetStringField(config, "bindings_field", "experience_bindings"))
	if len(bindings) == 0 {
		return nil, fmt.Errorf("bind_site_experience: no bindings supplied for %q", patternName)
	}

	instanceKey := datahelpers.GetStringField(config, "instance_key", "default")

	// ── load the entry ─────────────────────────────────────────────────────
	var (
		patternStatus string
		schemaJSON    []byte
		docsJSON      []byte
	)
	err := params.DB.QueryRowContext(ctx, `
		SELECT status,
		       binding_schema,
		       jsonb_build_object(
		         'criteria_template', criteria_template,
		         'contract', contract,
		         'states', states,
		         'automatic_triggers', automatic_triggers,
		         'data_contract', data_contract,
		         'degraded_states', degraded_states,
		         'entry_points', entry_points,
		         'destination_roles', destination_roles,
		         'section_types', section_types,
		         'honesty_clauses', honesty_clauses,
		         'latency_envelope', latency_envelope)
		FROM experience_patterns WHERE name = $1`, patternName).
		Scan(&patternStatus, &schemaJSON, &docsJSON)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("bind_site_experience: no register entry named %q", patternName)
	}
	if err != nil {
		return nil, fmt.Errorf("bind_site_experience: loading %q: %w", patternName, err)
	}

	schema := map[string]interface{}{}
	if err := json.Unmarshal(schemaJSON, &schema); err != nil {
		return nil, fmt.Errorf("bind_site_experience: decoding binding_schema: %w", err)
	}
	var docs interface{}
	if err := json.Unmarshal(docsJSON, &docs); err != nil {
		return nil, fmt.Errorf("bind_site_experience: decoding entry documents: %w", err)
	}

	// The placeholders the entry ACTUALLY uses, read from its stored documents
	// rather than from its schema — the schema is what it says it needs, the
	// documents are what it really reads, and the two can disagree.
	used := map[string]bool{}
	collectExperiencePlaceholders(docs, used)

	problems := validateExperienceBindings(used, schema, bindings)

	// ── page roles must resolve to real pages of THIS site ─────────────────
	pageBindings := experiencePageBindings(schema, bindings)
	if len(pageBindings) > 0 {
		missing, err := missingSitePages(ctx, params.DB, siteID, pageBindings)
		if err != nil {
			return nil, fmt.Errorf("bind_site_experience: resolving pages: %w", err)
		}
		for _, m := range missing {
			problems = append(problems, fmt.Sprintf(
				"binding %q names page %q, which does not exist on this site — binding a promise to a page that is not there is the defect this register exists to end (bugs_open/023)",
				m.binding, m.value))
		}
	}

	if len(problems) > 0 {
		sort.Strings(problems)
		logger.Warn("bind_site_experience: REFUSED",
			zap.String("site_id", siteID), zap.String("pattern", patternName),
			zap.Strings("problems", problems))
		return nil, fmt.Errorf("bind_site_experience: refused, %d problem(s): %s",
			len(problems), strings.Join(problems, "; "))
	}

	// A draft entry is visible but UNSELECTABLE, so a fork of one is a proposal
	// rather than a commitment. Recording it as `bound` would let an unapproved
	// contract look like a live one.
	status := "bound"
	if patternStatus == "draft" {
		status = "proposed"
	}

	bindingsJSON, err := json.Marshal(bindings)
	if err != nil {
		return nil, fmt.Errorf("bind_site_experience: marshalling bindings: %w", err)
	}

	var forkID string
	err = params.DB.QueryRowContext(ctx, `
		INSERT INTO site_experiences (site_id, pattern_name, instance_key, bindings, status)
		VALUES ($1, $2, $3, $4::jsonb, $5)
		ON CONFLICT (site_id, pattern_name, instance_key) DO UPDATE
		SET bindings = EXCLUDED.bindings,
		    status = EXCLUDED.status,
		    updated_at = now(),
		    -- a re-bind invalidates the previous run: whatever was verified was
		    -- verified against the OLD bindings.
		    last_checked_at = NULL,
		    last_check_result = NULL
		RETURNING id`,
		siteID, patternName, instanceKey, string(bindingsJSON), status).Scan(&forkID)
	if err != nil {
		return nil, fmt.Errorf("bind_site_experience: upsert: %w", err)
	}

	logger.Info("bind_site_experience: bound",
		zap.String("site_id", siteID),
		zap.String("pattern", patternName),
		zap.String("instance_key", instanceKey),
		zap.String("status", status),
		zap.String("pattern_status", patternStatus),
		zap.Int("bindings", len(bindings)))

	return map[string]interface{}{
		"site_experience_id": forkID,
		"pattern_name":       patternName,
		"instance_key":       instanceKey,
		"status":             status,
		"pattern_status":     patternStatus,
		"bound_count":        len(bindings),
	}, nil
}

// validateExperienceBindings reports every problem at once, for the same reason
// the write path does: each refusal round trip back to a generating agent is an
// LLM call.
func validateExperienceBindings(used map[string]bool, schema, bindings map[string]interface{}) []string {
	var problems []string

	for name := range used {
		if _, ok := bindings[name]; !ok {
			problems = append(problems, fmt.Sprintf(
				"placeholder {{binding.%s}} is used by the entry but not bound — the check would run against the literal placeholder text, which is the -EDIT failure mode: it runs, it looks green, it asserts nothing", name))
		}
	}

	for name := range bindings {
		if !used[name] {
			// A binding the entry never reads. Not fatal to correctness, but it
			// is a claim about what this fork asserts, and it is false.
			problems = append(problems, fmt.Sprintf(
				"binding %q is supplied but the entry never reads it — a fork that carries values nothing uses misrepresents what it asserts", name))
		}
	}

	for name, raw := range bindings {
		value := strings.TrimSpace(experienceString(raw))
		if value == "" {
			problems = append(problems, fmt.Sprintf(
				"binding %q is empty — an empty selector satisfies closure and produces a check that cannot fail, which is worse than leaving it unbound", name))
			continue
		}

		kind := experienceBindingKind(schema, name)
		if kind != "" && !experienceBindingKinds[kind] {
			problems = append(problems, fmt.Sprintf(
				"binding %q declares unknown type %q in the entry's binding_schema", name, kind))
		}

		if kind == "selector" {
			if experienceAnchorRE.FindStringSubmatch(value) == nil {
				problems = append(problems, fmt.Sprintf(
					"binding %q = %q has no leftmost #id/.class/tag anchor — Tier 2 SKIPS a selector it cannot anchor, and a skipped check reads as green", name, value))
			}
			if strings.Contains(value, "://") {
				problems = append(problems, fmt.Sprintf(
					"binding %q is typed selector but looks like a URL: %q", name, value))
			}
		}

		if strings.Contains(value, "{{") {
			problems = append(problems, fmt.Sprintf(
				"binding %q still contains a placeholder: %q — a fork is where placeholders END", name, value))
		}
	}

	return problems
}

func experienceBindingKind(schema map[string]interface{}, name string) string {
	decl, ok := schema[name].(map[string]interface{})
	if !ok {
		return ""
	}
	return experienceString(decl["type"])
}

type experiencePageBinding struct{ binding, value string }

// experiencePageBindings returns the bindings the entry's schema types as
// `page`, which are the ones that must name something real.
func experiencePageBindings(schema, bindings map[string]interface{}) []experiencePageBinding {
	var out []experiencePageBinding
	for name, raw := range bindings {
		if experienceBindingKind(schema, name) != "page" {
			continue
		}
		if v := strings.TrimSpace(experienceString(raw)); v != "" {
			out = append(out, experiencePageBinding{binding: name, value: v})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].binding < out[j].binding })
	return out
}

// missingSitePages returns the page bindings that match no page of this site.
// A page may be named by url or by name — a binding written either way resolves,
// because which one an author reaches for is not a thing worth failing over.
func missingSitePages(ctx context.Context, db *sql.DB, siteID string, want []experiencePageBinding) ([]experiencePageBinding, error) {
	if db == nil || len(want) == 0 {
		return nil, nil
	}
	values := make([]string, 0, len(want))
	for _, w := range want {
		values = append(values, w.value)
	}
	rows, err := db.QueryContext(ctx, `
		SELECT url, name FROM pages
		WHERE site_id = $1::uuid AND (url = ANY($2) OR name = ANY($2))`,
		siteID, pqStringArray(values))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	present := map[string]bool{}
	for rows.Next() {
		var url, name string
		if err := rows.Scan(&url, &name); err != nil {
			return nil, err
		}
		present[url] = true
		present[name] = true
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	var missing []experiencePageBinding
	for _, w := range want {
		if !present[w.value] {
			missing = append(missing, w)
		}
	}
	return missing, nil
}
