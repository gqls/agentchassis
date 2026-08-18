# 300 — `page_component_status_drift` keys its finding on `page_components.id`, which the platform's own rule says is unstable; the longer a row waits, the more likely it dispatches into a hard failure

**Filed:** 2026-08-17 · **Branch:** `087_towards_multiple_domains` · **Status:** OPEN, diagnosed with
evidence, not fixed
**Severity:** medium — low incidence today (1 of 20 rows), delayed onset, and the damage lands on a
*shared* gate rather than on the page.
**Class:** violated-invariant / stale-key. The invariant is already written down in three places;
this producer and its handler are on the wrong side of it.
**Found by:** the `bugs_open/083` close-out session, while validity-checking four findings before
hand-canarying one of them.

---

## The rule this breaks, which the estate already knows

`016b_debugging_guide` §"Two traps met while building the re-validation pass", trap 1, verbatim:

> **Do not key re-validation on a stored `component_id`.** `page_components.id` is not stable across
> re-renders. Keyed on `spec->>'component_id'`, 30 of 30 parked `needs_section_data` items and 11 of
> 45 `required_fields_missing` items resolve to a component that no longer exists — which reads as
> "the target was deleted" when the section is right there under a new row id. **Key on
> `(page_name, slot_name)`.**

Two live code sites already obey it and say why:

- `platform/orchestration/actions/revalidate_review_queue_action.go:648-649` — *"It keys the
  component lookup on (page_name, slot) and NEVER on `spec.component_id`: `page_components.id` is
  not stable across re-renders."*
- `platform/orchestration/actions/create_report_page_action.go:21` — *"NEVER by a remembered
  `page_components.id` (ids are not stable across …)"*

## What violates it

**The producer** stores the id as the finding's only handle on its subject —
`platform/orchestration/actions/discovery_checks/check_page_component_status_drift.go:173-180`:

```go
specJSON, marshalErr := json.Marshal(map[string]interface{}{
    "check":             "page_component_status_drift",
    "fix_type":          "repair_page_component_status",
    "page_component_id": r.pcID,          // <- the unstable key
    "slot_name":         r.slotName,
    ...
```

`slot_name` is right there in the same object, and `page_id` is a real column on the row — so the
stable `(page_id, slot_name)` pair the guide prescribes is already carried. Nothing reads it.

**The handler** resolves by that id alone and returns a **hard error** when it misses —
`platform/orchestration/actions/fix_component_template_action.go:801-826`:

```go
pcIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.spec.page_component_id")
...
err = params.DB.QueryRowContext(ctx, `
    SELECT ... FROM page_components pc JOIN pages p ON p.id = pc.page_id WHERE pc.id = $1
`, pcID).Scan(&observed, &slotName, &pageStatus, &hasHTML)
if err != nil {
    return nil, fmt.Errorf("failed to load page_component %s: %w", pcIDStr, err)
}
```

`sql.ErrNoRows` therefore becomes an action error, and the work item goes to `failed`.

## The measured instance

Of the four `page_component_status_drift` rows escalated to `needs_human_review` on 2026-08-17,
**one named a component that no longer existed**:

| | |
|---|---|
| item | `bc041cfb-3512-4d4d-b403-9503fd45656a` |
| page | `loanandmortgagecalculator.co.uk/guides/how-loans-affect-mortgage-affordability.html` |
| `spec.page_component_id` | `0f02ca76-c710-4fb1-bab2-58e210bb17d5` — **0 rows in `page_components`** |
| what the slot holds now | `a9550607-b97f-4556-904f-4bb133f10548`, same slot `prose-0`, `build_status='deployed'` |
| page re-rendered | 2026-08-15 21:22, i.e. **after** the finding was filed on 2026-08-10 |

So the drift was real when filed, was repaired by an ordinary re-render five days later, and the
row survived pointing at a corpse. Keyed on `(page_id, slot_name)` it would have resolved to the
live component and been closed as already-fixed. It was closed by hand instead
(`status='complete'`, `resolution_path='manual:revalidated'`, evidence in `result.revalidation`).

