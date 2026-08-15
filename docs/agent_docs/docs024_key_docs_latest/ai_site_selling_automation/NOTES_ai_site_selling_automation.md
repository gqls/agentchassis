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

## 2026-08-11 (close of session) — council rounds 3+4 REVISE; stopped by design at four; the trail stands

- **Round 3 gated by prior_art** (shipped-at-HEAD unverifiable from the SQL
  tier), **round 4 by reuse_agent** (the two-client-stores question). Round 4
  ran with the runbook's own remedy applied — the nine verification outputs
  in a doc_notes row (id 22a7ee9e, subject 'decision'/
  'council-submission-4ac1fe52') — and one seat read it and discounted it as
  prose. Objecting seats GREW round over round (3 → 6, architecture newly
  firing): fresh panels have no cross-round memory, and a shipped-code
  submission's central claims all live outside their queryable schema.
- **The reuse HIGH is answered by the owner's own ruling, on the record**:
  migration 375's committed header — customer identity lives on
  clients → networks → sites (ruling 2026-08-10, BIZ-014 shape); Stripe
  linkage stays on clients.external_id; and clients_info is the tenant store
  the OLD /admin/clients endpoints read (this lane's own landmine, synced).
  billing_orders/vouchers FK against clients(id) is the ruled chain, not a
  store mix-up. ADM-011 (the Customers surface) reads the same chain.
- **Architecture seat's "deserves an RFC"**: per the 2026-07-29 ruling an RFC
  triggers on a change to what a shared mechanism GUARANTEES; this adds a new
  registered mechanism (PAY-009) with named consumers. Its objection to
  after-the-fact review is an objection to the ruling that made review
  after-the-fact the design — routed to the owner via this note, not
  relitigated in rounds.
- **Stopping at four rounds, per the estate's own norms**: the gate is
  ADVISORY (owner 2026-07-24); "one council run per coherent task, not per
  iteration"; and every real finding of the four rounds — three of them —
  is landed and committed (b9bea5e1d, 895029d24, plus this note). The
  Council-Submitted trailer stays honest: 098 lists the commit as
  un-reviewed until an approval resolves it, which is the true state.
  Editquality's round-4 note that 391's verify block does not assert the
  vouchers/billing_orders columns is true and unfixable in place (editing an
  applied+recorded migration breaks the ledger checksum); the
  information_schema check in the RUNBOOK covers it, and a future 39x may
  add a standing assertion if wanted.
- Memory lesson filed: after-the-fact-council-review-of-shipped-code-loops
  (cap such reviews at ~2 rounds once real findings are landed; submit
  before-or-alongside the shipping commit to give seats a judgeable plan).

## 2026-08-11 (evening, post-roll) — billing surface VERIFIED LIVE on the fresh build, keyless by design

- Owner deployed a fresh build. Post-roll recipe run per the RUNBOOK:
  (1) provenance stamp = bb5348642, and `git merge-base --is-ancestor
  1834bd3c0 <stamp>` confirms the billing commits are aboard;
  (2) startup line = "billing mounted without a payment provider —
  STRIPE_SECRET_KEY / STRIPE_WEBHOOK_SECRET not set" (pool opened, mounted,
  keyless — the expected outcome of the three);
  (3) routes probed from core-manager via wget (curl absent in that image —
  the recipe's curl needs swapping for wget): admin settings → 401 (route
  exists behind auth), webhook POST → 503 (mounted, refusing keyless).
  Step 4 (real voucher POST) needs an admin JWT — owner-side, owed at first
  dashboard use.
- PAY-009 status advanced: built → **deployed (keyless)**. The go-live
  remainder is entirely owner-side: Stripe keys + webhook exposure decision.
- Sibling lane still live (19:04 activity; lock verification now against
  REAL rerender dispatches, a3240564a) — copy migration stays deferred.
- `bugs_open/239` gained a fail-closed fix commit (a097e3e26, their lane) —
  re-check fixed-AND-LIVE state next session; it gates the trigger seam (P4).

---

## 2026-08-12 — copy migration, half one: the £149 register is LIVE and detection is ARMED

The source-of-truth half of work item 1. `evidence_base` for webdesign.uk now
describes the £149 offer; the page copy is regenerated separately, next.

> **CORRECTED — the handoff's description of the live site was STALE, and I
> checked before acting on it.** `HANDOFF_2026-08-11b` §2.1 says the live site
> "still says '£1,200 is the total price', 'you only pay if you like it'".
> Measured 2026-08-12 at the served artefact: **it says none of those things.**
> The sibling lane re-priced the site to a £75-deposit model on 08-10
> (`da5fb0d`, `f4e77c7fb`), and **no price at all now appears on the served
> pages** — the homepage says "One fixed price" and "Tell us what you need.
> We'll tell you what it costs". What IS live, and false under the 08-11
> rulings, is the deposit/refund/window/two-rounds family. The retired £1,200
> copy survives only on the **archived** `index-rejected-v1-20260806` page.
> A handoff written on 08-11 evening was already describing an 08-10 state:
> the cheap check was one curl, and I ran it because the DB disagreed.

- **`webdesign.uk` 302s to `webdesign.co.uk` and always has (intentional,
  owner-confirmed 2026-08-10, `47615b5e1`). The shopfront serves at
  `preview.webdesign.uk`.** Every path on the apex collapses to the co.uk
  homepage, which is a different site (tools and guides). Anyone verifying
  "the live site" on the apex host measures the wrong site and sees no offer
  copy at all. The `-L` follow is what makes it obvious; without it you get a
  302 and 143 bytes.
- **Scope is FIVE surfaces, not the four the handoff lists.** Scanning all 25
  components rather than the pages I had in mind added
  `tool-website-brief-starter-guide` (`article-body`), which carries the full
  £75 / 14 days / two rounds terms inside a guide. Enumerating the pages I
  knew about would have shipped a migration that left the old offer live in
  prose nobody was looking at.
- **Migrated by SUPERSEDE, not in-place UPDATE** (`SQL_2026-08-12_evidence_base_149.sql`).
  The 08-10 deposit change was an in-place UPDATE and destroyed the row before
  it; the owner ruled the £1,200 offer "archived restorably", and a superseded
  row makes that true in the database and not only in a file snapshot. Old row
  `bccf42a7` (`is_current=false`, 10 facts / 18 bans); new row `6f9e8e7c`
  (12 facts / 26 bans / 4,774-char writer_block), `pinned` inherited.
- **What the register now says**: £149 total with no VAT; payment after the
  customer approves the site; **no refund**; **one set of changes**; only a few
  sites at a time with no number published; **the site is AI-built, stated
  plainly**; delivery is a private preview link then a ZIP the customer hosts;
  hosting and the domain are not included and stay with the customer.

### Two defects caught BEFORE this shipped, both by running the tool rather than reasoning

