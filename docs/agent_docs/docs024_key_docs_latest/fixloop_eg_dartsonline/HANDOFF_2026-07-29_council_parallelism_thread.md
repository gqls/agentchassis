# HANDOFF — council parallelism thread (2026-07-29)

**Cold-start for a fresh chat. This supersedes
`HANDOFF_2026-07-28_council_parallelism_thread.md` as the entry point** — that file
stays as the working record (the wrapper design, the three wrong calls, the
truncation write-up in full). Read this one first; go there for detail.

Everything below is committed. Current chassis: **`v1.0.1197`** (both pods 08:05Z,
identical `imageID sha256:458c9ad0…`).

---

## 1. What this thread was for, and why it is essentially DONE

**"Could/should we run several councils in parallel?"** Yes to both. The answer was
a thin orchestrator (`council-gate-orchestrator`) that spawns the council into its
own pod, releasing the request lane in ~8 s so N councils run through ONE lane.

**Built, applied, and PROVEN end to end** (07-28 15:20Z: wrapper →
`AWAITING_RESPONSES` in 37 s → 16 seats in a dedicated pod → `council_report` →
`COMPLETED`, ~10.75 min, generic-lane LAG 0 throughout).

**The default was deliberately NOT flipped, and that decision stands.** The flip
proposal went to the gate and came back **REVISE** (corr `f5da8f65`), gated by the
guardian, `unreadable 0` — a real verdict, not the harness. The objections were
accepted rather than argued. The urgency is gone anyway: the dedicated lane already
stops councils blocking other work, and `replicas=2` already delivers two
concurrent councils.

**Reaching it:** `TARGET_AGENT_TYPE=council-gate-orchestrator ./097_TRIGGER_council_review_v1.sh sub.json`

## 2. What was done 2026-07-28/29 (this is the live part)

### The council was losing ~a third of its rounds to seat truncation. Fixed.

`review_editquality` was truncating at its 8000-token ceiling on **12 of 48 calls
(25%)** in 24h. Because it is one of the two `ALWAYS_ON` seats
(`099_SYNC_gate_roster.py:48`) it runs every round, so **12 of 37 council rounds in
10 hours decided with at least one seat's opinion partial or lost.**

**Owner chose the one-seat option.** Raised `max_tokens` **8000 → 16000** on
`review_editquality` only, written to **both** `fix-proposer` and `council-gate`.

**VERIFIED IN EFFECT — cite the positive control, not the zero:**

| window | calls | truncated | **max output** |
|---|---|---|---|
| 24h before | 51 | **13 (25%)** | 7773 (cap 8000) |
| since 21:43:00Z 07-28 | 10 | **0** | **10071** |

**10071 output tokens is structurally impossible under the old 8000 cap.** That is
the proof. The `0 truncated` is 0-of-10 against a 25% baseline (~2.5 predicted) —
*consistent* with success, but too small to carry the claim alone. **The truncation
share needs days.**

### Three landmines paid for, all still live

1. **Patch BOTH `fix-proposer` and `council-gate` — this is an OBSERVED near-miss,
   twice.** The gate is mirrored from fix-proposer; `transform_step` (`099:71-90`)
   rewrites only `error_step`/`input_fields`/`prompt_template` and copies
   `config.ai_service.max_tokens` **verbatim**. Another thread ran `099 --apply`
   **twice** since the change (07-28 22:37:14Z, roster 16→17 adding the
   `architecture` seat; and 07-29 08:03:59Z). Both times the value survived **only
   because it is on the mirror's SOURCE**. A gate-only patch would have been gone
   within the hour. **Any hand-tuned seat value must go to fix-proposer, or it has a
   half-life of hours.** Verify with the dry-run reporting `drift: (none)`.
2. **The obvious truncation detector counts ZERO.**
   `count(*) FILTER (WHERE output_tokens >= max_tokens)` is **structurally blind**:
   truncated rows have `output_tokens IS NULL`, so the comparison is never true.
   Count `error_message LIKE 'TOLERATED%'`.
