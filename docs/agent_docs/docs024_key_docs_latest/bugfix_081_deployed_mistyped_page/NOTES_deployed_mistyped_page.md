# NOTES — bugs_open/081 (append-only, newest at the bottom)

## 2026-07-31 — bug selection cost more than the fix, and four picks were wrong

Recording the missteps first because they are the transferable part.

**Misstep 1 — I trusted `who-owns.py` as a filter and it has no discriminating
power at this concurrency.** With 55 sessions live it returned "OWNED or recently
active" for nearly every candidate, because ANY commit touching the file in 14
days trips it. Two bugs it flagged as owned (`097`, `158`) had no owning
workstream at all; two it was quiet about were being actively fixed. The tool is
not wrong — its header says it reads commits — I used it as though it read
sessions. **The check that would have caught it:** its own output says
`[ACTIVE, N commits/14d]`, i.e. a statement about history.

**Misstep 2 — I then counted bug FILENAMES in transcripts, and the numbers were
meaningless.** Every candidate came back at 50–90 "mentions". They were `ls
bugs_open/` output. A directory listing mentions all 60 bugs equally, so the
measurement answered "how often does someone list the directory", not "who is
working this". I nearly picked on it. **Caught by** noticing that the spread was
too narrow to be real — 52 to 92 across bugs I knew differed wildly in activity.

**Misstep 3 — three bugs picked and abandoned after the symbol check.** Each was
a genuine near-collision:

- `155` — its own file says "OPEN, unowned" and the diagnosis loop had CONFIRMED
  it that morning. Transcript `693556a1` had **57** hits on
  `resolveStorageURIFromAsset`. Someone was fixing it right then.
- `162` — filed as "unowned by this lane on purpose". `7eed05a9` live with
  `repair_step`×86, `persist_plan`×68.
- `097` — `git log` on `resolve_internal_links_action.go` showed
  `f6b4aea5a feat(cta): schema-derived CTA pairing (bug 023)`. Bug 023's lane is
  arriving at 097's fix candidate from the other side.

**What finally worked:** count `bugs_open/NNN`-shaped references (not filenames)
across every transcript touched in 4h, rank ASCENDING. `081` = **0**, alone. The
next coldest was 11. Then confirm with the code symbols. Command in the RUNBOOK.

**The rule I would write for next time:** an ownership check that reads committed
history cannot see the session that will collide with you, and a bug file saying
"OPEN, unowned" is a statement about the day it was written.

## 2026-07-31 — validity re-checked at BOTH halves, and the data moved

Not assumed from the file (filed 5 days ago). Checked:

- **Code, unchanged at HEAD.** `apply_gap_plan_action.go` still had
  `DO UPDATE SET title, sections, updated_at` with no `page_type`;
  `findStrandedNavPages` still carries `COALESCE(build_status,'') <> 'deployed'`.
- **Data, still live.** Both mistyped deployed rows exactly as filed. The
  `missing_news_page` item from 2026-05-01 is still `unresolved`; a second from
  2026-07-24 is still `detected`. Three months of loop.

**The file's own measurement was stale in the direction that matters.** It
recorded FOUR rows matching `sections @> ["news-listing"] AND page_type <>
'news-index'`, one of them a false positive (`gripper-catalog-index`). There are
now **five**, and the fifth is `robot-hands.com/learning-center-index` —
`status='archived'`, and the same page `bugs_open/098` is about. So the
discriminator 081 rejected as unwritable is **worse** than it measured: two false
positives out of five, not one out of four. That strengthened the decision to
route to a human rather than build a predicate.

## 2026-07-31 — a wrong turn on scope that a query settled

