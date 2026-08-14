# Netlify delivery + client editor — critical review and framework-native plan

## Context

The £149 offer delivers a finished static site as a private preview plus a ZIP
the customer hosts themselves. The owner now wants: (1) publish each finished
site to a **free Netlify account** as well as the ugg2.com preview, and
(2) point an **editor** at that published copy — either his own Go editor
("one for each client/website") or a third party's — so the customer can take
the site further themselves. A reference architecture was supplied
(`~/Downloads/Automated_CMS_Architecture.pdf`) to be reviewed critically and
then adapted "our own, but through the framework". Two small copy items ride
along: buttons stay ink (no action), and the attested build time changes from
"three or four days" to "usually next day".

## Part 1 — critical review of the PDF (what survives, what changes)

**Sound, keep:**
- Server-hosted engine (no distributed executables → no code-signing tax).
- Netlify zip-deploy API as the hosting layer; free tier fits £149 economics;
  the deploy loop (edit → zip → POST /deploys) is simple and robust.
- Editing STRUCTURED data, never parsing HTML — right instinct.
- Ownership transfer to the client's free Netlify account — matches the ruled
  offer exactly ("the files are yours", no lock-in, nothing we run for ever).
- Ops table: PAT in env only, strict schema on user input, deploy rate limits,
  tenant→site_id mapping in a DB.

**Fights the estate, change:**

1. **The "central Go engine on a VM" duplicates the platform.** The PDF
   reinvents multi-tenant auth, a tenant store and an orchestration brain —
   all of which the cluster already is (clients→networks→sites, auth-service,
   core-manager admin APIs, billing). A second brain with its own tenant DB is
   the two-stores drift the council already flagged once on PAY-009.
   → The CLUSTER is the engine. Anything public-facing follows the proven
   webdesign.uk box pattern (small Go service, systemd, nginx, WireGuard to
   cluster, **no cluster credentials on the box**).

2. **content.json would re-architect our deliverable, and we already have
   better.** The PDF's HTML/JS-fetches-JSON template is worse for SEO/FOUC on
   brochure sites and would replace our server-rendered pages. The framework
   already holds the structured source: `page_components.content_data`, typed
   per component schema, with `rendered_html` derived from it.
   → The editor edits **content_data through the framework**, the framework
   re-renders, the adapter redeploys. The framework IS the CMS engine the PDF
   sketches — the PDF's engine is a subset of what we run.

3. **The PDF never confronts the two-writers problem; we must, because it is
   this estate's best-documented failure class.** If a client edits a site the
   framework can still rewrite, the next regeneration clobbers them (the CTA
   saga, `page_divergence_overwritten`, bugs 226/229/268). The offer's own
   terms supply the answer: "no changes included" means that at handover the
   framework **stops writing** — handover must be a recorded STATE that
   excludes the site from discovery/improvement/rebuild, not merely a copy.
   After handover, the editor is the sole writer.

4. **"One editor per client/website" as separate processes is an ops trap for
   a one-person operation** — N systemd services to patch, monitor and secure.
   → One multi-tenant editor SERVICE with hard per-site isolation (per-client
   auth scoped to their site, per-site URL), which is logically "one editor
   per site" while operationally a singleton. A truly per-site embedded
   editor (shipped inside the exported site) is kept as a later variant, not
   the MVP.

5. **Netlify DNS delegation of the master domain is optional, not step one.**
   Published sites get `<name>.netlify.app` free; ugg2.com stays fully ours
   for previews; the customer's own domain connects after transfer under
   their control (matching the ruled DNS stance). Delegating ugg2 to Netlify
   DNS can come later if pretty preview URLs matter.

6. **The transfer step is manual (Netlify has no transfer API) — acceptable**
   at queue capacity 3–4; it becomes a runbook step at handover.

7. **Third-party editors lose on our constraints** (verified research, in the
   register trail): CloudCannon ~$55/site kills £149 economics; Decap/Tina are
   git+SSG-shaped and break for non-technical clients holding plain HTML.
   → Own Go editor, through the framework.

