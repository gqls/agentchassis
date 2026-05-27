# PLAN — idea.uk (working name): hosted AI ideation tool

**Status: aspirational goal, needs work.** The goal for this thread is to host the
ideation tool on **idea.uk** (working name). Build the **internal** version first
(it generates ideas for our own domains), then host a version for others if the
defensibility holds up. The moat is genuinely uncertain and is discussed openly
below — we may not make it, but that is the target.

Companion to `PLAN_simple_paid_multidomain_chat.md` (§10 differentiator
framework, §11 ideation agent) and `FOCUS_site_chatbot_edge_worker_and_context_
pack.md` (the worker/paywall idea.uk would reuse).

---

## 1. The goal

Turn the §10 differentiator method into a real tool, and host it on idea.uk as a
paid product that helps a business find AI opportunities worth building. Get
there via an internal-only version first, so we prove the output is good before
exposing it.

---

## 2. What the tool actually is (plain language)

It is the §10 method written precisely enough to run: take an **audience**, a list
of **assets** that audience's business has, and a list of **AI capabilities worth
using right now**, cross them to find combinations that do something a free tool
cannot, then verify, score, and rank the candidates. §10 describes this in prose;
§4 below names the inputs, steps, and output so it can be run the same way every
time — first by hand, later by an agent.

**Two front-ends, one shared method.** This matters from the start:

- **Internal version** — input is *our* domain + *our* asset list (our build
  pipeline output, any feeds we buy, our partnerships). Its value is that it knows
  our specific assets.
- **Hosted version (idea.uk)** — input is a *stranger's* business; it knows
  nothing about our assets and uses only what the user tells it about theirs, plus
  our maintained capability list.

The method in the middle is the same. The inputs, the data it may see, and who it
serves differ. So the rule from day one: **keep our asset list as input data, never
built into the method.** If our assets are passed in rather than hardcoded, the
hosted version is just the same method called with a different (user-supplied)
asset list. If they are wired into the logic, we would have to untangle them later
exactly when we want to ship idea.uk — the same separability discipline as the
cluster work, in a new place.

**The loop.** idea.uk is itself an instance of the paid multi-domain chat
(`PLAN_simple_paid_multidomain_chat.md`): a chat domain whose payable
differentiator *is* the ideation tool. So building idea.uk reuses that plan's
worker, paywall, and pass mechanism — idea.uk is one more configured domain, with
the ideation method as its bound tool. And we can run the tool on itself to look
for its own differentiators (see §5, dogfooding).

---

## 3. How we got here (discussion to date)

- **§10 differentiator framework.** The AI model is not the differentiator;
  everyone has the same models. What is hard to reproduce is the specific asset we
  apply AI to — proprietary data, an owned process/output, a well-built tool, a
  partnership, or being first to a new capability. A payable idea is asset × AI,
  for people who will pay.
- **§11 ideation agent.** An internal, low-risk use of the agent framework that
  runs the cross-product and proposes scored candidates; re-runnable whenever a
  new capability is added.
- **Internal vs hosted split.** Identified this session — the same method serves
  two different users; keep assets as data so both can share the core.
- **Defensibility caveat.** By the framework's own test, a hosted "AI finds AI
  opportunities for you" tool is a low-static-moat, chat-is-the-product case (the
  harder kind), because a capable user could prompt a free model to do something
  similar. That is why §5 exists.
- **Pricing note.** An ideation result for a business owner is worth more than the
  £1–2 of a light chat — possibly a higher one-off (a "report"). idea.uk likely
  sits at a different price point than the light domains; flagged in §7.

---

## 4. The method, written to run

Relating directly to §10 — this is the part I earlier called an "engine spec";
it is just the method made concrete.

