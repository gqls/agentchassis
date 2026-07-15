# HANDOFF — landing an image on a page can SILENTLY BLANK its article body

**Purpose.** Start a fresh chat from exactly here. This is a **platform data-loss
trap**, not a bug in one site. Landing (or rebuilding) an image on a page fires a
*scoped section re-render*; on any page whose article-body content was never
unwrapped from its LLM JSON envelope, that re-render renders the body **empty** and
**overwrites the good HTML with a blank shell** — the article vanishes from the live
page. Found in the imagery workstream (imagery RUNNING_NOTES Turns 39–41); the full
root-cause + recovery write-up is the parent handoff (see §7).

**Filed:** 2026-07-15
**Severity:** High — silent live content loss on SEO pages (guides/blog). Already
happened to 9 pages across 5 sites; 13 pages sit vulnerable right now.

**⚠️ THE TRAP IS LIVE IN PRODUCTION AS OF THIS WRITING** (prod = `v1.0.1122`). The
committed guard that would stop it is NOT yet in the running binary — see §3. **Until
it is deployed AND verified, do not land an image or trigger a scoped re-render on any
page in §5.**

## Working rules (hold these)
Go, not Python. British English. **Schema first**: read `\d <table>` before SQL, read
the function before changing it. Structural fixes over patches. Reuse existing
functions. Go changes are **inert until the chassis image is rebuilt** — DB config is
live immediately. Verify a deploy by grepping the RUNNING POD's binary, never git or
timing:
`kubectl exec -n ai-persona-system <pod> -- sh -c 'strings /app/agent-chassis | grep -c "<a log string from your change>"'`.
DB: `PSQL="kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db"`.

## 1. The mechanism (why an image landing destroys text)

1. An image asset lands on a page → `flag_page_image_rebuild_action.go:133` creates a
   `page_rerender` work item with `spec.reason = "image_landed"` and the dependent
   `component_id`.
2. `create_rerender_items_action.go:89` sees `reason ∈ {image_landed,
   section_data_resolved}` + a component_id → marks the item **scoped**, so
   `page-rerender` runs `rerender_page_sections` (section re-render) instead of the
   plain assemble-only path.
3. `rerender_page_sections` re-renders each section from its **stored content_data**
   through the template. The article-body template needs `{{.content}}`, but the
   affected rows store `content_data = {"type":"text","result":"{ \"content\": \"<html>\" }"}`
   — the LLM's JSON envelope as a **raw, never-parsed string**. There is no top-level
   `content` key.
4. Go templates run with `Option("missingkey=zero")` (`call_agent.go:1152`), so the
   missing `{{.content}}` renders as **empty — no error, no warning**. The section's
   `rendered_html` is overwritten with a ~1,326-byte empty shell.
5. Page assembly then drops the visually-empty section
   (`sectionHasVisibleContent`, `rerender_single_page_action.go:436`), so the article
   simply **disappears from the served page**. The words survive only in
   `content_data.result`.

The same unparsed envelope has a second, milder failure mode: where NO scoped
re-render has run, the raw envelope string is served verbatim, so readers see
`{ "content": "` printed above the first heading (a "JSON leak" page).

## 2. Trigger surface — what is safe vs unsafe

- **UNSAFE (blanks the article):** anything that produces a `page_rerender` with
  `spec.reason = "image_landed"` or `"section_data_resolved"` on an affected page —
  i.e. **an image asset landing on the page**, or a deferred section field becoming
  resolvable. The imagery workstream lands images (per-page heroes etc.), which is how
  the original 9 were blanked.
- **SAFE:** an **assemble-only** `page_rerender` (no `spec.reason`) — it only
  re-concatenates existing `rendered_html`, never re-renders sections. This is how the
  gate page was repaired without risk (imagery Turn 41).

## 3. Current state — READ THIS, it is time-sensitive

Two fixes for this exist **in source** and are being carried by other workstreams, but
the state is mixed:

- **The guard** (`rerender_page_sections_action.go` ~line 174–207, from the
  `empty_sections_loop_integrity` workstream): before re-rendering, it checks each
  section and, if content_data is absent **or missing a schema-required
  `source:"llm"` field** (`missingRequiredLLMFields`), it escalates the whole page to
  the writer via `escalateRerenderToWriter` **instead of blanking**. This closes the
  trap.
  - **Committed in source** (HEAD). **NOT in production `v1.0.1122`:** the running
    binary has `escalateRerenderToWriter` but NOT `missingRequiredLLMFields` /
    `"escalating page to writer instead of blanking"` — i.e. it has an earlier partial
    version that escalates only on *absent* content_data, not on the
    *missing-required-field* case the article-body rows hit. **So the trap is still
    live in prod for these pages.** Verify after the next deploy by grepping the pod
    for `escalating page to writer instead of blanking`.
- **The writer-side envelope repair** `ParseLLMJSON` (`json_envelope.go`, wired at
  `ai_actions.go:478`): repairs escaping-only malformations so future writes store a
  clean `content`. **But its test fails on 14 fixtures** (`002_HANDOFF…` §B —
  `TestParseLLMJSON_RepairsLiveEnvelopes`), which are almost certainly these same
  envelopes: some are escaping-malformed (repairable) and some are **truncated
  mid-string** (unrecoverable — the tail is not in storage). So escalation-to-writer
  may NOT cleanly regenerate every affected page.
