# BUG 024 — a tool-improver fix is written durably and NEVER reaches the live page

**Filed:** 2026-07-19 · travelling-docs thread · site `e33263f4-74f8-494f-b191-546845dbbddf` (gamesdesign.co.uk)
**Severity:** high — silently defeats the whole self-verifying fix loop for tool components.
**Status:** OPEN. Diagnosed with live evidence; no fix applied.

---

## Symptom

`tool-loot-table-balancer` has failed `mobile-fit@mobile` on every Tier-4 acceptance
run since 2026-07-17, across **three** tool-improver cycles. Each cycle wrote a
plausible, correctly-aimed fix and reported success. Every re-verification came
back RED with a byte-identical signal (`fieldset (419px)`).

This was diagnosed in `bugs_open/010` as *fix-loop non-convergence* — the fixer
aiming at the wrong element. That diagnosis drove a real improvement (the
drill-down attribution, T25), but it was **not the reason re-verification stayed
RED**.

## The actual finding

**The live page has never changed since the tool was born.** All three fixes are
present in the durable component template; none has ever been rendered.

Proof (2026-07-19):

| where | `.ltb-row-grid` rule | `fieldset` rule | len |
|---|---|---|---|
| `content_components.html_template` | `minmax(0,2fr) minmax(0,1fr) minmax(0,1fr) auto` | has `min-width:0; width:100%; max-width:100%` | 10,626 |
| `page_components.rendered_html` | `2fr 1fr 1fr auto` | bare, no constraints | 9,901 |
| live page (curl) | `2fr 1fr 1fr auto` | bare, no constraints | — |

`9,901` is the **v1 born length** (`component_versions.version_number = 1`).
`page_components.rendered_html` has not been regenerated since birth.

## Root cause — three defects in series

### (1) `update_component_html` relies on a flag nothing reads
`platform/orchestration/actions/update_component_html_action.go:248-255` writes
the new template to `content_components.html_template`, then sets
`page_components.build_status='pending'`, with the comment *"The rerender pipeline
regenerates rendered_html"*.

Nothing anywhere filters on `build_status='pending'`. The codebase says so itself
at `store_generated_component_action.go:455-458`: *"Without this the
build_status=pending flag is informational only — nothing downstream scans
page_components for pending rows."* **The flag is dead state.**

### (2) tool-improver's rerender request routes to assemble-from-stale
The page-rerender agent gates on the work item's `spec.reason` (LIVE definition):

```
condition: input_data.spec.reason == 'image_landed'
        OR input_data.spec.reason == 'section_data_resolved'
        OR input_data.spec.reason == 'cta_links_stale'
then_step: rerender_sections     # true template re-render
else_step: render_page           # rerender_single_page — assemble STORED html
```

`tool-improver`'s `create_rerender_item` config sets no `spec_data` at all, so
**no `reason`** → always `else_step` → `rerender_single_page`, whose own header
says *"Simple concatenation - no template re-rendering"*. It reads the stale
`rendered_html`, assembles it, deploys it, and marks the page `deployed`.

So the fix loop reports a successful deploy of an unchanged page. This is the
`complete` ≠ *the work happened* invariant, one layer deeper than usual: the work
item, the orchestration and the page status are all genuinely green.

### (3) even on the correct branch, a tool section escalates instead of rendering
Forcing the good branch (`reason=section_data_resolved`, item `478c44c9`) proved
the gate diagnosis right — `rerender_sections` ran — and then exposed the last
defect. `rerender_page_sections_action.go:186-206` refuses to render any section
whose `content_data` is empty, escalating the whole page to a full rebuild:

```go
if len(s.contentData) == 0 { reason = "no stored content_data" }
... escalateRerenderToWriter(...) ; out["escalated"] = true ; return out, nil
```

A **tool** section legitimately has `content_data = {}` — a tool is self-contained
HTML with no LLM-authored content fields (verified: `cd_type=object, cd_len=2`,
and the component has **no `input_schema`** at all).

`check_escalated` then routes `escalated==true` → `complete`, bypassing
`save_sections` — the only step that writes `rendered_html`. The render is
computed and thrown away.

That guard is correct for its original purpose (it is the fix for the blanked
article bodies, `bugs_closed/004` / `bugs_closed/005`). It simply has no concept
of a section type that has no content_data *by design*.

**Net: there is no path by which a tool template fix can reach a tool page.**

## Fix candidates

1. **(smallest, unblocks the loop)** Give tool-improver's `create_rerender_item`
   a `spec_data.reason`, and teach the guard that a section with no
   `input_schema` needs no `content_data` — render it from the template instead
   of escalating. Both are needed; either alone leaves the path dead.
2. **Escalate to the right handler.** For a tool page the escalation currently
   emits `needs_page` (full rebuild by the writer). That is the wrong remedy —
   a rebuild regenerates the tool and would destroy it (the known
   *re-plan clobbers built pages* landmine). If a tool section ever does escalate,
   it should route to the tool pipeline, not the page writer.
3. **Kill or honour `build_status='pending'`.** It reads as a working mechanism
   and is not one; it is what makes defect (1) look wired. Either make something
   scan it, or remove it and its comments.
4. **Structural:** `rerender_single_page` deploying stale HTML while reporting
   success is the load-bearing lie. It could compare the assembled output against
   the component templates it references and refuse (or flag) when a referenced
   template is newer than the stored render.

## How to verify a fix

```sql
-- component template vs stored render must agree on the fixed rule
SELECT (cc.html_template LIKE '%minmax(0, 2fr)%') AS template_has_fix,
       (pc.rendered_html LIKE '%minmax(0, 2fr)%') AS render_has_fix
FROM page_components pc JOIN content_components cc ON cc.id = pc.component_id
WHERE pc.id = '45229c85-5600-4e8c-b4ae-e8058f74b185';
```
Then `curl -s https://gamesdesign.co.uk/tools/tool-loot-table-balancer.html | grep 'ltb-row-grid' -A3`
must show the `minmax(0, …)` columns. Then a Tier-4 acceptance run should finally
go GREEN on `mobile-fit@mobile`.

## Consequences for other records

- **`bugs_open/010` needs re-reading.** Its non-convergence evidence (two
  identical RED re-verifications, two "materially identical" fixes) is explained
  by an unchanged page, not by a fixer that cannot aim. Its candidate (a)
  (drill-down) was still a genuine improvement; candidate (b) (convergence guard)
  is still worth having, but it would have fired here on a loop that was never
  actually being given a chance.
- **T24's claim that "the durable fix reached the live page (`max-width` present
  ×10)" is wrong.** `max-width` appears 10× on that page in unrelated site-chrome
  rules; the tool's own `fieldset` rule has none. Verifying a specific fix by
  grepping a *generic* CSS property is the trap — match the specific rule.
  (This thread made the identical mistake once today before catching it: see
  NOTES 2026-07-19.)

## Scope

Not tool-specific in principle: **any** component whose fix is written via
`update_component_html` and whose section carries no `content_data` is affected.
Tools are simply the population where that is the normal shape.
Adjacent to `bugs_open/021` (the durable write guard covers one path only) — same
`page_components.rendered_html` surface, different failure.
