# RUNBOOK — leopardessconsulting.co.uk

Two kinds of entry: **H** = a task only the human owner can do. **O** = an operator
procedure (commands), safe for me to run and for you to repeat.

Site: `leopardessconsulting.co.uk` · `site_id = 4851f6fc-71cf-4160-a270-e03d6d3e0732`
Postgres shortcut used throughout:

```bash
kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -c "<SQL>"
```

---

## Human tasks

| # | Task | Why it needs you | Status |
|---|------|------------------|--------|
| **H1** | Confirm the **name** to publish on the About page, or confirm it stays unattributed. | A1 removed the invented "Peter Grenfell". | **still open** — worldsoccernews.com and the background may be published; no public name given yet. Draft unattributed ("Founder and engineer") unless a name is supplied. |
| **H2** | Confirm **A7**: Go emits static SVG + inline JS enhancement, instead of literal go-echarts. | Deviates from a decision confirmed 2026-07-08. | **resolved 2026-07-10** — confirmed, data layer first (simple, extractable, reusable), display/colour/infographic treatment as a separate, later consumer. |
| **H3** | Confirm the **contact details** are current. | Every page carries them. | **resolved** — confirmed real. |
| **H4** | Approve the generated **logo** before it is committed as the permanent mark. | A logo is for the life of the site. | **resolved 2026-07-10** — owner chose c2. The exact approved image was background-knocked-out, gold-normalised to #C8A951, and deployed **live** to `/assets/images/logo.png` (+ favicon.ico, apple-touch-icon, og-card), all verified 200 and byte-identical. `sites.logo_url` set. Masters in `brand/`. |
| **H5** | Name worldsoccernews.com / Bumble publicly? | Real credibility signals, owner's to disclose. | **resolved 2026-07-10** — worldsoccernews.com yes, with the fuller, stronger detail (see H7). Bumble dropped from the bio entirely. |
| **H6** | Engagement shape (day rate, project, retainer). | Needed a commercial frame; owner asked for a full positioning pass first. | **resolved 2026-07-10** — pilot-first ladder (bounded fixed-price pilot → licence/day-rate/retainer decided by what the pilot reveals). Not yet drafted into the `/engagement-model.html` page content — L5 work. |
| **H7** | Source for "third busiest sports site"? | No independent evidence found (Wayback/press/rankings checked 2026-07-10). | **resolved 2026-07-10** — owner has no citable source and wants the claim stated anyway, explicitly flagged as unproven recollection, with real boundary context (bigger than the BBC's sports coverage then, smaller than ESPN's Soccernet). Combined with two owner-asserted facts I have not independently verified: ~12 million unique users/month at peak, and coverage in a media trade magazine plus Microsoft's own advertising material. Drafted into `specs/identity.json` team bio with the hedge intact — this is the one place on the site where an unprovable claim is allowed to stand, because it is labelled as one. |
| **H8** | Approve the data-sovereignty positioning (A8). | Pitched as a capability built *with* a client, never a standing guarantee. | **resolved 2026-07-10** — owner confirmed the framing ("I like your angle with that"). Drafted into `specs/portfolio.json` as a use_case. |
| **H9** | Approve the startup/founder "faster start" angle as a client category. | Owner raised this new idea 2026-07-10 — a startup building its own agent product could start from this platform's already-solved operational layer. | **resolved 2026-07-10** — drafted into `specs/portfolio.json` as a use_case, honestly labelled not-yet-done-for-a-client. Owner explicitly deferred wording judgement to me. |

| **H10** | *(withdrawn 2026-07-10)* — there is no platform-wide asset-deploy emergency. See AUDIT D8 correction. The only real, low-priority nit is that `assets.url` is cosmetically stale (presigned) for 83 rows; render already ignores it. Optional cleanup: generalise the idea.uk `w9_04` url-flip backfill. | **not urgent** |

---

## Operator procedures

