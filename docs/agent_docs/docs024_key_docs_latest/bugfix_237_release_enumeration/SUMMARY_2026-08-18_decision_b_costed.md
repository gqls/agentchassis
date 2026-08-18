# SUMMARY — 2026-08-18 · the frozen checks, costed

Current state only. Written to be read aloud.

---

## What we're trying to do

Make it impossible for a service to be running in production and be invisible to
the release tooling. Not "fix the one service that was invisible" — remove the
condition that let it happen, so the next one cannot happen quietly.

## Where we've come from

The release tooling held four separate hand-written lists of "the services a release
deploys", and nothing checked those lists against what was actually on disk. A
service could exist, run in production, and appear in none of them, in which case no
release would ever move it. That is what happened to `render-audit-adapter`: it sat
86 image tags behind the fleet for months, still serving a binary with a known
credential leak, and it was invisible because the obvious way of checking — "is the
image live?" — looks at the *other* service that shares that image, which was
perfectly up to date.

The individual case was fixed on 10 August. The class fix followed: one declaration
per set instead of four hand-written lists, the lists became loops, and a new gate,
`check-release-coverage`, reads the filesystem itself and fails the release when a
service exists that no release path would ever deploy. That shipped, and the fleet
release v1.0.1307 exercised it without incident.

Two questions were left open. The first — four old deploy rules that shadowed the
generic one and skipped its safety check — was settled on 17 August; the rules are
deleted and the generic rule serves them. The second was a list of six services the
new gate deliberately ignores, because the release does not build their images at
all. Four of those six are the estate's own daily health checks.

## What we've done

We have costed those six, which was the one task standing between the bug and
closure, and the answer changed the shape of the problem.

The four checks are not merely running old code. Two of them work by walking a list
of every action the platform knows about — and that list is compiled into the binary
when the image is built. A frozen image therefore has a frozen *inventory*. Any
action added since the build is simply absent from the list, the loop never visits
it, and nothing is reported. Measured: `removed-config-keys-check` can see 165 of
169 registered actions, `shared-output-fields-check` 161 of 169. Both ran on the
morning of 18 August, on those frozen images, confirmed at the cluster rather than
inferred from the repository. Two specific actions, `publish_site` and
`retract_asset_files`, are missing from all four check binaries — and those are the
same two actions already on record as invisible to a different check for an entirely
different reason.

We were also careful about what the measurement does *not* say. The third check
does not read that list at all, though one function it calls has changed, so it is
suspect and unproven. The fourth looks genuinely healthy: neither shared symbol it
uses has changed since it shipped. We have kept those two out of the affected group
deliberately, and we have not claimed any check has produced a wrong answer. The
demonstrated defect is under-coverage — findings that should have been raised and
were not — which is the kind that accumulates silently, because a blind check and a
healthy estate produce the same empty report.

We confirmed the frozen list is exactly six. Several other daily checks looked
suspicious but turned out to run SQL against a stock Postgres image, so they carry
no build of ours and belong to a different staleness problem.

## Where we are now

One decision is open, and it is the owner's. It can be answered two ways at once,
because the six services divide cleanly into two groups with the same shape and very
different consequences.

The four checks are silently under-reporting on the estate's own immune system, and
that compounds every day nobody acts. The recommendation is to fold them into the
fleet release so they can never be more than one release behind, and then to add a
gate that fails when a service's pinned image predates the last change to the code
it is built from — the first fixes today, the second means a check added next month
is covered without anyone having to remember.

The two GitHub runners are different in kind. That image is Ubuntu plus a pinned
upstream tarball, project content changing about twice a year, so rebuilding it every
release churns it for nothing. Their separate cadence is defensible on its merits.
What is not defensible is that the exemption currently exists only as an absence —
no entry anywhere, indistinguishable from an oversight, which is how one of them
reached an image missing `rsync` and `ssh`. The recommendation is to unstick them
once, record the exemption where the tooling can see it, and give the second runner
the retag target it has never had.

## Where we're going

Once the decision lands: build and repoint the four checks, then re-run the registry
census as the acceptance test — it should read 169 of 169, and it is a test that can
fail. Then fold them into the release lists, verifying at `deploy-agents` and
`deploy-core` rather than at `release`, which since a separate change no longer
reveals anything under `make -n`. Then the runners and their written exemption. The
content-change trigger comes last, as its own change with its own review. One open
question stays open and does not block: whether the render check was functionally
affected by the function that moved under it.

`bugs_open/237` closes when all six are ruled and implemented — not when the gate is
green, because a green gate on a compliant tree could only ever have passed.
