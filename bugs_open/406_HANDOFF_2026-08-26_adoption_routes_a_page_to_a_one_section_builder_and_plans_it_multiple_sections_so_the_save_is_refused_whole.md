# 406 — adoption routes a page to a ONE-SECTION builder and, in the same transaction, plans it MULTIPLE sections, so every save is refused whole

**Filed 2026-08-26 by the `bugs_open/357` lane.** Found by running two real site adoptions to
give 357's phase 2 its first production firing; this is what was actually blocking it, and it
is a bigger and more general defect than 357 itself.

**This is NOT `bugs_open/357`.** 357 is the *mislabelling* — a row claiming to be the shared
`hero` while holding a tool. This is the *refusal* — a save thrown away whole, leaving the page
with **zero** rows. They share a cause, which is why filing them together would hide both.

---

## 1. The defect in one paragraph

`apply_adoption_plan` decides a crawled page is interactive and routes it to
`tool-recreation-handler`, whose save can only ever emit **one** section — and **the same
action, in the same transaction, writes that page a multi-entry `pages.sections` plan.**
`save_page_sections`' prune floor then divides the one projected insert by the planned count
and refuses the entire save. Nothing is written, the page keeps zero `page_components` rows,
and a `save_refused_incomplete` item is filed for a human who never comes.

**The route chooser and the guard's denominator are written by one action, moments apart, and
they disagree about what the page is.**

## 2. Evidence [ALL MEASURED 2026-08-26 unless dated otherwise]

**The contradiction, in the adoption's own words.** The `needs_tool_recreation` item for
`cv1.co.uk/tool-example` (page `f763ca0e-a5ad-4d25-9e6c-37d158c13493`) carries:

```json
{"name":"Interactive Job Prep Checklist","type":"tool","self_contained": true, …}
```

`source=adoption`, `created_by=site-adoption-agent`, `item_key=needs_page:tool-example`,
`spec.mode=recreate`. That page's `pages.sections`, written by the same action:
`["generic-text-block","features","call-to-action"]` — **three**. `self_contained: true` and a
three-section plan cannot both be true.

**The refusals, verbatim from the work items** [2026-08-25 12:40Z / 12:42Z]:

```
index        planned sections 25% (1 of 4)   prune_floor_ratio=0.50   -> whole save REFUSED
tool-example planned sections 33% (1 of 3)   prune_floor_ratio=0.50   -> whole save REFUSED
```

Both pages then held **0** `page_components` rows. The recreated fragments themselves were
perfect — 26,271 B and 21,265 B, no `<section>`, no `data-component`, not full documents, both
carrying the `tool-page` wrapper — so nothing about the *content* was wrong.

**The arithmetic is closed-form.** A one-section save scores `1/planned`. The floor trips on
`ratio < floor` (`prune_floor.go:128`), so:

| planned | ratio | outcome |
|---|---|---|
| 1 | 1.00 | passes |
| 2 | 0.50 | passes — exactly at the floor |
| **3** | **0.33** | **refused** |
| **4** | **0.25** | **refused** |

**Any adopted interactive page planned with ≥3 sections is unsaveable. Always. By arithmetic.**

**The cross-check that makes this the explanation for 357's population, not merely adjacent to
it.** If this is the mechanism, 357's mislabelled rows must live almost entirely on pages
planned with ≤2 sections — the only ones whose one-section save could ever have completed:

| planned sections on the page | rows in the 357 population | one-section save clears floor? |
|---|---|---|
| 1 | 1 | yes |
| 2 | **20** | yes (exactly at 0.50) |
| 4 | 1 | no — predates the floor |

**21 of 22.** The floor has been silently *selecting* which tool pages get a row at all: plan
≤2 and it saves (mislabelled — that is 357); plan ≥3 and it is refused outright and the page
stays empty. **357's population is the survivors of this defect.**

**Blast radius.**

