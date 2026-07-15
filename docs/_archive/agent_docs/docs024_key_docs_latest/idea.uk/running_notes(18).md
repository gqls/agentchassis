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

### Honest diagnosis: the generation step has been narrow (user critique, accepted)
Four runs of useful filtering, but the candidates generated were pedestrian.
Root causes diagnosed: (a) **supply-side only** — step 2 crossed assets ×
capabilities, never started from user demand or generalist-failure modes;
(b) **generic capability menu** — "text reasoning, vision, voice..." is too
coarse to provoke novel combinations and doesn't reflect what's actually new
in the last 6–12 months; (c) **no explicit "where does the generalist fail
here?" prompt** in generation (only in the cut, by which point it's too late);
(d) **one-shot generation** settles on the obvious. The filter does its job
but the funnel is too narrow at the top. **Kept here** (also reflected in
method v2 patch).

### Method v2 changes — multi-lens generation + richer capability menu
Patched into `idea_uk_method_v0.md`:

- **Capability menu rewritten** by what specialism uniquely does (knowledge &
  retrieval / reasoning & computation / multi-modal in & out / action-taking /
  coordination / personalisation & continuity / quality & safety) with concrete
  examples of what each enables. Replaces the generic v0 list. Currently
  watching: agentic browsing, million-token contexts, reasoning-model pricing,
  real-time voice agents, video generation, precise image editing.
- **Step 2 (generation) made multi-lens.** Four lenses run in addition to the
  v0 asset×capability sweep: a **demand lens** (what does the audience deeply
  want or struggle with), a **generalist-failure lens** (where does general AI
  fail this audience), a **frontier lens** (what just became possible in the
  last 6–12 months), and an **outcome lens** (the dream result, not just the
  help). Aim for 3–6 candidates per lens (12–24 before dedup) rather than
  one pass.
- **Step 1 (audience framing) given an audience-fit challenge** (confirmed from
  the gaswholesalers and robot-hands runs — was kept-here-only, now promoted).
- **Step 3 (cut) given a seller-bundles-support-free check** (confirmed from
  robot-hands and gaswholesalers — same promotion).

### Where specialism wins over generalism (for the generation lens)
Captured concretely so it can be used as a thinking aid: stale/shallow
knowledge → curated grounding wins; confident-wrong on technical specifics →
verification + computation wins; statelessness → persistent memory/case
context wins; description-only → agentic action wins; precision → actual
computation wins; domain multi-modal inputs → tuned vision/voice/sensors win;
real-time/live data → monitoring/alerting wins; niche long-tail expertise →
specialist coverage wins; editorial judgement → curated taste wins;
multi-model orchestration → quality-at-price wins. Each is a generative
prompt the generalist-failure lens can run. **Doc:** method v1 capability
menu + lens b.

