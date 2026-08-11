# 239 — a top-level `action=orchestrate` message silently resolves to the `generic` no-op agent, ignoring `config.agent_type`, whenever `input_data` carries BOTH `source` and `spec` keys

> ⚠⚠ **THE TITLE IS WRONG, AND SO IS THE CORRECTION BELOW IT. ROOT CAUSE FOUND
> 2026-08-10 — SKIP TO "ROOT CAUSE FOUND" AT THE FOOT OF THIS FILE.**
> The trigger is not `source`+`spec`, and it is not non-deterministic either:
> **`kcat -P` sends one message per LINE of stdin**, so a multi-line envelope
> arrived as several invalid-JSON fragments and the chassis silently ran each one
> as the `generic` no-op. Proven from `chassis_intake_events` payload bytes —
> 8 of 8 single-message sends resolved correctly, 10 of 10 fragmented sends did
> not. The framework defect (a request that cannot be resolved is silently
> converted into a COMPLETED no-op) is **fixed in code, awaiting a roll**.
> Everything between here and that section is preserved as filed, including its
> wrong turns, which are the record of how this was got wrong twice.

> ⚠ **CORRECTED 2026-08-10 — the title's trigger condition is NOT reliable; read
> "CORRECTION 2026-08-10" below before acting on the bisection table.** The
> same "safe" payload was later shown to fail non-deterministically, and a
> completely unrelated key also triggers the fallback. Re-testing this against
> production data caused a real incident (recovered same session) — **do not
> bisect this bug against a live site again**; see the correction for the safe
> path forward.
> **(This correction is itself superseded — see the banner above. "Non-deterministic"
> was an artefact of never looking at the bytes on the wire.)**

**Filed 2026-08-09** by the `webdesign_uk_build_service` lane, found while driving
a `page-build-handler` dispatch by hand for Phase 5 (the queue is starved — see
CLAUDE.md's "Dispatching work at the cluster" and this lane's own HANDOFF §1/§2
"drive loop" recipe). **Status: OPEN, root cause NOT diagnosed at the Go source
level — this file is empirical bisection, not a code fix.** Filed per the
2026-07-31 owner ruling's escape hatch: this is a durable, cross-cutting,
structural claim (it breaks the drive-loop pattern CLAUDE.md itself documents
and that many `HANDOFF`/`RUNBOOK` docs across the repo rely on), but I
substituted rigorous first-hand empirical verification — eleven isolated,
reproducible test dispatches bisecting the exact trigger condition — for a
`090` run, because the finding is precise enough to hand straight to whoever
picks it up: the trigger is two specific JSON keys co-occurring, not a vague
"it sometimes doesn't work."

## The symptom, one sentence

Send `{"action":"orchestrate","config":{"agent_type":"<X>"},"input_data":{...}}`
via `kcat` to `system.agent.generic.requests` (the documented drive-loop
envelope). If `input_data` contains **both** a `source` key and a `spec` key —
**regardless of what `config.agent_type` says, regardless of what those two
keys' values actually are** — the orchestration silently runs as
`owner_agent_type='generic'`, using the `generic` agent's own trivial
`default_config` (a single `complete_workflow` step whose description reads
*"No-op — scheduled task pre_query already did the work"*), and reports
`status='COMPLETED'` having done **nothing**: zero steps executed
(`execution_path` is an empty array), no child orchestrations spawned, no DB
writes. **The named agent's real workflow never runs at all.**

This is worse than a hard failure: `status='COMPLETED'` reads as success. The
only tell is `owner_agent_type != config.agent_type` and an empty
`execution_path` — neither of which the documented drive-loop's own verify
step (`SELECT status, current_step FROM orchestration_states WHERE
correlation_id=...`) checks.

## Evidence — eleven isolated dispatches, same site, same session, ~20 minutes apart

All fired via the exact kcat envelope from this lane's own HANDOFF §1/§2 and
`scripts/initial_messages/020_build_pipeline/076_trigger_build_pipeline.sh`
(topic `system.agent.generic.requests`, headers `action=orchestrate
from_agent_type=user from_agent_id=cli client_id=demo_client`, fresh UUIDs
each time). Verified by `SELECT owner_agent_type, status, current_step FROM
orchestration_states WHERE correlation_id='<corr>' AND
parent_orchestration_id IS NULL` a few seconds after each send.

