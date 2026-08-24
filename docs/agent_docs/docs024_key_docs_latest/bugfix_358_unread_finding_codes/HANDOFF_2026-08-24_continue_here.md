# HANDOFF 2026-08-24 — bugs_open/358, continue here

**Supersedes `HANDOFF_2026-08-22_continue_here.md`, which is now wrong in two places** (§0 numbers,
and §2's "Phase 2 not started"). Read this one; that one only for the traps in its §3, which still
hold.

**Re-measure before quoting any number here.** Everything decays on a 30-day sliding window, and
this lane exists because of that. One command re-establishes the state:

```bash
./scripts/audit-finding-codes.sh          # exit 0 = every observed code is declared
```

---

## 0. What this lane is, in one paragraph

`agent_error_log` carries deliberately-written **finding codes** — a detector's record of something
it noticed and will not fix. `bugs_open/358` measured that most have no automated reader and are
deleted at 30 days unresolved (14 if resolved — **marking a row resolved makes it die faster**). The
fix is not a reader per code: some are legitimately human evidence, some are time-boxed
instrumentation, and operational plumbing is *correctly* consumed by generic newest-N diagnostic
reads. The fix is that **a code cannot enter the estate with no declared disposition and nobody
notice.**

## 1. Status — DONE, and what is left

**Phase 1 (registry + checker)** — done 2026-08-22, council `be1fd678`.
**Phase 2 (the daily CronJob)** — **DONE AND LIVE 2026-08-24.** Council `be252395`, APPROVED round 4.

| thing | path |
|---|---|
| the registry | `docs/agent_docs/docs024_key_docs_latest/architecture_review/finding_code_registry.json` |
| the checker | `cmd/config-key-audit/findingcodes.go` + `findingcodes_test.go` |
| hand-run wrapper | `scripts/audit-finding-codes.sh` |
| **the CronJob** | `deployments/kustomize/services/finding-code-registry-check/`, **07:30 UTC daily** |
| **the image** | `build/docker/backend/finding-code-registry-check.dockerfile` (in `RELEASE_IMAGES`) |
| **source-side scan** | `platform/orchestration/actions/findingcodes_scan_test.go` |
| **its commit-time runner** | `scripts/check-finding-code-registry.sh` → `.githooks/pre-commit` (hook 5) |
| shared hook mechanics | `scripts/lib/precommit-gotest.sh` — ⚠ **fleet-critical, see §4** |
| register | `docs026_concept_register/register/debugging.md` → **DBG-075** |

**LIVE READING 2026-08-24 evening, after the v1.0.1335 roll:** 43 codes observed, **55 declared**,
**0 findings**, **25 unruled** (cap 25), 12 registered-but-unobserved, retention parity 0
disagreements, `_scan_baseline` 13. **The day-one loop is CLOSED and proven at the artefact**: the
deployed image's build stamp (`48f55f218…`) has both the declaration commit and the scan/baseline
commit as ancestors (`git merge-base --is-ancestor`), and a manual in-cluster Job exits 0 with a
clean row. First *scheduled* fire: 07:30 UTC 2026-08-25 — worth one glance at the doc_notes row
that morning, then leave it alone.

### What is NOT done — in order of who decides

1. **B1, the remaining 25 `unruled` codes.** *Owner's to rule*, batch at a time. He ratified batch 1
   (7 codes) on 2026-08-23. **This is the real remaining work** and the count is the progress
   metric: 32 → 25.
2. **`_scan_baseline`, 13 codes written by `platform/orchestration/actions` and declared nowhere.**
   All zero-row today. Same shape as (1) — each needs a ruling, and the list may only shrink.
3. **A reader for `LINK_CONTEXT_UNAVAILABLE`** — *not this lane's*, see §3.

## 2. The three things you must not get wrong

**(a) `agenterrors.go:3` says "The ONE writer against `agent_error_log`". IT IS NOT** — a check
placed there covers one writer in five while reading completely clean. Re-census rather than
quoting the number (a census goes stale by ADDITION):
```bash
grep -rn "INSERT INTO agent_error_log" --include='*.go' --include='*.sql' --include='*.py' --include='*.sh' .
```
That is why the authority is `SELECT DISTINCT error_code` and not a source scan.

**(b) FOUR ways a code hides from a source scan**, all of which have produced a wrong answer here:
a Go **constant** (§3.2 — caught the bug's own author); a **positional** argument to
`LogActionError(ctx, params, siteID, domain, action, code, …)`; a code built at **runtime**; and a
writer **outside** the actions package (one is SQL). The new scan resolves constants and **states
the other three as blind spots at the top of the file**. Do not read it as coverage.

**(c) The scheduled job runs `--no-source`, deliberately.** In a container the two arms that open a
`consumed` entry's reader FILE would raise 5 `reader-unreadable` findings **every morning against a
healthy registry** (measured). They grade the registry against source, both halves change only by
commit, so they run at commit time instead. **Every run states which arms it skipped.**

## 3. The finding on day one, and it is NOT this lane's to fix

The CronJob's first run exited 1 on `[undeclared] LINK_CONTEXT_UNAVAILABLE`, ~2h after that code's
first row. **The code is recording a real degradation:** both rows are `page-content-writer` hitting
a **query timeout (SQLSTATE 08P01)** loading linkable pages — `severity=error`, `degraded=true`,
outcome *"writer instructed to emit NO internal links"*. Two pages were written with no internal
links, the system wrote that down, and nothing reads it.

**Owner: `bugfix_092_writer_link_constraints`** (quiet 14d — `scripts/who-owns.py 092`). Writer:
`platform/orchestration/actions/prepare_link_context_action.go:488`. It is declared
`human-evidence` here **by this lane, not by its author**, and the registry entry says so — the
disposition is the owner's to overrule, the measurement (no reader anywhere) is not in doubt.

**If you take one thing from this lane to another: that row is the best argument in the registry
for giving a code a real reader.**

*(Loop status: both day-one codes were declared the same day, the v1.0.1335 roll carried the
declarations, and the in-cluster run is green. The DEGRADATION itself — writers losing their link
context on DB timeouts — remains unfixed and unowned-in-practice; it is a candidate for a fresh
bug file routed at the 092 lane, not for silent adoption.)*

## 3a. The Fable review pass — DONE (second attempt), five findings, all fixed

The owner asked for an independent pass by Fable. The first attempt died on an API session limit;
the second (budget-aware brief) completed. **Five findings, all verified against the code, all
fixed and mutation-proven the same day.** The one that matters:

> **HIGH — the scan test was a hand-run tool wearing a commit-time label.** Its header claimed a
> new `ErrorCode:` is caught at commit; `check-finding-code-registry.sh` tested only
> `./cmd/config-key-audit/` and never named an actions file. The exact commit shape that produced
> `LINK_CONTEXT_UNAVAILABLE` would still have walked through the hook. **Now fixed:** the hook keeps
> two relevance sets and runs the actions package when a STAGED actions file carries `ErrorCode:`.
> This was yesterday's wrong call one level up — *existence is not execution* — and it is in
> `WRONG_CALLS.md` as a repeat.

The other four (const aliases dropped silently → now resolved and every unresolvable site
REPORTED; function-local vars misattributed → file-scope consts only; a vacuous `-run` control →
one shared `scanOrFatal`; an overstated comment → corrected) and what Fable verified HOLDS are in
NOTES under "The Fable pass, second attempt". **Nothing in this lane needs a further review pass, and nothing is awaiting a verdict** — every
correlation this lane opened (`be1fd678`, `be252395`, `2e5f687d`, `4d5c1523`) is APPROVED.

**If you run the scan verbosely** (`go test ./platform/orchestration/actions/ -run
TestEveryErrorCode -v`) you will see four `UNRESOLVED ErrorCode:` lines. Those are the stated
runtime blind spot — local variables at `component_write_guard.go:501`, `log_action_error.go:252`,
`v3_site_actions.go:4197`, `:4261` — made visible, not a defect. Only the daily live-table check
sees what those sites write.

## 4. Traps this lane created or found — read before touching

- **⚠ `scripts/lib/precommit-gotest.sh` runs on EVERY session's commit** (via the pre-existing
  RFC_022 guard, which now sources it). A `set -e`, or a `return $rc` that looks more honest than
  `return 0`, turns an advisory guard into a blocking one **for the whole fleet at once**. Every
  path must return 0. It has a LANDMINES entry. Its `subject` and `failure-headline` parameters must
  stay separate — merging them printed a line asserting a finding and disclaiming one at once.
- **⚠ Bumping `IMAGE_TAG` arms a fleet-wide `ImagePullBackOff`.** `deploy-agents` does not deploy
  the tag an overlay names — it **`sed`s every overlay to `$(IMAGE_TAG)`** and applies. Between a
  new service's birth commit and the next release, one image exists at the new tag and ~32 do not.
  Run the **whole** `make release`, never `deploy-` alone. LANDMINES entry, measured.
- **The hook runs the actions package with `-run '^TestFindingCode'`, and that prefix is a
  CONTRACT, not a roster:** `TestFindingCodeTestsFollowTheHookConvention` fails on any `Test*` in
  `findingcodes_scan_test.go` or `finding_code_roster_test.go` outside the prefix. If you add a
  registry-grading test, name it `TestFindingCode…` — in one of those two files, or it is unseen
  at commit time (stated limit). Whole-package was rejected by the guardian for a reason this lane
  agrees with: an unrelated flake would surface as "this package writes an undeclared code".
- **⚠ A comment saying a test runs "at commit time" is a claim about a HOOK.** Verify it with
  `grep -n 'go test' .githooks/pre-commit scripts/check-*.sh` — not by confirming the test file
  exists. This lane got it wrong both ways in two days (file did not exist; then file existed and
  nothing ran it).
- **⚠ `UNKNOWN` is a PREFIX of `UNKNOWN_HANDLER_VERDICT`** (a `_scan_baseline` code, writer
  `complete_work_item_verification.go:394`). The checker's prefix-collision rule is pairwise and
  unconditional over every declared code, so **the day that code is declared, the daily check
  exits 1 on `prefix-collision` and its author will not know why.** No live `LIKE 'UNKNOWN%'`
  exists (measured 2026-08-24), so it is a rule artefact today, not a query hazard. Remedies:
  rename the code, or scope the rule to family prefixes — the owner's call, because the rule
  exists for real `LIKE` families and relaxing it is a guarantee change.
- **The hook helper's build-failure classifier is now the toolchain marker** (`[build failed]` /
  `[setup failed]`), not `cannot find|undefined:|syntax error`. If a real failure ever reads as
  "NOT CHECKED", that is the first place to look — and the old bag-of-words is the reason it was
  changed.
