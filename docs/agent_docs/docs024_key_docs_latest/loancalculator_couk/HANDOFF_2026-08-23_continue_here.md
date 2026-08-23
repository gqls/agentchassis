# HANDOFF — loancalculator.co.uk · the 08-17 incident is CLOSED; two items remain (2026-08-23)

> Supersedes `HANDOFF_2026-08-17b_continue_here.md`, whose state block is now stale in
> almost every line (it predates the retraction, two chassis rolls and the deploy outage
> clearing). **Read 17b only for the mechanism and the missteps** — its "THE DAMAGE" and
> "DEPLOY ⛔" sections are history. Evidence for everything below: NOTES `## 2026-08-17`
> §1–§12, `## 2026-08-18`, `## 2026-08-20`, `## 2026-08-23`. Owner prose:
> README_where_we_are.

```
site      loancalculator.co.uk   0162cde4-633e-45e9-8ca6-87a6b2fe1d26
chassis   v1.0.1328, stamp 2dbe12f1d — VERIFIED 2026-08-23 (pod digest sha256:5e740e38…
          matches the local image for that tag; stamp present in /proc/1/exe, previous
          stamp 2d13d530d ABSENT, so the probe discriminated; ancestor of HEAD)
pages     29 active · 28 serving 200 · guides-index (/guides/index.html) the ONLY 404
          14 guides at /guides/<slug>.html · 11 tool pages · 15 archived
locks     12/12 held throughout everything below
plan      9463e31d-ee50-482e-94a9-7e186ef25543 is_current (created 08-17; no replan since)
flags     honour_realised_identity + twin_identity_snap + stem_twin_snap ALL TRUE,
          url_shape flat, 27-entry pages list intact — seeded 08-18, re-verified 08-23
deploy    healthy (the 08-17 fleet outage cleared; retraction committed and published
          the same minute)
```

## SETTLED — do not re-litigate, all measured

- **`bugs_open/282` proven on the motivating case.** The recompose placed 11/11 locked
  calculators on their own pages against a 0/11 baseline, joined on
  `content_components.function`. The homepage then rebuilt with its calculator at
  **position 2** (was appended at 6). `toolgolden --compare
  GOLDEN_2026-08-17_post_rebuild` → **exit 0**, "all 11 tools reproduce their golden
  values exactly".
- **The 08-17 duplicate-page incident is CLOSED at the artefact.** 14 `/blog/` pages
  archived (08-20) and retracted (08-23, corr `d7f7f5b3`, orchestration `8045c4a9`):
  considered 14, retracted 14, **0 refusals**, nav_retired 0, editorial_inbound null,
  stranded_targets null, one git commit `a1508b92` to `gqls/sites`. **All 14 URLs now
  404; the guides, index and tools verified still 200.**
- **The cause is fixed, not just cleaned up.** The planner moved those pages because
  `CanonicalisePage` cannot express a flat `/guides/<slug>.html` for `role=guide` at all,
  so this site's real URL shape was unrepresentable and `blog-post` was the nearest
  expressible role. `bugs_open/241`'s identity policy — written while planning THIS site
  — stops the write path re-deriving a live page's identity, and is now ON here.

## OWNER'S FOUR INSTRUCTIONS (2026-08-23): *"1. deleted, 2. release the rebuilds. 3. build and restore the Guides link 4. we only need one of them."*

| # | state |
|---|---|
| 1 | ✅ done and verified — 14/14 `/blog/` URLs 404, controls held |
| 2 | **canary PASSED**, nine released and queued (see below) |
| 3 | ✅ `/guides/index.html` serves **200** and lists all 14 guides. ⚠ the LINK needs the queued `nav_drift` |
| 4 | `tool-credit-roadmap` ARCHIVED + its 3 tickets cancelled; **file retraction still owed** |

**(2) "Release" is NOT a status flip — the 11 `owned_page_review` tickets have
`handler_agent = ''`.** They are review MARKERS (TP-004: "no handler by design"); triaging
them rebuilds nothing. The rebuild is a separate `needs_page` / `page_rerender:<page>` /
`page-build-handler` item — shape copied from the row that rebuilt `tool-credit-roadmap` on
08-22. Nine are queued at priority 15, `source='loancalc_owner_release_20260823'`.

**The canary (`tool-overpayment-calculator`) passed on the untested arm** — a tool-role page
whose calculator is LOCKED. Served order is now `hero · CALCULATOR · prose · faq · cta`, and
**the locked row was never written** (`updated_at` still 2026-08-09 while every sibling shows
13:24:19). No copy changed: every archived/saved pair is md5-identical.

⚠ **THE ACCEPTANCE HARNESS IS DOWN.** `toolgolden --compare` fails with `timeout waiting for
Runtime.evaluate` — **on rebuilt AND un-rebuilt pages alike**, so it is the environment, not
the change. Do not read it as rebuild damage, and do not quote a toolgolden pass after
2026-08-23 without checking it actually captured. The 08-17 golden is ALSO stale (a FAQ
heading changed in an 08-17 19:00 re-deploy, after the capture) — re-baseline once it works.

