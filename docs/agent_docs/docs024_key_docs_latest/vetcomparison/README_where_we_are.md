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

Here's where things stand after this stretch.

The consultation draft is ready for your review at docs/agent_docs/docs024_key_docs_latest/vetcomparison/CONSULTATION_RESPONSE_funding_DRAFT_2026-07-16.md. Before drafting I verified the key fact from the
consultation portal itself: the levy is charged flat per practice site, which sharpened the argument — a 500-site group pays 500×, but per site a single-vet independent bears the same charge as one branch of a
national chain with a central finance office. The draft asks for a small-business rate tied to the CMA's existing under-15-sites threshold (so no new definitions needed), plus an RCVS transparency requirement on
levy spending, and stays deliberately inside this consultation's narrow scope. The amounts (£150–250 setup, £450–550/year — CMA guidance, corroborated by Vet Times) are marked for you to confirm against the Notice
PDF in the portal before submitting — deadline 30 July, and it goes out only when you submit it; nothing leaves without your sign-off.

Phase 0 of the build is complete, and it ran deeper than the plan expected:

- The unowned-domain hazard is dead at every layer. The Go default is gone and the action now fails closed without an explicit domain (unit-tested, passing); the same .co.uk value turned out to also be seeded in
  the agent definition and the scheduled task in production — both blanked, per your go-ahead. The Go part takes effect on the next chassis deploy; until then the blanked configs and disabled task cover it.
- The "cosmetic" name cleanup uncovered real rot. Of 36 suspect names: 17 were real practices with scraped page-title names (now cleaned — "Eastcott Vets: Award-winning Swindon & Marlborough Vets Near Me" is just
  "Eastcott Vets" again), but 20 weren't practices at all — Yelp, StarOfService, ThreeBestRated and BestLocalRated listing pages, a US clinic, a college course page — all sitting in the live directory as "verified
  practices". Dismissed. A wider sweep then found 177 more entries whose "website" was a wheree.com mirror page rather than the practice's own site — real practices, unsound verification — now marked pending for
  genuine re-verification.
- The live directory is now 2,389 practices, every website the practice's own, re-exported and verified live. The plan records the follow-ups this generated: extend the sweep pipeline's deny-list with the mirror
  domains we found, and use the 176 wheree practices as the first live run of re-verification.

One process note worth knowing: mid-cleanup, a push went out before its integrity check because set -e doesn't reliably abort scripts in this harness. No harm resulted (that push was still strictly an improvement,
and the gate caught the issue on the next cycle — including one final straggler, "BestVets In Newcastle", whose spelling had dodged my regex). But the lesson is recorded in the plan and memory: verify and push now
happen as separate, gated steps.

Next up is Phase 1 — the insertPrice rewrite onto the unified schema, applying migration 006, and seeding the CMA's 36-item taxonomy that everything downstream maps to. That's the biggest single chunk of Go work in
the plan; say the word and I'll start it, or it's cleanly specified for a cheaper model if you'd rather switch now.


----

Phase 1 is complete — the unified price schema is live in production, the CMA taxonomy is seeded, and the Go cutover is written and tested. Here's the shape of it.

The Go rewrite (Go-B, the piece that blocked this since May). insertPrice no longer writes the deprecated business_prices table — it upserts a canonical products row (kind service) and records the observation in
product_prices, retiring prior current rows for that practice-and-service pair. A new insertMedicinePrice handles the verifier's medicine_prices[] output the same way (kind medicine), so per-practice medicine
prices — the V1.5 dream in the old handoff — now have a write path. The price reader cut over to the unified schema with an unchanged output shape, so no caller changes. The subtle risk in all this was slug drift:
migrated rows and future live writes must land on the same product rows, so I pinned the Go slug function byte-identical to the migration's SQL expression with unit tests, including a cross-check run against real
production values (trailing hyphens and all). Build and the full actions test suite pass.

Migration 006 applied. 512 distinct service offerings created from the historical data, 1,953 price observations migrated, 762 current. Three integrity results worth knowing: zero rows were unmigratable, the 41-row
gap between old and new "current" counts is the migration's designed same-instant dedup (originals remain in the old table, which is deprecated-not-dropped), and — the one I most wanted to see — zero quarantined
fabrication rows came through as current in the unified table.

The CMA taxonomy seeded (007, new migration in the vetcomparison docs dir). All 36 mandated services as canonical product rows — 12/6/6/9/3 across the five categories, checkbox-disclosure metadata on exactly the
four starred surgical items, and a pet_band dimension on price observations modelling the six CMA pet categories. Every scraped, claimed, or future RCVS-feed price now has a fixed regulatory vocabulary to map onto.
It carries a header note to re-verify wording against the final Order before launch.

The one open caveat: the Go changes aren't deployed. The chassis image builds from the local filesystem, and the working tree also holds other in-flight changes (including the article-body fixes marked
deploy-pending from another workstream) — so building an image now would bundle those, and that's your call rather than mine. It's not urgent: the verifier tasks are all disabled so nothing writes prices today, and
006 is idempotent — if the old binary ever wrote a stray business_prices row in the interim, re-running 006 sweeps it up. When you're ready to deploy, the new code rides along with whatever else is going out.

Next in the plan is Phase 2 — the generic directory_export_json action and its orchestrator pair, which replaces the hand-run export I've been doing via psql and bakes in the publication rules you decided
(own-domain-only attribution, opt-out exclusion, min_n=3 aggregates, claimed-supersedes-scraped, seed_import excluded unconditionally). Say the word and I'll keep going.

---

Done — Phase 2 is built and seeded, and the full documentation set is written. The summary of this stretch:

Phase 2 (the generic exporter) is code-complete and seeded, pending deploy. The new directory_export_json action is fully config-driven and fail-closed: it refuses to run without an explicit domain and vertical,
attributed prices default off (the vetcomparison task config turns them on per your decision), aggregates enforce the n≥3 floor in SQL, opted-out practices are excluded, and seed_import rows are excluded
unconditionally. The agent pair is seeded and the scheduled task is disabled until a chassis image containing the action ships. Along the way I found and fixed a pre-existing landmine: med_export_json was never
registered in the action registry at all — the med exporter agent could never have dispatched its own action.

The smoke test surfaced the most important data fact of the day. Three outputs validated perfectly against production (directory exactly matches the live 2,389; aggregates produce 15 publishable rows across 14
areas; claimed is empty as expected). Attributed came back zero — and that's correct behaviour, not a bug: every historical price row has an empty source URL (my earlier "all 803 rows have provenance" was wrong —
the NULL check passed on empty strings), and nothing is recoverable from the observation records either. So under your no-provenance-no-publication policy, historical prices feed aggregates only, and per-practice
attributed prices begin with fresh scrapes through the new write path — which means any future scraper/verifier must persist per-price source URLs, now written into the plan as an acceptance criterion.

The documents you asked for, all in docs/agent_docs/docs024_key_docs_latest/vetcomparison/:

- SUMMARY_2026-07-16_where_we_are.md — the full arc: the fabrication found and killed, the regulatory grounding, your four decisions, the honest rebuild, Phases 0–2, and the road ahead through the consultation
  dates to the Sep 2027 RCVS horizon.
- RUNBOOK_vetcomparison.md — operator procedures: the sites-repo deploy pattern, how to enable the exporter after deploy (including the image-tag bump it needs), manual claim and opt-out handling, the hard
  provenance rule, DB access, and every date with a clock on it.
- RUNNING_NOTES_vetcomparison.md — the session-by-session operational log, newest first.
- PLAN updated in place — Phases 0, 1 and 2 marked done with their verification numbers and findings.

Standing items for you: the funding-consultation response awaits your review and portal submission before 30 July; the substantive draft Order should appear on the CMA case page any day (I'd draft the
clause-referenced response as soon as it does); and the next chassis image build ships all of Phases 0–2, after which the exporter can be smoked and enabled per the runbook. Next build phase when you want it: Phase
3, the claim flow — the exporter already publishes claimed prices the moment the first practice is claimed.