- **⚠ `.dockerignore` excludes `docs/`** with one `!` un-ignore per shipped file. A new
  registry/acks-shipping check needs a line there or the `COPY` fails at build — the loud direction,
  and why the registry is a COPY not a mount.
- **⚠ 098's coverage report kept a SECOND, hand-kept path list** (`SCOPE_PATHS`) as a pre-filter, so
  `cmd/config-key-audit` commits were invisible to it — not unreviewed, **absent** — for a fortnight
  (22 commits, four lanes). Fixed 2026-08-24 in both files. If you widen council scope, **edit both**.
- **Do NOT hand-roll `git archive HEAD | tar`.** Use `scripts/verify-head-builds.sh`; it takes
  repeated `--with <file>` (you will need the test *and* the registry together, or the baseline
  reads as missing). I burned 2.6 GB of scratch this way before the rule landed.
- **A schedule slot is claimed by the REPO, not the cluster.** Census
  `deployments/kustomize/services/*/base/cronjob.yaml`, not `kubectl get cronjobs`.
- **Testing a pre-commit hook needs an isolated scratch repo**, and restore with
  `git checkout HEAD -- <file>`, never `git checkout -- <file>` (that restores from the INDEX, which
  still holds the copy you just staged, so your control reports a failure that is not real).

## 5. How to prove any of this — the artefact, never the tag

