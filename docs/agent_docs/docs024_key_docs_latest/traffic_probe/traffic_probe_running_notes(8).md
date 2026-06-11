# Traffic-probe — running notes (in-chat reasoning, choices, caveats)

**Purpose:** capture the reasoning, suggestions, caveats and choice-points from
chat that aren't fully recorded in the runbook/plan. Living journal; appended
each session.

**Conventions:** chronological, one section per session; entries name the topic,
the reasoning, what got chosen (if anything), and where it landed. *Suggested but
not pursued* and *Caveats* are flagged. "Standing observations" and "Open
threads" sit at the end.

---

## 2026-06-10 — sessions so far (backfilled)

### Session 1 — first cut as a standalone service
Forked idea.uk's Go service into `probe-go`: a single multi-vhost binary picking
its page by Host header, file-based JSON store keyed by host, intent capture
(search / categories / free-text), no cookies/JS. Built and smoke-tested.
*Caveat raised next session:* this drifted into a separate project.

### Session 2 — reframing: not a separate project
Flagged that the standalone cut sat too far from the website-building chassis.
Read the chassis consolidation docs (`CONSOLIDATION_where_it_all_fits.md`,
`PARALLEL_engine_deployment_and_layer5.md`). Conclusion: the probe is **Layer 4
(build a targeted site for a domain) + a thin slice of Layer 5 (deploy a tiny
backend to a VM instead of static files to B2)** — one stack, not a side project.
Decided to keep the existing **git → self-hosted Actions** deploy seam and only
**swap the target** from B2 to the VM (the light path; the heavier chassis
`service-deployer` adapter is the eventual move, not now).

### Session 3 — schema grounding + engine trimmed to a backend
Read the real schema (`\d`): `sites` already has `github_repo`, `github_branch`,
`last_deployed_at`, and **`deploy_config jsonb`** (target switch with no DDL);
`site_work_items` is the build/maintenance pipeline (`item_type`/`handler_agent`/
`priority`/`depends_on`/`pipeline`); `maintenance_queue` FKs to `sites`;
`thunder_instances` is the precedent for a `service_instances` registry (minus
the reaper/uptime-cap).

Key consequence for the **maintenance/improvement-loop requirement**: if each
probe domain is a normal `sites` row with `site_specs` + work items, the existing
heartbeats and discovery agents (`build-pipeline-trigger`, `quality-discovery`,
`design-discovery`, `completeness-discovery`, `site-review-agent`) pick it up
automatically — they scan the live site over HTTP, so a VM-served site is covered
exactly like a B2 one. No separate registry needed for the loop.

Trimmed the engine to its correct shape: **API-only capture backend**. nginx
serves the chassis-built static files; the engine handles only `/intent`,
`/api/hit` (a no-JS/no-cookie 1×1 visit beacon for the events-per-1k denominator),
`/stats` (key-gated), `/health`. Removed `page.go` and `domains.json` — page
content + the per-domain "invited action" + the privacy line are now **chassis
build outputs**, not Go code. Builds and vets clean; smoke test passes (accepted
host stored, unaccepted host dropped, beacon counts visits, stats gated).

### Session 3 — open decisions raised (under discussion, not settled)
1. **Separate workflow** for probe-type sites vs reusing the existing build
   pipeline. Concern: current workflows are large/monolithic.
2. **Repo layout**: a separate (shared) repo for VM sites vs the existing `sites`
   repo arrangement; repo-per-domain judged likely overkill at 100s–1000s scale.
3. **Deploy mechanism**: per-site-repo Action ("commit is deploy", target swapped
   to VM) vs the heavier chassis-driven `service-deployer`.
4. **`needs_vm_deploy`** as a sibling terminal item to `needs_rerender` — or is
   the deploy difference not at the terminal build item at all (see plan).

---

## 2026-06-10 — decisions resolved
- **Same workflow:** reuse `build-dispatch-loop` (site work items); current build
  pipeline confirmed as the dispatch-loop, not `pageflow-builder` (latter may be
  deprecated separately). No separate probe build workflow.
- **Repo:** `git-adapter` already writes per-domain subpaths into a shared repo,
  and the B2 Action syncs a domain-named first-level path in one bucket
  (e.g. bucket `portfolio-sites` with `agritec.uk/`, `gamedesign.uk/`, …). VM
  sites get a **separate shared repo**, same layout, with its own VM-deploy
  Action. A site's `github_repo` value selects the target; the static repo and
  its B2 Action stay untouched.
