# DESIGN — Gripper Selection & Integration Dossier pilot (robot-hands.com)

*Consolidated from two design passes (island+site half; cluster half), each
grounded against the live repo, live clients_db, and the live site. Seam
discrepancies reconciled in §5 — where this doc and a half-design disagree,
THIS doc wins. Parent workstream: `../per_site_ai/` (PLAN D15/D17);
features_open/013. Status: designed, awaiting owner answers (§8) — nothing
built.*

## 1. Shape (one paragraph)

A visitor on `robot-hands.com/gripper-report/` chats with a tightly-prompted
Haiku intake assistant (widget JS → island VM at tools.apis.uk), which fills a
structured application spec; on completion the visitor enters their email and
submits. The cluster pulls pending requests (one-way, no PII), and for each
runs a `report-builder` workflow: deterministic Go scoring (MatchMatrix v2
physics ported server-side over the 10 verified grippers) → fact-block-bound
LLM prose → deterministic prose verification → compose page → validate →
render → git deploy at `/reports/<uuid>.html` + a status sidecar
`/reports/<uuid>.json`. The island polls the sidecar and emails the visitor
the link (or an apology on failure/timeout). Email never enters the cluster;
the public never touches the cluster.

## 2. Island + site half

**Service** `cmd/gripper-intake/` (flat package main, idea.uk shape):
main / service (mux, handlers, CORS) / chat (raw Anthropic client, system
prompt, structured-output schema, spec merge, caps) / store (guarded status
transitions `UPDATE … WHERE status=$expected`) / ratelimit (banded per-IP,
lift from `docs/.../idea.uk/golang_files/audience_check.go:31-110`) / mailer
(lift from `service.go:771-910`; file-write fallback when SMTP_HOST empty) /
poller (60s tick: sidecar polling, fulfilled→email, expiry→apology, GC) /
schema.sql (go:embed, idempotent) / service_test.go (fake Anthropic, fake
SMTP, hardening tests). Dockerfile:
`build/docker/backend/gripper-intake.dockerfile` (core-manager shape).

