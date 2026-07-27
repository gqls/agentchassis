-- 230_finetuning_remove_unevidenced_80pct_claim.sql
-- OWNER RULING 2026-07-27: remove the "~80% reduction in quote preparation time"
-- claim from finetuning.uk. It was surfaced by the item-(b) survey as the ONE
-- genuine unevidenced business claim across four sites, and nothing on record
-- supports the figure.
--
-- ── WHAT IS BEING REMOVED ──────────────────────────────────────────────────
-- finetuning.uk /index, position 4, component `case-studies-grid`:
--
--   card1_client_name : Facilities management company (Midlands, UK)
--   card1_stat_value  : ~80%                                  <-- REMOVED
--   card1_stat_label  : reduction in quote preparation time   <-- REMOVED
--
-- Nothing else on the card changes. The card does not depend on the figure —
-- its own excerpt already makes the point qualitatively, and truthfully:
--
--   "Response times dropped substantially and the team stopped losing bids to
--    faster competitors."
--
-- ── WHY BLANKING IS SAFE HERE, AND HOW THAT WAS CHECKED ────────────────────
-- Blanking a stat field is only safe if the template GATES it; an ungated
-- field leaves an empty <strong></strong> on the page, which is bugs_closed/073.
-- Checked before writing this file — `case-studies-grid`'s html_template:
--
--   {{if .card1_stat_value}}<span class="csg-card-stat"><strong>{{.card1_stat_value}}</strong> {{.card1_stat_label}}</span>{{end}}
--
-- The gate is there (it is migration 217's work, 043 candidate 1), so an empty
-- value removes the whole span cleanly. This is the "honest empty stat" path
-- 073 was closed to make possible.
--
-- ── AND WHY BLANKING ALONE WOULD NOT BE ENOUGH ─────────────────────────────
-- Deleting a figure from content_data does not stop the writer inventing it
-- again on the next rebuild — that is this lane's entire thesis. finetuning.uk
-- has NO evidence_base row at all (checked: 0 rows, current or historic), so it
-- has neither the preventive rail (writer_block) nor the detective one
-- (banned_claims). This file adds both, which is what makes the removal durable
-- rather than cosmetic.
--
-- ── KNOWN CONSEQUENCE, STATED RATHER THAN DISCOVERED LATER ─────────────────
-- Creating this row turns the claims scans ON for finetuning.uk (banned_claims
-- is non-empty, so ParseEvidenceBase returns non-nil). The site publishes four
-- other numbers in prose, measured 2026-07-27 with the real scanner, and they
-- WILL be raised as `claims_unverified` items on the next discovery run:
--
--   "…anyone under the age of 16"                     (privacy policy)
--   "…the 5 to 50 employee businesses that make up…"  (audience descriptor, x3)
--
-- None is a business claim; all four are `bugs_open/102`'s class (the layer has
-- no notion of page_type or of policy text). They are NOT registered as facts
-- here, because a fact must come from a live query or an explicit owner
-- attestation and an audience descriptor is neither — inventing a fact row to
-- silence a false positive would be the exact failure this register exists to
-- prevent. Four dismissible review items is the honest price of the rail.

BEGIN;

DO $fix$
DECLARE
    v_site uuid;
    v_pc   uuid;
    n      int;
