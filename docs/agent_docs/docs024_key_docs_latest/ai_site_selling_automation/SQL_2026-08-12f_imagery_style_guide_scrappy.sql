-- FILE: SQL_2026-08-12f_imagery_style_guide_scrappy.sql
--
-- webdesign.uk: replace the imagery style guide for the £149 "take it or leave
-- it" brand (owner direction, 2026-08-12 evening).
--
-- THE ONE IDEA, and every motif is a restatement of it: an unorthodox, scrappy
-- PROCESS producing a genuinely good RESULT. A Heath Robinson contraption whose
-- output is immaculate. A scruffy goose that lays a golden egg. A plain
-- cardboard box with something well made inside.
--
-- THE RULE THAT GOVERNS EVERY IMAGE, stated as an `avoid` so it reaches the
-- negative prompt rather than living in a comment: **the cheapness is in the
-- SERVICE, never in the PRODUCT.** The machinery may be improvised, absurd,
-- held together with string. The thing it produces is always clean, finished
-- and good. That is the whole commercial proposition in a picture, and an image
-- that gets it backwards — shoddy output, junk, dereliction — sells the
-- opposite of what we do. It is also why the junkyard idea is NOT the base
-- register: a junkyard says the goods are scrap.
--
-- WHY THIS MEDIUM. Heath Robinson's own work is pen-and-ink line drawing, which
-- solves three problems at once: it carries a contraption, a goose and a
-- cardboard box in one coherent hand; it reads as deliberate craft rather than
-- as a cheap photograph; and line art survives the card-size gate that
-- photographic output kept failing (the D14 note in imagery_style_guide.go).
--
-- TEXT STAYS BANNED, deliberately kept from the previous guide. Generated
-- lettering garbles, and the shipping-label typography this brand wants is a
-- job for HTML type over the image, not for the generator.
--
-- SUBJECT vs STYLE: this spec carries STYLE. generate_image composes per-KIND
-- direction from medium/mood/palette and sends `avoid` to the negative prompt;
-- it does not choose subjects. The motif family is therefore named in `mood`,
-- where it reaches the generator as direction, and the per-page mapping is
-- recorded in the lane's NOTES for the imagery pipeline to work from rather
-- than being hand-assigned here.
--
-- A `kinds.content_hero` OVERRIDE REPLACES THE BASE WHOLESALE — including
-- empty values (imagery_style_guide.go, Phase I3.1). So the override below
-- restates palette and avoid in full rather than relying on inheritance.
--
-- ROLLBACK: the previous guide is superseded, not deleted —
--   UPDATE site_specs SET is_current=false, superseded_at=now()
--    WHERE site_id='1fcfa4f3-ec80-4010-878b-b971cd46711f' AND aspect='imagery_style_guide' AND is_current;
--   UPDATE site_specs SET is_current=true, superseded_at=NULL WHERE id='<the old row>';

BEGIN;

UPDATE site_specs SET is_current = false, superseded_at = now()
 WHERE site_id = '1fcfa4f3-ec80-4010-878b-b971cd46711f'
   AND aspect = 'imagery_style_guide' AND is_current = true;

INSERT INTO site_specs
       (site_id, aspect, data, source, notes, is_current, created_by, pinned)
