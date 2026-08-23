-- SQL_2026-08-23b_convert_unplaced_library_templates.sql
--
-- RFC_032 §8 completion, the three rows the placement-driven seed could not reach.
--
-- WHY NOT A WORK ITEM. The conversion pipeline files one work item per component
-- and site_work_items.site_id is NOT NULL, while content_components has no site
-- column -- a template's site is only reachable through the pages that place it.
-- These three have ZERO placements, so there is no honest site to file them
-- against and the seed's JOIN skips them entirely. Filing them under an
-- arbitrary site would put a false relationship in that site's queue to satisfy
-- a constraint, so the write is done here instead, mirroring writeScopedTemplate
-- (fix_component_template_action.go) exactly: snapshot the current template into
-- component_versions at MAX+1, then UPDATE.
--
-- THE CONVERSION WAS NOT HAND-WRITTEN. Each new template is the output of the
-- REAL converter (actions.ConvertTemplateToInstanceScope) and passed the REAL
-- acceptance gate (actions.GateConvertedTemplate: render two instances, run the
-- collision detector), with a CONTROL proving the gate REFUSES the unconverted
-- original -- otherwise a gate that accepts everything would vouch for nothing.
-- All three: ok=true, swaps=1, needsJudgedPool=false, control refused. The diff
-- is one line per template, the wrapper id, and nothing else.
--
-- ROWS (as of 2026-08-23):
--   pricing 6175e049  ACTIVE,   0 placements -- owner: convert, keep as library inventory
--   header  3f2fb70c  inactive, 0 placements -- dead since the chrome system replaced it
--   footer  fe11c787  inactive, 0 placements -- dead, same
-- The two inactive rows are converted too, deliberately: they cannot render today,
-- but they are a REACTIVATION HAZARD once the {{.ComponentID}} bindings are gone --
-- a reactivated row would render id="" on every instance, which the collision
-- detector cannot see (reElementID requires a non-brace character).
--
-- GUARDED: each UPDATE requires the row to still spell the placeholder, so
-- re-running is a no-op and a concurrent edit by another session is not clobbered.
--
-- VERIFY: SELECT count(*) FROM content_components WHERE html_template LIKE '%ComponentID%';
--         -- must be 0 before the render bindings are deleted.

-- pricing
INSERT INTO component_versions (component_id, version_number, html_template, change_source)
SELECT c.id,
       COALESCE((SELECT MAX(v.version_number) FROM component_versions v WHERE v.component_id=c.id),0)+1,
       c.html_template, 'rfc032_unplaced_library_conversion'
FROM content_components c
WHERE c.function='pricing' AND c.html_template LIKE '%ComponentID%';

UPDATE content_components
SET html_template = $tpl$<section id="{{.InstanceID}}" class="section section--pricing">
  <div class="container">
    <h2 class="section__title section__title--center">{{.section_title}}</h2>
    <div class="pricing-tiers grid grid--3">
      <div class="pricing-tier card">
        <h3 class="pricing-tier__name">{{.tier_1_name}}</h3>
        <div class="pricing-tier__price">{{.tier_1_price}}</div>
        <ul class="pricing-tier__features">
          <li class="pricing-tier__feature">{{.tier_1_feature_1}}</li>
          <li class="pricing-tier__feature">{{.tier_1_feature_2}}</li>
          <li class="pricing-tier__feature">{{.tier_1_feature_3}}</li>
        </ul>
        {{if .tier_1_cta_url}}<a href="{{.tier_1_cta_url}}" class="button button--secondary button--full-width">{{.tier_1_cta}}</a>{{end}}
      </div>
      {{if and .tier_2_name .tier_2_price}}<div class="pricing-tier pricing-tier--featured card">
        <div class="pricing-tier__badge">Popular</div>
        <h3 class="pricing-tier__name">{{.tier_2_name}}</h3>
        <div class="pricing-tier__price">{{.tier_2_price}}</div>
        <ul class="pricing-tier__features">
          <li class="pricing-tier__feature">{{.tier_2_feature_1}}</li>
          <li class="pricing-tier__feature">{{.tier_2_feature_2}}</li>
          <li class="pricing-tier__feature">{{.tier_2_feature_3}}</li>
          <li class="pricing-tier__feature">{{.tier_2_feature_4}}</li>
        </ul>
        {{if .tier_2_cta_url}}<a href="{{.tier_2_cta_url}}" class="button button--primary button--full-width">{{.tier_2_cta}}</a>{{end}}
      </div>{{end}}
      {{if and .tier_3_name .tier_3_price}}<div class="pricing-tier card">
        <h3 class="pricing-tier__name">{{.tier_3_name}}</h3>
        <div class="pricing-tier__price">{{.tier_3_price}}</div>
        <ul class="pricing-tier__features">
          <li class="pricing-tier__feature">{{.tier_3_feature_1}}</li>
          <li class="pricing-tier__feature">{{.tier_3_feature_2}}</li>
          <li class="pricing-tier__feature">{{.tier_3_feature_3}}</li>
          <li class="pricing-tier__feature">{{.tier_3_feature_4}}</li>
          <li class="pricing-tier__feature">{{.tier_3_feature_5}}</li>
        </ul>
        {{if .tier_3_cta_url}}<a href="{{.tier_3_cta_url}}" class="button button--secondary button--full-width">{{.tier_3_cta}}</a>{{end}}
      </div>{{end}}
    </div>
  </div>
