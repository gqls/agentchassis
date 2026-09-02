# HANDOFF — news_feed_ingestion — continue here

Written 2026-09-02, mid-session, at the owner's request to switch sessions.
**Read this file first**, then `PLAN_2026-09-02_news_feed_ingestion.md` (design +
decisions), `NOTES_news_feed_ingestion.md` (full chronological account incl.
two mistakes made and fixed this session), `RUNBOOK_news_feed_ingestion.md`
(commands), `README_where_we_are.md` (plain-prose account for the owner).

## 0. What this lane is

Charter: the raw news-ingestion pipeline — source enrollment, fetch
scheduling/cadence/queue fairness, and normalising/extracting structure out of
what arrives. Concretely: `content-feed-orchestrator`,
`find_news_sites`/`content-feed-trigger`, `feed-triage`, `content-feed-refresh`,
`content_feed_items` as a data asset. Full charter + explicit exclusions in
PLAN's opening section.

## 1. DONE — bugs_open/427 fix candidate #1 (LIVE, PROVEN, do not redo)

Extraction step that turns a confirmed news item into a dated, citation-verified
`evidence_base` fact. **Fully shipped and tested against real data
(boxingonline.com) — nothing left to do here.**

- Code: `platform/orchestration/actions/feed_event_extraction_actions.go` (new),
  `evidence_citations.go` (extended field pass-through, kind="entity" — NOT
  "event", see PLAN's correction), `registry.go`. Council-approved
  (`4849c95f-2594-48e6-87b9-acee6341b0f8`).
- Migration `684_content_feed_items_event_extraction_column.sql` — applied by
  hand and recorded (`run-migrations.sh --record-only`), NOT via bulk `--apply`.
- Image `v1.0.1352` (agent-chassis), built from committed HEAD, pushed,
  rolled, verified at the binary (sha controls on `/proc/1/exe`, both live pods).
- Workflow wired live into `feed-triage`'s `agent_definitions.default_config`:
  `apply_scores → load_for_event_extraction → check_for_events →
  extract_event_facts → register_event_facts → mark_event_extracted → complete`.
- **Tested end-to-end against boxingonline.com**: 6 new dated event facts
  registered in its `evidence_base`, verified live against source articles.
  49 `content_feed_items` rows now carry `event_extracted_at`.
- Two real mistakes made and fixed THIS session, full account in NOTES: (1) the
  migration was written but never applied — caught by the first live test
  failing with `column ... does not exist`; (2) `check_has_items` was found to
  short-circuit past the whole extraction chain whenever a cycle had zero fresh
  ingested items, regardless of an unextracted backlog — restructured so both
  entry paths converge on the loader, gated on its own live count.

**Peer status (as of last contact, may have moved on)**: the `bugs_open/427`
peer session (a.k.a. `calendar`) is building fix candidate #2 (correction path
— turned out to need almost nothing new, `composeWriterBlock`'s event-token
substitution was the one real gap, fixed and verified in `f865153f8`) and the
render target (`query.upcoming_events` resolver, committed `da2ab0d44`,
submitted `08f56b7e-61e4-42d1-a3b6-13d700dd833c`). They said they'd wire the
component that declares the new source against boxingonline's real facts
"within the hour" of their last message (mid-way through `bugs_open/428` at
that point) and would ping when live. **Check for that ping / re-check
`bugs_open/427`'s own file §9 for a newer status update before assuming this
is still pending.** Candidate #3 (entity-directory page role) is separately
diagnosed and filed as `bugs_open/428` — not this lane's concern.

## 2. DONE — bugs_open/316: core already fixed, one rider refuted

