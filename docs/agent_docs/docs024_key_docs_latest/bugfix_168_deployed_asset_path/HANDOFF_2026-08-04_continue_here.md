# HANDOFF — 2026-08-04 — the second adopter is LIVE, APPROVED, and its justification was FALSE

**Supersedes `HANDOFF_2026-08-03b_continue_here.md`.** That file is still worth reading for the
first adopter's history and its §3/§4 traps — but **§4.1's trap list and §3's producer check are
both now known to be incomplete, and this file says how.**

Read this, then `WRONG_CALLS.md` (2026-08-04 entry), then `NOTES_deployed_asset_path.md` (bottom).

**Nothing here is half-applied. Everything described is committed and live. The open item is a
DECISION, not an implementation.**

---

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

## 4. THE DECISION OWED — three options, costed. This is the only open item.

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
