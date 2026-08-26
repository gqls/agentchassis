// FILE: platform/orchestration/actions/feed_due_lookahead.go
//
// The due predicate for content_sources, with a look-ahead of HALF the
// content-feed trigger's cadence (bugs_open/410, the next_fetch_at phase lock).
//
// WHY A LOOK-AHEAD EXISTS AT ALL. Both writers of next_fetch_at stamp it
// relative to the moment of the FETCH (NOW() + fetch_interval), while the
// trigger fires on a fixed grid (scheduled_tasks.content-feed-refresh, every
// interval_seconds). A fetch happens seconds-to-minutes after its trigger
// (dispatch is sequential), so a source whose fetch_interval EQUALS the trigger
// cadence comes due seconds AFTER the next pass fires — it is skipped every
// other pass and runs at half its labelled cadence, for ever. Measured
// 2026-08-26: 10 of 12 news sites on a 12 h cadence under a 6 h label.
//
// THE RULE THE LOOK-AHEAD ENCODES: serve a source on the grid tick NEAREST its
// due time, not the first tick strictly after it. With look-ahead = cadence/2,
// a source is fetched at most half a cadence early, the error does not
// accumulate (next_fetch_at is re-anchored at each fetch), and an interval
// larger than the cadence still waits its full multiple of ticks. The phase
// lock cannot form because the seconds-scale gap between "stamped due" and
// "trigger fired" is dwarfed by the half-cadence window.
//
// THE CADENCE IS READ LIVE from scheduled_tasks so a capacity change there
// propagates here without a code change. The COALESCE fallback is HALF OF THE
// CADENCE AT THE TIME OF WRITING (21600 s / 2, i.e. 3 hours — asserted against
// the live row by migration 653's guard): if the task row is renamed or
// deleted the predicate degrades to today's designed value, never to the bare
// NOW() that produced the phase lock.
//
// ⚠ THIS PREDICATE HAS A CONFIG TWIN. The site-level selection in
// content-feed-trigger.find_news_sites (agent_definitions, migration 653)
// carries the same look-ahead and MUST stay in step: config-only admits sites
// whose sources the dispatcher then refuses (no-op runs that consume cap
// slots); Go-only is unreachable because the site is never admitted. See
// LANDMINES.md (footprint: this file) and bugs_open/410.
package actions

// feedDueLookaheadSQL is the look-ahead interval, evaluated by the database at
// query time: half the live trigger cadence, falling back to 3 hours (half of
// the cadence this was written against) if the task row is missing.
const feedDueLookaheadSQL = `COALESCE((SELECT make_interval(secs => interval_seconds / 2.0) FROM scheduled_tasks WHERE name = 'content-feed-refresh'), interval '3 hours')`

// feedSourceDuePredicate is the one due test for content_sources rows. Both Go
// readers (DispatchFeedSourcesAction, LoadDueSourcesAction) embed this constant
// rather than restating it — the two queries drifting apart is exactly how
// bugs_open/316 documented this subsystem going wrong before.
const feedSourceDuePredicate = `(next_fetch_at IS NULL OR next_fetch_at <= NOW() + ` + feedDueLookaheadSQL + `)`
