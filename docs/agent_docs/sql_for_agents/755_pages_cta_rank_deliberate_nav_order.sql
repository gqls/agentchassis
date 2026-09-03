-- 755: pages.cta_rank_deliberate_nav_order — "this page winning the site's
-- primary CTA is DELIBERATE, stop flagging it" (bugs_open/436, owner ruling
-- 2026-09-03; bugfix_436_cta_eligibility lane).
--
-- WHAT IT SAYS, and why the estate could not say it before. As of 714 the
-- system can express "never use this page as a CTA destination"
-- (eligible_as_cta_target = false). It cannot express the OPPOSITE judgement:
-- "this page SHOULD win the primary button, I have looked, leave it alone."
-- Those are different statements and only the first existed, so any site whose
-- correct primary button happens to sit on a low nav_order carries a
-- permanent cta_rank_anomaly review item that no action can retire.
--
-- THE WORKED CASE (owner, 2026-09-03). boxingonline.com's rank-1 is
-- tool-fight-calendar at nav_order 3 against a pack at 200 — the exact fossil
-- SHAPE the check fires on, and the RIGHT button for a boxing site. The
-- remedies the check offers both make the site worse: opting the calendar out
-- hands the button to a trivia quiz, and demoting nav_order does the same
-- while also moving the visible menu. The owner's response — "why would
-- boxingonline swap out the calendar for anything, it is prime content?" — is
-- correct, and the absence of a third option is this column's whole reason.
--
-- ⚠ WHY "DISMISS THE WORK ITEM" IS NOT THAT THIRD OPTION, and why the check's
-- own header comment was WRONG about this until today. idx_swi_dedup is
--   UNIQUE (site_id, item_key) WHERE item_key IS NOT NULL
--     AND status <> ALL (ARRAY['complete','verified','rejected','wont_fix',
--                              'failed','unresolved','cancelled'])
-- so the dedup slot is held only by items in OPEN statuses. Closing an item as
-- wont_fix/rejected RELEASES the key and the next discovery pass files an
-- identical one. Measured on cv1.co.uk 2026-09-03: item resolved to `complete`
-- at 10:00:35Z, a fresh identical item inserted at 10:02:24Z. So a dismissal
-- is not durable, and — perversely — leaving the item OPEN is what currently
-- prevents duplicates. This column is the durable form of that dismissal.
--
-- WHY AN integer AND NOT A boolean. The acknowledgement is about a SHAPE, not
-- a page: "reviewed while this page sat at nav_order N". Storing N makes the
-- acknowledgement SELF-EXPIRING — change the number and the check speaks again
-- — which mirrors the check's own item key (cta_rank_anomaly_<page>_<nav>_<site>,
-- keyed on nav_order for exactly this reason). A boolean would silence the
-- check for that page for ever, including for a future shape nobody reviewed,
-- which is the "an acknowledgement outlives what it acknowledged" failure the
-- estate keeps re-learning. NULL = never acknowledged.
--
-- THE UNSAFE SIDE IS OFF BY DEFAULT (2026-08-02 §2). NULL on every existing
-- row = today's behaviour byte for byte; the new authority is "silence a
-- detector", and it must be typed in per page, per nav_order, by a person.
-- Nullable rather than NOT NULL DEFAULT -1 because "no acknowledgement" is
-- genuinely absence, not a sentinel value the readers must decode.
--
-- WHO READS IT (the enumeration RFC_022 requires, as of 2026-09-03 — ONE
-- consumer, and deliberately not the ranking):
--   * discovery check cta_rank_anomaly (check_cta_rank_anomaly.go) — the only
--     reader. It looks the value up for the ALREADY-CHOSEN rank-1 and, when it
--     equals that page's current COALESCE(nav_order,100), takes the positive
--     Resolved branch instead of filing.
--   ⚠ THE RANKING DOES NOT READ THIS, BY DESIGN, and must never start.
--     RankCTAPositionalCandidates already chooses this page; the column records
--     that a HUMAN AGREED, which is a statement about the DETECTOR's finding,
--     not about candidacy. A reader in datahelpers/cta_positional.go would turn
--     a review note into a ranking input and give one page an unearned pin.
--     That is why this column is queried by the check directly and is NOT
--     carried on CTAPositionalCandidate: the shared supply's stated job is
--     "which pages of this type exist and may be linked, nothing more".
--
-- WHAT IT DOES NOT TOUCH: candidacy, ranking order, nav, listings, linkability,
-- the label match, and eligible_as_cta_target (orthogonal — a page could be
-- both opted out and acknowledged, which is contradictory but harmless: opted
-- out wins because it never reaches rank-1, so the acknowledgement is inert).
--
-- ORDERING: safe either side of the carrying image. Older Go never mentions the
-- column; newer Go requires it, so this applies before that image serves.
-- No data change — no page is acknowledged here. Those are owner decisions.

BEGIN;

ALTER TABLE pages
    ADD COLUMN IF NOT EXISTS cta_rank_deliberate_nav_order integer;

COMMENT ON COLUMN pages.cta_rank_deliberate_nav_order IS
    'Set = a human reviewed this page winning the site primary CTA while it sat '
    'at THIS nav_order, and accepted it: cta_rank_anomaly stays silent and '
    'positively retracts (bugs_open/436). Self-expiring — if nav_order changes, '
    'the acknowledgement no longer matches and the check speaks again. NULL = '
    'never acknowledged (default). Read ONLY by the discovery check; the CTA '
    'ranking must never read it.';

-- Verify inside the transaction. DO/RAISE, not bare SELECTs: ON_ERROR_STOP
-- ignores a non-empty result, so a SELECT-only block cannot abort the COMMIT.
DO $$
DECLARE
    col_type text;
    acked    integer;
BEGIN
    SELECT data_type INTO col_type FROM information_schema.columns
     WHERE table_name = 'pages' AND column_name = 'cta_rank_deliberate_nav_order';
    IF col_type IS NULL THEN
        RAISE EXCEPTION '755: column cta_rank_deliberate_nav_order absent after ADD COLUMN';
    END IF;
    IF col_type <> 'integer' THEN
        RAISE EXCEPTION '755: cta_rank_deliberate_nav_order is %, expected integer', col_type;
    END IF;

    -- The whole point of the default: this migration acknowledges nothing.
    SELECT count(*) INTO acked FROM pages WHERE cta_rank_deliberate_nav_order IS NOT NULL;
    IF acked <> 0 THEN
        RAISE EXCEPTION '755: % rows already carry an acknowledgement — this migration must add none', acked;
    END IF;
END $$;

COMMIT;
