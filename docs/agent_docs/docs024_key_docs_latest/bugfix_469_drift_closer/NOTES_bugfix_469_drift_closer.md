# NOTES — `bugs_open/469`, the detector half (a closer for `check_section_source_drift`)

Append-only, newest at the bottom. Technical log: evidence, commands, what the system
actually said, and every misstep.

Lane opened 2026-09-03 by session "bugs_open/469". The bug was filed earlier the same day
by the `bugs_open/427` lane and left explicitly **OPEN, unowned**; no `469` session and no
lane directory existed, so this is a resume, not a competing thread.

---

## 2026-09-03 — session 1

### Ownership check, before anything else

`scripts/who-owns.py 469` returns **OWNED or recently active**, naming
`bugfix_427_event_render` (8 mentions) and `idea_uk_vm_site` (6). Both are **citations**,
not ownership of the fix: 427 FILED the bug while triaging its own backlog, and idea.uk
replied to a notification about one of its pages. `ListAgents` shows no session named
`469`. The bug file's own status line says "OPEN, unowned".

**So `who-owns.py` says OWNED here and the honest answer is "filed by, not worked by".**
Worth noting because the script's verdict is deliberately conservative and a reader could
stop at it. The tell was that every hit is a cross-reference in another lane's docs, and the
bug file itself declares no owner. Confirmed by messaging the 427 session directly, which
replied agreeing the split (see below).

### Is the bug still valid? [MEASURED 2026-09-03]

Three separate questions, answered separately.

**(a) Is there live drift right now? No — zero, with demand controls.**

```sql
-- tier 1 (site_plan_sections for the current plan) vs pages.sections
WITH tier1 AS (
  SELECT sp.site_id, sps.page_name,
         jsonb_agg(sps.component_name ORDER BY sps.ordering) AS auth
  FROM site_plan_sections sps JOIN site_plans sp ON sp.id=sps.plan_id
  WHERE sp.is_current GROUP BY 1,2),
cache AS (
  SELECT p.site_id, p.name AS page_name, p.sections AS cache FROM pages p
  WHERE COALESCE(p.status,'') <> 'deleted'
    AND jsonb_typeof(p.sections)='array' AND jsonb_array_length(p.sections)>0)
SELECT count(*) AS joined, count(*) FILTER (WHERE t.auth = c.cache) AS agree
FROM cache c JOIN tier1 t ON t.site_id=c.site_id AND t.page_name=c.page_name;
```

→ `joined = 398, agree = 398`. Tier 2 (the `site_specs.site_plan` aspect, restricted to
pages tier 1 does not cover): **34 compared, 0 diverged**.

**The demand control matters and I ran it deliberately.** A first cut printed
`0|0|0` and a bare zero here has two causes with opposite meanings — no drift, or a join
that matched nothing. The control says 398 and 34 comparisons were actually made.

**And the zero is CONSERVATIVE, not optimistic.** My SQL does not apply the loader's
locked-row merge (`datahelpers.MergeLockedPageSlots`), which the real check applies to both
sides. The merge only ever inserts locked rows into BOTH lists, so it can turn a raw
mismatch into agreement, never agreement into a mismatch. A raw zero therefore implies a
merged zero. (Stating the direction because "I did not model the merge" would otherwise
read as an unquantified gap.)

**(b) Are all six items closed? Yes** — all six `section_source_drift` rows are `complete`,
closed by migration `753` earlier today, three of them with `direction: "authority_won"` in
the receipt.

> **Caveat from the 427 lane, and it is a good one.** Closing those items FREED their dedup
> keys (`idx_swi_dedup` excludes terminal statuses), so a page that is still genuinely
> broken will re-file on its next completeness sweep. Until each of the six sites has had a
> discovery pass, "zero open items" is partly "not yet re-asked". **This does not weaken
> (a)** — (a) compares the STORES directly and never consults the item table — but it is the
> right way to read the item count, and a new `section_source_drift` row appearing overnight
> would be a real finding rather than a duplicate.

**(c) Is the damage still there? Yes.** `robot-hands.com/gripper-catalog`:

| store | today |
|---|---|
| `site_plan_sections` (current plan) | `[hero, generic-text-block, info-card-grid, call-to-action]` |
| `pages.sections` | identical |
| live `page_components` (4 rows, all `deployed`, none locked) | identical |

`gripper-spec-sheet` is genuinely absent from that page. 469 §2 stands.

### A premise I nearly built on, and the check that refuted it

I was about to propose a new `page_sections_history` table + trigger, reasoning that a
destroyed composition is unrecoverable. **Before designing it I asked whether the estate
already records this, and it partly does.**

```sql
SELECT count(*) FROM page_component_history WHERE slot_name ILIKE '%gripper-spec%';  -- 24
```

Migration `357`'s trigger pair (`trg_page_component_artefact_archive_upd/_del`) archives
every deleted `page_components` row with `slot_name`, `position`, `rendered_html` and a
`divergence` classification. So from ~2026-08-09 onward a dropped section's **identity and
bytes survive the loss** — 24 rows for `gripper-spec-sheet` alone (on `product-detail` and
`gripper-detail`, ~12.4 KB each; it was never lost from those pages).

**What does NOT survive is the LIST.** `DELETE`+`INSERT` is the rebuild lifecycle, so the
history holds a delete for *every* section on *every* rebuild; "which one was dropped" is
only derivable by diffing consecutive builds. That is the gap worth a receipt — and it is a
much smaller gap than a second archive table.