---

Adoption is triggered, mostly worked, and hit a real blocker at the final step — here's the honest state.

Your instinct was half right. The site was already hosted by chassis infrastructure, but never registered with it — no sites row, so the build loops and improvement sweeps didn't know it existed. That registration
is what adoption does, and the May handoff had planned it but never got there.

What happened: I fired the adoption (vetcomparison.uk adopting itself as ground truth, fidelity locked so our hand-built pages are treated as canonical). The pipeline crawled the live site, fingerprinted it,
analysed it, classified the archetype and derived the content direction — all successfully. It then failed at the very last step, apply_adoption_plan, with a database error: the work-item insert expects a unique
index (idx_swi_dedup) that doesn't exist in production.

Why, and why I stopped rather than patched it: the code comment traces it to migration 157, written yesterday by the fixloop workstream — it changed the work-item dedup rules, and the Go and schema halves of that
change must land together, byte-matched, or every keyed insert fails exactly like this. Hand-creating the index now, without confirming which version of the clause the deployed binary emits, has a coin-flip chance
of recreating the same error from the other side. Your own build practice notes say verify against the pod, never git — so this waits for the deploy.

No damage: the sites row exists (harmless), no specs or work items were written, and the live site is untouched.

The unblock is the chassis deploy you already have queued, which is now carrying three workstreams' worth: the fixloop migration-157 Go, our Phases 0–3, and the fix for this. The sequence after deploy is in the
running notes: apply the 157 schema in the same window, re-run the same adoption command (safe to re-trigger), then bump the exporter agents' image tags and enable the export per the runbook. One standing rule I've
recorded from the adoption-faithfulness doc: once adoption succeeds, never re-plan this site to fill gaps — only the first-plan branch is faithful today, and re-plans can silently clobber adopted pages.

Also worth saying plainly: it's now the 17th, so the funding consultation closes in 13 days.

---

● CLAUDE.md read and in force — and following it immediately paid off. Here's what it changed and where things now stand.

What the file told me that mattered: this repo and cluster are worked by multiple concurrent Claude sessions — one working tree, one image sequence, one live database — with strict rules: pathspec commits per task,
forward-only git, builds from committed HEAD only, deploys verified against the running pod, and a queue check before dispatching anything at the cluster.

Applying it surfaced three facts I'd otherwise have got wrong:

1. My Phases 0–3 work had been swept into other sessions' commits (f51a7accc, d076c3c8e, 37468ba65) — the exact hazard the doc documents. Per its ruling: forward-only, nothing lost; I committed my remainder
   narrowly with pathspec commits and said so in the message.
2. The deployed chassis does contain our code. I'd assumed v1.0.1130 predated our changes; pod verification (strings on the binary, per the doc — never trust the tag) shows directory_export_json and
   insertMedicinePrice are in the running binary. The sweeping commits had, ironically, shipped us. The fleet moved to v1.0.1134 mid-session — other sessions are deploying continuously — and it still carries our
   symbols.
3. The adoption blocker was gone: idx_swi_dedup now exists in production with the cancelled predicate, landed by the fixloop session with its paired Go.

So I re-ran the adoption, by the book — queue coverage check first (clear), pod age outside the 300-second dispatch window — and this time it completed end-to-end: crawl, fingerprint, analysis, classification, and
apply_plan seeded all five adoption specs. The dispatch loop had already run the domain-research-classifier before I'd finished watching; the site is classified (as a hub), the spec cascade is current, and the
build pipeline now owns vetcomparison.uk. Phase 4 — the step that stalled in May and failed on Wednesday — is done.

Two watch items carried into the notes and memory: the classifier emits no content_features (the lost May patch never landed anywhere, so a news feed needs a one-off manual spec patch if you want one), and the
first build-cascade rerender needs checking against the live site to confirm the hand-authored pages survive the faithful pass — with the standing rule unchanged: never re-plan this site to fill gaps.

