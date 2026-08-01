# 081 — A **deployed** but mistyped page has no repair path: the retype arm cannot see it and `new_page` overwrites it without fixing `page_type` (OPEN, live instance)

**Found:** 2026-07-26 while closing `bugs_closed/015`. This is the half of the
mistyped-`page_type` class that 015's fix does **not** cover, and it has a live
instance with an open work item that has already failed once.

## The gap in one line

015's repair path (`retype_existing`) only considers **stranded** pages —
nav-visible, `sections=[]`, and **`build_status <> 'deployed'`**. A page that is
mistyped *and already deployed* matches none of that, so nothing can re-type it;
and the fallback arm silently makes things worse.

## Mechanism — the two doors are both shut

**Door 1, the retype arm, cannot see it.** `findStrandedNavPages`
(`check_news_feed.go:682-692`) requires `COALESCE(build_status,'') <> 'deployed'`:

```sql
WHERE site_id = $1
  AND COALESCE(page_type, '') <> 'news-index'
  AND (COALESCE(in_header, false) OR COALESCE(in_footer, false))
  AND jsonb_array_length(COALESCE(sections, '[]'::jsonb)) = 0
  AND COALESCE(build_status, '') <> 'deployed'
```

That clause is **correct and deliberate** — it was added by 015 after the
predicate without it flagged six live, working pages on
`ai-agent-orchestration.com` as "dead links today". The point here is not that the
clause is wrong; it is that excluding deployed pages leaves the deployed-mistyped
case with no owner. With no candidates the spec carries no `retype_candidates`,
and the prompt makes approach E conditional on exactly that: *"ONLY available when
the gap description lists stranded candidate pages"*.

**Door 2, `new_page`, overwrites the live page and leaves the mistype.** The
upsert at `apply_gap_plan_action.go:394-404`:

```sql
INSERT INTO pages (site_id, name, url, title, page_type, build_status, sections, ...)
VALUES ($1, $2, $3, $4, $5, 'planned', $6::jsonb, ...)
ON CONFLICT (site_id, name) DO UPDATE SET
        title = EXCLUDED.title,
        sections = EXCLUDED.sections,
        updated_at = NOW()
```

`page_type` is in the INSERT but **not in the `DO UPDATE SET`**. So when the
planner reuses the existing name:

- the live deployed page's `title` and `sections` are **overwritten** by the LLM's plan;
- its `page_type` is **left wrong**, so it stays orphaned from the news machinery;
- a `needs_content_page` item is filed, so the deployed page gets rebuilt;
- the site still has zero `news-index` pages, so `MissingNewsPageCheck` fires
  again on the next sweep. **The loop does not converge.**

If instead the LLM picks a *different* name, a duplicate page row is created —
the outcome 015's candidate 3 exists to prevent.

## Live evidence (queried 2026-07-26)

Two pages are mistyped and deployed, so neither is repairable today:

| domain | name | page_type | url | build_status | nav |
|---|---|---|---|---|---|
| ai-agent-orchestration.com | `news` | **`content`** | `/news.html` | deployed | in_footer |
| idea.uk | `news-index` | **`section-index`** | `/news/index.html` | deployed | header+footer |

And `ai-agent-orchestration.com` has the work item that will walk into door 2:

| item_type | status | spec.page_name | spec.page_type | has retype_candidates |
|---|---|---|---|---|
| `missing_news_page` | **detected** | `news` | `news-index` | **false** |
| `missing_news_page` | `unresolved` | `news` | `news-index` | false |

`spec.page_name` is `news`, which is exactly the existing row's name — so the
`ON CONFLICT` branch is the one that fires. **The second row is the same check
from 2026-05-01 that already ran out of attempts and went `unresolved`** — this
has been looping for roughly three months, which is the strongest evidence that
door 2 does not resolve it.

## Why this was invisible until now

015's scope measurement asked *"how many sites want a separate news page and have
no `news-index` page?"* — answer, one — and then *"how many stranded candidates
does that site have?"* — answer, zero. Both true, and together they read as "no
live damage". The question neither asked was *"is there a page already doing the
job under the wrong type that is **deployed**?"* — and there is, on that very site.

