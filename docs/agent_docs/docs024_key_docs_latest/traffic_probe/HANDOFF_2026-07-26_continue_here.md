# HANDOFF — continue from here (26 July 2026)

**Read this first in a fresh chat.** It covers two workstreams that this session left in
a clean state: **relojistas.com** (build list empty, one owner action outstanding) and
**vm_estate** (new, design settled, nothing built). Everything marked ✅ was fetched or
queried live on 2026-07-26; anything not verified is marked.

Repo-root `CLAUDE.md` binds: pathspec commits, forward-only, build-from-committed-HEAD,
pod-verify deploys, council gate advisory, the standing five docs, bugs_open/bugs_closed.

---

## 0. Sixty-second orientation

| | relojistas.com | vm_estate |
|---|---|---|
| state | **finished & self-running**; build list empty | **design only, nothing built** |
| blocked on | **owner's ONE box session** (§2) | nothing — next step is free (§5) |
| docs | `docs024_key_docs_latest/traffic_probe/` | `docs024_key_docs_latest/vm_estate/` |
| entry doc | `HANDOFF_RESUME_relojistas_rebuild.md` (older, still accurate) | `PLAN_2026-07-25_framework_controlled_vm_estate.md` |
| latest read-out | `SUMMARY_2026-07-26_relojistas_rebuild.md` | `SUMMARY_2026-07-25_vm_estate.md` |

Site id `ecf15e75-a966-4900-bcb0-1c85f689dbfd`. Box `167.233.33.159` (Hetzner, root SSH,
key-only). DB: `kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`.

---

## 1. relojistas — verified live state (2026-07-26) ✅

```bash
curl -s https://relojistas.com/feed.xml | grep -c "<item>"                     # 30
curl -s https://relojistas.com/feed.xml | grep -o "<lastBuildDate>[^<]*"       # 26 Jul 13:49 UTC
curl -s https://relojistas.com/ | grep -c 'href="/ferias\|href="/archivo'      # 0 phantoms  <- SEE CORRECTION
curl -s https://relojistas.com/ | grep -c 'mailto:'                            # 0           <- SEE CORRECTION
curl -s https://relojistas.com/ | grep -c 'action="/intent"'                   # 1 (search box)
curl -s https://relojistas.com/ | grep -c '<article'                           # 12 (server-rendered)
```

> **CORRECTED 2026-07-27 — lines 3 and 4 above are vacuous checks. Do not reuse them.**
>
> **"0 phantoms" greps for the two phantoms we already fixed**, so it returns 0 forever —
> including on a page full of different ones. A real sweep of all 19 deployed pages found
> **`/contact.html` → 404 linked from three of them, and it is the homepage's primary hero
> button** (`<a href="/contact.html">Leer las últimas noticias</a>`). Also `favicon.png` 404s
> on all 19. Use this instead — it follows every internal href and probes each target:
> ```bash
> curl -s "https://relojistas.com$PAGE" | grep -o 'href="/[^"]*"' | sed 's/href="//;s/"$//' \
>   | grep -v '^/cdn-cgi' | cut -d'#' -f1 | sort -u | while read -r h; do
>       printf '%s %s\n' "$(curl -s -o /dev/null -w '%{http_code}' "https://relojistas.com$h")" "$h"; done
> ```
> **"0 mailto" is vacuous on any Cloudflare-proxied site** — CF rewrites every mailto into
> `/cdn-cgi/l/email-protection#<hex>`, so the grep can never match. Decode (XOR each byte after
> the first against the first) before believing it: the homepage does carry
> `relojistas@contactforsales.com`, and the footer contact anchor is **empty**.
>
> Cause and full evidence: running notes 2026-07-27; contributed to `bugs_open/071` (slug
> `validate_gate_detects_every_broken_link_then_discards_the_finding`) as a sighting in
> `b96acad7d`. It is a **platform default**, not a writer invention — `component_library.go`
> fills an empty `cta_url` with a hardcoded English `/contact.html`. **Owned by
> `brochure_component_library`; do not start a competing fix.**

- Pipeline **runs unattended** — feed rebuilt today with no human involvement.
- Corpus 72 relevant + 5 review, newest 2026-07-26 07:56. 18 pages deployed.
- **The homepage-rewriting fix HOLDS**: 20 `page_rerender` complete (newest today
  13:49); last `source='render_news_section'` `needs_page` was **2026-07-24 13:45**,
  i.e. before v1.0.1155 rolled. Nothing since.

