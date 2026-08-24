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

`loadContentHubs` and `loadInteractivePages` have **three** callers **as of 2026-08-22**, not two:

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

## 2026-08-22 (evening) — round 4 was never reviewed: the estate's LLM budget is exhausted until 2026-09-01

Round 4 dispatched at 18:10:08Z and completed at 18:15:39Z on the terminal step
**`complete_invalid`**, with **no `council_report` row written**. My first reading was that my own
change had broken the payload — I had just replaced hand-written sketches with generated diff
hunks, so "the gate refused my submission as invalid" was the obvious and entirely wrong
conclusion. It is worth naming that I reached for it before checking, because it is the trap:

```
__step_error: step review_debug_historian failed: failed to execute action execute_llm_prompt:
AI endpoint unavailable: provider=anthropic model=claude-sonnet-5 ...
API request failed with status 400: {"type":"error","error":{"type":"invalid_request_error",
"message":"You have reached your specified API usage limits.
You will regain access on 2026-09-01 at 00:00 UTC."}}
```

**HTTP 400 and `invalid_request_error`, not 429** — the error's own type name reads as "your
request was malformed", which is what makes the wrong conclusion so easy. And `DRY_RUN=1` had
passed on the same file seconds earlier, because it validates locally and never calls a model:
**a green dry run followed by `complete_invalid` is the signature of this trap, not evidence
against it.** Filed as a LANDMINE.

**Not confined to this lane** [MEASURED 2026-08-22, 18:15:39Z–18:26:25Z]:

```sql
SELECT COALESCE(collected_data->'__step_error'->>'failed_step','?'), count(*)
FROM orchestration_states
WHERE collected_data->>'__step_error' ILIKE '%API usage limits%' GROUP BY 1 ORDER BY 2 DESC;
```

| failed_step | n |
|---|---|
| `call_content_writer` | 3 |
| `review_architecture` | 2 |
| `review_debug_historian` | 1 |
| `review_editquality` | 1 |

**7 failed steps across 5 orchestrations, and `call_content_writer` is live site content
generation — not a council seat.** So this is an estate-wide condition, not a property of my
submission.

**Scope stated honestly, because I nearly overstated it.** 95 orchestrations reached `COMPLETED`
in the 18:00Z hour, so it is *not* a total outage. But **zero** orchestrations have completed
carrying `__usage_output_tokens` since the last failure at 18:26:25Z — i.e. no LLM-producing work
has succeeded since. `[UNVERIFIED]` whether the block is model-specific (the error names
`claude-sonnet-5`) or account-wide across models; the window since 18:26 is only minutes and thin
on traffic, so that zero is consistent with a hard block but does not on its own prove one.

### What I did NOT do, deliberately

**I did not resubmit.** A retry cannot succeed before 2026-09-01 and each one re-runs the seats
that did answer. The correct action is to stop and tell the owner.

### Where that leaves Phase A

Unchanged and unaffected. The council is **advisory and cannot block a commit** — Phase A is
committed (`288ce3e7a`), live on both chassis replicas, and carries `Council-Submitted:` rather
than `Council-Reviewed:`, which asserts nothing and can never become a false claim. `098` resolves
the correlation at report time, so if a verdict ever lands approved the commits are credited with
no amend. **Three rounds of substantive review did happen** (rounds 1–3, 10 approvals in round 3),
and every objection they raised has been answered in the tree. What is missing is the final
verdict, not the review.

## 2026-08-23 (afternoon) — the cap was on the WRONG ACCOUNT; Phase B measured, written, committed

### 0. The blocker this lane recorded yesterday was false, and I nearly inherited it

Yesterday's NOTES, the handoff banner, the summary and the bug file all said the same thing: the
account's LLM budget is exhausted until 2026-09-01, do not resubmit, do not start Phase B. First
action of this session was to check it rather than obey it:

```sql
SELECT date_trunc('hour',created_at) hr, provider, success, count(*),
       min(left(coalesce(error_message,''),90))
FROM llm_call_log WHERE created_at > now() - interval '6 hours' GROUP BY 1,2,3 ORDER BY 1 DESC;
```

