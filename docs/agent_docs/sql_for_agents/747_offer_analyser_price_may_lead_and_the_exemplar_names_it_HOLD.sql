-- 747_offer_analyser_price_may_lead_and_the_exemplar_names_it.sql
--
-- OWNER RULING 2026-09-03, verbatim (relayed by the copy_quality_two_stage lane from the
-- owner's own words; his message was to them, not to this lane's log):
--
--   "we can add price in the exemplar, we don't need to be so strict as to not enable us
--    to say what we need to. It's not 'never a description of our inventory' its a
--    deprioritise and mostly leave it very brief or out altogether. Currently there is
--    nothing wrong with this sentence that I can see: 'why would I pay £29 when I can get
--    AI to analyse my idea for free?' (if we can't offer anything better than the free
--    model then we should think of another tool to create.)"
--
-- TWO edits, on DISJOINT anchors, both in `offer-analyser`'s ordering prompt. They must
-- ship together or the prompt argues with itself: naming price in the exemplar while the
-- absolute four sentences earlier forbids "a description of our inventory" would tell the
-- model to raise a doubt it is simultaneously barred from answering.
--
-- ══ WHY THIS IS THE FIX, MEASURED — AND WHY THE TWO OBVIOUS ALTERNATIVES ARE NOT ═══════
-- The lane reached this ruling from a finding it then had to withdraw TWICE. Both wrong
-- accounts are recorded because each one would have produced a different, wrong migration.
--
--  ✗ FIRST ACCOUNT — "our sites do not address price" (WITHDRAWN). It was a property of
--    this prompt, not of the sites: `money_flow`, `price`, `cost`, `pay`, `charge` and `£`
--    occur ZERO times in all 9,590 characters (9,642 bytes) of it. That fix would have been an editorial
--    campaign across 18 sites.
--  ✗ SECOND ACCOUNT — "the question side reads the whole strategy register, the answer
--    side reads only TASK 1's four fields, so a money_flow question is unanswerable BY
--    CONSTRUCTION" (ALSO WITHDRAWN, and this is the one that matters here, because the
--    obvious migration it implies is "add money_flow to TASK 1's four" and that would be
--    the WRONG EDIT). [MEASURED 2026-09-03] `from_field` on lead_with points is OPEN and
--    the model already uses it outside the four: competitive_position 17, content_strategy
--    13, **money_flow 5 across 4 sites**, revenue_models 2, growth_path 1. And money_flow
--    questions are not unanswerable — 2 of 7 ARE answered. So the field list is a strong
--    PRIOR, not a wall, and widening it is not the constraint.
--
--  ✓ THE ACTUAL MECHANISM, and it is visible in the two sites where both halves exist.
--    [MEASURED 2026-09-03] answered-rate by the question's own field: trust_threshold
--    29/30 · satisfaction_condition 24/24 · competitive_position 9/9 · recurring_value
--    15/16 · value_proposition 9/10 · **money_flow 2 of 7**. And uniquely, **ZERO
--    money_flow questions are answered by a money_flow point**, though five such points
--    exist. Exactly two sites carry both a money_flow point AND a money_flow question, and
--    on both the question is unanswered — so I read them:
--
--      finetuning.uk  Q: "How much is this going to cost me, and how long before I see
--                         anything working?"
--                     P: "The discovery call is a conversation about your specific process,
--                         with no sales pitch and no commitment to spend anything."
--      idea.uk        Q: "Why would I pay £29 for a report when I can get AI to analyse my
--                         idea for free?"
--                     P: "a verified report covers the market, competition, what is
--                         defensible, and a sensible next step — researched and sourced."
--
--    **THE MODEL WAS RIGHT TO REFUSE BOTH JOINS.** The first point answers "will I be
--    pressured to spend", the second lists contents; neither answers its question, and
--    marking either answered would be precisely the stretch the join rule forbids. Note
--    what both points share: **they cite `money_flow` and NEITHER STATES A PRICE.** That is
--    the absolute doing it — "never a description of us or of our inventory" makes
--    "£29, one-off, no subscription" unwritable, so a point sourced from money_flow must
--    launder the price into something that is not a price.
--
--    **So repealing the absolute is what lets a point STATE the price, and only a point
--    that states the price can answer "why would I pay £29 when the alternative is free".**
--    The exemplar change makes the question get asked; the absolute repeal makes it
--    answerable. The second is what makes the first do anything.
--
-- ⚠ DELIBERATELY NOT DONE: adding `money_flow` to TASK 1's four named fields. The
--   measurement above says the field list is not the binding constraint, and the owner did
--   not rule it. Doing it anyway would be scope the evidence does not support.
--
-- ══ HOW EDIT A IS WORDED, AND WHY NOT MORE LOOSELY ═══════════════════════════════════
-- "Deprioritise" is vaguer than "never", and vague rules drift. The absolute was doing real
-- work: it stops a page leading with "we have 500 templates" instead of "find the one you
-- need in 30 seconds". His intent is *let us say the necessary thing, briefly, low down* —
-- not *inventory description is now fine*. So the replacement keeps three teeth: it ranks
-- such a point LAST by default, caps it to ONE CLAUSE, and states the reader-benefit form
-- is preferred wherever one exists. Same house position he ruled into the page writer the
-- same morning ("say less or leave it out", migration 739), and worded to match.
-- `avoid_leading_with`'s own guidance is untouched and stays consistent: it names
-- "any self-description carrying NO READER BENEFIT", which the softened rule does not
-- license.
--
-- ⚠ IDEMPOTENCE — the trap this lane shipped six days ago. Migration `723` edited THIS SAME
-- PROMPT with a `replace()` whose replacement re-embedded its own anchor, so a second run
-- stacks a second copy. Both edits here are guarded NOT-EXACTLY-ONCE with a RAISE, on both
-- the pre-state AND the post-state, and NEITHER replacement contains its own anchor
-- (asserted mechanically below, not by eye). Anchors are 4,927 chars apart and each occurs
-- exactly once.
--
-- ══ TWO NOTES FROM THE PEER REVIEW OF THIS FILE, BOTH ACTED ON ═══════════════════════
-- `copy_quality_two_stage` read this migration on disk before it shipped (live prompt
-- confirmed untouched at the time). Two points, neither blocking, both taken:
--
-- 1. **"rank it LAST" was an INSTRUCTION where the owner said "mostly".** His words were
--    *"a deprioritise and mostly leave it very brief or out altogether"*, and "mostly" is
--    doing work. As first written, a site whose actual differentiator IS price — a
--    deliberately cheap alternative — would have had its strongest reader benefit forced to
--    the bottom. **Changed to "rank it LAST BY DEFAULT — unless it is genuinely this site's
--    strongest reader benefit, which is rare".** The escape hatch is narrow on purpose:
--    "which is rare" is what stops it becoming the ordinary case.
--
-- 2. **`Say less or leave it out.` now appears in TWO live prompts** — here, and the
--    page-content-writer's rule 18 via migration `739` (applied 2026-09-03 12:34Z). **That
--    is deliberate, not accident:** it is the owner's house phrasing and both prompts should
--    carry it identically. Recorded because a future session grepping that string will find
--    two owners and needs to know neither is a stray copy of the other.
--
-- They also retracted half of their own "necessary but not sufficient" objection on reading
-- the file: the *"prefer the reader-benefit form wherever one exists"* clause already pushes
-- against the bare-price outcome they predicted (the reader-benefit form of "£29" is "£29 for
-- a report that says what is defensible, sourced", not the naked figure). What survives is
-- narrower and is the pre-registered check below.
--
-- ⚠ THE PROMPT'S LENGTH, because two figures were in circulation and BOTH were being quoted
-- as "chars": it is **9,590 CHARACTERS and 9,642 BYTES**. The 52-byte difference is exactly
-- 26 em dashes at two extra bytes each; `wc -c` reads 9,643 because psql adds a trailing
-- newline. This is the estate's standing bytes-vs-chars trap in miniature — measure with
-- `awk length()` or python `len()` and SAY WHICH, because a length quoted without its unit
-- cannot be reconciled by the next reader.
--
-- ══ COUNCIL ROUND 1: APPROVED, 3 advisory mediums — ALL ANSWERED BY MEASUREMENT ═══════
-- Corr `aeaf9f88-4348-4453-8c9e-213e7fd548a7`, 2026-09-03 16:01:45Z, "approved with 3
-- advisory objection(s) — none high-severity", 8 abstained.
--
-- ⚠ **THE APPROVAL IS FOR THE PREVIOUS WORDING AND IS NOT BEING CLAIMED.** r1 reviewed
-- "rank it LAST"; this file now says "rank it LAST BY DEFAULT — unless…", which is a
-- semantic softening, not a typo. So this goes back as r2 on the SAME correlation and the
-- commits carry `Council-Submitted:`, never `Council-Reviewed:`, until a verdict approves
-- the text that actually exists. **An APPROVED verdict is the DANGEROUS one here**: a REVISE
-- forces the resubmission anyway, while an approval creates both the temptation and the
-- paper trail for "approved, ship it" over wording the file no longer contains.
--
-- 1. `editquality` + `debug_historian` (medium, same point, both right): **the row locator
--    had no multi-row guard.** PL/pgSQL's non-STRICT `SELECT … INTO` takes ONE ARBITRARY ROW
--    with no error, and a landmine records that some agent types carry TWO active
--    `agent_definitions` rows of which only the higher `version` loads live — so this could
--    have edited the stale row while the verify, re-running the same unordered predicate,
--    read back the row that WAS written and reported success.
--    [MEASURED 2026-09-03] **`offer-analyser` has exactly ONE row (version 1)**; the four
--    types that do carry duplicates are `chief-strategist`, `content-creator`,
--    `content-creator-contact`, `site-component-architect`. And it carries exactly ONE step
--    with a prompt (`run_offer_analysis`, 9,590 chars — the DB's own count, agreeing with
--    the character figure above). **So the objection's premise is false today, and the guard
--    goes in anyway**, on both the locator and the verify: a migration that is safe only
--    because of a fact nobody re-checks is precisely what this council exists to stop.
--    Proven by induced failure — widening the count predicate gives
--    `ABORT: 9 live offer-analyser (row, step) pairs carry anchor A…`, an abort rather than
--    an arbitrary pick.
--
-- 2. `guardian` (medium): **does 747 leave migration `723` rollbackable?** 723 edited this
--    same prompt row, and a later migration rewriting sentences a prior one touched can leave
--    the prior rollback unable to find its own anchor — without anyone deciding that.
--    **ANSWERED BY RUNNING IT, not by inspection.** Applying 747 inside a transaction and
--    probing before rollback: 723's ROLLBACK anchor
--    (`OUTPUT. Return ONE JSON object and nothing else`) still occurs **exactly once** in the
--    post-747 prompt, and both blocks 723 inserted (`THE QUESTION HIERARCHY`, `THE JOIN IS
--    THE POINT`) are intact. **723 stays rollbackable.** The two edit regions are disjoint:
--    747's anchors sit at offsets 2407 and 7334 in TASK 1 and the hierarchy exemplar, neither
--    inside 723's inserted block.
--
-- 3. ⚠ **AND THE FILE IS NOW `_HOLD`, WHICH IS A DEFECT ROUND 1 DID NOT CATCH AND A PEER DID.**
--    An unapplied migration sitting in `sql_for_agents/` is not "held" — any session running
--    `run-migrations.sh --apply` without scoping `MIGRATIONS_DIR` takes EVERY pending file, so
--    this could have been applied by a lane that had never heard of it, in wording that
--    diverges from what was reviewed. Verified in the runner rather than taken on trust:
--    `run-migrations.sh:65` sets `SIDECAR_RE='_[A-Z][A-Z0-9_]*\.sql$'`, line 284 filters the
--    pending list with `grep -vE "$SIDECAR_RE"` and line 294 still LISTS it. **Drop the
--    `_HOLD` suffix at the moment of applying, not before.**
--
-- ══ PRE-REGISTERED POST-APPLY CHECK — and it can come out AGAINST this migration ═══════
-- Raised by `copy_quality_two_stage` before apply, and recorded here rather than argued,
-- because it is cheap and decidable. Their point: **stating the price is NECESSARY but may
-- not be SUFFICIENT.** Decompose idea.uk's question — "why would I pay £29 for a report
-- when I can get AI to analyse my idea for free?" — into what an answer must do:
--   (a) acknowledge the cost exists          — needs the price statable. THIS FIXES IT.
--   (b) name what the cost buys              — the existing point ALREADY partly does this
--                                              ("verified… researched and sourced" is
--                                              exactly what free AI is not).
--   (c) contrast it with the NAMED free alternative — the point does NOT do this, and the
--                                              absolute was never what stopped it.
-- So a model freed to write "£29, one-off, no subscription" may state the price and STILL
-- not answer why you would pay it — which would pass a naive "do money_flow points state
-- prices now?" check while the question stayed unanswered, and read as partial success.
--
-- THE CHECK, to be run after enough sites have re-analysed (it is the same query that
-- produced the numbers above, so no new instrument):
--   answered-rate by the question's OWN from_field, for money_flow, against the pre-apply
--   baseline of **2 of 7** and the other fields' 24/24, 29/30, 9/9, 15/16, 9/10.
--   * moves toward the others  -> this migration's account is complete.
--   * money_flow points start STATING prices while the questions stay unanswered
--     -> the binding constraint was only HALF the absolute, and the other half is that
--        nothing in the prompt asks a point to engage the ALTERNATIVE the visitor names.
--        That is a one-clause addition of a different kind, and a separate migration.
-- ⚠ NOT folded in here on purpose: the owner's ruling is clear and right on its own terms,
-- and adding a second speculative clause is how a reviewable change becomes unreviewable.
--
-- ⚠ AND THE MECHANISM ABOVE IS STRONG-BUT-THIN, STATED AS SUCH. "Zero of seven, uniquely,
-- while five such points exist" is a very clean number, and cleanness is exactly the tell
-- this lane got wrong twice today. It was not quoted alone — the two co-occurring pairs
-- were hand-read — but that is **n=2 carrying the mechanism**. The post-apply re-run is
-- what converts it from strong-but-thin to settled.
--
-- ⚠ AND ONE THING THIS MIGRATION CANNOT DO, from the owner's own aside: *"if we can't
-- offer anything better than the free model then we should think of another tool to
-- create."* A doubt we cannot answer is information about the OFFER, not a copy defect to
-- close. If idea.uk's price question is still unanswered after this, the next move may be a
-- PRODUCT change, and that belongs to idea.uk's lane, not to this prompt.
--
-- ══ ⚠ AND THE VERIFY CAUGHT ME, ON THIS FILE, AFTER THE PEER REVIEW ══════════════════
-- Worth recording because it is the difference between a control and decoration, and it
-- happened for real rather than in a mutant. Acting on note 1 above I changed `repl_a`'s
-- wording ("rank it LAST" -> "rank it LAST BY DEFAULT — unless…") **and did not update the
-- verify block's needle**, which still searched for the old phrase. The next clean dry run
-- failed with `ABORT: the softened rule is not in the LIVE row — the write did not land`.
-- The UPDATE had run and the NOTICE had printed; only the independent live-row read caught
-- that what landed was not what the verify was written to expect. **A verify that re-reads
-- the LIVE row catches an author's own half-finished edit; one that inspects the block's own
-- in-memory copy cannot.** Needle re-pointed, re-run green.
--
-- ══ GUARDS PROVEN, NOT ASSERTED ═══════════════════════════════════════════════════════
-- [MEASURED 2026-09-03] dry run against live data (COMMIT->ROLLBACK): both edits land, the
-- independent verify passes reading the LIVE row, prompt 9,590 -> 9,987 characters, step
-- `run_offer_analysis`. THREE induced failures prove the guards are controls, not
-- decoration:
--   * replacement A made to re-embed its own anchor -> "ABORT: replacement A re-embeds
--     anchor A — this is migration 723's defect"   (723's exact defect, caught)
--   * ANCHOR A mistyped so it matches nothing       -> "ABORT: no live offer-analyser step
--     carries anchor A — the prompt has moved"     (asked for by the peer review, because A
--     is the anchor carrying the semantic change)
--   * ANCHOR A set to a string occurring MANY times -> "ABORT: anchor A occurs 83 times,
--     expected exactly 1"                          (the count guard in the other direction —
--     a presence check would have passed here, which is exactly 723's defect)
--   * anchor B mistyped so it matches nothing       -> "ABORT: anchor B occurs 0 times"
--   * the verify's needle changed to one no write produces -> "ABORT: the softened rule is
--     not in the LIVE row — the write did not land"
--   * the locator's count predicate widened to match many pairs -> "ABORT: 9 live
--     offer-analyser (row, step) pairs carry anchor A…"  (an ABORT, not an arbitrary pick)
--
-- Reversible: 747_..._ROLLBACK.sql restores both sentences verbatim.
-- Source: this lane's NOTES 2026-09-03 "I MEASURED MY OWN PROMPT"; owner ruling relayed
-- 2026-09-03 via copy_quality_two_stage.

