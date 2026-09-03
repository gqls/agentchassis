# NOTES — bugfix_410_feed_phase_lock (append-only, newest at bottom)

## 2026-08-26 ~17:30–18:00Z — pickup, verification at HEAD, fix built

- Ambiguity resolved first: TWO bugs named 410. Messaged the other `bugs_open/410` session
  (scan-loss lane, `[0bb588]`) — no overlap; messaged `idea.uk [e14ec9]` (the filer) — handoff
  confirmed, three refinements received (recorded in PLAN: (c) refutation criterion needs the
  trigger's actual fire time; I own recording (c)/(d); roll-before-apply sequencing; plus:
  316 has no active owner, and remortgagecalculator.uk's 13:43Z run must be excluded from any
  census).
- Queue check: no open `site_work_items` row covers the cadence defect → fired `090`
  17:31Z, RUN_CORRELATION_ID `15d56c13-2081-431a-ad70-9516c5fcfbc7`. [verdict pending]
- **Bug still valid at HEAD, verified first-hand:** both stamp arms unchanged
  (`dispatch_feed_sources_action.go` optimistic stamp; `feed_actions.go`
  `UpdateSourceTimestampsAction` success+failure arms); live `find_news_sites` query read
  from `agent_definitions` — byte-identical to migration 556's post-image, bare
  `cs.next_fetch_at <= NOW()`; `scheduled_tasks.content-feed-refresh` = 21600 s, enabled,
  last fired 14:46:30Z.
- **CORRECTION to the bug file's reader census:** live workflow uses `dispatch_feed_sources`
  (its own due query) — `LoadDueSourcesAction` (`feed_actions.go`, the file's cited "selector
  the orchestrator itself uses") has **zero live workflow callers** (`jsonb_each` over all
  active agent_definitions steps: only `feed-ingester.update_timestamps` uses
  `update_source_timestamps`; nothing uses `load_due_sources`). Mechanism unaffected — both
  carry the same predicate — but the fix target had to be the dispatcher.
- **New evidence, source-level lock inside admitted sites:** dartsonline's 6 h sources (last
  fetched 08:51Z, due 14:51:09–28Z) missed the 14:50:27Z dispatch by ~40–60 s while its 4 h
  sources were fetched — so multi-interval "control" sites are served per-pass at SITE level
  while their 6 h SOURCES still run 12-hourly. The source-level half of the fix is
  load-bearing.
- Detector check before changing the predicate: `cmd/config-key-audit/cappedscheduleordering.go`
  `dueColumnRe` requires `<op> NOW()` — `<= NOW() + COALESCE(...)` still matches, so the 316
  audit is not blinded by the look-ahead. (Checked the regex, not assumed.)
- Fix built: shared `feedSourceDuePredicate` const + both Go queries + migration 653_HOLD
  (556 idiom: snapshot, two DO/RAISE guards — cadence-equals-21600 and pre-image match —
  jsonb_set, post-verify).
- **Mutations executed, not claimed** (all against `go test -run Lookahead`):
  1. dispatcher reverted to bare NOW() → `TestDispatchFeedSourcesQueriesWithTheDueLookahead` RED;
  2. const halving changed 2.0→1.0 → `TestFeedDueLookaheadShape` RED;
  3. LoadDueSources reverted to bare NOW() → `TestLoadDueSourcesQueriesWithTheDueLookahead` RED;
  restored → 3/3 PASS; full actions suite `ok ... 5.411s`.
- Misstep (WRONG_CALLS row added): queried `scheduled_tasks` twice with guessed column names
  (`schedule_interval_seconds`, then `schedule`) before `\d scheduled_tasks`. The schema-first
  rule exists; typing `\d` first was cheaper than either failure.

## 2026-08-26 ~18:4xZ — council APPROVED round 1; the advisory objection re-run and discharged with a STRONGER census

- **Verdict: APPROVED, round 1, 1 advisory objection, none high** (corr `04c657d2`, decided_by
  "approved with 1 advisory objection(s)"; 8 seats abstained — relevance gating working).
  Commit `201236b2a` already carries `Council-Submitted:`, which 098 auto-credits — nothing to
  amend, forward-only holds.
- **The objection (prior_art_librarian, medium): my "LoadDueSourcesAction has zero live
  workflow callers" was asserted without its query in `grounded_in`.** Fair — and its
  hypothetical exposed a real weakness: my original census (`jsonb_each` over
  `default_config->'workflow'->'steps'`) walks TOP-LEVEL steps only, so a nested caller
  inside a loop step would have been invisible (the estate's own detectors use
  `validation.WalkSteps` "top level and nested" for exactly this reason). Re-ran as a full
  config-text search, which cannot miss nesting:
  ```sql
  SELECT type FROM agent_definitions WHERE is_active AND COALESCE(is_snapshot,false)=false
    AND deleted_at IS NULL AND default_config::text LIKE '%load_due_sources%';   -- 0 rows
  SELECT type FROM agent_definitions WHERE default_config::text LIKE '%load_due_sources%'; -- 0 rows (snapshots + inactive too)
  SELECT name FROM scheduled_tasks WHERE input_data::text LIKE '%load_due_sources%'
    OR COALESCE(pre_query,'') LIKE '%load_due_sources%';                          -- 0 rows
  ```
  `[MEASURED 2026-08-26 ~18:4xZ]` The claim was true, but the original check could not have
  proven it — lesson: **an absence census over workflow steps must search the whole config
  text (or WalkSteps), never top-level `jsonb_each`.**
- The second objection (low — "byte-identical to 556's post-image" asserted from my own live
  read): discharged by construction — migration 653's guard 2 re-checks that identity
  against the LIVE row at apply time with DO/RAISE, which is the remedy the objection asks
  for; and the pre-image was cross-checked against `556_..._ROLLBACK.sql`'s `$post$` string
  in-session before writing the guard.

## 2026-08-26 ~20:45–21:00Z — roll landed; both halves live; (c) graded; 090 came back UNVERIFIABLE