```bash
kubectl -n ai-persona-system get cronjob finding-code-registry-check \
  -o jsonpath='{.spec.jobTemplate.spec.template.spec.containers[0].image}'; echo
# NEVER wait for the schedule:
kubectl -n ai-persona-system create job --from=cronjob/finding-code-registry-check fcrc-manual-$(date +%s)
# then the POD's exit code and log, then:
```
```sql
SELECT created_at, left(body,600) FROM doc_notes
 WHERE source = 'finding-code-registry-check' ORDER BY created_at DESC LIMIT 7;
```
A **missing row means the job did not run**, never "nothing is wrong". The body states the registry's
declared-code count and the binary's build commit — a row disagreeing with a local
`./scripts/audit-finding-codes.sh` **is** image staleness, visible rather than silent.
Note a failing job writes **two** rows (`backoffLimit: 1` retries), which is expected.

Full command set with gotchas: `RUNBOOK_unread_finding_codes.md`.

## 6. Correlations, for the trail

- Council `be1fd678…` — phase 1, APPROVED round 2.
- Council `be252395-9d51-4427-b2ae-5f581337b16d` — phase 2, **APPROVED round 4**. Rounds 1–3 were
  all gated on the same defect of mine: a fact true of the tree and unshowable from the plan. Round
  4 generates every sketch from the committed diff. `WRONG_CALLS.md` 2026-08-23.
