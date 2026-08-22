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

~~**What a filer needs, in order:** (1) does a stub actually become a `page_components` row … (3) then `090`.~~
**DONE, 2026-08-20 14:45Z — all three steps taken, and the answer changed the claim again.**

- **(1) The stub hypothesis is REFUTED by measurement.** Only **11 of 1,855** `page_components` rows
  have `component_id IS NULL`, and they are not stubs standing in for holes: two are `lendzy.co.uk`
  tool pages carrying 13,262 / 14,747 chars whose served pages return **2 and 3 `<input>`**. The
  count is not being inflated.
- **(2) The artefact check, on the ORIGINATING case.** `remortgagecalculator.uk`/`index` plans six
  sections, holds five, and the missing one is **`mortgages-repayment`** — this bug's own step 1.
  `build_status='needs_rebuild'` (so **the gate fired and was right**), and the page nonetheless
  serves 200 / **40,726 bytes / 0 `<input>`**.
- **The mechanism, therefore:** *the refusal is a status write — it neither retracts the published
  artefact nor reaches a worker, and `needs_rebuild` has no consumer, so the page serves the hole
  indefinitely while the DB reads correct.* A **convergence** defect, not an absent gate. Nothing
  else in `bugs_open/` carries it (`210` is the inverse: a stamp wrongly APPLIED).
- **(3) `090` verdict is IN: `UNVERIFIABLE` (stopped: scope-not-narrowing), corr
  `e9555fad-5b25-46bc-9908-f40db98e16a4`** — and it killed two of the bullets above as they were
  first written. **Read `NOTES_311_fix.md`'s 15:30Z entry before touching this.** In short:
  "the gate fired and was right" was an **attribution with no evidence** (four writers leave an
  identical `needs_rebuild` row and there is no attribution column; the loop found **zero**
  `agent_error_log` rows for three shortfall pages), and "`needs_rebuild` has no consumer" is
  **refuted** — `webdesign.co.uk`/`tool-ab-test-calculator` was rerendered, republished and serves
  the new calculator while still flagged `needs_rebuild` with a `deployed_at` six days stale.
  **The missing piece is ATTRIBUTION, not observation:** make the guards distinguishable in
  `agent_error_log`, or catch one in the act. And the "does a refusal retract the published file"
  half **cannot be answered from the Go code at all** — it needs `sites.deploy_config` /
  `published_hash` / `published_at`. The status-column half was contributed to `bugs_open/315`
  (same class, opposite direction) rather than filed as a new number.

**Repairing the originating page is BLOCKED, not forgotten:** that site is locked,
`locked_by = "portfolio_positioning: owner HALT 2026-08-18 pending classifier register-input (RFC) +
builder-flow decision"`. The incumbent (`b89f91e1`, html `a2c00f1c66ce6f4ef72b48083f1e3da6`) is
re-pinned and the `needs_new_component:mortgages-repayment` key is held only by `cancelled`/`failed`
rows, so the two-item recipe is insertable the moment the halt lifts.

## STATE AS OF 2026-08-22 18:00Z — **311 CLOSED · 345 APPROVED, LIVE AND FIRING · nothing owed by this lane**

**Read this block first; everything below is history.**

| thread | state |
|---|---|
| `bugs_closed/311` | **CLOSED.** Originating page serves its calculator; 7 diversions, every incumbent byte-identical |
| `bugs_open/345` | **Fix complete in THREE parts, all live; council APPROVED r5 (`67b07528`, 4 advisories none high).** Go half (v1.0.1322+), migration `533` (prompt), migration `555` (dispatcher mapping — **without which the first two were inert**). Path **firing**: `input_data ? 'last_error'` = **6 as of 2026-08-22 18:00Z**, was 0 all-history. Target case `ceea0c07`: rejected → retry carried the reason → **complete at attempt 1**. **Causation n=1, NOT claimed.** **No residual: the `completed_at` hardening is LIVE on v1.0.1326 too** — verified 18:10Z by probing the capability (`wi.error, wi.completed_at` PRESENT on both replicas) after the sha hunt failed to find the build point; positive and near-miss controls both correct |
| loanzy.uk | **8 of 11 tool pages serve calculators** (was 1 on 08-21). Two are `337`'s (other lane); one has no calculator section by design |
| `RFC_034` | **Both owner questions CLOSED**: new components needing scope go inline; existing extracted-JS ones are left. Delivered to the 283 lane with the sweeper-widening interaction flagged |
| `bugs_open/351` | `bugfix_198_roundtrip_writers` lane's. Both this lane's findings recorded there; backfill-ordering is an owner call **when that fix lands** |
| `bugs_open/337` | another lane's (owner, 08-22) |

