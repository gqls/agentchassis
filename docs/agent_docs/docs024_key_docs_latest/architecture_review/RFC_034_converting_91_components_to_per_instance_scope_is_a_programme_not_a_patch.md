# RFC_034 — Converting components to per-instance scope is a PROGRAMME, not a patch, and it needs an owner decision on shape

**Status:** OPEN, filed 2026-08-17. **Blocks** `bugs_open/283`'s remaining work.

**Why this is here rather than in the bug file.** The council's `architecture` seat approved 283's
seam under RFC_022's narrow exception and named the expiry precisely:

> *"The moment the 22 templates start consuming `InstanceID`, condition 3 of the exception (zero
> live consumers) stops holding and this becomes a real load-bearing contract across the component
> library. That conversion PR, not this one, is where an RFC or at minimum a fresh architecture
> pass belongs."*

This is that RFC. The seam it licenses is live and inert (`v1.0.1305`, 0 of 244 active components
reference `{{.InstanceID}}` — the `instance-token-adoption-check` CronJob reads it daily).

---

## 1. The premise in `bugs_open/283` has drifted, and the new number changes the decision

283 says "the 22 calculator templates are unconverted". **Measured live 2026-08-17, that is not the
shape of the work.** Every figure below shares one denominator — component ROWS that bind by
`getElementById` and are placed on a live page, which is the unit that actually gets converted:

| measure | value |
|---|---|
| **component rows to convert** | **91** |
| distinct `function`s among them | 83 |
| functions carrying MORE THAN ONE active row (forks) | 4 — `tool-llm-cost-calculator` ×4, `tool-automation-savings-estimator` ×3, `tool-affordability-complaint-checker` ×3, `tool-model-approach-selector` ×2 |
| live pages affected | 94 |
| domains affected | 22 |
| literal `id=` attributes across them | 1,346 |
| `getElementById` calls across them | 886 |

**The blast radius per unit is small, and that is the good news.** Measured per component ROW, not
per function: **1** row is placed on more than one domain (max 2), and **3** rows are placed on more
than one page (max 2). Essentially, converting one row changes one page.

> ⚠ **An earlier draft of this section said "4 components are shared across up to 5 domains".**
> That was measured by `function`, and the four functions with the widest apparent spread are
> exactly the four that carry FORKS — so the grouping merged several single-domain rows into one
> apparently-shared function. Re-measured at the row level it is 1, not 4. Convert by
> `content_components.id`, never by `function`; a `function`-keyed conversion would also silently
> skip 9 forked rows.

## 2. Two findings that decide the SHAPE, both measured rather than argued

### 2.1 Converting the ids alone produces a page that reads clean and still cross-talks

Namespacing element ids is the mechanical, regex-able, obvious half. Doing only that gives a page
with **zero duplicate ids** — and every button still runs the last instance's logic.

The reason is collision class 2, which has nothing to do with ids: both instances' scripts declare
`function runCalc()` at top level, the second declaration replaces the first, and the inline
`onclick="runCalc()"` on every instance resolves to that one surviving function.

**Proven, not asserted** — `TestIDOnlyConversion_readsCleanOnIDsAndIsStillBroken`
(`component_instance_scope_test.go`): namespace the ids on today's real shape, render two
instances, and the detector reports **0 duplicate ids, 2 surviving `window.onload` assignments, 2
surviving global-scope scripts**. Mutation-checked — making the fixture script-scoped fails the
test, so it is sensitive to the thing it claims.

**Consequence for sequencing: ids and scripts must convert in ONE step, per component.** A phased
"ids first, scripts later" plan is worse than doing nothing, because it removes the only signal
(duplicate ids) that anything is wrong while leaving the wrong answer in place.

### 2.2 The IIFE route is FORCED, not chosen — the token is not a valid JS identifier

`{{.InstanceID}}` renders as `c-mortgages-repayment`. The obvious way to de-collide a global
function name is to suffix it with the token — `function runCalc_{{.InstanceID}}()` — and that is a
**syntax error**, because the token contains hyphens. Asserted by
`TestInstanceToken_isNotAValidJSIdentifier` so a converter author meets it in a test rather than on
a shipped page.

So each component's script must be wrapped in an IIFE, **and that forces the inline handlers to be
rewired** (`onclick="runCalc()"` cannot see a function inside an IIFE). That is the step which is
not safely mechanical.

## 3. What is mechanical and what needs judgement

| surface | rows affected (of 91) | mechanical? |
|---|---|---|
| `id="…"` attributes | 91 (1,346 attributes) | **yes** |
| `getElementById('…')` | 91 (886 calls) | **yes** |
| `<label for="…">` | **58** | yes — but silently breaks label-click if missed, with no error anywhere |
| id referenced from CSS in a `<style>` block | **33** | yes — same silent-breakage profile |
| `querySelector('#…')` | 34 | mostly, but the selector may be built by string concatenation |
| top-level `function` declarations | 21 | **no** — needs the IIFE |
| inline `on*=` handlers | 22 | **no** — must become `addEventListener` inside the IIFE |
| `window.onload =` | 8 | **no** — one slot, last write wins |

**58 `<label for=>` and 33 CSS `#id` references are the quiet ones.** Neither throws. A conversion
that handles ids and `getElementById` but forgets these ships pages whose labels no longer focus
their input and whose component-specific styling silently stops applying — on 94 live pages.

### 3a. ⚠ TWICE-CORRECTED (2026-08-17, same day) — the regex said 24, the detector said 88, and the truth is 25, because the DETECTOR had a defect the corpus run exposed

The history is kept because it is the useful part. Three classifications of the same 91 templates,
hours apart:

