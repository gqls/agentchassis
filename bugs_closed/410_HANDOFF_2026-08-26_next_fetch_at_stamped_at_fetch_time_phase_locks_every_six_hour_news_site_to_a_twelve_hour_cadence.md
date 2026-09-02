# 410 — `next_fetch_at` is stamped at FETCH time (`NOW() + fetch_interval`), so a 6-hour interval on a 6-hour trigger falls due SECONDS after the next pass fires: every news site whose sources are all 6-hourly is served every OTHER run — a 12-hour cadence wearing a 6-hour label

**Filed 2026-08-26 by the `idea_uk_vm_site` lane** (`docs/agent_docs/docs024_key_docs_latest/idea_uk_vm_site/RUNNING_NOTES_idea_uk_vm_site.md` §X.65).
Found because one site's `content_sources.last_fetched_at` did not move across a trigger pass the
handoff had recorded as "COMPLETED".

**Diagnosed first-hand rather than via the `090` loop — declaring the substitute per the 2026-07-31
ruling.** Every link below is quoted from the live system (the live `agent_definitions` step query,
the `scheduled_tasks` row, `content_sources` timestamps, `orchestration_states` runs) or from the Go
at a cited line; the claim is closed-form arithmetic on those; a **positive control** (sites with a
sub-6h source) behaves as the arithmetic predicts; and a **prospective prediction** is recorded in §5
BEFORE the pass that tests it. A `090` run is welcome and would be cheap — the fixing lane should fire
it if any of §2 reads as inference.

**This is NOT `bugs_open/316`.** 316 is *who wins when more sites are due than the cap admits*
(alphabetical ordering; fixed by migration `556`, live). This is *nobody being due*: the cap was not
binding in the measured pass (10 due, 10 dispatched, `LIMIT 10`). 316's fix is in the query quoted
below and this defect survives it. The `bugfix_316_news_feed_ordering` lane's README (line 138,
2026-08-22) says the seven 6-hourly sites are *"now fully served"*: **refuted by §3** — they are
served every other run, exactly as before, for a different reason.

---

## 1. Mechanism in one paragraph

`scheduled_tasks.content-feed-refresh` fires `content-feed-trigger` every 21,600 s. Its
`find_news_sites` step selects a site only if some active source has `next_fetch_at IS NULL OR
next_fetch_at <= NOW()`. Both writers of `next_fetch_at` stamp it **relative to the moment of the
fetch**, not the moment of the trigger: `NOW() + fetch_interval`. A fetch happens 10 s – 9 min after
its trigger (dispatch is sequential, ~50 s per site). With `fetch_interval` = **6 h** (the column
default, and every source on 12 of 12 news sites), a source fetched at *T + δ* is due at *T + 6h + δ*,
and the next trigger fires at *T + 6h + ε* with ε ≈ 3–30 s of scheduler drift. δ > ε for every site
but (at best) the first one dispatched in the first seconds, so **the source is not due, the site is
skipped, and it is served at T + 12h.** Sites that ALSO hold a 3 h or 4 h source are due at every
pass and are served every pass — the control.

## 2. Evidence `[ALL MEASURED 2026-08-26 14:15–14:35Z unless dated]`

**The trigger's live site-selection query** (`agent_definitions` type `content-feed-trigger`,
`default_config->workflow->steps->find_news_sites->config->query`, post-316):

```sql
... EXISTS (SELECT 1 FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true
            AND (cs.next_fetch_at IS NULL OR cs.next_fetch_at <= NOW()))
... ORDER BY due_at ASC NULLS LAST, domain ASC LIMIT 10
```

**The two stamp arms**, both `NOW() + fetch_interval`:

- `platform/orchestration/actions/dispatch_feed_sources_action.go:272-279` — *"Optimistically update
  next_fetch_at to prevent re-dispatch before completion"*: `SET next_fetch_at = NOW() + fetch_interval`.
- `platform/orchestration/actions/feed_actions.go` `UpdateSourceTimestampsAction` success arm (the
  `UPDATE content_sources SET last_fetched_at = NOW(), … next_fetch_at = NOW() + fetch_interval`
  block; the failure arm multiplies the interval by `LEAST(error_count+1, 4)` — same anchor).
- The per-site source selector the orchestrator itself uses is the same shape:
  `feed_actions.go:962` *"Returns all active sources for a site where next_fetch_at <= NOW()"*, :1007.

**The cadence and its drift** (`orchestration_states`, `owner_agent_type='content-feed-trigger'`):

