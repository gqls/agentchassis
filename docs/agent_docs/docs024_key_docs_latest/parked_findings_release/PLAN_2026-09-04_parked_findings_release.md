# PLAN 2026-09-04 — releasing the parked findings

**Lane opened 2026-09-04** on the owner's instruction: *"Please tell an appropriate lane to work
through all the parked findings. I can approve each as they go if they wish."* Relayed by the
`site_delivery_and_editor` lane (`docs/agent_docs/docs024_key_docs_latest/site_delivery_and_editor/HANDOFF_2026-09-04_continue_here.md` §1.3b).

## 1. The owner's ruling, and the scope conditions inside it

> *"I think the hitl loop should, for the moment, default to the system approving the changes. I
> don't need to check every fix. The damaged pages need fixing, the good ones don't. If a good one
> gets rewritten at this stage (fresh build) then as long as it is still good even if different
> then it is still ok."*

**"For the moment" and "at this stage (fresh build)" are conditions, not filler.** The accepted risk
is *a rewrite landing on a page that was already fine*, accepted **because the sites are fresh**.
That acceptance is not established for a site a customer has approved, or one already delivered.
**Re-confirm before carrying it past those boundaries** — do not inherit it.

It is also **not** a finding that the parking was a mistake. RFC_056 parked these because model
findings dispatching unattended is what `bugs_open/238` was; the ruling accepts that risk for now.
Keep the two apart in anything written.

## 2. What is parked — `[MEASURED 2026-09-04 16:1x–16:2xZ]`

`status='deferred'` AND `spec->>'filing_mode'='record'` AND empty `handler_agent`: **3,184 rows
across 39 sites**, filed 2026-08-25 → still arriving during this measurement.

| routed handler | item_type | rows |
|---|---|---|
| content-gap-planner | needs_content_planning | 1,571 |
| page-build-handler | content_rewrite | 829 |
| page-build-handler | needs_content_page | 441 |
| webdesign-agent | needs_design_review | 158 |
| copy-editor | needs_copy_edit | 53 |
| component-template-fixer | responsive_fix | 50 |
| component-template-fixer | spacing_fix | 49 |
| css-patch-agent | dark_section_audit | 32 |
| *(none — a rule-3b `capability_gap`, not releasable)* | capability_gap | 1 |

Seven seats are still filing in record mode, all `is_active`: `brief-fidelity-auditor`,
`content-quality-auditor`, `improvement-loop`, `offer-analyser`, `reader-experience-auditor`,
`site-review-agent`, `visual-design-auditor`. **The population grows while we work it.**

## 3. ⚠ THE FINDING THAT DECIDES THE DESIGN — the documented release recipe is INERT for 89% of the rows

Every row carries its own recipe in `spec.release_recipe`:

```sql
UPDATE site_work_items SET status = spec->>'routed_status', handler_agent = spec->>'routed_handler',
       updated_at = now()
 WHERE id = <id> AND status = 'deferred' AND spec->>'filing_mode' = 'record'
```

`spec.routed_status` is **`detected`** on every row. A `detected` row reaches dispatch only via
`detected-item-promoter`, which promotes to `triaged`; `find_dispatchable_site` selects
`status IN ('triaged','approved')` only. **The promoter has five doors, and door 5 holds these rows.**

Migration `629` (`bugs_open/405`, applied — 1 ledger row; predicate confirmed **verbatim in the
live `pre_query` 2026-09-04**) added:

```
(COALESCE(wi.spec->>'origin', '') <> 'model_opinion') AS origin_ok
...
WHERE pipe_ok AND handler_ok AND known_good AND floor_ok AND origin_ok
```

**`[MEASURED 2026-09-04]` 2,824 of the 3,184 parked rows carry `spec.origin = 'model_opinion'`** —
written by `write_audit_findings`, the same producer that parks them. Running the documented recipe
on the whole population would move them from one parked state to another **and look like it worked**:
status changes, handler is named, nothing errors, nothing ever dispatches.

Simulating all five doors over the parked population:

| outcome of running the documented recipe today | rows |
|---|---|
| **flows** (passes all five doors) | **352** |
| held by door 5 (origin) alone | 2,800 |
| held by door 5 **and** door 3 | 24 |
| held by door 3 alone (`dark_section_audit`→`css-patch-agent`: **0** completions, 1 failure, ever) | 8 |

**11% effective.** `629`'s own header says it does not change record mode because *"deferred rows
were never the promoter's candidates"* — true while parked, and false the instant the recipe runs.
The door's held-reason even names the exit it blocks: *"model opinion - release by hand or via
record mode"*.

