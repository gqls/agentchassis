# RFC_064 — may a non-planner action replace the CURRENT plan's section rows for one named page, and what must it do to `built_from_plan_version`?

**Status: OPEN — owner decision requested.**
Raised 2026-09-03 by session "427" (`bugs_open/427` lane) after shipping migration `750`,
the fifth hand-written instance of the same three-store correction in seven weeks.

**The question is one sentence, and it is at the top so it does not get lost:**

> May a **non-planner** action replace the **current** plan version's `site_plan_sections`
> rows for **one named page**, and if so what must it do to that page's
> `built_from_plan_version`?

---

## 1. What the thing IS, in plain terms

A page's list of sections — its composition — is stored in three places at once:

1. **`site_plan_sections`**, rows belonging to the site's current plan. This one wins.
2. the **`site_specs.site_plan` aspect**, an older generation's store. Used if (1) has
   nothing for the page.
3. **`pages.sections`**, a JSON array on the page row. A *cache*, used if neither of the
   others has the page.

Whenever a page is built, the build reads the highest of those three that knows about the
page, and **copies it down over `pages.sections`**. So the cache is not an independent
opinion; it is a photocopy, refreshed on every build.

## 2. The rule that follows, and why it keeps being broken

**If you correct a page's composition by editing only `pages.sections`, the next build
undoes you.** Not the next re-plan — the next *build*. Nothing errors; the page simply
comes back with its old sections.

This is written down. It is in `LANDMINES.md`. It is stated in the concept register as
**PLAN-029** with the status `deployed`. It is the header of migration `154`. And it keeps
happening anyway:

| when | who | what happened |
|---|---|---|
| 2026-07-15 | migration `153` | wrote the cache + the aspect, not the table. The rebuild resurrected the deleted components. `154` fixed it by hand. |
| 2026-08-22 | `bugfix_357` | designed the two-leg write itself, from scratch, for one site (`701_HOLD`) |
| 2026-09-02 | migrations `719`/`727`/`728` (`bugs_open/427`) | wrote only the cache. Caught before a build fired. |
| 2026-09-03 | migration `750` (this lane) | the fix — written by hand again |
| 2026-09-03 | migration `754` (`apis.uk` lane) | same fix, same day, different lane, written by hand again |

And twice it has *completed unremarked* (`bugs_open/469`): `robot-hands.com/gripper-catalog`
lost `gripper-spec-sheet` — **the very component `154` was written to rescue** — and
`idea.uk/guides-index` lost `guide-list`.

**The knowledge is not the missing part. The door is.** There is no typed way to do this, so
every lane hand-writes SQL, and roughly half get it wrong. "Keep one writer" has meant, in
practice, *six writers, all in SQL, none reviewed against the loader.*

## 3. Why this needs a ruling rather than a commit

`reconcile_site_plan_action.go:596-601` says, in its own words, that `site_plan_sections`
*"is per-plan and immutable, so it can"* tell an unchanged page from a re-composed one. Its
`decideEmit` restamp logic is **licensed by that immutability**. In Go the table is
insert-only today: two `INSERT`s, zero `UPDATE`s, zero `DELETE`s. And `bugs_closed/001` says
in terms: *"Do not 'fix' this by … hand-writing `site_plan_sections` — the writer is
`write_site_plan_action.go` … and it must stay the [only one]."*

Adding a writer narrows a guarantee another action's correctness rests on. Under the owner
ruling of **2026-07-29 §1** — an addition needs an RFC when it changes what the shared
mechanism *guarantees* — that is squarely architecture-scope. The **2026-08-11 RFC_022
narrowing** does **not** exempt it: that narrowing is about an opt-in *field on an existing
shared action*, and this is a new writer to a table whose immutability is load-bearing
elsewhere.

## 4. How the current case measures against that rule

The immutability that `decideEmit` actually needs is narrower than "nobody may ever write":

- It compares the **current** plan's rows against the rows of the plan a page was
  `built_from_plan_version`. The danger is mutating a **superseded** plan's rows: that
  falsifies build history for every page stamped to it, and a false `restamp` marks a page as
  built without rebuilding it, unrecoverably (the only record of the old composition is the
  row you overwrote).
