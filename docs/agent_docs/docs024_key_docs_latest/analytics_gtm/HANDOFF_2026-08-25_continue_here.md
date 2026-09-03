> **SUPERSEDED 2026-09-03 → `HANDOFF_2026-09-03_continue_here.md`** (this file kept as the arc history — its banner stack is the chronology).

# HANDOFF — analytics / GTM / GA4 · continue here (supersedes HANDOFF_2026-07-31b)

**Written 2026-08-25 16:40 BST, session "google".** This lane is the home for fleet Google tracking
— the owner's 08-25 request (recorded verbatim in
`docs/agent_docs/docs024_key_docs_latest/apis_uk_bees_homepage/HANDOFF_2026-08-25_continue_here.md`
§4a) was spun out "to a new lane", and this is it. Background everyone must read first:
`docs/agent_docs/docs024_key_docs_latest/039_REFERENCE_traffic_and_tracking.md`.

> ## ▶ STATE CHANGED 2026-08-26 — read this block first, the 08-25 one below is history
> The rotation (re-enabled 08-25, `bugs_open/401`) stripped GTM from **10 artefact-only sites
> overnight** (7/7 keyed sites kept it, 10/10 unkeyed lost it — 397 §10). **c2 IS APPLIED**
> (2026-08-26 10:12:11Z by the rows' own stamp — an earlier ~10:50Z here was estimated, corrected — owner "please carry on", 17 sites, 334 pages, all post-conditions passed):
> census now **A 16 durable · C 15 spec-only awaiting their per-site `stale_chrome` rebuild · B 0 ·
> D 0**. The container still has **0 tags** — publishing is still the owner's click. ⚠ Until the
> rebuilds land, **apis.uk serves `gtm=0` — do NOT use it for the walkthrough's step D**; test
> Realtime on a durable site (idea.uk, vonc.com, loancalculator.co.uk). apis.uk's own rebuild page
> item was expected to FAIL — **CORRECTED 12:30 BST: disproven by the apis.uk lane, three
> page_rerenders completed there overnight (that is how its tag was stripped); expect its wave
> items to COMPLETE**; the apis.uk
> lane settles its `build_status` themselves and has been told c2 ran. Verify drain with
> `scripts/check_gtm_state.sh --db` (C → A) and at served bytes.

> ## ▶ ONE-LINE STATE `[ALL MEASURED 2026-08-25 15:00–16:20 BST]`
> The snippet is on 26 sites; **the container it loads has 0 tags — nothing is recorded, nothing
> ever has been.** 12 of the 26 carry the snippet only in the stored artefact and lose it on their
> next chrome render (one has already: agritec.uk). Fix written and dry-run; **not applied — it is
> a 241-page rebuild wave and the owner picks the moment.** Publishing GA4 is his click; Search
> Console needs his service account. Nothing is mid-flight.

> **OWNERSHIP — owner ruling 2026-08-25 ~17:15 BST, relayed by the apis.uk session:** *"section 4 has
> google in it which is taken by another lane, please communicate to that lane that that is what they
> take and we will take the rest here."* **Everything Google is this lane's:** GA4 publication and the
> consent decision before it · `bugs_open/397` (the c2 rebuild AND the §6.2 structural half) · Search
> Console · **the fleet traffic dashboard script** (039 §2's Cloudflare query batched 8 zones at a
> time, our own curl/headless traffic as its own visible line — offered by apis.uk, never started,
> ours to build or decline) · `docs024_key_docs_latest/039_REFERENCE_traffic_and_tracking.md` (ours to
> keep current). apis.uk keeps the page, per-section subjects, image accuracy A+C, their two deferred
> rewrites; they DROPPED the `RenderFallbackHead` per-site-id build. Their table:
> `CONTRIB_2026-08-25_from_apis_uk_bees_homepage_owner_ruling_you_take_everything_google_we_keep_the_rest.md`
> (this dir). **Standing obligation: when c2 runs, tell apis.uk** — their index page refuses
> page-level re-renders (measured: `failed`, `result={}`), so the wave's page item there fails by
> design and they settle `build_status` themselves. 397 §9.

> **2026-09-02:** D refilled to 8 in a week (~1 unkeyed new site/day); second c2 run at 20:10:33Z
> cleared it (census A 30 / C 9 / B 0 / D 0). 397 §11. **Until the §6.2 seeder is built this is a
> weekly manual tax — the seeder + §6.3 detector are this lane's next build.** Customer container
> `GTM-TH5XGNQ4` created and verified same day; estate publish still pending (0 tags, 19:57Z).

> **2026-09-02 ~20:11Z — GA4 IS LIVE.** Owner published: `GTM-PQ3WCTBD` v3, one Google Tag →
> **`G-Y26N29T4KH`**. History begins here; consent position changed in fact (cookies on ~30 sites,
> no banner — the open compliance item). Customer container `GTM-TH5XGNQ4` verified still empty
> (the control). ⚠ the check script's parser was fixed the same minute (WRONG_CALLS 2026-09-02) —
> pull before trusting an old copy. Remaining: §6.2 seeder+detector build (mine), 11 review rows,
> cv1 membership, Search Console SA, consent banner.

> **2026-09-02 late — CONSENT (Option A) SHIPPED to the templates.** Owner chose A; **STY-060**:
> Consent Mode v2 all-denied + self-contained banner, inside the GTM gate, before the loader, in
> all three head templates at **20:55:43Z**; proven 26/26 in headless Chromium against the live
> container BEFORE applying (`consent/consent_block.html`, `sql/c3_consent_mode_banner.sql`,
> NOTES §28). Reaches each site via its stale_chrome rebuild — watch `check_gtm_state.sh --sites`
> (`consent=` column; also `gtm=`). Fail-safe: broken banner ⇒ consent stays denied. **OWED:
> per-site cookie/privacy policy pages (phase 2, through the framework) + the §6.2 seeder.**

> **2026-09-03: the §6.2 SEEDER IS BUILT (STY-061, `fe7359158`, Council-Submitted 45ae3ad3…).**
> New sites get `{analytics: {gtm_container_id: <network default>, mode: default}}` at creation;
> network value opt-in; `mode:none` honoured everywhere (c2 predicate added `90c787355`). **Go half
> inert until the next roll; migration 733 sets the Default Network value — check
> `schema_migrations` for it before assuming.** Consent wave: 22/38 heads overnight, live
> behavioural test on noted.co.uk 5/5 (NOTES §29). Owed: §6.3 detector, policy pages, and READ the
> council verdict for 45ae3ad3.

## 1. Verify before doing anything — one command

```bash
docs/agent_docs/docs024_key_docs_latest/analytics_gtm/scripts/check_gtm_state.sh --all
```
Section 1 reads the live container (the only artefact that can say whether GA4 is on); 2 curls every
deployed domain (our own traffic — Cloudflare counts it); 3 is the durable/artefact-only census.
Expected today: `NOT PUBLISHED, version 2, 0 tags` · 24/25 serve · A 14 / B 12 / D 4 / E 1.

## 2. The three owner actions, in order

1. **Publish the GA4 tag** — his walkthrough is §4a of the apis.uk handoff above (A: Measurement
   ID from the *Agent Chassis* property's data stream; B: a **Google Tag** — not "GA4 Event" — with
   that id, trigger All Pages; C: **Submit → Publish**; D: Realtime; E: delete any tag carrying a
   different `G-` id). ⚠ 039 §4a: this is a **change of compliance position** — the estate sets no
   cookies today because the container is empty; publishing sets `_ga` on ~24 sites with no consent
   banner anywhere. Decide that deliberately, before step C.
2. **Time the durable-tag rebuild** (`bugs_open/397`, §3 below) — "now", "tonight", "after the 130
   queued drain", and whether to include the 4 untagged new sites in the same wave.
3. **Search Console** — a Google Cloud service account with Site Verification + Search Console APIs.
   There is **no Google credential anywhere on our side** (checked `~/.config`, `gcloud`, cluster
   secrets) so nothing can start until he supplies one. Then 039 §5's automation shape.

## 3. bugs_open/397 — what it is and what to run

**What:** head templates emit GTM inside `{{if .gtm_container_id}}`; the id comes from
`site_specs.site_config.analytics.gtm_container_id` (STY-050, this lane, 07-31). The 08-24 backfill
wrote `site_components.rendered_html` on 13 sites and no key. Chrome regenerates on `nav_rebuild`
etc.; no key → gate false → tag gone → `chrome_divergence_overwritten` filed as if an operator
hand-edit had been corrected. agritec.uk, 08-24 19:20:53, both instruments.

**The fix, when the owner says when:**
```bash
L=docs/agent_docs/docs024_key_docs_latest/analytics_gtm
P="kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db"
$P -v DRY=1 -v GO=yes -f - < $L/sql/c2_gtm_spec_key_for_artefact_only_sites.sql   # look first
$P          -v GO=yes [-v UNTAGGED=1] -f - < $L/sql/…c2…sql                          # apply
```
Then watch `needs_rerender` items with `item_key='stale_chrome'` appear per site and complete; bucket
B reads 0 at once, served `gtm=2` only after the rebuilds and the deploy queue land. **Tell the lanes
on those sites first** (397 §9: loanzy, webdesign_uk_build_service, bugfix_357/384, agritec).

**The structural half — not started, council scope, REQUIREMENTS NOW SET (2026-08-26):** customer
builds default to the owner's container on the hosted copy only, per-site override, ZIP clean
(`webdesign_uk_build_service/DECISION_2026-08-26_default_tag_hosted_copy_only.md`; design notes in
397 §6.2 incl. the `analytics.mode` field and the one-place default). ⚠ ~~And a collision for the
owner's consent decision~~ **RULED 2026-08-26 night: a SECOND cookie-light container for customer
sites — estate GA4 publication into `GTM-PQ3WCTBD` is unblocked.** Creating it is THIS lane's task,
blocked on access only (owner's ~2 min in the GTM dashboard — walkthrough in README — or the same
service-account credential Search Console needs). Its id = the one-place customer default. ⚠ Empty
= records nothing: spec the Consent-Mode-denied GA4 tag when it publishes (397 §6.2).
**DONE 2026-09-02: the owner created it — `GTM-TH5XGNQ4`, verified live 20:01Z (v1, 0 tags, no
`G-` id — 0 tags is CORRECT for this one).** Estate `GTM-PQ3WCTBD` still 0 tags same read — the
owner's publish click remains the one outstanding Google action. Previously: no Go writer touches `site_config`, so a new
site is born without the key. Where the opt-in gets set for our-network sites, and how a
handed-over (third-party) site is guaranteed never to get it, is the design question. 397 §6.2.

## 4. Hard-won specifics — do not re-derive

- **Nothing you can see on a page tells you whether GA4 is on.** Only `gtm.js` does: `"tags":[]`
  = nothing. Three lanes said "live" off the snippet. `WRONG_CALLS.md` 2026-08-25.
- **A value in `rendered_html` is a cache, not a setting.** LANDMINES 2026-08-25 (this lane). Write
  the spec the schema names; expect the rebuild; merge into the existing row.
- **Writing `site_config` IS firing a rebuild** — `ChromeRenderInputsSQL` fingerprints it;
  `stale_chrome` dispatches (20 ever, all complete). Size it in pages before the first UPDATE.
- **`c1` (07-31) replaced `site_config` wholesale** and dropped relojistas.com's `intent_probe`.
- **`RenderFallbackHead` is the failure path** — never the place for a tag.
- **`webdesign.uk` 302s to `webdesign.co.uk`**; curl with `-L` or it reads `gtm=0`.
- Everything in HANDOFF_2026-07-31b §"Hard-won specifics" still holds (idea.uk is two applications;
  `input_schema` has two shapes; `webdesign.co.uk Document Head` lowercase charset; one page-rerender
  = one commit = one Actions run).

## 5. Cross-lane, this session

- `apis_uk_bees_homepage` ← `CONTRIB_2026-08-25_from_analytics_gtm_the_backfill_is_artefact_only_and_the_per_site_seam_exists.md`
  (their backfill is 397; their §3 per-site-id design is superseded by STY-050).
- `agritec_uk` — the `chrome_divergence_overwritten` head row on their site is 397, not a hand-edit.
- `dartsonline_traffic` — their LANDMINES 2026-08-25 entry carried "a GA4 tag was added the evening of
  2026-08-25"; corrected in place, dated.
