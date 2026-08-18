# 307 — a transient infrastructure burst terminally kills work items, because all three attempts fit inside the outage and the transient classifier has never heard of the git adapter

**Filed 2026-08-18** by the `staged_component_build` lane, recording an **owner ruling of the
same day: "A terminal blip should return the item to queued."**

> ## Why this is filed on first-hand verification instead of a `090` run
> (per the 2026-07-31 ruling): the incident is fully enumerated from `agent_error_log` and
> `site_work_items`, and the mechanism is quoted from `FailWorkItemAction`
> (`load_work_item_actions.go:1053-1160`) read this session. Nothing inferred; every count
> below could have come out otherwise and the attempt-count split DID come out two ways.

## 1. The incident (all measured 2026-08-18)

2026-08-17 **13:31Z → 16:14Z** (2 h 43 m): every git-adapter call fleet-wide failed with
`failed to get latest commit/base tree for branch "master": github API request failed with
status: 404 Not Found`. ~**815 failed steps** across 10+ agent types (`page-rerender` 487,
`build-dispatch-loop` 253, `asset-deployer` 40, …) and 9+ domains. **Zero occurrences since
16:15Z** — the platform's own record; the cause of the GitHub-side outage is not established
here and does not matter to this bug.

Casualties: **100 work items reached status `failed`** in the window (81 `page_rerender`) and
sat there untouched for 21+ h — `failed` is dispatch-terminal (`claim_work_item` claims from
`triaged`/`approved` only; `idx_swi_handler` covers only those two). **Disposed of 2026-08-18
by owner directive: all 100 marked `cancelled`** (audit note in `result.cancelled_reason`), so
do NOT re-count them as evidence that the backlog persists — the residual defect is the
mechanism, not those rows.

## 2. The mechanism — the retry machinery exists and worked; it retried into the same outage

`FailWorkItemAction` already does the right *shape* of thing (read at
`load_work_item_actions.go:1148-1160`): `attempt_count+1`, back to `triaged`, terminal `failed`
only when `attempt_count+1 >= max_attempts` (default 3). And it already has a transient class:
`isAIUnavailable(errorMsg)` (`ai_errors.go:49`) releases to `triaged` **without** consuming an
attempt — built for exactly this situation, for LLM endpoints.

Two gaps, both measured on the 100:

1. **88/100 exhausted all 3 attempts INSIDE the 2 h 43 m burst.** There is no backoff between
   attempts — a re-triaged item is immediately re-claimable, so during an outage the loop
   burns the whole budget against the same dead dependency in minutes. Retry-without-delay is
   equivalent to no retry for any outage longer than a few dispatch cycles.
