# NOTES — `bugfix_378_usage_count_derived`

Append-only, newest at the bottom. Technical log: evidence, commands, what the system actually
said, and every misstep.

---

## 2026-08-24 — session opens, lane claimed

`bugs_open/378` was filed 2026-08-24 by the `bugs_open/351` lane and left **OPEN, not started**.

**Ownership checked before claiming.** `scripts/who-owns.py 378` returns **OWNED or recently
active**, which is the known false positive: the only commits it can see are the three that
*filed* the bug, all by the 351 lane. That lane closed itself the same morning and its handoff
(`docs024_key_docs_latest/bugfix_351_section_template_predicate/HANDOFF_2026-08-22_continue_here.md`)
says in terms: *"`usage_count` path-blindness is a different defect, filed as `bugs_open/378` …
Not a 351 residual."* Its session transcript ends with the 351 closure. No `bugfix_378*` directory
existed. So: no active thread, lane resumed here.

### Re-validating the bug at HEAD before doing anything

The bug is **still valid**, and the mechanism is unchanged at HEAD. But three of its structural
claims are now wrong or incomplete, and one of its ranked candidates is refuted outright.

#### 1. There are THREE resolution paths, not two `[MEASURED 2026-08-24, read at HEAD]`

`plan_sections_action.go`'s section loop:

| path | line | how it resolves | counts? |
|---|---|---|---|
| **Path 0** | ~1208 | the page's own stored `page_components.component_id` (`bugs_open/204`) | no |
| **Path 1** | ~1258 | direct name/function match against the `components` map | no |
| **Path 2** | ~1279 | `resolveSectionComponent` → the `section_type` selector | **yes** |

