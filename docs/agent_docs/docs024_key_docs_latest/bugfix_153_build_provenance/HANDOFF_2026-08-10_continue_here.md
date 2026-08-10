# HANDOFF — `bugfix_153_build_provenance`, 2026-08-10 night · **start here**

## 1. Headline

**`bugs_open/153` is FIXED AND LIVE.** Every backend service now says, in its own binary, which
commit built it. Before the `v1.0.1283` roll that was **0 services, for the platform's entire
life** — the makefile promised provenance "verified against the running pod" and there was
nothing in the pod to read.

```
live stamp:   d3c09cc746e563b6339831cfb69576eb52135c43
services:     14 of 14, verified at the running binary on each
before:       0 of 14
```

**The file stays in `bugs_open/`** (owner ruling 2026-08-06), and one real proof is still owed
— see §4.

## 2. What was built (candidates 1 + 4 of the bug's own ranking)

The framework lever was the makefile's two build macros, so one edit covers all 14 services
rather than one:

- **`pkg/buildinfo`** — one `GitCommit` var, the stamp target. Separate package so a single
  identical `-ldflags` string works for every service regardless of build style.
- **`ref_build` / `tree_build`** — pass the sha `ref_build` already computed at `makefile:119`
  purely to echo and discard, as `--build-arg GIT_COMMIT` plus OCI `image.revision`/`created`
  labels. `tree_build` stamps `<shortsha>-tree` **unconditionally**, so a WIP image can never
  wear a clean sha.
- **14 dockerfiles** — `ARG GIT_COMMIT` + the ldflags clause on the existing `go build`, each
  file's own flags and output path preserved.
- **14 mains** — one import + one `build provenance` log line. **Load-bearing:** `-X` against
  an unlinked package is silently ignored, so the import is what makes the stamp take.
- **`verify-agent-images`** — reads the label and the pod's binary back.

**Commits:** `e743e6cfc` · `e5f31dcdb` · `1054ec36c` · `8d270c68a` (editquality fix) ·
`c4a932680` (probe fix). Register: **BLD-019**.

## 3. Council + owner — read this before touching the trailers

Round 1 on corr `44fa6a98-acaa-46b5-9ada-f0c34ca5475d` returned **REJECTED — hard veto from
`guardian`, on SCOPE not soundness.** It said plainly: *"The mechanism itself… is sound and
well-evidenced — that part I'd approve on a single-service pilot."* Its objection was that one
round bundled the shared macro with all 14 dockerfile and all 14 main edits.
**`bug_historian` and `reuse_agent` both APPROVED** — the seats disagreed.

Per CLAUDE.md's 2026-07-28 ruling (a scope veto is not answered by resubmitting), it went to
the owner with three costed options. **The owner chose option 1: the code stands as committed.**
Same call as `bugs_closed/124`, eleven days apart.

> **Therefore: LIVE AND PROVEN ≠ APPROVED.** The commits carry `Council-Submitted:`, which is
> accurate. **No `Council-Reviewed:` trailer exists on any of them and none may be added** —
> the verdict was REJECTED, and writing that trailer is the coverage report's MISMATCH bucket.
> The honest one-liner is: *reviewed, vetoed on scope, overruled by the owner, recorded.*

**Noted, deliberately not acted on:** the guardian's "MANY packages" trigger fires on edit
count and file spread and does not appear to distinguish a mechanically-identical, provably
inert, every-instance-verified change from N independent judgements. The owner has now twice
sided against it on that shape. **If a third case lands the same way, that is a rate and the
calibration deserves its own RFC** — raised by whoever hits it third, with three data points.
Do not open it on the strength of this lane alone.

## 4. What is still owed

1. **The induced-fault test (RUNBOOK R6) — the only outstanding real proof.** Bump
   `IMAGE_TAG`, run push+deploy **without** `build-*`, confirm the pod comes up wearing the
   NEW tag while honestly reporting the **OLD** sha. Everything verified so far covers an
   *honest* roll; only this covers a **dishonest** one, which is the actual defect. Needs an
   owner cycle. **Expect it to be visible, not refused** — this detects; refusing is
   candidates 2/3.
