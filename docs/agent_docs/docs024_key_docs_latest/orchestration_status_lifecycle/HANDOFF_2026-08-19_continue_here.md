# HANDOFF 2026-08-19 — orchestration status lifecycle, cold start

> ## ✅ LANE CLOSED 2026-08-19 16:05Z — the round that was in flight came back **APPROVED**
>
> `466` round 2: **APPROVED**, 4 advisories, none high-severity (`Council-Reviewed: f0e95e58`).
> All four migrations live and approved; the Go half live on `v1.0.1315` across **all 27** chassis
> pods. **Nothing on this lane is outstanding.** Read `SUMMARY_2026-08-19_….md` for the read-out.
>
> The three items below under "decided, not yet built" are tracked **elsewhere by design** and are
> not this lane's work: **RFC_039** awaits an owner ruling, **`bugs_open/329`** is filed with no
> known incident, and `kafka-scheduler`'s `base/deployment.yaml` still declares `128Mi` (noted at the
> head of `bugs_closed/240`).

**Read this first, then `NOTES_…md` (missteps) and `RUNBOOK_…md` (commands).** The lane is **DONE**.

---

## 1. What this lane was, and where it got to

A job that stopped in certain internal states was never cleaned up by anything, ever — the oldest
found had sat **19 days** through several fleet restarts, each pinning two Kafka topics for good.
Four layers, each found by fixing the one above:

| layer | bug | fix | state |
|---|---|---|---|
| a `RUNNING` row is unreachable by every recovery path | `bugs_closed/294` | migration `463` | **LIVE + council APPROVED** (`860d87d9`) |
| the same gap one status over | `bugs_closed/310` | migration `464` | **LIVE + APPROVED** (`e973d2aa`) |
| the reaper's coverage was a *list*, so any unlisted status was immortal | — | migration `465` + `case StatusRunning` | **LIVE + APPROVED** (`1c212b15`) |
| "which statuses are terminal" was still literals in two `pre_query` texts | — | migration `466` (vocabulary table + FK) | **LIVE, round 2 IN FLIGHT** (`f0e95e58`) |

Also done: the dead `WorkflowMonitor` cluster **deleted** (APPROVED `25fa8173`, live on `v1.0.1314`,
proven by absence with a discriminating control pair); the dead pause vocabulary **deleted**;
`bugs_open/240` **closed** on symptom + mitigation.

**Fleet state at handoff:** stranded rows over 4h **0**; vocabulary **7 rows**, FK **enforced**; both
sweeps read the table; Go half **live on `v1.0.1315`** (both `case StatusRunning` markers PRESENT in
the running binary with controls, `orchestrator_state` still ABSENT).

## 2. THE ONE THING IN FLIGHT

**Migration `466`'s council round 2, correlation `f0e95e58-1361-442b-87a5-56b5870943b6`.**
Round 1 returned **REVISE**, gated by `guardian` at **HIGH**. Round 2 was submitted 2026-08-19 with
every objection answered and was at `review_reuse_agent` when this was written.

```sql
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='f0e95e58-1361-442b-87a5-56b5870943b6' AND kind='council_report' ORDER BY created_at;
```

⚠ **`466` is APPLIED AND LIVE while its round says REVISE.** Normal on this tree (review is after the
fact by design) but do not mistake it for unshipped. **If round 2 comes back APPROVED**, commit the
verdict record with `Council-Reviewed: f0e95e58-1361-442b-87a5-56b5870943b6` — but **read the verdict
first**; an unread `Council-Reviewed:` is the coverage report's dishonesty surface.
**If it comes back REVISE again**, the objections and their answers are already in `NOTES`.

**The HIGH objection, and why it does not materialise** — worth understanding before you touch
anything: I validated the vocabulary seed by grepping **Go** only, while SQL embedded in an
`agent_definitions` step is *data*, invisible to that grep. Audited since: only the reaper writes a
status (one literal, `FAILED`); **0** agent steps write the table; the 5 that mention it are all
reads (3 are LLM prompt prose, 2 of those inactive). Plus the FK has taken production traffic for
hours with **0** violations.

## 3. Decided by the owner, NOT yet built

1. **RFC_039 needs a ruling.** Raised and committed
   (`architecture_review/RFC_039_the_estate_now_has_two_patterns…`). It asks where the *rule* lives
   for choosing between the two vocabulary patterns. Recommendation is option 1 (codify the test —
   *if DB-resident SQL must read it, use the table; else the CHECK*). **Nothing is blocked on it.**
2. **`bugs_open/329`** — both takeover arms decide by a 5-minute clock, not a lock, so two pods can
   run one step. Filed with candidates ranked and a both-directions verify recipe. `TakeOverOrchestration`
   already exists as the guarded CAS. **Not a fire** — the heuristic has run for months.
3. **A SUMMARY doc** — deliberately not written yet. The rule is that summaries are milestones, not a
   diary; the moment is once round 2 rules. **This is the natural next act if you are closing the lane.**
4. **`base/deployment.yaml` still declares `128Mi`** for `kafka-scheduler` while production overrides
   to `256Mi` — noted at the head of `bugs_closed/240`. Deploying that service anywhere else
   reproduces the original OOM condition.

## 4. What would make this lane closable

Round 2 APPROVED + verdict recorded + a SUMMARY written. Items 1, 2 and 4 above are *tracked
elsewhere by design* (an RFC, a bug file, a closed bug's banner) and do not belong to this lane.

## 5. Traps this lane hit — do not re-learn them

- **A mandated pre-flight check can have EXPIRED.** `294` said to re-run a census before applying; it
  returned 0 in every band and could not have come out otherwise. Ask what today's disconfirming
  result looks like *before* treating any mandated check as a licence.
- **A structural licence does not transfer between sibling states.** `RUNNING` is transient *by
  construction*; `INITIALIZED` *waits on a queue*, so it had to be measured (max 6.31s over 5,736
  rows). A parallel lane reused the argument and was ~1000× out.
- **Read the caller graph, not the function body.** Got this wrong **three times in two days** —
  `monitoring.go`, `TimeoutMonitor`, and my own correction about mounted endpoints.
- **An md5 check can pass on a truncated file**, and an extractor can report the md5 of a bare
  newline for both payloads. Assert *structure* as well as content.
- **Seed a vocabulary from what the CODE CAN WRITE, never from `SELECT DISTINCT`.** Only 5 statuses
  existed when `466` was built; seeding from the table would have rejected every new orchestration.

## 6. Commands

All in `RUNBOOK_orchestration_status_lifecycle.md` — reading the live `pre_query`, the four traps in
rewriting one, the two house guards and proving them by mutation, applying without `--apply`,
inducing the reaper with a negative control, adding a status, pod-verifying a Go change, and the
in-pod Kafka topic count that silently returns 0 if you pipe it.