One deliberate deferral: enabling the directory exporter (bump the two directory-* agents' image tags from v1.0.1126 to current, one kcat smoke run, then enable) is ready to go per the runbook, but with the build
cascade actively working the site right now and the fleet mid-deploy every few minutes, firing an exporter commit into the same window felt like avoidable contention. It's a ten-minute job whenever you say.

The consultation clock still stands: funding response closes 30 July, draft awaiting your review.


---
The exporter is live, and with it the last hand-cranked part of the data path is gone. Summary of this stretch:

First autonomous publish, done and verified. I bumped the two exporter agents to the pod-verified v1.0.1134, fired one smoke run through Kafka using the scheduled task's exact payload, and it completed in seconds:
queried production, built the five JSON artefacts, committed them through the git-adapter (ac3314fd), and the deploy served them. Until now, every data update on this site has been me running psql and pushing by
hand; from here the platform does it.

Every publication rule held under real conditions. The directory came out at exactly 2,109 (matching the dedupe); the aggregates file has 13 rows, each with its n shown and none below 3 — down from 15 in the
pre-dedupe smoke because removing priced duplicates dropped two area groups below the floor, which is precisely the behaviour you'd want; claimed and attributed prices are both empty arrays, correctly, since
nothing yet has consent or provenance; and the metadata file carries the policy statement. The moment a first practice claims, or a provenance-carrying scrape lands, those files populate on the next cycle with no
further work.

The schedule is on — every 48 hours, same config, and the smoke run stamped the task's completion time so the bookkeeping loop is proven too.

Where that leaves the whole build: Phases 0 through 4 are complete and operational — schema unified, CMA taxonomy seeded, claim flow proven, site adopted onto the chassis with its cascade actively building
(classifier, tool recreation and two content pages already done), and the exporter publishing on schedule. What remains from the plan is Phase 5, the provenance-first scraping that fills the attributed-prices file
— the biggest open build item, and sensibly a fresh session's work.

Two watch items, one of them yours: the rerender work item is still queued, and when it fires I want to confirm the hand-authored homepage and guides survive the first faithful pass — I'd check the live site after
it completes. And the funding consultation closes on the 30th — the draft response has been sitting ready for your review since Wednesday; the site it names as our operation is now something you can point the CMA
at without a caveat.

---

Both documents are written and committed, and the pre-handoff state check turned up something important that's now recorded in both.

The handoff is at docs/agent_docs/docs024_key_docs_latest/vetcomparison/HANDOFF_2026-07-18_vetcomparison_uk.md. A new thread can start cold from it: verified current state, the hard rules that must not be relaxed
(no price without provenance or consent, never re-plan this site, the sites-repo deploy pattern, CLAUDE.md's multi-session discipline), open items in priority order, and all the operational pointers. Pointing a
fresh session at that file is enough.

The read-out summary is beside it (SUMMARY_2026-07-18_readout.md) — the full arc in plain prose: what we set out to achieve, the fabrication we found and how we dealt with it, the regulatory grounding and your four
decisions, what got built in four days, where it stands, and where it goes. It's written to be read aloud straight through.

What the state check found: the adoption cascade ran the entire build overnight — strategy, design, seven imagery items, new about/contact pages, and full rerenders. The audit results cut both ways. The good: the
three guides survived at their original URLs with our sourced content intact, and a price audit across every rebuilt page found no unsourced figures anywhere — the only prices on the whole site are the CMA's own
£21/£12.50. The bad: the rebuilt homepage dropped the directory search UI and the claim/opt-out section. The data files and all the machinery behind them are untouched — it's a page-markup restoration — but it's
the site's core function plus our licensing funnel, so it's the handoff's top priority, ideally fixed at spec level so the next render doesn't revert it. There are also seven build items waiting for human review in
the admin queue, which only you can work through.

The two dated items remain: the funding consultation closes 30 July (draft ready for your review and portal submission), and the substantive draft Order should appear on the CMA case page imminently.

Ready for the new thread whenever you are.

---

I need to correct what I told you an hour ago. Re-reading CLAUDE.md sent me to check the rendered page rather than the work-item statuses, and the truth is worse than my handoff says.

My earlier diagnosis was wrong. I reported the directory search and claim CTA were "gone". They aren't. The chassis planned the homepage sensibly — its five sections include a filtered-result-grid and a
call-to-action — and it built a genuinely good search component with filtering, sorting, pagination and region controls. What it did not do is wire that component to our data. The live homepage says "No results
found." The directory doesn't work at all.

And the rebuilt page has acquired claims we cannot back. This is the part that matters, because it's the same failure class we spent four days remediating — not fabricated prices this time, but fabricated
capability:

- It describes the site as "a directory of UK veterinary practices with pricing information, ownership data…" — we publish neither per-practice. We deliberately stripped ownership fields on 15 July because they
  were unverified.
- It offers a whole panel headed "Ownership and Group Information — see whether a practice is independently owned…". We don't publish that.
- The results grid has "Price: Low to High / Price: High to Low" sort controls, on a directory with no published prices.
- And it carries a disclaimer calling the directory "a representative sample for demonstration and comparison purposes" — which is untrue in the other direction: these are 2,109 real verified practices, not a demo
  sample.

So right now the site is both broken and overclaiming, and every work item behind that reads complete. That's exactly CLAUDE.md's warning: trust the rendered artefact, not the status. I had verified the guides and
the price audit against the live site, which is why I caught "no unsourced prices" correctly — but I checked the homepage only for the presence of markers, not for what it actually says. That was a real gap in my
check and I'd rather flag it than let it stand.

The good news: our working hand-authored homepage is intact in git (b2896815), one restore away, and the chassis's search component is well-built — it needs pointing at /data/vet-full-index.json, not replacing.

Two ways forward, and the choice is yours because it's a platform question you know better than I do:

The fast path is restoring the known-good homepage over the rendered one. The site works again within minutes, and per the adoption-lock design a hand edit takes a permanent lock, which protects it from the next
rerender. The cost is that we're back to a hand-maintained page the chassis doesn't understand.

The structural path — which CLAUDE.md's "structural fixes over patches" favours — is to fix it at spec level: wire the filtered-result-grid to our directory JSON and strip the unbacked pricing/ownership/demo-sample
copy from the section content, so the next render produces the right page rather than reverting. Slower, and the site stays broken meanwhile.

My recommendation is both, in that order: restore now so the site works and stops overclaiming today, then fix the spec so the restoration isn't fighting the platform. Shall I go ahead on that basis?

I'll also correct the handoff either way — as written it sends the next thread hunting for a missing search box that isn't missing.

---

Both read-out documents are written, the emergency is contained, and the handoff no longer carries my wrong diagnosis.

The urgent thing first: the site is clean again. I restored the verified homepage and confirmed against the published page that the generated practices are gone, the real 2,109-practice directory is wired back in, the
claim and opt-out routes are back, and none of the unsupported pricing/ownership/demo-sample claims remain. That took priority over writing anything, because fabricated veterinary practices were live while we talked.

Two documents for you to read out, both in the vetcomparison docs folder:

SUMMARY_2026-07-18_bugs_journey.md — the bugs account you just asked for. Ten of them, in the order they mattered: the founding fabrication and its statistical tell (one in five practices charging exactly £48); the
exporter aimed at a domain you don't own; an action that was never registered and so could never run; a directory containing Yelp pages and a college course; 280 duplicates from one missing URL normalisation; prices
whose provenance was never recorded — including the check of mine that gave a false pass by testing for null instead of empty; adoption blocked by a missing database index; and then the big one, the platform
regenerating fabricated data by itself. It ends on the pattern that connects nearly all of them: a system reporting success while the artefact it produced was wrong.

SUMMARY_2026-07-18_readout.md — the aim, what we've done, where we are, where we're going, rewritten to tell the truth about today rather than yesterday's optimistic version.

What I'd want said plainly: today was a correction, not an advance. The platform we adopted the site onto rebuilt it overnight with invented practice data — the exact defect the whole project exists to remedy,
reintroduced four days after we removed it, by our own tooling, with every work item reporting success. It is contained but not fixed. The homepage restore is a hand edit; the underlying specification still describes a
search section with no data source, so a future automated render could regenerate it. That spec-level fix is the first job, and this needs filing as a platform bug rather than a site bug, because an agent that invents
data when it cannot reach the real source is dangerous on every site, not just this one. Both are written into the handoff as top priority.

I also owe you a correction on my own work: an hour ago I told you the rebuild had dropped the search box. It hadn't — the search was there and well built, filled with fake practices. I found that only because you sent
me back to CLAUDE.md, whose "trust the rendered artefact, not the status" is precisely the check I'd skipped.


---

Yes — it's bug 020: /bugs_open/020_HANDOFF_2026-07-18_tool_recreation_invents_data_when_it_has_no_source.md.

To be straight with you: when you asked, it didn't have a number. I'd written "this needs filing as a platform bug" and left it at that, which is exactly the kind of intention that evaporates when a thread ends. It's
filed now, along with the transferable pattern in the debugging guide's §9 and a row in the bug index.

Filing it forced a proper root cause, which I hadn't had. I'd been describing the symptom ("an agent that invents data"). Reading the actual prompt and the adoption path, it's two compounding structural defects:

The recreation path has no data-dependency contract. Adoption's fingerprinting step captures that a tool is interactive, but never carries its fetch() target through to the tool-recreation agent, whose brief is to
produce self-contained HTML/CSS/JS. So a tool whose behaviour is its data — a directory search — literally cannot be recreated faithfully. The model's only options are a dead empty widget or invented records that make
search, filter and pagination visibly work. It chose the latter and documented the decision in a comment.

And the prohibition that should have stopped it is scoped to arithmetic. Rule 9 of that prompt reads "No fake data or dummy outputs — calculations must be mathematically correct", sitting among rules about function
completeness. Read in context it's a statement about calculators. It doesn't tell the model that inventing records is forbidden, and the model plainly didn't read it that way.

That second point generalises beyond this agent, which is why it's worth the guide entry: a prohibition phrased about correctness doesn't cover invention. Any generative step that must produce list-shaped output while
its real source is unreachable has this failure available to it — and the fabricated version is more convincing than a broken one, because it's plausible and deterministic across page loads.

The case file carries four ranked fix candidates (carry the data source through adoption; rewrite rule 9 to bind records; add a fabrication grep-gate beside the existing completeness check; make "audited-content site" a
machine-readable flag rather than documentation an agent never reads), plus the grep tells and a verification method that checks the rendered artefact rather than the work item.

One thing I noticed while filing: bug 016 is currently used twice by two unrelated cases from concurrent threads, so the numbering has already collided once. I took 020 (019 was the highest) and noted the collision in
the file.

The handoff now points at 020 with its root cause, so the next thread starts from the diagnosis rather than re-deriving it. Still outstanding on the site side is the spec-level fix — the homepage spec still has that
search section with no data source, so the restore is protection for one page rather than a cure.

--

Re-read done — CLAUDE.md had gained a "standing four" working-docs directive since this morning, which my workstream was violating, and then I got on with the build.

Docs brought into line first. The notes file is renamed NOTES_vetcomparison.md and reordered newest-at-the-bottom — I'd deliberately reordered it newest-first this morning, so that was doubly wrong. The directive also asks for
corrections to be made visibly rather than silently edited away, so the file now carries two marked corrections: my wrong "the rebuild dropped the search UI" claim, and the earlier "all 803 rows have source URLs" claim that was
a false pass from testing NULL instead of empty.

The real build work: the fabrication was still in the database. My restore yesterday fixed the published file, not the source. Querying the components showed the generator sitting in page_components.rendered_html, deployed and
unlocked — so the next render would have republished it. It also wasn't where I'd assumed: the whole recreated tool lives in the hero slot (18KB of it), not filtered-result-grid. The false claims were spread across three
components.

I fixed it at source while keeping the chassis's better UI — its component has a region filter and pagination that our hand-written one lacked. Only the data layer was rewritten: same markup, IDs and styling, but fetching the
real /data/vet-full-index.json, with an honest disclaimer and a comment forbidding record generation that points at bug 020. Then the copy: price-sort controls gone, the "pricing information / ownership data" claim rewritten to
what we actually publish, and the "Ownership and Group Information" card — a feature we don't have — replaced with the claim-your-listing route, which is a product we do have. All four corrected components are now permanently
locked; before today exactly one component in the entire fleet had such a lock.

The sweep of other pages was worth doing, and worth doing by reading rather than by regex. Three more hits. Only one was genuinely false — the about page claimed "the directory identifies independently owned practices
separately from those owned by corporate groups", which we stripped on 15 July. The other two describe the CMA's findings and obligations accurately; a blanket find-and-replace would have corrupted correct content.

What I could not prove, and won't pretend otherwise: no render has been run against the fixed source. I tried twice to dispatch one manually and neither topic produced an orchestration state — rerender-pages is marked
experimental. So the database and the live site are both verified correct, but the render path is unverified. That's written into the handoff as the one thing to watch, with a one-line curl check for whoever sees the first
natural render.

One self-inflicted error worth flagging because it nearly did damage: I tried to load the corrected component using \set html \cat file`through a piped psql, and thecatran *inside the database pod*, which has no such file — 
  silently blanking the hero to zero characters while reportingUPDATE 1`. Caught it immediately by verifying rather than trusting the status, which is the same lesson this whole day has been about. It's recorded in the notes with
the correct method.

✻ Worked for 16m 49s

※ recap: We're rebuilding vetcomparison.uk as an honest vet price comparison site; the live site is clean and the fabricated data is now removed from both the page and the database, with components locked. Next: watch the first
natural render to confirm the fix survives it.


---
--

Docs and handoff refreshed so a new chat can start cold. The handoff is now HANDOFF_2026-07-19_vetcomparison_uk.md, rewritten with every figure re-checked against the live system rather than carried forward from
yesterday's file — I re-fetched the site, re-queried the database and re-read the CMA case page before writing a word of it. That grounding changed three things I would otherwise have written from memory.

The render path is still unexercised. Nothing has touched the site since Friday's restore, so the fabrication fix is verified in the database and on the live page, but no render has ever joined the two. When one finally
runs the homepage will visibly change — the live page is our simpler hand-written version, while the database now holds the richer chassis component with a region filter and pagination reading the real data. That change is
expected and is an improvement, but it is also the untested moment, so the handoff makes watching it the first job with the exact checks to run.

The substantive CMA Order still is not published. The case page's latest entry is still 30 June, so it is now overdue against its own July timetable with a hard 23 September deadline. The funding consultation closes in
eleven days and the draft response is still sitting waiting on you.

I also found the RCVS research you have been adding here — about 150KB across three documents on their institutional efficacy, financial stewardship and software delivery record. That is strategically live and is not folded
into any plan yet: the RCVS both builds the official comparison tool that will be our main competitor and writes the approval criteria for third parties, so how well they deliver bears directly on the argument we are making
to the CMA. It deserves a session of its own.

One correction to something I told you earlier today. I described this file as stale and told the next session to disregard it, because its opening still talks about the strip needing a push and quotes a practice count from
the 15th. That was a misreading — this is a running history, so of course its early entries are old; that is what a history is for. CLAUDE.md has since made it explicit that this is your document, appended to and never
rewritten, and that I should be adding to it whenever I write a substantial reply in chat. I have not been doing that, which is why the last few days of it are your paste rather than my writing. I have corrected the handoff
so it describes this file properly instead of dismissing it, and this entry is me starting to keep my side of it.

※ recap: the site is clean and working, the docs are current and a new chat can pick up from the handoff. Two things are waiting on you — the consultation response before the 30th, and seven build items in the admin queue.
The one thing to watch is the first automatic render, to confirm the fabrication fix survives it.

---
--

I did not wait to watch the render. The handoff said the first job was to watch the next automatic
render and check the fabrication fix survived it, and that the homepage would change for the better
when it happened. Instead of waiting, I put together what that render *would* produce and read it.
It was as well I did, because two of the five sections on the new homepage were broken, and one of
them would have been embarrassing.

The first was a second search box. The real one works — you type a name, town or postcode, filter by
region, page through the results, and it reads the genuine list of 2,109 practices. Directly beneath
it the render would have placed another search box over an empty grid that said "No results found."
and would have said that forever, because nothing was ever wired to fill it. Two search boxes, one
of them permanently empty. I have removed it, after taking a copy of everything first. The working
one already does the entire job, so nothing was lost.

The second was the news section: a confident headline about the CMA's remedies order with absolutely
nothing underneath it. It was trying to load news from a file that does not exist on the site.

I want to be clear that neither of these was invented data. Both of them refused to make anything up
and showed an empty state instead, which is exactly what this site is supposed to do and the whole
reason it exists. They were simply dead — and dead is still not good enough to publish.

You asked me to build the news feed properly rather than delete it, and that turned out to be the
more interesting job, because the thing blocking it was hiding in plain sight. Item seven on the
handoff's own list was described as optional and cosmetic — a classifier setting, if a news feed is
ever wanted. It was not optional. It was the gate. The system that goes out and fetches news only
looks at sites whose settings say a news feed is recommended, and this site's settings had no such
entry at all. So a news source could have been added and would simply have sat there doing nothing.
That is now fixed.

Then a decision I want to flag, because it is the kind of thing that could quietly undo the last
month of work. When you tell this platform a site should have news, it can create the sources
automatically, and one of the kinds it creates is an AI that *writes* the news itself from a prompt.
On any other site that is a reasonable convenience. On this one it is precisely the thing we spent
weeks removing. So I have restricted this site to real feeds only, and written the reason into the
setting itself, so that a future session that comes along and thinks it is being helpful by adding
the AI source has to read why it was excluded before it can do it.

The feed I have given it is the CMA's own announcements, filtered to veterinary matters. That
matters for two reasons beyond the site. The unfiltered CMA feed is all mergers and parking charges
and dentistry, with no vet content at all, so the filter is doing real work. And the very first item
in it today is the funding consultation that closes on the 30th — which means the site will now
carry it, and it means we get told the moment the substantive draft Order finally appears, instead
of you or me remembering to go and look at the case page. That has been sitting overdue since the
end of June.

Two other things worth knowing. I found a mistake in our own handoff: the command it gave for
checking the site was healthy was looking in the wrong file, and would have reported a perfectly
clean site as broken. Anyone following it faithfully would have raised a false alarm. That is
corrected. And I have sent the underlying platform question — how a section with no data in it can
still pass the check that is meant to catch empty sections — off to be diagnosed properly rather
than assert the answer myself, since that one affects every site, not just this one.

The news should start appearing on its own at the next six-hourly sweep. I will confirm it actually
arrived rather than trust that it should have.

※ recap: the homepage would have published a dead second search box and an empty news headline; the
first is removed and the second is now a real CMA veterinary feed, restricted to genuine sources so
nothing on this site is ever AI-written. The consultation still closes on the 30th and still needs
you.

---
--

The news feed is half working, and the half that is not working turned out to be a three-day
window. Two real CMA items came in overnight — the funding consultation that closes on the 30th,
and the case page for the market investigation — both with proper links back to gov.uk and no
invented text anywhere. So the part I built is doing what it should.

But nothing reached the site, and the reason is almost silly. The system only shows news
published in the last three days. Both our items were published on 30 June, so they are about
nineteen days old, and the renderer quietly decided there was nothing to show. When it finds
nothing it skips the publishing step entirely, which is why the file the page asks for was never
created at all. Three days is a sensible window for a site about gas prices or watch releases,
where something happens daily. It is quite wrong for a site about a regulator that publishes a
few times a month. I have widened it to thirty days, which is the same age at which the system
already throws items away, so the two now agree with each other instead of contradicting.

I had to make that change in two places, and I only found that out by getting it wrong. I changed
the live setting yesterday evening, and the run an hour later still used the old value, because
the setting also lives in a setup file that quietly reapplies itself. So the fix had not taken at
all. It is in both places now.

I also owe you a correction on something I said earlier. I told you the news system looked broken
across every site, not just ours, and I sent that off to be diagnosed. That was wrong. It was
running late, not failing — it caught up about an hour later and has run normally since. What I
actually had was ten minutes of nothing happening on a system that queues work, and I read
absence as evidence. The likely explanation is mundane and already written down in our own notes:
when a new build is deployed, jobs started in the first few minutes get silently dropped, and a
new build had just gone out. I have withdrawn that report with an explanation attached so nobody
wastes time on it.

Now the thing that actually matters more than any of the above. **The directory exporter is
failing.** That is the job that refreshes the list of 2,109 practices — the core of the site. It
ran on Sunday evening for the first time in two days and stopped with an error saying it had no
domain to work with, even though the domain is plainly sitting there in its configuration. It
last worked on the 17th, and several new builds have gone out since, so this looks like something
that broke rather than something misconfigured. Nothing is visibly wrong on the site: the existing
file is still being served and still has all 2,109 practices in it. But it has stopped being
updated, and the same fault appears likely to affect the equivalent job on the vet-medicines side
too. I have sent that for a proper diagnosis rather than guess at it.

※ recap: the news feed is ingesting real CMA items correctly; they were invisible because of a
three-day window that I have widened to thirty and fixed in both places it lives. I was wrong
that the news system was broken fleet-wide — it was just late, and I have withdrawn that. The
more serious thing is that the directory exporter has been failing since Sunday, so the practice
list has quietly stopped refreshing even though the site still looks fine.

---
--

Good news to start with: the news feed is now actually working, and the homepage has finally
rebuilt itself into the proper version. If you look at vetcomparison.uk now you'll see the richer
page with the working practice search, and underneath it three real CMA news items — the funding
consultation, the market investigation, and a new impact-assessment publication — each linking
back to gov.uk. And importantly, that news is now written into the page itself, so a search engine
that doesn't run JavaScript can read it. That was the whole point of the exercise.

The thing that unblocked it was the bug I found yesterday — the one where a setting you change in
the database silently does nothing. That fix went out in the new build, so the thirty-day news
window is finally real, which is why items from a few weeks ago now show up instead of being
thrown away. I've marked that bug closed, because it's now both fixed and live and I could prove it
by watching the news actually appear.

Two honest caveats on the render, because it dragged a couple of things back in when it rebuilt the
page. It restored the dead second search box I'd removed on Saturday, and it wiped the protective
locks we'd put on the components after the fabrication incident. I've removed the dead box again,
more thoroughly this time so it should stay gone. The locks being wiped is the more interesting
finding: it turns out those locks don't survive a full page rebuild at all — the rebuild throws the
old components away and makes fresh ones, and the lock goes with the old one. The reassuring part is
that when it rebuilt, it produced clean content with no invented data, so nothing bad actually
happened this time. But it means we can't rely on those locks as protection, and I've written that
up against the fabrication bug so nobody trusts them again.

Now the thing that still needs attention. The job that refreshes the list of 2,109 practices is
still failing, and has been since Sunday. The site doesn't look broken — it's still serving the
last good copy — but the list has quietly stopped updating. It'll try again tonight around 8:25. I
should be straight with you: the fix I made yesterday does NOT cover this one. Mine handled numbers
not reaching the code; this is a text value (the website's own name) not reaching the code, which
is a close cousin I deliberately left alone because fixing it carelessly could hide real bugs
elsewhere. So this one probably needs someone to sit down with that specific bit of code rather than
wait for the automated diagnosis — which, frankly, hasn't been delivering. Every diagnosis I filed
this session got stuck partway and never produced an answer; I've written that up as its own
problem.

I also owe you some corrections on my own work, which I've filed properly this time. I told you
yesterday that a couple of my review submissions had been "dropped" and I resubmitted them. That was
wrong twice over: they weren't dropped, they were just slow — about an hour, which the guidance now
says is normal — and they were invalid anyway because I'd formatted them in a way the reviewer
rejects. And earlier I declared the whole news system broken across every site when it was simply
running late. Both of those, and one where I reported something as filed when it hadn't been, are
now written into our shared log of wrong calls, because the pattern in all three is the same: I read
"nothing here yet" as "it failed", on systems that queue.

I've written a fresh handoff dated today so a new chat can pick up cleanly. The short version: the
news feed works, the big config bug is fixed and shipped, and the one live problem left is the
practice-list exporter, which needs a human look rather than the automated loop.

※ recap: news feed live and server-rendered; the config bug that blocked it is fixed and shipped and
closed; the dead search box is removed again and the protective locks turned out not to survive a
rebuild (no harm done this time). Still failing: the directory exporter that refreshes the practice
list — next attempt tonight ~20:25, and it needs a person, not the loop. New handoff written for the
next thread.

---

Today the medicine price comparison came back to life — the data side of it, not yet the
pages. You asked for the medicine scraping to feed vetcomparison again, and for us to read
the latest notes first because things had changed. That reading mattered twice over. First,
it corrected a mistake I'd written down yesterday: I had claimed the medicine price tables
didn't exist any more. They do — I had simply looked in the wrong part of the database, like
checking one filing cabinet and declaring the archive empty. That correction is now written
where the mistake was. Second, it surfaced the real distinction: the retailer prices we
collected in the spring were genuinely scraped, with the receipts kept — but the "typical
vet price" figure we used to show beside them was the invented one from the fabrication
episode.

So before switching anything on, two protections went into the code. The invented figure
family is now stripped out of everything we publish — it cannot reappear unless someone
deliberately builds a sourced version. And the exporter now refuses to publish any price
that lacks its source link and capture date; anything withheld is counted, and the count is
printed in the published metadata, so silence is impossible. There's a test that proves the
refusing branch actually refuses.

We also found and defused a trap: the setup file for the medicine exporter still carried the
old wrong web address (the .co.uk one we don't own), in a place where re-running that file
would have quietly put it back into the live system. It's blanked now, in both places. And
the reason medicine prices never refreshed on a schedule turned out to be simple: the
scraping task's database row had never actually been created — the setup script that was
supposed to point it at the right agent was updating a row that didn't exist, succeeding at
doing nothing. That row now exists.

The pipeline is coming up one step at a time, checking each before the next: the discovery
step ran and found two new product pages; the price scraping is running as I write this; the
export to the live site comes last, pointed at the real vetcomparison.uk. The prices land as
data files only for now — putting pages back in front of them is a separate decision, and
given what happened last time a tool was rebuilt automatically, one to take deliberately.

※ recap: medicine data pipeline switched back on step by step with provenance enforced in
code (no source link and date, no publication — withheld prices are counted publicly); the
invented "typical vet price" is stripped and cannot silently return; two dormant traps
removed; yesterday's wrong claim about missing tables corrected where it was made.

---

**2026-07-24 — the broken-looking "second search box" is finally gone for good (I hope).**

There was a stray, empty search box sitting on the homepage under the real one: a box with
a "sort by price" dropdown and a permanent "No results found." message beneath it. It never
worked and it was never meant to — the real search directly above it already does the whole
job. Two earlier sessions deleted it by hand and both times it came back a day or two later,
which is why the last handoff flagged it as the thing to watch.

I found out why it kept coming back. The site has a "plan" — a master list of what each page
is made of — and that plan still had this dead box in it. So every time the homepage rebuilt
itself (which it does automatically every few hours to pull in the latest CMA news), it
faithfully put the dead box back, exactly as the plan told it to. Deleting the box off the
page was like crossing a name off a printout while leaving it on the master list: the next
reprint brings it straight back.

So this time I removed it from the master list itself, not just the page. That's the part
that should make it stick. The website you see won't change this instant — it only updates
when it next rebuilds for the news refresh, roughly every six hours — but at that point the
dead box should be gone and stay gone. Nothing else on the page is affected; the real search,
the news, and the practice directory are untouched. I've kept a backup of the old settings in
case, and left a one-line check for the next person to confirm it's actually gone once the
site rebuilds this evening.

One smaller thing worth flagging: the platform has a "hide this section" button in the admin
screen, but from reading the code it looks like it only hides a section from the internal
health-checks, not from the actual published page — so it wouldn't have fixed this. I've noted
that as something to look at properly another time; I haven't touched it.

**2026-07-24 (correction to the note just above).** I said the homepage rebuilds "roughly every
six hours" — that isn't quite right and I watched the 7:45pm rebuild cycle do nothing. The
six-hourly news check only rebuilds the homepage when there's actually new or changed news; when
there isn't, it does nothing. In practice the homepage seems to rebuild about once a day (it did
so yesterday evening and again this afternoon). So the stray empty box should disappear on its
own within about a day, not within six hours. Nothing to worry about — the underlying change is
done and permanent; this is only about when the public page catches up. If it's still showing a
day from now, it can be nudged, just carefully.

**2026-07-24 — the big CMA document we were waiting for has landed (owner action needed).**

I checked the CMA's case page. The substantive draft Order — the main remedies document we'd
been watching for, described in the last handoff as "overdue and imminent" — was **published on
21 July 2026**, which was the same day the handoff was written, so it must have gone up just after
that day's check. It's now open for consultation and **that consultation closes at 23:59 on 20
August 2026** (about four weeks away). This is the one the briefing note (§3) was prepared for:
our line is pro-independent — an express right to reuse the mandated price lists, machine-readable
lists, and no selective blocking. It's drafted in outline but nothing gets submitted without your
sign-off.

Separately, and sooner: the **funding consultation closes at 23:59 on 30 July 2026** — about six
days away. I confirmed that date directly on the CMA case page (the last handoff had asked for it
to be re-checked against the notice). The draft response for that one is already written and
waiting for you to verify the levy figures and submit through the CMA portal.

So there are now two live CMA deadlines: the funding one in ~6 days, and the bigger substantive
one in ~4 weeks. Both need you; I've not submitted anything.

---

**24 July 2026 (evening) — the medicine price scraper was inventing prices, and two invented
ones reached the live site. Found, cleaned up, and fixed — fix awaiting the next deploy.**

The bug filed yesterday (two Advocate prices on the site that don't match the shop's actual
page) is now fully explained, and the explanation is uncomfortable but tidy: when our normal
price-reading code finds nothing on a page, we fall back to asking a small local AI model to
read the page instead. That fallback was being shown a snippet of the page that contained no
prices at all — and instead of saying "no prices here", the model made some up. Convincing ones.
In one case it literally copied the example price we'd used to show it the answer format. And
because we filed those made-up prices under the same label as properly-read ones, nothing
downstream could tell them apart.

The damage, once I swept everything against the saved page evidence: 212 invented price entries
in total — 8 from this week (two of which were on the public site) and 204 from back in April,
when the pipeline first ran. All 212 are now removed from the live data (kept in a quarantine
table, nothing destroyed), and the public site cleans itself at the next scheduled data export.
Important nuance: 19 entries at the suspicious £17.48 turned out to be *genuine* — so the sweep
checks each price against its own saved evidence, never just the number.

The fix: prices can now only be stored if they actually appear in the page text we save as
evidence — anything else is skipped and logged loudly. AI-read prices also get their own label
so they're always auditable. That code is written, tested, and committed, but it doesn't take
effect until the next software deploy; until then the scraper runs every 6 hours, so I've left
a re-check query in the runbook and the sweep should be re-run until the deploy lands. The
change has also gone to the review council (verdict pending at time of writing).

---

**2026-07-25 — the empty search box is gone for good, but I found something worse on the same page**

The duplicate, permanently-empty search box that kept coming back on the homepage is now
genuinely gone. It has survived two full page rebuilds since I fixed it, which is the proof I
was waiting for — the two earlier attempts to remove it were undone by the very next rebuild,
because they deleted the box without removing it from the site's build instructions. This time I
removed it from the instructions, so there is nothing left to bring it back. The homepage is now
about 4KB lighter and shows only the four sections it should.

While checking that, I found a bigger problem on the same page. The homepage's main block of six
cards — "Search practices by location", "Compare prices before you go", "See who owns your
practice", and so on — **has six links and every single one is broken.** Anyone clicking any of
them gets a "page not found". This is the most prominent thing on the page after the search
itself, and it has been dead the whole time.

Five of the six point at pages that were never built: a search page, a pricing explainer, an
ownership explainer, a pet-owner-rights guide, and a "claim your listing" page. The sixth points
at a guide that **does** exist — the CMA compliance guide — but at slightly the wrong address, so
it 404s too. That last one I have fixed, since there was only one sensible answer. The other five
I have deliberately left alone, because deciding where they should go is a decision about what
the site offers, not a typo — and three of those missing pages are already sitting in your review
queue waiting for exactly that call.

Worth knowing why nobody spotted this: we have an automatic checker that finds broken links, and
it works — it has flagged 188 of them across six other sites. It has never produced a single
finding for vetcomparison. I can't prove why (the records that would show it get deleted after 24
hours), but the checker is reliable enough that if it had run on this page it would have found
these. So it looks like this site is simply not being checked. I have written all of this up
against the platform bug that owns broken links, where another session is actively working; I
have not started a competing fix.

**On the CMA front, the news feed is now doing its job by itself.** The draft Order published on
21 July has appeared on the homepage automatically, with a link to the real gov.uk consultation.
I checked the wording our system generated around it against the actual legal documents, and it
is accurate: the Order really does require published price lists (Article 7), really does cap
prescription fees (Article 18), and really does require practices to disclose who owns them
(Article 5). I had initially thought two of those three were unsupported — that was me reading
the summary page instead of the documents themselves, and I was wrong.

**One thing you should know before we build the compliance deadline calculator.** Every number
and every date in the draft Order is written in square brackets — `[£21]`, `[£12.50]`,
`[X March 2027]`. That is the CMA's way of saying "not settled yet"; they are placeholders until
the Order is actually made. What *is* fixed is the relative timing: 3, 6, 9 or 12 months from the
Order coming into force, with smaller practices getting longer. So a deadline calculator can
honestly say "six months after the Order is made" but cannot honestly say "by 31 March 2027" —
and publishing a specific date as fact is the exact thing this site was cleaned up for. I have
checked, and none of those provisional figures appear anywhere on the live site today. Worth
deciding on that basis rather than cancelling the tool outright — it can be built, just not to
output calendar dates yet.

The two consultation deadlines still need you: the funding one closes **30 July at 23:59** (five
days), and the substantive one **20 August at 23:59**.

---

**26 July — the invented-prices bug is closed, and it very nearly closed on a false premise.**

You'll remember the medicine price scraper was caught making prices up: when its normal parser
found nothing on a page, it handed the text to a small local AI model and asked it to find
prices, and the model — shown a page with no prices on it — invented a plausible-looking table
instead of saying "none". Two of those invented prices were published on the site. We tracked
down 212 of them in total, took them out, and wrote a guard: before any price is saved, the
system now checks that the price actually appears in the page text it keeps as proof. If it
doesn't, the price is thrown away and the refusal is logged.

That guard was written two days ago but was sitting in code, not yet running. Today I found it
had in fact shipped in the meantime and has been running on every scrape since. All the checks
were green: every one of the 2,577 prices we have ever stored can be found in its own evidence,
the published data file has been rebuilt without the two bad prices, and — this is the bit that
misled me — the AI fallback had been asked to extract prices sixteen times since the fix and had
returned "nothing found" every single time. I concluded from that the model had stopped making
things up, so the guard was really just a belt-and-braces measure, and there was nothing left to
test.

**I tested it anyway, and the model made prices up on the first try.** I forced the scraper to
revisit the exact page that caused the original problem. It found no prices by normal means,
handed the text to the AI, and the AI invented three products — £19.25, £34.99 and £68.75,
including a pack size that does not exist on that page. The guard caught all three and stored
nothing. None of those figures appear anywhere in the page we captured, while the page's two
real prices do — so the guard threw out exactly the right things.

So the honest position is the opposite of what I was about to write down: **the guard is not a
backstop, it is the thing doing the work.** The sixteen empty answers were other pages, not a
cured model. It is also worth knowing the model invents *different* numbers each time — it
didn't repeat July's fabrications — which is why we never delete data by matching a known bad
price, only by checking the evidence.

Two smaller things for the record. The AI call took eight and a quarter minutes, which is slow
enough that one awkward page eats most of a scraping run; no wrong data comes out of it, but it
caps how much we can scrape and is worth measuring before anyone turns the batch size up. And
the review council did approve this fix, but its verdict arrived after the commit was written,
so the commit can't carry the usual "reviewed" marker and our coverage report will always show
it as unreviewed. It was reviewed; the note is in the case file.

The case has moved to the closed pile.

---

**2026-07-26 — I found why the site is thin, and it is not what I expected**

You said the site isn't strong enough, and dropped the CMA consultation work. Agreed on both, and
I've stopped all consultation work.

My first instinct was that the site is thin because pages are missing — six dead cards, three
pages never built. I now think that's the wrong read, and acting on it would have made the site
*look* finished without making it any more useful.

Here is what's actually going on. Almost everything genuinely useful that we hold, we are not
allowed to publish under our own rule that nothing goes on the site without a source. The prices:
762 of them, not one carries a link to where it came from, and that's unrecoverable because they
pre-date the rule. The ownership information: we have a group name for about 1,100 practices, but
it has no evidence behind it and it **contradicts itself on 870 of them** — every one of the 55
Medivet practices, 27 CVS practices and 34 Vets4Pets practices is currently flagged as
"independent". Publishing that would tell pet owners the exact opposite of the truth, on the one
subject this site exists to clarify. The practice details we hold — species treated, emergency
cover, accreditations — are real but undated and unsourced.

None of that is live, I should say plainly. The site publishes six fields per practice and none of
the problem data. So nothing false is on the site. We are under-publishing, not mis-publishing.

**The reason is that the data pipeline has been switched off since March.** All three collection
jobs are disabled. Nothing has been verified since 18 March — zero practices in the last thirty
days. They were turned off during the July clean-up, quite rightly, and everything we've done
since has been about *publishing* — the exporter, the guides, the news feed. Nobody ever turned
collection back on. So the site has been serving a four-month-old snapshot, and the reason it
feels thin is that it is.

The good news is that the machinery is all there and I don't need to build much. The verifier
already picks up each practice's Companies House registration number straight from its own website
footer using a plain pattern match — no AI guessing involved, which matters a lot here. That
number is the key to doing ownership properly, sourced to a company number anyone can check.

Two false starts worth recording, both caught cheaply. I first planned to switch on a scraper I
thought had never been installed — reading the file to the end showed it installs itself and then
deletes itself, with a note saying "forget it, we'll use the regex in the verifier". Then I
assumed that replacement was never written. It was; it's sitting there working, just never run.
Both times I concluded a part was missing when it was simply idle.

**One thing I need to measure before doing anything else.** When the verifier runs, it's supposed
to record *where* each fact came from. But all 2,970 records we have show that field empty. So I
don't yet know whether switching collection back on would give us publishable, sourced data — or
just fresher unpublishable data. That's the difference between "turn it on" and "fix it first,
which needs a code change and a deploy". I'm going to run a small pilot over about ten practices
to find out, rather than restart the whole thing and hope. It's one query's worth of work to
re-enable and I deliberately haven't.

---

**2026-07-26, later that evening.** I picked up where the last note left off, with the one thing
that needed measuring before anything else: does the verifier record *where* each fact came from?

**It doesn't, and it can't — and I could show that without running anything at real practices.**
The verifier asks the AI for six things: the business details, practice details, staff, prices, a
confidence score and some notes. It never asks it where the information came from. But the part of
the code that saves the results goes looking for a "source URL" *in the AI's answer* — a field
nobody ever asked the AI to produce. So it's always blank. Meanwhile the part of the system that
actually fetched the page, and therefore genuinely knows the address, isn't connected to the part
that saves the record. The knowledge exists; the wiring doesn't.

I checked that against all 2,970 records we hold and not one of them contains that field at all —
not empty, simply absent. So this isn't a setting that got missed or something that broke; the
information was never on its way in the first place.

**Worth saying plainly what the fix is not.** The obvious-looking repair is to add "and tell us
where you found it" to the AI's instructions. That would be a bad mistake on this site of all
sites. It would mean the AI *telling us* what its evidence was — a claim about evidence rather
than evidence — and that is exactly the thing we spent July cleaning up. The right fix is to hand
the saving code the address the fetcher already used. That is a genuine code change, so it needs
review, a build and a deploy before any large crawl is worth doing.

**On the crawl you approved: I have the hit rate you asked for, and I got it without touching
anybody's records.** Running the real verifier over 25 practices would have written a batch of
AI-extracted, unsourced facts over their current entries just to learn a percentage. Instead I
wrote a small probe using the *exact same* pattern-matching code the live system uses, pointed it
at 25 practices chosen at random, and had it only read.

**About one in six practices publishes its company number on its home page** — 4 of 25. **Widening
the search to terms and privacy pages takes it to about one in four** — 7 of 25. That second
number matters because it costs nothing: it's a settings change, live the moment it's made, no
deployment. Right now the verifier looks at pages called "fees", "prices", "about", "team",
"contact" and "services" — and not one of those is the legal page where a British company normally
prints its registration number. Three of my seven hits were on exactly those missed pages.

Please treat both figures as "roughly a sixth" and "roughly a quarter" rather than precise: 25 is
a small sample and the true numbers could reasonably be a fair bit either side. Now the probe
exists, running it over a few hundred is cheap if you want firmer ground.

**The quality of what we find is excellent, which is the encouraging part.** Six of the seven
numbers matched a real veterinary company at Companies House. And two of them immediately told us
something we currently have wrong: Heywood Veterinary Centre is owned by VetPartners, and Animed
in Whitstable by CVS — both sourced to a company number anyone can look up. That is precisely the
evidence needed to fix the ownership mess where 870 practices are flagged "independent" while also
carrying a group name. So when the pattern match does fire, it is worth a great deal; the question
is only how often it fires.

**One thing I noticed by accident.** One of my 25 "practices" isn't a practice at all — it's a
listing on the Lancashire Telegraph's business directory. That's the same problem as the 176
entries pointing at a site called wheree.com. I haven't counted how many of our 3,419 entries are
directory pages rather than real practices, but it's worth counting before we start building a
page for each one.

**What I'd suggest next**, in order: make the settings change to look at legal pages (free,
immediate); get the code change reviewed that records where each fact came from; then crawl at
scale, once a crawl can produce something we're allowed to publish. I've put the provenance
finding through the diagnosis review before treating it as settled, rather than after.

**Correction, an hour later the same evening — and it matters more than the thing it corrects.**

I told you above that getting from one-in-six to one-in-four was a free settings change. **That
was wrong, and I should have checked before saying it.** I looked at the settings on that step and
read them as instructions; they aren't. The setting that says "also look at the fees, prices,
about, team, contact and services pages" **is not read by the software at all**. Nor is the one
saying "fetch up to three pages", nor the one saying "if the practice has no website, use the top
search result". The code that does the fetching looks at exactly one address and asks for one
page, and quietly ignores the rest of the settings.

So two things change. First, the one-in-six figure is *exactly* right rather than cautious — the
home page is genuinely all the system ever looks at. Second, getting to one-in-four is a code
change, not a settings tweak, so it should be done in the same batch as the "record where each
fact came from" fix: one review, one build, one deployment, rather than pretending there's a quick
win to grab first.

**The wider point is the worrying one.** That step is written to look like a six-page crawl with a
sensible fallback. It is a single page fetch. Nothing anywhere flags the difference — unused
settings are simply ignored in silence, so the configuration reads like a description of what
happens without being one. I've only checked these four settings on this one step, so I don't know
how far that goes, but it is the kind of thing that makes people confident about the wrong thing.
I've recorded it in our log of wrong calls, with the one-line check that would have caught it.

**Later still — the bigger sample, and one thing I got ahead of myself on.**

I ran the same read-only check over 100 practices instead of 25. **About one practice in five
publishes its company number on its home page — 22 of 100. Looking at terms and privacy pages as
well takes it to 30 in 100, roughly three in ten.** Please treat those as "about a fifth" and
"about three in ten"; even at 100 the true figures could reasonably sit a few points either side.
My earlier one-in-six was from too small a sample, which is why I flagged it at the time.

**The quality is what makes this worth doing.** Sixteen of the twenty distinct companies we found
match a real veterinary company at Companies House. And here is the striking part: those 30 hits
are only 20 different companies, because **eight of the hundred practices turned out to be owned by
the same company, VetPartners.** Three more are CVS. So thirteen of the thirty are three
corporate groups. That is precisely the picture our records currently get wrong — we have 870
practices flagged "independent" that also carry a group name — and now it is evidence rather than
a label, because anyone can look the number up.

**Where I got ahead of myself.** In my correction above I said the one-in-six figure was now
*exact*. I should not have said that. Besides the sample being small, there is a part of the
pipeline I hadn't read: our system doesn't fetch these pages the way my test did — it goes through
a third-party service, and the setting controlling whether that service keeps or discards the
footer of a page is, because of how the code is written, never actually sent. So the service falls
back to its own preference, and I don't know what that is. Since company numbers live in footers,
that could mean the real figure is lower than mine.

I tried to settle it from 2,450 pages we've already fetched that way. The evidence is genuinely
mixed — three quarters of them kept their footer menus, which suggests footers survive, but none
contained company-registration wording, which could equally mean the footer was stripped or simply
that shop product pages don't print one. **So I'm leaving it open rather than guessing**, and I've
written it at the top of the relevant bug as the first thing to check, because it decides whether
the fix we're planning would actually work. One real run answers it.

Worth naming the pattern: I'd just been caught over-claiming in one direction, and immediately
over-claimed in the other. Being corrected about one thing tells you nothing about the next thing.

## 2026-08-24 — the "Claim your practice listing" page exists now

You reported that the page behind "Claim your practice listing" was missing — the homepage card
and the independent-practices guide both pointed at a page that showed "not found". That page had
been sitting half-planned since mid-July: the system had a slot for it but nothing that knew how
to build it, because it was originally imagined as a per-practice profile page, and the machinery
for those was never built (that's the separate bug 206, which another thread is looking at).

What I did: rather than wait for the profile machinery, I re-aimed the page as what the links
actually promise — a page that explains what claiming your listing means and how to do it. I gave
the build system a clear brief first (what to cover, and hard rules: invent nothing, no prices, no
dates from the draft CMA Order, only link to pages that exist), then sent it through the same
build route that successfully built the directory page earlier this month.

It built first time and is live: https://vetcomparison.uk/entities/practice.html — it says
claiming is free, explains that we never publish a price we can't attribute to a source, and gives
the email route (vetcomparison@contactforsales.com). I read the whole page: no invented figures,
no fake claims, all links work. The two automatic "dead link" alarms about this page are closed.

Still open, unchanged: the per-practice profile pages (needs the bug-206 thread's machinery), and
the compliance deadline calculator page (held deliberately — it must never print a calendar date
until the CMA Order is actually made). One small thing to remember: this page, like the
independent-practices guide, talks about "the remedies Order" in the present tense; when the Order
is actually made, both pages deserve one pass to check the wording against the final text.

## 2026-08-24, afternoon — claim intake by email, the first real request, and the nurse with the odd eyes

You asked whether practices need a form to claim a listing. My answer was no, and you agreed:
the whole value of a claimed listing is that we've verified the person speaks for the practice,
and an email from the practice's own domain does that verification for us — a web form would just
collect unverified text and we'd have to email them anyway. You ruled that a practice-domain email
is good enough evidence, and that we ask for a company number as a second check. The claim page
now has a "Claim your listing by email" button at the top and bottom that opens a pre-filled
email asking for exactly what we need, plus the same list written out on the page in case the
button doesn't work in someone's email program.

The first real request arrived the same morning — Vet Home Certs, a mobile network of vets doing
animal health certificates, asking to be added. I checked them before doing anything: they're a
real registered company (Companies House SC786251) and genuinely not in our directory. One
caution I flagged: their email came through a relay service, so the "from their own domain" check
isn't fully proven until they reply — which they have to do anyway to send us their locations and
prices. Their listing record exists but cannot publish until that data arrives and is verified;
you sent the reply asking for it.

The image with the odd eyes: it was made by an older image model (the 2023-era one) on 17 July —
one day before the framework switched hero images to a much better model. So it was a leftover,
not a live fault. I regenerated it (a calm clinic waiting room, dog and cat, nobody's face in
shot — I looked at it myself before calling it done) and, because the site's stored design brief
still said "no imagery at all" from the July clean-up, I replaced that with a real direction: white
and teal, photographic, and no close-up generated faces ever. Along the way I found the old model
was still being used for some other sites' images through a gap in the routing; I filed that as
bug 382, another thread found the cause and fixed it the same day, and it went out in the
afternoon's release. That's closed.

Finally, the "CSS is broken" moment: nothing on the server changed between when you saw it broken
and when you saw it fixed — you caught the page mid-update while my edits were landing back to
back. Not bug 198 (that one is fixed and stayed fixed), nothing to change.

## 2026-09-03 — the green checkmarks are live, and the whole design thread behind them

You asked for the homepage to be better designed, then ruled on the three questions that came
out of it: use the green deliberately, per-section illustrations, build the deadline calculator
fully. Where that all landed:

The deadline calculator is live and in the site menu. It does exactly what we agreed and nothing
more: you pick your practice size, enter the date the CMA Order is made, and it computes every
obligation's deadline from that date — all seventeen obligations copied word-for-word from the
draft Order itself, no money figures, no dates of its own, and a plain statement that the Order
is still a draft. When the real Order is made it will need one review pass — that's by design.

The green now means something. Claimed listings in the directory carry a green checkmark, and
they sort to the front — so Vet Home Certs' 51 locations are the first thing a visitor sees,
each with its mark. That's the "make them prominent" promise and the "use the green" ruling
delivered as one thing: the colour is the badge of a practice that verified itself and published
its prices with us. The independence banner stays on the shelf as you ruled.

Getting there found two things worth knowing. First, a rendering subtlety nobody had written
down: refreshing a page after a template change does NOT refresh the data underneath it — a
page can look updated while showing week-old listings. It's now documented as a trap for every
other site, credited to the review seat that predicted it before it happened. Second, another
thread asked to build the site's evidence register (the machinery that watches published claims
against their sources); I handed them the full map of what the site asserts, including one thing
they should find and fix: our guides state the £21/£12.50 prescription caps as settled, and the
draft Order actually leaves those numbers open until it's made.

Still with you: the August header question (a hand-edited header was overwritten by an automatic
rebuild; the platform saved your version and is waiting on restore-or-drop). Still in the queue:
the illustrations, and three small text-contrast fixes inside the comparison tool.
