# SUMMARY — where we are, 2026-08-09

*Concept register. Written to be read aloud. Previous milestone:
`SUMMARY_where_we_are_2026-08-04.md`; the series is the record, so this is a new
file rather than an edit of that one.*

## What we're trying to do

Keep one place that answers **"does this already exist, and is it real?"** for
every mechanism, contract, agent, tool and idea in the platform. Not
documentation — documentation says how a thing works. The register says *that* it
exists, who built it, what it is for, what will mislead you about it, and whether
it is genuinely live or merely written down.

Two audiences use it, and the second is why accuracy is not housekeeping: a
session about to build something, who needs to discover it was built in June; and
the council seats that review our changes, whose prompts are seeded from these
entries. A wrong entry is not misleading prose — it is false evidence inside a
machine review.

## Where we've come from

Built in July by sweeping ~4,100 files under `docs/`: 2,185 raw concepts extracted,
merged to 1,627, then every one checked against live code and database — which
corrected 124 of them and, on an adversarial second pass, overturned 106 *proposed*
corrections. Stage three turned its categories into the council seats that now
review platform changes.

Since 20 July it has had no dedicated thread. It grows because each session
registers what it builds in the same commit, watched by a coverage check. That is
the whole maintenance model, and five days ago we found out it was not working.

**2026-08-04** — brought it up to date. Two defects, both the same shape, *a check
whose result could not have come out otherwise*:

- **34 concepts had a register entry and no index row**, including the entire
  first half of the claims-verification layer. The index is what people search, so
  those read as *does not exist*. It survived ~20 careful re-measurements because
  each compared the row count to the **previous row count** — blind to a row nobody
  ever wrote.
- **The coverage check's accepted-backlog list ignored any line a session had
  annotated**, so 12 of 17 "new" items every run were already-settled decisions.

Both fixed, and the same day 1,339 superseded duplicate documents were deleted
(441 documents that existed as 1,973 numbered copies).

**2026-08-04, later** — you asked for a watcher in the framework rather than in a
session's local tooling. It exists: a daily CronJob that reads the register at a
pinned ref and writes its verdict to the notes table. Not a commit-time hook,
deliberately — a hook only fires for the session already editing the register,
which is the one most likely to have got it right.

## What we've done since

**The watcher has been firing unattended and it earned its place immediately.**
Five runs by 08-08: four scheduled, one manual. It caught two things.

First, **a headline mismatch on three consecutive days that nobody corrected.**
Second, and worse, **the missing-row defect came straight back**: `SCH-024`, the
reaper-accounting mechanism, was filed on 08-08 with an entry and no index row —
four days after 34 such rows were backfilled. That is the finding that changed the
picture. 08-04 was not a backlog that got cleared; the class recurs at roughly one
every few days.

**So today, at your direction, we retired the stored counts entirely** rather than
keep watching a number drift. The evidence for that call, measured rather than
argued:

- **Four different commands count the index, and all four answers are correct.**
  On 08-08: 1,792 by the documented row regex, 1,799 by a looser one quoted inside
  the file's own history, 1,792 unique entry ids, 1,800 raw headings. Twice in four
  days a careful session published the wrong one.
- **The per-category files had the same disease, unwatched.** All 109 carried a
  stated count; **32 were already wrong**, totalling 90 concepts of drift. One
  claimed five concepts and held none.

The number is now derived, never written down. The old chain of "previously N,
re-grepped after X" is preserved verbatim in a frozen log at the foot of the index,
because how it failed is worth more than the numbers were.

**The watcher's fourth check was inverted rather than deleted** — with no headline
to compare, it would have found nothing and reported nothing, which looks exactly
like passing. It now reports any stored count that has *come back*. That inversion
is proven by a mutation test: re-add a count and the check must name that file.

**Two things the tests caught before they shipped**, both worth more than the
feature: the frozen log quotes the old headlines verbatim, so a whole-file search
would have flagged a finding every run for ever — a watcher crying wolf about its
own archive; and an assertion I wrote assumed a broken register would look broken
from every angle, when in fact the headline was perfectly accurate while 34
concepts were missing. Agreement proved nothing. That is now what the test asserts.

**Also fixed, by the duplicate-id arm on live data:** two lanes claimed `LNK-031`
hours apart on 08-08. The second is renumbered `LNK-032` with its own row.

## Where we are now

The register holds ~1,800 concepts across 109 category files — **and that figure is
deliberately approximate here**, because the exact number is now derived by the
watcher and by the commands in the index header, not stored in prose. Entries and
index rows agree exactly, no id is used twice, and no file states a count except
one.

That exception is honest and owed: `rebuild-cascade.md` has been dirty in the shared
working tree since 04-08 with another session's edit, and retiring its one line
would have swept five days of their uncommitted work into a commit about something
else. The watcher reports it daily, which is correct — it really does carry a
stored count.

What is **not** true: nothing has re-verified whether the ~1,800 entries are
*accurate*. Verification ran once, in July. The register's honest claim is that it
is complete, consistent and now self-monitoring — not that it is current.

**Owed, and small:** the CronJob's script changed today, so its ConfigMap is stale
until `make deploy-concept-register-drift-check` is re-run. Until then the daily
run still does the old four checks, which is safe but blind to a returning count.

## Where we're going

**Staleness is still the open flank**, unchanged from 08-04 and now the biggest
thing between us and a register that can be trusted without reading the code.
Coverage answers "is it here?", drift answers "does it agree with itself?" — and
nothing answers "is it still true?" A register status is already a known landmine:
a snapshot that outlives its truth, quoted by council seats as ground truth. The
building blocks exist (`covers-through` stamps, the landmine-verifier, the
bugs-open staleness sweep). Pointing them at register entries is a design question
worth its own session.

**The missing-row half may yet need a gate, not a watcher.** Retiring the counts
removes one recurring defect at the source. The other — an entry landing without
its row — is still only *reported*, and today's evidence says it recurs every few
days. The candidate is a pre-commit rule that fires when a commit adds a `### ID`
heading without its index row. Worth doing if the daily row keeps naming new ones;
worth *not* doing yet, because we now have a mechanism that will tell us the rate
instead of us guessing it.

**And one honest risk to watch:** everything above assumes somebody reads the daily
row. The headline mismatch sat uncorrected for three days while the watcher
reported it every morning. A mechanism that reports into a table nobody opens has
the same failure mode as the convention it replaced — which is exactly why the
retirement (removing the artefact) is a better answer than the report (watching it),
and why the next intervention should also remove a thing rather than watch one.
