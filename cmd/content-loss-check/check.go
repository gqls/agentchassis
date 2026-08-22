// FILE: cmd/content-loss-check/check.go
//
// The queries and pure predicates. main.go owns the flow; this file owns what
// is asked and how a finding is judged, because the pure parts are what the
// tests hold and the SQL is what the RUNBOOK mirrors.
//
// THE DEFINITION OF "NON-LLM KEY", named because four partial definitions
// exist and they disagree (bugs_open/355, parallel-session census):
// planSection's resolver arm excludes ""|llm|renderer|static
// (plan_sections_action.go), its carry arm excludes only ""|llm (widened by
// bugs_open/268). THIS check uses: source PRESENT and NOT LIKE 'llm%'.
//   - renderer/static INCLUDED: they are machine-derived values a regeneration
//     must preserve — the 268 incident (214 anchors, 19 sites) and every one of
//     the 72 known funnel losses live in exactly this class, so excluding them
//     would blind the check to the only losses ever measured AND zero its own
//     demand control, refusing for ever.
//   - blank/absent source EXCLUDED: matches the carry arm's ""|llm exclusion —
//     an undeclared field is the writer's to rewrite, and its disappearance is
//     not this class.
//
// SCHEMA READ: input_schema->'fields' raw. Licensed by migration 437's CHECK
// constraint chk_input_schema_no_legacy_dialect — content_components REFUSES
// the legacy properties/required dialect at the table, so the projection
// datahelpers.SchemaContentFields performs for in-memory schemas cannot be
// needed here (verified 2026-08-22: 0 active rows carry a top-level
// "properties" key). If that constraint is ever dropped, this read fails OPEN
// on legacy rows — the constraint is the guarantee, not this comment.
package main

import (
	"fmt"
	"sort"
	"strings"
)

// lossRow is one (archived generation, key) pair where a schema-declared
// non-llm key went present-and-non-blank -> absent-or-blank across consecutive
// generations of the same (page, slot).
type lossRow struct {
	HistoryID string // page_component_history.id of the pre-image
	PageID    string
	SiteID    string
	SlotName  string
	Op        string // 'delete' = the funnel; 'overwrite' = the in-place writers
	Writer    string // application_name of the pre-image row (A1 stamps land here)
	Key       string
	Source    string // the field's declared source
	LostAt    string // pre-image created_at
}

// The detection query. One parameter, the lookback in days (fmt'd in as an
// int — both connection routes need literal SQL). Coverage notes that belong
// next to the query, not in a doc nobody reads with it:
//   - the FK schema route silently loses 58% of rows (ON DELETE SET NULL);
//     the slot fallback recovers 24% -> 73% coverage (bugs_open/355 §3.1);
//   - op is a PROXY for the writer: 'delete' rows are the funnel only because
//     save_page_sections is DELETE+INSERT (bugs_open/355 §7);
//   - INSERT is never archived (pch_op_check) — rows born incomplete are
//     invisible here and covered by the live-damage census below instead.
const lossQueryTemplate = `
WITH pre AS (
  SELECT h.id, h.page_id, h.site_id, h.slot_name, h.component_id, h.op,
         h.content_data AS before_data, h.created_at,
         COALESCE(h.application_name,'') AS writer
    FROM page_component_history h
   WHERE h.source='artefact_archive_trigger'
     AND h.created_at > now() - make_interval(days => %d)
     %s
), sch AS (
  SELECT p.*,
         COALESCE(
           (SELECT cc.input_schema->'fields' FROM page_components pc
              JOIN content_components cc ON cc.id = pc.component_id
             WHERE pc.id = p.component_id),
           (SELECT cc2.input_schema->'fields' FROM page_components pc2
              JOIN content_components cc2 ON cc2.id = pc2.component_id
             WHERE pc2.page_id = p.page_id AND pc2.slot_name IS NOT DISTINCT FROM p.slot_name
             ORDER BY pc2.updated_at DESC LIMIT 1)) AS fields,
         COALESCE(
           (SELECT h2.content_data FROM page_component_history h2
             WHERE h2.source='artefact_archive_trigger' AND h2.page_id=p.page_id
               AND h2.slot_name IS NOT DISTINCT FROM p.slot_name AND h2.created_at > p.created_at
             ORDER BY h2.created_at ASC LIMIT 1),
           (SELECT pc3.content_data FROM page_components pc3
             WHERE pc3.page_id=p.page_id AND pc3.slot_name IS NOT DISTINCT FROM p.slot_name
             ORDER BY pc3.updated_at DESC LIMIT 1)) AS after_data
    FROM pre p
)
SELECT COALESCE(json_agg(json_build_object(
         'history_id', t.id, 'page_id', t.page_id, 'site_id', t.site_id,
         'slot_name', t.slot_name, 'op', t.op, 'writer', t.writer,
         'key', t.key, 'source', t.source, 'lost_at', t.created_at))::text, '[]')
FROM (
  SELECT s.id, s.page_id, s.site_id, COALESCE(s.slot_name,'') AS slot_name,
         s.op, s.writer, k.key,
         s.fields->k.key->>'source' AS source, s.created_at
    FROM sch s, LATERAL jsonb_object_keys(s.fields) AS k(key)
   WHERE s.fields IS NOT NULL AND s.after_data IS NOT NULL
     AND s.fields->k.key->>'source' IS NOT NULL
     AND s.fields->k.key->>'source' NOT LIKE 'llm%%'
     AND s.before_data ? k.key AND btrim(COALESCE(s.before_data->>k.key,'')) <> ''
     AND (NOT (s.after_data ? k.key) OR btrim(COALESCE(s.after_data->>k.key,'')) = '')
) t`