### These four are EXPECTED failures, not bugs

```bash
curl -o /dev/null -w "%{http_code}\n" "https://relojistas.com/external.php?type=RSS2"  # 200 ✅
curl -o /dev/null -w "%{http_code}\n" "https://relojistas.com/external.php?type=rss2"  # 404 ← §2 fixes
curl -o /dev/null -w "%{http_code}\n" "https://relojistas.com/external.php"            # 404 ← §2 fixes
curl -o /dev/null -w "%{http_code}\n" "https://relojistas.com/buscar?q=rolex"          # 404 ← §2 fixes
```

They are the *measure* of the pending owner session. **Do not "fix" them another way** —
the repair is already written into the generator and waiting.

---

## 1b. 2026-07-27 — owner ruling: NO contact route; about page takes the for-sale block

**Applied.** relojistas carries no contact surface. `contacto` archived **and its file
deleted from `vm-sites`** (`c74fe58`) — archiving alone leaves it serving (`bugs_open/098`);
`/contacto.html` → 404. Its nav row set `inactive` (nav lives in **`site_nav_items`**, NOT
derived from `pages` — archiving does not touch it; contributed to 098). Hero CTAs on
`index`/`noticias-index`/`glosario-index` given real destinations; `identity.contact.phone`
cleared (an inherited `+34` number from the dead forum); the `needs_section_data` request
for a business email closed `wont_fix`.

**VERIFIED LIVE 2026-07-27 17:45 — 18 of 18 pages, and a full probe of every internal link:**

```
18 pages: 0 href="/contacto.html"   0 href="/contact.html"   0 hero "Get Started"
/contacto.html -> 404 ;  feed.xml still 30 items
every internal target probed -> only 2 non-200, BOTH pre-existing and unrelated:
   404 /assets/images/favicon.png   (long-known gap)
   301 /guias/mantenimiento          (harmless redirect)
```

**Getting the footer to change needed a SITE-CHROME rebuild, not a page re-render** — chrome
is a stored artefact (`site_components.rendered_html`), frozen here since 07-16, used verbatim
by `assemblePage`. Trigger: a `nav_drift` work item handled by `nav-updater`
(→ `render_site_components`). See `bugs_open/117`. **Then re-render the pages.**

**The contact column is GATED but still renders, correctly** — `sites.email` holds
`relojistas@contactforsales.com` (a 13-site house convention; NOT
`site_specs.identity.contact`, which is null). The gate correctly dropped the empty phone
line. **OWNER DECISION OPEN:** that address is a contact route on a site ruled to have none —
but it is a *sales* address on a domain that is for sale, and clearing it edits a fleet
convention. `bugs_open/111`'s "8 sites show an empty block" claim is **refuted** (corrected
in place). Gate applied to BOTH `footer-theme-chrome` and `footer-4-column` because chrome
selection has no `is_active` filter and picks alphabetically — the deactivated
`footer-4-column` is what actually renders fleet-wide (`bugs_open/118`). Backups:
`bak_footer_theme_chrome_20260727`, `bak_footer_4column_20260727`.

**Two traps burned in, both worth more than the task:**

1. **You cannot remove a CTA by clearing its data.** Deleting `cta_text` made
   `component_library.go` refill *both* fields, and the refilled pair then **satisfied** the
   template's `{{if and .cta_text .cta_url}}` guard — shipping English **"Get Started" → 404**
   on a Spanish site, worse than before. Every hero must carry an explicit `cta_url` forever.
2. **`page_rerender` only regenerates from `content_data` when `spec.reason` is stamped**
   (`section_data_resolved` / `image_landed`); reason-less is assemble-only and redeploys
   stale HTML. Precondition: any section with NULL `content_data` escalates the whole page
   to the **LLM writer** — check before firing.

**For-sale block: CONFIGURED, deliberately NOT rendered.** `site_specs.commercial` =
`portfolio / tier 2 / for_sale_requested / marketplace_url=forsale.godaddy.com/forsale/relojistas.com`.
Nothing renders it: `about-commercial-block` has **zero Go references** (Phase-2 default-set
unbuilt) and no section was inserted. **Blocked on language** — the block is hardcoded
English and this site must be *íntegramente en español*; raised with the owning workstream
(`about_page_commercial/NOTES`, `7c0ee759c`). **[UNVERIFIED]** the Afternic listing is the
owner's dashboard reading — GoDaddy 403s automated probes. Its **Minimum Offer is 0**, so the
design's anti-lowball floor is absent.

