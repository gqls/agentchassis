-- FILE: SQL_2026-08-14b_build_duration_next_day.sql
-- Owner, 2026-08-14: the build time becomes "usually next day", re-attesting a
-- figure carried over from the £1,200 offer since 08-04. Fact + writer_block
-- (the wire) + a ban arming detection on the retired figure, one supersede.
-- Ban proven non-inert before writing: 13 findings over the live corpus.
BEGIN;
UPDATE site_specs SET is_current=false, superseded_at=now()
 WHERE site_id='1fcfa4f3-ec80-4010-878b-b971cd46711f' AND aspect='evidence_base' AND is_current=true;
INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by, pinned)
SELECT '1fcfa4f3-ec80-4010-878b-b971cd46711f','evidence_base', $json$
{
  "facts": [
    {
      "id": "price_total",
      "kind": "metric",
      "claim": "The website build costs one hundred and forty nine pounds, as a single one-off payment covering the whole build.",
      "value": 149,
      "source": {
        "attested_by": "owner, 2026-08-11, ai_site_selling_automation PLAN §1b ruling 2 (supersedes the £1,200 price attested by the owner on 2026-08-03)"
      },
      "verified_at": "2026-08-11",
      "writer_line": "£149"
    },
    {
      "id": "price_is_total_no_vat",
      "kind": "attestation",
      "claim": "The owner is not VAT registered, so no VAT is added. £149 is the total the customer pays.",
      "source": {
        "attested_by": "owner, 2026-08-04 (VAT status); the price it applies to was restated by owner, 2026-08-11, ai_site_selling_automation PLAN §1b ruling 2"
      },
      "verified_at": "2026-08-11",
      "writer_line": "£149 is the total. No VAT is added to it."
    },
    {
      "id": "payment_after_approval",
      "kind": "capability",
      "claim": "The customer sees the finished site on a private preview link and pays after they have approved it. Nothing is taken before that. This is the current setting of a switch the owner can move to payment up front, so it must be re-checked against billing_settings.payment_timing before it is restated.",
      "source": {
        "sql": "SELECT payment_timing FROM billing_settings",
        "attested_by": "owner, 2026-08-11, ai_site_selling_automation PLAN §1c ruling 4 (collected after approval while the system is being tested; moves to up front later)"
      },
      "verified_at": "2026-08-11",
      "writer_line": "You pay after you have seen the site and approved it."
    },
    {
      "id": "no_refund",
      "kind": "attestation",
      "claim": "No refund is offered. The customer approves the site before paying, so there is nothing to return, and once the payment is made the price is not refundable.",
      "source": {
        "attested_by": "owner, 2026-08-11, ai_site_selling_automation PLAN §1c ruling 4 (refunds are manual and behind the scenes, never offered visibly)"
      },
      "verified_at": "2026-08-11",
      "writer_line": "We do not offer refunds. You pay once you have approved the site."
    },
    {
      "id": "no_changes_included",
      "kind": "attestation",
      "claim": "No changes are included. The site is built and handed over as it is; we do not make revisions to it afterwards. This is what the price buys and the site states it plainly rather than burying it.",
      "value": 0,
      "source": {
        "attested_by": "owner, 2026-08-12, ai_site_selling_automation PLAN §1d (\"At £149 we can't be doing corrections\") — supersedes the one-set-of-changes term attested 2026-08-11"
      },
      "verified_at": "2026-08-12",
      "writer_line": "No changes are included. You get the site as it is built."
    },
    {
      "id": "yours_to_change",
      "kind": "capability",
      "claim": "The customer receives the finished site's own files, which they or anyone they choose can edit. Changing a site that already works is far less work than commissioning one from nothing, so what they are buying is a working starting point they own outright, not a finished article they must accept for ever.",
      "source": {
        "attested_by": "owner, 2026-08-12: a site is much easier to edit than to start from scratch, so even if it is not quite what they would like, it is a start they can move forward from more easily than from a blank page"
      },
      "verified_at": "2026-08-12",
      "writer_line": "The files are yours. Editing a site that already works is a great deal easier than starting from a blank page."
    },
    {
      "id": "taking_it_further",
      "kind": "capability",
      "claim": "Building on the site after it is delivered is not a service offered here. The customer owns the files outright and can take them to any web developer, or any service they choose, to have the site changed or extended. Where it helps, the site points at third-party options rather than at us.",
      "source": {
        "attested_by": "owner, 2026-08-12: third-party services may be listed for customers who want to take their sites further; we could do that work but probably will not, for lack of bandwidth. NOTE: whether paid follow-on work is offered is NOT SETTLED — the site must therefore neither offer it nor rule it out, and this fact licenses only the outward pointer."
      },
      "verified_at": "2026-08-12",
      "writer_line": "Taking the site further is not something we do. The files are yours to hand to any developer you like."
    },
    {
      "id": "third_party_options",
      "kind": "entity",
      "claim": "Named third-party services the site may point customers at, by category, for things a delivered static site does not do on its own. Hosting: Cloudflare Pages and Netlify, both of which serve static files on a free tier. Visitor statistics: Fathom Analytics and Plausible, both privacy-first and cookie-free. Contact forms: Formspree and Basin, both of which accept form posts from a static page and email the result. Each may be named and described in these terms and no further: no ranking, no superlative, no claim about price, and no promise about what any of them will do for the customer.",
      "source": {
        "checked": "each service verified as currently operating and fit for a plain static site, 2026-08-12",
        "attested_by": "owner, 2026-08-12: list third-party services for customers who want to take their sites further; quality matters more than price, and not the agencies everyone else promotes"
      },
      "verified_at": "2026-08-12"
    },
    {
      "id": "queue_limited",
      "kind": "capability",
      "claim": "Only a few sites are built at a time. When every slot is taken, submissions close until one opens again.",
      "source": {
        "attested_by": "owner, 2026-08-11, ai_site_selling_automation PLAN §1b ruling 2 (a visible queue, capacity starting at three or four sites). The capacity number is owner-settable and the counter is not built yet, so no number may be published."
      },
      "verified_at": "2026-08-11",
      "writer_line": "We only build a few sites at a time. When we are full, submissions close until a slot opens."
    },
    {
      "id": "ai_built",
      "kind": "attestation",
      "claim": "The sites are built by an AI system rather than designed page by page by a person. The site states this plainly rather than implying otherwise.",
      "source": {
        "attested_by": "owner, 2026-08-11, ai_site_selling_automation PLAN §1c ruling 2 (no-frills positioning: say plainly that the sites are AI-built)"
      },
      "verified_at": "2026-08-11",
      "writer_line": "The site is built by AI, not designed page by page by a person."
    },
    {
      "id": "delivery_preview_and_zip",
      "kind": "capability",
      "claim": "What the customer receives is a private preview link to review, and a downloadable ZIP of the finished website for them to host wherever they choose.",
      "source": {
        "attested_by": "owner, 2026-08-11, ai_site_selling_automation PLAN §1b ruling 4 (deliverable = private preview plus a downloadable ZIP the customer hosts themselves)"
      },
      "verified_at": "2026-08-11",
      "writer_line": "You get a private preview link, then a ZIP of the finished site to host wherever you like."
    },
    {
      "id": "hosting_and_domain_not_included",
      "kind": "attestation",
      "claim": "Hosting and the domain name are not included. The customer keeps their own domain and their own DNS. Hosting by us and a manual domain transfer are available as paid extras and are optional.",
      "source": {
        "attested_by": "owner, 2026-08-11, ai_site_selling_automation PLAN §1b ruling 4; hosting recommendations and the written setup guide are owner, 2026-08-11, ai_site_selling_automation PLAN §1c ruling 5"
      },
      "verified_at": "2026-08-11",
      "writer_line": "Hosting and your domain are not included. You keep your own domain and DNS."
    },
    {
      "id": "no_lock_in",
      "kind": "capability",
      "claim": "There is no ongoing fee payable to us and no proprietary platform the customer is tied to. The finished site is theirs to host and to change.",
      "source": {
        "attested_by": "owner, PLAN 7a productised-offer ruling 2026-07-29; unchanged by the 2026-08-11 pricing ruling, reworded for ZIP delivery"
      },
      "verified_at": "2026-08-11",
      "writer_line": "No monthly fee, and nothing to be tied into."
    },
    {
      "id": "build_duration",
      "kind": "capability",
      "claim": "From having what is needed from the customer, the site is usually ready the next day.",
      "source": {
        "attested_by": "owner, 2026-08-14 ('usually next day') — supersedes the three-or-four-days figure attested 2026-08-04 under the £1,200 offer and carried over unre-attested until now"
      },
      "verified_at": "2026-08-14",
      "writer_line": "usually ready the next day"
    },
    {
      "id": "contact",
      "kind": "entity",
      "claim": "Enquiries reach webdesign@contactforsales.com and +44 (0) 7934 524 911.",
      "source": {
        "sql": "SELECT email, phone FROM sites WHERE domain='webdesign.uk'"
      },
      "verified_at": "2026-08-12"
    }
  ],
  "audit_doc": "docs/agent_docs/docs024_key_docs_latest/ai_site_selling_automation/ (PLAN 2026-08-10 §1b/§1c — the £149 offer, owner rulings 2026-08-11). The retired £1,200 offer and its copy are archived in that directory under snapshot_2026-08-11_gbp1200_offer/; the evidence_base row it governed is superseded in site_specs, not deleted.",
  "schema_notes": "facts[]: {id, claim, kind, source, verified_at, value?, writer_line?}. banned_claims[]: {pattern (case-insensitive regex; an invalid regex degrades to a literal substring), reason}.",
  "writer_block": "HOW THIS SITE SHOULD SOUND, AND THE THINGS IT MUST NOT DO.\n\nNever use an em dash. Not anywhere, not once. Where you want one, use a full stop, a comma, a colon or brackets. The owner reads em dashes as a tell that copy was machine-written, and on a site selling web design that impression is fatal.\n\nDo not use the word 'honest' or 'honestly' anywhere. The owner's standing instruction, applied across every site on 2026-08-12: the honesty has to be shown by what the page says, not claimed by labelling it. A page that calls itself honest is doing the opposite.\n\nDo not say that a person checks, reviews or looks over the site before the customer sees it. The earlier version of this page said 'a person checks it before you ever see it' and the owner removed it, because it makes the service sound like a template that gets a quick glance. It undersells what is actually done. Write about the work itself rather than about it being inspected.\n\nAsk the visitor a little about themselves and what they want, and only a little. A couple of questions in the page's own voice, or one short set of fields. Not a long form, not a questionnaire, and nothing that has to be completed before they can talk to us. The point is to open a conversation, not to qualify a lead.\n\nSay how long it takes: usually ready the next day from having what is needed. Give it plainly, as a normal fact about how the work goes. Never say three or four days, and never a range of days: that was the old offer's figure and it is retired.\n\nDo not oversell. This service is new and has nothing to boast about yet, so restraint is not modesty here, it is accuracy. No client counts, no testimonials, no awards, no years of experience, no team size, no superlatives about being the best or the fastest. If a sentence needs a boast to work, the sentence is wrong.\n\nTHE OFFER, IN FULL, AND THIS IS THE ONLY VERSION YOU MAY STATE.\n\nThe site costs £149. That is the total: the owner is not VAT registered, so nothing is added to it. It is one payment, not a subscription, and there is no monthly fee.\n\nThe customer pays after they have seen the finished site on a private preview link and approved it. Nothing is taken before that.\n\nWe do not offer refunds. Write it in those words, with 'do not' or 'never' in the same clause, and never as 'no refund' on its own: the platform's claims gate treats a bare 'no' as an intensifier rather than a negation, so 'there is no refund' reads to it as a refund promise and blocks the page. Never describe a refund, a refund window, a deposit, or money coming back in any form. There is no deposit and no fourteen-day window any more; both belonged to the £1,200 offer and were retired on 2026-08-11.\n\nNo changes are included, and the page says so plainly rather than hiding it in a clause. We do not revise the site after it is built. Never offer a round, a revision, an amendment or a tweak, and never imply that asking nicely might get one.\n\nImmediately alongside that, make the argument that makes it fair, because it is true and it is the reason the offer works: the customer gets the site's own files and they are theirs outright. They, or anyone they hire, can change anything they like. Editing a site that already works is far less work than starting from a blank page, so what they are buying is a working starting point rather than a finished article they are stuck with. Put it in that order — the hard term first, then this — and never as an apology or a softener. It is not a consolation for the missing revisions; it is the better half of the deal, and it should read like one.\n\nWhere the reader is plainly asking for more than the offer contains, point OUTWARD, never at us. Taking the site further is not a service offered here, and third-party options may be listed for anyone who wants them. Two things you must not write, for opposite reasons: never offer our time, at any price, and never state that we would never do it either. The first promises work we have not agreed to; the second forecloses a decision the owner has explicitly left open. Say what the customer can do, name where they can go, and stop. Wherever the page answers the question of taking the site further — the FAQ's answer on that subject in particular — DO name the six services, in the answer's own voice, as a short factual list grouped by what they are for. This is standing content, not an option: an answer about taking the site further that names nowhere to go is incomplete. The six and their groups, exactly and no others: For hosting the files: Cloudflare Pages and Netlify. For seeing who visits: Fathom Analytics and Plausible. For making a contact form work: Formspree and Basin. Never rank them, never call any of them the best, never quote a price for something we do not sell, and never promise what one of them will do for the reader. We are pointing at a door, not vouching for the room.\n\nWe build only a few sites at a time, and when we are full, submissions close until a slot opens. Do not put a number on how many: the capacity is set by the owner and the queue counter is not built yet.\n\nSay plainly that the site is built by AI. This is the positioning, not an admission: what the customer gets for £149 is a real, complete website built by software rather than an agency's hours, and pretending otherwise would be both untrue and off-strategy. Do not dress it up and do not apologise for it.\n\nSay what the customer gets and what they do not, with equal clarity. They get a private preview link and then a ZIP of the finished site, which they host wherever they like. Hosting and the domain are not included and stay theirs. Hosting by us and a manual domain transfer are optional paid extras. Never say that we handle the setup, the hosting or the domain: that promise belonged to the old offer and is now false.\n\nWhat you may state as fact is in facts[] above and nothing else: the £149 price, that it is the total with no VAT, that payment comes after approval, that there is no refund, that no changes are included, that the files are the customer's own to edit, that taking the site further is not a service offered here, that only a few sites are built at a time, that the site is AI-built, the preview link and ZIP, that hosting and the domain are not included, that there is no lock-in, the rough duration, the contact details, and the six named third-party services (Cloudflare Pages, Netlify, Fathom Analytics, Plausible, Formspree, Basin). Every other number, date, quantity or proportion is off limits. If a claim seems necessary and is not in that list, rewrite around it or leave it out.\n\nRegister: plain, direct British English, written for a business owner who is not technical and does not enjoy being sold to. Short sentences. Ordinary words. No agency vocabulary, no 'solutions', no 'bespoke digital experiences', no 'we're passionate about'. Say the thing, then stop.\n\nNow the harder half, and it is the whole strategy rather than a tone preference. This is a take-it-or-leave-it offer and the copy has to say so hard enough that nobody arrives expecting otherwise. State the terms early, flatly, in the plainest words available, and never soften them with a qualifier, a 'but', or a reassuring clause. What you pay for is what you get. Write it like someone behind a trade counter who knows exactly what he is selling and is not going to talk you into it: unhurried, unbothered, faintly amused, and completely unapologetic. Confidence, never aggression, and never rudeness for its own sake — the flatness IS the confidence.\n\nTwo things this register must never become. It is not self-deprecating: we are not admitting to a cheap service, we are describing a deliberate one, and a sentence that sounds like an apology has got it exactly backwards. And it is not scruffy WRITING. The words are the one thing on this site that must be immaculate — every sentence earning its place, no padding, no throat-clearing, nothing that could be cut. The offer is stripped back; the craft never is. A reader should finish a page thinking these people are good, and cheap because they have taken the fuss out rather than because they cut corners.",
  "banned_claims": [
    {
      "reason": "STYLE, NOT A CLAIM. The owner's explicit instruction 2026-08-04: the em dash reads as automated and unprofessional. Use a full stop, a comma, a colon or brackets. This is banned as a regex because it is the only enforcing lever on this path.",
      "pattern": "—"
    },
    {
      "reason": "The owner's instruction 2026-08-04: this phrasing makes the product sound like a template that gets a brief glance, which undersells the work and is the opposite of the positioning. Describe what is actually done, not that it is inspected.",
      "pattern": "(a (person|human)|someone|we) (then )?(check|checks|reviews|looks over|goes over|casts an eye)"
    },
    {
      "reason": "Same instruction. The original line 'a person checks it before you ever see it' is the specific sentence the owner asked to remove.",
      "pattern": "(before|until) you (ever )?see it"
    },
    {
      "reason": "Do not describe the product this way even to deny it. Denying a frame repeats it.",
      "pattern": "(template|templated|off.the.shelf|cookie.cutter)"
    },
    {
      "reason": "Social-proof class. This is a new service with no customers; any number here is fabricated.",
      "pattern": "(trusted|used|loved|chosen) by [0-9]"
    },
    {
      "reason": "Scale class. No delivery history exists.",
      "pattern": "[0-9][0-9,]* ?(clients|customers|businesses|companies|websites|sites) (built|served|helped|delivered|launched)"
    },
    {
      "reason": "Tenure class. This service is new and has no trading history.",
      "pattern": "years of (experience|expertise)|since [0-9]{4}"
    },
    {
      "reason": "Reputation class. No awards and no press coverage exist.",
      "pattern": "(award.winning|multi.award|as seen (in|on)|featured in)"
    },
    {
      "reason": "Headcount class. Do not state or imply a team size.",
      "pattern": "(our|the) (team|studio|agency|designers|developers) (of|are|is) [0-9]"
    },
    {
      "reason": "Testimonial shape: a long quotation followed by an attributed name. There are no customers and therefore no testimonials.",
      "pattern": "\"[^\"]{20,}\" ?[—,-]? ?[A-Z][a-z]+ [A-Z]"
    },
    {
      "reason": "Overclaims the guarantee. The real term is narrower and specific: a refund is available until the customer accepts the site. State that, not a slogan.",
      "pattern": "(100%|fully) (guaranteed|satisfaction)|money.back guarantee, no questions"
    },
    {
      "reason": "Speed class. The attested figure is about three or four days; anything faster is unsupported and also fights the positioning.",
      "pattern": "(instant|instantly|in minutes|in seconds|overnight|same day)"
    },
    {
      "reason": "No SEO outcome has been measured or promised, and ranking guarantees are unsupportable.",
      "pattern": "(seo|search engine) (optimised|optimized|ranking|rankings|guaranteed)"
    },
    {
      "reason": "Do not claim hand-coding. The sites are framework-built, which is the actual strength; claiming the opposite is both untrue and off-strategy.",
      "pattern": "(bespoke|custom|tailor.made) (from scratch|hand.coded|hand.written)"
    },
    {
      "reason": "Superlative market-position class: unmeasurable, unsupportable for a new service, and the prompt-side house voice discourages but does not forbid it. Added 2026-08-06 after the prompt review.",
      "pattern": "(best|leading|number ?one|no\\.? ?1|top.rated|premier) (web ?design|agency|studio|developer|choice|service)"
    },
    {
      "reason": "Caps ruling, owner 2026-08-09: revisions are capped (see facts revision_rounds_included); an uncapped-changes promise contradicts the offer.",
      "pattern": "(unlimited|no limit to|no limits on|as many (changes|revisions))"
    },
    {
      "reason": "Caps ruling, owner 2026-08-09: no open-ended time promises; the review window is the bound.",
      "pattern": "(at any ?(point|time)|any time before|whenever you like|no time limit)"
    },
    {
      "reason": "RETIRED OFFER (owner ruling 2026-08-11, PLAN §1c.1): the £1,200 price is off the table and superseded by £149. Any page still quoting it is stating a price we will not honour.",
      "pattern": "£ ?1,?200\\b|\\bone thousand two hundred\\b"
    },
    {
      "reason": "RETIRED OFFER (owner ruling 2026-08-11): the £75 non-refundable deposit and the £1,125 balance belonged to the £1,200 refund model. Neither exists now.",
      "pattern": "£ ?75\\b|£ ?1,125\\b"
    },
    {
      "reason": "RETIRED OFFER (owner ruling 2026-08-11): there is no deposit. Payment is taken once, after the customer approves the site.",
      "pattern": "\\bdeposits?\\b"
    },
    {
      "reason": "RETIRED OFFER (owner ruling 2026-08-11, PLAN §1c.4): no refund is offered and none may be described. NOTE FOR THE WRITER: the negation guard does not treat a bare 'no' as a negator, so 'there is no refund' IS flagged and will block the page. Write 'we do not offer refunds', which the guard suppresses correctly.",
      "pattern": "\\brefunds?\\b|\\brefundable\\b|\\bmoney.back\\b"
    },
    {
      "reason": "RETIRED OFFER (owner ruling 2026-08-11): the fourteen-day review window was the bound on the refund right, and went with it.",
      "pattern": "\\b(14|fourteen)[ -]days?\\b"
    },
    {
      "reason": "RETIRED OFFER (owner ruling 2026-08-11, PLAN §1b.4/§1c.7): hosting and the domain are NOT included and stay with the customer. The FAQ's 'we handle the setup as part of getting the site live' is the specific sentence this retires.",
      "pattern": "\\bwe (handle|take care of|sort out|deal with|manage) (the )?(setup|set.up|hosting|domain)"
    },
    {
      "reason": "RETIRED OFFER (owner ruling 2026-08-11): these promises belonged to the £1,200 preview-then-refund model. Payment after approval is stated by fact payment_after_approval, which is a switch and may flip to up front; a slogan built on it would then be false.",
      "pattern": "\\byou only pay if you (like|love|are happy|want)\\b|\\bbefore any money changes hands\\b"
    },
    {
      "reason": "STYLE, NOT A CLAIM. Owner instruction 2026-08-12, applied across every site, extending the leopardessconsulting ruling of 2026-07-18: 'overused; show the honesty, do not label it'. Banned mechanically here because the content_direction ban is prompt-side and advisory only.",
      "pattern": "\\bhonest(ly|y)?\\b"
    },
    {
      "reason": "RETIRED OFFER (owner ruling 2026-08-12, PLAN §1d): no changes are included at £149. This supersedes BOTH the two-rounds term of 2026-08-09 and the one-set-of-changes term of 2026-08-11 — the second was live on all five pages for four hours before the ruling retired it.",
      "pattern": "\\brounds? of (revisions?|changes)\\b|\\b(two|2) rounds?\\b|\\b(one|1|a) (set of )?(changes?|revisions?) (is |are )?included\\b|\\bincludes? (one|1|a) (set of )?(changes?|revisions?)\\b"
    },
    {
      "reason": "RETIRED OFFER (owner ruling 2026-08-12, PLAN §1d): we do not revise the site after it is built, for any price stated on this site. The customer's remedy is that the FILES ARE THEIRS to edit (fact yours_to_change) — never an offer of our time.",
      "pattern": "\\b(revisions?|amendments?|tweaks?) (are |is )?(included|free|at no (extra )?(cost|charge))\\b|\\bwe('ll| will| can) (make|do) (the |any |your )?(changes?|revisions?|tweaks?|amendments?)\\b"
    },
    {
      "pattern": "\\bthree (or|to) four days\\b|\\b3[-–]4 days\\b|\\bthree[-–]to[-–]four\\b",
      "reason": "RETIRED FIGURE (owner 2026-08-14): the build time is 'usually ready the next day' (fact build_duration). Three-or-four-days belonged to the £1,200 offer."
    }
  ],
  "governing_rule": "This site sells a website build service. Every commercial term stated on it (price, what is included, how long it takes, what happens if the customer is unhappy) must trace to a fact below, each of which is an owner attestation. There is no trading history, no client list, no testimonials and no awards, so no claim of scale, tenure, popularity or reputation may appear in any form. Where something is not yet settled, the page says nothing about it rather than guessing: silence is always publishable and a plausible-sounding invention never is.",
  "allowed_entities": [
    "webdesign.uk",
    "Stripe",
    "the United Kingdom",
    "UK",
    "Fathom Analytics",
    "Plausible",
    "Formspree",
    "Basin",
    "Cloudflare",
    "Netlify"
  ]
}
$json$::jsonb,'owner_ruling',
 'build_duration re-attested: usually ready the next day (owner 2026-08-14). '
 'Retires three-or-four-days; ban added; writer_block updated in the same edit.',
 true,'ai-site-selling-automation-2026-08-14',COALESCE(pinned,false)
