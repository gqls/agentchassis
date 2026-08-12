-- FILE: SQL_2026-08-12_evidence_base_149.sql
--
-- webdesign.uk: migrate evidence_base from the retired £1,200 offer to the
-- ruled £149 one (owner rulings 2026-08-11, ai_site_selling_automation
-- PLAN §1b/§1c). This is the SOURCE-OF-TRUTH half of the copy migration: it
-- changes what the site is ALLOWED to say, and arms detection against what it
-- must stop saying. The page copy itself is regenerated separately, through
-- the framework.
--
-- WHY SUPERSEDE RATHER THAN UPDATE IN PLACE. site_specs carries its own
-- history (is_current / superseded_at / idx_site_specs_history) and the
-- platform's own writer uses it (refresh_evidence_base_action.go
-- writeRefreshedEvidenceBase). The previous £1,200 -> £75-deposit change was
-- an in-place UPDATE, which destroyed the pre-deposit row; the owner ruled the
-- £1,200 offer "ARCHIVED restorably", and a superseded row is what makes that
-- true in the database rather than only in a file snapshot.
--
-- WHAT CHANGED, measured with cmd/claimscan against all 25 live components
-- BEFORE running this (see NOTES 2026-08-12):
--   facts        10 -> 12   banned_claims 18 -> 26   writer_block 2,932 -> 4,774 chars
--   the new ban set raises 3 findings -> 36, across five live surfaces
--   (faq 8, how-it-works 6, brief-starter-guide 4, what-you-get 3, index 1,
--   plus 14 on the archived index-rejected page). The intended replacement
--   copy scans CLEAN against the same set, with exactly one match suppressed
--   by the negation guard ("we do not offer refunds"), which is the phrasing
--   the writer_block now mandates and the ban's own reason explains.
--
-- ROLLBACK: the £1,200 row is one UPDATE away.
--   UPDATE site_specs SET is_current=false, superseded_at=now()
--    WHERE site_id='1fcfa4f3-ec80-4010-878b-b971cd46711f' AND aspect='evidence_base' AND is_current;
--   UPDATE site_specs SET is_current=true, superseded_at=NULL
--    WHERE id='<the superseded row id printed below>';

BEGIN;

\set site_id '1fcfa4f3-ec80-4010-878b-b971cd46711f'

-- 1. Supersede the current row, keeping it readable.
UPDATE site_specs
   SET is_current = false, superseded_at = now()
 WHERE site_id = :'site_id' AND aspect = 'evidence_base' AND is_current = true;

-- 2. Insert the £149 row, inheriting `pinned` from the row it replaces.
INSERT INTO site_specs
       (site_id, aspect, data, source, source_agent, notes, is_current, created_by, pinned)
