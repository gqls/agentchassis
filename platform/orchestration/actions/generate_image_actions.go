// internal/backend/agent-chassis/internal/actions/image_actions.go
package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	checks "github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
	"github.com/gqls/agentchassis/platform/orchestration/agenterrors"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// ImageGenerationResult represents the result of an image generation request
type ImageGenerationResult struct {
	Success              bool   `json:"success"`
	ImageURI             string `json:"image_uri,omitempty"`
	ErrorMessage         string `json:"error_message,omitempty"`
	RequestID            string `json:"request_id"`
	ChildOrchestrationID string `json:"child_orchestration,omitempty"`
	ChildResponsesTopic  string `json:"child_responses_topic,omitempty"`
	TargetAgentType      string `json:"target_agent_type"`
	TopicSentTo          string `json:"topic_sent_to"`
	StableIdentity       string `json:"stable_identity"`
	AwaitResponse        bool   `json:"await_response"`
}

// ============================================================================
// Phase 2H — per-kind generation defaults
// ============================================================================

// imageDefaults captures per-kind generation parameters that callers may
// override via inputData. Zero values mean "no opinion; adapter applies
// its own fallback".
type imageDefaults struct {
	NegativePrompt string
	CfgScale       float64
	Steps          int
	Width          int
	Height         int
}

// kindDefaults keyed by the `kind` enum from site_plan_imagery.
// Tuned conservatively — biggest expected win is logo getting any
// negative prompt at all. Tune from observation.
var kindDefaults = map[string]imageDefaults{
	"logo": {
		NegativePrompt: "people, faces, text, signature, watermark, low quality, blurry, photorealistic, complex background",
		CfgScale:       8,
		Steps:          35,
		Width:          1024,
		Height:         1024,
	},
	"hero": {
		NegativePrompt: "text, watermark, signature, low quality, blurry, distorted",
		CfgScale:       7,
		Steps:          30,
		Width:          1024,
		Height:         1024,
	},
	"content_hero": {
		// Phase I3.1 (D14): per-article editorial hero (article page + card
		// source). Neutral defaults matching hero — the visual language comes
		// from the style guide's per-kind override, not from here, so a site
		// whose content heroes ARE photographic isn't fought by a baked-in
		// "no photorealism" negative.
		NegativePrompt: "text, watermark, signature, low quality, blurry, distorted",
		CfgScale:       7,
		Steps:          30,
		Width:          1024,
		Height:         1024,
	},
	"illustration": {
		NegativePrompt: "photorealistic, watermark, text, signature, low quality",
		CfgScale:       7,
		Steps:          30,
		Width:          1024,
		Height:         1024,
	},
	"icon": {
		// SDXL v1.0 only supports a fixed dimension whitelist (see
		// invalid_sdxl_v1_dimensions API error). 1024x1024 is the
		// smallest allowed square. Asset-deployer can downscale at
		// deploy time if smaller files are wanted on disk.
		// Stability's SDXL v1.0 endpoint only accepts a specific list of dimensions: 1024x1024, 1152x896, 1216x832, 1344x768, 1536x640, 640x1536, 768x1344, 832x1216, 896x1152
		NegativePrompt: "background, shadows, photorealistic, text, complex, busy",
		CfgScale:       7,
		Steps:          25,
		Width:          1024,
		Height:         1024,
	},

	"infographic": {
		NegativePrompt: "decorative, blurry, photorealistic, low quality",
		CfgScale:       8,
		Steps:          35,
		Width:          1024,
		Height:         1024,
	},
}

// imageKind is the resolved identity of an image generation: what it is, which
// source said so, whether ANYTHING said so, and whether the sources disagreed.
//
// The last two fields exist because of bugs_open/417. A bare kind string cannot
// distinguish "this is positively not a logo" from "nobody told us anything",
// and it cannot notice a caller that labels a logo generation as something else
// — and BOTH of those are silent routes to an ungoverned prompt, which is the
// defect 417 records. A resolution that carries its own provenance can be acted
// on; a string cannot.
type imageKind struct {
	Kind     string // effective kind; "" when nothing identified this generation
	Signal   string // which source answered ("kind", "step_name", …); "" if none
	Answered bool   // did ANY source identify it, even as a non-logo?
	Conflict string // set when a STATED kind disagrees with the step's own name
}

// stepNameKindHint reads the kind out of a step's NAME. This is the only signal
// the two legacy parents carry (site-work-orchestrator and pageflow-builder call
// generate_image from a step named `call_logo_generation` while mapping no kind
// at all), so without it those paths are ungoverned — which is bugs_open/417's
// own root cause reproduced on a second axis.
//
// [MEASURED 2026-08-31] every live step name containing "logo" or "hero", across
// all active agent definitions: call_logo_gen, call_logo_generation,
// generate_logo, store_logo_asset, deploy_logo_image, check_logo_or_hero.
// An enumeration, not an assumption — the guardian seat rightly refused the
// assertion without it. Note `check_logo_or_hero` names BOTH kinds, which is
// exactly why an ambiguous name returns no hint rather than guessing.
func stepNameKindHint(name string) string {
	n := strings.ToLower(name)
	hasLogo := strings.Contains(n, "logo")
	hasHero := strings.Contains(n, "hero")
	if hasLogo == hasHero {
		// neither, or both (check_logo_or_hero) — no usable hint
		return ""
	}
	if hasLogo {
		return "logo"
	}
	return "hero"
}

// resolveKind picks the effective kind, in order of precedence:
//  1. inputData["kind"]         — caller-supplied via step input_mapping
//  2. inputData["default_kind"] — step-level fallback set by phase_2h.4
//     for legacy callers (call_logo_gen → "logo", call_hero_gen → "hero",
//     call_variant_gen → "hero")
//  3. inputData["purpose"] / inputData["spec"]["purpose"] — the key the STORE
//     path already reads for the same asset (storeImageAsset), so honouring it
//     here removes a disagreement rather than inventing a signal
//  4. the step's own config — purpose, then kind
//  5. the step's NAME — the legacy parents' only signal
//
// This is the ONE kind resolver for this action. A parallel resolver was written
// in round 2 of this change's council review and the reuse seat gated on it:
// two functions resolving the same axis in one file is the second-way-to-do-
// something pattern, and the objection was right — the step-name arm belongs
// HERE, where every caller already reads its kind from.
func resolveKind(inputData map[string]interface{}, agentConfig map[string]interface{}, step models.Step) imageKind {
	stated := func(k, signal string) (imageKind, bool) {
		if k == "" {
			return imageKind{}, false
		}
		r := imageKind{Kind: k, Signal: signal, Answered: true}
		// A stated kind is BELIEVED — that is what keeps hero/icon/imagery out
		// of the logo policy. But when the step's own name says otherwise, the
		// disagreement is recorded rather than swallowed: a caller mislabelling
		// a logo generation would otherwise skip the text policy silently, with
		// no error surface at all (bugs_open/417 on a third axis, raised by the
		// bug_historian seat in round 2). This is a deterministic disagreement
		// between two declarations, not a classifier guessing at the truth.
		if hint := stepNameKindHint(step.Name); hint != "" && hint != k {
			r.Conflict = fmt.Sprintf("step %q implies kind %q but the caller stated %q",
				step.Name, hint, k)
		}
		return r, true
	}

	if k, ok := inputData["kind"].(string); ok {
		if r, done := stated(k, "kind"); done {
			return r
		}
	}
	if k, ok := inputData["default_kind"].(string); ok {
		if r, done := stated(k, "default_kind"); done {
			return r
		}
	}
	if p, ok := inputData["purpose"].(string); ok {
		if r, done := stated(p, "input_purpose"); done {
			return r
		}
	}
	if spec, ok := inputData["spec"].(map[string]interface{}); ok {
		if p, ok := spec["purpose"].(string); ok {
			if r, done := stated(p, "spec_purpose"); done {
				return r
			}
		}
	}
	if step.Config != nil {
		if p, ok := step.Config["purpose"].(string); ok {
			if r, done := stated(p, "step_purpose"); done {
				return r
			}
		}
		if k, ok := step.Config["kind"].(string); ok {
			if r, done := stated(k, "step_kind"); done {
				return r
			}
		}
	}
	if hint := stepNameKindHint(step.Name); hint != "" {
		return imageKind{Kind: hint, Signal: "step_name", Answered: true}
	}
	return imageKind{}
}

