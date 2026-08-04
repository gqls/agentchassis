-- ============================================================================
-- webdesign.uk — site row + evidence_base + imagery_style_guide
-- Written 2026-08-04. Applied out of band (psql -f), NOT via the migration
-- runner: this is per-site setup, not a platform schema change.
--
-- WHY THIS FILE EXISTS AT ALL
--   The first webdesign.uk shopfront was HAND-WRITTEN as a single HTML file and
--   uploaded straight to the bucket (2026-08-03). That was wrong on two counts:
--   we sell framework-built sites, so a hand-rolled shopfront demonstrates
--   nothing; and it skips every control the framework applies. The owner called
--   it and asked for the site to be built through the framework instead. This
--   seed is the front half of that; the dispatch is 082_submit_domain_unified.sh.
--
-- WHY THESE THREE, AND WHY BEFORE SUBMISSION  (same reasoning as oufe's seed)
--   1. sites row with an EMAIL. bugs_open/063: the hallucinated-email check
--      FAILS OPEN when a site has no contact email. ensure_site_record upserts
--      on domain, so pre-creating is safe and wins the race.
--   2. evidence_base. The whole claims layer is gated on the PRESENCE of this
--      aspect: loadEvidenceBase returns nil and every lane silently no-ops
--      (validate_page_content.go:727-746). Seeding it BEFORE the first page is
--      written is the only way the first page is covered.
--   3. imagery_style_guide. bugs_closed/027: content_hero generates unstyled on
--      a site that has none, and a brand-new site has none.
--
-- THE FACTS ARE POPULATED HERE, WHICH IS THE OPPOSITE OF oufe's SEED
--   oufe shipped with facts[] EMPTY and a writer_block forbidding every number,
--   because nothing had been verified. Here the commercial facts ARE settled by
--   the owner, so they are registered as attestations. That is what lets the
--   writer state the price at all: the oufe runbook's "no figures in the brief"
--   rule pushes numbers OUT of the mission brief and into exactly this table,
--   where each one carries who attested it and when.
--
-- THE EM DASH BAN IS A STYLE RULE IN A CLAIMS MECHANISM, DELIBERATELY
--   banned_claims is a case-insensitive regex sweep over written copy. It is
--   built for fabrication patterns, and an em dash is a punctuation preference,
--   not a claim. It is here anyway because it is the only ENFORCING lever on
--   this path: writer_block is instruction the model may drift from, and the
--   owner's objection ("sounds very automated") is precisely the kind of thing
--   that returns quietly three pages later. If a cleaner style-rule aspect is
--   wired later, move it and delete this note.
--   [UNVERIFIED] I have not read the sweep's call site to confirm it runs over
--   every slot; if the em dash survives the first build, that is the thing to
--   check first, not the writer_block.
-- ============================================================================

\set ON_ERROR_STOP on

BEGIN;

-- ---------------------------------------------------------------- site row --
INSERT INTO sites (domain, name, network_id, status, email, phone, company_name)
VALUES (
  'webdesign.uk',
  'webdesign.uk',
  '00000000-0000-0000-0000-000000000002',
  'active',
  'webdesign@contactforsales.com',
  '+44 (0) 7934 524 911',
  'webdesign.uk'
)
ON CONFLICT (domain) DO UPDATE
  SET email = COALESCE(sites.email, EXCLUDED.email),
      phone = COALESCE(sites.phone, EXCLUDED.phone);
-- NOTE: the email is SHARED with webdesign.co.uk, which is a different product
-- (a 63-tool guide site). That is deliberate: it is a real address on a domain
-- the owner controls, and inventing webdesign.uk@... would be the fabricated-
-- contact defect class. OWNER: if enquiries need separating, change it here
-- BEFORE the build writes it onto the contact page.

-- ----------------------------------------------------------- evidence_base --
UPDATE site_specs SET is_current = false, superseded_at = now()
WHERE site_id = (SELECT id FROM sites WHERE domain = 'webdesign.uk')
  AND aspect = 'evidence_base' AND is_current;

INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, pinned, created_by)
SELECT
  id,
  'evidence_base',
  $eb${
    "governing_rule": "This site sells a website build service. Every commercial term stated on it (price, what is included, how long it takes, what happens if the customer is unhappy) must trace to a fact below, each of which is an owner attestation. There is no trading history, no client list, no testimonials and no awards, so no claim of scale, tenure, popularity or reputation may appear in any form. Where something is not yet settled, the page says nothing about it rather than guessing: silence is always publishable and a plausible-sounding invention never is.",
    "audit_doc": "docs/agent_docs/docs024_key_docs_latest/webdesign_uk_build_service/ (PLAN 7a offer, 7d price, HANDOFF 2026-08-03 P1)",
    "schema_notes": "facts[]: {id, claim, kind, source, verified_at, value?, writer_line?}. banned_claims[]: {pattern (case-insensitive regex; an invalid regex degrades to a literal substring), reason}.",
    "facts": [
      {
        "id": "price_total",
        "claim": "The website build costs one thousand two hundred pounds, as a single one-off payment covering the whole build.",
        "kind": "metric",
        "source": {"attested_by": "owner, 2026-08-03, in-session confirmation of PLAN 7d recommendation"},
        "verified_at": "2026-08-03",
        "value": 1200,
        "writer_line": "£1,200"
      },
      {
        "id": "price_is_total_no_vat",
        "claim": "The owner is not VAT registered, so no VAT is added. £1,200 is the total the customer pays.",
        "kind": "attestation",
        "source": {"attested_by": "owner, 2026-08-04"},
        "verified_at": "2026-08-04",
        "writer_line": "£1,200 is the total. No VAT is added to it."
      },
      {
        "id": "build_duration",
        "claim": "From having what is needed from the customer, the build takes up to about three or four days before they are shown it.",
        "kind": "capability",
        "source": {"attested_by": "owner, 2026-08-04"},
        "verified_at": "2026-08-04",
        "writer_line": "usually about three or four days"
      },
      {
        "id": "refund_until_acceptance",
        "claim": "The customer sees the finished site on a private preview link before paying. A full refund is available at any point until they accept it. Acceptance ends the guarantee and hands the site over.",
        "kind": "capability",
        "source": {"attested_by": "owner, PLAN 7a ruling 2026-07-29"},
        "verified_at": "2026-07-29"
      },
      {
        "id": "changes_paid_defects_free",
        "claim": "After the customer accepts the site, changes they ask for are charged as work, and anything that is our own mistake is fixed at no cost.",
        "kind": "capability",
        "source": {"attested_by": "owner, PLAN 7a fee boundary ruling 2026-07-29"},
        "verified_at": "2026-07-29"
      },
      {
        "id": "no_lock_in",
        "claim": "The site is delivered on the customer's own domain. There is no ongoing fee payable to us and no proprietary platform they are tied to.",
        "kind": "capability",
        "source": {"attested_by": "owner, PLAN 7a productised-offer ruling"},
        "verified_at": "2026-07-29"
      },
      {
        "id": "contact",
        "claim": "Enquiries reach webdesign@contactforsales.com and +44 (0) 7934 524 911.",
        "kind": "entity",
        "source": {"sql": "SELECT email, phone FROM sites WHERE domain='webdesign.uk'"},
        "verified_at": "2026-08-04"
      }
    ],
    "banned_claims": [
      {"pattern": "—", "reason": "STYLE, NOT A CLAIM. The owner's explicit instruction 2026-08-04: the em dash reads as automated and unprofessional. Use a full stop, a comma, a colon or brackets. This is banned as a regex because it is the only enforcing lever on this path."},
      {"pattern": "(a (person|human)|someone|we) (then )?(check|checks|reviews|looks over|goes over|casts an eye)", "reason": "The owner's instruction 2026-08-04: this phrasing makes the product sound like a template that gets a brief glance, which undersells the work and is the opposite of the positioning. Describe what is actually done, not that it is inspected."},
      {"pattern": "(before|until) you (ever )?see it", "reason": "Same instruction. The original line 'a person checks it before you ever see it' is the specific sentence the owner asked to remove."},
      {"pattern": "(template|templated|off.the.shelf|cookie.cutter)", "reason": "Do not describe the product this way even to deny it. Denying a frame repeats it."},
      {"pattern": "(trusted|used|loved|chosen) by [0-9]", "reason": "Social-proof class. This is a new service with no customers; any number here is fabricated."},
      {"pattern": "[0-9][0-9,]* ?(clients|customers|businesses|companies|websites|sites) (built|served|helped|delivered|launched)", "reason": "Scale class. No delivery history exists."},
      {"pattern": "years of (experience|expertise)|since [0-9]{4}", "reason": "Tenure class. This service is new and has no trading history."},
      {"pattern": "(award.winning|multi.award|as seen (in|on)|featured in)", "reason": "Reputation class. No awards and no press coverage exist."},
      {"pattern": "(our|the) (team|studio|agency|designers|developers) (of|are|is) [0-9]", "reason": "Headcount class. Do not state or imply a team size."},
      {"pattern": "\"[^\"]{20,}\" ?[—,-]? ?[A-Z][a-z]+ [A-Z]", "reason": "Testimonial shape: a long quotation followed by an attributed name. There are no customers and therefore no testimonials."},
      {"pattern": "(100%|fully) (guaranteed|satisfaction)|money.back guarantee, no questions", "reason": "Overclaims the guarantee. The real term is narrower and specific: a refund is available until the customer accepts the site. State that, not a slogan."},
      {"pattern": "(instant|instantly|in minutes|in seconds|overnight|same day)", "reason": "Speed class. The attested figure is about three or four days; anything faster is unsupported and also fights the positioning."},
      {"pattern": "(seo|search engine) (optimised|optimized|ranking|rankings|guaranteed)", "reason": "No SEO outcome has been measured or promised, and ranking guarantees are unsupportable."},
      {"pattern": "(bespoke|custom|tailor.made) (from scratch|hand.coded|hand.written)", "reason": "Do not claim hand-coding. The sites are framework-built, which is the actual strength; claiming the opposite is both untrue and off-strategy."}
    ],
    "allowed_entities": [
      "webdesign.uk", "Stripe", "the United Kingdom", "UK"
    ],
    "writer_block": "HOW THIS SITE SHOULD SOUND, AND THE THINGS IT MUST NOT DO.\n\nNever use an em dash. Not anywhere, not once. Where you want one, use a full stop, a comma, a colon or brackets. The owner reads em dashes as a tell that copy was machine-written, and on a site selling web design that impression is fatal.\n\nDo not say that a person checks, reviews or looks over the site before the customer sees it. The earlier version of this page said 'a person checks it before you ever see it' and the owner removed it, because it makes the service sound like a template that gets a quick glance. It undersells what is actually done. Write about the work itself rather than about it being inspected.\n\nAsk the visitor a little about themselves and what they want, and only a little. A couple of questions in the page's own voice, or one short set of fields. Not a long form, not a questionnaire, and nothing that has to be completed before they can talk to us. The point is to open a conversation, not to qualify a lead.\n\nSay how long it takes: usually about three or four days from having what is needed. Give it plainly, as a normal fact about how the work goes.\n\nDo not oversell. This service is new and has nothing to boast about yet, so restraint is not modesty here, it is accuracy. No client counts, no testimonials, no awards, no years of experience, no team size, no superlatives about being the best or the fastest. If a sentence needs a boast to work, the sentence is wrong.\n\nWhat you may state as fact is in facts[] above and nothing else: the price, that it is the total with no VAT, the rough duration, the refund-until-acceptance term, that changes after acceptance are charged and our own mistakes are not, that there is no lock-in, and the contact details. Every other number, date, quantity or proportion is off limits. If a claim seems necessary and is not in that list, rewrite around it or leave it out.\n\nRegister: plain, direct British English, written for a business owner who is not technical and does not enjoy being sold to. Short sentences. Ordinary words. No agency vocabulary, no 'solutions', no 'bespoke digital experiences', no 'we're passionate about'. Say the thing, then stop."
  }$eb$::jsonb,
  'manual',
  'Seeded at site creation, before any page was written, so the first page is covered. facts[] is populated (unlike oufe) because the commercial terms are owner-attested and the writer needs them to describe the offer at all.',
  true, true, 'webdesign-uk-build-service-2026-08-04'
