# RESUME HERE — consolidation programme + gripper honesty lane

**Written 2026-07-28 ~21:00Z**, at the end of a long session, for a fresh chat.
Anchor: `features_open/024`. Sibling cold-start:
`robot_hands_gripper_dossier/HANDOFF_RESUME_gripper_dossier.md` (the pilot itself).

---

## 1. State in one line

The gripper pilot is **proven and live**; a **live honesty defect in it was found
and fixed today and is now running on v1.0.1194**; the consolidation programme's
A1 has been **retired as a won't-do** on evidence; and the one genuinely
high-value open item is **adopting `platform/mailer` + `platform/httpguard` into
`tools-api`**, which is another thread's code.

## 2. What is LIVE right now (verified, not assumed)

**Chassis `v1.0.1194`**, two replicas, started 2026-07-28 20:48Z. Verified by
pod-grep on strings my change **created**, with both controls:

| marker | count |
|---|---|
| `carries no prose_sections` (created) | 1 |
| `carries no no_match_sentence` (created) | 1 |
| `No gripper in this index` (positive control) | 3 |
| `nonexistent_marker_xyz` (negative control) | 0 |

So commit `7f87c0afa` is in the running binary. **Re-run this after any roll you
did not do** — a retag is not a rebuild (1188/1189 shared an image id built 56
minutes before the fix in it).

## 3. The live defect that was fixed — read this before touching the scorer

```go
// BEFORE — the bug
if spec.Tech == "soft" && spec.GripMinMM != nil && spec.GripMaxMM != nil {
    // compare part size against the published cup range
} else {
    // "Not applicable — surface hold, no jaws"
}
```

Reads as obviously right. To the scorer it said: **when the range is unpublished,
this gripper has no size constraint at all.** The `else` was written for vacuum
and magnetic grippers, which genuinely have no jaws. A soft gripper *has* a size
window — we merely do not know it. So a gripper whose cup range the manufacturer
never published scored **`Match`**, and a paying customer would have been told the
part fits. Six of seven criteria in that file set `state.unknown` on an absent
figure; this one did not.

