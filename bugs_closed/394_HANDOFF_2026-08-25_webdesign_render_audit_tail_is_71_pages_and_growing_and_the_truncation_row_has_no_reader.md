# 394 — webdesign.co.uk's render audit covers 60 of 131 pages, the tail GROWS every week, and the truncation row `bugs_closed/242` built has no reader

> ## CLOSED 2026-09-03 — **FIXED AND LIVE, both commissioned halves, each proven at the artefact.**
>
> Councils `f67593f5` (cursor) and `f49da30d` (CronJob) both APPROVED.
> Lane: `docs/agent_docs/docs024_key_docs_latest/bugfix_394_render_audit_rotation_cursor/`
> (read `HANDOFF_2026-09-03_closed.md` there for the full account and the open questions).
>
> ### The two things the owner commissioned on 2026-08-25, and where each ended up
>
> **1. The rotation.** `request_render_audit` no longer takes the same prefix every run. It carries
> a keyset coverage cursor (opt-in `rotate_coverage`, migration **660**), and the window is a UNION:
> the pages carrying an open `contrast_failure` ride EVERY run, so the 3-day repair-grading latency
> migration 469 was an owner instruction to buy is preserved.
>
> **This file's own acceptance test passes.** `[MEASURED 2026-09-02]` over three consecutive
> SCHEDULED runs on webdesign.co.uk the union of `audited_paths` is **151 distinct pages of 151
> live — ZERO missed**, graded against the site rather than against itself. The third run cleared
> the cursor (`cursor_cleared=true`, final window 37 pages, last page `tool-llm-cost-calculator` at
> `nav_order` 201). The second arm is met too: in cursor mode the message no longer claims "the
> SAME pages every run". A second site, `loanandmortgagecalculator.co.uk`, cycled to completion
> unaided.
>
> **What that replaced:** the same first 60 pages for ever, with 91 never audited — including all
> 45 `tool-*-guide` pages, unreachable at ANY cap below 98.
>
> **2. The reader.** `cmd/config-key-audit --render-truncation`, shipped as CronJob
> `render-truncation-check` (07:50 UTC), registry flipped to `consumed`. `[MEASURED 2026-09-03]` it
> **fired on its own schedule** at `07:50:00Z` and wrote its durable row at `07:50:14Z`: *22 rows
> across 4 sites and 2 callers, 0 findings, 1 dormant group named.* One row per run, clean
> included, so an ABSENT row means the job did not run and can never read as "nothing is wrong".
>
> ### Three things a later reader should not have to rediscover
>
> - **The tail was a CLASS, not a count** — which is why "raise the cap" was rejected on a
>   measurement rather than a preference.
> - **Never parse a `contrast_failure` `item_key`.** A selector may contain `#` and so may a page
>   URL (`idea.uk` has `/tools.html` AND `/tools.html#audience-check` active). Match forward with
>   `workItemKey(...)` + `HasPrefix`. The council caught this as a LIVE defect in round 1.
> - **The cursor key is the RUNNING agent, not the dispatcher.** Keyed on `Sender.AgentType` it
>   silently kept a separate cursor per dispatch path. No unit test could see it — the fixture set
>   both identities to one literal.
>
> ### Deliberately NOT done, and each is a question rather than a defect
>
> `design-critique-agent` still takes a prefix (manual sampler; rotate-vs-curate is a product
> decision, not a bug) · webdesign's NEW-defect detection latency moved 3d → ~1 cycle in exchange
> for 91 pages going from never to one cycle · the dormancy window is 14 days, a judgement · and
> `page_names` is declared in the action's spec and read by nothing (`bugs_open/452`).
>
> ---
>
> ## STATUS 2026-09-02 (superseded, kept for the trail) — CURSOR HALF FIXED, LIVE AND ACCEPTED.
>
> **The acceptance test in §2 PASSES.** `[MEASURED 2026-09-02]` over three consecutive SCHEDULED
> rotation runs (08-27, 08-30, 09-02), the union of `audited_paths` on webdesign.co.uk is
> **151 distinct pages of 151 live — ZERO missed**, graded against the site rather than against
> itself. The third run cleared the cursor (`cursor_cleared = true`, final window 37 pages, last
> page `tool-llm-cost-calculator` at `nav_order` 201). The second acceptance arm is met too: in
> cursor mode the message no longer claims "the SAME pages every run".
>
> For contrast, what this replaced: the same first 60 pages for ever, with 91 never audited —
> including all 45 `tool-*-guide` pages, unreachable at any cap below 98.
>
> Also proven in production, unattended: the mode split (`design-critique-agent` writes `prefix`,
> `render-audit-agent` writes `cursor`), and a SECOND site cycling to completion
> (`loanandmortgagecalculator.co.uk`, 61 pages, a 2-page final window on 08-30).
>
> ### WHY IT IS STILL OPEN
>
> The owner commissioned TWO things (decision 4, 2026-08-25) and only one is driven.
> `cmd/config-key-audit --render-truncation` is built, tested, mutation-proven on four arms and
> registered in `finding_code_registry.json` as the `consumed` reader — but **no CronJob runs it**
> (`kubectl get cronjob` has no `render-truncation-check`). Until it is scheduled the registry's
> `consumed` claim is true about the code and false about the estate — the state `DBG-075` exists
> to prevent. `/bugs_closed/`'s bar is "fixed AND live"; half of this is not live.
>
> **Next step is one kustomize service**, cloned from `ungraded-completions-check`.
> Full state: `docs/agent_docs/docs024_key_docs_latest/bugfix_394_render_audit_rotation_cursor/HANDOFF_2026-09-02_continue_here.md`
>
> ---
>
> ## STATUS 2026-08-26 (superseded, kept for the trail) — OWNED, FIX BUILT AND COMMITTED, NOT YET LIVE.
>
> Lane: `docs/agent_docs/docs024_key_docs_latest/bugfix_394_render_audit_rotation_cursor/`.
> Council `f67593f5-90cb-4a35-9cc0-926254645192` — round 1 REVISE (acted on), round 2 submitted.
>
> **INERT until BOTH: an image carrying the code rolls, AND
> `sql_for_agents/649_render_audit_coverage_cursor_HOLD.sql` is applied BY HAND afterwards.**
> Do not read the commits as a fix in production. The `_HOLD` suffix is there because the
> runner hard-fails on a capability the binary lacks.
>
> ### What the measurement changed about this file
>
> `[MEASURED 2026-08-26]` webdesign.co.uk is **146 live pages, 60 audited, 86 never** — the
> file's 131/71 was two days old. And the tail is a **CLASS, not a count**: `nav_order` bands are
> 0..90 (6 nav pages), 100 (94 tools, alphabetical), 200 (**48 `tool-*-guide`**), 201 (1), so the
> cap cuts between `tool-head-architect` and `tool-html-minifier` and **every remaining guide page
> is unreachable at any cap below 98**. That kills §2 candidate 3 on a measurement.
>
> **§1's `[UNEXPLAINED]` "5 of 26" row is RESOLVED.** Its own `agent_error_log.context` reads
> `{"max_pages": 5, "pages_total": 26, "pages_audited": 5}` under `agent_type =
> render-audit-agent`, step `audit` — the standing agent with a **per-dispatch override**, not a
> config regression. The originating `orchestration_states` row has since aged out, so where the
> override came from is unrecoverable; recorded as such. **The conclusion that matters: `max_pages`
> is per-dispatch, so no design may assume "the cap is 60".**
>
> **There is a SECOND caller the file does not name.** `design-critique-agent` (seeded 2026-08-25,
> `sql_for_agents/645`) also runs `request_render_audit`, at `max_pages` **8**, and truncated twice
> on its first day. At cap 8, **25 sites** truncate, not one. It is deliberately left in prefix mode
> — it is a manual sampler with no cadence and its 8 pages are plausibly the intended sample — and
> it is acknowledged by name in the reader's acks file.
>
> ### What shipped (committed, not live)
>
> **Candidate 1 (cursor) AND candidate 2 (reader), both** — the reader is not discharged by the
> cursor, because the cursor changes what the row MEANS rather than removing it.
>
> - Keyset cursor in the audit's own ordering, opt-in via `rotate_coverage` (default **false**),
>   per `(site_id, agent_type)` in new table `render_audit_page_cursor`, advanced at DISPATCH and
>   written AFTER a successful produce.
> - **The window is a UNION, and that is the decision this turned on.** A plain cursor takes
>   webdesign's per-page re-measurement latency to ~9 days; migration `469` is an **owner
>   instruction of 2026-08-18** that cut the window 7d→3d precisely because the render audit is the
>   only thing that GRADES a contrast repair. A plain cursor would exceed the condition the owner
>   ordered removed. The pages carrying an open `contrast_failure` (**3 of 146** on webdesign,
>   bounded at `max_pages/2`) therefore ride EVERY run.
> - Mode-split message and nine context keys, all present on every cursor-mode row including zeros.
> - `cmd/config-key-audit/rendertruncation.go` — three mutation-proven alarm arms; registry entry
>   flipped to `consumed` (DBG-075).
>
> ### The acceptance arm this file asks for is NOT runnable until 649 applies
>
> §2's acceptance ("the union of audited pages reaches all 131, by the audit's own durable page
> list") could not be run at all: **there was no durable per-page list** — every persisted use of
> `pages_audited` was a COUNT. `audited_paths` now carries it. The union query is in the lane
> RUNBOOK.
>
> ### ⚠ Still owed, and none of it is optional
>
> 1. Roll an image carrying the code, then apply **649** by hand.
> 2. **Re-apply the `optional-key-budget-check` kustomize overlay** — `rotate_coverage` took
>    `request_render_audit` to 7 optional keys and the CronJob's literal moved with it in the
>    repo; the cluster keeps the old literal until the overlay is applied.
> 3. Run the union acceptance over one full cycle on webdesign (~9 days at a 3-day cadence).
>
> ### One open question for the owner
>
> Head-page DETECTION latency on webdesign goes from 3 days to one cycle (~9 days) in exchange for
> the 86-page tail going from **never** to one cycle. If that is the wrong trade, `max_pages` is
> now a pure latency dial with no coverage cliff behind it.

