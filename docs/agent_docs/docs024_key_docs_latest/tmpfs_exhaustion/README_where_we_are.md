# README — where we are (tmpfs / scratch exhaustion)

The owner's running log. Append only, newest at the bottom. Plain prose.

---

## 2026-08-24 — the fix worked, and the problem moved house

You asked this lane to own the `/tmp` problem, not just describe it. So the first thing I did was
re-measure yesterday's fix rather than read the write-up again.

**`/tmp` is genuinely better.** It is sitting at 4.9 GB of 16, about a third full, which is what
yesterday's note predicted. Everything large in there is a day or two old and belongs to sessions
that started before the change and are still using their old settings. That half is working.

**But the problem did not stop. It moved.** Yesterday's fix pointed all the session scratch at the
hard disk instead of at memory. That was the right call — the disk is 60 times bigger and it isn't
RAM. What nobody checked afterwards was the disk.

The scratch folder on disk is now **147 GB**. It was created on 3 August, the day of that change,
so that is three weeks of accumulation with nothing ever clearing it out.

**130 GB of that 147 GB is one thing**: 308 abandoned throwaway copies of this repository. They are
the "does the committed code still compile" check that we ask every session to run. Each copy is
about 450 MB, and nothing deletes them.

**Why it's still happening, when yesterday's fix was supposed to stop it.** Every version of that
check contains a delete command, which is why it reads as if it tidies up after itself. It doesn't.
The delete runs at the *start*, clearing the folder the check is about to use — and each session
picks a different folder name. One session this morning left six copies behind: `headtree`,
`headtree2`, `headtree3`, `headfinal`, `ht5`, `ht6`. Nearly 3 GB in one morning, and each one's
delete command dutifully cleared only itself.

**The number that decides how urgent this is.** In the last seven days those copies grew by
**73 GB — about 10 GB a day**. There is **123 GB free** on the disk. So we are roughly **twelve
days from a completely full disk**, and that is a worse day than a full `/tmp`: a full `/tmp` gives
you an annoying error, a full disk stops git, stops builds, and stops every session on the machine.

I want to be straight that this is our own remedy having a side effect we didn't look for, not a
new problem. Moving something unbounded into a bigger container buys you time in proportion to how
much bigger the container is, and it costs you the warning light that was telling you the thing
existed at all. That is what happened here.

**There is also a plain reason yesterday's documentation fix didn't take.** The write-up said the
recipe appeared in nine documents, and eight of them were corrected. I counted this morning: it is
in **seventy-three**. And the one file that every session reads automatically, `CLAUDE.md`, doesn't
mention the fix or the trap at all. So the correction went to the people who already knew, and the
other sixty-five documents carry on telling sessions to do the old thing.

**What I've built.** A cleaner-upper, `scripts/scratch-janitor.sh`, that handles both places —
memory and disk. Two things about it that are deliberate and worth your knowing:

It does **not** clear the disk by age. A session's scratch folder holds the disposable copy of the
repository *and* the session's actual work — its notes, its analysis — side by side in the same
place. Deleting "anything old" would take both. So on disk it only removes two things it can prove
are regenerable: a bare copy of this repository (identified by what's inside it, not by its name,
so a folder name nobody has invented yet is still caught) and the Go compiler's own leftovers.

It also **does nothing at all unless you tell it to.** Running it prints what it would delete and
stops. And it has a self-test that deliberately creates each hazard and checks the safety catch
actually engages — which caught me out immediately, though the fault turned out to be in my test
rather than in the safety catches.

**What it says today: 261 folders, 108 GB.** I have not deleted any of it. That is other sessions'
data on shared ground, and last time this was cleaned it was done on your say-so for that reason —
so it's your call:

1. **Delete the 108 GB now** at the settings I've shipped (memory: untouched for a day; disk:
   untouched for two days). That takes the disk from 87% to about 76% and buys back the twelve
   days.
2. **Be more cautious** — only clear disk copies untouched for a week. Less back, but no chance at
   all of catching a slow session.
3. **Leave it** and just watch the number.

**The second question is whether it should run on a schedule.** As it stands it's a tool somebody
has to remember, and this problem has now recurred four times because nobody remembers. There's
already a crontab on this machine with two entries in it, so adding a third is the normal way here.
Without that, the pile starts rebuilding the moment attention moves elsewhere.

