# PLAN — idea.uk demand (started 2026-07-29, session "idea.uk vm 9")

## The brief (from the vm-site handoff, owner-directed 2026-07-28)

idea.uk is complete, correct, and has never had a customer who is not its owner.
Find out whether anyone wants it. The engineering queue on the product is EMPTY;
this workstream's job is buyers, not features.

**Rails inherited from the product work, binding here too:**
- No overstating, anywhere. Acquisition copy survives the same bar as the site
  (`bugs_open/043`, the "no figure in any brief" rail).
- Outward-facing prose uses the plain-human style
  (`travelling_docs/pitch_pdf_source/` style prompt).
- The measured baseline (18–28 Jul, bots filtered): 26 views of /report.html
  from 20 IPs, 8 form submissions — **dominated by us and bots. Never compute a
  conversion rate from it.** Genuine external buyers to date: zero.
- Outward actions (posting, mailing, listing, spending) get owner sign-off
  first. Analysis, engineering and drafts don't need it.

## What we found on day one (all evidence in NOTES, 2026-07-29)

1. **Google's index entry for idea.uk is the domain's PREVIOUS life: a
   Dan.com "Domain Name For Sale" parking page.** A `site:idea.uk` search
   returns that, not the site. The one organic path a buyer has to us currently
   says the site does not exist.
2. **Real Googlebot (rDNS-verified, 66.249.0.0/16 — 253 hits ever) has never
   crawled /report.html or any guide.** It fetches `/`, assets, legal one-offs,
   and a 404 robots.txt (107 times).
3. **No robots.txt, no sitemap.xml** (both 404). 7 of 21 pages have EMPTY meta
   descriptions — including the homepage, the one page Google does crawl.
4. Crawler/referrer log data is heavily polluted: fake-Googlebot vulnerability
   scanners and referrer spam. Only rDNS-verified conclusions are usable.

## Phases

### Phase 1 — be findable (engineering, this session)
The demand experiments are pointless while the front door is mislabelled.
1. robots.txt + sitemap.xml, committed to `gqls/vm-sites` under `idea.uk/`
   (sitesync is `rsync --delete` from that repo — files dropped on the box are
   deleted within 5 min). Sitemap uses the site's own canonical URL form
   (`/guides/x/index.html`, matching internal links).
2. Meta descriptions for the 7 empty pages (`sql/` in the vm_site dir, p4_33) +
   plain (assemble-mode) rerenders — assemble mode rebuilds the `<head>` from
   `pages.meta_description` and cannot LLM-escalate, verified against
   `rerender_single_page_action.go:357` before choosing it.
3. OWNER: Google Search Console + Bing Webmaster verification, then request
   re-indexing of `/` to kill the "Domain For Sale" snippet. We prepare exact
   steps; the owner clicks. THIS IS THE HIGHEST-LEVERAGE SINGLE ACTION.

### Phase 2 — put the offer in front of people (owner-gated, drafts this session)
The product's own advice, applied to the product: pick a channel, put the offer
in front of a countable number of relevant people, count. Candidate channels
(each gets a draft + a measurement plan before any owner sign-off):
- The £8 example place (10 places, 0 used) as the demand experiment: honest
  outreach in UK founder/small-business communities.
- Direct outreach where "is my idea any good?" is already being asked.
- Small paid test (~£50–100) only if organic channels say nothing — owner call.

### Phase 3 — read the result honestly
Attribution BEFORE volume: every channel gets a distinguishable path or UTM so
the nginx log can attribute a visit. Success is a genuine external order (money
moved); interest signals (views, form starts) are secondary and bot-filtered by
rDNS, not by user-agent string.

## Decisions

- **D1 (2026-07-29):** This session (vm 9) runs both the specimen refresh and
  the demand thread — owner picked both via AskUserQuestion.
- **D2 (2026-07-29):** Specimen provenance = reword + refresh (owner call).
  Done in the vm_site lane: `sql/p4_32`.