## FINDING 2026-07-27 (bugs thread) — **candidate 2's premise does not survive measurement, and this is the blocker**

I set out to build candidate 2 (a second candidate class for deployed-but-mistyped
pages, keeping 015's fail-closed authorisation) and stopped, because the predicate it
needs **cannot be written from the evidence available**. Recording the measurement so
the next thread does not re-derive it.

**The discriminator I chose, and why.** Not name vocabulary — `bugs_open/044` is exactly
about that failing on non-English sites, and it is the trap 015 already stepped in. The
structural, language-independent signal is that the page carries the `news-listing`
component: that is what a news listing *is*.

**Fleet-wide, `sections @> ["news-listing"] AND page_type <> 'news-index'` returns four
rows, and one is a false positive:**

| domain | name | page_type | sections | is it the news listing? |
|---|---|---|---|---|
| ai-agent-orchestration.com | `news` | `content` | `["hero","news-listing"]` | **yes** — mistyped |
| idea.uk | `news-index` | `section-index` | `["hero","news-listing"]` | **yes** — mistyped |
| robot-hands.com | `news-index` | `section-index` | `["news-listing"]` | **yes** — mistyped |
| robot-hands.com | **`gripper-catalog-index`** | `section-index` | `["news-listing"]` | **NO — it is the catalog index, which embeds a news feed** |

**And the shapes are byte-identical.** `gripper-catalog-index` and `news-index` on the
same site both hold exactly `["news-listing"]`. So does `webdesign.co.uk/news`, which is
**correctly** typed `news-index`. There is no section-count, ordering, or composition
signal that separates them:

```
robot-hands.com | gripper-catalog-index | section-index | ["news-listing"]   <- NOT news
robot-hands.com | news-index            | section-index | ["news-listing"]   <- IS news
webdesign.co.uk | news                  | news-index    | ["news-listing"]   <- IS news, already right
```

**The site's own config does not resolve it either.** I checked
`classification.content_features.news_feed` on both affected sites hoping it named the
intended page; it carries `recommended`, `separate_page`, `source_types`,
`vertical_keywords` and a `reason` — **no page id, name or URL**. So nothing authoritative
points at which page is meant to be the news listing.

**What this means for the candidates:**

- **Candidate 2 is blocked, not merely more work.** It needs a discriminator that does
  not exist yet. Writing the predicate anyway would offer `gripper-catalog-index` to the
  planner as a re-type candidate, and re-typing the catalog index to `news-index` would
  point `render_news_section` at the wrong page and break a live, working page. That is a
  worse outcome than the current silent loop.
- **The honest options are to CREATE the missing signal**, not to infer it: either record
  the intended page on `news_feed` config when the page is created, or accept that this
  is a human judgement and take candidate 3 (detect and route to review) — which is
  precisely the shape 015's fail-closed model was built for, with a human rather than an
  LLM choosing.
- **Candidate 1 remains what this file says it is:** it would repair the instance and
  hand broad re-type authority to a generic arm.

**This also blocks `bugs_open/080`'s residual.** The settled section-index convention says
robot-hands should keep `news-index` at `/news/index.html`, re-typed to `news-index`, and
retire `/news.html` — see 080's correction box. That repair is *decided*; what is missing
is a mechanism that can identify the page without also selecting the catalog index. Until
one exists it is a hand-repair, and both rows are deployed and live.

---

## Fix candidates (none applied)

1. **Add `page_type` to the `new_page` upsert's `DO UPDATE SET`.** One line, and it
   converts the silent no-op into a repair — the conflicting row gets re-typed to
   the planner's intended type. Cheapest by far. Risk: `new_page` is generic, so
   this lets any gap plan re-type any same-named existing page; that is a real
   widening of authority and wants the fail-closed treatment `retype_existing` got.
2. **Widen the candidate predicate to a second, separate class** — deployed pages
   whose `page_type` disagrees with the role the check is looking for — and pass
   them as `retype_candidates` with a flag saying "deployed, re-render after
   re-typing". Keeps the fail-closed authorisation model 015 established. More work,
   correct shape. Note the sections-overwrite must NOT be applied blindly to a
   deployed page with real content.
3. **Detector only** — flag "site wants role X, has a deployed page occupying it
   under type Y" as a work item for a human. Weakest, but it stops the silent loop.

Candidate 2 is the honest structural fix; candidate 1 alone would repair the live
instance but hands a broad mutation power to a generic arm.

## Immediate, contained option for the live instance

Same targeted shape as 015's original relojistas workaround — re-type the two rows
by hand and let the normal renderers pick them up:

```sql
UPDATE pages SET page_type='news-index' WHERE site_id=$1 AND name='news';       -- ai-agent-orchestration.com
UPDATE pages SET page_type='news-index' WHERE site_id=$2 AND name='news-index'; -- idea.uk
```

**Not done here — it needs an owner call**, because both pages are live and
re-typing them changes their behaviour immediately: `render_news_section` starts
emitting `data/news-archive.json` for them, and `MissingNewsPageCheck` stops
firing. For `ai-agent-orchestration.com` that is probably desirable (it closes a
three-month-old loop). For `idea.uk` note it is VM-served and its news page's
`content_data` is stale/empty (`bugs_closed/026` Defect B part 1), so re-typing
alone will not make it correct.

## Related

- `bugs_closed/015` — the mistyped-`page_type` class; fixed for the *stranded* and
  *newly-planned* cases, and its closing section names this as residual 1's sibling.
- `bugs_closed/026` — routed these same two pages here ("handed to `015`'s owner");
  this file is where that hand-off actually lands.
- `bugs_open/080` — the other 015 offshoot: `new_page` bypasses `CanonicalisePage`,
  which is what makes the "different name → duplicate row" outcome above possible.

---

# CLOSED 2026-07-31 — fixed and committed, **NOT LIVE**

> **READ THIS BEFORE CITING THIS FILE AS FIXED.** Go is inert until the chassis
> image is rebuilt and rolled. The fix is committed so the **next** build carries
> it; nothing about production has changed yet. The standing bar in CLAUDE.md is
> *fixed AND live*, and this file is being moved to `bugs_closed/` at the owner's
> instruction with that gap stated rather than papered over — the same treatment
> as `bugs_closed/167` (`306130ba3`).
>
> **The live verification is scripted and OWED**:
> `docs024_key_docs_latest/bugfix_081_deployed_mistyped_page/RUNBOOK_deployed_mistyped_page.md`
> § "OWED — verify live after the next chassis roll". Both branches, with the
> `title`/`sections` snapshot taken BEFORE the induction — otherwise "unchanged"
> and "changed back" are indistinguishable.
>
> **Council: APPROVED at round 2** — `ccd4384c-aff9-45ed-80b2-01c3ced573bb`,
> 2026-08-01 08:20Z. "approved with 4 advisory objection(s) — none high-severity";
> 13 seats reviewed, 3 abstained, 0 unreadable. Round 1 was REVISE. The verdict was
> READ before the `Council-Reviewed:` trailer was written — see § ROUND 2 for what
> each objection was answered with.

## Re-validated before any code was written, at BOTH halves

Not taken from this file — it was five days old on a tree ~30 sessions are
editing.

- **Code, unchanged at HEAD 2026-07-31.** The upsert still carried
  `DO UPDATE SET title, sections, updated_at` with no `page_type`;
  `findStrandedNavPages` still carried `COALESCE(build_status,'') <> 'deployed'`.
- **Data, still live.** Both mistyped deployed rows exactly as filed. The
  `missing_news_page` item of 2026-05-01 is still `unresolved`; a second raised
  2026-07-24 is still `detected`. The loop this file describes has run for three
  months and was still running.

> **CORRECTION to § FINDING 2026-07-27 — the discriminator is WORSE than this
> file measured, not better.** That section reports **four** rows matching
> `sections @> ["news-listing"] AND page_type <> 'news-index'`, one of them a
> false positive. Re-run 2026-07-31 it returns **five**, and the fifth is
> `robot-hands.com/learning-center-index` — `status='archived'`, and the same
> page `bugs_open/098` is about. So two of five are false positives, not one of
> four, and the archived one would have been offered to the planner as a re-type
> candidate had candidate 2 been built. **The conclusion "candidate 2 is blocked"
> holds and is strengthened.** Nothing else in that section needed correcting.

## What was done — neither candidate 1 nor candidate 2

`applyNewPage` now only ever **creates**. The upsert is
`ON CONFLICT (site_id, name) DO NOTHING ... RETURNING id`; on `sql.ErrNoRows` the
arm **reads the row it collided with** — which the old code never did, and is the
whole defect — and then branches:

| collision | behaviour |
|---|---|
| `build_status='deployed'` AND `page_type` differs | **mutates nothing.** Files `mistyped_deployed_page` at `needs_human_review` (spec carries `existing_type`, `wanted_type` and the exact remediation SQL), sets the originating item `blocked` with a message naming the conflict, returns `applied:false, reason:"deployed_page_type_conflict"` |
| deployed, `page_type` already agrees | refresh `title`/`sections` as before — the plan is for this page's actual role |
| not deployed | refresh as before, byte-for-byte unchanged |
| no collision | create, as before |

**Why this is available when candidate 2 is not.** Candidate 2 needs a predicate
that says *which* page should hold a role, and this file proved none exists. The
refusal needs no such predicate, because **the planner has already named the
page**: we never have to work out which page is the news listing, only to notice
that the name is taken by a page holding a different role. That asymmetry is the
whole reason this was buildable today.

**Why not candidate 1** (add `page_type` to the `DO UPDATE`). It converges and
hands a *generic* arm authority to re-type any live row it collides with — the
widening this file's own candidate list flags. The fix goes the other way: it
**removes** authority the arm should never have had. The loop still stops, but
because the item is `blocked` rather than because anything was guessed at.

**`blocked`, deliberately not `complete`.** An item stamped complete over an
untouched defect is the false green that let this run three months (016b §9,
"a `complete` work item is not a repaired artefact").

## Scope was set by a measurement

```sql
SELECT COALESCE(build_status,'(null)'),
       count(*) FILTER (WHERE sections @> '["news-listing"]'::jsonb
                          AND page_type <> 'news-index')
FROM pages GROUP BY 1;
--  deployed 5 · needs_rebuild 0 · planned 0
```

**Every mistyped page fleet-wide is deployed.** A draft of this fix also re-typed
never-shipped rows; that half was cut on this count — it would have repaired
nothing extant while widening a generic arm's authority. Pinned by
`TestApplyNewPage_UndeployedTypeConflictStillRefreshes`, which fails if a later
session widens it back without re-running the query.

## Files

- `platform/orchestration/actions/apply_gap_plan_action.go` — the branch, and
  `refuseDeployedPageTypeConflict`. Result now reports `page_created` so a caller
  can tell a created page from a refreshed one (`bugs_open/091`'s treatment, one
  field over).
- `platform/orchestration/actions/apply_gap_plan_deployed_conflict_test.go` —
  **1 firing branch + 3 controls.** The pairing is deliberate: a test that only
  proves the refusal fires is satisfied by deleting the guard and refusing
  everything.

  > **CORRECTED 2026-08-01 — the first version of this line was WRONG, and so was
  > the test.** It said the load-bearing assertion was `ExpectationsWereMet()`,
  > "an unexpected `UPDATE pages` fails it, so nothing-was-mutated is checked, not
  > asserted." `ExpectationsWereMet` reports registrations made and not consumed —
  > it never sees an EXTRA call. **Induced and confirmed: an `UPDATE pages` added
  > to the refusal path passed the test.** The claim was a decoration, and I had
  > written it into this file, 016b, the commit message and the council
  > submission.
  >
  > **Caught by** a `LANDMINES.md` entry another session appended hours earlier
  > (`bugs_open/162`'s lane) — "`mock.ExpectationsWereMet()` is NOT 'no database
  > call happened'" — which arrived in this tree inside my own commit
  > `89e037a31`, because a pathspec commit takes the working tree's copy of a
  > shared file. I read it because the pre-commit hook flagged that commit for
  > removing lines from an append-only ledger.
  >
  > **Now real, and proved both ways:** the refusal path runs in a transaction
  > that checks and propagates every statement's error, so the same induced
  > `UPDATE pages` now FAILS at the caller. Error propagation carries the negative
  > assertion; the mock's bookkeeping never could.
- `platform/orchestration/actions/discovery_checks/verifier_coverage_test.go` —
  `mistyped_deployed_page` classified on the way IN. Produced outside that
  package, so neither the sensor (scans that package's source) nor the ratchet
  (a DB snapshot) could have seen it.
- Concept register `WII-008`; 016b §9 (the `ON CONFLICT DO UPDATE` pattern);
  `WRONG_CALLS.md` (three); workstream docs under
  `docs024_key_docs_latest/bugfix_081_deployed_mistyped_page/`.

## Still owed, and NOT done here

1. **The live verification** above. Until it runs this is committed, not proven.
2. **The two live mistyped pages are untouched.** Re-typing them changes what
   they serve immediately and **needs an owner call**, exactly as this file's
   § "Immediate, contained option" says. The fix makes the platform *ask* instead
   of loop; it does not answer for the owner. `idea.uk` additionally will not be
   correct from a re-type alone (`bugs_closed/026` Defect B part 1).
3. **`bugs_open/080`'s residual is still blocked** on the same missing signal —
   this fix does not create it, and deliberately does not pretend to.
4. **No verifier** for `mistyped_deployed_page`. One is writable
   (`pages.page_type = spec.wanted_type`); recorded as `catMechanical` rather
   than written, so it is a decision on the record.
5. **The sibling class is filed, not fixed** — `bugs_open/172`. The council's
   `bug_historian` seat asked whether a third write path shares this shape; the
   grep says **four more do** (`create_report_page_action.go:164`,
   `deploy_tool_action.go:376` and `:514`, `create_tool_component_action.go:416`)
   and a sixth carries the opposite risk (`apply_adoption_plan_action.go:532`
   re-types on conflict). Deliberately not fixed here: six call sites in one bug
   patch is the scope creep the guardian seat exists to veto, and the right
   answer differs per site.

---

# ROUND 2 — the council said REVISE, and what changed

`ccd4384c` came back **REVISE**, gated by the `guidelines` seat at HIGH. Four
seats (`guidelines` HIGH, `improvement_guardian` HIGH, `guardian` MEDIUM,
`reuse_agent` MEDIUM) converged on **one** thing and they were right:

**The refusal hand-rolled `INSERT INTO site_work_items ... ON CONFLICT DO
NOTHING` instead of going through `insertWorkItem`.** A bare `ON CONFLICT DO
NOTHING` does not match `idx_swi_dedup`'s partial predicate, carries none of the
two-strike anti-churn, and ignores `recurrenceExpected` — so the dedup this file's
own rationale claimed ("a re-firing check dedups onto the open decision") was not
actually guaranteed. Now routed through `insertWorkItem`, which owns that idiom.

That change required a `*sql.Tx`, which **also** discharged the `debug_historian`
seat's MEDIUM: the read-then-write on `pages.build_status` was unlocked and raced
the sweep that publishes a page. It is now one transaction with `SELECT ... FOR
UPDATE`.

**Answered with evidence rather than changed:**

- `recurrenceExpected` (editquality / architecture / guardian): left **false**,
  deliberately. It exempts items whose re-request is normal; this is a detected
  defect, not an action request. While the decision is open the row is
  non-terminal so dedup blocks a duplicate and two-strike is never reached — and
  if a human resolves it and the collision returns, "we fixed this and it came
  back" is exactly what the two-strike label is for.
- `handler_agent` (guardian, HIGH — "a live handler_agent risks the dispatch loop
  re-claiming this item"): the seat read my abbreviated sketch's `created_by`
  value as `handler_agent`. It was never set and is empty. **Measured rather than
  asserted:** `claim_work_item_action.go:102` and `load_work_item_actions.go:632`
  both claim `status IN ('triaged','approved')`, so a `needs_human_review` row
  cannot be picked up at all.
- `applied` consumers (prior_art_librarian, MEDIUM — an unevidenced absence, and
  a fair hit): **exactly one** active `agent_definition` uses `apply_gap_plan`
  (`content-gap-planner`), its step is `"next_step": "complete"` unconditionally,
  and **zero** active definitions reference `applied`/`page_created` anywhere.
- `pages.build_status` really holds `'deployed'` (editquality, LOW — the sibling
  landmine says `site_components.build_status` never does): `GROUP BY build_status`
  over `pages` returns `deployed 453, needs_rebuild 45, planned 17`.
- `mistyped_deployed_page` never observed live (prior_art_librarian, LOW):
  `SELECT count(*) ... WHERE item_type='mistyped_deployed_page'` → **0**.

**Not done, and named:** the `tooling_provenance` seat asked for a
`doc_notes`/`doc_plans` consultation on this subject before editing. The
workstream docs under `bugfix_081_deployed_mistyped_page/` are the NOTES entry it
asks be left; the prior-decisions lookup was not run, and that is a real gap
rather than a discharged one.


---

# ROUND 2 VERDICT — APPROVED, and the four advisories discharged rather than banked

`ccd4384c` round 2, 2026-08-01 08:20Z: **approved with 4 advisory objections, none
high-severity.** 13 seats, 3 abstained, 0 unreadable. `guidelines` and
`improvement_guardian` — both of whom raised the round-1 HIGHs — moved to
**approve** with zero objections, as did `reuse_agent`, `debug_historian`,
`adoption_guardian`, `render_guardian`, `constitution` and `mission`.

Three of the four advisories were one command each, so they were answered rather
than recorded:

- **`refuseDeployedPageTypeConflict` changed signature from `*sql.DB` to
  `*sql.Tx` — any caller outside this file?** (guardian, architecture)
  `grep -rn "refuseDeployedPageTypeConflict\|resolveNewPageConflict" --include=*.go .`
  → **every hit is in `apply_gap_plan_action.go`** plus two comment references in
  its test. Both functions are file-local and unexported. No compile break.
- **"the plan ASSUMES `insertWorkItem` takes a `*sql.Tx` on the strength of a
  round-1 remark, not a fresh symbol check"** (prior_art_librarian) — a fair
  challenge, and the check is one line:
  `load_work_item_actions.go:1154  func insertWorkItem(ctx context.Context, tx *sql.Tx, item workItem, logger *zap.Logger) (bool, error)`.
  Three other call sites already pass a `Tx` (`evidence_citations.go:414`,
  `seed_build_queue_action.go:165`, `emit_imagery_items_action.go:121`), so this
  is the established pattern rather than an assumption. It also compiles, which
  is the strongest available proof of a signature.
- **The `doc_notes`/`doc_plans` prior-decisions lookup, which round 2 declared NOT
  DONE** (tooling_provenance, medium; also prior_art_librarian). **Now run**, and
  it resolves to the benign branch the seat named:
  `doc_plans` → **0 rows** for `apply_gap_plan`/`content-gap-planner`.
  `doc_notes` → 6 rows, **all of them this session's own landmine** (synced from
  `LANDMINES.md`), plus one unrelated entry that merely carries
  `apply_gap_plan_action.go` in its footprint (`site_specs.audience` is read by
  nothing). **No prior decision on this subject existed to be lost.**

The fourth stands and is already tracked: **`bug_historian` and `architecture`
both note the census in `bugs_open/172` shows the identical defect at four more
call sites**, and `architecture` asks whether a canonical page-upsert helper
should exist. That is `172`'s fix candidate 2, and `172` already records it as
architecture-scope needing the RFC route rather than a bug patch. Nothing further
owed here beyond not pretending this closed the class — which the file says twice.

**Trailer earned:** `Council-Reviewed: ccd4384c-aff9-45ed-80b2-01c3ced573bb`.
