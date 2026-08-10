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
