# SPEC — V5: researched, cited, re-verifiable external facts

**Raised:** 2026-07-20 by the owner, for gaswholesalers.com:

> "I'd like the gaswholesalers site to consistently use numbers that are verified
> from web deepsearch cited references, so not manual but part of the chassis'
> capability."

**Status:** designed here, not built. This is the claims-verification thread's
next build. The site's own repositioning is `features_open/006` and is not this
thread's work.

---

## 1. What already exists (checked, 2026-07-20)

Most of the machinery is present and simply not connected to the evidence base.

| piece | what it does | state |
|---|---|---|
| `research-agent` (active) | `web_search` → `prepare_urls` → `batch_webscrape` → LLM `synthesize` → `format_research_content` → `insert_research_result` | live, used by page-content-writer |
| `web_search`, `batch_webscrape`, `rag_lookup` / `rag_index` | the retrieval primitives | registered actions |
| page-content-writer prompt | already injects `## Research Findings` + a `Sources:` list of title/domain | live |
| V1 number scan | flags any number asserted about the business that no fact supports | live |
| V2 whitelist | tells the writer the registered facts are the ONLY numbers it may assert | live |
| V4 freshness | re-runs `sql` facts, re-syncs values, raises drift for a human | live |

**So the enforcement lanes already exist.** A researched number that is *in the
register* is automatically usable by the writer (V2) and automatically protected
from drift (V4); one that is *not* in the register is automatically flagged (V1).

## 2. The three gaps

1. **Sources are page-level, not claim-level.** The writer receives a prose summary
   plus a list of source titles. Nothing binds *this number* to *that source*. The
   model can therefore state a figure and gesture at a citation list that does not
   actually support it — a fluent, well-cited-looking falsehood, which is worse than
   an obviously unsourced one.
2. **Research output never becomes evidence.** Findings land in `research_results`,
   flow into one prompt, and evaporate. The register learns nothing, so the next
   page re-researches from scratch and may state a *different* number for the same
   fact.
3. **Nothing re-verifies a citation.** V4 re-runs SQL facts. An external citation
   is never re-checked, so a number stays "verified" long after its source page has
   changed, moved, or been withdrawn.

## 3. The design

### 3a. A new source kind: `citation`

```jsonc
{
  "id": "MKT-lng-trade-2025",
  "claim": "global LNG trade reached 411 million tonnes in 2024",
  "value": 411,
  "unit": "million tonnes",
  "kind": "metric",
  "source": {
    "citation": {
      "publisher": "International Gas Union",
      "title": "World LNG Report 2025",
      "url": "https://…",
      "published": "2025-06",
      "quote": "global LNG trade reached 411 MT in 2024",   // VERBATIM from the source
      "accessed": "2026-07-20"
    }
  },
  "verified_at": "2026-07-20",
  "tolerance": "exact",
  "staleness_days": 400,
  "writer_line": "global LNG trade reached {value} million tonnes in 2024 (IGU, World LNG Report 2025)"
}
```

The `quote` field is the load-bearing part — see 3c.

### 3b. Acquisition: research → register, not research → prompt

A new agent, `evidence-researcher`, reusing the existing primitives:

```
load evidence gaps  → the claims a page WANTS to make but has no fact for
  → web_search (targeted per claim, not per topic)
    → batch_webscrape the candidate sources
      → execute_llm_prompt: EXTRACT ATOMIC CLAIMS, not a prose summary —
         each = {value, unit, verbatim quote, url, publisher, date}
        → verify_citations (Go, deterministic — 3c)
          → write verified ones into evidence_base as `citation` facts
            → unverifiable ones become a needs_human_review item, never a fact
```

**The critical prompt difference from `research-agent`:** it must return *atomic
claims with verbatim quotes*, not a synthesis. A summary cannot be checked; a quote
can.

### 3c. Verification is deterministic, and that is the whole point

For an external citation the check is not "does an LLM believe this" — it is:

> **fetch the URL, and assert the `quote` string still appears in the fetched text.**

That is a string comparison, exactly like the email check that started this
programme. It gives us, for free:

- **acquisition-time verification** — a hallucinated citation fails immediately,
  because the quote will not be found at the URL. This is the defence against the
  classic failure of research agents: a plausible-looking reference that does not
  say what it is cited for, or does not exist.
- **re-verification on a schedule** — V4's freshness pass extends to citations:
  re-fetch, re-match, and raise `stale_evidence` when the quote no longer appears
  (page changed, paywalled, withdrawn) or when `staleness_days` is exceeded.
- **an audit trail a reader could follow** — publisher, title, date, URL and the
  exact sentence relied upon.

A quote that cannot be re-found does **not** silently drop the fact: it raises a
work item, because a claim already published on the strength of it is now
unsupported.

### 3d. Nothing else needs building

Once a citation fact is in the register:

- **V2** injects it into the writer's whitelist, so the number becomes usable —
  and the `writer_line` carries the attribution, so the copy cites its source in
  the reader's language.
- **V1** flags any number the writer states that is *not* in the register.
- **V3** judges prose assertions against the register as before.
- **V4** re-verifies on the schedule.

That is the payoff of having built the lanes first: the external-evidence feature
is an *acquisition* problem, not an enforcement problem.

## 4. What makes this hard (do not skip these)

1. **Atomicity.** "The market grew strongly" is not a fact. The extractor must
   refuse anything without a number and a verbatim quote, or the register fills
   with unfalsifiable prose.
2. **Quote drift.** Publishers reformat. Whitespace, unicode dashes, thousands
   separators and HTML entities will break naive matching — normalise both sides
   before comparing, and prefer a distinctive substring over a whole sentence.
3. **Paywalls and JS-rendered pages.** Many authoritative energy sources are both.
   A fact whose source cannot be re-fetched by the platform is not re-verifiable;
   decide explicitly whether it may still be used (probably yes, flagged
   `reverifiable: false`, with a human attesting they read it).
4. **Source quality is a judgement, not a fetch.** A number correctly quoted from a
   bad source is still a bad number. The feed pipeline already scores credibility
   (`credibility`, `source_tier` on `content_feed_items`) — reuse that vocabulary
   rather than inventing a second one.
5. **Aggregator laundering.** A blog quoting a report is not the report. Prefer
   primary sources; record when a citation is second-hand.
6. **Numbers move.** Market figures are revised. `staleness_days` per fact, not one
   global policy: a reserves estimate and a spot price age very differently.

## 5. Why this matters beyond one site

gaswholesalers is the forcing case, but this generalises: it is the mechanism by
which **any** site can state an external fact honestly. Everything the layer does
today is about our own database (leopardess) — this is what it takes to be truthful
about the world. It is also the prerequisite for the AI-influence page
(`features_open/006`), which the owner requires to be "very well researched and
verified", and it is the same mechanism a future advisory chatbot
(`features_open/007`) would need if its answers are ever to be grounded.

## 6. Build order

1. `citation` source kind + verifier (Go, deterministic, testable offline with
   fixture HTML) — smallest piece, and it is the part that makes the rest honest.
2. Extend V4's freshness pass to re-verify citations.
3. `evidence-researcher` agent (acquisition), reusing web_search/batch_webscrape.
4. Gap-driven research: derive "what do we need a fact for" from the page plan,
   so research is targeted at claims the site intends to make.
