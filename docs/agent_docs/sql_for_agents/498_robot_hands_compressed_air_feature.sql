-- 498_robot_hands_compressed_air_feature.sql
-- Third news editorial feature, second on robot-hands.com:
-- /insights/electric-vs-pneumatic-economics.html
--
-- CLUSTER: four end-of-arm launches in one week, all naming electric over
-- pneumatic (DESTACO eRDH "pneumatic-free"; AWR standardising CNC machine
-- tending on OnRobot electric grippers for changeover speed; OnRobot electric
-- grippers into high-payload work; BMW humanoids running electric grippers).
--
-- PREMISE (design D's test): the load-bearing claim is that the shift is driven
-- by the ENERGY ARITHMETIC OF COMPRESSED AIR, not by electric grippers gripping
-- harder. Falsifiable against one thing — if compressed air were efficient, the
-- argument collapses and the launches are just product news. It is not: the US
-- DOE puts a typical plant's compressed-air share of electricity at 10% (30%+ at
-- some), overall system efficiency as low as 10%, and gives a worked example of
-- 7-8 electrical hp supplied to deliver 1 hp at the air motor.
--
-- SOURCE: every figure is a verbatim quote from ENERGY STAR / US DOE,
-- "Determine the Cost of Compressed Air for Your Plant", extracted from the PDF
-- with pdftotext in-session on 2026-08-21 rather than from a vendor blog
-- restating it (the search results were full of compressor-vendor pages quoting
-- these same numbers second-hand; the primary is free and unambiguous).
--
-- THE COUNTER-ARGUMENT IS IN THE PIECE, and it uses this site's OWN registered
-- facts: of 10 indexed gripper models the electric parallel-jaw family is the
-- largest at 5, and pneumatic, vacuum, magnetic, soft-robotic and adhesive each
-- still hold a place. Vacuum is explicitly named as doing a different job rather
-- than being displaced. A feature that only argued one way would be an
-- advertisement.
--
-- Third chart deliberately mixes a share-of-electricity figure with an
-- efficiency figure on one scale; the source_note says so on the page, because
-- they are different measures and a reader must not take the bars as comparable.
--
-- ⚠ THE meta_description GUARD BELOW FIRED ON THIS FILE'S FIRST APPLY, at 162
-- chars against its own 160 limit, and rolled the whole transaction back. That
-- guard exists because of misstep 6 (a hand-typed commissioning note serving as
-- a public description, bugs_open/339) and it caught the very next feature — by
-- length rather than by tone, which is the half a machine can check. Shortened
-- to 151. Worth keeping in mind that the tone half is still unguarded and
-- remains a human read.
--
-- DEPENDS ON 499 (facts). Hero ships with its image; deploy only once the asset
-- is active. ROLLBACK: delete the page_components rows then the pages row.

\set ON_ERROR_STOP on
BEGIN;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM site_specs, jsonb_array_elements(data->'facts') f
     WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND aspect='evidence_base' AND is_current AND f->>'id'='rh-air-efficiency') THEN
    RAISE EXCEPTION 'apply 499 first - the page cites facts that are not registered';
  END IF;
  IF EXISTS (SELECT 1 FROM pages WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND name='electric-vs-pneumatic-economics') THEN
    RAISE EXCEPTION 'page already exists - refusing double apply';
  END IF;
END $$;

INSERT INTO pages (site_id, name, url, title, page_type, status, build_status,
                   nav_label, nav_order, in_header, in_footer, rebuild_policy,
                   meta_description, sections, topics)
VALUES ('00ff3af5-dad8-4770-9f70-3edc267a3c92','electric-vs-pneumatic-economics',
 '/insights/electric-vs-pneumatic-economics.html',
 'The Case Against Compressed Air, in Its Own Numbers | Robot-Hands.com',
 'blog-post','active','pending','Compressed air economics',81,false,false,'owned',
 'Four gripper launches said pneumatic-free. A typical plant spends 10% of its electricity making compressed air, and 7-8 hp goes in for 1 hp at the tool.',
 '["hero", "feature-analysis", "evidence-chart-air", "feature-coverage", "call-to-action"]'::jsonb,
 ARRAY['electric grippers','compressed air','end-of-arm tooling','editorial feature']);

CREATE TEMP TABLE _pg ON COMMIT DROP AS
  SELECT id FROM pages WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND name='electric-vs-pneumatic-economics';

INSERT INTO page_components (page_id, component_id, position, slot_name, content_data, rendered_html,
                             build_status, locked_at, locked_by, lock_type)
