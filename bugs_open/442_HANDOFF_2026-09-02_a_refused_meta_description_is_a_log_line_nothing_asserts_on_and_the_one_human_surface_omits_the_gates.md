# 442 — a REFUSED meta description is a log line nothing asserts on, and the one human-facing surface omits the three gate reasons that actually fire

**Filed** 2026-09-02 by the `bugsweep_2026_08_26` lane. **Status: OPEN.**
Needs a config change (and possibly a Go change — two candidates below).

> **Resolve by SLUG** (`refused_meta_description_is_a_silent_log_line`) — bug numbers
> collide on this tree, and `git log` the FILE PATH, not the number.

**Found by the council gate reviewing someone else's fix — mine.** The `bug_historian`
seat raised it as a [medium] advisory objection against the `bugs_open/338` submission
(`106802fc-ad14-4beb-b622-147c3a0ab982`, verdict APPROVED). The round approved and still
found a real hole, which is the case for reading the objections rather than stopping at
the verdict. Its exact ask is §3 below; investigating it made the finding **sharper than
the objection knew** (§4).

---

## 1. What a reader needs first

`save_page_meta_description` (register **SEO-004**, `bugs_open/320`) writes a search-result
sentence onto a page. Before writing, it runs the site's **copy gates** — the voice gate
and the banned-claims sweep. When a gate fires, the action **refuses**: it writes nothing.

Refusing is correct behaviour. **The defect is what a refusal LOOKS LIKE from outside.**

## 2. The mechanism, read first-hand

`save_page_meta_description_action.go:192-199` returns `(map, nil)` — a **nil error**:

```go
if reason, detail := metaDescriptionFailsCopyGates(ctx, params, candidate, logger); reason != "" {
    logger.Warn("candidate refused by the copy gates — nothing written",
        zap.String("reason", reason), zap.String("detail", detail))
    return map[string]interface{}{"updated": false, "reason": reason, "detail": detail}, nil
}
```

So the step **succeeds**. Downstream, the live `meta-description-backfiller` config
(read 2026-09-02 from `agent_definitions`, not from a repo seed):

- `backfill_loop.sub_workflow` is `save_description → done` (`loop_complete`). **There is no
  conditional on `save_result.updated` anywhere in the workflow.**
- `continue_on_error: true` — moot, since the action does not error.
- Nothing reads `save_result` again.

**Net: a refusal is one `logger.Warn` plus a field in `collected_data` that nothing asserts
on.** No work item, no retry, no error, no failed status. The orchestration COMPLETEs and
the scheduled task stamps a clean run. `[MEASURED 2026-09-02]` `orchestration_states`
returns **zero rows** carrying a `save_result.reason` fleet-wide — the rows age out, so
even the field that does exist is not readable after the fact.

This is the estate's most-repeated failure shape (016b §9, *"a silent fallback deploys a
hollow section as success"*), on a page that then stays blank on an hourly schedule.

## 3. The objection, verbatim, so the ask is not paraphrased

> "The plan should either confirm an existing failure-surfacing mechanism (work item,
> error log, retry) already exists downstream of `metaDescriptionFailsCopyGates`'s
> non-empty return, or add one — otherwise this is a symptom fix on a known-recurring
> silent-loss mechanism, not a fix of the mechanism."

**Answer: it does not exist.** §2 is that confirmation, and it is the reason this file
exists rather than a line in `bugs_open/338`.

## 4. ⚠ AND THE ONE HUMAN-FACING SURFACE OMITS THE GATES — a stale list, by addition

The workflow's `complete` step carries a `result_message`, which is the only place anything
tells a person how to read the outcome. Quoted verbatim from the live row:

> "Meta description backfill finished. Read each save_result: `updated` true is a write,
> false carries a named reason (**empty_candidate / candidate_looks_internal /
> candidate_too_long / already_has_description**)."

**Those four are the cheap refusals. The list omits all three copy-gate reasons —
`voice_tell`, `banned_claim`, `voice_gate_unreadable`** — which are the expensive ones, the
ones that need a human, and the ones that caused `bugs_open/338`. A reader following this
message's own instruction would conclude the gates cannot refuse.

