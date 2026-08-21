# NOTES — agritec.uk

Append-only, newest at the bottom. Technical log: what was tried, what the system actually said,
and every misstep. The missteps are not an appendix — they are the point.

---

## 2026-08-21 — session 1, scoping and measurement

### What the owner asked for

Rebuild `agritec.uk` inside the framework. An adoption, but rebuilt from scratch so the tools and
guides are deconstructed and fully managed. Full mission, spec and plan. More imagery and
infographics. News, editorial and directory listings later. Same hosting, replacing what is there.
Two clarifications arrived mid-session: **everything eventually migrates**, and the copy is
**written afresh, to the same level of detail and greater**.

### MISSTEP 1 — I took the repo copy as the inventory. It is stale.

`/home/ant/projects/domains/agritec.uk/01/` has **6 tools and no `/deepdives/` directory**. The
live site has **13 tools and a seven-part deep-dive series**. I only found this because I fetched
`https://agritec.uk/` to check hosting, and the returned HTML had a nav link (`/deepdives/`) that
does not exist anywhere in the repo.

Had I not fetched it, the plan would have silently scoped out **14 of 26 pages** — more than half
the site — while appearing complete, and the subject ledger would have certified that as
"everything".

**The check:** for any adoption, inventory from the **live site**, not from a local copy. A repo
directory is a snapshot of when someone last synced it; here that was seven months and one whole
content section ago. Cheap: one `curl` of the homepage and a diff of the nav links against the
local tree.

### MISSTEP 2 — my first crawl double-counted pages.

The recursive link-follow emitted `/deepdives/../tools/lens-calculator.html` alongside
`/tools/lens-calculator.html`. Both fetch 200, so both looked like real, distinct pages, and the
first inventory over-counted. Normalise paths before counting. Recorded in `RUNBOOK` §5.

### MISSTEP 3 — I nearly read a nav census as a fleet-wide orphan rate.

I ran a census of pages with no `site_nav_items` row and got, e.g., finetuning.uk 21 of 50,
robot-hands.com 27 of 45. Read naively that says half the fleet's pages are orphaned.

**It does not.** `site_nav_items` is the *top nav*. A page absent from it can be perfectly well
linked from a section-index hub. The measurement answers "is it in the nav bar", not "is it
reachable", and those are different questions.

**The check:** reachability is measured at the rendered page — count inbound hrefs — not at the
nav table. That is what `RUNBOOK` §4 does, and it is what produced the one finding that is real
(below).

### Column names I got wrong, cheaply (recorded so the next session does not repeat them)

- `site_specs.spec_type` → the column is **`aspect`**.
- `site_nav_items.nav_group_id` → **`group_id`**.
- `site_nav_groups.label` → does not exist; the label lives on the item.
- `site_publish_checks` is **not** a check registry — it has two columns, `site_id` and
  `last_checked_at`. The checker vocabulary lives in `site_work_items.item_type`.

### What is actually true, measured

Confirmed the owner's recollection that not all articles are linked from the index, and it is
worse than remembered:

- `/guides/vapor-pressure-deficit.html` — **0 inbound links from any of the four index pages**.
  Reachable only from the VPD calculator.
- `elms-calculator`, `insect-waste-converter`, `seaweed-carbon-est` — on the home page, absent
  from `/tools/`. The tools index lists 10 of 13.
- `ai-labor-calc` — the inverse: on `/tools/`, absent from home.

Two further findings the owner had not flagged:

- **The market ticker was fabricated.** `data-collector/v1/cmd/updater/main.go` is a `rand`
  simulator that stamps its own output `"Source": "Simulated Exchange / National Grid"` and writes
  a `market-ticker.json` the site never reads. The homepage hardcoded the values instead. Someone
  has already commented the ticker out on the live site.
- **The data layer is dead.** `fetch(` count across all six agri calculators is **0**. The six
  `/data/*.json` files are read by nothing; every number is hardcoded per tool. The files are
  still publicly served — `/data/crop-dli-table.json` returns 200.

That last point matters for the cannabis ruling (owner, same day): the cannabis rows are in a
publicly-reachable JSON that no page fetches. **A client-side absence is not an absence** — a
server-side reader uses the same public URL. `b2 sync --delete` at cutover is what actually
retires it, not the fact that nothing links to it.

### The measurement that decided the build shape

The owner asked for "the same level of detail and greater", so I measured the existing depth
rather than guessing at it: explainers 315–453 words, deep dives 434–605, and exactly **one**
diagram across all six explainers.

Then I measured what the framework itself produces, because "greater" needs a destination:

| page_type | Sites | Pages | Avg words |
|---|---|---|---|
| `blog-post` | 4 | 40 | 1,568–1,876 |
| `guide` | 1 | 5 | 511 |

**`page_type` alone decides whether an explainer lands at ~1,600 words or ~500.** The old agritec
explainers sit at the `guide` shape. Building them as `page_type='guide'` — the obvious choice,
since the site calls them guides and the URL is `/guides/` — would have reproduced the exact
depth we were asked to exceed, while looking like the natural mapping.

`[MEASURED, sample noted]` The `guide` figure is one site, n=5. Strong signal, not a law.
Re-measure before leaning on it again.

### What has NOT been done yet

Nothing has been dispatched at the cluster. No `sites` row, no specs, no submission. The next
session's first action is the seed SQL (PLAN §4 Phase 1), and **the seed must land before the
first page is written** or the claims layer silently no-ops for the whole build.