- **Inputs:** the audience (and how much they can/will pay); the asset list
  (passed in — ours for internal, the user's for hosted); the current
  AI-capability list (the maintained watchlist, see §5).
- **Steps:**
  1. For each relevant (asset × capability) pair, draft a candidate and a one-line
     reason it beats a free tool.
  2. Critique each candidate (a separate pass, ideally a different model) to drop
     the generic or non-defensible ones.
  3. Verify the survivors with web research — does the data feed exist and at what
     cost; do competitors already do it; is the audience's willingness to pay
     real.
  4. Score on four factors: how hard the asset is to reproduce, willingness to
     pay, build cost, reuse across domains.
  5. Rank, and split into "test now (cheap)" and "score/consider (expensive)".
- **Output:** a ranked candidate list, each naming the asset it depends on, the AI
  capability it uses, the one-line reason it beats a free tool, the four scores,
  and the cheap-test it suggests.

This can be run by hand as a structured prompt sequence before any agent is built.

---

## 5. The moat — competing with a good prompt in a commercial LLM

The honest starting position: a single good prompt into a frontier model with web
search is already strong, and **search grounding is not a moat** — the commercial
models do it well. So the question is what a multiagent + multi-model + web-
research orchestration genuinely adds. Sorted by how hard each is to reproduce.

**Reproducible by a skilled prompter (weak on their own):**

- *Decomposition.* Breaking the job into research-audience, list-assets,
  scan-competitors, generate, critique, score is more thorough than one prompt —
  but a determined expert can do it manually across several prompts. We package
  the process; we don't own it.
- *Verification.* Checking claims with search is better than asserting them, but
  again a careful user can do it by hand. Laborious, not exclusive.

**Harder to reproduce (where the real defensibility is):**

1. **Currency — a maintained capability watchlist that beats the model's
   self-knowledge.** Models are poor at knowing their own newest features (training
   cutoffs lag, and they don't reliably know what was shipped this month). A list
   we keep current — what the latest models, image/video/voice generators, and
   tools can newly do — is something the base model does not have about itself.
   Re-running ideation when a new capability lands is the practical mechanism for
   being early adopters of ideas the capability just unlocked, before users know
   it exists. This is the single strongest durable advantage, and it is exactly
   the early-adopter point raised. Its cost is constant maintenance.
2. **Multi-model, multi-modal ensemble.** Generating candidates from several LLMs
   (different training, different idea distributions) and cross-critiquing
   produces a broader, less generic set than any one model. And some candidates
   can only be explored by orchestrating capabilities a single chat surface does
   not combine — e.g. generate an idea, then actually produce a sample image to
   judge whether an image-tool idea is any good. A user pasting one prompt into one
   model cannot do this; assembling it by hand is real effort.
3. **Verification with evidence attached.** Beyond search-grounding, the
   orchestration can check whether a proposed feed exists and its price, scan
   competitor offerings, and attach the evidence to each candidate — so the output
   is checked candidates, not plausible-sounding ones. Partially reproducible,
   laborious by hand.
4. **Memory across runs.** Accumulating which ideas were generated, tested, and
   converted improves scoring over time and avoids re-suggesting failures. A single
   prompt has no memory across sessions. A slow-building advantage, modest at
   first.
5. **The build bridge — unique to us.** We already operate a build pipeline. An
   idea from idea.uk can be taken straight to a cheap test (a demand-test page) or
   a working prototype, because we can build it. A prompt-paster gets prose; we can
   get them a tested thing. This connects idea.uk to the whole platform and is the
   advantage hardest for anyone without a build pipeline to copy.

**Honest verdict.** The strongest, most durable parts are **currency (1),
verification (3), and the build bridge (5)**, with the **multi-model ensemble (2)**
as a quality multiplier and **memory (4)** building slowly. But all of these are a
**process, freshness, and integration** advantage — kept alive by continuous
effort — not a static asset nobody else has. So idea.uk survives only as long as
we keep it current and keep it connected to building; it is the harder kind of
product from §10, and that realism should stay front of mind. The framing that is
defensible is not "a better idea chatbot" but "ideas that come verified, kept
current with capabilities the base models don't yet know about, and that we can
immediately prototype or build for you."

**Dogfooding as the first test.** Run the tool on itself. If it cannot surface a
compelling, defensible differentiator for idea.uk, that is direct evidence the
tool is not good enough yet. It is also the cheapest possible first run.

---

## 6. What it reuses

- **The paid multi-domain chat** (`PLAN_simple_paid_multidomain_chat.md`) — idea.uk
  is one configured chat domain; the worker, paywall, and pass mechanism are the
  same. The bound "tool" for this domain is the ideation method.
- **The multiagent framework** — the internal ideation agent (§11) and sub-agents
  for research/verification.
- **Web research tools** — for verification and for maintaining the capability
  watchlist.
- **Multi-model orchestration** — the ensemble in §5.2; reuse the existing LLM-call
  routing, extended to call several providers.
- **The build pipeline** — the action bridge in §5.5: turning an idea into a cheap
  test or prototype.

---

## 7. Sequence

1. **Write the method (§4) precisely**, parameterised so assets are input data,
   not built-in logic.
2. **Run it by hand** against two or three real domains *and* on idea.uk itself
   (dogfood). Judge whether the candidates are genuinely good.
3. **Refine the scoring** based on those runs — this is what decides if the output
   is worth anything.
4. **Build the internal agent (§11)** only if the hand-run output is good.
5. **Host on idea.uk** as a paid chat domain only if the moat (§5) holds up in
   practice — and decide its price point then (likely a higher one-off than the
   light domains, §3).

---

## 8. Open questions / risks

- **Is maintained currency enough?** The base models keep improving and may close
  the capability-knowledge gap; the watchlist advantage needs constant work to
  stay ahead.
- **Maintenance cost** of the capability watchlist and the multi-model ensemble —
  this is the ongoing price of the moat.
- **Hosted defensibility** is low on static assets (§5); the product may not clear
  its own bar. Accept this as a real possibility.
- **Pricing** — idea.uk is probably a higher-value one-off (a verified report +
  prototype) rather than a £2 day-pass; different price point, possibly different
  payment shape, than the light chat domains.
- **Verification depth vs cost** — deeper verification is better but costs more
  model and research calls per run; find the level that keeps a run affordable.
- **Scope creep into the build product** — the build bridge (§5.5) is powerful but
  blurs idea.uk into the build-as-a-service plan; keep the boundary deliberate.
