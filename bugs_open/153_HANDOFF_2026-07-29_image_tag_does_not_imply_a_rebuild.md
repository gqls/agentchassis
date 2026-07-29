# BUG 153 — an IMAGE_TAG bump does not imply a rebuild: the fleet ran v1.0.1202 on a binary older than its own roll commit

**Filed:** 2026-07-29 18:30 BST · found incidentally while auditing the auto-memory
index (a banner claimed `v1.0.1192`; checking it against the live pods opened this).
**Status:** OPEN, unowned. **Not a code defect in the chassis** — a defect in the
build/deploy contract, so it bites every service and every session.

**Affects:** anyone verifying "did my fix ship?", and — right now — at least three
lanes (`104`, `138`, `144`) whose notes say "INERT until a roll" when the roll has
already happened and delivered none of them.

---

## Symptom

The fleet reports `v1.0.1202` consistently in every place we look — deployment, pods,
both replicas. The binary inside those pods **predates the roll commit that named it
by ≥47 minutes**, and does not contain three fixes that were committed *before the
pods started*.

Nothing anywhere reports an error. Every consistency check we own passes.

## Evidence

All of it is **string presence in the running binary**, not timestamp arithmetic —
see the § "Why this is not the trap that corrected 066" below for why that matters.

Running image, both replicas (`kubectl -n ai-persona-system get pods -l app=agent-chassis`):

```
agent-chassis-cfd4d7cf7-bpdfb   docker.io/aqls/agent-chassis:v1.0.1202
agent-chassis-cfd4d7cf7-q9tzc   docker.io/aqls/agent-chassis:v1.0.1202
imageID (both): docker.io/aqls/agent-chassis@sha256:9590b3b7779d0d94d2e68ba92fa02c9de6fa812653480f6cde476906410d4e93
startTime:      2026-07-29T17:00:48Z  /  17:00:27Z          (= 18:00 BST)
```

Three commits, all made **before** those pods started, all **absent** from the binary:

| lane | commit | committed (BST / UTC) | marker grepped in `/app/agent-chassis` | count |
|---|---|---|---|---|
| 138 | `3a59b5012` | 17:14:54 / 16:14:54Z | `"blocked because architecture ran out of room"` | **0** |
| 138 | `3a59b5012` | ” | `"truncated review becomes a blocking review"` | **0** |
| 138 | `3a59b5012` | ” | `"a Degraded object always gates"` | **0** |
| 104 | `116fdffd8` | 17:24:22 / 16:24:22Z | `"the pattern has stopped matching"` | **0** |
| 104 | `116fdffd8` | ” | `"rigour over reassurance"` | **0** |
| 144 | `54fbfdf8b` | 17:41:11 / 16:41:11Z | `validation.WalkSteps` present in HEAD, absent in pod | — |

**Positive controls in the same exec** (so a zero means "absent", not "my grep is broken"):

```
orchestration                        7132
unknown execution-context field         1     (⇒ chassis ≥1191, bug 124's landmine satisfied)
Checking disconnected step              1     (⇒ 144's pre-fix code IS what is running)
They are CODE                           0     (⇒ the v1.0.1200 D12 guard did survive into this image)
```

The roll commit that named the tag:

```
8f26cf719  2026-07-29T18:02:13+01:00 (17:02:13Z)  chore(chassis): roll v1.0.1202 …
           ^ makefile IMAGE_TAG bump — committed 85s AFTER the pods had already started
```

**Therefore:** the image tagged `v1.0.1202` was built from code older than 16:14:54Z,
while the commit that declares "roll v1.0.1202" is timestamped 17:02:13Z — a gap of
**47m19s**, and everything committed inside that window is missing from production.

## Root cause — two gaps, only one of which is documented

**1. `push-*` / `deploy-*` are git-blind. This is BY DESIGN and is written down**
(`makefile:106-108`, repeated at `makefile:1068`):

> `# push-*/deploy-* are git-blind — they ship whatever is tagged $(IMAGE_TAG).`
> `# Provenance is got right HERE, at build time, and verified against the running`
> `# pod (never git, never the tag).`

That contract is sound and `ref_build` (`makefile:114-133`) honours its half properly:
it refuses a REF that is not a commit, prints how many uncommitted changes it is
leaving out, and `git archive`s into a clean context so no WIP can enter the image.

**2. But the image carries NO provenance, so the second half of that contract —
"verified against the running pod" — is unperformable.** Measured:

```
strings /app/agent-chassis | grep -cE "v1\.0\.1[0-9]{3}"   →  0
strings /app/agent-chassis | grep -cE "\b[0-9a-f]{40}\b"   →  0
```

`build/docker/backend/agent-chassis.dockerfile` has **no `ldflags`, no `ARG`, no
`LABEL`**. The binary cannot say what it was built from, and neither can the image.