**(3) The Guides 404's cause: the plan composed ZERO sections for that page** while `about`
and `legal` had two each — verbatim what its build kept reporting. Fixed with two
`site_plan_sections` rows (`hero`, `guide-list`, the fleet convention for `section-index`).
⚠ **Restoring the LINK is a SECOND mechanism**: the chrome was last rendered 08-20 and
carries no `/guides/` link, because the renderer correctly declined to link a 404ing page.
A `nav_drift` → `nav-updater` item (`nav_rebuild:e31c71a8…`) is queued to rebuild the nav
tables and re-render chrome. **A session that only built the page would see a working page
and a menu that never mentions it.**

**(4) Keep `tool-credit-health-check`** — both pages carry the same component function, but
its instance is the LOCKED one; credit-roadmap held an unlocked duplicate. Archived, tickets
cancelled, and excluded from the nine rebuilds. ⚠ **Its file still serves and its retraction
will be REFUSED**: 15 pages carried 16 links to it plus an active nav row. The sequence is
archive (done) → let the nine rebuilds regenerate their cross-links without it → then
`retract_page_deployment`. **[INFERRED] that the rebuilds drop the links — the nine are its
test.** If they persist, the retraction stays refused and it becomes a prose decision.

## THE TWO REMAINING ITEMS

### 1. The calculators sit at the BOTTOM of ten tool pages — 11 tickets await the owner

Measured on `tool-overpayment-calculator`, typical of the ten:

```
live now :  hero, ported-prose, faq, tool-cta, [calculator LAST, position 5]
plan says:  hero, [calculator], ported-prose, faq, tool-cta
```

So on a site whose product IS the calculators, the calculator is below the article text,
the FAQ and a call-to-action on ten of eleven calculator pages. **The improvement is
demonstrated, not theoretical:** the homepage made this move (6 → 2), and
`tool-credit-roadmap` re-deployed 08-22 and now matches the plan exactly.

Gated by 11 `owned_page_review` items at `needs_human_review`, open since 08-15 — TP-004's
deliberate human gate on tool pages, because the generic builder clobbers them. **Releasing
them is the owner's call.** After any rebuild, verify at the artefact with
`toolgolden.py --compare acceptance/GOLDEN_2026-08-17_post_rebuild_tool_values.json <the 11 urls>`
(URLs from `pages.url`, never name-derived) and re-check locks 12/12.

### 2. `guides-index` — the last 404

`/guides/index.html` has never built: its rerender returns `needs_human_review` with *"no
sections ready to build (empty spec sections)"*. It IS in the plan as `section-index` and
**its nav entry already exists** (`site_nav_items`, primary, position 2), so building it
restores the Guides menu entry. It needs sections composed, which means a planner run.

## ⚠ BEFORE ANY PLANNER RUN — this is what bit us

1. **Run `./checkpoint_postplan.sh` the MINUTE a new current plan lands**, not later. Its
   step 1 (new page rows after fire — expect NONE) and step 6 (page-identity md5) are
   exactly what caught the 08-17 invention, and I ran them ~35 minutes late, by which time
   the builds had dispatched. **A post-condition check is worth what its latency prevents.**
2. **Check the item KEYS against the plan the run just wrote**, not against what the last
   session's notes led you to expect. `needs_page:can-i-overpay` is not
   `needs_page:guide-can-i-overpay`; I read that list twice without seeing the missing
   prefix because an inherited framing told me it was routine churn.
3. **Re-verify the identity flags first** — a re-adoption silently drops them
   (LANDMINES), and the drop looks identical to them working.
4. **Pass C2 will NOT save you.** The guard that drops a re-proposed duplicate is
   structurally unreachable on a re-plan: its index is built from `noCurrentPlanPages`,
   "empty whenever the site has a current plan". Worth its own bug; not filed.

## Standing cautions (carried, all still true)

- **Prove a deploy at the artefact.** Compare the pod's `imageID` digest against the local
  image's, then `git merge-base --is-ancestor <fix> <stamp>`. Never grep the binary for the
  fix's own sha — it carries only its build commit. Two "fresh builds" this week were not
  live: one same-tag rebuild the nodes ignored, one where only the tag had moved.
- Verify tool placement at `site_plan_sections`, never `pages.sections` (LOCK-008 merges
  locked rows into the latter).
- The phase-2 script's judge query `component_name LIKE 'tool-%'` returns 26 either way —
  it matches `tool-cta`/`tool-list`. Use the locked-function join in 17b.
- A hand-filed or un-parked work item must be `triaged`; the dispatcher cannot see
  `detected` and fails silently.
- Query runs BY CORRELATION, never `now()`-interval. **A planner run's `collected_data`
  purged within ~2 hours** on 08-17, though the older cautions say ~2 days — read a run's
  payload the moment you have a question about it.
- `retract_page_deployment` REFUSES an active page, so archive first; and its DEFAULT
  selection is every non-active page with a deploy stamp, which on this site would also
  take `tool-standard-calc`. **Use explicit `page_ids`** — `retract_blog_duplicates.sh` is
  the worked example.
- A single 404 sampled during a rerender wave proves nothing; re-sample.
