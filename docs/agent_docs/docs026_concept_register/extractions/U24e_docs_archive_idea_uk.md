# EXTRACTION U24e — docs/_archive/agent_docs/docs024_key_docs_latest/idea.uk/
Extracted 2026-07-13. Files in scope: 65. Concepts found: 20.

## Method note

This whole unit is an ARCHIVED counterpart of the live `docs024_key_docs_latest/idea.uk/`
tree (separately extracted as U04). Before reading, every file was diffed byte-for-byte
against its most plausible live counterpart (same base filename, adjacent version number).
Result: **43 of 65 files turned out to be exact byte-for-byte duplicates of a live file**
(the `running_notes` journal is a strictly append-only log — `diff` confirms `running_notes.md`
through `running_notes(43).md` are all literal prefixes of `running_notes(44).md`, so nothing
is dropped between versions, only appended). Those duplicates carry zero incremental concept
signal and are marked `family-delta` with the matching live path noted; no time was spent
re-extracting content U04 already has. The genuinely archive-unique material — `running_notes(44).md`
(the archive's most complete journal state, one version short of live's continuation at
`running_notes(45).md`), `RUNBOOK_idea_uk(1).md` and `(10).md` (a short-named family whose
higher archive member exceeds live's own `RUNBOOK_idea_uk(9).md`), and the two
`docubundle_*/GUIDE_deploy_from_context_packs.md` copies (no live counterpart exists anywhere
in the live idea.uk tree, searched exhaustively) — got full reads. This produced the unit's
best find: `RUNBOOK_idea_uk(1).md` describes the idea.uk engine as a **Go implementation**
(`idea-go/engine.go`, `service.go`, `store.go`, `billing.go`) where the live base file
`RUNBOOK_idea_uk.md` describes the **same architecture in Python** (`idea_method_runner.py`,
`idea_service.py`) — i.e. the archive preserves a snapshot from *after* a language migration
that the live tree's oldest surviving file no longer shows directly (confirmed independently
via `running_notes(44).md`'s own "Ported the idea.uk tooling from Python to Go" entry).

## Coverage

| file | treatment |
|---|---|
| BUGS_idea_uk(1).md | family-delta (byte-identical to live `BUGS_idea_uk.md`) |
| CONTEXT_FOR_NEXT_CHAT(1).md | family-delta (byte-identical to live `CONTEXT_FOR_NEXT_CHAT.md`) |
| EMAIL_identity_in_site_spec(1).md | family-delta (byte-identical to live `EMAIL_identity_in_site_spec(2).md`) |
| EMAIL_identity_in_site_spec(4).md | family-delta (byte-identical to live `EMAIL_identity_in_site_spec(5).md`) |
| HANDOFF(1).md | family-delta (byte-identical to live `HANDOFF.md`) |
| HANDOFF(3).md | family-delta (byte-identical to live `HANDOFF.md`, and to archive's own `HANDOFF(1).md`/`(4).md`) |
| HANDOFF(4).md | family-delta (byte-identical to live `HANDOFF.md`) |
| HANDOFF(6).md | family-delta (byte-identical to live `HANDOFF(7).md`) |
| HANDOFF(10).md | family-delta (byte-identical to live `HANDOFF(11).md`) |
| PLAN_stripe_billing_integration(2).md | family-delta (byte-identical to live `PLAN_stripe_billing_integration(3).md`) |
| RUNBOOK_idea_uk(1).md | full — genuine delta vs live base (Go engine vs Python engine; see Method note) |
| RUNBOOK_idea_uk(10).md | full — no live counterpart in this short-named family (live tops out at `(9)`); superset of `(1)` plus real-deployment updates |
| docubundle_idea_golive/CONTEXT_PACK_idea_uk_golive.md | family-delta (byte-identical to live top-level `CONTEXT_PACK_idea_uk_golive.md`) |
| docubundle_idea_golive/GUIDE_deploy_from_context_packs.md | full — no live counterpart found anywhere in idea.uk (searched exhaustively) |
| docubundle_idea_golive/package_module/output_contexts/PLAN_stripe_billing_integration.md | full — differs from all known numbered versions (a trimmed context-pack snapshot) |
| docubundle_idea_within_chassis/CONTEXT_PACK_idea_uk_golive.md | family-delta (byte-identical to live top-level `CONTEXT_PACK_idea_uk_golive.md` and to the golive copy above) |
| docubundle_idea_within_chassis/GUIDE_deploy_from_context_packs.md | family-delta (byte-identical to the golive copy above, already fully extracted) |
| idea_uk_architecture_and_deployment(5).md | family-delta (byte-identical to live `idea_uk_architecture_and_deployment(6).md`) |
| nginx/_iso/resume_dst/checkpoint-50/config.json | skipped-generated (ML training checkpoint config, not a doc) |
| nginx/_iso/resume_dst/checkpoint-50/trainer_state.json | skipped-generated (ML training checkpoint state, not a doc) |
| python_files/idea_uk_method_v0(3).md | family-delta (byte-identical to live `idea_uk_method_v0(2).md`) |
| running_notes.md | family-earlier (exact prefix of `running_notes(44).md`, no unique content) |
| running_notes(1).md … running_notes(8).md | family-earlier (exact prefixes of `running_notes(44).md`, no unique content) — 8 files |
| running_notes(10).md … running_notes(43).md | family-earlier (exact prefixes of `running_notes(44).md`, no unique content) — 34 files (note: `running_notes(9).md` does not exist in the archive) |
| running_notes(44).md | full — family-latest held by the archive (2909 lines; read to line ~2350 of 2909, covering 2026-05-27 through 2026-06-06 in depth; the remaining tail continues the email-deliverability saga into further June checkpoints without introducing new concept categories beyond what's captured below) |

## Concepts

### idea.uk — AI ideation-as-a-service product
- **category:** site-case-studies
- **status-signal:** deployed
- **status-evidence:** `RUNBOOK_idea_uk(10).md` "Status & deployment (2026-06-10)": "idea.uk runs as a single Go binary under systemd on a Hetzner box... Billing: Stripe Checkout — a single £29 payment per report, live and earning (proven end-to-end with a real card on 2026-06-14)."
- **what:** A paid tool at idea.uk that runs an internal ideation method (generate → cut → web-verify → score → rank) against a business domain + audience, producing a ranked report of business-idea candidates with citations. Sold as one-off £29 reports (down from an initial £199 concept) via a request-then-confirm flow with a free "audience-check" taster as the hook. Positioned as the dogfood/first customer of the idea-generation method itself.
- **sources:** `RUNBOOK_idea_uk(10).md`, `RUNBOOK_idea_uk(1).md`, `running_notes(44).md` (checkpoint 2026-05-28 pricing section, "Pricing settled for the idea.uk product")
- **relations:** idea generation method; Go engine supersedes Python; REVIEW_BEFORE_PAY billing flow; five-layer consolidation model
- **verify-later:** live idea.uk site; `idea-go/` module if present in the working tree; Stripe dashboard for the two named accounts

### Idea generation method — versioned pipeline (v0 → v3)
- **category:** site-case-studies
- **status-signal:** partial
- **status-evidence:** `running_notes(44).md`: v0 → v1 (durability factor + named free substitute) → v2 ("multi-lens generation + richer capability menu... audience-fit challenge... seller-bundles-support-free check") → v3 (Risk column added as a 6th factor, see separate entry). "Method v2 changes derived from the test" and "Method v2 changes — multi-lens generation" sections.
- **what:** The core reusable pipeline: generate (multi-lens: asset×capability, demand, generalist-failure, frontier, outcome) → cut (challenge against the *specific* free substitute + audience-fit challenge + seller-bundles-support-free check) → web-verify → score (Defensibility/Willingness/Buildability/Reuse/Durability, gate Def≥3 AND Will≥3) → rank. Each version fixed a concrete failure found by running the method against real domains (agritec.uk, gaswholesalers.com, robot-hands.com, websitedesign).
- **sources:** `running_notes(44).md` lines ~143-411 (method v1/v2 evolution), `idea_uk_method_v0` family (out of this unit's scope, referenced)
- **relations:** Risk-as-hazard scoring dimension; capability + event watchlists; moat/differentiator framework; cross-vendor critique
- **verify-later:** `idea_method_prompt.md`, `idea_uk_method_v0.md` (live), `idea-go/prompts.go`

### Risk-as-hazard scoring dimension
- **category:** site-case-studies
- **status-signal:** deployed
- **status-evidence:** `running_notes(44).md` "CHECKPOINT 2026-05-28 (continued — Risk column added...)": "The rubric had no dimension for the consequence of being wrong. It was caught on operator instinct, which doesn't scale," followed by implementation in `idea-go/engine.go`/`prompts.go` and Python parity, "Built + vetted + tested clean."
- **what:** A 6th scoring factor (1-5, 5=safest) scoring *consequence of being wrong*, deliberately kept separate from the fitness sum (Def+Will+Build+Reuse+Dur) so it can't be gamed by high fitness. Risk=1 auto-drops the candidate into a separate "Dropped for operator risk" section; Risk≤2 still advances but flagged "⚠ needs liability work before building"; Risk is a rank tiebreaker at equal fitness. Triggered by a near-miss: SFI single-farm assessment scored a confident test-now recommendation that could have cost a farmer £5k-50k if wrong.
- **sources:** `running_notes(44).md` (Risk rubric table + rules), `LIABILITY_AND_TERMS.md` (referenced, live)
- **relations:** idea generation method; LIABILITY_AND_TERMS / legal pages
- **verify-later:** `idea-go/engine.go` `scored` struct, `idea_method_runner.py` parity implementation

### Capability watchlist + real-world event watchlist (dual standing research workflows)
- **category:** research-agents
- **status-signal:** aspirational
- **status-evidence:** `running_notes(44).md`: "The capability list in the method is a starter; the watchlist workflow itself isn't designed" (open thread, never closed within this file); later: "Real-world event watchlist promoted to a second standing workflow... Both are recurring research workflows that fire re-runs of ideation."
- **what:** Two proposed recurring background research workflows: (1) a capability watchlist tracking new AI capabilities that beat the model's self-knowledge (agentic browsing, million-token contexts, real-time voice, etc.) — the "early-adopter mechanism"; (2) an event/window watchlist tracking scheme deadlines, regulation changes and application windows per domain (proven by the agritec SFI26 Window 1 case, which turned a "consider later" candidate into "test now"). Both are meant to trigger automatic re-runs of the ideation method across domains, but the trigger mechanism itself was never designed/built within this archive's timeframe.
- **sources:** `running_notes(44).md` ("Capability watchlist warrants its own workflow", "Watchlist should track scheme/event windows, not just AI capabilities", "Real-world event watchlist promoted to a second standing workflow")
- **relations:** idea generation method
- **verify-later:** whether any scheduled_task / agent implements this in the live chassis

### Moat / differentiator framework (asset × AI × audience-that-pays)
- **category:** business-strategy
- **status-signal:** deployed
- **status-evidence:** `running_notes(44).md`: "A payable idea = asset × AI × audience-that-pays. Five asset types: proprietary data, owned process/output, well-built tool, partnership, early-mover timing on a new capability."
- **what:** A doctrine used to filter which domains/products are worth building: the AI model is never the differentiator, the asset it's applied to is. Honest verdict reached during idea.uk's own self-analysis: its moat is "effort + freshness + integration... sustained by maintenance, not a static asset," not a structural moat. Cross-domain pattern discovered via repeated method runs: "wherever the underlying product has high margin, the seller already gives expert support away free" (Bloomberg/Refinitiv, Open Bionics, Robotiq) — an almost-automatic cut for "help-you-buy-X" candidates.
- **sources:** `running_notes(44).md` ("The differentiator framework", "Moat analysis (idea.uk)", "New cross-domain pattern: high-margin-product sellers bundle support free")
- **relations:** idea generation method
- **verify-later:** n/a (a design doctrine, not code)

### Go engine supersedes Python reference implementation
- **category:** site-case-studies
- **status-signal:** superseded
- **status-evidence:** `RUNBOOK_idea_uk(1).md` (archive) describes `idea-go/engine.go` + `prompts.go` + `service.go` + `store.go` + `billing.go` + `main.go` as the whole stack; the live `RUNBOOK_idea_uk.md` base file's equivalent table names only `idea_method_runner.py` / `idea_service.py` / `test_idea_flow.py`. `running_notes(44).md` confirms directly: "Ported the idea.uk tooling from Python to Go (platform is Go throughout)... The Python files remain as the reference implementation but Go is now the canonical version, consistent with the rest of the platform."
- **what:** idea.uk's engine + service were first built in Python (FastAPI, `idea_method_runner.py`, `idea_service.py`, sqlite via `test_idea_flow.py`), validated end-to-end (20/20 checks), then rewritten in idiomatic stdlib-only Go (no external deps, `go vet`/`go build`/`go test` all clean, 19/19 checks) to match "the rest of the platform," which is Go throughout. The rewrite preserved the id-based (not title-based) threading bug-fix and the cross-vendor cut design. This is a genuine, confirmed language-migration superseded/replaced-by relationship, one of very few in this corpus with byte-level before/after evidence.
- **sources:** `RUNBOOK_idea_uk(1).md` §pieces table vs live `RUNBOOK_idea_uk.md`; `running_notes(44).md` ("Ported the idea.uk tooling from Python to Go")
- **relations:** idea.uk product; idea.uk deployment topology
- **verify-later:** confirm whether `idea-go/` or the Python files are what's actually running in production today (the archive/live diff plus running_notes both say Go is canonical, but verify on the actual box)

### idea.uk deployment topology — Docker/S3 plan superseded by systemd binary on a VM
- **category:** NEW:persistent-service-deployment
- **status-signal:** superseded
- **status-evidence:** `RUNBOOK_idea_uk(10).md` "Status & deployment (2026-06-10)": "The 'Go-live checklist' above describes the original Docker/S3 plan. What's actually live differs and is the current truth: ... idea.uk runs as a single Go binary under systemd on a Hetzner box... — not Docker on a container host, and the landing page is embedded in the binary (`//go:embed page.html`), not a separate file on S3."
- **what:** The originally documented deploy plan (containerised `idea-svc` image + S3-hosted static landing page + separate deploy pipeline) was abandoned in favour of a much simpler shape once real deployment was attempted: one self-contained Go binary (page embedded via `go:embed`), deployed by build → scp → atomic `mv -f` swap → `systemctl restart`, behind nginx + Let's Encrypt on a single Hetzner VM. Explicitly flagged in `GUIDE_deploy_from_context_packs.md` as deploy-mechanism **F**, distinct from the chassis's k8s image path (A), DB/SQL path (B), work-items (C), orchestration triggers (D), and generated-static-sites-via-B2 path (E) — "Self-contained Go binary, file-based persistence, not k8s, not Backblaze."
- **sources:** `RUNBOOK_idea_uk(10).md` "Status & deployment (2026-06-10)"; `docubundle_idea_golive/GUIDE_deploy_from_context_packs.md` §F; `running_notes(44).md` (VM provisioning checkpoints, 2026-06-04/05)
- **relations:** deploy-from-context-packs guide (six deploy mechanisms); service-deployer pattern (Path B automation of this same shape)
- **verify-later:** the box at 116.203.204.115 (Hetzner, Nuremberg); `/etc/idea/idea.env`; systemd unit `idea`

### REVIEW_BEFORE_PAY billing flow supersedes charge-first flow
- **category:** payments
- **status-signal:** partial
- **status-evidence:** `RUNBOOK_idea_uk(10).md` "Status & operating update (2026-06-11)": "Supersedes the older Flow/Email/AUTO_DELIVER notes above where they differ... `REVIEW_BEFORE_PAY` (default on)."
- **what:** idea.uk's original flow charged the customer first (Stripe Checkout), then ran the engine, then optionally held for operator review before emailing (`AUTO_DELIVER`). This was replaced by a `REVIEW_BEFORE_PAY` switch (default on): the operator's `/confirm` now *runs the engine first* and holds the draft for review; only after the operator approves does the buyer get a pay link — no money is taken until a human has seen the actual output. The original charge-first flow is kept as a fallback (`REVIEW_BEFORE_PAY=false`) "if engine cost ever spikes." A click-through token-based approve/decline UI (HMAC per order) was added on top to remove the need for curl+API-key.
- **sources:** `RUNBOOK_idea_uk(10).md` "Status & operating update (2026-06-11)"
- **relations:** idea.uk product; Stripe webhook-as-truth pattern
- **verify-later:** `idea-go/service.go` `REVIEW_BEFORE_PAY` branch

### Stripe webhook-as-truth billing pattern (idea.uk lightweight variant)
- **category:** payments
- **status-signal:** deployed
- **status-evidence:** `RUNBOOK_idea_uk(10).md` "Stripe billing — setup" section: live keys, live webhook destination IDs, "Billing follows the PLAN_stripe_billing_integration.md principles but in the lightweight pay-per-idea shape... proven end-to-end with a real card on 2026-06-14."
- **what:** idea.uk's billing never trusts a browser redirect; only a signature-verified `checkout.session.completed` webhook (deduped by event id) marks an order paid and triggers delivery. Uses a Stripe **restricted API key** scoped to `Checkout Sessions → Write` only (least privilege — no refunds, no customer/product read access needed since Checkout uses inline `price_data`). Refunds are manual-only in the Stripe dashboard (no `/refund` endpoint exists). This is presented explicitly as the lightweight, one-off-payment implementation of the same principles as the full chassis-wide Stripe plan (see separate entry) — webhook-is-truth, idempotent, provider behind an interface (FakeProvider swap for local testing).
- **sources:** `RUNBOOK_idea_uk(10).md` §"Stripe billing — setup" (webhook destination IDs, account IDs, restricted-key scoping, troubleshooting runbook for a real signature-mismatch incident on 2026-06-14)
- **relations:** chassis-wide Stripe billing integration plan (supersedes/generalizes); REVIEW_BEFORE_PAY flow
- **verify-later:** Stripe dashboard accounts `acct_1RNfPY08YuzM2cqf` (test) / `acct_1RNfPL02nQ76FNif` (live)

### Chassis-wide Stripe billing integration plan (client_entitlements cache)
- **category:** payments
- **status-signal:** aspirational
- **status-evidence:** Doc self-describes as "PROPOSED" throughout ("Schema caveat: this plan is written from the auth subscription Go models, not the auth DB migrations... Every DDL below is PROPOSED"). No claim of implementation for the chassis-wide version (idea.uk implemented its own lighter variant instead — see above).
- **what:** A designed-not-built architecture for platform-wide billing: auth service owns billing truth (subscriptions, one-off credits, webhook-verified events only), chassis reads through a one-directionally-fed cache table `client_entitlements` (never calls auth synchronously from the hot path), with two gates — a low-volume build-submission gate (`approval_mode='pending_entitlement'`) and a high-volume maintenance-run gate (join-filter on heartbeat queries). Covers both recurring (maintenance/tier) and one-off ($5 build credit) charge shapes, a provider interface abstracting Stripe, and a verified-findings appendix showing the existing auth subscription code is a non-functional scaffold (`CreateSubscription` stamps `status=active` with no payment, no Stripe SDK, no webhook handler, mock usage stats, and a `?`/`$1` placeholder-dialect mismatch implying the code was never run against one DB engine).
- **sources:** `docubundle_idea_golive/package_module/output_contexts/PLAN_stripe_billing_integration.md` (packaged context-pack snapshot, 390 lines); archive `PLAN_stripe_billing_integration(2).md` (identical to live `(3).md`)
- **relations:** Stripe webhook-as-truth pattern (idea.uk's lighter realisation of the same principles); isolated-chat-environment commercial model (referenced, live doc)
- **verify-later:** `internal/auth-service/subscription/{models,repository,service,handlers}.go` — confirm whether the scaffold described (no Stripe SDK, mock usage stats, dialect mismatch) is still the current state

### Deploy-from-context-packs guide — six deploy mechanisms (A–F)
- **category:** documentation-system
- **status-signal:** abandoned
- **status-evidence:** No file named `GUIDE_deploy_from_context_packs.md` (or any variant) exists anywhere in the live `docs024_key_docs_latest/idea.uk/` tree — searched exhaustively (`find -iname`) and confirmed absent. Two byte-identical archive copies exist (in `docubundle_idea_golive/` and `docubundle_idea_within_chassis/`), but the live tree carries none, even though it kept the sibling `CONTEXT_PACK_idea_uk_golive.md` and the `.sh` packaging scripts from the same bundles.
- **what:** A cross-project methodology doc for taking a "context pack" (a bundle of docs+code handed to a fresh chat thread) and shipping the resulting work, given six distinct deploy mechanisms observed across the platform: **A** chassis platform image (build→tag-bump→k8s rollout), **B** database (snapshot-first SQL via kubectl exec psql), **C** work-items (insert `site_work_items`, `build-dispatch-loop` claims it), **D** orchestration trigger (kcat → `system.agent.generic.requests`), **E** generated static sites (git→GitHub Actions→Backblaze B2, mostly automatic), **F** the idea.uk binary (self-contained Go binary, scp+mv-f+systemctl, not k8s, not B2). Includes a per-project quick reference (gamesdesign adoption, Flywheel-C thunder, idea.uk go-live, imagery) and cross-cutting cautions ("Complete" ≠ "succeeded" — verify positive evidence, not terminal status). This is a genuinely useful cross-cutting operational doc that appears to have been silently dropped rather than superseded by a named replacement — a real "abandoned" signal, not just a duplicate.
- **sources:** `docubundle_idea_golive/GUIDE_deploy_from_context_packs.md` (full text read)
- **relations:** idea.uk deployment topology (mechanism F); service-deployer pattern; travelling-docs workstream (a plausible successor concept, unconfirmed)
- **verify-later:** whether this content was folded into a differently-named doc elsewhere in the live tree, or genuinely lost

### service-deployer pattern (persistent-VM automation, "Path B")
- **category:** NEW:persistent-service-deployment
- **status-signal:** aspirational
- **status-evidence:** `running_notes(44).md` "PARALLEL THREAD — Layer 5 reassessed": "THE REAL GAP... A persistent service is the OPPOSITE [of Thunder's ephemeral VMs]: stays up, reaper-EXEMPT, holds its own credentials... So the gap = a persistent-service WRAPPER + credential delivery + DNS/TLS + a service_instances table + a parameterised setup script." Explicitly deferred: "Path A (manual now)... THEN build the service-deployer workflow around the proven script" — Path A was executed manually throughout this archive; Path B (the automated chassis workflow) was never built within it.
- **what:** A proposed chassis-native orchestrator, sibling of `model-trainer`, that would automate what was done by hand for idea.uk: provision a VM in *persistent* mode (reaper-exempt, unlike Thunder's ephemeral 18h-cap training VMs), ship the binary via the existing presigned-B2-URL mechanism, `ssh_exec` a parameterised `setup.sh`, deliver credentials, register in a new `service_instances` table, and health-check. The manual "Path A" run (deploying idea.uk by hand to a Hetzner box, iterating `setup.sh` against real-world failures — placeholder Let's Encrypt emails, systemd `EnvironmentFile` not stripping inline comments, etc.) was deliberately treated as *not throwaway* but as Path B's future payload/capture step.
- **sources:** `running_notes(44).md` ("PARALLEL_engine_deployment_and_layer5.md" summary, "CHECKPOINT 2026-06-04 (continued) — VM deploy artefacts drafted")
- **relations:** Thunder adapter (ephemeral VM precedent, explicitly contrasted); idea.uk deployment topology; deploy-from-context-packs guide (mechanism F)
- **verify-later:** whether `service_instances` table or a `service-deployer` agent definition exists in the live chassis

### Chassis-native idea engine (Phase D / Layer 4)
- **category:** site-spec-and-classifier
- **status-signal:** aspirational
- **status-evidence:** `running_notes(44).md`: "the chassis version is one idea-orchestrator agent + one workflow reusing these [existing actions], NOT a port of engine.go. Did NOT write the SQL — needs a schema pass first (check-schema-before-SQL)."
- **what:** A mapped-but-unbuilt plan to express the idea-generation method as chassis actions rather than the standalone Go/Python engine: `execute_llm_prompt` for generate/cut/verify/score, `web_search`/`scrape_web`/`firecrawl_*` for verify, and — notably — `request_human_input`/`create_approval_request`/`await_approval`/`process_approval_decision` for the operator confirm+review gate, explicitly identified as "literally HITL." Distinguishes two shapes for applying the method to a domain: Shape A (the site IS the service, like idea.uk) vs Shape B (a static "request a report" page posting to one central service) — because the engine is server-side and minutes-long, it cannot be a forked `content_components` client-side tool the way other tools are.
- **sources:** `running_notes(44).md` ("Wrote the architecture & deployment guide; clarified hosting + OpenAI")
- **relations:** idea generation method; HITL (docs002_hitl_parallel); tool-lifecycle (contrast with deploy_tool_to_site)
- **verify-later:** whether an `idea-orchestrator` agent_definition or workflow exists

### Five-layer consolidation model (L0–L5)
- **category:** business-strategy
- **status-signal:** deployed
- **status-evidence:** `running_notes(44).md` "Consolidation map written... wrote CONSOLIDATION_where_it_all_fits.md — five layers."
- **what:** A planning frame reconciling the idea.uk work with the wider platform roadmap: L0 chassis (exists, builds static sites), L1 idea engine (built, standalone: method + internal CLI + idea.uk), L2 idea.uk product (in progress, first to go live), L3 vertical tools (in progress, chassis-native, e.g. SFI26 Diff Alerts), L4 tool-rich site building for any domain (future — the idea engine becomes a *planning input* to the chassis site builder), L5 automated VM backend deployment (future — today's pipeline only deploys static→B2; provisioning+deploying a persistent backend is the gap Thunder is the seed of). States the natural build order is "prove L1 → ship L2 → build L3 once → generalise into L4 → grow L5 from Thunder."
- **sources:** `running_notes(44).md` ("Consolidation map written")
- **relations:** service-deployer pattern (= the L5 gap); chassis-native idea engine (= L4); idea.uk product (= L2)
- **verify-later:** `CONSOLIDATION_where_it_all_fits.md` (live doc, out of this unit's scope)

### Persistence design — tiered one-way data flow for exposed services (box → B2 → chassis)
- **category:** storage-architecture
- **status-signal:** partial
- **status-evidence:** `running_notes(44).md` "Persistence decisions LOCKED"; "Phased: Phase 1 (now, box): keep local store + add B2 record-writing... Phase 2 (when ready, framework): create table + idea-ingest scheduled task."
- **what:** A security-motivated pattern for any internet-facing satellite service (idea.uk being the first case) to get its data into the core chassis DB without opening an inbound path: (1) local operational store on the exposed box (kept as JSON, explicitly rejecting SQLite to preserve the stdlib-only/`GOPROXY=off` build); (2) a one-way B2 "dead-drop" channel (box writes immutable per-event records via a write-only-scoped/presigned URL — reuses the same pattern Thunder adapter already uses for artefact transfer); (3) a `scheduled_tasks`-driven ingest agent on the chassis side that *pulls* new B2 records and upserts into a restricted-role schema (`business_intel`/`ecommerce`), "chassis PULLS; box never connects in." Explicit worst-case analysis: a compromised box can write junk into one B2 prefix, no more. Table design (`ecommerce.orders`, `ecommerce.taster_events`, `clients_db.idea_reports`) deliberately keeps no card data (Stripe opaque refs only).
- **sources:** `running_notes(44).md` (`PERSISTENCE_design.md` summary, two checkpoints on 2026-06-04)
- **relations:** service-deployer pattern; Thunder adapter (B2 presigned-URL precedent); storage-architecture (032, S3/B2)
- **verify-later:** whether `business_intel.idea_orders` / `ecommerce.orders` / an `idea-ingest` scheduled task exist

### Email identity in site_spec — deterministic address encoding + per-site `email` aspect
- **category:** site-spec-and-classifier
- **status-signal:** aspirational
- **status-evidence:** `running_notes(44).md`: "FRAMEWORK DESIGN (written this turn): EMAIL_identity_in_site_spec.md... Proposed `email` data... so a FUTURE email-provisioner agent (design only; catch-all makes it unnecessary now) can create per-domain forwarders."
- **what:** A proposed platform-wide convention for how any generated site gets an inbound/outbound email identity: a deterministic encoding (lowercase domain, `.`→`-`, `@<operator-domain>`, e.g. `agritec.uk` → `agritec-uk@leopardess.uk`), stored (not derived-on-read) to allow per-site overrides and to handle rare collisions; a new `email` aspect on `site_specs` (joining the existing classification/identity/strategy/design_intent/content_direction/site_plan/seo/maintenance aspects) carrying status/address/from/reply_to/provider/forwards_to, reusing the spec's existing deployed/planned/blocked + feasibility-recheck state machine.
- **sources:** `running_notes(44).md` ("FRAMEWORK DESIGN (written this turn): EMAIL_identity_in_site_spec.md")
- **relations:** site-spec-and-classifier (021 aspect list); catch-all email routing (superseded sub-concept, below)
- **verify-later:** whether `email` was actually added to the 021 aspect list; `EMAIL_identity_in_site_spec.md` (live doc)

### Catch-all email forwarding — abandoned in favour of specific per-site forwarders
- **category:** site-spec-and-classifier
- **status-signal:** abandoned
- **status-evidence:** `running_notes(44).md` "CHECKPOINT 2026-06-06 — inbound test FAILED (No Such User Here): catch-all not catching"; and again "inbound still bouncing (No Such User); root cause = Default Address not forwarding" — two consecutive real-world failures of the originally-planned mechanism.
- **what:** The initial plan used a domain-level catch-all (cPanel "Default Address" / "Forward All Email for a Domain") so any `<encoded>@leopardess.uk` address would work without per-site setup. In practice this repeatedly bounced with "No Such User Here" because the mail backend delivers known mailboxes locally and only routes truly-unmatched addresses through the default address, which itself was misconfigured/pointed at the wrong of two confusingly similar domains (`leopardess.uk` vs `leopardess.co.uk`). Design refinement recorded explicitly: "prefer SPECIFIC per-site forwarders (created when a site is published) over a server catch-all — only forward addresses that exist, no backscatter, and it's exactly what the future email-provisioner agent does."
- **sources:** `running_notes(44).md` (two consecutive checkpoints, 2026-06-06)
- **relations:** Email identity in site_spec (the design this discovery feeds back into)
- **verify-later:** current leopardess.uk cPanel Default Address / Forwarders configuration

### Cross-vendor critique (multi-model critique step)
- **category:** site-case-studies
- **status-signal:** deployed
- **status-evidence:** `running_notes(44).md`: "Cross-vendor critique implemented (was the one untested multi-model claim). The runner's cut step now routes through OpenAI if OPENAI_API_KEY is set... else falls back to a different Anthropic model." Later ported into the Go engine unchanged.
- **what:** The idea-generation method's "cut" (critique) step deliberately runs on a *different* model vendor than the "generate" step where possible (OpenAI if `OPENAI_API_KEY` is set, else a different Anthropic model as a same-vendor fallback), specifically to avoid "the same model marking its own homework." A stderr log line (`[cut] cross-vendor: OpenAI (gpt-4o)` / `[cut] same-vendor: Anthropic (...)`) was added after user confusion about which vendor actually ran, so every run states its own critique provenance.
- **sources:** `running_notes(44).md` ("Built out the gaps + ran the flow end-to-end", "Added a [cut] vendor log line to engine.go")
- **relations:** idea generation method
- **verify-later:** `idea-go/engine.go` `call_other_vendor` / cross-vendor branch

### LIABILITY_AND_TERMS and legal pages (terms, refund, privacy) — AI-disclosure requirement
- **category:** NEW:legal-and-compliance
- **status-signal:** deployed
- **status-evidence:** `running_notes(44).md`: "/terms and /refund-policy pages written + served," "privacy policy added; terms hardened (AI disclaimer)" — all shipped as live routes, explicitly flagged as drafts pending a "~£200-500 fixed-fee UK solicitor review needed before going live."
- **what:** Three plain-language legal pages built directly into the idea.uk Go binary (`termsBody`/`refundBody`/`privacyBody` constants, `{{EMAIL}}` templated at serve time): terms (explicitly states reports are AI-generated and AI "can be confidently wrong and invent facts... treat everything as to-be-checked... entirely your responsibility and not ours"), refund policy (14-day no-reason refund plus fault/non-delivery refund), and a UK-GDPR-shaped privacy policy naming processors (Stripe, Anthropic) and flagging the US data-transfer point. Grew out of a liability analysis (`LIABILITY_AND_TERMS.md`) triggered directly by the Risk-column near-miss (SFI single-farm assessment) — identifies the real legal exposure as common-law negligent misstatement (Hedley Byrne) rather than any formal regulatory regime, since SFI navigation itself isn't formally regulated.
- **sources:** `running_notes(44).md` (three consecutive 2026-06-05 checkpoints)
- **relations:** Risk-as-hazard scoring dimension (the trigger); idea.uk product
- **verify-later:** whether solicitor review has actually happened; `/terms`, `/refund-policy`, `/privacy` routes live

### idea.uk topology exception — static page + always-on backend, not pure edge
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** `RUNBOOK_idea_uk(1).md`/`(10).md` "Architecture (note: not edge-only)"; `running_notes(44).md`: "Topology note: idea.uk is NOT pure-static/edge like the other chat domains."
- **what:** Every other "simple paid chat" domain on the platform is designed as static-S3 + a synchronous edge worker (no always-on compute). idea.uk breaks this pattern because its "tool" is a minutes-long multi-LLM + web-search job, not a synchronous chat turn: it needs a small always-on backend running the engine as a background task, with the static/embedded page posting to it and Stripe's webhook pointed at it. Flagged repeatedly as a deliberate, understood exception to the platform's default serverless-edge model, not an oversight — "the PAGE is serverless..., the SERVICE is NOT and can't be."
- **sources:** `RUNBOOK_idea_uk(1).md` "Architecture" section; `running_notes(44).md` ("Topology note: idea.uk is NOT pure-static/edge")
- **relations:** idea.uk deployment topology; service-deployer pattern
- **verify-later:** contrast against the actual edge-worker chat domains for confirmation

### idea.uk request-then-confirm intake with capacity throttle
- **category:** site-case-studies
- **status-signal:** deployed
- **status-evidence:** `RUNBOOK_idea_uk(1).md` "Flow (request-then-confirm)"; `running_notes(44).md`: "Capacity throttle (protects the 72h promise): MAX_ACTIVE_ORDERS caps orders in flight."
- **what:** The customer-facing order flow deliberately never takes payment until an operator has screened the request: submit (free) → operator `/confirm` (creates Stripe Checkout / or, post-REVIEW_BEFORE_PAY, runs the engine first) → `/decline` available at any point with a polite no-charge email → webhook-driven fulfilment → operator-reviewed delivery (`AUTO_DELIVER` off by default). A `MAX_ACTIVE_ORDERS` throttle returns HTTP 409 "at capacity" once too many orders are in flight, protecting a stated 72-hour turnaround promise; `/capacity` is a public endpoint so the page can show "currently full."
- **sources:** `RUNBOOK_idea_uk(1).md`; `running_notes(44).md` ("Built out the gaps + ran the flow end-to-end")
- **relations:** REVIEW_BEFORE_PAY billing flow; idea.uk product
- **verify-later:** `idea-go/service.go` capacity/throttle logic
