# PLAN — one flow, three brief sources; and the regulated-identity gate

**Status:** design, for the owner. Written 2026-08-19 from his decisions and his invitation to
discuss. Nothing here is built.

**The owner's decisions this rests on:**
- *"I think the correct flow is flow A because most times we will have a brief."*
- *"For my thousands of domains I will not have time to write a brief for each so I think maybe
  we have a step that writes a comprehensive brief (not a short one) for each domain… but maybe
  this could be an optional agent step that includes config enabled hitl step if I want it."*
- *"The webdesign.uk tool will have third party briefs of all sorts of quality and completeness
  that we should be able to handle too."*
- *"For now we reject a request until they write to us by email with proof."*

---

## 1. The decision dissolves the A/B question, and that is the point

Flow A and Flow B were never two pipelines — measured at the handler, all three built sites
produced identical work-item types handled by identical agents. The only difference was whether
a `mission_brief` spec existed before dispatch.

**So choosing Flow A does not mean "a human writes every brief". It means the brief is always
present.** Where it comes from is a separate question with three answers:

| source | who | volume | quality |
|---|---|---|---|
| **human** | the owner or a lane | tens | high, and the current proof |
| **generated** | a new brief-writer step | thousands | unknown until measured |
| **third party** | webdesign.uk customers | unbounded | *"all sorts of quality and completeness"* |

**One contract, three producers.** That is the shape, and it is worth stating because it makes
the third-party case free rather than extra: a thin customer brief and a bare domain name are
the same problem — *an incomplete brief that must be brought up to contract before dispatch.*
The generated-brief step and the third-party-brief handler are the same component, differing
only in how much they are given to start with.

**Where it plugs in, unchanged:** `mission_brief` is already a `site_specs` aspect, and
`domain-submitter` already writes it (12 of 15 live briefs came through it; the hand-written
ones arrived via its `--mission-file` argument). So a brief-writer step sits beside
`domain-submitter` and writes the same aspect. **Nothing downstream changes** — which is the
strongest argument for this shape over any alternative.

## 2. What "comprehensive" has to mean, or the briefs will promise things we cannot build

The owner wants briefs rich in *"content types and sections, news, directories, tools, games,
editorials, research, and the rest as appropriate."* Every one of those already maps to a real
platform mechanism:

| the brief asks for | the platform has | note |
|---|---|---|
| news | `NEWS-001` news-feed pipeline | live |
| directories | `DIR-001`, six kinds | live; a NEW kind is a seven-place change |
| tools | the tool library + `add_tool` → `deploy_tool_to_site` | live, but see §4 |
| editorials / research / guides | page types + `page-content-writer` | live |
| games | **no mechanism found** | would be new build |