BEGIN
    SELECT id INTO v_site FROM sites WHERE domain = 'finetuning.uk';
    IF v_site IS NULL THEN
        RAISE EXCEPTION '230: no site row for finetuning.uk';
    END IF;

    -- ── 1. Remove the claim from the stored content ────────────────────────
    SELECT pc.id INTO v_pc
    FROM page_components pc
    JOIN pages p ON p.id = pc.page_id
    LEFT JOIN content_components cc ON cc.id = pc.component_id
    WHERE p.site_id = v_site
      AND COALESCE(cc.name,'') = 'case-studies-grid'
      AND pc.content_data->>'card1_stat_value' = '~80%';

    IF v_pc IS NULL THEN
        RAISE EXCEPTION '230: the ~80%% stat was not found where expected (finetuning case-studies-grid card1) — re-survey before forcing this';
    END IF;

    UPDATE page_components
    SET content_data = jsonb_set(
            jsonb_set(content_data, '{card1_stat_value}', '""'::jsonb),
            '{card1_stat_label}', '""'::jsonb),
        updated_at = now()
    WHERE id = v_pc
      AND content_data ? 'card1_stat_value'
      AND content_data ? 'card1_stat_label';

    GET DIAGNOSTICS n = ROW_COUNT;
    IF n <> 1 THEN
        RAISE EXCEPTION '230: expected to update 1 component, updated %', n;
    END IF;

    -- ── 2. Make the removal durable: the site's first evidence_base ────────
    -- Guarded against a concurrent seeding by another session.
    IF EXISTS (SELECT 1 FROM site_specs WHERE site_id = v_site
               AND aspect = 'evidence_base' AND is_current) THEN
        RAISE EXCEPTION '230: finetuning.uk already has a current evidence_base row — another session seeded one; merge by hand rather than superseding blind';
    END IF;

    INSERT INTO site_specs (site_id, aspect, data, source, created_by, notes, is_current)
    VALUES (
      v_site, 'evidence_base',
      $eb$ {
        "audit_doc": "docs/agent_docs/docs024_key_docs_latest/fabricated_stats_043/ — owner ruling 2026-07-27",
        "governing_rule": "finetuning.uk publishes NO quantified client outcome unless the owner has attested it and it is registered as a fact below. Qualitative outcome language is fine and is what the case-study excerpts already use.",
        "facts": [],
        "banned_claims": [
          {"pattern": "~?\\s*80\\s*%[^.]{0,40}(quote|quoting|preparation)",
           "reason": "2026-07-27 owner ruling: '~80% reduction in quote preparation time' (index, case-studies-grid card1) had nothing on record behind it and was removed. It must not return."},
          {"pattern": "[0-9]{1,3}\\s*%\\s*(reduction|increase|faster|improvement|saving|savings|uplift|growth)",
           "reason": "unevidenced-outcome class: no client engagement on this site has an attested measured result. A percentage outcome is an invention until an owner attestation puts it in facts[]."},
          {"pattern": "(saved|saving|cut|reduced)\\s+[0-9,.]+\\s*(hours|hrs|days|weeks)",
           "reason": "same class in absolute units — a time-saved figure is an invention until attested."}
        ],
        "writer_block": "NUMBERS (state only these): none are registered for this site yet. NOT TRACKED, NEVER STATE: any quantified client outcome — percentage reductions or increases, hours or days saved, response-time improvements, quote-turnaround figures, ROI or payback periods. No engagement on this site has a measured, attested result, so every such number at any value is an invention. Describe outcomes QUALITATIVELY instead ('response times dropped substantially', 'the team stopped losing bids'), which is what the existing case studies do and is not a claim that can be false. Audience descriptors ('5 to 50 employees') are positioning, not counts, and are fine."
      } $eb$::jsonb,
      'owner_ruling', 'bugfix-043-lane',
      'Created 2026-07-27 to make the ~80% removal durable. See sql_for_agents/230.',
      true
    );
END $fix$;

-- ── Post-conditions ────────────────────────────────────────────────────────
DO $post$
DECLARE
    v_left  int;
    v_bans  int;
BEGIN
    -- The claim is gone from stored content, in EVERY form, on the whole site.
    SELECT count(*) INTO v_left
    FROM page_components pc
    JOIN pages p ON p.id = pc.page_id
    JOIN sites s ON s.id = p.site_id,
    LATERAL jsonb_each_text(pc.content_data) e(k,v)
    WHERE s.domain = 'finetuning.uk' AND e.v ILIKE '%80%%';
    IF v_left <> 0 THEN
        RAISE EXCEPTION '230: % content_data field(s) on finetuning.uk still carry an 80%% figure', v_left;
    END IF;

    SELECT jsonb_array_length(ss.data->'banned_claims') INTO v_bans
    FROM site_specs ss JOIN sites s ON s.id = ss.site_id
    WHERE ss.aspect='evidence_base' AND ss.is_current AND s.domain='finetuning.uk';
    IF v_bans <> 3 THEN
        RAISE EXCEPTION '230: expected 3 banned_claims on the new register, found %', v_bans;
    END IF;

    RAISE NOTICE '230 OK: ~80%% claim removed from content_data; finetuning.uk now has a register with % banned_claims and a writer_block', v_bans;
END $post$;

COMMIT;
