# NOTES — `bugs_open/323`, cta_improvement refusal completes green

Running technical record, append-only, newest at the bottom. What was tried, what the system said,
every misstep. Figures are `[MEASURED <date>]` with the query in `RUNBOOK_cta_improvement_refusal.md`
unless marked `[INFERRED]`/`[UNMEASURED]`.

---

## 2026-08-19 (session 1, afternoon) — orientation and re-validation

**Ownership.** `scripts/who-owns.py 323` → no owning workstream; the filing lane
(`bugfix_302_design_repair_verification`) says in its own README (l.327) that it filed the
measurement and deliberately did NOT chase it. Dirty tree at session start contained none of the
files this lane touches (`write_audit_findings_action.go`, `fix_component_template_action.go`,
`complete_work_item_*.go`, the fixer's agent row). Picked up.

**The bug is still valid** `[MEASURED 2026-08-19 ~15:30Z, archive-inclusive]`: 993 `cta_improvement`
completions, 22 sites, **0 ever `fixed=true`**; 468 carry `fixed=false` +
`reason='fix_type requires LLM-driven changes, not programmatic HTML edits'`; the other 525 carry a
FOREIGN payload in `result` (441 a webdesign-agent design-token blob from Mar–May, 12 spawn records
per `bugs_closed/287`, ~70 triage decisions). Live table: 34 rows 08-11→08-17 across 12 sites, from
FIVE producers — `design-audit`, `content-quality-audit`, `site-review`, `offer-analysis`,
`brief-fidelity-audit` — all `complete`. Still flowing, still closing green.

**The mechanism, read rather than inferred.**
- `write_audit_findings_action.go` Rule 3: `componentCategories = {cta, nav_restructure}` →
  `ItemType cta_improvement`, `HandlerAgent component-template-fixer`, `spec.fix_type` from
  `categoryToFixType["cta"] = "cta_improvement"`.
- `fix_component_template_action.go` dispatch switch, `case "cta_improvement", "cta",
  "nav_restructure"` (arm dated `e535d4f52`, **2026-03-14**): returns `{fixed:false, fix_type,
  reason:"fix_type requires LLM-driven changes, not programmatic HTML edits", action:"needs_review"}`.
- The fixer's live workflow (`agent_definitions`, updated 08-19 12:14Z by another lane — the
  `fix_type_field` key) branches `check_needs_rerender` on `fix_result.fixed == true` ONLY; the
  else path is compose_note → append_note → `complete_workflow` (success-labelled).
- `build-dispatch-loop` `process_item` sub-workflow: `call_handler` → `mark_complete`
  (`complete_work_item`, `result! = handler_result`).
- `CompleteWorkItemAction`: gate 1 `handlerReportedFailure` reads `response.status` ∈
  {failed,failure,error} — the fixer never sets `response.status` (complete_workflow has no
  config key for it; `workflow_actions.go` builds the envelope from output_fields only). Gate 1b
  `noChangeGates` is opt-in per item_type, numeric `CounterPaths` only — `cta_improvement` is not
  on it and its payload is a boolean. Gate 2 (verifier) — none registered for the type. → `UPDATE
  … status='complete'`.

