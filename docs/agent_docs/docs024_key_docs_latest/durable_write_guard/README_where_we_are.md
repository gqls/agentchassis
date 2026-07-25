# Where we are — the durable-write guard (bug 021)

Plain-prose running log, newest at the bottom. Append only.

---

**2026-07-21 — opening this workstream, and a plot twist.**

Bug 021 is a follow-up to bug 012. Back on the 18th we caught a fix-agent
truncating a working interactive tool and saving the wreckage — 10,000 characters
of a working tool cut down to 1,253 characters of dead CSS — straight over the
real thing, and reporting success. We built a guard that refuses that kind of
mangled write. But the guard only sits on ONE door. A reviewer on that fix
correctly said: this is a shape the platform keeps repeating, and the same door
exists in a couple of other places — please put a guard on those too before the
next accident. Bug 021 is that "please do the others". The two doors it named
were the page's rendered HTML, and the page header/footer/head fields.

I went to put guards on those two doors and found something worth telling you:
**both doors are already bricked up, for different reasons.**

- The header/footer/head fields on pages: **nothing writes them at all.** They're
  leftover columns — empty on all 301 live pages. The real page chrome lives
  somewhere else and is assembled fresh each time. A guard there would be guarding
  a door nobody uses.
- The page's rendered HTML: this is a *photocopy*, not the original. The real,
  editable source is the component template (which the existing guard already
  protects and which we keep version history for). The rendered HTML is produced
  by stamping that template out with the page's data — so if it's ever damaged, we
  just re-stamp it. I checked this against the live database: for every
  interactive tool, the rendered copy is the same size as its template. A guard
  that compares "new vs old" here would either do nothing (because the writes are
  just re-stamps, not fresh AI output) or actively block legitimate fixes.

So the literal task — "put the same guard on those two things" — turns out to be
the wrong task. But the reviewer's *instinct* was dead right, because there IS an
unguarded door, and it's currently letting damage through onto live customer
sites right now.

**Here's the real hole.** When we GENERATE a brand-new tool (as opposed to editing
an existing one), the only check on the AI's output is "does it have the little
header comment we expect at the top?" That header sits at the very start. If the
AI's answer gets cut off at the *end* — mid-JavaScript — the top is still fine, so
the check passes, and we save the broken tool. There's a proper check for this
that we already built and shipped last week (it looks at whether every `<script>`
etc. is closed) — but we only run it when *reading* a tool back, not when first
*saving* one. So a broken tool can still be born; we just quietly refuse to
re-display it later.

This isn't hypothetical. Another thread (bug 046) fetched the live sites two days
ago and found **8 tools serving broken, non-working JavaScript on 6 real customer
domains** — arena on vonc.com, the grip-force calculator on robot-hands.com, and
others. Most of them were never healthy: they were born broken through exactly
this gap. That thread is handling the *cleanup* (restoring or regenerating the 8);
its own plan explicitly says the *prevention* — the guard — belongs to this bug
021. So there's a clean split, no toes trodden on: they mop up, we fix the tap.

**What I recommend.** Add the check we already have to the tool-creation step, so
a cut-off tool can't be saved in the first place. It reuses proven, already-live
code, it's low-risk, and it closes the exact gap that produced the live breakage.
Optionally, do the same (a lighter version) for the general section-creation step,
which has a similar but smaller weakness. And formally record that the two doors
the original bug named don't need guarding — with the evidence, so nobody re-opens
that question in three weeks.

I've written all this up. Before I write code, I want your call on scope — because
this changes *what* we build and *where* from what bug 021 originally asked, and
the bug itself says a human should decide the scope first. Putting the choice to
you next.

---

**2026-07-21 (later) — you chose "do both gates"; built and committed.**

You picked the fuller option: guard the tool-creation step AND the general
section-creation step. Both are done and committed (`ba702c8c6`), with tests. In
plain terms:

- When we generate a **new tool**, we now check that its HTML is actually
  complete — every `<script>`/`<style>` block properly closed, ending on a real
  closing tag — before we save it. If it's cut off, we refuse to save it, the job
  is flagged for a human to look at, and we record why. This is the exact check we
  already trusted elsewhere (when *reading* tools back); we've just moved it
  earlier, to *before* the broken tool can be born.
- When we generate a **new section**, we now also refuse one with an unterminated
  script/style block. (The old check only looked at whether the `<section>` wrapper
  closed, which a half-cut tool could satisfy while its JavaScript was still
  chopped off.)

One small thing worth telling you, because it's the kind of thing that bites us:
while committing, I found I'd been about to accidentally scoop up **another
session's unfinished work** — they had half-written changes sitting in the same
file I'd tweaked. I backed my (cosmetic) change out of that file so their work
stays theirs, and committed only my own four files. Nothing lost; their work
untouched.

**Where this leaves us:** the fix is written, tested and committed, but it's Go
code — it does nothing until we build a new image and roll it out. That's a
live-cluster change, so I've stopped here for your go-ahead rather than deploying
on my own. When you're ready I'll build it, roll it, and then *prove* it works the
honest way — by deliberately feeding the tool-creator a cut-off tool and watching
it get refused, not just by checking the code shipped. (The 8 tools already broken
on live sites are a separate, already-tracked cleanup job — bug 046 — not this
one; this fix stops *new* ones.)

---

**2026-07-21 (later still) — turns out it's ALREADY live.**

Small twist: I didn't need to build or deploy anything. While I was tidying up the
notes, you ran one of your "sweep" builds (v1.0.1146), and because my code was
already committed, it got picked up and shipped automatically. I checked the
actual running server and my new checks are in it — all the tell-tale strings are
there, and a known-good control string confirms I'm looking at the right binary.
So the guard is *live on the cluster right now*.

