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

## 2026-06-10 — P3: pre-flight clean, repo surface complete, class-vs-instance naming fixed
- **Pre-flight result (operator):** `github_repo` empty on all 8 sites → the
  fallback patch is safe to land; nothing switches until a site is designated.
- **Repo-resolution surface complete (grep + uploaded sources):** exactly three
  sites. `git_commit` (config → default "sites"); `deploy_image_asset` line 463
  **hardcodes** "sites" inside the adapter request — MUST get the same fallback
  or a probe site split-brains (pages → VM repo, logo/hero → sites);
  `vet_med_export` is a dedicated single-site pipeline with its own default +
  config override — leave it. Confirmed absences: `deploy_tool` and
  `save_page_sections` have no git path (sections persist to DB);
  `generic_actions.go`'s DeployToHostingAction / HTTPRequestAction are
  simulation stubs, not real deploy paths.
- **Patch spec revised:** one private helper in the actions package
  (`resolveGitRepoName(config, collected)`: config `repo_name` →
  `site_record.github_repo` via datahelpers → "sites") used by `git_commit` and
  `deploy_image_asset`. Plus the upsertSite RETURNING + ensure_site_record map
  additions from before.
- **Naming correction (operator: "intent-probe is the wrong label"):** the CLASS
  is "sites with a server-side backend"; the traffic probe is instance #1 (chat/
  board sections join the same class later). Fixed: component keeps its narrow
  accurate name `intent-probe`; the invented site type is GONE
  (`suitable_site_types: []`); the planner gate now keys on the class marker —
  `AND NOT (COALESCE(semantic_tags,'[]'::jsonb) ? 'requires-backend')`; the
  site-side designation is `deploy_config || {"target":"vm","capabilities":
  ["backend"]}` set at onboarding alongside `github_repo`. *Correctness note:*
  with the site type dropped, the OLD suitable_site_types-based gate would have
  leaked intent-probe into the generic list — the tag-based gate is the right
  one. SQL revised + re-validated (quotes, JSON, vars==schema).
- **Open naming decisions (blocking small things):** (1) the VM content repo
  name (suggest `vm-sites`) — needed before any `sites.github_repo` write;
  (2) the engine repo name (suggest `site-engine`); (3) whether to neutralise
  the "probe" defaults in the box artifacts (service/user/paths/hook in
  setup.sh + both workflow YAMLs) to class-level names, since the box will host
  the whole backend-site class, not just probes — cheap now, nothing deployed.

## 2026-06-11 — component live; class-level rename applied (operator-confirmed)
- **intent-probe INSERTED into the live library** (`INSERT 0 1`, verify row
  active); the second run's `INSERT 0 0` is the ON CONFLICT idempotency working.
- **Names confirmed:** content repo `vm-sites`, engine repo `site-engine`,
  neutralise "probe" box defaults. Applied + revalidated (engine: vet, amd64
  build, smoke test through the renamed env var; deploy: `bash -n`,
  nginx -t "test is successful" with renamed zone/webroot, both workflows'
  YAML + inline bash clean; post-rename grep over all artifacts: zero strays).
- **RENAME MAP (every changed name, per the standing rule):**
  | old | new |
  |---|---|
  | env var `PROBE_DB_PATH` | `ENGINE_DB_PATH` (Go + setup.sh + env example + runbook) |
  | store default `probe_events.json` | `intent_events.json` |
  | go.mod `module probe` | `module site-engine` |
  | binary/build output `probe` | `site-engine` |
  | service user default `probe` | `site-engine` |
  | unit `probe.service` | `site-engine.service` |
  | `/opt/probe` | `/opt/site-engine` |
  | `/var/lib/probe` | `/var/lib/site-engine` |
  | `/etc/probe/probe.env` (+ example) | `/etc/site-engine/site-engine.env` |
  | webroots `/var/www/probe/<d>` | `/var/www/vm-sites/<d>` |
  | nginx conf `probe.conf` | `vm-sites.conf` |
  | rate-limit zone `probe_rl` | `engine_rl` |
  | env `PROBE_BINARY_URL/PATH` (+ `/tmp/probe`) | `ENGINE_BINARY_URL/PATH` (`/tmp/site-engine`) |
  | hook `/usr/local/sbin/probe-deploy-engine` | `/usr/local/sbin/site-engine-deploy` |
  | sudoers `probe-deploy` | `site-engine-deploy` |
  | apt/sshd conf `51-probe-noreboot`, `99-probe-hardening.conf` | `51-site-engine-noreboot`, `99-site-engine-hardening.conf` |
  | outputs folders `probe-go/`, `probe-deploy/` | `site-engine/`, `vm-deploy/` |
  Unchanged on purpose: the `intent-probe` component (accurate instance name),
  the engine's `ProbeSearch/ProbeCategory/ProbeFreeText` kind constants (they
  name the intent-capture feature, not the class), the living docs' project
  name (traffic-probe IS this project), and `probe-pipeline/` outputs folder.

