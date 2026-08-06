# 189 — resolving a previously-unresolvable LOCKED section duplicates it on the page

**Filed 2026-08-03**, discovered while inducing the live verification for
`bugs_closed/182` (component_id-first resolution) on loancalculator.co.uk's
`tool-loan-vs-savings` page. **This is a real consequence of 182's fix, not a
flaw in it** — 182's resolution is correct and desired; this bug is a
pre-existing defect in `save_page_sections_action.go` that 182's fix newly
*reaches*, because before 182 these sections never resolved at all.

## What happened, measured

Firing the documented `section_data_resolved` re-render on
`tool-loan-vs-savings` (work item `b46a134b-c17c-4f6b-8aa1-23fa942ec354`,
2026-08-03 11:12 UTC, chassis v1.0.1240 — 182's fix) produced **five**
`page_components` rows where there should be four, and the served page
(`https://loancalculator.co.uk/tools/loan-vs-savings.html`) rendered the
loan-vs-savings calculator **twice**:

```
 slot_name             | position | locked | html_len
 ported-prose          |    1     |  f     |   728
 ported-prose          |    2     |  f     |   200
 tool-loan-vs-savings  |    3     |  f     |  11844   <- NEW, fresh render
 ported-prose          |    4     |  f     |   685
 tool-2                |    5     |  t     |  11845   <- OLD, locked 2026-08-02
```

Both rows shared `component_id='448422ce-fbf0-4e3d-98a1-fab0a6e856ed'` and
**identical `content_data`** (same md5) — the fresh render reproduced the
locked, proven content almost exactly (1-byte difference, incidental). The
served page had 5 top-level `<section>` blocks and `data-component="ported-prose"`
×3 where there should be 3 distinct positional names.

**Remediated live, same session**: deleted the duplicate (position 3, unlocked),
repositioned the locked row back to position 3, restored the three prose rows'
slot_names from the generic `ported-prose` back to their original `prose-0`/
`prose-1`/`prose-3`, and fired an assemble-only redeploy (no `reason`) to ship
the correction. Verify this landed before reading anything else in this file as
current: `curl -s https://loancalculator.co.uk/tools/loan-vs-savings.html | grep -c '<section'` should read 4.

## The mechanism — two pre-existing defects that compound

**1. `extractSectionsFromMetadata` prefers `component_function` over
`component_name`** (`save_page_sections_action.go:896-902`):

```go
componentName := "section"
if fn, ok := m["component_function"].(string); ok && fn != "" {
    componentName = fn
} else if name, ok := m["component_name"].(string); ok && name != "" {
    componentName = name
}
```

`RerenderPageSectionsAction`'s successful-render entry sets BOTH fields —
`component_name: s.slotName` (the stored positional name, e.g. `tool-2`) and
`component_function: comp.Function` (the component's own identity, e.g.
`tool-loan-vs-savings`) — and has done so unchanged since before 182. Before
182, a positionally-named slot NEVER reached this success path (nothing
resolved, so `carryStoredSection` ran instead, which sets only
`component_name`, no `component_function` key at all) — so this precedence
rule was dormant for exactly the population 182 fixes. 182 makes resolution
succeed, `component_function` gets populated, and the persisted `slot_name`
silently becomes the generic component identity — **destroying the
deliberate positional naming** the loancalculator decomposition chose
specifically "so that a dropped-section warning names which paragraph
vanished" (`bugs_closed/182`'s own rationale).

**2. `matchLockedRow` matches by `section.ComponentName`**
(`save_page_sections_action.go:586`):

```go
if lr := matchLockedRow(lockedRows, section.ComponentName); lr != nil {
    lr.consumed = true
    // reposition the locked row, discard the fresh copy, continue
}
```

The locked-row guard (`bugs_closed/058`) is supposed to make exactly ONE of
{locked row, fresh copy} survive. It works by matching the INCOMING section's
name against the locked rows' `slot_name`. Once defect #1 renames the incoming
section from `tool-2` to `tool-loan-vs-savings`, the match against the locked
row (still named `tool-2`) **fails silently** — `matchLockedRow` finds
nothing, the guard never fires, the fresh copy is inserted as a **new** row,
and the locked row (excluded from the DELETE-all by its own protection) also
survives. Both defects individually look like "reasonable behaviour"; together
they mean **a locked, positionally-named section is duplicated, not
protected**, the first time it becomes resolvable.

## Blast radius — measured 2026-08-03, re-run before trusting

Sections that are BOTH positionally-named (unresolvable by name/function,
resolvable only by `component_id` — 182's repair population) AND currently
locked:

```sql
SELECT s.domain, count(*) FROM page_components pc
JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
LEFT JOIN content_components cc  ON cc.function = pc.slot_name
LEFT JOIN content_components cc2 ON cc2.name    = pc.slot_name
WHERE cc.function IS NULL AND cc2.id IS NULL
  AND pc.component_id IS NOT NULL AND pc.locked_at IS NOT NULL
GROUP BY s.domain;
--  loancalculator.co.uk | 12
--  oufe.com             |  2
```

**14 sections across 2 sites are armed**: any `section_data_resolved` (or
`image_landed`) re-render that reaches one of these will duplicate it exactly
as above, on the next roll or the next manual fire — whichever comes first.
This is a LATENT trap `182`'s fix newly created reachability for; it did not
exist before `182` because these sections could never resolve, so the locked
guard's silent mismatch never had a fresh copy to fail to catch.

## Fix candidates, ordered by what closes the door

1. **Match `matchLockedRow` by the STORED slot_name / position, not the
   post-resolution `ComponentName`.** `RerenderPageSectionsAction` (and any
   other producer of `sections_metadata`) already knows the ORIGINAL
   `page_components.slot_name` for every section it re-renders — thread it
   through as a stable key (e.g. a `stored_slot_name` field on the metadata
   entry, separate from the display/diagnostic `component_name`) and match
   locks against THAT, never against a name that can change between the read
   and the write.
2. **Stop letting `component_function` silently overwrite the stored
   `slot_name` in `extractSectionsFromMetadata`.** The positional name is
   information (which section this is), the function is a different piece of
   information (what renders it) — collapsing them into one field is the
   underlying defect. Candidate 1 is required regardless of whether this one
   is taken; this one prevents the *diagnostic* regression even outside the
   locked case (any positionally-named, unlocked section loses its identity
   the first time it resolves, fleet-wide, once 182 is live — 65 sections,
   not just the 14 locked ones).
3. Weakest: detect the resulting duplicate after the fact (same `component_id`
   twice on one page) and raise a work item. Does not prevent the live
   duplication, only shortens how long it's visible.

**1 and 2 are complementary, matching 182's own candidate reasoning**: 1 stops
the duplication (the acute harm), 2 stops the silent rename (the property this
site's design depends on) even where nothing is locked.

## How to verify a fix

Do NOT re-induce on `loancalculator.co.uk` again casually — every fire of
`section_data_resolved` on a page with one of the 14 armed sections currently
reproduces this defect until it's fixed. To verify a fix without touching
production content: fire the same re-render on `tool-loan-vs-savings` (page_id
`558f9f3f-ebac-4e4a-8265-30721054f351`, site_id
`0162cde4-633e-45e9-8ca6-87a6b2fe1d26`) and confirm the result is **exactly 4**
rows afterward, `tool-2` still locked at position 3, `locked_at`/`locked_by`/
`id` unchanged (058's own invariant) — not 5.

## Related

- `bugs_closed/182` — the fix whose success exposed this; not a defect in 182
  itself. Cross-linked both directions.
- `bugs_closed/058` — the lock-preservation guard this defeats; its own
  verification ("locked row's id/md5/updated_at unchanged, unlocked sibling
  rebuilt") was true for every case it was tested against because none of
  those cases involved a NAME CHANGE between the locked row and its own
  incoming re-render.

## Diagnosis loop

Not run through `090` — filed from direct, first-hand verification: the exact
live rows (before/after, md5-compared), the exact two functions read and
quoted above with line numbers, and the mechanism reproduced once, deliberately,
in a controlled test that was then remediated in the same session. Per the
2026-07-31 owner ruling's stated escape hatch: this substitutes for the loop
because the causal chain is fully read, not inferred, and re-running it live to
generate a second data point would recreate the very duplication being
reported.

---

## §Blast radius extension 2026-08-06 — the BUILD path becomes a second armed route (added by the 204-fixing session, 7fffb7ef)

`13252f714` (the `bugs_open/204` fix, committed 2026-08-06, inert until an image
rolls) gives the BUILD path (`plan_sections` → `page-content-writer` →
`compile_page_sections` → `save_page_sections`) the same component_id-first
resolution 182 gave the re-render path. **That makes this bug's trap reachable
from a second direction, and the build path's version is WORSE for the rename
half**, because unlike the re-render path it never carries the stored slot name
at all:

- `RenderComponentAction` (`v3_site_actions.go:1899-1903`) outputs
  `component_name: comp.Name` and `component_function: comp.Function` — the
  COMPONENT's identities. The planned section's positional name
  (`sectionPlanItem.Name`, e.g. `prose-0`) is on `current_section.name` and is
  copied into NEITHER.
- `extractSectionFromMap` (`v3_site_actions.go:2142+`) forwards only
  `component_id`/`component_name`/`component_function`/`content_data` into
  `sections_metadata`.
- So `extractSectionsFromMetadata`'s function-first preference (§mechanism
  defect 1 above) persists the slot as the component function — on the 204
  canary page BOTH prose slots would come back named `ported-prose`, the
  positional naming destroyed with no field anywhere still holding it. On the
  re-render path at least `component_name: s.slotName` preserved it.
- Defect 2 (matchLockedRow by post-resolution name) then fires identically:
  the 14 armed locked sections (12 loancalculator, 2 oufe) duplicate on any
  build-path run that reaches them, once 13252f714 is live.

**Consequences, until this bug is fixed:**
- The 204 closure canary must run on an UNLOCKED page (or a throwaway page)
  and must expect + restore the slot rename, exactly as this file's
  remediation did; it must NOT run on any page holding one of the 14 armed
  locked rows.
- Fix candidate 1 (thread the stored slot name through the metadata as a
  stable key) now needs the producer fixed on BOTH paths:
  `RerenderPageSectionsAction`'s entries AND the build path's
  (`RenderComponentAction` output or `extractSectionFromMap` — the loop's
  `current_section.name` is available in collected_data at compile time).
- 204's fix is NOT the defect here (same verdict as 182 in §header: resolution
  is correct and desired); this save-path defect predates both and is the
  remaining half of the decomposed-site rebuild story.

---

## §Fix committed + APPROVED 2026-08-06 (session 7fffb7ef) — awaiting a roll, and the config half needs care

**Commit `92e14493b`** (plus the `v3_site_actions.go` half at `1d11827c1`,
swept into another session's commit while this was being written — nothing
lost, both are at HEAD). **Council APPROVED round 1**, corr
`87444080-72e4-43a6-a089-b327a8285563`, 6 advisory objections, none high.
Registered as **PBP-035** in the concept register (the seam: `stored_slot_name`
is a reserved key on the shared `sections_metadata` contract, and
`slot_name_from` a new step-config convention — the architecture seat raised
this twice and it is the standing owner requirement for a shared seam).
Decision also in `doc_notes c23ce8cb` on the `page-content-writer` subject.

**What shipped:** candidates 1+2 as one move. `sections_metadata` gains
`stored_slot_name`, emitted by `RerenderPageSectionsAction`'s success entry,
`carryStoredSection`, and `RenderComponentAction` (via optional
`slot_name_from` config). `extractSectionsFromMetadata` prefers it VERBATIM;
absence is byte-for-byte today's behaviour. `matchLockedRow` untouched — once
the incoming name is the stored slot name, its existing match is correct.
15 tests, 10 mutations, all caught, including a control proving the locked-row
assertion is capable of failing.

### ⚠ Two things the next session must not skip

1. **The config half is UNAPPLIED and the Go half alone leaves the BUILD path
   dormant.** Apply `slot_name_from` on `render_section` and
   `render_from_template` from the `jsonb_set` block appended to
   `docs/agent_docs/sql_for_agents/023_page_content_writer_agent.sql`. Do NOT
   re-run that file's full-workflow UPDATE block — it is stale against live and
   would revert a later `prompt_template` patch that exists only as a pasted
   transcript at the end of the file.
2. **Back it up first (debug_historian, MEDIUM).** `page-content-writer`
   dispatches continuously. The block verifies AFTER the write, which is not
   the same as being able to undo it:
   ```sql
   \copy (SELECT default_config FROM agent_definitions WHERE type='page-content-writer'
          AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL)
     TO '/tmp/pcw_default_config_backup_20260806.json'
   ```
   Rollback is the same `UPDATE … SET default_config = <that file>`.

### Objections carried forward (advisory, none blocking)

- `bug_historian` LOW: **`tool-recreation-handler` is a third producer with no
  `stored_slot_name`** — safe only because it has no structured slot identity
  to offer, which is a fact about that producer today, not an enforced
  mechanism. The same shared-mechanism-fixed-at-some-call-sites shape this
  estate keeps recording. Worth a look if that handler ever gains a plan.
- `editquality` LOW: `carryStoredSection`'s hunk is a no-op today (a carry sets
  no `component_function`, so the save already fell through to
  `component_name`). Kept deliberately so the field's meaning does not depend
  on which branch produced the entry.
- `architecture` MEDIUM ×2 — answered by PBP-035, which states the rule for a
  future producer rather than leaving it in this diff.

### To close (fixed AND live bar)

Roll past v1.0.1257 → pod-grep `stored_slot_name` (**0 today**, measured — the
ready-made negative control) → apply the config (step 1, with the backup) →
then this file's §How to verify, unchanged: fire `section_data_resolved` on
`tool-loan-vs-savings`, assert **exactly 4** rows, `tool-2` still locked at
position 3, `id`/`locked_at`/`locked_by` unchanged. Then `bugs_open/204`'s
canary runs un-gated and both close.

---

## §BEHAVIOURAL VERIFICATION PASSED 2026-08-06, v1.0.1259 — the re-render half is proven; the BUILD half still needs its config

**Live**: pod-grepped one pod per chassis-image deployment (agent-chassis,
business-intel, vet-intel): `stored_slot_name` = **1** on each. It was **0** at
v1.0.1257, measured before the roll — so that earlier zero is this
verification's negative control, and a fabricated string returned 0 in the same
exec, proving the grep discriminates.

**Induced, exactly as §How to verify prescribes.** Work item
`b4de13fb-8b69-4927-9e20-3c457a85bfc2`, orchestration
`b807b035-3865-4068-9687-29f160f6c362`, `reason: section_data_resolved` on
`tool-loan-vs-savings`. Dispatched by kcat to
`system.agent.generic.requests` (pods were 46 min old — past the ~300s
post-restart drop window) and confirmed at the DB, because `kcat -P` exits 0
having sent nothing.

| | baseline (pre-fire) | after | verdict |
|---|---|---|---|
| row count | 4 | **4** | PASS — the old code produced **5** |
| slot names | prose-0, prose-1, tool-2, prose-3 | **identical** | PASS — the old code renamed the prose rows to `ported-prose` |
| `tool-2` row id | `10be4f71-…` | `10be4f71-…` | PASS — locked row preserved, not re-inserted |
| `tool-2` updated_at | 2026-08-02 23:01:03 | **unchanged** | PASS — 058's invariant holds |
| `tool-2` position / locked | 3 / permanent | 3 / permanent | PASS |
| served page `<section>` count | 4 | 4 | PASS |

**The save genuinely ran — this is not a no-op passing as a fix.** Two
independent proofs: the three unlocked prose rows are **NEW row ids**
(`6961ac4f`, `8cccbde8`, `97a73e35`) stamped `11:36:54`, i.e. DELETE+INSERT
actually executed; and the action's own counters report
**`rerendered=4, carried=0`**. A carried re-render would have produced an
identical row count and slot list, so the row-id check is what makes this
gradable rather than merely green.

The verification item finished `blocked` (not `complete`), which is the
platform recording that a lock prevented a full overwrite — the guard doing its
job. No `lock_blocked_change` item was raised this round; not investigated
further, since the row-level evidence above is direct.

### ⚠ STILL OPEN, and precisely why

The BUILD path remains armed: `slot_name_from` is **NOT applied** to
`page-content-writer`, so `RenderComponentAction` still emits no
`stored_slot_name`, the save still derives the name from
`component_function`, and a build-path run over a locked positional slot would
still rename and duplicate. The defect is therefore still reproducible by one
route, which is the fixed-AND-live bar. **One gated command closes it** — the
classifier blocked my write to `agent_definitions` (expected here for
production config), so it needs an operator:

```
! kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 < <(tail -45 docs/agent_docs/sql_for_agents/023_page_content_writer_agent.sql)
```

Self-verifying (`DO $verify$ … RAISE EXCEPTION` refuses the COMMIT unless both
keys read back; expect `NOTICE: slot_name_from present on both render steps`).
Backup taken and validated as JSON beforehand (20,339 B, both keys confirmed
absent): `scratchpad/pcw_default_config_backup_20260806.json`; rollback is the
same UPDATE from that file.

After it applies: confirm `slot_name_from` on both steps, then `bugs_open/204`'s
canary runs un-gated on the BUILD path and grades both bugs at once. Then close
189 and 204 together (`git mv`, naming BOTH paths on the commit).

### §CONFIG APPLIED 2026-08-06 — the BUILD half is now live too

The operator ran the gated command. Output: `BEGIN / UPDATE 1 /
NOTICE: slot_name_from present on both render steps / DO / COMMIT`, and
confirmed independently rather than trusting the NOTICE:

```
render_section=current_section.name | render_from_template=current_section.name
```

So both halves of this fix are now live: the re-render path (behaviourally
proven above) and the build path. **The defect is no longer reproducible by
either route**, which is the fixed-AND-live bar.

Closing verification is now shared with `bugs_open/204`, whose canary exercises
exactly this build path over the two positional prose slots of
`guide-how-loans-are-calculated` — if the slot names come back `prose-0`/
`prose-1` rather than `ported-prose`, that is this fix working on the build
path, and it is graded there.

## ✅ BOTH HALVES FIXED, LIVE AND BEHAVIOURALLY VERIFIED — 2026-08-06, v1.0.1259

**Kept in `bugs_open/` deliberately.** Owner direction 2026-08-06: *"please leave
the bugs that you've found in bugs_open not in the closed bug file."* That
overrides CLAUDE.md's `/bugs_closed/` bar. Do not `git mv` this file.

- **Re-render half** — proven above: 4 rows not 5, positional names intact,
  locked `tool-2` keeping its row id AND its 2026-08-02 `updated_at`,
  `rerendered=4 / carried=0`.
- **Build half** — proven by `bugs_open/204`'s canary
  (orchestration `fa89217a-768b-4f22-bd7b-12209f58cbf3`, 11:53): a build-path
  rebuild of `guide-how-loans-are-calculated`'s two positional slots persisted
  them as **`prose-0` and `prose-1`**, with `pages.sections` still
  `["prose-0", "prose-1"]`. **Before this fix that path would have written both
  as `ported-prose`** — the rename that makes `matchLockedRow` miss. Both rows
  are new ids (`a608c953`, `b05e3477`), so the save genuinely executed rather
  than carrying.

So the defect is no longer reproducible by **either** route, which is the
fixed-AND-live bar, and each route was induced rather than argued.

### What is deliberately NOT closed by this

- **`tool-recreation-handler` remains a third `save_page_sections` producer that
  supplies no `stored_slot_name`** (council `bug_historian`, LOW). It is safe
  only because it regenerates single-tool HTML with no structured slot identity
  to offer — **a fact about that producer today, not an enforced mechanism**. If
  it ever gains a plan, it needs the field. Recorded in PBP-035.
- **The architecture seat's standing question** (`doc_notes d9d67807`): the
  tri-state id-resolution judgement is now written inline at two call sites. A
  THIRD consumer should get one shared helper first, not a third copy.

### Evidence index for a future reader

Pod-grep `stored_slot_name` = 1 on agent-chassis / business-intel / vet-intel
(**0** at v1.0.1257, the negative control) · re-render induction: item
`b4de13fb`, orchestration `b807b035` · build induction: item `996b9619`,
orchestration `fa89217a` · config `slot_name_from` applied and independently
read back on both writer steps · backup
`scratchpad/pcw_default_config_backup_20260806.json` (20,339 B) · council
`87444080` APPROVED · register **PBP-035** · commits `92e14493b` (+
`1d11827c1` for the `v3_site_actions.go` half, swept into another session's
commit while this was being written).
