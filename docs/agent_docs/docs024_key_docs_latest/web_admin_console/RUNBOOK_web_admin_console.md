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
| **`laptop-safe.conf`** | **← USE THIS ONE.** Desktop WireGuard client. `laptop.conf` with the `DNS =` line removed — see the correction below; the DNS line locked the owner out on 2026-08-22 |
| `laptop.conf` | the raw config exactly as the cluster generates it. **Do not feed this to `wg-quick` on a desktop** |
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
  `sudo cp ~/webdesign-admin-vpn/laptop-safe.conf /etc/wireguard/wg0.conf && sudo chmod 600 /etc/wireguard/wg0.conf`
  — **`laptop-safe.conf`, not `laptop.conf`.** Or skip `wg-quick` entirely and let
  NetworkManager own it (see below), which is the safer desktop route.
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

## ⚠ USE `laptop-safe.conf`, NOT `laptop.conf` — CORRECTED 2026-08-22, after it locked the owner out

**The `DNS =` line in the generated config broke name resolution on an Ubuntu desktop and cost
the owner his internet until he purged WireGuard entirely.** This section used to list that
line as an optional preference at the bottom of the page. It is not a preference. Removing it
is the default, and the file to use is **`~/webdesign-admin-vpn/laptop-safe.conf`**, which is
`laptop.conf` with `DNS =` and `ListenPort =` stripped.

**The mechanism, because the symptom points at the wrong thing.** `wg-quick` only calls
`resolvconf` if the config has a `DNS =` line. When it does, it runs:

```
resolvconf -a wg0 -m 0 -x
```

The **`-x` makes the tunnel's DNS exclusive for the entire machine** — every lookup, not just
cluster names. On a desktop running systemd-resolved behind the `resolvconf` compatibility
shim, that write frequently lands nowhere, and you are left with **no working resolver at
all**. The tunnel is fine; the routes are fine; the internet is fine. Only names stop working,
which reads exactly like "the VPN broke my connection".

**The proof that the DNS *server* was not the problem:** the owner changed `DNS` to
`1.1.1.1, 8.8.8.8` — public resolvers that do not need the tunnel and are reachable directly —
and resolution **still** failed. So the fault is in the `resolvconf` write, not in the choice
of nameserver. `[MEASURED 2026-08-22]` cluster DNS itself is healthy and does forward external
names: `nslookup google.com 10.21.0.10` from the wireguard pod answers correctly, as does
`admin-dashboard.…svc.cluster.local`. The `DNS =` line was not a broken idea; it is a broken
delivery path on this desktop.

**You do not need it.** The admin console is reached by IP — `http://10.21.171.225:8080`. Cluster
DNS only buys you the convenience of service names.

## Notes on the config

- `AllowedIPs = 10.20.0.0/16,10.21.0.0/16` — **split tunnel**. Only cluster traffic goes over
  the VPN; your normal browsing is untouched. This is why the lockout was DNS-only: the two
  routes added are narrow, and `ip route get 1.1.1.1` still went out over wifi throughout.
- `ListenPort = 51820` in the `[Interface]` block pins your local port. Harmless unless you run
  another WireGuard tunnel — stripped from the safe config anyway, because it buys nothing on a
  client.

## Bring it up with a dead-man's switch, at least the first time

Never bring up a tunnel you cannot get out of. This puts it up, waits, and takes it down again
whatever happens — so the worst case is a two-minute outage that repairs itself:

```bash
sudo sh -c 'wg-quick up wg0; sleep 120; wg-quick down wg0'
```

Test in a second terminal while it is up:

```bash
sudo wg show                                     # want: a recent handshake
curl -sS -o /dev/null -w '%{http_code}\n' http://10.21.171.225:8080/health   # want: 200
ping -c2 1.1.1.1 && ping -c2 google.com          # want: BOTH — the second is the DNS canary
```

If `google.com` fails while `1.1.1.1` succeeds, that is the resolvconf fault above, and the
tunnel will take itself down shortly. To recover immediately: `sudo wg-quick down wg0` then
`sudo systemctl restart NetworkManager`.

## The desktop-friendly alternative: let NetworkManager own it

`wg-quick` is a shell script that edits system state as root. NetworkManager handles the same
tunnel as a normal connection you can toggle from the network menu, which is much harder to get
stranded by:

```bash
nmcli connection import type wireguard file ~/webdesign-admin-vpn/laptop-safe.conf
nmcli connection modify laptop-safe ipv4.never-default yes    # belt and braces: never the default route
nmcli connection up laptop-safe        # and: nmcli connection down laptop-safe
```

It then appears in the GUI network menu as `laptop-safe`, with an off switch.

## The noisy errors that are NOT the problem

Every `wg-quick` run on the owner's machine printed:

```
stat: cannot read table of mounted file systems: Permission denied
/usr/bin/wg-quick: line 47: ((: ( &  & 0007) == 0: syntax error: operand expected
Warning: `/etc/wireguard/wg0.conf' is world accessible
```

**All three are one cosmetic fault and none of them stopped anything.** Line 47 is wg-quick's
own file-permission check: it runs `stat -c '%#a' "$CONFIG_FILE"` twice inside an arithmetic
expression. `stat` fails, both substitutions come back empty, the arithmetic becomes
`(( ( & & 0007) == 0 ))` — hence the syntax error — and the "world accessible" warning is the
fallback branch. **That is why `chmod 600` did not silence it**, which is the tell: the check
is not reading the mode at all.

`[INFERRED, not verified on the owner's machine]` the likely cause is **uutils coreutils**, the
Rust reimplementation Ubuntu now ships by default, whose `stat` differs here. The evidence is
in the owner's own output: `cat: /etc/wireguard/wg0.conf: Permission denied (os error 13)` —
the `(os error 13)` suffix is Rust's error format; GNU `cat` prints no such suffix. Confirm
with:

```bash
stat --version | head -1      # "uutils coreutils" vs "stat (GNU coreutils)"
```

If it is uutils, this is a known class of incompatibility with distro shell scripts. It makes
`wg-quick` noisy, not broken.

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
