# HANDOFF 2026-08-08 — webdesign.uk: REBUILD IN FLIGHT, clean classification PROVEN; the pipeline must be HAND-DRIVEN item by item

**Start here cold. Supersedes `HANDOFF_2026-08-06_continue_here.md`** (kept; its
§1 state table is still right for box/tunnel/preview/token — not repeated here).
Read with: `NOTES` 08-06 evening → 08-08 (the asset-path discharge, the firecrawl
cache, the queue starvation) · `SUBMISSION_2026-08-08_*` (the envelope + roadmap
actually submitted) · `WRONG_CALLS.md` 08-08 (the cache wrong-call).

## 0. One paragraph

The rejected v1's four causes are all closed: the asset path needed no fix
(row-driven routing; row now correct), `content_direction` + `classification` +
`identity` + `design_intent` were **regenerated clean on 2026-08-08 22:08Z** from
a resubmission carrying the owner's mission verbatim plus a five-page
authoritative roadmap, and the classifier this time read a PARKED domain (522) —
verified in its own `detected_signals`: *"No existing site (522 timeout)"*,
*"Five named pages in roadmap"*. New classification: **brochure, 0.97, 5 pages,
has_existing_site=false**. The catch discovered on the way: **the fleet's build
dispatch queue does not self-drive** (no CronJob; FIFO with 306 older items from
4 other sites ahead of ours), so every pipeline stage for this site must be
dispatched BY HAND until either the backlog drains or the queue gets fixed. The
pattern is proven and below. Stage 2 (`vertical-exemplar-researcher`) was
dispatched 22:10:54Z (corr `f37d5cd6…`) and was in flight when this handoff was
written — **check its outcome first.**

## 1. THE DRIVE LOOP — how to run the rest of the build by hand

The chain: each stage's orchestration completes → files the NEXT `site_work_items`
row (handler_agent names the agent) → which nothing will dispatch. So, per stage:

1. **Find the item**: `SELECT id, item_type, handler_agent, status FROM
   site_work_items WHERE site_id='1fcfa4f3-ec80-4010-878b-b971cd46711f' AND
   status='triaged';`
2. **Claim it** (prevents a stray heartbeat double-driving it):
   `UPDATE site_work_items SET status='claimed', updated_at=now() WHERE id='<id>'
   AND status='triaged';`
3. **Orchestrate the handler directly** — envelope
   `{"action":"orchestrate","config":{"agent_type":"<handler_agent>"},
   "input_data":{"domain":"webdesign.uk","site_id":"1fcfa4f3-…","work_item_id":
   "<id>","item_type":"<item_type>","source":"<source>","spec":<spec>,
   "current_page":<spec>}}` — kcat block exactly as
   `082_submit_domain_unified.sh`'s (topic `system.agent.generic.requests`,
   `kcat -P -c 1`, fresh UUIDs, `client_id=demo_client`). Worked examples in
   NOTES 08-08 (classifier corr `b3e472f1…`, vertical corr `f37d5cd6…`).
4. **Verify by payload** (kcat exit 0 is nothing):
   `SELECT status, current_step FROM orchestration_states WHERE
   correlation_id='<corr>';`
5. **On COMPLETED**: check the expected spec/page rows landed, then
   `UPDATE site_work_items SET status='complete', updated_at=now(),
   resolution_path='<how>' WHERE id='<id>';` and go to 1 for the item the stage
   just filed.
6. On FAILED: read `agent_error_log` (`occurred_at`, `context->'issues'`), fix,
   reset to `'triaged'`, re-claim, re-dispatch. Spawn→call handshake failures
   happen (~half, memory `spawn-call-handshake-races`) — never cancel the row
   pre-diagnosis.