> **O7 — Commit a specific (pre-approved) brand asset via the git-adapter.**
> `scripts/commit_brand_assets.sh <domain> <brand_dir>` sends optimised files
> straight to the git-adapter (the same commit `deploy_image_asset` sends after
> optimising). Use this when injecting a **specific approved image** (e.g. an
> owner-chosen logo) rather than one the pipeline generates — there's no generation
> step to spawn. Assets must already be the right size; the script does not optimise.
> Verify by artifact: `curl` each path for 200 and diff the bytes. This put the
> leopardess logo live. **Not** a workaround for a broken pipeline — the normal
> generate→spawn asset-deployer→commit flow works (robot-hands, idea.uk serve fine).

### O1 — Inspect state
```bash
S=4851f6fc-71cf-4160-a270-e03d6d3e0732
# Pages and their sections
kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -c "
SELECT p.url, p.build_status, count(pc.id) FILTER (WHERE pc.build_status='deployed') AS deployed
FROM pages p LEFT JOIN page_components pc ON pc.page_id=p.id
WHERE p.site_id='$S' GROUP BY p.id ORDER BY p.nav_order;"

# Work items
kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -c "
SELECT item_type, status, handler_agent, priority, LEFT(error,80)
FROM site_work_items WHERE site_id='$S' ORDER BY status, priority;"

# Assets — url MUST be a git path, storage_path MUST be non-empty
kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -c "
SELECT asset_key, purpose, status, url, storage_path, origin_model FROM assets WHERE site_id='$S';"
```

### O2 — Trigger an agent
Copy the `kcat` block from `./082_submit_domain_unified.sh`. Only the body changes:
```json
{"action":"orchestrate","config":{"agent_type":"<AGENT>"},"input_data":{...}}
```
- Re-render + commit every deployed page:
  `agent_type=rerender-pages`, `input_data={"site_id":"…","domain":"leopardessconsulting.co.uk"}`
- Dispatch this site's queued work items:
  `agent_type=build-dispatch-loop`, `input_data={"site_id":"…","domain":"leopardessconsulting.co.uk"}`

### O3 — Rebuild pages *(the big gotcha)*
**`pages.build_status='needs_rebuild'` does nothing on its own.** The dispatch loop reads
`site_work_items` and never scans `pages`. Only `write_build_items` converts
`needs_rebuild` into work items, and it lives inside `site-work-orchestrator` /
`build-site-planner` — not inside the loop. This is why six pages have sat at
`needs_rebuild` doing nothing.

Insert the items explicitly:
```sql
INSERT INTO site_work_items (site_id, source, pipeline, item_type, severity, summary,
   spec, page_id, priority, handler_agent, status, created_by, item_key)
SELECT p.site_id,'manual','build','needs_content_page','high','Rebuild page: '||p.name,
   jsonb_build_object('id',p.id,'name',p.name,'url',p.url,'title',p.title,
                      'page_type',COALESCE(p.page_type,'content'),'sections',p.sections),
   p.id, 12, 'page-build-handler','triaged','manual','needs_page:'||p.name
FROM pages p
WHERE p.site_id='4851f6fc-71cf-4160-a270-e03d6d3e0732'
  AND p.build_status='needs_rebuild' AND p.status='active';
```
Then run O2's `build-dispatch-loop`, or wait ≤60s for the scheduled trigger.

### O4 — Unstick work items
```sql
UPDATE site_work_items SET status='triaged', claimed_by=NULL, claimed_at=NULL,
       attempt_count=0, error=NULL
WHERE site_id='4851f6fc-71cf-4160-a270-e03d6d3e0732' AND status IN ('claimed','failed');
```
If items still will not dispatch, check two things: the `handler_agent` exists in
`agent_definitions` (`claim_work_item` *blocks* the item otherwise), and the handler's
AI endpoint is healthy in `ai_endpoint_health` (it *releases* the item otherwise).

