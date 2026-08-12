# HANDOFF 2026-08-12 — fact-assignment front: everything built is LIVE and thrice roll-verified. Cold-start for a fresh chat

**Supersedes `HANDOFF_2026-08-11b_…` (and its two UPDATE blocks) as the entry point.** The
08-11 job queue is fully executed, reviewed and applied; this file exists because a fresh
chassis roll (`v1.0.1290`) needed re-verification and that re-verification found something.

**There is no verdict-gated work and nothing half-applied. Read §3 for what is actually left.**

Site id: **`199733a8-ac9c-4c30-b2ce-65ecdac6f3bd`**. Verdicts by CORRELATION, never
`doc_notes … LIMIT 1`. **This is ONE OF TWO fronts in this directory** — the other is
`HANDOFF_2026-08-09_sweep_front_continue_here.md` (read it before dispatching at the site;
its 2026-08-11 addendum records what my census left it).

## 1. Verified live state (all checked 2026-08-12, not inherited)

| thing | state |
|---|---|
| Chassis | **`v1.0.1290`**, pods `agent-chassis-cc7b7f7b8-8tjhm`/`-vj2rt` |
| The three detections | **LIVE, THIRD roll survived.** Binary probe both replicas: `FACT_ASSIGNMENT_ABSENT`/`RECOMPOSE_INTENT_NOT_REALISED`/`PLAN_PAGE_MERGE_LOSSY` →1; CTRL `FACT_CARRY_UNMATCHED_SECTION` →1; NEG `RECOMPOSE_INTENT_UNREALISED` →0 |
| Seed 385 (planner sees `recompose_pages`) | APPLIED + row-verified, prompt 20,445 chars, marker + range present. `Council-Reviewed: 62d2463f` |
| Seed 386 (writer: no invented commitments) | APPLIED + row-verified, 13,974 chars. `Council-Reviewed: d1e8c36e` |
| Seed 387 / 388 (nested-contract rule, both rosters) | APPLIED + row-verified, 8,695 / 8,732 chars, **377 cache breakpoint still at char 174** |
| Current plan | `40a66d3a-b80e-4f92-9033-c6de1f43bcd1` (the census replan, corr `e74974b3`) |
| Register | **PBP-041** names the `spec.recompose_pages` seam and its two readers; PBP-037 carries the `section_plan` nested contract |

**How to re-verify after the NEXT roll, and one thing that has changed:** CLAUDE.md's
preferred route is `kubectl logs -l app=agent-chassis --tail=300 | grep 'build provenance'`
— **on 08-12 that line had already ROTATED out of 15h-old pods**, so it only works on a
freshly-started pod. Fallback that did work, with controls in the same breath (never
`strings`, never a discovery grep):
`kubectl -n ai-persona-system exec <pod> -- grep -ac "<literal>" /proc/1/exe` for each of the
five literals above. A zero exits non-zero, which is itself the signal the zero is real.

## 2. ⚠ THE TRAP FOUND ON 08-12 — read this before quoting any census figure

**The `content_duplication` fact-overlap count falls when a numeric fact's REGISTERED VALUE
drifts, with no repair whatsoever, and it moves in the direction that looks like success.**
Now a LANDMINES entry (synced to `doc_notes`, rows verified 13:04Z — and note
`landmines-sync.py --check` reports all 426 entries stale, so trust the DB rows, not the
checker). Measured here: the evidence base refreshed overnight, **four facts drifted**
(F1 21→22, F9 9545→9706, F10 8297→8468, F11 244→266, F12 264→283) while **zero pages
rebuilt**; every chart still saying "21" stopped matching F1 and each pair it fed lost a
fact. A pair at exactly 3 shared facts leaves the count entirely on one drift.

