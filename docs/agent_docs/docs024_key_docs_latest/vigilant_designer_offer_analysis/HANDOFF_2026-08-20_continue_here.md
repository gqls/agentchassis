# HANDOFF — vigilant designer + offer analyser (2026-08-20)

**COLD-START = this file + `bugs_open/335` (filed today, the live thread and a defect in THIS
lane's own agent) + `features_open/030` §10 (the v2 backlog) + `features_open/034`.**
`bugs_closed/295` and `bugs_closed/301` are both closed — read them only for their residuals, which
are now `bugs_open/333` and `bugs_open/335`.
**This supersedes `HANDOFF_2026-08-18_continue_here.md`.**

> **Re-run every liveness claim here before acting.** In the ~16 hours before this file was written
> **298 commits** landed, `bugs_open/301` was fixed AND closed by another lane, a new bug (`333`)
> was filed on its residual, and this lane's own agent produced a false claim on a live site. None
> of that was visible from the previous handoff.

## The one-line state

> **B4 is enrolled, sweep-driven, and proven — and today it is also the thing that needs fixing.**
> `offer_ordering` is on **5 of 23** sites; the sweep is **CLOSED** (owner's cost control).
> **`bugs_open/335`** is the live thread: the analyser lifted a stale factual claim off a page's
> meta description into `lead_with` **rank 1** and stamped it `from_field: trust_threshold`. It was
> caught by **another lane**, not by us, one step before it reached a writer.

## Verified live today (2026-08-20, ~07:05 UTC)

| fact | value |
|---|---|
| Chassis | `v1.0.1317`, pods 8h |
| 295's save-path backstop still in the binary | **yes** — marker 1, positive control 3, negative control (plausible fake sha) 0 |
| `offer_ordering` | **5 of 23** sites |
| offer-analysis items | **31** — 18 complete, 8 needs_human_review, 3 failed, 2 cancelled |
| `owned_page_review` by producer | `save_page_sections` 66 (newest 08-19 11:43) · `load_page_record` **3** (newest 08-19 21:20) · `get_pages_to_build` 1 |
| `improvement-sweep` | **disabled** (last fired 08-17 12:30) |

## `bugs_open/335` — the live thread, and it is ours

`load_offer_surface` passes every page's `title` and `meta_description`. On leopardess the model
lifted *"eight live sites"* out of a page meta into `offer_ordering.lead_with[0].point` — **the
thing the site should lead with** — with `from_field: "trust_threshold"`.

- **The number is false:** `SELECT count(*) FROM sites WHERE status='deployed'` = **23**.
- **It is not in the premise:** searching every `is_current` spec for the phrase returns **only
  `offer_ordering` itself**. Two pages carry it in `meta_description`. So the analyser is the first
  spec-layer carrier of a page's claim.
- **The attribution is the defect, not just the staleness.** `from_field` is this lane's own honesty
  machinery; it names a premise field that does not contain the number, so an auditing reader sees a
  sourced claim.
- **We did not catch it.** The **leopardess lane** held all five findings at `needs_human_review`
  with a `grading` note naming the stale figure. The artefact passed every structural check B4 makes.

**Preferred fix (in the file):** a cardinal in a `lead_with[].point` must appear in its cited
`from_field`, asserted at write time — makes the bad state unrepresentable rather than asking the
model to be careful. ⚠ **Verify with a NEGATIVE control** (gaswholesalers' legitimately
premise-sourced specifics must survive), or "no numbers anywhere" passes trivially and guts the
artefact.
⚠ **Do NOT fix it by dropping `meta_description` from the surface** — those metas are load-bearing
for the surface's real job (two of the first five gaswholesalers findings were grounded in
missing/generic metas).

## What changed beneath this lane while it was away

- **`bugs_open/301` → `bugs_closed/301`.** Another lane took, fixed, council-approved (round 2) and
  rolled it: the owned-page guard now refuses at `load_page_record`, before the LLM writer. Proven
  on live demand both ways. **This lane's contribution is the CONTRIB block in that file** (the
  per-run join it declared unmeasured). **Do not re-take it.**
