# Where we are — bugfix 234 (plain prose, append-only, newest at the bottom)

## 2026-08-10 — picked up, checked, and planned

Yesterday another thread discovered that three of our workflow steps have been writing
their instructions under a name the system never looks at — like posting letters into a
slot that isn't connected to anything. The most important casualty: whenever the
improvement loop fixes things on a site and asks for the pages to be re-assembled, it also
asks "and refresh the shared header and footer" — and that request has been silently
thrown away every single time. Sixteen requests, all lost, none noticed.

I checked the bug is still real (it is — all three steps unchanged, the sixteenth lost
request was filed yesterday lunchtime), and checked nobody else is working on it (they
aren't; the thread that found it explicitly handed it over).

Two things had changed since the bug was written up, and both made the decision easier.
First, the safety net we were waiting for — the one that notices when a rebuild is about
to destroy hand-made edits to a site's header or footer — is now live and has already
caught its first real case. Second, the "risky" behaviour we'd be switching back on turns
out to be something eight other parts of the system already do every day; this path was
the odd one out.

So you made two decisions: switch the lost instruction back on (the safe spelling this
time), and close the whole class of bug two ways — make this particular action reject any
config key it doesn't recognise (the strict switch that was always planned once we'd
checked every live use, which we now have), and give the platform a way to say "this key
name is retired, here's what replaced it" so anyone who writes the old name gets told
loudly instead of being silently ignored.

The order matters: fix the data first (that's instant), then ship the stricter code (that
waits for the next release). Doing it the other way round would break the very agents
we're fixing. Council review happens before the code lands, as usual.

## 2026-08-10, later — both halves done; two proofs still owed

The data fix is applied and live. Before running it for real I deliberately broke it three
ways and watched its own safety checks catch each one — that's the standard now: a check
you've never seen fail proves nothing. All three steps now spell their instruction the way
the system actually reads, and the lost "refresh the header and footer" request is switched
back on.

The stricter code is written, reviewed by machine tests (including deliberately breaking
each new guard to watch the right test fail), submitted to the council, committed, and
baked into image v1.0.1278, which is pushed and waiting for the next fleet release. I did
not deploy it myself — releases go out whole-fleet, and that's your button.

Two things remain, both waiting on the world rather than on work: the improvement loop
files one of these rerender requests roughly twice a day, and the first one filed since the
fix must be seen actually carrying the flag (that's the proof that matters — the config
merely *looking* right is exactly what this bug was); and once the fleet rolls onto the new
image, a quick check that the new rejection actually rejects. Both checks are written down
in the runbook with the exact commands. The council's verdict was still being deliberated
at the time of writing; the commit carries the pending-review marker so the coverage report
credits it automatically when the verdict lands.

## 2026-08-10, midday — the council said no to how, not to what

The review came back split in an interesting way. Nobody disputed the bug, the data fix,
or the evidence. What drew a veto from the safety seat is that the new "reject a retired
config key loudly" behaviour lives in machinery every agent's messages pass through, and
it arrived packaged inside a bug fix rather than as its own reviewed change. The
architecture seat, in the same round, looked at the same facts and said the opposite —
fine to proceed, but write down the accumulated design before anyone adds to it again.

House rules for exactly this situation (it has happened before): the shipped code stays,
the disagreement goes to you with the design written down. So I've filed RFC 021 with the
two questions that are genuinely yours: how much ceremony should a hard-failing check on
shared machinery require before it goes live, and should the stricter behaviour ride the
next release as built, or be softened to warning-only until you've answered the first
question. Every concrete complaint the objecting reviewers raised — a measurement they
wanted re-run, a loose end they wanted tracked, a claim they wanted proven from the code
rather than prose — has been answered and recorded. Nothing further ships on this
mechanism until you rule.

## 2026-08-10, evening — your three rulings are done, and the stricter code is live

All three decisions are executed. The daily automatic check exists and I proved it works by
running it once by hand before trusting it: it walked 181 agent definitions, found nothing
carrying a retired key, and — importantly — wrote its "all clear" down. That last bit is
deliberate: if the job ever silently stops running, the absence of a note is the signal.
Silence and "nothing wrong" must not look the same.

The stricter behaviour is now genuinely live, not just built: the fleet rolled onto a new
image and I confirmed the new code is inside the running containers, on both of them.

