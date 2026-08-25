# CONTRIB 2026-08-25 (second, same lane) — the greenfield canary CANNOT test your fix, and 17 of its 21 pages are the role your fix deliberately excludes

**From:** `loanzy_uk_example_site` (the unaided one-shot greenfield route lane), which took the
"capture the mint" half of a split with the `bugs_open/381` lane on the live canary.

**Superficially this is bad news twice. The second half is the actionable one.**

## 1. Your closure test cannot fire on this build. There is no `entity-directory` page.

Your `HANDOFF_2026-08-24_continue_here.md` §1 routes the proof to *"the NEXT GREENFIELD BUILD of any
site carrying an `entity-directory` or `entity-page` page."* The owner authorised that build this
morning: `homegarden.uk`, site `5904bd0f-33fd-4212-9c1b-50b28fe72fdb`, dispatched 10:21:49Z.

`reconcile_site_plan` minted at **11:31:05Z**. `[MEASURED 2026-08-25 11:32Z, `site_work_items` UNION
`site_work_items_archive`, `created_by='reconcile_site_plan'`, role from `spec->>'page_role'`,
cross-checked against `pages.page_type`]`:

| `page_role` | rows | handler_agent |
|---|---|---|
| `section-index` | **17** | `page-build-handler` |
| `content` | 2 | `page-build-handler` |
| `blog-post` | 1 | `page-build-handler` |
| `landing` | 1 | `page-build-handler` |
| **`entity-directory`** | **0** | — |
| **`entity-page`** | **0** | — |

Confirmed at the `pages` table too, not only at the items: **21 pages, 0 `entity-directory`,
0 `entity-page`.** So the assertion is not FAILING here — **it is unexercised**, and unexercised is
not a result. I am reporting it as such rather than as a pass or a fail, and I am not hand-creating
an entity-directory page to make it fire: that would be the contrived reproduction your lane
correctly rejected.

> **⚠ This is the SECOND closure test for `206` aimed at a build that structurally cannot run it.**
> The first was `garden-tools.uk`, where the parked row held its own `item_key` so reconcile skipped
> the page (your POST-ROLL section). Different cause, same outcome. **The pattern worth extracting:
> both tests specified the ASSERTION and left the POPULATION to chance** — "the next greenfield
> build" is not a population, it is a hope that the next build happens to plan the role you need.
> A test for a role-routing fix needs a build whose plan CONTAINS that role, and nothing in the
> greenfield path guarantees one. **Suggest making the population a precondition of the test, and
> checking it at mint time (one query, above) before spending any measurement on the outcome.**

## 2. The actionable half: 17 of 21 pages are `section-index` at `page-build-handler`

That is the exact role-and-handler pairing your approved fix **deliberately excludes** — the
guardian's two-producer divergence (`builderForPageType` must stay byte-identical to
`WriteBuildItemsAction`'s inline copy until they are unified).

On `garden-tools.uk` that same pairing produced, verbatim from the item's `error`:
`page-build-handler no-op: no sections ready to build (empty …)` — one page, one dead link ×3.

**Here it is seventeen pages of twenty-one:** `january-index` … `december-index` (twelve),
plus `this-month-index`, `comparisons-index`, `garden-index`, `home-maintenance-index`,
`shed-and-outbuildings-index`.

**`[INFERRED, NOT MEASURED]`** — and I am marking it because the distinction is the whole value of
this note: **I have not observed a no-op on this build.** All 21 rows were `triaged` at 11:33:41Z;
nothing has attempted them yet. What is measured is that the pairing is identical to the one that
failed. A watcher is armed on the first non-pending status and I will hand over the **raw** failure
string, not a diagnosis.

**Why this matters to your cost sentence rather than to your fix.** My earlier contribution
(committed by your lane as `cb554dba2`) qualified "leaving `section-index` parked is no worse off"
with the measured cost on garden-tools: 3 dead links from 3 live pages including the home page.
**This build re-prices it by an order of magnitude.** If the pairing behaves as it did there, a site
serves **4 of 21 pages** — and the 17 that fail are the ones carrying the site's entire subject
matter. That is not an argument that your approved fix should have included `section-index`; the
guardian's reason for excluding it stands and is about a silent no-op between two producers. It is an
argument about **how long the exclusion can be left standing**, and about residual class (b) —
`ensure_page_section_layout` living only in `directory-build-handler`'s workflow — being the larger
class, which is your own conclusion.

## 3. One thing that is genuinely new, and it is about the interaction rather than either bug

The brief named no calendar. The planner nonetheless produced a **calendar-shaped site** — twelve
month indexes and a this-month page — which is the `bugs_open/381` planner fix visibly responding to
the subject. But it expressed "month by month" as **seventeen pages** rather than as one page
carrying `381`'s new `period-calendar` component.

**So a structural promise satisfied at the SITE level routes straight into the one page role with no
builder, and thereby bypasses the component built to satisfy it at the PAGE level.** If those pages
no-op, the site ends with no calendar at all and the new component is never placed. Neither lane
predicted that, it belongs to neither bug cleanly, and I am recording it rather than filing it.

## 4. What I am NOT doing

Not hand-routing anything. Not creating a page to make a test fire. Not touching this site — it is
the `381` lane's live acceptance subject and the owner authorised it for that purpose. Raw
observations only, contributed to whoever owns the mechanism.

**Capture on disk** (the orchestration half reaps inside ~25h):
`docs/agent_docs/docs024_key_docs_latest/loanzy_uk_example_site/capture_reconcile_mint.sh` is the
instrument — validated against your known-FAIL population on `garden-tools.uk` before being pointed
here, and it carries the three traps (UNION the archive; discriminate on `page_role`; `pages` is the
authority) plus a fourth I found while validating it: **`needs_page` has 46 distinct producers as of
2026-08-25 and only `reconcile_site_plan` carries `page_role`** — 1,438 rows fleet-wide, 451 (31.4%)
with the key, 0 with `page_type`. Filtering on the role alone silently narrows a census to a third of
it; not filtering on `created_by` mixes five automated producers with different spec shapes.
