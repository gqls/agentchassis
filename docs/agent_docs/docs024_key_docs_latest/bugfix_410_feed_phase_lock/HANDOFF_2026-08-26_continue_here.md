# HANDOFF — 410 (the PHASE-LOCK one) — ✅ CLOSED 2026-09-02. NOTHING IS OUTSTANDING.

> ## ⚠ STOP — THIS DOC IS SUPERSEDED. The cold-start doc is now:
> ## `docs/agent_docs/docs024_key_docs_latest/bugfix_410_feed_phase_lock/HANDOFF_2026-09-02_continue_here.md`
>
> (added 2026-09-02) — go there for current state, the two owner decisions, the one optional
> check, and the traps. Everything below is kept only as the record of what was believed on the
> 26th. It is **not** instructions.
>
> **The bug is CLOSED and the file has MOVED: `bugs_closed/410_HANDOFF_2026-08-26_next_fetch_at_…`.**
> Everything under "§4 THE ONE OUTSTANDING TEST" is **DONE and its test set is EXPIRED**. Do not
> run it. Current state, in one place:
> **`SUMMARY_2026-09-02_410_feed_phase_lock.md`** (milestone read-out) · `NOTES` (queries, tail
> section 2026-09-02) · `RUNBOOK` (the test that actually works).
>
> **Verdict: the fix is confirmed working on live traffic.** Pass at 08:58:27Z on 2026-09-02
> served three 6h-only sites — remortgagecalculator.uk, **idea.uk** (this bug's own site) and
> vetcomparison.uk — **2 m 15 s, 4 m 58 s and 10 m 57 s before their earliest possible due time**.
> Impossible under the pre-fix predicate. Both halves re-probed live (chassis **v1.0.1352**).
>
> **⚠ §4's test below was UNSATISFIABLE AS WRITTEN — do not copy its shape.** It demanded that
> *every* site from one pass reappear in the next, while §6 residual 5 of this same document
> predicted the fix would push demand past the `LIMIT 10` cap. Measured: **4 of 8** discriminating
> sites were correctly capped out of the pass that proved the fix works. `WRONG_CALLS.md` + a
> LANDMINE ("a fix that relieves a bottleneck raises demand on the NEXT constraint"). §4's
> hardcoded window is also dead: **the trigger drifted ~:46 → ~:57 in six days.**
>
> **What is genuinely left, and it is NOT this bug's:** 6h-only sites now run at **~9 h**, not the
> designed 6 h — 14 eligible sites against a cap of 10. That is `bugs_open/316` (CONTRIB filed
> 2026-09-02 with the numbers) and an owner spend decision. Residual 4 below (the provocation
> twin) is **RESOLVED: not affected**. Residual 6's "keep open until…" is **discharged**.

**Written 2026-08-26 ~21:00Z (all times in this lane are DB time — `SELECT now()`) by the
`bugfix_410_feed_phase_lock` lane. Superseded 2026-09-02 by the block above — the body below is
kept as the record of what was believed on the 26th, not as instructions.**

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

## 4. ~~THE ONE OUTSTANDING TEST~~ — ✅ DONE 2026-09-02, but **NOT with the test below; that test was broken**

> **CORRECTED 2026-09-02.** The test in this section was **unsatisfiable by a working fix**, and
> the reason is in §6 residual 5 of this same document: the look-ahead makes ~all 6h-only sites due
> at every pass, demand exceeds the query's `LIMIT 10`, and the surplus is *correctly* displaced —
> which a pass-membership test scores as a failure. Measured 2026-09-02: of 8 discriminating sites,
> **3 passed, 1 was vacuous, 4 were capped out**. The pinned test set below is also **expired**
> (`orchestration_states` prunes at ~2 days) and its `02:40:00` window is dead (**trigger drift
> ~:46 → ~:57**). What replaced it — a lower bound a cap cannot confound — is in the RUNBOOK under
> "THE ACCEPTANCE TEST". `WRONG_CALLS.md` + LANDMINE, same date. **Everything below is the 26th's
> record, not a task.**

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

**THE TEST SET, RECORDED HERE** so it does not have to be re-derived from `orchestration_states`
(pruned at ~2 days). Captured 21:02Z, with each site's re-stamped `next_fetch_at` — note every
one falls **after** the expected ~02:46:5xZ trigger, which is the defect, and **within** the 3 h
look-ahead, which is the fix:

| site | dispatched | re-stamped due | discriminating? |
|---|---|---|---|
| webdesign.co.uk | 20:47:02 | 02:47:31–02:47:54 | **YES** (6h-only) |
| ai-agent-orchestration.com | 20:48:26 | 02:48:48–02:49:05 | **YES** |
| fundamentallyai.com | 20:50:31 | 02:50:54–02:51:14 | **YES** |
| robot-hands.com | 20:53:03 | 02:53:26–02:53:59 | **YES** |
| gaswholesalers.com | 20:59:08 | 02:59:35–02:59:42 | **YES** |
| mortgagecalculator.co.uk | 21:01:36 | ≈03:01–03:02 | **YES** |
| relojistas.com | 20:55:20 | 23:55:46–02:55:43 | no — 3 h source due 23:55, served either way |
| dartsonline.com | 20:57:12 | 00:57:55–02:58:06 | no — 4 h source due 00:57, served either way |

⚠ The pass was still dispatching at 21:02Z; sites may have been added after this snapshot —
re-read if you need completeness, but **the six discriminating rows above are sufficient**.

⚠ **Use only the six.** relojistas and dartsonline carry sub-6h sources that come due *before*
the 02:46 trigger, so they are served under either predicate — they are the same vacuous shape
as prediction (d), and reporting them as evidence would be the mistake this lane already made
once.

```sql
-- the result: the six discriminating domains must all appear here
SELECT s.domain, to_char(o.created_at,'DD HH24:MI:SS')
FROM orchestration_states o JOIN sites s ON s.id=o.site_id
WHERE o.owner_agent_type='content-feed-orchestrator'
  AND o.created_at > timestamptz '2026-08-27 02:40:00+00'
ORDER BY o.created_at;
```
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

- ⚠ **`date -u -d '<naive timestamp>'` PARSES THE INPUT IN LOCAL TIME — `-u` affects OUTPUT
  ONLY.** On this BST (+0100) box a watcher armed for "20:53:00 UTC" actually fired at 19:53Z,
  an hour early, and I misread my own off-by-one as a cluster clock problem. **Append an
  explicit zone to the string:** `date -d '2026-08-26 20:53:00 UTC' +%s`. Proof (identical
  epochs with and without `-u`; different only with the explicit zone):
  ```
  date -ud '2026-08-26 20:53:00' +%s -> 1787773980 = 19:53:00 UTC   ← -u did nothing
  date -d  '2026-08-26 20:53:00 UTC' +%s -> 1787777580 = 20:53:00 UTC
  ```
  > **CORRECTION 2026-08-26:** an earlier version of this bullet claimed *"this workstation's
  > clock is ~1 hour ahead of the cluster's"*. **That was FALSE and is retracted.** All three
  > clocks agree within 4 s (local `date -u` 20:57:37 · postgres container OS 20:57:39 ·
  > `now()` 20:57:41, TimeZone UTC). Caught when a peer lane suggested promoting the false
  > finding to a LANDMINE and filing it forced me to ask *which* clock was wrong. Full account,
  > including the `[MEASURED]` marker I put on an inference: `WRONG_CALLS.md`.
- ⚠ Still true and unrelated to the above: **both ends of any age/window arithmetic must come
  from one clock.** Ask the DB for the age (`SELECT now() - created_at`) rather than
  subtracting a shell reading from a DB timestamp.
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
