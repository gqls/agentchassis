{
"workflow": {
"steps": {
"generate_all_content": {
"action": "execute_llm_prompt",
"config": {
"prompt_template": "Generate content for ALL pages: {{.page_list}}..."
},
"output_field": "all_pages_content"
},
"generate_shared_styles": {
"action": "execute_llm_prompt",
"config": {
"prompt_template": "Generate CSS for entire site..."
},
"output_field": "shared_styles"
},
"assemble_multipage_site": {
"action": "assemble_multipage_site",
"config": {
"index_html_field": "all_pages_content.index",
"pages_field": "all_pages_content.pages",
"shared_styles_field": "shared_styles.result",
"navigation_field": "site_architecture.navigation"
},
"output_field": "site_files"
}
}
}
}

# Using HTML Actions Properly

## The Problem

The current `html-developer` agent uses `execute_llm_prompt` directly, bypassing the well-designed HTML action architecture in `html_actions.go`:

```json
❌ Current (bypassing actions):
{
  "action": "execute_llm_prompt",
  "config": {
    "prompt_template": "Generate HTML... {{.site_architecture}} {{.site_content}}"
  }
}
```

This causes:
- No intelligent context gathering
- No HTML processing (meta tags, responsive, etc.)
- No validation
- Manual prompt construction
- Duplication of logic

## The Solution

Use the proper HTML action architecture:

```json
✅ Correct (using actions):
{
  "steps": {
    "generate_html": {"action": "generate_html"},
    "process_html": {"action": "process_html"},
    "validate_html": {"action": "validate_html"}
  }
}
```

## The HTML Actions

### 1. `generate_html` - Intelligent HTML Generation

**What it does:**
- Automatically gathers context from CollectedData:
    - `analyze_domain` → domain analysis
    - `architect_site` → site structure
    - `create_content` → content
    - `input_data` → business info
- Builds optimized prompt
- Calls LLM
- Extracts clean HTML (removes markdown blocks)

**Configuration:**
```json
{
  "action": "generate_html",
  "config": {
    "generation_type": "full",  // or "structure", "styles", "content"
    "max_tokens": 16000
  }
}
```

**Output:**
```json
{
  "raw_html": "<!DOCTYPE html>...",
  "generation_type": "full",
  "generated_at": "2025-12-06T...",
  "prompt_used": "...",
  "tokens_used": 16000
}
```

### 2. `process_html` - HTML Enhancement

**What it does:**
- Parses HTML with goquery
- Ensures proper structure (html, head, body)
- Adds meta tags:
    - charset (UTF-8)
    - viewport
    - description
    - Open Graph tags
- Ensures responsive design
- Optimizes images (lazy loading, srcset)
- Minifies inline CSS/JS

**No configuration needed** - it automatically finds HTML from:
- `generate_html` step
- `raw_html` field

**Output:**
```json
{
  "processed_html": "<!DOCTYPE html>...",
  "original_size": 15432,
  "processed_size": 14567,
  "processing_steps": ["structure_validation", "meta_tags", "responsive_design", "optimization"],
  "business_info": {...}
}
```

### 3. `validate_html` - HTML Validation

**What it does:**
- Validates structure
- Checks required elements (html, head, body, title)
- Checks meta tags
- Validates images (src, alt)
- Validates links (href)
- Checks accessibility

**No configuration needed** - finds HTML from `process_html` step

**Output:**
```json
{
  "valid": true,
  "errors": [],
  "warnings": ["Image 3 missing alt text"],
  "validation_time": 0.015,
  "checks_passed": 12,
  "checks_failed": 0
}
```

## Complete Workflows

### Standard Single-Page Workflow

```json
{
  "workflow": {
    "start_step": "generate_html",
    "steps": {
      "generate_html": {
        "action": "generate_html",
        "next_step": "process_html",
        "output_field": "html_generation"
      },
      "process_html": {
        "action": "process_html",
        "next_step": "validate_html",
        "output_field": "html_processing"
      },
      "validate_html": {
        "action": "validate_html",
        "next_step": "complete",
        "output_field": "html_validation"
      },
      "complete": {
        "action": "complete_workflow"
      }
    }
  }
}
```

### Chunked Generation Workflow

```json
{
  "workflow": {
    "start_step": "generate_structure",
    "steps": {
      "generate_structure": {
        "action": "generate_html",
        "config": {
          "generation_type": "structure",
          "max_tokens": 4000
        },
        "next_step": "generate_styles",
        "output_field": "structure_gen"
      },
      "generate_styles": {
        "action": "generate_html",
        "config": {
          "generation_type": "styles",
          "max_tokens": 8000
        },
        "next_step": "generate_content",
        "output_field": "styles_gen"
      },
      "generate_content": {
        "action": "generate_html",
        "config": {
          "generation_type": "content",
          "max_tokens": 12000
        },
        "next_step": "assemble_parts",
        "output_field": "content_gen"
      },
      "assemble_parts": {
        "action": "assemble_html_parts",
        "config": {
          "structure_field": "structure_gen.raw_html",
          "styles_field": "styles_gen.raw_html",
          "content_field": "content_gen.raw_html"
        },
        "next_step": "process_html",
        "output_field": "assembled"
      },
      "process_html": {
        "action": "process_html",
        "next_step": "validate_html",
        "output_field": "processed"
      },
      "validate_html": {
        "action": "validate_html",
        "next_step": "complete",
        "output_field": "validation"
      },
      "complete": {
        "action": "complete_workflow"
      }
    }
  }
}
```