- Ordering + capacity (the bug's original substance): already fixed and live
  by a predecessor session (migrations 554, 556). Nothing to do.
- Rider "unenrolled sites never reach the seeder"
  (`CONTRIB_2026-08-25_from_loanzy_lane_...md`): **investigated and refuted**
  this session. The claimed structural exclusion doesn't exist — the
  `NOT EXISTS` eligibility arm has been in `find_news_sites` since its
  original seed, verified against migration 554's own header and today's live
  population (14/15 recommended sites have sources; the one exception just
  hadn't had its first cycle yet). Full evidence in NOTES and in a RESPONSE
  section appended directly to the CONTRIB file. Do not re-investigate this.

## 3. IN PROGRESS — UK-news default for `.uk`/`.co.uk` sites — RESUME HERE

**This is the live thread, interrupted mid-investigation.** Owner ask (via
`CONTRIB_2026-08-31_from_loanzy_lane_owner_wants_uk_news_default_for_uk_tlds...md`):
*"The news is from America. I'd like it to be UK news for all .co.uk and .uk
sites, perhaps as a flag with a UK default."* User confirmed via
`AskUserQuestion` ("Build it now") to proceed with this build in this session,
before the session-switch request arrived.

**What's already measured (by the routing session, news_editorial_features
lane — read before re-measuring):**
1. Fleet-wide, the capability does not exist: across 48 `news_search` configs,
   zero region/country/locale/`gl`/`hl` keys anywhere. Must be built end to
   end, not just toggled on.
2. The seam is `web_search_action.go` — it already reads per-source config
   values of exactly this shape (`num_results`, `time_range`). A region key
   belongs beside them.
3. The default belongs at SEED time (`seed_content_sources_action.go`) so a
   human inspecting a row can see it, not derived implicitly at query time.
4. Open design question (theirs, unresolved): check whether the estate
   already derives anything from TLD elsewhere before inventing a second
   derivation.
5. Scope caution: a TLD-keyed default changes live content for every existing
   `.uk`/`.co.uk` site with `news_search` sources — count the affected
   population BEFORE it ships, not after.

**What I'd done in this session before the interrupt** (not yet a design,
just the first read): opened `platform/orchestration/actions/web_search_action.go`
and confirmed the pattern — `num_results` and `time_range` are read from the
step's `config` map with simple type-asserted lookups (`config["num_results"].(float64)`,
`config["time_range"].(string)`), around lines 58-68, then threaded through to
the actual search call and echoed back in the result map (~166-226). **Next
step**: read the rest of that file to see exactly what search provider/API is
called and whether it has a native region/country parameter (xAI web_search,
or whichever provider — check `config["provider"]`), then read
`seed_content_sources_action.go` to see how a source's config gets built at
seed time and where a TLD-derived default would slot in. Then design (don't
build yet): a new config key (e.g. `region` or `country`, check what the
underlying search API actually calls it before inventing a name), a seed-time
derivation from the site's domain suffix, and the population/blast-radius
query design item 5 above asks for — before writing any code, run that query
for real: `content_sources` joined to `sites` on domain suffix `.uk`/`.co.uk`
with a `news_search`-shaped source, count them.

**Not yet done, in order**: finish reading `web_search_action.go` +
`seed_content_sources_action.go` → design (add to PLAN) → measure blast
radius → build (Go change + migration if a new column, or just a new
optional config key if additive within existing jsonb config — check which)
→ tests → council submission → build/roll/deploy following the exact same
process as candidate #1 above (image before workflow/seed wiring) → verify
against a real `.uk` site.

## 4. Standing practices this lane has been following (keep doing these)

- Verify a peer's or a CONTRIB's claim against the live system before acting
  on it — this session refuted one CONTRIB and caught a defect in its own
  design comment (`kind="event"` vs `"entity"`) by checking source directly,
  not by trust.
- Test end-to-end against real data before declaring anything done — this
  session's two real mistakes (unapplied migration, workflow short-circuit)
  were BOTH caught this way, not by code review.
- Pathspec every commit, explicitly — this session made exactly one mistake
  here (`git commit --allow-empty` with no pathspec swept in stale staged
  files from other sessions; checked for harm, found none, documented
  visibly in NOTES rather than hidden). `git status --short` before any
  commit with no explicit file list.
- Never reuse a working tree's uncommitted `IMAGE_TAG`/kustomize tag bump —
  check `docker manifest inspect` for whether it's already pushed/live before
  building; bump to a fresh, confirmed-unused tag instead.
- Migrations: apply your OWN migration by hand + `run-migrations.sh
  --record-only`, never the bulk `--apply` on a shared queue (it takes every
  pending file, most of which are not yours).
- Verify a deploy at the binary (`/proc/1/exe` sha grep, positive AND
  negative control — never "40 zeros", it matches Go's binary padding), never
  at the roll status alone.
