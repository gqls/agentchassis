
<!-- SOURCE: U04_idea_uk.md -->
### Stripe integration pattern: webhook as the only source of truth
- **category:** payments
- **status-signal:** deployed
- **status-evidence:** Live £29 payments proven end-to-end 2026-06-14 (incl. resolving the stray-character webhook-secret incident); full setup documented from the real dashboards.
- **what:** The reference payments pattern proven by idea.uk: entitlement/fulfilment granted only on a signature-verified `checkout.session.completed` (browser redirects prove nothing); webhook handling idempotent via an event-dedup table; a **restricted** API key scoped to Checkout Sessions:Write only; test and live are separate accounts with separate webhook destinations and secrets ("a sandbox webhook does not cover live"); the signing secret must be byte-exact (one pasted stray character 400'd every event and stalled a paid order — recovered by resending the event); Stripe keeps its fee on refunds; no SDK — raw HTTP + HMAC verify.
- **sources:** idea.uk/RUNBOOK_idea_uk(9).md (Stripe billing — setup + troubleshooting); idea.uk/PLAN_stripe_billing_integration(3).md (idea.uk reference block); idea.uk/golang_files/billing.go (header)
- **relations:** platform billing plan (adopts these principles); request-then-confirm flow.
- **verify-later:** billing.go webhook verify (HMAC-SHA256 over timestamp+body, constant-time compare).

<!-- SOURCE: U04_idea_uk.md -->
### Platform Stripe billing integration plan (auth-service truth + chassis entitlement cache)
- **category:** payments
- **status-signal:** aspirational
- **status-evidence:** "the auth service has a subscription scaffold… but no working payment integration — no Stripe SDK, no checkout creation, no webhooks"; every DDL marked PROPOSED.
- **what:** The chassis-wide billing design for the build/host/chat product: truth lives in the auth DB mutated only by verified webhooks; the chassis gates on a one-way-fed `client_entitlements` cache (Kafka entitlement-changed events + reconciliation sweep) because the maintenance heartbeat can't call auth per site; two charge shapes — recurring tier subscription per client and a one-off **$5 build credit** (Checkout mode=payment, consumed via the atomic-claim idiom); build-submission gate reuses the `approval_mode` hold; provider interface from day one. idea.uk is the cited working reference for the one-off path.
- **sources:** idea.uk/PLAN_stripe_billing_integration(3).md; idea.uk/RUNBOOK_idea_uk(9).md (reference implementation)
- **relations:** Stripe webhook pattern; admin-dashboard-and-api (auth service); scheduler heartbeats.
- **verify-later:** auth service repo subscriptions tables; any client_entitlements table (expect absent).

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Entitlement gate architecture (build-submission + maintenance-run gates)
- **category:** payments
- **status-signal:** aspirational
- **status-evidence:** PLAN doc §8 describes both gates in the future tense as design; build order (§10) lists them as unbuilt steps
- **what:** Two entitlement checkpoints reusing existing chassis mechanisms: (1) build-submission gate — a new `pending_entitlement` hold state on `site_work_items.approval_mode` (mirroring the existing hitl/pending_review pattern), parking the first expensive work item until a billing check clears, with atomic credit consumption via the same UPDATE...RETURNING idiom as `claim_work_item`; (2) maintenance-run gate — a join-filter added to the three heartbeat selection queries requiring `maintenance_active`, valuable even before any domain is sold.
- **sources:** stripe/PLAN_stripe_billing_integration.md#§8, stripe/001commentary.md#final turns
- **relations:** Two-plane billing architecture; Ownership hierarchy reuse for entitlement scoping; One-off credit vs recurring subscription billing model
- **verify-later:** site_work_items.approval_mode values; build-pipeline-trigger/improvement-loop/content-feed-trigger selection SQL

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Two-plane billing architecture (auth-service truth + chassis entitlement cache)
- **category:** payments
- **status-signal:** aspirational
- **status-evidence:** PLAN §3: "Truth = auth DB, mutated only by webhooks... Gate reads = chassis client_entitlements cache, fed one-directionally from auth" — all proposed, table marked PROPOSED
- **what:** Splits billing across two databases/services with one directional bridge: the auth service owns billing truth (subscriptions, credits, webhook-driven events); the chassis reads a local `client_entitlements` cache table fed by an entitlement-changed Kafka event plus a reconciliation sweep backstop — required because the maintenance heartbeat must join across thousands of sites per tick.
- **sources:** stripe/PLAN_stripe_billing_integration.md#§3,§5
- **relations:** Entitlement gate architecture; Isolated chat/satellite architecture (Y-copy); Pluggable billing provider abstraction
- **verify-later:** proposed table client_entitlements; entitlement-changed Kafka event/consumer (not built)

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Pluggable billing provider abstraction (Stripe as implementation #1)
- **category:** payments
- **status-signal:** aspirational
- **status-evidence:** PLAN §4 gives a full Go interface sketch explicitly labelled "Sketch, not final"; current code has "no Stripe SDK" imported at all
- **what:** A `Provider` interface (`EnsureCustomer`, `CreateSubscriptionCheckout`, `CreateOneOffCheckout`, `CreatePortalSession`, `CancelSubscription`, `ParseWebhook`) behind which Stripe is the first implementation, normalising provider-specific webhook payloads into a provider-agnostic `Event` type. Justified as "zero retrofit cost" specifically because no Stripe integration exists yet.
- **sources:** stripe/PLAN_stripe_billing_integration.md#§4,#TL;DR
- **relations:** Two-plane billing architecture; Existing but non-functional auth-service subscription scaffold
- **verify-later:** internal/auth-service/subscription/{models,repository,service,handlers}.go

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Webhook-as-only-source-of-truth billing principle
- **category:** payments
- **status-signal:** partial
- **status-evidence:** "service.go imports only uuid, zap, time, context, fmt — no Stripe SDK. CreateSubscription is a bare DB insert that sets Status = 'active' with no payment step... There's no webhook handler anywhere." Confirmed by direct code read
- **what:** The organising design principle for the billing plan: client-side success redirects must never grant entitlement — only a signature-verified Stripe webhook, deduplicated by `provider_event_id`, may mutate entitlement state. Directly motivated by the audited finding that today `status = active` merely means "a row exists" with zero payment verification.
- **sources:** stripe/PLAN_stripe_billing_integration.md#§2,#Appendix, stripe/001commentary.md#Stripe audit turn
- **relations:** Existing but non-functional auth-service subscription scaffold; Two-plane billing architecture
- **verify-later:** internal/auth-service/subscription/service.go

<!-- SOURCE: U13_docs024_small_dirs.md -->
### One-off credit vs recurring subscription billing model
- **category:** payments
- **status-signal:** aspirational
- **status-evidence:** PLAN §7 and §5 `billing_credits` DDL are both marked PROPOSED; no credit ledger exists in code
- **what:** Two distinct charge shapes: recurring (maintenance/tier subscription, reusing the existing but non-functional subscription scaffold) and one-off (the $5-per-site build and first-site-free grant, modelled as a `billing_credits` ledger — granted/consumed counts per client). Build proceeds only once a credit is atomically consumed via the entitlement gate.
- **sources:** stripe/PLAN_stripe_billing_integration.md#§5,§7, stripe/001commentary.md#pricing discussion turn
- **relations:** Entitlement gate architecture; Existing but non-functional auth-service subscription scaffold
- **verify-later:** proposed billing_credits, billing_events tables (auth DB)

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Existing but non-functional auth-service subscription scaffold
- **category:** payments
- **status-signal:** partial
- **status-evidence:** "GetUsageStats returns hardcoded zeros with the comment 'returning mock data,' which makes CheckQuota always pass... repository.go mixes ? (MySQL-style) placeholders... with $1 (Postgres-style)... a strong sign this module has never actually been exercised." (stripe/PLAN_stripe_billing_integration.md#§1,§Appendix); independently confirmed from the tools-workstream side (PLAN_isolated_chat_environment(5).md §13: "Billing is scaffolded, not wired," correcting that same document's own earlier "billing largely exists" assumption)
- **what:** A pre-existing `subscription` package in the auth service (models, repository, service, handlers) with a `subscriptions` table, tier constants (free/basic/premium/enterprise), `stripe_customer_id`/`stripe_subscription_id` columns, a `CheckoutSession` type, and JWT claims already carrying `client_id`+`tier` — all reusable — but verified as not wired: no Stripe SDK import anywhere, `CreateSubscription` is a bare insert with no payment step, no webhook handler exists, `CheckUsage`/`GetUsageStats` returns hardcoded zeros so quota checks always pass, and a placeholder-dialect inconsistency in `repository.go` is a strong sign the module has never run against a live database. Security consequence: any entitlement gate trusting `subscription.status` today only reflects "a row exists," not "payment cleared."
- **sources:** stripe/PLAN_stripe_billing_integration.md#§1,§Appendix, stripe/001commentary.md#Stripe audit turn, tools/tool_widget_clobber/PLAN_isolated_chat_environment(5).md#13
- **relations:** Webhook-as-only-source-of-truth billing principle; Pluggable billing provider abstraction; Ownership hierarchy reuse for entitlement scoping; Entitlement gate architecture
- **verify-later:** internal/auth-service/subscription/{models,repository,service,handlers}.go; presence/absence of a Stripe webhook handler

<!-- SOURCE: U22_recent_small_docs.md -->
### Commercial model + entitlement seams (billing adapter)
- **category:** payments
- **status-signal:** partial
- **status-evidence:** "billing/identity is mostly reuse, not new" — the auth service already has a `subscriptions` table with `stripe_customer_id`, tier definitions, JWT carrying client_id+tier; "live checkout-session creation and webhooks were not evident ... verify before relying."
- **what:** The saleability design: operator-primary (operate thousands of domains), vendor-optional (sell a domain + its backend, rarely the whole framework). Isolation unit = the satellite; separability unit = the domain (partition by site_id/domain, extractable + swappable credentials). Seams to honour now: ownership via existing clients→networks→sites hierarchy (re-parent network_id to sell), a pluggable billing adapter (Stripe first, generalise stripe_* columns to provider_*), two entitlement gates (build-submission reusing site_work_items.approval_mode → a pending_entitlement hold; maintenance-run filtering the heartbeat site-selection queries as a cost valve), a saas_cheap-vs-portfolio build-tier riding the existing batch/sync rail, and snapshot-able building blocks for whole-instance sales.
- **sources:** docs025.../PLAN_isolated_chat_environment(4).md#13, docs025.../PLAN_simple_paid_multidomain_chat(1).md#2
- **relations:** auth-service subscriptions, site_work_items.approval_mode, batch processing (scheduled→batch), building-as-a-service
- **verify-later:** auth subscriptions table + Stripe webhook wiring; site_work_items.approval_mode; heartbeat site-selection queries

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### REVIEW_BEFORE_PAY billing flow supersedes charge-first flow
- **category:** payments
- **status-signal:** partial
- **status-evidence:** `RUNBOOK_idea_uk(10).md` "Status & operating update (2026-06-11)": "Supersedes the older Flow/Email/AUTO_DELIVER notes above where they differ... `REVIEW_BEFORE_PAY` (default on)."
- **what:** idea.uk's original flow charged the customer first (Stripe Checkout), then ran the engine, then optionally held for operator review before emailing (`AUTO_DELIVER`). This was replaced by a `REVIEW_BEFORE_PAY` switch (default on): the operator's `/confirm` now *runs the engine first* and holds the draft for review; only after the operator approves does the buyer get a pay link — no money is taken until a human has seen the actual output. The original charge-first flow is kept as a fallback (`REVIEW_BEFORE_PAY=false`) "if engine cost ever spikes." A click-through token-based approve/decline UI (HMAC per order) was added on top to remove the need for curl+API-key.
- **sources:** `RUNBOOK_idea_uk(10).md` "Status & operating update (2026-06-11)"
- **relations:** idea.uk product; Stripe webhook-as-truth pattern
- **verify-later:** `idea-go/service.go` `REVIEW_BEFORE_PAY` branch

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### Stripe webhook-as-truth billing pattern (idea.uk lightweight variant)
- **category:** payments
- **status-signal:** deployed
- **status-evidence:** `RUNBOOK_idea_uk(10).md` "Stripe billing — setup" section: live keys, live webhook destination IDs, "Billing follows the PLAN_stripe_billing_integration.md principles but in the lightweight pay-per-idea shape... proven end-to-end with a real card on 2026-06-14."
- **what:** idea.uk's billing never trusts a browser redirect; only a signature-verified `checkout.session.completed` webhook (deduped by event id) marks an order paid and triggers delivery. Uses a Stripe **restricted API key** scoped to `Checkout Sessions → Write` only (least privilege — no refunds, no customer/product read access needed since Checkout uses inline `price_data`). Refunds are manual-only in the Stripe dashboard (no `/refund` endpoint exists). This is presented explicitly as the lightweight, one-off-payment implementation of the same principles as the full chassis-wide Stripe plan (see separate entry) — webhook-is-truth, idempotent, provider behind an interface (FakeProvider swap for local testing).
- **sources:** `RUNBOOK_idea_uk(10).md` §"Stripe billing — setup" (webhook destination IDs, account IDs, restricted-key scoping, troubleshooting runbook for a real signature-mismatch incident on 2026-06-14)
- **relations:** chassis-wide Stripe billing integration plan (supersedes/generalizes); REVIEW_BEFORE_PAY flow
- **verify-later:** Stripe dashboard accounts `acct_1RNfPY08YuzM2cqf` (test) / `acct_1RNfPL02nQ76FNif` (live)

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### Chassis-wide Stripe billing integration plan (client_entitlements cache)
- **category:** payments
- **status-signal:** aspirational
- **status-evidence:** Doc self-describes as "PROPOSED" throughout ("Schema caveat: this plan is written from the auth subscription Go models, not the auth DB migrations... Every DDL below is PROPOSED"). No claim of implementation for the chassis-wide version (idea.uk implemented its own lighter variant instead — see above).
- **what:** A designed-not-built architecture for platform-wide billing: auth service owns billing truth (subscriptions, one-off credits, webhook-verified events only), chassis reads through a one-directionally-fed cache table `client_entitlements` (never calls auth synchronously from the hot path), with two gates — a low-volume build-submission gate (`approval_mode='pending_entitlement'`) and a high-volume maintenance-run gate (join-filter on heartbeat queries). Covers both recurring (maintenance/tier) and one-off ($5 build credit) charge shapes, a provider interface abstracting Stripe, and a verified-findings appendix showing the existing auth subscription code is a non-functional scaffold (`CreateSubscription` stamps `status=active` with no payment, no Stripe SDK, no webhook handler, mock usage stats, and a `?`/`$1` placeholder-dialect mismatch implying the code was never run against one DB engine).
- **sources:** `docubundle_idea_golive/package_module/output_contexts/PLAN_stripe_billing_integration.md` (packaged context-pack snapshot, 390 lines); archive `PLAN_stripe_billing_integration(2).md` (identical to live `(3).md`)
- **relations:** Stripe webhook-as-truth pattern (idea.uk's lighter realisation of the same principles); isolated-chat-environment commercial model (referenced, live doc)
- **verify-later:** `internal/auth-service/subscription/{models,repository,service,handlers}.go` — confirm whether the scaffold described (no Stripe SDK, mock usage stats, dialect mismatch) is still the current state

<!-- SOURCE: U04_idea_uk.md -->
### Stripe integration pattern: webhook as the only source of truth
- **category:** payments
- **status-signal:** deployed
- **status-evidence:** Live £29 payments proven end-to-end 2026-06-14 (incl. resolving the stray-character webhook-secret incident); full setup documented from the real dashboards.
- **what:** The reference payments pattern proven by idea.uk: entitlement/fulfilment granted only on a signature-verified `checkout.session.completed` (browser redirects prove nothing); webhook handling idempotent via an event-dedup table; a **restricted** API key scoped to Checkout Sessions:Write only; test and live are separate accounts with separate webhook destinations and secrets ("a sandbox webhook does not cover live"); the signing secret must be byte-exact (one pasted stray character 400'd every event and stalled a paid order — recovered by resending the event); Stripe keeps its fee on refunds; no SDK — raw HTTP + HMAC verify.
- **sources:** idea.uk/RUNBOOK_idea_uk(9).md (Stripe billing — setup + troubleshooting); idea.uk/PLAN_stripe_billing_integration(3).md (idea.uk reference block); idea.uk/golang_files/billing.go (header)
- **relations:** platform billing plan (adopts these principles); request-then-confirm flow.
- **verify-later:** billing.go webhook verify (HMAC-SHA256 over timestamp+body, constant-time compare).

<!-- SOURCE: U04_idea_uk.md -->
### Platform Stripe billing integration plan (auth-service truth + chassis entitlement cache)
- **category:** payments
- **status-signal:** aspirational
- **status-evidence:** "the auth service has a subscription scaffold… but no working payment integration — no Stripe SDK, no checkout creation, no webhooks"; every DDL marked PROPOSED.
- **what:** The chassis-wide billing design for the build/host/chat product: truth lives in the auth DB mutated only by verified webhooks; the chassis gates on a one-way-fed `client_entitlements` cache (Kafka entitlement-changed events + reconciliation sweep) because the maintenance heartbeat can't call auth per site; two charge shapes — recurring tier subscription per client and a one-off **$5 build credit** (Checkout mode=payment, consumed via the atomic-claim idiom); build-submission gate reuses the `approval_mode` hold; provider interface from day one. idea.uk is the cited working reference for the one-off path.
- **sources:** idea.uk/PLAN_stripe_billing_integration(3).md; idea.uk/RUNBOOK_idea_uk(9).md (reference implementation)
- **relations:** Stripe webhook pattern; admin-dashboard-and-api (auth service); scheduler heartbeats.
- **verify-later:** auth service repo subscriptions tables; any client_entitlements table (expect absent).

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Entitlement gate architecture (build-submission + maintenance-run gates)
- **category:** payments
- **status-signal:** aspirational
- **status-evidence:** PLAN doc §8 describes both gates in the future tense as design; build order (§10) lists them as unbuilt steps
- **what:** Two entitlement checkpoints reusing existing chassis mechanisms: (1) build-submission gate — a new `pending_entitlement` hold state on `site_work_items.approval_mode` (mirroring the existing hitl/pending_review pattern), parking the first expensive work item until a billing check clears, with atomic credit consumption via the same UPDATE...RETURNING idiom as `claim_work_item`; (2) maintenance-run gate — a join-filter added to the three heartbeat selection queries requiring `maintenance_active`, valuable even before any domain is sold.
- **sources:** stripe/PLAN_stripe_billing_integration.md#§8, stripe/001commentary.md#final turns
- **relations:** Two-plane billing architecture; Ownership hierarchy reuse for entitlement scoping; One-off credit vs recurring subscription billing model
- **verify-later:** site_work_items.approval_mode values; build-pipeline-trigger/improvement-loop/content-feed-trigger selection SQL

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Two-plane billing architecture (auth-service truth + chassis entitlement cache)
- **category:** payments
- **status-signal:** aspirational
- **status-evidence:** PLAN §3: "Truth = auth DB, mutated only by webhooks... Gate reads = chassis client_entitlements cache, fed one-directionally from auth" — all proposed, table marked PROPOSED
- **what:** Splits billing across two databases/services with one directional bridge: the auth service owns billing truth (subscriptions, credits, webhook-driven events); the chassis reads a local `client_entitlements` cache table fed by an entitlement-changed Kafka event plus a reconciliation sweep backstop — required because the maintenance heartbeat must join across thousands of sites per tick.
- **sources:** stripe/PLAN_stripe_billing_integration.md#§3,§5
- **relations:** Entitlement gate architecture; Isolated chat/satellite architecture (Y-copy); Pluggable billing provider abstraction
- **verify-later:** proposed table client_entitlements; entitlement-changed Kafka event/consumer (not built)

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Pluggable billing provider abstraction (Stripe as implementation #1)
- **category:** payments
- **status-signal:** aspirational
- **status-evidence:** PLAN §4 gives a full Go interface sketch explicitly labelled "Sketch, not final"; current code has "no Stripe SDK" imported at all
- **what:** A `Provider` interface (`EnsureCustomer`, `CreateSubscriptionCheckout`, `CreateOneOffCheckout`, `CreatePortalSession`, `CancelSubscription`, `ParseWebhook`) behind which Stripe is the first implementation, normalising provider-specific webhook payloads into a provider-agnostic `Event` type. Justified as "zero retrofit cost" specifically because no Stripe integration exists yet.
- **sources:** stripe/PLAN_stripe_billing_integration.md#§4,#TL;DR
- **relations:** Two-plane billing architecture; Existing but non-functional auth-service subscription scaffold
- **verify-later:** internal/auth-service/subscription/{models,repository,service,handlers}.go

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Webhook-as-only-source-of-truth billing principle
- **category:** payments
- **status-signal:** partial
- **status-evidence:** "service.go imports only uuid, zap, time, context, fmt — no Stripe SDK. CreateSubscription is a bare DB insert that sets Status = 'active' with no payment step... There's no webhook handler anywhere." Confirmed by direct code read
- **what:** The organising design principle for the billing plan: client-side success redirects must never grant entitlement — only a signature-verified Stripe webhook, deduplicated by `provider_event_id`, may mutate entitlement state. Directly motivated by the audited finding that today `status = active` merely means "a row exists" with zero payment verification.
- **sources:** stripe/PLAN_stripe_billing_integration.md#§2,#Appendix, stripe/001commentary.md#Stripe audit turn
- **relations:** Existing but non-functional auth-service subscription scaffold; Two-plane billing architecture
- **verify-later:** internal/auth-service/subscription/service.go

<!-- SOURCE: U13_docs024_small_dirs.md -->
### One-off credit vs recurring subscription billing model
- **category:** payments
- **status-signal:** aspirational
- **status-evidence:** PLAN §7 and §5 `billing_credits` DDL are both marked PROPOSED; no credit ledger exists in code
- **what:** Two distinct charge shapes: recurring (maintenance/tier subscription, reusing the existing but non-functional subscription scaffold) and one-off (the $5-per-site build and first-site-free grant, modelled as a `billing_credits` ledger — granted/consumed counts per client). Build proceeds only once a credit is atomically consumed via the entitlement gate.
- **sources:** stripe/PLAN_stripe_billing_integration.md#§5,§7, stripe/001commentary.md#pricing discussion turn
- **relations:** Entitlement gate architecture; Existing but non-functional auth-service subscription scaffold
- **verify-later:** proposed billing_credits, billing_events tables (auth DB)

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Existing but non-functional auth-service subscription scaffold
- **category:** payments
- **status-signal:** partial
- **status-evidence:** "GetUsageStats returns hardcoded zeros with the comment 'returning mock data,' which makes CheckQuota always pass... repository.go mixes ? (MySQL-style) placeholders... with $1 (Postgres-style)... a strong sign this module has never actually been exercised." (stripe/PLAN_stripe_billing_integration.md#§1,§Appendix); independently confirmed from the tools-workstream side (PLAN_isolated_chat_environment(5).md §13: "Billing is scaffolded, not wired," correcting that same document's own earlier "billing largely exists" assumption)
- **what:** A pre-existing `subscription` package in the auth service (models, repository, service, handlers) with a `subscriptions` table, tier constants (free/basic/premium/enterprise), `stripe_customer_id`/`stripe_subscription_id` columns, a `CheckoutSession` type, and JWT claims already carrying `client_id`+`tier` — all reusable — but verified as not wired: no Stripe SDK import anywhere, `CreateSubscription` is a bare insert with no payment step, no webhook handler exists, `CheckUsage`/`GetUsageStats` returns hardcoded zeros so quota checks always pass, and a placeholder-dialect inconsistency in `repository.go` is a strong sign the module has never run against a live database. Security consequence: any entitlement gate trusting `subscription.status` today only reflects "a row exists," not "payment cleared."
- **sources:** stripe/PLAN_stripe_billing_integration.md#§1,§Appendix, stripe/001commentary.md#Stripe audit turn, tools/tool_widget_clobber/PLAN_isolated_chat_environment(5).md#13
- **relations:** Webhook-as-only-source-of-truth billing principle; Pluggable billing provider abstraction; Ownership hierarchy reuse for entitlement scoping; Entitlement gate architecture
- **verify-later:** internal/auth-service/subscription/{models,repository,service,handlers}.go; presence/absence of a Stripe webhook handler

<!-- SOURCE: U22_recent_small_docs.md -->
### Commercial model + entitlement seams (billing adapter)
- **category:** payments
- **status-signal:** partial
- **status-evidence:** "billing/identity is mostly reuse, not new" — the auth service already has a `subscriptions` table with `stripe_customer_id`, tier definitions, JWT carrying client_id+tier; "live checkout-session creation and webhooks were not evident ... verify before relying."
- **what:** The saleability design: operator-primary (operate thousands of domains), vendor-optional (sell a domain + its backend, rarely the whole framework). Isolation unit = the satellite; separability unit = the domain (partition by site_id/domain, extractable + swappable credentials). Seams to honour now: ownership via existing clients→networks→sites hierarchy (re-parent network_id to sell), a pluggable billing adapter (Stripe first, generalise stripe_* columns to provider_*), two entitlement gates (build-submission reusing site_work_items.approval_mode → a pending_entitlement hold; maintenance-run filtering the heartbeat site-selection queries as a cost valve), a saas_cheap-vs-portfolio build-tier riding the existing batch/sync rail, and snapshot-able building blocks for whole-instance sales.
- **sources:** docs025.../PLAN_isolated_chat_environment(4).md#13, docs025.../PLAN_simple_paid_multidomain_chat(1).md#2
- **relations:** auth-service subscriptions, site_work_items.approval_mode, batch processing (scheduled→batch), building-as-a-service
- **verify-later:** auth subscriptions table + Stripe webhook wiring; site_work_items.approval_mode; heartbeat site-selection queries

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### REVIEW_BEFORE_PAY billing flow supersedes charge-first flow
- **category:** payments
- **status-signal:** partial
- **status-evidence:** `RUNBOOK_idea_uk(10).md` "Status & operating update (2026-06-11)": "Supersedes the older Flow/Email/AUTO_DELIVER notes above where they differ... `REVIEW_BEFORE_PAY` (default on)."
- **what:** idea.uk's original flow charged the customer first (Stripe Checkout), then ran the engine, then optionally held for operator review before emailing (`AUTO_DELIVER`). This was replaced by a `REVIEW_BEFORE_PAY` switch (default on): the operator's `/confirm` now *runs the engine first* and holds the draft for review; only after the operator approves does the buyer get a pay link — no money is taken until a human has seen the actual output. The original charge-first flow is kept as a fallback (`REVIEW_BEFORE_PAY=false`) "if engine cost ever spikes." A click-through token-based approve/decline UI (HMAC per order) was added on top to remove the need for curl+API-key.
- **sources:** `RUNBOOK_idea_uk(10).md` "Status & operating update (2026-06-11)"
- **relations:** idea.uk product; Stripe webhook-as-truth pattern
- **verify-later:** `idea-go/service.go` `REVIEW_BEFORE_PAY` branch

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### Stripe webhook-as-truth billing pattern (idea.uk lightweight variant)
- **category:** payments
- **status-signal:** deployed
- **status-evidence:** `RUNBOOK_idea_uk(10).md` "Stripe billing — setup" section: live keys, live webhook destination IDs, "Billing follows the PLAN_stripe_billing_integration.md principles but in the lightweight pay-per-idea shape... proven end-to-end with a real card on 2026-06-14."
- **what:** idea.uk's billing never trusts a browser redirect; only a signature-verified `checkout.session.completed` webhook (deduped by event id) marks an order paid and triggers delivery. Uses a Stripe **restricted API key** scoped to `Checkout Sessions → Write` only (least privilege — no refunds, no customer/product read access needed since Checkout uses inline `price_data`). Refunds are manual-only in the Stripe dashboard (no `/refund` endpoint exists). This is presented explicitly as the lightweight, one-off-payment implementation of the same principles as the full chassis-wide Stripe plan (see separate entry) — webhook-is-truth, idempotent, provider behind an interface (FakeProvider swap for local testing).
- **sources:** `RUNBOOK_idea_uk(10).md` §"Stripe billing — setup" (webhook destination IDs, account IDs, restricted-key scoping, troubleshooting runbook for a real signature-mismatch incident on 2026-06-14)
- **relations:** chassis-wide Stripe billing integration plan (supersedes/generalizes); REVIEW_BEFORE_PAY flow
- **verify-later:** Stripe dashboard accounts `acct_1RNfPY08YuzM2cqf` (test) / `acct_1RNfPL02nQ76FNif` (live)

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### Chassis-wide Stripe billing integration plan (client_entitlements cache)
- **category:** payments
- **status-signal:** aspirational
- **status-evidence:** Doc self-describes as "PROPOSED" throughout ("Schema caveat: this plan is written from the auth subscription Go models, not the auth DB migrations... Every DDL below is PROPOSED"). No claim of implementation for the chassis-wide version (idea.uk implemented its own lighter variant instead — see above).
- **what:** A designed-not-built architecture for platform-wide billing: auth service owns billing truth (subscriptions, one-off credits, webhook-verified events only), chassis reads through a one-directionally-fed cache table `client_entitlements` (never calls auth synchronously from the hot path), with two gates — a low-volume build-submission gate (`approval_mode='pending_entitlement'`) and a high-volume maintenance-run gate (join-filter on heartbeat queries). Covers both recurring (maintenance/tier) and one-off ($5 build credit) charge shapes, a provider interface abstracting Stripe, and a verified-findings appendix showing the existing auth subscription code is a non-functional scaffold (`CreateSubscription` stamps `status=active` with no payment, no Stripe SDK, no webhook handler, mock usage stats, and a `?`/`$1` placeholder-dialect mismatch implying the code was never run against one DB engine).
- **sources:** `docubundle_idea_golive/package_module/output_contexts/PLAN_stripe_billing_integration.md` (packaged context-pack snapshot, 390 lines); archive `PLAN_stripe_billing_integration(2).md` (identical to live `(3).md`)
- **relations:** Stripe webhook-as-truth pattern (idea.uk's lighter realisation of the same principles); isolated-chat-environment commercial model (referenced, live doc)
- **verify-later:** `internal/auth-service/subscription/{models,repository,service,handlers}.go` — confirm whether the scaffold described (no Stripe SDK, mock usage stats, dialect mismatch) is still the current state
