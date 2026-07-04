// PATCH — platform/orchestration/input_contracts/input_mapping.go
// Function: ValidateInputContract
//
// WHY
// ---
// call_agent (extractDataForAgent) resolves input_mapping into `inputData`, then
// (a) validates it against the target agent's input_contract and (b) returns the
// SAME map as the child's input_data. Handlers driven by the dispatch loop read
// spec-derived fields from input_data.spec.* (doc 003 §984-995: "reading the nested
// path is the one contract-compliant way"). But ValidateInputContract only checks
// the TOP LEVEL of inputData, so a required field that (correctly) arrives inside the
// spec object is reported missing.
//
// This is exactly why component-creator (input_contract.required = ["section_type"])
// cannot be dispatched through build-dispatch-loop: the loop maps the whole spec to
// inputData["spec"] but never flattens section_type to the top level, so validation
// fails with `missing required fields: [section_type]` before the agent runs. The
// generic loop deliberately does NOT carry per-handler field knowledge (doc 002
// §414), so the correct place to fix this is the validator, not the loop mapping.
//
// FIX
// ---
// When a required field is absent at the top level, also accept it if it is present
// inside inputData["spec"] — the documented location handlers consume it from. This
// is strictly more permissive and only in a principled direction: a field that is
// genuinely absent (neither top-level nor in spec) still fails loudly, preserving the
// hard-fail contract. No agent-definition or schema change is required; no change to
// the generic dispatch loop.
//
// SCOPE / SAFETY
// --------------
// - Backward compatible: agents called with flat top-level fields are unaffected
//   (top-level check still runs first). No agent relies on the validator REJECTING a
//   spec-nested field.
// - "spec" is the one documented container for handler inputs (input_data.spec.*), so
//   this checks the same place the child reads — not an arbitrary hunt. We do NOT also
//   scan current_page or other aliases: keep the rule to the documented convention.
// - The spec value is type-asserted to map[string]interface{}; a non-map spec is
//   ignored safely.
//
// Replace the existing ValidateInputContract with the version below.
// (logger.Info is used for the spec-satisfied note because Debug is not surfaced in
//  our log pipeline.)

func ValidateInputContract(
	agentType string,
	data map[string]interface{},
	contract *InputContract,
	logger *zap.Logger,
) error {
	if contract == nil {
		logger.Debug("No input contract defined, skipping validation",
			zap.String("agent_type", agentType))
		return nil
	}

	// spec is the documented container for handler inputs (input_data.spec.*).
	// A required field may be delivered here rather than flattened top-level.
	spec, _ := data["spec"].(map[string]interface{})

	var missing []string
	var satisfiedViaSpec []string

	for _, required := range contract.Required {
		// 1) top-level (the historical check)
		if _, exists := data[required]; exists {
			continue
		}
		// 2) input_data.spec.* — the path handlers actually read (doc 003).
		if spec != nil {
			if _, inSpec := spec[required]; inSpec {
				satisfiedViaSpec = append(satisfiedViaSpec, required)
				continue
			}
		}
		missing = append(missing, required)
	}

	if len(missing) > 0 {
		providedFields := MapKeys(data)
		var specKeys []string
		if spec != nil {
			specKeys = MapKeys(spec)
		}
		return fmt.Errorf(
			"contract violation for agent '%s': missing required fields: %v\n"+
				"Provided fields: %v\n"+
				"Provided spec.* fields: %v\n"+
				"Hint: required fields must be present top-level or under input_data.spec.*; check input_mapping in the step config",
			agentType, missing, providedFields, specKeys,
		)
	}

	if len(satisfiedViaSpec) > 0 {
		logger.Info("Input contract satisfied (some fields via input_data.spec.*)",
			zap.String("agent_type", agentType),
			zap.Strings("via_spec", satisfiedViaSpec))
	}

	logger.Debug("Input contract validated successfully",
		zap.String("agent_type", agentType),
		zap.Int("required_count", len(contract.Required)),
		zap.Int("provided_count", len(data)))

	return nil
}
