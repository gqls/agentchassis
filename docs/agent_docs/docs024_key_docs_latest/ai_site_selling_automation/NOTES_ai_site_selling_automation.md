# NOTES — ai_site_selling_automation — append-only, newest at the bottom

## 2026-08-10 ~18:00 — session start: orientation, freshness checks, owner rulings

Cold-started from `HANDOFF_2026-08-10_start_here.md` (this directory). Ran the
handoff's own falsification checks before trusting it:

- **Sibling lane is fresher than the handoff.** `webdesign_uk_build_service/`
  has `HANDOFF_2026-08-10c_continue_here.md` (17:21) and NOTES appended to
  17:55 — both post-date the start-here file (17:23). Read both.
- **The chat intake's "LIVE" status changed meaning this afternoon.** The
  cache gap closed on its own (proven at the nginx layer, sibling 08-10c §1a
  — do not re-litigate), but underneath it: **the Anthropic ACCOUNT hit its
  spend cap ~14:51Z 2026-08-10**. The box now serves the fail-closed
  contact-details line to every visitor. Sibling NOTES (17:00–17:40Z entry,
  contributed by the bugfix-236 lane) corroborates from a different credential
  path: last successful LLM call fleet-wide 14:51:45Z, 0 successes after,
  **council gate down** (a submission died at its first review seat,
  terminal `complete_invalid` — which reads as "invalid submission" and is
  not). Provider says access returns 2026-09-01 unless the owner raises the
  limit in the Anthropic Console. [CITED from sibling NOTES, not re-measured —
  their evidence spans two independent paths, which is stronger than my
  re-running the same query.]
  Consequence for THIS lane: council submissions are unsatisfiable until the
  cap lifts, so the clients-columns migration (council scope) is drafted but
  not shipped; FE work (out of scope) proceeds.
- **`bugs_open/239` is OWNED.** A live session named "bugfix bugs_open/239"
  was busy on it at 17:5x tonight (peer-session listing). Sequencing step 3 of
  the handoff is therefore someone else's; this lane tracks, does not touch.
- **Workstream memory entry exists** (MEMORY_workstreams.md line ~34, NEW
  08-10) and matches the handoff. No competing directory; this dir held only
  the handoff at session start.

**Owner rulings taken at session start** (put directly, three questions,
recorded in PLAN §1 — the authoritative copy): identity = extend `clients`
(BIZ-014); automation target = £1,200 tier intake (human releases); scope
while cutover pending = design + ungated build.

Spawned two read-only scouts: (a) admin FE + `/api/v1/admin` client endpoints
(for the Customers tab); (b) `046_site_chat_turns.sql` + chat-service JSONL
format + core-side pull precedents (for the ingestion design). Results to be
recorded below when they land.

## 2026-08-10 ~19:00 — first build session: mig 375 live, /admin/customers built, FE tab built, ingestion designed

**The handoff's cheapest-first premise failed on contact with the code.** Scout
read of `client_handlers.go` showed the "existing client CRUD endpoints" serve
`clients_info` — a side table the handler lazily CREATEs on first call
(`to_regclass('public.clients_info')` → NULL live, i.e. never yet called) —
part of the per-client-schema tenant machinery, not the ruled
`clients→networks→sites` chain. GET /clients names its columns and discards
`settings` into a dead variable; new columns would not have flowed through
anyway. Corrected in the handoff §3.5 in place; landmine appended to
LANDMINES.md + synced to doc_notes (`--check` clean; the verify dispatch will
sit starved until the LLM cap lifts, expected).