## Context Gathering (How It Works)

The `generate_html` action looks for these keys in `CollectedData`:

| Key in CollectedData | What It Contains | Used For |
|---------------------|------------------|----------|
| `analyze_domain` | Domain analysis results | Understanding the business |
| `architect_site` | Site architecture | Structure and sections |
| `create_content` | Content from content creator | Actual page content |
| `input_data` | Original request | Business name, domain, description |

**Example CollectedData:**
```json
{
  "input_data": {
    "domain": "techcorp.com",
    "business_name": "TechCorp Solutions"
  },
  "analyze_domain": {
    "result": {
      "industry": "technology",
      "target_audience": "B2B enterprises"
    }
  },
  "architect_site": {
    "result": {
      "sections": ["hero", "features", "pricing", "contact"],
      "theme": "modern-tech"
    }
  },
  "create_content": {
    "result": {
      "hero": "Innovative Tech Solutions...",
      "features": "..."
    }
  }
}
```

The action automatically extracts and combines this into an optimized prompt.

## Implementation Steps

### Step 1: Update Existing html-developer

```bash
# Apply the corrected workflow
psql $DATABASE_URL -f html_developer_using_actions.sql

# Delete old pods to pick up new config
kubectl delete pods -l agent-type=html-developer
```

### Step 2: Add Enhanced HTML Actions (for chunked support)

```bash
# Option A: Replace existing html_actions.go
cp html_actions_enhanced.go platform/orchestration/actions/html_actions.go

# Option B: Add alongside existing (safer)
cp html_actions_enhanced.go platform/orchestration/actions/html_actions_v2.go
# Then update action registry to use enhanced versions
```

### Step 3: Add AssembleHTMLPartsAction

```bash
# Add the assembly action
cp html_assembly_actions.go platform/orchestration/actions/

# Register it
# In action_registry.go:
actionHandlers["assemble_html_parts"] = AssembleHTMLPartsAction
```

### Step 4: Rebuild

```bash
make build-agent-chassis
docker tag ... :v1.0.510
docker push ... :v1.0.510
```

### Step 5: Create Chunked Agent (Optional)

```bash
# Update image_tag in SQL to v1.0.510
psql $DATABASE_URL -f html_developer_chunked_using_actions.sql
```

## Benefits of Using HTML Actions

| Aspect | Raw LLM Call | Using HTML Actions |
|--------|--------------|-------------------|
| Context gathering | Manual | ✅ Automatic |
| Prompt construction | Manual, error-prone | ✅ Optimized |
| HTML extraction | Manual regex | ✅ Robust parsing |
| Meta tags | Not added | ✅ Added automatically |
| Responsive design | Hope LLM includes it | ✅ Ensured |
| Validation | None | ✅ Built-in |
| Image optimization | None | ✅ Lazy loading, srcset |
| Minification | None | ✅ CSS/JS minified |
| Reusability | Copy-paste prompts | ✅ Reuse actions |

## Debugging

### Check what context is gathered:

```bash
kubectl logs <html-developer-pod> | grep "Generating HTML"
# Should show: "Generating HTML content" followed by context details
```

### Check processing steps:

```bash
kubectl logs <html-developer-pod> | grep "Processing HTML"
# Should show: processing_steps: ["structure_validation", "meta_tags", ...]
```

### Check validation results:

```bash
kubectl logs <html-developer-pod> | grep "Validating HTML"
# Should show: errors: [], warnings: [...]
```

## Migration Checklist

- [ ] Understand HTML action architecture (generate → process → validate)
- [ ] Update html-developer to use actions instead of execute_llm_prompt
- [ ] Test with simple single-page generation
- [ ] Add html_actions_enhanced.go for chunked support
- [ ] Add assemble_html_parts action
- [ ] Create chunked variant agent
- [ ] Test chunked generation
- [ ] Update website-builder to use new html-developer
- [ ] Monitor logs for errors
- [ ] Verify output quality

## Files Provided

**Core Actions:**
- [html_actions_enhanced.go](computer:///mnt/user-data/outputs/html_actions_enhanced.go) - Enhanced with chunking support
- [html_assembly_actions.go](computer:///mnt/user-data/outputs/html_assembly_actions.go) - Assembly action

**Agent Definitions:**
- [html_developer_using_actions.sql](computer:///mnt/user-data/outputs/html_developer_using_actions.sql) - Standard single-page
- [html_developer_chunked_using_actions.sql](computer:///mnt/user-data/outputs/html_developer_chunked_using_actions.sql) - Chunked version

**Documentation:**
- This file - Complete usage guide

## Key Takeaway

**ALWAYS use the HTML actions instead of raw LLM calls.**

They provide:
✅ Intelligent context gathering
✅ Optimized prompts  
✅ HTML processing & enhancement
✅ Validation
✅ Better results with less code

The architecture is already there - use it!