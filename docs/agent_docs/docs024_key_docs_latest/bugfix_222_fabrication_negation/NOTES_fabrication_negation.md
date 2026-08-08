# NOTES — bugfix 222 (fabrication gate negation-blindness)

Running record, append-only, newest at the bottom.

## 2026-08-08 evening — picked up, research done

**Ownership check first (per CLAUDE.md).** `git status` on session start showed
substantial uncommitted WIP already in the tree for `bugs_open/220` — a PLAN,
NOTES, README, RUNBOOK, a council submission JSON, and two SQL migrations, all
under `bugfix_220_unbuilt_link_dispatch/`. File mtimes were **2–4 minutes old**
against wall-clock at the time I looked (19:09–19:10 vs 19:13), and `git log`
showed the code fix already committed (`a60a13cbb`, `03433f4b5`) with the
council submission still in flight (correlation `def4441c-…`). That is an
active thread mid-task, not stale leftovers — did not touch it, did not commit
any of its files.

Surveyed the rest of `bugs_open/` for ownership with `scripts/who-owns.py` plus
manual `git log --grep` and directory checks on the newest bugs (217–226):
- 220, 221, 226: actively owned (216/220/226 all have same-day commits or
  in-flight council rounds).
- 224, 225: filed by `loanandmortgagecalculator_couk`, which is itself very
  active today (40 commits/14d) and whose own README says the SDLT fix (225)
  needs an **owner decision** before any code changes ("Do not change the
  number without the owner seeing this file first"). Skipped both — owned lane,
  and 225 is explicitly blocked on the owner, not on an implementer.
- 223: genuinely unowned (self-declared "OPEN, UNOWNED" in the bug file,
  `who-owns.py` found no workstream), but its own fix candidates route to
  `architecture_review` (RFC_005 §3.2 is their mechanism) — picking it up cold
  risks colliding with a lane that has explicitly claimed the surrounding
  design space. Left for that lane or a session with more RFC_005 context.
- **222: picked.** Filed by `mortgagecalculator_couk_adoption` but explicitly
  left as a workaround-until-fixed item in their own handoff ("the 222
  comment-style workaround clause … until 222 is fixed") — they are not
  implementing the platform fix themselves. Self-contained: one file
  (`platform/orchestration/actions/check_tool_fabrication_action.go`), clear
  repro payload already in the bug file, clear fix candidates, no DB schema
  involvement, no cross-lane architecture question.

**Bug still valid?** Read `check_tool_fabrication_action.go` at current HEAD
(clean in the working tree — no other session has a WIP diff on it). The Tier A
declaration regexes (`fabQualifierNearData`, `fabDataNearQualifier`,
`fabGenerateVerbData`, lines 91–93) still have no negation awareness. Confirmed
valid.

**Prior art checked:**
- `016b_debugging_guide_8_consolidated.md` §9 already carries the write-up for
  this exact bug (added by the filing lane, "A proximity detector convicts the
  DENIAL…", line ~11314) — diagnosis and transferable rules already recorded,
  nothing to re-derive.
- `LANDMINES.md` / `WRONG_CALLS.md`: no prior entry about
  `check_tool_fabrication_action.go`'s negation-blindness specifically. One
  unrelated WRONG_CALLS entry (2026-07-22) about this same file's config-path
  resolution bug (fixed; the code's current comment at line ~226 explains why
  `ExtractActionInputs` is deliberately not used here) — different bug,
  same file, no interaction with this fix.
- Memory topic files `prompt-text-poisons-its-own-detector` and
  `mutate-the-code-to-prove-the-guard` are the right precedent class (proximity
  detector fooled by its own vocabulary; the standing rule that a guard needs a
  mutation that survives, not just one that fails). No existing negation-aware
  regex helper in the codebase to reuse — this will be the first.
- No DB/architecture involvement: `DetectToolFabrication` is a pure function
  (recreation HTML string, original HTML string, bool) with no DB reads. No
  need to check live `agent_definitions` — the recreate-tool prompt's exact
  wording is not being changed, only the detector's precision.

**Test file already has 12 cases** covering Tier A/B, corroboration, fail-safe,
and `dataSourceIsExternal` — the negation fixture slots in as a 13th, using the
exact payload from the bug file, plus the required inverse (a genuine
declaration containing a negator elsewhere in the string, to prove narrowing
doesn't blind the detector to real fabrications framed defensively).

> **CORRECTED 2026-08-08, same evening, by the fable planning pass.** The
> paragraph above says "No existing negation-aware regex helper in the
> codebase to reuse — this will be the first." **False.**
> `negatedClaimMatch` in `platform/orchestration/datahelpers/claims.go:593`
> already does exactly this, registered as **CLM-017**, with two consumers
> (`ScanBannedClaims`, `ScanAttributedUncitedStats`) and a deliberately pinned
> residual test. I had grepped `LANDMINES.md`/`WRONG_CALLS.md` for
> "check_tool_fabrication" and the two named memory topics, and separately
> grepped the codebase for "negation-aware regex helper"-shaped absence, but
> never grepped `platform/` for the plain word **"negat"** before writing the
> "this will be the first" sentence — that grep (run afterward, on a hunch
> about the claims-verification lane) returned 30 files instantly, including
> the exact mechanism. **The cheap check that would have caught it:**
> `grep -rln "negat" platform/ --include=*.go` before asserting absence of a
> mechanism, not after. Textbook instance of "a grep proves absence only for
> the spelling it searches" (memory topic of that name) — "negation-aware
> regex helper" was never going to be the register's own spelling
> ("negation guard", CLM-017). Logged to `WRONG_CALLS.md` the same evening.
> The plan (`PLAN_2026-08-08_negation_aware_declaration_tier.md`) is built on
> reusing CLM-017's algorithm, not reinventing it — no downstream damage, but
> the false claim sat in this file for roughly 40 minutes before the fable
> pass caught it, and would have shaped a worse design if I had started
> implementing straight from the first pass of research instead of routing it
> through a planning step.

## 2026-08-08 evening — plan complete, implementing next

Fable's plan (`PLAN_2026-08-08_negation_aware_declaration_tier.md`) is in:
extract `NegationGuard` (algorithm) from `claims.go`, keep its vocabulary
untouched, give the fabrication gate its own cue set including bare
"no"/"without"/"zero" (deliberately excluded from the claims layer for
different, domain-specific reasons), scan from the qualifier token's submatch
position (not match start) because `fabDataNearQualifier`/`fabGenerateVerbData`
can put a negator inside the matched span (the CLM-019 lesson). Full test list
incl. a mutation-and-control triple and two CLM-017-style pinned residuals.
Two pre-existing, unrelated precision defects spotted in passing
(`invent\w+` matches "inventory"; `realistic` has no left word-boundary) —
noted, not folded into this fix; will grep bugs_open/closed before filing
separately so as not to duplicate if already known.

## 2026-08-08 evening — implemented, tested, mutation-proven

Implementation followed the plan exactly: `NegationGuard`/`NegatedAt`/
`claimNegationGuard` extracted in `claims.go` (`negatedClaimMatch` now a
one-line wrapper); T-0 confirmed the claims-layer suite green with **zero**
test-file edits. Fabrication-domain vocabulary, `declPattern`/
`firstAssertedDeclaration`, and the rewired Tier A chain landed in
`check_tool_fabrication_action.go`. One implementation deviation from the
plan's literal pseudocode: the plan sketched a nested if/else cascade for the
three Tier A arms; that produced a real `go build` scoping error (a
short-circuit `if`'s inner `suppressed` var went out of scope with no `else`
to hold it) — rewritten as a `for` loop over an ordered slice of
`{pattern, label}`, same priority (first non-empty match wins), same
suppressed-notes accumulation, cleaner. Caught by the compiler, not by
review — worth remembering that a plan's illustrative code is a sketch, not a
diff to paste.

**Reproduction-first discipline paid off twice.** Wrote T-1/T-3/T-4/T-7
*before* touching the fix and ran them red-first, per the plan's execution
order:
- T-1, T-3 (both sub-cases), and 5 of 6 T-4 rows failed exactly as expected
  against unfixed code.
- **One T-4 row didn't fail at all**: `"we fetch real data instead of
  generating it"` — turned out `fabDataNearQualifier`'s own data-noun
  alternation has no bare `\bdata\b` (only `dataset`/`records`/`entries`/…),
  so the fixture never matched ANY Tier A regex, guarded or not. Fixed the
  fixture to `"...real records instead of generating them"`, which does
  match. **A fixture that doesn't reproduce isn't proof of a fix — it's proof
  of nothing**, and this would have shipped as a silently-vacuous test row
  had I not checked each one failed red first.
- **A second, subtler fixture bug** in the same test: `"zero fake records —
  nothing is pre-seeded"` still convicted AFTER the fix, but not via the
  arm the row was meant to exercise. `fake records` (the intended
  `fabQualifierNearData` match) WAS correctly suppressed by "zero" — but the
  sentence also contains an accidental SECOND, independent Tier A match:
  `records … seeded` via `fabDataNearQualifier`, because that regex's
  alternation for the qualifier includes bare `seeded` with no word boundary,
  so it matches inside "pre-seeded". That second match sits in a different
  clause (an em-dash boundary intervenes) and 36 bytes from "zero" — outside
  the 32-byte window even before the clause trim. **The fix behaved
  correctly**; the test fixture accidentally exercised an unrelated,
  pre-existing regex looseness (`seeded` unanchored) I hadn't gone looking
  for. Reworded to `"zero fake records here — every one is user-entered"`,
  which isolates the one match the row is about. Both fixture bugs were
  caught by the row failing when it should have passed, not by reading the
  regex — the same "mutate/observe, don't just read" discipline the plan's
  own mutation step is built on, showing up one level earlier.

**Mutation triple (T-6), run against a `git archive HEAD` overlay with just
these four files copied in** (the shared tree carries unrelated dirty files
from other sessions — verified the two failures below are pre-existing at
clean HEAD, unrelated to this change, before trusting any pass/fail here):
- **Mutation A** (`NegatedAt` → always `false`): T-1, T-3, T-4 failed red;
  T-2, T-5, T-7 stayed green. Exactly as predicted.
- **Mutation B** (`NegatedAt` → always `true`): T-2 and T-7 failed red, as
  predicted. **T-5's `TestDetect_VetcompFabrication_Gated` and
  `TestCheckToolFabricationAction_ReadsDottedConfigPath` did NOT fail**,
  contrary to the plan's prediction — and the reason is informative rather
  than a bug: the vetcomp fixture also trips the synthetic-PII arm
  (`makePostcode`), which is deliberately NOT negation-guarded, so it stayed
  convicted through that independent path even with every Tier A declaration
  arm suppressed. This is the risk register's R1/R4 claim (defence in depth;
  the PII arm is untouched) demonstrated by the mutation itself, not just
  argued for in prose.
- **Control** (`fabLiteralRecordThreshold` 15→14, an unrelated Tier B knob):
  every test above passed. Confirms none of the new tests are vacuously
  sensitive to any change in the file.
All mutations reverted; clean archive left in the scratchpad, not committed.

**Pre-existing unrelated failures, confirmed NOT caused by this change**
(reproduced on a clean `git archive HEAD` with nothing of mine applied, before
I touched anything): `TestValidDocSubjectTypes_LockstepWithMigrationCheck`
(cites `bugs_open/064`, though the fix actually lives in `bugs_closed/064` —
a stale string in the test's own message, not mine to fix here) and
`TestEveryCheckProducedItemTypeIsClassified` (`decision_regression` /
`check_decision_guards.go`, unrelated file, someone else's open gap). Left
alone — out of scope, already pre-existing at HEAD.

**Council submission (round 1):** `SUBMISSION_CORR=aa2d0d62-4aba-480e-aedc-8be264d53b01`,
run orchestration `7937ed59-2055-423d-b8c3-3c3a80c02ca9`. Dispatch lane
(`system.agent.generic.requests`) showed LAG 0 at submit time — expect the
run to start promptly, verdict within the standard ~30 min budget. Verify by
payload, not the printed id:
```sql
SELECT current_step, status FROM orchestration_states
 WHERE collected_data->'input_data'->>'fix_correlation_id' = 'aa2d0d62-4aba-480e-aedc-8be264d53b01';
```
Committing now with `Council-Submitted:` per the 2026-07-30 trailer rule —
not holding code for the verdict. Next: update bugs_open/222, commit by
pathspec (this workstream's files only), then check the verdict later.

**Live same-file-passenger hit, before I could commit.** Re-checked `git
status` immediately before committing (per CLAUDE.md: the session-start
snapshot goes stale within minutes) and found my earlier `WRONG_CALLS.md`
addition no longer shows as a diff — `git log` explains why: commit
`d53d04786` ("WRONG_CALLS: cited a neighbouring INSERT as the convention…"),
an unrelated entry from another session, landed on this file while my
addition sat uncommitted in the same shared working tree. Their pathspec
commit took the file's whole working-tree content (pathspec commits ignore
the index, per CLAUDE.md), so it carried my bugfix-222 entry in as a passenger
— confirmed present verbatim at `HEAD:docs/…/WRONG_CALLS.md` line 23796.
Nothing lost, forward-only holds: I am simply not naming `WRONG_CALLS.md` in
my own commit's pathspec (there is nothing left to add — the diff is empty).
This is the exact landmine `LANDMINES.md` already documents ("a pathspec
commit still takes a SAME-FILE passenger"), now hit live rather than just
read.

## 2026-08-08 evening — council round 1 blocked fleet-wide, not on this submission

Checked the verdict:
```sql
SELECT current_step, status FROM orchestration_states
 WHERE collected_data->'input_data'->>'fix_correlation_id' = 'aa2d0d62-4aba-480e-aedc-8be264d53b01';
-- current_step=complete_invalid, status=COMPLETED
```
`__step_error`: `step review_editquality failed ... AI endpoint unavailable:
... "Your credit balance is too low to access the Anthropic API."` — this is
the **same fleet-wide Anthropic credit exhaustion** the 220 lane's docs
recorded the same evening (31 failures 18:25-19:20Z, owner billing action
needed), not a defect in this submission. Nothing to fix on my side; a retry
now would fail identically. **Not resubmitting** — this needs the owner's
billing action first, then a fresh round (reuse `RESUBMIT_CORR=aa2d0d62-…`
so the trail accumulates). Commit already carries `Council-Submitted:`,
which is correct either way: it asserts nothing, and 098 will resolve the
verdict once a round actually completes.

**Status at end of this session's work:** fix implemented, tested (including
mutation-proof), committed (`f8cbaf551` + `06241a516`), council round 1
blocked on infra rather than reviewed. Owed: re-submit once Anthropic
credits are restored; read the real verdict then; if APPROVED, no code
change needed — just note the approval (098 auto-credits the existing
commit via the correlation, forward-only forbids an amend anyway). If
REVISE/REJECTED, act on it in a follow-up commit.

## 2026-08-08 21:27 BST — the "fresh chassis build" the user reported does NOT carry this fix

The user reported a fresh chassis build had been deployed and asked to carry
on. **Checked at the artefact rather than trusting the report** (per
CLAUDE.md — a roll is not evidence a specific fix shipped):

```bash
kubectl -n ai-persona-system get pods -l app=agent-chassis   # both replicas, age 89m
kubectl -n ai-persona-system exec <pod> -- sh -c 'strings /app/agent-chassis | grep -c "negated declaration ignored"'   # 0, both replicas
kubectl -n ai-persona-system exec <pod> -- sh -c 'strings /app/agent-chassis | grep -c "NegationGuard"'                  # 0, both replicas
kubectl -n ai-persona-system exec <pod> -- sh -c 'strings /app/agent-chassis | grep -c "declared synthetic/fake data"'   # 1, both — POSITIVE CONTROL: the gate exists, methodology sound
```

**Why, once the timestamps are lined up — not a bug, just ordering.** The
running pods are 89 minutes old (started ≈19:58 BST). My fix committed at
**20:35:33** — 37 minutes AFTER these pods started. Whatever build the user
saw deployed was built from a HEAD that necessarily predates my commit; it
cannot contain code that did not exist yet when it was built. **`f8cbaf551`
has never been built into any chassis image.** `makefile` still reads
`IMAGE_TAG ?= v1.0.1268` — unbumped since before this fix.

**Not yet done, in order, for whoever continues this:**
1. Bump `IMAGE_TAG` in `makefile` (currently `v1.0.1268`) — a same-tag
   rebuild ships the stale cached binary.
2. `make build-agent-chassis` (builds from committed HEAD — my fix is
   committed, so this is safe; check for "leaving out N uncommitted changes"
   in the build output, should be none of mine).
3. Push + deploy per the makefile's own `push-agent-chassis`/
   `deploy-agent-chassis` targets (or whatever the owner's usual release
   command is — `releases are WHOLE-FLEET, owner runs make release` is a
   standing note in memory; check whether this needs to go through that
   route rather than a one-service deploy).
4. **Re-verify at the artefact**, both replicas, same grep as above — now
   expect `negated declaration ignored` ≥0 occurrences is not the test (it's
   conditional on the string being reachable in a live run); the load-bearing
   check is `NegationGuard` / `fabNegationCueRe`-shaped literals present
   (count ≥1), same command as above.
5. Behavioural check: re-run the portfolio recreation item (or dispatch a
   fresh one) and read `check_fabrication` in `orchestration_states.collected_data`
   same-day (it purges) — expect `fabricated:false`.
6. Tell `mortgagecalculator_couk_adoption` (their handoff docs) the workaround
   clause can come out.
7. Annotate `016b` §9's existing entry for this bug (deferred deliberately
   until live, per the PLAN §10) — do NOT do this before step 4 confirms live.
8. Re-check the council verdict — round 1 was blocked fleet-wide on Anthropic
   credit exhaustion (`SUBMISSION_CORR=aa2d0d62-4aba-480e-aedc-8be264d53b01`),
   not reviewed. Resubmit with `RESUBMIT_CORR=aa2d0d62-…` once credits are
   restored (check: has any council/LLM call succeeded fleet-wide in the last
   15 minutes?). If APPROVED lands, add `Council-Reviewed: <id>` — but only
   after reading the verdict, never speculatively.
