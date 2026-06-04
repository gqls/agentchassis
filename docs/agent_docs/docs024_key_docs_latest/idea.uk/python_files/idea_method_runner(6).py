#!/usr/bin/env python3
"""
idea_method_runner.py
─────────────────────
Standalone runner for the idea.uk ideation method (v2).

This is the script version of the method described in `idea_uk_method_v0.md`.
It reproduces the *procedure* used in the by-hand runs (`idea_uk_testrun_v0.md`,
`idea_uk_testrun_v2.md`): frame the audience, generate across four lenses, cut
against the specific free substitute (with a DIFFERENT model), verify survivors
with web search, score on five factors, rank and split.

IMPORTANT — what this is and isn't:
  • The by-hand runs were not produced by this script. This is a faithful
    reconstruction of the same steps, so output will be SIMILAR, not identical.
  • LLM output is non-deterministic. Two runs will differ. That is expected and
    is why the method leans on the cut + verification steps rather than the
    generator's first guess.
  • This is v0 of the runner. Treat its scores as a starting point for human
    judgement, not as ground truth.

Setup:
  pip install anthropic
  export ANTHROPIC_API_KEY=sk-ant-...

Run:
  python idea_method_runner.py \
      --domain "agritec.uk" \
      --audience "UK small farmers (3-50ha), many eligible for SFI26 Window 1" \
      --assets "operate the site; can curate UK scheme docs; no proprietary data" \
      --out report_agritec.md

  (Or edit the EXAMPLE_INPUTS block and run with no args.)
"""

import os
import re
import sys
import json
import argparse

try:
    from anthropic import Anthropic
except ImportError:
    sys.exit("Missing dependency. Run: pip install anthropic")


# ─────────────────────────────────────────────────────────────────────────────
# CONFIG — which model runs which step.
# The cut step deliberately uses a DIFFERENT model from the generator so the
# method doesn't mark its own work (this is the "[diff-model]" note in the method).
# Same-family models (Opus vs Sonnet) are weaker diversity than cross-vendor.
# For stronger diversity, replace the cut step with an OpenAI/Gemini call — see
# the call_other_vendor() stub at the bottom.
# Model strings verified against docs.claude.com (May 2026).
# ─────────────────────────────────────────────────────────────────────────────
GEN_MODEL      = "claude-opus-4-7"            # generation — broad, creative
CRITIQUE_MODEL = "claude-sonnet-4-6"          # cut — different Anthropic model (fallback)
VERIFY_MODEL   = "claude-opus-4-7"            # verification — runs web search
SCORE_MODEL    = "claude-sonnet-4-6"          # scoring

# Cross-vendor critique: if OPENAI_API_KEY is set, the cut step runs on OpenAI
# instead of Anthropic — genuine cross-vendor diversity (the one multi-model
# claim that was previously untested). Set the model to a current one; defaults
# are overridable because vendor model names change.
OPENAI_CRITIQUE_MODEL = os.environ.get("OPENAI_CRITIQUE_MODEL", "gpt-4o")

WEB_SEARCH_TOOL = {"type": "web_search_20250305", "name": "web_search", "max_uses": 6}

client = Anthropic()  # reads ANTHROPIC_API_KEY from env


# ─────────────────────────────────────────────────────────────────────────────
# THE CAPABILITY MENU (v2) — grouped by what specialism does that generalists
# don't. The "frontier" entries are the watchlist; keep them current.
# ─────────────────────────────────────────────────────────────────────────────
CAPABILITY_MENU = """\
- Knowledge & retrieval: curated/grounded RAG over the right sources; long-context
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
  THIS audience that wasn't possible 18 months ago?
"""


# ─────────────────────────────────────────────────────────────────────────────
# PROMPTS — these ARE the method, encoded per step. They are the "equivalent
# prompt" you would paste into an LLM. (A single-shot consolidated version is in
# idea_method_prompt.md.)
# ─────────────────────────────────────────────────────────────────────────────

SYSTEM_BASE = (
    "You run a disciplined ideation method that finds payable AI product ideas. "
    "Core principle: the AI model is NOT the differentiator (everyone has the same "
    "models); the hard-to-reproduce ASSET it's applied to is. A payable idea is "
    "one asset x one AI capability, for an audience that will pay, doing something "
    "a free model with a good prompt cannot. Be rigorous and honest; refusing to "
    "advance a weak idea is a correct outcome, not a failure."
)

AUDIENCE_PROMPT = """\
STEP 1 — FRAME THE AUDIENCE (and challenge it).

Domain: {domain}
Stated audience: {audience}
Assets available (input data, may be sparse): {assets}

Do two things:
1. State the audience and their willingness to pay in two sentences.
2. CHALLENGE the audience: is this the right audience, or is there a better-fit
   audience for this domain that would pay more, or that sits on a softer free
   substitute? List up to 3 alternative audiences with a one-line reason each.
   Flag which audience you'll carry forward (it may not be the stated one).

Return JSON only:
{{"carried_audience": "...", "willingness_to_pay": "...",
  "alternatives": [{{"audience": "...", "why": "..."}}]}}
"""

