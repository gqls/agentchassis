# PLAN — 2026-08-18 — `bugs_open/029`: the retry replay kills the child that was still working

> ## ⚠ CORRECTED 2026-08-18, later the same day — THE TITLE OF THIS PLAN IS WRONG
>
> **The replay does NOT kill a live child. It re-executes one that had already frozen
> ~7–10 minutes earlier**, duplicating a non-idempotent spawn and resetting the 4-hour
> reaper clock. The takeover arm *cannot* fire on a healthy child, because every
> `UpdateStateWithVersion` bumps `last_activity` — so >5 min of staleness is the
> precondition, not the consequence.
>
> Caught by the Fable design pass (briefed to contradict the diagnosis) and then verified
> here: 17 of 18 wedged children carry **two** spawn-init `awaited_requests` rows for the
> same step; the 18th carries one and is exactly the outlier this plan's own evidence
> reported and kept. Full working: `NOTES_retry_kills_live_child.md`, final section.
>
> **Survives:** the window truncation (worse than stated — it also fires one level down,
> inside the child), the 33-step blast radius, the 4-hour unreachability, and that the
> replay is destructive. **Withdrawn:** that it destroys *live work*.
> **Now the centre of 029, and OPEN:** what kills the child's continuation after its first
> spawn handshake. Headings below are left unedited so the correction reads against what
> it corrects.

**Lane opened 2026-08-18.** Owner thread: this one (session "bugs_open/029 hung spawns").
The owner started 029 here deliberately and told the `site_ai_agent_orchestration`
lane so, which is blocked on it (`6731ffa31`, 2026-08-17: *"propagation BLOCKED by
bugs_open/029"*).

## Why this lane exists, in one paragraph

`bugs_open/029` has been open since 2026-07-19 and has been through two framings.
The first — *hung spawns saturate the `dispatch` concurrency group* — was **refuted**
in the file itself on 2026-07-21. The second — *029 is `bugs_open/003`'s blast radius,
delivered through claimed-item starvation* — is right about the damage but stops short
of naming what makes a child hang **now**, after 003's F2/F3 retry work shipped. This
lane's claim is that there is a third, sharper mechanism, and that it is **self-inflicted
by the retry machinery itself**.

## The claim

~~**The parent's own retry kills the child that was still legitimately working.**~~

**WITHDRAWN — see the banner.** Corrected claim: **the parent's retry RE-EXECUTES a child that has already frozen**, duplicating a non-idempotent spawn (a real `page-rerender` agent and K8s job) and re-stamping `last_activity`, which resets the 4-hour reaper clock so the corpse outlives its own first death by another four hours.

The chain, as measured (evidence in `NOTES_retry_kills_live_child.md`):

1. A step declares `config.timeout_seconds` — `call_dispatch` declares **900s**.
2. The callee legitimately takes longer than that a few percent of the time.
3. The parent times out and **replays** the original request (correct in itself —
   `bugs_closed/129` established replay-not-reconstruct, and that must not regress).
4. But the retry's window is **silently recomputed**, not reused: capped to **5 minutes**,
   and dropped to **3 minutes** when the declared timeout exceeds 30 minutes. So the
   longer a step declares, the *shorter* its retry gets — an inversion.
5. That truncation drives the parent to retry_version 3 in ~25 minutes instead of ~60.
6. The final replay lands on a child that is **still running**, and **~15 seconds later
   the child freezes** in `EXECUTING_STEP`, mid-spawn.
7. A row frozen in `EXECUTING_STEP` with nothing in its awaited set is invisible to
   `TimeoutMonitor` and to the retry driver. Only the reaper's *4-hour* arm touches it.
8. Meanwhile the site's work item stays `claimed`, so that site is undispatchable until
   `claimed-item-timeout` resets it at 40 minutes.

## Three separable defects — do not conflate them

| | defect | shape | urgency |
|---|---|---|---|
| **A** | retry window truncated to 5 min (3 min if declared > 30 min) | config honoured on attempt 0, silently overridden on every retry | this is the *driver* — it manufactures the premature v3 |
| **B** | a replay delivered to a **live** child wedges it | no liveness check before replaying | this is the *damage* |
| **C** | a row frozen in `EXECUTING_STEP` is unreachable for 4 hours | no re-drive path; only the reaper's slowest arm | this is the *dwell time* |

**A is the cheapest and has the widest blast radius; B is the one that actually destroys
work; C is why one instance costs hours rather than minutes.** A plan that fixes only A
makes the bug rarer without making it safe. A plan that fixes only B leaves the estate
retrying on a window nobody configured.

## Blast radius of A alone `[MEASURED 2026-08-18]`

**33 live workflow steps across 25 agent types** declare `timeout_seconds > 300`, so every
one of them has its retry window silently shortened. Declared values run 600, 900, 1200,
1800, 2100, 3600, 21600, 43200 and **86400** — that last is a human-approval step, which
on retry would be given **three minutes** for a human to answer.

## Decisions and their reasons

- **Diagnosis loop filed BEFORE asserting the cause**, per CLAUDE.md's "Diagnosis before
  debugging" — the claim is structural, cross-cutting, and changes behaviour fleet-wide,
  which is squarely the "always file" list. Intake correlation
  `0e4f89fb-fb46-4bb3-a658-aa939713fd88`, run correlation
  `c8312dce-db45-4554-b2ab-5ac50e7e0c8a`. **A REFUTED verdict is a success here** and will
  be recorded as a visible correction, not quietly dropped.
- **Not filing a new bug number.** This is 029's own mechanism, sharpened. Forking a second
  account that drifts is exactly what the working-docs rules forbid. It goes into
  `bugs_open/029` as a contribution with its own dated section.
- **Fix design routed through Fable** at the owner's instruction, briefed to prefer a
  framework-level fix over the individual case, and explicitly asked to contradict the
  diagnosis if the code does not support it.

## Open questions (to be closed by the diagnosis run and Fable's read)

- **Why** does the replay freeze the child? Candidate: two concurrent drivers of one
  orchestration row racing on the optimistic lock, the replay winning and doing nothing
  useful, the real worker's update lost. `[UNVERIFIED]` — this is a hypothesis about the
  mechanism of B, not part of the measured claim.
- Is suppressing a retry while the child is demonstrably alive (the child row's
  `last_activity` is a cheap signal already in the schema) better than extending the await?
- Does `RSH-004`'s local-action deadline (7200s default, `bugs_closed/169`) interact here?
  The wedged rows sat past 2h without it firing, which needs an explanation either way.

## Guardrails taken from neighbouring lanes

- `bugs_closed/129` — **replay, never reconstruct.** Any fix that goes near
  `handleRecoverableError` must not regress this.
- `bugs_closed/169` trap 1 — *"`timeout_seconds` is read by no action; grep for the READER,
  not the key."* True of **action handlers**, and not in tension with this lane: the
  *coordinator* does read it, via `ConvertStepTimeout → step.Timeout → getTimeout →
  awaited.TimeoutAt`. The live proof is that attempt 0's window is exactly the declared
  900s. Stated here because the two claims look contradictory and are not.
- `bugs_open/294` — no reaper arm for `RUNNING`. Adjacent to defect C; different status,
  same shape of hole.
