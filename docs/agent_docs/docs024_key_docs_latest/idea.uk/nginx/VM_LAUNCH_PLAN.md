# VM launch plan (Path A) — hardened box, idea.uk as first tenant

This is the **infrastructure track**, deliberately a bit separate from the
idea.uk application work. The goal: a repeatable way to stand up a hardened
Ubuntu box (nginx + TLS + firewall + fail2ban + logging), run it by hand now
(Path A), and have the exact same artefacts become the payload for the chassis
`service-deployer` later (Path B). idea.uk is the first tenant; the box recipe
is general.

It builds directly on your year-old OVH reverse-proxy setup (the uploaded
Terraform + nginx + fail2ban + logrotate + prometheus + log-monitor files),
improved. A full assessment of those files — what's solid, what's broken — is in
§4, because "improve it all" means naming the specific problems.

---

## 1. Decision: dedicated box, not the existing reverse proxy

Your current OVH box (`51.89.148.216`) is a **shared multi-domain reverse
proxy** that terminates TLS and forwards everything to one Kubernetes NodePort
(`35.214.74.66:30080`). idea.uk is a different animal — the engine is a Go
service that runs *on a box*, not in your k8s cluster. So there are two options:

- **A dedicated VM for idea.uk** (recommended). Clean isolation; the engine runs
  on it; nginx on the same box reverse-proxies `127.0.0.1:8080`. This is what
  `deploy/setup.sh` does. idea.uk's DNS points at this new box.
