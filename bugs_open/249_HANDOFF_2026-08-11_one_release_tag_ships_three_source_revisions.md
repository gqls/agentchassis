# 249 — One `make release` ships THREE source revisions under ONE image tag

> ## STATUS 2026-08-11 evening — **FIX COMMITTED AND EXERCISED; the first release under it came out COHERENT but does NOT yet PROVE it**
>
> Candidate 1 (pin the ref once per release) is built, committed and registered as **BLD-020**
> (`21b9772a9`). `v1.0.1286` is the first release to run under it: **one revision, `c3b424c8e`,
> across 14/14 backend services**, agreeing at both instruments — image labels on the build
> machine and each service's own startup line at the pods (chassis needed the binary probe, with
> controls; its startup line had already scrolled past `--tail=3000`).
>
> **Why that is not the close condition.** [MEASURED] **zero commits landed inside that 5m52s
> build window** — the nearest was `c3b424c8e` itself, 10 seconds *before* the first build
> started, which is simply what the release pinned to. The old unpinned code would have produced
> the identical result, so the check could not have come out otherwise and is therefore not
> evidence. Contrast the defect itself: `v1.0.1284`, 6m22s, **two** commits inside, **three**
> revisions out.
>
> **This bug closes on the first release that straddles a commit and still ships one revision.**
> That costs nothing and should be soon — the busiest 7-minute window on 2026-08-11 held 13
> commits. Grade it with **RUNBOOK R9b(ii)** (lane dir), which asks what committed inside the
> window *before* counting revisions. R9b as first written counted revisions only, and would have
> called every quiet release a success — including the quiet ones from before the fix existed.
>
> Candidate 2 (the cross-service assertion) was offered to the owner and **not taken** — recorded
> as a decision in BLD-020, not left silent. Candidate 3 (runbook-only) is moot.
>
> **Not council-reviewed:** the gate refused it client-side — a makefile-only change touches none
> of `platform/`, `internal/`, `pkg/` (owner ruling 2026-07-17). No credits spent, `FORCE=1` not
> used, and no commit carries a trailer claiming otherwise.

**Filed:** 2026-08-11 · **Status:** OPEN, detected on the first release after the fix that made
it visible · **Found by:** `bugfix_153_build_provenance`, verifying the `v1.0.1284` roll

---

## 1. Symptom

`v1.0.1284` is a single tag. The fleet running it is built from **three different commits**,
spanning five commits of history. Before `bugs_open/153` shipped (the roll before this one,
`v1.0.1283`) this was **undetectable** — nothing in the pod could say what built it.

```
09:07:40Z  auth-service             55fc8fc35
09:08:06Z  core-manager             55fc8fc35
09:08:58Z  agent-chassis            55fc8fc35
09:09:44Z  reasoning-agent          55fc8fc35
09:10:03Z  content-creator-agent    55fc8fc35
                    ← e2afedaaf committed here (another session)
09:10:22Z  remote-job-spawner       e2afedaaf
                    ← a41dec8e5 committed here (another session)
09:10:58Z  kafka-scheduler          a41dec8e5
09:11:32Z  web-search-adapter       a41dec8e5
09:12:00Z  web-scrape-adapter       a41dec8e5
09:12:20Z  git-adapter              a41dec8e5
09:12:37Z  image-generator-adapter  a41dec8e5
09:12:57Z  thunder-adapter          a41dec8e5
09:13:45Z  analyser-adapter         a41dec8e5
09:14:02Z  browser-runner-adapter   a41dec8e5
```

The whole build took **6m22s**. Two commits from other sessions landed inside it. The cut points
match the commit times to within seconds — this is not a coincidence to be argued about, it is
the timeline read off two independent clocks.

## 2. Root cause — one line of the makefile, evaluated 14 times

`makefile:111` — `REF ?= HEAD`
`makefile:128` — `GIT_COMMIT=$$(git rev-parse '$(REF)^{commit}')`

That `git rev-parse` sits **inside `ref_build`**, which is `$(call)`-ed once per service. With
the default `REF=HEAD` each of the 14 builds resolves HEAD **independently, at the moment it
runs**. `release` (`makefile:2357`) is `build-backend push-backend deploy-core deploy-agents …`
— a sequential sweep with no ref pinned, so on a tree ~40 sessions commit to, a release straddles
whatever lands during those six minutes.

Nothing here is broken code. `ref_build` does exactly what it says: it builds committed state.
The unstated assumption is that "committed state" is *one* state for the duration of a release,
and on this tree it is not.

## 3. Why it matters (and why today's instance is harmless)

**Today: harmless.** The five commits in `55fc8fc35..a41dec8e5` are docs plus two `_test.go`
files. No production code differs between the three images. Verified:

```bash
git diff --name-only 55fc8fc35..HEAD | grep -E '\.(go|sql)$|^platform/|^internal/|^pkg/'
#  platform/orchestration/actions/diagnose_load_runtime_schema_test.go
#  platform/orchestration/actions/write_audit_findings_verifier_join_test.go
```

**The general case is not harmless.** A platform change committed at 09:10 would have reached
the eight adapters and *not* the chassis, auth-service, core-manager, reasoning-agent or
content-creator-agent — under one tag, with both halves reporting "deployed at v1.0.1284". Every
downstream statement of the form "the fix is live, it rolled at 1284" would be half true, and
the half that was false would look identical to the half that was true. This is `bugs_open/153`'s
own disease displaced one level up: **the tag stopped identifying the binary; now it turns out
the tag never identified the *revision* either.**

