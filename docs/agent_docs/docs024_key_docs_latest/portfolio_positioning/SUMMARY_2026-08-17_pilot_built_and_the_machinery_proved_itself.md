# Summary — the pilot built itself, and the directory machinery worked first time

2026-08-17. The previous summary (`SUMMARY_2026-08-16_phase_b_complete.md`) ended by saying Phase B was "finished and unproven in the round" — everything built, nothing yet exercised by a real site. That sentence is no longer true, which is why this one exists.

## What we're trying to do

Build a fleet of finance sites where each one carries a directory of the relevant UK providers, with every fact cited to a named source and machine-checked against it. The directories are the thing a competitor cannot cheaply copy. Before committing to a hundred and forty domains, we build one site end to end and see what actually happens.

## Where we've come from

Yesterday all the machinery existed and had passed review, but no site had ever been through it. Five separate pieces sit between "we have verified data" and "a site publishes it", and a kind missing from any one of them fails in a way that looks like success everywhere else. So the pilot was designed around three specific questions, asked in order, each answerable by a query rather than an opinion.

## What we've done

We seeded the pilot site, wrote its brief, and dispatched it. It sat in the queue for an hour — not because anything was wrong with it, but because the whole fleet's build pipeline was stuck behind a single unrelated job belonging to another workstream. That cleared on its own, and then the site built itself unattended in about two hours: research, strategy, brief, page plan, design, and pages.

**All three questions came back right.**

The site was correctly marked as wanting a mortgage-lender directory when it was classified. The planner then read that mark and produced exactly what it should have: a dedicated lender directory page, named and typed precisely the way the rest of the machinery expects, built from a hero, the directory listing, and a call to action — plus a directory panel on the home page. And nothing was silently discarded along the way, which we checked directly rather than inferring from the page looking right.

That last point is worth dwelling on. The gap we closed on Friday was that a site could be marked as wanting a directory and the planner would ignore the mark, so the site only got its directory later, after being built and published once without it. On this build the planner got it right the first time, unprompted, from the rule alone.

**The cost, measured rather than estimated:** 43 model calls, roughly 390,000 tokens in and 121,000 out, and 11 images. **[CORRECTED 2026-08-18: this figure was ~70% low — the real total is 73 calls, 664,000 in, 185,000 out, about $3.81 per site today and $4.83 from 1 September when an introductory model rate ends. I measured while the build was still running, so jobs still in flight were invisible to the count. See NOTES, 2026-08-18.]** Nearly all of it is page writing. I have taken care not to inflate that figure by counting other workstreams' activity that happened to run in the same two hours — it is attributed to this site's own jobs — and it should be read as a floor rather than a total.

**What broke is the pilot's real yield**, and none of it is the directory work:

- Eleven jobs failed trying to deploy, all with the same git error. The obvious reading — "the site's repository was never created" — is wrong: most sites here have no repository and publish by a different route entirely. So the question is why the deployer took the git path at all.
- Two pages were blocked by twenty instances of raw template syntax leaking into the finished HTML. I checked specifically whether my own new content rules had caused this. They had not; they blocked nothing.
- A handful of component and re-render failures, and a normal queue of things flagged for a human to look at, which is expected on a brand-new site.

**One thing to be clear about, because it affects what to do next.** A fresh deployment went out this afternoon and the pods restarted, but the image tag did not change, so the running software is the same as before and contains none of this week's fixes — including the one I made this morning. I checked the binary rather than trusting the restart. Another workstream independently found the same thing today and measured it as 203 unshipped commits. A restart is not a release; the tag has to be bumped.

## Where we are now

The directory machinery is proven end to end on a real site. That was the open question and it is now closed.

The pilot site itself is built but not published: the deploy failures mean nothing is live yet, and there is a queue of items waiting for a human. So "the machinery works" and "the site is finished" are two different statements, and only the first is true today.

## Where we're going

Three things, in order. Get the deploy path understood, since eleven failures share one cause and it is not a directory problem. Work the human-review queue for this site so the pages can complete. Then get a release out with a bumped tag, which ships this week's fixes and is also what makes the random-directory-decision fix real.

After that: the cost figure above is the number to pace the fleet against, and Phase D's outstanding decisions are the owner's before any wave goes out.
