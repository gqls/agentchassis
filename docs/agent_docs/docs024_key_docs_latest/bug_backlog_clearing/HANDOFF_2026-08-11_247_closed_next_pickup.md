# HANDOFF 2026-08-11 — 247 closed, proven live; start a fresh pickup here

Continues this series (`HANDOFF_2026-08-07_both_done_next_pickup.md` and earlier). Standing
owner brief for the whole "find the next unowned bug and fix it" task is still
`HANDOFF_2026-08-05_next_bug_pickup.md` — re-read it, it is the job description. This entry
exists because the previous session's context grew long (heavy verification: fable-model
planning, council submission, and a from-scratch build-provenance investigation) and it
handed off rather than risk sloppier work on a second bug in the same thread.

## `bugs_open/247`: DONE — fixed, council-approved, proven live, closed in place

**The bug:** `platform/config/agent_config_loader.go`'s `AgentConfigLoader.LoadAgentConfig`
held an unmutexed `map[string]*models.AgentConfig` cache — a data race waiting for a second
caller — and its only caller, `MessageProcessor.processRequest`
(`platform/messaging/processor.go`), was itself dead code, along with a whole connected
dead-twin cluster (`selectWorkflowOLD`, `determineWorkflowModeOLD`, `isComplexRequest`,
`getDefaultTaskWorkflow`, `determineWorkflowMode`). `selectWorkflowOLD` additionally carried
a defect (`bugs_open/239`'s) already fixed in its live sibling `selectWorkflow`.

**The fix:** pure deletion, 205 lines, 2 files (`processor.go`, `agent_config_loader.go`).
`AgentConfigLoader` the *type* stays live (real external caller:
`internal/agents/contentcreator/agent.go`) — only the dead cache method and the fully-dead
`processRequest`/`selectWorkflowOLD` cluster came out. `go build`/`vet`/`test`/`test -race`
all clean. Council: **APPROVED, round 1, all reviewers**
(`cf96f869-8b48-45fb-98bb-081b0f87df1c`).

**Commits:** `8cb8938bb` (the fix), `97207ef9c` (ratchet-file quieting), `a78853b52` (verdict
recorded), `e44b49578` (closed in place with post-roll proof).

**Proven live 2026-08-11** on a fresh whole-fleet chassis build: `agent-chassis` is the only
service whose binary actually contains the changed code (traced via import graph:
`platform/agentbase` → `platform/messaging`, and only `cmd/agent-chassis` imports
`platform/agentbase`). Both running pods share one imageID digest (no straddled release).
Found the exact build commit via the deployed image's own OCI label rather than the
scrolled startup log or the exact-match-only binary grep:
```
docker image inspect docker.io/aqls/agent-chassis:v1.0.1288 \
  --format '{{index .Config.Labels "org.opencontainers.image.revision"}}'
# -> bb534864249117003ac758e50adc0df9176ef370
git merge-base --is-ancestor 8cb8938bb bb534864... && echo LIVE
```
This is a reusable trick worth keeping: **if the image is already cached locally** (it was
here — `make release` had just run), reading the OCI revision label is faster and more
precise than either the log-tail method (rotates on busy services, `bugs_open/153`'s known
gap) or the binary-grep method (only matches an *exact* stamped commit, useless once even
one later commit lands on the same file before the build cuts — which happened here: two
more `processor.go` commits, both bug 246, landed between mine and the release).

**File stays in `bugs_open/`** — owner direction 2026-08-06 overrides CLAUDE.md's
`/bugs_closed/` bar; closure evidence is written inside the file. Do not `git mv` it.

Full account: `bugs_open/247_HANDOFF_2026-08-10_dead_agent_config_loader_holds_an_unmutexed_cache_keyed_on_agent_type.md`,
workstream docs at `docs/agent_docs/docs024_key_docs_latest/bugfix_247_dead_agent_config_loader/`
(PLAN, RUNBOOK, NOTES, README, SUMMARY — all current).

## Next pickup, mechanically (unchanged from the last entry in this series)

1. `ls bugs_open/` — as of 2026-08-11 the newest is `253`. **The 238–253 range is almost
   entirely claimed** — spot-checked via `who-owns.py` during this session:
   `244`→`bugfix_168_deployed_asset_path`, `243` (tool-acceptance one)→`staged_component_build`,
   `241`→`loancalculator_couk`, `240`→`bugfix_209_deploy_purpose_keyed_source`,
   `248` (undeployed-asset one)→`bugfix_203_phantom_cta_cleanup`. **This will already be
   stale** — 35 peer sessions were active concurrently at last check (`ListAgents`); re-run
   the checks, don't trust this list.
2. Given the recent range is saturated, also check the **older backlog** this session did not
   reach: `029`, `033`, `040`, `071`, `083`, `085`, `093`, `096`, `107`, `113`–`161` — several
   are OPEN per `bugs_sweep`/`bug_backlog_clearing`'s own prior notes but may since be picked
   up or fixed-pending-roll. Same check applies.
3. Standing four before touching anything: `scripts/who-owns.py <n|slug>` (resolve number
   collisions by SLUG — several numbers, including `243` and `248` right now, name two
   unrelated bugs), `git log` the file the bug's own text names, grep the live `.jsonl`
   transcripts under `~/.claude/projects/-home-ant-projects-agentchassis/` for the bug number
   AND the code path (recently-modified files only — `ls -la --time-style=full-iso *.jsonl |
   sort -r`), and check `site_work_items` for open work on the target.
4. Re-verify the defect against the live system before planning — grep the actual zero-caller
   /reproduction claim fresh, don't trust the bug file's own text as current.
5. Then the brief as written: **use the fable model** to prepare the fix plan (`Agent` tool,
   `model: "fable"`, read-only, get exact line ranges — see this session's transcript /
   `bugfix_247`'s `NOTES` for a worked prompt template), implement, build+test, submit to the
   council gate (platform/internal/pkg only — `097_TRIGGER_council_review_v1.sh`), commit per
   task with an explicit pathspec, keep the five standing docs current, missteps to
   `WRONG_CALLS.md`.
6. **If a fresh chassis build lands while you're mid-task**, don't assume your fix is in it —
   check per-service (trace the import graph first to find which of the 14 backend binaries
   actually compile the changed file) and per-pod (all replicas share one imageID), using the
   OCI-label trick above if the image is cached locally, falling back to
   `EXPECT_SHA=<sha> grep -aq "$sha" /proc/1/exe` (matches only the *exact* build commit,
   not ancestors) or the startup log (`grep 'build provenance'`, unreliable once busy) if not.

## Method lessons from this arc (the transferable part)

1. **A dead-code deletion still needs the fable-planning step done carefully** — the obvious
   scope ("delete the two named functions") undershot the real dead subgraph by three more
   functions (`determineWorkflowModeOLD`, `isComplexRequest`, `getDefaultTaskWorkflow`, plus
   the EOF `determineWorkflowMode`), all only reachable from the doomed cluster. A
   line-numbered plan from a fresh read caught the connected-subgraph shape that the bug
   file's own two-function summary missed.
2. **A struct field is not dead just because its type looks dead** — `AgentConfigLoader` had
   one dead method and five live ones (a real external caller, `contentcreator/agent.go`).
   Checking "does the TYPE have callers" instead of "does THIS METHOD have callers" would
   have produced a wrong, too-narrow (or if reasoned about carelessly, too-wide and breaking)
   fix.
3. **`--update-ratchet` is a shared-tree trap for a single-task session.** It "accepts the
   current state as baseline", which silently absorbs every OTHER session's new workstream
   directory as your own commit's business. Hand-append your own one line instead.
4. **A recited procedural rule can be as wrong as a recited code fact.** Wrote "fixed AND
   live is the bar to move to bugs_closed/" from CLAUDE.md's literal text, twice, despite the
   override sitting in the auto-loaded memory index the whole session. Caught on a re-read,
   not a fresh check — worth deliberately grepping the memory index before writing any
   *procedural* claim into a durable file, the same discipline already used for code claims.
   Logged in `WRONG_CALLS.md` 2026-08-11.
5. **The binary-grep provenance recipe only proves an EXACT commit, never an ancestor.** If
   even one more commit lands on your file between your fix and the release cut (routine on
   this tree), `EXPECT_SHA=<your-sha> grep -aq ... /proc/1/exe` correctly returns NO MATCH
   and is easy to misread as "not shipped". Get the actual stamped commit first (image OCI
   label if cached locally; otherwise the startup log within its shelf life), THEN
   `git merge-base --is-ancestor <your-sha> <stamped-sha>`.
