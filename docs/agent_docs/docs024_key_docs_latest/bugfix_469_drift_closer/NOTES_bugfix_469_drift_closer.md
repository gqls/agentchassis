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

### Correction to my own claim, same session

I told the robot hands lane that archived-but-serving was "a fact nobody had". **Wrong.**
`bugs_closed/359` is the archived-page-still-serving lane and names `gripper-catalog` by
exact byte count on 2026-08-26; its detector is live (migration `648`) and it FIRED:

```sql
SELECT status, count(*), min(created_at)::date FROM site_work_items
WHERE item_type='archived_page_still_serving' GROUP BY 1;
--  detected | 9 | 2026-08-26
```

All nine still `detected`, none triaged, `handler_agent=''`. Two are robot-hands.com
(`gripper-catalog`, `news`).

> **CORRECTED, same session, by the robot hands lane.** I first wrote "all filed in the same
> minute". **False, and I had not looked.** The truth: eight across 2026-08-26 22:32 →
> 2026-08-27 02:02 in per-site pairs (four sites), and a **ninth on 2026-09-02 14:03**,
> standalone. I had read `min(created_at)::date` — a DATE — plus the two robot-hands rows
> sharing a timestamp, and generalised from two to nine. The correction **sharpens** the
> point rather than softening it: the detector is still finding FRESH instances a week later,
> so this is not one stale August batch. Logged in `WRONG_CALLS.md`.

What was genuinely new was narrower and I should have said only that: **the page carrying
469's composition loss is the same page carrying an un-triaged serving flag**, and nobody
had connected them.

### A distinction that will mislead the next stale-backlog sweep

These two look identical in a backlog query — "old flag, nobody acted" — and want **opposite
remedies**:

| | `archived_page_still_serving` (9 open) | `section_source_drift` (469) |
|---|---|---|
| has a `Resolved` arm? | **yes** — one of the 19 | **no** |
| why is it still open? | because the finding is **still TRUE** — the pages really are serving | because the check has no closer (⚠ NOT "nothing ever closes one" — see the correction below) |
| what the item says today | accurate | describes a state that **no longer exists** |
| what it is blocking | nothing | the dedup key — fresh drift on that page cannot be filed |
| the remedy it wants | triage / routing to a handler | a closer that **cannot ratify the loss it just observed** |

**So "an old flag-only item" is not one defect.** A check with a working retraction arm and a
real un-fixed defect produces exactly the same backlog row as a check with no arm at all and
a defect that completed weeks ago. The discriminator is whether the item's own predicate
still holds — which is precisely what nothing re-derives today, and precisely what a
direction-aware closer must re-derive before it touches anything.

---

## 2026-09-03 — session 1, the build

### The design, and the two places the plan changed under scrutiny

Fable drafted the plan against the grounding above. Two of its refinements were better
than my brief and both are now the design:

1. **`lost` is the predicate; `direction` is only the label.** I had proposed classifying
   by migration `753`'s three-way equality. That cannot separate `oufe.com/contact` (the
   authority **restored** a section the cache had dropped) from `gripper-catalog` (the
   authority **deleted** one). Both are `authority_won`. Anything keying on the label
   grades a repair and a destruction identically.
2. **The receipt coupling belongs in `resolveWorkItems`, not in the check.** I had it in
   the check. `resolveWorkItems` has two callers (`discovery_checks.go:274`,
   `work_item_retraction.go:230`), so a control in the check protects one path and every
   future caller inherits nothing.

### Two of Fable's `[UNVERIFIED]` items, measured

```sql
SELECT COALESCE(build_status,'(null)'), count(*), count(*) FILTER (WHERE slot_name IS NULL)
FROM page_components GROUP BY 1;
--  deployed 3219 | removed 56 | pending 32 | approved 26   — null_slot: 0 in every group
```

- **NULL `slot_name` — REFUTED.** Zero of 3,333 rows. The name match has no silent
  under-report, which was Fable's stated worry.
- **`removed` rows — 56 of 3,333 (1.7%).** The projection carries no `build_status`, so a
  removed row still matches and over-grades to `high`. Safe direction, small population —
  but it is why the spec key is `would_drop_present` and **not** `would_drop_deployed`:
  the check has not measured deployment and must not say it has.

### Three things the estate's own guards caught, all kept