One honest caveat, and it's the same one I always insist on: seeing my code in the
running server proves it *shipped*, not that it *works*. This guard's whole job is
to catch a broken tool, and the only way to truly prove a catcher works is to
throw it something broken and watch it catch. I haven't done that live yet — so
right now the status is "shipped and in place, but the trap hasn't been sprung in
anger." To finish the job properly I'd run one deliberate test: hand the
tool-creator a tool that's cut off at the end and confirm it gets refused (and
flagged for a human), while a healthy tool still sails through. That's a real bit
of work on the live cluster (it involves kicking off a job and waiting on the
queue, ~half an hour), so I'll do it on your say-so rather than assume. Until then
I'm leaving bug 021 open and honest about it.

**Your call:** leave it live-but-unexercised for now — don't spend the half-hour
dispatch this session. So that's where it rests: the guard is live and protecting
new tool/section births, bug 021 stays open with an honest "not yet sprung in
anger" note, and the one deliberate live test is written down for whenever it's
worth doing (or for the first real cut-off generation to trip it for us).

---

**2026-07-23 — you said go, so I sprang the trap. It works.**

A fresh image had rolled (v1.0.1149); I first checked my guard survived the
rebuild — it did — then ran the real test on the live cluster. I set up a
throwaway one-step job that hands the tool-creator a tool I control, so I could
feed it a deliberately broken one without waiting on the AI.

- I fed it a **cut-off tool** (looks fine at the top, chopped off mid-script). The
  guard **refused it**: nothing was saved, the job ended in the error path, and it
  logged exactly why ("generated HTML is structurally incomplete"). That's the
  trap catching what it's meant to catch.
- Then I fed it a **healthy, complete tool**. The guard **let it straight
  through** — it only tripped later on a deliberately-fake site id, which proves it
  got *past* the guard. So we're not going to start refusing good tools.

Both tests created nothing on real sites, and I cleaned up every trace afterwards
(including the test error-log entries, so nothing sweeps them up as a real fault).

**Bottom line:** bug 021's prevention job is now *done and proven* — live, and
demonstrated to actually fire. The only reason the bug file stays open at all is
the unrelated second item in it, which belongs to a different workstream. The 8
already-broken tools remain the separate cleanup job (bug 046). Nothing further
from me on this unless you want it.

---

**2026-07-25 — you asked me to finish 021 and close it. Done. Three things
happened that are worth knowing about.**

**First, the review we thought was pending had never happened.** Back on the 24th
we sent the second half of this work to the reviewer council and recorded it as
"waiting for a verdict". It wasn't waiting. It had died six seconds after
starting, on a typo-grade problem: one of the four listed edits said it would
"create" a file, and the council only accepts the words "modify", "add",
"remove", or "config change". No reviewer ever saw it. The nasty part is that a
submission which dies this way looks *exactly* like one that's still queued — in
both cases there's simply nothing there yet — so it sat overnight looking
healthy. I've resent it with the one word corrected and the new evidence
attached. Two other sessions were bitten by the same thing yesterday and wrote it
up hours apart, which tells me the submission tool should check this before
spending anything; I've said so where the tool's owners will see it, rather than
going and changing their script myself.

**Second, the actual test — and this is the part I'd want you to see.** The
change adds a gate that asks, when a job claims to have fixed hardcoded colours
on a site, "did it actually?" The obvious way to write that gate is to re-run the
detector that found the problem. That would have been a quiet disaster, and I can
now show you why with real numbers rather than an argument. Across the fleet the
detector currently flags 32 components on 8 sites — but on **five of those eight
sites, not one of them is something our colour fixer was ever built to change**
(they're pale colours, short hex codes, colours written directly on the element).
So a gate built the obvious way would have refused every completion on those five
sites for ever, retried them twice each, and then filed them as failures — and
the sites it would have punished hardest are the ones where nothing was wrong.

So I tested it both ways on the live system. On robot-hands.com, where three of
three flagged components *are* in the fixer's range, the gate **refused** the
completion, sent the job back for another attempt, and wrote down exactly which
component it was unhappy about — and it named the right one, which I knew in
advance because I'd computed the answer first. On finetuning.uk, where eight
components are flagged and **none** are in the fixer's range, the gate **let it
through** and recorded why. That pair is the whole point: refusing when it should
and, just as importantly, not refusing when it shouldn't. I also found the gate
had already quietly passed a real job on vonc.com at 10:18 this morning, with no
prompting from me.

**Third, I got something wrong and want to flag it rather than bury it.** When my
first test message seemed to vanish, I decided a known glitch had eaten it,
changed the command, and re-sent. That was wrong. Nothing had been eaten — the
system's single message-reader was jammed for about fifteen minutes behind
*another* session's big review job, and when it freed up both of my messages ran,
twenty-four seconds apart. My "fix" fixed nothing; it just happened to coincide
with the jam clearing. I'd already written it into our runbook as a standing
trap, complete with a cost figure, before I disproved it. It's now corrected
there and logged in the fleet-wide list of wrong calls — where it turns out to be
the *second* entry of that exact shape filed today, by two sessions who couldn't
warn each other.

**Where that leaves it.** Bug 021 is closed and moved to the closed folder. Both
halves are live in the current build and both have been demonstrated firing, not
just shipped. One piece of unfinished business was split out as its own small
bug (077): the detector and the fixer disagree about what counts as a problem,
which leaves eight items sitting in the backlog labelled as if a fixer kept
failing on them when in truth there was never anything for it to do. It costs
nothing and breaks nothing — it just makes the backlog lie a bit — and how to
resolve it is a judgement call for whoever owns that check, so I've written down
the three options rather than picking one. The council verdict, whenever it
arrives, is advisory and doesn't change any of the above.