- **Deploy:** light per-repo Action; the terminal build item stays
  target-agnostic (just assemble + commit to the site's repo).
- **D4 moot:** no `needs_vm_deploy` terminal item. *Caveat (don't lose):* the
  one-time per-domain VM setup still needs a home — Path A manual now,
  provisioning step/`service-deployer` later.
- *Deferred:* one VM repo → one Action → one box to start; routing a relocated
  domain to a second box (Action reads `deploy_config`/`service_instances`) only
  when traffic forces it.

---

## 2026-06-10 — box setup artifact (Path A)
Adapted idea.uk's authoritative `setup.sh` into a multi-vhost probe version: ONE
engine (single `probe.service`), per-domain nginx `server_name` blocks that
**serve the chassis-built static site from `/var/www/probe/<domain>` and proxy
only `/intent`, `/api/hit`, `/stats`, `/health`** to the engine. Per-domain
webroot `certbot` (graceful → stays HTTP, re-run upgrades to HTTPS); `MODE=full`
(idempotent; add a domain = extend `DOMAINS`, re-run) and `MODE=update` (binary
swap + restart). Kept ufw/fail2ban/logrotate/unattended-upgrades/ssh-hardening
guard, inline confs, and the presigned-URL binary path (chassis-ready).
Validated: `bash -n` clean; both branches rendered; `nginx -t` "test is
successful" (only the sandbox's missing IPv6 needed stripping — `listen [::]` is
correct on a real VM). Companion `probe.env.example` added. Per-domain CONTENT is
NOT in this script — it arrives via the deploy Action's rsync into the web roots.
*Caveat:* the deploy Action's SSH user needs write access to
`/var/www/probe/<domain>` — settle that ownership when writing the Action (P2).

---

## 2026-06-10 — VM-deploy Action (P2)
Saw the real `deploy-to-b2.yml` + the Cloudflare Worker. Static deploy is
serverless (B2 + Worker), hosted runner (`ubuntu-latest`), changed folders via
`git diff HEAD~1 HEAD -- sites/`, `b2 sync --delete` per changed `sites/<domain>/`,
CF cache purge per host. Wrote `deploy-to-vm.yml` as a near-mirror for the
VM-sites repo: same trigger/detection/runner, but `rsync -az --delete` over SSH
into `/var/www/probe/<domain>`; no CF purge (nginx serves direct). Secrets:
`VM_HOST`, `VM_USER`, `VM_SSH_KEY`. Resolved the earlier ownership caveat: added a
**`WEBROOT_OWNER`** param to `setup.sh` (default `www-data:www-data`; set
`deploy:www-data` so the deploy user can rsync). Validated: `setup.sh bash -n`,
owner-split, YAML parse, inline `bash -n` all clean.
*Scope boundary:* the Action deploys CONTENT for domains already provisioned
(in `DOMAINS`, with vhost+cert). A NEW domain still needs the one-time
provisioning re-run (extend `DOMAINS`, run `setup.sh`). The ENGINE binary deploys
via its own workflow in the probe-go repo (build amd64 → ship → `MODE=update`) —
not written yet.
*Note:* the outputs mount rejects dotfile dirs, so the workflow ships flat as
`deploy-to-vm.yml`; in the repo it belongs at `.github/workflows/deploy-to-vm.yml`.

## 2026-06-10 — engine-deploy workflow + Cloudflare clarification
- **probe-go clarified:** it's my working name for the engine's Go source (forked
  from idea.uk), NOT an existing repo. Proposed layout: engine source + `setup.sh`
  + the engine workflow in their own small repo; per-domain CONTENT in a separate
  VM-sites repo. Operator to confirm the real repo name.
- **Engine-deploy workflow written** (`deploy-engine-to-vm.yml`): build
  `linux/amd64` (static, stripped) → scp to box → run a narrow sudo hook. Verified
  the build command produces a static stripped ELF; YAML + inline `bash -n` clean.
- **Privilege model (low-risk):** no root key in CI. `setup.sh` gains
  `DEPLOY_USER`; when set it installs `/usr/local/sbin/probe-deploy-engine`
  (root-owned: atomic binary swap + restart) and a sudoers rule scoped to ONLY
  that script. Deploy user can swap the engine and nothing else; the binary runs
  as the unprivileged `probe` user. `setup.sh` does not create the user/key — the
  operator does, during onboarding. Validated `bash -n` + rendered hook.
- **Cloudflare-in-front answered:** keep DNS on Cloudflare via a PROXIED record
  (orange cloud) → VM, NOT a second Worker. No second copy of content ⇒ no sync
  problem; the VM is the single origin, Cloudflare just caches (purge on deploy).
  A Worker would only reintroduce sync if it served a copy — avoid. Adjustments:
  cache-bypass the API paths, set nginx real-IP-from-Cloudflare (so rate-limiting
  works), TLS Full(strict) keeping certbot or a CF Origin Cert. Bonus: CF-IPCountry
  (engine already reads it) + instant relocation. Added as runbook §8 (option,
  not the base path).
- **Runbook scope addition (requested):** explicit new-domain onboarding
  procedure (DNS → extend `DOMAINS` → re-run `setup.sh` for vhost+cert → deploy
  content → verify) in §4.

## 2026-06-10 — guidelines audit (001/002/003)
Read the dev guide, architecture, and contracts. **Existing code: no violations.**
- Engine is standalone `package main` — the action/agent/Kafka/adapter contracts
  don't bind it. Reuse, complexity-in-Go, no `logger.Debug` all honoured. The
  `kind` values are single lowercase words (convention allows). `setup.sh` +
  workflows are infra, outside the chassis contracts.
- D4 confirmed compliant: `needs_rerender` is the terminal item that assembles +
  git-commits + triggers the Action; `github_repo` selects the VM repo/target, so
  no new terminal item.
Forward constraints the guidelines impose (matter for P3+):
- **Designating a VM/probe site** = set `sites.github_repo` to the VM-sites repo
  (the target selector; no schema change). Per-site cadence lives in
  `sites.settings.maintenance_profile`. Data ownership: `content_components` is a
  manually-maintained library; cross-domain coordination is via `site_work_items`
  only (no agent-to-agent calls).
- **Capture component** must follow: kebab-case `function`; JS Content Separation
  (ours is a plain HTML `<form>` POST + beacon `<img>` → no JS, trivially
  compliant); Component Input Schema v2; the component-creator return contract.
  STEP ZERO first — check the existing library for a reusable form/search/
  lead-capture section before creating `intent-probe`.
- **Handler/dispatch rules** (if ever relevant): new item_types must be
  `pipeline='build'`; dispatch `input_mapping` spec.* fields need the `?` suffix;
  handlers read `input_data.spec.*`. We currently add NO new item_type (github_repo
  selects target), so this mostly doesn't apply.
- **Adapter Response Envelope Contract** (big trap, for P4 collection / P5
  service-deployer if done as chassis adapters): reply headers MUST be a typed
  struct with real bool `is_complete`/`is_error`, REUSE the incoming `request_id`
  (+ `in_response_to_request_id`), set `message_id`, send via
  `ProduceWithValidation`. Getting this wrong = silent drop until timeout
  (documented multi-day thunder-adapter fault). The engine's public HTTP is NOT
  this and is unaffected.

## 2026-06-10 — P3: repo selection traced, intent-probe drafted, capability gate proposed
**Repo selection (from git_deployer_actions.go / site_db_actions.go /
github_client.go):** `git_commit` reads `config["repo_name"]` → default
`"sites"`. `sites.github_repo` is DORMANT end-to-end: `upsertSite` doesn't
SELECT it, `ensure_site_record` doesn't return it, nothing reads it.
`CommitToRepo` prefixes every file with `domain/` for ANY repo (shared layout
confirmed); `createOrGetRepo` auto-creates missing repos as **public**
(`private:false`) — decide deliberately for the VM-sites repo.
**The patch (guide's "small patch" pattern, workflows untouched):**
1) `upsertSite` RETURNING += `COALESCE(github_repo,'')`, Scan += `&site.GithubRepo`;
2) `EnsureSiteRecordAction` return map += `"github_repo"`;
3) `git_commit` fallback chain: config `repo_name` →
`datahelpers.ExtractNestedFieldString(CollectedData, "site_record.github_repo")`
→ `"sites"`. *Pre-flight before landing:* `SELECT COALESCE(github_repo,''),
count(*) FROM sites GROUP BY 1;` (stale values would silently switch targets).
*Unverified:* whether other git-touching actions (deploy_image_asset?) resolve
repo_name separately — `grep -rn '"repo_name"' platform/` and apply the same
fallback wherever it's resolved.
**STEP ZERO (component library, 83 sections):** nothing captures anonymous
intent server-side. contact-form = PII collection (the opposite of the probe's
posture); tool-* = client-side JS; call-to-action = link-out. Verdict: NEW
`intent-probe` section. Drafted against the live `\d` (dollar-quoted, v2 schema,
no JS, CSS-var theming, conditionals only); validated: quotes balanced, JSON
parses, template vars == schema fields. **v1 limit (deliberate):** single
text-input action (search/freetext kinds); the {{range}}-based categories
variant is deferred until the renderer's array handling is verified.
**Capability link (operator's design point → open decision D5):** it's a
capability match, not a site-type match. Facts: planner's load_components loads
ALL active sections (no suitable_site_types filter → leak risk onto static
sites); roadmap section_types ARE enforced by prompt (positive direction
solved). Proposal, layered: (1) now, zero DDL — component carries
`suitable_site_types: ["intent-probe"]` + one-line planner gate
`AND suitable_site_types = '[]'::jsonb` making restricted components opt-in
(roadmap-only); (2) later — capability tags both sides
(`sites.deploy_config.capabilities` vs `semantic_tags: ["requires-backend"]`,
already on the component) + one audit check writing site_work_items findings.
*Unverified paths flagged in the SQL:* `config.intent_probe.*` resolution and
`site_specs.identity.contact_email` — check against a real site_specs row.

## Standing observations
- Static sites today are **serverless**: the B2 Action `b2 sync`s each changed
  `sites/<domain>/` to the `portfolio-sites` bucket; a **Cloudflare Worker**
  serves requests by mapping `hostname+path` → B2 object. No origin server. The
  VM path replaces both halves for probe domains (nginx serves + engine captures;
  DNS → box).
- Every agent is an orchestrator; keep workflows thin, push complexity into Go
  actions; spawn sub-agents rather than SQL sub-workflows (clean logs, separate
  responsibilities).
- Reuse/alter existing functions and agents before recreating similar structures.
- The probe must stay a first-class `sites` record so the maintenance/improvement
  loop covers it for free.
- Privacy posture (UK GDPR/PECR, low risk appetite): no cookies, no JS, no IP
  stored, referer reduced to host, country only from a coarse CDN header.
- "Commit is deploy" is the seam we are preserving; only the destination moves.

## Open threads

Resolved since opened:
- The four decisions (D1–D4) — see "Decisions — RESOLVED" in the plan.
- `git-adapter` repo/path logic: confirmed it already writes per-domain subpaths
  into a shared repo (so the VM path uses a separate shared repo, same layout).

Still open / next:
- **Engine-deploy workflow** — DONE (`deploy-engine-to-vm.yml` + `setup.sh`
  `DEPLOY_USER` hook). Build verified.
- **New-domain provisioning** — documented (runbook §4: extend `DOMAINS`, re-run
  `setup.sh`). Manual now; a provisioning step/`service-deployer` later.
- **P3 pipeline wiring (next focus)**: designate a site as VM/probe (its
  `github_repo` = the VM-sites repo selects the target) + the capture component
  the planner includes. Verify the dispatch loop's `input_mapping` and
  `handler_agent` resolution before writing anything.
- **Engine repo name**: confirm where the engine source lives (own repo vs folder)
  so the workflow `paths:`/location is right.
- **Off-box collection (P4)** + **registry/relocation (P5)** remain.

`site_work_items` schema facts (from the pasted `\d`, for P3):
- `item_type` is free `text` — NO CHECK/enum, so a new type needs no migration;
  `pipeline` defaults to `'build'`, `status` to `'detected'`.
- Dispatch picks up work via `idx_swi_handler (handler_agent, status)` and
  `idx_swi_site_pending (site_id, priority)` for status IN ('triaged','approved');
  so a new item must reach those statuses with a `handler_agent` set.
- `idx_swi_dedup UNIQUE (site_id, item_key)` (for non-terminal statuses) — set
  `item_key` to make new items idempotent.
- Still to verify before writing any item: the dispatch loop's `input_mapping`
  and how `handler_agent` is resolved to a running agent.
