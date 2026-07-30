# HANDOFF — dartsonline_traffic, 2026-07-30: start here

Written at the end of session 2 (2026-07-29, ~13:00–19:50Z), for a fresh thread.

**One-line state:** the site is honest, readable, navigable, and has 8 buying
guides + a live news page + 15 new images; **the imagery cause is now fixed at
every level — code approved and live on v1.0.1207, config applied 2026-07-30.
Nothing is owed.**

---

## 0. THE IMAGERY CAUSE IS CLOSED — done 2026-07-30, before this handoff was filed

This section was originally "the first thing to do". The chassis rolled, cluster
access came back, and it was finished. Recorded because the *evidence* is the
reusable part.

**Pod-verified on v1.0.1207, both replicas** (`agent-chassis-6c448c66d6-fjpd7`
and `-kmm9b`), with three markers rather than one:

```
"no other background colour"             -> 1, 1   the new clause: ADDED
"composedPaletteDirection: query failed" -> 0, 0   first draft's string: DELETED
"composed palette unavailable"           -> 1, 1   its replacement: ADDED
```

The middle one is load-bearing. A positive control alone proves only that *some*
build after `bd9ebfec6` shipped; **the delete-marker proves the binary also
carries `88dee2a8d`**, the council follow-up, because that commit is what replaced
the string. A marker your change DELETED is the strongest deploy evidence there
is — and the timeline (both rolls postdating both commits) stayed [INFERRED]
throughout, because a retag is not a rebuild.

**Then the config half applied** —
`SQL_2026-07-30t_planner_drop_hardcoded_icon_ground.sql`. `build-site-planner` no
longer pins every icon prompt to `#EEEEEE`/`#4A4A4A`; every flatness guard is kept
verbatim, and the prompt now explicitly refuses to name a colour because the
palette supplies one at generation time. Dry run, then commit, then re-read on a
**fresh connection** — the in-file verify runs inside the same transaction as the
UPDATE and can only agree with itself.

**Two gotchas found doing it, both in the file:**

- **A `ROLLBACK` at the foot of a script is only safe if the `BEGIN` at the head
  is real.** As first written, `BEGIN;` was commented and `ROLLBACK;` was live —
  every statement would have autocommitted and the trailing rollback would have
  been a no-op with a warning. The "safe default" would have committed while
  appearing not to.
- **`snapshot_agent` has two overloads with two destinations.**
  `snapshot_agent(type, reason)` writes to **`agent_definitions_backup`**;
  `snapshot_agent(type)` inserts an `is_snapshot` row into `agent_definitions`. I
  checked the latter, found zero rows fleet-wide, and was one step from reporting
  a silent no-op. Check `agent_definitions_backup`, and assert
  `backup_has_old_hex = t` — a snapshot carrying the POST-change config restores
  nothing.

**Expect no visible effect.** It changes what FUTURE plan rows say. The 92
existing rows carrying the literal are deliberately untouched — rewriting them
would change 9 sites' plans at once, 4 of them light sites where the literal is
harmless. Absence of a visible change is not evidence it failed.

---

## 2. What the site is now — **re-measured live 2026-07-30 11:30Z**, not carried forward

| | state |
|---|---|
| **Guides** | 8 live, each with a hero photograph serving 200. The 9th, `grip-styles`, belongs to the `gemini_content_provider` lane — **do not build it** |
| **News** | `/news/index.html` serving 200 and well populated |
| **News feed** | **all 3 RSS sources have now fetched, `error_count` 0** (they were unfetched at the last handoff, waiting on the 6-hourly tick). **Relevant items 14 → 52 overnight** — the lane is genuinely flowing, not just armed |
| **Nav** | Guides · News · Start Here · Deals + About/Contact/Shipping. Zero dead links on any served page (07-29) |
| **Honesty** | wide sweep (49 phrases, 11 pages) → 3 hits, all the DARTS sense of "checkout" (07-29) |
| **Internal links** | Guides hub and homepage each serve 8 links to the 8 guides (07-29) |
| **Contrast** | guide pages `contrast=0`; homepage 2; news page 28 (all `.news-list-tag`, marginal) (07-29) |
| **Imagery** | 15 generated and **all 15 looked at**; 8 heroes wired and serving; **7 icons still on no page** |
| **Meta descriptions** | 16 of 22 active pages. The 6 without are deliberate skips (unbuilt / retail-hub pages) |

