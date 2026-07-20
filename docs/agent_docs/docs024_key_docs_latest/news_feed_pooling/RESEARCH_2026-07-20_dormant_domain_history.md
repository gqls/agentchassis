# RESEARCH — what the opaque traffic domains used to be

**Date:** 2026-07-20. Method: deep-research workflow `wf_2f8a91fd-b77` (103 agents,
83 completed, 3.07M tokens) over archive.org CDX/snapshots + live-web residue,
with 3-vote adversarial verification per claim; gaps (`zdec`, `ijih`) filled by
direct CDX pulls afterwards. 20 verifier agents and the synthesis step failed on
a session usage limit — synthesis below is hand-written from the surviving
verified claims; per-claim vote counts retained. Verification status is marked
per domain: **[verified]** = survived 3-vote adversarial check; **[unverified]**
= claim extracted but verifiers hit the limit; **[direct]** = my own CDX/snapshot
read, single-source.

## The two findings that outrank the per-domain detail

**1. Four of the nine "dormant domains we own" are not dormant-and-ours in the
expected sense — the inventory needs reconciling before anything is built:**

| domain | state found | status |
|---|---|---|
| bigotime.com | 307-redirects to an Afternic (GoDaddy) **for-sale** page | [verified 3-0] |
| buysportskit.com | same Afternic for-sale redirect reported | [unverified] |
| nanangmrk.com | **live site**: Indonesian Linux/OpenWrt tutorial archive, Cloudflare-fronted, 403 to bots | [verified 2-1] |
| ijih.com | Wayback shows **live blog captures in 2025–26** (`/about/`, `/blog/`, `/ads.txt`, `/bonvoy`) | [direct] |

If we own bigotime/buysportskit, why do they resolve to for-sale pages? If we own
nanangmrk/ijih, who is serving that content? Either the ownership list or my
reading of it is wrong somewhere, and **which one changes what we build**. The
403-to-bots finding on nanangmrk also means "hosting nothing" per our measurement
can be false — a crawler-based view of the portfolio undercounts Cloudflare-fronted
sites.

**2. High views ≠ valuable views.** zdec.com (409 — the highest of the opaque
set) is almost certainly poisoned traffic: see below. The views column measures
demand *arriving*, not demand *worth serving*; three of the top four opaque
domains have legacy traffic worth inheriting, but the single biggest number is
the one to distrust.

## Per domain

### smartbusinesssupplies.com — 748 views
**[verified 3-0 + 2-1]** Shadows **SMART Business Supplies Limited**, a real UK
B2B office-supplies dealer (Leatherhead, Surrey) whose actual site was
`smartbusinesssupplies.co.uk` — a Squarespace brochure site (stationery, managed
print, PPE, furniture, catering, facilities), live from at least 2015-06-25,
last full capture 2024-08-03, dead by 2025-03-10 (Squarespace expiry page), DNS
now gone entirely. **[unverified]** The .com's views are plausibly brand/typo
traffic from the dead company's orphaned customers.
**Rebuild:** UK office/business-supplies content and commerce — the intent
arriving is purchase-shaped, from a supplier that vanished without a forwarding
address. ⚠️ **Do not impersonate the company**: generic supplies content under
our own identity, not their brand, name or Surrey address. The company may also
still exist legally even with the site dead — check before trading near the name.
**Pool:** business-services (reverses the earlier "no-feed" call).

### zdec.com — 409 views
**[direct]** Continuous captures 2000→2018 (200s), 500 errors 2019–2022, 200s
again 2024–26. The 2017 homepage decodes as **GBK Chinese**: an industrial
control-systems company (控制芯片/control chips, 控制系统/control systems,
after-sales & repair nav) — but the page body is stuffed with **injected Macau
casino/betting spam links** (百家乐, 皇冠投注, hg0088…) and a literal "出售外链"
("backlinks for sale") + QQ contact. A hacked company site used as a link farm.
A 2023 capture shows a Chinese live-chat widget URL (`chat.7k35.com…eprId=zdec`).
**Rebuild: no usable clean legacy.** The 409 views are the residue of a
26-year-old domain plus a spam-era backlink profile that is more liability than
asset. If this domain is used at all, it should be with the expectation that
search engines may already distrust it; the traffic should be *verified human*
before anything is invested. Treat the biggest number in the opaque set as the
least valuable.

### komunikatif.com — 253 views
**[verified 3-0 ×2]** An **Indonesian-language regional news site** covering
Central Java (Semarang dateline; categories Berita/Politik/Headlines/Ekbis/
Hukrim/Hankam), actively publishing to at least 2023-05-13 (machine-readable
publish timestamp captured same-day).
**Rebuild:** Indonesian-language Central Java news/information — a Decision 13
language exception with a real former audience. This is the strongest
*news-shaped* legacy in the set: the domain was a news brand, and an
Indonesian-language feed (per-language sources, relojistas pattern) is exactly
what its residual readers were coming for. **Pool:** a future Indonesian-language
pool or dedicated sources; not the English pools.

