# Start Here: Next 2 Weeks Action Plan

## Goal
Get multipage-website-builder working reliably (Phase 0)

## Current Problem
- Multipage-website-builder tries to generate 4 pages at once
- Race conditions between spawned agents
- Timeouts and incomplete sites
- Landing-page-builder works, multipage doesn't

## Solution
Make multipage-builder work like landing-page-builder:
- Sequential, not parallel
- One page at a time
- Clear wait points between spawns

---

## Week 1: Fix the Workflow

### Day 1-2: Update multipage-builder workflow

**File to modify:** Agent definition for `multipage-website-builder`

**Current (broken) structure:**
```json
{
    "generate_batch_1": {
        "action": "spawn_multiple_writers",  // ❌ Spawns 4 at once
        "pages": [1, 2, 3, 4]
    }
}
```

**New (working) structure:**
```json
{
    "generate_pages_loop": {
        "action": "loop",
        "config": {
            "iterate_over": "page_plan.pages",
            "loop_var": "current_page",
            "substeps": {
                "write_page": {
                    "action": "call_agent",
                    "config": {
                        "agent_type": "content-creator",
                        "input_fields": ["current_page"]
                    }
                }
            }
        }
    }
}
```

**SQL to update:**
```sql
UPDATE agent_definitions
SET default_config = '{
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
                        "write_page": {
                            "action": "call_agent",
                            "config": {
                                "agent_type": "content-creator",
                                "target_role": "writer",
                                "input_fields": ["current_page"],
                                "timeout_seconds": 180
                            },
                            "output_field": "page_html"
                        }
                    }
                },
                "next_step": "wrap_pages",
                "output_field": "all_pages"
            },
            
            "wrap_pages": {
                "action": "wrap_multipage",
                "config": {
                    "pages_field": "all_pages"
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
}'::jsonb
WHERE type = 'multipage-website-builder';
```

**Deliverable:** Updated workflow in database

---

### Day 3-4: Implement wrap_multipage action

**New Go action needed:**

```go
// File: actions/multipage_actions.go

func WrapMultipageAction(ctx ActionContext) error {
    // 1. Get all page HTML from loop output
    pages := ctx.InputData["all_pages"].([]interface{})
    
    // 2. Generate navigation
    nav := generateNavigation(pages)
    
    // 3. Create index.html if not exists
    if !hasIndexPage(pages) {
        indexHTML := createIndexPage(pages[0], nav)
        pages = append([]interface{}{indexHTML}, pages...)
    }
    
    // 4. Inject navigation into each page
    for i, page := range pages {
        pages[i] = injectNavigation(page, nav)
    }
    
    // 5. Collect CSS/JS files
    assets := collectAssets(pages)
    
    // 6. Return wrapped site
    ctx.Output = map[string]interface{}{
        "files": pages,
        "assets": assets,
        "navigation": nav,
    }
    
    return nil
}

func generateNavigation(pages []interface{}) string {
    nav := "<nav>\n"
    for _, page := range pages {
        pageData := page.(map[string]interface{})
        title := pageData["title"].(string)
        path := pageData["path"].(string)
        nav += fmt.Sprintf(`  <a href="%s">%s</a>`, path, title)
    }
    nav += "</nav>\n"
    return nav
}

func injectNavigation(page interface{}, nav string) interface{} {
    pageData := page.(map[string]interface{})
    html := pageData["html"].(string)
    
    // Inject nav after <body> tag
    html = strings.Replace(html, "<body>", "<body>\n"+nav, 1)
    
    pageData["html"] = html
    return pageData
}
```

**Register action:**
```go
// In action registry
ActionRegistry["wrap_multipage"] = WrapMultipageAction
```

**Deliverable:** Working `wrap_multipage` action

---

### Day 5: Test with 3-page site

**Test workflow:**
```json
{
    "domain": "test-site.com",
    "objective": "Generate 3-page consulting site",
    "pages": ["home", "services", "contact"]
}
```

**Verify:**
- [ ] All 3 pages generated
- [ ] Navigation appears on each page
- [ ] No timeouts
- [ ] All pages deployed to git

**Debug if issues:**
- Check Kafka logs for message flow
- Verify loop completes all iterations
- Check spawned agent status

**Deliverable:** Working 3-page test site

---

## Week 2: Add Simple Voice Variation

### Day 6-7: Create simple flow table

**SQL:**
```sql
CREATE TABLE IF NOT EXISTS site_flows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    orchestration_id UUID NOT NULL,
    domain TEXT NOT NULL,
    
    flow_config JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- Simple function to get voice params for a page
CREATE OR REPLACE FUNCTION get_voice_for_page(
    p_orchestration_id UUID,
    p_page_name TEXT
)
RETURNS JSONB AS $$
DECLARE
    v_config JSONB;
BEGIN
    SELECT flow_config INTO v_config
    FROM site_flows
    WHERE orchestration_id = p_orchestration_id;
    
    -- Simple mapping: home = casual, others = professional
    IF p_page_name = 'home' OR p_page_name = 'index' THEN
        RETURN '{"formality": 0.5, "technical_depth": 0.3}'::jsonb;
    ELSE
        RETURN '{"formality": 0.7, "technical_depth": 0.5}'::jsonb;
    END IF;
END;
$$ LANGUAGE plpgsql;
```

**Deliverable:** Flow table and helper function

---

### Day 8-9: Update content-creator to use voice params

**Update content-creator workflow:**