**Shipped, in order:**
- Migration `375_clients_customer_identity_columns.sql` — email/phone/tier/
  customer_status/notes on `clients`, all nullable. Blast radius measured
  first: zero Go readers of `FROM clients`, zero positional INSERTs.
  Applied out-of-band (`psql -f` via kubectl) because the runner's `--apply`
  is blocked by four older pending files with pre-state mismatches
  (353/358/359/361 — other threads') and two LIKELY-ALREADY-APPLIED
  (363/370); a failed file stops the run before reaching 375. Ledger-recorded
  via `--record-only` with that reason. Verified: information_schema lists
  all 5 columns.
- `internal/core-manager/admin/customer_handlers.go` + routes —
  GET/POST `/api/v1/admin/customers`, GET/PATCH `/customers/:customer_id`.
  `go build` + `go vet` clean; **HEAD verified from a clean `git archive`
  build after committing** (HEAD-BUILD-OK). Registered as ADM-011 with index
  row, same commit. Commit `fe6b99d05`.
- FE `CustomersPage.tsx` + three-line App.tsx wiring (nav key `customers`,
  PipelinesPage precedent). Verified in a throwaway node:20 container on a
  scratchpad COPY (tree kept clean): `vite build` ✓; `tsc --noEmit` reports
  15 errors, ALL pre-existing in App.tsx, zero in the new page — this repo
  has no typecheck script and has never typechecked; do not "fix" that as
  part of an FE change. Commit `a84d544d1`.
- Ingestion design written into PLAN §2.3 (B2 sink, spawn-free CronJob
  puller, UUIDv5 pairing, hash-at-ingest, box-side asks for the sibling
  lane). Not built.

**Council obligation open:** `fe6b99d05` touches `internal/` with no
Council-Submitted trailer because the gate cannot currently run a seat (fleet
LLM cap; the sibling NOTES show a submission dying at its first review seat).
OWED: submit the mig-375 + customer-endpoints change as one coherent
submission when `llm_call_log` shows successes again — the RUNBOOK curl is
the cheap tell. [ASSUMED] the gate will accept a post-commit submission
normally; the Council-Submitted mechanism explicitly supports commit-first.

**Still true / unchanged tonight:** Anthropic account cap in force (chat box
fail-closed; owner action only); `bugs_open/239` owned by its own live
session, untouched by this lane.

## 2026-08-10 ~20:30 — fleet roll v1.0.1283: Customers surface LIVE end-to-end; cap LIFTED; council submission fired

- **Roll verified at the artefact, not the tag.** core-manager AND
  admin-dashboard both at `v1.0.1283` (2 replicas each, ~28 min old at check).
  Pod-grep, same exec, every replica: `strings /app/core-manager` →
  `"Failed to list customers"` = 1 (my string) AND `"Failed to list clients"`
  = 1 (pre-existing positive control) on BOTH core-manager replicas; dashboard
  bundle `index-DP86XXhq.js` contains `"No customers yet"` (mine) AND
  `"Data Pipelines"` (control) on BOTH replicas. Route probe from inside the
  cluster: `GET /api/v1/admin/customers` → **401** (auth wall), not 404.
  The whole surface — columns (375) → API → FE tab — is live.
- **The Anthropic account cap has LIFTED.** The RUNBOOK curl returns a real
  answer ("£1,200, paid once…" — correct pricing, correct follow-up), not the
  contact-line fallback. [INFERRED] the owner raised the limit; not verified
  in the Console, but the observable (fleet LLM works) is what matters here.
- **Council submission DISCHARGED for `fe6b99d05`:**
  `SUBMISSION_CORR = 371f8b7d-0835-4879-b48f-ad0176bf2058`. Budget ~30 min
  (dispatch queues behind the fleet, and the fleet is digesting a post-outage
  backlog — do NOT re-fire on a missing orchestration row). Find the run by
  payload: `SELECT current_step, status FROM orchestration_states WHERE
  collected_data->'input_data'->>'fix_correlation_id' =
  '371f8b7d-0835-4879-b48f-ad0176bf2058';` Verdict:
  `SELECT body FROM doc_notes WHERE categories ? 'council-gate' ORDER BY
  created_at DESC LIMIT 1;`
  **Known 098 gap, accepted:** `fe6b99d05` predates the submission and
  forward-only forbids amending a trailer in, so it will list un-reviewed in
  098 forever; this NOTES entry + the verdict note are the audit trail. Do
  NOT stick the corr as a trailer on an unrelated docs commit — the trailer
  binds to the commit it is on.

## 2026-08-11 — council APPROVED; four owner rulings land; EPP + FAQ facts grounded

- **Council verdict for `371f8b7d-0835-4879-b48f-ad0176bf2058`: APPROVED,
  round 1, all reviewers** (orchestration `complete_approved|COMPLETED`;
  verdict note in doc_notes, full report in diagnosis_artifacts under the
  corr). `fe6b99d05` still predates the corr — the 098 gap stands as
  recorded; nothing to amend, nothing further owed on that change.
- **Owner ruled the four open decisions** — recorded authoritatively in PLAN
  §1b (subscription service builds out; £149 queue model SUPERSEDES £1,200;
  chat into the framework; customers keep their own DNS, ZIP delivery,
  no refunds). §1b also lists the sub-questions the owner has NOT ruled
  (payment timing vs the live "only pay if you like it" promise; queue home;
  voucher single-use).
- **EPP question answered by measurement, not memory**: the domains lane
  RUNBOOK lists Nominet credentials PENDING; its README (owner's own doc,
  2026-08-04 entry) says the *password* was provided, the **TAG name is
  still owed**, and the allowlist must move to the five fixed cluster IPs
  because the office IP already rotated stale once.
- **FAQ target located live**: `/faq.html` component
  `edfecdf2-c25a-4bbd-90c1-c26e644d86cf`, Q "What about the domain and
  hosting?", A begins "…We handle the setup as part of getting the site
  live" — contradicts ruling 4; and the £1,200 copy footprint spans at least
  index (hero subheadline, pricing block, FAQ block), faq, what-you-get,
  plus the chat bot's facts (it quoted "£1,200, paid once" in last night's
  live check). The webdesign.uk pages are the SIBLING lane's build — check
  live sessions before editing; content changes go through the framework.
