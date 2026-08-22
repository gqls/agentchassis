# NOTES — bugs_open/308, CTA destination provenance (append-only, newest at the bottom)

## 2026-08-22 — lane opened; the bug re-verified before any design work

**Why a lane at all.** `bugs_open/308`'s routing line says *"Route to
`cta_target_content_pass` rather than opening a competing lane"*. Read against that lane's
own PLAN, the two are not the same object: that lane is a **content pass** (reword CTA
LABELS so the existing resolver picks better targets) and carries the widening only as a
phase-1 open question. 308's owner ruling of 2026-08-18 chose **fix candidate 1 — build a
real provenance record**, which is platform Go + a shared seam. This lane does that half.
The content pass is not competed with; a CONTRIB goes into its NOTES and into the bug file.

Checked before opening: `scripts/who-owns.py 308` → OWNED/recently-active
(`cta_target_content_pass`, 2 commits/14d, last 2026-08-18). Its files were read in full
before writing anything here. `git status` on the files this fix would touch
(`resolve_internal_links_action.go`, `rerender_page_sections_action.go`,
`discovery_checks/`, `datahelpers/`, `save_page_sections_action.go`,
`section_editor_actions.go`) → **clean**, so no in-flight session is mid-edit on them.
That check is LAGGING by construction and is re-run at each phase boundary.

### The bug is STILL VALID, and it has grown [MEASURED 2026-08-22, live DB]

The file's own query, re-run verbatim:

```sql
SELECT count(*) FROM site_work_items swi, LATERAL jsonb_array_elements(swi.spec->'findings') f
WHERE swi.item_key LIKE 'misdirected_cta:%'
  AND f->>'suggested_target' ~ '^/(contact|about|privacy|terms|legal)(\.html|/|$)';
```

- filed 2026-08-17: **149**
- today 2026-08-22: **200**

Split by status — and this split is the churn made visible:

| status | items | findings |
|---|---|---|
| `complete` | 71 | **112** |
| `unresolved` | 53 | 86 |
| `cancelled` | 2 | 2 |

**112 findings sit on work items the platform marked `complete`.** That is the bug's
sentence "the repair runs, completes green, and changes nothing" as a number.

### DEMAND CONTROL — the flow has stopped, and that must not be read as a fix

Before treating 200 as a current rate, I asked whether the detector is running at all
(`a-post-fix-zero-needs-a-demand-control`, and this is its mirror image — a *stock* read as
a *flow*):

```sql
SELECT date_trunc('day', created_at)::date, count(*) FROM site_work_items
WHERE item_key LIKE 'misdirected_cta:%' GROUP BY 1 ORDER BY 1 DESC;
```

→ 08-19: **3**, 08-18: 128, 08-17: 208, 08-16: 28, 08-15: 21, 08-14: 84 …

So **nothing has been detected for three days**. 200 is a stock. `[INFERRED, not verified]`
the cause is `bugs_open/230` (site discovery has no recurring driver) — I have not opened
230's mechanism to confirm it, and it does not change this bug's design either way. What it
DOES change: any post-fix measurement of this population is worthless without first
inducing a discovery run, because the number will sit at 200 whether the fix works or not.

### The finding I did not expect: `suggested_target` HAS NO CONSUMER

```
grep -rn "suggested_target" --include=*.go platform/ internal/ pkg/
```
→ three hits, ALL of them writers or a test: `check_misdirected_cta.go:130`,
`check_cta_nonpage.go:79-80`, `check_cta_nonpage_test.go:145`. **Nothing reads it.**

The detector emits a `page_rerender` item with `spec.reason = "cta_links_stale"`;
`rerender_page_sections_action.go:528` gates its CTA recompute on that *reason string* and
then re-derives the destination independently from `candidatesFromHubs`. So the bug is one
rung deeper than "two candidate universes": the detector's answer is **computed, written
down, and thrown away**, and the repairer re-answers the same question from less
information. Two universes is the symptom; *the repair not consuming the detection* is the
shape — and it is the same shape as `bugs_open/071` (the content gate detects every broken
link then discards the finding), which is open on a different seam.

That generalisation is this lane's argument for a framework-level fix rather than a CTA
patch, and it is the thing to hold the design to.

## 2026-08-22 (later) — a THIRD consumer of the candidate loaders, which no CTA doc names

Enumerating the seam's consumers before designing (the 2026-07-29 owner ruling wants the
other consumers NAMED, and enumerated by query rather than asserted):

```bash
for sym in candidatesFromHubs loadContentHubs loadInteractivePages storedCTADestinationIsAuthored \
           ctaExcludedDestination areasExcludedFromCTA BestLabelMatch ctaFieldNames; do
  grep -rn "\b$sym\b" --include=*.go platform/ internal/ pkg/ cmd/ | grep -v _test.go | sed 's/:.*//' | sort | uniq -c
done
```

