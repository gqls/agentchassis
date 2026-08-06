# 204 — `plan_sections` resolves a section by NAME/FUNCTION only, so a decomposed page can never be rebuilt — and the build path asks the fleet to manufacture junk components

**Filed 2026-08-05** from the `loancalculator_couk` lane, while carrying out the
owner's instruction to rerun a site's copy through the framework rather than by hand.

**This is `bugs_closed/182` in the sibling call site.** 182 fixed exactly this
blindness in the RE-RENDER path (`rerender_page_sections`) by resolving
`page_components.component_id` first and falling back to name. The BUILD path
(`plan_sections` → `page-build-handler` → `page-content-writer`) never got the same
fix — **and 182's own commit edited this very file.**

## The defect

`pages.sections` on a decomposed site is a list of **positional slot names**:

```
loancalculator.co.uk / guide-how-loans-are-calculated:  ["prose-0", "prose-1"]
```

`plan_sections` resolves those against component **name and function only**:

- `loadComponentSchemas` (`plan_sections_action.go:1144`) — its own comment says it
  builds `componentInfo` records *"keyed by both name and function (the lookup pattern
  planSection expects)"*.
- `:918` — `comp, ok := components[sectionName]`.

`prose-0` is neither a component name nor a function (the function is `ported-prose`,
attached via `page_components.component_id`). So the lookup misses, control falls to
the selector at `:937`, and the section is deferred.

## Measured

**0 of 57** section names on loancalculator.co.uk resolve to any component by name or
function. Fleet-wide, **86 unresolvable across 5 sites**:

```sql
WITH s AS (
  SELECT si.domain, jsonb_array_elements_text(p.sections) AS sec
  FROM pages p JOIN sites si ON si.id=p.site_id
  WHERE p.status='active' AND jsonb_array_length(COALESCE(p.sections,'[]'::jsonb))>0)
SELECT domain, count(*) AS names,
       count(*) FILTER (WHERE NOT EXISTS (
         SELECT 1 FROM content_components cc WHERE cc.function=s.sec OR cc.name=s.sec)) AS unresolvable
FROM s GROUP BY domain ORDER BY 3 DESC;
```

```
loancalculator.co.uk          57    57   (100% — fully blocked)
gaswholesalers.com           122    11
finetuning.uk                152    10
leopardessconsulting.co.uk   106     6
oufe.com                      20     2
```

## Induced end to end, not argued

One real `content_rewrite` work item (`created_by='voiceh-canary'`) at
`guide-how-loans-are-calculated` — 2 prose blocks, no calculator, so a failure could
not touch arithmetic:

```
status: needs_human_review
error:  page-build-handler no-op: no sections ready to build (empty spec sections,
        or all sections deferred for missing data) — the target section was NOT rebuilt
```