- **`"value": "after_approval"` would have silently disarmed the claims checker
  for the entire site.** `EvidenceFact.Value` is `*float64` in Go, so a string
  there fails the whole `evidence_base` unmarshal, `ParseEvidenceBase` returns
  nil, and the checker switches off **with no error anywhere** — this lane's
  own documented landmine, which I walked straight into and `cmd/claimscan`
  caught on the first run. The setting now lives in the claim text and in
  `source.sql`, where it is checkable.
- **My own verify block would have aborted the migration wrongly.** I wrote
  "abort if £1,200 appears in the new document" — but the retired price
  legitimately appears throughout the new `banned_claims` reasons, which is
  where it is named as the thing not to say. The assertion now checks the
  **facts** only, and is paired with a positive one (`price_total = 149`),
  because a check that only forbids cannot tell a correct document from an
  empty one.

### The negation-guard trap, measured both ways

`negationCueRe` **deliberately excludes a bare "no"** (it is an intensifier in
marketing prose: "there are no exceptions: every claim is verified"). So under
the new `refund` ban, **"there is no refund" IS flagged and would block the
page**, while **"we do not offer refunds" is correctly suppressed**. The copy
must use `do not` / `never` in the same clause. This is written into the ban's
own `reason` field and into `writer_block`, so the next writer meets it at the
point of use rather than at the point of failure.

### The measurement, and what could have come out otherwise

- **Ban set proven NON-INERT**: same 25 live components, same engine —
  old register **3 findings**, new register **36** (faq 8, how-it-works 6,
  brief-starter-guide 4, what-you-get 3, index 1, archived page 14). A ban set
  that returned the same 3 would have been decoration.
- **Intended replacement copy proven CLEAN**: 19 sentences built from the new
  `writer_line`s plus the FAQ answers the offer must be able to give →
  **0 findings, 1 suppressed** (the refund sentence above). `[LIMIT]` This is
  a fixture I composed, so it proves the ban set does not block the phrasing I
  intend; it does **not** predict what the writer will produce. That is what
  the regeneration + a re-scan of the real output has to establish.
- **Verify block proven to DISCRIMINATE**: both assertions were run against the
  old row first and both raised (`facts: expected 12, got 10`;
  `fact price_total is not 149`) — so a green run means the write landed, not
  that the check was asleep.
- **Checked at the READER, not the writer**: the row was pulled back out of the
  DB and re-scanned through the platform's own parser, reproducing 36 / 0
  exactly. A jsonb round-trip that had mangled anything would have shown here.

### Owner asks this raises (none blocking; all in the handoff's ask list)

1. **`build_duration` (three to four days) is CARRIED OVER unre-attested.** It
   was attested on 08-04 for the £1,200 offer; the 08-11 rulings did not
   restate it. It is in the register, flagged in its own `source`, and needs a
   yes or a new number.
2. **`changes_paid_defects_free` was DROPPED, not carried.** "Anything that is
   our own mistake is fixed at no cost" was a 07-29 fee-boundary ruling under
   the £1,200 offer; at £149 with no ongoing service it is an open-ended
   liability, so it is not in the register and the site will not say it.
   Silence is publishable; put it back if the owner wants it.
3. **The payment-timing switch has a COPY dependency the ask list does not
   record.** Handoff §1.4 calls it "yours to flip, no build needed". Flipping
   `payment_timing` to `upfront` makes fact `payment_after_approval` false and
   with it every page that states it. The fact now names
   `billing_settings.payment_timing` as its source so the coupling is visible,
   but nothing enforces it: flipping the switch is a copy migration, not a
   one-field UPDATE.
4. **No queue capacity number is published** (the ruling says start at 3–4, the
   counter is not built). The copy says "only a few sites at a time".
5. **ZIP delivery is promised in the register while work item 4 is unbuilt** —
   fulfilment today is manual (pull the B2 prefix, zip it). Deliverable, but by
   hand.

### Two things for the fleet, not this lane

