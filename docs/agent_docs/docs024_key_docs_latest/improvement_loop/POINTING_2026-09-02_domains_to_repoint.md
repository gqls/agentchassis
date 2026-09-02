# Domains to point — 2026-09-02

Produced by the improvement_loop lane at the owner's request ("I need to point them
manually, please get a lane to tell me which to point").

**Answer: two domains. Both need their NAMESERVERS changed, not their A records.**

> **⚠ CORRECTED 2026-09-02, same day, BEFORE the owner acted — the two domains are NOT
> in the same state, and the first version of this document said they were.**
>
> I wrote "40 pages of finished work are built and waiting". That is true of ONE of them.
>
> - **boxingonline.com — 21 pages BUILT AND DEPLOYED**, most recently `2026-09-02
>   13:59:46Z`, under an hour before this correction. The artefacts exist and are being
>   refreshed; only the delegation is missing. **Pointing this domain WILL make the site
>   appear.**
> - **adversecreditmortgage.co.uk — ZERO pages ever deployed.** All 19 are `planned`
>   (18) or `needs_rebuild` (1), every one with `deployed_at IS NULL`. **Pointing this
>   domain will show NOTHING**, because there is nothing at the other end yet. It needs a
>   build first, and pointing it is the second step, not the first.
>
> Why this matters more than the tidiness of the number: point them both, see one work
> and one stay blank, and the natural conclusion is that the pointing failed. It did not.
> **The count was real; the claim I hung on it was not** — I had the page counts from
> `pages.status='active'`, which says a page is wanted, and read it as a page that exists.
>
> This also RETRACTS the detector-gap suspicion at the foot of this file and in NOTES §(n).
> `head_essentials_missing` filed nothing for adversecreditmortgage **because its
> eligibility gate is `PageHasShippedPredicateFor`, and not one of those 19 pages has ever
> shipped.** The check was right to stay silent. There is no gap; there is an unbuilt site.

---

## The list

| domain | pages built | what it serves now | nameservers now | what it needs |
|---|---|---|---|---|
| **boxingonline.com** | 21 built **and deployed** (latest 09-02 13:59Z) | 114-byte stub redirecting to `/lander`, on every path | `ns1.afternic.com`, `ns2.afternic.com` | **delegate — and the site appears** |
| **adversecreditmortgage.co.uk** | 19 planned, **0 ever deployed** | same 114-byte stub, every path | `ns1.dan.com`, `ns2.dan.com` | **build FIRST**, then delegate — pointing it today shows nothing |

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

~~`adversecreditmortgage.co.uk` filed no such findings at all... That is a detector gap.~~
**RETRACTED the same day — see the correction box at the top.** The check skipped it
because `loadStructuralPopulation` gates on `PageHasShippedPredicateFor`, and all 19 of
that site's pages have `deployed_at IS NULL` with `build_status` in
(`planned`, `needs_rebuild`). A check that probes served pages correctly says nothing
about a site that has never served one. **No gap. An unbuilt site.**