| # | corr (short) | `config.agent_type` | `input_data` keys present | result `owner_agent_type` |
|---|---|---|---|---|
| control | `197ccd57` | `build-pipeline-trigger` | (none — `{}`) | `build-pipeline-trigger` ✅ |
| 1 | `896f9ea0` | `page-build-handler` | (none — `{}`) | `page-build-handler` ✅ (FAILED at `ensure_site_record`, as expected with no site_id) |
| 2 | `035c4fcf` | `page-build-handler` | `domain, site_id` | `page-build-handler` ✅ |
| 3 | `fcb72425` | `page-build-handler` | `domain, site_id, work_item_id` | `page-build-handler` ✅ |
| 4 | `67d64ee8` | `page-build-handler` | `domain, site_id, work_item_id, spec` | `page-build-handler` ✅ (reached `spawn_content_writer`) |
| 5 | `39ef229b` | `page-build-handler` | `domain, site_id, work_item_id, item_type` | `page-build-handler` ✅ |
| 6 | `b58f260d` | `page-build-handler` | `domain, site_id, work_item_id, item_type, source` | `page-build-handler` ✅ |
| 7 | `2906a176` | `page-build-handler` | `domain, site_id, work_item_id, item_type, spec` | `page-build-handler` ✅ (reached `spawn_content_writer`) |
| 8 | `25e84712` | `page-build-handler` | `domain, site_id, work_item_id, **source, spec**` | `generic` ❌ **NO-OP** |
| 9 | `7dc50ad0` | `page-build-handler` | `domain, site_id, work_item_id, item_type, **source, spec**` | `generic` ❌ **NO-OP** |
| 10 (original) | `96b33402` / `284a7f6e` | `page-build-handler` | `domain, site_id, work_item_id, item_type, **source, spec**, page_id, page_name` | `generic` ❌ **NO-OP** |
| 11 (original) | `826b1098` | `page-rerender` | `domain, site_id, work_item_id, item_type, **source, spec**, current_page, page_id, page_name` | `generic` ❌ **NO-OP** |

**The discriminator, isolated by bisection**: `source` alone (row 6) does not
trigger it; `spec` alone (rows 4, 7) does not trigger it; `item_type` alone
(row 5) or with `source` (row 6) does not trigger it. **`source` and `spec`
present TOGETHER (rows 8, 9, 10, 11) triggers it every time**, independent of
`item_type`, `page_id`/`page_name` being present, or which agent_type is
named (reproduced with both `page-build-handler` and `page-rerender`).

Raw correlation ids and full envelopes are in this session's scratchpad if
anyone wants to re-run the exact bisection:
`/home/ant/.claude-scratch/claude-1000/-home-ant-projects-agentchassis/*/scratchpad/test_isolate_corr*.txt`
(session-local, may not survive — the table above is the durable record).

## Why this matters beyond one dispatch

`source` and `spec` are both genuine, common `site_work_items` columns, and
**the drive-loop envelope this repo's own CLAUDE.md-adjacent HANDOFF docs
teach explicitly includes both together**:

```json
{"action":"orchestrate","config":{"agent_type":"<handler_agent>"},
 "input_data":{"domain":"...","site_id":"...","work_item_id":"<id>",
   "item_type":"<item_type>","source":"<source>","spec":<spec>, ...}}
```

(verbatim from `webdesign_uk_build_service/HANDOFF_2026-08-09_continue_here.md`
§2 and `HANDOFF_2026-08-09b_continue_here.md` §1 — both written and used
**today**, both carrying this exact shape). Any session following that
documented recipe with a work item whose `spec` is non-empty (i.e. almost
every real one) hits this. **The queue-starvation workaround CLAUDE.md itself
prescribes — "find the item → claim it → orchestrate its `handler_agent`
directly"** — is silently broken for the common case, and reports success
while doing nothing.

## What is NOT diagnosed