1. **A constant is invisible to the coverage sensor.** I wrote `ItemType:
   sectionDriftItemType` and `verifier_coverage_test.go` demanded the file be declared in
   `computedItemTypeSites`. Taking that escape hatch would have made **both** of this
   file's item types invisible to CLASSIFICATION — a hole in the guard exactly where it
   claims coverage. So the field sites spell their strings and a pin test stops them
   drifting from the constants.
2. **My own comment poisoned that scanner.** The comment explaining the above quoted the
   pattern it described, and the guard duly reported that this file produces an item type
   called `"literal"`. The landmine "a source-scanning test makes your COMMENTS
   load-bearing", experienced rather than read. The comment now describes the pattern
   instead of spelling it, and carries a warning not to re-spell it.
3. **`pattern-check` caught a roll-claim with no commit** on my own register entries —
   `WII-039`/`WII-040` said "inert until the next roll" and named no sha. Fixed in
   `e48ecbe75`.

### The mutation table — and the one that SURVIVED

Fifteen mutations, each applied by script, the named test run, the source restored and
`diff`ed byte-identical. **Eleven on the check, four on the seam. Fourteen killed on the
first pass. One survived, and it is the most useful line in this file:**

> Deleting the receipt-insert error return left `TestReceiptFailureWithholdsTheRetraction`
> **green**. Not because the guard was redundant — because the flow fell through to the
> `!inserted` arm, whose presence `SELECT` is unmocked and errors too. **A guard in SERIES
> did the work.** The safety property held throughout; the test's *claim to pin that line*
> did not.

The wrong reading is "the mutation passed, so the line is fine". The test now asserts the
specific message (`could not be written`) and the mutation is killed. The survival is
recorded in the test file's own header, because the next person to run that table needs to
know it happened once.

### What I did NOT build, and why each

- **No `pages.sections` history table** — refuted for one `count(*)`, above.
- **No writer to `site_plan_sections`** — RFC_064, the `427` lane's, open with the owner.
  Coordinated directly with that session, which narrowed its RFC to reference this work
  rather than duplicate it.
- **No refresh of the frozen spec**, although `refreshOnConflict` exists on the write
  seam. Those frozen lists **are** the baseline `lost` is computed from; refreshing them
  would erase the only surviving record of the destroyed composition. The dedup blindness
  is fixed by FREEING the slot, not by updating the row.
- **No changes to the other nine flag-only checks with no closer.** The seam is reusable;
  each predicate is its own, and a shallow arm on nine checks is how a naive closer gets
  written nine times.
- **An exported predicate for RFC_064.** The `427` lane asked for
  `SectionSourceStateForPage` exported behind a querier interface. I did the interface
  (both `*sql.DB` and `*sql.Tx` satisfy it, which is what lets their writer re-run the
  drift predicate inside its own transaction) but kept everything **unexported**: an
  exported helper with no caller reads as a finished refactor. Their commit is what
  exports it, at which point it has one. They agreed, unprompted, that this was the better
  call.

### Test state at commit

`go build ./...` clean; `./platform/orchestration/...` green **except two pre-existing
failures that are not mine** — `TestFindingCodeScanEveryWriteIsRegistered` and
`TestTemplateExecutorsAreDeclared`, both about `FAIL_WORK_ITEM_MESSAGE_TEMPLATE_FALLBACK`
/ `renderFailWorkItemMessage` from commit `83407cd37` (the 440 lane). **Verified they
pre-date my change** by running `scripts/verify-head-builds.sh --test` against committed
HEAD `c68932577` before my work was committed: identical two failures. `verify-head-builds.sh`
on my own HEAD `e48ecbe75`: **OK — HEAD builds.**

---

## 2026-09-03 — council round 1: APPROVED, and acting on it found a defect no test of mine could

`009fabca-71c8-4f7b-9b23-f0b6605eb531` — **APPROVED**, 13 reviewers, 4 advisory objections,
none high, 4 abstained. Committed with `Council-Reviewed:` on `2cca1b085`.

### The defect I found while answering the objections

**Not one of the objections — it came from asking a blast-radius question none of them
asked.** *What does the closer do with items that are ALREADY closed?*

`loadOpenDriftItems` carries **no status vocabulary**, following
`findResolvedEmptySections`, which argues the point explicitly: this package already holds
two hand-rolled copies of the closed-status list and they disagree, so *"resolveWorkItems
owns the predicate; this function owns the observation. Reading a few already-closed rows
costs a **no-op UPDATE and nothing else**."*

