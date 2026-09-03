// FILE: platform/orchestration/datahelpers/unified_extractor.go
// THE MASTER EXTRACTOR - Uses all our helper functions together

package datahelpers

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/types"
)

// ExtractFields is THE ONLY function you should use to extract fields
// It uses ALL our existing helpers in the right order
// When "input_data" is in fieldNames, this function:
//  1. Extracts core fields (domain, objective, model) to ROOT level -> {{.domain}} works
//  2. ALSO creates result["input_data"] map with same data -> {{.input_data.domain}} works
//
// This ensures backwards compatibility while supporting both template patterns.
func ExtractFields(
	collectedData map[string]interface{},
	fieldNames []string,
	logger *zap.Logger,
) map[string]interface{} {

	logger.Info("=== MASTER EXTRACTOR START ===",
		zap.Strings("requested_fields", fieldNames),
		zap.Strings("available_keys", getMapKeys(collectedData)),
	)

	result := make(map[string]interface{})

	// Special handling for "input_data" field
	if contains(fieldNames, "input_data") {
		logger.Info("Special case: extracting input_data (dual access pattern)")

		// Use aggressive search to get core data
		coreData := ExtractCoreInputData(collectedData, logger)

		// Create the input_data map for {{.input_data.domain}} access
		inputDataMap := make(map[string]interface{})

		// Populate both root level AND input_data map
		for k, v := range coreData {
			result[k] = v       // Root level: {{.domain}}
			inputDataMap[k] = v // Nested: {{.input_data.domain}}
			logger.Info("Added to both root and input_data",
				zap.String("key", k),
				zap.String("type", fmt.Sprintf("%T", v)),
			)
		}

		// Also try to get input_data map directly for any additional fields
		if existingInputMap := getInputDataMap(collectedData, logger); existingInputMap != nil {
			for k, v := range existingInputMap {
				if _, exists := result[k]; !exists {
					result[k] = v       // Root level
					inputDataMap[k] = v // Nested
					logger.Info("Added additional field from input_data map",
						zap.String("key", k),
					)
				}
			}
		}

		// Store the input_data map so {{.input_data.domain}} works
		result["input_data"] = inputDataMap
		logger.Info("Created input_data map for nested access",
			zap.Strings("input_data_keys", getMapKeys(inputDataMap)),
		)
	}

	// ========================================================================
	// Special handling for "current_section" - needed for loop iterations in content generation
	// IMPORTANT: Check top-level FIRST before recursive search to avoid finding
	// input_mapping.current_section (which contains the literal string "current_section")
	// ========================================================================
	if contains(fieldNames, "current_section") {
		logger.Info("Special case: extracting current_section")

		var currentSection interface{}
		var source string

		// PRIORITY 1: Direct top-level lookup (set by setLoopVariable)
		if val, exists := collectedData["current_section"]; exists {
			// Check if it's a valid section map (not a literal variable name)
			if csMap, ok := val.(map[string]interface{}); ok {
				// Valid map - use it
				currentSection = csMap
				source = "direct_top_level"
				logger.Info("✓ Found current_section at top level as map",
					zap.Strings("keys", getMapKeys(csMap)),
				)
			} else if str, ok := val.(string); ok {
				// It's a string - check if it's the literal variable name (bug case)
				if str == "current_section" {
					logger.Warn("current_section contains literal variable name, will try fallback",
						zap.String("value", str),
					)
					// Don't use this value - fall through to fallback
				} else if len(str) > 0 {
					// Try to parse as JSON
					if parsed := tryParseJSON(val, logger); parsed != nil {
						if csMap, ok := parsed.(map[string]interface{}); ok {
							currentSection = csMap
							source = "direct_top_level_json_parsed"
							logger.Info("✓ Parsed current_section from JSON string at top level",
								zap.Strings("keys", getMapKeys(csMap)),
							)
						}
					}
				}
			}
		}

		// PRIORITY 2: Use __current_loop_item_key__ to get the actual loop item
		if currentSection == nil {
			if itemKey, ok := collectedData["__current_loop_item_key__"].(string); ok && itemKey != "" {
				if item, exists := collectedData[itemKey]; exists {
					if csMap, ok := item.(map[string]interface{}); ok {
						currentSection = csMap
						source = "loop_item_key_fallback"
						logger.Info("✓ Retrieved current_section via __current_loop_item_key__ fallback",
							zap.String("item_key", itemKey),
							zap.Strings("keys", getMapKeys(csMap)),
						)
					} else if str, ok := item.(string); ok && len(str) > 20 {
						// Loop item might be JSON serialized
						if parsed := tryParseJSON(item, logger); parsed != nil {
							if csMap, ok := parsed.(map[string]interface{}); ok {
								currentSection = csMap
								source = "loop_item_key_json_parsed"
								logger.Info("✓ Parsed current_section from loop item JSON",
									zap.String("item_key", itemKey),
									zap.Strings("keys", getMapKeys(csMap)),
								)
							}
						}
					}
				}
			}
		}

		// PRIORITY 3: Check process_sections_loop_item (generic loop item key)
		if currentSection == nil {
			if item, exists := collectedData["process_sections_loop_item"]; exists {
				if csMap, ok := item.(map[string]interface{}); ok {
					currentSection = csMap
					source = "generic_loop_item"
					logger.Info("✓ Found current_section via process_sections_loop_item",
						zap.Strings("keys", getMapKeys(csMap)),
					)
				}
			}
		}

		// PRIORITY 4: Recursive search (last resort, but avoid config areas)
		if currentSection == nil {
			// Only search in specific safe areas, not in __raw_message__ or agent_config
			safeKeys := []string{"input_data", "section_components", "render_context"}
			for _, safeKey := range safeKeys {
				if safeData, exists := collectedData[safeKey].(map[string]interface{}); exists {
					if found := findFieldRecursive(safeData, "current_section", 0, logger); found != nil {
						if csMap, ok := found.(map[string]interface{}); ok {
							currentSection = csMap
							source = "recursive_in_" + safeKey
							logger.Info("✓ Found current_section via recursive search",
								zap.String("searched_in", safeKey),
								zap.Strings("keys", getMapKeys(csMap)),
							)
							break
						}
					}
				}
			}
		}

		// Store result
		if currentSection != nil {
			result["current_section"] = currentSection
			logger.Info("current_section extraction complete",
				zap.String("source", source),
			)
		} else {
			logger.Warn("✗ Could not find current_section via any method")
		}
	}

	// ========================================================================
	// Special handling for "reviewed_brief" - often deeply nested in input_data
	// This field is commonly needed for site planning and content generation
	// ========================================================================
	if contains(fieldNames, "reviewed_brief") {
		logger.Info("Special case: extracting reviewed_brief (aggressive recursive search)")

		// Use findFieldRecursive which searches through all nested structures
		if reviewedBrief := findFieldRecursive(collectedData, "reviewed_brief", 0, logger); reviewedBrief != nil {

			// NEW: Unwrap .response wrapper if present
			// reviewed_brief often comes as {"response": {...actual data...}, "response_status": "complete"}
			if rbMap, ok := reviewedBrief.(map[string]interface{}); ok {
				if response, ok := rbMap["response"].(map[string]interface{}); ok {
					// Check if this looks like a wrapped response (has response_status sibling)
					if _, hasStatus := rbMap["response_status"]; hasStatus {
						logger.Info("✓ Unwrapping reviewed_brief.response wrapper",
							zap.Strings("response_keys", getMapKeys(response)))
						// Use the unwrapped response as the reviewed_brief
						reviewedBrief = response
						rbMap = response
					}
				}

				// Log what we found for debugging
				logger.Info("✓ Found reviewed_brief",
					zap.Strings("keys", getMapKeys(rbMap)),
					zap.Bool("has_company_name", rbMap["company_name"] != nil),
					zap.Bool("has_about_us", rbMap["about_us"] != nil),
					zap.Bool("has_services", rbMap["services"] != nil),
					zap.Bool("has_tagline", rbMap["tagline"] != nil),
					zap.Bool("has_contact_email", rbMap["contact_email"] != nil),
				)
			} else {
				logger.Info("✓ Found reviewed_brief (non-map type)",
					zap.String("type", fmt.Sprintf("%T", reviewedBrief)))
			}

			result["reviewed_brief"] = reviewedBrief
		} else {
			logger.Error("✗ Could not find reviewed_brief anywhere in collected data",
				zap.Strings("top_level_keys", getMapKeys(collectedData)))
		}
	}

	// ========================================================================
	// Special handling for "db_sync" - contains navigation from sync_pages_to_db
	// ========================================================================
	if contains(fieldNames, "db_sync") {
		logger.Info("Special case: extracting db_sync")

		if dbSync := findFieldRecursive(collectedData, "db_sync", 0, logger); dbSync != nil {
			result["db_sync"] = dbSync

			// Log navigation info for debugging
			if dsMap, ok := dbSync.(map[string]interface{}); ok {
				hasNav := false
				navCount := 0
				if nav, ok := dsMap["navigation"].(map[string]interface{}); ok {
					hasNav = true
					if items, ok := nav["items"].([]interface{}); ok {
						navCount = len(items)
					}
				}
				logger.Info("✓ Found db_sync",
					zap.Bool("has_navigation", hasNav),
					zap.Int("nav_item_count", navCount),
					zap.Bool("db_available", dsMap["db_available"] == true),
				)
			}
		} else {
			logger.Warn("✗ Could not find db_sync")
		}
	}

	// ========================================================================
	// Special handling for "site_record" - also commonly needed
	// ========================================================================
	if contains(fieldNames, "site_record") {
		logger.Info("Special case: extracting site_record")

		if siteRecord := findFieldRecursive(collectedData, "site_record", 0, logger); siteRecord != nil {
			result["site_record"] = siteRecord
			logger.Info("✓ Found site_record")
		} else {
			logger.Warn("✗ Could not find site_record")
		}
	}

	// ========================================================================
	// Special handling for "current_page" - needed for content generation
	// ========================================================================
	if contains(fieldNames, "current_page") {
		logger.Info("Special case: extracting current_page")

		if currentPage := findFieldRecursive(collectedData, "current_page", 0, logger); currentPage != nil {
			result["current_page"] = currentPage
			if cpMap, ok := currentPage.(map[string]interface{}); ok {
				logger.Info("✓ Found current_page",
					zap.String("name", fmt.Sprintf("%v", cpMap["name"])),
					zap.String("title", fmt.Sprintf("%v", cpMap["title"])),
				)
			} else {
				logger.Info("✓ Found current_page (non-map type)")
			}
		} else {
			logger.Warn("✗ Could not find current_page")
		}
	}

	// ========================================================================
	// Track which fields we've already handled specially
	// ========================================================================
	// The set lives at package level (speciallyHandledInputFields) rather than
	// here, so an offline check can ask THIS package what the extractor treats
	// specially instead of keeping its own copy — bugs_open/453 names an
	// un-sourced copy of exactly this list as the way a lint over the seam
	// inherits the classifier gap it was written to close.

	// Extract each specific field (skip already handled)
	for _, fieldName := range fieldNames {
		if IsSpeciallyHandledInputField(fieldName) {
			continue // Already handled above
		}

		logger.Info(">>> Extracting field", zap.String("field", fieldName))

		value := extractSingleField(collectedData, fieldName, make(map[string]bool), logger)

		if value != nil {
			// Unwrap LLM output wrappers like {type: "text", result: "{...}"}
			// so Go templates can traverse directly into parsed fields.
			// Safe on already-clean values — returns them unchanged.
			value = UnwrapDeep(value, logger)

			// Store with simple name (last part of path) — the rule itself is
			// TemplateRootForInputField, so a reader asking "what does this
			// input_fields entry make available to the template?" gets the same
			// answer the extractor acts on.
			simpleKey := TemplateRootForInputField(fieldName)
			result[simpleKey] = value

			logger.Info("✓ Field extracted",
				zap.String("requested", fieldName),
				zap.String("stored_as", simpleKey),
				zap.String("type", fmt.Sprintf("%T", value)),
			)
		} else {
			logger.Warn("✗ Field not found",
				zap.String("field", fieldName),
			)
		}
	}

	// Always ensure domain and objective exist at root level
	ensureCoreFields(result, collectedData, fieldNames, logger)

	// Also ensure they exist inside input_data if that map exists
	if inputDataMap, ok := result["input_data"].(map[string]interface{}); ok {
		syncCoreFieldsToInputData(result, inputDataMap, logger)
	}

	logger.Info("=== MASTER EXTRACTOR COMPLETE ===",
		zap.Int("fields_extracted", len(result)),
		zap.Strings("result_keys", getMapKeys(result)),
	)

	return result
}

