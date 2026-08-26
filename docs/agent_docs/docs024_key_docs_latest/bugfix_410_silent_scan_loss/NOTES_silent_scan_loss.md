# NOTES — bugfix 410, silent scan loss

Append-only, newest at the bottom. Technical log: evidence, commands, what the system
actually said, and every misstep.

---

## 2026-08-26 (session `bugs_open/410`) — picking the lane up, and what the first hour measured

### Why this session exists, and the ownership check that licensed it

`bugs_open/410` is a **pattern file** filed 2026-08-26 by the `dartsonline_traffic` lane at the
`news_editorial` lane's suggestion. Three seams, three lanes, one week, all failing toward the
quiet default (assemble/skip), all reporting `complete`, all leaving the artefact looking
freshly built.

Ownership checked before touching anything:

- `scripts/who-owns.py 410` → **"OWNED or recently active"**, but every commit it listed is a
  commit *to the bug file itself* (`bf9b3f607`, `29e50c5ce`, `44e951896`, `15c09a218`), all
  today, all documentation. No fix commits.
- The bug file says it in its own words: *"Nobody owns the seam, which is why it is filed
  rather than carried."*
- `news_editorial`'s NOTES say of instance 3: *"**Deliberately not fixed here**: a shared seam
  on the busiest pipeline that this change merely surfaced, and bolting it into a feature
  commit is the scope veto §6.1 warns about."*
- `ListAgents`: no session named for this bug or for `dartsonline_traffic`. The filing lane is
  not on the machine.

So: filed but **unowned**, and declined on purpose by the lane best placed to have carried it.
Resumed here. Both live lanes messaged before any edit (`news editorial`, `bugs_open/384`,
`bugs_open/399`) — replies logged below.

### The bug is still valid `[MEASURED 2026-08-26]`

`loadStoredSections` (`platform/orchestration/actions/rerender_page_sections_action.go`) still
carries the defect verbatim:

```go
if err := rows.Scan(&s.id, &s.parentInstanceID, &s.componentID, &s.slotName,
                    &cdJSON, &s.renderedHTML, &s.position, &s.componentVersionID); err != nil {
    logger.Warn("rerender_page_sections: row scan failed", zap.Error(err))
    continue
}
```

Returns `out, rows.Err()`. A scan failure yields **fewer sections, or none**, with **no error**.

**Cited by symbol, not by line, deliberately** — the bug file's own correction records that
`:1206` expired the same afternoon it was written (`bd811fa93` moved it ~32 lines). The
signature is at `:1206` today and the Warn at `:1237`; both will move again.

### The caller propagates — so the fix is not a no-op one level up

`rerender_page_sections_action.go:270`:

```go
stored, err := loadStoredSections(ctx, params.DB, pageID, logger)
if err != nil {
    return nil, fmt.Errorf("load stored sections: %w", err)
}
```

Checked because a guard whose error is re-swallowed by its own caller is the defect this file
is about, reintroduced as its own fix. It is not: the error surfaces and fails the step.

### The class is 225 sites, not one `[MEASURED 2026-08-26]`

A brace-matched census over `platform/ internal/ pkg/ cmd/`, classifying **every** `.Scan(`
call by what its **own** error branch does (not "is there a `continue` nearby"):

| shape | count |
|---|---|
| `scan err -> continue` (swallow) | **225 production sites** |
| `scan err -> return/break` (propagate) | 491 |
| other / unclassified | 360 |

Script: `scratchpad/scancensus2.py` (brace-matches the `if err != nil {` block that wraps the
Scan, rather than a fixed lookahead window).

**My first cut of this census said 237 and was worthless** — it looked ahead 12 lines for any
`continue`, which matches every loop that legitimately filters *after* a successful scan. The
number was in the right region by accident. Recorded because a plausible number from a wrong
encoding is the failure this whole bug is about. See missteps below.

**Consequence for the design: converting 225 sites is a programme, not a patch.** The plan must
convert the motivating site and *ratchet* the rest.

### Three precedents, and the fix is propagation rather than proposal

410 cites one. There are three, and the closest is the one it does not know about.