⇒ **Bumping `IMAGE_TAG` and running `push-*`/`deploy-*` without re-running `build-*`
produces a retag of the previous binary, and nothing in the system can detect it.**
The tag is the only claim of provenance we have, and it is unbacked.

## Why the machinery we already own does not catch this

`verify-agent-images` (`makefile:1937`) compares the tag across the DB
(`agent_definitions`), the generic-orchestrator statefulset, running dynamic-agent
pods and the chassis deployment. It verifies that **the tag is consistent everywhere**
— not that the tag corresponds to any particular code. **On this exact defect it prints
all-green**, because every one of those places genuinely does say `v1.0.1202`.

That is worth stating plainly: the one verification target we have is structurally
blind to this failure, so "we checked" has been true and useless simultaneously.

## Why this is NOT the trap that corrected 066

`bugs_open/066` carries a correction warning that this machine is **BST**, `git log`
prints BST and `kubectl` prints UTC, so a naive comparison "makes a live fix look
un-shipped". That trap is real and it is the first thing to suspect here.

It does not apply: **no conclusion above rests on comparing a git time to a kubectl
time.** The finding is that five marker strings from three commits are *absent from
the binary* while four control strings in the same exec are present. Timestamps are
quoted for narrative only, and are shown in both zones. If every clock on this box
were wrong, the finding would be unchanged.

## Fix candidates — ordered by what makes the bad state UNREPRESENTABLE

