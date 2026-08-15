# DECISION INPUT — the seven duplicate page pairs (`bugs_open/215` O2)

**Prepared 2026-08-12 for the owner's survivor decision.** This document decides
nothing. It puts the seven pairs side by side with what each version *actually
serves*, so the choice is a short conversation. Execution procedure is unchanged and
lives in `RUNBOOK_2026-08-11_duplicate_page_identity_remediation.md`.

All figures [MEASURED 2026-08-12 ~16:30Z] against the **live sites**, not the database.

## The headline, and it reverses the working assumption

The runbook's step 1 offers **component count** as the "which side has content" input,
and records robot-hands as "5/3/4 components on the bare side against 1 each on the
`tool-` side" — which reads as *the bare side is the substantial one*.

**Measured at the served page, that is wrong on most pairs.** On **4 of the 7**, the
bare side has **no interactive element at all** and the `tool-` side is the working
tool. On three of those four the component count pointed the *opposite* way:

| pair | bare side | `tool-` side | component count said |
|---|---|---|---|
| robot-hands payload-calculator | **0 inputs** | 4 inputs + form | bare, 3 vs 1 |
| robot-hands matchmatrix | **0 inputs** | 4 inputs + form | bare, 4 vs 1 |
| finetuning ai-readiness-quiz | **0 inputs** | **32 inputs** | bare, 2 vs 1 |
| ai-agent-orch llm-cost-calculator | **0 inputs**, 727 words | 8 inputs | tool-, 0 vs 1 ✓ |

A component is a container, not a quantity of content: one component holding a
calculator outweighs four holding prose. **This is the estate's own "trust the
rendered artefact, not the status" rule, applied to a count nobody thought of as a
status.** Recorded as a landmine.

## The seven pairs

`IN PLAN` marks the side the site's current plan names — it matters because archiving
an in-plan page re-arms the refile loop (runbook step 2), so the non-plan side is the
cheaper retirement.

### 1. ai-agent-orchestration.com — `llm-cost-calculator`

| | bare `/llm-cost-calculator.html` | `tool-` `/tools/tool-llm-cost-calculator.html` |
|---|---|---|
| components | 0 | 1 |
| served | 11.3 KB, **727 words, 0 inputs** | 46.4 KB, 3,369 words, **8 inputs** |
| created / last deployed | 04-18 / **04-18** | 04-09 / 08-11 21:16 |
| in current plan | no | no |

**Clearest case of the seven.** The bare side is a 727-word stub with no calculator,
last deployed in **April** and never since. The `tool-` side is the maintained tool.
**Recommend: `tool-` survives, bare retires.** Neither is in the plan, so no plan edit
is needed first.

### 2. finetuning.uk — `ai-readiness-quiz` ⚠ DECOMPOSED SITE

| | bare `/ai-readiness-quiz.html` | `tool-` `/tools/tool-ai-readiness-quiz.html` |
|---|---|---|
| components | 2 | 1 |
| served | 39.2 KB, 2,089 words, **0 inputs** | 36.3 KB, 2,430 words, **32 inputs** |
| created / last deployed | 04-18 / 08-12 04:01 | 04-04 / 08-12 04:10 |
| in current plan | no | no |

The `tool-` side **is** the quiz — 32 inputs against zero. On content this is not a
close call.

**⚠ But this site is one of the six decomposed sites (`bugs_open/204`)**, which no
prior document connected to the twin population. Its `pages.sections` hold positional
slot names, so a rebuild of either side is already unreliable. **Recommend: decide the
survivor now (`tool-`), but do not execute until 204 is fixed** — and do not enable
any identity gate here meanwhile.

### 3–4. fundamentallyai.com — the two `-guide` pairs

| | bare `/blog/…` **IN PLAN** | `tool-` `/guides/…` |
|---|---|---|
| `automation-savings-estimator-guide` | 3 comps, 27.3 KB, 2,368 words, 0 inputs | 3 comps, 28.5 KB, 2,417 words, 0 inputs |
| `model-approach-selector-guide` | 3 comps, 26.2 KB, 2,195 words, 0 inputs | 3 comps, 31.6 KB, 2,859 words, 0 inputs |
| last deployed | 08-12 14:26 / 14:36 | 08-12 14:47 / 14:48 |

