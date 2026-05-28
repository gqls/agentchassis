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
