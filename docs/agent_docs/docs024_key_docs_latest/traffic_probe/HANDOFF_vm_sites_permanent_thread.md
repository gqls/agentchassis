# Handoff — permanent/scalable VM-hosted-sites thread (NOT P4)

This chat is carrying **P4** (off-box collection: `intent_events` + the
collector). This document hands off the OTHER side — turning the probe from a
hand-made one/two-page pattern into a first-class, framework-built, scalable
class of VM-hosted sites — plus the immediate relojistas manifest. Start a new
chat from this file.

Read first: `traffic_probe_plan.md` (esp. "How it all fits" + P3/P5 +
Decision 5 + the framework-integration mapping table), `traffic_probe_running_notes.md`
(decision log + rename map), and the per-domain notes (`relojistas_notes.md`,
`wayfaringlondoner_notes.md`). The chassis guidelines 001/002/003/004 govern.

## Where things stand (shared with P4)
- Engine (`site-engine`, stdlib Go) live on a dedicated EU box for
  relojistas.com (CPX22, 167.233.33.159): HTTPS, capturing, daily-JSONL store,
  `/events` export endpoint, retention prune timer. `intent-probe` component is
  in the live `content_components` library. Deploy is "commit is deploy" via
  two GitHub Actions (vm-sites content, site-engine binary), self-hosted runner.
- relojistas verdict: traffic is ~all bots/crawlers (Claude-SearchBot,
  SemrushBot, Yandex, a Chrome-spoof clone) hitting dead vBulletin paths; human
  intent ≈ 0. BUT the domain still has value (below).
- wayfaringlondoner.com page built + grounded (travel blog), not yet deployed;
  goes on an EXISTING box (no new boxes for now).

## Thread A — relojistas.com static-build manifest (do first; concrete)
Decision: relojistas is worth a real static site, not abandonment. Assets to
exploit: (1) an RSS feed that real aggregators still pull — populate it with
OUR content; (2) heavy search-crawler presence already indexing the domain —
keep them, feed them; (3) the dead-forum 404 paths + (once P4 lands) the
referer/landing-query log — these reveal what inbound links and searches want.
Task: **package everything we know into a manifest and hand it to the framework
to build a multi-page static site** (deployed to the VM via the same vm-sites
Action, or to B2 — decide per the capability model; it no longer needs a
backend unless we keep a capture box on it).
Manifest should include, per the chassis brief/roadmap contract (001/021):
- provenance (Spanish watch forum; boards: general, marcas, ferias, ventas/
  outlet/classifieds; the saved Wayback snapshot in the project);
- intended language Spanish; vertical = watches (brands/repair/marketplace);
- an RSS/news section seeded from our content (news-feed pipeline, doc 006);
- the top inbound 404 paths + referer/landing-query clusters (from P4 data once
  it accrues — gate the build on a first data pull, or build now and enrich);
- pinned sections via roadmap `section_types` (the planner honours these);
- whether to retain `intent-probe` on it (capability=backend) or go pure-static.
Open: build now from heritage alone, or wait ~1–2 weeks for P4 intent data to
inform content? (Lean: scaffold now, enrich from data.)

## Thread B — bring the build engine INTO the framework (the scalable side)
This is the "permanent, scalable" thread the operator asked to spin up. Goal:
a probe/backend site becomes a NORMAL chassis build (planner → content →
design → assemble → deploy) that differs only by (a) deploying to a VM and
(b) optionally including backend-requiring components. Most of this is the
existing P3/P5 analysis — pull it forward:
- **Multi-page, real languages, richer pages** ("when do we stop the one/two-
  page pattern?"): the moment the build runs through the chassis. The hand-made
  pages were only to unblock go-live. Once a probe site is a `sites` row with
  `github_repo='vm-sites'` + `deploy_config.target=vm`, the planner builds it
  like any site; `intent-probe` is just one pinned section. Language follows
  the site spec, not a hand-written file. THANKS_PATH-per-box stops mattering
  because the assembled site carries its own thanks page.
- **Chassis patch (P3, still pending):** the `resolveGitRepoName` helper +
  `git_commit`/`deploy_image_asset` call-sites + `upsertSite`/`ensure_site_record`
  github_repo plumbing (full spec in running notes 2026-06-10/11). Land this so
  the pipeline can target the VM repo at all.
- **Planner gate (Decision 5):** `load_components` +=
  `AND NOT (COALESCE(semantic_tags,'[]'::jsonb) ? 'requires-backend')` so
  backend components are opt-in via roadmap only.
- **P5 provisioning adapter (`vmhost`):** the analyser-adapter README skeleton;
  reuses thunder/ssh; `service_instances` registry (thunder_instances minus the
  reaper); onboard-domain = extend nginx + cert (setup.sh is already idempotent
  and multi-vhost). This automates what the runbook §3 does by hand.
- **Proposed doc 024 "VM-Hosted Backend Sites (site-engine)"** (Infrastructure
  Reference): the genuinely-new material — persistent non-reaped internet-facing
  VM class, DNS/public-TLS as managed state outside k8s, off-cluster data return
  (that's P4), commit-is-deploy seam + credential placement, capability gate.
  Draft it in this thread once the shape is agreed.

## Thread C — more domains (ongoing)
- Use EXISTING boxes only (no new provisioning now). wayfaringlondoner.com →
  deploy to an existing box: add its domain to that box's setup.sh DOMAINS and
  re-run (idempotent), set the shared `THANKS_PATH=/thanks.html`, rsync/commit
  `wayfaringlondoner-site/`. Page is built and grounded.
- Grounding new domains: operator supplies a Wayback URL/snapshot, or Claude
  works from web search + the name (Claude CAN web_fetch archive.org pages only
  when a search surfaces the exact URL; it canNOT enumerate CDX on demand).
- Each domain gets its own `<domain>_notes.md` (provenance/decisions/choices),
  per the relojistas/wayfaringlondoner template.

## Thread D — global bot blocklist (operator idea, cross-cutting)
relojistas is a good harvesting ground for illegitimate-crawler IPs (the
Chrome-spoof clone et al.) to block GLOBALLY across all boxes/sites. Design
sketch for this thread: derive candidate IPs from the nginx access logs
(high-volume, spoofed-UA, 404-storming, ignoring robots.txt), maintain a shared
denylist, and apply it at the edge (nginx geo/map deny, or Cloudflare if a box
is proxied). Keep legitimate crawlers (Googlebot, Bing, real Claude-SearchBot)
allow-listed. This is separate from intent capture but shares the log source.

## Cross-thread coordination
- P4 (this chat) creates `intent_events` + the collector + (next) the
  `backend_unreachable` discovery check. Thread A consumes P4 data for the
  manifest; Thread B consumes nothing from P4 directly but shares the capability
  model. Keep the living docs (plan/running-notes/per-domain) as the single
  source of truth across both chats — append, don't fork.