---

## 2. THE ONE OUTSTANDING ACTION — the owner's box session

Only the owner can do this (it needs the box). It converges four standing items at once.
Full procedure: `relojistas_rebuild_runbook.md` § "owner convergence run".

1. **scp the reconciled `setup.sh`** to the box and re-run it (it is idempotent; the box
   no longer has a copy). Source of truth:
   `traffic_probe/deploy_setup/vm-deploy/setup.sh`.
   Installs: `/buscar` route · legacy `/external.php` + the 3 failing variants ·
   `/events` (the collector needs it) · **Cloudflare real-ip** (`CF-Connecting-IP`).
   The rendered conf was verified a **strict superset** of the live conf, so the
   surgical hand-edit currently on the box is not lost.
2. **Append two lines** to `/etc/site-engine/site-engine.env` then
   `systemctl restart site-engine` — verified absent today, so this is still needed:
   ```
   WEBROOT_DIR=/var/www/vm-sites
   RESULTS_PATH=/buscar
   ```
   (The script deliberately never overwrites this file — it holds `INTERNAL_API_KEY`.)
3. **Cluster side, after `/events` answers**: retarget the collector task to
   `intent-collection-orchestrator` and `SET enabled=true`; verify rows land in
   `intent_events`.

### Verify afterwards (the before/after is quantified)

```bash
# all four should now answer:
for u in "external.php?type=RSS2" "external.php?type=rss2" "external.php" "buscar?q=rolex"; do
  curl -s -o /dev/null -w "$u %{http_code}\n" "https://relojistas.com/$u"; done
# real IPs, not 104.x/172.x Cloudflare edges:
ssh root@167.233.33.159 'tail -5 /var/log/nginx/access.log'
```

Residual failures were **~5/day in exactly 3 shapes** (bare path 25, lowercase
`type=rss2` 11 — the box's hand-edit is case-**sensitive** — `/ventas/` 4, over 8 days).

---

## 3. relojistas — what is CLOSED, so nobody reopens it

- **Per-forumid category feeds: DEFERRED on evidence, not forgotten.** Board-param feed
  traffic is ~88% self-identified crawlers with **zero conditional GETs**; the only
  real-subscriber signal (42 × `304`) is on the **bare** feed URL. The recorded "8
  subscribed forumids" was an unchecked sample of a real 123. The corpus cannot fill
  those boards anyway (Seiko 1 item, Louis Erard 0, Sorteos 0).
  Evidence + method + 123-board table: `EVIDENCE_2026-07-25_legacy_board_feed_demand.md`;
  decision + ready-to-build design: `relojistas_rebuild_plan.md` §P8.
  **Reversal trigger (check it after §2):** if board-param requests then show distinct
  real IPs or conditional GETs, there is someone to serve and the design is written.
- **Reactivation MEASURED:** 404-only until **17 July**, 200s since (26 Jul: 30-item feed
  serving). Distinct *people* remain uncountable until real-ip lands — §2.
- **027 (no-JS news), the emitter mis-route, P5.2 (search) — all closed.** See
  `SUMMARY_2026-07-24` and `_2026-07-25`.

## 4. Corrections made to my own record (do not re-derive these)

1. **The `needs_page` rows are NOT a regression.** 6 of them are `operator:bugfix_028` —
   another session's glosario repairs (verified live and good: `/glosario/calibre.html`
   now has its own headline). A `GROUP BY` total is not evidence about a cause; read the
   `source` column.
2. **The suppressor rows never "vanished".** They are present at `status='failed'`; I
   searched for `status='blocked'`. They are inert — `failed` is in `idx_swi_dedup`'s
   terminal list, so they hold no dedup key, and their key namespace
   (`page_rerender:index`) predates the live emitter's (`page_rerender_<page>_<site>_<reason>`).
3. **"Subscribed forumids" was requests miscounted as people** — logged in
   `WRONG_CALLS.md` with the one-column check that would have caught it.

---

## 5. vm_estate — the new workstream (nothing built)

**Goal (owner):** bring `setup.sh` under framework control and **merge** the three boxes
rather than run three projects. Read `PLAN_2026-07-25_framework_controlled_vm_estate.md`
— Part 1 is a full walkthrough of the script, Part 2 the argument, Part 3 the merge.

