# HANDOFF — `bugs_open/201`, 2026-08-07 · **start here.** Supersedes `HANDOFF_2026-08-05_continue_here.md`

> ## ✅ UPDATE 08:40Z — THE PROOF LANDED. **Both symptoms are now FIXED, LIVE and PROVEN. §3 below is history.**
>
> Attempt 2 (corr `78e15724…`) did it. The handler ran, rebuilt the page, and completion was
> **refused**:
>
> ```
> completion blocked: post-fix verification found the defect still present:
> 18 finding(s) still present across 3 component(s); first: slot "news-listing"
> field "items[1].summary" pattern bold in content_data — "**the `animation`**"
> ```
>
> That is `VerifyLiteralMarkdownResolved`'s own `Detail` verbatim — the verifier was consulted,
> returned `Resolved:false`, and blocked the stamp. **Before this change that run would have been
> marked `complete`**, exactly as the gaswholesalers item was.
>
> **The artefact confirms both halves:** all three components rewritten at 08:37:26Z against the
> baseline in §3, and `news-listing` **still carries the markdown**. So the repair genuinely ran
> and genuinely failed — and was caught.
>
> **⚠ SECOND FINDING, and it is `bugs_open/184`'s, not this lane's:** symptom 1's fix makes the
> dispatch *work*; it does not make the repair *effective*. `page-content-writer` wrote markdown
> straight back into the field it was dispatched to clean. Contributed to 184, which is now
> unblocked and has a different failure than before — visible instead of silent.
>
> **Nothing is outstanding on 201 except `RFC_017`** (the generic fail-open policy — owner
> decision, not a blocker). Item `efaa39a2…` is terminally `failed` at 3/3 → human review, which
> is the correct destination; a further attempt would need `attempt_count` reset too.
>
> Per the owner ruling of 2026-08-06, `bugs_open/201` **stays in `bugs_open/`** although fixed.

**Both symptoms are fixed, council-approved and deployed.** ~~ONE proof is outstanding, and it is
the only work left.~~ **Proof landed 08:40Z — see the box above.** Nothing below needs re-deriving.

## 1. State in one table

| | fix | council | deployed | behaviour proven |
|---|---|---|---|---|
| **Symptom 1** — direct `page-content-writer` dispatch hard-fails 11/11 | `37afbb847` | APPROVED | ✅ | ✅ **yes**, 08-06 11:34Z |
| **Symptom 2** — completion stamped on the handler's word | `dc4f4e6b2` + `7e62f4a07` (r2) | **APPROVED r2** | ✅ `v1.0.1262`, pod-verified 4/4 both replicas | ❌ **no — this is the job** |
| Migration **331** (claim-timeout exclusion) | — | — | ✅ applied, verified at the live column | — |

**`bugs_open/201` STAYS OPEN** on symptom 2's behavioural proof.

## 2. What symptom 2's fix actually is

`check_literal_markdown.go` registers **`VerifyLiteralMarkdownResolved`** on the estate's existing
`ItemVerifier` registry, which `CompleteWorkItemAction` already consults before stamping
`complete`. Reuse of a declared-gap mechanism — **not** a new seam, and nothing about
`complete_work_item`'s general trust of `handler_result` changed.

Three things a successor should not re-litigate:

1. **Whole-page scope is deliberate** and matches `page-build-handler`'s remit (it rewrites all
   the page's spec sections). Narrower would be wrong; stricter is the `page_rerender` trap that
   stranded 1,849 items.
2. **A zero returns `Resolved:false`, NOT an error.** r1 returned an error and the council gated
   the round on it: the registry **fails open on error**, so the error branch stamped `complete`
   on exactly the content-loss case the verifier existed to prevent.
3. **Registering a verifier does not make it gate** — the claim-timeout sweep walks past it at 15
   minutes unless the type is excluded in **both** `220_*.sql` and the live `pre_query`. Both done.

## 3. THE OUTSTANDING PROOF — method, and where the last attempt got to

**Goal:** get one `literal_markdown` item to *completion* on `v1.0.1262`+ so the verifier is
actually consulted. Either outcome proves it ran:

- repair worked → markdown gone from `content_data`, md5 changed, `Resolved:true`, item `complete`;
- repair wrote nothing → **completion REFUSED**, item to attempts rather than `complete`. *This is
  the outcome that demonstrates symptom 2 is closed.*

**Target: `webdesign.co.uk` / page `news` (`c60e79f9-a5e0-4b88-b26f-db6f3a844c73`) / slot
`news-listing`.** Work item **`efaa39a2-7bd9-4b2d-9e25-669245d28e46`** — one of 201's own 11
original failures, already re-armed onto `page-build-handler`.

