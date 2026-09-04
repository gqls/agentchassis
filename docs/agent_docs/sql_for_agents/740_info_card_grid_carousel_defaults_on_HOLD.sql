-- 740_info_card_grid_carousel_defaults_on.sql
--
-- OWNER RULING 2026-09-03 ("switch the switches"): `info-card-grid`'s `carousel`
-- flag is to be DEFAULT-ON. [MEASURED 2026-09-03 12:41:06Z] it is set on
-- 1 of 40 live instances (deployed+active pages) across 21 sites.
--
-- ══ THE OPEN DESIGN QUESTION, ANSWERED ═══════════════════════════════════════
-- The handoff left it open: "where does the default live — schema default plus a
-- backfill of the instances, or resolution-time?" It is RESOLUTION-TIME, and the
-- mechanism already exists. `plan_sections_action.go:2886` (the renderer/static
-- branch), verbatim:
--
--     if source == "renderer" || source == "static" || ... {
--         if !carryStored() && fallback != nil {
--             resolvedData[fieldName] = fallback
--         }
--         continue
--     }
--
-- So a `fallback` on a `source: static` field IS the default, applied per render,
-- and a stored value BEATS it. No backfill is needed and none is done here.
--
-- ⚠ NO BACKFILL IS NEEDED ONLY BECAUSE THE KEY IS ABSENT, NOT FALSE.
-- [MEASURED 2026-09-03 12:41:06Z] of the 40 live instances: 1 carries
-- `carousel: true`, 0 carry `carousel: false`, 39 do not carry the key at all.
-- That matters because `carryStored()` -> `storedFieldValue()` rejects only
-- values `IsEmptyContentValue` calls empty, and a Go `bool` hits that function's
-- default arm — so a stored `false` is NOT empty, IS carried, and WOULD beat this
-- fallback. Had the 39 stored an explicit false, this migration would have
-- applied cleanly, verified green and changed nothing. Re-run the census before
-- assuming that is still true.
--
-- ══ THE MECHANISM IS PROVEN LIVE, NOT JUST READ ══════════════════════════════
-- [MEASURED 2026-09-03] `ai-readiness-quiz` is the existing worked example on the
-- SECTION path (not chrome): 12 `source: static` fields with fallbacks, and both
-- live instances (finetuning.uk, leopardessconsulting.co.uk) carry the fallback
-- text persisted into `page_components.content_data` byte-for-byte —
-- `quiz_back_label = 'Back'`, `quiz_badge_label = 'AI Readiness Assessment'`.
-- Static fields are never offered to the writer, so those values can only have
-- come from this path.
--
-- ══ THE PRE-FLIGHT GATE — the arrows would be inert without the JS ═══════════
-- The field's own guidance says the layout "requires the hero-card-carousel
-- js_snippet to be bundled for the site ... without it ... only the arrows are
-- inert". Turning arrows on for a site that cannot drive them is worse than the
-- grid, so this was measured AT THE ARTEFACT before writing the migration:
--   * `js_snippets` has ONE matching row, `hero-card-carousel`, is_active,
--     applies_to = ["hero-card-carousel", "info-card-grid"] — so the bundle
--     follows the component, not the flag.
--   * [MEASURED 2026-09-03] all 21 sites carrying info-card-grid serve
--     /assets/js/snippets.js at HTTP 200 with 15 `data-hcc` occurrences each.
--   * NEGATIVE CONTROL, because a constant 15 could mean the grep matches
--     something always present: 6 sites WITHOUT info-card-grid serve 0 — and one
--     of them (fundamentallyai.com) has a 10,928-byte bundle, so it is not
--     "small bundle = zero". The instrument discriminates.
--
-- ══ ⚠ THE ACCEPTANCE TEST IN THE HANDOFF IS WRONG. DO NOT USE IT ═════════════
-- HANDOFF_2026-09-03 §3.2 names `overflow-x` as the NEGATIVE control — "it reads
-- 2 on both ... a flip that moves `overflow-x` is doing something other than what
-- it says". That is FALSE, and following it would make a CORRECT flip read as a
-- defect.
-- The template's ONLY `overflow-x` (line 204) sits INSIDE the `{{if $.carousel}}`
-- style block, so a correct flip ADDS ONE per flipped instance. The equal count
-- of 2 was a coincidence of unrelated CSS on two different pages of two different
-- sites, [MEASURED 2026-09-03] at the served bytes:
--   * leopardess/services.html: 1 from `--trp-track-gap` (another component) +
--     1 from the info-card-grid carousel block itself (`--icg-track-gap`).
--   * designblog/index.html:   2 from `.category-strip`, emitted twice.
-- CORRECTED TEST — before/after on the SAME page:
--   POSITIVE (0 -> n): data-hcc-track, data-hcc-prev, data-hcc-next,
--                      data-hcc-slide, info-card-grid__grid--carousel,
--                      scroll-snap-type
--   EXPECTED TO MOVE:  overflow-x, +1 per flipped instance
--   NEGATIVE (must NOT move): the count of `info-card-grid__card` articles and
--                      the card titles — this is a LAYOUT change; the content
--                      must be byte-stable. A flip that changes the card count
--                      is doing something other than what it says.
--
-- ══ ROUND 2 — THE COUNCIL RETURNED REVISE, AND THE GATING OBJECTION WAS RIGHT ════
-- Corr `2ac895f3-ca82-4dbe-8f4e-3335a04b8925`, r1: 9 seats approve, gated by
-- `bug_historian` (HIGH). Its objection, fairly: this migration's whole safety claim is
-- "a stored value beats the fallback", and the LANDMINES register carries an entry keyed
-- to the IDENTICAL call site saying the opposite —
--
--   "A `static`-source field in a component's `input_schema` OVERWRITES your stored
--    `content_data` on every section resolve … footprint: rerender_page_sections /
--    the section-planner resolve pass"
--
-- I had asserted the precedence from a code read and a code comment and never reconciled
-- it against the register. If the landmine were current, an instance that later stored an
-- explicit `carousel: false` would be silently flipped BACK to true — not "made inert",
-- which is what my risk section assumed. That is a materially worse failure and the seat
-- was right to gate on it.
--
-- ⚠ THE RECONCILIATION: THE LANDMINE IS HALF STALE, AND THE HALF THAT BEARS ON THIS
-- MIGRATION IS THE STALE HALF. Both were true when written; the code moved underneath it.
--   * The landmine was measured and added **2026-08-03** (brochure lane, tool-cta
--     `secondary_cta_label`).
--   * `carryStored` entered `plan_sections_action.go` on **2026-08-11** (`d26c26a9a`,
--     bugs_open/238) — but NOT yet on the static branch.
--   * The renderer/**static** branch got its `carryStored()` call on **2026-08-14**
--     (`8f899cc8d`, "fix(268): renderer/static fields now reach the 238 carry"), i.e.
--     ELEVEN DAYS AFTER the landmine was written. `git log -S 'if !carryStored() &&
--     fallback != nil'` returns exactly that one commit.
--   * So for `source: static` the landmine describes pre-268 behaviour and is stale.
--     ⚠ **Its `query.*` half is STILL LIVE** — a query that resolves writes
--     `resolvedData[field]` directly and beats the stored value. Do not read this
--     reconciliation as retiring the whole entry.
--
-- AND NOT ONLY FROM GIT — the surviving values are visible in live data. [MEASURED
-- 2026-09-03] 11 live instances across 6 sites store a value for a `source: static` field
-- that DIFFERS from that field's schema fallback and is still there. The load-bearing ones
-- are the POST-FIX rows, because a pre-08-14 row proves nothing:
--     mortgagecalculator.co.uk/index  tool-list.card_link_label
--         schema fallback "Open tool" → stored "Work it out", updated_at 2026-09-01 02:44Z
--     mortgagecalculator.co.uk/index  tool-list.eyebrow_label
--         schema fallback "Our Tools" → stored "Calculators",  updated_at 2026-09-01 02:44Z
--     cookly.uk/index                 testimonials-modern.cta_url
--         schema fallback "/contact"  → stored "/contact.html", updated_at 2026-08-26 12:36Z
-- Rows written after the fix, carrying authored values against a live fallback. The
-- oufe.com rows in the same result are 2026-07-29 and prove nothing either way; they are
-- excluded from the claim rather than counted toward it.
--
-- ⚠ THE LANDMINE ITSELF IS BEING CORRECTED IN THE SAME COMMIT AS THIS REVISION. A stale
-- register entry that contradicts live behaviour is worse than no entry: it gated a
-- correct change here, and next time it will license a wrong one.
--
-- ── `bug_historian` MEDIUM (the string-vs-boolean footgun is a SHARED property, and this
--    migration guards only its own call site) — MEASURED, and the class is currently EMPTY.
-- [MEASURED 2026-09-03] fleet-wide, active components, fields declaring a `fallback`:
--     fallback JSON type | declared type | fields | components
--     string             | text          |  1893  | 131
--     string             | number        |    87  |  23
--     string             | image         |     8  |   8
--     string             | url           |     6  |   5
--     boolean            | boolean       |     2  |   2   ← the whole boolean population
--     string             | string / html |     2  |   2
-- Fields declaring `type: boolean` whose fallback is NOT a JSON boolean: **ZERO**. So the
-- footgun is real as a shape and has no live instance, across a population of two. A CHECK
-- constraint or an authoring lint over ~2,000 fields to guard 2 is disproportionate today.
-- What would change that: the boolean population growing past a handful, or a first
-- non-boolean fallback appearing. The query above is the detector; it is one statement and
-- it is recorded here so the next author can re-run it rather than re-derive it.
--
-- ── `editquality` MEDIUM (the UPDATE predicate was LOOSER than the drift guard's) — FIXED
--    below. The guard validated `source='static' AND type='boolean'`; the UPDATE matched on
--    name alone. Now identical, so no row can be written that the guard did not validate.
--
-- ── `debug_historian` LOW (no real pre-image, only derived counts) — FIXED: `_pre_740` now
--    stores the whole `input_schema`, so the exact prior value is recoverable within the
--    transaction and the rollback has something to be checked against.
--
-- ── `guardian` LOW (state HOW the fleet actually re-renders) — it is the existing
--    `page_rerender` queue, and it is busy: [MEASURED 2026-09-03 15:27Z] 9,723 complete
--    (latest 15:26:59Z today), 1,751 unresolved, 51 triaged. No new trigger is created or
--    needed. ⚠ But the fleet WILL sit in a mixed state — some instances carousel, some
--    grid — until each page's turn comes, and that is a visible-to-visitors interim, not a
--    silent one. A lane wanting a site sooner files `page_rerender` items for it. ⚠ And a
--    `complete` rerender is NOT evidence the layout moved (standing landmine): read the
--    served bytes with the corrected acceptance test above.
--
-- ── `tooling_provenance` MISSING (no evidence the author consulted the travelling docs for
--    this component) — CORRECT, I had not, and doing so found something the whole council
--    missed:
--
-- ⚠⚠ `info-card-grid` HAS A TRAVELLING PLAN WITH AN ACCEPTANCE FENCE, AND ONE OF ITS SIX
--    CHECKS IS `no_horizontal_overflow`, ON DESKTOP AND MOBILE.
--    `doc_plans`, subject_type=`component`, subject_key=`info-card-grid`, authored
--    2026-08-05 by lane `staged_component_build`. A carousel is by construction a
--    horizontally overflowing track, so the obvious worry is that this default fails the
--    component's own fence on every placement.
--    IT DOES NOT, and the reason is explicit in the checker rather than inferred:
--    `internal/adapters/browserrunner/run_checks_action.go:1094-1104`, the `cut` predicate,
--    verbatim —
--        for (let n = el.parentElement; n; n = n.parentElement) {
--            const o = getComputedStyle(n).overflowX;
--            if (o === 'auto' || o === 'scroll') return false;
--        }
--    with the comment "a scroll container makes the width reachable, and is the standard
--    fix for wide tables, which must then pass this very check." The carousel sets
--    `overflow-x: auto` on `.info-card-grid__grid--carousel`, the DIRECT parent of every
--    `.info-card-grid__card`, so the cards are exempt; and the track is width-constrained
--    inside `.info-card-grid__inner`, so the document itself does not scroll (the check's
--    other clause). The arrows are `position: absolute`, which the same predicate exempts.
--    ⚠ **STATED AS A MECHANISM READ, NOT A RUN.** The recorded acceptance pass for this
--    component (`doc_notes`, categories `acceptance-run`, 2026-08-05, 10 of 10 across both
--    profiles) was taken on ai-agent-orchestration.com/services.html — a **flag-unset,
--    grid** placement. **The fence has never run against a carousel-enabled instance.**
--    So the honest position is that the mechanism says it passes and nothing has tested it.
--    ⚠ WHOEVER APPLIES THIS: re-run the component fence against a flipped placement and
--    record the result. That is the one check this migration cannot make for itself.
--
-- ══ ROUND 3: APPROVED — and two of its advisories CORRECT THIS FILE'S OWN CLAIMS ═══════
-- Corr `2ac895f3-ca82-4dbe-8f4e-3335a04b8925`, 2026-09-03 16:02:39Z, "approved with 4
-- advisory objection(s) — none high-severity". Approved, and two advisories are right about
-- things asserted above, so they are corrected here rather than left to the next reader.
--
-- ⚠ 1. `render_guardian` (medium) — **"THE EXISTING QUEUE LANDS IT" WAS THE WRONG
--    DENOMINATOR, AND THE SEAT HAD TO TELL ME.** This file cites the `page_rerender` queue's
--    busy-ness as evidence the default reaches the fleet, grounded only on
--    `rerender_page_sections_action.go:1450` — the SCOPED path. **Assemble-mode**
--    `page_rerender` (a `page_id` with no `spec.reason`) re-embeds each section's EXISTING
--    stored HTML and never re-renders the template against a re-resolved schema, so it can
--    never apply this fallback. I had not checked the split. [MEASURED 2026-09-03]:
--        status      items   carries spec.reason (scoped)   assemble-only
--        complete    9,781   1,264                          8,517  (87%)
--        unresolved  1,751   1,712  (98%)                      39
--    **87% of the completes I cited are a mode that cannot land this change.** The busy-ness
--    was real and was the wrong evidence for the claim. The forward-looking news is better
--    and is a DIFFERENT fact: the pending queue is 98% scoped. **Quote the unresolved split,
--    never the completes.** Consequence: the mixed carousel/grid interim's length depends on
--    SCOPED rerenders reaching each of the 21 sites, not on general throughput.
--
-- ⚠ 2. `editquality` + `guardian` + `tooling_provenance` — file a work item to force the
--    acceptance-fence run. **I NAMED THE WRONG VEHICLE AND THEN MEASURED IT.** Round 3 called
--    `acceptance_run` "the right vehicle (277 rows)" — named from the type name and a row
--    count, without checking what the rows ARE. [MEASURED 2026-09-03] **all 277 are
--    `handler_agent='tool-acceptance-agent'` with `spec.check` of `tool_acceptance_due` or
--    `manual` — it is a TOOL vehicle**, and this component's own 2026-08-05 fence run was
--    driven by **no work item at all**. **So no row was filed, deliberately:** routing a
--    component fence at a handler I cannot show runs component fences is exactly what
--    `bugs_open/395`'s routing rule 3b exists to stop, and an unhandled row is worse than a
--    recorded gap. What the seats actually wanted — the decision and its open follow-through
--    carried somewhere durable — **is written to `doc_notes`**, `subject_type='component'`,
--    `subject_key='info-card-grid'`, categories `decision`/`council-approved`/
--    `open-follow-through`, 2026-09-03 16:12:33Z. **Still owed: a component-capable
--    acceptance vehicle, or a hand-run of the fence against
--    `leopardessconsulting.co.uk/services.html`** — the one live carousel placement, which
--    means this can be answered BEFORE apply.
--    ⚠⚠ **THAT PAGE IS A PAYING CLIENT'S LIVE SITE**, flagged by the copy_quality_two_stage
--    lane and confirmed in the owner's own record (`about_page_commercial/
--    PLAN_2026-07-24_about_page_commercial.md` D4, verbatim: *"a paying client's site
--    (leopardess)"*). **It remains the correct target, on measured grounds: THE FENCE WRITES
--    NOTHING.** `internal/adapters/browserrunner/` contains zero `INSERT INTO`/`UPDATE … SET`;
--    `check_tool_acceptance.go` files no work item and triggers no rerender on failure;
--    screenshots go to the runner's own store. What it does to the site is an HTTP GET of an
--    already-served page.
--    **AND THE RISK ORDERING INVERTS.** [MEASURED 2026-09-03 16:15:50Z] leopardess/services is
--    the ONLY live placement with the `carousel` key set at all, so a "safer" portfolio target
--    (finetuning.uk has one `info-card-grid` instance and no flag) would first require a
--    `content_data` write plus a rerender — **choosing the portfolio site to avoid touching a
--    client site means performing the very write the concern is about, aimed elsewhere.** The
--    client page is lower-impact precisely because nothing has to change for it to be testable.
--    ⚠ **This grounding EXPIRES if the fence gains a write.** ⚠⚠ **AND THE STANDING CHECK IS
--    NOT "does it contain a write statement" — it is "does it have a DATABASE EXECUTION PATH
--    AT ALL",** which the peer lane established independently and is the stronger form because
--    a write statement can be added in one line while a DB dependency cannot be wired in
--    accidentally. [MEASURED 2026-09-03] across `internal/adapters/browserrunner/*.go`
--    (non-test): `.Exec(` **0** · `.ExecContext(` **0** · `.Query(` **0** · `.QueryContext(`
--    **0** · `sql.DB` **0** · `pgxpool` **0** · `database/sql` **0**. **Control, because a
--    grep that returns 0 everywhere proves nothing:** `sql.DB` in
--    `plan_sections_action.go` returns **11**. So the greps discriminate.
--    **Re-run that set before hand-running the fence**; if any of them stops returning 0,
--    create a carousel placement on a portfolio site and accept the write there instead.
--
--    ⚠ **AND A CENSUS TRAP FOR WHOEVER RE-RUNS THE "only placement" CLAIM.** The peer's first
--    pass used `content_data::text ILIKE '%carousel%'` and got **5 placements across 4 sites**,
--    which reads as a refutation and would have sent this fence to a portfolio page that both
--    needed the write AND does not carry the component. [MEASURED 2026-09-03] resolving the
--    VALUE rather than the text: only `leopardessconsulting.co.uk` `info-card-grid` has the key
--    present and `true`; `ai-agent-orchestration.com` (`case-studies-grid`), `seotools.co.uk`
--    (`Generic Text Block`, ×2) and `webdesign.co.uk` (`article-body`) merely MENTION the word
--    in a different component's prose. **Test `content_data ? 'carousel'` and read the value —
--    never `::text ILIKE`.** Same "true number over the wrong population" shape as the 87%
--    above, one table along.
--    Full reasoning: `doc_notes` subject_type=`component`, subject_key=`info-card-grid`,
--    2026-09-03 16:16Z.
--
-- 3. `bug_historian` (medium/low) and `debug_historian` (low) are registered, not disputed:
--    a shared-component default with no staged rollout fans out to every dependent instance
--    at once (inherent to the mechanism and the owner's stated intent), and `_pre_740` is an
--    in-transaction pre-image rather than a `pg_dump` — it guards a wrong write inside this
--    transaction, not a later unrelated clobber of a contended row.
--
-- ══ ⚠⚠ BEFORE YOU APPLY THIS: IT WILL CAROUSEL A SECTION ON A PAID CUSTOMER'S ═════
--        INDEX PAGE THAT NOBODY ASKED TO BE A CAROUSEL
-- [MEASURED 2026-09-04] The owner asked (via the boxingonline lane, 2026-09-04) for the
-- LATEST-ARTICLES section on `boxingonline.com/index` to become a carousel. That section is
-- **`content-listing`** ("Latest from the site"), whose schema has **NO** `carousel` field.
-- The SAME PAGE also carries an **`info-card-grid`** ("A few places to start"), which does.
--
-- **So applying this migration and letting that page rebuild puts a carousel on his index
-- page in the WRONG SECTION** — and he could reasonably read the request as actioned. It is
-- not hypothetical: `boxingonline.com/index` is `build_status='needs_rebuild'` as of
-- 2026-09-04, so a rebuild is already pending and this default would be picked up by it.
--
-- **This does NOT block the migration** — the owner's default-on ruling stands on its own and
-- the other 20 sites are unaffected by boxingonline's request. **But whoever applies it owes
-- the boxingonline lane (`site_delivery_and_editor`) a word first**, because on that one site
-- the visible effect will be mistaken for an answer to a different request. The real request
-- needs `content-listing` to gain its own opt-in carousel — a separate, unbuilt change
-- (its template has ZERO carousel affordance today; 19 instances across 11 sites).
--
-- ⚠ AND A CENSUS NOTE THAT COST ME THE FIRST ANSWER: querying that page with
-- `build_status='deployed' AND status='active'` — the pairing this file's own header
-- recommends — returns **ZERO ROWS for the whole page**, because the page is `needs_rebuild`.
-- That pairing answers *"what is being SERVED"*. For *"what is this page MADE OF"* the filter
-- is `status='active'` ALONE. Same table, same landmine, opposite direction: the safe-looking
-- filter produced a confident zero.
--
-- ══ WHAT CHANGES, AND WHEN ═══════════════════════════════════════════════════
-- Config: live on apply, no image build. But nothing on a served page moves
-- until that page re-renders. The re-render path DOES apply this —
-- `rerender_page_sections_action.go:1450` calls the same `planSection`, and its
-- own comment records that `plan.ResolvedData` merges LAST and wins over stored
-- content_data. So a page-rerender is sufficient; a full rebuild is not needed.
-- A completed rerender is NOT evidence: read the served bytes with the test above.
--
-- IDEMPOTENT BY CONSTRUCTION. `jsonb_set` to a fixed value; running it twice
-- leaves the same single key. Stated because THIS LANE shipped the opposite two
-- days ago: 723 used `replace()` on text that re-embedded its own anchor, so a
-- second run would stack a second copy of the guidance. A JSON path write cannot
-- do that.
--
-- Reversible: 740_..._ROLLBACK.sql removes the fallback key.
-- Source: HANDOFF_2026-09-03 §3.2, owner ruling relayed via designblog.co.uk.

BEGIN;

-- DRIFT GUARD. Abort rather than clobber if the component is not in the state
-- this migration was written against.
DO $$
DECLARE n int; tpl text;
BEGIN
    SELECT count(*) INTO n
      FROM content_components
     WHERE is_active AND name = 'info-card-grid'
       AND input_schema->'fields'->'carousel'->>'source' = 'static'
       AND input_schema->'fields'->'carousel'->>'type'   = 'boolean';
    IF n <> 1 THEN
        RAISE EXCEPTION
            'ABORT: expected exactly 1 active info-card-grid declaring a static boolean '
            '`carousel` field, found %. The component or its schema has moved.', n;
    END IF;

    SELECT count(*) INTO n
      FROM content_components
     WHERE is_active AND name = 'info-card-grid'
       AND input_schema->'fields'->'carousel' ? 'fallback';
    IF n <> 0 THEN
        RAISE EXCEPTION
            'ABORT: `carousel` already declares a fallback — another session has edited '
            'it, or this migration has ALREADY applied. Re-read before re-running.';
    END IF;

    -- The template gate is the whole point: a fallback on a field no template
    -- reads is inert config, and inert config verifies green.
    SELECT html_template INTO tpl FROM content_components
     WHERE is_active AND name = 'info-card-grid';
    IF position('{{if $.carousel}}' in tpl) = 0 THEN
        RAISE EXCEPTION
            'ABORT: the info-card-grid template no longer gates on {{if $.carousel}} — '
            'setting the default would change nothing while verifying green.';
    END IF;
END $$;

-- Pre-image: the schema must gain exactly ONE key, inside the carousel
-- descriptor, and the `fields` set itself must NOT change. jsonb_set on the
-- wrong path satisfies every named assertion below while destroying the schema.
-- Pre-image. `schema_before` is the WHOLE prior input_schema, not just derived counts
-- (debug_historian r1, low: the needle-gate convention calls for a real pre-image). The
-- counts are kept alongside it because they are what the assertions read.
CREATE TEMP TABLE _pre_740 ON COMMIT DROP AS
SELECT id,
       input_schema AS schema_before,
       (SELECT count(*) FROM jsonb_object_keys(input_schema->'fields')) AS n_fields,
       (SELECT count(*) FROM jsonb_object_keys(input_schema->'fields'->'carousel')) AS n_carousel_keys,
       input_schema->'fields'->'carousel'->>'llm_guidance' AS guidance
  FROM content_components
 WHERE is_active AND name = 'info-card-grid';

UPDATE content_components
   SET input_schema = jsonb_set(
           input_schema,
           '{fields,carousel,fallback}',
           'true'::jsonb,       -- JSON boolean, NOT the string "true": a Go
                                -- template {{if}} treats ANY non-empty string as
                                -- truthy, so "false" would render a carousel too.
           true),
       updated_at = now()
 -- PREDICATE MIRRORS THE DRIFT GUARD EXACTLY (editquality r1, medium). It used to
 -- match on name alone, which is LOOSER than what the guard validated — so a row
 -- the guard never inspected could have been written.
 WHERE is_active AND name = 'info-card-grid'
   AND input_schema->'fields'->'carousel'->>'source' = 'static'
   AND input_schema->'fields'->'carousel'->>'type'   = 'boolean';

-- VERIFY. DO/RAISE, not SELECTs: ON_ERROR_STOP does not fire on a non-empty
-- result set, so a block of SELECTs cannot stop the COMMIT.
DO $$
DECLARE
    rows_seen  int;
    is_true    boolean;
    is_boolean boolean;
    bad_fields int;
    bad_keys   int;
    lost_guid  int;
    drifted    int;
BEGIN
    SELECT count(*) INTO rows_seen
      FROM content_components WHERE is_active AND name = 'info-card-grid';

    SELECT (input_schema->'fields'->'carousel'->'fallback') = 'true'::jsonb,
           jsonb_typeof(input_schema->'fields'->'carousel'->'fallback') = 'boolean'
      INTO is_true, is_boolean
      FROM content_components WHERE is_active AND name = 'info-card-grid';

    -- The `fields` SET must be untouched: this adds a key one level deeper.
    SELECT count(*) INTO bad_fields
      FROM content_components cc JOIN _pre_740 pre ON pre.id = cc.id
     WHERE (SELECT count(*) FROM jsonb_object_keys(cc.input_schema->'fields')) <> pre.n_fields;

    -- The carousel descriptor must gain exactly one key (source/type/required/
    -- llm_guidance all survive).
    SELECT count(*) INTO bad_keys
      FROM content_components cc JOIN _pre_740 pre ON pre.id = cc.id
     WHERE (SELECT count(*) FROM jsonb_object_keys(cc.input_schema->'fields'->'carousel'))
           <> pre.n_carousel_keys + 1;

    -- Named survivor, because a count can be satisfied by a swap.
    SELECT count(*) INTO lost_guid
      FROM content_components cc JOIN _pre_740 pre ON pre.id = cc.id
     WHERE cc.input_schema->'fields'->'carousel'->>'llm_guidance' IS DISTINCT FROM pre.guidance;

    IF rows_seen <> 1 THEN
        RAISE EXCEPTION 'ABORT: % active info-card-grid rows after the UPDATE, expected 1', rows_seen;
    END IF;
    -- NULL here means the ROW is present but the PATH is not — i.e. jsonb_set
    -- wrote somewhere else. Distinguished from the missing-row case above,
    -- because "no row" is the wrong thing to tell someone whose path is wrong.
    IF is_true IS NULL THEN
        RAISE EXCEPTION 'ABORT: fields.carousel.fallback does not resolve after the UPDATE — '
                        'jsonb_set wrote to the wrong path';
    END IF;
    IF NOT is_boolean THEN
        RAISE EXCEPTION 'ABORT: the fallback is not a JSON boolean — a string "false" '
                        'renders a carousel, because {{if}} is truthy on any non-empty string';
    END IF;
    IF NOT is_true THEN
        RAISE EXCEPTION 'ABORT: the fallback did not land as true';
    END IF;
    IF bad_fields > 0 THEN
        RAISE EXCEPTION 'ABORT: the `fields` set changed size — jsonb_set hit the wrong path';
    END IF;
    IF bad_keys > 0 THEN
        RAISE EXCEPTION 'ABORT: the carousel descriptor does not have exactly one MORE key than before';
    END IF;
    IF lost_guid > 0 THEN
        RAISE EXCEPTION 'ABORT: the carousel llm_guidance did not survive the write';
    END IF;
    -- Whole-schema equality against the real pre-image: removing the one key we added
    -- must reproduce the prior schema BYTE-FOR-BYTE. This is the assertion the derived
    -- counts above cannot make — a same-count swap anywhere else in the document would
    -- satisfy every check so far and fail this one.
    SELECT count(*) INTO drifted
      FROM content_components cc JOIN _pre_740 pre ON pre.id = cc.id
     WHERE (cc.input_schema #- '{fields,carousel,fallback}') IS DISTINCT FROM pre.schema_before;
    IF drifted > 0 THEN
        RAISE EXCEPTION 'ABORT: the schema differs from its pre-image by more than the one '
                        'added fallback key — something else in the document changed';
    END IF;

    RAISE NOTICE '740: info-card-grid carousel now defaults ON at resolution time. '
                 '39 of 40 live instances carry no stored value and will flip on their next '
                 'render; the 1 with a stored true is unchanged. Read the SERVED bytes, and '
                 'note overflow-x is EXPECTED to move +1 (the handoff said otherwise).';
END $$;

COMMIT;
