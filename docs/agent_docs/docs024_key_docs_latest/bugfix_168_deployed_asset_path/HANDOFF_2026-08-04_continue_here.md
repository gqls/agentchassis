# HANDOFF — 2026-08-06 — the retraction lane is DONE. One inherited limit is open.

> **START HERE. Sections 1–8 below are HISTORY (pre-option-A) — read for context, do not act on
> them.** Plain-prose account: `SUMMARY_2026-08-06_deployed_asset_path.md`.

**Everything this lane set out to do is finished, live, council-approved and proven by effect.**

| | |
|---|---|
| First adopter `check_empty_sections` | **Working unattended** — 10 findings retracted, drawing down as sites are swept |
| Second adopter (`check_required_fields_missing`) | **Reverted** — it duplicated `revalidate_review_queue`. See `WRONG_CALLS.md` 2026-08-04 |
| The real gap (sweep built a queue it could not drain) | **FIXED** in `revalidate_review_queue`, council `1cec55d2` APPROVED, live `v1.0.1256` |
| Proven effective | **Yes** — a scoped hand-run closed `needs_page:self-correction-leopardessconsulting`, raised 2026-07-20, sitting at `unresolved`, a status the sweep could not previously see |
| Scheduled | **Yes, owner-approved 2026-08-06** — `scheduled_tasks.review-queue-revalidate-daily`, 86400s, verified end-to-end |
| `unresolved` across the four covered types | **0** |

## 0. THE OPEN ITEM — the sweep does 94% wasted work and cannot reach its own tail

**Not caused by this lane; exposed by it.** The first scheduled run measured it exactly:

```
scanned 500 · resolved 0 · still_holds 31 · unknown 469 · dry_run false
```

- `loadParkedReviewItems` takes the **oldest N across ALL parked rows**, with no type filter.
- **779 rows are parked; only ~168 are in the four types `reviewRevalidators` covers.**
- The other ~611 return `unknown` ("no revalidator for this type"), which is **deliberately
  non-terminal** — so they stay parked, stay oldest, and are re-selected every single run.
- Result: the same ~500-row head is re-judged daily, **279 rows are never reached at all**, and
  94% of each batch is work that cannot possibly resolve.

⚠ **YOU CANNOT FIX THIS FROM `scheduled_tasks`. I tried and measured it.** The action reads
`max_items` from the **step config** (`GetIntField(config, "max_items", 50)`) and the `sweep` step
has **no `input_mapping`**, so anything in `scheduled_tasks.input_data` is inert — I set 1000 and
the run reported `capped_at: 500`, the step config's value. The row's `input_data` is now `{}` and
its `description` says where the real knob is.

**Two honest fixes, both an `agent_definitions` change (config — live immediately, no build):**
1. **Add an `item_type` filter** so the sweep only loads types it can actually judge. Cleanest, and
   it makes the coverage gap visible rather than absorbed as wasted scans. But the action takes a
   single `item_type`, so this means either several scheduled rows or widening the action.
2. **Raise the step config's `max_items`** past 779 (with headroom). One-line, but it treats the
   symptom — the uncovered head still grows.

**Recommendation: (1), as several scheduled rows — one per covered type.** It is the only version
where a growing uncovered backlog cannot silently push covered work out of the batch.

```sql
-- size it before choosing
SELECT count(*) FILTER (WHERE item_type IN ('required_fields_missing','needs_section_data','unresolved_cta','needs_page')) AS covered,
       count(*) AS total_parked
FROM site_work_items WHERE status IN ('needs_human_review','unresolved');
```

### 0b. Loose thread — one eligible row was neither closed nor stamped

On `fundamentallyai.com`, a `needs_page`/`needs_human_review` row created 2026-08-05 13:44 got **no
`result.revalidation` key at all**, while its siblings were judged. **Its `item_key` is
`page_rerender:tools` — the prefix disagrees with its `item_type`**, the drift `workItemKey` warns
about. **Whether that causes the skip is [UNVERIFIED]** — selection is on `item_type`, which
matches. Safe direction (a skipped row closes nothing), so a thread, not a fire.