### makeitaquote.com — 226 views
**[verified 3-0 + unverified cluster]** Never the famous service's home — the
"Make it a Quote" **Discord bot** (installed in ~1.12M servers per top.gg) lives
at `miq.moe` / `miq.suzuneu.com`, and nothing it serves references our domain.
The views are **name-collision demand**: users guessing the .com of a brand they
know, plus bot-directory listings (Botwiki names makeitaquote.com as the
associated site). A separate @MakeItAQuote Twitter/X bot shares the name too.
**Rebuild:** a web-based quote-image generator — the arriving intent is *exactly*
"turn this text into a quote image", the platform already builds interactive
tools, and the portfolio already holds memecreator.co.uk / memegenerator.uk as
siblings (a genuine future `position` case). Not a news domain at all. ⚠️ Do not
imitate the bot's branding — serve the intent under our own identity.

### buysportskit.com — 215 views
**[unverified — verifiers hit the limit]** Previously the e-commerce shop of
**BSK Pro**, a UK sports kit/teamwear supplier (Errea stockist; deep URLs still
indexed: `/club-shops/`, `/teamwear/errea-teamwear/`, `/product/errea-gerome-…`);
the brand now trades at **buysportskit.shop**. Companies House shows BUYSPORTSKIT
PRO LIMITED (10368787, inc. 2016) — name-match inference only. The domain
reportedly 307s to an Afternic for-sale page today.
**Rebuild:** teamwear/sports-kit content-and-commerce would inherit indexed deep
URLs — but **ownership must be reconciled first**, and the predecessor still
trades at .shop, so nothing that could read as impersonating BSK Pro.
**Pool:** would reverse the no-feed call (sport-retail adjacency) *if* ours.

### nanangmrk.com — 95 views
**[verified 3-0 ×3 + 2-1 ×2]** **Not dormant.** Serves a live Indonesian-language
tutorial archive today — Mikrotik RouterOS, OpenWrt, Linux/Armbian, repurposing
ex-ISP set-top boxes (ZTE B860H, HG680P) — behind Cloudflare returning 403 to
non-browser agents. The traffic source is unambiguous: the **NanangMrk YouTube
channel (551k subscribers, 64.5M views, Indonesia)** lists nanangmrk.com in its
video descriptions and told viewers to go read it.
**Rebuild:** n/a until ownership is reconciled — someone is operating this. If it
is ours, its content instrument is Indonesian networking *tutorials* (evergreen,
YouTube-companion), not a news feed.

### outfax.com — 64 views
**[verified 3-0 ×2]** A free **internet-fax web service** for ~17 years: Java
Struts app (`/fax/Main.do` login, `CreateUser.do`, `faq.jsp`), captures 200 from
2004-05-07 through 2021-08-20, died some time after. `/fax/Main.do` is still
indexed — the views are bookmarked logins and crawlers hitting a dead app.
**Rebuild:** online-fax / document-sending guides and tools ("how to send a fax
without a fax machine") — the arriving intent is task-shaped and commercial
(fax-service affiliate space is real). **Pool:** web-tech or business-services,
plus this as a dedicated tool/content theme.

### bigotime.com — 90 views
**[verified 3-0 ×2]** Currently 307→Afternic for-sale. The one indexed legacy URL
(`/index/selectLogistic?coll_id=…`) matches the route signature of **templated
dropship storefront shells**, and the CDX has **zero captures** of it — evidence
is thin by finding, not by failure.
**Rebuild: no usable legacy** on present evidence, and the ownership question
comes first. Also **corrects an earlier guess**: the relojistas audience-row note
speculated bigotime might be watch-adjacent ("big-o-time"); the evidence says
dropship shell, not watches.

### ijih.com — 31 views
**[direct]** Captures 2012→2026. The 2018 locale pages (`/de-de`, `/es-es`) are a
**domain-for-sale parking page** (Undeveloped/Dan.com, "steht zum Verkauf"); the
2017 assets are landing-page templates — so most of its life was parked. 2025–26
captures show `/about/`, `/blog/`, `/disclaimer/`, `/ads.txt` and `/bonvoy` — a
recently-operated monetised blog (the `/bonvoy` path suggests travel-points
content), operator unknown.
**Rebuild: no usable legacy.** Four-letter brandable; 31 views is near-noise.
Ownership/operator reconciliation applies here too.

## What this changes

1. **Reconciliation question for the owner** (blocking for these 9 only):
   are bigotime/buysportskit actually ours (Afternic redirects), and who operates
   nanangmrk/ijih (live content)?
2. **Pool re-assignments**: smartbusinesssupplies → business-services;
   buysportskit → sport-retail-adjacent *if* ours; komunikatif → Indonesian
   language exception; makeitaquote → tool-build, not news; outfax → web-tech
   with a fax-tools theme; zdec/ijih/bigotime → stay no-feed.
3. **Decision 14 dedicated sources**: komunikatif (Indonesian Central Java news)
   and smartbusinesssupplies (UK business-supplies trade news) are the two with
   feed-shaped legacies deserving dedicated sources at onboarding.
4. **Measurement caveat fleet-wide**: a Cloudflare-fronted site can look empty to
   crawler-based measurement (nanangmrk's 403s). "Hosting nothing" claims about
   any domain need a browser-shaped check before being believed.
5. **Impersonation lines**: smartbusinesssupplies and buysportskit both shadow
   real, findable businesses (one dead, one trading). Inheriting *intent* is
   legitimate; inheriting *identity* is not. Any rebuild on these serves the
   arriving demand under our own branding, with no use of their names, marks or
   addresses.
