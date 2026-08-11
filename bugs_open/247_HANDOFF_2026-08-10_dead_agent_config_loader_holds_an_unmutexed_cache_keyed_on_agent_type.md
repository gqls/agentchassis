# 247 — the dead `AgentConfigLoader` holds an unmutexed map cache keyed only on agent type, and returns a hard-coded two-step workflow

**Filed 2026-08-10** by the `bugfix_239` lane, found while ruling out "a workflow cache
keyed coarser than the message" as the cause of `bugs_open/239`. **Status: OPEN, latent.**
It is not biting production today, and the reason it is filed anyway is that it is a loaded
gun with a plausible-looking safety catch.

## What is there

`platform/config/agent_config_loader.go` holds:

- `cache map[string]*models.AgentConfig` **keyed on `agentType` alone**, with **no mutex**
  (:20, :152, :179);
- a `Workflow` it synthesises rather than loads: a hard-coded `process → complete` plan.

Its only reader is `MessageProcessor.processRequest` (`platform/messaging/processor.go`),
and **`processRequest` has no callers** — verified fleet-wide 2026-08-10.

## Why a dead function earns a bug file

1. **It looks like the live path.** A session tracing "where does the chassis load an
   agent's workflow?" finds a well-named `AgentConfigLoader` with a cache, and can spend
   real time reasoning about cache invalidation in a mechanism that never runs. This lane
   did exactly that: "a workflow-plan cache keyed on something coarser than the full
   message" was `bugs_open/239`'s own leading hypothesis, and it was wrong — the live path
   (`loadAgentDefinition` + `FindByType`) hits Postgres on **every** message and caches
   nothing.
2. **The day it acquires a caller it is a data race**, `go test -race` clean until then
   because nothing exercises it. Concurrent `processMessage` goroutines share one
   processor.
3. **Its hard-coded workflow is the same failure shape as `bugs_open/239`**: a synthetic
   plan substituted for the agent's real one, indistinguishable downstream from a real
   resolution.

## Fix candidates, ordered by what closes the door

1. **Delete `processRequest` and the loader.** Nothing calls either; the live path does not
   need them; deletion makes the bad state unrepresentable and removes the misleading
   signpost. Check first whether `AgentConfigLoader` has readers outside `processRequest`
   (the constructor is wired into `MessageProcessor` at :101, so the field assignment must
   go too).
2. Keep it, add a mutex, and key the cache on something that can identify a version.
   Strictly worse: it preserves a second, hard-coded answer to "what is this agent's
   workflow?" — and the estate's rule is one implementation per judgement.
3. Leave it with a comment saying it is dead. Rejected, listed for completeness: the file
   already reads as live to anyone who does not run the grep.

## Check before deleting

```
grep -rn "AgentConfigLoader\|processRequest\|configLoader" --include=*.go .
```
Confirm the callers are still zero (this is a shared tree; another session may have wired
it up since 2026-08-10) before removing anything.

---

## Appended 2026-08-10 — a second dead twin in the same class, found by the pre-commit pattern check

`platform/messaging/processor.go` also carries **`selectWorkflowOLD`** (:1325) and
**`processRequest`**, both with **zero callers**, and `selectWorkflowOLD` carries the
*exact defect* `bugs_open/239` just fixed in its live sibling: the same
parse-failure-skips-everything structure falling through to the consuming agent's own
`DefaultConfig["workflow"]`.

This is why the class is worth a bug rather than a shrug. The pattern check flagged
`selectWorkflow` changed but `selectWorkflowOLD` not — correctly, and by the reasoning in
016b §9 #26: *"if the change is a fix, the twin probably has the same defect"*. It does.
The twin was left unfixed **deliberately**, because fixing dead code preserves the thing
that makes it dangerous: a second, plausible-looking answer to "how does this chassis pick
a workflow?", now with the fix applied so it looks *more* current, not less.

**The remedy for all three (`AgentConfigLoader`, `processRequest`, `selectWorkflowOLD`) is
the same: delete them**, with the caller-count grep re-run first on this shared tree.
Whoever takes this should treat them as one task.

---

## Appended 2026-08-11 — fix committed, awaiting roll (Status: FIXED, not yet live)

Picked up by a fresh thread (workstream:
`docs/agent_docs/docs024_key_docs_latest/bugfix_247_dead_agent_config_loader/`), unowned per
`scripts/who-owns.py 247` and independently corroborated still-dead by
`architecture_review/RFC_023`, which cites `selectWorkflowOLD` as already-dead in an
unrelated argument. Re-verified every zero-caller claim fresh via grep immediately before
editing (shared tree).

Deleted, per fix candidate 1 above: `AgentConfigLoader.LoadAgentConfig` + its unmutexed
`cache` field (`platform/config/agent_config_loader.go`); `MessageProcessor.processRequest`,
`selectWorkflowOLD`, `determineWorkflowModeOLD`, `isComplexRequest`, `getDefaultTaskWorkflow`,
and the EOF `determineWorkflowMode` (a fifth dead function found by the pattern-check
appendix's own sweep logic, pulled into this same deletion since it was only reachable from
the doomed cluster), plus the now-dead `configLoader` field/constructor-wiring/import in
`platform/messaging/processor.go`. Left `AgentConfigLoader` itself and its other methods
untouched (`internal/agents/contentcreator/agent.go` is a live external caller); left the
always-nil `db *sql.DB` field in the same struct untouched too — same dead-weight class, not
named by this bug, out of scope for this diff.

`go build ./...`, `go vet`, `go test`, and `go test -race` all clean on
`platform/messaging`, `platform/config`, `internal/agents/contentcreator`. Final grep for
every deleted symbol name returns zero hits fleet-wide (the only near-miss,
`platform/agentbase/agent.go`'s `processRequests`, is a distinct, live, plural function,
confirmed unrelated and untouched).

Submitted to the council gate: `SUBMISSION_CORR=cf96f869-8b48-45fb-98bb-081b0f87df1c`.
Committed with `Council-Submitted:` per the 2026-07-30 norm rather than holding code for the
verdict. **Per the owner's 2026-08-06 ruling this bug stays in `bugs_open/` even once the
fix is live** — the bar for moving to `bugs_open/` → closed is fixed AND live, and a Go
change is inert until the next image build + roll. Not marking closed here; the owning
workstream's NOTES file will record the post-roll proof.

Also noted, deliberately out of scope for this commit: `sendSuccessResponseOLD` and
`sendErrorResponseOLD` in `processor.go` are the same dead-`OLD`-twin class, zero callers,
not named by this bug file. Worth a follow-up filing so the class gets swept.