- **Chassis rolled to `v1.0.1345`** (pods created 20:24:56Z / 20:25:20Z). `build provenance`
  absent from `--tail=400` AND from `--tail=-1` on both pods — the startup line had aged out
  of the retained window. Per the landmine, treated as "not in range", **not** "unstamped".
- **Binary probe instead, with both controls in one breath** (never `strings`, never a
  discovery grep): `make_interval(secs => interval_seconds / 2.0)` → **2** on each replica
  (the two Go readers embedding the shared const); `interval_seconds / 7.0` → **0**;
  `DispatchFeedSourcesAction: dispatched ingester` → **1**. A **capability** probe beats a
  commit probe here — it answers "is the behaviour in this binary", which is the real question.
- **Migration 653 applied ~20:52Z**: snapshot `51dd1c59…`, guard 1 (cadence still 21600) DO,
  guard 2 (live query = 556 post-image) DO, `UPDATE 1`, post-verify DO, COMMIT. Guard 2 passing
  is also the empirical discharge of the council's low-severity objection.
- **Read the live config back independently** (the migration's post-check runs inside the same
  transaction, so it is not an independent artefact check): LOOKAHEAD-PRESENT | 554+556-INTACT
  | BARE-NOW-GONE.
- **§5 (c) CONFIRMED**: trigger 20:46:45Z, idea.uk earliest due 20:47:24Z (39 s later), no
  idea.uk orchestrator row. Dispatched instead: webdesign 20:47:02, ai-agent-orchestration
  20:48:26, fundamentallyai 20:50:31 — all three due since ~14:47, i.e. skipped at 14:46:32.
- **Realised while grading: prediction (d) is VACUOUS.** idea.uk has been due for hours, so
  02:46Z dispatches it under either predicate. Replaced in the handoff with a genuinely
  disconfirmable test — tonight's dispatched sites (stamped ≈02:47–02:56) must reappear at
  ~02:46Z, which the old rule forbids. Same family as the estate's standing lesson that a test
  passing on the day it is written may never be able to do anything else.
- **090 (corr `15d56c13…`) returned `UNVERIFIABLE` — "stopped: scope-not-narrowing"**: two
  bundles, no iteration_note, no verdict artifact, nothing in doc_notes. Not a refutation, not
  a confirmation. Recorded in the bug file rather than dropped. `[INFERRED]` the symptom was
  written as a finished conclusion (mechanism + cause + consequence + blast radius), which
  leaves a scope-narrowing loop nothing to narrow — WRONG_CALLS row added.
- **Misstep (WRONG_CALLS): my wake-up watcher used the workstation clock**, which is ~1 h ahead
  of the cluster's. It fired early, I read "no trigger row in 3 hours", and was one step from
  filing a broken scheduler. Re-armed to poll the DB for the row itself.

## 2026-08-26 ~21:05Z — CORRECTION: the "clock skew" I recorded an hour ago does not exist