Last `usage limits` refusal **2026-08-23 10:10:40Z**; 15 successes in the 10:00Z hour, 40 in
11:00Z, 79 in 12:00Z. The two failures since are `stop_reason=max_tokens` truncations — a
different defect. The cause is in `memory/the-fleet-key-is-not-on-the-default-console-org.md`: the
console the owner lands on by default is **not the org the fleet's key belongs to**, so it read
`0% used` while the API refused calls, and `2026-09-01` was the *other* account's reset date.

**Round 4 resubmitted** on the same correlation (`RESUBMIT_CORR=e4336931-…`), and it returned a
verdict in ~35 minutes — which is itself the proof the cap is gone.

### 1. Round 4 verdict: REVISE (the fourth), and again NOT the code

`decided_by: gating objection from editquality`, 13 reviewers, 4 abstained. Its HIGH objection is
correct and is the same class as rounds 2 and 3: **edit 2 claimed `storedCTADestinationIsAuthored`
as its symbol but its sketch was call-site-only and byte-identical to edit 3's**, so the one
function whose logic actually changed was never shown in any sketch. Also flagged: edits 2 and 3
duplicate each other, and the LANDMINES/register edits the notes claim shipped are absent from the
plan. Four rounds, four submission defects, zero code defects. The handoff's advice held.

### 2. The `bug_historian` objection, ANSWERED by a code read (it asked for one)

> "…never checks whether `apply_section_edit` / `ApplySectionEditAction` can write a new `cta_url`
> into a `ctaFieldNames` slot without going through `SetCTAMinted`."

**It can, and the answer is that this is safe BY THE VALUE-BINDING, not by luck.**
`applyContentEdit` merges `field_updates` key-by-key into the existing map
(`section_editor_actions.go:1025-1026`) and the file contains **0** occurrences of
`SetCTAMinted`/`SeedCTAMinted`/`CTAMintedKey` (grep -c, 2026-08-23). Walking the four reachable
states:

| stored record | editor writes | `CTAMintedCovers` | verdict | outcome |
|---|---|---|---|---|
| none | `/contact.html` | false | authored | KEPT ✓ |
| `{cta_url: /tools/x}` | `/contact.html` | false (names another url) | authored | KEPT ✓ |
| `{cta_url: /contact.html}` | label only, url unchanged | true | minted | KEPT by Phase B's new keeps ✓ |
| `{cta_url: /contact.html}` | `/about.html` | false | authored | KEPT ✓ |

**A stale record cannot vouch for a value it does not name — that is what value-bound means**, and
it is exactly the property LNK-035's design note claims for the section-editor merge. The objection
is right that it was asserted rather than established; it is established now.

### 3. The calibration — and it changed the design

`CALIBRATION_2026-08-23_phase_b_widening_report.md` has the full numbers. Method mirrors the
2026-08-11 precedent: throwaway harness, **real `datahelpers` imported**, frozen JSON dumps, and a
control (the harness's local ranking mirror re-scored against the shipping `BestLabelMatch` on all
1,266 pairs → **0 disagreements**; a mismatch would have invalidated everything).

Three results, in order of how much they changed what I then wrote:

1. **Widening rewrites 428 stored CTA urls fleet-wide (32 today).** ~2/3 outside 308's population.
2. **263 of 1,146 wide-pool matches are decided by alphabetical order alone; 137 of the writes.**
   Two families in there are wrong and would have been executed: finetuning.uk `"how we work"`
   (13 findings) moving OFF `/how-we-work.html`, dartsonline `"Read the guides"` (6) moving off
   `/guides/index.html`. Both are one-token ties where the loser matched on its own NAME and the
   winner on marketing copy in its TITLE.
3. **A name-tier key and a path-depth key were measured and BOTH rejected** — each repairs some and
   breaks others, e.g. name-tier moves *"Try the Password Strength Physics tool…"* off the password
   tool, because this estate names every tool page `tool-…` so the token `tool` is in all of them.
   Third rejected key in two calibrations. **So: refuse the tie instead of breaking it.**

