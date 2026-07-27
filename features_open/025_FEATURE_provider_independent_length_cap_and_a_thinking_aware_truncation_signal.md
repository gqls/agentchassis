# 025 — A provider-independent hard length cap, and a thinking-aware truncation signal

**Filed** 2026-07-27 by the `gemini_content_provider` workstream · **Source:** the
council gate's own recommendations on corr `a1a5cf20-a70d-48c3-8fda-842d2a91b651`
(APPROVED, 4 advisory objections) — raised independently by the **edit-quality**,
**guardian**, **guidelines** and **llm-reliability** seats · **Blocks nothing**

---

## Why this exists

`bugs_open/107` changed the Gemini client so `maxOutputTokens` = the caller's
`max_tokens` + a thinking reserve. That fixes the starvation, and it has two
consequences the council flagged and I did not fix in that change. Both are
recorded here rather than left in a code comment, because both are
**provider-independent problems that the Gemini work merely made visible**.

## (a) `max_tokens` is not a hard length cap, and something wants one

For a thinking model the API ceiling can no longer bound visible output — the
reserve is part of the same ceiling. `3ea9d718c` describes content-creator's
twitter tier this way (verbatim, line breaks as the commit wraps them):

```
2000->3000, long 4000->6000. twitter's 100 is untouched - that's an
intentional platform-side limit tied to the platform's character count,
not a model-verbosity ceiling, and wasn't observed truncating on Anthropic.
```

So there **is** a caller that means `max_tokens` as a real limit, and it means it
in **characters** — a platform constraint, not a model one. Guardian's phrasing:
*"this repurposes max_tokens from a hard ceiling into 'best-effort, thinking model
may ignore for length' on the Gemini path"*.

**Why 107 did not fix it, stated so this is a decision and not an oversight.**
A clamp in the provider client would truncate the returned text, and
`page-content-writer` sets `output_format: json` — cutting a JSON section
mid-string yields an unparseable artifact, which is strictly worse than a long
one. And a *character* limit tied to a publishing platform does not belong in an
LLM transport at all: Anthropic has the same exposure the moment extended
thinking is enabled on that path, so a Gemini-side clamp would be the wrong fix
in the wrong layer.

**What to build instead.** A cap where the constraint actually lives — in
`content-creator`, on the tweet text, in characters, applied to every provider.
The guidelines seat's option (a). Scope it by first answering the question
guardian asked and I only partly answered: **which callers enforce their own
truncation after the call, and which rely on `max_tokens`?** I confirmed
`estimateCost`'s blast radius (one call site, `agent.go:299`) but not this.

## (b) `llm_call_log`'s truncation heuristic goes quiet for thinking models

The standing rule — `output_tokens == max_tokens` means the completion was CUT —
compares visible output against a total that now includes the reserve, so it will
**silently under-report truncation** for Gemini. Raised by llm-reliability, which
also noted the mitigation is real: the client returns a typed `*TruncatedError` on
`finishReason=MAX_TOKENS`, and `__usage_thinking_tokens` /
`__usage_total_tokens` / `__sent_visible_budget_tokens` are all recorded, so no
data is lost — only the arithmetic shortcut breaks.

Its actual point is about **ownership**: *"This should be called out to whoever
owns that heuristic/dashboard, not just noted in code comments."* That is what
this file is for. Two options:
1. Teach the heuristic the split — compare `usage_output_tokens` against
   `sent_visible_budget_tokens` when it is present, falling back to
   `sent_max_tokens`. Small, and keeps every existing query working.
2. Retire the heuristic in favour of the typed error, which is authoritative and
   provider-independent. Bigger, and needs whoever reads the dashboards.

Option 1 first. Note the rule is quoted in `CLAUDE.md` and in `016b`, so both need
a line if the semantics change — a rule that silently means something different
per provider is worse than either version of it.

## Not in scope

Whether Gemini writes acceptably. That is `gemini_content_provider`'s P5/P6 and
needs a real generation, not a feature.