// parseAspectRatio takes a "W:H" string (e.g. "16:9", "1:1") plus
// anchor dimensions, returns (width, height, ok). Snaps to multiples
// of 64 per SDXL's requirement; keeps the longer side at the anchor.
func parseAspectRatio(ar string, anchorW, anchorH int) (int, int, bool) {
	parts := strings.Split(ar, ":")
	if len(parts) != 2 {
		return 0, 0, false
	}
	w, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	h, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err1 != nil || err2 != nil || w <= 0 || h <= 0 {
		return 0, 0, false
	}
	anchor := anchorW
	if anchorH > anchor {
		anchor = anchorH
	}
	var width, height int
	if w >= h {
		width = anchor
		height = anchor * h / w
	} else {
		height = anchor
		width = anchor * w / h
	}
	width = (width / 64) * 64
	height = (height / 64) * 64
	if width < 256 || height < 256 {
		return 0, 0, false
	}
	return width, height, true
}

// GenerateImageAction handles image generation requests with dynamic topics
// It should send to the adapter and wait for response
func GenerateImageAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Executing GenerateImageAction action",
		zap.String("agent_type", params.AgentType),
		zap.Any("collected_data_keys", datahelpers.GetMapKeys(params.CollectedData)),
		zap.String("action", params.ExecutionContext.Action),
		zap.Bool("has_db", params.DB != nil),
		zap.String("step_name", params.ExecutionContext.StepName),
		zap.String("orchestration_id", params.ExecutionContext.OrchestrationID),
		zap.String("correlation_id", params.ExecutionContext.CorrelationID),
		zap.Any("DEBUGaa: full params in GenerateImageAction", params),
	)

	// initialise
	if params.ExecutionContext.Action == "initialize" {
		params.Logger.Info("handling initialization")
		return map[string]interface{}{"status": "initialized"}, nil
	}

	// Method 2: Check for initialization flag in collected data
	if isInit, ok := params.CollectedData["is_initialization"].(bool); ok && isInit {
		params.Logger.Info("initialization detected via flag")
		return map[string]interface{}{"status": "initialized"}, nil
	}

	// Method 3: Check the action from collected data (if passed through)
	if action, ok := params.CollectedData["action"].(string); ok && action == "initialize" {
		params.Logger.Info("initialization detected via action field")
		return map[string]interface{}{"status": "initialized"}, nil
	}

	// removed NormalizeCollectedData was destroying accumulated state
	// Only ensure essential topic fields if missing:
	if params.ExecutionContext.RequestsTopic != "" {
		if _, exists := params.CollectedData["__my_requests_topic__"]; !exists {
			params.CollectedData["__my_requests_topic__"] = params.ExecutionContext.RequestsTopic
		}
	}
	if params.ExecutionContext.ReplyToTopic != "" {
		if _, exists := params.CollectedData["__parent_responses_topic__"]; !exists {
			params.CollectedData["__parent_responses_topic__"] = params.ExecutionContext.ReplyToTopic
		}
	}

	// Get the agent's configuration
	var agentConfig map[string]interface{}
	var ok bool

	// First try CollectedData (for orchestrated agents)
	agentConfig, ok = params.CollectedData["agent_config"].(map[string]interface{})
	params.Logger.Info("Checking agent_config in CollectedData",
		zap.Bool("found", ok),
		zap.Bool("is_nil", agentConfig == nil))

	// If not found, load it directly from the database
	if !ok && params.AgentType != "" {
		params.Logger.Info("Agent config not in collected data, loading from database",
			zap.String("agent_type", params.AgentType))

		agentDef, err := loadAgentDefinitionForImageAction(ctx, params.DB, params.AgentType)
		if err != nil {
			params.Logger.Error("Failed to load agent definition",
				zap.String("agent_type", params.AgentType),
				zap.Error(err))
			return nil, fmt.Errorf("failed to load agent definition: %w", err)
		}

		agentConfig = agentDef.DefaultConfig
		params.CollectedData["agent_config"] = agentConfig
	}

	if agentConfig == nil {
		params.Logger.Error("Agent configuration is nil after all attempts")
		return nil, fmt.Errorf("agent configuration not found")
	}

	// Extract input data
	inputData := datahelpers.GetInputData(params.CollectedData, params.Logger)

	// Optional parameters — existing
	style, _ := inputData["style"].(string)
	width, _ := inputData["width"].(int)
	height, _ := inputData["height"].(int)

	// Phase 2H — kind resolution and per-kind defaults
	kindRes := resolveKind(inputData, agentConfig, params.StepConfig)
	kind := kindRes.Kind
	defaults := kindDefaults[kind] // zero value if kind unknown

	// Width / height: caller wins, then per-kind default, then 1024 fallback
	if width == 0 {
		width = defaults.Width
	}
	if width == 0 {
		width = 1024
	}
	if height == 0 {
		height = defaults.Height
	}
	if height == 0 {
		height = 1024
	}

	// Phase 2H — negative_prompt: caller wins; otherwise per-kind default
	negativePrompt := defaults.NegativePrompt
	if np, ok := inputData["negative_prompt"].(string); ok && np != "" {
		negativePrompt = np
	}

	// Phase 2H — fold constraints from imagery spec into negative_prompt.
	// constraints arrives as map[string]interface{} from input_data.spec.constraints.
	// Known keys: no_text, no_logos, no_people, no_humans. Unknown keys ignored.
	if c, ok := inputData["constraints"].(map[string]interface{}); ok {
		var addNegatives []string
		if v, ok := c["no_text"].(bool); ok && v {
			addNegatives = append(addNegatives, "text", "letters", "writing")
		}
		if v, ok := c["no_logos"].(bool); ok && v {
			addNegatives = append(addNegatives, "logo", "brand mark")
		}
		if v, ok := c["no_people"].(bool); ok && v {
			addNegatives = append(addNegatives, "people", "faces", "humans")
		} else if v, ok := c["no_humans"].(bool); ok && v {
			addNegatives = append(addNegatives, "people", "faces", "humans")
		}
		if len(addNegatives) > 0 {
			extra := strings.Join(addNegatives, ", ")
			if negativePrompt != "" {
				negativePrompt += ", " + extra
			} else {
				negativePrompt = extra
			}
		}
	}

	// Phase 2H — style_hints.aspect_ratio overrides width/height if supplied.
	// Format: "16:9", "1:1", "4:3", etc.
	// Phase 2I — also captured as a top-level aspectRatio string so it can
	// be forwarded to the adapter, letting the chosen provider snap to its
	// own valid dimensions (SDXL whitelist vs Gemini's ratio set).
	var aspectRatio string
	if h, ok := inputData["style_hints"].(map[string]interface{}); ok {
		if ar, ok := h["aspect_ratio"].(string); ok && ar != "" {
			aspectRatio = ar
			if w2, h2, ok := parseAspectRatio(ar, width, height); ok {
				width, height = w2, h2
			} else {
				params.Logger.Warn("generate_image: aspect_ratio not understood; keeping default",
					zap.String("aspect_ratio", ar))
			}
		}
	}

	// Phase 2H — cfg_scale and steps. Caller wins; per-kind default otherwise.
	cfgScale := defaults.CfgScale
	if cs, ok := inputData["cfg_scale"].(float64); ok && cs > 0 {
		cfgScale = cs
	}
	steps := defaults.Steps
	if s, ok := inputData["steps"].(int); ok && s > 0 {
		steps = s
	}

	// Phase 2H — seed. 0 = random/unspecified.
	seed, _ := inputData["seed"].(int)

	// Phase I1 — per-site imagery style guide (site_specs aspect
	// 'imagery_style_guide'): the structured brand signal. Loaded once here;
	// used below for the prompt direction (superseding the free-text
	// design_intent.imagery_direction when it yields one), the negative
	// prompt (avoid terms), and reference-image style anchors.
	siteID, _ := inputData["site_id"].(string)
	styleGuide := getImageryStyleGuideForSite(ctx, params.DB, siteID, params.Logger)
	// avoidForKind handles the logo exclusion and D14 per-kind overrides
	// (an override's avoid replaces the guide-level one wholesale).
	if avoid := styleGuide.avoidForKind(kind); avoid != "" {
		if negativePrompt != "" {
			negativePrompt += ", " + avoid
		} else {
			negativePrompt = avoid
		}
	}

	// bugs_open/011 R1 — the site's provider preference. The adapter routes by
	// kind alone because it has no DB access; resolving the preference here and
	// passing it as data is what makes the routing decision site-owned instead
	// of hardcoded. Empty is the norm (no guide, or no preference in it) and
	// leaves the adapter's per-kind default untouched.
	providerHint := styleGuide.providerForKind(kind)
	if providerHint != "" {
		params.Logger.Info("generate_image: site provider preference applied",
			zap.String("site_id", siteID),
			zap.String("kind", kind),
			zap.String("provider_hint", providerHint))
	}

	// Phase 2I — reference image URIs. Plural form preferred (multiple
	// references for style consistency); singular form accepted for
	// backward compat with the Phase 2H field name. Stability v1 REST
	// ignores; Banana fetches via the injected ReferenceFetcher and
	// sends as multimodal inlineData parts.
	var referenceImageURIs []string
	if uris, ok := inputData["reference_image_uris"].([]interface{}); ok {
		for _, u := range uris {
			if s, ok := u.(string); ok && s != "" {
				referenceImageURIs = append(referenceImageURIs, s)
			}
		}
	}
	if len(referenceImageURIs) == 0 {
		if u, ok := inputData["reference_image_uri"].(string); ok && u != "" {
			referenceImageURIs = append(referenceImageURIs, u)
		}
	}
	if len(referenceImageURIs) > 0 {
		params.Logger.Info("generate_image: reference image URIs supplied",
			zap.Int("count", len(referenceImageURIs)),
			zap.String("kind", kind))
	}

	// Phase I1 — style-anchor references from the style guide when the caller
	// supplied none. referenceKeysForKind (D14) gates: per-kind override keys
	// flow ungated (deliberate), guide-level keys remain photographic-kinds
	// only (anchoring an icon or logo to a photographic brand hero invites
	// contamination). Banana consumes these; Stability ignores them.
	if refKeys := styleGuide.referenceKeysForKind(kind); len(referenceImageURIs) == 0 && len(refKeys) > 0 {
		refs := resolveReferenceAssetURIs(ctx, params.DB, siteID, refKeys, params.Logger)
		if len(refs) > 0 {
			referenceImageURIs = refs
			params.Logger.Info("generate_image: style-guide reference anchors applied",
				zap.Int("count", len(refs)),
				zap.String("kind", kind))
		}
	}

	// Extract prompt template
	// Get prompt using three-tier priority system
	promptTemplate, promptSource := getImagePromptWithPriority(params, agentConfig)

	// REFUSE the generic fallback rather than paint from it (bugs_open/210).
	//
	// getImagePromptWithPriority returns ("Generate content based on the
	// provided context.", generic_fallback) when NO caller supplied a prompt.
	// Generating from that produces a meaningless image which the caller then
	// stores AS A BRAND ASSET — store_logo_asset would save it as the site's
	// logo. A silently wrong logo is strictly worse than a loud failure, and
	// this is the one place every image the platform generates passes through,
	// so refusing here does not depend on any producer remembering anything.
	//
	// This is not a theoretical rung. Priorities 2 and 3 do not exist on this
	// fleet: exactly one step runs this action (image-generator's `generate`)
	// and neither its agent default_config nor its step config carries a
	// prompt_template, so it is Priority 1 or this. Measured 2026-08-09:
	// 0 of 344 assets with a recorded origin_prompt were generated from it.
	if promptSource == imagePromptSourceGenericFallback {
		params.Logger.Error("generate_image REFUSED: no prompt supplied by any caller",
			zap.String("agent_type", params.AgentType),
			zap.String("kind", kind),
			zap.String("site_id", siteID))
		return nil, fmt.Errorf("generate_image refused: no prompt supplied for kind %q (site %q) — "+
			"the caller must map a prompt into input_data.prompt or step config; generating from "+
			"the generic fallback would store a meaningless image as a brand asset", kind, siteID)
	}

	// Phase 0.1 / Phase 2I: prepend design_intent.imagery_direction (read from
	// site_specs) when site_id is available AND the kind is photographic.
	// Direction is enrichment, not a requirement — if site_id is missing or no
	// design_intent exists, the prompt is used as-is. Parents pass site_id
	// through input_mapping; see migration phase_0_combined_migration.sql.
	//
	// Flat-vector kinds (icon, logo) are EXCLUDED — see directionAppliesToKind.
	// Prepending a photographic direction ("industrial photography of robotic
	// grippers...") to an icon prompt makes the model render a photo with the
	// icon composited into a corner (observed on icon_cycle_time 2026-05-20).
	if siteID != "" {
		// Phase I1 — the style guide's per-kind direction wins when present
		// (structured, brand-approved, kind-gated inside directionForKind:
		// full voice for photographic kinds, palette-only for icons, nothing
		// for logos). Free-text design_intent.imagery_direction remains the
		// fallback for sites without a guide.
		direction := styleGuide.directionForKind(kind)
		directionSource := "+style_guide"
		if direction == "" {
			if directionAppliesToKind(kind) {
				direction = getImageryDirectionForSite(ctx, params.DB, siteID, params.Logger)
				directionSource = "+imagery_direction"
			} else if kind != "logo" {
				// Flat-vector kinds are excluded from the free-text fallback
				// above for a good reason (a photographic direction prepended
				// to a flat prompt composites the subject onto a photo), and
				// that exclusion left them with NO colour direction at all on
				// any site without a style guide — which is 9 of 10 sites. The
				// only colour such a prompt then carries is build-site-planner's
				// hardcoded "#4A4A4A line on #EEEEEE background", a LIGHT ground
				// shipped to every site regardless of scheme. On dartsonline
				// that produced 17 near-white tiles for a #111520 page.
				//
				// So derive the palette clause from the composed palette. Logos
				// stay excluded — generated once, human-approved, then locked
				// (the 2026-05-20 contamination lesson), and a locked asset must
				// not acquire a new colour instruction on a re-run.
				direction = composedPaletteDirection(ctx, params.DB, siteID, params.Logger)
				directionSource = "+composed_palette"
			}
		}
		if direction != "" {
			composed, truncated := composeImagePromptWithDirection(promptTemplate, direction)
			if truncated {
				// bugs_open/027 §4b: a silently clipped direction reads as a
				// deliberate style choice in the output. Say so, loudly, with
				// what was lost. NOTE for deploy verification: this format
				// string is the marker to grep the pod binary for (log-message
				// literals survive the build; identifier symbols may not —
				// RUNBOOK A6.3).
				params.Logger.Warn("Imagery direction TRUNCATED before generation",
					zap.String("site_id", siteID),
					zap.String("kind", kind),
					zap.Int("direction_len", len(direction)),
					zap.Int("cap", maxImageryDirectionInPrompt))
			}
			params.Logger.Info("Prepended imagery direction",
				zap.String("site_id", siteID),
				zap.String("kind", kind),
				zap.String("source", directionSource),
				zap.Bool("truncated", truncated),
				zap.String("direction_preview", datahelpers.TruncateString(direction, 100)),
				zap.String("composed_preview", datahelpers.TruncateString(composed, 250)))
			promptTemplate = composed
			promptSource = promptSource + directionSource
		} else if !directionAppliesToKind(kind) {
			params.Logger.Info("No imagery direction for flat-vector kind",
				zap.String("site_id", siteID),
				zap.String("kind", kind))
		}
	}

	// bugs_open/417 — LOGO TEXT POLICY, applied to every kind=logo generation
	// regardless of where the prompt came from.
	//
	// The rule ("no lettering") already existed, in discovery_checks/
	// default_brand_prompt.go, but it was coupled to the prompt's SOURCE rather
	// than the asset's PURPOSE: it lived inside the FALLBACK builder, which runs
	// only when a plan supplies no logo prompt. Every planner-built site supplies
	// one, so the rule protected exactly the population that never needed it and
	// the planned prompt reached the model ungoverned. That produced invented
	// brand names on two live sites — "Farm Shield Info" on farmerinsurance.uk
	// and "BOXING NEWS" on boxingonline.com, the first paid customer.
	//
	// WHY HERE. This is the one point every image prompt from every tier and
	// every producer must pass — getImagePromptWithPriority above has already
	// collapsed step config, collected data, input_data, the agent's own
	// prompt_template and the workflow fallback into one value, and the 210
	// refusal a few lines up is the precedent that this function enforces policy.
	// Applying the rule HERE rather than at plan time also governs work items
	// ALREADY QUEUED with an unwashed prompt, which is what migrations 669/670
	// structurally could not do: 670 washed the stored rows, and a plan created
	// 41 seconds later (boxingonline, migration 680) missed it. The model also
	// REWORDS the licence, so no literal match can bound the class — this guard
	// needs no detection at all, which is why it is a bound and not a floor.
	// bugs_open/417 — LOGO TEXT POLICY, driven off the single kind resolution
	// above, which now carries its own provenance (which signal answered, whether
	// ANY did, and whether the declarations disagreed). Round 2 of this change's
	// council review gated on a parallel resolver living beside resolveKind; the
	// seat was right, so the extra signals moved INTO resolveKind and this is
	// now a consumer of one resolution rather than a second opinion.
	//
	// Three durable outcomes, because each is a distinct way an ungoverned
	// prompt reaches the model and none of them is visible in a log:
	//   - the policy applied, and the artefact records which signal let it;
	//   - NOTHING identified the generation, so no per-kind policy could apply;
	//   - a stated kind DISAGREED with the step's own name, so the policy was
	//     skipped on a caller's say-so that another declaration contradicts.
	if kind == "logo" {
		wordmarkText := logoWordmarkTextFromInputs(inputData)
		var ident logoIdentity
		if wordmarkText != "" {
			// Only opting IN costs a query; the default path stays free.
			ident = loadLogoIdentity(ctx, params.DB, siteID, params.Logger)
		}
		var rejected string
		promptTemplate, negativePrompt, promptSource, rejected = applyLogoTextPolicy(
			promptTemplate, negativePrompt, promptSource, wordmarkText, ident, params.Logger)
		// bugs_open/424 — background-key policy, same choke point as the text
		// policy immediately above and for the same reason: this is the one place
		// every logo prompt from every producer and every tier converges before
		// dispatch, so applying it here (rather than in the adapter) means the
		// clause lands in assets.origin_prompt where a census can see it — see
		// LogoBackgroundKeyClause's doc comment.
		promptTemplate, negativePrompt, promptSource = applyLogoBackgroundPolicy(
			promptTemplate, negativePrompt, promptSource, params.Logger)
		promptSource += "+via_" + kindRes.Signal
		if rejected != "" {
			// bugs_open/034's shape, raised as a MEDIUM objection in round 1: a
			// validation outcome that exists only as a Warn is invisible to every
			// census, discoverable only by reading logs after the fact.
			recordImagePolicyEvent(ctx, params, siteID, "logo_wordmark_rejected", "warning",
				fmt.Sprintf("Logo wordmark opt-in REJECTED and degraded to a text-free mark. "+
					"Requested %q; site naming — company_name %q, logo_text %q, domain %q. "+
					"The requested text is not this site's own name, so it was not licensed.",
					rejected, ident.CompanyName, ident.LogoText, ident.Domain),
				map[string]interface{}{
					"requested_wordmark": rejected,
					"company_name":       ident.CompanyName,
					"logo_text":          ident.LogoText,
					"site_domain":        ident.Domain,
					"bug":                "bugs_open/417",
				})
		}
	}

	// The residual detectors. Gated on Answered, NOT on kind == "" — round 2
	// keyed the note on an empty kind, and the editquality seat caught that a
	// hero call setting `purpose` but no `kind` would file a false
	// "nothing identified it" note, polluting the very liveness measurement the
	// note exists to provide. "Did any source answer?" is the question that was
	// actually meant.
	if !kindRes.Answered {
		params.Logger.Warn("image generation with NO resolvable kind — no per-kind "+
			"policy could be applied (bugs_open/417 legacy-parent gap)",
			zap.String("site_id", siteID),
			zap.String("agent_type", params.AgentType),
			zap.String("step", params.StepConfig.Name))
		recordImagePolicyEvent(ctx, params, siteID, "image_generation_without_kind", "warning",
			fmt.Sprintf("An image generation reached GenerateImageAction with NO resolvable "+
				"kind from any source, so no per-kind policy (logo text policy, size "+
				"defaults, negative prompt) could be applied. Fix: give the calling step an "+
				"input_mapping supplying `kind` or `purpose`, as image-build-handler's "+
				"branches do."),
			map[string]interface{}{
				"step":          params.StepConfig.Name,
				"prompt_source": promptSource,
				"bug":           "bugs_open/417",
			})
	} else if kindRes.Conflict != "" {
		// A stated kind that the step's own name contradicts. The stated kind is
		// still believed — a classifier overriding a caller is a worse failure —
		// but the disagreement is now recorded, so a mislabelled logo is
		// findable instead of silently ungoverned.
		params.Logger.Warn("image kind CONFLICT — the caller's kind disagrees with the step name; "+
			"the caller's kind is being used and per-kind policy follows it",
			zap.String("site_id", siteID),
			zap.String("conflict", kindRes.Conflict))
		recordImagePolicyEvent(ctx, params, siteID, "image_kind_conflict", "warning",
			"Image kind declarations disagree: "+kindRes.Conflict+". The caller's kind was "+
				"used, so per-kind policy (including the logo text policy) followed it. If the "+
				"step name is right, the caller is mislabelling the asset and its prompt is "+
				"ungoverned.",
			map[string]interface{}{
				"stated_kind": kind,
				"stated_via":  kindRes.Signal,
				"step":        params.StepConfig.Name,
				"bug":         "bugs_open/417",
			})
	}

	params.Logger.Info("Selected prompt for execution",
		zap.String("source", promptSource),
		zap.String("agent_type", params.AgentType),
		zap.String("kind", kind),
		zap.String("prompt_preview", datahelpers.TruncateString(promptTemplate, 350)))

	// When running within the image-generator agent, we send to the adapter topic
	// The adapter will respond back to us (the image-generator agent)
	imageGeneratorAdapterTopic := "system.adapter.image-generator.requests"

	// Get our own responses topic - where the adapter should respond to
	myResponsesTopic := params.ExecutionContext.ResponsesTopic
	if myResponsesTopic == "" {
		// Fallback to environment variable if not in context
		myResponsesTopic = os.Getenv("RESPONSES_TOPIC")
	}
	if myResponsesTopic == "" {
		// Build it from the context
		myResponsesTopic = fmt.Sprintf("job.%s-%s-%s-%s.responses",
			extractShortCorrelation(params.ExecutionContext.CorrelationID),
			extractShortOrchestration(params.ExecutionContext.OrchestrationID),
			"image-generator",
			params.ExecutionContext.StepName,
		)
	}

	params.Logger.Info("Image-generator agent sending to adapter",
		zap.String("adapter_topic", imageGeneratorAdapterTopic),
		zap.String("my_responses_topic", myResponsesTopic),
		zap.String("DEBUGaa: DEBUGbb: which one: prompt template", promptTemplate),
		zap.String("prompt source", promptSource),
	)

	// Phase 2H — extended imageData. Base fields preserved; new ones omitted
	// when empty/zero so adapter sees identical JSON for legacy callers that
	// don't pass them.
	imageData := map[string]interface{}{
		"prompt": promptTemplate,
		"style":  style,
		"width":  width,
		"height": height,
	}
	if negativePrompt != "" {
		imageData["negative_prompt"] = negativePrompt
	}
	if cfgScale > 0 {
		imageData["cfg_scale"] = cfgScale
	}
	if steps > 0 {
		imageData["steps"] = steps
	}
	if seed > 0 {
		imageData["seed"] = seed
	}

	// Phase 2I — kind, aspect_ratio, reference_image_uris for provider
	// abstraction. Adapter routes by kind; provider snaps aspect_ratio
	// to its own valid dimensions; references go through to Banana
	// (Stability ignores). Fields omitted when empty so legacy adapter
	// builds see identical JSON.
	if kind != "" {
		imageData["kind"] = kind
	}
	if aspectRatio != "" {
		imageData["aspect_ratio"] = aspectRatio
	}
	if providerHint != "" {
		imageData["provider_hint"] = providerHint
	}
	if len(referenceImageURIs) > 0 {
		imageData["reference_image_uris"] = referenceImageURIs
	}
	// bugs_open/424 — the structured half of the background-key policy applied
	// above: the adapter has no DB access and cannot re-derive "this is a
	// governed logo" from prose, so the exact key colour travels as data,
	// alongside (never instead of) the prose clause baked into promptTemplate.
	// Both are derived from the one LogoBackgroundKeyHex constant.
	if kind == "logo" {
		imageData["key_ground"] = checks.LogoBackgroundKeyHex
	}

	newRequestID := uuid.NewString()

	// The adapter expects certain fields in a specific format
	// Build the image generation request for the adapter
	adapterRequest := map[string]interface{}{
		"headers": map[string]interface{}{
			"correlation_id":          params.ExecutionContext.CorrelationID,
			"orchestration_id":        params.ExecutionContext.OrchestrationID,
			"orchestration_name":      params.ExecutionContext.OrchestrationName,
			"parent_orchestration_id": params.ExecutionContext.ParentOrchestrationID,
			"client_id":               params.ExecutionContext.ClientID,
			"step_name":               params.ExecutionContext.StepName,
			"step_id":                 params.ExecutionContext.StepID,
			"request_id":              newRequestID,
			"message_type":            "request",

			// flat sender fields (not nested!)
			"sender_agent_type":    params.ExecutionContext.Sender.AgentType,
			"sender_agent_id":      params.ExecutionContext.OrchestrationID,
			"sender_pod_name":      params.ExecutionContext.Sender.PodName,
			"sender_agent_version": params.ExecutionContext.Sender.AgentVersion,
			"sender_role":          params.ExecutionContext.Sender.Role,

			// Response routing
			"responses_topic":        myResponsesTopic,
			"parent_responses_topic": myResponsesTopic,
		},
		"body": map[string]interface{}{
			"action": "generate",
			"data":   imageData,
			// Tell adapter where to reply
			"reply_to_topic":         myResponsesTopic,
			"parent_responses_topic": myResponsesTopic,
			"metadata": map[string]interface{}{
				"requesting_agent_id":   params.ExecutionContext.OrchestrationID,
				"requesting_agent_type": params.ExecutionContext.Sender.AgentType,
				"requesting_step":       params.ExecutionContext.StepName,
			},
		},
	}

	// produce map[string]string headers for validation
	rawHeaders := adapterRequest["headers"].(map[string]interface{})
	headers := make(map[string]string)

	for k, v := range rawHeaders {
		if str, ok := v.(string); ok {
			headers[k] = str
		} else {
			headers[k] = fmt.Sprintf("%v", v) // fallback stringify
		}
	}

	// Convert to JSON
	messageBytes, err := json.Marshal(adapterRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal adapter request: %w", err)
	}

	// Send to adapter
	key := []byte(params.ExecutionContext.CorrelationID)

	params.Logger.Info("Sending request to image adapter",
		zap.String("topic", imageGeneratorAdapterTopic),
		zap.String("request_id", newRequestID),
		zap.String("reply_to_topic", myResponsesTopic),
		zap.Any("request", adapterRequest),
	)

	// Send the message
	if err := params.Producer.ProduceWithValidation(
		ctx,
		imageGeneratorAdapterTopic,
		headers,
		key,
		messageBytes,
	); err != nil {
		return nil, fmt.Errorf("failed to send to adapter: %w", err)
	}

	// Store tracking info in collected data
	if params.CollectedData != nil {
		params.CollectedData[params.ExecutionContext.StepName] = map[string]interface{}{
			"request_id":      newRequestID,
			"adapter_topic":   imageGeneratorAdapterTopic,
			"responses_topic": myResponsesTopic,
			"prompt":          promptTemplate,
			"status":          "awaiting_adapter_response",
			"width":           width,
			"height":          height,
		}
	}

	// The image-generator agent will receive the response from the adapter
	// on its responses topic and can then complete the workflow
	result := ImageGenerationResult{
		Success:             true,
		RequestID:           newRequestID,
		ChildResponsesTopic: myResponsesTopic,
		TargetAgentType:     "adapter",
		TopicSentTo:         imageGeneratorAdapterTopic,
		AwaitResponse:       true, // Agent needs to wait for adapter response
	}

	params.Logger.Info("Image generation request sent to adapter, awaiting response",
		zap.String("request_id", newRequestID),
		zap.String("expecting_response_on", myResponsesTopic),
	)

	return result, nil
}

