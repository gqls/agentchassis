# BUG 156 (2026-07-30) — a page can hold content-identical duplicate slots, render itself twice, and nothing in the platform notices

**Status:** OPEN — narrowed 2026-07-31 to a **PREVENTION gap**. The vonc data
instance is **FIXED AND VERIFIED LIVE** (2026-07-30, owner-approved; see § "The live
instance"). **The DETECTION half is no longer this bug's** — the `151` lane built
`check_content_duplication` + a repair handler on 2026-07-31, deliberately held
inert; see the correction under point 3. What remains, and is unowned: **nothing
stops the state being created.** No uniqueness constraint (the only unique index on
`page_components` is the `id` pkey, re-verified 2026-07-31), no guard in the save
compares the incoming section set against itself (re-verified — the two `dedup`
mentions in that file are about `needs_new_component` items and an `idx_swi_dedup`
comment), and `page_components.content_hash` is still never written by any code path.

**Diagnosis-loop record (owner ruling 2026-07-31).** This file asserts a structural
root cause, so it was put through `090` — intake `6b09dbde`, run correlation
`8e5594a4-0a15-40eb-bdd9-636d8849dff5`. **The run FAILED and produced no graded
verdict**: `call_diagnoser` timed out after 3 retries over ~40 minutes (the known
spawn→call handshake class). It is recorded as attempted-and-failed rather than
skipped, and rather than claimed as a CONFIRMED. It was **not** wasted — it wrote
five evidence bundles first, and reading those by hand is what corrected point 3 and
surfaced `v3_site_actions.go:sameSectionList`, a second comparison function I had
never read (it compares *proposed vs realised* positionally, so a doubled 12-entry
list matched against 12 realised rows returns `true` — it does not refute the
prevention claim, it is another instance of the same pattern: every comparison in
this area is against what already exists, never within the incoming set).

**Found by:** running the `bugs_open/151` dedup census against vonc
(`docs/agent_docs/docs024_key_docs_latest/gauntlet_dead_cta/HANDOFF_2026-07-30_D_vonc_says_the_same_things_twice.md`,
Finding 1). **This is NOT the 151 mechanism** — 151 is the section writer having
no memory of facts already used, producing *near-duplicate* copy across siblings.
This is *byte-identical* rows on one page, a different cause and a different fix.
151's vonc instance was contributed separately in `55fe5eaf9`.

## Symptom

`vonc.com/about.html` has 12 `page_components` rows that are 6 identical pairs, so
a visitor reads the entire about page twice over. Live and visible; re-verified in
the served HTML 2026-07-30 (`A provocation, not a prompt` ×2, `The Gauntlet is
open` ×2 in a 90,220-byte response, `http_code=200`).

```
 position |        slot_name        | html_md5 |  cd_md5  | locked_by
        1 | hero-about              | 4b18a2d0 | d03273d1 |
        2 | hero-about              | 4b18a2d0 | d03273d1 |
        3 | content-block-about     | 80d26c85 | 59dd020d |
        4 | content-block-about     | 80d26c85 | 59dd020d |
        5 | game-master-explanation | a330c5ba | 8e723976 |
        6 | game-master-explanation | a330c5ba | 8e723976 |
        7 | platform-comparison     | 7e783dc3 | 0444dc9b |
        8 | platform-comparison     | 7e783dc3 | 0444dc9b |
        9 | differentiators         | 7087e491 | 6bffa737 |
       10 | differentiators         | 7087e491 | 6bffa737 |
       11 | gauntlet-cta            | c86b7007 | f76a6bd3 |
       12 | gauntlet-cta            | c86b7007 | f76a6bd3 |
```

Each pair matches on `rendered_html`, on `content_data` **and** on `component_id`.
No locks, no `parent_instance_id`/`content_item_id`/`research_id` references.
`page_id = a28abcd7-186b-4a33-9b89-5d7bfd727012`, site `9ec3b9ee-5b08-461b-b4f8-9e1e03579c74`.

## What is measured, and what is NOT

**Every persisted source of truth says 6.** The doubling exists only in
`page_components`:

| source | count | verdict |
|---|---|---|
| `site_plan_sections` (plan `77493277-…`, `is_current`) | 6 | correct, right order |
| `pages.sections` (jsonb) | 6 | correct, right order |
| `pages` rows named `about` | 1 | not a duplicated page |
| `page_components` | **12** | the only doubled thing |

**It was one sequential write pass, not two builds.** All 12 rows were created
between `15:12:50.37849+00` and `15:12:50.471673+00` on 2026-07-28 — a 93 ms
window — with `created_at` strictly increasing with `position`, and `positions
1..12 distinct`.

**That last detail rules out the obvious race.** `save_page_sections_action.go:538`
does `DELETE FROM page_components WHERE page_id = $1 AND <agent-writable>` and then
inserts at `position = i+1` (`:657`). Two concurrent saves interleaving
delete-delete-insert-insert would each number their own rows 1..6, giving **two
rows at each position 1–6** — not 12 sequential positions. So this was a single
loop over an input list that **already contained 12 entries, each planned section
duplicated adjacently** (1,1,2,2,3,3 — a re-run of the whole loop would have given
1,2,3,4,5,6,1,2,3,4,5,6).

**[UNRECOVERABLE] The producer of that 12-entry list is not identifiable from
retained data.** The transient list lived in the run's `collected_data`.
`orchestration_states` and `site_work_items` for that window have both aged out
(queried both; 0 rows). Narrowed to, and no further:

- `save_page_sections_action.go` did not create it — its two extraction paths are
  mutually exclusive (`:190` `if len(sections) == 0` guards the HTML fallback), so
  they cannot concatenate.
- `CompilePageSectionsAction` (`v3_site_actions.go:1941-1990`) did not create it —
  every branch appends exactly once per input item.
- `loop_actions.go:510` appends exactly once per iteration, so the loop's **item
  list** is the last place the doubling can have entered.

Do not write a root cause into a handoff on this evidence. The adjacency signature
means each *iteration* emitted its section twice, which is where a fixing thread
should start — but that is a lead, not a finding.

## The durable defect: nothing detects or prevents it

1. **No constraint.** `page_components` has no unique index on
   `(page_id, slot_name)` — 9 indexes, none unique but the `id` pkey; 8
   constraints, none a relevant unique.
2. **No guard in the save.** `save_page_sections_action.go` has five refusing
   guards (ownership, content-regression, interactivity, locked-slot, claims
   floor — see LANDMINES) and **none of them compares sections to each other.**
   A 12-entry list where the plan says 6 passes every one.
3. ~~**No detector anywhere.** There are 60+ discovery checks; none looks for
   duplicate rows. `grep -rn "HAVING count(\*) > 1" platform/ internal/ pkg/
   scripts/` returns **nothing fleet-wide**.~~ **OBSOLETE — see the correction
   below.** `content_hash` is EMPTY on all 12 rows, so even the column that could
   have flagged identity is unpopulated (that half still stands).

> **CORRECTED 2026-07-31 — point 3 is obsolete, and the probe behind it was never
> sound.** A detector now exists: `discovery_checks/check_content_duplication.go`
> (`ec8ad7959`, **2026-07-31 09:08** — 12½ hours after this file was written), built
> by the `151` lane as its candidate 3, with a deterministic repair handler
> (`remove_duplicate_page_sections_action.go`). Its `findIdenticalSamePage` groups a
> page's sections by `datahelpers.SectionIdentityKey(slot_name, raw)` — **same slot,
> byte-identical blob, per page**, which is exactly the shape described in this file,
> so it would have caught the vonc instance. It also independently reached this
> file's central fix constraint: identity is the **blob**, not the prose
> (`43492ec94`).
>
> **The claim was true when filed and the grep still could not have established it.**
> `HAVING count(*) > 1` can only find detection written as SQL aggregation; this
> checker compares in Go, so the probe was blind to the very thing it was cited for
> and would have returned the same empty result had the checker already existed. Full
> write-up in `WRONG_CALLS.md` (2026-07-31). Do not reuse that grep as an absence
> proof.
>
> **Deliberately inert, and that is not a defect to re-report.** `content_duplication`
> is registered in Go and configured in **no** agent — verified three ways: no
> `agent_definitions` row mentions it (active, inactive, snapshot or deleted); none of
> the three agents that actually run `run_discovery_checks` list it
> (`design-`/`completeness-`/`quality-discovery-agent`); no migration wires it. That
> is a **known, recorded decision**, not an oversight: `786722206` measured that
> enabling it today would delete nothing (11 duplicate groups fleet-wide, 0
> content-identical, 10 legitimately repeated, 1 NULL-content), so there is no backlog
> to clear and the enable switch waits on `brochure_component_library`. **A thread
> arriving here should not "discover" the inertness and route work at it.**
>
> Their 11 groups reconcile exactly with the 17 measured below: `17 − 6` (vonc's six,
> fixed 2026-07-30) `= 11`, and the "11 legitimate" figure below counts the NULL-content
> group they have since separated out.

The page shipped doubled on 2026-07-28 and was found on 2026-07-30 only because a
human ran a census by hand.

## The fix constraint that matters — do NOT add a naive unique index

A unique index on `(page_id, slot_name)` is the obvious fix and it is **wrong**.
Fleet census, 2026-07-30 — 17 duplicate `(page_id, slot_name)` groups exist, and
**11 of them are legitimate**:

```sql
WITH dups AS (SELECT page_id, slot_name FROM page_components GROUP BY 1,2 HAVING count(*)>1)
SELECT s.domain, p.name, pc.slot_name, count(*) AS rows,
       count(DISTINCT md5(pc.content_data::text)) AS distinct_content
FROM page_components pc JOIN dups d USING (page_id, slot_name)
JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
GROUP BY 1,2,3;
```

| verdict | groups | what they are |
|---|---|---|
| content **differs** | 11 | genuine repeated slots — `generic-text-block` used 2–3× per page on ai-agent-orchestration, leopardess, gaswholesalers, finetuning, idea.uk; `info-card-grid` ×2 on webdesign.co.uk |
| content **IDENTICAL** | 6 | all six are vonc `about` — this bug, and the only true duplication in the fleet |

So the discriminator is **content identity, not slot repetition**. A guard must
dedup on `(slot_name, md5(content_data))`, or reconcile the incoming count against
`pages.sections`/`site_plan_sections` — not forbid a repeated slot name.

Two footnotes that will bite whoever measures this again:
- `finetuning.uk/our-position-on-ai` reports `distinct_content = 0`, not 2: both
  rows have **NULL** `content_data`. The census in HANDOFF D filtered
  `content_data IS NOT NULL` and could not see rows of this shape at all.
- The handoff's stated worry — "the same cause may be live on other pages and
  other sites" — measures **false**. One event, one page, one timestamp.

## Fix candidates, ordered by what closes the door

1. **A dedup guard in `save_page_sections_action.go`** before the DELETE+INSERT:
   collapse adjacent sections identical on `(slot_name, md5(content_data))`, and
   log loudly when it fires. Makes the bad state unrepresentable at the only
   choke point every write passes through. Does not need the root cause.
2. ~~**A discovery check** (`check_duplicate_sections`) using the census query
   above, gated on content identity.~~ **DONE by the `151` lane, 2026-07-31** —
   `check_content_duplication` + `remove_duplicate_page_sections_action`, keyed on
   `SectionIdentityKey(slot_name, raw)`, deliberately not yet enabled. The note
   this candidate carried — *"a check is inert unless it is configured in an agent
   that actually runs `run_discovery_checks`; wire it, don't just register it"*
   (`bugs_open/149`) — turned out to describe the shipped state exactly, but **as a
   deliberate hold with the no-op measured** (`786722206`), not an oversight. Do not
   route work at it.
3. **Populate `content_hash`** on insert. It exists and is empty; a populated
   hash makes both of the above cheap and makes identity visible in a plain
   `SELECT`.
4. Find the producer — needs a fresh reproduction, since the run is past
   retention. Lowest priority: (1) contains the damage without it.

## The live instance — FIXED AND VERIFIED 2026-07-30

Applied with owner approval (the `DELETE` was refused by the permission classifier
on the first attempt, which is the correct default for a destructive write to a
live site):

1. `DELETE` of the six even-position rows — each byte-identical to its survivor on
   `rendered_html`, `content_data` and `component_id`, so nothing was lost — then a
   renumber of the survivors to positions 1–6 (`DELETE 6`, `UPDATE 5`; position 1
   was already correct). One transaction.
2. Assemble-only rerender + deploy:
   `docs/agent_docs/docs024_key_docs_latest/gauntlet_dead_cta/scripts/rerender_page_vonc.sh a28abcd7-186b-4a33-9b89-5d7bfd727012`
   — a generalisation of the arena script, which hardcoded its page id. `COMPLETED`.

Verified at the served artefact, cache-busted: **90,220 → 53,372 bytes**, the
doubled strings now appear **×1** (were ×2), and `data-component` appears exactly
**6** times (was 12), which is the check that ties the served page to the 6 rows.
Controls unaffected. The plan already said 6, so a later legitimate rebuild will
not restore the doubling.

## How to verify

```bash
# 1. rows: expect 6, positions 1..6, no identical pairs
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
SELECT position, slot_name, left(md5(rendered_html),8) FROM page_components
WHERE page_id='a28abcd7-186b-4a33-9b89-5d7bfd727012' ORDER BY position;"

# 2. served page: expect x1, not x2. Print the code — a B2 NoSuchKey body
#    reads as page content (see HANDOFF D landmines).
curl -s -w "http_code=%{http_code}\n" https://vonc.com/about.html \
  | grep -c "A provocation, not a prompt"
```

## Landmines

- **`pages.url` is `/about.html`, not `/about/index.html`.** Constructing the
  path returns a B2 `NoSuchKey` whose 286-char JSON body reads as page content —
  HANDOFF D briefly concluded the page was blank this way. Read `pages.url`.
- **Much of vonc's text arrives client-side**, so a raw `curl | grep` under-counts
  (4 of the 6 doubled strings show 0 in the raw HTML). Render, or match on strings
  the server actually emits.
- **Fixing `page_components` alone can be undone**: `site_plans` →
  `site_plan_sections` → `pages` → `page_components` regenerates it
  (LANDMINES, dartsonline lane). Here the plan is already correct at 6, so the
  cleanup is safe — but check that before hand-fixing any *other* page.

---

## Update 2026-08-04 — candidate 1 is BUILT, and this file's own proposed key was WRONG

**Still OPEN.** The code is committed and **inert until the next chassis roll**; the bar is
*fixed AND live*. Do not close it on the commit. Lane docs:
`docs/agent_docs/docs024_key_docs_latest/bugfix_156_duplicate_sections/`.
Council: correlation `1a3f4f27-a3b9-4388-b899-a36a911a976e`.

### Validity re-measured before starting — and the premise had moved

| this file's claim | 2026-08-04 | verdict |
|---|---|---|
| no unique index beyond the pkey | `\d page_components`: 9 indexes, only `page_components_pkey` unique | STANDS |
| no guard compares sections to each other | read all seven refusal/record blocks; every one compares incoming against **existing rows** or a floor | STANDS |
| `content_hash` never written | no writer fleet-wide | STANDS |
| 17 duplicate groups, 6 content-identical | **12 groups, ZERO content-identical** | **CHANGED** |

The vonc six are gone (fixed 07-30). So **there was nothing to repair**, and the guard had to
be judged solely on whether it could ever destroy a live section — not on what it cleans up.

### The correction that matters most: candidate 1's key would have deleted a live section

This file's candidate 1 says to collapse on `(slot_name, md5(content_data))`. **Its own census
footnote, forty lines above, says finetuning.uk/our-position-on-ai has two rows with NULL
`content_data` on both.** Under the proposed key those two are "identical", so the literal
recommendation would have deleted a live section on a shape this same file flags as a trap.
The two halves were written for different readers — the footnote warns a future *measurer*,
the candidate instructs a future *fixer* — and nothing joined them up.

**Shipped rule instead:** collapse an entry only when **every value the INSERT would bind is
equal, `position` excluded** — `slot_name`, `rendered_html`, `component_id` and
`content_data`, each with the insert loop's own normalisations (nil and empty `content_data`
both bind SQL NULL; an unparseable `component_id` binds NULL). Then the collapsed row would
have been *indistinguishable from its survivor in the database*, so nothing representable can
be lost. It still catches the whole recorded incident — vonc's six pairs matched on all four.

