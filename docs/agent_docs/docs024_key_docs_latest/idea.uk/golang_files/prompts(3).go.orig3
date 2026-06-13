package main

// prompts.go — the method, encoded as per-step prompts. {placeholder} tokens are
// substituted by fill() (strings.NewReplacer). JSON examples use single braces
// (unlike the Python .format version which doubled them).

const systemBase = "You run a disciplined ideation method that finds payable AI product ideas. " +
	"Core principle: the AI model is NOT the differentiator (everyone has the same " +
	"models); the hard-to-reproduce ASSET it's applied to is. A payable idea is " +
	"one asset x one AI capability, for an audience that will pay, doing something " +
	"a free model with a good prompt cannot. Be rigorous and honest; refusing to " +
	"advance a weak idea is a correct outcome, not a failure. " +
	"WRITING STYLE: everything you write is read by a busy, non-technical business " +
	"owner. Use plain, everyday English and short sentences. No jargon, no acronyms " +
	"unless you spell them out in full, and no marketing or technical buzzwords " +
	"(avoid words like 'leverage', 'robust', 'schema', 'pipeline', 'tuned', " +
	"'structured output'). When you name something, say concretely what it actually " +
	"is and does, with a brief everyday example where it helps a layperson picture it."

const capabilityMenu = `- Knowledge & retrieval: curated/grounded RAG over the right sources; long-context
  (load entire user corpora or a regulatory body of rules at once); persistent
  cross-session memory. Beats generalists on stale or shallow knowledge.
- Reasoning & computation: domain-tuned reasoning; ACTUAL computation (call a
  solver/calculator/simulator, not LLM approximation); multi-step workflows with
  checkpoints + verification. Beats generalists on confident-wrong technical output.
- Multi-modal input: technical image understanding (engineering drawings, medical
  images, maps, forms); voice/audio; sensor/data feeds. Beats generalists on
  domain-specific inputs they weren't tuned for.
- Multi-modal output: precise image editing/in-painting; video generation; voice;
  schema-guaranteed structured data. Beats generalists on output fidelity.
- Action-taking: agentic browsing / computer use; integration with the user's
  stack (calendar, CRM, accounting, devices, gov portals); workflow execution.
  Beats DESCRIPTION-ONLY generalists — does the thing rather than telling you how.
- Coordination: multi-model ensembles (cheap-fast for some steps, deep-slow for
  others); specialised sub-models per task. Beats single-model chats on quality-at-price.
- Personalisation & continuity: cross-session memory; user-profile tuning;
  ongoing-project context. Beats stateless chats for coaching/case-management/projects.
- Quality & safety: domain validation; source-grounded outputs with citations;
  uncertainty signalling; refusal/escalation. Beats generalists in regulated/sensitive
  domains where confident-wrong is harmful.
- FRONTIER (watchlist — keep current): agentic browsing/computer-use reliability;
  million-token context; reasoning-model pricing; real-time voice agents; video
  generation usefulness; precise image editing. Ask for each: is there a product for
  THIS audience that wasn't possible 18 months ago?`

const audiencePrompt = `STEP 1 — FRAME THE AUDIENCE (and challenge it).

Domain: {domain}
Stated audience: {audience}
Assets available (input data, may be sparse): {assets}

Do two things:
1. State the audience and their willingness to pay in two sentences.
2. CHALLENGE the audience: is this the right audience, or is there a better-fit
   audience for this domain that would pay more, or that sits on a softer free
   substitute? List up to 3 alternative audiences with a one-line reason each.
   Flag which audience you'll carry forward (it may not be the stated one).

Write in plain language a layperson understands; spell out any specialist term in
full (for example write "staff who build software or AI systems", never "ML staff").

Return JSON only:
{"carried_audience": "...", "willingness_to_pay": "...",
  "alternatives": [{"audience": "...", "why": "..."}]}`

const generatePrompt = `STEP 2 — GENERATE CANDIDATES across FOUR LENSES plus an asset x capability sweep.
Do NOT do a single pass. Run every lens. Aim for 3-6 candidates per lens
(12-24 total). Be willing to propose non-obvious ideas.

Domain: {domain}
Audience (carried forward): {audience}
Willingness to pay: {wtp}
Assets: {assets}

Capability menu (cross these with assets in the sweep; pull frontier items in lens c):
{capabilities}

Lenses:
a. DEMAND — what does this audience deeply want, struggle with, or pay for today?
   What workflow takes longest? What error-prone task do they re-do? What expertise
   is unevenly distributed? What do they pay specialists for that's mostly pattern-
   following? What can't they get done for lack of a piece of expertise?
b. GENERALIST-FAILURE — where does a generalist LLM fail this audience? (stale/wrong
   on specifics; confident-wrong on detail; generic where domain-specific matters;
   can't act, only describe; forgets between sessions; can't compute precisely;
   can't access live/proprietary data). Each "yes" is a candidate seed.
c. FRONTIER — what just became possible/cheap in the last 6-12 months (from the
   watchlist) that enables a NEW product for this audience?
d. OUTCOME — what's the DREAM result (not "help me with X" but "X is done,
   correctly, ready to use")? Reverse-engineer a product that delivers it.
e. SWEEP — cross the asset list with the capability menu for obvious combinations
   the lenses missed.

For each candidate give: a title, the lens it came from, the asset it depends on,
the capability it uses, and a one-line reason it beats a free model with a good prompt.
Write the title, asset, capability and reason in plain words a non-technical owner
understands: say what each thing actually is and does, with a concrete everyday
example where useful (e.g. "the scanned floor plans and wiring diagrams your
customers email you", not "customer drawings"; "an AI that reads those images and
returns the measurements as a tidy table", not "vision model with structured
output"). Avoid technical or product-spec phrasing.

Return JSON only:
{"candidates": [{"title": "...", "lens": "...", "asset": "...",
  "capability": "...", "beats_free_because": "..."}]}`

