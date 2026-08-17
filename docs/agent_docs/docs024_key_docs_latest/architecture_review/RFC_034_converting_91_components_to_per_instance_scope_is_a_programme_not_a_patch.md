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

## 4. The decision the owner needs to make

Not *whether* — 283's defect is real and the owner already ruled that reuse should be a genuine
property of the platform. The question is **shape**, and there are three candidates:

**A. Deterministic converter as a new `fix_component_template` fix_type.** Reuses live machinery —
that action already rewrites `content_components.html_template` (`repair_template_slots`) and
already routes through the work-item/dispatch loop. Auditable, idempotent, re-runnable, and every
conversion is a reviewable diff. Cost: the script half (21 + 22 + 8 rows) is not safely regex-able,
so this converts the mechanical surfaces and **parks the rest for human or LLM handling**, which
by §2.1 means those components must not be half-converted at all.

**B. LLM rewrite per component**, through the component-creator/tool-improver path. Handles the
script half, which A cannot. ⚠ **This is the truncation class**: `bugs_open/012` saw a 10,272-char
component saved back as 1,253 chars, reported as success. Any LLM rewrite of 91 templates needs a
byte-level structural check on every result, and `output_tokens == max_tokens` means CUT.

**C. Hybrid — A for the mechanical surfaces, B for the ~30 rows whose scripts need judgement**,
with the detector as the acceptance gate for both and nothing shipped half-converted.

**My recommendation is C**, sequenced per component rather than per surface, with
`DetectInstanceCollisions` run against the assembled page as the accept/reject gate. But the
sequencing question is the owner's, because it decides whether 94 live pages change over days or
weeks.

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