`rendered_html` in the key is what fixes the NULL-content case. `component_id` is free
narrowing.

### What was built

- `platform/orchestration/actions/save_sections_dedup.go` (new) — pure seam
  (`collapseDuplicateSections`) + DB seam + the durable record. Sibling-file shape, so the
  footprint in `save_page_sections_action.go` is one call and one result key.
- Call site placed **immediately after the "sections reaching save" diagnostic** (so that log
  keeps the TRUE arrival count) and **before every guard**. That placement is not tidiness —
  see below.
- `platform/orchestration/datahelpers/section_text.go` — comment-only cross-reference on
  `SectionIdentityKey`, so the next person to widen it reads about the pre-persist rule.
- 14 tests, each naming the mutation that makes it fail; all seven mutations were **run**.

### Three things found while building it that this file did not know

1. **A doubled list makes four other measurements lie.** The content-regression guard's
   `newTextLen` is doubled, so a page truly cut to 13% of its deployed text reads as 26% and
   clears the 25% floor. The completeness floor's numerator is doubled, so a save that saw 2
   of 6 planned sections but emitted them twice scores 67% and clears the 0.5 floor. The
   claims record and the `content_data` record both double-count.
2. **The locked-slot path MANUFACTURES a duplicate of human-locked copy.** With a doubled
   list the first copy of a locked slot consumes the lock and is discarded
   (`lr.consumed = true`); the **second copy falls through and is INSERTed beside the locked
   row**. So this defect does not merely duplicate agent-written sections — it can put a
   second copy of a slot next to one a human locked. Independent argument for collapsing at
   the save rather than relying on a post-hoc detector.
