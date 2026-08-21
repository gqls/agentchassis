# 307 — a transient infrastructure burst terminally kills work items, because all three attempts fit inside the outage and the transient classifier has never heard of the git adapter

> ## ✅ CLOSED 2026-08-21 — FIXED, LIVE on v1.0.1322, and PROVEN END TO END by a full canary pass
> The whole failure-write contract (WII-024) held on one live canary, all arms (§9c): re-triage
> with a scaling cooldown (+30 m → +60 m) that **survives** the dispatch loop's completion
> (`bugs_open/344`'s fix, live), a `wont_fix` decision preserved through a live failure write
> (skip line captured from the job pod), honest terminal `failed` at 3 of 3 (the §9b 42P18 fix,
> live, council APPROVED r1 `df0748bf`), and the transient release proven earlier on natural
> traffic (§9). §5's outage-scale acceptance is converted to a **standing watch** with a reopen
> trigger, per the owner ruling of 2026-08-21 — see the closing section. Residuals live in
> `bugs_open/344` (sweep SQL half), `bugs_open/341`/migration `524_HOLD`, and `033` D2.

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

**Status: ~~Go half COMMITTED (`069015add`) and INERT until the next chassis roll; DB half
APPLIED and deliberately inert until then~~ — SUPERSEDED 2026-08-21, see §9: the roll happened
2026-08-20 16:09Z (v1.0.1320) and the whole contract is LIVE, with first natural-traffic proof.** Council corr `4cdec68b-fa17-436d-8e25-8c422ee6c8c5`
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

**The bug still does not close.** ~~Nothing has changed about that: the Go half is committed and
inert until the next chassis roll~~ *(superseded 2026-08-21 — the roll happened; see §9)*, and §5's
acceptance evidence is live — the next adapter outage leaving zero terminal `failed` rows
attributable to it.

---

## §9 — 2026-08-21: the roll HAPPENED, the contract is LIVE, and natural traffic has already exercised the transient arm

*(Recorded by the session picking the lane up post-approval; every figure below measured this
morning, not carried forward.)*

### The roll, proven at the artefact

- **v1.0.1320**, both `agent-chassis` replicas started **2026-08-20 16:09Z** — binary stamp
  `buildinfo.GitCommit=a255551e0…`, read from `/proc/1/exe` on both pods with the positive
  control (`gqls/agentchassis`, 9911 hits). Superseded the same evening by **v1.0.1321**
  (pods `agent-chassis-68ff4d794c-*`, started **19:51Z**, stamp `0483e7f4e`, control 9926).
- `git merge-base --is-ancestor` **true for all three 307 commits** (`069015add`, `5e1a0ac1e`
  round-1 fix, `29b32d0d8` round-2 advisories incl. `deferred` in both guard lists) against
  BOTH stamps — so the contract has been live continuously since 16:09Z on the 20th, including
  the round-1 completion-guard fix (the running binary never carried the wrong list).
- The DB half (migrations `505`/`506`) was applied on the 20th and became operative with this
  roll — `retry_after` is no longer structurally NULL (see below).

### First natural-traffic proof — the transient arm, end to end [MEASURED 2026-08-21 ~10:15Z]

Demand side first (a post-fix zero needs a demand control): **288** `agent_error_log` events
since 16:09Z — failures happened.

**Two `page_rerender` items took the transient-release path exactly as designed**, on a real
Kafka topic-creation blip (~18:34Z and ~18:54Z on the 20th):

| observable | row 1 (`5f52413b`) | row 2 (`6de492b9`) |
|---|---|---|
| error | `transient (ai_unavailable): … failed to create requests topic …` | same shape, responses topic |
| attempt consumed? | **no** — `attempt_count=0` | **no** — `attempt_count=0` |
| cooldown stamped? | `retry_after` 18:34:00Z | `retry_after` 18:54:33Z |
| stamp honoured? | re-claimed 18:34:25Z — **after** the stamp | re-claimed 18:56:58Z — **after** |
| outcome | **`complete`** 18:34:51Z | **`complete`** 18:57:30Z |

That is the owner's 2026-08-18 ruling — *"a transient blip should return the item to queued"* —
observed working on live traffic, un-induced: blip → released without spending an attempt →
cooldown → retry → success. Pre-fix, both rows would have burned attempts against a dead
dependency or died terminally on the §2.2 path.

**Zero terminal `failed` writes since the fix went live** — 18 hours against a pre-fix
archive-inclusive baseline of 401/558 in 14d (~29/day, 72% dying at attempt 1). Both halves of
the 341 carve-out read zero (`error NOT LIKE 'Claim timed out%'`: 0; the carve-out itself: 0).
With 288 error events on the demand side, an 18-hour terminal-failure zero is signal, not quiet.

**Guard, passive arm:** the `wont_fix` population is now 57 (301's early guard is filing);
4 touched since 16:09Z are all **new refusals** (`OWNED_PAGE_GUARD` errors, decisions standing)
— no decision-status row observed overwritten. One `needs_imagery` row completed at
`attempt_count=1` with no `retry_after` — consistent with the known parked/complete-write
increment residual (§8 residuals), **not over-read** as ladder evidence.

### What this does and does not yet prove

LIVE-PROVEN: transient classification · release-without-consuming · cooldown stamping · the
claim path honouring `retry_after` · retried items completing. NOT yet live-proven: the
attempt-consuming ladder's scaling (`triaged` + backoff×N on a non-transient failure), honest
terminal `failed` at `max_attempts`, and the terminal-decision guard skip — those three are the
subject of the owner-authorised canary (next section when run; choreography in the lane
RUNBOOK). §5's outage-scale acceptance remains a live watch by its nature.

## §9b — 2026-08-21, the canary's verdict: the ladder works, and TWO defects around it fail the acceptance. **CLOSE BLOCKED.**

Canary `f4f15466` (pool-web-tech.internal, `content_rewrite`, spec deliberately carrying neither
`page_name` nor `page_id` → deterministic hard error at `load_page_record_action.go:197`,
non-transient text, one item so the burst conjunction cannot fire). Predictions P1/P2 recorded
before the first cycle. Three cycles driven through the REAL dispatch loop; row DELETEd after,
verified 0 (the one `needs_diagnosis` filed in the window targets the NATURAL render failure on
`system.internal`, not the canary).

**What the ladder itself did — correct on every cycle it could run:**

| cycle | ladder write | fingerprint |
|---|---|---|
| attempt 1 | `triaged`, `attempt_count=1`, `retry_after` **+30 m** (10:33:32Z), claim cleared, error preserved | exactly WII-024's stated acceptance |
| attempt 2 | `triaged`, `attempt_count=2`, `retry_after` **+60 m** (10:36:40Z) | backoff **scales** 30m×N |
| attempt 3 (terminal) | **the write ERRORED — SQLSTATE 42P18** | see defect 2 |

**Defect 1 — `bugs_open/344` (filed + committed `25bb9b91a`): the loop's `mark_complete`
overwrote the re-triaged row to `complete` ~2 s after cycles 1 AND 2.** The child ends via a
success-labelled `complete_workflow`, `notifyParentOfSuccess` hardcodes `Status: "complete"`,
gate 1 (196's fix) reads exactly that, and `triaged` is in no completion guard — harmless
pre-307, load-bearing the moment this fix made `triaged` the post-failure state. **A natural
case preceded the canary by 42 s**: `0c65f9fa` (mortgagecalculator.co.uk, real render failure) —
ladder stamp 10:32:50Z, `complete` 10:32:52Z. Fingerprint: `retry_after > completed_at` on a
`complete` row. Net: the §2.2 daily-bleed population converts from honest reds to FALSE GREENS.
344 carries the eight-arm chain, the fix candidates (1: completion refuses a future
`retry_after` — discriminating and contained) and the damage census.

**Defect 2 — the ladder's TERMINAL transition has never worked: SQLSTATE 42P18, both writers.**
`computeLadder` passes backoff **0** exactly on the terminal attempt (`:397-403`);
`writeCountingLadder`'s Go-side branch then collapsed the `retry_after` fragment, dropping `$4`
from the statement text while the bind list kept five parameters — "could not determine data
type of parameter $4" at PREPARE, so the terminal write never lands, the item stays `claimed`,
a sweep later resets it, and it re-dispatches **for ever at attempt max−1** (natural victim
`95fe67da`, `needs_new_component`, remortgagecalculator.uk, cycling now; natural
`fail_work_item` hit 10:41:29Z). The 15 sqlmock tests passed throughout — a mock matches text
and never TYPES a placeholder. **FIXED this session, `bc80dde4a`** (statement + bind list
assembled together in both writers; zero-backoff folded into the SQL CASE; the dead
column-absent fallback repaired as a side effect; placeholder-audit + text-invariance tests,
both mutation-proven), `Council-Submitted: df0748bf-a22e-4e27-ba40-c70adaa11130` — **inert
until the next chassis roll.** WRONG_CALLS row filed: fifteen tests proved the writer and none
asked what the next writer does to the row, and none could 42P18.

**Guard-skip arm: not demanded.** The mid-cycle `wont_fix` flip was overtaken by the two
discoveries (cycles 1–2 completed before a flip window existed; cycle 3 errored). Passive
evidence only (refusal rows' decisions standing, §9) plus the mutation-proven tests. Owed to
whoever re-runs the canary post-roll.

**Consequence for the close.** The owner's morning ruling — close on fair-weather proof — had
as its premise that the proof would succeed. It did not: of the canary's three arms, scaling
PASSED, honest-terminal FAILED (defect 2, fixed-not-live), and the re-triage's effect is
negated by defect 1 (open, unowned). **307 stays OPEN.** Its close bar is now: the 42P18 fix
live at the artefact on the next roll, `344` fixed or explicitly dispositioned by the owner,
then the SAME canary recipe re-run clean end-to-end (all three arms + the guard flip), with
§5's outage acceptance converted to the standing watch as already ruled.

## §9c — 2026-08-21 evening: §9b's bar is MET — one canary, every arm, on v1.0.1322. CLOSED.

The 16:54Z roll (v1.0.1322, stamp `bac189921…`, both replicas, positive control) carries BOTH
repairs: the §9b 42P18 fix (`bc80dde4a` + `d97557a04` + `54032b2dd`; council **APPROVED round 1**,
corr `df0748bf`, 2 advisories — the guardian's caller enumeration answered with the nested-walk
census below, editquality's untested-recovery advisory answered with a test that drives a genuine
42703 through the latch) **and** `344`'s candidate 1 (`0f80f5ea1`, the `bugfix_307_terminal_write_contract`
lane's build on the owner's ruling — committed inside the build window; a completion now refuses
an item whose `retry_after` is in the future). Binary literal probe: the new CASE fragment and
`countingLadderStatement` present, absent-needle control 0.

**The canary pass** (`c192a2b2`, pool-web-tech.internal, same recipe as §9b, torn down after —
0 rows by key, no immune-sweep pollution; every timestamp UTC 2026-08-21):

| arm | observed | verdict |
|---|---|---|
| attempt 1 | 18:23:46 `triaged`/1/`retry_after`+30 m, claim cleared, error preserved — **`completed_at` NULL** | ladder ✓ **and 344's fix ✓** (yesterday's binary stamped it `complete` 2 s later) |
| guard flip | flipped to `wont_fix` at 18:26:23 while claimed; handler failed at 18:26:33; row **untouched** (`wont_fix`/1, `updated_at` unmoved), job pod logged `work item failure ladder: skipped — a deliberate status is already recorded, not overwriting` | decision guard ✓, demanded live |
| attempt 2 | 18:29:36 `triaged`/2/`retry_after`**+60 m**, `completed_at` NULL | scaling ✓, survives completion ✓ |
| attempt 3 | 18:31:15 **`failed`, 3 of 3, `retry_after` NULL, `completed_at` NULL** | honest terminal ✓ (§9b's 42P18 dead), completion refused ✓ |
| transient release | §9's two natural rows, 2026-08-20 | ✓ (unchanged) |

Natural corroboration: **0 false-green rows** (`complete` with `retry_after > completed_at`)
since the roll, and the §9b damage row `0c65f9fa` was re-driven past the defect by its own lane
(now `needs_human_review` at attempt 2) — the pre-fix damage list is empty.

**The guardian advisory's caller enumeration** [MEASURED, nested walk]: 13 live steps across 11
agents route through the two ladder writers — `fail_work_item`: build/diagnose/report-dispatch-loops,
site-work-orchestrator, component-template-fixer ×2, tool-improver, page-build-handler
(`mark_needs_review`, the 033 D2 override branch); `update_work_item_status(failed)`: the five
§2.2 agents. The column-absent liveness claim: `retry_after` exists (migration `505`, applied
08-20) and the latch flips only on a genuine 42703 — none observed; the fallback guards the
binary-before-migration window only, which has passed.

**Close basis, in the owner's split style.** The bar is fixed-AND-live. The rule from the
2026-08-21 morning ruling: close on fair-weather proof, with §5's outage-scale arm as a dated
standing watch rather than a hold. This case: all three filed defects (no-delay ladder, ladder-less
§2.2 writers, missing decision guard) plus the two defects the acceptance itself surfaced (344's
overwrite, §9b's 42P18) are fixed, live on v1.0.1322, and proven above — so the bar is met.

**STANDING WATCH (the §5 outage arm, not waived — converted):** on the next infrastructure
outage, expect burst-release log lines and **zero** terminal `failed` rows attributable to it.
**Reopen trigger:** any adapter outage leaving terminal `failed` rows attributable to it, or a
confirmed outage during which no burst-release fires. Check:
`SELECT count(*) FROM site_work_items WHERE status='failed' AND error LIKE '%<outage signature>%';`

**Residuals, each with an owner, none re-opening this file's defects:**
- `bugs_open/344` stays OPEN for ~~its **sweep SQL half**~~ **its council round only** —
  CORRECTED same evening by the owning lane: the sweep half is DEAD SQL (all three
  `claimed-item-timeout` arms carry `WHERE wi.status='claimed'`; a re-triaged row is `triaged`
  with claims cleared and cannot be re-claimed mid-cooldown — 0 such rows measured, 0
  sweep-attributable false greens, `341` §5c). This bullet's author transferred 317's
  mechanism by analogy without re-reading a predicate already in this session's transcript —
  WRONG_CALLS row filed. What 344 actually still owes: the lane's council round `2c21e214`
  (r1 REVISE — a "byte-identical" claim on the claim gate, since measured and test-pinned;
  r2 dispatched) and its close.
- `bugs_open/341` + migration `524` — ~~`_HOLD`~~ **RELEASED and APPLIED** the same evening
  (its condition was candidate-1-LIVE, deliberately not candidate-1-committed: the owner's
  deferred roll would otherwise have stamped cooldowns all day against an unguarded
  `mark_complete`). The sweep now stamps `retry_after` from `reaper_policies`.
- `033` D2 (`fail_work_item`'s `status_override` branch) — its own owner ruling.
- `update_work_item_status`'s parked/complete writes still increment `attempt_count` (§8) — a
  parked item's count is not a count of attempts.
