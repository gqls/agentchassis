# PLAN — news_feed_ingestion

**Candidate #1 STATUS: LIVE and PROVEN, 2026-09-02.** Image v1.0.1352 built
from committed HEAD, pushed, rolled, verified at the binary (sha controls, not
the roll status). Workflow wired into `feed-triage`'s live config, migration
684 applied by hand and recorded (see NOTES — it was initially MISSED and
caught by the first end-to-end test failing, not by review). Re-dispatched
against boxingonline.com (the motivating site) end to end: 6 new dated event
facts registered in its `evidence_base`, verified live against their source
articles, real fight results with venues/participants, nothing fabricated.
49 `content_feed_items` rows (backlog + fresh) now carry `event_extracted_at`.
Council-approved (`4849c95f...`) before the roll. What remains on bug 427:
candidate #2 (peer, in progress) and candidate #3 (blocked on a separate
diagnosis) — this lane's own build is complete.

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
   as `directory_claims.go`'s own precedent. Candidates carry
   `"kind": "entity"` (`datahelpers.FactKindEntity`, "a named thing that
   exists" — a specific dated fight qualifies).

   > **Correction, 2026-09-02 (caught by the `bugs_open/427` peer session
   > before this was committed, reading the dirty tree).** The design
   > originally said `"kind": "event"`. `EvidenceFact.Kind`'s vocabulary is
   > CLOSED (`platform/orchestration/datahelpers/claims.go`, `bugs_open/105`):
   > `metric | capability | entity | attestation`, with `count`/`metrics`/
   > `counts` as the only live aliases. `"event"` is in neither set —
   > `CanonicalKind()` would have silently demoted every registered fact to
   > `"metric"`, and `UnrecognisedKinds()` (called from
   > `validate_page_content.go`, every build) would have logged an
   > "unrecognised kind" warning on every site, forever, until someone fixed
   > it. `"entity"` is already canonical and fits the meaning exactly, so this
   > is a one-word fix, not a rework — caught before the LLM prompt (item 4,
   > not yet written) could bake the wrong literal in.
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

## Fix candidate #2 (revisit/correction path) — being built by the `calendar`/427 peer lane

Bug 427 §6: dates get corrected, not just added — `refresh_evidence_base`'s
staleness/drift machinery is the closest analogue, reuse rather than re-derive.
The peer session took this after checking for overlap; not this lane's build.

**Part 1 done, verified independently (commit `f865153f8`, 2026-09-02):**
`refreshCitationFact` needed no changes — event facts carry a plain
`source.citation` regardless of kind, so the existing live-URL-recheck path
already covers them. The real gap, found by the peer tracing
`composeWriterBlock` directly (not something this lane's own design had
spotted): that function only substitutes `{value}` for facts carrying a
numeric value; anything else falls into a CAPABILITIES bucket that carries
`writer_line` **completely verbatim, no substitution**. Since this lane's event
facts have no numeric `value`, a `writer_line` phrasing `{event_date}`/`{venue}`/
`{participants}`/`{broadcaster}` (the natural thing to write) would have shipped
those tokens **unsubstituted** into the writer's prompt on any site with
`writer_block_managed: true`. Fixed with a third "SCHEDULED EVENTS" bucket,
missing fields render `"TBC"` rather than a bare brace. Matches this lane's
exact field names — verified by reading the diff and running
`TestComposeWriterBlock`/`TestComposeWriterBlockEventFacts` (both green) before
accepting the peer's account rather than taking it on trust.

## Fix candidate #3 (entity-directory page role)

Blocked on diagnosis `d6d350ec-e16b-4792-9282-ca5155369791` (UNVERIFIABLE,
iteration-capped) — site-planner territory. Not this lane's build; watching, not
duplicating.

## Council verdict on candidate #1's submission (2026-09-02)

**APPROVED** (`4849c95f-2594-48e6-87b9-acee6341b0f8`, verified directly against
`diagnosis_artifacts`, not taken on a peer's report), 5 advisory objections,
none high-severity. Two worth recording:

- **`architecture` seat**: the four new event fields are written only via raw
  map keys, and the typed `EvidenceFact` struct (`claims.go`) used elsewhere
  for parsing may silently drop them on any round-trip. **Checked directly,
  not just accepted**: `LANDMINES.md` already documents this exact hazard
  ("Parsing evidence_base through its own typed struct and writing it back
  DELETES every citation, writer_line, unit and staleness_days on the site") —
  `EvidenceFact` was *already* missing `unit`/`writer_line`/`staleness_days`
  before this lane's change, and the landmine states plainly that both live
  writers (`refresh_evidence_base_action.go`, `evidence_citations.go`)
  "avoid it only because they never use the struct: both work on
  `map[string]interface{}`... that is deliberate." This lane's extension
  follows the identical, already-established pattern — not a new risk, a
  fourth confirmation of an existing one. No action needed; noted for whoever
  next touches `EvidenceFact` that the gap is now 4 fields wider, not 4 fields
  newer.
- **`guardian` seat**: two shared-consumer checks, both verified clean —
  `grep -rn "range fact\b\|for.*fact\[" platform/orchestration/` finds nothing
  that enumerates `evidence_base` fact keys positionally/exhaustively (the
  widened allowlist can't break an exhaustive reader because none exists), and
  a query against `agent_definitions.default_config` for either new action
  name returns zero rows (no collision with any live workflow).
- The `architecture` seat's broader point — is `evidence_base`'s flat
  `facts[]` array the right shape for a *growing* corpus of dated, multi-field
  events (participants arrays, venues, no index, no dedup) — is not a defect
  in this submission and not acted on here. Recorded as a real open question
  for if/when a second or third event/calendar site is built on the same
  mechanism: at that point `entity_ids`/`duplicate_of`'s original,
  never-answered purpose (§ above) may be worth revisiting properly rather
  than continuing to bolt scalar fields onto the array.

Also recorded for next time: `prior_art_librarian` correctly noted this
submission's load-bearing claims (closed `Kind` vocabulary, `verifyCitationLive`/
`writeCitationRegister`'s existence, the three-writer enumeration) were
asserted from having read the source, but not attached as evidence *in the
submission itself*. All were independently re-verified true here regardless —
but the next submission should quote the evidence inline rather than making
the council re-derive it.

## UK-news default for `.uk`/`.co.uk` sites — STATUS: built, council-approved, image shipped by owner; verification PENDING (kubectl down)

Code, tests and migration 691 (+ROLLBACK+VERIFY) built and committed
(`0a408f8db`), council-APPROVED round 1 (`8842fe96-9a71-4ea5-9993-2483f10712cb`,
`Council-Reviewed:` on `2f8411b7e`). Built and pushed both images
(`agent-chassis`, `web-search-adapter`) at `v1.0.1353`; asked the owner
before rolling (see NOTES — a standing memory records the owner reserves
single-service deploys for himself). **The owner then deployed a fresh
build directly** (`v1.0.1355` on both, confirmed genuinely pushed via
`docker manifest inspect`; overlay files show this uncommitted in the
working tree as of this note). **Not yet verified**: pod rollout status,
the running binary's commit (kubectl access is down fleet-wide — the known
3-day token expiry, confirmed via the JWT's own `exp` claim, not a bug —
so no binary sha grep was possible), whether migration 691 has actually
been applied, and the live `.uk`-site dispatch this design's own
verification plan calls for. **All four are the next session's first job**;
see `HANDOFF_2026-09-02b_continue_here.md`.

## UK-news default for `.uk`/`.co.uk` sites — design (2026-09-02, resumed after session switch)

Owner ask (`CONTRIB_2026-08-31_from_loanzy_lane_owner_wants_uk_news_default_for_uk_tlds...md`):
*"The news is from America. I'd like it to be UK news for all .co.uk and .uk
sites, perhaps as a flag with a UK default."*

**Root cause, found by reading the actual provider code and fetching the
providers' own API docs (not assumed):** `providers.SearchOptions` (`provider.go:37-44`)
already declares a `Region string // us, uk, etc.` field — **but nothing has
ever populated it**, and nothing reads it except a dead comment in
`duckduckgo.go`. The adapter's own `SearchOptions{...}` construction
(`adapter.go:217-220`) sets only `SearchType`/`TimeRange`. Checked live:
`PRIMARY_SEARCH_PROVIDER=firecrawl` (`kubectl get deploy web-search-adapter`
env, 2026-09-02) — Firecrawl is the operative provider, not a fallback.
**Fetched Firecrawl's own `/v2/search` API docs directly**: it supports a
`country` parameter (ISO code) that **defaults to `"US"`** and applies to both
`web` and `news` sources — this default is almost certainly the literal,
load-bearing cause of "the news is from America", not a downstream ranking
effect. ScrapingBee's Google-proxy API (the first fallback) separately
supports `country_code` (also verified against its own docs) — both APIs use
`"UK"` as their United Kingdom value, matching `SearchOptions.Region`'s own
doc comment exactly, so no value-translation table is needed.

**DuckDuckGo (the second fallback) is out of scope, checked not assumed:**
`TestDuckDuckGoDeclinesNews` (`search_options_test.go:173`) proves the DDG
provider declines `search_type: "news"` outright
(`ErrUnsupportedSearchType`) — and `FetchNewsSearchAction` always sets
`search_type: "news"` for a `news_search` source
(`feed_fetch_async_actions.go:158`). So DDG structurally never serves this
pipeline's requests regardless of region — its hardcoded, unconditional
`kl=uk-en` (`duckduckgo.go:139`, a separate pre-existing latent oddity: every
*web*-type DDG fallback search today is silently UK-locked regardless of the
requesting site) is real but irrelevant to this fix and is being left alone —
touching it would be scope creep with no effect on the bug being fixed.

**Where the value flows (single choke point, verified by reading, not
inferred):** `content_sources.config` (jsonb) → `FetchNewsSearchAction`
(`feed_fetch_async_actions.go`, reads `sourceConfig[...]`, sets
`params.StepConfig.Config[...]`, mirrors the existing `time_range` block at
lines 163-168) → `WebSearchAction` (`web_search_action.go`, mirrors the
existing `timeRange`/`provider` reads at lines 62-75, adds `region` to
`adapterRequest["body"]["data"]`) → adapter's `RequestPayload.Data` →
`SearchOptions.Region` (`adapter.go:217-220`) → each provider's `Search`.
This is the ONLY path `news_search` sources take — `api_news` (xAI/Grok
sources, a different content_sources.source_type) is a separate mechanism,
untouched, out of charter (matches the routing session's own measurement of
"48 `news_search` configs").

**Value convention:** lowercase `region` key, value `"uk"` — matches
`SearchOptions.Region`'s existing doc comment ("us, uk, etc."). Upper-cased
only at each provider's call site where the external API's documented
examples are upper-case (`strings.ToUpper(opts.Region)`).

**No existing TLD-derivation mechanism anywhere in the codebase** (grepped
`platform/orchestration/actions/*.go` for TLD/suffix logic before building a
second one, per this lane's own open design question #4) — the closest
precedent is `isBlockedDomain`'s `strings.HasSuffix` pattern (`helpers.go:151`),
reused for the new `.uk` check (`.uk` as a suffix already covers `.co.uk`
without a separate branch).

**Blast radius, measured before building** (`[MEASURED 2026-09-02]`):
```sql
SELECT count(DISTINCT s.id), count(*) FROM sites s
JOIN content_sources cs ON cs.site_id = s.id
WHERE cs.source_type = 'news_search'
  AND (lower(s.domain) LIKE '%.co.uk' OR lower(s.domain) LIKE '%.uk');
```
**6 sites / 26 of 52 fleet-wide `news_search` source rows** — `farmerinsurance.uk`,
`idea.uk`, `loanandmortgagecalculator.co.uk`, `mortgagecalculator.co.uk`,
`remortgagecalculator.uk`, `webdesign.co.uk`. None carry any region-shaped key
today (confirms the routing session's fleet-wide zero). Because
`seedNewsSearchSources` inserts `ON CONFLICT (site_id, name) DO NOTHING`, a
seed-time-only fix would **never** touch these 26 already-provisioned rows —
satisfying the owner's actual ask ("for all .co.uk and .uk sites", not just
future ones) needs a backfill migration alongside the code change, scoped
exactly to this measured set.

### Build plan

1. **`seed_content_sources_action.go`**, `seedNewsSearchSources`: thread
   `domain` (already loaded at the call site, `line ~187`) into the function,
   add `if strings.HasSuffix(strings.ToLower(domain), ".uk") { config["region"] = "uk" }`
   before marshalling — same "a human inspecting a row can see it" rationale
   as PLAN item 3 above. Covers every future `.uk`/`.co.uk` site from the next
   `content-feed-orchestrator` seed cycle onward.
2. **`feed_fetch_async_actions.go`**, `FetchNewsSearchAction`: read
   `sourceConfig["region"].(string)`, pass through to
   `params.StepConfig.Config["region"]` — mirrors the `time_range` block
   exactly.
3. **`web_search_action.go`**, `WebSearchAction`: read `config["region"].(string)`
   (mirrors `timeRange`/`provider`), add to `adapterRequest["body"]["data"]["region"]`,
   the log line, and `Metadata` — parity with every other passthrough param.
4. **`internal/adapters/websearch/adapter.go`**: add `Region string
   \`json:"region,omitempty"\`` to `RequestPayload.Data`; set
   `opts.Region = req.Data.Region` in the `SearchOptions{...}` construction;
   add to the "Executing search" log line.
5. **`firecrawl.go`**: `if opts.Region != "" { payload["country"] =
   strings.ToUpper(opts.Region) }`.
6. **`scrapingbee.go`**: `if opts.Region != "" { params.Add("country_code",
   strings.ToUpper(opts.Region)) }`.
7. **Migration `691_...sql`** (additive, no schema change — jsonb key only):
   backfill UPDATE on the exact 26 measured rows (`source_type='news_search'`
   AND site domain suffix `.uk` AND `NOT (config ? 'region')`, idempotent),
   DO/RAISE verify block asserting the touched-row count against a
   pre-transaction `SELECT count(*)` (CLAUDE.md: a bare `SELECT` verify cannot
   stop a bad `COMMIT`). `_ROLLBACK.sql` strips the key back out on the same
   predicate; `_VERIFY.sql` for a human eyeball of the after-state.

**No `agent_definitions.default_config` workflow change needed** — unlike
candidate #1, `region` flows entirely through already-wired config lookups in
already-wired steps; nothing new to seed into a live workflow.

**Two images to build and roll, not one** — `platform/orchestration/actions`
(steps 1-3) lives in the `agent-chassis` binary; `internal/adapters/websearch`
(steps 4-6) is its own deployable, `web-search-adapter`
(`cmd/web-search-adapter`, confirmed via its own kustomize dir and a running
pod `web-search-adapter-*`). Both need `IMAGE_TAG` bumps and rolls.

**Verification plan**: dispatch a real `fetch_news_search` against a live,
non-parked `.uk` site (`webdesign.co.uk`, `idea.uk`, or
`farmerinsurance.uk` — NOT `mortgagecalculator.co.uk`, flagged elsewhere in
the fleet's own comments as parked/404-on-the-wire) after the backfill,
confirm the request actually reaching Firecrawl carries `country: "UK"`
(adapter logs), and spot-check returned result URLs/publishers skew UK
(BBC/Sky/Reuters UK bureau etc.) rather than trusting a 200 status alone.

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
