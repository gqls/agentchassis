# 420 — the negation gate's prose walker skips `name` fields, so heading-surface tics ship unscanned

> ## STATUS 2026-09-04 — FIXED IN CODE, INERT UNTIL THE NEXT ROLL. STAYS OPEN.
>
> Commit **`60091e140`** (7 files, `platform/orchestration/{datahelpers,actions}`), council
> correlation **`3e9e8ce8-fb9b-4f5b-a610-016b57427a27`**, verdict pending at time of writing.
> Bar for `bugs_closed/` is fixed AND live, so this stays here until a roll ships it and the
> post-roll check below passes.
>
> **⚠ THE FIX IN THIS FILE'S "Fix candidates" SECTION WAS UNSAFE. Do not apply candidate 1 as
> written** — the section is left intact below because it is the record of what we believed,
> and the two reasons it is wrong are the whole value of this bug. See §RESOLUTION.
>
> **What a reader most needs to know:** `name` is an IDENTITY where the item has a `url`
> sibling and DISPLAY copy where it does not — 908 vs 825 items fleet-wide with zero
> crossover `[MEASURED 2026-09-03]`. Neither "skip `*.name`" nor "scan `*.name`" is right.

**Filed:** 2026-08-31, `copy_quality_two_stage` lane, from the v1.0.1349 post-roll canary
(finetuning.uk approach page, rebuild item `4641e02d`, orchestration `2eecc01d-…`).
**Resumed 2026-09-04** by the `420 425` session with the copy lane's explicit handover
("Take it — I'm not on it and won't be"), after the file had sat untouched for three days.

**Diagnosis-loop note (owner ruling 2026-07-31):** filed on first-hand verification instead of a
090 run, and here is the substitution — every link in the chain is measured on one artefact, in
one run, with the walker's own instrumentation as witness: the gate's marker shows sibling
`description` fields of the SAME array targeted and repaired while the `name` fields never
entered `hits_before`; the exclusion regex names `name$` on a single greppable line; and the
served page carries both skipped strings verbatim. No inference bridges the steps. A 090 run
would re-read the same three artefacts.

## Symptom

After a full rebuild through the corrected stack (v1.0.1349: seven gate shapes, empty mild list,
ruling-7 repair), the served page still carries two define-by-negation constructions **on
heading surfaces**:

- "Privacy built into the deployment, not added afterwards"
- "Facts checked against sources, not asserted"

Both are `features[N].name` values (iter_2's stored result: `features[1].name`,
`features[6].name`), rendered as card headings.

## Evidence (all 2026-08-31, canary run 15:24–15:28Z)

- `process_sections_loop_iter_2_rewrite_negations` marker (orchestration `2eecc01d`,
  collected_data): `hits_before: 3`, `exempt: 0`, targets = `features[1].description`,
  `features[5].description`, `features[6].description`. **The two comma-not `name` strings in
  the same `features` array are absent from every list** — not exempted (exempt would be ≥1 and
  `exempt_reasons` populated), not rejected, never counted. The scanner never saw them.
- The same run proves the machinery works where the walker looks: across the page, 12
  replacements proposed, 10 spliced and serving, 2 refused with recorded reasons
  (`still_negative_reveal`, `gutted`) — so this is not a wiring failure.
- Battery on the served page (`count_negation_tells.py`): 3 × "X, not Y", of which 2 are these
  headings.

## Root cause (read, not inferred)

`platform/orchestration/datahelpers/negation_content.go:46`:

```go
var nonProseFieldRe = regexp.MustCompile(`(?i)(^|_)(url|…|target|rel|name)$`)
```

`name` is on the never-holds-prose list, so `WalkContentStrings` → `prosey()` returns false for
every `*.name` field and neither consumer (the renderer's annotation, the repair action) ever
scans one. The file's own comment names the cost — "A false NEGATIVE here costs a missed tell"
— and `IsHeadlineField`'s comment calls headline surfaces the place the construction is "least
forgivable". A feature card's `name` IS its heading: the field the exclusion list treats as an
identifier is the page's highlight surface. (The Gemini G1 trial read flagged exactly this
surface — "tic ON the highlight surface" — one model trial earlier, on the same page family.)

Every other entry in the list (`url`, `hex`, `price`, `email`…) is genuinely non-prose;
`name` is the only member that collides with a heading use. `title`/`heading`/`headline` are
scanned (and get headline severity).

## Fix candidates, ordered by what closes the door

1. **Drop `name` from `nonProseFieldRe` and let `prosey()`'s value tests decide** (≥12 chars,
   contains whitespace, fails `nonProseValueRe`). A prose-bearing name is then structurally
   scannable — the bad state becomes unrepresentable. Cost: person/product names like
   "Farm Shield" mostly fail the length/space tests anyway; one that passes gets scanned, which
   is harmless — scanning only flags negation shapes, and every splice still passes the claim
   and structure guards. One-line change + walker test.
2. Scan `name` but only when `IsHeadlineField`-style sibling heuristics say the object is a
   card/feature (has `description`/`icon` siblings). Closes less: a new classifier whose gaps
   the guarantee inherits (see memory: a guarantee conditional on a classifier).
