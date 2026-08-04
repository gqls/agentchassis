# SUMMARY — 2026-08-04 — one chrome link policy (bugs_open/191)

## What we're trying to do

Stop the site framework shipping navigation buttons that point at pages which do
not exist. Specifically: make the question *"which page is a piece of site
furniture allowed to link to?"* have **one** answer in the codebase instead of
two, so the header's menu and the header's button beside it cannot disagree.

## Where we've come from

A site's header is built in one pass: the navigation links and the call-to-action
button come out of the same code, reading the same list of pages. But they were
checked against two different rules. The menu got the careful rule — only pages
that have actually been published. The button got a looser one that had no
publication check at all.

The visible result, found by the mortgagecalculator lane on 3 August and
confirmed on the wire on 4 August: that site's menu had correctly shrunk to its
single published page, while the button next to it pointed at a stamp-duty page
that has never been published and returns 404. Because the header is copied onto
every page, that is a dead button site-wide — and because the header is written
once and then left alone, it does not heal.

This is the second time this exact *shape* has been fixed in this one file. Four
days earlier, the sibling question — *"which template serves this header slot?"*
— turned out to be answered three different ways, and was consolidated into one
named decision with a test that fails if a fourth answer appears. That fix
explicitly did not touch link targets. This is link targets.

## What we've done

Confirmed the bug was still real, in the code, in the database and on the wire,
and that no other session was working it.

Had **fable** draft the plan against the actual code, then ran its own
blast-radius commands rather than trusting its numbers — which was the right
instinct, because one of them nearly produced a false fact (below).

Built `ChromeLinkPolicy`: one named decision, used by both the menu filter and
the button. The important part is *what* got extracted. The careful rule already
existed — but the two escape hatches that make it safe (don't filter on a brand
new site; don't filter if the database lookup failed) were written *inside* the
menu code, so no other caller could reach them. That is why the button's author
reached for the wrong helper: there was nothing else to reach for. So the thing
lifted out is not the rule, it is the whole decision.

Every guard is proved by deliberately breaking the code and watching the test go
red, in a throwaway copy of the repository so the shared working tree was never
touched. A passing test proves nothing unless you have seen it fail.

Put it through the review council. **It took four rounds, and this is the part
worth telling**, because three of those rounds changed the code:

- The council caught that the fix was **inert for every header already built**.
  Correcting the rule only helps sites that happen to get rebuilt. I had written
  the repair for existing sites as a manual step on my own to-do list, which is
  not a mechanism — it is me remembering.
- It then demanded evidence that repairing a stored header actually reaches an
  already-published page. It does — every page assembly re-reads the header fresh
  from the database — but nothing I had submitted showed that, and I had to go and
  read the deciding code rather than argue from a function name.
- Two reviewers, independently, asked whether the *previous* fix in this same file
  had already built the machinery I was adding. **It had, and I had not looked.**
  My version invented a second mechanism beside the supported one and, worse,
  bypassed the guard that protects human-locked content — a header a person had
  deliberately locked would have been overwritten. No test of mine would have
  caught that. It is now built on the existing machinery instead.

Approved on round four with three advisory notes, all acted on.

## Where we are now

The code is committed in four narrow commits, the mechanism is registered so other
workstreams can find it, the trap is written into the fleet-wide landmines file,
and three of my own mistakes are in the wrong-calls log.

The most useful of those three: to justify a design decision I needed to know how
many sites would take the "brand new site" escape hatch. The first number was
**19 of 38 — half the fleet**, which would have sunk the argument. It was the
wrong question: 18 of those sites have no pages at all, so the rule never applies
to them. The real answer is **one site**. The design was fine, but the evidence I
was about to write down for it pointed the other way, and it was dated, marked and
compliant with every rule we have for recording measurements.

**The bug is not closed, and that is deliberate.** Go changes do nothing until a
new image is built and rolled out, which happens on the owner's schedule, not
mine. The 404 on mortgagecalculator.co.uk is still live as of this writing. The
house rule is that a bug moves to `bugs_closed/` only when the fix is **live**,
not when it is merged — a fix that is committed but dormant is still a
reproducible defect. So it stays open with a short, ordered list of what proves it.

## Where we're going

After the next chassis roll, in this order: grep the running pods for the new
symbols (a wire check proves an effect, but not *which* binary produced it), then
re-run the header builder on mortgagecalculator, then curl the button. Only then
does the file move.

One thing is deliberately **not** fixed here and is somebody's next job: getting a
repaired header out to pages that are already published requires those pages to be
re-assembled, and nothing schedules that. It is an existing open bug
(`bugs_open/117`) owned by another lane. This change makes the header repair
itself automatic, which is the half that was missing; it does not take on the
other lane's half, and it says so rather than quietly implying the live 404
disappears on merge.

The council also made a broader point worth carrying forward: this is now the
**second** time this class of defect — one shared decision, answered twice by
hand, reconciled only after it breaks something visible — has been fixed with a
one-off named type and a bespoke test. Two of those is a pattern, and the seat
suggested it deserves an architecture round of its own rather than a third
instance. That is recorded against the new entry, unresolved, as a question for
whoever picks it up.
