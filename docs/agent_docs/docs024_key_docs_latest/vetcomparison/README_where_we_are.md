The push to the live shared repo is being gated by the permission classifier — twice — and that's the one irreversible, outward step, so I'll stop and hand it to you rather than try to route around it.
Everything up to that point is done and verified.

Here's exactly where things stand.

The strip is built, verified, and ready — it just needs your push. On the sites repo I created branch strip-vetcomparison off current origin/master and cherry-picked the strip cleanly onto it. It's a
plain fast-forward — origin/master plus one commit — touching only vetcomparison.uk/ paths, nothing else. To take the live fabrication offline, run:

! git -C /home/ant/projects/sites push origin strip-vetcomparison:master

That triggers deploy-to-b2.yml and the live site updates. Until it runs, vetcomparison.uk is still serving invented prices — I confirmed origin/master still has the 1.25 MB fabricated index; my strip only
exists locally. (Heads-up on that repo: your local clone is 1,701 commits behind origin — a stale manual checkout. I deliberately did not touch local master or the other commits; the clean branch
sidesteps all of it.)

What the strip contains, all verified by serving it locally: every price and the price UI gone; the medicine calculator and all price-bearing data files deleted; the false "proprietary data / do not
scrape" claim gone; the two CMA guides pulled (one had a fabricated quote attributed to the CMA, another an unsourced price claim against a named practice — same liability class as the directory); and the
directory itself regenerated from 2,579 genuinely verified practices in the database rather than the prototype's list of uncertain origin. The homepage is now an honest directory with a
claim-your-listing call to action.

Still to do on the database: you approved quarantining the 997 fabricated seed_import price rows (they sit on 235 real verified practices). That's a separate write to production I haven't made yet — I'll
do it once the live push is sorted, unless you'd rather I do it now in parallel.

The CMA research came back authoritative, and it sharpens the strategy. Sourced to the actual Final report (24 March 2026):

- The Order hasn't been made yet — statutory deadline 23 September 2026. Price lists go live ~Dec 2026 (large groups) / ~Mar 2027 (independents). There's a genuinely time-sensitive item: a consultation
  closes 30 July (~2 weeks), but it's the funding Order (the RCVS levy), not the price remedies.
- The mandated price list is 36 specific services in 5 categories with 6 weight bands, VAT-inclusive, no free text allowed — which means the comparison columns are defined for us and scraped pages will be
  unusually clean to parse. This should be the schema, not anything we invent.
- On your copyright instinct: the CMA explicitly treats scraping as legitimate and even names an existing comparison site (VetHelpDirect) — but nothing compels practices to allow scraping, and the
  database-right question is genuinely untested (get an IP opinion before republishing at scale). The claim-your-listing model remains the clean path.
- Strategic catch worth knowing now: the "no paid rankings" rule binds only sites that take the official RCVS data feed — a self-scraping site isn't bound, but you can't hold the "RCVS-approved" badge and
  run paid placement. And the RCVS is building its own comparison tool, so "better than the regulator's" is the bar.

Once you've run the push (or told me to hand it off differently), I'd like to lay out the real build as a proper plan — the generic, config-driven directory/price exporter on the chassis (parameterised by
vertical + domain, so it serves the next comparison site too, per your constraint), the 36-item schema, and the claim flow. Want me to quarantine the DB rows now while you handle the push?

----
Fable

Review done, quarantine executed, and the legal record is written. Here's what came out of it.

The one thing that still needs you: the push hasn't happened. I re-checked the live site first — it is still serving the fabricated dataset (the invented Ardene House prices, the calculator, all of it),
and the repo history shows it has been public since 2 February 2026. The strip commit is sitting ready as fbc0b929 on strip-vetcomparison, exactly one commit ahead of current origin/master. When you run:

! git -C /home/ant/projects/sites push origin strip-vetcomparison:master

the deploy workflow takes it live. Every day before that push extends the exposure window the legal doc now documents — it's genuinely the most urgent item on the board.

What I verified fresh, and what changed on re-review:

- The database quarantine is now done and verified. 997 fabricated current price rows → 0; all seed rows retained for audit; the 803 genuine rows (every one with a source URL) untouched.
- Two clobber risks checked and cleared. There is no sites table row for vetcomparison, so the chassis render loop cannot overwrite the hand-stripped page (a real hazard given the re-render landmines
  we've hit on other sites). And every med/vet scheduled task — including med-export-json with its hardcoded .co.uk default — is disabled, so no automation can republish prices.
- CMA facts re-grounded first-hand against the case page and the CMA's business guidance (both fetched yesterday's-date fresh). The load-bearing claims all held: final report 24 March 2026, Order not yet
  made (statutory deadline 23 September 2026), price lists ~Dec 2026/Mar 2027, £21 prescription cap, RCVS feed to approved third parties ~September 2027, funding-order consultation closing 30 July. Two new
  details surfaced: practices must also publish a parasiticides price list (flea/tick/worm products — directly relevant to the medicine-comparison arm), and price lists must sit at most one click from the
  homepage, which makes compliance-detection scraping much easier. The weight-band discrepancy also resolved: six pet categories, but cat and small dog share the <10kg band — I've flagged it for
  verification against the final Order before we freeze the schema.