Rows marked (07-29) are yesterday's and were not re-fetched this morning; everything
else is this morning's.

**Site id:** `5fe8785b-223d-41a3-88ee-c07187622381`
**Current plan id:** `0fb05b75-04f4-4f4c-8890-c34d6a71012c`

---

## 3. Next work, in the order I would do it

1. **Homepage icons — and it is TWO steps, not one.** The homepage card grid still
   renders 4 emoji (🤝 ✈ 🎯 📰).
   - **(a) The icons are generated but not DEPLOYED.** Measured 07-30: 19
     `undeployed_asset` items sit at `detected` — 17 icons (the 7 new ones plus 10
     orphans from an older plan), the favicon and the og_card. The 8 heroes went
     out because their `needs_imagery` items were routed; these were not. Promote
     the 7 new icons **per site, as data** — do not touch the promoter
     (`bugs_open/083`, see §5).
   - **(b) Then swap the component.** From `info-card-grid` (schema mandates
     emoji; global, 23 placements on 11 sites — **do not edit it**) to the
     existing global `image-hover-card-grid`, typed for `cards[].image` and
     already reading `surface_alt` for its tile. Update
     `page_components.component_id` + `content_data`, then a `needs_page` /
     `page_rerender` with `reason:"image_landed"` — plain reassembly will not pick
     up the swap.

   Doing (b) without (a) wires the page to images that 404.
2. **Discovery files** — `scripts/site-discovery-files.py dartsonline.com --write`
   → robots.txt, sitemap.xml, llms.txt. Currently sitemap and llms.txt 404. **Run
   it when the site is NOT mid-build**: it probes every URL and drops 404s, and a
   page being rebuilt at that moment is dropped as "broken".
3. **The setup-builder tool.** Planned since July, never made; the hold was lifted
   2026-07-24. It is also the first tool the ratio asks for.
4. **Editorial cadence** — `blog-content-planner` made news-aware, plus a weekly
   `blog-analysis-refresh` scheduled task. This is what keeps the site producing
   without a thread driving it. `growth_config` is already in place.
5. **Affiliate resolver + feed ingester**, built dark. Largest piece; PLAN phase 6
   has the design. Would confirm with the owner before starting.

---

## 4. Landmines specific to THIS lane

- **`pages.sections` is a materialised cache; `site_plan_sections` is what the
  build reads.** I edited the cache, the rebuild read the source, and the change
  did nothing. Same for `pages.title` vs `site_plan_pages.title`.
- **`page_rerender` reassembles and PRESERVES content; `needs_page` regenerates
  it.** Right tool for a head-only fix, wrong tool for a stale listing. And a
  hand-made `page_rerender` needs `page_id` in the spec **and** in the `page_id`
  column — mine failed 3/3 with `page_id not found in input` because I copied the
  `needs_page` shape. **Copy the shape from a COMPLETED row, not from the action's
  source.**
- **A stale listing is invisible in every status.** `build_status='deployed'`,
  `deployed_at` today, item `complete`, page serves 200 — and the listing was
  written 9 days earlier. Only `page_components.updated_at` disagrees.
- **`sites.github_repo` is EMPTY for this site** and falls back to the default
  `"sites"` repo, which is correct here (B2-hosted). Do not "fix" it.
- **Archiving a page does not undeploy it** (`bugs_open/098`). This is why `sale`
  and `new-arrivals` were repurposed rather than deleted — an archived
  `/sale.html` would go on serving "we cut prices" indefinitely.
