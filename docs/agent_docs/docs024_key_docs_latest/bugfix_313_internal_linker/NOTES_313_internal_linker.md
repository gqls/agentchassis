# NOTES — bugfix 313/298 (internal linker dead branch + cap)

Append-only, newest at the bottom. Missteps are the point — record them.

## 2026-08-19 — session start, re-verification, design

- **Ownership checked before anything** (`who-owns.py 313` / `298`): both filed by
  `bugfix_275_silent_row_caps`, whose handoff carries a "LANE CLOSED 2026-08-19 10:30Z" banner and
  lists 313→298 as standing tickets **nobody owns**. Diagnosis queue
  (`item_type='needs_diagnosis' AND status='awaiting_diagnosis'`): empty. No competing thread.
- **Bug still valid, re-verified live 2026-08-19** (not trusted from the 08-18 filing):
  - live row `93cffe67-baf4-4fb1-bec9-ba546fb24a54`, step listing via `jsonb_each`:
    `load_candidate_pages` `output_format=array`, `check_candidates`
    `condition='candidate_pages.count > 0'`, `then_step=load_specs`,
    `else_step=complete_no_candidates`. Matches seed `101` exactly.
  - `llm_call_log WHERE agent_type='internal-linker'`: **0 rows, all history**.
  - 20 open (`unresolved`) work items name it as handler.
  - 298 census re-run with the query's own predicate (incl. the HAVING): 26 sites with
    candidates, **8 over the cap**, median 11.5, worst **69**. (Filing said 24/8/12/68-9 — moved
    slightly with the fleet, same shape.)
  - `updated_at = 2026-08-19 07:51` on the row is the v1.0.1314 roll-window bulk write, NOT a
    targeted edit — the 313 file's own landmine; the step config read is what settles it.
- **Code re-read at the cited lines** (unchanged since the CONFIRMED verdict `c4aa3559`):
  `database_actions.go:129` returns the bare slice for array; `:132-145` object branch =
  `rows`/`count`/`columns` **+ first row flattened to top level** (this is why
  `target_page.page_id` works — worth knowing: a column literally named `count`/`rows`/`columns`
  would clobber the metadata; our columns don't collide, and the four legit `.count` readers in
  the fleet may be *relying* on the flatten when they alias `count(*) AS count`).
  `conditional_branch_action.go:275-284`: numeric arm, `ToFloat64(nil)` fails → WARN →
  `return false, nil`. `:397-412`: strategy 5 requires `map[string]interface{}`.
- **No 090 re-run, substitution stated** (owner ruling 2026-07-31): 313 has a CONFIRMED verdict
  already (first iteration, 08-18); a second run on the same mechanism is a duplicate round. The
  substitute is the first-hand re-verification above. (Sought the verdict's doc_notes row by
  correlation for extra colour: not present under `%c4aa3559%` — orchestration_states prunes ~2
  days and the note body evidently doesn't carry the corr string. The bug file's quotation of the
  verdict is the record.)
- **Template/render path checked before designing the fix** (the half that trades a dead branch
  for a broken prompt): `ExecuteLLMPromptAction` → `ExtractFields` (generic arm:
  `extractSingleField` + `UnwrapDeep`; a `{rows,count,columns}` map has no `type`/`result` keys so
  passes unchanged) → `RenderPromptTemplate` = plain Go text/template. So
  `{{range .candidate_pages.rows}}` + `{{.name}}` works; nothing shape-special in between.
- **Design inputs measured, not assumed:**
  - Fleet census of `conditional` step config keys (top-level AND nested, both spellings):
    exactly `condition`/`then_step`/`else_step`, 145 uses each → a declared ActionInputSpec is
    complete; the new key makes 4.
  - `conditional` registers **no** ActionInputSpec today; unknown config keys are ignored at
    execution (bugs_open/101 header in cmd/config-key-audit/main.go) → opting `check_candidates`
    into the flag pre-roll is inert, not breaking. No `_HOLD` needed.
  - `evaluateStringCondition` has two external callers (both test files) → keep its signature,
    add a strict variant.
  - Latest migration number in the dir: 487 → this lane takes **488**. (Re-check at commit time —
    other sessions allocate concurrently.)
