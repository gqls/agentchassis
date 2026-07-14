# Register — investigation-discipline

2 concepts, consolidated from 4 raw extractions across unit U13 (each of the 2 distinct blocks
appeared byte-identically twice within the cluster input file — treated as duplicate copies of
one extraction, not independent corroboration).

### INVD-001 — Abandoned "no owner" claim (checked and found false)
- **status:** abandoned
- **status-evidence:** "An earlier claim in this plan — 'no agent owns ensuring a tool page has a working widget...' — was checked and found false." (PLAN_tool_widget_clobber(9).md §2.6).
- **what:** During the M2 diagnosis, the investigation asserted that no agent owned ensuring an adopted tool page gets a working widget. Verification against apply_adoption_plan_action.go, check_tool_completeness_action.go, and the agent-definitions backup showed tool-recreation-handler is a real, registered, active agent that already owns exactly this responsibility. The claim was retracted before any redundant handoff was built.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md#2.6, tools/tool_widget_clobber/NOTES_running_tool_widget_investigation.md#Phase-3,#Phase-4
- **relations:** Adoption interactivity misroute; Verify-before-acting investigation discipline
- **verify-later:** confirm tool-recreation-handler agent_definition remains registered/active

### INVD-002 — Verify-before-acting investigation discipline (diagnosis methodology)
- **status:** convention
- **status-evidence:** NOTES_running_tool_widget_investigation.md, whole document, esp. "the diagnosis changed three times, and each change came from refusing to act on the current theory until it was checked".
- **stage2-verified (2026-07-14):** deployed → convention — verify-later explicitly says 'n/a — process/design record, no separate code artifact'; what: describes working principles/methodology (verify before acting, prefer structural fixes), not a built artifact. 'deployed' is the wrong status label for a process doctrine; reclassified as convention/process, not a false-pos...
- **what:** A recorded set of working principles used through the tool-widget investigation: don't jump to conclusions, verify architectural claims by code search before turning them into tasks, prefer structural fixes over quick hacks, reuse existing helpers rather than building parallel ones, check the schema before writing SQL, make falsifiable predictions rather than declarative claims. Framed as reusable guidance and as raw material for a fix-loop council member.
- **sources:** tools/tool_widget_clobber/NOTES_running_tool_widget_investigation.md#How-the-diagnosis-evolved,#Principles-that-actually-drove-the-work
- **relations:** Abandoned "no owner" claim; fix-loop / diagnosis-loop council concept (fix-loop register)
- **verify-later:** n/a — process/design record, no separate code artifact beyond what is already cited in sources
