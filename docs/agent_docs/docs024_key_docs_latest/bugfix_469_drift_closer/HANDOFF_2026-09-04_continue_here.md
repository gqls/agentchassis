# HANDOFF — `bugs_open/469`, the section-source drift closer — 2026-09-04

**START HERE.** Written for a stranger with no context. Everything below was verified on
2026-09-04 unless dated otherwise.

**Lane:** `docs/agent_docs/docs024_key_docs_latest/bugfix_469_drift_closer/`
**Bug:** `bugs_open/469_HANDOFF_2026-09-03_the_tier1_sync_down_has_destroyed_composition_on_two_live_sites_and_the_flag_describing_it_went_stale.md`

---

## 1. State in one paragraph

The **code half is DONE, LIVE and CLOSED**. There is nothing to build. There is one held
migration waiting for a ruling, and one piece of evidence still owed (the closer has never
actually fired, because the estate has no drift for it to act on).

> **UPDATED 2026-09-04 (later):** this read *"three owner decisions"*. **Q2 has since been
> ANSWERED by measurement** (§5) and it collapses the rest into **ONE binary decision**: a
> shipped, live guard (`bugs_open/266`) refuses every deploy of a `status='archived'` page,
> and it has **already refused this exact page 3 times**. So Q1 alone cannot make the repair
> render, and **Q3 is prior**: either un-archive the page (then Q1 becomes operative and
> `760` applies), or leave it archived (then `760` is **moot** and the real defect is that it
> still serves 200). There is no third branch.

## 2. What was built, and where it is

| | |
|---|---|
| commit | `fc9cad600` (the closer + seam), `2cca1b085` (council follow-ups), `5f676db58` + `152f47b65` (a correction) |
| council | `009fabca-71c8-4f7b-9b23-f0b6605eb531` — **APPROVED round 1**, 13 reviewers, 4 advisories, none high, all actioned |
| live since | **2026-09-03 22:06Z**, chassis `239ab3626` (v1.0.1360) |
| register | **WII-039** (the seam), **WII-040** (the receipt type) in `docs026_concept_register/register/work-item-integrity.md` |

**The one-sentence design**, because it is the part worth not re-deriving:

> For a divergence check, *"the finding no longer reproduces"* and *"the damage completed"*
> can be **the same observation** — the two stores agree again precisely BECAUSE the build
> overwrote one with the other. So the obvious closer (retract when they agree) is **worse
> than no closer at all**: it certifies destroyed pages as resolved, automatically and
> fleet-wide. `ResolvedFinding.Receipt` therefore makes the record of what was destroyed a
> **precondition** of the close, enforced inside `resolveWorkItems` (the shared seam, 2
> callers) rather than in the check.

**And the test is `lost`, not `direction`.** Migration `753`'s three-way label
(`cache_held`/`authority_won`/`third_list`) grades an authority that RESTORED a section and
one that DELETED one **identically** — both are `authority_won`, and both cases are real
(`oufe.com/contact` vs `robot-hands.com/gripper-catalog`). Do not "simplify" the loss
computation back to the label.

## 3. Proven at the artefact — and the instrument I got wrong first

```bash
git merge-base --is-ancestor fc9cad600 239ab3626   # YES → the fix is in the running binary
git merge-base --is-ancestor 29d611750 239ab3626   # NO  → control: a later commit is absent
```

> ⚠ **Do NOT `grep` your commit sha in `/proc/1/exe` to answer "did my fix ship".** The binary
> carries **one** stamp — the commit it was BUILT from — not a list of ancestors. My commit
> reads **absent** in a binary that fully contains it, and that absence looks exactly like
> "not shipped". The stamp probe tells you *which* commit built it; the **ancestry test** is
> what answers the question. I ran them in the wrong order and nearly filed a false negative.

## 4. Live status — four zeros, each with its demand control

`[MEASURED 2026-09-04]`, ~14h after the roll: **0** new drift items, **0**
`section_composition_lost` receipts, **0** rows carrying `result->'resolution_evidence'`,
**0** `DISCOVERY_CHECK_ERROR` rows naming this check.

Four zeros mean nothing without controls, so:

