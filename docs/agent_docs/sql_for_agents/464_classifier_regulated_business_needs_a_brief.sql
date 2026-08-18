-- 464_classifier_regulated_business_needs_a_brief.sql
--
-- OWNER INSTRUCTION 2026-08-18: "we should hint to the classifier that this isn't a valid
-- option unless the brief specifically asks for it."
--
-- WHAT HAPPENED `[MEASURED 2026-08-18]`. loanzy.uk was submitted with NOTHING but the domain
-- string — no mission, no contact details, no seed (owner's experiment: let the framework
-- determine the direction, which it is designed to do). `domain-research-classifier` searched
-- the web, found two unrelated companies trading as "Loanzy" abroad, and wrote an identity
-- whose services are "Personal Loan Matching", "Loan Comparison Tool", "Eligibility Checker"
-- and "Lender Lead Facilitation". `domain-strategist` then designed the business around it:
-- money_flow = per-referral fees from lenders, primary_model = lead_generation, panel of
-- FCA-regulated lenders, representative APR on every rate page. The planner emitted 20 pages
-- including tool-eligibility-checker, tool-compare-loans, lenders-index and lender-profile.
--
-- That is CREDIT BROKING — a regulated activity. One page reached the live domain before the
-- build was stopped: "About Loanzy — A Credit Broker, Not a Lender", telling visitors we
-- search a panel of FCA-authorised lenders. Nothing behind it exists.
--
-- ⚠ THE SYSTEM ALREADY NOTICED, WHICH IS WHY THE FIX GOES HERE AND NOT DOWNSTREAM.
-- `build-briefing-agent` recorded in its own gaps list: "FCA authorisation number — not yet
-- known; must be obtained before launch", "Lender panel — specific lenders not confirmed",
-- "Legal entity name — not confirmed". It knew, wrote it down, and had no authority to stop
-- anything, because a gap is a note. The briefing agent is where the system NOTICES; the
-- classifier is where the direction is CHOSEN, and every spec downstream inherits that
-- choice. Turning the briefing agent's detection into a hard gate is the obvious second net
-- and is deliberately NOT in this migration — see the lane's PLAN.
--
-- WHAT THIS CHANGES. Two static insertions into the classifier's prompt. No new template
-- variables (a variable added without its `input_fields` entry renders EMPTY and errors
-- nothing — LANDMINES), no schema change, no new key: the classifier already has everything
-- it needs to obey this, including the Pre-Defined Mission block that licenses the exception.
--
-- SCOPE, stated plainly because this is a SHARED SEAM: this narrows what the classifier may
-- propose for EVERY domain the fleet builds, not just finance ones. It is a narrowing, never
-- a widening — the only behaviour it removes is "invent a regulated business nobody asked
-- for". Registered in the concept register in the same commit (owner ruling 2026-07-29,
-- condition 2), and the other consumers are named there rather than merely measured.
--
-- ROLLBACK: 464_classifier_regulated_business_needs_a_brief_ROLLBACK.sql

BEGIN;

SELECT snapshot_agent('domain-research-classifier',
                      '464_classifier_regulated_business_needs_a_brief.sql: pre-update');

DO $do$
DECLARE
  tpl       text;
  newtpl    text;
  anchor1   text := 'not as things to force into the first build.';
  anchor2   text := 'If no existing site found and no adoption, infer from the domain name and industry norms.';
  rule      text := $rule$

## Regulated business models are NOT available unless the brief asks for one

Some ways of making money are REGULATED ACTIVITIES: carrying one on without authorisation is an offence, and a site that merely PRESENTS as one is making that claim about itself whether or not any business exists behind it. In the UK these include lending and credit broking (introducing or referring a borrower to a lender, "compare and apply" journeys, eligibility checkers that route into lender applications, lead generation for lenders), debt advice, debt adjusting and debt management, mortgage advice or arranging, insurance distribution, investment advice or arranging, payment services and e-money, claims management, and funeral plans. Other jurisdictions have their own equivalents — treat this list as the shape of the problem, not its limit.

Do NOT propose a regulated business model — not in identity.services, not in unique_selling_points, not in the positioning, not anywhere — unless a Pre-Defined Mission is present above and explicitly asks for one. This is not a matter of degree or confidence: absent that instruction, the regulated option is not on the menu at all.

