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
