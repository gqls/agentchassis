# HANDOFF — analytics / GTM · continue here

**Written 2026-07-31.** Cold-start for a new session. Read this, then
`PLAN_2026-07-30_analytics_gtm.md` §5a/§6 and `NOTES_analytics_gtm.md` (bottom entry).

---

## One-paragraph state

The owner asked for Google Tag Manager container **`GTM-PQ3WCTBD`** on every page of
idea.uk, then for the best way to track all the framework's domains. **Phase A (idea.uk's
static site) is DONE and LIVE and survived the 2026-07-31 chassis roll.** **Phase B (the
11 pages served by idea.uk's own Go binary, including both conversion pages) is written,
committed, built and tested but NOT DEPLOYED** — it restarts the live Stripe service.
**Phase C (13 other domains) is blocked** on two owner inputs. Nothing is half-applied:
every phase is either fully live or fully unstarted.

## What is LIVE (do not redo)

- GTM on **20 real static pages** of idea.uk. Script immediately after
  `<meta charset="UTF-8">`; noscript the first thing after `<body>`.
- Verified at the artefact (20/20 rendered) and live (19/19 fetchable; `/privacy.html`
  is a 301 — see below). Re-verified after the chassis roll: 5/5 spot-checks return 2
  hits, `site_components.updated_at` unchanged at `2026-07-30 19:37:05`.
- Applied by `idea_uk_vm_site/sql/p4_34_gtm_container.sql` (transactional, pre-guards +
  post-assertions, backup tables `bak_ideauk_gtm_20260730_*`).
- Blast radius confirmed contained: the 8 other sites sharing `Document Head` return
  **0** GTM hits live.

Commits: `274785024` (Phase A + Phase B code), `c8fe9bb5f` (STY-050 register + gofmt).

## The two decisions blocking everything

### 1. Deploy Phase B? (idea.uk's conversion pages)

**Why it matters more than it sounds.** idea.uk is **two applications behind one
domain**. nginx proxies 16 routes to a Go binary; 11 render HTML through one
`App.page()` wrapper and none are in the static build — including **"Payment received"**
(`/order/success`) and **"Request received"**. Until this ships, Google shows idea.uk
traffic and **zero conversions**, which misreads as "the site doesn't convert".
`/privacy`, `/terms`, `/refund-policy` are also served there, and the static `.html`
copies **301 to them**, so the tag already shipped on those three can never fire.

**Ready to go.** Code committed; `GTM_CONTAINER_ID` env-driven (empty = fully inert);
5 tests in `gtm_test.go` assert *placement*, not presence. Binary already built at
`CGO_ENABLED=0 GOOS=linux GOARCH=amd64` (9,909,585 bytes; current live binary is
9,999,070, same shape). **Rebuild it — do not trust a scratchpad copy from a prior
session.**

```bash
G=docs/agent_docs/docs024_key_docs_latest/idea.uk/golang_files
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -C "$G" -trimpath -o /tmp/idea_new ./...
ssh root@116.203.204.115 'cp /opt/idea/idea /opt/idea/idea.prev-$(date +%Y-%m-%d)-pre-gtm
                          cp /var/lib/idea/orders.json /var/lib/idea/orders.json.bak-$(date +%Y-%m-%d)-predeploy'
scp /tmp/idea_new root@116.203.204.115:/opt/idea/idea.new
ssh root@116.203.204.115 'mv /opt/idea/idea.new /opt/idea/idea && chmod +x /opt/idea/idea'
# add the container on its OWN LINE — see the trap below
ssh root@116.203.204.115 "printf '\nGTM_CONTAINER_ID=GTM-PQ3WCTBD\n' >> /etc/idea/idea.env && systemctl restart idea"
# verify the pages that could not be verified any other way
for p in /order/success /order/cancel /privacy /terms /refund-policy; do
  printf '%-18s %s\n' "$p" "$(curl -s https://idea.uk$p | grep -c googletagmanager)"; done   # expect 2 each
```

> ⚠ **`/etc/idea/idea.env`: systemd EnvironmentFile does NOT strip inline comments.**
> `PORT=8080 # the port` makes the port `"8080 # the port"` → exit 1 → restart loop →
> nginx 502. This has already crashed one real deploy. Own line, no trailing comment.

> ⚠ **Check `curl -s https://idea.uk/capacity` first.** It reported `{"active":1}` on
> both 07-30 and 07-31. A restart mid-order is *handled* — the service returns the order
> to `requested`, releases the slot and emails the operator — but it is customer-visible.
> Waiting for `active:0` avoids it entirely.

**Rollback:** `ssh … 'cp /opt/idea/idea.prev-<date>-pre-gtm /opt/idea/idea && systemctl restart idea'`.
Twelve prior `.prev-*` binaries are already on the box; this is routine there.

### 2. Which container(s), and which seam survives? (Phase C)

**Cannot proceed without the owner.** `GTM-PQ3WCTBD` was given *for idea.uk*. Rolling it
to 13 more domains puts 13 businesses in one container — that is the recommendation
(§6: one GA4 property + one container, per-domain reporting off the hostname dimension)
but it is the owner's call, and per-domain containers do not exist yet.

**And there is a second, pre-existing analytics seam that conflicts** — found 07-31:

| seam | component | domains | mechanism | state |
|---|---|---|---|---|
| `config.analytics.gtm_container_id` | `Document Head` | 9 | GTM | live, idea.uk only |
| `config.analytics_id` | `head-seo-standard` | 4 | **gtag.js / GA4 direct** | **dormant, 0 sites** |
| — | `webdesign.co.uk Document Head` | 1 | neither | — |

A site carrying both loads GA4 directly **and** through GTM → **double-counted
pageviews**. Recommendation: GTM only; retire `analytics_id` while it is still dormant
(0 sites ⇒ removal changes no rendered byte).

## How to do Phase C when unblocked

Per domain, the same five steps as `p4_34` (read that file — it is the worked template):

1. `site_specs` aspect `site_config` ← `{"analytics":{"gtm_container_id":"<ID>"}}`.
   Supersede any existing current row first (`idx_site_specs_current` is UNIQUE on
   `(site_id, aspect) WHERE is_current`).
2. Head template: gated `{{if .gtm_container_id}}` block after the charset meta, plus a
   **map-valued** `input_schema` field `source: config.analytics.gtm_container_id`.
3. Header template: gated noscript block **prepended** (it renders directly after
   `<body>`, which is a Go literal in `assemblePage:577`).
4. Patch **both** `site_components.rendered_html` artefacts — template-only is inert
   (`bugs_open/117`), artefact-only is reverted by the next chrome rebuild.
5. Re-assemble every page, then verify at the artefact and live.

Only **3 head + 6 header components** cover all 14 deployed domains, so steps 2–3 are
9 template edits total, not 28.

### Four traps that make Phase C silently no-op

1. **A scalar `input_schema` entry is skipped** as "not a field descriptor"
   (`render_site_components_action.go:612-615`). Must be map-valued.
2. **Two schema shapes exist** — flat `{name:{…}}` and wrapped `{"fields":{name:{…}}}`
   (`:604-607`). `Document Head` is flat; `head-seo-standard` is wrapped. **Querying the
   wrong jsonb path returns NULL, not an error** — it made a working seam look
   undeclared for me on 07-31. Check both.
3. **`webdesign.co.uk Document Head` uses lowercase `<meta charset="utf-8">`.**
   `replace()` is case-sensitive → 0 rows updated, reported as success. Guard per site.
4. **Never publish re-render messages via `kubectl run -i … kcat -P`** — loses ~4 of 5
   at exit 0. Use `scripts/fire_reassemble_idea_uk.sh`'s pattern (payload in the
   container COMMAND, `--command` required, every publish prints `PUBLISH_OK`).

