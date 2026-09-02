# 397 — GTM was backfilled into the stored head artefact only: the next chrome render reverts it (it already has, once), and every site born since 07-31 is untagged

**Filed 2026-08-25** by the `analytics_gtm` lane (session "google"), from
`docs/agent_docs/docs024_key_docs_latest/analytics_gtm/`. Found while checking what a GA4 tag
publication would actually record.
**Status: OPEN. Fix written and dry-run, NOT applied — applying it is an owner-timed rebuild wave (§7).**
**Severity: medium.** Nothing a visitor sees is broken. The harm is that the estate's analytics
coverage is silently shrinking: **12 of 26 tagged sites will lose the tag on their next chrome
render, one already has, and 4 new sites never had it** — and every document that says "27/27 heads
carry it" was true for six hours.
**Class:** a backfill wrote the ARTEFACT and not the SOURCE the template reads. Same family as
`bugfix-117` (template vs stored chrome) and CLC-030 (`page_divergence_overwritten`), one layer up.

> **On the 090 loop: NOT run — first-hand verification substituted, and here is what it consisted of**
> (owner ruling 2026-07-31 requires this to be stated, not skipped). The mechanism is three readable
> lines: the gate `{{if .gtm_container_id}}` in the head template (read from `content_components`,
> §3); the input source, `site_specs.site_config.analytics.gtm_container_id`, declared in the
> component's `input_schema` and fingerprinted by `ChromeRenderInputsSQL`
> (`platform/orchestration/datahelpers/chrome_render_inputs.go:100-120`, read); and the check that
> makes the fix a rebuild, `StaleSiteComponentsCheck`
> (`platform/orchestration/actions/discovery_checks/check_integrity.go:375-432`, read). The claim
> "the next chrome render drops it" was then tested against its **disconfirming case** — a head
> re-rendered after the backfill that KEPT the tag — and instead of a refutation the query returned
> one site that had **lost** it, on **two independent instruments** that agree to the second:
> `site_component_history` (§4) and the platform's own `chrome_divergence_overwritten` row (§4). A
> diagnosis agent re-reading the same template would add nothing this file does not already cite.

## 1. The one-paragraph version

Every head template on the estate emits the GTM snippet inside `{{if .gtm_container_id}}`, and that
value comes from one place: the site's current `site_config` spec, key `analytics.gtm_container_id`.
The 2026-07-31 rollout set that key on 14 sites **and** rendered the artefacts — durable. The
2026-08-24 fleet backfill (`apis_uk_bees_homepage` lane) inserted the snippet directly into
`site_components.rendered_html` on 13 more sites and wrote no key. That served the tag immediately
(pages inject the stored head), which is why every check passed. But chrome is regenerated from
template + inputs whenever the platform renders it — a `nav_rebuild`, a `stale_chrome` item, a site
rebuild — and with no key the gate is false, the snippet is gone, and the divergence detector
records the hand-edit as an overwrite. agritec.uk did exactly that at 19:20 on 08-24, six hours
after the backfill. Separately, nothing in the build pipeline writes `site_config` for a new site
(every current `locale`/`analytics` row was written by an operator migration or session), so the
four sites built since — agritec.uk, cv1.co.uk, homegarden.uk, lampenkap.com — have no key and no
tag, and the owner's 08-24 instruction "make it standard for new builds" has no mechanism behind it.

## 2. Evidence — all `[MEASURED 2026-08-25 15:00–16:20 BST]`, predicates inline

