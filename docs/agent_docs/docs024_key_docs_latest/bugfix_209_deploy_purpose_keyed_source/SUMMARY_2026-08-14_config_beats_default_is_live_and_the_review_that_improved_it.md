# SUMMARY — 2026-08-14. The config-beats-default fix is live, and the review round is the most useful thing in this document

Milestone read-out, written to be read aloud. Previous in the series:
`SUMMARY_2026-08-11b_census_closed_and_the_signature_collapse.md`.

**Status of the verdict, stated up front because the rest of this leans on it:**
round 1 came back **REVISE** and has been read in full and acted on. **Round 2 was
resubmitted and its verdict has NOT landed** — the round is still running (last seen at
the guardian seat). Nothing below claims approval, and no commit in this lane carries a
`Council-Reviewed:` trailer.

> **⚠ CORRECTED, same day, ~20 minutes after this file was written — ROUND 2 CAME BACK
> AND IT IS ALSO `REVISE`.** 12 seats, 5 abstained, no truncation, **gated by
> `prior_art_librarian` on a HIGH objection**, with `bug_historian` repeating its round-1
> point. The paragraph above is left as written because the series is the record; this
> note is what changed. **The gating objection is that I answered a gating objection
> badly, and it is right:**
>
> - **HIGH — I argued down the architecture gate with a citation the seats cannot read.**
>   Council seats have no access to `CLAUDE.md`, so "the owner ruled to ship this" and
>   "the owner retired the staged-rollout requirement" are, from their side, unverifiable
>   appeals to authority used to dismiss a gate. There is a landmine registered against
>   exactly this seat for exactly this. The rulings are real and they do apply — but that
>   is not the point. **The point is that a reviewer who cannot check a claim is right to
>   refuse it**, and the correct move was to route the objection (which I did, to
>   `RFC_028`) and then say nothing further, rather than also litigating it with evidence
>   the reviewer has no way to see. Round 3 drops the argument and quotes any ruling it
>   must reference verbatim.
> - **MEDIUM — my headline measurement had no query attached.** The "27 rounds, 8
>   `needs_rfc`, 1 veto" figure is doing load-bearing work in the argument, and I put the
>   numbers in the evidence list without the SQL that produced them. This is the seat
>   whose entire job is verifying such claims against `diagnosis_artifacts`, and I gave
>   it nothing to run. Unverified, it is, in its words, "folklore dressed as a
>   measurement" — and it is a fair description, even though the number is correct and
>   the query is three lines and sitting in `RFC_028`.
> - **LOW — my "no duplicate classifier exists" proof is a content grep**, and content
>   search over behaviour patterns is documented as unreliable here. A miss does not
>   prove absence, which is a rule this estate has written down more than once and I
>   leaned on anyway.
> - **`bug_historian`, repeated: the remedy I built is inert.** `--report` exists but no
>   CronJob drives it, so for an unbounded interval a mistyped setting on a defaulted
>   field is refused with no fleet-visible signal at all. My reason for not shipping the
>   overlay was sound (the image must exist first, or the cluster reports the failure as
>   "still running"), but the seat's point stands regardless: **a deferred remedy is not
>   a closed objection.** It asks specifically whether the job has run even once and
>   produced a real row — which it has not.
>
> **None of the four is a defence of the code**, and none disputes that the fix works.
> Three are about how I argued, and one is about a control that is built but switched
> off. That is a more uncomfortable set of findings than round 1's and a more useful one.

---

## What we're trying to do

Make a step's written-down configuration actually take effect.

Every action in this platform declares a small specification: which settings it
accepts, and what each one falls back to when nobody says otherwise. The intent is
ordinary — the fallback is there for the common case, and a step that wants something
different says so in its own configuration.

That is not what happened. The resolver filled in every fallback **first**, and then
went looking for what the configuration had asked for — and each of those later
lookups skipped any setting that already had a value. A fallback *is* a value. So the
fallback always won, and a large class of configuration was decoration: written,
readable, reviewed, and inert.

The goal of this piece of work was to invert that one rule — an explicit
configuration value should beat a fallback — and to do it in a way where the risk
could be *shown* rather than asserted.

## Where we've come from

This began as a wrong-logo problem. A step said "this is a logo", the resolver
delivered "hero" (the fallback), and because the filename is derived from that value,
the logo's bytes were published under the hero's name. Eleven sites were affected.
Those were repaired, and the immediate cause was fixed for the two definitions that
carried it.

The more important finding was that the logo case was one instance of a general defect,
and nobody knew how large the class was. A census settled it: **99 live configuration
entries were inert.** Seventy-eight of them happened to repeat their fallback word for
word, so nothing looked wrong. Twenty-one disagreed with the fallback that silently
beat them.

A third face turned up during that census and is the one worth remembering: four
different auditing agents each set their own distinctive label on the findings they
file, and all four labels were dead — every finding any of them had ever written was
stamped with the same fallback name. The discriminator was the thing that was broken,
which means the records cannot be attributed after the fact even in principle. That was
repaired separately by another lane, by a stronger route than the one originally
proposed: the field is now required, with no fallback at all, so a fifth author
repeating the mistake fails loudly instead of landing a fifth silent row.

The owner then ruled that the general fix should ship. That ruling is what this
milestone delivers.

