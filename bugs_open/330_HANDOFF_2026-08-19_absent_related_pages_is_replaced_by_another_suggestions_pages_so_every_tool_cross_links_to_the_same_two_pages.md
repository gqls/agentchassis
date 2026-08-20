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
**VERDICT: CONFIRMED** (2026-08-19 21:13Z, first iteration, `stopped_by: confirmed`). It read the
one link §2.3 had *inferred* rather than opened, and the citation is worth keeping verbatim
because it names the exact predicate:

> Only fields Strategy 0 actually resolved (present in `result.Values` and **not merely
> `Defaulted`**) are excluded from what `ExtractFields`' whole-tree search is asked for; **a field
> whose declared dot-path yields nothing is never marked resolved**, so it is handed to
> `ExtractFields`/`extractSingleField` under its bare name exactly as an undeclared field would be.

So the defect is in `ExtractActionInputs`' `strategy0Resolved` bookkeeping, one layer above the
cascade — "I tried and found nothing" and "I was never asked" are recorded identically. The loop
explicitly **declined** to answer the scope-widening half (which other live steps declare a path
now supplied by the search) — it is marked `[context]`, "a scope-widening audit request … not an
observed instance", pursued via a data_request rather than asserted. **That audit is still owed
and is the sizing question for fix candidate 2.**

Bonus confirmation from the loop's own evidence bundle: at 20:56:43Z the *real*
`tool-ab-test-calculator` was built and cross-linked to those same two pages — and for THAT tool
they are **correct**, because it IS `suggestions[0]`. One tool's right answer is being handed to
nine others.

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

## 9. ADDENDUM 2026-08-20 — candidate 2 SIZED (the audit §6.2 called for), and §4's silent population now has a measured floor

By the staged_component_build lane (the owner of the resolver-scope half, per §6.2's own note).
Method and caveats in the lane RUNBOOK ("Sizing the wired-but-missing rescue population");
scripts regenerate it from live data.

**Static surface (fleet-wide, live definitions, recursive walk incl. sub-workflows):**
**451 plain Strategy-0 wires** — a spec field (Required∪Optional) whose step config carries a
dotted-path value — across **309 (agent, field) pairs on 83 agents**; only 3 wires are `!`-strict.
That is the full surface whose behaviour candidate 2 could change.

**Runtime sample (the 8 highest-demand agents, last ≤12 runs each, 40 non-array-indexed wires
evaluated against FINAL collected_data):** 30 wires resolve on every sampled run; **10 are
RESCUE-PRONE** — the wired path missed AND the field name exists elsewhere in the tree, so the
whole-tree search is (upper bound) supplying the value today:

| agent | field ← wired path | miss rate |
|---|---|---|
| page-build-handler | `mode` ← input_data.spec.mode | **12/12** |
| page-build-handler | `section_facts` ← spec_sections.section_facts | 11/12 |
| page-build-handler | `page_id` ← page_record.id · `page_name` ← page_record.name · `page_sections_fallback` ← page_record.sections · `sections` ← spec_sections.sections | 7/12 each |
| page-build-handler | `page_id` ← input_data.spec.page_id · `authoritative_page_id` ← input_data.page_id | 1/12 each |
| page-rerender | `reason` ← input_data.spec.reason | **12/12** |
| tool-generator | `related_pages` ← input_data.spec.related_pages | 7/12 (**this bug**) |

build-dispatch-loop, page-content-writer, component-creator, tool-deployer: all wires resolved
on every sampled run. rerender-pages: 0 recent runs — unsampled, not clean.

**Caveats, load-bearing:** (a) "rescueable" is an UPPER bound — the LIKE probe does not honour
the search's infrastructure-key exclusions (e.g. `retry_payload` is skipped since 306 cand 3),
and finding the name is not finding a usable value; (b) sampled against the FINAL tree — a path
filled after the reading step counts as resolved here but was absent at step time; (c) 34
array-indexed wires were skipped; (d) the other 75 agents (269 pairs) are UNSAMPLED — this is
the high-demand slice, not the fleet.

**What it means for candidate 2:** flipping "wired-but-missing → absent" fleet-wide would
change behaviour at 10 live wire-sites immediately — including pbh's `mode` (every run) and the
`page_record.*` family, where the rescue is very likely finding the RIGHT value from a sibling
envelope (two wires for the same field at different steps; when one misses the search finds the
other's source). So candidate 2 needs one of: (i) per-field fallback CHAINS in config (ordered
paths — the shape `renderEnvelopeIdentity` uses in Go), so the legitimate dual-envelope cases
are declared rather than rescued; or (ii) config repairs on the 10 wires first, then the flip,
then this audit re-run fleet-wide as the gate. Either way the pbh envelope family is the first
work item, not tool-generator: it is 8 of the 10 wires and 56 runs/24 h of demand.

> **CORRECTED 2026-08-20 ~10:3xZ, same session, ~1 h after §9 — the "10 RESCUE-PRONE / needs
> fallback chains" conclusion was an artefact of my own probe.** The LIKE test encoded "the
> field name appears anywhere in the RAW tree", but the search skips `agent_config`,
> `__raw_message__` (and siblings) and `retry_payload` (`isInfrastructureKey`,
> `unified_extractor.go:720`) — and enumerating actual paths on a missing run showed most hits
> lived exactly there. Re-run with those subtrees stripped, same 12-run samples:
>
> | wire | miss | genuinely rescueable |
> |---|---|---|
> | pbh `page_id` ← page_record.id | 6/12 | **6** (run's own input: input_data.page_id / current_page.page_id / spec.page_id all present 6/6 — right value, dual-envelope) |
> | pbh `page_name` ← page_record.name | 6/12 | **6** (same family) |
> | tool-generator `related_pages` | 8/12 | **8** (THIS bug — the rescue is the wrong value) |
> | page-rerender `reason` | 12/12 | **1** |
> | pbh mode / sections / section_facts / page_sections_fallback / authoritative_page_id / page_id←spec | 1–12/12 | **0 — clean absences** |
>
> **So candidate 2's real behaviour-change population on the high-demand agents is FOUR wires,
> not ten, and it decomposes cleanly:** (a) pbh page_id+page_name — declare the fallback the
> search performs today (one config edit on the reading step, e.g. a `?`-optional wire at
> input_data.page_id, or re-point; the value is proven present in-run 6/6); (b) related_pages —
> absence is the DESIRED outcome (this bug's whole point); (c) reason — 1/12, trace that one
> run before deciding. **"Needs per-field fallback chains" is WITHDRAWN as a general
> requirement** — with (a) repaired, candidate 2 is close to a straight flip on this slice.
> The 75-agent / 269-pair unsampled remainder still needs the (corrected) probe before a
> fleet-wide flip; the RUNBOOK method now includes the strip.
> What caught it: enumerating real paths on one missing run instead of trusting the LIKE —
> "your measurement answers the question you ENCODED, not the one you asked."

> **Sharpening, same session (~10:5xZ):** the pbh pair's rescue candidates AGREE (one page,
> several paths — 19/19 name, 19/19 id where both exist, 0 genuine disagreements over 40 runs;
> the record misses 21/40 and the input then carries the identity 6/6). Agreeing candidates are
> never refused by RFC_029's step-5 flip, so **the pbh wires do not block step 5** — they are in
> scope ONLY for THIS bug's candidate 2, where the options are a two-element declared chain or
> a safe-by-inspection note (now fully writable from the measurement above). Candidate 2's
> wrong-value population on the sampled high-demand slice is exactly one wire: this bug's.
