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

## CHECKPOINT 2026-06-04 (continued) — timeout bug fixed; consolidation map written

### Bug 3 — verify step client timeout
`verify step: Post ".../v1/messages": context deadline exceeded (Client.Timeout
exceeded while awaiting headers)`. Different class from bugs 1+2 (those were fast
400s; this one the request was ACCEPTED and the server held the connection while
thinking+searching). Cause: Opus 4.8 at xhigh effort + 6 web searches runs for
minutes; shared http.Client timeout was 180s. Fix: raised to 900s. Noted
streaming as the durable long-term answer (keeps connection alive, dodges proxy
timeouts) but deferred — bigger change. Build/test clean.

### Debugging guide
Built **016_debugging_guide_v2_32.md** on the user's latest upload v2_30b
(items 1-23, where 23 = the Risk-column lesson). Added:
- 24: hosted tool auto-injects a helper tool → duplicate-name 400 (web_search v2
  injects its own code_execution).
- 25: thinking API differs by model family (Opus 4.7/4.8 adaptive+effort vs
  Sonnet 4.6 manual budget).
- 26: long agentic call holds connection for minutes → client timeout must
  allow worst-case; streaming is the durable fix.
- 27: STANDING DISCIPLINE (user explicitly elevated this) — confirm the exact
  current request shape of any model/tool API from live docs before coding
  against it; shapes drift between generations, a remembered shape is a guess.
  Cites 24+25 as worked examples.

LINEAGE NOTE: v2_30b has the Risk-column item 23, NOT the gpu-provisioner
output_fields item that was in the *other* v2_30 upload. v2_32 is built on
v2_30b and does not contain that gpu-provisioner item. If the user wants both,
they need to reconcile their branches — flagged to them.

### Consolidation map written
User asked to step back and see how everything fits. Wrote
**CONSOLIDATION_where_it_all_fits.md** — five layers:
- L0 chassis (exists): builds static sites, has all the reusable actions.
- L1 idea engine (built, standalone): the method, internal CLI + idea.uk.
- L2 idea.uk product (in progress): £29 + taster + real-door; FIRST to go live.
- L3 vertical tools (in progress): turn recommendations into products; SFI26
  Diff Alerts first; CHASSIS-NATIVE (vs idea.uk standalone).
- L4 tool-rich site building for any domain (future): the ORIGINAL problem
  statement; idea engine becomes a PLANNING INPUT to the chassis site builder.
- L5 automated VM backend deployment (future, the gap the user named): today
  deploys static→B2; can't yet provision+deploy persistent backend services;
  THUNDER ADAPTER is the seed of this layer.
Key framings captured: two kinds of "tool" (static client-side widget vs backend
service) and two kinds of "deploy" (static→B2 done, backend→VM is the gap);
idea.uk standalone vs vertical-tools chassis-native and why; the natural order
is the order we're already in (prove L1 → ship L2 → build L3 once → generalise
into L4 → grow L5 from Thunder).