// helper function to load agent definition
func loadAgentDefinitionForImageAction(ctx context.Context, db interface{}, agentType string) (*AgentDefinition, error) {

	query := `
		SELECT id, type, display_name, COALESCE(description, ''), category,
		       image_repository, image_tag, 
		       resources, default_config, capabilities, topics,
		       health_config, env_vars, is_active
		FROM agent_definitions
		WHERE type = $1 AND is_active = true
		LIMIT 1
	`

	var def AgentDefinition
	var defaultConfigJSON json.RawMessage // Read as RawMessage first
	var capabilitiesJSON json.RawMessage

	// Handle both *sql.DB and *pgxpool.Pool
	switch d := db.(type) {
	case *sql.DB:
		// For *sql.DB, we need to handle the Command field differently
		err := d.QueryRowContext(ctx, query, agentType).Scan(
			&def.ID,
			&def.Type,
			&def.DisplayName,
			&def.Description,
			&def.Category,
			&def.ImageRepository,
			&def.ImageTag,
			&def.Resources,
			&def.DefaultConfig,
			&def.Capabilities,
			&def.Topics,
			&def.HealthConfig,
			&def.EnvVars,
			&def.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to query agent definition: %w", err)
		}
		// Note: Command field is not loaded here as it's not needed for LLM prompt
	case *pgxpool.Pool:
		err := d.QueryRow(ctx, query, agentType).Scan(
			&def.ID,
			&def.Type,
			&def.DisplayName,
			&def.Description,
			&def.Category,
			&def.ImageRepository,
			&def.ImageTag,
			&def.Resources,
			&def.DefaultConfig,
			&def.Capabilities,
			&def.Topics,
			&def.HealthConfig,
			&def.EnvVars,
			&def.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to query agent definition: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported database type: %T", db)
	}

	// Now unmarshal the JSON into the map
	if len(defaultConfigJSON) > 0 && string(defaultConfigJSON) != "null" {
		if err := json.Unmarshal(defaultConfigJSON, &def.DefaultConfig); err != nil {
			return nil, fmt.Errorf("failed to unmarshal default_config: %w", err)
		}
	}

	// Unmarshal capabilities
	if len(capabilitiesJSON) > 0 {
		json.Unmarshal(capabilitiesJSON, &def.Capabilities)
	}

	// Validate that we have a config
	if def.DefaultConfig == nil {
		return nil, fmt.Errorf("agent %s has no default config", agentType)
	}

	return &def, nil
}

// ProcessImageResponse processes the response from the image adapter
// This is called when the adapter sends back the result
func ProcessImageResponse(params *ActionParams) (interface{}, error) {
	logger := params.Logger.With(
		zap.String("action", "process_image_response"),
		zap.String("step_name", params.ExecutionContext.StepName),
	)

	// Get the response from collected data
	// The response from the adapter should be stored here
	stepData, ok := datahelpers.GetStepData(params.CollectedData, params.ExecutionContext.StepName, logger)
	if !ok {
		// Check if response came in a different way
		if responseData, ok := params.CollectedData["adapter_response"].(map[string]interface{}); ok {
			stepData = responseData
		} else {
			return nil, fmt.Errorf("no response data found from adapter")
		}
	}

	// Extract image  URI  and  URL  from response
	imageURI, _ := stepData["image_uri"].(string)
	imageURL, _ := stepData["image_url"].(string)
	if imageURI == "" {
		// Check in body.data structure
		if body, ok := stepData["body"].(map[string]interface{}); ok {
			if data, ok := body["data"].(map[string]interface{}); ok {
				imageURI, _ = data["image_uri"].(string)
				imageURL, _ = data["image_url"].(string)
			}
		}
	}

	if imageURI == "" {
		// Check for error
		if errorMsg, ok := stepData["error"].(string); ok {
			return nil, fmt.Errorf("image generation failed: %s", errorMsg)
		}
		return nil, fmt.Errorf("no image URI in adapter response")
	}

	// Store result for parent agent with both S3 URI and presigned URL
	result := map[string]interface{}{
		"success":   true,
		"image_uri": imageURI, // S3 URI for storage reference
		"image_url": imageURL, // Presigned URL for web use (valid for 7 days)
	}

	// Update collected data so parent can access the result
	if params.CollectedData != nil {
		if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
			inputData["generated_image_uri"] = imageURI
			inputData["generated_image_url"] = imageURL
			inputData["image_generation_complete"] = true
		}
		params.CollectedData["image_result"] = result
	}

	logger.Info("Image generation complete",
		zap.String("image_uri", imageURI),
		zap.String("image_url", imageURL),
	)

	return result, nil
}