- **`bugs_open/333` FILED** on the owner's decision — the residual both `295` and `301` left
  untaken: producers route content findings at `page-build-handler` without reading
  `rebuild_policy`, so owned pages queue findings that can only ever be refused. Their census:
  **142 open on 57 pages / 9 sites**, largest producer the **tool pipeline itself**. Another lane's;
  contribute, do not compete.
- **`bugs_closed/295` gained a cross-reference** from the 301 lane (additive only — nothing of ours
  altered), recording **0 save-path refusals since the 08-19 roll**, which is the earlier guard
  working exactly as this lane's CONTRIB predicted.
- **Two of this lane's items were cancelled** under an owner directive as casualties of a fleet-wide
  `git-adapter` 404 burst (2026-08-17 13:31–16:14Z). That window also forced a correction to
  `bugs_closed/295` — see it.

## What the next session should do

1. **Take `bugs_open/335`.** It is this lane's own agent, it is a live false claim on a live site,
   and the fix batches naturally with **v2(b)** — which this defect promotes from "intermittent,
   does not justify a migration alone" to load-bearing.
2. **Then the v2 batch** (`features_open/030` §10) — one migration, one re-proof:
   **(d)** machine-checkable acceptance predicates (strongest; census says ~8 of 22 tests are
   expressible and the one that failed is among them) · **(b)** attribution in `why` clauses ·
   **(a)** head-of-hero excerpt ⚠ *invalidates v1's truncation baseline — re-run that check on
   webdesign.co.uk after it* · **(c)** `primary_model` in the degraded arm ⚠ *latent, no live
   instance — must not motivate the batch*.
3. **`features_open/034`** — claims audit over `site_specs` prose, owner-approved 2026-08-14, still
   not designed. **`335` sharpens its case and does not replace it:** 034 would catch this *after*
   the write; 335's fix stops the import.
4. **Sweep window** — 18 sites still lack a ranked record. ~1 site / 15 min, every site currently
   `audit_due` so every visit is the expensive shape; **≈4.5 hours to finish the estate.** Owner's
   call. Enable by direct `UPDATE`, never a migration, and **disable in the same session.**

## Watch-outs (the ones that have actually bitten)

- **⚠ psql prints UTC, your shell prints BST.** Always toward alarm. Make the DATABASE subtract
  (`now() - last_activity`). I made this error twice in twenty minutes, the second time after
  writing the landmine about it.
- **⚠ `count(*) = count(DISTINCT item_key)` is the WRONG dedup test** — `idx_swi_dedup` is
  `UNIQUE (site_id, item_key)`, per site.
- **⚠ A new producer on a deduped item type reads ZERO while working** — the first filer keeps
  `refused_by`. **Corrected 2026-08-20:** suppression is per SUBJECT, not global, which is worse — a
  non-zero count "proves it works" while saying nothing about the repeat cases.
- **⚠ A cause you did not read is not a weaker version of one you did.** I proved one failure and
  let its neighbour ride on the resemblance; an incident log later supplied a competing explanation
  and retention had closed. `WRONG_CALLS.md`.
- **⚠ A roll is not a deploy.** Same-tag rebuilds serve the cached image — one such roll shipped
  **none** of a day's 222 commits. Probe `/proc/1/exe` with a negative control **capable of being
  absent** (a plausible fake sha, never 40 zeros).
- **⚠ A site with `created_at` today, 0 pages, or `status='active'` is UNDER CONSTRUCTION** and is
  not a fact about the estate.

## Who owns what nearby

`bugs_open/333` and `bugs_closed/301` belong to the 301 lane — contribute, do not compete.
**`bugs_open/335` is ours.** The **leopardess lane** is actively working that site and is holding
our findings pending an owner design report — **coordinate before firing B4 there again.**
`copy_quality_two_stage` + the LMC lane still work loanandmortgagecalculator.co.uk.