### O5 — Commission imagery
1. Fix `design_intent.avoid` first (it bans leopard imagery; the prompt prepender will
   otherwise fight the prompt).
2. Insert a `site_plan_imagery` row against the site's **current** plan:
```sql
INSERT INTO site_plan_imagery (plan_id, scope, scope_ref, key, kind, prompt, style_hints, constraints, source)
SELECT sp.id,'site',NULL,'logo','logo', '<prompt>',
       '{"aspect_ratio":"1:1"}'::jsonb, '{"no_text":true,"transparent_background":true}'::jsonb, 'manual'
FROM site_plans sp WHERE sp.site_id='4851f6fc-…' AND sp.is_current=true;
```
   Constraints: `kind ∈ {logo,hero,illustration,icon,infographic}` — enforced **twice**,
   in `chk_kind` *and* in `validImageryKinds` (`write_site_plan_action.go`). Changing the
   set needs a migration **and** the Go edit, together.
   `scope='site'` requires `scope_ref IS NULL`; `scope='section'` requires `<page>:<ordering>`.
3. Either wait for the `improvement-loop` sweep to emit `needs_imagery`, or insert the
   work item directly (`handler_agent='image-build-handler'`).
4. **Verify by artifact.** `assets.url` must be `/assets/images/…`, not an
   `s3.…backblazeb2.com/…?X-Amz-…` presigned URL. If it is presigned, `deploy_image_asset`
   was not passed `asset_id` and did not rewrite the row.

### O6 — Verify a deploy (never trust "complete")
```bash
for u in / /about.html /services.html /assets/images/logo.png /favicon.ico; do
  printf "%-32s" "$u"; curl -s -o /dev/null -w "%{http_code}\n" "https://leopardessconsulting.co.uk$u"
done
curl -s https://leopardessconsulting.co.uk/assets/css/styles.css \
  | grep -E -- '--color-(card-bg|header-bg|footer-bg|cta-bg)'
```
Expect: all 200; charcoal `#0D0D0D` chrome, gold `#C8A951` CTA, no `#0f172a`/`#1e40af`.

### O8 — Re-render pages (two modes — pick the right one)
`scripts/rerender_pages.sh <site> <domain> <page_name>…` — **regenerate section HTML from
content_data** (`spec.reason=section_data_resolved`, `spec.page_name`). Use after editing
a page's `content_data`.
`scripts/reassemble_pages.sh <site> <domain> <page_name>…` — **re-assemble only** (embed
current header/footer + deploy, no section regen). Use after a header/footer/palette change.
`page_name` = the page `name`, NOT the url.

### O9 — Make the whole site uniform after a header/footer change
`scripts/reconcile_headers.sh <site> <domain>` — idempotent: each round re-fires
page-rerender only for pages whose deployed HTML still shows the old header, waits, repeats.
**Throughput note:** the prod `agent-chassis` runs ONE replica and processes page-rerenders
serially (~45s each), so a full 30-page site takes ~15–20 min. Fire all, then let it churn;
re-run the reconcile to mop up stragglers. This is a platform throughput limit, not a bug.

### O10 — Change a site's palette (per-site, safe)
Fork palette→theme→collection (see `scripts/L3_fork_palette.sql`), repoint
`sites.style_collection_id`, then reconcile `design_intent`: **core** slots go in
`design_intent.color_scheme` AND `palette.reference_values` (the `analyze_design` LLM reads
the latter — give it prescriptive guidance or it invents a palette); **specialised** slots
live on the forked palette row. Re-render CSS via `webdesign-agent`. NEVER edit the shared
seed. Validate every text/bg pair with the WCAG formula in `color_util.go` first — the
platform does NOT gate specialised-slot contrast.

---

## Landmines (learned the hard way, do not re-learn)

1. **`needs_rebuild` is inert.** See O3.
2. **Never trust a `complete` work item.** This platform has shipped ten `complete_error`
   builds that reported success while building nothing. Verify the artifact.