It also silently weakens the closing move `153` gave the estate — `git merge-base --is-ancestor
<commit> <stamp>` answers "did my fix ship?" **per service**, not per fleet. A lane that reads
the chassis stamp and generalises to the fleet will be wrong whenever a release straddles a
commit. One lane has already settled 19 register entries this way (`ebaac39c0`); those answers
happen to be safe (all predate 55fc8fc35) but the method needs the caveat.

## 4. Fix candidates, ranked by what makes the bad state unrepresentable

1. **`release` pins the ref once and passes it down.** Resolve `git rev-parse HEAD` at the start
   of `release` and export it as `REF` for every `build-*` beneath. The machinery already exists
   — `REF ?= HEAD` means an operator can do this **today** with
   `make release REF=$(git rev-parse HEAD) …`. Making `release` do it itself is what removes the
   operator's chance to forget. Costs a handful of makefile lines; changes no contract; the
   escape hatch (`REF=<other>`) survives untouched.
2. **`verify-agent-images` asserts one revision across all 14 local images at `IMAGE_TAG`.**
   Detects rather than prevents, but it is cheap, read-only, and it is the check that would have
   caught today's instance without anyone looking for it. Note the existing stanza checks
   **only `agent-chassis`** and defaults `EXPECT_SHA` to `git rev-parse HEAD` — on a tree this
   busy that prints `NO MATCH` routinely and correctly, which trains the reader to ignore it.
   Worth fixing in the same pass: compare against the **image label**, not against live HEAD.
3. **Runbook: "always pass `REF=` to a release".** Weakest, and by the standing rule an
   operator-must-remember step is a defect rather than a fix. Acceptable only as an interim.

Candidates 1 and 2 are complementary; 1 without 2 leaves nothing watching for the next way skew
gets in (a re-push at an existing tag, a single-service rebuild).

## 5. How to verify a fix

```bash
# After a pinned release, every image at the tag must carry ONE revision:
for S in auth-service core-manager agent-chassis reasoning-agent web-search-adapter \
         web-scrape-adapter git-adapter image-generator-adapter thunder-adapter \
         analyser-adapter browser-runner-adapter content-creator-agent \
         remote-job-spawner kafka-scheduler; do
  docker image inspect docker.io/aqls/$S:$IMAGE_TAG \
    --format '{{index .Config.Labels "org.opencontainers.image.revision"}} '"$S"
done | sort | awk '{print $1}' | uniq -c        # EXPECT: exactly one line
```

At the pods, the second independent witness (does not depend on any probe of the binary):

```bash
kubectl -n ai-persona-system logs -l app=<service> --tail=300 | grep -m1 'build provenance'
```

> **GOTCHA — do not verify this with a binary probe alone, and never with `strings`.** The
> `grep -aq "<sha>" /proc/1/exe` form is right, but with `2>/dev/null` a failed exec, a missing
> tool or a permissions refusal is indistinguishable from a true "not stamped". While finding
> this bug, eight services first read as `NO MATCH` from exactly that command, and it took a
> **positive control on the same pod** (`grep -aq <its own sha>` → MATCH, `grep -aq <the other
> sha>` → no match) to establish the negatives were real. The startup log line is the cheaper
> and more honest instrument. See `bugs_open/153` §CONTRIBUTION and `LANDMINES.md`.

## 6. Verification status of the report itself

Everything above is first-hand, from two independent instruments that agree:

- **Image labels** (`docker image inspect`, build machine) — revision + created timestamp, 14/14.
- **Running pods** (startup `build provenance` log line) — 14/14, agreeing with the labels.
- **Binary probe** (`grep -aq <sha> /proc/1/exe`) on a sample, with positive and negative
  controls on the same pod (`git-adapter`: own sha MATCH, chassis sha NO MATCH).
- **The cause** is read off `makefile:111`/`makefile:128`, not inferred from behaviour.

> **On the 2026-07-31 owner ruling** (a `bugs_open/` file asserting a cross-cutting root cause
> goes through the `090` diagnosis loop, or the filing session states plainly why it substituted
> equivalent first-hand verification): **substituted, deliberately, and here is the why.** The
> ruling exists for causes that live somewhere other than the symptom — the loop's value is that
> it reads the function you skipped. Here the cause and the symptom are the same three lines of
> makefile, the mechanism is arithmetic on two timestamp series, and there is no downstream
> consequence being asserted. A `090` run remains available and would be a reasonable belt-and-
> braces call if the fix grows past candidate 1.

## 7. Relations

- **`bugs_open/153`** — image tag does not imply a rebuild. This bug is only *visible* because
  153 shipped; it is the first thing the new mechanism found, and it found it unprompted.
- **`bugs_open/237`** — `render-audit-adapter` is in no release path. Adjacent: 237 is a service
  the release never touches, 249 is a release that touches services at different revisions.
- Concept register **BLD-019** (`register/build-pipeline.md`) — the provenance mechanism.
- `docs/agent_docs/docs024_key_docs_latest/bugfix_153_build_provenance/` — the lane, its RUNBOOK
  (R5 is the per-pod verification, R6 the still-owed induced-fault test).
