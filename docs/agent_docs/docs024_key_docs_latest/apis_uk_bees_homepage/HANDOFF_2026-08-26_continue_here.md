# HANDOFF 2026-08-26 — continue here

**Supersedes `HANDOFF_2026-08-25_continue_here.md`** (whose top correction block records the
08-25 owner split — still worth reading for the page's traps; its Google half is dead, see below).

> ## ▶ ONE-LINE STATE
> The page is done and locked. **Everything Google left this lane on 08-25** (owner ruling;
> `analytics_gtm` is the cold-start for tracking, `bugs_open/397`). The lane's remaining build,
> **per-section subjects, SHIPPED on 08-26**: Go + migration committed (`35905c547`), migration
> 638 applied, council **APPROVED at round 2** (`4bd35ed8`, ~12:35 BST; 3 advisories, all ACTED ON in
> `fa98a1961` — see §1a), three config seeds `_HOLD` awaiting the next chassis roll — and — NEW at
> 14:25 UTC — **the improvement loop added two tools + two guides (4 `planned` pages) to this
> single-page-by-ruling site; their publication is PARKED (`deferred`) pending the owner, §5c.**
> Otherwise only the image-accuracy A+C build remains, not started.

## 1. Per-section subjects — what a next session owes, in order

Full design + falsifiers: `PLAN_2026-08-26_per_section_subjects.md` · register **PBP-049** ·
council submission `scratchpad` copy is gone with the session — the artifacts are keyed under the
correlation.

1. **Council: SETTLED.** Round 1 REVISE (gating: prior_art_librarian, the applied-DDL tense —
   answered with the pre-state and the 2026-07-29 after-the-fact posture) → round 2 **APPROVED**
   ~12:35 BST, 3 advisories none high. `fa98a1961` carries `Council-Reviewed:`; `35905c547` and
   `52085b410` carry `Council-Submitted:` and are credited by `098` automatically.

1a. **The advisories, acted on same hour (`fa98a1961`):** rule 17's repeat-requires-subject was
   prompt-only (the 016b §9 decorative-decision pattern, bug_historian MEDIUM) → observe-only
   durable finding **`SUBJECT_MISSING_ON_REPEATED_COMPONENT`** in `write_site_plan`
   (`subjectlessRepeatFindings`, GATED on the plan carrying any subject so pre-640 plans are
   silent; declared in `finding_code_registry.json` same commit; gate mutation-proven). Every
   `_HOLD` seed header now carries the **pod-verification commands** (provenance stamp +
   `merge-base --is-ancestor 35905c547`, capability probe with positive control) and the
   **APPLIED-line convention** (a `_HOLD` file never reaches the ledger — the file IS the
   record). 641's owner read must be RECORDED (NOTES line + named in the APPLIED line).
   Guardian's caller doubts settled by measurement: `site-planner` does NOT call
   `write_site_plan`; the loader's caller set is still exactly `page-build-handler`.

2. **After the next chassis roll** (any session's — a roll ships this commit), apply the seeds
   **in order, by hand** (they are `_HOLD`; the runner skips them):
   `639` (wiring) → `640` (planner rule 17) → `641` (writer prompt v5).
   Every file carries drift-guarded anchors and aborts rather than mis-applies.
3. **`641` is double-gated: the owner must read the inserted block first** (quoted in full in the
   seed header — 4 lines). RFC_016 §5.2: the v4 approval attaches to its committed text and voids
   on edit. Do not apply 641 on an old approval.
4. **Then un-defer apis.uk's two `content_rewrite` items** (swarm, pollination — their rows name
   this chain as the unblock condition). ⚠ The locks refuse page re-renders by design, and adding
   a section needs plan surgery or a replan — that step is a decision, not a mechanical follow-on:
   write the two subjects into `site_plan_sections` for apis.uk (or replan), un-defer, and expect
   to settle `pages.build_status` afterwards (a render re-queues; settle → verify → settle again).
5. **Falsifiers before claiming anything works**: a plan with no subjects must be byte-identical
   end to end (NULL column, `omitempty` items); a replan with a repeated component must produce
   DISTINCT non-NULL `site_plan_sections.subject` on the repeats; after 641, same-named sections
   must differ in TOPIC, not wording. Adoption query (also the copy_quality lane's control):
   `SELECT count(*) FILTER (WHERE subject IS NOT NULL), count(*) FROM site_plan_sections sps
    JOIN site_plans sp ON sp.id=sps.plan_id WHERE sp.is_current;` — 0/N until the chain is live.

**Stated gaps, so nobody re-derives them:** the Pass B/B2 carry's object-realised arm is untested
(realised lists are plain strings everywhere today; the string arm is mutation-proven); the
carry's `unmatched` list stays facts-worded and drops an unmatched subject silently; seeding
`page_components.content_brief` from the subject at save time was considered and deferred.

## 2. ~~Known-red at HEAD~~ FIXED same day — the 396 lane declared their code

The morning CONTRIB worked: `a0ec90eb9` declared `WORK_ITEM_STATUS_OVERRIDE_REFUSED`, and the
full actions suite **including the finding-code scan is green at HEAD** (re-run 2026-08-26 ~12:45
via the verify-head-builds overlay). Nothing left to watch here.

## 3. RFC_022 parity — settled, and the lesson

`35905c547` grew plan_sections' Optional list 7→8 and I did NOT regenerate the cron literal in
the same commit — the **333 lane caught it at HEAD before I did** (parity test). Settled in
`339474ca4`: literal regenerated from a committed-HEAD extract, overlay re-applied, verified at
the mounted ConfigMap (`plan_sections: 8` with an unchanged neighbour as control). **If you touch
any ActionInputSpec Optional list in this lane again, check.py + overlay ride the same commit.**

## 4. Everything else this lane once held

- **Google/tracking** → `analytics_gtm/HANDOFF_2026-08-25_continue_here.md` (owner split, verbatim
  in the 08-25 handoff's correction block). apis.uk is in their `c2` rebuild wave's bucket B; they
  were asked to tell us when it runs so we re-verify the served bytes.
- **Image accuracy A + C** — designed 08-24 (see the 08-25 handoff §3 bullet 3), owner-approved,
  NOT started. D is live. C is agent config (`execute_vision_prompt` on `visual-design-auditor`),
  canary one apis.uk image first.
- **The page itself**: 200 / 67,877 B, `<h1>A closer look at bees</h1>`, 7 permanent locks,
  `tools.apis.uk` probe via `POST /api/v1/tools/gauntlet/round` (root 404 is by design).
  `[UNEXPLAINED, 08-25]` `sites.build_status='pending'` while the page row is `deployed` — no
  served effect, left alone, still true when checked 08-26 morning.

## 5. Design checks are visiting apis.uk again — measured, and the source is now settled

> **CORRECTED 2026-08-26 ~11:00:** the 00:40 visit was the **improvement-loop** (owner re-enabled
> it ~21:18Z on 08-25, loanzy lane's phased plan; it dispatches design-discovery + completeness as
> children — apis.uk got full cycles at 00:39, 04:47 and 08:40Z), NOT the rotation, which
> `webdesign-tool-rebuilds` re-enabled separately at 09:20Z (`bugs_open/401`). **Both are active
> now**, so expect findings roughly every 4h from the loop plus rotation visits. Their lane
> corrected its own NOTES on our timestamps (their commit `7baa7a4f1`).

Measured here rather than assumed: **six findings landed on apis.uk at 00:40:15–56 UTC**,
all status `detected`, **all with `handler_agent` empty — so NONE are promotable** by
`detected-item-promoter` (it requires `COALESCE(handler_agent,'')<>''`, plus the mig-629 origin
door). No auto-dispatch today; re-check that premise if anyone maps handlers onto these types.

| finding | read |
|---|---|
| `head_essentials_missing`: index missing `skip_link`, `footer` | **the footer half is BY DESIGN** (owner: no footer, no email — the empty `site_components.footer` row is the mechanism). Do NOT let anything "repair" it. The `skip_link` half is a real accessibility gap — a legitimate small fix, but see the trap below |
| `image_url_404` ×2: chrome references `/assets/images/favicon.png` + `og-card.png`, no active asset | real, mild (404 favicon/og-card). Fix = deploy the two assets, NOT edit chrome |
| `prerequisite_missing` ×2 (page_research, feed_sources) · `structure_floor_unmet` (1 of 6 structures) | the single-page-by-design shape failing fleet norms — expected; annotate rather than "fix" if they nag |

⚠ **THE TRAP fired overnight — §5b.** The locks guard `page_components` only, and that is
exactly what the 00:45–08:46 completions proved: sections survived, chrome did not.

## 5b. 2026-08-26 08:46 — the chrome refresh stripped GTM AND brought a fallback footer back

Measured ~13:15 BST, served page: 200 / 68,248 B · h1 + 6 sections + 7 images INTACT (7/7
permanent locks held, artefact-verified) · no email anywhere · **`googletagmanager` ×0** ·
**one `<footer>`** — the minimal `RenderFallbackFooter` shell (brand h3 + copyright only).
All three `site_components` rows regenerated together at 08:46:26 by the improvement-loop's
completed `needs_rerender`; page re-queued (`needs_rebuild` since 09:15). My §5 prediction that
apis.uk "cannot re-render" was WRONG — see `WRONG_CALLS.md` 2026-08-26, both entries.

**The finding that outlives the incident: the owner's no-footer ruling had the SAME latent
defect as the GTM backfill.** Emptying `site_components.footer` was artefact-only state; no
spec-level suppression exists (grepped: no `footer_disabled`/`suppress_footer`/`no_footer` key
anywhere), so ANY chrome refresh regenerates the shell via `RenderFallbackFooter`
(`component_library.go:1976`). Re-emptying the row now is 397-class churn — c2's imminent wave
would revert it again.

**State ~13:35 BST:** c2's spec key CONFIRMED on apis.uk (`site_config` current, row 10:12 —
an earlier read of mine said "no row" — RESOLVED: the row's own stamps say inserted 10:12:11 UTC by their session, my read simply preceded it); head
artefact still tagless (08:46); page `needs_rebuild`; their stale_chrome wave not yet here.
And CORRECTED: apis.uk does NOT refuse ordinary page re-renders (3 completed overnight — that
was the strip vector); expect the wave's items to COMPLETE and gtm to return on the served page.

**Sequence for whoever acts next (after `analytics_gtm`'s c2 wave lands on apis.uk):**
1. Verify GTM at the served bytes (c2 writes the spec key; the wave re-renders chrome WITH it).
2. **OWNER DECISION on the footer:** (a) accept the minimal fallback shell (brand + ©, no email,
   no disclosure — the harmful content is absent), or (b) commission the durable mechanism: an
   opt-in `site_config.chrome.footer_disabled` respected in `RenderFooter`/`InjectFooter`
   (council-scope, small; the 2026-08-02 opt-in ruling's shape). Interim row-emptying is NOT
   recommended — it re-fights every chrome refresh.
3. Settle `pages.build_status` (render → settle; it is `needs_rebuild` right now).
4. Open oddity: one `deactivated_component` item sits `unresolved` (08:40) — read before touching.
5. **NEAR-MISS, resolved by another lane (2026-08-26, migration `644`, register IMG-074):** the
   `Illustrated Text Block` component sourced `image_url` from `site_assets.image`, which
   ALIASES to hero — so the very c2/rebuild wave this section waits for would have run
   plan_sections and **overwritten all six distinct illustrations with hero-home.jpg**
   ("live resolution always wins" beats carryStored). The `vigilant_designer_offer_analysis`
   lane retargeted it to `site_assets.illustration` (no alias ⇒ resolves nothing here ⇒ carry
   PRESERVES the six) and fixed `image_alt` (it was handing the image URL to screen readers).
   **Verified at the component and at our six stored values, all distinct, this session.**
   ⚠ CONSTRAINT ON STEP 4's un-defer (from their unfixed pointer): apis.uk's five
   `site_plan_imagery` illustration rows are INERT (scope='page'; no resolver arm reads them),
   and even scope='section' resolution maps by kind FIRST-WINS — several illustration blocks on
   one page would all get the SAME image. So the two NEW sections (swarm, pollination) cannot
   rely on live imagery resolution: put their images in `content_data` + lock, exactly as the
   existing six were done (CLC-030), until the resolver learns per-section illustration mapping.
6. `chrome_divergence_overwritten` `2e4e5f51…` (00:44, the head strip's receipt) is
   `needs_human_review` — the platform ARCHIVED the pre-strip head (48,471 B), so the tagged
   artefact is recoverable evidence. Owner disposition queued via `bugs_open/397` §10-addendum;
   do not close it from this lane.

## 5c. 2026-08-26 09:18–11:47 UTC — the improvement loop ADDED TWO TOOLS to a single-page-by-ruling site; publication PARKED at 14:25 UTC pending the owner

**What happened (measured, path CORRECTED 14:50 from the loop owner's code read):** two
`add_tool` items (filed 09:18Z by the **design-discovery seat's `missing_tools` check →
`evaluate_tools` → `tool-suggester`** — not completeness; it runs in both the loop's audit pass
AND the design rotation; handler `tool-generator`) COMPLETED at 11:43/11:47 and
created **four `planned` pages** on apis.uk — `tool-bee-identifier-quiz` ("What Bee Did I See?"),
`tool-bee-foraging-calendar` ("UK Bee Foraging Calendar") and a 3-section companion guide for
each — **outside the site plan** (`site_plans` still 1 page) and **against the current
`roadmap_brief`** (single page, nothing to sell; 2026-08-23 row, `is_current`). Nothing is live:
all four are `planned`, unbuilt, undeployed.

**What would have published it:** nine `triaged` items — 2× `page_rerender` (tool pages),
2× `needs_content_page` (guides), 2× `content_rewrite` ("add <tool> reference to index page"),
1× `nav_drift` (nav rebuild + chrome re-render), 2× `improve_tool` ("tool page is planned, not
rendered" — would push the renders). None claimed when found.

**What this session did:** set exactly those nine to **`deferred`** (one transaction, count
asserted =9, reason + unblock written into `result` — the same reversible shape as this lane's
08-24 deferrals). The index's own `page_rerender` (11:40, restores GTM) was deliberately left
`triaged`. Pages were NOT deleted, items NOT cancelled: enforcing the owner's standing ruling
until he speaks is this lane's job; choosing the future of the tools is his.

**OWNER DECISION (the second one in this handoff):**
- **(a) accept tools on apis.uk** → set the nine back to `triaged` (`UPDATE … WHERE result->>'deferred_by'='apis.uk session 2026-08-26'`), and expect: two tool pages, two guides, nav links, and a rewritten index with tool references — i.e. apis.uk stops being the single quiet page.
- **(b) keep the ruling — ORDER MATTERS:** (1) FIRST get the durable refusal built — there is
  **no per-site exclusion today** (loop owner, from code: the check's only gates are deployed-tool
  count, an opt-in growth ratio that can only ADD pressure, and a cooldown; tool-suggester reads no
  roadmap/single-page key). The fix shape is a per-site refusal DECLARATION the check honours —
  `sites.settings->maintenance_profile->growth->>'refusal'`, mirroring 624's
  `structure_floor->>'refusal'` — small Go + a council round; recorded by the loop owner as an
  **RFC_056 follow-up** and being put to the owner. (2) ONLY THEN cancel the nine and delete the
  four `planned` rows. ⚠ Do NOT cancel first: `deferred` is an OPEN status for `idx_swi_dedup`, so
  the nine parked rows HOLD their dedup keys and every re-file dedups onto them and inserts
  nothing; cancelling releases the keys and the next cycle re-creates the whole set.
- Default while undecided: leave deferred — **the park IS the hold**, and it is durable for as
  long as the rows stay deferred. Nothing publishes.

**Also on the page's queue since 644:** `image_source_unsatisfiable` (`needs_human_review`,
11:24) — the checker now flags that `illustrated-text-block` sources `image_url` from
`site_assets.illustration`, which resolves nothing here. **That is the PROTECTION working**
(644 made the six stored illustrations carry instead of being hero-clobbered) — the item is a
true statement and a false alarm; told the 644 lane, leave it for them/the owner.
