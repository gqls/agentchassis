# HANDOFF — analytics / GTM · continue here (supersedes HANDOFF_2026-07-31)

**Written 2026-07-31, late afternoon.** All three phases are APPLIED. What remains is a
**deploy queue draining on its own** and one open owner decision (consent).

---

## State in one paragraph

GTM container **`GTM-PQ3WCTBD`** is applied estate-wide: **14/14 deployed domains** carry
the gated tag in both their `head` and `header` chrome artefacts and in `site_specs`, and
the competing gtag.js seam is retired. idea.uk's separate Go tool binary is deployed and
serving the tag on all its pages including both conversion pages. All **377** pages were
re-rendered successfully. The only outstanding *mechanical* item is that ~221 GitHub
Actions deploy runs were still queued at the time of writing, so **5 domains may still be
serving pre-GTM bytes until that drains (~100 min from ~13:05 UTC)** — it is self-healing
and needs no intervention, only a re-check.

## Verify current reality before doing anything (things may have drained)

```bash
# per domain, homepage: expect 2
for d in ai-agent-orchestration.com dartsonline.com finetuning.uk fundamentallyai.com \
         gamesdesign.co.uk gaswholesalers.com idea.uk leopardessconsulting.co.uk \
         oufe.com relojistas.com robot-hands.com vetcomparison.uk vonc.com webdesign.co.uk; do
  printf '%-30s %s\n' "$d" "$(curl -s --max-time 25 https://$d/ | grep -c googletagmanager)"
done
# the deploy queue that gates the last few
gh run list --repo gqls/sites --status queued --limit 400 | wc -l
```

**If a domain reads 0, check `last-modified` before suspecting the tag.** Bytes older
than 2026-07-31 mean the deploy has not landed yet, not that the render failed — the DB
side is provably correct (see below).

## What is done, with the evidence

| phase | state | evidence |
|---|---|---|
| **A** idea.uk static (20 pages) | ✅ live | verified live; survived the 07-31 chassis roll |
| **B** idea.uk tool binary (11 pages) | ✅ **deployed 07-31** | `/order/success`, `/order/cancel`, `/privacy`, `/terms`, `/refund-policy` all `script=1 noscript=1 body_adjacent=yes`; `/health` shows `StripeProvider`; `/capacity` unchanged |
| **C** DB, all 14 domains | ✅ applied | `sql/c1_gtm_fleet_rollout.sql` post-conditions: 14/14 head artefacts, 14/14 headers with noscript FIRST, 14/14 specs, 9/9 templates gated, `analytics_id` 0 matches |
| **C** page re-render | ✅ 377/377 | 0 publish failures, 0 FAILED orchestrations, 377/377 artefacts carry both tags |
| **C** deploy to B2 | ⏳ draining | 77 done / 221 queued at 13:05 UTC; ~2/min |
| **D** account structure | ✅ decided | one container estate-wide; `analytics_id` retired |

**Commits:** `274785024` (Phase A + B code), `c8fe9bb5f` (STY-050), `27bb75322` (seam
correction), plus this session's rollout commit.

**Rollback:** `bak_gtm_fleet_20260731_site_components` / `_content_components` (DB);
`/opt/idea/idea.prev-2026-07-31-123615-pre-gtm` (binary, plus an `orders.json` backup).

## The one thing still needing the owner

**Consent.** These are UK-facing sites; analytics cookies now fire on 14 domains and
there is no consent banner anywhere. Adding GTM did not create the exposure but it has
made it concrete and estate-wide. GTM Consent Mode v2 is the mechanism and it needs a
banner to feed it. This was flagged before the rollout and the owner chose to proceed;
it remains open and is the natural next workstream.

## Hard-won specifics — do not re-derive these

- **idea.uk is TWO applications.** nginx proxies 16 routes to a Go binary
  (`docs024/idea.uk/golang_files/`, live at `/opt/idea/idea` on 116.203.204.115). Its 11
  HTML pages go through one `App.page()` wrapper and are NOT in the static build. The
  static `/privacy.html|/terms.html|/refund-policy.html` **301** to the binary's copies.
  Tag is env-driven: `GTM_CONTAINER_ID` in `/etc/idea/idea.env`.
  ⚠ **systemd EnvironmentFile keeps inline comments** — own line only, or the service
  restart-loops into an nginx 502.
  ⚠ After any restart, check `/health` shows `"provider":"*main.StripeProvider"` — a
  missing Stripe var silently downgrades it to `FakeProvider`.
- **`input_schema` has two live shapes**, and `header-theme-chrome`'s is **NULL**. Flat
  ⇒ descriptor at top level; wrapped ⇒ under `fields`. Wrong path = the gap-fill never
  sees it, silently. A wrong jsonb path **returns NULL, not an error**.
- **`webdesign.co.uk Document Head` uses lowercase `<meta charset="utf-8">`** — the
  uppercase UPDATE hit 12 rows, the lowercase one hit 1. And that component emits **no
  `<head>` open tag at all**; its 99 pages have an implicit head. GTM works there, but the
  missing tag is a real pre-existing defect, deliberately not fixed (blast radius).
- **`pages.deployed_at` is not reliably refreshed** — `contact` on vetcomparison still
  reads 2026-07-18 while serving new chrome. Never use it to decide what shipped.
- **`b2 sync --delete` is not concurrency-safe** on one prefix: two runners racing gave
  `FileNotPresent … Incomplete sync` and a red workflow. Self-heals on the next sync.
- **One page-rerender = one commit = one Actions run.** 377 pages queued ~230 redundant
  whole-directory syncs on a 2-slot self-hosted runner. **Next time, decouple render
  fan-out from deploy fan-out:** the workflow syncs *every* domain when its `CHANGED` list
  comes back empty, so a single commit touching only a root-level file triggers one
  full-estate deploy instead of hundreds.
- **Never publish via `kubectl run -i … kcat -P`** (loses ~4 of 5 at exit 0). Use
  `scripts/fire_reassemble_site.sh <domain> [parallel]` — payload in the container
  COMMAND, `--command` required, every publish prints `PUBLISH_OK`.

## Corrections carried forward

- **STY-050's "first real consumer of the schema-driven fill" claim is WITHDRAWN.**
  `head-seo-standard` had the same pattern from 2026-05-13 (gtag.js at
  `config.analytics_id`). Now retired. Logged in `WRONG_CALLS.md` 2026-07-31, with the
  general lesson: **prior-art greps over prose miss config-only mechanisms — query
  `content_components`, not just the docs.**
- `tool-audience-check` on idea.uk is a **stub** (0 sections, never deployed), correctly
  excluded from re-renders. Not a missing page.