- **`missingkey=zero`** (`call_agent.go:1152`) and the empty-section drop
  (`sectionHasVisibleContent`) are **unchanged**. The guard prevents *this* blanking
  path; the underlying "render a required field as empty, silently" pattern still
  exists elsewhere (cf. the product-page defect,
  `HANDOFF_2026-07-14_empty_product_sections.md`).

## 4. What actually needs doing (beyond the committed guard)

1. **Deploy + VERIFY the guard reaches prod** (grep the pod for `escalating page to
   writer instead of blanking`). Until then, hold the §2 operating rule.
2. **Recover the 13 already-broken pages** — the guard prevents *future* blanking; it
   does not fix existing bad rows. Recovery = extract the article from
   `content_data.result`, set `content_data.content`, re-render. Recipe + caveats
   (the envelope is NOT valid JSON — bare HTML quotes, so string-surgery not a parse;
   some are truncated → partial recovery) are in the parent handoff §5.
3. **Fix `ParseLLMJSON`'s 14 fixtures** (`002…` §B) — decide which are repairable vs
   quarantine the truncated ones — so writer-escalation actually regenerates these
   envelopes rather than looping.
4. **Consider hardening the silent path**: a component whose schema marks a field
   `required:true` should not render empty — fail the step or flag, not blank
   (same class as the product-page defect).

## 5. Affected pages (13, as of 2026-07-15 — re-run the query in §6)

**BLANKED (9 — article already gone from the live page; text still in content_data):**
- ai-agent-orchestration.com `/why-kubernetes-kafka-postgres-for-ai-agent-orchestration.html`
- finetuning.uk `/blog/why-most-ai-projects-fail-in-the-first-three-months.html`
- finetuning.uk `/guides/llm-cost-calculator-guide.html`
- finetuning.uk `/guides/tool-ai-data-risk-checker-guide.html`
- gamesdesign.co.uk `/guides/tool-xp-curve-designer-guide.html`
- leopardessconsulting.co.uk `/blog/hierarchical-multi-agent-orchestration-explained.html`
- leopardessconsulting.co.uk `/blog/why-most-ai-agent-projects-never-reach-production.html`
- leopardessconsulting.co.uk `/guides/tool-agent-complexity-estimator-guide.html`
- robot-hands.com `/guides/tool-gripper-cycle-time-estimator-guide.html`

**JSON LEAK (4 — `{ "content": "` visible; a scoped re-render would blank them):**
- ai-agent-orchestration.com `/blog/multi-agent-failure-isolation-patterns-production.html`
- finetuning.uk `/blog/how-to-audit-your-business-for-ai-automation.html`
- gamesdesign.co.uk `/guides/tool-drop-rate-tuner-guide.html`
- robot-hands.com `/blog/tool-gripper-payload-calculator-guide.html`

(robot-hands `/guides/tool-grip-force-friction-calculator-guide.html` was the 14th; it
was hand-repaired in imagery Turn 41 and is now healthy — a worked example of §4.2.)

## 6. Detection query

```sql
-- state of every article-body instance
SELECT CASE WHEN btrim(pc.rendered_html) LIKE '{%' THEN 'JSON LEAK'
            WHEN length(pc.rendered_html) = 1326 THEN 'BLANKED'
            WHEN pc.content_data ? 'content' THEN 'healthy' ELSE 'other' END AS state,
       s.domain, p.url
FROM page_components pc
JOIN content_components cc ON cc.id = pc.component_id
JOIN pages p ON p.id = pc.page_id JOIN sites s ON s.id = p.site_id
WHERE cc.name = 'article-body' ORDER BY state, s.domain;

-- generalised: any component one scoped re-render away from blanking
-- (schema requires a source:"llm" field the row hasn't got as a top-level key)
```

## 7. Code map & cross-references

| what | where |
|---|---|
| emits `reason:"image_landed"` (the trigger) | `flag_page_image_rebuild_action.go:133` |
| routes it to the scoped section re-render | `create_rerender_items_action.go:89` |
| the section re-render (blank path + the committed guard) | `rerender_page_sections_action.go:174–207` |
| silent empty render | `call_agent.go:1152` (`Option("missingkey=zero")`) |
| empty section dropped at assembly | `rerender_single_page_action.go:436` (`sectionHasVisibleContent`) |
| writer-side envelope repair (test failing) | `json_envelope.go`, wired `ai_actions.go:478` |

- **Parent — full root cause, all 14 pages, recovery recipe:**
  `../HANDOFF_2026-07-14_article_body_json_envelope.md`
- **The `ParseLLMJSON` failing-fixtures test:** `002_HANDOFF_2026-07-15_errors_to_fix.md` §B
- **Same class (authoritative store vs deployed diverging; silent overwrite):**
  `001_HANDOFF_replan_clobbers_built_pages_FIX.md` and `002…` §C
- **Sibling silent-required-field defect (product pages):**
  `../HANDOFF_2026-07-14_empty_product_sections.md`
