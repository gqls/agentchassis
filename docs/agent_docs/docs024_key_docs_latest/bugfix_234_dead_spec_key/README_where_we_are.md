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