SELECT :'site_id', 'evidence_base', $json$
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
        "attested_by": "owner, 2026-08-11, ai_site_selling_automation PLAN §1c ruling 4 (collected after approval while the system is being tested; moves to up front later)",
        "sql": "SELECT payment_timing FROM billing_settings"
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
      "id": "changes_included_one_set",
      "kind": "attestation",
      "claim": "One set of changes is included in the price. The customer asks for their changes once and they are made; anything after that is charged as work.",
      "value": 1,
      "source": {
        "attested_by": "owner, 2026-08-11, ai_site_selling_automation PLAN §1b ruling 2 (\"One set of changes included\"; supersedes the two-rounds term attested 2026-08-09)"
      },
      "verified_at": "2026-08-11",
      "writer_line": "One set of changes is included in the price."
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
      "claim": "From having what is needed from the customer, the build takes up to about three or four days before they are shown it.",
      "source": {
        "attested_by": "owner, 2026-08-04, attested under the previous £1,200 offer. The 2026-08-11 rulings did not restate it, so it is carried over unchanged and is DUE RE-ATTESTATION at £149."
      },
      "verified_at": "2026-08-04",
      "writer_line": "usually about three or four days"
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
  "writer_block": "HOW THIS SITE SHOULD SOUND, AND THE THINGS IT MUST NOT DO.\n\nNever use an em dash. Not anywhere, not once. Where you want one, use a full stop, a comma, a colon or brackets. The owner reads em dashes as a tell that copy was machine-written, and on a site selling web design that impression is fatal.\n\nDo not use the word 'honest' or 'honestly' anywhere. The owner's standing instruction, applied across every site on 2026-08-12: the honesty has to be shown by what the page says, not claimed by labelling it. A page that calls itself honest is doing the opposite.\n\nDo not say that a person checks, reviews or looks over the site before the customer sees it. The earlier version of this page said 'a person checks it before you ever see it' and the owner removed it, because it makes the service sound like a template that gets a quick glance. It undersells what is actually done. Write about the work itself rather than about it being inspected.\n\nAsk the visitor a little about themselves and what they want, and only a little. A couple of questions in the page's own voice, or one short set of fields. Not a long form, not a questionnaire, and nothing that has to be completed before they can talk to us. The point is to open a conversation, not to qualify a lead.\n\nSay how long it takes: usually about three or four days from having what is needed. Give it plainly, as a normal fact about how the work goes.\n\nDo not oversell. This service is new and has nothing to boast about yet, so restraint is not modesty here, it is accuracy. No client counts, no testimonials, no awards, no years of experience, no team size, no superlatives about being the best or the fastest. If a sentence needs a boast to work, the sentence is wrong.\n\nTHE OFFER, IN FULL, AND THIS IS THE ONLY VERSION YOU MAY STATE.\n\nThe site costs £149. That is the total: the owner is not VAT registered, so nothing is added to it. It is one payment, not a subscription, and there is no monthly fee.\n\nThe customer pays after they have seen the finished site on a private preview link and approved it. Nothing is taken before that.\n\nWe do not offer refunds. Write it in those words, with 'do not' or 'never' in the same clause, and never as 'no refund' on its own: the platform's claims gate treats a bare 'no' as an intensifier rather than a negation, so 'there is no refund' reads to it as a refund promise and blocks the page. Never describe a refund, a refund window, a deposit, or money coming back in any form. There is no deposit and no fourteen-day window any more; both belonged to the £1,200 offer and were retired on 2026-08-11.\n\nOne set of changes is included. The customer asks for their changes once, they are made, and anything after that is charged as work. Never say rounds, never say two, and never leave the number of changes open.\n\nWe build only a few sites at a time, and when we are full, submissions close until a slot opens. Do not put a number on how many: the capacity is set by the owner and the queue counter is not built yet.\n\nSay plainly that the site is built by AI. This is the positioning, not an admission: what the customer gets for £149 is a real, complete website built by software rather than an agency's hours, and pretending otherwise would be both untrue and off-strategy. Do not dress it up and do not apologise for it.\n\nSay what the customer gets and what they do not, with equal clarity. They get a private preview link and then a ZIP of the finished site, which they host wherever they like. Hosting and the domain are not included and stay theirs. Hosting by us and a manual domain transfer are optional paid extras. Never say that we handle the setup, the hosting or the domain: that promise belonged to the old offer and is now false.\n\nWhat you may state as fact is in facts[] above and nothing else: the £149 price, that it is the total with no VAT, that payment comes after approval, that there is no refund, that one set of changes is included, that only a few sites are built at a time, that the site is AI-built, the preview link and ZIP, that hosting and the domain are not included, that there is no lock-in, the rough duration, and the contact details. Every other number, date, quantity or proportion is off limits. If a claim seems necessary and is not in that list, rewrite around it or leave it out.\n\nRegister: plain, direct British English, written for a business owner who is not technical and does not enjoy being sold to. Short sentences. Ordinary words. No agency vocabulary, no 'solutions', no 'bespoke digital experiences', no 'we're passionate about'. Say the thing, then stop. The offer is a cheap, fast, no-frills one and the copy should sound like it knows that: what you pay for is what you get, said without apology and without inflation.",
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
      "pattern": "£ ?1,?200\\b|\\bone thousand two hundred\\b",
      "reason": "RETIRED OFFER (owner ruling 2026-08-11, PLAN §1c.1): the £1,200 price is off the table and superseded by £149. Any page still quoting it is stating a price we will not honour."
    },
    {
      "pattern": "£ ?75\\b|£ ?1,125\\b",
      "reason": "RETIRED OFFER (owner ruling 2026-08-11): the £75 non-refundable deposit and the £1,125 balance belonged to the £1,200 refund model. Neither exists now."
    },
    {
      "pattern": "\\bdeposits?\\b",
      "reason": "RETIRED OFFER (owner ruling 2026-08-11): there is no deposit. Payment is taken once, after the customer approves the site."
    },
    {
      "pattern": "\\brefunds?\\b|\\brefundable\\b|\\bmoney.back\\b",
      "reason": "RETIRED OFFER (owner ruling 2026-08-11, PLAN §1c.4): no refund is offered and none may be described. NOTE FOR THE WRITER: the negation guard does not treat a bare 'no' as a negator, so 'there is no refund' IS flagged and will block the page. Write 'we do not offer refunds', which the guard suppresses correctly."
    },
    {
      "pattern": "\\b(14|fourteen)[ -]days?\\b",
      "reason": "RETIRED OFFER (owner ruling 2026-08-11): the fourteen-day review window was the bound on the refund right, and went with it."
    },
    {
      "pattern": "\\brounds? of (revisions?|changes)\\b|\\b(two|2) rounds?\\b",
      "reason": "RETIRED OFFER (owner ruling 2026-08-11, PLAN §1b.2): two rounds of revisions is superseded by one set of changes. Say changes, not rounds, and never two."
    },
    {
      "pattern": "\\bwe (handle|take care of|sort out|deal with|manage) (the )?(setup|set.up|hosting|domain)",
      "reason": "RETIRED OFFER (owner ruling 2026-08-11, PLAN §1b.4/§1c.7): hosting and the domain are NOT included and stay with the customer. The FAQ's 'we handle the setup as part of getting the site live' is the specific sentence this retires."
    },
    {
      "pattern": "\\byou only pay if you (like|love|are happy|want)\\b|\\bbefore any money changes hands\\b",
      "reason": "RETIRED OFFER (owner ruling 2026-08-11): these promises belonged to the £1,200 preview-then-refund model. Payment after approval is stated by fact payment_after_approval, which is a switch and may flip to up front; a slogan built on it would then be false."
    },
    {
      "pattern": "\\bhonest(ly|y)?\\b",
      "reason": "STYLE, NOT A CLAIM. Owner instruction 2026-08-12, applied across every site, extending the leopardessconsulting ruling of 2026-07-18: 'overused; show the honesty, do not label it'. Banned mechanically here because the content_direction ban is prompt-side and advisory only."
    }
  ],
  "governing_rule": "This site sells a website build service. Every commercial term stated on it (price, what is included, how long it takes, what happens if the customer is unhappy) must trace to a fact below, each of which is an owner attestation. There is no trading history, no client list, no testimonials and no awards, so no claim of scale, tenure, popularity or reputation may appear in any form. Where something is not yet settled, the page says nothing about it rather than guessing: silence is always publishable and a plausible-sounding invention never is.",
  "allowed_entities": [
    "webdesign.uk",
    "Stripe",
    "the United Kingdom",
    "UK"
  ]
}
$json$::jsonb,
       'owner_ruling',
       NULL,
       'The £149 offer. Owner rulings 2026-08-11 (ai_site_selling_automation '
       'PLAN §1b rulings 2 and 4, §1c rulings 1, 2, 4 and 5) retire the £1,200 '
       'price, the £75 non-refundable deposit, the fourteen-day refund window '
       'and the two-rounds revision cap, and add: payment after approval, no '
       'refund, one set of changes, a capped queue, AI-built stated plainly, '
       'ZIP delivery with the customer keeping their own domain and DNS. '
       'build_duration is CARRIED OVER from the previous offer and is due '
       're-attestation. Ban set verified non-inert with cmd/claimscan before '
       'writing: 3 findings -> 36 over the live components.',
       true,
       'ai-site-selling-automation-2026-08-12',
       COALESCE(pinned, false)
  FROM site_specs
 WHERE site_id = :'site_id' AND aspect = 'evidence_base' AND is_current = false
 ORDER BY superseded_at DESC NULLS LAST
 LIMIT 1;

