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
	err = db.QueryRowContext(ctx, `
		SELECT data FROM site_specs
		WHERE site_id = $1 AND aspect = 'identity' AND is_current = true LIMIT 1
	`, siteID).Scan(&raw)
	switch {
	case err == sql.ErrNoRows:
		// Expected on most sites — 18 of 39 carry no identity spec — so this is
		// Info, not Warn. It still has to be SAID: the prompt that results is
		// thinner, and a reader looking at a bare-domain logo needs to be able
		// to tell "this site had no identity recorded" from "the read broke".
		logger.Info("DefaultBrandImagePrompt: no identity spec for this site; "+
			"prompt will be built from the domain alone",
			zap.String("site_id", siteID.String()))
	case err != nil:
		// A DB blip must NOT read as "this site has no brand". Across a
		// 2,000-domain rollout an unlogged failure here produces silently
		// degraded logos fleet-wide with no signal at all, indistinguishable
		// from working — the council's bug_historian seat named this shape and
		// it is the platform's most recurrent one.
		logger.Warn("DefaultBrandImagePrompt: identity spec read FAILED — "+
			"the prompt is degraded, and this is a fault, not an absent spec",
			zap.String("site_id", siteID.String()), zap.Error(err))
	case len(raw) > 0:
		var d map[string]interface{}
		if uerr := json.Unmarshal(raw, &d); uerr != nil {
			logger.Warn("DefaultBrandImagePrompt: identity spec is not parseable JSON — "+
				"the prompt is degraded",
				zap.String("site_id", siteID.String()), zap.Error(uerr))
		} else {
			f.Industry = stringField(d, "industry")
			f.Tone = stringField(d, "tone")
			f.Audience = stringField(d, "target_audience")
			if f.Tagline == "" {
				f.Tagline = stringField(d, "tagline")
			}
		}
	}

	// One line per built prompt saying how much signal it actually had. This is
	// what makes a fleet-wide degradation visible as a pattern rather than as
	// 2,000 individually-plausible logos: `clauses:1` everywhere means the
	// identity reads are failing, not that the estate has no brands.
	logger.Info("DefaultBrandImagePrompt: composed",
		zap.String("site_id", siteID.String()),
		zap.Int("clauses", f.signalCount()),
		zap.Bool("has_name", f.Name != ""),
		zap.Bool("has_industry", f.Industry != ""))

	return f
}

// signalCount is how many identity clauses the prompt will carry. A site with
// only a domain scores 1 legitimately; the whole FLEET scoring 1 is a fault.
func (f siteBrandFacts) signalCount() int {
	n := 0
	for _, v := range []string{f.Name, f.Industry, f.Tagline, f.Audience, f.Tone} {
		if v != "" {
			n++
		}
	}
	if f.Domain != "" {
		n++
	}
	return n
}

// LogoTextFreeClause is the estate's ruled text policy for generated logo marks,
// stated POSITIVELY and as an explicit override of anything earlier in the prompt.
//
// WHY IT IS A CONSTANT AND NOT A SENTENCE IN ONE FUNCTION (bugs_open/417). Until
// 2026-08-31 this rule lived only inside composeBrandImagePrompt below — the
// FALLBACK builder, which runs only when a site plan supplies no logo prompt at
// all. Every planner-built site supplies one, so the rule protected exactly the
// population that never needed it, and the ruled path was, in the bug file's
// words, "the fallback nobody reaches". The rule is now applied at the
// generation choke point to every kind=logo prompt from every producer
// (applyLogoTextPolicy in generate_image_actions.go); this constant is what both
// sites share so they cannot drift.
//
// WHY POSITIVE AND WHY "OVERRIDES". [MEASURED 2026-08-31] The failing
// boxingonline generation DID receive a text prohibition: the banana adapter
// folded negative_prompt (including "text") into the positive prompt at
// 12:55:50Z, prompt_len 232 -> 407, and the model still lettered "BOXING NEWS".
// A folded prohibition demonstrably LOSES to a positive licence sitting in the
// same prompt ("no text other than the wordmark itself" presupposes a wordmark).
// So the rule has to be a positive instruction that explicitly voids the earlier
// wording; the negative channel is belt, never the mechanism. This also corrects
// an assumption the old wording rested on — a purely negative clause was weaker
// than its own comment claimed, for as long as it stood.
//
// The favicon-size rationale still holds and is why the default is text-free at
// all: generated wordmarks reliably produce malformed text, and this asset is
// re-derived into a favicon and an og_card.
const LogoTextFreeClause = "Render a text-free mark: a single pictorial symbol, " +
	"one composition on one plain background, containing no lettering, words, " +
	"letters, numerals or typography of any kind. The brand name is set in HTML " +
	"beside the logo and must never be painted into the image. This instruction " +
	"overrides any earlier wording in this prompt that mentions, permits or " +
	"presupposes a wordmark or any text."