**One small correction to yesterday's note, in your favour.** It says nothing reaps idle scratch.
Something does — Ubuntu's own tidier runs daily. Its rule is "delete anything untouched for ten
days", and `/tmp` used to fill up in four, so it could never fire in time. Not "nothing is
watching" but "the watchman is set too slow", and those two point at different fixes. Shortening it
is a one-line system file and needs your root password; it would only help `/tmp` though, and 88%
of the volume is now on the disk side where that tidier doesn't look.

## 2026-08-24 (later) — I built something we already had, and that turned out to be the finding

An hour after writing the above I went to register the new cleaner-upper in the estate's catalogue
of reusable parts, and read the entries either side of where mine would go. **We already had one.**

`scripts/scratch-report.py`, built on 3 August — the same day as the change that started all this —
does what I spent the morning building. Same idea, same safety rules, three weeks older, and in one
respect better designed than mine. When I ran it, it said:

> **250 folders, 97.1 GB, safe to delete.**

At its own default settings. It has been able to say that for weeks. Nothing has ever run it.

So the honest correction is that **my headline was wrong in an interesting way**. I told you the
problem was that nothing cleans up. It isn't. We built the cleaner-upper three weeks ago and then
never scheduled it. The pile did not grow for want of a tool; it grew because the tool has no alarm
clock. That changes what I'm asking you for: not "shall I build a janitor" but **"shall I put one
line in the crontab"** — which is a much smaller thing, next to the two entries already there.

I've deleted mine. Two cleaner-uppers that drift apart is a problem this place files bugs about
most weeks, and the older one is better.

**Why I didn't find it first, since that's the part worth learning from.** Our own rules say to
check the catalogue before deciding something doesn't exist. I checked it before deciding to *tell
you* about it — which is a completely different and much later moment, by which point the thing was
written and tested. The symptom felt new, so I never asked whether the machinery was.

**One genuinely useful thing did come out of building the duplicate.** Comparing the two showed the
existing tool has a blind spot that has been hiding in plain sight. It is *supposed* to cover both
places — memory and disk — and its own documentation says so in as many words. It doesn't. It only
ever looked at the disk, because of how it goes hunting: it expects a particular folder structure
that the `/tmp` side has never had. Three weeks looking covered and covering half.

What makes that one worth telling you about is **how it hid**. There was no error, no wrong number,
no zero. The report just quietly printed one heading where there should have been two. A missing
line is the hardest kind of mistake to spot, because there's nothing there to read and disagree
with. I've fixed it — `/tmp` now shows up, with 2.7 GB of leftovers in it.

**So, unchanged, the two things I need from you:**

1. **The ~100 GB.** Delete it, or don't. The existing tool at its normal setting says 97 GB across
   250 folders; with the `/tmp` fix included and a slightly keener setting it's 106 GB across 272.
   Either way it leaves the 14.8 GB of actual session work completely alone — that's the part of
   the design I'd most want you to trust, and it's the same in both tools independently.
2. **The crontab line.** Without it we are back here in a fortnight, and this is now the fifth time.

## 2026-08-24 (end of day) — done, and the number that did *not* move is the interesting one

You chose the two-day setting and a nightly report that deletes nothing. Both are in place.

**The clear-out freed 98.7 GB.** The disk went from 87% full to **75%** — from 124 GB free to
**231 GB**. The scratch folder went from 147 GB to 38 GB. It took about seven minutes.

**Three things it deliberately did not touch, which is how I'd judge whether it was the right
tool:**

The **14.8 GB of actual session work** — notes, analysis, half-finished files belonging to other
sessions — is untouched, because the tool identifies what it deletes by looking inside it, not by
how old it is. That is the part of the design I'd most want kept if anyone ever changes it.

**`/tmp` didn't move at all.** Everything in there is about a day and a half old, and your setting
was two days. A tool that had helpfully cleared it anyway would have been the wrong tool.

**Every system folder survived** — the five named ones plus all eleven of the private ones. That
exclusion list has now earned its keep twice.

**And the number that didn't move: swap is still completely full.** That's worth a line, because it
settles something. Yesterday's clear-out freed 11 GB from `/tmp` and gave the machine back 6.9 GB
of swap, which made it look as though scratch files were the cause of the memory pressure. Today I
deleted **nine times as much** — and swap did not shift by a single megabyte. The difference is
that `/tmp` is memory and the disk is not.

So the memory pressure on this box is not scratch and never was, beyond that one episode. It is the
roughly **fifty Claude sessions holding about 25 GB between them**. That is not something this lane
can fix, and I'd rather say so plainly than leave a tidy story standing that today refuted.

