# HANDOFF — Article bodies: raw JSON envelope never parsed → 9 pages have SILENTLY LOST their article; 5 more show raw JSON on the page

**Created 2026-07-14 (imagery Turn 39). Found while trying to land I2.5 (sprite
bullets on a component wrapper) — the wrapper wasn't in the deployed markup, and
pulling that thread exposed this. NOT fixed; handed off deliberately.**
**Start a fresh chat from this file.** Severity: **live content loss across 5
sites.** The lost words are all RECOVERABLE (see §5).

---

## 1. One-paragraph summary

The content writer stores the LLM's **JSON envelope as a raw string** in
`page_components.content_data.result` — it is never parsed. The article-body
template needs `{{.content}}`, which is buried *inside* that string. Two
consequences, both live:

1. **9 article bodies have been BLANKED.** The light re-render path
   (`rerender_page_sections`) looks for `.content`, doesn't find it, renders an
   **empty** `.article-body__content`, and overwrites the good `rendered_html`.
   Page assembly then drops the visually-empty section — so the article silently
   **vanishes from the live page**. The words are still in the DB.
2. **5 article bodies LEAK RAW JSON.** Where no re-render has run, the raw
   envelope string is concatenated into the page, so readers literally see
   `{ "content": "` printed above the first heading.

Only **2 of 16** article-body instances are healthy. **14 of 16 are one
re-render away from being blanked (or already are).**

## 2. The trigger — and why this is urgent

`rerender_page_sections` runs on a **scoped** rerender, gated on
`spec.reason ∈ {image_landed, section_data_resolved}`
(`create_rerender_items_action.go:86`). **`image_landed` fires automatically when
an image asset lands on a page.** So routine imagery work silently destroys
article bodies. The imagery workstream (per-page heroes, Phase I0) is the most
likely cause of the 9 already-blanked pages — the latent bug is in the
writer/renderer, but landing images is what pulls the trigger.

**⚠️ DO NOT trigger a scoped rerender (or land an image) on any page in §4 until
this is fixed — it will blank the article.** Assemble-only rerenders (no
`spec.reason`) are safe.

## 3. Root cause, precisely

- `content_components.article-body.input_schema` declares exactly one field:
  `content` — `{"type":"text","source":"llm","required":true}`.
- The template is
  `…<div class="article-body__content">{{.content}}</div>…`.
- But the stored `content_data` is `{"type":"text","result":"{\n \"content\": \"<h2>…\"\n}"}`
  — i.e. **`result` is a STRING whose contents are a JSON document**. The `content`
  key exists only *inside* that string, so `{{.content}}` resolves to nothing.
- `RenderTemplate` flattens `content_data` verbatim into the template map
  (`contextToInterfaceMap`, `component_library.go:~740`), and `executeGoTemplate`
  runs with `Option("missingkey=zero")` (`call_agent.go:1152`) — so a missing
  `content` renders as **empty, silently**. No error, no warning.
- **All 14 broken envelopes are MALFORMED JSON** (literal newlines inside the
  string values), so a strict `::jsonb` cast fails on every one. That is very
  likely *why* the writer never parsed it and fell back to storing the raw text.

**Verified offline (safe probe, no live change):** rendering the real template
against the stored `{type,result}` content_data produces
`<div class="article-body__content"></div>` — empty. Against a `{content: …}`
content_data it renders correctly. That probe is what stopped this handoff from
becoming a fourteenth blanked page.

## 4. Blast radius (2026-07-14)

| state | components | sites |
|---|---|---|
| **BLANKED** — article silently lost from the live page (`rendered_html` = 1,326-byte empty shell) | **9** | 5 |
| **JSON LEAK** — `{ "content": "` visible to readers | **5** | 4 |
| healthy (`content_data` has a real `content` key) | 2 | 1 |

**Blanked** (article gone from the live page — content still in the DB):
leopardessconsulting.co.uk `/blog/why-most-ai-agent-projects-never-reach-production.html`,
`/guides/tool-agent-complexity-estimator-guide.html`,
`/blog/hierarchical-multi-agent-orchestration-explained.html`;
finetuning.uk `/blog/why-most-ai-projects-fail-in-the-first-three-months.html`,
`/guides/tool-ai-data-risk-checker-guide.html`, `/guides/llm-cost-calculator-guide.html`;
ai-agent-orchestration.com `/why-kubernetes-kafka-postgres-for-ai-agent-orchestration.html`;
robot-hands.com `/guides/tool-gripper-cycle-time-estimator-guide.html`;
gamesdesign.co.uk `/guides/tool-xp-curve-designer-guide.html`.

