-- ============================================================================
-- remortgagecalculator.uk — CORRECTION to SEED_2026-08-17 (same day, applied)
--
-- WHAT WAS WRONG: every one of the six banned_claims patterns was INERT, and
-- the seed's own verify block passed while they were.
--
-- CAUSE — two escaping layers, counted once. The patterns live in a
-- dollar-quoted ($eb$...$eb$) SQL string, which passes bytes LITERALLY, and are
-- then parsed as JSON, which DOES unescape. Writing `\\\\b` in the file
-- therefore stored `\\b` — a literal backslash followed by 'b' — instead of the
-- word-boundary `\b`. The `£` escapes (`\\u00a3`) went the same way and stored
-- as the six characters `£`.
--
-- WHY IT WAS SILENT, AND WHY THIS IS THE DANGEROUS DIRECTION
-- `datahelpers/claims.go:284-289` compiles each pattern with
-- `regexp.Compile("(?i)" + p)` and falls back to `regexp.QuoteMeta` only when
-- compilation ERRORS. `(?i)\\bguaranteed ...` is a PERFECTLY VALID regex — it
-- just matches a literal backslash — so it compiles, the fallback never fires,
-- and the guard is loaded, listed, counted, and matches nothing for ever.
-- **A pattern that fails to compile is caught; a pattern that compiles WRONG is
-- not.**
--
-- HOW IT WAS CAUGHT, and the lesson worth more than the fix: the original
-- seed's verify block asserted `jsonb_array_length(banned_claims) = 6`. It
-- passed. Counting six inert patterns proves exactly nothing — the count comes
-- out identical whether the guards work or not. What caught it was PROBING:
-- four strings that MUST match and one that MUST NOT. Four came back false.
--
-- Note the probe itself then needed correcting: the first pass used Postgres's
-- `~` operator, but production compiles with Go's RE2, and the two disagree on
-- word boundaries (PG ARE spells it `\y`; `\b` means backspace). A probe in the
-- wrong engine is its own false signal. The authoritative check now lives in
-- Go, beside the production compile call — `datahelpers/claims_banned_pattern_
-- escaping_test.go` — with these exact patterns and probe strings.
--
-- ALSO FIXED HERE: the redundant leading `(?i)` on each pattern. Harmless (Go
-- accepts a repeated flag) but misleading, since claims.go prepends its own.
-- ============================================================================

\set ON_ERROR_STOP on

BEGIN;

UPDATE site_specs SET data = jsonb_set(
    jsonb_set(
      data,
      '{banned_claims}',
      $bc$[
        {"pattern": "\\bguaranteed (acceptance|approval|rate|saving)\\b",
         "reason": "Guarantee language is a financial-promotion exposure and is never true of a remortgage outcome."},
        {"pattern": "\\b(best|cheapest|lowest) (rate|deal)s? (available|on the market|in the uk)\\b",
         "reason": "A market-wide superlative is unverifiable and dates within hours."},
        {"pattern": "\\bsave (up to )?£[0-9,]+",
         "reason": "A specific saving figure is a per-borrower calculation, not a site fact; it must come from the calculator with the user's own inputs, never from copy."},
        {"pattern": "\\b(we are|we're) (fca|financially) (regulated|authorised)\\b",
         "reason": "This site is not a regulated firm and must never imply it is."},
        {"pattern": "\\b(you (will|would) save|this will save you)\\b",
         "reason": "Second-person outcome promises assert a result the site cannot know."},
        {"pattern": "\\b[0-9]+(\\.[0-9]+)?% (apr|apcr|rate)\\b",
         "reason": "A literal rate in copy is a price fact - stale within days and, under a named lender, a financial promotion. Rates belong in the calculator inputs, never in prose."}
      ]$bc$::jsonb
    ),
    '{writer_block}',
    to_jsonb($wb$NUMBERS: this site has NO registered facts yet, so state no rate, percentage, threshold, fee or saving as fact. Do not write a figure you cannot point at. Calculator worked examples are allowed and must be visibly hypothetical ('if your balance were £200,000 ...'), never presented as current market values.
LENDERS: name a lender only via the verified lender directory, and only for NON-PRICE facts - regulator status, product types, underwriter, established year. Never a rate, APR or fee for a named lender, whatever the source.
STANCE: urgency without alarm. The reader's fix is ending within six months; the job is a clear deadline, a clear calculation and a clear next step. No pressure language, no guarantees, no 'act now'.
UNCERTAINTY: 'this depends on your lender' and 'we have not verified this' are always publishable and are preferred to a confident guess.$wb$::text)
  ),
  notes = COALESCE(notes,'') || ' | CORRECTED 2026-08-17b: banned_claims patterns and writer_block were double-escaped and INERT; see SEED_2026-08-17b.',
  updated_at = now()
