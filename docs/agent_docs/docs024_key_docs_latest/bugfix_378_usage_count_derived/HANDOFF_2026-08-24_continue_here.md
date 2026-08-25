# HANDOFF — `bugfix_378_usage_count_derived`, 2026-08-24 (cold start: read this top to bottom)

> ## ✅ LANE COMPLETE, 2026-08-25 — `bugs_open/378` is CLOSED. Read this block, then stop.
>
> **Fixed, live AND demand-proven.** The file moved to
> `bugs_closed/378_HANDOFF_2026-08-24_component_usage_count_is_written_on_one_of_two_resolution_paths_and_scored_on_the_other.md`
> — `git log` on the old `bugs_open/` path will not find the closure. Live on chassis
> **`635f2d32f5bbe3789867a978284c9c125d718eb0`** (2026-08-25 08:49:31Z); first shipped on `48f55f21…`
> 2026-08-24 18:55:21Z. Commit `5074367f7`, council `ca01b81a` APPROVED.
>
> **The evidence gap this file was written around is CLOSED.** `[MEASURED 2026-08-25]` since the fix
> went live: **403** `page_components` rows across **125** pages, **73** `page-build-handler` runs,
> **880** `page-rerender` runs, **17** `component-creator` runs (the only consumer of the changed
> contract reader) and **5** components born — and all 12 counted components read **byte-identically**
> to the pre-fix snapshot. `needs_new_component` ran **5** against **6** the prior day, so nothing
> broke. §4 below is spent; it is kept as the record of what was owed.
>
> ## ⏳ THE ONE REMAINING ACTION — apply migration `610`, but NOT YET: the fleet is split
>
> **Checked 2026-08-25 evening against a fresh build and REFUSED.** `[MEASURED]` the fleet was
> running **two** chassis builds at once — `4c996e1b5…` on **139** pods (does NOT contain the fix)
> and `a7459a44b…` on **12** (does). Applying would have broken component creation on 92% of pods.
>
> **The precondition in the first version of this handoff was wrong** — *"a build containing the fix
> is LIVE"* was TRUE and still unsafe. It needed a quantifier: **EVERY live build must contain it.**
> 610's header now carries the fleet-wide enumeration; use it, not this summary:
>
> ```sql
> SELECT git_commit, count(DISTINCT pod_name) AS pods FROM service_binary_capabilities
> WHERE service LIKE '%chassis%' AND last_seen_at > now() - interval '30 minutes'
> GROUP BY 1 ORDER BY 2 DESC;
> ```
> `git merge-base --is-ancestor f403113f4 <X>` must be **YES for every X**, and the binary probe must
> show ABSENT on a pod of **each** build. One NO anywhere → do not apply.
>
> ⚠ **How it was caught, because it is the reusable part:** the ancestry check and the binary probe
> DISAGREED — the stamp query (ordered by `last_seen_at`) returned an old-build pod while
> `-l app=agent-chassis` returned a new-build one. **If they disagree, STOP; do not pick the
> convenient one.** A redundant check earns its place when it can disagree.
>
> ### (the original framing, kept — its precondition is the one that was wrong)
>
> #### apply migration `610` after the next chassis roll
>
> Everything else in this lane is done. `610_content_components_drop_dead_usage_count_HOLD.sql`
> drops the dead column and is **written, tested and deliberately unapplied**. It is `_HOLD` because
> a banner cannot hold a file back from the runner.
>
> **Precondition:** a chassis build containing commit `2c1a5d0…`-era part 1 (the removal of the birth
> INSERT's `usage_count` from `store_generated_component_action.go`) must be LIVE. Against an older
> binary the DROP makes **every component creation fail**. The artefact check with its control is in
> 610's own header — and the control is not optional: this lane's first ancestry control was itself
> worthless because it also predated the build.
>
> Then apply **by hand** (`608` was pending from another lane; `--apply` takes every pending file) and
> record with `--record-only` + `--note`. Its guard aborts if `usage_count` has moved since it was
> killed; that guard has been induced and fires correctly.
>
> **Done already (2026-08-25):** the last writer is removed, and migration **`609` is APPLIED,
> recorded and verified at the artefact** — the column comment used to read *"Times this component
> has been assigned to a page. Incremented by selector. Higher = more battle-tested."* (false in all
> three clauses) and now begins `SUPERSEDED AND DEAD`. Council `ac7b62e6` (follow-up round).
> ⚠ `agent_definitions.usage_count` is a DIFFERENT column, is LIVE, and is deliberately untouched.
>
> **Not a residual of this lane:** the 27-section_type contract mismatch is now **`bugs_open/388`**,
> filed 2026-08-25 with its own evidence and `[UNMEASURED]` list. 378 *reduced* it from 29 to 27.
> Do not re-derive it here.
>
> ⚠ **One check that looked like evidence and was not**, recorded so it is not repeated: grepping
> `-l app=agent-chassis` logs for selector errors returned `0` — and so did the control asking whether
> those logs contain *any* selector lines at all. The work runs in dynamically-named pods
> (`agent-page-content-writer-…`), not the two pods carrying that label. **That zero was vacuous.**
>
> ## (original 2026-08-24 state block, kept as the record)
>
> ### STATE: FIX IS LIVE AND PROVEN AT THE BINARY. Lane is NOT closeable yet — one evidence gap, three residuals.
>
> `bugs_open/378` — `content_components.usage_count` was written on one of three resolution paths and
> read as a merit signal. Fix committed `5074367f7`, council **`ca01b81a` APPROVED** round 1, live on
> chassis **`48f55f21834ac3e2d95aa43716f6e63e40ac12ee`** (pod started 2026-08-24 18:55:21Z).
>
> **Do not re-derive the diagnosis. It is complete and in NOTES.** What is left is listed under
> §What to do next, and it is short.