| fired (UTC) | gap |
|---|---|
| 08-25 14:45:19 | |
| 08-25 20:45:32 | 6h 00m 12s |
| 08-26 02:45:35 | 6h 00m 04s |
| 08-26 08:46:06 | 6h 00m 30s |

**The interval**: `information_schema.columns` default for `content_sources.fetch_interval` =
`'06:00:00'`; every active source on all 12 news-eligible sites carries `06:00:00`; dartsonline adds a
`04:00:00`, relojistas a `03:00:00` and a `04:00:00` (`SELECT domain, string_agg(DISTINCT
fetch_interval) … GROUP BY domain`).

**The worked case — idea.uk at the 08:46:06 pass.** Its five sources were fetched 02:46:15–02:46:31
(dispatched 02:45:52, 17 s after the 02:45:35 trigger); `next_fetch_at` = **08:46:15 – 08:46:31**;
the trigger fired at **08:46:06** — **9 to 25 seconds too early**. `orchestration_states` holds ONE
`content-feed-orchestrator` run for the site today (02:45:52). Not the cap: exactly 10 sites were due
(the seven fetched at 20:45–21:02 on 08-25 and skipped at 02:45 for this same reason; the two
always-due sites; one newly armed site) and exactly 10 were dispatched, 08:46:20 → 08:54:23.

## 3. The 48-hour census — every 6h-only site runs 12 h apart; the sub-6h sites run every pass

`SELECT s.domain, string_agg(DISTINCT cs.fetch_interval::text,','), string_agg(to_char(o.created_at,'DD HH24:MI'),' | ' ORDER BY o.created_at) FROM orchestration_states o JOIN sites s ON s.id=o.site_id LEFT JOIN content_sources cs ON cs.site_id=s.id AND cs.is_active WHERE o.owner_agent_type='content-feed-orchestrator' AND o.created_at > now()-interval '48 hours' GROUP BY 1;` (distinct times, de-duplicated by hand):

| site | intervals | orchestrator runs (UTC) | effective cadence |
|---|---|---|---|
| ai-agent-orchestration.com | 6h | 25 20:47 · 26 08:47 | **12 h** |
| fundamentallyai.com | 6h | 25 20:49 · 26 08:48 | **12 h** |
| gaswholesalers.com | 6h | 25 20:58 · 26 08:51 | **12 h** |
| mortgagecalculator.co.uk | 6h | 25 21:00 · 26 08:52 | **12 h** |
| robot-hands.com | 6h | 25 20:51 · 26 08:48 | **12 h** |
| webdesign.co.uk | 6h | 25 20:45 · 26 08:46 | **12 h** |
| idea.uk | 6h | 25 16:25/16:29 (manual) · 26 02:45 — skipped 20:45 and 08:46 | **12 h** |
| remortgagecalculator.uk | 6h | 26 02:48 · 26 13:43 (off-cadence run) | — |
| loanandmortgagecalculator.co.uk | 6h | 26 08:54 (first run) | — |
| vetcomparison.uk | 6h | 25 21:02 · 25 23:00 · 26 08:53 | mixed |
| **dartsonline.com** | 4h + 6h | 25 14:47 · 20:56 · 26 02:47 · 08:50 | **6 h** ← control |
| **relojistas.com** | 3h + 4h + 6h | 25 14:45 · 20:54 · 26 02:46 · 08:49 | **6 h** ← control |

The disconfirming result would have been a 6h-only site appearing at ~02:4x AND ~08:4x; none does.
The control would have failed if the 3h/4h sites also alternated; they do not.

## 4. Blast radius and why nothing surfaced it

