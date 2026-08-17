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

### 3a. ⚠ CORRECTION 2026-08-17 — the split above is REGEX TRIAGE, and it was wrong in the direction that made the job look easy

An earlier version of §4 said the script half was *"the ~30 rows whose scripts need judgement"*,
derived from three regexes (`window.onload`, inline `on*=`, a `function` keyword near the top of a
script). **That number does not survive contact with the real classifier.**

`DetectInstanceCollisions` — the detector this lane already built, and the one that will gate
acceptance — was run over all 91 live templates
(`cmd/instanceaudit`, one-off, reads the templates and calls the production function):

| | rows |
|---|---|
| script bodies **already scoped** (0 unscoped inline scripts) | **3** |
| **declare into global scope** (≥1 unscoped) | **88** |
| of those, also assign `window.onload` | 8 |
| **would produce duplicate ids if placed twice** | **91 — every one** |
| total duplicate ids across all 91, doubled | **1,345** |

So the script work is not a minority tail: **88 of 91 need it.** The regex triage found 24 because
it searched for three specific old-fashioned spellings; the other 64 declare globals in ways it did
not look for.

Two notes on trusting this number rather than the regexes:

- **It is the same classifier that will accept or reject each conversion**, so its verdict is the
  operative one even where a hand-reading might disagree. The detector's own docstring says it errs
  toward *reporting* — a script wrapped in a form it does not recognise reads as unscoped — which
  makes 88 a ceiling on "needs work" and, more usefully, a floor on "must satisfy the gate".
- **It corroborates the independent SQL census**: 1,345 duplicate ids by the Go detector against
  1,346 literal `id=` attributes counted in SQL — two different code paths, agreeing within one.

## 4. The decision the owner needs to make

Not *whether* — 283's defect is real and the owner already ruled that reuse should be a genuine
property of the platform. The question is **shape**, and there are three candidates:

**A. Deterministic converter as a new `fix_component_template` fix_type.** Reuses live machinery —
that action already rewrites `content_components.html_template` (`repair_template_slots`) and
already routes through the work-item/dispatch loop. Auditable, idempotent, re-runnable, and every
conversion is a reviewable diff. ⚠ **Post-correction, this finishes 3 of 91 components.** The other
88 declare into global scope, and §2.1 forbids shipping them half-converted — so on its own this
option converts the names on 88 rows that must then sit unshipped until something does their
scripts. It is a *component* of the answer, not an answer.

**B. LLM rewrite per component**, through the component-creator/tool-improver path. Handles the
script half, which A cannot, and post-correction that is 88 of 91 rows rather than a tail. ⚠ **This
is the truncation class**: `bugs_open/012` saw a 10,272-char component saved back as 1,253 chars,
reported as success. An LLM rewrite of 91 templates needs a byte-level structural check on every
result, and `output_tokens == max_tokens` means CUT. It also rewrites 1,207,640 bytes of working
production markup to fix a scoping problem, which is a wide blast radius for a narrow defect.

**C. Hybrid — A for the mechanical surfaces, B for the script half**, one component at a time, with
`DetectInstanceCollisions` as the accept/reject gate for both halves and nothing shipped
half-converted.

**My recommendation is still C, and the correction strengthens rather than weakens it** — but the
balance inside C has shifted a long way toward B, and that is the thing to be clear-eyed about.
Before the measurement, C looked like "a program does most of it, an LLM mops up ~30". It is
actually "a program does the names on all 91 reliably, and an LLM must touch 88 scripts". If that
LLM exposure is unacceptable, the honest alternative is **not** option A — it is to narrow the
*scope*: convert only the components someone actually wants to place twice, and leave the rest
literal. That is 283's original candidate B, which the owner declined on 2026-08-15 for good
reasons, but it is the only shape that avoids rewriting 88 working scripts.

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
