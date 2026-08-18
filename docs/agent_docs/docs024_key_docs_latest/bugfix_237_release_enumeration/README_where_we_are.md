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
