# 278 — webdesign.co.uk's homepage renders `info-card-grid` TWICE, same component, different content, neither locked

**Filed 2026-08-15**, from owner feedback on webdesign.co.uk after the `bugs_open/122` ink canary
rebuilt that site's stylesheet. He reported the copy reading as AI-written *and* "the component is
duplicated on the home page". Both true; the copy half is the `copy_quality_two_stage` lane's and is
taken (their NOTES `b2cccc5c9`). **This file is the composition half only.**

> ## ⚠ UPDATED 2026-08-15, SAME SESSION — LOCATED: **the duplication is in the PLAN, not in the page-build or save path.** §4 below said no cause was asserted; that is now superseded for the *location*, though not for the *reason*.
>
> `[MEASURED 2026-08-15]` `site_plan_sections` for this site's `index` page carries
> `info-card-grid` **twice**, at `ordering` 1 and 2, both rows stamped
> **`2026-07-25 16:26:18.203946+00` — identical to the microsecond**, i.e. written in one
> transaction alongside `hero` (0) and `call-to-action` (3):
>
> ```sql
> SELECT sps.page_name, sps.ordering, sps.component_name, sps.created_at
> FROM site_plan_sections sps JOIN site_plans sp ON sp.id = sps.plan_id
> JOIN sites s ON s.id = sp.site_id
> WHERE s.domain='webdesign.co.uk' AND sps.page_name='index' ORDER BY sps.ordering;
> -- index | 0 | hero | 2026-07-25 16:26:18.203946+00
> -- index | 1 | info-card-grid | 2026-07-25 16:26:18.203946+00
> -- index | 2 | info-card-grid | 2026-07-25 16:26:18.203946+00
> -- index | 3 | call-to-action | 2026-07-25 16:26:18.203946+00
> ```
>
> **So the page is composed exactly as planned.** `save_page_sections_action.go` and the
> re-render path are **exonerated** — §5's first two candidates are closed, and 189's
> neighbourhood is the wrong neighbourhood.
>
> **And it is a TRUE duplicate, not two sections sharing a heading.** Both carry four cards with
> the *same four titles in the same order* — "Sixty-three tools, zero installs" / "Guides on how
> things actually work" / "Built by one person, in the open" / "Notes from the workbench". The
> `content_data` md5s differ only because the body copy under each card was generated
> independently; the structure and headings are identical. A reader sees the same section twice.
>
> **WHAT IS STILL UNKNOWN, and stays unasserted: WHY the plan contains it.** I have measured the
> plan's *contents*; I have not opened the planner. Whether this is a planner defect, a bad input
> spec, or a deliberate two-grid design that the writer then filled identically is **undiagnosed**.
> §4's rule still applies to that question: **a structural claim about the planner needs `090`
> before it is asserted**, because it would predict recurrence on other sites and the census (§3)
> currently says N=1.
>
> **The fix is therefore probably a plan edit plus a re-render, not a code change** — but do not
> apply one until the "why" is settled, or the same plan may simply regenerate it.

**NO ROOT CAUSE IS ASSERTED HERE.** The mechanism is undiagnosed. That is deliberate — see §4.
*(Superseded in part by the update above: the LOCATION is now measured; the REASON is not.)*

## 1. The symptom, measured

`[MEASURED 2026-08-15]` `page_components` for `webdesign.co.uk` `/index.html`:

| position | slot_name | component_id | locked | content md5 | html md5 | len |
|---|---|---|---|---|---|---|
| 1 | `hero` | `23f95f00` | f | `0768847e` | `383dd6d3` | 3177 |
| **2** | **`info-card-grid`** | **`fc56f085`** | **f** | `430b352e` | `f665c013` | 7353 |
| **3** | **`info-card-grid`** | **`fc56f085`** | **f** | `c4a89a0b` | `695f34b9` | 7409 |
| 4 | `call-to-action` | `0197e8d7` | f | `731dd700` | `901a4a8c` | 2464 |

Both rows carry the **same `slot_name`**, the **same `component_id`**, **neither is locked**, and
their `content_data` **differs**. Both render the identical `<h2>` — *"A workbench, not a sales
pitch"* — so the page shows the same heading twice with different cards beneath it. Visible on the
served page today.

## 2. It is NOT `bugs_open/189`, and the discriminators are explicit