3. **`apply_section_edit` sets `build_status='approved'`**, and every discovery check
   filters on `'deployed'` — so an edited section goes silently invisible to the whole
   audit surface. Fix by updating to `deployed`, then clearing
   `schema_mode`/`locked_at`/`locked_by` in a *second* statement (the first will trip
   `auto_lock_on_deploy`).
4. **Style collection `3196d966` is shared by four sites.** Fork; never edit in place.
5. **Reference images only work on Banana** (`kind=='icon'` today). SDXL silently
   ignores them. Brand consistency is impossible until the routing changes.
6. **`deploy_image_asset` only rewrites `assets.url` to the git path when passed
   `asset_id`.** Otherwise the row keeps a presigned URL that expires in 7 days — which
   is exactly how this site's logo died.
7. **`js_snippets` has no `updated_at` column.**
8. `site_work_items` has no `domain` column any more — filter on `pipeline`.
9. The partial unique index `idx_swi_dedup` will **silently suppress** a new work item if
   an open one shares `(site_id, item_key)`.
10. `sites.github_repo` being empty is normal. Every site deploys into the single
    `<GITHUB_ORG>/sites` repo, under a top-level `{domain}/` directory.
11. **No tenant isolation exists on this platform today** (no row-level security;
    shared Postgres, Kafka, and `ollama-adapter` pod across every site). Relevant if
    positioning copy or a client conversation implies otherwise — see AUDIT §4b, P5.
12. Only **two** text-generation providers work end-to-end: `anthropic` and `ollama`.
    `openai` is a stubbed error in `createAIClient`; nothing else is wired for text.
    "Mistral" is a model name run through Ollama, not a separate provider. Don't let
    site copy imply more provider choice than this.
13. **`page-rerender` needs `spec.reason='section_data_resolved'` + `spec.page_name`** to
    regenerate sections; without the reason it only re-assembles stored HTML. `page_name`
    is the page `name`, not the url. (`check_rerender_mode`.)
14. **Section resolvers override `content_data` every render.** Hero `background_image`
    auto-resolves to `/assets/images/hero.jpg` (change the FILE, not content_data).
    `source:"static"` schema fields (e.g. system-stats suffixes) re-apply their `fallback`
    from the schema and can't be overridden per-instance; the path is
    `{fields,<field>,fallback}`. A *section* component fork does NOT survive rerender
    (`save_page_sections` re-links to the canonical component by function); a *header/footer*
    fork DOES (wired via the style_collection).
15. **`rerender-site`'s page loop stalls** on a lost child response. Don't rely on it;
    drive pages with O8/O9.
16. **`analyze_design` invents a palette** unless `design_intent.palette.reference_values`
    is set with prescriptive guidance (it does NOT read `color_scheme`).
17. **Back up before ANY change, including component/theme forks.** Use `bak_*` tables.
18. `kubectl exec … <<HEREDOC` silently no-ops without `-i`. Prefer `kubectl cp` + `psql -f`.
19. **`dimensions` on `assets` is jsonb**, not text (`'{"width":1024,"height":1024}'`).

---

## Reference: threads deferred out of this workstream

- **UK-sovereign stack exploration** (compute+storage+models all UK/self-hosted) —
  explicitly deferred to a separate chat by the owner, 2026-07-10. Baseline facts
  captured in memory (`uk-sovereign-stack-exploration`) and `AUDIT_verified_facts.md`
  §4b P6, so that chat doesn't have to re-derive them. Do not start this unprompted.
- **llama3.3:70b → live inference** — real TODO, logged in its home document rather
  than duplicated here:
  `docs/agent_docs/docs024_key_docs_latest/009_model_infrastructure.md` "Future"
  checklist. Not this workstream's to execute, but the data-sovereignty positioning
  (H8) becomes stronger once it lands.