`loadContentHubs` and `loadInteractivePages` have **three** callers, not two:

| caller | what it does with them |
|---|---|
| `resolve_internal_links_action.go` | build-path CTA resolution (`setCTAField`) |
| `rerender_page_sections_action.go` | repair-path recompute (`applyCTARecompute`) |
| **`render_site_components_action.go:182-190`** | **the site HEADER's CTA fallback** — `chooseCTATargets("", "", interactive, hubs)` |

The third is the one that matters, and **`bugs_open/308`, `bugs_open/248` and LNK-033's
landmine all name only the first two** ("this set is shared by BOTH writers"). Greps for
`render_site_components` in both bug files return **0**. The register knows the header CTA
exists (LNK-030 lists it as a `ChromeLinkPolicy` consumer) and LNK-033's landmine knows
`site-header` carries a `/contact.html` schema fallback — but neither says the header shares
the resolver's *candidate loaders*, which is the fact that widens the blast radius.

**Why it changes the design, decisively.** LNK-033's landmine lists three ways to break the
invariant: widen `loadContentHubs`/`loadInteractivePages`, drop `candidatesFromHubs`' filter,
or drop `rank()`'s excluded-area test. Those are not equivalent. Widening at the **loaders**
also silently re-picks **every site's header button**, because the header derivation reads the
same two functions and takes `ordered[0]` by nav_order — a page newly admitted to the loader
result can outrank today's pick without going anywhere near a utility area. So:

> **The widening must happen at the candidate-assembly seam (`candidatesFromHubs`), never at
> the loaders.** The loaders are shared with chrome; the assembly seam is not.

**And the change would be invisible to the obvious instrument.** [MEASURED 2026-08-22]

```sql
SELECT slot_name, count(*) rows, count(*) FILTER (WHERE content_data ? 'cta_url') has_cta_url
FROM site_components GROUP BY 1;
--  footer 24 / 0 · header 24 / 0 · head 24 / 0
```

The header CTA is **not persisted anywhere** — `site_components.content_data` carries no
`cta_url` on any of the 24 header rows. It is derived at chrome-render time and written
straight into `rendered_html`. So a before/after diff of `content_data` (the natural check,
and the one the 355/RFC_042 content-loss work is building) would read **clean** across all 24
sites while every header button silently moved. The only instrument that could see it is a
diff of rendered chrome HTML.

[UNVERIFIED] I have not confirmed that a widened loader result would in fact outrank the
current pick on any specific site — that is a claim about nav_order values I did not query.
The design constraint does not depend on it (the point is that the instrument is blind, not
that the change is certain), but if the plan ever proposes touching the loaders, that query
is owed first.

### Note on where LNK-033's scope is genuinely sound

I went looking for a live falsification and did not find one, so recording the negative: the
header CTA legitimately resolves to `/contact.html` from the nav's own contact item
(`render_site_components_action.go:162-164`), which looks at first like a resolver path that
CAN mint a utility url — falsifying LNK-033. It does not, because chrome lives in
**`site_components`**, a different table from `page_components`, and `site-header` is not in
`ctaFieldNames`, so `storedCTADestinationIsAuthored` is never consulted for it. The predicate's
scope is exact. Recorded because the near-miss is the useful part: the invariant is stated in
prose that sounds table-agnostic and is only true because of a scoping nobody wrote down.

## 2026-08-22 (later still) — where a stamp WOULD be dropped, read rather than reasoned

The design direction I want the plan to evaluate is to invert the burden: rather than
recording "a human authored this", have the **resolver stamp what IT wrote**, so
`authored == valid && not-stamped`. One producer cooperates; every other writer is correct by
default (a REPLACE drops value and stamp together; a section-editor field update that writes a
new url without a stamp is genuinely authored).

The dangerous case is a writer that carries the **value** forward without the **stamp**. I
went to check whether one exists rather than assuming, and it does — it is PBP-039's carry:

`plan_sections_action.go:2110-2126`, `carryStored()`:

```go
carryStored := func() bool {
    if source == "" || source == "llm" { return false }
    value, ok := resolver.storedFieldValue(ctx, sectionName, fieldName)
    if !ok { return false }
    resolvedData[fieldName] = value
    carriedFields = append(carriedFields, fieldName)
    ...
}
```

It is **per-field by construction**: it looks up `fieldName` in the stored row and writes that
one key into `resolvedData`. A sibling provenance key (`cta_url__origin`, or an entry in a
`__origin` map) is a *different* key, so it is not looked up and not carried.

Two things follow, and both are load-bearing for the plan:

1. **A naive sibling-key stamp fails in the worst direction.** A resolver-minted url that goes
   through the carry arrives on the far side **unstamped**, is therefore read as authored, and
   is frozen for ever by the very keep-branch this bug is trying to make honest. That is the
   current bug's failure mode reintroduced through the fix.
2. **CTA url fields do reach this code path.** `carryStored` declines only `source == ""` and
   `source == "llm"`; CTA url fields are `source: "renderer"` (`ctaFieldNames`' own comment
   requires it), which resolves to nothing at plan time and so falls to `handleMissingField()`
   → `carryStored()`. `bugs_open/312` says the same thing from the other end: the values that
   survived a fresh build were "the carry re-shipping the stored row".

So the carry must carry the stamp **with** the value, and that is a specific, checkable,
one-place code change rather than a property to be hoped for. It is also a mutation target:
no-op the stamp-carry and the test that must fail is one asserting a carried resolver value is
still recomputable — i.e. that it did NOT become authored.

> **[UNVERIFIED] discharged, same session.** I had marked "whether `storedFieldValue` is the
> only per-field read in that path" as owed. It is: `grep -rn storedFieldValue --include=*.go
> platform/ | grep -v _test` returns **one definition (`:255`), one call site (`:2114`), and
> two comment mentions** — no second reader. So the carry is a single point, and stamping it is
> a one-place change. Recording the discharge rather than deleting the marker, because the
> marker is the evidence that the check happened rather than being assumed.

## 2026-08-22 — CORRECTION to my own entry above: right mechanism, WRONG REMEDY LOCATION

> **CORRECTED 2026-08-22, same session, caught by the fable planning agent's disagreement and
> then by reading the resolver's call site.** The entry above concluded: *"the carry must carry
> the stamp with the value, and that is a specific, checkable, one-place code change"* — naming
> `plan_sections_action.go`'s `carryStored` as the place to fix. **The mechanism I described is
> right and the remedy is wrong**, and a reader acting on it would have edited a function that
> cannot fix this.

**What I missed: `setCTAField` does not read the carry's output for its decision — it reads the
DB.** At the call site (`resolve_internal_links_action.go:207-212`):

```go
resolved := sectionResolvedData(section)          // <- what the carry populated
existing := existingLabels[sectionName]           // <- a FRESH DB read
setCTAField(resolved, existing, fields[0], ...)   // `stored` = existing, not resolved
```

`existingLabels` comes from `loadExistingSectionContentData`, a fresh `page_components` read. So
the stamp — living in stored `content_data` — is in `setCTAField`'s hand **directly**, whatever
the carry did or did not do. The predicate's input never travels through `plan_sections`.

**Where the leak actually is, located exactly.** `setCTAField`'s final branch (the tail of the
function, ~`:455-462`) writes **nothing** to `resolved[field]`:

```go
	*unresolved = append(*unresolved, map[string]interface{}{ ... })
}
```

Every other branch assigns `resolved[field]`. So on the unresolved fallthrough — no valid
positional target — whatever `plan_sections`' carry left in `resolved[field]` is what gets
persisted, **and the save is a REPLACE** (RFC_042 §2: the plan→save funnel is DELETE+INSERT), so
the previous cycle's stamp is not carried forward by the persist either. Next cycle reads an
unstamped valid value → authored → frozen.

**So the fix belongs in `setCTAField`, not in `carryStored`** — and it is cheaper there, because
`stored` is already a parameter: at the unresolved path, if `CTAMintedCovers(stored, field,
resolved[field])` then re-stamp the surviving value. One function, one branch, no change to
`plan_sections` at all, and no risk to PBP-039's carry (a seam whose register entry says in
terms: *do not remove it, do not reorder it*).

**What I got right and am keeping:** the failure direction (value without stamp → false
authored → freeze), that CTA url fields reach the carry (`source: "renderer"`, declined only for
`""`/`"llm"`), and that `storedFieldValue` has exactly one call site.

**What caught it.** The planning agent asserted the opposite ("the carry structurally cannot
carry the stamp, and under value-binding it must not"). Its *reason* is also wrong — the whole
stored map is in hand at `storedFieldValue:270` (`data[field]`), so a stamp copy is two lines and
"structurally cannot" is false — but the disagreement was the prompt to go and read the call
site, which is where both of us were wrong about something. **Neither account survived contact
with the call site; the disagreement is what sent me to it.**

**The cheap check I skipped.** I reasoned about `carryStored` in isolation and never asked the
one question that decides it: *where does the consumer of this value get its input from?* One
`sed -n '207,212p'` at the call site answers it. General form: **a claim that data flows from A
to B is not checkable at A — it is checkable at B's call site**, and I had already read that call
site earlier in the session for a different purpose without connecting it.
