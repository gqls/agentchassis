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
