# NOTES — bugs_open/308, CTA destination provenance (append-only, newest at the bottom)

## 2026-08-22 — lane opened; the bug re-verified before any design work

**Why a lane at all.** `bugs_open/308`'s routing line says *"Route to
`cta_target_content_pass` rather than opening a competing lane"*. Read against that lane's
own PLAN, the two are not the same object: that lane is a **content pass** (reword CTA
LABELS so the existing resolver picks better targets) and carries the widening only as a
phase-1 open question. 308's owner ruling of 2026-08-18 chose **fix candidate 1 — build a
real provenance record**, which is platform Go + a shared seam. This lane does that half.
The content pass is not competed with; a CONTRIB goes into its NOTES and into the bug file.

Checked before opening: `scripts/who-owns.py 308` → OWNED/recently-active
(`cta_target_content_pass`, 2 commits/14d, last 2026-08-18). Its files were read in full
before writing anything here. `git status` on the files this fix would touch
(`resolve_internal_links_action.go`, `rerender_page_sections_action.go`,
`discovery_checks/`, `datahelpers/`, `save_page_sections_action.go`,
`section_editor_actions.go`) → **clean**, so no in-flight session is mid-edit on them.
That check is LAGGING by construction and is re-run at each phase boundary.

### The bug is STILL VALID, and it has grown [MEASURED 2026-08-22, live DB]

The file's own query, re-run verbatim:

```sql
SELECT count(*) FROM site_work_items swi, LATERAL jsonb_array_elements(swi.spec->'findings') f
WHERE swi.item_key LIKE 'misdirected_cta:%'
  AND f->>'suggested_target' ~ '^/(contact|about|privacy|terms|legal)(\.html|/|$)';
```

- filed 2026-08-17: **149**
- today 2026-08-22: **200**

Split by status — and this split is the churn made visible:

| status | items | findings |
|---|---|---|
| `complete` | 71 | **112** |
| `unresolved` | 53 | 86 |
| `cancelled` | 2 | 2 |

**112 findings sit on work items the platform marked `complete`.** That is the bug's
sentence "the repair runs, completes green, and changes nothing" as a number.

### DEMAND CONTROL — the flow has stopped, and that must not be read as a fix

Before treating 200 as a current rate, I asked whether the detector is running at all
(`a-post-fix-zero-needs-a-demand-control`, and this is its mirror image — a *stock* read as
a *flow*):

```sql
SELECT date_trunc('day', created_at)::date, count(*) FROM site_work_items
WHERE item_key LIKE 'misdirected_cta:%' GROUP BY 1 ORDER BY 1 DESC;
```

→ 08-19: **3**, 08-18: 128, 08-17: 208, 08-16: 28, 08-15: 21, 08-14: 84 …

So **nothing has been detected for three days**. 200 is a stock. `[INFERRED, not verified]`
the cause is `bugs_open/230` (site discovery has no recurring driver) — I have not opened
230's mechanism to confirm it, and it does not change this bug's design either way. What it
DOES change: any post-fix measurement of this population is worthless without first
inducing a discovery run, because the number will sit at 200 whether the fix works or not.

### The finding I did not expect: `suggested_target` HAS NO CONSUMER

```
grep -rn "suggested_target" --include=*.go platform/ internal/ pkg/
```
→ three hits, ALL of them writers or a test: `check_misdirected_cta.go:130`,
`check_cta_nonpage.go:79-80`, `check_cta_nonpage_test.go:145`. **Nothing reads it.**

The detector emits a `page_rerender` item with `spec.reason = "cta_links_stale"`;
`rerender_page_sections_action.go:528` gates its CTA recompute on that *reason string* and
then re-derives the destination independently from `candidatesFromHubs`. So the bug is one
rung deeper than "two candidate universes": the detector's answer is **computed, written
down, and thrown away**, and the repairer re-answers the same question from less
information. Two universes is the symptom; *the repair not consuming the detection* is the
shape — and it is the same shape as `bugs_open/071` (the content gate detects every broken
link then discards the finding), which is open on a different seam.

That generalisation is this lane's argument for a framework-level fix rather than a CTA
patch, and it is the thing to hold the design to.