- **Add idea.uk to the existing reverse proxy.** Possible, but the engine still
  has to run somewhere — either on the OVH box (mixing the proxy and the app, and
  the proxy currently only knows how to reach k8s) or on a separate box anyway
  (in which case you've built the dedicated box, so just do that).

**Recommendation: dedicated box.** It isolates idea.uk's blast radius, keeps the
engine's credentials off the shared proxy, and is the cleaner unit for the
chassis to provision and tear down later. Your reverse-proxy pattern is still
the right model for *content sites* fronting k8s — idea.uk is simply not one of
those.

---

## 2. What we reuse from your prior setup

Your files give us a proven baseline. Reused (improved):

- **The reverse-proxy shape** — `proxy_pass` + the standard `X-Forwarded-*`
  headers + a `limit_req` rate zone. Your `nginx_reverse_proxy_conf.tpl` is the
  template; `setup.sh` writes the idea.uk-specific equivalent.
- **fail2ban** — your `jail.local` (sshd + an nginx jail) is the basis. We fix
  the broken bits (§4).
- **logrotate** — your size-based nginx logrotate (`size 100M rotate 14`) is a
  genuine improvement over Ubuntu's default; folding it in.
- **Monitoring** — your `prometheus.yml` + node-exporter pattern; offered as an
  optional add-on (§5), off by default to keep the base lean.
- **certbot for TLS** — same as yours, but driven so re-runs are idempotent (§4).

---

## 3. Path A — the launch sequence

On a fresh small Ubuntu 22.04/24.04 VM (1 vCPU, 512MB–1GB; the engine is
I/O-bound on LLM calls). Full detail in `deploy/README.md`; the spine:

1. **Provision the VM.** By hand via the provider console/CLI, or with a small
   Terraform (your `main.tf` pattern — improved per §4 — provisions the OVH
   instance; do *not* use it to push per-file configs, see §4-D).
2. **Get the binary onto the box.** `go build -o idea .`, then `scp` it up (or,
   in the chassis path, a presigned B2 URL).
3. **Point idea.uk DNS** at the box IP; wait for propagation (certbot needs it).
4. **Fill the env file.** Copy `deploy/idea.env.example` → `/etc/idea/idea.env`,
   set `ANTHROPIC_API_KEY` + Stripe test keys. `AUTO_DELIVER=false`.
5. **Run `setup.sh`.** `DOMAIN=idea.uk LETSENCRYPT_EMAIL=you@… bash setup.sh`.
   Installs nginx + TLS + ufw + fail2ban + the hardened systemd unit, starts the
   service.
6. **Verify.** `systemctl status idea`, `journalctl -u idea -f`,
   `curl -sS https://idea.uk/health`.
7. **Stripe webhook** → `https://idea.uk/stripe/webhook`, put `whsec_…` in the
   env, restart. Walk one test order end-to-end.
8. **Fold any box tweaks back into `setup.sh`.** The script is the single source
   of truth and Path B's payload — don't let the box drift from it.

Re-running `setup.sh` is the **rebuild** path (idempotent). `MODE=update` swaps
just the binary for redeploys.

---

## 4. Assessment of your prior files (and the fixes)

Your files are a year old and from a much earlier model. The structure is sound;
there are several concrete bugs worth fixing before reusing them.

### A. The `if ($remote_addr = "…") { break; }` "whitelist" is a no-op
In `workdomain.conf` / `felines.conf`, every server block tries to whitelist
your IP with `if ($remote_addr = …) { break; }`. **This does nothing useful for
access control.** `break` only stops the current rewrite-module directives; it
does not bypass `limit_req`, basic auth, or anything else. (This is the classic
"if is evil" nginx trap.) Your IP was never actually exempt from rate limiting.

**Fix** (folded into `setup.sh`, optional via `WHITELIST_IPS`): use the
canonical `geo` + `map` pattern so the exemption actually works —

```nginx
geo $rate_limited { default 1; 176.25.120.48 0; 127.0.0.1 0; }
map $rate_limited $limit_key { 1 $binary_remote_addr; 0 ""; }
limit_req_zone $limit_key zone=idea_rl:10m rate=10r/s;
```

An empty `$limit_key` is not counted, so whitelisted IPs are genuinely exempt.
For idea.uk this is optional anyway — operator endpoints are gated by
`INTERNAL_API_KEY`, not IP, and `/audience-check` is rate-limited in the app.

### B. `felines.conf` serves the WRONG TLS certificate
Both `felines.co.uk` and `felines.uk` server blocks point
`ssl_certificate` at `/etc/letsencrypt/live/workdomain.co.uk/…` — the
*workdomain* cert. A browser hitting `felines.co.uk` would get a certificate for
`workdomain.co.uk` → **name-mismatch warning**. Either it's been silently
serving a bad cert, or those sites aren't actually live.

**Fix:** each domain gets its own cert (`certbot certonly -d felines.co.uk -d
felines.uk` → its own `live/` dir), or one SAN cert covering all names.
`setup.sh` issues a per-domain cert for whatever `DOMAIN` you pass.

### C. Terraform marks secrets as `sensitive = false`
In `variables.tf`, `password_for_ovh_ssh` and `htpasswd_password` are
`sensitive = false` — so they'd print in `terraform plan/apply` output and logs.
(`main.tf`'s `ssh_password` is correctly `sensitive = true`; the others aren't.)

**Fix:** `sensitive = true` on every secret var. Better: drop SSH **password**
auth entirely and use key-only (your `main.tf` supplies both a key *and* a
password — the password is an unnecessary attack surface). `setup.sh` disables
password auth (guarded against lockout).

Good news: your `terraform.tfvars` and `terraform.tfstate` **do not** contain
the password, htpasswd, or any private key — they're passed at apply-time. So
nothing leaked. (As general hygiene, tfstate can contain secrets for other
resource types; keep it out of any shared/committed location.)

### D. `main.tf` is only an SSH connection test
The uploaded `main.tf` has `required_providers {}` empty and a single
`null_resource` that SSHes in and runs `whoami`. The resources that actually
*push* your `.tpl` configs (the `templatefile()` + `remote-exec`/`file`
provisioners) aren't in the uploaded set — they live in a `.tf` you didn't
upload, or this is an early stub.

**Recommendation:** keep Terraform for **provisioning the VM only** (create the
instance, attach the SSH key, output the IP). Do **box configuration** with
`setup.sh`, not with per-file `templatefile()` resources. Reasons: (1) it's the
same artefact the chassis will run, (2) one idempotent script is easier to
reason about than a spray of file-push resources, (3) it sidesteps the
certbot-clobber problem in E.

### E. Terraform-pushed nginx templates fight with certbot
Your `nginx_reverse_proxy_conf.tpl` is HTTP-only (port 80); certbot then
*edits the live file in place* to add the 443 block (visible in the deployed
`workdomain.conf`). That means **re-applying Terraform overwrites the file and
wipes certbot's 443 edits** — every `terraform apply` would silently break TLS
until the next certbot run.

**Fix** (in `setup.sh`): own the full conf ourselves (write both the 80→443
redirect and the 443 TLS block) and use `certbot certonly` (which only issues
certs, never edits nginx). Re-runs are then idempotent and TLS never gets
clobbered. This is the single biggest robustness improvement over the old setup.

### F. fail2ban: a misconfigured jail and a missing filter
- `etc_fail2ban_jail_d_ssh_custom_usernames.conf` sets `filter = %(sshd_log)s`
  — that's a **log-path macro, not a filter name**, so the jail can't load its
  filter correctly.
- `jail.local`'s `nginx-rate-limit` and `ssh-custom-usernames` jails reference
  custom filters (`nginx-rate-limit`, `ssh-custom-usernames`) whose filter files
  weren't uploaded — so they may be silently failing to start.

**Fix:** for idea.uk, start with the stock `sshd` jail (bans failed SSH logins)
plus the standard `nginx-http-auth` / `nginx-limit-req` jails that ship with
fail2ban — no custom filters needed. `setup.sh` writes a correct minimal
`jail.local`. If you want the "invalid username" jail, it needs a real filter
file (the stock `sshd` filter in `aggressive` mode already catches most of it).

### G. The Python `nginx_log_monitor.py` — homegrown WAF, several issues
It tails the access log and writes `deny`/`return 403` rules. It overlaps with
fail2ban and has real problems:

- **`127.0.0.0/8` is in `CHINESE_IP_PREFIXES`** — localhost is in the
  block-on-match set (only saved because the whitelist is checked first; still a
  latent bug, and the /8 prefixes are extremely broad).
- **Over-aggressive UA blocking** — it blocks `curl`, `wget`, `python-requests`,
  `Go-http-client`, `Java/`, `HttpClient`. That would block legitimate
  server-side callers. (Stripe's webhook UA isn't matched, so webhooks survive —
  but it's luck, not design.)
- **Reload/clear logic is mis-nested** — nginx is only reloaded when a *UA* is
  blocked, not when only an IP is; the periodic `recent_*.clear()` is nested
  inside the UA branch.
- **Unbounded growth** — `append_to_file` re-reads the whole block file on every
  hit (O(n²)) and the block lists never shrink.
- **No systemd unit** — it's a foreground tail loop with nothing keeping it alive.

**Recommendation: don't run it on idea.uk initially.** fail2ban + nginx
`limit_req` + the in-app per-IP rate limit on `/audience-check` cover the real
threats without the risk of blocking legitimate traffic (including our own
server-side calls and Stripe retries). If you want it later, it needs: the /8
localhost bug removed, the UA list narrowed to genuine scanners
(`zgrab|masscan|sqlmap|nmap|nikto|…`), the reload/clear logic fixed, a bounded
block file, and a systemd unit. I can produce that corrected version when you
want it — flagged as a follow-up rather than done here, since the
recommendation is to launch without it.

### H. No security headers / HSTS
None of the deployed confs set HSTS or the standard hardening headers.

**Fix** (in `setup.sh`'s 443 block): `Strict-Transport-Security`,
`X-Content-Type-Options: nosniff`, `X-Frame-Options: SAMEORIGIN`,
`Referrer-Policy: strict-origin-when-cross-origin`.

---

## 5. Optional monitoring (your prometheus/node-exporter pattern)

Your `prometheus.yml` scrapes a node-exporter on `:9100`. For a single idea.uk
box this is nice-to-have, not required, so it's **off by default** in `setup.sh`.
To add it later: `apt install prometheus-node-exporter` bound to localhost, and
either run a local Prometheus or have your existing Prometheus scrape it over a
firewalled, authenticated channel. The engine could also expose its own
`/health` to a blackbox-exporter probe. Keep this for after the box is serving.

---

## 6. The bridge to chassis automation (Path B)

Everything above is built so the manual run becomes the automated one:

- `setup.sh` is what `service-deployer` will `ssh_exec` — unchanged.
- The binary arrives via presigned B2 URL (`IDEA_BINARY_URL`) — the mechanism
  the Thunder adapter already has (`prepare_artefact_url`).
- The VM provisioning (your Terraform / the provider) is what the adapter will
  do in **persistent mode** (no reaper, no uptime cap — the training safeguards
  must not apply).
- The one genuinely new piece is **credential delivery** — getting
  `/etc/idea/idea.env` onto the box safely (a service VM holds its own keys,
  unlike the credential-free training VMs).

See `PARALLEL_engine_deployment_and_layer5.md` for the full Path B design.

---

## 7. Immediate next actions

1. Provision one small VM (any provider; OVH if you want to reuse your Terraform).
2. `go build -o idea .`; scp the binary up.
3. DNS A record for idea.uk → the box IP.
4. Fill `/etc/idea/idea.env` (Anthropic key + Stripe test keys, `AUTO_DELIVER=false`).
5. `DOMAIN=idea.uk LETSENCRYPT_EMAIL=you@… bash setup.sh`.
6. Walk one test order; fold anything you had to fix back into `setup.sh`.
7. Report back what broke — that's what we harden before building `service-deployer`.
