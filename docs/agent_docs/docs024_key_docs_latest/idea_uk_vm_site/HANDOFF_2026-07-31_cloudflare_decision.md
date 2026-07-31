# HANDOFF — idea.uk is not behind Cloudflare, and §4a of the runbook is built on the belief that it is

**Written 2026-07-31 by the `webdesign_uk_build_service` lane, which measured this
while answering a DNS question about a different domain. This lane is not mine —
I have corrected the facts and left every decision to its owner.**

Read with: `RUNBOOK_idea_uk_vm_site.md` (now carries a dated correction at line 12
and under §4a) · `LANDMINES.md` → *"idea.uk is NOT behind Cloudflare"* ·
`webdesign_uk_build_service/NOTES_…` (2026-07-31 entries, the raw measurements).

---

## 1. The finding

`RUNBOOK_idea_uk_vm_site.md:12` states `FRONT nginx + Let's Encrypt, DNS
(Cloudflare) → the VM`, and §4a opens *"idea.uk is behind Cloudflare"* under the
heading **"Restore the real client IP — nothing else works until this is done."**

That section also instructs the reader to *"first confirm whether the DNS record is
actually proxied (orange) or DNS-only (grey)"*. **Confirmed 2026-07-31 ~19:20 UTC:
neither. Cloudflare is not in idea.uk's request path at all.**

```bash
$ dig +short NS idea.uk
oxygen.ns.hetzner.com.   helium.ns.hetzner.de.   hydrogen.ns.hetzner.com.

$ dig +short A idea.uk
116.203.204.115                       # the VM itself, bare — no anycast in front

$ curl -sS -o /dev/null -D- https://idea.uk | grep -i '^server\|cf-ray'
server: nginx/1.28.3 (Ubuntu)         # and NO cf-ray header
```

The zone is delegated to **Hetzner**, not Cloudflare, so the orange/grey question
does not even arise — there is no Cloudflare zone to be orange or grey.

**`relojistas.com`, the other VM site, IS proxied** (`server: cloudflare`, CF
anycast A records, Cloudflare NS). So **the two VM sites do not share a front-end
shape**, though the estate docs treat them as one pattern. Any reasoning that
moves a conclusion from one box to the other by similarity is unsafe.

## 2. What it invalidates, and what it reveals

**§4a's premise is false, so its prescribed work is unnecessary — but the section
was pointing at something real, and that thing is worse than it thought.**

| §4a said | actually |
|---|---|
| `$binary_remote_addr` is a Cloudflare edge IP | it is the **true client address** |
| the `limit_req` zone (`setup.sh:86,226,299`) buckets all traffic as one | it is **already correct** and rate-limits per visitor |
| a `geo` deny or fail2ban jail would ban Cloudflare | it would ban **the actual spammer** |
| Cloudflare WAF/Turnstile is *"reachable as the blocking layer"* | **there is no WAF, no Turnstile and no DDoS layer at all** |

So: **the remediation is not needed, and the exposure is larger than the document
describes.** A live, earning, card-taking service is on the open internet with
nginx `limit_req` as its only abuse control. That is the finding worth acting on,
and it replaces §4a rather than being a footnote to it.

**One consequential side effect, easy to trip over:** `RUNBOOK:435`'s *"purge the
Cloudflare cache for idea.uk"* is a **no-op**. Debugging a stale page, you would
run it, see no change, and conclude the purge had failed or the deploy had not
landed — chasing a cache that does not exist.

## 3. The decision this lane's owner needs to take

**Both options are coherent. The dangerous outcome is the middle one — moving the
box behind Cloudflare and treating the follow-up work as optional.**

### Option A — leave it direct

- Strike §4a; its work is unnecessary. Correct `RUNBOOK:12` and `:435`.
- Accept: no WAF, no bot protection, no DDoS absorption, and **the origin IP is
  public** (it is the A record).
- Cost: nothing. Risk: unchanged from today, but now *known* rather than assumed
  away.

### Option B — move it behind Cloudflare (recommended by me, but it is not my call)

Gains the WAF, Turnstile, bot-fight and DDoS absorption that §4a assumed existed.
**Two things stop being optional the moment you do it**, and both fail silently:

1. **Real-IP configuration** — `set_real_ip_from <22 published CF ranges>` +
   `real_ip_header CF-Connecting-IP` in nginx. Without it, §4a's described failure
   becomes **actually true**: every visitor keys to a Cloudflare edge address and
   `limit_req` degrades into one global bucket **that still looks like a working
   rate limiter**. The estate plan already specifies this module —
   `vm_estate/PLAN_2026-07-25_framework_controlled_vm_estate.md:44`
   (`cloudflare_realip`) — so do not hand-roll it.
   > **The discriminating check is `count(DISTINCT ip) > 1` from TWO networks.**
   > One test machine cannot tell a constant from a working key. This is the
   > `bugs_open/139` landmine: the island's per-IP limiter was keyed on a constant
   > (sha256 of the docker gateway, **83/83 rows identical**) and read as fine.
2. **Firewall 443/80 to Cloudflare's ranges only.** Otherwise the origin IP —
   already public, and additionally exposed by having been used as the origin for
   `webdesign.uk`/`ugg2.com` on 07-31 — is directly reachable and every attacker
   simply bypasses the WAF. A proxy in front of an open origin is decoration.

**Sequence, if B:** add the zone at Cloudflare → set records **grey/DNS-only**
first and verify the site still serves → apply the real-IP nginx config and prove
it with the two-network check → **then** go orange → **then** firewall the origin
→ then re-test checkout end to end. Do not go orange first; the rate limiter is
wrong for the whole window between orange and step 3.

**A third option worth pricing before choosing:** `cloudflared` tunnel instead of a
proxied A record — the profile the estate plan already documents for the Mythic
Beasts box (`PLAN:187`, *"cloudflared tunnel, no inbound"*). It removes the
firewall step entirely (nothing can reach the box but the tunnel), makes
`CF-Connecting-IP` **unforgeable** rather than merely conventional, and matches the
owner's own pull-only island ruling. Costs a daemon on the box and a dependency on
the tunnel staying up.

## 4. What I changed, and what I deliberately did not

**Changed** — facts only, appended, never rewritten:
- `RUNBOOK_idea_uk_vm_site.md:12` — inline `[CORRECTED 07-31: NOT Cloudflare — see §4a]`.
- `RUNBOOK_idea_uk_vm_site.md` §4a — a dated correction block appended **below** the
  existing text, with the evidence and the three consequences. The original
  paragraph is untouched and still readable.
- `LANDMINES.md` — a fleet-wide entry so the next session to touch this box gets
  the check before it has a symptom (synced into `doc_notes`).

**Not changed, because they are this lane's calls, not mine:**
- §4a's body and its ordering of work — it is wrong about the premise but it is
  the owner's plan; deleting someone's security section on a live earning service
  from an adjacent lane is not a correction, it is a decision.
- The security posture itself. Nothing was reconfigured on the box. **No command in
  this handoff has been run against `116.203.204.115` except read-only probes.**
- `PublicBaseURL`, the Stripe configuration, and the `Host` finding below.

## 5. One thing found in passing that this lane should decide on

**The engine does not validate the `Host` header.** `grep -n 'r\.Host' main.go
service.go billing.go` → **no match**; redirect targets come from the configured
`PublicBaseURL`.

On 2026-07-31 the owner briefly pointed `webdesign.uk` and `ugg2.com` at
`116.203.204.115`. All three domains then served **byte-identical** idea.uk
content (`md5 cf4c46c2b4e0`), and because there is no `Host` check the shop was
fully functional on both — **a real, payable order was creatable from
`webdesign.uk`**, with the buyer bounced mid-checkout to a domain they had never
visited. **Not fired: doing so would have created a live order and a real Stripe
session.** Stated as reachable, not as exercised.

Those records have since been removed by the owner. **The exposure was the
configuration, not the domains** — any hostname pointed at this box reproduces it.
A `Host` allow-list (or an nginx `default_server` returning 444) closes it
permanently and costs a few lines. Recorded as a fleet landmine; the fix is this
lane's call.

## 6. Verify any of the above in one command

```bash
for d in idea.uk relojistas.com; do
  echo "=== $d ==="; dig +short NS $d; dig +short A $d
  curl -sS -o /dev/null -D- -m 12 "https://$d" 2>&1 | grep -iE '^server|cf-ray'
done
```

`cf-ray` present ⇒ proxied. `server: nginx/...` with no `cf-ray` ⇒ you are talking
to the origin directly. **Run it per domain** — sharing an owner, a provider or a
runbook does not mean sharing an ingress.