// demandControlFilter pins the control to the funnel's KNOWN pre-fix losses
// (72 measured 2026-08-22: 08-09 x4, 08-11 x63, 08-12 x5 — the population the
// 268 carry extension ended on 08-14). The clock is PINNED, not relative: a
// control that slides forward with now() would drain to zero as retention or
// deletions eat the evidence and the check would start refusing on a
// vanished baseline it could no longer explain. If this control ever returns
// zero, the INSTRUMENT is blind (schema pointers gone, trigger dropped, table
// truncated) — the check refuses (exit 2) rather than reporting a clean sweep,
// and a human re-pins the control against whatever losses are then known.
const demandControlFilter = `AND h.op = 'delete' AND h.created_at < '2026-08-15'`

// coverageQueryTemplate reports how much of the window the instrument can
// actually judge — a zero over 30 judgeable pairs and a zero over 5,000 are
// different facts, and the heartbeat must say which one it is reporting.
const coverageQueryTemplate = `
WITH pre AS (
  SELECT h.id, h.page_id, h.slot_name, h.component_id, h.op, h.created_at
    FROM page_component_history h
   WHERE h.source='artefact_archive_trigger'
     AND h.created_at > now() - make_interval(days => %d)
)
SELECT json_build_object(
  'pairs_total', count(*),
  'pairs_delete', count(*) FILTER (WHERE op='delete'),
  'pairs_overwrite', count(*) FILTER (WHERE op='overwrite'),
  'judgeable', count(*) FILTER (WHERE
     EXISTS (SELECT 1 FROM page_components pc JOIN content_components cc ON cc.id=pc.component_id
              WHERE pc.id = pre.component_id AND cc.input_schema ? 'fields')
     OR EXISTS (SELECT 1 FROM page_components pc2 JOIN content_components cc2 ON cc2.id=pc2.component_id
              WHERE pc2.page_id = pre.page_id AND pc2.slot_name IS NOT DISTINCT FROM pre.slot_name
                AND cc2.input_schema ? 'fields')))::text
FROM pre`

