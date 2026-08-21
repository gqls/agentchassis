-- FILE: docs/agent_docs/sql_for_agents/538_bugfix_238_contact_block_gates_and_repoints_phone_location.sql
--
-- bugs_open/238 §11.12 — stop `contact-block` shipping a DEAD TELEPHONE LINK,
-- and let its phone/address reach the store that actually holds them.
--
-- THE LIVE DEFECT, measured at the served page 2026-08-21 (a row census cannot
-- see it — this was found only by fetching the page):
--
--     <div class="cb-detail-value"><a href="tel:"></a></div>
--
-- **6 deployed rows across 3 sites are serving that right now.** It is
-- bugs_open/238's own defect class — an UNGATED field inside an attribute
-- rendering empty — on `href=` instead of `src=`.
--
-- TWO CAUSES, so two halves, and either alone leaves a site broken:
--
--   (a) THE SOURCE IS UNREACHABLE. `contact_phone` declares
--       `site_specs.contact.phone` and `contact_location` declares
--       `site_specs.contact.address`. `resolveSpecAlias` step 2 is hard-gated
--       (`if aspect != "identity" { return nil, false }`), and
--       `siteRowIdentityColumns` maps `phone -> phone` and
--       `address -> contact_address`. So the identity spelling reaches the
--       `sites` row and the contact spelling reaches nothing — the same
--       mismatch migration 525 fixed for `contact_email`.
--       Repointing alone fixes ONE site: leopardessconsulting.co.uk has a phone
--       (`+44 (0) 7934 524 911`). robot-hands.com and gamesdesign.co.uk have
--       none in either store and would keep serving the dead control.
--
--   (b) THE TEMPLATE IS UNGATED. A site with no phone SHOULD render no phone
--       row at all. This is RFC_009's option C, which the owner already chose
--       once: migration 295 gated 68 ungated fields on 2026-08-03 for exactly
--       this reason. Gating alone fixes all three sites' dead control but
--       leaves leopardess's real number unpublished.
--
-- Both, therefore. Owner-authorised 2026-08-21 ("please go ahead and do both"),
-- following the ruling the same day that the contact email should appear.
--
-- ⚠ THE GATE WRAPS THE WHOLE `cb-detail-item`, NOT THE VALUE. Each item is
-- icon + label + value. Gating only the value leaves an icon and a heading
-- standing over nothing, which is `bugs_closed/111` exactly ("footer contact
-- heading renders over an empty mailto"). The standing LANDMINE on
-- `on_missing: skip_field` makes the general point: read what ENCLOSES the
-- field and gate the smallest VALID unit, because a field in a fixed-arity row
-- is either a no-op to gate or emits malformed HTML.
--
-- ⚠ ALL THREE ITEMS ARE GATED, including the email 525 just repointed.
-- gamesdesign.co.uk has no email either, so without this it would render
-- `<a href="mailto:"></a>` on its next rebuild — the same defect one attribute
-- over. Fixing the phone and leaving that armed would be absurd.
--
-- PROVED BEFORE SHIPPING, in Go against the real renderer
-- (`contact_block_gate_test.go`, three tests):
--   * all three fields supplied      -> 3 detail items, nothing reported dead;
--   * phone missing                  -> NO `tel:` anywhere, exactly 2 items,
--                                       no orphaned label;
--   * nothing supplied               -> 0 items, `cb-details` container intact.
--   MUTATION: un-gate the phone item -> the test fails printing
--   `<a href="tel:"></a>`, i.e. it reproduces the live defect verbatim.
--
-- ⚠ A SIDE EFFECT WORTH KNOWING: once gated, `missingBareFields` can no longer
-- see these fields (it walks root-scope actions only), so `dead_url_control`
-- will never report this component again. That is correct — there is nothing to
-- report once the control is not emitted — but do not read the silence as
-- proof the data arrived. Check `content_data`, or the served page.
--
-- ⚠ CONFIG IS LIVE ON APPLY; THE PAGES ARE NOT. This changes what the next
-- render produces. The six pages keep serving the dead `tel:` until each is
-- re-rendered (`page_rerender`, `spec.reason='section_data_resolved'` — the
-- no-LLM MERGING path). Check each page's divergence stamp first: a rebuild
-- silently discards hand-patched `rendered_html` (bugs_open/229). All six were
-- verified `machine_made` on 2026-08-21.
--
-- ON THE NUMBER ITSELF, because publishing a phone is not a technical act:
-- `bugs_closed/140` already traced `+44 (0) 7934 524 911` across six sites and
-- established it is the OWNER'S OWN, correctly propagated, with "whether six
-- businesses should share one number" recorded as an owner question rather than
-- a defect. It was put to him again here and authorised.
--
-- Rollback: 538_..._ROLLBACK.sql (restores the ungated block and both sources).

\set ON_ERROR_STOP on

BEGIN;

-- Pre-flight: the row, the two sources, and the EXACT block this file rewrites.
-- A replace() that matches nothing is a silent no-op that reports success.
DO $guard$
DECLARE
    v_rows int;
    v_hits int;
    v_p    text;
    v_l    text;
BEGIN
    SELECT count(*) INTO v_rows FROM content_components
     WHERE function = 'contact-block' AND COALESCE(is_active, true);
    IF v_rows <> 1 THEN
        RAISE EXCEPTION '238/538: expected exactly 1 active contact-block row, found % — a fork appeared; convert by content_components.id (RFC_034) and re-derive', v_rows;
    END IF;

    SELECT input_schema->'fields'->'contact_phone'->>'source',
           input_schema->'fields'->'contact_location'->>'source'
      INTO v_p, v_l
      FROM content_components WHERE id = '4ebdccb5-6eb0-4b03-9fe1-0fa5cc315fb4';
    IF v_p IS DISTINCT FROM 'site_specs.contact.phone' THEN
        RAISE EXCEPTION '238/538: contact_phone.source is % (want site_specs.contact.phone) — already applied or moved', COALESCE(v_p,'(absent)');
    END IF;
    IF v_l IS DISTINCT FROM 'site_specs.contact.address' THEN
        RAISE EXCEPTION '238/538: contact_location.source is % (want site_specs.contact.address) — already applied or moved', COALESCE(v_l,'(absent)');
    END IF;

    SELECT count(*) INTO v_hits FROM content_components
     WHERE id = '4ebdccb5-6eb0-4b03-9fe1-0fa5cc315fb4'
       AND position($old$      <div class="cb-details">
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
$old$ IN html_template) > 0;
    IF v_hits <> 1 THEN
        RAISE EXCEPTION '238/538: the cb-details block this file rewrites was not found verbatim — the template has changed since 2026-08-21; re-derive the block, do NOT force';
    END IF;
END $guard$;

UPDATE content_components
   SET html_template = replace(html_template, $old$      <div class="cb-details">
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
$old$, $new$      <div class="cb-details">
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
$new$),
       input_schema = jsonb_set(
           jsonb_set(input_schema,
               '{fields,contact_phone,source}', '"site_specs.identity.phone"'::jsonb, false),
               '{fields,contact_location,source}', '"site_specs.identity.address"'::jsonb, false),
       updated_at = now()
 WHERE id = '4ebdccb5-6eb0-4b03-9fe1-0fa5cc315fb4';

-- VERIFY: gates present, sources moved, neighbours intact, and the OLD ungated
-- shape gone. Every comparison treats absence as failure.
DO $verify$
DECLARE
    v_gates int;
    v_ends  int;
    v_old   int;
    v_p     text;
    v_l     text;
    v_ptype text;
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
$old$ IN html_template),
           input_schema->'fields'->'contact_phone'->>'source',
           input_schema->'fields'->'contact_location'->>'source',
           input_schema->'fields'->'contact_phone'->>'type'
      INTO v_gates, v_old, v_p, v_l, v_ptype
      FROM content_components WHERE id = '4ebdccb5-6eb0-4b03-9fe1-0fa5cc315fb4';

    IF v_gates <> 3 THEN
        RAISE EXCEPTION '238/538 verify FAILED: % {{if .contact_*}} gate(s), want 3', v_gates;
    END IF;
    IF v_old <> 0 THEN
        RAISE EXCEPTION '238/538 verify FAILED: the ungated block is still present — replace() did not take';
    END IF;
    IF v_p IS DISTINCT FROM 'site_specs.identity.phone' OR v_l IS DISTINCT FROM 'site_specs.identity.address' THEN
        RAISE EXCEPTION '238/538 verify FAILED: sources are phone=% location=%', COALESCE(v_p,'(absent)'), COALESCE(v_l,'(absent)');
    END IF;
    -- jsonb_set replaces a subtree; a path typo would drop the neighbours while
    -- the source read fine. Assert one of them rather than assume.
    IF v_ptype IS DISTINCT FROM 'text' THEN
        RAISE EXCEPTION '238/538 verify FAILED: contact_phone lost its type (got %) — jsonb_set hit the wrong path', COALESCE(v_ptype,'(absent)');
    END IF;

    SELECT (length(html_template) - length(replace(html_template, '{{end}}', ''))) / length('{{end}}')
      INTO v_ends FROM content_components WHERE id = '4ebdccb5-6eb0-4b03-9fe1-0fa5cc315fb4';
    IF v_ends < 3 THEN
        RAISE EXCEPTION '238/538 verify FAILED: only % {{end}} in the template — unbalanced gates would fail every render', v_ends;
    END IF;

    RAISE NOTICE '238/538: contact-block gated (3 items) and phone/location repointed at the identity spelling';
END $verify$;

COMMIT;

-- ---------------------------------------------------------------------------
-- AFTER APPLYING — nothing on any page has changed yet.
--
-- Expected per site once re-rendered:
--   leopardessconsulting.co.uk  email + PHONE render; no address row.
--   robot-hands.com             email renders; NO phone row (was a dead tel:).
--   gamesdesign.co.uk           no rows at all (has none of the three) — and,
--                               importantly, no empty mailto/tel either.
--
-- Verify at the SERVED page, never the row:
--   curl -s https://<domain>/contact.html | grep -c 'href="tel:"'   -> expect 0
--   curl -s https://<domain>/contact.html | grep -o 'href="tel:[^"]*"' | head