GENERATE_PROMPT = """\
STEP 2 — GENERATE CANDIDATES across FOUR LENSES plus an asset x capability sweep.
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

Return JSON only:
{{"candidates": [{{"title": "...", "lens": "...", "asset": "...",
  "capability": "...", "beats_free_because": "..."}}]}}
"""

CUT_PROMPT = """\
STEP 3 — CUT. You are a DIFFERENT model from the generator. Be ruthless; most
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
{{"results": [{{"id": "...", "title": "...", "free_substitute": "...",
  "verdict": "keep"|"drop", "reason": "..."}}]}}
"""

VERIFY_PROMPT = """\
STEP 4 — VERIFY the surviving candidates with web search. Assert nothing you can
check. For each candidate, check the claims its premise rests on:
 - does the data feed / partnership / tool actually exist, and roughly what does it cost?
 - do competitors already offer this?
 - is the willingness-to-pay real (evidence, not assumption)?
Use the web_search tool. Then drop any candidate whose premise fails verification.

Domain: {domain}
Surviving candidates (each has an "id" — echo the SAME id for each):
{survivors_json}

After searching, return JSON only (no prose outside the JSON):
{{"results": [{{"id": "...", "title": "...",
  "findings": "<what you verified, with specifics>", "premise_holds": true|false}}]}}
"""

SCORE_PROMPT = """\
STEP 5 — SCORE each candidate whose premise held. Six factors, 1-5, where 5 is
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
  reported separately. Risk = 1 gets dropped automatically downstream; Risk <= 2
  gets flagged as "needs liability work before building."

GATE: a candidate ADVANCES only if Defensibility >= 3 AND Willingness >= 3.
(Risk is NOT a hard gate — Risk = 1 is dropped in post-processing; Risk = 2 still
advances but flagged.)

Flag any advancing candidate with Durability <= 2 as "short_lived": true.

For each advancing candidate, give the cheapest demand test. IMPORTANT: if Risk
<= 2, the cheapest_test MUST explicitly say "validate demand first; do not
build until PII insurance is in force and T&Cs are reviewed by a UK solicitor."
For Risk >= 3 candidates the cheapest_test can be a normal fake-door / landing
page / etc.

Mark "test_now" if Buildability >= 4 OR the demand test is cheap AND Risk >= 3;
else "consider".

Candidates with findings (each has an "id" — echo the SAME id for each):
{verified_json}

Return JSON only:
{{"scored": [{{"id": "...", "title": "...", "defensibility": n, "willingness": n,
  "buildability": n, "reuse": n, "durability": n, "risk": n, "sum": n,
  "advances": true|false, "short_lived": true|false,
  "cheapest_test": "...", "flag": "test_now"|"consider"|"failed"}}]}}
"""


# ─────────────────────────────────────────────────────────────────────────────
# HELPERS
# ─────────────────────────────────────────────────────────────────────────────
def call(model, user, system=SYSTEM_BASE, tools=None, max_tokens=4096):
    """One Anthropic Messages API call; returns concatenated text blocks.
    With the server-side web_search tool, the model loops internally and we
    just read the final text."""
    kwargs = dict(model=model, max_tokens=max_tokens, system=system,
                  messages=[{"role": "user", "content": user}])
    if tools:
        kwargs["tools"] = tools
    resp = client.messages.create(**kwargs)
    return "".join(b.text for b in resp.content if getattr(b, "type", None) == "text")


def parse_json(text):
    """Pull the first JSON object/array out of a model response, tolerating
    ```json fences and surrounding prose."""
    text = text.strip()
    text = re.sub(r"^```(?:json)?", "", text).strip()
    text = re.sub(r"```$", "", text).strip()
    # find the outermost {...} or [...]
    start = min((i for i in (text.find("{"), text.find("[")) if i != -1), default=-1)
    if start == -1:
        raise ValueError(f"No JSON found in response:\n{text[:400]}")
    depth, opener = 0, text[start]
    closer = "}" if opener == "{" else "]"
    for i in range(start, len(text)):
        if text[i] == opener:
            depth += 1
        elif text[i] == closer:
            depth -= 1
            if depth == 0:
                return json.loads(text[start:i + 1])
    raise ValueError("Unbalanced JSON in response.")


def log(step, msg):
    print(f"\n=== {step} ===\n{msg}", file=sys.stderr)