> **[INFERRED, not yet observed]** No `model_opinion` row has ever sat at `detected` (**0** today,
> against 1,010 unstamped ones), so the door has never been seen to fire. The claim rests on reading
> the live `pre_query` verbatim, not on an observation. **The cheap confirmation is `629`'s own
> direction-1 recipe** — one synthetic `detected` row, proven pair, `origin='model_opinion'`, assert
> it is still `detected` after ≥2 ticks (900 s each), then cancel it by hand. It is safe precisely
> because a held row cannot dispatch. **Do this before any bulk release.**

## 4. The three ways to release, and why the recommendation is (1)

The doors are not bureaucracy: door 2 requires a live handler, door 3 requires the
`(item_type, handler)` pair to have completed at least once, door 4 requires ≥25% success once a
pair has ≥5 terminal outcomes. **Doors 1–4 pass for 7 of the 8 pairs** (3,152 rows); the 8th is
`dark_section_audit`→`css-patch-agent`, which door 3 holds **correctly**.

1. **RECOMMENDED — release to `detected`, and widen door 5 to admit a row carrying an explicit
   release stamp** (e.g. `spec.released_by`). Keeps doors 1–4, keeps provenance, keeps door 5 doing
   its job for *un*released opinions, and keeps the promoter's `LIMIT 20` / 900 s as the throttle.
   A migration, so live on apply; council-scope.
2. **Release straight to `triaged`.** Bypasses all five doors — including handler-liveness and the
   competence floor — and would dispatch the 32 `dark_section_audit` rows to a pair that has never
   once succeeded. Fastest, least safe.
3. **Release to `approved`.** Semantically the closest to the owner's words ("the system approves
   them") and it satisfies the dispatch predicate, but it bypasses doors 1–4 exactly as (2) does.

## 5. Pacing — the throttle already exists, and the FIFO ordering is the surprise

- `detected-item-promoter`: **900 s** tick, `candidates … LIMIT 20` ⇒ **80 rows/hour, ~1,920/day**
  ceiling. 3,184 rows is ~40 hours of promotion. **We do not need to invent pacing; we need to not
  defeat it.** Option (2)/(3) would defeat it, releasing straight into the dispatch pool.
- `find_dispatchable_site`: 30 s tick, **one site per tick**, `ORDER BY MIN(created_at) ASC LIMIT 1`,
  and it skips any site that already has a `claimed` row — so each site drains serially.
- ⚠ **A released row keeps its ORIGINAL `created_at`** (the recipe touches `updated_at` only).
  These are dated 2026-08-25 onward, so they sort to the **FRONT** of a fleet-wide FIFO, ahead of
  today's live work. This is the opposite of "3,184 positions at the back of the queue".
- `governor_admits()` **admits all eight item types** — checked, that door is clear.

## 6. Sequence

1. Confirm door 5 empirically (§3 box). **Not yet done.**
2. Owner picks a release mechanism from §4 and a first site.
3. **Canary one site**, not a batch. Release, watch it through promotion → dispatch → artefact,
   and check the *page*, not the work-item status (a `complete` row is not a repaired artefact).
4. Widen per site, oldest-first or owner's order, with the parked count and the outcome recorded
   per site in NOTES.
5. Re-census before each wave — the population is still growing.

## 7. Holds to respect

- ~~Chassis roll v1.0.1361 not landed.~~ **CLEARED 2026-09-04 16:2xZ** — both `agent-chassis` pods
  stamp `06c0b18f233bc600918ef481d32b40f29535f78f`, started 16:01Z; the ~300 s no-dispatch window
  is long past. (The handover said pods still stamped `239ab3626`; they rolled in between.)
- **`boxingonline.com` — talk to session `457` first** (`uds:/run/user/1000/cc-socks/862673.sock`).
  They report `/articles/index.html` has 7 occupants in a guessed slot, and that dropping occupancy
  to 1 arms an overwrite of a legitimate prose block until a binary carrying `828b22c7c` rolls.
  **Unverified by this lane** — passed on because a bulk release would hit it blind.
- **`bugs_open/396`'s parks are a DIFFERENT population** — `deferred` with a *named* handler (256
  rows), not this one. Of those, **60 carry another lane's live "un-park after rebuild verify"
  condition**. `unpark_work_items` is scoped to one `parked_by` for that reason. **Do not sweep
  them.** Any query that does not split on `handler_agent` is answering a different question.
- The `capability_gap` rule-3b row is **not releasable** and that is correct — no handler can do it.

## 8. Open questions for the owner

- Mechanism: §4 (1), (2) or (3)?
- Order and granularity: one site at a time with a check between, or open the tap fleet-wide once a
  canary is clean?
- Does "approve each as they go" mean per **site** or per **finding**? Per finding is 3,184 decisions.
