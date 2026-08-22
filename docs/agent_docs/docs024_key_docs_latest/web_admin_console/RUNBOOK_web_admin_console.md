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

---

## Rotating a peer's keys WITHOUT restarting the wireguard pod

Done for `peer_laptop` on 2026-08-22, after its preshared key was pasted into a chat
transcript. Recorded because the obvious method is the dangerous one.

**Why not the obvious way.** The documented way to re-key a `linuxserver/wireguard` peer is to
delete `/config/peer_<name>/` and restart the pod so the entrypoint regenerates it. **Do not do
that here.** The same instance carries `peer_webdesignbox`, and that tunnel is the webdesign.uk
box's only route to `core-manager` — which serves the chat bot's facts relay. The bot is built
to **refuse to start** without it. A restart also drops every other peer and rebuilds
`wg0.conf` from `PEERS`, which risks renumbering.

**Scope the rotation first — do not assume peers share key material.** Each peer has its own
preshared key, so a leak of one is not a leak of all:

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=wireguard -o jsonpath='{.items[0].metadata.name}')
kubectl -n ai-persona-system exec "$POD" -c wireguard -- sh -c '
for p in laptop phone webdesignbox; do
  printf "%-14s psk-sha256=%s pub=%s\n" "$p" \
    "$(sha256sum < /config/peer_$p/presharedkey-peer_$p | cut -c1-16)" \
    "$(cat /config/peer_$p/publickey-peer_$p)"
done'
```

`[MEASURED 2026-08-22]` all three digests differed, so only `laptop` needed rotating.

**Take a baseline of `wg show` before you touch anything** — specifically the box peer's
handshake age and transfer counters. They are how you prove afterwards that you disturbed
nothing, and you cannot take that measurement retrospectively.

**The rotation itself** (full script: this session's transcript; the shape is what matters):

1. `cp -a` the peer dir and `wg_confs/wg0.conf` to timestamped backups under `/config/`.
2. `wg genkey` / `wg pubkey` / `wg genpsk`; write `privatekey-`, `publickey-`,
   `presharedkey-peer_<name>`.
3. Rewrite `peer_<name>.conf`. **Omit `DNS =` and `ListenPort =`** — see the correction above;
   the DNS line is what took the owner's desktop off the internet.
4. `qrencode -o peer_<name>.png -t png < peer_<name>.conf`.
5. Patch **only** the target stanza of the persisted `wg_confs/wg0.conf`, with `awk` keyed on
   the `# peer_<name>` comment — not a global `sed`, which would rewrite every peer.
6. Apply to the running interface, which touches that peer alone:
   ```bash
   wg set wg0 peer "$OLDPUB" remove
   wg set wg0 peer "$NEWPUB" preshared-key /tmp/psk allowed-ips 10.13.13.2/32
   ```
   `wg set` takes the preshared key as a **file path**, never a literal.

**Verify four things, and the last two are the ones people skip:**

```bash
# 1. old key gone, new key present at the right address
kubectl -n ai-persona-system exec "$POD" -c wireguard -- wg show

# 2. DURABILITY — keyfile, persisted wg0.conf and live interface must all agree,
#    or the rotation is cosmetic and a pod restart silently reverts it
kubectl -n ai-persona-system exec -i "$POD" -c wireguard -- sh -s <<'CHK'
LIVE=$(wg show wg0 peers | tr '\n' ' ')
for n in laptop phone webdesignbox; do
  F=$(cat /config/peer_$n/publickey-peer_$n)
  C=$(awk -v n="# peer_$n" '$0==n{f=1;next} /^\[Peer\]/{f=0} f&&/^PublicKey/{print $3;exit}' /config/wg_confs/wg0.conf)
  [ "$F" = "$C" ] && S=AGREE || S=MISMATCH
  case "$LIVE" in *"$F"*) L=on-interface;; *) L=NOT-LIVE;; esac
  printf '  %-14s keyfile-vs-wg0.conf=%-9s interface=%s\n' "$n" "$S" "$L"
done
CHK

# 3. the copy you hand over really carries the new key — DERIVE it, do not trust the file
kubectl -n ai-persona-system exec -i "$POD" -c wireguard -- sh -c \
  "awk -F'= ' '/^PrivateKey/{print \$2}' | wg pubkey" < ~/webdesign-admin-vpn/laptop.conf

# 4. the BOX is undisturbed — end to end, not just on the interface
curl -sS -X POST https://preview.webdesign.uk/api/chat -H "Content-Type: application/json" \
  -d '{"message":"How long does a site take?"}'
```