- Mutating the **current** plan's rows falsifies no comparison. `decideEmit` returns
  `skip_built` on `BuiltFromPlanVersion == planID` **before** any section comparison
  (`:612-614`), and rows are keyed `(plan_id, page_name)` so no other page moves.

**But "falsifies no comparison" is not the same as "harmless", and this is the sharp edge.**
If a deployed page is stamped at the current plan and you change that plan's rows for it, the
stamp becomes a **false statement** — the page was built from the *old* rows — and
`skip_built` means the reconciler will never rebuild it, so the correction only takes effect
if something else happens to build the page. Hence the second half of the question.

### 4a. A second, stronger case — where q2 is BLOCKING, not theoretical (contributed by the `bugs_open/469` lane, 2026-09-03)

Migration `750`, this RFC's motivating case, was the *easy* shape: a rename at a fixed
`ordering`, on a page that was already correct, on a site with exactly one plan. It could be
argued as a precedent-following one-off. **`robot-hands.com/gripper-catalog` cannot, and it is
blocked on question 2 specifically.**

The loss is **damage, not an intended removal**, with provenance: migration
`SQL_2026-07-24_r9_gripper_catalog_real_grid.sql` placed `gripper-spec-sheet` at position 3 as
an owner-backed call — spec cards being the honest fit where `product-grid` had empty
e-commerce fields. A reasoned July addition, wiped by the tier-1 sync-down, never restored.

Why `750`'s safety argument does not transfer — `[MEASURED 2026-09-03]`, DB facts re-verified
independently here:

| | migration 750 | gripper-catalog |
|---|---|---|
| shape of the write | RENAME at fixed `ordering` | **INSERT at ordering 2**, shifting `info-card-grid`/`call-to-action` down — the renumbering §5 warns about |
| plans on the site | exactly **1** | **5** — so "nothing superseded to falsify" is unavailable |
| `built_from_plan_version` | = current plan | = current plan |
| does the correction render? | yes — page already correct, nothing to re-render | **NO.** `decideEmit` returns `skip_built` before any section comparison, so a corrected plan is **a no-op** |

**That last row is question 2, made load-bearing.** On this page the repair does not merely
lose provenance if the stamp is left alone — it **does not happen at all**. Without a stamp
withdrawal the corrected plan rows sit there and the page never rebuilds from them. This is
the case that turns q2 from a tidiness argument into a blocking one.

The renumbering hazard was **checked rather than carried across**: all four plan rows have
`assigned_fact_ids = NULL` and `subject = NULL`, and the site's only imagery row for this page
is page-scoped (`'gripper-catalog'`), not `'<page>:<ordinal>'` — so on *this* page the shift is
benign. That is the right discipline and it is worth naming: §5's warning is a reason to
enumerate the four consumers, not a reason to refuse.

**One further wrinkle, which is a question this RFC does not answer:** `gripper-catalog` is
`pages.status = 'archived'` and `build_status = 'deployed'` with a NULL `last_built_at`
(verified here), while the 469 lane reports it **serving 200** at `/gripper-catalog.html`
(their probe, via `scripts/probe-page-url.sh` with an invented-URL 404 control and a sibling
200 control; my own `WebFetch` returned 403, a transport-level block that is evidence of
neither). If that holds, then even with the stamp withdrawn it is an open question whether the
build path will touch an archived page at all — so a ruling for q2 may still leave this
specific repair needing a separate answer. `gripper-catalog-index` is not a substitute; it is a
single `news-listing`.

The 469 lane is writing it as `_HOLD.sql` and routing the go/no-go to the owner alongside this
RFC rather than applying it. That is the right call: it is the same decision, met from the
other end.

## 5. The proposal

A single typed action, `apply_page_composition`, plus one file
(`site_plan_section_rows.go`) holding **every** mutating statement against the table. Net
writer count goes *down*, because the six hand-written SQL instances stop.

Structural guarantees, not conventions:
- **`plan_id` is never an input.** The only lookup is `WHERE is_current … FOR UPDATE`, so a
  superseded plan is *unreachable*, and a concurrent re-plan either blocks until we commit or
  makes us refuse.