- **10 of 12 news sites** (as of 2026-08-26) hold only 6-hourly sources and refresh **twice a day, not
  four times** — half the designed and documented cadence (`scheduled_tasks.description`: *"every 6
  hours to refresh news feeds"*), since the trigger was armed.
- Nothing fails, so nothing files: the 6-hourly `stale_news_section` check keys on newest-item age vs
  `max_age_hours` (default **72**), which a 12 h cadence never approaches; the trigger's own run is
  COMPLETED; every orchestrator run that does happen is COMPLETED.
- The 316 lane measured "late" against each site's `fetch_interval` and read the cap boundary as the
  cause. Their fix made the queue fair (true) and did not change the cadence (this file).
- Aside, not the defect: dispatch is sequential at ~50 s per site (08:46:20 → 08:54:23 for 10), which
  is what makes δ minutes rather than seconds.

## 5. Prospective test — written at 14:38Z, BEFORE the ~14:46Z pass

idea.uk has been due since 08:46:15. Prediction: **(a)** the ~14:46 pass dispatches it (an
`orchestration_states` row, `owner_agent_type='content-feed-orchestrator'`, `site_id` idea.uk,
`created_at` ≈ 14:46–14:55); **(b)** its five `next_fetch_at` land at (fetch time + 6 h) ≈
20:46–20:55; **(c)** the ~20:46 pass does NOT dispatch it; **(d)** the ~02:46 pass on 08-27 does.
Refutation of (c) — an idea.uk orchestrator row at ~20:46–20:55 — kills this file's mechanism.
Outcome to be recorded below by whoever is watching (the filing lane will record (a) and (b)).

## 6. Fix candidates — ordered by what makes the bad state unrepresentable

1. **Give the due predicate a look-ahead of half the cadence, in BOTH layers, in one migration.**
   `next_fetch_at <= NOW() + interval '3 hours'` in `find_news_sites` (config, DB-live on apply) AND
   in the orchestrator's per-site source selector (`feed_actions.go:1007` — Go, so this half needs a
   roll; until it rolls the trigger would dispatch a site whose orchestrator then finds 0 due sources
   and completes as `no_sources`). A source due at any point before the midpoint to the next pass is
   served now; worst-case lateness falls from 6 h to ≤ 3 h and the phase lock cannot form. **The two
   layers are the trap** — 316 §"two layers" already documents that the site-level query and the
   source-level query drift apart.
2. **Anchor the stamp to the schedule, not the fetch**: `next_fetch_at = <this pass's trigger time> +
   fetch_interval`, carried into the ingester's payload. Structurally closes the door for any
   interval ≥ cadence, but the ingester has no notion of the trigger time today (Go, two arms, roll).
3. **Set `fetch_interval` below the cadence** — column default and seeder to `05:30:00`, and an
   `UPDATE` for the 12 existing sites. Cheapest, DB-live, and the least closing: it re-opens the moment
   anyone sets an interval equal to the cadence again, which is the obvious value to set.

**Per-site mitigation (candidate 3, one site) — left for the OWNER**: the filing session's attempt
was refused by its permission classifier (a production `UPDATE`), which is the right refusal — it
doubles that site's search-fetch spend and the owner has a cost-scare history on rotations.
```sql
CREATE TABLE bak_ideauk_fetch_interval_20260826 AS SELECT * FROM content_sources WHERE site_id='1244516d-014d-421c-88c6-090bb1e9552a';
UPDATE content_sources SET fetch_interval='05:30:00' WHERE site_id='1244516d-014d-421c-88c6-090bb1e9552a' AND is_active AND fetch_interval='06:00:00';
```

## 7. How to verify a fix

The §3 census, re-run 48 h after the roll: every 6h-only site shows **four** distinct run hours per
day (≈ 02:4x, 08:4x, 14:4x, 20:4x). Per site, `max(last_fetched_at)` should never be > 6 h 15 m old
while the trigger is running. Control stays: dartsonline/relojistas unchanged at every pass.

## 8. Consumers told / pointers

`bugs_open/316` (CONTRIB appended today, same commit) · `bugfix_316_news_feed_ordering/` (its
"fully served" claim, README:138) · `news_feed_pooling/` · `seed_content_sources_action.go` (inherits
the 6 h column default) · 016b §9 entry "a due-stamp of fetch-time + period is phase-locked to the
scheduler" (same commit) · idea.uk lane RUNBOOK 6g.

### §5 outcomes, recorded as they land

- **(a) CONFIRMED 2026-08-26 14:47Z** — trigger fired **14:46:32** (gap from 08:46:06 = 6h 00m 26s,
  the drift as measured); idea.uk `content-feed-orchestrator` run created **14:46:58**, the site's
  second of the day.
- **(b) CONFIRMED 2026-08-26 ~14:48Z** — the five `next_fetch_at` moved to **20:47:01–20:47:06**
  (the dispatch arm's optimistic stamp) and then **20:47:24–20:47:42** (the ingestion arm's), both
  AFTER the next trigger's expected ~20:46:58–20:47:02. Margin this cycle: **~25–45 s** (morning
  cycle was 9–25 s) — the phase lock re-arms itself every pass, δ > ε again.
- **(c)/(d) pending** — the ~20:46–20:47 pass should SKIP idea.uk (an idea.uk orchestrator row
  there refutes this file); the ~02:46 pass on 08-27 should serve it.