**Cost of the check that caught this: one `count(*)`. Cost had I skipped it: a new table, a
new trigger on a hot path, and a council round defending it.** Logged in `WRONG_CALLS.md`
as a near-miss, because "reuse existing machinery before building new" is a standing rule
here and I was one step from breaking it.

### The class measurement (why this is framework-shaped, not one check's problem)

- 71 discovery check files; **19** populate `CheckResult.Resolved`.
- **18** file at least one flag-only item (`HandlerAgent: ""`); **10 of those have no closer
  at all**, `check_section_source_drift` among them.
- Handler-less open items fleet-wide, top of the tail: `capability_gap` 331 (oldest 42
  days), `owned_page_review` 171 (47), `cta_names_unknown_destination` 107 (51),
  `image_source_unsatisfiable` 77 (48), `needs_section_data` 43 (**172 days**).

So the accumulation 469 §3 describes is a fleet-wide pattern, not a quirk of one check.
This lane fixes one instance of it properly rather than all of them shallowly.

### The check runs often enough for a closer to matter

`completeness-discovery-agent` (whose live `run_checks.config.checks` array does contain
`section_source_drift` — 46 checks) filed items on **22 sites today, 18 yesterday, 17 the
day before**, every day for the last ten. A `Resolved` arm therefore fires within about a
day for an active site. (This counts items CREATED, so it is a lower bound on runs.)

### The sharp edge, restated because it is the whole design

A naive closer — "the stores agree again, so retract" — would be **worse than today's
silence**: it would automatically and silently ratify the destruction, fleet-wide, forever.
Migration `753` did the classification by hand and got three `authority_won` out of six.
Whatever this lane builds must make that classification mechanical and unskippable.

---

## 2026-09-03 — session 1, continued: §5.1 answered, and a fact nobody had

### The owning lane answered: it is damage

I messaged the `robot hands` session rather than guessing. It replied with provenance I
could not have derived: `docs/agent_docs/docs024_key_docs_latest/robot_hands/SQL_2026-07-24_r9_gripper_catalog_real_grid.sql`
put `gripper-spec-sheet` on `gripper-catalog` at position 3 on 2026-07-24, deliberately
*instead of* `product-grid` — the migration's own comment says product-grid's e-commerce
fields (price/rating/badge/image) would be empty for grippers and "invite fabrication",
where spec cards are "the honest fit". The lane's HANDOFF frames it as an owner call. They
checked the rest of that lane for a later reversal and found none.

**So §5.1's factual half is closed: the absence is this bug's handiwork, not a change of
mind.** That is a fact no query would have produced — it came from asking the lane that
made the decision.

### The fact nobody had: the page is ARCHIVED, and it still serves

```sql
SELECT name, status, build_status, url, updated_at FROM pages WHERE site_id=<robot-hands>;
```

`gripper-catalog` → `status = 'archived'`, `build_status = 'deployed'`, last built
**2026-08-11**, while every *active* page on the site rebuilt today at ~16:20-16:56.

Archived is not "not serving". Probed at the recorded URL with both controls, because a
composed URL's 404 has filed a false bug on this estate before:

```
./scripts/probe-page-url.sh robot-hands.com gripper-catalog gripper-catalog-index
gripper-catalog        /gripper-catalog.html        200 SERVING
gripper-catalog-index  /gripper-catalog/index.html  200 SERVING
controls: invented=404 (want non-200)  sibling=200 (want 200)
```

**Both controls hold, so the 200 means something.** The loss is visitor-facing.
`gripper-catalog-index` is NOT a replacement — it carries a single `news-listing` section.

### Why the repair is BLOCKED, and it is not an excuse

Migration `750` is the only council-approved template, and it does not transfer:

| | 750 (boxingonline) | this (gripper-catalog) |
|---|---|---|
| shape | in-place RENAME at a fixed `ordering` | **INSERT** at ordering 2, shifting two rows down |
| direction | plan aligned to an already-correct page | page and plan BOTH wrong; a section must come back |
| `site_plans` rows on the site | exactly 1 (pre-check asserts it) | **5** — the "nothing superseded" premise fails |
| effect | none on the artefact, by design | the page must REBUILD to render the restored section |

And the blocker proper: the page is `built_from_plan_version = <current plan>` with
`build_status = 'deployed'`, so `reconcile_site_plan_action.go`'s `decideEmit` returns
`skip_built` **before** it compares any section list. A corrected plan would sit there and
never render. Making it render means withdrawing the build stamp — which is
**RFC_064 §7 question 2, open with the owner**. Here it is load-bearing, not theoretical:
without the ruling the repair is a no-op.

Renumbering risk, checked rather than carried across from 750's general warning: all four
plan rows have `assigned_fact_ids = NULL` and `subject = NULL`, and the page's only
`site_plan_imagery` row is page-scoped (`scope_ref = 'gripper-catalog'`), not
`<page>:<ordinal>`. So the shift is low-risk **on this page**; the warning still stands in
general.

**Decision: write the repair as `_HOLD.sql` with full pre/post checks, route the go/no-go
to the owner alongside RFC_064, do not apply.** Both affected lanes told.

### An archived page that still serves is a third question

Even with the stamp withdrawn, whether an archived page is reachable by the build path at
all is unknown to me. Recorded as an open question rather than assumed either way.
