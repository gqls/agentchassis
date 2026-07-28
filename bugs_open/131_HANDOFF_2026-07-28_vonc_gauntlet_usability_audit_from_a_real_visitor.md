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

**THIS IS A SECOND INSTANCE OF AN ALREADY-FILED CLASS — see
`bugs_open/112_…_shipped_css_diverges_from_the_pinned_palette_and_makes_text_invisible`**
(filed 2026-07-27 by `brochure_component_library`, after the owner looked at
fundamentallyai.com and reported unreadable text). Same mechanism, different site:

| site | element | ratio |
|---|---|---|
| fundamentallyai.com | eyebrow: `--color-primary` on `--color-background` | **1.11:1** |
| fundamentallyai.com | card title: `--color-heading` on `--color-card-bg` | **1.21:1** |
| **vonc.com** | **hero accent: `--color-primary` on a `--color-primary` background** | **1.00:1** |

**Two sites, two workstreams, one mechanism — a token used as foreground where the
same or near-same token is the background — and BOTH were found by the owner
looking at the page, not by any check.** That is the strongest possible argument
for the `contrast_ratio` check proposed in
`experience_loop/HANDOFF_2026-07-28_appeal_dimension.md`: it is not one site's
cosmetic problem, it is a recurring structural fault with no detector.

⚠ **112 is also an AMBIGUOUS NUMBER** — `bugs_closed/112` is an unrelated Gemini
API-key case. Refer to the CSS one by slug. (Its file also shows as deleted in the
working tree by another session as of 2026-07-28; recover with
`git show HEAD:bugs_open/112_…` if it has gone.)

**The general defect remains open elsewhere:** any component whose accent token is
also its background token has this fault latent. Worth a fleet grep for
`color: var(--color-primary)` inside a section whose background is the same var.

---

## B. Content bleeds off-screen on mobile — and `no_horizontal_overflow` CANNOT detect it [MEDIUM, structural] — PAGE-SIDE FIXED · CHECK-SIDE FIXED IN CODE, INERT

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
   > **CORRECTED 2026-07-28 (~15:45): the premise was wrong — the table ALREADY
   > sits in `div.pc-table-wrapper` with computed `overflow-x: auto`** (ancestor
   > chain measured live in Chromium). The content is reachable by scrolling,
   > which is exactly the fix this item prescribed. The "9 elements" this file
   > reports were the table's own internals (`table/thead/tr/th/td`) — a raw
   > crossing-rect scan cannot tell scrollable-within-a-wrapper from cut, which
   > is ALSO why the check clause below needed a scrollable-ancestor escape.
   > Nothing to fix here; at most a design nicety (a visible scroll affordance).