**FINDING — the handler ALREADY separates refusal from no-op, and nothing reads it.** The 302 lane
wrote (bug file + roster comment) that `spacing_fix`'s "already has flex CSS" and
`cta_improvement`'s "requires LLM" are the SAME SHAPE ("fixed:false") and only `reason` prose
separates them. That is wrong on the handler's own vocabulary: every REFUSAL arm in
`fix_component_template_action.go` (13 emission sites: the cta/nav punt, the unrecognised
fix_type default, chrome-lock refusals ×2, `chrome_overflow_fix` missing/unsafe slot or selector
×3, `repair_page_component_status` refusals ×4, the locked-template partial) carries
`action: "needs_review"`; every IDEMPOTENT arm ("already has flex CSS", "already has responsive
CSS", "already deployed", "already patched for this selector", "no component_id and no rendered
HTML") carries NO `action` key. `[MEASURED 2026-08-19, archive-inclusive, all component-template-fixer
rows]`: `action=needs_review` on 468 `cta_improvement` + 2 `responsive_fix` (the two "missing
spec.slot_name" refusals the 302 roster comment itself names); `fixed=false` WITHOUT action on
226 `spacing_fix` + 72 `responsive_fix` + 1 `instance_scope_conversion`. **Zero overlap in either
direction.** The disconfirming result — an idempotent reason carrying `action=needs_review`, or a
refusal without it — did not occur. So honouring the key would have parked 470 refusals and
blocked 0 legitimate completions.

**Who reads `action`?** `grep -rn '\["action"\]' --include=*.go platform/orchestration` → no
reader of a RESULT's action key (the hits are ExecutionContext/headers/config). No live workflow
`condition` mentions `needs_review` (query in RUNBOOK) — the two agent rows that match the string
are step NAMES (`mark_needs_review`). **The comment at `fix_component_template_action.go:58`,
"action:needs_review … is what stops the dispatch loop recording the work item as done", is FALSE**
— nothing stops it. Candidate LANDMINE (filed below once the diagnosis lands).

**"Does another route do the work?" (the bug's open question) — partially, and only for the
DESTINATION class.** Graded at the artefact via `page_component_history`:
- robot-hands.com/index, items 33cf3c40 + 233387b5 (08-17 12:34/12:37, "both CTAs → matchmatrix"):
  TRUE when filed — `call-to-action` content_data had `secondary_cta_url=/tools/matchmatrix/…` from
  08-16 14:34 through 08-17 14:23; at **14:50:36** it became `/gripper-catalog/index.html` with a
  `secondary_cta_target_title` (the resolver's signature, `setCTAField`). Live page confirmed
  distinct 08-19 ~16:00Z. So the destination defect was repaired ~2h later by the deterministic
  CTA machinery (resolver / `cta_links_stale` recompute), not by the item — `[INFERRED]` which of
  the 12:32 `cta_links_stale` rerender (cancelled for index) or the 14:35 `section_data_resolved`
  rerender carried it; not chased further, the point stands either way.
- gaswholesalers.com/index (08-15, label "Request a Rack Pricing Quote" → calculator): hero
  `cta_text` went "Learn how to buy wholesale fuel" (15:54) → "Get in touch about supply" (16:38)
  → "Work out your break-even volume" (20:45) — copy churn from content_rewrite items the same
  audit filed, URL unchanged throughout. The label now matches the destination by accident of a
  whole-page rewrite, not by anyone acting on the CTA finding.
- LABEL/COPY-class findings (brief-fidelity "no Discover/Explore/Learn More", "CTAs should be
  task-oriented") — nothing deterministic touches them; `[UNMEASURED]` whole-page content_rewrite
  may incidentally rewrite them.
**So: the 468 are NOT "468 unimproved CTAs" — and they are not "done by another route" either.
Destination defects have a deterministic repair path that runs regardless of the item;
copy defects have no handler at all.**

**Estate context found.** (1) `improvement-loop.md:51` in the concept register already records
that component-template-fixer "PUNTS on CTAs (cta_improvement → needs_review)" — known for months
as a fact, never as a bug. (2) The `copy_quality_two_stage` lane received, TODAY, a CONTRIB from
the `bugfix_277` lane asking whether its stage-2 (LLM → per-component `field_updates`, never
applied, human checkpoint) can be aimed at "one named component with one named defect" — because
`277` (`required_fields_missing`) and `301/083` (owned-page targeted edits) are stuck on the SAME
missing piece: something that turns "this component is wrong in this way" into a `field_updates`
payload for `section-editor` (`apply_section_edit`, 220 ok / 5 failed lifetime). `cta_improvement`
is a THIRD queue with that exact missing piece. Unanswered as of `HANDOFF_2026-08-19` (copy-editor
lane, 17:04). **Not building that here** — it is a new LLM write-route into live pages and the
copy-editor lane's proposal-only posture is an owner-level decision; but the plan must route
`cta` findings so that a future handler can pick them up, and must say so to that lane.
(3) `misdirected_cta` discovery check → `page_rerender` reason `cta_links_stale` → `applyCTARecompute`
is the live deterministic repair for the destination class on the `ctaFieldNames` components.

**Diagnosis loop fired** ~15:55Z: intake `31375905-…`, run `b218f39d-…`; coverage clear; loop
dispatch (not direct publish). Origin is 325 commits behind HEAD; mechanism is on origin.

## 2026-08-19 (session 1, evening) — diagnosis CONFIRMED, Layer 1 LIVE+PROVEN, Go half committed

**Diagnosis verdict** (run `b218f39d`, complete 16:27Z, 3 iterations): **CONFIRMED**, grounding on
the same four points I had read — the switch arm, `handlerReportedFailure` reading only
`resp["status"]`, the live `check_needs_rerender` condition, and `CompleteWorkItemAction`'s UPDATE.
Symptom coverage: all four sub-claims `[explained]`. It did not answer the "other route" question
(that was a request, not a symptom — grade it myself, done above).

**Layer 1 — migration 495, APPLIED 20:02Z** (`record-only` in the ledger, note says applied by hand
after a rolled-back probe). Pre-image guards: exactly one live row; `check_refused`/`park_refused`
absent; `check_needs_rerender.else_step = compose_note`; compose_note prompt ends with the expected
sentence. Post-image verified by `DO/RAISE` and by re-reading the row (updated_at 20:02:22Z).

**Layer 1 PROVEN by a real dispatch** (not by reading config): probe item `4dca88ba` inserted on
`system.internal` (status `claimed`, max_attempts 1, created_by `bugfix_323_probe`), then the REAL
`component-template-fixer` dispatched at it via a confirmed-publish kcat (payload in the container
command, `PUBLISH_OK` echoed; corr `64f89f97`). Run COMPLETED 20:04:51Z; `collected_data` keys show
the path `ensure_site_record → apply_fix → check_needs_rerender → check_refused → park_refused →
compose_note → append_note → complete`; item went `claimed → needs_human_review`, `error` = the 495
literal, `attempt_count` 0, `handled_by` generic; doc_notes entry titled `## refused:
cta_improvement`, Fix line = the reason verbatim, "Verified: work item was parked for human review".
Probe item and note DELETED afterwards (counts 0/0). The disconfirming outcome — status `complete`,
or the note titled "no-op fix pass" — did not occur. ⚠ `park_refused` has no `output_field`, so
`collected_data.refusal_parked` is null; the step's presence in `collected_data` is the evidence.
Harmless; noted so nobody "fixes" the absence.

**Why no proof via the build-dispatch-loop:** pointing the loop at a real site dispatches EVERY open
item on that site (other lanes' work, credits, page writes) — not an acceptable side effect for a
proof; the direct-agent dispatch exercises the identical fixer workflow.

**Go half** (Layers 2+3) committed `0e4622bab` (+ gofmt fix-up `f2525b3c8`) by pathspec with
`Council-Submitted: 92829711-aecb-4e1a-8457-d011b4a635af` (round 1, 20:08Z; submission JSON beside
this file). Full `./platform/orchestration/actions/...` suite green; lockstep test mutation-proven
(both halves RED on re-adding `cta`; restored GREEN). The pre-commit "architecture signal"
(migration + platform code in one commit → staged rollout order) is answered: 495 is
independent of the Go half and was applied FIRST on purpose; the Go half is inert until the roll and
needs no config to precede it. The pattern-check `logged-model-output` hit at
`fix_component_template_action.go:1602` is the 283 lane's judged branch, not this change.

**LANDMINES entry appended + committed from a TEMPORARY INDEX (`43ee7f196`)**: another session had
an uncommitted update to the promoter entry in the same file; a pathspec commit would have taken it
as a same-file passenger, so `git read-tree HEAD` into `GIT_INDEX_FILE`, `git apply --cached` my
hunk only, `write-tree`/`commit-tree`, `update-ref HEAD <new> <old>` (compare-and-swap, so a
concurrent commit would have failed it rather than been overwritten). Then `git reset -- <file>` to
re-sync the shared index — after `update-ref` the stale index entry read as a STAGED REVERT of my
hunk (`MM`), which a bare commit by anyone would have shipped. `landmines-verify-dispatch.sh` run
(first attempt died on a kubectl EOF, second dispatched 2 — mine and the other session's edit).

**Still owed:** council verdict read + acted on; bug file status; WRONG_CALLS; 016b §9; register
entry; CONTRIB to `copy_quality_two_stage`; README; memory.
