# DESIGN NOTE — premise extraction, branch pages, and the "deepthink" composition

**Raised by the owner, 2026-07-28**, against the Thames Water dossier and stated as
generic for "this type of site and similar to the news editorial requirement":

> extract from the content pertinent points and then extend the article with maybe
> different pages for each major point … for Thames Water the pertinent points will
> be the stakeholders involved and the data the deal is based on, it also touches
> the politics and economy forecasts, population growth, competitors … for each of
> these — probably just the major one or two — branch off to a deeper exploration
> that may have historical data graphs and tools and factual commentary that may be
> updated regularly … a sort of deepthink workflow that works with the experience
> workflows and the tool checking workflows and tool builder workflows and graphing
> tools.

**Status: thinking, nothing built.** This note records the analysis and the
recommendation, not a decision.

---

## 1. This is `features_open/001`, rotated ninety degrees

`features_open/001_FEATURE_packaged_topic_features.md` ("living dossiers") was
raised by the owner on 2026-07-19 out of the news-feed-pooling design. Its framing
there is almost word-for-word the same instrument:

> "select a pertinent topic … break down its parts — e.g. if it is gas prices then
> we'd look at previous gas, oil prices, the Hormuz war, inflation rates and
> collected opinions etc — and create a packaged feature about it, updated as we go
> until it gets irrelevant."

001's central design is the **substrate/angle split**: the expensive research
substrate (timeline, price history, figures, citations) is built **once**, and the
cheap per-site **angle** is generated many times, one per domain, from that site's
`audience` aspect.

**The owner's request today is the same substrate on the other axis.** 001 spreads
one topic across many *sites*. This spreads one topic across many *pages* of one
site:

| | 001 (across sites) | today's request (across pages) |
|---|---|---|
| built once, expensive | research substrate | research substrate |
| generated many times, cheap | one angle per domain, from `audience` | one branch page per premise, from the premise |
| what varies | who is reading | which question is being answered |

**This is the single most important observation in this note.** They are one
feature with two projections. If a Thames-specific decomposition is built now, the
substrate concept forks, and we end up with two incompatible implementations of
the same idea — which is precisely the drift class the council gate exists to
catch. **Whatever is built must treat the substrate as the shared entity and the
page (or the site angle) as the projection.**