**Island tenancy** (2nd tenant on the ALREADY-RUNNING gauntlet island —
`/opt/island/`, postgres:16 + caddy:2 compose; repo copies under
`docs/.../gauntlet_dead_cta/infra/island/` are source of truth, scp on
change): own DB `gripper_intake` + role; own spend-capped Anthropic key
(never the debate engine's); compose block `gripper-intake:8090`; Caddyfile
`@gripper path /api/gripper/*` → reverse_proxy, 64KB body cap; backup_pg.sh
second dump line; RUNBOOK "Tenant 2" section. Secrets in `/opt/island/.env`
(on-box only): GRIPPER_PG_PASSWORD, GRIPPER_ANTHROPIC_API_KEY,
GRIPPER_PULL_KEY (32 hex bytes), SMTP_*, OPERATOR_EMAIL.

**API** (all `/api/gripper/v1/`): `POST /session` (per-IP 6/h, 20/d) →
`{session_id, greeting}` · `POST /chat` `{session_id, message≤2000}` →
`{reply, spec, missing_fields, complete}` (per-IP 60/h 200/d; ≤30
turns/session; ≤60K tokens/session; global daily turn cap env, default
2,000; 409 on cap, 503 on Anthropic fail) · `POST /submit` `{session_id?,
email, company_website(honeypot), _elapsed, spec? (plain-form mode)}` → 201
(byte-identical on silent honeypot/timing drop; per-IP 3/h 10/d) ·
`GET /health` · `GET /requests?since=` (X-Internal-Key auth; cluster pull —
see §5.1 for the reconciled contract; **email absent from payload by
design**).

**Chat call**: claude-haiku-4-5, max_tokens 1024, `output_config` json_schema
forcing exactly `{reply, spec(8 nullable fields), complete}`
(additionalProperties:false); fallback flag for prompt-level JSON if
output_config rejected; 25s timeout, one retry on 429/5xx honouring
retry-after. Cost ≈ $0.15/session worst case; ~$20/day global worst case;
spend-capped key backstop.

**Prompt contract** (design, copy written at build): intake assistant for
this dossier only; British English; one question/turn; reply ≤120 words;
fields with per-field guidance — payload_kg, part_geometry (shape + key mm),
part_surface (material/finish from the μ-table list), environment (IP,
washdown, temperature), cycle_rate (ppm), mounting (robot + flange std),
budget (optional), notes (optional); record only stated values, normalise
metric, never infer, ambiguous → clarify + stay null; full spec every turn
(server merges — non-null never regresses); complete only when required
non-null; NEVER ask for email in chat (widget owns it — PII stays out of
transcripts); off-topic/injection → one-line redirect, spec unchanged.
Containment is structural, not prompt-level: schema-forced output, nothing
executed, widget renders via textContent, spec = untrusted data downstream
(cluster treats it so too).

**Island schema** (DB `gripper_intake`): `chat_sessions` (id uuid pk,
created/last_activity, client_ip, user_agent, turns, input_tokens,
output_tokens, transcript jsonb, spec jsonb, status
active|submitted|expired|blocked) · `report_requests` (id uuid pk — IS the
report slug; session_id fk, email, spec, spec_summary, report_url, status
pending→pulled→fulfilled→emailed | email_failed|expired|discarded,
created/expires(+24h)/first_pulled/last_pulled/fulfilled/emailed_at,
email_attempts, failure_notified_at, next_check_at, client_ip, user_agent) +
partial index on next_check_at for the poller. Retention proposal: transcript
GC at 24h idle; null email+ip 90 days after terminal.

**Site side**: page `/gripper-report/` (`rebuild_policy='owned'`,
in_header=false, in_footer=true, nav_label "Gripper Dossier"); one section
component `gripper-report-intake` (mount div `data-gri-root
data-gri-endpoint="https://tools.apis.uk"` + explainer copy via content_data;
js_content EMPTY — 041 lesson); widget = one `js_snippets` row
`applies_to:["gripper-report-intake"]`, ≤8KB IIFE, textContent-only
rendering, honeypot + elapsed tracker; **degraded mode**: on 503 the widget
swaps to a plain 7-field form POSTing the same /submit with spec inline.
Delivery landmines: owned page ⇒ section-editor edits only; assemble-only
rerender via the 086-style direct envelope; verify served page AND served
snippets.js.

## 3. Cluster half

**Ground truths that shaped it** (verified): `rerender_single_page`
CONCATENATES stored `page_components.rendered_html` (it does not render
templates) ⇒ `create_report_page` renders final HTML in Go and writes
rendered_html + content_data on the instance row. `validate_page_content`
check 8 fails ANY number not in evidence_base ⇒ run it with
`check_claims:false` and add a purpose-built deterministic
`verify_report_prose` bound to the per-request fact block. Dispatch: clone
the **diagnose-dispatch-loop** lane (own scheduled task + custom statuses
`awaiting_report`/`reporting` + `FOR UPDATE SKIP LOCKED` claim + own reaper)
— NOT the stalled fix-loop/dispatch group (bugs_open/029). site_work_items
status/item_type are free text; `idx_swi_dedup` gives idempotent inserts on
item_key.

**New Go actions** (platform/orchestration/actions/, registered in
registry.go, all inert until image roll):
- `pull_report_requests` (report_request_pull_action.go) — mirrors
  CollectIntentEventsAction; loops sites `WHERE deploy_config ?
  'report_island'`; GET `{base_url}/requests?since=` w/ X-Internal-Key;
  checkpoint = max(submitted_at) − 2min; INSERT work items
  (item_type='report_request', handler_agent='report-builder',
  status='awaiting_report', item_key='report_request:<uuid>',
  max_attempts=1) ON CONFLICT DO NOTHING + NOT-EXISTS guard across all
  statuses.
- `score_grippers` (score_grippers_action.go) — deterministic port of
  MatchMatrix v2 physics (extracted from content_components fdfeaa7a-…,
  32,195 chars): μ table {glass .10, steel-dry .15, aluminium .20, plastic
  .25, cardboard .30, rubber .50}; dyn=m·a·S; fJaw=dyn/(μ·n); fDir=dyn;
  mEq=dyn/9.81; assessJaw (force/stroke/payload/IP, missing→unknown never
  fail), assessMagnetic (ferromagnetic gate, lower published figure),
  assessPayloadRated (vacuum/adhesive/soft + grip-range window), jaw
  conflict note (implied-μ), verdicts Match/Marginal(<1.25×)/Insufficient
  data/No match, rank+headroom sort. Reads normalised
  `products.content_data->'matchmatrix'` block (seeded from the verified
  GRIPPERS array — same figures, same provenance; runtime string-parsing of
  published spec text rejected as fragile). Output includes
  requirements+printed formulas and the **fact_block** — the ONLY
  numbers/names the writer may use; match_count==0 ⇒ mandatory first line
  `No gripper in this index meets the requirement.`
- `create_report_page` (create_report_page_action.go) — adapted from
  deploy_tool_action.go:265-349 page-creation half; NO fork (shared
  site-owned `report-dossier` component), NO CTA component (bugs_open/023);
  renders verdict summary, criteria cards, two inline SVG charts, prose
  sections, formulas panel, provenance footer (source_url + verified_date
  per candidate); INSERT pages (name='report-<uuid>',
  url='/reports/<uuid>.html', page_type='report', in_header=false,
  in_footer=false, build_status='pending', rebuild_policy='owned',
  sections=["report-dossier"]) ON CONFLICT DO UPDATE; upsert ONE
  page_components row keyed by (page_id, slot_name) lookup — never by
  remembered id (landmine).
- `verify_report_prose` (verify_report_prose_action.go) — deterministic
  gate: every numeric token in prose ∈ allowed set from scoring (formatting
  tolerance); every product/maker name ∈ candidate set; match_count==0 ⇒
  summary contains the mandatory sentence and no "meets/passes/suitable";
  no empty section. Violation ⇒ nil,error (error_step fires).
- `emit_report_status_files` — tiny helper returning `{files:
  {"reports/<uuid>.json": {status, url?, generated_at}}}` for git_commit;
  needed because the failure branch must publish a files map when
  create_report_page never ran.
- **Charts: dependency-free inline SVG** (report_charts.go, unexported
  helpers: headroom bar chart w/ 1.0 and 1.25 reference lines; payload
  scatter) — NOT go-echarts (absent from go.mod; new fleet-wide dependency
  + JS runtime tag + heavier council review for two static charts; SVG is
  golden-file-testable and cannot break at view time).

> **CORRECTED 2026-07-25 — deltas against this section, all shipped:**
> - **ONE chart, not two.** The payload scatter was dropped: published payload
>   is only comparable *within* a payload-rated technology, so a cross-tech
>   scatter would mislead. Declared to the council rather than folded in
>   silently (round 1, editquality).
> - **The chart reports its omissions.** A candidate with no comparable
>   capacity figure gets no bar (correct — inventing one is the dishonesty this
>   exists to prevent) but the omission is now NAMED on the page, so a figure
>   lost to an upstream bug is distinguishable from one never published.
> - **The renderer inlines its own CSS.** rerender concatenates stored
>   `rendered_html` and collects no component stylesheets; robot-hands.com's
>   site_specs define zero `report-*` classes, so the deliverable would have
>   shipped unstyled (bugs_open/027 shape). Two drift-guard tests hold it.
> - **NO Go template engine** in the renderers, deliberately and permanently:
>   `text/template` renders a missing field as empty with no error
>   (`missingkey=zero`), which on this page is the worst available failure.
>   Pure `strings.Builder`/`Fprintf` only.
> - **The prose gate gained** a truncation check, a vendor-trace check sourced
>   from live product data, and a `context_field`. See NOTES for the council
>   trail.
> - **`fail_workflow` is a new core primitive**, needed by the failure path
>   below: `complete_workflow` always signals success, so a handler that tidies
>   up and then completes gets its work item stamped 'complete' beside the
>   evidence it failed.

**Agent definitions** (3, all orchestrator):
- `report-request-collector` v1: pull → complete.
- `report-dispatch-loop` v1 (diagnose-lane clone): reap_stuck (reporting >
  30min → failed) → claim_item (SKIP LOCKED, awaiting_report→reporting) →
  check_claimed → spawn_handler → call_handler (timeout 1200,
  error→mark_failed) → mark_complete → notify_scheduler.
- `report-builder` v1: load_request (query_database join sites) → score
  (score_grippers) → write_prose (execute_llm_prompt, prompt_template in
  step config: fact_block is the only source of numbers/names; not-published
  figures must be said to be unpublished; no-match must be stated verbatim,
  never softened, no purchase recommendation; response_format json:
  summary_html, candidates_html, integration_html, vendor_questions_html) →
  verify_prose → compose_page → validate_page (validate_page_content,
  check_claims:false) → render_page (rerender_single_page) → check_skipped
  (skipped==true ⇒ mark_failed — a report page with no sections is a
  failure) → deploy_page (git_commit) → update_status (deployed) →
  emit_ready → publish_ready (git_commit of the sidecar AFTER the page —
  ready never precedes the artefact) → complete. Failure path: mark_failed
  (fail_work_item) → emit_failed → publish_failed → complete_failed.
  **An honest no-match report is a SUCCESS** (deploys, status ready).

**Scheduled tasks** (both enabled=false at seed; isolated concurrency
groups; interval = period − 30s tick): `report-request-pull`
(report-request-collector, interval 270, group report-pull, pre_query gates
on any site having deploy_config?'report_island') · `report-dispatch`
(report-dispatch-loop, interval 90, group report-dispatch, pre_query
EXISTS(awaiting_report OR stuck reporting)).

**Seeds** (docs/agent_docs/sql_for_agents/; ON CONFLICT DO UPDATE; NO
migrations needed — statuses/item_type free text).

> **CORRECTED 2026-07-25 — RENUMBERED 205–208 → 207–210.** Other sessions
> took 205 and 206 while this lane was building (206 twice over: both
> `206_content_gap_planner_retype_approach` and
> `206_planner_news_index_page_type`). Seed numbers are claimed by whoever
> commits first, so **re-check `ls docs/agent_docs/sql_for_agents/` at the
> moment you write one** — the number in a design doc goes stale the same way
> a figure does. 204 is unaffected.

- **204** matchmatrix normalised spec blocks into the 10 products rows
  (`content_data ||` merge, never clobber) — [pre-image OK] — *not yet applied*
- **207** report-dossier component (`component_level='section'` NOT `'tool'`,
  which keeps it out of the deployable tool catalogue;
  `render_mode='template'`; `category='report'`) — [pre-image OK].
  `created_from` must be one of `manual|generated|adopted|tool|forked` — a
  CHECK constraint the dry run found.
- **208** `deploy_config.report_island` {base_url, pull_key} — [pre-image OK,
  but see below]. The key is NOT in the file: it comes from `psql -v` and is
  bridged into plpgsql via `set_config`, echo to `/dev/null`. **Must be applied
  via stdin or `-f`, never `psql -c`** (no variable interpolation with `-c`,
  and none inside `$$ … $$` either — both verified). Applying it early is not
  free: `pull_report_requests` selects sites by this key, so the pull starts
  hitting the endpoint once the row exists.
- **209** the 3 agent_definitions — [STRICTLY AFTER the image: names six
  actions no earlier image carries].
- **210** the 2 scheduled_tasks, both **disabled**, in isolated concurrency
  groups; `enabled` deliberately not overwritten on conflict — [after image].

## 4. Build order

1. Go code (5 actions + charts + registry + table-driven tests incl.
   physics fixtures, no-match fixture, verify_report_prose
   rejects-invented-number/name/softened-no-match, SVG golden files) —
   **council review** on the platform diff; commit per task w/ pathspec.
2. Island service code + tests (offline: fake Anthropic, file-mode SMTP);
   dockerfile; local compose smoke.
3. Seeds 204, 205 (name no actions).
4. Image: IMAGE_TAG bump, build from committed HEAD, roll; discriminating
   pod-grep on literals the change CREATED ('report-dossier',
   'pull_report_requests') + positive control; ~300s no-dispatch window;
   verify spawned handler pods pick the new tag (bugs_open/066 class).
5. Seeds 206–208 (tasks stay disabled).
6. Induced E2E WITHOUT the island (§6 fixtures, direct-fire kcat per 049b
   pattern) — all three branches observed live.
7. Island deploy (tenant-2: role+DB, .env, scp compose/Caddyfile, docker
   load, up; on-box health, then public health via tunnel; one manual
   backup run; RUNBOOK ledger).
8. Pull key into deploy_config (206 with real values); authed-curl pull
   proof + 401 proof.
9. Site-side SQL: snippet + component + owned page + page_component;
   assemble-only rerender; verify served page + snippets.js symbol.
10. Full E2E from a real browser (real intake → island → pull → build →
    sidecar → email arrives, link works, SPF/DKIM pass); failure drill
    (expired request → apology + operator alert); abuse drills (injection,
    honeypot, cross-origin, outage→plain-form).
11. Owner checkpoint on the sample dossier + prompt + privacy posture →
    enable report-dispatch, then report-request-pull → watch first real
    pull. Soft-launch unlinked (recommended) → footer nav link.

## 5. Seam reconciliation (where the two halves disagreed — RESOLVED)

1. **Pull contract**: island path wins, cluster parsing wins. Endpoint =
   `GET {base_url}/requests?since=<RFC3339>` where
   `deploy_config.report_island.base_url =
   "https://tools.apis.uk/api/gripper/v1"`. Response = **NDJSON** lines
   `{id, host, submitted_at, spec:{…}}` + trailing `{"_meta":…}` (the
   intent-collector stream contract the cluster action already mirrors).
   Auth X-Internal-Key. All requests since checkpoint regardless of status;
   cluster dedups by item_key; island sets first/last_pulled_at +
   pending→pulled on first pull. NO email in payload.
2. **Completion signalling**: the cluster's status sidecar wins over the
   island half's poll-the-HTML-page. Island polls ONE url:
   `https://robot-hands.com/reports/<uuid>.json` (2min, then 5min×1h, then
   15min; cache-buster + no-cache; 15s timeout). `ready` (+url) → link
   email → emailed; `failed` → apology; neither by expires_at (24h) →
   apology anyway (covers cluster-dead and reaped-pod gaps). Sidecar
   commits AFTER the page commit. Island's report_url column = base +
   `/reports/<uuid>.html` (page URL scheme: `.html`, matching pages.url
   platform convention — the island half's `/reports/<id>/` directory form
   is dropped).
3. **Spec mapping** (chat fields → physics inputs): deterministic and
   CLUSTER-side (in score_grippers input handling), island stays dumb:
   material string → μ key (fixed list the chat prompt offers); environment
   → ip_min int; part geometry dims → travel_mm; cycle_rate → safety-factor
   tier (2/3/4); budget + mounting → prose context ONLY (no numbers
   derivable; products.price unpopulated). Exact tier table = owner
   question §8.3; frozen before the chat prompt is written so the prompt's
   field list matches.

## 6. Test fixtures (cluster E2E without the island)

Hand-INSERT work item + direct-fire report-dispatch-loop (kcat, 049b
pattern). (1) **Success** (2.5kg steel, a=12, S=2, IP54): page 200 live,
contains the substituted formula literal `(2.5 × 12 × 2) ÷ (0.15 × 2)` (a
string no other page carries), sidecar ready, absent from nav/listings. (2)
**Honest no-match** (500kg, IP67, glass): deploys as SUCCESS containing
verbatim `No gripper in this index meets the requirement`, zero
Match/Marginal in content_data, no purchase recommendation —
verify_report_prose unit test proves the gate bites. (3) **Induced failure**
(mass_kg="not-a-number"; separately a sabotaged write_prose in a test copy):
item failed, NO page committed, sidecar `failed` live — per
verify-the-failing-branch, the lane is not "verified live" until the failed
sidecar has been observed. Cleanup: source='manual-test' rows deleted, test
files git-rm'd (owner approval §8.9).

## 7. Cost & risk envelope

Chat ≈$0.15/session worst case, ~$20/day global cap, spend-capped dedicated
key; report build = one Haiku-class prose call + deterministic everything
else; island compromise blast radius unchanged in kind (leads + 2 capped
keys + 1 mailbox cred; zero cluster credentials); prompt injection
structurally contained (schema-forced output, nothing executed, textContent,
spec untrusted downstream); no-match/fabrication risk gated by
verify_report_prose + the no-invention prompt + provenance footer.

## 8. Consolidated owner questions

> **ANSWERED 2026-07-24 (owner):** Q1 sender = **shared sender**
> (robot-hands@contactforsales.com, fleet convention). Q9 live fixtures on
> prod = **approved** (delete artefacts after). Q6 launch = **soft-launch
> unlinked**, footer link only after real-traffic confidence. Q2 key cap =
> **~$50/month** (owner issues the key → /opt/island/.env
> GRIPPER_ANTHROPIC_API_KEY). Defaults accepted without objection: Q3
> mapping table as proposed, Q5 UUID + noindex (no robots.txt entry), Q7
> retention (24h transcript GC, 90d email/IP null), Q8 max_attempts=1.
> Remaining owner ACTIONS (not questions): issue the $50-capped key; confirm
> the tunnel authorisation click. Q10 (sitemap in GH Action) = ours to check.

1. **Sender address**: robot-hands@contactforsales.com (fleet shared-sender
   convention) or new reports@robot-hands.com mailbox (+SPF/DKIM)? (Also
   incidentally closes the site's contact_form_undeliverable gap.)
2. **Second spend-capped Anthropic key** for gripper-intake — issue at what
   cap (suggest ~$20/mo)?
3. **Spec-mapping table sign-off** (§5.3): cycle_rate→safety-factor tiers,
   geometry→travel, material list. Blocks freezing the chat prompt + scoring
   input contract.
4. **Tunnel status**: has the Cloudflare tunnel authorisation click happened
   (RUNBOOK said awaiting, 2026-07-24)? Blocks all public access.
5. **Report privacy**: UUID URL sufficient? Add robots.txt
   `Disallow: /reports/` (reveals the prefix exists)? noindex meta is free —
   proposed yes.
6. **Launch posture**: soft-launch unlinked until E2E passes (recommended),
   then footer link?
7. **Retention**: island transcripts GC 24h-idle; email+IP nulled 90 days
   post-terminal; report pages accumulate with no TTL for the pilot —
   acceptable?
8. **max_attempts=1** (failure ⇒ apology, retry is manual re-queue) —
   acceptable at pilot volume?
9. Approval to run the three live fixtures against production
   robot-hands.com and then remove the artefacts.
10. Does the static-site GH Action generate sitemap.xml/RSS from committed
    files (cluster has no sitemap emitter)? If yes, /reports/ needs
    excluding there.