**Neither side is a tool at all** — despite the `tool-` prefix, both are prose guides,
and the two versions are near-identical in length and structure. **The content choice
is close to free; the cheap discriminator is the plan.** The bare `/blog/` side is
IN PLAN, the `tool-` `/guides/` side is not.

**Recommend: bare `/blog/` survives, `tool-` `/guides/` retires** — it is the loop-safe
direction and needs no plan surgery. Counter-argument worth one moment of your time:
`/guides/` is the better-shaped URL for a guide and its pages are slightly fuller
(+2% and +30% words). If you prefer `/guides/`, the runbook's step 3 (remove the loser
from the plan first) becomes mandatory rather than optional.

**Both sides of both pairs were re-deployed today, 20 minutes apart.** The duplication
is not dormant; the pipeline is actively maintaining both copies.

### 5–7. robot-hands.com — three pairs, **both sides IN PLAN**

| pair | bare | `tool-` (dir-style `/tools/<name>/index.html`) |
|---|---|---|
| `gripper-cycle-time-estimator` | 5 comps, 46.0 KB, 3,828 words, **5 inputs** | 1 comp, 32.0 KB, 2,125 words, **8 inputs + form** |
| `gripper-payload-calculator` | 3 comps, 23.0 KB, 1,700 words, **0 inputs** | 1 comp, 34.2 KB, 2,275 words, **4 inputs + form** |
| `matchmatrix` | 4 comps, 29.0 KB, 2,214 words, **0 inputs** | 1 comp, 47.7 KB, 4,655 words, **4 inputs + form** |

**Two of the three are unambiguous**: `payload-calculator` and `matchmatrix` have **no
calculator on the bare side**. The `tool-` side is the product; the bare side is prose
about it. **Recommend: `tool-` survives for both.**

**`cycle-time-estimator` is the one genuine judgement call in the seven.** Both sides
are interactive (5 inputs vs 8 + a form) and the bare side carries nearly twice the
prose (3,828 vs 2,125 words). Either could defensibly survive; the honest options are
to keep `tool-` for consistency with its two siblings, or to merge the bare side's
prose into the `tool-` page before retiring it.

**⚠ All three have BOTH sides in the plan**, so per runbook step 2 there is **no
loop-safe archive order** — the plan must be fixed first (step 3) or the retired page
is re-created. This is the one site where the plan edit is not optional.

## What I have NOT established

- **Search-engine indexing.** The runbook names it as an input and I have no data
  source for it. On age alone the bare/flat URLs are older on 5 of 7 pairs and are the
  likelier indexed ones, but **that is an inference, not a measurement** — if any pair
  matters commercially, check Search Console before executing.
- **Inbound links: measured, and there are none.** Zero `link_registry.target_page_id`
  rows reference any of the 14 pages, re-measured today. Nothing internal breaks
  whichever side goes, and no redirect conflicts.
- **Whether the prose differs in substance**, as opposed to length, on the two
  fundamentallyai pairs and robot-hands `cycle-time`. Word counts are not a diff.

## Summary of recommendations

| # | pair | recommend keep | confidence | blocker |
|---|---|---|---|---|
| 1 | ai-agent-orch `llm-cost-calculator` | **`tool-`** | high | none |
| 2 | finetuning `ai-readiness-quiz` | **`tool-`** | high | **`bugs_open/204`** |
| 3 | fai `automation-savings-…-guide` | bare `/blog/` | medium (loop-safety, not merit) | none |
| 4 | fai `model-approach-selector-guide` | bare `/blog/` | medium (loop-safety, not merit) | none |
| 5 | robot-hands `payload-calculator` | **`tool-`** | high | plan carries both |
| 6 | robot-hands `matchmatrix` | **`tool-`** | high | plan carries both |
| 7 | robot-hands `cycle-time-estimator` | either — **your call** | low | plan carries both |

