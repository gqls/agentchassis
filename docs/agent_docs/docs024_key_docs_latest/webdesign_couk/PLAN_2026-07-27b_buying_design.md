# PLAN — "Buying design": the enterprise buyer section, and where the site is really going

**Written 2026-07-27**, replacing the buyer-track design in
`PLAN_2026-07-27_phase2_buyer_track.md` after the owner redirected it the same
day. That earlier plan is kept, marked superseded, because how the audience was
first mis-sized is itself useful.

**This is a proposal built on the owner's direction, including the parts he asked
me to improve rather than transcribe.** Where I am arguing rather than recording,
it says so.

---

## 1. What changed, in the owner's terms

> *"It pitches against the multimillion or multi-hundred-thousand pound web design
> buyer… They probably will be focused on the scalability and reliability as much
> as the brand proposition and expressing their offline brand online, or being
> able to extend their online brand further online without diluting the strengths
> of it."*

Three rejections, each with a reason worth keeping:

- **`Hire` is dead.** *"A bit cheap-sounding and sounds like we're hiring an Upwork
  or Fiverr worker, which is completely the wrong end of the quality and price
  scale."* The label was setting the price expectation before a word was read.
- **"How to judge a quote" is dead**, and for a better reason than the risk I
  raised. My objection was that we have no verified figures. The owner's is
  stronger: *"it focuses the content on money and not design."* At £500k the buyer
  is not price-shopping — they are de-risking a decision they will be held to.
- **The general small-business buyer is dead.** The audience is enterprise.

And one addition that changes the section's character entirely:

> *"Fully AI aware. We can lead with the strengths and limitations of the solutions
> out there and include our own and be fully truthful."*

## 2. The audience, sized properly

Not "people who want a website". A £100k–£multi-million commissioner: marketing
or brand director, CDO, digital lead, with procurement alongside and a board
behind. Usually running a formal supplier selection, often with an incumbent
agency, and personally exposed if it goes wrong.

**Their real fear is not overpaying. It is choosing wrong and being unable to
show they chose carefully.** Everything below follows from that.

What they are drowning in: agency marketing that is all self-assessment; AI
vendor claims; procurement-consultant content selling procurement consultancy.
What they cannot find: **a neutral account written by someone with no stake in
which supplier wins.**

## 3. Positioning — the one thing this site can say that others structurally cannot

[ARGUED, not recorded — this is the part I most want challenged.]

We operate an AI web-build system in production across roughly a thousand sites.
That gives us standing nobody in this market has:

- **An agency has an incentive to say AI changes little** — it protects day rates.
- **An AI vendor has an incentive to say it changes everything** — it sells seats.
- **We run one, and we are not bidding for the buyer's project.** We can say
  exactly where it works, exactly where it fails, and show the receipts.

**The honesty is the product.** To a buyer who is being pitched AI by everyone,
the single most credible thing they can read is a detailed, specific account of
where an AI build system *gets it wrong* — written by the people running it. That
is more persuasive than any capability claim, because their dominant experience
of this market is being sold a story.

We have unusually concrete material for this. This project has, on the record:
invented statistics twice; shipped a component truncated from 10,272 characters to
1,253 and reported success; published a homepage whose main call-to-action was a
dead link while an automated check reported the site clean. **Those are not
embarrassments to hide from this audience — they are the exact texture of "what
actually goes wrong when you build this way", and no competitor will publish
theirs.**

> **OWNER DECISION NEEDED — this is a business call, not mine.** The positioning
> above is only differentiating to the extent we publish things most companies
> conceal. There is a spectrum: (a) speak about failure classes generically,
> (b) publish specific anonymised cases, (c) publish our own named failures with
> evidence. **(c) is the most credible and the most exposing.** I would not
> publish anything at (c) without an explicit decision, and the decision should be
> made once, in writing, rather than page by page.

## 4. Naming

**Proposed: `Buying design`** — the owner's own suggestion, and better than mine.
Plain, sets no price anchor, and says the section is about the *act of
commissioning* rather than about us.

[ARGUED] One alternative worth a moment: **`Commissioning`**. Its advantage is
self-selection — a £5k buyer does not use that word, a £500k buyer's procurement
team does, so the label filters the audience before the click. Its cost is that
it reads as formal and slightly institutional. **My recommendation is still
`Buying design`**: self-selection is better achieved by the content's register
than by a word that risks reading as stuffy. URL `/buying-design/`.

## 5. Content pillars — the owner's list, improved

His starting list was: what makes the buyer's project successful; the elements of
a top campaign; AI and the top-tier website. All three survive. Reordered by what
a buyer needs first, and extended where I think there are gaps:

