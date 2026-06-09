# guidelines (the constitution) — grab the highest version of each
ls docs/agent_docs/docs024_key_docs_latest/ | grep -E '^(001_development_guide|002_system_architecture|003_contracts_and_standards)'
find docs/agent_docs -iname '*debugging_guide*' | sort        # want the newest 016_debugging_guide_v2_NN

# the gamesdesign defect list (THE source for this task) — highest version
find docs/agent_docs -iname '*CATALOGUE*gamesdesign*' | sort

# component / hero / CTA / content-quality design docs (any that exist)
ls docs/agent_docs/docs024_key_docs_latest/ | grep -iE 'component|hero|content_direction|content.quality|019_tool_library|020_tool_lifecycle'

# this session's running notes + any newer handoff
find docs/agent_docs -iname 'running_notes_1*' -o -iname 'HANDOFF_2026-06*' | sort