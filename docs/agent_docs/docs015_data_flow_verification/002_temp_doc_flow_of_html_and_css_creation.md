# HTML Structure & CSS Colour Bugs — Trace Summary

## Bug 1: `<head>` Ends Up Inside `<body>`

### The Call Chain

A page goes through two agents sequentially. Each adds structure:

```
page-content-writer (compile_page_sections)
    │
    ├── buildPageHTML()
    │   Creates: <!DOCTYPE><html><head>small</head><body>{sections}</body></html>
    │
    ├── InjectHeader()
    │   Removes existing <header>, inserts rendered header after <body>
    │
    └── InjectFooter()
        Inserts rendered footer before </body>
    │
    ▼ Output: correct structure
    <!DOCTYPE html>
    <html><head>small</head><body><header>...</header>{sections}<footer>...</footer></body></html>
    │
    │  This HTML is returned as page_content.response.page_html
    │
    ▼
pageflow-builder (assemble_page)
    │
    ├── CleanHTMLString()
    │   Strips markdown fences only — no structural change
    │
    ├── cleanHTMLStructure()              ◄── PROBLEM POINT 1
    │   Deduplicates <head> tags. If an LLM section contained a <head> block,
    │   it keeps the LARGER one regardless of position. This can remove the
    │   correct <head> (before <body>) and keep a misplaced one (inside <body>).
    │
    ├── InjectHead()                      ◄── PROBLEM POINT 2
    │   Does IN-PLACE replacement of <head>...</head>.
    │   If the surviving <head> is inside <body>, the replacement
    │   stays inside <body>. Does NOT enforce correct positioning.
    │
    ├── InjectHeader()                    (second injection — re-replaces header)
    └── InjectFooter()                    (second injection — re-replaces footer)
```

### Why It Breaks

Two independent problems compound:

**cleanHTMLStructure** (multipage_actions.go line 444-483):
```go
// Keeps the LARGER <head>, removes the smaller — wrong heuristic
if secondHeadLen > firstHeadLen {
    html = html[:firstHeadStart] + html[firstHeadEnd+7:]  // removes correct one
}
```
If any LLM-generated section output includes `<head>` tags (which does happen — LLMs sometimes output full HTML pages), the small correct `<head>` gets removed and the large misplaced one (inside body content) survives.

**InjectHead** (component_library.go line 1634-1661):
```go
// Replaces <head> in-place — preserves wrong position
headRe := regexp.MustCompile(`(?is)<head(?:\s[^>]*)?>.*?</head>`)
if headRe.MatchString(html) {
    html = headRe.ReplaceAllString(html, headHTML)  // stays wherever it was
}
```
Should remove ALL `<head>` blocks then insert before `<body>`, not replace in-place.

### The Fixes

**Fix 1 — InjectHead** (component_library.go): Remove all `<head>` blocks, then always insert before `<body>`:
```go
// Step 1: Remove ALL existing <head>...</head>
headRe := regexp.MustCompile(`(?is)<head(?:\s[^>]*)?>.*?</head>`)
html = headRe.ReplaceAllString(html, "")

// Step 2: Always insert before <body>
bodyRe := regexp.MustCompile(`(?i)(<body[^>]*>)`)
html = bodyRe.ReplaceAllString(html, headHTML+"\n$1")
```

**Fix 2 — cleanHTMLStructure** (multipage_actions.go): When deduplicating `<head>`, remove ones that are after `<body>` (misplaced), not the smaller one:
```go
// Find <body> position, remove any <head> that appears after it
bodyPos := strings.Index(lowerHTML, "<body")
for i := len(headPositions) - 1; i >= 0; i-- {
    if headPositions[i][0] > bodyPos {
        // This <head> is inside <body> — remove it
    }
}
```

### Files to Change
- `component_library.go` — `InjectHead` function
- `multipage_actions.go` — `cleanHTMLStructure` function

---

## Bug 2: Light Text on Light Background

### Root Cause

The webdesign-agent's CSS prompt instructs the LLM to generate rules that break colour inheritance in dark sections:

```
Prompt says:        What gets generated:              What it should be:
──────────────────  ──────────────────────────────     ──────────────────────────
h1-h6: color:       h1-h6 { color: var(--primary); }  h1-h6 { color: inherit; }
  var(--primary)
p: color:           p { color: var(--color-text); }    p { margin: 0 0 1rem; }
  var(--color-text)                                      (no color)
```

The generated styles.css also adds:
```css
strong, b { color: var(--color-primary); }     /* should be: no color set */
blockquote { 
    background-color: var(--color-surface);     /* should be: no background */
    color: var(--color-text-muted);             /* should be: no color set */
}
```

### How It Manifests

Dark section components (social-proof, CTA) set `color: #fff` on their container. Children should inherit white text. But the global CSS forces dark colours on `p`, `h1-h6`, `blockquote`, `strong` — overriding the inherited white.

The screenshot shows the testimonials section: blockquotes get a light `#f7fafc` background forced by global CSS, and text colours fight between the component's white and the global dark values.

### The Design System Rule

From the architecture doc's colour inheritance model:
- `body` sets `color: var(--color-text)` — the ONLY default text colour
- `h1-h6` use `color: inherit`
- `p`, `blockquote`, `strong`, `em`, `cite` — do NOT set `color`
- `blockquote` — do NOT set `background-color`
- Dark section components set `color: #fff` on container, children inherit

### The Fixes

**Fix 1 — webdesign-agent prompt** (agent_definitions, webdesign-agent, generate_css step):

Change the Required Base Styles section to:
```
- body { font-family: var(--font-family); color: var(--color-text); line-height: 1.6; }
- h1-h6 { color: inherit; line-height: 1.2; margin: 0 0 1rem; }
- p { margin: 0 0 1rem; }
- a { color: var(--color-accent); }
```

Add a COLOUR INHERITANCE RULES section:
```
- body sets color: var(--color-text) — ONLY place default text colour is set
- h1-h6 MUST use color: inherit (NOT var(--color-primary))
- p, li, blockquote, strong, em, cite — do NOT set color, they inherit
- blockquote — do NOT set background-color (components handle contextually)
- Dark sections set color: #fff on container, ALL children inherit
```

**Fix 2 — leopardessconsulting.co.uk styles.css** (immediate deploy):

```css
/* These lines need changing: */
h1-h6 { color: inherit; }           /* was: color: var(--color-primary) */
p { margin: 0 0 1rem; }             /* was: + color: var(--color-text) */
strong, b { font-weight: 700; }     /* was: + color: var(--color-primary) */
blockquote {                         /* remove background-color and color */
    margin: 0 0 1rem;
    padding: 1rem 1.5rem;
    border-left: 4px solid var(--color-accent);
    font-style: italic;
}
```

### Files/Data to Change
- `agent_definitions` table — webdesign-agent `prompt_template` in `generate_css` step
- `styles.css` for leopardessconsulting.co.uk — git commit via deployer or manual
- All future sites will be correct once the agent definition is updated

---

## Deployment Order

1. Fix `InjectHead` in `component_library.go` — systemic, affects all page builds
2. Fix `cleanHTMLStructure` in `multipage_actions.go` — prevents misplacement
3. Update webdesign-agent prompt in `agent_definitions` — prevents bad CSS for new sites
4. Deploy corrected `styles.css` for leopardessconsulting.co.uk — immediate visual fix
5. Re-run `rerender_pages` for leopardessconsulting.co.uk — applies both fixes to all pages