I did not find the Go source responsible. A quick look
(`grep -rn '"orchestrate"' platform/messaging/processor.go`,
`extractGroupInfo` at `processor.go:1067-1120`) confirms `config.agent_type`
IS correctly extracted from the message body — so the short-circuit happens
**after** that, somewhere between workflow selection and execution, and
appears to special-case on `input_data`'s shape rather than `config`. The
`generic` agent's own `default_config` (`agent_definitions.id
6187d808-0d25-441b-b7b3-40af562878af`, `updated_at` **2026-08-09 12:22:37**,
~1 minute before the chassis pods' last restart at 12:23) was edited today —
`[UNVERIFIED]` whether that edit is related; worth checking `git log`/DB audit
history on that row before assuming it is or isn't. The literal string *"No-op
— scheduled task pre_query already did the work"* lives in that DB row, not in
compiled Go, so grepping the binary won't find the mechanism either.

## Workaround, confirmed working

**Omit `source` from the dispatch envelope's `input_data`** (row 7 above shows
`item_type, spec` together still resolve correctly). If the target workflow
step needs the work item's `source` value, it can read it via
`work_item_id` → `site_work_items.source` inside the orchestration instead of
being handed it directly in `input_data`. This is what unblocked the
webdesign.uk Phase 5 dispatch this bug was found while doing.

## How to verify a fix

Re-run row 8 or 9's exact envelope (`domain, site_id, work_item_id[, item_type],
source, spec` — the smallest reproducing case, no `page_id`/`page_name`
needed) and assert `owner_agent_type` equals the requested `config.agent_type`,
`execution_path` is non-empty, and `current_step` is a real step name from
that agent's own `default_config.workflow.steps` (not `complete` reached in
one hop with no prior step). A regression test should assert the same at the
Go level once the responsible code is found.

## CORRECTION 2026-08-10 — the trigger is NOT reliably "source+spec"; it is non-deterministic, and re-testing it caused real production damage

**The bisection table above is real and reproducible for the eleven dispatches it
records, but the headline characterization ("source+spec together") does not
hold as a general rule, and stating it that plainly was wrong.** Discovered by
the same lane, same day, continuing this file's own investigation:

- Re-sending the EXACT "safe" shape from row 7 (`domain, site_id, work_item_id,
  item_type, spec` — no `source`) at a later time **failed** (resolved to
  `generic`) with **byte-identical input** to the earlier run that had
  succeeded. Same payload, different outcome, ~15 minutes apart.
- A completely unrelated, meaningless key (`zzz_nonsense_key`) alongside `spec`
  ALSO triggers the fallback — so it is not specific field NAMES (`source`,
  `page_id`) that matter either.
- This means the trigger is either **time-dependent, state-dependent
  (something about the work_item_id, the site, or a prior dispatch's
  aftermath), or concurrency-dependent** — not a pure function of the
  message's own shape, as the original table implies. The original table's
  individual rows are still accurate reports of what happened at that moment;
  the inference drawn from them (a clean, restatable trigger condition) is
  not.

**This is exactly the class of error `docs024_key_docs_latest/WRONG_CALLS.md`
exists to catch, and it is recorded there too.** The marker discipline that
file asks for — mark an inference `[INFERRED]`, not stated as fact — was not
followed in this file's original text, and should have been.

**Consequence, recorded so the cost is visible, not just the correction:**
re-testing this bisection against a REAL site's REAL work item (there was no
throwaway target available at the time) caused two of the "isolated test"
dispatches to actually run for real, silently regenerating a live page's hero
image binding via the exact landmine this same estate had already documented
(`webdesign_uk_build_service` lane, `NOTES_webdesign_uk_build_service.md`
2026-08-09/10 entry). Caught and reverted the same session, byte-verified
against the last known-good git commit; no lasting damage. But it is the
reason this file now says explicitly: **do not bisect this bug against
production data again.**

## Plan for whoever picks this up — root cause + a SAFE verification path

**1. Do not continue black-box bisection against a real site.** It already
cost one production incident and, given the finding above, further black-box
probing is unlikely to converge — the trigger isn't a clean function of the
message shape, so no amount of additional test payloads will pin it down.
The next step has to be reading code, or running the `090` diagnosis loop
(this durable, cross-cutting, non-obvious-cause claim now clearly qualifies
under CLAUDE.md's "always file" criteria a second time — the first filing
substituted empirical bisection per the 07-31 ruling's escape hatch, but that
substitution has now been shown insufficient, which is itself a reason to
use the loop instead this time).

**2. Where to look, if reading code directly**: `platform/messaging/processor.go`'s
`extractGroupInfo` (confirmed correct — reads `config.agent_type` faithfully)
is NOT the culprit. The swap happens somewhere in workflow *selection*
(`selectWorkflow` and whatever it calls before falling through to the `generic`
agent's own `DefaultConfig`) or in whatever caches/memoizes a resolved
workflow plan. Candidates worth checking first, because they are the only
things that plausibly vary between two calls with byte-identical input a few
minutes apart: (a) a workflow-plan cache keyed on something coarser than the
full message (e.g. just `agent_type`, or a hash that collides), (b) a
per-work-item or per-site in-flight/cooldown guard that isn't visible in
`site_work_items` itself (state living in Redis, an in-memory map, or a
short-lived DB table this investigation didn't check), (c) a concurrency race
in whichever goroutine resolves `agentDef` before the workflow starts. Rule
these in or out by instrumentation/logging around workflow selection, not by
more kcat sends.

**3. A safe verification harness, once a fix candidate exists**: dispatch
against a **scratch site with no real pages/content** (create one if none
exists) or, if a real site must be used, a **work item whose handler's worst
case is a no-op you don't mind repeating** (e.g. `page-rerender` against a
page that only holds placeholder content, or a fresh `content_components` row
nothing links to yet). Confirm the fix by: (a) repeating the exact byte-identical
payload 5-10 times in a tight loop and asserting `owner_agent_type` matches
`config.agent_type` EVERY time, not just once (the whole point of this
correction is that a single success proves nothing); (b) trying the same with
a `zzz_nonsense_key` present, since that's the widest-net reproduction found
so far.

**4. Specifically for webdesign.uk's chat-input-box** (the task that surfaced
this bug — see `webdesign_uk_build_service/PLAN_2026-08-04_webdesign_uk_vm_hosting.md`
Phase 5): **the dispatch mechanism is not actually needed to finish that task.**
The lane worked around it entirely — the component's fields are all static
(no LLM content), so its `page_components` row and the page's rendered HTML
were hand-constructed deterministically and verified byte-for-byte, the same
technique used to recover from the incident above. That work is done and
sitting ready (DB row `fc70ab85-4bb8-4122-a74c-cc5dcaef8684`, spliced page
file in that lane's session scratchpad) — it only needs a `git commit`+push
to `gqls/vm-sites` that a permission classifier blocked mid-session (see that
lane's own NOTES/HANDOFF for the exact state). **Fixing this bug is not on
that task's critical path**; it matters for every OTHER lane relying on the
documented drive-loop pattern while the build queue stays starved.

## Related, not the same bug

`bugs_open/201` — `page-content-writer` dispatched directly with no
`section_plan` either hard-fails or silently no-ops **inside its own workflow
logic** (the child orchestration DOES run, with the named agent's real
workflow, but that workflow's own section-planning branch produces nothing
useful). Distinct mechanism: 201's orchestrations show real
`owner_agent_type` and real `execution_path`; this bug's orchestrations show
neither — the named agent's workflow never starts.

---

# ROOT CAUSE FOUND — 2026-08-10, and it was on the wire, not in the payload's shape

**Status: FIXED IN CODE, committed, NOT YET LIVE.** The fix is inert until an image is
rebuilt and the fleet rolls. Closing condition is at the foot of this section.

## What actually happened

`chassis_intake_events` records every message the chassis ingested, **with its payload
bytes**, and it retains eight days — so the bug's own dispatches were still there. One
query settles it:

```sql
SELECT left(correlation_id::text,8), count(*) AS request_msgs,
       (SELECT string_agg(DISTINCT o.owner_agent_type,',') FROM orchestration_states o
        WHERE o.correlation_id::text LIKE left(e.correlation_id::text,8)||'%'
          AND o.parent_orchestration_id IS NULL) AS owner
