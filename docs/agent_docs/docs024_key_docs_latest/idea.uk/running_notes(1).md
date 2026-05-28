# Running notes — in-chat ideas, reasoning, and suggestions

**Purpose:** capture the reasoning, suggestions, caveats, and choice-points
surfaced in chat that aren't fully recorded in the formal docs (plans, FOCUS,
method, test runs). Living journal; appended each turn going forward.

**Conventions:**
- Chronological. One section per session/day; entries within are brief.
- Each entry names the topic, the idea/reasoning, what got chosen (if anything),
  and which formal doc it landed in (or "kept here only").
- *Suggested but not pursued* items are flagged — they may be worth revisiting.
- *Caveats* are flagged separately because they're easy to lose.
- "Standing observations" at the end holds cross-cutting principles.
- "Open threads" tracks what's in flight at the end of the session.

---

## 2026-05-27 — session arc (backfilled)

### Chatbot framing: serving is the new piece, not bounded context or recording
The chat question decomposed into three parts with very different reuse stories:
bounded context (reuse the existing RAG/`knowledge_base` and `site_specs`);
recording each turn (one new table; deliberately separate from the build-time
`llm_call_log` for retention/PII reasons); serving (the genuinely new bit,
blocked because Layer 2 doesn't exist).
**Doc:** `FOCUS_site_chatbot_edge_worker_and_context_pack.md`.

### Option A vs B for the runtime turn
Suggested A (synchronous edge worker) over B (Kafka-orchestrated agent) because
B can't stream tokens to the browser and inherits the chassis's async failure
modes (offset replay, OOM cascades, 10-min retry timeouts). The Kafka fabric is
right for durable build work, wrong for live request/response. **Caveat:** if a
spawned-per-turn agent is used, the platform's own ~12s `DefaultRemoteStartupWait`
makes it a non-starter regardless. **Doc:** FOCUS §2.

### Serverless edge vs central nginx VM (security)
Read the sister `terraform-nginx-reverse-proxy` project and found concrete
weaknesses: SSH password auth, secrets marked `sensitive = false`, exposed
Grafana/Prometheus, merge-conflict markers in `main.tf` (config drift). The
central VM model would also drag static content behind the box (DNS leaves S3),
so it loses the hack-resistance we like about S3. **Suggested:** serverless edge
worker, provider-agnostic via a thin shim. **Doc:** FOCUS §3 + Appendix A.

### Phase 4 long pole is routing, not the router code
The harder structural problem in standing up Layer 2 isn't the Go service — it's
that `/api/*` on a static-S3 domain needs a proxy/edge in front that splits
static from API. Most likely thing to be underestimated. Flagged early.
**Doc:** discussed in chat; the eventual serverless route bypassed this.

### Isolated chat environment — three vectors
Reframed isolation into load, hack, and bug as separable concerns. Crucial
correction surfaced in chat: **live traffic already never hits the cluster** (it
hits the edge worker), so isolation is about turn data, drain/analytics, and any
chat workflow code — not the request path. **Doc:** `PLAN_isolated_chat_environment.md`
§§1–2.

### Don't reuse Phase 4a multi-cluster — it's a *coupling* mechanism
The existing `va001` multi-cluster pattern shares Kafka and Postgres on purpose;
that's the inverse of isolation. Reuse the chassis binaries, action code, and B2
storage pattern, but not the shared-Kafka dispatch. **Doc:** §3 of isolation plan.

### Self-correction: ownership already exists (not "net-new identity layer")
Initially called the identity/ownership layer "net-new." Reading the schema
showed `clients → networks → sites` already exists, with `clients.external_id`
as a back-reference to external identity/billing. Corrected.
**Doc:** isolation plan §13 (corrected subsection).

### Self-correction (twice): billing is scaffolded, not "largely exists"
First turn after reading the auth Go *models* I said "billing mostly exists."
After reading `subscription/{service,repository,handlers}.go` it had to be
corrected again: no Stripe SDK, no webhook handler, `CreateSubscription` stamps
`status=active` with no payment, `GetUsageStats` returns mock zeros, and the
repository mixes `?` and `$1` placeholders (never run on one DB). **Caveat
for future:** model-first reads can mislead; verify against implementation.
**Doc:** `PLAN_stripe_billing_integration.md` §1 + corrected §13 of isolation plan.

### Build-as-a-service example reframes the satellite
Working through "type a domain, get a site" via the chatbot on design.co.uk
showed the satellite isn't a chat box — it's a customer-facing instance of the
whole platform. Strongest justification yet for full chassis. "Hybrid S3 +
lambdas" *is* the edge pattern. Two `sites` populations (core = portfolio;
satellite = customer SaaS), not a stale copy. **Doc:** isolation plan §12.