- **The `honest` ban is now enforced MECHANICALLY on this site.** The
  fleet-wide instruction of 2026-08-12 (`claude-ideauk-copy-20260812`, applied
  to 14 sites' `content_direction`) is prompt-side and advisory; a
  `banned_claims` pattern is the enforcing layer, exactly as this lane's own
  superlative ban was on 08-06. Site-side only, deliberately: banning it
  fleet-wide is that lane's call, not mine.
- **The claims checker files work items against ARCHIVED pages.** 14 of the 36
  findings are on `index-rejected-v1-20260806`, which is `status='archived'`
  and served to nobody; there is already an open `claims_unverified` item for
  it. Pre-existing, not caused by this change, and not fixed here — but this
  change quadruples the noise, so it is worth someone's attention.

---

## 2026-08-12 — copy migration, half two: the `faq` canary is LIVE and correct, and it exposed a hole the ban set cannot see

Two `content_rewrite` items with `spec.mode='edit_live'` (`SQL_2026-08-12b`,
commit `0854162c6`). Dispatch was **fast, not queued**: created 16:37:47Z,
claimed by `build-dispatch-loop` at 16:39:14Z, `faq` complete at 16:43:54Z.

> **CORRECTED, and the correction was one query.** I read
> `find_dispatchable_site` (it orders `wi.created_at ASC` before priority) and
> concluded my items sat behind the fleet's ~700-item backlog, hours away. That
> is wrong twice over: the 722 figure counts every status, and only **35**
> items across **2 sites** were actually dispatchable and older than mine. The
> selector takes one SITE per firing, so webdesign.uk came up within two
> minutes. **A backlog figure quoted from another lane's note is not the
> queue depth for your item** — the filter that matters is
> `status IN ('triaged','approved') AND attempt_count < max_attempts`.

- **`edit_live` did its job: the page GREW.** faq body 4,512 → 5,339 bytes
  (+18%), hero 3,063 → 2,881, cta 2,389 → 2,200. Compare `bugs_open/178`'s
  measured failure without it: 4,439 → 1,806, one paragraph in three surviving.
  The mode is the difference between a migration and a quiet amputation.
- **The rewritten page scans CLEAN** against the live register: 0 findings, 1
  suppressed — and the suppressed match is the writer's own sentence, *"Once
  you approve the site and pay, we do not offer refunds, so take the time you
  need before you approve."* **This is the prediction I marked `[LIMIT]` above
  coming out right on real output**: I could only show that my composed fixture
  passed, not that the writer would choose the mandated phrasing. It did,
  because the rule is in `writer_block` and in the ban's own `reason`, which is
  where the writer meets it.
- **Verified at the SERVED artefact, not the status**: `preview.webdesign.uk/faq.html`
  returns 200 with £149 ×4, "we do not offer refunds", "built by AI", "ZIP",
  and none of £75 / deposit / 14 days / two rounds / "we handle the setup".
  `last-modified 16:45:03Z`, `cf-cache-status: DYNAMIC` (HTML is not edge
  cached, so no purge is needed — unlike the `.js` bundle trap this lane hit on
  08-10). Work item → writer → save → render → deploy → served, all of it.

### The hole: `edit_live` PRESERVES claims the register no longer licenses, and nothing flags it

The rewritten FAQ says **"Anything that's our mistake, we fix at no cost."**
That is `changes_paid_defects_free` — the fact I deliberately did **not** carry
into the £149 register, because it was attested on 2026-07-29 under the £1,200
offer and is an open-ended liability at £149 (owner ask 2 above).

It survived because `edit_live` keeps what it is not told to change, and
**nothing mechanical objects**: it is not a banned pattern, it carries no
number, and the `governing_rule` that says every commercial term must trace to
a fact is prose the writer complies with, not a gate. So the asymmetry is:

> **Removing a fact from the register does not remove the claim from the page.**
> A ban catches a claim you named. A *missing* fact catches nothing at all —
> the register's silence is invisible to every check we have.

Not fixed here, deliberately, and it is not a regression: the sentence was
already live before this migration on `what-you-get` and `how-it-works`.
Stripping a customer-favourable promise without the owner's word is as much a
decision as keeping it. **It is now owner ask 2 with a live example attached**:
either re-attest it at £149 (and I put it back in the register) or say it goes
(and I write it into the guidance for the remaining pages). What must not
happen is it sitting on the page with nothing behind it.

The general form is worth carrying: **a claims register enforces its bans and
merely hopes for its facts.** Anything you want *gone* has to be named in
`banned_claims`; dropping the fact is necessary and never sufficient.

---

## 2026-08-12 (close) — all five pages MIGRATED, LIVE and CLEAN; one item says `failed` and is lying

> **⚠ CORRECTED 2026-08-12 (later) — "CLEAN" IN THIS HEADING IS FALSE, and the
> entry below is kept unedited as the original account.** The migration had also
> deleted **every call-to-action button on the four offer pages** — 14 anchors
> across 7 components, including both on the home page hero. All five checks
> recorded below passed and not one of them could see a missing `href`. Found
> ~20 minutes after this entry was written, while chasing an unrelated
> `page_divergence_overwritten` item; repaired and verified live at 17:37Z.
> The account, the mechanism and the check I should have run are in the entry
> at the foot of this file and in `WRONG_CALLS.md` (2026-08-12).

**The migration is done.** Final sweep over the five served pages, cache-busted,
17:16Z: **zero** occurrences of £1,200 / £75 / deposit / 14 days / two rounds /
"rounds of changes" / "we handle the setup" / "you only pay if you like" /
"before any money changes hands" — and **£149 present on all five** (index 3,
faq 4, how-it-works 1, what-you-get 3, guide 5). Every page scans **0 findings**
against the live register, each with exactly one match suppressed by the
negation guard, and in every case it is the writer's own "we do not offer
refunds" sentence.

| page | body bytes before → after | scan findings before → after | served |
|---|---|---|---|
| faq | 4,512 → 5,339 | 8 → 0 | 16:45:03Z |
| index | 9,617 → 9,713 | 1 → 0 | 16:55:29Z |
| how-it-works | 1,521 → 1,927 | 6 → 0 | 17:06:03Z |
| guide (article-body) | 5,512 → 5,847 | 4 → 0 | 17:11:12Z |
| what-you-get | 1,477 → 1,985 | 3 → 0 | 17:16:20Z |

**Not one page shrank.** `edit_live` held on all five, including the two whose
whole commercial paragraph had to be replaced.

### The guide page's link set was preserved, and the gate that says so was proven able to fail

`gate_page_links.py` (reused from the `loanandmortgagecalculator_couk` lane —
it reads `pages.content_direction->'required_links'`, the PAGE-level column, so
it does not collide with the fleet voice pass editing `site_specs`) reports
**all 4 required links present** after the rewrite. That pass means something
only because the same gate was run `--self-test` first and correctly FAILED on
an impossible link. The set was declared as data before the rewrite was queued,
and the insert transaction refuses to queue the guide's rewrite at all if
`required_links` is missing — an undeclared set makes the gate pass vacuously,
which is the failure mode that would look most like success.

### `what-you-get` is marked `failed` and the page is CORRECT AND LIVE

```
step deploy_page failed: workflow completed but its result could not be
delivered to the parent (failed_transient): message validation failed
(code: CHILD_ORCHESTRATION_FAILED)
```

This is the **spawn→call handshake race** (memory: `spawn-call-handshake-races`),
firing *after* the work was done. The evidence that the work landed, none of it
the work item's own status: components written 17:13:23Z and all three
`build_status='deployed'`; `pages.deployed_at` 17:13:58Z; **the deploy repo took
the commit** (`Rerender: what-you-get.html`, 17:13:55Z); the served page updated
at 17:16:20Z; and the content scans clean.

**Left as `failed`, deliberately.** It is a truthful record that the
orchestration's result delivery failed, and fabricating a `complete` would put a
lie in the one place a future reader is most likely to trust. **It will not
retry**: `find_dispatchable_site` selects only `triaged`/`approved`, and both
sweeps that could re-triage it (`vet-sweep-continue`, `improvement-sweep`) are
disabled — checked, not assumed. So the row is inert, not a pending rewrite of
an already-correct page.

> **The general shape, and it is the mirror of this estate's own rule.** "Trust
> the artefact, not the status" is usually a warning that `complete` overstates.
> Here `failed` UNDERSTATES: the page was written, deployed and served. A status
> is a claim about work by the thing that did it, and the handshake that carries
> that claim is a separate mechanism that can fail on its own. **Read the
> artefact in both directions.**

### Box sync lag, now with five samples rather than one

`build_status='deployed'` leads the served file by **58s, 74s, 89s, 142s, 285s**
(guide, faq, how-it-works, what-you-get, index). Same mechanism every time, so
that spread is a **timer period with random phase**, not a variable delay — the
box pulls on a schedule. I very nearly filed the 285s case as a broken deploy
after checking 46 seconds in. The check that separates the two without waiting:
ask the deploy repo whether it has the file
(`gh api repos/gqls/vm-sites/contents/<domain>/<page>.html`). If the repo has it,
the deploy worked and you are waiting on a timer.

### Sibling lane, closed out

Their `contact / chat-input-box` lock is untouched: exactly one row, still
`lock_type=permanent`, `updated_at` unchanged from 2026-08-11 15:03, and present
on the served contact page. Their NOTES carry the change notice (commit
`7adce5896`), including that the register their in-flight facts relay serves has
moved under them and that their bot's compiled-in `systemPromptFacts` are now
the stale half.


---

## 2026-08-12 (later) — the migration deleted every CTA button on the site, and five green checks could not see it

Found while reading a `page_divergence_overwritten` work item I had assumed was
routine bookkeeping. It said a page rebuild had overwritten a **hand-patched**
index hero — and the hand patch it destroyed was the sibling lane's own
restoration of those buttons from their 08-11 incident.

**The damage: 14 anchors across 7 components** — hero and call-to-action on
`index`, `faq`, `how-it-works`, `what-you-get` (index's cta block excepted, see
below). Every offer page lost its "Get in touch" / "Send us an email" buttons.

### Mechanism, read from the code and then confirmed empirically

The writer preserved the button **labels** (`cta_text`, `primary_cta`,
`secondary_cta`) and dropped the **destinations** (`cta_url`,
`primary_cta_url`, `secondary_cta_url`). Both templates gate the anchor on the
URL, not the label — `{{if and .cta_text .cta_url}}` — so the button renders as
**nothing at all**. No error, no missing text, no shrunken byte count.

Those URL keys are declared in `content_components.input_schema` with
`source: "renderer"`. In `plan_sections_action.go`, `sourceResolver.resolve`
short-circuits that source: `if source == "" || source == "llm" || source ==
"renderer" || source == "static" { return nil, true }` — value nil, **found
true**. The field is therefore never "missing", `handleMissingField` never runs,
and `carryStored` — the `bugs_open/238` carry (PBP-039) — never runs either,
because it guards fields that FAIL to resolve and a renderer-sourced field
always "succeeds" with nothing.

> **The carry IS live in the running binary**, checked at the artefact rather
> than inferred: agent-chassis `v1.0.1291` was built from `da5a7eb8f` (image
> label `org.opencontainers.image.revision`), and
> `git merge-base --is-ancestor d26c26a9a da5a7eb8f` passes, with controls in
> both directions (the stamp is its own ancestor; my commit from 20 minutes
> earlier is correctly NOT aboard). So this is a **gap in the carry's
> coverage**, not an unshipped fix — and `bugs_open/238`'s file, whose banner
> still says the fix is "inert until the next roll", is stale on that point.

**One hypothesis of mine was refuted on the way**, which is why the mechanism
above is filed rather than asserted: I first thought the URL keys were simply
**undeclared** in the schema, so the carry's field loop never saw them. The
schema query returned all four, declared, with `source: renderer`. Filed as
`090` run `97ef39f0-19df-4935-834d-c80514fbc43e` for independent diagnosis.

### Repair, in two parts, because the first was necessary and not sufficient

1. `SQL_2026-08-12d` restored the URLs into `content_data` (recovered from the
   pre-migration rendered HTML, not invented) and declared `required_links` on
   all five pages so the next rewrite is gated mechanically.
2. A `page_rerender` dispatched **after** that still rendered no buttons — the
   render path re-resolves renderer-sourced fields to nothing rather than
   reading the stored value. So `SQL_2026-08-12e` splices the anchors into
   `rendered_html`, built from `content_data`'s own fields so no label is
   retyped, refusing to write unless each insertion point matches exactly once.
3. The deployed files were patched in `gqls/vm-sites` by the sibling lane's own
   documented route for a verified deterministic edit (`b538295`), with the
   markup copied byte-for-byte from the repaired component rows.

**Live and verified 17:37Z**: index 2 buttons, faq 4, how-it-works 4,
what-you-get 4 — and retired terms still 0 on all four.

**Known cost, taken deliberately:** hand-patched `rendered_html` re-arms
`bugs_open/229` — the next rebuild will overwrite these bytes and file a
divergence item, exactly as it did to the sibling lane's patch this afternoon.
The alternative was leaving a sales site with no call-to-action buttons. The
durable fix is the platform change the `090` is for.

**`index/call-to-action` deliberately NOT repaired.** It carried no URLs before
the migration either. Repairing it here would improve the page and, in doing so,
erase the boundary of what my migration actually broke.

### Why five green checks missed it — the part worth carrying

`claimscan` asks whether a banned **phrase** is present. Byte deltas ask whether
the text was **truncated**. The retired-term grep asks about the terms **I
named**. The served fetch asks whether the new **copy** arrived. And
`gate_page_links.py` asks about the **one page I declared a link set for** —
the guide, because that was the page I judged at risk. It passed 4/4, and the
damage was on the four pages I had not declared a set for.

**A gate covers what you point it at.** Its green light was load-bearing in my
confidence, and it only ever meant "links are fine on one of five pages".

**The check that would have caught it needed no new tool and used data I had
already exported** — count `href="` per component as a matched before/after
pair. I ran precisely that query twenty minutes after calling the work
complete. Run against the first canary it would have failed loudly, before four
more pages were queued.

**The general form: I verified what the change was FOR, never what it might
COST.** A verification suite assembled from your own intent is structurally
blind to collateral damage, because every check was chosen by the same author
with the same expectation. For any regeneration, diff the invariants — link
count, image count, component count — and treat whatever you did not intend to
change as the finding. Full entry in `WRONG_CALLS.md`.


---

## 2026-08-12 (close) — the 090 REFUTED my mechanism, and my run was authored so that it almost had to

`090` run `97ef39f0-19df-4935-834d-c80514fbc43e` returned **REFUTED** on the
claim that renderer-sourced fields fall outside `carryStored` and that is why
the CTA keys were lost. **Recording it as a visible correction, per the standing
rule — a refuted verdict is the cheapest place to be wrong, and this one cost
one run.**

### What still stands, and what does not

- **The DAMAGE is not in question.** I measured it first-hand before any repair:
  the live `page_components` rows for index/faq/how-it-works/what-you-get heroes
  and CTA blocks had `cta_url` / `primary_cta_url` / `secondary_cta_url` **absent**
  (`content_data ? 'cta_url'` → false), and all 7 components rendered **0**
  anchors. `page_component_history` holds the outgoing rows that still carried
  the keys. That is unchanged by this verdict.
- **The MECHANISM is NOT established.** My reading of
  `sourceResolver.resolve` returning `(nil, true)` for `source: "renderer"`,
  and of `carryStored` therefore never running, is a plausible reading of the
  code that the loop declined to accept as the cause here. Treat it as an open
  question, not a finding.

### The authoring error, which is the part worth carrying

**I repaired the state and then asked the loop to find the damage in it.** The
verdict's own citations are `page_components.content_data` rows carrying
`"cta_url": "/contact.html"` — the values *I put back* at 17:23, four hours
into the incident and sixteen minutes before the run started. Of course the key
was present; I had restored it. The loop measured a repaired system and
correctly reported that nothing was missing.

My symptom text asked it to "determine whether a renderer-sourced field can be
lost from `page_components.content_data`" — a question about the LIVE table,
which by then no longer held the evidence. The evidence had moved to
`page_component_history`, and I never pointed the run at it.

> **The general form, and it is a specific one this estate already knows in
> another costume: your own remedial action can silence the very check you then
> commission.** The existing lessons are about a detector you disarm by
> declaring a key, or by filing a row that bumps `updated_at`. This is the same
> shape at a coarser grain: **fix first, ask second, and the answer is about the
> fix.** Where a repair is urgent (it was — a live sales site had no buttons),
> the run must be pointed at the *archive* of the broken state, and the symptom
> must say plainly that the live table has since been repaired and when.

### What the run DID give, and it is worth more than my hypothesis was

The revised hypothesis corrects something I had wrong about the parent bug:
**`bugs_open/238` as tracked in the codebase is the dead-URL-CONTROL defect** —
a section that renders while leaving a URL attribute empty, recorded
non-fatally on the rerender path — **not** a case of a renderer-sourced field
being dropped from `content_data`. I had been reading 238 as the latter and
filing my case under it on that basis.

`next_scope` names where to look next: `dead_url_guard.go`,
`emitSectionDeadControlItem`, `recordDeadURLControls`. Its cited comment is the
tell I should have read first — *"DeadURLSlots names sections that RENDERED,
but with a URL attribute left empty (bugs_open/238)"* — which is exactly the
symptom I had, described in the codebase, one layer away from where I was
looking.

**Owed, and NOT done in this session:** a re-run authored against
`page_component_history` for the 16:37–17:23 window, with the symptom stating
that the live rows were repaired at 17:23. Until then the cause of the CTA loss
is **open**, and the corrections above have been made everywhere I asserted it
(`HANDOFF` §3b, `LANDMINES.md`, `bugs_open/238` §9).


---

## 2026-08-12 (evening) — the offer hardens to take-it-or-leave-it, and the brand gets a visual system

Owner ruling recorded in full as PLAN §1d. **It retires §1b ruling 2 — "one set
of changes included" — which I implemented in the register and shipped to all
five pages this afternoon.** That line is now live and false on every page. It
is the same class of problem I spent today fixing, arriving from the other
direction: not a stale claim I inherited, but a true one that a ruling made
false four hours after it shipped.

### What changed

No corrections at £149. Two private remedies (re-run, refund), **neither
advertised** — the refund by the owner's explicit instruction, the re-run by the
same logic, since printing it reopens the negotiation the offer exists to close
and lets the customer define "bad". The copy says no changes and no refund, hard
and early, because that IS the cost control rather than a tone choice.

### The research, and the one point that changed the owner's plan

- **The no-frills mechanic works because promise and delivery match**, not
  because of the rudeness. Ryanair's marketing set correct expectations rather
  than flattering ones, so customers were not disappointed
  ([case study](https://themarketingagency.ca/blog/ryanair-case-study/),
  [Marketing Week](https://www.marketingweek.com/ryanair-cmo-brand-love/)).
- **Surprises are the number one trigger of refund demands**, and stating terms
  up front is the biggest single lever on them
  ([Sentisum](https://www.sentisum.com/customer-complaints),
  [CPD Online](https://cpdonline.co.uk/knowledge-base/business/customer-complaints-expectations-psychology/)).
  So the hard copy is the mechanism, not the decoration.
- **Nobody in cheap web design positions on "no revisions."** The category
  advertises unlimited and then quietly caps it at the design stage or a fixed
  number of rounds ([Feedbird](https://feedbird.com/blog/best-affordable-web-design-services/),
  [ManyPixels](https://www.manypixels.co/blog/web-design/website-design-packages)).
  Genuinely differentiating, and no proved copy to borrow.
- **The anti-design risk is real and specific to us.** Ugly reads as broken (the
  Aesthetic-Usability Effect), brands inherit the aesthetic's adjectives, and
  the failure mode is rawness leaking out of the visual layer into interaction
  ([Divimode](https://divimode.com/brutalism-web-design/),
  [Toptal](https://www.toptal.com/designers/ux/minimalist-brutalist-web-design)).
  **Our site is the demo**: a scrappy-LOOKING site undersells a product that is
  actually clean and professional. Owner agreed and stated the rule himself —
  utmost quality in the product, as if delivering for a corporate.

### The visual system, and the constraint that makes it safe

One idea, restated by every motif: **an unorthodox, scrappy PROCESS producing a
genuinely good RESULT.** Heath Robinson contraptions and marble runs; a scruffy
goose with one golden egg; plain cardboard boxes and flat-pack; cash-and-carry
pallets; a trade counter. The owner's two additions are stronger than my
cardboard-only pick, because the contraption makes the *process* the joke rather
than the product — which is exactly what "built by AI" needs.

**The governing rule is written into `avoid`, not into a comment: the cheapness
is in the SERVICE, never in the PRODUCT.** `avoid` is the only field that
reaches the negative prompt (`generate_image` composes medium/mood/palette as
direction and sends `avoid` negatively), so a rule that lives anywhere else in
the spec is decoration. It bans any suggestion the output is poor — no rust,
refuse, decay, dereliction or broken objects — and that is also why the junkyard
is NOT the base register: a junkyard says the goods are scrap.

Medium is pen-and-ink line drawing with one flat wash — Heath Robinson's own
hand. It carries a contraption, a goose and a box in one style; it reads as
deliberate craft rather than a cheap photograph; and line art survives the
card-size gate that photographic output kept failing (D14). Palette: kraft
ground, near-black ink, one warm gold used only on the output element. Text
stays banned — generated lettering garbles, and the shipping-label typography
this brand wants belongs in HTML type over the image.

Live as `f20ceedb9` (`SQL_2026-08-12f`), superseding the 08-04 studio guide.
**Images are for the framework's imagery pipeline to generate from these
guidelines — the owner explicitly does not want them made by hand at the CLI.**

> **Misstep, recorded because the missteps are the point: backticks in a
> `git commit -m` message EXECUTED.** The word `avoid`, wrapped in backticks in
> three places, was command-substituted away, so `f20ceedb9`'s message reads
> "written into , not into a comment, because / is the only field". The content
> is unaffected and the SQL file carries the full explanation. **This is a
> landmine already in my own memory index** ("backticks in `-m` execute") and I
> hit it anyway — knowing a trap is not the same as having a habit that avoids
> it. The habit: single-quote the message body, or write it to a file and use
> `-F`. Forward-only forbids an amend, so the corrected sentence is here.

### Owed next, and a sequencing hazard worth stating

1. **The copy rewrite**: remove "one set of changes is included" from the
   register and all five pages, and state the take-it-or-leave-it terms hard.
   One pass, not two — every rewrite is another chance to lose the CTA buttons.
2. **`design_intent` is now at odds with the imagery.** It still describes "a
   well-printed document": near-white ground, forest-green accent, and
   explicitly "No hero image, no full-viewport splash". A kraft-ground
   contraption illustration on that page will look borrowed. The palette and
   layout preference need to move with it.
3. **⚠ A rebuild will delete the CTA buttons again.** Whatever triggers the copy
   rewrite or a design change will rebuild these pages, and the button loss is
   unfixed at platform level (§3b of the handoff). Re-apply `SQL_2026-08-12e`
   and re-push the deployed files afterwards, and check the href counts as a
   matched pair before and after.


---

## 2026-08-13 (morning) — the fleet-fix handoff exists, and round one of the gap-fills failed in the known way

- **The fleet fix is now another thread's to start**:
  `bugfix_268_cta_buttons_fleet/HANDOFF_2026-08-13_start_here.md` (commit
  `a99e95efd`), registered in MEMORY_workstreams. It front-loads the two
  refuted hypotheses, specifies the 090 re-run against `page_component_history`
  (windows 16:37–17:23Z and 20:20–20:45Z, repair times stated), ranks the fix
  candidates, separates FIX from REPAIR (fix first — repairing first re-armed
  the loss once already), and says plainly that webdesign.uk is locked and
  handled: leave it out of sweeps, unlock it last.
- **Overnight state, verified at the served artefact**: buttons held (index 2,
  faq 4, how-it-works 4, what-you-get 4), no refund promise anywhere, the
  no-changes term stated on faq/how-it-works/what-you-get.
- **Round one of the two gap-fills failed in the documented way.** Both items
  wrote (`updated_at` 21:14/21:17) and neither did its additive job: the faq
  answer named zero of the six services, index's summary still omits the
  no-changes term. A prose instruction to INCLUDE a set inside a larger brief
  is not followed — same class as the guide-links loss, from the insertion
  side. `gapfill_index` also drew the handshake-race `failed` (work landed,
  delivery lost), second sighting.
- Round two dispatched (`gapfill2_*`), one job per item, each phrased as THE
  TEST THIS PAGE MUST PASS. **Capped at two rounds** per the council lesson —
  if round two misses, it goes to the owner as a parked item, not a loop.


---

## 2026-08-13 — why there are no Heath Robinson images, and why the writer refused the service names twice

Owner asked why the imagery discussion produced nothing on the site. Traced,
and both this and the gap-fill failures come down to the same shape: **a spec
that is live is not a spec that is driven.**

### The imagery: the guide is a rulebook, and nothing has opened it

1. **The style guide is consumed at GENERATION time and generates nothing by
   itself.** `generate_image` composes per-kind direction from it when an image
   is being made. No image has been made for this site since 2026-08-10 (last
   `needs_imagery` completed). The guide superseded at 19:35Z on the 12th has
   never been read by a live generation.
2. **No detector fires on a style change.** `check_content_image_missing` and
   `check_unfulfilled_imagery_plan` key on ABSENCE — a planned image with no
   asset. Every planned image exists (heroes + logo, built 08-08/08-10 under
   the OLD brand), so nothing flags, for ever. The memory lesson verbatim: a
   silent mechanism is usually undriven, not missing.
3. **The guide is only half the input anyway.** The SUBJECT of each image comes
   from `site_plan_imagery.prompt`, and those still describe the old studio
   brand ("minimal, light abstract composition", "a clean modern workspace").
   I changed the style half and left the subject half stale — regeneration
   without new prompts would have drawn old subjects in new ink.
4. **The locked heroes do NOT block new imagery.** A regenerated asset ships
   under the same `asset_key` and therefore the same served path; the locked
   component's `hero_url` reference never changes, only the file does.

**Canary dispatched through the framework** (owner's explicit channel — no CLI
generation): `imagery-brand-canary:page:index:hero_home`, `image-build-handler`,
prompt = a Heath Robinson contraption delivering one immaculate finished page,
brand_update=true per `imageryplan.BrandUpdate`. The current plan's `hero_home`
prompt updated in step so the plan and the item cannot drift. Junkyard is NOT
in the motif set — the owner's final list (contraption/marble run, goose+egg,
boxes, pallets, trade counter) does not include it, and the guide's `avoid`
bans dereliction.

### The gap-fills: the writer was OBEYING, and the defect was mine — the 2026-08-09 lesson recurring

Rounds 1 and 2 of "name the six services" both produced good prose with zero
names. Cause, found by asking what the writer actually receives:
`SQL_2026-08-12i` put the names into `facts[]` and `allowed_entities`, and
writer_block referred to them as "in fact `third_party_options`" — **a data
structure the writer never sees.** Its own whitelist paragraph then says: not
in this list → "rewrite around it or leave it out". Asked twice to name
services outside its whitelist, it refused twice, correctly. **facts[] is
bookkeeping; writer_block is the wire** — this lane wrote that lesson on
2026-08-09 and I re-enacted it with a pointer instead of a copy.

Fixed at the wire (`SQL_2026-08-13a`: the six names now in writer_block text);
round 3 dispatched. The two-round cap was not breached in spirit: rounds 1–2
failed on MY spec defect, not on writer noncompliance, and round 3 runs with
the cause removed.

**index, reassessed:** the term IS now stated — the stat tile reads
"0 Changes included" (deadpan, exactly the register) and step 3 says "you
approve the site, pay £149, then get the finished files to keep for good".
My round-2 grep looked for prose spellings and missed the tile. Accepting
this; not dispatching a round 3 for index.

**Still owed on imagery after the canary is judged:** the remaining heroes
(faq, how-it-works, what-you-get, contact) + the section icons under the new
brand; the `design_intent` palette/layout pass (still describes the
"well-printed document" brand and says "no hero image"); owner look at the
canary before the fleet of five.

## 2026-08-13 — CROSS-LANE NOTE from the webdesign_uk_build_service lane (not this thread; your words above are untouched)

Two things you flagged as falsifiers/blockers are now resolved or connectable.

**1. The site-facts relay SHIPPED and is LIVE (your §5 falsifier).** CHAT-010
is on core-manager `v1.0.1294`. The webdesign.uk chat bot now reads its facts
from THIS register's `evidence_base` over the WireGuard tunnel, refreshed every
5 min. Proven at the artefact 2026-08-13: the bot answers **"£149, one-off, you
approve before you pay, no refund"** with zero retired £1,200/£75/14-day terms.
**Before this, the bot was still quoting £1,200 while your five pages said £149**
— a live contradiction a visitor would have hit. It's gone. Net effect for you:
**the register's `facts` are now the single source of truth the bot speaks
from — change a fact, the bot changes within one refresh, no redeploy.** (The
`writer_block` is deliberately NOT served to the bot; only `facts[].claim`.)

**2. Your webhook-exposure decision #2 recommends "proxy from the webdesign.uk
box over the existing tunnel — needs the sibling lane." That tunnel now
EXISTS.** WireGuard peer `webdesignbox` is up and proven (box → in-cluster
ClusterIP, real forwarded traffic, `ip_forward` latent fault found and fixed —
see the CHAT-010 register entry + the LANDMINE). Today it carries box→cluster
*outbound* (the facts fetch). For a Stripe webhook you need the reverse
direction — the box receives the public HTTPS POST and proxies it *inbound* to
auth-service over the same tunnel. The tunnel is bidirectional once up; the box
would need a small nginx `location /stripe/webhook { proxy_pass
http://<auth-service ClusterIP>:<port>; }`. That's box config, small, and it
avoids standing up an Ingress or a second Cloudflare tunnel. Whoever picks up
your decision #2 should talk to this lane's RUNBOOK for the tunnel details.

**3. The owner's live question (2026-08-13), recorded here because it lands on
your `payment_timing` switch:** *"What do we need to make it create the site
after taking money. Perhaps we should take money first, I don't know yet."*
My read, for your consideration, not a ruling:
- Your current design is `after_approval` (build → preview → approve → pay),
  and the copy now states it. That has **zero refund risk and zero
  chargeback-for-copying-then-walking risk in the sense that they haven't paid**
  — but it means every build is speculative (built before any money), and for
  an AUTOMATED flow that is exactly the "anonymous internet-triggered,
  token-spending build" SAAS-001 warns about. Automation removes the human gate
  that makes a speculative build safe today.
- **Pay-first fits the £149 "no refund" terms MORE cleanly than pay-after
  does** — "pay £149, we build it, you get the files, no refund" is a product
  purchase; "pay after approval, no refund" is slightly odd (why name a refund
  you'd never need?). Pay-first also makes payment its own rate-limiter and
  abuse guard: no paid order, no build.
- The catch you already documented (decision #4): the five pages now SAY
  pay-after-approval, so flipping `payment_timing` to `upfront` is a copy
  migration, not a one-field UPDATE.
- **So the owner's instinct ("take money first") is, I think, the right one
  for the automated path specifically** — but it's your switch and your copy,
  so I'm recording it, not flipping it.


---

## 2026-08-13 (midday) — canary VERIFIED BY EYE; round 3 of the FAQ services failed and is PARKED; fresh build v1.0.1295

- **The imagery canary is live and on-brief, verified at the artefact by
  looking at it**: pen-and-ink Heath Robinson contraption on kraft ground,
  machine drawn as if it works, gold used ONLY on the immaculate web page
  rolling out of it. Deployed to `hero-home.jpg` IN PLACE (same asset_key ⇒
  same path), served `last-modified 2026-08-13 10:45:02Z` — so the locked hero
  component was never touched, exactly as designed. The pipeline even filed and
  completed its own "Re-render index after its image asset landed" item.
- **gapfill3_faq FAILED with the names in the wire** — the answer is unchanged
  , no services named, three rounds total. Rounds 1–2 were my
  facts-vs-wire defect; round 3's cause is UNKNOWN and I am not running a
  round 4 blind. PARKED with three options in the handoff (§2.2): section-data
  insertion, read the rendered prompt from `llm_call_log` (the 08-09 method),
  or the owner drops the ask.
- **Fresh chassis build `v1.0.1295`** (image label ← `69612d692`, pods started
  13:53Z). My 268 filing is aboard; the 268 fix paths (`plan_sections`,
  `save_page_sections`, `link_repair`) show no commits today, so the fleet
  defect is unchanged and the locks stay on.
- SUMMARY_2026-08-13 written (offer hardened + brand landed + fleet defect
  found); HANDOFF_2026-08-13_continue_here supersedes 08-12's.


---

## 2026-08-13 (afternoon) — the brand is LIVE on every hero; the FAQ services item fails a FOURTH round; the cluster token expired mid-session

- **All five heroes generated and verified BY EYE at the served artefact**:
  contraption (home, the canary), marble run (how-it-works), open cardboard box
  with the framed page (what-you-get), trade counter with the gold page on it
  (faq), scruffy goose beside the golden egg (contact). One hand, one kraft
  ground, gold only on the output — a coherent set.
- **En route: four heroes painted `/assets/images/hero.jpg`, which 404s**, while
  their per-page files had existed since 08-08. References fixed in
  content_data AND rendered_html through my own locks (the verify assertion
  caught the guide as a fourth case my page-list missed). what-you-get and faq
  converged via pipeline redeploys; **how-it-works and the guide were pushed
  directly to vm-sites (`c075a6c`)** because the cluster token expired before a
  rerender could be dispatched — the DB already carried the fix, so the push
  CONVERGES file and DB rather than diverging them.
- **`gapfill4_faq` FAILED — four rounds now.** Round 4 ran with the six names
  in writer_block AND in the whitelist ENUMERATION (SQL_2026-08-13b, committed
  ~16:46Z, item ran 16:53–16:59Z). The round-3 diagnosis (enumeration excluded
  them) was read from the rendered prompt and was true, but fixing it did not
  suffice. **PARKED HARD. Do not run a round 5 on guidance changes.** Next
  moves when the cluster returns: read round 4's rendered prompt
  (`llm_call_log`, ~16:53–16:59Z, page-content-writer iter for the faq
  section) to confirm the enumeration edit reached the call; then either
  section-data insertion or an owner drop. Buttons and hero background on faq
  are intact (verified at the served page — note `grep -c` counts LINES not
  matches; the four anchors sit on two lines).
- **The kubeconfig token expired ~17:00Z** (fleet-wide Unauthorized; the known
  3-day expiry, owner refreshes). Everything after that point was done at the
  served artefact and the deploy repo; DB-side verification of the last hour
  is OWED once access returns.

## 2026-08-14 — from the webdesign_uk_build_service lane: your Stripe keys have a trap waiting — `personae-platform-secrets` is terraform-owned, and a kubectl-added key dies at the next release

Your handoff's owner decision 1 says: add `STRIPE_SECRET_KEY` +
`STRIPE_WEBHOOK_SECRET` "to `personae-platform-secrets`". **Do NOT do that with
`kubectl patch`/`edit`.** That secret's whole `data` map is reconciled by every
`make release` (`deploy-core` → `deploy-047-base-configs` → `terraform apply`),
and any key terraform does not declare is DELETED. Proven live 2026-08-13: my
`SITE_FACTS_TOKEN` was kubectl-added in the morning, wiped by the 13:53Z
release, and the chat bot's facts relay 401'd for the rest of the day. For
Stripe the same failure is worse: payment breaks on the FIRST release after
the keys go in — silently, on a seam where "money was taken" and "webhook
processed" must not diverge.

The right way (worked example: `site_facts_token`, commit `ebf839ce2`):
declare `stripe_secret_key` + `stripe_webhook_secret` variables in
`deployments/terraform/environments/production/uk001/047-base-configs/variables.tf`,
map them into the secret's `data` in `main.tf`, put the values in the local
gitignored `terraform.tfvars.secret`, then `terraform apply` (or let the next
release carry them). Env is read at pod start, so auth-service needs a restart
after the apply. Full trap write-up: LANDMINES § "An additive `kubectl patch`
to `personae-platform-secrets` … is deleted by the NEXT `make release`".


---

## 2026-08-14 (morning) — round FIVE lands the services; the dead channel is proven by discrimination; bug 271 filed

- **The FAQ names all six services, live and well-phrased** ("A few names
  worth knowing, if you want to look further…"), buttons 4, retired terms 0.
  What worked: the instruction as an IMPERATIVE in writer_block
  (`SQL_2026-08-13c`) — standing content, so it survives future rewrites with
  no per-item brief.
- **Root cause of all five rounds, proven twice over**: `content_guidance` has
  four writers and NO reader (grep at HEAD; independently observed in
  `bugs_closed/177`'s close-out footnote on 08-02, where it sank), and the
  guidance strings are absent from all six probed writer prompts while
  writer_block appears verbatim. Round 5 was the discriminating experiment:
  guidance inert, wire carries the ask → names appear. **Filed as
  `bugs_open/271`** with the blast radius (every content_rewrite brief
  fleet-wide, including the platform's own gap-planner) and two fix candidates.
- **My rounds 1–4 post-mortem, honestly**: three register defects were real
  and worth fixing (pointer-not-copy; enumeration exclusion) but INCIDENTAL —
  the channel was never connected. The tell I missed for a day: every ask that
  "worked" via guidance was also in writer_block.
- `gapfill5_faq` reads `failed` with the page deployed and correct — handshake
  race, fourth sighting.
- **Remaining on this front: the palette pass** (design_intent still the old
  white brand under kraft heroes) — values proposed to the owner in chat, not
  applied.


---

## 2026-08-14 (evening) — the kraft palette is LIVE, verified served, and the anchor held exactly

- **Served stylesheet verified**: background `#f7f1e6`, surface `#efe6d3`,
  border `#d8c3a0`, text `#1a1a1a`, muted `#6f6553`, accent `#8a6410` — all
  seven approved values BYTE-EXACT, zero green or old-white residuals. Given
  `reference_values` are advisory by construction (the corrected colour-churn
  landmine), exactness was the thing to verify, not assume: baseline copy taken
  before dispatch, matched-pair diff after.
- **Contrast, computed not eyeballed**: text 15.5:1, accent-as-link 4.8:1,
  muted 5.1:1, button text 17.4:1 — all past AA's 4.5.
- **Buttons are INK, not the proposed egg-gold**: `.btn-primary` is built
  structurally from a primary→secondary gradient (`#1a1a1a`→`#4a4437`) with
  white text — the renderer's CTA slot is not palette-addressable, so the gold
  fill would need a template change. Reported to the owner with the honest
  case for keeping ink (gold stays reserved for the artwork's output element);
  awaiting his preference, not chasing it unprompted.
- **Route notes for the next palette change on any site**: write all FOUR
  copies in one transaction (design_intent + content_data.color_scheme +
  palettes + style_collections — worked SQL `SQL_2026-08-14_kraft_palette.sql`);
  do NOT stack duplicate design runs — a discovery-filed `needs_design_review`
  at priority 10 will front-run yours, so cancel the duplicate and ride it
  (fewer LLM passes = less drift surface); the acceptance is the SERVED
  stylesheet against a baseline, never the item status. My transaction's
  all-copies-agree union check was WEAKER than intended (passes if ANY copy
  matches) — verified by direct read instead; do not reuse that block.
- Sequence that actually delivered it: discovery's own prio-10 design item ran
  the agent (mine cancelled as duplicate) → repo commit 18:38:41Z → box sync →
  served. Monitor caught each stage; pages re-verified green throughout (link
  gate 5/5, buttons intact).

**The brand work is COMPLETE**: imagery guide → five kraft heroes (verified by
eye) → hard-terms copy → services named → kraft palette served. Remaining on
the site front are the parked/owner items only (271 fix thread for the
guidance field; Stripe keys; webhook route; "three or four days").

## 2026-08-14 (late) — from the webdesign_uk_build_service lane: your fact edits now reach the chat bot live AND durably; the chat box became a library tool (gated); two filings you may care about

- **When you change `evidence_base` (e.g. if the owner re-attests "three or
  four days"), the chat bot follows within 5 minutes** — no redeploy, and as
  of this morning that pipe is release-proof: yesterday's 13:53Z fleet
  release DELETED the relay's token (`personae-platform-secrets` is
  terraform-owned; the kubectl-added key died at the next release — my
  Stripe warning above is this incident's lesson) and the bot ran on
  last-good facts until the owner's one-command fix at 08:12Z. The token now
  lives IN terraform (047-base-configs), so releases re-assert it. Practical
  for you: **include the bot's replies in your retired-term/claims greps** —
  it speaks the same register your pages do, from the same rows.
- **`chat-input-box` is now `component_level='tool'`, `category='interactive'`**
  (PLAN step 1, 2026-08-14), and tool-suggester gained the requires-backend
  gate (migration 406, council APPROVED, corr c78ed496): a backend-requiring
  tool is only offered to sites with `deploy_config.capabilities:['backend']`.
  **Your locked contact-page instance is untouched** — the page-instance
  loader resolves by name/function regardless of level (verified served after
  the flip), and the lock still holds (an improvement sweep tried to remove
  it 08-11; the block is in `needs_human_review`, a4cd5dc8).
- **Two filings from today's work**: `bugs_open/275` (tool-suggester's
  LIMIT 30 hides 38 of 68 library tools — affects any tool suggestions you
  rely on) and `bugs_open/276` (section-level `requires-backend` components,
  e.g. intent-probe, are still placeable on backend-less sites through
  generic planning — the gate only covers the tool path; same broken-widget
  class as your 228 references).
- Standing owner instruction today: both webdesign threads keep each other
  aware — I read this file's tail at my session starts and natural breaks;
  anything you need from the box/tunnel/relay side, write it here and I will
  pick it up.


---

## 2026-08-14 (night) — Phase 1 of the delivery/editor plan DONE: "usually next day" live on all five pages

- **Served, verified**: zero occurrences of the retired three-or-four-days
  figure on any page; "ready the next day" present on all five (index 3,
  how-it-works 2, what-you-get 2, faq 1, guide 1).
- The three locked heroes moved surgically (href-asserted); the six unlocked
  mentions moved by the framework with the instruction in writer_block.
- **The invariant monitor caught the run's one casualty in real time**: site
  links 30 → 29, and the gate named it — the guide's `/what-you-get.html`
  AGAIN, third time. This round the writer kept the sentence and stripped
  the ANCHOR (label-without-destination, in prose). Restored in both columns,
  gate 5/5, redeploy queued. **Pattern now firm: any rewrite of the guide's
  article-body is likely to cost that link; the gate + this restore recipe
  are the standing counter until 268's class fix covers prose links.**
- `nextday_how-it-works` reads `failed` with its content landed and correct —
  handshake race, FIFTH sighting.
- Register trail: `SQL_2026-08-14b` (fact + wire + ban, ban proven on 13 live
  mentions before writing). Phases 2–6 live in the
  `site_delivery_and_editor` workstream (plan approved verbatim, cold-start
  ready); this lane's part of Phase 1 is closed.

## 2026-08-15 — cross-lane note (bugfix_268_cta_buttons_fleet): your 8 CTA locks are OFF, by owner directive

The 8 permanent hero/call-to-action locks from `SQL_2026-08-12k` were lifted
today (`SQL_2026-08-15_unlock_cta_components.sql`, this directory —
UPDATE 8, verified: chat-input-box lock untouched, all 9 destination rows
keep their urls). Safe because bugs_closed/268's carry fix (`8f899cc8d`,
live v1.0.1298+) now preserves renderer-sourced CTA keys through rewrites —
proven by canary and a permanence rewrite. Your locked_by text prescribed
unlock→edit→relock; the RELOCK step is retired with the locks. A
`lock_blocked_change` item will no longer be filed for these rows;
`content_rewrite` on the offer pages works normally again.
