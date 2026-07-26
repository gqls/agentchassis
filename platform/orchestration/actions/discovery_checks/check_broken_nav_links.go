// FILE: platform/orchestration/actions/discovery_checks/check_broken_nav_links.go
//
// Detects header/footer navigation that renders anchor links (href="#slug")
// where it should render page URLs, and files work for the nav-link-fixer.
//
// THE ARTEFACT MISMATCH — read this before changing either side
// -------------------------------------------------------------
// The detector reads RENDERED html in site_components. The handler
// (fix_nav_link_templates) never touches that; it rewrites the backing
// content_components.html_template with literal find/replace strings, and a
// re-render then regenerates the rendered html. So "is this fixable?" is a
// question about the TEMPLATE, and a rendered href="#services" produced by a
// template that hardcodes the anchor is a guaranteed no-op for the handler.
//
// That is why the check partitions on the template (bugs_open/077, third
// instance): everything the handler's patterns would not change is filed as a
// capability_gap instead of as work nobody can do.
//
// AND THE REMIT IS NOT ALL IN GO. The seeded agent supplies its own `patterns`
// config, which OVERRIDES the Go defaults entirely (fix_nav_link_templates_action.go:
// "If no patterns configured, use sensible defaults"). Measured 2026-07-26 the
// live nav-link-fixer row carries THREE patterns where DefaultNavLinkPatterns has
// four. So the check reads the live config via HandlerStepConfig and partitions
// on that — partitioning on the Go list would credit the handler with a pattern
// it does not apply, which is the same defect one level down.
//
// The transform and the default pattern list are homed HERE, alongside the check
// that partitions on them, exactly as ReplaceHardcodedColors is: actions imports
// discovery_checks and never the reverse, so this is the only package both the
// handler and the check can reach them from. Do not fork a private copy back into
// package actions — a mirrored predicate drifts, and the drift is invisible until
// it has filed a few hundred items.

package discovery_checks

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"go.uber.org/zap"
)

func init() { Register(&BrokenNavLinksCheck{}) }

// ============================================================================
// The handler's transform (shared with fix_nav_link_templates)
// ============================================================================

// NavLinkPattern is one literal find/replace the handler applies to a template.
type NavLinkPattern struct {
	Find    string
	Replace string
}

// DefaultNavLinkPatterns is what fix_nav_link_templates applies when its step
// config supplies no `patterns`. It is NOT necessarily what the live agent runs —
// see the file header. Anything relying on the real remit must go through
// NavLinkPatternsForHandler.
var DefaultNavLinkPatterns = []NavLinkPattern{
	{Find: `href="#{{.slug}}"`, Replace: `href="{{.url}}"`},
	{Find: `href="#{{ .slug }}"`, Replace: `href="{{ .url }}"`},
	{Find: `href="#{{.name}}"`, Replace: `href="{{.url}}"`},
	{Find: `href="#{{ .name }}"`, Replace: `href="{{ .url }}"`},
}

// ApplyNavLinkPatterns is THE transform the fix_nav_link_templates action applies
// to a template, returning the new template and how many replacements were made.
// The check uses it as the remit predicate, so keep it the single copy.
func ApplyNavLinkPatterns(template string, patterns []NavLinkPattern) (string, int) {
	out := template
	applied := 0
	for _, p := range patterns {
		if p.Find == "" {
			continue
		}
		if n := strings.Count(out, p.Find); n > 0 {
			out = strings.ReplaceAll(out, p.Find, p.Replace)
			applied += n
		}
	}
	return out, applied
}

// ParseNavLinkPatterns reads a step config's `patterns` array in the shape the
// action parses it. Returns nil when the config carries none, which is the
// action's own signal to fall back to DefaultNavLinkPatterns.
func ParseNavLinkPatterns(config map[string]interface{}) []NavLinkPattern {
	raw, ok := config["patterns"].([]interface{})
	if !ok {
		return nil
	}
	var out []NavLinkPattern
	for _, p := range raw {
		pMap, ok := p.(map[string]interface{})
		if !ok {
			continue
		}
		find, _ := pMap["find"].(string)
		replace, _ := pMap["replace"].(string)
		if find != "" {
			out = append(out, NavLinkPattern{Find: find, Replace: replace})
		}
	}
	return out
}

// navLinkHandlerAgent / navLinkHandlerAction identify the handler whose remit
// this check partitions against.
const (
	navLinkHandlerAgent  = "nav-link-fixer"
	navLinkHandlerAction = "fix_nav_link_templates"
)

// NavLinkPatternsForHandler resolves the patterns the LIVE handler would apply,
// mirroring the action's own precedence: step config if it names any, otherwise
// the Go defaults. The bool reports whether the handler agent is registered at
// all — if it is not, there is no remit and the whole population is residue.
func NavLinkPatternsForHandler(dctx DiscoveryCheckContext) ([]NavLinkPattern, bool, error) {
	cfg, exists, err := HandlerStepConfig(dctx.Ctx, dctx.DB, navLinkHandlerAgent, navLinkHandlerAction)
	if err != nil {
		return nil, false, err
	}
	if !exists {
		return nil, false, nil
	}
	if configured := ParseNavLinkPatterns(cfg); len(configured) > 0 {
		return configured, true, nil
	}
	return DefaultNavLinkPatterns, true, nil
}

// ============================================================================
// The check
// ============================================================================

type BrokenNavLinksCheck struct{}

func (c *BrokenNavLinksCheck) Name() string { return "broken_nav_links" }