### Commercial model resolved: operator-primary, vendor-optional
Operating thousands of domains ≠ thousands of clusters. Isolate at the
*satellite*; partition for sale at the *domain*. Per-domain sell-on = re-parent
`site.network_id` + swap credentials. Pluggable billing = adapter discipline
same as everything else. **Doc:** isolation plan §13.

### Pivot to simple paid multi-domain chat — the real cost question
Surfaced the reframe: legitimate inference on a cheap model is fractions of a
cent; the cost drivers are abuse + model choice, not honest usage. So
"can't afford free" may be overstated; free-with-tight-caps could be affordable.
Charging the visitor is also the *expensive* path because card processing has a
fixed per-transaction fee that punishes sub-£5 charges. **Doc:** `PLAN_simple_paid_multidomain_chat.md` §3.

### Per-domain monetisation by domain type
Suggested split: chat-is-the-product domains (visitor pays) vs business-chasing-
leads domains (free + capture). Same worker, different mode flag. Resolved for
the first batch: chat-is-the-product, day-pass. **Doc:** §5.

### Day-pass collapses payment complexity
A time-pass token is stateless and validates by signature + expiry — no edge KV,
no decrement, no webhook on the critical path. Issue via a synchronous
`/redeem` endpoint against the payment provider. Counted credits would need
edge state; only worth it if a pass doesn't fit. **Doc:** §6.

### The differentiator framework
The AI model isn't the differentiator; the asset it's applied to is. A payable
idea = asset × AI × audience-that-pays. Five asset types: proprietary data,
owned process/output, well-built tool, partnership, early-mover timing on a new
capability. Maintain two menus (assets, current capabilities) and cross them.
**Doc:** §10.

### The reusability trap
Bespoke tools don't carry across unrelated domains the way grounding does, so
"many domains each with a payable tool" is many separate builds. *Suggested
shape:* one or two reusable differentiators + a few high-value single domains;
avoid the bespoke-tool-for-one-mid-value-domain combination. **Doc:** §10.

### idea.uk: internal-first, sale-ready, recursive
Suggested treating idea.uk as one configured paid-chat domain whose tool *is*
the ideation method. Internal version uses our assets; hosted version uses the
user's. Keep assets as input data (so the same method serves both). Also keep
the workflows/actions it uses identifiable and minimal so it can be sold as a
working site. Dogfood: run the method on idea.uk itself. **Doc:** `PLAN_idea_uk.md`.

### Moat analysis (idea.uk)
Search grounding isn't a moat. Reproducible by a determined prompter:
decomposition, verification. Harder to reproduce: a maintained capability
watchlist that beats the model's self-knowledge (the early-adopter mechanism),
multi-model ensemble with cross-critique, evidence-attached verification,
accumulated memory, and the build bridge (we can prototype the idea, not just
describe it). Honest verdict: this is an effort/freshness/integration moat
sustained by maintenance, not a static asset. **Doc:** §5 of idea.uk plan.

### Capability watchlist warrants its own workflow
Suggested that maintaining the watchlist isn't ad-hoc — it's a recurring
research workflow whose output (new entries) triggers a re-run of ideation
across domains. That re-run loop *is* the early-adopter mechanism. **Doc:** §8
of idea.uk plan.