- One correction to how we describe the "legal imperative" in marketing and planning: the precise obligation is that vets must publish standardised prices and submit data to the RCVS; nothing compels them
  to use a third-party comparison site. Our pitch to vets is therefore "you already have to publish — claiming your listing here makes that compliance work twice as hard for you," not "the law says you
  need us."

The legal doc is at docs/agent_docs/docs024_key_docs_latest/vetcomparison/LEGAL_2026-07-15_vetcomparison_factual_record.md. As asked, it states facts only, no liability conclusions: what was published and
when (2 Feb 2026, itemised including the fabricated CMA quote and the Taylor Vets claim), how the fabrication was detected (the statistical fingerprint, zero source URLs, the animeddirect.co.uk domain
error, the seed-file lineage), the remediation table with dates and commit hashes, the grounded regulatory context with sources, and the publication policy now in force. It honestly records the deploy as
pending — update that row once you've pushed. Its §8 also lists the two things a solicitor should look at before per-practice prices return at scale: the record itself, and the untested database-right
question.

The strategy itself survived the re-review intact: evidence-not-publication scraping, claim-your-listing as the licence mechanism and the product, aggregates and dated deep-links otherwise, and everything
built generic on the chassis (config-driven exporter per vertical and domain). Whenever you're ready, the next piece is the build plan for that exporter and the claim flow — say the word and I'll plan it
properly. 

---
All four decisions are now baked into the plan, and the consultation briefing is written. Here's what changed where.

The plan (PLAN_2026-07-15_rebuild.md) — the old "Standing decision points" section is now "Decisions (owner, 2026-07-16)" with all four resolved, and the details are pushed into the phases where an
implementer will actually meet them:

1. Attributed prices ON — the Phase 2 exporter config now has attributed_prices: true with five enforced rules: prices only from the practice's own domain (checked against their recorded website, never a
   third-party compilation), always with source URL and capture date, opted-out practices excluded, quarantined fabrication excluded unconditionally, and claimed figures always superseding scraped ones. The
   disclaimer text is written verbatim in the plan so the implementer uses it as-is — it cites the CMA's recognition of comparison services and offers the email opt-out, with a compact per-price label ("From
   the practice's price list, {date} · source ↗ · correct/remove").
2. No RCVS badge — Phase 5 now states the independent-operator decision, with the two consequences an implementer must respect: any future paid placement must be clearly labelled (the CMA's not-misleading
   standard still applies to everyone), and we monitor the approval criteria anyway for competitive intelligence about what badge-holding rivals may do.
3. min_n = 3 — set in the exporter config, with the n always published beside each statistic so a reader can judge the sample, and a note to revisit if small-area reverse-inference ever becomes plausible.
4. Consultations: responding — recorded with your stance, pointing at the new briefing.

Phase 3 also gained the opt-out mechanism your decision 1 requires: an publication_optout flag on the businesses table, operator verification of the requester, prompt re-export on opt-out, and the rule
that opting out removes price display only — the practice stays in the directory and in unnamed aggregates, and a later claim reverses it.

The briefing (CONSULTATION_2026-07-16_briefing.md) is written to be read straight through: how the CMA turns its final report into binding Orders, the two consultations (funding — open now, closes 30
July, via the CMA portal; substantive — still not published as of today, expected imminently), how responding works (who can, the portal, what gets published, what a response can realistically achieve),
and then our positions under two declared hats. Four arguments for the independents — a size-scaled levy instead of the flat ~£500/yr that hits a single-site practice hardest, closing the three-month
"corporates look transparent first" window with free templates, hard enforcement of ownership disclosure, and keeping the no-free-text rule absolute. Four for us — proportionate third-party approval
criteria, an express right to reuse the mandated price lists (the one sentence that would close our database-right ambiguity), machine-readable structure, and treating scrape-blocking of mandated lists as
non-compliance. One flag inside it: the levy figures (~£150–250 setup, ~£450–550/yr) come from our earlier document review and must be re-verified against the draft Order PDF before they go in any
submission.

The legal record now carries your decision contemporaneously: attributed publication proceeds ahead of solicitor review, on the narrowing conditions above, with the review kept as an open advisory item.
That's the honest version of events if anyone ever asks.

Two timing notes. The funding consultation is the only clock running — if you want the levy-scaling point made, it's a one-page portal submission before 30 July, and I can draft it whenever you say. The
substantive draft Order should appear within days for the CMA to hit its September deadline; when it does, I'll turn the briefing's positions into a clause-referenced draft response for your sign-off —
nothing gets submitted without it.

--