## What we've done

**The fix is written, reviewed once, revised, and live.** It runs on the current
chassis build, on both machines.

The part I would most want to defend is the *shape* of the safety argument. Because the
old resolver could only ever set one of these settings by a single specific route,
anything the new rule can touch is something that previously did nothing at all. That
is a property of the code, not an estimate — so there is no path by which this disturbs
behaviour that currently works. The counting then follows: of the 99 inert entries, 78
are word-for-word identical to their fallback, and the remaining 21 belong to actions
that were already reading their configuration directly through a private route. Net
change to live behaviour today: none. The value is that the trap is shut for whoever
writes the next one.

Two deliberate limits, both with evidence behind them. A value containing a dot is
still ignored when it fails to resolve, because a dotted value is a *pointer to data
elsewhere* — and this platform has already paid for treating one as a literal, with
over 150 broken image links named after the pointer itself. And I dropped an extension
I had planned, after listing all 48 live cases and finding every one of them is plainly
a word an author typed rather than a pointer.

**The tool that reports this class of defect was rewritten in the same commit**, because
two thirds of what it used to call "broken" is now working configuration, and a checker
describing last week's code is worse than no checker. After the rewrite it reported
zero problems. I did not trust a zero from a checker I had just edited, so I took the
live configuration, deliberately broke one entry, and fed it back: it caught it. The
zero is a real zero.

**And then the review round, which is the most useful thing that happened.** Eleven
reviewers; six had nothing to say; four of the objections were real defects rather than
misunderstandings, and all four are now fixed:

- My safety argument was an argument, not a test. True of today's code, and nothing
  stopped someone adding another rule next month that quietly broke it. It is now a
  test that fails if anyone does.
- A step using both the old and the new spelling of a setting got the **old** one.
  That is backwards for a migration, and it is fixed.
- The change to the old-spelling path needed tests for the case it wasn't about.
- When the new code *rejects* a badly-typed setting it only said so in the logs, which
  on this system vanish within minutes. There is now a mode that writes a permanent
  record instead.

There is an embarrassment attached to the first of those and it belongs in the record.
The new test says "the automatic search did not override the fallback" — a sentence
that also passes if the search could never have found anything, in which case the test
guards nothing while looking as though it guards something. I checked, it does find
values, so the test is real. But I only checked because we have a standing habit of
asking that question, and the check is now written into the test itself.

## Where we are now

**The code is live and proven to be running.** The binary carries the exact commit, and
— the half that makes it meaningful — a *different, later* commit is correctly absent
from it, so the check can tell the difference rather than agreeing with everything.
Both machines. The offline census re-confirms zero dead entries.

**The behaviour has not been witnessed, and that is recorded as owed rather than
passed.** I went to the logs to watch the new rule fire and found nothing. The
explanation is not that it isn't working: the machine retains about **ninety seconds**
of logs, on a machine that had been running five hours. I know that rather than guess
it, because an older log line that has always been there is missing from the same
window too. So the logs cannot answer the question either way.

The irony is worth stating plainly: that is precisely what one reviewer objected to,
and I ran into it myself an hour after answering them, while trying to check my own
work. It is now the empirical justification for the permanent-record mode — which is
built, and deliberately **not** switched on, because it needs a container image first
and enabling it before that exists produces a failure this cluster reports as "still
running" rather than "broken".

**One structural question is now in front of the owner rather than settled by me.** The
reviewer responsible for architecture flagged that this shared resolver is a piece of
machinery nobody formally owns, and it could not tell whether that had been flagged
before. I measured it: **27 review rounds have touched this resolver, 8 of them raised
the same signal, and one was vetoed outright.** So roughly one round in three notices
it, and each time the change ships anyway — because each individual change is
well-argued and blocking it would cost something real. That is written up as a proposal
with three questions, not a recommendation I've acted on.

## Where we're going

Four things, in order:

1. ~~**Read the round-2 verdict** and act on it.~~ **DONE — it is REVISE (see the
   correction at the top). Round 3 is the next action, and it is mostly not code:** drop
   the owner-ruling argument entirely and let `RFC_028` carry the architecture question
   to a human; attach the SQL behind the 27-round measurement; replace the content grep
   with a declarations-based check; and decide the one thing that genuinely needs the
   owner — whether to build and push the CronJob image so the permanent-record mode
   stops being inert. That last one is an outward-facing registry push, so it is a
   decision to take rather than assume.
2. **Witness the new rule firing.** Not from a log tail on an old machine; from a live
   stream, with an older log line in the same filter as the liveness control. This is
   the last outstanding claim in the whole piece of work.
3. **Decide whether the permanent-record mode gets switched on**, which means building
   its image first and then wiring it to the daily schedule alongside the two sibling
   checks that already run there.
4. **The three architecture questions**: does the resolver's rule chain get an owner;
   should the one load-bearing distinction in it be expressed once in code instead of
   twice in prose; and is there a point at which adding another rule to it needs more
   than a review round.

Deliberately left open: the other half of the original defect — 96 configuration
entries that point at data which is sometimes absent, and fall back silently when it
is. That stays open by design, because whether a pointer resolves is only knowable
while the system is running, so it is a detection problem rather than a precedence one.
