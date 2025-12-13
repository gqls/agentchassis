# Consolidated Action Architecture

## Current Problem: Too Many Similar Actions

You have:
1. `assemble_from_library` - assembles from component library
2. `assemble_full_page` - assembles full pages
3. `html_actions` - various HTML operations
4. `assembly_multipage_site_actions.AssembleMultipageSiteAction` - takes multiple pages, adds nav
5. `html_assembly_actions.AssembleHTMLPartsAction` - combines structure + styles + content
6. `wrap_multipage_action.WrapMultipageAction` - wraps single page into multipage

**Too much overlap. Hard to know which to use when.**

---

## Proposed: 3 Clear Actions

### 1. `assemble_page` (replaces: assemble_full_page, AssembleHTMLPartsAction)

**Purpose:** Build ONE complete HTML page from parts

**Config:**
```json
{
    "action": "assemble_page",
    "config": {
        "structure_field": "architect.structure",
        "styles_field": "architect.styles", 
        "content_field": "writer.content"
    },
    "output_field": "page_html"
}
```

**Does:**
- Takes structure (skeleton HTML)
- Takes styles (CSS)
- Takes content (body content)
- Combines into valid HTML document
- Returns single complete page

**Implementation:** Merge logic from `AssembleHTMLPartsAction` + any useful bits from others

---

### 2. `assemble_multipage_site` (replaces: AssembleMultipageSiteAction, WrapMultipageAction)

**Purpose:** Take multiple pages, add navigation, create site

**Config:**
```json
{
    "action": "assemble_multipage_site",
    "config": {
        "pages_field": "all_pages",
        "add_navigation": true,
        "generate_missing_pages": ["about", "contact"]
    },
    "output_field": "site_files"
}
```

**Input:**
- Loop output: `{"index": "...", "services": "...", ...}`
- OR single page that needs wrapping

**Does:**
- Takes pages map/array
- Generates any missing standard pages (about, contact)
- Adds consistent navigation to all pages
- Returns files map ready for deployment

**Implementation:** Merge best of `AssembleMultipageSiteAction` + `WrapMultipageAction`

---

### 3. `assemble_from_components` (new, but use existing component logic)

**Purpose:** Build page from component library (Phase 3 feature)

**Config:**
```json
{
    "action": "assemble_from_components",
    "config": {
        "page_type": "landing",
        "requirements_field": "architect.requirements"
    },
    "output_field": "page_html"
}
```

**Does:**
- Queries component library
- Selects appropriate components
- Assembles into page
- Returns complete HTML

**Implementation:** Keep existing `assemble_from_library` logic, just ensure consistent interface

---

## Remove SQL From Workflows

### Problem: This in agent config

```json
{
    "get_voice": {
        "action": "execute_sql",
        "config": {
            "query": "SELECT get_voice_for_page($1, $2)",
            "params": ["orchestration_id", "current_page.name"]
        }
    }
}
```

**Bad because:**
- SQL in config is rigid
- Hard to test
- Breaks encapsulation
- Can't evolve without changing configs

### Solution: New action

```json
{
    "get_voice_params": {
        "action": "get_page_voice",
        "config": {
            "orchestration_id_field": "orchestration_id",
            "page_field": "current_page.name"
        },
        "output_field": "voice_params"
    }
}
```

**Action implementation:**

```go
func GetPageVoiceAction(ctx context.Context, params ActionParams) (interface{}, error) {
    orchestrationID := extractField(params.CollectedData, params.Config["orchestration_id_field"])
    pageName := extractField(params.CollectedData, params.Config["page_field"])
    
    // Query database
    voiceParams := queryVoiceParams(orchestrationID, pageName)
    
    return map[string]interface{}{
        "formality": voiceParams.Formality,
        "technical_depth": voiceParams.TechnicalDepth,
    }, nil
}
```

**Benefits:**
- SQL hidden in action
- Can add caching
- Can add fallbacks
- Testable
- Evolvable

---

## Simplified Multipage Builder Workflow

### Current (broken):
- Tries to batch generate
- Race conditions
- Complex data flow

### Proposed (works):

```json
{
    "workflow": {
        "start_step": "call_strategist",
        "steps": {
            "call_strategist": {
                "action": "call_agent",
                "config": {
                    "agent_type": "chief-strategist",
                    "timeout_seconds": 120
                },
                "next_step": "generate_pages_loop",
                "output_field": "page_plan"
            },
            
            "generate_pages_loop": {
                "action": "loop",
                "config": {
                    "iterate_over": "page_plan.pages",
                    "loop_var": "current_page",
                    "max_iterations": 10,
                    "substeps": {
                        "generate_page": {
                            "action": "call_agent",
                            "config": {
                                "agent_type": "content-creator",
                                "input_fields": ["current_page"],
                                "timeout_seconds": 180
                            },
                            "output_field": "page_html"
                        }
                    }
                },
                "next_step": "assemble_site",
                "output_field": "all_pages"
            },
            
            "assemble_site": {
                "action": "assemble_multipage_site",
                "config": {
                    "pages_field": "all_pages",
                    "add_navigation": true
                },
                "next_step": "deploy",
                "output_field": "site_files"
            },
            
            "deploy": {
                "action": "call_agent",
                "config": {
                    "agent_type": "deployer-agent",
                    "input_fields": ["site_files"]
                },
                "next_step": "complete"
            },
            
            "complete": {
                "action": "complete_workflow"
            }
        }
    }
}
```

