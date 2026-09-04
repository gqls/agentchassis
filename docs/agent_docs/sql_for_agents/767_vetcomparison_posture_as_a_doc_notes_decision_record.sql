-- 767_vetcomparison_posture_as_a_doc_notes_decision_record.sql
--
-- Answers the council REVISE on migration 761 (corr 5d54f835, gated by prior_art_librarian HIGH,
-- with converging objections from editquality, guardian and reuse_agent). Two of the three
-- substantive objections are answered by MEASUREMENT below; this migration is the fix for the
-- third.
--
-- ─── OBJECTION 1 (HIGH, gating): "the map round-trip claim contradicts the landmines" ────────
--
-- prior_art_librarian, editquality and guardian all converged on the same gap, and they were
-- RIGHT that I had not closed it: 761's safety case rests on unknown keys surviving the daily
-- writer, and I verified that at ONE writer (`refresh_evidence_base_action.go`) while the
-- landmine register carries entries warning that parsing `evidence_base` through its own TYPED
-- struct and writing it back DELETES every `citation`, `writer_line` and unknown field. I never
-- established whether such a write path is LIVE. That is the "editing one file is not knowing
-- the package" class, and it matters far beyond this key: if a typed writer were live, every
-- register on the fleet would be losing its citations.
--
-- **CENSUSED 2026-09-04. THE WRITER SET IS THREE, AND ALL THREE ARE MAP-BASED.** Derived, not
-- recalled: every .go file mentioning `evidence_base` intersected with those writing `site_specs`
-- (`grep -rlE "INSERT INTO site_specs|UPDATE site_specs|writeSiteSpec"`).
--
--   1. `refresh_evidence_base_action.go:1695` — `writeRefreshedEvidenceBase(… eb map[string]interface{} …)`,
--      `json.Marshal(eb)` at :1698. MAP. Unknown keys survive.
--   2. `evidence_citations.go:449`  — `writeCitationRegister(… eb map[string]interface{} …)`,
--      `json.Marshal(eb)` at :453. MAP. And its facts slice is read straight back out of the same
--      map at :320 (`facts, _ := eb["facts"].([]interface{})`) with new facts APPENDED, so
--      per-fact unknown keys (`citation`, `writer_line`, `rule`, `draft_status`,
--      `corrects_site_citation`, `no_citation_because`) survive too.
--   3. `internal/core-manager/admin/site_admin_handlers.go:283` — writes `body.Data`, the caller's
--      RAW JSON. It calls `ParseEvidenceBase` at :225, but ONLY as a validation guard (to refuse a
--      save that would parse to nothing scannable). The parsed struct is never marshalled back.
--
-- **So the typed struct is a READ and VALIDATION path only; it is not a write path today.** The
-- landmines describe a real HAZARD SHAPE with no live instance — which is why they are worded as
-- a warning about what a future writer would do. `[MEASURED 2026-09-04]`, and this is the census
-- RUNBOOK §8d asked for and only half-supplied. **The date is load-bearing: this is a claim about
-- a SET OF CALLERS, and such a set grows by addition while reading as current for ever. Re-run
-- the grep before quoting this; `git log --since=2026-09-04 --diff-filter=A -- platform/orchestration/actions/`
-- lists what has been added since.**
--
-- ⚠ The corollary, and it is the useful half: a future writer that round-trips this aspect through
-- `EvidenceBase` would silently strip `citation` from **every** register on the fleet, not just
-- this key. The census makes that a testable invariant rather than folklore.
--
-- ─── OBJECTION 2 (MEDIUM, reuse_agent): "doc_notes already IS the who/when/why mechanism" ────
--
-- Also right, and also not checked: 761's rationale grepped Go and SQL for `posture_rung` /
-- `claims_posture` and concluded the Q4 record has no home, without asking whether the platform's
-- EXISTING decision-record mechanism already covers it. It does. `[MEASURED 2026-09-04]`
-- `doc_notes` carries live decision records under `categories ? 'decision'` and
-- `? 'decision-record'` — 4 rows `["decision","provenance","experience-council","decision-record"]`,
-- 4 `["council-gate","decision-record"]`, 4 `["decision"]` — with native `created_by`/`created_at`
-- and a `(subject_type, subject_key, created_at DESC)` index.
--
-- **THIS MIGRATION IS THE FIX, AND IT IS "BOTH", NOT "INSTEAD".** The `posture` key stays on the
-- register because that is where a consumer that has just LOADED the register will look, and it
-- travels with the row it governs. The `doc_notes` row is added because that is the platform's
-- typed, queryable, indexed record of who declared what and when — and because a hand-rolled
-- provenance object competing for space inside a shared config blob is exactly the
-- reuse-before-create smell the seat exists to catch. Cost of both: one INSERT. The two are
-- cross-referenced so neither can be found without the other.
--
-- ─── OBJECTION 3 (MEDIUM, debug_historian): a `<>` I fixed elsewhere in the same file ─────────
--
-- ALSO RIGHT, and the sharpest of the three because it is my own lesson not carried through.
-- 761's verify contains `IF rung <> 'relied_upon'`. On the path where the UPDATE matches no row,
-- `rung` is NULL and `NULL <> 'relied_upon'` is NULL, not TRUE — so that guard cannot fire. It is
-- masked only INCIDENTALLY by the next check (`declared IS NULL OR declared = ''`), which does
-- raise; remove or reorder that line and the guard goes silent. This is the identical defect a
-- mutation test caught two migrations earlier and that 761's own comment block explains at
-- length, three lines further down.
--
-- Nothing needs re-running: 761 applied against a matching row and its data is verified correct
-- (`rung=relied_upon`, 21 facts, 6 banned_claims, read back through a JOIN on `sites`). But the
-- FILE is a template others copy, so the correction is recorded here and in the lane's NOTES, and
-- **the verify below uses `IS DISTINCT FROM` throughout**. The general rule now in LANDMINES: in a
-- verify block, `<>` is wrong wherever the value can be NULL on the failure path — which is nearly
-- always, because the failure being guarded against is usually "the thing is gone".
--
-- Rollback: 767_..._ROLLBACK.sql

