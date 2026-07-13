# FOCUS — Language

A short reference covering where language assumptions sit in the platform
today, what works implicitly, and which surfaces will need attention as
non-English sites come into scope. Not a plan — a map.

## Where we are today

Everything we've built so far works through the assumption that the
content language is whatever the brief, site_specs, and existing content
say it is. That assumption is implicit, not declared. For English sites
it's invisible; for non-English sites it's mostly invisible too, but
with seams where English bias can leak in through prompts, examples,
and hardcoded helper text.

There is no `language` field on `sites`, `pages`, `content_components`,
or `site_specs` today. Routing, deployment, and HTML generation all
proceed without consulting a target language because they don't need to
yet.

## What works implicitly

The LLM follows the surrounding context. When the brief, content_direction,
existing_content, and admin briefs are all in language X, an instruction-
tuned model produces output in language X without being told. This has
been the de facto mechanism on every site we've built.

Two reasons it works:

- The brief comes from the human owner of the site. They write it in the
  language they expect the site to be in.
- Adopted sites (recreate mode) carry the original content into
  `existing_content.raw_markdown`. The prompt feeds it back as
  reference, and the LLM matches register and language.

## Where English is hardcoded today

After Step 3 the page-content-writer prompt has only one explicit
language signal — a `## Language` section that tells the LLM to match
the input language. Strict rules use abstract phrasing without
English-only examples.

Other surfaces still carry English assumptions that will need attention
when non-English sites land:

- **Component `llm_guidance` strings** on `content_components.input_schema`.
  These are the schema author's per-field instructions ("Primary heading
  for the tool list section. 6-12 words."). Today they're English. The
  prompt translates the intent at runtime, but the guidance itself is
  authored in English. When schema authors begin producing components
  for non-English sites, they'll either author guidance in the target
  language or keep it structural ("primary section heading, short").
  No system change required — it's a schema-authoring concern.

- **Static fallback values** in input_schema (e.g. `fallback: "Browse
  All Tools"` on a Tier B label). These bypass the LLM. For a non-English
  site, the fallback would render in English unless the schema declares
  a different default — or the LLM is allowed to override (see
  "soft static" below).

- **Admin content briefs** stored on `page_components.content_brief`.
  Written by admins; assumed to match the site's language. No
  enforcement.

- **Site spec auto-generated text**. `site_type_reasoning`,
  `growth_path`, `content_strategy` etc. produced by the domain
  strategist are currently English-output even when the target site
  isn't. Those fields are mostly internal — not user-facing — so the
  language mismatch is a reasoning-vs-output question rather than a
  user-experience one. Worth knowing.

- **Other agent prompts**. domain-research-classifier, domain-strategist,
  briefing-agent, site-planner, component-creator, content-reviewer —
  each has its own prompt template that may carry English-specific
  examples or instructions. Step 3 has only audited page-content-writer.
  Each agent needs the same sweep when non-English sites land.

- **HTML output**. `<html lang="...">` is not currently set per site.
  For SEO and accessibility, non-English sites will need it. Lives in
  the head template / site-components, not the per-section prompt.

## Surfaces likely to need attention

In rough priority order if non-English sites become a near-term goal:

1. **Other agent prompts.** The page-content-writer audit found three
   classes of English bias: named examples in rules, English-implying
   guidance, and the meta-explanation in "What To Write". Same patterns
   likely exist in the other content-producing agents.

2. **HTML `lang` attribute.** Cheap to add, important for accessibility
   and SEO. Sits in the head template; needs a per-site language string
   to read from.

3. **Component schema guidance.** Decide whether `llm_guidance` should
   be language-neutral structural advice or per-site translated. The
   per-section prompt today asks the LLM to translate the intent, which
   works but adds a cognitive step. Language-neutral guidance ("primary
   heading, 6-12 words") is robust across languages.

4. **Static fallbacks in input_schema.** A fallback like
   `"Browse All Tools"` rendering on a Spanish site is wrong. Options:
   (a) the LLM may override Tier B labels — see the "soft static"
   concept in the Step 2 observations; (b) per-site fallback overrides
   on the page_components row; (c) language-aware fallback tables. No
   decision needed yet.

5. **Existing-content language detection.** During adoption, the crawler
   pulls page text. A small language-detection step at adoption time
   would let the system flag the site's primary language and propagate
   it everywhere. Cheap (langdetect/cld3), useful as a signal even
   before any system change responds to it.

## Open design questions to defer

- **Should `sites` carry a `primary_language` column?** Probably yes
  eventually. Drives the HTML lang attribute, can be inspected by
  prompts, makes language an explicit not implicit signal. Don't add
  until there's a concrete consumer (e.g. the head template needs it).

- **Should the prompt explicitly pass a target language to the LLM,
  or continue to rely on input-context matching?** Today's approach
  (match the language of the brief/specs/existing-content) works as
  long as the inputs are coherent. If you ever need to write a section
  in language A when most context is in language B (e.g. an English
  "About Us" on an otherwise-Spanish site for international visitors),
  an explicit language parameter helps. Until that requirement is
  real, the implicit approach is simpler.

- **Static fallback handling for non-English sites.** As above — the
  fallback bypass-LLM behaviour is correct for English where the
  fallback is in-language, but produces wrong output for other
  languages. Resolves into the soft-static question and per-site
  overrides.

## What Step 3 changed

The page-content-writer prompt:

1. Adds an explicit `## Language` section instructing the LLM to write
   in the same language as the brief, site specs, and existing content.

2. Removes English-language examples from strict rules. Rule 13 used to
   say `No "Sarah Mitchell, Founder"`; now it says "Never invent fake
   people, client names, or attributed quotes."

3. Acknowledges in "What To Write" that the schema-author's
   `llm_guidance` may be in English even when the output should be in
   another language: "translate the intent into the output language
   as needed."

4. Rule 4 (placeholder text): broadened from "Lorem ipsum" alone to
   "[Your Company] or Lorem ipsum, in any language."

These are the smallest possible language-aware changes. They don't
solve cross-agent consistency, HTML lang attributes, or schema-authoring
conventions. They make the per-section prompt language-agnostic so it
doesn't bias toward English when the inputs aren't.

## Verification when a non-English site lands

When the first non-English site is built:

- Confirm `## Language` actually steered output (read the
  `agent_chassis` logs for `generate_content` — the LLM response
  should be in the target language).
- Check static fallback labels (Tier B) — if the site uses a component
  whose fallback is `"Browse All Tools"`, that English string will
  appear on the rendered page. Decide per-site whether to override
  or accept.
- Check the `<html lang>` attribute — likely still hardcoded.
- Check site_specs auto-generated text — likely still in English, but
  may not matter if it's internal-only.

## Files relevant to this document

- `step3b_prompt_template.txt` — the page-content-writer prompt with
  the language changes
- `step3b_workflow.json` — workflow JSON containing the prompt
- `STEP3_changelog.md` — full Step 3 change description

## Pointers

- Step 2 observations flagged the prompt's English bias as
  out-of-scope for that change. Step 3 addressed it.
- The "soft static" concept (allowing LLM override of Tier B labels
  when site tone warrants) was raised in Step 3 seam-4 findings. Not
  yet a feature, but the place where per-site Tier B fallbacks would
  most naturally live.
