-- noted.co.uk — experience patterns.
-- Applied out of band (psql -f). Per-site/product behaviour, not a schema change.
--
-- WHY THESE EXIST, AND WHY BEFORE THE APP IS BUILT
--   The owner's requirement is DECOMPOSITION: every part of this site under
--   framework control, including the parts that are behaviour rather than prose.
--   Sections and components decompose a brochure. An application is mostly
--   behaviour, so if decomposition stopped at the section level we would have a
--   nicely decomposed set of marketing pages wrapped around one opaque lump of
--   app — the very thing being ruled out, moved down a level.
--
--   `experience_patterns` is the mechanism that prevents that: a behaviour
--   becomes a named object with a contract, honest degraded states, a data
--   contract and a criteria template the checkers can run. Written BEFORE the
--   build so the build has something to satisfy, rather than being reverse
--   engineered afterwards into a description of whatever got made.
--
-- CHECK TYPES USED ARE ONLY THOSE THE RUNNER ACTUALLY SUPPORTS
--   Tier 2 (static): selector_exists, selector_count, interaction, asset_loads,
--   page_status_ok, attribute_absent, attribute_matches.
--   Tier 4 (browser): no_horizontal_overflow, no_console_errors,
--   has_visible_area, computed_values.
--   Anything else in a check object is INERT — the runner never reads it — so
--   unsupported intent is recorded in `_unsupported`, in the open, rather than
--   written as a check that silently never runs.
--
-- THE RUNNER'S LIMITS BITE THIS PRODUCT HARDER THAN MOST, and they are stated
-- per check below rather than in a footnote:
--   * checks cannot be ORDERED or made conditional on one another;
--   * there is no expect_within_ms and no retries;
--   * a failing dependency cannot be induced, so degraded states cannot be
--     tested by the platform at all — only observed opportunistically.
--   The behaviours most worth checking here ("it is still there after a
--   reload", "the same notes appear in a second browser") are ordered and
--   stateful BY NATURE. Do not read a green run as covering them.

\set ON_ERROR_STOP on

BEGIN;

-- ========================================================================
-- 1. authenticated-note-sync — the reason the rebuild exists
-- ========================================================================
INSERT INTO experience_patterns (
  name, kind, display_name, aka, description,
  primitives, section_types, destination_roles, funnel_stage, suitable_site_types,
  contract, states, degraded_states, data_contract,
  entry_points, requires_component_contract, requires_invariant,
  binding_schema, criteria_template, status, source, created_by
) VALUES (
  'authenticated-note-sync',
  'micro-journey',
  'Sign in and find the same notes',
  '["account-backed editor","cross-device notes"]'::jsonb,
  'A person writes something, closes the tab, and later opens the site on a different browser or device, signs in, and finds the same writing there. The rule that makes it honest: the editor only reports work as saved once the server has said so. A local echo of the text the person just typed is not evidence that anything was stored, and on this product a false "Saved" is the worst possible lie — it is the exact moment they stop worrying and close the tab.',
  '["submit","reveal"]'::jsonb,
  '["tool"]'::jsonb,
  '["account"]'::jsonb,
  -- funnel_stage is CHECK-constrained to awareness|consideration|conversion.
  -- Neither 'retention' (what this actually is) nor 'onboarding' exists: the
  -- vocabulary was built for marketing funnels, and a product behaviour that
  -- happens AFTER someone converts has nowhere to sit. 'conversion' is the
  -- least wrong, not the right answer. Flagged rather than quietly fudged —
  -- if more app-shaped patterns land, this column needs a wider vocabulary.
  'conversion',
  '["tool","app"]'::jsonb,
  $contract$[
    {
      "primitive": "submit",
      "control_role": "sign_in",
      "outcome": "on a successful sign-in response the person is shown their own notes, fetched from the server. On failure: one message that does not say whether the account exists, and no partial signed-in state.",
      "must_not": ["reveal whether an email address has an account", "show a signed-in shell with an empty note list as though the account had no notes"],
      "evidence": "server.go login handler returns one message for ErrNoAccount and ErrBadPassword and spends equal work on both; engine_test.go TestLoginDoesNotRevealWhetherAnAccountExists"
    },
    {
      "primitive": "submit",
      "control_role": "save_note",
      "outcome": "the note is sent, and the saved indicator changes to a saved state ONLY as a consequence of a successful response",
      "must_not": ["show saved on keystroke", "show saved while a request is in flight", "show saved after an error response"],
      "evidence": "the legacy app set its indicator locally, which is why this is written down as a contract rather than assumed"
    },
    {
      "primitive": "reveal",
      "control_role": "note_list",
      "outcome": "the list shows every note belonging to the signed-in account and no note belonging to anyone else",
      "must_not": ["render a note fetched by id without the account scope"],
      "evidence": "store.go scopes every read and write by account_id in the SQL; engine_test.go TestOneAccountCannotSeeAnothersNotes fails if that scope is removed (verified by mutation, 2026-08-11)"
    }
  ]$contract$::jsonb,
  '["signed_out","signing_in","signed_in","saving","saved","save_failed"]'::jsonb,
  $degraded$[
    {
      "when": "the engine is unreachable or returns an error while saving",
      "outcome": "the person is told plainly that the note is NOT saved, the text stays exactly where it is and remains editable, and the indicator does not read saved",
      "must_not": ["discard the typed text", "show saved", "silently retry for ever with no visible state"],
      "evidence": "NOT YET OBSERVED. The platform cannot induce a failing dependency, so this clause is unverifiable by the checkers and must be exercised by hand before launch.",
      "note": "This is the clause that protects the thing that cannot be recovered. A save that fails loudly costs a person ten seconds; a save that fails silently costs them the note."
    },
    {
      "when": "the session has expired",
      "outcome": "the person is asked to sign in again and their unsaved text survives that round trip",
      "must_not": ["redirect away from an editor containing unsaved text"]
    }
  ]$degraded$::jsonb,
  $data$ {
    "source": "the noted engine, same origin, {{binding.api_base}}",
    "endpoints": ["/api/login","/api/register","/api/logout","/api/me","/api/notes","/api/media/{id}","/api/import"],
    "session": "an HttpOnly, Secure, SameSite=Lax cookie. The browser never sees the token in JavaScript, so an XSS cannot lift a session.",
    "cross_origin": "none by design — the API is same-origin behind nginx, so there is no CORS surface to get wrong",
    "rendering_rule": "note text is rendered as TEXT, never as markup. A note is arbitrary content typed by a person and may contain anything.",
    "media": "fetched per item by id, never inlined into the note list, so a list of notes never carries megabytes"
  } $data$::jsonb,
  '["/","/app"]'::jsonb,
  '[]'::jsonb,
  '["notes are scoped by account in the query, not in the handler"]'::jsonb,
  $binding${
    "type": "object",
    "required": ["tool_section","sign_in_form","email_input","password_input","sign_in_submit","note_list","note_editor","save_indicator"],
    "properties": {
      "api_base": {"type":"string"},
      "tool_section": {"type":"string"},
      "sign_in_form": {"type":"string"},
      "email_input": {"type":"string"},
      "password_input": {"type":"string"},
      "sign_in_submit": {"type":"string"},
      "note_list": {"type":"string"},
      "note_editor": {"type":"string"},
      "save_indicator": {"type":"string"},
      "sample_email": {"type":"string"},
      "sample_password": {"type":"string"}
    }
  }$binding$::jsonb,
  $criteria${
    "container": "{{binding.tool_section}}",
    "profiles": ["desktop","mobile"],
    "checks": [
      {"id":"page_ok","type":"page_status_ok"},
      {"id":"sign_in_form_present","type":"selector_exists","selector":"{{binding.sign_in_form}}"},
      {"id":"email_input_present","type":"selector_exists","selector":"{{binding.email_input}}"},
      {"id":"password_input_present","type":"selector_exists","selector":"{{binding.password_input}}"},
      {"id":"password_field_is_a_password_field","type":"attribute_matches","selector":"{{binding.password_input}}","attribute":"type","matches":"password",
       "_why":"a text-typed password field shoulder-surfs and lands in autofill history"},
      {"id":"editor_present","type":"selector_exists","selector":"{{binding.note_editor}}"},
      {"id":"save_indicator_present","type":"selector_exists","selector":"{{binding.save_indicator}}"},
      {"id":"editor_actually_has_a_box","tier":4,"type":"has_visible_area","selector":"{{binding.note_editor}}",
       "_why":"three tools on this estate shipped work areas measuring 1146x0 — in the DOM and invisible — and selector_exists passed all three"},
      {"id":"no_console_errors","tier":4,"type":"no_console_errors"},
      {"id":"no_overflow","tier":4,"type":"no_horizontal_overflow","profiles":["mobile"]},
      {"id":"sign_in_round_trip","tier":4,"type":"interaction",
       "steps":[{"action":"fill","selector":"{{binding.email_input}}","value":"{{binding.sample_email}}"},
                {"action":"fill","selector":"{{binding.password_input}}","value":"{{binding.sample_password}}"},
                {"action":"click","selector":"{{binding.sign_in_submit}}"}],
       "expect":{"selector":"{{binding.note_list}}","text_matches":".{1,}"},
       "_unsupported":"expect_within_ms does not exist; the runner asserts 300ms after the step, and a sign-in is a network round trip. This check is therefore FLAKY BY CONSTRUCTION and a failure is not proof of a broken sign-in. Do not tighten it into a gate until the runner can wait."},
      {"id":"notes_survive_a_reload","tier":4,"type":"interaction",
       "steps":[{"action":"reload"}],
       "expect":{"selector":"{{binding.note_list}}","text_matches":".{1,}"},
       "_unsupported":"cannot be ordered after sign_in_round_trip — checks cannot be made conditional on one another — so this only means anything when the runner happens to arrive signed in. The REAL assertion behind it (the same notes appear in a second, independent browser session) is not expressible at all today and is covered by box/smoke-test.sh instead."}
    ]
  }$criteria$::jsonb,
  'draft',
  'manual',
  'noted-rebuild-workstream-2026-08-11'
) ON CONFLICT (name) DO NOTHING;

-- ========================================================================
-- 2. legacy-local-data-adoption — the migration, and an obligation
-- ========================================================================
INSERT INTO experience_patterns (
  name, kind, display_name, aka, description,
  primitives, section_types, destination_roles, funnel_stage, suitable_site_types,
  contract, states, degraded_states, data_contract,
  entry_points, requires_component_contract, requires_invariant,
  binding_schema, criteria_template, status, source, created_by
) VALUES (
  'legacy-local-data-adoption',
  'micro-journey',
  'Bring notes from the old browser-only version',
  '["local data rescue","IndexedDB adoption"]'::jsonb,
  'A site that used to store its data in the visitor browser now stores it server-side. Because the origin has not changed, that old data is STILL THERE in the browser, and the rebuilt site can read it. This journey finds it, shows the person what was found before asking anything of them, and offers to copy it into their account. The rule that makes it safe: it is READ-ONLY against the old store, always. Someone who starts this and wanders off must find their notes exactly as they left them.',
  '["reveal","submit"]'::jsonb,
  '["tool"]'::jsonb,
  '["account"]'::jsonb,
  'conversion',  -- see the note above: 'onboarding' is not in the vocabulary
  '["tool","app"]'::jsonb,
  $contract$[
    {
      "primitive": "reveal",
      "control_role": "detect",
      "outcome": "on finding a legacy store, the page states what it found — how many notes, how many recordings, how many photographs — BEFORE asking the person to sign up or do anything else",
      "must_not": ["ask for an email address before showing what was found", "report a count it has not actually read", "show an error to a visitor who simply has no legacy data"],
      "evidence": "browser-verified 2026-08-11: a different page on the same origin opened NotedDB and read notes, content and audio; all four object stores visible"
    },
    {
      "primitive": "submit",
      "control_role": "adopt",
      "outcome": "on a successful response the person is told exactly how many notes, recordings and photographs were brought across, including anything skipped",
      "must_not": ["delete or modify anything in the legacy store", "report success for items the server skipped", "leave the person unable to tell whether it worked"],
      "evidence": "server.go importBackup returns notes/recordings/photos/skipped counts and continues past a per-item failure rather than aborting the whole import"
    },
    {
      "primitive": "submit",
      "control_role": "download_instead",
      "outcome": "the person can take a file away without creating an account at all",
      "why": "an obligation, not a feature. These people trusted a tool that promised their notes would stay on their own machine. Requiring an account to retrieve their own writing would be a poor way to repay that."
    }
  ]$contract$::jsonb,
  '["scanning","found","none_found","adopting","adopted","adopt_failed"]'::jsonb,
  $degraded$[
    {
      "when": "the legacy database is absent (a new visitor, a different browser, or cleared storage)",
      "outcome": "a plain statement that there is nothing to bring across, with no error styling and no suggestion the person has done something wrong",
      "must_not": ["show a technical error", "imply data was lost"]
    },
    {
      "when": "the engine is unreachable during adoption",
      "outcome": "the legacy data is left completely untouched and the person is told to try later; the download route is offered as the fallback",
      "must_not": ["delete legacy data after a failed upload", "report partial success as success"],
      "evidence": "NOT YET OBSERVED — the platform cannot induce a failing dependency. Exercise by hand before launch."
    }
  ]$degraded$::jsonb,
  $data$ {
    "source": "browser IndexedDB, database NotedDB, stores notes/history/audio/images",
    "open_rule": "open WITHOUT a version number. Naming a version triggers onupgradeneeded and would run a migration against the only copy of somebody's notes.",
    "access": "READ-ONLY. This journey never deletes, never writes, never upgrades the legacy store.",
    "upload": "POST /api/import, accepting exactly the full-backup format the legacy app has been producing since 2026-08-10 — a shape fixed by files already on people's disks and therefore not tidyable",
    "media": "carried as data URIs inside the import payload, which is why the endpoint and nginx both allow a large body"
  } $data$::jsonb,
  '["/legacy"]'::jsonb,
  '[]'::jsonb,
  '["the legacy store is never written to, only read"]'::jsonb,
  $binding${
    "type": "object",
    "required": ["tool_section","summary_region","adopt_control","download_control"],
    "properties": {
      "tool_section": {"type":"string"},
      "summary_region": {"type":"string"},
      "adopt_control": {"type":"string"},
      "download_control": {"type":"string"},
      "empty_state": {"type":"string"}
    }
  }$binding$::jsonb,
  $criteria${
    "container": "{{binding.tool_section}}",
    "profiles": ["desktop","mobile"],
    "checks": [
      {"id":"page_ok","type":"page_status_ok"},
      {"id":"summary_region_present","type":"selector_exists","selector":"{{binding.summary_region}}"},
      {"id":"download_route_present","type":"selector_exists","selector":"{{binding.download_control}}",
       "_why":"the no-account escape route is a contract clause, so its absence must fail rather than be noticed by a person reading the page"},
      {"id":"no_console_errors","tier":4,"type":"no_console_errors",
       "_why":"this page opens IndexedDB on arrival; a thrown error here is the difference between rescuing someone's notes and appearing to have lost them"},
      {"id":"no_overflow","tier":4,"type":"no_horizontal_overflow","profiles":["mobile"]},
      {"id":"summary_is_actually_visible","tier":4,"type":"has_visible_area","selector":"{{binding.summary_region}}"}
    ],
    "_unsupported":"The checks above verify the page is present and does not throw. They CANNOT verify the behaviour that matters — that a browser holding legacy data is shown the right counts — because the runner cannot seed IndexedDB before a check. That is covered by a Playwright probe in the workstream (the same technique that proved same-origin readability on 2026-08-11) and MUST NOT be assumed from a green criteria run."
  }$criteria$::jsonb,
  'draft',
  'manual',
  'noted-rebuild-workstream-2026-08-11'
) ON CONFLICT (name) DO NOTHING;

COMMIT;

-- ------------------------------------------------------------------ verify --
SELECT name, kind, status,
       jsonb_array_length(contract)        AS contract_clauses,
       jsonb_array_length(degraded_states) AS degraded_states,
       jsonb_array_length(criteria_template->'checks') AS checks
FROM experience_patterns
WHERE created_by = 'noted-rebuild-workstream-2026-08-11'
ORDER BY name;

-- Every check type used must be one the runner actually executes. A typo here
-- is INERT, not an error, so this is the check that catches it.
SELECT DISTINCT c->>'type' AS check_type,
       (c->>'type') IN ('selector_exists','selector_count','interaction','asset_loads',
                        'page_status_ok','attribute_absent','attribute_matches',
                        'no_horizontal_overflow','no_console_errors','has_visible_area',
                        'computed_values') AS runner_supports_it
FROM experience_patterns p,
     LATERAL jsonb_array_elements(p.criteria_template->'checks') c
WHERE p.created_by = 'noted-rebuild-workstream-2026-08-11'
ORDER BY 2, 1;