**So: pin ONE evidence dump and use it for both halves of any comparison.** The 08-11
census's 34→9 is sound because it did exactly that (before any rebuild ran) — that was luck
of method, and it is now the written practice. The instrument is
`fact_overlap_census.py` (a faithful port of the shipped check's fact half); if the
scratchpad copy is gone, re-derive it from
`platform/orchestration/actions/discovery_checks/check_content_duplication.go`
(`findResidue` + `assignFacts` + `containsStandalone`, ≥120-char usable floor, cross-page
only, ≥3 shared). **Read the pair COMPOSITION, never only the total** — two censuses can
both say 9 while naming different pairs, which is what surfaced this.

## 3. WHAT IS ACTUALLY LEFT — three items, in priority order

1. **The live recompose proof (the arc's last measurement).** Seed 385 makes the planner SEE
   `REDESIGN REQUESTED` on any page named in a `needs_site_plan` item's
   `spec.recompose_pages`; nothing has yet proven the marker changes what the planner emits.
   **Do it on the next GENUINE redesign need — do not manufacture one** (a replan of this
   site is a build dispatch and still a phantom-page generator; the sweep front hand-archives
   after). Dispatch with the intent in BOTH the field and the briefing (the operator rule
   still stands until this is proven), then read whether the marked page's composition
   deviates while unmarked peers hold. On success: retire the prose escape's load-bearing
   status in seed 362's instruction (a follow-up seed), update the LANDMINES 2026-08-10
   recompose entry, close the arc in `features_open/012`.
2. **Centralise the exactly-one-active-row guard (approved-round advisory, architecture
   seat).** The guard now exists **verbatim in four seed files** (385/386/387/388):
   `IF (SELECT count(*) FROM agent_definitions WHERE type='<x>' AND is_active AND
   COALESCE(is_snapshot,false)=false AND deleted_at IS NULL) <> 1 THEN RAISE …`. The seat's
   words: "fine to ship this instance, but the pattern deserves centralising before a 5th
   copy diverges." Shape to consider: a DB function (`assert_single_active_agent(text)`) or
   an assertion in `scripts/migration/run-migrations.sh`. **This is a shared-mechanism change
   → it wants its own council round**, and a register entry if it becomes callable.
3. **The census's owner items (his calls, recorded not actioned).** The **215 revisit trigger
   TRIPPED** (`PLAN_PAGE_MERGE_LOSSY` = 2, both composed-vs-composed guide-pair merges, all
   four page rows live) — noted in `bugs_open/215` + README. The **evidence-chart composition
   question**: the same register-derived chart serves on index, capabilities and
   digital-asset-recovery, which is 3 of the residual pairs — does it belong on all three?
   Compliance findings 3.5/3.6 remain unscheduled.

## 4. What the census actually established (so nobody re-measures it)

Replan `e74974b3` → plan `40a66d3a`: 18/18 singular built pages preserved by name+order (seed
362 proven; zero carry rows — the mechanism, not the safety net); all 9 writer-visible facts
assigned one-place-each; rebuilt pages state assigned facts in register-mandated FLOOR form
("10+"/"12+"/"0" — invisible to a value-matcher by design, read the copy) and nothing
unassigned. Whole-site pairs **34 → 9** on one pinned dump. The 8 fact-blind sites showed zero
writer/build/planner orchestrations all round. `bugs_open/151`'s candidate-1 arc is
**measured-done** (file stays in `bugs_open/` per the 08-06 owner ruling).

## 5. Traps this front pays for repeatedly

- **`orchestration_states` prunes ~24h** — every orch id quoted in NOTES is dated for that
  reason; re-run censuses, never quote them as live.
- **A failed `needs_page` item can be lying about the work.** On 08-11, THREE builds reported
  `deploy_page`/`call_content_writer` "failed" (`message validation failed (code:
  CHILD_ORCHESTRATION_FAILED)`) on a **SUCCESS** result while the page had persisted and
  deployed. Check the page row and `page_components.updated_at` before re-dispatching. Not
  squarely 207/216/217; deliberately NOT filed (no root cause diagnosed — a filer should run
  090 or declare the substitution).
- **A full rebuild re-resolves from source, so it can revert another lane's `content_data`
  patch** — mine reverted bugfix-235's logo fix on index (`bugs_open/235` addendum; the 238
  resolver-path family). Check who has patched a page's data before rebuilding it.
- **`LANDMINES.md` is the fleet's most-appended file and is a same-file-passenger magnet** —
  twice in three days my appends rode into another lane's commit (`f8ca05594` on 08-12).
  Nothing is lost, but verify at HEAD (`git show HEAD:<file> | grep -c`) rather than assuming
  your pathspec commit carried it.
- **`097_TRIGGER_council_review_v1.sh` prints validation errors and exits 0** — a grep for
  `SUBMISSION_CORR` alone reads a rejected submission as silence. `operation` is an enum:
  `modify|add|remove|config_change` (a new file is `add`, not `create`).
- **After a resubmission, watch the RUN, not the correlation** — a corr-keyed watcher matches
  the previous dead run's terminal status. The 012 round's first run was reaped un-reviewed
  (kafka `context canceled` at `review_mission`, the 029 hung-spawn family); the resubmission
  reviewed normally.
- **Council queue latency is ~30–40 minutes**, not 2. A missing orchestration row is latency.

## 6. Commit trail (08-11 → 08-12, this front)

`c3b424c8e` census complete · `fc01fdbc2` 012 built+submitted · `28436b190` seeds 386/387
submitted · `009aa7325` + follow-up seeds REVISE answered (388, guards) · `73d89c266` seeds
APPROVED+APPLIED · `d110da193` the reaped run · `e22ebe635` 012 REVISE answered (PBP-041,
rollback file) · `fd8731043` 012 APPROVED+APPLIED · `10038434f` post-roll v1.0.1290 +
the drift landmine (its LANDMINES half rode in `f8ca05594`) · this file's commit.

---

## UPDATE 2026-08-12 afternoon — §3 gains an item, and it is the urgent one

I ran the factless arm's last available test (rebuild `production-backend-engineering`: all
four sections `[]` in the plan, copy still stating 6 facts). **It was REFUSED at
`validate_content` — 20 blockers, all `unrendered_template`** — so:

- **The factless arm is still UNPROVEN on that page** and the 4 residual pairs it feeds are
  still there. Nothing persisted; no damage (stored sections still 08-09, page still
  `needs_rebuild`, live page untouched). Re-run the test once the defect below is fixed.
- **A NEW fleet-wide defect was surfaced and FILED to the diagnosis loop:**
  `RUN_CORRELATION_ID=b885a92e-d308-4b9c-99ee-306ca2f6b373`. The assembled `page_html`
  carries `mechanism-flow`'s Go template control structures verbatim (`{{if .eyebrow}}…{{end}}`,
  `{{range $s := .steps}}`) with the field placeholders inside them already substituted. All
  four writer LLM responses are clean of `{{`, so the leak is AFTER generation, in whatever
  assembles `collected_data->'page_content'->>'response'` inside `page-content-writer`. The
  type is new since **08-11 15:39** across **3 domains**, against a recorder that has 157 rows
  since 07-14 (so this is not the instrument's birth). **Read the loop's verdict before
  rebuilding any page on any site.**
- **Seed 386 is named in that filing as a chronological suspect** (applied 08-11 12:36Z; first
  row 15:39Z; the writer prompt is fleet-wide) even though its mechanism does not fit (one
  prose sentence, no braces; and the model output is clean). **If the loop CONFIRMS 386, the
  rollback is ready:** `agent_definitions_bak_386` + the `agent_definitions_backup` snapshot,
  both verified pre-update. Do not pre-emptively roll it back — that would remove an
  owner-approved control on a guess.
- Related, for whoever fixes it: `bugs_open/149` §B1 — the registered `unrendered_templates`
  discovery check is configured in NO agent and has never run, so nothing sweeps pages that
  already serve this. That is the detection half if the loop confirms a renderer defect.

---

## UPDATE 2026-08-12 late — the 090 run finished and produced NO locatable verdict. The halt STANDS.

`b885a92e` completed (5 iterations, bundles only). **No conclusion exists that this session
could find** — no non-`bundle` artifact on the correlation, `final_result` NULL on all three
rows, no `doc_notes` row for it, no `conclusion`/`verdict` column on any diagnosis table, and
the final bundle contains no verdict language. **Do not read that as "refuted" or "no problem
found".** Nothing has cleared seed 386 and nothing has named the cause, so:

- **The page-rebuild halt stays in force** (any site, any page) until someone establishes the
  cause. Rebuilding risks parking pages at `needs_human_review` — it does not corrupt anything,
  because the gate refuses before persisting, but it wastes builds and buries the queue.
- **Before re-firing 090, find out where its conclusion is supposed to land** (the fixloop
  091/090 read-out step in `fixloop_eg_dartsonline/`). A second run costs the same and may
  land in the same silent place. If a conclusion genuinely is not being written, that is its
  own defect and worth more than this one bug.
- **Correct my symptom before reusing it.** The loop's own check asked whether the
  `CONTENT_VALIDATION_BLOCKER_DETAIL` rows belong to `page-content-writer` runs and got
  **0 rows** — `validate_content` is a **page-build-handler** step, so my "inside
  page-content-writer" phrasing was unverified inference and may have misdirected the run.
- **The strongest lead is the bundle's own in-scope list, not my guess:**
  `multipage_actions.go` (`AssemblePageAction`, `buildRenderContextFromCollectedData`,
  `extractFieldValue`, `cleanHTMLStructure`), `assemble_from_library.go:AssembleOutput`,
  `datahelpers.CleanHTMLString`. Hypothesis to TEST (not a finding): a field-substituting
  assembler that never executes `{{if}}`/`{{range}}`. Unowned.
- Evidence that stands regardless: assembled `page_html` carries `mechanism-flow`'s control
  structures with field values substituted; all four writer LLM responses clean of `{{`;
  the failure type is new since 08-11 15:39 on 3 domains against a recorder holding 157 rows
  since 07-14; seed 386 remains a chronological suspect with a verified rollback ready
  (`agent_definitions_bak_386`), NOT to be used on a guess.