**The makefile is the guide here (owner's direction), and it already does the hard
part**: `ref_build` computes `git rev-parse --short $(REF)` at `makefile:119` purely to
echo it. The commit is already in hand at build time; it is simply thrown away.

1. **Stamp the commit into the binary and the image; verify it at the pod.**
   Pass the sha `ref_build` already computes as `--build-arg GIT_COMMIT=` plus
   `--label org.opencontainers.image.revision=`, and add
   `-ldflags "-X main.GitCommit=$GIT_COMMIT"` to the dockerfile build.
   This turns the existing pod-grep discipline into an **exact, universal** check —
   `strings /app/agent-chassis | grep <sha>` answers "what is running?" for every
   service and every fix, and **retires per-fix marker hunting entirely**. (That
   hunting is itself producing defects: bug 144's stated marker
   `"Checking disconnected step"` → 0 is unachievable — the string is untouched at
   `platform/validation/workflow.go:328` — so that entry could never be closed by its
   own test.) Closes the door on *verification*.

2. **Make the tag imply the build** — the strongest, because it removes the bad state
   rather than detecting it. Either derive `IMAGE_TAG` as `v1.0.<n>-<shortsha>`, or
   have `push-<service>` refuse when the local image's `revision` label ≠ the ref
   being shipped (`FORCE=1` to override). A retag then cannot be produced by accident.

3. **Gate `push-*`/`deploy-*` on a build stamp.** `ref_build` writes
   `.build/<service>.<tag>` recording `(tag, commit, built_at)`; `push-<service>`
   refuses if it is missing or names a different commit. Cheapest of the three — no
   image-format change, no registry dependency — and it is the same "fail toward a
   wasted cycle, not a bad prod ship" direction the build macro already chose.

4. **Widen `verify-agent-images` to print provenance, not just the tag** — per-pod
   image `revision` label and `.CreatedAt` alongside the tag. Detects rather than
   prevents, so it is not a fix on its own, but it pairs with (1) and is the smallest
   useful change we could ship today.

**Do not** remove the `IMAGE_TAG` bump requirement while fixing this: a same-tag
rebuild ships the node's stale cached binary (CLAUDE.md, "Building & deploying
images"), so the bump is load-bearing for a different reason.

## How to verify a fix

- **Induced fault (the discriminating test):** bump `IMAGE_TAG`, then run `push-*` +
  `deploy-*` **without** `build-*`. Candidates 2/3 must refuse; candidate 1 must show
  the pod still reporting the *old* sha. If nothing objects, the fix is inert.
- **Positive control:** a full `make build-agent-chassis push-… deploy-…` cycle — the
  pod must report exactly `git rev-parse HEAD` of the ref built.
- **Regression guard:** `make build-<service> REF=<older-commit>` must stamp *that*
  commit, not HEAD.
- Verify at the **running pod**, never at git and never at the tag — the rule this bug
  exists to make performable.

## Immediate operational note

The makefile is already at **`IMAGE_TAG ?= v1.0.1203`** (`makefile:17`). Whoever rolls
it next should confirm the build ran **after 17:41 BST 2026-07-29**, or 104/138/144
will miss a second roll. Their index entries have been corrected to say they owe a
pod-grep rather than a status edit.

## Landmines

- **`verify-agent-images` will say all-green.** It cannot see this defect.
- **BST vs UTC**: `git log` prints BST, `kubectl` prints UTC (the trap that corrected
  066). Prefer string presence in the binary over any timestamp comparison.
- **A 0 from a marker grep is ambiguous** without a positive control in the same exec:
  it can mean "fix absent", "image older than the commit", or "my marker was never a
  real string". Always grep a control.
- **`imageID` alone does not prove a retag** — two tags sharing one digest does, but
  you need the previous tag's digest to compare, and we did not capture 1201's.

---

## CONTRIBUTION 2026-07-29 (bugsearch-7, the 144 lane) — the marker table cannot support the conclusion, and one of its markers is mine

**Read this before acting on §"Evidence".** I am not disputing the finding; I am
disputing the binary evidence for it, having just measured the same markers on the
NEXT image. The timeline argument (the roll commit timestamped 85s *after* the pods
started) is independent of everything below and is untouched by it.

### 1. All five marker strings are structurally incapable of appearing in ANY binary

Measured 2026-07-29 against the repo, not against an image:

| marker | where it actually lives |
|---|---|
| `blocked because architecture ran out of room` | **no `.go` file, ever** — `git log -S … -- '*.go'` is empty. It is a phrase from `bugfix_138_degraded_gates/README_where_we_are.md` |
| `truncated review becomes a blocking review` | **no `.go` file, ever** — it is that README's TITLE |
| `the pattern has stopped matching` | **no `.go` file, ever** — prose in `bugfix_104_fleetwide_claim_patterns/RUNBOOK` line 312 |
| `rigour over reassurance` | **no `.go` file, ever** — site copy, quoted in `bugs_open/147` |
| `a Degraded object always gates` | a **Go COMMENT**, `diagnose_council_decide_action.go:709`. Comments are not compiled in |

So all five grep to 0 against every image ever built, and every image that ever will be
built. They were harvested from the workstreams' *documentation* rather than from their
code. A zero here is not evidence about the image.

### 2. The 144 row's marker is also blind, and the "positive control" is blind in the opposite direction

`validation.WalkSteps` greps **0** on v1.0.1203 — an image I have separately proven
contains `WalkSteps` (its string literals are present on both replicas). A Go symbol
name is not a reliable `strings` target; a **string literal the code emits** is.

And `Checking disconnected step` → 1 was read as *"⇒ 144's pre-fix code IS what is
running"*. It cannot mean that. **That marker is my error, from my own bug file** — I
wrote it as a delete-marker, and it is not one: the replacement message is
`"Checking disconnected step for cycles"`, which **contains the old phrase as a
prefix**. It returns 1 with or without the fix. (Your §"Fix candidates" note spotted
that the marker was unachievable, and attributed it to the string being "untouched at
workflow.go:328" — the line WAS changed; the phrase survives inside the new string.
Same conclusion, different mechanism, and the mechanism is the reusable part.)
The discriminating form is `Checking disconnected step: ` — **with the colon and
space**, which only the deleted `fmt.Printf` format had.

### 3. What is actually true of v1.0.1203 (measured, both replicas)

144's fix — including round 2, `54fbfdf8b`, 16:41:11Z — **is live**:
`"uses fan_out, which cannot work inside a sub-workflow"` → 1, `"Substep declares
fields"` → 1, `"Checking disconnected step: "` → **0**, `"Checking disconnected step
for cycles"` → 1. Functionally: 22 orchestration runs carrying a `sub_workflow` since
the roll, 21 COMPLETED, 0 validation errors.

Since 138's and 104's commits are **earlier on the same branch** than `54fbfdf8b`,
v1.0.1203 necessarily contains them too. Their markers reading 0 is explained entirely
by §1, not by their absence.

### 4. What this does to the bug

- The **conclusion may still be right** — a tag bump that does not imply a rebuild is a
  real hazard, and the 85-second timeline is real evidence for it on v1.0.1202. Nothing
  above touches that.
- The **evidence table should be withdrawn or re-run** with real emitted string
  literals, or the fix candidates will be argued for on a table that proves nothing.
- **Fix candidate 1 gets stronger, not weaker.** Every failure above is a per-fix
  marker being hand-chosen wrongly, three times, by two sessions, in one day. Stamping
  the commit sha into the binary retires the entire practice — that is the argument,
  and it now has three worked examples instead of one.
- Suggested addition to the Landmines list: **a marker must be a string the binary
  EMITS.** Not a symbol name, not a comment, not a phrase from the workstream's own
  docs — and a "deleted" marker must be one the new code cannot contain as a
  substring. Cheapest check, before you exec anything:
  `git grep -c "<marker>" -- '*.go'` on the commit you expect to be running.