---

## 2026-08-26 ~18:00Z — PICKED UP by the `bugfix_410_feed_phase_lock` lane; fix built (candidate 1); two corrections

Lane docs: `docs/agent_docs/docs024_key_docs_latest/bugfix_410_feed_phase_lock/`. Handoff
confirmed by message with the filing session. ⚠ Number collision: 410 ALSO names an unrelated
scan-loss case (its own lane, `bugfix_410_silent_scan_loss`) — resolve by slug.

**Correction 1 — the live second layer is `dispatch_feed_sources_action.go`'s own due query,
not `feed_actions.go`'s.** §2's "per-site source selector the orchestrator itself uses"
(`feed_actions.go:962/:1007`, `LoadDueSourcesAction`) has **zero live workflow callers**
(verified against `agent_definitions` steps 2026-08-26: the orchestrator's step is
`dispatch_sources` → `dispatch_feed_sources`). Same predicate text, so the mechanism and all
arithmetic stand unchanged — but a fix applied where §6 pointed would have patched dead code.

**Correction 2 — the §3 controls are only controls at SITE level.** dartsonline's 6 h sources
(last fetched 08:51Z, due 14:51:09–28Z) were skipped by its OWN 14:50:27Z dispatch by
~40–60 s while its 4 h sources were fetched: the phase lock also operates per-SOURCE inside
an admitted site. "dartsonline **6 h** effective cadence" in §3 is true of the site's runs
and NOT of its 6 h sources, which run 12-hourly like everyone else's.

