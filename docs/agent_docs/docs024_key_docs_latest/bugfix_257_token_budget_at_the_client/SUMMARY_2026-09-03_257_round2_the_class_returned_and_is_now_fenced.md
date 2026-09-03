# 257 — round two: the same fault came back in a shape that beat the fix, and it is now fenced by a test

*2026-09-03. Third in the series, after `SUMMARY_2026-08-16_257_budget_at_the_client.md` and
`SUMMARY_2026-08-16b_257_live_and_verified.md`. Written to be read aloud.*

## What we are trying to do

When we ask a language model to write something, we have to tell it how long the answer may be. That
number should be a setting an operator can change, not something typed into the program. This bug is
about making that true everywhere, and — just as important — about being able to *tell* whether it is
true, because the failure is silent: a limit that is being ignored looks exactly like a limit that is
being obeyed, right up until an answer gets cut in half.

## Where we have come from

In August we found that only one function in the whole system actually read that setting and passed it
on. Anything that talked to a model provider directly quietly used its own built-in number instead. The
configuration looked perfect and did nothing.

We fixed that in the right place: the provider clients themselves now read the setting from the
configuration they were built with, so a caller that asks for no particular limit inherits the
configured one. That went through the review council, shipped on 16 August, and was verified against the
running binaries. It is still true and nothing since has undermined it.

## What we have done

Three weeks later, the same fault reappeared in a new shape — in code written *after* the fix, by
authors who had not read the note warning against exactly this. Two new pieces of code each
re-implemented the "which limit wins" rule by hand, and each finished with a number typed in: 2000.

That is worse than the original fault, for a reason that is easy to miss. A limit sent with the request
always beats the one the client would have worked out for itself. So a caller that types in a number can
*never* inherit the configured one — passing nothing is now safer than passing a number. And in one of
the two cases the typed-in number happened to equal the configured number, which meant our own records
read `2000` whether the configuration was being honoured or thrown away. There was no query anyone could
run that would tell the two apart.

This session deleted both hand-written copies, along with three more we found on the way, and put all
five places that talk to a model through the one shared piece of code that reads the setting properly.
Then we added a check that fails the build if anybody types a limit into that part of the system again —
package-wide, not a list of files, which is why the previous check missed both new offenders. We also
proved the check can fail: we put each of the four mistakes back, one at a time, in a throwaway copy of
the tree, and confirmed each is caught with a message that says what to do.

**The three extra places are the interesting part.** Two were listed in this morning's plan as "fine"
and were not — one was handing the model an empty settings object, the other would pass a limit of zero
straight through to a provider that rejects it. The third was not on any list we have ever made: the
news-fetching code talks to a different supplier over ordinary web requests rather than through our
client, so every survey we have run — four of them, over three weeks — was structurally incapable of
seeing it. It had 4,096 typed in. The check found all three on its first run, because it asks a
different question: not *who calls our client*, but *is there a number typed in next to a word meaning
"how long may the answer be"*.

## Where we are now

The code is committed and has gone to the review council; the verdict is not in yet. Nothing is live
until the next fleet build, and this bug's own bar for closing is *fixed and live*, so it stays open.

For the sites we build today, nothing changes: every step affected already states its limit in
configuration, and each will send exactly the number it sends now. We measured that rather than assuming
it. What changes is that those numbers are now genuinely the operator's, and that the next person to
type one in gets told by a failing build on the day they do it.

We also found, and deliberately did not touch, four steps of the site-adoption agent that state a limit
— one of them asks for 32,000 — in a place nothing reads. All four are running at 16,000. Nothing is
being cut off, but somebody once asked for double and quietly got half. Moving that setting to where it
would be read takes effect the instant it is applied and costs money, so it is a decision to be taken,
not a side effect of a code change.

## Where we are going

Three things wait on a decision rather than on work.

The first is whether to merge the two remaining near-duplicate copies of the "which limit wins" rule.
One of them serves most of the fleet and cannot simply call the other — the languages of the two
programs point the wrong way round for that — so this is a real design question, not a tidy-up.

The second is whether making direct model calls visible to our truncation monitoring belongs to this bug
or to a lane of its own. Today's work sharpens the question: more reporting is not automatically more
truth, because reporting a number that a typed-in default could also have produced tells you nothing.

The third is the four dead settings above.

And the routine remainder: when the next fleet build goes out, confirm at the running binary that the
change is aboard, then watch a step whose configured limit is *not* 2000 — the 16,000 one is the right
subject — send the larger number. Choosing a 2000 step for that check would reproduce the very blindness
this round was about, in our own verification, where it would be least visible.
