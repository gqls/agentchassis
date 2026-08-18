# DECISION INPUT 2026-08-18 — 185 unreadable-text findings that will never clear themselves, and the reason we can now safely act on them

**For the owner.** Raised by the contrast front out of `bugs_open/296`. Everything here was
measured on the live system between **15:40 and 18:12 UTC on 2026-08-18**; where something
is unmeasured it says so.

**The one-line version:** the backlog we parked in August has started clearing itself, the
tool we were afraid to use now works, and what is left — 185 findings — has been re-checked
by the machine and is genuinely still broken. **We need a decision on those 185.**

*(The filename says 186 because that was the count when I started writing; one more was
withdrawn while I measured. The live figure is **185**, and it will keep drifting down
slowly as pages get repaired — every number here carries the time it was taken.)*

---

## 1. What a "contrast finding" actually is

When we render one of our sites in a real browser, we measure the text against whatever is
behind it. If the difference is too small to read comfortably, we write a ticket. The number
on each ticket is a ratio: **4.5:1 is the accessibility standard**, and **1.0:1 means the
text is exactly the same colour as its background — completely invisible.**

On 10–11 August that check wrote **226 tickets** across 16 sites.

## 2. What we decided in August, and why that decision has now expired

We did not try to fix those 226 straight away. We **parked** them — put them in a holding
status where nothing would pick them up.

**The reason was specific.** The tool that fixes them, `css-patch-agent`, had a known defect
(`bugs_open/213`): it reported work as finished when it had not done it. Letting it loose on
226 tickets would have produced 226 tickets marked "fixed" that weren't — turning an honest
backlog into a dishonest one.

**Both halves of that reasoning have since gone away, and this is the news in this document:**

- `bugs_open/213` — the false-completion defect — **is fixed and closed.**
- `css-patch-agent` **is now working.** Since yesterday afternoon it has processed **58
  contrast tickets, completed all 58, and failed none.** It writes real CSS, e.g.
  `p.p { color: #4a4a40; }`.
- **I checked its work at the actual web pages, not at its own status.** On
  `noted.co.uk/index.html` both pairings it patched are now gone; `noted.co.uk/contact.html`
  measures **completely clean, zero failures.**

So the thing we were protecting the backlog from is no longer the thing it was.

## 3. Meanwhile, the backlog has begun clearing itself

There is a second machine involved: the weekly render audit. It re-visits each site, measures
the pages again, and **withdraws any ticket whose problem has genuinely gone.** It only ever
withdraws on a fresh measurement, so a withdrawal is evidence, not an assumption.

It reached these sites for the first time overnight. Of the original 226:

| | count |
|---|---|
| withdrawn — re-measured and genuinely fixed | **40** |
| still open, still failing | **185** |
| cancelled earlier by hand | 1 |

**A worked example so you can see it is real.** `dartsonline.com/about.html` had six
unreadable items yesterday afternoon, including five headings at 1.06:1 — effectively
invisible. This morning the audit re-measured it and withdrew fourteen of that site's
seventeen tickets. I then re-rendered the page myself: **it is down to one item, and the
heading problem is gone.** The page was genuinely repaired and the withdrawal was correct.

> **I got this wrong yesterday and it is worth saying so.** I predicted this audit would
> clear "approximately none" of the backlog. It cleared 40. I had the evidence that would
> have corrected me — I could see hundreds of page re-renders running — but I read that
> traffic only as something getting in the audit's way, never as the thing repairing the
> pages. The correction is recorded in `bugs_open/296` and `WRONG_CALLS.md`.

## 4. What is left, and why it will not clear itself

**185 findings, across 14 sites, in 65 distinct colour combinations.**

The important property: **every single one has now been re-measured by the machine in this
pass and deliberately not withdrawn.** They are not "not yet looked at". They were looked at
again, this week, and they still fail.

**And re-rendering will not save them.** Every one of these sites has been re-rendered
heavily since 15 August — `vonc` 44 times, `robot-hands` 93, `ai-agent-orchestration` 150,
`webdesign` 956. A re-render rebuilds the page from the same templates and the same colours,
so it reproduces the same unreadable pairing. **Repetition cannot fix these. Only a change
to the colours themselves will.**

### How bad they are

| | how bad | findings | sites | colour combinations |
|---|---|---|---|---|
| **A** | **invisible** (1.0–1.2:1) | **60** | 7 | 18 |
| **B** | severely unreadable (1.2–3.0:1) | 40 | 6 | 19 |
| **C** | fails the standard but readable (3.0–4.0:1) | 59 | 8 | 18 |
| **D** | marginal (4.0–4.5:1) | 26 | 5 | 9 |

**Class A is the one to look at.** Sixty pieces of text on seven live sites are, to a
visitor, simply not there.

### The three shapes behind most of it

- **`robot-hands.com` — white text on a white button.** 1.00:1, nine findings. The button's
  background colour never arrived, so the label is invisible. This is the same family as
  `bugs_open/113`.
