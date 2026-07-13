The 5-items problem — full explanation
What's supposed to happen
When webdesign-agent runs, one of its early steps is load_site_context. That step is supposed to read everything the agent needs about the site — including which content_components are actually used on its pages — and stash that in collected_data under site_context.
Specifically, there's a field site_context.all_component_functions which should be an array of every distinct content_components.function value used by every page_component on every page of the site. For gaswholesalers, that real list is 25 items (from the diagnostic earlier):
about-content, call-to-action, case-studies-list, contact-form,
contact-info, differentiators, faq, features, generic-text-block,
hero, hero-about, hero-case-studies, hero-contact, hero-services,
info-card-grid, latest-news, pricing, services-grid, social-proof,
testimonials, tool-breakeven-volume-calculator,
tool-fuel-budget-forecaster, tool-fuel-cost-estimator,
tool-gas-unit-converter, use-cases-list
Then later in the workflow, render_css_from_spec_action.go calls extractCSSComponents(collectedData) which reads that array and uses it as the input to loadComponentCSSSnippets. The CSS snippet matching WHERE applies_to && <component list> uses this. So every CSS snippet whose applies_to overlaps the site's actual components gets included in the deployed styles.css.
What's actually happening
The diagnostic queried collected_data -> 'site_context' -> 'all_component_functions' on the three most recent webdesign-agent orchestrations across two different sites. All three returned the same 5 items, just in different orders:
["hero", "services-grid", "differentiators", "social-proof", "call-to-action"]
These aren't a subset of the real components — they're a fake list. Some piece of code is putting these 5 items in all_component_functions regardless of what the site actually contains. Same 5 items for gaswholesalers (which has 25 real components) as for the other test site (which presumably has its own different list).
I noticed this is exactly the hardcoded fallback list in extractCSSComponents:
go// render_css_from_spec_action.go line 363
return []string{"hero", "services-grid", "differentiators", "social-proof", "call-to-action"}
That's the function's fallback when site_context.all_component_functions is missing. But the field IS present in the collected_data (the query found it), with these exact 5 items. So either:

load_site_context is calling that fallback list and storing it as if it were the real component list. Possible but odd.
OR another step writes the fallback into collected_data after extractCSSComponents runs.
OR there's a different (older) hardcoded list somewhere in load_site_context that happens to match. The orders varying between runs suggests something stochastic — maybe a SELECT that returns rows in non-deterministic order, but always the same 5 rows.

Why it matters
Every css_snippet whose applies_to contains ONLY component functions outside those 5 will never reach any site via the proper pipeline. From the 21 css_snippets we saw:
Snippetapplies_toWould match the fake list?fade-in-up["hero", "section", "card"]yes (matches "hero")responsive-grid["features", "cards", "gallery"]noLatest News Grid["latest-news"]noNews Listing Page["news-listing"]noinput-modern["form", "input", "newsletter"]nocard-bordered["card", "feature"]nocard-glass["card", "feature", "testimonial"]nohover-lift["card", "feature", "testimonial"]nobtn-3d, btn-glass, btn-icon, btn-minimal, btn-outline-fill["button", "cta", ...]nohover-glow["button", "card", "cta"]nohover-scale["card", "image", "button"]nopulse-attention, shake-attention["cta", "button", ...]noborder-animate["link", "nav-item", "card"]noborder-gradient["card", "feature", "cta"]nofloat["hero-image", "icon", "decoration"]no
Only fade-in-up out of 21 css_snippets ever reaches any site. The other 20 have been dead weight since they were added. This is also why every site has the same surprisingly-bare aesthetic — none of the button styles, card styles, hover effects, or section-specific styling has been getting included.
