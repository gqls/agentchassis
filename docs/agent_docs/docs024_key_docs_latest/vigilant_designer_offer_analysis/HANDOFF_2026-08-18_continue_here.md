# HANDOFF — vigilant designer + offer analyser (2026-08-18)

**COLD-START = this file + `features_open/030` (§10, the v2 backlog) + `bugs_open/301` (filed
today, the live thread) + `bugs_closed/295` (closed, but its residuals are 301's premise).**
NOTES tail 08-16 → 08-18 carries the evidence, the predictions-before-firing, and **six corrections
of my own claims** — most of them caught inside the session that made them.
**This supersedes `HANDOFF_2026-08-17_continue_here.md`.**

> **Re-run every liveness claim in this file before acting on it.** Not a formality. On this tree,
> in one day, I recorded: a site as missing a field (it was being built while I measured), a sweep
> as stalled (UTC vs BST), and the cadence as broken (same clock error, second time). **Every
> number below is dated; treat an undated inference as suspect, including mine.**

## The one-line state

> **B4 is enrolled, sweep-driven, and proven to grow the estate unattended.** `offer_ordering` is on
> **5 of 23** sites. **The sweep window is CLOSED** — it is the owner's cost control, and reopening
> it is his call. **`bugs_closed/295` is fixed, live and behaviourally proven.** Its fix immediately
> exposed a bigger, previously invisible defect, now filed as **`bugs_open/301`**: the platform runs
> an LLM writer and a link resolver on owned pages *before* refusing them — **39 full chains
> discarded in one night on one site.**

## Verified live today (2026-08-18, ~12:20 UTC)

| fact | value |
|---|---|
| Chassis | `v1.0.1308`, pods ~07:58 UTC |
| 295's fix present in that binary | **yes** — marker 1, positive control `OWNED_PAGE_GUARD` 3, negative control (plausible fake sha) 0 |
| `offer_ordering` | **5 of 23** sites |
| offer-analysis items | 26 total — 18 complete, 5 failed, 3 needs_human_review |
| `owned_page_review` from `save_page_sections` | **59 rows, 5 sites**, all `needs_human_review`, first 08-17 18:57 |
| `improvement-sweep` | **disabled** (last fired 08-17 12:30) |

## What the owner has open

1. **Another sweep window?** 18 sites still lack a ranked record. Measured cost: **~1 site per 15
   minutes**, every site currently `audit_due` so every visit is the expensive full-audit shape,
   ~4–5 items filed each. **≈4.5 hours of open window to finish the estate.** Enabled by direct
   `UPDATE` (never a migration — a migration would re-enable it for anyone who later runs the
   migration set), and **it must be disabled again in the same session.**
2. **`bugs_open/301`'s preferred fix needs a chassis roll** and is unowned. It is cheap (move a
   guard from step 12 to step 2) and it is the one that stops the waste.
3. **`features_open/034`** — claims-audit over `site_specs` prose. Owner-approved **2026-08-14**,
   still not designed. Unchanged by the last two days.

## What happened since the 08-17 handoff

- **`bugs_closed/295` CLOSED — fixed, live, behaviourally PROVEN.** 8 rows at the time of closing
  where there had been **zero for all history**; the **negative control** (6 content items
  completing on `generic` pages in the same window, emitting nothing) is what makes it a proof
  rather than a coincidence. Dedup held on a double refusal. Council APPROVED round 1
  (`d4f49ea5`). Register **PBP-036** updated. Moved with both paths on the commit and verified at
  HEAD with `git ls-tree`.
- **`bugs_open/301` FILED today** — see below. It exists only because 295 made refusals visible.
- **Two chassis rolls, one of which shipped nothing.** The 08-17 14:42 roll reused tag
  `v1.0.1305` and served the node's cached image: the binary was still `6a782274b` with **222
  commits** missing, and *every* Go change committed that day was inert across several lanes. The
  17:06 roll (`v1.0.1307`) was real. **Check the binary, never the roll.**

## `bugs_open/301` — the live thread, and the most valuable thing on this list

The ownership guard is the **last** step of `page-build-handler`. `rebuild_policy` is knowable at
step 2 (`load_page_record` already reads the row) and is not consulted until step 12
(`save_sections`). So an owned page gets a full `page-content-writer` LLM run and an
`internal-link-resolver` run, and is then refused.

`[MEASURED 08-18]` webdesign.co.uk, 02:30–05:00: **39 writer runs COMPLETED, 39 resolver runs
COMPLETED, 39 `page-build-handler` at `complete_error`**, 38 new review rows in the same window.
Fleet since the fix: **59 refusals in 14 hours, 49 of them webdesign — half its 97 owned pages.**

