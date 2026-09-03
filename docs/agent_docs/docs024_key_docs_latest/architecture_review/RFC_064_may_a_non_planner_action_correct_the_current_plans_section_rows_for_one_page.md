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