**Fix shipped (this commit): the half-cadence due look-ahead — candidate 1, both layers.**
Due test becomes `next_fetch_at <= NOW() + COALESCE((SELECT make_interval(secs =>
interval_seconds / 2.0) FROM scheduled_tasks WHERE name = 'content-feed-refresh'), interval
'3 hours')` — serve on the nearest grid tick, not the tick after; cadence read live;
fallback = today's half-cadence, never bare `NOW()`. One shared Go constant
(`feed_due_lookahead.go`) feeds both Go readers (dispatcher + the callerless
`LoadDueSourcesAction`, fixed so a future caller inherits it); migration
`653_content_feed_due_lookahead_HOLD.sql` is the site-admission half — **held, hand-applied
only after the chassis roll** (config-first admits sites the old binary refuses: no-op runs
burning cap slots). Guards: cadence-still-21600 and 556-post-image pre-image match. Tests
mutation-proven (three mutations RED, restored GREEN — lane NOTES). Candidates 2/3 not
taken: 2 needs trigger-time plumbed through both stamp arms for the same closure; 3 re-opens
at the next default-valued source. Caps 10/10 untouched (owner capacity decision, 556);
post-fix ~12 sites due per pass means **cap hits become normal** — expected demand, not a
regression — and fetch volume returns to designed (~2× today's). Council submission
alongside this commit; 090 independent run fired 17:31Z
(RUN_CORRELATION_ID `15d56c13-2081-431a-ad70-9516c5fcfbc7`).

**§5 (c) criterion, refined before the ~20:47Z pass tests it** (per the filing lane, by
message): the sources are due 20:47:24–42Z and trigger drift is seconds — so read the
trigger's FIRE TIME first. (c) = fired **before 20:47:24Z** AND no idea.uk orchestrator row
→ mechanism confirmed. If the trigger happens to drift PAST 20:47:42Z, a dispatch would
CONFIRM the arithmetic (margin flipped), not refute it. This lane records (c)/(d).

**2026-08-26 ~18:4xZ — council APPROVED round 1** (corr `04c657d2`, 1 advisory objection —
the "zero callers of LoadDueSourcesAction" absence claim lacked its query in evidence; re-run
as a full config-text census incl. nested steps and snapshots: still 0 rows, recorded with
the queries in the lane NOTES). Awaiting: the chassis roll (Go half), then hand-apply of 653;
090 verdict (corr `15d56c13`); §5 (c)/(d) tonight.

---

## 2026-08-26 ~21:00Z — BOTH HALVES LIVE. §5 (c) CONFIRMED. The 090 loop returned UNVERIFIABLE.

**§5 (c) CONFIRMED, on the refined criterion.** The evening trigger fired **20:46:45Z** —
**39 s before** idea.uk's earliest due stamp (20:47:24) — and **no idea.uk
`content-feed-orchestrator` row exists for that pass**. Fired-before-the-window AND not
dispatched is exactly the predicted skip. The same pass dispatched **webdesign.co.uk 20:47:02,
ai-agent-orchestration.com 20:48:26, fundamentallyai.com 20:50:31** — each due since ~14:47,
i.e. each *skipped by the 14:46:32 pass*: the phase lock caught in the act one final time.

⚠ **(d) IS NOT A DISCRIMINATING TEST and must not be reported as one.** idea.uk has been due
since 20:47:24, so the ~02:46Z pass will dispatch it whatever the predicate says. The real
acceptance test replaces it: **every site dispatched in the 20:47Z pass must be dispatched
AGAIN at ~02:46Z** (their stamps are ≈02:47–02:56; under the old rule each misses by seconds
and waits until ~08:46). Query and disconfirming result: lane HANDOFF §4.

**Deployment, verified at the artefact.** Go half **LIVE on `v1.0.1345`, both replicas** —
the `build provenance` line had already scrolled (empty grep = *not in range*, not
*unstamped*), so proven by binary probe with both controls: the look-ahead SQL literal present
**2** (the two readers), a near-miss variant **0**, a known existing literal **1**. Config half:
migration **653 applied ~20:52Z**, both guards passed, and re-read independently afterwards —
look-ahead present, 554 ordering + 556 caps intact, bare `NOW()` gone. Pre-change snapshot
`51dd1c59-69e6-4625-baf6-203c35052f18`. **653's guard 2 passing also settles the council's
low-severity objection empirically**: the live query *was* byte-identical to 556's post-image.

**The 090 run did NOT confirm this file — recorded plainly rather than quietly dropped.**
Corr `15d56c13-2081-431a-ad70-9516c5fcfbc7` returned **`UNVERIFIABLE` — "Diagnosis NOT
confirmed (stopped: scope-not-narrowing)"**: two evidence bundles, no iteration note, no
verdict artifact. **That is neither a confirmation nor a refutation** — the loop reached no
conclusion. This file's standing claim therefore rests where §"090 substitution" said it did:
first-hand verification plus the prospective predictions, (a)/(b)/(c) all now confirmed live,
and council APPROVED at round 1 (`04c657d2`). `[INFERRED]` cause of the stall: the symptom was
submitted as a fully-formed conclusion, against this repo's own symptom-authoring guidance —
logged in `WRONG_CALLS.md`.

**Bug stays OPEN** until the ~02:46Z acceptance test passes and §7's 48 h census shows four
run-hours/day for every 6h-only site.

**2026-08-26 21:02Z — the acceptance test set, pinned before `orchestration_states` prunes it.**
The 20:46:45Z pass (the last under bare-NOW() admission) served, with each site's re-stamped
earliest `next_fetch_at`: webdesign.co.uk 02:47:31 · ai-agent-orchestration.com 02:48:48 ·
fundamentallyai.com 02:50:54 · robot-hands.com 02:53:26 · gaswholesalers.com 02:59:35 ·
mortgagecalculator.co.uk ≈03:01 — **six 6h-only sites, every one re-stamped AFTER the expected
~02:46:5xZ trigger** (the defect) **and inside the 3 h look-ahead** (the fix). All six must be
served at ~02:46Z. ⚠ relojistas.com and dartsonline.com were also served but are **NOT**
evidence: their 3 h/4 h sources come due at 23:55 and 00:57, i.e. before the trigger, so they
pass under either predicate — the same vacuous shape as prediction (d). Full table: lane
HANDOFF §4.

---

## 2026-09-02 — CLOSED. §7's first criterion met on live traffic; the second reassigned to `bugs_open/316`

`[ALL MEASURED 2026-09-02 12:38–13:00Z, DB clock]` by the `bugfix_410_feed_phase_lock` lane.
Full record: that lane's `NOTES` (queries + traps) and `README_where_we_are` (owner prose);
milestone read-out in `SUMMARY_2026-09-02_410_feed_phase_lock.md`.

### The fix is confirmed working on real traffic

The §4 acceptance test named a pass on 2026-08-27; `orchestration_states` prunes at ~2 days, so
it was gone. **It was also unsatisfiable as written** — see the correction below. Replaced with a
test that needs no inference: a site fetched in pass N cannot have been fetched before it was
*dispatched* in pass N, so its next due stamp is **≥ (pass-N dispatch) + 6 h**. If pass N+1's
trigger fires before that bound and the site is admitted, `next_fetch_at <= NOW()` cannot explain
it. Pass N fired 02:57:57, pass N+1 fired **08:58:27**:

| site (6h-only) | dispatched pass N | earliest possible due | trigger fired EARLY by | pass N+1 |
|---|---|---|---|---|
| remortgagecalculator.uk | 03:00:42 | 09:00:42 | 2 m 15 s | **SERVED 09:14:20** |
| idea.uk | 03:03:25 | 09:03:25 | 4 m 58 s | **SERVED 09:16:56** |
| vetcomparison.uk | 03:09:24 | 09:09:24 | 10 m 57 s | **SERVED 09:18:35** |

Three sites served on a tick that fired minutes before they could possibly have been due —
arithmetically impossible under the pre-fix predicate. **`idea.uk`, this file's own site, skipped
by 39 s on 08-26, is one of them.** `mortgagecalculator.co.uk` also reappeared but its bound falls
**4 s the wrong side**, so it is served under either predicate: **excluded as vacuous**, the same
shape as prediction (d). Straggler-source objection closed: `off_pattern`=0, `max_err`=0, source
spread ≤19 s on all four, and **vetcomparison.uk has exactly one source**, so it is decisive alone.

Both halves re-probed rather than assumed: chassis **v1.0.1352** (rolled on from 1345) carries the
look-ahead — capability probe **2** hits, negative control **0**, positive control **1**; the live
`find_news_sites` still carries the look-ahead with 554+556's tail intact and no bare `NOW()` arm;
cadence still 21600 s, so the `interval '3 hours'` fallback is still the designed value.

### ⚠ CORRECTION to §4 of the 08-26 handoff — the acceptance test was unsatisfiable when written

§4 required *every* site from one pass to reappear in the next, while §6 residual 5 — same
document — predicted the fix would put ~12 sites against a cap of 10 and that cap hits would
become routine. Both cannot hold. Measured: **4 of 8** discriminating sites were capped out of the
very pass that proved the fix. A pass-membership test cannot separate "phase-locked" from
"displaced by the cap". `WRONG_CALLS.md` row + LANDMINE filed; the RUNBOOK now carries the
lower-bound test instead. Also **the trigger drifted ~:46 → ~:57 in six days** — never reuse a
handoff's hardcoded window.

### §7's second criterion is NOT met, and it belongs to `bugs_open/316`

*Four run-hours/day per 6h-only site* is unreachable at the current cap. `[MEASURED 2026-09-02]`
**14** eligible news sites, **12** of them 6h-only; the two multi-interval controls are due every
pass and take 2 of the 10 slots, leaving **8 slots for 12 sites**. Over the three successful
passes, 36 demanded / 24 served — **exactly 24 observed**, spread evenly (max−min = 1), so the cap
is fully binding and 554's `due_at` rotation is working. Effective cadence **≈9 h** (was 12 h,
designed 6 h). Four sites were carrying 9 h 24 m–9 h 32 m staleness.

