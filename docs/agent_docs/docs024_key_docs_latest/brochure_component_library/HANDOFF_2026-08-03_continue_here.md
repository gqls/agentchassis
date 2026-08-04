# HANDOFF 2026-08-03 — brochure component library / fundamentallyai.com

**Supersedes `HANDOFF_2026-08-02_continue_here.md`.** That file is accurate about
TL-035's arming and proof; two of its facts have since moved (§4 below) and its §5 open
list is now mostly done. Owner instruction this session: *"do them all in the order you
choose"* — all four fronts were taken.

## 1. Where the lane is, in one paragraph

TL-035 is proven AND consumed: the renders were fetched (the `[UNFETCHED]` caveat is
spent), read by a person, and the reading found a camera defect no check asserts — the
shutter fires after the checks have driven the page. The eye half exists as
`scripts/contact_sheet.py` (TL-039), and the first sheet is published to the owner as a
private artifact. The 151 candidate-3 checker is **ENABLED** (seed 296) and its first
watched sweep deleted nothing and filed one flag-only capability_gap. `/tools.html`
exists; the dead CTAs work; both companion guides are built and live through the real
pipeline. What remains open is editorial, not technical (§5).

## 2. What was done, where the proof lives

- **151 enable** — seed `296_enable_content_duplication_on_completeness_discovery.sql`
  (applied by hand, guards induced first, ledger row `applied_by='record-only'`).
  Preconditions re-measured the same morning: guard strings pod-grepped on both
  v1.0.1238 replicas with a nonsense-string discriminator; shipped-rule census 0/0 over
  1,189 rows (`gauntlet_dead_cta/scripts/dedup_census_shipped.go` — re-run it, never
  quote it); plan repetition still exactly 1 (webdesign.co.uk/index/info-card-grid,
  guard refuses it). First sweep via a one-shot `scheduled_tasks` row (disabled after
  firing): 0 deletions, 1 `capability_gap` — 9 fact-overlap pairs + 1 near-duplicate on
  fundamentallyai, `do_not_auto_rewrite`, naming candidate 1. Commit `30dde02d1`.
- **The eye** — `scripts/contact_sheet.py` (commit `1f375991f`, register TL-039).
  Owner's copy: the private artifact "Acceptance renders — contact sheet". What the
  first look found is in NOTES 2026-08-03 and on TL-035's register entry.
- **Viewport on Renders** — implemented end to end, commit `d0a873f97`, council
  `a18db904` (`Council-Submitted:`; read the verdict when it lands). **INERT until
  adapter + chassis images roll** — either order is safe (`omitempty`). New note-line
  form: `(mobile 390x844@3x)`; old refs keep `(mobile)` exactly.
- **Content** — `sql/086b_tools_index_and_dead_cta_fix.sql` (read its CORRECTED blocks;
  they are half the value). `/tools.html` live (hero-tool + tool-cta, resolver-fed
  items); "Explore All Tools" → `/tools.html` on both tool pages; calculator hero
  "Run the calculator" → `#input-tokens`, "Review the methodology" → its guide; both
  guides built via `needs_page` items through `page-build-handler` and serving 200;
  decision-record stub archived. Simulator re-render probed: **47 checks, 0 failed**
  (the probe grew — trust its exit code, not a remembered count).

## 3. Traps this session paid for (all in LANDMINES/WRONG_CALLS too)

- **A queue `page_rerender` item with no `reason` is ASSEMBLE-ONLY.** It completes, the
  page deploys, and your content_data edit is not on it. For content/template changes
  use `scripts/rerender_page_sections_direct.sh` (proven three times today). RUNBOOK
  corrected — its INSERT recipe had also drifted from the live schema twice in one
  morning (`category`, then `pipeline`; the table has neither).