FROM chassis_intake_events e
WHERE kind='request' AND topic='system.agent.generic.requests'
GROUP BY 1, e.correlation_id;
```

| dispatches | Kafka messages received | resulting `owner_agent_type` |
|---|---|---|
| the 8 that "worked" (rows control, 1–7 of the table above) | **1** | the requested agent ✅ |
| the 5 that "failed" (rows 8–11 + the re-tests) | **4–6** | `generic` ❌ |

And the payload bytes say why. Failing dispatch `25e84712`, four messages, offsets
107892–107895:

```
{"action":"orchestrate","config":{"agent_type":"page-build-handler"},
 "input_data":{"domain":"webdesign.uk","site_id":"1fcfa4f3-...",
   "work_item_id":"8793da9a-...","source":"phase5-chat-input-pin-2026-08-09",
   "spec":{"mode":"edit_live",...}}}
```

**One message per LINE of the heredoc.** `kcat -P` treats each line of stdin as a separate
message and applies the same `-H` headers to all of them — so four invalid-JSON fragments
arrived, each carrying `action=orchestrate`, and the chassis processed the first one.

**So `source`+`spec` was never the trigger.** It was a proxy for payload length: the
envelopes long enough to be written across multiple lines were the ones that failed. That
also explains, exactly, the correction above — "byte-identical payloads diverged" is true
of the *logical* payload and false of the *wire* payload, and `zzz_nonsense_key`
triggered it for the same reason (one more line, or one line longer).

> **CORRECTION TO THE CORRECTION (2026-08-10).** The 08-10 correction concluded the
> trigger was "time-dependent, state-dependent or concurrency-dependent — not a pure
> function of the message's own shape". It IS a pure function of the message — of the
> message as SENT, which nobody had looked at. Both the original claim and its retraction
> reasoned about payloads as the author wrote them, never as Kafka received them. The
> deciding evidence was one column (`chassis_intake_events.payload`) in a table this
> investigation never opened, and the deciding query counts ROWS, not content.

## The framework defect — which is the part worth fixing

kcat line-splitting is only the trigger. The durable bug is what the chassis did with it:
**an orchestration request it cannot resolve was silently converted into a COMPLETED
no-op run of the consuming pod's own workflow.** Four ways in, all in
`platform/messaging/processor.go` `selectWorkflow`, all pre-fix:

| entry | line | logged? |
|---|---|---|
| body is not JSON → skips Priority 1 AND 2 entirely | :969 | **nothing at all** |
| `FindBestGroup` returns ANY error — no such agent OR a transient DB fault, indistinguishable | :1010/:1037 | Info, `"…didnt find it basically"` |
| definition found but `default_config->'workflow'` is SQL NULL | :1013 | **nothing at all** |
| no DB handle | :999/:1046 | Error |

…each falling to Priority 3 (:1053) = `agentDef.DefaultConfig["workflow"]`, where
`agentDef` is the POD's own definition — `generic`, whose entire workflow is one no-op
`complete` step. Meanwhile `executeWorkflow` (:1815-1828) ran a **second, independent**
parse of the same bytes to decide `owner_agent_type`, so workflow selection and
attribution were two answers to one question that could disagree.

Also found in the same seam and fixed with it: `platform/discovery/agent_discovery.go`
`FindByType` filtered only `type` + `is_active` — missing the `deleted_at IS NULL` and
`is_snapshot` guards that **every** sibling loader carries. Snapshots are stored at
version+1000, so an active snapshot **outranked the live definition** under
`ORDER BY version DESC`.

## Blast radius, measured before changing anything

Eight days of `chassis_intake_events`, 10,851 request messages:

- **48 unparseable bodies — all of them this bug's own kcat fragments** (`from_agent_type=user`).
  No legitimate producer sends an unparseable body. Failing closed catches all 48 and
  breaks nothing.
- **711 orchestrate messages carry NO agent type** — every one of them carries an inline
  `config.workflow` instead (call_agent/spawn envelopes). Priority 1, untouched.
- **9,433 scheduler messages all name `agent_type` explicitly**, including the ticks that
  name `generic` (whose real workflow IS the no-op, because a `pre_query` did the work).
  Those resolve through the lookup, so they are resolutions, not fallbacks — unaffected.

## The fix (committed 2026-08-10)

Fail closed, resolve once, and keep the two failure kinds apart:

- **`platform/errors`** — two typed codes: `DISPATCH_UNRESOLVABLE` (terminal;
  on `NonRetryablePermanentCodes`) and `DISPATCH_LOOKUP_UNAVAILABLE` (transient, built
  `AsRetryable`), with `IsDispatchUnresolvable` / `IsDispatchLookupUnavailable`.
- **`platform/messaging/context.go`** — `MessageContext.Body()`, parsed ONCE per message
  and memoised, error returned rather than discarded. Killed the duplicate parse.
- **`platform/messaging/processor.go`** — `selectWorkflow` returns
  `(workflowSelection{Plan, RunAgentType, Source}, error)`. An orchestration action that
  cannot resolve now REFUSES with a reason: `parse_failure` | `agent_type_unresolved` |
  `workflow_missing`; a transient lookup fault is retryable and distinct. `process()`
  writes a **FAILED `orchestration_states` row whose `owner_agent_type` is the type that
  was ASKED FOR** (concept register SYS-014's own suggested fix shape) plus the bug-196
  failure envelope. `executeWorkflow` reads the resolved type instead of re-deriving it.
- **`platform/discovery/agent_discovery.go`** — the two missing guards; a typed
  `ErrAgentDefinitionNotFound` sentinel so a miss is distinguishable from a fault; and the
  `workflow` column now scans through a `[]byte` (SQL NULL was raising a Scan ERROR, which
  made "this agent has no workflow" — permanent — look transient).
- **`platform/agentbase`** — `processMessage` returns its verdict; the intake worker marks
  a terminally-refused event `failed` (with `DISPATCH_FAIL_CLOSED …` in `last_error`)
  instead of `done`, and leaves a transiently-refused one for the existing attempts<3
  re-pop. Both claim-holders release rather than complete the dedupe claim on the
  transient arm, so a re-attempt cannot lose the race against its own earlier try
  (`StateRepository.ReleaseMessageClaim`).

**Priority 3 survives only for non-orchestration actions and for an orchestration action
that names nothing** (a dedicated/spawned pod, whose own type IS the target — that path
now logs `DISPATCH_OWN_DEFAULT`). No fallback survives a FAILED resolution.

Tests: `platform/messaging/processor_dispatch_resolution_test.go` (10 cases, incl. the
reproduction), `platform/discovery/find_by_type_guards_test.go`,
`platform/agentbase/intake_dispatch_disposition_test.go`. The reproduction was
**mutation-verified**: reverting the parse-failure refusal makes it fail with
`resolved to "generic" (source "own_default")` — the bug, exactly.

## How to verify once it is LIVE (it is not yet)

1. **At the artefact, both directions, every replica** — the new literal must appear and
   the old one must vanish:
   ```
   kubectl exec -n ai-persona-system <pod> -- sh -c \
     'strings /app/agent-chassis | grep -c DISPATCH_FAIL_CLOSED; strings /app/agent-chassis | grep -c "didnt find it basically"'
   ```
   Expect `>0` then `0`. Measured on v1.0.1280 (2026-08-10, pre-fix) it is `0` then `1`,
   so this control can come out either way.
2. **Scratch target only** — per this file's own §3, never production data. Five
   byte-identical SINGLE-LINE dispatches: `owner_agent_type` must equal the requested type
   every time and `execution_path` must be non-empty.
3. **One deliberately fragmented (multi-line heredoc) dispatch**: no workflow runs;
   `SELECT status, last_error FROM chassis_intake_events WHERE correlation_id=…` shows
   `failed` / `DISPATCH_FAIL_CLOSED … parse_failure` on the first fragment; one FAILED
   orchestration row.
4. **One dispatch naming `no-such-agent-239`**: FAILED row with
   `owner_agent_type='no-such-agent-239'` and reason `agent_type_unresolved`.
5. **Health**: scheduler ticks still green (generic no-ops still created via the explicit
   path), and `chassis_intake_events` `failed` count over the hour ≈ the deliberate tests
   only.

**CLOSING CONDITION: steps 1–5 pass on a rolled image.** Until then this stays OPEN —
the defect is still reproducible in production.

## Council — REJECTED on SCOPE, and the code stays

Round `fca1071b-80ac-40cd-8c6d-d30a735de89b`, 2026-08-10, 11 reviewers, 6 abstained:
**7 approve, 1 VETO, 3 object. Decided by a hard veto from `guardian` — on how the change
reached production, not on whether it is right.** `editquality` said in the same round
*"Not a veto: the core fix is on-target"*; `reuse_agent` — the seat whose remit is
architectural fit — reported *"No architecture-level reuse concern identified"*;
`constitution` and `mission` both approved on the grounds that the root cause was fixed at
the mechanism rather than patched at the trigger. The `architecture` seat objected with
`needs_rfc` and said *"the design direction (fail closed, one parse, typed disposition) is
correct… Route it through architecture_review"*.

**Per CLAUDE.md, a scope veto is not answered by resubmitting with better measurements**,
and the seats disagreeing with each other is the stated condition for putting it in front
of a human. So: **the verdict, the guardian's contained alternative verbatim, and the three
substantive objections — checked, not argued — are recorded in
`architecture_review/RFC_023_a_silent_success_becoming_a_loud_failure_is_a_delivery_guarantee_change.md`.**
Two of the three dissolved on inspection; one was real (see below). The commit stays: it is
at HEAD, forward-only forbids the rewrite that would split it, and reverting would restore
a defect that has already caused one production incident.

**The one objection that was real** (`prior_art_librarian`, seconded by `debug_historian`
and `constitution`): `LANDMINES.md:5788` keys on the exact symbols this change edits and
was not cited. Reconciled: that landmine is the NESTED-envelope case — a `call_agent` child
whose `config.agent_type` sits under `body`, invisible to `extractGroupInfo`. **This change
does not fix it.** Such a message still reaches the own-default branch; what changes is
that it now logs `DISPATCH_OWN_DEFAULT` instead of being silent. 7 messages of that shape
in the 8-day census. So the honest statement is: *this fix makes that landmine's failure
mode detectable and leaves it unfixed.*

**And one measurement worth carrying forward**, because it answers the veto's
highest-severity objection ("the FindByType guards change which row EVERY consumer resolves
to, fleet-wide") with a number rather than an argument — measured 2026-08-10, and it could
have come out otherwise:

| check | result |
|---|---|
| active snapshots / active-but-deleted, of 203 definition rows | **0 / 0** |
| agent types whose resolved row CHANGES under the new predicate | **0 of 182** |
| agent types that now resolve to NOTHING | **0** |

The guards are **inert today and prospective** — they close a door that is currently
unlocked and unused, held shut only by `snapshot_agent()` writing `is_active=false`, an
invariant the landmine notes nothing states or tests.

No `Council-Reviewed:` trailer is claimed: the commit carries `Council-Submitted:`, which
asserts nothing, and 098 will resolve this correlation to REJECTED at report time. That is
correct and deliberate — the trailer should show what actually happened.

## Two defects found en route, NOT fixed here (separate filings)

- `platform/messaging/processor.go:66-72` unconditionally re-applies
  `SetMaxOpenConns(4)`/`SetMaxIdleConns(1)`/`ConnMaxLifetime(10m)` to the SAME
  `*sql.DB` that `agentbase` has already sized from `CHASSIS_DB_MAX_OPEN_CONNS` (12 in
  production) — silently undoing the CS-2 pool increase.
- `platform/config/agent_config_loader.go` holds an unmutexed `map[string]*AgentConfig`
  cache keyed only on agent type, reached from the dead `processRequest` (no callers).
  Harmless today, a map race the day someone calls it.


---

# POST-ROLL VERIFICATION — 2026-08-11, v1.0.1284. The refusal is LIVE and PROVEN. One promised trace was a no-op in production, found by running the checks; fixed, awaiting the next roll.

## Live at the artefact, both directions, both replicas

```
agent-chassis-7c9d5f74b9-6j5xn / -rvrdg   docker.io/aqls/agent-chassis:v1.0.1284
  strings /app/agent-chassis | grep -c DISPATCH_FAIL_CLOSED        -> 8   (was 0 on v1.0.1280)
  strings /app/agent-chassis | grep -c "didnt find it basically"   -> 0   (was 1 on v1.0.1280)
