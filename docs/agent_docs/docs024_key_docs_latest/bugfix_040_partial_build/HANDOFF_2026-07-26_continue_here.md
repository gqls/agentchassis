# HANDOFF — bugs_closed/040-partial-build, candidate 2: LIVE but behaviour NOT yet induced

**Written 2026-07-26 21:15 UTC.** Read this first; the case file
`bugs_closed/040_HANDOFF_2026-07-20_failed_page_build_leaves_page_deployed_and_partially_composed.md`
holds the full history, this holds only **what is owed and how to finish it**.

> **CORRECTED 2026-07-26 22:00 UTC — §4's "UNSOLVED PROBLEM" was never a defect, and
> §5 of this same document contains the answer.** The five publishes were not dropped:
> `generic-requests-group` was stalled at a frozen committed offset of **105196**, and
> §5's own backlog dump lists the probe messages queued at **105197, 105202 and 105204**
> — immediately behind the stall. When the lane cleared they were consumed and ran
> normally. **The probe completed at 21:16:09 and 21:16:16 UTC, one minute after this
> handoff was written**, and both assertions passed. None of §4's three hypotheses were
> needed; the agent-definition row was fine all along.
>
> **Owed items 1 and 3 are therefore DISCHARGED** — see the case file's
> "VERIFIED LIVE 2026-07-26 21:16Z" section and `NOTES_040_partial_build.md`. Owed item 2
> needed a new probe (see the correction in §3). The cheap check that would have saved
> five attempts and three hypotheses: **when a dispatch looks eaten, re-read the outcome
> table before theorising about the cause — a stalled queue and a dropped message are
> indistinguishable from the publisher's side.** Logged in `WRONG_CALLS.md`.

---

## 1. Where it stands in one paragraph

Bug **040-partial-build is CLOSED** — the defect (a partial build stamped `deployed`
+ plan-version, invisible to the reconciler for ever) was fixed at every layer and
verified end to end on 2026-07-24. **Candidate 2**, its last unowned residual, is the
*reporting* half: `update_work_item_status` now records the routed step error on the
work item instead of leaving `site_work_items.error` blank. It is **committed
(`43002d3a4`), council-APPROVED (`66d77d4d`, 8 seats, 2 medium objections both
answered), and LIVE in the running binary** — pod-grep confirmed on v1.0.1170 and
again on **v1.0.1171**. What is **not** done is a live behavioural proof: no work
item has been stamped `failed` by this path since the fix rolled, and my induced-fault
harness has failed to dispatch five times for reasons unrelated to the fix.

**Nothing is blocked on this.** It is a verification debt, not a defect.

## 2. What IS verified (with the evidence)

**Deployment, on the running pod — not git, not the tag** (v1.0.1171, pod
`agent-chassis-5b4456686c-s5fkc`):

```
strings /app/agent-chassis | grep -c "no error_message literal"   -> 1   # created by this change
strings /app/agent-chassis | grep -c "build is short of its plan" -> 1   # positive control (040 guard, live since v1.0.1146)
strings /app/agent-chassis | grep -c "candidate two placeholder xyzzy" -> 0   # negative control
```

The first string exists **only** inside the new fallback branch, so it cannot be
satisfied by the old binary — that is the discriminating half. The control proves the
grep itself works, which a bare `grep -c` of your own new string never does.

