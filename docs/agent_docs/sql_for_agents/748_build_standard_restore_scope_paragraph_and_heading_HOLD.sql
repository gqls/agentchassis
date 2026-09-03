-- 748_build_standard_restore_scope_paragraph_and_heading_HOLD.sql
--
-- ⚠ _HOLD — DRAFTED FOR THE OWNER'S APPROVAL OF THE WORDING, NOT APPLIED. Drop the suffix at
-- the moment of applying, and move the canary needle (below) in the SAME commit.
--
-- Two defects in the build-standard carrier row (675), one migration, because both are edits to
-- the same text and the second is free once the first is made:
--
-- (1) THE SCOPE PARAGRAPH. 675's header asserts its wording is "verbatim … confirmed
--     byte-identical … with ONE deliberate trim". [MEASURED 2026-09-03] the source block
--     (049_domain_research_classifier.sql:2593, the `## Build standard` heading to the next `##`)
--     is TWO paragraphs and 675 took one. The dropped ~70 words are the counterweight — "governs
--     QUALITY and FIT, not scope … do not invent services, pages, features, or facts beyond what
--     the evidence supports". The three opted-in rows (build-site-planner/plan_site,
--     content-gap-planner/plan_gaps, visual-designer/design) are exactly the agents that decide
--     what pages and sections EXIST, and they have been receiving the aspiration without its
--     limit. build-site-planner has consumed it twice (10:40Z, 14:15Z on gamedesign.uk), so this
--     is a correction to a live influence, not a pre-emption.
--
--     It cannot be restored verbatim: "say so honestly in the confidence fields" and "adopted
--     sites stay faithful to their source at first" are classifier-specific. And it cannot be
--     restored in its own register either: the source paragraph, and my first generalisation of
--     it, measured 4 negation tells in 54 words against this lane's own detector — two x_not_y,
--     a "rather than", a "do not" — i.e. the exact register the lane exists to remove, about to be
--     injected into every planner prompt. (675's own header trimmed ONE "rather than" from the
--     first paragraph for precisely this reason, then left a not_x_but_y and three em dashes in.)
--
--     WORDING BELOW IS THE OWNER'S, 2026-09-03: a middle ground, since "sometimes the negation
--     makes it easier to understand; it is just that AI overuses it". ONE negation kept, in the
--     first sentence, where the contrast carries the load-bearing distinction (a quality standard
--     is not licence to build more); the other three sentences positive. His two edits on my
--     draft: "real evidence" on the first, weight-bearing use of the word, and "is responsible
--     for" in place of "governs" ("a dense word"). Measured: 1 tell in 57 words. He will revisit
--     the copy another day; this is the approved-for-now form.
--
-- (2) THE HEADING. The opt-in templates insert `## {{.build_standard}}`, which was correct
--     against the source's shape (title · newline · body). 675 rewrote the title into a run-on
--     sentence — "…regardless of inputs). Aim for…" — replacing the line break with a full stop,
--     so `##` spans the whole block. [MEASURED] 897 chars from `##` to the next newline in a live
--     build-site-planner prompt. Restoring the line break makes the same three templates render a
--     heading plus body with no template change.
--
-- ⚠ THE NEEDLE COUPLING. The canary that proves the block RENDERS keys on 675's run-on form —
--   `BUILD STANDARD (applies to every site, regardless of inputs). Aim` — precisely because that
--   mangled form is what distinguishes carrier output from domain-research-classifier's own
--   hard-coded copy of the same block. This migration removes the ". Aim" run-on. The new
--   discriminator is the CAPS title itself: the classifier's copy is sentence-case
--   ("## Build standard (…)"), [MEASURED 2026-09-03] position('BUILD STANDARD (applies' IN its
--   config) = 0. So the needle becomes `BUILD STANDARD (applies to every site` and MUST be
--   matched CASE-SENSITIVELY (position() or LIKE, never ILIKE). Three places carry the old
--   needle and move at apply time, same commit:
--     - copy_quality_two_stage/HANDOFF_2026-09-03b_continue_here.md  (canary bullet)
--     - copy_quality_two_stage/NOTES_two_stage_copy.md               (2026-09-03 entries)
--     - scripts/fire-content-gap-planner.sh                          (the post-fire SELECT)
--
-- Safe to rehearse any time under BEGIN/ROLLBACK. Single-row guarded on config_name (UNIQUE).
-- Versioned: `version` 1 → 2, `updated` stamped, `source` extended rather than replaced so the
-- provenance chain reads 049 → 675 → 748.

BEGIN;