2. **Two CronJob images** (`component-render-check`, `shared-output-fields-check`) get the OCI
   label but their binaries are unstamped by design. Labelled ≠ stamped.
3. **Candidates 2 and 3 remain unbuilt on purpose** — both change the push/deploy contract
   fleet-wide and want owner sign-off. This change is their prerequisite.

## 5. ⚠ The verification probe is itself a trap — it fooled me twice in one hour

**Do not verify this (or anything) with `strings`.**

- **`strings` is ABSENT from `browser-runner-adapter`** — the fleet's only debian-slim image
  (Chromium needs glibc), which ships no binutils. Behind the customary `2>/dev/null` it
  returns a clean `0`, indistinguishable from "not stamped". It briefly convinced me the new
  mechanism had caught a stale binary on its first run. **This makes CLAUDE.md's own fleet-wide
  pod-grep recipe unsafe on debian-based services.**
- **`ls /path/binary` can fail on PERMISSIONS** (`git-adapter`, as `appuser`), a second false
  "not there".
- **Never use a DISCOVERY grep.** Dropping `strings` for `grep -aoE "[0-9a-f]{40}"` also drops
  its line boundaries, so `^`/`$` mean nothing and the match lands in Go's internal digit
  table — every service returns `0001020304050607…`, identically, with no error.

**The settled form — verify a value you already know:**

```bash
kubectl -n ai-persona-system exec <pod> -- grep -aq "<expected-sha>" /proc/1/exe \
  && echo MATCH || echo "NO MATCH"
```

`/proc/1/exe` resolves the running binary for any image base and any binary path. **Always run
a control in the same breath** — a fabricated sha must return no-match.

## 6. Lane files

`PLAN_2026-08-10_build_provenance.md` (the fable-drafted plan; §8 lists the non-goals) ·
`RUNBOOK_build_provenance.md` (R1–R8, each command with its gotcha; **R6 is the owed test**) ·
`NOTES_build_provenance.md` (append-only; three missteps recorded) · `README_where_we_are.md`
(owner's plain prose) · `SUBMISSION_2026-08-10_build_provenance.json`.

Outside the lane: `bugs_open/153` (status block has the full verification record),
`docs026_concept_register/register/build-pipeline.md` **BLD-019**, `LANDMINES.md` (three
entries from this lane), `WRONG_CALLS.md` (two entries).

## 7. Also filed by this lane, not part of 153

**`bugs_open/243` (`243-anthropic-cap`)** — the Anthropic account hit its usage limit at
14:51:47Z, killing every LLM step fleet-wide. **Resolved the same evening: the owner added
credit, fleet back at 18:12:11Z**, 21 days before the API's stated auto-restore — by action,
not calendar. **Stays open**: adding credit restores service, it does not prevent recurrence,
and this was the second Anthropic exhaustion in eleven days (third single-provider outage
counting `bugs_open/202`'s Gemini 429). All 127 LLM steps across 55 agents name one provider
and one key. **The actionable prevention is `bugs_open/244`** (another lane): council-gate is
**87.8% of the fleet's August spend** at 790,551 input tokens per round, with ~76% available
from prompt caching.

## 8. If you are picking this up cold

The lane's work is done bar item §4.1. Highest-value next steps, in order:

1. **Run the induced-fault test** with the owner (R6). It closes the last real gap.
2. **Use the stamp.** It is live and under-exploited — `git merge-base --is-ancestor <commit>
   d3c09cc74` now answers "did my fix ship?" exactly, for any commit, on any service. Another
   lane already used it to settle **19** roll-blocked register entries in one pass
   (`ebaac39c0`). Anywhere a doc says "inert until the next roll", that is now a query.
3. **Retire marker-hunting where you find it.** CLAUDE.md and many lane handoffs still
   prescribe per-fix `strings` markers — the practice that produced three false readings in one
   day (`bugs_open/153`'s CONTRIBUTION section) and which this mechanism exists to replace.
   Updating CLAUDE.md's "Building & deploying images" section is owed and was **not** done by
   this lane.
