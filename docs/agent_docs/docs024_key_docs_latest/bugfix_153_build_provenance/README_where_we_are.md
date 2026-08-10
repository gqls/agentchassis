# README_where_we_are — bugfix_153_build_provenance

*(the owner's running plain-prose log — append only, newest at the bottom)*

## 2026-08-10

Picked up `bugs_open/153` — a real, unowned bug filed 2026-07-29 that nobody had gotten back
to. The short version: when we build a docker image for one of our services, nothing inside
the image or the running binary says which git commit it was built from. So if someone bumps
the version tag and pushes/deploys without actually rebuilding, we get a container running
old code labelled with a new version number — and nothing anywhere notices. The person who
filed this bug actually walked into that trap themselves and burned most of a day chasing
ghosts because of it.

Checked today (10 August): still true. Live production is on `agent-chassis` version
`v1.0.1279` and the binary running in those pods still has zero trace of what commit built it.

The fix, in plain terms: bake the git commit hash into the binary and the image at build time
(this is a standard practice — most serious software does this), so anyone can ask a running
container "what code are you actually running?" and get an exact, unambiguous answer instead
of trusting the version label, which — this bug proves — can lie.

I had a planning pass done by our "fable" model to turn that idea into a concrete list of file
edits (the plan is in this folder, `PLAN_2026-08-10_build_provenance.md`), scoped
conservatively: do the stamping + verification part now, fleet-wide (it's cheap and safe),
and leave the more invasive options — like refusing to push an image that doesn't match what
was built — for a later decision, since those change how the deploy process works for
everyone and deserve a proper sign-off rather than one session deciding unilaterally.

Next: build it, get it reviewed by the automated council (our advisory review process for
platform-code changes), commit it, and prove it works on one service (`agent-chassis`) end to
end before leaving the other 13 services' matching changes for their next normal rebuild.

## 2026-08-10, later — it's built and committed, but not yet switched on

The change is written and committed. Every backend service will now stamp the exact version of
the source code it was built from into the program itself, and into the container image, so
anyone can ask a running service "what code are you actually running?" and get a precise
answer instead of a version label that can lie.

I proved it works before committing, and — this is the part that matters — I proved it the
sceptical way: I built the program *with* the change and confirmed the stamp appeared, then
built it *without* and confirmed it didn't. A test that can only come out one way isn't a test.

**Two things it needs from you.**

First, **a release**. It's built but it isn't running anywhere yet. Releases here are
whole-fleet and yours to run, so when you're ready:

```
! date; make release redeploy-agents ENVIRONMENT=production REGION=uk001; date
```

Afterwards I'll verify it on both live pods. Then there's one more test worth doing
deliberately: bump the version tag and deploy *without* rebuilding — the exact mistake this
bug is about. The service should come up wearing the new label while honestly reporting the
old code. That's the bug becoming visible for the first time.

Second, and unrelated but more urgent: **our Anthropic account has hit its usage limit.** It
stopped at 14:51 today and the API says access returns on 1 September. Everything that uses AI
is down — page builds, content, the review council, all of it. I've filed it as
`bugs_open/243`. The message says *"your specified API usage limits"*, which suggests a cap we
set ourselves rather than something imposed on us, so it's probably a setting you can change
in the console. This has happened before, on 31 July, and you fixed it the same way that day.
Worth noting it's the third single-provider outage in eleven days (this one, this one in July,
and a Gemini one on the 5th) — everything we run points at one provider with one key, so there
is nothing to fall back to. That's a decision worth making at some point, not today.

One honest caveat on the fix: because the council review system is itself down, my change
hasn't been reviewed. The commits say "submitted" rather than "reviewed", which is accurate,
and I'll resubmit properly once the AI service is back.

Also worth saying plainly: I made two mistakes today and caught both. I filed the outage bug as
if it were new when it had happened before and was already written down — I'd actually run the
check that would have told me, looked at the results, and talked myself out of them. And I
wrote a warning note that duplicated an existing one, for the same reason. Both are corrected,
and both are logged in the file we keep for exactly this.

## 2026-08-10, evening — the review came back and it said no, but not for the reason you'd expect

Thank you for the credit — the fleet came back to life at about 19:12 our time, roughly three
and a half hours after it stopped. Worth noting for the record that it recovered because *you
acted*, not because the block expired; the API said access would return on 1 September, so
that's three weeks bought back.

I resubmitted the change to the review council straight away. **It was rejected** — but read
what by, because it isn't what it sounds like.

Four of the reviewers looked at it properly. Two approved outright. One found a genuine bug in
my work (more on that below). And the "guardian" seat vetoed it — while saying plainly that the
mechanism itself *"is sound and well-evidenced — that part I'd approve on a single-service
pilot."*

**Its objection is about size, not correctness.** I changed fourteen services in one go, and it
thinks a change that touches the shared build machinery *plus* all fourteen services at once is
too much to review as a single unit, however mechanical the edits are. That's a fair position,
and reasonable people can disagree — which is exactly what happened, since two other reviewers
looked at the same change and approved it.

Our own written rule covers this case: when the veto is about *scope* rather than *soundness*,
the answer is **not** for me to resubmit with better arguments. It's to record it and put it in
front of a person. So that's what I've done, and **there's a decision I need from you** — I've
written the three options up in the bug file with the costs of each, but in short: let it stand
as is; pull it back to just one service and re-land the other thirteen one at a time; or send
the underlying question to architecture review.

**The bug the reviewer found is worth telling you about, because it's funny and slightly
humbling.** The whole point of this work is that our checks can quietly pass when they should
be raising the alarm. I wrote a new check to confirm a service is stamped — and got it wrong in
exactly that way. Because of how shell pipelines report success, my check would print a blank
line instead of "no stamp found" for an unstamped service. It could not report the one thing it
existed to report. It's fixed, and this time I tested it failing as well as succeeding.

I also answered the guardian's one factual complaint — that thirteen of the fourteen services
were unverified. They're verified now: I built all fourteen and confirmed every one carries the
stamp, including the three that are built in unusual ways. But I've published that as evidence
for you, not sent it back to the council, because arguing with a scope judgement by producing
more measurements is precisely what our rule says not to do.

Still waiting on the release when you're ready.

## 2026-08-10, night — it works, and it's running everywhere

The release went out and **it works**. Every one of our fourteen backend services now carries,
inside the program itself, the exact version of the source code it was built from. You can ask
any running service "what code are you actually running?" and get an exact answer. Before
tonight that number was zero — not "unreliable", zero, for the entire life of the platform.

The one-line version: all fourteen services report commit `d3c09cc74…`, and that commit is a
real, checkable point in our history.

I checked it the sceptical way rather than the flattering one. It's easy to grep for something
and find it; the question is whether you'd have found it anyway. So alongside "is the right
code version there?" I also asked "is a made-up version there?" (no), and — the one that
actually matters — "is a *different real* version there?" (also no). That last one is what
proves the stamp is specific rather than just "some numbers are present".

**It's already earning its keep.** Another session used it the same evening to settle nineteen
register entries that were stuck on "we can't tell if this shipped without hunting for a marker
string". That hunt is the thing this fix abolishes — it's now a one-line query, and all
nineteen came back "yes, shipped".

**A confession, because it's the most useful thing in this note.** Two services first looked
*unstamped*, and I had a very satisfying story ready: the new mechanism catching a stale binary
on its first night. It was nonsense. My *check* was broken — one of our services runs on a
different base image that doesn't include the tool I was using, and the way these commands
suppress errors, "I couldn't look" came back looking exactly like "it isn't there". Then my fix
for that was wrong in the same direction and confidently returned the same wrong answer for
every service.

I caught both, and only because something didn't add up rather than because I was careful. What
makes it worth telling you: **that's the third time today** I got a confident wrong answer from
a check that couldn't have told me anything else — and it's precisely the disease this whole
bug is about. A check that reads clean when it should be shouting. I've written all three up
in the file we keep for that, and the two traps are now warnings other sessions will see before
they touch this.

**One thing still owed**, and it's worth doing when you have a spare release cycle: the real
test. Bump the version tag and deploy *without* rebuilding — the exact mistake this bug
describes. The service should come up wearing the new label while honestly reporting the old
code. Everything I've proved so far shows it works when we do things properly; that test shows
it catches us when we don't, which is the whole point.

**On the review:** it was rejected on scope, you overruled it, and that's recorded honestly.
The commits say "submitted", not "reviewed", and I haven't dressed it up as approved.
