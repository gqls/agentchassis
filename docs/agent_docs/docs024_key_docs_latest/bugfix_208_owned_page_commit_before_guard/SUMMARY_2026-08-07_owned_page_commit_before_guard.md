# SUMMARY — bug 208: the rebuild that would have destroyed live tools

**2026-08-07.** Written to be read aloud. Five parts: what we're trying to do · where we've come
from · what we've done · where we are now · where we're going.

---

## What we're trying to do

Stop the platform destroying its own tool pages.

Some pages on our sites are not ordinary prose. They are working tools — the vonc arena, the
gauntlet, a password-entropy calculator, six games-design calculators. They are marked in the
database as "owned", meaning *this page belongs to a tool; the generic page builder must not
touch it*. The goal of this work was to make that marking actually protect the page, in every
pipeline, rather than in the two places somebody happened to think of.

## Where we've come from

The marker was created in March, in migration 164, after an incident we still call the vonc
arena clobber: a generic rebuild ran over an interactive page and replaced it with prose. The
migration added two refusals — one where pages get queued for building, one where a page's
sections get saved — and both work.

They were designed against one route into the builder. The design assumed a page reaches a
generic build by being *queued* for one, so it guarded the queue and the save. What nobody
modelled was a pipeline that skips the queue entirely and selects pages straight out of the
database by their build status. Three pipelines do exactly that, and all three do their work in
an order the migration did not anticipate: they **commit the regenerated page to the website's
repository first, and check whether they were allowed to a step later.**

So the guard fired correctly, every time, one step too late. It protected the database row. The
file the public actually sees had already been replaced.

Another session found this yesterday during pre-flight for a rebuild you had authorised on
`ai-agent-orchestration.com`. It stopped, filed the bug, and handed it on. I picked it up.

## What we've done

**Established the real size of it.** Not two pages on one site as filed, but fourteen across six
domains; not one pipeline but three; and a second symptom nobody had noticed — because these
loops are not configured to tolerate a failed page, the late refusal did not merely lose the
owned page, it **killed the entire run**, silently skipping every page queued after it. Anyone
who has seen a bulk rebuild quietly stop half-way through a site now has a candidate explanation.

**Checked whether it had already happened, rather than assuming.** I fetched all fourteen live
pages and hashed them. Thirteen were serving working tools; the fourteenth had never been built
at all. We were fixing this *before* it fired, not after — and those hashes became the control
set for proving we changed nothing.

**Fixed it at two ends.** Owned pages are now excluded when pages are selected for a generic
build, and — if one arrives by some other route — refused at the moment of composition, one step
before the commit that would destroy it. The refusal reuses a "skip this page" signal the system
already had and the committing step already understood, so **no pipeline configuration changed
anywhere**: no config half to forget, no ordering hazard, nothing for another session to trip on.

**Declined the obvious simplification, and wrote down why.** The tempting fix is to guard the
commit itself, which would cover everything at once. It would also have broken the only routes by
which tool pages *legitimately* go live — the migration says so in writing — and the symptom
would have looked like tools mysteriously ceasing to update, nothing like a guard misfiring. That
is now recorded as a landmine, because it is precisely what a future session would reach for in
good faith.

**Put it through the council.** Approved: thirteen reviewers, five advisory objections, none
high-severity. I read the verdict rather than banking the decision, and acted on three of the
five — unifying a predicate I had duplicated, and turning a silent failure window into a counted
one. A fourth I answered with an experiment instead of an argument (below). The fifth turned out
to be a false alarm we caused ourselves: a reviewer concluded there must have been a previous
attempt at this bug, because it found my own landmine entry — which I had written an hour
earlier, about my own unshipped code, and published to the corpus it searches.

### Mutation testing — the part worth dwelling on

The convention here is that a guard is not proven by a test that passes; it is proven by
**deliberately breaking the guard and requiring a specific, named test to fail.** You change one
line, run the suite, and watch. If nothing fails, the test was decoration.

