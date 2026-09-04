# HANDOFF — experience_loop, continue here (2026-09-03b)

**Full path of this file:**
`docs/agent_docs/docs024_key_docs_latest/experience_loop/HANDOFF_2026-09-03b_rule_d_shipped_and_the_routing_premise_is_refuted.md`

**SUPERSEDES `HANDOFF_2026-09-03_experience_loop_continue_here.md`.** Read this alone. The
older file stays for the trail; **its §5c contains an instruction that the code refutes — see
§4 here before you act on any of it.**

> **UPDATED 2026-09-04 15:5xZ.** The diagnosis this file was waiting on has landed and it is
> **UNVERIFIABLE, not CONFIRMED** — §4 now says so plainly and says what that does and does not
> license. The loop also named a symbol I had never opened, **`findOrSpawnAgent`**, which would
> have refuted the whole section had it been reachable; it is **dead code** and that is now the
> sharpest thing in §4. Rule D has since **run unattended on schedule** (§2). Everything else
> below stood up to a day and at least two chassis rolls.

---

## 1. Where this stands in one paragraph

The owner asked for `content-quality-auditor` to be put into the new build path. That turned
out to be three jobs. **(1) Give it sight — DONE 09-02 (migration 694, council-approved).
(2) Teach it the promise-keeping questions — DONE, and they fire.** Both verified again today
against chassis `v1.0.1358`. **(3) Route it into the build path — STILL OWED, and its stated
method does not work; §4 has the evidence and the corrected options.** Two side-quests that
became work of their own are now also DONE: **rule D** (the empty-index gap a peer lane found)
is built, ground-truthed and live, and both detectors' scoped-run control bugs are fixed. Both
outstanding peer promises are **discharged**.

---

## 2. What shipped today

### Rule D — a collection page with nothing in it (SQ-005)

| | |
|---|---|
| commits | `95f891a84` (detector + cronjob twin) · `6ad5e6c41` (register, 2 landmines, notes) · `1d96fd609` (peer reports + a third landmine) |
| live | ConfigMap `experience-promise-check-script-fh4ck725kb`, `cronjob configured`, triggered job **exitCode=0**, receipt in `doc_notes` 15:14:47Z |
| council | **NOT in scope** — `council-scope.sh` read, not assumed: `scripts/pattern-check.py` is the only in-scope script, and this is neither that, platform code, nor a migration |

**Fleet at ship, 2026-09-03:** 126 collection candidates — **71 list something, 28 bare
section-indexes skipped, 18 rule D findings across 9 sites, 8 never built, 1 render-vs-data
divergence.** Demand control 62 structured / 2 runtime / 7 own-links / **0 prose**.

**IT HAS SINCE RUN ITSELF, WHICH IS THE ONLY PROOF THAT COUNTS.** The CronJob fired on schedule
**2026-09-04 07:40:00Z**, succeeded at 07:41:04, and wrote its receipt at 07:41:00 with nobody
watching: **18 rule D of 127 candidates — 72 list something, 28 skipped, 8 never built, 1
divergence.** The fleet grew by one candidate overnight and that page LISTS something (71 → 72),
which is the demand control doing its job rather than a number drifting. **Rule B is now 0**
(was 1): boxingonline's fight-calendar, the page the whole family was built from, has been
repaired by another lane.