**What's now running by itself:** a nightly report at 06:41 that lists what could be cleared and
deletes nothing, appending one short timestamped block to `/home/ant/scratch-report.log`. I tested
the exact command in a stripped-down environment like cron's *before* switching it on, because a
cron job that quietly dies looks identical to a quiet machine. Your two existing crontab entries are
untouched — I backed the file up first and checked all three afterwards.

**One thing I got wrong and had to undo within the hour**, since it's the sort of thing worth
recording. I'd found that the older tool's blind spot showed up as a *missing heading* in its
report, and I wrote that up everywhere as the thing to watch for. Then the clear-out ran, `/tmp`
legitimately dropped below the threshold, and the heading disappeared for a perfectly innocent
reason. The same silence now meant both "all clear" and "broken". So I fixed the tool instead of
the documentation: every location now always prints a line, even when it has nothing to say. The
lesson I'd already written down and then walked straight into is that **you should never make an
absence carry meaning** — if something matters, print it.

**Where this leaves us.** At the recent rate the pile rebuilds by about 10 GB a day, but that rate
should now fall, because the pointer to the right command finally sits in the file every session
reads rather than in eight documents only their own lanes read. The nightly log will tell us which,
and that's the measurement I'd want before touching anything else — including whether `/tmp` should
be a memory-backed folder on this box at all, which is still open and still needs your root
password.

## 2026-08-25 — you asked if `/tmp` was full again. It isn't. Something much bigger is.

**Short answer: no.** `/tmp` is at 27%, slightly emptier than yesterday, and the scratch folder went
*down* from 33 GB to 25 GB. Both halves of the fix are holding.

**But the disk went from 74% full to 85% overnight** — 97 GB gone in a day — while the thing I'd
been watching was shrinking. So I went looking, and found two things I had never measured, one of
which dwarfs everything this lane has ever looked at.

**Go's build cache is 117 GB and grew 50 GB in a day.** That's the compiler's own store of
previously-compiled code, and it's genuinely useful — it's what makes the second build fast. Go
does clean it up, but only removes things untouched for five days, and with fifty sessions
compiling we're adding fifty gigabytes a day. Same story as before: a cleaner that works and can't
keep up.

**Docker's build cache is 539 GB, of which 438 GB is dead weight.** There are also 1,034 stored
container images taking 104 GB, of which exactly one is in use. That is over half a terabyte, and it
is more than five times everything my lane has ever been looking at.

**Now the part I want to own, because it's a mistake I've now made three times in three days.**

Every disk figure this lane has produced — including yesterday's confident before-and-after table —
came from a command that **cannot see Docker's storage**, because that lives in a root-owned folder
my account can't read. And it doesn't fail loudly when it can't: it reports **"4.0K"**, which looks
exactly like a small empty folder. Add up everything my command could see and you get 226 GB;
the disk says 754 GB is used. **I have been reporting on a quarter of the disk and presenting it as
the disk.**

There's a one-line check that catches this — compare the total your tool found against what the
filesystem says is used, and if they disagree wildly, you can't see the disk. It's free, and I
should have run it on day one.

**And the nightly report I set up for you yesterday never ran.**

The laptop was asleep from 23:15 to 09:54, and I'd scheduled it for 06:41 — inside the sleep. The
old-style scheduler I used simply skips anything it slept through; it would have missed it every
single night, silently. I've replaced it with the kind that catches up when the machine wakes, the
same mechanism Ubuntu's own cleaner uses on this box — and I tested that it actually *runs*, rather
than just confirming it was scheduled.

The one thing that worked exactly as intended was the warning I'd written next to it: *a missing
entry in the log means the job didn't run, and must never be read as "nothing is wrong"*. That's
the only reason I opened the log instead of glancing at the disk and telling you it was fine.

**What I'd like your decision on, and it's a bigger number than yesterday's:**

The 438 GB of dead Docker build cache is reclaimable with one command. The trade is that the next
few builds of each service will be slower while it rebuilds what it actually needs. There's also
around 90 GB of stored images that nothing is using. I haven't touched either — it's a much larger
irreversible action than yesterday's, it affects everyone's build times, and it's outside what this
lane was originally asked to do.

I'd also gently flag that **this is now the fourth different thing filling this machine**, and the
first three each looked like the whole story at the time. The honest position is that the disk needs
someone watching *all* of it, not this lane watching the corner it started in.
