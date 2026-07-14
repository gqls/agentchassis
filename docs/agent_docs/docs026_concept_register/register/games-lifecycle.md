# Register — games-lifecycle

1 concept, consolidated from 1 raw extraction across unit U13.

### GML-001 — Games quality lifecycle parity (new game_health / game-auditor / game-behavioral-tester / game-improver)
- **status:** aspirational
- **status-evidence:** "Games currently have no quality lifecycle. Add the analogues, reusing tool shapes wherever possible" (PLAN_tools_games_behavioral_qa_loop.md §7) — a proposal with no implementation evidence found.
- **what:** Proposes mirroring the entire tool-lifecycle quality apparatus for games, which today have no equivalent: `check_game_health.go` as the Tier-1 analogue of tool_health; `game-auditor` as the Tier-2 analogue of tool-auditor; `game-behavioral-tester` sharing the QA-loop harness with game-specific invariants; `game-improver` as the fix handler for `improve_game` items. Explicitly conditional on first confirming games are modelled compatibly with tools (component_level/page_type, fork model) so the existing tool pipeline can be forked rather than rewritten from scratch.
- **sources:** tools/tool_widget_clobber/PLAN_tools_games_behavioral_qa_loop.md §7, §11
- **relations:** Behavioral QA loop for tools & games (tool-lifecycle TL-019, the shared harness this would reuse); tool-lifecycle verification ladder (TL-008, the apparatus being mirrored); tool pipeline (TP-001, the shape to be forked)
- **verify-later:** whether component_level='game'/page_type='game' schema support already exists; whether any of check_game_health.go / game-auditor / game-behavioral-tester / game-improver have been built
