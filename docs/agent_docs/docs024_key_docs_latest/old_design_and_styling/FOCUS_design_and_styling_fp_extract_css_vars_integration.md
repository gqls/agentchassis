# Integration: Robust CSS Variable Extraction

## Changes to extract_design_fingerprint_action.go

### 1. Remove from regex block (line ~63)

Delete this line:
```go
fpCSSVarRe      = regexp.MustCompile(`(--[\w-]+)\s*:\s*([^;}{]+)`)
```

### 2. Remove old function (line ~388)

Delete the entire `fpExtractCSSVars` function:
```go
func fpExtractCSSVars(css string, vars map[string]string) {
	for _, match := range fpCSSVarRe.FindAllStringSubmatch(css, -1) {
		name := strings.TrimSpace(match[1])
		value := strings.TrimSpace(match[2])
		if fpIsDesignVar(name) {
			vars[name] = value
		}
	}
}
```

### 3. Add the new code

Either:
- Paste the contents of `fp_extract_css_vars_replacement.go` into the
  existing file (remove the duplicate `package actions` and `import` lines)
- Or create it as a new file in the same package (the `package actions`
  line is already correct)

Since the file declares functions that replace `fpExtractCSSVars` with the
same signature, existing callers (`extract_design_fingerprint_action.go`
and `enrich_fingerprint_with_css_action.go`) work unchanged.

### 4. What this fixes

| Problem | Before | After |
|---------|--------|-------|
| BEM selectors `.btn--primary:hover` | Captured as `--primary: hover` | Rejected — not inside a token block |
| Commented-out variables | Captured as live values | Stripped before parsing |
| Minified CSS | Regex might not match | Semicolon-splitting handles it |
| Variables on body/html/.dark | Not found (only regex-scanned) | Found via Strategy 2 |
| Multiple :root blocks (media queries) | Only found first match sometimes | All blocks found |
| Values with complex expressions | Regex captured partial values | Full value up to semicolon |
| No :root at all (Tailwind, utility CSS) | Nothing found | Falls back to frequency analysis (existing) |

### 5. How it handles different CSS sources

**Clean hand-written CSS (gamedesign.uk):**
```css
:root {
    --bg-color: #121212;
    --primary-color: #00bcd4;
}
```
→ Strategy 1 finds :root, extracts both variables. Highest confidence.

**Minified production CSS:**
```css
:root{--bg-color:#121212;--primary-color:#00bcd4}body{margin:0}
```
→ Strategy 1 finds :root, semicolon-split handles no-whitespace.

**WordPress theme with body variables:**
```css
body { --wp-primary: #2271b1; --wp-bg: #f0f0f0; }
```
→ Strategy 2 finds body block, extracts variables.

**Dark mode via data attribute:**
```css
[data-theme="dark"] { --bg-color: #1a1a2e; --text: #ffffff; }
```
→ Strategy 2 finds [data-theme block, extracts variables.

**Tailwind / utility CSS (no custom properties):**
→ No variables found by any strategy. That's correct — design data
  comes from frequency analysis of actual colour/font values instead.
  The existing fpExtractColors and fpExtractFonts handle this.

**CSS with BEM naming:**
```css
.card--primary:hover { background: var(--color-accent); }
```
→ Not inside a token block, never parsed. No false positives.

**Commented-out old values:**
```css
/* :root { --old-color: #ff0000; } */
:root { --color: #00bcd4; }
```
→ Comments stripped first. Only live values extracted.