### Status / next
Engine timeout fixed; awaiting the user's next live run to confirm clean
end-to-end (and to finally see the [cache] lines + Operator risk line). xhigh
verify will be slow (minutes) and ~£2-4. If too slow, dial verify xhigh→high.
Open: whether to add the gpu-provisioner item to the guide lineage (user's call).

## CHECKPOINT 2026-06-04 (continued) — LAYER 1 VALIDATED; search budget tuned; Layer 5 reassessed

### Layer 1 proven end-to-end (the validation gate)
The agritec.uk live run succeeded with the upgraded engine. Confirmed:
- Caching working: `[cache] claude-opus-4-8: created=13798 read=77940` — ~78k
  tokens served from cache (bills ~10% of input), system prompt reused across steps.
- Verify ran on Opus 4.8 at xhigh with the web-search loop (took a few minutes —
  the reason the 180s timeout had been failing).
- Risk column live and DISCRIMINATING: 2 candidates (Rejected-Claim Diagnoser,
  Compliance Evidence Readiness Audit) came back Risk 2 with ⚠ needs-liability-work
  AND their cheapest_test auto-rewritten to "validate demand first; PII + T&Cs
  reviewed before building" + a per-candidate rationale. The 2 test_now candidates
  came back Risk 3 with normal framing. Nothing hit Risk 1 so no dropped-for-risk
  section (correct). The model reasoned consequence-of-being-wrong on its own.
- Method consistency: landed on the advisor/larger-farm audience AGAIN and put
  guidance-change monitoring as #1 test_now — i.e. independently re-derived the
  SFI Diff Alerts product we'd already chosen for Phase C. Good signal before
  trusting it to PLAN sites (Layer 4).

### Search budget tuned
The run reported "search quota exhausted" (6 searches, 4 candidates) → several
premises came back "provisional." Made it configurable: `WEB_SEARCH_MAX_USES`
env, default 12 (was hard-coded 6). Added `envInt()` helper + strconv import.
Build/vet/test clean.

### PARALLEL THREAD — Layer 5 reassessed against the actual repo
User pushed back on "Layer 5 = far future." Read the repo; they're right.
Written up in **PARALLEL_engine_deployment_and_layer5.md**. Findings:

WHAT EXISTS (deployed):
- Thunder adapter (cmd/thunder-adapter) — PHASE 4 COMPLETE & DEPLOYED (verified
  prod 2026-05-24): provision VM, ssh_exec (verified exit 0 real A100),
  ssh_get_status, prepare_dataset_url/prepare_artefact_url (presigned B2 file
  transfer), decommission. All callable from workflows.
- model-trainer agent = working orchestration template: spawn data-preparer →
  provision (gpu-provisioner) → launch over SSH (training-launcher via ssh_exec).
- 007_adoption_pipeline_v4.md = the "former nginx/security/logging box recipe"
  the user remembered: OVH VM (Terraform pattern) + nginx + certbot + fail2ban +
  systemd + Go binary + Prometheus/Grafana; PLUS config-driven site_api_routes
  in site_specs (routes in DB, no code deploy).
- 032 storage + B2 presigned-URL plumbing for shipping the binary.

THE REAL GAP (not the plumbing):
- Thunder provisioning is DELIBERATELY ephemeral (provision→train→decommission):
  reaper every 15min, 18h hard uptime cap, concurrency 2, $100/day cap,
  CREDENTIAL-FREE VMs. (033 explicitly retracted the persistent-VM option for
  training.)
- A persistent service is the OPPOSITE: stays up, reaper-EXEMPT, holds its own
  credentials (ANTHROPIC_API_KEY + Stripe keys), stable inbound DNS+TLS,
  systemd keep-alive.
- So the gap = a persistent-service WRAPPER + credential delivery + DNS/TLS +
  a service_instances table + a parameterised setup script. Modest; mostly
  assembling existing pieces. NOT a greenfield build.

TWO PATHS TO DEPLOY THE ENGINE:
- Path A (manual now = B1): hand-start a small VM, apply the 007 recipe by hand,
  DNS→IP, set env incl. Stripe, AUTO_DELIVER=false, walk one order. CAPTURE the
  steps as script + nginx conf + systemd unit — that artefact is 80% of Path B.
- Path B (chassis workflow later): service-deployer orchestrator modelled on
  model-trainer — provision (persistent mode, reaper-exempt) → ship binary via
  presigned URL → ssh_exec the setup script → deliver credentials → register in
  service_instances → health-check. Reuses adapter primitives + orchestration
  pattern + 007 recipe.

"idea.uk in a normal workflow" = idea.uk as first consumer of service-deployer;
later the site-build pipeline invokes it for any site whose site_plan includes a
backend. Lightweight cousin (site_api_routes config) already exists in 007.
Distinction kept clear: deploying the engine BINARY (infra, Path A/B) vs
expressing the engine as chassis ACTIONS (Phase D / chassis-native, needs schema
pass) — complementary, not alternatives.

### Go-live mechanics captured (in the parallel doc)
Full env-var table + the fake-vs-Stripe provider switch (main.go: both
STRIPE_SECRET_KEY + STRIPE_WEBHOOK_SECRET set → real; else FakeProvider).
No Dockerfile in my outputs copy (user has one in their tree); binary builds fine.

### Next
- Immediate: push toward live = Stripe test keys (A8) + stand container/binary
  on one small box (B1 = Path A). First hands-on touch of the Layer 5 gap, done
  manually — exactly what Path B will later automate.
- Open: re-run agritec with WEB_SEARCH_MAX_USES=12 to confirm fewer provisional
  premises (optional; costs ~£2-4).
- Open (user's call): add the gpu-provisioner output_fields item to the guide
  lineage if they want both branches reconciled.

## CHECKPOINT 2026-06-04 (continued) — VM deploy artefacts drafted

User wants idea.uk deployed/maintained/CONTROLLED BY THE FRAMEWORK, not by hand,
worked toward the most efficient way. Also confirmed the chassis-native engine
(method as a site-planning input, Phase D) IS part of the plan.

The previous nginx/security/systemd/terraform files are NOT in the project
snapshot (only 007 documents the recipe). Drafted fresh from 007 + the engine's
real shape; flagged to the user to drop their real files in for me to align to.

### Drafted in idea-go/deploy/
- **setup.sh** — the one provisioning artefact. Idempotent, non-interactive,
  parameterised. Built to serve BOTH manual-now and chassis-later: the same
  script a person runs by hand is what service-deployer will `ssh_exec`.
  - Installs nginx (TWO-STAGE so it's idempotent and we OWN the conf rather than
    letting `certbot --nginx` rewrite it: stage 1 http+ACME-webroot+proxy →
    `certbot certonly --webroot` → stage 2 full http→https redirect + 443 TLS
    proxy), ufw (deny-in/allow-out + 22/80/443), fail2ban (sshd jail),
    unattended-upgrades, hardened systemd unit, binary.
  - Binary via IDEA_BINARY_URL (curl — the chassis presigned-B2-URL path, same
    as prepare_artefact_url) OR local IDEA_BINARY_PATH (manual scp path).
  - MODE=full (provision/rebuild) | MODE=update (swap binary + restart only).
  - Re-running = the rebuild path (idempotent).
  - HARDEN_SSH guarded: only disables password auth if authorized_keys exist
    (anti-lockout).
  - bash -n clean.
- **idea.env.example** — /etc/idea/idea.env template. NOT written by setup.sh
  (secrets stay out of the script). Sets REPORT_PRICE_GBP=29 (binary default is
  a stale 199), AUTO_DELIVER=false. Documents the Stripe-vs-Fake switch.
- **README.md** — manual Path A steps + chassis Path B design + the bridge
  ("setup.sh IS the capture; fold any box tweaks back into it") + the
  reused-vs-new breakdown.

### Bug caught while drafting (would have been mine)
makeDeliver() writes the SMTP-unset fallback report to a RELATIVE path
(delivered_*.md in CWD). systemd CWD defaults to / which ProtectSystem=strict
makes read-only → the write (error ignored with `_ =`) would SILENTLY fail.
This is exactly the AUTO_DELIVER=false operator-review phase the user starts in
(SMTP likely unset, operator reads the files). Fix: WorkingDirectory=$DATA_DIR
in the unit so relative writes land in the writable data dir (which is in
ReadWritePaths). Re-validated.

### The path the user chose
Framework-controlled is the goal. Most efficient route: do Path A manually now
to get live AND to produce the exact setup.sh, then Path B (service-deployer
workflow) consumes that same setup.sh. So Path A isn't throwaway — it's Path B's
payload. service-deployer = sibling of model-trainer; reuses adapter
provision/ssh_exec/presigned-URL transfer; the only genuinely new bits are
persistent-mode provisioning (reaper-exempt), credential/env delivery to the
box, a service_instances table, DNS wiring.

### Next options
1. Draft the service-deployer agent definition + workflow SQL (the chassis Path
   B) — but per house rule, needs a schema pass on agent_definitions + the
   adapter action contracts first (check schema before SQL).
2. Or the user runs setup.sh on a real box (Path A), reports what they hit, and
   we fold fixes back into setup.sh before automating.
3. Or align setup.sh to the user's actual previous nginx/systemd/terraform files
   once they paste them.

## CHECKPOINT 2026-06-04 (continued) — prior infra files reviewed; VM launch plan; setup.sh improved

User uploaded their year-old OVH infra files and chose Path A (run setup.sh on a
real box). Reviewed all of them; wrote **VM_LAUNCH_PLAN.md** (infra track,
separate from the idea.uk app work) with a full assessment + fixes + Path A
sequence + the chassis bridge.

### Their setup, understood
OVH box 51.89.148.216 = shared multi-domain reverse proxy terminating TLS and
forwarding ALL domains to one k8s NodePort (35.214.74.66:30080). idea.uk is NOT
in k8s — it's a Go service on a box. DECISION: dedicated VM for idea.uk (not the
shared proxy); engine on the box, nginx → localhost:8080.

### Secrets check — CLEAN
tfvars and tfstate do NOT contain the SSH password, htpasswd, or any private
key (passed at apply-time). No leak. Only PII = their home IP 176.25.120.48
throughout (their choice).

### Concrete bugs found in their year-old files (all catalogued in the doc)
A. `if ($remote_addr){ break; }` "whitelist" in nginx confs is a NO-OP — break
   doesn't bypass limit_req/auth (classic "if is evil"). Their IP was never
   actually exempt. FIX: geo+map pattern (folded into setup.sh, optional via
   WHITELIST_IPS).
B. felines.conf serves workdomain.co.uk's CERT for felines.co.uk/felines.uk →
   TLS name mismatch. FIX: per-domain cert (setup.sh does this).
C. variables.tf marks password_for_ovh_ssh + htpasswd_password sensitive=FALSE
   → would print in logs. FIX: sensitive=true; better, drop SSH password auth
   (key-only).
D. main.tf is only an SSH `whoami` test; the templatefile/remote-exec resources
   that push configs aren't in the upload. RECOMMEND: Terraform for VM
   provisioning ONLY; box config via setup.sh (the chassis-reusable artefact).
E. Their nginx .tpl is port-80-only and certbot edits it in place → re-applying
   Terraform WIPES certbot's 443 edits (silent TLS break). FIX: setup.sh OWNS
   the full conf + uses `certbot certonly` (never edits nginx) → idempotent.
   Biggest robustness win.
F. fail2ban ssh-custom-usernames jail uses `filter=%(sshd_log)s` (a log macro,
   not a filter) + custom filters not present → jails may silently fail. FIX:
   stock sshd + nginx-http-auth/limit-req jails; no custom filters.
G. nginx_log_monitor.py: 127.0.0.0/8 in the CHINESE prefix block (localhost
   bug); over-aggressive UA blocking (curl/wget/python-requests/Go-http-client/
   Java — would block legit server-side callers); reload/clear logic mis-nested;
   O(n^2) unbounded block files; no systemd unit. RECOMMEND: don't run it on
   idea.uk initially (fail2ban + nginx limit_req + in-app /audience-check limit
   suffice). Offered a corrected version as a follow-up rather than now.
H. No HSTS / security headers anywhere. FIX: added to setup.sh 443 block.

### setup.sh improvements folded in (bash -n clean, preamble verified both modes)
- Correct geo+map rate-limit whitelist via WHITELIST_IPS env (default empty) —
  replaces their broken `if break`. Verified it emits valid nginx with/without.
- Security headers on the 443 block: HSTS, X-Content-Type-Options,
  X-Frame-Options, Referrer-Policy.
- Size-based nginx logrotate drop-in (their `size 100M rotate 14` improvement),
  using `kill -USR1` postrotate.

### What I did NOT do (flagged, awaiting user)
- Did not rewrite nginx_log_monitor.py (recommend launching without it; will
  produce a corrected+systemd'd version if wanted).
- Did not rewrite their Terraform (recommend Terraform-for-provisioning-only;
  can produce a cleaned main.tf with sensitive=true + key-only SSH + a real
  provider block if wanted).
- Still want the user's ACTUAL config-pushing .tf (not in the upload) if they
  want me to align rather than supersede with setup.sh.

### Next (user runs Path A)
User stands up a real box with setup.sh, reports what breaks, we harden, THEN
build the service-deployer workflow around the proven script.

## CHECKPOINT 2026-06-04 (continued) — persistence design (requests/results → DB)

User wants idea.uk requests + results in a database, eventually the main
framework DB, but flagged (correctly) that going directly could be a security
hole. Wrote **PERSISTENCE_design.md**. Checked the live schema first (their rule).

### Schema facts confirmed (from the \d dump in schemas_all/schemas_some)
- Conventions: public + a `business_intel` schema owned by a SEPARATE restricted
  role `clients_user`; uuid PK gen_random_uuid(); jsonb; timestamptz now();
  snake_case; status text columns.
- business_intel holds collected/analytics data (businesses, business_prices,
  data_observations, discovery_candidates, companies_house_data, …) — the
  natural home for idea.uk orders, AND the restricted role is a second security
  boundary (ingest can run as clients_user, not superuser).
- Ownership hierarchy: clients → networks → sites (sites.id uuid, network_id,
  domain, settings jsonb, status).
- No existing customer-order/submission table (the *_requests tables are
  framework-internal: approval_requests, awaited_requests, input_requests,
  orchestration_requests, pending_requests). So a NEW table is needed.
- `scheduled_tasks` table IS the ingest-job mechanism (name, interval_seconds,
  target_agent_type, target_topic, pre_query, concurrency_group, enabled). A row
  fires an agent on an interval — exactly how the B2→Postgres ingest wires up.

### The design (three tiers, one-way flow)
1. Box LOCAL store (operational): orders/paid-state/idempotency/pending report.
   Recommend STAY JSON (keeps exposed tier dead-simple, no DB driver/server to
   secure) — OR upgrade to SQLite (modernc.org/sqlite, pure-Go, vendored) if
   they want local SQL. Honest caveat: SQLite = first non-stdlib dep, breaks
   "GOPROXY=off just works" (needs go mod vendor once). Lean: JSON on box,
   database lives in the framework.
2. CHANNEL: B2 dead-drop. Box writes immutable per-event record to idea-events/
   prefix via a write-only-scoped B2 key (or presigned PUT issued by trusted
   side = prepare_artefact_url, no standing creds). NO inbound path to cluster.
   Same pattern Thunder adapter already uses. Alternatives documented: Kafka
   (append-only but a path in), narrow HTTPS ingest (a path in), direct Postgres
   (THE HOLE — no). B2 chosen: zero inbound path, reuses existing pattern.
3. FRAMEWORK: a scheduled_tasks row fires an idea-ingest agent that pulls new B2
   records and upserts into business_intel.idea_orders (idempotent on order_ref
   via ON CONFLICT), runs as restricted clients_user role, marks B2 objects
   processed. Chassis PULLS; box never connects in. Down ingest → records pile
   safely in B2, replayable.

### Proposed table (clearly marked PROPOSED, reconcile with live schema first)
business_intel.idea_orders: id uuid pk, order_ref text unique (idempotency key),
site_id uuid (optional FK sites), domain, audience, notes, status
(requested|confirmed|paid|delivered|refunded), stripe_session (NO card data),
report_md text, result_meta jsonb, requested_at, delivered_at, ingested_at,
raw jsonb (audit). Optional idea_taster_events table for conversion analytics.

### Security properties
Exposed box holds NO core-DB creds, NO path into cluster. Worst case on
compromise: write junk to one B2 prefix (ingest validates). One-way drop,
trusted side has initiative. Ingest = least-privilege role on one table. No card
data anywhere (only Stripe opaque refs). Decoupled — framework view rebuildable
by replaying B2.

### Phased
Phase 1 (now, box): keep local store + add B2 record-writing on terminal events.
Durable + off the box, no hole. Phase 2 (when ready, framework): create table +
idea-ingest scheduled task. Box doesn't change again.

### Open decisions asked of user
1. Box store: JSON (recommended) vs SQLite (local SQL, one vendored dep)?
2. Schema home: business_intel (recommended, restricted role) vs public?
3. Store taster events too (conversion analytics) or orders only?
4. Channel: B2 polling now (recommended); Kafka later only if near-real-time needed.
Then I write (a) box-side B2 record-writing, (b) finalised DDL + idea-ingest
agent/scheduled-task — after a proper live-schema look.

## CHECKPOINT 2026-06-04 (continued) — persistence decisions locked; VM recommendation

### Persistence decisions LOCKED (updated PERSISTENCE_design.md §10)
1. Local store JSON (no SQLite) — box stays stdlib-only.
2. New `ecommerce` schema for commerce (orders + taster_events); idea reports in
   `clients_db`.
3. Store taster events (conversion signal).
4. B2 polling now, Kafka later.
Refined tables: ecommerce.orders (txn, NO report content, order_ref unique
idempotency key, stripe_session opaque/no card data), ecommerce.taster_events
(funnel signal), clients_db idea_reports (deliverable: report_md + candidates
jsonb, linked by order_ref).
FLAGGED FORK (awaiting user): is clients_db a separate DATABASE or a SCHEMA?
- Separate DB (likely, given _db suffix + clients_user role): NO cross-DB FK in
  Postgres → link ecommerce.orders ↔ idea_reports by order_ref VALUE; ingest
  holds two restricted connections. Clean commerce/deliverable separation.
- Schema: normal cross-schema FK works.
Designed B2 record to be Kafka-message-shaped (self-contained event w/ order_ref,
type, status, payload) so the Kafka upgrade is a transport swap, not a redesign.

### VM recommendation (search-grounded, 3 searches: Hetzner, Oracle, spot)
SPOT VERDICT: NOT for idea.uk. Consensus from search — spot/preemptible (60-90%
cheaper) is for fault-tolerant/stateless/batch; explicitly WRONG for user-facing
services + DBs needing continuous uptime. idea.uk is always-on with a Stripe
webhook; reclamation (30s-2min notice) kills in-flight reports; Stripe retries
~3 days so survivable but needless fragility for a ~€4/mo on-demand box. Spot is
right for the engine's batch/training work later, not the live front door.

RECOMMENDED: Hetzner Cloud.
- CAX11 (ARM Ampere) 2vCPU/4GB/40GB NVMe ~€3.79/mo, EU-only. Best price/perf.
- CX22 (x86) 2vCPU/4GB/40GB ~€4.35/mo, EU + US (Ashburn VA, Hillsboro OR).
- Both: 20TB traffic + IPv4 + DDoS + firewall included. No free tier, no spot.
  Price rise April 1 2026 on some plans. Has Server Auction for cheap dedicated.
- Engine is I/O-bound (waits minutes on LLM calls) → smallest box is plenty;
  4GB tiers give headroom for nginx + binary + JSON store.

FREE OPTION: Oracle Cloud Always Free — Ampere A1 up to 4 OCPU/24GB + 200GB +
10TB egress/mo, never expires, genuinely £0. BIG caveat: chronic "Out of
Capacity" for ARM A1 on free accounts; workaround = upgrade to Pay-As-You-Go
(still free within A1 limits; card + temp ~$100 hold released) and pick a 3-AD
region (Ashburn/London). More manual setup. Unbeatable on price if you tolerate
the capacity dance + ARM.

ARM NOTE: CAX11 and Oracle A1 are ARM → build the Go binary arm64:
`GOARCH=arm64 GOOS=linux go build -o idea .` (trivial for Go). setup.sh already
takes the binary via IDEA_BINARY_URL/PATH so arch is just the build flag.

Others (Vultr/DigitalOcean ~$6+/mo, Netcup, Lightsail $5) cost more for the same;
not recommended over Hetzner for this.

## CHECKPOINT 2026-06-04 (continued) — idea.uk box provisioned (Hetzner)

User provisioned the idea.uk VM. No ARM was available, so it's x86 (silver lining:
plain native amd64 build, no cross-compile, binary stays static/stdlib-only).

BOX FACTS:
- Hetzner CX23, 2 vCPU / 4GB / 40GB, €5.99/mo, Nuremberg (Germany).
- Name: idea1  (server #136872964, project 14860308)
- IPv4: 116.203.204.115
- IPv6: 2a01:4f8:1c18:7c31::/64 (primary likely ::1; confirm on box with `ip -6 addr`)
- Dual-stack (IPv4 + IPv6 as decided — IPv4 required for Stripe webhooks + IPv4-only users).

GO-LIVE SEQUENCE GIVEN (Path A, in chat):
1. DNS: A idea.uk→116.203.204.115; AAAA→box v6. Verify with dig before setup.
2. (rec) ssh-copy-id so setup.sh can harden SSH (else it skips, anti-lockout).
3. Build: GOPROXY=off GOTOOLCHAIN=local GOOS=linux GOARCH=amd64 go build -o idea .
4. scp idea + deploy/setup.sh + deploy/idea.env.example to root@IP.
5. Fill /etc/idea/idea.env (ANTHROPIC_API_KEY, PUBLIC_BASE_URL, OPERATOR_EMAIL,
   INTERNAL_API_KEY=openssl rand -hex 32, AUTO_DELIVER=false, REPORT_PRICE_GBP=29;
   leave STRIPE_* BLANK first → FakeProvider, test flow w/o money).
6. DOMAIN=idea.uk LETSENCRYPT_EMAIL=… IDEA_BINARY_PATH=/root/idea bash /root/setup.sh
7. Verify: systemctl status idea; journalctl -u idea; curl https://idea.uk/health.
8. Stripe later: webhook https://idea.uk/stripe/webhook (checkout.session.completed),
   put sk_test_/whsec_ in env, restart; test cards → live → refund own-card test.

ASSUMPTIONS FLAGGED TO USER: OS = Ubuntu (setup.sh uses apt/ufw; Debian ok, else no).
Hetzner Cloud Firewall is separate from box ufw + not applied by default → ufw
(22/80/443) from setup.sh is the active control.

AWAITING: user runs it, pastes journalctl + /health output. Likely first wrinkle =
empty ANTHROPIC_API_KEY (service starts but real work fails). Then harden setup.sh
against whatever real-box issues surface before it becomes service-deployer payload.

## CHECKPOINT 2026-06-04 (continued) — first real-box run: certbot placeholder-email abort

Path A first run on the Hetzner box (idea1, 116.203.204.115). DNS confirmed
correct (dig +short idea.uk → box IP). setup.sh got through packages, user,
binary (/opt/idea/idea), systemd unit, nginx stage-1 — then certbot FAILED:
"ACME server believes you@example.com is an invalid email address." Cause: the
user ran with the literal placeholder LETSENCRYPT_EMAIL=you@example.com from my
instructions (example.com is reserved; LE rejects it).

Knock-on: set -e aborted the script AT certbot, so everything after never ran —
including `systemctl restart idea`. Hence: status inactive(dead) but enabled;
journalctl no entries (never started); curl 443 connection refused (no service +
no stage-2/443 conf). Not a code bug — a half-completed run.

IMMEDIATE FIX given: re-run with a real email (idempotent → issues cert, writes
443 conf, ufw, fail2ban, starts service):
  DOMAIN=idea.uk LETSENCRYPT_EMAIL=<real> IDEA_BINARY_PATH=/root/idea bash /root/setup.sh

setup.sh HARDENED (the surface-and-harden loop):
1. Up-front guard: reject LETSENCRYPT_EMAIL ending @example.com/@example.org with
   a clear error before doing any work (fail fast, not minutes later).
2. certbot failure now NON-FATAL (|| log warning): the box still gets ufw +
   fail2ban + the service STARTED on HTTP, instead of aborting half-configured.
   stage-2 (443) already guards on the cert existing, so it stays HTTP-only until
   a successful re-run upgrades it to HTTPS. Much better degraded state.
bash -n clean; guard logic verified. User can re-run the on-box copy as-is with a
real email now (placeholder was the only blocker); the hardened copy is for
future/chassis use.

Candidate debugging-guide item: "a setup/provisioning step that aborts under
set -e can leave a box half-configured; make external-dependency steps (cert
issuance, DNS) degrade gracefully and validate obvious placeholders up front."

Also reminded user what journalctl is (reads systemd journal; -u unit, -n N,
--no-pager, -f follow).

## CHECKPOINT 2026-06-05 — second real-box issue: inline comments in env file (systemd)

Re-run with real email SUCCEEDED on the cert (curl now gives HTTPS 502, not
connection-refused → nginx up on 443 + cert issued + proxying). But the service
crash-looped (restart counter 39), nginx → 502.

ROOT CAUSE (my artefact bug): idea.env.example had INLINE comments after values
(e.g. `PORT=8080   # must match SERVICE_PORT in setup.sh`). systemd
EnvironmentFile does NOT strip inline comments — only treats # as a comment at
the START of a line. So PORT was read as the literal "8080   # must match..." →
net.Listen → "lookup tcp/8080...: unknown port" → exit 1 → restart loop. Same
swallowing made price=£0 (REPORT_PRICE_GBP read as "29   # NB...").
Log tell: `idea.uk service on :8080   # must match SERVICE_PORT in setup.sh
(auto_deliver=false, price=£0)` — the comment text embedded in the value.

FIX given to user (preserves their entered secrets):
  sed -i 's/[[:space:]]*#.*$//' /etc/idea/idea.env   # safe: no value contains #
  then verify ANTHROPIC_API_KEY + INTERNAL_API_KEY are actually filled
  systemctl restart idea; curl https://idea.uk/health

ARTEFACTS HARDENED:
- idea.env.example REWRITTEN: every comment on its own line; added a header
  warning about the systemd inline-comment trap. Verified no value has a trailing
  inline comment.
- setup.sh: added a non-fatal guard that greps the env file for inline comments
  after values and, if found, WARNS with the exact offending lines + the sed fix
  (so it fails loudly up front, not as a crash loop). bash -n clean; guard tested.

Progress state: TLS/nginx 443 working. Once the env is de-commented + keys
filled, service should bind :8080 and /health should return OK.

DEBUGGING-GUIDE CANDIDATES (two from this deploy, both artefact-level):
- "Let's Encrypt rejects example.com placeholder emails; validate obvious
  placeholders up front; make cert issuance non-fatal so the box isn't left
  half-configured under set -e."
- "systemd EnvironmentFile does not strip inline comments — a `# ...` after a
  value becomes part of the value (crashed PORT). Keep env-file comments on their
  own lines; guard for it."

## CHECKPOINT 2026-06-05 — idea.uk LIVE on the box (Path A working)

Service active(running) on the Hetzner box; HTTPS works; /health green:
  {"auto_deliver":false,"ok":true,"price_gbp":24,"provider":"*main.FakeProvider"}
Clean startup line confirms the inline-comment fix took:
  "idea.uk service on :8080 (auto_deliver=false, price=£24)"
So Layer 2 (idea.uk) is DEPLOYED and serving over TLS on a real box. FakeProvider
(Stripe not wired yet) as expected. The crash-loop journal entries above are
pre-fix history (counter ~197); current process is clean. (reset-failed optional.)

OPEN: price_gbp reads 24, not the locked £29. Can only come from the env file
having REPORT_PRICE_GBP=24 (sed only stripped comments; wouldn't change 29→24).
Flagged to user to confirm/fix:
  sed -i 's/^REPORT_PRICE_GBP=.*/REPORT_PRICE_GBP=29/' /etc/idea/idea.env; restart.

NEXT verification given:
- /health doesn't exercise ANTHROPIC_API_KEY. Test the taster to confirm real
  runs + the key: curl "https://idea.uk/audience-check?domain=agritec.uk&audience=UK%20small%20farmers"
  (cheap Step-1 call). Or the taster widget at https://idea.uk.
- Then: full path proven on FakeProvider → wire Stripe test keys → paid order
  walk → live keys → own-card test refunded via /refund.

ARTEFACT STATUS: both real-box artefact bugs (placeholder LE email; systemd
inline-comment env parsing) fixed in setup.sh + idea.env.example AND validated on
the real box (cert issued, service binds, /health green). The hardened setup.sh
is now a much more trustworthy payload for the future service-deployer (Path B).

## CHECKPOINT 2026-06-05 — page 404 fixed (embedded landing page); audience-check curl corrected

Price now £29 (confirmed via /health). Two issues this turn:

1. audience-check "POST only": my earlier test curl was wrong (GET + domain=).
   The handler is POST with FORM fields `business` and `audience` (not domain).
   Correct test:
     curl -sS -X POST https://idea.uk/audience-check \
       --data-urlencode "business=agritec.uk" --data-urlencode "audience=UK small farmers"
   Returns HTML verdict; errors if ANTHROPIC_API_KEY unset.

2. https://idea.uk → 404: the Go service is API-ONLY. routes() registers
   /health /capacity /audience-check /subscribe /request /confirm /decline
   /stripe/webhook /internal/run /order/success /order/cancel — but NOTHING at
   "/", and no static serving. The landing page (idea_uk_fakedoor.html) wasn't
   being served at all. Page uses same-origin relative paths (fetch('/audience-check'),
   action="/request", action="/subscribe", fields business/audience) → designed
   to be served from idea.uk itself.

FIX (structural, chassis-aligned): EMBED the page in the binary, serve at "/".
- Copied idea_uk_fakedoor.html → idea-go/page.html.
- service.go: added `_ "embed"`, `//go:embed page.html var pageHTML []byte`, and
  a home handler (serves pageHTML at "/", http.NotFound for any other unmatched
  path), registered mux.HandleFunc("/", a.home).
- Rationale over nginx-static: one self-contained artefact (binary carries the
  page) → matches the chassis "ship the binary" model; nginx stays trivial
  (TLS + proxy); page version-locked to the binary. Cost: rebuild to edit the
  page (fine for a fake-door). build/vet/test clean; confirmed page bytes are in
  the binary.

setup.sh ergonomic fix: MODE=update no longer requires LETSENCRYPT_EMAIL (only
full mode validates email + placeholder guard). DOMAIN still required. (Cleaned a
duplicate-declaration slip during the edit.) bash -n clean.

REDEPLOY to get the page live (binary change only; nginx untouched):
- page.html MUST be in the module dir on the laptop (go:embed needs it at build).
- GOPROXY=off GOTOOLCHAIN=local GOOS=linux GOARCH=amd64 go build -o idea .
- Atomic swap + restart (no setup.sh needed):
    scp idea root@116.203.204.115:/opt/idea/idea.new
    ssh root@116.203.204.115 'chmod 755 /opt/idea/idea.new && mv -f /opt/idea/idea.new /opt/idea/idea && systemctl restart idea'
  (or MODE=update via setup.sh). Then curl https://idea.uk/ → the page.

NOTE: nginx security headers + logrotate + geo-whitelist improvements live in the
canonical setup.sh but are NOT yet on the box (box has the first setup.sh copy;
those apply on a MODE=full re-run with the refreshed setup.sh — optional, separate
from getting the page up).

## CHECKPOINT 2026-06-05 — landing page copy rewritten (plainer, human tone)

Page is live and working. User asked to change the tone: stop sounding like an
LLM. Avoid words like honest / gate / deck / asset (and leverage, robust,
seamless, "surface" as a verb), and drop negative "X, not Y" framing (their
example: "A report you can act on, not a deck."). Keep punchy headlines, but make
the EXPLAINING copy plain and matter-of-fact so a human reader who doesn't read
AI reports all day focuses on the point, not on decoding words.

Edits to idea-go/page.html (and synced to idea_uk_fakedoor.html):
- h1: "Tested. Verified. Honest." → "Tested. Verified. Ranked."
- what-you-get h2: "A report you can act on, not a deck." → "A report you can act on."
- what-you-get body: dropped "asset", "premise fails on verification",
  "five-factor", "honest verdict", "advances the gate" → plain descriptions.
- specimen labels: "Asset:" → "Relies on:", "Capability:" → "Uses:" (×3 each;
  the real delivered report uses a different format so no mismatch).
- how-it-works h2: "Generated across four lenses. Filtered hard. Verified." →
  "Ideas from four angles. Most get cut. The rest get checked."
- steps: "challenge"→"check", "kill"→"cut", "survivors"→"what's left"; bodies
  plainer; dropped "Not one brainstorm — four" negative framing.
- callout: label "Why it isn't another AI brainstorm" → "How this differs from
  an AI brainstorm"; "Multi-model/Capability-current/Honest results" →
  "Different AI models / Kept up to date / Straight answers"; dropped
  "rubber-stamp", "stale self-knowledge".
- price list: "not a one-shot brainstorm" removed; "Honest scoring" → "A clear
  score"; "most valuable" → "most useful"; "Web-verified facts" → "Facts we've
  checked".
- "surface (anything)" → "find/turn up" in hero, disclosure, and the form's
  terms checkbox.
Verified: no flagged terms remain in visible copy; HTML tags balanced.

Kept untouched (good punchy headlines, no flagged words): hero eyebrow, taster
h2 ("30-second audience check, on the house"), price h2 ("One price. One report.
No payment until we confirm.").

REDEPLOY: page is embedded in the binary, so this needs a rebuild + binary swap
(same as before): ensure page.html in the module, go build (amd64), scp to
/opt/idea/idea.new, mv -f over /opt/idea/idea, systemctl restart idea.

## CHECKPOINT 2026-06-05 — post-submit pages + taster upsell rewritten; contact email wired

User: improve the pages shown after the request form and the newsletter form; they
need a contact email on the enquiry form; and the taster footer copy must explain
free-vs-£29 in plain non-AI language (it was dense: "idea generation across four
lenses, a critique pass against the free AI alternative ... operator-risk score").

Found: those responses were bare unstyled fragments (writeHTML emits no full page),
so post-form navigations rendered as black-on-white text after the styled landing
page. Also the live page showed literal CONTACT_EMAIL / MONTH_SLOTS placeholders
(page.html had unfilled placeholders, served raw).

Changes (idea-go):
- Config += ContactEmail, Slots. loadConfig reads CONTACT_EMAIL (falls back to
  OPERATOR_EMAIL) and MONTH_SLOTS (falls back to "a limited number of").
- App += landingHTML []byte; NewApp templates pageHTML once at startup, replacing
  CONTACT_EMAIL + MONTH_SLOTS; home() now serves a.landingHTML (placeholders gone
  on the live page).
- New a.page(title, body) helper: full brand-styled HTML doc (cream/ink/rust,
  Fraunces+IBM Plex via Google Fonts, wordmark bar, footer with the contact email
  + links). firstName() helper for a friendly greeting.
- Rewrote subscribe / handleRequest / orderSuccess / orderCancel to use page() with
  plain, warm copy (what we got, what happens next, when, refund, contact). Request
  page greets by first name (escaped).
- Taster footer rewritten (audience_check.go): plain free-vs-£29 explanation — what
  the free audience check is, then what the £29 report gives (3–6 ranked ideas, each
  explained in plain terms, fact-checked re competitors/prices/rules, cheapest demand
  test, straight "nothing worth building yet" if so, written report by email, pay
  only after we confirm, money back if nothing worth acting on). No method jargon.
  Also "Our reframe —" heading → "Who we think you should actually be selling to".
  CTA "Get the full report" → "Request the full report — £29 →".
- page.html: added .taster-upsell + .taster-cta CSS (upsell resets white-space:normal
  since .taster-result is pre-wrap). Re-synced idea_uk_fakedoor.html.
- idea.env.example documents CONTACT_EMAIL + MONTH_SLOTS (comments on own lines).
- Updated service_test.go assertion to the new CTA text. build/vet/test clean.

REDEPLOY (rebuild — page embedded + Go changed): set CONTACT_EMAIL (and optionally
MONTH_SLOTS) in /etc/idea/idea.env, then build amd64, scp to /opt/idea/idea.new,
mv -f over /opt/idea/idea, systemctl restart idea.

## CHECKPOINT 2026-06-05 — /terms and /refund-policy pages written + served

Linked-but-404 routes /terms and /refund-policy now exist. Plain-language DRAFTS
(flagged to user: not legal advice; UK solicitor must review before taking real
payments — runbook A6).

Implementation (idea-go/service.go):
- contactEmail() helper (ContactEmail, else OperatorEmail); page() now uses it.
- page() CSS extended for long-form: --paper-2 added; main 600→680px; h2, ul/ol/li,
  .meta, .note styles. (Benefits all page()-wrapped pages.)
- termsPage / refundPage handlers: strings.ReplaceAll(body, "{{EMAIL}}", contactEmail())
  then a.page(...). Registered /terms + /refund-policy in routes().
- termsBody / refundBody raw-string constants appended to service.go. {{EMAIL}} token
  filled at serve time; [bracketed] items left for the user.
- Verified by a throwaway in-package test (since background servers are flaky in the
  sandbox): both render, {{EMAIL}}→email (0 tokens left), terms 13 sections, refund 6.
  Test removed after; full suite green.

Content covered —
Terms: who we are [trading name/address placeholders], what you get, what this is NOT
(information service not professional advice; check cited sources), ordering/payment
(request first, pay only after confirm, £29), your part (accurate info, no unlawful
use, don't pass off as own advice), accuracy (no guarantee, world changes, verify),
liability (reasonable care; not liable for business losses; cap = amount paid; nothing
excludes unexcludable e.g. fraud/death/PI; consumer rights unaffected), using the
report, data (used only to deliver; not sold; [privacy policy placeholder]), refunds
(→ refund page), changes (version at order applies), governing law England & Wales
[confirm w/ adviser], contact.
Refund: short version; pay only after confirm; 14-day "not useful" full refund no
detailed reason; fault/non-delivery full refund; how to claim (email from order
address, 2 wd reply, 5–10 wd to original method); statutory rights on top [adviser to
confirm Consumer Contracts Regs 2013 for a bespoke report begun with consent]; contact.

Note: terms quote delivery "within five working days of payment, usually sooner" =
conservative outer bound; the friendlier pages (orderSuccess, landing) say ~72h. 72h
is within 5 wd + "usually sooner", so consistent (terms = outer bound by design).

Previews saved to outputs/terms_preview.html + refund_policy_preview.html (sample
CONTACT_EMAIL baked in for preview; live value from env).

REDEPLOY (Go change, page-embedded binary): rebuild amd64, scp to /opt/idea/idea.new,
mv -f over /opt/idea/idea, systemctl restart idea. CONTACT_EMAIL in /etc/idea/idea.env
feeds the footer + these pages.

## CHECKPOINT 2026-06-05 — privacy policy added; terms hardened (AI disclaimer); docs updated

Legal pages — all three now live (served from service.go constants via a.page(), {{EMAIL}}
filled at serve time, linked in both footers + the disclosure):
- /terms  (termsBody)        — see prior checkpoint; CHANGES this turn below.
- /refund-policy (refundBody) — see prior checkpoint (unchanged this turn).
- /privacy (privacyBody)     — NEW this turn.
All three are plain-language DRAFTS; UK solicitor review required before real payments (A6).

Terms changes this turn (user: low risk appetite, trading name idea.uk no address, make
the AI nature + its risks explicit, use-is-on-them):
- "Who we are": now "operated under the name idea.uk, in the United Kingdom" — trading
  name idea.uk, NO address (placeholders removed).
- Replaced the "Accuracy" section with "The report is produced using AI, and AI can be
  wrong": states plainly that reports are AI-generated; AI can be confidently wrong and
  invent facts/figures/companies/prices/quotes/sources ("hallucination") and be stale; a
  person reviews but cannot catch everything and we do not claim so; treat everything as
  to-be-checked; verify sources/figures/competitors/prices/rules and take own advice
  before spending/committing; we do not promise accuracy/completeness/currency/fitness;
  "What you do with the report, and any decision you make based on it, is entirely your
  responsibility and not ours."
- "The information you give us": placeholder removed; now links /privacy.
- (Unchanged and still doing the liability work: "What this is not" = info service not
  professional advice; "Our responsibility to you" = reasonable care, not liable for
  business losses, cap = amount paid, nothing excludes unexcludable, consumer rights
  unaffected.)

Privacy policy (privacyBody) — UK-GDPR-shaped, low-risk posture:
- Controller = idea.uk, UK, contact {{EMAIL}}. What we collect (request: name/email/
  business/audience/notes; updates: email; payment via Stripe — never full card; free
  audience check: processed, not stored; technical logs incl. IP). Lawful bases
  (contract / legitimate interests / consent / legal duty). Processors NAMED: Stripe
  (payments), Anthropic (AI — business+audience sent, NOT name/email), hosting+email
  [bracketed to confirm]. International transfer flagged (Anthropic, US) with safeguards
  [bracketed]. Retention [bracketed: ~6yr financial]. Rights + ICO (ico.org.uk). No
  cookies/analytics/tracking. Security wording measured ("reasonable steps", no
  absolute promise). Children: business service, not under-18. Changes; contact.
- Bracketed placeholders for user/adviser: hosting+email provider names, transfer
  safeguards per supplier, retention period.

Footers: page() wrapper footer + landing footer now include Privacy (and Refunds).
page.html re-synced to idea_uk_fakedoor.html. service_test.go untouched (still green).
build/vet/test clean. Verified all three render: {{EMAIL}}→email (0 tokens left),
terms 13 / privacy 12 / refund 6 sections; previews saved (terms_preview.html,
privacy_preview.html, refund_policy_preview.html).

Docs updated this turn (user request):
- 016_debugging_guide_v2_32.md: added §11 "idea.uk standalone service — page-serving
  and deploy gotchas": MISSING PAGES (404 on linked path = no mux.HandleFunc; check
  grep mux.HandleFunc vs hrefs; curl status to tell Go-404 from nginx) and MISSING
  DESIGN (writeHTML bare fragment renders unstyled on full navigations; wrap in
  a.page(); pre-wrap reset note), plus literal-placeholder fix and deploy gotchas
  (set -e half-config box, systemd inline-comment env, LE example.com, text-file-busy
  → mv -f swap).
- idea_uk_architecture_and_deployment.md: appended "Update — 2026-06-05" decisions
  section (live on Hetzner; page embedded not B2; shared page() wrapper; policy pages
  incl. privacy + AI disclaimer + trading name; CONTACT_EMAIL/MONTH_SLOTS config; plain
  copy; redeploy steps).

REDEPLOY (Go change, embedded page): rebuild amd64, scp to /opt/idea/idea.new, mv -f
over /opt/idea/idea, systemctl restart idea. CONTACT_EMAIL in /etc/idea/idea.env feeds
all three pages' footer + contact lines.

## CHECKPOINT 2026-06-05 — email sending decided: AWS SES (London) for sending, contactforsales.com for receiving

Question explored: how does idea.uk send email, and does Hetzner provide a mail server.

Findings (web-checked):
- Hetzner Cloud blocks OUTBOUND ports 25 AND 465 by default; unblock only after ~1
  month + first paid invoice, case-by-case. Port 587 (submission to an external
  relay) is OPEN by default → use a transactional provider on 587.
- Hetzner DOES have mailboxes, but via their **Web Hosting** product (konsoleH),
  separate from Cloud. It's a personal-inbox product (SMTP auth, throttled), not a
  transactional sender — fine for receiving, poor fit for app/transactional mail.
- Don't run our own mail server on the box: cloud-IP deliverability/reputation is
  poor and we'd own SPF/DKIM/DMARC/PTR. Not worth it for payment/report email.
- Transactional market: US (SendGrid/Mailgun/Postmark), French (Brevo/Mailjet), NZ
  (SMTP2GO). No genuine UK-HQ transactional provider found. EmailOctopus IS UK
  (London) but it's newsletter/marketing — its transactional path is just a UI over
  your own Amazon SES. UK mailbox hosts for receiving: Krystal, Mythic Beasts,
  Fasthosts (IONOS = German parent, UK ops).
- Amazon SES is available in the **London region (eu-west-2)**; pricing identical
  across regions at ~$0.10 / 1,000 emails (+$0.12/GB attachments); free tier 3,000
  message charges/month for the first 12 months. London region = email processed in
  the UK (data residency), though AWS is a US-parent company.

DECISIONS (locked):
- **Sending: AWS SES, London region (eu-west-2).** UK data residency, ~pennies at
  our volume, free year one. Drops into the service's existing SMTP_* vars on 587.
- **Receiving: contactforsales.com** — user's existing working email domain. Public
  CONTACT_EMAIL + OPERATOR_EMAIL → an address at contactforsales.com (exact local
  part TBC from user).
- **From/Reply-To (recommended, user to finalise):** send branded from idea.uk
  (e.g. SMTP_FROM=noreply@idea.uk, SMTP_FROM_NAME="idea.uk") with
  SMTP_REPLY_TO=<addr>@contactforsales.com so replies land in the working inbox; no
  mailbox needed on idea.uk. Alternative (simpler, off-brand): send from the
  contactforsales.com address directly. Newsletter later: EmailOctopus (UK, free to
  2,500 subs) — separate from transactional.

Code change this turn (service.go makeDeliver): added optional **SMTP_FROM_NAME**
(display name → "Name <addr>" From header) and **SMTP_REPLY_TO** (Reply-To header).
Envelope sender stays the bare SMTP_FROM (must be a verified SES identity).
Backward-compatible (unset → old behaviour; host unset → still writes delivered_*.md).
idea.env.example email block rewritten with the SES London settings + the new vars
(SMTP_HOST=email-smtp.eu-west-2.amazonaws.com, port 587, generated SES SMTP creds).
build/vet/test clean.

Privacy policy (privacyBody) updated: hosting bracket → named **Hetzner** (servers in
Germany); email bracket → **Amazon SES, London region** (email processed in the UK) +
a bracket for the contactforsales.com **mailbox provider** (name TBC). Transfers
section refined: UK/EU residency where possible (Germany servers, London email); the
main outside-UK transfer is **Anthropic (US)** for the business/audience content;
safeguards bracket left for adviser. Retention bracket still open.

SES setup checklist (to do, account-side):
1. SES console → London region (eu-west-2). 2. Verify the idea.uk domain (and/or the
   From address) → add the SES DKIM CNAME records to idea.uk DNS at Hetzner DNS; add
   SPF (TXT include amazonses.com) and a DMARC TXT (start p=none). 3. Request
   PRODUCTION ACCESS (new SES accounts are in sandbox = can only send to verified
   addresses; approval ~24h with a use-case note). 4. SMTP settings → Create SMTP
   credentials (these are an IAM user, NOT the AWS password). 5. Put SMTP_HOST/PORT/
   USER/PASS/FROM(/FROM_NAME/REPLY_TO) in /etc/idea/idea.env; set CONTACT_EMAIL +
   OPERATOR_EMAIL to the contactforsales.com address. 6. Flip AUTO_DELIVER only after
   a real report has been reviewed.

PENDING from user: exact contactforsales.com address; confirm From choice (branded
idea.uk vs contactforsales.com). Then set env on the box + redeploy (this turn's
mailer change + privacy edits need a rebuild: amd64 build → scp /opt/idea/idea.new →
mv -f → systemctl restart idea).

## CHECKPOINT 2026-06-05 — email architecture thread: operator domain leopardess.uk; START on Clook (defer SES)

(Separate "discuss email" thread — not fully finished; recording decisions + open items.)

Framing established: email = THREE independent parts, none of which is the Go box —
(1) DNS records (MX/SPF/DKIM/DMARC), (2) inbound handler (the MX target), (3) outbound
sender. Web (A → box) and mail (MX → mail host) point different ways; that's normal,
not "messy". So DNS can stay at Hetzner while mail lives elsewhere.

Landscape findings (web-checked):
- Hetzner: DNS Console CAN publish MX/SPF/DKIM/DMARC. Hetzner CLOUD has no mailboxes;
  mailboxes only via the separate Hetzner Web Hosting (konsoleH) product. No need for a
  2nd Hetzner account. Don't run our own mail server on the box (cloud-IP reputation).
- No genuine UK-HQ transactional provider. SES London (eu-west-2) = UK data residency,
  ~$0.10/1k + free 3k/mo for 12 months, but AWS is US-parent. EmailOctopus = UK but
  newsletter/marketing (transactional only via your own SES). UK mailbox/hosting: Clook
  (£5/mo, 100GB email, UK data centres), Krystal, Mythic Beasts, Fasthosts. ImprovMX =
  free forwarding (US), INBOUND ONLY (its sending is a paid add-on). Cloudflare Email
  Routing = free but requires DNS on Cloudflare.
- Clook (existing account, UK, cPanel specialist): has Email Packages (£5/mo, 100GB) and
  Managed Cloud Servers (+ an unmanaged cloud option). Can't run the Go service on cPanel
  shared hosting (no systemd/long process/root); could buy a Clook cloud server but no
  reason — keep the Go service on Hetzner.

Key principle for the generic domain: it's the OPERATOR IDENTITY (platform + transactional
+ support mail, and where replies land) — NOT a shared bulk-sender for thousands of
lead-gen sites (one site's spam complaints would poison the shared reputation). Any site
that emails at volume should use its OWN sending domain. contactforsales.com rejected
("sales" reads like a funnel, hurts trust/deliverability); idea.uk rejected (single
product, not generic).

DECISIONS (locked):
- **Operator/generic domain = leopardess.uk** (chosen from the user's domain list; a real
  brand, category-neutral; user already owns the leopardess cluster).
- **Start ALL email on Clook** (the £5/mo email package), BOTH inbound and outbound, for
  leopardess.uk. **Defer SES.** Rationale: at leopardess.uk's low volume (a few system/
  support emails a day) Clook is fine; SES and Clook are interchangeable to the app (it
  speaks plain SMTP), so we can move to SES later by changing env vars only — NO code
  change, no lock-in.
- **Use ONLY leopardess.uk as the sender — NOT idea.uk.** idea.uk's transactional/system
  mail will go out from a leopardess.uk address (so no idea.uk DKIM/verification needed,
  and everything funnels through one mailbox).
- **On the idea.uk pages, state clearly "by leopardess.uk"** so a leopardess.uk sender on
  an idea.uk receipt isn't confusing — establishes the product↔operator relationship
  on-site. (This reverses the earlier branded-idea.uk-sender plan; simpler, one identity.)

Mechanics for the chosen path:
- Outbound: Clook authenticated SMTP (port 587) + a leopardess.uk mailbox's credentials →
  set SMTP_HOST (Clook's outgoing server), SMTP_USER, SMTP_PASS, SMTP_FROM=<addr>@leopardess.uk
  on the box. Code already supports SMTP_FROM / SMTP_FROM_NAME / SMTP_REPLY_TO (no change).
- Inbound: Clook mailbox/forwarder for leopardess.uk → forward to aaa@designconsultancy.co.uk
  (the user's Gmail). MX (wherever leopardess.uk DNS lives) → Clook's mail servers.
- DNS for leopardess.uk: keep at Hetzner DNS Console (consistent with idea.uk) or let Clook
  host it; MX → Clook either way. (Open — not yet set up.)

KNOWN TRADEOFF to revisit (on record): Clook outbound is shared-IP SMTP with host send
limits / acceptable-use on automated sending, and little bounce/delivery visibility. Fine
for low volume; if idea.uk payment/report volume or stakes grow, or deliverability slips,
move idea.uk sending to SES London (env swap only). ACTION: ask Clook their hourly send
limit + policy on application/SMTP sending before relying on it.

OPEN / PENDING (thread not finished):
- Is designconsultancy.co.uk on Google Workspace? (decides whether a forwarder is needed,
  or whether to just add leopardess.uk to Workspace.)
- Exact local parts + From display name (hello@ / noreply@ / support@ @leopardess.uk).
- Buy/enable Clook email package; add leopardess.uk; create the mailbox.
- Then lay out exact DNS records (MX→Clook, SPF, Clook DKIM, DMARC) to paste in.
- PAGE CHANGE (not yet done): add "by leopardess.uk" branding to idea.uk pages — landing
  footer + the a.page() wrapper footer + (consider) the policy pages.
- Set env on the box (SMTP_* → Clook; CONTACT_EMAIL + OPERATOR_EMAIL → the leopardess.uk
  address) + redeploy.
- PRIVACY POLICY FIX (privacyBody currently names "Amazon SES, London region" as the email
  sender — now INACCURATE since we're on Clook): change the email processor to Clook (UK,
  both send + receive); keep Anthropic (US) as the only outside-UK transfer; the
  contactforsales.com mailbox bracket → leopardess.uk via Clook (forwarding to the
  designconsultancy.co.uk Gmail). Do this before going live.

## CHECKPOINT 2026-06-05 — email: address scheme + catch-all locked; DNS at Clook; framework design written

(Continues the email thread. Records decisions that earlier turns discussed but hadn't
written to file: the address-encoding scheme and catch-all.)

ADDRESS SCHEME (locked): deterministic encoding — lowercase the domain, replace every "."
with "-", append "@leopardess.uk".
  agritec.uk -> agritec-uk@leopardess.uk ; veterinary.co.uk -> veterinary-co-uk@leopardess.uk ;
  idea.uk -> idea-uk@leopardess.uk
- One-way: resolved by MATCHING an incoming address against the encoded forms of known
  domains (the framework has the domain set), NOT by reversing the dashes.
- Why dash not the dotted form (agritec.uk@leopardess.uk): single-label hyphenated local
  parts are accepted by all form validators/clients; multi-dot local parts are RFC-valid
  but some web forms reject them, and customers may type these. Dotted form is a possible
  switch before it's baked in.
- Collision caveat: the rule can collide when one domain's dot sits where another's hyphen
  does (e.g. a-b.uk and a.b.uk both -> a-b-uk). Rare in the current set; the framework must
  detect at assignment (it has the full set) and store a disambiguated address for the
  colliding site. This is why the address is STORED (defaulted to the encoding, overridable)
  — also lets flagships override (idea.uk -> idea@).
- Role variants by suffix where wanted: <encoded>-support@, -info@, -billing@. Company-level
  reserved words (no domain prefix): info@, postmaster@, abuse@ (must resolve).

CATCH-ALL (locked): operator domain leopardess.uk uses a catch-all (cPanel Default Address)
forwarding everything to aaa@designconsultancy.co.uk (Gmail). So <anything>@leopardess.uk
arrives with NO per-site setup; the encoded To: address identifies the site. For sites with
a contact FORM, the app already knows the domain (form posts tagged to the site) — the email
address is just the fallback for people who email directly.

DNS HOME (decided): **Clook hosts leopardess.uk DNS** (cPanel zone). User pasted the live
zone. Status of the pasted zone:
- MX -> mx1/mx2.email-cluster.com (5), failover1 (10): mail to Clook. Good.
- SPF TXT present: v=spf1 +a +mx +ip4:62.182.23.1 include:relay.email-cluster.com ~all —
  authorises Clook's relay for sending as leopardess.uk. Good.
- DMARC present: v=DMARC1; p=none; — monitoring mode. Good to start.
- A leopardess.uk -> 62.182.23.30 (cPanel box); www CNAME. Fine (default cPanel page; add a
  holding page later if wanted). Other records are standard cPanel service entries (harmless).
- **DKIM NOT in the paste.** ACTION: cPanel -> Email Deliverability for leopardess.uk —
  ensure DKIM (default._domainkey TXT) shows installed/green (and SPF green). Needed for
  deliverability + DMARC alignment; cPanel can auto-install it since it controls the zone.

cPANEL NEXT STEPS (Clook): (1) Default Address -> Forward to aaa@designconsultancy.co.uk
(the catch-all) — confirm Clook allows default-address forwarding to an EXTERNAL inbox.
(2) Create ONE mailbox for SMTP auth, e.g. system@leopardess.uk (or info@). (3) Confirm DKIM
via Email Deliverability. (4) Outbound From: if Clook lets the auth mailbox send as any
@leopardess.uk local part, From = the encoded address; if it restricts From to the auth
identity, set From = the one mailbox and Reply-To = the encoded address (catch-all routes
replies back, sorted). Code already supports SMTP_FROM/FROM_NAME/REPLY_TO. CONFIRM with
Clook: external catch-all forwarding allowed? forwarding volume cap? outbound send limit +
policy on automated/SMTP sending? Turn on Clook spam filtering BEFORE the forward (catch-all
attracts spam).

FRAMEWORK DESIGN (written this turn): EMAIL_identity_in_site_spec.md (outputs). The email
choice lives in the framework as TWO layers — (1) global/operator config (operator_domain
=leopardess.uk, provider=clook, SMTP creds, forward_to, the encoding rule) in framework
config; (2) per-site identity as a NEW site_specs ASPECT `email` (no DDL — site_specs is
aspect+jsonb; existing aspects: classification/identity/strategy/design_intent/
content_direction/site_plan/seo/maintenance). Proposed `email` data: status, operator_domain,
address (stored, defaulted to encoding, overridable), from, reply_to, provider, forwards_to,
published, provisioned/provisioned_at, notes. status/provisioned reuse the spec's existing
deployed/planned/blocked + feasibility-recheck machinery — so a FUTURE `email-provisioner`
agent (design only; catch-all makes it unnecessary now) can create per-domain forwarders/
mailboxes (and per-domain DKIM if per-domain sending is wanted) and write status back, same
shape as model-trainer/Thunder. idea.uk standalone: env mirrors what its `email` aspect would
say (SMTP_*->Clook, SMTP_FROM=idea-uk@leopardess.uk, CONTACT/OPERATOR_EMAIL=that). To fold
into chassis: add `email` to 021 aspect list; encoding rule as one shared pure function used
by spec-writing AND the inbound router; operator values in framework config not per-site.

DONE this turn: privacy policy (privacyBody) corrected SES -> **Clook** (UK, sends+receives)
+ **Google (Gmail)** (forwarded inbox); transfers now: Anthropic (US, report content) and
Google (US, forwarded email) are the outside-UK points, Clook UK for email, servers Germany.
build/vet/test clean; privacy_preview.html refreshed (sample idea-uk@leopardess.uk). The
SES-era privacy wording and the contactforsales.com mailbox bracket are gone.

STILL PENDING:
- "by leopardess.uk" branding on idea.uk pages (landing footer + a.page() footer + maybe
  policy pages) — needs user nod on exact wording/placement; pair with next redeploy.
- Confirm designconsultancy.co.uk = Google Workspace? (not needed now — forwarding handles
  it — but affects future options).
- Once Clook mailbox/catch-all live + DKIM confirmed: set env on the box (SMTP_* -> Clook;
  SMTP_FROM=idea-uk@leopardess.uk; SMTP_REPLY_TO if From restricted; CONTACT_EMAIL +
  OPERATOR_EMAIL=idea-uk@leopardess.uk) + redeploy (this turn's privacy change + the page
  branding both need a rebuild).
- arch/deploy doc "Update 2026-06-05" still says SES London for sending — corrected by the
  note appended this turn.

## CHECKPOINT 2026-06-05 — leopardess.uk one-pager built; "by leopardess.uk" added to idea.uk; Clook mail progressing

leopardess.uk SITE (new): leopardess_uk_index.html — one-page identity/holding site.
Refined dark editorial (Cormorant + Hanken Grotesk + Spline Sans Mono; warm near-black
ground, leopard-gold accent, subtle glow/spots/grain, staggered load). Positioned as "an
AI-agent-first development company" — deliberately NO promises (no claims, metrics,
testimonials, client logos, service guarantees, or portfolio). Contact: info@leopardess.uk
(catch-all handles it). Self-contained HTML (inline CSS, Google Fonts) → upload to Clook
public_html for leopardess.uk (its A record → the Clook cPanel box). Products (incl. idea.uk)
intentionally omitted to stay generic.

"by leopardess.uk" branding (DONE): added to idea.uk landing footer (page.html: "idea.uk ·
by leopardess.uk · operated from the UK") and the a.page() wrapper footer (service.go), both
linking https://leopardess.uk. build/vet/test clean; idea_uk_fakedoor.html re-synced. Needs
redeploy to go live.

Clook MAIL progress (user actions):
- Mailbox system@leopardess.uk created (250MB) → the SMTP AUTH account for outbound.
- Catch-all done via domain-level "Forward All Email": leopardess.uk → websy.uk, and
  websy.uk forwards to aaa@designconsultancy.co.uk (user confirmed). So
  <encoded>@leopardess.uk → websy.uk → Gmail; the original To: (encoded domain) is preserved
  in headers through the hops, so it's still sortable in Gmail.
- DELIVERABILITY CHECK NEEDED: the green DKIM/SPF/DMARC/PTR screen the user pasted is for
  leopardess.CO.UK, NOT leopardess.UK. ACTION: in cPanel → Email Deliverability, select
  leopardess.UK and confirm DKIM is valid there (its zone had SPF+DMARC; DKIM on .uk
  unconfirmed). .co.uk being green says nothing about .uk.

ENV to set on the box once leopardess.uk DKIM confirmed (then redeploy — also ships the
privacy fix + by-leopardess branding):
  SMTP_HOST=<Clook outgoing host: cPanel → Email → Connect Devices shows it; likely
            mail.leopardess.uk or rs17.uk-noc.com>
  SMTP_PORT=587
  SMTP_USER=system@leopardess.uk
  SMTP_PASS=<mailbox password>
  SMTP_FROM=idea-uk@leopardess.uk        (idea.uk's encoded operator address)
  SMTP_FROM_NAME=idea.uk
  SMTP_REPLY_TO=idea-uk@leopardess.uk    (set regardless — guarantees replies route)
  CONTACT_EMAIL=idea-uk@leopardess.uk
  OPERATOR_EMAIL=idea-uk@leopardess.uk
CONFIRM with Clook/test: does authenticating as system@ allow sending From idea-uk@ (a
different local part on the same domain)? If yes → From=idea-uk@; if restricted → From=
system@leopardess.uk and rely on Reply-To=idea-uk@ (catch-all routes the reply). Either way
the Reply-To is set, so per-site routing holds.

PENDING: redeploy idea.uk (rebuild amd64 → scp idea.new → mv -f → restart) to ship the
privacy correction + by-leopardess footers + embedded page; upload leopardess_uk_index.html
to Clook for leopardess.uk; confirm leopardess.uk DKIM; confirm the From-vs-ReplyTo question
with Clook.

## CHECKPOINT 2026-06-06 — leopardess.uk email deliverability CONFIRMED (DKIM/SPF/DMARC/PTR pass)

Email Deliverability now shows **leopardess.uk** (correct domain) all green: DKIM, SPF,
DMARC, PTR Valid. Test send system@leopardess.uk → aaa@designconsultancy.co.uk: at Gmail,
SPF PASS (IP 23.83.217.7), DKIM PASS (d=leopardess.uk, s=default), DMARC PASS (p=none),
delivered ~3s. DKIM is ALIGNED to leopardess.uk (header.from domain matches) — good inbox
placement. So OUTBOUND auth via system@ works and is fully authenticated.

Routing note: Clook sends outbound via **MailChannels** (relay.mailchannels.net; the
leopardess.uk SPF include:relay.email-cluster.com authorises it). Normal for cPanel hosts;
helps reputation. Header X-MC-Relay: Bad = brand-new sender with no history, not a block (it
delivered + passed all auth). WARM UP the new domain — don't blast volume on day one;
reputation builds with a little history.

TWO THINGS THE TEST DID NOT COVER:
1. From-as-different-local-part: the test was From=system@ (the auth account itself). The app
   will send From=idea-uk@leopardess.uk while authenticated as system@. Almost certainly
   fine — cPanel Exim lets an authenticated user send as any From on its own domain, and
   DKIM/SPF/DMARC key on the DOMAIN not the local part, so they'll still pass. Set
   From=idea-uk@ + Reply-To=idea-uk@; if the first app send errors "sender not allowed",
   fall back to From=system@leopardess.uk (Reply-To still routes the reply). Optional manual
   test: add idea-uk@ as a Roundcube identity and send.
2. Inbound catch-all not yet tested. Send a test FROM an external address (phone/Gmail) TO
   idea-uk@leopardess.uk; confirm it lands in the Gmail (via websy.uk) with the original
   To: header intact (proves receiving + per-site sorting end to end).

READY NOW: set box env (SMTP_HOST=cPanel "Connect Devices" host e.g. mail.leopardess.uk;
PORT 587; USER=system@leopardess.uk; PASS; FROM=idea-uk@leopardess.uk; FROM_NAME=idea.uk;
REPLY_TO=idea-uk@leopardess.uk; CONTACT_EMAIL+OPERATOR_EMAIL=idea-uk@leopardess.uk) then
REDEPLOY idea.uk (rebuild amd64 → scp idea.new → mv -f → restart) — ships email sending +
the privacy correction + the by-leopardess footers. Then idea.uk can send confirmations/
reports for real. Still pending: upload leopardess_uk_index.html to Clook (leopardess.uk).

## CHECKPOINT 2026-06-06 — inbound test FAILED (No Such User Here): catch-all not catching; fix = specific forwarder

leopardess.uk one-pager is LIVE at leopardess.uk (user confirmed). 

Inbound test to idea-uk@leopardess.uk BOUNCED: "No Such User Here", router=reject, delivery
domain=leopardess.uk — i.e. leopardess.uk's own mail server (Clook/email-cluster) rejected it.
Diagnosis: leopardess.uk is a normal LOCAL mail domain (has the system@ mailbox), so the
server delivers known mailboxes and sends everything else to the domain's DEFAULT ADDRESS,
which is still on cPanel's out-of-the-box "fail / No Such User Here". The "Forward All Email
for a Domain → websy.uk" entry set earlier is NOT intercepting unknown addresses (overridden
by local delivery / not in the routing path). So the assumed catch-all isn't working.

FIX (deterministic, do first): cPanel → Email → Forwarders → Add Forwarder: idea-uk @
leopardess.uk → forward to aaa@designconsultancy.co.uk. A named forwarder to an external
address always works. Re-test idea-uk@leopardess.uk → should land.

CATCH-ALL (long tail): cPanel → Email → Default Address → select leopardess.uk → change from
fail to "Forward to email address" → aaa@designconsultancy.co.uk; and REMOVE the "Forward All
Email for a Domain → websy.uk" entry (not working; removes the websy hop). CAVEAT: some hosts
restrict/discourage forwarding a catch-all to an EXTERNAL inbox (backscatter → IP blocks).
Confirm Clook allows it; if not/uneasy, do NOT use a server catch-all.

DESIGN REFINEMENT (recorded in EMAIL_identity_in_site_spec.md): prefer SPECIFIC per-site
forwarders (created when a site is published) over a server catch-all — only forward
addresses that exist, no backscatter, and it's exactly what the future email-provisioner
agent does (create forwarder, record on the site's `email` aspect). For now idea.uk = one
manual forwarder (above). The deterministic encoding still gives each site's address; the
forwarder is created per published site rather than relying on a wildcard.

(Outbound from leopardess.uk already confirmed working + fully authenticated last checkpoint.)

## CHECKPOINT 2026-06-06 — inbound still bouncing (No Such User); root cause = Default Address not forwarding (+ .uk/.co.uk trap)

"test 5" to idea-uk@leopardess.uk BOUNCED again: chain is Gmail → Clook gateway
(pmg-slave01.email-cluster.com, a Proxmox Mail Gateway) → cPanel box 62.182.23.1, which
said 550 "No Such User Here". So the cPanel backend still rejects idea-uk@ — the catch-all
(Default Address) is NOT yet forwarding. Every "error" the user has seen is THIS bounce
(MAILER-DAEMON non-delivery report with the original embedded) — it exists ONLY because
delivery is failing. Once the address is accepted, the original is delivered intact, no
wrapper. Reassured user: (a) inbound/forwarding does NOT affect sending reputation; (b) the
"DMARC FAIL" they saw is on the bounce notice (null return-path) — normal, irrelevant.

LIKELY TRAP: the user's cPanel screens keep showing leopardess.CO.UK (Spam Detection screen
says @leopardess.co.uk; Default Address dropdown lists leopardess.uk / leopardess.co.uk /
leopardess.co.uk.leopardess.uk). The Default Address must be set on **leopardess.uk**. Also
"Current Setting: leopardess" suggests it may be pointed at the SYSTEM ACCOUNT (would stop
the bounce but keep mail in the cPanel mailbox, not Gmail).

FIX given: Default Address → select domain **leopardess.uk** → "Forward to Email Address" →
aaa@designconsultancy.co.uk → Save. Re-test idea-uk@leopardess.uk; check inbox + spam +
PMG quarantine (panel.email-cluster.com; quarantine at score 6+, test scored ~0). Allow a
few min for the gateway's recipient cache. CLEAN FALLBACK (avoids external forwarding +
backscatter): Default Address → "Forward to your system account", then Gmail → Check mail
from other accounts (POP3) from that mailbox.

Infra learned: Clook = Proxmox Mail Gateway (email-cluster.com) in front + cPanel/Exim
backend (rs17.uk-noc.com / 62.182.23.1); outbound via MailChannels (confirmed working,
authenticated). CONFIRMED: designconsultancy.co.uk is on **Google Workspace** (bounce DKIM
d=designconsultancy-co-uk.20251104.gappssmtp.com) — answers the earlier open question; we
are staying on Clook per user, but Workspace-native is the fallback if Clook stays fiddly.
Forwarders/domain-forwarders now cleared (user removed the idea-uk@ forwarder + websy entry).

Still pending (unchanged): get catch-all delivering to Gmail; then set box env (SMTP_* →
Clook, FROM=idea-uk@leopardess.uk, etc.) + redeploy idea.uk (ships email + privacy fix +
by-leopardess footers). leopardess.uk one-pager is live.

## CHECKPOINT 2026-06-06 — inbound catch-all CONFIRMED; mailer now supports port 465 (implicit TLS) for Clook

INBOUND WORKING: Default Address for leopardess.uk → "Forward to Email Address" →
aaa@designconsultancy.co.uk. A test from an EXTERNAL sender arrived in Gmail. The earlier
"didn't arrive" was Gmail self-send DEDUPE (sending aaa@ → idea-uk@ → forwards back to aaa@;
Gmail hides its own message by Message-ID) — not a delivery failure. So the catch-all is good.
⚠️ The user then flipped the Default Address to system@leopardess.uk (POP route); must set it
BACK to forward-to-aaa@designconsultancy.co.uk for inbound to reach Gmail.

WORKSPACE GATES (parked): designconsultancy.co.uk is Google Workspace, and two features are
disabled at the Workspace ADMIN level for this account: "Check mail from other accounts"
(POP fetch — section absent) and "Send mail as" via external SMTP ("Functionality not
enabled. You must send emails through leopardess.uk SMTP servers..."). So Anthony's personal
replies will go out as aaa@designconsultancy.co.uk, not idea-uk@ (cosmetic only). This does
NOT affect idea.uk's AUTOMATED mail, which the Go service sends via Clook SMTP as idea-uk@.
(Workspace charges per USER not per domain — adding leopardess.uk as a secondary domain +
catch-all into the existing mailbox would be free — but staying on Clook per user.)

SMTP CONFIRMED (cPanel → Connect Devices for system@leopardess.uk): Outgoing
**mail.leopardess.uk**, **SMTP port 465 (SSL/TLS)** only — NO 587 advertised. Username
system@leopardess.uk + its password.

MAILER CHANGE (service.go): port 465 is implicit TLS, which Go's smtp.SendMail does NOT do
(it does STARTTLS). Added smtpSend(host,port,user,pass,from,to,msg): if port=="465" →
tls.Dial + smtp.NewClient + Auth/Mail/Rcpt/Data/Quit (implicit TLS); else smtp.SendMail
(STARTTLS, 587/25). Added crypto/tls import. Backward compatible (host unset → still writes
delivered_*.md). TLS verifies ServerName=host (mail.leopardess.uk); Connect Devices listing
it under "Secure SSL/TLS (Recommended)" implies the cert matches — if a cert error appears on
first send, override ServerName or relax verification (one-liner). idea.env.example email
block updated to Clook + 465 + the leopardess.uk identity. build/vet/test clean.

FINAL ENV for the box (/etc/idea/idea.env):
  SMTP_HOST=mail.leopardess.uk
  SMTP_PORT=465
  SMTP_USER=system@leopardess.uk
  SMTP_PASS=<system@ mailbox password>
  SMTP_FROM=idea-uk@leopardess.uk
  SMTP_FROM_NAME=idea.uk
  SMTP_REPLY_TO=idea-uk@leopardess.uk
  CONTACT_EMAIL=idea-uk@leopardess.uk
  OPERATOR_EMAIL=idea-uk@leopardess.uk
REDEPLOY (code changed → rebuild): GOPROXY=off GOTOOLCHAIN=local GOOS=linux GOARCH=amd64
go build -o idea . ; scp idea root@<box>:/opt/idea/idea.new ; ssh ... 'chmod 755
/opt/idea/idea.new && mv -f /opt/idea/idea.new /opt/idea/idea && systemctl restart idea'.
Box IP last known 116.203.204.115 (user to confirm). Ships email + privacy fix +
by-leopardess footers. AFTER restart: send a test confirmation; verify it arrives and From =
"idea.uk <idea-uk@leopardess.uk>". If send errors "sender not allowed", set
SMTP_FROM=system@leopardess.uk (keep Reply-To=idea-uk@).

## CHECKPOINT 2026-06-06 — laptop build-sync issue; docs updated; handoff written

REDEPLOY BLOCKED on a laptop sync, not a code problem. User's `go build` failed with
`undefined: App / Config / NewApp / writeHTML` (all defined in service.go) plus the
writeHTML refs in audience_check.go and Config/NewApp in main.go. Proven here: canonical
service.go is 658 lines, package main, defines Config(32)/App(44)/NewApp(56)/writeHTML(381)/
smtpSend(478), and builds from a CLEAN cache (BUILD_OK). All 8 .go files are package main.
=> the user's local golang_files/service.go is missing/empty/stale. FIX: re-download the
presented service.go → overwrite golang_files/service.go; verify: `ls -la service.go`,
`head -1 service.go` (package main), `wc -l service.go` (~658), `grep -c "func NewApp"` (1),
`ls -1 *.go` (8: audience_check billing engine main prompts service service_test store). Then
rebuild. WARNING: user ran build/scp/ssh on SEPARATE lines, so the failed build still let
scp/ssh push a STALE old `idea` binary (or error) — the box is NOT running the new code.
Re-run scp+ssh only after a clean build; offered the &&-chained one-liner.

DOCS UPDATED this checkpoint:
- idea_uk_architecture_and_deployment.md: added "Update — 2026-06-06" (email working both
  ways; the 465 implicit-TLS detail + why smtpSend exists; catch-all confirmed; Workspace
  gates; exact env; design note on per-site forwarders). Now 469 lines.
- HANDOFF.md: NEW standalone handoff (current state, the ONE next action, env, redeploy
  commands, email summary, backlog, file map).
- running_notes.md, EMAIL_identity_in_site_spec.md: current.

IMMEDIATE NEXT (unchanged, now in HANDOFF.md): (1) fix service.go in golang_files + clean
build; (2) flip leopardess.uk Default Address back to forward→aaa@designconsultancy.co.uk
(user had left it on system@ POP route); (3) set the env above; (4) redeploy (ships email +
privacy fix + by-leopardess footers); (5) test order → verify it arrives AND From =
"idea.uk <idea-uk@leopardess.uk>". Box IP last known 116.203.204.115 (confirm). Still on
FakeProvider — Stripe live keys remain the separate step to take real money.

## CHECKPOINT 2026-06-06 — two packager scripts (idea.uk go-live, chassis idea-engine)

Built two bundlers in the style of the user's package_page_build_debug.sh:
- **package_idea_uk_golive.sh** — bundles the idea.uk service (golang_files: code +
  page.html + go.mod + deploy/, tests + the `idea` binary excluded) + the go-live docs
  (HANDOFF, running_notes, architecture, email, liability, Stripe plan, runbooks, previews,
  016_v2_32). NO live capture (idea.uk has no DB and isn't on k8s). Env overrides:
  IDEA_ROOT/CODE_DIR/DOCS_ROOT. `--no-debugdoc` drops the big guide. Tested here against the
  outputs: 26 files, 628K (476K lean), no MISSING.
- **package_chassis_idea_engine.sh** — (A) engine to port: engine.go/prompts.go/
  audience_check.go + method docs (method_v0, method_prompt, testrun_v2, PARALLEL,
  CONSOLIDATION); (B) chassis framework: guideline docs (000/001/002/003/019/020/023/009 +
  recent running_notes_16/HANDOFF_2026-06-09), orchestration Go via package_page_build_debug
  paths (coordinator/state/helpers, registry/types/helpers, workflow/branch/basic/generic/
  spawn/call_agent/await actions, datahelpers search+extract, postgres.go, main.go), an
  LLM-action pattern find, schemas (002_intake_orchestrator + best-effort agent_definition_types/
  model-lifecycle/schemas dumps), seed (initial_messages, vet check). Site-merge files
  list_only. Optional read-only LIVE CAPTURE (reuse-discovery): \d agent_definitions/agents +
  existing agent types + a few default_config workflows. `--no-live`, `--no-debugdoc`. Env:
  PROJECT_ROOT/DOCS_ROOT/SQL_ROOT/IDEA_GO_DIR/IDEA_DOCS_DIR.

Both improve on the original: each item is resolved by path then by a repo-wide name search,
and anything unresolved is printed in a **MISSING report** (the original skipped silently).
Fixed a set -e bug (grep -c on zero matches aborted the run) → now uses wc -l guarded by -n.
Note: schemas_all/schemas_some and agent_definition_types.sql may not exist as files in the
repo; if MISSING, the live \d capture gives the authoritative agent/workflow schema.

## CHECKPOINT 2026-06-10 — packager scripts fixed for the real (messy) repo layout

User ran package_idea_uk_golive.sh from the repo root and hit path errors. Root causes +
fixes (both scripts rewritten, syntax-checked, and end-to-end tested against a simulated
messy tree mirroring the user's idea.uk/ dir):

1. **Wrong path default.** IDEA_ROOT defaulted to the script's own dir (= repo root when run
   from there), so CODE_DIR pointed at a non-existent top-level golang_files/. FIX: scripts
   now self-locate the repo root (go.mod) and default to
   $REPO_ROOT/docs/agent_docs/docs024_key_docs_latest/idea.uk (+ /golang_files). Run from the
   repo root with no env needed. Outputs default into the existing
   idea.uk/docubundle_idea_golive (and docubundle_idea_within_chassis) /package_module/output_contexts.
2. **Stale-version trap.** The idea.uk folder has dozens of (N)-versioned copies where the
   UN-suffixed name is the OLDEST (e.g. running_notes.md = 11KB May-27; running_notes(41).md =
   153KB Jun-9). A plain name search grabbed the stale one. FIX: new add_doc() resolves each
   logical doc to the NEWEST variant by mtime, matching both "name.ext" and "name(N).ext".
   Verified: running_notes -> running_notes(41).md, arch -> (3), EMAIL -> (2), 001_development_guide -> (3), etc.
3. **.orig backups + the 9MB binary** would have been swept by the code walk. FIX: write_directory
   now excludes *orig*, *~, *.bak, the `idea` binary, *_test.go, and the usual binaries; NOISE
   prune also skips old_golang_files/ python_files/ _iso/ output_contexts/.
4. **emit() bug** (dynamic scoping): `local path=$1 ... rel="${path#...}"` on one line read
   `path` from the CALLER's scope, so walked files got a blank "filepath = ./" header. FIX:
   compute rel on its own line after path is assigned. Verified headers correct.
5. Dropped refund_policy_preview.html (doesn't exist; terms/privacy previews live in
   golang_files and are captured by the code walk).

Notes for the user: PLAN_stripe_billing_integration.md is NOT in their idea.uk/ folder (it's
elsewhere in the repo; the script searches under docs024 so if it's outside that it'll show
MISSING — copy it into idea.uk/ or set DOC_SEARCH_ROOT). The chassis script's chassis Go paths
(platform/orchestration/…) resolve on the real repo; some SQL/docs may be MISSING if they
don't exist as files (model-lifecycle SQL, schemas dumps) — the live \d capture covers the
agent/workflow schema regardless. Both scripts print a MISSING report; nothing is silently
dropped. (User's idea.uk/ folder is very messy — many duplicate (N) copies — could be cleaned
up later, but the scripts now cope.)

## CHECKPOINT 2026-06-10 — outbound email: 587 not 465 (Hetzner egress) + mailer made async/bounded

Port sweep FROM the box (idea1) to mail.leopardess.uk (62.182.23.30):
  25 blocked · 465 blocked · 2525 blocked · 587 OPEN · 80 open · 443 open
So Hetzner blocks outbound 25/465/2525 but leaves 587 (submission) open — NOT a blanket
SMTP block. CORRECTION: the earlier "use 465 (cPanel Connect Devices advertises 465-only)"
was wrong for this box — 465 is unreachable, the send timed out (dial tcp …:465 connect
timeout, ~2min). Revert to **SMTP_PORT=587** (code takes the smtp.SendMail STARTTLS path for
any port != 465). 587 is also almost certainly what the earlier successful test used.

Mailer structural fix (service.go, deployed):
- makeDeliver now wraps the send in `go func(){…}()` — it was inline on the HTTP request
  path, so the failed 465 connect froze the visitor "thanks" page for ~2min. Now the request
  returns immediately; send result is logged.
- smtpSend 465 path now uses tls.DialWithDialer(&net.Dialer{Timeout:10s}, …) + a 30s
  conn.SetDeadline, so a network problem fails fast instead of hanging. Added "net" import.
- Verified: builds + `go vet` clean (installed Go 1.22 via apt: `apt-get install golang-go`;
  go.dev is not reachable from the sandbox, Ubuntu mirrors are). gofmt nits in the file are
  pre-existing (import order, App/NewApp field alignment), left untouched to keep the diff small.

NEXT: revert env to 587 + `systemctl restart idea`; place a request with own email → operator
/confirm; watch `journalctl -u idea -f` — quiet = sent; a STARTTLS/auth error = report it.

## CHECKPOINT 2026-06-10 — 587 submission WORKS; new blocker = MailChannels relay block

Test (placed a request → handleRequest → operator notification to OPERATOR_EMAIL=idea-uk@leopardess.uk):
- Box→Clook on **587** SUCCEEDED: delivery report shows Authentication=dovecot_plain,
  Sender IP=116.203.204.115 (the box), From idea-uk@leopardess.uk. No "email failed" in the
  journal → smtpSend returned nil. So the 587 + async/timeout mailer is functioning. "Thanks"
  page returned instantly (async fix confirmed).
- BLOCK is one hop later: recipient idea-uk@leopardess.uk is local w/ no mailbox → catch-all
  forwards to aaa@designconsultancy.co.uk → Clook relays via MailChannels (Transport:
  mailchannels_forwarded_smtp) → MailChannels returns **550 5.7.1 [CS] Message blocked**
  (bounce: console.mailchannels.net/insights/bounce?auid=rigxig7uor…). That's MailChannels
  (Clook's outbound relay) rejecting leopardess.uk's FORWARDED outbound — not our code/box/Google.

Likely cause (pending the bounce page): leopardess.uk is days old, no sending reputation; a hard
[CS] reject on a 768-byte plain message reads as reputation, not content. Lever = Clook support
(they own the MailChannels relationship; can check/clear or say what's missing e.g. a domain-
lockdown record). Fallback if a new domain keeps getting blocked via the shared relay: a dedicated
transactional sender — but ask Clook first.

NEXT: (1) open the MailChannels bounce link for the exact [CS] reason. (2) Test the CUSTOMER path
(normal external send, not a forward): request with a real Gmail as buyer → operator /confirm →
watch the delivery report for the message to that Gmail. Forwarded notifications are the hardest
case; the direct buyer send may behave differently.

## 2026-06-10 — docs refreshed for the email findings
Updated to reflect 587-not-465, the async/bounded mailer, and the MailChannels content-filter
block: HANDOFF.md (rewritten to current state + next steps), idea_uk_architecture_and_deployment.md
(new 2026-06-10 update; 06-06 section marked superseded on the port), EMAIL_identity_in_site_spec.md
(operational note on cloud-box ports + relay content-filtering), RUNBOOK_idea_uk.md (status &
deployment note: systemd binary on Hetzner not Docker/S3; email status). Left PLAN_stripe_billing_
integration.md untouched — it's the chassis build/host billing plan, a different product from the
idea.uk £29 report. running_notes is the live journal (this file).

## CHECKPOINT 2026-06-10 (evening) — direct-send test dispatched; awaiting delivery-report result

State of the email test:
- Ran the operator /confirm on the pending order **ord_1781120033520453998** (email
  aaa@designconsultancy.co.uk). Done ON THE BOX reading the key from env:
  `KEY=$(grep '^INTERNAL_API_KEY=' /etc/idea/idea.env | cut -d= -f2)` then
  `curl -s localhost:8080/confirm -H "X-Internal-Key: $KEY" -d '{"order_id":"ord_1781120033520453998"}'`.
  Returned `{"checkout_url":"https://idea.uk/order/success?o=…&fake=1","status":"awaiting_payment"}`
  — so the order advanced and the **buyer confirmation (pay-link) was dispatched** to
  aaa@designconsultancy.co.uk. (fake=1 URL expected on FakeProvider.)
- Gotcha that wasted a round: running the curl on the laptop with the literal placeholders
  `<INTERNAL_API_KEY>` and `"order_id":"ord_..."` returned **`unauthorised`**. Run it on the box
  with the real key from env (above), real order id.
- Order-id lookup command works:
  `ssh root@116.203.204.115 "python3 -c \"import json;o=json.load(open('/var/lib/idea/orders.json'))['orders'];r=sorted([x for x in o.values() if x['status']=='requested'],key=lambda x:x['created_at']);print(r[-1]['id'], r[-1]['email'])\""`

THE DISTINCTION THAT DECIDES IT (forward vs direct):
- The operator "new request" notifications keep getting `550 5.7.1 [CS]` "Blocked (Spam Content)"
  — but those are FORWARDS: idea-uk@leopardess.uk has no mailbox, so the catch-all forwards them
  to aaa@ (Transport: **mailchannels_forwarded_smtp**). Forwarded mail is the hardest case for a
  spam filter. (Another block came in at 20:33, 790 bytes.)
- The buyer confirmation just dispatched is a **direct** send (no forward) — expected Transport
  **mailchannels_smtp**. This is the clean test of whether MailChannels blocks our legitimate
  *direct* outbound or only forwards.

PENDING: read the cPanel delivery report for the 20:3x send to aaa@designconsultancy.co.uk —
need the **Result** (Accepted/delivered vs 550 [CS]) and **Transport** (mailchannels_smtp vs
_forwarded_). Branches:
- **Delivered** → blocks are forwards-only. Customer email is fine (always direct to buyer). Fix
  the operator notification by pointing **OPERATOR_EMAIL straight at aaa@designconsultancy.co.uk**
  (removes the idea-uk@ forward) — that's the real fix; the wording tidy just helps. Email = done.
- **[CS] blocked** → MailChannels rejects direct outbound too → move sending OFF MailChannels
  (ask Clook to disable outbound spam-filtering for the account, or a dedicated transactional
  sender over 443). Stop patching.

CODE: the tidied operator-notification body is in service.go (relabel From:→Requester:, dropped
the raw `POST /confirm {json}`; subject "[idea.uk] New report request <id>"). Builds + vet clean.
NOT confirmed whether the box is running this tidied binary yet — if it IS and the notification
still blocks, that confirms the forward (not the wording) is the cause.

## CHECKPOINT 2026-06-11 — DECISIVE: MailChannels blocks leopardess.uk DIRECT outbound too → must leave MailChannels

The deciding test came back negative. The buyer confirmation (direct send, To aaa@designconsultancy.co.uk,
"Your idea.uk report — confirmed, ready to pay", clean prose + pay-link) was BLOCKED at 09:54:
`550 5.7.1 [CS] Message blocked`, bounce `sender=idea-uk%40leopardess.uk` (NOT the srs0 forward) →
so it's a DIRECT send, not a forward, and not about wording. **MailChannels rejects leopardess.uk's
outbound wholesale as "Spam Content."** Tidying bodies / re-pointing OPERATOR_EMAIL will not fix it.

DECISION: email must come OFF MailChannels. The service speaks plain SMTP, so switching senders is
ENV-ONLY (SMTP_HOST/PORT/USER/PASS/FROM) — no code change. Work is provider + DNS side: authenticate
leopardess.uk (DKIM + SPF) at the new provider, records added to the leopardess.uk zone at Clook.
Routes:
- (1) Quick parallel: ask Clook to exempt the account from MailChannels outbound filtering / whitelist
  leopardess.uk. Free, might clear today, but a shared relay's spam filter is fragile for a txn service.
- (2) Durable: a dedicated transactional sender. SES London (cheapest at volume; already the documented
  swap; needs AWS account + production-access/sandbox-exit) OR a free-tier ESP — Brevo (~300/day) or
  Resend (~3,000/mo), no AWS, quick DKIM (confirm current limits). Box can reach 587 + 443 already.
AWAITING user's choice of route/provider; then give exact DNS records + env.

FORM ABUSE (noted this session): bots are hitting the open /request form — injection probes
(`gYAd'">`, `<'">…`) and spam submissions (e.g. "RobertEnacy", zekisuquc419@gmail.com, Albanian
"kam dashur të di çmimin"). Low risk: the flow is OPERATOR-GATED (no auto-charge, no auto-customer
email; each is just a `requested` order that waits), and the injection strings are inert (stored as
JSON, emailed as plain text). Cost is noise (garbage orders + a blocked notification each). PLAN
(after the email decision): add a honeypot field + basic validation to handleRequest → silently drop
obvious spam (create nothing, send nothing, return the normal thanks page). A "different response
page" is not the fix (bots don't read it; it teaches them what tripped). Also ensure any page that
echoes user input escapes it.

MINOR: the fake=1 success URL didn't load when clicked from the BOUNCE email (the bounce wrapped/
encoded the URL); it loaded fine from a clean address bar (duplicated tab). Test artifact, not a bug —
real flow sends a delivered Stripe checkout link; fake=1 is only the local test shortcut.

## CHECKPOINT 2026-06-11 — engine works end-to-end; idea-uk@ now a real mailbox; trying Clook 1-2 more then SES

GOOD: the fake payment (fake=1 success URL) ran fulfil → engine → produced a full 8-candidate DRAFT
REPORT for "small llm training service company" (Risk column present, per-candidate scores + cheapest
tests, web-verified premises). So the pipeline intake→confirm→pay→fulfil→draft→review is confirmed
working. Only email DELIVERY is broken.

MAILBOX CHANGE: idea-uk@leopardess.uk is now a REAL mailbox (was a catch-all-forwarded address).
Consequence to use deliberately:
- OPERATOR mail (NEW REQUEST notification + REVIEW-with-draft, both To idea-uk@) now delivers
  **locally on Clook** — no forward, no MailChannels, no [CS] block. READ IT IN CLOOK WEBMAIL. Do NOT
  set idea-uk@ to forward onward (a forward is an outbound send → MailChannels → blocked again). So
  the local mailbox fixes the OPERATOR emails only.
- CUSTOMER mail (pay-link, later the report) goes To the buyer's EXTERNAL address → outbound →
  MailChannels regardless → still blocked. The local mailbox does nothing for it. (The 9:54 test
  "customer" was aaa@designconsultancy.co.uk = the operator's own Workspace; a real Gmail hits the
  same wall.)

PLAN (user wants to keep Clook/MailChannels if possible — values the leopardess.uk identity):
1. MailChannels Insights → click "Not Spam" on the blocked confirmation(s).
2. Clook support ticket: legitimate transactional mail from leopardess.uk rejected by MailChannels
   550 5.7.1 [CS] even for a clean 4-line message to one recipient — ask them to check the domain's
   MailChannels reputation / whitelist the account / advise. (They own the relay = the real lever.)
3. Give it ~a day, then re-test a DIRECT send to an external GMAIL (request w/ gmail → confirm → does
   the pay-link arrive?). That is the clean customer-path check, not the aaa@ test.
If still blocked → AWS SES. NOTE: SES sends AS idea-uk@leopardess.uk with leopardess.uk DKIM — same
address the recipient sees; SES is just a more reliable pipe behind the same identity. No loss of
"calibre"; switch is env-only + DKIM/SPF for leopardess.uk added to the Clook DNS zone.

## CHECKPOINT 2026-06-11 — decided AWS SES (London) for sending; local mailbox confirmed for operator mail

Delivery report confirmed: operator notifications to idea-uk@ (12:32, 16:19) = **Accepted** (local
Clook mailbox, no MailChannels). Direct customer send (9:54 aaa@) + forward (10:19) = [CS] blocked.
So MailChannels blocks our outbound; moving sending to SES.

SES PLAN (verified current via AWS docs, June 2026):
- Region London = **eu-west-2**; SMTP `email-smtp.eu-west-2.amazonaws.com`, STARTTLS **587**
  (Hetzner allows 587; SES works on it). **Env-only switch — NO code change** (smtpSend already does
  STARTTLS for any port != 465). From stays idea-uk@leopardess.uk + leopardess.uk DKIM (same identity).
- Steps: (1) AWS acct → SES eu-west-2. (2) Verify domain leopardess.uk, Easy DKIM → add the 3 CNAMEs
  to leopardess.uk zone in cPanel Zone Editor at Clook; wait Verified + DKIM Successful. (3) SES → SMTP
  settings → Create SMTP credentials (IAM user; SMTP user/pass ≠ AWS keys). (4) Request production
  access (Transactional; point at idea.uk; describe honestly). SANDBOX until granted = verified
  recipients only, 200/day, 1/s; approval ~1-2 days. (5) Set env (below) + `systemctl restart idea`.
  (6) Test in sandbox: verify a test Gmail as an identity → request w/ that Gmail → confirm on box →
  pay-link arrives, From idea.uk <idea-uk@leopardess.uk>, DKIM PASS d=leopardess.uk. (7) Live on
  production access.
- ENV on the box:
  SMTP_HOST=email-smtp.eu-west-2.amazonaws.com / SMTP_PORT=587 / SMTP_USER=<SES SMTP user> /
  SMTP_PASS=<SES SMTP pass> / SMTP_FROM=idea-uk@leopardess.uk / SMTP_FROM_NAME=idea.uk /
  SMTP_REPLY_TO=idea-uk@leopardess.uk.
- Operator notifications still land in the idea-uk@ Clook mailbox via SES→leopardess.uk MX. Optional
  later: repoint OPERATOR_EMAIL=aaa@designconsultancy.co.uk for Gmail (must be verified while in sandbox).
- Optional polish: custom MAIL FROM mail.leopardess.uk (MX + SPF TXT) for SPF alignment; not required
  (DKIM alignment passes DMARC).
- leopardess.uk already has SPF + DMARC(p=none); SES Easy DKIM adds the DKIM SES needs. AWS now expects
  SPF/DKIM/DMARC in place before granting production — we have them.

AWAITING: leopardess.uk shows Verified in SES + env set → run sandbox test vs a verified Gmail.

## CHECKPOINT 2026-06-11 — SES live + in PRODUCTION; first send accepted; checking inbox placement

- leopardess.uk DKIM in SES (London) = SUCCESS (domain authenticated for sending).
- SES dashboard shows **50,000/day quota** → account is in PRODUCTION (sandbox is 200/day), so SES
  sends to ANY recipient — the earlier "verify the recipient" sandbox step no longer applies.
- Confirmed ord_1781205749466777546 (email aaa@designconsultancy.co.uk) on the box → returned
  awaiting_payment; SES "Emails sent 1" → **SES ACCEPTED the pay-link**. So the env switch to SES
  worked and the MailChannels [CS] wall is gone. The service now sends through SES.
- BUT the pay-link hasn't appeared in the aaa@ inbox yet. Since SES accepted it, this is now ordinary
  INBOX PLACEMENT (SES→Google), not a relay rejection. Most likely SPAM (brand-new sending domain's
  first message to Google) or a Google-side bounce.
- NEXT CHECKS: aaa@ Spam/All Mail for "Your idea.uk report — confirmed, ready to pay"; SES Account
  dashboard → Reputation/sending stats for bounces/complaints; SES Suppression list for aaa@;
  journalctl -u idea for any "email failed" (shouldn't be, given 1 sent); allow a few minutes. If in
  spam → mark Not spam (normal warmup now that DKIM is proper). If bounced → read the SES reason.

## CORRECTION 2026-06-11 — the SES "1 sent" was NOT the pay-link; SMTP auth is failing (535)

journalctl on the box:
  email to aaa@designconsultancy.co.uk failed: 535 Authentication Credentials Invalid
So the pay-link send FAILED at SES SMTP auth. The dashboard "Emails sent 1" was a different message
(likely a console "send test email"), not our pay-link. Earlier note that "SES accepted the pay-link"
was wrong — the async deliver only logs failures, and the confirm returns awaiting_payment regardless,
so the log is the source of truth, not the confirm response.

DIAGNOSIS: 535 = bad SMTP username/password (NOT an IAM-permissions error). The SMTP_USER/SMTP_PASS in
/etc/idea/idea.env are not valid SES SMTP creds for eu-west-2. Two usual causes:
 (a) IAM access key/secret pasted instead of the SES SMTP Username/Password (the SMTP password is a
     special value SES shows on "Create SMTP credentials" — NOT the IAM secret key);
 (b) SMTP creds generated in the wrong region (SMTP password is region-bound; must be made in
     eu-west-2, the region where the domain is verified).
 Plus watch: truncated password, trailing space, quotes, or an inline comment in the env line.

FIX: SES (region = eu-west-2) → SMTP settings → Create SMTP credentials → use the shown SMTP Username +
SMTP Password → put in idea.env each on its own line, no quotes/inline comment, full password →
systemctl restart idea → re-confirm → journalctl should show no 535. Optional isolation test: swaks
--auth LOGIN against email-smtp.eu-west-2.amazonaws.com:587 (235 = creds OK).
Domain/DKIM/production access are all fine — this is purely the SMTP login.

## ROOT CAUSE 2026-06-11 — SMTP_USER was the IAM user NAME, not the SES SMTP Username

The env had: SMTP_USER=ses-smtp-user.20260611-195505 — that's the auto-generated IAM USER NAME from
"Create SMTP credentials", NOT the SMTP Username. SES authenticates with the SMTP Username = the
ACCESS KEY ID (begins AKIA…), so the IAM user name is rejected → 535. Rest of the env is fine (host
ok, port 587, comments on own lines).
FIX: set SMTP_USER to the AKIA… access key id (from the downloaded SMTP-creds CSV, or IAM → Users →
ses-smtp-user.20260611-195505 → Security credentials → Access key ID). Ensure SMTP_PASS is the SES
SMTP Password (the computed value from the create screen/CSV), NOT the IAM secret access key — if
unsure, regenerate Create SMTP credentials for a matched pair. Own lines, no quotes/inline comment,
full password. systemctl restart idea → re-confirm → journalctl should show no 535.

## CHECKPOINT 2026-06-11 — EMAIL WORKING (AKIA fix); async fulfil fix; review-flow clarified

EMAIL FIXED: the 535 cause was SMTP_USER holding the IAM user name; setting SMTP_USER to the AKIA…
access key id (SMTP Username) fixed it. Pay-link for ord_1781209483578351719 reached aaa@ INBOX (not
spam); headers show DKIM d=leopardess.uk VALID, DMARC_PASS, SPF_PASS. SES now delivers customer mail.
MailChannels→SES saga closed.

CODE (idea-go/service.go) — slow fake-payment page fixed (needs sync+rebuild+redeploy):
- ROOT CAUSE: orderSuccess (fake=1 path) called a.fulfil(id) SYNCHRONOUSLY, so the "Payment received"
  page blocked for the whole engine run. The real Stripe path (webhook) already ran fulfil via
  a.dispatch(...) (goroutine in prod, inline in tests) — the fake path was the only inline caller.
- FIX 1: orderSuccess now does `a.dispatch(func() { a.fulfil(id) })` (reuses existing dispatch; page
  returns immediately; tests stay deterministic since test sets dispatch inline at service_test.go:27).
- FIX 2: added defer/recover at top of fulfil() — once fulfil runs in a goroutine, a panic would crash
  the whole process (net/http only recovers request-path panics). On panic it marks order failed +
  emails operator ([idea.uk] RUN PANIC <id>), matching the RUN FAILED pattern. No var renames.
- Also fixed a PRE-EXISTING broken test: service_test.go reqID helper still searched "NEW REQUEST "
  after the subject was reworded to "New report request " in an earlier session → returned "" →
  nil-panic at line 97. Updated helper string to "New report request " (test-only). vet+build+test green.

REVIEW FLOW clarified for the user: auto_deliver=false, so hitting the pay-link does NOT send anything
to the customer — it simulates payment, runs the engine, and emails the DRAFT to the OPERATOR as
"[idea.uk] REVIEW <id> (<business>)" (lands in idea-uk@ Clook mailbox a minute or two after, once the
engine finishes). Customer only got the pay-link.
GAP FLAGGED (not yet built): there is NO operator "approve & send to customer" endpoint. After review,
the order sits at awaiting_review and nothing forwards it to the buyer — sending is manual right now.
Proposed next: a /deliver operator endpoint (alongside /confirm, /decline) that emails the stored
o.Report to o.Email and marks delivered. Awaiting user go-ahead.

## DESIGN DECISION PENDING 2026-06-11 — re-sequence so money is taken only after operator approves the report

OBSERVATION: ord_1781209483578351719 (small agent framework operator) — the (fake) payment was taken
BEFORE the operator saw the report, and the report came back EMPTY ("No candidate advanced the gate";
Def≥3 AND Will≥3 not met; near-misses Stale-Doc Drift Detector Def2/Will4, Live Log-Grounded Incident
Summariser Def2/Will4; advised change audience/asset/monetisation). Empty result = engine being honest
(niche crowded by Cursor/Claude Projects, low defensibility), NOT a bug.

USER WANT: don't take money until the operator has approved a good report. The previously-floated
/deliver endpoint does NOT solve this (it automates the final send but keeps charge-first) — dropped.

PROPOSED RE-SEQUENCE (awaiting user's explicit go before coding — it's the money flow):
- /request → requested (unchanged).
- /confirm → now RUNS THE ENGINE (not the pay-link): running → draft stored → awaiting_review; operator
  gets the REVIEW email. (Reuse fulfil + dispatch; fulfil already produces draft→awaiting_review+emails
  operator. fulfil guard already allows "running".)
- /approve (NEW operator endpoint): awaiting_review → send pay-link to customer → awaiting_payment.
- /decline (extend to awaiting_review): → declined; optionally email the customer "no charge".
- pay (webhook + orderSuccess fake): awaiting_payment → paid → DELIVER STORED o.Report to customer →
  delivered. (No second engine run; report already vetted.) i.e. move the engine call OUT of the
  payment path (webhook line ~259 / orderSuccess) and INTO /confirm.
- Drop fulfil's auto_deliver=true branch (delivery now happens on payment after approval) — or leave
  unused; decide on build.
- Update TestFlow (currently confirm→webhook→expects delivered) to the new order.

TRADE-OFF flagged to user (their call): re-sequence moves engine cost to BEFORE payment. Customer no
longer commits upfront, so operator absorbs the (few-£) engine cost on empty reports (declined) and on
non-paying pay-link recipients. Operator-controlled (manual confirm). Future refinement once on live
Stripe: card auth-then-capture. AWAITING USER GO-AHEAD on this exact shape.

## CHECKPOINT 2026-06-11 — review-before-pay flag + email UTF-8 fix (needs sync+rebuild+redeploy)

USER DECISIONS: (1) wants a flag to switch pay-first vs pay-after-approval (so they can fall back if
engine cost spikes); during fake-gateway + throttled testing they're happy to absorb engine cost, so
default = review-before-pay. (2) report language/format too dense + provide optional PDF — DEFERRED to
next focused step (spans paid report prompts.go AND free audience-check audience_check.go). (3) email
mojibake (â‰¥ for ≥, â€ for —) — FIXED now.

CODE (idea-go) — done, tested, NEEDS the user to sync + rebuild + redeploy:
- NEW Config flag `ReviewBeforePay` (Config struct + main.go loadConfig). Env `REVIEW_BEFORE_PAY`,
  DEFAULT "true". true = review-before-pay; false = charge-first (current behaviour, untouched).
  Startup log now prints review_before_pay=…
- Flow, REVIEW_BEFORE_PAY=true (Mode B):
  /request → requested → /confirm RUNS THE ENGINE (status running → fulfil → awaiting_review; draft
  emailed to operator, NO pay link) → operator reviews → /approve (NEW endpoint) sends pay link →
  awaiting_payment → customer pays → deliverReport() sends the STORED report → delivered. No 2nd
  engine run. /decline at no charge as before.
- Flow, REVIEW_BEFORE_PAY=false (Mode A): exactly the current charge-first flow (confirm→pay-link→
  pay→fulfil→AutoDeliver decides). Unchanged.
- Reuse: new helper sendPayLink(o) shared by confirm (Mode A) + approve (Mode B); new helper
  deliverReport(id) releases stored report on payment in Mode B. fulfil's delivery branch changed to
  `if AutoDeliver && !ReviewBeforePay { deliver to buyer } else { hold for review }` (so Mode B never
  auto-delivers pre-payment; AutoDeliver only meaningful in Mode A). Payment paths (webhook + fake
  orderSuccess) branch deliverReport (Mode B) vs fulfil (Mode A). New /approve route added.
- TEST: newTestApp now variadic newTestApp(...func(*Config)); added TestReviewBeforePayFlow (confirm→
  awaiting_review+draft, no buyer email; approve→pay link; 2nd approve→409; pay→delivered+report).
  TestFlow unchanged (Mode A). vet+build+all tests green.

EMAIL UTF-8 FIX (service.go makeDeliver): added MIME-Version + Content-Type: text/plain; charset=UTF-8
+ Content-Transfer-Encoding: 8bit; RFC2047-encode Subject and From-name via mime.QEncoding.Encode
(added "mime" import). Fixes the â‰¥/â€ mojibake in all outgoing mail. No env change.

OPERATOR WORKFLOW NOW (Mode B), on the box:
  KEY=$(grep '^INTERNAL_API_KEY=' /etc/idea/idea.env | cut -d= -f2)
  confirm (starts engine): curl -s localhost:8080/confirm -H "X-Internal-Key: $KEY" -H 'content-type: application/json' -d '{"order_id":"ord_..."}'
  → wait for REVIEW email with draft → read it →
  approve (bills buyer):   curl -s localhost:8080/approve  -H "X-Internal-Key: $KEY" -H 'content-type: application/json' -d '{"order_id":"ord_..."}'
  or decline (no charge):  curl -s localhost:8080/decline  -H "X-Internal-Key: $KEY" -H 'content-type: application/json' -d '{"order_id":"ord_...","reason":"..."}'
ENV: add REVIEW_BEFORE_PAY=true (or =false to revert). New binary defaults to true if unset.

gofmt: only PRE-EXISTING nits in service.go (blank-import order, App/NewApp alignment); my edits added
none. Left as-is per earlier decision; can gofmt -w if wanted.

NEXT (deferred, agreed): rewrite report + audience-check language/format (less dense, nicer layout);
add optional PDF of the report (email attachment or link). The UTF-8 fix already lands the mojibake.

## SESSION DECISIONS LOG — 2026-06-11 (consolidated; detail in the checkpoints above)

Choices and decisions made this session, in one place:

1. EMAIL SENDER → AWS SES. Tried once more to keep Clook/MailChannels (operator wanted the
   leopardess.uk identity); confirmed MailChannels blocks leopardess.uk outbound wholesale
   (550 [CS], even a clean message). Decision: move sending to AWS SES, London eu-west-2, STARTTLS
   587. From stays idea-uk@leopardess.uk + leopardess.uk DKIM (same identity, better pipe). SES is
   in production. DONE + verified (pay-link in real inbox; DKIM/DMARC/SPF pass).
   - Sub-decision: idea-uk@ made a real Clook mailbox so operator notifications deliver locally (no
     MailChannels); read in Clook webmail; never forward idea-uk@ onward.
   - Gotcha recorded: SMTP_USER must be the SES SMTP Username (AKIA… access key id), NOT the IAM
     user name (ses-smtp-user.…) — that was the 535.

2. PAYMENT ORDER → review-before-pay, behind a flag. Operator was uncomfortable taking money before
   seeing the (sometimes empty) report. Decision: add REVIEW_BEFORE_PAY (default true): /confirm
   runs the engine → operator reviews the draft → /approve bills the buyer → payment delivers the
   stored report; /decline = no charge. REVIEW_BEFORE_PAY=false keeps the old charge-first flow as a
   fallback if engine cost spikes. Trade-off accepted (engine cost moves before payment; fine while
   on fake gateway + throttled; operator can revert). DONE + tested (both flows).

3. STRUCTURAL FIXES this session: slow fake-payment page fixed (route fulfil through the existing
   dispatch so the page returns immediately; added panic-recovery to fulfil since it now always runs
   in a goroutine). Email mojibake fixed (UTF-8 Content-Type + RFC2047 Subject/From-name). Stale
   test helper fixed ("NEW REQUEST" → "New report request").

4. DEFERRED (agreed, next focused pass): rewrite report + audience-check language/format (too dense)
   and add an optional PDF of the report. Form-abuse honeypot on /request (low risk, operator-gated).
   Stripe live mode remains THE earn step.

DOCS UPDATED this session: HANDOFF.md (where-it-stands one-pager) rebrought current — SES email,
review-before-pay flow + operator commands, env block, facts, backlog. Dated SES-supersession notes
prepended to idea_uk_architecture_and_deployment.md and EMAIL_identity_in_site_spec.md so their older
MailChannels/465 sections aren't followed by mistake.

## CHECKPOINT 2026-06-11 — report readability pass (layout + prose prompts); needs sync+rebuild

Started the deferred report-quality work (Stripe still waits on the operator's keys/own-card test).
DIAGNOSIS: the report is markdown but emailed as text/plain, so #, **, * showed as literal clutter;
and the "findings" / "cheapest_test" prose had no length/plain-language guidance, so the model wrote
long, vendor-name-dumped paragraphs. Both made it read dense.

CODE (idea-go) — done, tested, needs sync+rebuild+redeploy:
- engine.go render() rewritten to clean PLAIN TEXT (reads well unstyled in an email): "IDEA REPORT —
  <business>" + rule, WHO IT'S FOR / WHY THEY'D PAY, "ADVANCING IDEAS (best first)" with each idea as
  "N) Title [flag]" then aligned labels Idea / Built on / Checks out / Scores (Defensibility X/5,…,
  total N/25) / Risk / First test; then "DIDN'T MAKE THE CUT" and "SET ASIDE ON RISK" sections.
  Plain-English section names; no markdown symbols. **Engine logic/scoring/gating UNCHANGED — layout
  only.** Added const reportRule.
- prompts.go: verifyPrompt now tells the model to write "findings" as 1-2 short plain sentences, no
  vendor name-strings unless one name is the key evidence; scorePrompt caps cheapest_test at 1-2 plain
  sentences. (Style steers; effect seen on the next real engine run — can't verify output offline.)
- service_test.go: added TestRenderReadable (prints the layout + guards against markdown clutter
  returning). vet+build+all tests green.

NOTE the report's `domain` field is actually the business descriptor (not a domain name) — render
shows it after "IDEA REPORT —".

STILL OPEN on report quality (next step):
- Audience-check FREE taster (audience_check.go) is also dense — reformat next, same spirit.
- "Nicer format / optional PDF" decision: report is plain text now. Options — (1) HTML email
  (multipart/alternative: clean text + styled HTML), PURE Go stdlib, works with the offline
  GOPROXY=off build, gives real headings/spacing; (2) a true PDF file needs a PDF library in the Go
  service, which the stdlib-only offline build can't pull without relaxing it (vendor a module) — or
  render HTML and print-to-PDF. RECOMMEND HTML email; await user's choice.

## CHECKPOINT 2026-06-11 — Stripe ready (test keys received), taster logging added, mobile bug, sample shown

STRIPE: billing.go is complete + correct (Checkout via API with metadata[order_id]; HMAC webhook-sig
verify; no SDK). makeProvider switches to StripeProvider only when BOTH STRIPE_SECRET_KEY and
STRIPE_WEBHOOK_SECRET are set (else FakeProvider, which logs "No Stripe keys"). No code change needed.
User supplied TEST keys (sk_test_…/pk_test_…) — NOT stored anywhere by me; user sets them on the box.
GO-LIVE (test mode), operator steps:
 - Stripe Dashboard (TEST mode) → Developers → Webhooks → Add endpoint → URL
   https://idea.uk/stripe/webhook → event checkout.session.completed → copy the Signing secret whsec_…
 - /etc/idea/idea.env: STRIPE_SECRET_KEY=sk_test_… , STRIPE_WEBHOOK_SECRET=whsec_… → systemctl restart
   idea. (pk_test_ publishable key is NOT used server-side — Checkout is hosted; nothing to set.)
   Confirm the startup log NO LONGER says "No Stripe keys"; if it does, one env var is blank.
 - Test: request → /confirm (engine) → review → /approve (now makes a REAL Stripe checkout link) → pay
   with test card 4242 4242 4242 4242 (any future date/CVC/postcode) → Stripe webhook → order paid →
   report delivered. Webhook must be reachable at https://idea.uk/stripe/webhook (nginx proxies all);
   check delivery in the Stripe dashboard (200 = good).
 - SECURITY: user pasted the sk_test_ in chat — low risk (test key, no real money) but advised not to
   paste secrets; can roll in Dashboard. LIVE keys later must be box-env-only, never in chat.

LOGGING (answered the user):
 - FREE taster (audience_check.go) had NO usage logging — only an in-memory rate limiter. ADDED
   `log.Printf("free taster: business=%q audience=%q", …)` on each successful run → countable in
   journalctl (`journalctl -u idea | grep "free taster" | wc -l`). Added "log" import.
 - PAID tasks ARE recorded: the generated report is stored per order in /var/lib/idea/orders.json
   (Order.Report), plus status/email/timestamps. So full record of paid tasks + their reports exists.
   (Could add a per-run engine log line later if wanted; the JSON is the record.)

MOBILE BUG (logged + fixed): text too close to phone edges. Logged as a BUG in new BUGS_idea_uk.md.
Fix: side padding → 24px + safe-area-aware (max(24px, env(safe-area-inset-*))) in page.html mobile
media query AND the a.page() wrapper in service.go. Needs rebuild+redeploy; user to confirm on phone;
if still tight, identify the exact page.

HTML EMAIL: user approved ("sounds fine for now"). DEFERRED to next focused step (it's the substantial
build: an HTML renderer for the report + a multipart/alternative deliver path keeping the plain-text
fallback). Plain-text report already reads cleanly after the render rewrite.

SAMPLE: the new plain-text report layout was only in a test-log the user can't see; pasted it into the
reply this turn for their review/tweak.

FILES this turn: audience_check.go (taster log), service.go (a.page() mobile padding), page.html
(mobile padding), BUGS_idea_uk.md (new). vet+build+tests green.

## CHECKPOINT 2026-06-11 (b) — HTML email BUILT; taster now logs the RESULT; runbook updated; PDF noted

(Supersedes the "HTML EMAIL ... DEFERRED" line in the previous checkpoint — it's now built.)

TASTER RESULT LOGGING (corrected an assumption): the v1 log line logged only the visitor's INPUTS
(business + stated audience), NOT the result. Told the user honestly (they'd assumed it logged the
result). FIXED: the taster now logs the result too — carried audience, willingness, and the
alternatives — so the results are reviewable in journalctl (`grep "free taster"`). Paid reports were
already stored in orders.json (report + report_html).

HTML EMAIL — BUILT (structural, not a text re-parse):
 - engine.go: EngineFunc + RunMethod now return `renderedReport{Text, HTML}` (was a bare string).
   New `renderReport(...)` builds BOTH renderings from the same structured data; new `renderHTML(...)`
   emits inline-styled, email-client-safe HTML (brand palette; all dynamic text escaped). All 9 error
   returns → `renderedReport{}`, all 3 render returns → `renderReport(...)`. Added "html" import.
 - store.go: Order gained `ReportHTML` (json `report_html,omitempty`).
 - service.go: mailer refactored to `sendOne(...)` (plain OR multipart/alternative: text + HTML),
   with `makeDeliver` (plain) + `makeDeliverHTML`; App gained `deliverHTML`, wired in NewApp. `fulfil`
   stores both text+HTML and sends the report/REVIEW as multipart HTML; `deliverReport` sends stored
   text+HTML (plain fallback if no HTML). `internalRun` returns report + report_html. main.go CLI uses
   rep.Text.
 - service_test.go: engine stub returns the struct; deliverHTML stub records the text part (so asserts
   still pass); TestRenderReadable also renders HTML, asserts its structure, and writes a viewable
   sample to sample_report_email.html (write is a no-op if the outputs dir is absent → still passes on
   the box). gofmt -w on the touched files (also tidied the pre-existing alignment nits). vet+build+
   tests green (TestFlow 19 checks, TestReviewBeforePayFlow, TestRenderReadable, internal all pass).

DOCS: RUNBOOK_idea_uk.md got a 2026-06-11 operating-status section (new flow + operator commands, SES,
Stripe go-live, taster/paid logging, HTML email, PDF intent) → 159 lines. HANDOFF backlog item 2 split:
language/layout = DONE (render rewrite + HTML email); PDF promoted to its own item with the rationale
(makes the £29 tangible) + the stdlib-only/offline constraint (vendor a PDF lib or render HTML→PDF).

FILES this turn: engine.go, store.go, service.go, main.go, service_test.go (HTML email);
audience_check.go (result logging); RUNBOOK_idea_uk.md + HANDOFF.md (docs); sample_report_email.html
(regenerated, viewable). Sync the .go files + rebuild + redeploy to apply. HTML email needs no env
change; SES already sends multipart fine.

## CHECKPOINT 2026-06-11 (c) — report copy rewritten to plain English + HTML email redesigned

Acting on the user's review of the report email (it read too technical / LLM-speak / dense /
abbreviated; missing context; rejected sections too terse; design too "Claude-ish"). Two layers:
TEMPLATE (my code — labels, structure, summary, design) and CONTENT (prompts for real runs; stub for
the sample). Fixed both.

PROMPTS (prompts.go) — for real reports:
 - systemBase: added a global plain-language WRITING STYLE rule (plain everyday English, short
   sentences, no jargon/acronyms/buzzwords, spell terms out, describe concretely with an example).
 - generatePrompt: asset/capability/reason must be plain + concrete with an everyday example
   ("scanned floor plans your customers email you", not "customer drawings").
 - audiencePrompt: plain language, spell out specialist terms ("staff who build software or AI
   systems", never "ML staff").

RENDER (engine.go) — both renderers:
 - Added a plain SUMMARY at the top (reportIntro: what idea.uk is, what they asked for, what's in the
   report) and a plain FOOTER (reportFooter: what idea.uk does + "reply to this email").
 - Plain labels: Idea→(lead sentence, no label); Built on→"What it's built on"; Checks out→"What we
   found"; Scores→"How it scored" with PLAIN factor names (hard to copy / people will pay / easy to
   build / reusable elsewhere / built to last); First test→"A cheap first test"; Risk→"Risk to you".
   Flag shown in plain words ("worth considering"/"worth testing now") via flagLabel.
 - Heading "ADVANCING IDEAS" → "IDEAS WORTH PURSUING" (both renderers, consistent).
 - EXPANDED rejected sections: each dropped/risk idea now shows a plain description (BeatsFreeBecause,
   which the embedded candidate carries) + a plain reason (dropReason from the scores; risk reason for
   the risk-dropped). New helpers: reportIntro, reportFooter, flagLabel, dropReason.
 - riskNote: replaced "PII"/"T&Cs" with plain wording.
 - renderHTML REDESIGNED with its own professional palette/type (NOT the landing-page brand): deep
   navy #15243d headings, restrained gold #b08a3e accent (rule/badge/card edge), Georgia serif
   headings over Helvetica/Arial body, white "sheet" on #eceff3, idea cards, generous spacing.

SAMPLE (service_test.go TestRenderReadable): rewrote the placeholder content into plain English
(domain "a small firm that builds AI tools for solicitors"; spelled-out audience; concrete asset
descriptions; added BeatsFreeBecause to the dropped/risk stubs). Updated assertions to the new
headings/labels (IDEAS WORTH PURSUING, "A cheap first test:", "This report is from idea.uk"; HTML
asserts "Your idea report" + "Ideas worth pursuing"). sample_report_email.html regenerated. gofmt'd;
vet+build+tests green.

BUGS: logged the feedback in BUGS_idea_uk.md, categorised (copy / orientation / completeness / design)
with what was fixed + a "for future builds" principle each, so future report builds start better.

NOTE: real-report copy can only be confirmed on a live engine run; the sample uses placeholder text.
After the next real run, re-read actual model output and tighten prompts again if jargon slips through.

FILES this turn: prompts.go, engine.go (render+renderHTML+helpers+riskNote), service_test.go (stub +
asserts); BUGS_idea_uk.md + this note. sample_report_email.html regenerated. Sync the .go files +
rebuild to apply (prompts change only affects real runs; no env change).

## CHECKPOINT 2026-06-11 (d) — riskNote plain-up + report CTA added

- "refunds make customers whole" was OURS (a hardcoded string in riskNote, engine.go — not the model).
  Risk-4 note reworded to plain English: "(low — a mistake would be minor, and a refund would put it
  right)". (Another instance of the copy-clarity theme already in BUGS.)
- Added a bottom-of-email CTA (both renderers): "If you'd like us to help turn any of these ideas into
  a working tool — or you have any questions about this report — just email us at <contact>." In HTML
  it's a styled gold-tinted box with a mailto link, above a muted footer line. New helpers:
  reportContact() (reads CONTACT_EMAIL env, falls back to idea-uk@leopardess.uk) and reportCTA();
  reportFooter() trimmed (the CTA now covers "questions", so the duplicate "reply to this email" line
  was dropped).
- EMAIL ADDRESS NOTE: the user typed idea_uk@ (underscore); I used idea-uk@ (HYPHEN) — that's the live
  mailbox and the CONTACT_EMAIL/SMTP_FROM value. Flagged to the user to confirm. The CTA tracks
  CONTACT_EMAIL on the box, so it shows whatever that env is set to.
- gofmt'd; vet+build+tests green; sample_report_email.html regenerated. Sync engine.go + rebuild.

## CHECKPOINT 2026-06-11 (e) — sample off-domain item fixed + generate-prompt domain guard

- User spotted a MEDICAL item ("suggest a likely diagnosis…") in the "set aside on risk" section of a
  report about a SOLICITORS firm. It was MY stub placeholder (service_test.go riskDropped), NOT model
  output. Replaced with an on-theme legal example ("Give the public legal advice on their own case").
- Answered the prompt question: "medical/financial" DO appear in prompts.go — line 26 (capability
  menu, "medical images" as one input-type example) and the score step (risk examples: medical/legal/
  financial advice). They're scoring/illustration, not idea seeds; the generate step is audience-
  anchored, so off-domain drift on a real run is unlikely.
- SAFEGUARD added to generatePrompt: an explicit line that every candidate must be for THIS audience in
  THIS domain, and cross-sector mentions elsewhere are illustration only (not a suggestion to leave the
  domain). Cheap insurance against the exact drift the user worried about.
- Note: the render explanation still reads "things like medical, legal, or financial advice" — that's
  reader-facing only (never fed back to the model) and is a fair general description of regulated
  territory, so left as is.
- gofmt'd; vet+build+tests green; sample regenerated. Files: service_test.go (stub), prompts.go
  (generate guard). Logged in BUGS under a new "consistency / off-domain" theme.

## CHECKPOINT 2026-06-11 (f) — click-through operator links (Confirm/Approve/Decline in the email)

User confirmed the live review-before-pay flow works (request email → curl /confirm → engine ran).
Asked to make the operator step clickable from the email (a per-order auth id), or failing that, clear
instructions. Built the clickable version (the right call), securely:

DESIGN: a per-order capability token = HMAC-SHA256(INTERNAL_API_KEY, "op:"+orderID), base64url[:24] —
no storage, unguessable, authorises that ONE order. Emails contain a link
PUBLIC_BASE_URL/op?o=<id>&t=<token>.
 - Why a button, not a bare GET that acts: mail scanners / clients PRE-FETCH links. So /op is a
   SIDE-EFFECT-FREE GET that renders the order + status-appropriate buttons; the action fires only on a
   button POST carrying the token. A prefetch can't trigger an (expensive) engine run.
 - Token doubles as CSRF protection on the POST. Actions still gated by status (can't confirm twice).
 - Blast radius if a token leaks: one order (confirm=engine cost; approve=buyer gets a pay link they'd
   get anyway; decline=one order killed). No data exposure, no unauthorised charge. Acceptable; the
   operator inbox is trusted.

CODE (service.go): new helpers orderToken/opLink/parseOp/opAuthorised/wantsHTML/opRespond. confirm/
approve/decline now accept EITHER the X-Internal-Key header (curl → JSON, unchanged) OR a valid token
(browser → HTML result page) via opAuthorised + content-negotiated opRespond. New opPage GET handler at
/op (route added) renders details + buttons by status: requested→Confirm/Decline; awaiting_review→
Approve/Decline; running/awaiting_payment/delivered/declined→status only. Links wired into the request
email (handleRequest) and the review email (fulfil, text + HTML). Imports added: crypto/hmac,
crypto/sha256, encoding/base64.

TEST: TestOperatorLink — valid-token /op shows the Confirm button and is a no-op on status; bad-token
/op = 404; token POST /confirm advances the order (→awaiting_payment) and returns 200; wrong-token
POST = 401. Existing TestFlow/TestReviewBeforePayFlow unchanged (header+JSON path still returns JSON
because they send no Accept: text/html). gofmt clean; vet+build+all tests green.

DOCS: runbook 2026-06-11 section got a "Click-through (easiest)" note above the curl block (curl kept as
fallback). FILES: service.go (+ service_test.go test); RUNBOOK_idea_uk.md. Sync service.go + rebuild.
No env change (PUBLIC_BASE_URL already set; token uses INTERNAL_API_KEY).

## CHECKPOINT 2026-06-11 (g) — click-through confirmed working by user; OPEN: draft email not received

USER TEST: received the request email WITH the link, clicked through to the new page, clicked Confirm,
and got the correct "report is being generated / status: running" page. So the click-through link +
token + page + confirm all work end to end. BUT the draft report email did not arrive.

DIAGNOSIS SO FAR (not assuming a cause): after Confirm, fulfil runs in the background (engine minutes →
store draft → email review to OPERATOR_EMAIL). Found that the engine logs NOTHING and fulfil only
logged panics — so the run was invisible in journalctl, which is why this is hard to see.

OBSERVABILITY FIX (this turn): added log lines to fulfil — "running engine", "engine error: …",
"engine done (N chars)", "delivering report to buyer"/"draft ready, emailing review to <addr>" — and a
SUCCESS log to the mailer sendOne ("email to <addr> sent" vs the existing "… failed: <err>"). A re-test
will now show exactly where it stops. gofmt+vet+build+tests green.

DIAGNOSTIC PLAN given to user (for ord_1781349188999996431): (1) check status + whether report/
report_html are stored in orders.json; (2) grep old logs for panic / "email to … failed"; (3) redeploy
with the new logging and re-confirm a test order while tailing journalctl. Hypotheses: engine still
running / restarted mid-run; OR engine finished but the NEW multipart HTML draft email failed at SES
(plain request email works → suspect the new multipart path); OR engine errored (RUN FAILED email →
check spam). The draft is stored in orders.json even if the email failed, so it isn't lost.

QUESTIONS to user: how long after clicking did you check (engine takes minutes)? did you restart/
redeploy the service after confirming (kills in-flight runs)? which inbox are you checking and did you
check spam (draft goes to OPERATOR_EMAIL = idea-uk@leopardess.uk, the same inbox the request arrived in)?

DOCS: BUGS_idea_uk.md OPEN entry; RUNBOOK_idea_uk.md troubleshooting section. FILE: service.go (logging).
Sync service.go + rebuild to get the logging before re-testing.

## CHECKPOINT 2026-06-11 (h) — diagnosis: engine OK + draft SAVED; it's an email-delivery question

### What happens in the background after Confirm (review-before-pay) — the flow
1. **/confirm** (operator, via the token link or the X-Internal-Key header): checks the order is
   `requested` and there's capacity; sets status `running`; calls `a.dispatch(fulfil(id))`. In
   production `a.dispatch` runs fulfil in a GOROUTINE, so the HTTP response returns immediately (that's
   the "report is being generated" page the operator sees). (In tests `dispatch` runs inline.)
2. **fulfil(id)** (background goroutine), with a defer/recover that marks the order `failed` + emails a
   RUN PANIC notice if it panics:
   a. re-reads the order; only proceeds if status is `paid` or `running`.
   b. sets status `running`; logs `fulfil: <id> running engine`.
   c. calls `a.engine` (RunMethod): the method — audience → generate → cut → verify (web search) →
      score → rank. Real Anthropic/OpenAI + web_search calls; takes MINUTES; spends money. Returns
      `renderedReport{Text, HTML}`.
   d. on engine error: logs `engine error: …`, sets status `failed`, emails operator RUN FAILED.
   e. on success: logs `engine done (N chars)`; stores `o.Report` (plain) + `o.ReportHTML`; sets status
      `awaiting_review` (review-before-pay) — or, in charge-first auto-deliver, `delivered` + sends to
      the buyer.
   f. emails the DRAFT (the REVIEW email) to OPERATOR_EMAIL via `a.deliverHTML` (multipart: plain text +
      HTML), including the op link to approve/decline. `a.deliverHTML` → `go sendOne(...)` — ANOTHER
      fire-and-forget goroutine; as of the 2026-06-11 logging it logs `email to <addr> sent` or
      `email to <addr> failed: <err>`.
3. Operator reads the draft → op link → **/approve** (sends the buyer the pay link) or **/decline**.
4. Buyer pays → Stripe webhook → `deliverReport` sends the stored report to the buyer → `delivered`.
KEY POINTS: the engine runs in the background (the response returns at once); the DRAFT goes to
OPERATOR_EMAIL, NOT the requester; everything is persisted in orders.json, so a generated draft is
never lost even if its email fails; the email sends are fire-and-forget goroutines.

### Diagnosis result (order ord_1781349188999996431)
- `status: awaiting_review | report stored: True | html: True` → the engine SUCCEEDED and the draft is
  SAVED on the box. Not lost.
- The log grep returned NOTHING. The build that ran this order logged only email FAILURES (not
  successes), so "no lines" means the send did NOT log a failure. So: not an engine problem — an email
  DELIVERY question.
- LEADING explanation (to confirm): the draft (REVIEW) email is a large multipart HTML message with
  links, whereas the request email that DID arrive is a tiny plain-text one — so the draft is far more
  likely to have been filtered into SPAM/junk, or accepted by SES but not surfaced in the inbox.
- IMMEDIATE: (1) check the OPERATOR_EMAIL inbox's spam/junk (idea-uk@leopardess.uk, the same inbox the
  request email arrived in); (2) the stored draft can be read straight from orders.json right now
  (`['orders'][id]['report']`).
- CERTAINTY: redeploy with the (already-added) send logging and re-confirm a test order while tailing
  `journalctl -u idea -f` — we'll see `email to <addr> sent` (then it's pure deliverability/spam) or
  `failed: <err>` (then we fix the send). No code change needed first; logging is already in service.go.
