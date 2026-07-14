
<!-- SOURCE: U13_docs024_small_dirs.md -->
### Abandoned "no owner" claim (checked and found false)
- **category:** NEW:investigation-discipline
- **status-signal:** abandoned
- **status-evidence:** "An earlier claim in this plan — 'no agent owns ensuring a tool page has a working widget...' — was checked and found false." (PLAN_tool_widget_clobber(9).md §2.6)
- **what:** During the M2 diagnosis, the investigation asserted that no agent owned ensuring an adopted tool page gets a working widget. Verification against `apply_adoption_plan_action.go`, `check_tool_completeness_action.go`, and the agent-definitions backup showed `tool-recreation-handler` is a real, registered, active agent that already owns exactly this responsibility. The claim was retracted before any redundant handoff was built.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md#2.6, tools/tool_widget_clobber/NOTES_running_tool_widget_investigation.md#Phase-3,#Phase-4
- **relations:** Adoption interactivity misroute; Verify-before-acting investigation discipline
- **verify-later:** confirm `tool-recreation-handler` agent_definition remains registered/active

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Verify-before-acting investigation discipline (diagnosis methodology)
- **category:** NEW:investigation-discipline
- **status-signal:** deployed
- **status-evidence:** NOTES_running_tool_widget_investigation.md, whole document, esp. "the diagnosis changed three times, and each change came from refusing to act on the current theory until it was checked"
- **what:** A recorded set of working principles used through the tool-widget investigation: don't jump to conclusions, verify architectural claims by code search before turning them into tasks, prefer structural fixes over quick hacks, reuse existing helpers rather than building parallel ones, check the schema before writing SQL, make falsifiable predictions rather than declarative claims. Framed as reusable guidance and as raw material for a fix-loop council member.
- **sources:** tools/tool_widget_clobber/NOTES_running_tool_widget_investigation.md#How-the-diagnosis-evolved,#Principles-that-actually-drove-the-work
- **relations:** Abandoned "no owner" claim; fix-loop / diagnosis-loop council concept
- **verify-later:** n/a — process/design record, no separate code artifact beyond what is already cited in sources

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Abandoned "no owner" claim (checked and found false)
- **category:** NEW:investigation-discipline
- **status-signal:** abandoned
- **status-evidence:** "An earlier claim in this plan — 'no agent owns ensuring a tool page has a working widget...' — was checked and found false." (PLAN_tool_widget_clobber(9).md §2.6)
- **what:** During the M2 diagnosis, the investigation asserted that no agent owned ensuring an adopted tool page gets a working widget. Verification against `apply_adoption_plan_action.go`, `check_tool_completeness_action.go`, and the agent-definitions backup showed `tool-recreation-handler` is a real, registered, active agent that already owns exactly this responsibility. The claim was retracted before any redundant handoff was built.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md#2.6, tools/tool_widget_clobber/NOTES_running_tool_widget_investigation.md#Phase-3,#Phase-4
- **relations:** Adoption interactivity misroute; Verify-before-acting investigation discipline
- **verify-later:** confirm `tool-recreation-handler` agent_definition remains registered/active

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Verify-before-acting investigation discipline (diagnosis methodology)
- **category:** NEW:investigation-discipline
- **status-signal:** deployed
- **status-evidence:** NOTES_running_tool_widget_investigation.md, whole document, esp. "the diagnosis changed three times, and each change came from refusing to act on the current theory until it was checked"
- **what:** A recorded set of working principles used through the tool-widget investigation: don't jump to conclusions, verify architectural claims by code search before turning them into tasks, prefer structural fixes over quick hacks, reuse existing helpers rather than building parallel ones, check the schema before writing SQL, make falsifiable predictions rather than declarative claims. Framed as reusable guidance and as raw material for a fix-loop council member.
- **sources:** tools/tool_widget_clobber/NOTES_running_tool_widget_investigation.md#How-the-diagnosis-evolved,#Principles-that-actually-drove-the-work
- **relations:** Abandoned "no owner" claim; fix-loop / diagnosis-loop council concept
- **verify-later:** n/a — process/design record, no separate code artifact beyond what is already cited in sources