## 2026-06-11 — relojistas go-live bundle + framework-integration mapping
- **Relojistas grounded in the snapshot:** it was a Spanish watch FORUM
  (vBulletin boards: general, marcas, ferias, ventas/outlet/classifieds) — so
  the probe is Spanish, `kind=search`, thanks at `/gracias.html`, action label
  covers marca/modelo/reparación/compraventa. Delivered: `relojistas-site/`
  (index + gracias, hand-instance of exactly what intent-probe renders) and
  `relojistas_golive.md` (exact commands: repos/secrets → deploy user → build/
  scp → setup.sh → env (THANKS_PATH=/gracias.html, INTERNAL_API_KEY) → rsync
  content → end-to-end verify → Action-driven engine update test).
  *Self-correction:* removed the beacon from gracias.html — counting thanks
  views would inflate the visit denominator.
- **OPEN ITEM (blocks the content Action, not the manual path):** layout
  discrepancy — git-adapter's CommitToRepo prefixes `<domain>/` at repo ROOT,
  but deploy-to-b2.yml watches `sites/**` and syncs `sites/$site`. Both can't
  describe the same repo. Operator to `ls` the live sites repo root;
  deploy-to-vm.yml then matches whichever is real.
- **Integration mapping done (004 read this turn; analyser README = the adapter
  precedent).** What ISN'T new: adapter skeleton (README pattern), SSH
  (thunder/ssh + shared/ precedent), registry (thunder_instances →
  service_instances minus reaper), thin actions, deployer-family agent,
  scheduled collection (doc 010), discovery checks + 600s sweep →
  site_work_items (doc 004), in-cluster Actions runner, table+upsert.
  **Key simplification:** P4 collection needs NO adapter/SSH — engine already
  speaks key-gated HTTPS; add `GET /events?since=` to site-engine (gap: /stats
  is summary-only today), then one Go action pulls + upserts `intent_events`,
  scheduled. SSH/adapter is only for provisioning/onboarding (P5 `vmhost`
  adapter). **Genuinely new** → proposed doc **024 "VM-Hosted Backend Sites
  (site-engine)"**: persistent non-reaped internet-facing VM class + lifecycle;
  DNS/public-TLS outside k8s; off-cluster data-return path; the off-cluster
  commit-is-deploy seam + credential placement; capability-gate semantics.
- The discovery tie-in is concrete: a `backend_unreachable` check (engine
  /health via the public URL) + optionally `missing_beacon` (rendered index
  lacks /api/hit img) in discovery_checks/, findings → site_work_items —
  same table format as the existing checks in doc 004.

## 2026-06-11 — layout discrepancy RESOLVED; deploy-to-vm.yml rewritten
Operator's repo listing settles it: the live `gqls/sites` repo keeps **domain
folders at the repo ROOT** (no `sites/` dir), matching git-adapter's
`<domain>/` prefixing; the bucket mirrors it, and today's commit→sync timing
proves the deployed workflow works with root layout. The `sites/**` YAML I'd
been mirroring sits at `agentchassis/.git/workflows/deploy-to-b2.yml` — inside
the `.git` METADATA dir, which GitHub never reads and git can't track: a stale
accidental copy (suggested deleting it). The authoritative b2 workflow is the
one in `gqls/sites/.github/workflows/` (unseen; aligning further is
nice-to-have, but root-layout logic follows from the adapter regardless).
**deploy-to-vm.yml rewritten for root layout:** trigger = push to main with
`paths-ignore` (`.github/**`, README*, LICENSE); detection = first path
segment of the diff; guards skip dot-entries, file edits, and deleted folders
(deletion-propagation gap is shared with the b2 action — noted, not fixed).
Validated: YAML parse, inline `bash -n`, and a behavioural simulation of the
skip logic (.github/LICENSE/missing dir all skipped, real domain deploys).
Side effect: `vm-sites` repo uses the same root layout — golive checklist
updated; the content Action is no longer blocked.