A domain NAME is a topic signal, never a licence. "loan", "credit", "mortgage", "insurance", "invest", "claim" or "pension" in a domain tell you what the site is ABOUT; they do not tell you it may broker, advise, arrange, introduce or refer. Where the name points at a regulated subject and no mission says otherwise, take the UNREGULATED position on that same subject — explaining and teaching, comparing information rather than providers, calculators that compute rather than collect an application, published rules and rights — and note in the confidence fields that a regulated direction was available and deliberately not taken. Honest thinness beats an invented regulated business: a site that only explains something real can be built and stood behind; one that claims a lender panel it does not have cannot.

Facts you can never have for an unbuilt brand: an authorisation or registration number, a regulator relationship, a panel of providers, a legal entity behind the name. Do not assert or imply any of them, and do not treat a same-named company found elsewhere in the world as evidence that this site is that company.
$rule$;
  extra     text := ' Subject to the regulated-business rule above — a name tells you the SUBJECT, never that the site may carry on a regulated activity.';
  n         int;
BEGIN
  SELECT default_config #>> '{workflow,steps,classify_and_extract,config,prompt_template}'
    INTO tpl
    FROM agent_definitions
   WHERE type='domain-research-classifier' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF tpl IS NULL THEN
    RAISE EXCEPTION '464: no live domain-research-classifier prompt_template found';
  END IF;

  -- idempotence: refuse to double-apply
  IF position('Regulated business models are NOT available' in tpl) > 0 THEN
    RAISE EXCEPTION '464: the regulated-business rule is ALREADY present — not re-applying';
  END IF;

  -- anchors must each appear exactly once, or the prompt has moved under us
  n := (length(tpl) - length(replace(tpl, anchor1, ''))) / length(anchor1);
  IF n <> 1 THEN RAISE EXCEPTION '464: anchor1 found % times, expected exactly 1', n; END IF;
  n := (length(tpl) - length(replace(tpl, anchor2, ''))) / length(anchor2);
  IF n <> 1 THEN RAISE EXCEPTION '464: anchor2 found % times, expected exactly 1', n; END IF;

  newtpl := replace(tpl, anchor1, anchor1 || rule);
  newtpl := replace(newtpl, anchor2, anchor2 || extra);

  -- length must have grown by exactly what we inserted: nothing else may have changed
  IF length(newtpl) <> length(tpl) + length(rule) + length(extra) THEN
    RAISE EXCEPTION '464: unexpected length delta (% vs %) — aborting',
      length(newtpl) - length(tpl), length(rule) + length(extra);
  END IF;

  -- no template variable may have been introduced: the rule is static prose
  IF position('{{' in rule) > 0 OR position('{{' in extra) > 0 THEN
    RAISE EXCEPTION '464: inserted text contains a template variable — it would render EMPTY without an input_fields entry';
  END IF;

  UPDATE agent_definitions
     SET default_config = jsonb_set(default_config,
           '{workflow,steps,classify_and_extract,config,prompt_template}',
           to_jsonb(newtpl), false),
         updated_at = now()
   WHERE type='domain-research-classifier' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN RAISE EXCEPTION '464: updated % rows, expected exactly 1', n; END IF;
END
$do$;

-- post-condition, read back from the row rather than from the variable above
DO $verify$
DECLARE t text; BEGIN
  SELECT default_config #>> '{workflow,steps,classify_and_extract,config,prompt_template}'
    INTO t FROM agent_definitions
   WHERE type='domain-research-classifier' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF (length(t) - length(replace(t, 'Regulated business models are NOT available', ''))) /
     length('Regulated business models are NOT available') <> 1 THEN
    RAISE EXCEPTION '464 VERIFY: rule not present exactly once after update';
  END IF;
  IF position('Subject to the regulated-business rule above' in t) = 0 THEN
    RAISE EXCEPTION '464 VERIFY: the infer-from-domain bullet was not amended';
  END IF;
  RAISE NOTICE '464 OK: rule inserted once, bullet amended, template now % chars', length(t);
END $verify$;

COMMIT;
