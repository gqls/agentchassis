-- BRIEF_2026-08-15_site_chat_intake.sql — the DOC-076 experience brief for the
-- "site chat intake" experience-planner run (PLAN_2026-08-11 §3, owner GO
-- 2026-08-15). A REAL brief, not a probe: no test-artefact marking.
--
-- Kept OUT of docs/agent_docs/sql_for_agents/ so the migration runner can
-- never sweep it up: it is a one-shot doc_notes seed, not a change to the
-- estate (same reasoning as loancalculator_couk/probe_363_veto_arm_brief.sql,
-- whose INSERT shape this mirrors).
--
-- The body is what the planner's `compose` step is handed (load_brief selects
-- by subject_key + categories @> ["experience-brief"]). It distils decisions
-- the owner has already taken — with the deliberate exception that it states
-- MECHANISMS, not site facts: a figure written here would be a second source
-- of truth beside evidence_base, which is exactly the drift class this lane
-- was built to close.

INSERT INTO doc_notes (subject_type, subject_key, site_id, body, categories, source, source_agent, created_by)
SELECT 'experience',
       'site-chat-intake',
       s.id,
$brief$# Brief — site chat intake (an on-site chat box that answers from live site facts)

## What this is and why (owner-approved plan 2026-08-11; GO given 2026-08-15)

webdesign.uk carries a chat box on its contact page that answers visitor
questions about price, process and timing, and invites the visitor to start
their brief. It is live and working. This plan's job is to state, once, what a
safe and correct site chat experience IS — journeys, promises, data contract,
MVP cut — so that giving any other framework site a chat box becomes a
per-site decision that cites an approved contract, instead of re-deriving
safety from first principles each time.

## Decisions already taken — owner-accepted, do NOT relitigate

1. **Answers come from the live facts relay, never from the bot's own code and
   never from this brief.** The chat service fetches the site's facts
   (`site_specs.evidence_base`, served read-only over HTTPS by the in-cluster
   facts relay) at startup and refreshes every few minutes. A price or term
   stated anywhere else — compiled into the binary, or written into a brief —
   is a drift defect; this site has already been bitten once by a retired
   figure surviving in code. The plan must state the mechanism and must not
   restate the site's current facts.
2. **Fail-closed to real contact details.** If facts cannot be fetched at
   startup, the widget does not chat: it shows the site's real contact
   details instead. A running bot that loses the relay keeps serving
   last-good facts and logs the failure. This behaviour is built and proven
   in production; the plan states it as the contract.
3. **Four abuse/spend controls are built and mutation-tested in Go**, and the
   plan states them as promises rather than leaving them implicit in one Go
   file: a per-IP rate limit; a per-conversation turn cap; a daily spend
   ceiling; and the fail-closed fallback above. The chat backend runs
   off-cluster on the site's own box deliberately — an anonymous,
   internet-triggered, token-spending pipeline must not run on core. Do not
   propose moving it.
4. **Which sites may have one is already gated.** The widget is a library
   tool (`chat-input-box`) tagged `requires-backend`, and tool-suggester
   offers such tools only to sites whose `deploy_config.capabilities`
   includes `backend` (live). The plan cites this gate; it does not invent a
   second eligibility mechanism.
5. **The chat converts to a LEAD** — a visitor leaves an email address or
   starts their site brief. Payment is a separate checkout flow and is out of
   chat scope.

## Journeys wanted

- A visitor asks the price → gets the real current answer, stated plainly,
  from live facts → is invited to leave an email / start their brief.
- A visitor asks about process or timing → answered from live facts only;
  where the facts are silent the bot says so and offers the contact route
  rather than improvising an answer.
- An off-topic, hostile or runaway conversation → hits the turn cap or rate
  limit and is politely handed the real contact details.
- The relay or the model is down → the box degrades to real contact details;
  no dead ends, and no fabricated answers, ever.

## Data contract — what one deployment needs, per site

- The site's identity to the relay (site id / domain), the relay URL, and the
  relay bearer token (infrastructure-owned, rotated as a pair with the
  cluster side).
- The contact fallback details — themselves facts in the site's
  evidence_base, not deployment literals.
- Tone/brand parameters, and the per-site values of the four controls.

A deployment that needs a fact evidence_base does not hold is a gap to file
against the register, never something to hardcode into the service.

## MVP cut

One site (webdesign.uk) is already live end to end. The MVP for "reusable" is
the SECOND site on the same box: the same binary, different parameters, its
facts read from its own evidence_base through the same relay. A multi-tenant
chat relay service is a later, owner-level decision and is not this plan.

## What "done" looks like

An approved EXPERIENCE_PLAN whose journeys, promise ledger and data contract
match the above, cited by every future per-site chat decision. The honesty
bar: nothing in the plan may let a bot state a fact its site's evidence_base
does not currently attest.
$brief$,
       '["experience-brief"]'::jsonb,
       'webdesign_uk_build_service · PLAN_2026-08-11 §3 (owner GO 2026-08-15)',
       'cli',
       'webdesign_uk_build_service'
FROM sites s WHERE s.domain = 'webdesign.uk';
