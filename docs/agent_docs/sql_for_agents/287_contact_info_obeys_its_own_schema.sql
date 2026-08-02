-- 287_contact_info_obeys_its_own_schema.sql — 2026-08-02, bugfix_140_contact_info_fabrication
--
-- CLOSES bugs_open/140: the shared `contact-info` component FABRICATES a phone number
-- and office hours when the site's datum is absent. 8 of 8 live uses serve invented
-- hours today; vetcomparison.uk (rendered 2026-07-31) also serves the invented phone
-- `+1234567890`, confirmed on the wire 2026-08-02.
--
-- THE ORGANISING FACT: the component's own input_schema ALREADY declares the correct
-- behaviour and the template disobeys it.
--
--     "hours":   { "source": "site_specs.identity.hours",   "on_missing": "skip_field" }
--     "phone":   { "source": "site_specs.identity.phone",   "on_missing": "skip_field" }
--     "address": { "source": "site_specs.identity.address", "on_missing": "skip_field" }
--     "section_title": { "source": "llm", "fallback": "Contact Us", "on_missing": "use_fallback" }
--
-- `skip_field` is the contract. The template substituted an invented value instead. So
-- this is NOT a policy change requiring a choice between two defensible behaviours — it
-- is making the template obey the contract it already publishes. Note the schema already
-- distinguishes a legitimate LABEL default (`section_title.fallback`) from a fabricated
-- FACT: only the template ignored the distinction.
--
-- WHAT IT DOES
--   A. Gates every contact card on its own datum and DELETES the fabricated fallbacks:
--        tel: href      {{if .phone}}…{{else}}+1234567890{{end}}          -> gated, literal gone
--        phone text     {{else}}+1 (234) 567-890{{end}}                   -> gated, literal gone
--        hours          {{else}}Monday – Friday, 9am – 6pm{{end}}         -> gated, literal gone
--        email          {{else}}info@example.com{{end}}                   -> gated, literal gone
--      A datum nobody supplied can no longer render. This is the rule bugs_open/111
--      established for the footer (RenderFallbackFooter, d4731109d) and never applied to
--      the section component.
--
--   B. Repairs a template/schema desync the quality pipeline detected on 2026-05-18 and
--      nothing ever consumed (quality_score 80, schema_template_synced FALSE,
--      quality_issues ["template var {{.intro}} has no schema entry"]).
--      The template read `.title` and `.intro`; the schema declares `section_title` and
--      `intro_text`. Those branches were therefore PERMANENTLY FALSE.
--      **All 8 live sites do supply section_title and intro_text** — "Get in Touch",
--      "Contact Darts Online", "Reach us directly", … — so 8 real headings and every
--      intro paragraph have been silently discarded in favour of a hardcoded
--      "Contact Information". Renaming restores them.
--      `address` was declared and sourced but rendered nowhere; it is now rendered
--      (gated). After this file template vars == schema fields exactly, 6 == 6, so
--      schema_template_synced becomes satisfiable.
--
--   C. Gates the .contact-grid container on having at least one card, so an all-absent
--      component renders nothing rather than an empty shell (111's container rule).
--
-- WHAT IT DOES NOT DO, ON PURPOSE
--   * Does NOT stamp real hours/phone into any site. Only the owner can state those;
--     inventing them is the bug. `hours` is supplied by 0 of 1,089 page_components
--     fleet-wide — the Hours card has never once rendered a real datum, anywhere, so
--     every instance of it ever served was fabricated.
--   * Does NOT touch `+44 (0) 7934 524 911`, which appears on six sites and looks alarming
--     until traced: it comes from sites.content_data.phone, it is the owner's own number,
--     and these are the owner's own portfolio sites. Real datum, correctly propagated.
--   * Does NOT patch the 8 stored page_components rows. The template re-fabricates on any
--     rerender of an unlocked row, so per-instance patching is the wrong tool here
--     (140 fix candidate 3). The rows correct themselves on next rerender.
--   * Does NOT drop `phone_display`'s CAPABILITY silently: the key is removed because it
--     is schema-undeclared and used by 0 rows fleet-wide, and keeping it would leave the
--     component desynced. If a display/href split is wanted later, add the field to
--     input_schema FIRST — that is the contract the template must follow.
--
-- BLAST RADIUS, measured live 2026-08-02 before submitting (not left for a reviewer):
--     email   present 8/8  -> every site keeps its Email card.
--     phone   present 6/8  -> those six keep it. vetcomparison.uk loses a card that today
--                            publishes `+1234567890`; idea.uk loses one whose stored render
--                            shows a number its content_data no longer holds (the 117 drift
--                            family) and which under today's template would rerender AS
--                            `+1234567890` — so gating prevents a future fabrication there
--                            rather than removing a present truth.
--     hours   present 0/8  -> all eight lose the Hours card. THAT IS THE CORRECTION.
--     address present 0/8  -> renders nothing today; turns on a declared capability.
--
-- CONSUMERS TOLD (owner ruling 2026-07-29 §3 — measuring is not telling): the 8 sites
-- belong to other lanes. The change to their guarantee is "a contact card whose datum you
-- never supplied will stop rendering, so a contact page may carry fewer cards after its
-- next rerender". Recorded in bugs_open/140 and named in the council submission.
--
-- ROLLBACK
--   UPDATE content_components c SET html_template = b.old_value->>'html_template'
--     FROM migration_backups b
--    WHERE c.id = '0bd72302-e9bf-4dc0-a615-41a9c919bf17'
--      AND b.migration_name = '287_contact_info_obeys_its_own_schema.sql'
--      AND b.target_id = '0bd72302-e9bf-4dc0-a615-41a9c919bf17';
--   (Restores the fabricating template verbatim. The 8 stored renders are untouched by
--    this file in either direction.)

BEGIN;

-- Before-image, so the rollback recipe above is executable rather than aspirational.
INSERT INTO migration_backups (migration_name, target_table, target_id, old_value, notes)
SELECT '287_contact_info_obeys_its_own_schema.sql',
       'content_components',
       id::text,
       jsonb_build_object('html_template', html_template, 'input_schema', input_schema),
       'bugs_open/140: pre-fix template, fabricates phone/hours/email when the datum is absent'
  FROM content_components
 WHERE id = '0bd72302-e9bf-4dc0-a615-41a9c919bf17';

UPDATE content_components
   SET html_template = $mig$<section class="contact-info-section" data-component="contact-info">
        <div class="contact-info-container">
            <h2>{{if .section_title}}{{.section_title}}{{else}}Contact Us{{end}}</h2>
            {{if .intro_text}}<p class="contact-intro">{{.intro_text}}</p>{{end}}
            {{if or .email .phone .address .hours}}
            <div class="contact-grid">
                {{if .email}}
                <div class="contact-card">
                    <div class="contact-icon">&#9993;</div>
                    <h3>Email</h3>
                    <a href="mailto:{{.email}}">{{.email}}</a>
                </div>
                {{end}}
                {{if .phone}}
                <div class="contact-card">
                    <div class="contact-icon">&#9742;</div>
                    <h3>Phone</h3>
                    <a href="tel:{{.phone}}">{{.phone}}</a>
                </div>
                {{end}}
                {{if .address}}
                <div class="contact-card">
                    <div class="contact-icon">&#8962;</div>
                    <h3>Address</h3>
                    <p>{{.address}}</p>
                </div>
                {{end}}
                {{if .hours}}
                <div class="contact-card">
                    <div class="contact-icon">&#9200;</div>
                    <h3>Hours</h3>
                    <p>{{.hours}}</p>
                </div>
                {{end}}
            </div>
            {{end}}
        </div>
    </section>
<style>
.contact-info-section {
    padding: var(--spacing-section, 5rem 2rem);
    background: var(--color-surface, #f8f9fa);
}
.contact-info-container {
    max-width: 1000px;
    margin: 0 auto;
    text-align: center;
}
.contact-info-section h2 {
    font-size: clamp(1.75rem, 4vw, 2.25rem);
    margin-bottom: 1rem;
    color: var(--color-primary, #1a1a2e);
}
.contact-intro {
    color: var(--color-text-muted, #555);
    margin-bottom: 2.5rem;
    line-height: 1.6;
    max-width: 600px;
    margin-left: auto;
    margin-right: auto;
}
.contact-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
    gap: 2rem;
}
.contact-card {
    background: var(--color-background, #fff);
    padding: 2rem;
    border-radius: 8px;
    box-shadow: 0 2px 8px rgba(0,0,0,0.08);
}
.contact-icon {
    font-size: 2rem;
    margin-bottom: 1rem;
}
.contact-card h3 {
    font-size: 1.1rem;
    margin-bottom: 0.5rem;
    color: var(--color-primary, #1a1a2e);
}
.contact-card a,
.contact-card p {
    color: var(--color-text-muted, #555);
    text-decoration: none;
    line-height: 1.5;
}
.contact-card a:hover {
    color: var(--color-accent, #0f3460);
}
@media (max-width: 768px) {
    .contact-info-section { padding: 3rem 1.5rem; }
    .contact-grid { grid-template-columns: 1fr; }
}
</style>$mig$,
       updated_at = now()
 WHERE id = '0bd72302-e9bf-4dc0-a615-41a9c919bf17';

-- Guard our own shape. Anything unexpected rolls the whole file back.
DO $guard$
DECLARE
    tpl  text;
    n    int;
BEGIN
    SELECT html_template INTO tpl FROM content_components
     WHERE id = '0bd72302-e9bf-4dc0-a615-41a9c919bf17';

    IF tpl IS NULL THEN
        RAISE EXCEPTION '287: the contact-info row is missing';
    END IF;

    -- 1. Every fabricated literal is gone. These are the bug.
    IF tpl LIKE '%+1234567890%'            THEN RAISE EXCEPTION '287: fabricated tel literal survives'; END IF;
    IF tpl LIKE '%+1 (234) 567-890%'       THEN RAISE EXCEPTION '287: fabricated phone display literal survives'; END IF;
    IF tpl LIKE '%9am%'                    THEN RAISE EXCEPTION '287: fabricated hours literal survives'; END IF;
    IF tpl LIKE '%info@example.com%'       THEN RAISE EXCEPTION '287: fabricated email literal survives'; END IF;

    -- 2. Every fact-bearing card is gated on its own datum.
    IF tpl NOT LIKE '%{{if .email}}%'   THEN RAISE EXCEPTION '287: email card is not gated'; END IF;
    IF tpl NOT LIKE '%{{if .phone}}%'   THEN RAISE EXCEPTION '287: phone card is not gated'; END IF;
    IF tpl NOT LIKE '%{{if .hours}}%'   THEN RAISE EXCEPTION '287: hours card is not gated'; END IF;
    IF tpl NOT LIKE '%{{if .address}}%' THEN RAISE EXCEPTION '287: address card is not gated'; END IF;
    IF tpl NOT LIKE '%{{if or .email .phone .address .hours}}%'
        THEN RAISE EXCEPTION '287: the grid container is not gated on having a card'; END IF;

    -- 3. The desync is repaired: the schema-declared names are the ones rendered,
    --    and the phantom ones are gone.
    IF tpl NOT LIKE '%{{if .section_title}}%' THEN RAISE EXCEPTION '287: section_title not rendered'; END IF;
    IF tpl NOT LIKE '%{{if .intro_text}}%'    THEN RAISE EXCEPTION '287: intro_text not rendered'; END IF;
    IF tpl LIKE '%.phone_display%'            THEN RAISE EXCEPTION '287: schema-undeclared phone_display survives'; END IF;
    IF tpl LIKE '%{{if .title}}%'             THEN RAISE EXCEPTION '287: phantom .title branch survives'; END IF;
    IF tpl LIKE '%{{if .intro}}<%'            THEN RAISE EXCEPTION '287: phantom .intro branch survives'; END IF;

    -- 4. Contracts other machinery depends on: the data-component attribute must equal
    --    the function (component_validation.go) and <section> must balance
    --    (scoreComponent's TemplateClosed).
    IF tpl NOT LIKE '%data-component="contact-info"%'
        THEN RAISE EXCEPTION '287: data-component attribute lost'; END IF;
    IF (length(tpl) - length(replace(tpl, '<section', ''))) / length('<section')
       <> (length(tpl) - length(replace(tpl, '</section>', ''))) / length('</section>')
        THEN RAISE EXCEPTION '287: <section> tags do not balance'; END IF;

    -- 5. The before-image was actually taken, or the rollback recipe is a lie.
    SELECT count(*) INTO n FROM migration_backups
     WHERE migration_name = '287_contact_info_obeys_its_own_schema.sql'
       AND target_id = '0bd72302-e9bf-4dc0-a615-41a9c919bf17';
    IF n <> 1 THEN RAISE EXCEPTION '287: expected exactly 1 before-image row, found %', n; END IF;

    RAISE NOTICE '287: contact-info now obeys its own schema — 4 fabricated literals removed, 4 cards gated, desync repaired';
END
$guard$;

COMMIT;