## 0c. How to verify the schedule is still working

⚠ **`last_triggered_at` is written by the scheduler at publish time and is NOT proof the agent
ran.** Take the chain:

```sql
-- 1. did the scheduler publish?
SELECT last_triggered_at FROM scheduled_tasks WHERE name='review-queue-revalidate-daily';
-- 2. did an agent actually run?  (should appear within ~1s of the above)
SELECT orchestration_id, current_step, status, created_at FROM orchestration_states
WHERE orchestration_name ILIKE '%reval%' ORDER BY created_at DESC LIMIT 3;
-- 3. what did it do?  (note the key is revalidation_result, NOT sweep.result)
SELECT collected_data #>> '{revalidation_result,scanned}'  AS scanned,
       collected_data #>> '{revalidation_result,resolved}' AS resolved,
       collected_data #>> '{revalidation_result,unknown}'  AS unknown,
       collected_data #>> '{revalidation_result,capped_at}' AS capped_at
FROM orchestration_states WHERE orchestration_id='<from step 2>';
-- 4. cumulative effect (34 as of 2026-08-06 08:05:19)
SELECT count(*), max(completed_at) FROM site_work_items WHERE resolution_path='auto:revalidated';
```
**`resolved: 0` is not a failure** — it means nothing was closable. Only step 2 distinguishes that
from "it never ran", which is why a one-armed check is worthless here.

## 0d. The three lessons this lane paid for — read before adopting the seam anywhere else

1. **Count what already CLOSES an item type, not just what produces it.** The mandatory §3 producer
   check has a producer-shaped hole; it passed twice while the real problem sat at the other end.
   `SELECT status, count(*), left(result::text,120) FROM site_work_items WHERE item_type='<t>' AND
   status IN ('complete','verified') GROUP BY 1,3;`
2. **When you widen a shared mechanism, the blast radius is rarely on the consumer you are thinking
   about.** Three times in three days. The one time it was caught pre-ship, it prevented pointing an
   auto-closer at 21 rows under an open owner decision.
3. **A dispatcher has two gates and the one you can reach is not the one that decides.** Three
   instances here: `input_mapping` vs a claim query's `RETURNING`; the sweep's selection vs its two
   CAS guards; `scheduled_tasks.input_data` vs step config.

## 1. The one-paragraph version

`check_required_fields_missing` became the second adopter of the RFC_010 retraction seam. It was
approved by the council at round 1, shipped on `v1.0.1250`, and is pod-verified on both replicas.
**It is also redundant**, because a mechanism I did not find — `revalidate_review_queue` — has
been closing exactly these items, with the same predicate, since 2026-07-27. My submission's
central claim ("nothing can close these") was false, 14 council seats accepted it, and it was
caught only by re-running the sizing after the roll. The code is harmless and inert; what it needs
is an owner decision, below.

## 2. State — verified 2026-08-04, chassis `v1.0.1250`, both replicas

| thing | state |
|---|---|
| `check_required_fields_missing` retraction | **LIVE AND VERIFIED.** Council `64430363` APPROVED r1 |
| Commits | `ba3aae47f` (code) · `8b25fb4e0` (objection answers) · docs `b312c409a`, `9cd9c5227` |
| Retractions it has made | **0**, and on current populations it will make none — see §3 |
| Fleet `result ? 'resolved_at'` | **8** (relojistas 1, gaswholesalers 3, leopardess 4) — **all `empty_section`**, i.e. all from the FIRST adopter |
| First adopter (`empty_section`) | **Working.** Was 4 on 08-03, now 8 — the mechanism is drawing down as sites get swept |
| `required_fields_missing` population | 78 `needs_human_review`, 17 `complete` (was 59/11 on 08-03) |
| **The decision owed** | **§4. Do not skip it — it is the only open item.** |

**Deploy verification, and the two things it taught:**