3. **The plan guard had to be mirrored, with the failure direction INVERTED.**
   `remove_duplicate_page_sections` refuses to delete a repetition the effective plan source
   specifies (council trail `da3f2d9b`, owner decision 07-31). A save-time collapse ignoring
   the plan would make the two halves disagree about the same question on the same table. So
   the guard calls the same `datahelpers.PlanSpecifiedSectionCounts` with the same per-slot
   accounting — but where the repair **fails closed** (an unreadable plan aborts it, because
   it is about to DELETE), a collapse guard's conservative direction is **not collapsing**, so
   a plan read error returns the incoming set untouched. Both mean "do nothing destructive".

### The durable record, and why it is half the value

This file records `[UNRECOVERABLE] the producer of that 12-entry list is not identifiable
from retained data`. The guard writes `agent_error_log` with code
`CONTENT_DUPLICATE_SECTIONS_COLLAPSED` carrying exactly what that hunt lacked: which
extraction path built the list (`sections_source`), the metadata field and its origin, the
step, the driving work item — and the **adjacency signature**, `adjacent` / `non_adjacent` /
`mixed`, which preserves the `1,1,2,2,3,3` vs `1,2,3,1,2,3` distinction that ruled out the
concurrent-save race. **Candidate 4 (find the producer) becomes tractable the first time this
fires**, which is the first time it has been.

### Candidate 3 (`populate content_hash`) — DELIBERATELY NOT DONE

Recorded as a decision, not an omission. Seven Go call sites INSERT into `page_components`;
only this one would populate the column, so `content_hash IS NULL` would read as "not a
duplicate" for every row the other six wrote. A Go-marshal hash would also not equal a hash of
the jsonb `::text` read back, so the column would carry a value nothing else can recompute — a
third definition of section identity beside `SectionIdentityKey` and the guard's. **If it is
ever wanted, the honest shape is a DB-side generated column `md5(content_data::text)`**, which
covers all writers at once. Owned by nobody today.

### What is still owed

- **The roll.** Then the post-roll pod-grep and the induction in
  `bugfix_156_duplicate_sections/RUNBOOK_duplicate_sections.md` — a **behavioural** induction,
  not a grep: feed a save a doubled list and confirm 6 rows and one `agent_error_log` row.
- **Council verdict** on `1a3f4f27-a3b9-4388-b899-a36a911a976e`.
- Candidate 4 (the producer) still needs a fresh reproduction. The record above is what makes
  it findable.
