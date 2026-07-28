# 109 — the render-context allowlists drop any field not on them, silently; three are being dropped right now

**Filed:** 2026-07-27, by the brochure_component_library workstream, at the council gate's
request. `bugs_open/085`'s reviewers (bug_historian, severity **high**, gating) objected
that fixing `current_page` at its three drop points leaves the *mechanism* that dropped it
untouched: "any future render_context field that isn't added to the allowlist will be
silently dropped the same way, with the template still advertising it as present."

The objection was accepted. This file is the tracked follow-up it asked for.

**Severity:** Low today, structurally Medium — nothing is visibly broken, and that is the
complaint. It is a defect factory rather than a defect.
**Class:** structural (a contract advertised in one place and honoured by an allowlist in
another, with no relationship between them that anything checks).
**Status:** **ALL FOUR MAPS NOW DERIVED — candidate 1 complete in code (`f78cf8125`,
2026-07-28, council corr `1d082754`, verdict pending at commit time), INERT until a
chassis image roll.** See the 2026-07-28 box below; the earlier box records the first map.

> ## STATUS 2026-07-28 — the remaining three maps derived (`f78cf8125`)
>
> `setRenderContextScalarsFromData` is the write-side twin of `renderContextScalarFields`:
> build (`mergeIntoRenderContextEnhanced`) and restore (`mergeIntoRenderContext`) now
> accept exactly the step contract the serialiser emits — one predicate
> (`renderContextStepContractExcluded`), three call sites. Both render maps
> (`contextToInterfaceMap`, `contextToMap`) derive the template contract from the same
> projection plus explicit decorations. `schema_mode` moved to a new
> `renderContextControlFields` (machinery: never advertised, never data-settable — tested).
>
> **Two deliberate divergence closures** (the rest is transcription-pinned
> behaviour-preserving):
> 1. **Restore now sets twelve more struct fields**, so the regex-fallback renderer stops
>    painting hard-coded default colours and empty cta/industry/year for restored
>    contexts — `current_page`'s exact latent gap (085), generalised.
> 2. **`contextToMap` gains `logo_url`** — the two render literals had themselves drifted.
>
> **Still excluded, deliberately:** `theme_css`/`title`/`description` (per-page producers
> undecided — unchanged from the earlier ruling, now pinned by
> `TestBuildStillExcludesPerPageFields`).
>
> **Why the case is not closed:** the fix is inert until a chassis image roll ≥
> `f78cf8125`, and the post-roll check is owed: render a page on the fallback path (or
> unit-drive it in the pod image) and confirm restored colours arrive instead of
> `#1a1a2e`-family defaults. After that, closing is a judgement call on whether the
> remaining residue (the per-page trio's producers) stays here or moves to its own case.

> ## STATUS 2026-07-27 — the SERIALISE map is derived and **LIVE on v1.0.1177**; the other three are not
>
> **Rolled 19:22:02Z.** Verified in the running pod:
> `grep -c "genuinely dropped, tracked in bugs_open/109"` → **3**.
>
> **Done (`595c1f499`, council `d8517d30-c691-4e4d-9647-b17a51324cd3`, APPROVED, 10
> reviewers, 0 unreadable, 5 advisory objections, no veto).** `renderCtxToMap`'s scalar
> half is now **derived from `RenderContext`'s json tags** rather than hand-listed, so
> the default for a new field is *serialised* and an omission must be written into
> `renderContextUnserialised` **with a reason**. That is candidate 1, applied to one of
> the four maps.
>
> The refactor is **behaviour-preserving today** — the serialised key set is identical,
> and `TestRenderCtxToMapDerivationIsBehaviourPreserving` transcribes the old literal
> and asserts both directions, so a future change that adds or drops a key is reported
> rather than discovered. Three fields also gained json tags they never had
> (`Industry`, `Tone`, `TargetAudience`, `Services`); without them the struct was not a
> complete declaration of anything.
>
> The test's **duplicate copy** of the omission list is gone — it now derives from the
> same map. Two lists that must agree with nothing checking that they do is the same
> drift class as the defect itself.
>
> ### Still open — this is why the case is not closed
>
> Candidate 1 asked for **all four** maps. Three remain hand-maintained and carry the
> identical defect: `mergeIntoRenderContextEnhanced` (build), `mergeIntoRenderContext`
> (restore), `contextToInterfaceMap` (render). The council's `editquality` and
> `guardian` seats both said so explicitly, and `guardian` said in terms that 109
> should stay open rather than be closed on this step. The commit for this fix drew the
> `untouched-twin` advisory for exactly that reason, which is expected and recorded.
>
> **`theme_css` / `title` / `description` are still dropped**, deliberately. Reading
> their writers is what settled it: `Title` and `Description` are written **per page**
> (`rerender_pages_actions.go:191-192`, `multipage_actions.go:94-99`), so serialising
> them from a *site-level* context would bleed one page's title onto every page — the
> opposite of the per-page behaviour `current_page` needed. Each needs its producer
> decided; that is a behaviour change, not a mechanism fix.
>
> ### Two facts checked for the council, worth keeping
>
> - **`RenderContext` is marshalled nowhere**, and is embedded in no other struct. The
>   json tags were inert documentation before this change, which is what makes
>   promoting them to the authoritative declaration free.
> - **`renderCtxToMap` has exactly one production caller** — `BuildRenderContextAction`
>   (`v3_site_actions.go:1026`). Every other reference is a test.
>
> ### New failure mode this introduces, and its guard
>
> Deriving from tags means a **duplicate** tag silently collides (map insertion
> overwrites) and a **typo'd** tag silently becomes a key nothing reads.
> `TestRenderContextJSONTagsAreUnique` catches the first. The second is caught from the
> other side by the existing contract test. This is a narrower silent-failure surface
> than the one removed, but it is not zero and should not be described as such.

---

**The commit hook agrees, independently.** `085`'s fix commit drew a
`pattern-check.py` **untouched-twin** advisory — *"changed `mergeIntoRenderContext()`
but not its twin `mergeIntoRenderContextEnhanced()` … if the change is a fix, the twin
probably has the same defect"* (016b §9 #26). It does. Leaving the twin alone was
deliberate and is this file; recording the hit here so the two accounts do not drift.

## The mechanism

A `RenderContext` reaches a component template by four different maps, and two of them are
hand-maintained allowlists:

| step | function | shape |
|---|---|---|
| build | `mergeIntoRenderContextEnhanced` (`v3_site_actions.go`) | **allowlist** — ~20 named fields extracted from each source; everything else dropped |
| serialise | `renderCtxToMap` (`v3_site_actions.go:1316`) | **allowlist** — ~20 named keys written into `collected_data` |
| restore | `mergeIntoRenderContext` (`:1428`) | allowlist of ~9, then a catch-all into `ContentData` |
| render | `contextToInterfaceMap` / `contextToMap` (`component_library.go`) | the contract the template sees |

Nothing relates the four. Add a field to `RenderContext`, wire it into the template map,
and it is advertised to every component author while arriving empty forever. That is
exactly what happened to `current_page` — and the failure has no error surface at all,
because "empty" is a legal value for every one of these fields.

## Three fields are in that state today

Measured 2026-07-27 by diffing the two key sets with every scalar populated (so nothing is
missing merely for being a zero value):

```
advertised to templates but never written by renderCtxToMap:
  contact_email, description, logo_url, theme_css, title
```

- **`contact_email`** — benign. A render-time alias of `email`, which is serialised.
- **`logo_url`** — latent, not live: both writers set `ContentData["logo_url"]` alongside
  the struct field, and `renderCtxToMap` merges `ContentData` at the end. The struct field
  alone would not survive.
- **`theme_css`, `title`, `description`** — **genuinely dropped**, same shape as
  `current_page`. A component template reading `{{.title}}` or `{{.theme_css}}` on the
  page-build path gets empty, always. Whether that matters depends on whether anything
  reads them there, which is the first thing to check.

## What is already in place

`085`'s fix shipped a contract test —
`TestRenderContextSerialisationCoversTemplateContract` in
`platform/orchestration/actions/render_context_current_page_test.go`. It fails when a field
is advertised to templates and not serialised, unless it is named in
`knownUnserialisedContextFields` with a reason. Verified by injecting a new field into
`contextToInterfaceMap` and watching it fail.

**That stops the gap growing. It does not close the gap**, and it covers only the
serialise↔render pair — the build-side allowlist in `mergeIntoRenderContextEnhanced` has no
equivalent, because there is no declared contract on that side to compare against.

## Fix candidates, ordered by what closes the door

1. **Make the bad state unrepresentable: derive the maps from one declaration.** Tag the
   `RenderContext` struct fields with their template key and generate/reflect all four maps
   from that. Then "add a field and forget an allowlist" cannot be expressed. Largest
   change, and the only one that ends the class.
2. **Close the three live gaps** (`theme_css`, `title`, `description`) by serialising them,
   after checking what reads them — cheap, and it shrinks
   `knownUnserialisedContextFields` toward the empty list the test wants.
3. **Extend the contract test to the build side** by asserting the set of keys
   `mergeIntoRenderContextEnhanced` consumes against the set of source fields the live step
   configs actually supply. Catches the drop that started `085` — but it needs a
   declaration of what a source is allowed to carry, which does not exist yet.
4. Documentation on the allowlists. Listed for completeness and last on purpose: "operators
   must remember to update four places" is a schema defect in a documentation costume.

## How to verify a fix

`knownUnserialisedContextFields` shrinking, with the test still passing, is the direct
measure. Beyond that: pick a field, set it in `BuildRenderContextAction`, and assert it
arrives in a rendered component — end to end, not per function. `085` is the worked example
of why per-function assertions pass while the value never arrives.

## What this is NOT

- Not `bugs_open/085` — that is the one field, fixed. This is the mechanism that dropped it,
  which `085`'s fix deliberately did not touch.
- Not a claims or content defect. No wrong value is ever displayed; a field is absent.