## 1. What the bug was, in one paragraph

`content_components.usage_count` had exactly **one** incrementing writer (`IncrementUsageCount`),
reachable from exactly **one** of the **three** resolution paths in `plan_sections_action.go`'s
section loop — Path 0 stored `component_id`, Path 1 name/function, Path 2 the `section_type` selector.
Only Path 2 counted. The same column was read as a merit signal by **three** readers. So it recorded
*which route found a component*, not whether the component is any good. It also **over**-counted: the
increment fired before `planSection` decided ready/deferred/skipped and again on every re-plan, so the
column's two largest values belonged to components with **zero** page bindings.

## 2. What shipped

1. `IncrementUsageCount` **deleted** with its only call site.
2. `ComponentUsageSitesSQL` — **one** named constant in `component_selector.go`,
   `count(DISTINCT p.site_id)` over `page_components` JOIN `pages`, excluding `build_status='removed'`.
3. **The selector's scoring term is REMOVED, not repaired.** The derived figure is still SELECTed and
   logged; nothing scores on it.
4. `load_existing_component_action.go`'s **contract-row** `ORDER BY` uses the constant.

**Why removal, not repair — this is the decision to understand before touching anything:** simulated
over the 4,888 contested `(section_type, site_type, page_type)` contexts, *removing* the term changes
**0** winners; *repairing* it changes **3,246**. And a working usage term is a preferential-attachment
loop (selected → count rises → scores higher → selected again), which is exactly what `bugs_open/107`
("every site gets the same homepage skeleton") is the open complaint about, citing this file.
**Repairing the measurement would have armed the homogeneity engine.**

## 3. Proof it is live (each arm has a control — reuse these, do not invent new ones)

```bash
# the stamp (the startup log line has scrolled; this table has no shelf life)
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c \
 "SELECT service, git_commit, started_at FROM service_binary_capabilities
  WHERE service LIKE '%chassis%' ORDER BY last_seen_at DESC LIMIT 1;"

git merge-base --is-ancestor 5074367f7 48f55f21834ac3e2d95aa43716f6e63e40ac12ee   # YES
git merge-base --is-ancestor 4ad5b10fb 48f55f21834ac3e2d95aa43716f6e63e40ac12ee   # NO  <- the control

POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
kubectl -n ai-persona-system exec $POD -- grep -aq 'count(DISTINCT p.site_id)' /proc/1/exe        # PRESENT
kubectl -n ai-persona-system exec $POD -- grep -aq 'usage_count, 0)::float / 50.0' /proc/1/exe    # ABSENT
kubectl -n ai-persona-system exec $POD -- grep -aq 'component_selector: selected' /proc/1/exe     # PRESENT (control)
```

⚠ **My first ancestry control was worthless** — I picked a commit that also predated the build, so
both arms returned YES. The control above (`4ad5b10fb`, committed *after* the build) is the corrected
one. If you re-verify against a *later* build, pick a fresh control the same way.

## 4. ⚠ THE ONE EVIDENCE GAP — it is why this is not closed

**The fix is not demand-proven.** `usage_count` values are byte-identical to the pre-fix snapshot,
which is the expected post-fix observation and **currently proves nothing**, because:

```sql
SELECT count(*) FROM page_components WHERE created_at > '2026-08-24 18:55:21+00';  -- 0
```

**No page has been built since the roll**, so the old code would have incremented nothing either.
The frozen counter becomes evidence only after a page build has actually run.

**To close this gap, in order of preference:**
1. **Wait for organic traffic**, then re-run both queries. The 12 counted components must still read
   `20, 19, 12, 7, 4, 2, 1×6` — any increase means the old binary is still serving somewhere.
2. If you need it sooner, drive one page build/rerender on a site whose homepage uses a
   `section_type` with a selector fallback, and check the chassis log for
   `plan_sections: resolved via section_type selector` — that line proves Path 2 ran, which is the
   only path that used to count.

