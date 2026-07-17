# The concept register — where we came from, where we are, a deep dive on the bug-historian, and where we're going

*2026-07-17. The calm, read-aloud version. For the technical detail see
`PILOT_bug_historian_reviewer.md` (the reviewer's full design), `PLAN_concept_register.md`
(the living plan), and `006_VERIFICATION_stage2.md` (the verification method).
Turn-by-turn log: `RUNNING_NOTES_concept_register.md`.*

---

## Where we came from

Short version, for anyone joining mid-story: this platform's history lives in
about four thousand scattered documents, so we built a register that reads
all of them and pulls out every distinct thing anyone ever built, half-built,
or abandoned — around sixteen hundred entries. Then we checked every single
one against the real running code rather than trusting what the documents
claimed, and found about one in thirteen was wrong. We fixed those, with a
second independent check trying to disprove every fix before we trusted it.

The whole point of doing any of this was never the register itself. It was
to grow the fix-loop's review panel — the part of the system that looks at a
proposed bug fix before it becomes a real pull request. That panel had two
members. We wanted a third, and we wanted to pick it with evidence instead of
a guess.

## Where we are now

We picked it, we built it, and it's live.

The evidence came from counting which lessons this platform's history keeps
teaching over and over, to people who didn't know someone else had already
learned them. One pattern stood out sharply: content quietly vanishing during
a rebuild, with no error and no warning, discovered independently at least
seven times across completely different corners of the system — a tool
widget destroyed by a routine page rebuild, an early defence layer built
because generated pages kept coming out blank, a component regeneration that
silently broke every page depending on it, and — the freshest instance — an
article body that went empty because a template rendered a missing field as
nothing rather than failing loudly. Seven different bugs, seven different
root causes, one shared shape. That's exactly the kind of thing a reviewer
sitting on the panel should recognise on sight.

Since then we've also kept the register itself honest and current. The
fix-loop delivered its first real diagnosis this week — genuinely confirmed,
with evidence, not a benchmark — and in the same stretch of work a second,
unrelated bug turned up: a config setting that looks like it's taking effect
but silently isn't, on a noticeable slice of the platform's agents. Both are
now written up properly, checked against the real code rather than taken on
faith, sitting in the register alongside everything else.

## A deep dive on the bug-historian

Here's what we actually built and how it thinks.

**What it is.** The review panel used to be two members: one that checks
whether a proposed fix is well-made, and one that checks whether it's safe
for the rest of the platform. The bug-historian is a third member with a much
narrower job: it doesn't judge quality, and it doesn't judge safety. It asks
one question only — *has something like this happened before, and does this
plan account for it?* It sits in the review order between the other two,
reads the fix that's being proposed, and compares it against a short, curated
account of that seven-times-recurring pattern. If the proposed fix risks
repeating the shape, or only patches the one spot it happened to notice
without touching the underlying cause, it says so.

**What it can't do.** It cannot block a fix on its own. Only one of the
original two reviewers holds that power, and we kept it that way on purpose —
this new member's job is to raise a hand and say "wait, I recognise this,"
not to have a veto. That distinction turned out to matter more than expected:
while designing exactly how the review panel makes its decisions, we found
that the panel's actual rule is that *any* member's objection at the
strongest level gets treated as a block, regardless of which member raised
it. Knowing that, we deliberately built the historian so it can only ever
raise a concern, never issue that strongest-level objection — a genuine
choice, not an oversight.

**Getting it live.** Wiring it in meant editing the actual production
definition of how the review panel works and applying that change to the
live database — not a copy, not a test environment, the real thing the
fix-loop uses every time it proposes a fix. Before touching it, an automatic
safety check in our own tools stepped in and refused even to *look* at the
production database, because nobody had specifically said out loud "yes, that
exact database, that exact box." That's the system working as intended, not
a fault — so we stopped, explained exactly what we needed and why, and waited
for you to say the specific words back. Once you did, we applied it: the
prior version of the workflow was backed up automatically first, so the
change can be undone in one step if it ever needs to be, and we checked
afterwards, directly in the database, that everything was wired the way it
was meant to be — the new reviewer in the right place in the sequence, its
opinion correctly included wherever the other two reviewers' opinions already
were.

**Where it stands right now.** Live, correctly wired, and it has not yet
actually reviewed a real fix. That's not a gap — the fix-loop only just
confirmed its first real diagnosis this week, and sending that diagnosis on
to become an actual fix proposal is a decision that's still sitting with you,
not something that happens automatically. So the historian is standing at
its post; it just hasn't had a case walk through the door yet. We also
double-checked something important once that second, unrelated config bug
turned up this week: it happens to affect how a number of agents read their
settings, and we confirmed directly that the reviewer panel's own workflow is
not one of the ones affected. Good news, and worth confirming rather than
assuming.

## Where we're going

Two things are queued up, both waiting on the same kind of decision — yours,
not ours to make unilaterally.

The nearer one: whenever the confirmed diagnosis from this week gets sent on
to the fix-writing stage, that will be the bug-historian's actual first
outing. Worth watching for, if only to see whether a reviewer built from a
paper trail of past incidents turns out to be useful in practice the way the
evidence suggested it would be.

The further one: there's a second strong candidate for a fourth review-panel
member, built the same evidence-driven way — someone whose whole job is
checking "have we built this before," aimed at a different but equally
well-documented habit this platform's history keeps repeating: building the
same thing twice because nobody checked first. It's fully designed, sitting
ready, waiting only on whether a second new panel member is wanted yet or
whether one at a time is the better pace.

And underneath both of those, the quieter work continues: keeping the
register itself honest as the platform keeps moving, catching new bugs and
new subsystems the moment they surface rather than letting the map go stale
again the way it did the first time around.
