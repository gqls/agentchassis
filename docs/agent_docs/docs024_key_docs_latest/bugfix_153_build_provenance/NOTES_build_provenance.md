# NOTES — bugfix_153_build_provenance (append-only, newest at bottom)

## 2026-08-10 — picking this bug up: ownership + validity check

Asked to find the next `bugs_open/` case not already owned by another thread, research it,
plan a robust framework-level fix (via a Fable-model planning pass), check the fleet's own
guidelines and the council, and see it through to a committed, buildable fix.

**Triage across all ~79 open bugs.** Ran `scripts/who-owns.py` against every number in
`bugs_open/`. Nearly all come back "OWNED or recently active" — this fleet is extremely busy,
and the script's heuristic fires on any commit that merely cites a bug number in its subject,
including the filing commit itself, so the raw verdict over-reports ownership. Cross-checked
against the actual `docs024_key_docs_latest/bugfix_*` workstream directory listing (grep for a
directory whose name embeds the bug's number) to separate "someone filed this and moved on"
from "someone is actively iterating on this". That narrowed the field to bugs with either no
matching `bugfix_NNN_*` directory, or a matching directory that has gone quiet (no commits in
14+ days).

**Considered and rejected `bugs_open/040` (kafka-dial timeouts).** Genuinely quiet — its
workstream's last commit was 2026-07-26/27, 14+ days ago. But it's an open-ended,
still-root-cause-unknown infra investigation awaiting more passive measurement days; not a
scoped fix a single session can responsibly plan and close.

**Considered and mostly-rejected `bugs_open/093` (stat-audit has one guarded call site).**
Root cause fully known, candidate 1 (the actual code fix) already shipped live in `v1.0.1172`
back on 07-27. Checked the DB directly: `site_work_items` now has 30 live `claims_unverified`
items (up from the "has never fired" state the file itself recorded as of 08-07), including a
same-day (08-10 08:44Z) revalidation of the exact vonc.com contradiction the file predicted —
so the detection gap is closed in substance, just not yet formally verified/documented/closed
out. What's left is process (council round 6 — confirmed via `diagnosis_artifacts`/`doc_notes`
that no round 6 exists yet) plus an optional, and per the file's *own* analysis NOT
recommended, prevention-path build (candidate 2). Real, unclaimed, small task — but not much
of a "fix" to plan, mostly paperwork. Kept in mind as a fallback, not picked.

**Picked `bugs_open/153`** — "an IMAGE_TAG bump does not imply a rebuild: nothing ties a tag
to the code it was built from". Self-declared `Status: OPEN, unowned` in its own header (the
strongest kind of signal — the filing session said so directly), no `bugfix_153_*` workstream
directory exists, `who-owns.py 153` shows the only touching commits are the filing session's
own (07-29/07-30), nothing since. Root cause is fully diagnosed and the file itself already
ranks fix candidates by robustness — exactly the "prefer a robust, framework-level solution"
brief. Chose it over 093 because there's a real, concrete, still-undone build to plan and do.