## 5. Residuals — none blocking, all stated rather than closed by assertion

| # | what | severity | state |
|---|---|---|---|
| A | **27 of 117 section_types predict a contract the store would not enforce.** `load_existing_component`'s advisory names one row's fields while the store overwrites the row whose `function = NormaliseToKebab(section_type)`. **Pre-existing, NOT created by this change** — this change *reduced* the mismatch from 29 to 27. **Not filed as a bug yet; probably should be.** | medium | OPEN, unowned |
| B | **`content_components.usage_count` still exists**, written by nothing and read by nothing in Go. Deliberately deferred so the code cannot roll back onto a missing column. Now that the code IS live, a migration to `COMMENT ON COLUMN` (or drop) is safe — the `guidelines` seat suggested the comment. ⚠ **Until then it still reads as a maintained figure.** | low | OWED |
| C | `reuse_agent`'s objection was **valid** — prior art exists (`component_write_guard.go:437`, `store_generated_component_action.go:1179`) and I had not checked. Not reused deliberately (they compute a write-fence blast radius over all rows; this is a merit signal excluding `removed`), reasoning recorded in NOTES. | low | ANSWERED |

**Objections resolved, for the record:** `editquality` (no test call sites existed — `go vet` clean);
`guardian` (cost **measured**: 10.7 ms batch, index-backed, vs 0.36 ms stored — ~30× relative, ~10 ms
absolute, once per page build); `prior_art_librarian` (CLC-026 figures are mine with a control, and
the register bullet is commit `1103c5cbd`); `architecture` (`ARCHITECTURE_SIGNAL: point_fix`; the
in-code TODO it asked for was already in the shipped comment); `bug_historian` — **INVERTED**, see below.

## 6. The `bug_historian` objection is the most interesting thing in this lane

It warned that switching the contract row could silently re-shape the enforced schema for pages bound
to the old winner — the family that has bitten this platform before. Right shape, **backwards
direction**, and reading the file's own fallback is what showed it.

`resolveContractViaStorageIdentity`'s comment states the design intent: the function name is derived
*"exactly as `store_generated_component_action.go` derives it"* so *"the prediction and the enforcement
agree by construction rather than by coincidence"*. The store resolves what it overwrites by
**function name**. So the test is not "did the winner change" but "does the winner's `function` equal
what the store enforces":

| ordering | agreements, of 117 section_types |
|---|---|
| OLD | **88** |
| NEW | **90** |

Both changed types moved **from disagree to agree** (`hero`: `hero-about`→`hero`;
`tool-archetype-taster-quiz`: `archetype-taster-quiz`→`tool-archetype-taster-quiz`).

## 7. Other lanes — already told, nothing owed

- **`bugs_open/107`** — carries a cross-thread note plus an addendum. **Its most useful content is a
  finding, not a warning:** only **4** section_types have more than one candidate, so the sameness is
  *one candidate per slot*, not a scorer with a favourite. No scoring change can fix that.
- **`bugs_open/357`** — their mis-bound rows (**22**, all `hero`) become bindings under the new
  substrate; under DISTINCT SITES they collapse to **3 of hero's 27** and are functionally inert. Their
  mint is still open (phase 2 default-OFF), so treat it as a floor with a growth rate.
- **`bugs_open/381`** — no collision (different files); I corrected my own 107 note's scope for them,
  because the incumbency trap is a Path-2 property and does not bite their planner-level fix.
- **register `CLC-026`** carries a downstream-consumer bullet naming this lane, with the switch
  condition for adopting the provenance stamp later (**not** a coverage percentage — see the entry).

## 8. Where the record is

- `NOTES_378_usage_count_derived.md` — evidence, every measurement, every misstep. **Read §"the fix,
  and the misstep that produced it"** before trusting any number in here.
- `README_where_we_are.md` — the owner's plain-prose log.
- `WRONG_CALLS.md` 2026-08-24 — I measured what *removing* a term does and used it to license
  *replacing* it (0 vs 3,246). **A measurement licenses exactly the diff it simulated.**
- `LANDMINES.md` — the prospective entry, added once the fix was live so it could name the remedy.
- `016b` §9 — the transferable pattern (written by the 351 lane when it filed the bug).

## What to do next

1. **Close the demand gap** (§4). This is the only thing standing between the lane and closure.
2. **Then move `bugs_open/378` → `bugs_closed/`** — the bar is fixed AND live, and it is both; the
   demand check is this lane's own stricter standard, and I would hold to it.
3. **Write residual B's migration** (comment or drop the dead column) — small, safe now the code is live.
4. **Consider filing residual A** as its own bug. It is a real prediction/enforcement mismatch on 27
   section_types and nobody owns it.