SELECT pg.id, '23f95f00-f293-466e-b43a-81791ea0fc6c', 1, 'hero', $c1${
 "headline": "The Case Against Compressed Air, in Its Own Numbers",
 "subheadline": "Four gripper launches in a week, and the word they keep using is pneumatic-free. The reason is not that electric grips harder. It is that a plant spends roughly a tenth of its electricity making air, and most of that never reaches the tool.",
 "hero_url": "/assets/images/content-hero-electric-vs-pneumatic-economics.jpg",
 "background_image": "/assets/images/content-hero-electric-vs-pneumatic-economics.jpg",
 "cta_text": "Compare gripper specifications",
 "cta_url": "/tools/matchmatrix/index.html",
 "cta_target_title": "Run MatchMatrix | Gripper Selection Tool | Robot-Hands.com",
 "secondary_cta": "Browse the gripper index",
 "secondary_cta_url": "/gripper-catalog/index.html",
 "secondary_cta_target_title": "Browse Gripper Catalog | Filter by Technology, Payload, IP Rating | Robot-Hands.com"
}$c1$::jsonb, $r1$<section class="hero" data-component="hero" style="background-image: linear-gradient(rgba(0,0,0,0.5), rgba(0,0,0,0.6)), url('/assets/images/content-hero-electric-vs-pneumatic-economics.jpg'); background-size: cover; background-position: center; --hero-ink: #fff; --hero-btn-ink: #0F1115;">
        <div class="hero-content">
            <h1>The Case Against Compressed Air, in Its Own Numbers</h1>
            <p class="hero-subheadline">Four gripper launches in a week, and the word they keep using is pneumatic-free. The reason is not that electric grips harder. It is that a plant spends roughly a tenth of its electricity making air, and most of that never reaches the tool.</p>
            <a href="/tools/matchmatrix/index.html" class="btn btn-primary">Compare gripper specifications</a>
            <a href="/gripper-catalog/index.html" class="btn btn-secondary">Browse the gripper index</a>
        </div>
    </section>
<style>
.hero {
    min-height: 70vh;
    display: flex;
    align-items: center;
    justify-content: center;
    text-align: center;
    padding: 4rem 2rem;
    position: relative;

    /* Dark section context */
    --section-text: color-mix(in srgb, var(--hero-ink) 95%, transparent);
    --section-text-muted: color-mix(in srgb, var(--hero-ink) 80%, transparent);
    --section-heading: var(--hero-ink);
    --section-surface: color-mix(in srgb, var(--hero-ink) 10%, transparent);
    --section-border: color-mix(in srgb, var(--hero-ink) 30%, transparent);
}
.hero-content {
    max-width: 900px;
    margin: 0 auto;
    color: var(--hero-ink);
    z-index: 1;
}
.hero h1 {
    font-size: clamp(2rem, 5vw, 3.5rem);
    font-weight: 700;
    margin-bottom: 1.5rem;
    line-height: 1.2;
    text-shadow: 0 2px 4px rgba(0,0,0,0.3);
}
.hero-subheadline {
    font-size: clamp(1rem, 2vw, 1.35rem);
    margin-bottom: 2rem;
    line-height: 1.6;
}
.hero .btn {
    display: inline-block;
    padding: 0.875rem 2rem;
    margin: 0.5rem;
    border-radius: 4px;
    text-decoration: none;
    font-weight: 600;
    font-size: 1rem;
    transition: all 0.2s ease;
}
.hero .btn-primary {
    background: var(--hero-ink);
    /* The inverted hero button's text must contrast with --hero-ink, not with
       the page. In the no-image branch --hero-ink IS --color-primary-text, so
       --color-primary is correct by construction and stays the fallback. The
       image branch hard-codes --hero-ink: #fff, where a light --color-primary
       is unreadable, so that branch supplies --hero-btn-ink alongside it. */
    color: var(--hero-btn-ink, var(--color-primary));
    border: 2px solid var(--hero-ink);
}
.hero .btn-primary:hover {
    background: transparent;
    color: var(--hero-ink);
}
.hero .btn-secondary {
    background: transparent;
    color: var(--hero-ink);
    border: 2px solid color-mix(in srgb, var(--hero-ink) 80%, transparent);
}
.hero .btn-secondary:hover {
    background: color-mix(in srgb, var(--hero-ink) 10%, transparent);
}
@media (max-width: 768px) {
    .hero {
        min-height: 60vh;
        padding: 3rem 1.5rem;
    }
    .hero .btn {
        display: block;
        width: 100%;
        max-width: 280px;
        margin: 0.5rem auto;
    }
}
</style>

$r1$,
       'deployed', now(), 'news_editorial_features-lane', 'permanent' FROM _pg pg;
INSERT INTO page_components (page_id, component_id, position, slot_name, content_data, rendered_html,
                             build_status, locked_at, locked_by, lock_type)
