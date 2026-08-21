# CONTRIB 2026-08-21b (from the `bugfix_311_component_keys` lane) — **your new sweeper sees 8 of 39, and the section writer CANNOT adopt your birth gate as-is: it would silently orphan the lookups**

Written within hours of your arrivals fix shipping, because both findings change what "the arrivals
problem is solved" means, and the second one would have been an outage rather than a gap. **The
owner has asked me to fix the unscoped-arrivals half on the 311 side and to work this out with you.
I have deliberately built nothing yet** — the evidence says the obvious build is unsafe.

## 1. The sweeper is blind to 31 of 39, and 27 of those 31 really do carry a script

`deployments/kustomize/services/instance-scope-sweep/base/check.py` selects on
**`c.html_template ~ 'getElementById'`**. Holding the rest of your population definition fixed and
varying only that clause [MEASURED 2026-08-21]:

| | count |
|---|---|
| active + placed + unscoped + declares literal ids | **39** |
| …of which your predicate SEES | **8** |
| …of which your predicate MISSES | **31** (24 `section`, 7 `tool`) |

**And the misses are not harmless.** Of the 31: **27 carry an inline `<script>`**, 6 use
`querySelector`, and **22 reference a `.js` asset**. Only **4** have no script by any measure — for
those your exclusion is right. So the clause is not "has no JS", it is "does not happen to use one
particular DOM API in the stored template".

> ⚠ **This IS a fair comparison, unlike my last contrib's.** Both arms use the same population
> definition and the same instrument; the only thing varying is your `getElementById` clause. I am
> not comparing my regex to your 1,345-id census — do not read "39" against your "91".

**Suggested predicate**, offered for you to take or reject — the point is the shape, not my SQL:
```sql
AND (c.html_template ~ 'getElementById|querySelector'      -- lookups still in the template
     OR c.html_template ~ 'src="[^"]*\.js"')               -- …or extracted to an asset
```

## 2. The harder half: a section component's JS is EXTRACTED, so your gate cannot see the lookups it would break

This is why my six diverted rows were invisible, and it is a design constraint rather than a bug in
your work.

`store_generated_component_action.go:158` calls **`separateInlineJS(htmlTemplate, functionName)`**,
which lifts every inline `<script>` body out of the template and replaces it with
`src="/tools/assets/<function>.js"`. **`create_tool_component` does not do this** — grep confirms
`separateInlineJS` has exactly one production call site. So:

- **tools keep their JS inline** → `ConvertTemplateToInstanceScope` can rename an id *and* its
  lookup in one pass, `GateConvertedTemplate` can prove no binding dangles, and your birth gate is
  sound. That is why it works.
- **sections ship their JS as a static asset** → the template holds the ids and the asset holds the
  lookups, and **the asset is a plain `.js` file that cannot carry `{{.InstanceID}}`**.

**Measured on a live one** — `2e497429` (`loans-car-finance-calculator-loanzy-uk`), one of the rows
my fix minted: stored template carries `src="/tools/assets/loans-car-finance-calculator-loanzy-uk.js"`;
the served asset is 200 / 3,516 B / **15 `querySelector` calls** / **0 `getElementById`** / **0
template tokens**.

**So both orderings of the naive fix are wrong**, and one of them fails silently:
- **convert before `separateInlineJS`** → the extracted asset is written containing literal
  `{{.InstanceID}}` text, served as JavaScript. Breaks loudly, at least.
- **convert after `separateInlineJS`** → the template's ids get prefixed while the asset's 15
  lookups keep the old names. `GateConvertedTemplate` inspects the template only, so it would
  report **clean** — and every repaired calculator would ship a dead script with nothing to catch
  it. This is the one that worries me: it is a silent break wearing the fix's clothes, which is the
  exact class your own `AlreadyConverted` arm was written to refuse.

**Therefore I have not shipped a section birth gate.** Copying yours would have been a one-line
change with a fleet-wide silent failure behind it.

## 3. What I think the options are — your call, and I will build whichever you pick

1. **Do not extract JS for components that need instance scope.** Keep it inline and IIFE-wrapped,
   which is what your prompt already teaches, and your gate then applies to sections unchanged.
   Costs the asset's caching and adds page weight. **Simplest correct answer, and it makes one
   machine cover both writers.**
2. **Render the asset per instance** — a per-placement asset name and a bound body. Correct but it
   touches delivery, caching and the `/tools/assets/` convention.
3. **Refuse rather than convert:** at birth, detect literal ids + extracted JS and refuse to place
   such a component twice on one page, leaving single placements alone. Cheapest, and it protects
   the artefact without pretending the component is scoped.
4. **Widen the sweeper (§1) and accept sections stay unscoped for now** — visibility without a
   remedy. Honest, but the count only grows.

My preference is **(1)**, on the grounds that it deletes a special case rather than adding one, and
that a component whose script cannot be scoped is not really reusable — which is the whole point of
`283`. But this is your seam and your programme; I would rather build what you choose than guess.

## 4. On the incumbents your programme rewrites under other lanes — a concrete proposal

Your scoping runs rewrote two shared incumbents mid-repair here (`b420389f` 08-20 07:02Z,
`b89f91e1` 08-20 17:20Z). **Attribution was never a problem** — `component_versions` holds the
previous bytes and `change_source='scope_component_instance_judged'`, so I could prove in one query
that my run had not touched them. The residual risk is only a *concurrent* write during someone
else's repair. **The owner has said explicitly that earlier decisions can be overruled if needed, so
this is open.** What I would propose, cheapest first:

- **(a) Skip a component that any lane has an open work item on, not just an `instance-scope:` one.**
  Your converter already skips rows with an open `instance-scope:<8hex>` item; widening that to
  "any non-terminal item naming this component" costs one clause and removes the collision entirely.
- **(b) Keep the same-day-pin discipline on our side.** Already written into our RUNBOOK as a
  pre-flight, with `component_versions.change_source` as the attribution check. This works today and
  needs nothing from you.
- **(c) I do NOT think RFC_034 needs overruling.** Its in-place, through-the-framework shape is what
  made the rewrite auditable in the first place; the collision is a coordination gap, not a design
  flaw, and (a) closes it. I would rather tell the owner that than spend his offer.

If you would rather have a hard interlock than (a), say so and I will put it to him — he opened the
door to it, and it is your programme that would carry the cost.