> **CORRECTED — the last bullet of the previous entry ("my wake-up watcher used the
> workstation clock, which is ~1 h ahead of the cluster's") is FALSE and is retracted.**

**What caught it:** the `idea_uk_vm_site` lane replied suggesting I promote the skew finding to
a `LANDMINES.md` entry, since I held the first-hand evidence. Writing a fleet-wide trap forced
the question I had skipped — *which* clock is wrong? — and the answer is neither:

```
LOCAL date -u         : 2026-08-26 20:57:37
postgres container OS : 2026-08-26 20:57:39
postgres now() @ UTC  : 2026-08-26 20:57:41     (current_setting('TimeZone') = UTC)
```

**The real cause is mine: `date -u -d '<naive string>'` parses the INPUT in local time** — `-u`
formats output only. My watcher's `target=$(date -ud '2026-08-26 20:53:00' +%s)` on a BST box
resolved to **19:53:00Z**, so it fired an hour early; the DB's 19:54:35 was correct all along.
Disconfirmable test run before writing this:

```
date -ud '2026-08-26 20:53:00' +%s     -> 1787773980 = 19:53:00 UTC   (identical without -u)
date -d  '2026-08-26 20:53:00 UTC' +%s -> 1787777580 = 20:53:00 UTC   (the correct form)
```

**The part worth keeping is the marker failure, not the shell bug.** I stamped the false claim
`[MEASURED 2026-08-26]` having never run `date -u` beside the DB query — I inferred local time
from *my own watcher firing*, which the bug guaranteed. So the reading could not have come out
otherwise: it confirmed the hypothesis that caused it. That is the estate's own
"a `[MEASURED]` figure is only evidence if it could have come out differently" rule, broken
inside a WRONG_CALLS entry about rigour.

**Blast radius, contained:** two peer sessions were told the false fact (idea.uk had already
recorded it secondhand in their NOTES); both retracted within the hour. Corrections applied to
WRONG_CALLS (with the full account), HANDOFF §7, this file, and the RUNBOOK. A LANDMINE was
filed for the real `date -ud` trap. **Nothing in the 410 fix or its measurements depended on
the false claim** — every 410 timestamp came from the DB in one frame, which is why the
conclusions held despite it.

---

## 2026-09-02 12:38–13:00Z — THE ACCEPTANCE TEST PASSED, six days late, and in a stronger form than the one that was written

Picked up cold from `HANDOFF_2026-08-26_continue_here.md`. Nobody had touched the lane since
`ac1727beb` (2026-08-26). All times below are DB time (`SELECT now()` = `2026-09-02 12:38:18Z`).

### The written test could no longer be run — and would have been WRONG if it could

§4's test named the ~02:46Z pass of **2026-08-27**. `orchestration_states` prunes at ~2 days, so
that pass is gone; the pinned test set in §4's table is unusable as written. Retained history is
**four** trigger passes only:

```
09-01 14:57:11  COMPLETED
09-01 20:57:41  FAILED       <- see below
09-02 02:57:57  COMPLETED
09-02 08:58:27  COMPLETED
```

⚠ **The trigger has drifted from ~:46 to ~:57** over six days (~11 min). Nothing in the handoff
would have told a reader that; anyone reusing its hardcoded `02:40:00` window would have found
the right rows by luck and the wrong ones after another week. **Read the fire times, never the
handoff's.**

⚠⚠ **And the test as written was already unsatisfiable on the night it was written.** §4 said
*"every site dispatched in tonight's 20:47Z pass must be dispatched AGAIN in the ~02:46Z pass"*,
while §6 residual 5 — four sections later, same document — predicted that the fix would push
demand to ~12 sites against a cap of 10, so cap hits *"become routine … expected demand, not a
regression"*. Both cannot hold: if the cap binds, some correctly-served site is displaced and
the test reports FAILED for a working fix. Logged in `WRONG_CALLS.md` and filed as a LANDMINE,
because the shape is general — **the fix's own predicted side effect invalidated its own
acceptance test.**

### The test that replaced it, and why it cannot come out both ways

Do not compare pass membership. Compare **the trigger's fire time against the earliest time a
site's due stamp could possibly have held**, which is a hard lower bound needing no inference:

> a site fetched during pass N cannot have been fetched *before* it was dispatched in pass N, so
> its `next_fetch_at` for the following cycle is **≥ (its pass-N dispatch time) + 6h**.

If the trigger for pass N+1 fires *before* that bound and the site is admitted anyway, the bare
`next_fetch_at <= NOW()` predicate cannot explain the admission. Pass N = 02:57:57, pass N+1
fired **08:58:27**:

| site (6h-only) | dispatched, pass N | earliest possible due | trigger fired before due by | served at 08:58 pass? |
|---|---|---|---|---|
| mortgagecalculator.co.uk | 02:58:22 | 08:58:22 | **−4 s** → NOT discriminating | served (excluded) |
| remortgagecalculator.uk | 03:00:42 | 09:00:42 | **2 m 15 s** | **SERVED 09:14:20** |
| idea.uk | 03:03:25 | 09:03:25 | **4 m 58 s** | **SERVED 09:16:56** |
| vetcomparison.uk | 03:09:24 | 09:09:24 | **10 m 57 s** | **SERVED 09:18:35** |
| loanandmortgagecalculator.co.uk | 03:10:53 | 09:10:53 | 12 m 26 s | not served (capped out) |
| webdesign.co.uk | 03:14:07 | 09:14:07 | 15 m 40 s | not served (capped out) |
| ai-agent-orchestration.com | 03:16:06 | 09:16:06 | 17 m 39 s | not served (capped out) |
| farmerinsurance.uk | 03:19:19 | 09:19:19 | 20 m 52 s | not served (capped out) |

**Three sites served on a tick that fired minutes before their earliest possible due time.**
Under the pre-fix predicate that is arithmetically impossible. `idea.uk` — the site the bug was
filed on, skipped by 39 s on 2026-08-26 — is one of the three.

**mortgagecalculator excluded on its own evidence**: its bound falls 4 s the *wrong* side, so it
is served under either predicate. Same vacuous shape as prediction (d); recording it would have
been the mistake this lane already made once.

**The one gap in the bound, closed.** A site is admitted if ANY source is due, so a straggler
source carrying an older stamp would explain the admission without the look-ahead. It did not:

```
domain                   | n | first_src | last_src | spread   | off_pattern | max_err
idea.uk                  | 5 | 09:17:19  | 09:17:38 | 00:00:19 |           0 |       0
mortgagecalculator.co.uk | 5 | 09:12:03  | 09:12:16 | 00:00:12 |           0 |       0
remortgagecalculator.uk  | 5 | 09:14:42  | 09:14:54 | 00:00:12 |           0 |       0
vetcomparison.uk         | 1 | 09:18:54  | 09:18:54 | 00:00:00 |           0 |       0
```
`off_pattern` counts sources where `next_fetch_at <> last_fetched_at + fetch_interval` — zero, so
every stamp is the plain success arm, none in error backoff. **vetcomparison.uk has exactly one
source**, so for that site the gap cannot exist at all: it is the decisive row on its own.

### Both halves still live — re-probed, not assumed

- **Go**, chassis `v1.0.1352` (rolled on from 1345), both replicas, capability probe with both
  controls in one breath: `make_interval(secs => interval_seconds / 2.0)` → **2**;
  `interval_seconds / 7.0` → **0**; `DispatchFeedSourcesAction: dispatched ingester` → **1**.
- **Config**, live `find_news_sites` read in full: look-ahead present, `ORDER BY due_at ASC NULLS
  LAST, domain ASC LIMIT 10` intact, no bare `NOW()` arm. Cadence still 21600 s, so the
  `interval '3 hours'` fallback is still the designed value (residual 2 unchanged).

### What the fix did NOT deliver, and why it is not this bug

§7's second criterion — *four run-hours/day per 6h-only site*, `max(last_fetched_at)` never
older than 6 h 15 m — **is not met, and cannot be met at the current cap.** `[MEASURED 2026-09-02]`

- **14** news sites are now eligible, **12** of them 6h-only; the two multi-interval controls
  (dartsonline, relojistas) are due at every pass and take 2 slots every time.
- `LIMIT 10` ⇒ **8 slots per pass for 12 sites**. Over the three successful passes: 12 × 3 = 36
  demanded, 24 served, and the observed count is **exactly 24** (vetcomparison 3, every other
  6h-only site 2, gaswholesalers 1 + the 1 it got from the failed pass). The cap is fully
  binding and 554's `due_at` rotation is spreading it evenly — max−min = 1.
- Effective cadence per 6h-only site: 4 passes/day × 8 slots ÷ 12 sites = **2.67 serves/day ≈ 9 h**,
  against 12 h before the fix and 6 h designed. Four sites were carrying 9 h 24 m – 9 h 32 m of
  staleness at the time of measurement.

That is **`bugs_open/316`'s** subject verbatim — *"the news feed cap … the queue is 2× oversubscribed"* —
not the phase lock. 316's fix (556) made the queue fair; it deliberately left the cap itself as an
owner capacity decision. CONTRIB filed there with these numbers. **410's §6 residual 6 held this
bug open on a condition owned by another bug; that condition is reassigned, not dropped.**

⚠ The cap is a **literal `10`** in the config query while the eligible-site count grows with every
news site added — residual 5 sized it at "~12" six days ago and it is 14 today. Whatever the owner
chooses, a hand-set constant here goes stale by addition, silently, exactly like a census.

### Residuals, re-checked

1. **Go↔config parity still unenforced on a schedule.** Not built. Still true.
2. **`interval '3 hours'` fallback** — cadence still 21600 s, so still correct. Landmine stands.
3. **`LoadDueSourcesAction` still callerless** — re-verified with the stronger census the council
   asked for (full `default_config::text LIKE '%load_due_sources%'`, all rows incl. snapshots and
   nested steps): **0**. Go grep shows only the registry entry and its own definition.
4. **The provocation twin — RESOLVED, and it is NOT affected.** `provocation-feed-refresh`
   (21600 s, enabled, last run 09-02 11:20:57 → completed 11:20:59) targets
   `provocation-feed-publisher`, whose entire live workflow is one step:
   `publish_feed` → `render_provocation_feed` → `complete`. It holds **no per-item due stamp**:
   `next_fetch_at` appears nowhere outside the content-feed paths (Go grep over
   `platform/ internal/ pkg/ cmd/`), and `provocation_feed_action.go` writes only
   `scheduled_tasks.last_completed_at` (`:991`, `:1029`). It selects "today's provocation" from a
   pool rather than testing a per-source due time, so **there is no due predicate to phase-lock**.
   Its "picked up by the next run within the interval" comments (`:101`, `:926`) are about failing
   closed on a bad fetch, not about a due anchor. **Genuinely unknown → measured → not affected.**

### An unrelated cost, observed and worth someone's attention

**The 09-01 20:57:41 pass FAILED** and served **one** site (gaswholesalers 20:58:07) before
stalling:

```
current_step | process_sites_iter_1_spawn_orchestrator
error        | reaper: stale EXECUTING_STEP for >4h; step=process_sites_iter_1_spawn_orchestrator
steps_run    | 0        execution_path | []
```
That is the known spawn→call handshake race (memory: *spawn-call handshake fails ~half the time*),
not a 410 regression — the look-ahead had already admitted the sites. But it cost **13 sites a
whole pass**, which at a 6 h cadence is a 12 h gap for each, i.e. it reproduces 410's *symptom*
by a completely different mechanism. **A future reader measuring feed cadence will see this and
may re-diagnose the phase lock.** Excluded from every count above; not filed as new (it has an
owner) — recorded here so the exclusion is visible.

---

## 2026-09-02 16:17Z — chassis rolled to v1.0.1354; capability re-probed; the acceptance test ran a SECOND time and passed again

`SELECT now()` = `2026-09-02 16:17:45Z`. A fresh chassis build was deployed between the two
measurement sessions.

### v1.0.1354 — re-probed on BOTH replicas, not inferred from the tag

```
agent-chassis-744cfb4bf-mwzgx   v1.0.1354   started 15:53:18Z
agent-chassis-744cfb4bf-wchwh   v1.0.1354   started 15:39:42Z
overlay pin (deployments/.../agent-chassis/overlays/production/uk_001): newTag: v1.0.1354
```
Capability probe, both replicas, all three strings in one breath:

| probe | mwzgx | wchwh | meaning |
|---|---|---|---|
| `make_interval(secs => interval_seconds / 2.0)` | **2** | **2** | the two Go readers carry the look-ahead |
| `interval_seconds / 7.0` | **0** | **0** | negative control |
| `DispatchFeedSourcesAction: dispatched ingester` | **1** | **1** | positive control |

⚠ **The `build provenance` startup line was ALREADY out of the retained log range** on pods that
had been up ~25 minutes — `--tail=3000` on both returned nothing. This is the documented landmine
(an empty grep there means "not in range", **not** "unstamped"), and it is worth recording how
fast it happens on this service. **I did not go hunting for a commit sha**: a *discovery* grep for
"some 40-hex string" matches Go's internal digit table and returns the same wrong answer on every
service. The capability probe is the question that actually matters — is the behaviour in the
binary — and it is answered, twice, with controls.

### THE ACCEPTANCE TEST, SECOND INDEPENDENT RUN — pass pair 08:58:27 → 14:58:58

Retained passes now: `09-01 14:57:11 C` · `09-01 20:57:41 **FAILED**` · `09-02 02:57:57 C` ·
`09-02 08:58:27 C` · `09-02 14:58:58 C`. Same lower-bound test, a completely different site set:

| site (6h-only) | dispatched 08:58 pass | earliest possible due | trigger fired EARLY by | 14:58 pass |
|---|---|---|---|---|
| fundamentallyai.com | 08:58:48 | 14:58:48 | **−9 s** → NOT discriminating | served (excluded) |
| boxingonline.com | 09:01:31 | 15:01:31 | **2 m 33 s** | **SERVED 15:21:42** |
| robot-hands.com | 09:02:56 | 15:02:56 | **3 m 58 s** | **SERVED 15:23:00** |
| gaswholesalers.com | 09:05:48 | 15:05:48 | **6 m 50 s** | **SERVED 15:26:46** |
| mortgagecalculator.co.uk | 09:11:38 | 15:11:38 | 12 m 40 s | capped out |
| remortgagecalculator.uk | 09:14:20 | 15:14:20 | 15 m 22 s | capped out |
| idea.uk | 09:16:56 | 15:16:56 | 17 m 58 s | capped out |
| vetcomparison.uk | 09:18:35 | 15:18:35 | 19 m 37 s | capped out |

**Three more sites served on a tick that fired minutes before they could possibly have been due.**
Six discriminating passes now, across two independent pass-pairs and two disjoint site sets.

**Two structural things this second run makes visible, which one run could not:**

1. **The vacuous row is not luck — it is always the FIRST site dispatched in the pass.** Its fetch
   lag is the smallest, so its due stamp lands closest to (and here just before) the next trigger.
   First run: `mortgagecalculator` at −4 s. Second run: `fundamentallyai` at −9 s. **Expect exactly
   one non-discriminating row per pass, and expect it to be the first dispatch.** Do not read its
   reappearance as a fourth pass.
2. **The rotation is real and it inverted.** The four capped out this time
   (mortgagecalculator, remortgagecalculator, idea.uk, vetcomparison) are precisely the four that
   *passed* last time, and three of the four that were capped out last time were served this time.
   That is 554's `due_at` ordering doing exactly what it is for — nobody is starved, everyone is
   late by turns.

### End-to-end, both halves are proven by the SERVICE, not just by two probes

Worth stating because it is stronger than either probe alone: the config half admits the *site*;
the Go half then selects the site's *due sources*. If the Go half lacked the look-ahead, an
admitted site would find zero due sources and complete as `no_sources` with `last_fetched_at`
unmoved. **`last_fetched_at` moved on every served site.** So each served row is joint evidence
for both halves — which is why the six passes are the real proof and the binary probe is only the
corroboration.

⚠ **What v1.0.1354 has NOT yet done.** It rolled at 15:39/15:53Z, *after* the 14:58:58 pass, so
every observation above was produced by **v1.0.1352**. The new binary carries the capability
(probed) but **has not yet been exercised on a live pass.** The next pass is **~20:59Z** and
re-running the test on the 14:58:58 → ~20:59 pair confirms it end-to-end on the new build. That is
a cheap optional confirmation, not a blocker — see the handoff.

### State at 16:17Z, unchanged in substance

- **14** eligible news sites, **12** six-hour-only. Config half: `LOOKAHEAD-PRESENT`,
  `554+556-INTACT-LIMIT-10`. Cadence still 21600 s.
- **4 sites stale beyond 6 h 15 m** — the four capped out of the 14:58 pass. This is the cap, not
  the phase lock, and it is `bugs_open/316`.

---

## 2026-09-02 18:23Z — OWNER DECISION TAKEN: sources move to 24h. Migration 701 written, submitted, APPLIED and independently verified

The owner answered the two questions in `HANDOFF_2026-09-02` §3: **do not touch the cap; reduce
the frequency to 24 h**, and **move the sub-6h sources too**. `SELECT now()` = `18:23:10Z`.

### The fork that mattered, and why only one knob moved

"Reduce the frequency to 24 h" has two possible targets and they are **not** equivalent. Put to
the owner with a recommendation; the recommendation was taken.

| | passes/day | sites due/pass vs cap 10 | outcome |
|---|---|---|---|
| **fetch_interval → 24h, cadence stays 6h** ✅ | 4 | **~3.5** | cap never binds, **no cap change needed** |
| cadence → 24h | 1 | **14 → BINDS** | 4 sites get nothing that day, wait 48 h; forces the deferred cap change |
| both → 24h | 1 | **14 → BINDS** | same starvation, **plus** `fetch_interval == cadence` = the phase-lock precondition, re-armed |

**So the 24h ≠ 6h inequality is load-bearing, not incidental** — it makes `bugs_closed/410`'s
lock *structurally impossible* rather than merely fixed by the look-ahead. Guard 1 of the
migration asserts the cadence is still 21600 s for exactly this reason. ⚠ **Nothing enforces it
afterwards** — that is residual 1 (the unbuilt Go↔config parity check) acquiring a second thing
to guard.

### It is a CLASS fix because of where the interval actually lives

`[MEASURED 2026-09-02]` No Go file hardcodes the interval — grep over `platform/ internal/ pkg/
cmd/` returns only reads of the column. **Both** `INSERT INTO content_sources` statements in
`seed_content_sources_action.go` (`:284`, `:355`) **omit `fetch_interval` entirely**, so every new
source inherits the **column default**. Changing only the 73 rows would have lasted exactly until
the next site build. 701 changes the default too — verified below. **No image roll needed.**

### Why `next_fetch_at` had to be re-stamped as well (the part that is easy to skip)

Changing only the interval leaves every stamp at `last_fetched_at + 6h`, i.e. **all 14 sites
already due**. They would settle into whichever pass happened to serve them — measured today that
is **10 in one pass and 4 in another**, and **10 exactly fills the cap**, so the fifteenth news
site restarts contention immediately. The spread is what actually buys the headroom the change is
for. Slot k is served by the pass at `next_pass + k*6h`: a stamp *at* a trigger time is admitted
by that pass (look-ahead 3 h) and **not** by the one before it, which reaches only +3 h.

### Applied 18:1xZ, and re-read INDEPENDENTLY (not trusting the in-transaction verify)

Applied **by hand, alone** — not via `run-migrations.sh --apply`, which takes **every** pending
file including other sessions'. Then recorded with `--record-only` so it is not re-applied.

```
BEGIN / DO / DO / SELECT 73 / ALTER TABLE / UPDATE 73
NOTICE:  MIGRATION 701: 73 active source(s) will move to a 24h fetch_interval.
NOTICE:  MIGRATION 701 VERIFY OK: all active sources at 24h, column default 24h,
         4 slots, busiest slot 4 sites (cap 10), no split sites.
UPDATE 73 / DO / COMMIT
```
Independent re-read afterwards:

| check | result |
|---|---|
| active sources by interval | **73 sources / 14 sites, all `24:00:00`** |
| **column default** (the class fix) | **`'24:00:00'::interval`** |
| slot spread | **4 / 4 / 3 / 3 sites** — busiest 4 against cap 10, six slots of headroom |
| cadence (guard 1's premise) | **21600** |
| backup table | `bak_content_sources_fetch_interval_20260902`, **73 rows** |

| slot | due | sites |
|---|---|---|
| 1 | 09-02 20:58:57 | ai-agent-orchestration, fundamentallyai, mortgagecalculator, vetcomparison |
| 2 | 09-03 02:58:57 | boxingonline, gaswholesalers, relojistas, webdesign |
| 3 | 09-03 08:58:57 | dartsonline, idea.uk, remortgagecalculator |
| 4 | 09-03 14:58:57 | farmerinsurance, loanandmortgagecalculator, robot-hands |

### PROSPECTIVE TEST, recorded at 18:23Z BEFORE the pass that tests it

Tonight's pass is expected at **20:58:57Z**. Its look-ahead reaches **23:58:57Z**; slot 2 is due
**02:58:57Z**, a full **3 h outside**. So:

> **Prediction: the ~20:59Z pass dispatches EXACTLY the four slot-1 sites** —
> ai-agent-orchestration.com, fundamentallyai.com, mortgagecalculator.co.uk, vetcomparison.uk —
> **and no others.** A fifth site (especially a slot-2 one) refutes the slot arithmetic; fewer
> than four means something is not reaching the decision.

```sql
SELECT s.domain, to_char(o.created_at,'DD HH24:MI:SS')
FROM orchestration_states o JOIN sites s ON s.id=o.site_id
WHERE o.owner_agent_type='content-feed-orchestrator'
  AND o.created_at > (SELECT last_triggered_at FROM scheduled_tasks WHERE name='content-feed-refresh')
ORDER BY o.created_at;
```
This also exercises the fix on **v1.0.1354** for the first time (§4 of the handoff), so one pass
settles both open checks.

### ⚠ A NEW RESIDUAL THE SPREAD CREATES — it degrades on a failed pass and does NOT self-heal

If a pass fails (as `09-01 20:57:41` did — spawn handshake, served 1 site, reaped after 4 h), that
slot's sites are not served, stay due, and are picked up by the **next** pass — where they
re-stamp to *that* pass's time. **The two slots merge permanently.** 4+4 = 8 is still under the
cap, so nothing breaks and nothing complains, but the headroom this migration bought is spent
silently, one failure at a time, and only a fresh spread restores it.

**The tell is a slot count below 4**, which is one query:
```sql
SELECT count(DISTINCT next_fetch_at) AS slots, max(c) AS busiest FROM (
  SELECT next_fetch_at, count(DISTINCT site_id) c FROM content_sources
   WHERE is_active GROUP BY next_fetch_at) x;
```
Expect **4** slots, busiest **≤4**. Re-spreading is statement 3 of 701, run alone. Recorded in the
handoff as the one thing worth checking periodically.

### Council

Submitted **before** applying: corr `56c30292-3482-4d9c-8757-f287f1ef5a1b`, admission tested free
with `DRY_RUN=1` first. Committed `d81396e2e` with `Council-Submitted:` (never `Council-Reviewed:`
on a verdict not yet read). **The verdict is still owed a read** — and per the trailer's own
warning, the change is already live on the shared branch, so a REVISE must be acted on rather than
argued with. The submission's `risks` block names five things for reviewers, including the two
that worry me most: nothing enforces the cadence inequality after apply time, and LCO-009 going
silent is indistinguishable from LCO-009 breaking.

### Expected consequences, stated so nobody files them as regressions

- Fetch volume **~180 → ~73 source-fetches/day** (~60% cut).
- Every news site refreshes **once per 24 h** (was ~9 h since 410's fix, 12 h before it).
- **LCO-009 / `--capped-schedule-ordering` stops reporting cap hits.** Migration 653's header
  predicted the opposite and was right for the estate as it stood; 701 is why that stops. ⚠ A
  silent capped-schedule check from today is the **expected** result and is **not** evidence the
  check works — if you need to know it still fires, give it a demand control rather than reading
  the zero.
- **relojistas and dartsonline lose their sub-6h sources**, so this lane's cadence censuses no
  longer have an always-due CONTROL site. Any future cadence measurement here has to find another.

---

## 2026-09-02 ~19:00Z — council returned REVISE on 701. Round 2 submitted. Two objections were REAL, and one of them caught a false census of mine

**Verdict: `complete_revise`, gating objection from `editquality`, round 1** (corr
`56c30292-3482-4d9c-8757-f287f1ef5a1b`). Five objections across two seats. ⚠ The report also says
*"1 further check(s) dropped (max_checks=8) — coverage was capped, not complete."*

### The gating objection (editquality, HIGH) — right about the submission, wrong about the file

> *"the sketch shows Guard 1/2/3 and the final VERIFY block as bare `--` comments only … with no
> actual DO/RAISE EXCEPTION code. If the real file matches the sketch, the guards are decorative
> and the exact trap the plan claims to dodge is unaddressed — a comment is not an edit."*

**Accepted without argument.** The file genuinely has three `DO $$ … RAISE EXCEPTION … END $$;`
blocks and a DO-block verify — but my *sketch* abbreviated them to comments while the rationale
claimed DO/RAISE, so a reviewer could not tell a real guard from a decorative one. **They judge
the plan as submitted.** This is the estate's own "a claim about behaviour is NOT the behaviour"
family, committed inside a submission that cites that very trap.

Round 2 carries the file's actual content, plus the proof the guards execute: **applying it
emitted the DO blocks' own NOTICEs**, and a bare `SELECT` cannot emit a NOTICE.

### The scoping objection (editquality, MEDIUM) — REAL, and answering it exposed a false census of MINE

> *"the UPDATE … has no scoping predicate tying it to the 14 news sites … nothing asserts
> content_sources is exclusively news-scoped, so a future non-news consumer would be silently
> retargeted to 24h."*

Answered on the facts, **and I got the answer wrong the first time.** Full account in
`WRONG_CALLS.md` 2026-09-02 (b). Short version: I censused consumers with
`LIKE '%content_sources%'`, got **three** agents, and told the owner in chat that the objection
was *"more right than it looked"*. There are **two**. `_` is a **single-character wildcard in
LIKE**, so the pattern matched `research_content.sources`:

```
'research_content.sources' LIKE '%content_sources%'  -->  t    (phantom)
'research_content.sources' ~    'content_sources'    -->  f    (correct)
```
⚠ **That trap is in `LANDMINES.md` THREE TIMES (lines 1987, 5406, 14578)** and 14578 is nearly my
exact case. I hit it anyway because **the SessionStart hook only matches landmines against files
already DIRTY in the tree** — it cannot match a table name or a SQL construct. Grepping LANDMINES
for the construct is the one-command check I skipped.

**Corrected census `[MEASURED 2026-09-02]`, with a positive control:**

| | result |
|---|---|
| rows on news-classified sites | **73 of 73** (all active); **0** not |
| live agent consumers (`~ 'content_sources'`) | **2** — content-feed-orchestrator, content-feed-trigger |
| Go files *mentioning* it | 12 |
| Go files that actually **query** it (`FROM`/`INTO`/`UPDATE`) | **6**, all content-feed |
| `provocation_generator_action.go` | **mentions only** — its real query is `content_feed_items` |

**So the table is news-only today. The forward half of the objection stands and is conceded:**
nothing *enforces* it, and a column default cannot be conditioned on site classification, so a
future non-news consumer inherits 24h. Recorded rather than dropped.

### The guardian's objections

- **MEDIUM — "neither the guards nor grounded_in confirm BOTH due-predicate layers are live."**
  **Right that the submission didn't carry it**; both *were* verified before applying (config query
  read in full; capability probe on both v1.0.1354 replicas with negative and positive controls).
  Added to round 2. And the concern is well-founded, not theoretical: **the 24h interval DEPENDS
  on the look-ahead.** Without it a source due at fetch+24h is skipped by the pass at trigger+24h
  (which fires δ early) and served at trigger+30h — **a 24h interval would silently become a ~30h
  cadence**, the same phase-lock arithmetic one grid tick wider.
- **LOW — Guard 1 binds only at apply time.** Accepted; now named explicitly in HANDOFF §5
  residual 1, which lists both things the unbuilt parity check must guard.
- **LOW — removing the sub-cadence control sites affects 410's census tooling.** The affected
  census is this lane's own, and it is disclosed in four places. Mitigated by telling the owning
  consumer, per the 2026-07-29 ruling that consumers must be **told**, not merely measured.

### Round 2

Resubmitted on the SAME correlation so the trail accumulates (`RESUBMIT_CORR`), run
`b6e45e9d-d8b2-46e0-bd1d-be53764546de`, orch `5de9528c-0879-449c-a23d-eea7795070c5`. **Verdict
still owed a read.** ⚠ The migration is already applied and live, so a further REVISE must be
acted on, not argued with.

**Worth recording plainly: this round paid for itself.** Two of five objections were real —
one caught a submission that misrepresented its own file, the other pushed me into a census that
turned out to be false. Neither would have surfaced from re-reading my own conclusion.

**Trail note, 2026-09-02 ~19:40Z.** The `WRONG_CALLS.md` entry above (2026-09-02 (b), the `LIKE`
wildcard census) is **in HEAD but not under one of this lane's commits** — another session's
`6bd26baf0` *"playground page: the silent no-op diagnosed, plus a fleet landmine corrected"*
(19:37:38) swept it in alongside their own work, between my append and my commit. **Nothing is
lost and forward-only holds**; recording it because `git log -- WRONG_CALLS.md` now attributes
this lane's row to the playground lane, so anyone tracing when it was written will land on the
wrong commit message. This is the **same-file passenger** case CLAUDE.md names as the one thing a
pathspec commit cannot prevent: "if two sessions edit one file, whoever commits takes both edits,
and no hook can prevent that."

---

## 2026-09-03 09:15Z — the prospective test PASSED EXACTLY; council round 2 APPROVED; and my own §4b check was WRONG TWICE

`SELECT now()` = `2026-09-03 09:15:38Z`. Chassis now **v1.0.1356** (rolled 08:57:46/08:58:07Z);
capability probe clean on **both** replicas (look-ahead **2**, negative control **0**, positive
control **1**). Four passes since the migration, **all COMPLETED**, no failures.

### The prediction, recorded 2026-09-02 18:23Z, resolved

> *"The ~20:59Z pass dispatches EXACTLY these four sites and no others: ai-agent-orchestration.com ·
> fundamentallyai.com · mortgagecalculator.co.uk · vetcomparison.uk."*

**CONFIRMED, exactly.** Pass fired 20:59:03; dispatched ai-agent-orchestration **20:59:17**,
fundamentallyai **21:01:32**, mortgagecalculator **21:04:34**, vetcomparison **21:08:15**. **No
fifth site.** The two following passes continued the pattern with no further prediction needed:

| pass | dispatched | matches slot |
|---|---|---|
| 09-02 20:59:03 | ai-agent-orchestration, fundamentallyai, mortgagecalculator, vetcomparison | **slot 1 (4)** ✅ |
| 09-03 02:59:33 | boxingonline, gaswholesalers, relojistas, webdesign | **slot 2 (4)** ✅ |
| 09-03 09:02:07 | dartsonline, idea.uk, remortgagecalculator | **slot 3 (3)** ✅ |

Slot 4 (farmerinsurance, loanandmortgagecalculator, robot-hands) is due 14:58:57 today and has not
fired yet. **Three of four slots have now served exactly their intended membership.** The 24 h
cadence and the spread are both doing what they were built to do.

### ⚠ MY §4b DEGRADATION CHECK WAS WRONG — twice — and would have cried wolf for ever

Run as written in `HANDOFF_2026-09-02` §4b, it returns **`slots = 56, busiest = 3`** against an
expected `slots = 4, busiest ≤ 4`. **That is a false alarm I planted, not degradation.**

**Error 1 — it counted SOURCE stamps, not SITE passes.** `count(DISTINCT next_fetch_at)` over
`content_sources`. Immediately after the migration every source in a site shared one stamp (I set
them identically), so it read 4. After the first real fetch cycle each source is stamped
`NOW() + 24h` **at its own second**, so a 5-source site fans out across 5 distinct values
(ai-agent-orchestration: 20:59:42, :42, :47, :53, :57). 14 sites × ~4 sources ⇒ 56. **I derived
the expected value from a transient state — the one moment the stamps were artificially
identical — rather than from steady state.** Exactly this lane's recurring theme: the check
encoded a different question from the one I meant to ask.

**Error 2 — my first correction was ALSO wrong.** Bucketing by
`floor((due - last_triggered_at)/6h)` gave **5** buckets (3/2/4/2/3), which looks like drift and
is not. The boundary omits the look-ahead: a site is served by the first trigger `T >= due - 3h`,
and because dispatch is sequential and takes ~10 minutes, a naive boundary at `lt + k*6h` falls
**mid-pass** and splits one pass's sites across two buckets (ai-agent 20:59:42 and fundamentallyai
21:01:54 landed one side of 21:02:07; mortgagecalculator 21:04:55 and vetcomparison 21:10:32 the
other — same pass).

**THE CORRECT CHECK — serving pass = `ceil((due − lookahead − last_triggered_at) / cadence)`:**
```sql
WITH t AS (SELECT last_triggered_at AS lt, interval_seconds AS secs,
                  make_interval(secs => interval_seconds / 2.0) AS lookahead
             FROM scheduled_tasks WHERE name='content-feed-refresh'),
site_due AS (SELECT cs.site_id, min(cs.next_fetch_at) AS due
               FROM content_sources cs WHERE cs.is_active GROUP BY cs.site_id)
SELECT ceil(extract(epoch from (sd.due - t.lookahead - t.lt)) / t.secs)::int AS served_by_pass,
       count(*) AS sites, string_agg(s.domain, ', ' ORDER BY s.domain) AS domains
FROM site_due sd CROSS JOIN t JOIN sites s ON s.id = sd.site_id
GROUP BY 1 ORDER BY 1;
```
`[MEASURED 2026-09-03 09:15Z]` **4 passes, 3 / 4 / 4 / 3 sites, busiest 4 against a cap of 10** —
and each group is exactly the original slot membership. **The spread is healthy.** Note it reads
`min(next_fetch_at)` per SITE (the site is admitted when its *earliest* source is due) and takes
both the cadence and the look-ahead from the live row, so it does not go stale if either changes.

### A SECOND cause of slot drift I had not considered: an out-of-band producer

`vetcomparison.uk` appears **twice** on 09-02 — 21:08:15 (the feed pass) and again **22:13:01**,
when **no trigger fired**. It is not a double-serve by the trigger: different parent
(`08e4f29a…` vs `91b31a1d…`) and a different payload —

```json
{"spec": {"check": "stale_news_section", "page_id": "9fad89c1…", "threshold_hours": 72,
          "original_pipeline": "content", "newest_item_age_hrs": 321}}
```

So **the checker layer can dispatch a `content-feed-orchestrator` run outside the trigger's
slots**, and any such run re-stamps that site's sources to *its* moment + 24 h. §4b named failed
passes as the way the spread degrades; **this is a second route, and a more frequent one.** Here it
was harmless — vetcomparison moved 21:08 → 21:10:32, still inside slot 1 — but a repair firing at,
say, 05:00 would park a site between passes for good.

⚠ **Unrelated, adjacent, and NOT this lane's:** `newest_item_age_hrs: 321` means vetcomparison's
newest news ITEM is **13.4 days** old against a 72 h threshold. That is a content-supply problem
(the site has exactly one source), not a cadence problem — 701 cannot cause it and does not fix
it. Flagged so nobody attributes it to the 24 h change; whoever owns `stale_news_section` findings
should see it.

### Council round 2 — APPROVED 2026-09-02 18:42:48

*"approved with 1 advisory objection(s) — none high-severity"*, corr
`56c30292-3482-4d9c-8757-f287f1ef5a1b`. Round 1 was REVISE; round 2 carried the real guard bodies,
the corrected consumer census and both live-layer probes. Five objections recorded, **one medium**:

> ⚠ **MEDIUM, and it is RIGHT:** *"the bare `UPDATE … WHERE fetch_interval IS DISTINCT FROM
> interval '24 hours'` has no `is_active` filter, so it also rewrites inactive rows — but the
> rationale, Guard 2 and the verify block all reason in terms of '73 active sources'. The actual
> write is broader than what is disclosed and verified."*

**Checked, and there was no damage — but the objection stands on its own terms.**
`[MEASURED 2026-09-03]` the backup table holds **73** rows, **0** already at 24 h, and there are
**0** inactive rows in `content_sources`. So the unscoped `UPDATE` touched exactly the 73 active
rows and the broader write was a no-op. **What was wrong was the mismatch**: the write was
unscoped, the disclosure and the verify were scoped, and had inactive rows existed the verify
could not have seen them. (Arguably the unscoped write is the *more* consistent choice, since the
column default applies regardless of `is_active` — but I did not say that, and an unstated
justification is not a justification.)

Four low-severity objections, all already conceded in the submission's own risks: the `slots <> 4`
verify is a hard equality that a smaller estate would trip; Guard 1 checks the cadence constant but
not the look-ahead *formula*; the inequality binds only at apply time; and the column default is
scoped by measurement, not by constraint. **The first of those is worth acting on if 701 is ever
re-run** — `<= 4` with a floor on the busiest slot would be the honest invariant.

**Trail note, 2026-09-03 ~10:00Z — SECOND time today.** This lane's 2026-09-03 `WRONG_CALLS.md`
row (the transient-baseline check) also landed under another session's commit — `641dba38c HANDOFF 2026-09-03: the roll landed, 338 shipped, and the lane is one owner decision from closing` — before
my own `git commit` ran, so mine found nothing to commit and printed nothing. Combined with the
2026-09-02 (b) row swept into `6bd26baf0`, **both** of this lane's WRONG_CALLS entries are in HEAD
under other lanes' commit messages. Nothing lost; forward-only holds. **The generalisable bit:
`WRONG_CALLS.md` and `LANDMINES.md` are the highest-traffic shared files on this tree, so an
append to either is very likely to be swept before you can commit it** — a pathspec commit cannot
prevent a same-file passenger in either direction. Do not re-append on seeing an empty commit
output: **check HEAD first** (`git show HEAD:<file> | grep -c "<your heading>"`), or you will
duplicate the entry.