// Helper functions
func extractShortCorrelation(correlationID string) string {
	if len(correlationID) >= 8 {
		return correlationID[:8]
	}
	return correlationID
}

func extractShortOrchestration(orchestrationID string) string {
	if len(orchestrationID) >= 8 {
		return orchestrationID[:8]
	}
	return orchestrationID
}

func createImageGenerationTopics(ctx context.Context, requestsTopic, responsesTopic string, logger *zap.Logger) error {
	// Topic creation logic if needed
	// This might be handled by the platform automatically
	logger.Info("Topics for image generation",
		zap.String("requests", requestsTopic),
		zap.String("responses", responsesTopic),
	)
	return nil
}

// imagePromptSourceGenericFallback is the source label returned when NO tier
// supplied a prompt. GenerateImageAction REFUSES on it (bugs_open/210): the
// accompanying text is meaningless and its output would be stored as a brand
// asset. Named rather than repeated as a literal so the guard and the value it
// guards cannot drift apart — a comparison against a duplicated literal is the
// shape where a later rename silences the guard with no test failing.
const imagePromptSourceGenericFallback = "generic_fallback"

// getPromptWithPriority implements three-tier priority for prompt selection
func getImagePromptWithPriority(params ActionParams, agentConfig map[string]interface{}) (prompt string, source string) {
	logger := params.Logger

	logger.Info("in getPromptWithPriority")

	// PRIORITY 1: Check incoming message for prompt (from parent's call_agent)
	// Check in StepConfig.Config first (this is where call_agent passes it)
	if configPrompt, ok := params.StepConfig.Config["prompt"].(string); ok && configPrompt != "" {
		// Check if prompt contains template syntax and needs interpolation
		if strings.Contains(configPrompt, "{{") && strings.Contains(configPrompt, "}}") {
			logger.Info("Prompt contains template syntax, interpolating against CollectedData",
				zap.String("raw_prompt", configPrompt))

			interpolated, err := datahelpers.RenderPromptTemplate(configPrompt, params.CollectedData, *logger)
			if err != nil {
				logger.Warn("Failed to interpolate prompt template, using raw prompt",
					zap.Error(err),
					zap.String("raw_prompt", configPrompt))
				// Fall through to use raw prompt
			} else if interpolated != "" && interpolated != configPrompt {
				logger.Info("Prompt interpolated successfully",
					zap.String("interpolated_preview", datahelpers.TruncateString(interpolated, 200)))
				configPrompt = interpolated
			}
		}

		logger.Info("Using prompt from step config (Priority 1 - from parent)",
			zap.String("prompt_preview", datahelpers.TruncateString(configPrompt, 100)))
		return configPrompt, "parent_message"
	}
	// Also check in CollectedData["prompt"] as a fallback
	if collectedPrompt, ok := params.CollectedData["prompt"].(string); ok && collectedPrompt != "" {
		logger.Info("Using prompt from collected data (Priority 1 - from parent)",
			zap.String("prompt_preview", datahelpers.TruncateString(collectedPrompt, 100)))
		return collectedPrompt, "parent_message"
	}

	// Check input_data.prompt (from parent's input_mapping resolution)
	// When parent uses input_mapping: {"prompt": "site_plan.response.image_prompts.logo"},
	// the resolved value lands in CollectedData["input_data"]["prompt"]
	inputData := datahelpers.GetInputData(params.CollectedData, logger)
	if inputPrompt, ok := inputData["prompt"].(string); ok && inputPrompt != "" {
		logger.Info("Using prompt from input_data.prompt (Priority 1 - from parent input_mapping)",
			zap.String("prompt_preview", datahelpers.TruncateString(inputPrompt, 100)))
		return inputPrompt, "input_data"
	}

	// PRIORITY 2: Check agent's own default_config.prompt_template
	// This comes from the agent_definitions table for this specific agent type
	if agentPrompt, ok := agentConfig["prompt_template"].(string); ok && agentPrompt != "" {
		logger.Info("Using prompt from agent config (Priority 2 - agent's default)",
			zap.String("agent_type", params.AgentType),
			zap.String("prompt_preview", datahelpers.TruncateString(agentPrompt, 100)))
		return agentPrompt, "agent_default"
	}

	// PRIORITY 3: Check workflow step config (fallback)
	// This is the hardcoded fallback in the workflow definition
	if stepConfig, ok := params.StepConfig.Config["prompt_template"].(string); ok && stepConfig != "" {
		logger.Info("Using prompt from workflow step config (Priority 3 - fallback)",
			zap.String("prompt_preview", datahelpers.TruncateString(stepConfig, 100)))
		return stepConfig, "workflow_fallback"
	}

	// Generic fallback if nothing found
	logger.Warn("No prompt found in any tier, using generic fallback")
	return "Generate content based on the provided context.", imagePromptSourceGenericFallback
}

