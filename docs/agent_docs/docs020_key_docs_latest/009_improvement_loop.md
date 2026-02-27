improvement-loop (orchestrator) → spawns and calls design-discovery-agent → runs run_discovery_checks with checks including "hardcoded_section_colors" → countHardcodedColorComponents() counts affected page_components → writes site_work_items with handler_agent: "color-variable-fixer", status: "detected" → improvement-loop triages (triage_detected_items promotes to "triaged") → dispatches via build-dispatch-loop → spawns color-variable-fixer → runs fix_hardcoded_colors → replaces hardcoded hex with CSS variables → improvement-loop inserts needs_rerender item → dispatch re-renders pages.

improvement-loop
→ spawns design-discovery-agent
→ run_discovery_checks with "forced_text_colors" check
→ findForcedTextColors() parses CSS rules in <style> blocks
→ only flags child text element rules (h2, p, li...) with color: #hex
→ skips container rules (.hero-section { color: #fff }) and links
→ inserts site_work_item: item_type="forced_text_colors", handler="color-variable-fixer"
→ triage promotes detected → triaged
→ build-dispatch-loop spawns color-variable-fixer
→ Step 1: fix_hardcoded_colors (existing — fixes background hex → var(--color-primary))
→ Step 2: fix_forced_text_colors (new — removes child text color declarations)
→ Loads site's color palette from style_collections
→ For each component: determines bg color, calculates what text color WILL BE after removal
→ WCAG AA contrast check (4.5:1 minimum) before removing
→ For dark sections missing --section-* contract: injects the contract variables first
→ Skips and logs warning if resulting contrast would be too low
→ Fixes both templates (permanent) and rendered HTML (immediate)
→ improvement-loop inserts needs_rerender → pages redeployed