## 2026-06-11 — live b2 action learned; siblings rewritten; store scaling fix; relojistas notes
- **The REAL sites-repo action (operator paste):** `runs-on: self-hosted`
  (the in-cluster runner), branch `master` (9-year-old repo), changed-domain
  detection via dotted-first-segment regex `^[^/]+\.[^/]+/` (structurally
  excludes `.github`, `.idea`, `LICENSE`, and `unknown-domain/` — explaining
  its absence from the bucket), a **full-sync fallback** when the diff is
  empty, secret-presence checks, `--skip-newer`, and CF purge with per-domain
  zone lookup. **Both VM workflows rewritten as faithful siblings:**
  self-hosted, same regex + fallback, secret checks; vm-sites stays on `main`
  (new repo default; noted in-file); no CF purge (lift from the live action if
  a domain goes proxied). Validated: YAML, inline `bash -n`, and the regex
  against the six edge-case paths.
- **Store scaling fix (structural, pre-launch):** the old store rewrote the
  ENTIRE ever-growing JSON file on EVERY beacon hit (MarshalIndent + rename
  under the global mutex) — a linear-growth write cliff. Now: `AddVisit` marks
  a dirty flag only; a background flusher persists at most every 5s;
  `AddEvent` still persists immediately (events are the product); SIGTERM/
  SIGINT flush on shutdown; compact `Marshal`. Burst-tested: 50 hits + 1 event
  → disk shows 50/1 at the event persist; 30 more hits + SIGTERM → final 80/1;
  file compact. **New names (per the rule):** `Store.dirty` field,
  `Store.Flush()`, `Store.flushLoop()`, `flushInterval` const; `main.go` now
  runs the server in a goroutine with signal handling. No existing names
  changed; `AddVisit` no longer persists per call (deliberate, documented).
  Crash window: ≤5s of visit COUNTS, never events.
- **relojistas_notes.md created** (per-domain provenance/decisions/choices
  file): forum provenance, dated decisions (no-results-page v1, apex-only,
  pre-launch store fix), open choices (CF proxied timing, categories variant,
  retention 90d proposal, graduation criteria proposal), the three operator
  questions answered with exact commands (stats curl, jq last-20, top-terms
  aggregation, nginx ground truth), and the traffic mitigation ladder
  (JSONL+rotation → proxied → own box) with thresholds.

## 2026-06-11 — store v2 (JSONL), dedicated-VM sizing, no collector VM
- **Store v2 (the operator's "assuming high traffic" question exposed the next
  cliff):** v1 still rewrote the whole file per EVENT and held all events in
  RAM. v2: events append to **daily JSONL** (`events-YYYYMMDD.jsonl`, one line
  per submission, O(1) at any volume, bounded RAM, rotation = the date,
  retention = delete old files, future collector tails lines); /stats counters
  live in a small `counters.json` flushed by the existing dirty/5s flusher;
  SIGTERM also fsyncs the events file. Burst-tested: 300 events + 100 visits →
  300 lines, counters 100/300, per-1k math right, first/last lines verified.
  **Names (per the rule):** REMOVED `Store.Events` (in-RAM event map) and
  `Store.Snapshot()` (uncalled); NEW `Store.EventCounts`, `dir`, `snapPath`,
  `eventsF`, `eventsDay`, `openEventsFileLocked()`; env var **`ENGINE_DB_PATH`
  → `ENGINE_DATA_DIR`** (propagated: main.go, setup.sh placeholder, env
  example, runbook, golive, notes jq commands — grep shows zero stale refs).
  On-disk format change is deliberate and pre-launch.
- **VM sizing (relojistas, its own box):** Hetzner CX22-class — x86, 2 vCPU,
  4 GB, 40 GB. Sized by disk/log headroom, not CPU: even at the claimed 1.2M
  visits/month (~0.5 req/s avg, tens/s peak), static nginx + O(1) appends are
  far inside a small box. **Caveat: x86 only** — the engine Action builds
  GOARCH=amd64; Arm (CAX) would need a build-matrix change. Small domains
  share the second multi-vhost box as planned.
- **No third "collector" VM:** the serving box buffers (JSONL); the CLUSTER
  pulls over key-gated HTTPS (P4 scheduled collector) into clients_db. Pull
  keeps every credential in the cluster — boxes never hold DB/cluster secrets;
  push or a middle VM inverts that and adds an attack surface + a hop for no
  gain. B2 remains optional cold backup.
- Plan doc gained a top-of-file **"How it all fits (plain English)"** section
  (operator asked for orientation).

