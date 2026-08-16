# PLAN 2026-08-16 — bug 265: make the legacy `input_schema` dialect unrepresentable

Lane: `bugfix_265_legacy_dialect_unrepresentable`. Bug file:
`bugs_open/265_HANDOFF_2026-08-12_legacy_schema_dialect_declared_extinct_is_being_reintroduced_and_its_tripwire_only_warns.md`.

## 0. Why this bug, and why now (session start, 2026-08-16)

This session was pointed at 213 and 274; both are **already closed at HEAD** (213 by owner
ruling 2026-08-15, 274 live+proven the same day). A census of `bugs_open/` for a bug that
is (a) still open, (b) unowned — no workstream dir, no live transcript editing it in 36 h,
last commit 2026-08-12 — and (c) a framework defect rather than a site one, produced 265.
`who-owns.py 265` → no owning workstream identified.

## 1. What the bug IS, in plain terms

A component's `input_schema` tells the platform which content fields it has and which are
required. The house dialect is `{"fields": {...}}`. An older dialect — JSON-Schema,
`{"type":"object","properties":{...},"required":[...]}` — was declared **extinct** in a code
comment on 2026-07-21 (0 of 173). Since then components in the old dialect have been added
again. The reader tolerates the old dialect (it projects it onto the new shape) and logs a
`Warn` that nobody reads. So the comment readers trust is false, and the detector that should
have said so is silent.

## 2. What was re-verified today (the bug is still valid, but the picture moved)

| claim in the file | today [MEASURED 2026-08-16] |
|---|---|
| 4 legacy rows, newest 2026-08-10 | **3** — `report-dossier`, `mechanism-flow`, `evidence-timeseries`. `loans-consolidation` was converted to v2 by the LMC lane's `b2_convert_oldshape.py` on 2026-08-15 14:06Z (`updated_at`), no `component_versions` row |
| producer "likely the component-creator path" `[UNVERIFIED]` | **WRONG direction.** All 3 (and the 4th) are `created_from='manual'`, `source_agent_type` NULL. The three seeds are on disk: `sql_for_agents/207` (gripper-dossier lane, committed 2026-07-25 — four days after the extinction census), `247` and `250` (oufe lane, 2026-07-28); each writes `"properties"` + `"required"` verbatim. The component-creator (`created_from='generated'`, 69 rows 03-31→07-06) has produced **0** legacy rows in its life |
| the tripwire only Warns | unchanged — `WarnLegacyDialect` is still a `logger.Warn` and nothing reads it |
| doc comment claims extinction | unchanged, `component_schema_fields.go:53-56` |

The producer finding is the decisive one: **the dialect is reintroduced by hand-authored SQL
seeds and scripts**, and no Go-side gate on the component-creator would ever have seen a single
one of the four rows. The estate already knows this about components in general — the
`component-fallback-check` CronJob header (RFC_009) says it in terms: *"content_components live
only in the database. A component is routinely changed by a migration or by hand with no commit
at all."* So the seam that sees every producer is the table itself.

Verification route (owner ruling 2026-07-31): no `090` run for the producer claim. Substituted:
a full-population enumeration on provenance columns (`created_from`, `source_agent_type`) that
could have come out otherwise (it would have shown `generated` rows carrying `properties`), plus
the three seed files read on disk. The **fix** goes through the council gate.

## 3. Design — three layers, ordered by what makes the bad state unrepresentable

**Layer 1 (the guarantee) — a CHECK constraint on `content_components.input_schema`.**
`chk_input_schema_no_legacy_dialect`: `input_schema IS NULL OR jsonb_typeof(input_schema)
<> 'object' OR NOT (input_schema ? 'properties')`. Refuses INSERT and UPDATE from every
producer — seeds, scripts, Go, admin UI, psql by hand, a restore from any `bak_*` table. This
is the only layer that would have stopped the four rows that actually happened. It refuses the
**top-level** key only: nested `properties` inside `fields.<x>.items` is the shape of an item
and is legitimate (mechanism-flow and evidence-timeseries both carry one).

