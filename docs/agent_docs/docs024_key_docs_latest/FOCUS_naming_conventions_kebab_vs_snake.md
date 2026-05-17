# FOCUS — Naming Conventions for String-Typed Values

Date: 2026-05-17
Status: applied (migration 051, page_canonical.go + page_role_validator.go updated, contracts doc v9, debug guide v2.10)

This is the reasoning behind the kebab/snake split documented in `003_contracts_and_standards_v9.md`. Reading this first will save the next person several hours of "why didn't they just pick one?"

## How we got here

The 2026-05-17 gamesdesign.co.uk readoption surfaced a structural inconsistency in `pages.page_type`:

```
content          | 110
blog-post        |  42   ← kebab
tool             |  29
blog-index       |   3   ← kebab
blog_post        |   3   ← snake (same conceptual type as blog-post above)
index            |   2   ← name conflated with type
game             |   2
blog_index       |   2   ← snake
news-index       |   1
entity_directory |   1
entity_page      |   1
```

Same conceptual type stored two ways in the same column. Reader code (e.g. `WHERE page_type = 'blog-post'` at chassis line 42374, `pageType == "blog-index"` at line 14833) was silently missing the snake rows. The bug had been latent because the snake rows were a minority and the missing-page consequences manifested only on sites with many blogs.

Root cause: `CanonicalisePage.normalisePageType()` was actively converting kebab inputs to snake outputs, while everything reading `page_type` expected kebab. The canonicaliser was a one-way translation that nothing else was aware of.

## What I considered

Three end-states were on the table:

| Approach | Direction | Migration cost | Reader changes | Standards-doc alignment |
|---|---|---|---|---|
| **A — kebab everywhere** | flip the writer | 7 rows | none — readers already kebab | matches existing kebab rule |
| **B — snake everywhere** | flip the readers | 189 rows | 6+ Go sites | requires standards-doc exception |
| **C — tolerant readers** | normalise at read time | 0 | every reader needs a wrapper | explicitly discouraged by standards |

C was rejected on principle (the standards doc says fallbacks should not be relied upon, and a fallback would just preserve the underlying inconsistency).

My initial lean was B (snake everywhere), reasoning that:
- The canonicaliser writes snake; tests assert snake; it's the "documented design"
- `site_work_items.item_type` is snake (`needs_blog_posts`); analogous "type column"
- snake matches Go map keys and switch-case constants

This was wrong on inspection.

## Why kebab won

Five things in order of weight:

### 1. The framework-convention argument is silent for `page_type`

I'd assumed "follow the language convention" would settle it. It doesn't. `page_type` never appears as:

- An HTML attribute (would force kebab — `data-foo`)
- A URL path segment (would force kebab — convention)
- A Go identifier (would force CamelCase — N/A for a string value)
- A SQL identifier (would force snake — but the value isn't an identifier, it's data in a `text` column)
- A switch-case key in any Go file (verified by grep — no `case "blog-post":` or `case "blog_post":` switching on page_type)

Both kebab and snake are equally legal everywhere `page_type` flows. So framework convention doesn't pick; the choice rests on consistency with existing codebase realities.

### 2. The data was already mostly kebab

189 of 196 rows were kebab. Migration cost is dominated by the smaller side: 7-row kebab fix vs 189-row snake fix. The migration that landed (051) touched 13 rows total including Phase 1.5; the snake-everywhere alternative would have touched 196.

### 3. The readers were already kebab

Every kebab-form reader I found in the codebase (`WHERE page_type = 'blog-post'`, `pageType == "blog-index"`, set-membership maps with `"blog-index": true`) was correct against the kebab data but broken against the snake data. Going kebab makes those work. Going snake would have required updating every one of them — bigger surface area.

### 4. `content_components.function` is kebab and is the documented canonical

The standards doc has an existing kebab rule for `function`. `pages.page_type` is conceptually parallel: pure metadata, never an identifier, used as data describing what a thing is. Extending the kebab convention to it makes the doc internally consistent — one rule for "this kind of column."

### 5. The snake forms were genuinely localised

I'd treated the snake usage in `page_canonical.go`, `page_role_validator.go`, and their test files as "documented design." They weren't. They were a leak. The tests asserted snake because the function returned snake; the function returned snake because someone wrote it that way; nothing in the doc justified it. Once I traced the lineage, "the canonicaliser writes snake" stopped being a reason to keep snake — it was the bug.

## The two conventions, and the rule for choosing

The standards doc now codifies what the codebase actually does:

- **Identifier-shaped values** → `snake_case`  
  Used as Go map keys, `switch case` constants, action registry names, work-item dispatch routes. Examples: `site_work_items.item_type` (`needs_blog_posts`, `orphan_blog_posts`), action names (`create_blog_posts`, `apply_adoption_plan`).
  
- **Data-shaped values** → `kebab-case`  
  Describe what a thing is, never used as identifiers in code. Examples: `content_components.function`, `pages.page_type`, `agent_definitions.type` (which appears in Kafka topic segments and Kubernetes labels where kebab is the convention).

- **Single-word values** → lowercase word (no separator)  
  Status enums and similar. Examples: `pages.build_status` (`planned`, `deployed`), `site_work_items.status` (`detected`, `triaged`, `claimed`).

**The decision rule:** for any new string-typed column or enum-like value, ask "is this value ever used as an identifier in Go code?" — meaning a `switch case` constant, a map key in a registry-shaped lookup, or a dispatch-route segment.

- Yes to any → snake
- No → kebab
- Single word → just the word

This isn't aesthetic; it follows a real distinction in how the value gets consumed. Code consumers (switch statements, registries, routers) prefer snake because the value reads as an identifier in surrounding Go code. Data consumers (CSS templates, URLs, HTML attributes, LLM prompt content) prefer kebab because kebab is the dominant convention there.

## Phase 1.5 — landing vs index, separately

Independent of kebab/snake, the homepage's page_type had been written as `"index"`. Wrong category: "index" is the page's name (storage convention for the homepage); "landing" is its type. The column is called `page_type`, so it should hold the type.

Renaming `index` → `landing` in `page_type` lets:
- The homepage and other landing pages share rendering / CTA / nav-suppression logic via `WHERE page_type = 'landing'`
- Future analytics distinguish landing pages from content pages without name inspection
- The same row act as the homepage (`name = 'index'`) and as a landing page (`page_type = 'landing'`) without one fact shadowing the other

The page name stays `"index"` for the homepage. Only the type column moved.

## What we didn't do, and why

### We didn't audit every type-like column in the database

There are at least ten "type-like" columns: `site_specs.aspect`, `pages.build_status`, `agent_definitions.category`, `agent_definitions.agent_category`, `research_results.result_type`, etc. Each could in principle have its own kebab/snake drift. Auditing them all would have ballooned this fix from "two-hour focused change" to "two-day project."

The plan: revisit when a specific column drifts, apply the same rule, document the constraint. The two-conventions rule in the standards doc is the persistent artefact that lets future contributors classify new columns without re-deriving the reasoning.

### We didn't drop the snake-input fallback in `normalisePageType`

The function still accepts `blog_post`, `section_index`, etc. as input and converts to kebab on output. Older planner prompts may still emit snake forms; the fallback prevents breakage.

This is a controlled exception to the standards doc's "fallbacks should not be relied upon" rule. The exception is bounded: it only applies during the migration tail, and once we confirm no upstream is still emitting snake we can simplify the function to passthrough. Until then, the fallback is the cheaper path than tracking down every old prompt.

### We didn't enforce snake_case on identifier-shaped columns

The standards doc declares the rule but `site_work_items.item_type` and similar columns don't yet have a CHECK constraint. That's a follow-up. The kebab constraint on `pages.page_type` is in place because the immediate bug was kebab-side; symmetric enforcement on the snake side awaits a separate change.

## Things to flag for future contributors

1. **A snake-form planner prompt is silently doing the right thing through the fallback.** If you see snake values in LLM outputs that end up as kebab in the DB, the fallback caught it. Fix the prompt rather than relying on the fallback.

2. **Tests document behaviour, not intent.** The original `page_canonical_test.go` asserting `wantType: "blog_index"` looked like a deliberate contract. It wasn't — it was just describing what the function did. Don't read tests as design.

3. **The "index" page_type is now a constraint-allowed value that nothing should emit.** The CHECK constraint accepts `index` because the regex permits single-word values. If a row appears with `page_type = 'index'` after migration 051, it's a writer that skipped the canonicaliser or pre-dates the deploy.

4. **`site_work_items.item_type` should NOT migrate.** It looks like another type column but it's identifier-shaped — used in switch-case routing in `load_work_item_actions.go`. Snake is correct there.

## References

- Migration: `051_pages_page_type_kebab.sql`
- Code: `platform/orchestration/datahelpers/page_canonical.go`, `page_role_validator.go`
- Tests: `page_canonical_test.go`, `page_role_validator_test.go`
- Standards: `003_contracts_and_standards_v9.md`, section "String-Value Naming Convention"
- Debug: `016_debugging_guide_v2_10_.md`, section 6.5 and the two new section-9 entries
