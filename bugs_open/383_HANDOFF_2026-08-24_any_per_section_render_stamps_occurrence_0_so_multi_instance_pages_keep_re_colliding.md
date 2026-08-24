# 383 — Any per-section render stamps occurrence 0, so a repaired multi-instance page re-collides within hours

**Filed** 2026-08-24 · **Status: FIX COMMITTED (`364e80b7f`), INERT until the next chassis roll — so this stays OPEN**
· lane `docs/agent_docs/docs024_key_docs_latest/bugfix_283_component_instance_scope/`
· council correlation `3fd0d026-8966-44c6-b0d8-bd8c0dfba187` (verdict pending at filing)

> **WHY THIS IS A NEW NUMBER AND NOT `283`.** `bugs_closed/283` is a DIFFERENT defect —
> interactive components could not be reused on one page because their element ids were
> **literal**. That was fixed and went live 2026-08-22 and stays closed. This bug is the
> defect that fix *exposed*: the ids are now derived from an occurrence, and two render
> paths supply the wrong occurrence. Until now the only open record of it lived inside
> `architecture_review/RFC_032…md` §9d/§10 and a PLAN file, so a reader grepping
> `bugs_open/` for duplicate element ids found **nothing**. Filed at the owner's direction
> 2026-08-24.

## 1. What the thing is, in plain terms

A component that can appear more than once on a page (a text block, an FAQ, a mechanism
flow) must give each of its copies **different** HTML element ids, or a script or a
stylesheet aimed at "the FAQ" hits whichever copy the browser found first. The estate does
that by putting a per-copy token in every id: `id="c-faq"`, `id="c-faq-2"`, `id="c-faq-3"`.

The token is `InstanceToken(function, occurrence)`. The **occurrence** is simply *how many
sections with the same function come before this one on the page*, counted in position
order. One function computes it for everybody: `InstanceCounter.Next`
(`platform/orchestration/actions/component_instance_scope.go:140`), which lowercases and
trims the function name so `"FAQ "` and `"faq"` count as the same thing.

## 2. The defect

Two render paths only ever see **one section at a time**, so they cannot count anything —
and both simply passed **0**:

| path | file:line (before the fix) | when it runs |
|---|---|---|
| `RenderComponentAction` | `v3_site_actions.go:2404` | every page **build** and every `content_rewrite` |
| `applyContentEdit` | `section_editor_actions.go:1104` | a section **edit** on a live page |
| `applyComponentSwap` | `section_editor_actions.go:1275` | a component **swap** on a live page |

So every copy on the page got occurrence 0, i.e. the **same** token — and the ids collide
again. The trigger is not "a rebuild": it is **any per-section render**, which is why the
damage kept coming back.

## 3. Evidence (all 2026-08-24, this session, at the artefact unless stated)

```bash
curl -s https://gaswholesalers.com/pricing-transparency.html \
  | grep -o 'id="c-generic-text-block[^"]*"' | sort | uniq -c
#   2 id="c-generic-text-block"          <-- two sections, one id
```
Same shape on `vetcomparison.uk/how-it-works.html`.

- **3** multi-instance pages carry a duplicated `c-` token in stored `rendered_html` (live DB).
- **30** multi-instance `(page, function)` pairs exist; **16 of 30** repeat a `slot_name`.
- **3 of the 12** pages repaired by the 2026-08-23 rerender queue were **re-collided within
  hours** by an unrelated lane's `content_rewrite` backfill. That is the defining symptom:
  *the repair works and does not hold.*
- `apis.uk/index.html` still holds `build_status='needs_rebuild'`, so it rebuilds and
  re-collides **on its own** — a repro that regenerates itself, no setup.

⚠ **A corpus count stays GREEN while this happens.** The same token twice is still two
`c-`-prefixed tokens, so any check that counts *prefixed* ids sees nothing wrong. Count
**distinct** tokens per page, or fetch the page.

## 4. Root cause

