# BUG 441 — the acceptance ladder selects by bare element id while instance-scoped tools render prefixed ids, so it reports a missing anchor for an element that is present

**Filed 2026-09-02** by the `mortgagecalculator_couk_adoption` lane, found while trying to answer
the owner's "verify the tools" on mortgagecalculator.co.uk and discovering that the platform's own
verifier cannot be trusted on any instance-scoped tool.

**Status: OPEN. Not fixed, nothing dispatched, no criteria edited.**

---

## 1. The claim, in one paragraph

`bugs_closed/283` converted interactive tool components to **instance-scoped element ids** — the
template carries `id="{{.InstanceID}}-foo"` and the page renders `id="c-<function>-foo"`. Nothing
updated the acceptance criteria, which still name `#foo`, and **neither checker knows the prefix
exists**. Tier 2 tests presence with an exact string match; Tier 4 hands the bare selector to
Playwright. Both therefore report a missing anchor for an element that is sitting on the page under
its new name. The failure raises an `improve_tool` item, so an automatic fixer is dispatched to
repair a tool that was never broken.

## 2. Verification status — READ THIS BEFORE QUOTING THE BUG

**The `090` diagnosis loop returned `UNVERIFIABLE`, not CONFIRMED** (intake `0c852424`, run
correlation `7177c2d6-fe22-40c4-b9bc-b53f93ec59c9`, work item `f49713ae`, stopped by
`iteration-cap`). It did **not** refute the mechanism; it stopped because its evidence bundle
lacked two things, which it named precisely:

> (1) … it is not established which one is the page actually served at
> `https://garden-tools.uk/tools/watch-service-interval-calculator/index.html` … (2)
> `check_tool_acceptance.go` (anchorPresent / evaluateStaticCriteria), named by the hypothesis, is
> not in this bundle's scope at all — there is no code evidence yet that the selector resolution is
> literal rather than instance-aware.

Per the owner ruling of 2026-07-31, this file therefore states plainly what first-hand verification
was substituted: **both named gaps were closed by hand, and §3 and §4 below are exactly those two
closures.** The filing rests on the artefact and the source, not on the loop's verdict. Do not cite
this bug as loop-confirmed — it is not.

## 3. Gap (2) closed: the resolution IS literal — the source, quoted

`platform/orchestration/actions/discovery_checks/check_tool_acceptance.go:563-567`:

```go
func anchorPresent(html, anchor string) bool {
	switch {
	case strings.HasPrefix(anchor, "#"):
		id := anchor[1:]
		return strings.Contains(html, `id="`+id+`"`) || strings.Contains(html, `id='`+id+`'`)
```

An exact substring match on `id="<id>"`. `#calc-btn` cannot match
`id="c-tool-watch-service-interval-calculator-calc-btn"`. There is no prefix-awareness anywhere in
the function, and `evaluateStaticCriteria` (`:423-512`) calls it for `selector_exists`,
`selector_count`, and — via `selectorAnchor` on every `steps[].selector` and `expect.selector` —
`interaction`. The Tier-4 half passes the same bare selector to Playwright
(`internal/adapters/browserrunner/run_checks_action.go`), which is why the failure surfaces as
`waiting for locator('#calc-btn')` rather than as a checker error.

## 4. Gap (1) closed: the same fence PASSES on one site and FAILS on the other — at the artefact

One `content_components.function`, `tool-watch-service-interval-calculator`, has **two active rows**
— one converted, one not — and `doc_plans` keys the criteria fence on `function`, so **both sites
are judged by one fence naming `#calc-btn`.** Curled 2026-09-02, both HTTP 200:

| domain | id actually served | `#calc-btn` matches? |
|---|---|---|
| `garden-tools.uk` | `id="c-tool-watch-service-interval-calculator-calc-btn"` | **NO** |
| `relojistas.com` | `id="calc-btn"` | **YES** |

`garden-tools.uk` is the URL the Tier-4 run failed against, at 2026-09-02 (doc_note `c8b6b243`,
`calc-shows-result@desktop: step 1 (click #calc-btn) failed: playwright: timeout`). `relojistas.com`
is the positive control: **same fence, same selector, same checker, and it passes** — so the fence
is not malformed and the checker is not simply broken. The only difference is whether that site's
row was converted.

