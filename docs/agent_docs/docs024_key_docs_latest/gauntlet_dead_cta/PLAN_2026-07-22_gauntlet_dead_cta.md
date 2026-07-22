# PLAN — vonc gauntlet dead CTAs + the generic dead-control detector hole

**Started:** 2026-07-22 · **Branch:** `085_debug_and_feature_loops` · **Owner-directed.**
Site: vonc.com (`9ec3b9ee-5b08-461b-b4f8-9e1e03579c74`), page `tool-gauntlet`
(`ecb637c1-845f-46bf-b174-9c92a43f9586`), component `gauntlet-interface`
(`content_components` `5da50747-7936-4b8f-a66d-c1ea98919c75`; page_component
`1048b344-f1fa-44ea-b936-951bc7eafc59`).

## Symptom (owner report)
`vonc.com/tools/gauntlet/index.html#` — "the link doesn't work … not sure we have a
working gauntlet yet." Reproduced: the page serves 200 and its widget (timer,
checkable objectives, progress, animated stat counters) is fully wired and works.
The **only** broken controls are the two hero buttons:
- **Enter the Gauntlet** → `<a href="#" data-gi-enter-btn>` → clicking appends `#`
  (the URL the owner pasted). Dead.
- **Preview Rules** → `<a href="#">`. Dead.

## Root cause (evidenced)
1. Both `href="#"` are **hardcoded in the component's `html_template`**. The
   `input_schema` parameterises the button *labels* (`cta_enter_label`,
   `cta_preview_label`) but has **no URL field** — dead by construction; no
   content_data/resolver/render could give them a destination. The primary CTA
   even carries a `data-gi-enter-btn` hook the JS never binds.
2. The stats (`12,847` competitors / `94,210` completed / `38%` win rate / `7` day
   streak) and the 5-name leaderboard (AxonFury, ZeroRush, NexVoid, Skorch,
   Proxima) are **fabricated placeholders** — stats slotted from `static` fallbacks,
   leaderboard hardcoded in the template. There is no real gauntlet behind
   "Enter the Gauntlet".

## Why nothing caught it (the generic gap)
- `misdirected_cta` (live, enabled) scans the page but **skips `href="#"`**:
  `ClassifyLinkScope("#")` = `LinkScopeAnchor`, and the check only inspects
  page/empty scopes (`check_misdirected_cta.go:234`). Documented in `bugs_open/023`.
- `dead_controls` (live in binary, enabled on completeness-discovery-agent) is the
  detector **built for exactly this** — its header names *"the vonc gauntlet … both
  href="#", live for weeks"* as its proof case. But it **never fires on the
  gauntlet**, because its query filtered `p.build_status = 'deployed'`
  (`check_dead_controls.go:65`) and the gauntlet page is `build_status =
  'needs_rebuild'` while serving 200 (`pc.build_status='deployed'`). It is one of
  ~34 fleet pages that serve live as `needs_rebuild` (`bugs_open/049/052/053`).
  **The detector missed its own proof case.**

## Owner decisions (2026-07-22)
1. **Make the gauntlet genuinely work** (not a mock): wire the CTAs to real on-page
   behaviour; strip the fabricated stats + leaderboard so nothing is simulated.
2. **Fix the generic detector** (`dead_controls` build_status predicate) so any new
   site's live-but-needs_rebuild tool page gets its dead CTAs flagged — via the
   council gate, coordinating with the owning thread.
3. **Backend:** deliberately NOT building a full competitive-gaming backend (accounts
   + live leaderboard) now — no real competitors exist, so a live leaderboard would
   be a *new* fabrication. Reuse the existing form-delivery backend (bugfix 006) for
   the one real action ("file your Position" → delivered to the owner). Full backend
   is an explicit follow-on once real traffic exists.
4. **Council directive** (owner, verbatim intent): *we shouldn't be creating
   placeholders like that, that don't work.* Carried to the council as the rationale
   of the detector-fix submission (the fix enacts the directive).

## Phases
- **P1 — generic detector fix (Go, image-gated, council-gated).** `check_dead_controls.go`:
  gate liveness on `pc.build_status='deployed'` (the component that actually serves),
  not the drifting page-level `p.build_status`. DONE (edit + local build green);
  council submission carrying the owner directive next; commit on APPROVED with
  trailer; ship on next chassis image; verify the gauntlet is flagged post-roll.
- **P2 — gauntlet component honesty + function (config/content, live via section-editor).**
  Rewrite `gauntlet-interface` template + js_content + input_schema: CTAs do real
  on-page things; remove fabricated stats/leaderboard. Because the page is
  `rebuild_policy='owned'`, deliver ONLY via `section-editor`/`apply_section_edit`
  (generic rerender is forbidden — `bugs_closed/024`). Verify live by curl (match the
  component's OWN rule, never a generic property — the 024/046 trap).
- **P3 — real action (optional, owner-gated).** "File your Position" submits via the
  existing contact/lead form delivery. Modest, real, honest. Follow-on.

## Coordination
Dead-control detection is guard-rail-3 of the experience loop, actively owned by the
`bugs_open/054` (chrome dead-control) / `cta_link_integrity` (`bugs_open/023`) threads.
Do NOT fork: the P1 fix goes through the council gate; the finding is contributed into
their record. `who-owns.py 054` confirms active ownership (cqls).
