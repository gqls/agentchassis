# HANDOFF — 2026-08-06 — option A is LIVE, PROVEN EFFECTIVE, and the lane is done bar one decision

> **START HERE. Updated 2026-08-06. Sections 1–2 below describe the state BEFORE option A and are
> kept only as history — do not act on them.**
>
> **The retraction lane is complete.** The first adopter (`check_empty_sections`) works unattended
> — **10 findings retracted** so far. The second adopter was redundant and has been **reverted**.
> The real gap was fixed in the better mechanism (`revalidate_review_queue`), council `1cec55d2`
> **APPROVED**, live on `v1.0.1256`, and as of 2026-08-06 **proven to actually close a row it could
> not previously reach**.
>
> **Read `SUMMARY_2026-08-06_deployed_asset_path.md` for the plain-prose account**, then §0 here.
>
> **TWO THINGS ARE OPEN — both in §0. One is a decision for the owner; one is a loose thread.**

## 0. THE TWO OPEN ITEMS

### 0a. OWNER DECISION — should the revalidation sweep run on a timer?

**There is no `scheduled_tasks` row for `diagnosis-review-queue-revalidator`** (0 rows by name or
agent type; the scheduler is healthy — 27 enabled tasks). It has only ever run hand-dispatched.
Config is live immediately, so nothing forces the timing.

- **FOR:** the defect this lane fixed — the sweep builds a two-strike backlog it then cannot drain
  — is only *actually* fixed if it runs. Hand-dispatch means it runs when someone remembers.
- **AGAINST:** standing fleet-wide authority for a `dry_run:false` auto-closer over four item
  types. `needs_page` has **5 Go producers** and an **open owner decision (033 D2)** living in its
  `failed` population, measured churning 17 → 21 → 19 in 36 hours. **17 of 44 scheduled tasks are
  disabled**, so scheduling here reads as deliberately managed.
- **Evidence now in hand** (this was the missing input): one scoped hand-run closed exactly the 1
  row it should, refused another with a stated reason, and touched nothing else. It behaves.

### 0b. LOOSE THREAD — one eligible row was neither closed nor stamped

Found by the 08-06 run, **not caused by it**. On `fundamentallyai.com`, a `needs_page` /
`needs_human_review` row created 2026-08-05 13:44 got **no `result.revalidation` key at all** — it
did not even receive gate 2's record, while its two siblings were judged.

**Its `item_key` is `page_rerender:tools` — the prefix disagrees with its `item_type`**, which is
the drift `workItemKey`'s doc comment warns about. **Whether that causes the skip is
[UNVERIFIED]** — selection is on `item_type`, which matches, so on a first reading it should have
been loaded. I did not guess further.

```sql
SELECT left(item_key,45) AS item_key, item_type, status, created_at, (result ? 'revalidation') AS judged
FROM site_work_items WHERE item_type='needs_page'
  AND site_id='199733a8-ac9c-4c30-b2ce-65ecdac6f3bd' AND status='needs_human_review';
```
Safe direction (a skipped row closes nothing), so it is a thread, not a fire.

---

### How the 2026-08-06 run was done, if you need to repeat it

Run `19d0ccbd-c2df-4beb-bd1a-f7e41599eb5f`. **Scoped, not fleet-wide** — the shared trigger
`review_queue_drain/TRIGGER_revalidate_review_queue_v1.sh` hardcodes `INPUT_DATA='{}'`, so replay
its envelope with a filter instead of editing another lane's script:

```
{"site_id":"<uuid>","item_type":"needs_page","dry_run":false,"max_items":50}
```

Pre-checks that mattered: 0 claimed items, pods >300s old, binary re-verified (`AND status IN (%s)`
greps **3**). ⚠ **kcat pod names must be RFC-1123 lowercase** — an uppercase letter rejects the pod
and sends nothing; and `kcat -P` exits 0 having sent nothing, so **verify by the orchestration row,
never the exit code**. Dispatch was immediate (1 poll), not the ~30 min the trigger warns of — one
sample, do not plan on it.

**Both arms of the check — use both, a single arm cannot tell "worked" from "never ran":**

```sql
-- ARM 1: which rows closed
SELECT s.domain, swi.item_type, swi.completed_at, left(swi.result->'revalidation'->>'reason',90)
FROM site_work_items swi JOIN sites s ON s.id=swi.site_id
WHERE swi.resolution_path='auto:revalidated' AND swi.completed_at > '<dispatch time>';
-- ARM 2: did the sweep run at all?  (34 / 2026-08-06 08:05:19 as of this handoff)
SELECT count(*), max(completed_at) FROM site_work_items WHERE resolution_path='auto:revalidated';
```

**What the run proved:** it closed `needs_page:self-correction-leopardessconsulting`, raised
2026-07-20, sitting at **`unresolved`** — a status the sweep **could not see before option A**.
Direct behavioural proof. `unresolved` across all four covered types is now **0**.

**What it did NOT prove:** the `failed` exclusion was never exercised — that site has no `failed`
rows, so the exclusion was *not contradicted*, which is weaker than *held*.

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