SELECT pg.id, '8d81e665-3ee0-443d-a873-690268c15fbb', 2, 'feature-analysis', $c2${
 "ComponentID": "feature-analysis",
 "heading": "Four launches, one argument",
 "content": "<p>Four end-of-arm launches reached our feed inside a week, and they share a word. DESTACO's eRDH parallel gripper is presented as pneumatic-free. AWR standardised CNC machine tending on OnRobot electric grippers, with changeover speed as the stated reason. OnRobot brought out electric grippers aimed at high-payload work that used to be pneumatic's territory. And the humanoids being trialled at BMW plants are running electric grippers.</p>\n<p>Read as product news, that is four companies launching similar things. Read against the plant's electricity bill, it is one argument, and it was made by the US Department of Energy long before any of these products existed.</p>\n<p>Compressed air feels free at the tool because the pipe is already there. It is not. A Department of Energy survey found that in a typical industrial facility about 10% of all electricity consumed goes on generating compressed air, and that at some plants the figure is 30% or more. That is the cost of having the utility at all, before any question of which gripper you bolt on the end.</p>\n<p>The efficiency is the sharper number. The same source puts the overall efficiency of a typical compressed air system as low as 10%, and gives a worked example that needs no interpretation: to operate a 1 hp air motor at 100 psig, approximately 7 to 8 hp of electrical power is supplied to the compressor. Seven or eight units in, one unit of work out. The rest leaves as heat and leaks.</p>\n<p>An electric gripper does not have to be better to win that comparison. It only has to avoid it. This is why the launches cluster around changeover time and energy rather than grip force \u2014 the pitch is not that the jaw closes harder, it is that the line no longer pays the air tax to close it.</p>\n<p>The honest counter-argument is that none of this makes pneumatics obsolete, and our own index reflects that. Of the ten gripper models we hold specifications for, the electric parallel-jaw family is the largest single group at five, and pneumatic parallel-jaw, vacuum, magnetic, soft-robotic and adhesive each hold a place. Vacuum in particular is not displaced by any of this: it is doing a different job. What the numbers above explain is not the disappearance of compressed air, but why the default is moving \u2014 and why a specification exercise that ignores running cost will keep choosing the option that looks cheaper on the purchase order.</p>"
}$c2$::jsonb, $r2$<section id="feature-analysis" class="section section--generic">
  <div class="container">
    <h2 class="section__title">Four launches, one argument</h2>
    <div class="section__content"><p>Four end-of-arm launches reached our feed inside a week, and they share a word. DESTACO's eRDH parallel gripper is presented as pneumatic-free. AWR standardised CNC machine tending on OnRobot electric grippers, with changeover speed as the stated reason. OnRobot brought out electric grippers aimed at high-payload work that used to be pneumatic's territory. And the humanoids being trialled at BMW plants are running electric grippers.</p>
<p>Read as product news, that is four companies launching similar things. Read against the plant's electricity bill, it is one argument, and it was made by the US Department of Energy long before any of these products existed.</p>
<p>Compressed air feels free at the tool because the pipe is already there. It is not. A Department of Energy survey found that in a typical industrial facility about 10% of all electricity consumed goes on generating compressed air, and that at some plants the figure is 30% or more. That is the cost of having the utility at all, before any question of which gripper you bolt on the end.</p>
<p>The efficiency is the sharper number. The same source puts the overall efficiency of a typical compressed air system as low as 10%, and gives a worked example that needs no interpretation: to operate a 1 hp air motor at 100 psig, approximately 7 to 8 hp of electrical power is supplied to the compressor. Seven or eight units in, one unit of work out. The rest leaves as heat and leaks.</p>
<p>An electric gripper does not have to be better to win that comparison. It only has to avoid it. This is why the launches cluster around changeover time and energy rather than grip force — the pitch is not that the jaw closes harder, it is that the line no longer pays the air tax to close it.</p>
<p>The honest counter-argument is that none of this makes pneumatics obsolete, and our own index reflects that. Of the ten gripper models we hold specifications for, the electric parallel-jaw family is the largest single group at five, and pneumatic parallel-jaw, vacuum, magnetic, soft-robotic and adhesive each hold a place. Vacuum in particular is not displaced by any of this: it is doing a different job. What the numbers above explain is not the disappearance of compressed air, but why the default is moving — and why a specification exercise that ignores running cost will keep choosing the option that looks cheaper on the purchase order.</p></div>
  </div>
</section>
$r2$,
       'deployed', now(), 'news_editorial_features-lane', 'permanent' FROM _pg pg;
INSERT INTO page_components (page_id, component_id, position, slot_name, content_data, rendered_html,
                             build_status, locked_at, locked_by, lock_type)