// convertToMapIfPossible tries to convert interface{} to map[string]interface{}
func convertToMapIfPossible(data interface{}, logger *zap.Logger) map[string]interface{} {
	// Already a map
	if m, ok := data.(map[string]interface{}); ok {
		return m
	}

	// Try unwrapping
	unwrapped := UnwrapDeep(data, logger)
	if m, ok := unwrapped.(map[string]interface{}); ok {
		return m
	}

	// Try JSON parsing
	if parsed := tryParseJSON(data, logger); parsed != nil {
		if m, ok := parsed.(map[string]interface{}); ok {
			return m
		}
	}

	return nil
}

// syncCoreFieldsToInputData ensures domain/objective/model exist in input_data map
// if they were recovered at root level by ensureCoreFields
func syncCoreFieldsToInputData(root map[string]interface{}, inputData map[string]interface{}, logger *zap.Logger) {
	coreFields := []string{"domain", "objective", "model"}

	for _, field := range coreFields {
		if _, existsInInputData := inputData[field]; !existsInInputData {
			if value, existsAtRoot := root[field]; existsAtRoot {
				inputData[field] = value
				logger.Info("Synced core field to input_data",
					zap.String("field", field),
				)
			}
		}
	}
}

// extractSingleField tries the inner chain's arms, in order, to find ONE field.
//
// THE ARM NAMES ARE DESCRIPTIVE, NOT NUMBERED (RFC_029 §9 D4, owner-delegated
// ruling 2026-08-15). This chain used to number its arms "Strategy 1..5" while
// ExtractActionInputs (action_inputs.go) independently numbers ITS arms
// "Strategy 0..6" — two different rules under the same names, in two files that
// call into each other, and migration 402's header already cited the wrong
// file's arm because of it. The arms here are: direct-path, input-data-prefix,
// input-data-map, whole-tree-search, alias. Do not reintroduce numbers.
//
// THERE IS A BUDGET ON THIS CHAIN TOO, enforced by a test, not by this comment:
// resolver_arm_budget_test.go counts this function's resolution return sites
// (floor 5, ceiling 8). Past the ceiling, another arm is an RFC — RFC_028's
// outer-chain rule, extended to the inner chain by RFC_029 §9 D4.
func extractSingleField(
	data map[string]interface{},
	fieldName string,
	seen map[string]bool,
	logger *zap.Logger,
) interface{} {

	logger.Info("Trying extraction strategies", zap.String("field", fieldName))

	// Prevent infinite loops from circular aliases
	if seen[fieldName] {
		return nil // Already tried this field
	}
	seen[fieldName] = true

	// direct-path: FindByPath (handles unwrapping)
	if value := FindByPath(data, fieldName, logger); value != nil {
		logger.Info("Found via FindByPath", zap.String("field", fieldName))
		return value
	}

	// input-data-prefix: retry with "input_data." prepended
	if !strings.HasPrefix(fieldName, "input_data.") {
		path := "input_data." + fieldName
		if value := FindByPath(data, path, logger); value != nil {
			logger.Info("Found via input_data prefix",
				zap.String("field", fieldName),
				zap.String("path", path))
			return value
		}
	}

	// input-data-map: look inside the input_data map directly
	if inputMap := getInputDataMap(data, logger); inputMap != nil {
		if value, ok := inputMap[fieldName]; ok {
			logger.Info("Found in input_data map", zap.String("field", fieldName))
			return UnwrapDeep(value, logger)
		}
	}

	// whole-tree-search: the aggressive last resort — collect-all /
	// unique-or-nothing since RFC_029 §9 (see findFieldRecursive)
	logger.Info("Trying aggressive search", zap.String("field", fieldName))
	if value := findFieldRecursive(data, fieldName, 0, logger); value != nil {
		logger.Info("Found via aggressive search", zap.String("field", fieldName))
		return value
	}

	// alias: retry the whole chain under a known alternative name
	if alias := getFieldAlias(fieldName); alias != "" {
		logger.Info("Trying field alias",
			zap.String("field", fieldName),
			zap.String("alias", alias))
		return extractSingleField(data, alias, seen, logger)
	}

	logger.Info("Field name not found using aggressive search in extractSingleField.",
		zap.String("field", fieldName),
	)

	return nil
}

