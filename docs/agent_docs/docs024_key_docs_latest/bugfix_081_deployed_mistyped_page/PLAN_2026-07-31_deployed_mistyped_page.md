# PLAN — bugs_open/081: a deployed but mistyped page has no repair path

**Started** 2026-07-31 · **Bug** `bugs_open/081_HANDOFF_2026-07-26_deployed_but_mistyped_page_has_no_repair_path.md`
**Council** `ccd4384c-aff9-45ed-80b2-01c3ced573bb` (submitted before commit)

## Why this bug, out of 60 open ones

Selection was a real cost here and is worth recording, because the cheap
ownership checks are all lagging and this estate runs ~30 sessions concurrently.

`scripts/who-owns.py` reads COMMITS, so a session mid-fix is invisible to it —
it returned "OWNED or recently active" for almost every candidate, which is
useless as a discriminator when 55 transcripts are live. Counting bug filenames
in transcripts is worse than useless: every session runs `ls bugs_open/`, so the
counts are dominated by directory listings and every bug looks equally hot.

**What worked: count `bugs_open/NNN`-shaped references across every transcript
touched in the last 4 hours, then rank ASCENDING.** That is engagement, not
listing. 081 came back at **heat = 0** — the only open bug with no mention
anywhere in the fleet — against 1382 for 149 and 1118 for 138.

Four candidates were abandoned mid-assessment on this evidence, and each would
have been a collision:

| candidate | why abandoned |
|---|---|
| `155` (deploy_image_asset resolves by purpose) | transcript `693556a1` had 57 hits on `resolveStorageURIFromAsset` — a live session, actively fixing it. Its own bug file says "OPEN, unowned". |
| `162` (fix-proposer plan discard) | `7eed05a9` live with `repair_step`×86, `persist_plan`×68. Filed as deliberately unowned; someone took it. |
| `097` (CTA integrity misses card links) | `git log` on the file showed `f6b4aea5a feat(cta): schema-derived CTA pairing` — bug 023's lane arriving at 097's fix from the other side. |
| `091` (dedup drops second drift) | `insertWorkItem` / `load_work_item_actions` appear in ~20 live transcripts, two above 100 hits. Its own file says "do not start a competing fix". |

**The transferable bit:** an ownership check that reads committed history cannot
see the session that will collide with you. Read the live `.jsonl` transcripts
for the CODE SYMBOLS you are about to touch, not for the bug number.

## What is actually wrong

`applyNewPage` (`apply_gap_plan_action.go`) is a CREATE arm whose INSERT carried
`ON CONFLICT (site_id, name) DO UPDATE SET title, sections, updated_at`.
`page_type` was in the INSERT and **not** in the DO UPDATE. So on a name
collision the arm silently became a PARTIAL update and produced a row nobody
asked for: the plan's content under the existing row's role.

On a **deployed** page that is two defects at once:

1. the live page's title and sections are overwritten by a plan written for a
   role it does not hold;
2. the mistype the plan existed to repair survives, so the check that raised the
   gap fires again next sweep. `ai-agent-orchestration.com` has looped this since
   **2026-05-01** — one item out of attempts (`unresolved`), one still `detected`.

## The decision, and why it is neither of the bug's candidates

081's own candidate 2 ("a second re-type candidate class") is **blocked and stays
blocked**: no predicate separates a real news listing from a catalog index that
embeds one. Both carry `sections=["news-listing"]` byte-identically on
robot-hands.com. I found a **third** page of that exact shape the bug file had not
recorded (`learning-center-index`, `status='archived'`), so the false-positive
rate is worse than filed, not better.

Candidate 1 ("add `page_type` to the DO UPDATE") converges but hands broad
re-type authority to a generic arm — a widening the file itself flags.

**So: take neither, and remove authority the arm should never have had.**
`new_page` now only ever CREATES. A name held by a DEPLOYED page of a DIFFERENT
type is refused: nothing is mutated, a `mistyped_deployed_page` item is filed for
a human, and the originating item is **blocked** with a message naming the
conflict.

This needs no discriminator — which is the whole reason it is available while
candidate 2 is not. **The planner has already named the page.** We never have to
guess which page should hold the role; we only have to notice the name is taken
by a page holding a different one.

## Scope is set by a measurement, not by taste

```sql
SELECT COALESCE(build_status,'(null)'),
       count(*) FILTER (WHERE sections @> '["news-listing"]'::jsonb
                          AND page_type <> 'news-index')
FROM pages GROUP BY 1;
--  deployed      | 5
--  needs_rebuild | 0
--  planned       | 0
```

**Every mistyped page fleet-wide is deployed.** So extending the re-type to
never-shipped rows repairs nothing that exists while widening a generic arm's
authority for free. Declined on that number, and pinned by a test
(`TestApplyNewPage_UndeployedTypeConflictStillRefreshes`) that fails if a later
session widens it back.

## Deliberately NOT done

- **The live repair.** Re-typing `ai-agent-orchestration.com/news` and
  `idea.uk/news-index` needs an owner call — 081 says so and it is right: both
  are live, and re-typing changes what they serve immediately
  (`render_news_section` starts emitting `data/news-archive.json`). This fix
  makes the platform ASK instead of loop; it does not answer for the owner.
- **`page_type` on the refresh path** (candidate 1). See the measurement above.
- **Same-type-deployed content overwrite.** Unchanged. When the type agrees the
  plan is for that page's actual role, so refreshing is coherent. If a live page
  should never be content-overwritten by a gap plan at all, that is a bigger
  change than this bug.

## Phasing

1. ~~Verify the bug is still valid at HEAD and in the live DB~~ — done, both.
2. ~~Measure the population that sets the scope~~ — done.
3. ~~Code + both-branches tests~~ — done, 4 tests, 3 of them controls.
4. ~~Classify the new item type in `verifier_coverage_test.go`~~ — done.
5. ~~Council submission~~ — `ccd4384c`.
6. Commit with `Council-Submitted:`; close the bug to `bugs_closed/`.
7. **OWED after the next chassis roll:** induce the refusal live (see RUNBOOK).