1. **Regex triage: 24 "need script work".** Three patterns; published as "~30" in an earlier §4.
2. **The production detector: 88.** ~~"That number does not survive contact with the real
   classifier … 88 of 91 need it."~~ **Wrong — and wrong because the classifier was wrong.**
   Sampling the 88 found that the estate's tool templates conventionally open their script with a
   `/* tool-doc */` comment block, and the detector's accepted-wrapper regex anchors at the body's
   first byte — so **62 correctly IIFE-wrapped scripts read as global.**
3. **The fixed detector** (leading comments skipped before the wrapper test; commit `5b30a831b`,
   mutation-proven both directions, council round 3 on the same correlation):

| | rows |
|---|---|
| script bodies **already scoped** | **66** |
| **genuinely declare into global scope** | **25** — of which 17 global-only, 8 also `window.onload` |
| **would produce duplicate ids if placed twice** | **91 — every one** (1,345 ids; the SQL census's 1,346 is one internal duplicate in `tool-spawn-rate-balancer`, an aria-title id repeated in markup and a JS string, bound by nothing) |

**The 25 are the 23 `loans-*`/`mortgages-*` calculators plus `tool-archetype-clash-calculator` and
`tool-bayesian-ranking`** — i.e. the original bug's "22 calculator templates" was very nearly the
right judged-work list all along. Also measured, because it is what makes "the 66 are actually
safe" a checked claim: all 20 inline `on*=` handlers and all 8 `window.onload` assignments sit
inside the 25, and there are **zero** `window.<name> =` assignments anywhere in the 66.

Three lessons, in the order they were paid for:

- **Do not size work with a hand-rolled proxy for an existing classifier** (the 24).
- **Do not read a gate's flag as ground truth without sampling the flags** (the 88). One eyeball of
  one flagged template found the comment; the false-flag rate was 70%.
- **A second implementation loses to the real one even when checking it**: an independent Python
  depth-walk said 65/26 and was wrong on `tool-css-specificity-calculator` — regex literals in its
  JS unbalanced the crude walk. The fixed production detector is right on it.

**The gate consequence is why the detector fix ships with this RFC rather than later:** this
programme uses `DetectInstanceCollisions` as its accept/reject gate, and the unfixed gate would
have refused 62 correct conversions — at which point someone mid-programme either "fixes"
components that are not broken or relaxes the gate in a hurry. Both are worse than one line now.

## 4. The decision the owner needs to make

Not *whether* — 283's defect is real and the owner already ruled that reuse should be a genuine
property of the platform. The question is **shape**, and there are three candidates:

**A. Deterministic converter as a new `fix_component_template` fix_type.** Reuses live machinery —
that action already rewrites `content_components.html_template` (`repair_template_slots`) and
already routes through the work-item/dispatch loop. Auditable, idempotent, re-runnable, and every
conversion is a reviewable diff. **Post-§3a-correction, this finishes 66 of 91 components** — their
scripts are already IIFE-scoped, so namespacing ids, `getElementById` strings, `label for=` and CSS
`#id` completes them, gate-verified. The remaining 25 must not ship half-converted (§2.1).

**B. LLM rewrite per component**, through the component-creator/tool-improver path, for scripts
that genuinely declare into global scope — **25 rows**, carrying all 20 inline handlers and all 8
`window.onload`s. ⚠ **This is the truncation class**: `bugs_open/012` saw a 10,272-char component
saved back as 1,253 chars, reported as success. Every result needs a byte-level structural check,
and `output_tokens == max_tokens` means CUT.

**C. Hybrid — A for the 66 and every mechanical surface, B for the 25 judged scripts**, one
component at a time, with the (now-corrected) `DetectInstanceCollisions` as the accept/reject gate
for both halves and nothing shipped half-converted.

**My recommendation is C, and after the §3a corrections it is a genuinely comfortable shape** — the
deterministic pass completes 73% of the estate on its own, and the LLM exposure is 25 components,
23 of which are the LMC calculators the bug was filed about, on the one domain with an independent
oracle (170 checks) to verify results against. The earlier draft of this paragraph said the LLM
must touch 88 scripts and floated narrowing the scope as the honest alternative; both followed from
the detector's false flags and are withdrawn.

## 5. What must be decided WITH the shape, not after it

1. **The byte-identical baseline.** Converting ends the byte-identical property the LMC lane's
   `b2_verify` verifies against. Rebaseline **before** the first conversion, deliberately, or the
   first batch discovers it.
2. **`oracle.py`'s 170 literal-id checks** move in lockstep with the component they address. This
   is why the token rule is function-based — one prefix per tool, not a per-page map — but the
   change still has to be made.
3. **Order relative to `RFC_032`.** 032 asks whether `ComponentID` should be re-pointed. If it is,
   and this conversion has already shipped, five more live components move a second time.
4. **When `enforce_instance_scope` is armed.** ⚠ The council's `render_guardian` seat: arming it
   before the 13 already-colliding pages are fixed would itself be *"a high-severity fail-loud
   violation"*. Convert → re-measure → arm.

## 6. Sources

- `bugs_open/283` (§9, §10 the council rounds; §11 this scoping); register **CLC-014**, **CLC-016**
- Council correlation `07635a2f-3605-4e67-9a6d-7636b07f16ca`, round 2 `architecture` seat
- `RFC_032` (the `ComponentID` unification this must be sequenced against); RFC_022's narrowing
- All figures: live `content_components` / `page_components` / `pages` / `sites`, 2026-08-17,
  queries in `bugfix_283_component_instance_scope/RUNBOOK` §10