// fieldCandidate is one match the whole-tree search collected for a field name.
type fieldCandidate struct {
	path  string // dotted path from the search root; "~unwrap" marks a hop through tryUnwrapMapPatterns
	depth int
	value interface{} // UnwrapDeep'd, never nil

	// rank is the DECLARED tie-break at equal depth (bugs_open/306). Lower wins.
	// It names the preference the resolver has always applied by accident of
	// append order: a direct key match at this level, then the ~unwrap hop
	// (input_data / result / *_result — the places a step's own inputs live),
	// then plain sibling recursion. Measured 2026-08-18 on page-build-handler,
	// 13 of 139 runs carried a genuinely different page at equal depth and the
	// unwrap-hop candidate was the correct one every time; nothing declared that
	// dependency, and a reordered append would have flipped it with no failing
	// test. Now the sort reads the rank, and a test pins it.
	rank candidateRank
}

// candidateRank orders equal-depth candidates. The numeric values are the
// collector's historical append order, so declaring them changes no winner.
type candidateRank int

const (
	rankDirectMatch candidateRank = iota // v[fieldName] at this level
	rankUnwrapHop                        // reached via tryUnwrapMapPatterns
	rankSibling                          // reached by recursing into another key
)

// findFieldRecursive is the whole-tree-search arm — the resolver's last resort,
// and the only nondeterministic one until 2026-08-15.
//
// RULING (RFC_029 §9, owner-delegated, 2026-08-15): COLLECT-ALL / UNIQUE-OR-NOTHING.
// The search collects EVERY match in the tree (same depth cap, same
// infrastructure-key skip list as always) instead of returning whichever match a
// randomised map iteration happened to meet first. If every match carries the
// same value, it resolves — deterministically, shallowest path first — and
// behaviour is unchanged from the old happy path. If the candidates CONFLICT,
// any picking rule is a guess, and the owner's stated ranking forbids a guess:
// no field at all is better than a wrong field.
//
// PHASE 2 IS THIS BUILD (flipped 2026-08-21, RFC_029 §10.13 step 5): a conflict
// resolves to NOTHING. The WARN below still fires and still names the field, every
// candidate path and the candidate the ranking WOULD have picked — the instrument
// is unchanged apart from `phase`, which now reads "2-refuse".
//
// PHASE 1 (builds before 2026-08-21) resolved conflicts to the stable
// shallowest-first winner and warned. If you are reading a conflict row, check its
// `phase` before assuming which behaviour produced it.
//
// The ranking below is therefore NOT dead code: it still decides what
// `winner_path` reports, which is the first thing anyone tracing an absent field
// wants to know. bugs_closed/306 is why it is declared rather than accidental.
func findFieldRecursive(
	data interface{},
	fieldName string,
	depth int,
	logger *zap.Logger,
) interface{} {
	candidates := collectFieldCandidates(data, fieldName, "", depth, nil, logger)
	if len(candidates) == 0 {
		return nil
	}

	// Stable sort by depth, then by the DECLARED rank (bugs_open/306): direct
	// match, then the ~unwrap hop, then sibling recursion. Remaining ties keep
	// the collector's deterministic DFS encounter order (sorted map keys, natural
	// slice order) — sorting by path STRING here would misorder slice indices
	// ("[10]" < "[2]" lexicographically). The rank was always the append order;
	// reading it here is what stops a reordered collector from silently changing
	// the winner in the 13-of-139 population where the candidates are different
	// pages.
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].depth != candidates[j].depth {
			return candidates[i].depth < candidates[j].depth
		}
		return candidates[i].rank < candidates[j].rank
	})

	winner := candidates[0]
	conflicting := false
	for _, c := range candidates[1:] {
		if !reflect.DeepEqual(c.value, winner.value) {
			conflicting = true
			break
		}
	}

	if conflicting {
		paths := make([]string, len(candidates))
		for i, c := range candidates {
			paths[i] = c.path
		}
		// THE PHASE 2 INSTRUMENT (RFC_029 §9 D2). Same message, same fields, so
		// the observation window's queries keep working across the flip — but
		// `phase` now reads "2-refuse", which is how a reader tells WHICH BUILD
		// produced a row without guessing from its timestamp. `winner_path` is
		// still reported: nothing is resolved from it any more, but it names the
		// candidate the old ranking WOULD have picked, which is the single most
		// useful thing to know when tracing a field that has stopped arriving.
		logger.Warn("aggressive search: conflicting candidates",
			zap.String("field", fieldName),
			zap.Strings("candidate_paths", paths),
			zap.String("winner_path", winner.path),
			zap.String("phase", "2-refuse"),
		)
		// ...and PERSISTED, every occurrence (resolver_findings.go): the log
		// line above scrolls out of a chassis pod in ~90s, and the window is
		// 48h+. With no recorder registered this is a no-op — log-only.
		recordResolverFinding(logger, ResolverFinding{
			Code:    ResolverFindingConflictingCandidates,
			Field:   fieldName,
			Message: "aggressive search: conflicting candidates",
			Context: map[string]interface{}{
				"field":           fieldName,
				"candidate_paths": paths,
				"winner_path":     winner.path,
				"phase":           "2-refuse",
			},
		})

		// PHASE 2, THE FLIP (RFC_029 §9 D2, owner-delegated ruling 2026-08-15;
		// path ruled 2026-08-18, sequence A). Conflicting candidates resolve to
		// NOTHING. Any picking rule here is a guess, and the owner's stated
		// ranking forbids a guess: no field at all is better than a wrong field.
		//
		// The precondition was NOT the "zero WARNs" branch — that branch can
		// never be sufficient, because a row requires the candidates to DIFFER
		// and a lone wrong candidate substitutes silently (bugs_open/330 §4,
		// bugs_open/350). It was the OTHER branch: every observed field/caller
		// pair given an explicit mapping first. All 19 pairs the instrument saw
		// between 2026-08-16 and 2026-08-21 are dispositioned — 11 closed by the
		// prune/gate/rename, 2 by their own existing wires, 3 recorded as
		// "absence is correct", and the three live ones fixed and PROVEN at the
		// artefact: pbh/page_type (mig 515), tg/reason (mig 512), and
		// bdl/commit_sha (mig 537). The record is
		// docs024_key_docs_latest/staged_component_build/HANDOFF_2026-08-20_continue_here.md
		// §2.4–§2.9.
		//
		// A caller that now gets nil where it used to get a value is a caller
		// that was being handed a guess. If that breaks something, the fix is an
		// explicit mapping at the step (`<field>?: <path>` — the OPTIONAL-EXPLICIT
		// marker, first proven live by mig 515), never a return to picking.
		return nil
	}

	logger.Debug("whole-tree-search resolved",
		zap.String("field", fieldName),
		zap.String("path", winner.path),
		zap.Int("candidates", len(candidates)),
	)
	return winner.value
}