</section>
$tpl$, updated_at = now()
WHERE function='pricing' AND html_template LIKE '%ComponentID%';

-- header
INSERT INTO component_versions (component_id, version_number, html_template, change_source)
SELECT c.id,
       COALESCE((SELECT MAX(v.version_number) FROM component_versions v WHERE v.component_id=c.id),0)+1,
       c.html_template, 'rfc032_unplaced_library_conversion'
FROM content_components c
WHERE c.function='header' AND c.html_template LIKE '%ComponentID%';

UPDATE content_components
SET html_template = $tpl$<header id="{{.InstanceID}}" class="site-header">
  <nav class="site-header__nav container">
    <div class="site-header__brand">{{.brand_name}}</div>
    <ul class="site-header__menu">
      <li class="site-header__menu-item"><a href="#hero" class="site-header__link">Home</a></li>
      <li class="site-header__menu-item"><a href="#features" class="site-header__link">Features</a></li>
      <li class="site-header__menu-item"><a href="#pricing" class="site-header__link">Pricing</a></li>
      <li class="site-header__menu-item"><a href="#faq" class="site-header__link">FAQ</a></li>
    </ul>
    <a href="#call_to_action" class="button button--primary button--small">{{.cta_text}}</a>
  </nav>
</header>
$tpl$, updated_at = now()
WHERE function='header' AND html_template LIKE '%ComponentID%';

-- footer
INSERT INTO component_versions (component_id, version_number, html_template, change_source)
SELECT c.id,
       COALESCE((SELECT MAX(v.version_number) FROM component_versions v WHERE v.component_id=c.id),0)+1,
       c.html_template, 'rfc032_unplaced_library_conversion'
FROM content_components c
WHERE c.function='footer' AND c.html_template LIKE '%ComponentID%';

UPDATE content_components
SET html_template = $tpl$<footer id="{{.InstanceID}}" class="site-footer">
  <div class="container">
    <div class="site-footer__content grid grid--4">
      <div class="site-footer__col">
        <h3 class="site-footer__brand">{{.brand_name}}</h3>
        <p class="site-footer__tagline">{{.tagline}}</p>
      </div>
      <div class="site-footer__col">
        <h4 class="site-footer__heading">Product</h4>
        <ul class="site-footer__links">
          <li><a href="#features" class="site-footer__link">Features</a></li>
          <li><a href="#pricing" class="site-footer__link">Pricing</a></li>
          <li><a href="#" class="site-footer__link">Updates</a></li>
        </ul>
      </div>
      <div class="site-footer__col">
        <h4 class="site-footer__heading">Company</h4>
        <ul class="site-footer__links">
          <li><a href="#" class="site-footer__link">About</a></li>
          <li><a href="#" class="site-footer__link">Blog</a></li>
          <li><a href="#" class="site-footer__link">Careers</a></li>
        </ul>
      </div>
      <div class="site-footer__col">
        <h4 class="site-footer__heading">Legal</h4>
        <ul class="site-footer__links">
          <li><a href="#" class="site-footer__link">Privacy</a></li>
          <li><a href="#" class="site-footer__link">Terms</a></li>
          <li><a href="#" class="site-footer__link">Contact</a></li>
        </ul>
      </div>
    </div>
    <div class="site-footer__bottom">
      <p class="site-footer__copyright">{{.copyright}}</p>
    </div>
  </div>
</footer>
$tpl$, updated_at = now()
WHERE function='footer' AND html_template LIKE '%ComponentID%';