**So the brief-writer needs the capability catalogue as an input, not just the register entry.**
A brief that asks for something unbuildable does not fail loudly — it produces a plan with a gap.
The platform already detects this: `capability_gap` work items exist and fire (**42 raised, 41
`deferred`, measured 2026-08-19** — e.g. *"1 finding(s) from revenue_shape need agent
affiliate-link-manager, which is not registered"*). At 2,000 auto-written comprehensive briefs
that number will not stay at 42, and a queue of deferred gaps nobody reads is the same as no
detection at all.

**Recommendation:** the brief-writer picks from an explicit list of what the platform can build
today, and anything outside it goes in a `wishlist` section of the brief that is *documented but
not planned*. That keeps ambition visible without turning it into silent build failures.

## 3. This merges with RFC_037 rather than competing with it

RFC_037's ruling is that the register moves to a database, becomes the source of truth, covers
every domain, and is fed to the classifier as **advisory** input.

**The brief-writer is the more natural consumer of that data than the classifier is.** The
register entry carries the proposition, the neighbours and the must-nots; a brief is exactly
where those belong; and the classifier already reads the brief. Feeding the register to the
brief-writer gets the differentiation into every downstream agent, not just the first one —
whereas feeding it to the classifier alone leaves the strategist, briefing agent and planner
still blind to the siblings.

**This does not supersede RFC_037** — the classifier input is still worth having, and the owner
has ruled on it. But if only one gets built first, **the brief-writer is the higher-leverage
one**, and the RFC should say so rather than have two lanes build overlapping readers of the
same table.

## 4. The HITL switch the owner wants ALREADY EXISTS — and has never been used

*"a config enabled hitl step if I want it"* does not need building. `site_work_items` has an
`approval_mode` column, and the dispatcher honours it
(`load_work_item_actions.go:709`):

```sql
AND (COALESCE(wi.approval_mode, 'auto') = 'auto' OR wi.status = 'approved')
```

**An item whose `approval_mode` is not `'auto'` is not dispatched until its status is
`approved`.** That is precisely a config-enabled human gate, it is fleet-wide, and it applies to
every work-item type including a future `needs_brief`.

**⚠ And it has never once been exercised: all 10,311 rows are `'auto'`** (measured 2026-08-19).
So it is **live and undriven** — the class this estate has been bitten by repeatedly, where a
mechanism reads as available but has never had a single real caller. **Before designing around
it, set one item non-auto and prove both halves**: that it is withheld while unapproved, and
that it dispatches once approved. A mechanism with zero live callers has an untested dependency
on everything around it.

**At 2,000 domains, review cannot be per-brief.** Nobody reads 2,000 briefs. The realistic
shapes are: review the first N of a batch; review only where the register's collision check
flags a clash with a sibling; or review only briefs whose subject falls in a sensitive category
(finance, health, legal). **The second is the one that scales**, and it is an argument for
reviving the *binding* collision check that RFC_037 has just deferred as advisory-only.

## 5. The regulated-identity gate — what exists, and the owner's rule

The owner is right that something has started. Searched the whole repo 2026-08-19 and verified
the three load-bearing findings independently.

### What is live

**Migration `464` → concept-register `CGV-032`.** A prose rule inserted into
`domain-research-classifier`'s prompt: regulated business models — lending, credit broking,
"compare and apply" journeys, eligibility checkers that route into lender applications, debt
advice, mortgage arranging, insurance distribution, investment advice, payment services, claims
management, funeral plans — *"are not on the menu at all"* unless a mission brief explicitly asks
for one. It also forbids asserting an authorisation number, a regulator relationship, a panel of
providers, or a legal entity for an unbuilt brand.

**Verified live 2026-08-19:** the rule is present in the running config of
`domain-research-classifier` and **of no other agent**. It is proven on the motivating case by a
genuine A/B — loanzy run 1 produced *"Personal Loan Matching, Lender Lead Facilitation"*; run 2
under the rule produced *"Loan Explainers, Borrowing Guides, Rights and Regulations Overview"*
with `"Regulated business model explicitly excluded by platform rules"` in its own reasoning.

### The five gaps, each verified

1. **No code guards this anywhere.** No `*_guard.go`, no discovery check, nothing in `platform/`
   or `internal/`. The single control is one paragraph of prose in one prompt.
2. **No fleet-wide banned-claim pattern for regulatory status.** `claims_global.go` carries 24
   patterns, all about accuracy/reliability overclaims; **zero** mention of FCA, regulated,
   authorised or broker (verified by grep). **A new site is born with no protection against
   "We are FCA authorised and regulated".**
3. **Exactly TWO sites carry a per-site pattern that would block it** — `remortgagecalculator.uk`
   and `adversecreditmortgage.co.uk` — and only because a human hand-seeded them (verified by
   querying every current `evidence_base`).
4. **The rule is on the classifier ONLY** — not the strategist, briefing agent, planner or
   page-content-writer. If the classifier is bypassed, re-run, or its answer edited, nothing
   downstream re-asserts it. This is exactly the hole P11 named.
5. **The positive control is owed.** That the classifier *declines* is proven; that a brief which
   legitimately *asks* for a regulated model still produces one has never been tested. Half a
   guard is proven.

Also: the briefing agent **already notices** — it wrote *"FCA authorisation number — not yet
known; must be obtained before launch"* into its own gaps list — and had no authority to stop
anything, **because a gap is a note**.

### The owner's rule, and what it needs

> *"even then they should provide some sort of proof I guess… for now we reject a request until
> they write to us by email with proof."*

**There is no proof mechanism of any kind today.** The classifier's test is "does the brief say
so", which is an unverified assertion in free text — a third-party brief from a webdesign.uk
customer can simply contain the sentence. The generic `attestation` machinery that exists is for
evidence freshness (180-day staleness on owner-attested commercial facts) and keys on nothing
regulatory.

**The smallest honest implementation of the owner's rule:**
- a brief that requests a regulated model is **refused at intake**, with a stated reason and an
  instruction to email;
- approval is an **attestation record** naming the firm, its FRN, and who attested — not a flag,
  because a flag cannot be audited later;
- the site's `evidence_base` then carries the FRN as a citable fact, so the existing claims layer
  can check what the pages say against it.

**Do the cheap half first regardless:** promote a regulated-status pattern into the **fleet-wide**
banned-claim set. It costs one entry, it protects all ~17 live sites and every future one instead
of two, and it is the backstop for the four cases where the classifier's prompt is bypassed. On
its own it does not implement the owner's policy — it stops the *claim*, not the *positioning* —
but it removes the worst failure mode while the rest is designed.

## 6. Open questions for the owner

1. **Who reviews the generated briefs, and on what sampling rule?** Per-brief review does not
   scale past a few dozen.
2. **Does a third-party brief get the same trust as an owner brief?** They arrive through the
   same door and the regulated-identity test is currently "the brief says so".
3. **Games** has no mechanism. Is it wanted enough to build, or does it go on the wishlist?
4. **Should the brief-writer read the register instead of / as well as the classifier?** (§3.)