// Add to platform/orchestration/actions/generate_image_actions.go
// This action stores the image URI/URL after call_agent returns from image-generator

// StoreGeneratedImageAction stores the S3 URI from image generation response
// and sets the relative URL for templates
func StoreGeneratedImageAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(
		zap.String("action", "store_generated_image"),
		zap.String("step_name", params.ExecutionContext.StepName),
	)

	logger.Info("Storing generated image reference",
		zap.Any("collected_data_keys", datahelpers.GetMapKeys(params.CollectedData)))

	// Get configuration
	purpose := "hero_home"
	if cfg, ok := params.StepConfig.Config["purpose"].(string); ok && cfg != "" {
		purpose = cfg
	}

	// Extract site_record to get site_id
	siteRecord := datahelpers.ExtractNestedFieldMap(params.CollectedData, "site_record")
	if siteRecord == nil {
		logger.Error("No site_record found")
		return map[string]interface{}{
			"success": false,
			"error":   "site_record not found",
		}, nil
	}
	siteID, _ := siteRecord["site_id"].(string)
	if siteID == "" {
		siteID, _ = siteRecord["id"].(string)
	}

	// Extract image URI from the call_agent response
	var imageURI string

	// Check possible locations for the image result
	possiblePaths := []string{"hero_image_result", "image_result", "generate_hero_image"}

	for _, path := range possiblePaths {
		if result, ok := params.CollectedData[path].(map[string]interface{}); ok {
			if uri, ok := result["image_uri"].(string); ok && uri != "" {
				imageURI = uri
				break
			}
			if response, ok := result["response"].(map[string]interface{}); ok {
				if uri, ok := response["image_uri"].(string); ok && uri != "" {
					imageURI = uri
					break
				}
			}
		}
	}

	if imageURI == "" {
		logger.Warn("No image URI found in result")
		return map[string]interface{}{
			"success": false,
			"error":   "No image URI in response",
			"skipped": true,
		}, nil
	}

	// Relative URL for templates - assemble_site will create this file
	relativeURL := fmt.Sprintf("/assets/images/%s.jpg", purpose)

	logger.Info("Storing image",
		zap.String("site_id", siteID),
		zap.String("purpose", purpose),
		zap.String("image_uri", imageURI),
		zap.String("relative_url", relativeURL))

	// Store in database. The {purpose}_uri write that used to precede this is
	// GONE (bugs_open/155 / 209 Phase 3, 2026-08-22): a site-wide,
	// last-write-wins cache of a per-asset fact whose last reader was deleted
	// on 2026-08-09 (91dda3243) — measured readerless across Go, live
	// agent_definitions, workflow_templates and active component templates
	// before removal. The source of an asset is its own row (storage_path/url).
	if params.DB != nil {
		// Store relative URL for templates
		urlField := purpose + "_url"
		if err := updateSiteContentField(ctx, params.DB, siteID, urlField, relativeURL); err != nil {
			logger.Error("Failed to store URL", zap.Error(err))
			return map[string]interface{}{
				"success": false,
				"error":   fmt.Sprintf("Failed to update content_data: %v", err),
			}, nil
		}
	}

	// Update collected_data for downstream steps
	if params.CollectedData != nil {
		params.CollectedData[purpose+"_url"] = relativeURL

		if siteRecord != nil {
			if contentData, ok := siteRecord["content_data"].(map[string]interface{}); ok {
				contentData[purpose+"_url"] = relativeURL
			}
		}
	}

	logger.Info("Image reference stored",
		zap.String("purpose", purpose),
		zap.String("s3_uri", imageURI),
		zap.String("relative_url", relativeURL))

	return map[string]interface{}{
		"success":   true,
		"site_id":   siteID,
		"purpose":   purpose,
		"image_uri": imageURI,
		"image_url": relativeURL,
	}, nil
}

