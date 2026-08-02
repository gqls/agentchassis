# SUMMARY — 2026-08-02 — one derivation for a deployed asset's path

Written to be read aloud. Five parts: what we're trying to do · where we've come from ·
what we've done · where we are now · where we're going.

---

## What we're trying to do

When the platform generates an image — a hero, a logo, an icon, a favicon, a social card —
two things have to agree about where that file lives: the code that **commits** it into the
site's repository, and the code that **points a page at it**. If they disagree, you get a
file nobody references or a reference to a file that isn't there, and both halves look
correct in isolation.

`bugs_open/168` was filed because the function everything asks for that path,
`storage.DeployedWebPath`, gets one case wrong: it answers `og_card.png` when the real file
is `og-card.png`. The goal was to fix that in a way that fixes the **class**, not the case.

## Where we've come from

The case had been found twice already and patched neither time at the root. The
`bugs_open/142` lane discovered the wrong path, declared a small map of the correct
filenames, and left a deliberate tripwire test saying *"if someone later teaches
`DeployedWebPath` about this case, this test fails and tells them to collapse the map into
it."* Then the `bugs_closed/128` lane nearly shipped a check that would have reported a
broken favicon and social card **on every site in the fleet**, because it trusted the
function's own doc comment. It caught itself, added a local branch, and its reviewers said
the residual should be its own item. It was written down as "its own item" and never filed —
until a council seat objected to exactly that, which is how 168 came to exist.

So the through-line is a defect that three lanes each contained locally, and nobody moved.

## What we've done

**Found that the filed explanation was wrong, before acting on it.** The bug said the file on
disk has a hyphen where the helper says underscore. For the main deploying code that is
false — it produces the underscore too; the two agree. Acting on the filed mechanism would
have meant fix candidate 2, "swap underscores for hyphens everywhere", which would have
*created* the drift it claims to remove. The real defect is that there are **two writers**
and the function's inputs cannot express which one published a given file.

**Made it one function.** `storage.DeployedAssetPath` is now the single derivation, called by
the writer *and* all six readers. They cannot drift, because there is nothing to drift from —
where before they were held in step by a comment claiming they matched.

**Proved every guard by breaking the code**, seven times, rather than trusting a green test
run. The most valuable of those was run *after* the fix, to check that removing a local
protection hadn't quietly removed the protection.

**Put it through the review council three times.** Round 1 found one real code defect and
three things I had done but not *shown*. Round 2 found something better: **I was wrong.**

## Where we are now

The council's most valuable catch was the one I argued against. My change made it possible
for the deploying code to **overwrite a site's live social card**, and I twice told the
council that was unreachable, with measurements attached. Two seats kept pressing. When I ran
the query to prove them wrong, it returned **eleven queued work items** that would do exactly
that — two of them in a state that gets picked up and run.

Both of my measurements were sound and neither could answer the question: one was about work
items created *from now on*, the other about *readers* when the risk was in a *writer*. The
population that answers it is the queue that already exists. **Fixing a check stops new bad
items; nothing goes back for the ones already queued.** The safety I was leaning on came from
a July fix, and was only ever true of July's code — which I had just changed.

So the guard is **built, not filed**: the deploying code now refuses to touch a favicon or
social card, before it downloads or commits anything. Mutation-proven both for existing and
for *position*, because a guard that fires after the commit is not a guard.

**One piece of luck, stated as luck:** none of this is live. The risky change and its guard
will ship in the same image. Had the earlier commit already rolled, this would have been a
live incident rather than a review comment.

Everything is committed. `HEAD` builds and tests green. Round 3 is with the council. Two
wrong calls of mine are logged fleet-wide with the cheap check that would have caught each,
and a follow-up (`bugs_open/179`) tracks the one hole that remains genuinely open.

**`bugs_open/168` is still OPEN, deliberately.** The bar here is *fixed AND live*, and a
pod-grep with three controls confirms the running binary predates the fix. Closing it now
would be a false claim.

## Where we're going

Three things, in order:

1. **Read the round-3 verdict and act on it.** Blocked right now on cluster access — the
   kubeconfig token has expired, which is a routine 3-day thing the owner refreshes.
2. **Verify after the next roll**, on both replicas, with a positive control *and* a negative
   one — a string the change removed, expected to read zero. Only then does 168 move to
   `bugs_closed/`.
3. **Decide two things that are not code.** Whether the eleven stale queued items should be
   re-pointed at re-derivation (what they actually want) rather than merely refused; and the
   larger question `RFC_009` raises — the platform reconstructs an artefact's identity from
   its metadata instead of reading what the writer recorded, which is the shared root of
   `bugs_open/152`, `155` and `179`. That one wants designing with those lanes, not inside a
   path helper.