-- ⚠ `subject_type` is CHECK-constrained to (tool|pipeline|experience|action|experience-pattern|
-- landmine|component|decision). The first cut used 'site' and the INSERT was refused — caught by
-- the dry run, not by review. `decision` is the right value and is the live convention (33 rows;
-- the newest, `copyonline.co.uk`, uses exactly this shape: subject_key = the domain).

BEGIN;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM sites WHERE id = '72b9e3a6-872f-4528-a6d6-7f205ea60f4d' AND domain = 'vetcomparison.uk') THEN
    RAISE EXCEPTION '767 ABORT: site_id does not resolve to vetcomparison.uk';
  END IF;
END $$;

-- The doc_notes row must AGREE with the register, or the two records diverge from birth.
DO $$
DECLARE rung text; n int;
BEGIN
  SELECT data->'posture'->>'rung' INTO rung FROM site_specs
   WHERE site_id = '72b9e3a6-872f-4528-a6d6-7f205ea60f4d' AND aspect = 'evidence_base' AND is_current;
  IF rung IS DISTINCT FROM 'relied_upon' THEN
    RAISE EXCEPTION '767 ABORT: the register says rung=% - refusing to write a decision record that disagrees with it', coalesce(rung,'(no posture key)');
  END IF;

  SELECT count(*) INTO n FROM doc_notes
   WHERE subject_type = 'decision' AND subject_key = 'vetcomparison.uk' AND categories ? 'posture-rung';
  IF n <> 0 THEN
    RAISE EXCEPTION '767 ABORT: % posture-rung decision record(s) already exist for this site - read them before adding another', n;
  END IF;
END $$;