**It refuses LOUDLY, which is correct and worth preserving** (the shape
`bugs_open/194`'s framework half shipped for). The defect is that it can never
succeed, not that it lies.

## Relationship to `bugs_closed/041` — same family, different cause, and 041's fix does NOT cover this

**Check this first, because 041 looks like this bug and is closed.** Its title is
*"section lookup never normalises… and the platform asks to rebuild a component it
already has"*, which is the same sentence you would write for 204.

They are not the same:

| | `bugs_closed/041` | **204** |
|---|---|---|
| cause | lookup used the RAW string, so `call_to_action` missed the existing `call-to-action` | lookup is keyed by name/function AT ALL, so a positional slot name can never match |
| fix | normalise before lookup (`NormalizeComponentFunction`), live v1.0.1146 | resolve by `page_components.component_id` first |
| does 041's fix help? | — | **No.** `prose-0` normalised is still `prose-0`, and no component bears that name or function under any spelling |

So 041 closed the *spelling* half of this lookup's blindness and 204 is the
*identity* half. Worth stating because the family now has four members (039, 041,
095, 204) and the next one will look like all of them.

⚠ **Consequently the second-order damage below is NOT a new finding** — 041 records
the same "asks to rebuild a component it already has" behaviour, and its closure
verified *"0 new `needs_new_component` since"*. What is new is only that a
positional-slot site reproduces it at scale (114 items for one site) through a cause
041's fix cannot reach.

## ⚠ The second-order damage: it asks the fleet to build junk

The selector reads the unresolvable name as an unknown **component type** and files
work items to create it. My single canary produced:

```
needs_new_component  "Need component template for section type: prose-0"
needs_new_component  "Need component template for section type: prose-1"
needs_section_data   "Section 'prose-0' on guide-how-loans-are-calculated needs: "
needs_section_data   "Section 'prose-1' on guide-how-loans-are-calculated needs: "
```

All four **cancelled** with an explanatory note before a component-creator could act.
A full-site attempt on loancalculator would have filed **114** of these (57 × 2), for
components that already exist. **Anyone attempting a build-path run on a decomposed
site must sweep for these afterwards** — see the query in the lane handoff.

## Root cause, and why it survived 182

`a43be1e70` (182's fix) **modified `plan_sections_action.go`** — 199 lines — to factor
`componentInfoFromRaw` so the template-truncation guard *"can't drift across the three
now-shared conversion sites"*. It refactored around this lookup and left it keyed by
name/function, adding `component_id`-first resolution only to the re-render path.

This is the documented shape in 016b §9: *one call site of a shared judgement gets the
rigorous fix, the sibling stays heuristic.* Same family as `bugs_closed/041` (section
lookup never normalises) and `bugs_closed/095` (wrong slot name renders nothing and
reports complete) — this is the third or fourth appearance of "the section→component
lookup has one more spelling than anyone checked".

Re-verified at chassis **v1.0.1254**: `plan_sections_action.go` untouched since
2026-08-04. Open and unowned.

## Fix candidates, ordered by what closes the door

1. **Resolve by `page_components.component_id` first, fall back to name/function** —
   exactly what 182 did one function over, and its `loadComponentSchemasByID` /
   `loadContentComponentsByID` already exist. **Do not write a third resolver.**
   ⚠ **The real design question:** `plan_sections` has no `pageID` at that point in
   the workflow — its own comment says so (`:1140`, *"plan_sections doesn't have a
   pageID at this point"*). Work out where the page id comes from before writing code;
   that, not the lookup, is the hard part.
2. **Make the miss fail loudly at the lookup**, naming the unresolved names, instead of
   silently routing to the selector. Weaker (it does not enable the rebuild) but it
   would have turned this into a one-line diagnosis, and it stops the junk work items.
3. Re-point `pages.sections` at component functions. **Rejected:** slot names are
   positional and a page with three prose blocks would collide on one function name —
   the positional naming exists precisely to disambiguate them.

Candidate 1 makes the bad state unrepresentable; candidate 2 only makes it visible.
Both are worth having, and 2 is cheap.

## How to verify a fix

- The census above returns **0 unresolvable** for loancalculator, or the lookup
  resolves them by id.
- Re-fire the canary and assert the prose actually changes:
  ```sql
  SELECT pc.slot_name, left(regexp_replace(pc.content_data->>'content','<[^>]+>',' ','g'),200)
  FROM pages p JOIN page_components pc ON pc.page_id=p.id
  WHERE p.name='guide-how-loans-are-calculated' ORDER BY pc.position;
  ```
  Pre-fix baseline: `prose-0` opens *"How Your Monthly Repayment is Actually
  Calculated / Demystifying the 'Amortisation' formula… Most people see a monthly loan
  repayment as a flat fee. In reality…"* (1993 b), `prose-1` 192 b.
- **Zero `needs_new_component` items filed** for the run.
- ⚠ The 12 tool rows on that site are `lock_type='permanent'` — a fix must leave them
  untouched. Backups: `page_components_bak_20260805_framework_rewrite` (63 rows).

## Filing basis

**CLAUDE.md requires a cross-cutting structural claim to go through the `090`
diagnosis loop, or the filing session to state plainly why it substituted equivalent
first-hand verification. This is that statement.** Substituted, because all three
links were read directly and the failure was then INDUCED on a live page rather than
predicted: the live `pages.sections` value, the live `content_components` schema, the
source of `loadComponentSchemas`/`planSection`, the live `page-content-writer` config,
a fleet-wide census with a non-zero result on 5 sites, and a real work item whose
error message names the failure. The one thing a `090` run would add that this does
not is an independent reader — worth having if the fixing thread wants it, and the
symptom to file would be *"plan_sections resolves pages.sections against
content_components.name/function; sites whose sections are positional slot names
cannot be rebuilt"*.

---

## §Fix committed 2026-08-06 — awaiting roll + live verification (this section by the fixing session, 7fffb7ef)

**Commit `13252f714`** (`fix(204): build path resolves sections by stored identity
first`), council correlation `d3e232b8-5456-4eb8-bb6b-851f3ac28610`
(`Council-Submitted:` trailer; read the verdict before closing). Plan with the
decision log:
`docs/agent_docs/docs024_key_docs_latest/bug_backlog_clearing/PLAN_2026-08-06_204_build_path_slot_identity.md`.

What shipped — candidate 1 plus the loud-miss half of candidate 2, as one Path 0:

- `loadPageSlotComponentIDs` reads the page's `page_components` slot→component_id
  map by UNIQUE `pages(site_id, name)` — the answer to §candidate 1's "where does
  the page id come from": it was never needed; (site_id, page_name) is already a
  complete identity and the action has always had both.
- The section loop tries the stored identity FIRST, with the re-render path's
  decided 182 semantics: id wins over a disagreeing name match (observe-only
  log), a template-guard-rejected pinned component defers LOUDLY (actionable
  `needs_section_data`, NO `needs_new_component`, NO silent name fallback), an
  id with no active row falls through to the name/selector paths unchanged.
- Duplicate slot_names (legitimate, 11 pages fleet-wide) map when repeats agree
  on the id, fall back to name resolution when they disagree.
- Seven sqlmock tests (`plan_sections_slot_identity_test.go`), each
  mutation-tested; package suite 1144 pass. Committed-HEAD archive builds and
  passes.

**Census re-measured 2026-08-06 before fixing: 87 unresolvable across 6 sites**
(was 86/5 at filing — loanandmortgagecalculator.co.uk gained one). loancalculator
still 57/57.

### To close (fixed AND live bar) — steps for this or any thread

1. Fix rides the next whole-fleet release (owner runs
   `make release redeploy-agents ENVIRONMENT=production REGION=uk001`); at
   commit time live was v1.0.1256 = HEAD's tag, so the next build needs a tag
   bump.
2. Pod-grep positive AND negative controls, one pod per ReplicaSet:
   `kubectl exec -n ai-persona-system <chassis-pod> -- sh -c 'strings /app/agent-chassis | grep -c "load page slot identities"'` → ≥1
   and `... grep -c "slot_name repeats with different component_ids"` → ≥1
   (both strings are NEW in this commit; a pre-fix binary has 0 of each).
3. Re-fire the canary (`voiceh-canary` shape, §Induced above) at
   `guide-how-loans-are-calculated` and assert: prose changed (baseline in §How
   to verify), zero `needs_new_component` filed, no empty-suffix
   `needs_section_data`.
4. Sweep for junk items any pre-fix build attempt filed:
   `SELECT id, item_type, summary FROM site_work_items WHERE item_type IN ('needs_new_component','needs_section_data') AND status NOT IN ('complete','cancelled','rejected') AND (summary LIKE '%prose-%' OR summary LIKE '%tool-%');`
5. The 12 `lock_type='permanent'` tool rows on loancalculator untouched (this
   change only alters resolution; the writer's lock handling is unchanged).
6. Unblocks: loancalculator H-voice copy rerun (owner instruction 2026-08-05).

### ⚠ AMENDMENT to closure step 3 (added 2026-08-06, same session): the canary is 189-gated

Tracing the writer's save path after committing the fix: once resolution
succeeds, `bugs_open/189`'s save-path defect (slot renamed to the component
function; locked rows duplicated on name mismatch) becomes reachable from the
BUILD path — and the build path carries the positional name NOWHERE
(`RenderComponentAction` outputs only the component's own identities; see the
§blast-radius-extension appended to 189). So, until 189 is fixed:

- **Never fire a build-path run on a page holding locked rows** (loancalculator's
  12 permanent tool slots, oufe's 2) — it will duplicate them (189 §measured).
- The closure canary at `guide-how-loans-are-calculated` (2 unlocked prose
  slots) WILL rebuild the prose (204's assertion) and WILL rename the slots
  `prose-0`/`prose-1` → `ported-prose` ×2 as a side effect. Either restore the
  slot names in one UPDATE afterwards (the 189 remediation shape) and say so,
  or run the canary on a throwaway page. The rename does NOT invalidate the
  204 verification — content change + zero junk items is still the assertion —
  but an unrestored rename breaks the page's next id-first resolution (the map
  is keyed by `pages.sections` names vs `slot_name`).
- Cleanest sequencing: fix 189 first (its candidate 1 needs the producer fixed
  on both paths now), THEN run the 204 canary un-gated.

### Council outcome + closure step 2 corrected (2026-08-06, same session)

**APPROVED round 2** (corr `d3e232b8`, "2 advisory objections, none
high-severity"; round 1 was REVISE on an evidence gap plus two real corrections
of mine — trail in the PLAN doc). The commit's `Council-Submitted:` trailer is
credited automatically by `098`. Decision recorded in `doc_notes`
(`d9d67807`, subject `action/plan_sections`).

**Closure step 2 as originally written hits a documented landmine**
(debug_historian seat; `logs-deploy-reads-one-pod-of-n`): "one pod per
ReplicaSet" under-counts, because ONE image serves MANY deployments —
measured today, **44 pods run `agent-chassis:v1.0.1256`** across
`agent-chassis`, `business-intel`, `vet-intel` and the per-site deployments.
Corrected check: enumerate every deployment running the chassis image at the
new tag, then pod-grep one pod of EACH (positive: `load page slot identities`
≥1; negative: a string the commit removed — none in this additive change, so
use a pre-fix-only absence instead: expect `plan_sections: slot_name repeats`
present, and on a PRE-fix pod expect 0 — the pre-fix pod is the negative
control):
```
kubectl -n ai-persona-system get pods -o jsonpath='{range .items[*]}{.metadata.name}{" "}{.spec.containers[0].image}{"\n"}{end}' | grep 'agent-chassis:v1.0.<NEW>' | awk '{print $1}' | sed 's/-[a-z0-9]*-[a-z0-9]*$//' | sort -u
# one pod per deployment name printed above:
kubectl -n ai-persona-system exec <pod> -- sh -c 'strings /app/agent-chassis | grep -c "load page slot identities"'
```
Also from the verdict, worth carrying: the loud-defer's `needs_section_data`
item_type (rather than a new type) is a deliberate reuse — bug_historian notes
consumers keyed on item_type shape changes have been missed before; the item's
spec carries `component_id` + the "repair, do not create" reason, which is the
queryable discriminator.