**Filed** 2026-08-25 by the `bugs_open/358` lane, on the owner's ruling of 2026-08-25 (decision 4:
commission a reader). Lineage: `bugs_closed/242` (*"a capped render audit is indistinguishable from
a complete one"*) — **CLOSED correctly**: its fix made truncation loud (`pages_total`/`truncated`
stamped into the durable result, plus an `agent_error_log RENDER_AUDIT_TRUNCATED` row before
dispatch), and migration `392_render_audit_rotation_max_pages_60.sql` (applied) raised the cap
25→60 as a stated MITIGATION. This file is the next rung: **the loud signal exists and nothing
reads it, and the mitigation has been outgrown.**

## 1. Evidence — the writer says it plainly, weekly, to nobody

```sql
SELECT occurred_at::date, left(error_message,120) FROM agent_error_log
 WHERE error_code='RENDER_AUDIT_TRUNCATED' ORDER BY occurred_at;
```

`[MEASURED 2026-08-25]`:

| date | message |
|---|---|
| 08-11 | `render audit truncated by max_pages: 5 of 26 live pages audited for loancalculator.co.uk …` |
| 08-18 | `… 60 of 109 live pages audited for webdesign.co.uk — the unaudited tail is the SAME pages every run` |
| 08-21 | `… 60 of 125 … webdesign.co.uk …` |
| 08-24 | `… 60 of 131 … webdesign.co.uk …` |