The bug file describes two. Path 0 landed with `bugs_open/204` (`d376ca9b8`, "ONE reader for a
page's stored slot identity") and is tried *first*, so it is the path a settled, rebuilt page
most often takes — the population least likely to be counted is also the most established one.

#### 2. A THIRD reader, and it decides more than a score `[MEASURED 2026-08-24]`

The bug file names the two scoring queries (`component_selector.go:181`, `:235`). It misses:

```
platform/orchestration/actions/load_existing_component_action.go:170
    ORDER BY usage_count DESC NULLS LAST, updated_at DESC
    LIMIT 1
```

That query picks the **canonical contract row** for a `section_type` — by its own comment, *"the
row the store will overwrite and enforce"*. So the biased column is not only a 0.1 nudge in a
score; it is the primary sort key deciding which component *is* the contract. Higher stakes than
anything the bug file records.

Full census of the column, repo-wide, `[MEASURED 2026-08-24]`:
- **Writers: 2.** `IncrementUsageCount` (`component_selector.go:133`, the only incrementer, one
  non-test call site at `plan_sections_action.go:1957`), and the birth INSERT at
  `store_generated_component_action.go:639` which writes the literal `0`.
- **Readers: 3**, all filtering `component_level='section'`.
- No DB trigger, function or view touches it (`pg_proc`/`pg_trigger` checked; the only trigger on
  `content_components` is `trg_cc_refuse_null_section_type`, migration 581 / CLC-029 from the 351
  lane). The three `agent_definitions` config hits are prompt text listing `060`'s filename, not
  queries. So the bug file's `[INFERRED]` "no third writer" **holds** — now measured, not inferred.

#### 3. The counter OVER-counts as well as under-counting — and this refutes fix candidate 2

This is the finding the bug file does not have, and it changes the fix.

`IncrementUsageCount` is called inside `resolveSectionComponent` **before** `planSection` returns
and therefore before the section's status (`ready` / `deferred` / `skipped`) is known, and before
any `page_components` row is written. It also fires again on every re-plan of the same page. So
the column counts **resolution attempts on one path**, not usages.

Proof at the artefact, not from the code alone `[MEASURED 2026-08-24]`:

```sql
SELECT name, usage_count,
       (SELECT count(*) FROM page_components pc WHERE pc.component_id=cc.id) AS live_bindings
FROM content_components cc
WHERE is_active AND forked_from IS NULL AND component_level='section' AND COALESCE(usage_count,0)>0
ORDER BY usage_count DESC;
```

| name | usage_count | live bindings |
|---|---|---|
| `bayesian-ranking-hero-tool_pre_037` | **20** | **0** |
| `case-studies-grid` | 19 | 4 |
| `testimonials-modern` | **12** | **0** |
| `contact-block` | 7 | 6 |

**The two largest values in the column are both non-usages.** `testimonials-modern`
(`202fe796-3a39-449f-9dbb-697726fbf5f1`) was **created 2026-08-23** and has **zero**
`page_components` rows *ever* — checked live, checked with `build_status='removed'` included, and
checked against the `page_components_bak_20260820_preplan_lmc` snapshot. Twelve counts, no binding,
in one day. `bayesian-ranking-hero-tool_pre_037` is a pre-migration-037 **backup copy** that is
still `is_active` and still a live selector candidate, carrying the highest count in the table.

**So the bug file's candidate 2 — "call `IncrementUsageCount` on Path 1 too" — is refuted, not
merely weak.** It would propagate a definition that is already wrong in both directions to more
paths. Any fix that keeps this increment keeps the overcount.

#### 4. The dead half is bigger than the bug file measured `[MEASURED 2026-08-24]`

Census by level (active, non-forked). The bug file censused `section` only.

| level | rows | with any count | bound but zero | bindings invisible to the term |
|---|---|---|---|---|
| section | 151 | 12 | 98 | 1,865 |
| **tool** | **115** | **0** | 104 | 107 |
| header | 4 | 0 | 0 | 0 |
| footer | 1 | 0 | 0 | 0 |
| site | 5 | 0 | 0 | 0 |
| element | 1 | 0 | 0 | 0 |

`tool` is **0 of 115** — that half is `bugs_closed/060`'s dead counter exactly, not 378's
half-written one. It has no reader today (all three readers filter to `section`), which is the only
reason it is harmless.

#### 5. The scoring term is provably INERT today — and the control proves the measurement had power

The bug file's central `[UNMEASURED]`: *"whether the 0.1 term has ever actually flipped a
selection."* Run it.

Simulating the real scoring expression over every `(section_type, site_type, page_type)` context
drawn from the live vocabulary, comparing `argmax(base + usage_term)` against `argmax(base)`, ties
broken identically by `id` in both so a tie cannot masquerade as a flip:

```
contested_contexts | flipped | flipped_section_types | contested_section_types
       4888        |    0    |          0            |           4
```

**Zero flips.** But a zero is only evidence if it could have come out otherwise, so the same query
was re-run with a counterfactual granting the *weakest-base* candidate in each group a full 50 uses
(the maximum the term allows):

```
contexts | flips_detectable | section_types_flippable
  4888   |       52         |           1
```

The instrument detects flips. The zero is real. **The reason is mechanical:** only **4**
`section_type`s have more than one active non-forked section candidate — `hero` (7),
`tool-archetype-taster-quiz` (2), `tool-gripper-payload-calculator` (2), `features` (2) — and
**every candidate in all four carries `usage_count = 0`**. The 12 components that *do* carry a
count are all the sole candidate for their own `section_type`, so they win regardless.

**Consequence, and it sets the schedule:** a change that removes or replaces the term **cannot
regress any current selection**. That is a closing window — it holds only while contests stay rare
and uncounted.

#### 6. A maintained counter has seven places to be forgotten `[MEASURED 2026-08-24]`

`INSERT INTO page_components` sites outside tests:

```
platform/orchestration/actions/save_page_sections_action.go:1074
platform/orchestration/actions/deploy_tool_action.go:510
platform/orchestration/actions/create_tool_component_action.go:526
platform/orchestration/actions/create_report_page_action.go:268
platform/orchestration/actions/rebuild_blog_listing_action.go:389
platform/orchestration/actions/adopt_verbatim.go:531
cmd/webdesignport/import.go:239
```

Six in `platform/` plus one command. This is the `RFC_008` trap in memory almost exactly — that
census found ten writers of `page_components.rendered_html` and an eleventh was *born* two weeks
later. **So "maintain the counter where the binding is written" is the wrong shape**: it is seven
chances to forget and an eighth tomorrow. Derive at read; there is then no counter to drift.

#### 7. Prior art the bug file missed

`docs/agent_docs/sql_for_agents/325_directory_listing_binds_to_business_directory_query.sql:8-9`,
dated **2026-08-06** — eighteen days before 378 was filed:

> *"`usage_count=0` and, more reliably (usage_count is not a trustworthy signal fleet-wide —
> checked separately), zero pages anywhere carry this component"*

Someone had already established the column was untrustworthy and had already stopped relying on it
— and recorded it in a **migration comment**, where no grep of `bugs_open/`, the register or 016b
would ever surface it. Worth noting as a routing failure, not a fault of that author.

#### 8. Cross-thread: this is upstream of `bugs_open/107`

`bugs_open/107` — *every site gets the same homepage skeleton* — cites `component_selector.go:176`,
the neighbouring scoring term. The usage term is a **preferential-attachment loop**: selected →
count rises → scores higher → selected again. It is inert today only because contests are rare
(finding 5). **Any 107 remedy that adds candidate variety per `section_type` immediately activates
a term that systematically favours the incumbent** — so 107's fix would be partly undone at the
moment it starts working. 378 should land before or with it. Raised with that lane.

### Method note

`090` diagnosis run filed: intake `d2072b88-67ac-48ca-a319-6be6d543aae7`, run correlation
**`1c62c1f7-cb06-4482-933e-8c08a622b5c1`**. Framed at `component_selector.go` (13.7KB) and
`load_existing_component_action.go` (15.7KB) rather than `plan_sections_action.go` (114KB),
per the LANDMINE that a `090` on a symbol in a file over ~60KB returns bundles and no verdict.
