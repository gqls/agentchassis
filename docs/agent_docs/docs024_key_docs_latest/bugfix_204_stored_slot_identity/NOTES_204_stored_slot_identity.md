# NOTES — 204, the stored-slot-identity blindness at its remaining call sites

Append-only, newest at the bottom. Technical log: what was tried, what the system
actually said, and every misstep.

---

## 2026-08-21 (a) — session start: what I found before touching anything

Picked up `bugs_open/204` on the owner's instruction. First job was to establish
what part of it is still open, because the file's own headline says
**"✅ FIXED, LIVE AND BEHAVIOURALLY VERIFIED END TO END — 2026-08-06, v1.0.1259"**
and is deliberately kept in `bugs_open/` (owner direction 2026-08-06, "leave the
bugs that you've found in bugs_open not in the closed bug file" — so the file's
presence there is NOT evidence the defect is live; read the foot of the file).

**The originally-filed defect IS fixed.** `plan_sections` (the BUILD path) resolves
`page_components.component_id` first via `loadPageSlotComponentIDs`
(`platform/orchestration/actions/plan_sections_action.go:1739`), shipped as
`13252f714`, council APPROVED `d3e232b8`, live v1.0.1257, canary proven at
v1.0.1259. Nothing here disputes that.

**What is still open is the two CONTRIBUTIONS appended later**, and the 2026-08-20
one is the real remaining bug: the same blindness at a THIRD call site,
`ValidateSitePlanAction`'s `validate_components` arm, which **drops** an
unresolvable name where `plan_sections` merely **defers** it.

### Ownership check (CLAUDE.md: check who owns it before routing work at a bug)

