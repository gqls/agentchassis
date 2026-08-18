# PLAN 2026-08-18 — bug 184: mechanical markdown repair + write-seam prevention

Bug: `bugs_open/184_HANDOFF_2026-08-03_llm_markdown_reaches_the_page_as_literal_asterisks.md`
Lane: `bugfix_184_literal_markdown`. Status at adoption: detection LIVE and proven; repair
BROKEN (pair `literal_markdown → page-build-handler` 1 complete / 28 failed lifetime, held by
the migration-444 promoter floor); 71 open items on 6 sites, new detections daily; defect
confirmed served at the artefact (fundamentallyai.com/news, curl 2026-08-18).

## The decision, and its reason

**Stop using an LLM to fix a mechanical defect.** Removing markdown syntax from a
plain-text field is a deterministic string operation. Every LLM-shaped repair has failed
for a *demonstrated* reason: the regenerating writer has the same habit as the original
writer — proven at the artefact on 2026-08-07, when a full regeneration wrote 18 markdown
findings back into the very field it was dispatched to clean, three days after the
prompt-rule hardening (migration 304) went live. A prompt is not a control; a rebuild is
not a repair (bugfix_201 lane conclusion, recorded in the bug file).

So the fix is one shared, deterministic, **strip-only** transform, used three ways:

1. **Repair**: a page rerender (the existing, proven no-LLM path — 5,044 lifetime
   completes as a work-item handler) re-renders every section from stored `content_data`;
   with the transform applied at that seam, *a plain rerender becomes the repair*.
2. **Prevention**: the same transform at the LLM render seam normalises new content at
   birth, on both surfaces at once.
3. **Detection stays the tripwire**, widened to the forms now live (markdown links), and
   the completion verifier keeps gating honestly — both share the transform's patterns by
   construction, so detector, verifier and repair cannot drift.

Why this is the door-closer (memory: order fix candidates by what closes the door): the
defect class becomes unrepresentable on the covered write paths, the repair cannot
reintroduce the defect (it is not generative), and the check catches any writer that
bypasses the seams — including writers that do not exist yet.

### Why strip-only is safe where markdown→HTML was rejected

CQ-019 deferred normalise-on-write for two stated reasons. Both are answered, not
overridden:
- *"markdown→HTML conversion into an unescaping pipe is an injection surface"* — we never
  insert markup. The transform only deletes marker characters (`**`, backticks, `#{1,6} `,
  `[…](url)` → link text). The render pipe (`text/template`, zero escaping) receives a
  strict subset of the characters it receives today.
- *"blind mutation in shared SavePageSectionsAction changes what a shared save
  guarantees"* — the transform is (a) NOT in the save action at all, and (b) **opt-in via
  step config, default OFF** (owner ruling 2026-08-02 §2 shape; per RFC_022's narrowing,
  opt-in-default-OFF with no live consumer naming it is not architecture-scope). It is
  enabled by migration on named steps of named agents.

## Evidence base (all verified this session unless cited)

- Live open items: 34 unresolved / 24 failed / 10 detected / 3 needs_human_review, 6 sites
  [MEASURED 2026-08-18]. Findings by pattern: heading 107+106, code_span 49+8, bold 8+8.
- Served now: fundamentallyai.com/news/index.html — 11 raw `#` headings, 2 raw md links
  [MEASURED, curl 2026-08-18].
- Detector gap: `check_literal_markdown.go:107-114` has no link pattern; fleet content_data
  carries 9 md-link components (largest raw bucket) [MEASURED 2026-08-18]. Bullets: 0
  multi-line; italic: 0 — deliberately NOT added (letter-guard discipline; no evidence).
- `RenderComponentAction` window `v3_site_actions.go:1988-2012`: LLM field map in memory,
  `comp.InputSchema` in hand, precedent transform (`reconcileGeneratedItemKeys`, :2005)
  already sits there; the same map feeds render context AND persisted `content_data`
  (:2007). A transform here lands on both surfaces.