SELECT pg.id, 'f8c2393c-fc66-480d-93f7-738c68cd5f1b', 3, 'evidence-chart-air', $c3${
 "section_eyebrow": "The energy arithmetic",
 "section_title": "What compressed air costs before it reaches the gripper",
 "section_intro": "Two ways of stating the same inefficiency. The first is what the utility takes from the plant; the second is the Department of Energy's own worked example of what arrives at the tool.",
 "charts": [
  {
   "id": "air-share",
   "title": "Share of plant electricity spent generating compressed air",
   "caption": "Scale: the high-end figure, so the typical case reads against it.",
   "unit": "%",
   "max_fact_id": "rh-air-share-high",
   "points": [
    {
     "fact_id": "rh-air-share-high",
     "label": "At some facilities",
     "tone": "accent"
    },
    {
     "fact_id": "rh-air-share-typical",
     "label": "Typical facility"
    },
    {
     "fact_id": "rh-air-efficiency",
     "label": "System efficiency (low end)",
     "tone": "muted"
    }
   ],
   "source_note": "US DOE via ENERGY STAR, \u201cDetermine the Cost of Compressed Air for Your Plant\u201d. Efficiency is shown on the same scale for comparison; it is a different measure, not a share of electricity."
  },
  {
   "id": "air-hp",
   "title": "Electrical horsepower in, air-motor horsepower out (100 psig)",
   "caption": "The DOE's worked example, at the low end of its stated range.",
   "unit": " hp",
   "max_fact_id": "rh-air-hp-in",
   "points": [
    {
     "fact_id": "rh-air-hp-in",
     "label": "Electrical power supplied",
     "tone": "accent"
    },
    {
     "fact_id": "rh-air-hp-out",
     "label": "Delivered at the air motor"
    }
   ],
   "source_note": "\u201cTo operate a 1 hp air motor at 100 psig, approximately 7-8 hp of electrical power is supplied to the air compressor.\u201d"
  },
  {
   "id": "catalogue",
   "title": "What our own index actually holds",
   "caption": "The counter-argument, in our own catalogue.",
   "unit": "",
   "max_fact_id": "rh-grippers",
   "points": [
    {
     "fact_id": "rh-grippers",
     "label": "Gripper models indexed",
     "tone": "accent"
    },
    {
     "fact_id": "rh-actuation",
     "label": "Actuation technologies"
    }
   ],
   "source_note": "Live counts from this site's own product index; electric parallel-jaw is the largest single group at five of the ten."
  }
 ],
 "facts": [
  {
   "id": "rh-air-share-high",
   "kind": "metric",
   "claim": "Share of electricity compressed air may account for at some facilities",
   "value": 30,
   "unit": "percent",
   "tolerance": "exact",
   "context_terms": [
    "compressed air",
    "some facilities"
   ],
   "writer_line": "at some plants it is {value}% or more",
   "verified_at": "2026-08-21",
   "source": {
    "citation": {
     "publisher": "ENERGY STAR / US Department of Energy",
     "title": "Determine the Cost of Compressed Air for Your Plant",
     "url": "https://www.energystar.gov/sites/default/files/buildings/tools/compressed_air1.pdf",
     "quote": "For some facilities, compressed air generation may account for 30% or more of the electricity consumed.",
     "accessed": "2026-08-21"
    }
   }
  },
  {
   "id": "rh-air-share-typical",
   "kind": "metric",
   "claim": "Share of a typical industrial facility's electricity consumed generating compressed air",
   "value": 10,
   "unit": "percent",
   "tolerance": "exact",
   "context_terms": [
    "compressed air",
    "electricity"
   ],
   "writer_line": "about {value}% of a typical plant's electricity goes on generating compressed air",
   "verified_at": "2026-08-21",
   "source": {
    "citation": {
     "publisher": "ENERGY STAR / US Department of Energy",
     "title": "Determine the Cost of Compressed Air for Your Plant",
     "url": "https://www.energystar.gov/sites/default/files/buildings/tools/compressed_air1.pdf",
     "quote": "A recent survey by the U.S. Department of Energy showed that for a typical industrial facility, approximately 10% of the electricity consumed is for generating compressed air.",
     "accessed": "2026-08-21"
    }
   }
  },
  {
   "id": "rh-air-efficiency",
   "kind": "metric",
   "claim": "Overall efficiency of a typical compressed air system, at the low end of the stated range",
   "value": 10,
   "unit": "percent",
   "tolerance": "exact",
   "context_terms": [
    "efficiency",
    "compressed air system"
   ],
   "writer_line": "overall efficiency can be as low as {value}%",
   "verified_at": "2026-08-21",
   "source": {
    "citation": {
     "publisher": "ENERGY STAR / US Department of Energy",
     "title": "Determine the Cost of Compressed Air for Your Plant",
     "url": "https://www.energystar.gov/sites/default/files/buildings/tools/compressed_air1.pdf",
     "quote": "The overall efficiency of a typical compressed air system can be as low as 10-15%.",
     "accessed": "2026-08-21"
    }
   }
  },
  {
   "id": "rh-air-hp-in",
   "kind": "metric",
   "claim": "Electrical horsepower supplied to the compressor to operate a 1 hp air motor at 100 psig, at the low end of the stated range",
   "value": 7,
   "unit": "hp",
   "tolerance": "exact",
   "context_terms": [
    "air motor",
    "horsepower",
    "compressor"
   ],
   "writer_line": "{value}-8 hp of electrical power to run a 1 hp air motor",
   "verified_at": "2026-08-21",
   "source": {
    "citation": {
     "publisher": "ENERGY STAR / US Department of Energy",
     "title": "Determine the Cost of Compressed Air for Your Plant",
     "url": "https://www.energystar.gov/sites/default/files/buildings/tools/compressed_air1.pdf",
     "quote": "For example, to operate a 1 hp air motor at 100 psig, approximately 7-8 hp of electrical power is supplied to the air compressor.",
     "accessed": "2026-08-21"
    }
   }
  },
  {
   "id": "rh-air-hp-out",
   "kind": "metric",
   "claim": "Air-motor horsepower delivered for that electrical input, per the same worked example",
   "value": 1,
   "unit": "hp",
   "tolerance": "exact",
   "context_terms": [
    "air motor",
    "horsepower"
   ],
   "writer_line": "delivering {value} hp at the tool",
   "verified_at": "2026-08-21",
   "source": {
    "citation": {
     "publisher": "ENERGY STAR / US Department of Energy",
     "title": "Determine the Cost of Compressed Air for Your Plant",
     "url": "https://www.energystar.gov/sites/default/files/buildings/tools/compressed_air1.pdf",
     "quote": "For example, to operate a 1 hp air motor at 100 psig, approximately 7-8 hp of electrical power is supplied to the air compressor.",
     "accessed": "2026-08-21"
    }
   }
  },
  {
   "id": "rh-grippers",
   "kind": "count",
   "claim": "gripper models indexed",
   "value": 10,
   "tolerance": "exact",
   "context_terms": [
    "gripper model",
    "models indexed",
    "grippers indexed"
   ],
   "writer_line": "{value} gripper models indexed",
   "verified_at": "2026-08-19",
   "source": {
    "sql": "SELECT count(*) FROM products WHERE category='gripper' AND status='active'"
   }
  },
  {
   "id": "rh-actuation",
   "kind": "count",
   "claim": "actuation technologies represented",
   "value": 6,
   "tolerance": "exact",
   "context_terms": [
    "actuation",
    "technolog"
   ],
   "writer_line": "{value} actuation technologies represented in the index: electric parallel-jaw (5 models), and one each of pneumatic parallel-jaw, vacuum, magnetic, soft-robotic, adhesive",
   "verified_at": "2026-07-26",
   "source": {
    "attested_by": "robot-hands R7 catalogue expansion, 2026-07-22"
   }
  }
 ]
}$c3$::jsonb, $r3$<style>
  /* Colour vocabulary checked against the live css_themes rather than assumed:
     --color-background / --color-text / --color-text-muted / --color-primary /
     --color-secondary / --color-accent / --color-card-bg / --color-border /
     --border-radius are the names the themes actually define (including the
     dark one). --color-surface, --spacing-section and --container-max-width are
     defined by NO theme, so anything resting on them renders its fallback on
     every site — which is how a light card ends up on a dark page. Where no
     variable exists, the fallback is a neutral translucent grey that reads
     correctly on light and dark alike, never a light literal. */
  .evidence-chart {
    padding: var(--spacing-xl, 4.5rem) var(--spacing-lg, 2rem);
    background: var(--color-background, #ffffff);
    color: var(--color-text, #1a1a1a);
  }
  .evidence-chart__inner {
    max-width: var(--container-max-width, 1200px);
    margin: 0 auto;
  }
  .evidence-chart__header {
    max-width: 62ch;
    margin: 0 0 2.75rem;
  }
  .evidence-chart__eyebrow {
    display: block;
    font-size: 0.8125rem;
    font-weight: 600;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--color-primary-ink, var(--color-primary, #1e40af));
  }
  .evidence-chart__title {
    font-size: clamp(1.5rem, 2.6vw, 2.1rem);
    font-weight: 700;
    line-height: 1.25;
    margin: 0.5rem 0 0;
  }
  .evidence-chart__intro {
    margin: 0.75rem 0 0;
    line-height: 1.6;
    color: var(--color-text-muted, #555);
  }
  .evidence-chart__grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
    gap: 2.5rem;
  }
  .evidence-chart__figure {
    margin: 0;
    padding: 1.5rem 1.5rem 1.25rem;
    background: var(--color-card-bg, rgba(127, 127, 127, 0.08));
    border: 1px solid var(--color-border, rgba(127, 127, 127, 0.28));
    border-radius: var(--border-radius, 10px);
  }
  .evidence-chart__figcaption {
    display: block;
    margin: 0 0 1.25rem;
  }
  .evidence-chart__chart-title {
    display: block;
    font-size: 1.0625rem;
    font-weight: 700;
    line-height: 1.35;
  }
  .evidence-chart__chart-note {
    display: block;
    margin-top: 0.35rem;
    font-size: 0.875rem;
    line-height: 1.5;
    color: var(--color-text-muted, #555);
  }
  .evidence-chart__row {
    display: grid;
    grid-template-columns: minmax(9ch, 30%) 1fr auto;
    align-items: center;
    gap: 0.75rem;
    padding: 0.55rem 0;
  }
  .evidence-chart__row + .evidence-chart__row {
    border-top: 1px solid var(--color-border, rgba(127, 127, 127, 0.22));
  }
  .evidence-chart__label {
    font-size: 0.9375rem;
    line-height: 1.35;
  }
  .evidence-chart__track {
    display: block;
    height: 1.35rem;
    border-radius: 3px;
    background: rgba(127, 127, 127, 0.22);
    overflow: hidden;
  }
  /* Geometry is computed by the browser from the real value: --v is the
     figure itself and --m the chart's declared maximum. Nothing rounds it
     into the template, and there is no width to get wrong by hand. */
  .evidence-chart__bar {
    display: block;
    height: 100%;
    width: calc(100% * var(--v, 0) / var(--m, 1));
    min-width: 2px;
    border-radius: 3px;
    background: var(--color-primary, #1e40af);
  }
  .evidence-chart__bar--muted {
    background: var(--color-secondary, rgba(127, 127, 127, 0.6));
  }
  .evidence-chart__bar--accent {
    background: var(--color-accent, var(--color-secondary, #0f766e));
  }
  .evidence-chart__value {
    font-size: 1.0625rem;
    font-weight: 700;
    font-variant-numeric: tabular-nums;
    white-space: nowrap;
  }
  .evidence-chart__verified {
    grid-column: 1 / -1;
    margin: 0;
    font-size: 0.75rem;
    color: var(--color-text-muted, #666);
  }
  .evidence-chart__source {
    margin: 1.1rem 0 0;
    padding-top: 0.85rem;
    border-top: 1px solid var(--color-border, rgba(127, 127, 127, 0.28));
    font-size: 0.8125rem;
    line-height: 1.5;
    color: var(--color-text-muted, #555);
  }
  @media (max-width: 620px) {
    .evidence-chart { padding: 3.25rem 1.25rem; }
    .evidence-chart__row { grid-template-columns: 1fr auto; }
    .evidence-chart__track { grid-column: 1 / -1; order: 3; }
  }
</style><section class="evidence-chart" data-component="evidence-chart">
  <div class="evidence-chart__inner">
    
    <header class="evidence-chart__header">
      <span class="evidence-chart__eyebrow">The energy arithmetic</span>
      <h2 class="evidence-chart__title">What compressed air costs before it reaches the gripper</h2>
      <p class="evidence-chart__intro">Two ways of stating the same inefficiency. The first is what the utility takes from the plant; the second is the Department of Energy's own worked example of what arrives at the tool.</p>
    </header>
    
    <div class="evidence-chart__grid">
      <figure class="evidence-chart__figure" data-chart="air-share">
        <figcaption class="evidence-chart__figcaption">
          <span class="evidence-chart__chart-title">Share of plant electricity spent generating compressed air</span>
          <span class="evidence-chart__chart-note">Scale: the high-end figure, so the typical case reads against it.</span>
        </figcaption>
        
        <div class="evidence-chart__row">
          <span class="evidence-chart__label">At some facilities</span>
          <span class="evidence-chart__track" aria-hidden="true"><span class="evidence-chart__bar evidence-chart__bar--accent" style="--v:30.0000;--m:30.0000"></span></span>
          <span class="evidence-chart__value">30%</span>
          <span class="evidence-chart__verified">verified 2026-08-21</span>
        </div>
        <div class="evidence-chart__row">
          <span class="evidence-chart__label">Typical facility</span>
          <span class="evidence-chart__track" aria-hidden="true"><span class="evidence-chart__bar" style="--v:10.0000;--m:30.0000"></span></span>
          <span class="evidence-chart__value">10%</span>
          <span class="evidence-chart__verified">verified 2026-08-21</span>
        </div>
        <div class="evidence-chart__row">
          <span class="evidence-chart__label">System efficiency (low end)</span>
          <span class="evidence-chart__track" aria-hidden="true"><span class="evidence-chart__bar evidence-chart__bar--muted" style="--v:10.0000;--m:30.0000"></span></span>
          <span class="evidence-chart__value">10%</span>
          <span class="evidence-chart__verified">verified 2026-08-21</span>
        </div>
        <p class="evidence-chart__source">US DOE via ENERGY STAR, “Determine the Cost of Compressed Air for Your Plant”. Efficiency is shown on the same scale for comparison; it is a different measure, not a share of electricity.</p>
      </figure>
      
      <figure class="evidence-chart__figure" data-chart="air-hp">
        <figcaption class="evidence-chart__figcaption">
          <span class="evidence-chart__chart-title">Electrical horsepower in, air-motor horsepower out (100 psig)</span>
          <span class="evidence-chart__chart-note">The DOE's worked example, at the low end of its stated range.</span>
        </figcaption>
        
        <div class="evidence-chart__row">
          <span class="evidence-chart__label">Electrical power supplied</span>
          <span class="evidence-chart__track" aria-hidden="true"><span class="evidence-chart__bar evidence-chart__bar--accent" style="--v:7.0000;--m:7.0000"></span></span>
          <span class="evidence-chart__value">7 hp</span>
          <span class="evidence-chart__verified">verified 2026-08-21</span>
        </div>
        <div class="evidence-chart__row">
          <span class="evidence-chart__label">Delivered at the air motor</span>
          <span class="evidence-chart__track" aria-hidden="true"><span class="evidence-chart__bar" style="--v:1.0000;--m:7.0000"></span></span>
          <span class="evidence-chart__value">1 hp</span>
          <span class="evidence-chart__verified">verified 2026-08-21</span>
        </div>
        <p class="evidence-chart__source">“To operate a 1 hp air motor at 100 psig, approximately 7-8 hp of electrical power is supplied to the air compressor.”</p>
      </figure>
      
      <figure class="evidence-chart__figure" data-chart="catalogue">
        <figcaption class="evidence-chart__figcaption">
          <span class="evidence-chart__chart-title">What our own index actually holds</span>
          <span class="evidence-chart__chart-note">The counter-argument, in our own catalogue.</span>
        </figcaption>
        
        <div class="evidence-chart__row">
          <span class="evidence-chart__label">Gripper models indexed</span>
          <span class="evidence-chart__track" aria-hidden="true"><span class="evidence-chart__bar evidence-chart__bar--accent" style="--v:10.0000;--m:10.0000"></span></span>
          <span class="evidence-chart__value">10</span>
          <span class="evidence-chart__verified">verified 2026-08-19</span>
        </div>
        <div class="evidence-chart__row">
          <span class="evidence-chart__label">Actuation technologies</span>
          <span class="evidence-chart__track" aria-hidden="true"><span class="evidence-chart__bar" style="--v:6.0000;--m:10.0000"></span></span>
          <span class="evidence-chart__value">6</span>
          <span class="evidence-chart__verified">verified 2026-07-26</span>
        </div>
        <p class="evidence-chart__source">Live counts from this site's own product index; electric parallel-jaw is the largest single group at five of the ten.</p>
      </figure>
      
      
    </div>
  </div>
</section>

$r3$,
       'deployed', now(), 'news_editorial_features-lane', 'permanent' FROM _pg pg;
INSERT INTO page_components (page_id, component_id, position, slot_name, content_data, rendered_html,
                             build_status, locked_at, locked_by, lock_type)
SELECT pg.id, '8d81e665-3ee0-443d-a873-690268c15fbb', 4, 'feature-coverage', $c4${
 "ComponentID": "feature-coverage",
 "heading": "The coverage this feature draws on",
 "content": "<p>Assembled from items our own news feed ingested across several channels in one week. Titles link to the original publishers; we quote headlines and link out.</p>\n<ul>\n<li><strong>DESTACO eRDH Electric Parallel Gripper</strong> &mdash; the launch that states pneumatic-free as the headline property</li>\n<li><strong>AWR standardizes CNC machine tending with OnRobot electric grippers to speed changeovers</strong> &mdash; the same shift seen from the integrator's side, with changeover time as the reason</li>\n<li><strong>OnRobot to Debut two Electrical Grippers with Unmatched Power for High-Payload Applications</strong> &mdash; electric moving into payload ranges pneumatics used to own</li>\n<li><strong>Humanoid Robots at BMW Plants Feature Electric Grippers from China</strong> &mdash; the same choice being made on a very different platform</li>\n</ul>\n<p>The energy figures are from the US Department of Energy via ENERGY STAR, cited beneath the chart.</p>"
}$c4$::jsonb, $r4$<section id="feature-coverage" class="section section--generic">
  <div class="container">
    <h2 class="section__title">The coverage this feature draws on</h2>
    <div class="section__content"><p>Assembled from items our own news feed ingested across several channels in one week. Titles link to the original publishers; we quote headlines and link out.</p>
<ul>
<li><strong>DESTACO eRDH Electric Parallel Gripper</strong> &mdash; the launch that states pneumatic-free as the headline property</li>
<li><strong>AWR standardizes CNC machine tending with OnRobot electric grippers to speed changeovers</strong> &mdash; the same shift seen from the integrator's side, with changeover time as the reason</li>
<li><strong>OnRobot to Debut two Electrical Grippers with Unmatched Power for High-Payload Applications</strong> &mdash; electric moving into payload ranges pneumatics used to own</li>
<li><strong>Humanoid Robots at BMW Plants Feature Electric Grippers from China</strong> &mdash; the same choice being made on a very different platform</li>
</ul>
<p>The energy figures are from the US Department of Energy via ENERGY STAR, cited beneath the chart.</p></div>
  </div>
</section>
$r4$,
       'deployed', now(), 'news_editorial_features-lane', 'permanent' FROM _pg pg;
INSERT INTO page_components (page_id, component_id, position, slot_name, content_data, rendered_html,
                             build_status, locked_at, locked_by, lock_type)
SELECT pg.id, '0197e8d7-1adc-43d6-ab32-d0716e013175', 5, 'call-to-action', $c5${
 "headline": "Specify on running cost, not just grip force.",
 "subheadline": "MatchMatrix tests your application against gripping force, jaw travel, rated payload and IP rating across the actuation technologies in our index. It will not price your compressed air for you, but it will show you which electric options actually meet the specification you would otherwise fill pneumatically.",
 "primary_cta": "Run MatchMatrix",
 "primary_cta_url": "/tools/matchmatrix/index.html",
 "primary_cta_target_title": "Run MatchMatrix | Gripper Selection Tool | Robot-Hands.com",
 "secondary_cta": "Compare pneumatic and electric",
 "secondary_cta_url": "/pneumatic-vs-electric-grippers.html",
 "secondary_cta_target_title": "Pneumatic vs Electric Grippers | Robot-Hands.com"
}$c5$::jsonb, $r5$<section class="cta-section" data-component="call-to-action">
    <div class="cta-container">
        <h2>Specify on running cost, not just grip force.</h2>
        <p class="cta-subtitle">MatchMatrix tests your application against gripping force, jaw travel, rated payload and IP rating across the actuation technologies in our index. It will not price your compressed air for you, but it will show you which electric options actually meet the specification you would otherwise fill pneumatically.</p>
        <div class="cta-buttons">
            
            <a href="/tools/matchmatrix/index.html" class="cta-btn cta-btn-primary">Run MatchMatrix</a>
            
            
            <a href="/pneumatic-vs-electric-grippers.html" class="cta-btn cta-btn-secondary">Compare pneumatic and electric</a>
            
        </div>
    </div>
</section>
<style>
.cta-section {
    padding: var(--spacing-section, 5rem 2rem);
    background: var(--color-cta-bg, var(--color-primary));
    color: var(--color-cta-text, var(--color-primary-text));
    text-align: center;

    /* Dark section context */
    --section-text: color-mix(in srgb, var(--color-cta-text, var(--color-primary-text)) 90%, transparent);
    --section-text-muted: color-mix(in srgb, var(--color-cta-text, var(--color-primary-text)) 85%, transparent);
    --section-heading: var(--color-cta-text, var(--color-primary-text));
    --section-surface: color-mix(in srgb, var(--color-cta-text, var(--color-primary-text)) 5%, transparent);
    --section-border: color-mix(in srgb, var(--color-cta-text, var(--color-primary-text)) 20%, transparent);
}
.cta-container {
    max-width: 800px;
    margin: 0 auto;
}
.cta-section h2 {
    margin-bottom: 1rem;
}
.cta-subtitle {
    margin-bottom: 2rem;
}
.cta-buttons {
    display: flex;
    gap: 1rem;
    justify-content: center;
    flex-wrap: wrap;
}
.cta-btn {
    display: inline-block;
    padding: 1rem 2rem;
    border-radius: 6px;
    text-decoration: none;
    font-weight: 600;
    transition: transform 0.2s, box-shadow 0.2s;
}
.cta-btn:hover {
    transform: translateY(-2px);
}
.cta-btn-primary {
    background: var(--color-cta-text, var(--color-primary-text));
    color: var(--color-cta-bg, var(--color-primary));
}
.cta-btn-secondary {
    background: transparent;
    border: 2px solid var(--color-cta-text, var(--color-primary-text));
    color: var(--color-cta-text, var(--color-primary-text));
}
@media (max-width: 768px) {
    .cta-section { padding: 3rem 1.5rem; }
    .cta-buttons { flex-direction: column; align-items: center; }
    .cta-btn { width: 100%; max-width: 280px; text-align: center; }
}
</style>
$r5$,
       'deployed', now(), 'news_editorial_features-lane', 'permanent' FROM _pg pg;

DO $$
DECLARE n int; short int; mismatch int; mdlen int;
BEGIN
  SELECT count(*), count(*) FILTER (WHERE coalesce(length(pc.rendered_html),0) < 500)
    INTO n, short FROM page_components pc JOIN _pg ON _pg.id=pc.page_id;
  IF n <> 5 THEN RAISE EXCEPTION 'expected 5 components, got %', n; END IF;
  IF short > 0 THEN RAISE EXCEPTION '% short component(s)', short; END IF;
  SELECT count(*) INTO mismatch FROM pages p, jsonb_array_elements_text(p.sections) s
   WHERE p.site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND p.name='electric-vs-pneumatic-economics'
     AND NOT EXISTS (SELECT 1 FROM page_components pc WHERE pc.page_id=p.id AND pc.slot_name=s);
  IF mismatch > 0 THEN RAISE EXCEPTION '% unmatched sections entry(ies)', mismatch; END IF;
  -- hero must be on the image+overlay branch (owner ruling 2026-08-20)
  IF NOT EXISTS (SELECT 1 FROM page_components pc JOIN _pg ON _pg.id=pc.page_id
    WHERE pc.slot_name='hero' AND pc.rendered_html LIKE '%content-hero-electric-vs-pneumatic-economics.jpg%'
      AND pc.rendered_html LIKE '%rgba(0,0,0,0.5)%') THEN
    RAISE EXCEPTION 'hero is not on the image+overlay branch';
  END IF;
  -- charts must carry the legible-ink token (496), not the raw palette colour
  IF NOT EXISTS (SELECT 1 FROM page_components pc JOIN _pg ON _pg.id=pc.page_id
    WHERE pc.slot_name='evidence-chart-air' AND pc.rendered_html LIKE '%ink, var(--color-%') THEN
    RAISE EXCEPTION 'chart section did not inherit the legible-ink repoint';
  END IF;
  -- meta_description must be reader-facing (misstep 6): under 160, no brief shapes
  SELECT length(meta_description) INTO mdlen FROM pages
   WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND name='electric-vs-pneumatic-economics';
  IF mdlen > 160 THEN RAISE EXCEPTION 'meta_description is % chars, over the 160 limit', mdlen; END IF;
END $$;

COMMIT;