## 2026-06-11 — retention timer, input sanitisation, bandwidth, imagery, box question
- **Retention (operator's logrotate question):** daily JSONL IS the rotation;
  logrotate on engine files would race the open handle. Added to setup.sh:
  `RETENTION_DAYS` param (default 90) + `site-engine-prune.service/.timer`
  (daily find-delete of old `events-*.jsonl`). nginx logs keep their existing
  size-based logrotate. **New names:** RETENTION_DAYS, site-engine-prune.*
- **Engine input hygiene:** new `sanitizeValue()` in service.go — strips
  control chars, caps by RUNES (multibyte-safe; `MaxValueLen` semantic changed
  bytes→runes, deliberate). Tested: cap=10 yields `caña españ` (ñ intact),
  control-only submission dropped. *Test-harness lesson:* first run "failed"
  because dash doesn't expand `$'…'` — the field name literally became
  `$value`; the engine was right, the harness wrong (don't-jump-to-conclusions
  applied; re-tested via urllib).
- **Hetzner bandwidth (searched, current):** EU cloud (CX class) includes
  20 TB/mo — entry CX23 = 2 vCPU/4 GB/40 GB ≈ €3.49/mo; overage €1/TB. US
  locations were cut (≈1 TB × vCPU/2, Dec 2024 change); Singapore tiny
  allowances with €7.40/TB overage → use EU. Relojistas at the claimed 1.2M
  visits ≈ 360 GB/mo (~2% of allowance) — bandwidth is a non-issue.
- **Imagery ("latest watches"):** no manufacturer/press photos (rights +
  shop-implication + search ANCHORING contaminates the signal). v1 text-only;
  v1.1 optional ONE brand-free generated hero via the image pipeline; v2 idea:
  "novedades" as CATEGORY BUTTONS so the latest-models display becomes
  measurement (`kind=categories`) — A/B against the plain box only.
- **"Could we use the existing box?"** Capacity: yes. Recommendation: no —
  setup.sh has box-takeover semantics (`ufw --force reset`, removes nginx
  default site) and coupling an unknown-traffic experiment to the live idea.uk
  product saves ~€3.49/mo. Dedicated CX23 for relojistas stands (their own
  earlier instinct); the other small domains share a separate multi-vhost box.

## 2026-06-12 — debug guide v2_46, sanitisation v2, runbook §3 command walkthrough
- **Debug guide updated (operator ask: catch similar misses earlier):**
  `016_debugging_guide_v2_46.md` = v2_45 + two §0 checklist entries in the
  house style (rule → dated instance → discipline). **#24:** a config/workflow
  file is only authoritative at its runtime read-path — the
  `agentchassis/.git/workflows/deploy-to-b2.yml` stale copy (GitHub never
  reads `.git/`) nearly produced a never-firing Action; catch by checking
  `.github/workflows/` on the default branch of the repo that RUNS it and by
  triangulating commit↔bucket timestamps. **#25:** when a test "fails", prove
  the harness delivered the intended input before debugging the system — the
  dash `$'…'` case (field literally named `$value`; the engine was right);
  catch by inspecting what the server received, knowing the bashisms list
  ($'…', brace expansion, source), and the background-PID/timeout pattern.
- **Sanitisation v2 (operator correctly flagged v1 insufficient for saved
  terms):** now strips Cc AND Cf (zero-widths, bidi overrides incl. U+202E,
  BOM, soft hyphen), collapses whitespace runs, caps by runes. **Real bug
  found by the new tests:** checking IsControl before IsSpace silently JOINED
  words — `\t` is both Cc and whitespace, so `gmt\t\tmaster` → `gmtmaster`
  (v1 had the same latent flaw); order is now IsSpace FIRST. NFD combining
  marks deliberately SURVIVE: NFC normalisation + lowercasing belong at the
  P4 collector (needs x/text; engine stays stdlib-only) — added to the plan's
  ingest contract. Tests green: ZWSP/bidi stripped, runs collapsed, NFD kept,
  junk-only dropped. `sanitizeValue` rewritten in place; no renames.
- **Runbook §3 replaced with the full copy-paste walkthrough** (operator ask):
  `export BOX/DOMAIN/EMAIL` header, then 3.1 repos+secrets (manual creation —
  adapter makes PUBLIC repos), 3.2 CX23 EU box, 3.3 DNS grey-cloud, 3.4 deploy
  user, 3.5 build+scp+setup.sh (installs prune timer), 3.6 env (KEY +
  THANKS_PATH=/gracias.html), 3.7 content rsync, 3.8 end-to-end verify (incl.
  `systemctl list-timers site-engine-prune.timer`), 3.9 both Action seams.
  `relojistas_golive.md` stays as the domain-specific instance of the same.

## 2026-06-12 — operator started 3.1; handover gaps found and fixed
Operator feedback from beginning execution (OWNER=gqls; outputs copied to
`~/projects/agentchassis/docs/.../traffic_probe/deploy_setup`; checkout at
`~/projects/site-engine`):
- **"No env.go / no go.mod":** both EXIST and are current in outputs — they
  were simply never re-presented after the early turns, so they never appeared
  in the operator's downloads. All five engine files re-copied (fresh
  timestamps), compile-verified as exactly that set, and re-presented. Lesson
  absorbed into the runbook: 3.1 now has an explicit FILE MANIFEST per repo
  plus a `go vet && go build` line that fails loudly if any file is missing.