**Which templates, and what they read** (`content_components`, deployed sites' `head` slot):

| component | sites | `html_template LIKE '%GTM-PQ3WCTBD%'` | `LIKE '%googletagmanager%'` | `input_schema` key |
|---|---|---|---|---|
| `Document Head` `116c5f91…` | 23 | false | **true** | `gtm_container_id` |
| `head-seo-standard` `aec98dbe…` | 4 | false | **true** | `gtm_container_id` |
| `webdesign.co.uk Document Head` `14cf6193…` | 1 | false | **true** | `gtm_container_id`, `cf_analytics_token` |

No template carries the literal. All three carry the gated block (verbatim, `Document Head`):
```
  <meta charset="UTF-8">
{{if .gtm_container_id}}<!-- Google Tag Manager -->
<script>(function(w,d,s,l,i){ … })(window,document,'script','dataLayer','{{.gtm_container_id}}');</script>
<!-- End Google Tag Manager -->{{end}}
```
The header slot on 11 of the 12 affected sites is `header-theme-chrome`, whose template also reads
the key (the `noscript` half). That is why those sites serve `gtm=1` (script only) while the 07-31
sites serve `gtm=2`.

**Where the key lives** — `analytics_gtm/sql/c1_gtm_fleet_rollout.sql:168-176` (07-31):
`INSERT INTO site_specs (site_id, aspect='site_config', data='{"analytics":{"gtm_container_id":"GTM-PQ3WCTBD"}}' …)`,
read "via input_schema source config.analytics.gtm_container_id". Post-condition there: 14/14.

**The census, 2026-08-25** (`sites.status IN ('deployed','active')`; spec key =
`EXISTS current site_config row WITH data->'analytics'->>'gtm_container_id'='GTM-PQ3WCTBD'`;
artefact = `site_components.rendered_html LIKE '%GTM-PQ3WCTBD%'` on `slot_name='head'`):

| bucket | n | sites |
|---|---|---|
| A durable — spec + artefact | **14** | the 07-31 fourteen |
| B **artefact only** — reverts on next chrome render | **12** | apis.uk cookly.uk garden-tools.uk lendzy.co.uk loanandmortgagecalculator.co.uk loancalculator.co.uk loancash.co.uk loanzy.uk mortgagecalculator.co.uk noted.co.uk remortgagecalculator.uk(active) webdesign.uk |
| D neither | **4** | agritec.uk cv1.co.uk homegarden.uk lampenkap.com — heads created 08-24/08-25, all `Document Head`, all `sites.settings = {}`, **no `site_config` spec row at all** |
| E no head component | 1 | adversecreditmortgage.co.uk (active) |

All 12 bucket-B heads: `updated_at = 2026-08-24 13:13:52.096044+00` — the backfill instant, none
re-rendered since. The 08-24 `psql` history rows: **13** heads written, **12** still carry it.

**Served, one curl per deployed domain, redirects followed:** 24 of 25 tagged domains serve the
snippet; `webdesign.uk` 302→`webdesign.co.uk` (gtm=2 there). agritec/cv1/homegarden/lampenkap
were not in that list because they are not tagged.

**What the fix triggers** — `check_integrity.go:378-388` selects chrome rows whose stamped
`render_inputs IS DISTINCT FROM (ChromeRenderInputsSQL)`; that expression md5s
`site_config`+`identity`+`design_intent` (`chrome_render_inputs.go:116-120`). Writing the key
changes the fingerprint on every slot → one `needs_rerender` item per site (`item_key
stale_chrome`, handler `rerender-pages`, "force-rerenders ALL slots on any firing" per the comment
at line 366). **It dispatches:** 20 `stale_chrome` items ever (18 archived + 2 live), all `complete`.
Sized: 12 sites → **241 pages**; with the 4 untagged, 16 sites → 280.

## 3. The mechanism, in three parts

**What the thing is.** A head template does not contain the tag. It contains a *gate* — "if this
site has a container id, emit the snippet with it" — and the id is looked up per site in its
`site_config` spec. The stored head in `site_components.rendered_html` is a *cache* of that
template's last output, which pages copy in when they render.

**The rule.** The cache is regenerated from template + inputs whenever chrome renders (nav change,
stale-inputs detection, site rebuild), and the platform checks the regenerated output against the
stored one: if they differ, the stored one is overwritten and a `chrome_divergence_overwritten`
review item is filed (CLC-030's chrome twin). So an edit to the cache alone is, by design, a
temporary state that the platform will detect and undo.

**This case.** The 08-24 backfill edited the cache on 13 sites and not the spec. On agritec.uk the
lane added a tool page at 19:01; `nav_drift` → `nav_rebuild` completed 19:21; the head was
regenerated at 19:20:53 with no key, the gate was false, the snippet vanished, and the platform
filed `chrome_divergence_overwritten:site_component:head:920676eab287` at `needs_human_review` —
19:20:53.239, the same instant. The other 12 have simply not had a nav change yet.

## 4. The production instance (two instruments, one second)

```sql
SELECT s.domain, h.created_at, h.application_name
  FROM site_component_history h JOIN site_components sc ON sc.id=h.site_component_id JOIN sites s ON s.id=sc.site_id
 WHERE sc.slot_name='head' AND h.rendered_html LIKE '%GTM-PQ3WCTBD%' AND sc.rendered_html NOT LIKE '%GTM-PQ3WCTBD%';
-- agritec.uk | 2026-08-24 19:20:53.178828+00 | app - 10.20.31.31:32834      (the ONLY row)

SELECT s.domain, x.status, x.created_at FROM (live ∪ archive site_work_items WHERE item_type='chrome_divergence_overwritten') x JOIN sites s …
 WHERE x.item_key LIKE '%:head:%';
-- agritec.uk | needs_human_review | 2026-08-24 19:20:53.23915+00          (the ONLY row)
```
The disconfirming result would have been a bucket-B head with a history row *after* 13:13:52 that
still carries the tag. There is none.

## 5. Why this matters this week

The owner is about to publish a GA4 tag into `GTM-PQ3WCTBD` (container is at version 2, 0 tags,
`[MEASURED 16:18 BST]` — see `analytics_gtm/scripts/check_gtm_state.sh`). When he does, **history
starts at that moment and only for sites that serve the snippet**. Every bucket-B site that takes a
nav change between now and the fix drops out of GA4 silently — no error, the page still renders,
Realtime just stops showing it. And the `chrome_divergence_overwritten` row it files reads as "an
operator hand-edited chrome and we corrected it", i.e. as the platform doing its job.

## 6. Fix candidates — ordered by what makes the bad state unrepresentable

1. **Write the source, not the cache — `analytics_gtm/sql/c2_gtm_spec_key_for_artefact_only_sites.sql`.**
   Merges `analytics.gtm_container_id` into each bucket-B site's current `site_config` (supersede +
   insert; **merge**, because 10 of 12 already hold `locale`/`chrome`, and the 07-31 rollout's
   wholesale replace dropped relojistas.com's `intent_probe` key — measured). Guarded (`-v GO=yes`),
   dry-runnable (`-v DRY=1`, verified: 12 targets, 241 pages, keys preserved, 0 rows written),
   post-conditioned (every target keyed, exactly one current row, no key lost), scoped to
   `network_id = …0002` so a third-party site can never receive our container from it.
   **Cost: it IS a rebuild wave** (§2) — 241 pages, one commit and one Actions run each on the
   two-slot runner, ~2 h of queue, on top of the 130 still queued from 08-24. Owner picks the moment.
   `-v UNTAGGED=1` adds the 4 bucket-D sites (280 pages) — "standard for new builds" made true for
   the four that missed it.
2. **Give new sites the key at birth — the structural half, and an open design question.**
   > **REQUIREMENTS LANDED 2026-08-26** (owner ruling relayed by the webdesign.uk lane,
   > `webdesign_uk_build_service/DECISION_2026-08-26_default_tag_hosted_copy_only.md`): customer
   > `<slug>.ugg2.com` builds default to the owner's container on the HOSTED copy only; per-site
   > override (customer id → hosted + ZIP; "none" → no tag anywhere); the ZIP always ships clean of
   > the owner default. Design consequences adopted here: (a) an explicit `analytics.mode`
   > (`default|custom|none`) must sit beside the id — once a seeder exists, "none" has to be a
   > stored fact it honours, not an absence it re-fills; (b) the default value lives in ONE place
   > read by both the seeder and the ZIP exporter, so "is this the owner default?" is a comparison,
   > never a literal; (c) the export should re-render with `mode=none` rather than strip markup.
   > Their doc's "no per-site tag field exists" is a false absence (measured at `sites.settings` +
   > Go; the seam is `site_specs` — STY-050) — corrected by message, their doc to carry the note.
   > ⚠ **Collision to put to the owner before customer intake ships:** the pending GA4 publication
   > goes INTO `GTM-PQ3WCTBD`; from that moment the same container on a hosted customer site sets
   > `_ga` cookies with no banner — the decision's own §5 re-ruling trigger. A second cookie-light
   > container for customer sites, Consent Mode, or a re-ruling are the options.
   > **RULED 2026-08-26 (night): option one — a SECOND cookie-light container** ("please go ahead
   > with a second GTM analytics container"; recorded in the webdesign.uk lane's DECISION doc §5).
   > This UNBLOCKS estate GA4 publication into `GTM-PQ3WCTBD` entirely. Execution is this lane's,
   > and is **blocked on access, not on decision**: a container is created inside the owner's
   > Google account, and no Google credential exists on our side (`[MEASURED 2026-08-25, re-checked
   > 2026-08-26]` — no gcloud, no `GOOGLE_APPLICATION_CREDENTIALS`, no cluster secret). Two paths:
   > the owner's ~2 minutes in the GTM dashboard (walkthrough in the lane README), or a service
   > account with the Tag Manager API — the same credential Search Console needs, one grant
   > unblocks both. Its `GTM-XXXX` id then becomes the ONE-place fleet default for customer builds.
   > ⚠ **Name the trap before it recurs: an EMPTY container is cookie-light AND records nothing** —
   > exactly the `GTM-PQ3WCTBD` 0-tags lesson — while the decision's stated purpose is seeing
   > whether delivered sites get visits. When the container is published it needs a GA4 tag with
   > **Consent Mode defaults DENIED** (cookieless pings, no `_ga`), speced at creation, or the
   > refund-judgement purpose silently gets zero data and someone rediscovers this in three weeks.
   > **CREATED 2026-09-02 by the owner: `GTM-TH5XGNQ4`** (second container under the
   > leopardessconsulting GTM account). Verified live at the artefact 20:01Z: HTTP 200, version 1,
   > **0 tags, no `G-` id — cookie-light by construction, and for THIS container 0 tags is the
   > CORRECT state** (the check script's NOT-PUBLISHED verdict reads inverted here). Until the
   > seeder/export mechanism is built, the id's canonical record is this bullet, the lane handoff,
   > and `webdesign_uk_build_service/DECISION_2026-08-26_default_tag_hosted_copy_only.md`; when the
   > mechanism ships, the value moves into the ONE machine-readable place both paths read. No Go
   writer touches `site_config` (every current row is `created_by` a migration or a session), so
   "standard for new builds" needs a seeding step. The seam is right as it is — opt-in per site,
   unsafe default OFF, exactly the shape the 2026-08-02 §2 ruling prescribes — so the fix is *where
   the opt-in is set*, not a default in the renderer. Candidates: the site-creation action
   (`site_db_actions.go`, the `INSERT INTO sites`) seeding `site_config.analytics` from a
   network-level setting for `network_id …0002`; or the `SEED_*.sql`/`082_submit_domain_unified.sh`
   path. **The third-party question must be answered first**: a handed-over site
   (`sites.handed_over_at`, `transfer_confirmed_at` exist) must not carry our container, and today
   nothing distinguishes one at creation. Council scope (`platform/`); belongs to `analytics_gtm` +
   `web_admin_console` (handover). Not started.
3. **A detector, so the next backfill cannot pass silently:** a nightly check (the
   `cmd/config-key-audit` fleet is the home) that flags any `…0002` site in `deployed`/`active`
   whose head artefact and spec key disagree, or whose head has neither. The manual form exists
   today: `check_gtm_state.sh --db`, bucket B ≠ 0 or D ≠ 0 is the finding.

## 7. What NOT to do

- **Do not backfill `rendered_html` again.** It is how this bug was made; it would pass every check
  it passed on 08-24 and revert the same way.
- **Do not add GTM to `RenderFallbackHead`** (`component_library.go:2243`), as
  `apis_uk_bees_homepage/HANDOFF_2026-08-25 §3` proposed under "per-site analytics id". That is the
  path taken only when the head component FAILS to render; a tag there hides a broken head. And the
  per-site id seam it designs (`sites.settings->>'analytics_container_id'`) **already exists** as
  `site_config.analytics.gtm_container_id` (STY-050, live since 07-31, 14 sites) — a second key would
  be two switches for one light.
- **Do not hand-stamp `render_inputs` to dodge the rebuild.** It would be a true statement about the
  head (the artefact matches) and a false one about the header (no `noscript`), and a stamp that
  lies is the trap this estate keeps paying for.

## 8. How to verify

- Before: `check_gtm_state.sh --db` → B=12, D=4. After apply: B=0; then `needs_rerender` items with
  `item_key='stale_chrome'` appear per target as discovery runs; after they complete, every target's
  head `updated_at` moves past the apply time and still `LIKE '%GTM-PQ3WCTBD%'`, and a served page
  reads `gtm=2` (script + noscript) where it read `gtm=1`.
- Falsifier for the whole file: a bucket-B site takes a `nav_rebuild` before the fix and **keeps**
  the tag. Query in §4 — the row would show `had before AND has now`.

## 9. Cross-lane

- `apis_uk_bees_homepage` — the backfill was theirs; told via
  `CONTRIB_2026-08-25_from_analytics_gtm_…` in their directory. Their handoff's "27/27 heads" and
  "GTM live across the estate" are dated true and now false by one site (agritec) with 12 pending.
- `agritec_uk` — their site is the instance; nothing they did was wrong (a nav rebuild is normal).
  The `chrome_divergence_overwritten` row at `needs_human_review` on their site is THIS bug, not an
  operator hand-edit to review.
- `loanzy_uk_example_site`, `webdesign_uk_build_service`, `bugfix_357`, `bugfix_384` — own or are
  working bucket-B/D sites; a rebuild wave lands on their pages. Tell before firing.

> **OWNERSHIP, owner ruling 2026-08-25 ~17:15 BST (relayed by the apis.uk session, verbatim: *"section 4
> has google in it which is taken by another lane, please communicate to that lane that that is what
> they take and we will take the rest here"*).** This bug — the c2 rebuild AND the §6.2 structural
> half — is `analytics_gtm`'s, along with GA4 publication, Search Console, the fleet dashboard script
> and `039_REFERENCE`. Recorded in
> `docs/agent_docs/docs024_key_docs_latest/analytics_gtm/CONTRIB_2026-08-25_from_apis_uk_bees_homepage_owner_ruling_you_take_everything_google_we_keep_the_rest.md`.

- **`apis_uk_bees_homepage` — ADDED 2026-08-25 at their request.** apis.uk is in bucket B, and its
  index page **refuses page-level re-renders**: `[MEASURED 2026-08-25]` the 11:19 BST
  `page_rerender_index_…_template_changed` item is `failed` with `result = {}` (no error recorded —
  the `bugs_open/099` pattern; the reason is in the orchestration's `__step_error`), and the 383
  lane's `9a843c06a` reads "apis.uk cannot re-render". **So expect the wave's page item on apis.uk to
  fail. That is not damage:** the served page already carries the tag, the head *artefact* will be
  right once the key exists, and a render re-queues the page (`build_status='needs_rebuild'` is
  queue membership — their trap). **After c2 runs, tell them**; they verify apis.uk at the served
  bytes and settle `build_status` themselves.
  > **CORRECTED 2026-08-26 (by the apis.uk lane, adopted here): "apis.uk refuses page re-renders"
  > is DISPROVEN — three `page_rerender`s COMPLETED on apis.uk overnight (00:45, 04:53, 08:46Z),
  > which is exactly how its tag was stripped.** The 08-25 11:19 `failed` item I measured was real
  > but was the 383 lane's canonical re-walk specifically, not a standing property; the permanent
  > locks hold the 7 `page_components` rows and a page re-render completes around them while
  > regenerating chrome. So expect apis.uk's wave items to COMPLETE and the served page to regain
  > the tag with no unblock step. The wrong generalisation was theirs (their WRONG_CALLS
  > 2026-08-26); the too-strong adoption after measuring one failed row was mine.

---

## §10 — 2026-08-26: the prediction landed fleet-wide overnight, and c2 is APPLIED

**The strip.** The design-discovery rotation was re-enabled 2026-08-25 (`bugs_open/401` — the
`webdesign_tool_rebuilds` lane; it had sat off 15 days). Its sweeps promote `needs_rerender` /
`deactivated_component` to `rerender-pages` — chrome-touching repairs. `[MEASURED 2026-08-26 ~10:15 BST]`
17 head re-renders between 08-25 18:00Z and 08-26 06:00Z: **every keyed site kept the tag (7/7:
robot-hands, vetcomparison, idea.uk, oufe, gamesdesign, vonc, gaswholesalers); every artefact-only
site lost it (10/10: cookly 21:21Z, garden-tools 22:07, loanzy 23:32, lendzy 23:33,
mortgagecalculator 23:58, loancash 00:13, apis.uk 00:44, remortgagecalculator 01:15, noted 03:34,
webdesign.uk 05:54).** With agritec that is 11 losses ever; only the two big loan-calculator sites
still carried the artefact. A natural experiment this clean is the §8 falsifier run 17 times: the
spec key is the only variable that predicts survival.

**A false reassurance corrected en route:** the relayed claim "no design finding type carries a
handler_agent — nothing promotable, no repair fires yet" was refuted by the same 00:40Z apis.uk
batch it was measured on (`needs_rerender`/`deactivated_component` → `rerender-pages`, complete).
The handler-less types are the *design* ones; the chrome-touching promotions were live all night.

**c2 APPLIED 2026-08-26 10:12:11Z** — the 17 rows' own `created_at`, one instant; my first write here said `~10:50Z`, an unmeasured estimate in a measured voice, caught by the apis.uk lane reading the stamp (corrected 2026-08-26) — (owner: "please carry on"), `-v GO=yes -v UNTAGGED=1`:
**17 sites** (the 15 stripped/new + the 2 remaining artefact-only + `adversecreditmortgage.co.uk`,
which had gained a head overnight), **334 pages** at apply time (323 at the dry run an hour earlier —
sites build under you; date your counts). `UPDATE 11 / INSERT 17`, all post-conditions passed, keys
preserved (`analytics[,chrome][,locale]`). §9's lanes notified beforehand (9 of 10 messages
delivered; the loanzy session's was blocked by a local classifier — this file and the handoff are
the record for that lane). Census after: **A durable 16 · C spec-only 15 · B 0 · D 0**; 17 current
`site_specs` rows `created_by='claude-session-google-2026-08-25'`.

**What remains open here:** (1) the 15 C-bucket sites regain the *served* tag only as each site's
`stale_chrome` → rebuild lands — verify with `check_gtm_state.sh --db` (C should drain to A) and at
served bytes; apis.uk's page item is expected to fail (§9) and its served page may stay `gtm=0`
until the 383 lane's render blocker clears. (2) The §6.2 structural half (new sites born unkeyed —
`adversecreditmortgage` proved it again by arriving overnight) is unchanged and still owed.
(3) webdesign.uk per-page baseline from its lane: tag was 5/7 pages before the strip; expect 7/7
after its rebuild; their hand-placed "Not active yet" label on index will be wiped by the rerender —
theirs, expected. (4) homegarden note from `copy_quality_two_stage`: any *planned* page the wave
builds (not just re-renders) runs the post-627/628 writer — copy differences across that boundary
are not this change's doing.

### §10 addendum, 2026-08-26 ~12:30 BST — the 11 review items, their disposition, and the banked facts

**Every `chrome_divergence_overwritten` `:head:` item at `needs_human_review` matches a strip event
to within 130 ms** — 11 items, one per stripped site (agritec `6b320f3e…`, cookly `103e8baa…`,
garden-tools `e4cf75ec…`, loanzy `94dd0a5e…`, lendzy `eb67c4ed…`, mortgagecalculator `75e76f77…`,
loancash `f11b3679…`, apis.uk `2e4e5f51…`, remortgagecalculator `e452e9b1…`, noted `efa20f84…`,
webdesign.uk `d23bb38c…`). All are THIS bug: the archived copy is the 08-24 backfill being correctly
archived out, not an operator hand-edit to restore. **A batch disposition (status→`complete`, cause
in `result`) was attempted and BLOCKED by this session's permission layer; not retried, not routed
around, no peer asked to run it.** The items stay open; this section is their written answer; the
one-line UPDATE is queued for the owner. noted.co.uk's older sibling (08-18, header slot,
`ab9afa54…`) predates the backfill by six days, no GTM link found — not this bug's.

Banked from the lanes' acks, 2026-08-26 (fuller versions in `analytics_gtm/NOTES` §15):
- **cv1.co.uk (357 lane): rerender measured-safe (6 completed), but a page REBUILD on
  `index`/`tool-example` crashes the agent pod** (`bugs_open/408`, `extractFieldValue` infinite
  recursion → stack overflow, measured ×3). c2's path rides `stale_chrome → needs_rerender` only —
  **nothing in this bug's remediation may set `build_status='needs_rebuild'` on cv1 until 408 is
  fixed.**
- **remortgagecalculator.uk:** served index still tagged; six pre-queued `page_rerender`s can open a
  temporary tag-less window before the keyed chrome rebuild restores it. Rotation ETA ~18 h of 09:20Z.
- **agritec.uk:** own 13-page rerender wave + 17 `needs_imagery` items interleave; dedup absorbs this
  bug's item. `/favicon.png` 404 and `head_essentials_missing` ×13 are pre-existing, not regressions.
- **loancalculator.co.uk:** locked calculators proven wave-safe; post-wave `toolgolden.py` vs
  `GOLDEN_2026-08-24_post_385_repair_tool_values.json` should read 11/11. Mid-wave, one page
  differing from peers = NOT YET REACHED, not skipped — re-sample before concluding a miss.
- **Open OWNER question: should `cv1.co.uk` and `lampenkap.com` report into the estate's GA4?**
  Portfolio call, both lanes agreed; both keyed today; retraction is one supersede (4 + 1 pages).
  > **lampenkap RULED 2026-08-26: "leave lampenkap google tag" — keep the key, no retraction**
  > (`analytics_gtm/CONTRIB_2026-08-26_from_bugs_open_384_owner_ruling_leave_lampenkap_tagged.md`,
  > `d3f04b95a`). **cv1.co.uk remains the open half.** And per the owner's attached question ("is
  > that what GA4 is?"): the ruling keeps the CONTAINER on the site; nothing reports into GA4 until
  > the container carries a published Google Tag — 0 tags at 2026-08-26 10:50Z, re-measured.

### §11 — 2026-09-02: the D bucket refilled to 8 in a week, and the second c2 run cleared it — §6.2 is now the cost centre

Routed back by `agentchassis-33` (their census 20:08Z, independently reproduced here): **8 sites born
since c2 (08-26) with no analytics key** — advertise.co.uk farmerinsurance.uk boxingonline.com
websitepromotion.co.uk seotools.co.uk designblog.co.uk oxenunity.com gamedesign.uk — all network
0002, all deployed, **158 pages**, growth ≈ 1 site/day exactly as §6.2 predicted. Three live lanes
(designblog, gamedesign, boxingonline) notified before firing; **c2 re-applied `-v UNTAGGED=1` at
2026-09-02 20:10:33Z (row stamp)** — INSERT 8, UPDATE 1 (boxingonline's `chrome` key preserved by
the merge), all post-conditions passed. Census after: **A 30 · C 9 (the 8 + adversecreditmortgage,
awaiting their rebuilds) · B 0 · D 0.**

**Consequence: re-running c2 is now a ~weekly manual tax.** Two data points (12+4 sites on 08-26,
8 on 09-02) make the seeder (§6.2) the fix that closes the door, and until it exists the honest
stopgap is the §6.3 detector so the drift is at least seen daily rather than found by routing.
designblog note for the wave: its four empty listing pages re-render empty by design (fill is
upstream, `bugs_open/444`) — chrome timestamps, not content changes.
