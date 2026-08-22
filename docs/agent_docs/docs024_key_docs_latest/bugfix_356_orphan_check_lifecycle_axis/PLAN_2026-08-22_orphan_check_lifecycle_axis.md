# PLAN 2026-08-22 — the orphan check selects on the BUILD axis only, so retired pages are filed as work

**Lane opened** 2026-08-22 from the session named `bugs_open/298`.
**Origin:** `bugs_closed/298`'s explicitly unclaimed adjacent finding — *"15 of 38 completed
`internal-linker` items found NO target page"*. 298's own subject (the 15-row candidate cap)
is **fixed, live and proven**; this lane takes the residual, which is a different defect.

---

## 1. What 298 asked, and what the answer turned out to be

298 recorded, without chasing it, that a large minority of `internal-linker` runs completed
having found no target page at all, and noted that a run which completes having linked
nothing is indistinguishable, from `status`, from one that linked successfully.

Re-measured 2026-08-22 with the correct JSON path (`result->'response'->'target_page'`, not
`result->'target_page'` — see NOTES misstep 1):

| completed `internal-linker` items | count |
|---|---|
| total | 34 |
| **found NO target page** | **17** |
| found a target page | 9 |
| unreadable (`result` is the spawn record — `bugs_closed/287`) | 8 |

**All 17 of the no-target items name a page whose `pages.status` is `archived`.** Not most —
all of them. That is not a coincidence to be explained; it is the mechanism.

## 2. The mechanism

`OrphanPagesCheck` (`platform/orchestration/actions/discovery_checks/check_orphan_pages.go`)
selects candidate orphans with `findOrphanPagesSQL`, whose only page-row predicate is

```go
AND ` + datahelpers.PageHasShippedPredicateFor("p") + `
```

That is the **BUILD** axis — "has this page ever been served". It carries **no LIFECYCLE arm**
(`datahelpers.PageWantedLivePredicateFor`, i.e. `status = 'active'`). The predicate family's own
doc comment (`platform/orchestration/datahelpers/links.go`) states the contract plainly:

> Pair this with whichever build-axis arm YOUR question needs; do not expect one combined helper.

The orphan check pairs nothing. So an **archived** page that was deployed once still satisfies
the predicate, is enumerated as an orphan, and is filed as a `site_work_items` row for a handler.

## 3. Why this is a producer defect and not three handler defects

The check has three routing branches. **Every one of the three remedy paths already applies a
lifecycle arm. The producer is the single outlier.** Each cited first-hand:

| branch | item type → handler | the remedy path's OWN page predicate | what an archived page does there |
|---|---|---|---|
| unflagged | `needs_internal_links` → `internal-linker` | live `agent_definitions` step `load_target_page`: `... AND p.status = 'active'` | target resolves empty → `check_target_found` takes `else_step` → run completes having done nothing |
| nav-flagged | `nav_drift` → `nav-updater` | `navPageScopeSQL` (`platform/orchestration/actions/nav_prune_floor.go:128`): `status IN ('active','deployed','pending')` | the nav rebuild **cannot** add the page, so the finding is **unsatisfiable** and recurs for ever |
| blog | `orphan_blog_posts` → `rerender-pages` | `blogPostsQuery` (`platform/orchestration/actions/rebuild_blog_listing_action.go:110-111`): `p.status IN ('active','deployed')` | the listing **cannot** include the page — same unsatisfiable shape |

So the producer is asking three different handlers to make a **retired** page more reachable, and
all three correctly refuse. The disagreement is entirely inside the platform.

## 4. The damage, measured 2026-08-22

Fleet census, running `findOrphanPagesSQL`'s own predicate verbatim and splitting by
`pages.status` (query in the RUNBOOK):

| page status | routing branch | pages | sites |
|---|---|---|---|
| active | blog | 63 | 16 |
| active | needs_internal_links | 1 | 1 |
| **archived** | **blog** | **16** | **3** |
| **archived** | **nav_drift** | **3** | **2** |
| **archived** | **needs_internal_links** | **15** | **4** |

**34 archived pages across the three branches are being filed as work right now.**

It recurs, because `idx_swi_dedup` excludes terminal statuses, so each discovery rotation
re-raises the same key:

```
needs_links:gripper-cycle-time-estimator:00ff3af5…   filed 4×   2026-07-17 → 2026-08-17   archived
needs_links:gripper-payload-calculator:00ff3af5…     filed 4×   2026-07-17 → 2026-08-17   archived
needs_links:case-study-news-pipeline:4851f6fc…       filed 3×   2026-04-24 → 2026-08-17   archived
```

Four months of the same retired page being re-detected, re-dispatched and re-completed.