// liveDamageQuery is the INSERT-blind-spot cover (bugs_open/355 §3.3): the
// archive never sees a row being born, so standing damage is measured at the
// CURRENT state — deployed rows whose schema declares a REQUIRED non-llm field
// that is absent or blank. Required-only, deliberately: an optional non-llm
// field that is absent is a template with a gate, not damage. This is also the
// check's DURABLE record of standing damage — finding rows in agent_error_log
// expire (30d unresolved / 14d resolved, mig 466), this census re-derives from
// live state every run.
const liveDamageQuery = `
SELECT json_build_object(
  'rows', count(*),
  'examples', COALESCE(json_agg(json_build_object(
      'domain', t.domain, 'page', t.page, 'slot', t.slot, 'key', t.key, 'source', t.source)
      ) FILTER (WHERE t.rn <= 10), '[]'))::text
FROM (
  SELECT s.domain, p.name AS page, COALESCE(pc.slot_name,'') AS slot, f.key,
         f.value->>'source' AS source,
         row_number() OVER (ORDER BY s.domain, p.name, pc.slot_name, f.key) AS rn
    FROM page_components pc
    JOIN pages p ON p.id = pc.page_id
    JOIN sites s ON s.id = p.site_id
    JOIN content_components cc ON cc.id = pc.component_id,
    LATERAL jsonb_each(cc.input_schema->'fields') f
   WHERE pc.build_status = 'deployed'
     AND f.value->>'source' IS NOT NULL
     AND f.value->>'source' NOT LIKE 'llm%'
     AND (f.value->>'required')::boolean IS TRUE
     AND btrim(COALESCE(pc.content_data->>f.key,'')) = ''
) t`

// ---------------------------------------------------------------------------
// Pure predicates — what the tests hold.
// ---------------------------------------------------------------------------

// isNonLLMLoss is the Go statement of the SQL predicate above, used by the
// runtime canary (main.go fabricates a pair and refuses to run if this
// disagrees with itself) and by the tests. Keep the two in step BY HAND — the
// runtime demand control is what catches the SQL side drifting (a query that
// cannot find the known funnel losses exits 2), and this function is what
// catches the DEFINITION drifting (a canary loss that stops being a loss
// fails the run before it can report clean).
func isNonLLMLoss(source, before, after string) bool {
	if source == "" || strings.HasPrefix(source, "llm") {
		return false
	}
	if strings.TrimSpace(before) == "" {
		return false // nothing was held, nothing was lost
	}
	return strings.TrimSpace(after) == ""
}

// healVerdict grades one finding against the CURRENT row state at its
// (page, slot). The verdicts:
//
//	"healed"   — a deployed row holds the key non-blank again
//	"row_gone" — no row exists at that (page, slot) at all: nothing serves the
//	             damage any more (the slot was removed or the page rebuilt
//	             without it — either way there is no artefact to repair)
//	"open"     — a row exists and the key is still absent/blank, or the row is
//	             mid-flux (not deployed); the finding stays unresolved
func healVerdict(rowsAtSlot, deployedAtSlot int, deployedHoldsKey bool) string {
	switch {
	case rowsAtSlot == 0:
		return "row_gone"
	case deployedHoldsKey:
		return "healed"
	default:
		return "open"
	}
}

// carryMissVerdict folds per-field verdicts for a STRUCTURAL_KEY_CARRY_MISS
// finding (one row names several fields). The row resolves only when EVERY
// field is healed or gone; one open field keeps the whole finding open, and
// the open fields are reported so a partially-healed row is never read as
// done.
func carryMissVerdict(fieldVerdicts map[string]string) (resolved bool, verdict string, openFields []string) {
	for f, v := range fieldVerdicts {
		if v == "open" {
			openFields = append(openFields, f)
		}
	}
	sort.Strings(openFields)
	if len(openFields) > 0 {
		return false, "open", openFields
	}
	// all healed/gone — pick the stronger word for the resolved_by stamp
	for _, v := range fieldVerdicts {
		if v == "healed" {
			return true, "healed", nil
		}
	}
	return true, "row_gone", nil
}

// lossFindingKey is the dedupe identity for a CONTENT_KEY_LOSS finding: one
// finding per (pre-image archive row, key). Re-runs over an overlapping
// window re-derive the same losses; this is what makes the insert idempotent.
func lossFindingKey(l lossRow) string {
	return l.HistoryID + ":" + l.Key
}

func buildLossQuery(lookbackDays int, control bool) string {
	filter := ""
	if control {
		filter = demandControlFilter
	}
	return fmt.Sprintf(lossQueryTemplate, lookbackDays, filter)
}

func buildCoverageQuery(lookbackDays int) string {
	return fmt.Sprintf(coverageQueryTemplate, lookbackDays)
}
