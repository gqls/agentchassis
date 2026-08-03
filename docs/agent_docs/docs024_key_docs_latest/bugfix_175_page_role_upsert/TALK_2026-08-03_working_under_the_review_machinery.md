# Working under the review machinery — a read-aloud account, 2026-08-03

**Not a series SUMMARY** (those are current-state-only; the series entry for today is
`SUMMARY_2026-08-03_what_counts_as_a_live_page.md`). This is one session's account of what
the review machinery actually did, written to be read out.

---

## The short version

Over about a day I fixed one bug, which produced a second bug, which produced a third. I
put the work through the council four times. It approved three and sent one back. Along the
way it caught five things I had written down as fact that weren't, and it found one defect
I would never have looked for. It also spent a good deal of its attention on things I had
already done, because it reviews what you *say* you will do, not what you did.

That mix — genuinely catching me out, and occasionally shadow-boxing — is the honest
picture, and it's more useful than either half alone.

## The machinery generates its own work, and the work is real

The bug I started on existed because a reviewer, on someone else's fix a week earlier, had
asked: *"you've fixed one of these — how do you know there aren't others?"* There were
four others.

I fixed those four. Three reviewers then asked the same question about a different thing:
*"does any other place in the codebase still make this judgement by hand?"* About ten did,
and twenty-eight live pages were invisible to them. That became the third bug.

So there's a chain: one fix, one question, four more instances; one fix, one question,
twenty-eight blind spots. Each link was a reviewer refusing to accept that a fix was the
whole of the problem. None of the three would have been found by the person doing the work,
because the person doing the work has just convinced themselves they're finished.

## What it actually caught

Five things, and they're all the same shape — a sentence written confidently, in the same
voice as the things I *had* checked.

**A claim sourced from a comment.** I stated that a particular queue had one producer and
nothing automated reading it. A reviewer pointed out my source was a comment in a test file,
not the database. I ran the query: the claim was true. But it had been folklore until
somebody insisted, and the reviewer's whole job is to notice that a comment is not evidence.

**A helper I built that already existed.** A reviewer asked whether I'd checked for an
existing version before writing mine. I checked for a *helper*, found none, answered, and
stopped. What already existed was one line of SQL — and mine was a worse version of it,
wrong on eleven live pages. The reviewer asked the right question; I answered a narrower
one than was asked, which is its own lesson.

**A measurement filtered on an assumption.** I published a figure — twenty-eight affected
pages — from a query that quietly excluded archived ones. A reviewer noted I'd filtered on
a column without ever looking at what values it had. Seven more rows were hiding there, and
whether they count is a real question, not a rounding error.

**A claim about production that had gone stale.** I told you twice that a fix wasn't live
yet. A reviewer declined to take that on trust, saying plainly it had no way to check a
running binary and that a human should look. I looked: another session had deployed in the
meantime, and the thing I said wasn't live had been live for an hour.

**A count I typed from memory.** Twice. Once "six" when it was four, once "seven" when it
was eleven — and running the second one properly also revealed that three of the eleven
were false alarms in a rule I'd just written.

## The one that was gating, and it was right

The second submission came back as *revise*. The objection: I'd claimed all four call sites
declared a new safety flag, but my submission only showed two of them being edited.

The code was correct — all four were done. What was wrong was that I'd folded two one-line
changes into prose instead of listing them, so no reviewer could confirm the claim. And the
reason it mattered is sharp: the flag defaults to off, so *every call site I didn't edit
silently changes behaviour*. That's exactly what makes the design safe, and exactly what
makes an incomplete list dangerous — a reviewer can't tell a deliberate omission from an
oversight.

The rule I took from it: **a change whose safety depends on having swept every call site
must show the sweep, not assert it.**

## Where it is weak, said plainly

It reviews a plan, not a diff. Several objections asked for things already in the commit —
a test description updated, a call site converted — because the reviewers can only see what
the submission describes. That's a structural limit, not a failure, but it means a
carelessly written submission gets a worse review than the code deserves, and a
well-written one could in principle get a better one. The gating objection above is that
limit working *for* us; it can equally work against.

It's also slow and it isn't free. Each round is twenty to thirty minutes and real spend, and
one run was killed outright when another session redeployed the cluster mid-review. Sixteen
reviewers exist; a given submission wakes only the relevant ones — six abstained on the last
round — which is what keeps it affordable.

And it doesn't catch everything. The three false alarms in my own new rule I found myself,
by running it rather than reasoning about it.

## The cheaper machinery does different work

Three tiers, and they fail differently.

**Grep and the commit-time checks** catch spelling: a pattern written the wrong way, a
statement missing a column. Milliseconds, free, and they only see what you thought to
encode.

**Breaking your own guards on purpose** catches decorations. I did this to every guard I
wrote, and twice it earned its keep. Once a guard I'd "verified" survived being reverted —
every test stayed green, because none of the test inputs could tell the new version from
the old, so I'd been protecting nothing. Once a mutation failed to *compile* and I nearly
recorded that as the test catching it; a compile error proves nothing about an assertion.

**The council** catches the confident unchecked sentence. Nothing else does. Not the
compiler, not the tests, not care.

## What I'd say to someone asking whether it's worth it

The expensive review isn't there for careless work. Careless work is caught by the cheap
tiers, immediately and for nothing. It's there for the work that looks finished — where
someone competent has measured, tested, mutated their own guards, written it up, and is
genuinely confident.

That's precisely the state in which I wrote five wrong things in one day, and it's the
state the cheap tiers cannot see into, because every one of those sentences was
syntactically fine, plausible, and consistent with everything around it.

The strongest single piece of evidence: the machinery found a defect through a chain of
three separate people's fixes, each time by asking one question the author hadn't — *are
you sure that's all of them?* It was never all of them.