2. **The check** — **DONE by another session, 2026-07-28, commit `5042d5ecb`**
   ("no_horizontal_overflow now sees clipped overflow"). Filed here at 12:5x,
   fixed by 15:39: the multi-session file-passing working as designed.

   **Their implementation is better than the clause proposed above**, and the
   difference is worth reading before anyone "improves" it: it excludes
   `position: fixed|absolute` (off-canvas drawers are a deliberate pattern) and
   anything inside a horizontally scrollable ancestor — because a scroll
   container is the standard fix for a wide table, and such a table must then
   PASS this check rather than be reported forever. It also attributes the
   offender to the deepest/widest element rather than the ancestor that merely
   inherited the width. Code at
   `internal/adapters/browserrunner/run_checks_action.go:652-700`.

   > **STATUS: FIXED IN CODE, NOT LIVE.** Per `/bugs_closed/README.md` the bar is
   > fixed AND live. `browser-runner-adapter` is its OWN service with its OWN
   > image — a chassis roll does nothing for it. Fix committed **15:39**; the
   > running pod is `v1.0.1189`, started **14:26**, i.e. 73 minutes EARLIER. This
   > item stays OPEN until that adapter is rebuilt and rolled, then re-verify
   > against this bug's own failing cases (`/` and `/about.html` at 390px).

   > ⚠ **LANDMINE for that verification: `strings` DOES NOT EXIST in the
   > browser-runner container.** CLAUDE.md's verify-against-the-pod recipe
   > (`strings /app/<binary> | grep -c`) returns 0 for EVERYTHING there — it
   > works on the chassis and silently fails on this adapter. Caught only by a
   > positive control (`no_horizontal_overflow` itself came back 0, which is
   > impossible). Use `grep -c '<marker>' /app/browser-runner-adapter` directly,
   > and always pair it with a marker you know is present.
   > **DONE 2026-07-28 (~15:40), commit `5042d5ecb`, council corr `845893c9`
   > pending; INERT until the browser-runner-adapter image rolls.** The clause
   > needed three filters the raw spec lacked, each field-proven: in-flow only
   > (off-canvas fixed/absolute UI is a pattern), visible only, and **no flag
   > when a horizontally scrollable ancestor exists** — without that last one
   > the clause false-positives on every properly-scrollable table, including
   > this file's own item 1 (see correction above), and a false acceptance
   > failure becomes an `improve_tool` fixer aimed at a correct page (126).
   > **Verified with the exact shipped JS on the live pages**: gauntlet + about
   > clean; **the homepage FLAGS — a real residual cut this file does not list:
   > `div.brief-explanation__stat`, right edge 434px on a 390px viewport (44px
   > cut), in-flow, no scrollable ancestor, cause `flex-wrap:nowrap` on the
   > stats row.** The morning's page fix measured element WIDTH (14→0
   > over-wide); a 106px chip POSITIONED past the edge is invisible to that
   > measure. Page-side fix of the stats row still owed.
   > **Check-side council APPROVED 14:42 UTC (corr `845893c9`); the verdict
   > post-dates commit `5042d5ecb`, so this note is the join, not a trailer.**
   > **Stats-row cut FIXED & VERIFIED LIVE ~15:50 BST:** `flex-wrap: wrap`
   > added to `.brief-explanation__stats` in the component's base rule —
   > `content_components.html_template` AND all five `page_components.
   > rendered_html` in one transaction (§10/§11 discipline; the CSS block has
   > no `{{.vars}}` so one string replacement, pre-counted to exactly 1 hit
   > per artefact, applies to both). vonc homepage redeployed (assemble-only
   > rerender, corr `a8528143`, COMPLETED, `__step_error` NULL) and verified
   > in the renderer: shipped JS `{over:0}` (was `clipped:true cutCount:3`),
   > computed `flexWrap === "wrap"`, 0 raw placeholders.
   > **⚠ The component is shared: robot-hands.com, idea.uk (/index + /tools)
   > and oufe.com carry the SAME fix in their DB rows but their live pages are
   > UNCHANGED until each next re-renders** — deliberately no unsolicited
   > deploy of other workstreams' sites. Their cut (if any — narrower type
   > scales may fit) heals on their next natural rerender.

---

## C. "Enter the Gauntlet" reveals nothing [MEDIUM] — FIXED & VERIFIED LIVE 2026-07-28 ~20:35

> **DONE — the owner chose "the button reveals the round" over
> "position-as-entry" (the two candidates below; ruling taken ~19:50).** The
> page now opens SEALED: header, an entry block (button at 512px on a phone —
> was 1,913px — plus the "How it works" link and its own status line), and the
> sidebar (clock + rules). The provocation panel is hidden until `/round`
> returns 200; the `gi-sealed` class is removed ONLY in that success handler or
> when resuming a genuinely live stored round — the reveal is itself an
> API-bound state change, per the standing rail. The feed pre-render was
> REMOVED (dated comment at the code records the reversal so nobody restores
> it); the two "press Enter the Gauntlet" recovery messages now point at
> filing a position (the button hides after reveal, and position-filing
> auto-starts a fresh round; a 404 also clears the stale round id so that
> auto-start genuinely fires).
> **Verified twice**: 22-check local harness against the LIVE engine (sealed
> load · injected-503 stays sealed with the error at the entry · real reveal ·
> real AI counter · mid-round reload resumes revealed · fresh context sealed ·
> overflow clean both states), then 16 checks on the DEPLOYED page, desktop +
> mobile, real browser CORS, one real round, 0 raw placeholders.
> **The new 131-B check earned its keep during this build**: it caught the
> entry button and then the rules link at 398px on a 358px row (content-box
> padding on top of max-width:100%) — two would-be cut elements stopped
> before shipping. Sources: `p4_sources/*2026-07-28d_sealed_reveal*`,
> harness `drive_reveal_131c.py`, delivery corr `824c7f1c`.
> Residual for E/F: the revealed panel's provocation framing and the page's
> visual ranking are unchanged — C's fix gives E its structure (input directly
> beneath the provocation, nothing between), E's affordance work remains.

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