- Worked shapes copied rather than reinvented: migration **484** (snapshot → count-gate →
  pre-state gate incl. refuse-to-double-apply → id-scoped jsonb_set → DO/RAISE verify with
  controls); **446** (`[…truncated]` marker); WFA-009 (the opt-in-required precedent and its
  scope argument); WFA-010 (`GetBoolFieldLoud` for the new key).

### Correction discipline for this file
Anything above that later turns out wrong gets a dated `> **CORRECTED:**` block here, and if it
was a durable claim in a commit/bug file, a WRONG_CALLS.md row.

## 2026-08-19 (later) — implementation record

(appended as work lands; see RUNBOOK for the exact commands)
- **Implementation landed (all uncommitted so far):** migration `490_…fail_loud.sql` + ROLLBACK;
  `conditional_branch_action.go` opt-in `fail_on_non_numeric` (wrapper keeps the 3-arg lenient
  signature; spec registered for both names); `cmd/config-key-audit/conditionalshape.go`
  (`--array-producer-conditions`) + tests + `scripts/audit-array-producer-conditions.sh`.
  Tests green; both flag guards MUTATION-PROVEN (default flipped → lenient test RED; error arm
  disabled → fail-loud + AND-propagation tests RED; both reverted, clean run green). First live
  run of the audit: **191 agents, 145 conditionals, exactly 1 finding = internal-linker** — the
  motivating case, independently re-deriving 313's census.

