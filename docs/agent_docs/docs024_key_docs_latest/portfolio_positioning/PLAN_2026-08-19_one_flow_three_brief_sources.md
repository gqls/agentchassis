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

## 2. ~~What "comprehensive" has to mean~~ — CORRECTED: the planner already knows, and games are buildable

> **⚠ CORRECTED 2026-08-19, twice, both times by the owner and both times verifiable in one
> query. This whole section rested on two claims of mine that are false.**
>
> **(a) "The brief-writer needs the capability catalogue, because the planner does not know what
> we can build." WRONG.** `build-site-planner` has a **`load_components`** step that queries the
> live library — `SELECT name, display_name, function, category, description FROM
> content_components WHERE is_active = true …` — and `plan_site`'s `input_fields` are
> `["input_data","site_specs","available_components","available_styles","existing_pages"]`.
> The planner is handed the catalogue every run. Nothing needed building; I had not looked.
>
> **⚠ But the query has a restriction worth knowing, and it connects to `bugs_open/311`:**
> `component_level IN ('section','element')` is unconditional, while `component_level='tool'`
> rows are included **only if** the site already has `structure.plan_includes_tools = true`
> **and** the tool is already on one of that site's pages. So on a **greenfield** build the
> planner can see no library tool at all. That is exactly why tools arrive afterwards via
> `tool-suggester` → `add_tool`, and it is the upstream half of 311: the planner, unable to see
> `tool-mortgage-repayment`, names a bespoke section instead, and the generator then collides
> with another site's component of that name.
>
> **(b) "Games have no mechanism." WRONG.** Owner: *"games can be created in the framework."*
> Confirmed: the tool framework generates arbitrary interactive components with real JS —
> `tool-drop-rate-tuner` (22,230 chars), `tool-xp-curve-designer` (17,787),
> `tool-gacha-pity-designer` (13,161), all live on `gamesdesign.co.uk`, plus section-level
> `game-list` and `game-master-explanation` and an interactive `tool-archetype-taster-quiz`.
> **A game is an interactive tool**, and that is a mechanism the estate already runs.
>
> **What survives from this section:** only the `capability_gap` observation — 42 raised, 41
> `deferred` — and it survives for a different reason than I gave. It is not evidence that
> briefs over-reach; it is evidence that **when the planner does hit a genuine gap, the record
> of it goes nowhere.** That is still worth fixing before 1,500 briefs multiply it.

## 2b. Superseded reasoning (kept so the correction is legible)

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

> **⚠ OWNER RULING 2026-08-19 — this section's recommendation is NARROWED, and the principle is
> better than what it replaces.** *"I don't think that brief writer necessarily needs to know
> what we can actually build though it would help I guess. It can be an aspirational brief.
> **The spec is aspirational, the plan is achievable.**"*
>
> That is a cleaner separation than the one I proposed and it belongs in the architecture, not
> just in this decision: **the brief says what the site SHOULD be; the plan says what we can
> build today.** Constraining the brief to today's capabilities would bake a snapshot of the
> platform into a document meant to outlive it — and would quietly delete the evidence of what
> we are missing, which is the most useful thing a corpus of 1,500 briefs could give us.
>
> **So the gap moves from the brief to the PLANNER, and that is where it must be handled.** Two
> consequences a builder must not skip:
> 1. **The planner must degrade explicitly, not silently.** A brief asking for a games section
>    should produce a plan without one *and a recorded reason*, not a plan that quietly omits it.
>    `capability_gap` is the existing carrier and it already fires — **42 raised, 41 `deferred`**
>    (measured 2026-08-19). A deferred queue nobody reads is the same as no detection at all, so
>    the volume from 1,500 aspirational briefs makes triaging that queue a prerequisite, not a
>    nicety.
> 2. **The aspiration becomes the roadmap.** If 300 briefs ask for games and nothing can build
>    one, that is the single best-evidenced feature request the estate could produce. It is only
>    worth anything if the gaps are aggregated and read — otherwise the ambition is written down
>    and thrown away.
>
> The wishlist idea below is therefore unnecessary: **the whole brief is the wishlist**, and the
> plan is the achievable subset of it.

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
3. ~~**Games** has no mechanism…~~ **ANSWERED by the aspirational-spec ruling:** briefs may ask
   for it, the planner degrades explicitly and records a `capability_gap`. The live question is
   now *who reads the gap queue*, since 41 of 42 sit deferred.
4. **Should the brief-writer read the register instead of / as well as the classifier?** (§3.)

---

# ADDENDUM 2026-08-19b — owner decisions, and the register question evaluated properly

## A. Review is PER-BRIEF with annotation, not sampling

Owner: *"I'd like to briefly look at each one. I may have a few words of direction on many of
them that I'd like to add to or change with the automated brief."*

This supersedes §6 question 1 and it changes the mechanism, not just the volume:

- **The brief is not generated-then-approved, it is generated-then-EDITED.** A pure
  approve/reject gate is the wrong shape; the owner needs to add or change wording. So the
  artefact must be reviewable and writable in its stored form, and the edit must survive the
  build rather than being overwritten by a re-run of the generator.