- Rerender path does NOT pass RenderComponentAction: `rerender_page_sections_action.go`
  renders via `RenderTemplate` from stored content_data (:486-491) and persists
  `mergedContent` (:543-564) through `save_page_sections`. So the transform needs its own
  hook there — which is exactly what makes rerender the repair.
- `page-rerender` is in `knownHandlerAgents` (handler_coverage_test.go:110); two checks
  already route at it (check_misdirected_cta — the `cta_links_stale` precedent — and
  check_contact_form_undeliverable); pair `page_rerender → page-rerender` is 5,044
  complete / 38 failed lifetime [MEASURED 2026-08-18].
- Live `check_rerender_mode` condition already accepts FOUR spec.reason values
  (image_landed, section_data_resolved, cta_links_stale, template_changed) — the
  vocabulary has grown before [MEASURED, live agent_definitions row 2026-08-18].
- Verifier `VerifyLiteralMarkdownResolved` is keyed on item_type, not handler — it
  survives the re-route and its whole-page remit still matches (rerender_sections rewrites
  every unlocked section). Claim-timeout exclusion (migration 331) also keys on item_type.
- Optional-key budget: `rerender_page_sections` is at 3 optional keys (budget N=10); the
  new flag is a ConfigKeys setting. `render_component` has no ActionInputSpec (uncounted
  bucket — posture unchanged).
- Residual risk sized: 1 live `content_components.html_template` fleet-wide matches
  markdown patterns; ≤9 open-item pages have rendered-only code spans [MEASURED
  2026-08-18]. Where the defect lives in a template, rerender cannot heal it — the
  verifier then blocks completion and routes to human review, which is honest; if canaries
  hit it, that is a separate bug to file (template hygiene), not a reason to weaken this.

### On the diagnosis-loop requirement (CLAUDE.md, owner ruling 2026-07-31)