**Nothing here has been executed. No page archived, no plan edited, no redirect
written.** Execution needs your survivor decision per pair, and then the runbook's
eight steps in order.

> **UPDATED 2026-08-13 — the blocker in the original last line is DISCHARGED.** That line
> read: *"until the archived-pages diagnosis (corr `38099787`) is read, any archive can be
> undone by the next build."* The diagnosis was read, the cause was **four** independent
> producers none of which checked `pages.status`, and the fix is **live and council-approved**
> (`bugs_open/266`, PBP-042, chassis `v1.0.1295`, artefact-verified on both replicas).
> **An archive now holds.** Two caveats, neither of which changes a recommendation above:
> the guard is **not yet behaviourally exercised** (nothing has dispatched a build at an
> archived page since it went live), and it **stops recurrence without undoing** the pages
> already archived-and-serving — for those, retraction is the route, and retraction is
> deliberately unaffected by the guard.

---

# OWNER RULING 2026-08-13 — all seven decided. Recorded verbatim; nothing executed yet.

| # | pair | **OWNER'S CALL** | vs my recommendation | execution gate |
|---|---|---|---|---|
| 1 | ai-agent-orch `llm-cost-calculator` | **keep `tool-`** | as recommended | none — neither side in plan |
| 2 | finetuning `ai-readiness-quiz` | **keep `tool-`, HOLD execution** | as recommended | **`bugs_open/204`** — decided now, executed later |
| 3 | fai `automation-savings-…-guide` | **keep `/guides/`** | **DIFFERS** — I recommended `/blog/` | **plan edit MANDATORY** |
| 4 | fai `model-approach-selector-guide` | **keep `/guides/`** | **DIFFERS** — I recommended `/blog/` | **plan edit MANDATORY** |
| 5 | robot-hands `payload-calculator` | **keep `tool-`** | as recommended | plan carries both |
| 6 | robot-hands `matchmatrix` | **keep `tool-`** | as recommended | plan carries both |
| 7 | robot-hands `cycle-time-estimator` | **MERGE the bare page's prose into `tool-`, then retire bare** | I offered this as the best-outcome option | plan carries both |

**On 3 and 4 the owner took the option I did not recommend, and he is right about the
merit — my recommendation was explicitly on execution cost, not quality.** `/guides/` is
the better home for a guide and those pages are fuller (+2% and +30% words). The cost is
exactly what this document said it would be: **the `/blog/` side is IN PLAN, so it must be
removed from the plan BEFORE it is archived** (runbook step 3, now mandatory rather than
optional). Skipping that re-arms the refile loop and the retired page comes straight back.

**Pair 7 is not a retire-one-side job, it is a content merge first.** The bare page carries
~1,700 more words than the tool page. Retiring it before merging loses them. This is the
only pair whose execution includes writing.

## The five archived-and-serving pages — OWNER: **"leave them for their own lanes"**

Not this lane's to execute. Each owning lane decides its own:

- **leopardessconsulting.co.uk** `/our-approach.html` — told in `docs/leopardessconsulting/HANDOFF.md`
- **robot-hands.com** `/gripper-catalog.html`, `/news.html` — told in that lane's handoff
- **fundamentallyai.com** `/blog/ai-readiness-checker-guide.html`,
  `/tools/llm-cost-calculator/index.html` — **still needs a call from the fundamentallyai
  sweep front**, which owns that site's execution (handoff §5). Routed there, not decided here.

**These five are a SEPARATE population from the seven pairs above** — no page appears in
both lists. Do not conflate them when executing.

## What is now unblocked, and what is not

`bugs_open/266` is fixed, council-approved and live (`v1.0.1295`), so **an archive holds** —
the standing "any archive can be undone by the next build" warning is discharged. Remaining
gates are per-pair and listed in the table: `204` for pair 2, and a plan edit for pairs 3–7.

---

# OWNER RULING 2026-08-14 — pairs 3+4 REVERSED, on the redirect finding

