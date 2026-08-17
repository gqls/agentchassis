# Handoff — the section-list assembler is lock-blind: every rebuild proposes a list without the page's locked sections, and only the terminal write guard stands between that and deletion

> ⚠ **285 IS AN AMBIGUOUS NUMBER** — filed the same day as
> `285_HANDOFF_2026-08-15_tool_improver_rewrites_a_shared_template_…` (an
> unrelated case). Refer to this one by slug
> (`section_list_assembly_is_lock_blind`), never by bare number, and
> `git log` the FILE PATH. **Ownership (owner, 2026-08-15 evening): the fix
> for THIS case is being implemented by a separate lane** — the filing lane
> (webdesign_uk_build_service) holds the locked chat box and runs the
> acceptance when the fix lands.

**Filed 2026-08-15** by the webdesign_uk_build_service lane, on the owner's
2026-08-15 ruling ("the improvement loop that tried to remove the chat box:
CHECK AND FIX IT"). Diagnosed through two 090 rounds — the first REFUTED the
obvious framing, the second CONFIRMED the relocated one. Both trails are
citable:

- Round 1 `c199c4bf-e433-4fa7-8bbf-c64b627e7373` — **REFUTED**: the accused
  `PlanSectionsAction` is a CONSUMER (`sectionsRaw := inputs.GetRaw("sections")`);
  its only `page_components` read resolves identity for names already in the
  input. Correction recorded in `WRONG_CALLS.md` (2026-08-15) and the lane's
  NOTES.
- Round 2 `d9f97c15-da88-459f-8fba-75add31227b2` — **CONFIRMED** at the
  composer. Verdicts live in the diagnose-agent orchestration's
  `collected_data->'verdict'` (the `diagnosis_artifacts` rows are iteration
  INPUT bundles, not output).

## The defect

`LoadPageSectionsFromSpecAction`
(`platform/orchestration/actions/load_page_sections_from_spec_action.go`) is
the single place the build/rebuild pipeline assembles a page's section list.
It has four source tiers — `site_plan_sections` (authoritative),
`site_specs.site_plan`, `pages.sections`, same-role-sibling synthesis — and
**none of the four reads `page_components` or any lock column** (round 2
citations: the tier-1 SELECT reads `component_name, assigned_fact_ids` only;
the tier-3 SELECT reads `pages.sections` only). A human-locked section that
exists only on the live page therefore CANNOT survive into the assembled
list, whatever the tier.

Downstream, `save_page_sections` diffs proposal against live rows and wants
to remove the missing section; the bugs_closed/058 write guard
(`loadActiveLockedRows`, predicate `NOT pageComponentAgentWritableSQL`)
preserves the row and files `lock_blocked_change` — so the page keeps the
component, but:

- the plan/cache narrative says the section is gone — measured live:
  `pages.sections` for webdesign.uk `contact` = `["hero","contact-info"]`
  while the locked `chat-input-box` row (component
  `7d3489c6-0586-491f-ab38-42a8f82b40f6`, `lock_type='permanent'`,
  locked 2026-08-11 14:48Z) sits in `page_components`. Every consumer that
  reads the list rather than the rows believes the page has no chat box;
- every improvement/rebuild pass re-attempts the removal and re-files
  `needs_human_review` noise (`lock_blocked_change:contact:chat-input-box`,
  2026-08-13T17:16Z, `blocked_action: remove`);
- **before the lock existed the same pass deleted the section for real**
  (2026-08-11 improvement sweep; the lane restored it by hand and locked it).
  The guard is the only thing that changed between then and now.

NOT this bug: the `lock_blocked_change` rows for `hero`/`call-to-action`
slots on the same site. Those locks exist for the CTA-destinations defect
(bugs_closed/268 family) and block content OVERWRITES of in-list sections —
different mechanism, same guard machinery. Round 1 refuted their use as
corroboration; do not re-bundle them.

## Root cause

The section list's assembler treats the CURRENT PLAN as the whole truth of
page membership. Human locks live on the live page (`page_components`), a
store the assembler never consults — so "the plan omits it" and "an operator
pinned it to the page" are indistinguishable to everything downstream of
assembly. The write guard (058) protects the ROW; nothing protects the LIST.

## Fix candidates, ordered by what closes the door

1. **Merge live locked sections into the assembled list, in the loader
   (recommended).** After tier assembly in `LoadPageSectionsFromSpecAction`:
   load the page's non-agent-writable rows using THE GUARD'S OWN predicate
   (`pageComponentAgentWritableSQL` → `datahelpers.AgentWritableSQLFor` —
   reuse it, do not approximate; the pin-predicate-vs-pool-predicate class is
   a filed landmine), and insert any row whose slot/identity is missing from
   the list, at its live `position` (append when position exceeds list
   length). This is the single chokepoint every build path passes through,
   and it makes the bad state (an assembled list missing a locked live
   section) unrepresentable at the source. Two alignment obligations in the
   same edit:
   - `specSectionFacts` must receive an aligned nil entry per insertion when
     tier 1 served, or the `len(specSectionFacts) == len(specSections)` guard
     silently drops the whole fact-scoping payload;
   - the pages.sections cache syncs (three sites in the function) must write
     the MERGED list, or the cache keeps telling consumers the section is
     absent.
2. **Make the plan producers lock-aware** (write locked sections into
   `site_plan_sections` at plan time). Correct long-term but there are
   multiple producers plus two legacy tiers the fix would never reach; does
   not close the door alone.
3. **Suppress the removal attempt save-side.** Already effectively the
   guard's behaviour; leaves the list/caches lying to every reader. Fails
   the owner's acceptance by construction.

## Two interactions the fixing thread MUST verify (candidate 1)

- > **CORRECTED 2026-08-16 (fixing lane, `7d9b7334a`):** this bullet mis-attributes
  > the function. `loadComponentNameResolver` is called by `ValidateSitePlanAction`
  > (`v3_site_actions.go:3407`) and `apply_gap_plan_action.go` ONLY — never by
  > `plan_sections` (`grep -n loadComponentNameResolver plan_sections_action.go` → 0
  > hits). On the page-build path a merged slot name resolves via `plan_sections`
  > Path 0 (`loadPageSlotComponentIDs`, slot→component_id from `page_components`,
  > no `component_level`/lock filter) or Path 1 (no level filter); a self-contained
  > tool is `ready`. So `bugs_open/282` is a co-requisite ONLY for candidate 2
  > (writing tools INTO the plan at replan) — not for candidate 1, which shipped.
  > Verified by driving the merged name through the fixture path in the loader
  > test and by reading the writer/render chain (below). WRONG_CALLS 2026-08-16.
  ~~**plan_sections' resolver excludes tool-level components from name
  normalisation** (`loadComponentNameResolver`, section/element only —~~
  deliberate: tools are placed by tool-deployer). A merged `chat-input-box`
  entry must survive plan_sections' planning/skip machinery
  (`persistSectionSkips`, the deferred/needs_new_component branches) all the
  way into save's input, or the merge is undone one step downstream.
  **This is not hypothetical — it is `bugs_open/282`'s exact mechanism**
  (016b §9, filed 2026-08-15 by the 407 thread): the same resolver ate every
  tool section the widened planner proposed, silently. Whatever fix 282
  lands (a `component_level='tool'` arm in the resolver, gated or not) is a
  prerequisite or co-requisite for candidate 1 reaching tool-level locked
  sections; coordinate with that thread. Verify by driving the real
  pipeline, not by reading alone.
- **The 058 guard's match path** (`matchLockedRow` — identity first, then
  slot, then kebab-normalised; see bugs_open/189 for the positional-slot
  duplication it was hardened against): with the section now IN the
  proposal, the expected behaviour is consume-and-keep (fresh copy discarded
  for a locked row, position kept, no duplicate row, item filed as overwrite
  rather than remove). Confirm no duplicate-slot regression on a page with
  positional slot names — 189's exact territory.