**Clean. Simple. Works like landing-page-builder.**

---

## Implementation Plan

### Week 1: Consolidate Actions

**Day 1-2: Create unified `assemble_page` action**
- Merge `AssembleHTMLPartsAction` logic
- Add structure + styles + content combination
- Test independently

**Day 3-4: Create unified `assemble_multipage_site` action**
- Merge `AssembleMultipageSiteAction` + `WrapMultipageAction`
- Keep navigation generation
- Keep page generation for missing pages
- Test with 3 pages

**Day 5: Update multipage-builder workflow**
- Use new consolidated actions
- Remove SQL from config
- Test end-to-end

### Week 2: Add Voice Params (if needed)

**Only if Week 1 works and you want voice variation**

**Day 1-2: Create `get_page_voice` action**
```go
func GetPageVoiceAction(ctx context.Context, params ActionParams) (interface{}, error) {
    pageName := extractStringField(params.CollectedData, 
        params.Config["page_field"].(string), params.Logger)
    
    // Simple logic: home = casual, others = professional
    if pageName == "index" || pageName == "home" {
        return map[string]interface{}{
            "formality": 0.5,
            "technical_depth": 0.3,
        }, nil
    }
    
    return map[string]interface{}{
        "formality": 0.7,
        "technical_depth": 0.5,
    }, nil
}
```

**Day 3-4: Update content-creator to use voice params**
- Add voice params to prompt
- Test variation

**Day 5: End-to-end test**

---

## File Consolidation

### Keep These (Updated):

**1. `multipage_actions.go`** - All multipage operations
```go
// Contains:
- AssembleMultipageSite() // merged from AssembleMultipageSiteAction + WrapMultipageAction
- GenerateStandardPage() // for about, contact
- AddNavigationToPages()
- helper functions
```

**2. `html_actions.go`** - Basic HTML operations
```go
// Contains:
- AssemblePage() // merged from AssembleHTMLPartsAction + assemble_full_page
- InjectStyles()
- InjectContent()
- ValidateHTML()
- helper functions
```

**3. `component_actions.go`** - Component library (Phase 3)
```go
// Contains:
- AssembleFromComponents() // keep existing assemble_from_library
- QueryComponents()
- RenderComponentTree()
```

**4. `voice_actions.go`** - Voice/persona helpers (Phase 2)
```go
// Contains:
- GetPageVoice()
- GetPersonaProfile() // for Phase 2
- helper functions
```

### Remove/Archive:
- `assembly_multipage_site_actions.go` → merge into `multipage_actions.go`
- `html_assembly_actions.go` → merge into `html_actions.go`
- `wrap_multipage_action.go` → merge into `multipage_actions.go`

---

## Updated Immediate Roadmap

### Phase 0: Fix Multipage (This Week)

**Goal:** Multipage builder works reliably

**Tasks:**
1. Consolidate actions (3 clear actions)
2. Update multipage-builder workflow to use consolidated actions
3. Remove SQL from configs
4. Test with 3-page site
5. Verify deployment works

**Success:** Can generate 3-5 page sites reliably

### Phase 1: Add Simple Voice (Next Week - Optional)

**Goal:** Pages sound different

**Tasks:**
1. Create `get_page_voice` action
2. Update content-creator to use voice params
3. Test voice variation
4. Measure difference

**Success:** Home page measurably more casual than contact page

### Stop Here and Assess

Before continuing to Phase 2+ (personas, components, etc.), evaluate:
- Is quality good enough?
- Do we need more sophistication?
- What's ROI of continuing?

---

## Key Principles

1. **One action, one purpose** - no overlap
2. **No SQL in configs** - hide in actions
3. **Test each action independently** - before using in workflow
4. **Keep it simple** - landing-page-builder pattern works, copy it
5. **Consolidate before adding** - don't add to the mess

---

## Questions to Answer

1. **Which actions do you actually use?**
    - Which of the 6 are in active workflows?
    - Which can we delete?

2. **What's the minimum to get multipage working?**
    - Can we just fix timing and call it done?
    - Or do we need consolidation first?

3. **Do you want voice params now?**
    - Or just get multipage working first?
    - What's the priority?

Let me know and I'll create the consolidated action files.