// Helper: Update a field in sites.content_data
func updateSiteContentField(ctx context.Context, db interface{}, siteID, field, value string) error {
	query := `
		UPDATE sites 
		SET content_data = jsonb_set(
			COALESCE(content_data, '{}'::jsonb),
			$2::text[],
			to_jsonb($3::text),
			true
		),
		updated_at = NOW()
		WHERE id = $1
	`
	jsonPath := fmt.Sprintf("{%s}", field)

	switch d := db.(type) {
	case *sql.DB:
		_, err := d.ExecContext(ctx, query, siteID, jsonPath, value)
		return err
	case *pgxpool.Pool:
		_, err := d.Exec(ctx, query, siteID, jsonPath, value)
		return err
	}
	return fmt.Errorf("unsupported database type")
}

// ============================================================================
// Phase 0.1 — imagery_direction composition helpers
//
// Today this captures only the strategic-spec layer (site_specs.design_intent).
// Once Phase 1 of doc 030 ships (site_plan_directives + brief renderer), this
// helper should also compose plan-time directives where category=design AND
// subject LIKE 'imagery%'. Tracked in PLAN_imagery_loop_closure.md as a
// successor work item.
// ============================================================================

// maxImageryDirectionInPrompt is the soft cap for the style-direction half of
// a composed image prompt. Set to fit comfortably under SDXL's 75-77 CLIP
// token limit (~350-380 chars total) while leaving 100-150 chars for the
// subject prompt.
//
// Provider-specific notes (relevant when provider routing lands):
//   - SDXL / Stable Diffusion family: hard wall at 77 tokens; this cap is correct
//   - FLUX (T5 encoder): ~256 tokens; cap could be raised to ~600 chars
//   - Imagen / Banana Pro: ~1000+ char effective; cap could be raised significantly
//   - Anthropic vision (audit-only): no meaningful cap; cap should not apply
//
// Until provider routing lands, 200 chars is the safe default for the only
// generation backend (Stability hosted SDXL).
const maxImageryDirectionInPrompt = 200

