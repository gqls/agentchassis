# Domains to point — 2026-09-02

Produced by the improvement_loop lane at the owner's request ("I need to point them
manually, please get a lane to tell me which to point").

**Answer: two domains. Both need their NAMESERVERS changed, not their A records.**

---

## The list

| domain | pages built | what it serves now | nameservers now | what it needs |
|---|---|---|---|---|
| **boxingonline.com** | 21 active | 114-byte stub redirecting to `/lander`, on every path | `ns1.afternic.com`, `ns2.afternic.com` | delegate to the estate's Cloudflare pair |
| **adversecreditmortgage.co.uk** | 19 active | same 114-byte stub, every path | `ns1.dan.com`, `ns2.dan.com` | delegate to the estate's Cloudflare pair |

**Target delegation** — every serving estate domain uses the same pair, so this is the
value to set, verified on cookly.uk, farmerinsurance.uk, agritec.uk and webdesign.uk:

```
alexis.ns.cloudflare.com
leah.ns.cloudflare.com
```

## Why it is the nameservers and not the A record

Both domains resolve today to `13.248.169.48` and `76.223.54.146`. Those are the parking
IPs of the domain marketplaces whose nameservers are still authoritative — Afternic for
one, Dan.com for the other (the same operator; hence the identical IPs). **Setting an A
record at the registrar would not take effect**, because the domains are not delegated to
where the record would live. The delegation is the fix; the addresses follow.

Read plainly: both were bought and never moved off the marketplace's DNS. Nothing on our
side is broken, and nothing needs rebuilding — **40 pages of finished work are built and
waiting behind a delegation that was never changed.**

## How this was found, and how to re-check it

`./probe_serving.sh` (this directory) probes every active estate domain and classifies it
from the served bytes rather than from `sites.status` — both of these are recorded
`deployed`, which is why no status query would ever have found them.

⚠ **The control is the load-bearing part.** A parked domain answers **200 to every path**,
including its front page, so a plain root probe reports it healthy. The script fetches an
invented path on the same domain alongside the root; identical bytes for both means the
host answers anything, and that is the signature.

`[MEASURED 2026-09-02]` 34 active domains probed: **31 SERVING**, **2 PARKED**, 1 flagged
for a look (see below).

## One more thing, not a pointing job

**noted.co.uk** answers 200 with the home page for URLs that do not exist — a soft 404.
Real pages are fine and distinct (`/privacy.html` and `/about.html` return their own
titles and byte counts), so this is not a delegation problem and needs no action from
you. It is worth one look by the `noted_rebuild` lane, because a soft 404 tells search
engines every mistyped URL is a real page. **Not this lane's to fix.**

## What this changes for the improvement loop

The 20 `head_essentials_missing` findings against boxingonline.com — "no title, no
skip-link, no footer" — are **true statements about a parking stub**, not about our
pages. They should clear on their own once the domain is pointed and the checker re-probes
a real page. I will confirm that rather than assume it.

`adversecreditmortgage.co.uk` filed **no such findings at all**, despite serving the same
stub. It is one of only two active sites with none. **That is a detector gap, not a clean
bill of health** — the check's page-eligibility gate is skipping it for a reason I have
not yet established. Logged in NOTES §(n) as unmeasured; it is on this lane's list.