**Ground truth, both directions, sampled away from the motivating site:** all **18** findings
confirmed EMPTY at the served body with a per-domain invented-URL control; **11 of 12** cleared
pages confirmed to serve items; the 12th (vonc's runtime-fill archive) checked at its source
instead — `https://vonc.com/data/provocations.json` serves **21 entries**.

**Three false positives the curl caught and the rule read cleanly past** — all three are now
fixtures, two are landmines: the inner join to `content_components` (hid two 20-odd-article
blog listings, because `rebuild_blog_listing` writes `component_id` NULL); matching the whole
URL path (made every page under `/guides/` a guides index); matching the whole title (imported
the site name — "Contact Us — Garden Tools UK").

### The other two

- **SQ-004 + SQ-005 scoped demand controls** now print `n/a` instead of `FAIL` when the
  control case or every clean page is outside `--site`. (SQ-004's half shipped this morning as
  `e535fc4f0`; I wrote the same bug into SQ-005 hours later and fixed it in `95f891a84`.)
- **`render_data_divergence`**, a bucket for the page that is neither empty nor clean.

---

## 3. Peer promises — BOTH DISCHARGED, do not re-promise them

| to | owed | state |
|---|---|---|
| `designblog.co.uk` | build the empty-index rule, re-run over their four pages, report | **DONE** — all four fire, re-confirmed at the served body; reported in their `NOTES_designblog_couk.md`, 2026-09-03 ~16:30Z. Both my caveats to them WITHDRAWN |
| `vetcomparison` | re-run both detectors post-701; fire the widened auditor (record mode) | **DONE** — 701's remainder verified at the ARTEFACT (it is NOT in `schema_migrations`, hand-applied); both detectors clean with non-vacuous controls; auditor RAN, correlation `bd349a6c-c7a6-432b-a2c1-f6f7a6b8d6fa`, COMPLETED in 51s, 8,235 input tokens, **4 findings, all record-mode**. Reported in their `NOTES_vetcomparison.md` |

---

## 4. ⚠ §5c's METHOD IS REFUTED BY THE CODE — read this before routing anything

The superseded handoff says, in bold: *"Use `call_agent`, not `spawn_agent` — the spawn→call
handshake fails ~half the time fleet-wide, and `apply_site_design` in that same workflow is
already a `call_agent` to copy verbatim."*

**A bare `call_agent` naming `content-quality-auditor` will fail on every build.** Read
`platform/orchestration/actions/call_agent.go`:

- `findTargetAgent` (~:225) has exactly three arms. With no `target_role` it tries
  `findAgentByType`, which **scans `CollectedData` for a SPAWN RESULT** of that type; then
  `isStandardAgent`, which (~:446) is the two-element list `{"search", "image"}`; then returns
  `fmt.Errorf("no spawned %s agent found")`.
- `apply_site_design` works **only because `spawn_webdesign_agent` ran earlier in the same
  workflow** and left a `webdesigner` role in collected data. It is not a copyable template
  for an agent that is never spawned.
- The two other candidate actions are not escapes: **`dispatch_agent`**
  (`dispatch_actions.go`) is a *remote-cluster alternative to spawn* whose own usage block
  shows `"next_step": "call_researcher"`, and **`StartOrchestrationAction`**
  (`spawn_actions.go` ~:1685) begins with `findSpawnData(params)`.

**So: every in-workflow path to another agent appears to require a prior `spawn_agent`.** That
is a structural claim, so per the 2026-07-31 owner ruling it was **filed for independent
diagnosis rather than asserted**: `RUN_CORRELATION_ID=ebb08a21-4460-427c-b2ca-dea66ef40c04`
(queue was empty; prior art checked — `bugs_open/210` is a *different* `call_agent` failure,
an `input_mapping` path miss, not this arm).

### ⚠ THE VERDICT IS **UNVERIFIABLE**, NOT CONFIRMED — read this before quoting me

The run COMPLETED 2026-09-03 15:49:56Z after 5 LLM calls and returned
**`status: UNVERIFIABLE`, `stopped_by: iteration-cap`** — *"Diagnosis NOT confirmed. Best-effort
trail attached for a human; no fix proposed."* **So the claim above rests on my code-read alone.
Do not cite the 090 run as corroboration; it explicitly declines to give any.**

What the loop *reasoned* agrees with me — its last hypothesis says `dispatch_agent` *"IS a spawn
(just remote), not a consumer of an earlier spawn step"* and that `StartOrchestrationAction`
*"explicitly requires prior spawn data … which is the opposite of an escape hatch: it enforces
spawn-before-use"*. But it then **refused to certify its own reasoning**, because its bundle
never contained those two function bodies — only an `ls` for one and a one-line fragment for the
other — so its claims were *"asserted but not quotable from anything in this bundle."* That is
the loop being honest about its evidence, and it is the right call. It is also the same
discipline this file asks of you.

### The loop earned its cost anyway: it named `findOrSpawnAgent`, and that is the real trap

Its symbol scan surfaced **`platform/orchestration/actions/call_agent.go:findOrSpawnAgent`** —
a function I had never opened, sitting in the very file I had just read end to end. **Its body
does exactly what §4 says is impossible:** it checks collected data, and failing that builds a
`spawn_agent` step config and spawns one on demand (`:749-761`).

**It has ZERO callers.** Measured 2026-09-04, repo-wide, with a positive control:

```bash
grep -rn "findOrSpawnAgent" --include=*.go .   # only its own definition + its own log line
grep -rn "findTargetAgent"  --include=*.go .   # CONTROL: returns call_agent.go:39, the live call site
```

`CallAgentAction` calls `findTargetAgent` (`:39`), never `findOrSpawnAgent`. So the claim stands
— **but a future reader who greps "can call_agent spawn on demand?" finds a function whose NAME
answers yes and whose body answers yes, and it is dead.** That is a landmine and is filed as one.

### The durable evidence the loop asked for, and what it does NOT prove

The loop's fair objection was that no runtime row shows the mechanism either way. That is
answerable from `agent_error_log`, which is **durable** — 56,450 rows back to 2026-07-24, unlike
`orchestration_states` (a ~28-hour window; see §6):

| measured 2026-09-04 | count |
|---|---|
| `call_agent` failures logged (**positive control** — this log does capture them; 4th most common failing action) | **6,074** |
| of those, `no spawned … agent found` | **0** |
| of those, `no agent with role …` | **0** |

**Read this twice, because it cuts both ways.** It is consistent with my claim — every live
`call_agent` step is preceded by a spawn, so the target-not-found arms are never reached. It also
means **my predicted failure has no production precedent in six weeks**, so wiring a bare
`call_agent` would be exercising a path this estate has never run. Neither reading licenses
skipping the fix.

### The options, once the verdict lands

1. **`spawn_agent` + `call_agent`**, the `site-review-agent` pattern. Honest cost: the
   handshake's ~50% failure. Mitigation worth designing rather than assuming: the audit is
   advisory and record-mode, so a failed handshake should not fail the build — make the step
   non-fatal. **This is the only option with production precedent** — see the 6,074-row table
   above: every `call_agent` this estate has ever run had a spawn in front of it.
2. **A cold-dispatch action.** What I did by hand proves the *transport* works: one
   `orchestrate` message to `system.agent.generic.requests` naming
   `content-quality-auditor` with `{site_id, domain}` ran the whole audit in 51 seconds. There
   is no workflow action that does this. Adding one is **platform code and architecture-scope**
   (a new shared mechanism on a shared seam — see CLAUDE.md's 2026-07-28 ruling), so it needs
   the concept register in the same commit and a council round. **⚠ `findOrSpawnAgent` is the
   trap on this route:** it looks like option 2 already built. Reviving a dead function is a
   behaviour change to a shared action with no callers to regress against — treat it as new
   code, not as a revert.
3. **Leave it out of the workflow** and let the sweep cover it, which it already does — 13
   distinct sites audited today, roughly one every 15–20 minutes, all COMPLETED. This is the
   option to argue against explicitly rather than by omission: the owner asked for the build
   path specifically, and the sweep does not gate delivery.

**Unchanged and still correct from the old §5c:** insert between `apply_site_design` and
`update_site_status` (verified live today — the tail is still
`fix_items_loop → apply_site_design → update_site_status → complete`); the auditor needs **no
storage client**; it is **not redundant with `content-reviewer`** (that is per-PAGE inside
`build_items_loop`, this is a GROUP/site auditor, and boxingonline is the argument — every page
passed individually and the SITE was wrong); a migration under `sql_for_agents/` **is** council
scope.

---

## 5. ⚠ NEW LANDMINE — do NOT run `001_audit_agent_trigger.sh`

`scripts/initial_messages/010_audit_triggers/001_audit_agent_trigger.sh` has a clean usage
header, worked examples, and a `SAFETY:` paragraph promising *"Each agent operates on ONE site
only (the site_id you pass). It will NOT affect other sites."* **All of that is false.** The
file is **895 lines, mode 755, with ELEVEN `kcat -P` publish blocks and no `exit` between
them.** Lines 36–38 require the three arguments; every later block reassigns
`AGENT_TYPE`/`SITE_ID`/`DOMAIN` to hardcoded literals before publishing. Running it as
documented dispatches **11 runs at six OTHER lanes' sites and none at yours** — including
`blog-content-planner` and `image-build-handler`, which write rather than audit.

**Use this instead** (it worked today, correlation `bd349a6c`):
```bash
. scripts/kafka-publish-lib.sh
CORR=$(cat /proc/sys/kernel/random/uuid)
kafka_publish_checked --topic system.agent.generic.requests \
  --payload '{"action":"orchestrate","config":{"agent_type":"content-quality-auditor"},"input_data":{"site_id":"<uuid>","domain":"<domain>"}}' \
  --correlation "$CORR" --header correlation_id=$CORR --header orchestration_id=$(cat /proc/sys/kernel/random/uuid) \
  --header request_id=$(cat /proc/sys/kernel/random/uuid) --header message_id=$(cat /proc/sys/kernel/random/uuid) \
  --header message_type=request --header client_id=demo_client --header action=orchestrate \
  --header sender_agent_type=cli --header sender_agent_id=cli-user \
  --header responses_topic=system.agent.generic.responses --header timestamp=$(date -u +%FT%TZ)
# then VERIFY BY YOUR OWN CORRELATION, never a time window:
#   SELECT count(*) FROM orchestration_states WHERE correlation_id='<CORR>';
#   SELECT count(*) FROM llm_call_log        WHERE correlation_id='<CORR>';
```
Target the **group agent** directly; that is what avoids the spawn→call handshake by hand.

---

## 6. Per-roll verification (unchanged, still cheap, still do it)

**Re-run 2026-09-04 15:5xZ, against `v1.0.1360` (chassis pods up 2026-09-03 22:06/22:07Z, both
stamped `239ab3626`). Every 694 marker intact** — allow-list gone, non-greedy strip,
`ORDER BY pc.position`, all four dimensions, `filing_mode` still `record`. The seat's
`updated_at` moved again (now 2026-09-03 22:05:57), as it does at roughly every roll, with **no
snapshot taken and nothing lost**. That is now **seven rolls survived**.

| measured 2026-09-04, last 24h, from `llm_call_log` | |
|---|---|
| auditor calls | **60** |
| average input tokens | **6,749** (pre-694 average was **1,744**) |
| failures | **0** |
| demand control: fleet calls / distinct agent types | **2,814 / 31** |

**⚠ A ROLL IS IN FLIGHT AS THIS IS WRITTEN.** A peer session reports
`make release redeploy-agents` running for **`v1.0.1361`, cut at `06c0b18f2`**, all 14 images
built 15:29–15:36Z and pushing. Nothing in this lane is Go, so **nothing here ships or breaks
with it** — rule D is a CronJob mounting a ConfigMap and is untouched by an agent-pod roll. Two
consequences that ARE yours: **re-run the marker check above once the new pods are up**, and
**do not dispatch for ~300s after they restart** (the spawn is silently dropped).

**Read the running commit from the table, not from a log line** (CLAUDE.md was reordered
2026-09-04 to say so):
```sql
SELECT pod_name, git_commit, started_at FROM service_binary_capabilities
 WHERE kind='build' AND pod_name LIKE 'agent-chassis-%' ORDER BY started_at DESC;
```
⚠ It is a **two-hour window**, not a history — it answers *what is running now*.

**Use `llm_call_log`, never `orchestration_states`,** for anything durable. That table is a
rolling window: `[MEASURED 2026-09-04]` oldest row **2026-09-03 11:47**, newest now, **6,096
rows** — about **28 hours**. A zero there means "reaped", not "never ran". For errors the
durable table is **`agent_error_log`** (56,450 rows back to 2026-07-24), which is what §4's
6,074-row control is built on.

---

## 7. What is left

1. **The routing (§4)** — the diagnosis came back **UNVERIFIABLE**, so there is no verdict to
   lean on: the claim is mine, from the code, and §4 now carries the six-week durable evidence
   and the `findOrSpawnAgent` trap. Pick an option and argue it. If you want independent
   confirmation, a **re-filed 090 naming the two function bodies the last run could not get**
   (`DispatchAgentAction`, `StartOrchestrationAction`) is the cheap next move — the loop said
   exactly what it was missing.
2. **CONTRIB ask 3, planner half** — refusing to SELECT a tool at planning time when we hold
   no data to put in it. The mechanical half shipped as SQ-005 rule B; this is the
   planner-side door-closer.
3. **Nothing else is owed to anyone.** Both peer promises are closed (§3).

---

## 8. The five standing docs

| doc | path |
|---|---|
| PLAN | `…/experience_loop/PLAN_experience_loop.md` |
| RUNBOOK | `…/experience_loop/RUNBOOK_experience_loop.md` |
| NOTES | `…/experience_loop/RUNNING_NOTES_experience_loop.md` — rule D entry at the foot |
| README (owner's) | `…/experience_loop/README_where_we_are.md` — **append only, never rewrite** |
| SUMMARY | `…/experience_loop/SUMMARY_where_experience_loop_is_now.md` — **deliberately not rewritten again.** The milestone is the routing going live; the five headings would still repeat the last one. Write a NEW dated file when §4 lands |

Register: **SQ-004** `listing-class-promise-check` (`25 7 * * *`) · **SQ-005**
`experience-promise-check` (`40 7 * * *`, now rules A–D).

**Refused deliberately, do not re-attempt as a regex:** *"does this page contain the thing its
title asserts?"* — disproof in SQ-005's docstring. Rule D is decidable because it counts items.
**Also refused:** widening rule D's collection nouns to catch designblog's `/criticism/` — it
would catch homegarden.uk's twelve month pages, which are articles. Stated gap, not an oversight.
