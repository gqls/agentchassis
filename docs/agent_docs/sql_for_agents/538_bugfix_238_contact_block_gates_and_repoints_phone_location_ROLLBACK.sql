-- FILE: docs/agent_docs/sql_for_agents/538_bugfix_238_contact_block_gates_and_repoints_phone_location_ROLLBACK.sql
--
-- Revert bugs_open/238 §11.12: restore the UNGATED cb-details block and put
-- contact_phone / contact_location back on the unreachable contact.* spelling.
-- Run BY HAND — the runner never applies a ROLLBACK sidecar.
--
-- ⚠ WHAT REVERTING RESTORES IS THE DEFECT, deliberately. Ungated + unreachable
-- is what was serving `<a href="tel:"></a>` on 6 rows across 3 sites. This is
-- not a return to a working state; it is a return to the broken one. Reach for
-- it only if the owner's 2026-08-21 authorisation is withdrawn, or if the gate
-- turns out to break a consumer nobody anticipated.
--
-- ⚠ AND IT DOES NOT UN-RENDER ANYTHING. Pages re-rendered while 538 was live
-- carry the gated markup (and leopardess's phone) in their deployed HTML and in
-- content_data. Reverting changes only what the NEXT render produces. To undo
-- the artefact you must also re-render each page — and note the PBP-039 carry
-- will re-supply a stored contact_phone on the next build even after the source
-- is un-pointed, because the carry reads the page's own stored row.

\set ON_ERROR_STOP on

BEGIN;

DO $guard$
DECLARE v_hits int;
BEGIN
    SELECT count(*) INTO v_hits FROM content_components
     WHERE id = '4ebdccb5-6eb0-4b03-9fe1-0fa5cc315fb4'
       AND position($new$      <div class="cb-details">
        {{if .contact_email}}<div class="cb-detail-item">
          <div class="cb-detail-icon" aria-hidden="true">
            <svg viewBox="0 0 24 24"><path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/><polyline points="22,6 12,13 2,6"/></svg>
          </div>
          <div>
            <div class="cb-detail-label">{{.email_label}}</div>
            <div class="cb-detail-value"><a href="mailto:{{.contact_email}}">{{.contact_email}}</a></div>
          </div>
        </div>{{end}}

        {{if .contact_phone}}<div class="cb-detail-item">
          <div class="cb-detail-icon" aria-hidden="true">
            <svg viewBox="0 0 24 24"><path d="M22 16.92v3a2 2 0 01-2.18 2 19.79 19.79 0 01-8.63-3.07A19.5 19.5 0 013.07 9.81 19.79 19.79 0 01.12 1.18 2 2 0 012.11 0h3a2 2 0 012 1.72 12.84 12.84 0 00.7 2.81 2 2 0 01-.45 2.11L6.09 7.91a16 16 0 006 6l1.27-1.27a2 2 0 012.11-.45 12.84 12.84 0 002.81.7A2 2 0 0122 14.92z"/></svg>
          </div>
          <div>
            <div class="cb-detail-label">{{.phone_label}}</div>
            <div class="cb-detail-value"><a href="tel:{{.contact_phone}}">{{.contact_phone}}</a></div>
          </div>
        </div>{{end}}

        {{if .contact_location}}<div class="cb-detail-item">
          <div class="cb-detail-icon" aria-hidden="true">
            <svg viewBox="0 0 24 24"><path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0118 0z"/><circle cx="12" cy="10" r="3"/></svg>
          </div>
          <div>
            <div class="cb-detail-label">{{.location_label}}</div>
            <div class="cb-detail-value">{{.contact_location}}</div>
          </div>
        </div>{{end}}
$new$ IN html_template) > 0;
    IF v_hits <> 1 THEN
        RAISE EXCEPTION '238/538 ROLLBACK: the gated block is not present verbatim — 538 is not applied, or the template moved since';
    END IF;
END $guard$;

UPDATE content_components
   SET html_template = replace(html_template, $new$      <div class="cb-details">
        {{if .contact_email}}<div class="cb-detail-item">
          <div class="cb-detail-icon" aria-hidden="true">
            <svg viewBox="0 0 24 24"><path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/><polyline points="22,6 12,13 2,6"/></svg>
          </div>
          <div>
            <div class="cb-detail-label">{{.email_label}}</div>
            <div class="cb-detail-value"><a href="mailto:{{.contact_email}}">{{.contact_email}}</a></div>
          </div>
        </div>{{end}}

        {{if .contact_phone}}<div class="cb-detail-item">
          <div class="cb-detail-icon" aria-hidden="true">
            <svg viewBox="0 0 24 24"><path d="M22 16.92v3a2 2 0 01-2.18 2 19.79 19.79 0 01-8.63-3.07A19.5 19.5 0 013.07 9.81 19.79 19.79 0 01.12 1.18 2 2 0 012.11 0h3a2 2 0 012 1.72 12.84 12.84 0 00.7 2.81 2 2 0 01-.45 2.11L6.09 7.91a16 16 0 006 6l1.27-1.27a2 2 0 012.11-.45 12.84 12.84 0 002.81.7A2 2 0 0122 14.92z"/></svg>
          </div>
          <div>
            <div class="cb-detail-label">{{.phone_label}}</div>
            <div class="cb-detail-value"><a href="tel:{{.contact_phone}}">{{.contact_phone}}</a></div>
          </div>
        </div>{{end}}

        {{if .contact_location}}<div class="cb-detail-item">
          <div class="cb-detail-icon" aria-hidden="true">
            <svg viewBox="0 0 24 24"><path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0118 0z"/><circle cx="12" cy="10" r="3"/></svg>
          </div>
          <div>
            <div class="cb-detail-label">{{.location_label}}</div>
            <div class="cb-detail-value">{{.contact_location}}</div>
          </div>
        </div>{{end}}
$new$, $old$      <div class="cb-details">
        <div class="cb-detail-item">
          <div class="cb-detail-icon" aria-hidden="true">
            <svg viewBox="0 0 24 24"><path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/><polyline points="22,6 12,13 2,6"/></svg>
          </div>
          <div>
            <div class="cb-detail-label">{{.email_label}}</div>
            <div class="cb-detail-value"><a href="mailto:{{.contact_email}}">{{.contact_email}}</a></div>
          </div>
        </div>

        <div class="cb-detail-item">
          <div class="cb-detail-icon" aria-hidden="true">
            <svg viewBox="0 0 24 24"><path d="M22 16.92v3a2 2 0 01-2.18 2 19.79 19.79 0 01-8.63-3.07A19.5 19.5 0 013.07 9.81 19.79 19.79 0 01.12 1.18 2 2 0 012.11 0h3a2 2 0 012 1.72 12.84 12.84 0 00.7 2.81 2 2 0 01-.45 2.11L6.09 7.91a16 16 0 006 6l1.27-1.27a2 2 0 012.11-.45 12.84 12.84 0 002.81.7A2 2 0 0122 14.92z"/></svg>
          </div>
          <div>
            <div class="cb-detail-label">{{.phone_label}}</div>
            <div class="cb-detail-value"><a href="tel:{{.contact_phone}}">{{.contact_phone}}</a></div>
          </div>
        </div>

        <div class="cb-detail-item">
          <div class="cb-detail-icon" aria-hidden="true">
            <svg viewBox="0 0 24 24"><path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0118 0z"/><circle cx="12" cy="10" r="3"/></svg>
          </div>
          <div>
            <div class="cb-detail-label">{{.location_label}}</div>
            <div class="cb-detail-value">{{.contact_location}}</div>
          </div>
        </div>
$old$),
       input_schema = jsonb_set(
           jsonb_set(input_schema,
               '{fields,contact_phone,source}', '"site_specs.contact.phone"'::jsonb, false),
               '{fields,contact_location,source}', '"site_specs.contact.address"'::jsonb, false),
       updated_at = now()
 WHERE id = '4ebdccb5-6eb0-4b03-9fe1-0fa5cc315fb4';

DO $verify$
DECLARE v_gates int; v_old int;
BEGIN
    SELECT (length(html_template) - length(replace(html_template, '{{if .contact_', ''))) / length('{{if .contact_'),
           position($old$      <div class="cb-details">
        <div class="cb-detail-item">
          <div class="cb-detail-icon" aria-hidden="true">
            <svg viewBox="0 0 24 24"><path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/><polyline points="22,6 12,13 2,6"/></svg>
          </div>
          <div>
            <div class="cb-detail-label">{{.email_label}}</div>
            <div class="cb-detail-value"><a href="mailto:{{.contact_email}}">{{.contact_email}}</a></div>
          </div>
        </div>

        <div class="cb-detail-item">
          <div class="cb-detail-icon" aria-hidden="true">
            <svg viewBox="0 0 24 24"><path d="M22 16.92v3a2 2 0 01-2.18 2 19.79 19.79 0 01-8.63-3.07A19.5 19.5 0 013.07 9.81 19.79 19.79 0 01.12 1.18 2 2 0 012.11 0h3a2 2 0 012 1.72 12.84 12.84 0 00.7 2.81 2 2 0 01-.45 2.11L6.09 7.91a16 16 0 006 6l1.27-1.27a2 2 0 012.11-.45 12.84 12.84 0 002.81.7A2 2 0 0122 14.92z"/></svg>
          </div>
          <div>
            <div class="cb-detail-label">{{.phone_label}}</div>
            <div class="cb-detail-value"><a href="tel:{{.contact_phone}}">{{.contact_phone}}</a></div>
          </div>
        </div>

        <div class="cb-detail-item">
          <div class="cb-detail-icon" aria-hidden="true">
            <svg viewBox="0 0 24 24"><path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0118 0z"/><circle cx="12" cy="10" r="3"/></svg>
          </div>
          <div>
            <div class="cb-detail-label">{{.location_label}}</div>
            <div class="cb-detail-value">{{.contact_location}}</div>
          </div>
        </div>
$old$ IN html_template)
      INTO v_gates, v_old
      FROM content_components WHERE id = '4ebdccb5-6eb0-4b03-9fe1-0fa5cc315fb4';
    IF v_gates <> 0 OR v_old = 0 THEN
        RAISE EXCEPTION '238/538 ROLLBACK verify FAILED: gates=% ungated_block_present=%', v_gates, (v_old <> 0);
    END IF;
    RAISE NOTICE '238/538 ROLLBACK: ungated block and contact.* sources restored — the dead tel: control returns on the next render';
END $verify$;

COMMIT;
