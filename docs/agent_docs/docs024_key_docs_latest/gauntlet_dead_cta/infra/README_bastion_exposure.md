# tools-api public exposure — bastion + tunnel design (P3)

Owner decisions (2026-07-23): engine in the k8s cluster; shared API domain on
**apis.uk** (owner names the exact subdomain — placeholder `<SUB>.apis.uk` below);
a **bastion host** between the internet and the cluster; Cloudflare at the edge;
sites stay static.

## Path
browser → https://<SUB>.apis.uk (Cloudflare: TLS, WAF, edge rate-limit)
       → Cloudflare Tunnel (outbound-only from the bastion; NO inbound ports)
       → bastion VM (cloudflared + Caddy: route allowlist, size caps, rate limit;
         holds NO k8s credentials)
       → WireGuard (precedent: admin-dashboard peering, 31820/UDP)
       → tools-api ClusterIP service (NetworkPolicy: WireGuard/bastion path only)

## Owner tasks (blocking, in order)
1. Name the subdomain (`<SUB>.apis.uk`) and confirm apis.uk is a Cloudflare zone.
2. Provision the bastion VM (small; Debian/Ubuntu; no public inbound needed —
   the tunnel dials out).
3. Approve WireGuard peering bastion↔cluster (new peer on the existing WG setup).
4. Create the Cloudflare Tunnel (one-time `cloudflared tunnel create tools-api`)
   + DNS route `<SUB>.apis.uk` → the tunnel. (I can script this given API access,
   or it's ~5 minutes in the dashboard.)

## Files here
- `cloudflared_config.yml` — tunnel → localhost Caddy
- `Caddyfile` — the bastion reverse proxy: ONLY /api/v1/tools/* forwarded,
  1MB body cap, per-IP rate limit, everything else 404
- `networkpolicy_tools_api.yaml` — cluster side: tools-api accepts traffic only
  from the WireGuard peer address
All three carry placeholders (<SUB>, <WG_CLUSTER_IP>, <TOOLS_API_PORT>) to be
filled when the owner tasks land and the service design (feature-builder) fixes
the port. NOT applied anywhere yet.