I ran three rounds. **Every round found something, and every time it was the mutation I expected
to be boring.**

**Round one — a test of mine passed with the guard deleted.** I had added an early exit and
written a test for it: call the function in the skipped state, assert it reports "skipped". Green.
Then I deleted the guard, and it was **still green**. The function has an older, unrelated path —
"page not found" — that returns the *same shape*. With my guard gone the code fell through, hit
that path, and produced an identical answer for a completely different reason. My test was not
testing my guard; it was reading a value two different roads arrive at. Had I stopped at green,
I would have shipped a guard whose test actively vouched for its absence, and the next person to
refactor it would have had a passing suite telling them everything was fine.

The fix is to assert the **discriminator** — the one field only my path fills — rather than the
shared flag. And there is a cheap way to see this coming without running anything: before writing
the assertion, look through the function for every other place that returns the key you are about
to assert on. More than one, and the flag cannot be your assertion.

**Round two — after refactoring, two guards had no test at all.** Answering the council I merged
two duplicate ownership checks into one shared function. Re-running the mutations found that
forcing the shared check to always say "not owned" broke only *half* the tests I expected:
nothing anywhere asserted that the **save** step refuses an owned page. That refusal shipped
months ago in migration 164 and had never been tested, because the existing suite feeds it a
page marked "generic" — so the guard was inert in every test that touched it, and my unification
quietly inherited that blind spot. A second mutation showed the failure-reporting I had *just
added to satisfy a reviewer* also had no test.

**The lesson, stated plainly: re-mutate after a refactor, not only after writing the guard.** A
refactor is exactly the moment a guard stops being load-bearing without anything turning red.

All told: six mutations across the session, four guards now provably load-bearing, and two tests
that existed only because a mutation embarrassed the previous version of them. The cost was
minutes. Three real defects in my own work — including one in the code I had written *to answer
a reviewer* — would otherwise have shipped looking tested.

## Where we are now

**The fix is live.** It went out on chassis v1.0.1262 this morning and I have verified it in the
running binary rather than trusting the tag: the strings my change adds are present, the string
my change *removed* is absent (which is what proves it is genuinely the latest version and not
an older image that happens to share some symbols), and a deliberately fabricated string returns
nothing, which proves the check itself is honest. All forty-one processes running that binary
report a single identical image, so this is the whole fleet by identity, not a sample.

I re-checked the fourteen pages: **thirteen byte-for-byte identical**, and the fourteenth is
still the 404 it always was, untouched in the database since June. Nothing was disturbed.

**One honest gap: the guard has not yet been seen to fire.** Everything above proves the code is
there and that nothing broke. None of it proves a refusal actually happens when an owned page is
selected, because no rebuild has been dispatched at a site with owned pages since the roll. A
clean result from a guard that has never been tried is not evidence — that is the trap this
estate has written down more than once, and I am not going to fall into it in my own summary.

## Where we're going

One thing left: **make the guard bite, in production, where we can watch it.**

The obvious candidate is vonc.com, which has exactly three owned pages waiting and no ordinary
ones, so a rebuild there would touch nothing else. I am deliberately *not* doing that, and the
reason is the whole point of this bug: those three pages are the arena, the gauntlet and the
archetype quiz. If my fix is wrong, that experiment destroys them. You do not test a safety net
by dropping the thing you built it for.

Instead I want to create a **throwaway** page, marked owned, on a site with nothing else queued,
and dispatch a real rebuild at it. If the guard works, nothing is built, nothing is committed,
and a review item appears explaining why. If it fails, the only casualty is a page I invented for
the purpose. That needs your go-ahead, because it is a real dispatch against a live site.

After that: the follow-up defect I found while fixing this one and deliberately left alone
(`bugs_open/210` — a page whose content generation fails gets marked "deployed" anyway, so the
rebuild request is silently forgotten). It is the same family, but fixing it changes retry
behaviour across the whole fleet, and that is a decision rather than a tidy-up.
