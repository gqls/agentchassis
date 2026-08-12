# SUMMARY — 2026-08-12: the offer track closes its own loop, and B4 arrives needing a decision

## What we're trying to do

Every site we build is supposed to have an answer to a simple question: what is this site
*for*, who is it for, and how does it make money? We call that the site's premise. The offer
track exists to make the estate answer that question about itself — to notice when a site has
no recorded premise, to get one written, and eventually to judge whether the offer a site
actually presents is any good.

## Where we've come from

Two weeks ago none of this existed and the question was asked only by humans, occasionally.
We built it in pieces: give the strategic review the site's own premise to judge against
(B1); make refreshing a premise safe on a live site, so we could do it without triggering a
rebuild (B2); and two checks that examine live sites and file findings when the premise is
missing or when the site's shape does not match how it says it makes money (B3).

Along the way we learned some uncomfortable things about our own machinery — most of them
about the difference between a system reporting success and a system having done anything.
Those lessons are the durable part; the checks themselves were the easy half.

## What we've done

**The loop closed, on its own, without us.** Three sites had no recorded premise. All three
now have one, and two of them were repaired entirely by machinery that already existed: our
check noticed, the platform's own dispatch queue picked it up, and a strategist wrote the
premise. Nobody steered it. That was the whole thesis of the track and it is now a fact
rather than a design.

**The last untested piece fired.** When a finding of ours is genuinely fixed, our checks are
meant to notice and withdraw their own complaint. That had only ever worked in tests. Today
it worked on a live site, closing its own finding with a written reason. It also, incidentally,
prevented a real accident — another team had just rewritten that site's strategy by hand, and
our stale finding would have caused the platform to overwrite their work.

**A piece of code shipped this afternoon and was verified where it matters.** It fixes a
silence: for a kind of site we have no rule for, a check used to return "nothing found", which
downstream is indistinguishable from "I looked and it was fine". It now says "I have no rule
here, so I examined nothing". It went out with the fleet and filed exactly the right row on
exactly the right site, marked as something no agent can be sent to fix — because there is
nothing to fix, only a decision to take.

**And we found one of our own claims to be fiction.** Our notes from the 10th recorded a
finding on a site that had never produced one — the site had never been examined at all. It
was not an unmarked guess: the row had a "verified how" column and it was filled in, with the
*reason* the check would have filed something rather than any observation that it had. That
distinction is the whole lesson, and it is now written down in three places, because a
justification sitting in an evidence column reads as evidence and will do it again.

## Where we are now

**B1, B2 and B3 are complete, live, and proven end to end** — detection, repair and
withdrawal, all three exercised on real sites with the evidence read at the artefact rather
than at a status column. The estate has twenty-two sites and every one of them now records
how it makes money.

**B4 — the analyser that judges whether an offer is actually good — has not started, and it
opens with a decision rather than a design.** We measured its inputs before designing against
them, and the measurement contradicted the reason B4 was scheduled next. The fields the
analyser needs to judge an offer exist on seven of our twenty-two sites, not on all of them.
The cause is benign: those fields were introduced in early August, and every site whose
strategy has been written since carries them. The back catalogue simply has not been
refreshed — using exactly the safe-refresh capability we built a fortnight ago and have never
used for its purpose.

## Where we're going

**One question for the owner, and it is a short one.** Refresh thirteen sites' premises first,
one dispatch each, so the analyser sees a consistent estate — or let the analyser work with
less on some sites and declare that it did. We recommend refreshing first: it is the same call
already taken on the 11th for the three sites that had no premise at all, that went three for
three, and it means the analyser's first verdicts are comparable across the estate instead of
carrying an asterisk per site.

**With one warning attached.** Two of those fifteen sites do not carry a machine-written
premise — they carry the owner's. A blanket refresh would write over his own voice direction a
day after he gave it. Those two are excluded until he decides separately, and the exclusion has
to be made by asking who wrote each record, not by looking at its date.

**B4 has also acquired a customer while we were not looking.** The team working on copy quality
asked us, independently, for exactly what B4 produces: a ranked answer to "what is this reader
trying to do, so what should this page say first?" Answering them turned up something better
than a new design — most of that answer is already written on each site and nothing reads it.
On the site whose copy the owner rejected, the site's own stored promise says the same thing
the owner said, and disagrees with the brief that produced the copy. Nobody had put those two
documents side by side, because nothing ever does. That comparison is now the cheapest useful
thing on the list, and it is smaller than B4.
