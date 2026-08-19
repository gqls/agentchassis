# Where we are — release enumeration (`bugs_open/237`)

Plain-prose running log for the owner. Append only, newest at the bottom.

---

## 2026-08-18 — the frozen checks are not just old, two of them are blind

Short version: the last open question on this lane has an answer, and it is worse
than we thought last time.

The lane started because a service could be running in production and be in none
of the release tooling's lists, so no release would ever move it. That was fixed —
there is now one list per set, and a gate that reads the filesystem and fails the
release if a service exists but no release path would ever deploy it.

What was left over was six services the gate deliberately ignores, because the
release does not build their images at all. Four of those six are the estate's own
daily checks — the ones that run every morning and tell us whether the platform is
healthy. Last session I said they were "running August-6th logic", then corrected
myself: their *own* code was up to date, and only the shared platform code they
link was stale. That correction was right as far as it went, and it made the
problem sound smaller than it is.

Here is the bit that changes the picture. Two of those four checks work by walking
a list of every action the platform knows about — and **that list is baked into the
binary when the image is built.** So a frozen image is not just running old logic;
it is looking at an old *inventory*. Any action added since the image was built is
simply not in the list, and the loop never visits it. The check does not say "I
skipped four things". It says nothing, and a clean report from a check that cannot
see half a problem looks exactly like a clean report from a healthy system.

Measured this morning:

- `removed-config-keys-check` can see 165 of the 169 registered actions.
- `shared-output-fields-check` can see 161 of 169.
- Both ran today, at 06:25 and 07:10, on those frozen images. I checked at the
  cluster, not in the repo.

So every clean run since 8 August has been vouching for actions the binary could
not look at. Nothing has gone *wrong* as a result that I can point to — the damage
is findings that should have been raised and were not, which is the kind you only
discover later.

The other two of the four are less bad and I want to be careful not to overstate
them. `component-render-check` does not walk that list; one function it calls has
changed, and somebody should read it before claiming anything. `verifier-remit-check`
looks genuinely fine — neither of the two shared symbols it uses has changed since
it shipped. I have deliberately not folded those two in with the first pair.

One detail worth knowing because it is a bit alarming: two specific actions,
`publish_site` and `retract_asset_files`, are missing from all four check binaries.
Those are the *same two actions* CLAUDE.md already records as having been invisible
to a different check for a different reason (a hand-maintained list in a Python
file). Two unrelated blind spots landed on the same pair. Both fail in the same
circumstance — whenever someone adds an action — and neither notices.

I also checked whether the frozen list is really six, or whether we had missed
more. It is six. The other daily checks that looked suspicious turned out to run
SQL against a stock Postgres image, so they carry no build of ours at all and are
not this problem (they have their own staleness trap, which is documented
elsewhere). Two Ollama services pin `latest`, which is a separate trap and not
part of this bug.

**What I need from you is one decision, and it can be split two ways** — one answer
for the four checks, a different answer for the two GitHub runners. The three
options and my recommendation are in the chat and in `PLAN_2026-08-18_...`. The
short form: for the four checks I recommend making the release build them like
anything else, and additionally making the coverage gate fail when a service's
pinned image predates the last change to the code it is built from. For the two
runners, that image is Ubuntu plus a pinned upstream tarball and rebuilding it
every release churns it for nothing, so I would unstick them once by hand and
exempt them explicitly, in writing, where the gate can see the exemption.

The one thing not to do in the meantime is run `release-github-runner` as a quick
fix. It moves one of the two runners and leaves the other untouched and unmovable,
which is precisely the state that produced this bug.

---

## 2026-08-18, later — you ruled, and it is built

You ruled all six into the fleet release: the four checks fold in now with the
content-change trigger to follow, and the two runners fold in as well rather than
getting the written exemption I suggested. That is done and committed.

Two things turned out better than expected. The `-vmsites` runner, which the
handoff described as having no way to move it at all, needed no new machinery —
both runners pin the *same* image at different tags, one image and two
deployments, so it just needed declaring in the list we already had. And folding
these six in has a second effect I had not fully appreciated when I wrote the
options up: the coverage gate only watches services whose image the release
builds, so all six were outside what it could ever see. Bringing them in means the
gate now watches them too. I proved that both ways — under the old lists none of
the six shows up however hard you probe it, under the new ones all six do.

One thing needed care. Retagging the runners is only safe *because* the release
now also builds and pushes that image; if someone later removes it from the build
list but leaves the runners in the deploy list, both runners break together with a
failure to pull the image. The old comment in the file warned about exactly that,
and its premise has now flipped, so I have written the new hazard down where the
old one was rather than just deleting it. I also checked mechanically that what
the build actually produces matches the declared list exactly, in both directions.

While I was in there I found `build-github-runner` was still building from the
whole shared working tree — the thing we inverted for every other service back in
July, so an image could pick up any session's half-finished work. It only needs
one tracked file, so switching it to the safe build was a one-line change.

**What is still true: nothing has actually moved yet.** This is makefile
plumbing, and it does nothing until a release runs — which is yours to run, since
releases are whole-fleet. Until then the two blind checks are still blind. When
you do run one, the test that tells us it worked is re-running the registry count
against the new build: it should read 169 of 169 on all four, and it is a test
that can genuinely fail.

There is also no council review on any of this, and that is not an oversight —
the gate refuses makefile-only submissions by scope, client-side, so no commit
here claims one.

---

## 2026-08-19 — it worked, and the lane is essentially done

Your release picked it all up. Every one of the six moved on the first release
after the ruling, with nobody having to remember anything, which was the whole
point.

The four daily checks are now on the current build and ran this morning on it. Both
GitHub runners moved too — and this is the one I would call out, because the
`-vmsites` runner had not moved since mid-July and the other one had not moved
since **April**. That older one was the one missing `rsync` and `ssh`. That gap is
closed now, on the same roll, without anyone doing anything special for it.

The test I said should be run is the one that could have failed, and it passed: the
inventory the checks carry now matches the platform exactly, 170 out of 170,
including all four of the actions that were invisible on Monday. Worth noting the
platform added a new action overnight — 169 became 170 — which is precisely the
churn that caused this problem in the first place, and it is now covered
automatically.

I want to be straight about one thing I could not show you. I tried to prove the
difference at the check's own output rather than just at the version number, and I
could not. The check's report is identical to the day before. That is not a
problem — it turns out none of the four newly-visible actions is the kind that
would show up in that particular report, so the report *couldn't* have changed. But
it also means the honest evidence here is "the right code is demonstrably in the
image", not "we watched the behaviour change". I have written that down rather than
quietly presenting the version numbers as if they were the same thing.

One small thing I found while checking. The four check images and the runner image
are built without the stamp that lets you ask a binary which commit it came from —
every other backend service has it. It did not matter while they were nobody's
concern; now that they are release images, it means proving one of them moved
requires reasoning about the release instead of just asking it. That is a small,
separate fix and I have written it up as its own item rather than letting it
enlarge this one.

**So: what is left.** Two things, and neither is this bug. The first is the trigger
you asked for as the follow-on — the rule that fails a release when a service's
pinned image is older than the code it is built from, which is what would catch a
*seventh* service without anyone remembering. The second is that stamp fix. There
is also one loose end I have deliberately never claimed either way: whether the
render check was ever actually wrong. My honest answer is that we do not know, it
is now moot going forward, and finding out is optional.

My recommendation is that we close this bug and carry the trigger separately, since
the thing the bug describes is fixed, live and verified — but the trigger is what
stops the *next* one, so it should not quietly evaporate when the bug closes.