```sql
UPDATE agent_definitions
SET default_config = jsonb_set(
    default_config,
    '{workflow,steps,generate,config,prompt_template}',
    '"Write content for: {{.page_name}}

Voice parameters:
- Formality: {{.voice_params.formality}} (0.5 = casual, 0.7 = professional)
- Technical depth: {{.voice_params.technical_depth}}

Content: {{.content_requirements}}

Adjust your writing style to match the formality level.
Lower = more casual and conversational
Higher = more professional and polished"'::jsonb
)
WHERE type = 'content-creator';
```

**Deliverable:** Updated content-creator

---

### Day 10: Update multipage-builder to query voice params

**Add SQL query before content generation:**

```sql
UPDATE agent_definitions
SET default_config = jsonb_set(
    default_config,
    '{workflow,steps,generate_pages_loop,config,substeps}',
    '{
        "get_voice": {
            "action": "execute_sql",
            "config": {
                "query": "SELECT get_voice_for_page($1, $2)",
                "params": ["orchestration_id", "current_page.name"],
                "output_field": "voice_params"
            }
        },
        "write_page": {
            "action": "call_agent",
            "config": {
                "agent_type": "content-creator",
                "input_fields": ["current_page", "voice_params"],
                "timeout_seconds": 180
            },
            "output_field": "page_html"
        }
    }'::jsonb
)
WHERE type = 'multipage-website-builder';
```

**Deliverable:** Voice params flowing to content creator

---

### Day 11-12: Test and measure

**Generate 2 test sites:**
```json
// Site 1: Without voice params (control)
{
    "domain": "control-site.com",
    "use_voice_params": false
}

// Site 2: With voice params (test)
{
    "domain": "test-site.com",
    "use_voice_params": true
}
```

**Measure:**
- Read home page vs contact page from both sites
- Assess if test site has noticeable voice variation
- Measure formality with text analysis tool
- Get human feedback

**Deliverable:** Evidence that voice variation works

---

## Success Criteria (End of Week 2)

### Must Have ✅
- [ ] Multipage-website-builder completes without timeout
- [ ] 3+ pages generated successfully
- [ ] All pages deployed to git
- [ ] Navigation works between pages

### Should Have ✅
- [ ] Voice parameters applied to pages
- [ ] Measurable difference between home and contact voice
- [ ] Process repeatable (can run multiple times)

### Nice to Have
- [ ] 5+ pages works
- [ ] Error handling graceful
- [ ] Logs clear and useful

---

## If You Get Stuck

### Multipage still timing out?
**Try:** Increase spawn delays in spawn_actions.go
```go
time.Sleep(20 * time.Second) // Instead of 10s
```

### Loop not iterating?
**Check:** Loop action implementation
**Verify:** `iterate_over` field exists in input data
**Debug:** Log each iteration

### Pages not deploying?
**Check:** Git adapter configuration
**Verify:** Repository exists
**Test:** deployer-agent independently

### Voice params not working?
**Simplify:** Just hardcode different prompts for home vs others
**Test:** Generate same page with different formalities manually

---

## Quick Wins

### If ahead of schedule:

**Quick Win 1:** Better navigation
```go
// Add active page highlighting
func generateNavigation(pages []interface{}, currentPage string) string {
    nav := "<nav>\n"
    for _, page := range pages {
        class := ""
        if page.path == currentPage {
            class = ` class="active"`
        }
        nav += fmt.Sprintf(`  <a href="%s"%s>%s</a>`, page.path, class, page.title)
    }
    nav += "</nav>\n"
    return nav
}
```

**Quick Win 2:** Basic CSS
```css
/* Include in wrapped pages */
nav {
    display: flex;
    gap: 1rem;
    padding: 1rem;
    background: #f5f5f5;
}

nav a {
    text-decoration: none;
    color: #333;
}

nav a.active {
    font-weight: bold;
    color: #0066cc;
}
```

**Quick Win 3:** Error recovery
```go
// In loop, if page generation fails, continue
for i, page := range pages {
    result, err := generatePage(page)
    if err != nil {
        log.Error("Failed to generate page", page.name, err)
        // Add placeholder page
        result = createErrorPage(page.name)
    }
    allPages = append(allPages, result)
}
```

---

## What NOT to Do

❌ **Don't** start on components yet
❌ **Don't** build interrogation system
❌ **Don't** implement full persona system
❌ **Don't** try to do multiple pages in parallel
❌ **Don't** add complexity before basics work

✅ **Do** focus on getting multipage working
✅ **Do** test after each change
✅ **Do** keep it simple
✅ **Do** measure if changes improve things

---

## After Week 2

**If successful:**
- You have working multipage generation
- Sites have basic voice variation
- Process is reliable

**Next decision:**
- Continue to Phase 2 (persona profiles)?
- Or stop here if good enough?

**Review:** `COMPREHENSIVE_ROADMAP.md` for next phases

---

## Files You Need

**To read first:**
1. This file (START_HERE.md)
2. COMPREHENSIVE_ROADMAP.md (full plan)
3. PRIORITY_MATRIX.md (decision guide)

**SQL to run:**
```bash
# Week 1: Just update multipage-builder workflow
psql $DATABASE_URL -c "UPDATE agent_definitions SET ... WHERE type='multipage-website-builder';"

# Week 2: Add flow table
psql $DATABASE_URL -f week2_flow_table.sql
```

**Go code to write:**
- actions/multipage_actions.go (wrap_multipage)
- Register action in action registry

---

## Summary

**Week 1:** Fix multipage-builder (sequential generation, wrap action)
**Week 2:** Add voice variation (flow table, voice params)
**Result:** Working multipage sites with voice variety

Start simple, test thoroughly, build from there.