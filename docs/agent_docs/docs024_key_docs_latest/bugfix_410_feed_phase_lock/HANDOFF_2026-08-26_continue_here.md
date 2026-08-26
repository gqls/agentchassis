# HANDOFF — bugs_open/410 (the PHASE-LOCK one) — both halves LIVE, one acceptance test outstanding

**Written 2026-08-26 ~21:00Z (DB clock; see the clock warning below) by the
`bugfix_410_feed_phase_lock` lane. This is the cold-start doc — read this first.**

⚠ **410 NAMES TWO UNRELATED BUGS.** This lane owns
`bugs_open/410_HANDOFF_2026-08-26_next_fetch_at_stamped_at_fetch_time_phase_locks_every_six_hour_news_site_to_a_twelve_hour_cadence.md`
(the news-feed cadence phase lock). The OTHER 410 — `..._three_seams_fail_toward_the_quiet_default...`
— is a different case with its own lane (`bugfix_410_silent_scan_loss`) and its own session.
**Resolve by slug; never by number.** (CLAUDE.md's ambiguous-number list now lists 410.)

---

## 1. What the bug is, in one paragraph

Both writers of `content_sources.next_fetch_at` stamp `NOW() + fetch_interval` at **fetch**
time, while the trigger (`scheduled_tasks.content-feed-refresh`, 21600 s) fires on a fixed
grid. Dispatch is sequential, so a fetch lands 10 s–9 min after its trigger, whereas the grid
drifts only seconds. A source whose `fetch_interval` **equals** the cadence (6 h — the column
default, and every source on 12 of 12 news sites) therefore comes due *seconds after* the next
pass fires, is skipped, and is served 6 h later: **a 12 h cadence under a 6 h label, with every
run reporting COMPLETED.** 10 of 12 news sites, since the trigger was armed.

## 2. What shipped, and where it is

**The fix is candidate 1: a due look-ahead of HALF the trigger cadence, in every layer that
asks "is this source due?"** — serve on the **nearest** grid tick, not the first tick strictly
after. Cadence is read **live** from `scheduled_tasks`; `COALESCE` falls back to `interval
'3 hours'` (half of today's 21600 s) so a renamed task degrades to the designed value, never to
the bare `NOW()` that caused the lock.

| layer | where | state |
|---|---|---|
| shared predicate | `platform/orchestration/actions/feed_due_lookahead.go` (`feedDueLookaheadSQL`, `feedSourceDuePredicate`) | **LIVE** |
| source-level (live path) | `dispatch_feed_sources_action.go` due query | **LIVE** |
| source-level (dormant) | `feed_actions.go` `LoadDueSourcesAction` — zero live callers, fixed so a future caller inherits it | **LIVE** |
| site-level admission | `docs/agent_docs/sql_for_agents/653_content_feed_due_lookahead_HOLD.sql` → `content-feed-trigger.find_news_sites` | **APPLIED 2026-08-26 ~20:52Z** |

Commits: `201236b2a` (fix), `4da33f40e` (gofmt), `8ac0d0558` (ratchet line), `f3ca6a9e5`
(316 CONTRIB), `86aea3804` (council verdict + objection discharge), `b34c24f4c` (WRONG_CALLS).
Council **APPROVED round 1**, corr `04c657d2-cbee-4528-b124-b53a747d2e96`.

## 3. Deployment state — VERIFIED AT THE ARTEFACT, not inferred

- **Go half live on `v1.0.1345`, both replicas.** The `build provenance` startup line had
  already scrolled out of the retained log (an EMPTY grep there means "not in range", **not**
  "unstamped"), so this was proven by binary probe **with both controls in the same breath**:
  ```bash
  kubectl -n ai-persona-system exec <pod> -- grep -acF 'make_interval(secs => interval_seconds / 2.0)' /proc/1/exe   # 2  (the two Go readers)
  kubectl -n ai-persona-system exec <pod> -- grep -acF 'interval_seconds / 7.0' /proc/1/exe                          # 0  (negative control)
  kubectl -n ai-persona-system exec <pod> -- grep -acF 'DispatchFeedSourcesAction: dispatched ingester' /proc/1/exe  # 1  (positive control)
  ```
  This is a **capability** probe, not a commit probe — it asks whether the behaviour is in the
  binary, which is the question that actually matters.
- **Config half applied and re-read independently** (not trusting the migration's own
  in-transaction post-check): look-ahead **present**, 554 due-ordering + 556 caps tail
  **intact**, bare `NOW()` arm **gone**. Both of 653's guards passed, which also settles the
  council's low-severity objection empirically — the live query **was** 556's post-image.
  A pre-change snapshot exists: `snapshot_agent` id `51dd1c59-69e6-4625-baf6-203c35052f18`.

## 4. THE ONE OUTSTANDING TEST — and it is genuinely disconfirmable

The evening pass fired **20:46:45Z** and was the **last bare-NOW()-admission pass**. Dispatch
is sequential (~2.5 min/site here), so it was still running when this handoff was written —
**do not use the partial list below as the test set; re-read it.** Confirmed dispatched by
20:55Z: webdesign.co.uk 20:47:02, ai-agent-orchestration.com 20:48:26, fundamentallyai.com
20:50:31, robot-hands.com 20:53:03 — each due since ~14:47, i.e. each *skipped by the 14:46:32
pass*, which is the phase lock itself, observed once more.

Those sites' sources are stamped **fetch time + 6 h ≈ 02:47Z onwards**. The next trigger fires
**≈ 02:46:5xZ** — seconds *before* them, which is the whole defect.

> **THE ACCEPTANCE TEST: every site dispatched in tonight's 20:47Z pass must be dispatched
> AGAIN in the ~02:46Z pass.** Under the old behaviour each misses by seconds and waits until
> ~08:46Z. This is the fix working, or not, on its first real grid tick.

```sql
-- the test set (read it, don't assume the four above are all of it)
SELECT s.domain, to_char(o.created_at,'HH24:MI:SS') AS dispatched_2047
FROM orchestration_states o JOIN sites s ON s.id=o.site_id
WHERE o.owner_agent_type='content-feed-orchestrator'
  AND o.created_at BETWEEN timestamptz '2026-08-26 20:40:00+00' AND timestamptz '2026-08-26 21:30:00+00'
ORDER BY o.created_at;

-- the result: same domains must appear here
SELECT s.domain, to_char(o.created_at,'DD HH24:MI:SS')
FROM orchestration_states o JOIN sites s ON s.id=o.site_id
WHERE o.owner_agent_type='content-feed-orchestrator'
  AND o.created_at > timestamptz '2026-08-27 02:40:00+00'
ORDER BY o.created_at;
```
⚠ `orchestration_states` is pruned to roughly **2 days**, so run the first query before it ages
out, or take the test set from `content_sources.last_fetched_at` (≈20:47–21:0xZ) instead.
**Disconfirming result:** tonight's sites absent from the 02:46Z pass ⇒ the look-ahead is not
reaching the decision; re-check the live config query and the binary probe before anything else.

⚠ **`bugs_open/410` §5's own prediction (d) is NOT discriminating** and must not be reported as
if it were: idea.uk has been due since 20:47:24 and will be dispatched at 02:46Z **whatever**
the predicate says. Prediction **(c) CONFIRMED**: the 20:46:45Z trigger fired **39 s before**
idea.uk's earliest due stamp (20:47:24) and idea.uk was **not** dispatched — the site-admission
phase lock, observed live, one last time before 653 closed it.

## 5. What the 090 diagnosis loop actually returned — state this honestly

**`UNVERIFIABLE` — "Diagnosis NOT confirmed (stopped: scope-not-narrowing)"**
(item complete, dispatch corr `15d56c13-2081-431a-ad70-9516c5fcfbc7`; two evidence bundles
written, **no** iteration_note, **no** verdict artifact, nothing in `doc_notes`).

**This is not a refutation and must not be reported as one — nor as a confirmation.** The loop
did not reach a conclusion. The mechanism rests instead on the declared 2026-07-31 substitute:
first-hand verification (live config read, both Go stamp arms read at their lines, the 48 h
census with sub-6h control sites, and a **prospective** prediction recorded before the pass that
tested it, (a)/(b)/(c) all since confirmed) plus council approval at round 1.

**Probable cause of the stall, marked as inference `[INFERRED]`:** the symptom I submitted was a
fully-formed conclusion — mechanism, cause, consequence and blast radius in one paragraph —
against CLAUDE.md's own authoring guidance ("state the MECHANISM, then POINT at the tables/
symbols… no downstream-consequence clauses"). A loop whose job is to narrow scope has nothing
to narrow when handed the answer. Logged in `WRONG_CALLS.md`.

## 6. Residuals — stated so they do not become this lane's own quiet default

1. **Go↔config predicate parity is pinned at commit time but NOT enforced against live config
   on a schedule.** A future migration could rewrite `find_news_sites` without the look-ahead
   and nothing would fire. Candidate: a `cmd/config-key-audit` mode asserting that any config
   query with a `next_fetch_at` due arm carries it. **Not built.**
2. **The `interval '3 hours'` fallback goes stale in BOTH layers if the cadence changes.**
   653's guard catches it only at apply time. Landmine filed (footprint `feed_due_lookahead.go`).
3. **`LoadDueSourcesAction` remains callerless** — fixed and guarded, but dead code.
4. **The provocation-feed twin (`provocation-feed-refresh`, also 21600 s) was NOT audited for
   this class.** Its publisher may or may not stamp fetch-relative due times. Genuinely unknown.
5. **Cost/capacity, for the owner, not for a session to decide:** the designed cadence roughly
   **doubles** feed ingestion (~+100 ingester runs/day), and ~12 sites now contend for caps of
   10 every pass, so cap-hit warnings (LCO-009, `--capped-schedule-ordering`) become routine and
   are **expected demand, not a regression** — 554's due_at ordering rotates the overflow
   fairly. Raising caps to 12/12 is the next capacity step **if the owner wants it**.
6. **The bug file stays OPEN** until the 02:46Z acceptance test passes and the 48 h census in
   `bugs_open/410` §7 shows four run-hours/day per 6h-only site.

## 7. Traps this lane hit — read before repeating the measurements

- ⚠ **THIS WORKSTATION'S CLOCK IS ~1 HOUR AHEAD OF THE CLUSTER'S.** Measured 2026-08-26: local
  `date -u` read 20:53 while `clients_db now()` read 19:54:35 in the same minute. A watcher
  armed off local time fired an hour early and I nearly recorded "the trigger never fired".
  **Take the clock from the DB for anything keyed to DB-timestamped events.** (WRONG_CALLS.)
- ⚠ **An absence census over workflow steps must search the whole config text**, not
  `jsonb_each` over top-level steps — the latter cannot see a nested caller inside a loop step.
  This was the council's medium objection and it was right to raise it.
- ⚠ **`\d <table>` before the first SELECT.** `scheduled_tasks`'s column is `interval_seconds`;
  two guessed names cost two round-trips (WRONG_CALLS).
- ⚠ **Excluded from any cadence census:** remortgagecalculator.uk's off-cadence 13:43Z run on
  2026-08-26 (origin unestablished).

## 8. Lane docs and pointers

`docs/agent_docs/docs024_key_docs_latest/bugfix_410_feed_phase_lock/` — PLAN (the decision
record and why candidates 2/3 were rejected), RUNBOOK (every command, with its gotcha),
NOTES (append-only technical log incl. all mutations), README_where_we_are (owner-facing prose).
Related: `bugs_open/410` (phase-lock slug) · `bugs_open/316` + its lane (CONTRIB filed there —
their README:138 "fully served" claim was refuted by this second mechanism) ·
`LANDMINES.md` (two-layer due predicate) · migration `653` + its `_ROLLBACK`.
