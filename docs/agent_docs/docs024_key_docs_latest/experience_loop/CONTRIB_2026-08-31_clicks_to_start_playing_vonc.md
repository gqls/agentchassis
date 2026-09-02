# CONTRIB — vonc Gauntlet: two clicks before a visitor can type anything

**From:** the vonc provocation-pipeline session
**Measured:** 2026-08-31 · **Filed:** 2026-09-02 (the session spanned the date change;
every figure below was taken on the 31st and has not been re-checked since)
**To:** the experience-loop lane — and, because of the supersession below, whoever
is actually holding the vonc Spark experience.
**Status:** measurement + a question. Nothing changed, nothing dispatched.

---

## Why this is filed here, and where it probably belongs

The owner asked me to "correspond with the experience loop and the offer and benefit
analysis threads to perhaps reduce the number of clicks that a user has to take to get
to start playing (only if that's the right thing to do)."

Routing note for whoever reads this: **this lane was SUPERSEDED on 2026-07-25.** The
`gauntlet_dead_cta` workstream owns the vonc Spark experience's build cycle end to end,
against a plan approved 2026-07-25 (corr `5316e79c`) that targets a real backend
(`tools-api` on the island), *not* run 8's CP2 plan. I am filing here because the owner
named this thread; if you are reading this and the build is still owned there, hand it on
rather than acting from this lane.

The same measurement has gone to the live
`offer analyser benefit analyser visual designer` session, which owns the visual-designer
and offer/benefit tracks.

## The measurement (first-hand, 2026-08-31, against the live site)

Path from the home page to the first point where a visitor can type:

| # | Where | Control | What it gets you |
|---|---|---|---|
| 1 | `vonc.com` | primary CTA **"File Your Position"** | → `/tools/gauntlet/index.html` |
| 2 | gauntlet page | **"Enter the Gauntlet"** (`data-gi-enter-btn`) | reveals the position box |
| 3 | position box | **"File Position"** (`data-gi-position-submit`) | opponent replies |
| 4 | defence box | **"Send Defence"** (`data-gi-defence-submit`) | verdict |

**Two clicks before any input is possible.** Step 2 is a splash screen whose entire
content is the heading "Enter the Gauntlet" and one sentence:

> One provocation. Twenty minutes. An AI opponent that argues the other side, and a
> judge that says who held up.

## The part that is a defect rather than a preference

Reducing clicks is a taste question and the owner explicitly made it conditional. This
is not:

**The home CTA says "File Your Position". The page it lands on does not let you file a
position.** It requires a further click on a differently-named button first.

That is a promise/delivery mismatch — the same class this lane was created to catch, and
which `misdirected_cta` structurally cannot see. `misdirected_cta` proves an anchor
reaches a real page; it has no way to know the page withholds the thing the anchor named.
This is a concrete, live instance of that gap on the pilot site, which may be worth more
to the lane as a test case than as a fix.

## What I would NOT do

Delete the splash outright. Twenty minutes is a large ask and setting that expectation is
doing real work — a visitor who starts without knowing may abandon mid-round, which is
worse than one extra click. The candidate worth costing is **landing on the position box
with the expectation set alongside it** rather than gating behind it.

## One false lead, pre-cleared

`href=""` appears three times in the home page markup. All three are inside
`data-runtime-fill="true"` shells (`provocation-card`, `lobby-grid`) that the JS fills
from `/data/provocations.json`. They are the deliberate empty-shell contract, **not dead
links** — three discovery checks have historically misfired on exactly this. Do not file
them.

## Context

vonc.com served the same provocation from 22 Aug to 31 Aug — nine days — because the
content shelf ran out and nothing noticed. I am fixing that in the provocation-pipeline
lane now, on the owner's instruction to make it daily and to stop requiring his sign-off.
If you canary anything on this site, pin the date you looked at: the daily content is
about to start moving again.