- **"Outputs has no .github/workflows dir":** by design — the delivery channel
  rejects dot-directories, so workflows ship FLAT in `vm-deploy/`; creating
  `.github/workflows/` is part of step 3.1. Now stated explicitly in both the
  runbook and the golive checklist. (Also: the outputs `vm-deploy/` dir itself
  was throwing an I/O error on listing — recreated and repopulated.)
- **"Should the repo live inside agentchassis?" — NO.** `site-engine` and
  `vm-sites` are standalone repos with their own remotes; working checkouts
  are SIBLINGS of agentchassis (`~/projects/site-engine` is correct as-is).
  Never nest a git repo inside the chassis tree. The `$OUTPUTS` copy under the
  docs tree is a reference snapshot only (contextkit pattern), not the working
  repo. Answer written into runbook 3.1.
- Runbook §3 var block gains `OWNER=gqls` and `OUTPUTS=<their real path>`;
  3.1 rewritten with full commands (gh repo create, manifest copy, gh secret
  set for VM_USER/VM_SSH_KEY now + VM_HOST deferred to after 3.2). Golive
  checklist aligned to the same ($OUTPUTS, shell-proof `cp $OUTPUTS/site-engine/*`
  instead of brace expansion — checklist #25 discipline applied to our own docs).
- **"Did I miss a go build step for go.mod?" — no.** go.mod is authored
  (once, via `go mod init` or by copying), never generated by `go build`.
  The delivered 2-line file (`module site-engine` / `go 1.22`) is the source
  of truth; `go mod init site-engine` would be equivalent. No go.sum exists
  by design (stdlib-only, zero deps → no downloads at build). Clarifying
  parenthetical added to runbook 3.1.
Operator paused here; next session resumes at 3.1 with the corrected docs.

## 2026-06-12 — operator execution: branch trap, secrets UI, box live, setup.sh usability
Operator progressed 3.1→3.5; findings and fixes:
- **Branch trap (vm-sites blocked):** local git defaulted to `master`; push of
  `main` failed (`src refspec main does not match any`). Subtlety: the retry
  `git add && git commit && git push -u origin master` chain STOPPED at commit
  ("nothing to commit" → non-zero) so the push **never ran** — repo is
  committed locally, unpushed. Both workflows trigger on `main`, so the fix is
  rename-then-push (`git branch -M main && git push -u origin main`), in BOTH
  repos. Runbook 3.1 now does `git branch -M main` explicitly before every
  first push.
- **Secrets:** operator found the Settings page — all three values go in as
  **Repository secrets** (not Variables, not environment secrets) because the
  workflows read `${{ secrets.* }}`; VM_SSH_KEY = full private-key file incl.
  BEGIN/END lines; both repos. UI walk added to runbook 3.1 alongside the gh
  commands.
- **Box provisioned:** Hetzner **CPX22** #140056673, nbg1 (EU ✓, x86 ✓ for the
  amd64 build), 2 vCPU/4 GB/80 GB/20 TB, **IP 167.233.33.159**, €11.39/mo —
  cheapest available at order time (CX23 not offered). Recorded in
  relojistas_notes coordinates; IPv6 unused v1 (no AAAA).
- **setup.sh invocation:** operator ran it bare and then positionally; both
  refused (params are env vars). Fixed two ways: runbook 3.5 now states the
  env-prefix requirement with an on-box form, AND setup.sh accepts positional
  domains (`bash setup.sh relojistas.com`) when $DOMAINS is unset — guard
  message rewritten; both forms behaviour-tested (bare fails on DOMAINS,
  positional passes to the EMAIL guard, env form passes).
- **"Apex-only" explained** + glossed in runbook 3.3 (apex = bare domain, no
  www; www needs its own record + server_name/cert later) + DNS-before-3.5
  note (certbot http-01; setup.sh idempotent re-run if DNS lags).
