# PLAN — a framework-controlled VM estate (setup.sh → chassis)

**Started 2026-07-25**, on the owner's direction: *"we should aim to make the
setup.sh all controlled by the framework"* and *"be aware of the tools-api
project in the vonc.com and bastion threads — we'll want to try and merge the
projects rather than keep them separate."*

This workstream owns the machines the platform runs sites and tools on, and how
their configuration is produced. It spans three threads that each grew their own
box: **traffic_probe** (relojistas.com), **idea_uk_vm_site** (idea.uk), and
**gauntlet_dead_cta / bastion** (the tools-api island). Nothing here is built
yet — this is the design record and the walkthrough that produced it.

---

## Part 1 — Walkthrough: what `setup.sh` actually does today

`docs/…/traffic_probe/deploy_setup/vm-deploy/setup.sh`, 585 lines, run as root on
a fresh Ubuntu box. It is genuinely good work for what it is: non-interactive,
idempotent, parameterised, self-contained. Every claim below was read out of the
file, and the two defects were verified.

### 1.1 Contract and modes (lines 34–102)

`set -euo pipefail`. Every input is an env var with a default — no prompts — so
it can be driven by a machine. `DOMAINS` and `LETSENCRYPT_EMAIL` are required
(the latter is checked against `@example.com` placeholders, which is a nice
touch: a placeholder email silently poisons cert registration). Domains are
normalised: trimmed, lowercased, leading `www.` stripped.

Two modes: `MODE=full` provisions or **rebuilds** — re-running *is* the rebuild
path — and `MODE=update` only swaps the engine binary and restarts. Content is
explicitly *not* this script's job; it arrives by rsync from a GitHub Action.

### 1.2 The nginx generator (lines 104–308) — the interesting half

Config is assembled from small emitter functions, then written whole:

| function | emits |
|---|---|
| `rate_limit_preamble` | `limit_req_zone`, with a `geo`+`map` exemption when `WHITELIST_IPS` is set |
| `api_locations` | exact-match proxies to the engine: `/intent`, `/api/hit`, `/stats`, `/events`, `/health`, `/buscar` |
| `legacy_feed_locations` | the vBulletin `/external.php` variants → `/feed.xml` |
| `cloudflare_realip` | 22 published CF ranges + `real_ip_header CF-Connecting-IP` |
| `static_body` | `root`/`index`, the two blocks above, and the `try_files` catch-all |
| `server_block_http_only` / `server_block_https` | the per-domain vhosts |

`write_nginx_conf` loops the domains and picks the HTTPS block **only if
`/etc/letsencrypt/live/<domain>` exists**. That is why the script writes nginx
**twice** (stage 1 at line 486, stage 2 at line 513): HTTP first so certbot's
`.well-known` challenge can be served, then again once certs exist, upgrading
each domain that succeeded. A domain whose cert fails stays on HTTP and the run
continues — failure is per-domain, not fatal. This is the best-designed part of
the script.

Deliberate omission worth keeping: **no `http2`**, because `listen … http2` is
deprecated on nginx ≥1.25 while `http2 on;` doesn't exist before 1.25.1 — so the
generated conf is version-neutral across the estate.

### 1.3 The rest (lines 310–585)

