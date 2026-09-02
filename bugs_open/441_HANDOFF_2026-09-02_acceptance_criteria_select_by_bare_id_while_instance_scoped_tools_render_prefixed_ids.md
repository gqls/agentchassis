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
whatever you write, one of the two sites fails. [MEASURED 2026-09-02] **10 tool functions are
"split" this way** (≥1 converted row and ≥1 unconverted row under one function), and **6 of them
have a current criteria fence**. Those six are unsatisfiable by construction today.

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
