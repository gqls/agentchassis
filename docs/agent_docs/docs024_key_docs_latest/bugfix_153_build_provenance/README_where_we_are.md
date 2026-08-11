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

---

## 2026-08-11, morning — it worked, and it caught something on its first outing

A fresh build went out this morning as `v1.0.1284`. Two things to report.

**First, the good news: the stamp held.** All fourteen services came up announcing which commit
built them. Last night's roll proved the mechanism worked on the release we watched it ship in;
this one proves it survives a release nobody from this lane touched. That's the difference
between a thing that worked once and a thing the platform now does.

**Second, and this is the interesting part: it immediately found a real problem.** `v1.0.1284`
is one version tag, but the fleet running it was built from **three different commits**.

The build takes about six and a half minutes to go through all fourteen services. During those
six minutes, two other sessions committed their work. Each service's build asks git "what's the
latest commit?" *at the moment it starts* — so the five services built first got one answer, the
one built at 09:10:22 got a second, and the eight built after that got a third. All fourteen
then went out wearing the same version label.

I want to be careful about how alarming that sounds, so: **today it did no harm at all.** The
commits that landed mid-build were documentation and two test files. Not a line of production
code differs between the three groups. I checked rather than assumed.

But the reason it's worth your attention is what it would have meant on a different morning. If
one of those mid-build commits had been a real code change, eight services would have it and six
would not — under a single version number, with every dashboard and every handoff note saying
"deployed at 1284". Half the fleet would have the fix, half wouldn't, and there'd be no way to
tell which half from anything except the stamps we just added. That is the same disease this
whole bug was about, one level up: we knew the version tag didn't prove a rebuild had happened;
it turns out it doesn't identify the *code* either.

**The fix is small and the machinery already exists.** The release can decide once, at the start,
which commit it's building, and hand that same answer to all fourteen builds. You can do it today
by hand by adding one argument to the release command; making the release do it itself is a few
lines and takes away the chance to forget. I've written it up properly as `bugs_open/249` with
the evidence and the options ranked — I'd like your steer before changing the release path,
because that's the one command on this estate everything else depends on.

**One correction I owe you**, because we advertised something slightly too broadly. We told
other sessions the stamp lets them ask "did my fix ship?" and get an exact answer. That is still
true — but it answers it **for one service at a time**, not for the fleet, precisely because of
what I found today. Anyone reading the chassis stamp and concluding "so the whole fleet has it"
can be wrong. I've attached that caveat where the claim lives.

**And a confession in the same shape as yesterday's, going the other way.** When I first checked
the twelve non-chassis services, eight of them came back "no". Yesterday a broken check gave me
a false alarm, so my instinct today was "my check must be broken again" — I nearly dismissed a
genuine finding as my own instrument playing up. It took a deliberate control (ask the same pod
about a commit it *should* have, and one it shouldn't) to establish the answer was real. Worth
recording because it's the mirror image of yesterday: being burned by a false positive makes the
next true positive easier to wave away, and the only thing that tells them apart is doing the
control both times.

---

## 2026-08-11, later — your four decisions are done

**The pin is in.** A release now decides once, at the very start, which commit it is building,
and hands that same answer to all fourteen services. It prints the commit before it starts
building, so you can see it. If you ever want a deliberately older release, that still works
exactly as before. It's committed, and it takes effect the next time you run a release — I
haven't run one, and won't.

**One thing it cost, which you should know before it surprises you.** `make -n release` — the
"show me what you'd do without doing it" command — now prints the release sweep rather than the
list of docker commands underneath. There's a way to get the old output back, but it works by
making the preview actually run the sub-commands and trusting them to stay in preview mode. On
an estate where you drive releases by hand, a preview that performs a real release if that trust
is ever misplaced isn't a trade worth taking for neater output. `make -n build-backend` still
shows you everything, unchanged.

**The regression guard passed.** I built the chassis from an older commit onto a throwaway tag,
never pushed it, and checked the result: it carries the older commit, and — the part that
actually matters — it does *not* carry the current one. That last check is what separates "built
from the commit I asked for" from "has a commit in it somewhere". Then I deleted the image. So
the mechanism the pin relies on is proven, and the production test you'd have had to run is no
longer owed. I've written that decision down where the "still owed" list lives, so nobody
re-files it in three weeks as an outstanding gap.

**CLAUDE.md is rewritten.** The old instruction told every session to prove a fix had shipped by
hunting for a chosen word inside the binary. That never worked properly and on one of our
services it silently returns "not found" because the tool isn't installed. It now says: ask the
service what commit it's running — it announces it at startup — and compare. The old sentence is
struck through and dated rather than deleted, because it was true when it was written and
explains why the section looked the way it did.

**One thing didn't happen, and I want to flag it rather than let it pass.** I put the release
change to the review council and it **refused to look at it** — automatically, before spending
anything. Its remit is the three main code directories, and the makefile isn't one of them. That
rule is yours, from July, so I left it alone rather than using the override. The observation
worth keeping: the release command is arguably the most shared thing we own, and it falls outside
review, while a one-line change inside the code directories gets a full round. I'm not proposing
anything on the strength of one instance — just noting it where you'll see it.

**So this commit is honestly un-reviewed** — no trailer claiming otherwise. Written up in the
lane notes with the refusal message quoted.