2. **12/100 died at `attempt_count = 1`** — some path writes `failed` without the CASE ladder
   (candidates: `update_work_item_status` with an explicit `failed`, or
   `complete_work_item_verification.go:381-385`'s own ladder at max already). `[UNVERIFIED]`
   which path those 12 took — the first task for whoever picks this up
   (`SELECT handled_by, item_type` on the cancelled rows narrows it in one query).
3. The git-adapter error never enters `isAIUnavailable` — that classifier is string-matched on
   LLM connection/auth patterns, and this failure is a clean HTTP 404 from a healthy adapter
   faithfully reporting GitHub. Nothing anywhere classifies "shared infrastructure dependency
   down" as transient.

## 3. The ruling, and the trap in implementing it naively

Owner, 2026-08-18: **a transient blip should return the item to queued** — not consume its
attempts, not land terminal.

The trap: the existing AI-unavailable branch does NOT increment `attempt_count`, i.e. it
retries FOREVER. For LLM credit exhaustion that is tolerable. For a **404 on a branch** it is
not: a genuinely deleted repo produces the same error as a GitHub blip, and an unclassified
infinite retry would grind on it for ever (compare `bugs_open/131`'s `sites.github_repo`
picking the wrong repo, *succeeding silently* — the same seam, misconfigured, is a permanent
404). Any fix must distinguish "the dependency is down" from "the dependency answered: no".

## 4. Fix candidates, ordered by what closes the door

1. **Detect the burst, not the string** (closes the door): the signature of an infrastructure
   outage is MANY items failing with the SAME error across DIFFERENT sites/types in a short
   window — visible in `agent_error_log` at fail time in one query. On match: release to
   `triaged` with `attempt_count` untouched AND stamp a cooldown (`spec.retry_after` or a
   `blocked` status until a timestamp), so the retry happens AFTER the storm, not during it.
   No string list to maintain; a genuinely-missing repo fails alone and still exhausts its 3
   attempts honestly.
2. **Backoff between attempts** (smaller, complements 1): make re-triage set a
   not-claimable-before timestamp scaling with `attempt_count` (e.g. 5 m / 30 m / 2 h). Even
   with no classification at all, 3 attempts then span ~2.5 h+ and most outages are survived.
3. **Extend `isAIUnavailable`-style classification to adapter infrastructure errors** (weakest:
   a string list that will miss the next novel error, and the 404 ambiguity in §3 sits exactly
   on its edge). Acceptable only as a stopgap inside 2's attempt budget, never with the
   count-free branch.
4. Whatever ships: the 12-at-attempt-1 path (§2.2) must be found and put through the same
   ladder, or the fix covers 88% of the incident and reads as covering all of it.

## 5. How to verify

Replay-shaped test: N items, a failing dependency, assert (a) no item reaches `failed` while
the burst detector/backoff holds, (b) all items complete after the dependency recovers, (c) a
single item with a permanent 404 and no burst STILL reaches `failed` in 3 attempts — the
disconfirming case that keeps the fix honest. Live: next adapter outage should leave zero
terminal `failed` rows attributable to it.

## 6. Relations

Owner ruling 2026-08-18 (this file) · the 100-item disposition (cancelled, same day) ·
`ai_errors.go` (the existing transient mechanism this generalises) · `bugs_open/131`
(wrong-repo landmine on the same seam) · MEMORY `order-fix-candidates-by-what-closes-the-door`.

---

## Contribution, 2026-08-18 (session `bugfix-277/083`) — a THIRD gap in the same UPDATE: `FailWorkItemAction` is the only one of the three work-item status writers with no terminal-status guard

**Not a rival diagnosis, not a fix attempt, and not filed as its own bug** — `307` owns
`FailWorkItemAction` and a second file would drift from it. I arrived from the other end
(shipping owner decision 1 of 2026-08-18, which makes an ownership refusal record `wont_fix`
instead of `failed`) and a council reviewer's gating objection sent me to read exactly the lines
this bug quotes. What I add is a defect in the same statement, adjacent to your two.

### The asymmetry

Three code paths write a work item's terminal status. Two of them refuse to overwrite a status a
handler deliberately set; the third does not, and it is this one.

| writer | guard |
|---|---|
| `CompleteWorkItemAction` (`load_work_item_actions.go:1017-1025`) | `AND status NOT IN ('needs_human_review','failed','unresolved','rejected','wont_fix','verified','blocked')` |
| `failUnverifiedCompletion` (`complete_work_item_verification.go:428-429`) | **the identical list** |
| **`FailWorkItemAction` (`load_work_item_actions.go:1146-1160`)** | **none — `WHERE id = $1`, nothing else** |

`CompleteWorkItemAction`'s guard carries its own reasoning in a comment: *"do NOT overwrite a
status a handler deliberately set to a flagged or terminal state … without this guard a handler
that flagged its item `needs_human_review` … would be re-stamped 'complete' here, silently undoing
the flag."* **Every word of that applies to the failure path too.** A handler that has just
written `needs_human_review`, `wont_fix` or `rejected` through its own `update_work_item_status`
step, and whose saga then errors on the way out, has that decision replaced by
`triaged`/`failed` — and `triaged` means it is re-dispatched to be refused again.

### How often, measured — small, and NOT zero

[MEASURED 2026-08-18] on the population I was working, `page-build-handler`'s owned-page
refusals. `handled_by` is the tell: it is written **only** by `fail_work_item` and
`complete_work_item`; `update_work_item_status` never writes it, so a NULL means the handler's
own write was never touched.

| `handled_by` | status | rows |
|---|---|---|
| **NULL** — the handler's own write, untouched | `failed` | **113** |
| NULL | `cancelled` | 3 |
| **`build-dispatch-loop`** — went through `fail_work_item` | `failed` | **2** |

So **~98% of refusals take the guarded path and ~2% do not**, the latter both on 2026-08-15.
Small, but it is the same shape as the 88/100 in §2: a mechanism that is right almost always and
silently wrong in the tail.

⚠ **Why this is invisible in today's data, and why that matters to YOUR fix.** Both paths
currently write `failed`, so no query on the existing rows can tell "the handler said failed and
it stood" from "the loop overwrote what the handler said". The 2 rows above are visible only
because `handled_by` happens to distinguish the writers. **Any change that makes a handler write
a status other than `failed` — including the one I have just shipped, and including anything
`307` does about returning items to `queued` — turns that invisible overwrite into a visible
wrong answer.**

### What I am NOT proposing

I have deliberately **not** added the guard, and the reason is the one this estate keeps filing
bugs about: `fail_work_item` is a shared action reached by every dispatch loop, and adding
`status NOT IN (…)` to it changes the retry ladder for all of them. In particular `failed` is
itself in the sibling list, so a naive copy would stop `fail_work_item` re-triaging a `failed`
row — which may be exactly right, and is exactly the kind of decision that belongs with whoever
owns the retry semantics rather than inside someone else's patch. Owner ruling 2026-07-28: a
change to a shared mechanism arriving inside a bug patch draws a veto, and rightly.

**What the fix probably wants** (your call, not mine): the guard list on the failure path is
almost certainly NOT the same list as on the completion path. `failed`/`unresolved` should
probably stay overwritable (that is the retry ladder); `needs_human_review`, `wont_fix`,
`rejected`, `verified`, `blocked` — the statuses that record a DECISION rather than an outcome —
should not. That is one line and one comment, but it is a decision about what those words mean.

### Relates

`bugs_open/301` (the owned-page refusals whose statuses this can overwrite) · `bugs_open/083`
(the promoter floor that reads those statuses) · register **WII-003** (the flag-preservation
guard, Fix A — this is the same defect class, one writer over) · **WII-019** (the change that
made this observable) · council corr `725b1f01-f4b5-42fc-92b5-6de8fc0daa85`, whose `editquality`
seat raised the objection that found it
