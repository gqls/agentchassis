# 243 — tool-acceptance's `look` step has no storage client on EITHER execution path, so the vision half of acceptance has never run

**Filed 2026-08-10** by `staged_component_build`, at the owner's direction, after the
batch-8 `tool-setup-builder` S6 run surfaced it. Symptom first recorded in that lane's
NOTES (`## 2026-08-10 (second session, later)`) and the 08-10 handoff's defect item 9.

**Status: OPEN. Diagnosed first-hand, fix candidates below, nothing committed as a fix.**

## Declared substitute for the 090 loop (owner ruling 2026-07-31)

This file asserts a root cause without a 090 run. The substitute: the owner directly named
the mechanism to check ("storage is injected at spawn time if the type is listed in the
spawn action code") and every link in the chain below was then read first-hand from the
live code, the live DB and the running pods, with the citation inline at each step. The
chain has no inferred link: the failing line, the nil's origin, the env gate, the spawn
list, the live category row, the per-run `processing_node`, the ConfigMap key→env mapping,
and the deliberate absence of the env on the standing chassis are each individually
quoted. The one inference I did make mid-investigation — "the runs execute inline on the
standing chassis" generalised from one run's shared `owner_agent_id` — was WRONG for 20 of
26 runs and was corrected by reading `processing_node`, which is the column the overlay
comment says to read before concluding anything about which env matters.

## Symptom (measured, could have come out otherwise)

Every `tool-acceptance-agent` orchestration in the retained window ends
`complete_no_look` with the identical `__step_error`:

```
step look failed: failed to execute action execute_vision_prompt:
execute_vision_prompt: no storage client — cannot download screenshots
```

Grouped over all rows (`owner_agent_type='tool-acceptance-agent'`): **26 of 26**
`complete_no_look` carry that message; the only other row is 1 `complete_error` from an
unrelated `request_run` timeout. The grouping by `__step_error` is what distinguishes
"uniformly broken" from "a designed skip branch" — the bare `complete_no_look` count
cannot. Retention bounds the claim: earliest retained row 2026-08-09, so "every run in the
retained window", not "every run ever". (`site_work_items` shows 51 complete
`acceptance_run` items all-history, so the true failure count is likely ~2× the retained.)

**What is NOT lost:** the check results. They are produced by `request_run` and land in
`browser_run.response` before `look` runs, and the judge acts on them. What is lost is the
screenshot/vision pass — the half that catches what a selector cannot see (a tool that is
present, correct and invisible).

## Root cause — one gate, two execution paths that both fail it

`execute_vision_prompt` hard-errors when `params.StorageClient == nil`
(`platform/orchestration/actions/execute_vision_prompt_action.go:87`). That client is the
agent's own: `agentbase/agent.go:335` hands `a.storageClient` to the coordinator
(`coordinator.go:101`, then into every action at `coordinator.go:1703`), and
`agentbase/agent.go:316` builds it **only when `IMAGE_BUCKET` is set** — otherwise it logs
"Storage client not configured (IMAGE_BUCKET not set)" and leaves it nil.

`orchestration_states.processing_node` splits the 26 failing runs across two environments,
and **neither has `IMAGE_BUCKET`**:

| path | runs | why the env is absent |
|---|---|---|
| spawned `agent-tool-acceptance-agent-<hash>` pods (the overnight `acceptance_run` due-sweep) | **20** | spawn injects storage env only when `isStorageEnabledAgent(type) \|\| category ∈ {orchestrator, code-driven}` (`spawn_actions.go:2556`). `tool-acceptance-agent` is **not** in the 12-name `storageAgents` list (`spawn_actions.go:3040-3053`), and its live `agent_definitions.category` is **`tools`** — so no branch fires and nothing storage-shaped is injected |
| standing `agent-chassis-*` pods (manual kcat dispatches on the generic topic, e.g. `tool_acceptance_run.sh`) | **6** | the deployment **deliberately** carries no `IMAGE_BUCKET`/`S3_ENDPOINT` — owner ruling 2026-08-08, recorded in `deployments/kustomize/services/agent-chassis/overlays/production/uk_001/patch-deployment.yaml` (~line 100): an earlier revision added them (`820a033c0`, 2026-08-05) and was reverted. The B2 credentials ARE present there (older, `019cf8d94`, explicitly not part of that revert); it is the bucket/endpoint config that is absent, and `agentbase` gates on the bucket |

Supporting facts, each checked:

- The spawn block's ConfigMap mapping is complete and correct for the listed types:
  `storage-config` keys `S3-ENDPOINT`/`S3-REGION`/`image_bucket`/`assets_bucket`/
  `S3_USE_PATH_STYLE` map to env names `S3_ENDPOINT`/`S3_REGION`/`IMAGE_BUCKET`/
  `ASSETS_BUCKET`/`S3_USE_PATH_STYLE` (`spawn_actions.go:2588-2634`), plus AWS/B2
  credentials as direct values (`spawn_actions.go:2568-2580`). So membership of the list
  is sufficient — a listed type gets everything `agentbase` needs.
- **The screenshots exist.** The browser-runner adapter has its own client and uploads
  fine — my run's `browser_run.response.renders` names
  `s3://personae-prod-uk001-images/acceptance-evidence/<site>/tool-setup-builder/…_desktop.png`
  and `…_mobile.png`. The orchestration-side download is the only missing piece.
- **Blast radius is exactly this agent.** A `default_config` scan over live
  `agent_definitions` finds `tool-acceptance-agent` is the **only** active agent whose
  workflow uses `execute_vision_prompt`. render-audit's vision runs adapter-side and its
  recent orchestrations complete with no step error on the same chassis pods.
- The overlay comment's reason 2 ("agents do not execute in this deployment") is true for
  the sweep and **false for manual dispatches** — 6 of 26 runs, including both of this
  lane's S6 runs, executed inline on standing chassis pods. Its own instruction ("read
  `processing_node` before concluding anything about which env matters") is what caught
  this; the comment should not be weakened, but its reason-2 wording overstates.

## Fix candidates, ordered by what closes the door

1. **Add `"tool-acceptance-agent"` to `storageAgents` (`spawn_actions.go:3040`).** One
   line. Fixes the scheduled sweep — 20 of the 26 observed runs, and 100% of the
   unattended path, which is the one that matters "for ever". Consistent with BOTH halves
   of the 2026-08-08 ruling: the list is the sanctioned mechanism for granting storage to
   a specific spawned type, rather than spreading env across the whole chassis
   deployment. Additive, opt-in, changes no shared guarantee → normal council gate, no
   RFC (2026-07-29 ruling). Verify after the next roll by the spawn log line
   `Injecting storage credentials {agent_type: tool-acceptance-agent}` and a sweep run
   reaching `complete` (not `complete_no_look`), with `__step_error` empty.
2. **The manual path needs an owner decision, and candidate 1 does not touch it.** kcat
   dispatches on the generic topic run inline on the standing chassis, which stays
   bucket-less under the 08-08 ruling. Options: (a) accept that manual runs keep losing
   the `look` step (check half still lands — it is how all of this lane's S6 greens were
   read); (b) reshape the manual trigger to go through the spawn path like the sweep;
   (c) reopen the 08-08 ruling for the chassis deployment. (a) costs nothing today and
   leaves a knowingly-degraded path; do not choose it silently.
3. **Make the degradation visible either way**: `complete_no_look` reads as COMPLETED and
   nobody noticed 26 consecutive losses. Whatever else is done, the `look` failure should
   mark the run's summary (e.g. `vision: skipped(<reason>)`) so a green with no vision is
   distinguishable from a green with vision — this is the same "a silent gate either did
   not look or approved" shape the memory index already records.

## How to verify (whoever fixes)

```sql
-- before: every retained sweep run says complete_no_look + the storage message
SELECT processing_node, current_step,
       collected_data->'__step_error'->>'message'
  FROM orchestration_states
 WHERE owner_agent_type='tool-acceptance-agent'
 ORDER BY created_at DESC LIMIT 5;
```

After candidate 1 rolls: dispatch one acceptance run (the due-sweep raises them
overnight, or `tool_acceptance_run.sh` — but note a MANUAL run tests the manual path,
which candidate 1 does not fix; to test the fix you need a SPAWNED run). Expect
`current_step='complete'`, no `__step_error`, and the vision verdict populated. The spawn
env can be checked directly on the pod while it lives:
`kubectl exec <agent-tool-acceptance-agent-…> -- env | grep -E "IMAGE_BUCKET|S3_ENDPOINT"`.

---

## UPDATE 2026-08-10 (same day, evening) — candidate 1 IMPLEMENTED, SUBMITTED, COMMITTED; still OPEN pending roll + spawned-run proof