That is 316's subject verbatim (*"the queue is 2× oversubscribed"*); its 556 fix made the queue
fair and deliberately left the cap as an owner capacity decision. **CONTRIB filed there with these
numbers. This bug is not held open on another bug's condition** — the phase lock is fixed, live and
no longer reproducible, which is the stated bar. ⚠ The cap is a **literal `10`** while the eligible
count grows by addition: residual 5 sized it "~12" on 08-26 and it is **14** today.

### Residuals at close

1. Go↔config parity still unenforced against live config on a schedule. **Not built.**
2. `interval '3 hours'` fallback — cadence unchanged at 21600 s, so still correct. Landmine stands.
3. `LoadDueSourcesAction` still callerless — re-verified by full config-text census (**0** rows).
4. **Provocation twin: RESOLVED — NOT affected.** `provocation-feed-publisher`'s entire live
   workflow is `publish_feed` → `render_provocation_feed` → `complete`; it holds no per-item due
   stamp (`next_fetch_at` appears nowhere outside content-feed paths), writing only
   `scheduled_tasks.last_completed_at`. No due predicate ⇒ no phase lock possible.
5. Cost/capacity — **owner's decision, untouched.** Raising the cap to ≥14 buys the designed 6 h.
6. ⚠ **A stalled pass mimics this bug exactly.** 09-01 20:57:41 FAILED on
   `process_sites_iter_1_spawn_orchestrator` (`reaper: stale EXECUTING_STEP for >4h`), served **1**
   site and cost 13 sites a whole pass — a 12 h gap with nothing to do with the phase lock. Known
   spawn→call handshake race, own owner. **Anyone re-measuring feed cadence will meet this and may
   re-diagnose 410.** Excluded from every count above.
