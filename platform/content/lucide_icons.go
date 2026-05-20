// FILE: <place in the package that fills component content_data, e.g.
//        platform/content/lucide_icons.go — adjust the package clause to match>
//
// Lucide icon-name allowlist + validation for components whose schema has an
// `icon` field rendered as <i data-lucide="...">  (the `features` grid is the
// first such component). Lucide renders the glyph by name at page load; an
// invalid/hallucinated name renders NOTHING — that is the "missing icons"
// failure mode.
//
// SINGLE SOURCE OF TRUTH: this allowlist is BOTH
//   (a) the set the content LLM is told to choose from (see AllowedLucideIcons), and
//   (b) the set ValidateLucideIcon enforces before content_data is stored.
// Keeping one list prevents drift between "what the model picks" and "what
// will actually render".
//
// ⚠️ VERIFY ONCE against your bundled Lucide version before relying on this.
// Lucide renames/deprecates icons across major versions (e.g. "tool" → "wrench").
// Every name below is a long-stable core icon, but confirm against the version
// you actually ship (see verify_and_wire_lucide.md) and prune any that don't
// render. The render-time fallback (LucideFallback) protects against residual
// misses regardless.

package content

import (
	"sort"
	"strings"
)

// LucideFallback is substituted for any name not in the allowlist. It MUST
// itself be a guaranteed-valid Lucide name. "circle" is minimal and neutral;
// while debugging you may prefer "help-circle" so misses are visually obvious.
const LucideFallback = "circle"

// lucideAllowlist — permitted Lucide icon names, kebab-case (matching the
// data-lucide attribute). Grouped by theme for readability only; grouping has
// no runtime effect. The ten confirmed-rendering names from robot-hands.com's
// deployed features sections are included and marked.
var lucideAllowlist = map[string]struct{}{
	// data / catalog / structure
	"database": {}, "layers": {}, "box": {}, "boxes": {}, "package": {},
	"archive": {}, "folder": {}, "file": {}, "file-text": {}, "files": {},
	"list": {}, "layout-grid": {}, "table": {}, "component": {},
	// charts / analysis
	"bar-chart-2": {}, "pie-chart": {}, "trending-up": {}, "trending-down": {},
	"activity": {}, "gauge": {},
	// tools / config / calculation
	"calculator": {}, "sliders": {}, "settings": {}, "wrench": {}, "cog": {},
	"ruler": {}, "scale": {}, "filter": {}, "search": {},
	// workflow / process
	"git-merge": {}, "git-branch": {}, "workflow": {}, "share-2": {},
	"check": {}, "check-circle": {}, "check-square": {}, "list-checks": {},
	"clipboard": {}, "clipboard-check": {}, "clipboard-list": {},
	// trust / verification / security
	"shield": {}, "shield-check": {}, "lock": {}, "key": {}, "eye": {},
	"badge-check": {}, "award": {},
	// knowledge / guides
	"book-open": {}, "book": {}, "graduation-cap": {}, "lightbulb": {},
	"info": {}, "help-circle": {}, "bookmark": {},
	// targeting / precision
	"target": {}, "crosshair": {}, "focus": {}, "zap": {},
	// hardware / industry
	"cpu": {}, "server": {}, "hard-drive": {}, "factory": {}, "truck": {},
	// generic neutral (safe fallbacks)
	"circle": {}, "square": {}, "hexagon": {},
}

// NormalizeLucideIcon lowercases, trims and converts common separators to
// kebab-case so minor LLM formatting variance ("Bar Chart 2", "bar_chart_2",
// " Database ") still matches the allowlist.
func NormalizeLucideIcon(name string) string {
	n := strings.ToLower(strings.TrimSpace(name))
	n = strings.ReplaceAll(n, " ", "-")
	n = strings.ReplaceAll(n, "_", "-")
	return n
}

// ValidateLucideIcon returns a guaranteed-renderable Lucide name and whether
// the input was already valid. Invalid/empty names map to LucideFallback.
func ValidateLucideIcon(name string) (string, bool) {
	n := NormalizeLucideIcon(name)
	if n == "" {
		return LucideFallback, false
	}
	if _, ok := lucideAllowlist[n]; ok {
		return n, true
	}
	return LucideFallback, false
}

// AllowedLucideIcons returns the sorted allowlist. Use it to build the
// content-LLM prompt so the model is told exactly which names are valid,
// guaranteeing the list it picks from equals the list validation enforces.
func AllowedLucideIcons() []string {
	out := make([]string, 0, len(lucideAllowlist))
	for k := range lucideAllowlist {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// SanitizeFeatureIcons walks a features-style content_data map and replaces
// any invalid icon name with the fallback, in place. Returns the number of
// replacements (log it so bad LLM names are visible). Expected shape:
//
//	{"features": [ {"icon": "...", "name": "...", "description": "..."}, ... ]}
//
// Call this after the content LLM returns and BEFORE writing content_data to
// page_components. Items with no "icon" key are left alone — the features
// template guards with {{if .icon}} so a missing icon is fine; only a PRESENT
// but INVALID name is the problem this fixes.
func SanitizeFeatureIcons(contentData map[string]interface{}) int {
	feats, ok := contentData["features"].([]interface{})
	if !ok {
		return 0
	}
	replaced := 0
	for _, f := range feats {
		fm, ok := f.(map[string]interface{})
		if !ok {
			continue
		}
		raw, ok := fm["icon"].(string)
		if !ok {
			continue
		}
		if valid, wasValid := ValidateLucideIcon(raw); !wasValid {
			fm["icon"] = valid
			replaced++
		}
	}
	return replaced
}
