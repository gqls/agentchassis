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

---

## 2026-08-22 (later) — ⚠ CORRECTION: I published "already armed" and it is REFUTED. The rerender is the AUTHOR, not the threat.

Recorded here in full because the correction is more useful than the finding it replaces.

**What I wrote** (in `bugs_open/357` §3, and said to the owner): the 13 rows with a complete hero
`content_data` are one rebuild away from having the tool replaced by a title band.

**What refutes it.** `vetcomparison.uk` `index` has **six completed `page_rerender_index_…` items
between 08-19 and 08-22** and the tool still serves. Then the timestamps:

```
rerender item   created 2026-08-22 08:44:51 → completed 08:50:19
all 4 page_components rows on the page       created 08:50:12
```

The rows were written **inside** the rerender window. `RerenderPageSectionsAction` DELETEs and
re-INSERTs; it carried the tool HTML forward and re-attached the hero identity and hero
`content_data`. **The rerender re-mints the mismatch on every pass.**

**Three failures, worth separating:**

1. **Imported a mechanism from an adjacent bug without testing its reach.** `ContentDataCanFillTemplate`
   is 277's arming mechanism. Its only non-test caller is
   `discovery_checks/check_literal_markdown.go:429` — a **detector's classifier**, not a rebuild
   gate. I ran that grep, saw "one caller", and went looking for the rebuild path anyway with the
   conclusion already formed.
2. **Predicted a codepath's behaviour instead of looking for its output.** The prediction was
   defensible from source. The system had already run the experiment six times and written the
   answer into `site_work_items`.
3. **Read `updated_at = created_at` as history.** I had it in these NOTES as "born wrong, never
   touched since". It means the opposite — this writer re-inserts rather than updates, so every row
   is *newly* born, repeatedly. **A column equal to its sibling is evidence about the WRITER's
   method, not about stillness.**

**What survives unchanged:** the identity/bytes disagreement on 22 rows; `save_page_sections` as the
writer; hero planned first on all 22; the positional enrichment in `enrichSectionsWithPlannedNames`;
the `data-component` predicate and its 1,550/0/27 census.

**What this changes about the fix:** the population is **self-renewing**, so stock repair without a
flow fix is wasted — the next rerender reproduces it. Flow before stock, and any repair needs a
re-check *after* a subsequent rerender, not just after the repair.

**STILL OPEN, assumed in neither direction:** whether a full page rebuild *through the writer*
(rather than a rerender) renders the hero template over the tool. Rerender demonstrably does not.

**Also established while checking this** (`rerender_page_sections_action.go`): the content pre-check
at ~380–445 exempts self-contained tools via `isSelfContainedSection` (~1270), keyed on
`component_level=='tool'` **and** an empty `input_schema`. `hero` is `component_level='section'`
**with** a schema [MEASURED], so the exemption never fires for these rows — they are processed as
ordinary sections end to end. That is the seam where a tool-shaped payload could be recognised and
is not.

## 2026-08-22 (later still) — the persistence mechanism, and the general form of the defect

`carryStoredSection` (`rerender_page_sections_action.go:1150-1169`) forwards the stored row as one
bundle: `rendered_html`, `stored_slot_name`, `component_id` and `content_data`. On the receiving
side `extractSectionsFromMetadata` prefers `stored_slot_name` over everything else when naming the
section. **Identity, bytes and content are propagated together, verbatim, and nothing compares
them.**

That the vetcomparison row was CARRIED rather than rendered is not an assumption: the hero template
opens `<section class="hero" data-component="hero"`, and the stored 11,326 bytes contain no
`data-component` at all, so they cannot be that template's output. Only a carry branch emits stored
HTML. **Which** carry branch (component unresolvable / section not ready / empty template) is not
established — three reach `carryStoredSection` and the logs for that run have rotated.

### The general form, which is the part worth fixing

> A component row's **identity** (`component_id`, `slot_name`), its **bytes** (`rendered_html`) and
> its **content** (`content_data`) are written, carried and re-written as an atomic bundle by every
> path in the page-composition system — and **no seam anywhere asserts that the three agree**.

Two independent consequences, and the second is why this is not a data-repair job:

1. **A pairing that is wrong once is wrong for ever.** The carry preserves it faithfully.
2. **It is refreshed on every pass**, so the population is self-renewing and cannot be repaired
   ahead of the producer.

Where identity is *invented* rather than carried, it is invented **positionally**
(`enrichSectionsWithPlannedNames`: `planned[Position-1]`), from a plan, without looking at the
bytes — which is how a tool comes to be called a hero in the first place. Origin and persistence are
therefore two different code paths, and a fix that addresses only one leaves the class alive:
fixing the origin leaves 22 rows re-minting for ever; fixing the carry leaves the next tool page
mislabelled at birth.

