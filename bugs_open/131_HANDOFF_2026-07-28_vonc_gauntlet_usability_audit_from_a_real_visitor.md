# 131 — vonc.com Gauntlet: eight defects from the first real visitor session

**Filed:** 2026-07-28 · **By:** gauntlet_dead_cta, from the owner's own use of the
live site · **Severity:** MIXED — one HIGH (A: headline invisible; long-standing, NOT a
regression — see the correction in §A), two MEDIUM structural, the rest
design/product · **Status:** OPEN — A fixed, B–H open

## Why this file matters more than its severity suggests

Every item here was found by **a person using the site**, not by any check in the
fleet. Several had passed automated verification hours earlier. Two of them — A
and B — are cases where **the check that should have caught it is structurally
incapable of doing so**, and B's blind spot is fleet-wide.

The site was declared honest and working on 2026-07-27, and it is. This is the
gap between *"nothing on this page lies"* and *"a person can tell what to do"*.

---

## A. The word "Gauntlet" is invisible — 1.00:1 contrast [HIGH] — FIXED 2026-07-28

`.gi-title-accent` renders `rgb(109,40,217)` on a resolved background of
`rgb(109,40,217)`. **The same colour. Contrast ratio 1.00:1** against a 4.5
threshold. Owner's screenshot shows the headline reading "Enter the ______".

> **CORRECTED — I first filed this as a regression I caused today. It is not.**
> `.gi-title-accent` has always been `color: var(--color-primary)`, and this
> section's background has always been `--color-primary` too. They resolve to the
> same colour whatever that variable holds. **With the previous `#7c3cff` the
> ratio was 1.34:1 — also invisible.** My palette change altered the shade and
> not the relationship. Filing my own change as the cause would have sent the
> next reader to the wrong place; the fault is that an accent is painted with its
> own background's token.

**Cause.** `.gauntlet-interface-section .gi-title .gi-title-accent { color:
var(--color-primary) }` while `.gauntlet-interface-section` has
`background: var(--color-primary)`.

When I changed the palette earlier that day I verified it with computed font
sizes, brace balance, line counts and page-level overflow. **Not one of those
could see a foreground and background being the same colour** — which is why this
survived a day of checks and was found by a person looking at the page.

**FIXED and live 2026-07-28.** Repointed to `--color-stage` (`#f59e0b`), an
existing site token, measured **3.31:1** against `#6d28d9` — passes AA for large
text (44–64px) and stays distinct from the white run of the headline. Candidates
measured rather than guessed: `--color-accent` #fc5c7d was **2.36:1 (FAIL)** and
`--color-primary-on-dark` #a78bfa **2.61:1 (FAIL)**; plain white reads 7.10:1 but
would erase the accent entirely.

**Verified live**, desktop and mobile: `rgb(245,158,11)` on `rgb(109,40,217)` =
**3.31:1**, 0 raw placeholders.

**The general defect remains open elsewhere:** any component whose accent token is
also its background token has this fault latent. Worth a fleet grep for
`color: var(--color-primary)` inside a section whose background is the same var.

---

## B. Content bleeds off-screen on mobile — and `no_horizontal_overflow` CANNOT detect it [MEDIUM, structural] — PAGE-SIDE FIXED, CHECK-SIDE OPEN

Measured live, Chromium, four widths:

| page | 320 | 360 | 390 | 414 | offending |
|---|---|---|---|---|---|
| `/` | ✓ | ✓ | **14 elements** | **14 elements** | `.brief-explanation__text-col` at **437px** |
| `/about.html` | **9** | **9** | **9** | **9** | a `TABLE` at **560px** |
| gauntlet / arena / provocations | clean | clean | clean | clean | — |

**The structural part.** On every one of those pages
`document.scrollWidth - document.clientWidth === 0`, because a parent clips. That
expression is exactly what the platform's `no_horizontal_overflow` acceptance
check computes — **so the check passes on a page whose content is visibly cut off**.
It detects *page scroll*, not *content overflow*, and those are different faults.

> **ROOT CAUSE FOUND, and the homepage half WAS mine.** Raising
> `--font-size-base` 16→20px the same morning scaled every **rem-based padding**
> by 25% as a side effect. `.brief-explanation-section`'s side padding went
> 64px→80px EACH SIDE, squeezing its container 262px→230px while its grid track
> stayed 437px. **Proven causal with a reversible toggle:** forcing the root back
> to 16px takes over-wide elements 14→0; back to 20px returns it to 14.
> **The type scale and the gutter are different decisions and must not be welded
> together** — that is the transferable lesson, and it applies to every site whose
> stylesheet expresses gutters in `rem`.

**FIXED page-side and live 2026-07-28** (homepage): mobile gutters capped and
`min-width: 0` added so a grid child can shrink below its content's intrinsic
width. Measured at 360/390/414px: **over-wide 14 → 0**, text column 112%
(overflowing) → 89–90%. Item D fixed in the same pass: the gauntlet's readable
column 74% → **83%**.

**A wrong turn worth recording:** my first attempt also added `0.75rem` padding
to `.gi-container`, which already had **zero**. That NARROWED the column 74%→71%
— the opposite of the goal. Caught by measuring after, not by reasoning before.

**STILL OPEN:**
1. `/about.html`'s 560px table — untouched, needs its own scroll container.
2. **The check**: `no_horizontal_overflow` should also assert no element's
   `getBoundingClientRect().right` exceeds the viewport. One extra clause; it
   would have caught both pages. This is fleet-wide and worth more than the two
   page fixes.