Two live facts:

1. **webdesign.co.uk has outgrown the 60-page mitigation and is diverging**: 109 → 125 → 131 live
   pages in six days, tail now **71 pages that are never audited — and the writer's own message
   says the tail is the SAME pages every run.** Whatever class of defect the render audit exists to
   catch is structurally invisible on more than half of the fleet's largest site.
2. **The 08-11 loancalculator row says `5 of 26`** — a cap of 5, not 25/60, on a site the 392
   migration's own header shows at 25-of-27 previously. `[UNEXPLAINED]` — a per-call override, or a
   config regression; whoever takes this should read that call's config before assuming the cap is
   uniform.

## 2. What is asked for

Commissioned (owner ruling 2026-08-25): **a reader for `RENDER_AUDIT_TRUNCATED`.** Candidates,
ordered by what closes the door:

1. **Make the rotation actually rotate**: persist a per-site cursor so the next run starts where the
   cap cut off. The writer's message says the tail is the same pages every run — a cursor makes the
   cap cost latency instead of coverage, and it retires the signal's cause rather than reporting it.
   (242's fix candidates discussed this; re-read that file's §4 before designing.)
2. **Read the row**: a daily-check-family consumer that alarms on any site truncated in the last N
   runs, so a site outgrowing the cap is a finding rather than folklore.
3. **Raise the cap again** — the 392 shape. Weakest: webdesign grew 22 pages in six days; a constant
   will always be outgrown.

**Acceptance**: for candidate 1 — over consecutive rotation runs on webdesign.co.uk, the union of
audited pages reaches all 131 (verify by the audit's own durable page list, not the status), and
the `RENDER_AUDIT_TRUNCATED` message changes meaning (tail no longer "the SAME pages"). For 2 —
synthetic truncation row → red; mutation-proved both ways. Registry follow-up: flip the code to
`consumed` with `reader`/`reader_sink` in the shipping commit (`DBG-075`).

## 3. Traps

- **Do not re-file 242** — the visibility half is DONE and live; this is the consumption half.
- The `5 of 26` row is `[UNEXPLAINED]`; resolve it by reading the dispatching config, not by
  assuming the cap is 60 everywhere.
- Rows expire (365d declared; 14d once a consumer resolves them) — extract before resolving.

---

**2026-09-02, contributed by the release-unblock session (not this lane):** the reader's image
could not build at the 2026-09-02 release — `.dockerignore` ignores `docs/`, and
`render_truncation_acks.json` had no `!` un-ignore line (its four elder siblings each do), so
the dockerfile's COPY of the acks file failed with "not found" and killed `make release` at
`build-render-truncation-check`. Fixed in `ebf27c603` (one line in `.dockerignore`);
`make build-render-truncation-check` then passed, and the fix is an ancestor of chassis stamp
`0d2feee2ff61` (the 21:00Z roll). **For whoever verifies this check's first run:** the image
only became buildable at `ebf27c603` — a "built" record from earlier the same day (`6dc288aaf`,
which is where the dockerfile and its COPY line were born) predates a buildable image. The
CronJob's live existence and first `doc_notes` row were NOT verifiable by the fixing session
(kubeconfig token expired); that verification stays with this lane. Trap distilled:
LANDMINES "A new ack-shipping check's image fails at RELEASE time".
