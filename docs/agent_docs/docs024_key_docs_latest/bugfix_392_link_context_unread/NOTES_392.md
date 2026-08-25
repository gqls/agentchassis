# NOTES — bugfix_392 (append-only, newest at the bottom)

## 2026-08-25 — session `bugs_open/387` opens the lane

- **Ownership before touching anything.** `who-owns.py 392` → no owning workstream; no
  `bugfix_392*` dir; `git log --since='45 minutes ago'` showed 104 fleet commits and **zero**
  touching 392/393/394 or `prepare_link_context`. The 358 lane that filed it has declared itself
  complete ("Nowhere, by design — this lane's work is complete"). Routed at
  `bugfix_092_writer_link_constraints`, quiet since 2026-07-31. ⇒ unowned, resumed.
- **No SUMMARY today, deliberately.** The five headings would say "we have just started", which
  is the cadence rule's own test for not writing one. First SUMMARY when the canary is proven.

### What the evidence says (all `[MEASURED 2026-08-25]`, queries in RUNBOOK)

- 2 `LINK_CONTEXT_UNAVAILABLE` rows ever. **Both already healed or never published** — so the
  bug's stated severity ("three pages… medium-high") rests on damage that no longer exists. This
  is a latent class defect. Said plainly in the bug file's correction block, not softened.
- 411/736 deployed pages carry zero writer-authored prose anchors; 187 on prose page types;
  **48 of those 48 owned pages are link-less**, which is a striking enough ratio to be its own
  finding and belongs to the 396/333 lanes, not here.
- `orchestration_id` on the log row resolves to the exact page — verified end to end on **both**
  rows. The bug file says only `site_id` is recorded, which is true of `context` and false of the
  row.
- `page-rerender` neither spawns the writer nor knows `edit_live`; `page-content-writer` is the
  only agent fleet-wide running `prepare_link_context`. **392's own fix candidate 1 is wrong.**
- `internal-linker` is LIVE (7 completions to 2026-08-24) while `check_orphan_pages.go:14` still
  says "(future handler)".

### MISSTEP (mine), caught in-session, and it is the reason the census exists

I first measured link-lessness on `page_components.rendered_html` and got **140 of 737**. That
instrument counts template nav, hero and CTA links, so a page whose writer emitted nothing still
reads 2–3. Measured on the writer's own output (`content_data`) the answer is **411 of 736** —
nearly triple, and it includes 31 blog-posts that the first instrument called perfect (0/164
there). I noticed only because the per-page-type split had no mechanism behind it: "every blog
post is fine" is not a thing a writer does. Logged in `WRONG_CALLS.md`.
**The check:** before trusting any "does this page contain X" census, open one known-positive and
one known-negative row and look at where X actually lives. Here, hero/CTA links live in
structured fields (`cta_url`, `link_url`) and carry no `href=` at all.

### Three claims of mine REFUTED by the third fable planner, before any code was written

I ran three planners: framework-first, reuse-first, and one asked only to stress the design. The
third earned its cost by refuting the other two.

1. **My canary induction was self-blinding.** I had settled on setting an unparseable `site_id`
   in the step config, having read `resolveLinkContextSiteID` and confirmed "an explicit config
   value wins". It does reach the degrade — but the row is then written with the garbage site
   context, and the reader arm selects by `site_id`. **The induction would have proved the
   writer and silently never exercised the reader**, while looking like a clean end-to-end pass.
   Same shape as the missing control that filed bug 387. Replaced with a site-scoped opt-in hook
   that shrinks the query's context deadline *after* site resolution.
2. **I had the owned-page door backwards.** I wrote that `writeWorkItem` would park items filed
   against owned pages — an argument for excluding them as tidiness. In fact neither
   `internal-linker` nor `page-build-handler` declares `refuse_owned_page`, so the door does not
   fire at all; the refusal comes late at `SavePageSectionsAction`'s `OWNED_PAGE_GUARD`, **after
   the LLM spend**, terminating `wont_fix`. Excluding owned pages at the query is therefore
   load-bearing, not hygiene.
3. **My item_key would have collided.** I proposed sharing `internal-linker`'s
   `internal_link:<page>` namespace and called the co-dedup a feature. `check_orphan_pages`
   already mints `needs_links:<name>:<site>`, and `idx_swi_dedup` is UNIQUE on
   `(site_id,item_key)` with **no item_type column** — so an inbound-orphan finding and an
   outbound-absence finding on one page would share a slot and one would vanish. They are
   different defects and both must stay actionable.

**The transferable lesson:** two planners asked to design agreed with each other and with me; the
one asked to attack found three defects in twenty minutes. Where a mechanism will be trusted to
run unattended, the adversarial seat is not a luxury round.

### Design, as it stands going into implementation

One new discovery check in `platform/orchestration/actions/discovery_checks/`, sibling to its own
inverse. No new service, no new image, no cron slot, no handler, **and no change to any live
agent** — the repair route (`content_rewrite` + `spec.mode='edit_live'` → `page-build-handler` →
`page-content-writer`) already exists and runs ~30/day, and because the repair re-runs the writer
it gets a fresh link allow-list and picks its own targets. Full design and the corrections block
in `PLAN_2026-08-25_392.md`.

⚠ `content_rewrite`'s 14-day health is 93 complete / 53 wont_fix / 45 failed / 21
needs_human_review — **~21% fail or are refused.** Any claim that a filed item equals a repaired
page is wrong, which is why acceptance is measured at the served page.
