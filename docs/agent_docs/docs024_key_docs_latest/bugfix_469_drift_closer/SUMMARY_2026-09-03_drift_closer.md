# Where we've got to — the detector that watched a page lose a section and then went quiet

Written to be read aloud. `bugs_open/469`, lane `bugfix_469_drift_closer`, 2026-09-03.

---

## What we're trying to do

Stop the platform destroying deliberate human work on live pages without anyone being told.

A page's list of sections is stored in three places at once, and one of them is only a
photocopy: every time a page is rebuilt, the real plan is copied down over it. So if
somebody corrects a page by editing the photocopy — which is the obvious place, and the
place the page itself reads back to you — the next rebuild quietly undoes them. Nothing
errors. The page just comes back wrong.

We have a detector that spots the two copies disagreeing. What we did not have was anything
that closed its warnings, so they piled up, went stale, and — worst of all — blocked the
detector from ever warning about the same page again.

## Where we've come from

This is the third time round on the same mechanism, and that is the point.

In July, a migration swapped a component on a robot-hands page by writing only the
photocopy. The rebuild resurrected the old layout. A second migration fixed it by hand, and
its author wrote the trap up carefully so nobody would repeat it.

Somebody repeated it. On 24 July a spec-sheet block was deliberately added to the gripper
catalogue page — an owner-backed call, chosen over a product grid whose empty price and
rating fields would have invited the writer to invent numbers — and it was written to the
photocopy. Four days later the detector raised a flag. Then nothing happened for
thirty-seven days, the page rebuilt, and the block was gone.

Another session found this on 3 September while clearing the backlog, filed it as bug 469,
and left it explicitly unowned. This lane picked it up the same evening.

## What we've done

**Checked that the problem is real, and that it is not currently on fire.** Nothing on the
estate is drifting right now — 398 pages compared against their plans, plus 34 more on the
older-style plan, all in agreement. The earlier lane's clean-up held. But the spec sheet is
genuinely gone from the gripper catalogue page, confirmed in all three stores and in the
live page itself.

**Found two things nobody had connected.** That page is marked "archived" in the database —
and it still serves a normal page to visitors, which I checked by fetching the real URL
with controls rather than assuming. And there is already an untriaged warning about that,
eight days old, from a different detector; all nine of its kind across the estate are
sitting untouched. So the page has two separate open questions, and answering one of them
implicitly answers the other.

**Built the fix, and the interesting part is what it refuses to do.** The obvious closer —
"the two copies agree again, so the warning must be resolved" — would have been *worse than
having no closer at all*. They agree again precisely *because* the rebuild destroyed
somebody's work. That closer would have converted a silence into a certificate,
automatically, across every site.

So instead the closer asks what agreement *cost*. If the page kept what the human wanted,
the warning closes quietly. If something was destroyed, the warning can only close by
filing a permanent record of exactly what went — and that record is written *first*, in the
same database transaction, so the two cannot come apart. We also made the original warning
say, up front, which sections the next rebuild is going to destroy, rather than printing two
lists and leaving a person to spot the difference.

**Proved it by breaking it.** Fifteen deliberate sabotages of our own code, each one
expected to make a specific test fail. Fourteen did on the first try. One did not — and
that one is the most useful thing we learned: the test we thought was guarding a particular
line was actually being rescued by a *different* guard further down. The safety was real;
our proof of it was not. We sharpened the test rather than shrugging.

**Nearly built something we already had.** I was about to propose a whole new archive so a
destroyed section list could be recovered. One query showed we already archive the deleted
blocks — their names, positions and full contents — and have done since early August. What
we *don't* keep is the ordering. That is a much smaller gap, and the fix now points at what
exists instead of duplicating it. Logged as a near miss.

## Where we are now

The fix is written, tested, committed and reviewed-pending. It does nothing yet: this is Go
code, and Go code is inert until the next fleet rebuild. It will start working on its own
after that, without anyone doing anything.

The repair for the gripper catalogue page is also written in full, with all its safety
checks, and rehearsed against the live database in a transaction we rolled back — including
deliberately feeding it wrong data to confirm it refuses. **It is held, not applied**, and
that is the honest answer rather than a stalling one: the page is already recorded as
"built from the current plan", so correcting the plan would leave the correction sitting in
the database and never reaching a visitor. Getting past that means withdrawing that record,
and whether we are allowed to is exactly the question another lane put to you today.

## Where we're going

Three things need you, and they are cheapest answered together because they are about one
page:

1. May we withdraw a page's "already built" record so a composition repair actually
   renders? (This is RFC_064's second question.)
2. Does the rebuild machinery even reach an archived page? I don't know, and I would rather
   say so than assume.
3. Should the gripper catalogue page be serving at all? If the answer is "no, retire it
   properly", questions 1 and 2 stop mattering for this page entirely.

Beyond that, two things are noted and not attempted. Nine other detectors raise warnings
that nothing ever closes — the mechanism built here is reusable for them, but each one needs
its own judgement about what closing safely means, and doing nine of them shallowly is how
you get nine naive closers. And the record this fix files goes into a queue the estate
demonstrably does not drain. That is worth saying plainly: what we have guaranteed is that a
destroyed page can never again be quietly marked "resolved" — not that somebody reads the
record. The reading is a different problem, and it already has a bug number.