- Council `2e5f687d-5753-441b-91f3-406c84a98394` — the source-side scan, **APPROVED 2026-08-24**,
  10 seats, 7 abstained, no gating objection. All four advisory objections say one thing: the sketch
  was truncated over the logic. Cause was mechanical (42 lines of header comment, cap applied in file
  order), fix is a RUNBOOK rule — strip comments before truncating. No code change owed.
- Council `4d5c1523-2453-4799-b828-25379affc41b` — the Fable-findings fixes. **Round 1 REVISE**
  (guardian HIGH: the hook ran the whole actions package — a misattributed headline on any
  unrelated flake; measured 86/411 commits would trigger). **Round 2 submitted** with the hook
  scoped to `-run '^TestFindingCode'`, a naming convention ENFORCED by
  `TestFindingCodeTestsFollowTheHookConvention`; all four proofs in NOTES. Commits `bce49226a`
  (round 1) + `71c4081cc` (round 2). **Round 2 APPROVED 2026-08-24**, 8 seats, 9 abstained; the
  four advisory objections are all already-stated limits (NOTES, "round 2: APPROVED").
  **Nothing in this lane is awaiting a verdict.**
- `090` `c965bfec…` — UNVERIFIABLE (scope-not-narrowing), **not a refutation**: its static tier is
  `.go`-only, so the SQL writer was outside its corpus.

## 7. The two corrections this lane owes its own past self

1. **`findingcodes.go` cited `findingcodes_scan_test.go` for two days and the file did not exist** —
   through a council round and into the concept register. Now real. **A comment naming a file is a
   claim, and it is the one class of claim that costs one command to verify.**
2. **"That blind spot is harmless BY CONSTRUCTION" was too strong.** An unfired code costs nothing
   *today*. The population is **thirteen**, and each becomes a live undeclared finding the moment it
   first fires — which is exactly how `LINK_CONTEXT_UNAVAILABLE` arrived. Bounded and short-lived is
   the honest claim.

Both in `WRONG_CALLS.md` (2026-08-24).

## 8. The five living docs

| doc | what it holds |
|---|---|
| `PLAN_2026-08-22_unread_finding_codes.md` | design, the four dispositions, §6a on what phase 2 actually needed |
| `RUNBOOK_unread_finding_codes.md` | every command, gotcha attached, incl. the phase-2 deploy/verify set |
| `NOTES_unread_finding_codes.md` | evidence log + **every misstep**, newest at the bottom |
| `README_where_we_are.md` | the owner's plain-prose log |
| `SUMMARY_2026-08-22_unread_finding_codes.md` | the milestone read-out (a new one is owed if day one's findings change the story) |

Bug file: `bugs_open/358_HANDOFF_2026-08-22_agent_error_log_finding_codes_are_write_only_and_expire_unread.md`.