INSERT INTO doc_notes (subject_type, subject_key, site_id, body, categories, source, created_by)
VALUES (
  'decision',
  'vetcomparison.uk',
  '72b9e3a6-872f-4528-a6d6-7f205ea60f4d',
  'POSTURE RUNG: relied_upon (RFC_060 Q4 decision record).

DECLARED BY: RFC_060 §3a/§3b, which names vetcomparison.uk in its own worked list ("lendzy.co.uk, loancalculator.co.uk, vetcomparison.uk -> relied_upon") and cites the site as the RFC''s relied_upon worked example because it carries animal-health claims; and the vetcomparison lane''s handover of 2026-09-03 ("Rung: the CITED bar. vetcomparison is RFC_060''s own relied_upon worked example"). RECORDED (not decided) by the bugfix_414 register-programme lane under owner ruling D1 of 2026-09-03.

DECLARED ON: 2026-09-03. Recorded as a doc_notes decision record on 2026-09-04.

BASIS: a reader may act on this site''s assertions to their financial, legal or animal-safety detriment. The animal-health-certificate guide governs whether a pet may lawfully travel — the 21-full-day rabies wait, the 10-day certificate validity, the 24-to-120-hour tapeworm window — and the CMA guides tell a veterinary practice which statutory obligations bind it and by when. Both are act-on-it content, which is the relied_upon test in RFC_060 §3b.

WHAT THE RUNG REQUIRES HERE: the CITED bar — every fact carries source.citation{url,quote,...} with the quote verified through the production matcher — EXCEPT where the primary source is a format that matcher cannot read, where source.attested_by is used and the reason is recorded on the fact. 8 of the register''s 21 facts are in that second class: the CMA draft Order and draft Schedule 1 are PDFs, measured 2026-09-03 to return false for EVERY quote including one certainly present, and false for the absent control too — so a citation there would report citation_lost drift every day for ever.

REVIEW WHEN: the substantive CMA Order is due by its statutory deadline of 23 September 2026. On the day it is made, every fact whose id begins CMA-DRAFT- (exactly 5, and migration 763 made that prefix mean exactly "provisional") needs re-verification, and the bracketed placeholder figures become real numbers.

WHERE THE MACHINE-READABLE COPY LIVES: site_specs.data->''posture'' on this site''s current evidence_base row (migration 761). This note and that key are the SAME record in two places, deliberately: the key travels with the register a consumer has just loaded, and this row is the platform''s typed, indexed, queryable decision record. If they ever disagree, the register key is the one a running check will act on — reconcile toward whichever a human last signed.

STATUS OF THE SHAPE: the site_specs ''posture'' key is OFFERED to the claims-verification lane as a shape for RFC_060''s Q4 record, NOT declared as the fleet convention. Zero Go consumers as of 2026-09-04 (grep, non-test); one site carries it. This doc_notes row uses the platform''s existing decision-record convention and needs no such caveat.',
  '["decision", "decision-record", "posture-rung", "claims-verification", "rfc-060", "provenance"]'::jsonb,
  'migration 767',
  'bugfix_414 register-programme lane (migration 767)'
);

DO $$
DECLARE n int; nrung int; body_has_key boolean; reg_rung text;
BEGIN
  SELECT count(*) INTO n FROM doc_notes
   WHERE created_by = 'bugfix_414 register-programme lane (migration 767)';
  IF n IS DISTINCT FROM 1 THEN
    RAISE EXCEPTION '767 VERIFY: expected exactly 1 decision record, found %', coalesce(n::text,'NULL');
  END IF;

  -- it must be findable the way a reader would actually look for it
  SELECT count(*) INTO nrung FROM doc_notes
   WHERE subject_type = 'decision' AND subject_key = 'vetcomparison.uk'
     AND categories ? 'decision' AND categories ? 'posture-rung'
     AND site_id = '72b9e3a6-872f-4528-a6d6-7f205ea60f4d';
  IF nrung IS DISTINCT FROM 1 THEN
    RAISE EXCEPTION '767 VERIFY: the record is not findable by (subject, decision, posture-rung, site_id) - found %', coalesce(nrung::text,'NULL');
  END IF;

  -- and it must POINT AT the register key, or the two records drift apart unnoticed
  SELECT body LIKE '%site_specs.data->''posture''%' INTO body_has_key FROM doc_notes
   WHERE created_by = 'bugfix_414 register-programme lane (migration 767)';
  IF body_has_key IS DISTINCT FROM true THEN
    RAISE EXCEPTION '767 VERIFY: the decision record does not name the register key it mirrors';
  END IF;

  -- and the register must still say the same thing (IS DISTINCT FROM, per objection 3)
  SELECT data->'posture'->>'rung' INTO reg_rung FROM site_specs
   WHERE site_id = '72b9e3a6-872f-4528-a6d6-7f205ea60f4d' AND aspect = 'evidence_base' AND is_current;
  IF reg_rung IS DISTINCT FROM 'relied_upon' THEN
    RAISE EXCEPTION '767 VERIFY: register and decision record disagree - register says %', coalesce(reg_rung,'NULL');
  END IF;

  RAISE NOTICE '767 OK: posture recorded as a doc_notes decision record, cross-referenced with site_specs.data->posture, both saying %', reg_rung;
END $$;

COMMIT;