`scripts/who-owns.py 204` — the two lanes that CITE it most (`brochure_component_library`,
`loanandmortgagecalculator_couk`) are the lanes that FILED the two contributions, on
2026-08-17 and 2026-08-20, from canaries fired for *other* fixes (215's identity work).
Neither is fixing 204 itself; both explicitly say so ("Not a reopening", "filed for
diagnosis rather than guessed"). `bug_backlog_clearing` owned the ORIGINAL fix and its
last commit in-lane was 2026-08-14 on an unrelated bug (264). Last commit touching the
204 file: `e102241cd`, 2026-08-20, the contribution itself. **No thread is fixing this.**

Open work-item check: no `site_work_items` in a non-terminal state mention
`plan_sections` / `validate_plan` / `validate_components` / positional slots. The three
non-terminal hits on that query are `section_source_drift` items from 07-28/08-04/08-10,
a different (and older) mechanism.

⚠ Both checks are LAGGING — `who-owns.py` reads COMMITS, so a session mid-fix with a
dirty tree is invisible. Re-run at each phase boundary.

### Is the bug still valid? Read at HEAD, 2026-08-21

Yes, and it is worse than when it was written. Three independent confirmations:

1. **The code is unchanged.** `v3_site_actions.go:3838` still does
   `fn, ok := resolver.resolve(name); if !ok { …; continue }` — `continue` being the
   drop. `loadComponentNameResolver` (`:4330`) still selects only
   `function, name, display_name FROM content_components WHERE component_level IN
   ('section','element')`. `resolve()` (`:4369`) has five arms and none of them can
   see a `page_components.slot_name`.
2. **The config that arms it is live on two agents** [MEASURED 2026-08-21]:
   ```sql
   SELECT type, step.key, step.value->'config'->>'validate_components', step.value->'config'->>'menu_field'
   FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config->'workflow'->'steps') AS step
   WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
     AND step.value->'config' ? 'validate_components';
   ```
   → `build-site-planner|validate_plan|true|available_components` and
   `site-planner|validate_plan|true|(null)`.
3. **The population grew.** Census of `pages.sections` names resolvable by neither
   `content_components.name` nor `.function`:

   | | 2026-08-05 (filing) | 2026-08-06 | **2026-08-21 (today)** |
   |---|---|---|---|
   | unresolvable names | 86 | 87 | **107** |
   | sites | 5 | 6 | **7** |

   Today: loanandmortgagecalculator.co.uk 70 (41 pages), gaswholesalers.com 11,
   finetuning.uk 10, leopardessconsulting.co.uk 6, dartsonline.com 4,
   robot-hands.com 4, oufe.com 2. loancalculator.co.uk — 57/57 at filing — is now
   **0 unresolvable**, which is the 08-06 fix's own footprint showing up in the census
   (its sections were re-pointed), not a shrinking problem.

### The measurement that settles it, and it is not one I expected to get

`bugs_open/282`'s lane shipped a DURABLE record of every dropped section name
(`agent_error_log`, `error_code='PLAN_SECTION_NAME_DROPPED'`, from
`recordDroppedSectionNames` / `droppedFindings` in `component_name_resolver_menu.go`).
It went live on the 08-16 roll. Since then [MEASURED 2026-08-21]:

```sql
SELECT action, count(*) AS drops,
       count(*) FILTER (WHERE context->>'section' ~ '-[0-9]+$') AS positional_shaped,
       count(DISTINCT context->>'page') AS pages,
       min(occurred_at)::date, max(occurred_at)::date
FROM agent_error_log WHERE error_code='PLAN_SECTION_NAME_DROPPED' GROUP BY action;
```
```
 validate_plan | 140 | 140 | 41 | 2026-08-17 | 2026-08-20
```
Breakdown: `prose-0` 70, `tool-1` 34, `prose-2` 18, `tool-0` 12, `prose-1` 6.

**140 of 140 — every recorded section drop in the fleet is this bug.** Not one
display-name variant, not one typo, not one stale function: the class of miss that
`validate_components` exists to catch has produced ZERO records, and the class it was
never designed to see has produced all of them.

⚠ **Why this figure could have come out otherwise, which is what makes it evidence**
(the estate's disconfirmability rule): the same query would have returned a mixture, or
mostly display-name drops, if the resolver's other arms were doing the work the comment
claims. It would have returned 0 if `validate_components` were off, or if the durable
record were unwired. It returned neither.

⚠ **What it does NOT say:** it covers 08-17 onward only, because that is when the record
shipped — 140 is a lower bound on the damage, not a total, and the 08-17 and 08-20
incidents are inside the window. Do not quote it as "the drops began on 08-17".

### The blast radius is wider than the bug file says: FOUR call sites, not one

The file names `validate_plan`. Grepping the resolver's own symbols
(`loadComponentNameResolver(`, `resolver.resolve(`, `recordDroppedSectionNamesFor`)
finds it consumed at **four** places, not one:

| call site | file:line | what it does with the drop |
|---|---|---|
| `validate_plan` | `v3_site_actions.go:3838` | drops from the plan; **before** the object-form→string split, so RFC_016 plan-time `facts` die with the entry |
| `applyAddToPage` | `apply_gap_plan_action.go:244` | drops from a `content_rewrite` item's `add_sections`. The file's own comment calls this **"the fleet's dominant placement path"** |
| `applyNewPage` | `apply_gap_plan_action.go:374` | drops before INSERT. A new page has no stored slots, so this one is probably CORRECT |
| `applyRetypeExisting` | `apply_gap_plan_action.go:905` | drops, then `UPDATE pages SET sections = $3::jsonb` **directly onto a live page** |

So the persistence surfaces are two, not one: the `pages` upsert
(`site_db_actions.go:1201`, `sections = EXCLUDED.sections`, unguarded) and the retype
UPDATE.

### The precedent that is sitting inside the very statement that does the damage

`site_db_actions.go:1201`'s two immediate neighbours in the same `ON CONFLICT DO UPDATE`
were given destructive-write guards on 2026-08-19 after blank overwrites were measured
on robot-hands.com:

```sql
nav_label        = COALESCE(NULLIF(pages.nav_label, ''), EXCLUDED.nav_label),
meta_description = COALESCE(NULLIF(EXCLUDED.meta_description, ''), pages.meta_description),
sections         = EXCLUDED.sections,          -- <- no guard
```

Same statement, same class of harm (an empty incoming value overwriting a real stored
one), fixed for the two text columns and not for the jsonb one.

### Filed for an independent read

Per CLAUDE.md's diagnosis-before-debugging default, the claim I am about to make is
cross-cutting and structural ("one shared resolver, four call sites, two write
surfaces"), so I filed a `090` run rather than assert it from my own greps:
intake fired 2026-08-21, **RUN_CORRELATION_ID `1588b0da-5657-451a-8dc5-a5f63324712f`**
(the dispatch loop's own correlation — the key the artifacts are written under, NOT the
intake id the script prints first). Verdict to be recorded below when it lands.

---

## 2026-08-21 (b) — the live wiring, which constrains the fix more than the code does

Read from `agent_definitions` (live rows: `is_active`, `NOT is_snapshot`,
`deleted_at IS NULL`), not from the seeds.

### The two agents that arm `validate_components` do NOT have the same workflow

```
build-site-planner: complete, emit_design, emit_imagery, ensure_site, load_components,
                    load_existing_pages, load_styles, plan_site, populate_nav,
                    reconcile_site_plan, read_specs, sync_pages, validate_plan, write_site_plan
site-planner:       complete, load_available_components, load_style_collections,
                    plan_site, validate_plan
```

Three consequences, and the first one nearly cost me a wrong design:

1. **`site-planner` has no `load_existing_pages` step**, so its `validate_plan` runs with
   an EMPTY `existingPages`. A fix keyed on the `existing_pages` collected-data field —
   which is the shape the 08-20 contribution suggested ("trust stored state over the
   component catalogue") — would be **inert on that agent by construction**. A fix that
   queries the DB covers both. I had been leaning toward the collected-data route because
   it needs no query; this measurement is what moved me off it. [MEASURED 2026-08-21]
2. `site-planner` also has no `write_site_plan` and no `sync_pages`, so **build-site-planner
   is the damaging one** — which matches the drop record, where all 140 rows carry
   `agent_type='build-site-planner'` and none carry `site-planner`.
3. `build-site-planner`'s `load_existing_pages` query DOES select `p.sections`, so for
   that agent the realised list is in hand. It runs via `query_database`, which
   stringifies jsonb — `realisedSectionsOf` already handles the string case.

### The write surface is shared by three agents, not one

`upsertPage` (`site_db_actions.go:1114`, holding the unguarded
`sections = EXCLUDED.sections` at `:1201`) has exactly **one** Go caller,
`SyncPagesToDBAction` — but that action is wired into **three** live agents:
`build-site-planner` (as step `sync_pages`), `pageflow-builder` and
`site-work-orchestrator` (both as `sync_pages_to_db`). So a guard there is one edit
covering three agents.

### `apply_gap_plan` is live, and its exposure is LATENT, not measured

`apply_gap_plan` runs as `content-gap-planner`'s `apply_plan` step. It is active —
46 `gap_plan_*` work items, most recent 2026-08-20. But **zero** of the 140
`PLAN_SECTION_NAME_DROPPED` rows come from any `apply_gap_plan:*` action.

⚠ **State that honestly.** It means the gap planner has not been pointed at a
decomposed page's existing sections since the record shipped on 08-16. It does NOT
mean those three call sites are safe — `applyRetypeExisting` writes
`sections = $3::jsonb` straight onto a live page with no upsert guard in front of it.
The right word is *latent*, and the evidence for the risk is the code path, not a row.
[MEASURED: zero drops. INFERRED: that this is exposure rather than immunity.]

### Baseline for attribution

A clean `git archive HEAD` tree (not the shared working tree, which carries at least
eight other sessions' dirty files today) builds `./platform/...` and passes
`go test ./platform/orchestration/actions/` — `ok … 2.519s`. Any failure I introduce
from here is mine.

---

## 2026-08-21 (c) — what shipped, the 090's actual verdict, and the 33 seconds HEAD did not compile

### Shipped

`d376ca9b8` (A: the shared reader, a pure move) → `7baaf513b` (a hotfix, below) →
`c6446f5da` (B: the rescue, its tests, PLAN-051 + the PLAN-027 amendment). Council
correlation `f73f4eeb-5d79-482c-bc9b-b33f0ab64f76`, `Council-Submitted:` — **verdict
not yet read.** Inert until the next chassis roll.

17 tests, **7 mutations run and 7 killed**: conflict rule deleted; `SlotNameSet` made
to inherit it; pod-greppable warning string altered; rescue keep removed; rescue
scoped to site rather than page; `slotUnknown` made to drop; kept name rewritten to
its component function.

### The 090 came back UNVERIFIABLE, and that is not the same as "no result"

Run correlation `1588b0da-5657-451a-8dc5-a5f63324712f`, terminated **UNVERIFIABLE at
the iteration cap** — not REFUTED, not CONFIRMED. Reading its trail rather than its
label is the whole value:

- **It confirmed the two halves that matter, independently and with its own
  citations**: that `loadComponentNameResolver`'s key space is exactly
  `content_components.function/name/display_name`, and that `resolve()`'s miss removes
  the entry from `pm["sections"]`.
- **And it found its own live evidence** — `PLAN_SECTION_NAME_DROPPED` rows naming
  `prose-0`/`tool-1` for five pages (`guide-when-repayments-are-a-struggle`, `legal`,
  `loans-application-tracker`, `loans-credit-health-check`, `loans-damage-checker`)
  where `page_components.slot_name` records **those same names for those same pages**.
  Different pages from the ones I had looked at. Its phrasing: *"a name with real
  stored slot identity was dropped by a resolver that never queries page_components."*
- **Why it stopped:** `site_db_actions.go` was omitted from its bundle for size, so it
  could not close the persistence half; and the code index it reads is stale (mirrors
  a commit two days old), so it correctly refused to call
  `component_selector.go:SelectComponentByType` absent rather than unread.

**I answered its open question rather than leaving it**: the selector never touches
`page_components` (zero grep hits), and its only caller is `plan_sections`'
`resolveSectionComponent`, whose own doc comment says it handles "a section name that
didn't match any function directly" — reached only after Path 0 has already tried
stored identity at `:1177`. So it is catalogue-only **by position in the chain**, and
correctly so. Not a fifth call site.

⚠ Carry its caveat forward: **the code index is stale, so an absence in it is
`unknown`, not `confirmed absent`.** I did not re-derive my own call-site enumeration
from the index — it came from grepping the resolver's symbols directly — but anyone
extending this work should not treat "the index found nothing" as an answer.

### The misstep: I put a call to an undefined symbol into HEAD

Sequence: wrote `stored_slot_rescue.go` (definition, **untracked**), added the
four-line caller hunk to `v3_site_actions.go`, moved on to the next file intending to
commit both within the minute. Commit `af4743464` — an unrelated `342` change by
another session — took `v3_site_actions.go` from the shared tree **with my caller in
it** and left my untracked definition behind. HEAD then held a call to
`storedSlotRescueFor` with no such symbol in the repo.

- **Window: 33 seconds** (15:03:27 → 15:04:00, restored by `7baaf513b`). No build
  started inside it. That is luck, not a control — and it is the second consecutive
  occurrence of this trap that closed on luck.
- **Nothing automatic caught it.** `go build` in my tree was green throughout (it
  reads the tree, not HEAD) and every test passed. I found it only because
  `git diff --stat` on a file I was about to commit came back **empty** when I
  expected a diff, so I went looking.
- **I had the rule in my PLAN and it did not save me.** LANDMINES said *"commit the
  definition FIRST, ALONE"*; my plan quoted it. The rule is under-specified: it only
  protects you once both halves exist, and **the exposure opens the moment you type
  the CALLER**, not when you choose a pathspec. Sharpened in `LANDMINES.md` and logged
  in `WRONG_CALLS.md`: **write the definition, COMMIT IT, then type the first line
  that calls it.**
- Separately, my PLAN-051 **index row** was swept into `d79e4243c` (a `bugs_open/335`
  commit) before I could commit it. Nothing lost; forward-only holds. Worth recording
  as a plain instance of the thing CLAUDE.md warns about rather than as a complaint.

### Two figures from the drafted plan that did not survive grounding

- *"the 141 non-decomposed sites"* — invented. `sites` holds **45 rows** (23 deployed,
  17 pool, 2 active, 2 test, 1 system); **27** have any active page; 7 carry
  unresolvable names, so the population is about **20**. Logged in `WRONG_CALLS.md`
  under *adjacent accuracy is not evidence* — the figures either side of it
  re-measured exactly, which is what made it credible.
- *"72 of 748 active pages at `sections=[]`, 60 of them tools"* — **re-measured and
  correct**, recorded here so the next reader knows which of the two was checked.

---

## 2026-08-21 (d) — the council APPROVED it and still found a real defect

Corr `f73f4eeb-5d79-482c-bc9b-b33f0ab64f76`: **approved with 4 advisory objections,
none high-severity**, `gated_by_truncation: false`. Seats: `editquality` approve,
`bug_historian` **object**, `reuse_agent` approve, `guidelines` approve, 4 abstained.

**An APPROVED verdict is not a reason to skip reading it.** The `bug_historian`
seat's two mediums were both worth work, and one of them was right about something I
had checked and got wrong.

### Objection 1 — the read-failure keep is indistinguishable from a real rescue

> *"slotUnknown collapses 'DB read failed' into the same keep-path as 'legitimately
> stored' … it silently absorbs an infrastructure fault into an apparently-clean
> validation pass."*

My first reaction was that this was half wrong, and it was: the seat read only the
submission's sketch, and the code already logged the two differently (a Warn at the
read, `read_failed` on the gap-plan per-entry Info, `stored_slot_read_failed` on
validate's summary line).

**But it was right where it counted, and I had not looked there.** The *durable*
record did not distinguish them. `keptFinding()` returned nil when `kept == 0`, so a
run that kept every name **because the database was unreachable** filed **no row at
all** — and therefore read, in the only channel that survives log rotation, exactly
like a clean pass. That is precisely the silent-absorb shape this whole lane exists
to remove, reproduced one level up in my own fix.

Fixed: `PLAN_SECTION_STORED_SLOT_READ_FAILED` is its own error code, filed
unconditionally on failure, carrying `kept_without_checking`; `keptCount()` stays 0
because nothing was *rescued*; validate logs the unchecked keep per entry at Warn.
Two mutations pin it — reusing the rescue's error code, and filing nothing (the
pre-objection behaviour). Both kill their test.

### Objection 2 — the other private slot loaders may already apply the wrong rule

> *"the two other cited prior fixes are NOT migrated … If SlotNameSet's
> deliberately-different conflict rule is correct, the two untouched loaders should
> be checked for whether they silently apply the WRONG rule already."*

Checked rather than accepted or waved off:

- The rerender path's `loadContentComponentsByID` **builds no slot map at all** — it
  loads component schemas by id, so it has no conflict rule to get wrong. The
  objection's premise does not hold for that one.
- The loader that *does* key on `slot_name` with **no rule and no `ORDER BY`** is
  `enrichSectionComponentsWithBriefs` (`v3_site_actions.go`). A page with a repeated
  slot_name carrying **different** briefs gets a non-deterministic last-write-wins.
- **[MEASURED 2026-08-21] 0 such pairs exist today** — and the measurement could have
  come out otherwise, which is what makes it worth recording: the population is real
  (1,619 rows carrying a brief across 553 pages) and the shape is reachable (18
  repeated slot groups fleet-wide). It is latent, not firing.

Recorded in PLAN-051 and left unfixed **here**: it is a different question (which
*brief*, not which *component*) and belongs to whoever owns that path. Filing it into
this lane's fix would be the scope creep the `editquality` seat was already gently
flagging.

### The two lows, recorded not actioned

- *"Refactoring `plan_sections_action.go` is adjacent-not-causal"* — fair. Argued in
  the PLAN (the drift class, and "unify later is how a seam becomes folklore"), and
  the risk is bounded by its existing tests passing unmodified.
- *"The submission's edit list does not enumerate the `000_concept_index.md` row"* —
  also fair, and with an irony worth writing down: that row had **already been swept
  into `d79e4243c`** by a concurrent session before I could commit it, so the plan's
  own account of its commit contents was wrong in a second way the seat could not see.

---

## 2026-08-21 (e) — the second council round found a LIVE defect I had asserted away

Corr `2466d82c-17f8-4ebc-948d-ff8dbab9cee4`: **approved, 5 advisory objections, none
high-severity**. Two seats objected (`editquality`, `bug_historian`, `guardian`), and
this round was more valuable than the first.

### The one that mattered: my "safe by construction" list was not measured

`bug_historian`, medium:

> *"the plan explicitly enumerates five other write paths … as safe 'by construction'
> or 'different authorities' — but that safety is ASSERTED, not measured with the same
> rigor the plan applies elsewhere (e.g. the 72/748 zero-sections census) … this
> council's history with exactly this class of incomplete guard [bugs_closed/001, 037,
> 050]."*

It named the precedent precisely, and it was right. Checking the five:

| path | claim | actual |
|---|---|---|
| `adopt_verbatim` | safe | **true** — always writes `[]string{portedPageSlot}`, one element |
| `create_blog_posts` | safe | **true** — has a floor: `if len(sections)==0 { hero, article-body, call-to-action }` |
| `apply_gap_plan` retype / `applyNewPage` conflict arm | safe | **true** — `defaultSectionsForPage` never returns empty, and `if len(resolved) > 0` floors the resolved path |
| `UpsertPageForRole` | different authority | not re-examined, out of scope |
| **`apply_adoption_plan`** | safe | **FALSE** |

`apply_adoption_plan_action.go` builds `sections := []string{}` and fills it only when
the plan page carries the key, then wrote it through an **unguarded `EXCLUDED`**, over
LIVE pages, via `ON CONFLICT (site_id, name)`, on the live `site-adoption-agent`
(`apply_plan`).

**And the damning part:** that statement *already carried* the `meta_description`
guard, with the comment **"Same guard as upsertPage"** (`bugs_open/320`). A previous
lane fixed one half of this exact omission on this exact statement and left the other
— which is the 001→037 shape happening a second time, on a second statement, and
exactly what the seat predicted from history rather than from reading my code.

Fixed in the same round: same CASE, no release arm (adoption has no
deliberate-emptying intent), SQL extracted to `applyAdoptionPlanPagesUpsertSQL` so a
test can pin it, mutation-checked.

### The one about my own test being fake

`editquality`, medium ×2: no test asserted the durable finding is written.

True — and worse than the seat could see. The test occupying that slot,
`TestSectionsRefusalMessageNamesPagesAndCause`, **built the message string itself and
then asserted the string it had just built.** It tested nothing. It would have passed
with the production code deleted. That is a new entry in my own catalogue of ways to
write a test that cannot fail, and it is a nastier one than usual because it *looks*
like a content assertion.

Replaced with one that drives `SyncPagesToDBAction`, forces a refusal, and asserts the
`agent_error_log` write is attempted. Mutating the record away fails it.

### The guardian's point I did NOT close

> *"Reusing recompose_pages … changes its semantics from 'planner-side redesign signal
> read by validate_plan' to also 'sync-time destructive-write authorization' — two
> different consumers now key off one flag with different blast radii."*

That is correct as stated, and I have recorded it in PLAN-052 rather than argued it
away. I still think the reuse is right — a second flag would drift from the first, and
"this page is released for redesign" and "this page may be emptied" are the same
intent — but it IS a semantic expansion of a shared field, and the honest place for
that is the register, where the next person to add a consumer will meet it.

### Answered with a measurement

`guardian` asked me to confirm exhaustive callers rather than say two were found:
`upsertPage(` has exactly **one** caller of this function (`site_db_actions.go:380`).
The only other hit is `cmd/webdesignport/import.go:134` — a **different function of the
same name in a different package**, which is precisely the confusion LANDMINES' "THREE
`pages` upsert helpers have OPPOSITE collision policies" entry exists to flag.
