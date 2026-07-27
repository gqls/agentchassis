# Register — idea-product

> **covers-through: 2026-07-13** · extraction freeze.
> Subsystems that shipped after this date may be absent from this file
> **entirely** — absence here is not evidence of absence in the platform. See `bugs_open/106`.

15 concepts, consolidated from 30 raw extractions across unit U04. (The cluster input file contained this category's raw blocks twice, back-to-back and byte-identical; each pair is merged into one entry below. No cross-unit duplication found — all raw blocks for this category came from U04_idea_uk.md.)

### IDEA-001 — Ideation method v0→v3 (staged, multi-model, web-verified pipeline)
- **status:** deployed
- **status-evidence:** Method doc carries v1/v2/v3 patches with dated rationale; engine validated end-to-end live 2026-06-04; live product runs it per paid order.
- **what:** The runnable method: (1) frame AND challenge the audience; (2) generate candidates across four lenses — demand, generalist-failure, frontier, outcome — plus the original asset×capability sweep; (3) cut each candidate against the specific named free substitute with a different model, incl. the seller-bundles-support-free check; (4) verify survivors with web research, evidence attached; (5) score; (6) rank and split test-now vs consider. Version history is itself conceptual: v1 added Durability + the specific-substitute cut; v2 added multi-lens generation + audience-fit challenge (single-lens generation diagnosed as "narrow — supply-side only"); v3 added the Risk column. Dogfooding rule: if the method can't find an advancing candidate for idea.uk itself, it isn't good enough.
- **sources:** idea.uk/idea_uk_method_v0(3).md; idea.uk/idea_uk_testrun_v2.md; idea.uk/KEY_DOC_idea_method_prompt.md; idea.uk/running_notes(63).md (method-run checkpoints)
- **relations:** IDEA-002 operator-risk column; IDEA-004 cross-vendor critique; IDEA-003 capability watchlist; IDEA-005 engine implementations
- **verify-later:** golang_files/prompts.go step prompts vs the method doc

### IDEA-002 — Operator-risk column: hazard scored separately from fitness, with gates
- **status:** deployed
- **status-evidence:** Debugging-guide item 23 (2026-05-28) documents the addition end-to-end; A1 acceptance notes the live report carries "Operator risk: N/5" with auto-flags.
- **what:** A sixth scored dimension, Risk to the operator (1-5, 5 safest), scoring the CONSEQUENCE of being wrong, not the probability. Deliberately not added to the fitness sum and not in the Def≥3∧Will≥3 gate: Risk=1 (regulated professions) is dropped automatically into a visible "Dropped for operator risk" section; Risk≤2 advances but flagged "needs liability work before building" with the cheapest_test forced to demand PII + solicitor-reviewed T&Cs first; Risk breaks ties toward safer builds. Generalisable lesson: when a scoring system recommends actions to an operator who carries downstream exposure, hazard must be a separate scored dimension — fitness sums cannot see it. First real effect: paused the SFI single-farm assessment.
- **sources:** idea.uk/016_debugging_guide_v2_32(1).md (item 23); idea.uk/idea_uk_method_v0(3).md (Risk rubric); idea.uk/LIABILITY_AND_TERMS(2).md (header)
- **relations:** LGL-001 liability framework; IDEA-014 SFI26 Diff Alerts swap; IDEA-001 ideation method
- **verify-later:** engine.go `scored` struct + riskNote; idea_method_runner.py parity

### IDEA-003 — Capability watchlist + real-world event-window watchlist
- **status:** aspirational
- **status-evidence:** PLAN_idea_uk §8: "the capability watchlist runs as its own recurring research workflow" (stance accepted); no evidence either watchlist was ever built as a workflow.
- **what:** Two maintained lists that feed re-runs of ideation: (1) AI capabilities worth using now, grouped by what specialism does that generalists don't — the mechanism for being early to ideas a new capability just unlocked, and the single strongest durable advantage; (2) real-world event windows per domain (scheme deadlines, regulation changes — e.g. SFI26 Window 1), because timing changes what's actionable. The capability menu v1 ships inside the method/prompts; the recurring maintenance workflows remain unbuilt.
- **sources:** idea.uk/idea_uk_method_v0(3).md (capability list v1); idea.uk/PLAN_idea_uk(3).md#8; idea.uk/running_notes(63).md ("Watchlist should track scheme/event windows")
- **relations:** BIZ-004 differentiator framework (currency moat); scheduler-and-tasks (would host the recurrence)
- **verify-later:** scheduled_tasks / agent_definitions for any watchlist workflow

### IDEA-004 — Cross-vendor critique (the cut step on a different vendor)
- **status:** deployed
- **status-evidence:** Architecture doc §9: `[cut] cross-vendor: OpenAI (gpt-4o)` / `[cut] same-vendor: Anthropic` stderr line added; env-switched via OPENAI_API_KEY.
- **what:** The method's quality gate (the cut) is run by a different model from the generator — ideally a different vendor — so the method isn't one model marking its own work. Implemented as an optional OpenAI branch on the cut step (OPENAI_API_KEY + OPENAI_CRITIQUE_MODEL); same-vendor fallback still uses a different model (Sonnet vs Opus). Cross-vendor comparison flagged as an untested open experiment.
- **sources:** idea.uk/idea_uk_architecture_and_deployment(6).md#9; idea.uk/idea_uk_method_v0(3).md (diff-model markers); idea.uk/RUNBOOK_idea_uk(9).md (go-live step 5)
- **relations:** IDEA-001 ideation method; multi-model ensemble moat claim
- **verify-later:** engine.go cut-step branch

### IDEA-005 — Engine implementations: single-shot prompt → Python runner → Go engine (with LLM feature upgrade)
- **status:** deployed
- **status-evidence:** "Ported the idea.uk tooling from Python to Go (platform is Go throughout)" (running notes 2026-05-28ff); A1 DONE 2026-05-28 with live validation 2026-06-04; python files retained as superseded parity copies.
- **what:** Three coexisting expressions of the method: a paste-anywhere single-shot prompt (weakest — one model marks its own work), the Python `idea_method_runner.py`/`idea_service.py` originals (superseded), and the shipped Go engine (`engine.go`+`prompts.go`, stdlib-only, no SDKs, offline `GOPROXY=off` build). The A1 upgrade set the model line-up (Opus for generate/verify, Sonnet for cut/score, all env-overridable) and added extended thinking per step (off for brainstorm breadth), prompt caching on static system blocks, `web_search_20260209` + code-execution filtering on verify, and a `WEB_SEARCH_MAX_USES` budget (raised 6→12 after a quota-exhausted run left premises "provisional").
- **sources:** idea.uk/DEVELOPMENT_RUNBOOK(3).md#A1; idea.uk/golang_files/engine.go (header); idea.uk/python_files/idea_method_runner.py (header); idea.uk/RUNBOOK_idea_uk.md (base — Python era, family-delta)
- **relations:** LLM API shape disciplines (the three bugs found during validation); IDEA-001 ideation method
- **verify-later:** engine.go callClaudeOpts / usesAdaptiveThinking

### IDEA-006 — idea.uk service: request-then-confirm flow, REVIEW_BEFORE_PAY, AUTO_DELIVER, capacity cap
- **status:** deployed
- **status-evidence:** Live and earning; full flow proven end-to-end with a real card 2026-06-14 ("LIVE BUG RESOLVED: paid + report delivered end-to-end").
- **what:** The order state machine: visitor `/request` (free) → operator confirm/decline → pay → fulfil. Two switchable shapes: charge-first (engine runs after payment; AUTO_DELIVER=false holds the report for operator review) and review-before-pay (default from 2026-06-11: confirm runs the engine first, operator reviews the draft, `/approve` sends the pay link — money is only taken after the operator has seen the deliverable). `MAX_ACTIVE_ORDERS` caps in-flight orders so capacity can't be oversold; `/capacity` exposes it. Orders live in a JSON file store (`/var/lib/idea/orders.json`) — deliberately no DB on the exposed box.
- **sources:** idea.uk/RUNBOOK_idea_uk(9).md (flow + 2026-06-11 update); idea.uk/golang_files/service.go (header); idea.uk/idea_uk_architecture_and_deployment(6).md#5
- **relations:** PAY-008 REVIEW_BEFORE_PAY billing flow (payments-category angle on the same change); PAY-002 chassis-wide Stripe plan; B2 dead-drop persistence design (future DB); LGL-001 liability framework (operator review as mitigation)
- **verify-later:** service.go state machine + service_test.go (19+ checks)

### IDEA-007 — Free audience-check taster endpoint
- **status:** deployed
- **status-evidence:** A2 DONE 2026-05-28 with acceptance ticks; live on the page; taster now logs the result (2026-06-11 checkpoint b).
- **what:** `/audience-check` — the method's step 1 (audience challenge + 2-3 alternative audiences) exposed as a free, no-auth, ~£0.02/run, ~10s taster: the conversion hook that replaced voluntary-pay. Per-IP sliding-window rate limiting (3/h, 20/day) with Retry-After; XSS-escaped HTML fragment for direct innerHTML insertion; TASTER_ENABLED kill switch; each run logs business/audience/result as market intelligence.
- **sources:** idea.uk/DEVELOPMENT_RUNBOOK(3).md#A2; idea.uk/golang_files/audience_check.go (header); idea.uk/idea_uk_open_discussion.md#5
- **relations:** BIZ-007 voluntary-pay rejection; IDEA-001 ideation method step 1
- **verify-later:** audience_check.go limiter + tests

### IDEA-008 — Click-through operator approval links (HMAC per-order tokens)
- **status:** deployed
- **status-evidence:** Checkpoint 2026-06-11 (f)/(g): built, then "click-through confirmed working by user".
- **what:** Request/review emails carry links to a page with Confirm/Approve/Decline buttons. The link carries an HMAC(order id, INTERNAL_API_KEY) token authorising that one order only; the link opens a safe GET page (mail-scanner prefetch can't trigger anything) and the action fires only on a button POST; actions stay gated by order status so a token can't double-fire. Curl + X-Internal-Key remains the fallback.
- **sources:** idea.uk/RUNBOOK_idea_uk(9).md (2026-06-11 update); idea.uk/running_notes(63).md (checkpoints f-g)
- **relations:** IDEA-006 request-then-confirm flow; hitl (same shape as approval flows)
- **verify-later:** service.go token mint/verify

### IDEA-009 — Fake-door → intent-capture-first launch discipline
- **status:** superseded
- **status-evidence:** PLAN §7 step 4 ("intent capture first, no payment"); superseded by the live request-then-confirm flow with real Stripe (the fakedoor page became the embedded landing page).
- **what:** Launch pattern: a static landing page offering the report at a flat price, capturing intent without charging ("we reply within 24h with a confirmed slot + payment link, or a polite decline") — deliberately avoiding charge-then-fail refund overhead — with a visible monthly slot count to throttle demand to manual capacity. Also prescribed as a parallel track for the strongest single-domain candidate (agritec SFI26 checker). The page evolved into the embedded `page.html` of the live service.
- **sources:** idea.uk/PLAN_idea_uk(3).md#7; idea.uk/running_notes(63).md ("Built the idea.uk fake-door page", "Fake-door modified to intent-capture-only"); idea.uk/idea_uk_fakedoor(9).html (deployment notes header)
- **relations:** IDEA-006 request-then-confirm flow (its successor); demand-test philosophy in the method's cheapest_test
- **verify-later:** n/a (historical)

### IDEA-010 — Deliverable quality standards for reports and product emails
- **status:** deployed
- **status-evidence:** BUGS_idea_uk 2026-06-11 entries all marked "Fixed this build" with standing "for future builds" rules.
- **what:** Standing rules distilled from report-email review: every customer-facing string in plain English for a non-technical owner (jargon/acronyms treated as defects); every standalone deliverable opens with a one-paragraph plain summary of what it is; rejected options always say what the thing was and why it died; deliverables get a deliberate professional design distinct from marketing surfaces (the £29 report email: navy/gold/serif "sheet" look, unlike the landing page); illustrative examples must not leak into generated output (audience-anchored generation). Transport rule: any HTML email must be base64/quoted-printable encoded (the SMTP 998-octet line-fold corrupted raw HTML mid-tag).
- **sources:** idea.uk/BUGS_idea_uk(4).md; idea.uk/RUNBOOK_idea_uk(9).md ("HTML emails are base64-encoded")
- **relations:** content-quality (platform analogue); transactional email realities
- **verify-later:** service.go b64Body; report HTML renderer

### IDEA-011 — Chassis-native idea engine (Phase D `idea-orchestrator`)
- **status:** aspirational
- **status-evidence:** DEVELOPMENT_RUNBOOK Phase D "Not started yet… needs a schema pass first"; architecture doc §8 declines to write the SQL until the action contracts are read.
- **what:** The second way to run the method: as a chassis agent + workflow reusing existing actions almost 1:1 (execute_llm_prompt for frame/generate/cut/score, web_search for verify, HITL actions for the operator gate, store_result/write_site_spec for persistence) — for running the method internally across our own domains on schedule (the Layer-4 planning input). The billing half deliberately stays in the standalone service ("a product/payment concern, not an agent workflow"). Bundle 2 packages exactly this port task.
- **sources:** idea.uk/idea_uk_architecture_and_deployment(6).md#8; idea.uk/DEVELOPMENT_RUNBOOK(3).md (Phase D); idea.uk/BUNDLE_2_chassis_idea_engine_workflow.md
- **relations:** BIZ-003 five-layer stack (L4); development-guide conventions (every agent an orchestrator; spawn sub-agents)
- **verify-later:** agent_definitions for any idea-orchestrator; the docubundle context file

### IDEA-012 — Multi-tenant branded intake pages on one central engine (white-label Option C)
- **status:** aspirational
- **status-evidence:** open_discussion §7: Option C "RECOMMENDED… Want me to do this in the next round?" — never built.
- **what:** Other sites offer the ideation product via their own branded static request page (built through the normal pipeline, own price/copy) POSTing to the central service with a tenant_id; per-tenant Stripe branding; iframe and CNAME/reverse-proxy options analysed and rejected. Needs ~100-200 lines (tenant field on Order, tenants config, tenant-aware /request). Shape A (site IS the service) vs Shape B (request panel on a content site) hosting split defined in the architecture doc; a forked-component tool is explicitly the wrong model for a server-side paid engine — sites only ever link to it.
- **sources:** idea.uk/idea_uk_open_discussion.md#7; idea.uk/idea_uk_architecture_and_deployment(6).md#7
- **relations:** tool-library boundary (why the engine is not a content_component); site_plan blocked/planned mechanism
- **verify-later:** service.go for any tenant handling (expect none)

### IDEA-013 — Real-door streaming progress page + programmatic refund endpoint (Phase A3/A4)
- **status:** aspirational
- **status-evidence:** DEVELOPMENT_RUNBOOK A3/A4 have outputs+acceptance defined but no DONE mark; refunds confirmed manual-only in the Stripe section ("There is no refund code").
- **what:** A3: post-payment page polls `/status/{order_id}` and renders live engine progress ("generating… cutting… verifying claim 1 of N"), report renders in-browser — the "real door" UX (option (a) of the real-door analysis; the honest 72h email model shipped instead). A4: operator-gated `/refund` calling Stripe POST /v1/refunds and marking the order refunded — refunds today are manual dashboard clicks and the app doesn't record them.
- **sources:** idea.uk/DEVELOPMENT_RUNBOOK(3).md#A3-A4; idea.uk/idea_uk_open_discussion.md#3-4; idea.uk/RUNBOOK_idea_uk(9).md (Refunds — manual)
- **relations:** IDEA-006 request-then-confirm flow; PAY-001 Stripe pattern
- **verify-later:** service.go routes (expect no /status, /refund)

### IDEA-014 — SFI26 Diff Alerts (first vertical tool) — replacing the single-farm assessment
- **status:** aspirational
- **status-evidence:** "Tool swapped 2026-05-28… paused on liability grounds"; Phase C fully specified (C1-C5) with no build evidence; the base DEVELOPMENT_RUNBOOK still carries the original single-farm Phase C (family-delta capture of the abandoned product).
- **what:** The first Layer-3 vertical tool: a subscription digest for UK farm advisors summarising what changed in Defra/RPA SFI26 guidance, from a versioned scraped corpus, with every change cited to source+version, weekly, operator-reviewed for 8 issues before auto-send. Scored 19/25 with Risk 4. It replaced the SFI26 single-farm assessment (abandoned/backlogged: Risk 2 — a wrong number could cost a farmer £5-50k), the first product decision the Risk column changed. Chassis-native by design (recurring, per-user state, scheduled), the opposite plumbing to standalone idea.uk.
- **sources:** idea.uk/DEVELOPMENT_RUNBOOK(3).md (Phase C + swap note); idea.uk/DEVELOPMENT_RUNBOOK.md (base — original single-farm Phase C); idea.uk/CONSOLIDATION_where_it_all_fits.md (Layer 3)
- **relations:** IDEA-002 operator-risk column; LGL-001 liability framework (SFI T&Cs draft); vet-med-pricing (sibling scraping shape)
- **verify-later:** any SFI corpus/agent in the repo or DB (expect none)

### IDEA-015 — idea.uk standalone service page-serving and deploy gotchas
- **status:** deployed
- **status-evidence:** Debugging guide §11 added for the idea.uk service; each gotcha tied to a fixed live incident.
- **what:** The operational failure catalogue for the single-binary service: every served path needs an explicit mux handler (bare 404s on linked pages); `writeHTML` fragments vs the `a.page()` full-page brand wrapper (navigation targets must wrap; injected fragments must not); startup templating of `CONTACT_EMAIL`/`MONTH_SLOTS` placeholders; systemd EnvironmentFile keeps inline comments (crash-loop + nginx 502); certbot failure made non-fatal in setup.sh; replace a running binary by scp-to-temp + `mv -f` (text-file-busy); Let's Encrypt rejects placeholder emails.
- **sources:** idea.uk/016_debugging_guide_v2_32(1).md#11; idea.uk/golang_files/README_setup_SETUP.md; idea.uk/BUGS_idea_uk(4).md (mobile safe-area padding)
- **relations:** setup.sh; VM launch plan
- **verify-later:** service.go routes() vs page.html hrefs
