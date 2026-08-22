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

## 2026-08-22 — Phase A implemented; council round 1 REVISE; two defects of my own found on the way

### Council round 1: REVISE, gated by `editquality`. It was worth the round.

`SUBMISSION_CORR=e4336931-487b-4db3-b4dc-a4b128b3566c`. 14 seats reviewed, 6 abstained; 4 objected,
8 approved. Verdict read **keyed on the correlation** against `diagnosis_artifacts`, not via the
`doc_notes ORDER BY created_at DESC LIMIT 1` recipe in CLAUDE.md — with ~40 live sessions that
returns whoever finished last (LANDMINES carries the trap; another lane hit it this same day).

Objections and what each produced. **Every one answered by a change, not an argument.**

| seat | objection | outcome |
|---|---|---|
| editquality | HIGH: the plan's prose lists the **non-page keep** among branches needing the stamp, but no diff covers it | **The prose was wrong, not the code.** The non-page keep must NOT stamp — `tel:`/`mailto:` are authored by construction and that branch's write is a URI *repair*; stamping would make a phone button recomputable (299's defect). Stated in code and prose |
| editquality | MED: the authored-utility keep's `if CTAMintedCovers(...)` guard is dead code | **Correct.** Reaching that branch proves the record does not cover the value. Removed; the code now says why in place |
| editquality | MED: no analogous fallthrough re-stamp in `applyCTARecompute` | **Answered by reading the persist**: the rerender MERGES (`mergedContent = stored ⊕ fresh`, `:725-733`), so an untouched field keeps value and record together. Only the build path's DELETE+INSERT loses them |
| bug_historian | HIGH: *"'owed' is not a control on a mechanism whose whole purpose is a stamp reaching the DB"* — read the persist path before merge | **Right, and done.** `save_page_sections_action.go:968-1001` marshals the WHOLE map, no key filter; `extractSectionsFromMetadata:1354` takes `content_data` wholesale. The 16-row measurement now corroborates a code read instead of substituting for one |
| guardian | MED: could an envelope guard's key allow-list refuse saves fleet-wide? | **Read it. No such allow-list.** `content_data_envelope_guard.go` keys on the envelope SIGNATURE (`type=="text"` AND a string `result`) — its own header says *"Key on the signature, never the arity"* |
| reuse_agent | MED: was an existing metadata sidecar checked? | No generic per-field provenance helper exists (zero hits for `FieldProvenance`/`StampField`/`ValueProvenance`/…). But `__` is an **established convention**: 15 distinct `__` markers already live in `platform/orchestration`, and `isEnvelopeMarkerKey` already treats any `__` key as platform marker rather than content |
| guidelines | MED: register entry required in the SAME commit | Done — LNK-035 added, **LNK-033 amended visibly** |
| debug_historian | MED: name the pod-verify symbol + control | Specified: grep the binary for `__cta_minted` with a must-be-absent control, per service |

One objection was **factually stale** and is recorded as such rather than silently accepted:
`bugs_open/097` does not exist — 097 is in `bugs_closed/`. Both it and `bugs_closed/023` were read;
neither shape reopens, because every stamping branch **and** the predicate itself require
`validPages.Contains`, which an unbuilt page fails.

### Two defects of my own, found by mutation rather than by reasoning

**1. The shallow merge would have dropped a sibling slot's record — the freeze, reintroduced by the
fix.** Both persist paths merge key-by-key and `__cta_minted` is a *nested map*, so a `resolved_data`
holding a record for the primary slot REPLACES the stored record and drops the secondary's. Four of
six `ctaFieldNames` components have two slots. `SeedCTAMinted` is the fix — and it lives **inside**
both writers, not in the callers' loops, because I mutation-tested the loop-level version and
**deleting the call left every test in the tree green**. That is the "a helper with no callers looks
like a finished refactor" shape, caught only because I ran the mutation instead of reasoning it.

**2. A second helper I wrote and then deleted.** `CarryCTAMinted` re-stamped at the unresolved
fallthrough. Mutation showed removing it changed no test — the seed had already carried the record,
so it was dead. Deleted rather than shipped: **two guards in series with one dead is how an
unexercised branch ships**, and the estate's own lesson is that a passing mutation usually means a
guard in series, not a guard that works.

### A claim of mine the test caught

I wrote — in code, in the submission, and in a test — that `NormalizePagePath` equates
`/contact.html` with `/contact/index.html`. **It does not.** It trims a trailing `index.html` and
trailing slashes, so `/contact/index.html` and `/contact/` both become `/contact`, while
`/contact.html` stays `/contact.html`. Those are *different pages* here; only
`ctaExcludedDestination` collapses them, and only to decide the AREA. The test now pins the
boundary in both directions, and the code comment says which forms are equated. I had asserted the
equivalence in three places before running it once.

### And one "(verified)" that was not

Correcting a mutation comment, I wrote that making `CTAMintedCovers` presence-bound would kill
`TestSetCTAFieldInventsNoProvenanceForAnAuthoredValue`, and appended **"(verified)"**. It does not —
with no record on the row at all, `Covers` returns false at its nil-map guard long before the
comparison. I ran it, it passed, and I replaced the claim with the mutation that actually kills it
(`SeedCTAMinted` inventing a record from whatever `resolved` already holds — verified by running).
**Writing "(verified)" is not a verification**, and it is worse than saying nothing, because it
tells the next reader the check has been done.