- **A static-source `input_schema` field overwrites your authored content_data on every
  resolve** (tool-cta's `secondary_cta_label` → "Learn how it works"). Read the schema
  before authoring; write only the `llm`/unsourced keys. Corollary: tool-cta `items`
  are `query.pages_where_type:tool`-fed — archiving the decision-record stub is what
  kept a 404 card off the new index.
- **Verify a link TARGET at the artefact before shipping the link.** The calculator's
  copy had promised a companion guide since 07-25; the guide's page row sat `planned`,
  0 components, serving 404 the whole time. Buttons at a 404 were live for ~2 minutes
  because the rerender beat the revert. WRONG_CALLS has the full shape.
- **Hours can pass between your turns.** A ledger note went in stamped "~00:20 BST"
  when the clock said 11:20. `date` before writing any timestamp claim.
- **`look.py`'s blank lower half on a short page is the vh-stretch artifact** (its own
  trap 2), not a page defect — doc_height tracks the 4000px probe viewport.

## 4. Corrections to the 08-02 handoff

- Its "a pre-delete guard is being BUILT now" — the guard was **built, approved
  (`6c5d1491`), and live** the same evening; my pod-grep re-proved it on 1238.
- Its §5.1 "NOBODY LOOKS" — closed in its cheapest form (TL-039). Cadence beyond
  on-demand is the remaining owner call.
- Its §5.3 viewport question — answered, implemented, inert until roll (§2).
- Its loose-ends list — all done except `tool-guide-intro` on the simulator page,
  which is **deliberately not added**: a NULL-content section escalates the whole page
  to the content writer, and the mutation-proven page is not worth risking for a strip
  whose need the new guide serves. Safe route if wanted: author the JSON from the
  guide's own copy and deliver via `section_edit` (no LLM near the page).

## 5. Open, in the order I would take them

1. **Owner call: contact-sheet cadence** — on-demand (today's state), a digest line, or
   a cron. Editorial.
2. **Owner call: camera ordering** — should renders capture the LANDING state (reload
   or capture before `evaluateOnPage` drives the page)? Changes what a render means;
   register TL-035 carries it as verify-later (d). Both known consumers carry the
   driven-state warning meanwhile, so nothing is urgent.
3. **Read the `a18db904` council verdict** (viewport change) and act on a REVISE — the
   code is already on the shared branch.
4. **151 candidate 1** — assign facts to sections at plan time. Now has a measured
   population on our own site (the capability_gap's 9 pairs) and remains this lane's
   largest unbuilt piece. The checker's residue items are the queue it should clear.
5. **Watch the checker's fleet behaviour** as other sites' discovery sweeps run — the
   census goes stale by design; a non-zero dedup item is worth reading, not alarming.

## 6. Commits this session

`30dde02d1` (seed 296) · `1f375991f` (contact_sheet.py) · `d0a873f97` (viewport,
Council-Submitted a18db904) · docs commit following this file. Artifact:
"Acceptance renders — contact sheet" in the owner's gallery (private).

---

## ADDENDUM 2026-08-04 — §5 items 1–3 are resolved; this file stays the cold start

- **§5.1 cadence: DONE.** Owner approved; `crontab -l` → Mondays 08:53,
  `scripts/weekly_contact_sheet_refresh.sh` (auth pre-check → regenerate →
  push-notify; log `~/acceptance_renders/refresh.log`). RUNBOOK §"The weekly
  contact sheet" carries the gotchas: `/snap/bin` on the cron PATH (kubectl is a
  snap — its absence reads as "token expired"), headless `claude -p` has NO
  Artifact tool (measured, so the gallery page refreshes on request only), push
  messages truncate ~200 chars. The artifact is a NEW url
  (`14a45889-e1f0-46e9-969a-08295cc36650`) — the 08-03 one was deleted from the
  gallery within a day; treat the URL as replaceable state.
- **§5.2 camera: DONE — landing state, owner-delegated.** `fe51ad611`, council
  `8e35caad` (Council-Submitted; READ THE VERDICT — open obligation). Renders
  now capture before `evaluateOnPage` and carry `Stage:"landing"`; failure
  evidence unchanged; note line gains `, landing state`. **INERT until the
  adapter image rolls** — after the roll, verify with a real acceptance run:
  the note line must carry the stage token, and the simulator's desktop render
  must show the DEFAULT preset, not the post-Clear empty panel.
- **§5.3 a18db904 verdict: READ — APPROVED r1**, one low-severity advisory
  (profileTag refactor minimality). No action.
- Still open: §5.4 (candidate 1) and §5.5 (watch fleet sweeps) — plus the
  `8e35caad` verdict, and the post-roll verification above.
