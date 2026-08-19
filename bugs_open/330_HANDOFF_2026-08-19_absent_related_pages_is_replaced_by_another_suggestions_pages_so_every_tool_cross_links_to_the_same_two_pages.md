# 330 — an absent `related_pages` is replaced by ANOTHER suggestion's, so every tool on a site cross-links to the same two pages

**Filed** 2026-08-19 ~21:00Z by the `staged_component_build` lane.
**Found by** this lane's own RFC_029 §9 Phase 1 conflict instrument
(`RESOLVER_CONFLICTING_CANDIDATES` in `agent_error_log`) — not by a symptom report.
**Status** OPEN. Mechanism confirmed end to end first-hand; damage measured; a negative
control at four other sites behaves correctly.
**090 diagnosis filed** 2026-08-19 21:56Z — `RUN_CORRELATION_ID=fad0675b-007e-4343-8887-5e6aede6e415`
(intake `5a47d9a2-…`; the RUN id is the one the artifacts are written under). Filed per the owner
ruling of 2026-07-31: this asserts a cross-cutting cause that lives in shared infra
(`extractSingleField`'s arm chain), not in the symptom's own file. Read it before acting on §6:
```sql
SELECT created_at, kind, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='fad0675b-007e-4343-8887-5e6aede6e415' ORDER BY created_at;
```
A missing row is latency, not a drop — do not re-trigger.

## 1. Symptom

On `webdesign.co.uk`, **nine different tools** were each given cross-link work items pointing at
**the same two pages** — `learn-algorithms-p-values-explained` and
`learn-algorithms-bayesian-theory` — regardless of what the tool does:

```
tool_crosslink:tool-svg-optimizer:learn-algorithms-p-values-explained:6b49db8e…
tool_crosslink:tool-json-cleaner:learn-algorithms-bayesian-theory:6b49db8e…
tool_crosslink:tool-smooth-shadow:learn-algorithms-p-values-explained:6b49db8e…
tool_crosslink:tool-css-variables:learn-algorithms-bayesian-theory:6b49db8e…
tool_crosslink:tool-prompt-permutator:learn-algorithms-p-values-explained:6b49db8e…
   … 9 distinct tools, 2 distinct target pages
```

"Add SVG Code Stripper tool reference to learn-algorithms-p-values-explained" is the shape of
every one of them. A per-tool value would vary by tool. **It does not vary — that constancy is
the tell**, and it is the only thing that distinguishes this from ordinary cross-linking:

```sql
SELECT s.domain, count(*) n,
       count(DISTINCT split_part(w.item_key,':',2)) AS distinct_tools,
       count(DISTINCT split_part(w.item_key,':',3)) AS distinct_target_pages
FROM site_work_items w JOIN sites s ON s.id=w.site_id
WHERE w.item_key LIKE 'tool_crosslink:%' GROUP BY 1 ORDER BY n DESC;
```
`webdesign.co.uk` → 32 rows, **9 tools, 2 pages**. Every other site → 1 tool, 3 pages.

## 2. Root cause — the full chain, each link read first-hand

1. **The config wiring is present and correct.** Migration `211` wired
   `save_tool.config.related_pages = "input_data.spec.related_pages"` (also `description`,
   `function`). Verified live, not assumed:
   ```sql
   SELECT d.type, s.key, s.value->'config'->>'related_pages'
   FROM agent_definitions d, jsonb_each(d.default_config->'workflow'->'steps') s
   WHERE d.is_active AND COALESCE(d.is_snapshot,false)=false AND d.deleted_at IS NULL
     AND s.value->>'action' IN ('create_tool_component','deploy_tool_to_site');
   ```
   → `tool-generator/save_tool` and `tool-deployer/deploy_tool`, both wired. **211 has NOT
   regressed** — see §5, this was my first hypothesis and it was wrong.
2. **But the `add_tool` spec usually has no `related_pages` to read.** Four of the five most
   recent `tool-generator` runs carry `input_data.spec` **without** the key:
   ```sql
   SELECT created_at, collected_data->'input_data'->'spec' ? 'related_pages'
   FROM orchestration_states WHERE owner_agent_type='tool-generator'
     AND created_at >= '2026-08-18' ORDER BY created_at DESC LIMIT 5;
   ```
   → `f, t, f, f, f`.
3. **A declared path that resolves to nothing falls through the cascade.**
   `extractSingleField` (`platform/orchestration/datahelpers/unified_extractor.go:417`) runs
   its arms in order: direct-path → input-data-prefix → input-data-map → **whole-tree-search**
   → alias. Nothing stops at "the step declared a path and the path was empty".
4. **The whole-tree search finds every `related_pages` in the tree** — all of them inside
   `load_brand_context.specs.tools.suggestions[N].related_pages` — they conflict, and the
   shallowest-first winner is `suggestions[0]`'s. Logged, every time:
   ```
   field=related_pages  winner=load_brand_context.specs.tools.suggestions[0].related_pages
   ```
   17 such rows since 2026-08-16.
5. **The guessed value pre-empts the correct fallback.**
   `relatedPagesFromInputs` (`create_tool_cross_link_items.go:443`) reads
   `inputs.GetRaw("related_pages")` FIRST and only falls back to
   `ExtractNestedField(collected, "input_data.spec.related_pages")` when that is empty. The
   resolver has already filled it with the guess, **so the fallback the 029 fix installed can
   never run.** The comment above it ("the direct read is the fallback for a config that
   predates it") describes a path that is now unreachable whenever the search finds anything.
6. **Confirmed at the source.** `suggestions[0]` on that site is the *A/B Test Significance
   Calculator*, whose `related_pages` are exactly the observed targets:
   ```sql
   SELECT ss.data->'suggestions'->0->>'name',
          ss.data->'suggestions'->0->'related_pages'
   FROM site_specs ss WHERE ss.site_id='6b49db8e-…' AND ss.aspect='tools' AND ss.is_current;
   ```
   → `A/B Test Significance Calculator`,
   `["learn-algorithms-p-values-explained","learn-algorithms-bayesian-theory","tool-bayesian-rank"]`.
   The first two are the observed targets; the third (`tool-bayesian-rank`) has no page row, so
   `resolveToolPageURL` drops it. **The match is exact.**

**In one sentence:** when the `add_tool` work item carries no `related_pages`, the correct
behaviour is to emit no cross-links at all (`create_tool_cross_link_items.go:148` treats that as
normal, not an error) — but the resolver substitutes an unrelated suggestion's pages, and the
pipeline dutifully builds work from the substitute.

## 3. Damage, measured — and the part that is NOT damaged

- **32 cross-link items on `webdesign.co.uk`** across 9 tools, all aimed at 2 unrelated pages.
- **None of them reached live content on that site** [MEASURED]: 22 `failed`, 6 `wont_fix`,
  2 `triaged`, 2 `unresolved`, **0 `complete`**. So the cost so far is wasted writer work and a
  polluted queue — *not* wrong links on served pages. Do not upgrade this to "wrong content
  live" without re-running the status query.
- **Negative control — the mechanism is fine where the spec carries the key.** `finetuning.uk`,
  `loancash.co.uk`, `loanandmortgagecalculator.co.uk` and `vonc.com` all cross-link coherently
  (`tool-affordability-complaint-checker` → `guide-how-to-complain-and-win`;
  `tool-archetype-clash-calculator` → `archetypes`). This is what makes the webdesign.co.uk
  constancy diagnostic rather than merely odd.

## 4. The instrument UNDER-COUNTS this class — read before sizing it `[INFERRED]`

The conflict row only fires when the collected candidates **differ**. `webdesign.co.uk` has six
suggestions with differing `related_pages`, so it conflicts and is visible. **A site whose tree
holds exactly one `related_pages` — or several that agree — resolves SILENTLY**: no WARN, no
`agent_error_log` row, and the same wrong-value substitution. The 17 logged rows are therefore a
floor on this bug class, not a measure of it. I have **not** measured the silent population;
the four coherent sites above are consistent with it being small, but they do not bound it.

## 5. My own wrong turn, recorded

My first hypothesis was that **migration 211 had regressed** — I queried
`config->'params'->>'related_pages'`, got NULL for every step, and had begun writing up "the
211 wiring is missing from the only live `create_tool_component` step". **211 writes to
`config.related_pages`, not `config.params.related_pages`.** Reading the migration file before
filing is what caught it. The cheap check that would have caught it sooner: dump the whole
`config` object for the step once, rather than probing the key I expected to find.
→ logged in `WRONG_CALLS.md`.

## 6. Fix candidates, ordered by what closes the door

1. **Phase 2 refusal (RFC_029 §9 D2) — the structural fix, already scheduled.** Make a
   conflicting whole-tree search resolve to **nothing**. This case is a clean argument for it:
   the guessed value is worse than absence, and absence is already a handled, documented state
   here. This lane's step 5. **Closes the conflicting half only — see §4, it leaves the silent
   half open.**
2. **Do not fall through to the whole-tree search for a field the step explicitly wired.** If
   the config names a path for `related_pages` and that path is empty, the answer is "absent",
   not "go and find something called related_pages anywhere in the tree". This makes the bad
   state unrepresentable for *every* explicitly-wired field, silent cases included, and is the
   only candidate here that does. Blast radius is fleet-wide and needs measuring — a declared
   path that is currently empty-and-rescued-by-the-search would start resolving to nothing.
3. **Have the producer write `related_pages` into the `add_tool` spec.** Narrowest, and it
   fixes the observed site, but it leaves the substitution mechanism armed for the next
   producer that omits the key. Treat as remediation, not as the fix.
4. **Make `relatedPagesFromInputs` prefer the declared path over `GetRaw`.** Cheap and local,
   but it only un-breaks this one consumer and leaves every other optional input reading a
   guess.

## 7. How to verify a fix

```sql
-- after the fix: a tool built with no related_pages in its spec must emit NO cross-link items
SELECT split_part(item_key,':',2) AS tool, count(DISTINCT split_part(item_key,':',3)) AS pages
FROM site_work_items WHERE item_key LIKE 'tool_crosslink:%'
  AND site_id='6b49db8e-d447-4467-8277-4f3018af9897' AND created_at >= '<roll time>'
GROUP BY 1;
```
**With a demand control** — the zero is only evidence if `tool-generator` actually ran:
`SELECT count(*) FROM orchestration_states WHERE owner_agent_type='tool-generator' AND created_at >= '<roll time>';`
(it runs ~5×/24h, so allow a real window). And the negative control: the four coherent sites in
§3 must still emit their per-tool cross-links.

## 8. Ownership note

Another session filed a `needs_diagnosis` on `create_tool_component` at 2026-08-19 20:50Z — a
**different** mechanism (regeneration/replacement of an existing tool component, item
`79980f18-…`). This bug is in `create_tool_cross_link_items.go` + the resolver cascade. The
files are adjacent; coordinate before editing `create_tool_component_action.go`.
