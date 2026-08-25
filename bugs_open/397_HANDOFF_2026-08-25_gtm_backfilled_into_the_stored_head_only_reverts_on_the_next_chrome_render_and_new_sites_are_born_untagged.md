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
2. **Give new sites the key at birth — the structural half, and an open design question.** No Go
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