-- 3. Assert, and ABORT if wrong. A verify block of bare SELECTs cannot stop a
--    COMMIT (ON_ERROR_STOP ignores a non-empty result), so these are RAISEs.
DO $$
DECLARE
  n_current int; n_facts int; n_bans int; has_block bool; is_pinned bool;
BEGIN
  SELECT count(*) INTO n_current FROM site_specs
   WHERE site_id='1fcfa4f3-ec80-4010-878b-b971cd46711f' AND aspect='evidence_base' AND is_current;
  IF n_current <> 1 THEN
    RAISE EXCEPTION 'expected exactly 1 current evidence_base row, found %', n_current;
  END IF;

  SELECT jsonb_array_length(data->'facts'), jsonb_array_length(data->'banned_claims'),
         data ? 'writer_block', COALESCE(pinned,false)
    INTO n_facts, n_bans, has_block, is_pinned
    FROM site_specs
   WHERE site_id='1fcfa4f3-ec80-4010-878b-b971cd46711f' AND aspect='evidence_base' AND is_current;

  IF n_facts <> 12 THEN RAISE EXCEPTION 'facts: expected 12, got %', n_facts; END IF;
  IF n_bans  <> 26 THEN RAISE EXCEPTION 'banned_claims: expected 26, got %', n_bans; END IF;
  IF NOT has_block THEN RAISE EXCEPTION 'writer_block missing: the writer would lose every number'; END IF;
  IF NOT is_pinned THEN RAISE EXCEPTION 'pinned was not inherited: an agent could overwrite the register'; END IF;

  -- The retired price must not survive in the FACTS, which are what the site
  -- is licensed to assert. It deliberately DOES still appear in banned_claims
  -- reasons and in writer_block, where it is named as the thing not to say, so
  -- a whole-document text check here would abort on its own explanations.
  IF EXISTS (SELECT 1 FROM site_specs
              WHERE site_id='1fcfa4f3-ec80-4010-878b-b971cd46711f' AND aspect='evidence_base'
                AND is_current AND (data->'facts')::text ~ '1,200 is|costs .{0,20}1,200|"value": ?1200')
  THEN RAISE EXCEPTION 'the retired £1,200 price is still asserted by a fact'; END IF;

  -- And the new price must actually be there: an assertion that only forbids
  -- cannot tell a correct document from an empty one.
  IF NOT EXISTS (SELECT 1 FROM site_specs
                  WHERE site_id='1fcfa4f3-ec80-4010-878b-b971cd46711f' AND aspect='evidence_base'
                    AND is_current
                    AND data->'facts' @> '[{"id":"price_total","value":149}]'::jsonb)
  THEN RAISE EXCEPTION 'fact price_total is not 149'; END IF;
END $$;

COMMIT;