- golive §2 gained the missing `ssh root@<IP>` / `exit` lines.

## 2026-06-12 — provisioning ran; guard failure decoded (2 causes); cert pending DNS
- **The "DOMAINS was set" failure had TWO independent causes, one symptom:**
  (1) the variables were set on separate prompt lines WITHOUT `export` —
  shell-local vars never reach a child `bash /tmp/setup.sh` (the prefix form
  works because prefix assignments are exported to that command); (2) the
  positional retry failed with the OLD guard wording ("space-separated") —
  proof the box still had the PRE-patch setup.sh, so positional support wasn't
  there to use. Both now in the runbook (3.5) and the debug guide as **#26**
  (v2_47: export/prefix + `env | grep` discipline; error-text-vs-source
  mismatch = stale deployed artifact, #24 applied to scp'd scripts).
- **Provisioning succeeded end-to-end** (operator log): engine ACTIVE, unit +
  deploy hook + prune timer installed, nginx OK, hardening applied. Box runs
  Ubuntu "resolute" (newer than docs' 24.04) — no issues; noted.
- **Certbot 403 root-caused from the log itself:** the validator fetched the
  challenge from **76.223.54.146** — the domain's current (parking) DNS
  target, NOT the box. Registrar A record not yet repointed; setup degraded to
  HTTP exactly as designed. Recovery = repoint DNS → `dig +short` shows
  167.233.33.159 → re-run setup.sh.
- **Idempotency gap fixed in install_binary:** a full re-run with no
  URL/PATH (e.g. after a reboot clears /tmp) now KEEPS the installed engine
  instead of erroring — re-runs need no binary parameter (this is the normal
  fix-DNS-then-rerun path). bash -n validated; outputs refreshed.
- Box state recorded in relojistas_notes Log; env key (3.6) and content (3.7)
  can proceed while DNS propagates — neither depends on the cert.

- **"Where is relojistas-site?"** — same delivery-gap class as env.go/go.mod:
  the two page files were presented on an earlier turn only, so they never
  reached the operator's $OUTPUTS copy. Verified intact (index: form+beacon;
  gracias: neither, by design) and re-presented; runbook 3.7 now names them as
  deliverables to place in `$OUTPUTS/relojistas-site/`, and offers the
  commit-to-vm-sites alternative that doubles as the 3.9 content-seam test.

## 2026-06-12 — cert issued (recovery path worked); http2 deprecation fixed at the generator
- **Operator repointed DNS and re-ran setup.sh: cert ISSUED** (expires
  2026-09-10, auto-renew timer registered), stage-2 HTTPS conf live, engine
  active. The fix-DNS→idempotent-re-run path behaved exactly as designed; the
  keep-installed-binary branch wasn't needed (/tmp/site-engine still present).
- **Field finding:** nginx 1.28.3 warns on `listen ... http2` (deprecated
  since 1.25). Local container nginx is 1.24, where the replacement
  `http2 on;` doesn't exist — so the generator now emits **version-neutral**
  `listen 443 ssl;` (no http2; opt-in comment in the conf for backend sites on
  ≥1.25.1 where it matters). Renders + `nginx -t` "test is successful"
  locally; the box's next re-run validates it on 1.28 for real. Negligible
  loss for a one-form probe page.
- **rsync flags explained + glossed in runbook 3.7:** `-a` archive (faithful
  recursive copy), `-z` wire compression, `--delete` exact-mirror semantics —
  with the foot-guns noted (empty/wrong source wipes the destination; trailing
  slash = contents vs nested folder).
- Reminder logged in relojistas_notes: run 2 printed no env warning only
  because run 1 wrote a PLACEHOLDER env — INTERNAL_API_KEY likely still unset;
  3.6 before trusting /stats.
- Remaining to go-live complete: 3.6 env key, 3.7 content, 3.8 verify, 3.9
  Action seams.

## 2026-06-12 — FIRST LIVE CAPTURE; empty-$KEY decoded (session scope, not file state)
- **First production event in the store** (13:03:44 UTC, one minute after cert
  issuance): kind=search, "correa Omega Seamaster - aa", visits 2 / events 1.
  The full HTTPS path — page → form → /intent → sanitise → daily JSONL +
  counters → redirect — is proven live. ref_host/country empty as expected
  (direct visit; DNS-only, no CF geo header).
- **Empty `STATS KEY:` echo decoded:** the restart+echo ran in a NEW ssh
  session (login 13:34) where the earlier session's $KEY shell variable no
  longer exists. The echo proves nothing about the FILE — which may hold a
  real key (set earlier) or still the placeholder. Resolution command issued:
  `grep -E '^(INTERNAL_API_KEY|THANKS_PATH)='` on the env file; if placeholder,
  re-run 3.6 as ONE single-quoted ssh command (so $KEY expands remotely).
- **Docs hardened:** runbook 3.6 now ends with the grep (file is the truth;
  never `echo $KEY`) + the single-quoting note for remote one-liners; debug
  guide #26 gained the corollary (variables die with the SESSION; read state
  back from the artifact; ssh-quote so expansion happens remotely).
