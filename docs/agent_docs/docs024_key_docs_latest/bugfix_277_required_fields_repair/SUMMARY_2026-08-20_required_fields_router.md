# SUMMARY — 2026-08-20 — the blocker changed shape: outside pages were thought to have no repair route at all, and it turns out there is one, at thirty-six successes out of thirty-seven

> ## ⚠⚠ CORRECTED THE SAME DAY, ~10:15Z — THE HEADLINE CLAIM IS RETRACTED
>
> **Left standing rather than rewritten, deliberately: this is what we believed at the milestone,
> and the series is the record.** Read it as that, not as current state.
>
> **What is wrong:** the route is real (36 of 37 on outside pages) but it **cannot reach these
> pages**. Their blemish is in the *finished* HTML, and every repair route we have rewrites the
> *source data* — which on these pages holds only a fingerprint and a note of where the page came
> from, no words at all. So there was nothing for the repair to work on. Measured: re-rendering one
> of these pages from its source data produces a page with **zero words**, and no error.
>
> **What that means for the read-out:** "the blocker changed shape" is **retracted**. It is where it
> was — these pages have no repair route, and giving them one means building something that edits
> finished HTML, which nothing here does. For seven instances of the mildest blemish (backticks
> around code words on developer pages), that is the owner's judgement call, not an obvious yes.
>
> **The one thing that genuinely did change**, and it survives: we had been blaming *ownership* for
> three days. It was never ownership. On this component, "we do not own it" and "its words cannot be
> regenerated" happen to be the same 100 pages — so the ownership guard took the blame for a
> problem it was not causing, and every fix anyone proposed, including both of mine, was about
> getting around it. Getting around it would have repaired nothing.
>
> Full account: `bugs_open/277` §5, and the correction box in `HANDOFF_2026-08-20_continue_here.md`.

## What we are trying to do

When our system inspects a website it has built, it files a note for anything it finds wrong — a
missing field, a placeholder left in, formatting characters showing through as literal asterisks.
Those notes are supposed to reach something that can fix them. This workstream exists because, for a
whole family of them, nothing could: the note was filed, correctly labelled, and then sat there for
ever. The aim is not a tidier queue. It is that a page which is wrong gets **repaired**, and that we
can point at one and say so.

## Where we have come from

We built the missing piece — a router that reads each finding and sends it to the right repairer —
and it went live, drained its backlog, and passed its council review. Alongside it we restored the
mechanism that decides when a repair is safe to attempt, and added a three-day timer so that
anything nothing can handle is escalated to a person rather than aging quietly.

That worked, and then it exposed the real problem underneath. A large group of findings sit on pages
we do not own — pages built for, or handed over to, someone else, which our system is deliberately
forbidden from rewriting. Our general-purpose repairer refuses those, correctly. So the notes were
now well-labelled and still unrepairable, which is a better failure but the same one.

Two nights ago we thought we had found the answer, and had it backwards: we proposed moving seven of
those stuck jobs onto a repairer that was working beautifully next door. That would have produced
seven fresh failures and fixed nothing, because the working repairer works on *our* pages and would
have refused these too. Catching that turned into the lane's most useful lesson: when you propose
moving work, you have almost certainly checked why it is stuck where it is, and not whether the new
place would accept it.

## What we have done

Since then, three things.

**We fixed a mechanism that was quietly handing people wrong directions.** When a job escalates to a
person, the system attaches a note saying which team should pick it up. All three entries in that
list were wrong — two named bug files by the folder they used to sit in, and both had since been
fixed and filed away; the third said "nobody owns this" about a problem that had had an owner for
three days. It matters more than ordinary stale paperwork because the note is stamped onto the job
**once**, at the moment it escalates, and never revisited — so correcting it afterwards does not help
anyone who already received it. And it had already happened once, on the only escalation this
mechanism has ever produced. Fixed, with the underlying rule written into the list itself so it
cannot rot the same way again: name the bug, not the folder it currently sits in.

**We refuted our own conclusion, in the useful direction.** "An outside page with a real, fixable
blemish has no repair route at all" is wrong. There *is* a route, it has been in the system's own
instructions the whole time, and on outside pages it has succeeded **36 times out of 37**. The
general-purpose repairer, measured the same way, has eight successes on our own pages and one failure
on the single outside page it ever tried. These were never two blocked routes; one is refused by
design and the other is how this estate has been editing outside pages all along.

We also checked the thing that should have killed that idea. There is a recorded warning that this
very route, used on a certain kind of page, silently blanks all the text while every quality check
still passes — and six of our seven pages are that kind of page. It does not apply: the warning needs
two conditions together and only one is present, and the blemish we care about is not in the risky
part of those pages anyway.

**And we corrected ourselves twice more, both worth recording.** The fix we applied at eight in the
morning contained three figures of our own that were false by quarter past eight the next morning —
one of them a prediction whose mechanism had already been made impossible by somebody else's work two
days earlier. And our first test of the "blanks all the text" warning came back clean but was
incapable of coming back dirty: it looked for leftover placeholders in the finished page, when the
whole failure being described is that placeholders are replaced with nothing.

## Where we are now

The main bug's blocker has **changed shape**, which is the real movement: it was "there is no route
for these pages", and it is now "there is a route, and one question remains about whether our
detector can hand work to it". That question is about code, not design — it can be answered by
reading, not by building.

Nothing has actually been repaired yet, and we are not claiming otherwise. The larger hole — 27 of
the 30 parked jobs, whose pages have no content at all — is completely untouched by any of this.

The seven jobs we tracked for two days are gone, and how they went is itself a finding. Their group's
success rate crossed a threshold overnight, they were released, tried, refused, and closed as "will
not fix" — all seven inside three and a half minutes. So the escalation we expected tomorrow will not
happen. A job that can only be refused reaches its dead end faster than it reaches a person, and the
quick path is the silent one. We have told the team who owns that seam, with the measurements.

The second bug in this lane is waiting out a week of quiet running, around the 25th. One companion
bug is closed, one new one was filed by a neighbouring team taking the half of this problem that
belongs to them, and one config-review gap has now been corroborated by three separate lanes.

## Where we are going

Answer the code question: can our detector file work for the route that works? That is the next real
piece, and it is what stands between this bug and the one repaired page it has always needed. After
that, the honest remaining work is the big one we have been circling — the 27 jobs whose pages have
no content, which needs a different repairer and is not solved by anything found this week.

Two smaller debts stay open and are named in the handoff beside this file: telling a review seat that
one of its standing checks is inverted, and a change to the dispatch rules that genuinely alters
behaviour and so goes past the owner rather than in on our own judgement. Full priority list:
`HANDOFF_2026-08-20_continue_here.md`.