// getImageryDirectionForSite reads design_intent.imagery_direction from
// site_specs. Returns empty string if not found or on error — never blocks
// image generation.
//
// Handles both *sql.DB and *pgxpool.Pool to match the action's database
// abstraction (params.DB is interface{}).
func getImageryDirectionForSite(ctx context.Context, db interface{}, siteID string, logger *zap.Logger) string {
	if siteID == "" || db == nil {
		return ""
	}
	const query = `
		SELECT data->>'imagery_direction'
		FROM site_specs
		WHERE site_id = $1
		  AND aspect = 'design_intent'
		  AND is_current = true
		LIMIT 1
	`
	var direction sql.NullString
	switch d := db.(type) {
	case *sql.DB:
		err := d.QueryRowContext(ctx, query, siteID).Scan(&direction)
		if err != nil {
			if err != sql.ErrNoRows {
				logger.Warn("getImageryDirectionForSite: query failed",
					zap.String("site_id", siteID),
					zap.Error(err))
			}
			return ""
		}
	case *pgxpool.Pool:
		err := d.QueryRow(ctx, query, siteID).Scan(&direction)
		if err != nil {
			// pgx returns its own no-rows sentinel; we treat any error as miss
			// without warning when the message indicates no rows — keeps logs
			// clean for sites without design_intent yet.
			if !strings.Contains(err.Error(), "no rows") {
				logger.Warn("getImageryDirectionForSite: query failed",
					zap.String("site_id", siteID),
					zap.Error(err))
			}
			return ""
		}
	default:
		logger.Warn("getImageryDirectionForSite: unsupported database type",
			zap.String("site_id", siteID))
		return ""
	}
	if !direction.Valid {
		return ""
	}
	return strings.TrimSpace(direction.String)
}

// directionAppliesToKind reports whether design_intent.imagery_direction
// should be prepended for a given image kind.
//
// imagery_direction describes the brand's PHOTOGRAPHIC style — lighting,
// setting, mood, e.g. "industrial photography of robotic grippers in real
// manufacturing environments". That is correct for photographic kinds
// (hero, illustration, infographic, and the empty/legacy default) but
// actively harmful for flat-vector kinds: prepending a photography directive
// to an icon prompt makes the model render a photograph with the icon
// composited into a corner (observed on icon_cycle_time, 2026-05-20 — Banana
// rendered the full industrial scene plus the requested flat stopwatch icon).
//
// Icons and logos carry their own complete style instructions from the
// planner ("minimalist flat icon ... line illustration style, single light
// grey colour on transparent background, no shadows, no photorealism") and
// must not receive the photographic direction.
//
// content_hero (D14) is a flat duotone editorial illustration kind routed to
// Banana, so it joins the excluded set for the same reason (bugs_open/027
// §5(a)). D14 added content_hero to directionForKind via the per-kind override
// map but NOT here — so an override-less site fell to the default:true branch
// and had its site's photographic free-text direction prepended to a flat
// kind, the exact 2026-05-20 contamination class. This gate governs the
// free-text imagery_direction fallback AND the reference-key fallback
// (referenceKeysForKind); the sibling directionForKind default is fixed in
// lockstep so content_hero is treated as flat in BOTH gating functions — the
// asymmetry §1 identified as the root cause.
func directionAppliesToKind(kind string) bool {
	switch kind {
	case "icon", "logo", "sprite_sheet", "content_hero":
		// sprite_sheet (Phase I2) and content_hero (D14) join the flat-vector
		// class: prepending a photographic direction (or anchoring to
		// photographic brand heroes) to a flat illustration reproduces the
		// 2026-05-20 contamination failure. Palette-only / per-kind-override
		// styling comes via the style guide instead.
		return false
	default:
		return true
	}
}

// composeImagePromptWithDirection prepends a design direction to the subject
// prompt, and reports whether the direction had to be truncated. The direction
// is truncated to maxImageryDirectionInPrompt with a preference for ending at
// a sentence or comma boundary so the truncation reads naturally. The full
// subject prompt is preserved — never truncated — because losing the subject
// is unrecoverable while losing trailing style modifiers is graceful.
//
// The truncated return exists so the caller can be LOUD about the loss
// (bugs_open/027 §4b): a silently clipped direction reads as a deliberate
// style choice in the generated image, which is how a brand palette went
// missing fleet-wide without anything erroring. The bool is returned from
// here — the one scope that knows — rather than recomputed at call sites.
//
// Format is intentionally label-less. SDXL doesn't reason about "Style:" /
// "Subject:" meta-instructions and they cost ~17 tokens of the 77-token
// budget for no model benefit.
func composeImagePromptWithDirection(prompt, direction string) (string, bool) {
	prompt = strings.TrimSpace(prompt)
	direction = strings.TrimSpace(direction)
	if prompt == "" || direction == "" {
		return prompt, false
	}
	truncated := len(direction) > maxImageryDirectionInPrompt
	if truncated {
		// Rune-safe cut (bugs_open/027): a raw byte slice here could split a
		// multi-byte rune landing on the boundary. The boundary preferences
		// below then operate on a valid string.
		cut := datahelpers.SafeCut(direction, maxImageryDirectionInPrompt)
		// Prefer the LAST sentence boundary within the cap — not the first
		// (changed 2026-07-20, bugs_open/027 §4b): with a long leading clause,
		// e.g. the palette composed first, the FIRST '. ' past minKeep is that
		// clause's end, and cutting there discards later clauses that fit.
		// LastIndex keeps every complete sentence that fits. Fall back to
		// comma; fall back to hard cut. The minKeep guard avoids degenerate
		// cases where a boundary near the start would lose most of the
		// direction.
		minKeep := maxImageryDirectionInPrompt / 2
		if idx := strings.LastIndex(cut, ". "); idx > minKeep {
			cut = cut[:idx+1]
		} else if idx := strings.LastIndex(cut, ", "); idx > minKeep {
			cut = cut[:idx]
		}
		direction = strings.TrimSpace(cut)
	}
	// Use period separator if direction doesn't already terminate with one.
	sep := ". "
	if endsWithSentenceBoundary(direction) {
		sep = " "
	}
	return direction + sep + prompt, truncated
}

// ---------------------------------------------------------------------------
// bugs_open/417 — logo text policy
// ---------------------------------------------------------------------------

// logoIdentity is the site's own naming, used to check that an opt-in wordmark
// really is this site's name rather than something a planner LLM invented.
type logoIdentity struct {
	CompanyName string
	LogoText    string
	Domain      string
}

// loadLogoIdentity reads the three naming columns. Deliberately tolerant, like
// DefaultBrandImagePrompt: a DB failure degrades to an empty identity, which
// makes any opt-in fail validation and fall back to a text-free mark. The safe
// side of every failure is the default, never the licence.
func loadLogoIdentity(ctx context.Context, db *sql.DB, siteID string, logger *zap.Logger) logoIdentity {
	var ident logoIdentity
	if db == nil || siteID == "" {
		return ident
	}
	var company, logoText, domain sql.NullString
	err := db.QueryRowContext(ctx, `
		SELECT COALESCE(company_name, ''), COALESCE(logo_text, ''), COALESCE(domain, '')
		FROM sites WHERE id = $1
	`, siteID).Scan(&company, &logoText, &domain)
	if err != nil {
		// Not fatal, but it must be SAID: a silent read failure here would look
		// exactly like "this site has no name", and the consequence (a valid
		// wordmark rejected) is invisible in the artefact.
		logger.Warn("logo text policy: identity read FAILED — any wordmark opt-in "+
			"will degrade to a text-free mark; this is a fault, not an absent name",
			zap.String("site_id", siteID), zap.Error(err))
		return ident
	}
	ident.CompanyName = company.String
	ident.LogoText = logoText.String
	ident.Domain = domain.String
	return ident
}

// logoWordmarkTextFromInputs reads the opt-in field off the existing
// input_data.spec.constraints seam (the same map the Phase 2H block above
// already reads, whose unknown keys it ignores — so this key was inert on every
// path until this guard shipped).
func logoWordmarkTextFromInputs(inputData map[string]interface{}) string {
	c, ok := inputData["constraints"].(map[string]interface{})
	if !ok {
		return ""
	}
	t, _ := c["wordmark_text"].(string)
	return strings.TrimSpace(t)
}