- Remaining: confirm/record the key → /stats check → 3.9 Action seams → done;
  then P4 `/events?since=` is the next build item.

## 2026-06-12 — /stats verified; traffic-claim assessment; WWW_ALIAS shipped
- **3.6 closed:** real key in the env file (recorded redacted in domain notes),
  THANKS_PATH set. **/stats over HTTPS works:** visits 4 / events 1 / 250 per
  1k — all operator self-tests; organic = 0 so far.
- **"Not 1.2M/month" — direction agreed, verdict withheld (don't-jump rule):**
  the claim predicts ~1,600 visits/hour; observed organic ≈ 0 in hour one.
  Three confounds keep it at "strong early signal": DNS propagation window
  (24–48h of old-TTL resolvers feeding the parking IP), beacon counts humans
  only (the claim likely counted bot-heavy requests → access.log is the
  ground-truth comparison, commands in domain notes), and the **www gap** —
  forum-era links likely target www., which has no record, so that traffic
  dies invisibly. Verdict criterion written into domain notes (48h +
  access-log UA split + www share).
- **WWW_ALIAS added to setup.sh** to close confound (3) structurally: opt-in
  param (default false — the v1 apex-only decision stands untouched); when
  true, every vhost's server_name gains www.<domain> and certbot requests the
  www SAN only if www DNS resolves (getent pre-flight; logged apex-only
  fallback otherwise; idempotent re-run upgrades). bash -n + rendered both
  modes + nginx -t "test is successful" ×2. New names: WWW_ALIAS, extra_san.
- Remaining on the checklist: 3.9 Action seams; then the box just accumulates
  evidence for 48h.

## 2026-06-12 — /events export endpoint built (P4 unblocked); www = CNAME; finish-up sequence
- **www DNS question:** yes — `www CNAME relojistas.com` is the right form
  (CNAMEs are fine on subdomains, only the apex can't be one) and beats a
  second A record: when the box relocates, one record updates and www follows.
  DNS alone isn't enough — the `WWW_ALIAS=true` re-run adds server_name + the
  cert SAN (pre-flighted on www resolving).
- **`GET /events` built + tested** (engine: store.StreamEvents + handler;
  setup.sh: nginx location beside /stats in the HTTPS block, `proxy_buffering
  off`, 120s read timeout). Design: key-gated NDJSON oldest-first, original
  line bytes preserved; params since (RFC3339, STRICTLY-after), host, limit
  (default 5000); final `_meta` line = {count, truncated, server_time} as the
  collector's checkpoint aid; checkpoint contract = max created_at received →
  duplicate-free pulls. **Lock-free by design** so a big export can never
  block live captures; the cost — a possibly torn mid-append tail line — is
  skipped and arrives next pull. Day-file skip uses the filename date vs
  since's UTC day. Tests green ×6: full export, since across day files, host
  filter, limit/truncated meta, 401 + 400, torn-tail resilience. Render +
  nginx -t clean; /events parity with /stats confirmed (HTTPS block; HTTP is
  301+ACME only). New names: Store.StreamEvents, App.events, strconv import.
- **Deploy convergence:** the 3.9 engine-seam test now SHIPS the endpoint
  (push updated store.go/service.go → Action builds & swaps), and the
  WWW_ALIAS re-run carries the nginx /events location — relojistas finish-up
  and the first P4 piece land in the same two operator actions.
- Remaining P4 (chassis side): `intent_events` DDL + collector action +
  scheduled task + `backend_unreachable` discovery check — next build items.

## 2026-06-13 — relojistas verdict (bots); passive signals already logged; wayfaringlondoner page
- **Relojistas VERDICT from the access log:** 14,961 reqs, 83% 404s on dead
  vBulletin paths, 2,242 redirects (the www/https 301s — NORMAL, redirect-loop
  fear cleared), 268×200, beacon 8× (all operator). UA = Chrome-spoof crawler +
  Claude-SearchBot + SemrushBot + YandexBot. Human intent ≈ 0. Clean probe
  result (domain not worth building), not a measurement failure. CF 1.13k and
  the 308/h spike were bot/crawler artifacts (spike = www-name discovery).
- **Operator chose passive referer+query capture — KEY FINDING: it's already
  captured.** setup.sh sets no custom log_format → nginx default `combined`
  already logs referer, UA, and the full request line (path+query). The engine
  CAN'T see the external referer on a static page load anyway (nginx serves the
  HTML; the beacon's referer is the page itself). So "capture passive signals"
  = a COLLECTION step (P4 collector parses the access log per domain), not an
  engine/page change, and not worth inverting the engine to serve HTML. Added
  to the plan's P4 ingest contract (referer + landing path + 404 paths + UA
  harvest). Avoided over-engineering.
- **Small legitimate engine enrichment shipped:** `landing_query` field on
  IntentEvent, populated from the submission's Referer query (the inbound
  ?q=/?utm=… that survives into the form page), so the structured /events
  export carries it without a log-join. Tested: present when the landing URL
  had a query, omitempty when not, external ref_host still recorded. New names:
  IntentEvent.LandingQuery, landingQuery() helper. Additive, no breaking change.