// collectFieldCandidates walks data depth-first with SORTED keys, collecting
// every non-nil value stored under fieldName — the collect-all half of
// findFieldRecursive. Depth cap and the infrastructure-key skip list are exactly
// the old walk's; only the traversal ORDER (sorted, so reproducible) and the
// stopping rule (never stops early) differ.
//
// A key holding null was never a resolvable value here — the old walk returned
// it as nil at the root (which read as not-found) and skipped past it mid-tree —
// so the collector skips null uniformly rather than treating it as a candidate.
//
// RANK (bugs_open/306): every candidate carries the rank of the FIRST hop that
// left the search root — direct match, ~unwrap hop, or sibling recursion — and
// that rank is inherited unchanged by everything found beneath it. A caller at
// the root passes rankDirectMatch; only the root's three branches assign the
// other two. This is the append order the collector has always used, made
// explicit so the sort in findFieldRecursive can read it instead of relying on it.
func collectFieldCandidates(
	data interface{},
	fieldName string,
	path string,
	depth int,
	out []fieldCandidate,
	logger *zap.Logger,
) []fieldCandidate {
	return collectFieldCandidatesRanked(data, fieldName, path, depth, rankDirectMatch, path == "", out, logger)
}

func collectFieldCandidatesRanked(
	data interface{},
	fieldName string,
	path string,
	depth int,
	inherited candidateRank,
	atRoot bool,
	out []fieldCandidate,
	logger *zap.Logger,
) []fieldCandidate {
	if depth > 20 {
		return out
	}

	// Below the root every hop inherits; at the root each branch names its own.
	rankFor := func(branch candidateRank) candidateRank {
		if atRoot {
			return branch
		}
		return inherited
	}

	switch v := data.(type) {
	case map[string]interface{}:
		// Direct match at this level
		if val, ok := v[fieldName]; ok {
			if unwrapped := UnwrapDeep(val, logger); unwrapped != nil {
				out = append(out, fieldCandidate{
					path:  joinCandidatePath(path, fieldName),
					depth: depth,
					value: unwrapped,
					rank:  rankFor(rankDirectMatch),
				})
			}
		}

		// Unwrap common nesting patterns ({X}_result.result, result, input_data)
		// and search inside — this reaches JSON-encoded payloads the plain key
		// walk cannot (a string is not a map), exactly as the old walk did.
		if unwrapped := tryUnwrapMapPatterns(v, logger); unwrapped != nil {
			out = collectFieldCandidatesRanked(unwrapped, fieldName, joinCandidatePath(path, "~unwrap"), depth+1, rankFor(rankUnwrapHop), false, out, logger)
		}

		// Recurse into all values, in sorted-key order so the candidate list is
		// reproducible. Skip infrastructure/metadata blobs that contain workflow
		// configs with literal path strings (e.g. input_mapping:
		// {"site_id": "site_record.site_id"}) which get confused with actual data
		// values. Direct match above still works if someone explicitly asks for
		// "agent_config" etc.
		keys := make([]string, 0, len(v))
		for key := range v {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			if isInfrastructureKey(key) {
				continue
			}
			if key == fieldName {
				// Already collected as the direct match above. A same-named key
				// nested INSIDE that value is not an independent source, and the
				// old first-match walk could never have returned it — descending
				// would manufacture conflicts no run has ever seen.
				continue
			}
			out = collectFieldCandidatesRanked(v[key], fieldName, joinCandidatePath(path, key), depth+1, rankFor(rankSibling), false, out, logger)
		}

	case []interface{}:
		for i, val := range v {
			out = collectFieldCandidatesRanked(val, fieldName, fmt.Sprintf("%s[%d]", path, i), depth+1, rankFor(rankSibling), false, out, logger)
		}
	}

	return out
}