func (c *BrokenNavLinksCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	broken, err := findBrokenNavLinks(dctx)
	if err != nil {
		return nil, err
	}
	if len(broken) == 0 {
		return &CheckResult{}, nil
	}

	patterns, handlerExists, err := NavLinkPatternsForHandler(dctx)
	if err != nil {
		return nil, err
	}

	// Partition on the TEMPLATE — the artefact the handler edits — not on the
	// rendered html the detector matched.
	var inRemit, residue []brokenNavLinkFinding
	for _, f := range broken {
		if !handlerExists || f.Template == "" {
			// No handler, or no template behind the slot for one to rewrite.
			residue = append(residue, f)
			continue
		}
		if _, applied := ApplyNavLinkPatterns(f.Template, patterns); applied > 0 {
			inRemit = append(inRemit, f)
		} else {
			residue = append(residue, f)
		}
	}

	result := &CheckResult{
		Findings: []map[string]interface{}{{
			"check":        "broken_nav_links",
			"count":        len(inRemit),
			"population":   len(broken),
			"out_of_remit": len(residue),
			"findings":     broken,
		}},
	}

	for _, finding := range inRemit {
		specJSON, _ := json.Marshal(map[string]interface{}{
			"check":        "broken_nav_links",
			"slot_name":    finding.SlotName,
			"link_count":   finding.LinkCount,
			"example_href": finding.ExampleHref,
			"fix": "Template uses #{{.slug}} — should use {{.url}}. " +
				"Fix template in content_components, then force re-render site_components.",
		})

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Pipeline:     "build",
			ItemType:     "broken_nav_links",
			Severity:     "high",
			Summary:      fmt.Sprintf("Navigation in %s uses anchor links (#slug) instead of page URLs", finding.SlotName),
			SpecJSON:     string(specJSON),
			Priority:     40,
			HandlerAgent: navLinkHandlerAgent,
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("broken_nav_links:%s", finding.SlotName),
			BatchID:      dctx.BatchID,
		})
	}

	if len(residue) > 0 {
		examples := make([]RemitCandidate, 0, len(residue))
		for _, f := range residue {
			examples = append(examples, RemitCandidate{Key: f.SlotName + " " + f.ExampleHref})
		}
		gapKind := GapHandlerRemit
		capability := "rewrite anchor-style nav links whose TEMPLATE does not contain one of the handler's " +
			"literal find strings — anchors hardcoded in the template, or built by any expression other than " +
			"the seeded #{{.slug}} / #{{.name}} forms. Literal find/replace cannot reach those; the fix needs " +
			"the template's link construction understood, not string-matched."
		if !handlerExists {
			gapKind = GapHandlerMissing
			capability = "agent " + navLinkHandlerAgent + " is not registered, so no nav link can be repaired at all"
		}
		result.WorkItems = append(result.WorkItems, CapabilityGapItem(dctx, CapabilityGap{
			Check:         "broken_nav_links",
			Pipeline:      "build",
			BuilderNeeded: navLinkHandlerAgent,
			GapKind:       gapKind,
			Capability:    capability,
			Population:    len(broken),
			Residue:       len(residue),
			Examples:      examples,
			CodePointers:  navLinkCodePointers,
		}))
	}

	return result, nil
}

var navLinkCodePointers = []map[string]string{
	{
		"path": "platform/orchestration/actions/discovery_checks/check_broken_nav_links.go",
		"why":  "ApplyNavLinkPatterns and DefaultNavLinkPatterns — the literal find/replace that IS the remit",
	},
	{
		"path": "docs/agent_docs/sql_for_agents/042b_nav_link_fixer_agent.sql",
		"why":  "the seeded patterns OVERRIDE the Go defaults, so widening the remit may be a seed rather than a build",
	},
	{
		"path": "platform/orchestration/actions/fix_nav_link_templates_action.go",
		"why":  "the handler; it writes content_components.html_template, which is why the partition reads the template",
	},
}

type brokenNavLinkFinding struct {
	SlotName    string `json:"slot_name"`
	LinkCount   int    `json:"link_count"`
	ExampleHref string `json:"example_href"`
	// Template is the backing content_components.html_template — what the handler
	// actually rewrites, and therefore what the remit test must be applied to.
	// Empty when the slot has no template component assigned.
	Template string `json:"-"`
}

func findBrokenNavLinks(dctx DiscoveryCheckContext) ([]brokenNavLinkFinding, error) {
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT sc.slot_name,
		       (LENGTH(sc.rendered_html) - LENGTH(REPLACE(sc.rendered_html, 'href="#', '')))
		           / LENGTH('href="#') as link_count,
		       SUBSTRING(sc.rendered_html FROM 'href="(#[a-zA-Z][^"]*)"') as example_href,
		       cc.html_template
		FROM site_components sc
		LEFT JOIN content_components cc ON sc.component_id = cc.id
		WHERE sc.site_id = $1
		  AND sc.slot_name IN ('header', 'footer')
		  AND sc.rendered_html IS NOT NULL
		  AND sc.rendered_html ~ 'href="#[a-zA-Z]'
		ORDER BY sc.slot_name
	`, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("broken_nav_links query failed: %w", err)
	}
	defer rows.Close()

	var findings []brokenNavLinkFinding
	for rows.Next() {
		var f brokenNavLinkFinding
		var exampleHref, template sql.NullString
		if err := rows.Scan(&f.SlotName, &f.LinkCount, &exampleHref, &template); err != nil {
			dctx.Logger.Warn("Failed to scan broken nav link", zap.Error(err))
			continue
		}
		if exampleHref.Valid {
			f.ExampleHref = exampleHref.String
		}
		if template.Valid {
			f.Template = template.String
		}
		findings = append(findings, f)
	}
	return findings, rows.Err()
}