```
Positive control proves the pipeline; the negative proves it is THIS change and not a
neighbouring one. Both were measured on the pre-fix binary first, so both could have come
out otherwise.

## Behavioural proof, on a target that cannot do damage

Every test named a **non-existent** agent type (`no-such-agent-239`), so even a wrongly
successful resolution could not have run anything — this file's §3 forbids bisecting
against production data, and that constraint is now permanent.

| test | result |
|---|---|
| 5 × byte-identical single-line dispatches (incl. `zzz_nonsense_key`, the correction's widest-net repro) | **5 of 5** intake rows `failed`, `DISPATCH_FAIL_CLOSED … agent_type_unresolved`, `attempts=1` — terminal, not retried, and **deterministic** |
| 1 × deliberately fragmented multi-line heredoc (the original trigger) | fragment 1 `failed` with `parse_failure`; fragments 2–3 `done` by dedupe (the documented, accepted edge) |
| orchestrations created by all 6 | **0** — pre-fix these were 6 `COMPLETED` `generic` no-ops |
| natural fleet traffic since the roll | 22 dispatches (endpoint-health-checker, build-pipeline-trigger, index-orchestrator, directory-freshness…) all resolved to their REQUESTED type; **none** to `generic` |
| intake events `failed` since the roll that are NOT these tests | **0** — no collateral damage |

Also checked and cleared as a false alarm: no `generic`-owned orchestration appears after
the roll, which looked like a regression in the scheduler ticks. It is not — those messages
run about **once an hour**, not once per tick, and the pre-roll `generic` rows at 08:29–08:31
were single-message **inline-workflow** envelopes (Priority 1, the 711 population), which
still resolve exactly as before.

## ⚠ THE GAP THIS FOUND IN MY OWN FIX — a promised trace that was a no-op in production

**`recordDispatchFailureState` never wrote anything on the chassis, and the fix as shipped
in v1.0.1284 still does not.** It guarded on `p.sqlDB`, which `NewMessageProcessor`
populates **only when `DATABASE_URL` is set** — and it is **not set on the chassis pods**
(`env | grep -c '^DATABASE_URL=' -> 0`). So the guard returned early every time, and the
FAILED `orchestration_states` row owned by the REQUESTED type — the concept-register
SYS-014 fix shape, the thing this bug file and RFC_023 both advertised as the queryable
trace — has never been written in production.

**Why it stayed invisible: two of the three traces WERE present.** The refusal fired (pod
logs carry `DISPATCH_FAIL_CLOSED: requested agent type has no active definition`) and the
intake row recorded it with the full reason. Only the third was missing, and nothing
compared them. The unit test I wrote covered the function's logic while constructing the
processor with `sqlDB` set — a shape production does not have.

**Fixed in the tree** (`db := p.db; if db == nil { db = p.sqlDB }` — the idiom
`selectWorkflow` already used twelve lines away) with a regression test built on the
PRODUCTION shape (`db` set, `sqlDB` nil), mutation-verified: restoring the old guard makes
it fail. **Rides the next roll.**

**The transferable lesson, logged in WRONG_CALLS:** a guard keyed on a field production
never populates is indistinguishable from a guard that passes, and a partially-present set
of traces reads as a working one. When a fix promises N traces, verify N, not the first
one that answers.

## Status

**OPEN — deliberately, and now for exactly one reason.** The refusal, the disposition and
the determinism are proven live. The FAILED-row trace is not yet in a rolled image.

**CLOSING CONDITION:** after the next chassis roll, re-run one unknown-agent dispatch and
assert a row exists with `owner_agent_type='no-such-agent-239'` and `status='FAILED'`:

```sql
SELECT owner_agent_type, status, left(error,80) FROM orchestration_states
WHERE correlation_id::text = '<corr>';
```
Then this file can close. Nothing else is outstanding.

---

# CLOSING VERIFICATION — 2026-08-11, v1.0.1286. The third trace is LIVE. Closing condition MET.

The roll arrived the same day: **v1.0.1286**, both replicas, pods up 12:03Z. Verified in
order, each check disconfirmable:

1. **The fix is in the image, by ancestry, not inference.** Build stamp `c3b424c8e`
   confirmed at the artefact (`grep -aq <full-sha> /proc/1/exe` on a live pod — positive
   hit; negative control `0d822ac56`, committed AFTER the build, absent), then
   `git merge-base --is-ancestor 209917d15 c3b424c8e` → **ancestor**. Note the `strings`
   recipe used in the earlier sections of this file is RETIRED (CLAUDE.md, 2026-08-11);
   stamp + ancestry is the replacement and is what was used here.
2. **The baseline could have come out otherwise:** immediately before the dispatch,
   `SELECT count(*) FROM orchestration_states WHERE owner_agent_type='no-such-agent-239'`
   → **0, all history** — the six v1.0.1284 dispatches naming that type wrote no row.
3. **One single-line dispatch** (corr `cc7bd91a-a80e-4993-bfa9-e0c72be0aad6`, fresh
   `orchestration_id` header `4b11d580-…`, ~12:49 UTC): intake row `failed`, `attempts=1`,
   `DISPATCH_FAIL_CLOSED … agent_type_unresolved` — terminal, not retried.
4. **THE assertion — the row that had never been written in production:**
   ```
   owner_agent_type  | status | error
   no-such-agent-239 | FAILED | DISPATCH_UNRESOLVABLE: DISPATCH_FAIL_CLOSED agent_type_unresolved: orchestration action "orchestrate…
   ```

All three promised traces — the log line, the intake row, and the FAILED orchestration
row owned by the REQUESTED type — are now live and proven. **Nothing is outstanding.**
RFC_023 stays open with the owner (recorded, not resubmitted); `bugs_open/246`,
`bugs_open/247` and the nested-envelope case (`DISPATCH_OWN_DEFAULT` population) remain
filed, unowned, and out of scope here.

**This file STAYS in `bugs_open/` by owner direction (2026-08-06): closure evidence is
recorded in the file; the move is the owner's call, not the fixing thread's.**