Owner directed candidate 1 ("please implement fix candidate 1 and put it through the
council gate"). Done:

- **Code**: `tool-acceptance-agent` added to `storageAgents` (`spawn_actions.go`), with
  the why-comment citing this bug. Commit `543206039`, trailer
  `Council-Submitted: 5eb4ad58-b873-4d6a-b61e-9cef1cbe4372`. HEAD verified to build from
  a clean `git archive` after the commit.
- **Council**: submitted pre-commit, correlation `5eb4ad58-b873-4d6a-b61e-9cef1cbe4372`.
  The submission names the tension with the MDL-040 register entry's strict reading of
  the 2026-08-08 ruling and the 2026-08-10 owner direction that resolves it; verdict to
  be read and acted on (~30 min budget, find the run by payload not printout).
- **Sibling filed**: `bugs_open/245` — the owner's companion directive (standing chassis
  should not carry B2 credentials). Its candidate 1 (secretKeyRef conversion) must land
  BEFORE any credential removal, or this fix's injection path breaks for every storage
  agent — sequencing recorded in both files and in the submission's risks.

**Remains OPEN because the bar is fixed AND live**: the change is inert until the next
chassis image roll, and the proof must come from a SPAWNED run (`processing_node` naming
an `agent-tool-acceptance-agent-*` pod, `complete` not `complete_no_look`, and the
first-ever `llm_call_log` rows from the vision step). A manual run cannot prove it — it
exercises the inline path this fix deliberately does not touch. Candidates 2 (manual
path) and 3 (make the loss visible) remain undecided.

## UPDATE 2026-08-10 (later) — council verdict: APPROVED, round 1, and both advisory objections answered by query

`5eb4ad58` returned **APPROVED** ("2 advisory objection(s) — none high-severity"; 11 seats
reviewed, 6 abstained, no truncation gating). The commit already carries
`Council-Submitted:`, which 098 resolves to this approval at report time — no further
trailer action (forward-only forbids the amend anyway).

Both objections asked for the same discipline — scope claims confirmed by query, not
prose — and both were run immediately (2026-08-10 evening):

1. **prior_art_librarian (medium): the two-active-rows trap** — could `category='tools'`
   have been read from a stale lower version? **No**: `agent_definitions` holds exactly
   ONE row for `tool-acceptance-agent` (version 1, category `tools`, active, not
   snapshot, not deleted, config references `execute_vision_prompt`).
2. **guardian (medium): the only-consumer claim** — re-measured with NO liveness filter:
   across all rows in any state (deleted/snapshot included), `tool-acceptance-agent` is
   the only type whose `default_config` mentions `execute_vision_prompt`. The fix
   under-covers nothing.

The guardian also put on record (low) that this is another touch to `spawn_actions.go`
under the stability preference — noted, and the reason the edit stayed a pure allow-list
append.

**Still OPEN**: awaiting the next chassis roll, then the SPAWNED-run proof (§ above).

## UPDATE 2026-08-10 (night) — v1.0.1283 rolled; the spawned-run proof is still OWED

Fleet rolled to **v1.0.1283** (chassis pods up 21:43Z), built after the fix commit
(18:16Z) — but this change added **no unique string literal** (`"tool-acceptance-agent"`
already existed in the binary from other call sites, and Go dedupes rodata), and the
binary carries no VCS stamp, **so a pod-grep cannot prove this fix shipped.** The proof
is behavioural and has not happened yet: no `tool-acceptance-agent` orchestration has
been created since the roll. The 8 spawned runs at 19:05–19:16Z all pre-date it (old
image) and correctly still read `complete_no_look`.

**Next session: check the overnight sweep's runs** —
```sql
SELECT correlation_id, processing_node, current_step,
       collected_data->'__step_error'->>'message'
  FROM orchestration_states
 WHERE owner_agent_type='tool-acceptance-agent' AND created_at > '2026-08-10 21:43+00'
 ORDER BY created_at;
```
PASS = a run on an `agent-tool-acceptance-agent-*` node reaching `complete` with no step
error, plus the first-ever vision `llm_call_log` rows. `complete_no_look` on a POST-roll
spawned run = the fix did not ship or did not work — re-open loudly.

## UPDATE 2026-08-11 — CANDIDATE 1 PROVEN ON A SPAWNED RUN. The vision half ran for the first time ever — and found a real defect on its first look.