**Current exposure, measured honestly and not overstated:** 1 of 20 lifetime rows. The other 16 are
`deferred` on `loancalculator.co.uk` and **every one of their ids still resolves today**. They are
parked by the loancalculator rebuild thread until that rebuild finishes — which is precisely the
event that reassigns component ids, so the exposure is queued rather than absent.

## Why this is worth fixing rather than noting: the failure lands on a shared gate

This is the part `016b`'s entry does not cover, because the mechanism did not exist when it was
written.

Since `bugs_open/083`'s promoter gained a success floor (migration `444`, corrected by `454`), a
pair that drops below **25% success over ≥5 terminal outcomes is HELD** — the promoter stops
dispatching that `(item_type, handler_agent)` pair altogether. Failures caused by a dead subject id
are indistinguishable, to that gate, from a handler that cannot do its job.

So the compounding path is: a finding waits → its page is re-rendered → its id dies → it dispatches
→ it fails → the pair's ratio falls → **the promoter stops dispatching `page_component_status_drift`
entirely, including the findings that are still true.** The type disables itself by ageing, and the
symptom at the far end ("this handler is unreliable") points at the wrong component.

The pair's whole lifetime record is currently **4 complete, 0 failed** (all four completed
2026-08-17). Five id-rot failures would put it under the floor.

## Fix candidates, ordered by what closes the door

1. **Resolve by `(page_id, slot_name)`, fall back to the id — in the handler.** `fixPageComponentStatus`
   already receives `input_data.spec.slot_name` and the work item carries `page_id`. Look the
   component up by the stable pair and use `spec.page_component_id` only to disambiguate. This is
   what `revalidate_review_queue_action.go` already does and it makes the bad state
   unrepresentable rather than merely rarer. ⚠ `page_id` is **not** in the dispatch loop's
   `call_handler` input_mapping today (verified live: the mapping has `spec`, `site_id`,
   `item_type`, `work_item_id`, `component_id?`, `page_name?`, … but no `page_id`), so this needs
   either a mapping addition or `page_name` from the spec.
2. **Have the producer stop writing the id as the handle** and write `(slot_name, page_id)` as the
   subject, keeping the id only as `spec.observed_page_component_id` for forensics. Cleaner, but it
   strands every row already filed under the old shape.
3. **Revalidate before dispatch.** The `revalidate_review_queue` sweep already knows how to
   re-derive a finding from current state; this type is not in its covered set. Closes this *and*
   the general "a parked finding ages into a lie" class — but it is a bigger change and it is not
   this bug's to make.
4. **Do nothing to the key; teach the floor to discount artefact failures.** Rejected as the primary
   fix — it treats the gate rather than the defect, and a failure reason string is a weak thing to
   make a gate depend on. Noted because it would also help `bugs_open/184`'s pairs.

## How to verify a fix

1. Take a row whose `spec.page_component_id` is dead but whose `(page_id, slot_name)` resolves —
   construct one by re-rendering a page that has an open drift finding — and dispatch it. It must
   close as already-fixed (or repair the live component), **not** `failed`.
2. Positive control in the same run: a row whose drift is genuinely still true must still be
   repaired, or the fix has simply made the handler blind.
3. `page_component_status_drift → component-template-fixer` keeps a success ratio above the `444`
   floor over ≥5 terminal outcomes.

## Landmines

- **A dead id and a real defect look identical from the queue.** Both are a well-formed row with a
  plausible spec. The only tell is a join to `page_components`, and no status field carries it.
- **Do not canary this pair on a stale row.** A canary is how a never-dispatched pair earns the
  promoter's trust; failing one for an artefact reason teaches the gate the opposite of the truth.
  This is now recorded in `LANDMINES.md`.
- `016b`'s entry warns the number will be **too clean** (30/30). Here it was 1 of 4 — a dirty
  fraction, which is what a genuine mixed population looks like and is easier to dismiss.

## On process, stated rather than omitted