**Live-collision check.** `who-owns.py` is blind to uncommitted work in other sessions
(documented fleet-wide practice: grep the live `.jsonl` transcripts). There are ~40+ sessions
active in this repo right now (checked `~/.claude/projects/-home-ant-projects-agentchassis/
*.jsonl` mtimes — dozens modified in the last few minutes). Grepped recent transcripts for the
bug's filename and for `check_unverified_claims_stats`/`scanStoredStatClaims` (093's markers)
and for 153's own filename. All matches traced back to generic `ls bugs_open/` directory
listings (every session's own first-look step, mine included) or to my own session's prior
turns — no other session shows any sign of being inside `153`'s actual content or the
build/dockerfile/makefile files it touches. No `orchestration_states` row and no
`diagnosis_artifacts` row exists yet for anything resembling this fix.

**Re-verified the defect is still live today, not stale.** Both `agent-chassis` pods
(`agent-chassis-8496665bb8-{f6svp,sskxd}`, image `v1.0.1279` — 76 versions past the `v1.0.1202`
the bug was originally filed against) still show `strings /app/agent-chassis | grep -cE
'v1\.0\.1[0-9]{3}'` → 0 and the same for a 40-hex sha → 0. `grep -c ldflags
build/docker/backend/agent-chassis.dockerfile makefile` → 0/0. Nobody has fixed this in the 12
days since it was filed. `[MEASURED]` 2026-08-10.

**Plan drafted via a `claude-fable-5` Plan-type agent** (per the brief's "using fable, prepare
a plan" instruction), given the full research above plus my own direct reads of the makefile
(`ref_build`/`tree_build` defines, `push-backend`, `verify-agent-images`), all 14 backend
Dockerfiles' `go build` lines, `go.mod`'s module path, and `cmd/agent-chassis/main.go`'s
logger idiom. See `PLAN_2026-08-10_build_provenance.md`. Scope: candidate 1 (stamp + verify,
via the makefile's two `define`s so it's generic across all 14 services, not a one-off) plus
candidate 4 (widen `verify-agent-images`); candidates 2 (tag-implies-build) and 3
(push-gating) deliberately deferred — both change the push/deploy contract fleet-wide and
warrant explicit owner sign-off rather than one session's unilateral call.

## 2026-08-10 (later) — built, committed, and the two things that went wrong

**Built and committed in four commits**, each with an explicit pathspec: `e743e6cfc`
(`pkg/buildinfo`), `e5f31dcdb` (14 mains), `1054ec36c` (14 dockerfiles + makefile),
`041aae02b` (lane docs). Register entry **BLD-019** in `4451b2a0a`.

**The mechanism was proven before commit, with a negative control** — see RUNBOOK R2. The
positive alone would have been worthless: it proves the string exists, not that my flag put
it there. WITH ldflags → 3 occurrences; WITHOUT → 0 for that sha and 1 for `unknown`.

**Two same-file passengers, both declared in their commit messages.** Another session's
uncommitted `IMAGE_TAG` bump (`v1.0.1278`→`v1.0.1280`, and it moved from 1279 to 1280 *while
I was working*) rode along in `1054ec36c`; their uncommitted `BLD-018` rode along in the
register commit. A pathspec cannot exclude a same-file edit — the rule protects you from
other sessions' *other files*, not their edits to yours. Saying so in the message is the only
available remedy.

**A working-tree `go build` was failing for reasons that were not mine.**
`platform/orchestration/datahelpers/page_canonical.go:185: undefined: nestedOrFlatURL` —
another lane's half-written `BLD-018` work. Building from `git archive HEAD` + my own files
isolated my change and showed it clean. Worth internalising: on this tree, *a build failure is
not evidence about your change until you have isolated it*.

### MISSTEP 1 — I dismissed my own grep hits and filed a duplicate-ish bug

Filing `bugs_open/243` (the Anthropic usage-limit outage that killed my council round), I ran
the prescribed "grep both bug dirs before filing", **got three hits**, decided from the
filenames alone that they were coincidental, printed `--- (empty above = not filed) ---` under
a non-empty result, and filed as though it were a first occurrence. It is a **recurrence** —
2026-07-31, same signature, already in `LANDMINES.md`, with the resolution recorded in
`bugs_closed/130` (owner raised the cap the same day). Another lane had already logged today's
recurrence, with better evidence than mine.

Corrected in place with a visible banner, logged in `WRONG_CALLS.md`. **The transferable half:
a dismissed check is worse than a skipped one** — it leaves you confident. The tell was
sitting in my own terminal: a label that contradicted the output directly beneath it.

### MISSTEP 2 — I filed a landmine that duplicated one from 2026-07-31, for the same reason

I wrote *"a dead council run is indistinguishable from queue latency"* without grepping
`LANDMINES.md` first. An entry filed 07-31 already covered it, better. **Withdrawn** (with the
reason left in place so the deletion is legible), and its one novel point — that
`complete_invalid` is the council's generic "I could not run" state and generalises past any
one outage — folded into the surviving entry. Caught only because `landmines-sync.py --apply`
reported `content changed (same slug)` on an entry I had never touched. That is luck: I ran
the sync because I had appended a landmine, not because I was checking my premise.

### What is genuinely owed, in order

1. **The owner's whole-fleet release** (RUNBOOK R4). Not this lane's to run — a single-service
   roll fragments the fleet.
2. **Pod verification on every replica** (R5), then the **induced-fault test** (R6), which is
   the only one that proves the mechanism catches a *dishonest* roll. A green R5 is not a
   close condition.
3. **A fresh council submission** once the provider outage lifts (R7). The current correlation
   has no verdict and never will; `Council-Submitted:` is honest but uncredited.

## 2026-08-10 (evening) — the outage lifted, the round ran, and it came back REJECTED on scope

**The owner added credit; the fleet recovered at 18:12:11Z** (last failure 17:02:12Z), ~3h20m
after the 14:51:47Z cutover — 21 days before the API's stated auto-restore, i.e. by ACTION not
by calendar, which is precisely what `bugs_open/243` §6 said to record. Confirmed by sustained
traffic (43 successes across 3 agent types), not one lucky call.

**Resubmitted on the same correlation** (`RESUBMIT_CORR=44fa6a98…`) so the trail accumulates,
and watched it through eleven seats. **Verdict: REJECTED, hard veto from `guardian`.**

### The verdict is on SCOPE, and the seats disagreed

`bug_historian` **approve**, `reuse_agent` **approve**, `editquality` object (one real bug),
`guardian` **veto (high)**, 8 abstained. The guardian's own words: *"The mechanism itself…
is sound and well-evidenced — that part I'd approve on a single-service pilot. My veto is
about scope, not soundness."* Its objection: one round bundled the shared `ref_build`/
`tree_build` macro change with all 14 Dockerfile edits and all 14 main.go edits — the
"MANY packages at once" trigger.

**I did NOT resubmit, and that is deliberate.** CLAUDE.md's owner ruling of 2026-07-28 says a
scope veto *"is not answered by resubmitting with better measurements… record it where the
change lives, route the seam to architecture review on its own merits, and let a human break
it — especially when seats disagree with each other."* Recorded in the bug file's status block
(three costed options), the register entry, and the index row — all three now say NOT RATIFIED,
so no later session or seat reads BLD-019 as blessed.

### `editquality` found a real bug and it was this lane's own disease

My new pod-provenance stanza ended `... | grep -E ... | head -1` with a `|| echo "no provenance
stamp"` fallback. **A pipeline's exit status is its LAST command's**, so `head` returns 0 even
when grep matched nothing — the fallback **could never fire for an unstamped binary**, which
is the one case the check exists to catch. It printed a silent blank instead.

That is the same shape as bug 153 itself: *a check that reads clean in exactly the condition it
was built to report.* I wrote it while fixing that class. Fixed in `8d270c68a` by capturing to
a variable so the test is the last command, and **proven in all four directions** rather than
assumed — old form + no match → exit 0 (bug reproduced); new + no match → exit 1; new + sha →
0 and prints; new + `-tree` → 0 and prints.

### The guardian's factual claim, answered by measurement (published, not fired back)

*"Correctness of 13 of the 14 edits is unverified by this submission."* True as stated. Built
**all 14** from committed HEAD with an injected sha: **14/14 stamped, 3 occurrences each**,
including the three structurally-unlike cases (`git-adapter` and `remote-job-spawner` build a
bare `main.go` by filename; `kafka-scheduler` builds `./cmd/scheduler`). Negative control on
four: **0** without the flag.

> **A detail from that negative control worth keeping.** The `unknown` default is **not** a
> standalone `strings` line in most binaries — present in `agent-chassis`, absent in
> `auth-service`, `git-adapter`, `kafka-scheduler`. So **"unstamped" must be tested as the
> ABSENCE OF A SHA, never as the PRESENCE OF `unknown`**. The shipped check does the former.
> Had I built the check the other way — which is the more natural way to write it — it would
> have reported three services as stamped when they were not.

### Still owed

1. **The owner's decision** on the scope veto (the three options are in the bug file). This is
   the blocking item and it is not mine to take.
2. **The whole-fleet release**, then pod verification and the induced-fault test (RUNBOOK R4–R6).
3. Whatever the decision implies — a revert-and-re-land, an architecture RFC, or nothing.

## 2026-08-10 (night) — LIVE on v1.0.1283. 14/14. And the probe fooled me twice.

**The roll landed and the close condition is met.** Every backend service carries
`d3c09cc746e563b6339831cfb69576eb52135c43`. That sha resolves to a real commit whose tree
holds `pkg/buildinfo` + 14/14 stamped dockerfiles + 14/14 importing mains, and it contains the
editquality fix. Before this roll: **0 on every service, for the platform's entire life.**

Controls on `agent-chassis` (both replicas): real sha **3**, fabricated **0**, **a real but
different commit (current HEAD) 0**, positive control `orchestration` **8562**. That third row
is the load-bearing one — it proves the stamp is *specific*, not merely "a sha is present".
The startup log agrees independently:
`{"msg":"build provenance","git_commit":"d3c09cc746e5…"}`.

### The mechanism paid for itself the same evening, in another lane

Commit `ebaac39c0` (concept-register staleness survey) used the stamp to settle **19
roll-blocked register entries** — `git merge-base --is-ancestor <entry's commit> d3c09cc74`
answers exactly what previously needed a hand-picked marker hunt per entry. All 19 came back
IN. Their own caveat is the right one: **the stamp only settles an entry that NAMES a commit**,
and 13 of 29 surveyed entries name none.

### MISSTEP 3 — I nearly published my broken probe as the mechanism's first catch

13 services stamped, `browser-runner-adapter` apparently not. The satisfying conclusion —
*"it caught a stale binary on day one"* — was wrong. That service is the fleet's only
**debian-slim** image (Chromium needs glibc); debian-slim ships **no binutils**, so `strings`
is command-not-found, and the `2>/dev/null` every published pod-grep recipe carries turned that
into a clean `0`.

**What caught it was a contradiction, not diligence:** the pod's `imageID` digest *matched* the
local image's `RepoDigest`, so the image could not be stale. Extracting the binary from that
image showed the stamp present.

**And the fix for it was wrong in the same direction.** Replacing `strings` with
`grep -aoE "[0-9a-f]{40}"` to *discover* the sha also discards `strings`' line boundaries, so
`^`/`$` stop meaning anything and the match lands in Go's internal digit table — every service
returned `0001020304050607…`, identically, no error. Caught only because I tested the fix
against a known-good service instead of trusting it.

**Settled form:** `grep -aq "<expected-sha>" /proc/1/exe` — no binutils, any image base, any
binary path (which also retires the `/app/x` vs `/root/x` vs `/x` guessing, and `ls` on
`git-adapter` fails on *permissions*, a second false "not there"). **Verify a known value;
never discover one.** In `LANDMINES.md` and `WRONG_CALLS.md`.

> **Three confidently-wrong readings in one day, all the same shape**: the jsonb `provider`
> path (0 rows, wrong path), the dismissed grep hits (non-empty result read as empty), and this
> (`strings` absent). Each was an instrument answering a different question than the one asked,
> in the believable direction. The common defence is the one I applied to `agent-chassis` and
> then dropped as the list got long: **a control in the same breath as the measurement.**

### Council + owner

Round 1 REJECTED — guardian veto on **scope**, not soundness (`bug_historian` and `reuse_agent`
both approved). Put to the owner per the 2026-07-28 ruling; **owner chose option 1: the code
stands as committed.** Same call as `bugs_closed/124`, eleven days apart. **Live and proven ≠
approved** — no `Council-Reviewed:` trailer exists or may be added.

### Owed

1. **Induced-fault test (RUNBOOK R6)** — the only proof that covers a *dishonest* roll.
2. Two CronJob images labelled but unstamped by design.
3. Candidates 2/3 unbuilt on purpose — this detects, it does not refuse.

---

## 2026-08-11 — second roll (`v1.0.1284`): the mechanism holds, and immediately finds something

**Stamp survived the second release.** 14/14 backend services print `build provenance` at
startup on `v1.0.1284`. So the fix is not a one-roll artefact of the night it shipped — it is
now the fleet's normal behaviour, on a release nobody in this lane ran.

**And the first thing it found was `bugs_open/249`, unprompted.** `v1.0.1284` is one tag
carrying **three source commits**:

| built (UTC) | service | revision |
|---|---|---|
| 09:07:40 → 09:10:03 | auth-service, core-manager, agent-chassis, reasoning-agent, content-creator-agent | `55fc8fc35` |
| 09:10:22 | remote-job-spawner | `e2afedaaf` |
| 09:10:58 → 09:14:02 | kafka-scheduler + the 7 remaining adapters | `a41dec8e5` |

One `make release`, 6m22s, two commits from other sessions landing inside it. Cause is
`makefile:128` — `git rev-parse` lives **inside `ref_build`**, so with the default `REF=HEAD`
each of the 14 builds resolves HEAD independently. Full account, evidence and ranked fix
candidates: `bugs_open/249`.

Today's skew is docs + two `_test.go` files, so nothing functional differs. That is luck, not
design, and it is worth saying plainly rather than filing this as low-severity: the same six
minutes with a platform commit in them ships half a fleet with the change and half without,
under one tag, both halves reporting the same version.

**Caveat this lane now owes the estate.** The closing move we advertised — `git merge-base
--is-ancestor <commit> <stamp>` answers "did my fix ship?" — is **per service**, not per fleet.
It is still exact for the service whose stamp you read. Generalising one service's stamp to the
fleet is only safe for commits that predate the whole release window. (The 19 register entries
another lane settled in `ebaac39c0` are safe on that test — all predate `55fc8fc35` — but the
method needs the qualifier attached wherever it gets quoted.)

### Misstep, caught before it was asserted (the fourth of the same shape)

My first sweep across the 12 non-chassis services ran
`kubectl exec <pod> -- grep -aq "$SHA" /proc/1/exe 2>/dev/null` and returned **NO MATCH for
eight of them**. The tempting read — and I had the sentence half-written — was "the probe is
blind again, like `strings` was". The opposite was true: the probe was fine and the negatives
were real.

What settled it was **a control on the same pod in the same breath**, which is the defence this
lane keeps rediscovering: on `git-adapter`, `grep -aq a41dec8e5…` → MATCH, `grep -aq 55fc8fc35…`
→ no match. A probe that can find one 40-hex sha in that binary and not another is not blind.

**The transferable bit is that the same reading was wrong in the opposite direction each time.**
On 08-10 a blind instrument produced a false *positive* finding ("the mechanism caught a stale
binary!"); today a working instrument nearly produced a false *dismissal* ("my probe must be
broken again"). Yesterday's lesson — distrust a clean zero — is not a rule that says "zeros are
false". Both errors are the same failure to run a control, and only the control tells them
apart. Note also that `2>/dev/null` is what made the first sweep unreadable: it is in every
recipe we write, and it is exactly what erases the difference between "no" and "couldn't look".

The cheaper instrument was available the whole time and I reached for it second, not first:
`logs -l app=<service> | grep 'build provenance'` needs no exec, no binary path, no tool inside
the image, and it is the *service's own statement* rather than my inference from its bytes.
**For "what commit is this service running?", read the log line; keep the binary probe for the
case where you doubt the logs.**

---

## 2026-08-11 (later) — the owner's four decisions, and what each cost

Decisions taken: **(1)** pin the ref in `release`; **(2)** the local regression guard instead of
the production induced-fault test; **(3)** yes to rewriting CLAUDE.md's build section;
**(4)** builds and deploys stay manual — so no refusal mechanism in `push-*`/`deploy-*`, and
this lane does not run releases.

### The pin (BLD-020)

`release` and `release-backend` were prerequisite lists, which is *why* the defect existed:
prerequisites are made before the recipe runs, so nothing in the recipe can reach them. Both are
now recipes calling `$(call pinned_sweep,<goals>)` — one `git rev-parse`, echoed, then each goal
in order as `$(MAKE) <goal> REF=$$PINNED`.

**A cost, not a bug, and I would rather it be found here than by a reader who thinks it broke:**
`make -n release` now prints the sweep rather than the docker commands underneath it. The `+`
recipe prefix would restore the old preview — it forces the sub-makes to execute under `-n` and
relies on MAKEFLAGS carrying `-n` down. **Rejected deliberately.** The owner drives releases by
hand; a preview command that performs a real release if that assumption ever fails is a far worse
failure than untidy output is a gain. `make -n build-backend` is unchanged and is the real
preview. Written into the makefile comment, BLD-020's landmine, and RUNBOOK R9d.

### The regression guard — and what makes it evidence rather than decoration

`make build-agent-chassis REF=d3c09cc74… IMAGE_TAG=scratch-153-guard` (scratch tag, never
pushed, image and extracted binary removed):

| probe | result | what it rules out |
|---|---|---|
| OCI label | `d3c09cc74…` | the label follows `REF`, not HEAD |
| OLD sha in binary | present | the stamp got in |
| **live HEAD `6235beb44…`** | **absent** | **the stamp is THIS ref's, not "some sha"** |
| fabricated sha | absent | the grep is not matching everything |
| `orchestration` | present | the probe can read this binary at all |

**The load-bearing control is the HEAD one.** A fabricated sha only shows the grep is not
promiscuous; a *real but different* commit is what distinguishes "built from the ref I asked for"
from "contains a sha". Two of this lane's three false readings would have survived a fake-sha
control, so this is not a pedantic distinction — it is the one that was missing each time.

That guard also happens to prove the mechanism the pin depends on: `REF=` reaching the linker.
One command, two purposes, which is why it was worth running even though the pin is inert until
the owner's next release.

### The council gate REFUSED this one, client-side, and it is right to

```
REFUSED: no edit touches the review scope (platform/, internal/, pkg/ — owner ruling 2026-07-17).
```

A makefile-only change is out of scope; no credits were spent. `FORCE=1` exists and **was not
used** — spending council credits to override an owner's scope ruling is not a call this lane
gets to make, particularly while `bugs_open/244` has the gate at 87.8% of the fleet's August
spend. So this commit carries **neither trailer**, and that is the honest state: not reviewed,
not submitted, out of scope.

Worth noticing rather than acting on: the release path is plausibly the most shared mechanism on
the estate, and it sits outside the review scope, while a one-line change inside `pkg/` draws a
full round. **One lane is not a rate** — recorded here, not opened as an RFC, per the same rule
this lane applied to the guardian's calibration in the last round.

> Note for whoever reads `SUBMISSION_2026-08-11_release_ref_pin.json` later: it is a **written
> and refused** submission, not a judged one. Nothing in this lane's history corresponds to a
> verdict on it. Also, the schema takes `plan.risks` as a **prose string**, not an array — the
> array form fails validation with a clear message, which is how I found out.

### CLAUDE.md § "Building & deploying images", rewritten

The `strings /app/<svc> | grep -c "<your symbol>"` recipe is gone. It was unsafe on the
debian-slim image (no binutils → silent 0) and it is the marker-hunting practice that produced
three false readings in one day. Replaced with: ask the service (`logs … | grep 'build
provenance'`), then `git merge-base --is-ancestor` for "did my fix ship?", with the binary probe
demoted to "if you doubt the logs" and carrying its controls. The old claim *"nothing downstream
of the build records whether it came from a commit"* is struck through and dated rather than
deleted — it was true when written and is the reason the section looked the way it did.

**Also added there, because the correction would otherwise overclaim in the other direction:**
the stamp answers "did my fix ship?" **per service**, not per fleet, until BLD-020 has a release
under it.

---

## 2026-08-11 (evening) — `v1.0.1286`, the first release under the pin: COHERENT, but not yet PROVEN

**Result: one revision across all 14 backend services** — `c3b424c8e`, agreeing at both
instruments (image labels on the build machine; the services' own startup lines at the pods).
Build window 11:47:15Z → 11:53:07Z. Fleet is 18 deployments on `v1.0.1286`.

Chassis needed the fallback: `logs --tail=300` (and `--tail=3000`) does **not** reach its startup
line — it is a busy service and the line has scrolled. Binary probe with controls instead:
both replicas carry `c3b424c8e` and **do not** carry the previous release's sha. 13 by log line,
1 by probe, 14/14.

### ⚠ I nearly wrote this up as "the pin is proven". It is not, and the reason is the estate's own rule

**Zero commits landed inside that 5m52s window.** The nearest was `c3b424c8e` itself, 10 seconds
*before* the first build started — which is precisely the commit the release pinned to, not a
commit that arrived during it. So the old, unpinned code would have produced **exactly the same
result**: fourteen builds each resolving HEAD independently, with HEAD not moving.

One revision here is **consistent with the pin and equally consistent with its absence.** The
measurement could not have come out otherwise, which by this estate's rule makes it not evidence.
The honest word for `v1.0.1286` is **coherent**, not *proven*.

The contrast is the whole point: `v1.0.1284` ran 6m22s with two commits inside it and produced
three revisions; `v1.0.1286` ran 5m52s with zero inside and produced one. **The proof arrives
free on the first busy release** — and today's own numbers say that is likely soon, since the
busiest 7-minute window on 08-11 held 13 commits.

**RUNBOOK R9b now carries the missing half** (R9b(ii)): take the window from the image labels and
ask `git log` what landed inside it, *then* interpret the revision count. The check as I first
wrote it counted revisions and stopped — it would have called every quiet release a success,
including the quiet releases that happened before the pin existed. Watch the timezone: labels are
UTC, `git log %ad` is local.

### The pin survived another session's makefile edit

`a9237f0c9` (11:42, another lane's council round 2) added `build-removed-config-keys-check` to the
makefile. Checked rather than assumed: `pinned_sweep` is intact at HEAD and `release` still calls
it. Pure addition, no conflict — but this is the same file two lanes edited within 40 minutes,
which is exactly the same-file-passenger shape, and it is worth re-checking the pin after any
makefile commit rather than trusting that it is still there.

### A third labelled-but-unstamped image — BLD-019's landmine 3, exactly as predicted

That new target builds `build/docker/backend/removed-config-keys-check.dockerfile`, which compiles
`./cmd/config-key-audit` and contains **0** occurrences of `GIT_COMMIT` and **0** of `ldflags`
(positive control: `agent-chassis.dockerfile` returns 2, so the grep spelling is sound). It calls
`ref_build`, so its **image** gets the OCI `revision` label from the docker CLI while its
**binary** stays unstamped. Third instance, joining `component-render-check` and
`shared-output-fields-check`. **Do not read a labelled image as a stamped binary.**

Recorded in BLD-019's landmine 3, **not fixed** — that lane is mid-council-round on it (`a9237f0c9`
is its round-2 REVISE), and adding two lines to another session's dockerfile while they iterate is
how two lanes end up fighting over one file.

> **A misstep of my own, caught by its own output:** the first version of that check was
> `ls "$D" && grep -c GIT_COMMIT "$D" || echo "no dockerfile"`. `grep -c` printed `0` and
> **exited 1**, so the `||` fired and it printed *"no dockerfile"* under a file that plainly
> exists. A measured zero wearing the costume of a missing file — the same shape as the
> `strings` trap, in my own one-liner, one day after writing the landmine about it. The fix is
> the boring one: separate the existence test from the count, and print the exit status.

---

## 2026-08-11 (night) — `v1.0.1287`, the second release under the pin: **PROVEN**

Picked up this session on the `HANDOFF_2026-08-11_continue_here.md` pointer, whose entire owed
item was "run R9b + R9b(ii) after the next release". The fleet had already moved: `makefile:17`
reads `v1.0.1287` (was `v1.0.1286` when the handoff was written), and `kubectl get deploy` showed
all 14 backend deployments already running it. So the release had happened; only the grading was
owed.

**R9b — image labels, all 14 services, freshly pulled:**

```
2026-08-11T14:17:39Z auth-service             9b7811d4b9cd412b7e11633aef112348b8eda337
2026-08-11T14:18:06Z core-manager             9b7811d4b9cd412b7e11633aef112348b8eda337
2026-08-11T14:18:57Z agent-chassis            9b7811d4b9cd412b7e11633aef112348b8eda337
2026-08-11T14:19:46Z reasoning-agent          9b7811d4b9cd412b7e11633aef112348b8eda337
2026-08-11T14:20:04Z content-creator-agent    9b7811d4b9cd412b7e11633aef112348b8eda337
2026-08-11T14:20:28Z remote-job-spawner       9b7811d4b9cd412b7e11633aef112348b8eda337
2026-08-11T14:21:15Z kafka-scheduler          9b7811d4b9cd412b7e11633aef112348b8eda337
2026-08-11T14:21:32Z web-search-adapter       9b7811d4b9cd412b7e11633aef112348b8eda337
2026-08-11T14:21:50Z web-scrape-adapter       9b7811d4b9cd412b7e11633aef112348b8eda337
2026-08-11T14:22:18Z git-adapter              9b7811d4b9cd412b7e11633aef112348b8eda337
2026-08-11T14:22:43Z image-generator-adapter  9b7811d4b9cd412b7e11633aef112348b8eda337
2026-08-11T14:23:06Z thunder-adapter          9b7811d4b9cd412b7e11633aef112348b8eda337
2026-08-11T14:23:53Z analyser-adapter         9b7811d4b9cd412b7e11633aef112348b8eda337
2026-08-11T14:24:18Z browser-runner-adapter   9b7811d4b9cd412b7e11633aef112348b8eda337
```

One revision, 14/14. Build window `14:17:39Z`–`14:24:18Z`, 6m39s.

**R9b(ii) — what landed inside that window, converted from BST to UTC** (system clock is BST,
`git log`'s local dates are UTC+1, so subtract an hour):

```
9b7811d4b  2026-08-11T15:17:15+01:00 = 14:17:15Z   bug(253): BestLabelMatch scores by RAW …
d80fbf4bf  2026-08-11T15:17:44+01:00 = 14:17:44Z   notes(203): manual detection run halted …
```

`9b7811d4b` is exactly the pinned revision (full match, not just the short form) — it was `HEAD`
24 seconds *before* the first image's `created` timestamp, i.e. it is what `pinned_sweep` resolved
once at the start of the release, before any build ran. `d80fbf4bf` landed 5 seconds *after* the
window opened and nearly 6m34s before it closed, from another session, touching only a docs file
(`NOTES_phantom_cta_cleanup.md` — irrelevant to whether the mechanism held; what matters is that
`HEAD` moved mid-sweep). **This is the discriminating case.** Under the pre-fix behaviour, every
service built after `14:17:44Z` — 13 of the 14, everything after `auth-service` — would have run
`git rev-parse HEAD` independently and picked up `d80fbf4bf`, producing at least two revisions
exactly as `v1.0.1284` did (two commits inside a 6m22s window → three revisions). Instead all 14
still show `9b7811d4b`. The pin held through a real commit landing mid-sweep.

**Second instrument — the pods**, per-service startup log line:

```
agent-chassis              <none in range>            # scrolled past --tail=300, expected
auth-service               "git_commit":"9b7811d4b…"
core-manager                       "         …
reasoning-agent                    "         …
web-search-adapter                 "         …
web-scrape-adapter                 "         …
git-adapter                        "         …
image-generator-adapter            "         …
thunder-adapter                    "         …
analyser-adapter                   "         …
browser-runner-adapter             "         …
content-creator-agent              "         …
remote-job-spawner                 "         …
kafka-scheduler                    "         …
```

13/14 by log line, matching the label exactly. `agent-chassis` needed the fallback again — same
as `v1.0.1286` — so the binary probe, with its three controls, on both replicas:

```
agent-chassis-8657f4d748-62nt7   OWN SHA (9b7811d4b…): MATCH · FAKE sha: no match · orchestration: present
agent-chassis-8657f4d748-8lm4f   OWN SHA (9b7811d4b…): MATCH · FAKE sha: no match · orchestration: present
```

14/14 across two independent instruments. **`bugs_open/249` closes; BLD-020 moves from
*exercised* to *proven*.** Both updated in place (owner ruling 2026-08-06: closed bugs stay in
`bugs_open/`, the closure evidence goes in the file, the file does not move).

No misstep this round — the timezone conversion was the one thing the runbook already flagged in
advance (R9b's own gotcha, and R9b(ii) as written for `v1.0.1286`), and checking it against
`git show -s --format='%h %cI'` directly (rather than trusting `%ad` + a mental UTC offset) kept
it from becoming this session's own version of the same mistake.