## E. The provocation does not read as the thing you must answer [MEDIUM, design] — FIXED & VERIFIED LIVE 2026-07-28 ~21:00

> **DONE — the remaining half (container + affordance) shipped on top of the
> flow reorder and the C reveal.** The provocation now renders inside one
> distinct card (`.gi-provocation-card`): translucent surface, 4px
> `--color-stage` accent edge (the token already measured 3.31:1 on this
> background; the edge itself is decorative, text colours untouched), with the
> existing owner-voiced `challenge_intro` copy relocated INTO the card's foot,
> restyled as the directive attached to the question — no new prose written.
> Combined with C (revealed by the press) and the reorder (input directly
> beneath), all three clauses of this item's own fix direction now hold.
> Verified in the harness (accent computed `rgb(245,158,11)`, card holds
> eyebrow+title+text+intro) and on the deployed page. Sources:
> `p4_sources/*2026-07-28e_question_card_emphasis*`.

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

### PARTLY FIXED — the mechanism was distance, and it was measurable

The type size was already fixed; the remaining fault was **the box you answer in
was 2,066px below the provocation on mobile — 2.4 screens.** Between them sat
intro copy, a status line, the Enter button, **three objective bullets and a
progress bar**. Objectives and progress REPORT; they are not things a visitor
does, and they were physically separating the question from its answer.

**Reordered so the panel reads: provocation → answer it → how you are doing.**
Pure DOM reorder inside `.gi-challenge-body`; the JS selects by `data-` attribute
so nothing rewired (selector counts identical, div balance 21/21).

| | mobile | desktop |
|---|---|---|
| gap before | 2,066px | 1,118px |
| gap after | **1,624px** | **749px** |
| objectives | between them | after the input |

Verified live: 0 over-wide elements, no page errors, and a real round still
completes (opponent replied, objectives wired, 33% progress).

### STILL OPEN, and it is a DESIGN DECISION not a bug

The largest remaining separator is **`.gi-cta-row` at 317px** — the "Enter the
Gauntlet" button, which is **item C's own subject**. Closing the rest of the gap
means resolving C, and C is a choice this file already framed:

> either the button moves to the top and the provocation is revealed BY pressing
> it, or the button stops being the entry point and filing a position is the only
> entry. **Do not do both** — two entry points is why this is confusing.

The JS already auto-starts a round when a position is filed with no round, so the
input is *already* a working entry point and the button is the redundant one.
Moving it below the steps would close ~317px and resolve C — but it changes which
control is primary, which is the owner's call, not a mechanical fix.
**Deliberately not bundled into this pass.**

---

## F. The page is busy; the job is not obvious [MEDIUM, design] — FIXED & VERIFIED LIVE 2026-07-28 ~21:00

Owner: *"the site is very busy with lots of sections and bits and pieces
everywhere — which is not a bad thing in itself — but it needs to be organised
and make it much clearer what you're there to do, not necessarily by changing the
text but by improving the design — making everything just a bit tighter."*

This is the parent of C, D and E. The page currently presents: hero, rules card,
provocation panel, position step, opponent panel, defence step, verdict panel,
timer, objectives, progress bar — with no visual ranking between "read this",
"do this now" and "this will happen later".

**Not actionable as a code change until someone decides the hierarchy.** See §H.

> **DONE — the hierarchy question resolved itself once C/E landed, and the
> ranking is now STATE-DRIVEN.** Sealed: one door (C). Revealed: the question
> card (E), then the steps carrying explicit rank — `is-current` (full
> emphasis, inset stage marker), `is-done` (receded), `is-future` (muted,
> opacity 0.55) — set ONLY by `applyStepEmphasis()`, which is called from the
> same handlers that advance the round on real API responses, re-derived on
> restore. **Muted is a ranking, never a gate**: every control stays enabled
> (asserted in the harness — the defence button already explains itself when
> pressed out of order, the anti-dead-control rule intact). Witnessed through
> a full live round: position current → filed → defence current → verdict
> current with the rest receded; reload mid-round re-derives correctly.
> "Read this / do this now / this will happen later" is now painted, not
> implied.

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

---

## B check-side — NO LONGER INERT, 2026-07-28 16:56Z. Still not *proven*, and the difference matters

