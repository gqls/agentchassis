# PRIOR ART — what earlier conversations and docs hold on domain valuation (2026-09-02)

## The search, and its honest result

The owner: *"I have discussed valuations of domains previously (.co.uk and .uk) …
Please search the previous conversations to get an initial idea and starting point."*

Searched: all 646 retained transcripts across both project dirs
(`~/.claude/projects/*/`), owner-typed messages only, three passes (valuation
terms; UK-domain + money terms; Afternic/Dan/floor terms). A subagent ran the
first pass and died on an auth error; the passes were re-run inline.

**FINDING: no dedicated .co.uk/.uk valuation conversation exists in this
machine's retained transcripts.** Nothing before 2026-09-02 discusses per-domain
portfolio valuations. The discussion the owner remembers most likely happened on
claude.ai (web), on another machine, or predates transcript retention (retained
transcripts effectively start late July 2026). **Asked the owner where it took
place** — if he can point at it (or paste its conclusions), it supersedes the
starting points below.

## What DID surface (use as the starting frame)

1. **Owner-set floor on a premium .com: relojistas.com Afternic minimum offer =
   $12,000** (owner-confirmed 2026-07-28; `WRONG_CALLS.md` §"the Afternic
   minimum offer is 0" — the 0 was a misread paste, the real floor is $12k).
   Also establishes: Afternic dashboard carries `Minimum Offer · Sale Lander ·
   Views · Leads · 30-day Searches` — demand signals worth ingesting from the
   export when it lands.

2. **£150 transfer-away fee** (owner, 2026-08-17, webdesign.co.uk lane:
   *"I'd like to charge £150 pounds as a transfer away fee. Basically they are
   buying the domain off me."*) — a de-facto floor under any "keen" BUY_NOW
   price: below ~£150 the sale is worth less than the fee he already charges
   for the same outcome.

3. **The value ladder** — `docs/architecture/010-domain-value-maximisation.md`
   (the portfolio's stated strategy): naked domain $500–2,000 (sits for years) →
   +basic site $1,000–3,000 → +traffic $5,000–20,000 → +revenue $10,000–100,000+.
   So valuation here is not static appraisal only: a domain carrying one of the
   framework's built sites belongs a rung up, and the bottom-500 sell list
   should generally be names the site-building programme will never reach.

4. **The Jul-31 portfolio subset** —
   `portfolio_positioning/PORTFOLIO_domains.txt` (152 domains, owner-supplied,
   finance-heavy: mortgages/loans/insurance/banking clusters). Classified
   2026-08-19: **124 of 152 parked at Dan.com**, 6 on Cloudflare (built sites),
   3 aftermarket.com, 17 never delegated (owner: *"No nameserver usually means
   I never set a nameserver"* — registered-but-never-delegated, NOT lapsed).

5. **Traffic ≠ value** — `news_feed_pooling/RESEARCH_2026-07-20_dormant_domain_history.md`:
   per-domain archive research on the .coms; its headline for pricing: *"High
   views ≠ valuable views"* (the highest-views domain of the opaque set was
   poisoned casino-spam traffic). Afternic Views/Leads columns get read with
   that caution.

6. **Category structure prior art** —
   `portfolio_positioning/REGISTER_positioning.md` (65KB): finance verticals
   already carved into families (mortgage tools, loans, insurance evaluation
   layer, …) with per-domain angles. The valuation categoriser should reuse
   these families, not invent parallel ones.

## Marked inferences

- [INFERRED] The prior valuation talk being off-machine. Absence proven only
  for retained transcripts here; "it never happened" is not claimed.
- [INFERRED] $12k relojistas floor generalises only as "owner prices premium
  brandables in five figures", not as a portfolio-wide anchor.
