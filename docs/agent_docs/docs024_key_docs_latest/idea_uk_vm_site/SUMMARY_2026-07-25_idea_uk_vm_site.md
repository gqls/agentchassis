# SUMMARY — idea.uk (2026-07-25)

*Previous summary: `SUMMARY_idea_uk_vm_site.md` (2026-07-18, the migration read-out). This one
marks a genuine turn: the site has changed from a migration project into a content-and-tools
business being built out. Written to be read aloud.*

## What we're trying to do

idea.uk should be the place someone with an idea actually works it out. Not a brochure — a
guided journey: plain-English guides for each stage of an idea's life (protecting it, funding
it, building and testing it), free tools that give an honest steer at each stage, and one paid
product — the £29 Verified Idea Report — that everything funnels towards. The site is also
meant to become the worked example the rest of the fleet copies, one maturity rung at a time.

## Where we've come from

A week ago this was a migration project: get the chassis-built static site and the
independently-running £29 report tool onto one origin on the Hetzner box, without breaking the
tool. That shipped on the 18th. Since then the work has been making the site honest and joined
up: the home page and header buttons actually reaching the paid tool, dead chrome links fixed,
the contact form actually delivering. As of yesterday the site was nine pages: sound, but
thin — a Guides section that was an empty heading, a News section the same, and one tool.

## What we've done

Today the pipeline itself started, and the first four stages of it are live:

- **Two protection guides** — "Patents: how to protect an idea in the UK" and "Copyright: what
  you already own" — hand-written, UK-specific, honest about unsettled law, each ending at the
  paid report. Hand-written because the platform's fact-checking gate isn't live yet, and legal
  guidance is the last place to trust generated text.
- **Two funding guides** — the *ways* (the eight mechanisms and what each really costs) and the
  *sources* (the actual UK map: Innovate UK, the British Business Bank, devolved agencies,
  angels, trusts). These deliberately name no amounts, rates or deadlines — those go stale and
  a stale figure is indistinguishable from an invented one.
- **The first free tool** — "Should you patent it?", six questions in the browser, nothing
  stored. Built to *gate* rather than score: if someone has already disclosed publicly, no
  number of good answers elsewhere should tell them they look patent-ready.
- **The plumbing that makes the next guide cheap**: both the Guides and Tools pages now build
  their listings from the database automatically, so every future guide or tool lists itself.
  The repeatable build recipe is written down (RUNBOOK Phase 5), with the traps documented.
- **The tools page repaired** after the owner's report: its buttons went to the contact form
  (now: free check and paid report), the diagram was a missing file (now the real illustration),
  an invented statistic claimed "8 free tools" (now the true number), and the paid report now
  appears in the tool listing alongside the free tools.
- **An audit of the paid tool against its own sales page** (owner's question). Answer: the tool
  hasn't changed — it is, and always was, an idea *finder* (it generates and web-verifies AI
  product ideas for your business, with real refusal outcomes and human review). The sales copy
  describes a different product — assessment of *your single idea*, with checkable citations —
  and two of its specific promises are not what a paying customer receives. Full detail in
  `AUDIT_2026-07-25_paid_tool_vs_copy.md`; the fix direction is the owner's call.

## Where we are now

Live and verified: six pipeline pages (guides hub, four guides, patent checker), the repaired
tools page, and the paid report — every button on every one of them pointing somewhere real.
The authored legal and funding content is locked so no automated pass can rewrite it; the
derived listings are deliberately *not* locked, after nearly freezing one — the day's most
instructive mistake, now written up. The one open decision is the paid-tool copy mismatch
above. The pattern the site follows is now stable: guide → free tool where it earns its place →
funnel to the report.

## Where we're going

Next stages of the pipeline: ideation, build, test, user acceptance and feedback loops — the
earlier, lower-stakes stages where generated copy becomes acceptable. A "funding-fit finder"
free tool is the natural next tool, built on the same gate-before-score rule. The paid-tool
copy question resolves either into a copy rewrite (cheap, immediate) or an engine extension
(single-idea assessment with citations — the better product, more work). And once the pipeline
is far enough along, idea.uk becomes the top-rung exemplar for the fleet-wide site maturity
ladder that's being planned separately.