**The blocker this item recorded is cleared.** `browser-runner-adapter` now runs
**v1.0.1190**, pod `browser-runner-adapter-cb96646c7-bcztg`, started `16:56:18Z`,
rolled via a new single-service makefile target (`35c8277a8`).

**Why a rebuild was needed and a redeploy would have been a no-op.** `v1.0.1188` and
`v1.0.1189` are the **same image id** (`bb9cb4a8b649`), built `13:43:31Z` — **56 minutes
before** the fix commit `5042d5ecb` (`14:39:22Z`). A retag is not a rebuild. Rolling
1189 again would have restarted the pod, looked exactly like a successful deploy, and
shipped the identical binary.

**Pod-grep, discriminating** (using this file's own landmine recipe — `strings` does not
exist in this container):

```
positive control  no_horizontal_overflow            1
fix marker        "while the content stays cut off" 1     <- was 0 on v1.0.1189
negative control  zzz_not_a_real_marker_zzz         0
```

The negative control is there because a `grep -c` that finds something proves the file
is greppable, not that the binary is the one you think. All three had to move together.

### What is NOT yet proven, and why re-running the named cases will not prove it

This item's own instruction was *"re-verify against this bug's own failing cases (`/` and
`/about.html` at 390px)"*. **Following that literally would now produce a green result
that means nothing**, and the reason is recorded earlier in this same file:

- `/about.html` was **never** a real failure — the premise was corrected at ~15:45: the
  table already sits in `div.pc-table-wrapper` with computed `overflow-x: auto`.
- `/`'s residual cut (`div.brief-explanation__stat`) was **fixed page-side at ~15:50**
  via `flex-wrap: wrap`, and verified in the renderer.

So **both named pages are clean, and would be clean with or without the new clause.**
A pass on them is a happy-path check: it proves the adapter is deployed and reachable,
which the pod-grep already proved. It says nothing about whether the clause fires.

> **What would actually prove it: a page that IS clipped.** Per
> [[verify-the-failing-branch]] — a green happy path proves deployment, not correctness;
> induce the fault. Until the clause is seen returning `clipped:true` with an attributed
> culprit from the *deployed* adapter, the check-side is **deployed, not demonstrated.**

### The good news: no manual dispatch is needed to get there

The clause is on an **actively exercised path**, so this will happen on its own. The
check runs inside tool acceptance as `{"id": "no_overflow", "tier": 4, "type":
"no_horizontal_overflow", "profiles": ["mobile"]}` — **14 orchestrations reference it,
most recently `2026-07-28 15:17:51Z`** (i.e. shortly before this roll). The next
acceptance run at the `mobile` profile exercises the new code with nothing fired by hand.

**What to watch for, and where:** a `CheckResult` carrying `pass:false` **plus** a
populated `culprit` / `component` / selector — the attribution fields are the tell. The
old clause could only fail on `scrollWidth - clientWidth > 2`; the new one also fails on
in-flow visible elements laid out past the right edge with no scrollable ancestor. A
failure with an attributed culprit on a page whose `scrollWidth` is clean **is** the new
clause, and nothing else produces that shape.

**One caution for whoever confirms it.** A false positive here is not cosmetic: it
becomes an `improve_tool` fixer aimed at a correct page (`bugs_open/126`). If the first
flag looks wrong, check for a horizontally scrollable ancestor before treating it as a
page defect — that escape is exactly what the three filters exist for, and it is what
kept this file's own item 1 from being reported forever.

**Item B check-side therefore stays OPEN** — but it is now open on *evidence*, not on a
deploy. `/bugs_closed/`'s bar is fixed AND live; this is live and unwitnessed.

> **UPDATE 2026-07-28 18:23Z — the tag moved, the fix survived.** A fleet deploy took
> `browser-runner-adapter` from the `v1.0.1190` above to **`v1.0.1192`** (pod
> `browser-runner-adapter-8f74cbd95-nj866`, started `18:23:07Z`). Re-grepped rather than
> assumed, because this bug's whole check-side history is a lesson in tags that lie:
> positive control **1**, fix marker **1**, negative control **0**. The clause is still
> in the running binary. **Do not go looking for `v1.0.1190`** — it is gone, and the
> version number in the section above is now history, not state.
>
> **Still unwitnessed.** `no_horizontal_overflow` has run **0** times since the 16:56Z
> roll (checked 18:2xZ). Everything in "what to watch for" above stands unchanged.