**Trigger:** execution of pair 1 found that **there is no redirect mechanism** (RUNBOOK
`⚠ CORRECTION 2026-08-14`). Retiring a URL 404s it. The 08-13 ruling on pairs 3+4 was taken
partly on my assurance that a redirect would protect the retired side; that assurance was
false, so the owner was asked again with the true trade.

| # | pair | 08-13 ruling | **08-14 RULING** | why it changed |
|---|---|---|---|---|
| 3 | fai `automation-savings-…-guide` | keep `/guides/` | **keep bare `/blog/`** | `/blog/` is the older, likelier-indexed URL; with no redirect, retiring it would 404 the indexed side |
| 4 | fai `model-approach-selector-guide` | keep `/guides/` | **keep bare `/blog/`** | same |

Pairs **1, 5, 6, 7 stand as ruled on 08-13** (keep `tool-`; pair 7 merge-then-retire), and
pair **2 remains decided-but-held** on `bugs_open/204`. The owner accepted the 404 for those,
where the retired side is a stub or a non-interactive prose page rather than the indexed one.

## What the reversal simplifies

**Step 3 (plan surgery) is no longer mandatory for pairs 3+4.** It became mandatory only
because the 08-13 choice retired the IN-PLAN side. Reverting to `/blog/` retires
`/guides/`, which the plan does **not** name — the loop-safe direction, and the reason this
was my original recommendation. **Re-verify plan membership at execution anyway**; it was
measured 2026-08-12 and the fundamentallyai plan has since been touched by other lanes.

**Pairs 5, 6, 7 still require step 3** — robot-hands carries both sides of all three in its
plan, and that is unchanged by this ruling.

## Standing consequence for all seven

There is no redirect. Every retirement is a 404. The owner has accepted that for the five
proceeding pairs on the grounds that the retired side is in each case the *worse* page, not
the indexed one. **If a redirect capability is ever built, pairs 3+4 are the ones worth
revisiting** — `/guides/` is the better URL and was the owner's preference on merit.

---

# OWNER RULING 2026-08-15 — pair 7's merge, scoped. The 08-13 ruling stands; this decides its CONTENTS.

**Trigger:** executing pair 7's merge established that the bare page's "~1,700 extra words"
are not prose awaiting an author. They are five discrete components, of which two carry
essentially all the words and move verbatim. The owner was offered three options and chose
the first.

| # | option | words moved | chosen |
|---|---|---|---|
| A | **explainer + FAQ only** | **1,587**, verbatim, zero links, no authoring | ✅ **CHOSEN** |
| B | + the closing CTA | 1,675 — CTA points at the survivor, needs a new destination | no |
| C | + the hero as well | 1,763 — hero duplicates the tool's heading and carries a live defect | no |

**What is deliberately left behind, and why it dies with the bare page:**

- **hero (88 w)** — the survivor's tool component already carries its own heading, so the
  hero would put two competing headings on one page. Its button also reads *"Run the
  Estimator"* while pointing at `/contact.html`: a defect **already live on the bare page
  today**, which porting would have carried across. Not this lane's to fix; it disappears
  when the bare page is retired.
- **call-to-action (88 w)** — its primary button points at
  `/tools/gripper-cycle-time-estimator/index.html`, i.e. **at the survivor**. On the bare
  page that is correct and useful; moved to the survivor it is a link to itself. Repointing
  it is a commercial choice about where a finished user should go next, not a mechanical
  one, so it was not made by a session.

**Consequence worth noting:** the survivor has **no CTA at all**, before or after this
merge, and once the bare page is retired the site loses that closing prompt for this tool
entirely. That is the status quo for the survivor and was not changed here — but if a CTA
on the tool page is wanted, it is a separate, small, decided-by-you piece of work.

**This ruling does NOT change the 08-13 decision** ("MERGE the bare page's prose into
`tool-`, then retire bare"). It scopes what "the prose" means. The retire half is
unchanged and outstanding — see `HANDOFF_2026-08-12_215_quiet_mode_continue_here.md` §18.4.
