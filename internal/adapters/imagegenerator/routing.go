// internal/adapters/imagegenerator/routing.go
//
// bugs_open/011 R1 — image provider routing policy, as data plus one pure
// function. Kept out of dynamic_adapter.go deliberately: routing is decided
// per request and had, until this file existed, no test that did not require a
// Kafka consumer and two live provider clients.
//
// WHY THIS IS A TABLE AND NOT A SWITCH. The previous form was a hand-maintained
// switch whose `default:` branch silently routed to Stability. That default is
// the mechanism behind two separate shipped defects — content_hero (found at
// the D13 card gate) and then hero (bugs_open/011, found when a gibberish
// diagram shipped as a client homepage). In both cases the kind simply was
// never added to the Banana branch, and nothing anywhere said so: generation
// reported success, brand anchors were dropped without a word, and the defect
// surfaced only when a human looked at the image.
//
// A table makes the routed set enumerable, so "is this kind routed?" is a
// question the code can actually ask — which is what lets routeProvider flag an
// unmigrated kind instead of quietly serving it from the weaker provider.
// Adding a kind is still a code change; noticing that you forgot is no longer
// left to chance.
package imagegenerator

import (
	"sort"
	"strings"
)

// Provider names as they appear in routing decisions and in the
// imagery_style_guide's `provider` field. These are the only two accepted
// spellings — aliases were considered and dropped (council gate e996bf0a,
// editquality): the field is written by planners and humans against this
// documented contract, so an alias set is surface area with no caller.
const (
	providerBanana    = "banana"
	providerStability = "stability"
)

// kindProviderRouting maps an image kind to the provider that serves it.
//
// Everything with a declared kind is on Banana: it renders legible text, holds
// a flat-illustration house style, and is the only provider that honours
// ReferenceImageURIs — without which per-site brand anchoring is structurally
// impossible, not merely degraded.
//
// The migration happened one kind at a time as each was found to be wrong:
//   - logo/illustration/infographic — 2026-07-10. A site's logo is the
//     reference image every other asset anchors to, so it has to be generated
//     by the provider that reads reference images (RUNNING_NOTES.md turn 4
//     decision A6; AUDIT_verified_facts.md C5).
//   - sprite_sheet — Phase I2. One coherent N×M grid of flat glyphs; Banana's
//     gridded-composition tendency is the feature.
//   - content_hero — Phase I3.1 (D14), after the D13 card gate failed on style
//     drift: SDXL ignores free-text style direction at card size.
//   - hero — 2026-07-18 (bugs_open/011). The last kind left behind, and left
//     by omission rather than decision. It was also the largest: 84 of 155
//     site_plan_imagery rows. Sites that genuinely want SDXL heroes set
//     provider:"stability" in their imagery_style_guide — data, not code.
var kindProviderRouting = map[string]string{
	"icon":         providerBanana,
	"logo":         providerBanana,
	"illustration": providerBanana,
	"infographic":  providerBanana,
	"sprite_sheet": providerBanana,
	"content_hero": providerBanana,
	"hero":         providerBanana,
}

// routingDecision is the outcome of routeProvider: which provider serves the
// request, and what the caller should warn about.
type routingDecision struct {
	// Provider is the resolved provider name (providerBanana | providerStability).
	Provider string

	// BadHint is true when a non-empty provider_hint was not understood. The
	// request still routes (by kind), because a typo in one site's style guide
	// must not take out image generation for that site.
	BadHint bool

	// UnmigratedKind is true when a non-empty kind is absent from
	// kindProviderRouting and therefore fell back to Stability. This is the
	// content_hero/hero failure shape and always deserves a loud warning.
	// An EMPTY kind does not set this: legacy callers that predate the field
	// are a documented, deliberate Stability path, not an oversight.
	UnmigratedKind bool
}

// routeProvider decides which provider serves a request.
//
// Precedence, highest first:
//  1. An explicit, understood provider_hint — the site's own preference,
//     resolved from its imagery_style_guide by GenerateImageAction. The
//     adapter has no DB access, so this is the only way site context reaches
//     the routing decision.
//  2. The kind's entry in kindProviderRouting.
//  3. Stability, as the fallback — flagged via UnmigratedKind whenever a
//     non-empty kind reached it, since that means the kind was never routed.
func routeProvider(kind, providerHint string) routingDecision {
	switch strings.ToLower(strings.TrimSpace(providerHint)) {
	case "":
		// No preference; fall through to kind routing.
	case providerBanana:
		return routingDecision{Provider: providerBanana}
	case providerStability:
		// A deliberate opt-out — e.g. a photographic site keeping SDXL heroes.
		// Not an unmigrated kind: someone chose this.
		return routingDecision{Provider: providerStability}
	default:
		// Understood as "no usable preference", but the caller warns.
		d := routeProvider(kind, "")
		d.BadHint = true
		return d
	}

	if p, ok := kindProviderRouting[strings.TrimSpace(kind)]; ok {
		return routingDecision{Provider: p}
	}
	return routingDecision{
		Provider:       providerStability,
		UnmigratedKind: strings.TrimSpace(kind) != "",
	}
}

// knownRoutedKinds returns the routed kinds, sorted — for the unmigrated-kind
// warning, so the log line itself tells whoever reads it what the valid set is
// rather than sending them to the source.
func knownRoutedKinds() []string {
	kinds := make([]string, 0, len(kindProviderRouting))
	for k := range kindProviderRouting {
		kinds = append(kinds, k)
	}
	sort.Strings(kinds)
	return kinds
}