## Verification commands that actually prove something

```bash
# live, per URL, WITH the status code — a bare hit-count hides the 301 case
curl -s -o /tmp/p -w '%{http_code}' https://<domain>/<path>; grep -c googletagmanager /tmp/p
```
```sql
-- artefact, not status: COMPLETED is not proof the render carried the tag
SELECT count(*) FILTER (WHERE (collected_data->'render_page'->>'html') LIKE '%gtm.js%')
FROM orchestration_states WHERE correlation_id IN (…);
-- blast radius: nobody unexpected should appear here
SELECT s.domain, sc.slot_name FROM site_components sc JOIN sites s ON s.id=sc.site_id
WHERE sc.rendered_html LIKE '%googletagmanager%';
```

## Also outstanding (owner-facing, not blocking)

**Consent is a real gap.** UK-facing sites, analytics cookies, no banner on any domain.
Adding GTM did not create it but does make it concrete; Consent Mode v2 needs a banner to
feed it. Worth settling before Phase C widens it to 13 more domains.

## Corrections carried forward — do not re-derive

- STY-050's "first real consumer of the schema-driven fill" claim is **withdrawn**;
  `head-seo-standard` had the same pattern from 2026-05-13. Correction is inline in the
  register entry and logged in `WRONG_CALLS.md` (2026-07-31).
- `tool-audience-check` is a **stub** (`/tools.html#audience-check`, 0 sections,
  `deployed_at` NULL) and is correctly excluded from re-renders. It is not a missing page.
- The gofmt complaint on `service.go` was **pre-existing import ordering**, fixed in
  `c8fe9bb5f`. `engine.go` is also unformatted and deliberately untouched.