No new root-cause claim is being filed. The mechanism ("the writer emits markdown into
text-typed fields; regeneration reintroduces it") was established by *first-hand
artefact-level experiment in production* on 2026-08-07 (bugfix_201 lane: full regeneration,
md5s changed, 18 findings re-introduced, completion blocked by the verifier) and is
recorded in the bug file. That is the declared substitute for a 090 run on an
already-demonstrated mechanism. The fix plan goes to the council gate, which is the right
reviewer for a fix.

## The edits (≤8, for the council submission)

1. **`platform/orchestration/datahelpers/literal_markdown.go` (new)** — single source of
   the pattern set (bold, code_span, heading, **md_link** new) + `ScanLiteralMarkdown`
   primitive + `StripLiteralMarkdown` (fixpoint-bounded, strip-only). Property test:
   for the live corpus + generated cases, scan(strip(x)) == ∅. datahelpers is imported by
   both `actions` and `discovery_checks`, so all consumers share one implementation.
2. **`check_literal_markdown.go`** — consume the shared patterns/scan (delete local vars);
   add `md_link` to the pattern vocabulary (content_data: only on markup-free values, like
   code_span; rendered_html: on assertion text). Re-point `HandlerAgent` to
   `"page-rerender"`; add `"reason": "literal_markdown"` to the item spec (the
   check_misdirected_cta precedent); rewrite the spec `fix` text (mechanical strip, not
   LLM rewording); CHANGES header entry. Verifier widens in lockstep automatically (shared
   scan).
3. **`v3_site_actions.go` (RenderComponentAction)** — in the :2005-:2007 window, when step
   config `strip_literal_markdown: true`, walk `contentData` string leaves (skip `_` keys)
   and apply the transform (code_span/md_link only on markup-free values). Info-log each
   strip with field path — the demand control that proves the seam fires.
4. **`rerender_page_sections_action.go`** — same gated transform applied to each loaded
   `storedSection.contentData` before planSection/render/persist; declare
   `strip_literal_markdown` in `RerenderPageSectionsInputSpec.ConfigKeys`.
5. **`section_editor_actions.go`** — same gated transform applied to the merged
   content_data map immediately before each of the two `RenderTemplate` calls
   (applyContentEdit :851, applyComponentSwap :949) — before render, so both surfaces are
   built from stripped values. (The :398-407 pre-persist window is too late: HTML is
   already rendered by then.)
6. **Migration `sql_for_agents/NNN_literal_markdown_mechanical_repair.sql`** — on
   `page-rerender`: extend `check_rerender_mode.condition` with
   `OR input_data.spec.reason == 'literal_markdown'`, and set
   `strip_literal_markdown: true` on `rerender_sections.config`. Fail-loud verify
   (DO/RAISE), backup table, idempotent guard. **Apply after the image is live.**
7. **Migration `sql_for_agents/NNN+1_writer_render_strip_optin.sql`** — set
   `strip_literal_markdown: true` on `page-content-writer`'s
   `process_sections_loop…render_section.config` (and on `section-editor`'s
   `apply_section_edit` step config). Same discipline. Apply after image.
8. **Docs/register in the same commit** — CQ-019 updated (status stale since 08-04; new
   seam + repair route recorded, per the ordering-exemption condition 2);
   `bugs_open/184` progress note; 016b §9 transferable pattern ("repair-by-regeneration
   cannot fix a defect the regenerator has the habit of producing — repair mechanically,
   prevent at the seam, keep the detector as the tripwire").

## Rollout (operational, after image + migrations live)

1. Pod-verify the binary (build-provenance stamp; `merge_base --is-ancestor`).
2. Apply migration 6, then 7. Run the optional-key parity test + budget audit.
3. **Canary**: hand-promote TWO items (one webdesign.co.uk news page, one robot-hands
   page): `status='triaged'`, `handler_agent='page-rerender'`, `attempt_count=0`,
   `spec = spec || '{"reason":"literal_markdown"}'`, keep `pipeline`, stamp
   `spec.original_pipeline` — the documented bootstrap path for a new pair (bugfix_277
   consumer notice). Two pages, because a stale page holds every improvement since it
   rendered — they must disagree to teach anything (memory rule).
4. Verify at the artefact (curl visible text, all four patterns), not at `result` — on
   these rows `result` can be the spawn record (bugs_open/287).
5. If clean: promote the remaining open population in batches; let the next discovery pass
   retract the leftovers (`Resolved` seam); watch the new pair's ratio establish itself.
6. Consumer notices: append to `bugs_open/184` (the shared account); note for the
   bugfix_277 lane that the held pair goes quiet and a fresh pair bootstraps — exactly
   their documented expectation.
7. Bug closes when: fixed AND live AND the founding rows + the widened-symptom pages
   verify clean at the artefact. Close-out must `git mv` with BOTH paths on the commit
   pathspec (ambiguous-number landmine, verified at HEAD via `git ls-tree`).

## What this deliberately does NOT do

- No markdown→HTML conversion, ever, anywhere (unescaped pipe).
- No transform in `SavePageSectionsAction` (the shared save's guarantees unchanged; the
  render-side seams cover both surfaces and the save ingests their output).
- No italic/bullet patterns (measured zero; letter-guard discipline over reach).
- No unhold of the old pair and no rollback of migration 444 — the floor did its job; the
  old pair simply stops receiving items.
- No hand-editing of `rendered_html` rows (the render regenerates it).

## Risks, stated

- **A strip could alter meaning** (e.g. `` `animation` `` → `animation`): bounded by
  letter-guarded patterns identical to the detector's — anything stripped would otherwise
  be SERVED as raw syntax, which is strictly worse; and every strip is logged with its
  field path.
- **Md-link strip drops the URL**: in a plain-text field the URL was never clickable —
  the text is the only faithful plain-text rendering. Logged when it fires.
- **Template-borne markdown** (1 fleet-wide): rerender cannot heal; verifier routes to
  human review honestly; file separately if canaries hit it.
- **Detector widening creates new findings**: with a working repair this drains rather
  than floods; new detections carry the new handler and dispatch normally (fresh pair, no
  floor hold).
- **Opt-in coverage gaps** (a writer without the flag): the check remains the tripwire;
  RFC_022's budget machinery watches optional-key accumulation.
