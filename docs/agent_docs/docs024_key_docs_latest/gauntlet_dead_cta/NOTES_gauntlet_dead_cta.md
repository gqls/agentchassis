# NOTES — gauntlet dead CTAs (append-only, newest at bottom)

## 2026-07-22 — diagnosis
- Symptom: `vonc.com/tools/gauntlet/index.html#` — hero buttons dead.
- Live page: HTTP 200, 41KB. JS asset `/tools/assets/gauntlet-interface.js` intact
  (3909B, balanced, ends `})();`), selectors match HTML hooks — the widget works.
- Two `href="#"` anchors = the CTAs "Enter the Gauntlet" (`data-gi-enter-btn`) and
  "Preview Rules". Sibling tools arena + archetype-taster-quiz have ZERO `href="#"`
  — gauntlet is the anomaly, NOT a blanket every-tool pattern. [CORRECTED my first
  hypothesis — I initially assumed a fleet-wide "all tools ship dead CTAs" pattern;
  the sibling check refuted it.]
- DB: page `tool-gauntlet` (`ecb637c1…`), page_type=tool, rebuild_policy=owned,
  **build_status=needs_rebuild** (but pc.build_status=deployed → serves live).
  One component `gauntlet-interface` (`5da50747…`), component_level=section,
  has_input_schema=t. `href="#"` is in BOTH html_template and rendered_html.
- Template: `<a href="#" data-gi-enter-btn>{{.cta_enter_label}}</a>` and
  `<a href="#">{{.cta_preview_label}}</a>`. input_schema has label fields
  (cta_enter_label/cta_preview_label) but **NO url field** → dead by construction.
- Stats (12,847/94,210/38%/7) slotted from static fallbacks; leaderboard (AxonFury,
  ZeroRush, NexVoid, Skorch, Proxima) hardcoded in template → fabricated.