3. **The config path is `config.ai_service.max_tokens`**, not `config.max_tokens`.
   A `->>` NULL means "wrong path" as readily as "unset" — I read one as "no seat
   sets max_tokens" and committed it wrong. Print the object once with
   `jsonb_pretty` minus the prompt instead of guessing. Logged in `WRONG_CALLS.md`.

**Do NOT "fix" the seed.** `0NN_council_gate.sql` carries `'max_tokens', 3000` at
nine sites — already stale against the live 8000 *before* this change, so replaying
it would cut every seat's budget. Live row + mirror are the source of truth.

## 3. The spawn-timeout class looks CLOSED — and why that is not luck

| | |
|---|---|
| reproducer (`build-pipeline-trigger`) spawns since 11:30Z 07-28 | **91** |
| timeouts fleet-wide in that window | **0** |
| last `%timed out after%` anywhere | **2026-07-28 11:29:57Z** (~21h) |

Against the morning rate (46–67%) 91 spawns predicted ~42–61 failures.

**The cause is named, not lucky:** `029`'s fix `CHASSIS_RESPONSES_START_AT=latest`
is **live on both replicas** (pod env + binary carries `NewConsumerFromLatest`). It
was "shipped dark" when `029` was last written; it is not dark now.

**The busy post-roll test — previously owed — has now been observed and it passes.**
Last night's window was worthless (six consecutive hours at **zero** reproducer
spawns). The 08:05:18Z roll to `v1.0.1197` landed into a busy fleet: **19
orchestrations and 4 reproducer spawns in the first 10 minutes, zero errors of any
kind.** Under the old regime every roll produced a post-roll burst. 4 spawns is a
thin control on its own — but it is a *real* one, and it sits on top of the 91/0.

### The one loose end, contributed into `029` (owned elsewhere — do not compete)

`CHASSIS_RESPONSES_START_AT` **exists only as cluster state**: present in the
Deployment spec, but absent from `deployments/`, absent from
`git log -S'CHASSIS_RESPONSES_START_AT'`, and absent from
`last-applied-configuration`.

**Updated 07-29: it SURVIVED a real deploy** (1196 → 1197), which narrows the risk —
this is not "any deploy". The exposure is a rebuild-from-overlay, a Deployment
recreate, or a fresh cluster: those yield a chassis **without** the fix, and nothing
in git says it is required. Given the mechanism (~2–3 h of response deafness per
restart, growing daily) that silently reintroduces a fleet-wide outage class.
**I am not claiming `apply -k` strips it** — env is a merge-keyed list and that path
is untested. Flagged in `bugs_open/029_…hung_spawns…`, not actioned.

## 4. Flip preconditions — 2 of 4, and this does NOT reopen the decision

| precondition | state |
|---|---|
| `003` closed | ✅ `bugs_closed/` |
| `124` fixed | ✅ `bugs_closed/` |
| `029` closed or accepted | ❌ **open** — ⚠ **number collision**: the phantom-tool-links `029` is CLOSED; the blocker is `029_…hung_spawns_saturate_dispatch_group…`. **Resolve by SLUG.** |
| Anthropic concurrency headroom | ❌ `[UNMEASURED]` |

The flip decision rested on the council's REVISE being sound and the urgency being
gone — **not** on the preconditions alone. Clearing two restores nothing. Reviving
it is an **owner call**.

## 5. What is actually left

1. **Watch the truncation share** over the next few days — the only genuinely owed
   item. One query, below. Success = the share stays near zero on a call count that
   matters (≥40 calls). If it creeps back, 16000 was not enough.
2. **Nothing else is owed by this thread.** The busy post-roll test is discharged
   (§3). The wrapper is proven and deliberately not default.
3. Optional, unowned, cheap: the other seats sit at 8000 with
   `guidelines` max 7727 / `guardian` max 7637 / `checkability` avg 6330 — pressed
   against the cap and likely to start truncating as submissions grow. Raising them
   is a **fleet-wide** change to the shared review apparatus and therefore an owner
   call, deliberately not taken.

