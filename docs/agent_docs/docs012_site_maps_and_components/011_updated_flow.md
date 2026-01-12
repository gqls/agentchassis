site-architect workflow:
┌─────────────────────────────────────┐
│ 1. load_components                  │
│    action: load_component_library   │
│    → queries content_components DB  │
│    → returns organized list         │
│    → includes "for_prompt" string   │
└─────────────────────────────────────┘
↓
┌─────────────────────────────────────┐
│ 2. design                           │
│    action: execute_llm_prompt       │
│    → receives component_library     │
│    → LLM sees available functions   │
│    → outputs pages with sections[]  │
└─────────────────────────────────────┘
↓
┌─────────────────────────────────────┐
│ 3. complete                         │
│    → returns site_architecture      │
└─────────────────────────────────────┘

LoadComponentLibraryAction Output

{
"by_function": {
"hero": {"id": "...", "function": "hero", "description": "..."},
"features": {"id": "...", "function": "features", "description": "..."}
},
"by_category": {
"hero": [...],
"features": [...]
},
"sections": [...],
"available_functions": ["hero", "features", "call_to_action", ...],
"for_prompt": "Available section components:\n\n## Hero\n- **hero**: Hero Section...",
"total_count": 29
}