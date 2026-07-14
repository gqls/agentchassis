
<!-- SOURCE: U13_docs024_small_dirs.md -->
### Games quality lifecycle parity (new game_health / game-auditor / game-behavioral-tester / game-improver)
- **category:** NEW:games-lifecycle
- **status-signal:** aspirational
- **status-evidence:** "Games currently have no quality lifecycle. Add the analogues, reusing tool shapes wherever possible" (PLAN_tools_games_behavioral_qa_loop.md §7)
- **what:** Proposes mirroring the entire tool-lifecycle quality apparatus for games: `check_game_health.go` as the Tier-1 analogue of tool_health; `game-auditor` as the Tier-2 analogue of tool-auditor; `game-behavioral-tester` sharing the QA-loop harness with game-specific invariants; `game-improver` as the fix handler for `improve_game` items. Explicitly conditional on first confirming games are modelled compatibly (component_level/page_type, fork model) so the tool pipeline can be forked rather than rewritten.
- **sources:** tools/tool_widget_clobber/PLAN_tools_games_behavioral_qa_loop.md §7,§11
- **relations:** Behavioral QA loop for tools & games; tool-lifecycle (020)
- **verify-later:** whether `component_level='game'`/`page_type='game'` schema support already exists

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Games quality lifecycle parity (new game_health / game-auditor / game-behavioral-tester / game-improver)
- **category:** NEW:games-lifecycle
- **status-signal:** aspirational
- **status-evidence:** "Games currently have no quality lifecycle. Add the analogues, reusing tool shapes wherever possible" (PLAN_tools_games_behavioral_qa_loop.md §7)
- **what:** Proposes mirroring the entire tool-lifecycle quality apparatus for games: `check_game_health.go` as the Tier-1 analogue of tool_health; `game-auditor` as the Tier-2 analogue of tool-auditor; `game-behavioral-tester` sharing the QA-loop harness with game-specific invariants; `game-improver` as the fix handler for `improve_game` items. Explicitly conditional on first confirming games are modelled compatibly (component_level/page_type, fork model) so the tool pipeline can be forked rather than rewritten.
- **sources:** tools/tool_widget_clobber/PLAN_tools_games_behavioral_qa_loop.md §7,§11
- **relations:** Behavioral QA loop for tools & games; tool-lifecycle (020)
- **verify-later:** whether `component_level='game'`/`page_type='game'` schema support already exists
