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

## §LIVE 2026-08-06 at v1.0.1257 — verified in the binary and against the data; ONE gated step remains

**The fix shipped.** Pod-grepped one pod per deployment running the chassis
image (`agent-chassis`, `business-intel`, `vet-intel` — one image, many
deployments, per the corrected step above):

```
load page slot identities                       -> 1   (added by 13252f714)
slot_name repeats with different component_ids  -> 1   (added by 13252f714)
stored component %s for slot                    -> 2   (added by 13252f714)
NEGATIVE CONTROL (a string in no build)         -> 0   (the grep discriminates)
```

**The fix works on the live data**, proven read-only (no build fired, so the
189 gate is respected). For loancalculator.co.uk, per section name in
`pages.sections`:

| | count |
|---|---|
| section names | 57 |
| unresolvable by name/function (the defect) | **57** |
| resolvable by stored `page_components.component_id` (the fix's route) | **57** |

Every section the old lookup could never resolve has a stored component_id
pointing at an ACTIVE component — which is exactly what Path 0 reads. The
defect's population is fully covered by the fix's route.

**Remaining to close: the behavioural canary only**, and it is still gated on
`bugs_open/189` (see the amendment above). 189's fix is committed
(`92e14493b`, council corr `87444080`) but NOT yet live — it needs the next
roll past v1.0.1257. Sequence for the next session:

1. Wait for / request a roll carrying `92e14493b`; pod-grep `stored_slot_name`
   (expect ≥1; it is 0 at v1.0.1257, measured — that is the negative control
   for 189's own verification).
2. Apply the writer config: `slot_name_from` on `render_section` and
   `render_from_template` (the seed's appended `jsonb_set` block in
   `docs/agent_docs/sql_for_agents/023_page_content_writer_agent.sql`).
   Without it the BUILD half of 189 is inert.
3. Verify 189 first (its own §how to verify: fire `section_data_resolved` on
   tool-loan-vs-savings, assert exactly 4 rows and `tool-2` still locked at
   position 3 with id/locked_at unchanged).
4. THEN fire this bug's canary un-gated at `guide-how-loans-are-calculated`,
   assert the prose changed against the §How-to-verify baseline and that zero
   `needs_new_component` items were filed. Then close and `git mv` to
   `bugs_closed/` (name BOTH paths on the commit).

## §CANARY RUNNING 2026-08-06 — plan_sections resolves both positional slots (the defect's exact inverse)

`slot_name_from` applied (189's config half; both keys confirmed independently),
so the BUILD path is fully live. Canary fired as 204 §How-to-verify prescribes:
work item `996b9619-46aa-4b5e-ab71-80e141e0d87e` reusing the original
`voiceh-canary` spec **verbatim** (copied by SQL from item `2517bc4b`, so the
prompt cannot have drifted), dispatched to `page-build-handler` as orchestration
`fa89217a-768b-4f22-bd7b-12209f58cbf3`.

**`plan_sections`' own output, read live from `collected_data`:**

```
ready_count=2  deferred=0  skipped=0  ready_names=["prose-0", "prose-1"]
resolved:  prose-0 -> ported-prose,  prose-1 -> ported-prose
```

This is the defect's exact inverse and it is the strongest single piece of
evidence in this file:

- **Both positional names resolved.** Pre-fix: 0 ready, 2 deferred, and the run
  died `page-build-handler no-op: no sections ready to build`.
- **Each resolved to the function `ported-prose`** — a value obtainable ONLY
  through `page_components.component_id`, since `prose-0` is neither a component
  name nor a function under any spelling. That is Path 0 executing.
- **The item keeps its positional `name` while carrying the resolved
  `function`**, which is the design: identity from the row, schema from the
  component.
- **Zero junk work items** filed in the window (pre-fix: 2
  `needs_new_component` + 2 `needs_section_data` for components that already
  existed; a full-site run would have filed 114).

The run then proceeded to `call_content_writer` — i.e. past the gate that used
to be a dead end. Save/deploy outcome and the prose comparison follow below.

## ✅ FIXED, LIVE AND BEHAVIOURALLY VERIFIED END TO END — 2026-08-06, v1.0.1259

**Kept in `bugs_open/` deliberately.** Owner direction 2026-08-06: *"please leave
the bugs that you've found in bugs_open not in the closed bug file."* That
overrides CLAUDE.md's `/bugs_closed/` bar; the fix being live is a fact about the
code, not permission to retire the ticket. Do not `git mv` this file.

Canary: orchestration `fa89217a-768b-4f22-bd7b-12209f58cbf3`, work item
`996b9619-46aa-4b5e-ab71-80e141e0d87e` (the original `voiceh-canary` spec copied
verbatim by SQL from `2517bc4b`, so the prompt cannot have drifted). Terminal
state `complete/COMPLETED`, page `deployed_at 11:53:21`.

### Every assertion this file asked for

| assertion (from §How to verify) | result |
|---|---|
| sections resolve instead of deferring | `ready_count=2, deferred=0, skipped=0`; `prose-0 -> ported-prose`, `prose-1 -> ported-prose` — a function reachable ONLY via `component_id` |
| the prose actually changes | prose-0 **1993 → 2358 b**, prose-1 **192 → 471 b** |
| zero `needs_new_component` filed | **0** (pre-fix: 2 + 2 junk items per page; 114 for a full site) |
| the 12 `lock_type='permanent'` tool rows untouched | untouched — the canary page holds no locked rows, and `tool-loan-vs-savings`' locked row was separately proven intact under 189 |

**The rewrite is on-spec, not merely different.** prose-0 now opens *"If you've
ever looked at your monthly…"* and prose-1 *"If you'd like to see these numbers
applied to your own loan…"* — the conditional/situational opening the canary
brief demanded — with the heading structure and the Main Loan Calculator link
preserved. So the framework produced the intended artefact, which is the thing
this bug was blocking.

**Proof the save ran rather than no-opped** (a carried run would show an
unchanged row count and slot list): both rows have NEW ids — `a608c953`,
`b05e3477` against the baseline `efdb1a61`, `79e62948` — stamped `11:53:01`.

**Verified at the ARTEFACT, not the status.** Cache-busted fetch of
`https://loancalculator.co.uk/guides/how-loans-are-calculated.html`: HTTP 200,
17,619 bytes, 2 `<section>` blocks, the new opening present **1**, and the
pre-fix opening (*"Most people see a monthly loan repayment as a flat fee"*)
**0** — a negative control on the served page, not just a positive match.

**A second bug's fix is proven in the same run:** the slot names came back
`prose-0`/`prose-1` and `pages.sections` still reads `["prose-0", "prose-1"]`.
Pre-`bugs_open/189` this build path would have renamed both to `ported-prose`.
(`data-component="ported-prose"` ×2 in the markup is correct and unrelated —
that attribute carries the component's function; the slot name lives in the DB
column, and that is what rebuild-ability depends on.)

### The full chain, for the record

`13252f714` (Path 0, council APPROVED `d3e232b8`) → live v1.0.1257, pod-grepped
with a fabricated-string control at 0 → read-only proof that 57/57 unresolvable
names ARE resolvable by stored id → `92e14493b` + config (189, APPROVED
`87444080`, PBP-035) removed the save-path gate → this canary. **Unblocks** the
owner's 2026-08-05 instruction to rerun loancalculator's copy through the
framework in the H voice: the mechanism is now proven on that site's own pages.

---

## CONTRIBUTION 2026-08-17 — the same blindness may live in a THIRD path: `write_site_plan` produced 0 sections for 41 positional-slot pages, on a site where this fix is live

From the `loanandmortgagecalculator.co.uk` D6 planner lane. **Not a reopening, and not a
claim about this fix's correctness** — the id route is live and this lane has no evidence
against it. What we have is a fresh instance of the same *shape* in a path this file's chain
does not name.

**The run.** One `build-site-planner` canary, site `ed633ada-f8af-424b-b4d4-8af79160dbcd`,
corr `6fe6ee93-67b9-4831-bf17-2ca473e1d30c`, COMPLETED 2026-08-17 12:07:05Z on chassis
`v1.0.1305` (well after `v1.0.1257`).

**What the plan carried.** `site_plan_sections` for the resulting plan holds **10 rows
total**, and all 10 belong to the four pages the FRAMEWORK itself built on 08-15
(`hero`/`article-body`/`call-to-action` on two new guides; the complaint-checker tool's
four). The **41 adopted pages got zero sections**, and `sync_pages_to_db` then wrote
`sections = []` onto 24 of them in the live `pages` table (restored from a snapshot within
the half hour; the live site never served it).

**Why it looks like this file's class.** Those 41 pages carry positional slot names —
`prose-0`, `tool-1`, `prose-2` — which are neither a component `name` nor a `function`,
which is precisely the lookup this file measured at 0/57 on the sibling site.

**What we have NOT established, and are not asserting:** whether the drop happened in the
name resolver at all. The same run also failed to preserve realised page identities with
`honour_realised_identity='true'` set, and a reconciler that never PAIRED a plan page with
its realised row would produce both symptoms at once — no pairing, so no sections carried
and no identity honoured. Filed for diagnosis rather than guessed: 090 run correlation
`33d4d7bc-62f8-4886-a8e2-7c39f0c0a302`.

**Explicitly NOT `bugs_open/282`.** 282 is validate's resolver missing tool-LEVEL functions;
these sections name nothing at all, so an acceptance-side fix cannot reach them.

Evidence, if this turns out to be yours: the plan rows above, plus snapshot table
`pages_bak_20260817_preplan_lmc` which holds the pre-run `sections` for all 46 pages. Full
account: `docs024_key_docs_latest/loanandmortgagecalculator_couk/NOTES_…md` entry
**2026-08-17 (d)**.

---

## CONTRIBUTION 2026-08-20 (from the `bugs_open/215` same-name identity lane) — a THIRD call site of this blindness, and it is the worst of the three: `validate_plan` DELETES the sections rather than deferring them, and takes the plan-time fact assignments with them

Found by firing a replan canary at `loanandmortgagecalculator.co.uk` (corr
`313368d2-b9ac-47d4-9465-da087eaf94f7`, chassis `v1.0.1320`) to prove an unrelated
identity fix. The identity half worked; **41 of 45 live pages had `pages.sections`
emptied**, and this file's defect is why. Detected by a pre-fire digest, contained, and
fully reversed within the hour (see below) — so this is a measurement, not an outage.

### The new call site

This file documents `plan_sections` (build path) and cites `bugs_closed/182` for
`rerender_page_sections`. The third is **`ValidateSitePlanAction`'s `validate_components`
resolver** (`v3_site_actions.go`, the block whose own comment reads *"unresolvable names
are dropped + logged"*), enabled by `"validate_components": true` in the
`build-site-planner` and `site-planner` step configs.

Same blindness, worse consequence. `plan_sections` **defers** an unresolvable slot;
validate **drops** it, and the drop happens *before* the object-form→string
normalisation, so the RFC_016 plan-time fact assignments travelling inside each entry
are destroyed with the entry.

### The chain, read from the run's own stored output rather than inferred

1. `load_existing_pages` returned the page correctly:
   `"sections": "[\"prose-0\", \"tool-1\", \"prose-2\"]"` (stringified jsonb — handled).
2. The planner proposed the **right** composition, in object form, with facts attached:
   `[{"name":"prose-0","facts":["sdlt-standard-nil-band-upper", …8 ids]}, {"name":"tool-1", …}, …]`.
   So this is not planner disobedience — the plan was correct.
3. `collected_data->'validate_plan'->'pages'` for that page: **`"sections": []`**.
   `prose-0` / `tool-1` / `prose-2` are positional slot names, not
   `content_components.function` values (the functions are `ported-prose` etc., attached
   via `page_components.component_id`) — so all three were unresolvable and dropped.
4. `write_site_plan` wrote **10 `site_plan_sections` rows for 45 planned pages**.
5. `sync_pages`' upsert does `sections = EXCLUDED.sections` **unconditionally** (no
   COALESCE, unlike its `nav_label`/`meta_description` neighbours), so `[]` overwrote
   three real slot names on each affected page.

**Population on this site [MEASURED 2026-08-20]: 41 active pages carry at least one
positional slot name** (`~ '-[0-9]+$'`) — and 41 is exactly the number that emptied. The
match is the confirmation.

### Why this is not `bugs_open/282`

282 is the same resolver dropping **tool** sections the widened menu had legitimately
offered — a menu-membership problem, fixed by teaching the resolver what the planner was
offered. This is different: a positional slot name is not in any menu and never will be,
because it is not a component identity at all. **Fixing 282 does not fix this.** The fix
is this file's fix — resolve `page_components.component_id` first, name second, exactly
as `bugs_closed/182` did for the re-render path — applied at a third call site.

### Severity, and why it is higher than it looks from the re-render path

`pages.sections` is the *only* record of a decomposed page's composition. Once emptied:
- `page_components` still holds the real rows (82 unchanged here, and the live artefact
  never moved — all real URLs 200, phantom paths 404 at the control's byte size), so the
  **site keeps serving** and nothing looks wrong;
- but the page's composition is now unrepresentable in the plan, so the next rebuild has
  nothing to rebuild, and every `needs_page` the run filed would have built an empty
  page over a live one.

That last part is the live hazard: the run filed **20 `needs_page` + 1 `needs_rerender`
+ 1 `needs_design` + 5 `needs_imagery`**, all `triaged` and claimable within seconds. I
cancelled all 32 (assertion inside the transaction, reason recorded on each row) and
restored `sections` on the 41 pages from a pre-fire snapshot, asserting both the sections
digest and the identity digest back to their pre-fire values inside the repair
transaction. **Anyone reproducing this must take the snapshot first and cancel the queue
before repairing** — the queue is what turns a DB-only regression into a deployed one.

### What this adds to this file's own fix candidates

The existing candidates target `plan_sections`. Add: **whatever fixes the resolution must
be applied at all three call sites, or the next one re-lands the damage** — and validate
is the one that must additionally *preserve* the entry rather than drop it, because an
unresolvable name there is not evidence the section is junk (here it was evidence the
site is decomposed). A conservative shape for validate alone: leave an unresolved name in
place when the page's realised `sections` already contains it, i.e. trust stored state
over the component catalogue.

**Cross-reference:** register PLAN-048's landmine already says *"do not opt the decomposed
sites in until 204 is fixed"* — that instruction is correct, and the reason it gives
(`normaliseRealisedToPlanPage` carrying positional names verbatim) is **adjacent to but not
the actual route**. The damage arrives through validate's resolver, which makes the
exclusion broader than its stated rationale: it applies to any replan of a decomposed
site, whatever the identity flags say. I did not read that landmine before firing —
recorded in `WRONG_CALLS.md`.

---

## §THE THIRD CALL SITE IS FIXED — committed 2026-08-21, INERT until the next chassis roll

**Do not read this as closed.** The Go is committed and the test suite is green; it
is inert until an image is built and rolled. The 08-06 half of this file (the BUILD
path) remains live and proven — this section is only about the `validate_plan` call
site the 08-20 contribution found, plus two of `apply_gap_plan`'s three.

Commits, in order: **`d376ca9b8`** (A, the shared reader) → **`7baaf513b`** (a hotfix,
see below) → **`c6446f5da`** (B, the rescue and the register). Council correlation
**`f73f4eeb-5d79-482c-bc9b-b33f0ab64f76`** (`Council-Submitted:` — the verdict was
not read before committing, and must be read before anyone calls this done). Lane:
`docs/agent_docs/docs024_key_docs_latest/bugfix_204_stored_slot_identity/`.

### What the evidence turned out to be, and it is sharper than this file had it

`bugs_open/282`'s lane shipped a durable record of every dropped section name on the
08-16 roll. Since then [MEASURED 2026-08-21]:

```sql
SELECT action, count(*) AS drops,
       count(*) FILTER (WHERE context->>'section' ~ '-[0-9]+$') AS positional_shaped,
       count(DISTINCT context->>'page') AS pages, min(occurred_at)::date, max(occurred_at)::date
FROM agent_error_log WHERE error_code='PLAN_SECTION_NAME_DROPPED' GROUP BY action;
--  validate_plan | 140 | 140 | 41 | 2026-08-17 | 2026-08-20
```

`prose-0` 70, `tool-1` 34, `prose-2` 18, `tool-0` 12, `prose-1` 6. **140 of 140.**
The class `validate_components` exists to catch — display names, typos, stale
functions — has produced **zero** records; the class it cannot see has produced every
one. ⚠ It is a **lower bound**, not a total: the record only exists from 08-16, so do
not quote `min(occurred_at)` as the date the drops began. ⚠ The column is
`occurred_at`, not `created_at`.

Census re-measured the same day: **107 unresolvable names across 7 sites** (86/5 at
filing, 87/6 on 08-06). loancalculator.co.uk is now **0** — that is the 08-06 fix's
own footprint in the census, not a shrinking problem.

### It was four call sites, not the one this file named

| call site | file:line | shipped |
|---|---|---|
| `ValidateSitePlanAction` `validate_components` | `v3_site_actions.go:3838` | **fixed** — the only site with measured damage |
| `applyAddToPage` | `apply_gap_plan_action.go:244` | **fixed** — latent exposure |
| `applyRetypeExisting` | `apply_gap_plan_action.go:905` | **fixed** — writes `sections` straight onto a LIVE page |
| `applyNewPage` | `apply_gap_plan_action.go:374` | **deliberately LEFT**, pinned by a test — a new page has no stored rows, so a positional name there points at nothing and dropping it is right |

⚠ **The two gap-plan arms are LATENT, not measured.** `content-gap-planner` is live
(46 `gap_plan_*` items, most recent 08-20) but has recorded **zero** drops since
08-16. The evidence there is the code path, not a row. Say "latent", not "damaged".

### The fix, and the constraint it had to survive

`LANDMINES.md` forbids the obvious version: *"do NOT fix the inconsistency by
widening `loadComponentNameResolver` itself"* — three of its four call sites belong
to a path whose menu PLAN-049 records as deliberately un-widened. So the resolver is
**untouched**. Instead each drop branch asks, **per page**, whether that page already
carries a slot under the proposed name (`stored_slot_rescue.go`; the shared reader is
`datahelpers/page_slot_identities.go`, register **PLAN-051**).

A menu widening enlarges what may be newly PLACED anywhere; this grants no placement
at all, only *"keep what this page already has"*. That difference is a property of
the map key, not a claim in a comment —
`TestStoredSlotRescue_IsScopedToTheProposedPage` fails if a slot stored on **another
page of the same site** rescues this one.

Decisions a later reader will want the reasons for:
- **Kept VERBATIM.** Rewriting a positional name to its component's function would
  collapse `prose-0`/`prose-1` onto one name — this file's own candidate-3 rejection —
  and break the page's next id-first resolution.
- **A read failure KEEPS, loudly.** The costs are asymmetric: a junk name surviving is
  deferred one step later by `plan_sections`; an emptied decomposed page is
  recoverable only from a snapshot somebody thought to take.
- **Lazy.** The rows are read at most once, only on the first miss, so a site whose
  names are honest functions issues no extra query — inert on the undecomposed estate
  by construction, asserted through sqlmock's `ExpectationsWereMet`.
- **One durable summary row per run** (`PLAN_SECTION_NAME_KEPT_BY_STORED_SLOT`), not
  one per keep. Without any record, *"the fix works"* and *"the planner proposed only
  catalogue names this time"* produce identical evidence.

⚠ **Corrects the 08-20 contribution's proposed shape.** It suggested trusting the
page's realised `pages.sections`. That names the wrong store twice: it is the column
this bug destroys, and it reaches validate through the `existing_pages` field —
which **`site-planner` does not have** (its live step list is `complete,
load_available_components, load_style_collections, plan_site, validate_plan`; no
`load_existing_pages`), so a fix keyed on it would be inert on one of the two live
consumers. `page_components` is ground truth and is what `plan_sections` already chose.

### The independent read

`090` run correlation **`1588b0da-5657-451a-8dc5-a5f63324712f`** returned
**UNVERIFIABLE at the iteration cap** — not REFUTED, and not a confirmation either.
It independently confirmed the two halves that matter, with its own citations and
its own live evidence: `PLAN_SECTION_NAME_DROPPED` rows naming `prose-0`/`tool-1` for
five pages where `page_components.slot_name` records those same names for those same
pages. It stopped short on the persistence half (`site_db_actions.go` was omitted
from its bundle for size) and on a full call-site enumeration, flagging
`component_selector.go:SelectComponentByType` as unread. **That open question
resolves NO:** the selector never touches `page_components`, and its only caller is
`plan_sections`' `resolveSectionComponent`, reached only after Path 0 has already
tried stored identity. Its other caveat is worth carrying — the code index it read
is stale (mirrors a commit 2 days old), so an absence there is *unknown*, not
*confirmed absent*.

### Council: APPROVED, and it found a real defect anyway

**Corr `f73f4eeb-5d79-482c-bc9b-b33f0ab64f76` — approved, 4 advisory objections,
none high-severity** (`editquality` approve, `bug_historian` object, `reuse_agent`
approve, `guidelines` approve). Both mediums were acted on in `e08450fd3` rather than
filed, and one of them was right about something I had checked and got wrong:

- **The read-failure keep filed NO durable row.** The logs distinguished "kept
  because recognised" from "kept because unreadable"; the durable record did not, so
  a run that kept every name because the database was unreachable read — in the only
  channel that survives log rotation — exactly like a clean pass. That is this bug's
  own silent-absorb shape, reproduced inside its fix. Now
  `PLAN_SECTION_STORED_SLOT_READ_FAILED` is its own code, filed unconditionally on
  failure with `kept_without_checking`; `keptCount()` stays 0 because nothing was
  *rescued*; validate warns per entry. Two mutations pin it.
- **"Check whether the untouched slot loaders already apply the wrong rule."** Done.
  The rerender path's `loadContentComponentsByID` builds no slot map at all, so the
  premise does not hold there. The one that does key on `slot_name` with **no rule
  and no `ORDER BY`** is `enrichSectionComponentsWithBriefs` — a repeated slot with
  differing briefs is a non-deterministic last-write-wins. **[MEASURED 2026-08-21] 0
  such pairs today**, against 1,619 brief-carrying rows across 553 pages and 18
  repeated slot groups fleet-wide: reachable, unoccupied. Latent, recorded in
  PLAN-051, deliberately not fixed here — different question, different lane.

### Still to do — the write-side guard, which is NOT in these commits

> **SUPERSEDED 2026-08-21, same day — this shipped as commit `af318f318`** (PLAN-052,
> council corr `2466d82c-17f8-4ebc-948d-ff8dbab9cee4`). Kept below because the
> reasoning for why it is SEPARATE is still the reasoning, and because the paragraph
> names what is still NOT guarded. See §The write-side guard below.

`upsertPage` (`site_db_actions.go:1201`) carried `sections = EXCLUDED.sections`
**unguarded**, while its `nav_label` and `meta_description` siblings **in the same
statement** were given destructive-write guards on 08-19. One Go caller
(`SyncPagesToDBAction`), reached by **three** live agents (`build-site-planner`,
`pageflow-builder`, `site-work-orchestrator`). Deliberately a separate commit and a
separate council round: different mechanism, different blast radius, and a class
defence rather than this bug's fix. Design is in the lane's PLAN §5 — including that
zero sections must stay representable (**72 of 748 active pages live there
legitimately, 60 of them tools** [MEASURED 2026-08-21]) via the existing
`recompose_pages` release rather than a new config key.

### How to verify after the roll

1. **Provenance per SERVICE, never `strings`:** the pod's own `build provenance`
   line → `git merge-base --is-ancestor c6446f5da <stamp>`; the line scrolls, so fall
   back to `grep -aq "<sha>" /proc/1/exe` **with a control in the same breath** (one
   sha that must be present, one that must be absent). One pod per *deployment*
   running the chassis image.
2. **Behavioural canary:** the 08-20 replan shape on loanandmortgagecalculator.co.uk,
   with that incident's own containment runbook — **pre-fire snapshot first, and
   cancel the emitted queue before any repair.** Assert `validate_plan`'s pages carry
   the positional names, that a `PLAN_SECTION_NAME_KEPT_BY_STORED_SLOT` row exists
   for the run, and that the post-`sync_pages` `pages.sections` digest equals its
   pre-fire value.
3. **Demand control on the zero.** Positional-shaped `PLAN_SECTION_NAME_DROPPED` rows
   going to zero is the success metric — but induce one drop with an invented name
   and confirm its row appears. **A zero without that control is a blind pass.**
4. **Negative control:** the same shape on an honest-function site → zero keeps. If
   the arm fires there, the scoping property failed; stop and re-diagnose.
5. ⚠ **The 107-names/7-sites census is NOT the fix's metric.** It measures
   decomposition, not damage, and should be unchanged by this fix.

---

## §The write-side guard — shipped 2026-08-21 as `af318f318` (PLAN-052), INERT until the roll

The read-side fix stops `validate_plan` PRODUCING an empty list. This refuses the
empty list **whatever produced it**, including causes nobody has found yet — which is
why it is a separate commit and a separate council round (corr
`2466d82c-17f8-4ebc-948d-ff8dbab9cee4`) rather than part of the fix above.

`sections = EXCLUDED.sections` becomes a four-arm `CASE`: a caller-declared release
wins, then a non-empty incoming list wins (**the plan stays authoritative — a
recomposition still recomposes**), then an empty stored list is freely replaced, and
only the **non-empty → empty** transition is refused. One statement, so no
read-then-write window.

- **Zero sections stays representable** — 72 of 748 active pages live there
  legitimately, 60 of them tools [MEASURED 2026-08-21]. Deliberate emptying travels
  through the **existing `recompose_pages` release**, so there is no new config key
  and nothing for an operator to remember.
- **A refusal is DURABLE** (`PAGE_SECTIONS_EMPTY_OVERWRITE_REFUSED`), naming the pages
  and pointing at the upstream cause. A silent keep would mean the write reports
  success while the plan and the database disagree about what a page is made of —
  this bug in a new place.
- ⚠ **Known edge:** `recompose_pages` names realised page names while the sync loop
  reads the CANONICALISED name, so a page whose canonical name differs may not match
  the release. The failure direction is safe (refused and recorded) but it is real.
- ⚠ **NOT guarded — and this list was CORRECTED by the council, having been wrong.**
  The submission called five other write paths safe "by construction"; the
  `bug_historian` seat objected that this was **asserted, not measured**, citing
  `bugs_closed/001 → 037 → 050` as a history of exactly this guard shipping for one
  path and being found incomplete on a sibling within weeks. It was right, and **one
  of the five was not safe**:

  | path | now checked |
  |---|---|
  | `adopt_verbatim` | safe — always writes one element (`[]string{portedPageSlot}`) |
  | `create_blog_posts` | safe — floors to `hero, article-body, call-to-action` when empty |
  | `apply_gap_plan` retype / `applyNewPage` conflict arm | safe — `defaultSectionsForPage` never returns empty, and the resolved path is floored by `len(resolved) > 0` |
  | **`apply_adoption_plan`** | **NOT safe — now guarded (`8922183a5`)** |
  | `UpsertPageForRole` | different authority, out of scope, not re-examined |

  `apply_adoption_plan_action.go` built `sections := []string{}`, filled it only when
  the plan page carried the key, and wrote it through an **unguarded `EXCLUDED`** —
  over LIVE pages, via `ON CONFLICT (site_id, name)`, on the live `site-adoption-agent`.
  **And that statement already carried the `meta_description` guard, commented "Same
  guard as upsertPage"** (`bugs_open/320`) — one half of this exact omission had been
  fixed on this exact statement and the other left. The 001→037 shape, a second time,
  on a second statement.

  **Do not read even the corrected list as fleet-wide cover for `sections` writes** —
  it is five paths checked on one day, and `UpsertPageForRole` was not among them.

⚠ **Two of five mutations SURVIVED the first pass**, and both gaps were invisible in
a green suite. (1) Mutating the `RETURNING` expression changed nothing, because
**sqlmock never executes SQL** — the test proved only that the Go *scans* the column.
It now pins the expression as text and says plainly that it establishes the statement
asks the right question and **not** that Postgres answers it correctly; that needs the
live induction in the verification list. (2) Making the release **site-wide** left
every test green, because they all stopped at `upsertPage`'s boundary — one
legitimately-recomposed page would have licensed emptying every other page in the run.
**Generalisable: a guard tested only at the helper's boundary is untested against the
wiring that decides its input.**