## 6. The queries that matter

```sql
-- OWED: truncation share since the raise (count TOLERATED, never output_tokens>=max_tokens)
SELECT count(*) AS calls,
       count(*) FILTER (WHERE error_message LIKE 'TOLERATED%') AS truncated,
       max(output_tokens) AS max_out, max(max_tokens) AS cap
FROM llm_call_log
WHERE step_name='review_editquality' AND created_at > '2026-07-28 21:43:00+00';

-- the seat ceilings, at the CORRECT json path
SELECT key, (value->'config'->'ai_service'->>'max_tokens')::int
FROM agent_definitions, jsonb_each(default_config->'workflow'->'steps')
WHERE type='council-gate' AND is_active AND COALESCE(is_snapshot,false)=false
  AND deleted_at IS NULL AND key LIKE 'review_%' ORDER BY 1;

-- timeout class + its CONTROL (a zero is meaningless without the control column)
SELECT date_trunc('hour', created_at) AS hour, count(*) AS orch_runs,
       count(*) FILTER (WHERE owner_agent_type='build-pipeline-trigger') AS bpt
FROM orchestration_states WHERE created_at > now() - interval '12 hours'
GROUP BY 1 ORDER BY 1;

SELECT agent_type, step_name, count(*), max(occurred_at)
FROM agent_error_log WHERE error_message ILIKE '%timed out after%'
  AND occurred_at > now() - interval '12 hours' GROUP BY 1,2;
```

**Per-roll ritual** (do this after ANY roll you did not perform yourself):

```bash
# 124's landmine — per replica, because a partial roll can split the binaries
for p in $(kubectl get pods -n ai-persona-system -l app=agent-chassis -o name | sed 's|pod/||'); do
  printf '%s : ' "$p"
  kubectl exec -n ai-persona-system "$p" -- sh -c 'strings /app/agent-chassis | grep -c "unknown execution-context field"'
done
# 029's flag — it is NOT in git, so treat its presence as a fact to re-check, not assume
kubectl get deploy -n ai-persona-system agent-chassis \
  -o jsonpath='{range .spec.template.spec.containers[0].env[*]}{.name}={.value}{"\n"}{end}' | grep START_AT
```

## 7. Measurement discipline this thread paid for (the transferable part)

- **A zero without a control is not evidence.** This thread called "ABATED" wrong
  once on a quiet hour, and nearly repeated it on a quiet post-roll night. **Always
  print the control column.**
- **Prefer a positive control to an absence.** "No truncation" is weak; "10071
  tokens, impossible under the old cap" is decisive. Look for the observation the
  old world could not have produced.