**This is also the proof that editing the criteria cannot fix this.** One fence, two id shapes;
whatever you write, one of the two sites fails.

> **CORRECTED 2026-09-02, same day, before anyone acted on it — I measured the split on the wrong
> surface and both the number and the CAUSE were wrong.** The original text read: *"[MEASURED
> 2026-09-02] **10 tool functions are "split" this way** (≥1 converted row and ≥1 unconverted row
> under one function), and **6 of them have a current criteria fence**."* That counted
> `content_components.html_template` — **templates**. The fence is judged against the **deployed
> page**, so the surface that matters is `page_components.rendered_html`. Re-measured there:
>
> | | value |
> |---|---|
> | tool functions with live placements | **214** |
> | **split AT THE RENDERING** (some placements scoped, some bare) | **16** |
> | …of those, holding a current criteria fence — unsatisfiable by construction | **8** |
> | all-scoped / all-bare | 162 / 36 |
>
> **And the cause is not what I said.** These are not unconverted templates. `tool-credit-health-check`
> and `tool-rate-stress-test` each have every active row scoped, and still serve **bare** ids on
> `loancalculator.co.uk` — because those two placements were last rendered **2026-08-02** and
> **2026-08-09**, before the conversion, and a tool's `rendered_html` is written once and served
> verbatim. **A converted template does not convert the pages already built from it.**
>
> **What caught it:** checking whether this lane's own 8 tools were safe to re-fence. A per-placement
> count (`regexp_matches(…,'g')` over `rendered_html`) disagreed with the per-template count, and the
> two stale dates explained the gap.
>
> **This makes candidate 1 stronger, not weaker.** You cannot fix the split by converting templates —
> they are already converted. You would have to re-render every placement, and **any page that later
> goes stale re-breaks its fence**. Only a checker that accepts both spellings is stable under a
> half-rendered estate.

Those eight are unsatisfiable by construction today.

## 5. Size of the damage, with the control

[MEASURED 2026-09-02, live DB, 45-day window]

| measure | value |
|---|---|
| `acceptance-fail` notes naming an absent anchor | **192** (147 distinct tools) |
| …of those, parseable to a named anchor | 187 |
| **…naming an element that EXISTS in that tool's own template under `{{.InstanceID}}-`** | **134 (72%)** |
| **distinct tools affected** | **99** |
| all `acceptance-fail` notes, same window (context) | 410 (184 tools) |
| **passing `acceptance-run` notes, same window (the control)** | **178 (127 tools)** |

The control matters: acceptance is not universally failing, so "the ladder is just broken" is not
the explanation. The measurement could have come out near zero had the absent anchors been genuinely
invented selectors; it came out at 72%.

Re-runnable:

```sql
WITH n AS (
  SELECT subject_key, substring(body from 'anchor #([A-Za-z0-9_-]+) absent') AS anchor
    FROM doc_notes
   WHERE categories ? 'acceptance-fail' AND body LIKE '%anchor%absent%'
     AND created_at > now() - interval '45 days')
SELECT count(*) FILTER (WHERE anchor IS NOT NULL) AS notes_with_anchor,
       count(*) FILTER (WHERE anchor IS NOT NULL AND EXISTS (
         SELECT 1 FROM content_components cc
          WHERE cc.function = n.subject_key AND cc.is_active
            AND cc.html_template LIKE '%{{.InstanceID}}-' || n.anchor || '"%')) AS exists_scoped
  FROM n;
```

Shape check, independent of the notes: **172 of 173** instance-scoped tool functions holding a
current criteria fence have at least one bare `"selector": "#…"` in it.

## 6. The second-order damage: a fixer is dispatched at a working tool

A failing check files an `improve_tool` item, and `tool-improver` acts on it. Observed live while
this file was being written — doc_note `8045f367`, 2026-09-02 17:44, subject
`tool-budget-kit-builder`:

> *Root cause: **unknown**; anchor element was missing from rendered HTML template. Fix: Updated
> tool-budget-kit-builder-garden-tools-uk component HTML (19565 chars) to include the
> #activity-maintenance interaction anchor…*

⚠ **That specific instance is NOT proven to be this bug** — that tool's template carries a *bare*
`id="activity-maintenance"`, so its anchor-absent finding has some other cause and I have not
diagnosed it. It is quoted only for the shape it demonstrates and which this bug industrialises: a
`Root cause: unknown` rewrite of a 19.5 KB working component, triggered by an anchor-absence
finding. With 134 false absences over 99 tools, that path is being walked on tools that are fine.
**Whoever fixes this should sample what the fixer actually changed on the affected 99 before
assuming the churn was harmless.**

## 7. Fix candidates, ordered by what closes the door

1. **Make the checkers instance-scope aware (recommended).** `anchorPresent` accepts `id="<id>"`
   **or** `id="c-<function>-<id>"`; the Tier-4 runner resolves a bare `#id` to the scoped id when
   the bare one is absent. Fixes all 99 tools at once with no data migration, keeps every existing
   fence valid, and is the only candidate that survives the split-function case in §4 — a fence
   shared by a converted and an unconverted row is satisfiable **only** if the checker accepts both
   spellings. Structural: makes the bad state unrepresentable rather than repaired.
2. **Re-emit every fence from the live page.** The lane tooling already does this
   (`toolgolden.py --emit-criteria` drives the deployed page, so it picks the scoped ids up
   automatically). Correct per tool, but it is 172 fences, it re-breaks on the next conversion, and
   **it cannot satisfy the 6 split functions at all.** Reasonable as a follow-up, wrong as the fix.
3. **Re-convert or un-convert the 10 split functions** so one fence can be right. Narrower than it
   looks and does not address the other 89 tools; worth doing anyway for hygiene, after (1).

**Do not** "fix" this by teaching the *fixer* to ignore anchor failures — that removes the alarm
rather than the false alarm, and `bugs_open/084` is what happens when a ladder stops reporting.

## 8. How to verify a fix

- **Negative→positive at the artefact:** `garden-tools.uk`'s watch-service-interval-calculator must
  move from fail to pass on `calc-shows-result` **without its rendered_html changing** (pin the md5
  first). If the bytes moved, something rewrote the tool and the test is void.
- **The control must stay green:** `relojistas.com`'s copy passes today and must still pass.
- **Induce the red:** a criteria fence naming a genuinely invented anchor
  (`#definitely-not-present-xyz`) must still FAIL after the change. Without this, a checker that
  accepts everything looks identical to a checker that was fixed — and this bug's whole class is
  "a check that cannot tell present from absent".
- **Re-run §5's query:** `exists_scoped` should fall to ~0 for notes created after the fix. Compare
  windows, not totals, and re-run the ORIGINAL query.

## 9. Related

- `bugs_closed/283` — the instance-scope conversion that renamed the ids. Its five `CONTINUE_HERE`
  files never mention criteria or acceptance; this is its unswept consequence.
- `bugs_closed/324` — "a converted component reads clean on every id check and every binding through
  a variable dangles". Same conversion, adjacent blind spot.
- `bugs_open/357` / migration `701` — retypes adopted rows to `component_level='tool'`, which moves
  the ladder's subject key from `<slug>` to `tool-<slug>` and orphans existing fences. Independent of
  this bug, but the same fences are involved; CONTRIB filed there 2026-09-02.
- `bugs_open/126` — gated-tool acceptance; the `acceptance_stuck` human path.

---

## UPDATE 2026-09-03 — this is not a backlog of stale fences. It is a LIVE GENERATOR of them, and it fired five times in four minutes.

Filed yesterday, this bug reads as damage left behind by `bugs_closed/283`'s August conversion — a
fixed population of 99 tools to repair once. **That framing is wrong and understates it.**

### What happened overnight, on one site, without anyone intending it

Migration `701` (owner-applied 2026-09-02 ~22:00Z, `bugs_closed/357`) retyped
mortgagecalculator.co.uk's 11 adopted tool rows to `component_level='tool'`. It created each new
component with an **instance-scoped template** (`{{.InstanceID}}-`) while **preserving the existing
rendered bytes** — and verified exactly that, by md5, per row. Correct, and green.