// joinCandidatePath extends a candidate path with one segment.
func joinCandidatePath(prefix, key string) string {
	if prefix == "" {
		return key
	}
	return prefix + "." + key
}

// isInfrastructureKey returns true for keys that contain framework metadata
// rather than business data. These should not be recursed into during
// aggressive search because they contain workflow configs with literal
// path strings (e.g. input_mapping: {"site_id": "site_record.site_id"})
// that can be mistaken for actual data values.
//
// See also: current_section extraction (line ~159) which already uses a
// safe-key whitelist for the same reason.
func isInfrastructureKey(key string) bool {
	switch key {
	case "agent_config",
		"__raw_message__",
		"__work_request__",
		"__execution_context__",
		"__my_requests_topic__",
		"__my_responses_topic__",
		"__parent_responses_topic__",
		"__reply_to_request_id__":
		return true
	case types.RetryPayloadKey:
		// bugs_open/306 candidate 3. A retry_payload subtree is the VERBATIM
		// Kafka message an awaited-request action produced, kept so the
		// coordinator can replay it on timeout (bugs_open/129) — including a
		// frozen copy of every input field the sender sent, which is exactly
		// the stale echo behind the genuinely-different-page candidate sets
		// (13/139 measured 2026-08-18). Its ONLY reader takes it by direct key
		// from the action result before that result reaches CollectedData
		// (coordinator.go extractRetryPayload → awaited_requests.request_payload);
		// nothing resolves fields OUT of it, so the search must not treat its
		// contents as live data. Like agent_config above, a direct request FOR
		// the key itself still resolves — this guards recursion only.
		return true
	}
	return false
}