Packages; a no-login system user; directories with explicit ownership;
`install_binary` (URL takes precedence over local path — the URL branch is the
chassis path, mirroring how the Thunder adapter already ships binaries) with an
atomic `mv` swap; a systemd unit with real hardening (`ProtectSystem=strict`,
`NoNewPrivileges`, `ReadWritePaths` scoped to the data dir); a **least-privilege
deploy hook** (`/usr/local/sbin/site-engine-deploy` + a single-line sudoers rule,
so a compromised deploy key can restart the engine and nothing else); a retention
timer that deletes `events-*.jsonl` older than `RETENTION_DAYS` (correctly *not*
logrotate, which would race the engine's open handle); ufw; size-based nginx
logrotate; fail2ban; unattended upgrades; and SSH hardening **guarded against
lockout** — it refuses to disable password auth unless it can find an
`authorized_keys` first. That guard is the kind of thing that only gets written
by someone who has been locked out once.

### 1.4 Defect A — `local` at top level [VERIFIED, FIXED 2026-07-25]

Line 496 (as was) declared `local extra_san=""` inside the top-level TLS loop.
`local` outside a function is a bash **error**, and under `set -e` it aborts the
run:

```
$ bash -c 'set -euo pipefail; for d in a; do if [[ ! -d /nope ]]; then local x=""; fi; done'
bash: line 1: local: can only be used in a function     # exit 1
```

It is dormant on a box whose domains already have certs (the branch is skipped —
so **the owner's pending relojistas re-run is safe**), and fatal on exactly the
path the file header advertises: *"Adding a domain = add it to DOMAINS and
re-run"*. It would die after writing stage-1 nginx and before stage 2, leaving a
new domain on HTTP and the engine unrestarted. Fixed in place with the
explanation attached.

**The idea.uk original does not have this bug** — it entered during the fork,
which is Part 2's argument in miniature.

### 1.5 Defect B — per-site policy emitted for every site [OPEN, latent]

`static_body` calls `legacy_feed_locations` and `api_locations` unconditionally,
so **every domain on the box** gets relojistas' vBulletin `/external.php` →
`/feed.xml` rewrites and a `/buscar` route. Harmless today (the box hosts only
relojistas.com — verified live), and wrong the moment a second site lands: a
watch-forum legacy URL scheme would be grafted onto an unrelated business site.

The generator has no concept of *whose* policy a location is. That is not a bug
to patch in bash; it is the thing Part 2 fixes structurally.

### 1.6 The friction that proves the point

Two things the owner must still do **by hand** on the box: append `WEBROOT_DIR`
and `RESULTS_PATH` to `/etc/site-engine/site-engine.env` (verified absent live
today), because the script deliberately never overwrites the env file — it may
hold `INTERNAL_API_KEY`. So the script owns *some* of the box's state, a human
owns the rest, and nothing owns the boundary.

---

## Part 2 — Why this should be framework-controlled

**The box's nginx conf is to the machine what `rendered_html` is to a page.**

That is not an analogy reached for; it is the same failure, already diagnosed in
this codebase. Contract doc `003` rejected HTML patching as an edit mechanism
because a patched artefact is silently destroyed by the next legitimate render.
On 2026-07-19 the box's nginx conf was **hand-edited surgically** to fix the
legacy feed; the generator did not know, so the next `setup.sh` run would have
deleted the fix. We reconciled it by hand on 07-24 — the same manual repair the
platform refuses to accept for pages.

The platform's own doctrine, applied to machines:

| pages | machines (today) | machines (target) |
|---|---|---|
| `content_data` is source of truth | env vars in a human's shell history | DB: site row + `deploy_config` |
| render → `rendered_html` | bash string-concatenation on the box | `render_vm_config` action in the chassis |
| deploy artefact, never edit it | hand-edit the live conf, reconcile later | apply rendered config; drift is detectable |
| per-site config drives per-site output | one global `static_body` for all | per-site profile decides its own locations |

Concretely, three properties we lack and would gain:

1. **The config becomes a rendered artefact of DB state.** Defect B dissolves:
   `legacy_feed_locations` is emitted because *relojistas' row declares a legacy
   feed*, not because it is in the shared function. Adding a site stops being an
   edit to a shared script.
2. **Drift becomes visible.** A rendered artefact can be diffed against the live
   box. Today the only way to know the box diverged is to notice.
3. **The owner stops being the transport.** "One box session" is currently a
   *dependency* in a workstream plan. It should be a dispatched job with a
   recorded outcome, like every other deploy.

### 2.1 Reuse, don't invent — the primitives already exist

CLAUDE.md's rule is to reuse existing machinery, and the machinery is already
here. The chassis provisions and drives remote machines **today** via the Thunder
path (registry.go):

- `dispatch_thunder_provision` → returns `instance_ip`, `ssh_user`,
  `ssh_key_secret_name`, `provisioning_id`
- `dispatch_thunder_ssh_exec` → **run a command on a provisioned instance**,
  returns `exit_code`/`stdout`/`stderr`
- `dispatch_thunder_ssh_status` → reachability + status command
- agents `gpu-provisioner`, `training-launcher`, `thunder-training-monitor`

That is provision → execute → monitor, already built, already in production for
GPU training. A VM estate needs the **same contract against a different
provider** (Hetzner, Mythic Beasts), not a new concept. The renderer half has a
precedent too: `render_rss_feed` reads DB rows and emits a file artefact for
commit — `render_vm_config` is the same shape with a different target.

Note the header of setup.sh already promises this: *"serves the manual path NOW
and the chassis service-deployer LATER"*. **There is no `service-deployer` in the
codebase** (verified: no match in `platform/`, `internal/`, `cmd/`). The promise
has been carried in a comment for months. This workstream is that promise.

---

## Part 3 — Merging with the tools-api island

Per the owner: merge, don't keep separate. The estate today is **three boxes
provisioned three ways**, with no shared source of truth:

| | relojistas box | idea.uk box | tools-api island |
|---|---|---|---|
| host | Hetzner CPX22 | (idea.uk VM) | Mythic Beasts VDS |
| provisioning | `setup.sh` (585 ln) | `setup.sh` (393 ln) | manual + scp'd compose/Caddyfile |
| web layer | nginx + certbot | nginx + certbot | Caddy in docker |
| ingress | public :80/:443 behind CF DNS | public :80/:443 | **cloudflared tunnel, no inbound** |
| app | site-engine binary + systemd | site-engine binary + systemd | containers (tools-api + Postgres) |
| backups | none (event prune only) | none | nightly `pg_dump` + off-box rsync |
| controlled by | a human running bash | a human running bash | a human running docker |

**The two `setup.sh` copies share 61 lines and differ on 614.** They are no
longer one script in two places; they are two scripts with a common ancestor and
divergent bug sets — Defect A exists in one fork only. A third divergence is
being born in the island's compose/Caddyfile. Left alone this becomes four.

### 3.1 What to merge — and the part that must NOT merge

Merge the **source of truth and the generator**. Do not automatically merge the
**control path**.

The island's whole security rationale is that *"the production cluster appears
NOWHERE in this path"* and *"NOTHING on this box holds any production
credential"* (RUNBOOK_island.md). Framework control in its naive form —
cluster holds SSH keys, pushes config to boxes — **inverts exactly that
isolation**, and would undo the reason Route B1 was chosen over the alternatives.

So the design splits cleanly:

- **Shared (merge these):** one profile schema describing a box; one renderer in
  the chassis producing every box's config; one repo location for the emitted
  artefacts; one drift check. This is where the 614-line divergence dies.
- **Per-box (keep separate):** how the rendered config *reaches* the box.
  - Public estate (relojistas, idea.uk): chassis pushes over SSH, reusing the
    Thunder `ssh_exec` contract via a provider adapter.
  - Island: the box **pulls** its rendered config — outbound-only, the same
    direction cloudflared already dials, so no inbound path and no cluster
    credential on the box is required. It stays isolated *and* stops being
    hand-maintained.

That distinction is the real design content of this plan: **merge the
generator, not the trust boundary.**

### 3.2 Profiles, not projects

One schema, two profiles, so the differences are declared rather than forked:

```
vm_profile:
  ingress:    public_tls | tunnel          # certbot+nginx  vs  cloudflared+Caddy
  runtime:    systemd_binary | compose      # site-engine    vs  containers
  backup:     none | pg_dump_offsite
  control:    push_ssh | pull_agent
  sites:      [ per-site policy from each site's deploy_config ]
```

`relojistas = {public_tls, systemd_binary, none, push_ssh}`;
`island = {tunnel, compose, pg_dump_offsite, pull_agent}`. Neither is a special
case; both are rows.

---

## Part 4 — Phasing (nothing built yet; each phase is independently useful)

- **P1 — Extract the truth.** Move what the two scripts hard-code (domains,
  ports, paths, per-site locations) into DB state: a `vm_hosts` table or a
  `deploy_config.vm` block, whichever survives a schema read. Deliverable: the
  existing boxes described as data, no behaviour change.
- **P2 — Render, then compare.** `render_vm_config` action emitting nginx conf +
  systemd unit + env skeleton from that data. Prove it by **rendering the
  relojistas box and diffing against the live conf** — a byte-level superset
  check, exactly the verification used when reconciling the generator on 07-24.
  Read-only; no box is touched.
- **P3 — Apply on the public estate.** Provider adapter implementing the Thunder
  `ssh_exec` envelope; apply the rendered config; verify against the running box
  (never against the intent — the CLAUDE.md pod-verify rule applies to boxes too).
- **P4 — Island by pull.** Small agent on the island fetching its rendered
  profile outbound; isolation preserved. Retires the scp-the-compose-file step.
- **P5 — Drift check.** Periodic render-and-diff across the estate; a difference
  is a work item, not a discovery someone happens to make.

Phase order is deliberate: **P2 pays for itself before anything is automated**,
because a renderer that can reproduce a live box is already the proof that the
DB description is complete.

---

## Part 5 — Open questions for the owner

1. **Island trust boundary** — ~~is pull-only acceptable as "merged"?~~
   **DECIDED (owner, 2026-07-25): pull-only.** *"I like that the island pulls its
   rendered profile outbound only."* The island joins the merged generator and
   drift check but fetches its own rendered profile outbound; the cluster never
   holds a credential to it. This is now a design constraint, not an option.
2. **Provider spread** — Hetzner *and* Mythic Beasts, or consolidate? Two
   providers means two adapters. There may be good reasons (UK sovereignty for
   the island; see the UK-sovereign-stack thread) — this plan does not assume.
3. **Scope of "the framework"** — should provisioning a *new* box (order the VM,
   DNS, first boot) be in scope, or only configuring boxes that exist? P1–P5
   above deliberately stop at "configure"; ordering hardware is a bigger ask.
4. **Where the emitted artefacts live** — `vm-sites` repo alongside content, or a
   separate infra repo? The repo-scoped-deploy-key argument in
   `REPORT_vm_sites_repo_architecture.md` probably applies here too.

## Related

- `traffic_probe/` — relojistas box, the pending owner convergence run, and the
  drift incident that motivates this.
- `gauntlet_dead_cta/infra/island/RUNBOOK_island.md` — the island as built.
- `idea_uk_vm_site/` — the older `setup.sh` fork.
- `REPORT_vm_sites_repo_architecture.md` — why vm-sites is its own repo.