BEGIN;

DO $mig$
DECLARE
    anchor_a  text := 'a benefit to the reader, never a description of us or of our inventory.';
    repl_a    text := 'a benefit to the reader. A point that describes us or our inventory — what we stock, how many, what it costs — is allowed only where the reader genuinely needs it: rank it LAST BY DEFAULT — unless it is genuinely this site''s strongest reader benefit, which is rare — keep it to a single clause, and prefer the reader-benefit form wherever one exists ("find the one you need in 30 seconds", not "we have 500 templates"). Say less or leave it out.';
    anchor_b  text := 'For most sites that is some form of what will this actually get me and how much work is it to get it;';
    repl_b    text := 'For most sites that is some form of what will this actually get me, how much work is it to get it, and what does it cost me;';
    p         text;
    n         int;
    agent_id  uuid;
    step_key  text;
BEGIN
    -- ⚠ COUNT THE MATCHES BEFORE TAKING ONE (council r1, `editquality` + `debug_historian`,
    -- both medium, same point, both right). PL/pgSQL's non-STRICT `SELECT … INTO` takes ONE
    -- ARBITRARY ROW on a multi-row result — no error, no warning — and a documented landmine
    -- says some agent types carry TWO active `agent_definitions` rows of which only the
    -- higher `version` is ever loaded live. Without this guard the migration could edit the
    -- stale row while the live one keeps the old absolute, and the verify block below,
    -- re-running the same unordered predicate, could read back the row that WAS written and
    -- report success. That is the "a check sharing state with the thing it checks confirms
    -- the author's intention, not the outcome" shape, one table along.
    --
    -- [MEASURED 2026-09-03] `offer-analyser` has exactly ONE row (version 1). The four types
    -- that DO carry duplicates are chief-strategist, content-creator, content-creator-contact
    -- and site-component-architect — offer-analyser is not among them. **The guard goes in
    -- anyway**: the objection's premise is false today and the class is real, and a migration
    -- that is safe only because of a fact nobody re-checks is the thing this council exists
    -- to stop. It counts (agent, step) PAIRS, so two matching STEPS on one agent aborts too.
    SELECT count(*) INTO n
      FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config->'workflow'->'steps') st
     WHERE ad.type = 'offer-analyser' AND ad.is_active
       AND COALESCE(ad.is_snapshot,false) = false AND ad.deleted_at IS NULL
       AND st.value->'config'->>'prompt' LIKE '%' || anchor_a || '%';
    IF n = 0 THEN
        RAISE EXCEPTION 'ABORT: no live offer-analyser step carries anchor A — the prompt has moved';
    END IF;
    IF n > 1 THEN
        RAISE EXCEPTION 'ABORT: % live offer-analyser (row, step) pairs carry anchor A. SELECT INTO '
                        'would take an arbitrary one and the verify could read back the same wrong '
                        'row. Decide which is live (highest version) before applying.', n;
    END IF;

    SELECT ad.id, st.key, st.value->'config'->>'prompt'
      INTO agent_id, step_key, p
      FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config->'workflow'->'steps') st
     WHERE ad.type = 'offer-analyser' AND ad.is_active
       AND COALESCE(ad.is_snapshot,false) = false AND ad.deleted_at IS NULL
       AND st.value->'config'->>'prompt' LIKE '%' || anchor_a || '%';

    IF p IS NULL THEN
        RAISE EXCEPTION 'ABORT: locator returned no prompt after the count said 1 — race or NULL prompt';
    END IF;

    -- PRE-STATE, both anchors, EXACTLY ONCE. A count, not a presence check: `723`'s defect
    -- was that `position(...) = 0` cannot tell one occurrence from two.
    n := (length(p) - length(replace(p, anchor_a, ''))) / length(anchor_a);
    IF n <> 1 THEN RAISE EXCEPTION 'ABORT: anchor A occurs % times, expected exactly 1', n; END IF;
    n := (length(p) - length(replace(p, anchor_b, ''))) / length(anchor_b);
    IF n <> 1 THEN RAISE EXCEPTION 'ABORT: anchor B occurs % times, expected exactly 1', n; END IF;

    -- Neither replacement may contain its own anchor, or a re-run stacks a copy (723's
    -- exact defect). Asserted mechanically rather than trusted to the eye.
    IF position(anchor_a in repl_a) > 0 THEN
        RAISE EXCEPTION 'ABORT: replacement A re-embeds anchor A — this is migration 723''s defect';
    END IF;
    IF position(anchor_b in repl_b) > 0 THEN
        RAISE EXCEPTION 'ABORT: replacement B re-embeds anchor B — this is migration 723''s defect';
    END IF;

    -- Already applied? Abort rather than double-apply.
    IF position(repl_a in p) > 0 OR position(repl_b in p) > 0 THEN
        RAISE EXCEPTION 'ABORT: a replacement is already present — this migration has applied';
    END IF;

    p := replace(p, anchor_a, repl_a);
    p := replace(p, anchor_b, repl_b);

    -- POST-STATE, before the write: each replacement exactly once, each anchor gone.
    n := (length(p) - length(replace(p, repl_a, ''))) / length(repl_a);
    IF n <> 1 THEN RAISE EXCEPTION 'ABORT: replacement A landed % times, expected 1', n; END IF;
    n := (length(p) - length(replace(p, repl_b, ''))) / length(repl_b);
    IF n <> 1 THEN RAISE EXCEPTION 'ABORT: replacement B landed % times, expected 1', n; END IF;
    IF position(anchor_a in p) > 0 THEN RAISE EXCEPTION 'ABORT: anchor A survives'; END IF;
    IF position(anchor_b in p) > 0 THEN RAISE EXCEPTION 'ABORT: anchor B survives'; END IF;

    UPDATE agent_definitions
       SET default_config = jsonb_set(default_config,
             ARRAY['workflow','steps',step_key,'config','prompt'], to_jsonb(p), false),
           updated_at = now()
     WHERE id = agent_id;

    RAISE NOTICE '747: offer-analyser ordering prompt — price may now LEAD (ranked last BY DEFAULT, one clause), '
                 'and the exemplar names cost. Step %, prompt now % chars.', step_key, length(p);