**That last clause is true for a retraction with no side effect and false for mine.** My
retraction writes a receipt, and `resolveWorkItems` writes it BEFORE the UPDATE.

```sql
-- what would actually have happened on the first pass after the roll
 idea.uk         | guides-index    | complete | {guide-list}
 robot-hands.com | gripper-catalog | complete | {gripper-spec-sheet}
```

Both already settled — `guides-index` **adjudicated a benign rename by its owning lane**,
`gripper-catalog` documented in 469 and routed. Both would have been re-raised as fresh
`needs_human_review` items while the retraction they belong to matched no row: **the receipt
without its retraction.**

Every test passed. All eleven mutations passed. It is not a logic error — it is a
**precedent whose licence had expired for my case**, and the precedent's own text told me so
in a clause I read as reassurance rather than as a condition. Fixed with an open-item
pre-check gating the whole receipt block (and a pre-check ERROR is not read as "nothing to
close"); both arms mutation-proved. `WRONG_CALLS` entry: **when you cite a precedent, quote
its stated COST and test that cost against your change.**

### The four objections, and what each cost

| seat | sev | outcome |
|---|---|---|
| `reuse_agent` — `Evidence` merges free-form into `result`'s top level, beside two live conventions | med | **CODE CHANGED.** Now nests at `result->'resolution_evidence'`. The seat proposed `revalidation` / `_verification`; **neither fits** and saying why mattered more than picking one — `_verification` = a COMPLETION gate ran, `revalidation` = the revalidator re-asked, this is what a RETRACTION observed. `retraction` unavailable (`write_audit_findings_retraction` owns it: 531 rows, 258 also carrying `resolved_by`). |
| `guardian` — test the COMBINED Receipt+Evidence path's arg count/order; check the second caller | med ×2 | **CODE CHANGED.** Exact-args test pinning `$7` to the seventh argument (mutation: move it first → fails). `work_item_retraction.go` verified to set neither field, pinned by a source test so the day it starts to is not silent. |
| `prior_art_librarian` — two load-bearing claims asserted, not shown | med ×2 | **VERIFIED AT SOURCE.** Exactly **2** non-test callers of `resolveWorkItems`. `insertWorkItem`'s false return arises at `:2104`, `:2406`, `:2425` plus ON CONFLICT — **three of the four leave no usable record**, which is exactly why `!inserted` confirms. |
| `architecture` — `needs_rfc`: a shared contract framed as reusable for nine checks, decided inside a single-symptom fix | med | **`RFC_066` FILED**, quoting the objection. Its §5 gives the blast radius the seat asked for — including that the honest number is **smaller than nine**, because a receipt only means anything where a resolution can be DESTRUCTIVE, and nobody has enumerated which of the nine those are. |

`editquality` (med) wanted the receipt's CONTENT in the sketch, not just its absence being
refused — a submission-shape gap rather than a code defect; the function exists and is
tested. `bug_historian` (med) objected that the receipt lands in a queue the estate does not
drain — already in the stated risks, `bugs_open/033`, and not fixable here.

**Total: 18 mutations run across the change, each killing a named test, sources restored and
diffed byte-identical every time.**

---

## 2026-09-03 — CORRECTION: "nothing ever closes one of these" is FALSE, and my instrument could not have told me otherwise

Raised by the `427` lane after I read the two archived rows it pointed me at. **It is right,
and I had repeated the claim in the shipped code comment, a LANDMINES entry, this file's own
table, the SUMMARY and the owner log.** Corrected in all of them; the code comment matters
most, because that is the one a future reader inherits without asking.

### What is actually true

```sql
SELECT item_key, status, created_at::date, completed_at::date, handled_by
FROM site_work_items_archive WHERE item_type='section_source_drift';
-- section_source_drift:contact       complete 2026-07-16 2026-07-19  bugfix thread (bugs_open/002 C)
-- section_source_drift:who-we-help   complete 2026-07-17 2026-07-19  bugfix thread (bugs_open/002 C)
```

**Eight of this type have ever been filed** (6 live + 2 archived), and **two were closed by
hand** — both on 2026-07-19, both by ONE thread, two and three days after filing. (The `427`
lane said "two people, same-day"; it was one thread, 2–3 days later. Stating it because the
correction should be more accurate than the error, not merely different from it.)

### Why neither of us could see them, which is the transferable half

**A census over `site_work_items` returns "nothing ever closes these" whatever the truth
is**, because closing a row archives it out of the table being queried. The instrument
cannot produce the disconfirming answer. That is
`MEMORY[a-closer-census-cannot-see-what-it-succeeded-at]` — *"`site_work_items` is a ROLLING
WINDOW; closing a row archives it out of the table you queried"* — about this exact table,
already in my index while I wrote its negation.

**And the same error inflated a claim I made to the owner about the OTHER nine checks.** I
wrote "nine other detectors raise warnings that nothing ever closes". Over the UNION,
`[MEASURED 2026-09-03]`:

| item type | ever filed | ever closed |
|---|---|---|
| `claims_unverified` | 61 | **16** |
| `voice_tells` | 72 | **5** |
| `image_source_unsatisfiable` | 94 | **2** |
| `decision_blocked_change` | 14 | **2** |
| `capability_gap` | 334 | **1** |
| `content_duplication` | 1 | **1** |
| `contact_form_undeliverable` | 7 | 0 |
| `section_source_drift` | 8 | **8** |

**Six of eight have real closures.** The true property is **"rare, manual, undriven, and
invisible once it happens"**, not "impossible".

### What survives, and what this changes about the fix

**Nothing in the design changes, and the fix is better motivated by the true version.** The
failure was never that resolution is impossible — it is that it depends on an individual
being present at the right moment and leaves no visible trace afterwards. `capability_gap`
at **1 closure in 334** is a sharper statement of that than "zero" would have been, because
zero invites "so nobody can", while 1-in-334 says "somebody did, once, and you would never
find out".

### The cheap check that would have

**When a claim is "X has never happened", ask what your query does to X when it DOES
happen.** Here: closing archives the row, so the census is blind by construction. The
general form is the one I already had in memory and did not apply to my own sentence —
which is the uncomfortable part, and the reason this is in `WRONG_CALLS` rather than only
here.

---

## 2026-09-04 — Q2 ANSWERED: the build path reaches the page; the DEPLOY is refused by a shipped guard

§5 Q2 of the handoff (and line 236 of the bug file) read **"Unknown; not assumed either
way."** It is now measured. The answer is in two halves that point opposite ways, which is
why "does it reach it" was the wrong shape of question.

### Half 1 — the reconciler DOES reach an archived page. Nothing filters `pages.status`.

`[MEASURED 2026-09-04]`, read at the source:

- `loadPlanPages` (`reconcile_site_plan_action.go:640`) selects from `site_plan_pages`,
  which has no status column at all.
- `loadRealisedPages` (`:662`) is `SELECT name, COALESCE(build_status,''), … FROM pages
  WHERE site_id = $1` — **no `status` predicate**.
- The emit loop iterates **plan** page names, not `pages` rows.

And `gripper-catalog` is in the current plan `7a40a0f9-a1cd-4259-8654-cc0922e942aa` as
`role=content, nav_order=2` — a top-nav page, not a stray. So Q1's stamp withdrawal really
would flip `decideEmit` from `skip_built` to `stale` and emit a build item. Half 1 says the
repair starts.

### Half 2 — and then the deploy is REFUSED, by a guard that is correct and live.

`archived_page_guard.go` (commit `580af7ff0`, 2026-08-12, `bugs_open/266`) refuses at **two**
seams, deliberately: `git_commit` (`git_deployer_actions.go:65`) and the `update_page_status`
deploy stamp (`v3_site_actions.go:899`). Its predicate is a literal
`status == "archived"` (`pageIsArchivedForGuard`, `:164`). Its header states the reasoning:

> `archived` = "nothing may deploy this" -> no deploy is legitimate. A state whose
> prohibition has no exceptions must be guarded at the seam every producer passes through.

**It is in the running binary** — ancestry test, not a stamp grep (this lane's own §3 lesson):

```
git merge-base --is-ancestor 580af7ff0 239ab3626   # YES
git merge-base --is-ancestor 2fae8baa4 239ab3626   # NO  <- control behaves
kubectl -n ai-persona-system get deploy agent-chassis -o jsonpath=…  # v1.0.1360
# pods agent-chassis-ffc9ddff9-{jvw92,k866t} started 2026-09-03T22:06:58Z
```

**And it is not theoretical — it has already refused THIS page.** `[MEASURED 2026-09-04]`:

| | |
|---|---|
| `ARCHIVED_PAGE_DEPLOY_REFUSED` rows, all-time | **308** (first 2026-08-14, last 2026-09-04) |
| naming `page_id 64fab29e-5d8a-4a50-ad1b-2f9b0721cef6` | **3** (last 2026-08-23 13:02Z) |
| that page_id | `robot-hands.com` / `gripper-catalog`, `status='archived'` — joined, not assumed |
| demand control (`agent_error_log` can find rows) | 57,780 all-time, **243 since the roll** |
| fail-open window `ARCHIVED_PAGE_GUARD_UNCHECKED` | 1, on 2026-08-26 — so the guard is not silently standing down |

The 3 refusals are the load-bearing evidence: they prove the page **did** reach the commit
seam under some producer and **was** turned away there. `[INFERRED]` — that a post-Q1
rebuild would meet the same refusal — but inferred from a path this page has already walked
three times, against a predicate that is string equality on a column that has not changed.

### What this does to the three decisions: Q3 is PRIOR, not third

The handoff presents Q1/Q2/Q3 as co-equal and notes the third "can moot the other two". The
guard makes that structural rather than a matter of preference: **while `status='archived'`,
no repair can ever deploy, whatever Q1 rules.** So there is no branch in which we correct an
archived page's composition and a visitor sees the result. The fork is binary, and the
guard's own refusal message already words it:

> "Un-archive it deliberately if it should be live, or retract it (page-retraction) if its
> file is still served."

- **(a) Un-archive** (`status='active'`) — guard stands down, Q1 becomes the operative
  ruling, `760` applies and can actually render.
- **(b) Stay archived** — `760` is **moot**. The live defect is then the serving 200, and
  the remedy is retraction, which the guard explicitly does **not** block (it dispatches
  `delete_file`, a different path from `git_commit` — stated in the header and checked there
  rather than assumed).

`gripper-catalog` is already one of the nine untriaged `archived_page_still_serving` items
(`archived_page_still_serving:64fab29e-…`, filed 2026-08-26, still `detected`) — the same
page id, so §7.3's "the same page carries both flags" is now joined by id, not by name.
Re-probed today: `200 SERVING`, invented-URL control 404, sibling control 200.

### The misstep worth recording

My first instinct was to answer Q2 from the reconciler alone. Reading `loadRealisedPages`,
finding no status filter, and stopping there would have produced a **confident and useless
"yes, it reaches it"** — true, and it would have sent the next session to apply `760` after
a Q1 ruling, straight into a refusal 308 rows deep in the database. **"Does the path reach
it" is not "does the work complete".** What caught it was following the page to the *end* of
its path rather than the start: the sibling archived pages on the same site had not rebuilt
either, which is the anomaly that made me look downstream for a blocker.

### Addendum, same session — the inbound-link cost of each branch, and the zero that needed a control

Decision-relevant to Q3, because it prices the two branches. `[MEASURED 2026-09-04]`, over
`page_components.rendered_html` of `status='active'` pages on `robot-hands.com`:

| link target | live pages linking |
|---|---|
| `gripper-catalog.html` (**the archived page**) | **0** |
| `gripper-catalog/` (the ACTIVE index) | **19** |
| `gripper-selection-guide` (control: a live page) | 1 |
| `href=` at all (control: the query can see links) | 31 |

**The first zero is real, and I nearly reported it without the controls.** On its own "0 pages
link to it" is exactly what a blind query returns — wrong column, wrong link spelling, wrong
status filter all produce it. The `gripper-catalog/` row is what makes it informative: the
same query, same column, same predicate, finds **19**. So the query can see links to this
page's own name-space; the archived `.html` genuinely has none.

**What it means for Q3.** Retracting the archived page breaks **no internal link** — the
site's navigation already points at `gripper-catalog/`, i.e. `gripper-catalog-index`, which is
`active` and rebuilt 2026-09-03. So branch (b) is cheap in link terms.

⚠ **But it is not therefore the obvious answer**, and §7.3 already has the counterweight:
`gripper-catalog-index` carries a **single `news-listing`** and is *not* a replacement for the
archived page's content (hero + text + info-card-grid + CTA). So the measured state is **19
live links pointing at a thin index, while the substantive page sits archived and
unreferenced**. That is an argument the archive was a mistake as much as it is an argument to
finish retiring it — which is precisely why Q3 is the owner's and not mine.
