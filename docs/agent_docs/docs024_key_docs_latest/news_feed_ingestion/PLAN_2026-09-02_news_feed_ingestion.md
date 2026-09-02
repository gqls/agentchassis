# PLAN — news_feed_ingestion

Started 2026-09-02. Charter proposed by the `calendar` session (peer handoff, in
chat) after it filed `bugs_open/427`, adopted here after independent verification
(read 427 in full, ran `who-owns.py` on 427/316, checked the tree for any in-flight
feed work — none found). Session named "feed lane" by the owner before the peer's
message arrived, which is why this charter was accepted rather than deferred.

**Charter**: the raw news-ingestion pipeline — source enrollment, fetch
scheduling/cadence/queue fairness, and normalising/extracting structure out of what
arrives (topics, credibility, dedup, and now structured entities/events).
Concretely: `content-feed-orchestrator`, `find_news_sites`/`content-feed-trigger`,
`feed-triage`, `content-feed-refresh`, and `content_feed_items` as a data asset.

**Explicitly NOT this lane**: `news_editorial_features` (what gets WRITTEN from
feed items — editorial prose; declined 427's ask correctly, authoring isn't
ingestion), `period-calendar` (recurring-NAME calendars, not dated one-off events —
`calendar_component` lane), `research-agent`/`evidence-researcher` (one-time
landscape research, different pipeline), site-planner/page-role work (`bugs_open/206`
and diagnosis correlation `d6d350ec-e16b-4792-9282-ca5155369791`, status
UNVERIFIABLE — don't duplicate, just don't let it block this lane's half).

## Priority order (from the peer handoff, verified before accepting)

1. **`bugs_open/427`** — no writer turns a confirmed news item into a dated,
   structured `evidence_base` fact. Verified: file reads coherently, its own header
   says "Status: OPEN, unowned"; `who-owns.py 427` shows the bug file itself owned
   by `calendar_component` (who filed it) but §7 explicitly disclaims the fix as
   "not this lane's build" for the calendar side. Severity MEDIUM-HIGH: owner
   ruling is fix-before-delivery on a paid site (boxingonline.com), so this is
   blocking, not queued. **Working this first.**
2. **`bugs_open/316`** (+ two riders parked in that lane: unenrolled sites never
   reaching the seeder, and UK-news-defaults-on-`.uk` absence) — after 427.
   `who-owns.py 316` shows the `bugfix_316_news_feed_ordering` workstream as
   "ACTIVE" by commit-recency, but the real-work commits cluster on 2026-08-22;
   everything from 08-25 onward is CONTRIB-shaped (other lanes routing findings
   in) or an owner routing commit, and the tree has no dirty files touching it —
   consistent with the peer's "genuinely stale" read. Will re-verify before
   touching.
3. **Not urgent, long-term background**: `news_feed_pooling` (owner-parked since
   2026-07-20), and reading (not reworking) both `bugfix_410_feed_phase_lock`
   variants for reusable patterns (410's cadence/staleness handling is the direct
   analogue for 427 fix candidate #2).

## 427 — design for fix candidate #1 (the one this lane owns)

Bug 427 §7 candidate #1: "a new extraction step, downstream of `feed-triage`, that
fills `entity_ids` (or a comparable typed field) for a confirmed-event item... and a
corresponding writer that registers it as a dated `evidence_base` fact."

**Correction to the literal reading, before building anything — checked, not
assumed.** `entity_ids` is `uuid[]` with **no FK constraint** (`\d content_feed_items`,
2026-09-02) and no documented target table. `directory_entities` exists but is a
different, unrelated feature (business/model directory claims, kind+slug keyed).
The `news_editorial_features` lane already hit this and left a LANDMINE
(`LANDMINES.md`, "entity_ids is an empty column with no writer... establish what it
was for before designing a grouping key"). Nobody today can state what it was for —
the original design doc (`006_news_feed_pipeline_v2.md:138`) declares it with zero
surrounding prose. **Decision: do not populate `entity_ids` for this fix.** Bug 427's
own wording anticipates this ("or a comparable typed field") — the store this fix
actually writes to is `evidence_base` (per §6: "do not invent a parallel store"), and
`entity_ids`/dedup-grouping is a distinct, still-open question this fix does not
need to answer. If a real "entities" concept gets built later or `entity_ids`' intent
resurfaces, revisit; not blocking today's fix.

**Design: extend two things that already exist, rather than build a parallel
pipeline** (platform idiom — reuse before building new; this is exactly the
`verify_and_register_directory_claims` / `verify_and_register_citations` idiom
applied a third time; see `directory_claims.go`'s own header: "This file adds
NOTHING to how a citation is verified... reused UNCHANGED").

1. **New Go action `load_feed_items_for_event_extraction`** — same shape as
   `LoadFeedItemsForTriageAction` (`feed_triage_actions.go:320`) but
   `WHERE status = 'relevant' AND event_extracted_at IS NULL`, ordered by
   `source_published_at DESC`. Returns items with topics included (extraction
   prompt can gate on topic shape).
2. **New Go action `mark_feed_items_event_extracted`** — sets
   `event_extracted_at = now()` on every item id considered by the extraction
   pass, whether or not it yielded a candidate fact. This is the idempotency
   mechanism (not a new "processed" status — `status` stays `relevant`, so the
   render path and everything reading `status` is unaffected). Without this,
   every triage cycle would re-spend LLM budget re-examining the same non-event
   articles forever.
3. **New migration**: `ALTER TABLE content_feed_items ADD COLUMN
   event_extracted_at timestamptz` (nullable, no default — absence means "not yet
   considered", matching the column's own name literally). One column, no index
   needed yet (queries filter on `status` first via the existing `idx_cfi_site_status`,
   then the `IS NULL` scan is bounded by that).
4. **New LLM step `extract_event_facts`** (workflow config, DB-live, no image
   needed for this half) — classifies each loaded item: does it confirm a
   specific, dated, real-world event (not "fights happen sometimes" — a named
   date/fixture)? If yes, extract `event_date`, `venue`, `participants`,
   `broadcaster` (whichever are stated; never invented — same "don't fabricate"
   discipline as the citation-verification layer) plus the citation fields
   (`claim`, `quote`, `url`, `publisher`, `title`, `published`) in exactly the
   shape `VerifyAndRegisterCitationsAction` already consumes as `candidates`.
5. **Extend `VerifyAndRegisterCitationsAction`'s field pass-through list**
   (`evidence_citations.go:299`, currently
   `[]string{"value", "unit", "tolerance", "writer_line", "staleness_days", "context_terms"}`)
   to add `"event_date", "venue", "participants", "broadcaster"`. This is the
   entire extension needed on the writer side — `verifyCitationLive` and the
   supersede-write pattern (`writeCitationRegister`) are reused UNCHANGED, same
   as `directory_claims.go`'s own precedent. Candidates carry `"kind": "event"`
   (the action already defaults `kind` to `"metric"` when absent — no code
   change needed there, just the prompt setting it).
6. **Wire into `feed-triage`'s own workflow** rather than a new agent +
   dispatcher. `apply_scores`'s `next_step` (currently `"complete"`) becomes a
   new `evaluate_condition` step gated on `relevant` count from `apply_scores`'
   own output (`0` → `complete`, default → the new chain:
   `load_for_event_extraction` → `extract_event_facts` (LLM) →
   `register_event_facts` (`verify_and_register_citations`, candidates path
   pointed at the LLM step's output) → `mark_event_extracted` → `complete`).
   Rides on feed-triage's existing trigger/cadence (already tuned by 410) — no
   new cron, no new dispatch path to get wrong (per
   [[detection-works-schedule-and-dispatch-do-not]] in memory: dispatch is the
   part that tends to silently not work).

**Sequencing**: image (new actions + extended field list) built and rolled BEFORE
the workflow-config UPDATE that references them — a seed naming an unregistered
action fails at runtime (CLAUDE.md). Migration can go in alongside the image build
(it's additive, no data dependency on the code).

## Fix candidate #2 (revisit/correction path) — designed, not built this pass

Bug 427 §6: dates get corrected, not just added — `refresh_evidence_base`'s
staleness/drift machinery is the closest analogue, reuse rather than re-derive.
`bugfix_410_feed_phase_lock`'s cadence/staleness handling is the directly
transferable pattern (see its RUNBOOK). Deferred until candidate #1 is live and
has real fact rows to revisit against — building the correction path first has
nothing to correct.

## Fix candidate #3 (entity-directory page role)

Blocked on diagnosis `d6d350ec-e16b-4792-9282-ca5155369791` (UNVERIFIABLE,
iteration-capped) — site-planner territory. Not this lane's build; watching, not
duplicating.

## Open questions

- Should the extraction LLM step run per-site (as designed, riding feed-triage's
  existing per-site trigger) or could it be made shared-once-per-article like
  `news_feed_pooling`'s design split? Deferred: pooling is owner-parked, and 427's
  urgency (paid delivery) argues for the smallest working thing now, not the
  fleet-optimal shape. Revisit if/when pooling unparks.
- Whether `VerifyAndRegisterCitationsAction` fired at all during boxingonline's
  build (bug 427 §3, `[UNMEASURED, left to the fixing thread]`) — worth checking
  once instrumented, though it doesn't change the fix: the action exists and is
  opportunistic-only either way.