**Artefact baseline (unchanged as of 08:28Z — the page has NOT been altered by any attempt):**

```
hero           | md5 3e77770ccea4619f6d7b8c78c733e3a9 |    304 B | 08-05 14:21:41Z | bold_md=false
news-listing   | md5 9df6c43d4eab12ab5600ca0f760daacc | 10 232 B | 08-05 14:21:41Z | bold_md=TRUE
call-to-action | md5 45f6f2b8154d441af30e7c334f7c8af1 |    331 B | 08-05 14:21:41Z | bold_md=false
```

**Attempt 1 (corr `c0c88fcb…`, 08:25Z) — INCONCLUSIVE, not a pass and not a refutation.**
The item reached `failed` at attempt 1/3 with
`step call_content_writer failed: workflow completed but its result could not be delivered to the
parent (failed_transient): message validation failed (code: CHILD_ORCHESTRATION_FAILED)`.
That is the known **spawn→call handshake race** (fails ~half the time, fleet-wide). The run died
**upstream of `complete_work_item`**, taking `mark_failed`, so **the verifier was never
consulted.** ⚠ It would be easy and wrong to report "the item wasn't stamped complete, so symptom
2 works" — it wasn't stamped complete because the workflow failed, not because the verifier
refused. **No damage: the artefact is byte-identical to baseline.**

**Attempt 2 (corr `78e15724-4efd-4b52-80a8-0654847c8bd2`) was in flight when this was written.**
`attempt_count` was deliberately left at 1 rather than reset, so the real attempt history shows.
**Check its outcome first**, and re-read the artefact against the baseline above.

### How to run another attempt

```sql
-- re-arm ONE item (status only; leave attempt_count as the honest record)
UPDATE site_work_items SET status='triaged', claimed_by=NULL, error=NULL, updated_at=NOW()
 WHERE id='efaa39a2-7bd9-4b2d-9e25-669245d28e46' AND status='failed';
```
then dispatch `build-dispatch-loop` — envelope in this lane's
`TRIGGER_fire_quality_discovery.sh` (adapt: spawn `build-dispatch-loop`, `input_data` needs BOTH
`site_id` and `domain`). ⚠ **`attempt_count` is 1 (or 2) of 3** — when it hits 3 the item is
exhausted and a further re-arm must reset `attempt_count` too.

### Four traps, each of which produces a confident wrong answer

1. **A re-arm MUST also set `handler_agent='page-build-handler'`** if the row still carries
   `page-content-writer` — otherwise it re-runs the pre-fix route and looks like the fix failed.
   (Already set on this row.)
2. **⚠ Do NOT canary on `gaswholesalers.com/how-pricing-works`** — it is the envelope-guard page
   (`type='text'` + `jsonb_typeof(result)='string'`), so any verdict is uninterpretable. **It is
   also the row 201's own §Symptom-2 evidence rests on** — flagged there, not overturned.
3. **"Still `claimed`" is not a hang.** Migration 331 removed `literal_markdown` from the
   claim-timeout auto-complete, so a stuck item no longer self-resolves at 15 minutes. By design.
4. **Read the orchestration rows by correlation, not by clock.** A *different*
   `build-dispatch-loop` COMPLETED 19s into attempt 1 — the scheduled `build-pipeline-trigger`'s
   own instance, on another site. A correct query read carelessly says "the run finished
   instantly".

## 4. Also open, and NOT blocking 201

- **`architecture_review/RFC_017`** — the registry's **fail-open-on-error** policy. Two council
  seats said the symptom-2 fix routes *around* it; every other registered verifier can still error
  and be waved through to `complete`. Four options costed. The deciding number is
  `[UNMEASURED]`: **how often verifiers actually error in production.** Owner decision.
- **`bugs_open/208`** — another thread's (fixed `cb7b4d759`, PBP-036). Do not pick it up here.
- **`ai-agent-orchestration.com` rebuild** — scoped, not started:
  `site_ai_agent_orchestration/HANDOFF_2026-08-05_rebuild_scope.md`.

## 5. Lane files

`PLAN` (the decision + three rejected shapes, with a visible correction) · `RUNBOOK` (R1–R7b; R6
has the verification traps, R7 explains why a pod-grep could not verify symptom 1 but **can**
verify symptom 2) · `NOTES` (evidence and every misstep, newest at the bottom) ·
`README_where_we_are` (owner's plain prose) · `SUMMARY_2026-08-06` ·
`TRIGGER_fire_quality_discovery.sh` (detect-only; the house trigger `075_trigger_discovery.sh`
**must not be used** — it triages a hardcoded other domain's backlog).