const cutPrompt = `STEP 3 — CUT. You are a DIFFERENT model from the generator. Be ruthless; most
candidates should die here.

For EACH candidate, do two checks:
1. SPECIFIC FREE SUBSTITUTE — name the concrete free thing this audience would
   actually do instead (e.g. "describe the idea to Bolt themselves", "call the
   supplier's free application engineer", "use the manufacturer's free app",
   "ask Perplexity"). If that gets them most of the way, DROP it.
2. SELLER-BUNDLES-SUPPORT — does the seller of the underlying product already give
   this support away free as part of their sales process? (For high-margin products
   they usually do.) If so, DROP unless the candidate clearly improves on the
   seller's offer. NOTE: sometimes a conflicted free incumbent (e.g. commission-paid
   brokers) is an OPPORTUNITY, not a barrier — flag that case as KEEP.

Candidates (each has an "id" — echo the SAME id in your result for each):
{candidates_json}

Return JSON only:
{"results": [{"id": "...", "title": "...", "free_substitute": "...",
  "verdict": "keep"|"drop", "reason": "..."}]}`

const verifyPrompt = `STEP 4 — VERIFY the surviving candidates with web search. Assert nothing you can
check. For each candidate, check the claims its premise rests on:
 - does the data feed / partnership / tool actually exist, and roughly what does it cost?
 - do competitors already offer this?
 - is the willingness-to-pay real (evidence, not assumption)?
Use the web_search tool. Then drop any candidate whose premise fails verification.

Write each "findings" in plain English a non-technical reader understands: one or two
short sentences saying what you checked and what you found. Do not list strings of
product or vendor names — name one only if that single name is the key evidence.

Domain: {domain}
Surviving candidates (each has an "id" — echo the SAME id for each):
{survivors_json}

After searching, return JSON only (no prose outside the JSON):
{"results": [{"id": "...", "title": "...",
  "findings": "<1-2 plain sentences: what you checked and what you found>", "premise_holds": true|false}]}`

const scorePrompt = `STEP 5 — SCORE each candidate whose premise held. Six factors, 1-5, where 5 is
ALWAYS more attractive / safer. Score honestly; guessing high defeats the point.

- Defensibility: 1 = a free model with a good prompt does this; 3 = needs our
  process/curation, a determined expert could copy with effort; 5 = depends on an
  asset others can't get (exclusive data, held partnership, genuinely hard tool).
- Willingness to pay: 1 = expect free / won't pay; 3 = some pay a little; 5 =
  budget + clear repeated pain, pays readily.
- Buildability: 1 = major bespoke build; 3 = moderate, mostly assembling what we
  have; 5 = trivial or already produced.
- Reuse across domains: 1 = bespoke to one domain; 3 = a few similar; 5 = many.
- Durability (resistance to base-model improvement): 1 = next release erodes it /
  substitute improving fast; 3 = holds a while, needs refresh; 5 = rests on
  something model progress doesn't erode.
- Risk to the operator (us, if we build and operate it): score the CONSEQUENCE
  of being wrong, not the chance of being wrong. 5 = pure analysis, customer
  makes their own decisions, no plausible loss beyond the fee we charged; 4 =
  minor downstream consequences possible, a refund would make the customer
  whole; 3 = meaningful financial/operational decisions ride on our output,
  customer can verify our citations, PII insurance is recommended; 2 =
  high-stakes decisions (real money, regulated or quasi-regulated matters,
  legal/medical adjacency), requires human review every report + PII +
  carefully reviewed T&Cs before building; 1 = regulated profession territory
  (medical advice, legal advice, FCA-regulated financial advice) — should NOT
  be built without proper qualifications, regardless of how attractive it looks
  on the other factors.

  Use the asymmetry test: if our output is wrong, what could it cost the
  customer? Time wasted = high score. Lost grant money / missed regulatory
  window / wrong medical decision = low score.

  Risk is NOT added to the sum. The sum is fitness; Risk is hazard. They are
  reported separately. Risk = 1 gets dropped automatically downstream; Risk ≤ 2
  gets flagged as "needs liability work before building."

GATE: a candidate ADVANCES only if Defensibility >= 3 AND Willingness >= 3.
(Risk is NOT a hard gate — Risk = 1 is dropped in post-processing; Risk = 2 still
advances but flagged.)

Flag any advancing candidate with Durability <= 2 as "short_lived": true.

For each advancing candidate, give the cheapest demand test. IMPORTANT: if Risk
<= 2, the cheapest_test MUST explicitly say "validate demand first; do not
build until PII insurance is in force and T&Cs are reviewed by a UK solicitor."
For Risk >= 3 candidates the cheapest_test can be a normal fake-door / landing
page / etc. Keep the cheapest_test to one or two plain, concrete sentences.

Mark "test_now" if Buildability >= 4 OR the demand test is cheap AND Risk >= 3;
else "consider".

Candidates with findings (each has an "id" — echo the SAME id for each):
{verified_json}

Return JSON only:
{"scored": [{"id": "...", "title": "...", "defensibility": n, "willingness": n,
  "buildability": n, "reuse": n, "durability": n, "risk": n, "sum": n,
  "advances": true|false, "short_lived": true|false,
  "cheapest_test": "...", "flag": "test_now"|"consider"|"failed"}]}`