Not "the paths disagree about the rule" — they use the same rule. They were given the
wrong **input**, and the constant was documented as safe on a measurement that was never
true of these templates: the binder's comment claimed the component "appears once per
page, which is every interactive component on every live page today (measured
2026-08-15)". That measurement was about `getElementById` tool components. It was never
true of the RFC_032 §8 repeatable templates — they are precisely the ones that repeat, up
to ×6.

## 5. The fix, committed `364e80b7f` (Go — INERT until the roll)

Both paths now feed the canonical rule its real input; nothing invents a second rule.

- **Build path**: it is always a loop iteration. Loop expansion already injects
  `loop_item_index` / `loop_name` into each substep's own config and parks every item in
  `CollectedData`, so the occurrence is counted from the sections **already rendered in
  this pass**. That is `InstanceCounter`'s arithmetic over the same list, one section at a
  time — and it is therefore right on a page's **first build**, when no `page_components`
  rows exist to count.
- **Editor path**: no loop, but it holds the stored row, so it counts same-function
  predecessors in the DB, position-exact, with a `(position, id)` tie arm — matched by
  tightening `loadStoredSections`' `ORDER BY`, because that query *is* the canonical walk.
- **Everything else**: occurrence 0, exactly as before. The derivation is an
  **input-improver, never a gate** — it cannot fail a render.
- New: `component_instance_occurrence.go`, `datahelpers/loop_keys.go`, two test files.

No migration and no workflow-config key, deliberately: the fix is live at the image roll
with no activation step.

## 6. How to verify — at the artefact, never at a status

**Before believing any of this is fixed, confirm the roll actually shipped it:**
```bash
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
git merge-base --is-ancestor 364e80b7f <the stamp>
```

Then:
```bash
# 1. the two standing repros — expect the bare token x1 and -2 x1
curl -s https://gaswholesalers.com/pricing-transparency.html \
  | grep -o 'id="c-generic-text-block[^"]*"' | sort | uniq -c
# 2. the first-build arm, free: apis.uk/index.html rebuilds itself -> 6 DISTINCT tokens
# 3. THE ACTUAL DEFECT — STICKINESS. After repair, await or trigger ONE content_rewrite on
#    a repaired page and re-count at the SERVED page. This is the step that flapped on
#    2026-08-23 17:41, and it is the only check that distinguishes "the fix works" from
#    "the rerender queue happened to fix it".
# 4. the editor arm: one content_edit on the SECOND instance of a multi-instance page,
#    then re-count -> the -2 token must survive.
```

## 7. What is NOT fixed by the commit

- **The stored damage.** A `page_rerender` goes through the canonical walk and repairs it
  correctly *even on today's code* — that is how 9 of 12 were fixed on 2026-08-23. Repair
  items are filed separately; the commit is what makes the repair **durable**.
- **A non-loop `render_component` caller** would still bind 0. There are **zero** today —
  exactly **one** active agent definition (`page-content-writer`) runs the action, as of
  **2026-08-24**, and that count goes stale *by addition*. A `LANDMINES.md` entry warns
  whoever adds the second one.
- **Editor errs-high edge case**: the canonical walk does not advance its counter for a
  *carried* section (invalid template), but the SQL counts every stored same-function row.
  Where an earlier same-function section is unresolvable the editor's count comes out one
  high — swapping which partner it collides with, not adding a collision, and no worse in
  collision count than the constant 0 it replaces. Documented on `storedPredecessorCount`.
- **Empty element ids** (`id=""`) are a *different* class with three distinct causes and
  belong to RFC_032 Half B (committed `120131549`, council `661bcf00`) — not here.

## 8. Related

- `bugs_closed/283` — the literal-id defect this one was exposed by. Resolve by SLUG.
- `architecture_review/RFC_032…md` §9c (the ruled fix shape), §9d (the correction that the
  trigger is ANY per-section render), §10 (owner ruling), §10a (the interim exposure record).
- `PLAN_2026-08-24_occurrence_derivation_and_empty_id_detector.md` — the originating plan.
  **Its Half A design is superseded**; the correction is recorded in that file.