END $mig$;

-- INDEPENDENT VERIFY, re-reading the LIVE row rather than the local variable — a block that
-- checks its own in-memory copy proves only that string functions work.
DO $$
DECLARE p text; n int;
BEGIN
    -- Same multi-row discipline as the locator: a verify that could read back an arbitrary
    -- one of two rows proves nothing about which row is live.
    SELECT count(*) INTO n
      FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config->'workflow'->'steps') st
     WHERE ad.type='offer-analyser' AND ad.is_active
       AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
       AND st.value->'config'->>'prompt' LIKE '%rank it LAST BY DEFAULT%';
    IF n <> 1 THEN
        RAISE EXCEPTION 'ABORT: % live (row, step) pairs carry the softened rule, expected exactly 1', n;
    END IF;

    SELECT st.value->'config'->>'prompt' INTO p
      FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config->'workflow'->'steps') st
     WHERE ad.type='offer-analyser' AND ad.is_active
       AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
       AND st.value->'config'->>'prompt' LIKE '%rank it LAST BY DEFAULT%';

    IF p IS NULL THEN
        RAISE EXCEPTION 'ABORT: the softened rule is not in the LIVE row — the write did not land';
    END IF;
    IF position('what does it cost me' in p) = 0 THEN
        RAISE EXCEPTION 'ABORT: the exemplar does not name cost in the LIVE row';
    END IF;
    IF position('never a description of us or of our inventory' in p) > 0 THEN
        RAISE EXCEPTION 'ABORT: the repealed absolute is still in the LIVE row';
    END IF;
    n := (length(p) - length(replace(p, 'Say less or leave it out.', ''))) / length('Say less or leave it out.');
    IF n <> 1 THEN RAISE EXCEPTION 'ABORT: the new clause appears % times in the LIVE row, expected 1', n; END IF;
    RAISE NOTICE '747 VERIFY: live row carries both edits, exactly once each, and the absolute is gone';
END $$;

COMMIT;