# ─────────────────────────────────────────────────────────────────────────────
# PIPELINE
# ─────────────────────────────────────────────────────────────────────────────
def run(domain, audience, assets):
    # STEP 1 — audience framing + challenge
    a = parse_json(call(GEN_MODEL, AUDIENCE_PROMPT.format(
        domain=domain, audience=audience, assets=assets)))
    carried = a["carried_audience"]
    wtp = a["willingness_to_pay"]
    log("STEP 1 audience", f"Carried: {carried}\nWTP: {wtp}\nAlternatives: "
        + "; ".join(x["audience"] for x in a.get("alternatives", [])))

    # STEP 2 — generate (multi-lens)
    g = parse_json(call(GEN_MODEL, GENERATE_PROMPT.format(
        domain=domain, audience=carried, wtp=wtp, assets=assets,
        capabilities=CAPABILITY_MENU), max_tokens=6000))
    candidates = g["candidates"]
    # Assign stable ids in code and thread them through every later step.
    # Steps are matched by id, never by title (titles get reworded by models).
    for i, cand in enumerate(candidates, 1):
        cand["id"] = f"c{i}"
    log("STEP 2 generate", f"{len(candidates)} candidates")

    # STEP 3 — cut (DIFFERENT vendor if OPENAI_API_KEY set, else different Anthropic model)
    c = parse_json(critique(CUT_PROMPT.format(
        candidates_json=json.dumps(candidates, indent=2))))
    keep_ids = {r["id"] for r in c["results"] if r.get("verdict") == "keep" and "id" in r}
    survivors = [x for x in candidates if x["id"] in keep_ids]
    log("STEP 3 cut", f"{len(survivors)} survive of {len(candidates)}")
    if not survivors:
        return render(domain, carried, wtp, [], note="No candidate survived the cut.")

    # STEP 4 — verify (web search)
    v = parse_json(call(VERIFY_MODEL, VERIFY_PROMPT.format(
        domain=domain, survivors_json=json.dumps(survivors, indent=2)),
        tools=[WEB_SEARCH_TOOL], max_tokens=6000))
    holds = {r["id"]: r for r in v["results"] if r.get("premise_holds") and "id" in r}
    verified = [{**x, "findings": holds[x["id"]].get("findings", "")}
                for x in survivors if x["id"] in holds]
    log("STEP 4 verify", f"{len(verified)} premises held")
    if not verified:
        return render(domain, carried, wtp, [], note="No premise survived verification.")

    # STEP 5 — score
    s = parse_json(call(SCORE_MODEL, SCORE_PROMPT.format(
        verified_json=json.dumps(verified, indent=2)), max_tokens=4000))
    # Merge by id. Drop any scored entry whose id we don't recognise (a model
    # hallucinated/reworded one) rather than rendering it with blank fields.
    by_id = {x["id"]: x for x in verified}
    scored = [{**by_id[sc["id"]], **sc} for sc in s["scored"] if sc.get("id") in by_id]
    skipped = len(s["scored"]) - len(scored)
    if skipped:
        log("STEP 5 score", f"dropped {skipped} scored entr(y/ies) with unmatched id")

    # STEP 6 — apply Risk rules, rank advancing, split.
    # Risk = 1 is regulated-profession territory; drop and surface separately so
    # the operator sees what got killed for risk vs what failed the Def/Will gate.
    # Risk <= 2 still advances if it passes the gate, but flagged as needing
    # liability work before any build.
    risk_dropped = [x for x in scored if x.get("risk") == 1]
    if risk_dropped:
        log("STEP 6 risk", f"dropped {len(risk_dropped)} for risk=1")
    remaining = [x for x in scored if x.get("risk") != 1]
    for x in remaining:
        x["needs_liability_work"] = (x.get("risk") or 5) <= 2
    advancing = [x for x in remaining if x.get("advances")]
    # Sum desc, then Risk desc as tiebreaker — prefer safer builds at equal fitness.
    advancing.sort(key=lambda x: (x.get("sum", 0), x.get("risk", 0)), reverse=True)
    return render(domain, carried, wtp, advancing,
                  dropped=[x for x in remaining if not x.get("advances")],
                  risk_dropped=risk_dropped)