**Second-order cost, already documented by another lane:** these no-op completions consume the
two-strike ladder, and `bugs_closed/313` established that this is what parked the linker's 20
queued items at `status='unresolved'` — terminal and undispatchable. So the producer defect
does not merely waste runs; it *retires the queue* for the pages that are legitimately affected.

## 5. What the fix must NOT be

`bugs_open/266`'s note from the `bugfix_168_deployed_asset_path` lane (2026-08-14) warns, in
terms, against the obvious-looking companion fix:

> **Do NOT let anyone "fix" the audits by filtering on `pages.status`.** … an archived page can
> be live, so filtering it out would stop auditing a page that really is asserting unsupported
> claims to the public. The audit is right to look.

That warning is correct and this plan does not contradict it. The distinction is **what the
finding's remedy does**:

- `check_unverified_claims` **observes** — it flags copy on a page that may still be serving to
  the public. Archived-and-serving is exactly the case it must not miss. No lifecycle filter.
- `check_orphan_pages` **prescribes a reachability change** — every one of its three remedies
  makes the page *more* reachable. Applied to a retired page that is still serving, the correct
  action is retraction (`bugs_closed/098`), never re-linking. Lifecycle filter.

This is the predicate family's stated doctrine — same sources, different questions, opposite
safe-failure directions — not an exception to it.

## 6. Fix candidates, ordered by what closes the door

**Candidate A — the individual case (necessary, not sufficient).**
Add the lifecycle arm to `findOrphanPagesSQL`:

```go
AND ` + datahelpers.PageHasShippedPredicateFor("p") + `
AND ` + datahelpers.PageWantedLivePredicateFor("p") + `
```

Two lines. Makes the producer agree with all three of its own remedies. Closes today's 34 rows.
Alone it fixes one check and teaches the next author nothing.

**Candidate B — the framework fix: make the omission unrepresentable, not merely fixed.**
The defect class is *"a discovery check queries `pages` and silently takes only one of the two
axes"*. Today nothing detects that; the two axes are a doc comment, and this bug is what a doc
comment enforces. Proposed: a **declared posture registry** in the `discovery_checks` package —
each check that selects from `pages` names its lifecycle posture and the reason — asserted by a
coverage test that fails when a new check queries `pages` without an entry.

This is the estate's existing idiom for exactly this shape: `InboundLinkSurfaces`' lockstep test,
`COMPONENT_WRITE_ALLOWED`'s reason-carrying allow-list, the optional-key budget register.

⚠ **The allow-list landmine applies and must be designed against** (`LANDMINES.md`: *a declared
key silences your own detector*). An entry reading `NO_FILTER: deliberate` is a rubber stamp that
converts a live debt into a false all-clear. Mitigation: the test asserts the **SQL matches the
declared posture** — a check declaring `LifecycleFiltered` whose SQL contains no lifecycle arm
fails — so the entry records a decision the code must then honour, rather than excusing it.

**Candidate C — NOT proposed: a blanket guard at the work-item filing seam.** Considered because
`bugs_open/266` fixed its four-producer problem exactly that way, at the one seam every producer
passes through. Rejected here: §5 shows the right answer differs per check, so a seam-level
filter would break `check_unverified_claims`, which is right to look at archived pages. 266's
state (`archived` = "nothing may deploy this") has no exceptions; this one does.

**Candidate D — NOT proposed: reap the existing rows.** The 34 rows resolve themselves once A
lands (they stop being re-raised, and the terminals age out of the dedup index — the
self-healing note in `bugs_closed/313`). Deliberately not doing a one-off deletion:
`RFC_006`'s ruling is that a one-off deletion is not a class fix.

## 7. Ordering and dependencies

A and B are independent; A is the behaviour change, B is the ratchet. Both are Go, so both are
inert until a chassis image is rebuilt and rolled — no config half, nothing live on apply.

**Council:** in scope (`platform/`). B changes a shared mechanism's contract (what a discovery
check must declare), so it goes to the gate before or alongside the commit, and the consumers —
every author of a `discovery_checks` check — are told, per the 2026-07-29 ruling §3.

## 8. Open questions at time of writing

- `[PENDING]` Which OTHER discovery checks select from `pages` without a lifecycle arm? A precise
  audit is running; a grep count is not the answer, because `status='active'` on a JOINED table
  (e.g. `site_nav_items sni`) is a false positive and `check_orphan_pages` itself scores 2 on a
  naive grep while having none on the page row.
- `[PENDING]` 090 diagnosis loop verdict — intake `4480a3a7-b4cd-4026-828c-5297878dfb7f`,
  run `7bac4520-651d-41f9-aa98-f4721c49902f`, filed 2026-08-22 per the 2026-07-31 owner ruling
  (this asserts a cross-cutting structural cause).