I had drafted the fix to ALSO add `page_type = EXCLUDED.page_type` on the refresh
path (081's candidate 1, restricted to never-shipped rows), reasoning that it
"closes the convergence door for the non-deployed population too".

Then I measured the population and it is **empty**: all 5 mistyped pages are
`deployed`, 0 `planned`, 0 `needs_rebuild`. So that half would have repaired
nothing that exists while handing a generic arm broad re-type authority — exactly
the widening 081's own text warns about, bought for nothing.

Cut it. The undeployed path keeps the old behaviour byte-for-byte, and
`TestApplyNewPage_UndeployedTypeConflictStillRefreshes` now fails if a later
session widens it back without re-running the measurement.

**Transferable:** "it also closes the door for X" is worth a query before it is
worth a line of code. Mine would have been dead config with live authority.

## 2026-07-31 — the tests are paired on purpose

First draft had one test: the refusal fires. That test is **satisfied by deleting
the guard and refusing everything**, which is the failure mode the fleet index
already records ("verify a narrowing with BOTH branches"). Rewrote as 1 firing +
3 controls, two of which assert the `UPDATE pages` refresh genuinely still
happens. The refusal test's load-bearing assertion is `ExpectationsWereMet` — an
un-expected `UPDATE pages` fails it, so "nothing was mutated" is actually checked
rather than asserted in prose.

## 2026-07-31 — PREPARE caught nothing, and that is the point

Ran all four statements through `PREPARE` against the live schema before
trusting `go build`, per the standing rule that the compiler cannot parse SQL.
All four passed first time. Recording it because a clean result is only evidence
if you say you ran the check — and it did surface one thing worth knowing:
`site_work_items.handler_agent` is `NOT NULL DEFAULT ''::text`, so omitting it
from the column list is fine and passing NULL would have failed at runtime, in
production, on the refusal path — i.e. the path nobody exercises until it
matters.

## 2026-07-31 — what is NOT proven

- **Nothing is verified live.** Go is inert until the chassis rolls. Both
  branches are induced under `sqlmock`, which is a statement about the code, not
  about production. The live induction is scripted in the RUNBOOK and is owed.
- `[UNMEASURED]` — whether any active `agent_definition` has a `conditional` step
  keying on this action's `applied` field. I checked that no definition branches
  on it and found none, but I did not enumerate every step type, so a reviewer
  confirming independently is worth more than my grep. Flagged to the council in
  the submission's `risks`.
- The two live mistyped pages are **not** repaired. That needs an owner call and
  this fix deliberately does not make it — it makes the platform ask instead of
  loop.

## 2026-07-31 — the landmine verifier was NOT fired, and the reason is measurable

`landmines-sync.py --apply` synced the corpus (616 owned rows) and printed
`NEEDS_VERIFICATION` for the new entry, per RFC_005 §3.2. **I did not dispatch
`scripts/landmines-verify-dispatch.sh`, deliberately.**

`bugs_open/163` (OPEN, unowned) records that the verifier's symbol lookup returns
0 rows for entries whose symbols post-date the code index, and then reports
`NEEDS_HUMAN_REVIEW` blaming index staleness. Every symbol in my entry —
`refuseDeployedPageTypeConflict`, the new `applyNewPage` branch,
`mistyped_deployed_page` — was written **today**, and the index it would query is
at 2026-07-28. So the verdict is knowable in advance: a false negative, costing a
run and adding a `NEEDS_HUMAN_REVIEW` note that says nothing about whether the
entry is true.

**Stated rather than skipped silently**, because "the verifier was not run" and
"the verifier passed" must not look the same in this file. What the entry rests on
instead: the code is in this commit and readable; the `NOT NULL DEFAULT ''::text`
claim about `handler_agent` came from `\d site_work_items` on the live DB; and the
dedup-slot behaviour is `WII-005`, not my inference. Re-verify when 163 closes.

## 2026-07-31 — council verdict

Submitted `ccd4384c-aff9-45ed-80b2-01c3ced573bb` before committing, and committed
with `Council-Submitted:` rather than holding the code — the tree is shared, so
holding buys nothing and the 2026-07-29 owner ruling retires the ordering
exemption that used to justify it. Verdict recorded below when read.
