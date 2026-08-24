-- 594 — the four pass-through prose slots are told to use structure (bugs_open/381, arm B1)
--
-- ✅ HOLD RELEASED 2026-08-24 — the release condition below was MET and VERIFIED, and this
-- banner replaces the "DO NOT APPLY" one that stood here for about an hour. (A resolved
-- warning left in place is its own defect: it makes the correct action look premature.)
--
-- THE EVIDENCE, with controls in both directions. The running chassis commit is
-- `70fd163c24eae0c444bae7a425bb3d3c3096f7e4`, read from the binary's own record —
--     SELECT git_commit FROM service_binary_capabilities
--      WHERE service='agent-chassis' AND kind='build' ORDER BY last_seen_at DESC LIMIT 1;
-- (`platform/buildcapability`, RFC_040 — this table has NO shelf life, unlike the startup
-- `build provenance` line, which had already scrolled). Then:
--     git merge-base --is-ancestor 714789d7b 70fd163c24…   -> PASSES  (the 305 </th fix: LIVE)
--     git merge-base --is-ancestor <a later commit> …      -> FAILS   (control behaves)
--     git merge-base --is-ancestor <an old commit> …       -> PASSES  (control behaves)
--
-- ⚠ WHY THE OBVIOUS METHODS FAILED, recorded because they are what anyone reaches for:
--   * `grep -a <sha> /proc/1/exe` — USELESS here. `buildinfo.GitCommit` is ONE STRING, not
--     an ancestry, so a binary that certainly contains your commit returns ABSENT for it.
--     Worse, the 40-zeros "must be absent" control comes back PRESENT (Go's internal digit
--     table matches a run of zeros), so that control cannot fail and converts "I did not
--     check" into "I checked and it passed". Two lanes were burned by this
--     (`bugs_open/215` on v1.0.1288, `bugs_open/299` on v1.0.1316).
--   * a pod's `.status.startTime` — dates the ROLL, not the IMAGE. Ours started 15:39Z,
--     an hour AFTER the fix was committed at 14:39:30Z, and that proves nothing either way:
--     a tag can be built before a commit and rolled after it.
--
-- ⚠ FORWARD-LOOKING, from the 305 lane: `</blockquote`, `</dd`, `</dt`, `</caption` and
-- `</section` are still ABSENT from that scanner's sentence-boundary set, deliberately,
-- because RULE 10 does not produce them. **If a later revision adds any of those to this
-- guidance or to RULE 10, tell that lane so it can probe and fixture them** — guessing at
-- prefixes is exactly how the `</th`/`</td` asymmetry arose.
--
--
-- > **⚠ CORRECTED 2026-08-24 by the `bugs_open/305` lane, after this file was applied — the
-- > premise for the hold above was WRONG, and the correction matters more than the hold did.**
-- > I said these two migrations were "what first let the writer emit a `<table>` into these
-- > slots", i.e. what made the `</th` scanner defect reachable. **Not so.** `[MEASURED
-- > 2026-08-24 by that lane]` `<th>` markup has been reaching `page_components` since **10
-- > August** — 2 components the week of 08-10, 14 the week of 08-17, 1 the week of 08-24, **17
-- > total** — and an inspected instance shows `rendered_has_real_table=true`, `escaped=false`.
-- > **The render path was already passing markup through at `type: text`.** These migrations
-- > raise the RATE; they did not open the path. That lane's fix was closing a fortnight-old live
-- > defect, not a prospective one. (Damage from that fortnight: **zero** — 0 of 76 `<th>` cells
-- > fleet-wide, all history, ever carried a define-by-negation construction, because header
-- > cells are labels like "Rate (£/ha)", not prose.)
-- >
-- > **The consequence for VERIFYING this migration, and it is a trap:** because that slot already
-- > rendered real tables before the retype, **looking for a `<table>` in `rendered_html` would
-- > show one either way and prove nothing.** It is a check that cannot fail — the same disease as
-- > the 40-zeros control described above, arriving from the other direction. **The demand control
-- > that DOES discriminate** is the one in the RUNBOOK: the planner's capability distribution
-- > moving 96/2 → 93/6 (`component_expresses` reading `{html-block,list,table}` where it read
-- > `{}`), plus each field's `llm_guidance` and declared type read back literally. Verify there,
-- > never at the served table.
-- >
-- > The hold itself cost about an hour and was harmless; the reasoning that a hold must be
-- > ENFORCED (`_HOLD.sql`) rather than documented still stands, since another session's `--apply`
-- > has no file scope. What was wrong was only my claim about what this change made reachable.
--
-- Applied by hand (the runner has no file or directory scope) and then recorded with
-- `--record-only`. See the lane RUNBOOK §8.
--
--
-- WHY, AND THIS IS THE MEASURED HALF OF THE FIX. The owner's "wall of text" is a
-- `generic-text-block` instance: 1,486 words on garden-tools.uk's how-we-assess, 14
-- paragraphs, no subheads, no list, no emphasis. Its template is a pass-through
-- (`<div class="section__content">{{.content}}</div>`, text/template, NO escaping),
-- so it would have rendered a <h3> or a <ul> unchanged. Nothing stopped the writer.
--
-- THE LEVER IS `llm_guidance`, NOT THE DECLARED TYPE — and the estate already proves
-- it. A controlled comparison with the type held constant `[MEASURED 2026-08-24, 30d
-- instances]`:
--
--   article-body.content        type: text, HAS llm_guidance ("Use h2 for main
--                               sections, h3 for subsections, p for paragraphs,
--                               ul/ol for lists")   -> 116/153 = 76% render a list
--   generic-text-block.content  type: text, NO llm_guidance at all
--                                                   -> 12/173 =  7% render a list
--
-- Same declared type, same renderer, same RULE 9 ("text fields: plain string, no HTML
-- wrapping") — and an ELEVEN-FOLD difference in outcome. So RULE 9 does not bind and
-- the field's own guidance does. An earlier draft of this fix had it the other way
-- round (retype to `html` and the writer is freed); `article-body` is the disproof and
-- the correction is recorded in the lane's NOTES and in WRONG_CALLS.md.
--
-- WHAT THE RETYPE IS STILL FOR, since this file does it too. The writer's prompt
-- prints each field as `` `content` (TYPE, required) `` and its rules are addressed by
-- type. RULE 10 currently names types `rich_text` and `content` — declared by ZERO
-- components fleet-wide — so the only rule permitting structure is addressed to
-- nobody, while RULE 9 tells every `text` field to emit no HTML. Retyping these four
-- to `html` is what ROUTES them from RULE 9 to RULE 10 (rewritten in 595). It is
-- honesty about what the slot is, not the force that changes behaviour.
--
-- ⚠ THE TYPE IS OTHERWISE INERT, SO THIS FILE ASSERTS THE LITERAL AFTERWARDS.
-- `DeclaredTypeSatisfied` (datahelpers/content_type_violations.go:262) is default-TRUE:
-- only `array`/`list` are ever checked. `hmtl`, `HTML` or `` would behave exactly like
-- `html` and NO downstream check could ever surface the typo (flagged by the
-- staged-component-build lane, 2026-08-24: "a check that cannot return false"). The
-- verify block therefore reads each of the four back and demands the literal 'html'.
--
-- FOUR FIELDS, NOT FIVE. `report-dossier.body` matched the census predicate (an llm
-- text field rendered directly inside a block container) and is EXCLUDED: its own
-- guidance says "Pre-rendered dossier HTML from create_report_page. Never authored by
-- an LLM and never assembled from a template." It is not a writer slot. Only the
-- guidance text distinguishes it; the structural predicate cannot.
--
-- WHY <img>/<figure>/<iframe> ARE FORBIDDEN IN THE GUIDANCE (requested by the
-- editorial_design_uplift lane, adopted). In-blob imagery is the loss class that
-- lane's inline-imagery phases and features_open/035 exist to retire — under 035 a
-- figure becomes an addressable child component instance, not a writer emission
-- inside a blob that dies on the next rewrite (the bugs_open/238 class). Without this
-- sentence, permitting HTML here would be the enabling edit for recreating it
-- fleet-wide. For the same reason the guidance names NO class attributes: the writer
-- owns prose STRUCTURE, the component system owns design, and a class name in
-- guidance is a comment, not a control.
--
-- ⚠ THE CONTAINER CHECK — the objection this file was nearly wrong about, now CLEARED
-- (council `ca400ba6`, bug_historian seat, severity medium; the sharpest objection of the
-- round). If any of these four templates wrapped its `{{.content}}` slot in a literal
-- `<p>`, then writer output containing `<h3>`/`<ul>`/`<table>` would be nested inside a
-- paragraph: invalid HTML, repaired inconsistently by each browser, and INVISIBLE
-- everywhere we look — the DB row is schema-valid, no check fails, and only the served
-- page is wrong. That is the platform's most-repeated failure family (016b §9, the
-- `<style>`-in-the-wrong-half case).
-- `[MEASURED 2026-08-24, live DB]` all four put the slot in a `<div>`, none in a `<p>`:
--     generic-text-block      <div class="section__content">{{.content}}</div>
--     about-content           <div class="about-text">{{.content}}</div>
--     illustrated-text-block  <div class="section__content">{{.content}}</div>
--     article-body            <div class="article-body__content sprite-bullets">{{.content}}</div>
-- Query, for re-checking after ANY template edit (the 283 lane is rewriting this column
-- fleet-wide, so this fact has a shelf life):
--     SELECT function, html_template ~* '<p[^>]*>\s*\{\{\s*\.?\s*[Cc]ontent\s*\}\}'
--     FROM content_components
--     WHERE function IN ('generic-text-block','about-content','illustrated-text-block','article-body')
--       AND is_active;
-- All four must read false. `about-content` DOES contain `<p>` elsewhere (in its
-- highlights block) — which is why the predicate must test the CONTAINER of the slot, not
-- the presence of a `<p>` anywhere in the template.
--
-- ⚠ NO GO PATH BRANCHES ON A FIELD TYPE OF `html` (council `ca400ba6`, guardian seat,
-- severity medium). Checked exhaustively rather than assumed: the only readers of a
-- schema field's declared type are `DeclaredTypeSatisfied` / `declaredTypeOf` /
-- `declaresArray` (`content_type_violations.go:261-285`, array/list only) and
-- `component_schema_fields.go:111-115`, which COPIES the type through into the writer's
-- field spec — i.e. into the prompt, which is the routing effect this file intends. Every
-- other `"html"` literal in platform/ and internal/ is a map key, an LLM step
-- `output_format`, a content-type sniff or a payload field, none of them a component field
-- type. So this retype changes what the WRITER is told and nothing else in the platform.
--
-- WHY THE FOUR GUIDANCE BLOCKS ARE NOT ONE SHARED STRING (council `ca400ba6`, reuse_agent
-- seat, severity low — a fair maintenance concern). They are deliberately NOT copies:
-- `about-content` must say "the separate highlights field already holds the short punchy
-- items, do not duplicate them here"; `illustrated-text-block` must forbid inline images
-- SPECIFICALLY because it owns image_url/image_alt/image_caption fields that an inline
-- <img> would bypass; `article-body` keeps <h2> because it is a whole article body rather
-- than a section slot. The shared half — what structure to use and what is forbidden —
-- IS centralised, in RULE 10 (595). Per-field guidance carries only what is true of that
-- field. A single shared string would have to drop all three of those clauses to be
-- shareable, and each one is the part that prevents a specific defect.
--
-- BLAST RADIUS. Four shared library rows, `[MEASURED 2026-08-24]` 181 + 153 + 27 + 6 =
-- 367 live instances. Existing instances are NOT rewritten by this file — stored
-- rendered_html and content_data are untouched; only future writes change. It touches
-- `input_schema` ONLY and never `html_template` (confirmed disjoint with the
-- RFC_032/bugs_open/283 lane, which has never edited input_schema and inserts no
-- content_components rows).
--
-- PAIRS WITH 595 AND SHOULD LAND WITH IT. Guidance alone works (article-body proves
-- it) but leaves RULE 9 contradicting the guidance on a now-`html` field. Neither
-- ordering is unsafe; both applied is the intended state.
--
-- SCOPE. Config/library data only. LIVE ON APPLY, no chassis roll.
-- ROLLBACK: 594_prose_slots_get_structure_guidance_ROLLBACK.sql

BEGIN;

-- ── Pre-state: exactly one active row each, and each target field is the
--    text/llm shape this file was written against ──────────────────────────
DO $$
DECLARE r record; n int; t text;
BEGIN
  FOR r IN SELECT * FROM (VALUES
      ('generic-text-block',     'content'),
      ('about-content',          'content'),
      ('illustrated-text-block', 'content'),
      ('article-body',           'content')
    ) AS v(fn, fld)
  LOOP
    SELECT count(*) INTO n FROM content_components
     WHERE function = r.fn AND is_active;
    IF n <> 1 THEN
      RAISE EXCEPTION '594: expected exactly 1 active row for %, found % — an id-scoped edit would be ambiguous', r.fn, n;
    END IF;

    SELECT input_schema->'fields'->r.fld->>'type' INTO t
      FROM content_components WHERE function = r.fn AND is_active;
    IF t IS NULL THEN
      RAISE EXCEPTION '594: %.% has no declared type — the schema is not the shape this file expects, refusing', r.fn, r.fld;
    END IF;
    IF t = 'html' THEN
      RAISE EXCEPTION '594: %.% is already typed html — already applied, refusing to double-apply', r.fn, r.fld;
    END IF;
    IF t <> 'text' THEN
      RAISE EXCEPTION '594: %.% is typed % (expected text) — another migration has changed it, refusing', r.fn, r.fld, t;
    END IF;
    IF (SELECT input_schema->'fields'->r.fld->>'source' FROM content_components WHERE function = r.fn AND is_active) <> 'llm' THEN
      RAISE EXCEPTION '594: %.% is not source=llm — it is not a writer slot, refusing', r.fn, r.fld;
    END IF;
  END LOOP;

  -- report-dossier is deliberately NOT in the loop. Assert it stayed out.
  IF (SELECT input_schema->'fields'->'body'->>'type' FROM content_components WHERE function = 'report-dossier' AND is_active) <> 'text' THEN
    RAISE NOTICE '594: report-dossier.body is no longer text — not touched by this file either way';
  END IF;
END $$;

-- ── generic-text-block.content — no guidance today; the owner's wall of text ──
UPDATE content_components
   SET input_schema = jsonb_set(
         jsonb_set(input_schema, '{fields,content,type}', '"html"'::jsonb),
         '{fields,content,llm_guidance}',
         to_jsonb($g$Write this section's body as HTML with real structure, not one long run of paragraphs. Use <h3> for a subheading whenever the material turns to a new point (in a long block, roughly every 150 words), <p> for paragraphs, <ul>/<ol> with <li> wherever the content is genuinely enumerable — months, steps, things to check, options being compared — <strong> for the term a reader is scanning for, <table> only when the data really is tabular, and <blockquote> for a quotation. Do not pad to earn a list: if the material is one idea, one paragraph is right. Never use <h1> or <h2> (the section supplies its own heading), and never emit <img>, <figure>, <iframe>, form controls, inputs, buttons, element ids, class attributes, inline styles or <script> — imagery and visual treatment belong to the component system, not to this text.$g$::text)
       ),
       updated_at = now()
 WHERE function = 'generic-text-block' AND is_active;

-- ── about-content.content — no guidance today ────────────────────────────────
UPDATE content_components
   SET input_schema = jsonb_set(
         jsonb_set(input_schema, '{fields,content,type}', '"html"'::jsonb),
         '{fields,content,llm_guidance}',
         to_jsonb($g$Write the about narrative as HTML with real structure. Use <h3> where the story turns (origins, approach, who it is for), <p> for paragraphs, <ul>/<ol> with <li> where the content is genuinely a set — capabilities, principles, stages — <strong> for the term a reader is scanning for, and <blockquote> for a quotation. Do not pad to earn a list. Never use <h1> or <h2> (the section supplies its own heading), and never emit <img>, <figure>, <iframe>, form controls, inputs, buttons, element ids, class attributes, inline styles or <script>. The separate `highlights` field already holds the short punchy items — do not duplicate them here.$g$::text)
       ),
       updated_at = now()
 WHERE function = 'about-content' AND is_active;

-- ── illustrated-text-block.content — guidance today RESTRICTS it to <p> ──────
UPDATE content_components
   SET input_schema = jsonb_set(
         jsonb_set(input_schema, '{fields,content,type}', '"html"'::jsonb),
         '{fields,content,llm_guidance}',
         to_jsonb($g$Prose for this section's own subject, as HTML. Paragraphs in <p>; where the content is genuinely enumerable use <ul>/<ol> with <li>; <strong> for the term a reader is scanning for; <h3> only if the block is long enough to need a turn. Do not pad to earn a list. Never use <h1> or <h2> (the section supplies its own heading), and never emit <img>, <figure> or <iframe> — this component already has its own image fields (`image_url`, `image_alt`, `image_caption`) and an image written into the prose would bypass them. No form controls, inputs, buttons, element ids, class attributes, inline styles or <script>.$g$::text)
       ),
       updated_at = now()
 WHERE function = 'illustrated-text-block' AND is_active;

-- ── article-body.content — the exemplar. Type only, plus the exclusions. ─────
-- Its guidance already works (76%) and is NOT rewritten: the only change is the
-- forbidden-elements sentence, which is the one thing it does not say and the one
-- thing the editorial lane needs it to say. <h2> stays permitted here and only here,
-- because this component IS a whole article body rather than a section slot.
UPDATE content_components
   SET input_schema = jsonb_set(
         jsonb_set(input_schema, '{fields,content,type}', '"html"'::jsonb),
         '{fields,content,llm_guidance}',
         to_jsonb(
           (input_schema->'fields'->'content'->>'llm_guidance')
           || $g$ Never emit <img>, <figure>, <iframe>, form controls, inputs, buttons, element ids, class attributes, inline styles or <script>: imagery and visual treatment belong to the component system, not inside this text.$g$
         )
       ),
       updated_at = now()
 WHERE function = 'article-body' AND is_active;

-- ── Verify ──────────────────────────────────────────────────────────────────
DO $$
DECLARE r record; t text; g text;
BEGIN
  FOR r IN SELECT * FROM (VALUES
      ('generic-text-block'), ('about-content'), ('illustrated-text-block'), ('article-body')
    ) AS v(fn)
  LOOP
    SELECT input_schema->'fields'->'content'->>'type',
           input_schema->'fields'->'content'->>'llm_guidance'
      INTO t, g
      FROM content_components WHERE function = r.fn AND is_active;

    -- The literal, not a type check: nothing downstream could ever tell us.
    IF t IS DISTINCT FROM 'html' THEN
      RAISE EXCEPTION '594 VERIFY: %.content reads back type=% — expected the literal html (DeclaredTypeSatisfied is default-TRUE and would never surface a typo)', r.fn, COALESCE(t, '<null>');
    END IF;
    IF g IS NULL OR length(g) < 100 THEN
      RAISE EXCEPTION '594 VERIFY: %.content has no substantial llm_guidance — the lever of this fix is missing', r.fn;
    END IF;
    IF position('<ul>' in g) = 0 AND position('<ul>/<ol>' in g) = 0 AND position('ul/ol' in g) = 0 THEN
      RAISE EXCEPTION '594 VERIFY: %.content guidance never mentions a list', r.fn;
    END IF;
    IF position('<img>' in g) = 0 THEN
      RAISE EXCEPTION '594 VERIFY: %.content guidance does not forbid <img> — the editorial lane''s condition is unmet', r.fn;
    END IF;
    -- The rest of the schema must be intact: content is not the only field on
    -- three of these four.
    IF (SELECT input_schema->'fields'->'content'->>'source' FROM content_components WHERE function = r.fn AND is_active) <> 'llm' THEN
      RAISE EXCEPTION '594 VERIFY: %.content lost its source=llm — jsonb_set has clobbered the field object', r.fn;
    END IF;
  END LOOP;

  -- Sibling fields survived (about-content has 3, illustrated-text-block has 5).
  IF (SELECT count(*) FROM content_components cc
        CROSS JOIN LATERAL jsonb_each(cc.input_schema->'fields')
       WHERE cc.function = 'illustrated-text-block' AND cc.is_active) <> 5 THEN
    RAISE EXCEPTION '594 VERIFY: illustrated-text-block no longer has 5 fields — sibling fields were lost';
  END IF;

  -- article-body kept its own proven wording rather than being overwritten.
  SELECT input_schema->'fields'->'content'->>'llm_guidance' INTO g
    FROM content_components WHERE function = 'article-body' AND is_active;
  IF position('Use h2 for main sections' in g) = 0 THEN
    RAISE EXCEPTION '594 VERIFY: article-body lost its original guidance — it is the exemplar and must only be APPENDED to';
  END IF;

  -- report-dossier untouched.
  IF (SELECT input_schema->'fields'->'body'->>'type' FROM content_components WHERE function = 'report-dossier' AND is_active) <> 'text' THEN
    RAISE EXCEPTION '594 VERIFY: report-dossier.body was modified — it is deliberately excluded (never LLM-authored)';
  END IF;

  RAISE NOTICE '594 OK: four prose slots typed html and told to use structure; report-dossier deliberately untouched';
END $$;

COMMIT;