DO $m$
DECLARE
  cur text; nxt text; n int;
  old_title text := 'BUILD STANDARD (applies to every site, regardless of inputs). Aim for best-in-class quality';
  new_title text := E'BUILD STANDARD (applies to every site, regardless of inputs)\nAim for best-in-class quality';
  scope_para text := E'\n\nThis standard is responsible for the quality and fit of what gets built, not for how much. Every service, page, feature and fact in the plan comes from real evidence; thin evidence means a small plan. Aspirations set the direction. The first build carries what the evidence supports today, and the rest arrives as the evidence does.';
BEGIN
  SELECT count(*) INTO n FROM agent_default_configs WHERE config_name='build_standard_block';
  IF n <> 1 THEN RAISE EXCEPTION '748: % carrier rows, want exactly 1', n; END IF;

  SELECT config->>'text' INTO cur FROM agent_default_configs WHERE config_name='build_standard_block';
  IF cur IS NULL OR cur = '' THEN RAISE EXCEPTION '748: carrier text is empty'; END IF;

  -- Idempotency and drift, both directions.
  IF position('This standard is responsible for the quality and fit' IN cur) > 0 THEN
    RAISE EXCEPTION '748: scope paragraph already present — this migration has applied, or 675 was re-cut';
  END IF;
  IF (length(cur) - length(replace(cur, old_title, ''))) / length(old_title) <> 1 THEN
    RAISE EXCEPTION '748: title anchor not-exactly-once — carrier text drifted from 675, re-base';
  END IF;
  IF right(cur, 1) <> '.' THEN
    RAISE EXCEPTION '748: carrier does not end with a full stop — cannot append the paragraph safely';
  END IF;

  nxt := replace(cur, old_title, new_title) || scope_para;

  -- The replacement must not re-embed its own anchor (migration 723''s defect).
  IF position(old_title IN nxt) > 0 THEN RAISE EXCEPTION '748: new text re-embeds the old title anchor'; END IF;

  UPDATE agent_default_configs
     SET config = config
                  || jsonb_build_object('text', nxt)
                  || jsonb_build_object('version', 2)
                  || jsonb_build_object('updated', '2026-09-03')
                  || jsonb_build_object('source',
                       (config->>'source') || '; 748: scope paragraph restored (generalised from 049 — classifier-specific nouns removed) and the title/body line break restored so `## {{.build_standard}}` renders a heading, not an 897-char H2'),
         updated_at = now()
   WHERE config_name='build_standard_block';
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN RAISE EXCEPTION '748: update touched % rows, want 1', n; END IF;

  -- VERIFY on the LIVE row, with RAISE — a verify of SELECTs cannot stop the COMMIT, and a verify
  -- that inspects `nxt` rather than the row confirms intention, not outcome.
  SELECT config->>'text' INTO cur FROM agent_default_configs WHERE config_name='build_standard_block';
  IF position(new_title IN cur) = 0 THEN RAISE EXCEPTION '748 VERIFY: new title (with line break) absent from the LIVE row'; END IF;
  IF position('This standard is responsible for the quality and fit of what gets built, not for how much.' IN cur) = 0 THEN RAISE EXCEPTION '748 VERIFY: scope paragraph absent from the LIVE row'; END IF;
  IF position('comes from real evidence' IN cur) = 0 THEN RAISE EXCEPTION '748 VERIFY: the owner''s "real evidence" edit is absent from the LIVE row'; END IF;
  IF position(old_title IN cur) > 0 THEN RAISE EXCEPTION '748 VERIFY: run-on title survives in the LIVE row'; END IF;
  IF position('confidence fields' IN cur) > 0 OR position('adopted sites' IN cur) > 0 THEN
    RAISE EXCEPTION '748 VERIFY: classifier-specific wording leaked into the carrier';
  END IF;
  IF (SELECT config->>'version' FROM agent_default_configs WHERE config_name='build_standard_block') <> '2' THEN
    RAISE EXCEPTION '748 VERIFY: version not bumped on the LIVE row';
  END IF;
  RAISE NOTICE '748 verify: carrier v2 — heading restored, scope paragraph restored (generalised), classifier nouns absent, % chars.', length(cur);
END $m$;

INSERT INTO schema_migrations (filename, checksum, applied_by, notes)
VALUES ('748_build_standard_restore_scope_paragraph_and_heading_HOLD.sql', :'mig_checksum', 'copy_quality_two_stage session',
        'Carrier 675 dropped the build standard''s scope paragraph and its title/body line break while asserting verbatim; both restored, paragraph generalised. Canary needle moves to the CAPS title in the same commit.');

COMMIT;