- the owning agent **ran** — `completeness-discovery-agent` filed across **9 hourly windows**;
- drift **is** genuinely absent — 398 tier-1 comparisons, all agree;
- the error query **can** find rows — **97** all-time, **5 since the roll** for *other*
  checks, none for this one.

**So: deployed, running, error-free, and NEVER EXERCISED.** That is correct on a clean
estate and it is *not* proof. **The evidence still owed is one live retraction.**

## 5. THE THREE DECISIONS — this is what the lane is waiting on

All three are about **one page**, `robot-hands.com/gripper-catalog`, and they are cheapest
answered together because the third can moot the other two.

### Q1 — May a composition repair withdraw a page's build stamp so it actually renders?

This is **`RFC_064` §7 q2** (the `427` lane's RFC, `Status: OPEN`). On this page it stops
being about provenance and becomes load-bearing: the page is
`built_from_plan_version = <current plan>` and `build_status = 'deployed'`, so
`reconcile_site_plan_action.go`'s `decideEmit` returns `skip_built` **before** comparing any
section list. **A corrected plan would sit in the database and never reach a visitor.**
Without this ruling the repair does not merely lose history — it does not happen.

### Q2 — Does the build path reach an ARCHIVED page at all? — **ANSWERED 2026-09-04**

> **CORRECTED 2026-09-04:** this read *"Unknown, not assumed."* It is now measured, and the
> answer is in two halves pointing opposite ways — which is why "does it reach it" was the
> wrong shape of question. Caught by following the page to the END of its path instead of
> the start: the sibling archived pages on this site had not rebuilt either.

**Yes, the reconciler reaches it — and no, the deploy never completes.**

- **It reaches it.** Nothing filters `pages.status`: `loadPlanPages` reads `site_plan_pages`
  (no status column) and `loadRealisedPages` is `… FROM pages WHERE site_id = $1` with no
  status predicate. `gripper-catalog` is in the current plan at `role=content, nav_order=2`.
  So Q1's stamp withdrawal *would* flip `decideEmit` to `stale` and emit a build item.
- **Then it is refused.** `archived_page_guard.go` (`580af7ff0`, 2026-08-12, `bugs_open/266`)
  refuses at BOTH `git_commit` and the `update_page_status` deploy stamp, on a literal
  `status == "archived"` test. Live in the running binary: `git merge-base --is-ancestor
  580af7ff0 239ab3626` -> YES (control `2fae8baa4` -> NO); pods run v1.0.1360.
- **It has already refused THIS page.** `[MEASURED 2026-09-04]` **308**
  `ARCHIVED_PAGE_DEPLOY_REFUSED` rows all-time; **3** name `page_id
  64fab29e-5d8a-4a50-ad1b-2f9b0721cef6` = `robot-hands.com/gripper-catalog` (joined by id,
  last 2026-08-23). Demand control: `agent_error_log` holds 243 rows since the roll.

**Consequence: Q1's ruling alone is provably insufficient**, and **Q3 is PRIOR to Q1/Q2, not
a co-equal third.** While `status='archived'` no repair can ever deploy, so the fork is
binary — and the guard's own refusal message already words it: *"Un-archive it deliberately
if it should be live, or retract it (page-retraction) if its file is still served."*

- **(a) Un-archive** -> guard stands down, Q1 becomes operative, `760` applies and renders.
- **(b) Stay archived** -> `760` is **moot**; the live defect is the serving 200, fixed by
  retraction, which the guard explicitly does NOT block (it dispatches `delete_file`).

There is no third branch in which the composition is repaired and a visitor sees it.

### Q3 — Should that page be serving at all?

It is archived **and returns HTTP 200** (`scripts/probe-page-url.sh robot-hands.com
gripper-catalog`, invented-URL control 404, sibling control 200). It is a real content page
(hero + text + info-card-grid + CTA); `gripper-catalog-index` is **not** a replacement — that
carries a single `news-listing`.

This is `bugs_closed/359`'s territory and **already has an untriaged flag**: nine
`archived_page_still_serving` items estate-wide, all still `detected` — eight from
2026-08-26/27 and one from 2026-09-02. One of the nine is **this page by id**
(`archived_page_still_serving:64fab29e-…`). **Answering Q1 alone implicitly un-retires this
page**, which is why the three go up together. *"Retire it properly"* moots Q1 and Q2.

**ADDED 2026-09-04 — the inbound-link price of each branch.** `[MEASURED]` over active pages'
`rendered_html` on this site: **0** link to `gripper-catalog.html` (the archived page), **19**
link to `gripper-catalog/` (the ACTIVE index). Controls: `gripper-selection-guide` 1, `href=`
31 — so the query can see links and the zero is real, not blindness. **Retraction therefore
breaks no internal link.** The counterweight is unchanged: the index carries a single
`news-listing` and is not a replacement, so the measured state is 19 live links pointing at a
thin index while the substantive page sits archived and unreferenced — an argument the archive
was a mistake as readily as an argument to finish it. Hence: owner's call.

### There is also a fourth, smaller decision, and it is not the owner's

**`RFC_066`** (`Status: OPEN`) asks whether the receipt's *shape* should be fixed now or left
free until a second consumer adopts the seam. Recommended: **leave it free and revisit at the
SECOND adopter**, with this RFC as the named trigger. Nothing is blocked on it.

## 6. The held migration

`docs/agent_docs/sql_for_agents/760_restore_gripper_spec_sheet_to_gripper_catalog_HOLD.sql`
(+ `_ROLLBACK`). **Re-verified 2026-09-04**: pre-checks pass, dry-run applies and rolls back
clean, the damage is unchanged, the page is still `archived`.

It restores `gripper-spec-sheet`, which the `robot hands` lane confirmed is **damage, not an
intended removal**, with provenance:
`docs024_key_docs_latest/robot_hands/SQL_2026-07-24_r9_gripper_catalog_real_grid.sql` — an
owner-backed July decision (spec cards chosen over a product grid whose empty price/rating
fields would "invite fabrication"), never reversed.

**To apply, once Q1 is ruled:** rename off `_HOLD`, **add the stamp-withdrawal statements the
ruling licenses**, re-run the pre-checks (they re-derive live state), apply, then drive a
rebuild and **verify at the artefact** — `probe-page-url.sh` plus a `gripper-spec-sheet` row
in `page_components`. ⚠ **The plan being right is not the verification. The page is.**

## 7. What would CLOSE this bug

1. Q1/Q2/Q3 ruled, and `760` either applied-and-verified or withdrawn as moot.
2. `RFC_064` decided (the `427` lane's; §5.3).
3. Ideally one live retraction observed, to move the closer from *running* to *proven*.

Nothing else. **Do not re-open the code** — it is approved, live, and mutation-proved (18
mutations, each killing a named test, sources restored and diffed byte-identical).

## 8. Traps this lane paid for — read before touching anything here

- **A census of OPEN work items cannot observe a closure.** Closing archives the row out of
  `site_work_items`, so "nothing ever closes these" is the answer that query returns
  *whatever the truth is*. Over the UNION with `site_work_items_archive`, **six of eight**
  flag-only types have real closures (`capability_gap` **1 of 334**). I wrote the false
  version into shipped code before it was caught.
- **A precedent's stated COST expires.** `check_empty_sections` argues that reading
  already-closed rows "costs a no-op UPDATE and nothing else" — true for a retraction with no
  side effect, **false** once one files a receipt. Following it unexamined would have
  re-raised two questions two lanes had already settled.
- **A mutation that PASSES may have hit a guard in SERIES.** One of mine did; the property
  held and the test's claim to pin that line did not.
- **A source-scanning test makes your COMMENTS load-bearing.** My comment quoted the pattern
  it described and the coverage guard reported this file produces an item type called
  `"literal"`.
- **`result` is a crowded shared namespace** — `revalidation` 1,822 rows, `commit_sha`
  12,909, `retraction` already owned by `write_audit_findings_retraction`. **Merge, never
  replace**, and nest under one key.

Full accounts: this lane's `NOTES` (newest at bottom) and `WRONG_CALLS.md`.

## 9. Who else is involved

- **`427`** — owns `RFC_064` and the write path. Coordinated throughout; do not duplicate it.
- **`robot hands`** — owns the site; answered §5.1 and asked that execution stay with this lane.
- **`idea.uk`** — its `guides-index` was the second near-miss; told, and settled.