### Method test run — the critique step changed the answer
First pass scored the websitedesign "spec-as-starter-prompt" idea 15/test-now.
Critique pass against the *real* free substitute (describe it to Bolt yourself,
plus Bolt's own planning step) dropped willingness-to-pay to 2 → **failed the
gate**. Adding durability also flagged it as eroding fast. The version that
survives is an orchestration layer (spec + run + critique + iterate), which is
expensive — consider, not test-now. **Single-pass would have built a weak
product.** **Doc:** `idea_uk_testrun_v0.md` (revised).

### Method v1 changes derived from the test
Two improvements went into the method: (a) Durability factor in the rubric, 1–5
on how much base-model improvement erodes the idea; ≤2 flagged as short-lived.
(b) Cut step must name the *specific* free substitute, not test against "free AI"
in the abstract. **Doc:** `idea_uk_method_v0.md` (now effectively v1).

### Multi-model critique caveat
The critique pass above was a rigorous opposing pass by the *same* model, not a
genuinely different model. The real tool should use a different provider here;
that would likely catch more. Honest limit of the test, not a hidden flaw.
**Kept here only** (also in test-run verdict).

---

## Standing observations (cross-cutting)

- **Reuse before rebuild.** Said many times; concrete cases this session:
  `clients→networks→sites` hierarchy, `approval_mode` hold for the build gate,
  heartbeat selection queries for the maintenance gate, chassis binaries + action
  code on the satellite, JWT carrying `client_id` + `tier`.
- **Verify against implementation, not models.** The billing reframe (twice)
  taught this; scaffolds can imply working systems that aren't there.
- **Live traffic never reaches the cluster in any of the chat designs.** The
  edge worker is the surface; everything cluster-side is async/post-hoc.
- **The fixed per-transaction fee is the real micro-payment constraint**, not
  inference cost. Pushes bundles to a few pounds, not pennies.
- **The reusability axis is what stops bespoke-tool-per-domain becoming a trap.**
- **Separability is a saleability property as well as a blast-radius one.** Same
  discipline applies in two different places (cluster level, domain level,
  workflows-and-actions level for idea.uk).
- **A single generative pass can send you to build a weak product.** The
  test-run is direct evidence; the critique step is not decoration.

---

## Open threads (end of session)

- **Method v1 work outstanding:** willingness-to-pay still estimated; multi-model
  critique should be run with a genuinely different model when the tool exists.
- **Next-step options I offered, not yet chosen:**
  - Build the fake-door demand test for the *orchestration-layer* version of the
    websitedesign idea (1b), not the bare prompt (1a) which the critique failed.
  - Run the method against a third domain with a *data-asset* candidate (e.g.
    gaswholesalers' "buy a proprietary feed") to test the verification step on a
    different asset type.
- **PLAN_stripe_billing_integration.md** is settled but gated on the auth DB
  engine question; not yet started.
- **PLAN_isolated_chat_environment.md** open decisions: X vs Y satellite shape;
  capability classes for chat; turn sink (queue vs D1); pack storage location.
- **Capability list in the method** is a starter; the watchlist workflow itself
  isn't designed.

### Method run against gaswholesalers.com — first negative result, useful
Three candidate types tested. *Buy proprietary feed* failed on verification: the
big-three PRAs (Argus, Platts, ICIS) are enterprise-billion-dollar scale, sell
through Bloomberg-style channels at five-figure-plus annual fees, and licensing
restricts redistribution / LLM-grounding. *Cheap public APIs + AI* failed on
defensibility: EIA is free, OilPriceAPI from $9/mo, $15–$129/mo for real-time —
all cheap, none proprietary, free models with tool-use can hit the same APIs.
*Specialised tool for high-paid traders* failed because the stated audience is
already over-served by Bloomberg/Refinitiv. The audience pivot (SMEs) advanced
to scoring but failed the gate (Defensibility 2, Willingness 2–3).
**Result:** no candidate advances. Method recommends not building until a new
asset is acquired (partnership, expert, data deal) or the monetisation is
pivoted (lead capture / free-with-caps). **Doc:** test-run §Run 3.

### Method now produces a usable range of verdicts (not just thumbs-up)
Three runs gave three different shapes: test-now (rare), consider-with-demand-
test (idea.uk), and don't-build-yet (gaswholesalers). A method that always
advances something is rubber-stamping. The range is direct evidence the gating
rule is doing real work. **Doc:** test-run cross-run summary.

### Audience-fit surfaced as a method gap (candidate v2 improvement)
The gaswholesalers run showed a problem the rubric doesn't currently check
explicitly: the stated audience can be the *wrong* audience for the asset+
capability. High-paid traders have institutional tooling and won't pay £2-pass;
the audience that fits the cheap-data product is different (SMEs) and brings
lower willingness. *Suggested for method v2:* add an audience-fit challenge as
part of step 1 ("is this the right audience, or is there a better-fit audience
that would pay more?") — same kind of step-3-style challenge that named the
specific free substitute. **Kept here only** for now; fold into method v2 if
confirmed on another run.

### Cheap commodity data is a commodity (verification finding)
Wider implication beyond gaswholesalers: for any domain where the candidate
relies on "easily-acquired data + AI commentary", defensibility is near-zero
because free models with tool-use will hit the same APIs. This pattern will
recur across data-themed domains (agritec? financial niches?) and is worth
remembering as a near-automatic cut. **Kept here only** (general lesson).

### Implications for the first-batch shape
Combined with Run 1, the picture for the first batch firms up: websitedesign's
*orchestration-layer* variant is the candidate worth a fake-door demand test;
gaswholesalers should be removed from the paid-chat batch (pivot or shelve);
the other domains in the original list (robot-hands, agritec, etc.) haven't
been run yet. The earlier suggestion — "one or two reusable differentiators
plus a small number of high-value single domains" — is *narrowing*: it may be
"one reusable differentiator (the orchestration-layer for design customers) +
zero high-value single domains until we find one." **Kept here only.**

### robot-hands.com run — three framings, all fail, new pattern surfaces
Ran the method across three plausible audience framings (prosthetic users,
industrial gripper buyers, hobbyist makers) since the domain name is ambiguous.
All three failed. Different reason each time: free manufacturer support
(Open Bionics provides a free Customer Success Officer + Sidekick App +
open-source community designs); free supplier pre-sales engineering for B2B;
free AI + open-source culture for hobbyists. Verified via web research:
prosthetic devices £8k–60k+, 70%+ insurance-funded; Open Bionics is the
established UK player; community designs and the training app are free.
**Doc:** test-run §Run 4.

### New cross-domain pattern: high-margin-product sellers bundle support free
Across gaswholesalers (Bloomberg/Refinitiv), prosthetics (Open Bionics' free
Customer Success Officer), and industrial grippers (Robotiq/OnRobot/Schunk
free application engineers), the same shape appears: **wherever the underlying
product has high margin, the seller already gives expert support away free**
because that support helps close the sale. A paid chatbot that "helps you with
X" is structurally undercut by X's seller in those markets. This is an almost
automatic cut going forward for any "help-you-buy-X" candidate where X is a
high-margin product. *Lesson for method v2:* add a step-3 check "is the seller
of the underlying product already giving this away free as part of their sales
process?" alongside "what's the specific free substitute?" **Kept here only**
for now; fold into method if confirmed on a fourth run.

### Audience-fit gap (Run 3) confirmed by Run 4 — promote to method v2
Running three framings for one domain materially changed the picture
domain-by-domain: each framing failed for a *different* reason. The original
v1 method took the stated audience at face value; v2 should challenge it.
*Suggested for method v2:* a step-1 sub-step asking "is this the right
audience, or would a different audience for the same domain pay more or
sit on a softer free substitute?" Combined with the seller-bundles-support
check above, that's two related v2 improvements. **Two confirming runs is
enough — promoting to a confirmed v2 change list.**

### Strategic implication: the first paid-chat batch is much smaller than the original list
Cumulative picture across runs: websitedesign orchestration-layer is the only
candidate that clears the bar so far (and even that needs a fake-door demand
test). gaswholesalers and robot-hands should both come out of the first
paid-chat batch (pivot to lead-capture / free-with-caps, or shelve). The
original "vast array of domains with paid chat" framing isn't surviving the
method honestly. *Possible reframe to discuss:* paid chat may be the wrong
default monetisation for most of these domains; lead-capture or free-with-
caps may fit more of them. The framework in
`PLAN_simple_paid_multidomain_chat.md` §5 already supports this as a per-domain
mode flag — we may be heading there for most domains. **Kept here only**
pending a discussion.
