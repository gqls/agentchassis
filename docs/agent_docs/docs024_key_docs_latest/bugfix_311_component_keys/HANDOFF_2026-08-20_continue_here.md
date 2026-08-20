# HANDOFF — `bugfix_311_component_keys`, 2026-08-20 (cold start: read this file, then §"What to do next")

> **Why this file exists at all.** The session of 2026-08-20 was pointed at
> `HANDOFF_2026-08-19_continue_here.md` in this directory and **it did not exist** — never written,
> confirmed by `git log` on the directory. Reconstructing state from NOTES + RUNBOOK cost the first
> twenty minutes. **If you do work in this lane, update this file before you stop.**

## The bug in one paragraph

`bugs_open/311`: the component **selector** looks a section up by `section_type`, the component
**writer** looks it up by `function`. A row carrying one key but not the other is therefore
invisible to the first and immovable to the second — so a site could neither REUSE an existing
calculator nor CREATE its own, and the page shipped with no calculator at all. The owner saw it as
*"remortgagecalculator.uk left out the actual tools."* The fix diverts on collision: when the row a
store would overwrite belongs to another site, write a **new site-scoped base row** instead
(`…-loanzy-uk`) and log `COMPONENT_COLLISION_DIVERTED`. There is a **tool-level twin** of the same
wall (`RFC_036 §9.3`, `create_tool_component`) and the owner ruled the pair is one logical change.

## Status — what is DONE, and the evidence, not the assertion

**Both halves are FIXED, COUNCIL-APPROVED, LIVE and DEMAND-PROVEN.**

| half | commit | council | proven by |
|---|---|---|---|
| section writer (`store_generated_component`) | `17d883333` | `fc3ac5f4` APPROVED r1 | six real collisions on loanzy.uk, 08-19 and 08-20 |
| tool writer (`create_tool_component`, RFC_036 §9.3) | `e24bc9c0f` | `ceae30f2` APPROVED r1 | `tool-ab-test-calculator` on webdesign.co.uk, 08-19, incl. the served page |

- **Six diversions, zero collateral damage.** Every one completed on **attempt 0**; every incumbent
  it could have overwritten is **byte-identical** to an md5 pinned minutes beforehand. Six
  `COMPONENT_COLLISION_DIVERTED` rows in `agent_error_log`.
- **Graded at the served artefact, not at a status.** loanzy `tool-car-finance-calculator` 0 → 4
  `<input>`; `tool-loan-comparison-calculator` 0 → **6**; `tool-overpayment-calculator` 0 → **5**;
  `tool-settlement-calculator` 0 → **5**. webdesign `tool-ab-test-calculator` serves **5**, and the
  markup was **discriminated** to the forked row (its ids exist in `8a315006` and in neither
  removed slot — `id="verdict"` alone would NOT have distinguished it from the ported original).
- **Live on the current image**, re-probed each roll rather than carried forward — see the RUNBOOK's
  post-roll recipe. As of v1.0.1317 and v1.0.1319: both capability literals present on both
  replicas, invented literal absent, and the stamp probe discriminates (other candidate shas from
  the same build window come back absent).

## What keeps `311` OPEN — ONE residual, and its framing was corrected on 08-20

**Candidate 3, the deploy gate — and read this before you act on it, because the obvious version of
the claim is WRONG.** `311`'s file used to say "nothing reads `needs_section_data` as a blocker".
A session asserted that again on 08-20 and **refuted it 40 minutes later**: `UpdatePageStatusAction`
(`v3_site_actions.go:819-960`) refuses the `deployed` stamp **twice** — on `pageHasComponents`
false, and on `pageSectionShortfall` reporting `rendered < planned` (`bugs_open/040`'s partial-build
rule, widened by `bugs_open/210`).

**What actually survives, as a hypothesis and not a finding:** both gates compare **counts**
(`pageSectionShortfall`, `v3_site_actions.go:1214-1233`, is `count(sections)-count(suppressed)` vs
`count(*) FROM page_components`, every row regardless of which component it points at). So the gate
sees an **empty** slot and cannot see a slot filled by the **wrong thing** — and
`loadSectionComponents` (`v3_site_actions.go:4936-4954`) produces exactly the second shape, giving
every unresolved section name a **stub** behind one `logger.Warn`.

