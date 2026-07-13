# Website Builder Architecture - Status Report

## 1. CSS Architecture

### Problem
Component inline CSS was overriding global CSS colors, breaking the design cascade.

### Solution
**Responsibility barrier:**
- Global CSS (styles.css): All colors, fonts, typography for h1-h6, p, a elements
- Component CSS (inline): Layout, positioning, structure only

**Exception:** Dark/inverted sections must override text colors to white.

### Files
- `26_css_responsibility_barrier.sql` - Updates components to layout-only CSS
- `01_webdesign_agent.sql` - Updated prompt with responsibility rules

---

## 2. Contact Information

### Problem
LLM hallucinated email addresses and phone numbers in content.

### Solution
`rerenderInjectContactInfo()` function in rerender action queries site's actual contact data from database and replaces contact-info sections with template-rendered content using real data.

### Files
- `10_rerender_pages_action.go` - Contains injection function

---

## 3. Content Validation

### Problem
Broken internal links and incorrect emails deployed to live sites.

### Solution
`validate_page_content` action runs before review mode determination:
- Extracts all `href` values from HTML
- Checks internal links against pages table
- Checks emails against site contact data
- Returns errors (broken links) and warnings (wrong emails)

Validation errors force HITL review, blocking auto-approval.

### Files
- `12_validate_page_content_action.go` - Validation action
- `11_content_reviewer_validation.sql` - Workflow integration

---

## 4. HITL Rejection Flow

### Problem
No way to reject content and flag for later attention.

### Solution
HITL reviewer can:
- Approve (proceed to deploy)
- Approve with edits (use edited HTML)
- Reject (mark page `needs_attention`, skip deploy)

Rejected pages are picked up by maintenance workflow.

### Files
- `11_content_reviewer_validation.sql` - Rejection handling
- `14_pageflow_handle_rejection.sql` - Skip rejected pages in build loop

---

## 5. Navigation Consistency

### Problem
Navigation built from planner output before pages exist. If pages fail or are rejected, nav still links to them.

### Solution
Query pages table directly when injecting header/footer:
```sql
SELECT nav_label, url FROM pages 
WHERE site_id = ? 
  AND in_header = true 
  AND status IN ('deployed', 'active')
ORDER BY nav_order
```

### Files
- `16_navigation_from_pages.go` - `GetHeaderNavFromPages()`, `GetFooterNavFromPages()`
- `17_patch_inject_nav.go` - Patch instructions for InjectHeader/InjectFooter
- `23_nav_labels_datahelper.go` - Shared `SimplifyNavLabel()` function

### Go Changes Required
- Add `FooterNavItems []NavItem` to RenderContext struct
- Update `quick_links` case in template rendering to prefer FooterNavItems
- Call nav query functions in InjectHeader/InjectFooter

---

## 6. Link Constraints for Content Writer

### Problem
LLM creates links to pages that don't exist.

### Solution
`prepare_link_context` action extracts available pages from `db_sync.pages` and builds constraint text. Output stored in `link_context` for prompt inclusion.

Example constraint text:
```
## Internal Links

When creating internal links, ONLY link to these pages:

- /index.html (Home)
- /about.html (About)
- /services.html (Services)
- /contact.html (Contact)

Do NOT create links to pages not in this list.
```

### Files
- `21_prepare_link_context_action.go` - Prepares link context
- `22_page_content_writer_link_context.sql` - Adds step to workflow

### Not Yet Done
Prompt template needs update to include `{{.link_context.link_constraint_text}}`

---

## 7. Maintenance Workflow

### Status
Not started.

### Purpose
- Scan deployed sites for broken links
- Process pages with `status = 'needs_attention'`
- Analyze content for link opportunities
- Suggest new pages based on content gaps
- Track orphan pages

### Proposed Approach
Separate `site-maintenance-agent` triggered on schedule or after build completion. Generates maintenance report and creates tasks for human review.

---

## File Inventory

| File | Purpose |
|------|---------|
| `01_webdesign_agent.sql` | CSS generation with responsibility rules |
| `10_rerender_pages_action.go` | Rerender with contact injection, nav from DB |
| `11_content_reviewer_validation.sql` | Validation step in review workflow |
| `12_validate_page_content_action.go` | Link/email validation action |
| `14_pageflow_handle_rejection.sql` | Skip rejected pages in deploy |
| `16_navigation_from_pages.go` | Nav query functions for header/footer |
| `17_patch_inject_nav.go` | Patch instructions for header/footer injection |
| `21_prepare_link_context_action.go` | Link context for content writer |
| `22_page_content_writer_link_context.sql` | Workflow step for link context |
| `23_nav_labels_datahelper.go` | Shared nav label simplification |
| `25_patch_quick_links_footer.go` | Patch for footer quick_links rendering |
| `26_css_responsibility_barrier.sql` | Component CSS - layout only |

---

## Outstanding Items

1. **Link constraints in prompts** - Update page-content-writer prompt template to include link constraint text

2. **Go patches** - Apply changes to types.go and component_library.go for footer nav

3. **Maintenance workflow** - Design and implement site-maintenance-agent