// tryUnwrapMapPatterns tries to unwrap common nesting patterns
func tryUnwrapMapPatterns(m map[string]interface{}, logger *zap.Logger) interface{} {
	// Pattern 1: {field}_result.result — in SORTED key order (bugs_open/306).
	// This loop used to range the map directly and return the first *_result it
	// met, which is the exact iteration-order coin flip RFC_029's collect-all
	// rewrite removed from the collector, surviving one call inside it. Measured
	// 0/139 able to fire on 2026-08-18 (no live root *_result object carried a
	// `result` child), so sorting changes no live winner today — which is the
	// cheap moment to close it, before a workflow produces two such keys.
	keys := make([]string, 0, len(m))
	for key := range m {
		if strings.HasSuffix(key, "_result") {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	for _, key := range keys {
		val := m[key]
		{
			if resultMap, ok := val.(map[string]interface{}); ok {
				if result, ok := resultMap["result"]; ok {
					if parsed := tryParseJSON(result, logger); parsed != nil {
						return parsed
					}
					return result
				}
			}
		}
	}

	// Pattern 2: Direct result field
	if result, ok := m["result"]; ok {
		if parsed := tryParseJSON(result, logger); parsed != nil {
			return parsed
		}
		return result
	}

	// Pattern 3: input_data field
	if inputData, ok := m["input_data"]; ok {
		return inputData
	}

	return nil
}

// ensureCoreFields makes absolutely sure critical fields exist eg domain, objective
// These are fields that templates commonly reference but might not be explicitly requested
//
// TWO CLASSES OF TOP-UP, gated differently (RFC_029 §10.13 step 3, owner ruling
// 2026-08-18; the measurement is in staged_component_build NOTES 2026-08-18
// evening and bugs_open/306):
//
//   - domain / objective / model are recovered UNCONDITIONALLY, as always.
//     Measured 2026-08-18: 39 / 10 / 6 live steps across 31 agent types read
//     them from templates WITHOUT declaring them. They are load-bearing and
//     must never be gated.
//
//   - current_page / current_section / render_context are recovered ONLY WHEN
//     REQUESTED (present in fieldNames). Measured the same day: zero undeclared
//     template consumers fleet-wide, and exactly one Go-side consumer that read
//     the injected copy rather than requesting it — html-developer-chunked —
//     which migration 483 gave an explicit input_fields list BEFORE this gate
//     shipped (order is load-bearing: config is live at once, this rides a roll).
//     Before the gate, 63% of the RFC_029 observation window's conflict rows
//     were this function searching for a page on behalf of build-dispatch-loop,
//     which declares no page anywhere and reads the answer nowhere: a wrong
//     value, found by a whole-tree walk per step, delivered into a slot nobody
//     opens. A request list is the only honest statement of who wants what.
//
// A requested page-ish field still gets the special-case extraction in
// ExtractFields (which runs first and usually fills it) and this fallback if
// that came up empty — exactly as before. Only the UNREQUESTED path is closed.
func ensureCoreFields(
	result map[string]interface{},
	source map[string]interface{},
	fieldNames []string,
	logger *zap.Logger,
) {
	requested := func(name string) bool { return contains(fieldNames, name) }
	// Name the fields the gate is about to skip ON the line this function has
	// always emitted, so "why is current_page missing here?" is answerable from
	// the log of the very call that skipped it — the alternative (silence) is
	// the no-error-no-warning shape this estate keeps re-learning about
	// (council REVISE round on this change, bug_historian seat, 2026-08-19).
	// Empty when everything page-ish was requested or nothing was.
	var skippedPageFields []string
	for _, f := range []string{"current_page", "current_section", "render_context"} {
		if !requested(f) {
			skippedPageFields = append(skippedPageFields, f)
		}
	}
	logger.Info("Ensuring core fields present",
		zap.Strings("unrequested_page_fields_not_injected", skippedPageFields))

	// Check domain
	if _, hasDomain := result["domain"]; !hasDomain {
		logger.Warn("Domain missing from result, searching aggressively")
		if domain := FindDomainAggressive(source, logger); domain != "" {
			result["domain"] = domain
			logger.Info("✓ Recovered domain via aggressive search", zap.String("domain", domain))
		} else {
			// Try site_record.domain explicitly
			if siteRecord := findFieldRecursive(source, "site_record", 0, logger); siteRecord != nil {
				if srMap, ok := siteRecord.(map[string]interface{}); ok {
					if d, ok := srMap["domain"].(string); ok && d != "" {
						result["domain"] = d
						logger.Info("✓ Found domain in site_record.domain", zap.String("domain", d))
					}
				}
			}
		}
	}

	// Check objective
	if _, hasObjective := result["objective"]; !hasObjective {
		logger.Warn("Objective missing from result, searching aggressively")
		if objective := FindObjectiveAggressive(source, logger); objective != "" {
			result["objective"] = objective
			logger.Info("✓ Recovered objective via aggressive search",
				zap.Int("length", len(objective)))
		} else {
			// Build objective from context if we can't find one
			objective := buildObjectiveFromContext(result, source, logger)
			if objective != "" {
				result["objective"] = objective
				logger.Info("✓ Built objective from context", zap.String("objective", objective))
			} else {
				logger.Warn("✗ Could not find or build objective")
			}
		}
	}

	// Check model
	if _, hasModel := result["model"]; !hasModel {
		if model := findFieldRecursive(source, "model", 0, logger); model != nil {
			if modelStr, ok := model.(string); ok {
				result["model"] = modelStr
				logger.Info("✓ Recovered model via aggressive search", zap.String("model", modelStr))
			}
		}
	}

	// Check current_page - only when REQUESTED (see the doc comment above)
	if _, hasCurrentPage := result["current_page"]; !hasCurrentPage && requested("current_page") {
		if currentPage := findFieldRecursive(source, "current_page", 0, logger); currentPage != nil {
			result["current_page"] = currentPage
			if cpMap, ok := currentPage.(map[string]interface{}); ok {
				logger.Info("✓ Auto-recovered current_page",
					zap.String("name", fmt.Sprintf("%v", cpMap["name"])),
					zap.String("title", fmt.Sprintf("%v", cpMap["title"])),
				)
			} else {
				logger.Info("✓ Auto-recovered current_page (non-map)")
			}
		}
	}

	// Check current_section - only when REQUESTED
	if _, hasCurrentSection := result["current_section"]; !hasCurrentSection && requested("current_section") {
		if currentSection := findFieldRecursive(source, "current_section", 0, logger); currentSection != nil {
			result["current_section"] = currentSection
			logger.Info("✓ Auto-recovered current_section")
		}
	}

	// Check render_context - only when REQUESTED
	if _, hasRenderContext := result["render_context"]; !hasRenderContext && requested("render_context") {
		if renderContext := findFieldRecursive(source, "render_context", 0, logger); renderContext != nil {
			result["render_context"] = renderContext
			logger.Info("✓ Auto-recovered render_context")
		}
	}
}

// buildObjectiveFromContext constructs an objective string from available context
// Used when no explicit objective is found but we have enough context to build one
func buildObjectiveFromContext(result, source map[string]interface{}, logger *zap.Logger) string {
	var parts []string

	// Get section info
	sectionName := ""
	if cs, ok := result["current_section"].(map[string]interface{}); ok {
		if fn, ok := cs["function"].(string); ok && fn != "" {
			sectionName = fn
		} else if n, ok := cs["name"].(string); ok && n != "" {
			sectionName = n
		}
	}

	// Get company name
	companyName := ""
	if rb, ok := result["reviewed_brief"].(map[string]interface{}); ok {
		if cn, ok := rb["company_name"].(string); ok {
			companyName = cn
		}
	}

	// Get domain
	domain := ""
	if d, ok := result["domain"].(string); ok {
		domain = d
	}

	// Build objective based on what we have
	if sectionName != "" {
		parts = append(parts, fmt.Sprintf("Generate content for %s section", sectionName))
	}
	if companyName != "" {
		parts = append(parts, fmt.Sprintf("for %s", companyName))
	}
	if domain != "" && companyName == "" {
		parts = append(parts, fmt.Sprintf("for %s", domain))
	}

	if len(parts) > 0 {
		return strings.Join(parts, " ")
	}

	return ""
}

// getInputDataMap extracts the input_data map if it exists
func getInputDataMap(data map[string]interface{}, logger *zap.Logger) map[string]interface{} {
	// Try direct lookup
	if inputData, ok := data["input_data"].(map[string]interface{}); ok {
		// Check for double nesting
		if nested, ok := inputData["input_data"].(map[string]interface{}); ok {
			logger.Debug("Unwrapped double-nested input_data")
			return nested
		}
		return inputData
	}

	// Try unwrapping data first
	unwrapped := UnwrapDeep(data, logger)
	if unwrappedMap, ok := unwrapped.(map[string]interface{}); ok {
		if inputData, ok := unwrappedMap["input_data"].(map[string]interface{}); ok {
			return inputData
		}
	}

	return nil
}

// getFieldAlias returns alternative names for common fields
// NOTE: Do NOT alias site_content<->content because "content" is too common
// (appears in sections[0].content, etc.) and causes wrong field extraction
func getFieldAlias(fieldName string) string {
	aliases := map[string]string{
		"site_architecture":  "architecture",
		"architecture":       "site_architecture",
		"domain_analysis":    "analysis",
		"analysis":           "domain_analysis",
		"available_builders": "builders",
		// Removed: "site_content": "content" - too dangerous, matches sections[0].content
		// Removed: "content": "site_content"
	}
	return aliases[fieldName]
}

// Helper functions
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