> ⚠ **CORRECTED 2026-08-26, within the hour, BEFORE anyone quoted it — the first version of
> this section said "34 items across 16 domains" as though all 34 were this defect. They are
> not, and I had not classified them when I wrote it.** Classifying every parked item by the
> cohort numbers in its own reason string:
>
> | shape | items | domains |
> |---|---|---|
> | **THE 406 SHAPE — `1 of ≥3`** | **6** | **5** |
> | no cohort captured (older reason format) | 26 | 15 |
> | other shrinkage (`2 of 5`, `7 of 20`) | 2 | 2 |
>
> **Six are demonstrably this defect**, and two of those six are cv1's, which I caused by
> running the adoptions — so **four pre-existing victims across four domains**:
> `finetuning.uk/blog` (1 of 3), `fundamentallyai.com/tool-model-approach-selector` (1 of 3),
> `mortgagecalculator.co.uk/game-fact-finder` (1 of 4),
> `webdesign.co.uk/tool-llm-cost-calculator` (1 of 4).
>
> **The 26 are UNATTRIBUTED, not attributed elsewhere.** Their reason strings predate the
> `planned sections` cohort, so the numbers simply are not in the record; several are named
> tool pages (`loanzy.uk/tool-credit-health-check`, `-eligibility-checker`,
> `-interest-rate-stress-test`, `mortgagecalculator.co.uk/tool-affordability`,
> `webdesign.co.uk/tool-mind-map`, `ai-agent-orchestration.com/tool-automation-savings-estimator`,
> `gamesdesign.co.uk/tools-index`) and are *consistent* with this shape without being evidence
> of it. ⚠ **Do not attribute them by present-day page state** — a page's plan and row count
> today is a PRESENT-tense census used to explain a PAST event, which is a logged wrong-call
> shape on this estate. If they matter, attribute them from `page_component_history` or leave
> them unattributed.

**What IS true of all 34, and is a finding in its own right: nobody reads this queue.** 34
items sit in `needs_human_review` with no `handler_agent`, spanning 2026-07-31 → 2026-08-25,
and the count grew from 32 in the fourteen hours I was watching. A refusal that files a work
item instead of erroring has no other alarm. Named tool pages among them: `webdesign.co.uk/tool-llm-cost-calculator` (1 of 4),
`fundamentallyai.com/tool-model-approach-selector` (1 of 3),
`mortgagecalculator.co.uk/game-fact-finder` (1 of 4), `loanzy.uk/tool-credit-health-check`,
`loanzy.uk/tool-interest-rate-stress-test`, `loanzy.uk/tool-eligibility-checker`,
`ai-agent-orchestration.com/tool-automation-savings-estimator`,
`loanandmortgagecalculator.co.uk/loans-application-tracker`, `gamesdesign.co.uk/tools-index`.

⚠ **Several older rows have EMPTY cohort captures** — the `planned sections` cohort postdates
them, so a blank is a reason-string format difference and **must not** be read as a different
cause.

⚠ **`site_work_items` is a ROLLING WINDOW.** Joining `needs_tool_recreation` to `pages` finds
only the two cv1 rows because the history is archived out. Do **not** conclude "only two pages
were ever routed to tool recreation" ([[a-closer-census-cannot-see-what-it-succeeded-at]]).

## 3. Root cause, at the lines

- **The route:** `platform/orchestration/actions/apply_adoption_plan_action.go:719` —
  `if len(page.Features) > 0` → `itemType = "needs_tool_recreation"`,
  `handlerAgent = "tool-recreation-handler"`.
- **The plan:** the same action writes `pages.sections` (`applyAdoptionPlanPagesUpsertSQL`).
  `save_sections_prune_floor.go`'s own comment names `apply_adoption_plan` as one of the
  writers of that column.
- **The one-section guarantee:** `tool-recreation-handler` declares
  `expects_no_sections_metadata: true`, so `save_page_sections` cannot take the metadata path
  and always falls through to `saveSectionsExtractFromHTML`, whose no-`<section>` arm
  (`save_page_sections_action.go:1561`) stores the whole fragment as **exactly one** section.
- **The refusal:** `measurePageSectionCompleteness`
  (`save_sections_prune_floor.go:~127-148`) reads `jsonb_array_length(p.sections)` as the
  `planned sections` cohort's denominator, and the floor refuses the whole save because delete
  and insert are one operation.

**Nothing here is individually wrong.** The floor is well-designed, its cohorts were measured
rather than assumed, and refusing beats half-pruning. The defect is that one action emits two
statements about the same page that cannot both hold.

## 4. Fix candidates, ordered by what makes the bad state unrepresentable

