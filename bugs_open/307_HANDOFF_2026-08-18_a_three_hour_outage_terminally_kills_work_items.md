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
2. **12/100 died at `attempt_count = 1`** — some path writes `failed` without the CASE ladder.
   ~~`[UNVERIFIED]` which path~~ **VERIFIED 2026-08-19 (filing session): it is
   `update_work_item_status`, and the class is FIVE live agents wide.** All 12 have
   `handled_by` NULL — the tell the §third-gap contribution below established (`handled_by` is
   written only by `fail_work_item`/`complete_work_item`), so neither ever touched them. The
   item types map to their handlers' error steps exactly: `content_rewrite`/`needs_content_page`/
   `needs_page` (10) → `page-build-handler.mark_item_failed`; `needs_imagery` (2) →
   `image-build-handler.mark_work_item_failed`. Census via the nested walk
   (`jsonb_path_query(default_config,'$.**.steps')` — a top-level `jsonb_each` undercounts;
   LANDMINES): **five live agents carry an `update_work_item_status` step with
   `status: 'failed'`** — page-build-handler, image-build-handler,
   image-source-unsatisfiable-handler, image-url-404-handler, required-fields-missing-handler.
   `update_work_item_status`'s UPDATE (`v3_site_actions.go:~5496`) sets `status=$2,
   attempt_count=attempt_count+1` with **no CASE ladder and no terminal-status guard**
   (`WHERE id = $1` alone) — **one failure is terminal on these paths regardless of
   `max_attempts`=3, in fair weather as well as outages.** Consequence for the candidates in
   §4: candidate 2 (backoff) as originally written only reaches items routed through
   `fail_work_item`'s ladder; these five bypass any ladder, so the remedy must cover this
   writer too or it repairs 88% of the incident and reads as repairing all of it — which §4.4
   already demanded, and which this verification now makes concrete.
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


## §7 — 2026-08-19 (filing session): the verified 12-path plus the §third-gap contribution converge on ONE remedy shape

Three defects in one seam are now individually evidenced:
(a) `fail_work_item` retries with no delay — 88/100 burned all attempts inside one burst (§2.1);
(b) five agents' `update_work_item_status(status='failed')` steps have no ladder at all —
    one failure is terminal (§2.2, verified);
(c) `FailWorkItemAction` has no terminal-status guard where its two sibling writers do
    (§third-gap contribution) — and `update_work_item_status` has none either.

Fixing them separately produces three council rounds and leaves the seam as it was: FOUR
writers with FOUR different guarantees. The converged shape (design, not yet built): **one
work-item terminal-write contract** — attempts honoured (ladder), terminal statuses never
overwritten (guard), transient classification releasing without consuming attempts BUT with a
not-claimable-before backoff (never count-free-forever; the §3 404 trap), and the burst
detector of §4.1 as its transient signal. Owner ruling stands ("a terminal blip should return
the item to queued"). Interaction warning inherited from the contribution: the owned-page
`wont_fix` change (owner decision 1, shipped) makes any overwrite VISIBLE, so the guard half
should land before or with anything that widens status vocabulary further. This is a coherent
council-gated piece; whoever builds it should treat this section as the spec's skeleton and
the two measured populations (88/100, 12/100 → five agents) as its regression fixtures.

---

## §8 — 2026-08-20: §7's converged contract is BUILT, submitted and half-live (`bugfix_307_terminal_write_contract` lane)

**Status: Go half COMMITTED (`069015add`) and INERT until the next chassis roll; DB half
APPLIED and deliberately inert until then.** Council corr `4cdec68b-fa17-436d-8e25-8c422ee6c8c5`
(`Council-Submitted:` trailer — **no verdict claimed**). Register entry **WII-024**, written in
the same commit as the seam. Lane docs:
`docs/agent_docs/docs024_key_docs_latest/bugfix_307_terminal_write_contract/`.

**This bug stays OPEN.** The defect is still reproducible until the chassis rolls, and the
acceptance evidence in §5 is a live one: the next adapter outage leaving zero terminal `failed`
rows attributable to it.

### What was built, against §7's skeleton

One helper, `applyWorkItemFailureLadder` (`platform/orchestration/actions/work_item_failure_ladder.go`),
through which **both Go failure writers** now write. It carries all four of §7's requirements:
attempts honoured (the ladder), terminal decisions never overwritten (the guard), transient
classification releasing without consuming an attempt **but always with a not-claimable-before
cooldown**, and §4.1's burst detector as the transient signal. New column
`site_work_items.retry_after` (migration `505`), honoured by **four** read sites — `claim_work_item`
and `LoadWorkItemsAction` in Go, `build-pipeline-trigger.pre_query` and its `find_dispatchable_site`
selector in SQL (migration `506`).

§4's candidates as they landed: **candidate 1** (burst, not string) and **candidate 2** (backoff)
both shipped, together, because §7 is right that separately they leave the seam as it was.
**Candidate 3** (extending `isAIUnavailable`'s string list) was **not** done and should not be —
see the correction below. **§4.4's demand** — that the 12-at-attempt-1 path be put through the
same ladder — is met: `update_work_item_status`'s `failed` arm now calls the same helper.

### Corrections to this file, from building it

1. **§2.2's tell is contaminated, and the correction matters to anyone re-running its census.**
   `handled_by IS NULL` does not mean "written by `update_work_item_status`". A **fifth** writer
   leaves it NULL too: the `claimed-item-timeout` scheduled task (enabled, every 120s) runs its
   **own** copy of the CASE ladder in raw SQL. [MEASURED 2026-08-19] 23 `failed` rows in 14 days
   carry its `Claim timed out…` error, 18 of them `tool-auditor` at `attempt_count=3` with
   `handled_by` NULL. Two further writers set `handled_by` on `complete`-only paths
   (`plan_sections_action.go:2606`, `apply_gap_plan_action.go:1164`). The §2.2 attribution still
   holds for the rows it named, but the predicate needs `AND error NOT LIKE 'Claim timed out%'`
   to stay honest. **That task is left unchanged and is now the one remaining divergent writer** —
   named as a residual rather than quietly folded in.
2. **The outage is not the larger half of this bug.** [MEASURED 2026-08-19] **141 of 270** failed
   rows in 14 days died before exhausting their budget — **52%**, rising to **71%** over the last
   48 hours — and **139 of those 141** were at `attempt_count=1` with `max_attempts=3` and
   `handled_by` NULL, i.e. the §2.2 path, in fair weather. `content_rewrite` (61), `needs_page`
   (35) and `needs_content_page` (18) are 114 of the 159 `handled_by`-NULL failures, all routed
   to `page-build-handler.mark_item_failed`. **The incident was the alarm; the daily bleed is the
   fire.** A fix that had shipped candidate 2 alone would have repaired 88% of the incident and
   none of this.
3. **§4.1's "release to `triaged` … or a `blocked` status until a timestamp" — the `blocked` half
   is not available.** `feasibility-recheck` (enabled, every 600s) releases **every** `blocked`
   row whose handler exists, with **no timestamp condition**, and sets `error = NULL`, destroying
   the reason. That is also why the table holds **zero** `blocked` rows and always has: drained,
   not unused. `deferred` fails differently — absent from `idx_swi_dedup`'s exclusion list, so the
   row holds its dedup slot and its own detector cannot re-file it. Hence a column.
4. **§4.3 is weaker than written, and for a reason worth recording: the shared classifier does not
   match this incident either.** `RetryDisposition` (RSH-007) returns `(false, "")` for
   `github API request failed with status: 404 Not Found` — both needle lists miss it. §2.3 says
   "the git-adapter error never enters `isAIUnavailable`"; the fuller truth is that **no string
   classifier in the estate classifies it, and none should**, because a deleted repo emits the
   same bytes. This is what makes §4.1 the only candidate that can satisfy the owner's ruling for
   this class, and it is why the two classifiers are **layered** here rather than one replacing
   the other (their needle sets disagree in both directions — `EOF`/`401`/`402`/`credit`/`api key`
   on one side, `temporary`/`service unavailable`/`bad gateway` on the other).
5. **`agent_error_log.site_id` is 89.8% NULL over 7 days** — and **2,025 of 2,084 rows were NULL
   inside the 2026-08-17 outage window itself**. A burst detector keyed on distinct *sites*, which
   is the natural reading of §4.1's "DIFFERENT sites/types", **would not have fired during the very
   outage it exists for**. It keys on `domain` (0% NULL). It also cannot key on `error_code`:
   `agenterrors.ClassifyError` labels this 404 `LLM_API_ERROR`.

### The false-positive measurement §4.1 needs, and its limit

[MEASURED 2026-08-19, 7 days excluding the outage] With the conjunction ≥10 matching rows, **≥2
distinct domains AND ≥2 distinct agent types** in 10 minutes: exactly **three** signatures fire —
Kafka partition-leader (3,517 rows / 47 agent types / 23 domains), an Anthropic usage-limit, and
this git 404. **All three are genuine infrastructure outages; zero single-item permanent faults
fire**, including ones that failed repeatedly (`rebuild_protected`, `save_page_sections: SEC…`,
claims-floor blocks) — they are single-domain and single-agent-type by nature, which is the
conjunction doing its job. ⚠ **That evidence rests on ONE dominant incident**, so it establishes
that the *conjunction* discriminates, not that N=10 is the right N. Thresholds are env-tunable and
this limit is stated in the submission rather than hidden.

### §5's verification, as implemented

The replay-shaped test exists (`work_item_failure_ladder_test.go`, 15 tests — the **first** ever to
drive this path; before today `grep fail_work_item` across `*_test.go` returned two prose mentions
and no test). It includes §5(c), the disconfirming case: a lone permanent 404, same error text as
the outage, must still reach `failed` in three attempts. Five mutations were applied, run and
restored — dropping `wont_fix` from the guard list, never building the guard clause, never stamping
`retry_after`, dropping the agent-type leg of the conjunction, and releasing a `max_attempts=1`
one-shot item — each caught by a named test.

**Still owed before this bug closes:** the roll, then the artefact check (build-provenance stamp
per service + `merge-base`), then §5's live acceptance. Plus one demand control on the guard, whose
effect is currently **unprovable from data** — both writers write `failed`, so no query separates
"the handler said failed and it stood" from "the loop overwrote it". The 9 live `wont_fix` rows are
the population where an overwrite would now be visible.

### Residuals deliberately left open, with their owners

- `fail_work_item`'s `status_override` branch: untouched. `bugs_open/033` **D2** owns it, with its
  own owner ruling of 2026-07-25.
- `claimed-item-timeout`'s duplicate SQL ladder: the fifth writer, unchanged (WII-002).
- `update_work_item_status`'s six `needs_human_review` and six `complete` steps still increment
  `attempt_count` on every write. Not a failure path, so out of scope here — but it means a
  parked item's attempt count is not a count of attempts.

### §8b — corrections to §8 itself, from the council round (corr `4cdec68b`, verdict REVISE)

Four seats found real things. Recorded here because §8 is already in the shared account.

1. **A REAL CODE DEFECT, caught by `editquality` and gating the round.** §8's `complete`-arm guard
   reused the FAILURE-path list, which deliberately omits `failed` and `unresolved` so the ladder
   can move a row through them. On the COMPLETION path both must be protected — otherwise a
   `complete` write silently overwrites a row that already failed or was given up, which is the
   very defect class this change exists to close. My rationale had claimed it was *"the same guard
   its two siblings already have"*, and that was **false in exactly the two entries that mattered**.
   Fixed: a separate `workItemCompletionGuardStatuses` (the siblings' seven, plus `cancelled`),
   with the two lists sitting adjacent so the difference is visible, and two tests pinning the
   divergence — mutation-proved by reverting to the wrong list.
2. **MY MEASUREMENT WAS UNDERSTATED, caught by `guardian`.** §8's "141 of 270 in 14 days" was read
   from `site_work_items` alone, which the `work-item-archiver` drains after ~7 days — a 14-day
   claim over a 7-day table. Re-run archive-inclusive [MEASURED 2026-08-20]:
   **401 of 558 (72%)** died before exhausting their budget, **398 of them with `handled_by` NULL**
   — against the 141 of 270 (52%) I reported. The correction makes the case **stronger**, which is
   exactly why it still had to be made: an understated figure is as wrong as an overstated one, and
   I would have kept quoting it.
3. **"Reapable at 48 h" is imprecise, caught by `bug_historian` against WII-018.** Clearing
   `claimed_at` makes a re-triaged row *eligible* for `stale-work-item-reaper`, but every write
   bumps `updated_at` (trigger) and the reaper keys on it — so the clock **restarts on each
   failure**. Correct statement: *48 h after the last write*.
4. **The residual is now TRACKED, not prose** — `bugs_open/341`, filed at the insistence of three
   seats. `claimed-item-timeout` runs a fifth copy of the ladder in SQL, with no cooldown, no
   guard and no release, and it also leaves `handled_by` NULL (so §2.2's tell needs
   `AND error NOT LIKE 'Claim timed out%'`).

Not accepted, with reasons: `debug_historian` asked for a pre-state needle gate on 506's
`jsonb_set` — right as discipline, and 506 was already applied, so I **verified no clobber
occurred** (mine is the only write to that row: `updated_at` 14:14:08Z, no other lane's migration
touched that agent in the window) and added the gate to the ROLLBACK sidecar instead.
`architecture` asked for an `architecture_review` record and explicitly did not gate on it.

### §8c — 2026-08-20: council APPROVED at round 2, and the three advisories dispositioned

Corr `4cdec68b-fa17-436d-8e25-8c422ee6c8c5`, round 2: **approved, 3 advisory objections, none
high-severity**, verdict read. All three answered with a change or a measurement rather than a note:

- **`bug_historian` (medium) — the guard lists are stated as exhaustive, so an omission is a claim.**
  It was one: `deferred` — the estate's canonical parking state, **344 live rows** — was in neither
  list, while `blocked` (**0 rows**, self-unparking within 600s) was in both. Same class, opposite
  treatment, inconsistency running the wrong way. `deferred` now in both. The seat's specific worry
  (a `PAUSED_FOR_HUMAN` spelling) does **not** apply: no such constant in the tree, no such status
  in the data live or archived.
- **`reuse_agent` (medium) — is the siblings' list a named constant to import?** No: two
  independently-written **inline literals** that happen to match
  (`load_work_item_actions.go:1032`, `complete_work_item_verification.go:429`), no named constant
  anywhere in the package. So `workItemCompletionGuardStatuses` is the **first** named version of
  that vocabulary, not a fifth copy. Answered in the code where the next reader stands.
- **`guardian` (low) — does any live step write a terminal status other than `failed`/`complete`
  through `update_work_item_status`, and so bypass the ladder?** The seat noted SQL could not see
  this. It can, via the nested walk: all **17 live steps across 5 agents** configure exactly three
  statuses — `complete` (6), `needs_human_review` (6), `failed` (5). No `rejected`, `cancelled` or
  `unresolved`. **The `failed`-only scoping is complete for the live population**, not merely for
  the population I had enumerated.

Coverage: all three 307 commits are credited REVIEWED against this correlation by `098`, no
MISMATCH — the `Council-Submitted:` trailer resolved automatically on approval, as designed.

**The bug still does not close.** Nothing has changed about that: the Go half is committed and
inert until the next chassis roll, and §5's acceptance evidence is live — the next adapter outage
leaving zero terminal `failed` rows attributable to it.
