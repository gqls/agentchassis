-- 595 — page-content-writer: RULE 10 is addressed to a type that exists (bugs_open/381, arm B2)
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
-- WHY. The writer's rulebook has exactly two sentences about markup, and between them
-- they instruct every prose slot on the estate into paragraphs:
--
--   RULE 9  "For fields of type `text`: return a plain string with no HTML wrapping."
--   RULE 10 "For fields of type `rich_text` or `content` that contain multiple
--            paragraphs: use proper HTML markup, wrapping each paragraph in <p> tags"
--
-- `[MEASURED 2026-08-24]` across active section components: 940 llm fields are typed
-- `text` (135 components), 2 are `html`, and `rich_text` and `content` are declared
-- **ZERO** times. So the only rule that permits structure is addressed to nobody, the
-- rule that forbids it covers essentially everything, and the most structure RULE 10
-- would have permitted even if it were reachable is a <p>.
--
-- WHAT CHANGES. Migration 594 retypes four pass-through prose slots to `html`. This
-- file re-addresses RULE 10 to `html` and tells the writer what structure means, and
-- narrows RULE 9 so it stops contradicting a field's own guidance.
--
-- ⚠ RULE 9 IS NARROWED, NOT WEAKENED. Its real job — and 304's addition to it — is the
-- markdown ban, which stays absolute and is restated. What it stops asserting is that
-- a `text` field may never contain markup, because that was already false in practice:
-- `article-body.content` is typed `text`, carries guidance asking for headings and
-- lists, and renders a list in 76% of its instances. RULE 9 was being ignored where a
-- field's own guidance contradicted it, so the rulebook is being made to match what
-- the estate already does rather than the reverse.
--
-- THE EXAMPLES ARE DELIBERATELY NOT FIRST-PERSON PRACTICE CLAIMS. "Months of the year,
-- steps a reader takes, options being compared" — never "our testing process, step by
-- step". Coordinated with the bugs_open/380 lane, whose writer arm forbids exactly
-- that shape on a site with no operating history; an example inviting it here would
-- have pulled against their fix.
--
-- ⚠ 304's ROLLBACK WILL REFUSE AFTER THIS FILE, BY DESIGN. 304
-- (`304_forbid_markdown_in_text_fields.sql:65`) guards on the literal
-- "10. For fields of type `rich_text` or `content`", which this file replaces. That is
-- a refusal, not a corruption: 304's rollback declines rather than mis-splicing. This
-- file's own rollback restores both rules verbatim, so the pair is recoverable.
--
-- ⚠ TIMING, from the 305 lane (2026-08-24). Their define-by-negation scanner splits
-- sentences on `</p`, `<br`, `</li`, `</h`, `</div`, `</td` — and `</th` was MISSING
-- (it is not covered by the `</h` arm: the third character differs). A
-- define-by-negation construction inside a table HEADER cell therefore produced a
-- markup-bearing "sentence", and their repair splices over exactly that span, which
-- would have replaced the cell tags with prose and broken the table. Found by asking
-- them about this change; fixed by that lane in 714789d7b, mutation-proven — but it is
-- Go, so it is INERT UNTIL THE NEXT CHASSIS ROLL. PREFER APPLYING THIS FILE AFTER THAT
-- ROLL. The exposure is narrow (it needs the construction inside a <th> specifically),
-- lists and subheads were already safe, and nothing here is unsafe on its own.
--
-- ⚠ THE ANCHOR GUARDS ABORT THE TRANSACTION, THEY DO NOT WARN (council `ca400ba6`,
-- debug_historian seat, severity medium — the objection was that the submission's sketch
-- did not SHOW the guard failing). It does: the pre-state block below RAISEs EXCEPTION
-- unless RULE 9 and RULE 10 each appear EXACTLY ONCE verbatim, inside the same transaction
-- as the UPDATE, so a `replace()` that would silently no-op on a moved anchor while still
-- reporting `UPDATE 1` can never reach COMMIT. Mutation-proven 2026-08-24: with RULE 10
-- edited to a different spelling, this file refuses with
-- '595: RULE 10 does not appear exactly once verbatim'. Same discipline in 594, whose
-- per-field pre-state loop runs inside its own transaction for the same reason.
--
-- PAIRS WITH 594. Guidance is the lever that moves behaviour (see 594's header); this
-- file removes the rulebook's contradiction of it. Neither ordering is unsafe.
--
-- SCOPE. Config-only, one prompt, two anchored replaces. LIVE ON APPLY, no roll.
-- Anchored replace() with exact-count prechecks — never a whole-prompt rewrite, which
-- would silently revert the bugs_open/380 lane's concurrent edits to the same prompt
-- (their anchors: the evidence_base writer_block block; mine: rules 9 and 10 — disjoint,
-- agreed 2026-08-24).
--
-- ROLLBACK: 595_writer_rules_9_and_10_name_html_ROLLBACK.sql

BEGIN;

SELECT snapshot_agent('page-content-writer', '595_writer_rules_9_and_10_name_html: pre-update');

DO $$
DECLARE n int; p text;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'page-content-writer'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '595: expected exactly 1 live page-content-writer row, found %', n;
  END IF;

  SELECT default_config#>>'{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}'
    INTO p FROM agent_definitions WHERE id = '5946a27b-38ab-41e8-8b49-7bc1a4b626b8';
  IF p IS NULL THEN
    RAISE EXCEPTION '595: generate_content.config.prompt_template is NULL — the loop sub-workflow has changed under me, refusing';
  END IF;

  IF (length(p) - length(replace(p, '9. For fields of type `text`: return a plain string with no HTML wrapping.', '')))
     / length('9. For fields of type `text`: return a plain string with no HTML wrapping.') <> 1 THEN
    RAISE EXCEPTION '595: RULE 9 does not appear exactly once verbatim — refusing to splice blind';
  END IF;
  IF (length(p) - length(replace(p, '10. For fields of type `rich_text` or `content` that contain multiple paragraphs:', '')))
     / length('10. For fields of type `rich_text` or `content` that contain multiple paragraphs:') <> 1 THEN
    RAISE EXCEPTION '595: RULE 10 does not appear exactly once verbatim — 304 or another migration has moved it, refusing';
  END IF;
  -- 304's markdown ban must still be inside rule 9; this file restates it and would
  -- otherwise silently drop it.
  IF position('Plain string also means NO markdown syntax' in p) = 0 THEN
    RAISE EXCEPTION '595: 304''s markdown ban is missing from RULE 9 — the prompt is not the shape this file was written against, refusing';
  END IF;
  IF position('10. For fields of type `html`' in p) > 0 THEN
    RAISE EXCEPTION '595: already applied — refusing to double-apply';
  END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}',
         to_jsonb(
           replace(
             replace(
               default_config#>>'{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}',
               $old10$10. For fields of type `rich_text` or `content` that contain multiple paragraphs: use proper HTML markup, wrapping each paragraph in <p> tags: <p>Paragraph 1</p><p>Paragraph 2</p>.$old10$,
               $new10$10. For fields of type `html`: write real structure, not a run of paragraphs. Use <h3> for a subheading each time the material turns to a new point (in a long block, roughly every 150 words); <p> for paragraphs; <ul> or <ol> with <li> wherever the content is genuinely enumerable — the months of the year, the steps a reader takes, the things to check before deciding, the options being compared; <strong> for the term a reader is scanning for; <table> only when the data really is tabular; <blockquote> for a quotation. Do NOT pad to earn a list: if the material is one idea, one paragraph is the right answer, and a three-item list of near-identical clauses is worse than the paragraph it replaced. Never use <h1> or <h2> — the section supplies its own heading — and never emit <img>, <figure>, <iframe>, form controls, inputs, buttons, element ids, class attributes, inline styles or <script>. Where this field's own description gives more specific guidance, that description wins.$new10$
             ),
             $old9$9. For fields of type `text`: return a plain string with no HTML wrapping. The template handles paragraph wrapping for these fields.$old9$,
             $new9$9. For fields of type `text`: return a plain string — a heading, a label, a sentence — with no HTML wrapping around it. The template handles the wrapping for these fields. If this field's own description asks for particular markup inside it, follow the description: it knows what its slot renders. (Structure belongs in `html` fields — see rule 10.)$new9$
           )
         )
       ),
       updated_at = now()
 WHERE id = '5946a27b-38ab-41e8-8b49-7bc1a4b626b8'
   AND type = 'page-content-writer'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE p text;