- **`vonc.com` — semi-transparent white on the brand purple.** 3.19–3.37:1, nineteen
  findings. Deliberate design that lands just under the standard.
- **`idea.uk` — muted grey on cream.** 3.11–3.35:1, twenty-one findings. A "quiet text"
  colour that is a little too quiet.

### Where they are

| site | findings | colour combinations | worst |
|---|---|---|---|
| vonc.com | 37 | 9 | 1.63:1 |
| robot-hands.com | 33 | 7 | **1.00:1** |
| idea.uk | 23 | 3 | 3.11:1 |
| mortgagecalculator.co.uk | 22 | 6 | 1.92:1 |
| ai-agent-orchestration.com | 17 | 6 | **1.00:1** |
| finetuning.uk | 11 | 6 | **1.00:1** |
| gamesdesign.co.uk | 9 | 7 | **1.00:1** |
| leopardessconsulting.co.uk | 8 | 6 | **1.00:1** |
| webdesign.co.uk | 7 | 3 | 4.31:1 |
| lendzy.co.uk / vetcomparison.uk | 5 each | 1 / 3 | 2.90 / 3.77:1 |
| dartsonline.com / loanandmortgagecalculator.co.uk | 3 each | 2 / 3 | 1.06 / **1.00:1** |
| relojistas.com | 2 | 2 | 2.68:1 |

## 5. THE DECISION

**The question: do we turn the fixer loose on these 185, and if so on which ones?**

The safety argument has changed. Releasing them is no longer an act of faith, because there
are now **two independent machines**: `css-patch-agent` proposes and applies a fix, and the
render audit — which does not trust it — re-measures the page and only withdraws the ticket
if the problem has actually gone. **A fix that does not work leaves its ticket open.** That
is the grading we were waiting for in August, and it exists now.

| | option | cost | risk |
|---|---|---|---|
| **1** | Release all 185 to the fixer | one run, ~185 fixes | class A may get a wrong-but-passing fix (see below) |
| **2** | **Release C and D (85), hold A and B** | ~85 fixes | low — these are pure colour-value problems |
| **3** | Fix at source, per colour family | engineering, 65 families | slowest; needs design decisions |
| **4** | Accept and close them | nothing | 60 pieces of invisible text stay invisible |

**My recommendation is option 2, then option 3 for what remains — but this is your call, and
here is the reasoning you should push back on.**

Classes C and D are straightforward: the colour is close and needs nudging, which is exactly
what the fixer does well. **Class A is different and that is why I would hold it.** When text
is invisible because its *background* failed to appear, darkening the *text* makes the
ticket pass while leaving the real fault — a missing background — in place. We would get a
green tick and a button that still looks wrong. Those sixty want a person to look at them,
or a source fix per family (option 3).

**One caveat I want on the record:** I have verified the fixer at two pages. That is real
evidence and it is a small sample. Option 2 is partly chosen *because* it is the version
that limits what a wrong answer would cost.

## 6. What I am NOT proposing

- **Not** releasing class A in bulk without a decision from you.
- **Not** building a second withdrawal mechanism — one exists, is shared, and works.
- **Not** treating August's ink fix as evidence about these. It repaired a different family.

## 7. Two blind spots I found while measuring — recorded, not fixed

- **Sites in the `pool` and `system` states are never audited at all.** The audit only looks
  at `active` and `deployed` sites. Today that is 23 deployed and 2 active, versus **17
  `pool` and 1 `system`** — and those pool sites have only 2 published pages between them,
  so the live blind spot is 2 pages. It is unguarded, though: if a pool site ever publishes
  properly, it becomes invisible to this check and **nothing would tell us.**
  *(A related note, because it is a natural worry: there is a `building` state in the code
  that would also be skipped — but nothing ever sets it. No live workflow writes it, so no
  site is ever in it. It is a latent gap, not a live one.)*
- **`[UNMEASURED]` Whether the audit covers every page of a large site.** The runs I could
  inspect measured 5–24 pages, while `webdesign.co.uk` has 108 published pages. The record
  I needed was already deleted by routine cleanup, so **I could not check this either way
  and am not asserting it.** If audits are capped below a site's page count, problems on the
  unmeasured pages are never found and never withdrawn. Worth one check on the next
  `webdesign.co.uk` audit.

## 8. Separately — we have made the loop faster

The audit used to re-visit each site **once a week**, so a repair could wait up to seven days
to be confirmed. On your instruction that has been shortened to **three days** (migration
`469`). Three rather than two because the audit competes with build work for its slot, and I
measured it being turned away on 9 of 14 attempts during a busy period — at two days it would
start missing turns silently. Three roughly doubles the speed and keeps headroom.

## Where the working is

`bugs_open/296` §8 — every query re-runnable. Defect families: `features_open/026`
(brand-colour-as-text) and `bugs_open/113` (missing palette slot).