001 also settles a dependency that oufe happens to already satisfy: the `audience`
aspect is a **prerequisite, not a parallel task** ("building packages before
profiles exist produces 231 variations on one article"). oufe has one, seeded
2026-07-27 (`audience.v1`, `sophistication: institutional`, students explicitly
out of scope).

## 2. The genuinely new primitive is premise extraction, and its definition matters

Nothing in the estate decomposes an artefact into what it rests on. That is the
new piece, and it is the piece that generalises across site types — which is what
the owner asked for.

**The definitional trap: extract PREMISES, not TOPICS.** They look similar and
produce completely different sites.

- *Topic:* "the stakeholders in the Thames Water restructuring." Produces an
  encyclopaedia page. Nobody needs it; it is a worse Wikipedia.
- *Premise:* "the outcome turns on whether the Class A group's valuation of the
  relevant alternative is the one the court accepts." Produces a page with a
  reason to exist, and it tells you **what tool and what graph belong on it** —
  because a premise is a claim that can be tested, and a tool is a way to test it.

A premise is a load-bearing assumption the main article stands on. The test:
**if this turned out to be false, would the main article change?** If not, it is
background, and it belongs in a sentence rather than a page.

That test also answers "probably just the major one or two" mechanically, rather
than by taste: rank premises by how much of the parent article depends on them.

Applied to Thames Water, the plausible premise set:

| candidate premise | load-bearing? | what the branch page would carry |
|---|---|---|
| The relevant alternative (what happens if the plan is not sanctioned) is the pivot of the whole cram-down test | **yes — the article's central mechanism** | a tool: move the counterfactual recovery and watch which classes lose their veto |
| The Class A / Class B split determines who has a real economic interest | **yes** | waterfall tool (built), extended with class voting |
| Ofwat's determination sets the revenue envelope the plan is modelled on | **yes** | historical: determinations over time vs actual returns |
| Population growth drives demand forecasts | weakly — one input among many | a paragraph, not a page |
| Competitors | **no** — a regulated regional monopoly has no competitors in the ordinary sense | nothing. Worth stating why, because its absence is itself informative |

That last row is the point of doing this deliberately: a generic premise
extractor asked for "competitors" will happily invent a competitive landscape for
a regional water monopoly. **The extractor must be allowed to return "this premise
does not apply here, and here is why", and that must be a first-class output
rather than an empty result.**

## 3. The real blocker is the substrate, not the rendering

Corrected 2026-07-28 (see `WRONG_CALLS.md`): the claim that there is no chart
renderer was wrong. There are two.

- **`evidence-chart`** — live, active, section-level. CSS-drawn horizontal bars,
  no SVG, no dependency. **Every plotted point resolves through a `fact_id` into
  the evidence register**; the denominator can itself be a `max_fact_id`; each row
  renders `verified <date>` and the figure carries a `source_note`. A chart point
  structurally cannot carry its own number. This is exactly the doctrine oufe
  needs, already built.
- **`report_charts.go`** — `renderBarChartSVG`, `renderHeadroomChart`; inline SVG,
  dependency-free, unexported, bound to report pages.
- **`features_open/023`** — the designed rule that infographic prompts are built
  from `evidence_base`, plus the boundary that matters here: **generated images
  explain, code-rendered SVG states.** Anything exact, selectable or translatable
  must be code-rendered. "Historical data graphs" is squarely on the
  code-rendered side.

**What is missing is a time-series renderer — and underneath it, a schema gap that
is the actual blocker.** An `evidence_base` fact looks like this:

```json
{ "id": "CIT-…", "kind": "metric", "value": 75, "unit": "percent",
  "source": { "citation": { "url": "…", "quote": "…", "accessed": "2026-07-26",
                            "published": "2020" } },
  "verified_at": "2026-07-26", "staleness_days": 800 }
```

`accessed`, `published` and `verified_at` are all **provenance** dates: when we
looked, when the source was issued, when we last checked. **None of them is the
date the value applies to.** A fact holds *one* value.

A historical graph needs an **observation series**: many values, each with an
`as_of`, each independently sourced. That is a different shape, and it is a
substrate question before it is a drawing question. Building a line-chart renderer
first would produce a component with nothing legitimate to plot — and the failure
mode is not an empty chart, it is a writer filling the series from the model,
which is the one thing this site's entire posture forbids.

**So the ordering is forced: series-shaped facts first, renderer second.** The good
news is that `evidence-chart` already proves the resolution pattern (values arrive
by `fact_id`, never inline), so this extends a working design rather than
inventing one.

## 4. What this multiplies, and why that needs saying out loud

Every branch page in this scheme is **figure-dense by construction**. Population
growth, economy forecasts, historical determinations, price series — these pages
are *nothing but* numbers over time. The parent article can be mostly mechanism
with a handful of cited facts; a branch page cannot.

So this feature multiplies the exact surface that oufe's honesty posture depends
on, by roughly the branch factor. That is not an argument against it. It is an
argument that:

- the substrate must be the evidence register, extended, and never a parallel
  store (`bugs_open/043`: a number in a spec is a *given* and outranks every
  writer-side rule);
- `023`'s boundary is not optional here — no generated imagery for any of it;
- "updated regularly" needs 001's **two update classes**: substrate-only (new
  observations, angles re-render unchanged) versus narrative (the story moved,
  derived pages must be regenerated). Only the second costs per-page money, and
  only the second needs a blast-radius count. 001 already flags that the fan-out
  record is the step most likely to be skipped and most likely to hurt.

## 5. On composing with the existing loops — one hard rule