// normaliseBrandToken lowercases and strips everything that is not a letter or
// digit, so "Robot-Hands", "robot hands" and "ROBOTHANDS" all compare equal.
func normaliseBrandToken(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// domainStem reduces "www.boxingonline.com/x" to "boxingonline" — the label
// before the first dot, which is the part a brand name actually echoes.
func domainStem(domain string) string {
	d := strings.ToLower(strings.TrimSpace(domain))
	if i := strings.Index(d, "://"); i >= 0 {
		d = d[i+3:]
	}
	if i := strings.Index(d, "/"); i >= 0 {
		d = d[:i]
	}
	d = strings.TrimPrefix(d, "www.")
	if i := strings.Index(d, "."); i >= 0 {
		d = d[:i]
	}
	return normaliseBrandToken(d)
}

// wordmarkGroundsInIdentity reports whether the requested wordmark is this
// site's own name. It is the check that makes the opt-in safe against its own
// producer: a planner LLM can write this field, so "the field exists" cannot be
// the licence — "the field names THIS site" is.
func wordmarkGroundsInIdentity(text string, ident logoIdentity) bool {
	want := normaliseBrandToken(text)
	if want == "" {
		return false
	}
	for _, candidate := range []string{
		normaliseBrandToken(ident.CompanyName),
		normaliseBrandToken(ident.LogoText),
		domainStem(ident.Domain),
	} {
		if candidate != "" && candidate == want {
			return true
		}
	}
	return false
}

// logoTextNegatives are the negative-prompt tokens that contradict a
// deliberately lettered mark. "watermark" and "signature" are NOT here: they
// forbid artefacts, not the brand's own name.
var logoTextNegatives = map[string]bool{
	"text": true, "letters": true, "lettering": true, "words": true,
	"writing": true, "typography": true, "numerals": true,
}

// stripLogoTextNegatives removes the text prohibitions from a negative prompt so
// an approved wordmark is not fighting the same prompt's own instruction. Order
// and all other tokens are preserved.
func stripLogoTextNegatives(negativePrompt string) string {
	if negativePrompt == "" {
		return ""
	}
	parts := strings.Split(negativePrompt, ",")
	kept := make([]string, 0, len(parts))
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t == "" || logoTextNegatives[strings.ToLower(t)] {
			continue
		}
		kept = append(kept, t)
	}
	return strings.Join(kept, ", ")
}

// appendClause joins a governing clause onto a prompt at a sentence boundary.
func appendClause(prompt, clause string) string {
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return clause
	}
	sep := " "
	if !strings.HasSuffix(prompt, ".") && !strings.HasSuffix(prompt, "!") &&
		!strings.HasSuffix(prompt, "?") {
		sep = ". "
	}
	return prompt + sep + clause
}

// applyLogoTextPolicy is the pure half of the guard, so it is directly testable
// against every shape that matters. It returns the governed prompt, the
// reconciled negative prompt, and the tagged prompt source.
//
// Three behaviours, in order:
//
//  1. DEFAULT (no opt-in): append LogoTextFreeClause and add the lettering terms
//     to the negative prompt as belt. This is the arm that fires on every site.
//
//  2. OPT-IN THAT GROUNDS: append LogoWordmarkClause naming the exact string,
//     and REMOVE the text prohibitions from the negative prompt so the two
//     channels do not contradict each other.
//
//  3. OPT-IN THAT DOES NOT GROUND: degrade to (1) with a Warn — never refuse.
//     Refusing would mint an unhandleable item, which is the bugs_open/210
//     lesson; and a text-free mark is safe by construction, so the degraded
//     result is always publishable. This is also what closes the door against
//     the field's own producer: the only word a logo can carry is the site's
//     own name, so an invented one cannot be smuggled in through the escape
//     hatch that exists for deliberate ones.
//
// Idempotent on LogoTextFreeSentinel, so a prompt already carrying the clause
// (a 670/680-washed row, or a retried generation) gains no second copy.
func applyLogoTextPolicy(
	prompt, negativePrompt, promptSource, wordmarkText string,
	ident logoIdentity,
	logger *zap.Logger,
) (newPrompt, newNegative, newSource, rejectedWordmark string) {
	const tag = "+logo_text_policy"

	if wordmarkText != "" {
		if wordmarkGroundsInIdentity(wordmarkText, ident) {
			logger.Info("logo text policy applied: wordmark opt-in ACCEPTED",
				zap.String("wordmark_text", wordmarkText))
			return appendClause(prompt, checks.LogoWordmarkClause(wordmarkText)),
				stripLogoTextNegatives(negativePrompt),
				promptSource + tag + "+wordmark", ""
		}
		// The whole point of the grounding check. Loud, because the caller asked
		// for something specific and is getting something else.
		logger.Warn("logo text policy: wordmark opt-in REJECTED — the requested text "+
			"is not this site's own name; degrading to a text-free mark",
			zap.String("wordmark_text", wordmarkText),
			zap.String("company_name", ident.CompanyName),
			zap.String("logo_text", ident.LogoText),
			zap.String("domain", ident.Domain))
		promptSource += "+wordmark_rejected"
		rejectedWordmark = wordmarkText
	}

	if strings.Contains(prompt, checks.LogoTextFreeSentinel) {
		logger.Info("logo text policy: clause already present, not duplicated")
		return prompt, negativePrompt, promptSource + tag, rejectedWordmark
	}

	logger.Info("logo text policy applied: text-free mark (default)")
	negatives := negativePrompt
	for _, t := range []string{"text", "lettering", "words"} {
		if !strings.Contains(strings.ToLower(negatives), t) {
			if negatives != "" {
				negatives += ", "
			}
			negatives += t
		}
	}
	return appendClause(prompt, checks.LogoTextFreeClause), negatives, promptSource + tag, rejectedWordmark
}

// logoBackgroundNegatives are the negative-prompt tokens added as belt for the
// background-key policy — bugs_open/424. "checkerboard" and "transparency pattern"
// name this bug's own failure mode directly: the model's observed response to a
// transparency request was to paint the UI symbol for transparency as opaque pixels.
var logoBackgroundNegatives = []string{"checkerboard", "transparency pattern", "magenta", "#ff00ff"}

// applyLogoBackgroundPolicy is the pure half of the background-key guard, mirroring
// applyLogoTextPolicy immediately above — bugs_open/424.
//
// "Transparent background" is not promptable (see LogoBackgroundKeyClause's doc
// comment for why), so this does not attempt to ask for transparency at all. It
// appends the positive key-colour clause, which OVERRIDES any earlier wording about
// transparency or plain backgrounds that a caller's own prompt may already carry —
// the word "transparent" must not survive into a governed logo prompt, because a
// prompt asking for both transparency and a solid key colour hands the model a
// contradiction it resolves unpredictably (bugs_closed/390: co-present instructions
// are adjudicated by the model, not by precedence wording).
//
// Idempotent on LogoBackgroundKeySentinel, same shape as applyLogoTextPolicy's
// idempotence on LogoTextFreeSentinel: a prompt already carrying the clause (a
// retried generation, or a row washed by a future migration) gains no second copy.
func applyLogoBackgroundPolicy(prompt, negativePrompt, promptSource string, logger *zap.Logger) (newPrompt, newNegative, newSource string) {
	const tag = "+logo_background_key_policy"

	if strings.Contains(prompt, checks.LogoBackgroundKeySentinel) {
		logger.Info("logo background-key policy: clause already present, not duplicated")
		return prompt, negativePrompt, promptSource + tag
	}

	logger.Info("logo background-key policy applied: keyed magenta ground (bugs_open/424)")
	negatives := negativePrompt
	for _, t := range logoBackgroundNegatives {
		if !strings.Contains(strings.ToLower(negatives), t) {
			if negatives != "" {
				negatives += ", "
			}
			negatives += t
		}
	}
	return appendClause(prompt, checks.LogoBackgroundKeyClause), negatives, promptSource + tag
}

// recordImagePolicyEvent files a durable row for an outcome that would otherwise
// exist only as a log line.
//
// REUSES the estate's action-level recording family (LogActionEntry /
// agenterrors.Entry) rather than hand-writing an INSERT. Round 2 of this
// change's council review had a hand-rolled doc_notes insert here, and the
// reuse seat objected that a fourth ad-hoc writer was being added for exactly
// the need that family exists to serve. It was right — and the objection was
// worth more than tidiness: the hand-rolled insert used
// subject_type='site', which is NOT one of the eight values
// doc_notes_subject_type_check permits, so EVERY insert would have failed. It
// was best-effort, so it would have failed SILENTLY — nulling out the very
// detector this round offers as its compensating control for the legacy-parent
// gap. A reused writer cannot get its own table's constraints wrong.
//
// Best-effort by construction (agenterrors.Write warns and returns false), so a
// recording failure can never change a generation's disposition.
func recordImagePolicyEvent(ctx context.Context, params ActionParams, siteID, code, severity, message string, context map[string]interface{}) {
	LogActionEntryInheritingProvenance(ctx, params, agenterrors.Entry{
		SiteID:       siteID,
		AgentType:    params.AgentType,
		StepName:     params.StepConfig.Name,
		Action:       "generate_image",
		ErrorCode:    code,
		Severity:     severity,
		ErrorMessage: message,
		Context:      context,
	}, params.Logger)
}

// endsWithSentenceBoundary reports whether s ends with ., !, or ?.
func endsWithSentenceBoundary(s string) bool {
	if s == "" {
		return false
	}
	last := s[len(s)-1]
	return last == '.' || last == '!' || last == '?'
}