| # | pillar | why it earns a place |
|---|---|---|
| 1 | **AI and the top-tier website: what actually changed, and what didn't** | The flagship, per §3. Every board is asking it; nobody neutral is answering it. |
| 2 | **Why large web projects fail** | The failure modes are consistent and rarely written honestly: decision rights unclear, brand signed off late, **content not ready** (the real killer), design by committee, no performance budget. Written from the delivery side, not the pitch side. |
| 3 | **Expressing an offline brand online without flattening it** | The owner's point, and genuinely under-served. Most writing on this is agency portfolio-bait. |
| 4 | **Extending an online brand without diluting it** | His second point — sub-brands, campaign sites, international rollout, where the brand system stops holding. |
| 5 | **Scalability and reliability as commitments, not adjectives** | [ARGUED — my addition, and I think the most immediately useful page.] Every agency says "scalable and reliable". This page turns those into things you can hold a supplier to contractually: performance budgets, conformance level, uptime, recovery, security posture. It converts an adjective into a clause. |
| 6 | **What you should own at the end** | [ARGUED — my addition.] Code, design system, CMS licences, data, domain, analytics. Lock-in is the most expensive mistake at this tier and the least discussed before signature. |
| 7 | **Running a selection that surfaces substance, not theatre** | How to structure a pitch so you learn something. Neutral, and only credible from someone not pitching. |
| 8 | **Accessibility as a board-level duty** | Equality Act 2010 / WCAG 2.2. Primary sources only — this is the class the vetcomparison workstream was burned by. |

**Deliberately absent:** anything shaped like "how much should this cost".

## 6. Tools for this market — the part the owner asked me to think hardest about

> *"My guess is that the early tools should be simple to use and have strong impact
> and that they may come back and back to them during their pitch for supplier
> process."*

**That guess is right, and it has a sharper form.** The test for a tool here is not
"is it useful" but **"does the same tool get used at three different stages of the
selection?"** A buyer's journey runs: build the internal case → write the brief →
longlist → evaluate pitches → contract → oversee delivery. A tool that serves one
stage is a novelty; one that serves three becomes the reason they keep the tab
open for four months.

Assessed against that test, and against what we can actually build:

### T1 — Side-by-side site benchmark *(the anchor)*
Enter your own site plus up to five others — competitors, or the portfolio work an
agency is claiming. Get objective, comparable results.

**Three returns across the journey:** stage 1 (evidence for the board that the
current site is behind), stage 4 (does this agency's *actual delivered work* stand
up, as opposed to the case study about it), stage 6 (is what we just paid for
measurably better). **That is the strongest repeat-use case in the set.**

**Buildable now, mostly.** `internal/adapters/browserrunner/` already runs real
headless Chromium with desktop and mobile profiles and checks
`no_horizontal_overflow`, `no_console_errors` and `page_status_ok`
(`run_checks_action.go`), plus screenshots. That is a genuine foundation — **but
it is not a performance or accessibility suite, and I should not imply it is.**
Page weight and load behaviour would be new work.

### T2 — Accessibility duty check *(cheapest strong move — build first)*
Contrast, touch-target size, focus visibility, run against any URL, framed as
legal exposure under the Equality Act rather than as a developer checklist.

**Nearly free, and that is the point.** We already have these as practitioner
tools — the contrast tools, *The 44px Rule*, *The Invisible Focus*. The buyer
version is the same measurement with a different frame and a different output: not
"your contrast ratio is 3.9:1" but "these fourteen elements would fail an
accessibility audit, here is what that means for you". **One engine, two
audiences, two framings** — which is exactly "add rather than remove", and it
makes the practitioner library an asset to the buyer section rather than a
distraction from it.

### T3 — Brand consistency across a site
Extract the palette, type scale and spacing actually in use across a set of pages
and show the drift. Stage 1 ("our own site has quietly fragmented") and stage 6
("did they deliver a system, or twelve pages?"). We have palette and design-token
machinery; this is a real but bounded build.

### T4 — Brief scaffold
Structured questions in, a usable brief out. Lower priority as a tool — **but note
it is the same interaction shape as the website-creation form in §8, so building it
teaches us that, and it may simply become the first step of it.**

### T5 — Ownership & handover checklist
Interactive, exports a document procurement can attach to a contract. Cheap, and
it is the physical artefact of pillar 6.

**Suggested order: T2 → T1 → T3 → T5 → T4.** T2 because it is nearly free and
carries legal weight; T1 because it is the anchor and the hardest to copy.

### The rail these tools must not cross

**They run on URLs the buyer supplies, and the results are shown to the buyer.**
We do **not** publish league tables ranking named agencies or their clients'
sites. Publishing comparative quality judgements about named real firms is a
different risk class from pointing at software — reputational and potentially
legal — and it would destroy the neutrality that makes the section worth reading
in the first place. A buyer running our tool on an agency's portfolio privately is
research; us publishing the same output is a hit list.

