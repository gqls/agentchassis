# 328 — a page that failed to build is still linked from the pages that did, so one blocked page becomes a visibly broken site

**Filed 2026-08-19** by the `loanzy_uk_example_site` lane, from a greenfield one-shot build.
**Status: OPEN, UNOWNED. Live on `loanzy.uk` today.**

> **On the 090 loop:** not run. This is not a hypothesis about a mechanism — it is two URLs,
> one serving and one 404, and the anchor between them, all read from the live site.

## The one-paragraph version

When a page fails to build, the build knows. The item is `needs_human_review`, the `pages` row
never reaches `deployed`, and the failure is recorded. **Nothing tells the pages that link to
it.** They are built, deployed and served with anchors pointing at a page that does not exist,
so a single blocked page turns into a live site with dead navigation — and the reader gets a
404 with no indication that anything is unfinished.

## Evidence (measured 2026-08-19, live)

`https://loanzy.uk/` serves 200 and contains:

| link on the live home page | target status |
|---|---|
| `/your-rights.html` | **404** — build refused at `validate_content` (`bugs_open/260`, 20 blockers) |
| `/guides/index.html` | **404** — builder no-op'd honestly: *"no sections ready to build"* |
| `/tools/loan-comparison-calculator/index.html` | 200, **but empty of its tool** (`bugs_open/311`) |

Two of the three defects behind those targets are already filed and owned. **This bug is about
the third fact: the links shipped anyway.** The home page was built AFTER `your-rights` had
already failed, so the information was available at build time and simply not used.

## Why it is its own defect and not part of 260/311

Fixing 260 removes this instance; it does not remove the class. **Any** page can fail to build —
truncation, a validation blocker, a missing component, an exhausted retry budget, an
infrastructure burst (`bugs_open/307`) — and every one of them produces the same shape: a live
site advertising a page that is not there. The route needs one rule, not one rule per cause.

It is also the difference between the two possible readings of a build. A site with four good
pages and no links to a fifth reads as *small*. The same site with two dead links reads as
*broken* — which is the judgement a customer makes about the product, not about the page.

## Fix candidates, ordered by what makes the bad state unrepresentable

1. **Resolve links against `build_status` at deploy time.** A link whose target is not
   `deployed` is dropped (or rendered as plain text) by the deploy step. The bad state stops
   being representable because the anchor cannot outlive the missing page. ⚠ must consult
   `pages.build_status` **and** the artefact, given `bugs_open/315` (`deployed_at` is stamped
   whether or not the object was written).
2. **Hold the first publish until the corpus is complete**, then publish atomically. Strongest,
   and the most disruptive: it changes the route from incremental to all-or-nothing, and it
   would also have prevented the stale-nav window this build served for ~1 hour.
3. **File a visible item per dangling link and re-render the referring pages** once the target
   builds. Weakest — it is repair after publication, and it depends on the repair path
   actually running (see `bugs_open/313`, where `plan_links` has never run at all).
4. **At minimum, a post-deploy assertion**: after a build completes, fetch every internal link
   on every deployed page and fail the run if any 404s. This does not prevent the defect, but it
   converts "nobody noticed" into "the build reported it", which is what happened here only
   because a human went looking.

## How to verify a fix

Build a site in which one page is *forced* to fail (a deliberately invalid section is enough),
and require that no deployed page links to it — asserted **at the served HTML**, not at the
plan, the sections, or `page_components`. Positive control in the same run: a page that DID
build must still be linked, or the fix has simply stopped emitting internal links (which
`bugs_open/313` shows is a state this platform can reach and not notice).

---

## CONTRIB 2026-08-21 — from the `mortgagecalculator_couk_adoption` lane: the detector already exists, so candidate 3 is much cheaper than it looks (and candidate 4 already half-runs)

**Second site, same shape, 25 days older.** `mortgagecalculator.co.uk` has carried this defect
since adoption, so this is not a greenfield-build artefact — it is what the shape looks like once
it has had a month to settle. Measured live 2026-08-21 ~10:20Z.

### The instance

