# NOTES — `bugs_open/357`, component identity vs stored bytes

Append-only, newest at the bottom. Technical log: evidence, commands, what the system said,
and every misstep.

---

## 2026-08-22 — session start: 277 was the ask, 357 is the work

Asked to take up `bugs_open/277`. **It is not open** — it moved to `bugs_closed/` earlier the same
day (`4a193b390`, owner ruling). Verified the close rather than trusting the file:

```sql
SELECT COALESCE(result->>'route','(UNROUTED)'), status, count(*)
FROM site_work_items WHERE item_type='required_fields_missing' GROUP BY 1,2;
--  no_content_data/needs_human_review 27 · asset_sourced 2 · no_plan_owned 1 · (UNROUTED)/complete 37
```
Matches 277's stated close exactly, and its served-bytes proof still passes today
(cubic-bezier `<code>ease-in-out</code>` = 1; gas-unit-converter `<tr` = 6).
Owner picked the named successor: **`bugs_open/357`**.

## The population is 22, not 9 — and 357's own query says so

357 states nine rows [MEASURED 2026-08-22 ~10:00Z] and prints a "re-runnable" population query.
**Run today, that query returns 22.** The nine were the subset that also had a parked
`required_fields_missing` work item (all of them one-component pages); the query as written is
unrestricted and picks up multi-component pages too. Newest row: **`vetcomparison.uk` `index`,
created 2026-08-22** — a site homepage, born today. So this is a live producer, not a June artefact.

⚠ **The file's stated population and the file's own query disagree.** Anyone re-measuring 357 from
its query and comparing to its table will think 13 rows appeared overnight. They did not.

## Which writer — closed by fingerprint, which the 090 run never tried [MEASURED 2026-08-22]

`63d4d1a7` came back UNVERIFIABLE with "which writer assigns the component_id" open. The writers
leave distinguishable marks on the row, so the row itself answers it:

| writer | position | content_brief | build_status |
|---|---|---|---|
| `save_page_sections_action.go` | `i+1` | **written** (`{slot} section`) | deployed |
| `deploy_tool_action.go` | 2 | never | deployed |
| `create_tool_component_action.go` | 2 | never | deployed |
| `adopt_verbatim.go` | 0 | never | approved |

**All 22 rows: `position=1`, `content_brief.section_guidance='hero section'`.** Only
`save_page_sections_action.go` writes `content_brief` at all. The writer is settled.

⚠ **A near-misstep, recorded because I nearly filed it as a second discriminator.** The nine older
rows have `rendered_html_digest` NULL and the thirteen newer ones have it set — which reads like two
writers. It is not: the INSERT writes `md5($3)` unconditionally, and the column postdates the older
rows (`bugs_open/229` / IMP-052). **The split is by DATE, not by writer.** A digest-based
"two writers" claim would have been confidently wrong.

## Root cause, in committed code, cited

1. `saveSectionsExtractFromHTML` (`save_page_sections_action.go`) — tool/game pages emit
   `<div class="tool-page">…</div>` with no `<section>` element, so the section regex matches
   nothing. The documented fallback stores the whole fragment as ONE section named `"section"` —
   the sentinel for *identity unknown*. Its own comment says it stays `"section"` "unless a planned
   name exists".
2. `enrichSectionsWithPlannedNames` — treats `"section"` as needing enrichment and assigns
   **`planned[Position-1]`** from `pages.sections`. Purely positional.
3. `enrichSectionsWithComponentIDs` — resolves that name to the shared component's UUID.
4. The single INSERT persists `slot_name='hero'`, `component_id=<hero>`,
   `rendered_html=<the whole tool>`.

**`hero` is planned FIRST on all 22 pages** (`["hero","generic-text-block"]`,
`["hero","tool-list"]`, `["hero","info-card-grid","latest-news","call-to-action"]`), so
`planned[0]` is always `hero`. The tool is called a hero because hero is first in the list —
**the sentinel meaning "I do not know what this is" is converted into a confident wrong answer.**

