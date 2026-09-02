-- 718_planner_imagery_content_expected_prompt_and_exemplar.sql
--
-- OWNER DIRECTIVE 2026-09-02 (designblog.co.uk critique session): "please go ahead with the
-- planner prompt and exemplar changes." Context: the estate has planned ONE infographic ever
-- (hero 399 / icon 211 / logo 50 / illustration 25 / infographic 1, measured 2026-09-02), and
-- the cause is NOT missing vocabulary — the `kind` enum already carries illustration and
-- infographic. Three things in this same prompt suppress them (inline_guide_imagery
-- NOTES_inline_guide_imagery.md §15, quotes verified against the live row 2026-09-02):
--   (1) the sections bullet instructs "Use sparingly in v1 — most plans will have zero
--       section-scope entries";
--   (2) rule 13's minimum is chrome-only (logo + heroes), no floor for content imagery;
--   (3) the worked example's sections block shows ONLY icons and the Return-JSON skeleton
--       shows "sections": {} — and exemplars ship verbatim (LANDMINES:
--       a-quoted-exemplar-in-a-prompt-is-copied-verbatim; the copy lane's proven mechanism:
--       demonstrations govern, instructions don't).
-- So the near-zero is the planner OBEYING. This migration flips all four surfaces in one
-- change: the bullet now states content-carrying imagery is expected (with two limits: only
-- onto sections whose component can display it — the 4.1 undisplayable-asset trap — and only
-- where there is something real to depict); rule 13 gains a section-scope illustration-or-
-- infographic floor for the index page and EXEMPTS empty-sections pages (rule 3: blog-post
-- etc. — article pages have no structure to host imagery, bugs_open/114; that is a separate
-- ask, deliberately NOT forced here); the worked example gains an index:1 illustration and a
-- tools:1 infographic (both text-free — image models render garbage lettering; headings live
-- in HTML); the skeleton's sections map demonstrates one entry.
-- Rule 16 (one image per entry) is deliberately UNTOUCHED and its presence is asserted in the
-- verify block — raising section-scope volume without it produces multi-panel collages.
-- Downstream: IMG-075 per-section binding is live (2026-09-02 15:39 roll); kinds pass the DB
-- CHECK (rule 15). Cost: more generated images per build — accepted by the owner in ruling.
--
-- Anchors: five, all asserted exactly-once against the live text (which two other lanes also
-- edit — LANDMINES 17024: anchored replace() with exact-count guards, never wholesale
-- jsonb_set). Local dry-run: all five spliced, balance checks pass, ROLLBACK roundtrip
-- restores the original byte-exactly.
--
-- Apply: psql -f THIS FILE ONLY (never an unscoped runner --apply). Companion ROLLBACK
-- alongside. SITE_DEFECT_CATEGORIES.md §4.6 documents the category this closes.

BEGIN;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '718 REFUSED: expected exactly 1 active build-site-planner row, found %', n;
  END IF;
  PERFORM snapshot_agent('build-site-planner',
                         '718_planner_imagery_content_expected_prompt_and_exemplar.sql: pre-update');
END $$;

DO $do$
DECLARE
  tpl text; newtpl text; n int;
  ifs_before int; ends_before int; elses_before int; vars_before int;
  anchor_A text := $A718$Use sparingly in v1 — most plans will have zero section-scope entries. Only emit a section entry when a specific section's imagery need is not covered by the page hero.$A718$;
  repl_A text := $RA718$Content-carrying imagery is EXPECTED here, not exceptional: wherever a section presents an idea a picture can carry, emit a section entry — an `illustration` for a concept, process or scene, an `infographic` for numbers, comparisons or steps, `icon` entries for feature/benefit groups. A page hero is chrome; it does not satisfy a section's need to SHOW what that section says. The index page and other section-bearing content pages should rarely ship with zero section-scope entries. Two limits keep this honest: attach an entry only to a section whose component can actually display it (an illustrated, media-capable or card-style component from the Available Section Components list — never a plain prose block), and only where the section has something real to depict — an entry must carry content, not decorate it.$RA718$;
  anchor_B text := $B718$13. Populate the `imagery` block per the Imagery Block rules above. At minimum: one site-scope `logo` entry, one page-scope `hero` entry under `pages.index`, and one page-scope `hero` entry for every other page whose `sections` array contains a hero-class component$B718$;
  repl_B text := $RB718$13. Populate the `imagery` block per the Imagery Block rules above. At minimum: one site-scope `logo` entry, one page-scope `hero` entry under `pages.index`, and one page-scope `hero` entry for every other page whose `sections` array contains a hero-class component, and at least one section-scope `illustration` or `infographic` entry under the index page. Pages with empty sections arrays (rule 3) take NO section-scope entries — never attach an imagery entry to a section that does not exist. Rule 16 applies to every section-scope entry$RB718$;
  anchor_C text := $C718$"sections": {
    "index:2": [$C718$;
  repl_C text := $RC718$"sections": {
    "index:1": [
      {
        "key": "illustration_how_it_works",
        "kind": "illustration",
        "prompt": "A stylised cutaway of an industrial robotic gripper holding a delicate object, technical-drawing feel, no text or labels anywhere in the image",
        "style_hints": {"aspect_ratio": "4:3"}
      }
    ],
    "tools:1": [
      {
        "key": "infographic_selection_steps",
        "kind": "infographic",
        "prompt": "A clean vertical flow of four unlabelled stages from part shape to gripper choice, abstract geometric forms connected by arrows, flat design, no text anywhere in the image — headings are set in HTML beside the graphic",
        "style_hints": {"aspect_ratio": "4:3"}
      }
    ],
    "index:2": [$RC718$;
  anchor_D text := $D718$The three icon entries in `sections."index:2"` demonstrate the key decomposition principle: each conceptually-distinct image gets its own entry — never describe multiple images in a single prompt.$D718$;
  repl_D text := $RD718$The `index:1` illustration and the `tools:1` infographic demonstrate content-carrying section imagery: each is attached to the specific section it serves, describes ONE image, and keeps all wording out of the image (headings and labels are set in HTML beside the graphic). The three icon entries in `sections."index:2"` demonstrate the key decomposition principle: each conceptually-distinct image gets its own entry — never describe multiple images in a single prompt.$RD718$;
  anchor_E text := $E718$"sections": {}$E718$;
  repl_E text := $RE718$"sections": {
      "index:1": [
        {"key": "illustration_how_it_works", "kind": "illustration", "prompt": "..."}
      ]
    }$RE718$;
BEGIN
  SELECT default_config #>> '{workflow,steps,plan_site,config,prompt_template}' INTO tpl
    FROM agent_definitions WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF tpl IS NULL THEN RAISE EXCEPTION '718: plan_site.config.prompt_template not found'; END IF;

  n := (length(tpl) - length(replace(tpl, anchor_A, ''))) / length(anchor_A);
  IF n <> 1 THEN RAISE EXCEPTION '718: anchor A found % times, expected 1', n; END IF;
  n := (length(tpl) - length(replace(tpl, anchor_B, ''))) / length(anchor_B);
  IF n <> 1 THEN RAISE EXCEPTION '718: anchor B found % times, expected 1', n; END IF;
  n := (length(tpl) - length(replace(tpl, anchor_C, ''))) / length(anchor_C);
  IF n <> 1 THEN RAISE EXCEPTION '718: anchor C found % times, expected 1', n; END IF;
  n := (length(tpl) - length(replace(tpl, anchor_D, ''))) / length(anchor_D);
  IF n <> 1 THEN RAISE EXCEPTION '718: anchor D found % times, expected 1', n; END IF;
  n := (length(tpl) - length(replace(tpl, anchor_E, ''))) / length(anchor_E);
  IF n <> 1 THEN RAISE EXCEPTION '718: anchor E found % times, expected 1', n; END IF;

  ifs_before   := (length(tpl) - length(replace(tpl, '{{if ', ''))) / length('{{if ');
  ends_before  := (length(tpl) - length(replace(tpl, '{{end}}', ''))) / length('{{end}}');
  elses_before := (length(tpl) - length(replace(tpl, '{{else}}', ''))) / length('{{else}}');
  vars_before  := (length(tpl) - length(replace(tpl, '{{.', ''))) / length('{{.');

  newtpl := tpl;
  newtpl := replace(newtpl, anchor_A, repl_A);
  newtpl := replace(newtpl, anchor_B, repl_B);
  newtpl := replace(newtpl, anchor_C, repl_C);
  newtpl := replace(newtpl, anchor_D, repl_D);
  newtpl := replace(newtpl, anchor_E, repl_E);

  IF length(newtpl) <> length(tpl) + (length(repl_A) - length(anchor_A)) + (length(repl_B) - length(anchor_B)) + (length(repl_C) - length(anchor_C)) + (length(repl_D) - length(anchor_D)) + (length(repl_E) - length(anchor_E)) THEN
    RAISE EXCEPTION '718: unexpected length delta';
  END IF;

  IF (length(newtpl) - length(replace(newtpl, '{{.', ''))) / length('{{.') <> vars_before THEN
    RAISE EXCEPTION '718: replacement introduced a template variable — it would render EMPTY without input_fields';
  END IF;
  IF (length(newtpl) - length(replace(newtpl, '{{if ', ''))) / length('{{if ') <> ifs_before
     OR (length(newtpl) - length(replace(newtpl, '{{end}}', ''))) / length('{{end}}') <> ends_before
     OR (length(newtpl) - length(replace(newtpl, '{{else}}', ''))) / length('{{else}}') <> elses_before THEN
    RAISE EXCEPTION '718: template if/else/end balance changed';
  END IF;

  UPDATE agent_definitions
     SET default_config = jsonb_set(default_config,
           '{workflow,steps,plan_site,config,prompt_template}', to_jsonb(newtpl), false),
         updated_at = now()
   WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN RAISE EXCEPTION '718: updated % rows, expected exactly 1', n; END IF;
END $do$;

-- Verify (DO/RAISE): suppressors gone, new surfaces present, rule 16 and the kind enum intact.
DO $$
DECLARE tpl text; n int;
BEGIN
  SELECT default_config #>> '{workflow,steps,plan_site,config,prompt_template}' INTO tpl
    FROM agent_definitions WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF position('Use sparingly in v1' in tpl) > 0 THEN
    RAISE EXCEPTION '718 VERIFY: default-to-zero instruction still present';
  END IF;
  IF position('most plans will have zero section-scope entries' in tpl) > 0 THEN
    RAISE EXCEPTION '718 VERIFY: zero-expectation sentence still present';
  END IF;
  IF position('Content-carrying imagery is EXPECTED' in tpl) = 0 THEN
    RAISE EXCEPTION '718 VERIFY: new sections bullet missing';
  END IF;
  IF position('at least one section-scope `illustration` or `infographic` entry under the index page' in tpl) = 0 THEN
    RAISE EXCEPTION '718 VERIFY: rule-13 floor missing';
  END IF;
  IF position('infographic_selection_steps' in tpl) = 0 THEN
    RAISE EXCEPTION '718 VERIFY: worked-example infographic missing';
  END IF;
  n := (length(tpl) - length(replace(tpl, '16. Each entry in `imagery` produces exactly ONE image.', '')))
       / length('16. Each entry in `imagery` produces exactly ONE image.');
  IF n <> 1 THEN RAISE EXCEPTION '718 VERIFY: rule 16 not present exactly once (found %)', n; END IF;
  n := (length(tpl) - length(replace(tpl, 'one of: `logo`, `hero`, `illustration`, `icon`, `infographic`. No other values permitted.', '')))
       / length('one of: `logo`, `hero`, `illustration`, `icon`, `infographic`. No other values permitted.');
  IF n <> 1 THEN RAISE EXCEPTION '718 VERIFY: kind enum not present exactly once (found %)', n; END IF;
  IF position('"sections": {}' in tpl) > 0 THEN
    RAISE EXCEPTION '718 VERIFY: empty-sections skeleton still present';
  END IF;
  RAISE NOTICE '718 OK: content imagery expected, floor set, exemplars demonstrate it, rule 16 intact.';
END $$;

COMMIT;
