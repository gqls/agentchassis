# 138 — a TRUNCATED advisory review silently becomes a BLOCKING one: 17 council rounds in 14 days were revised on token budget, not judgement

**Filed** 2026-07-29 by thread "bugsearch 5", found while verifying the
`review_architecture` seat it had shipped hours earlier (`FIX-054`) ·
**Status** OPEN, unowned · **Severity** medium — no correctness risk, pure waste
and a misleading signal: each occurrence costs a full council round (credits +
~30 min latency) and returns `revise` for a reason no reviewer intended ·
**Distinct from `bugs_closed/076` and `bugs_open/119`** — see below, this is
neither of them.

---

## Symptom

A council round returns `revise` with

```
decided_by : gating objection from architecture
```

while every objection that seat raised is `"severity": "medium"` — and the seat's
own prompt tells it that medium is *"recorded and returned to the author without
blocking"*. Nothing in the verdict says why it actually gated.

## Root cause — a correct carve-out with an uncosted consequence

Three mechanisms compose, and each is individually right:

1. A reviewer's LLM call exceeds `max_tokens`. Logged, but as a **tolerated**
   failure:
   ```
   llm_call_log.success = false
   error_message = "TOLERATED (step continued on the partial): response truncated:
                    stop_reason=max_tokens (output_tokens=8000 reached the configured
                    cap, 6112 chars recovered); raise max_tokens or shorten the prompt"
   ```
2. `tolerate_truncation: true` recovers the partial JSON, and the council marks
   that review **`degraded: true`**.
3. `diagnose_council_decide_action.go:684` — **a `Degraded` object gates
   unconditionally**, ignoring severities:
   ```go
   if r.Degraded || len(r.Objections) == 0 {
       return true   // gates
   }
   ```
   The reasoning is sound and documented at `:675-677`: a high-severity objection
   *may have been cut off before we ever saw it*, so a truncated objection cannot
   be waved through on the severities that survived.

**So the defect is not in any of the three.** It is that (3) converts a *token
budget overrun* into a *forced revise*, and nothing anywhere measures or surfaces
that. An advisory seat that runs long becomes a blocking seat, silently, and the
verdict names the seat rather than the truncation.

## Evidence — measured, all-seats, 14 days

```sql
WITH r AS (
  SELECT rev->>'reviewer' AS reviewer, rev->>'verdict' AS verdict,
         COALESCE((rev->>'degraded')::boolean,false) AS degraded
  FROM diagnosis_artifacts d, LATERAL jsonb_array_elements(d.body::jsonb->'reviews') rev
  WHERE d.kind='council_report' AND d.created_at > now() - interval '14 days'
)
SELECT reviewer, count(*) AS reviews,
       count(*) FILTER (WHERE degraded) AS degraded,
       count(*) FILTER (WHERE degraded AND verdict='object') AS degraded_objections_that_gate
FROM r GROUP BY 1 HAVING count(*) FILTER (WHERE degraded) > 0 ORDER BY 3 DESC;
```

| reviewer | reviews | degraded | degraded objections that GATE |
|---|---|---|---|
| editquality | 241 | 14 | **9** |
| prior_art_librarian | 200 | 6 | **4** |
| guidelines | 142 | 4 | 0 |
| **architecture** | **3** | **2** | **2** |
| checkability | 3 | 1 | **1** |
| guardian | 261 | 1 | **1** |
| tooling_provenance | 110 | 1 | 0 |

**17 forced revises in 14 days.** The absolute rate is low per seat (editquality
5.8%) but the cost per occurrence is a whole round, and the *rate is not the
point* — the point is that the verdict is unattributable to its real cause.

**`architecture` at 2 of 3 (67%) is the outlier and is my own doing** — a longer
prompt demanding four-dimension reasoning against the same fleet-wide
`max_tokens: 8000`. Already fixed for that seat (below); the mechanism is
untouched.

## Why this is NOT `bugs_closed/076`

076 asked whether *consumers check the truncation marker* and found them already
guarded; its fix is what **added** the `Degraded` carve-out in
`diagnose_council_decide_action.go`. That carve-out is working as designed. 076
made truncation *visible to the decider*; it did not ask what the decider then
does with an advisory seat's round, and the answer — block it — was never costed.
**This is 076's fix behaving correctly and having a consequence nobody measured.**

## Why this is NOT `bugs_open/119`