189 ("resolving a previously-unresolvable LOCKED section duplicates it on the page") is the nearest
prior art and was checked first. Its signature requires **a locked row** and **identical
`content_data`** (189's own evidence: same `component_id`, same md5, one row `locked=t`). Here:
`locked_rows = 0` and `distinct_content = 2`. **Both discriminators fail, in the same direction, on
this row pair.** Do not fold this into 189.

## 3. Scope: N = 1 fleet-wide — and my first census said 14, which was wrong

`[MEASURED 2026-08-15]` **webdesign.co.uk `/index.html` is the only page in the fleet with a NAMED
component slot occupied more than once.**

```sql
WITH d AS (
  SELECT p.site_id, pc.page_id, pc.slot_name, pc.component_id, count(*) n,
         count(DISTINCT md5(pc.content_data::text)) distinct_content
  FROM page_components pc JOIN pages p ON p.id = pc.page_id
  WHERE pc.slot_name NOT IN ('generic-text-block','section','ported-prose')
  GROUP BY 1,2,3,4 HAVING count(*) > 1)
SELECT s.domain, p.url, d.slot_name, d.n, d.distinct_content
FROM d JOIN sites s ON s.id=d.site_id JOIN pages p ON p.id=d.page_id;
-- webdesign.co.uk | /index.html | info-card-grid | 2 | 2      <- one row, and only one
```

> **⚠ CORRECTED BEFORE FILING, and the error is worth more than the number.** My first census
> counted *any* repeated `slot_name` and returned **14 duplicated slots across 9 sites**. I was one
> keystroke from filing that as a fleet-wide class. **12 of the 14 were `generic-text-block`**, a
> slot a page may legitimately repeat — the control confirms it: **12 pages carry more than one
> `generic-text-block` and that is normal composition, not a defect.** I had encoded "duplicate
> `slot_name`" and read the result as "duplicate component". The question the query answered was not
> the question I was asking. Sibling of MEMORY `narrow-filter-defines-the-conclusion` — except here
> the filter was too WIDE, which is the rarer direction and produces an over-alarming class rather
> than a missed one.

## 4. Why no root cause, and why no `090` run

The plausible read — a planner emitting the section twice — **is inference, not measurement.** I
have not opened the planner, and a second `site_plan_sections` row would be the evidence. Stating it
would be exactly the shape this repo's 2026-07-19 correction warns about: a confident structural
claim built from a plausible mechanism whose code was never read.

`090` was **not** run, and the reason is that the 2026-07-31 ruling's trigger is a **cross-cutting or
structural** root-cause claim. This is N=1 and asserts no cause, so the trigger does not fire. **If
the next thread forms a structural theory — a planner defect, a save-path defect, anything that
would touch other sites — file `090` before asserting it.** That is the point at which it becomes
required.

## 5. Where to look first (candidates, unranked, none verified)

- `site_plan_sections` for this site/page: does the PLAN carry the section twice? One query settles
  whether this is a planning or a saving defect, and it is the cheapest discriminator available.
  (⚠ the table's join column is **not** `site_id` — check `\d site_plan_sections` first; my own
  query failed on that assumption.)
- `save_page_sections_action.go` — 189 §"The mechanism" documents two pre-existing defects in
  section identity resolution in this same file. Different signature, same neighbourhood.
- Whether the two rows arrived in one build or two: `created_at`/`updated_at` on both rows.

## 6. ⚠ ORDERING CONSTRAINT — composition BEFORE voice

Recorded at the `copy_quality_two_stage` lane's request, because the two fixes interact:

**The voice rewrite must NOT run until this bug's cause is known.** A page-level `content_rewrite`
could coincidentally collapse or re-duplicate the section, **destroying the diagnostic state**; and a
section-scoped rewrite of position 2 alone leaves position 3 serving the old copy. **Composition
first, then voice.** When this closes, ping that lane — or their
`copy_quality_two_stage/HANDOFF_2026-08-15` picks it up.

## 7. Do not confuse with the ink work

webdesign.co.uk's stylesheet was rebuilt ~10:46Z 2026-08-15 as the `bugs_open/122` 5.0 canary. Every
in-prose link moved `#2b2b2b` → `#915e2c` (5.15:1). **That is intended and owner-approved.** It
touched `styles.css` only and cannot have created or affected these rows — `page_components` were
not rewritten by it.

## 8. PRESERVED EVIDENCE — the duplicate is a natural experiment, and the fix will destroy it

**Banked 2026-08-15, before any repair.** The two sections are the same planned component, filled by
the writer twice from the same context in the same build. That is a controlled comparison nobody
would have thought to run, and **a plan edit + re-render deletes it permanently.** Recorded here so
the fix does not have to be delayed to preserve it.

Card titles are identical across both (§ update above). The **bodies** are not, and the split is the
finding:

| card | pos 2 vs pos 3 | body (position 2) | body (position 3) |
|---|---|---|---|
| 1 | **IDENTICAL** | "Colour contrast checks, layout generators, SVG tools and more, all running client-side" | *(same)* |
| 2 | **DIVERGENT** | "**Practical explanations** of the CSS, accessibility rules and security mechanics **that sit…**" | "**Explanations** of the CSS, accessibility rules and security mechanics **behind the tools,…**" |
| 3 | **DIVERGENT** | "**No team, no investors, no growth targets. Just someone who builds websites**, writing do…" | "**One person builds websites and writes down what they learn here**, with no team, no inve…" |
| 4 | **IDENTICAL** | "Short updates on new tools, changed guides and anything else worth flagging as it's ad…" | *(same)* |

**`[MEASURED]` 2 of 4 identical, 2 of 4 materially reworded — same facts, different sentences.**

**Why this is worth more than "the md5s differ":** it is not blanket non-determinism. Cards 1 and 4
are short and enumerative and converged **exactly**; cards 2 and 3 are longer and more discursive and
diverged in structure while preserving content. `[INFERRED — N=1, not tested]` the plausible read is
that constrained, list-shaped content converges and open-ended prose does not, but **one page cannot
establish that** and it must not be repeated as a finding.

**Consequence for anyone building acceptance gates on generated copy:** a gate that compares prose to
prose would fail this pair on cards 2 and 3 while nothing is wrong — the same generator, the same
inputs, twice. Compare declared sets, types and structure instead. Contributed to the
`copy_quality_two_stage` lane, whose stage-2 design uses exactly that rule.

**If the "why" diagnosis re-fills this section as part of a repro, capture the third generation** —
the second data point is free at that moment and expensive at any other.
