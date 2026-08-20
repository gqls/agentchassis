# DESIGN 2026-08-20 — the narrow sibling: one named component, one named defect, a `field_updates` payload

**Status: NOT BUILT, and not in flight in this lane.** This is the spec three lanes have now
independently asked for, written down once so the fourth asker does not re-derive it. Whoever builds
it owns it; this lane owns `copy-editor` (stage 2) and is not competing.

**Who asked, and the demand measured `[MEASURED 2026-08-20, live + archive]`.** ⚠ Counted across
**both** `site_work_items` and `site_work_items_archive` — the live table alone reports
`cta_improvement` as **29** when the true lifetime is **999**, a 97% undercount, because the archive is
bigger than the live table (existing LANDMINE; grep it for a table footprint, the SessionStart hook
cannot match one).

| asker | type | lifetime | not terminal |
|---|---|---|---|
| `bugfix_323_cta_improvement_refusal` | `cta_improvement` | **999** | 6 |
| `bugfix_277_required_fields_repair` | `required_fields_missing` | 160 | 30 |
| `bugs_open/301` / `083` + `bugfix_184` | `literal_markdown` | 98 | 69 |

## 1. It is a DIFFERENT agent from stage 2, and the reason is structural rather than stylistic

`copy-editor` reads a **whole page** because the faults it exists to find are only visible across
sections — the same argument made five times, one thing under four names, the most useful thing
buried third. That page-scoped read is its entire justification and it is **pure cost** for a named
defect on a named component: you already know which component and what is wrong with it.

So the sibling is not a mode of stage 2 and should not be built as one. It is: *given one component,
one stated defect and an acceptance test, emit `field_updates` for the named fields and nothing else.*

## 2. What it should inherit from stage 2, because each item was paid for

1. **Enumerate the invariant as DATA, never as prose.** Stage 2's own comment: *"a prose instruction
   to preserve a set is not reliably followed"* — the page had lost six required links under exactly
   such an instruction. For the sibling, the equivalent is the auditor's `acceptance_test`, which is
   already written on every `cta_improvement` row and is free.
2. **Declared-schema-in, same-type-out.** `applyContentEdit` overwrites every field the agent NAMES
   (`section_editor_actions.go:746-748`), so a type change is a live defect, not a lint. Stage 2's
   type gate has its own induced control for this reason (`bugs_open/260`).
3. **A bounded edit set.** Stage 2 truncated at `max_tokens` on a diffuse page and was fixed with a
   **budget** (3 edits) rather than a bigger cap: *an edit set bounded at the source cannot truncate*.
   The sibling's natural bound is one component, so this is nearly free — but state it.
4. **Compare sets, types and structure — never prose to prose.** `bugs_open/278 §8`: same generator,
   same inputs, 2 of 4 card bodies diverged with nothing wrong. A prose-diff gate fails that pair.
5. **The lock is in the SELECT** (`locked_at IS NULL`), not in the instructions. In the 08-09 arm test
   both prompt versions tried to overwrite the owner's approved opening and were stopped only by the
   lock.
6. **A pre-write type check exists as a pure function** — `datahelpers.ContentTypeViolations`
   (`content_type_violations.go`, from the `bugs_open/260` lane): no DB, no render, no logger, indexed
   nested `Path` like `steps[2].branches`, both live `items` dialects understood, and absent/nil/empty
   are never violations. **A Go executor should call it before it writes** and get the same verdict the
   renderer will reach.

## 3. What is different, and the traps that are specific to it

- ⚠ **Do NOT edit `_url` fields.** On the `ctaFieldNames` components the url fields' schema source is
  `renderer`, so the resolver re-resolves them into `resolved_data` on **every** render and **merges
  last** — a `field_updates` write to `cta_url` is overwritten at the next render
  (`resolve_internal_links_action.go`; `bugs_open/238`). **Labels are safe; destinations are not**, and
  the destination class already has a deterministic owner (the internal-link resolver / `cta_links_stale`
  recompute, which fixed robot-hands.com/index ~2h after the auditor flagged it, by that route and not
  by the item). Edit TEXT fields and leave `_url` alone, or the two mechanisms will fight.
- **`literal_markdown` may need no LLM at all.** Migration **473** makes a page-rerender the mechanical
  repair. Check that before building an LLM path for the most deterministic of the three types.
- **The RFC_015 citation gate.** `apply_section_edit` requires `acknowledges_decision` /
  `supersedes_decision`; a hand-filed `section_edit` can silently omit it and the gate returns a SKIP.
  A machine-filed one should populate it deliberately rather than inherit the SKIP.
- **Its own producer cannot birth an item `detected`.** `checkpoint_for_review_action.go:202`
  hardcodes both `handler_agent='human-review'` and `status='needs_human_review'` in one INSERT, so
  the promoter route the `083` lane suggested is not available without a Go change on a shared action.

## 4. The safety posture is the real decision, and it is not a detail

Stage 2 is **proposal-only, human-approved**, and every run so far has been a hand-fired canary —
nothing dispatches `copy-editor` by choice. A repair queue with ~99 open items across three types
will want less than one human approval per item, and **that is a change of posture, not a
configuration** (owner decision D2, 2026-08-12). The containable shape, if someone wants volume:
approval gates the **dispatch** rather than the typing — a human releases a batch, the agent proposes,
a mechanical gate grades, and only then does it apply. That keeps D2 intact and removes the typing
cost, which is the cost the asking lanes actually named.

## 5. Is `gate_stage2_edit.py` reusable? Partly, and here is the seam

Reusable: the type gate (both schema dialects), the link-set comparison, the markup-parity count, the
no-invented-figure check. **Not** reusable: the volume floor's discriminator, which asks *"is every
removed figure and link still reachable elsewhere ON THE PAGE?"* — that question needs the page-scoped
read the sibling deliberately does not do. For one component, the equivalent is the auditor's
`acceptance_test` plus the type check in §2.6.

⚠ And the lesson that made the gate trustworthy: **after changing a check, re-run the controls.**
Three of its holes were found by USING it on something harder, all the same family — it reported
"checked" for something it had not looked at.

---

**Pointers:** `bugfix_277_required_fields_repair/CONTRIB_2026-08-19_reply_different_agent_and_check_473_before_you_build_anything.md`
· `copy_quality_two_stage/CONTRIB_2026-08-19b_from_the_323_lane_…` ·
`copy_quality_two_stage/CONTRIB_2026-08-20_260_renderer_half_shipped_…` · register **CQ-024** (stage 2)
