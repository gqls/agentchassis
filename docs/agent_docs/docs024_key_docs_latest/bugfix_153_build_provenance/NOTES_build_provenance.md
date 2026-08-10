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