---

## C. "Enter the Gauntlet" reveals nothing [MEDIUM]

Measured on a 390px viewport. Before pressing: the provocation is **already on
screen** (pre-rendered from the feed), timer reads `20:00`, status says "No round
started yet". After pressing: status changes, round state changes, the timer
starts counting, the page scrolls 512px. **The provocation does not change,
because it was already there.**

So the primary CTA's entire visible effect is *a clock starting*. Nothing is
revealed, nothing is unlocked. The owner's words: *"doesn't do anything useful."*

**Compounding:** the button sits at **y ≈ 1913px** on that viewport — roughly two
and a half screens down. A visitor meets the whole page before meeting the thing
that starts it.

**Fix candidates (design, needs a decision):** either the button moves to the top
and the provocation is revealed *by* pressing it, or the button stops being the
entry point and filing a position is the only entry (the JS already starts a
round automatically when a position is filed with no round). **Do not do both** —
two entry points is why this is confusing.

---

## D. The readable column is a narrow strip on mobile [MEDIUM]

`.gi-challenge-text` measures **288px of a 390px viewport — 74%**. Desktop is 62%
(788 of 1280). Owner: *"it renders as a narrow column down the middle which looks
like a mistake."*

On a phone a text column should be near full width; the side gutters are costing
a quarter of the screen for no reading benefit at that measure.

**Fix:** reduce horizontal padding at mobile breakpoints specifically. Note the
site stylesheet's `--container-pad-x` and the component's own padding both apply —
check which dominates before changing either.

---

## E. The provocation does not read as the thing you must answer [MEDIUM, design]

Owner: *"someone trying to figure out how the site works will skip right past it
thinking it's AI slop text and not realise that it **is** the challenge."*

Type size is no longer the problem — it was raised to 46px (35 mobile) with its
label at 20px on 2026-07-28. **The problem is that nothing frames it as a
question addressed to the reader.** It sits in a panel that looks like editorial
copy, on a page where several other blocks also look like editorial copy.

**Fix direction (not a text change — the owner was explicit):** make the
provocation structurally distinct from every other block — its own container
treatment, an explicit "you are arguing against this" affordance adjacent to it,
and the input immediately beneath it rather than a scroll away.

---

## F. The page is busy; the job is not obvious [MEDIUM, design]

Owner: *"the site is very busy with lots of sections and bits and pieces
everywhere — which is not a bad thing in itself — but it needs to be organised
and make it much clearer what you're there to do, not necessarily by changing the
text but by improving the design — making everything just a bit tighter."*

This is the parent of C, D and E. The page currently presents: hero, rules card,
provocation panel, position step, opponent panel, defence step, verdict panel,
timer, objectives, progress bar — with no visual ranking between "read this",
"do this now" and "this will happen later".

**Not actionable as a code change until someone decides the hierarchy.** See §H.

---

## G. FEATURE — a won verdict should be recorded somewhere [owner request]

Owner, after winning: *"I'd like it to be posted somewhere probably or somehow
recorded perhaps."*

**This is not the fabrication class** that was stripped from this site. A visitor's
own real verdict, on their own real round, is a true fact about something that
actually happened. Recording it is honest; **inventing a leaderboard population is
not.** The distinction is the same one already documented for the Gauntlet's
`sessionStorage` resume.

**Constraint that must survive:** the site has essentially no traffic (the
engine's request log holds only our own calls), so anything resembling a
leaderboard would be a room with one person in it, presented as a crowd. Options
in increasing cost:
1. Persist the round's verdict client-side and let the visitor keep/share a
   permalink to their own round (the round already has a real `round_id` and rows
   in the island DB).
2. A shareable card generated from the real verdict text.
3. A public archive of verdicts — **only** honest once there is more than one
   participant, and requires a moderation answer.

**Recommend 1 or 2. Do not build 3 yet.**

---

## H. PRODUCT — "why argue with an AI when Perplexity or Google is free?" [owner-relayed user comment]

Not a defect; the sharpest thing in this list. The Gauntlet's premise is that
arguing with an AI is worth twenty minutes, and a visitor can already do that
free, anywhere, without a clock.

**This cannot be fixed in CSS and should not be answered by engineering.** It is
the question of what the Gauntlet offers that a chat window does not — the clock,
the judgement, the adversarial framing, a record of whether you held up. Items
E, F and G are all downstream of the answer.

**Owner decision needed before further design work on this page.** Doing C–F
without it risks tightening a page whose proposition has not been settled.

---

## Ordering for whoever takes this

1. **A** — a headline nobody can read, and a same-day regression. Fix first.
2. **B check-side** — one clause, fleet-wide value, prevents recurrence.
3. **B page-side + D** — mechanical, low risk.
4. **H** — the owner's call; blocks the rest.
5. **C, E, F** — only after H.
6. **G** — after H, options 1 or 2.

## What this says about our verification

A, C, D and E were all present while the page passed every automated check it
has. The checks assert that controls exist, that they do what they promise, and
that nothing lies. **None of them asks whether a person can tell what to do.**
That gap is what `experience_loop/HANDOFF_2026-07-28_appeal_dimension.md` was
opened to close, and item A is the strongest argument yet for its first
recommendation: add a `contrast_ratio` check before anything else.
