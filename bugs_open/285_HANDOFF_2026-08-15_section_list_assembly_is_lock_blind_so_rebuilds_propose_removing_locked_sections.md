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

- **plan_sections' resolver excludes tool-level components from name
  normalisation** (`loadComponentNameResolver`, section/element only —
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