- **wayfaringlondoner.com page built** (`wayfaringlondoner-site/`), grounded in
  the project snapshot: a 2015–16 travel blog (Csilla; London + Bangkok/
  Transylvania/Jersey). BLOG framing, not marketplace — asks for a destination/
  London spot/story. Validated (form→/intent, beacon on index only).
- **Design point — THANKS_PATH is engine-wide:** one env var serves all vhosts
  on a box, so domains on the SHARED box must share a thanks filename →
  standard `THANKS_PATH=/thanks.html`, each domain ships its own `thanks.html`
  (wayfaringlondoner does). relojistas keeps `/gracias.html` on its own box.
  Recorded in wayfaringlondoner_notes.
- wayfaringlondoner targets the SHARED multi-vhost box (low-traffic blog);
  provision-now-or-batch is an open choice in its notes.

## 2026-06-13 (b) — scope split: P4 on this chat, permanent thread handed off; P4 started
- **Decisions:** (1) relojistas HAS value → static-build it (RSS we populate +
  crawler presence + 404/referer signal are assets); becomes a manifest→
  framework build (handoff Thread A). (2) gracias/multi-page/languages → the
  permanent thread: once the build engine is IN the framework, a backend site
  is a normal multi-page chassis build that deploys to the VM; hand-made pages
  were only to unblock go-live (handoff Thread B). (3) NO new boxes — next
  domains go on existing boxes. (4) archive.org: Claude CAN web_fetch archive
  pages but ONLY when a search surfaces the exact URL; canNOT enumerate CDX on
  demand and the sandbox can't reach archive.org directly — so grounding new
  domains = operator supplies Wayback URL/snapshot, or Claude uses web search +
  name. Existing two domains already have project snapshots (richer anyway).
- **wayfaringlondoner copy:** added "and under new ownership" to the tagline
  (operator request).
- **Scope:** P4 (collection) stays on THIS chat; everything else →
  `HANDOFF_vm_sites_permanent_thread.md` (Threads A manifest / B framework
  integration / C more domains / D global bot blocklist).
- **P4 STARTED:** `probe-pipeline/intent_events_migration.sql` — intent_events
  table (engine_event_id UNIQUE = structural idempotency; CHECK kind/value len;
  host→site_id resolve; checkpoint = max(event_created_at), no extra storage) +
  disabled `intent-collection` scheduled_tasks row (modeled on live
  improvement-sweep per-row dispatch + thunder-monitor insert-disabled
  convention) + the per-site deploy_config.engine.{base_url,stats_key}
  onboarding UPDATE. `probe-pipeline/intent_collector_actions.go` —
  CollectIntentEventsAction skeleton in the git_deployer_actions.go pattern
  (gofmt-clean; can't compile standalone — chassis pkgs absent — VERIFY markers
  V1–V4 at the DB/params/registry/egress integration points). Key storage
  decision: the box INTERNAL_API_KEY (read-only capture-export key) lives in
  deploy_config.engine.stats_key — low sensitivity, one accessor, movable to a
  secrets table later.
- **Remaining P4:** verify+register the action + intent-collector agent def;
  access-log harvest mode (referer/landing-path/404/UA — needs the box's
  access.log readable, likely a second engine endpoint or a log-tail action);
  `backend_unreachable` discovery check (needs live site_work_items DDL +
  discovery_checks pattern — request before writing); the ranking query.

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
