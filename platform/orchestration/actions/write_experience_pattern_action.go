// FILE: platform/orchestration/actions/write_experience_pattern_action.go
//
// write_experience_pattern — the ONLY way a row enters the experience register.
// Requires migration 218_experience_register_substrate.sql.
//
// The register is a library of small reusable user experiences (component
// contracts and micro-journeys), held once and forked per site. This action
// owns the write, and it is deliberately the only door: migration 218 seeds no
// entries by raw INSERT precisely so that the first rows in the register are
// ones its own contract accepted. A register whose contents never passed its
// own validator has no evidence that the contract is enforceable.
//
// Three integrity rules are worth reading before changing anything here,
// because each closes a door rather than asking an operator to remember
// something:
//
//  1. STATUS IS NOT WRITABLE. A new entry is always 'draft'. Nothing this
//     action accepts can set 'approved' or 'proven'. Approval is a council
//     verdict; 'proven' is the first live green run of bound criteria. Both are
//     things that HAPPENED, and a writer that can assert them turns evidence
//     into a claim. Ignoring a status field silently would be its own trap, so
//     supplying one is an error, not a no-op.
//
//  2. A CONTRACT CHANGE DEMOTES. If an approved or proven entry's contract,
//     criteria, bindings schema or any clause-bearing field changes, the row
//     returns to 'draft'. The approval was of the old contract; carrying it
//     forward would make "approved" mean "was approved of something, once".
//     Cosmetic fields (display_name, description, aka) do not demote — the
//     distinction is explicit in experiencePatternContractFields, so that
//     widening it is a decision rather than an accident.
//
//  3. VALIDATION REFUSES THE WRITE. Criteria are validated here, on the way in,
//     because the live failure class is criteria parsed at READ time only: a
//     template asserting markup the page no longer ships reads green while
//     asserting nothing. Deferred checks are recorded and do NOT block — a
//     check the platform cannot execute is declared, never dropped, and never
//     counted as a pass.
//
// Registration (registry.go):
//
//	"write_experience_pattern": {
//	    Handler:     WriteExperiencePatternAction,
//	    Category:    "experience_register",
//	    Description: "Validate and write one experience-register entry (always as draft)",
//	    IsLocal:     true,
//	}
//
// The travelling doc for an entry is NOT written here. It is a plain
// write_doc_plan step with subject_type 'experience-pattern' — the same
// machinery a tool's doc uses, which is what the owner ruled on 2026-07-24.
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

var WriteExperiencePatternInputSpec = datahelpers.ActionInputSpec{
	Optional: []string{
		"pattern_field", "created_by", "source", "allow_update",
	},
	Defaults: map[string]interface{}{
		"pattern_field": "experience_pattern",
		"allow_update":  true,
	},
}

func init() {
	datahelpers.RegisterActionInputSpec("write_experience_pattern", WriteExperiencePatternInputSpec)
}

// experiencePatternNameRE — kebab-case, the register's own naming convention
// (taxonomy_seed.md). Names are the join key for site_experiences and for the
// travelling doc's subject_key, so they must be stable and typo-resistant.
var experiencePatternNameRE = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// experiencePatternKinds — mirrors the CHECK on experience_patterns.kind
// (migration 218). Held in lockstep by
// TestExperiencePatternKinds_LockstepWithMigrationCheck.
var experiencePatternKinds = []string{"component-contract", "micro-journey"}

// ── the four column classes ────────────────────────────────────────────────
//
// Every column of experience_patterns belongs to exactly ONE of the four lists
// below, and TestExperiencePatternColumns_EveryColumnIsClassified fails if a
// column exists in the migrations that none of them names. That test is the
// answer to the objection three council seats raised independently on corr
// 2e71f640: the demotion list was hand-maintained with no mechanical guard,
// unlike the kinds vocabulary, and "a list drawn too narrow silently lets a
// changed clause keep 'approved' status — precisely the failure this rule
// exists to prevent" (guardian).
//
// A derived list is not possible here, because which fields are clause-bearing
// is a judgement rather than a fact about the schema. So the guard forces the
// judgement to be MADE: adding a column without classifying it fails the build.
// That converts a silent omission into a required decision, which is the most a
// mechanical check can do for a question that is not mechanical.

