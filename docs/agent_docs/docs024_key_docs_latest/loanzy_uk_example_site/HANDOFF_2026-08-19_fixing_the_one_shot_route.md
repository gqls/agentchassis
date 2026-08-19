# HANDOFF 2026-08-19 — fixing the one-shot route, from one greenfield build watched end to end

**What this is.** `loanzy.uk` was built on 2026-08-18 from its domain name and nothing else —
no mission, no contact details, no seed — and watched from dispatch to served page. This file
lists what the route got wrong, in the order that matters, with the evidence and the fix
candidates. It is written for whoever owns the one-shot route, not for this lane.

**The route today** (measured, not from a diagram): `082_submit_domain_unified.sh` →
`domain-submitter` → `needs_domain_research` → classifier → `needs_strategy` → strategist →
`needs_briefing` → briefing agent → `needs_site_plan` → planner → `needs_page` × N →
`page-build-handler` (which **deploys each page itself**, via its own `deploy_page` step) →
imagery → re-render. The webdesign.uk chat box is NOT in this chain: it stores transcripts on
its own box and dispatches nothing, so "one-shot" today means an operator runs the trigger with
the customer's brief.

**Live artefact this was measured on:** https://loanzy.uk — home, calculators, glossary,
get-help serving; `your-rights` and `guides/index` 404; one tool page serving with no tool in it.

---

## 1. A FAILED BUILD CANNOT BE RETRIED — re-submitting a domain silently does nothing

**The most serious item here, because it is the one a paying customer meets.**

`[MEASURED]` Re-running `082 … loanzy.uk` returned a **COMPLETED** orchestration and did
nothing at all. `domain-submitter`'s `create_research_item` returned:

```json
{"deduped": true, "inserted": false, "item_key": "research_loanzy.uk", "item_type": "needs_domain_research"}
```

`create_work_item` dedups on `item_key` **in any status, including `complete` and
`cancelled`**. Every stage key (`research_`, `strategy_`, `briefing_`, `site_plan_` + domain) is
therefore consumed for ever by the first attempt. A build that fails halfway can never be
re-run through the front door; the only way we got a second run was to **rename 78 rows'
`item_key` with a `_run1` suffix by hand**.

**Why it is not merely inconvenient:** the operator sees `COMPLETED`. Nothing anywhere says
"deduped". A support person would reasonably tell the customer the rebuild is running.

**Fix candidates, ordered by what closes the door:**
1. **Make the key carry the attempt** — `research_<domain>_<attempt>` or `…_<submission_id>` —
   so a re-submission is a new unit of work by construction and no dedup logic has to be right.
2. Dedup only against **non-terminal** rows (the partial-index contract `idx_swi_dedup` already
   implies), so a completed or cancelled stage does not block a fresh one.
3. At minimum, **make `deduped: true` fail loudly at the trigger** — a re-submission that
   silently no-ops while reporting COMPLETED is the dishonest surface here.

## 2. TOOL COMPONENTS: 7 of 7 lost, and the page ships hollow anyway — `bugs_open/311`

`[MEASURED]` All seven planned tool sections failed. Six were refused at
`store_generated_component`:

> *regeneration removes/renames 10 existing schema field(s) (button_1 … para_1) that dependents'
> content_data is keyed on*

The incumbent is **another live site's component** — `Loans Car Finance Calculator
(loanandmortgagecalculator.co.uk)` — which the selector could not find for reuse (it keys on
`section_type`) but the writer did find for overwrite (it keys on `function`). The guard is
correct: overwriting would empty that site's live sections. The seventh failed separately at
`generate_template` with `stop_reason=max_tokens` (a real truncation, and a reminder that
`output_tokens == max_tokens` means CUT, not finished).

**Then the page deployed anyway.** `https://loanzy.uk/tools/loan-comparison-calculator/index.html`
serves **200, 22,600 bytes, zero `<input>` elements** — prose describing a calculator that is not
there. Nothing tells the reader, and nothing stopped the deploy.

Full mechanism and fix candidates live in `bugs_open/311`; this lane's contribution (third site,
7/7, served-page evidence) is appended there. **The fleet consequence belongs in the fix
decision: ~140 finance domains are planned and the L-family shares calculator functions by
construction, so whichever site creates a function name first owns it and every later site ships
that tool hollow.** A shared component library that the second consumer cannot reuse is the
defect; the guard is only the messenger.