- **Two modes.** `align_to_live` derives the list from the page's own live `page_components`
  (invents nothing) and leaves the stamp alone — the plan becomes *more* true, so `skip_built`
  is correct and no content is regenerated. `apply_list` takes an explicit list with an
  `expected_sections` compare-and-swap, and **must** set
  `built_from_plan_version = NULL, build_status = 'needs_rebuild'` — restoring honesty and
  making `decideEmit` return `not_built` so the correction actually lands.
- **Rename in place; never renumber.** `ordering` is a positional join key for four
  consumers: `assigned_fact_ids` (where `'[]'` and `NULL` are different instructions),
  `subject`, `page_components.position`, and `site_plan_imagery.scope_ref`, which for section
  scope is literally `'<page>:<ordinal>'` (`[MEASURED 2026-09-03]`: `index:1`, `index:2`,
  `about:2`). A renumbering correction silently re-points every section figure on the page.
> **NARROWED 2026-09-03 (later) — the DETECTOR half of this RFC is no longer mine to propose.**
> A session working `bugs_open/469` has taken `check_section_source_drift.go`: a
> direction-aware `Resolved` arm plus a `section_composition_lost` receipt so an
> `authority_won` resolution can never close silently. That overlaps what §5 originally
> proposed for that file, so **this RFC no longer asks for changes to the check** — it
> depends on them. Read their work as the detector half; this RFC is only about the WRITE
> path (`apply_page_composition`, `site_plan_section_rows.go`, and whether the current
> plan's rows may be replaced at all).
>
> One dependency remains and is stated rather than assumed: the postcondition below needs
> the per-page drift predicate callable with a **querier interface** (so it can run against
> an open `*sql.Tx`), not `*sql.DB`. That has been asked of the 469 lane; if it lands
> differently, this RFC's postcondition needs its own extraction and that is a change to
> their file, to be coordinated rather than raced.

- **A postcondition inside the transaction** re-runs the drift predicate, so the action
  cannot commit a state the build would resolve differently.
- `request_rebuild` defaults **OFF** (owner ruling 2026-08-02 §2: the unsafe branch — the one
  that regenerates copy — must be named by the caller).

## 6. The alternative, costed, and why it is not recommended

**Mint a new plan version for every correction.** Preserves immutability perfectly. Against
it: it puts *every* deployed page on the reconciler's restamp path, and any page whose
built-from version lacks rows falls through to `stale` and is rebuilt **with its content
regenerated** (`bugs_closed/038` exists to prevent exactly that). It also means copying four
tables, including `site_plan_imagery`, where a lossy copy files real image-generation
**spend**. A one-page correction becomes a site-wide copy event, for a population of two
pages.

**The action's interface is identical under either answer**, so a ruling for the fork changes
one function, not the design. That is deliberate: it keeps this a genuine choice rather than
a fait accompli.

## 7. What the owner is actually being asked

1. **May a non-planner action write the current plan's section rows for one page?**
   Recommended: **yes**, with `plan_id` structurally unreachable as an input.
2. **If yes, must `apply_list` withdraw the build stamp?** Recommended: **yes** —
   `built_from_plan_version = NULL`, `build_status = 'needs_rebuild'`. The cost is one page's
   build provenance, on a page whose composition we have just declared superseded.
3. **Or should a correction mint a new plan version instead?** Recommended: **no**, on the
   cost in §6 — but this is the purist answer and it is a legitimate call.

## 8. Relations

- `bugs_open/427` §19–§21 — the case, and the corrected mechanism
- `bugs_open/469` — the two losses that have already completed
- `bugs_open/443` — the adjacent class: pages born with a layout in the cache and nothing in
  the authority
- **`RFC_063`** — DECIDED (option B). Its *composition half* is unassigned and needs exactly
  the row writer proposed here, so the two should be sequenced, not duplicated. Distinct
  question though: 063 creates a **first** plan for plan-**less** sites; this corrects **one
  page** on a site that already has one.
- Concept register **PLAN-029** (`deployed`), **PLAN-040** (`aspirational` — proposes the
  drift auditor that now exists as `check_section_source_drift`), **PLAN-052** (the sync guard)
- `LANDMINES.md` — "`pages.sections` is a materialised CACHE", and the two entries added
  2026-09-03