- **`approval_mode` gives the hold, not the edit.** The existing switch
  (`load_work_item_actions.go:709`) withholds a work item until `status='approved'` — that is
  exactly the pause needed and it needs no building. What it does NOT provide is a place to put
  the owner's words. That is a `mission_brief` spec revision, and `site_specs` already supersedes
  rather than mutates, so the owner's edited version becomes the current one with the generated
  version preserved underneath. **The provenance matters**: a brief the owner touched should be
  distinguishable from one he waved through, so the edited revision should record that.
- **Ordering falls out of the volume.** ~1,500 briefs read "briefly" is still ~1,500 decisions.
  Generate in batches, hold the batch, and let the owner work a queue — the generator must not
  run ahead of the review or the queue becomes the thing nobody reads (which is precisely what
  happened to `capability_gap`: 42 raised, 41 deferred).

## B. Third-party briefs need SECURITY and REASONABILITY screening — and they are a different population

Owner: *"A third party brief will need security and reasonability screening."*

Worth separating the two, because they fail differently:

- **Security** — the brief is untrusted text that reaches an LLM prompt and then a build.
  Prompt-injection ("ignore your instructions and…"), attempts to name internal systems, links
  to material we should not fetch, and anything trying to steer the site toward a regulated
  identity (now backstopped by CGV-033, but the brief is where the ask arrives).
- **Reasonability** — is this a site we are willing to build at all? Legality, sector, claims it
  wants to make, and whether it is coherent enough to build from.

**And they are not on our domains.** Owner: *"I don't want to use my domains for their sites."*
That draws a clean boundary this plan did not have: **the positioning register covers OUR estate
only.** A third-party site is outside it entirely — no register entry, no neighbour rule, no
collision invariant, because it does not compete with our portfolio for our benefit. It also
means the register-reading question below applies to our briefs and not to theirs.

**Consequence:** third-party sites need domains, which we do not own. Recorded as a separate
workstream in §D.

## C. Should the brief-writer read the register? — EVALUATED, and the answer changed

The owner asked for this to be worked through rather than asserted. Having done so, I think the
answer is **yes, and it should REPLACE the classifier input rather than sit beside it** — which
is a stronger claim than §3 made.

**The argument that decides it is new, and it comes from decision A above.** The owner is going
to read every brief. The brief is the human-readable artefact he reviews; the classifier's
prompt input is not. **Positioning that lands in the brief is positioning he can see and
correct. Positioning fed straight to the classifier is invisible to him** — it would shape the
site with no point at which a human could disagree with it. For an estate whose entire premise
is that 1,500 domains are 1,500 different businesses, putting the differentiation where the
owner can edit it is worth more than putting it one step earlier.

Three supporting reasons:

1. **Reach.** The classifier is one agent. The brief is read by the classifier AND inherited by
   the strategist, briefing agent and planner through the specs derived from it. Feeding the
   register to the brief-writer gets the neighbours and must-nots into every downstream
   decision; feeding the classifier leaves the rest blind, which is the gap `RFC_037` §4 itself
   identifies.
2. **Risk.** `RFC_037`'s change adds an input to `classify_and_extract` — a shared seam every
   fleet site passes through, needing a council round and careful inertness for the ~40
   non-register sites. A brief-writer is a NEW agent that nothing depends on yet. Same
   information, materially lower blast radius.
3. **One reader, not two.** Two consumers of the same register table, with different shapes and
   different update paths, is the drift class `099_SYNC_gate_roster.py` exists to prevent.

**What this costs, stated because it is a real loss:** the classifier would no longer see its
siblings *as siblings*. If a binding collision check is ever wanted — and §A's volume argues for
one, since review-on-collision is the only sampling rule that scales — it needs sibling data at
a point where it can fail a classification. That is an argument for keeping `RFC_037` open as
the home for the **binding** check, while the **advisory** half moves to the brief-writer.

**Recommendation:** build the brief-writer as the register's reader; narrow `RFC_037` to the
binding collision check and leave it unbuilt until the brief-writer has run enough to show
whether convergence still happens. Do NOT build both readers.

## D. NEW WORKSTREAM — finding and buying domains for third-party customers

Owner: *"We still need to look at search and buy domains for the third parties — I don't want to
use my domains for their sites. That might be a completely separate workflow."*

Agreed that it is separate, and it is genuinely a different kind of thing from anything this
lane does. Recording the shape so it is not lost:

- **Search** — availability lookup across registrars, plus a name-suggestion step given the
  customer's brief. The brief-writer's output is a natural input.
- **Buy** — a spend action with a real financial consequence, on a customer's behalf. Nothing in
  this estate currently spends money; that is a first, and it should be a deliberate one.
- **Ownership and handover** — whose account holds it, what happens if the customer leaves,
  who renews it. This is the question that decides the design, and it is commercial, not
  technical.
- **What already exists to build on:** `scripts/domains/classify_nameservers.py` (is a domain
  live or parked, from public DNS), the Cloudflare zone + worker-route recipe in
  `RUNBOOK_dns_pointing_a_domain_at_the_serving_worker.md`, and the registrar credential state
  in `domains_cloudflare_rollout/`. **What does not exist:** any availability search, any
  purchase path, and any registrar API key beyond Nominet's EPP (Dynadot / Porkbun / Spaceship
  keys are still outstanding).

**Not started. Needs its own lane and a commercial decision first.**