Per the owner ruling of 2026-07-31, a `bugs_open/` file asserting a structural cause should go
through the `090` diagnosis loop or say plainly why first-hand verification was substituted.
**Substituted here, deliberately:** the structural claim in this file is not mine and is not new —
`page_components.id` is unstable across re-renders is already recorded in `016b` and in two live code
comments. What I add is a measured instance (the dead id, queried directly), the code path that
turns it into a hard error (read at source), and the second-order consequence for the `444`/`454`
floor. Each was verified first-hand and each is a single query or file:line a reader can re-run.

## Related

- `bugs_open/083` — the promoter whose floor this defect would poison; where the instance was found.
- `016b` §re-validation trap 1 — the invariant, with its own 30/30 and 11/45 measurements.
- `bugs_open/287` — why the canary's evidence was taken at `page_components.build_status` and the
  served page rather than the item's `result`.
- `bugs_open/184` / `201` — the other pairs currently near or under the `444` floor, for a different
  reason.

---

## FIX BUILT 2026-08-18 (candidate 1, owner-chosen) — and the exposure figures in this file are now stale in the direction that matters

Built by the `bugfix-277/083` lane on the owner's decision of 2026-08-18 (*"as recommended — resolve
by `(page_id, slot_name)`, id as tiebreak"*). **Go change; INERT until the next chassis roll.**

### The exposure is larger than this file recorded, and the ageing is now OBSERVED rather than predicted

[MEASURED 2026-08-18, all lifetime `page_component_status_drift` rows, live table]

| | rows | carry `page_id` | carry `spec.slot_name` | `spec.page_component_id` resolves | **`(page_id, slot_name)` resolves** |
|---|---|---|---|---|---|
| `complete` | 66 | 66 | 66 | 59 | **66** |
| `deferred` | 16 | 16 | 16 | 11 | **16** |
| **total** | **82** | **82** | **82** | **70** | **82 — 100%** |

So **12 of 82 stored ids are already dead (15%)**, not 1 of 20; and the stable pair resolves for
every single row. The measurement could have come out otherwise in both columns — it did not.

**The ageing this file predicted has happened, in one day.** This file recorded on 2026-08-17 that
*"the other 16 are `deferred` on `loancalculator.co.uk` and **every one of their ids still resolves
today**"*. Today **11 of those 16 resolve**. Five died in a queue nobody touched, exactly as the
compounding path here describes. That is the strongest evidence in this file and it belongs to the
original diagnosis, not to me.

### One correction to a claim this bug and its handoff both carried

> ⚠ `page_id` is **not** in the dispatch loop's `call_handler` input_mapping today (verified live)