The owner asks for a deepthink workflow that works *with* the experience
workflows, the tool-checking workflows, the tool-builder workflows and the
graphing tools. Those all exist and the composition is real. One caution, from a
defect found today:

**`bugs_open/126`** — a failing Tier-4 tool acceptance auto-raises an
`improve_tool` item **carrying the failing criteria as the specification**, and
dispatches it to a rewriting agent. On the waterfall tool, the only way to satisfy
my own bad fence was to weaken or delete the legally load-bearing consent gate. It
was cancelled by hand.

That defect is survivable today because tools are built **one at a time, by a
human who is watching**. A deepthink workflow that generates tools per premise,
across pages, across sites, removes the human from exactly that position while
multiplying the number of auto-repair dispatches.

**Proposed rule, before any of this is automated: an artefact generated by the
deepthink lane on an evidence-gated site enters human review; it never enters
auto-repair.** This is the same posture already chosen for `grounded-explainer`,
which cannot publish and ends at `needs_human_review` — and that rail is listed in
the oufe handoff as one that must not be relaxed. The lane composing *with* the
loops must not mean the loops acquiring authority over its output.

The experience loop is, by contrast, the right instrument for the genuinely hard
editorial question here — *is this branch page worth existing?* — because its
honesty critic already holds a hard veto.

## 6. Recommendation: do not build the workflow first

Three reasons to hand-build one worked example before designing the lane.

1. **001's open questions are unanswered and expensive to reverse.** Who picks the
   topic; what the update trigger and retirement rule are; whether an updated
   package keeps one URL (accrues authority) or publishes a new one (fragments
   it); whether the substrate is a first-class `topic_packages` entity or just a
   work item output. The SEO one bites this proposal hardest: **branching one good
   page into six thin ones is a reliable way to dilute a site**, and that is the
   default outcome if premise extraction is run without the load-bearing test in
   §2.
2. **`features_open/015`, the maturity ladder**, warns specifically against
   top-rung-first. A generalised multi-site deepthink lane is several rungs above
   where oufe is (one dossier, one tool, live for three days).
3. **One hand-built branch tells us what the lane must automate.** Guessing
   produces a lane that automates the wrong step — and the expensive step here is
   almost certainly premise *selection*, not page generation.

**Suggested first slice, in order:**

1. Extend `evidence_base` with a **series** fact kind: many observations, each
   with `as_of` and its own citation. Schema first, one worked series (Ofwat
   determinations, or Thames debt over time), verified end to end.
2. A **time-series component** on the `evidence-chart` pattern — values resolved
   by `fact_id`, never inline; `verified` and `source_note` preserved; CSS or
   inline SVG, no dependency.
3. **One hand-authored branch page** off the Thames dossier, on the strongest
   premise (the relevant alternative), carrying that series and one extension to
   the existing waterfall tool. Human-written, so we learn what good looks like.
4. Only then design the lane, against a real example, and put it through the
   council gate — where the reuse seat should be asked explicitly whether it is
   `001` with different parameters. It probably is.

## 7. What I would push back on

- **"Competitors" for a regulated monopoly.** Included above because it is the
  clearest case where a generic premise list produces confident nonsense. The
  extractor needs to decline premises, with a reason.
- **"Updated regularly" on a site that is three days old.** 001's own cost shape
  assumes a pool with volume behind it. One dossier with six branch pages, each
  needing refresh, is a maintenance commitment before there is an audience to
  serve — and `C7` in this workstream's PLAN already deferred the news feed on
  exactly that reasoning ("one flagship dossier kept current beats ten stale
  ones"). Six branches make that harder, not easier.
- **The tools-per-premise idea is the strongest part of the request** and I would
  weight it above the branch pages. "How do the stakeholders' expected recoveries
  move if one input changes" is what the waterfall tool already does; extending it
  per premise is cheap, needs no new substrate, and is the thing a professional
  reader would actually use twice.