3. Prompt-side ban ("never negate in a heading"). Closes nothing — the gate exists because
   prompt rules leak.

Candidate 1 should also add `name` to the headline-severity question: a card name is a heading,
so a hit there should carry `headline: true` weighting (extend `headlineFieldRe` or special-case
the walker's path tail).

## How to verify a fix

Re-run the canary shape: rebuild any page whose section uses `features[].name` +
`description`, then assert the marker's `hits_before` counts name-field hits (a unit test on
`WalkContentStrings` with `{"features":[{"name":"X, not Y case","description":"…"}]}` is the
cheap half — it must return the name field). Control: the same test against the current binary
returns no name path (proves the test can fail).

## Scope note

> **⚠ CORRECTED 2026-09-03 by the `420 425` lane, verified independently before accepting — THE
> CLAIM BELOW IS FALSE AND IT WOULD HAVE WASTED A MEASUREMENT.** `cmd/brief-negation-check` does
> **NOT** import the walker. `grep -rn "WalkContentStrings\|nonProseFieldRe\|prosey"
> cmd/brief-negation-check/` returns **nothing**. It has its own traversal (`assessValue`,
> `flatten`) and reaches `datahelpers` only for text-taking scanners — `ScanDefineByNegation`,
> `NegationExempt`, `ScanBannedRegisterWords`, `ScanPracticeClaims`, `SplitPlainAssertionText`,
> plus the evidence/practice helpers. **None of them takes a field name, so none consults
> `nonProseFieldRe`.** It scans SPEC text; the walker scans `content_data`. Different populations,
> no shared filter.
>
> **So a before/after on the nightly is a CONTROL, not a measurement:** the prediction is NO
> movement attributable to a walker change, and movement on 09-04 against the 09-03 baseline means
> something ELSE moved and both lanes should want to know. Capture it as **N-of-M** — baseline
> 2026-09-03 07:41Z = **11 of 39** brief-supplies — because the denominator grows as sites are
> added and a bare count either side is uninterpretable.
>
> How the false claim arose: the two components sit next to each other, share the same scanners,
> and both talk about negations — so "shares the walker" was an inference from adjacency that one
> grep refutes. It went into a Scope note, which is the half of a bug file a later reader trusts
> without re-checking.