And one counter-intuitive result worth carrying: **the NARROW widening (utility pages only) is
WORSE** — 108 writes vs 291 and *more* of them wrong, because a pool that omits the label's real
target gives the matcher a monopoly rather than a choice. It also settles 308's standing "add
`about` to LabelStopwords" suggestion: it would suppress the four `Talk to us about …` false
matches AND the correct `Learn More About Us` → `/about.html`.

### 4. What shipped (commit `7f85aa814`, inert until the roll)

`LoadCTALabelUniverse` (LNK-036) consumed by the detector and both writers; `candidatesFromHubs`
deleted; `BestLabelMatch` returns `ambiguous` (LNK-037); **both writers' keeps changed** so the
positional pick can no longer DISPLACE a utility destination it could never have chosen. That last
part is the subtle half — without it the widening re-creates `bugs_open/248`'s clobber through its
own fix, because a MINTED utility destination whose label goes generic takes no keep at all.

New invariant, replacing LNK-033's: **the positional pick may neither CHOOSE nor DISPLACE a utility
destination.** `rank()` enforces the first half; the keeps enforce the second.

Mutation-proven 7/7 against a clean `git archive HEAD` tree + my files. RFC_047 filed for the
guarantee change on `BestLabelMatch` (the commit hook's architecture signal fired on it, correctly).

### 5. My own errors this session

- **My sketch generator silently produced an EMPTY sketch for both NEW files** — `git diff` reports
  nothing at all for an untracked file, and the submission JSON looked complete (78 chars of
  header). This is round 4's own objection (a missing body) reproduced *inside the tool built to
  prevent it*, one layer down. Caught by printing sketch sizes before dispatch — 78 vs ~4,800.
  Fixed with a `--no-index` fallback **and a hard refusal**: the generator now raises rather than
  emit an empty sketch. Then a second pass found the budget had also dropped 2 of 6 hunks from
  `resolve_internal_links_action.go` — including the deletion of `candidatesFromHubs` — so the
  final run asserts **hunks-kept == hunks-in-diff for every file** and prints the ratio.
- **I reported 435 writes / 298 post-refusal, then corrected to 428 / 291.** The first pass counted
  every pick differing from the stored value; the shipping writers additionally gate on
  `validPages.Contains`. **I measured the MATCHER and called it the WRITER.** Corrected in place in
  the calibration report with the cause named.
- **I assumed the "index/home" exclusion was dropping section-index pages** (it would have explained
  "Browse all guides" landing on `/about.html`). Measured: **28 pages named index/home fleet-wide,
  all 28 at the site root, 0 non-root**. The hypothesis was wrong and the real mechanism was the
  alphabetical tie. One query, and it stopped me "fixing" something that was not broken.

### 6. Phase B council verdict: **APPROVED** (13 reviewers, 4 abstained, 4 advisory objections)

`SUBMISSION_CORR = 00732119-4e24-43c3-bd5e-ba30ced47f15`, verdict 2026-08-23 13:14:05Z,
`decided_by: approved with 4 advisory objection(s) — none high-severity`. The commit already
carries `Council-Submitted:`, and `098` resolves the correlation at report time, so it is credited
without an amend (forward-only forbids one anyway).

**Approval is not silence. What the seats found, and what I did about each:**

- **`architecture` (low) — ACTED ON.** "If a 4th caller appears it would silently inherit
  ambiguity-refusal semantics tuned for CTA repair — worth a comment at the symbol noting the
  consumer count is closed, not open." Correct and cheap: `BestLabelMatch`'s doc now lists the
  three call sites with the grep that enumerates them, and says plainly that refusing is the right
  default when the consequence is rewriting a live button and may be wrong when it is only a weaker
  suggestion.
- **`bug_historian` (medium) — CHECKED, both are CLOSED.** It asked whether `bugs_open/092` and
  `bugs_open/097` are distinct from this. They are, and neither is open: `bugs_closed/092` is the
  page WRITER never being told which pages exist; `bugs_closed/097` is in-body CARD links to unbuilt
  pages. Different surfaces — this change touches the CTA url fields of six components. Worth
  knowing that 097 is the same *family* as my build-state finding (§ the 43 planned-never-deployed
  pages), which is a point in favour of the predicate, not against it.
- **`editquality` (medium) — a REAL submission gap, not a code gap.** It observed that
  `BestLabelMatch` "almost certainly has its own dedicated unit tests… none are shown in the edit
  list", and that if so the package would not compile and "full suite green" would be false. The
  file (`label_match_test.go`) **is** updated and **is** in the commit — I simply left it out of
  the seven-edit plan. That is the same class of defect that gated Phase A four times: the
  submission understating what the change touches. **A generated edit LIST would have caught it,
  the way generated sketches caught the drift** — the edit list is still hand-written here.
- **`guardian` (medium) — NOT actionable by this lane, and it is the right objection.** "A
  wide-surface, silent (next-build-triggered) content change across the whole multi-tenant fleet
  with no staged/canary path, justified only by an owner 'no opt-out flags' ruling." This is my own
  risk #1 in the submission, said back to me by a reviewer, and it belongs to the owner: a per-site
  canary would need a flag, and the flag is what the 2026-08-18 ruling forbids. Recorded in the
  handoff and in `RFC_047`; not resolved here.
- **`prior_art_librarian` (medium)** — wanted the three-call-site enumeration shown rather than
  asserted. Fair; it is now in the code comment above, with the command that produces it.

## 2026-08-24 — PROVEN AT THE ARTEFACT, and the backlog question is answered

### 1. The build

Chassis **v1.0.1332**, deployed 09:39Z. Both replicas probed for `LoadCTALabelUniverse` (Phase B)
and `BestLabelMatchForPage` (the self-link refusal) with `LoadCTALabelUniverseNOTREAL` as a control
that must come out absent — all six readings as required, in the same exec each time.

### 2. The repair, at the served page

`finetuning.uk/ai-for-uk-small-business`, hero: **"Book a Discovery Call"** →
`/tools/password-entropy.html` became `/contact.html`, in the row, in the committed bytes
(`gqls/sites` `07f664323`) and on the live page. `__cta_minted` records **both** slots. Four
buttons on that page repaired; all four destinations 200. Detail in the bug file.

### 3. The detector, before and after, on one query

| | old runs | new run (v1.0.1332) |
|---|---|---|
| findings on finetuning.uk | 169 | 70 |
| bare `"how we work"` → `/about.html` (FALSE) | **15** | **0** |
| findings suggesting the page they sit on | **7** | **0** |

And the correct member of the same family survived: `"See how we work with clients from first call
to…"` → `/how-we-work.html`. That is the discriminating result — a change that silenced the whole
family would have taken this one too.

### 4. `detected` is NOT dispatchable — the handoff's `[UNVERIFIED]` is now resolved

`load_work_item_actions.go:711` (read at HEAD, because the file is dirty from another lane):

```sql
AND wi.status IN ('triaged', 'approved')
AND wi.attempt_count < wi.max_attempts
AND (COALESCE(wi.approval_mode,'auto') = 'auto' OR wi.status = 'approved')
```

So a freshly-`detected` item waits for triage. **It does get triaged** — fleet-wide, `cta_links_stale`
page_rerenders sit at 344 `complete`, and **zero** items have been stuck in `detected` for more than
two days. Today's 32 will drain on their own, and will now actually repair rather than completing
unchanged.

**The stuck pool is `unresolved`: 215 items, 2026-07-16 → 2026-08-18.** Those will NOT drain by
themselves, and they are what the Phase C requeue (`555_requeue_misdirected_cta_stock.sql`) exists
for. ⚠ **The requeue must set `triaged`, not `detected`** — this is exactly the fact the migration
needed and did not have.

### 5. My own error this session, and it forged evidence for the bug I was testing

My first induced rerender put `reason` at the top level of `input_data`. The workflow's
`check_rerender_mode` conditional reads **`input_data.spec.reason`**, so the run took `render_page`,
deployed the page, and completed green **with the button unchanged** — which is precisely bug 308's
symptom. I got as far as inspecting the workflow for a live defect before checking my own payload.
Filed as a LANDMINE, because when the thing under test IS a silent no-op, a malformed dispatch
manufactures a false positive that looks like the bug.

---

## 2026-08-24b — the stuck backlog re-measured against the CURRENT artefact, not against itself

Question asked: the 215 `unresolved` items, "some since 16 July" — do those sites still need
fixing? Answer: **mostly no, and the backlog is a bad map of what IS broken.**

### 1. Method — the detector's own predicate, re-run read-only over today's `rendered_html`

Not an eyeball comparison of the filed `href` against `content_data`. A standalone program
(scratchpad, not committed) that imports the REAL `datahelpers` — `CTALabelCandidateRow`,
`ExtractAnchors`, `ClassifyLinkScope`, `BestLabelMatchForPage`, `NormalizePagePath`,
`NewPageURLSet` — and copies `ctaClassifyAnchor`'s six-line body and the unknown-destination
switch verbatim. Data pulled by psql using `CTALabelUniverseSQL` and `ctaComponentScanQuery`
verbatim. Every query a SELECT; nothing dispatched.

**CONTROL, and it could have come out otherwise.** `robot-hands.com / electric-vs-pneumatic-economics`
was filed by the LIVE detector at 2026-08-24 12:49 and is still `triaged` (unrepaired), and its
page HTML last changed 08-21 — so the probe must reproduce it exactly. It did: same slot, text,
href, `suggested_target` AND `suggested_target_title`. The negative half also holds — the same
site's 16 repaired items are gone from the probe's output.

⚠ **Two data-extraction traps, both silent.**
- `kubectl exec … | psql` **truncated at 708 of 893 rows** with only an "unexpected EOF" on stderr.
  Per-site dumps with a row-count assertion against a `count(*)` taken separately; three retries.
  Without the assertion this whole measurement would have been 79% of the corpus, reported as all.
- `COPY … TO STDOUT` **text format escapes backslashes**, so a title containing `\"` arrives as
  `\\"` and a JSON newline inside `rendered_html` arrives as `\\n`. One line failed to parse and
  gave it away; the newline corruption would NOT have — it would have silently split anchors.
  Fixed by `translate(encode(convert_to(row_to_json(t)::text,'UTF8'),'base64'), E'\n','')`.

### 2. What became of the 215 [all figures MEASURED 2026-08-24]

215 rows are only **325 distinct findings** (`domain`,`page`,`slot`,`text`) — the rest are the same
key re-filed, up to 5×, which `unresolved` permits (see §4).

| | |
|---|---|
| repaired — the href now agrees with the copy | **124** |
| the CTA is no longer on the page at all (copy rewritten, slot gone) | **65** |
| unchanged button, detector now REFUSES to judge (copy names its own page) | **10** |
| unchanged button, detector now refuses (ambiguous — Phase B's tie refusal) | **54** |
| copy reworded, now names no page | **(in the 65 above)** |
| **still live damage, and auto-repairable** | **65** |
| still flagged, but only to the human-review arm | **7** |

**78% of the backlog is obsolete.** Note the two refusal rows (64 findings): those buttons are
UNCHANGED on the page. "Gone from the detector" is not "fixed" — Phase B's self-link and ambiguity
refusals make them invisible on purpose, because the platform decided a label naming its own page
is a CONTENT defect, not a destination to guess at.

### 3. The July stock is one site, and none of it is machine-fixable damage

All 15 July items are **vonc.com**, and they are 3 distinct pages / 7 distinct findings:
4 CTAs no longer exist (verified at the served page: `/about.html` now offers "Find Your
Archetype" → the clash calculator and "Enter the Gauntlet" → the gauntlet round, both sensible),
2 are unchanged buttons on `/archetypes.html` ("Explore All Archetypes" and "Explore Your
Archetypes", both → `/tools/gauntlet/index.html`) whose copy names the page they sit on — the
class the detector now declines, and 1 is ambiguous. vonc.com's live total today is **1**
misdirected CTA, in a slot the repairer cannot touch.

### 4. The backlog is not blocking anything, and releasing it is the weaker lever

- `idx_swi_dedup`'s predicate and `workItemTerminalStatuses` BOTH contain `unresolved` (checked
  in the same breath — they are the lockstep pair). So a stuck row **holds no dedup slot**: a
  fresh discovery pass files a new item for the same page. Verified in the data — vonc.com holds
  5 rows under one `item_key`.
- `suggested_target` still has **no consumer** (`grep -rn` → 4 hits, all in the two detectors).
  `rerender_page_sections_action.go` never reads `spec.findings`. So **releasing a stale item
  cannot write its stale suggestion** — it triggers a recompute against today's data with today's
  code. The risk in a release is the re-render wave, not the stale content.
- Releasing all 215 would address **65** of the **301** live findings.

### 5. What is actually broken, fleet-wide [MEASURED 2026-08-24]

**301 misdirected CTAs on 182 pages across 22 live sites** — of which **171** sit in a slot the
repairer can rewrite (`ctaFieldNames`: hero, call-to-action, archetype-grid,
archetype-combinations, gauntlet-cta, content-block-about) and **130** do not (`article-body` 37,
`ported-page` 31, `info-card-grid` 16, `tool-cta` 15, …) and escalate to human review.

Worst live case, confirmed in the served bytes, not the DB:
`gaswholesalers.com/how-it-works.html` serves **"Contact Our Sales Team" → the fuel budget
forecaster** and **"Review Supply Terms" → the break-even calculator**; `/fuel-supply-by-industry.html`
serves **"Contact our sales team" → the break-even calculator**. 25 findings on that site, all 25
machine-fixable.

### 6. The pattern that names the remedy

Last CTA sweep per site vs machine-fixable findings left today: **finetuning.uk (swept 08-24) → 0**
and **robot-hands.com (swept 08-24) → 1**; every site last swept 08-17/18/19/22 — i.e. before Phase B
went live — still carries a stack (gaswholesalers 25, leopardess 23, ai-agent-orchestration 22,
lendzy 20, dartsonline 19, …).

robot-hands.com is the clean before/after: **17 findings filed 12:49, 16 `complete` by 13:18, no
human involved**, and the probe now sees 3 — one still in the queue and two in `info-card-grid`,
which is the known non-repairable class. A completeness sweep is DB-queries-only (no LLM):
`./scripts/initial_messages/170_work_item_flow_build/075_trigger_discovery.sh <domain> completeness`
(the `misdirected_cta` check is on `completeness-discovery-agent`, verified in `agent_definitions`).

### 7. Stale figures corrected

- The handoff's "**11 client sites**" for the stuck stock is **7 as of 2026-08-24**
  (webdesign.co.uk 108, gaswholesalers.com 55, finetuning.uk 20, vonc.com 15, dartsonline.com 12,
  gamesdesign.co.uk 4, leopardessconsulting.co.uk 1).
- Migration `555_requeue_misdirected_cta_stock.sql` is aimed at a population that is 78% obsolete
  and 0% blocked. Do not build it on the "these are the broken pages" premise.

### 8. Independent audit (Fable, 2026-08-24, read-only, own probe, did not open my dumps) — corrections

Owner asked for a second model to check §§1-7 before anything was dispatched. All ten claims
held; three carry corrections, and it found the mechanism I had not named.

> **CORRECTED 2026-08-24 (audit):** §3 said "Explore Your Archetypes" was an *unchanged* button.
> Its href moved `/provocations/index.html` → `/tools/gauntlet/index.html`; it is the self-page
> refusal that hides it now, not immobility. Conclusion unchanged.
> **CORRECTED:** §6 said two sites were swept since Phase B. `orchestration_states` shows FOUR
> `completeness-discovery-agent` runs today: finetuning.uk 10:43Z, remortgagecalculator.uk 11:49Z
> (0 misdirects), robot-hands.com 12:49Z, cookly.uk 13:49Z (1 finding, repaired 14:05Z).
> **CORRECTED:** §6 said robot-hands.com had 1 machine-fixable finding left. That item is
> `triaged` at `attempt_count 2/3` with `error = OWNED_PAGE_GUARD` (page is `rebuild_policy=owned`,
> two rerenders FAILED at `save_sections`) — it will not be machine-fixed. Effective: **0**. Of the
> fleet's 171 covered findings, 1 is on an owned page → **170** writable.

**The reframing that matters — the 215 were never stuck jobs.** 212 of 215 carry the summary
`[unresolved after N attempts]`: `load_work_item_actions.go:1507-1547`, the two-strike rule,
writes a fresh finding *born* `unresolved` when its `item_key` already has ≥2 `complete`/`failed`
rows in the last 7 days. They were never dispatchable. "Release the backlog" was never releasing
held repairs; it would have re-triaged labels whose repair had already run twice and not stuck.
The other 3 are stale-reaper rows.

**And that is the mechanism that MANUFACTURES the stock, and it regrows under a sweep.** The
detector files one `page_rerender` per page whatever the slot; a page whose only findings sit in
an uncovered component (`article-body`, `ported-page`, `info-card-grid`, `tool-cta` … — **130** of
301 fleet-wide today) gets a rerender that completes as a no-op, strikes twice, and the third
sweep inside 7 days births an `unresolved` row. finetuning.uk has **32 keys at 2 strikes** in the
window right now. So: (a) space sweeps of a site more than 7 days apart, or the third one
manufactures stock; (b) **a re-sweep within 3 h of a terminal row is SUPPRESSED at Info level
(`:1526`) and reads as clean** — do not read a fast second sweep as "0 remaining"; (c) the real
fix is Phase C — the detector should not file a rerender for a page with zero covered findings,
and a completion verifier should refuse "complete and unchanged". This is `bugs_open/308` §Phase C,
not a new bug.

**Two caveats on my own claims.** C5 (a released item cannot write a stale target) holds, but a
released item is not a no-op on a page whose finding is gone: the recompute visits every
`ctaFieldNames` field on the page and can move any covered CTA whose label now matches elsewhere.
And C4 (an `unresolved` row holds no dedup slot) is true of the running code but contradicts an
owner ruling on record — `RFC_010` Decision 2 (line 173): "`unresolved` IS OPEN. Not terminal …
deduplication must be able to reach it." Unimplemented (87 duplicate rows block the index change).
**If someone builds Decision 2, a stuck row WILL hold the slot and a sweep will no longer refile.**

**500 AMBIGUOUS anchors are discarded unrecorded fleet-wide** — more than the 301 misdirects (Phase
B design choice, `check_misdirected_cta.go:110-119`; the third return is dropped on purpose). Some
are real (vonc "Enter the Gauntlet" → `/tools/gauntlet/round.html`). **"0 remaining" is not
"healthy".** gaswholesalers.com: 22 of them.

**gaswholesalers.com against every trap above [MEASURED 2026-08-24 ~15:00Z]:** 12 pages with
findings, **0** with only uncovered findings (no no-op rerenders → no stock growth); 1 key with a
single strike, 160 h old (no suppression, no birth-unresolved); 25/25 findings covered; last sweep
08-17 22:37Z. It is the clean first case as well as the worst one.

### 9. Per-site release, day 1 — gaswholesalers.com REPAIRED AT THE SERVED PAGE [MEASURED 2026-08-24]

Owner chose the worst-5 sequence (gaswholesalers → leopardess → ai-agent-orch → lendzy →
dartsonline) over releasing the 215. Pre-flight per site before firing (strikes_7d per key,
3h-suppression window, uncovered-only pages), sweep via
`scripts/initial_messages/170_work_item_flow_build/075_trigger_discovery.sh <domain> completeness`.

- **gaswholesalers.com** (corr `9917776c-3c75-4ab1-9a43-774086abe3f3`, fired ~15:06Z): discovery
  COMPLETED, 12 items filed `detected` → all `triaged` by ~15:50Z → repairs landed unattended.
  **Verified at the served bytes ~18:35Z: 8/8 of the morning's worst buttons now point where the
  copy says, and all 6 destinations 200** ("Contact Our Sales Team" → /contact.html on
  how-it-works; "Review Supply Terms" → /supply-terms-and-eligibility.html; "Contact our sales
  team" → /contact.html and "See who we serve" → /who-we-serve.html on fuel-supply-by-industry;
  all three news/pricing CTAs on fuel-industry-insights). **Detection → served fix ≈ 3.5 h,
  nobody involved.** The 25-finding DB re-probe is owed when cluster access returns.
- **dartsonline.com DEFERRED to after 2026-08-29 ~18:40Z**: 18 keys at 2 strikes (created
  08-22 18:37Z, the pre-Phase-B no-op sweep) — a sweep before they age out births `unresolved`
  rows instead of repairs (§8's mechanism, observed prospectively this time).
- ⚠ **The kubeconfig token expired MID-DRAIN (18:07 local)** and the DB-polling monitor's
  `|| echo 0` fallback turned auth failure into "no items filed" warnings — an instrument
  failure that read exactly like a dropped dispatch. The drain itself was unaffected (it runs
  in-cluster); the served page was the observability that survived. Next time: make the poll
  distinguish QUERY_ERR from zero (it did in one arm and not the other), and remember the
  3-day expiry hits observers, not the fleet.

### 10. Day 1 continued — leopardessconsulting.co.uk, and a fleet roll mid-drain [MEASURED 2026-08-24 ~19:20Z]

- **gaswholesalers.com FINAL: 12/12 complete, detector findings 25 → 0** (fresh dump, count-asserted,
  probe re-run). Both verifications now on record: served bytes (§9) and DB.
- **leopardessconsulting.co.uk** (corr `9b52142b-0e25-4a40-b108-bde1bc0805db`, filed 18:14:23Z):
  18 items, **15 complete within ~25 min**, 16/23 tick-list buttons confirmed moved at the served
  page. Residual: 2 items `OWNED_PAGE_GUARD` (llm-cost-calculator, tool-ai-vendor-trust-checklist —
  both were on the uncovered-only list anyway, so no repairable damage lost); 1 item
  (`who-we-help`) at `attempt_count 0` — its dispatch fell in the roll window below. 5 tick-list
  buttons repaired IN THE ROW (18:32-18:38Z) but the served page not yet synced — the known
  B2-lag; watcher armed.
- **Fleet rolled to v1.0.1335 mid-drain** (pods 18:31:55/18:32:19Z). Observed effects, all
  recoverable: in-flight `claimed` items fell back to `triaged` and re-dispatched; one `triaged`
  item's spawn was silently dropped (the ~300s post-restart rule, seen live) and waits for the
  loop's next pass. No item was lost.
- ⚠ **Tick-list gotcha (mine): a label truncated to 40 chars fails its own exact-match check** —
  9 buttons read as "gone from page" until re-checked with full labels. Truncate for display,
  never in the comparison key.

### 11. leopardessconsulting.co.uk FINAL — 23/23 at the served page; and §10's "B2 sync lag" was MY URL BUG

> **CORRECTED 2026-08-24 ~20:00Z:** §10 attributed 5 unverified buttons to B2 sync lag. FALSE —
> the deploy had landed. My tick-list built page URLs from the page NAME
> (`/<name>.html`); those pages live at `pages.url = /guides/<name>.html`. A wrong URL fetched
> *something* (a 404/fallback page with no matching anchors) and read as "button gone" — the
> same shape as the parked-domain lesson: **a URL census needs the page's OWN url column, never
> a URL derived from its name.** Checked at the real URLs: committed bytes AND served page
> both carry all 5 repairs.

Final: 16/18 items complete (2 = OWNED_PAGE_GUARD pages, no covered findings on them, left to
exhaust retries), **23/23 tick-list buttons verified in the served bytes**, including
who-we-help's pair after its roll-dropped dispatch was re-picked. Detection → served ≈ 25 min
for the bulk.
