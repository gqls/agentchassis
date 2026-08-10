// FILE: platform/orchestration/actions/discovery_checks/default_brand_prompt.go
//
// OWNER RULING 2026-08-09 (bugs_open/210, needs_logo slug). When a site needs a
// logo and nobody ever planned a prompt for it, the item goes to human review
// for guidance — but that review DEFAULTS to "create a logo that suits the
// mission, target market and the domain character" and proceeds on its own.
//
// The reason is scale, and it is the whole design constraint: there are ~2,000
// domains to populate and the owner cannot author or approve that many logos.
// A disposition that BLOCKS on a human is therefore not a safe default here — it
// is a queue that will never drain. So the default must always produce a usable,
// site-specific prompt, and the human path stays available as an OVERRIDE rather
// than a gate.
//
// WHY THIS DOES NOT REOPEN THE CONTAMINATION LESSON. imagery_style_guide.go
// deliberately gives logos NO style-guide direction ("logos get nothing — the
// 2026-05-20 contamination lesson"), because prepending a site's PHOTOGRAPHIC
// imagery direction to a flat logo prompt makes the model composite the mark
// onto a photograph. This builder reads nothing from the imagery signal. It
// reads BRAND IDENTITY — who the site is, who it serves, what it sounds like —
// which is a different axis, and it is the axis the owner named. The two must
// not be conflated: if you ever find yourself reaching for design_intent
// .imagery_direction or the style guide here, that is the excluded path.
//
// WHY IT LIVES IN THIS PACKAGE. Two producers need it — this package's
// check_placeholder_image_in_use and the actions package's WriteBuildItemsAction
// — and the import runs actions -> discovery_checks, never the reverse. A leaf
// helper here is reachable from both without a new package or a cycle. The name
// is exported for that reason and for no other.
//
// WHAT IT DOES NOT DO. It never returns the empty string for a site that exists:
// the domain alone is enough to write a defensible prompt, and every site has a
// domain. That property is what makes it safe to use as a default, and
// TestDefaultBrandPrompt_NeverEmptyForARealSite pins it — a builder that could
// return "" would hand a caller straight back to the generic-fallback refusal
// (IMG-069) that this whole lane exists to keep unreachable.

package discovery_checks

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// BrandPromptSourceDefault is recorded in the work item spec as `prompt_source`
// whenever the prompt was synthesised here rather than planned. It is the audit
// trail for "this logo was chosen by the default, not by a person" — which is
// exactly the population an operator will want to review first when there is
// time, and the only way to find them later.
const BrandPromptSourceDefault = "default_from_brand_identity"

// BrandPromptSourcePlanned marks a prompt that came from the site plan.
const BrandPromptSourcePlanned = "site_plan"

// siteBrandFacts is the identity signal, normalised. Every field is optional;
// the builder degrades one clause at a time rather than all at once.
type siteBrandFacts struct {
	Name     string
	Domain   string
	Industry string
	Tagline  string
	Audience string
	Tone     string
}

// DefaultBrandImagePrompt builds a logo or hero prompt from the site's brand
// identity. purpose is "logo" or "hero".
//
// It is deliberately tolerant: any query failure or missing spec degrades to a
// thinner prompt rather than an error, because the caller's alternative is
// filing an item that cannot be handled at all.
func DefaultBrandImagePrompt(
	ctx context.Context,
	db *sql.DB,
	siteID uuid.UUID,
	purpose string,
	logger *zap.Logger,
) string {
	facts := loadSiteBrandFacts(ctx, db, siteID, logger)
	return composeBrandImagePrompt(facts, purpose)
}

// loadSiteBrandFacts reads the site row and its identity/audience specs.
//
// ⚠ The `identity` spec's key names are NOT consistent across sites — a census
// on 2026-08-09 found ~70 distinct top-level keys, including
// `target_audience`, `target_market_and_locale`, `audience_and_tone` and
// `industry,target_audience,tone,messaging_pillars` as a single literal key.
// That is the same inconsistency bugs_open/072 records for contact fields. So
// this reads only the three keys that are actually dependable (`industry` 21/21,
// `tagline` 19/21, `tone` 10/21 of sites carrying the spec) and treats
// everything else as a bonus. Do NOT add a long key-guessing ladder here: the
// prompt degrades gracefully by design, and a missing clause costs far less
// than a wrong one.
func loadSiteBrandFacts(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) siteBrandFacts {
	var f siteBrandFacts
	if db == nil {
		return f
	}

	var name, companyName, tagline, domain sql.NullString
	err := db.QueryRowContext(ctx, `
		SELECT domain, name, company_name, tagline
		FROM sites WHERE id = $1
	`, siteID).Scan(&domain, &name, &companyName, &tagline)
	if err != nil && err != sql.ErrNoRows {
		logger.Warn("DefaultBrandImagePrompt: site row read failed", zap.Error(err))
	}
	f.Domain = domain.String
	f.Name = firstNonEmpty(companyName.String, name.String)
	f.Tagline = tagline.String

	var raw []byte
	if err := db.QueryRowContext(ctx, `
		SELECT data FROM site_specs
		WHERE site_id = $1 AND aspect = 'identity' AND is_current = true LIMIT 1
	`, siteID).Scan(&raw); err == nil && len(raw) > 0 {
		var d map[string]interface{}
		if json.Unmarshal(raw, &d) == nil {
			f.Industry = stringField(d, "industry")
			f.Tone = stringField(d, "tone")
			f.Audience = stringField(d, "target_audience")
			if f.Tagline == "" {
				f.Tagline = stringField(d, "tagline")
			}
		}
	}

	return f
}

