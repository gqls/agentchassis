# idea.uk method — single-shot prompt

This is the whole method collapsed into one prompt. Paste it into any capable
model **with web search enabled** (Claude, ChatGPT, Gemini), fill in the three
inputs at the top, and you'll get output similar to the by-hand runs. It's
"similar, not identical" — LLM output varies between runs, and a single-shot
prompt is weaker than the staged multi-model script (`idea_method_runner.py`),
because one model does generation, critique, and scoring in a single pass rather
than a different model critiquing the generator. For a stronger result, run the
staged script, or at least run steps 3 (cut) and 4 (verify) again in a *fresh*
chat / different model so the critique isn't the generator marking its own work.

────────────────────────────────────────────────────────────────────────────
PASTE FROM HERE
────────────────────────────────────────────────────────────────────────────

You are running a disciplined ideation method that finds payable AI product
ideas. Core principle: the AI model is NOT the differentiator — everyone has the
same models, including the customer. The hard-to-reproduce ASSET it's applied to
is the differentiator. A payable idea is one asset × one AI capability, aimed at
an audience that will pay, doing something a free model with a good prompt
cannot. Be rigorous and honest: refusing to advance a weak idea is a correct
outcome, not a failure. Use web search to verify claims — assert nothing you can
check.

INPUTS
  Domain:   <<< e.g. agritec.uk >>>
  Audience: <<< e.g. UK small farmers, 3–50ha, many eligible for SFI26 Window 1 >>>
  Assets:   <<< the hard-to-reproduce things this business has or could acquire;
                may be sparse, e.g. "we operate the site; can curate scheme docs;
                no proprietary data feed" >>>

CAPABILITY MENU (cross with assets; pull "frontier" items in lens c)
  - Knowledge & retrieval: curated/grounded RAG; long-context (load whole corpora
    or rule-sets); persistent memory. Beats generalists on stale/shallow knowledge.
  - Reasoning & computation: domain-tuned reasoning; ACTUAL computation (call a
    solver, not LLM approximation); multi-step workflows with verification. Beats
    confident-wrong technical output.
  - Multi-modal input: technical images (drawings, medical, maps, forms); voice;
    sensors/feeds. Beats generalists on inputs they weren't tuned for.
  - Multi-modal output: precise image editing; video; voice; schema-guaranteed data.
  - Action-taking: agentic browsing/computer-use; stack integration (CRM, accounting,
    gov portals); workflow execution. Beats description-only generalists.
  - Coordination: multi-model ensembles; specialised sub-models. Quality-at-price.
  - Personalisation & continuity: cross-session memory; profile tuning; project context.
  - Quality & safety: domain validation; source-grounding; uncertainty signalling.
  - FRONTIER (keep current): agentic browsing reliability; million-token context;
    reasoning-model pricing; real-time voice agents; video gen; precise image edit.
    For each: is there a product for THIS audience that wasn't possible 18 months ago?

PROCEDURE — do every step in order. Show your working for each.

STEP 1 — FRAME & CHALLENGE THE AUDIENCE.
  State the audience and willingness to pay in two sentences. Then CHALLENGE it:
  is this the right audience, or is there a better-fit one for this domain that
  would pay more or sits on a softer free substitute? List up to 3 alternatives;
  pick the one to carry forward (may not be the stated one).

STEP 2 — GENERATE across FOUR LENSES + a sweep. Not one pass. 3–6 per lens:
  a. DEMAND — what does the audience deeply want / struggle with / pay specialists
     for / can't get done for lack of expertise?
  b. GENERALIST-FAILURE — where does a generalist LLM fail them (stale/wrong on
     specifics; confident-wrong; generic; can't act; forgets; can't compute; can't
     access live/proprietary data)? Each "yes" is a seed.
  c. FRONTIER — what just became possible/cheap (last 6–12 months) that enables a
     NEW product for them?
  d. OUTCOME — what's the dream result ("X is done, ready to use", not "help me
     with X")? Reverse-engineer a product to deliver it.
  e. SWEEP — cross assets × capability menu for obvious combos the lenses missed.
  For each candidate: title, lens, asset it depends on, capability it uses, and a
  one-line reason it beats a free model with a good prompt.

STEP 3 — CUT (be ruthless; most should die). For each candidate:
  (i) name the SPECIFIC free substitute the audience would actually use (e.g.
      "describe it to Bolt themselves", "call the supplier's free engineer", "use
      the maker's free app", "ask Perplexity"). If that gets them most of the way,
      DROP it.
  (ii) does the seller of the underlying product already give this support away
      free as part of selling it? (High-margin products usually do.) If so, DROP
      unless the candidate clearly beats the seller's offer. EXCEPTION: a
      conflicted free incumbent (e.g. commission-paid brokers) can be an
      OPPORTUNITY — keep and note it.

STEP 4 — VERIFY survivors WITH WEB SEARCH. For each: does the feed/partnership/tool
  exist and what does it cost; do competitors already do it; is willingness-to-pay
  real (evidence, not assumption)? Attach findings. Drop any whose premise fails.

STEP 5 — SCORE survivors, 1–5 each (5 always better):
  - Defensibility: 1 free-model-does-it · 3 needs our process/curation · 5 exclusive asset.
  - Willingness to pay: 1 won't pay · 3 some pay a little · 5 budget + repeated pain.
  - Buildability: 1 major bespoke build · 3 moderate · 5 trivial/already have it.
  - Reuse across domains: 1 bespoke · 3 a few · 5 many.
  - Durability: 1 next model release erodes it · 3 holds a while · 5 model progress
    doesn't erode it.
  GATE: advance only if Defensibility ≥3 AND Willingness ≥3. Flag Durability ≤2 as
  "short-lived". For each advancing candidate give the cheapest demand test and
  mark "test now" (Buildability ≥4 or cheap test) vs "consider".

STEP 6 — OUTPUT. Rank advancing candidates by score sum. For each: title, idea
  (capability × asset → what it does), the verified findings, the five scores +
  sum, the cheapest test, and test-now/consider. Then list what was dropped and
  why. If nothing advances, say so plainly and give the route to a better audience,
  a new asset, or a different monetisation.

────────────────────────────────────────────────────────────────────────────
PASTE TO HERE
────────────────────────────────────────────────────────────────────────────

## Notes on reproducing the by-hand runs

- The runs in `idea_uk_testrun_v2.md` used this procedure but with me doing the
  searching and judgement, not a single paste. So pasting this gets you the same
  *shape* of output and usually similar verdicts, but exact candidates and scores
  will differ.
- The biggest fidelity gap in single-shot mode is STEP 3 (the cut). When one model
  generates and critiques in the same pass, it tends to be too kind to its own
  candidates. To get the by-hand behaviour, copy the candidate list into a fresh
  chat (ideally a different model) and run STEP 3 there before continuing. The
  staged script does this automatically.
- STEP 4 needs real web search. Without it, "verification" is just the model
  asserting again — which is exactly what the method exists to avoid.