def render(domain, audience, wtp, advancing, dropped=None, risk_dropped=None, note=None):
    out = [f"# Idea report — {domain}\n",
           f"**Audience:** {audience}  ",
           f"**Willingness to pay:** {wtp}\n"]
    if note:
        out.append(f"> {note}\n")
    if not advancing:
        out.append("**No candidate advanced the gate** (Defensibility ≥3 AND "
                    "Willingness ≥3). That is a real outcome — consider a "
                    "different audience, a new asset, or a different monetisation.\n")
    else:
        out.append("## Advancing candidates (ranked)\n")
        for i, x in enumerate(advancing, 1):
            tags = ""
            if x.get("short_lived"):
                tags += " — *short-lived (low durability)*"
            if x.get("needs_liability_work"):
                tags += f" — **⚠ needs liability work before building** (risk {x.get('risk')}/5)"
            out += [
                f"### {i}. {x['title']}  [{x.get('flag','')}]{tags}",
                f"- **Idea:** use *{x.get('capability','?')}* on *{x.get('asset','?')}* "
                f"to {x.get('beats_free_because','')}",
                f"- **Verification:** {x.get('findings','(none)')}",
                f"- **Scores:** Defensibility {x.get('defensibility')}/5 · "
                f"Willingness {x.get('willingness')}/5 · Buildability {x.get('buildability')}/5 · "
                f"Reuse {x.get('reuse')}/5 · Durability {x.get('durability')}/5  (sum {x.get('sum')})",
                f"- **Operator risk:** {x.get('risk')}/5 {_risk_note(x.get('risk'))}",
                f"- **Cheapest test:** {x.get('cheapest_test','')}\n",
            ]
    if dropped:
        out.append("## Did not advance (failed the Defensibility/Willingness gate)\n")
        for x in dropped:
            out.append(f"- **{x['title']}** — Def {x.get('defensibility')} / "
                       f"Will {x.get('willingness')} (gate is ≥3 on both)")
        out.append("")
    if risk_dropped:
        out.append("## Dropped for operator risk (Risk = 1)\n")
        out.append("Regulated-profession territory or comparable. These may be genuine "
                   "opportunities, but they aren't safe for us to build without the right "
                   "qualifications/cover — surfaced here for visibility, not as a recommendation.\n")
        for x in risk_dropped:
            out.append(f"- **{x['title']}** — Risk 1/5 (Def {x.get('defensibility')} / "
                       f"Will {x.get('willingness')} / Build {x.get('buildability')})")
    return "\n".join(out)


def _risk_note(r):
    """Short label shown next to the Risk score in the report."""
    return {
        5: "(pure analysis; customer decides)",
        4: "(low — refunds make customers whole)",
        3: "(moderate — cite sources; PII recommended)",
        2: "(high — needs review, insurance, tight T&Cs before building)",
        1: "(regulated territory — should not build without proper qualifications)",
    }.get(r, "")


# ─────────────────────────────────────────────────────────────────────────────
# Optional cross-vendor critique (stronger diversity than same-family).
# Fill in with your OpenAI / Gemini client and point CRITIQUE_MODEL's step at it.
# ─────────────────────────────────────────────────────────────────────────────
# ─────────────────────────────────────────────────────────────────────────────
# Cross-vendor critique for the cut step (STEP 3). If OPENAI_API_KEY is set, the
# cut runs on OpenAI — genuine cross-vendor diversity, so the method isn't a
# single vendor marking its own work. Falls back to a different Anthropic model
# (CRITIQUE_MODEL) when no OpenAI key is present.
# ─────────────────────────────────────────────────────────────────────────────
def call_other_vendor(system, user):
    """OpenAI Chat Completions call for the cut step. No max_tokens set, to avoid
    the max_tokens vs max_completion_tokens divergence across OpenAI models."""
    from openai import OpenAI
    oc = OpenAI()  # reads OPENAI_API_KEY
    resp = oc.chat.completions.create(
        model=OPENAI_CRITIQUE_MODEL,
        messages=[{"role": "system", "content": system},
                  {"role": "user", "content": user}],
    )
    return resp.choices[0].message.content


def critique(user):
    """Run the cut on a different vendor if available, else a different Anthropic model."""
    if os.environ.get("OPENAI_API_KEY"):
        return call_other_vendor(SYSTEM_BASE, user)
    return call(CRITIQUE_MODEL, user, max_tokens=4000)


EXAMPLE_INPUTS = dict(
    domain="agritec.uk",
    audience="UK small farmers (3-50ha), many eligible for SFI26 Window 1 (June 2026)",
    assets="we operate the site; can curate UK scheme docs; no proprietary data feed",
)


if __name__ == "__main__":
    p = argparse.ArgumentParser(description="Run the idea.uk ideation method.")
    p.add_argument("--domain")
    p.add_argument("--audience")
    p.add_argument("--assets")
    p.add_argument("--out", help="write report to this file (else stdout)")
    args = p.parse_args()

    inp = dict(
        domain=args.domain or EXAMPLE_INPUTS["domain"],
        audience=args.audience or EXAMPLE_INPUTS["audience"],
        assets=args.assets or EXAMPLE_INPUTS["assets"],
    )
    if not os.environ.get("ANTHROPIC_API_KEY"):
        sys.exit("Set ANTHROPIC_API_KEY first.")

    report = run(**inp)
    if args.out:
        with open(args.out, "w") as f:
            f.write(report)
        print(f"Wrote {args.out}", file=sys.stderr)
    else:
        print(report)
