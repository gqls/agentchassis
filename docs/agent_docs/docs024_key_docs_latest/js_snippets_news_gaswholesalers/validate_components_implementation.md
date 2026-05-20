# `validate_components` Implementation

Implements the dead `validate_components: true` flag in
`ValidateSitePlanAction`. Scope is deliberately narrow: **resolve each
section name to a real `content_components.function`, or drop it.** It does
NOT deduplicate or make content-intent decisions (see "Scope" below).

Reuses the existing `NormalizeComponentFunction` (~line 31067) and adds
only a display-name/name → function lookup, which normalisation alone
can't do (`"FAQ Section"` normalises to `faq-section`, still not the real
`faq`).

## Scope — what it does and does NOT do

**Does (deterministic, safe):**
- Resolves `call_to_action` → `call-to-action`, `FAQSection` →
  `faq-section` via existing normalisation.
- Resolves `"FAQ Section"` (display_name) → `faq` and any `name` → its
  `function` via DB lookup.
- Drops + logs a section name that resolves to nothing (it would orphan
  the page_component downstream).

**Does NOT (intent decisions it can't make safely):**
- Does NOT deduplicate or strip a `generic-text-block` sitting next to a
  `faq`. Validation cannot know whether that block is a legitimate intro
  or a redundant duplicate. Guessing risks deleting wanted content or
  keeping the empty-FAQ pairing. The duplicate-surface problem is solved
  at the planner prompt (don't emit the bad pairing) and by per-section
  briefs (disambiguate legitimate pairings) — NOT here.

This keeps the action's behaviour predictable: a section name either
resolves to a valid component or is removed for being unresolvable. No
silent intent-based deletions.

## Code

### 1. Helper: load the resolution maps

Add near `loadSiteChromeNames` (same file). One query, three maps.

```go
// componentNameResolver holds lookup maps for resolving section names in a
// site plan to canonical content_components.function values.
type componentNameResolver struct {
	validFunctions map[string]bool   // function -> true
	displayToFunc  map[string]string // lower(display_name) -> function
	nameToFunc     map[string]string // lower(name) -> function
}

// loadComponentNameResolver loads section/element component identity from
// the DB so plan section names can be resolved to a canonical function.
// Returns an empty resolver (not nil) on error so callers can no-op safely.
func loadComponentNameResolver(ctx context.Context, db *sql.DB, logger *zap.Logger) *componentNameResolver {
	r := &componentNameResolver{
		validFunctions: make(map[string]bool),
		displayToFunc:  make(map[string]string),
		nameToFunc:     make(map[string]string),
	}
	if db == nil {
		return r
	}
	rows, err := db.QueryContext(ctx,
		`SELECT "function", name, COALESCE(display_name, '')
		   FROM content_components
		  WHERE component_level IN ('section','element')
		    AND is_active = true
		    AND "function" <> ''`)
	if err != nil {
		logger.Warn("loadComponentNameResolver: query failed", zap.Error(err))
		return r
	}
	defer rows.Close()
	for rows.Next() {
		var fn, name, display string
		if err := rows.Scan(&fn, &name, &display); err != nil {
			continue
		}
		r.validFunctions[fn] = true
		if name != "" {
			r.nameToFunc[strings.ToLower(name)] = fn
		}
		if display != "" {
			r.displayToFunc[strings.ToLower(display)] = fn
		}
	}
	return r
}

// resolve attempts to map a raw section name to a canonical component
// function. Returns (function, true) if resolved, ("", false) if not.
func (r *componentNameResolver) resolve(raw string) (string, bool) {
	if raw == "" {
		return "", false
	}
	// 1. Already a valid function.
	if r.validFunctions[raw] {
		return raw, true
	}
	// 2. Normalise (underscore->hyphen, camelCase->kebab) and re-check.
	norm := NormalizeComponentFunction(raw)
	if norm != raw && r.validFunctions[norm] {
		return norm, true
	}
	// 3. Display-name lookup (handles "FAQ Section" -> "faq").
	if fn, ok := r.displayToFunc[strings.ToLower(raw)]; ok {
		return fn, true
	}
	// 4. Component name lookup (handles a row's `name` differing from `function`).
	if fn, ok := r.nameToFunc[strings.ToLower(raw)]; ok {
		return fn, true
	}
	// 5. Last try: display lookup on the normalised form.
	if fn, ok := r.displayToFunc[strings.ToLower(norm)]; ok {
		return fn, true
	}
	return "", false
}
```

### 2. The validation block in `ValidateSitePlanAction`

Insert AFTER the existing site-chrome-strip `if params.DB != nil { ... }`
block and BEFORE the final
`params.Logger.Info("ValidateSitePlanAction: Complete", ...)` /
`return plan, nil`.

Gated on the existing `validate_components` config flag (currently set
true for site-planner but never read).

```go
	// ── Resolve section names to canonical component functions ───────────
	// Implements config flag `validate_components`. Each section name must
	// map to a real content_components.function. Names that are display
	// names ("FAQ Section"), wrong-case, or underscore variants are
	// normalised/resolved; unresolvable names are dropped and logged.
	// This does NOT deduplicate or make content-intent decisions — it only
	// guarantees every surviving section name is a valid component function.
	validateComponents := false
	if vc, ok := config["validate_components"].(bool); ok {
		validateComponents = vc
	}
	if validateComponents && params.DB != nil {
		resolver := loadComponentNameResolver(ctx, params.DB, params.Logger)
		if len(resolver.validFunctions) > 0 { // only act if we actually loaded components
			for _, p := range pages {
				pm, ok := p.(map[string]interface{})
				if !ok {
					continue
				}
				sectionsRaw, ok := pm["sections"].([]interface{})
				if !ok {
					continue
				}
				resolved := make([]interface{}, 0, len(sectionsRaw))
				for _, s := range sectionsRaw {
					name, ok := s.(string)
					if !ok {
						// Non-string (e.g. a section-brief object) — leave as-is.
						resolved = append(resolved, s)
						continue
					}
					fn, ok := resolver.resolve(name)
					if !ok {
						params.Logger.Warn("ValidateSitePlanAction: dropped unresolvable section name",
							zap.Any("page", pm["name"]),
							zap.String("section", name))
						continue
					}
					if fn != name {
						params.Logger.Info("ValidateSitePlanAction: resolved section name to function",
							zap.Any("page", pm["name"]),
							zap.String("from", name),
							zap.String("to", fn))
					}
					resolved = append(resolved, fn)
				}
				pm["sections"] = resolved
			}
		} else {
			params.Logger.Warn("ValidateSitePlanAction: validate_components set but no components loaded — skipping name resolution")
		}
	}
```

Notes on the design:
- **Non-string entries pass through untouched.** When per-section briefs
  land (objects, not strings), this loop won't mangle them — it only
  resolves string entries. Forward-compatible with the briefs work.
- **Empty-resolver guard.** If the DB load returns nothing (query error,
  empty table), it skips rather than dropping every section. Fail-safe:
  better to pass an unvalidated plan than to empty every page.
- **Drop, don't substitute.** An unresolvable name is removed, not
  replaced with a guess. A wrong substitution would be worse than a
  missing section (which downstream gap-detection can re-surface).

## 3. The gap-planner path

`content-gap-planner` applies via `apply_gap_plan` → `applyNewPage`, which
does NOT route through `validate_site_plan`. So the resolver must also run
there, or the `"faq-section"`-style names from that planner's prompt slip
through.

In `applyNewPage` (`apply_gap_plan_action.go`), where it reads
`newPlan["sections"]` and falls back to the default, resolve each name
before writing the page record:

```go
	sections := []string{"hero", "generic-text-block", "call-to-action"}
	if sectionsRaw, ok := newPlan["sections"].([]interface{}); ok && len(sectionsRaw) > 0 {
		raw := make([]string, 0, len(sectionsRaw))
		for _, s := range sectionsRaw {
			if str, ok := s.(string); ok {
				raw = append(raw, str)
			}
		}
		// NEW: resolve names to canonical functions, drop unresolvable.
		resolver := loadComponentNameResolver(ctx, params.DB, logger)
		if len(resolver.validFunctions) > 0 {
			resolved := make([]string, 0, len(raw))
			for _, name := range raw {
				if fn, ok := resolver.resolve(name); ok {
					resolved = append(resolved, fn)
				} else {
					logger.Warn("applyNewPage: dropped unresolvable section name",
						zap.String("page", pageName), zap.String("section", name))
				}
			}
			if len(resolved) > 0 {
				sections = resolved
			}
		} else if len(raw) > 0 {
			sections = raw // resolver unavailable — use as-is rather than lose them
		}
	}
```

(If `applyNewPage` doesn't currently have `ctx`/`params.DB`/`logger` in
scope, thread them in — they're available on the action's `params`.)

## What this fixes and what it doesn't

Fixes (deterministically):
- `"FAQ Section"` → `faq` (the display-name leak that orphaned the
  component on `containment-first-architecture`).
- `call_to_action` ↔ `call-to-action` inconsistency across sites.
- Any future typo'd/display-named section that would otherwise orphan.

Does NOT fix (by design — handled elsewhere):
- The `generic-text-block` + `faq` duplicate pairing → planner prompt edit
  (Defect 1) + per-section briefs.
- Empty structured components when a pairing does occur → post-build
  validation (Fix D) catches it before deploy.

## Testing it

After deploying the chassis with this change, the existing isolated-build
harness covers it: trigger a build whose plan contains a deliberately
display-named section (e.g. inject `"FAQ Section"` into a test page's
sections) and confirm the resulting `page_components.component_id` is
linked (resolved to `faq`) rather than NULL (orphaned). The same
faq-test page pattern used for the writer test works here — only the
input section name changes.