**JSON leak** (raw JSON visible):
robot-hands.com `/guides/tool-grip-force-friction-calculator-guide.html` (the sprite-bullet gate page),
robot-hands.com `/blog/tool-gripper-payload-calculator-guide.html`,
finetuning.uk `/blog/how-to-audit-your-business-for-ai-automation.html`,
ai-agent-orchestration.com `/blog/multi-agent-failure-isolation-patterns-production.html`,
gamesdesign.co.uk `/guides/tool-drop-rate-tuner-guide.html`.

These are the SEO content pages — guides and blog posts. Losing their body copy
is the worst possible page to lose it on.

## 5. Recovery — the words are all still there

Every one of the 14 has a recoverable envelope (`result` matches
`^\s*\{\s*"content"` and contains real `<h2>`/`<p>` HTML). The repair per
component:

1. **Parse tolerantly.** A strict `::jsonb` cast FAILS on all 14 (literal
   newlines). Extract with a lenient parser — Go's `json.Unmarshal` after
   escaping raw newlines, or a targeted extraction of the `content` value.
2. **Normalise `content_data`** so it matches the schema the template and the
   re-render path both expect: `content_data.content = <the article HTML>`
   (keep `result`/`type` — don't delete evidence).
3. **Re-render the section** through the template (scoped rerender is now SAFE
   for that component) → correct `rendered_html`, wrapper present, JSON gone.
4. Verify on the served page: article text present, no `{ "content": "`, and
   `<div class="article-body__content">` in the HTML.

Back up first: `CREATE TABLE bak_pc_articlebody_<date> AS SELECT * FROM
page_components WHERE …`.

## 6. The real fix (upstream — do this too, or it recurs)

Repairing the 14 rows fixes today's pages; the writer will keep producing the
same envelope tomorrow. Candidates, in order:

1. **Writer: parse the envelope.** Where the LLM result is persisted to
   `content_data`, unwrap a JSON envelope into its fields (the code already knows
   this shape — `html_actions.go:358` and `site_db_actions.go:863` both probe for
   `html`/`content`/`result` keys; that normalisation just isn't applied here).
   Demand strict JSON from the model, or repair the newlines before parsing.
2. **Renderer: fail loudly, not silently.** `missingkey=zero` turns "the required
   field is missing" into "render an empty div". A component whose schema marks a
   field `required: true` should NOT render empty — it should fail the step or
   emit `needs_human_review`, not quietly ship a blank section. This is the same
   class of bug as the product-page one (`HANDOFF_2026-07-14_empty_product_sections.md`).
3. **Assembly: don't hide the evidence.** `sectionHasVisibleContent`
   (`rerender_single_page_action.go:436`) silently drops the empty section, which
   is why nobody noticed 9 articles disappearing. Dropping a section that the plan
   says should have content deserves at minimum a work item.
4. **A discovery check:** article-body (or any component) whose schema-required
   fields are absent from `content_data` → work item. Sibling of
   `check_required_fields_missing` (which already exists for the product-page
   defect — extend it rather than duplicate).

## 7. Code map

| what | where |
|---|---|
| light re-render that blanks the section | `platform/orchestration/actions/rerender_page_sections_action.go` (renders stored content_data through `RenderTemplate`) |
| the `image_landed` / `section_data_resolved` gate | `platform/orchestration/actions/create_rerender_items_action.go:86` |
| verbatim flatten of content_data into template data | `component_library.go` → `contextToInterfaceMap`, `RenderTemplate` |
| silent empty render | `call_agent.go:1152` — `Option("missingkey=zero")` |
| empty section dropped at assembly | `rerender_single_page_action.go:436` `sectionHasVisibleContent` |
| existing envelope normalisation (not applied here) | `html_actions.go:358`, `site_db_actions.go:863` |

## 8. Queries

```sql
-- The three states
SELECT CASE WHEN btrim(pc.rendered_html) LIKE '{%' THEN 'JSON LEAK'
            WHEN length(pc.rendered_html) = 1326 THEN 'BLANKED'
            WHEN pc.content_data ? 'content' THEN 'healthy' ELSE 'other' END AS state,
       s.domain, p.url
FROM page_components pc
JOIN content_components cc ON cc.id = pc.component_id
JOIN pages p ON p.id = pc.page_id JOIN sites s ON s.id = p.site_id
WHERE cc.name = 'article-body' ORDER BY state, s.domain;

-- Anything one re-render away from being blanked (schema wants `content`, row hasn't got it)
SELECT count(*) FROM page_components pc
JOIN content_components cc ON cc.id = pc.component_id
WHERE cc.name = 'article-body' AND NOT (pc.content_data ? 'content');   -- 14 of 16
```