~~`cmd/brief-negation-check` imports the same walker for spec scanning — a fix here widens what
that check sees on its next rebuild; expect its suppression/finding counts to move and do not
read that movement as drift (its standing daily read is documented in the copy lane's handoff).~~

---

## RESOLUTION 2026-09-04 — and why the fix candidate above was unsafe

### Why candidate 1 ("drop `name` from `nonProseFieldRe`") must not be applied as written

**Reason 1 — that regex was the ONLY thing keeping identity fields away from the LLM repair.**
It is load-bearing on the rewrite side, not a scan optimisation, and the file's own header
already stated both costs in one comment without anyone noticing they are different questions:
"a false NEGATIVE here costs a missed tell; a false POSITIVE sends a URL to a model and asks it
to rewrite the sentence". Scanning and overwriting are different questions with opposite risk
directions, and `name` is exactly where the two answers differ. Confirmed at the code by the
`copy_quality_two_stage` lane: `AcceptNegationRewrite` had `invented_name` but **no
`dropped_name`** — nothing downstream stopped a rewrite LOSING an identity string.

**Reason 2 — `name` carries two OPPOSITE contracts by producer.** Found by the components lane
(`bugs_open/425`) and then censused fleet-wide. `[MEASURED 2026-09-03]`, over all **1,729**
objects at any depth in `page_components.content_data` carrying a `name` key — reproduced
independently by that lane from a query sharing no code:

| `url` sibling | `name` prose-shaped | n | contract |
|---|---|---|---|
| no `url` key at all | **752 of 825** | 825 | **DISPLAY** — directory / feature / tracker items (`business_directory.go:185`, `directory_items.go:224`). "Freedom Health Insurance", "170+ Agents Running in Production" |
| `url` key, non-empty | **0 of 908** | 908 | **IDENTITY** — listing / nav items (`queryresolve.go` ~747). `name` is the real `pages.name` and the item's own `url` is built from it, so a rewrite desynchronises the item from BOTH in one stroke — **and the page still renders**, which is the expensive kind |
| `url` key, empty or null | — | **0** | the state does not occur; checked explicitly because a clean split is a reason to doubt the instrument |

**Zero crossover.** Two lexical signals partition the same items identically (908/908 lowercase
slugs, 825/825 containing uppercase) and are deliberately **not** used: case is a property of
who the current producers are, the `url` sibling is a property of the shape.

⚠ **Listing `name`s are hyphenated slugs, so the value tests skip them today BY LUCK, NOT BY
PROTECTION** — one tokeniser change from silent and estate-wide. That is why the fix does not
lean on it.

### What shipped (commit `60091e140`)

1. **The predicate split**, following the estate's own doctrine rather than inventing a seam —
   `markup_spans.go:63-74` ("a writer exclusion is deliberately NOT a detector exclusion") and
   `resolve_internal_links_action.go:81-86` from `bugs_open/248`'s clobber ("'never newly SEND'
   … 'never TRUST an existing' is a different and much stronger claim that happens to reuse the
   same set"). `runtime_fill.go:29-38` is the mechanical form copied.
   `isProseContentField` (SCAN, fails toward scanning) and `identityContentField` (OVERWRITE
   guard, fails toward exclusion). Bare `name` leaves the never-prose list; **the `_name` SUFFIX
   arm STAYS** and is what still protects `company_name` (84), `current_page_name` (71),
   `cardN_client_name`, `*_author_name`, `tool_name` — zero of which carry a gate shape.
2. **One walk, one count preserved.** The identity flag rides on the yielded field; the repair
   records it as an exemption, so `total = exempt + withinBudget + targets` still reconciles with
   the annotation. A filter would have made the populations differ — the thing the walker's
   header forbids.
3. **`dropped_name` in `AcceptNegationRewrite`** — the missing LOSE half, keyed on `protectFrom`
   as `dropped_figure` is. **In the judge, not the walker, deliberately:** a filter is bypassable
   by any future caller that enumerates fields itself; a rejection reason is not. Both mutating
   call sites inherit it (`rewrite_negations_action.go`, `judgeRegisterRewrite`).
4. **`name` joins `headlineFieldRe`**, and the identical dual-purpose defect in `IsHeadlineField`
   is closed **by ORDER**: the identity exemption runs before the headline branch, so an identity
   field can never be *forced* to the model by headline severity. A test pins the ordering.
5. **OWNER RULING 2026-09-03 (in session): a 2-word HEADING floor**, separate from the untouched
   5-word sentence floor. `[MEASURED 2026-09-03]` truncating the 36 live `x_not_y` headings at the
   construction leaves 1 word ×1, 2 ×10, 3 ×6, 4 ×8, 5 ×6, 6 ×3, 7 ×2 — so the sentence floor
   would have refused **25 of 36** as `gutted`: visible, then unrepaired.

### The live population this fixes, dated so it can be re-measured

`[MEASURED 2026-09-03]` **37** values, **15** domains, **23** pages, all `status='deployed'`; all
in LLM-written slots (`differentiators` 22, `features` 11, `differentiators-section` 4).
gaswholesalers.com 6 · leopardessconsulting.co.uk 5 · robot-hands.com 4 · gamesdesign.co.uk 4 ·
ai-agent-orchestration.com 3 · finetuning.uk 3 · garden-tools.uk 2 · lendzy.co.uk 2 · vonc.com 2 ·
dartsonline.com, remortgagecalculator.uk, cookly.uk, seotools.co.uk, farmerinsurance.uk,
vetcomparison.uk 1 each.

**OWNER DECISION 2026-09-04: these are left to heal on rebuild — no sweep.** ⚠ **A
`page_rerender` would repair NONE of them**, and I offered that as an option before checking,
which was wrong (recorded in `WRONG_CALLS.md`). `page-rerender` has neither `rewrite_negations`
nor `copy_gate_annotate`, and `rerender_page_sections_action.go:3-9` re-renders stored
`content_data` "WITHOUT invoking the content writer (no LLM)". The tell lives in
`content_data.name`, so a rerender faithfully re-renders it. Only a `page-content-writer`
rebuild regenerates these values. One query settles it:

```sql
SELECT default_config::text ~ 'rewrite_negations' FROM agent_definitions
WHERE type='page-rerender' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

### How to verify after the roll

1. **At the binary, not the tag** — capability literal, with BOTH controls:
   `identityContentField` must be PRESENT, and a nonsense symbol ABSENT, in the same breath.
2. Re-run the 37-value census above; it should fall as pages rebuild, not all at once.
3. In the gate marker, expect `hits_before` to rise on card sections and
   `exempt_reasons.identity_name_with_url` to appear. **`dropped_name` in the rejection log is
   the instrument working, not a fault.**
4. ⚠ **Do not read a low rewrite count as "the fix did nothing" — read the rejection reasons.**
   The copy lane's `gutted` rework (`7cc16a5d0`) makes short repairs more likely to be refused,
   and the heading floor is what stops that swallowing this fix.
5. **Control that must NOT move:** the nightly `brief-negation-check` (`40 7 * * *`), baseline
   **11 of 39** at 2026-09-03 07:41Z — recorded N-of-M because the denominator grows. The struck
   scope note above is why: that check does not share this walker, so movement means something
   else changed.

### Still open, recorded rather than closed by assumption

- **`sectionAssetKeyLike`** (`section_text.go:45`) is a FOURTH member of this dual-purpose class
  and the highest-consequence one: shared between a read-only duplication detector and
  `remove_duplicate_page_sections_action.go:297`, which executes a `DELETE`. Not touched here.
- `ScanContentDataForNegation` never calls `NegationExempt` while the repair does, so annotation
  and repair already disagree in the other direction.
- Mirrored listing items carry 58 `meta_description` + 4 `excerpt` gate hits the annotation
  reports and the repair can never see — fixable only at the source page.
- `bugs_open/457`'s NULL-`component_id` rows re-ship stored HTML verbatim; a copy rewrite through
  the normal path will not reach them.
