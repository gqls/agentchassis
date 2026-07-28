-- p4_30 — the £8 example place on the request form
--
-- Both boxes ship UNTICKED and the wording sits with them: a pre-ticked box is
-- not consent, and consent given without the terms in front of you is not
-- consent either. Two separate checkboxes because they are two decisions, which
-- is exactly how the Go reads them — asking for the tier without agreeing to
-- publication gets the standard price, and agreeing without asking records no
-- consent at all.
--
-- No JavaScript. The block is always visible rather than revealed by a script:
-- the server enforces both conditions regardless, and a consent control that
-- depends on JS to appear is one that can silently fail to appear.
--
-- Safe to edit directly: this component is forked to idea.uk (1 site, 1
-- instance, checked before p4_22).
--
-- ORDER OF OPERATIONS this was applied in, and it matters: binary first (tier
-- code present, tier OFF), then the env vars, then this. That way there was
-- never a visible offer the backend would ignore.

\set comp '8a88fcd4-83fe-4f7a-bed1-f270e6edee53'

SELECT (html_template LIKE '%rr-example%') AS already_present_before FROM content_components WHERE id = :'comp';

BEGIN;

UPDATE content_components
SET html_template = replace(html_template,
      '      <div class="rr-hp" aria-hidden="true">',
      $h$      <div class="rr-example">
        <p class="rr-example-h">The £8 example place</p>
        <p>We are building a library of example reports so people can see the work before they buy. Take one of those places and you get the full report — same research, same length — for <strong>£8</strong>.</p>
        <p>In return, you let us publish it as an example. We publish the idea and the report. Your name, your email and your company stay out of it.</p>
        <p>Which reports go in the library is our call — yours might appear next week, or never. Taking it out again is yours: email us and it comes down within five working days.</p>
        <p>Everything else matches the full report — same delivery, same refund if it turns up nothing worth acting on. Want your idea kept private? That is what the £29 report is for.</p>
        <label class="rr-check"><input type="checkbox" name="tier" value="example"> Yes — I would like an £8 example place</label>
        <label class="rr-check"><input type="checkbox" name="publish_consent" value="yes"> I agree you may publish this report as an example, anonymously</label>
      </div>
      <div class="rr-hp" aria-hidden="true">$h$),
    updated_at = now()
WHERE id = :'comp' AND html_template NOT LIKE '%rr-example%';

UPDATE content_components
SET html_template = replace(html_template,
      '.rr-form { display: flex; flex-direction: column; gap: 1.5rem; }',
      $c$.rr-form { display: flex; flex-direction: column; gap: 1.5rem; }
.rr-example { border: 1px solid var(--color-border, #ccc); border-radius: 4px; padding: 1.25rem;
  display: flex; flex-direction: column; gap: .75rem; }
.rr-example-h { font-weight: 600; margin: 0; }
.rr-example p { margin: 0; font-size: .95rem; line-height: 1.55; }
.rr-check { display: flex; gap: .6rem; align-items: flex-start; font-size: .95rem; cursor: pointer; }
.rr-check input { margin-top: .3rem; flex: 0 0 auto; }
$c$),
    updated_at = now()
WHERE id = :'comp' AND html_template NOT LIKE '%.rr-example {%';

COMMIT;

-- Both must be 1: the markup and its styling.
SELECT (html_template LIKE '%name="tier"%')::int          AS tier_input,
       (html_template LIKE '%name="publish_consent"%')::int AS consent_input,
       (html_template LIKE '%.rr-example {%')::int         AS css_present,
       (html_template LIKE '%checked%')::int                AS any_prechecked_box
FROM content_components WHERE id = :'comp';
