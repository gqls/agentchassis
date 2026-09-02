# Where we are — Nominet domain management

## 2026-09-02 — the lane opens, and immediately finds four domains going dark

You asked for a single home for everything Nominet, so this is it: the tag, the
EPP connection, the domain list, registrations, transfers, and moving domains'
nameservers.

First thing found: the nameserver change you made at Nominet for the four remake
domains — advertise.co.uk, designblog.co.uk, seotools.co.uk and
websitepromotion.co.uk — points them at Cloudflare, but no Cloudflare zones
exist for them yet. Until a zone exists, Cloudflare refuses to answer for those
names, so as the internet's caches empty over the next day or two each domain
stops resolving entirely. advertise.co.uk is already unreachable from here. The
old Drupal site appearing to still work earlier today was just stale cache.

The fix is ready and safe to re-run: one command creates the four zones and
wires them the same way as every other portfolio domain. My session isn't
allowed to run it (it involves the Cloudflare key), so please run, in this chat:

    ! scripts/domains/cf-zone-bootstrap.sh advertise.co.uk designblog.co.uk seotools.co.uk websitepromotion.co.uk

That makes the four domains answer again (serving the portfolio router; the
remade sites appear on them as each build lands).

Two things only you can move, when you have a moment:

1. **The domain list.** Your ~1,500 .uk domains at Nominet have never been
   listed out — every other registrar got counted today, Nominet is the one
   still missing. Easiest: export the CSV from Nominet Online Services and drop
   it somewhere I can read. Alternatively run the prepared walk (runbook §2).
2. **The second tag.** The application went in on 11 August and we've heard
   nothing since — has anything arrived from Nominet?

Small tidy-up, no rush: the old read-only Cloudflare key (the file called
`token`) has stopped working entirely. Everything now uses the newer
`portfoliotoken`, so nothing is broken — but if you meant to keep two working
keys, that one needs remaking.
