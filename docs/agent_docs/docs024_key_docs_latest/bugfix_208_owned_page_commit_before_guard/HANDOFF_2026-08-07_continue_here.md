# HANDOFF — bugs_open/208 ~~, continue here~~ — **THE ASKED-FOR WORK IS DONE (2026-08-08)**

> **CLOSED 2026-08-08, same lane.** The one remaining task this file was written for — the
> behavioural canary — **ran on 2026-08-08 with owner authorisation and every assertion passed,
> twice** (runs `a00dac64-…`, `d09ad12e-…` on oufe.com; full table in `NOTES_…`, verdict in
> `bugs_open/208`'s state section). 208 is FIXED, LIVE (v1.0.1262) and BEHAVIOURALLY PROVEN;
> the canary was cleaned up with provenance. **Do not execute the canary section below again** —
> it is kept as the record of what was run and as the template for proving a future guard of this
> shape. Still genuinely open for a future session: `bugs_open/210`, and the class sweep
> ("what else has a DB-row guard behind a git commit?") — see "Then, in priority order", items
> 2 and 3.

**Written 2026-08-07** by the `bugfix_208_owned_page_commit_before_guard` lane, at a clean
stopping point. Everything below is committed. **Cold-start reading order:** this file, then
`SUMMARY_2026-08-07_owned_page_commit_before_guard.md`, then `NOTES_…` (newest at the bottom).

---

## State in one paragraph

An `owned` (tool/widget) page swept into a **generic** page-composition loop was recomposed by an
LLM and **git-committed over the live tool** one step *before* the ownership guard refused the
database write. Fixed at two ends — excluded at selection, refused at composition — **LIVE on
chassis v1.0.1262 and pod-verified on the whole fleet.** Council **APPROVED**. The bug stays
**OPEN** for one substantive reason: **the guard has never been observed to fire.** One task
remains, and it needs an owner go-ahead because it is a real dispatch at a live site.

## What is done (do not redo any of it)

| | |
|---|---|
| Fix | `cb7b4d759` (both layers) · `6a9d85777` (narrowing) · `f5710d6b0` (council objections) |
| Council | **APPROVED**, corr `5d1dcb10-7929-431e-b9e5-496992ce3229`, 13 reviewers, 5 advisory objections, none high. Verdict **read**; 3 acted on, 1 answered by induction, 1 was a false positive we caused |
| Register | **PBP-036** in `docs026_concept_register/register/page-build-pipeline.md`, shipped in the same commit as the seam |
| Landmine | `LANDMINES.md` — "the owned-page guard belongs on `assemble_page`, NOT on `git_commit`", synced to `doc_notes` |
| WRONG_CALLS | 3 entries (my worthless test, my producer misattribution, my un-carried scoping decision) |
| Consumer told | `feature_021_operator_bulk_page_rebuild/NOTES_…` — their entry point's guarantee changed |
| Follow-up filed | `bugs_open/210` (content-failed build stamped `deployed`) |
| Live proof | pod-grep with **added, REMOVED and fabricated** strings; fleet by **digest identity** (41 pods, one sha256); baseline 13/14 byte-identical |

## THE ONE REMAINING TASK — the behavioural canary

**Why it matters:** everything verified so far proves the code is *present* and that nothing
broke. Neither is *bite*. `SELECT * FROM site_work_items WHERE item_type='owned_page_review'`
returns 12 rows, **all `source='reconcile_site_plan'`, all 2026-08-02, pre-roll — zero from
`get_pages_to_build` or `assemble_page`.** Do not let "fix live + baseline clean" stand in for
this; that is 016b §9's silent-gate trap and this lane's own docs would be the thing that misled
you.

### Do NOT use vonc.com, however tempting

It is the **only** site whose `needs_rebuild` set is owned-only (3 owned, 0 generic), so a
dispatch there sweeps in nothing else — mechanically perfect. **Refused:** those three are
`tool-arena-interface`, `tool-gauntlet`, `tool-archetype-taster-quiz`, and the arena is the page
whose clobber created the ownership marker. A canary whose failure mode destroys exactly what the
bug exists to protect is not a canary. RUNBOOK R5 already said it: **synthetic page, never a real
tool.**

### The canary, ready to run

**Target:** a site with a completely empty build queue, so the dispatch sweeps in *only* the
throwaway. Verified 2026-08-07 — `oufe.com` (9 active pages), `relojistas.com` (21),
`loancash.co.uk` (18), `loancalculator.co.uk` (26): all have **0** pages at `needs_rebuild` **or**
`planned`. Re-check before firing, it goes stale:

```sql
SELECT count(*) FROM pages p JOIN sites s ON s.id=p.site_id
WHERE s.domain='<target>' AND p.status='active'
  AND COALESCE(p.build_status,'planned') IN ('needs_rebuild','planned');   -- must be 0
```

**Step 1 — create the throwaway.** `\d pages` first (do not trust this column list):

```sql
INSERT INTO pages (site_id, name, url, title, page_type, status, build_status, rebuild_policy, sections)
SELECT id, 'zz-canary-208', '/zz-canary-208.html', 'Canary 208', 'content',
       'active', 'needs_rebuild', 'owned', '[]'::jsonb
FROM sites WHERE domain='<target>';
```

The `/zz-canary-208.html` URL is deliberately obvious: if the guard fails, that is the file to
delete from the site repo, and it is the only casualty.

**Step 2 — dispatch `page-rebuild`** at the target (the operator entry point is
`docs024_key_docs_latest/feature_021_operator_bulk_page_rebuild/scripts/rebuild_pages.sh`; read
its header first — `DRY_RUN=0` is what makes it real). Save the correlation id.

**Step 3 — assert, all of these:**

1. **Nothing built:** `get_pages_to_build` returned 0 pages and the run reached
   `complete_no_pages`, not the build loop.
2. **A review item exists:** exactly one row, `item_type='owned_page_review'`,
   `item_key='owned_page_review:zz-canary-208'`, **`source='get_pages_to_build'`** — the source
   is what distinguishes my guard from reconcile's pre-existing emission. **This is the positive
   observation the whole canary is for.**
3. **Nothing committed:** no `/zz-canary-208.html` in the site repo.
4. **The row is untouched:** `pages.updated_at` unchanged, still `needs_rebuild`, **not** stamped
   `deployed` (that last one exercises the `update_page_status` arm).
5. **The run COMPLETED**, rather than failing — pre-fix, an owned page in the set failed the whole
   workflow.
6. **Re-dispatch files no second item** (proves the dedup key), and the run is identical.

**Step 4 — the negative control the canary structurally cannot supply.** If nothing is built, a
"guard excludes everything" bug looks exactly like success. Cite the controls that already exist
rather than inventing one: `TestAssemblePage_GenericPageStillAssembles`, and the live
row-identity proof on `ai-agent-orchestration.com` (old predicate 5 rows → new 3; the two owned
tools drop, three generic blog-posts stay). **Better if cheap:** add a second throwaway page that
is `generic` + `needs_rebuild` to the same site before dispatching — then one page must build and
the other must not, in the same run, which is a real control. It costs one LLM page build and one
junk commit to clean up.

**Step 5 — clean up.** Delete both throwaway `pages` rows, cancel the `owned_page_review` item
with provenance in its `result`, and remove any committed file. Record the correlation id and the
outcome in `NOTES_…` and in PBP-036's `status-evidence`.

## Then, in priority order

1. **Update PBP-036 and `bugs_open/208` from "behaviourally unexercised" to proven** — and only
   then is 208 genuinely finished. (It **stays in `bugs_open/`** regardless: owner ruling
   2026-08-06 overrides CLAUDE.md's fixed-AND-live bar.)
2. **`bugs_open/210`** — content-failed build stamped `deployed`, so the rebuild request is
   forgotten. Filed with a *rejected* measurement recorded (the obvious proxy is confounded by
   legitimate rerenders — read that section before measuring). Its fix changes fleet-wide retry
   behaviour, so it is a decision, not a tidy-up.
3. **The class question 208's filing raised and nobody has answered:** what *else* has a DB-row
   guard sitting behind a git commit? `bugs_closed/143` is the `assets` instance, 208 is the
   `pages` instance. A sweep of actions that regenerate an artefact **and** upsert the row
   describing it would find the rest. Not started.

## Traps this lane hit, so you do not

- **`scripts/who-owns.py` says OWNED for a bug filed hours ago**, because the *filing* commit
  touches the file. It reads commits, so it cannot tell a filing from a claim, and it cannot see
  an uncommitted session at all. Grep the live `.jsonl` transcripts too (RUNBOOK R4).
- **The shared tree often does not compile, and it is usually not you.** Verify against
  `git archive HEAD` + only your own files (RUNBOOK R6). This lane's build was broken for hours by
  another session's uncommitted call to a function that existed nowhere.
- **A top-level `jsonb_each` over `{workflow,steps}` under-reports** — steps inside a loop's
  `sub_workflow` are invisible. Use the `jsonb_path_query('$.**.steps')` walk (R1).
- **A changed Go signature is a free consumer census** — the compiler found a caller
  (`WriteBuildItemsAction`) that the action-level DB query could not (R8).
- **Mutation-prove every guard, and re-mutate after a refactor** (R7). Six mutations here found
  three real defects, including a test of mine that passed with its guard deleted, and migration
  164's own refusal which had been untested since it shipped.
- **Writing a landmine for your own unshipped change manufactures prior-art evidence about it** —
  a council seat read my own hour-old `doc_notes` footprints as proof of a prior attempt at this
  bug. Harmless here; know the shape.
