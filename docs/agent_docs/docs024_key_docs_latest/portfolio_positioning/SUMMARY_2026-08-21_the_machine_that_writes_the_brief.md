# SUMMARY — 2026-08-21: the machine that writes the brief

*Written to be read aloud. The previous read-out is `SUMMARY_2026-08-19_first_sites_live_and_the_wall_the_fleet_would_have_hit.md`, and it is worth keeping — two of its three "where we are now" statements have since been overtaken, which is what makes the series useful.*

## What we're trying to do

Turn a portfolio of around 1,500 owned domains into a fleet of genuinely different, framework-built sites — each positioned so it does not compete with its neighbours, each built by the pipeline rather than by hand, each carrying the compliance guards its subject needs. The measure of success is not that sites exist; it is that a stranger could tell them apart, and that nothing on them is asserted without a citation.

## Where we've come from

Two days ago the read-out was: the machinery is proved, the first sites are live, and the thing that would quietly degrade the next fifty is understood but not fixed. Three things were open and named. The domain estate had never been inventoried, so nobody knew what we actually owned. The brief writer — the piece that would let 1,500 domains be specified without 1,500 hours of writing — did not exist. And the regulated-identity guard was written but not running.

## What we've done

**The estate is inventoried.** The registry export you supplied covers 1,567 domains, all registered, none suspended. Roughly 1,250 are parked at marketplaces, 207 have never had nameservers set, 14 are the sites we have built, and **62 sit on real hosting** — of which 25 are serving actual content. Five domains are family names. We now know what we own, which we did not on Monday.

**You made the decisions that unblock the fleet.** Flow A, with the brief always present but written by an agent when you have not written one. The spec is aspirational and the plan is achievable — a distinction that turned out to be better than the one we proposed, and which we have built to. You read every brief and may add or change a few words. Third-party work goes on domains we buy for the customer, not on yours. And the positioning register moves into a database, which is now the source of truth.

**The brief writer is built, live, and proved twice.** Given nothing but a domain name it researches the subject and writes a founding document — the proposition, the audience, what a reader actually comes for, a content plan of a dozen or more specific items graded core, valuable or aspirational, concrete tools with real inputs and outputs, and the things the site must never claim. Then it stops and waits for you. It writes to exactly the same place a hand-written brief goes, so nothing downstream changed, and because it stores rather than overwrites, your edited version keeps the machine's original underneath it — the difference between what it proposed and what you changed is preserved for free.

The first brief, for a houseplant-pot domain, named its real competitors, took a position against them, proposed fifteen content items across eight kinds and three working tools, and — unprompted — marked its own uncertainty where the research had not supported a claim.

**And it reads the register.** The positioning document is now a database table of 189 rows. Given a registered domain, the brief writer is handed that domain's entry and its siblings before it writes a word. The second brief, for a buy-to-let calculator, put it plainly: *"The site deliberately does not serve the general homebuyer (M2 territory), does not compare rates (M3 territory), and does not go deep on company structures (M12 territory)."* It stayed off three neighbours' ground by name, and where a boundary was genuinely unclear it asked you rather than guessing.

**The regulated-identity guard is live.** Every site the platform builds is now checked for claiming to be an authorised firm about itself, and refused unless a record says otherwise. A client who proves authorisation gets a proper entry — firm, registration number, who checked it, what they saw — and may then say it, with the number becoming a fact the system can check their pages against.

**Sitemaps became a mechanism.** You ruled that all future sites should have them; only eight of twenty-five live sites did. A generator existed but nothing ran it. It is now an action the pipeline can call, on by default rather than opt-in.

## Where we are now

> **⚠ CORRECTED 2026-08-24: this paragraph is overtaken.** `bugs_open/311` is **closed** — fixed,
> live, and the originating page verified healed at the served artefact (`remortgagecalculator.uk`
> now serves a real calculator: 6 inputs, 69,704 bytes, where on 08-18 it was 40,726 bytes with
> none). The precondition is gone and that site is unlocked; `adversecreditmortgage.co.uk` remains
> locked on the owner's call rather than on anything technical. Left in place unedited because the
> summary series is a record of what we believed at each milestone, and a summary that was
> overtaken is evidence about how the understanding moved.

**Builds are still halted, and the reason has narrowed to one thing.** The decisions that were blocking are answered. What remains is a defect: two sites cannot both have a calculator with the same internal name, so on a fleet of finance domains that all want the same tools, whichever site is built first owns the name and every later one ships its tool hollow — with nothing on the page to show a reader anything failed. That is understood, written up, ruled to be fixed as one piece of work, and **not yet built**. It is the last thing standing between us and running the pipeline at scale.

**Twenty-two of your hosted sites are cleared for a remake**, three protected. They are all small — the biggest is 22 pages and nine are single-pagers — so there is very little to lose and a great deal to gain.

**The honest position on quality**: the brief writer has produced two briefs and both were good. Two is two. The regulated guard is installed but has not yet refused anything in production, because both our sites are locked — so we can say it is running, not that it works.

## Where we're going

**First, fix the component collision.** It needs a review round and a chassis roll, and it is a precondition rather than an improvement: building fifty sites before it lands means fifty sites to repair afterwards.

**Then start remaking, from the small end.** The single-pagers with strong generic names — advertise, conferences, catalogues, copyonline — have nothing to lose, no regulated angle, and each gives the brief writer a different shape of subject to work with. The insurance one waits until the compliance machinery has been seen to refuse something.

**Then the fleet.** With a brief writer that reads the register, a review step that holds until you have looked, and a positioning table that covers the estate, the thing that was a hundred hours of writing becomes a queue you work through. That was the point of this week.

**Still owed, and worth saying rather than leaving implicit:** 21 portfolio domains have no register row because the document never named them; the register document's own fate — kept as a view, or retired — is undecided and must not be left with two editable copies; and the card-sending mechanism you want on the Christmas domains would be the first thing this platform builds that takes a stranger's input and delivers it to a third party, which is a genuinely new kind of thing and needs its delivery half designed before its front end.
