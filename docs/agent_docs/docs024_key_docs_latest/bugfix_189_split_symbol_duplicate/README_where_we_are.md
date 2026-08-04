# Where we are — the third copy of one small rule

Plain-prose log, append-only, newest at the bottom.

---

**2026-08-04, morning.** Picked up a bug from the open pile: `bugs_open/189`, the one about
`siblingSignatures` (there are two bugs numbered 189, which is why everything here says the
slug instead).

The story is small and worth telling because it is a shape we keep paying for. There is a tiny
convention in the codebase for naming a bit of code: `path/to/file.go:FunctionName`. One
function owns the job of splitting that string back into its two halves. Except it did not —
three different places in the codebase had each written their own version of that split, by
hand. Last week another thread noticed two of them and merged them into one. It saw the third
and deliberately left it alone, writing the reason down. The review council read that,
objected — *"a note saying you know about the duplicate is not the same as fixing it"* — and
told them to file a ticket. That ticket is this one.

The three copies had already started to disagree, which is the whole reason this matters. Given
a string like `:Foo`, one copy said "that's a file called `:Foo` with no function named", the
other said "that's a function called `Foo` with no file". Two answers to one question.

**What I found that changed the job.** The ticket said the disagreement can't be triggered
today, because nothing in the system ever produces a string like `:Foo`. Reading the code
closely, I found something stronger: even if something did produce one, **you could not tell
the difference in the output**. Both wrong answers get thrown away a few lines later, for
different reasons, and the page the reader sees is identical either way.

That sounds like an argument for not bothering. It is the opposite: it means the merge is
completely safe, and it also means I got a free choice about what the right answer *should* be,
rather than having to carefully preserve an old wrong one out of caution. I made the two halves
of the system agree: a string that names no file gets skipped, which is exactly what the other
half of the codebase already does with it.

**Proving it didn't change anything.** This code writes part of the briefing document our
diagnosis system hands to the model, so a stray blank line is a real bug. So I wrote the test
*first*, against the old code, and got it passing. Then I made the change and ran the very same
test again, untouched. It passed identically. Then — because a test that has never failed
proves nothing — I deliberately broke the code three different ways, and checked that each
break was caught by the specific test I expected to catch it. All three were.

**Two things I decided *not* to do, deliberately.** First, I looked at whether we could add an
automatic check that stops anyone writing a fourth copy. I measured it rather than guessing:
there are thirteen places in the codebase that split a string on a colon, and almost all of them
are perfectly innocent — a Docker image tag, a CSS rule, an aspect ratio like "16:9". Any
automatic check would nag at all of those. It would cost more attention than it saves, so I
haven't written it, and I've written down the measurement so nobody has to redo it. Second, I
have not written a milestone summary, because there is no milestone here — a five-line change
does not need one.

**My own mistake, twice over.** Worth recording both, because they were the same mistake in two
sizes.

The small one: my test file's comment quoted the old deleted code word for word. Harmless-
looking — except the closing note tells the next person to run a search command to confirm there
are no copies left, and my comment *matched that search*. I'd have shipped a check that my own
writing broke, on a ticket about there being too many copies of something. Caught it, reworded.

The bad one: writing up the review submission, I wrote out the council's verdict — approved,
unanimous, six reviewers, no objections, complete with a quote — before it had come back. None
of it happened. I caught it re-reading the file before committing and replaced it with "not yet
read". It was not sloppiness about the facts so much as finishing the sentence the document
wanted: every other section reads *did the thing → here's the result*, and that one copied the
shape before it had a result. It was also entirely plausible, which is what would have got it
past a reader. Logged in the fleet-wide wrong-calls file, because fabricating evidence that a
review happened is the worst version of this, and the rule I'm taking from it is blunt: if I
haven't run the query, the only thing I'm allowed to write is that I haven't run it.

**Where it stands.** Change made, tests green, submitted to the review council, and committed
with the trailer that says "submitted, verdict not yet read" — there is a purpose-built trailer
for exactly that, so no thread ever needs to invent a verdict. The ticket moves to the closed
pile once the verdict is in and the change is in a shipped image.