**If you are picking this lane up: there is no queued work.** The only future event is 345's
causation population growing past n=1 — check with:
`SELECT count(*) FROM orchestration_states WHERE collected_data->'input_data' ? 'last_error';`
and, for outcomes, whether post-fix items with a rejection then complete rather than repeating it.

**Two process lessons this lane paid for, both now in `WRONG_CALLS.md`:** a fix can be deployed,
correct, and *connected to nothing* — a zero with a comfortable explanation is the most dangerous
measurement there is; and when the code moves, **re-derive the whole council submission from the
code** rather than patching the edit you are thinking about (that defect cost rounds 2 and 4).

---

## CLOSE-OUT ASSESSMENT 2026-08-21 evening — what would close 311, and what genuinely remains

**The titled defect is DONE by every bar this estate uses**: both halves fixed, council-approved,
live on the current image (v1.0.1322, provenance-stamped `bac189921`, re-verified each roll), and
demand-proven — six real diversions, every incumbent byte-identical, and the scoped name derived
correctly on the originating case. CLC-020's register entry now says so.

**The ONE thing between here and `bugs_closed/`:** the originating PAGE
(`remortgagecalculator.uk/index`, 0 `<input>` since 08-17). Its repair was blocked by `345`; both
halves of 345's fix went live today (v1.0.1322 + migration `533` applied and recorded); demand test
`e9e5a10b` is in flight. **If it lands and the page serves a calculator, close 311** — move the
file by `git mv` naming BOTH paths on the commit (the LANDMINES `git mv` + pathspec trap), and the
residuals below stay open in THEIR files, not 311's:

| residual | where it lives | why not 311's |
|---|---|---|
| retry-blindness | `bugs_open/345` (fix live, council round 2 pending — read the verdict, then close on a second-attempt-differs observation) | its own mechanism, its own file |
| the 16k output cap | `bugs_open/337`, unowned | upstream of the store; cannot exercise 311's seam |
| unscoped diverted rows + sweeper blind spot + section birth gate | the `283` lane (`CONTRIB_2026-08-21b…` in their dir; no reply yet as of 18:15Z) | their seam, their programme; owner has the design question |
| needs_rebuild convergence / status-column truth | `bugs_open/315` (contributed) + the `090` UNVERIFIABLE trail | a fleet convergence question, not the component deadlock |
| loanzy leftovers | archived page + `253` floor page = loanzy/owner calls; `tool-loan-vs-savings` = one free `needs_page` | documented in their lane's contrib |

**If the demand test FAILS**, the failure is evidence about `345` (or a third gate), not about 311 —
the diversion on that page is already proven. In that case the honest close is still available:
close 311 on its titled defect and point the page's repair at 345's file. The owner asked; this is
the recommendation either way, with the page's outcome deciding only which sentence goes in the
close-out commit.

## ⚠ UPDATE 2026-08-21 — the originating site is UNBLOCKED, the diversion FIRED there, and a third gate now blocks it (`bugs_open/345`)

- **`remortgagecalculator.uk` is UNLOCKED** (owner instruction, 2026-08-21). Only that row moved;
  `adversecreditmortgage.co.uk` is still held with its 41 items. Old lock values and the reasoning
  are in `portfolio_positioning/CONTRIB_2026-08-21_from_311_lane_your_halt_is_LIFTED_on_remortgagecalculator_only.md`.
- **311's fix is PROVEN on its own originating case.** The store's refusal names
  `function="mortgages-repayment-remortgagecalculator-uk"` — the site-scoped name — so identity
  resolution diverted correctly. **Identity resolution precedes pre-store validation.**
- **Blocked downstream by `bugs_open/345`** (filed from here): the generated template declares
  `site_specs.locale.currency_symbol`, an aspect that exists nowhere (grep-proven), and
  `generate_template`'s inputs carry **no previous-failure field**, so every retry reproduces the
  identical rejection. 99 such rejections fleet-wide, every repeat item with exactly one distinct
  reason, one item burning **52 generations** under a 3-attempt budget. Item `95fe67da` **cancelled
  at attempt 2** rather than pay for a third. **Do not seed a `locale` aspect to get past it** —
  that is fixing the checker to agree with broken output.
- **Three unrelated gates have now each blocked a 311 repair, and none is 311:** `pages.status='archived'`
  (loanzy loan-repayment), the `253` component floor on an unrelated slot (loanzy stress-test), and
  `345` (remortgage index). The collision fix is sound; what is left between a diverted component
  and a serving page is a queue of other people's guards.
- **The originating page is a one-command repair the moment `345` is fixed:** site unlocked, both
  dedup keys free, incumbent pinned same-day (`b89f91e1`, html `a453a6565489c348ad6a9156a8af812f`
  — it moved AGAIN on 08-20 via `scope_component_instance_judged`, so re-pin same-day), served
  before = 200 / 40,726 B / 0 `<input>` / md5 `89910f6e7875f1d310d962f83e443989`.

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
