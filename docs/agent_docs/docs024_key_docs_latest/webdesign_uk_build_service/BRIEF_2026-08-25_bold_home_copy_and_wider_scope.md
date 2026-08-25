# BRIEF 2026-08-25 — bolder home-page copy, wider site-type scope (owner, verbatim)

First installment of the copy/design brief promised at the 08-25 pause
(`SUMMARY_2026-08-25_webdesign_uk_build_service.md`). Owner's words verbatim, then the
lane's readings and what was done. Applied by `SQL_2026-08-25_bold_audience_categories.sql`.

## The owner's words (chat, 2026-08-25, unedited)

> First copy changes: I think we can be more bold and honest on the home page, before
> they sign up about what they're getting and what they need to know and what they need
> to be in order to use the sites that come as a .csv and the sites aren't just business
> sites. Example copy:
> This is for experienced webdesigners because we're just giving you a .csv file with
> your static site after which it's yours to host, yours to edit and yours to maintain.
> We are not a hosting company so, instead, we can help you set it up on free hosting
> like netlify. You can see what we've built for 30 days at a link we'll give you.
> Why would you want a starter site like this? We think that editing an existing site to
> suit what we want is a whole lot easier than having to start from scratch. So these
> starter sites are exactly that. A more or less complete starter template for you to
> carry on with.
> How to edit the site? This is the trickier part of our offering - We can't really help
> you much here until we do start hosting. Here are several ways you can do it.
>
> Brief
> too limited scope in what sort of sites we offer, other types of sites are probably
> going to be common. sites like dartsonline or robot-hands - what category would they
> be called for instance? And more categories too.

## Readings the lane applied — each one flagged back to the owner in chat

1. **".csv" read as ".zip".** The attested and built deliverable is a ZIP
   (`delivery_live_link_and_zip`; DGH-011 zip-deliverer), and a static site cannot be a
   .csv. Nothing was attested as csv. **If the owner really means csv, say so and this
   reading reverses.**
2. **The register was already most of the way there.** `starter_site_initial_copy`,
   `yours_to_change`, `any_site_type`, `keep_it_online`, `no_presales_service` (which
   already says "bare-bones starter website") all pre-date this brief. The gap was the
   SERVED HOME PAGE, which still led with "A complete website for your business" and
   was never rebuilt after the 08-18/08-22 register rounds (it also still carries the
   retired "two or three days" figure). So the work splits: attest what is genuinely
   new (below), then rebuild the page.
3. **Genuinely new, attested by this brief:** the audience gate (experienced web
   designers / comfortable with files — bounds the BUYER, not the site's subject);
   "We are not a hosting company" in those words; the month fixed at **30 days**; a
   home-page "how to edit the site" answer; named CATEGORIES of site.
4. **"until we do start hosting" is deliberately NOT attested.** A future tease invites
   pre-sales questions nobody may answer (`no_presales_service`), same as the
   2026-08-21 "for now" precedent on TLDs. If hosting launches, the copy changes then.
5. **Example sites stay OFF the page.** The owner attested 2026-08-18 (inside
   `any_site_type`) that no example sites are named until sites exist through this
   route. Today's brief asks what CATEGORY dartsonline/robot-hands would be — answered
   with categories, which go on the page; the example names stay in the fact's source
   notes only.
6. **Category vocabulary,** grounded in the estate's own classifier
   (classification.category: interactive 13 · hub 5 · editorial 4 · brochure 2, counted
   2026-08-25) and its archetype labels ("financial calculator utility", "lighting
   education content hub", …). Customer-facing names chosen: **business and service
   sites · tools and calculators · guides, reviews and enthusiast sites · reference and
   comparison directories · portfolios, personal, community and project sites.**
   dartsonline.com = guides/reviews/enthusiast (editorial); robot-hands.com = reference
   directory with interactive tools (hub/interactive).
7. ~~**NOT ruled, left for the owner:** whether to state that online shops that take
   payment are out of scope.~~ **RULED 2026-08-25 (evening), all three flags in one
   message:** *"template can remain a banned word. .zip is better. I don't think we can
   do online shops yet so we can exclude that."* → template stays banned (writer_block
   already says "starter site", SQL_2026-08-25c); ZIP confirmed; online shops attested
   as a capability limit in `any_site_type` (`SQL_2026-08-25d`).

## A defect found while applying this (fixed in the same SQL)

`writer_block`'s "HOW THE SITE LEADS" paragraph instructed the writer to *"Name the
real sites built by this system, which fact `any_site_type_examples` attests"* — **a
fact that does not exist**, and the opposite of the no-examples rule two paragraphs
earlier. A ghost reference from before the 08-18 examples-deferred ruling; two
contradictory instructions in one prompt produce a confident wrong answer (§8 trap 3).
Rewritten to match the no-examples state.

## Still owed after the SQL (the page half)

Home-page rebuild so the served copy catches up (it must also pick up the 08-22
"three or four days" figure); the hero leads with the audience gate + starter-site
argument; a "how to edit the site" section; categories named where pages say what can
be asked for. ⚠ The rebuild removes the hand-placed "Not active yet" label
(`RUNBOOK_go_live_webdesign_uk.md` gate 2 — re-place it) and may hit the known
`validate_content` "1 blockers" failure (chat-on-home-page item; grep chassis logs at
`validate_page_content.go:412`).
