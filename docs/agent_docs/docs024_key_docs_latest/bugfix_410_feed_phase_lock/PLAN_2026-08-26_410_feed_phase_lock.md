# PLAN — bugs_open/410 (phase lock): the half-cadence due look-ahead

Lane: `bugfix_410_feed_phase_lock` · started 2026-08-26 ~17:30Z · fixing thread for
`bugs_open/410_HANDOFF_2026-08-26_next_fetch_at_stamped_at_fetch_time_phase_locks_every_six_hour_news_site_to_a_twelve_hour_cadence.md`
(filed by the `idea_uk_vm_site` lane; handed off explicitly — confirmed by message with that
session 2026-08-26 17:3xZ).

⚠ **410 names TWO unrelated bugs.** The other (three-seams scan loss) is owned by another
session with its own lane (`bugfix_410_silent_scan_loss`); confirmed by message, no overlap.
Resolve by slug, always.

## The defect, one paragraph

Both writers of `content_sources.next_fetch_at` stamp `NOW() + fetch_interval` at **fetch**
time; the trigger fires on a fixed ~6 h grid. Fetch lags trigger by dispatch latency
(10 s–9 min), the grid drifts only seconds, so a 6 h-interval source is due seconds AFTER
the next pass and is served every other pass: a 12 h cadence under a 6 h label, all runs
COMPLETED. 10 of 12 news sites, since arming (bug file §2–3, prospectively confirmed).

## Decision: candidate 1 (look-ahead), and why not 2 or 3

**Chosen: due look-ahead of HALF the trigger cadence, in every layer that asks "is this
source due?"** — semantics: *serve on the nearest grid tick, not the first tick strictly
after*. Worst-case earliness cadence/2, error non-accumulating (the stamp re-anchors each
fetch), intervals > cadence still wait their full multiple of ticks, and the phase lock
cannot form because seconds-scale stamp lag is dwarfed by the 3 h window.

- **Not candidate 2 (anchor stamps to the schedule):** needs the trigger's fire time plumbed
  into the ingester's payload and coordinated across BOTH stamp arms (optimistic + ingester),
  plus a catch-up rule for long-dead sources. More moving parts for the same closure; the
  look-ahead achieves unrepresentability with stamps untouched.
- **Not candidate 3 (intervals below cadence):** re-opens the moment anyone sets interval =
  cadence, which is the column default and the obvious value. "Operators must remember X" is
  a defect (register: order-fix-candidates-by-what-closes-the-door).

**The look-ahead reads the cadence live** from `scheduled_tasks.interval_seconds`
(`name='content-feed-refresh'`), so a capacity/cadence change propagates with no code change;
`COALESCE` falls back to `interval '3 hours'` (half of today's 21600 s — migration 653's
guard 1 asserts that equality at apply time) so a renamed task degrades to the designed
value, never to bare `NOW()`. The fallback direction was chosen deliberately: the quiet
default must be the FIXED behaviour, not the defect (the other 410's whole thesis).

## Where it lands (3 code surfaces + 1 config)

| layer | file | live when |
|---|---|---|
| shared predicate | `platform/orchestration/actions/feed_due_lookahead.go` (new: `feedDueLookaheadSQL`, `feedSourceDuePredicate`) | chassis roll |
| source-level, LIVE path | `dispatch_feed_sources_action.go` due query (was bare `NOW()`) | chassis roll |
| source-level, dormant path | `feed_actions.go` `LoadDueSourcesAction` (**no live workflow caller as of 2026-08-26** — fixed so a future caller inherits it) | chassis roll |
| site-level admission | migration `653_content_feed_due_lookahead_HOLD.sql` → `find_news_sites` config query | hand-applied AFTER the roll |

**CORRECTION to the bug file (recorded there too):** its "second layer" citation
(`feed_actions.go:962/:1007`) is the dormant `LoadDueSourcesAction`; the live second layer is
`dispatch_feed_sources_action.go`'s own due query. Both now share one predicate constant, so
the distinction can no longer drift into a defect.

**Also strengthened at the source level:** the census shows even the "control" sites'
6 h SOURCES were phase-locked *inside* admitted sites — dartsonline's 6 h sources
(last fetched 08:51, due 14:51:09–28) missed the 14:50:27 dispatch by ~40–60 s while its 4 h
sources were served. Site admission alone would not have fixed those; the source-level
half is load-bearing, not belt-and-braces.

## Sequencing (the trap idea.uk flagged, adopted)

Go first (commit → next chassis roll), migration 653 by hand only after the roll's build
provenance confirms the commit. Config-first admits sites the un-rolled dispatcher then
refuses → no-op COMPLETED runs consuming cap slots and poisoning the census. Go-only is a
no-op (admission still gates on bare NOW()). Neither half alone regresses anything — the
failure mode of partial deployment is "no change yet", not damage.

## What this deliberately does NOT change

- **Caps stay 10/10** (migration 556 — an owner capacity decision). Post-fix ~12 sites are
  due every pass, so cap hits become the norm; 554's due_at fair ordering bounds the
  overflow (avg ≈ 7.2 h vs the 6 h label for 12 sites over cap 10). Raising to 12 is the
  owner's cost call, flagged in the council submission, not taken here.
- **Stamps and the backoff arm unchanged** (see candidate 2 above; backoff stays
  now-anchored, which is right for a failing source).
- **COST, stated plainly:** restoring the designed cadence roughly **doubles** feed
  ingestion volume (~10 sites × ~5 sources go from 2 to 4 fetches/day ≈ +100 ingester
  runs/day). This is the spend the 6 h design always implied; the owner has cost
  sensitivity, so it is flagged, loudly, rather than assumed.

## Verification

1. Mutation-proven tests (all three run RED then restored GREEN pre-commit — NOTES).
2. `090` independent diagnosis fired 17:31Z, RUN_CORRELATION_ID `15d56c13-2081-431a-ad70-9516c5fcfbc7` — read verdict before council submission if landed, else before closing.
3. Prospective (c)/(d) from the bug file §5: recorded tonight (~20:47Z) and tomorrow (~02:46Z),
   with idea.uk's refinement — read the trigger's actual fire time first; skip is predicted
   only if it fires before 20:47:24Z.
4. Post-roll+apply: bug file §7 census — every 6 h-only site at four run-hours/day;
   `max(last_fetched_at)` never > 6 h 15 m stale; exclude remortgagecalculator.uk's
   off-cadence 13:43Z run from any before/after comparison.

## Residuals (stated so they don't become quiet defaults)

1. Go↔config predicate parity is pinned at commit time (shared const + tests + migration
   guard) but NOT enforced against live config on a schedule. A future migration could
   rewrite `find_news_sites` without the look-ahead and nothing fires. Candidate follow-up: a
   `cmd/config-key-audit` mode asserting any config query with a `next_fetch_at` due arm
   carries the look-ahead. Not taken in this commit (scope).
2. `LoadDueSourcesAction` remains callerless — fixed and guarded, but dead code is dead code.
3. The provocation-feed twin (`provocation-feed-refresh`, also 21600 s) was NOT audited for
   the same class — its publisher may or may not stamp fetch-relative due times. Follow-up
   candidate for whoever owns it.