One page cannot build (`scorecard-simulator`, refused by `bugs_closed/260`'s template leak).
**Six deployed pages each serve exactly one anchor to it.** Full audit, not a sample: all 29
`build_status='deployed'` pages fetched, every non-absolute `href` extracted, deduped to 33
distinct targets, each resolved once cache-busted.

| measure | value |
|---|---|
| deployed pages fetched | 29 |
| raw internal hrefs | 1,030 |
| distinct internal targets | 33 |
| targets returning 200 | **32** |
| targets returning 404 | **1** — `/scorecard-simulator.html`, 404 on 3 cache-busted repeats |
| deployed pages serving an anchor to it | **6** — `/disclaimer.html`, `/guides/first-time-buyer/`, `/guides/how-banks-decide/`, `/guides/lender-restrictions/`, `/guides/mortgage-scorecard/`, `/tools/affordability/` |

### The correction that matters: **the platform is NOT blind to this**

§"The one-paragraph version" says *"Nothing tells the pages that link to it."* On this site that
is **false in a way that changes the fix.** There are **seven `unbuilt_internal_link` work items**
on the blocked page's `page_id`, one per linking component, each naming the linking page, the
component function, and quoting the href verbatim:

```
unbuilt_internal_link in page_component (guide-mortgage-scorecard:generic-text-block):
  href "/scorecard-simulator.html" points at a page that has never been deployed
```

`page_id` on the item is the **target** page; the linking page is named inside the summary text.

**So detection exists, is per-linking-page, and is accurate.** What is missing is the route: all
seven sit at `status='needs_human_review'` with no handler, which is `bugs_open/033`'s "no working
surface" queue. The information is produced correctly and then parked unread.

That reprices the candidate list in §"Fix candidates":

- **Candidate 3 is not the weakest option on this evidence — it is the one that is already 80%
  built.** It does not need a new detector; it needs the existing item type wired to a handler
  that re-renders the referring page (or drops the anchor) once the target reaches `deployed`.
- **Candidate 4 is also already half-running**, for the same reason — the finding exists, it just
  never reaches anyone. "Convert nobody-noticed into the-build-reported-it" is a delivery problem
  here, not a detection one.
- Candidates 1 and 2 stand unchanged and remain the only ones that make the state
  **unrepresentable**; this contrib is about cost, not about preferring repair to prevention.

### A second finding: a parked item is not evidence its condition survives

**Six of the seven items match the live site. The seventh is stale.** `8a230338` names
`contact-index`, and the href is in neither the served page nor the stored content:

| check | result |
|---|---|
| `curl /contact/index.html \| grep -c scorecard-simulator` | **0** |
| `page_components.content_data LIKE '%scorecard-simulator.html%'`, this site | **6 pages — `contact-index` absent** |

Something repaired that link and nothing closed the item. Whoever builds the handler needs this:
**it must re-check the condition at handling time, not trust the summary**, or it will re-render a
page to fix a link that is not there. [UNVERIFIED] whether this staleness is general — measured on
this site only, n=1 of 7.

⚠ **And the stored/served split bites here.** `contact-index` is `build_status='needs_rebuild'`
with `deployed_at` NULL, yet serves a healthy 1,267-word page from its previous build. A `pages`
row that does not say `deployed` is **not** evidence the URL is dead — relevant to candidate 1,
which proposes gating anchors on `build_status`: on this site that predicate alone would have
dropped every anchor pointing at a perfectly healthy served page. Candidate 1's own ⚠ about
`bugs_open/315` is the same hazard from the other direction; the honest version of candidate 1
consults the artefact, and `build_status` is at best a cheap pre-filter.

### The self-fuelling property, which is what makes this worse than a static count

Every page the framework writes here **adds another anchor to the page it cannot build**, because
the site's `design_intent` names `Scorecard Simulator` as a delivered feature. Measured over this
lane's own work: while three unrelated dead links were being fixed on 08-16 by building the pages
they pointed at, live instances of *this* href went **4 → 6** — because the two pages the framework
had just built each linked to it unprompted. The count is not stable; it grows with productivity.
Three separate work items on this site (`2dca1a09`, `d5a9ae7d`, `0c65f9fa`) exist purely because
different detectors each noticed the brief promises a page that is not there.

**That is the argument for candidate 1 or 2 over 3 or 4**: repair-after-publication has to keep
pace with a producer that never stops, and on this site the producer is the brief itself.

### Cross-reference

`bugs_closed/260` closed 2026-08-20 and its fix is proven aboard `agent-chassis` v1.0.1321
(binary probe, four controls). **It does not close this instance** — 260's own closure note says
so explicitly: the fix makes a mistyped field fail early and named, it does not make the parked
content build. This lane re-fired the build on 08-21 to find out which. Whatever the outcome, the
six anchors were served for 25 days regardless, which is this bug's point and not 260's.

Lane record: `docs/agent_docs/docs024_key_docs_latest/mortgagecalculator_couk_adoption/NOTES_mortgagecalculator_couk.md`,
`## 2026-08-21`.