> **CORRECTED 2026-08-19 (same session):** the line above saying "this lane takes **488**" was
> stale within ~3 hours — 488 was taken TWICE concurrently (301 lane's applied migration + 320
> lane's renumbered HOLD) and 489 applied too, exactly the "486/487 taken concurrently" event
> repeating. Renumbered everything to **490** BEFORE committing, so no damage. The check that
> caught it: re-listing the dir + `schema_migrations` before commit, which the NOTES entry itself
> prescribed. Number allocation on this tree has ~hours of shelf life.

- **Pre-existing red test at HEAD, not ours:** `optional_budget_cron_parity_test.go` fails on
  clean `git archive HEAD` — `save_page_meta_description` (320 lane, commit `aeccfc595`) declares
  5 optional keys, cron literal says 0. The 320 lane is ACTIVE today → contributing the finding to
  them, not fixing competitively.

## 2026-08-19 (afternoon) — council submitted, 321 interaction, pre-commit checks

- **Council: SUBMITTED, corr `aef24a7f-2992-4d4f-a6e0-422cd77fcca3`** — admitted without FORCE=1
  (the platform/ files put it in scope; the config-only half alone would have been refused —
  bugs_open/314). Committing with `Council-Submitted:` per the 2026-07-30 norm; budget ~30 min for
  the verdict, find the run by payload not printed id.
- **The pre-existing parity failure resolved itself**: `go test ./cmd/config-key-audit/` is green
  at the current tree — a session fixed the `save_page_meta_description` literal in the interim.
  The planned contribution to the 320 lane is MOOT; nothing to do.
- **Cross-session message from the 321 lane (item_key collisions), and it names a REAL residual of
  this fix:** `create_rewrite_item` carries `item_key_prefix: internal_link` with NO
  `item_key_suffix_field`, so once 490 revives the loop, an N-link plan files ONE content_rewrite
  item and silently drops N−1 (idx_swi_dedup collision — 321's exact class; plan_links asks for
  1–3 links, so the loss is up to ~2/3 of the agent's output). **They fix it independently**
  (their migration, next free number, single jsonb_set on the create_items_loop subtree — fully
  disjoint from 490's paths and gates; apply order immaterial, confirmed both sides). Replied
  2026-08-19: proceed independently; joint disconfirming check = an N-link plan produces N items
  keyed `internal_link_<domain>_<source_page>`. **Until theirs applies, expect ~1 item per plan —
  do not misread that as 490 failing.**
- **Fresh chassis build v1.0.1315 (12:15Z) predates this commit** — the Go halves (WFA-019 flag,
  WFA-018 audit changes) ride the NEXT build; the audit is runnable from the tree meanwhile
  (`go run`), and migration 490 is live-on-apply regardless.
- **Pre-commit checks:** same-file passenger scan on the four shared modified files —
  `git diff --numstat` shows additions only, counts matching my edits exactly (4/2/21/46+5).
  **HEAD + only-my-files** build: `git archive HEAD` + my six code files → build green, tests
  green (`cmd/config-key-audit`, `actions`) — my change does not lean on other sessions'
  uncommitted work.

## 2026-08-19 (late afternoon) — committed, applied, and the detector's first scalp

- **Committed `5315c8a19`** (15 files, pathspec, `Council-Submitted: aef24a7f`), after: same-file
  passenger scan clean (additions only, counts match) and a HEAD+only-my-files archive build+test
  (green — no dependence on other sessions' dirty tree state).
- **Migration 490 APPLIED** 2026-08-19 (direct psql): snapshot row, gates passed, **UPDATE 1**,
  verify incl. PREPARE parse-check passed, COMMIT. Recorded via `--record-only` (note names this
  lane + commit). Post-apply read-back: `output_format=object` · no row LIMIT · truncation marked ·
  template on `.rows` · `fail_on_non_numeric=true` · condition/routing untouched.
- **Post-fix audit run: internal-linker CLEAN — and the detector caught a NEW second instance that
  did not exist this morning.** `meta-description-backfiller.check_has_pages` tests
  `pages_missing_meta.count > 0` against an `output_format: array` producer (fleet moved
  191→193 agents, 145→146 conditionals — the 320 lane seeded the agent today). So "blast radius
  is ONE" was true at filing and was already going stale the same day; the class detector is the
  half that mattered. Contributed to the 320 lane's NOTES + committed `ab02e8145` + messaged the
  likely session. **My RUNBOOK's "clean fleet expected post-490" line is therefore wrong as
  written** — the honest pass is "0 findings once the 320 lane fixes theirs"; the audit exits 1
  until then and that exit is CORRECT.
- **313/298 bug files: FIXED banners added** (`9be3a2202`); both stay OPEN pending their own
  §How-to-verify — the first `plan_links` row in `llm_call_log` (still 0), the prompt listing
  pages, a run passing `check_candidates`. Natural traffic ~2/day, 20 open items queued.
- **Council round mid-flight** at `review_guardian` (orchestration row found by payload, per the
  runbook — not by printed id). Verdict owed a read; background watcher armed for
  verdict-or-first-run.
- **321 lane landed their `--loop-sitewide-item-keys` mode in main.go on top of my commit** — the
  two modes coexist; their internal-linker suffix migration is still to apply on their side.

## 2026-08-19 (evening) — council round 1: REVISE, answered by measurement; round 2 dispatched

- **Verdict: REVISE, gated by prior_art_librarian (HIGH)** — the round-1 submission failed to
  carry the internal-linker vs internal-link-resolver landmine in grounded_in, so the reviewer
  could not see the correct-agent question had been settled. Fair catch: it WAS settled in the
  lane docs (RUNBOOK §1) and not in the submission. Round-2 evidence, newly measured:
  handler_agent='internal-linker' → **75 work items (50 complete needs_internal_links)**;
  handler_agent='internal-link-resolver' → **0**; resolver llm_call_log → **0**. (⚠ counts move
  with the reaper — 82→75 between morning and evening reads; direction stable.)
- **Two objections found real things:**
  1. **debug_historian: 490's `snapshot_agent` fires BEFORE the double-apply guard** — a re-run of
     the applied file writes one redundant backup row labelled 'pre-update' before the first gate
     RAISEs. **Accepted explicitly** (the objection's own second branch): the file is applied and
     recorded, and editing a recorded migration trips the checksum drift guard. Cost bounded: one
     spurious `agent_definitions_backup` row per accidental re-run, zero mutation. **Adopted
     forward: snapshot AFTER the pre-state gates in every future migration this lane writes.**
  2. **reuse_agent forced precision on the budget claim** — "1 optional key against N=10" was
     wrong: the RFC_022 counter counts `spec.Optional`; `fail_on_non_numeric` is ConfigKeys-class,
     so the budget count stays 0 and the parity literal is untouched (suite incl. parity test:
     PASS). Claim retracted in round 2, PLAN corrected in place, **WRONG_CALLS.md row added**
     ("grep the METER before claiming you are on it"). Residual surfaced for the counter's owner:
     ConfigKeys-class opt-in settings are structurally invisible to the budget as built.
- Remaining objections answered by existing measurements now quoted in grounded_in: repo-wide
  census proves "first ActionInputSpec" (no spec under any conditional spelling); 0 prior council
  rounds on conditional_branch_action.go; 124/WFA-009/WFA-010 precedent quotes attached; edit-3
  sketch now shows the strict returns INSIDE the numericOps loop (the null-probe carve-out is
  structural); pod-probe with controls added to WFA-019 verify-later; cron follow-up shown tracked
  in three places.
- **Round 2 dispatched with `RESUBMIT_CORR=aef24a7f`** (same correlation, trail accumulates).
  Note for the reader: the reviewers' round-1 checks read back the POST-apply state because 490
  applied between submission and review — the estate's review-after-the-fact norm, stated in the
  round-2 header.

## 2026-08-19 (close of day) — APPROVED, and the fleet goes clean

- **Council round 2: APPROVED, all reviewers** (report read in full). Advisories, each closed:
  - tooling_provenance (medium): "register entries claimed same-commit but not a plan edit" —
    the entries DID land in commit `5315c8a19` (`git show --stat`: workflow-authoring.md +21,
    000_concept_index.md +2); the plan under-declared them to stay within 8 code edits. Delivery
    verified, on record here.
  - guardian: "confirm no other live step already carries the key" — ran the nested-aware census:
    **2 carriers**, both intended: `internal-linker.check_candidates` (mine, via 490) and
    `meta-description-backfiller.check_has_pages` — **the 320 lane acted on the contribution
    within the hour**: fixed their producer to `object` AND adopted the tripwire.
  - architecture: `evaluateStringConditionMode` is unexported (lowercase) — cannot become an
    external contract; confirmed.
  - debug_historian: the accepted snapshot-ordering gap stays visible at low severity — agreed,
    it is recorded here and in the round-2 rationale.
- **Fleet audit now CLEAN: 193 agents, 147 conditional steps, 0 findings, exit 0.** The class is
  closed at config time fleet-wide, with both known instances repaired by their owning lanes.
- Trailer state: earlier commits carry `Council-Submitted:` and are auto-credited by 098 now the
  correlation is approved; commits from here may carry `Council-Reviewed: aef24a7f…` (verdict
  READ and APPROVED — never before).
- **SUMMARY written** (`SUMMARY_2026-08-19_313_internal_linker.md`) — the milestone is real:
  approved + applied + fleet clean + second instance fixed by its owner, same day.
- **The lane's single remaining gate: the artefact proof.** `llm_call_log` for the linker is
  still 0 all-history; ~20 queued jobs, natural cadence a few runs/day. The verification is
  RUNBOOK §"Verify the fix at the artefact". On its first pass: move 313 + 298 to `bugs_closed/`
  (numbering preserved), close the lane. Post-next-roll checks for the Go halves live in
  WFA-018/019 §verify-later.

## 2026-08-19 (evening, later) — the Go halves are LIVE (v1.0.1316), and a second production repro on record

- **Chassis v1.0.1316 rolled; both Go halves are LIVE — verified first-hand, not from the peer's
  report:** `git merge-base --is-ancestor 5315c8a19 07eeba4a1` → ancestor; pod
  `agent-chassis-5ddd9744-86nqf` (image v1.0.1316) `/proc/1/exe` carries the new error literal,
  with both controls behaving (old literal PRESENT, fake literal ABSENT). The 320 lane probed the
  same independently (absent on 1315, present on 1316) before telling us. **The
  `fail_on_non_numeric` tripwire on `check_candidates` is armed as of this roll.** Register
  WFA-018/019 statuses updated to LIVE.
- **A second, fully-evidenced production instance of 313's mechanism** from the 320 lane's message:
  their backfiller's first run COMPLETED with no error while `collected_data` held an
  11-element `pages_missing_meta` array and the `.count > 0` gate routed to else — correlation
  `56822c5b-5207-4e39-b752-1578d61323d9`. They inherited the shape by modelling their workflow on
  internal-linker: the copy-inheritance path is exactly why the class-closer matters.
- **Their count-ish-condition census answered:** the producer join they left to this lane is
  already mechanical and already ran clean — `scripts/audit-array-producer-conditions.sh` joins
  every one of the 147 live conditionals (their 10 listed conditions included) to its producer,
  nested steps included, and returns 0 unsatisfiable. Honest scope note: conditions whose
  producer is a Go action (e.g. `unswept_areas.count`, whose action sets count in code) are
  SKIPPED by design — statically unjudgeable, and the runtime tripwire is the layer that covers
  them if opted in.

## 2026-08-19 (night) — PROVEN AT THE ARTEFACT. Both bugs closed. Two things this lane had wrong.

Picked up by a later session (named `bugs_open/taken 313 and 298 and 314`) to supply the one
remaining gate. It is supplied, and closing it exposed two errors in what this lane wrote.

**The proof.** Canary correlation `50ea3037-4602-40f5-b7de-a0b3a537ce39`, 2026-08-19
21:18:35→21:19:33Z, webdesign.co.uk (the worst site: 69 candidates), target page `about`,
chassis v1.0.1316. All three arms, each against an established "before":

| arm | before | after |
|---|---|---|
| `plan_links` ever ran | 0 rows all-history (`step_name='plan_links'`) | **1 row**, 21:18:53Z, success, haiku-4-5, 21,946 in / 446 out |
| prompt rendered pages, not map keys | n/a (step unreachable) | `## Candidate Pages` → `### domains (/domains/index.html)`; no `rows`/`count`/`columns` |
| terminal step with non-empty candidates | every run ended `complete_no_candidates` | **`complete`**, `candidate_pages.count = 68` |

298's disconfirming arm specifically: **`tool-white-balance`, alphabetical position 69 of 69, is in
the rendered prompt** — structurally impossible under `ORDER BY p.name LIMIT 15`. Position 15
(`tool-csp-builder`) is present too, so the check discriminates rather than merely observing.
Downstream: 2 links planned → **2** `content_rewrite` items, distinct keys,
`created_by='internal-linker'`.

### MISSTEP 1 — "the proof should arrive within hours". It never would have.

This lane's SUMMARY and the 313 banner both said natural traffic (~2 runs/day, "20 open work items
queued") would supply the proof within hours. **Those 20 items are `status='unresolved'`, which is
TERMINAL** (`work_items_common.go:40-46`) — invisible to the promoter (reads `detected` only), to
the selector and to the atomic claim (`workItemDispatchableStatuses = {triaged, approved}`,
`work_items_common.go:172-175`). They can never dispatch. Worse, **this very bug parked them**: the
dead branch produced no-op "complete" runs, and the two-strike rule
(`load_work_item_actions.go:1336+`) counted its own successes as attempts and branded the third
re-raise `unresolved`. Fresh items arrive only on a `site-discovery-rotation-completeness` tick
(hourly, ONE site, 7-day per-site stamp) that finds new orphans.

**The cheap check I skipped:** read `workItemDispatchableStatuses` before calling a queue "open
work". A count of rows is not a count of *dispatchable* rows. Recorded in `WRONG_CALLS.md`.

*(Not a hazard for the fix: `unresolved` does not hold the `idx_swi_dedup` slot, so once those
terminals age past 7 days the rotation re-raises the findings and they dispatch against the fixed
config. The 20 dead rows need no revival.)*

### MISSTEP 2 — the verification query this lane wrote into BOTH bug files cannot come out true.

Both files nominate "a `plan_links` row in `llm_call_log` for `agent_type='internal-linker'` — a
table that has zero rows in all history today, so the 'before' arm is already established and
**cannot be faked**". It cannot be faked, and it also cannot be SATISFIED: after a demonstrably
successful run that query still returns **zero**, because the row was written under
`agent_type='generic'`. That column carries the DISPATCH context, not the workflow's agent type —
the exact sibling of the `orchestration_states.owner_agent_type` landmine this lane already knew
about and cited. We swapped one mislabelled-provenance column for another.

Had the proof been left to arrive on its own, the honest reading of this instrument would have been
"still not proven" *for ever*, on a fix that works. **The instrument that works is
`step_name='plan_links'`** — genuinely zero all-history, now 1. RUNBOOK corrected in place; the old
recipe struck through rather than deleted. New LANDMINES entry filed for the column.

[MEASURED for a hand-fired inline-workflow dispatch. Whether a loop-dispatched `spawn_agent` run
labels `agent_type` differently is **UNMEASURED** — and does not affect the proof, because
`step_name` settles it either way.]

### Lane state

Both bugs moved to `bugs_closed/`. `scripts/fire-internal-linker.sh` promoted out of a scratchpad
into the repo, because the next person to verify this agent needs it and "wait for traffic" is not
a recipe. The lane is CLOSED.
