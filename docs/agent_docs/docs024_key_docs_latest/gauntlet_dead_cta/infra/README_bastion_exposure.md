# tools-api public exposure — bastion + tunnel design (P3)

Owner decisions (2026-07-23): engine in the k8s cluster; shared API domain on
**apis.uk**; a **bastion host** between the internet and the cluster; Cloudflare
at the edge; sites stay static.

Owner decisions (2026-07-24, "bastion host" session): subdomain = **tools.apis.uk**;
bastion on a **UK-based/owned provider** (recommendation below); apis.uk
nameservers already moved to Cloudflare.

## Zone state (verified live 2026-07-24)

- `dig NS apis.uk` → `alexis.ns.cloudflare.com` / `leah.ns.cloudflare.com` —
  the zone is ACTIVE on Cloudflare, not just "directed".
- A **proxied wildcard record** exists: `*.apis.uk` (and the apex) all resolve
  to Cloudflare IPs and serve **error 525** (origin TLS handshake fails — the
  wildcard points at a dead origin, probably imported when the zone was added).
  Delete the wildcard before creating the tunnel route; nothing real serves today.

## Path
```
browser → https://tools.apis.uk (Cloudflare: TLS, WAF, edge rate-limit)
       → Cloudflare Tunnel (outbound-only from the bastion; NO inbound ports)
       → bastion VM (cloudflared + Caddy: only /api/v1/tools/* forwarded,
         1MB body cap, per-IP rate limit; holds NO k8s credentials)
       → WireGuard to a DEDICATED in-cluster instance `wireguard-bastion`
         (NodePort 31821/UDP — see correction below)
       → tools-api ClusterIP service (pin spec.clusterIP in the PR)
```

## CORRECTED 2026-07-24 — the original peering design did NOT protect the cluster

The first draft said: peer the bastion onto the existing admin WireGuard and
restrict it with a NetworkPolicy keyed on the bastion's WG address
(`ipBlock <WG_BASTION_PEER_IP>/32`). **Both halves were wrong**, found by
reading the live cluster (this session):

1. **The existing `wireguard` pod MASQUERADES.** Verified in the running pod:
   `iptables -t nat -S POSTROUTING` → `-A POSTROUTING -o eth+ -j MASQUERADE`.
   Every peer's traffic reaches other pods carrying the **WireGuard pod's IP**,
   not the peer's WG address. The ipBlock policy could therefore NEVER match;
   under the namespace's `default-deny-ingress` it would have blocked
   everything — fail-closed, but not a working design.
2. **Any peer of that instance has full-namespace reach.** ai-persona-system
   carries `allow-same-namespace` (`podSelector {}` ← `podSelector {}`): all
   pods accept ingress from all pods in the namespace. Because peer traffic
   masquerades to the WG pod (itself in the namespace), a peer reaches EVERY
   service — including `postgres-clients:5432` (the `database-access-policy`
   app-label allowlist is unioned away by `allow-same-namespace`). Right for
   the owner's laptop/phone; unacceptable for an internet-adjacent bastion.

**The fix — a dedicated, contained WireGuard instance** (`wireguard_bastion.yaml`):

- Second tiny `linuxserver/wireguard` deployment `wireguard-bastion`, NodePort
  **31821/UDP**, internal subnet **10.13.14.0/24** (main WG uses 10.13.13.0/24),
  exactly ONE peer (the bastion). Client-side AllowedIPs = the tools-api
  ClusterIP /32 only.
- **In-pod containment** (one-time edit of the generated `wg0.conf` PostUp):
  forward from wg0 ONLY to `<TOOLS_API_CLUSTERIP>:<TOOLS_API_PORT>`, drop the
  rest. Cryptokey routing already stops source spoofing (peer may only source
  its /32).
- **Calico egress NetworkPolicy on the wireguard-bastion pod** — the layer that
  does not trust the pod itself: enforced at the node, it lets the pod open
  connections ONLY to `app: tools-api` on the service port. Even a compromised
  or misconfigured WG pod reaches nothing else. (Calico enforces egress here;
  verified the cluster runs Calico, policies active 359d.)