**This also puts D11 under new light.** A "best third-party tools" directory is
uncontroversial. A *UK agency* directory at this tier is not, for exactly the
reason above. D11 (editorial only, never affiliate) still stands, but the
directory should be understood as **tools, not suppliers**, unless the owner rules
otherwise.

## 7. Figures — the rail, and where verified numbers could actually come from

Owner: *"We should look for verified figures from the top agencies, but no figures,
no page."* **Unchanged and now stronger — at this tier an unsourced number does
more damage than a missing one**, because the audience is professionally trained
to check.

[ARGUED] Where genuinely citable figures could come from, in descending order of
solidity:

1. **Companies House** — UK agency size, turnover, headcount from filed accounts.
   Actually verifiable, actually public, and almost nobody uses it in this
   context. **Caveat, checked not assumed:** the vetcomparison workstream has
   related company-number matching machinery, but it currently cannot record
   provenance (`bugs_open/100`/`101`) — so treat reuse as a *lead*, not a
   dependency, until that is resolved.
2. **HTTP Archive / Chrome UX Report** — real public performance data on real
   sites, including the buyer's own and their competitors'. Strong and free.
3. **W3C / WCAG, ICO, legislation.gov.uk** — for anything legal. Primary source or
   it does not ship.
4. **Published agency and industry reports** — usable only with the source and its
   method named, because they are marketing with a methodology section.

**Explicitly not acceptable:** a figure recalled, inferred, or carried over from
another page without a live check.

## 8. Recorded future direction — the website creation form (NOT this phase)

The owner's stated next build after this section:

> *"A website creation form using our system. So we would migrate the site onto a
> VM and use the tools-api probably. There will be new build for integrating it
> into our chassis but we can look at that when we get there, probably we stand up
> a copy chassis in its own cluster with its own database."*

Recorded now so the buyer section is designed compatibly, **not started**.

Grounded checks, so the next thread starts from fact:
- **`cmd/tools-api` and `internal/tools-api` exist**, with one live endpoint
  (`/api/v1/tools/gauntlet`, `internal/tools-api/api/server.go:23`). So adding a
  `/api/v1/tools/…` endpoint is an established pattern with exactly one worked
  example — thin, but real.
- **A VM estate workstream exists** (`vm_estate/`), and `idea_uk_vm_site/` is the
  precedent that matters: a VM-hosted site that has run the full chain to a
  completed paid transaction in production.
- **The owner's isolation constraint is the important architectural bit** —
  separate cluster, separate database. That is a decision, and it should not be
  quietly eroded into "just another site on the main chassis" when it is built.

**Design consequence for now:** T4 (brief scaffold) is the same shape as this
form. Build T4 so it could become its first step.

## 9. The strategic tension I need to name

[ARGUED — the owner said two things that pull in different directions, and it is
better to say so than to average them.]

> *"For now we need to get a whole load of webdesign related traffic any way we
> can and to get away from duplicated content from the other sites."*
> *"It pitches against the multimillion… buyer."*

Broad traffic and a tiny high-value niche are different games. They coexist only
if we are deliberate about which asset does which job:

- **The 94 practitioner pages are the traffic engine.** They rank for tool and
  how-to queries. High volume, low commercial intent. Keep them, improve them.
- **"Buying design" is the positioning layer.** Low volume, very high value per
  visitor. It does not need traffic to be worth its cost.
- **The risk is register collision.** An enterprise buyer landing on *Fractional
  Layouts: The Math of CSS Grid* concludes this is a developer site and leaves.
  So the buyer section needs its own unambiguous front door, and the home page has
  to serve both without confusing either. That is a real design problem, not a
  copywriting one.

**And a reframe that raises the rewrite's priority.** Because the two source
domains stay live, the 94 imported pages *duplicate* them — so W2's copy rewrite
is doing double duty: quality **and** de-duplication. Meanwhile **the buying-design
section is the only part of the site with zero duplication risk, because it is
entirely new writing.** The owner's instinct to build it therefore serves the SEO
goal directly, not just the positioning one.

## 10. Open for the owner

1. **§3 — how exposing do we get?** (a) generic failure classes, (b) anonymised
   cases, (c) our own named failures with evidence. Most credible is most
   exposing. One decision, in writing.
2. **`Buying design` confirmed**, or `Commissioning` for its self-selection?
3. **§5 pillars** — which of the eight, and what is missing? Pillars 5 and 6 are
   mine, not yours.
4. **§6 tool order** — T2 first because it is nearly free, T1 because it is the
   anchor. Agreed?
5. **§6 rail** — confirm we never publish comparative rankings of named agencies,
   and that D11's directory means **tools, not suppliers**.
6. **Do the designers stay?** You said *"ultimately we will not focus on the
   designers"*. §9 argues they are the traffic engine and should stay as such,
   without further investment beyond the rewrite. Worth an explicit call, because
   it decides whether W2's practitioner half is a rewrite or a holding action.