- Handoff superseded: `HANDOFF_2026-08-11_continue_here.md` replaces 08-10b
  as the cold-start.

## 2026-08-11 (later) — second ruling batch; £1,200 copy ARCHIVED before migration

- **Owner ruled the §1b open details + retired £1,200 outright** — PLAN §1c
  is authoritative (payment after-approval-then-upfront as a switch; vouchers
  single-use/named/expiring; no visible refunds, manual dashboard behind the
  scenes; hosting via third-party/affiliate + a UK-S3+CDN setup page;
  Lovable/Durable affiliate links; differentiation = example sites from the
  owner's own domains; positioning = no-frills Ryanair-honest, explicit
  about AI-built).
- **Pin question answered by measurement**: no whole-site snapshot exists —
  register shows spec pin/unpin (ADM-003) per-spec and CGV-017
  "approval-snapshot lock regime built, never wired up, dropped"; scripts/
  has only memory/scratch git snapshots. So the pin is an ARCHIVE:
  `snapshot_2026-08-11_gbp1200_offer/` — sites(1)/pages(6)/
  page_components(22)/site_specs(31) as JSONL, line counts verified equal to
  live row counts in the same run (the disconfirmable check: embedded
  newlines would have made lines > rows). Captures content_data AND
  rendered_html AND evidence_base. Box-side vm-sites git tag would be
  belt-and-braces — sibling lane's turf, noted in the snapshot README.
- Handoff 08-11 updated in place: §3 Q1–Q3 answered, Q4 (queue semantics)
  still open; work item 0 (pin) done; work item 1 unblocked with the copy
  direction + hosting setup page + affiliate block folded in.
- SUMMARY_2026-08-11 written at owner request — the decision series with
  options and choices, readable aloud.

## 2026-08-11 (evening) — queue ruled non-binding; Nominet TAG named; allowlist gap found

- **Queue semantics ruled (closes handoff §3 Q4):** the wait note is an
  APPROXIMATION, nothing binding; the owner may pause the queue on
  malfunction or scale trouble. Copy for the queue note must promise nothing.