SELECT '1fcfa4f3-ec80-4010-878b-b971cd46711f', 'imagery_style_guide', $json$
{
  "medium": "Pen-and-ink line drawing with a single flat wash, in the manner of an early-twentieth-century patent diagram or a Heath Robinson contraption cartoon. Confident, controlled hand-drawn linework with visible construction: string, pulleys, planks, clamps, improvised joints and counterweights. Real draughtsmanship. Never sketchy, never scribbled, never unfinished.",

  "mood": "An absurd, over-engineered contraption whose output is unexpectedly excellent. Dry, deadpan, faintly comic; the joke is played straight. The recurring motif family, any of which may carry an image: a Heath Robinson machine assembling a finished website; a marble run or conveyor of improbable steps; a plain cardboard box or flat-pack panel with something well made inside; a scruffy goose beside one perfect golden egg; a pallet of goods in their shipping boxes under strip lighting; a trade counter with a price list on the wall. All of them say the same thing, so any one of them fits any page.",

  "palette": "Kraft cardboard ground (#d8c3a0 through #c8ab82), near-black ink (#1a1a1a), and a single warm gold (#c8961e) used only on the one element in the image that matters — the finished thing, the golden egg, the output. Nothing else carries colour.",

  "avoid": "ANY suggestion that the OUTPUT is poor quality: the thing the machine produces, the object in the box, the egg, the finished website is always clean, complete and well made. No rust, refuse, decay, dereliction, bins, skips, litter or broken objects. No chaos: the contraption is absurd but every part of it is drawn as if it works. No stock photography of business people, handshakes, laptops on desks or people pointing at screens. No cliched web-design imagery: wireframe sketches, colour swatches fanned out, pencils and rulers, browser windows drawn in perspective, code on screens. No rocket ships, lightbulbs or jigsaw pieces. No photorealism or photographic texture. No 3D rendering, gradients or drop shadows. No text, lettering, numerals, logos or watermarks of any kind. No cuteness, no whimsy for its own sake, and nothing apologetic.",

  "kinds": {
    "content_hero": {
      "medium": "Pen-and-ink line drawing with a single flat wash, in the manner of a patent diagram or a Heath Robinson contraption cartoon. One clear motif, strong silhouette, generous air around it. Controlled linework, no crosshatching so dense it muddies at card size.",
      "mood": "One absurd machine, or one scruffy creature, or one plain box, producing or containing one immaculate result. Deadpan. Played straight.",
      "palette": "Kraft cardboard ground (#d8c3a0 through #c8ab82), near-black ink (#1a1a1a), a single warm gold (#c8961e) on the output element only.",
      "avoid": "Any suggestion the output is poor quality. No rust, refuse, decay, dereliction or broken objects. No photorealism, photographic texture, heavy gradients, 3D rendering or drop shadows. No text, lettering, numerals, logos or watermarks. No busy detail that collapses at card size. No colour outside the palette. No dark or saturated full-bleed backgrounds.",
      "reference_asset_keys": []
    }
  },

  "reference_asset_keys": []
}
$json$::jsonb,
       'owner_ruling',
       'The £149 take-it-or-leave-it brand. Owner direction 2026-08-12 evening: '
       'Heath Robinson contraptions, a scruffy goose with a golden egg, plain '
       'cardboard boxes and cash-and-carry pallets, all telling one story — an '
       'unorthodox process with a genuinely good result. Supersedes the '
       '"assured modern studio / flat two-colour editorial illustration" guide '
       'of 2026-08-04. The governing constraint is in `avoid` so it reaches the '
       'negative prompt: the cheapness is in the SERVICE, never in the PRODUCT. '
       'Owner asked that the framework generate from these guidelines rather '
       'than images being made by hand at the CLI.',
       true,
       'ai-site-selling-automation-2026-08-12',
       COALESCE(pinned, false)
  FROM site_specs
 WHERE site_id = '1fcfa4f3-ec80-4010-878b-b971cd46711f'
   AND aspect = 'imagery_style_guide' AND is_current = false
 ORDER BY superseded_at DESC NULLS LAST
 LIMIT 1;

DO $$
DECLARE n_current int; is_pinned bool; has_override bool; bans_output_quality bool;
BEGIN
  SELECT count(*) INTO n_current FROM site_specs
   WHERE site_id='1fcfa4f3-ec80-4010-878b-b971cd46711f' AND aspect='imagery_style_guide' AND is_current;
  IF n_current <> 1 THEN RAISE EXCEPTION 'expected exactly 1 current guide, found %', n_current; END IF;

  SELECT COALESCE(pinned,false), data->'kinds' ? 'content_hero',
         data->>'avoid' ILIKE '%output is poor quality%'
    INTO is_pinned, has_override, bans_output_quality
    FROM site_specs
   WHERE site_id='1fcfa4f3-ec80-4010-878b-b971cd46711f' AND aspect='imagery_style_guide' AND is_current;

  IF NOT is_pinned THEN RAISE EXCEPTION 'pinned not inherited — an agent could overwrite the brand'; END IF;
  IF NOT has_override THEN RAISE EXCEPTION 'content_hero override missing — heroes would fall back to the base voice'; END IF;

  -- The governing rule must actually be in the negative prompt, not just in
  -- this file's comments. A guide that merely *reads* correctly is the failure
  -- mode: `avoid` is the only field that reaches the generator as a constraint.
  IF NOT bans_output_quality THEN
    RAISE EXCEPTION 'the cheapness-is-in-the-service rule is not in `avoid` — it would not reach the negative prompt';
  END IF;
END $$;

COMMIT;