### Strategic open thread (not yet resolved)
Even with richer generation, the four runs so far suggest the *visitor-paid*
chat model is the wrong default for most of the operated domains — high-margin
sellers bundle support free; B2B doesn't fit retail price points; hobbyists
won't pay. The richer generation should be tested to see if it surfaces
candidates that survive, or whether the right move is to reframe most domains
as lead-capture / free-with-caps and reserve paid chat for the few where the
asset is exceptional (e.g. websitedesign's orchestration-layer). **Not
resolving here** — waiting on at least one run with the new generation step
before concluding.

### v2 method run across four domains — generation was the missing piece
Ran v2 (multi-lens generation + audience-fit challenge + seller-bundles-free
check + richer capability menu) on websitedesign, gaswholesalers, robot-hands,
and agritec.uk. Result: **13 advancing candidates across four domains**, versus
1 from v0 across three of these. The user's discipline call ("we shouldn't
conclude the model won't work until generation is strong enough") was correct —
generation was the bottleneck, not the model. **Doc:** `idea_uk_testrun_v2.md`.

### Audience-fit challenge does the heaviest lifting
Single biggest unlock: gaswholesalers shifted from "high-paid traders
(overserved by Bloomberg)" to "mid-market procurement managers (underserved)"
and went from 0 advancing to 3–4. Robot-hands also gained a candidate from the
challenge though brand-fit mismatch limits it. **Lesson:** the v0 method's
implicit acceptance of the stated audience was a major weakness. The audience
challenge belongs as step 1, not optional.

### Two recurring specialism wins surfaced as patterns
- **Regulatory specificity wins for non-expert audiences.** websitedesign
  (WCAG/GDPR) and agritec (SFI26/RPA) had their strongest candidates rooted
  in this pattern: generalists are stale and confident-wrong on specifics,
  while curated rule sets + checking are exactly what specialism does well.
  Reusable pattern across domains where regulations exist and the audience
  is non-expert.
- **Agentic action on bad portals wins.** RPA Rural Payments portal, supplier
  RFQ systems, CRO on the user's own site — three different candidates, same
  shape: a generalist can describe what to do, an agent can do it. Worth
  flagging as a pattern for future generation.

### Watchlist should track scheme/event windows, not just AI capabilities
Agritec SFI26 Window 1 (June 2026) creates a uniquely urgent moment that
won't exist in the same form 6 months later. The capability watchlist concept
needs a sister watchlist for **real-world events/windows** per domain. **Kept
here**; potential v3 method addition.

### Payment-model fragmentation — the simple-paid plan may need expansion
The strongest v2 candidates (compliance audit, procurement advisor, SFI
helper) are **B2B SaaS-shaped**: £20–500/month subscriptions or per-job
pricing — not £2 day-pass. The simple-paid-multidomain-chat plan assumed
visitor-paid £2 day-pass as default; the v2 candidates suggest "chat + tool,
B2B subscription" is the model that fits where the asset is strong. The plan's
mode flag already supports per-domain modes, but the *modes themselves* may
need to include B2B SaaS. **Worth a discussion when the user has bandwidth.**
**Kept here for now.**

### Robot-hands.com — candidate looking for a domain
v2 found a good candidate for robot-hands (cross-device funding case-builder)
but the URL doesn't market it. Some domains may be URLs in search of a product
rather than products in search of a URL. **Implication:** treat the asset
collection (what we build/curate) as separate from the domain portfolio (the
URLs we own), and match them later. **Kept here.**

### Outcome lens earned less than expected
On these four domains the outcome lens didn't surface much that the demand
and generalist-failure lenses didn't already catch. Worth watching whether it
earns its place on later runs (different domain types might use it more) before
trimming it. **Kept here.**

### Strongest concrete next move (changed from before): agritec SFI26
The agritec eligibility checker / window alerts is the strongest test-now
candidate across all four domains because of the unique time-window: SFI26
Window 1 opens **June 2026** (within a month of this writing) and the prior
SFI 2024 closed abruptly when funding ran out, so urgency is genuine. Cheap
to build a fake-door, time-sensitive to test. Websitedesign compliance audit
is a strong second. **Doc:** testrun v2 cross-domain summary.

### Pricing settled for the idea.uk product: pay-per-idea, cost-plus
User decision: **pay-per-idea is the primary monetisation for the idea.uk
product itself** (a one-off flat price per report), with **B2B SaaS as a later
mode** after pay-per-idea is validated. This matters because several of the
v2 candidates *for other domains* (compliance audit, procurement advisor,
SFI helper) are B2B SaaS-shaped — but the idea.uk ideation tool itself is
deliberately one-off. Cost-plus pricing from PLAN_idea_uk.md §8; fake-door
sets the headline at £199 per report (room to adjust based on conversion).
**Doc:** PLAN_idea_uk.md §7 and §8 updated.

### Sequence in PLAN_idea_uk.md updated to reflect actual progress
Steps 1–3 marked done (method written, runs done, scoring refined via v1/v2
patches). Step 4 is the **idea.uk fake-door build** (this turn). Step 5 is
**fulfil the first paid reports manually** — both validates demand AND forces
us to ship the method as a real product. Step 6 (agent build) gated on
demand validation + manual delivery becoming a bottleneck. Step 7 (B2B SaaS
as a second mode) deferred.

### Real-world event watchlist promoted to a second standing workflow
Confirmed: alongside the **capability watchlist** (new AI capabilities ⟶
trigger re-runs across domains), a **real-world event watchlist** tracks
scheme deadlines, regulation changes, and application windows per domain
(e.g. SFI26 Window 1 opens June 2026). Both are recurring research workflows
that fire re-runs of ideation. Agritec was the proof — timing turned a
candidate from "consider later" into "test now."
**Doc:** PLAN_idea_uk.md §8.

### Built the idea.uk fake-door page (deployable static HTML)
Deliverable: `idea_uk_fakedoor.html` — single self-contained static page
ready for git → GitHub Actions → B2 deployment. Editorial-refined aesthetic
(Fraunces + IBM Plex Sans, warm cream + ink + rust accent, generous
whitespace, subtle paper grain, staggered fade-in). One SKU at £199/report,
honest "manually delivered during early access" disclosure, 72h refund
guarantee. Three wire-up placeholders the user replaces before going live:
`ORDER_LINK` (Stripe Checkout payment link), `INTEREST_FORM_ACTION`
(form endpoint), `CONTACT_EMAIL`. Includes a specimen report mock so
visitors see the actual report shape, not just claims about it.

The page deliberately states "no candidate advances" is a real outcome — we
return analysis + refund if so. This is a credibility move: it signals the
filter is honest rather than performative, which is the only way the
positioning ("survives scrutiny") doesn't sound like marketing.

### Headline message commits to a position
"Most AI 'opportunities' don't survive scrutiny. We find the ones that do."
This puts the **filter** at the centre of the value proposition rather than
the generation, which matches what we've actually built: the filter is the
mature part (proven through four runs), the generation is the part still
improving. If demand validates, we know they're paying for filter quality
specifically — useful signal.

### Next concrete steps after this turn
1. **User wires up the fake-door** — set Stripe payment link, form endpoint,
   email. Deploy to idea.uk.
2. **Parallel: fake-door for agritec SFI26** — same discipline, different
   product, time-sensitive (June 2026 window). Could reuse most of the page
   structure with different copy and price (likely lower, e.g. £49 for an
   SFI eligibility report + recommended actions, given the audience is small
   farms rather than business operators).
3. **Watch for conversions / interest signals.** Even a handful of paid
   orders validates the £199 price point and the report format.
4. **Fulfil the first orders manually** using the v2 method, by hand, with
   real verification. Each fulfilment is also a v3 stress test of the
   method.

### Fake-door modified to intent-capture-only (no payment yet)
User's call: avoid taking money up-front to limit manual refund work. Updated
`idea_uk_fakedoor.html`:
- Replaced Stripe `ORDER_LINK` CTA with a structured **Request a report** form
  (name, email, business description, audience description, optional notes).
- New flow: visitor submits request → we reply within 24h with confirmed slot
  + Stripe payment link, or polite decline + next-batch offer.
- Visible throttle indicator (`MONTH_SLOTS`) in the header label.
- Disclosure rewritten to reflect "no payment until we confirm."
- Kept lower-commitment email-only capture as a small strip below the main form.
- Phase 2 (when demand validates): swap to direct Stripe Checkout.

The price (£199) stays visible — that's still the willingness-to-pay signal we
want. The change is just *when* money changes hands. **Doc:** PLAN_idea_uk.md §7
step 4 updated; HTML updated.

### Claude-as-preview refinement of the websitedesign candidate (user's aside)
User proposed: "does our site spec/plan help when using Claude's design tool,
and could we offer a flow that lets people test potential sites in Claude
before moving to Lovable?" This is a meaningful refinement, not just a side
thought.

**Why it matters.** v0 candidate 1a (bare prompt for Bolt) failed on
willingness-to-pay because a user about to use Bolt can describe their idea to
Bolt themselves for free — paying for "a better prompt" is a hard sell. But if
the flow is **"pay for the spec, test it for free in Claude before you commit
to Lovable credits,"** the framing changes: the user isn't paying for a prompt
to a paid tool, they're paying for a preview-ready spec that de-risks the paid
step. Same asset (the build pipeline's spec output), different positioning.

A side-by-side demonstration is available: "your idea pasted into Claude" vs
"your idea pasted into Claude via our spec" — a concrete, testable claim the
user can verify *before* paying. That's a real value lift versus v0.

**Sketched score under v2 rubric (refined candidate, call it 1c):**
- Defensibility 3 — same asset, same question of whether our spec is measurably
  better than the user's own description.
- Willingness 4 — improved over 1a because free preview lowers paying risk.
- Buildability 4 — mostly the existing spec generator, wrapped with clear
  "test in Claude" instructions and example formatting.
- Reuse 4 — works for every site-building prospect.
- Durability 3 — depends on whether Claude artifact quality and Bolt/Lovable
  planning both improve faster than our spec stays better than them.
- Sum ~18, advances comfortably. Better than 1b (orchestration layer, sum 17)
  and far cheaper to build.

**Honest caveats:**
- Claude Artifacts has real constraints (single-file React/HTML, specific
  library list, no localStorage, certain output sizes). Best for landing pages
  and single-page sites; multi-page production sites still need Lovable/Bolt.
  The preview is genuinely a preview, not a substitute.
- The flow requires the user to be on Claude.ai (free tier works for casual
  use; heavy iteration may push toward Claude Pro).
- The "demonstrably better" claim has to actually hold — if our spec doesn't
  produce noticeably better artifacts than a casual description, the wedge
  collapses.

**The cheap test (for the user to run, not Claude in this session):**
1. Take a real spec/plan output from the build pipeline (small-business site
   case) and paste it into Claude.ai as: "Build an HTML/CSS landing page based
   on this spec. Use it directly."
2. Take the *same business idea* described casually in one or two sentences
   ("a one-page site for a local accountancy serving sole traders") and paste
   it into Claude.ai the same way.
3. Compare the two artifact outputs side by side. Is the spec-driven one
   visibly better — more on-brand, more conversion-aware, better content
   structure, better typography? If yes, the wedge is real and worth building
   a flow around. If no, the spec isn't differentiated enough yet for this
   use-case and would need work first.

This is a 30-minute test the user can do directly; Claude in this session
can't run it usefully (I'm one model approximating both passes, so the
comparison wouldn't be honest). **Recommended next step on the websitedesign
front, in parallel with the idea.uk fake-door.**

### Capability-menu addition (small but real)
The v2 capability menu didn't include "free-tier preview environments / consumer
chat tools as a testing layer" — which is exactly what the Claude-preview
candidate uses. Worth adding to the menu under **Action-taking** (or its own
category, "Preview & free trial environments") so the generation step picks it
up next time. Concrete examples: Claude Artifacts as a free design preview,
ChatGPT canvas as a free draft environment, v0.dev's free tier, Bolt's free
tier. The pattern is "use a competitor's free tier as the trial layer for our
paid asset" — non-obvious, but the user surfaced it from their own thinking
and the method should be able to find it next time. **Kept here**; fold into
method capability menu when next updated.

### Next concrete steps after this turn
1. **User wires up the fake-door** (REQUEST_FORM_ACTION, UPDATES_FORM_ACTION,
   CONTACT_EMAIL, MONTH_SLOTS), deploys to idea.uk.
2. **Agritec SFI26 fake-door** — same discipline, different product, sharper
   time pressure (June 2026 window opens within weeks). Lower price point
   (£29–£79) for the audience (small UK farmers). Probably the next build.
3. **User runs the 30-minute Claude-preview test** to validate or kill the
   websitedesign refinement.
4. **Method maintenance**: add the preview-environment category to the
   capability menu when we next touch the method.

### First automated run of idea_method_runner.py (agritec.uk) — it works, and outperformed the by-hand run
The script ran end-to-end: 20 candidates generated → 10 survived the cut →
7 premises held on verification → 5 advanced. Two ways it beat by-hand Run 8
on the same domain:
1. **Audience challenge made a sharper call autonomously.** Run 8 carried
   "small farmers"; the script challenged and carried "advisors / land agents /
   agronomists" — with sound reasoning (farmers cash-tight, churn after
   submission, £50-150 one-off at most; advisors bill £400-800/day, run dozens
   of applications, pay £30-100/mo per seat). That's a better audience than I
   picked by hand. The v2 audience-fit step did real work unsupervised.
2. **Verification was more specific** than mine — surfaced the real scheme
   history (SFI22→23→24→26), mid-cycle GRH6 removal, handbook version numbers,
   named real advisors and Land App as a partial competitor, concrete
   action-code conflicts (IPM4 vs CNUM3). Web grounding doing its job.

Top candidate was one I never generated by hand: **a weekly diff/changelog of
Defra/RPA SFI guidance for advisors** (score 18). Leans directly on the
freshness/currency advantage that the whole moat thesis rests on; verification
confirmed no competitor offers it. Genuinely better agritec candidate than
anything in Run 8.

### Bug found and fixed: title-based merge produced an empty candidate
Candidate 5 rendered with blank asset/capability/findings but real scores —
because the pipeline threaded data between steps by matching on the `title`
string, and the scoring model reworded a title with special chars ("SFI × CS ×
BNG"), breaking the match. **Fix:** assign stable `id`s in code after
generation and thread them through cut → verify → score; merge by id, never
title; drop any scored entry whose id doesn't match a verified candidate
(rather than rendering blanks). Prompts updated to require echoing the id.
Recompiles clean. **Doc:** `idea_method_runner.py`.

### The script independently confirmed agritec is B2B SaaS, not consumer chat
Left to challenge the audience freely, it landed on advisors at £30-100/mo per
seat — the same B2B SaaS conclusion the by-hand analysis was circling, and
further evidence the simple-paid-multidomain-chat day-pass model doesn't fit
the strongest agritec framing. Reinforces the open thread: most strong
candidates want B2B SaaS, not £2 day-pass.

### Honest read on the moat, post-run
The script's strongest candidate (diff tracker) scores Defensibility 3,
Durability 4 — an operational/freshness moat (keep the scrape current), not a
structural one. That's consistent with the whole idea.uk thesis: the moat is
effort + freshness + integration, not a static asset. The run is encouraging
but doesn't change that the defensibility is sustained by work, not owned.

### Open: cross-vendor critique still untested
The cut ran on Sonnet while generation ran on Opus (real diversity, same
vendor). The `call_other_vendor` stub for OpenAI/Gemini critique is still
unfilled. Whether genuinely cross-vendor critique changes outcomes remains the
one untested claim in the multi-model part of the moat. **Kept here.**

### Built the idea.uk service layer with billing (idea_service.py)
Wired the engine (idea_method_runner.run) into a working service with two front
doors and one-off billing. Internal: authenticated /internal/run, no payment,
for our own domains. External: the customer flow. Both call the same engine
with assets passed as data (keeps idea.uk sale-ready).

**Flow aligned to the page's request-then-confirm design** (the page had been
edited to "Request a report · no payment yet · we reply within 24h · payment
only on confirmation" — a better early-access flow than pay-immediately):
  request (free, public /request, form-encoded) → operator reviews →
  /confirm (creates Stripe Checkout, emails pay link) OR /decline (polite, no
  charge) → customer pays → /stripe/webhook (verify + idempotent) → fulfil
  (run engine) → AUTO_DELIVER ? email report : hold for operator review.

Why request-then-confirm matters: it lets us screen out businesses where the
method would return "no candidate advances" BEFORE taking money — which
protects the refund promise and is honest. The /decline path emails a polite
"we'd rather not sell you a weak report" message.

**Billing follows the PLAN_stripe_billing_integration.md principles** but in the
lightweight pay-per-idea shape: webhook is source of truth (payment never
trusted from the browser redirect); provider behind an interface (StripeProvider
+ FakeProvider for local testing, swappable); idempotent webhooks (dedup on
event id in sqlite). No subscription, no chassis client_entitlements cache —
this is the one-off variant.

**AUTO_DELIVER defaults OFF** — a paid run is held for operator review before
sending, honouring the "manually delivered during early access" stance. Flip on
once output is trusted. Engine failure flags the order for refund.

Wired the landing-page request form action to /request. Parses clean. Deps:
fastapi uvicorn python-multipart stripe anthropic.

### Topology note: idea.uk is NOT pure-static/edge like the other chat domains
The other simple-paid-chat domains are static-S3 + synchronous edge worker. But
idea.uk's "tool" is a minutes-long multi-LLM + web-search job, not a synchronous
chat turn — so it needs a small always-on backend (the FastAPI service) running
the engine as a background task, plus the static page on S3 posting to it, plus
the Stripe webhook pointed at /stripe/webhook. Worth flagging because it breaks
the pure-serverless model: idea.uk is static-page + small-service, not edge-only.
This is fine (it's the §11-agent-as-a-service stage) but it's a different
deployment shape from the day-pass chat domains. **Kept here.**

### Remaining wiring before idea.uk goes live (handoff list)
1. Deploy idea_service.py (small container/box) with env: ANTHROPIC_API_KEY,
   STRIPE_SECRET_KEY, STRIPE_WEBHOOK_SECRET, PUBLIC_BASE_URL, INTERNAL_API_KEY,
   OPERATOR_EMAIL, SMTP_* (or accept the file-drop fallback during testing).
2. Point Stripe webhook at PUBLIC_BASE_URL/stripe/webhook (checkout.session.completed).
3. Serve the static page; ensure /request POSTs reach the service (same origin
   or CORS/proxy). The page's separate UPDATES_FORM_ACTION (newsletter capture)
   still needs a form endpoint — left as the user's placeholder, separate concern.
4. Decide concurrency/throttle: each fulfil run costs API money + minutes. Add a
   "currently full" state if manual review can't keep the 72h promise at volume.
5. Optionally fill the call_other_vendor stub (cross-vendor critique) — still the
   one untested multi-model claim.

### Built out the gaps + ran the flow end-to-end (20/20 checks pass)
Closed the gaps flagged last turn:

- **Cross-vendor critique implemented** (was the one untested multi-model claim).
  The runner's cut step now routes through OpenAI if OPENAI_API_KEY is set
  (OPENAI_CRITIQUE_MODEL, configurable since vendor model names drift), else
  falls back to a different Anthropic model. So the method can genuinely be one
  vendor generating and a different vendor critiquing — not marking its own work.
  `call_other_vendor` stub is now a real OpenAI Chat Completions call.
- **Capacity throttle** (protects the 72h promise): MAX_ACTIVE_ORDERS caps orders
  in flight (awaiting_payment..awaiting_review). `/confirm` returns 409 at_capacity
  when full; `/capacity` is public so the page can show "currently full".
- **Newsletter /subscribe** endpoint + subscribers table; page updates form wired
  to it.
- **CORS** middleware (ALLOWED_ORIGINS) so the S3 page can call the service.
- **Deployment artifacts**: Dockerfile, .env.example, RUNBOOK_idea_uk.md.

**Validated by actually running it.** Wrote test_idea_flow.py (FastAPI TestClient
+ FakeProvider + stubbed engine — no Stripe, no LLM spend) and ran it:
**ALL 20 CHECKS PASSED**. Covers: request→confirm→pay(webhook)→fulfil→deliver,
operator-key auth gates, customer pay-link email, webhook idempotency (duplicate
ignored), decline path, capacity gate (3rd confirm blocked at MAX=2), internal
run, subscribe. The state machine is proven, not just parsed.

This is the first time in the project something has been validated by execution
rather than by-hand reasoning or syntax check. The flow logic holds.

### idea.uk is now a runnable internal + external tool with billing
Internal: /internal/run (auth, no billing) for our own domains — same engine.
External: request-then-confirm + Stripe one-off, operator review gate.
Remaining before live is config/deploy (Stripe keys + webhook, deploy container,
serve page, set capacity), not code — captured in RUNBOOK_idea_uk.md go-live
checklist. The cross-vendor critique question (does a different vendor's cut
change outcomes) is now testable for real by setting OPENAI_API_KEY and
re-running a domain — still worth doing as the moat-validation step.

### Ported the idea.uk tooling from Python to Go (platform is Go throughout)
Rewrote the engine + service in idiomatic Go, stdlib-only so it builds offline
with no module fetch. Installed Go 1.22 (apt). New module in `idea-go/`:
- engine.go + prompts.go — the method pipeline; calls Anthropic/OpenAI directly
  over net/http (no SDKs). Cross-vendor cut preserved (OpenAI if OPENAI_API_KEY,
  else a second Anthropic model). Stable ids threaded by id not title (the bug
  fix carried over). EngineFunc is a swappable field for tests.
- store.go — Order + JSON-file Store behind a small method set (Save/Get/Update/
  ActiveCount/MarkEventSeen/AddSubscriber), mutex-guarded, atomic-ish file
  replace. Production swaps in chassis Postgres behind the same methods.
- billing.go — Provider interface; StripeProvider (checkout via form-POST to
  api.stripe.com; webhook verified with crypto/hmac + sha256, no SDK);
  FakeProvider for local testing.
- service.go — App struct; request-then-confirm flow, idempotent webhook,
  capacity throttle, /internal/run, /subscribe, CORS; `dispatch` field so
  background fulfilment runs inline in tests, goroutine in prod.
- main.go — config from env, serve; `idea internal D A S` CLI runs the engine
  with no server/billing.
- service_test.go — Go port of the flow test.

**Built, vetted, and tested offline: go vet clean, go build OK, go test PASS
(19 checks)** — same flow the Python test covered: request→confirm→pay→fulfil→
deliver, auth gates, webhook idempotency, decline, capacity gate (3rd confirm
blocked at MaxActive=2), internal path, subscribe. Validated by execution, like
the Python version.

Updated Dockerfile (Go multi-stage, distroless) and RUNBOOK to Go build/test/run
commands. The Python files remain as the reference implementation but Go is now
the canonical version, consistent with the rest of the platform.

Design choices worth noting:
- stdlib-only on purpose: no external Go deps to fetch/vet/audit, and it builds
  in a locked-down environment. The cost is hand-rolled Stripe calls + webhook
  HMAC, which are small and well-understood.
- Store behind an interface-shaped method set so the JSON file (standalone/MVP)
  swaps cleanly for the chassis Postgres (production/integration) — matches the
  separability/sale-readiness line in PLAN_idea_uk.md §2.
- App struct with injected engine/deliver/dispatch/provider = testable without
  real keys, real Stripe, real LLM calls, or real time delays.

### Wrote the architecture & deployment guide; clarified hosting + OpenAI
Created `idea_uk_architecture_and_deployment.md` — plain-language map of the
pieces, the diagram, hosting reality, Stripe flow, deploy checklist, applying to
other domains, chassis integration, and how to run the engine vs the test.

Grounded the chassis-integration and site-spec sections in the REAL registry
(read registry_go.txt, 020_tool_lifecycle.md, 021_site_spec_and_classifier.md):
- The method maps almost entirely onto EXISTING actions: execute_llm_prompt
  (generate/cut/verify/score), web_search/scrape_web/firecrawl_* (verify),
  request_human_input/create_approval_request/await_approval/process_approval_decision
  (the operator confirm+review gate is literally HITL), send_notification,
  store_result/write_my_state, read_site_spec/write_site_spec. So the chassis
  version is one idea-orchestrator agent + one workflow reusing these, NOT a
  port of engine.go. Did NOT write the SQL — needs a schema pass first
  (check-schema-before-SQL).
- Honest distinction surfaced: existing "tools" (deploy_tool_to_site) are
  self-contained client-side HTML/JS widgets forked into static sites. The
  ideation engine is server-side, minutes-long, paid — it CANNOT be a forked
  content_components tool. So "apply to a domain" = either Shape A (the site IS
  the service, like idea.uk) or Shape B (a static "request a report" page that
  posts to the ONE central service). Many pages, one engine.
- site_specs mechanism already fits: an ideation feature is a site_plan item
  with status blocked→planned→built; feasibility-recheck promotes it when the
  agent exists.

Clarified the "serverless" terminology for the user: the PAGE is serverless
(static on B2), the SERVICE is NOT and can't be (minutes-long multi-LLM job +
stable webhook endpoint) — it's a small always-on container. This was a
recurring point of confusion.

### Added a [cut] vendor log line to engine.go (resolves OpenAI confusion)
The user wasn't sure whether their run used OpenAI. The cut step uses OpenAI
only if OPENAI_API_KEY is set in the env, else Claude Sonnet. Added a stderr
line: "[cut] cross-vendor: OpenAI (gpt-4o)" or "[cut] same-vendor: Anthropic
(claude-sonnet-4-6)" so every run states which vendor critiqued. Rebuilds clean.
Explained in the guide: echo $OPENAI_API_KEY to check; the TEST never calls
OpenAI/Anthropic (stubbed engine), only `go run . internal` / real orders do.

### The Go engine produced a real agritec report (live APIs)
User's `go run . internal agritec.uk ...` returned a full report: advisor
audience again, top candidates = Client Case File Memory, Scheme Change Diff
Alerts (test_now), BNG/biodiversity cross-check, multi-client portfolio,
stacking checker, tenancy interaction, inspection pre-mortem. Consistent with
the Python automated run and sharper than by-hand Run 8. Confirms the Go port
works end-to-end against real Anthropic (and OpenAI for the cut if key set).

## CHECKPOINT 2026-05-28 — paused to discuss pricing, real-door, self-host, white-label
Stopping here to think before pushing further. Five open questions captured
fully in **idea_uk_open_discussion.md** (the canonical pickup point next session).
Short version of each, with the working answers:

### 1. Per-run cost
Verified Anthropic pricing (May 2026): Opus 4.7 $5/$25, Sonnet 4.6 $3/$15, Haiku
4.5 $1/$5 per MTok. Pipeline = 5 LLM calls; verify step (Opus + web_search,
~20-25k input tokens from search results) dominates. **~£0.30–0.60 per run,
£0.50 working estimate.** Optimisable to £0.20–0.30 with Haiku for scoring +
prompt caching on the static prompt parts. Real cost will appear in Anthropic
console after runs.

### 2. Stripe break-even + recommended pricing
UK Stripe: 1.5% + £0.20 (UK), 3.25% + £0.20 (international). **Stripe does NOT
return the processing fee on refunds.** Refund cost = £0.93 fee + £0.50 engine
= ~£1.43 per refunded £49 order. Break-even charge is ~£0.72. **Recommended
early-access price: £29-£49** (not £199), strong refund guarantee retained.
Move up after 5-10 unrefunded orders.

### 3. Real door — instant result on payment?
Current flow: pay → webhook → engine (2-10min) → operator review (AUTO_DELIVER
off) → email. NOT instant. Three options analysed; the realistic best is a
**streaming-progress page after Stripe redirect** that polls /status/{order_id}
and shows "generating… cutting… verifying claim 1 of 8… done" with the report
rendering in-browser. Modest extension (add status field + polling endpoint +
post-pay page). **Want to build this when you decide on price.**

### 4. Voluntary pay / two free goes — honest verdict
**Not recommended in that form.** Voluntary pay converts 1-10% B2B, doesn't
filter serious users, attracts abuse, no demand signal. Two-free-goes is
trivially circumventable by new email. **Better pattern: free "audience
challenge" taster (~£0.02 to run, often the most valuable line in the report)
+ £29 paid full report with refund guarantee.** Gets the same hook benefit
without the abuse risk.

### 5. Self-hosted LLMs — the honest realities
**Don't self-host for idea.uk now.** At early-access volume, commercial models
are dramatically cheaper than self-hosted. Llama 3.3 70B Q4 needs ~40GB VRAM
(£25k H100 or £5k used A6000); cloud H100 rental ≈ same cost per run as
commercial API but with lower cut-step quality (open models tend to agree with
the generator — exactly the failure the cut exists to prevent). Self-hosting
starts to make sense at 100s of runs/day with GPU ops capability. **2027
decision, not 2026.** Cheap wins now: Haiku for scoring (5× cheaper than
Sonnet), prompt caching (90% off cached input on the static prompt parts).

### 6. White-label / branded URLs from other sites
**Option C recommended:** branded request page per tenant on each site (built
through normal pipeline, own copy/brand/price), POSTs to central idea.uk
service with tenant_id; one central engine; Stripe Checkout supports merchant
branding. Per-tenant pricing/tracking; cheap to add tenants. ~100-200 lines of
Go to add tenant support to the service. Stripe Checkout URL is unavoidably
checkout.stripe.com unless we build embedded checkout. Iframe and subdomain-
proxy options also documented but C is the right answer for "offer something,
charge not too much, get feedback, don't lose money."

### Charging + refunds — operational notes captured
Stripe account setup is ~30 minutes once you have UK company/sole-trader
details + bank account; test cards before live. Manual refund = one click in
dashboard. Programmatic refund = ~30 lines of Go to add /refund endpoint
operator-gated. **Decision pending.**

### Open decisions for next session (in idea_uk_open_discussion.md §7)
1. Price (£29 vs £49 vs free-taster+£29)
2. Real-door streaming page (build or defer)
3. Programmatic refund endpoint (add or use dashboard)
4. Multi-tenant Option C (build or single-site for now)
5. Cost optimisation (Haiku for scoring + caching) — likely worth doing
6. Self-host (parked until 100+ runs/week)
7. Stripe account setup (test then live)

### A small engine improvement landed this turn
Added a stderr log line: `[cut] cross-vendor: OpenAI (gpt-4o)` or
`[cut] same-vendor: Anthropic (claude-sonnet-4-6)` so every run states which
vendor critiqued. Resolves the previous ambiguity ("did my run use OpenAI?").
Rebuilds clean.

## CHECKPOINT 2026-05-28 (continued) — decisions confirmed, new concerns surfaced

### Decisions locked
- **idea.uk pricing**: £29 full report, with a **free 30-second audience check** as the
  hook. Not voluntary pay, not multi-free-goes — a single instant taster (Step 1
  only, ~£0.02 cost, no web search) so visitors see real value before paying.
- **First vertical tool**: **SFI26 single-farm assessment** (£49–99 one-off,
  not subscription yet). Farmer/advisor feeds in farm details → gets back
  Window recommendation, eligible action stack within £100k cap, what to avoid.
  Closest thing to "buy a consultant for an hour, in 5 minutes." Subscription
  layered on top later once engine + corpus are validated.
- **Real-door delivery for idea.uk**: build the streaming-progress page after
  Stripe redirect. Customer sees "generating candidates… cutting… verifying
  claim 1 of 8… done" with the report rendering in-browser. ~5–10 min on page.
- **Quality direction**: use the best models and latest features. Don't
  cost-optimise yet. Extended thinking on the cut step, the newer
  `web_search_20260209` with code execution on verify, prompt caching for
  context-packing (not just cost), file attachments for whole-PDF reading.
- **Vendor swap-ability**: protected by step-as-function discipline, not by
  avoiding vendor features. Each step (generate/cut/verify/score) has one input
  and one output shape; swapping vendors = replacing one function's body, not
  rewriting the workflow. The chassis is already built this way.

### Disambiguations and corrections
- **"Manual fulfilment" had two meanings; I used both ambiguously.** The
  intended meaning is *engine runs the report automatically, operator reviews
  the draft before sending*. That's what `AUTO_DELIVER=false` does. NOT "person
  writes the report by hand." Should always say "operator review" or
  "auto-deliver" going forward.
- **"Fake door" is no longer accurate for idea.uk.** Engine produces real
  reports, service is built and tested, billing is wired. It's an MVP that
  isn't deployed yet — that's a different thing. The other candidate tools
  (SFI26, websitedesign compliance, etc.) genuinely don't exist yet — those
  are the real build items.
- **Real per-run cost (from user's actual data)**: ~$1.30 Claude + $0.02 OpenAI
  ≈ **£1.05 per run**, not the £0.50 I estimated. Reasons: Opus 4.7 tokenizer
  generates ~35% more tokens than 4.6 for the same text (Anthropic flagged at
  launch); verify step's web search pulls more result text into context than
  budgeted. At a £29 charge there's still healthy margin; just don't run at £5.

### Audience-check taster — exact UX
Two input fields: business (domain or one-line description) + stated audience.
One Opus 4.7 call, ~10 sec response. Shows: the reframed audience with reasoning,
three alternative audiences with one-line reasons each, and a CTA to buy the
£29 full report. ~1,200 input + 500 output tokens = ~£0.02. At 1,000 free
tasters/month = £20 — easily covered by a 1–2% conversion to £29.

### New concerns raised (this checkpoint)
1. **Liability**: bad info from us could cause a farmer to lose grant money,
   miss a window, commit to a wrong agreement, or pass on a better option.
   Direct financial consequence. Need T&Cs, disclaimers, insurance, and
   technical mitigations (citations, "verify before acting" framing).
   **Doc:** LIABILITY_AND_TERMS.md.
2. **Page wording too clever**: the editorial framing ("Most AI 'opportunities'
   don't survive scrutiny. We find the ones that do.") is intellectual; people
   scan, they don't read. Simpler English needed. **Doc:** rewritten
   idea_uk_fakedoor.html (design kept, words simpler).

### Build sequence agreed
Captured in **DEVELOPMENT_RUNBOOK.md** as Phases A → D.


### Files added this checkpoint
- **LIABILITY_AND_TERMS.md** — risk analysis (idea.uk vs SFI tool sharper),
  technical/operational/legal mitigations, draft starter T&Cs for both
  products, insurance notes, complaint sequence. Top of doc carries the
  "not a lawyer" caveat; ~£200–500 fixed-fee UK solicitor review needed
  before going live. Important real-regulation finding: SFI navigation
  isn't formally regulated (no FCA/SRA/BASIS/CAAV requirement), but the
  *negligent-misstatement* common-law route (Hedley Byrne) is the real
  liability exposure — effective, conspicuous, proximate disclaimers in
  the report itself are the practical answer.
- **DEVELOPMENT_RUNBOOK.md** — Phase A (pre-launch quality + liability
  hardening for idea.uk, 8 tasks), Phase B (deploy idea.uk + first 10
  operator-reviewed orders), Phase C (SFI26 single-farm assessment tool,
  5 tasks), Phase D (chassis-native version, deferred). Each task has
  output + acceptance + unblocks. Right-now shortlist: A1 (engine upgrade),
  A2 (taster endpoint), A5 (page rewrite), A6 (solicitor review kickoff),
  A7 (PII quote kickoff).
- **idea_uk_fakedoor.html updated** — plainer English throughout; new
  free "Try free first" section with working audience-check taster widget
  (posts to /audience-check, ~10s spinner, inline result with CTA);
  honest "what this is and isn't" paragraph integrated into the disclosure
  section (informational service not professional advice; person reviews
  every report before sending; sources cited and checkable); £29 not £199
  everywhere; T&Cs acceptance checkbox on the request form; T&Cs and
  refund-policy links in the footer; corrected the inaccurate "produced
  manually by a person" line to "produced by our analysis engine and
  reviewed by a human." Editorial design unchanged (Fraunces + IBM Plex
  Sans, warm cream + rust accent).

### Honest correction
The disclosure section previously said "produced manually by a person
following the same method" — that was incorrect. The engine produces the
report automatically; an operator reviews the draft before sending
(AUTO_DELIVER=false). Fixed.


## CHECKPOINT 2026-05-28 (continued — Risk column added, SFI single-farm paused)

### What we caught
The v2 method scored agritec/SFI single-farm assessment as **sum 17,
test-now** — a confident "build this" recommendation. A wrong action stack in
that £49 report could have cost a farmer £5k–50k of lost grant money. **The
rubric had no dimension for the consequence of being wrong.** It was caught on
operator instinct, which doesn't scale.

### What we changed
Added a sixth scoring factor — **Risk to the operator** (1–5, 5 = safest,
scores *consequence* not probability). The rubric:

| Risk | What it means |
|---|---|
| 5 | pure analysis, customer decides, no plausible loss beyond fee paid |
| 4 | minor consequences, refunds make customers whole |
| 3 | meaningful decisions, customer can verify citations, PII recommended |
| 2 | high-stakes / regulated-adjacent, needs review + PII + reviewed T&Cs |
| 1 | regulated profession (medical/legal/FCA) — do not build without qualifications |

Rules:
- **Risk NOT in the sum.** Sum is fitness (Def+Will+Build+Reuse+Dur, /25);
  Risk is hazard. Kept separate, reported separately.
- **Risk = 1 dropped automatically**, surfaced in a separate "Dropped for
  operator risk" section so the operator sees what was killed for risk vs
  what failed the Def/Will gate.
- **Risk ≤ 2 still advances** if it passes the gate, but flagged
  "⚠ needs liability work before building"; cheapest_test for these must
  require "validate demand first; PII + T&Cs reviewed before any build."
- **Risk as tiebreaker** in the rank (prefer safer at equal fitness).
- Gate unchanged: Def ≥ 3 AND Will ≥ 3 (fitness gate only).

### Files changed this checkpoint
1. **idea-go/engine.go** — Risk + NeedsLiabilityWork in `scored` struct; score
   step drops Risk=1 + marks needs_liability_work; sort by Sum desc, Risk desc
   tiebreaker; `render` takes new `riskDropped` arg + shows "Operator risk:
   N/5 (label)" + new "Dropped for operator risk" section; `riskNote()` helper.
   All three render call sites updated. **Built + vetted + tested clean.**
2. **idea-go/prompts.go** — `scorePrompt` rewritten with Risk rubric, "asymmetry
   test", not-added-to-sum rule, cheapest_test framing for Risk ≤ 2, new JSON
   field `"risk": n`.
3. **idea_method_runner.py** — Python parity: `SCORE_PROMPT` mirrors Go; `run()`
   drops Risk=1, marks needs_liability_work, sorts by (sum, risk) desc;
   `render()` shows Operator risk + Dropped-for-risk section; `_risk_note()`
   helper. Compile-clean.
4. **idea_uk_method_v0.md** (v3) — Risk added to rubric with full 1–5 ladder;
   new "Risk rules" section; output template updated.
5. **idea_method_prompt.md** — single-shot STEP 5 + STEP 6 updated for Risk.
6. **DEVELOPMENT_RUNBOOK.md** — **Phase C swapped**: SFI single-farm
   assessment paused; **SFI26 Diff Alerts** substituted (re-scored under new
   rubric: Def 4 / Will 4 / Build 4 / Reuse 3 / Dur 4 / **Risk 4**, sum 19,
   test-now). C1–C5 rewritten for the diff product (corpus, weekly diff
   engine, first 8 weekly digests, etc.). Single-farm goes to backlog,
   revisit after Diff Alerts proves operational credibility + PII in force +
   named UK agricultural advisor on retainer.
7. **LIABILITY_AND_TERMS.md** — added "first-line filter" section explaining
   the Risk column as the method's automatic check before this doc gets used;
   notes single-farm SFI paused on these grounds and Diff Alerts replacing it.
8. **016_debugging_guide_v2_30.md** (NEW version) — Section 0 item **23**
   added in the established bold-principle / concrete-instance / disciplines
   style. Cross-references the implementation files and the runbook swap. The
   wider lesson: when a scoring system recommends actions for an operator who
   carries downstream exposure, score the operator's risk as a SEPARATE
   dimension; fitness scores cannot capture hazard.

### What's still open from the prior checkpoint
- Phase A1–A8 tasks (engine upgrade to latest LLM features, /audience-check
  endpoint, streaming progress page, refund endpoint, T&Cs solicitor review,
  PII quote, Stripe live mode) — all unchanged by the Risk-column work.
- Right-now shortlist still: A1 + A2 together is the next coherent build
  session.

### Open question (not blocking)
Cross-vendor critique (OpenAI cut step) vs same-vendor (Sonnet cut) — never
A/B tested yet. Cheap to test once we have a stable agritec run to compare
against: run with `OPENAI_API_KEY` set and without, see whether verdicts
change. Worth doing before the A1 engine upgrade so we have a clean baseline.


## CHECKPOINT 2026-05-28 (continued) — A1 + A2 done in one session

### A1 engine upgrade — landed
Verified current Anthropic features by web search (not memory) before coding:
- Extended thinking shape: `thinking: {type:enabled, budget_tokens:N, display:omitted}`
  with N ≥ 1024 and < max_tokens. Billed at output token rates.
- web_search v2 is `web_search_20260209`, paired with `code_execution_20250522`
  for dynamic filtering. ~11% accuracy lift on BrowseComp / DeepsearchQA,
  ~24% fewer input tokens (Anthropic's own benchmarks). Code execution free
  when paired with web_search. Server-side at $10/1k searches.
- Opus 4.8 is current frontier (released a few days ago; "no breaking changes
  from 4.7"). Sonnet 4.6 stays the second tier.
- Prompt caching: 90% off cached input, 5-min TTL by default. Wrap system as
  a content block with `cache_control: {type:ephemeral}`.

Engine changes:
- `callClaude` → `callClaudeOpts(callOpts)`: options-struct call helper with
  optional `ThinkBudget` and `CacheSystem`. Original `callClaude` signature
  preserved as a thin wrapper so all existing call sites unchanged.
- Default models: **Opus 4.8** on generate + verify; Sonnet 4.6 on cut + score
  (cross-model diversity preserved). All overridable via env vars.
- **Cut step**: extended thinking on the Anthropic branch (budget 4000),
  caching on. OpenAI branch unchanged (no Anthropic-side thinking applies).
- **Verify step**: upgraded to `web_search_20260209` + `code_execution_20250522`.
  Extended thinking (budget 8000), max_tokens raised to 16000 for headroom.
- **Score step**: extended thinking (budget 2000), caching on. Carefully
  applying the new Risk column rewards a little deliberation.
- **Audience + generate**: thinking deliberately off (breadth, not depth);
  caching on. Audience step extracted into `runAudience` for reuse by /audience-check.
- Stderr `[cache]` log line on every call that shows create/read counts so the
  operator can see caching actually working.

### A2 audience-check endpoint — landed
New file `idea-go/audience_check.go`:
- POST /audience-check, public, no auth, no billing.
- Form fields: `business` + `audience`. Length-capped at 200 chars each.
- Calls `a.audience(business, audience, "")` — `runAudience` injected on App
  so tests can stub.
- HTML response uses the page's existing `.taster-result` CSS classes; all
  user-supplied input HTML-escaped (XSS-safe).
- Per-IP sliding-window rate limiter: 3/hour AND 20/day (both must clear).
  In-memory; resets on restart (acceptable for this risk profile).
- Kill switch: `TASTER_ENABLED=false` returns 503 with a polite message.
- Rate-limit response includes Retry-After header and "have another go in N
  minute(s)" body.

Tests added (in service_test.go):
- GET rejected with 405.
- Missing either field rejected with 400.
- Happy path asserts response shape + CTA + £29 line.
- XSS escaping explicitly asserted (`<script>` and `<b>` both escaped).
- Rate-limit cap-and-window asserted (4th call in same hour blocked).

All 19 original flow tests still pass; 5 new tests pass. **24/24 total.**

### What this means in practice
- Engine is now using the best Anthropic features currently available.
- The page's taster widget now has a live endpoint to call.
- The Risk column added earlier this session will appear in any new run's
  output, scored more carefully thanks to thinking on the score step.
- Cost per run will go up a bit (extended thinking is billed at output rate;
  more max_tokens headroom), but user explicitly said don't cost-optimise yet.
  Probably £1.50–2.50/run now vs the £1.05 we measured before. Still healthy
  margin at £29.

### Open validation step (not blocking, costs ~£1.50)
Run agritec end-to-end once with the new engine to confirm the upgrade works
against real APIs:

```
cd idea-go
GOPROXY=off GOTOOLCHAIN=local go run . internal "agritec.uk" "UK small farmers" "curate scheme docs"
```

Watch stderr for:
- `[cut] same-vendor: Anthropic (claude-sonnet-4-6) with extended thinking`
- `[cache] claude-opus-4-8: created=N read=M ...` on steps 2–5
- Output report includes an "Operator risk: N/5" line per advancing candidate

### Next up — runbook right-now list
- **A3** streaming progress page (real-door UX)
- **A4** programmatic refund endpoint
- **A5** verify the page's taster widget wires cleanly to the live endpoint
  (page rewrite already done earlier; just confirm wiring on deploy)
- **A6** solicitor T&Cs review (background)
- **A7** PII quote (background)

## CHECKPOINT 2026-06-04 — validation run surfaced two API bugs, both fixed

Running the upgraded engine end-to-end (agritec.uk) against real APIs to validate
A1+A2 surfaced two real bugs, fixed in sequence. The cut step worked first try
(extended thinking confirmed); the verify step failed twice before going clean.

### Bug 1 — duplicate auto-injected tool (web_search v2)
`verify step: anthropic 400: "Auto-injecting tools would conflict with existing
tool names: ['code_execution']. Each tool name must be unique."`
- Cause: `web_search_20260209` injects its OWN `code_execution` tool server-side
  to do dynamic filtering. We were ALSO declaring `code_execution_20250522` in
  the same tools array → duplicate name → 400.
- My misread: the docs say v2 search "requires code execution enabled," which I
  took as "add it yourself." It enables it for you.
- Fix: pass only `web_search_20260209`; removed the explicit code_execution.
  Filtering still runs, still free.

### Bug 2 — thinking API differs by model family
`verify step: anthropic 400: "thinking.type.enabled" is not supported for this
model. Use "thinking.type.adaptive" and "output_config.effort"`
- Cause: Opus 4.7/4.8 (and Mythos) dropped MANUAL thinking budgets
  (`thinking:{type:enabled,budget_tokens:N}`) for ADAPTIVE thinking
  (`thinking:{type:adaptive}` + `output_config:{effort}`). Manual form 400s on
  those models. Sonnet 4.6 STILL takes the manual form — which is why the cut
  step (Sonnet) succeeded on the same code path the verify step (Opus 4.8)
  rejected in the same run.
- Fix: `callOpts.ThinkBudget int` → `callOpts.Effort string`. New
  `usesAdaptiveThinking(model)` predicate (true for opus-4-7/opus-4-8/mythos):
  emits adaptive+effort; else emits manual budget (via `effortToBudget`,
  clamped < max_tokens). One helper, two wire formats, chosen by model string.
- Call sites now: cut → Effort "high" (Sonnet manual budget 8000), MaxTokens
  12000; verify → Effort "xhigh" (Opus adaptive — Anthropic recommends xhigh for
  agentic+search), MaxTokens 32000; score → Effort "medium" (Sonnet manual 4000),
  MaxTokens 10000; audience + generate → no thinking (Effort "").
- Build + vet + test clean (24/24). No stray ThinkBudget refs.

### Key facts learned (verified via web, not memory)
- Opus 4.8 released 2026-05-28; current frontier; 1M context, 128k output;
  adaptive thinking only; manual budgets 400; temperature/top_p/top_k 400 if
  non-default (carry-over from 4.7).
- effort levels: low | medium | high | xhigh | max. Default high. Anthropic:
  start xhigh for coding/agentic, high floor for reasoning-heavy, step down only
  after evals. For xhigh/max set large max_tokens (they suggest 64k+ for heavy
  coding; we use 32k for verify).
- web_search_20260209 + auto code_execution = dynamic filtering: ~11% accuracy
  lift, ~24% fewer input tokens (Anthropic benchmarks). $10/1k searches.

### Debugging guide
Added items 24 (duplicate auto-injected tool) and 25 (model-family thinking API
divergence) to Section 0. Written as `016_debugging_guide_v2_31.md`.

### DIVERGENCE FLAG — debugging guide item 23
The user's uploaded v2_30 has item 23 = a gpu-provisioner `output_fields`
(plural) bug dated 2026-06-03 — NOT the Risk-column lesson I drafted as item 23
last session. So the user's canonical guide evolved separately after our last
session; my Risk-column write-up is NOT in this lineage. Did NOT silently
re-insert it. **Open question for the user: do they want the Risk-column lesson
folded into the guide too (it would become item 26), or is it tracked
elsewhere?** The Risk column itself IS in the method + engine; only the
debugging-guide writeup of it is absent from their current guide.

### Cost note
Verify at xhigh effort + Opus 4.8 + web search will be slower and pricier than
before — expect maybe £2–4 for this validation run. User said use the best /
don't cost-optimise yet, so this is expected, not a regression.

### Status
Engine fix in place; awaiting the user's next live run to confirm clean
end-to-end. If `claude-opus-4-8` 404s on their account, one-line fallback:
`export GEN_MODEL=claude-opus-4-7 VERIFY_MODEL=claude-opus-4-7` (4.7 is also
adaptive-thinking, so the new code path handles it identically).