**The framing that makes it click:** the box's nginx conf is to a machine what
`rendered_html` is to a page — the same defect contract doc `003` already ruled on. A
hand-edit to the live conf on 07-19 would have been destroyed by the next generator run;
it was reconciled by hand on 07-24. That manual repair is exactly what the platform
refuses to accept for pages.

**Reuse, do not invent** — the primitives exist and are in production for GPU training:
`dispatch_thunder_provision` → `dispatch_thunder_ssh_exec` → `dispatch_thunder_ssh_status`,
agents `gpu-provisioner` / `training-launcher` / `thunder-training-monitor`. The renderer
has a precedent too: `render_rss_feed` reads DB rows and emits a file artefact.
**`service-deployer` does not exist** despite `setup.sh`'s header promising it for months.

**DECIDED by the owner 2026-07-25: the island pulls, outbound-only.** Merge the
*generator*, not the trust boundary — the island's rationale is that the production
cluster appears nowhere in its path and it holds no production credential; naive push
control would invert exactly that.

**Merge order (reasons attached):** relojistas → idea.uk → island.

**NEXT STEP, and it is deliberately free:** P1 describe the relojistas box as DB state,
then **P2 render its config and diff byte-for-byte against the live conf**. Read-only,
touches nothing, and a renderer that reproduces a live box *is* the proof the description
is complete. Do not skip to applying.

**Still open for the owner** (PLAN Part 5): provider spread (Hetzner + Mythic Beasts or
consolidate); whether *ordering* hardware is in scope (plan stops at "configure"); where
rendered artefacts live (vm-sites repo vs a separate infra repo).

**Defect B is deliberately unpatched:** `static_body()` emits relojistas' legacy-feed and
`/buscar` locations for **every** domain on the box (latent — one domain there today). A
bash conditional would entrench the design being replaced; it belongs in per-site DB state.

---

## 6. Landmines (all paid for at least once)

- **`grep` prints NOTHING on non-UTF-8 files** (archived/ISO-8859 pages) in a UTF-8
  locale — no error, no warning. `file` it, then `LC_ALL=C grep -a` + `iconv`.
- **Scope log greps to the endpoint before extracting params** — a bare `forumids=` grep
  also sweeps crawlers walking old thread URLs.
- **A 404 spike may be pre-fix history in the same retained log file.** Split by day
  (numerically — the natural sort is alphabetical on `01/Jul`) before disbelieving a
  proven metric.
- **Committed code rides ANYONE's next build** — commit implementation only when you are
  content for it to ship; the council trailer carries review status, not the commit.
- **Park suppressors BEFORE repairing the data they protect** — a feed cycle destroyed a
  repair in the gap.
- **Council sketches must be FINAL-state**, not intermediate commits (cost two rounds).
  Stop after ~3 REVISE rounds when substantively approved (the 053 precedent).
- **idea.uk/relojistas are served from a VM** (sitesync ~5 min), so live HTML lags a
  chassis publish — do not read a stale fetch as a failed deploy.
- A missing orchestration row is usually **queue latency (~16–30 min)**, not a dropped
  dispatch. Do not resubmit on that evidence.

## 7. Doc map

```
traffic_probe/
  HANDOFF_2026-07-26_continue_here.md      <- you are here
  HANDOFF_RESUME_relojistas_rebuild.md     <- older entry point, items 4+5 updated
  SUMMARY_2026-07-26_relojistas_rebuild.md <- latest read-out (series: 07-19/24/25/26)
  EVIDENCE_2026-07-25_legacy_board_feed_demand.md
  relojistas_rebuild_plan.md               <- §P8 = the deferral decision
  relojistas_rebuild_runbook.md            <- "owner convergence run" + feed forensics
  relojistas_rebuild_running_notes.md      <- technical log, newest at the bottom
  README_where_we_are.md                   <- owner's plain-prose log (APPEND ONLY)
vm_estate/
  PLAN_2026-07-25_framework_controlled_vm_estate.md   <- walkthrough + design + open Qs
  SUMMARY_2026-07-25_vm_estate.md · NOTES_vm_estate.md · README_where_we_are.md
WRONG_CALLS.md                             <- fleet-wide; my 07-25 entry is in it
```

Commits this session: `d24e4ffa7` (P8 + measurement), `099a2bfee` (vm_estate opened +
setup.sh fix), `09e3666c4` (island decision + first vm_estate summary).