- **`site_plan_imagery.source`** takes `llm|classifier|manual|adoption`;
  `site_specs.source` takes `hand_authored`. They do not share a vocabulary and
  nothing says so — a whole transaction rolled back on it.
- **The verification query trap, hit three times on this site.** A blob `ILIKE`
  matches the prohibition text you just wrote (`honesty_rails` matched `%stock%`;
  `cta_style.never_use` matched `%Add to Bag%`; my own imagery prompts match
  `%packaging%` inside "No packaging"). Check per-key with `jsonb_each`, or read
  the rows rather than counting them.

---

## 5. Owed, open, and explicitly NOT mine

**Owed by this lane**
- The planner config change (§1b) — the only one.

**Open, recorded, not mine to fix**
- `bugs_open/122` — the last homepage contrast failure. `--color-primary` is used
  as both a fill and an ink; **measured that no value satisfies both**, including
  the site's own brand accent (4.41 against a 4.5 floor), so repointing would
  trade one failure for another. The arithmetic is in the bug file. Needs the
  generator.
- `bugs_open/122` also now carries `.news-list-tag` at 3.94:1 — `text_muted` on
  `border`, a component using the border slot as a fill. Arrived on a page created
  that day, so the class is still being reproduced by new components.
- `info-card-grid`'s image variant reads `var(--color-icon-chip-bg, #EEF2F8)` — a
  light literal no palette can reach (**0 of 18 layouts declare the slot**).
  `image-hover-card-grid` already reads `surface_alt` and is fine. Fixing the
  former means editing a fleet-shared template with 23 placements on 11 sites:
  its own change, its own review.

**Not this lane's**
- `grip-styles` → `gemini_content_provider`.
- `bugs_open/083` (stranded `detected` findings) — owner has ruled *"routing is
  NOT the bottleneck"* and *"decision pending — do not act"*. I promoted items
  **per site, as data**, and did not touch the mechanism. Do the same.
- `bugs_open/113` / `/114` / `/117` / `/098` — siblings of what I found; contribute
  into them, do not fork.

---

## 6. Where everything is

```
docs/agent_docs/docs024_key_docs_latest/dartsonline_traffic/
  PLAN_2026-07-29_…                      phases, owner decisions D1–D5, copy rails
  NOTES_…                                technical log — READ THE SESSION-2 TAIL
  README_where_we_are.md                 owner's plain-prose log
  RUNBOOK_…                              commands with their gotchas
  SUMMARY_2026-07-29_…                   session 1 (superseded on two points)
  SUMMARY_2026-07-29b_…                  session 2 — the current read-out
  PREVENTION_2026-07-29_…                18 defects → 6 levers  ← the owner asked for this
  SQL_2026-07-29{a..s}_…                 every applied change, with its reasoning
  SQL_OWED_planner_drop_…                the one unapplied change
```

**Read order for a cold start:** this file → `SUMMARY_2026-07-29b` → the session-2
tail of `NOTES`. The PLAN only if you are picking up phases 4–6.

**Council trail:** tool-ratio `f5fc3014-973c-49a2-8d42-4bf9b401eaeb` (APPROVED r2);
imagery `bf208075-6df2-4e5c-9a2a-a49acd0b63ec` (APPROVED, 5 advisory).

---

## 7. The one thing to carry forward

Every claim this workstream got wrong — and there were four — was **a claim about
an artefact, made by looking at something other than the artefact.** The phrase
sweep read stored HTML instead of the page. The nav diagnosis read a database
column instead of fetching the site. The stale listing looks perfect in every
status field it has. The pale icons were "known" from a document until I fetched
one and looked.

The corrective is not care, it is a habit: **fetch the thing you are about to make
a claim about.** `curl` is ten seconds; `scripts/render_audit.py` reads the painted
page; `Read` shows you a PNG. Two of these were caught by accident, an hour later,
for unrelated reasons — which is not a system.

Both wrong calls are written up in `WRONG_CALLS.md` with the cheap check that would
have caught each, and the transferable half is in `016b §9`.
