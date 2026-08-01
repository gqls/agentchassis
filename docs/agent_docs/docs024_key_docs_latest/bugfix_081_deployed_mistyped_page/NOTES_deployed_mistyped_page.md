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

## 2026-08-01 — ROUND 1 VERDICT: REVISE. And the worst finding was mine, not theirs

**Council `ccd4384c` → REVISE**, gated by `guidelines` at HIGH. Approvals from
`bug_historian`, `architecture` (point_fix), `constitution`, `mission`,
`render_guardian`; objections from `guidelines`, `improvement_guardian`,
`guardian`, `reuse_agent`, `editquality`, `debug_historian`,
`tooling_provenance`, `prior_art_librarian`. 4 abstained.

**Four seats independently named the same defect** and it was a real one: the
refusal hand-rolled `INSERT INTO site_work_items ... ON CONFLICT DO NOTHING`
instead of calling `insertWorkItem`. A bare `ON CONFLICT` does not match
`idx_swi_dedup`'s partial predicate over non-terminal statuses, and carries
neither the two-strike rule nor `recurrenceExpected`. **So the dedup my own
rationale advertised — "a re-firing check dedups onto the open decision" — was
not actually guaranteed.** I had reasoned from the sibling `needs_content_page`
INSERT in the same function, which is also hand-rolled; local convention is not
the same as the contract, and `reuse_agent` said exactly that.

Routing through `insertWorkItem` needs a `*sql.Tx`, which discharged
`debug_historian`'s objection in the same edit: the read-then-write on
`pages.build_status` was unlocked and raced the sweep that publishes a page. A
page mid-deploy could be read as not-yet-deployed and overwritten as it went live
— **this fix causing the damage it exists to prevent.** Now one transaction,
`SELECT ... FOR UPDATE`.

### The misstep that matters, and no seat caught it

**My round-1 test was a decoration, and I had written its false claim into five
places.** I asserted that `mock.ExpectationsWereMet()` proved no `UPDATE pages`
was issued on the refusal path. It does not: it reports registrations made and
**not consumed**, never an *extra* call. And because my code discarded the `Exec`
error, sqlmock's own complaint about the unexpected statement went nowhere.

Induced rather than argued:

```
# round 1: add `UPDATE pages SET title=$2 WHERE id=$1` to the refusal path
ok  github.com/gqls/agentchassis/platform/orchestration/actions  0.017s      <- PASSED
# round 2, after every error is checked and propagated: the SAME edit
--- FAIL: TestApplyNewPage_DeployedTypeConflictIsRefused
    applyNewPage: induced: ExecQuery: could not match actual sql:
    "UPDATE pages SET title = $2 WHERE id = $1" with expected regexp "UPDATE site_work_items"
```

**What is actually load-bearing is the production code propagating errors**, not
the mock's bookkeeping — which means a swallowed `Exec` error is not merely a
production smell, it silently disarms the test that guards it.

**How it was caught, which is the uncomfortable part.** Not by me and not by the
council. My commit `89e037a31` tripped the pre-commit pattern check for *removing
lines from an append-only ledger* — those lines were another session's
heading-level edit on `LANDMINES.md`, riding in because a pathspec commit takes
the working tree's copy of a shared file. Reading that diff to check I had not
destroyed someone's entry is how I saw the landmine they had appended hours
earlier: *"`mock.ExpectationsWereMet()` is NOT 'no database call happened'"* —
`bugs_open/162`'s lane, who found the identical thing in four assertions the same
day. **A hook firing about an unrelated thing is the only reason this was found.**

I had written, in that same commit, three paragraphs about pairing tests so a
guard cannot be vacuous. Pairing tells you nothing about whether *either* test can
fail. Logged in `WRONG_CALLS.md`; `016b` §9 now carries the correction inline.

### Answered with evidence rather than changed

- **`recurrenceExpected`** (editquality, architecture, guardian): left `false`,
  deliberately. It exempts an item whose *re-request* is normal; this is a
  detected defect. While the decision is open the row is non-terminal so dedup
  blocks a duplicate and two-strike is never reached; if a human resolves it and
  the collision returns, that genuinely *is* "we fixed this and it came back".
- **`handler_agent`** (guardian, HIGH): the seat read my abbreviated sketch's
  `created_by` value as `handler_agent`. It was never set. **My sketch's fault**,
  and the underlying question deserved a measurement anyway:
  `claim_work_item_action.go:102` and `load_work_item_actions.go:632` both claim
  `status IN ('triaged','approved')`, so a `needs_human_review` row cannot be
  picked up by the dispatch loop.
- **`applied` consumers** (prior_art_librarian): a fair hit — I had marked this
  `[UNMEASURED]` in this very file and then asserted it in the submission. The
  query took thirty seconds and made the claim *stronger*: exactly ONE active
  definition uses `apply_gap_plan`, its step is `"next_step":"complete"`
  unconditionally, ZERO definitions reference `applied`/`page_created`.
- **`pages.build_status` really holds `'deployed'`** (editquality, low — the
  sibling landmine is about `site_components`): `deployed 453, needs_rebuild 45,
  planned 17`.

### The sibling class — filed as `bugs_open/172`

`bug_historian` asked whether a third write path shares this shape rather than
assuming it isolated. **Four more do** (`create_report_page_action.go:164`,
`deploy_tool_action.go:376` and `:514`, `create_tool_component_action.go:416`),
and a sixth carries the *opposite* risk (`apply_adoption_plan_action.go:532`
re-types on conflict). Censused and filed, deliberately not fixed: six call sites
in one bug patch is the scope creep the guardian seat exists to veto, and the two
failure modes are opposite so one answer would be wrong somewhere.

### Not done, named rather than discharged

`tooling_provenance` asked for a `doc_notes`/`doc_plans` lookup on this subject
*before* editing. Not run. The workstream docs are the NOTES entry it asks be
left; the prior-decisions half is a genuine gap.