> **⚠ CORRECTED 2026-09-03 (later the same day): the id rewriting was NOT 701.** This section says
> 701 created the adopted components with instance-scoped templates. It did not — 701 adopted the
> bodies verbatim, with **bare** ids, exactly as its own md5 evidence says. The rewriting was a
> **separate actor**: the instance-scope sweep filed 11 `instance_scope_conversion` items at
> **07:40:15** on 2026-09-03 (*"uses getElementById without `{{.InstanceID}}`"* — it found them
> unconverted, which they were), those completed **08:36–08:46**, and every 701-born row's
> `content_components.updated_at` moves from its `created_at` of `2026-09-02 21:06:35` to that
> window. Each conversion then filed a `page_rerender` (`reason: template_changed`), and five ran at
> 08:46–08:49.
> **Everything below about the CONSEQUENCE stands** — half-converted estate, fences broken by a
> render, 441 as a live generator. Only the cause changes, and it changes for the better: **three
> actions, each correct in isolation** (adopt verbatim → convert an unconverted component → publish
> it), composed into broken verification that none of them could see. That is a predictable sequence
> every future adoption will meet, not a defect in anyone's migration.
> **What caught it:** the site's own completed work items, which were visible all along. I inferred
> cause from which migration had run most recently — coincidence in time, not mechanism.



Those two states agree only until something renders. The routine rebuild wave of **2026-09-03
08:46–08:49Z re-rendered 5 of the 10**, and their served element ids changed:

```
tool-simple      id="amt"          ->  id="c-tool-simple-amt"
tool-repayment   id="calculateBtn" ->  id="c-tool-repayment-calculateBtn"
```

**Five acceptance fences went from satisfiable to unsatisfiable in four minutes**, as a side effect
of a rebuild nobody connected to acceptance. Measured immediately after: template scoped **10 of
10**, rendering scoped **5**, rendering still bare **5** — and those 5 will convert whenever they
next render.

**Nothing was broken by the conversion itself.** All ten checked for dangling JS bindings: **0**.
The converter rewrites bindings alongside ids, as designed. The tools work. Only the *checks* broke.

### Why this changes the fix, not just the size

- **Candidate 2 (re-emit the fences) is now clearly insufficient**, not merely a follow-up. A fence
  regenerated today is invalidated by the next render of its own page. There is no point at which
  the estate is "converted" and the fences can be brought into line once — **every re-render of an
  adopted-then-scoped tool is another instance.**
- **Candidate 1 (scope-aware checker) is the only stable option**, because it is the only one whose
  correctness does not depend on when a page last rendered.
- **The half-rendered state is normal and will persist.** 5 of 10 tools on this one site are mid-
  conversion right now. Any fix that assumes a page is either "converted" or "not" will be wrong for
  whichever half it did not expect — which is exactly the split-function case in §4, arriving by a
  second route.

### A general lesson for adoption migrations, worth taking beyond this bug

**"Bytes unchanged, md5-verified" is a true claim that expires.** It proves the migration moved
nothing. It cannot promise the *next render* will not, when the migration has changed the template
that render will use. `701`'s verification was sound and its closing evidence was green; the
divergence appeared eleven hours later on a rebuild that had nothing to do with it. **An adoption
migration that rewrites a template should either state that the next render will change the served
bytes, or render one page itself and check.** Contributed to `bugs_closed/357` post-close as
information, not as a reopen — the migration did what it said.

### This lane's own fences are repaired, and that is NOT a fix for this bug

All 8 mortgagecalculator fences were re-addressed today (subject key `<slug>` → `tool-<slug>`, forced
by 701; selectors re-pointed to the `c-tool-<slug>-` prefix on the 5 already-rendered pages), each
verified present in the live page first, with the old fences re-tested as the control (4/4/5/7/6
anchors absent — so the transform did real work). **That is candidate 2 applied by hand to one site,
and it will rot the moment the other five pages render.** It buys this lane working verification
today; it is not evidence the bug is smaller.