**Do not file it on a work-item census.** The 08-20 attempt ("12 deployed pages carry an open
`needs_section_data`") does not survive contact: the cleanest specimen
(`finetuning.uk`/`password-entropy`) holds the RIGHT component at `build_status='pending'` — a stale
item, not a hole — and two others lost their slots *after* the stamp, which no stamp-time gate can
catch. **An open work item is not a live defect; the artefact is.**

**What a filer needs, in order:** (1) does a stub actually become a `page_components` row on a real
build — read `save_page_sections`' writer, or watch one build; (2) a per-page **artefact** check
(does the served page contain the planned section's markup) instead of a work-item join; (3) then
`090`, because the claim is structural and CLAUDE.md's default applies.

## What to do next (ranked)

1. **Nothing is required for the fix itself.** Both halves are done. If you are here to close 311,
   the only question is candidate 3 above — either file it properly (steps 1-3) or ask the owner to
   close 311 and drop candidate 3 into its own unfiled note.
2. **`bugs_open/337`** (filed from this lane, 08-20): `loans-credit-health-check` blows
   `generate_template`'s 16,000-token ceiling on **every** site that plans it — 46,637 / 47,436 /
   48,553 chars, three attempts each, nine cap-hits, zero successes. Two live pages
   (`loanzy.uk`/`tool-credit-health-check`, `loancalculator.co.uk`/`tool-credit-roadmap`). Ranked
   candidates are in the file; candidate 3 there is architecture-scope, do not take it as diagnosed.
3. **Two loanzy pages remain unrepaired and NEITHER is this bug** — leave both unless the owner says
   otherwise, and neither is this lane's call:
   - `tool-loan-repayment-calculator` — `pages.status='archived'`. Component and page built fine;
     the archived-page guard correctly refused the deploy stamp. **Unarchiving is the loanzy lane's
     decision.**
   - `tool-interest-rate-stress-test` — refused **deterministically** by the `bugs_open/253`
     component floor over an **unrelated** `hero-tool` slot (12→5 class attributes, identical
     figures on two independent runs). The guard's own guidance names the remedy: give the writer
     the component vocabulary in `content_direction`, NOT `section_component_floor=0` ("the
     deliberate escape hatch, not a fix"). That is a site-wide content change — **the loanzy lane's
     call.**
   - Free and unclaimed: `tool-loan-vs-savings` has a good component and serves 0 `<input>` only
     because `needs_rebuild` has no consumer. **One `needs_page` item, no generation.**

## Traps this lane paid for — read the RUNBOOK section, do not re-derive

`RUNBOOK_311_fix.md` carries all of these with the exact queries: the batch re-drive recipe; **read
`idx_swi_dedup`, don't recall it**; **re-pin incumbent md5s immediately before a run** and attribute
drift via `component_versions.change_source` (a shared incumbent was rewritten mid-session by
`scope_component_instance_judged`); **a served-page baseline needs TWO reads** (one 404 was a
transient that would have credited the fix with publishing a page it never touched); **check
`pages.status`, not just `build_status`** (cost one generation + one page build on an archived page,
and the predicate was already in `LANDMINES.md`); and **a refused page build is terminally parked at
attempt 1 of 3** — file a fresh item, don't wait for a retry that cannot come.

`NOTES_311_fix.md` is the running record, newest at the bottom, and it carries the corrections —
including three the 08-20 session made against its own earlier claims in the same session.

## Cold-start commands

```bash
# is the fix still in the running binary? (per SERVICE, controls both ways — RUNBOOK has the full recipe)
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name | head -1 | cut -d/ -f2)
kubectl -n ai-persona-system exec $POD -- grep -aq "COMPONENT_COLLISION_DIVERTED" /proc/1/exe && echo section-half-present
kubectl -n ai-persona-system exec $POD -- grep -aq "library tool claims this function" /proc/1/exe && echo tool-half-present
```
```sql
-- every diversion that has ever happened, and what it saved
SELECT occurred_at, context->>'requested_function', context->>'final_function'
FROM agent_error_log WHERE error_code='COMPONENT_COLLISION_DIVERTED' ORDER BY occurred_at;
-- loanzy's tool pages, the one query that tells you what each one needs
SELECT p.name, p.status, p.build_status, jsonb_array_length(coalesce(p.sections,'[]'::jsonb)) AS planned,
       (SELECT count(*) FROM page_components pc WHERE pc.page_id=p.id) AS slots
FROM pages p JOIN sites s ON s.id=p.site_id WHERE s.domain='loanzy.uk' AND p.name LIKE 'tool-%'
ORDER BY p.status, p.name;
```

**Docs in this lane:** `PLAN_2026-08-19_divert_on_foreign_collision.md` ·
`RUNBOOK_311_fix.md` · `NOTES_311_fix.md` · `README_where_we_are.md` (owner's, append-only) ·
this handoff. **Bug files:** `bugs_open/311_HANDOFF_2026-08-18_selection_keys_on_section_type_…md`,
`bugs_open/337_HANDOFF_2026-08-20_one_section_type_reliably_exceeds_…md`.