## How to verify (the owner's stated acceptance)

Drive an improvement/rebuild pass over webdesign.uk `contact` (single-page
path is fine) and assert, in order:

1. the PROPOSED section list handed to `save_page_sections` **contains
   `chat-input-box`** — not merely "the lock blocked the removal again";
2. `pages.sections` for contact contains it after the pass (the cache tells
   the truth);
3. the locked row's `rendered_html`/`updated_at` are unchanged (the guard's
   original 058 assertion — artefact, not status);
4. an UNLOCKED sibling section on the same page IS still rebuilt (don't fix
   this by never rebuilding anything — 058's own control);
5. the pass files no `lock_blocked_change … blocked_action: remove` item for
   the slot.

## Coordination

- The chat-input-box lock on `contact` STAYS ON until this is fixed and
  live (owner ruling 2026-08-15). The `a4cd5dc8` needs_human_review row is
  answered by this fix, not dismissed.
- Pending migrations 418/419/420 (another thread, bugs_open/276 class) edit
  the PLANNERS' component-loading steps — different step configs from this
  loader, but the same neighbourhood: check `who-owns.py` + live transcripts
  before editing, and expect their pre-state probes to keep reading
  "concurrent edit?" while both efforts are in flight.
- Owning lane + full session trail:
  `docs/agent_docs/docs024_key_docs_latest/webdesign_uk_build_service/`
  (NOTES 2026-08-15 evening/late entries, HANDOFF_2026-08-15b §2.1).

## Implementation — 2026-08-16, fixing lane `bugfix_285_lock_blind_section_list/` (session `390a1ae1`)

**Status: FIX COMMITTED `7d9b7334a` + round-2 `57336c127` (Go — inert until the next chassis roll), council
corr `79f70435-fadc-4e1b-b9d3-6d41f437f7fd` **APPROVED round 2 (2026-08-16 15:31Z)** after a round-1 REVISE
(arm-order confirmation, durable merge-skip trace, ROW/LIST loader lockstep test, no-locked-rows regression test — lane NOTES).
Advisory residuals from the approving round are in register LOCK-008. This file stays OPEN until the owner's five criteria above pass
post-roll — run by the filing lane, which holds the chat-box lock.**

What shipped (candidate 1, made shared — register **LOCK-008**, `register/locks.md`):
- `datahelpers/locked_page_sections.go` (NEW): `LoadLockedPageSlots(ForSite)` — the guard's
  exact predicate string `NOT AgentWritableSQLFor("pc.")`, never re-typed (a test refuses a
  `locked_at IS NOT NULL` lookalike); `build_status <> 'removed'` as a separate MEMBERSHIP
  condition (0 such rows). `MergeLockedPageSlots` — pure; pairs list entries with locked rows
  arm-for-arm as `matchLockedRow` does (slot exact → slot kebab → component function/name),
  consume-once, then inserts unpaired rows at their live position (clamped, ascending).
  `NormalizeComponentFunction` moved down; `actions` delegates.
- `load_page_sections_from_spec_action.go`: merge after any tier serves (NOT when none did —
  a locked-only list is neither plan nor page and a rebuild on it would delete unlocked
  siblings); `section_facts` kept index-aligned (nil at each merged index); result gains
  `locked_sections_merged` + `locked_merge_count`; best-effort on the locked query (Warn,
  proceed unmerged — the guard still protects the row). ONE jsonb-compared cache sync
  replaces the three per-tier syncs whose `sections::text IS DISTINCT FROM $1` was ALWAYS
  true (LANDMINES 2026-08-16).
- `check_section_source_drift.go` (ENABLED live): both sides through the same merge — else
  one drift item per fixed page (13 on day one). Loud if the locked query fails.
- Tests: 12 merge cases + predicate pin; the FIRST loader tests (sqlmock; MUTATION-proven —
  merge removed → fails alone on `sections`; facts alignment removed → fails alone on
  `section_facts`; no-tier case proves the locked query never ran); drift check ×3.

Fleet census that made this a class, not a page (2026-08-15, RUNBOOK C1–C3): 26 locked rows
(all `permanent`), **13 on tier-1 pages whose plan omits them** — contact + 12
loancalculator.co.uk calculators (`tool-1..4` positional slots); **5 remove-blocked items
filed 17:11–17:48Z that day** by `page-build-handler` (`spec_sections.source=site_plan_tables`,
list `[hero, ported-prose, faq, tool-cta]`); the tail-exile moved `index/tool-3` 4→6 and
`tool-settlement-calculator/tool-2` 3→5 within two hours.

Second correction (with the 282 one above): the proposal `save_page_sections` diffs is the
content WRITER's `sections_metadata` (`render_component` emits `component_id` +
`stored_slot_name` from `slot_name_from: current_section.name` → `compile_page_sections`), one
hop past `plan_sections`. That is WHY the guard pairs the merged entry: identity arm, then
slot — consume-and-keep, `overwrite` item (058's design; deduped by `item_key`), no duplicate.
All 26 live locks are `static`-field components → `render_from_template`, no LLM spend.

Consequence to expect after the roll: the `remove` items STOP; an in-list locked section
still draws/refreshes ONE `overwrite` item per pass (LANDMINES "a lock_blocked_change item does
NOT mean the copy differed"). A carry-stored-row shortcut for locked sections in the
writer/render loop would make them silent — left as LOCK-008's open review question.

How to close: RUNBOOK C5 in the lane dir — stamp check (`git merge-base --is-ancestor 7d9b7334a
<stamp>`), ONE page-build-handler pass over contact via the 081c work-item recipe (NOT
`run_improvement_sweep_once.sh`), then the five criteria; expect `locked_sections_merged =
["chat-input-box"]` in that run's `collected_data->'spec_sections'`, and no new `remove`
items on the next loancalculator rebuilds. Then `git mv` this file to `bugs_closed/` naming
BOTH paths on the commit.

## LIVE — 2026-08-16 15:45Z: round 1 is running; the merge has not yet fired

**Round 1 (`7d9b7334a`) IS in the running chassis (`v1.0.1304`, both replicas, pods started
10:41Z — after that commit's 10:09Z). Round 2 (`57336c127`, 15:10Z) is NOT.** Proven at the
artefact, not at the tag and not at git, with controls in the same exec
(`agent-chassis-5d95ddddfd-48lv6`, `grep -a … /proc/1/exe`):

| probe | result | why it is the right probe |
|---|---|---|
| `locked_sections_merged` | **PRESENT** | round-1 result key, introduced by this fix |
| `LOCKED_MERGE_SKIPPED` | absent | round-2 durable trace — correctly not in a binary built before it |
| `load_page_sections_from_spec` | PRESENT | positive control: the grep can see strings in this binary |
| `deadbeefcontrolstring` | absent | negative control: the grep is not matching everything |

The startup `build provenance` line had already scrolled (absent from `--tail=3000`), which is
the documented shelf-life problem — so the probe above is the evidence, and it is corroborated
independently **at the data**: every one of the 20 `page-build-handler` runs since the roll
carries `spec_sections.locked_sections_merged` and `locked_merge_count`, keys no earlier binary
could emit.

**What is NOT yet proven, stated plainly: the merge has never fired in production.** All 20
post-roll runs were on pages with no locked rows, so `locked_merge_count = 0` and
`locked_sections_merged = []` every time. That is the no-locked-rows regression case behaving
correctly — it is not evidence the merge works.

**And the symptom counter is currently uninformative.** `remove`-blocked items filed since the
roll: **0** — but locked pages rebuilt since the roll: **also 0**. There has been no demand, so
the zero says nothing (LANDMINES/memory: *a post-fix ZERO needs a DEMAND control*). The 7
existing `remove` items are all pre-roll (6 on 2026-08-15 17:11–17:58Z, 1 on 08-13).

**Consequence for closing this file:** step 1 of "How to close" (is it live?) is DONE and
recorded here; steps 2–3 (the owner's five acceptance criteria on webdesign.uk `contact`, via
the 081c work-item recipe) are still owed, and by the owner's ownership note at the head of this
file they belong to the filing lane, which holds the chat-box lock. **A rebuild pass on `contact`
rewrites its unlocked sections and redeploys a live shopfront page, so it is not something a
second lane should fire unannounced** — the recipe is ready in the lane RUNBOOK (C5) with the
exact assertions. The alternative demand is a natural rebuild of any of the 12 loancalculator
calculator pages, which would exercise the merge with no new action from us; the first one to run
should show `locked_merge_count = 1` and file no `remove` item.

Round 2's durable `LOCKED_MERGE_SKIPPED` trace stays inert until the next roll: today, a locked
query that fails mid-build still degrades to lock-blind with only a log line.

## PROVEN — 2026-08-17: the merge fired on its own, and every acceptance criterion holds

**Both halves are now LIVE** (`v1.0.1305`, pods started 2026-08-16 22:07Z; probed at
`/proc/1/exe` on `agent-chassis-5657f446c7-q7b82`: `locked_sections_merged` PRESENT,
`LOCKED_MERGE_SKIPPED` PRESENT — round 2's durable trace shipped this time — with
`load_page_sections_from_spec` PRESENT as the positive control and `deadbeefcontrolstring`
absent as the negative).

**And the mechanism has been exercised for real, without anyone firing anything.** At
2026-08-16 16:09:17Z — under round 1 on `v1.0.1304`, ~24 minutes after the previous entry was
written — `page-build-handler` rebuilt **loancalculator.co.uk `index`**, a page carrying the
locked `tool-3` calculator (`tool-loan-repayment`, `lock_type=permanent`). Correlation
`b45d9965-1efb-4d58-aa05-447ce4bc83a8`, COMPLETED. This is the fleet half of the acceptance,
arrived as a natural rebuild rather than a driven one.

The owner's five criteria, each measured against that run:

| # | criterion | measured |
|---|---|---|
| 1 | the PROPOSED list contains the locked section | ✅ `spec_sections.sections` = `["hero","info-card-grid","tool-list","guide-list","call-to-action","tool-3"]`, `source=site_plan_tables`, `locked_sections_merged=["tool-3"]`, `locked_merge_count=1`. The plan tier served 5 names; the 6th is the merge. |
| 2 | the cache tells the truth afterwards | ✅ `pages.sections` for `index` now carries the same 6, written 16:23:57Z |
| 3 | the locked row is untouched (artefact, not status) | ✅ row still `created_at 2026-08-02`, **`updated_at 2026-08-09 14:44:40`** — both a week BEFORE this run; `md5(rendered_html)` `99a7f143…`; still `lock_type=permanent`, position 6 |
| 4 | an UNLOCKED sibling is still rebuilt (058's own control) | ✅ all five — hero, info-card-grid, tool-list, guide-list, call-to-action — carry `created_at = updated_at = 16:23:23Z`, i.e. deleted and re-inserted by this very save |
| 5 | no `lock_blocked_change … blocked_action: remove` for the slot | ✅ **0 `remove` items fleet-wide since the fix rolled**; the newest in the whole table is still `2026-08-15 17:58:27Z`, pre-fix. The page's only item is the `overwrite` one from 2026-08-08, `updated_at` unchanged (deduped by `item_key`, 058's design) |

**This time the zero has its demand control.** The previous entry warned that "0 `remove` items"
meant nothing while 0 locked pages had rebuilt. A locked page HAS now rebuilt — the whole
mechanism ran end to end on one — and it produced no `remove` item. The counter is answering a
question that was actually asked.

**Sixth check, beyond the owner's five: the SERVED page.** `https://loancalculator.co.uk/` →
200, 56,257 bytes, and the locked calculator is in it: **16 occurrences of its own namespaced
class `tool-loan-repayment-section`** (positive control `tool-list` = 25, negative control
`zzz-absent-control-zzz` = 0). So the rebuild kept the section in the list, kept the row, and
shipped it to visitors.

⚠ **A near-miss worth copying, because it nearly became a false alarm here:** the first artefact
check grepped the served page for `tool-3` and for `data-component="tool-3"` and found **zero**,
which reads exactly like "the locked calculator is not being served". Both markers were
INVENTED — the slot name is a database label that never appears in the markup, and this tool's
template emits no `data-component` attribute at all. The correct check takes its literal FROM
the artefact (`substring(rendered_html …)` → the namespaced class the template actually writes)
and runs a present-control and an absent-control in the same command. Recorded in `WRONG_CALLS`.

**Status: the defect is no longer reproducible, and the fix is fixed-AND-live by the stated bar.**
What remains is not a defect but a hand-over: the owner's acceptance text names webdesign.uk
`contact`, whose chat box is the filing lane's and which has not rebuilt since the roll. That
page's evidence will be identical in shape to the table above (its `chat-input-box` is the same
case: a tool-level locked row on a tier-1 page whose plan omits it). **Recommendation: this file
moves to `bugs_closed/` on the owner's word; the chat-box lock decision stays with the filing
lane, whose ruling said the lock holds "until this is fixed and live" — it now is.**
