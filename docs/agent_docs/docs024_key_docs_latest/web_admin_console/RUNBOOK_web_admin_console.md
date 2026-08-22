# RUNBOOK — reaching the admin console over the existing VPN

Written 2026-08-22. This is the **zero-build** path: everything below already exists and is
running. Nothing here needs a deploy, an image or a public domain.

## What you are connecting to

A WireGuard VPN that has been running in the cluster for a month
(svc `wireguard`, NodePort **31820/UDP**). It has three peers: `laptop`, `phone` and
`webdesignbox`. **The box's peer is in constant use; `laptop` and `phone` have never once
connected** — `wg show` records no handshake for either. The config files were generated
2026-07-20 and are still valid.

## Step 1 — get your config files

Already extracted to **`~/webdesign-admin-vpn/`** (mode 0700):

| file | for |
|---|---|
| `laptop.conf` | import into the desktop WireGuard client |
| `phone-qr.png` | scan with the WireGuard phone app |
| `phone.conf` | if you would rather import the file than scan |
| `laptop-qr.png` | if the desktop client prefers a QR |

⚠ **These contain private keys.** They are outside the git repo on purpose — do not move them
into `agentchassis/`, where another session's `git add -A` would commit them.

To regenerate them later:

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=wireguard -o jsonpath='{.items[0].metadata.name}')
mkdir -p ~/webdesign-admin-vpn && chmod 700 ~/webdesign-admin-vpn
for p in laptop phone; do
  kubectl -n ai-persona-system exec "$POD" -c wireguard -- cat "/config/peer_$p/peer_$p.conf" > ~/webdesign-admin-vpn/$p.conf
  kubectl -n ai-persona-system cp "ai-persona-system/$POD:/config/peer_$p/peer_$p.png" ~/webdesign-admin-vpn/$p-qr.png -c wireguard
done
chmod 600 ~/webdesign-admin-vpn/*
```

## Step 2 — install a WireGuard client

- **Linux desktop:** `sudo apt install wireguard` then
  `sudo cp ~/webdesign-admin-vpn/laptop.conf /etc/wireguard/wg0.conf`
- **macOS / Windows:** the official WireGuard app, "Import tunnel from file".
- **Phone:** the WireGuard app, "+" → "Scan from QR code", scan `phone-qr.png`.

## Step 3 — bring the tunnel up

```bash
sudo wg-quick up wg0        # or click Activate in the GUI app
sudo wg show                # you should now see a handshake
```

A successful handshake looks like `latest handshake: <a few> seconds ago` with non-zero
transfer. **No handshake means UDP 31820 is not getting out** — some cafe and hotel networks
block it; a phone hotspot is the quickest way to tell the difference.

## Step 4 — open the admin console

```
http://10.21.171.225:8080
```

That is the `admin-dashboard` ClusterIP. Verified answering `200` today.

Log in with your admin email and password. The app requires `role == "admin"` and will say
*"Admin access required"* if the account is not an admin.

⚠ `[UNVERIFIED 2026-08-22]` **whether an admin account exists for you.** The login endpoint
is alive and correctly returns 401 for bad credentials, so the chain works — but I could not
check the user list, because identity lives in an **external MySQL**
(`catalogu_vectordb_chassis:3306`), not in this cluster. If the login fails, that is the next
thing to fix, and it is not a VPN problem.

## What you can do once you are in

Sites overview · work-item queue (retry / resolve / approve) · a three-level page and
component browser · spec editing with pin and propagate · component regenerate and
section restore · media browser · customers · pipelines.

**What is missing is the build-steps view.** The API serves it
(`GET /api/v1/admin/workflows/:correlation_id`) but the SPA never calls it — 0 references in
`App.tsx`. That is the screen to build.

## Notes on the config

- `AllowedIPs = 10.20.0.0/16,10.21.0.0/16` — **split tunnel**. Only cluster traffic goes over
  the VPN; your normal browsing is untouched.
- `DNS = 10.21.0.10` — while the tunnel is up, your DNS goes to the cluster's CoreDNS. That
  lets you use names like `admin-dashboard.ai-persona-system.svc.cluster.local:8080`, but it
  also sends *all* your lookups through the cluster. If that is unwanted, delete the `DNS =`
  line and use the IP above.
- `ListenPort = 51820` in the `[Interface]` block pins your local port — harmless unless you
  run another WireGuard tunnel, in which case remove it.

## Known fragility — worth fixing before you rely on it

`Endpoint = 134.213.168.37:31820` is the address of **one worker node**
(`prod-instance-…1148`). The cluster runs on **Rackspace Spot**, where nodes can be reclaimed.
If that node goes, the tunnel stops with no warning and the fix is to point `Endpoint` at
another node's IP (`kubectl get nodes -o wide`). A LoadBalancer service or a DNS name would
remove the single point of failure.

## What you can reach, and what you deliberately cannot

A fence was applied 2026-08-22 (`wireguard-egress-containment`). Peers can reach kube-dns,
`core-manager:8088`, `auth-service:8081` and `admin-dashboard:8080` — and **nothing else,
including postgres**. If you need a new destination from your laptop, it must be added to
that policy; otherwise it fails as a timeout that looks exactly like the service being down.
Full detail and the verification probe:
`webdesign_uk_build_service/RUNBOOK_webdesign_uk_build_service.md`, "The wireguard egress
fence".