**Security shape:** the Netlify master PAT lives ONLY in cluster secrets
(`personae-platform-secrets`, like the Stripe keys); the box-hosted editor
holds no Netlify credential and no cluster credential — it calls a cluster
API (the CHAT-010 facts-relay pattern) which performs renders and deploys.

## Part 2 — what the estate already has (explored, verified)

**Delivery:** commit-is-deploy via git-adapter → `gqls/sites` → B2
`portfolio-sites` → one Cloudflare Worker serving 38 zones **including a live,
proven wildcard `*.ugg2.com`** (publish a preview = write objects under the
slug; already works, zero new code). Per-site target = `sites.github_repo`
(single string, one consumer — a non-git target needs a new field). **Zero
Netlify code. Zero Cloudflare-API automation.** The only generic outbound HTTP
is the registered `http_request` action. `S3Client` binds one bucket; nothing
binds `portfolio-sites`; **no `archive/zip` anywhere** — the ZIP deliverable is
genuinely unbuilt.

**Editing:** the platform IS a CMS already — `PATCH …/components/:component_id`
(CTS-003: history snapshot → content_data write → **auto-lock** → re-render
item) and the richer `apply_section_edit` action (field-level merge + the
whole gate set; one known defect: leaves `build_status='approved'`, hiding the
section from discovery). The admin SPA has a working structured-form editor to
crib. Rendered sections carry contract-enforced `data-component` markers
(section-addressable from the DOM; field-level markers don't exist).

**Auth gap (the one big hole):** customers cannot log in to anything — ADM-011
is admin CRUD *about* them. The only end-user auth precedent is noted-engine's
own accounts+sessions. And there is **no public route into the cluster** (the
Stripe webhook shares this problem) — so the editor lives on the box
(webdesign-chat/noted-engine systemd pattern) and calls a token-gated cluster
endpoint over WireGuard (the CHAT-010 facts-relay precedent), holding no
cluster or provider credentials itself.

**Two writers, resolved by what the owner's answer implies:** since the
framework keeps write access, customer edits must flow THROUGH the framework —
content_data in the one DB, never a divergent copy on the host. Then the
estate's existing, this-week-proven lock machinery arbitrates: a customer edit
auto-locks the component (exactly as admin edits already do), framework
rewrites skip locked rows and record the skip, history archives everything.
Two writers, one store, existing referee.

## Part 2b — provider think-hard (owner asked; matrix for the design pass)

| | automated? | customer ownership | free-tier risk |
|---|---|---|---|
| Netlify master acct + zip API | fully | transfer is DASHBOARD-ONLY — fails the bar | ~100GB/mo bandwidth + site cap, shared |
| **Netlify via customer OAuth connect** | fully (customer clicks Connect once) | **real — site created in THEIR account** | per-customer tiers, no shared ceiling |
| Cloudflare Pages, our acct | fully (Direct Upload API) | none, ever (no transfer mechanism) | effectively none (no bandwidth cap) |
| In-estate (B2 + worker + ugg2) | already live | ZIP only | none, but custom domains need paid CF features |

Direction: a provider-agnostic **publish seam** with the first backend chosen
by the design pass; ugg2 stays the preview either way; master credentials (if
any) live in cluster secrets beside the Stripe keys, never on the box.

## Part 2c — the architecture (one paragraph)

A provider-agnostic **publish seam** (`platform/publish/`) mirrors the
already-proven built artefact tree in `b2://portfolio-sites/<domain>` to a
hosted copy, per-site opt-in via new `sites.publish_target` (NULL = OFF, the
2026-08-02 opt-in ruling; deliberately NOT overloading `sites.github_repo`).
A scheduler reconciler publishes on **hash drift** — never inline after
`git_commit` — which sidesteps the async GitHub-Actions→B2 race and gives
deploy batching free. **Handover** is one nullable column
(`sites.handed_over_at`) with exactly one reader: the editor auth gate — the
framework keeps writing, per the owner. The **editor** is one multi-tenant Go
box service (port 8083, systemd/nginx/cloudflared per the chat-service
pattern; holds no cluster or provider credentials) that exchanges an emailed
magic-link token at a token-gated cluster endpoint (`sitefacts.go` cloned
shape) and edits through a shared extraction of the existing CTS-003
component-edit transaction with `locked_by='customer'` → permanent lock. Two
writers, one DB, the proven lock machinery as referee.

## Part 2d — provider: Cloudflare Pages, and why the Netlify question dissolves

**Primary backend: Cloudflare Pages Direct Upload, our account.** The
deciding move: because the framework keeps write access (owner ruling), a
site *transferred away* is incompatible anyway — so Netlify's dashboard-only
transfer stops being a missing feature and ownership reverts to what the
£149 terms already say it is: **the ZIP**. That frees the choice to be about
automation and failure modes, where CF Pages wins: pure-API end to end,
unmetered free bandwidth (Netlify free = ~100GB/mo SHARED across all client
sites — a quota suspension takes down a paying customer because of someone
else's traffic), free API custom domains later, one master token in cluster
secrets beside Stripe's, and loud API errors rather than silent quota deaths.
**Netlify-via-customer-OAuth stays available behind the seam** for a future
"hosted ownership" paid extra; in-estate B2+`<slug>.ugg2.com` is the day-one
second backend (near-zero code) and remains the preview channel.

## Part 2e — phases

**Phase 1 — rider, no council: "usually next day".** New
`SQL_2026-08-1X_build_duration_next_day.sql` in the lane dir: re-attest
`build_duration` in webdesign.uk's `evidence_base` AND update `writer_block`
(the wire — per `bugs_open/271`, never steer via item spec prose); then
`content_rewrite` items for pages whose SERVED html states a duration.
Verify: served grep old-string N>0 → 0; href counts unchanged; expect and
accept lock-skip items on the locked CTA components.

**Phase 2 — publish seam + CF Pages backend.** New `platform/publish/`
(`publisher.go` interface, `cfpages.go` Direct Upload manifest dance,
`b2worker.go` slug copy), `publish_site_action.go` (no-op on NULL target) in
`registry.go`; migration `sites.publish_target/publish_project/
published_hash/published_at`; second `S3Client` bound to `portfolio-sites`
(one bucket per client — `platform/storage/s3.go:20`); `CF_PAGES_API_TOKEN`
into `personae-platform-secrets`; reconciler entry in the scheduler.
Council + register (publish seam; opt-in field). Verify on a **quiet
portfolio canary site, not webdesign.uk**: per-page sha256(B2) ==
sha256(`<project>.pages.dev`) — the SERVED hashes, never the API 200; second
sweep with no change performs zero deploys; every other site still NULL.
Risk: Direct Upload is multi-step — atomic on CF's side once the deployment
is created, but acceptance is served-hash equality only.

**Phase 3 — ZIP deliverable.** New `zip_deliverable_action.go`: ListObjects
prefix → stream through `archive/zip` (first use in repo) → Upload under
`deliverables/<domain>/` → presigned URL. Council + register. Verify:
`unzip -l` count == object count; extracted index.html sha == B2 object;
presigned 200 in-expiry / 403 after; nothing on any box. Risk: memory on
image-heavy sites — stream per object, alert (never truncate) past a size
threshold; a truncated ZIP is a silent contractual failure.

**Phase 4 — handover state.** Migration `sites.handed_over_at timestamptz
NULL`; `POST /api/v1/admin/sites/:site_id/handover` (sets stamp, files the
ZIP item, mints the Phase-5 token, sends the link); admin SPA button.
`platform/mailer` EXISTS and is the platform's one sanctioned mailer
(verified: `platform/mailer/mailer.go`, ported+tested from idea.uk) — use it;
confirm SMTP creds are configured at implementation time, and if not, v1
shows the magic link in the admin UI for the owner to send himself. Council +
register, the entry explicitly stating what handover does NOT gate (deploys,
rewrites, locks). Verify: exactly one Go reader of the column; a rewrite
against a handed-over canary is byte-identical to before.

**Phase 5 — customer auth: cluster-issued magic-link tokens.** NOT
auth-service (platform-user space), NOT box-local accounts (forks identity
off the `clients` chain). Migration `customer_access_tokens(client_id,
site_id, token_hash, expires_at, used_at, …)`; new
`internal/core-manager/handlers/editor_gateway.go` cloned from
`sitefacts.go` (static header token, constant-time compare, fail-closed,
ClusterIP+WireGuard only): `POST /api/v1/editor/session-exchange` validates
single-use + `handed_over_at IS NOT NULL`. Add `"customer"` to
`humanLockSources` in `lock_policy.go` (registration; permanent ⇔ human-set
invariant holds). Council + register. Verify: used token → 401; pre-handover
→ 401; empty env → all 401; `LockPolicyFor("customer")` → permanent;
`kubectl get ingress` still empty. Risk: the link is the credential — short
expiry, single-use, re-issue button; the box session carries ongoing access.

**Phase 6 — the editor.** Extract `HandleUpdateComponent`'s transaction into
shared `applyComponentEdit(tx, …, source, lockSource, fieldUpdates)` with
server-side field-merge (closes the full-replace clobber for admin SPA too);
editor gateway gains scoped reads + `POST /components/:id/edit`
(`source='customer-edit'`, site_id ALWAYS from the session, never the body);
new box service `site-editor` (port 8083, `edit.webdesign.uk` vhost,
env = gateway token + cookie secret only), UI = server-rendered structured
form ported from the admin SPA's auto-form. Write path reuses the PATCH
internals NOT `apply_section_edit` (its citation gates are agent-facing and
its known `build_status='approved'` defect would hide customer-edited
sections from discovery — that defect files separately). Small edit:
`emitLockBlockedChangeItem` names `locked_by='customer'` in the skip item.
Council + register. Verify on two canaries: history row `customer-edit`
with prior blob; content_data diff exactly the field; real `content_rewrite`
at the section → skip item naming the customer lock; reconciler changes only
that page's served hash; **cross-tenant probe: session A requesting site B's
component IDs → 404 every time**. Sharpest risk of the whole plan: tenant
scoping must be structural (session→site_id joined once, all queries keyed
on it), proven by the cross-tenant probe.

## Verification (end-to-end acceptance)

A full dress rehearsal on a canary site: build → publish to pages.dev
(hash-equal) → handover (ZIP arrives, link arrives) → customer session →
one field edited → lock recorded → framework rewrite skips it and says so →
reconciler republishes only the changed page → ZIP re-cut on request.
Item statuses are never the acceptance; served artefacts and DB rows are.

## Register/council obligations roll-up

One council run per phase (2–6); concept-register entries: publish seam,
publish_target field, ZIP deliverable, handover state (with its
does-NOT-gate list), customer session exchange, customer lock source,
site-editor service. All new authority ships opt-in default-OFF.

## Part 3 — immediate copy item riding along

- Re-attest `build_duration` in webdesign.uk's register: "usually next day"
  (owner, 2026-08-14) superseding the carried-over "three or four days";
  writer_block updated in the same edit (facts[] is bookkeeping, the block is
  the wire); one rewrite pass on the pages that state a duration; verify with
  the href/link gates as always. Buttons: stay ink — record only.

## Owner decisions taken (2026-08-14)

1. **Provider: OPEN — "completely automated" is the bar**; Netlify not
   presumed; Cloudflare or another route welcome. (Part 2b matrix; design pass
   recommends.)
2. **Editor: one multi-tenant Go service, per-site isolation.**
3. **The framework KEEPS write access to delivered sites** — so two-writer
   arbitration is designed in (locks as referee, Part 2), not avoided.
4. **Customers get the editor after handover only.** Preview stays read-only.
Also: buttons stay ink (record only); `build_duration` re-attests to
"usually next day".

## Verification

[TO FILL after design]