// LogoTextFreeSentinel is the idempotence key for LogoTextFreeClause and the
// substring a census greps for in assets.origin_prompt to prove the policy
// REACHED a generation. Kept separate from the clause so a future rewording of
// the clause does not silently break idempotence on prompts already carrying it.
const LogoTextFreeSentinel = "Render a text-free mark"

// LogoWordmarkClause is the OPT-IN escape hatch: a deliberately lettered logo.
//
// OWNER RULING 2026-08-31: logos carry no words BY DEFAULT, and a lettered logo
// is allowed only when someone names the EXACT string it must read. That shape
// is the point — the field's value IS the text, so "a wordmark" with no text
// named (the bugs_open/417 licence, which is what let the model invent
// "Farm Shield Info" and "BOXING NEWS") cannot be expressed at all. Per the
// 2026-08-02 owner ruling on new authority on a shared seam, this is an opt-in
// field whose unsafe side is OFF by default and which is visible to a reviewer
// of the caller rather than licensed by a comment.
func LogoWordmarkClause(text string) string {
	return fmt.Sprintf("The only text in the image is the exact wordmark %q, "+
		"spelled exactly as written, rendered once, as a single composition on "+
		"one plain background; no other lettering, words or numerals anywhere. "+
		"This instruction overrides any earlier wording in this prompt about text.",
		text)
}

// LogoBackgroundKeyHex, LogoBackgroundKeySentinel and LogoBackgroundKeyClause —
// bugs_open/424.
//
// "Transparent background" is not a promptable property of this estate's image
// models: Gemini's image family (the banana provider) has no alpha-channel output
// at all, so asking for transparency makes it paint the checkerboard PICTURE of
// transparency as opaque pixels — verified by PNG chunk scan (colour type 2, no
// tRNS) on the asset that exposed this. No prompt wording closes that gap, because
// the property being asked for does not exist in the model's output space.
//
// The fix is architectural, not textual: ask for something the model CAN paint —
// a flat, deterministic, saturated colour no real mark would ever use — and remove
// it mathematically after generation (KeyOutBackground, package imagegenerator).
// This constant is the prompt half of that pair; the two must never drift apart,
// which is why both the clause text and the value handed to the matting step are
// derived from the one hex constant, never restated.
//
// WHY MAGENTA, AND WHY A FIXED CONSTANT NOT A SITE COLOUR. #FF00FF is maximally
// distant from the greys, blacks, whites and metallics a real mark is likely to
// use, and unlike black or white it never reads as a "natural" ground a model
// might blend a mark into. It is deliberately NOT derived from the site's brand
// palette: imagery_style_guide.go already excludes logo prompts from palette
// direction (the 2026-05-20 contamination lesson, see the file header above) —
// tying the key colour to a site's own palette would reopen exactly that.
const LogoBackgroundKeyHex = "#FF00FF"

// LogoBackgroundKeySentinel is the idempotence key for LogoBackgroundKeyClause and
// the substring a census greps for in assets.origin_prompt to prove the policy
// REACHED a generation — the same role LogoTextFreeSentinel plays for the text
// rule. Kept separate from the clause so a future reword of the clause does not
// silently break idempotence on prompts already carrying it.
const LogoBackgroundKeySentinel = "single flat, uniform, edge-to-edge field of pure magenta"

// LogoBackgroundKeyClause is stated positively and as an explicit override, the
// same shape as LogoTextFreeClause and for the same reason (bugs_open/417
// measured that a folded NEGATIVE prohibition demonstrably loses to a positive
// licence sitting earlier in the same prompt — the converse is assumed to hold
// here too until measured otherwise).
const LogoBackgroundKeyClause = "The entire background is a " + LogoBackgroundKeySentinel +
	" (" + LogoBackgroundKeyHex + "), with no gradient, vignette, shadow, glow, texture, " +
	"panel or border, and the artwork must not touch the image edges. This instruction " +
	"overrides any earlier wording in this prompt about transparency, a plain background, " +
	"or any other ground colour."

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
	// imagery style — see the file header on why that separation matters.
	b.WriteString(" Flat vector mark, minimal and geometric, a single clear silhouette that stays " +
		"legible at favicon size, centred on a plain background, " +
		"no photographic texture, no drop shadows.")
	// The text rule is no longer written out here. It is LogoTextFreeClause, so
	// that this fallback path and the generation choke point cannot drift apart
	// — bugs_open/417 was exactly that drift, in the other direction: the rule
	// existed ONLY here, and every planned prompt reached the model without it.
	// A parity test pins the two together.
	b.WriteString(" " + LogoTextFreeClause)

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
