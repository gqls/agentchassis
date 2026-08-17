# 291 — `tool-auditor` files dispatchable work at `hitl-review`, an agent that has never existed, and it is bleeding now

**Filed 2026-08-17** by the `bugs_open/284` lane, which found it while verifying its
own fix at the live table. **Status: OPEN, evidence measured, producer identified,
root cause NOT diagnosed beyond the producer** — per the 2026-07-31 owner ruling
this file asserts no structural root cause; it records what was measured first-hand
and names the `090` run as the next step. Grepped `/bugs_open/` and `/bugs_closed/`
for `hitl-review` first: **zero hits**, so this is not a duplicate. The
`needs_diagnosis` queue was empty at filing.

This is `bugs_closed/077`'s class alive again — *"a check that filed items at an
agent which had never existed, so every one was blocked at claim time"* — reached
this time from a **config** producer rather than Go.

## `[MEASURED]` 2026-08-17 — 14 rows, 2 sites, three days, still arriving

```sql
SELECT count(*), count(DISTINCT site_id), min(created_at)::date, max(updated_at)
FROM site_work_items WHERE status='blocked' AND error LIKE 'Handler agent not registered%';
-- 14 | 2 | 2026-08-14 | 2026-08-17 06:55:35
SELECT EXISTS(SELECT 1 FROM agent_definitions WHERE type='hitl-review' AND deleted_at IS NULL);
-- f   ← the handler every one of them names does not exist, and never has
```

**It is growing, not settled**: the same census on 2026-08-16 returned **5**; a day
later it is **14**. Newest arrival 2026-08-17 06:55.

