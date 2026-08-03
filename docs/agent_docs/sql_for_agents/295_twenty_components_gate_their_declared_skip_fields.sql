-- 295_twenty_components_gate_their_declared_skip_fields.sql — 2026-08-03,
-- bugfix_140_contact_info_fabrication / RFC_009
--
-- CLOSES the one open item left by bugs_closed/140: 68 fields across 20 shared components
-- declare `"on_missing": "skip_field"`, are referenced by their template, and are GATED BY
-- NOTHING. When the datum is absent the element renders anyway, empty. Requested by the
-- owner 2026-08-03 ("gate the 68 blank-rendering fields").
--
-- THE ORGANISING FACT is 287's, unchanged: the schema already declares the correct
-- behaviour and the template disobeys it. This is not a policy choice between two
-- defensible renderings — it is making 20 templates obey a contract they already publish.
--
-- WHY THE TEMPLATES AND NOT THE RENDERER. RFC_009 option A (teach executeGoTemplate to
-- enforce on_missing) is DECIDED AGAINST by the owner, on a measurement: ~90% of fields
-- (1,938 of 2,163) declare no on_missing at all, so a render-time gate would be inert for
-- nine fields in ten while being the only option that can break a live page. This file is
-- the per-template repair that decision implies, for the 10% where the contract IS
-- declared and therefore checkable.
--
-- THE SHAPE OF THE DEFECT, which is why one rule does not fit all 68: 62 of the 68 are the
-- ungated PARTNER of a gated field —
--     {{if .spec_1_name}}<tr><th>{{.spec_1_name}}</th><td>{{.spec_1_value}}</td></tr>{{end}}
-- The row is gated on the NAME; the VALUE is not. So the element that must disappear
-- differs by component, and gating "the field" blindly would produce malformed HTML.
-- Four treatments, chosen per component:
--
--   1. GATE THE ELEMENT (30 fields) — a block element that is independently skippable:
--      a hero subheadline <p>, a headline <h2>, a stat label <div>, a trust label <span>.
--          <p …>{{.subheadline}}</p>  ->  {{if .subheadline}}<p …>{{.subheadline}}</p>{{end}}
--
--   2. WIDEN THE ROW GATE (23 fields) — product-specs and platform-comparison put these
--      fields in <td>s of fixed-arity rows. A <td> CANNOT be dropped: removing one from a
--      4-column comparison row misaligns every column after it, which is worse than the
--      blank. "Skip the field" can only mean "do not render the row this field is a cell
--      of", so the EXISTING row gate is widened rather than a second gate added:
--          {{if .row1_feature}}<tr>  ->  {{if and .row1_feature .row1_platform1_value …}}<tr>
--      An empty cell in a comparison table is not neutral either — it reads as "this
--      platform does not have this", which is nearer a false claim than a blank.
--
--   3. GATE THE CARD (6 fields) — Pricing Tiers 2 and 3 are `required:false,
--      on_missing:skip_field`: OPTIONAL tiers. A site with one tier should render ONE card,
--      not three of which two are empty. tier_1 is DELIBERATELY UNTOUCHED — it declares
--      `required:true, on_missing:needs_human_review`, an escalation policy, not a skip,
--      and another mechanism's business. featured_article's image wrapper and meta row are
--      gated on `or` of their contents so gating the children cannot leave an empty shell
--      behind (the container rule from bugs_open/111 that migration 287 also followed).
--
--   4. GATE THE SECTION (1 field) — about-commercial-block.domain is woven through six
--      prose branches in two languages ("The {{.domain}} name is available to acquire"),
--      where an inline gate would render "The  name is available to acquire". The whole
--      block asserts something ABOUT the domain, so the section is the honest skip unit.
--      It is also the ONE field of the 68 that RenderContext independently supplies
--      (ctx.Domain reaches the template contract via its json tag), so it is never actually
--      absent in practice — 0 blank of 2 live instances. That is RFC_009's hard question 2,
--      and measuring it is what kept this gate from being written on a false premise.
--
-- ONE OF THE 68 IS NOT A BLANK, and RFC_009 calling the whole class mild is too kind to it:
-- featured_article.featured_image sits bare in <img src="{{.featured_image}}">. Absent, it
-- renders src="", which a browser resolves to the page URL and re-requests — a BROKEN
-- IMAGE, the `inURLAttr` dead-control class of bugs_open/018. hero-tool's two CTA labels
-- are the same family: a present URL with an absent label renders <a href="…"></a>, an
-- invisible unclickable control. Both are now gated on url AND label — the idiom `hero` in
-- this same library already uses ({{if and .cta_text .cta_url}}), so the library is being
-- made to agree with itself.
--
-- BLAST RADIUS, measured live before writing, at the ARTEFACT and not merely in the data:
--   * 20 of the 68 fields have ZERO live instances (product-specs 8, featured_article 6,
--     archetype-result-card 3, bayesian-ranking-hero-tool 3) — those gates are purely
--     prophylactic and cannot change any page.
--   * Of the rest, the datum is absent from content_data in 75 field-instances, 47 of them
--     hero.subheadline. BUT the stored artefact tells a different and more honest story:
--     only THREE stored rows attributable to these 20 components carry the empty element
--     today (1 hero <p>, 1 empty <h2>, 1 Pricing Tiers card). The other 46 hero rows have
--     real subheadline text in their stored HTML from a legacy render and an EMPTY
--     content_data — their blank is latent, not served.
--     [The content_data count is the forward-looking risk; the artefact count is present
--     damage. Quoting the first as if it were the second would overstate this file 25x.]
--   * Nothing changes on any live page until that page rerenders WITH A REASON — a
--     reason-less rerender re-staples stored section HTML and will not pick this up
--     (the bugfix_140 landmine, and the mistake that thread made in the middle of 140).
--
-- PROVEN BEFORE SHIPPING, on the real path and not a replica of it: every one of the 20
-- gated templates was rendered through actions.RenderTemplate (-> contextToInterfaceMap ->
-- executeGoTemplate, the production entry point, same missingkey=zero and same FuncMap)
-- twice — once with the datum present, once with it absent. 20/20: the element vanishes
-- when absent, AND still renders when present. The positive control is the load-bearing
-- half: a gate that over-fires would pass a "does it disappear" test perfectly.
--
-- CONSUMERS TOLD (owner ruling 2026-07-29 §3 — measuring is not telling): these 20
-- components belong to other lanes. The change to their guarantee is: "a field you never
-- supplied, which your component's own schema says to skip, will now actually be skipped —
-- so an element may DISAPPEAR after your next reasoned rerender rather than render empty.
-- Nothing that has a value renders differently." Recorded in RFC_009's decision record.
--
-- WHAT IT DOES NOT DO, ON PURPOSE
--   * Does NOT touch the 1,938 fields that declare no on_missing — undeclared is not
--     disobeyed, and inventing a policy for them is RFC_009 option A, decided against.
--   * Does NOT patch any stored page_components row. The templates are the defect; the
--     rows correct themselves on their next reasoned rerender (287's rule).
--   * Does NOT change input_schema. The schema was right; the templates were wrong.
--   * Does NOT gate tier_1 (needs_human_review), nor the 55 fields that ARE gated
--     somewhere — a field gated in one place and used bare in another is invisible to the
--     lint's approximation and to this file. That gap is stated in RFC_009 and unchanged.
--
-- ROLLBACK
--   UPDATE content_components c SET html_template = b.old_value->>'html_template'
--     FROM migration_backups b
--    WHERE b.migration_name = '295_twenty_components_gate_their_declared_skip_fields.sql'
--      AND c.id::text = b.target_id;
--
-- VERIFY (after COMMIT, from a session):
--   python3 scripts/check_placeholder_fallbacks.py     -- expect: 0 fabricated, 0 ungated

BEGIN;

-- Before-image, so the rollback recipe above is executable rather than aspirational.
INSERT INTO migration_backups (migration_name, target_table, target_id, old_value, notes)
SELECT '295_twenty_components_gate_their_declared_skip_fields.sql',
       'content_components',
       id::text,
       jsonb_build_object('html_template', html_template, 'input_schema', input_schema),
       'RFC_009: pre-gate template — declared skip_field fields render an empty element'
  FROM content_components
 WHERE is_active AND name IN ('hero', 'about-hero', 'case-studies-hero', 'contact-hero', 'services-hero', 'use-cases-hero', 'testimonials', 'social_proof', 'portfolio-showcase', 'content-listing', 'system-stats', 'archetype-result-card', 'bayesian-ranking-hero-tool_pre_037', 'hero-tool', 'case-studies-grid', 'product-specs', 'platform-comparison', 'Pricing Tiers', 'featured_article', 'about-commercial-block');


-- hero — 1 field(s): subheadline
UPDATE content_components SET html_template =
  replace(html_template,
         $o0$<p class="hero-subheadline">{{.subheadline}}</p>$o0$,
         $n0${{if .subheadline}}<p class="hero-subheadline">{{.subheadline}}</p>{{end}}$n0$),
    updated_at = now()
 WHERE is_active AND name = 'hero';

-- about-hero — 1 field(s): subheadline
UPDATE content_components SET html_template =
  replace(html_template,
         $o0$<p class="hero-subheadline">{{.subheadline}}</p>$o0$,
         $n0${{if .subheadline}}<p class="hero-subheadline">{{.subheadline}}</p>{{end}}$n0$),
    updated_at = now()
 WHERE is_active AND name = 'about-hero';

-- case-studies-hero — 1 field(s): subheadline
UPDATE content_components SET html_template =
  replace(html_template,
         $o0$<p class="hero-subheadline">{{.subheadline}}</p>$o0$,
         $n0${{if .subheadline}}<p class="hero-subheadline">{{.subheadline}}</p>{{end}}$n0$),
    updated_at = now()
 WHERE is_active AND name = 'case-studies-hero';

-- contact-hero — 1 field(s): subheadline
UPDATE content_components SET html_template =
  replace(html_template,
         $o0$<p class="hero-subheadline">{{.subheadline}}</p>$o0$,
         $n0${{if .subheadline}}<p class="hero-subheadline">{{.subheadline}}</p>{{end}}$n0$),
    updated_at = now()
 WHERE is_active AND name = 'contact-hero';

-- services-hero — 1 field(s): subheadline
UPDATE content_components SET html_template =
  replace(html_template,
         $o0$<p class="hero-subheadline">{{.subheadline}}</p>$o0$,
         $n0${{if .subheadline}}<p class="hero-subheadline">{{.subheadline}}</p>{{end}}$n0$),
    updated_at = now()
 WHERE is_active AND name = 'services-hero';

-- use-cases-hero — 1 field(s): subheadline
UPDATE content_components SET html_template =
  replace(html_template,
         $o0$<p class="hero-subheadline">{{.subheadline}}</p>$o0$,
         $n0${{if .subheadline}}<p class="hero-subheadline">{{.subheadline}}</p>{{end}}$n0$),
    updated_at = now()
 WHERE is_active AND name = 'use-cases-hero';

-- testimonials — 1 field(s): headline
UPDATE content_components SET html_template =
  replace(html_template,
         $o0$<h2>{{.headline}}</h2>$o0$,
         $n0${{if .headline}}<h2>{{.headline}}</h2>{{end}}$n0$),
    updated_at = now()
 WHERE is_active AND name = 'testimonials';

-- social_proof — 1 field(s): headline
UPDATE content_components SET html_template =
  replace(html_template,
         $o0$<h2>{{.headline}}</h2>$o0$,
         $n0${{if .headline}}<h2>{{.headline}}</h2>{{end}}$n0$),
    updated_at = now()
 WHERE is_active AND name = 'social_proof';

-- portfolio-showcase — 1 field(s): headline
UPDATE content_components SET html_template =
  replace(html_template,
         $o0$<h2>{{.headline}}</h2>$o0$,
         $n0${{if .headline}}<h2>{{.headline}}</h2>{{end}}$n0$),
    updated_at = now()
 WHERE is_active AND name = 'portfolio-showcase';

-- content-listing — 1 field(s): section_title
UPDATE content_components SET html_template =
  replace(html_template,
         $o0$<h2 class="section__title">{{.section_title}}</h2>$o0$,
         $n0${{if .section_title}}<h2 class="section__title">{{.section_title}}</h2>{{end}}$n0$),
    updated_at = now()
 WHERE is_active AND name = 'content-listing';

-- system-stats — 8 field(s): stat1_description, stat1_label, stat2_description, stat2_label, stat3_description, stat3_label, stat4_description, stat4_label
UPDATE content_components SET html_template =
  replace(replace(replace(replace(replace(replace(replace(replace(html_template,
         $o0$<div class="stat-label">{{.stat1_label}}</div>$o0$,
         $n0${{if .stat1_label}}<div class="stat-label">{{.stat1_label}}</div>{{end}}$n0$),
         $o1$<p class="stat-description">{{.stat1_description}}</p>$o1$,
         $n1${{if .stat1_description}}<p class="stat-description">{{.stat1_description}}</p>{{end}}$n1$),
         $o2$<div class="stat-label">{{.stat2_label}}</div>$o2$,
         $n2${{if .stat2_label}}<div class="stat-label">{{.stat2_label}}</div>{{end}}$n2$),
         $o3$<p class="stat-description">{{.stat2_description}}</p>$o3$,
         $n3${{if .stat2_description}}<p class="stat-description">{{.stat2_description}}</p>{{end}}$n3$),
         $o4$<div class="stat-label">{{.stat3_label}}</div>$o4$,
         $n4${{if .stat3_label}}<div class="stat-label">{{.stat3_label}}</div>{{end}}$n4$),
         $o5$<p class="stat-description">{{.stat3_description}}</p>$o5$,
         $n5${{if .stat3_description}}<p class="stat-description">{{.stat3_description}}</p>{{end}}$n5$),
         $o6$<div class="stat-label">{{.stat4_label}}</div>$o6$,
         $n6${{if .stat4_label}}<div class="stat-label">{{.stat4_label}}</div>{{end}}$n6$),
         $o7$<p class="stat-description">{{.stat4_description}}</p>$o7$,
         $n7${{if .stat4_description}}<p class="stat-description">{{.stat4_description}}</p>{{end}}$n7$),
    updated_at = now()
 WHERE is_active AND name = 'system-stats';

-- archetype-result-card — 3 field(s): stat_1_label, stat_2_label, stat_3_label
UPDATE content_components SET html_template =
  replace(replace(replace(html_template,
         $o0$<span class="arc-stat-label">{{.stat_1_label}}</span>$o0$,
         $n0${{if .stat_1_label}}<span class="arc-stat-label">{{.stat_1_label}}</span>{{end}}$n0$),
         $o1$<span class="arc-stat-label">{{.stat_2_label}}</span>$o1$,
         $n1${{if .stat_2_label}}<span class="arc-stat-label">{{.stat_2_label}}</span>{{end}}$n1$),
         $o2$<span class="arc-stat-label">{{.stat_3_label}}</span>$o2$,
         $n2${{if .stat_3_label}}<span class="arc-stat-label">{{.stat_3_label}}</span>{{end}}$n2$),
    updated_at = now()
 WHERE is_active AND name = 'archetype-result-card';

-- bayesian-ranking-hero-tool_pre_037 — 3 field(s): stat_one_label, stat_three_label, stat_two_label
UPDATE content_components SET html_template =
  replace(replace(replace(html_template,
         $o0$<span class="brht-trust-label">{{.stat_one_label}}</span>$o0$,
         $n0${{if .stat_one_label}}<span class="brht-trust-label">{{.stat_one_label}}</span>{{end}}$n0$),
         $o1$<span class="brht-trust-label">{{.stat_two_label}}</span>$o1$,
         $n1${{if .stat_two_label}}<span class="brht-trust-label">{{.stat_two_label}}</span>{{end}}$n1$),
         $o2$<span class="brht-trust-label">{{.stat_three_label}}</span>$o2$,
         $n2${{if .stat_three_label}}<span class="brht-trust-label">{{.stat_three_label}}</span>{{end}}$n2$),
    updated_at = now()
 WHERE is_active AND name = 'bayesian-ranking-hero-tool_pre_037';

-- hero-tool — 5 field(s): cta_primary_label, cta_secondary_label, stat_one_label, stat_three_label, stat_two_label
UPDATE content_components SET html_template =
  replace(replace(replace(replace(replace(replace(html_template,
         $o0$<span class="htl-trust-label">{{.stat_one_label}}</span>$o0$,
         $n0${{if .stat_one_label}}<span class="htl-trust-label">{{.stat_one_label}}</span>{{end}}$n0$),
         $o1$<span class="htl-trust-label">{{.stat_two_label}}</span>$o1$,
         $n1${{if .stat_two_label}}<span class="htl-trust-label">{{.stat_two_label}}</span>{{end}}$n1$),
         $o2$<span class="htl-trust-label">{{.stat_three_label}}</span>$o2$,
         $n2${{if .stat_three_label}}<span class="htl-trust-label">{{.stat_three_label}}</span>{{end}}$n2$),
         $o3${{if or .cta_primary_url .cta_secondary_url}}<div class="htl-cta-row">$o3$,
         $n3${{if or (and .cta_primary_url .cta_primary_label) (and .cta_secondary_url .cta_secondary_label)}}<div class="htl-cta-row">$n3$),
         $o4${{if .cta_primary_url}}<a href="{{.cta_primary_url}}" class="htl-btn-primary">{{.cta_primary_label}}</a>{{end}}$o4$,
         $n4${{if and .cta_primary_url .cta_primary_label}}<a href="{{.cta_primary_url}}" class="htl-btn-primary">{{.cta_primary_label}}</a>{{end}}$n4$),
         $o5${{if .cta_secondary_url}}<a href="{{.cta_secondary_url}}" class="htl-btn-secondary">{{.cta_secondary_label}}</a>{{end}}$o5$,
         $n5${{if and .cta_secondary_url .cta_secondary_label}}<a href="{{.cta_secondary_url}}" class="htl-btn-secondary">{{.cta_secondary_label}}</a>{{end}}$n5$),
    updated_at = now()
 WHERE is_active AND name = 'hero-tool';

-- case-studies-grid — 5 field(s): card1_stat_label, card2_stat_label, card3_stat_label, card4_stat_label, card5_stat_label
UPDATE content_components SET html_template =
  replace(replace(replace(replace(replace(html_template,
         $o0$<strong>{{.card1_stat_value}}</strong> {{.card1_stat_label}}</span>$o0$,
         $n0$<strong>{{.card1_stat_value}}</strong>{{if .card1_stat_label}} {{.card1_stat_label}}{{end}}</span>$n0$),
         $o1$<strong>{{.card2_stat_value}}</strong> {{.card2_stat_label}}</span>$o1$,
         $n1$<strong>{{.card2_stat_value}}</strong>{{if .card2_stat_label}} {{.card2_stat_label}}{{end}}</span>$n1$),
         $o2$<strong>{{.card3_stat_value}}</strong> {{.card3_stat_label}}</span>$o2$,
         $n2$<strong>{{.card3_stat_value}}</strong>{{if .card3_stat_label}} {{.card3_stat_label}}{{end}}</span>$n2$),
         $o3$<strong>{{.card4_stat_value}}</strong> {{.card4_stat_label}}</span>$o3$,
         $n3$<strong>{{.card4_stat_value}}</strong>{{if .card4_stat_label}} {{.card4_stat_label}}{{end}}</span>$n3$),
         $o4$<strong>{{.card5_stat_value}}</strong> {{.card5_stat_label}}</span>$o4$,
         $n4$<strong>{{.card5_stat_value}}</strong>{{if .card5_stat_label}} {{.card5_stat_label}}{{end}}</span>$n4$),
    updated_at = now()
 WHERE is_active AND name = 'case-studies-grid';

-- product-specs — 8 field(s): spec_1_value, spec_2_value, spec_3_value, spec_4_value, spec_5_value, spec_6_value, spec_7_value, spec_8_value
UPDATE content_components SET html_template =
  replace(replace(replace(replace(replace(replace(replace(replace(html_template,
         $o0${{if .spec_1_name}}<tr>$o0$,
         $n0${{if and .spec_1_name .spec_1_value}}<tr>$n0$),
         $o1${{if .spec_2_name}}<tr>$o1$,
         $n1${{if and .spec_2_name .spec_2_value}}<tr>$n1$),
         $o2${{if .spec_3_name}}<tr>$o2$,
         $n2${{if and .spec_3_name .spec_3_value}}<tr>$n2$),
         $o3${{if .spec_4_name}}<tr>$o3$,
         $n3${{if and .spec_4_name .spec_4_value}}<tr>$n3$),
         $o4${{if .spec_5_name}}<tr>$o4$,
         $n4${{if and .spec_5_name .spec_5_value}}<tr>$n4$),
         $o5${{if .spec_6_name}}<tr>$o5$,
         $n5${{if and .spec_6_name .spec_6_value}}<tr>$n5$),
         $o6${{if .spec_7_name}}<tr>$o6$,
         $n6${{if and .spec_7_name .spec_7_value}}<tr>$n6$),
         $o7${{if .spec_8_name}}<tr>$o7$,
         $n7${{if and .spec_8_name .spec_8_value}}<tr>$n7$),
    updated_at = now()
 WHERE is_active AND name = 'product-specs';

-- platform-comparison — 15 field(s): row1_platform1_value, row1_platform2_value, row1_spark_value, row2_platform1_value, row2_platform2_value, row2_spark_value, row3_platform1_value, row3_platform2_value, row3_spark_value, row4_platform1_value, row4_platform2_value, row4_spark_value, row5_platform1_value, row5_platform2_value, row5_spark_value
UPDATE content_components SET html_template =
  replace(replace(replace(replace(replace(html_template,
         $o0${{if .row1_feature}}<tr>$o0$,
         $n0${{if and .row1_feature .row1_platform1_value .row1_platform2_value .row1_spark_value}}<tr>$n0$),
         $o1${{if .row2_feature}}<tr>$o1$,
         $n1${{if and .row2_feature .row2_platform1_value .row2_platform2_value .row2_spark_value}}<tr>$n1$),
         $o2${{if .row3_feature}}<tr>$o2$,
         $n2${{if and .row3_feature .row3_platform1_value .row3_platform2_value .row3_spark_value}}<tr>$n2$),
         $o3${{if .row4_feature}}<tr>$o3$,
         $n3${{if and .row4_feature .row4_platform1_value .row4_platform2_value .row4_spark_value}}<tr>$n3$),
         $o4${{if .row5_feature}}<tr>$o4$,
         $n4${{if and .row5_feature .row5_platform1_value .row5_platform2_value .row5_spark_value}}<tr>$n4$),
    updated_at = now()
 WHERE is_active AND name = 'platform-comparison';

-- Pricing Tiers — 4 field(s): tier_2_name, tier_2_price, tier_3_name, tier_3_price
UPDATE content_components SET html_template =
  replace(replace(replace(replace(html_template,
         $o0$      <div class="pricing-tier pricing-tier--featured card">
        <div class="pricing-tier__badge">Popular</div>
        <h3 class="pricing-tier__name">{{.tier_2_name}}</h3>$o0$,
         $n0$      {{if and .tier_2_name .tier_2_price}}<div class="pricing-tier pricing-tier--featured card">
        <div class="pricing-tier__badge">Popular</div>
        <h3 class="pricing-tier__name">{{.tier_2_name}}</h3>$n0$),
         $o1$        {{if .tier_2_cta_url}}<a href="{{.tier_2_cta_url}}" class="button button--primary button--full-width">{{.tier_2_cta}}</a>{{end}}
      </div>$o1$,
         $n1$        {{if .tier_2_cta_url}}<a href="{{.tier_2_cta_url}}" class="button button--primary button--full-width">{{.tier_2_cta}}</a>{{end}}
      </div>{{end}}$n1$),
         $o2$      <div class="pricing-tier card">
        <h3 class="pricing-tier__name">{{.tier_3_name}}</h3>$o2$,
         $n2$      {{if and .tier_3_name .tier_3_price}}<div class="pricing-tier card">
        <h3 class="pricing-tier__name">{{.tier_3_name}}</h3>$n2$),
         $o3$        {{if .tier_3_cta_url}}<a href="{{.tier_3_cta_url}}" class="button button--secondary button--full-width">{{.tier_3_cta}}</a>{{end}}
      </div>$o3$,
         $n3$        {{if .tier_3_cta_url}}<a href="{{.tier_3_cta_url}}" class="button button--secondary button--full-width">{{.tier_3_cta}}</a>{{end}}
      </div>{{end}}$n3$),
    updated_at = now()
 WHERE is_active AND name = 'Pricing Tiers';

-- featured_article — 6 field(s): featured_author, featured_category, featured_date, featured_excerpt, featured_image, featured_read_time
UPDATE content_components SET html_template =
  replace(replace(replace(html_template,
         $o0$          <div class="featured-article__image">
            <img src="{{.featured_image}}" alt="{{.featured_title}}" loading="lazy">
            <span class="featured-article__category">{{.featured_category}}</span>
          </div>$o0$,
         $n0$          {{if or .featured_image .featured_category}}<div class="featured-article__image">
            {{if .featured_image}}<img src="{{.featured_image}}" alt="{{.featured_title}}" loading="lazy">{{end}}
            {{if .featured_category}}<span class="featured-article__category">{{.featured_category}}</span>{{end}}
          </div>{{end}}$n0$),
         $o1$<p class="featured-article__excerpt">{{.featured_excerpt}}</p>$o1$,
         $n1${{if .featured_excerpt}}<p class="featured-article__excerpt">{{.featured_excerpt}}</p>{{end}}$n1$),
         $o2$            <div class="featured-article__meta">
              <span class="featured-article__author">{{.featured_author}}</span>
              <span class="featured-article__date">{{.featured_date}}</span>
              <span class="featured-article__read-time">{{.featured_read_time}}</span>
            </div>$o2$,
         $n2$            {{if or .featured_author .featured_date .featured_read_time}}<div class="featured-article__meta">
              {{if .featured_author}}<span class="featured-article__author">{{.featured_author}}</span>{{end}}
              {{if .featured_date}}<span class="featured-article__date">{{.featured_date}}</span>{{end}}
              {{if .featured_read_time}}<span class="featured-article__read-time">{{.featured_read_time}}</span>{{end}}
            </div>{{end}}$n2$),
    updated_at = now()
 WHERE is_active AND name = 'featured_article';

-- about-commercial-block — 1 field(s): domain
UPDATE content_components SET html_template =
  replace(replace(html_template,
         $o0$<section class="about-commercial-block" data-component="about-commercial-block">$o0$,
         $n0${{if .domain}}<section class="about-commercial-block" data-component="about-commercial-block">$n0$),
         $o1$  </div>
</section>$o1$,
         $n1$  </div>
</section>{{end}}$n1$),
    updated_at = now()
 WHERE is_active AND name = 'about-commercial-block';

-- =====================================================================
-- VERIFY — a DO block, NOT a list of SELECTs. `ON_ERROR_STOP` ignores a
-- non-empty result set, so a verify block made of SELECTs cannot stop the
-- COMMIT; only RAISE can. Two distinct failures are caught:
--
--   1. a replace() that matched NOTHING — the silent failure mode of this
--      technique. If another session edited one of these templates between the
--      fetch that built this file and its run, the replace is a no-op, the
--      UPDATE still reports success, and the defect survives looking fixed.
--      Caught by comparing each row against its own before-image.
--   2. a field still ungated afterwards — using the SAME regex the lint uses
--      (check_placeholder_fallbacks.field_gated), so a pass here means the lint
--      will report 0, not merely that this file ran.
-- =====================================================================
DO $verify$
DECLARE
    unchanged text;
    ungated   text;
BEGIN
    SELECT string_agg(c.name, ', ')
      INTO unchanged
      FROM content_components c
      JOIN migration_backups b
        ON b.target_id = c.id::text
       AND b.migration_name = '295_twenty_components_gate_their_declared_skip_fields.sql'
     WHERE c.html_template = b.old_value->>'html_template';
    IF unchanged IS NOT NULL THEN
        RAISE EXCEPTION 'template UNCHANGED after its UPDATE — a replace() matched nothing, so the gate was never applied: %', unchanged;
    END IF;

    SELECT string_agg(t.cname || '.' || t.field, ', ')
      INTO ungated
      FROM (VALUES
         ('hero','subheadline'),
         ('about-hero','subheadline'),
         ('case-studies-hero','subheadline'),
         ('contact-hero','subheadline'),
         ('services-hero','subheadline'),
         ('use-cases-hero','subheadline'),
         ('testimonials','headline'),
         ('social_proof','headline'),
         ('portfolio-showcase','headline'),
         ('content-listing','section_title'),
         ('system-stats','stat1_description'),
         ('system-stats','stat1_label'),
         ('system-stats','stat2_description'),
         ('system-stats','stat2_label'),
         ('system-stats','stat3_description'),
         ('system-stats','stat3_label'),
         ('system-stats','stat4_description'),
         ('system-stats','stat4_label'),
         ('archetype-result-card','stat_1_label'),
         ('archetype-result-card','stat_2_label'),
         ('archetype-result-card','stat_3_label'),
         ('bayesian-ranking-hero-tool_pre_037','stat_one_label'),
         ('bayesian-ranking-hero-tool_pre_037','stat_three_label'),
         ('bayesian-ranking-hero-tool_pre_037','stat_two_label'),
         ('hero-tool','cta_primary_label'),
         ('hero-tool','cta_secondary_label'),
         ('hero-tool','stat_one_label'),
         ('hero-tool','stat_three_label'),
         ('hero-tool','stat_two_label'),
         ('case-studies-grid','card1_stat_label'),
         ('case-studies-grid','card2_stat_label'),
         ('case-studies-grid','card3_stat_label'),
         ('case-studies-grid','card4_stat_label'),
         ('case-studies-grid','card5_stat_label'),
         ('product-specs','spec_1_value'),
         ('product-specs','spec_2_value'),
         ('product-specs','spec_3_value'),
         ('product-specs','spec_4_value'),
         ('product-specs','spec_5_value'),
         ('product-specs','spec_6_value'),
         ('product-specs','spec_7_value'),
         ('product-specs','spec_8_value'),
         ('platform-comparison','row1_platform1_value'),
         ('platform-comparison','row1_platform2_value'),
         ('platform-comparison','row1_spark_value'),
         ('platform-comparison','row2_platform1_value'),
         ('platform-comparison','row2_platform2_value'),
         ('platform-comparison','row2_spark_value'),
         ('platform-comparison','row3_platform1_value'),
         ('platform-comparison','row3_platform2_value'),
         ('platform-comparison','row3_spark_value'),
         ('platform-comparison','row4_platform1_value'),
         ('platform-comparison','row4_platform2_value'),
         ('platform-comparison','row4_spark_value'),
         ('platform-comparison','row5_platform1_value'),
         ('platform-comparison','row5_platform2_value'),
         ('platform-comparison','row5_spark_value'),
         ('Pricing Tiers','tier_2_name'),
         ('Pricing Tiers','tier_2_price'),
         ('Pricing Tiers','tier_3_name'),
         ('Pricing Tiers','tier_3_price'),
         ('featured_article','featured_author'),
         ('featured_article','featured_category'),
         ('featured_article','featured_date'),
         ('featured_article','featured_excerpt'),
         ('featured_article','featured_image'),
         ('featured_article','featured_read_time'),
         ('about-commercial-block','domain')
           ) AS t(cname, field)
      JOIN content_components c ON c.name = t.cname AND c.is_active
     WHERE c.html_template !~ ('\{\{-?\s*(if|with)\s+[^}]*\.' || t.field || '\y');
    IF ungated IS NOT NULL THEN
        RAISE EXCEPTION 'still UNGATED after this migration: %', ungated;
    END IF;

    RAISE NOTICE '295 OK — 20 templates changed, all 68 declared skip_field fields now gated';
END
$verify$;

COMMIT;