119 is a seat emitting **structurally malformed** JSON → `unreadable` → the round
is voided with `decided_by: unreadable reviewer(s)`. Here the JSON is **well-formed
but incomplete** → `degraded` → the round is *decided*, by a seat that did not
intend to decide it. Different field, different decided_by, different fix. They
are siblings in one family ("a seat's output problem costs a round") and each
needs its own fix.

## Fix candidates, ordered by what closes the door

1. **Make the cause visible in the verdict.** `decided_by` should distinguish
   `gating objection from X` (a judgement) from `gating TRUNCATED objection from X`
   (a budget overrun). Cheapest, no behaviour change, and it stops the wrong
   conclusion being drawn — which is the live harm, because a high object-rate
   with no signal line is *also* the documented kill-switch for retiring a seat.
   **A seat could be pulled for being noisy when it was being cut off.**
2. **Alert on the rate.** The query above is a one-line check; nothing runs it. A
   seat whose degraded rate crosses a threshold is misconfigured, not noisy.
3. **Right-size `max_tokens` per seat.** Every seat currently sits at 8000
   regardless of how much its prompt asks for. The two seats with the longest
   analytical remits (editquality, prior_art_librarian) are also the two with the
   most degradations — that is not a coincidence and it is measurable per seat.
4. **Emit the load-bearing field FIRST in every seat's output schema.** Truncation
   eats the tail, so whatever must survive belongs at the head. Generalises the
   fix already applied to `review_architecture`.

Not recommended: relaxing the `Degraded` gate. It is the conservative half of 076
and removing it would let a cut-off high objection through silently — strictly
worse than a spurious revise.

## What was already done (2026-07-29, this thread, `review_architecture` only)

- `max_tokens` **8000 → 16000** on `fix-proposer`, mirrored to `council-gate`.
- **`notes` moved ahead of `objections`** in the output schema. The mandated
  `ARCHITECTURE_SIGNAL` lives in `notes`, which was emitted **last** — so
  truncation destroyed exactly the field that makes the seat measurable, and the
  result was indistinguishable from "the seat is noise, pull it".
- An explicit length budget in the prompt, naming this bug's mechanism so the
  seat knows why brevity is a correctness constraint and not a style note.

**This fixes one seat. It does not fix the mechanism** — the other six seats in
the table above are unchanged.

> ## ✅ THE SEAT FIX IS NOW PROVEN (2026-07-29 ~12:30). The MECHANISM is still open — that is what keeps this bug open.
>
> 12 further `review_architecture` reviews have run since the cutover.
>
> | | before fix | after fix |
> |---|---|---|
> | reviews | 3 | **12** |
> | truncated (`stop_reason=max_tokens`) | 2 | **0** |
> | `degraded` | 2 (67%) | **1 (8%) — and it is explained, below** |
> | carries `ARCHITECTURE_SIGNAL` | 1 (33%) | **11 (92%)** |
> | verdict object | 2 of 3 | **2 of 12** |
>
> `llm_call_log`: all 12 calls at `max_tokens=16000`, all `success=t`, **zero
> truncations**, peak output **4,443 tokens — 28% of the cap**. Note the outputs got
> *shorter*, not just the ceiling higher: the length budget in the prompt did as much
> work as the extra headroom.
>
> **The single remaining `degraded` is not a residual defect.** It belongs to
> orchestration `815b38c3`, spawned **07:16:59 — before the 07:19:36 cutover** — so
> it carried the old 8,000 config. An orchestration keeps the workflow definition it
> loaded at spawn. **Among rounds spawned after the cutover: 0 of 11 degraded.**
>
> **What this settles beyond the seat:** the object rate fell from 2-of-3 to 2-of-12
> the moment it stopped truncating. The seat was never noisy — **it was being cut off,
> and a degraded review gates unconditionally.** That is the confounder this bug is
> named for, now demonstrated rather than argued: acting on the raw object rate would
> have retired a seat that was working.
>
> **STILL OPEN, and it is the whole bug:** the other six seats in the table above are
> unchanged, all still at `max_tokens: 8000`, and nothing surfaces a degraded gate as
> distinct from a judged one. Fix candidates 1–4 stand.