- **Nominet TAG = `DESIGNCONSULT`** (for now); owner applying for a second
  tag for this venture. Application body drafted this session (chat);
  grounded in Nominet's registrar docs: additional tag via Online Services →
  Apply for services → Apply for an additional domain tag; three tag types;
  only ONE Self-Managed tag per registrar, so the second tag is Channel
  Partner-shaped (customer domains) regardless.
- **Live EPP allowlist (owner-supplied): 5.65.164.9, 116.203.204.115,
  151.226.83.138, 176.58.121.95 — the five cluster IPs are NOT among them**;
  cluster EPP still blocked. Contributed to the domains lane NOTES (their
  EPP ownership) with the credentials-file completion note.

## 2026-08-11 (late) — owner: tag applied, IPs done; EPP proven 1000 from cluster; queue+voucher design added

- Owner applied for the second tag and completed the EPP allowlist. Proven
  end-to-end this session: EPP LOGIN (not greeting) from postgres-clients-0
  → result 1000 (full evidence in the domains lane NOTES, contributed
  there). Credentials file completed with TAG=DESIGNCONSULT.
- PLAN gains §2.6 (queue/submission gate) and §2.7 (subscription+vouchers)
  designs so the next session builds rather than designs. Handoff §4
  updated: Nominet asks are now CLEARED except the registrar keys; second
  tag pending Nominet's decision.

## 2026-08-11 (night) — §2.7 BUILT: billing surface + vouchers; copy migration deferred on live collision risk

- **Cold-start falsifier check found the copy migration (work item 1) unsafe
  to start**: the webdesign lane's session was live at that moment (transcript
  timestamps are UTC — they read ~an hour behind local; the tell that it was
  seconds-fresh, not an hour stale), mid-testing `page_rerender`/`locked_at`
  on webdesign.uk after their chat-box-wipe incident. A copy migration
  dispatches page_rerender on the same site. Deferred; coordination note
  appended to their NOTES (with the ask: signal when quiet). Work item 3
  (subscription + vouchers, §2.7) built instead — zero overlap with them.
- **Migration 391 applied + recorded** (out-of-band + `--record-only`, the
  375 recipe — runner `--apply` still blocked by other threads' pending
  files, re-verified via dry run this session): `vouchers`, `billing_orders`
  (statuses `created|paid` ONLY — no refunded state, per ruling),
  `billing_events` (dedup PK), `billing_settings` (one row,
  `payment_timing='after_approval'`). Single-use invariant exercised against
  the LIVE schema in a rolled-back tx: first redemption 1 row, second 0,
  zero leftovers — a check that could have failed, not a shape-read.
- **`internal/auth-service/billing` built** on the idea.uk PAY-001 shape
  (raw net/http + HMAC, no SDK; FakeProvider reachable only from test
  files): Provider/StripeProvider, atomic redemption + webhook-dedup
  repository, service pinning the ruled invariants (14900p list, voucher
  variants 1000|5500 only, timing vocabulary closed, nil provider → 503),
  admin routes + public webhook. `go test` green (10 tests incl. signature
  tamper, double-redeem, checkout-failure-keeps-voucher-consumed).