FROM site_specs WHERE site_id='1fcfa4f3-ec80-4010-878b-b971cd46711f'
  AND aspect='evidence_base' AND is_current=false
ORDER BY superseded_at DESC NULLS LAST LIMIT 1;
DO $$
DECLARE wb text; f jsonb; is_pinned bool; n_bans int;
BEGIN
  SELECT data->>'writer_block',
         (SELECT x FROM jsonb_array_elements(data->'facts') x WHERE x->>'id'='build_duration'),
         COALESCE(pinned,false), jsonb_array_length(data->'banned_claims')
    INTO wb, f, is_pinned, n_bans
    FROM site_specs WHERE site_id='1fcfa4f3-ec80-4010-878b-b971cd46711f' AND aspect='evidence_base' AND is_current;
  IF NOT is_pinned THEN RAISE EXCEPTION 'pinned not inherited'; END IF;
  IF f->>'writer_line' <> 'usually ready the next day' THEN RAISE EXCEPTION 'fact not re-attested'; END IF;
  IF wb NOT LIKE '%usually ready the next day%' THEN RAISE EXCEPTION 'the wire still says the old figure'; END IF;
  IF wb LIKE '%about three or four days from having%' THEN RAISE EXCEPTION 'old sentence survived in the wire'; END IF;
  IF n_bans <> 28 THEN RAISE EXCEPTION 'ban count: expected 28, got %', n_bans; END IF;
END $$;
COMMIT;
