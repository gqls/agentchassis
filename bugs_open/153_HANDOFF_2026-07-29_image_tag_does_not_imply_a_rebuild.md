# BUG 153 — an IMAGE_TAG bump does not imply a rebuild: nothing ties a tag to the code it was built from

> # ✅ FIXED AND LIVE — `v1.0.1283`, 2026-08-10 21:43Z. **14 of 14 backend services now say what commit built them.**
>
> **The close condition in § "How to verify a fix" is met.** Both `agent-chassis` replicas, and
> every other backend service, carry the sha of the commit their image was built from. Before
> this roll the count was **0** on every service, for the whole life of the platform.
>
> ```
> stamp on the live fleet:  d3c09cc746e563b6339831cfb69576eb52135c43
> resolves to:              commit d3c09cc, 2026-08-10 20:40:17 +0100
> that tree contains:       pkg/buildinfo ✓ · 14/14 dockerfiles stamped ✓ · 14/14 mains importing ✓
> ```
>
> **Controls, on `agent-chassis`** (the discriminating set — a positive alone proves nothing):
>
> | probe | result | meaning |
> |---|---|---|
> | the real sha | **3** | stamped |
> | a fabricated sha | **0** | the grep is not matching everything |
> | **a real but DIFFERENT commit** (current HEAD) | **0** | **the stamp is specific, not "some sha"** |
> | `orchestration` (positive control) | 8562 | the probe works at all |
>
> **Second, independent witness** — the startup log, which agrees with the binary:
> `{"msg":"build provenance","git_commit":"d3c09cc746e563b6339831cfb69576eb52135c43"}`
>
> **All 14 services verified** `[MEASURED]`: agent-chassis (both replicas), auth-service,
> core-manager, reasoning-agent, web-search-adapter, web-scrape-adapter, git-adapter,
> image-generator-adapter, thunder-adapter, analyser-adapter, browser-runner-adapter,
> content-creator-agent, remote-job-spawner, kafka-scheduler — same sha on every one.
>
> > **⚠ THE PROBE THAT VERIFIES THIS IS ITSELF A TRAP, AND IT FOOLED ME FIRST.** Two services
> > initially read as unstamped. Both were false:
> > - **`git-adapter`** — `ls /root/git-adapter` fails on *permissions* as `appuser`, which
> >   reads as "no binary". Use `/proc/1/exe`.
> > - **`browser-runner-adapter`** — the fleet's only **debian-slim** image ships no binutils,
> >   so **`strings` is command-not-found**, and the `2>/dev/null` in every published pod-grep
> >   recipe turns that into a silent `0`. I briefly believed this mechanism had caught a stale
> >   image on its first run. It had not; my probe had failed. Caught by the pod's `imageID`
> >   digest **matching** the local image's `RepoDigest`, which contradicted the conclusion.
> >
> > **So verify with `grep -aq "<expected-sha>" /proc/1/exe`** — no binutils, any image base,
> > any binary path — and **never with a discovery grep**: without `strings`' line boundaries
> > there is nothing to anchor, and `grep -aoE "[0-9a-f]{40}"` returns Go's internal digit
> > table (`000102030405…`) identically on every service. **Ask "does this pod carry sha X",
> > never "what sha does this pod carry".** Both traps are in `LANDMINES.md`; the near-miss is
> > in `WRONG_CALLS.md`. This is why CLAUDE.md's own `strings`-based recipe is now unsafe on
> > debian-based services fleet-wide.
>
> **What is still genuinely owed** (so this is not over-reported):
> 1. **The induced-fault test (RUNBOOK R6)** — bump `IMAGE_TAG`, push+deploy *without* build,
>    confirm the pod reports the OLD sha under the NEW tag. Everything above proves the
>    mechanism works on an **honest** roll; only R6 proves it catches a **dishonest** one,
>    which is the actual defect. Needs an owner cycle.
> 2. **The two CronJob images** (`component-render-check`, `shared-output-fields-check`) carry
>    the OCI label but their binaries are unstamped by design — labelled ≠ stamped.
> 3. **Candidates 2 and 3 remain unbuilt on purpose.** This **detects** a retag; it does not
>    refuse one.
>
> **File stays in `bugs_open/` per the owner ruling of 2026-08-06** (fixed bugs stay), and
> because item 1 above is a real outstanding proof, not bookkeeping.
>
> ---
>
> ## UPDATE 2026-08-11 — second roll (`v1.0.1284`) holds, and the stamp found `bugs_open/249` on its first outing
>
> **14/14 again on a release this lane did not run.** Every backend service came up printing
> `build provenance` on `v1.0.1284`, so the mechanism is the fleet's normal behaviour rather
> than an artefact of the roll it shipped in `[MEASURED]`, from the startup log line on each
> service plus `grep -aq <sha> /proc/1/exe` on a sample with per-pod controls.
>
> **What it revealed immediately: one tag, THREE source commits.** `v1.0.1284` was built over
> 6m22s; two other sessions committed inside that window; each service's build resolves `HEAD`
> independently (`makefile:128`), so five services carry `55fc8fc35`, one carries `e2afedaaf`,
> and eight carry `a41dec8e5`. Today's drift is docs + two `_test.go` files, so nothing
> functional differs — luck, not design. Filed as **`bugs_open/249`** with the timeline, the
> cause and ranked fix candidates. **This bug is why 249 is visible at all**; before `v1.0.1283`
> the question could not be asked.
>
> **Caveat now owed on this bug's own headline claim.** `git merge-base --is-ancestor <commit>
> <stamp>` answers "did my fix ship?" exactly — **for the service whose stamp you read**. It is
> not a fleet-wide answer while 249 is open, because a release can straddle a commit. Anyone
> quoting the chassis stamp for the fleet needs that qualifier.
>
> **Item 1 above (R6) is unchanged and still owed** — 249 is a *different* fault (skew inside an
> honest release), not the dishonest-roll fault R6 tests. But see the lane RUNBOOK: the local
> `REF=<older-commit>` regression guard covers most of R6's value without touching production.
>
> ---
>
> ## STATUS 2026-08-10 (earlier) — **OWNED, FIX BUILT AND COMMITTED, NOT YET LIVE.** Candidates 1+4 done; 2+3 deferred on purpose
>
> Lane: `docs/agent_docs/docs024_key_docs_latest/bugfix_153_build_provenance/` (standing five).
> Register: **BLD-019** in `docs026_concept_register/register/build-pipeline.md`.
> Commits: `e743e6cfc` (pkg), `e5f31dcdb` (14 mains), `1054ec36c` (14 dockerfiles + makefile).
>
> **The defect was re-verified before anything was written**, on live pods
> `agent-chassis-8496665bb8-{f6svp,sskxd}` at `v1.0.1279` — version-string grep **0**, 40-hex
> sha grep **0**, `grep -c ldflags` over the makefile and every backend dockerfile **0**.
> Unchanged 12 days after filing and 77 tag versions on. `[MEASURED]`
>
> **What is now built** (candidate 1 exactly as this file recommended, plus candidate 4):
> `ref_build`/`tree_build` pass the sha they already computed as `--build-arg GIT_COMMIT` and
> as OCI `image.revision`/`image.created` labels; all 14 backend dockerfiles stamp it through
> `-ldflags` into a new `pkg/buildinfo.GitCommit`; all 14 mains log it at startup;
> `verify-agent-images` gained three read-only stanzas that print the image label and the sha
> actually baked into the running pod's binary. **Full 40-hex, not `--short`** — this file's
> own positive control asks for `git rev-parse HEAD` exactly, and 40-hex is a measured-zero
> pattern in the binary so extraction is unambiguous.
>
> **Mechanism proven before commit, with a negative control**, against a clean
> `git archive HEAD` + the new files: WITH the ldflags clause the injected sha appears **3**
> times in `cmd/agent-chassis`; the same source WITHOUT it gives **0** for that sha and **1**
> for the literal `unknown`. Repeated on the two bare-`main.go` file-builds (`git-adapter`,
> `remote-job-spawner`) and `cmd/scheduler`: 3 each. `[MEASURED]`
>
> **STILL OPEN, and these are the honest reasons — none of them is bookkeeping:**
> 1. **Nothing is live.** No roll has happened. Releases here are whole-fleet and owner-run
>    (`make release redeploy-agents`), so a single-service roll is not this lane's to do —
>    see RUNBOOK R4. Until then the defect remains fully reproducible, which is the
>    `/bugs_closed/` bar.
> 2. ~~**The council round DIED UNJUDGED.**~~ **SUPERSEDED — the owner added credit at
>    ~18:12Z and the round was resubmitted and JUDGED. Verdict: REJECTED, hard veto from
>    `guardian`, on SCOPE. See the dedicated section below — it is an owner decision, not a
>    resubmit.**
> 3. **13 of 14 services are inert.** Their dockerfile+main.go edits are committed but their
>    binaries are unstamped until each is next rebuilt. Expect a MIXED fleet, and do not read
>    an unstamped adapter as a failed fix. Checklist below.
> 4. **The induced-fault test is unrun** (RUNBOOK R6) — bump the tag, push+deploy *without*
>    build, confirm the pod reports the OLD sha under the NEW tag. R5 alone proves the
>    mechanism works on an *honest* roll; only R6 proves it catches a dishonest one, which is
>    the actual bug. **A green R5 is not a close condition.**
>
> **Candidates 2 and 3 are deliberately NOT built**, and this is a decision, not an omission:
> both change the push/deploy contract fleet-wide and want explicit owner sign-off. Candidate
> 1 is their prerequisite — the label a refusal would compare against did not exist until now.
> So what is shipped **detects** a retag; it does not refuse one. Say so when reporting this
> fixed.
>
> ---
>
> ### COUNCIL VERDICT, round 1 on corr `44fa6a98-acaa-46b5-9ada-f0c34ca5475d`: **REJECTED — hard veto from `guardian`, on SCOPE not soundness**
>
> **The seats disagreed, which is the part that decides what happens next.**
>
> | seat | verdict | substance |
> |---|---|---|
> | `bug_historian` | **approve** | "does not touch any rebuild/rerender/regeneration/template-render code path… additive tooling, inert until read, consistent with the owner ruling cited" |
> | `reuse_agent` | **approve** | "a genuinely novel gap per the diagnosis's own evidence… edit 6 correctly extends the existing `verify-agent-images` target rather than creating a parallel mechanism — good reuse discipline" |
> | `editquality` | object (medium) | **a real bug, now fixed — see below** |
> | `guardian` | **VETO (high)** | scope: "this round bundles the shared makefile macro change with edits to all 14 Dockerfiles and all 14 mains in one plan, which is precisely the 'MANY packages at once' trigger" |
> | 8 others | abstained | (relevance-gated) |
>
> **The guardian was explicit that the mechanism is sound:** *"The mechanism itself (stamp git
> sha via ldflags + OCI label, prove with pos/neg control) is sound and well-evidenced — that
> part I'd approve on a single-service pilot. My veto is about scope, not soundness."* Its
> **contained alternative**: land `pkg/buildinfo` + the macro change + agent-chassis only, and
> track the remaining 13 Dockerfile+main edits as a named per-service follow-up.
>
> **`editquality`'s objection was correct and is FIXED (`8d270c68a`).** It is worth reading
> because it is this bug's own disease: my new pod-provenance check ended in
> `... | grep | head -1`, and a pipeline's exit status is its *last* command's — so `head`
> returned 0 even when grep matched nothing, and the `|| echo "no provenance stamp"` fallback
> **could never fire for an unstamped binary**, printing a silent blank instead. A check that
> reads clean in exactly the case it exists to catch. Fixed by capturing to a variable so the
> test is the last command, and **proven in all four directions** (old form + no match → exit
> 0, reproducing the bug; new form + no match → exit 1; new form + sha → 0 and prints; new
> form + `-tree` → 0 and prints).
>
> **The guardian's factual claim was "correctness of 13 of the 14 edits is unverified by this
> submission". That was true, and it is now measured.** Built all 14 services from committed
> HEAD with an injected sha: **14 of 14 carry the stamp, 3 occurrences each** — including the
> two bare-`main.go` file-builds (`git-adapter`, `remote-job-spawner`) and `cmd/scheduler`,
> the three structurally-unlike cases. Negative control on a sample of four: **sha count 0**
> without the flag. `[MEASURED]` 2026-08-10.
>
> > **Honest detail from that negative control:** the `unknown` default is *not* a standalone
> > `strings` line in most binaries (it is in `agent-chassis`, not in `auth-service`,
> > `git-adapter`, `kafka-scheduler`). So **"unstamped" must be tested as *absence of a sha*,
> > never as *presence of `unknown`*.** The shipped check does the former. Do not "improve" it
> > into the latter.
>
> **WHAT HAPPENS NEXT IS AN OWNER DECISION, AND DELIBERATELY NOT A RESUBMISSION.** CLAUDE.md's
> owner ruling of 2026-07-28 is directly on point: *"A veto on SCOPE is not answered by
> resubmitting with better measurements. It is a judgement about how a capability reached
> production. Record it where the change lives, route the seam to architecture review on its
> own merits, and let a human break it — especially when seats disagree with each other."*
> They did disagree, twice over. So this section **is** that record, and the measurements
> above are published as evidence for the human, **not** fired back at the gate.
>
> ### ⚖️ OWNER RULING, 2026-08-10 (same evening): **OPTION 1 — THE CODE STANDS AS COMMITTED.** All 14 services, one round.
>
> Put to the owner with the three options below and their costs, per the 2026-07-28 ruling that
> a scope veto is a human's call. **The owner chose option 1: let it stand.** No revert, no
> re-slicing, no architecture RFC. The whole-fleet release was authorised in the same breath.
>
> **This is the SECOND time this exact call has gone this way.** `bugs_closed/124` drew a
> REJECTED verdict from the guardian on the same ground — a shared seam arriving inside a
> larger change — and the owner's ruling then was *"the code stays and the precedent gets
> fixed."* Same shape, same outcome, eleven days apart.
>
> **So the generalisable observation, recorded but NOT acted on** (the owner picked option 1,
> not option 3, and filing an RFC anyway would be exactly the "answer a scope veto with more
> argument" move the ruling forbids): the guardian's *"MANY packages at once"* trigger fires on
> **edit count and file spread**, and appears not to distinguish a change that is *mechanically
> identical, provably inert, and verified across every instance* from one that is genuinely
> N independent judgements. Two of the eleven seats read this change as approvable; the
> guardian read it as a veto; the owner has now twice sided with the former. **If a third case
> lands the same way, that is a rate, not a coincidence, and the trigger's calibration is worth
> an RFC on its own merits** — raised by whoever hits it third, with three data points instead
> of two. Do not open it on the strength of this entry alone.
>
> What this ruling does **not** license: reading the change as council-APPROVED. It was not.
> The commits carry `Council-Submitted:`, which is accurate; **no `Council-Reviewed:` trailer
> exists on any of them and none may be added**, because the verdict was REJECTED and writing
> that trailer would be the coverage report's MISMATCH — its dishonesty surface. The honest
> summary is: *reviewed, vetoed on scope, overruled by the owner, and recorded.*
>
> ---
>
> The three options, costed, as they were put to the owner (**option 1 chosen**):
> 1. **Let it stand as committed** (14 services, one round). Cheapest; the guardian's specific
>    risk — "a missed import silently no-ops per-service" — is now measured away for all 14,
>    and every edit is inert until each service is next rebuilt. Accepts a scope precedent the
>    guardian objected to.
> 2. **Honour the contained alternative retroactively**: forward-commit a revert of the 13
>    non-pilot Dockerfile+main edits, roll agent-chassis alone, then re-land the 13 as a named
>    follow-up. Most faithful to the verdict; costs a revert plus 13 services' worth of
>    re-review for edits already proven correct.
> 3. **Route to architecture review** (`architecture_review/`) as a seam question — *"may a
>    mechanical, inert, N-service edit ride one round when the shared macro rides with it?"* —
>    which is the generalisable version and the one the 124 precedent points at.
>
> **Precedent worth knowing:** `bugs_closed/124` shipped a platform seam inside a bug patch
> and drew a REJECTED verdict on exactly this ground. The owner's ruling then was *"the code
> stays and the precedent gets fixed."* That is a reason to expect option 1 or 3, not a reason
> to assume it.
>
> ---
>
> **Per-service liveness checklist** — tick when that service's *binary* greps its build sha:
> `agent-chassis` ☐ (pilot) · `auth-service` ☐ · `core-manager` ☐ · `reasoning-agent` ☐ ·
> `web-search-adapter` ☐ · `web-scrape-adapter` ☐ · `git-adapter` ☐ ·
> `image-generator-adapter` ☐ · `thunder-adapter` ☐ · `analyser-adapter` ☐ ·
> `browser-runner-adapter` ☐ · `content-creator-agent` ☐ · `remote-job-spawner` ☐ ·
> `kafka-scheduler` ☐

**Filed:** 2026-07-29 18:30 BST · found incidentally while auditing the auto-memory
index (a banner claimed `v1.0.1192`; checking it against the live pods opened this).
**Status:** ~~OPEN, unowned~~ → **OPEN, owned by `bugfix_153_build_provenance` (2026-08-10)**;
see the status block above. **Not a code defect in the chassis** — a defect in the
build/deploy contract, so it bites every service and every session.

> ## ⚠ CORRECTED 2026-07-30 BY THE FILER — the original SYMPTOM is WITHDRAWN; the ROOT CAUSE stands
>
> This file was titled *"the fleet ran v1.0.1202 on a binary older than its own roll
> commit"* and §"Evidence" below still argues it. **That claim is withdrawn.** Read
> §CONTRIBUTION (bugsearch-7, the 144 lane) — it is correct, and I have re-verified
> every part of it:
>
> - **All five ADD-markers in the evidence table are unfindable in any binary.** Four
>   are prose from the 138/104 workstreams' own README/RUNBOOK; the fifth
>   (`a Degraded object always gates`) is a Go **comment** at
>   `diagnose_council_decide_action.go:709`. I harvested them by regexing quoted strings
>   out of `git show <commit>`, which returns doc changes alongside code. **The zeros
>   measured nothing**, so the 47-minute gap is unsupported.
> - **My "the string is untouched at `workflow.go:328`" was also wrong.** `583f31eae`
>   *did* remove the `fmt.Printf` form; the phrase survives as a **prefix** of the
>   replacement `"Checking disconnected step for cycles"`. Same conclusion (the marker
>   cannot discriminate), different and more reusable mechanism — substring containment,
>   not an untouched line. Note `git log -S` on the phrase is blind to this too, because
>   the occurrence count never changes.
> - **Four positive controls passed and could not see any of it.** A positive control
>   validates the *instrument*, not the *probe*. The tell I missed: a zero that does not
>   move when the image does. Confirmed on `v1.0.1207` — the same markers still read 0.
>
> **What is NOT affected, because it was measured directly rather than inferred from
> markers** (re-verified on `v1.0.1207`, 2026-07-30): the image carries no provenance —
> **0 shas, 0 version strings** in the binary, no `ldflags`/`ARG`/`LABEL` in the
> dockerfile — while `ref_build` computes the sha at `makefile:119` and discards it; and
> `verify-agent-images` (`makefile:1937`) checks tag *consistency*, never tag↔code. The
> fix candidates are unchanged and candidate 1 is now argued for by two independent
> marker failures in one day (see `WRONG_CALLS.md` 2026-07-29).
>
> Logged fleet-wide by the 144 lane in `WRONG_CALLS.md`; not duplicated here.

**Affects:** anyone verifying "did my fix ship?" — the reason this stayed invisible is
that the only available verification method is per-fix marker hunting, which is itself
error-prone enough to have produced two false readings in a single day.

> **[SUPERSEDED 2026-07-30]** ~~and — right now — at least three lanes (104, 138, 144)
> whose notes say "INERT until a roll" when the roll has already happened and delivered
> none of them.~~ Withdrawn with the symptom above. The memory index records that
> v1.0.1205/1206 (built 21:35+ BST 07-29 from HEAD) carry those three commits.

---

## Symptom

The fleet reports `v1.0.1202` consistently in every place we look — deployment, pods,
both replicas. The binary inside those pods **predates the roll commit that named it
by ≥47 minutes**, and does not contain three fixes that were committed *before the
pods started*.

Nothing anywhere reports an error. Every consistency check we own passes.

## Evidence — ⚠ WITHDRAWN 2026-07-30, retained only as the worked example of a broken probe

**Do not cite the marker table below.** All five markers are unfindable in any binary
(see the correction banner at the top). It is kept, not deleted, because the *shape* of
the mistake is the reusable part: every number in it is real, every control passed, and
the conclusion was still false.

The original framing was: *"All of it is string presence in the running binary, not
timestamp arithmetic"* — which was true and beside the point. Being immune to the
BST/UTC trap does not make a probe valid; I had guarded the failure mode I knew about
(§"Why this is NOT the trap that corrected 066", still sound) and never checked the
markers were code at all.

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

## CONTRIBUTION 2026-08-17 (`bugfix_277_required_fields_repair` lane) — the trap fired again, at scale, and BLD-019 is the only reason anyone noticed

**A roll happened this afternoon and shipped nothing.** The owner rebuilt and deployed the chassis
at **the same `IMAGE_TAG` as the morning's build** (`v1.0.1305`), so the node served its cached
layer. The pods restarted, looked healthy, and are running **this morning's binary**. Filed here
rather than as a new bug because this is exactly candidate 2's territory: the tag did not imply
the build.

### Three independent instruments, all disconfirmable, all agreeing

| instrument | reading |
|---|---|
| OCI label on the **local image** at `v1.0.1305` | `revision=89a0cbeb7…`, `created=2026-08-17T14:30:02Z` — a genuinely new build |
| **running binary** (`grep -a … /proc/1/exe`) | `6a782274b…` **PRESENT** (the morning's build) · `89a0cbeb7…` **ABSENT** |
| pod `imageID` vs local repo digest | pods on `sha256:f90a7e88…`; local image `sha256:6039e19c…` — **different images** |

Pods restarted `14:42:48Z` / `14:43:11Z`, i.e. after the 14:30 build — so this is not a timing
race. The rebuild simply never reached the node.

> **Control discipline, because a wrong reading here looks exactly like a right one.** The negative
> control was a **real but different commit sha**, never a constant-character run — a
> 40-zeros needle is present in every binary (LANDMINES, 2026-08-17) and would have made a sound
> probe look broken. Both arms behaved: the expected sha present, the other absent, same `exec`.

### The cost, quantified

**252 commits** sit between the running revision (`6a782274b`) and HEAD, of which **26 touch
`platform/`, `internal/`, `pkg/` or `cmd/`** — i.e. 26 commits of Go changes that their authors
have every reason to believe are live and are not. Among them are `remit.go`'s doc unification,
`283`/RFC_034's deterministic `scope_component_instance` half, and the `284` tie-break unification
of the agent-registration predicate.

**This is the failure mode CLAUDE.md's "bump `IMAGE_TAG` for every build" line exists to prevent,
and a written rule did not prevent it** — the same lesson as the owner's 2026-08-02 §2 ruling
(*"a comment is not a control on a tree this many sessions share"*). Candidate 2 or 3 is what turns
the rule into a control.

### What this says about the candidates

- **Candidate 1 is doing its job and is the only reason this was caught.** Nothing else in the
  estate would have distinguished "fresh pods running new code" from "fresh pods running the
  cached old code" — both look identical at `kubectl get pods`, at the tag, and in a health check.
  BLD-019's own register entry says it *"detects; it does not refuse"*, and today is that sentence
  happening in production.
- **Candidate 2 (make the tag imply the build) would have prevented it outright**, at either
  strength: a sha-suffixed tag makes a same-tag rebuild unrepresentable, and a `push-*` refusal on
  a revision-label mismatch would have failed the push instead of silently shipping nothing.
- **Candidate 3 (gate push on a `.build/<service>.<tag>` stamp) would NOT have caught this one.**
  Worth saying plainly so the ordering is honest: the stamp file *would* have recorded the new
  commit against the same tag and the push would have been permitted. Candidate 3 catches
  "pushed without building"; today's case is "built, pushed, and the node ignored it".
- Both 2 and 3 are recorded in this file and in register **BLD-019** as *deliberately not built,
  pending explicit owner sign-off* because they change the push/deploy contract fleet-wide. This
  instance is evidence for that decision, not a licence to take it.

### Immediate operational note

**Nothing is wrong with the code; it is simply not deployed.** To ship it: bump `IMAGE_TAG` in the
makefile (currently still `v1.0.1305`) and run a release — whole-fleet, owner-run, never a
single-service apply at its own tag. Until then, any session verifying "did my Go fix ship?"
against this chassis will correctly get **no**, and should not read that as their change being
missing from the build.

**Not applicable to config-shipped work.** `agent_definitions` / `scheduled_tasks` changes are live
at COMMIT and are unaffected — this lane's own work today (migrations `444`, `453`) is live and was
verified at the live rows, not at the binary.
