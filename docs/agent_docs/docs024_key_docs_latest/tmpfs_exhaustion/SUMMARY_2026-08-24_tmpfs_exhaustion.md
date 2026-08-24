# SUMMARY 2026-08-24 — scratch exhaustion: the fix worked, moved house, and the reaper was already built

## What we're trying to do

Stop this machine running out of space every few days. It has done it four times before today, and
each time the diagnosis was correct and the recurrence happened anyway. So the real goal is not a
diagnosis — it is to make the fix *reach* the sessions that cause the problem, and to bound what
gets left behind when it does not.

## Where we've come from

The trigger is one command. Sessions are told, correctly, to check that the shared branch still
compiles after they commit, by extracting the committed code into a throwaway folder and building
it. Each extract is about 450 MB.

On 3 August that scratch was moved off `/tmp` — which is *memory*, not disk — onto the hard disk.
On 23 August the command itself was turned into a script that cleans up after itself, the Go
compiler was pointed at disk too, eight documents were corrected, and 11 GB was cleared from `/tmp`.
Clearing it returned 6.9 GB of swap, which looked like proof that scratch files were the cause of
the machine's memory pressure.

The next day's measurement of `/tmp` showed it refilling at only 0.5 GB a day against a prior 3.
On that evidence the remaining work — something to sweep up automatically — was set aside as low
priority.

## What we've done

Re-measured, rather than re-read, and the conclusion did not survive.

`/tmp` was genuinely fine. **The problem had moved to the disk it was redirected onto**, where
nobody was looking: 147 GB, of which 130 GB was 308 abandoned copies of the repository, growing 73
GB in seven days. That was about twelve days from a full root filesystem — sooner than the problem
it replaced, and a worse failure, because a full disk stops git, builds and every session on the
box.

The mechanism is a detail everyone had read past. Every copy of the command *contains* a delete,
which is why none of them read as leaky — but it runs at the *start*, clearing the folder the run is
about to use. Each run picks a new name, so it clears nothing. One session left six copies in a
single morning.

Two other things were simply wrong in the record and are now corrected: the command is in **73**
documents, not nine, and 66 of those never delete anything; and `/tmp` *is* already swept
automatically, by the operating system, at a ten-day threshold — against a folder that used to fill
in four. Not "nothing is watching" but "the watchman is set too slow", which points at a different
and much cheaper fix.

Then the most useful mistake of the day. I built a sweeper — and when I went to register it,
discovered **we already had one**, built on 3 August, better designed than mine, listed in the
catalogue the whole time. It reported 97 GB ready to clear at its own default setting. It had never
been run. So the finding was never "we need a sweeper"; it was "the sweeper we have has no alarm
clock". I deleted my duplicate and folded its one real gap into the original — which exposed a live
defect: that tool's own documentation claims it covers both memory and disk, and it only ever
covered disk. Three weeks inert while reading as covered.

With the owner's decision: **98.7 GB cleared**, and a nightly report scheduled that lists what could
go and deletes nothing.

## Where we are now

The disk is at 75%, with 231 GB free, up from 87% and 124 GB. The scratch folder is 38 GB, down from
147. The 14.8 GB of irreplaceable session work was not touched, because the tool identifies what it
deletes by looking inside it rather than by age. All system folders survived.

Swap is still completely full — and that is the useful negative. Nine times yesterday's volume was
deleted today and swap did not move a megabyte, because the disk is not memory. **The memory
pressure on this box is the fifty-odd concurrent sessions holding 25 GB, not scratch.** Yesterday's
tidy single-cause story is retracted rather than quietly dropped.

The pointer to the right command now sits in the file every session loads, which is the half that
was missing: the 23 August fix was correct, and unadopted — of the fifty extracts created after it
shipped, none used it.

## Where we're going

The nightly log is the next measurement, and it decides everything else. If the rate falls, the
recipe really was the mechanism and the pointer reaching `CLAUDE.md` was the fix. If it does not,
the sweep gets promoted from reporting to deleting, which is one word in one crontab line.

Two smaller things stay open, both cheap: a commit-time rule so document number 74 cannot spell the
command out again, and the question of whether `/tmp` should be memory-backed on this box at all —
a one-line system change needing root, and the only option that removes the failure class outright
rather than bounding it.

The thing I would most want carried forward is not any of the numbers. It is that **a bigger
container is not a bound**: moving an unbounded producer somewhere roomier buys time in proportion
to the size ratio, changes nothing else, and costs you the warning light that was telling you the
producer existed. After such a move, the only measurement that means anything is taken at the
destination — a healthy reading where the symptom used to be is what success and failure both look
like.