BEGIN
  SELECT default_config#>>'{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}'
    INTO p FROM agent_definitions WHERE id = '5946a27b-38ab-41e8-8b49-7bc1a4b626b8';

  IF position('10. For fields of type `html`: write real structure' in p) = 0 THEN
    RAISE EXCEPTION '595 VERIFY: RULE 10 was not re-addressed to html';
  END IF;
  IF position('rich_text' in p) > 0 THEN
    RAISE EXCEPTION '595 VERIFY: the prompt still names rich_text — a type no component declares';
  END IF;
  IF position('9. For fields of type `text`: return a plain string — a heading' in p) = 0 THEN
    RAISE EXCEPTION '595 VERIFY: RULE 9 was not narrowed';
  END IF;
  -- 304's markdown ban must have SURVIVED: it is a separate sentence appended to
  -- rule 9 and the replace above must not have swallowed it.
  IF position('Plain string also means NO markdown syntax' in p) = 0 THEN
    RAISE EXCEPTION '595 VERIFY: 304''s markdown ban was destroyed by the rule 9 replace';
  END IF;
  IF position('<ul> or <ol> with <li> wherever the content is genuinely enumerable' in p) = 0 THEN
    RAISE EXCEPTION '595 VERIFY: the list instruction is absent';
  END IF;
  IF position('never emit <img>, <figure>, <iframe>' in p) = 0 THEN
    RAISE EXCEPTION '595 VERIFY: the forbidden-elements sentence is absent';
  END IF;
  -- The two rules must each appear exactly once and in order.
  IF position('9. For fields of type `text`' in p) > position('10. For fields of type `html`' in p) THEN
    RAISE EXCEPTION '595 VERIFY: rules 9 and 10 are out of order';
  END IF;
  -- Nothing else in the prompt was disturbed: three landmarks that belong to other
  -- lanes and must still be there.
  IF position('STRICT RULE -- NEVER PROMISE ACCURACY YOU CANNOT GUARANTEE.' in p) = 0
     OR position('19. Never write the words' in p) = 0
     OR position('{{range .current_section.llm_field_specs}}' in p) = 0 THEN
    RAISE EXCEPTION '595 VERIFY: an unrelated part of the prompt is missing — this file must only touch rules 9 and 10';
  END IF;

  RAISE NOTICE '595 OK: rule 10 addresses html and asks for structure; rule 9 narrowed, markdown ban intact';
END $$;

COMMIT;
