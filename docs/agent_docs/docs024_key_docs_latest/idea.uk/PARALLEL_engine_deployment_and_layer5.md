# Parallel thread — deploying the engine, and where Layer 5 really stands (2026-06-04)

You pushed back on my earlier "Layer 5 is a far-future gap" framing, and you're
right. I went and read the repo. The honest reassessment: **the hard plumbing
for Layer 5 already exists and is deployed in production.** What's missing isn't
the difficult part (provisioning, SSH, file transfer) — it's a thinner wrapper
for a different *operational mode*. Below is what's actually there, the real
gap, and two concrete paths to get the engine onto idea.uk.

---

## What already exists (grounded in the repo, not assumed)

### 1. The Thunder adapter — provisioning + SSH, deployed and verified

`cmd/thunder-adapter` is a long-running cluster adapter (same pattern as
`ollama-adapter`, `image-generator-adapter`). Per `FOCUS_finetuning_flywheel_and_service_23_.md`:
**"PHASE 4 STATUS: COMPLETE & DEPLOYED (verified in production 2026-05-24)."**
Its actions, all exposed to workflows:

- **provision** a VM (Thunder Compute), capturing IP + SSH port into a
  `thunder_instances` table.
- **`ssh_exec`** — verified, exit 0, real stdout from a real A100.
- **`ssh_get_status`** — built on the same dial path.
- **`prepare_dataset_url` / `prepare_artefact_url`** — presigned B2 URLs for
  moving files on/off the box without putting credentials on it (verified GET of
  a real ~21.5MB dataset).
- **`decommission_instance`** — verified, including the idempotent
  already-deleted path.

So: *provision a box, run commands on it, move files to/from it, tear it down* —
all real, all deployed, all callable from a workflow. This is the part I wrongly
implied didn't exist.

### 2. The orchestration pattern — provision → run-over-SSH, as a chassis workflow

The `model-trainer` agent (in `bk_agent_definitions_backup.sql`) is a working
template for exactly the shape we'd need: a workflow that spawns a data-preparer,
**provisions compute** (`gpu-provisioner`), then **launches work over SSH**
(`training-launcher` via `ssh_exec`), tracking state in a table. The
`training-launcher` that actually drives a real workload over `ssh_exec` was the
in-flight milestone (Phase 5) as of those docs — so the *pattern* is proven and
the *first real consumer* was the piece being finished.

### 3. The box recipe — nginx + TLS + security + systemd

`007_adoption_pipeline_v4.md` documents the "former code that pushed nginx and
security and logging to a box" you remembered. Its "Minimum viable Layer 2":

```
OVH VM (extending existing Terraform pattern)
  ├── nginx (reverse proxy, SSL via certbot, rate limiting)
  ├── site-api-router (Go binary, systemd service)
  ├── Postgres (form submissions, route configs)
  ├── Prometheus + Grafana (monitoring)
  ├── fail2ban (security)
  └── cron jobs (data exports, cert renewal)
```

Plus a **config-driven `site_api_routes`** mechanism in `site_specs` — routes
defined in the DB, reusing the existing action library, no code deploy. Whether
this is fully built, partial, or just designed, you'll know better than the docs
show — but either way it's the *content of the setup script* a deployment needs
to run on the box.

### 4. Storage + credentials plumbing

`032_storage_architecture_and_credentials.md` + the B2 presigned-URL pattern the
Thunder adapter already uses give us a clean way to ship the built engine binary
to a box (PUT to a presigned URL, fetch on the box) without baking credentials
into an image.

---

## The real gap — it's the operational mode, not the plumbing

Here's the crucial bit, and it's spelled out in the Thunder adapter design doc
(`033_thunder_adapter_design_1_.md`). Thunder provisioning is **deliberately
built for ephemeral, per-run instances**: provision → train → decommission. The
design *explicitly retracted* the persistent-VM-with-a-service option (its
"Option B") because for training, ephemeral is cheaper and lets the VMs stay
credential-free. The safeguards reflect that:

- **18-hour hard uptime cap** (a reaper kills the VM).
- **Reaper every 15 minutes** reconciling Thunder ↔ DB, decommissioning.
- **Concurrency limit of 2**, **$100/day cost cap**.
- **Credential-free VMs** (presigned URLs only).

A persistent idea.uk service is the **exact opposite shape**:

| | Ephemeral training VM (exists) | Persistent service VM (needed) |
|---|---|---|
| Lifetime | minutes–hours, then killed | stays up indefinitely |
| Reaper | yes — decommissions it | **must be exempt** — reaper would kill the service |
| Uptime cap | 18h hard | none |
| Credentials | none on VM (presigned URLs) | **must hold its own** (ANTHROPIC_API_KEY, Stripe keys) |
| Networking | outbound only | **stable inbound DNS + TLS** for the page + Stripe webhook |
| Restart | n/a (one job) | **systemd keep-alive** |

So you can't just point the Thunder adapter at idea.uk — its reaper and uptime
cap are designed to tear instances down, which is correct for training and fatal
for a service. **The gap is a persistent-service operational wrapper, plus
credential delivery to the box — not the SSH/provisioning plumbing, which is
done.** That's a much smaller gap than "build Layer 5 from scratch."

---

## Two paths to deploy the engine to idea.uk

### Path A — manual VM now (this is B1; do it to get live)

Exactly what you proposed: start a remote VM by hand and reuse the 007 box
recipe manually. Concretely:

1. Spin up a small VM (Hetzner/OVH/DO — 1 vCPU, 512MB–1GB is plenty for the
   engine; it's I/O-bound on the LLM calls, not CPU-bound).
2. Build the binary (your tree has the Dockerfile; `go build` also fine) and get
   it onto the box.
3. Apply the 007 recipe by hand: nginx reverse-proxy → the binary on `:PORT`,
   TLS via certbot, fail2ban, a systemd unit so it restarts on crash/reboot.
4. Point `idea.uk` DNS at the box's IP. Certbot issues the cert.
5. Set the service env (full list below), including the real Stripe keys, and
   register the Stripe webhook at `https://idea.uk/stripe/webhook`.
6. `AUTO_DELIVER=false`. Walk one test order end-to-end.

This is the fastest path to live, and — the important part — **the exact manual
steps you run here become the script that Path B automates.** Capture them as
you go (even just a shell script + an nginx conf + a systemd unit file).

### Path B — a chassis "service-deployer" workflow (the Layer 5 generalisation)

Once Path A has taught us the exact steps, generalise. A new `service-deployer`
orchestrator agent + workflow, modelled directly on `model-trainer`, that:

1. **provisions** a VM — reuse the adapter's provisioning, but in a
   **persistent mode**: no reaper enrolment, no uptime cap. Cleanest is a small
   sibling adapter (or a `persistent: true` flag + a reaper-exclusion) so the
   training safeguards stay intact for training.
2. **ships the binary** — PUT to a presigned B2 URL, fetch on the box. This is
   `prepare_artefact_url` used for a service binary instead of a training
   artefact — same machinery.
3. **`ssh_exec`s the setup script** — the 007 recipe (nginx + certbot +
   fail2ban + systemd + the binary). This is the one genuinely new artefact: a
   parameterised setup script. It's the thing Path A produces by hand.
4. **delivers credentials** — the real new concern. A service VM must hold its
   own `ANTHROPIC_API_KEY` + Stripe keys (unlike credential-free training VMs).
   Either drop a root-owned env file over SSH, or have the service fetch secrets
   at boot from a store. This is a deliberate credential-boundary change from
   the training path, and worth designing carefully (it's the one place the
   "credential-free VM" property we have for training doesn't carry over).
5. **registers** the service in a `service_instances` table (sibling to
   `thunder_instances`): domain, IP, health endpoint, status — so the reaper
   *skips* it and monitoring *watches* it.
6. **health-checks** via the existing Prometheus/Grafana pattern from 007;
   systemd handles restart.

What's reused vs new, explicitly:

- **Reused (exists, deployed):** provisioning, `ssh_exec`, presigned-URL file
  transfer, decommission, the spawn→provision→launch orchestration pattern, the
  007 box recipe content, B2 storage/credentials plumbing.
- **New (the gap):** persistent-mode provisioning (reaper/uptime-cap exemption),
  credential delivery to the box, DNS+TLS wiring as a step, the
  `service_instances` table, the parameterised setup script.

That's the real size of Layer 5 for this use case: a wrapper and a table and a
script, on top of plumbing that already works — not a greenfield build.

---

## "Include idea.uk in a normal workflow" — what that means

Two readings, both landing on Path B:

1. **Deploy idea.uk *through* the chassis** rather than by hand — i.e. idea.uk
   becomes the first consumer of the `service-deployer` workflow. (model-trainer
   is the template; idea.uk's service-deployer is the sibling.)
2. **Let the site-build pipeline produce sites that have a backend**, not just
   static files. Today the planner can specify static content + client-side
   tool widgets, and the build deploys them to B2. Adding a "this site needs a
   backend service" path means the build workflow can invoke `service-deployer`
   for sites whose `site_plan` includes a backend tool. The 007 `site_api_routes`
   mechanism is the lightweight end of this (routes as config, no code deploy);
   a full service like the engine is the heavyweight end (a deployed binary).

One distinction to keep clear, because it's two different things you might mean
by "the engine as an action":

- **Deploying the engine binary to a VM** (Path A/B) — infrastructure. This is
  for idea.uk-the-product: external, standalone, paid.
- **Expressing the engine as chassis actions** (the Phase D / chassis-native
  version) — the method runs as a workflow *inside* the chassis, no separate
  binary, reusing `execute_llm_prompt` / `web_search` / HITL. This is for
  running the method *internally* on our own domains as a planning input (Layer
  4), and it needs the schema pass first.

These are complementary, not alternatives. idea.uk wants the deployed binary;
the internal site-planner wants the chassis-native actions.

---

## Concrete go-live checklist (Stripe + box)

The service picks its payment provider in `main.go`: if both `STRIPE_SECRET_KEY`
and `STRIPE_WEBHOOK_SECRET` are set it uses the real StripeProvider, otherwise a
FakeProvider (local testing, no money). Full env the service reads:

| Env var | Purpose |
|---|---|
| `ANTHROPIC_API_KEY` | the engine (required) |
| `OPENAI_API_KEY` | optional — cross-vendor cut step |
| `GEN_MODEL` / `VERIFY_MODEL` / `CRITIQUE_MODEL` / `SCORE_MODEL` | model overrides (defaults: Opus 4.8 / Opus 4.8 / Sonnet 4.6 / Sonnet 4.6) |
| `OPENAI_CRITIQUE_MODEL` | default `gpt-4o` |
| `WEB_SEARCH_MAX_USES` | verify search budget (default 12, was 6) |
| `STRIPE_SECRET_KEY` / `STRIPE_WEBHOOK_SECRET` | live billing; absent → FakeProvider |
| `PUBLIC_BASE_URL` | `https://idea.uk` (used in Stripe redirect + webhook URLs) |
| `INTERNAL_API_KEY` | gate for operator endpoints (`/confirm`, `/decline`, `/refund`) |
| `OPERATOR_EMAIL` | where new-order notifications go |
| `AUTO_DELIVER` | keep `false` until reports are trusted |
| `MAX_ACTIVE_ORDERS` | capacity cap (default 8) |
| `REPORT_PRICE_GBP` | the £29 price |
| `ALLOWED_ORIGINS` | CORS for the page's form posts (if cross-origin) |
| `SMTP_HOST/PORT/USER/PASS/FROM` | email delivery; absent → reports written to files |
| `PORT` / `IDEA_DB_PATH` | listen port / JSON store path |

Sequence:
1. **Stripe** (test mode first): create account, get `sk_test_…`, add a webhook
   endpoint → `https://idea.uk/stripe/webhook` for `checkout.session.completed`,
   copy the `whsec_…`. Test cards before live.
2. **Box** (Path A): provision, deploy binary, nginx+certbot+fail2ban+systemd,
   DNS → IP.
3. **Env**: set everything above; `AUTO_DELIVER=false`.
4. **Walk one order**: request → confirm → pay (test card) → engine runs →
   operator review → deliver. Then switch to live keys and do one real £0.x test
   with your own card, and refund it via `/refund`.

---

## Bottom line

- The provisioning/SSH/file-transfer/decommission plumbing for Layer 5 **exists
  and is deployed** (Thunder adapter). I was wrong to file it under "far future."
- The remaining gap is a **persistent-service wrapper** (reaper exemption,
  credential delivery, DNS/TLS, a `service_instances` table, a setup script) —
  modest, and largely *assembling* existing pieces.
- **Do Path A now** (manual VM, reuse the 007 recipe) to get idea.uk live and to
  learn the exact steps. **Those steps become Path B's script** — the
  `service-deployer` workflow, a sibling of `model-trainer`.
- "idea.uk in a normal workflow" = idea.uk as the first consumer of
  `service-deployer`, and later the site-build pipeline invoking it for any site
  whose plan includes a backend. The lightweight cousin (`site_api_routes`
  config) already exists in 007 for route-shaped backends.

Next concrete step on this thread: when you do Path A, capture the setup as a
script + nginx conf + systemd unit. That artefact is 80% of Path B.