**One thing this lane would add to 311's candidate list:** whatever the keying fix, a section
whose component never materialised should **block its page's deploy**, or deploy it without the
section and file a visible item. A page that promises a tool it does not contain is worse than a
missing page.

## 3. A PAGE BLOCKED BY VALIDATION LEAVES A DEAD LINK ON A LIVE PAGE — `bugs_open/260`

`[MEASURED]` `your-rights` was refused at `validate_content` with **20 blockers, all
`unrendered_template` `{{end}}`** (`agent_error_log.error_code='CONTENT_VALIDATION_BLOCKER_DETAIL'`).
That is 260's known leak — a component rendering through a regex path no template uses. It fired
here on a **greenfield** site with zero component history, which narrows 260's trigger.

The route-level defect is what happens next: **the home page still links to `/your-rights.html`,
which 404s.** The build knows the page did not build; nothing removes or suppresses the link.
Same for `/guides/index.html`, which no-op'd honestly (*"no sections ready to build"* — correct,
there were no guides to index) and is still linked from the home page.

**Fix candidate:** at deploy time, a link to a page that is not `deployed` should be dropped or
degraded, not shipped. This is cheap and it is the difference between "a small site" and "a
broken site" from a customer's point of view.

## 4. THE DISPATCH ITSELF CAN VANISH SILENTLY

`[MEASURED]` One of three trigger runs published nothing: the script printed its correlation and
exited 0, and **no orchestration row and no work item ever appeared**, while 29 orchestrations
were created fleet-wide in the same ten minutes (so the chassis was consuming normally). This is
the documented `kcat -P` trap, hit for real. It cost nothing only because the item was checked
rather than the exit code.

**Fix candidate:** the trigger should verify its own landing — poll for the `needs_domain_research`
row for ~30s and exit non-zero if it never appears. A customer-facing route cannot have a
"submitted" that means nothing happened.

## 5. NAV AND CTAs: the site contradicted itself for about an hour

- **Stale nav served.** The nav rebuild refused (*"saw too little of its page corpus to replace
  the stored nav"* — a correct guard), so the live pages carried the PREVIOUS build's menu:
  *Check Eligibility · Lenders · Debt Consolidation · Car Loans*, pointing at pages that no
  longer existed, on a site whose body copy says it does not introduce anyone to lenders. It
  **self-corrected** on a later re-render, ~1 hour after the first page went live. On a customer
  build the first hour is exactly when someone looks.
- **9 × `unresolved_cta`**, every one *"no real-page destination"* — hero and call-to-action
  blocks pointing at pages that do not exist yet or never built. They are filed for a human and
  nothing resolves them.

**Fix candidate:** order the deploy so that nav rebuild and CTA resolution run once the corpus is
complete, and hold the first publish until then — or publish, but re-render the chrome
immediately rather than on the next unrelated trigger.

## 6. Smaller, but each cost time

- **`build_status='deployed'` survives retraction.** After `page-retraction` removed a page, the
  row still read `deployed` with a `deployed_at`. It is a lie about the artefact and it silently
  blocks the strategist's build-chaining gate (`bugs_open/304` family).
- **An archived same-name page row satisfies the planner.** After archiving old rows, the plan
  still named `index`, but the builder had an archived `index` row to work with; archived pages
  are excluded from rendering, so the home page would have built and never served. Un-archiving
  by hand is what made the apex live. **Any reset story for a site needs this in it.**
- **`site_unreachable` fires on deliberate maintenance** — it detected the window when routes were
  removed on purpose. Harmless, but it is noise in the queue a customer build would carry.

## 7. What the route got RIGHT, so a fix does not remove it

The refusals in this build were mostly **correct behaviour**: the nav rebuild refused rather than
writing a half-corpus menu; `guides-index` no-op'd rather than shipping an empty shell; the
component guard protected another site's live content; `validate_content` refused a page carrying
raw template syntax. **The defect is almost never the refusal — it is that the build carries on
and publishes as though the refusal had not happened.** Any fix that makes these quieter is the
wrong fix.

## 8. Where to start

`bugs_open/311` first (it is the one with fleet-scale multiplication and it is already
diagnosed), then §1 (retry impossible — cheap, and it is the customer-visible one), then §3's
dead-link suppression. §4 is a one-line addition to the trigger.
