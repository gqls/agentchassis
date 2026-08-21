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

---

## UPDATE 2026-08-19 (afternoon) — conferred with the owning lanes; §2 is now LIVE-FIXED and the next run is its test

**`bugs_open/311` — FIX IS LIVE on `v1.0.1315`.** Settled by that lane with controls, not
asserted: the running binary stamps `590ca3a20…` (positive) with a fake sha absent on the same
pod (negative), `git merge-base --is-ancestor 17d883333 590ca3a20` is TRUE, and the fix's own
literal `COMPONENT_COLLISION_DIVERTED` probes PRESENT.

> **⚠ My own probe said ABSENT and was measuring the wrong question.** The binary carries only
> the BUILD commit's sha, so **any ancestor commit reads ABSENT even when its code is in the
> image**. Grepping for your own fix's sha is not the check; getting the stamp and then
> `merge-base --is-ancestor` is — exactly as CLAUDE.md already says. I did not follow it.

**The section/tool distinction, which decides whether our run tests anything** (from the 311
lane, verbatim in substance): the one-shot path — `plan_sections → needs_new_component →
component-creator → store_generated_component` — is **entirely the SECTION-level writer**, and
that is the half now live. `create_tool_component` belongs to the **other** route
(`tool-suggester → add_tool → tool-generator`), which neither finance build has ever triggered:
no `add_tool` item exists on either site. So the outstanding tool-level half — the owner's
stated precondition for wave 1 — **does not gate a clean-domain run**, though it will fire on
portfolio sites later and on webdesign's two parked tools.

### The protocol for the next clean-domain run (agreed with the 311 lane)

Check BOTH halves, and the artefact independently of both:

1. **Diversion worked.** Per diverted section: a new `content_components` row
   (`function='<function>-<domain>'`, `forked_from` NULL, `section_type` = the requested section
   name, so the selector can see it), one `COMPONENT_COLLISION_DIVERTED` row in
   `agent_error_log`, and the work item **complete** rather than failed. Expect **6 of 7** —
   the `generate_template` / `stop_reason=max_tokens` failure is upstream of the store and this
   fix does not touch it.
2. **No collateral damage.** The 311 lane pinned the incumbents' HTML md5s BEFORE our run
   (`bugfix_311_component_keys/NOTES_311_fix.md`, commit `527193376`): `824e3309` →
   `e6ee4b07f11d0b43c1c5a62667f4999f`, `b89f91e1` → `a2c00f1c66ce6f4ef72b48083f1e3da6`,
   `7d8b0503` → `5f9534982e7f2bd776605ed78e755010`. **Re-read them after the run; they must be
   UNCHANGED.** A run where our site gets its calculator by overwriting
   `loanandmortgagecalculator.co.uk`'s is the exact damage the old guard prevented.
3. **The served page, independently.** *"The fix guarantees the diverted component gets STORED
   and LINKED, not that the LLM generates a GOOD calculator."* So keep the artefact check that
   caught the hollow page last time: fetch the tool URL and **count `<input>` elements**. A
   stored, linked, selector-visible component that still renders no tool is a different failure
   and must not be reported as success.

### `bugs_open/260` — owned, fix in flight, and it WILL still fire

That lane took the renderer half this morning and is the only session on it. The defect is Go
(the silent fallback in `RenderTemplateReportingMissing`, `component_library.go:965`, fallback
at `:1032-1057`), so **inert until an image rolls; there is no config half that can land
sooner.** Two things they asked for and got: occurrence contributions continue as two-line
pointers (domain · date · work item · what differs), and the reproduction is **held** until
after the roll, when this lane runs the greenfield build as their "after" test.

**A design constraint the owner's tools ruling created for them, recorded because it will shape
what "fixed" means:** tool pages legitimately contain brace literals (a prompt library, a syntax
gallery), so a fix keying on "contains `{{ }}`" would start refusing pages that work. The
discriminating shape is Go CONTROL syntax with no executed output, not brace presence.

**And a measurement caveat that now applies to this lane's numbers too:** distinct-token sets
inherit `validate_content`'s 10-per-detector cap, so any "N tokens" figure is a **ceiling**, not
a count. Report it as such.

---

## READINESS CHECK 2026-08-21 — for the next clean-domain run (`garden-tools.uk`, owner's choice)

**`bugs_open/260` is CLOSED** (2026-08-20): *"the renderer half is FIXED AND LIVE, verified"*, and
its headline's ~~"no live damage"~~ was restated as "zero corruption of STORED content" as that
lane said it would. The blocker that stopped the last run is gone.

**Also closed since:** `286`, `317`, `331`, `323`. **`311`'s section-level half remains live** and
is what our route uses.

### Still open, and what each would do to this run

| bug | risk to a clean-domain build |
|---|---|
| **`307`** | a transient infrastructure burst kills work items **terminally** — all three attempts fit inside the outage. Its own record: ~815 failed steps, 100 items reaching a terminal state. A build running through an outage loses pages permanently, with no retry left. |
| **`326`** (this lane) | **if the build fails partway, it cannot be re-run.** Re-submitting reports COMPLETED and queues nothing. The recovery path is hand-renaming `item_key`s. |
| **`327`** (this lane) | the trigger can publish nothing and exit 0 (1 of 3 last time). Mitigated by procedure — verify the `needs_domain_research` row, never the exit code — but unfixed. |
| **`328`** (this lane) | any page that fails to build stays linked from the pages that did, turning one failure into a visibly broken site. |

### The blocker that is NOT a bug: `garden-tools.uk` cannot serve anything yet

`[MEASURED 2026-08-21]` The domain is **parked**: NS are `ns1/ns2.dan.com`, A records
`13.248.169.48` / `76.223.54.146`, and it serves **HTTP 200** from a parking page. There is **no
Cloudflare zone** (`GET /zones?name=garden-tools.uk` → `success: true`, empty result), therefore
no worker route. A build would write pages to the bucket that **nothing serves**.

Prerequisites, in order — the same path `loanzy.uk` took: confirm ownership (⚠ the `webzy.uk`
precedent: a zone was created for a domain the owner did not own, and the GODADDY tag was the
tell) → delegate to Cloudflare (Nominet EPP, DESIGNCONSULT tag) → create the zone → add the
`garden-tools.uk/*` and `*.garden-tools.uk/*` worker routes to `portfolio-sites-router` → verify
the apex returns the router's 9-byte 404 **before** dispatching.

### Confirmed clean

0 rows in `sites`, 0 work items, no register entry — and it is **deliberately** absent from the
register: `portfolio_positioning/RESERVED_test_domains.md` names it as a reserved test domain,
*"parked, not in the register, not among the 50 … a subject that naturally exercises guides, a
supplier/brand directory, a calculator-shaped tool and editorial, with no regulated angle"*. So
no register work is owed and no other lane's wave plan is disturbed.

### Correction to a premise, measured rather than assumed

**`lendzy.co.uk` DOES have tools, and they work.** Four tool pages serve 200 with real inputs
(affordability-complaint-checker 41 inputs, price-cap-checker 3, true-cost-calculator 1,
complaint-deadline-calculator 2), and three are linked from the home page body. What is missing is
**navigation**: the nav reads *About · Check your loan · Your rights · Free help now* — no tools
or calculators entry — so browsing the menu you would never find them. The estate-wide census
also refutes "the framework does not build tools": webdesign.co.uk has 92 deployed tool pages,
mortgagecalculator.co.uk 14, loancash.co.uk 8. **The defect is discoverability, not absence** —
and it is worth its own bug if the owner wants tools in the nav by default.