Check 3 matters because every other check reads the *server's* view. It is the only one that
proves the file in the owner's hands is the file that works.

**Rollback:** the timestamped backups under `/config/.bak_peer_<name>_*` and
`/config/.bak_wg0_*.conf`. Restore the files, then re-apply with the same `wg set` pair
reversed. **The old key is dead the moment `wg set … remove` runs** — a client still holding it
cannot handshake, which is the point.

**Result 2026-08-22:** `laptop` moved from `1kw5qf…` to `NvmN0w92…`; `phone` and
`webdesignbox` untouched; box peer still handshaking (55s) with unchanged counters; bot, site
and `/c/` all verified answering afterwards.

⚠ **`phone.conf` still carries a `DNS =` line** and has *not* been rotated (its preshared key
never leaked). On a phone that line is safe — mobile WireGuard apps scope DNS to the tunnel
rather than writing it system-wide through `resolvconf`, which is the desktop-only fault above.

---

## Troubleshooting: tunnel comes up, but nothing works (2026-08-22, NARROWED — the loss is upstream of the client)

Symptom on the client: `wg-quick up` succeeds, routes are added, DNS is fine, but `wg show`
lists the peer with **no `latest handshake` and no `transfer` line at all**, and anything
addressed into the cluster times out.

**Read those two absent lines carefully — they are the whole diagnosis.** WireGuard learns a
peer's address from its first *valid* packet. On the SERVER, the peer showed **no endpoint, no
handshake, zero bytes received**, while the `webdesignbox` peer beside it kept handshaking
normally throughout. So this is not a key problem, not a policy problem and not a routing
problem: **nothing from the client is arriving at all.**

### What was ruled out, and how

| candidate | verdict | evidence |
|---|---|---|
| The `wireguard-egress-containment` fence | **NOT the cause** | `policyTypes: ["Egress"]` only — it cannot drop inbound handshakes. And if it dropped the *reply*, the server would still show an endpoint and non-zero rx for that peer. It shows neither |
| Wrong / mismatched keys after the rotation | **NOT the cause** | The peer's public key is on the interface at `10.13.13.2/32`, and the preshared key digest is **identical** across all three copies: the keyfile, the live interface (`wg showconf`) and the client's own file |
| Wrong server public key or endpoint in the client conf | **NOT the cause** | Client's `PublicKey` matches the server's; `Endpoint` interpolated correctly to `134.213.168.37:31820` |
| `externalTrafficPolicy: Local` sending traffic to a node without the pod | **NOT the cause** | Policy is `Cluster`, **and** the pod is on `prod-instance-…1148`, which *is* `134.213.168.37` |
| The node being unreachable from that network | **NOT the cause** | TCP to `134.213.168.37:30080` connects from the same laptop; a control port, `30099`, is properly refused |
| UDP egress being blocked wholesale | **NOT the cause** | `dig @1.1.1.1` over UDP/53 works from the same machine |

### ⚠ A probe that CANNOT answer this, so do not use it

The obvious test — `nc -u` to the port and watch for an ICMP port-unreachable — **is useless
here, and its own control proves it.** Probing `1.1.1.1:5399`, which is certainly closed,
produced exactly the same silence as the open port. ICMP unreachables are not returning to this
network at all, so "no ICMP back" carries no information: open, closed and filtered are
indistinguishable. **A probe whose control returns the same answer as its subject has measured
nothing.**

### The measurement that WILL settle it — needs root on the client

Watch whether the handshakes even leave the machine. Run the capture first, in its own
terminal:

```bash
sudo tcpdump -ni any 'udp port 31820'
```

Then, in another terminal, bring the tunnel up with the dead-man's switch and force traffic:

```bash
sudo sh -c 'wg-quick up wg0; sleep 60; wg-quick down wg0'
curl -m 10 -sS -o /dev/null -w '%{http_code}\n' http://10.21.171.225:8080/health
```

Read it as follows — the three outcomes are genuinely different faults:

| tcpdump shows | meaning | next step |
|---|---|---|
| **nothing at all** | WireGuard is not even trying. Client-side config or routing | check `ip route get 10.21.171.225` returns `dev wg0` |
| **outbound packets, no replies** | packets leave; the path or the far end drops them | try another network (below); then a different node IP |
| **outbound and inbound** | the tunnel is working and the fault is above it | check the egress fence's allowlist for the destination |