- Result: bastion compromise ⇒ ability to POST to tools-api, nothing more. The
  claim the first draft made is now actually true.

**Never add the bastion as a peer of the main `wireguard` deployment.**

## Bastion provider (owner asked for UK-based/owned, cheap, reliable)

- **Recommended: Mythic Beasts** — independently UK-owned for 25 years
  (Cambridge/London DCs), ISP-grade, developer-friendly; small VPS is a few
  pounds/month. The reliability pick.
- **Budget: Clouvider** — UK-based (London DCs), VPS from ~£3.50/mo, 100% SLA
  marketing. Fine for a box whose job is two daemons.
- **Hetzner fallback fails the UK test**: German-owned, and no UK datacentre —
  a Hetzner box lands in Germany/Finland/US/SG.
- Spec needed: smallest tier anywhere — 1 vCPU / 1GB / Debian 12. No inbound
  ports required (tunnel dials out; WG dials out to NodePort 31821).

## Owner tasks (updated 2026-07-24)

1. ~~Name the subdomain~~ → **tools.apis.uk**. ~~Confirm Cloudflare zone~~ → ACTIVE.
2. Pick the provider (above) and provision the VM (Debian/Ubuntu, smallest tier).
3. In Cloudflare DNS for apis.uk: **delete the `*` wildcard record** (dead 525 origin).
4. Approve the dedicated `wireguard-bastion` deploy (31821/UDP NodePort) —
   applied only after the tools-api PR lands and fixes port + ClusterIP.
5. Tunnel: either give session an API token to script it, or run the runbook
   below on the bastion (~10 min, one browser login).

## Tunnel runbook (locally-managed, config in this repo)

On the bastion VM, as root:
```bash
# 1. install cloudflared (Cloudflare's apt repo)
mkdir -p --mode=0755 /usr/share/keyrings
curl -fsSL https://pkg.cloudflare.com/cloudflare-main.gpg \
  | tee /usr/share/keyrings/cloudflare-main.gpg >/dev/null
echo "deb [signed-by=/usr/share/keyrings/cloudflare-main.gpg] https://pkg.cloudflare.com/cloudflared any main" \
  > /etc/apt/sources.list.d/cloudflared.list
apt-get update && apt-get install -y cloudflared

# 2. authenticate (prints a URL — open it in a browser, pick the apis.uk zone)
cloudflared tunnel login

# 3. create the tunnel + DNS route (CNAME tools.apis.uk → <UUID>.cfargotunnel.com)
cloudflared tunnel create tools-api        # note the UUID; writes ~/.cloudflared/<UUID>.json
cloudflared tunnel route dns tools-api tools.apis.uk

# 4. install config (this repo's cloudflared_config.yml) + credentials
mkdir -p /etc/cloudflared
cp cloudflared_config.yml /etc/cloudflared/config.yml   # set credentials-file to the <UUID>.json path
cp ~/.cloudflared/<UUID>.json /etc/cloudflared/tools-api.json

# 5. run as a service
cloudflared service install
systemctl enable --now cloudflared
```
Cloudflare dashboard afterwards (apis.uk zone): SSL/TLS → Full (strict);
Always Use HTTPS on; one rate-limiting rule on `tools.apis.uk/*` (free plan
includes one); Free Managed WAF ruleset on.

## Files here
- `cloudflared_config.yml` — tunnel → localhost Caddy (hostname filled: tools.apis.uk)
- `Caddyfile` — bastion reverse proxy: ONLY /api/v1/tools/* forwarded, 1MB body
  cap, per-IP rate limit, everything else 404; proxies to the tools-api
  ClusterIP through the WG tunnel
- `wireguard_bastion.yaml` — the dedicated WG instance (deployment + PVC + NodePort)
- `networkpolicy_tools_api.yaml` — CORRECTED: Calico egress containment on the
  WG pod (the enforcement) + ingress documentation on tools-api

Placeholders remaining: `<TOOLS_API_CLUSTERIP>`, `<TOOLS_API_PORT>` — fixed by
the feature-builder PR (ask the PR to PIN `spec.clusterIP` so the Caddyfile,
wg0.conf PostUp and egress policy never drift). NOT applied anywhere yet.
