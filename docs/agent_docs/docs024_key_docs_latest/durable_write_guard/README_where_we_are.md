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