Every row: `item_type='needs_human_review'`, `handler_agent='hitl-review'`,
`created_by='tool-auditor'`, `source='tool-auditor'`, `spec->>'check'='tool_auditor'`.
Their summaries are real, specific and useful findings about live tools — *"None of
the four number inputs (#w1, #h1, #w2, #h2) have associated labels"*, *"The entire
tool depends on `policy-generator.js` via a bare …"*. This is good audit output
being recorded as a routing failure.

## `[MEASURED]` They are BORN dispatchable — no promoter is involved

```sql
SELECT count(*) FILTER (WHERE spec ? 'original_pipeline') AS via_a_promoter,
       count(*) FILTER (WHERE triaged_at IS NOT NULL)     AS triaged_at_set
FROM site_work_items WHERE status='blocked' AND error LIKE 'Handler agent not registered%';
-- 0 | 0
```

Both zero, so nothing promoted them: they were filed straight into a dispatchable
status, claimed by the dispatch loop, and refused by
`claim_work_item_action.go`'s **handler-not-registered** branch (:191-214) — the
sibling of the empty-handler branch that `bugs_open/284` is about.

**So 284's fix does NOT cover this, and that is the point of filing separately.**
284 put the claim path's routability test at the *promoter*; these rows never go
near the promoter. It is the "born dispatchable" second path 284 records as
unclosed — with a machine producer instead of a hand insert, and a **named but
unregistered** handler instead of an empty one. Note also that the CHECK constraint
284 proposes as the eventual closer would **not** catch this either: a CHECK cannot
subquery `agent_definitions`, so "this handler exists" is not expressible there.

## Producers naming `hitl-review` — two, and one is Go

- **`tool-auditor`** (live `agent_definitions` config) — the producer of all 14 rows.
  `SELECT type FROM agent_definitions WHERE deleted_at IS NULL AND default_config::text
  LIKE '%hitl-review%'` → one row: `tool-auditor`.
- **`resolve_composition_layout_action.go:390`** (`handlerAgent: "hitl-review"`) —
  has produced none of these 14, but is the same mistake waiting to fire. Its
  comment at :375 says the pairing *"matches the existing"* convention, which is how
  a wrong value spreads: it was copied because it looked established.

## Why it matters

1. **The findings are lost.** `blocked` is not terminal, so each row holds its
   `(site_id, item_key)` slot in `idx_swi_dedup` — the auditor's later findings for
   the same key are dropped silently. The count understates the loss.
2. **Nothing will ever release them.** `feasibility-recheck` promotes a blocked row
   only `WHERE EXISTS (… ad.type = wi.handler_agent …)`; `hitl-review` will never
   satisfy that until an agent of that name exists.
3. **It is the live half of this family.** 284's class is currently paused because
   its driver (`improvement-sweep`) is disabled; this one is arriving daily.

## Fix candidates, ordered by what closes the door (pending diagnosis)

1. **Decide what `hitl-review` was meant to be** — a real agent, or the empty-handler
   HITL-terminal idiom the sibling checks use (`check_unverified_claims`,
   `check_voice_tells`: `HandlerAgent: ""` + `status: needs_human_review`, which is
   never dispatched and never blocked). If the latter, both producers are simply
   wrong and the fix is to stop naming a handler at all. **Measure before choosing:**
   `needs_human_review`'s other rows (5 cancelled, 2 at `needs_human_review`) carry
   an empty handler, so the idiom already exists in the data.
2. **Refuse an unregistered handler at the WRITE door**, not at claim — the check
   claim already performs, moved to `writeWorkItem`. Covers every future producer
   that uses the shared door; ~20 raw `INSERT INTO site_work_items` sites bypass it,
   so it is partial. (284's `workItemHandlerRegisteredSQL` is the predicate to reuse
   — do not write a fourth copy.)
3. **A minting ratchet for `handler_agent`**, the shape `bugs_open/279` built for
   `item_type`: a closed set plus a CI test, so a config that names a nonexistent
   agent fails before it ships.
4. **Repair the 14 rows** once (1) is decided — they are useful audit findings, so
   the repair is a re-file at the right shape, not a cancel.

**Next step: a `090` diagnosis run** — point it at `tool-auditor`'s live
`default_config`, `claim_work_item_action.go`'s handler-not-registered branch,
`resolve_composition_layout_action.go:375-390`, and the `needs_human_review`
population split by `handler_agent`. Check the queue first.

---

## 2026-08-17 (session bugfix-291) — taken on; root cause pinned; fix live in part

Owning workstream: `docs024_key_docs_latest/bugfix_291_hitl_review_phantom_handler/`.
`090` run filed: `3555b514-ca8f-4f31-9f55-e105ce73e961`.

> **CORRECTED 2026-08-17 (two claims above, both `[MEASURED]` against the live table):**
> 1. Fix candidate 1 says `needs_human_review`'s other rows "carry an empty handler,
>    so the idiom already exists in the data". **False as written**: all 7 non-291 rows
>    carry `handler_agent='human-review'` (checkpoint producers: `tool-recreation-handler`,
>    `image-url-404-handler`, `generic`) — measured
>    `SELECT handler_agent, status, count(*) ... WHERE item_type='needs_human_review' GROUP BY 1,2`.
>    The EMPTY-handler idiom is real but lives in the Go discovery checks
>    (`check_unverified_claims.go`, `check_voice_tells.go`) and fleet-wide
>    (544 `''` vs 22 `'human-review'`, `refresh_evidence_fact_drift.go:698-703`) —
>    not in this item_type's rows. What caught it: re-running the file's own query.
> 2. The producers section calls `resolve_composition_layout_action.go:390` "the same
>    mistake waiting to fire". **Half-right**: it names the same phantom handler, but
>    :391 sets `status: "needs_human_review"` explicitly, so its items are never claimed
>    and can never take the blocked flip. The wrong HANDLER spread by copying; the safe
>    STATUS did not. The bleeding difference is that tool-auditor's config has NO status
>    key, so `create_work_item_action.go:208-211` births its rows at the dispatchable
>    default `'triaged'`.

**Root cause**: the missing `status` key + the never-built handler. `hitl-review` was
documented 2026-04-19 as *"a convention, not a registered agent"* with "build the
hitl-review handler agent proper" as a roadmap row that was never done
(`old_design_and_styling/HANDOFF_2026-04-19_…update4(3).md:136`); 016 has carried
"Known missing handlers: `internal-linker`, `hitl-review`" for months. The intended
consumer of these items already exists and is NOT a dispatch agent: the admin confirm
endpoint (`confirm_work_item_handler.go:77,95-117`) turns a confirmed
`spec.check='tool_auditor'` review item into an `improve_tool` follow-up.

**Fix, decided and (in part) live** — candidates 1+2 of this file, candidate 3
discharged rather than built (the write-door probe against `agent_definitions` IS the
write-time registry check `bugs_closed/279`'s ratchet header said did not exist):

- **LIVE 2026-08-17**: migration `450` adds `status: "needs_human_review"` to
  `create_review_item.config` (status ONLY — the live binary hard-errors on an empty
  config handler at `create_work_item_action.go:184-187`, so flipping the handler now
  would silently lose every finding under `continue_on_error`); migration `451`
  re-parked all 14 rows at `needs_human_review`/`''`, error cleared, stamped
  `result.repair_291`. Verified: 0 rows at the blocked predicate; snapshot proven
  pre-update in `agent_definitions_backup`.
- **COMMITTED `c8400e452`, inert until the next chassis roll**: the class fix —
  `writeWorkItem` demotes a dispatchable-born item at an unregistered NAMED handler to
  born-`blocked` (claim's own predicate via `workItemHandlerRegisteredSQL`; demote not
  refuse; feasibility-recheck self-heals on registration); `create_work_item` accepts
  the parked idiom from config; `resolve_composition_layout` drops the phantom for
  `''`. Register **WDS-018**. Council `4d1ed8a5-20c4-420f-b619-6197ab9af1b2`:
  round 1 REVISE (guardian, gating — answered with a kill-switch
  `DISABLE_UNREGISTERED_HANDLER_DEMOTION`, ships armed, + measured probe cost
  0.107ms × ~613 inserts/day, commit `f629f4530`); **round 2 APPROVED**
  (2 advisories, none high; the dedup KEY-GRANULARITY loss is recorded as a
  residual owned by the 285 lane's per-finding review keys — no routing fix can
  close it).
- **STAGED** (`bugfix_291_.../STAGED_tool_auditor_review_handler_to_empty.sql`):
  the flip of tool-auditor's inert parked handler to `''` — MUST NOT apply before the
  roll (ordering gate in the file header).

**`090` verdict (run `3555b514-…`, completed 12:22): outcome CONFIRMED — with a
timeline caveat a later reader MUST have.** The run independently confirmed the core:
`hitl-review` has never existed (0 rows, direct query), the config route files at it,
and it cited the exact config path. **But its `symptom_check` marks three legs
"contradicted"/"unexplained" — the no-status-key leg, the blocked-rows leg, and the
claim leg — and every one of those readings was caused by THIS lane's own fix landing
between filing (12:03) and the diagnoser's queries (~12:10+):** migration `450` added
the status key at 12:07, `451` zeroed the blocked population at 12:08. The legs were
true at filing time — first-hand, timestamped measurements are in this file and the
workstream NOTES (14 rows at the exact error, measured 12:27Z and again 12:04Z;
the config walk showing NO status key, measured pre-450). Do not read the run's
"contradicted" lines as a refutation of the mechanism; they are unintended proof the
fix works. Sequencing misstep logged in `WRONG_CALLS.md` ("mutated the state a filed
instrument was about to measure"). Bonus finding from the run: one parked
`hitl-review` row from 2026-08-12 (`needs_new_layout_candidate`,
`created_by='site-design-planner'`) — **proof `resolve_composition_layout_action.go`
is live and reachable**, safely parked; the staged Phase 3 migration sweeps it.

**Stays OPEN until fixed-AND-live**: the roll carrying `c8400e452` is
provenance-verified, the staged flip applied, and one auditor run files
`''`+`needs_human_review` review items end-to-end.

### `[MEASURED]` 2026-08-17 17:2x — a chassis build WAS deployed and it shipped none of this code

Checked before touching Phase 3; the staged file's ordering gate held and **Phase 3 was
NOT applied** (against this binary it would hard-error every review-item filing inside
`continue_on_error` — every finding silently lost).

- Pods restarted 14:43Z on image **`v1.0.1305` — the same tag as before these commits**;
  `makefile` `IMAGE_TAG` never bumped; `imagePullPolicy: IfNotPresent`.
- Binary probe with both controls discriminating: the kill-switch literal
  `DISABLE_UNREGISTERED_HANDLER_DEMOTION` is **ABSENT** from the running binary
  (positive control `Handler agent not registered: ` present; negative control absent).
- **The build was fine — the delivery was not.** The LOCAL `v1.0.1305` image, built
  14:30Z, **does** contain the literal. Digests: local `sha256:6039e19c…` vs running
  `sha256:f90a7e88…`. A same-tag rebuild serves the node's cached image.
- Fleet-wide, not this lane: another lane measured the running chassis still at commit
  `6a782274b` with **203 commits in HEAD but not in it** — every Go change committed
  2026-08-17 is inert.
- **Remedy (owner-run, whole-fleet): `make release IMAGE_TAG=v1.0.1306`** — a tag BUMP;
  re-applying at the same tag re-serves the same cache. New fleet-wide LANDMINE records
  the 10-second digest check.

**Config half remains LIVE and unaffected** (migrations 450/451 — DB config is live on
apply): the bleed is still stopped and the 14 findings are still parked and actionable.
Only the Go class-fix and Phase 3 wait on a real roll.