Expected sequence after vertical research (from the 08-04 build's spec
timestamps): **strategy → briefing → resolved_composition/site plan → `needs_page`
fan-out (5 pages) → per-page content/render → deploy to vm-sites → sitesync pulls
to the box ≤5 min.** The planner enforces the roadmap ("build ONLY the pages
listed"). Respect the 300s post-roll window before each dispatch — the fleet was
rolling repeatedly on 08-08 evening (1264→1269 in one evening; check pod age).

## 2. What is DONE and PROVEN (do not redo)

- **Asset path**: no gap. All four artefact kinds route by `sites.github_repo`
  (`resolveGitRepoNameDB`); row says `vm-sites`; idea.uk's 08-05 deploys prove
  the path live. v1's assets went to `gqls/sites/webdesign.uk/` only because the
  row flipped AFTER they deployed. **Tidy later** (after the rebuild's own assets
  land in vm-sites): delete stale `gqls/sites/webdesign.uk/` — nothing serves it.
- **Rejected v1 page row**: archived AND renamed `index-rejected-v1-20260806`
  (rename load-bearing — `(site_id,name)` unique, sync upsert never updates
  `status` on conflict, an archived `index` would block the new home page
  silently). Its two `needs_human_review` items cancelled.
- **Specs now current**: mission_brief (verbatim resubmit), roadmap_brief (NEW,
  five pages incl. faq VAT/price answers via evidence_base; chat box explicitly
  a later phase), classification/identity/content_direction/design_intent (all
  22:08Z, clean), evidence_base UNTOUCHED and intact (15 bans, 7 facts).
- **302s RESTORED and verified** (apex+www 302 → webdesign.co.uk, preview 200).
  Parked window was 20:19–22:07 UTC, deliberate, owner-approved route.
- **Classifier scrape config now sends `max_age: 0`** (fleet, deliberate, kept;
  revert SQL in NOTES 08-08 §2). Without it Firecrawl serves a DAYS-old cached
  snapshot — see LANDMINES ("Firecrawl serves a CACHED snapshot").

## 3. Platform findings a next session must not re-derive

- **No build heartbeat CronJob exists.** 076 is the only driver, one SITE per
  firing, fleet-FIFO by item age (the 08-02 fix in `bugs_closed/169` part B +
  `176`). 306 dispatchable items ahead of this site's on 08-08 across
  leopardess (61, oldest 07-26), webdesign.co.uk (188), loancash (18),
  loanandmortgagecalculator (39).
- **`build-dispatch-loop` invoked with bare `action=orchestrate` silently
  no-ops**: loads items, `has_items=true`, loop expansion "handled", processes
  NOTHING, reports COMPLETED:complete, items untouched (orch `4e26e881…`,
  20:23:50Z). The working path is the trigger's spawn+call (`action: process`) —
  leopardess's item processed fine through it the same evening. NOT yet filed as
  a bug — a next session touching dispatch should file it via `090` (it is
  cross-cutting; CLAUDE.md's ruling applies) or fold it into the queue-starvation
  case. Do NOT re-diagnose from scratch: the log trail and orch ids are in NOTES
  08-08 §3.
- The petclinic.jollyes.co.uk vet-intel loop was FAILING every ~30 min at
  `extract_and_reconcile` on 08-06→08-07 — another lane's; ignore it when
  reading orchestration_states.

## 4. Owner ledger (delta from 08-06)

Nothing new owed. The 08-06 ledger stands: scoped Anthropic key (chat service) ·
correction-fee number · terms before live Stripe · final review + cutover
approval. The chat box/input-box phase stays AFTER owner review of the five-page
build. Cutover unchanged (DNS-only; delete the two Page Rules at that moment).

## 5. Access map

Unchanged from 08-06 §6: box key `~/.ssh/webdesign_box_ed25519` · CF token
`~/.config/cloudflare/token` (works; expires 09-01; PageRules PATCH proven —
zone `746f81e6…`, rule ids `6d4d5b67…` apex / `88794916…` www) · `gh`=gqls ·
kubectl → cluster/DB. Session permission note: Cloudflare API writes trip the
permission classifier — the owner approved the calls on 08-08; a new session
should expect the same prompt.
