package actions

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
)

// ScopeToolBirthTemplate is the FLOW half of bugs_open/283. The conversion
// programme fixed every template that already existed (the stock); this guard
// sits at the only seam every NEW tool passes through — create_tool_component's
// write — so the estate stops minting the defect it just finished converting
// (WRONG_CALLS 2026-08-21: 23 old-style tools were born during the conversion
// programme itself, seven of them on the day the backlog finished).
//
// It runs the SAME deterministic transform and acceptance gate the programme
// proved on 24 live components (ConvertTemplateToInstanceScope +
// GateConvertedTemplate — the latter executes the doubled template through the
// real renderer), rather than trusting the generator's LLM to follow a prompt:
// the prompt reduces the failure rate, this guard bounds it.
//
// Two modes, keyed by the step's enforce_instance_scope config — the
// 2026-08-02 §2 shape (new authority on a shared seam ships opt-in, unsafe
// default OFF):
//
//	unarmed: record-only. The returned bytes are the caller's input, verbatim;
//	         info names what WOULD have happened. No new authority.
//	armed:   convert-and-refuse. A mechanically convertible template is
//	         CONVERTED (ids prefixed, bindings renamed) and gated; a template
//	         the deterministic pass cannot prove safe (global-scope script,
//	         dynamic unprefixable ids, dangling bindings, hex-ambiguous id) is
//	         REFUSED. At birth a refusal is cheap and safe in a way it is not
//	         post-hoc: the step fails, the work item retries, and the
//	         generator produces a fresh sample against a prompt that teaches
//	         the convention — no live page is waiting on the outcome.
//
// Returns:
//
//	tpl      — bytes to persist as content_components.html_template
//	rendered — bytes for page_components.rendered_html writes. Tools carry
//	           the template VERBATIM as rendered_html (the assembler never
//	           re-renders a tool slot from its template), so this is the
//	           template with {{.InstanceID}} BOUND for occurrence 0 — an
//	           unbound placeholder here would ship literal "{{.InstanceID}}-"
//	           text inside live ids.
//	info     — result/log fields; always non-nil.
//	refuse   — non-nil ⇒ persist NOTHING (armed mode only).
func ScopeToolBirthTemplate(html, function string, armed bool, logger *zap.Logger) (tpl, rendered string, info map[string]interface{}, refuse error) {
	info = map[string]interface{}{"instance_scope_armed": armed}
	converted, rep, ok := ConvertTemplateToInstanceScope(html)

	// gateAndBind proves a placeholder-bearing candidate and produces the
	// occurrence-0 rendering for the verbatim rendered_html writes.
	gateAndBind := func(candidate string) (string, error) {
		needsJudged, gateErr := GateConvertedTemplate(function, candidate, logger)
		if gateErr != nil {
			return "", fmt.Errorf("instance-scope gate error: %w", gateErr)
		}
		if needsJudged {
			ub := UnprefixedBindings(candidate)
			return "", fmt.Errorf("script is not mechanically provable — it declares into global scope and/or %d binding(s) would dangle (%s); regenerate following the scoping rules the prompt states (IIFE-wrapped script, addEventListener not inline on*=, no window.onload, no dynamic ids without a static prefix)",
				len(ub), strings.Join(ub, "; "))
		}
		rc := &RenderContext{}
		BindInstanceToken(rc, InstanceToken(function, 0))
		bound, _, _, rerr := RenderTemplate(candidate, rc, logger)
		if rerr != nil {
			return "", fmt.Errorf("instance-scope occurrence-0 bind failed: %w", rerr)
		}
		return bound, nil
	}

	switch {
	case rep.AlreadyConverted:
		// The generator emitted the convention itself (the prompt teaches it).
		// Verify, never trust: a self-converted template that fails the gate is
		// exactly as broken as an unconverted one, wearing the fix's clothes.
		bound, err := gateAndBind(html)
		if err != nil {
			info["instance_scope"] = "generator emitted {{.InstanceID}} but the template fails the gate"
			info["instance_scope_defect"] = err.Error()
			if armed {
				return "", "", info, fmt.Errorf("tool birth refused (instance scope, self-converted template): %w", err)
			}
			return html, html, info, nil
		}
		info["instance_scope"] = "generator emitted the convention; gated clean"
		return html, bound, info, nil

	case ok:
		bound, err := gateAndBind(converted)
		if err != nil {
			info["instance_scope"] = "mechanical conversion produced an ungateable template"
			info["instance_scope_defect"] = err.Error()
			if armed {
				return "", "", info, fmt.Errorf("tool birth refused (instance scope): %w", err)
			}
			return html, html, info, nil
		}
		info["instance_scope_ids"] = len(rep.IDsDeclared)
		info["instance_scope_binding_literals"] = rep.Bindings.LiteralIDsRenamed
		info["instance_scope_concat_sites"] = rep.Bindings.ConcatSitesRenamed
		if rep.TemplatedIDSwaps > 0 {
			// A newborn spelling the RETIRED placeholder (RFC_032). Reported
			// because it is the one conversion that says something about the
			// GENERATOR — the prompt taught {{.InstanceID}} and this sample
			// used the old name — and a count nobody surfaces is a count
			// nobody can act on.
			info["instance_scope_templated_id_swaps"] = rep.TemplatedIDSwaps
		}
		if !armed {
			info["instance_scope"] = "record-only: mechanically convertible; arming enforce_instance_scope on this step would convert it at birth"
			return html, html, info, nil
		}
		info["instance_scope"] = "mechanically converted at birth"
		return converted, bound, info, nil

	case rep.NoLiteralElementIDs && !rep.ComponentIDUnswappable:
		// Nothing to scope and nothing to collide on — inert-safe both modes.
		//
		// ⚠ TWO CONDITIONS, and the second is load-bearing: a template that
		// declares no literal ids but carries {{.ComponentID}} somewhere pass 0
		// cannot rewrite reaches the same early return in the converter, and it
		// is NOT inert — that placeholder renders the same value on every
		// instance, which is the collision this seam exists to remove. It
		// belongs in the default arm below (refused when armed).
		//
		// Both are TYPED FIELDS rather than a substring of RefusedReason, which
		// is what this arm used to key on. That text is prose: it gets reworded,
		// and a future refusal whose wording happens to contain the old phrase
		// would silently inherit this pass-through — the failure mode that put
		// the ComponentID case here in the first place (council cd6a5ef6 r2,
		// bug_historian; LANDMINES 2026-08-23). Route on facts, not sentences.
		info["instance_scope"] = "no literal ids — nothing to scope"
		return html, html, info, nil

	default:
		// The judged class at birth: hex-ambiguous id, dynamic id with no
		// static prefix, unrecognised binding construction. Post-hoc these
		// route to an LLM+human pipeline because the live behaviour must be
		// preserved; at birth there is nothing to preserve — regenerating is
		// strictly cheaper and safer than judging.
		info["instance_scope"] = "not mechanically convertible"
		info["instance_scope_defect"] = rep.RefusedReason
		if armed {
			return "", "", info, fmt.Errorf("tool birth refused (instance scope): %s — regenerate following the scoping rules the prompt states", rep.RefusedReason)
		}
		return html, html, info, nil
	}
}