```bash
for POD in $(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name | sed 's|pod/||'); do
  A=$(kubectl -n ai-persona-system exec $POD -- sh -c "strings /app/agent-chassis | grep -c 're-observed filled'" 2>/dev/null|tail -1); A=${A:-0}
  B=$(kubectl -n ai-persona-system exec $POD -- sh -c "strings /app/agent-chassis | grep -c 'filing stops here'" 2>/dev/null|tail -1); B=${B:-0}
  echo "$POD adopter=$A capfix=$B"
done
```

⚠ **Use ASCII-ONLY substrings.** My first probe grepped for a string containing an **em-dash**;
it got mangled crossing the `kubectl exec -- sh -c` boundary and returned **0**, which reads
exactly like "my change did not ship". The string was there. Never let a non-ASCII character
cross that boundary.

⚠ **There was no valid negative control, and I nearly pretended otherwise.** The old cap log line
is a strict *prefix* of the new one, so grepping the old matches either binary. For a purely
additive change the dated **0→1 transition is the whole proof** — and it exists only because it
was taken **before** the roll. `re-observed filled` was 0 on both replicas of `v1.0.1244`
(2026-08-03) and is 1 on both of `v1.0.1250`. If you ship an additive change, take that baseline
first; you cannot get it afterwards.

## 3. THE THING TO UNDERSTAND BEFORE TOUCHING THIS SEAM AGAIN

**§3 of the previous handoff makes counting PRODUCERS mandatory. Count the CLOSERS too.**

I ran the producer check thoroughly and two independent ways (grep for `ItemType:` literals;
corroborate with `created_by` over every row ever). Single producer, confirmed. **It could never
have surfaced the actual problem**, which is at the other end:

`platform/orchestration/actions/revalidate_review_queue_action.go:161` —

```go
"required_fields_missing": revalidateNamedFields("missing_fields"),
```

- It selects `status = 'needs_human_review'` — **the status every one of these items is born in**,
  so it covers **100%** of the population I claimed was unreachable.
- Its predicate is mine: *"every field this item reports missing (%s) is populated on the deployed
  component"*.
