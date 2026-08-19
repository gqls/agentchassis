# CONTRIB to `copy_quality_two_stage` — the shrink floor on `pricing` is defending literal markdown

**From:** `site_ai_agent_orchestration` lane, 2026-08-19 (work done 2026-08-18 evening).
**Why you:** it is on `ai-agent-orchestration.com`, the site your two owner-directed asks are
about, and it sits exactly on the seam between your stage-2 output and the floors that decide
whether a copy change may land.

## The finding

`pricing` on `ai-agent-orchestration.com` cannot be rebuilt. The page-scoped rebuild was filed
**2026-08-17 20:28Z** (item `889a0687-cc0a-4f5e-8693-9ee6ca98751a`, `page-rerender` →
`page-build-handler`, `spec.reason='content_data_backfill'`) and **failed**:

```
save_page_sections: SECTION SHRINK REFUSED for page "pricing" —
call-to-action 483→213 chars of VISIBLE text (44% kept, floor 50%). Nothing was written.
```

I extracted the **483 chars the floor is protecting**. It is this, served live:

> Ready to Talk Architecture, Not Estimates? If you're evaluating AI agent infrastructure for
> production, the real costs are in failure modes, re-architecture, and provider lock-in — not the
> initial build. Before we talk scope, use our
> **`[LLM Provider Cost Comparison Calculator](/tools/tool-llm-cost-calculator.html)`** to get a
> concrete number on token costs across providers and whether self-hosting makes sense at your
> scale. Then let's have a direct conversation about what your system actually needs to run
> reliably. Discuss Your Requirements Explore Our Services

**The link is literal markdown, printed rather than rendered, on the live page.** So a slice of the
baseline the guard defends is not copy at all — it is a rendering defect, and the URL and bracket
punctuation are inflating the "kept" denominator. The site already carries open `literal_markdown`
items (1 `failed`, 2 `needs_human_review`), so this is a known family here rather than a one-off.

Two things follow, and only the first is established:

1. **[MEASURED]** the 44%-kept figure is computed against an inflated baseline, so the regenerated
   CTA is plausibly losing much less real copy than 56%.
2. **[NOT ESTABLISHED]** whether the replacement is *acceptable*. That needs the generated text,
   which was never written — the save was refused, correctly, and nothing is in `page_components`.
   Recovering it means `orchestration_states.collected_data` for that run, compared on the
   **visible-text** axis, not tag-stripped (the two axes agree on zero pairs in eight days —
   LANDMINES, `bugs_open/293`). ⚠ A log grep cannot recover it: chassis retention here is ~4
   minutes, so the query returns nothing **with its control also zero**.

## Why it is yours rather than mine

The owner has asked you to "ensure that sort of copy never leaves this framework again". This is
the same question from the other end: **a floor that measures a defective baseline will preferentially
refuse the repair.** Your stage-2 proposals face the same arithmetic — an edit that removes
unrendered markup, tightens a define-by-negation sentence, or cuts restatement is a *shrink*, and
the floor cannot tell it from a truncation. The existing LANDMINE ("the per-slot TEXT floor counts
CSS and JAVASCRIPT as text") records that the tag-stripped axis fires, across 117 archived pairs,
**exactly once — on the repair**. This is that shape again on the visible axis: the worse the
current copy, the harder it is to replace.

I have not touched the floor, filed anything against it, or re-run the rebuild. **The owner's
decision is "investigate the shrink first", and this is step 1 of that.**

## What I did do on this site, so it does not surprise you

Migration `469` (`469_departments_grid_and_leadership_team_consume_site_tokens.sql`, applied,
recorded, verified) tokenised two components' colours and re-rendered **`index` and `about`** only,
page-scoped, `spec.reason='template_changed'`. Contrast went 32 → 8 firm failures; `index` and
`about` are at 0. **No copy was changed** — it is a CSS-token change, and the re-render regenerates
from existing `content_data`. `pricing` was not touched and still serves its April render.

⚠ One thing worth your attention: those two components' `about` placement renders
`<p class="section-intro"><p>…</p></p>` — content that already carries its own `<p>`, wrapped again
by the template. The parser auto-closes the outer tag, so `.section-intro`'s styling reaches
nothing. Censused fleet-wide: **exactly 2 placements**, both on this site's `about`
(`content-block-about`/`body_text`, `leadership-team`/`section_intro`). Same family as the literal
markdown — LLM-authored prose arriving with block markup into a slot that assumes plain text.

## Pointers

- My full account: `site_ai_agent_orchestration/NOTES_site_improvement.md`, 2026-08-18 evening.
- The gate trap I filed: LANDMINES, *"`rebuild_policy='generic'` answers ONE of at least six
  independent refusals in `save_page_sections`"* — relevant to you because `copy_edit_proposed`
  applications face the other five.
- Reply into `site_ai_agent_orchestration/` as a CONTRIB, or straight into my NOTES.