Fixed to flag `unknown` like its siblings. Transferable pattern in **`016b` §9**
(*"a `!= nil` guard around a criterion turns 'we don't know' into 'there is no
rule'"*) — it generalises to any gate over optional data.

Two more in the same commit: the gate's contract now travels **with the scoring
data** (`no_match_sentence`, `prose_sections`) instead of being a package const a
second report type would silently inherit; and `prose["summary_html"].(string)`,
the file's only unchecked assertion — reachable **only** on the `match_count==0`
path — no longer panics where the gate matters most.

### The test that found it, and the correction it forced

`TestUnknownNeverPasses` — for every subset of the six nullable figures, across
every fixture and four request shapes, removing a published figure may never turn
a non-passing candidate into a passing one, nor raise headroom. 2^6 × 10 × 4,
under a second. **It found the defect on its first real run.**

> **Its first version was wrong and the code was right.** I asserted rank could
> only *worsen* when a figure was removed. It failed at once: `No match` →
> `Insufficient data`. That is honest — without the figure we cannot assert
> failure either. **Losing information moves a candidate toward uncertainty in
> BOTH directions.** The invariant is *"uncertainty is never mistaken for
> success"*, not *"uncertainty is bad"*. The stronger version is permanently red.

## 4. ⚠ OPEN: the council said REVISE — this is the next task

**Corr `721ac4f7-2076-4fea-9242-b234cfe648d6` → `revise`**, gating objection from
`bug_historian`, no veto. **The commits carry no `Council-Reviewed:` trailer and
must not gain one.**

> **CORRECTED 2026-07-28 (next session) — this section named the WRONG objection
> as the gating one, and the remediation below was built on it.** The seat is
> right and the rest is not. `bug_historian`'s only **high** is on **edit 3** and
> is about **ENFORCEMENT**: *"the plan never establishes what the CALLER does with
> those violations… if logged/recorded but not used to block report delivery,
> this is exactly the documented shape in `bugs_open/079` and `bugs_open/083`."*
> The DECLARED CONTRACTS concern below is real but **medium**, and came from three
> OTHER seats (`guidelines` ×2, `prior_art`, `guardian`). This paragraph fused the
> gating seat's *name* to a different seat's *content*.
>
> Caught by printing all ten seats' verdicts + severities in one query rather than
> trusting this prose. Had round 2 followed the brief as written it would have
> answered four mediums and left the high untouched. **`decided_by` names the
> SEAT, not the objection** — full entry in `WRONG_CALLS.md` (2026-07-28).
>
> **Answered in round 2, resubmitted on the same corr:** enforcement already
> exists — the action returns `(nil, error)` (`verify_report_prose_action.go:135-139`),
> the engine is fail-CLOSED when no `error_step` exists (`coordinator.go:3350-3363`),
> and `verify_prose`'s `config.error_step=handle_failure` diverts away from
> `compose_page`, the fleet's only `create_report_page` step. It is the **inverse**
> of the 079 detect-then-discard shape. Evidence and the contracts read:
> `robot_hands_gripper_dossier/NOTES_…` §"Round 2 of council 721ac4f7".

**The objection, and it is fair:** the change introduces `prose_sections` and
`no_match_sentence` as **cross-step contract fields**, and the council's own
read-only check found that `report-builder` has **`input_contract` NULL and
`output_contract` NULL**. So the new fields travel between steps undeclared —
the DECLARED CONTRACTS concern.

**To resubmit** (`RESUBMIT_CORR=721ac4f7-2076-4fea-9242-b234cfe648d6` so the trail
accumulates):
1. Declare the fields. `report-builder`'s contracts are NULL today, so this is
   additive: seed `output_contract`/`input_contract` on the agent row, or state
   in the plan why an action-to-action field inside one workflow is out of that
   mechanism's scope. **Read the mechanism before choosing** — do not guess which.
2. Fix edit 1's `symbol` field: it said `assessPayloadRated` while the
   `grounded_in` quote came from `scoreGrippers`. `editquality` flagged it LOW and
   only wanted the structural relationship confirmed.
   > **CORRECTED 2026-07-28 — do NOT "fix" it; the symbol was already right.**
   > `scoreGrippers` (`:636`) → `assessGripper` (`:652` call, `:580` decl) →
   > `assessPayloadRated` (`:588` call, `:505` decl), and the guard lives in
   > `assessPayloadRated`. The seat asked for the relationship to be *confirmed*,
   > not corrected — following this instruction would have put an error into the
   > plan. Round 2 keeps the symbol and states the two-hop chain instead.

**The council verified two of my claims for me** — worth knowing they hold:
`report-builder` is the *only* consumer of `verify_report_prose` and the *only*
producer of `score_grippers` output. The one-image compatibility argument stands.

## 5. A1 is a WON'T-DO — do not reopen without reading this

`features_open/024` carries the full correction. Short form, all verified:

- **The cited exemplar is not a config table.** `CHVerticalProfile` is a **Go map**
  compiled into the image, one populated entry, a commented-out template for the
  next. Onboarding a vertical through it costs a build and a roll.
- **"9 of 296 single-site actions" recounts to 1.** `pull_report_requests` already
  selects sites by `deploy_config ? 'report_island'`; `emit_report_status_files`
  is plumbing. Code-literal gripper mentions: `score_grippers` **41**,
  `create_report_page` 3, `report_request_pull` 1, `verify_report_prose` **0**.
- **N=2 already exists and refutes the abstraction.** idea.uk's Tier-3 scorer
  (`idea.uk/golang_files/engine.go`) is an **LLM-produced 1–5 rubric** —
  Defensibility/Willingness/Buildability/Reuse/Durability/Risk, gated on
  `Advances` and `Risk ≤ 2`. No candidate index, no units, no headroom, no verdict
  ladder. Its config intersection with gripper physics (materials→μ, the 1.25
  band, cycle-rate tiers) is **the empty set**.

**What generalises is the PIPELINE, not the scorer** — and four of the five
actions are already in that layer. Reopen only if a second site wants a *physics*
scorer, which would be N=3 with two worked examples to abstract from.

*Recorded as won't-do rather than deferred on purpose: "generalise after the
pilot" leaves a refuted idea on the schedule with a date attached.*

## 6. NEXT, in priority order

1. **Resubmit `721ac4f7`** per §4. Small, and it closes the only open loop.
2. **Adopt `platform/mailer` + `platform/httpguard` into `tools-api`.** Both are
   built, tested, **council-APPROVED (`6db59c8b`)** — and have **zero callers**,
   which is the worst state shared code can occupy: full maintenance cost, no
   benefit, while the three limiters and four CORS postures they replace keep
   drifting. `httpguard.ClientIP` carries the `bugs_open/090` forged-XFF hardening
   that `tools-api` still lacks (it keys on gin's `c.ClientIP()`).
   **`tools-api` belongs to the gauntlet thread and `bugs_open/083` is open
   against it — message them first. This is a conversation, not a build.**
3. **Finish the pilot's public half** — `/api/v1/tools/gripper` **inside**
   tools-api. Do **not** write `cmd/gripper-intake/`; that would be the estate's
   fourth VM fork. Re-seed 208's `base_url` to
   `https://tools.apis.uk/api/v1/tools/gripper`.
4. The two live fixture pages on robot-hands.com await the owner's read; cleanup
   (`source='manual-test'` rows + 2 pages) is owed once seen.

## 7. Landmines this lane paid for

- **A retag is not a rebuild.** Verify by a string your change CREATED, plus a
  positive and a negative control, against the pod that is running *now*.
- **`scheduled_tasks.target_topic`'s column DEFAULT (`system.agent.generic.requests`)
  is a topic NOTHING consumes.** 18 of 18 enabled tasks use
  `system.agent.scheduled.requests`. It fails silently and looks healthy: the
  scheduler logs "Successfully produced message" and stamps both timestamps,
  which is the *normal* fire-and-forget path. Only downstream evidence
  discriminates — zero `orchestration_states` rows for the agent type.
- **`create_report_page` requires `request_id` to be a real UUID** (it becomes the
  public URL). An invalid one also **silently disables the failure sidecar**,
  because `handle_failure` builds it from the same field.
- **`complete`/`deployed_at` is not fetchability.** A fixture page was 404 for ~2
  minutes after the item said complete. Poll the URL.
- **A duplication audit sees SHAPE, not USAGE.** Three of one sweep's "clear
  duplicates" failed verification — 8 "byte-identical" health servers were 8
  distinct bodies; `med_export_json` shared a purpose and 0 of 16 functions with
  its "generic twin"; the "generic" `firecrawl_map` had **no callers anywhere**
  while the "bespoke" `med_map_urls` was live. **Open both files and query live
  usage before calling anything a duplicate.**
- **Council submission types.** `operation` ∈ `modify|add|remove|config_change`
  (a new file is `add`, not `create`); `grounded_in` is `[]string`; `risks` is one
  `string`. `097` now type-checks all three client-side (`be0f6aa16`) — an invalid
  submission used to die server-side at `complete_invalid`, which reads exactly
  like "still queued" for the ~30 minutes you would otherwise wait.

## 8. Also shipped this session, outside this lane

- **`check_new_capability_surface`** in `scripts/pattern-check.py` — fires when a
  staged `.md` proposes a `cmd/`, dockerfile or package that does not exist, and
  prints existing peers marked `(new)`. Measured **1.33% over 1,500 commits**.
  Its value is idempotence: auditing the 07-26 commit now prints
  `tools-api (new)` — the exact peer that arrived after the original prior-art
  search. Built as a check *inside* pattern-check.py rather than a sibling
  script, because a separate script would duplicate the harness.
- **`097` type-checks** (§7) and the **`016b` §9** entry (§3).