## What 357 does not record, and what makes it worse than filed

- **13 of the 22 carry a complete hero `content_data`** (`headline`, `subheadline`,
  `background_image`, `hero_url`, `cta_text`, `secondary_cta`) alongside tool HTML. 357's nine all
  have `content_data` NULL, which is why its "the parked state is the protection" argument holds
  for them — **it does not hold for these thirteen.** `ContentDataCanFillTemplate` is already true
  on them, so the regeneration 357 warns a backfill would arm **is already armed**, with no backfill.
- **16 of the 22 are `rebuild_policy='generic'`** — rebuildable, not owned.
- The rows were **born** wrong: `updated_at = created_at` on all 22. No later overwrite.
- `vetcomparison.uk/index.html` serves `<div class="tool-page">` at the hero position and
  `data-component="hero"` **zero** times [MEASURED 2026-08-22, http 200, 44,496 bytes]. The hero
  copy its writer produced is stored and never rendered.

## The discriminator for a guard — measured, not guessed

357's fix candidate 2 warns a born-wrong guard "must be measured against the whole live table
before arming" because the static-template-prefix test also fires on legitimately drifted templates.
Correct: that test flags **158** rows fleet-wide. So test the component's own self-declaration
instead — `data-component="…"`, which `saveSectionsExtractFromHTML` already trusts as the identity
carrier in this same file:

```sql
tmpl_attr := substring(cc.html_template from 'data-component="([^"{]+)"')
html_attr := substring(pc.rendered_html from 'data-component="([^"]+)"')
-- flag when tmpl_attr IS NOT NULL AND html_attr IS DISTINCT FROM tmpl_attr
```

| | rows |
|---|---|
| components declaring the attribute | 149 of 339 |
| template and HTML **agree** | **1,550** |
| template and HTML **disagree** (both declared) | **0** |
| template declares, stored HTML silent | **27** |
| static-prefix mismatches (the noisy test) | 158 |

**Zero false positives fleet-wide, and the 1,550 agreements are the demand control** — the
predicate could have come out otherwise, 1,550 times, and did. The 27 are a strict subset of the
158: the 22 hero rows, 3 `loancash.co.uk` Ported Page rows, one `leopardessconsulting.co.uk`
`blog-listing_pre_037`, and one `idea.uk` `tool-list` row whose `rendered_html` is **zero bytes**.
The ~131 drift-only rows the noisy test flags are **not** in it.

⚠ **Stated as a limit, not glossed:** this predicate is silent for the 190 components whose
template declares no `data-component` at all, and 97 of the 100 `Ported Page` prefix-mismatches are
NOT flagged. It is a **refuse-on-certainty** test, not a census of everything wrong. That is the
right trade at a write seam and the wrong one for a sweep.

## [UNMEASURED] — whether the destruction has already fired

A page born this way, then rebuilt from its hero `content_data`, would serve a title band and
**leave the population query**, so the survivors cannot tell me about the casualties. There is no
systematic `page_components` history to check (only ad-hoc `_backup_*` tables). Candidates exist —
`mortgagecalculator.co.uk` `game-fact-finder` is `page_type='game'` with 2 components, 4,363 bytes
and no `<script>` — but "never had a tool" and "had one and lost it" are indistinguishable from
current state alone. **Not claimed in either direction.**

## Diagnosis loop

Re-filed sharper (357 itself says the first run stalled because the symptom asked one question
while pointing at three files): intake `f7aedef7-0bee-4c68-8cde-c86ac552e3e2`,
**`RUN_CORRELATION_ID=e580b34a-d284-4f80-ac96-81af1c4adaba`**. It is asked the one thing the row
fingerprints cannot settle: which leg pairs a genuine hero `content_data` with the tool's bytes,
since the `saveSectionsExtractFromHTML` fallback sets no `ContentData` at all.