**The one place both converge** is `save_page_sections_action.go`'s single
`INSERT INTO page_components` (~line 999) — documented in-code as *"the single INSERT every
page-composition path flows through"*, and already carrying a precedent guard
(`sectionIsUnresolvableStub`, `bugs_open/039`) that refuses a bad row, raises a typed work item and
continues. That is the seam, and the precedent is the shape.

---

## 2026-08-22 — COUNCIL ROUND 1: **REVISE**, and checking the gating objection found a defect in my plan worse than the objection

Trail `62aac6c2-996f-4b5d-8f8f-72e3daf4c82e`. 13 reviewers, 4 abstained, gated by `bug_historian`.

### The finding that resized the submission — mine, not theirs, but theirs is why I looked

`bug_historian` (HIGH, edit 1) asked whether the honest-unknown degrade path had been tested against
the Layer 2 carry-forward, reasoning that if carry-forward decides protection *by inspecting the
component/template*, a re-typed row might stop qualifying.

**Its stated mechanism is wrong**, and that is checkable in one read: Layer 2's PROTECTION keys
purely on the bytes — `interactiveHTMLSQL(col)` is `ILIKE` over markers on `rendered_html` (:1698),
`sectionHTMLIsInteractive(html)` is `strings.Contains` over the same markers (:1677). Neither reads
`component_id`, `slot_name` or a template. Re-typing cannot remove Layer 2 protection.

**But the question was the right one**, because Layer 2's **MATCHING** is:

```go
for i := range sections {
    if sections[i].ComponentName == p.slot { matchedIdx = i; break }   // :517 — slot NAME, and nothing else
}
```

Rename the stored slot to `adopted-fragment` while an incoming plan-driven section set still names
`hero`, and the match fails. Control falls to `default:` — *"Slot dropped entirely — re-append the
tool so it survives"* — which **appends** the tool as an extra section while the incoming hero title
band is **also** saved. **The page gains a hero band and the tool moves position.**

> **So my "byte-preserving re-type" would have changed what 22 live pages serve**, on four sites,
> as a side effect of a change I described in the submission as preserving bytes exactly. It needs
> `pages.sections` updated in the same transaction, or identity-agnostic matching — and either is
> its own round with its own measurement.

That is the second time today the estate's protective machinery has been the thing I mis-modelled
(the first is the correction above). **The pattern to carry: a mechanism that preserves BYTES may
still key its decisions on IDENTITY, and those are different questions.**

### Round 2: withdrawn, kept, and answered

**Withdrawn** — Layer A (fallback typing), the `adopted-fragment` seed, the positional-enrichment
narrowing. All three change what is persisted, and all three are blocked on the slot-matching
question above. This also answers `guardian`'s three objections (unconditional changes, three
simultaneous change vectors on the busiest save path) without argument.

**Kept** — the seam guard (record-only; refusal behind an opt-in seeded on nobody), the detector, the
register entry, the tests. None changes any persisted output.

**Objections answered by checking:**

| seat | objection | what the check said |
|---|---|---|
| `prior_art` | is `adopted-fragment` dormant-machinery duplication? | **No.** `SELECT name FROM content_components WHERE regexp_replace(html_template,'\s','','g') ~ '^\{\{\.?[A-Za-z_]+\}\}$'` → **0 rows**. No byte-preserving passthrough exists |
| `prior_art`, `reuse_agent` | is the claimed inline duplication real? | **Yes** — `dataComponentRe := regexp.MustCompile(...)` at **:1392 AND :1498**, byte-identical, used at :1427/:1463/:1509. Both deleted and repointed; `grep -c 'dataComponentRe :='` must read 0 after |
| `editquality` | does `ON CONFLICT` protect idempotency? | **Yes** — `content_components_name_key` UNIQUE on `(name)` exists |
| `architecture` | state the optional-key count before adding the Nth key | **It cannot be read, and that IS the finding.** The audit covers 123 actions against N=10 and lists **99 more as `uncounted`** — `save_page_sections` is one of them (no `RegisterActionInputSpec`), so a new key enters counted as ZERO. The `retract_asset_files`/`publish_site` trap exactly |
| `tooling_provenance` | do not hand-edit the concept index count | **Correct.** Stored counts are RETIRED; the checker reads only the first 4,000 bytes and its own line says *"any that reappears is the finding"*. CLAUDE.md's "update the index count" is the anti-pattern here. Row added, headline untouched, two-way `comm` parity run instead |
| `bug_historian` | Layer D is detection-only at five other writers | **Accepted as residual exposure, not closed.** Named in the submission |

Round 2 dispatched on the same trail (`RESUBMIT_CORR`), run envelope `3ad2cff2`.