- **Verify the numerator and denominator are the same population.** The truncation
  errors carry `agent_type='generic'` (the spawned *pod's* identity), not the
  orchestration owner — reading them as a separate population is the available
  denominator trap here. Join and check.
- **Use `agent_error_log`, never `orchestration_states`, for timeouts.** A timed-out
  awaited request does not reliably fail its orchestration.
- **Re-run filtered measurements UNFILTERED before believing a zero.** If a fix
  changed the error *text*, an `ILIKE` filter reports a clean fleet. Doing this is
  what surfaced the truncation class in the first place.

## 8. Files

`HANDOFF_2026-07-28_council_parallelism_thread.md` (full working record),
`bugs_open/029_…hung_spawns…` (open, owned elsewhere; my contribution at the foot),
`bugs_open/119` (a DIFFERENT defect — complete-but-invalid JSON, not truncation),
`bugs_closed/019` (the truncation case; its fix is what keeps rounds alive),
`0NN_council_gate_orchestrator.sql`, `097_TRIGGER_council_review_v1.sh`,
`099_SYNC_gate_roster.py`, `WRONG_CALLS.md` (four entries from this thread).

---

## 2026-07-29 ~08:45Z — watch check (session "council parallelism 3")

**Per-roll ritual: CLEAN on `v1.0.1198`.** The fleet rolled again at 08:19Z
(image `3730e90d…`, not the `458c9ad0…` this file was written against — a roll
this thread did not perform). Both replicas carry 124's marker (`1` each);
`CHASSIS_RESPONSES_START_AT=latest` survived a **second** real deploy, further
narrowing 029's residual risk to rebuild-from-overlay / recreate / fresh-cluster.

**The owed truncation watch: 11 calls since the raise, 0 truncated, max output
10071 (cap 16000).** Only one new call since this handoff was written — councils
were quiet overnight. The ≥40-call bar is not met; **the watch stays open.** The
positive control (10071 > old 8000 cap) remains the load-bearing evidence.
Nothing owed changed.

**Timeout class: still closed.** 0 `%timed out after%` rows in 12h against a live
control (38 `build-pipeline-trigger` spawns, 40–129 orch/hour). Re-ran the sweep
UNFILTERED per §7 — no renamed timeout error hiding behind the ILIKE; the only
truncation-family rows are the council-seat ones below.

**Roster is now 17 seats** (architecture added 07-28 22:37Z, per §2's landmine 1
— and that seat was raised to 16000 by its own thread at 07:19:36Z on 07-29,
correctly on BOTH `fix-proposer` and `council-gate`; landmine 1 respected).

**§5.3 has moved from prediction to observation:** `review_prior_art` truncated
2 of ~58 calls in 36h at its 8000 cap (20:16:18Z, 21:50:32Z on 07-28). Evidence
contributed into `bugs_open/138` (owned by the architecture_review workstream —
that file also now records its architecture-seat fix as EXERCISED, verified by
this session: 2 calls at cap 16000, 0 truncated). Raising the remaining seats
stays an owner call, per §5.3 — now with a live instance to point at.

## 2026-07-29 ~09:35Z — OWNER RULED on §5.3: prior_art + guidelines raised to 16000

**The call was put with 14-day per-seat evidence and the owner chose the narrow
option** — raise only the two seats with *observed* truncation losses; guardian
and bug_historian stay at 8000 until they actually truncate:

| seat (gate+generic populations) | calls 14d | truncated | p95 out | max out |
|---|---|---|---|---|
| `prior_art` | 224 | **7 (3.1%)** | 7,129 | 7,900 |
| `guidelines` | 154 | **7 (4.5%)** | 6,953 | 8,000 (cut) |
| `guardian` — NOT raised | 287 | 1 (0.3%) | 6,959 | 7,934 |
| `bug_historian` — NOT raised | 198 | 0 | 6,499 | 7,531 |

**Applied ~09:35Z** with the same guarded `jsonb_set` as the editquality raise
(HANDOFF_2026-07-28 §"the fix"), to BOTH `fix-proposer` and `council-gate`
(landmine 1): each UPDATE returned 2 rows individually. **099 dry-run:
`drift: (none)`** — the mirror is a no-op, so a future `--apply` cannot revert
this. Roster state: 4 seats at 16000 (architecture, editquality, guidelines,
prior_art), 13 at 8000, gate ≡ proposer throughout.

**The watch now covers three seats.** editquality is at **18 calls / 0
truncated** since its 21:43Z raise (traffic resumed ~09:00Z; the single 8000-cap
call at 21:43:54Z was an in-flight pre-cutover round — spawn-carries-config, see
138). For prior_art (3.1% baseline) and guidelines (4.5%) a zero needs MORE
calls than editquality's 25% baseline did to mean anything — expect **days**;
the decisive observation for any of them is a single output >8000.

**Caveat on populations:** no `generic` (spawned-pod) council has run
editquality since 07-26 14:46Z, so the raised caps are unexercised in that
population. The config is shared — one definition row, two pod identities
`[INFERRED from the architecture seat's pre-raise truncations logging as
agent_type='generic' off the same rows]` — but the first spawned council after
the raises is worth a glance.

**Out of this lane, flagged in 138, not actioned:** the experience-approval
council's `review_deferral_honesty` truncated **3 of 5 calls at cap 12000** —
the same mechanism family in a different apparatus, owned by the experience-loop
workstream.