1. **Plan a tool-routed page with ONE section** — at the point of routing, in
   `apply_adoption_plan`, write `pages.sections` with a single entry for any page that goes to
   `tool-recreation-handler`. This makes the contradiction unrepresentable: the same branch
   that chooses the one-section builder writes the one-section plan. Shared seam → council gate.
   ⚠ Decide what the single planned name should be: keeping `planned[0]` preserves the slot
   name, and **the slot name is what Layer 2 matches on with exact string equality** — renaming
   it arms the carry-forward landmine (357's first council round was rejected for exactly that).
2. **Make the floor aware that a declared one-section producer is not a shrinkage.** The
   handler already declares `expects_no_sections_metadata`; the floor could read that and skip
   the `planned sections` cohort for such a caller. Narrower blast radius than (1), but it
   leaves `pages.sections` still lying about the page, and every other consumer of that column
   keeps reading the wrong number.
3. **Lower `prune_floor_ratio` on `tool-recreation-handler`'s save step.** What the refusal
   message itself prescribes, one config row, live immediately. **Rejected as the fix** — it
   removes the truncation guard for every tool recreation on the estate, so a truncated LLM
   completion could overwrite a good tool page (the `bugs_open/012` shape). Acceptable only as
   a deliberate, time-boxed operational unblock.
4. **Per-page plan correction** — what was done on cv1 on 2026-08-25 under owner decision
   (`docs024_key_docs_latest/bugfix_357_component_identity/OPERATION_2026-08-25_correct_cv1_tool_page_plans.sql`).
   It works and it is reversible; it fixes two pages and nothing else. **Not a class fix**
   ([[a-one-off-deletion-is-not-a-class-fix]]).

**Note (1) and (2) are not exclusive** — (1) stops new contradictions, (2) makes the guard
robust to any future one-section producer. Neither retroactively saves the 34 parked pages;
those need a re-run once the producer is fixed.

## 5. How to verify a fix

```sql
-- must be EMPTY after the fix: pages routed to tool recreation whose plan they cannot satisfy
SELECT s.domain, p.name, jsonb_array_length(p.sections) AS planned
FROM site_work_items wi JOIN pages p ON p.id = wi.page_id JOIN sites s ON s.id = p.site_id
WHERE wi.item_type = 'needs_tool_recreation'
  AND jsonb_typeof(p.sections) = 'array' AND jsonb_array_length(p.sections) > 2;

-- and the queue must stop growing (dated; compare against 34 / 16 domains at 2026-08-26)
SELECT count(*), count(DISTINCT s.site_id) FROM site_work_items wi JOIN sites s ON s.id = wi.site_id
WHERE wi.item_type = 'save_refused_incomplete' AND wi.status = 'needs_human_review';
```

**End to end:** adopt a site with an interactive page, and assert the page ends with **≥1**
`page_components` row rather than a `save_refused_incomplete` item. ⚠ Verify at the row and at
the served page, not at the item status — `complete` is not a repaired artefact
(`bugs_closed/287`).

## 6. Diagnosis-loop status — read this before citing the file as confirmed

Filed through the loop per the owner ruling of 2026-07-31: intake
`f2fa4b9e-28b6-4f45-9ffa-2627c2031af0`, **RUN_CORRELATION_ID
`fbdaca97-a97e-41e6-b422-2475521e6a6c`**.

**It returned `UNVERIFIABLE` (`stopped: scope-not-narrowing`) — NOT `CONFIRMED`, and NOT
`REFUTED`.** It agreed the numeric signature was present and declined to conclude, naming two
gaps. **Both were then closed first-hand, and that substitution is declared here rather than
left silent:**

- *"nothing ties this page's tool-recreation route and this sections write to a single
  `apply_adoption_plan` transaction"* → closed by the work item's own columns
  (`source=adoption`, `created_by=site-adoption-agent`, `item_key=needs_page:tool-example`,
  `spec.mode=recreate`, populated `interactive_features`) — the literal `pageSpec` shape built
  at `apply_adoption_plan_action.go:710-724`, plus `applyAdoptionPlanPagesUpsertSQL` being the
  same action's page writer.
- *"the actual body of `measurePageSectionCompleteness` showing the denominator is
  `pages.sections`"* → read directly: `jsonb_array_length(p.sections)`,
  `m.Planned = planned - suppressed - m.LockedRows`,
  `{Label: "planned sections", Confirmed: projected, Stored: m.Planned}`.

⚠ **Being in the loop's symbol scope is not being retrieved** — `measurePageSectionCompleteness`
was listed in its own scope and it still could not read the body.

## 7. Relations

- **`bugs_open/357`** — the mislabelling. This defect explains **21 of its 22** rows. 357's
  phase 2 is proven and unaffected; it was this refusal that kept it from firing for two days.
- `bugs_closed/165` — the prune floor's own origin.
- `bugs_closed/156` — why the refusal is all-or-nothing (a partial prune duplicates rows).
- `bugs_open/012` — the truncation shape that makes candidate (3) unsafe.
- Full evidence trail, including the two adoptions that found it:
  `docs/agent_docs/docs024_key_docs_latest/bugfix_357_component_identity/NOTES_component_identity.md`
  (2026-08-25 afternoon onward) and `HANDOFF_2026-08-26_continue_here.md` §4.