WHERE site_id = (SELECT id FROM sites WHERE domain = 'remortgagecalculator.uk')
  AND aspect = 'evidence_base' AND is_current;

-- ------------------------------------------------------------------ verify --
-- Structural only. The SEMANTIC check cannot live here: Postgres regex and Go
-- RE2 disagree on \b, so a PG probe would either pass wrongly or fail wrongly.
-- Semantics are pinned by claims_banned_pattern_escaping_test.go, which
-- compiles exactly as claims.go does.
DO $do$
DECLARE
    n int;
    bad int;
    sid uuid;
BEGIN
    SELECT id INTO sid FROM sites WHERE domain = 'remortgagecalculator.uk';

    SELECT jsonb_array_length(data->'banned_claims') INTO n FROM site_specs
    WHERE site_id = sid AND aspect = 'evidence_base' AND is_current;
    IF n <> 6 THEN
        RAISE EXCEPTION 'verify: expected 6 banned_claims, found %', n;
    END IF;

    -- The count above is exactly the check that PASSED while every pattern was
    -- inert, so it is not trusted alone. This one can actually come out false:
    -- no stored pattern may contain a DOUBLE backslash (the corruption's
    -- signature) and none may still carry a literal \u escape.
    -- position()/chr(92), NOT LIKE. In LIKE, backslash is itself the default
    -- ESCAPE character, so a pattern like '%\\%' means "one literal backslash"
    -- and '%\b%' means "the letter b" — which would have made the positive
    -- control below pass on ANY pattern containing 'b'. A guard whose own
    -- quoting has to be reasoned through three layers (file -> SQL literal ->
    -- LIKE escape) is a guard nobody can check. position() has no escape
    -- semantics at all, and chr(92) cannot be rewritten in transit.
    SELECT count(*) INTO bad
    FROM site_specs ss, jsonb_array_elements(ss.data->'banned_claims') e
    WHERE ss.site_id = sid AND ss.aspect = 'evidence_base' AND ss.is_current
      AND (position(chr(92) || chr(92) in (e->>'pattern')) > 0
           OR position(chr(92) || 'u00' in (e->>'pattern')) > 0);
    IF bad > 0 THEN
        RAISE EXCEPTION 'verify: % pattern(s) still double-escaped - they would compile and match nothing', bad;
    END IF;

    -- Positive control on the corrected shape: at least one pattern must now
    -- contain a single-backslash \b, or the UPDATE silently did nothing.
    SELECT count(*) INTO n
    FROM site_specs ss, jsonb_array_elements(ss.data->'banned_claims') e
    WHERE ss.site_id = sid AND ss.aspect = 'evidence_base' AND ss.is_current
      AND position(chr(92) || 'b' in (e->>'pattern')) > 0;
    IF n < 5 THEN
        RAISE EXCEPTION 'verify: only % pattern(s) carry a single-backslash word boundary - expected at least 5', n;
    END IF;

    -- writer_block must carry a real £, not the escape text.
    -- (This assertion was INVERTED on first write — it searched for '£', which
    -- is the DESIRED character, and so refused the correct data. The dry run
    -- caught it. Kept as a comment because an inverted assertion is the same
    -- family as the bug this file fixes: a check that reports on the wrong
    -- condition is worse than no check, and only running it reveals which.)
    -- The escape text is built with chr(92) rather than typed. Typing it is how
    -- this check was wrong TWICE: the authoring channel rewrites a backslash-u
    -- sequence into the character it denotes, so the file ended up searching for
    -- '£' — the very thing it wanted to confirm — and refused correct data.
    -- chr(92) cannot be rewritten. (MEMORY: escape-sequence-emission-trap.)
    SELECT count(*) INTO bad FROM site_specs
    WHERE site_id = sid AND aspect = 'evidence_base' AND is_current
      AND (data->>'writer_block') LIKE '%' || chr(92) || 'u00a3%';
    IF bad > 0 THEN
        RAISE EXCEPTION 'verify: writer_block still carries the literal escape text instead of a pound character';
    END IF;

    -- Positive control for the same field: the £ must actually be present, or
    -- the writer_block replacement silently did nothing.
    SELECT count(*) INTO n FROM site_specs
    WHERE site_id = sid AND aspect = 'evidence_base' AND is_current
      AND (data->>'writer_block') LIKE '%£200,000%';
    IF n <> 1 THEN
        RAISE EXCEPTION 'verify: writer_block does not carry the £200,000 worked example - the replacement did not land';
    END IF;
END $do$;

COMMIT;