- It has been running since at least **2026-07-27**.
- **It is better than mine in two ways.** It refuses when `content_data` is empty (*"the component
  renders from a template, a DERIVED source or a static fallback… content_data cannot answer the
  question, so we do not pretend it can"*) — a refusal I do not have. And it keys the component
  lookup on `(page_name, slot)`, **never** on `spec.component_id`, having measured that component
  ids are unstable across re-renders (11 of 45 `required_fields_missing` items resolved to a
  component that no longer existed when keyed that way).

**The ten-second check that would have caught it, before any code:**

```sql
SELECT status, count(*), left(result::text,120)
FROM site_work_items WHERE item_type='<your type>' AND status IN ('complete','verified')
GROUP BY 1,3 ORDER BY 2 DESC;
```

**Ask what closed the ones that are already closed, before claiming nothing can close them.** A
`result.revalidation` block is the tell. I had even *seen* the count — I measured "11 complete" the
previous afternoon and read it as history rather than as a mechanism running weekly.

Also in `LANDMINES.md` (2026-08-04) with the code-side grep that finds closers, and in
`WRONG_CALLS.md` with the full account of how an *evidenced* claim was still wrong: I proved
"the handler dispatch path cannot close it" and stated "nothing can close it".

## 4. ~~THE DECISION OWED~~ — **DECIDED 2026-08-04: OPTION A. DONE, committed `b4c64f433`.**

> **The owner chose option A: teach the revalidator `unresolved`, revert my adoption.** Both halves
> are committed; council `1cec55d2-5928-4785-8598-dfd7870a39d8` submitted alongside. **Not live** —
> both replicas run `v1.0.1251`, which still carries the reverted adoption. Needs the next roll.
>
> **Three things changed from the plan below, all from reading the mechanism rather than the
> summary:**
>
> 1. **The gap is not "empty today" — it is imminent.** `revalidate_review_queue`'s own closes feed
>    `insertWorkItem`'s two-strike counter, so after the second close in 7 days the third re-raise is
>    born `unresolved`, which the sweep could not see. **It generates the rows it goes blind to.**
>    5 `required_fields_missing` keys already sit at 1 strike.
> 2. **`status` is checked in THREE places**, not one — the selection plus two write-time CAS
>    guards. Widening only the selection selects the new rows and silently updates none.
> 3. **`failed` was EXCLUDED, against my own first draft.** RFC_010 Decision 2 pairs it with
>    `unresolved` and I copied the pair. Measuring all four covered types showed the blast radius is
>    on **`needs_page`: 17 `failed` rows** — the population this action's header defers by name to
>    **owner decision 033 D2**. Narrowed radius: **1 row**. Preventive, not a drain.
>
> Full working in `NOTES_deployed_asset_path.md` (bottom), including a mutation harness that was
> **invalid on its first run** and how.
>
> **Council `1cec55d2` APPROVED it** (`e35044546` answers the objections). Two are worth carrying:
> the `editquality` seat caught that **my own negative control was the shape I keep writing
> landmines about** — it counted the removed literal over the whole file, so a future comment
> explaining the removal would fail the test on a correct fix (now strips `//` lines; proven in both
> directions). And the `guardian` seat asked the co-dedup question and **found something**:
> **`needs_page` has 5 Go producers**, and it is the only one of the four types with a live row in
> scope (1). Pre-existing in kind — the revalidator already serves that type — but the
> `bugs_open/187` lane is told.
>
> **VERIFY AFTER THE NEXT ROLL. This is the one change in this lane with a REAL negative control**,
> because a revert removes a string:
> ```bash
> for POD in $(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name | sed 's|pod/||'); do
>   N=$(kubectl -n ai-persona-system exec $POD -- sh -c "strings /app/agent-chassis | grep -c 're-observed filled'" 2>/dev/null|tail -1); N=${N:-0}
>   P=$(kubectl -n ai-persona-system exec $POD -- sh -c "strings /app/agent-chassis | grep -c 'auto:revalidated'" 2>/dev/null|tail -1); P=${P:-0}
>   echo "$POD removed_expect_0=$N positive_control_expect_nonzero=$P"
> done
> ```
> Baseline dated **2026-08-04, `v1.0.1251`, both replicas: `re-observed filled` = 1, `auto:revalidated` = 2.**
> After the roll the first must be **0** and the second must stay non-zero — a 0/0 means the probe
> broke, not that the revert landed.

## 4b. The original options, kept because the reasoning is the record

The change is **redundant, not harmful**: `resolveWorkItems` skips rows already in
`workItemClosedStatuses`, so the two closers cannot double-close or clobber each other.

**The one genuine gap is narrow and empty today.** `revalidate_review_queue` only ever looks at
`needs_human_review`. `resolveWorkItems` deliberately also closes **`unresolved` and `failed`**
(RFC_010 Decision 2 — neither status means "this stopped being a problem"). **0 of the 95
`required_fields_missing` rows are in those statuses today**, so the gap is real and has never
been exercised. It becomes reachable the moment the two-strike counter starts marking these
`unresolved` — which is RFC_010 **Q1**, already tracked.

| option | cost | what it buys |
|---|---|---|
| **A. Teach the revalidator `unresolved`/`failed`, revert my adoption** *(my recommendation)* | one predicate widened in `revalidate_review_queue_action.go`; one revert | **One closer, and it is the better one** — it has the `content_data`-empty refusal and the stable `(page_name, slot)` key that mine lacks. Closes the gap at the mechanism that already owns the type |
| **B. Keep both, documented** | zero now | Keeps a second closer on a different cadence (discovery sweep vs review-queue pass). Costs a permanent duplicate predicate — the exact drift class this estate keeps paying for, and the one `findResolvedRequiredFields`' own comments argue against |
| **C. Revert, close nothing** | one revert | Simplest. Leaves the `unresolved`/`failed` gap open for RFC_010 Q1 to deal with |

**I did not choose unilaterally.** A revert is also a change needing justification, and option A
touches a mechanism I did not write, on a shared path, which is exactly the `bugs_closed/124`
shape this lane keeps getting told about. **Owner call.**

## 5. What is next, after §4 is decided

### 5.1 The first adopter is working — watch it, do not re-measure it
`empty_section` retractions went **4 → 8** without intervention (relojistas +1, gaswholesalers
+3). That is the seam doing its job on its own cadence. The remaining ~10 close as their sites get
swept.

```sql
SELECT s.domain, swi.item_type, swi.result->>'resolved_by', count(*)
FROM site_work_items swi JOIN sites s ON s.id=swi.site_id
WHERE swi.result ? 'resolved_at' GROUP BY 1,2,3 ORDER BY 4 DESC;
```

### 5.2 More adopters — but the entry bar has gone up
The remaining candidates from the old §4.1 are `cta_names_unknown_destination` (107 open, and
still being filed), `needs_sprite_css` (10), `voice_tells` (25). **Before any of them: run the
CLOSER check in §3 as well as the producer check.** `cta_names_unknown_destination` is
`needs_human_review` too, so it is squarely in `revalidate_review_queue`'s selection — check its
`reviewRevalidators` map first, because if the type is in there the answer is probably "extend
that, not this".

⚠ `check_image_url_404` — another lane is still editing it. Co-ordinate.

### 5.3 The armed sibling, censused so nobody has to do it again
Five discovery checks carry a per-pass cap. Three already `break`. Mine was fixed. **Exactly one
remains armed: `check_image_source_unsatisfiable.go` (`return result, nil`).** It is **inert
today** — that check does not populate `Resolved`, so its `return` is currently correct. **The
commit that adopts the seam there is the commit that must also change it to `break`.** In
`LANDMINES.md` with the re-run command.

### 5.4 Still open, unchanged
`bugs_open/179` (unowned). Decision 2's dedup half (blocked on the 87 duplicate rows). RFC_010 Q1
(two-strike; owner-ruled accept-as-is and tracked) — **note it is now load-bearing for §4**, since
it is what would make the `unresolved`/`failed` gap real.

## 6. What the council round taught (worth keeping regardless of §4)

**APPROVED r1, 14 seats, one objecting seat, none high.** Three objections were checkable and were
checked rather than filed — that is the third round running where that was the right move.

- **`bug_historian` (medium)** asked whether other checks carry the same armed cap. Answered with
  the census in §5.3 rather than an intention.
- **`bug_historian` (low)** said the "healthy++ from one place" discipline was held by review, not
  mechanism. Now mechanical: `TestHealthyIsIncrementedFromExactlyOnePlace` parses the source with
  `go/ast` (not a grep — the file's own comments discuss `obs.healthy`) and cannot pass vacuously.
- **`editquality` (low)** flagged `build_status='deployed'`. The answer was structural, not
  statistical: `page_components` has **no separate status column** — retirement is
  `build_status='removed'`, *inside* the column being filtered. The `pages` trap needs two columns
  to drift apart.

⚠ **The finding worth carrying, and it is uncomfortable:** the seats that most wanted to check my
absence claim (`prior_art_librarian`, `tooling_provenance`) recorded that they **could not** —
`code_checks` cannot see function bodies. `prior_art_librarian` then praised the claim as
*"unusually well-evidenced for an absence claim"*. **A council cannot falsify an absence claim
about code it cannot read, and it will reward you for citing evidence for a narrower proposition
than the one you stated.** The gate is not a substitute for the ten-second query in §3.

## 7. Correlations

| what | id |
|---|---|
| council (APPROVED r1) — **rationale contains the false premise** | `64430363-a42a-4028-b84a-9a25ab707441` |
| first adopter's council (APPROVED r1) | `97923026-2b2d-4925-b9a3-de6f70c49d2b` |
| RFC_010 Decision 1 council | `846f4f3d-8958-4e4c-be81-d5f02e20852d` |

## 8. Environmental

**`/tmp` is a 16G tmpfs and has been near-full.** The Go linker writes there, so `go build ./...`
fails with `mapping output file failed: no space left on device`, which reads exactly like a code
error. `TMPDIR=/home/ant/.cache/buildtmp go build ./...` (235G free on `/`). Used throughout.

`gofmt -l` reports `check_image_url_404.go` and `check_misdirected_cta_test.go` dirty in the
shared tree — **another session's, not mine.** Leave them.