The overnight sweep could never have produced this proof: `check_tool_acceptance_due.go:92-102`
suppresses any tool with an acceptance verdict in the last 7 days, and every batch-8 tool ran
on 08-10. So the proof was driven: work item `ae33ed59-9a43-49b3-ae05-3a8a6177aa27`
(`acceptance_run:tool-setup-builder:5fe8785b…`, dartsonline — this lane's own subject, raised
09:40Z mirroring the A4 items' shape), claimed by `build-dispatch-loop` within a minute.

**Every PASS criterion met, on chassis v1.0.1284:**

- Run `0ee53904-4c9f-475e-ab93-c2252c4e6a9d` on **`agent-tool-acceptance-agent-649a6c11-q9mlk`**
  (a SPAWNED node) reached **`complete`** — not `complete_no_look` — with **no `__step_error`**.
  Checks 15/0/9, matching the S6 greens.
- **Pod env captured while it lived**: `IMAGE_BUCKET`, `ASSETS_BUCKET`, `S3_ENDPOINT`,
  `S3_REGION`, `S3_USE_PATH_STYLE` present via `configMapKeyRef: storage-config` — the
  `storageAgents` injection fired for this type for the first time. (Credentials arrived as
  `secretKeyRef`, which is `bugs_open/245`'s half of the same spawn block.)
- **The first-ever `llm_call_log` rows for `tool-acceptance-agent`** (0 all-history before):
  step `look`, provider `anthropic`, model `claude-sonnet-5`, success, 2 images sent /
  0 dropped, 5.7s. This also finally exercises MDL-040's provider path for vision.

**And the opened eye saw something.** Its very first result reports a genuine visual defect on
the dartsonline setup-builder page: several form options ("Beginner", "Smooth and fluid",
"Pinch grip") and the "Get my recommendation" button render near-invisible against their dark
backgrounds, consistently on desktop AND mobile — precisely the "present, correct and
invisible" class the check half cannot see. Every selector check passed while this was true.

**Candidate 3's case just made itself**: the run completed green, raised nothing, and that
vision finding is recorded only inside `collected_data->'look'` where nobody looks. A vision
result that names a defect should mark the verdict or raise something visible — as filed above.

**Status: candidate 1 is FIXED + LIVE + PROVEN. The file stays in `bugs_open/` (owner practice
2026-08-06) because candidates 2 and 3 are still undecided owner calls:**
- **Candidate 2** — the manual/inline path still loses the vision half by design (08-08 ruling
  keeps the standing chassis bucket-less). Options (a) accept and say so / (b) reshape the
  manual trigger to the spawn path / (c) reopen the ruling.
- **Candidate 3** — make a vision-skip or vision-finding visible in the verdict rather than
  silent (today's run is the worked example: found a defect, run reads green, nothing raised).

## UPDATE 2026-08-11 (afternoon) — OWNER DECIDED both remaining candidates; candidate 2 (option b) is DONE and PROVEN

Owner, in chat 2026-08-11: candidate 3 **YES — build it** (spec and constraints:
`staged_component_build/HANDOFF_2026-08-11_continue_here.md` §3b item 1; the parallel
session's measurement that `render-critique` has NO consumer raises its priority — the
restored eyes currently write to a channel nothing reads). Candidate 2: **option (b),
reshape the trigger with an orchestrator wrapper.**

Candidate 2b implemented same-day: `docs/leopardessconsulting/scripts/tool_acceptance_run.sh`
rewritten — it no longer kcats the generic topic (the inline path this bug documented);
it inserts the due-sweep's own `acceptance_run` work item and `build-dispatch-loop`
spawns the pod, so a manual run now exercises BOTH halves. Preflights refuse the two
quiet no-op modes (unresolvable page, missing PLAN) and a duplicate open item, loudly,
at submit time. RUNBOOK §10 box has the operational notes (async, ~3–10 min; per-site
dispatch rotation; the 7-day cooldown gates only the sweep).

**Proven end-to-end on its first real run**: work item `4ef3c11a…` → a SECOND
independently spawned pod (`agent-tool-acceptance-agent-d3a4a56a-vtw9d`) → `complete`,
no step error, 15/0/9, vision ran (2 images; `llm_call_log` now 2 rows all-history, one
per spawned run). One capture bug found and fixed by foreground-testing the script:
psql `-t -A` prints the `INSERT` command tag after a `RETURNING` row, so the printed
follow-queries carried a two-line id until `head -1` was applied.

**Remaining OPEN on this bug: candidate 3 only** (the build is specified and decided,
not yet written). The inline path is now unreachable from any documented trigger; a
hand-built kcat dispatch would still lose the vision half — that residue is accepted
and documented rather than defended.