// composeBrandImagePrompt is the pure half — no DB, so it is directly testable
// against the degraded shapes that matter (identity present, identity absent,
// nothing but a domain).
func composeBrandImagePrompt(f siteBrandFacts, purpose string) string {
	subject := f.Name
	if subject == "" {
		subject = domainCharacter(f.Domain)
	}
	if subject == "" {
		subject = "the site"
	}

	var b strings.Builder

	if purpose == "hero" {
		b.WriteString(fmt.Sprintf("A hero image for %s", subject))
	} else {
		b.WriteString(fmt.Sprintf("A simple, distinctive logo mark for %s", subject))
	}
	if f.Domain != "" {
		b.WriteString(fmt.Sprintf(" (%s)", f.Domain))
	}
	b.WriteString(".")

	if f.Industry != "" {
		b.WriteString(fmt.Sprintf(" Sector: %s.", trimClause(f.Industry)))
	}
	if f.Tagline != "" {
		b.WriteString(fmt.Sprintf(" Positioning: %s.", trimClause(f.Tagline)))
	}
	if f.Audience != "" {
		b.WriteString(fmt.Sprintf(" Intended audience: %s.", trimClause(f.Audience)))
	}
	if f.Tone != "" {
		b.WriteString(fmt.Sprintf(" Brand character: %s.", trimClause(f.Tone)))
	}

	if purpose == "hero" {
		b.WriteString(" Photographic or illustrative, appropriate to the sector, " +
			"with clear space for overlaid headline text and no embedded words or lettering.")
		return b.String()
	}

	// Logo craft constraints. Deliberately about FORM, not about the site's
	// imagery style — see the file header on why that separation matters. The
	// no-lettering instruction is not decoration: generated wordmarks reliably
	// produce malformed text, and this asset is used at favicon size.
	b.WriteString(" Flat vector mark, minimal and geometric, a single clear silhouette that stays " +
		"legible at favicon size, centred on a plain background, no lettering or words, " +
		"no photographic texture, no drop shadows.")

	return b.String()
}

// domainCharacter turns a bare domain into something nameable when the site row
// has no name yet — "robot-hands.com" -> "robot hands". This is the last
// fallback and the reason the builder can never return an empty prompt: at
// ~2,000 domains, most will reach the imagery stage before anyone has written
// them a company name.
func domainCharacter(domain string) string {
	if domain == "" {
		return ""
	}
	host := domain
	if i := strings.Index(host, "/"); i >= 0 {
		host = host[:i]
	}
	// Drop the public suffix — ".co.uk" is two labels, so trim known compound
	// suffixes before the simple one.
	for _, suffix := range []string{".co.uk", ".org.uk", ".me.uk", ".com", ".uk", ".org", ".net", ".io", ".ai"} {
		if strings.HasSuffix(host, suffix) {
			host = strings.TrimSuffix(host, suffix)
			break
		}
	}
	host = strings.NewReplacer("-", " ", "_", " ", ".", " ").Replace(host)
	return strings.TrimSpace(host)
}

func stringField(d map[string]interface{}, key string) string {
	s, _ := d[key].(string)
	return strings.TrimSpace(s)
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

// trimClause caps one interpolated clause. The identity spec's free-text fields
// run to hundreds of characters on some sites (`strategic_brief`-style prose),
// and an image prompt that is 90% positioning copy generates worse than one that
// is 90% craft instruction.
func trimClause(s string) string {
	s = strings.TrimSpace(strings.ReplaceAll(s, "\n", " "))
	const cap = 160
	if len(s) <= cap {
		return strings.TrimRight(s, ".")
	}
	cut := s[:cap]
	if i := strings.LastIndex(cut, " "); i > 40 {
		cut = cut[:i]
	}
	return strings.TrimRight(cut, " .,;:")
}