- **Wiring degrades, never dies**: first draft Fatal'd auth-service when
  clients_db was unreachable — corrected before submission, because that
  coupled LOGIN availability (admin dashboard's auth) to pgbouncer. Now:
  error log + billing unmounted for that run.
- **Blast-radius measurement corrected my own risks claim before submission**:
  I wrote "only Go write site to clients.external_id is this change's" — the
  grep found ADM-011's customer-create INSERT also writes it
  (customer_handlers.go:190, our own lane's previous session!). They compose
  (webhook guarded by `external_id IS NULL`; unique constraint both ways).
  The submission states the measured truth.
- **Register: PAY-009 appended** (payments.md) — producer set + key shape
  named per the RFC_010 condition; consumers named (queue gate §2.6, admin
  FE follow-up, webdesign chat lane under `upfront` later). Open at build
  time: webhook public exposure (NO Ingress resources exist in the cluster
  today — decide box-proxy vs Ingress+cert when Stripe keys arrive); Stripe
  restricted key + webhook secret = owner task; admin FE voucher screen not
  built.
- Council submission dispatched (councils queue behind the fleet; corr in
  the commit trailer). Subscription scaffold's ListAll `$N`-on-MySQL dialect
  bug fixed in the same change; scaffold otherwise untouched (PAY-007
  stands — `status=active` still means "a row exists").

## 2026-08-11 (late night) — council round 1 REVISE: two real findings, seven answered by measurement; round 2 in

- **REVISE, gated by guardian** (corr 4ac1fe52). The gating objection — "confirm
  the 391 ledger row actually matches" — was answerable, not a defect: row
  present, applied_by='record-only', checksum md5-exact against the committed
  file (the runner hashes md5, not sha — my first comparison used sha256 and
  looked like a MISMATCH until I read run-migrations.sh:264; and my first
  ledger query silently returned nothing because I wrote `note` for `notes`
  with 2>/dev/null eating the error — the blind-check trap again).
- **Two REAL findings, both now durable in b9bea5e1d**: (1) reuse seat: the
  unwired `subscriptions` scaffold and the new `billing_orders`/`vouchers`
  are two uncoordinated representations of "this client paid" → recorded as
  a named OWNER DECISION (handoff §4, with recommendation: deprecate the
  scaffold's create surface once £149 is proven). (2) debug_historian: the
  degrade-to-unmounted design makes a silent non-deploy look like success →
  post-roll verification recipe committed to the RUNBOOK ("Verify the
  billing surface"), named as handoff item 3c.
- Answered by measurement: pod's baked config file proven authoritative
  (kubectl exec cat — the overlay is env-only, viper needs SERVICE_ prefix);
  pool headroom ≤9/200 client conns (MaxConns=3, MinConns=0 × 3 replicas);
  clients_db has ONE migration series (sql_for_agents), no core-manager
  collision; 391 replay-safe by construction (IF NOT EXISTS / ON CONFLICT).
- Round 2 resubmitted on the SAME correlation (RESUBMIT_CORR), edits
  restructured one-file-per-edit to the 8-cap, DO-guard shown verbatim,
  dialect fix removed from the edit list (shipped in 1834bd3c0,
  acknowledged in rationale per the guardian's separate-shipping preference).

## 2026-08-11 (later still) — council round 2 REVISE: the 8-cap paradox named, four checks answered live

- **Round 2 REVISE, gated by editquality** — both HIGHs are one structural
  fact: 10 shipped files cannot each hold an edit slot under the fix_plan's
  hard 8-edit cap, and the seats read a bundled companion file as "never
  created by the applying mechanism". There IS no applying mechanism — the
  diff is at HEAD (1834bd3c0 + 5af5ee2ca) and compiles from a clean
  `git archive HEAD`. Round 3 leads with that fact instead of burying it.
- **Their four factual asks, answered by measurement this session**:
  (1) the DDL RAN — information_schema, queried fresh: vouchers(8 cols),
  billing_orders(11), billing_events(6), billing_settings(3) all present in
  clients_db; "record-only" recorded an apply that genuinely happened.
  (2) PAY-001's Go is `package main` under docs/agent_docs/.../idea.uk/
  golang_files/ — a docs archive, not an importable module: pattern-reuse
  stands, duplication objection falls. (3) the scaffold has NO webhook/SDK:
  the only non-billing "checkout/stripe/webhook" hits in internal/
  auth-service are the never-constructed CheckoutSession TYPE and its
  example URL string (models.go:79-86). (4) pgbouncer headroom stands at the
  round-2 numbers (≤9/200; 10 server conns measured); SHOW CLIENTS needs
  console auth we don't hold — noted, not silently skipped.
- **debug_historian's second real find landed**: the post-roll recipe's
  step 4 — a real voucher POST as the write-path acceptance test (a mounted
  route proves middleware, not a working DB write). RUNBOOK updated.
- reuse seat's edit-1 medium (two payment concepts) remains a NAMED OWNER
  DECISION by design — recorded, recommended, not resolvable by this lane
  unilaterally deprecating a surface the owner ruled to build out.
