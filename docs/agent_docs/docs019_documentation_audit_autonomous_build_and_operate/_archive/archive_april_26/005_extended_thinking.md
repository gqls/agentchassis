// ============================================================================
// Workflow config example with extended thinking:
//
//   "classify_and_extract": {
//       "action": "execute_llm_prompt",
//       "config": {
//           "ai_service": {
//               "model": "claude-opus-4-6",
//               "provider": "anthropic",
//               "max_tokens": 4000,
//               "api_key_env_var": "ANTHROPIC_API_KEY",
//               "budget_tokens": 10000
//           },
//           "prompt_template": "..."
//       }
//   }
//
// When budget_tokens is set:
// - AnthropicClient adds {"thinking": {"type": "enabled", "budget_tokens": N}}
// - Temperature is removed (Anthropic requirement)
// - Response parsing skips thinking blocks and returns only the text block
// - Response time increases by 30-90 seconds
//
// When budget_tokens is absent or 0:
// - Standard behaviour, no thinking, temperature applied as normal
// ============================================================================