⚠ **The 39/39/39 correspondence is three aggregates over one window, NOT a per-orchestration
join** — stated as unmeasured in the bug file. Retention is ~24h, so **do that join on a FRESH
burst, not on this one.** No token-cost figure exists either; "39 runs" is a count, not spend.

## What the next session should do

1. **Take `bugs_open/301`.** Preferred fix: refuse at `load_page_record`, file the same
   `owned_page_review` row, never reach the writer — **and KEEP the save-path guard**, which is the
   backstop for other callers and removing it re-opens 295. Needs a roll. Verify with both controls
   (owned page → no writer orchestration, row still filed, item still `failed`; generic page →
   writer runs normally), or "no writer ran" is indistinguishable from a broken writer.
2. **v2(d) — the analyser's acceptance predicates.** The strongest of the four v2 items and the one
   with a live worked case: on webdesign the repair **reintroduced the exact fault it was filed to
   remove** (test said "before any count of tools or articles"; the new hero opens "Sixty-three
   browser tools"). Census of all 22 live acceptance tests: **~8 expressible as a text/DB predicate,
   ~6 partly, ~8 judgement only** — and the one that failed is in the expressible set. Ship it as a
   **per-finding opt-in, unsafe default OFF** (2026-08-02 ruling), emitting a predicate only where
   one exists. ⚠ **Never emit a predicate for a judgement test** — it would grade confidently,
   wrongly, and carry a green tick.
3. **v2(a)+(b)+(c) batch with it** — one migration, one re-proof (`features_open/030` §10).
   ⚠ (a) invalidates v1's truncation baseline; re-run that check on webdesign.co.uk after it.
   ⚠ (c) is **latent, no live instance** after a correction — do not let it motivate the batch.
4. **`features_open/034`**, the owner-approved claims audit, still unstarted.
5. **Watch the 59 review rows.** They are terminal and unclaimable so they cannot inflate dispatch
   or trip the sweep's 50-item guard (which counts only `triaged`/`detected`) — but nothing
   actions them, which is `bugs_open/115`'s shape. If 301 lands, the rate should fall sharply;
   **that fall is itself a good post-fix measurement.**

## Watch-outs added in the last two days

- **⚠ psql prints UTC; your shell prints BST.** Subtracting across them overstates every age by an
  hour, **always toward alarm** — a sweep idle 2m28s reads as a 72-minute stall, which is exactly
  what `bugs_open/294` looks like. **Make the DATABASE do the subtraction** (`now() - last_activity`).
  I made this error twice in twenty minutes, the second time *after* writing the landmine.
- **⚠ `count(*) = count(DISTINCT item_key)` is the WRONG dedup test.** `idx_swi_dedup` is
  `UNIQUE (site_id, item_key)` — per SITE. Any page name the estate reuses (`privacy`, `terms`,
  `about`, `index`, `llm-cost-calculator`) gives one legitimate row per site, so the aggregate
  under-reports by exactly the number of shared names and looks like a dedup failure. Group by
  `(site_id, item_key)`, or prove it on a repetition you caused.
- **⚠ A site with `created_at` = today, `0 pages`, or `status='active'` rather than `'deployed'` is
  UNDER CONSTRUCTION**, and nothing about it is a fact about the estate. The `[MEASURED]` marker
  does not help — it certifies the number was taken, not that its subject had stopped moving.
- **⚠ A roll is not a deploy.** Same-tag rebuilds serve the cached image. Probe `/proc/1/exe` for a
  specific marker **with a negative control capable of being absent** — 40 zeros matches every
  binary and cannot discriminate; use a plausible fake sha.
- **⚠ The sweep selector is `ORDER BY s.updated_at ASC`**, so any touch by another lane moves a site
  to the back. A "which site is next" census is stale within minutes.
- **⚠ Repeated failures on one page may be DIFFERENT producers, not one key churning.** Two
  failures 35 minutes apart on the same page were `offer-analysis_…` and `gap_plan_…`. Read the
  `item_key`s, not the timestamps — I recorded the wrong mechanism from the timestamps alone.

## Who owns what nearby

Unchanged from 08-15 except: **`bugs_closed/295` is done** (residuals live in `301`).
**`bugs_open/208`'s lane last touched it 08-08** and is not active — 301 is the same ordering class
one route over, so read 208 before designing the fix. **`bugs_open/294`** (stalled orchestrations)
is another lane's and is the bug the UTC/BST error impersonates — never file against it on a clock
reading. **`copy_quality_two_stage` + the LMC lane** still work loanandmortgagecalculator.co.uk;
do not fire at LMC without checking.
