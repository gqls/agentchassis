# deploy/ — getting the idea.uk engine onto a box

These artefacts deploy the engine as a persistent service behind nginx + TLS,
hardened. They're written to serve **both** the manual path now and the chassis
path later — the same `setup.sh` a person runs by hand is what the future
`service-deployer` workflow will `ssh_exec`. Nothing here is throwaway.

> **Drop your previous files in.** The actual nginx conf / systemd unit /
> Terraform from the earlier "push nginx + security + logging to a box" work
> aren't in the repo snapshot I can see — only `007_adoption_pipeline_v4.md`
> documents the recipe. `setup.sh` is drafted from that recipe plus the engine's
> real shape. If you paste your previous working files, I'll align these to them
> (ports, hardening choices, monitoring, the OVH Terraform pattern) rather than
> keep my drafted versions.

## Files

- **`setup.sh`** — the one artefact. Idempotent, non-interactive, parameterised.
  Installs nginx (two-stage with certbot TLS), ufw, fail2ban,
  unattended-upgrades, a hardened systemd unit, and the binary. Re-running it is
  the **rebuild** path. `MODE=update` just swaps the binary and restarts.
- **`idea.env.example`** — copy to `/etc/idea/idea.env`, fill in secrets. Holds
  `ANTHROPIC_API_KEY`, the Stripe keys, `REPORT_PRICE_GBP=29`, `AUTO_DELIVER=false`.
  Deliberately NOT written by `setup.sh` (secrets stay out of the script).

## Manual path now (Path A)

On a fresh small Ubuntu 22.04/24.04 VM (1 vCPU, 512MB–1GB is plenty — the engine
is I/O-bound on LLM calls):

```bash
# 1. get the binary onto the box (build locally first: `go build -o idea .`)
scp idea root@<vm-ip>:/tmp/idea
scp deploy/setup.sh deploy/idea.env.example root@<vm-ip>:/root/

# 2. point idea.uk DNS A record at <vm-ip> and let it propagate (certbot needs it)

# 3. on the box: fill in the env, then run setup
ssh root@<vm-ip>
cp /root/idea.env.example /etc/idea/idea.env   # (setup.sh makes /etc/idea on full run;
                                                #  or mkdir -p /etc/idea first)
nano /etc/idea/idea.env                          # ANTHROPIC_API_KEY + Stripe test keys
DOMAIN=idea.uk LETSENCRYPT_EMAIL=you@example.com bash /root/setup.sh
```

Then:
```bash
systemctl status idea
journalctl -u idea -f
curl -sS https://idea.uk/health
```

Stripe: create a webhook endpoint in the dashboard → `https://idea.uk/stripe/webhook`
for `checkout.session.completed`, put the `whsec_…` in the env, restart
(`systemctl restart idea`). Test-mode keys first; walk one order; then live keys
and a real £0.x test refunded via `/refund`.

Redeploying a new engine build later:
```bash
scp idea root@<vm-ip>:/tmp/idea
ssh root@<vm-ip> 'MODE=update DOMAIN=idea.uk LETSENCRYPT_EMAIL=you@example.com bash /root/setup.sh'
```

**Capture nothing extra — `setup.sh` IS the capture.** Whatever you tweak on the
box, fold back into `setup.sh` so it stays the single source of truth. That's
what makes it Path B's input for free.

## Framework path later (Path B) — the goal

You want idea.uk deployed/maintained/controlled by the chassis, not by hand. The
target is a `service-deployer` orchestrator agent + workflow, a sibling of
`model-trainer`, that does what you did manually:

1. **provision** a VM via the adapter — in a **persistent mode** (no reaper
   enrolment, no 18h uptime cap; the training safeguards must NOT apply or they'll
   tear the service down).
2. **ship the binary** by PUT to a presigned B2 URL, then `setup.sh` fetches it
   via `IDEA_BINARY_URL` — the same `prepare_artefact_url` mechanism the adapter
   already uses for training artefacts.
3. **`ssh_exec bash setup.sh`** with `DOMAIN`/`LETSENCRYPT_EMAIL`/etc. passed as
   env — this exact script, unchanged.
4. **deliver the env file** (`/etc/idea/idea.env`) over SSH — the one genuinely
   new concern vs training, because a service VM must hold its own credentials
   (ANTHROPIC_API_KEY + Stripe), unlike the deliberately credential-free training
   VMs. Design this carefully (drop a root-owned env file, or fetch secrets at
   boot from a store).
5. **register** the service in a `service_instances` table (sibling to
   `thunder_instances`): domain, IP, health endpoint, status — so the reaper
   skips it and monitoring watches it.
6. **wire DNS** to the new IP and let `setup.sh`'s certbot step issue TLS.

What's reused (already deployed): provisioning, `ssh_exec`, presigned-URL file
transfer, decommission, the spawn→provision→launch orchestration pattern.
What's new (the modest gap): persistent-mode provisioning (reaper exemption),
the env/credential delivery, the `service_instances` table, DNS wiring as a step
— and this `setup.sh` as the parameterised setup payload.

See `PARALLEL_engine_deployment_and_layer5.md` for the full reasoning.

## "idea.uk in a normal workflow"

idea.uk becomes the **first consumer** of `service-deployer`. Later, the
site-build pipeline can invoke the same workflow for any site whose `site_plan`
includes a backend service (the lightweight, route-shaped cousin —
`site_api_routes` config, no code deploy — already exists in `007`). And
separately, the chassis-native engine (Phase D — the method as `execute_llm_prompt`
+ `web_search` + HITL actions) is what runs the method **internally as a
site-planning input**, which you've confirmed is part of the plan. The deployed
binary (this dir) and the chassis-native actions (Phase D) are complementary:
the binary is idea.uk-the-product; the actions are the internal planner.