FROM sites WHERE domain = 'webdesign.uk';

-- ------------------------------------------------------ imagery_style_guide --
UPDATE site_specs SET is_current = false, superseded_at = now()
WHERE site_id = (SELECT id FROM sites WHERE domain = 'webdesign.uk')
  AND aspect = 'imagery_style_guide' AND is_current;

INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, pinned, created_by)
SELECT
  id,
  'imagery_style_guide',
  $img${
    "medium": "clean, confident editorial illustration: simple geometric composition, generous white space, and calm structural forms suggesting order and craft",
    "mood": "assured, modern, understated; the register of a small studio that is good at its job and does not need to shout about it",
    "palette": "warm off-white paper (#fbfaf8), near-black ink (#16151a), a single confident blue accent (#1a4fd6), cool grey mid-tones (#5a5763, #e2ded7)",
    "avoid": "stock photography of business people, handshakes, laptops on desks, or people pointing at screens; cliched web-design imagery such as wireframe sketches, colour swatches fanned out, pencils and rulers, or browser windows drawn in perspective; code on screens; rocket ships, lightbulbs, jigsaw pieces and other startup cliches; photorealistic depictions of identifiable individuals; text, lettering, numerals, logos or watermarks of any kind",
    "kinds": {
      "content_hero": {
        "medium": "flat two-colour editorial illustration",
        "mood": "one clear abstract motif suggesting structure being composed: aligned blocks settling into a grid, a page taking shape, elements finding their place. Minimal detail, strong silhouette, plenty of air",
        "palette": "warm off-white ground, near-black linework, the blue accent used sparingly on one or two shapes only",
        "avoid": "photorealism, photographic texture, heavy gradients, 3D rendering, drop shadows, text, lettering, numerals, logos, watermarks, busy detail, colour outside the palette, dark or saturated full-bleed backgrounds",
        "reference_asset_keys": []
      }
    },
    "reference_asset_keys": []
  }$img$::jsonb,
  'manual',
  'Seeded pre-build: bugs_closed/027 - content_hero generates unstyled on a site with no style guide, and a fresh site has none. The avoid-list bans web-design cliche imagery explicitly, since a site about web design is the most likely place for it to appear.',
  true, true, 'webdesign-uk-build-service-2026-08-04'
FROM sites WHERE domain = 'webdesign.uk';

COMMIT;

-- Verify
--   SELECT domain, email, phone, status FROM sites WHERE domain='webdesign.uk';
--   SELECT aspect, is_current, pinned, created_by FROM site_specs ss
--     JOIN sites s ON s.id=ss.site_id WHERE s.domain='webdesign.uk' ORDER BY aspect;
--   SELECT jsonb_array_length(data->'banned_claims') AS bans,
--          jsonb_array_length(data->'facts') AS facts
--     FROM site_specs ss JOIN sites s ON s.id=ss.site_id
--    WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current;