**Layer 1a (data) — convert the 3 legacy rows to v2 in the same migration**, mirroring
`SchemaContentFields`' projection exactly (the six copied keys, `minItems`→`min_items`,
`string`→`text`, `description`→`llm_guidance`, `source` defaults to `llm`, `required[]` folded
in). Backup table first; DO/RAISE verify: 3 before, 0 after, every converted row now reads
`ok=true, fromLegacy=false`, field NAMES identical. Equivalence of the projection is proven in Go
by a test that runs `SchemaContentFields` on the before/after JSON captured from the live rows.

**Layer 2 (legibility at the one LLM producer) — a fourth birth check in
`store_generated_component`.** If the generated `input_schema` is the legacy dialect, fail
the step with an error that names the house dialect, before the INSERT/UPDATE. Without it, the
constraint would still refuse — but as SQLSTATE 23514 naming a constraint, on a step that also
derives `render_mode` from `fields` only (`deriveRenderMode`), so a JSON-Schema-shaped
generation would previously have been stored as `render_mode='template'` with an empty
field set. Reuses `SchemaContentFields`; no new classifier.

**Layer 3 (the stale invariant) — rewrite the doc comment and the tripwire's message** so
they cite the constraint, not a census. A census goes stale the day after it is taken; a
constraint is checkable at any moment (`\d content_components`). The tripwire stays (defence
in depth — it now fires only if the constraint has been dropped or a schema arrived from
outside `content_components`), moves from `Warn` to `Error`, and its message says what a firing
now means. Correct the stale prior-art pointer (`bugs_open/026` → `bugs_closed/026`).

**Explicitly NOT done, and why:**
- No new CronJob check. RFC_006's daily-check pattern exists for invariants the database
  cannot express; this one it can. A constraint does not drift.
- Not deleting the legacy projection branch in the reader. After Layer 1a nothing in
  `content_components` can reach it, but it is the fail-safe if the constraint is ever
  dropped — the alternative is the fail-OPEN blind spot bug 026 was opened about.
- Not rewriting the three applied seeds (`207`/`247`/`250`). Applied migrations are history.
  A hand re-run of one now fails loudly on the constraint, which is the intended behaviour.
- Not touching `component_versions` (history) or the `bak_*` tables.

## 4. Scope ruling, applied

Is this architecture-scope? It **changes what a shared table guarantees** ("no legacy dialect
can be stored"), so by the 2026-07-29 ruling §1 it is more than a bare vocabulary addition. But
it is a *strengthening* every consumer already assumes — every Go reader either reads only
`fields` or projects legacy through one helper; the guarantee removes a case, it does not add
authority. Consumers enumerated: writers of `input_schema` are `store_generated_component_action.go`
(INSERT :528, UPDATE :466), `deploy_tool_action.go:303` (copies an existing row's schema —
transitively safe), the LMC/loancalculator decomposition loaders (Python, write v2), and any
seed. Readers all go through `SchemaContentFields` or `["fields"]`. Named in the council
submission; council gate, one round.

## 5. Sequence

1. Bug file: today's re-verification + producer correction (visible correction, not an edit).
2. Migration `437_content_components_refuse_legacy_input_schema_dialect.sql` + `_ROLLBACK`.
3. Go: Layer 2 + Layer 3 + equivalence test. `go build ./... && go test ./platform/orchestration/...`.
4. Council submission (JSON in this dir), `Council-Submitted:` trailer on the commit.
5. Apply the migration by hand (scoped), record with `--record-only`; induce: try to INSERT a
   legacy-dialect scratch row → must be refused; re-run the census → 0.
6. Commit Go + migration + docs by pathspec. Go is inert until the next chassis roll; the
   constraint is live the moment the migration is applied.
7. Verify post-roll: birth-path literal in the binary; tripwire silence with the constraint as
   its explanation. Register the constraint in the concept register (it is a mechanism another
   lane will hit when a seed is refused).
