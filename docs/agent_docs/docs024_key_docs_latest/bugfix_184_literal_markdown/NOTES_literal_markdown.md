# NOTES — bugfix 184 (literal markdown reaches the page) — append-only, newest at the bottom

## 2026-08-18 ~19:35Z — lane opened; ownership check; bug re-validated

- **Ownership**: `who-owns.py 184` says "OWNED or recently active", but read the detail:
  every recent commit is a *contribution* from another lane (`bugfix_201` 08-07 — "No action
  requested of this lane"; `bugfix_277` 08-17 — "This lane is not claiming the bug"). Grepped
  live session transcripts (`~/.claude/projects/.../\*.jsonl`) for `literal_markdown`: the one
  session with real hits (15) is the `bugfix_277` promoter-floor lane (migrations 466/471),
  touching the pair as *data*, not working 184. **Taking the bug.** [VERIFIED — transcript grep,
  2026-08-18 19:30]
- **Bug still valid** [MEASURED, live DB 2026-08-18]:
  - `site_work_items` `item_type='literal_markdown'`: 34 unresolved / 24 failed / 10 detected
    (newest **2026-08-18**, parked by the migration-444 promoter floor) / 3 complete /
    3 needs_human_review. 71 open across 6 sites; webdesign.co.uk carries 46, robot-hands.com 16.
  - Open items' findings by pattern: heading 107 (content_data) + 106 (rendered_html),
    code_span 49+8, bold 8+8. Samples include `## [` — headings that *contain markdown links*.
  - Fleet content_data quick scan (my own regexes, cruder than the check's): bold 1,
    heading 5, **md_link 9**, code_span 1 of 1,723 components. **Markdown links are now the
    largest raw bucket and the detector has NO link pattern** (check_literal_markdown.go:107-114
    covers bold/code_span/heading only). Confirms the 08-17 contribution's "symptom has widened"
    warning at the code level.
- Migration 304's prompt rule IS live (`rule9_extended = t` on `page-content-writer`), and the
  writer still emits markdown (bugfix_201's 08-07 artefact evidence: 18 findings re-written into
  the very field it was dispatched to clean). A prompt is not a control — again.
- Repair pair `literal_markdown → page-build-handler`: 1 complete / 28 failed lifetime (worst in
  fleet), HELD by the migration-444 success floor. New findings park at `detected`.
- Dispatched two Explore agents: (1) component write path / repair seam / field-type schema,
  (2) writer generation path + bugfix_201 lane conclusions + CQ-019 register state.

## 2026-08-18 ~19:45Z — artefact confirmation + status semantics

- **Served defect confirmed at the artefact** [MEASURED, curl 2026-08-18 19:42Z]:
  `fundamentallyai.com/news/index.html` visible text carries **11 ATX headings and 2 raw
  markdown links** right now. Not a stale-DB claim.
- `status='unresolved'` semantics: the two-strike rule (complete_work_item_verification.go /
  create_tool_cross_link_items.go:88 comment) parks a RE-detection as a non-dispatchable
  zombie once the page has ≥2 terminal failures; a stale-triaged 48h reaper also writes it.
  So the 34 unresolved rows are the same defect re-found on pages whose repairs already
  failed twice — consistent with "no working repairer exists".
- Diagnosis queue checked: no open `needs_diagnosis` item on this mechanism. No duplicate filing.

## 2026-08-18 ~20:20Z — research done, design fixed, part 1 committed, council submitted

- Two Explore agents mapped the machinery. Load-bearing facts (each verified in code by me
  before use): RenderComponentAction's :1988-:2012 window has the LLM map + schema in hand
  and feeds BOTH surfaces (v3_site_actions.go:2007 "capture before merge"); the no-LLM
  rerender does NOT pass RenderComponentAction (renders via RenderTemplate from stored
  content_data :486-491, persists mergedContent :543-564) — so it needs its own hook, and
  that hook is what makes a plain rerender the repair; `page-rerender` is in
  knownHandlerAgents (:110) with a 5,044-complete lifetime as an item handler; the live
  check_rerender_mode condition already grew to four reasons (read from the live row, not
  the seed).
- **Design**: strip-only shared primitive in datahelpers (scan + strip, property test
  scan(strip(x))==∅), md_link widening, HandlerAgent → page-rerender with spec.reason,
  opt-in default-OFF strip flags at three seams, migrations 473/474. Full reasoning:
  PLAN_2026-08-18_mechanical_markdown_repair.md. CQ-019's deferral of normalise-on-write
  answered (strip-only ≠ markdown→HTML; not in the save action; opt-in).
- **Council**: submitted, corr `060bcc0a-1ba5-4525-8fea-03de021e26f5` (~30 min budget).
- **SAME-FILE PASSENGER incident, handled**: rerender_page_sections_action.go carries the
  299 lane's uncommitted KEEP #3 hunk whose helpers (links_tel.go) are NOT at HEAD. Their
  session messaged me proposing "either order works" — WRONG in one detail: my strip hunk
  calls datahelpers.StripLiteralMarkdownFromContentData, not at HEAD at that moment, so
  their commit-first would have broken HEAD's build via MY passenger hunk. Resolution
  agreed by direct message: I committed the datahelpers primitive first (`019fb0616`,
  gofmt follow-up `5fbe549f7`) so my hunk compiles as their passenger; they commit next
  (naming my block as a passenger); I land part 2 (check re-route + rerender hook) after.
  **The check re-route is deliberately NOT at HEAD yet** — re-route without the strip hook
  in one image burns literal_markdown attempts on an unequipped handler.
- Migrations 473 (+ROLLBACK) and 474 (+ROLLBACK) authored; anchors verified against LIVE
  rows (page-content-writer step `render_section` action `render_component`;
  section-editor step `apply_edit` action `apply_section_edit`). Both safe pre-image
  (flags unread, reason unemitted); intended post-image.
- Register CQ-019 updated (status was stale two weeks — said "inert until 303" while
  303/304 were live since 08-04); bugs_open/184 progress note appended; 016b §9 entry
  appended ("repair-by-regeneration cannot fix a defect the regenerator has the habit of
  producing").