// experiencePatternContractFields are the fields whose change invalidates an
// approval (rule 2 above): they are what a reviewer was looking at when they
// approved the entry.
var experiencePatternContractFields = []string{
	"kind",
	"contract",
	"criteria_template",
	"binding_schema",
	"states",
	"automatic_triggers",
	"data_contract",
	"degraded_states",
	"entry_points",
	"requires_component_contract",
	"requires_invariant",
	"primitives",
	"destination_roles",
}

// experiencePatternSelectionFields decide WHERE an entry is offered, not
// whether it is sound. Changing them does not invalidate an approval, because
// nothing a reviewer approved becomes untrue when a pattern is offered on a
// different kind of page. Deliberately separate from the contract list rather
// than lumped in: over-wide demotion produces spurious warnings, and a warning
// that fires spuriously is one people learn to click past.
var experiencePatternSelectionFields = []string{
	"section_types",
	"suitable_site_types",
	"funnel_stage",
}

// experiencePatternCosmeticFields are presentation only.
var experiencePatternCosmeticFields = []string{
	"display_name",
	"description",
	"aka",
}

// experiencePatternSystemFields are written or owned by the platform, never by
// a submitting caller: identity, provenance, lifecycle, and the validator's own
// findings. `status`, `executable_checks` and `deferred_checks` are here rather
// than in any caller-facing list precisely because they are refused as input —
// an entry that scores itself walks through the constraint that reads the score.
var experiencePatternSystemFields = []string{
	"id", "name", "status", "executable_checks", "deferred_checks",
	"source", "source_agent", "harvested_from", "created_by",
	"created_at", "updated_at",
}

// experiencePatternJSONFields are every jsonb column this action writes, so the
// marshal loop cannot silently miss one that was added to the table.
var experiencePatternJSONFields = []string{
	"aka", "primitives", "section_types", "destination_roles", "suitable_site_types",
	"contract", "states", "automatic_triggers", "data_contract", "degraded_states",
	"entry_points", "requires_component_contract", "requires_invariant",
	"binding_schema", "criteria_template",
}

func WriteExperiencePatternAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "write_experience_pattern"))
	if params.ExecutionContext != nil && params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	config := params.StepConfig.Config

	field := datahelpers.GetStringField(config, "pattern_field", "experience_pattern")
	entry := datahelpers.ExtractNestedFieldMap(params.CollectedData, field)
	if len(entry) == 0 {
		return nil, fmt.Errorf("write_experience_pattern: no entry at %q", field)
	}

	sourceAgent := ""
	if params.ExecutionContext != nil {
		sourceAgent = params.Headers["agent_type"]
	}
	createdBy := datahelpers.GetStringField(config, "created_by", sourceAgent)
	source := datahelpers.GetStringField(config, "source", sourceAgent)

	// ── shape ──────────────────────────────────────────────────────────────
	problems := validateExperiencePatternShape(entry)

	// ── criteria (the formalised acceptance side) ──────────────────────────
	// The contract and the other clause-bearing documents are passed as extras
	// so that a placeholder used in a contract clause but never declared in
	// binding_schema is caught here too — it is the same defect as one used in
	// a check, and only differs in which document happens to carry it.
	var validation ExperienceCriteriaValidation
	if tmpl, ok := entry["criteria_template"].(map[string]interface{}); ok {
		validation = ValidateExperienceCriteria(tmpl,
			asMap(entry["binding_schema"]),
			entry["contract"], entry["states"], entry["data_contract"],
			entry["degraded_states"], entry["entry_points"])
		for _, e := range validation.Errors {
			problems = append(problems, fmt.Sprintf("criteria[%s].%s: %s", e.CheckID, e.Field, e.Detail))
		}
	} else if entry["criteria_template"] != nil {
		problems = append(problems, "criteria_template must be an object")
	}

	if len(problems) > 0 {
		sort.Strings(problems)
		logger.Warn("write_experience_pattern: REFUSED",
			zap.String("name", datahelpers.GetStringField(entry, "name", "")),
			zap.Strings("problems", problems))
		return nil, fmt.Errorf("write_experience_pattern: refused, %d problem(s): %s",
			len(problems), strings.Join(problems, "; "))
	}

	name := datahelpers.GetStringField(entry, "name", "")

	// ── referential integrity for invariants ───────────────────────────────
	// requires_invariant is jsonb, so the database cannot enforce this. An
	// entry naming an invariant that does not exist would silently reference
	// nothing, which is worse than inlining the clause: it LOOKS like the
	// clause is held centrally.
	missing, err := missingExperienceInvariants(ctx, params.DB, entry["requires_invariant"])
	if err != nil {
		return nil, fmt.Errorf("write_experience_pattern: checking invariants: %w", err)
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("write_experience_pattern: refused, entry %q requires invariant(s) that do not exist: %s",
			name, strings.Join(missing, ", "))
	}

	cols, vals, err := experiencePatternColumns(entry)
	if err != nil {
		return nil, fmt.Errorf("write_experience_pattern: %w", err)
	}

	// The validator's own accounting is stored on the row (migration 230), so
	// the approval question — does this entry actually assert anything? — is
	// answerable from the register alone, without re-running the validator or
	// reading a log. The database refuses 'approved' when executable_checks is
	// zero, which is what stops an entry that defers every hard check from
	// being approved on the strength of a promise.
	deferredJSON, err := json.Marshal(experienceDeferredRecords(validation))
	if err != nil {
		return nil, fmt.Errorf("write_experience_pattern: marshalling deferrals: %w", err)
	}
	cols = append(cols, "executable_checks", "deferred_checks")
	vals = append(vals, validation.Executable, string(deferredJSON))

	tx, err := params.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("write_experience_pattern: begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck — no-op after Commit

	// Read the existing row inside the transaction: the demotion decision and
	// the update must not straddle another writer.
	var prevStatus string
	prev := map[string]interface{}{}
	var prevJSON []byte
	err = tx.QueryRowContext(ctx, `
		SELECT status, to_jsonb(p) FROM experience_patterns p WHERE name = $1 FOR UPDATE`,
		name).Scan(&prevStatus, &prevJSON)
	switch {
	case err == sql.ErrNoRows:
		prevStatus = ""
	case err != nil:
		return nil, fmt.Errorf("write_experience_pattern: read existing: %w", err)
	default:
		if uerr := json.Unmarshal(prevJSON, &prev); uerr != nil {
			return nil, fmt.Errorf("write_experience_pattern: decode existing: %w", uerr)
		}
	}

	if prevStatus != "" && !datahelpers.GetBoolField(config, "allow_update", true) {
		return nil, fmt.Errorf("write_experience_pattern: %q exists and allow_update is false", name)
	}

	demoted, changedFields := false, []string(nil)
	if prevStatus == "approved" || prevStatus == "proven" {
		changedFields = changedExperienceContractFields(prev, entry)
		if len(changedFields) > 0 {
			demoted = true
		}
	}

	// Build the upsert. status is never taken from the entry (rule 1): a new
	// row is 'draft'; an existing row keeps its status unless a contract field
	// changed, in which case it returns to 'draft'.
	setClauses := make([]string, 0, len(cols)+3)
	args := make([]interface{}, 0, len(vals)+3)
	insertCols := []string{"name"}
	insertPlaceholders := []string{"$1"}
	args = append(args, name)
	for i, c := range cols {
		if c == "name" {
			continue
		}
		args = append(args, vals[i])
		insertCols = append(insertCols, c)
		insertPlaceholders = append(insertPlaceholders, fmt.Sprintf("$%d", len(args)))
		setClauses = append(setClauses, fmt.Sprintf("%s = EXCLUDED.%s", c, c))
	}
	setClauses = append(setClauses, "updated_at = now()")
	if demoted {
		setClauses = append(setClauses, "status = 'draft'")
	}

	query := fmt.Sprintf(`
		INSERT INTO experience_patterns (%s, created_by, source, source_agent, status)
		VALUES (%s, $%d, $%d, $%d, 'draft')
		ON CONFLICT (name) DO UPDATE SET %s
		RETURNING id, status`,
		strings.Join(insertCols, ", "),
		strings.Join(insertPlaceholders, ", "),
		len(args)+1, len(args)+2, len(args)+3,
		strings.Join(setClauses, ", "))
	args = append(args, nullIfEmpty(createdBy), nullIfEmpty(source), nullIfEmpty(sourceAgent))

	var id, newStatus string
	if err := tx.QueryRowContext(ctx, query, args...).Scan(&id, &newStatus); err != nil {
		return nil, fmt.Errorf("write_experience_pattern: upsert %q: %w", name, err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("write_experience_pattern: commit: %w", err)
	}

	fields := []zap.Field{
		zap.String("name", name),
		zap.String("pattern_id", id),
		zap.String("status", newStatus),
		zap.Int("executable_checks", validation.Executable),
		zap.Int("deferred_checks", len(validation.Deferred)),
	}
	if demoted {
		// Loud on purpose: an approval that quietly stops applying is exactly
		// the thing this rule exists to prevent, so it must be visible in the
		// log of the run that caused it.
		fields = append(fields, zap.Strings("demoted_because_changed", changedFields))
		logger.Warn("write_experience_pattern: entry DEMOTED to draft — approved contract changed, re-approval required", fields...)
	} else {
		logger.Info("write_experience_pattern: entry written", fields...)
	}

	return map[string]interface{}{
		"pattern_id":        id,
		"name":              name,
		"status":            newStatus,
		"created":           prevStatus == "",
		"demoted":           demoted,
		"demoted_because":   changedFields,
		"executable_checks": validation.Executable,
		"deferred_checks":   experienceDeferredRecords(validation),
		"validation":        validation.Summary(),
	}, nil
}

// validateExperiencePatternShape returns every problem with the entry, rather
// than the first — a writer that has to resubmit once per defect learns the
// contract one refusal at a time.
func validateExperiencePatternShape(entry map[string]interface{}) []string {
	var problems []string

	name := datahelpers.GetStringField(entry, "name", "")
	switch {
	case name == "":
		problems = append(problems, "name is required")
	case !experiencePatternNameRE.MatchString(name):
		problems = append(problems, fmt.Sprintf("name %q is not kebab-case", name))
	}

	kind := datahelpers.GetStringField(entry, "kind", "")
	if !containsString(experiencePatternKinds, kind) {
		problems = append(problems, fmt.Sprintf("kind must be one of %s, got %q",
			strings.Join(experiencePatternKinds, "|"), kind))
	}

	if datahelpers.GetStringField(entry, "display_name", "") == "" {
		problems = append(problems, "display_name is required")
	}

	// Rule 1: status is not writable. Refusing is deliberate — silently
	// ignoring it would let a caller believe it had approved something.
	if s, present := entry["status"]; present {
		problems = append(problems, fmt.Sprintf(
			"status is not writable here (got %v): 'approved' is a council verdict and 'proven' is a live green run; both are things that happened, not things a writer asserts", s))
	}

	// Same rule, same reason: these are the VALIDATOR's findings about the
	// entry, not claims the entry may make about itself. A writer that can
	// declare its own executable-check count can walk straight through the
	// approval constraint that count exists to enforce (migration 230).
	for _, derived := range []string{"executable_checks", "deferred_checks"} {
		if _, present := entry[derived]; present {
			problems = append(problems, fmt.Sprintf(
				"%s is derived from validation and is not writable — an entry that scores itself defeats the approval constraint that reads it", derived))
		}
	}

	contract, _ := entry["contract"].(map[string]interface{})
	if len(contract) == 0 {
		problems = append(problems, "contract is required and must be a non-empty object")
	} else {
		problems = append(problems, validateExperienceContractTriggers(contract)...)
	}

	if fs, ok := entry["funnel_stage"].(string); ok && fs != "" {
		if !containsString([]string{"awareness", "consideration", "conversion"}, fs) {
			problems = append(problems, fmt.Sprintf("funnel_stage %q is not awareness|consideration|conversion", fs))
		}
	}

	return problems
}

// validateExperienceContractTriggers enforces the two clauses that make a
// contract checkable at all: a trigger must name an observable outcome, and a
// navigating trigger must name a destination ROLE rather than a URL. A base
// entry carrying a concrete URL is the bugs_closed/045 class — a static value
// re-applied on every render that a per-site binding cannot override.
func validateExperienceContractTriggers(contract map[string]interface{}) []string {
	var problems []string
	triggers, ok := contract["triggers"].([]interface{})
	if !ok || len(triggers) == 0 {
		return []string{"contract.triggers must be a non-empty array"}
	}
	for i, t := range triggers {
		tm, ok := t.(map[string]interface{})
		if !ok {
			problems = append(problems, fmt.Sprintf("contract.triggers[%d] is not an object", i))
			continue
		}
		label := datahelpers.GetStringField(tm, "when", fmt.Sprintf("#%d", i))
		if datahelpers.GetStringField(tm, "then", "") == "" {
			problems = append(problems, fmt.Sprintf(
				"contract.triggers[%s]: `then` is required — a trigger with no observable outcome cannot be checked, and is how a control that does nothing gets written down as behaviour", label))
		}
		if role := datahelpers.GetStringField(tm, "destination_role", ""); role != "" {
			if strings.Contains(role, "/") || strings.Contains(role, ":") {
				problems = append(problems, fmt.Sprintf(
					"contract.triggers[%s]: destination_role %q looks like a path or URL — a base entry names a ROLE; the concrete page is bound per site", label, role))
			}
		}
		if u := datahelpers.GetStringField(tm, "destination", ""); u != "" {
			problems = append(problems, fmt.Sprintf(
				"contract.triggers[%s]: `destination` is a site-specific value and does not belong in a base entry; use destination_role", label))
		}
	}
	return problems
}

// changedExperienceContractFields compares the clause-bearing fields of the
// stored row against the incoming entry. Comparison is on canonical JSON, so
// key order and numeric formatting do not produce phantom changes — a spurious
// demotion would train people to ignore the warning.
func changedExperienceContractFields(prev, next map[string]interface{}) []string {
	var changed []string
	for _, f := range experiencePatternContractFields {
		if !sameCanonicalJSON(prev[f], next[f]) {
			changed = append(changed, f)
		}
	}
	return changed
}

func sameCanonicalJSON(a, b interface{}) bool {
	// A field absent from the incoming entry is "not changed", not "set to
	// null": a partial update must not silently empty a clause it did not
	// mention. Emptying one is done by sending an explicit empty value.
	if b == nil {
		return true
	}
	ab, err1 := json.Marshal(a)
	bb, err2 := json.Marshal(b)
	if err1 != nil || err2 != nil {
		return false
	}
	var av, bv interface{}
	if json.Unmarshal(ab, &av) != nil || json.Unmarshal(bb, &bv) != nil {
		return false
	}
	rea, _ := json.Marshal(av)
	reb, _ := json.Marshal(bv)
	return string(rea) == string(reb)
}

// missingExperienceInvariants returns the names in requires_invariant that have
// no row in experience_invariants.
func missingExperienceInvariants(ctx context.Context, db *sql.DB, requires interface{}) ([]string, error) {
	names := stringSliceOf(requires)
	if len(names) == 0 || db == nil {
		return nil, nil
	}
	rows, err := db.QueryContext(ctx, `SELECT name FROM experience_invariants WHERE name = ANY($1)`, pqStringArray(names))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	present := map[string]bool{}
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		present[n] = true
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	var missing []string
	for _, n := range names {
		if !present[n] {
			missing = append(missing, n)
		}
	}
	return missing, nil
}

// experiencePatternColumns marshals the entry into (columns, values) for the
// upsert. Scalars are passed through; every jsonb column is marshalled, so a
// column added to the table and to experiencePatternJSONFields is written
// without touching this function.
func experiencePatternColumns(entry map[string]interface{}) ([]string, []interface{}, error) {
	cols := []string{"name", "kind", "display_name"}
	vals := []interface{}{
		datahelpers.GetStringField(entry, "name", ""),
		datahelpers.GetStringField(entry, "kind", ""),
		datahelpers.GetStringField(entry, "display_name", ""),
	}

	for _, f := range []string{"description", "funnel_stage", "harvested_from"} {
		if v := datahelpers.GetStringField(entry, f, ""); v != "" {
			cols = append(cols, f)
			vals = append(vals, v)
		}
	}

	for _, f := range experiencePatternJSONFields {
		v, present := entry[f]
		if !present || v == nil {
			continue
		}
		b, err := json.Marshal(v)
		if err != nil {
			return nil, nil, fmt.Errorf("marshalling %s: %w", f, err)
		}
		cols = append(cols, f)
		vals = append(vals, string(b))
	}
	return cols, vals, nil
}

// experienceDeferredRecords renders the deferrals for storage and for the
// action result in one shape. Each carries its REASON: a deferral without one
// is indistinguishable from a check somebody could not be bothered to write,
// and the whole justification for allowing deferral at all is that the clause
// stays in the record.
func experienceDeferredRecords(v ExperienceCriteriaValidation) []map[string]string {
	out := make([]map[string]string, 0, len(v.Deferred))
	for _, d := range v.Deferred {
		out = append(out, map[string]string{
			"check_id": d.CheckID, "field": d.Field, "reason": d.Detail,
		})
	}
	return out
}

func asMap(v interface{}) map[string]interface{} {
	m, _ := v.(map[string]interface{})
	return m
}

func stringSliceOf(v interface{}) []string {
	switch t := v.(type) {
	case []string:
		return t
	case []interface{}:
		out := make([]string, 0, len(t))
		for _, e := range t {
			if s, ok := e.(string); ok && s != "" {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}