`PersistentKeepalive = 25` has been added to the client config, so the tunnel now sends every
25s on its own — you no longer need to generate traffic to test it, and the server side becomes
a live indicator. Watch it from the repo directory with
`~/webdesign-admin-vpn/watch-server.sh`, which polls `wg show wg0 dump` for this peer.

### The cheap discriminator before any of that: a different network

The `phone` peer is configured, valid and untouched by the rotation. Connect it **over mobile
data, not wifi** (scan `~/webdesign-admin-vpn/phone-qr.png`).

- **Phone works, laptop does not** → the home network or its router is dropping UDP/31820. The
  fix is a different transport, not a different config.
- **Phone also fails** → the fault is at the node or in front of it, despite the box's tunnel
  being healthy, and the next question is what is different about the box's path.

⚠ The phone config still carries a `DNS =` line. That is safe on a phone — mobile WireGuard
apps scope DNS to the tunnel instead of writing it system-wide through `resolvconf`.

### UPDATE — the client is exonerated, with packet-level evidence

`tcpdump` on the client, while the tunnel was up:

```
15:44:49.565364 wlp0s20f3 Out IP 192.168.0.45.57504 > 134.213.168.37.31820: UDP, length 148
15:44:55.372408 wlp0s20f3 Out IP 192.168.0.45.57504 > 134.213.168.37.31820: UDP, length 148
...
```

**Outbound only, five retries, nothing inbound.** 148 bytes is exactly a WireGuard handshake
initiation, so the client is behaving correctly and the packets do leave the NIC.

Simultaneously, on the server, `wg show wg0 dump` for that peer: `endpoint=(none)`,
`last_handshake=0`, `rx=0` — while `peer_webdesignbox` handshook **11 seconds earlier** with
6.5 MB received. Two peers, one instance, one port; one works continuously and the other has
never delivered a byte.

**And the client's network passes UDP on high ports perfectly.** Three STUN binding requests,
each with a validated reply — `stun.l.google.com:19302` (112 ms), `stun1.l.google.com:19302`
(127 ms), `stun.cloudflare.com:3478` (152 ms). So outbound UDP works, replies return, and NAT
is not the problem. **This matters because it is the check that would have exonerated the
network in the wrong direction too** — had the STUN probes failed, the answer would have been
"the home network blocks UDP" and nothing further would have been needed.

So the packets leave a working network, and never arrive. **The loss is in transit or at the
node**, and it is not:

- the client config (proven at the packet level),
- the client's network (proven by STUN),
- the keys (PSK digest identical in all three places),
- the egress fence (`Egress`-only; and a dropped reply would still populate the endpoint),
- node selection (`externalTrafficPolicy: Cluster`, and the pod is on the very node addressed),
- anything in our terraform (no firewall or security-group resource mentions 31820 at all).

### Next test: is the block node-specific?

`~/webdesign-admin-vpn/try-nodes.sh` sweeps all five node IPs as the `Endpoint`, 20 s each, and
**checks the SERVER's `wg show` dump rather than the client's** — because the client cannot
distinguish "no reply yet" from "never arriving", and the server can.

```bash
sudo -E ~/webdesign-admin-vpn/try-nodes.sh
```

It prints the working endpoint if one lands, and says so plainly if none does. If none does,
the next step is a capture inside the pod (`apk add tcpdump` — ephemeral, cleared on restart)
to separate "arrives but is rejected" from "never arrives", which the `rx` counter alone cannot
do: **`rx` counts only valid, authenticated packets**, so a malformed or mis-keyed arrival is
indistinguishable from an absent one at that counter.

### ⚠ The strategic point, before spending more on this

We already run a **proven** internet-to-cluster path that does not use UDP or a NodePort at
all: the island (`tools.apis.uk`), live since 2026-07-24 — Cloudflare tunnel, outbound-only, no
inbound ports. And the webdesign.uk box **already holds a working tunnel into the cluster** and
already proxies to `core-manager`.

So there is a route to browser-based admin that needs no VPN on any device: put
`admin-dashboard` behind the box's nginx with an access control in front. That is exactly
options **B/C** in `PLAN_2026-08-22_web_admin_console.md` §1, and it is gated on owner decision
**D-A**. Continuing to debug WireGuard is reasonable, but it is fixing the path we do *not*
have working while a path we *do* have working sits unused.