**That is no longer true** [VERIFIED 2026-08-18, `build-dispatch-loop`'s live `sub_workflow`]: the
mapping is `{spec, domain, issue?, source, site_id, page_id?, purpose?, asset_id?, item_type,
page_name?, current_page, work_item_id, component_id?, reviewed_brief?}` — `page_id?` is there. The
`?` makes it optional, so it is omitted for item types whose row has a NULL `page_id`, which is why
a sampled `component-template-fixer` payload (an `instance_scope_conversion` item) showed none.
Drift rows carry `page_id` on 82 of 82, so it would arrive. **The fix does not depend on it either
way** — it joins `page_components` through `site_work_items.page_id` using `work_item_id`, which is
unconditionally mapped, so it cannot be broken by a change to an optional mapping entry.

### What was built

`resolveStatusRepairComponent` in `fix_component_template_action.go`, called by
`fixPageComponentStatus`:

1. **`(page_id, slot_name)` via `work_item_id`** — one match wins outright, and if it differs from
   the stored id that is logged as a stale-key resolution.
2. **The stored id is the TIEBREAK when the pair is ambiguous** — and that is not decoration.
   [MEASURED 2026-08-18] **17 `(page_id, slot_name)` pairs on the estate carry more than one
   component, worst case 4.** None is a drift row today, which is precisely why it needed a test:
   resolving by the pair alone is correct on every row that exists now and silently wrong on the
   first one that is not.
3. **Genuine ambiguity is REFUSED, not guessed** — `fixed:false, action:needs_review` with the
   reason on the row, matching the two guards already in this function ("refusing rather than
   guessing"). A refusal is not a hard error, so it does not feed the `444` floor either.
4. **Fallback to the stored id** when the pair yields nothing, so every caller that worked before
   works identically.
5. Every outcome now carries `resolved_by` (`page_id+slot_name` / `…+stored_id_tiebreak` /
   `spec.page_component_id`), so a census can tell which key did the work.

The measured instance in §"The measured instance" above now lands on the `"already deployed"` arm
instead of a hard error — which is the correct reading of what happened to it.

### Residual, stated

If **both** keys miss (the page itself is gone), it is still a hard error and still reaches the
floor. Deliberate: candidate 1 is what was approved, and softening a genuinely-missing subject is a
different judgement about what `complete` means. Recorded rather than quietly widened.

**Verification is still owed at the artefact** (this file's own "How to verify a fix", unchanged):
after the roll, a row with a dead stored id must close without failing, **and** a row whose drift is
genuinely still true must still be repaired — without the second control, "no more failures" is
equally consistent with having made the handler blind.

Mutation-proven before submission: stable key never consulted → 4 tests RED; tiebreak removed so an
ambiguous pair takes the first match → 2 RED; fallback removed → 2 RED.

### Council APPROVED round 1 (corr `203d858b`), verdict READ — four medium advisories, answered by measurement not by filing

- **`editquality` / `guardian` — "82/82 `page_id` is a historical snapshot, not an invariant".** Fair, and the better answer is structural rather than statistical: **the producer always fills it.** `check_page_component_status_drift.go` builds `WorkItemSpec{PageID: pageIDPtr}` from a row whose own query is `JOIN pages p ON p.id = pc.page_id`, so the page necessarily exists and the only nil path is a uuid parse failure. **Residual stands for a hand-filed row:** a NULL `page_id` falls through to the stored id — safe, because that is today's behaviour, but the fix is then a no-op for that row.
- **`reuse_agent` / `constitution` / `prior_art_librarian` — "you wrote a THIRD implementation without searching for a shared helper".** The strongest objection, and the honest answer is that I did not state the search. Having now done it: `revalidateNamedFields` keys on `(page_name, slot)` from a parked item's spec; `create_report_page_action.go:208` keys on `(page_id, slot_name)` with page_id already in hand; this one must DERIVE page_id from a work item id, break a tie, and refuse an unbreakable one. **An extraction is probably right at three sites, but it needs the ambiguity semantics decided for all three** — wider than a bug fix. Named in register **WII-020** so the fourth caller extracts rather than writing a fifth.
- **⚠ AND THE SEARCH FOUND SOMETHING.** `create_report_page_action.go:208` is `SELECT id FROM page_components WHERE page_id = $1 AND slot_name = $2` executed as a **`QueryRow` on a pair that is not unique** — 17 such pairs exist fleet-wide. It takes an arbitrary row if one ever collides. Not touched (different lane; its pages are freshly created so a collision is unlikely today), but it is the first thing to check if that shape is extracted.
- **`bug_historian` — "the ambiguous refusal returns `needs_review` and the item still COMPLETES; nothing routes it to a human".** Correct, and it joins three arms in this same function that already do it. **The audit it asks for — is ANY `action:"needs_review"` value surfaced anywhere? — is NOT done and is not claimed.** Recorded rather than quietly inherited.
- **`debug_historian` — no deploy-verification named.** Added to WII-020's `verify-later`: probe the RUNNING pod's binary for `resolveStatusRepairComponent` with a must-be-absent control in the same breath, never a tag.

### ⚠ One more defect found while answering, NOT fixed here

The producer's `item_key` is `page_component_status_drift:<page_component_id>` — **also keyed on the unstable id.** So after a re-render the same drift on the same slot re-raises under a NEW key rather than deduping against the open row. Same root cause as this bug, one table over, and it is a producer change rather than a handler one. Left for whoever takes fix candidate 2.