## 2026-07-22 — why nothing caught it (code-read, not grep-guess)
- `check_misdirected_cta.go:234`: only inspects LinkScopePage/LinkScopeEmpty →
  `href="#"` = LinkScopeAnchor → skipped. (bugs_open/023's documented hole.)
- `check_dead_controls.go`: EXISTS, live in binary (16 symbol hits, pod
  agent-chassis-7d4ff8b54-cm786 started 2026-07-22T13:56Z), enabled on
  completeness-discovery-agent. Header names the vonc gauntlet as its proof case.
  BUT query filtered `p.build_status='deployed'` (line 65) and gauntlet page is
  needs_rebuild → **skipped its own proof case**. Confirmed: only 2 dead_control
  items ever filed for vonc (brief-explanation), none for gauntlet. 15:47 sweep
  today produced misdirected/phantom/empty items but no gauntlet dead_control.
- IsNoopHref (links.go:109) recognises bare "#" as no-op; DeadControlAnchors uses
  it; DropDeadURLControls (chrome sibling, render-time) is for site_components.

## 2026-07-22 — P1 fix (generic detector)
- Edit: `check_dead_controls.go` — moved `= 'deployed'` predicate from
  `p.build_status` to `pc.build_status` (component liveness, not drifting page flag).
  Local `go build ./platform/orchestration/actions/discovery_checks/` GREEN.
- Council submission fired (owner directive carried as rationale): SUBMISSION_CORR
  `1834a349-c652-4889-b8bf-fcf5b553ad21`, orch `591acbf1…`, name council-gate-174600.
  Await verdict (~30min queue). Commit on APPROVED with trailer; ship next image.
- NOT yet committed (awaiting verdict; will commit narrowly by pathspec).

## 2026-07-22 — landmines to respect for P2 (gauntlet rewrite)
- Owned tool page → deliver ONLY via section-editor/apply_section_edit; generic
  rerender is FORBIDDEN and REFUSED (bugs_closed/024). Dispatch lane is cron-starved
  (bugs_open/030) — drive via kafka 085 envelope, don't wait on the queue.
- Verify live by the component's OWN rule, never a generic property (024/046 trap).
- collectJSAssets republishes js_content as /tools/assets/*.js.

## 2026-07-22 — P2 delivery DONE + verified live
- Rewrote gauntlet-interface: removed both `href="#"` CTAs, removed fabricated
  stats bar (12,847/94,210/38%/7) + 5-name leaderboard (AxonFury…), added an
  honest #gi-rules card. Primary CTA is now a real `<button data-gi-enter-btn>`
  that starts the timer + scrolls to the challenge + focuses objective 1;
  secondary is `<a href="#gi-rules">` (a NAMED anchor — a real link, not a dead
  control) that reveals+highlights the rules. Template 25054→22912, JS 3910→4090.
- Applied via UPDATE content_components (dollar-quoted, verified 1 row) then
  DELIVERED via section-editor content_edit (corr 7fe2143d, COMPLETED). field_updates
  carried honest copy (eyebrow TODAY'S GAUNTLET, softened subtitle, rules list).
- LANDMINE HIT: apply_section_edit reassembles+deploys the HTML but does NOT run
  collectJSAssets, so the JS asset stayed stale (old 3909B, didn't wire the new
  button). collectJSAssets runs ONLY in rerender_single_page (page-rerender's
  render_page/else branch). Fixed by an assemble-only page-rerender.
- LANDMINE HIT: the 049b bare `action=orchestrate` page-rerender envelope did NOT
  ingest (the kubectl-run stdin race its own comment warns about — no orch row, no
  work item, no log trace). The 086-style DIRECT orchestrator envelope
  (spawn_agent+call_agent, action=process, full inline workflow) routed reliably.
  Script saved: scripts/republish_gauntlet_js.sh. USE THIS PATTERN, not bare 049b.
- LANDMINE: section-editor left pc.build_status='approved' (drift). Set back to
  'deployed' before the assemble rerender (an assemble path could otherwise drop a
  non-deployed component). This drift is itself more evidence for the P1 fix.
- VERIFIED LIVE (cache-busted): page 0 dead href=#, 0 fabricated data, real
  <button> CTA, #gi-rules card; JS asset last-modified 2026-07-22T18:17:48Z, 4090B,
  wires enter-btn + rules-btn, no dead stat-counter. Hero copy honest (only
  remaining 'Live' is aria-live on the timer).

## 2026-07-22 — P1 council + commit
- Council APPROVED (corr 1834a349, round 1, 3 advisory objections none high-sev;
  10 reviewers / 6 abstained). Committed check_dead_controls.go 01e18019a with
  trailer Council-Reviewed: 1834a349. INERT until next chassis image roll (did NOT
  roll a fleet image for a detection-only change — ships with next build).
  Post-roll verify: re-run dead_controls on vonc → a dead_control item should now
  be filed for tool-gauntlet (it currently is NOT, because the page is
  needs_rebuild). NOTE the gauntlet's OWN dead CTAs are now gone, so the post-roll
  proof needs a DIFFERENT needs_rebuild page that still has a dead control, OR a
  temporary re-check before this fix would have flagged it.

## 2026-07-22 — MISSTEPS this session (recorded on owner request; the point is the pattern)

Two reached WRONG_CALLS.md (fleet-wide ledger); the rest are wasted cycles / near-misses
that a check would have saved. Recorded so the next thread doesn't re-walk them.

1. **Framed the dead CTAs as a fleet-wide "tools ship dead CTAs" pattern before checking
   siblings.** The request ("generic to any new site") primed a fleet-wide framing;
   `arena` + `archetype-taster-quiz` on the same site had ZERO `href="#"`. Caught by a
   2-line sibling curl before it reached a durable claim. → WRONG_CALLS. LESSON: "make the
   fix generic" is about the FIX's blast radius, not evidence the SYMPTOM is widespread.

2. **Four SQL queries against non-existent columns** (`page_name`, `component_type`,
   `schema_migrations.version/id`, `orchestration_states.name/agent_type`). Each threw;
   four retries. → WRONG_CALLS. LESSON: `\d <table>` first (the standing CLAUDE.md rule I
   skipped four times).

3. **Fired the bare `049b` page-rerender envelope and waited on it — it never ingested.**
   Six polls (60s) returned no orchestration row, no work item, no log trace. The script's
   OWN header warns of a kubectl-run stdin race with `action=orchestrate`. I trusted "works
   in seconds" (my memory note) over the caveat in front of me. Recovered by switching to
   the 086-style DIRECT orchestrator envelope (`action=process`, spawn+call, full inline
   workflow), which routed first try. LESSON: after firing a fire-and-forget dispatch,
   confirm INGEST (an orch row within ~30s) before waiting on the outcome; read the
   trigger's own caveats before reusing it. Durable route saved as
   `scripts/republish_gauntlet_js.sh`.

4. **Near-miss: nearly declared the page fixed on the HTML alone.** After the section-editor
   delivery the DB `rendered_html` + live HTML were correct and the orchestration was
   COMPLETED — I could have stopped there. The JS ASSET was still stale (old 3909B), so the
   primary `<button>` was inert. Caught only because I checked the asset's last-modified +
   content separately. LESSON: for a tool whose behaviour lives in a JS asset, "the HTML is
   right + status COMPLETED" is NOT proof the tool works — verify the served
   `/tools/assets/*.js` (last-modified + a symbol it must contain). This is the
   status-vs-artefact trap one layer out, and it's now a 016b §9 pattern.

5. **Minor: a shell grep broke on an apostrophe** (`grep … "TODAY'S GAUNTLET"` → unexpected
   EOF) mid-verification. Cost one re-run; switched the verification to a Python heredoc.
   LESSON: quote apostrophe-bearing literals in Python/`-F`, not inline double-quoted bash.