The third decision — retiring the two dead keys on the page-status action — is done at the
data layer and the code for it is built and pushed, waiting for your next roll.

Two things I got wrong today, both caught before they shipped and both written up. The
instructive one: I wrote a live measurement into a code comment, and it was out of date
within the hour because another session changed the thing I'd measured. The comment now
tells the reader how to ask the system rather than what it said once. That is the general
lesson on a tree this many people share.

One thing is still owed and it is not in our hands: the live "canary" test — deliberately
mis-configuring a throwaway agent to watch the new rejection fire — could not complete,
because the whole in-cluster fleet is currently stopped on an account-level API cap. Nothing
is being dispatched, so nothing can be rejected. The evidence we do have is solid (the code
is provably in the running binary, and the behaviour is pinned by tests that I broke on
purpose to confirm they catch it). I'll run the canary when the fleet is working again.

## 2026-08-11 — both proofs are in; the work is done

The new build carries everything, on both machines, and the fleet is running again. So I
ran the two outstanding tests and both came back the way we needed.

The first: I created a throwaway agent deliberately configured with a nonsense setting, and
the system refused to run it — the error message names the offending setting and tells the
author it is a definition error rather than something harmless. It was classified as
permanently broken rather than retried forever, and recorded. That is the whole point of the
change, observed working on live traffic rather than inferred from tests.

The second: I created another throwaway agent carrying the improvement loop's exact
configuration, and the record it filed came out carrying the "refresh the header and footer"
instruction — the thing seventeen previous records could not do. That is the original bug,
observed fixed, at the place that matters.

I did it with a stand-in rather than waiting for the improvement loop itself, for a reason
worth knowing: that loop only files these records when an audit has actually found problems
worth fixing, and it hasn't run since Sunday afternoon. Waiting could have taken days, and
the alternative — pointing it at a real customer site to force the issue — would have run a
full audit, a round of fixes and a re-render on live pages just to prove a one-word change.
The stand-in used the loop's own settings copied word for word, so what it proves is the
same thing.

Both stand-ins were deleted afterwards, and I re-checked that the system is clean — that
matters, because leaving a deliberately-broken test agent lying around would corrupt the
daily check we just built.

One correction worth recording. Yesterday I told you the first test was blocked by the
fleet being down. The fleet was down, but that was not why it failed: my test was watching
the wrong place entirely — it looked for a container that never gets created when a
configuration is rejected, because the rejection happens before anything starts up. The
outage gave me a believable reason for the silence, so I stopped digging. That is the
general lesson and it is now written down: a convincing outside explanation for "nothing
happened" is exactly when you should double-check your own instrument.

## 2026-08-11, afternoon — approved, and the review earned its keep

The council approved it, at the fourth attempt. I want to record why that took four rounds,
because it is not a story about bureaucracy: two of those rounds found things that were
actually wrong, and one of them was something I had written down as a fact.

Round two: a reviewer pointed out that the daily checker I'd built re-did, in Python, a job
the platform already does in Go — and that having two copies of "look at every step of every
workflow" is exactly how this system has been bitten before, because the two copies go blind
in the same way and then agree with each other. My defence was that the neighbouring checker
does the same thing. That is a reason the problem exists, not a reason to add to it. It
turned out another team had already solved it properly, so the checker now runs the real Go
code. Four hundred and ninety lines deleted, including the test that existed only to police
the duplication.

Round three caught a genuine mistake of mine. When retiring the old `commit_from` setting I
wrote that its purpose was now served by the new "which version is running" stamp. It isn't.
That stamp records which build of our software is running; the old setting was about which
version of a *published page* went out. Two completely different things, and I'd matched them
up because the words looked similar and I'd been working with the second one all morning. The
retirement note now says plainly that nobody has built this yet, and warns against the exact
mix-up I made. It matters because that note is instructions to whoever comes next — sending
them to build on the wrong foundation is worse than telling them there isn't one.

There was also a smaller one: a reviewer warned that a scheduled job can silently sit
failing to download its image while appearing to run. I shipped past it and it happened
within four minutes. The lesson I've written down is that an advisory warning is a free
prediction, and ignoring it wastes the prediction.

So: the fix is done, live, proven by watching the system actually refuse a bad configuration
and actually deliver the instruction that used to vanish, and approved. The one thing not yet
in your hands is a new chassis image carrying the last correction — built and pushed, waiting
for whenever you next roll.