~~**AND IT IS NOT YET PROVEN.**~~ *(Discharged — see above.)* Config cutover was **2026-07-29 07:19:36Z**
(`agent_definitions.updated_at`, not guessed). At the time of writing **no council
round has spawned since**, so `llm_call_log` still shows `max_tokens = 8000` on
every `review_architecture` call. That is **not** the change failing — both
observed calls belong to rounds spawned at 07:16:59 and 07:18:26, i.e. **before**
the cutover:

> **An orchestration carries the workflow definition it loaded at SPAWN.** A live
> config edit does not reach a round already in flight. So "DB config is live
> immediately" is true of the row and false of any running orchestration — which
> also means a config change cannot corrupt a round that is already going, and that
> a verification looking at in-flight rounds will read as a failed change.

**To close this out, confirm on a round spawned after the cutover:**
```sql
SELECT count(*) FROM llm_call_log WHERE step_name='review_architecture' AND max_tokens=16000;
-- must become > 0; until then the fix is applied but UNEXERCISED
```
and then re-check that seat's degraded rate is 0 over its next handful of reviews.

## How to verify a fix

Induce it rather than wait: set a seat's `max_tokens` low (say 500) on a scratch
agent, run one round, and confirm the verdict distinguishes truncation from
judgement. A green round proves nothing here — **the failing branch is the whole
bug** ([[verify-the-failing-branch]]).

Re-run the evidence query above; the count should not grow for a seat after its
budget is right-sized.

---

## 2026-07-29 ~08:45Z — the seat fix is now EXERCISED (contributed by the council-parallelism thread; this bug stays yours)

Ran the owed check while doing this thread's truncation watch:

```
SELECT count(*) FROM llm_call_log WHERE step_name='review_architecture' AND max_tokens=16000;
-- returns 2
```

Two `review_architecture` calls have run under the raised cap — 07:44:23Z
(output 3773) and 08:13:03Z (output 1308), **neither truncated**. The
"applied but UNEXERCISED" caveat above is discharged: the config took, on rounds
spawned after the 07:19:36Z cutover.

One more truncation DID occur after the cutover — 07:26:06Z, but at **cap 8000**:
that round spawned pre-cutover, so it confirms the spawn-carries-config note
above rather than contradicting the fix. Full per-call record:
3 truncations at 8000 (22:10:52Z, 22:18:27Z on 07-28; 07:26:06Z on 07-29),
then clean at 16000.

Still owed (yours): the degraded-rate re-check "over its next handful of
reviews" — 2 clean calls is not a handful, and neither output has yet exceeded
the OLD cap, so there is no positive control that 16000 is *enough*, only that
the row is live.

**New evidence for fix candidate 3** (right-size per seat): `review_prior_art`
has started actually truncating, not merely pressing the cap — 2 of ~58 calls in
36h at 8000 (20:16:18Z and 21:50:32Z on 07-28, both `TOLERATED%` in
`llm_call_log`; the 21:50Z one produced a `damaged by a TRUNCATED response` row
in `agent_error_log` at 21:55:47Z). That matches this file's 14-day table naming
prior_art_librarian second-worst, and moves it from "likely to start" to
"observed".

## 2026-07-29 ~09:35Z — candidate 3 actioned for two more gate seats (owner call, council-parallelism thread)

The per-seat 14-day measurement (`llm_call_log`, gate + generic populations
combined): `prior_art` 7 truncations of 224 calls (3.1%), `guidelines` 7 of 154
(4.5%), `guardian` 1 of 287, `bug_historian` 0 of 198, all other 8000-cap seats
clean with p95 ≤ 6,960. **Owner ruled: raise `prior_art` + `guidelines` 8000 →
16000, leave guardian/bug_historian until they actually truncate.** Applied to
both `fix-proposer` and `council-gate` (guarded `jsonb_set`, 2 rows returned
each; 099 dry-run `drift: (none)` after). Roster now: 4 seats at 16000, 13 at
8000. Note this right-sizes three of your table's seven seats — **the mechanism
(Degraded gates unconditionally, cause invisible in `decided_by`) is still
untouched; candidates 1, 2 and 4 remain open.**

**A data point from outside the council gate, same mechanism family:** the
experience-approval council's `review_deferral_honesty` truncated **3 of 5 calls
in 14d at cap 12000** (llm_call_log, `agent_type='experience-approval-council'`)
— worst rate of any seat anywhere, already above 8000, which is direct evidence
for your candidate 4 (schema order / length budget) over pure cap-raising.
Owned by the experience-loop workstream; flagged here because this file is where
the mechanism lives, not actioned by me.