`[MEASURED 2026-09-02]` the action returns **7** distinct reasons; the message names **4**.
⚠ **It takes TWO greps, and that asymmetry is the whole point** — the four the message
names are string literals in the action, and the three it omits are returned as bare
strings from `metaDescriptionFailsCopyGates`, so the obvious single grep finds exactly the
four already documented and reports the list as complete:

```bash
# the 4 the message names — literals in the result map
grep -oE '"reason": *"[a-z_]+"' save_page_meta_description_action.go | sort -u   # -> 4
# the 3 it omits — returned from the gate helper, invisible to the grep above
grep -nE 'return "[a-z_]+",' save_page_meta_description_action.go               # -> :316 voice_gate_unreadable, :334 voice_tell, :341 banned_claim
```

(I wrote the single-grep version into this file first and it did not reproduce my own
number — logged in `WRONG_CALLS.md` as the same fault one step along: **a citation of a
command is not its output**.)
The gates were added to the action by the owner requirement of 2026-08-19 (`bugs_open/320`)
and **the message was not updated**. Correct when written, wrong by addition, reading as
current ever since.

> **This is the THIRD stale-by-addition list found in one task**, after `bugs_open/338`
> §4's check list (missing `negation_density`, from `bugs_open/305`) and the LANDMINES
> entry's remedy. The owner ruling of 2026-08-22 requires a COUNT to carry the date it was
> counted; **these are not counts, they are ENUMERATIONS, and they rot the same way with no
> equivalent rule.** Worth considering as its own norm.

## 5. Fix candidates, ordered by what closes the door

1. **Make a gate refusal a filed work item.** The strongest: it moves the failure onto a
   queue with a surface, and the copy is exactly what a human must judge. ⚠ Check the
   queue's volume first — `bugs_open/033`/`083` record `voice_tells` items parked in a
   review queue that has closed ~one item ever, so this must not simply relocate the
   silence. Cheapest honest version: file only on `voice_tell`/`banned_claim`, not on the
   four cheap reasons.
2. **Branch in the workflow on `save_result.updated`.** Config-only, live immediately, no
   roll. Weaker — it needs somewhere to send the branch, so it is really candidate 1's
   plumbing.
3. **Fix the `result_message` list.** ~10 minutes, config-only, closes §4. It does not make
   a refusal loud, so it is a documentation fix, not a mechanism fix — but it removes an
   actively misleading surface and should ship regardless of 1 and 2.
4. ~~Make the action return an error~~ — **rejected, name it so nobody re-proposes it.**
   A refusal is correct behaviour; erroring would fail the whole batch of up to 25 pages
   for one bad sentence, and `continue_on_error: true` would swallow it anyway.

## 6. How to verify a fix

**Induce both arms, or the test proves nothing:**
1. A candidate containing a banned phrase must produce a **visible** artefact — a work item
   row, or a distinguishable status — not merely a `logger.Warn`.
2. A clean candidate must still write silently, with **no** spurious item filed.

⚠ **Do not verify at `orchestration_states`** — it ages out (§2) and returns zero rows for
this today. Verify at whatever durable surface the fix adds.

## 7. Blast radius

`save_page_meta_description` is today the only caller of the copy gates on a single-value
field, so the silent path is one action wide. **It widens the moment the gates are reused
on a title, nav label or alt text** — which `bugs_open/338` §5 and CQ-035 both expect —
because the refusal shape travels with the gate, not with the field.

## 8. Provenance

Not run through `090`. **Substituting first-hand verification per the owner ruling of
2026-07-31, and stating the substitution rather than omitting it:** the deciding code path
was read (`save_page_meta_description_action.go:192-199`, the nil return); the live
workflow config was read from `agent_definitions` rather than a repo seed (the loop step,
its `continue_on_error`, the absence of any conditional on `save_result`, and the
`result_message` quoted verbatim); the seven-vs-four reason count is a grep over the
action. The claim is narrow — one action's refusal has no downstream reader — and asserts
no cause outside the symptom. It was independently raised by the council's `bug_historian`
seat from the submission alone, which is corroboration from a reader with no access to
this tree.

**Related:** `bugs_open/338` (the false-positive trigger, fix committed `425398a01`, inert
until the next roll) · `bugs_open/320` (SEO-004, the owner requirement that added the
gates) · CQ-035 (the single-value classification) · 016b §9 silent-fallback family.
