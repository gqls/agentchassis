# RFC_049 — "is this internal URL a legitimate link target?" has now been hand-rolled three times. What consolidates it?

**Raised 2026-08-23** by the `bugs_open/328` lane, at the **explicit instruction of the council gate's
`architecture` seat**, which approved 328's mechanism and then said what to do about the pattern:

> *"approve this fix to proceed, but record on LNK-030 (or open a fresh RFC ticket) that this is now
> three instances, so it doesn't quietly become four."*

Both halves were done: this ticket, and an amendment on `LNK-030` carrying the count.

**Status: OPEN, UNOWNED.** Nothing is blocked on it. `bugs_open/328` shipped.

## What happened, in order

The estate has answered one underlying question — *may this internal URL be linked to?* — three
separate times, each in its own hand-rolled predicate, each correct, each with its own tests:

| # | where | question it answers | filed under |
|---|---|---|---|
| 1 | **CLC-013** `ResolveChromeComponent` | which library component serves chrome function F | `bugs_closed/118` |
| 2 | **LNK-030** `ChromeLinkPolicy` | which page may a piece of CHROME link to | `bugs_closed/191` |
| 3 | **LNK-038** `PageLinkRefusedPredicateFor` | which page may page BODY content link to | `bugs_open/328` |

After the second, the `architecture` seat wrote a standing advisory onto LNK-030: two bespoke fixes of
one class is *"a candidate RFC subject, not a solved problem … Do not treat LNK-030 as closing the
class."* That advisory did its job — the 328 author read it and **declared the third instance in the
submission** rather than letting a seat find it. Four seats (`architecture`, `guardian`, `reuse_agent`,
`constitution`) then objected on it anyway, all four correctly: **declaring a debt is not discharging
it.**

## Why the obvious consolidation is not obviously right

The three are not the same predicate wearing three names, and this is the part a consolidation has to
survive. They deliberately disagree, and the disagreements are measured:

- **Chrome negates `NeverDeployedPagePredicate`.** Chrome ships on every page, renders once behind an
  idempotence gate and has no repair pass, so its failure is a site-wide 404 nothing later corrects. It
  can afford to be strict.
- **Page content may NOT use that predicate.** Measured 2026-08-23 against live HTTP with a per-domain
  control: it selects **9 pages that return 200**. Delisting a working page is what this estate calls
  "worse than the bug" (`bugs_open/052`).
- **`PageMayBeLinkedPredicateFor`** — the floor written *because* of that — excludes only
  `planned` + never-deployed, and **misses 3 `needs_rebuild` rows that were never built at all**, 3/3
  returning 404.
- So 328 needed a fourth spelling: never-deployed **and** no rendered components, which separates the
  mixed class 20/20 vs 9/9.

**The naïve consolidation — one predicate for everyone — is therefore already refuted by measurement.**
Any real answer has to let callers differ while making the difference visible and deliberate, rather
than an accident of which helper each author happened to find.

## The question for the owner

Not *"should these be one function"* — measurement says no. The question is:

> **What shape makes the FOURTH case a configuration of an existing decision rather than a new
> predicate?**

Sketches, none costed, none preferred:

1. **A named policy enum** — `LinkTargetPolicy{Chrome, BodyContent, Listing, Suggestion}` resolving to
   one of the family's predicates, so adding a case means naming your policy and being told which
   predicate you get.
2. **A predicate registry with a stated question** — each family member declares the question it answers
   and its measured population, and a caller picks by question. Closest to what `datahelpers/links.go`
   already is informally; the change is making the choice explicit at the call site rather than by import.
3. **Leave it three, and make the advisory mechanical** — a check that fails when a new `deployed_at`/
   `build_status` predicate appears outside the family. Cheapest; does not consolidate anything, but it
   would have caught all three at authoring time.

## What the seat says the cost of doing nothing is

> *"the next site-URL edge case (redirects, archived-but-served pages, etc.) will again be a fourth
> hand-rolled predicate rather than a shared decision."*

Not a crash today — an accumulation. The estate already has the shape of that failure written down for
optional keys (`RFC_022`, budget N=10), and this is the same accumulation argument one layer over.

## If you are about to write a fourth

Read this first, and either consolidate or **add your case to the table above** so the count stays
honest. The one thing that must not happen is a fourth arriving with no record, which is how the
second became invisible until the third was written.

## Sources

- `docs/agent_docs/docs026_concept_register/register/link-management.md` — LNK-030 (the standing
  advisory + the 2026-08-23 amendment), LNK-038 (the third instance and its measurements)
- `bugs_open/328` — the fix that raised it; council corr `21c19c1f-e614-49bd-82ac-0bb5b58082e0`
- `platform/orchestration/datahelpers/links.go` — the predicate family as it stands
