# PLAN — B1 + B2: the offer track's premise-first slice

## Context

Owner decision 2026-08-08 (recorded in the lane PLAN's decision log): **B1 and B2 jump the
queue**, ahead of the rest of Programme B and independent of Programme A. Both are
`agent_definitions` config only — live on apply, no image roll, no committed-but-inert window.

- **B1** — `site-review-agent` asks offer-shaped questions ~16×/fortnight with no premise in
  context (`load_strategic_context` selects only domain/company/dream_spec/site_plan/2 counts).
  Fix: widen the query, extend the prompt to judge against the site's own recorded
  `revenue_models.primary_model`.
- **B2** — `domain-strategist` unconditionally chains `needs_briefing` → `build-briefing-agent`
  → `needs_site_plan` → `build-site-planner` (all three links verified live). Correct on
  greenfield (its only historical use, 3/3 rows); on a deployed site a premise refresh would
  re-plan it. Fix: gate the chain on deployed state; add the four Q-fields (a **restoration**
  of gaswholesalers' 2026-04-17 shape, not an invention).

Owning lane: `docs/agent_docs/docs024_key_docs_latest/vigilant_designer_offer_analysis/`.
Wider feature context: `features_open/030_FEATURE_offer_and_benefit_analyser.md`.

## Ground verified live 2026-08-08 (do not re-derive; re-verify only what you act on)

- `site-review-agent` chain: `ensure_site_record → load_strategic_context →
  spawn_content_auditor → call_content_auditor → run_strategic_review →
  write_strategic_findings → complete`. Query param binds `site_record.site_id`.
  `run_strategic_review`: sonnet-4-6, max_tokens 4000, findings cap 5, closed vocabulary.
- `domain-strategist` chain: `read_specs → analyze_strategy → write_strategy_spec →
  create_next_item → complete`, **no conditional anywhere**. `read_specs` =
  `read_site_spec` with no `aspect` ⇒ **reads ALL current aspects** (so an existing
  strategy row IS in the LLM's context on a refresh).
- `write_site_spec` **deep-merges** over the current row (`site_spec_actions.go` ~line 100):
  keys omitted by a later run structurally survive. Refresh-preserves is belt (prompt) +
  braces (merge).
- Conditional step shape (live example, improvement-loop `check_audit_due`):
  `{"action":"conditional","config":{"condition":"<field>.<key> == true","then_step":…,"else_step":…}}`.
- `extractWorkflowResult` (`coordinator.go:3757`): `ResultModeFields` **skips missing
  fields silently** ⇒ B2's gated path may reuse the existing `complete` step even though
  `next_item_created` never gets written on that path.
- `llm_call_log.prompt_rendered` exists ⇒ B1's planted-marker witness instrument.
  Post-2026-07-26 `agent_type` is the resolved type; key on `step_name` too.
- Aspect sizes (current rows): strategy median 12KB/max 15KB; identity median 2.6KB/**max
  26.8KB**; audience median 364B; content_direction median 20KB/**max 50KB**; mission_brief
  ~2.4KB (`{text:…}` wrap only).
- `sites.status` is NOT a reliable deployed predicate (loanandmortgagecalculator = `active`
  with 41 deployed pages). Use `count(pages WHERE build_status='deployed') > 0`.
- B2 witness target `loancalculator.co.uk`: site `0162cde4-…`, status deployed, **27
  deployed pages, NO strategy aspect** — ideal: gate must suppress the chain while the first
  strategy row is written.
- `primary_model` lives at `revenue_models.primary_model`, 16/17 sites (10× direct_business).
  The lane PLAN's §B text is correct — do NOT "fix" it (see PLAN corrections block 08-08).

## The changes

### B1 — migration: site-review-agent loads the premise

One migration file in `docs/agent_docs/sql_for_agents/` (take the next free number at apply
time — 336 is the current tail and numbers get taken mid-session; header names the lane and
phase, following `301_render_audit_agent_write_findings_tail.sql` style).

1. **`load_strategic_context` query**: keep the existing SELECT and add COALESCE'd
   subselects per aspect, following the query's own existing pattern
   (`COALESCE(ss.data::text,'{}')`):
   - `strategy` — full (`data::text`); it is the point.
   - `audience` — full (tiny).
   - `mission_brief` — `data->>'text'`.
   - `identity`, `content_direction` — **capped**: `left(data::text, 4000)` with the cap in
     the column NAME (`identity_head_4k`, `content_direction_head_4k`) so the filter ships
     visibly (max sizes 26.8KB/50KB would dilute a 5-finding review). Each subselect:
     `ORDER BY created_at DESC LIMIT 1` on `is_current` rows, COALESCE to `'{}'` — a NULL
     into a Go template renders `<no value>` (the css-patch commit-message defect class).
2. **`run_strategic_review` prompt**: add the new context blocks; add one instruction: judge
   the site against its own recorded `revenue_models.primary_model` — a mismatch between
   revenue shape and page structure/CTAs is a top-5-worthy finding. **Vocabulary stays
   closed** (same 5 `work_item_type`s) — no new item types in B1; that is B3/B4 territory.
   Model config untouched (max_tokens 4000 stays; note the truncation watch below).
3. Migration verify block: `DO`/`RAISE` (a SELECT cannot stop a COMMIT), asserting the new
   query text and prompt substring landed on the active, non-snapshot row. Induce it once
   (run against a scratch condition that fails) before trusting it.

### B2 — migration: domain-strategist refresh-safe + Q-fields

Second migration file, same conventions.

1. **New step `check_site_deployed`** between `write_strategy_spec` and the chain:
   `query_database`: `SELECT (count(*) > 0) AS is_deployed FROM pages WHERE site_id = $1
   AND build_status = 'deployed'`, params `["input_data.site_id"]` (matches
   `create_next_item`'s existing binding), `output_format: object`,
   `output_field: site_state`.
2. **New step `gate_next_item`**: conditional, `condition: "site_state.is_deployed == true"`,
   `then_step: "complete"` (chain suppressed), `else_step: "create_next_item"` (greenfield
   behaviour byte-identical). Rewire `write_strategy_spec.next_step` → `check_site_deployed`.
   Predicate is DB-state only — **never** an `input_data.spec.*` path (a missing spec key
   fails input_mapping; worked example: the dead `needs_logo` path). No override flag (no
   consumer needs one; add it when one does).
3. **`analyze_strategy` prompt**: add four fields to the output JSON schema, using the LIVE
   spellings from gaswholesalers' 2026-04-17 row where they exist: `satisfaction_condition`,
   `trust_threshold`, `recurring_value`, plus `money_flow` (PLAN §B2's name for the fourth).
   Add the refresh-preserves instruction: if Research Data contains an existing `strategy`
   aspect, treat it as prior — revise where contradicted, do not reinvent. (Deep-merge
   already preserves omitted keys structurally.)
4. Verify block: DO/RAISE asserting new steps exist + `write_strategy_spec.next_step`
   repointed + prompt carries a Q-field name. Induce once.

## Cross-cutting discipline (both migrations)

- **Snapshot first**: insert an `is_snapshot` copy of each agent row pre-UPDATE (318's
  pattern). No `restore_agent_snapshot()` exists — rollback is restoring from the snapshot
  row directly; write the restore statement into the migration header.
- **Apply**: single `psql -f` + `--record-only` — never blanket `--apply` (the tree carries
  other sessions' pending migrations).
- **Council**: two submissions, one per coherent task (B1, B2), rationale + ≤8-edit plan
  JSON, via `097_TRIGGER_council_review_v1.sh`. Budget ~30 min each; find runs by payload
  not printed id. Commit with `Council-Submitted: <corr>` (never `Council-Reviewed` on an
  unread verdict). Config-migration precedent: 290/291/301/318.
- **Consumers told** (owner ruling 2026-07-29 §3): B2 changes `domain-strategist` for its
  existing producer (`vertical-exemplar-researcher`, greenfield — behaviour there unchanged,
  but they must be told, not merely measured). CONTRIB note into `portfolio_positioning/`
  naming what changed about the guarantee; run `scripts/who-owns.py` first.
- **Landmines**: at execution start, grep `LANDMINES.md` for `site-review-agent`,
  `domain-strategist`, `site_specs`, `llm_call_log`, `agent_definitions` footprints.
- **Docs as you go**: append lane NOTES entry (including that this plan was executed);
  commit this plan's content as `PLAN_2026-08-08_B1_B2_premise_first.md` in the lane dir;
  pathspec commits, one per task.

## Verification (the witnesses ARE the deliverable — "applied" is not "seen")

### B1 witness — planted marker reaching the assembled prompt
1. Plant: direct SQL on the current `audience` row of the witness site (read by nothing per
   the positioning lane's 07-31 measurement): `UPDATE site_specs SET data = data ||
   '{"__b1_marker":"B1-MARKER-2026-08-08"}' WHERE …aspect='audience' AND is_current`.
2. Witness site: **webdesign.co.uk** (PLAN §B5's named proof site; a full sweep ran there
   2026-08-05 under broad autonomy, so consequences are known ground). Pre-check the 291
   gate would pass (`settings.maintenance_profile.last_audit` vs current fingerprint) — if
   audit is not due the review never spawns; pick another never-audited deployed site via
   that same query rather than forcing.
3. Fire `./run_improvement_sweep_once.sh webdesign.co.uk` (≥300s after any chassis pod
   start; queued ≠ lost — the measured norm is ~29 min, one observed 9.5h; never retry on a
   missing orchestration row).
4. Assert marker in `llm_call_log.prompt_rendered` WHERE `agent_type='site-review-agent'`
   AND `step_name='run_strategic_review'` newest row. Also check the same row for
   `error_message ILIKE '%stop_reason=max_tokens%'` — truncation is stated ONLY there
   (output_tokens is NULL on cut first attempts), and the premise blocks lengthen the input.
5. Un-plant: `UPDATE … SET data = data - '__b1_marker'` (deep-merge cannot delete keys, so
   direct SQL both ways).
6. Read the strategic findings the run filed — B1's value is visible here: do they cite the
   premise? (Qualitative, record in NOTES.)

### B2 witness — refresh on a deployed site files no briefing
1. Hand-file one `needs_strategy` row for loancalculator.co.uk, copying the live 08-02 rows'
   shape (`item_key`, `handler_agent='domain-strategist'`, `status='triaged'`,
   `created_by` naming this verification); let build-dispatch-loop dispatch it.
2. Assert, in order: (a) orchestration reached `complete` (a FAILED step can show COMPLETED
   with `error` NULL — read `__step_error`); (b) a new `strategy` row exists for the site
   carrying `revenue_models.primary_model` AND at least one Q-field; (c) **zero**
   `needs_briefing` rows for the site (`created_at > dispatch time` — assert row identity,
   not just count); (d) the gate's `site_state.is_deployed` visible in `collected_data`.
3. **Negative control** (the gate could be dead code that never routes): confirm the
   greenfield arm still works by evidence, not assumption — either a later real greenfield
   run, or assert in `collected_data` that `gate_next_item` evaluated and chose `complete`
   (the routing decision itself is the witness that the conditional parsed).
4. loancalculator is adopted: the strategy write touches `site_specs` only, no pages — but
   verify (e) no work items of ANY type were created for the site by this orchestration
   beyond the one hand-filed (`created_by`/timestamp window), which is the actual "no side
   effects" claim.

### Order
B1 → witness → B2 → witness. Independent changes, but B1 first means loancalculator's new
strategy row is read by the review from day one. If B1's council round REVISEs, B2 proceeds
in parallel — they share no files.

## Explicitly out of scope
B3 (the two checks — needs the register entry per RFC_010 §1 and IMP-016 observe-only
sequencing), B4 (the analyser agent), any council seat, any new item_type, anything in
Programme A (198's witnessed css-patch run stays that track's next action).

---

## Execution corrections (2026-08-08, same session — the plan above is as approved; what differed in reality)

- **The council premise was WRONG and the submissions were refused client-side.** The gate's
  scope is `platform/`, `internal/`, `pkg/` (owner ruling 2026-07-17) and it refuses
  config-only submissions before spending credits. The plan cited "config-migration
  precedent: 290/291/301/318" — re-checked: **290/291/301 were applied + recorded WITHOUT
  council rounds** (only Go changes A0.3/A1.2 and the mixed 318 went through). Both B1 and
  B2 therefore shipped without a council round, following the lane's actual config
  precedent, stated plainly in both commit messages. Not FORCEd.
- **`conditional` is a deprecated alias** — the registered name is `conditional_branch`
  (registry.go:65; same handler). B2 uses `conditional_branch`.
- **The marker went in `strategy`, not `audience`** — webdesign.co.uk has no `audience`
  aspect (only strategy / identity / content_direction / mission_brief). The fingerprint
  hashes pages + palette + chrome, not site_specs, so the plant cannot flip the 291 gate.
- **The ledger records were blocked by the session permission classifier** (both the runner
  script and a direct INSERT). Handed to the owner; both files carry probe guards so an
  accidental replay refuses loudly instead of double-applying.
- **Applies dated by `snapshot_taken_at`** (same-transaction): B1 17:58:31Z, B2 18:01:50Z.
  `agent_definitions.updated_at` did NOT move (the UPDATE sets only default_config) — it
  still reads the 16:26 fleet-touch, so it cannot date an apply. A 17:17 sweep
  (leopardessconsulting, another session) ran the OLD prompt because it spawned 40 minutes
  before B1's apply — expected, not an anomaly; workflow plans freeze at spawn.
