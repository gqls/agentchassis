-- 718_planner_imagery_content_expected_prompt_and_exemplar_ROLLBACK.sql
-- Reverses 718 by anchored replace of each NEW text back to the ORIGINAL, with the same
-- exact-count guards. Roundtrip proven byte-exact against the 2026-09-02 live text before
-- shipping. Refuses if any new text is not present exactly once (i.e. 718 not applied, or a
-- later edit overlapped it — resolve by hand from the snapshot taken by 718's apply).

BEGIN;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '718R REFUSED: expected exactly 1 active build-site-planner row, found %', n;
  END IF;
  PERFORM snapshot_agent('build-site-planner',
                         '718_..._ROLLBACK.sql: pre-rollback');
END $$;

DO $do$
DECLARE
  tpl text; newtpl text; n int;
  anchor_A text := $A718$Content-carrying imagery is EXPECTED here, not exceptional: wherever a section presents an idea a picture can carry, emit a section entry — an `illustration` for a concept, process or scene, an `infographic` for numbers, comparisons or steps, `icon` entries for feature/benefit groups. A page hero is chrome; it does not satisfy a section's need to SHOW what that section says. The index page and other section-bearing content pages should rarely ship with zero section-scope entries. Two limits keep this honest: attach an entry only to a section whose component can actually display it (an illustrated, media-capable or card-style component from the Available Section Components list — never a plain prose block), and only where the section has something real to depict — an entry must carry content, not decorate it.$A718$;
  repl_A text := $RA718$Use sparingly in v1 — most plans will have zero section-scope entries. Only emit a section entry when a specific section's imagery need is not covered by the page hero.$RA718$;
  anchor_B text := $B718$13. Populate the `imagery` block per the Imagery Block rules above. At minimum: one site-scope `logo` entry, one page-scope `hero` entry under `pages.index`, and one page-scope `hero` entry for every other page whose `sections` array contains a hero-class component, and at least one section-scope `illustration` or `infographic` entry under the index page. Pages with empty sections arrays (rule 3) take NO section-scope entries — never attach an imagery entry to a section that does not exist. Rule 16 applies to every section-scope entry$B718$;
  repl_B text := $RB718$13. Populate the `imagery` block per the Imagery Block rules above. At minimum: one site-scope `logo` entry, one page-scope `hero` entry under `pages.index`, and one page-scope `hero` entry for every other page whose `sections` array contains a hero-class component$RB718$;
  anchor_C text := $C718$"sections": {
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
    "index:2": [$C718$;
  repl_C text := $RC718$"sections": {
    "index:2": [$RC718$;
  anchor_D text := $D718$The `index:1` illustration and the `tools:1` infographic demonstrate content-carrying section imagery: each is attached to the specific section it serves, describes ONE image, and keeps all wording out of the image (headings and labels are set in HTML beside the graphic). The three icon entries in `sections."index:2"` demonstrate the key decomposition principle: each conceptually-distinct image gets its own entry — never describe multiple images in a single prompt.$D718$;
  repl_D text := $RD718$The three icon entries in `sections."index:2"` demonstrate the key decomposition principle: each conceptually-distinct image gets its own entry — never describe multiple images in a single prompt.$RD718$;
  anchor_E text := $E718$"sections": {
      "index:1": [
        {"key": "illustration_how_it_works", "kind": "illustration", "prompt": "..."}
      ]
    }$E718$;
  repl_E text := $RE718$"sections": {}$RE718$;
BEGIN
  SELECT default_config #>> '{workflow,steps,plan_site,config,prompt_template}' INTO tpl
    FROM agent_definitions WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF tpl IS NULL THEN RAISE EXCEPTION '718R: plan_site.config.prompt_template not found'; END IF;

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

  newtpl := tpl;
  newtpl := replace(newtpl, anchor_A, repl_A);
  newtpl := replace(newtpl, anchor_B, repl_B);
  newtpl := replace(newtpl, anchor_C, repl_C);
  newtpl := replace(newtpl, anchor_D, repl_D);
  newtpl := replace(newtpl, anchor_E, repl_E);

  IF length(newtpl) <> length(tpl) + (length(repl_A) - length(anchor_A)) + (length(repl_B) - length(anchor_B)) + (length(repl_C) - length(anchor_C)) + (length(repl_D) - length(anchor_D)) + (length(repl_E) - length(anchor_E)) THEN
    RAISE EXCEPTION '718R: unexpected length delta';
  END IF;

  UPDATE agent_definitions
     SET default_config = jsonb_set(default_config,
           '{workflow,steps,plan_site,config,prompt_template}', to_jsonb(newtpl), false),
         updated_at = now()
   WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN RAISE EXCEPTION '718R: updated % rows, expected exactly 1', n; END IF;
END $do$;

DO $$
DECLARE tpl text;
BEGIN
  SELECT default_config #>> '{workflow,steps,plan_site,config,prompt_template}' INTO tpl
    FROM agent_definitions WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF position('Use sparingly in v1' in tpl) = 0 THEN
    RAISE EXCEPTION '718R VERIFY: original sections bullet not restored';
  END IF;
  IF position('Content-carrying imagery is EXPECTED' in tpl) > 0 THEN
    RAISE EXCEPTION '718R VERIFY: 718 text still present';
  END IF;
  RAISE NOTICE '718R OK: original prompt restored.';
END $$;

COMMIT;
