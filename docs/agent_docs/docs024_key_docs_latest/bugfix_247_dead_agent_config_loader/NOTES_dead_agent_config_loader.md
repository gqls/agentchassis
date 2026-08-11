# NOTES — bugfix 247 (append-only, newest at the bottom)

## 2026-08-11 — picked up, validated, plan prepared

Ownership check (`scripts/who-owns.py 247`): no owning workstream, no workstream directory
existed. Live-session check: grepped active `.jsonl` transcripts under
`~/.claude/projects/-home-ant-projects-agentchassis/` for mentions of the bug file / target
symbols — several recent sessions were themselves running `who-owns.py` sweeps over the
240-253 range (i.e. other threads doing the same "find an unowned bug" pass this task
description asks for), but none showed a Write/Edit tool call against
`platform/config/agent_config_loader.go` or `platform/messaging/processor.go`. One session
(3402547c…) was mid-flight on bug 239 in the same file (`processor.go`) but at unrelated
line ranges (offsets 595, 1060, 215) and had already committed `7a29195c7` and reported 239
closed — read-only overlap, not a collision.

Re-validated the bug's claims fresh (not trusted from the file): grepped for
`.processRequest(`, `.selectWorkflowOLD(`, `.LoadAgentConfig(` call sites — all zero, only
the function definitions exist. Cross-checked independently against
`architecture_review/RFC_023_a_silent_success_becoming_a_loud_failure_is_a_delivery_guarantee_change.md`,
which — for an unrelated argument — states `selectWorkflowOLD` "is already one, dead,
recorded in bugs_open/247". Two independently-arrived-at confirmations of the same dead-code
fact is a good validity signal without needing a 090 diagnosis run (this is a self-evidencing
fix per CLAUDE.md's debugging-before-diagnosis section: grep proves the claim, the build
proves the fix).

Dispatched a fable-model agent (general-purpose, read-only) to prepare the implementation
plan per the task's instruction to "use fable" for planning. Full plan text below,
unedited from the agent's output, because it names exact line ranges the implementer will
re-verify at edit time — reproducing it here rather than only in chat keeps it durable.

<details>
<summary>Full fable-agent plan (2026-08-11)</summary>

Rationale, exact deletions (E1-E11 across processor.go and agent_config_loader.go), what
must not be touched, test-file impact (none), architecture-scope judgment (not in scope,
still council-gate it), and verification commands — see `PLAN_2026-08-11_dead_agent_config_loader.md`
for the condensed version. The agent's key findings, worth keeping verbatim:

- `processor.go`'s dead cluster is bigger than the bug file named: `selectWorkflowOLD`
  (:1336) pulls in `determineWorkflowModeOLD` (:1394), `isComplexRequest` (:1418, but this
  one has a SECOND dead caller — `determineWorkflowMode` at EOF :2343, itself dead, so both
  die together), and `getDefaultTaskWorkflow` (:1437). This is a bigger blast radius than
  "delete two named functions" — it's a connected dead subgraph.
- The `configLoader` field in `MessageProcessor` becomes fully dead once `processRequest`
  goes (its only reader), so the field, its constructor wiring, and the now-unused `config`
  package import in processor.go all come out together — confirmed by the fact that `go
  build` will fail on the unused import if this is missed, which doubles as a verification
  step.
- `AgentConfigLoader` the *type* stays: `internal/agents/contentcreator/agent.go` is a live
  external caller of `LoadFromDatabase`/`GetDefaultConfig`. Only `LoadAgentConfig` (the
  cache method) and the `cache` field die.
- Two more same-class dead twins exist but are NOT named by bug 247:
  `sendSuccessResponseOLD` (~:2097) and `sendErrorResponseOLD` (~:2242). Left out of this
  commit deliberately — noted here and in the bug file's close-out for a follow-up filing,
  per CLAUDE.md's "don't widen the commit beyond the bug's stated remedy" norm implicit in
  the pathspec-commit rule.

</details>

Next: implement per the plan, re-grepping every claimed line range fresh immediately before
editing (this is a shared tree, per CLAUDE.md).

## 2026-08-11 — implemented, verified, submitted to council

Re-ran the pre-edit go/no-go greps fresh immediately before editing (per the plan's own
warning). All matched the fable agent's plan exactly, including the exact line numbers and
even the file line counts (2365 / 181) — the tree had not moved since the plan was drawn up
minutes earlier.

Applied all 5 grouped edits (E1/E2/E4 in processor.go as one contiguous-block deletion each,
E5/E6/E7 field+wiring+import, E8/E9/E10 in agent_config_loader.go). Deliberately left out
E11 (the always-nil `db *sql.DB` field) — same dead-weight class but not named by the bug,
and not part of either named risk (race, signpost); keeping the diff scoped to what 247
actually documents. `git diff --stat`: 2 files, 205 deletions, 0 insertions, nothing else in
the shared tree swept in.

Verification: `go build ./...` clean, `go vet` clean on the three affected packages,
`go test` and `go test -race` both green on `platform/messaging` and
`internal/agents/contentcreator`. Final grep sweep for every deleted symbol name returns
zero hits anywhere except `platform/agentbase/agent.go`'s unrelated, live, plural
`processRequests` — confirming the word-boundary distinction held and nothing else was
caught in the blast radius.

Submitted to the council gate: `SUBMISSION_CORR=cf96f869-8b48-45fb-98bb-081b0f87df1c`,
orchestration `da90d4a2-212a-4d1e-9cb0-728d97a5f0df`. Queue was clear (LAG 0) at submission
time. Committing now with `Council-Submitted:` per the 2026-07-30 norm rather than holding
the code for the ~30-minute round; will read the verdict and update this file plus the
commit trailer (a fresh commit, not an amend — forward-only) once it lands.

Also hand-appended one line to `docs026_concept_register/102_coverage_ratchet.txt` for this
workstream's own directory (pure dead-code deletion, nothing new callable, nothing to
register) — deliberately did NOT run `--update-ratchet`, which would have swept in 11 other
sessions' new workstream directories that are none of this task's business to characterise.

## 2026-08-11 — verdict: APPROVED, round 1, all reviewers

Checked back sooner than the ~30-minute budget because the dispatch lane was clear (LAG 0)
at submission time — verdict landed in well under 5 minutes. Confirmed via both queries
CLAUDE.md names:

```
SELECT current_step, status FROM orchestration_states WHERE ... = 'complete_approved' | 'COMPLETED'
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts ... = 'approved' (2026-08-11 15:56:37 UTC)
```

Human-readable note (`doc_notes`, category `council-gate`): "COUNCIL GATE — APPROVED — all
reviewers approve (round 1)". No objections to answer.

Per CLAUDE.md: **no amend** (forward-only) — the existing `Council-Submitted:` commit
(`8cb8938bb`) is credited automatically by the `098` coverage report at report time once it
resolves this correlation as approved. Not writing a fresh `Council-Reviewed:` commit on top
of a no-op change; the trailer rule exists to make the commit<->verdict join correct, and a
second commit with no code change would just be noise. Recording the verdict here is the
durable evidence trail.

**Still owed:** the fix is committed and approved but not yet LIVE — it ships on the next
image build + roll for `agent-chassis`/whichever services embed `platform/messaging` and
`platform/config`. Per the owner's 2026-08-06 ruling, `bugs_open/247` stays in `bugs_open/`
until proven live post-roll, even though it is fixed in substance. A future session (or this
one, later) should check the build-provenance stamp on the rolled service and append a
post-roll proof section to `bugs_open/247`, then move it to `bugs_closed/`.