**Unit behaviour:** `platform/orchestration/actions/update_work_item_status_error_test.go`,
6 cases, all branches, fixtures taken from the real live error shapes. Run against
`git archive HEAD` + the two changed files (the shared tree is dirty with other
sessions' work) — package green, `platform/... internal/... cmd/...` build clean.

**Guardian's containment check (the one that mattered):** every literal-less,
non-`complete` `update_work_item_status` step fleet-wide, resolved across all six edge
kinds (`error_step`/`next_step`/`else_step`/`then_step`/`on_success`/`on_failure`):

```
image-build-handler | mark_work_item_failed | failed | error_step
page-build-handler  | mark_item_failed      | failed | error_step
```

Exactly two, each reachable **only** as an `error_step` target — so `__step_error` is
set by construction wherever the fallback can fire. Query is in the case file.

## 3. What is NOT verified, precisely

**No work item has been stamped `failed` by `mark_item_failed` since the fix went
live.** The blank-error census has not grown — but it has not been exercised either,
so that is *absence of traffic*, not evidence. Do not write "verified live" on the
strength of it. (This is the [[verify-the-failing-branch]] trap: a quiet census plus a
pod-grep proves deployment, never correctness.)

Owed, in order:

1. **Induce one failure through `mark_item_failed`** and assert the item's `error`
   reads `step deploy_page failed: …` (or `step <name> failed: …`) instead of blank.
2. **Negative control:** a `needs_human_review` park must still read its configured
   literal (`page-build-handler no-op: …`), not a routed error.
3. **Negative control 2 (the load-bearing one):** a `complete` stamp reached *after* an
   error route must leave `error` untouched — `__step_error` is never cleared, and the
   fleet census found exactly one literal-less `complete` step (image-build-handler)
   that would otherwise inherit a stale failure.

The **prefix branch** (`step X failed: ` prepended when the message does not already
say so) is exercised only by the awaited-request-timeout shape, which needs a real
`call_agent` timeout. It is unit-tested and is pure string logic; mark it
`[UNVERIFIED LIVE]` rather than claiming it.

> **UPDATED 2026-07-26 22:00 UTC — items 1 and 3 are now DISCHARGED; item 2 is not, and
> history cannot discharge it.**
>
> Items 1 and 3 both passed on the 21:16Z run (case file, "VERIFIED LIVE"). This
> paragraph's `[UNVERIFIED LIVE]` on the prefix branch **stands, and is now positively
> confirmed as the right call**: the run's `__step_error.message` already read
> `step boom failed: …`, so `HasPrefix(errorMessage, "step ")` was true and the prefix
> was correctly skipped. I briefly claimed the opposite from the work item's text before
> checking — **the output of a prefix-if-absent branch is indistinguishable from the
> output of no branch at all**, so read `collected_data->'__step_error'` in
> `orchestration_states`, never the work item, to tell them apart.
>
> **Item 2 needs an induced run, not a census.** The only live rows carrying the literal
> (`page-build-handler no-op: no sections ready to build …`, 4 rows) were all written
> **before** the roll that made the fallback live (v1.0.1170, 18:35Z). They prove a
> literal was written when nothing competed with it — not that a literal still wins now
> that a fallback exists and `__step_error` is set. The harness below has been extended
> with **PROBE B** for exactly this.

## 4. The probe harness — WORKS, and now covers all three owed items

**It was never broken.** See the correction at the top: the five "failed" publishes were
queued behind a stalled consumer and ran fine when it cleared. The harness is good; the
only thing that ever went wrong was reading a stalled queue as a dropped message.

Extended 2026-07-26 22:00Z with **PROBE B**, so one run now tests all three owed items.

| thing | value |
|---|---|
| agent type | `scratch-cand2-probe` (`agent_definitions`, `is_active=true`) |
| PROBE A item (must end `failed` + non-blank error) | `d418240c-f88f-480a-85c8-b328c901b7f5` |
| PROBE B item (must end `needs_human_review` + the **literal**) | `b0b0b0b0-1111-4222-8333-444444444444` |
| PROBE C item (must end `complete` + error still BLANK) | `1b001fec-8e4e-4e4d-b6d3-2eb17d9e4c4c` |
| item_type | `scratch_cand2_probe` — no handler, so the dispatch loop cannot pick them up |
| site | dartsonline.com (`5fe8785b-223d-41a3-88ee-c07187622381`) |

Workflow: `boom` (deliberate error — `update_work_item_status` with a `work_item_id_field`
that resolves to nothing and `skip_if_missing:false`) → `error_step: mark_failed`
(status `failed`, **no** literal) → `mark_park` (status `needs_human_review`, **with** a
literal) → `mark_complete` (status `complete`, **no** literal) → `done`. Every step after
`boom` runs with `__step_error` set, which is the whole point: A must inherit it, B must
ignore it in favour of its literal, C must ignore it because it is a `complete`.

**Reset before each run** — `attempt_count` is not reset by the workflow and `max_attempts`
will eventually refuse:

```sql
UPDATE site_work_items SET status='detected', error='', attempt_count=0, result=NULL, max_attempts=9
 WHERE item_type='scratch_cand2_probe';
```

**Fire it with** (`setup` + `fire` scripts are reproduced in `RUNBOOK_040_cand2_probe.md`):

```bash
CORR=$(uuidgen); ORCH=$(uuidgen)
PAYLOAD='{"action":"orchestrate","config":{"agent_type":"scratch-cand2-probe"},"input_data":{"item_a":"d418240c-f88f-480a-85c8-b328c901b7f5","item_b":"b0b0b0b0-1111-4222-8333-444444444444","item_c":"1b001fec-8e4e-4e4d-b6d3-2eb17d9e4c4c"}}'
kubectl -n kafka run "kcat-cand2-$(date +%s)" --rm --restart=Never \
  --image=edenhill/kcat:1.7.1 --attach --quiet --command -- \
  sh -c "printf '%s' '$PAYLOAD' | kcat -P -c 1 \
    -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
    -t system.agent.generic.requests \
    -H correlation_id=$CORR -H request_id=$(uuidgen) -H message_id=$(uuidgen) \
    -H orchestration_id=$ORCH -H orchestration_name=cand2-probe \
    -H step_name=start -H client_id=demo_client -H message_type=request \
    -H action=orchestrate -H from_agent_type=user -H from_agent_id=cli \
    -H responses_topic=system.agent.generic.responses && echo PUBLISH_OK"
```

> **Two publish traps, both silent, both defeated by the same discipline.** The original
> form above (`printf | kubectl run -i --rm ... -- kcat -P`) exits 0 having published
> nothing. Replacing it with `-- sh -c '…'` fails differently: the image's **entrypoint is
> kcat**, so `sh -c …` arrives as kcat *arguments* and you get the usage text and
> `-b <broker,..> missing`. `--command` is what replaces the entrypoint.
> **Put the payload in the container command and make the container print `PUBLISH_OK`** —
> a publish with no positive confirmation is not evidence of a publish. Then confirm the
> offset it landed at before concluding anything:
> `kcat -C ... -p 0 -o <offset> -c 8 -e -f '%o %T %h\n'`.

Assert:

```sql
SELECT summary, status, COALESCE(NULLIF(error,''),'<<BLANK>>')
FROM site_work_items WHERE item_type='scratch_cand2_probe' ORDER BY summary;
-- PROBE A -> failed             | step boom failed: …        <- the fix firing
-- PROBE B -> needs_human_review | cand2 probe literal: …     <- the literal winning
-- PROBE C -> complete           | <<BLANK>>                  <- the exclusion holding
```

And to tell the prefix branch apart from the plain copy — the work item alone **cannot**:

```sql
SELECT collected_data->'__step_error'->>'failed_step',
       collected_data->'__step_error'->>'message'
FROM orchestration_states WHERE orchestration_id = '<the ORCH you published>';
```

### THE UNSOLVED PROBLEM — read before re-firing

**Five publishes produced ZERO orchestration rows.** The message reaches the topic
(proved by consuming it back), the chassis subscribes to that topic and is actively
processing other traffic, the consumer group reaches lag 0 — and no
`orchestration_states` row for `scratch-cand2-probe` is ever created, with **no error
line in the pod log naming the type**. Ruled out along the way, each by a check:

| ruled out | how |
|---|---|
| message never published | consumed it back off the topic — present, correct JSON |
| the ~300s post-restart drop window | waited it out (310 s) and re-fired; same result |
| `processing_mode: "task"` | changed to `orchestrator` (matching `council-gate`); same result |
| `description` NULL — `loadAgentDefinition` Scans it (`processor.go:354-376`) | set a description; same result. **Still the best remaining hypothesis for attempts 1–3**, since it would fail before any log line naming the type |
| empty `capabilities` array | set `["diagnose"]` (matching `council-gate`); same result |
| the consumer stall of §5 | that cleared; lag 0 and moving during attempts 4–5 |

**Next hypotheses, cheapest first** (do NOT just re-fire — that is what burned five
attempts):

1. Read `platform/messaging/processor.go` `ProcessMessage` around the `orchestrate`
   action and find where `config.agent_type` is actually consumed. The pod logs
   `Agent definition loaded … type=generic` for every message, so the *consumer's own*
   type is loaded first; find what then re-dispatches to the named type, and what it
   requires that `council-gate` has and this row lacks. **Diff the two rows column by
   column** — that is the check I should have run at attempt 2 instead of changing one
   field at a time.
2. Compare against a working publish captured live: consume a real `council-gate`
   message off the topic and diff its **headers** against the ones above (the payload
   shape is already identical — the difference may be in a header, not the body).
3. If it still resists, abandon the scratch harness and induce the fault on a **real**
   handler instead: give a page-build a `needs_page` item for a page whose build will
   error deterministically. Slower and messier, but it exercises the real workflow.

**Cleanup when done** (leave nothing behind):

```sql
DELETE FROM site_work_items WHERE item_type='scratch_cand2_probe';
DELETE FROM agent_definitions WHERE type='scratch-cand2-probe';
```

## 5. Found on the way — the generic request lane stalled for ~25 minutes

While probing, `generic-requests-group` on `system.agent.generic.requests` sat at a
**frozen committed offset (105196) with lag 13 for over 20 minutes**, while the chassis
pod was alive and busily processing responses and the *scheduled* lane (
`generic-requests-group-lane-system-agent-scheduled-requests`) sat at lag 0 throughout.

**It was not only my traffic.** Reading the backlog off the topic:

```
105196 18:30:05 council-gate       105202 18:36:58 scratch-cand2-probe
105197 18:32:40 scratch-cand2-probe 105203 18:38:42 council-gate
105198 18:32:53 page-rerender      105204 18:40:24 scratch-cand2-probe
105199 18:33:32 council-gate       105205 18:40:36 council-gate
105200 18:34:15 council-gate       105206 18:41:04 scratch-069-chromelock
105201 18:35:36 council-gate       105207 18:44:34 council-gate
                                   105208 18:45:11 page-rerender
```

Six other threads' council submissions, two `page-rerender`s, and another thread's
scratch probe — all stuck, none of those sessions able to see why. The stall began
~18:30, i.e. **before** the 18:35:07Z roll to v1.0.1170, and the fresh pod resumed from
the committed offset and consumed nothing further. **It cleared on its own by 21:04**
(offset 105196 → 105247, lag 0).

This is the lane `bugs_closed/030` cleared today (its own measurement at 13:26 showed
lag 0 and ~1 s publish→run), so this is a **recurrence on a just-closed case, with a
different signature**: 030 was *slow but progressing* (one partition, in-order queue,
~18–36 min); this was *stopped* — 13 messages, zero progress, 20+ minutes, while the
same pod processed other topics normally.

**Not filed, deliberately.** 030 was closed hours earlier by another session that is
plainly still active, and `who-owns` discipline says contribute to the owner rather
than open a competing front. **It is owed as a contributed observation on
`bugs_closed/030`** (that file already carries two such sections) — or a reopen, which
is the owning thread's call, not mine.

**The one command** anyone should run before concluding their dispatch was dropped:

```bash
kubectl -n kafka exec personae-kafka-cluster-combined-pool-prod-0 -- bash -c \
 '/opt/kafka/bin/kafka-consumer-groups.sh --bootstrap-server localhost:9092 \
  --describe --group generic-requests-group'
```

A **frozen `CURRENT-OFFSET`** (not merely a large lag) is the signature. Sample it
twice, 30 s apart — one reading cannot tell "stalled" from "busy".

## 6. Standing traps this episode added

- **`2>/dev/null` in a polling loop turns every error into your own default branch.**
  I polled `WHERE id='<uuid>'` — the column is `orchestration_id` — with stderr
  suppressed, and read my own hand-written `<no row yet: queued>` sixty times for a
  council run that had died in **7 seconds**. Run the query once in the foreground
  first; give a watcher a fail-fast case for terminal states, not just the success one.
  (`WRONG_CALLS.md`, 2026-07-25.)
- **Council submission schema is stricter than the 097 header:** `risks` is a **string**,
  not an array; `edits[].operation` is `modify|add|remove|config_change` — **`create` is
  refused**. An invalid run writes **no artifacts**, so a watcher keyed on
  `diagnosis_artifacts` alone waits for ever. Three threads hit this in one morning;
  written up in `RUNBOOK_council_gate.md`.
- **`abstained` is NOT a health signal on the 16-seat gate** — it counts
  relevance-filtered seats, by design. Check **`unreadable: 0`** and
  `reviews-in-body == reviews-carrying-a-verdict`. I read my own unanimous 8-seat
  APPROVED as "every seat abstained". The `abstained: 0` rule belongs to the 4-seat
  experience-loop council and does not transfer.
- **Change one thing per attempt, but diff first.** Five probe attempts each altered a
  single field against a guess. A column-by-column diff of my row against the
  known-working `council-gate` row was available from attempt one and is where the
  next session should start.

## 7. Related

- Case file: `bugs_closed/040_HANDOFF_2026-07-20_failed_page_build_leaves_page_deployed_and_partially_composed.md`
  (§ "Candidate 2 — BUILT 2026-07-25" and § "Council: APPROVED round 1").
- `016b_debugging_guide_8_consolidated.md` §10, row `040` — layer (4).
- `bugs_closed/030` — the lane; §5 above is owed to it.
- Council: `Council-Reviewed: 66d77d4d-0f12-4c16-84c7-e1e3487411aa`.