1. **`scanBlogArticles`** (`rebuild_blog_listing_action.go:465-520`) — **the same guard, on the
   same `rows.Scan` shape, already live.** `attempted, scanFailures := 0, 0`; per-row Warn and
   continue; **error** when `attempted > 0 && len(articles) == 0` (*"refusing to report an empty
   listing as 'no posts'"*); Warn when `scanFailures > 0` but some survived. It is **graded** —
   partial loss and total loss are different outcomes — which the all-or-nothing form is not.
   Independently verified by the `news_editorial` lane on my report.
2. **`collectPageSections`** (`validate_page_content_surface_sections.go:81`, guard at `:98`) —
   the count guard 410 cites, protected by a mutation-checked test.
3. **`TestFindingCodeScanEveryWriteIsRegistered`** with an `_unruled_cap` of **0** — supplied by
   the `bugs_open/399` lane, unprompted: a live instance of the ratchet shape, which *caught
   their own new code within a minute of them adding it*.

Plus the estate's ratchet idiom itself, `work_item_type_minting_ratchet_test.go`, whose header
states the rule this fix must obey: a blocking Go source-sensor test paired with an **advisory**
`scripts/pattern-check.py` check — *"Two layers, one pattern: change them TOGETHER"* — and
comments stripped before matching, citing the fleet-wide load-bearing-comments trap.

### The guardian's blast-radius question, answered before it is asked

The `news_editorial` lane warned (from their own REVISE round this morning, `53d71504`) that
guardian objected at **HIGH** severity to touching `rerender_page_sections` at all without a
canary or fast-revert path, on the grounds that it is the fleet's busiest pipeline — and that
architecture, render_guardian and debug_historian independently made the same point. Their
sharpest version: *"your guard's first live effect could be to start failing rerenders that
today quietly half-work, and how many that is on day one is a number a seat will want."*

**It is zero, and by schema rather than by luck** `[MEASURED 2026-08-26]`.

Every column the loader scans is structurally incapable of failing a scan on today's data:

| projected | destination | why it cannot fail |
|---|---|---|
| `id::text` | `string` | `id uuid NOT NULL` (PK) |
| `COALESCE(parent_instance_id::text,'')` | `string` | COALESCE'd |
| `COALESCE(component_id::text,'')` | `string` | COALESCE'd |
| `COALESCE(slot_name,'')` | `string` | COALESCE'd |
| `content_data` | `[]byte` | NULL-safe: scans to nil. 54 live rows ARE NULL |
| `COALESCE(rendered_html,'')` | `string` | COALESCE'd |
| `position` | `int` | `position integer NOT NULL` |
| `COALESCE(component_version_id::text,'')` | `string` | COALESCE'd |

```sql
SELECT count(*) AS rows_live,
       count(*) FILTER (WHERE position IS NULL)     AS null_position,
       count(*) FILTER (WHERE content_data IS NULL) AS null_content_data
FROM page_components WHERE build_status IS DISTINCT FROM 'removed';
--  2194 |  0 |  54
```

`\d page_components` confirms `position integer NOT NULL` and `id uuid NOT NULL`.

**So the guard cannot convert a working rerender into a failing one on today's data.** It can
only speak when the SELECT list and the Scan destinations diverge — which is a **code defect**,
introduced by an edit, and is exactly the edit that produced seven red tests today. That is the
answer to the canary objection: the change is inert on current data by construction, and the
population it would newly fail is empty and measurably so.

⚠ **This is a claim about TODAY's schema and TODAY's projection.** It expires the moment either
changes — a nullable column added to the SELECT without a COALESCE would make it false. The
disconfirming query is the one above; re-run it, do not quote this table.

### The relayed 6,428 / 3 figure — re-run here, reproduced exactly, and then REFUTED

> **CORRECTED 2026-08-26, within the hour, by the `bugs_open/384` lane (their commit `4a31c6b8f`):**
> **the figure is 203 of 17,285, not 3 of 6,428 — and my "independent verification" below was
> worthless, because I reproduced their number by repeating their exact mistake.**
>
> `site_work_items` is a **ROLLING WINDOW**: closing a row archives it into
> `site_work_items_archive`, out of the table both of us queried. We each measured the live slice
> and read it as the population. Re-run by me across both tables `[MEASURED 2026-08-26, mine]`:
>
> | table | items | with a reason |
> |---|---|---|
> | `site_work_items` (live) | 6,428 | 3 |
> | `site_work_items_archive` | 10,857 | 200 |
> | **TOTAL** | **17,285** | **203** |
>
> **What this does to the argument: it strengthens it.** All 203 carry `section_data_resolved`,
> which the Go reader knows; `template_changed` / `literal_markdown` via that path is **zero rows
> in both tables**. So "no unknown reason has ever reached the stale reader" now holds over the
> full recorded history rather than over a live window — a claim 404 could not previously make.
>
> **And the 384 lane's second point is the one that matters more here, which I had reached
> separately: this figure does not size my blast radius and must not appear in my rationale even
> corrected.** It counts work items carrying a `spec.reason`. My guard is about what
> `loadStoredSections` does when a `.Scan` fails. Different populations, different axes. The
> blast radius has to be measured **at the scan**, which is the NOT NULL / COALESCE table above,
> not at the item.
>
> Everything below this box is left as originally written, including the wrong headline, because
> the way it was wrong is the record.

### The relayed 6,428 / 3 figure — re-run here, and it reproduces exactly

410's corrections block records that the filing lane stamped a **relayed** number with
`[MEASURED]`, and retracted the attribution. I re-ran it rather than relay it again:

```sql
SELECT source, created_by, count(*) AS n,
       count(*) FILTER (WHERE COALESCE(spec->>'reason','') <> '') AS with_reason
FROM site_work_items WHERE item_type='page_rerender'
GROUP BY source, created_by ORDER BY n DESC;
```

`rerender-pages / rerender-pages → n = 6428, with_reason = 3` `[MEASURED 2026-08-26, mine]`.

**The 384 lane's figure is exactly right.** It is now first-hand for this lane, and the wider
decomposition is richer than the single ratio:

| producer | items | with a reason |
|---|---|---|
| `rerender-pages` (the fleet sweep) | 6,428 | **3** |
| `discovery` / generic | 429 | 429 |
| `discovery` / completeness-discovery-agent | 375 | 375 |
| `side_effect` / component-template-fixer | 326 | 326 |
| `render_news_section` | 222 | 222 |
| *(all page_rerender, all producers)* | **7,957** | **1,512** |

Reason distribution across all 7,957: `(none — assemble only)` **6,445**, `cta_links_stale` 820,
`template_changed` 389, `section_data_resolved` 260, then a long tail — including **5 items whose
`spec.reason` is a 143-character prose sentence about migration 415**, which is its own small
finding for the 404 lane: the reason field is being used as a free-text note by at least one
producer, and 404's whole mechanism is a hard-coded equality test against it.

The file's argument survives intact and is now better grounded: the fleet's own sweep is 81% of
all `page_rerender` items and stamps a reason on 3 of 6,428. **Assemble is the overwhelming norm
and should be — which is precisely why every drift lands there silently.**

### The seven red tests — corrected from six, by the lane that owns them

410 (and my first message) said **six**. It is **seven**. The `news_editorial` lane corrected it
unprompted and the way they got it wrong belongs in this file: they piped a test run through
`tail -6`, saw one failure, fixed that mock, re-ran, saw six more, and reported the second run's
count. **Their own evidence was truncated by their own command, and they formed a count from the
truncated view** — in a bug about instruments that quietly return less than they were given.

Cite by NAME, not line (they moved several today):

- `TestRerenderPageSections_SuccessEntryCarriesTheStoredSlotName`
  (`save_sections_stored_slot_identity_test.go`)
- `TestRerenderPageSections_FailsWhenComponentUnresolvedByNameOrID`
- `TestRerenderPageSections_ResolvesToolByComponentIDWithoutEscalating`
- `TestRerenderPageSections_ComponentIDWinsOverNameWhenBothResolve`
- `TestRerenderPageSections_InvalidTemplateByID_IsFatalAndNamed`
- `TestRerenderPageSections_EmptyTemplateCarriesWithoutFailing`
- `TestRerenderPageSections_StructuralCarryMakesANotReadySectionRerender`
  (the last six all in `rerender_page_sections_resolve_test.go`)

**And the sharper form of the finding, verified by them:** those tests' assertions are
outcome-shaped (*"expected exactly one section, got %d"*, *"stored_slot_name = %q"*). None
asserts rows-yielded against rows-scanned. So they **would have passed** on a column change that
was wrong in a way that still scanned. The reproduction is *"invisible to the tests that cover
it"*, not *"unhelpfully reported by them"*.

**The detail that makes it worse, and it is the best single line in this whole case:** those
tests call `mock.ExpectationsWereMet()`, so **sqlmock's own completeness assertion ran and
passed.** The query was issued exactly as expected; only the scan silently produced nothing.
Even the mocking framework's completeness check cannot see this failure.

### Peer replies, both useful, both changed something

- **`news editorial`** — *"GO FIRST, THE FUNCTION IS YOURS."* Not editing `loadStoredSections`'
  body; their council work is in `RerenderPageSectionsAction`'s per-row loop (~462-770), a
  different region of the same file. Their submission came back **REVISE** this morning with two
  HIGH objections, so they will not write that region today, and will take my version as base.
  They still warn: different regions is **not** protection from a same-file passenger — commit
  narrowly and read `git diff` on the file first.
- **`bugs_open/399`** — no collision; they add no `pattern-check.py` check and no Go source
  sensor (their source-scanning is two package-local tests pinning their own call sites). They
  took the *propagate-the-existing-idiom* framing and say it let them **reject** the fix their
  own bug file proposed, which would have been a third definition of "misdirected" beside two
  that already exist. They handed over the `_unruled_cap = 0` ratchet precedent unprompted.

### Missteps this session

1. **My first census encoded the wrong question and returned a plausible number.** "Is there a
   `continue` within 12 lines of a `.Scan(`" → 237. The real question is "does the Scan's **own**
   error branch `continue`" → 225. The two agree to within 5%, which is the trap: a wrong
   encoding that lands near the right answer is indistinguishable from a right one *by looking at
   the number*. Only rewriting it with brace matching separated them. Logged to `WRONG_CALLS.md`.
2. **I measured `spec->>'component_id'` on `site_work_items` and reported 0 "scoped to
   component" — on a column that the producer never writes, and which also exists as a real
   table column I had not read.** `component_id` IS a column on `site_work_items`; it is also 0;
   so my number was *accidentally right by a route that could not have come out otherwise*.
   `create_rerender_items_action.go`'s INSERT names `site_id, source, pipeline, item_type,
   severity, summary, page_id, priority, handler_agent, status, created_by, spec, item_key,
   batch_id` — **no `component_id` at all**. A zero from a column nobody writes is not evidence
   of an unused code path; it is evidence of nothing. Caught by reading the INSERT before
   writing the claim down anywhere durable. Logged to `WRONG_CALLS.md`.
3. **I nearly published a contradiction of the 384 lane's 6,428/3 figure.** My first query
   counted `page_rerender` items across **all** producers (7,957 total, 1,512 with a reason) and
   read as though "3" were off by five hundred-fold. The file scopes its claim to the
   `rerender-pages` producer, which I had not re-read carefully. Scoping to the same population
   reproduced their number **exactly**. Not sent, because the discrepancy was too large to be a
   real disagreement and that itself was the signal to re-read rather than to report. Logged.

---

## 2026-08-26 (later) — the fix is built, mutation-proved, and AT COUNCIL

### ⚠ SUBMISSION_CORR = `c8385154-17b4-43f5-94b2-41f552f43867`

Recorded here rather than left in the session transcript, for the reason the
`news_editorial` lane already paid once: a correlation in scrollback dies with the session.

```sql
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
 WHERE correlation_id='c8385154-17b4-43f5-94b2-41f552f43867' AND kind='council_report'
 ORDER BY created_at;
```

APPROVED → `Council-Reviewed: c8385154-…`. Committing first → `Council-Submitted: c8385154-…`,
which asserts nothing and is credited at report time.

Two client-side schema corrections needed before admission passed, both worth having in the
RUNBOOK for the next lane: `plan.grounded_in` must be an array of **strings** (not objects — put
the source inside the quote text), and `plan.risks` must be a **single string** (not an array —
join them into one prose block). `DRY_RUN=1` caught both for free.

### > **CORRECTED — the class is 207, not the 225 I told three lanes**

I sent "225 production sites" to `news editorial`, `bugs_open/384` and `bugs_open/399`, and
wrote it into this file. **It is 207** `[MEASURED 2026-08-26]`, in 127 files.

**What was wrong:** my classifier matched *any* `.Scan(` whose error branch continues. That also
catches `db.QueryRow(q, u).Scan(&id)` inside a loop over some other collection — a single-row
lookup where `continue` is ordinary control flow over an **item**, not loss of a **row the
database handed us**. There is no cursor-yielded count to compare against in that shape, so the
guard does not apply to it and it must not be in the population.

**What caught it:** generating the baseline and diffing the two classifiers, which disagreed on
eight files **in both directions**. Chasing one of them (`process_area_vet_sweep.go`, old 2 / new
0) showed it was a `QueryRowContext(...).Scan(...)` with an explicit `sql.ErrNoRows` arm —
nothing like the defect. The newer classifier had excluded it **by accident** (a regex quirk: it
required `{` straight after `nil`, and that branch reads `err != nil && err != sql.ErrNoRows {`),
which is right answer, wrong mechanism. So I re-encoded the shape as what the defect actually is:

> a `rows.Scan` inside a `for rows.Next()` loop whose error branch `continue`s.

**Why this matters beyond the number.** The looser shape would have produced a ratchet that
fires on `QueryRow` sites, where the remedy it names (`ScanShortfall`) is meaningless — a guard
that fires where its own fix does not apply is the "fires constantly, gets loosened, dies" trap,
arriving through the detector instead of through the data.

Third instance this session of the same error: **my measurement answered the question I encoded.**
Logged to `WRONG_CALLS.md`; the three lanes have been sent the correction.

### What shipped, and every mutation was RUN rather than promised

| file | what |
|---|---|
| `datahelpers/scan_completeness.go` | `ScanShortfall(offered, kept, subject)` — one implementation, strict policy only |
| `datahelpers/scan_completeness_test.go` | 4 cases; two exist to stop the guard being over-eager |
| `actions/rerender_page_sections_action.go` | `loadStoredSections` counts and refuses; marker on the branch |
| `actions/rerender_page_sections_scan_completeness_test.go` | 4 sqlmock cases; header names the seven red tests |
| `actions/scan_swallow_ratchet_test.go` | blocking source sensor + classifier StillBites |
| `actions/scan_swallow_baseline.txt` | 207 sites / 127 files, shared by both layers |
| `scripts/pattern-check.py` | advisory twin, tree-wide, same classifier, same baseline |

**Mutations, all executed:**

| guard | mutation | result |
|---|---|---|
| `ScanShortfall` | body → `return nil` | RED, restored GREEN |
| loader refusal | delete the call, restore `return out, rows.Err()` | **both** refusal tests RED — and the failure text reproduced the defect verbatim: *"returned 1 section(s) with no error"* and *"(0 sections, nil error)"* |
| ratchet, door-closing | remove a baseline line | RED: *"3 NEW silent scan loss site(s)"* |
| ratchet, anti-drift | inflate a baseline count | RED: *"FELL from 4 to 3 — good, now ratchet down"* |
| classifier | 7 synthetic cases incl. comment prose, `QueryRow`, non-default cursor name | all discriminate |

**Classifier parity verified rather than asserted** — the Go reader, the Python reader and the
baseline file all agree at **207 sites / 127 files, zero disagreements**, plus six shared edge
cases run through both. This is the property that makes "one baseline, two readers" true instead
of decorative. It is **not** mechanically enforced (no test invokes the Python classifier from
Go); that is stated as a residual in the submission rather than claimed as covered.

Full package suite green: `actions`, `discovery_checks`, `queryresolve`, `datahelpers`. And
`scripts/pattern-check.py` still exits 0 on the tree — checked deliberately, because it runs on
**every commit in every session** and breaking it would break every lane.

### A design decision worth recording, because the plan I was handed had it right and I nearly didn't

Fable's plan corrected me on something I had asserted: **the swallow shape is not the defect.**
`scanBlogArticles` is *in* the baseline — it has the exact shape AND a correct guard, because its
guard lives after the loop. Had I built the ban-style ratchet I first imagined, it would have
convicted the estate's own best precedent on this pattern. That is why the ratchet pins per-file
**counts** with an at-site opt-out marker, rather than banning the shape.

### The residual I am NOT fixing, named so it does not become 410's own quiet default

`loadStoredSections` still does `_ = json.Unmarshal(cdJSON, &s.contentData)`. On corrupt JSON
that **keeps the row and empties its content**, so `offered == kept` and the count guard cannot
see it. A second silent-loss class in the same loop, on a different axis (content, not rows).
Commented at the site, listed in the ratchet header's "what it cannot see", and stated in the
submission's risks. Not fixed here: it needs a decision about whether an unparseable section may
render as an empty one, which is a different question from whether a row may vanish.

---

## 2026-08-26 (evening) — VERDICT: APPROVED round 1, and the close-out adjudication

### The verdict

`c8385154` → **APPROVED, round 1** (11:20:59Z artifact): *"approved with 4 advisory
objection(s) — none high-severity."* Twelve seats; six abstained; objecting seats were
reuse_agent, guardian, debug_historian, prior_art_librarian (bug_historian and architecture
approved but attached objections). The guardian objection that hit the `news_editorial` lane at
HIGH that same morning arrived here at medium — the schema-not-luck blast-radius argument was
already in the submission, which is the difference.

`7c443aac6` carries `Council-Submitted:`, so `098` credits it automatically now the verdict is
approved. No amend (forward-only).

### The close-out review — fable was asked, died on a session limit (resets 20:40 BST), and the
### owner directed this model to carry it. Dispositions, each re-verified against the tree:

| seat | objection | disposition |
|---|---|---|
| editquality (low) | other callers may inherit the stricter error | **CLOSED** — exactly one production caller (`:270`), and it propagates |
| guardian (low arm) | confirm pattern-check stays advisory | **CLOSED** — `.githooks/pre-commit:41` runs it `\|\| true`; cannot block |
| prior_art (medium) | cited rounds/idiom may not exist as described | **CLOSED** — `3ed2b792` (2 rows) and `170147b4` (4 rows) both resolve in `diagnosis_artifacts`; the minting ratchet file exists with the two-layer rule verbatim. Honest note: I had RELAYED both round ids (from the guard comments and the 384 lane) before checking them — the seat was right to ask |
| debug_historian (medium) | blast-radius query covered 2 of 8 columns | **REFUTED, and the refutation strengthens the claim** — full-column measurement below |
| bug_historian (medium) | the comment must say the `continue` is safe ONLY because the trailing check survives | **ACTIONED** — comment sharpened at the site, naming the mutation test that goes red on deletion |
| reuse_agent (medium) | three implementations; track convergence | **ADJUDICATED + ACTIONED** — see below; converge-on-touch for the true sibling, false-sibling finding for the other |
| debug_historian (low) | state the pod-verification recipe | **ACTIONED** — RUNBOOK §12, three-way probe per the LANDMINES form |
| architecture (low) | parity verified once, not enforced | **ACTIONED** — recipe in the baseline header; mechanical enforcement REJECTED with the reason stated there |

### The full-column measurement (debug_historian's own terms) `[MEASURED 2026-08-26]`

```sql
SELECT count(*), count(*) FILTER (WHERE parent_instance_id IS NULL),
       count(*) FILTER (WHERE component_version_id IS NULL),
       count(*) FILTER (WHERE component_id IS NULL),
       count(*) FILTER (WHERE slot_name IS NULL),
       count(*) FILTER (WHERE rendered_html IS NULL)
FROM page_components WHERE build_status IS DISTINCT FROM 'removed';
-- 2295 | 2295 | 1064 | 10 | 0 | 0
```

`parent_instance_id` is NULL on **every live row** and `component_version_id` on 1,064 — and
both are `COALESCE(…::text,'')` in the projection, so they scan as `''`. The only bare columns
are `id` and `position`, both NOT NULL by schema. **The columns most likely to be NULL are
precisely the guarded ones.** The seat read the two-column control query as the whole argument;
the argument was always the eight-row table (NOTES above), and now every row of it has its own
measured control. Also note the live count moved 2194 → 2295 in a few hours — the census-ages-
by-addition rule demonstrating itself inside one day.

### The reuse_agent adjudication — one true sibling, one false one

The seat is right that three implementations of "compare counts, don't trust shape" now exist,
and right that an untracked residual is how duplication becomes permanent. But the three are not
one concept, and the adjudication splits them:

- **`scanBlogArticles` is a TRUE sibling** — a DB cursor with hand-rolled offered/kept counters.
  **Converge on next touch**, recorded AT ITS COUNTERS (where the next editor actually is), in
  `ScanShortfall`'s doc, and in DBI-027. Not converted now, for the reason already approved by
  the round: it is guarded, and a behaviour-neutral refactor widens a bug fix's blast radius for
  nothing. The convergence needs a graded variant of the helper, which ships WITH that first
  caller, not before it.
- **`collectPageSections` is a FALSE sibling, and saying so IS the tracking.** It compares the
  lengths of an in-memory metadata array — no cursor, no scan — and its response is to degrade
  loudly (fall back to whole-page grain), not to refuse. Routing it through `ScanShortfall`
  would use the error as a boolean and stamp *"refusing the partial result"* onto a path whose
  whole point is that it does not refuse. **Shared shape is not shared concept.** Written into
  the helper's doc so future convergence pressure cannot force the false unification.

### The fresh-eyes checks fable was asked for, run here instead

- **The tombstone refactor is a no-op for this lane, now committed by its owner** (`18853ade6`
  — the `datahelpers.NotRemoved("")` extraction). `NotRemovedSQL` is the byte-identical literal
  `build_status IS DISTINCT FROM 'removed'`, so the composed SQL is unchanged and the
  `removed_test` sqlmock pattern still matches what the code sends. Both guard test sets green
  against their commit.
- **Multi-loop files classify correctly**: first loop without a Scan → 1; two swallowing loops
  → 2. (The StillBites suite had no multi-loop case; the Python side is now spot-checked. Worth
  adding to StillBites next time that file is touched, not worth a commit alone.)
- **A deleted/renamed baseline file fails the ratchet with an explicit remove-its-line
  instruction** — by design, and it is the concrete form of guardian's CI-friction advisory: a
  lane renaming one of the 127 files owes a one-line baseline edit in the same commit. The
  failure message says exactly that, which is the mitigation.

---

## 2026-08-26 (post-roll) — THE FIX IS LIVE, verified at the artefact per RUNBOOK §12

Fresh chassis roll (both `agent-chassis` pods 10 min old, ReplicaSet `5864bf97c5`, one pod per
node). Verification `[MEASURED 2026-08-26, post-roll]`:

**The provenance line was NOT reachable** — absent from `--tail=2000` on 10-minute-old pods,
exactly the scroll behaviour the RUNBOOK warns about. Recorded as "not in range", not
"unstamped", and step 2 carried the verification instead, as designed.

**Three-way probe, BOTH pods (different nodes — the cached-image landmine means one tag can
serve different bytes per node, so one pod's probe does not vouch for its sibling):**

| probe | pod 5l8xd | pod 68t5h |
|---|---|---|
| `refusing the partial result` (capability under test) | **GUARD-PRESENT** | **GUARD-PRESENT** |
| `rerender_page_sections: row scan failed` (must-present control) | CONTROL-PRESENT | *(not re-run; same probe mechanism proven on 5l8xd)* |
| `xq410zz-not-a-real-marker` (must-absent control) | CONTROL-ABSENT | CONTROL-ABSENT |

**Step 3, guard silence: 0 refusals on both pods — and 0 `rerender_page_sections` invocations
in the window, so that zero is UNDEMANDED at the live tier.** Stated plainly rather than read as
proof: a zero with no traffic distinguishes nothing (the a-post-fix-zero-needs-a-demand-control
trap). The demand control is where §12 puts it: the mutation-proved test suite fires the guard
on every build, and the live zero will become meaningful as rerender traffic flows. Nothing in
this verification claims otherwise.

**Consequence for the estate:** instance 3 of `bugs_open/410` is now **fixed AND live**. The
blocking ratchet is live in every future build; the advisory twin was already live. The bug file
stays OPEN (pattern file — instances 1–2, candidates 1–2 and the content-loss residual are other
lanes' work or undecided), with instance 3 marked closed inside it.
