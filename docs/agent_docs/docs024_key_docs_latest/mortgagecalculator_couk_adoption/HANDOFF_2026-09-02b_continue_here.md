# HANDOFF — mortgagecalculator.co.uk — cold start, read this first (2026-09-02b)

**Supersedes `HANDOFF_2026-09-02_continue_here.md` (this morning's).** That file's §1, §3, §4 and
§5 still hold. **Its §2 is REFUTED — do not act on it**, see §2 below for what it got wrong and why
the mistake was structural rather than careless.

Everything here measured **2026-09-02 evening**.

---

## 0. The one-paragraph state

18 tool pages, all serving 200. None is known broken; **9 of the 18 cannot be checked by the
platform at all**, and where it does check, it is being lied to (`bugs_open/441`, filed today).
All 18 have their own hero image generated, active and deployed; **14 pages display no image**, for
two distinct reasons, neither of which is a render-path bug. **Nothing was changed on the live site
today** except one stale work item cancelled. The two real fixes both belong to other lanes and have
been written to them.

## 1. ⚠ The morning handoff's §2 is refuted: the tool pages have no hero, not a hero that fails

§2 said the `hero` component renders a background on guide pages and not on tool pages, "same
component, same field", and sent the reader to diff the render path. **There is no render path to
diff.** On the ten adopted tool pages the `hero` slot contains **the entire calculator**:

| page | slot | length | what the bytes are |
|---|---|---|---|
| `tool-simple` | `hero` | 9,590 B | `<div class="tool-page">…<h1>Simple Mortgage Calculator</h1>` |
| `tool-equity-release` | `hero` | 14,164 B | the calculator |
| `tool-equity-release-guide` | `hero` | 3,267 B | an actual hero, with `url('/assets/images/hero.jpg')` |

A hero is ~3.2 KB. `content_data.background_image` is inert on those rows because **nothing on the
page renders a hero**. ⚠ **Following §2 would have destroyed ten working calculators** — "make the
hero slot render" replaces the tool with a title band. That is `bugs_open/357`, and migration `701`
exists to make it impossible.

**Corroborated independently the same day, and they supplied the cause.** The
`bugfix_114_imagery_wiring` lane filed
`CONTRIB_2026-09-02_from_114_your_render_path_diff_is_bug_357_not_a_divergence.md` into this
directory with the same measurement and the mechanism I lacked: `adopt_fragment_section.go:14-15`
(verified, quoted from its own header) — a fragment is stored under the sentinel name `"section"`
(*identity unknown*) and **that sentinel is then replaced by `planned[Position-1]` from
`pages.sections`**, so the row inherits whatever identity the plan held at that position. The fix
is 357's constructive adoption (built, opt-in, default OFF, their lane mid-flight) — **not the
resolver and not more wiring**.

⚠ **Two corrections to the morning handoff's §2 framing that came with it:** the 12 content-hero
images are **not** waste — the event-driven card derive (IMG-073, live, 193/193 fleet-wide) uses a
content hero as the **card** source, and our tool cards exist and serve 200. And the
`undeployed_asset` mislabelling generalises to **1,651 rows fleet-wide**; there is a new LANDMINES
entry on why draining them is the trap.

**Why it happened, which is the reusable part:** the morning session reasoned from `content_data`
and `html_template` without once measuring `length(rendered_html)`. One column would have shown it.
And the CONTRIB that says so in its first paragraph
(`CONTRIB_2026-09-02_from_357_lane_migration_701_retypes_11_of_your_tool_rows.md`) was sitting
unread in this directory. **Read the inbound CONTRIBs before trusting the outbound handoff.**

## 2. The tools — verified as far as anything here can verify them, which is not far enough

**Done and sound:** all 18 probed with `scripts/probe-page-url.sh` (200, invented-URL control 404,
sibling control 200). All 18 fetched; every literal JS binding checked against the page's own ids:
**0 dangling, 0 template residue.**

⚠ **That is not a verdict and must not be quoted as one.** It only sees bindings written as literal
strings; `btl-investor` and `fee-analyser` report **0 literal bindings** (they bind through
variables — `bugs_closed/324`'s exact blind spot), so for those two the check saw nothing and said
nothing was wrong. **A real verdict needs the browser runner. Two things stop it:**

**(a) 9 of 18 pages are outside the ladder's eligibility.** Run the predicate — RUNBOOK **§15**
(added today) has the query. Ineligible: affordability, bridging-loan, equity-release, fee-analyser,
overpayment, portfolio, rate-forecaster, repayment, stamp-duty — each is multi-component with no
`component_level='tool'` row, satisfying neither eligibility clause. **7 of the 9 still hold a
current criteria fence** (installed 08-10/11/17) that nothing loads. **A page leaves this set by
GAINING a component** — adding a `generic-text-block` beside a calculator silently ends its
verification, with no error anywhere.

**(b) Where it does run, the criteria are unsatisfiable — `bugs_open/441`, filed today.**
`bugs_closed/283` renamed every tool's element ids to instance-scoped (`#foo` → `#c-<function>-foo`);
nothing updated the fences and neither tier knows the prefix. Fleet-wide: **134 of 187 anchor-absent
failures in 45 days name an element that exists**, across **99 tools**. The `090` came back
**UNVERIFIABLE** (iteration-cap, not refuted) and named two gaps; **both were closed by hand** and
are §3/§4 of the bug file. Key finding for whoever fixes it: **10 functions have both a converted and
an unconverted row and 6 have a fence — so re-emitting the fences CANNOT work**, only a scope-aware
checker can.

**Next action on the tools, in order:** (1) `441` candidate 1 (make the checkers scope-aware) — it is
a platform change, so council gate; (2) then re-run acceptance on the 9 eligible pages and read real
verdicts; (3) the other 9 become eligible only after migration `701`.

**Corrected from the morning handoff:** it recorded both failed `improve_tool` items as
"INFRASTRUCTURE, not the tool… acceptance unverified, not failed". That conflates two runs. The
`error` column is the *fixer's* infrastructure failure (`input_data.spec.page_id resolved to nil` —
real, still open). The `summary` is the *acceptance* verdict, which ran and failed on a concrete
selector — and that selector is stale criteria, i.e. `441`.

## 2a. The site divides in two on ONE property, and exactly one tool is verifiable today

Measured after §2 was written, and it is the sharpest way to hold this site in your head. All 48
`#id` selectors in the lane's 9 stored fences are **present on the live pages** — the fences are
not stale, they are unread. Why splits perfectly:

| shape | pages | ladder-eligible? | fence |
|---|---|---|---|
| **instance-scoped** `id="c-tool-…"` | 8 (bridging-compound, btl-investor, credit-health-check, deposit-tracker, overpayment-priority, rate-scenarios, rate-stress-test, remortgage-savings) | **yes** | **stale — `441`** |
| **bare ids** | 10 (affordability, bridging-loan, equity-release, fee-analyser, overpayment, portfolio, rate-forecaster, repayment, simple, stamp-duty) | **no, for 9** | **valid** |

The two axes are one axis: the conversion only touched components with a `content_components` row,
which is exactly the eligibility test.

> **CORRECTED 2026-09-02 evening, once the token came back.** This section originally concluded
> *"every tool the ladder can see has a broken fence"*. **Wrong — only TWO are broken.** I inferred
> staleness from "the page is scoped"; it actually requires the FENCE to name a scoped id, and most
> of these fences anchor on classes or on wrapper ids the conversion never touched. Re-tested by
> reimplementing `selectorAnchor` + `anchorPresent` exactly: `tool-deposit-tracker` (8 of 9 anchors
> absent) and `tool-remortgage-savings` (7 of 9) are stale; `tool-bridging-compound`,
> `tool-overpayment-priority` and `tool-rate-scenarios` are **satisfiable today**. Three more
> (`tool-btl-investor`, `tool-credit-health-check`, `tool-rate-stress-test`) have **no fence at
> all** — a cheaper, different problem. ⚠ My first cut of that test was ALSO wrong the other way:
> it matched the whole selector as an id, so `#bridgeForm button[type=submit]` read as missing.
> **Implement the platform's rule, don't approximate it.**

**The verification scoreboard, from the verdict record (2026-09-02):**

| tool | state |
|---|---|
| `simple`, `tool-overpayment-priority`, `tool-rate-scenarios` | **PASSING** (⚠ `simple`'s pass is desktop only — 4 mobile checks SKIPPED) |
| `tool-bridging-compound` | **stale FAIL** — repaired 08-26 12:21, page rebuilt 23:07, never re-run. Fresh run `21b2d81d` fired 21:27 |
| `tool-deposit-tracker`, `tool-remortgage-savings` | FAILING — `441` stale fence, and their fixer is blocked by `448` |
| `tool-btl-investor`, `tool-credit-health-check`, `tool-rate-stress-test` | no fence (`needs_criteria`) |
| the 9 adopted pages | ineligible — needs `701` |

⚠ **A FAIL outlives its own repair when the fixer does not re-run the check.** `bridging-compound`'s
newest verdict is a failure that predates its own fix by nine hours. Do not read the verdict record
as current state without checking `page_components.updated_at` against the verdict timestamp.

**`tool-simple` is the only exception** — bare ids *and* eligible (key `simple`). It is the one tool
here that can be verified today. ⚠ **And it is migration 701's designated pilot**: 701 moves its key
to `tool-simple` and orphans its fence, so **the one verifiable tool becomes unverifiable, and the
pilot looks perfectly healthy while it happens.** That is the concrete case for the re-key `UPDATE`
in the CONTRIB to `bugs_open/357` — it is no longer a hypothetical.

**Second defect on the two stuck items, diagnosed from source: `bugs_open/448` (new).** Their
`improve_tool` died at `input_data.spec.page_id resolved to nil`. `JudgeAcceptanceResultsAction`
re-derives `page_id` with the site filter on a LEFT JOIN and an unordered `LIMIT 1`, so a function
with rows on two sites yields empty and the key is omitted — while the sound code (read the handed
`input_data.spec.page_id`) sits 300 lines below in the same file, used only for the ported route.
**441 and 448 stack on those two tools: 441 makes the acceptance fail spuriously, 448 stops the
repair it queues from starting.**

## 3. The images — two mechanisms, both diagnosed, neither ours to apply

All 18 tool pages have an active asset at exactly their `ContentHeroKey`
(`'content_hero_'||replace(name,'-','_')`) — verified row by row. The resolver
(`plan_sections_action.go` `ensureAssets`) already prefers plan hero → content hero → site hero.
**It is never asked.**

**(a) Ten pages have nowhere to put an image** — the hero slot holds the calculator (§1). They need
their composition rebuilt, and **migration `701` is the prerequisite**: it retypes those rows to
`component_level='tool'`, after which the page can gain a real hero without touching the tool.
Piloting on `tool-simple` today. **Do not pre-empt it** — 701 aborts on an exact md5/position census
and our edits would fight it.

**(b) Four pages have a hero that structurally cannot show their image.** `hero-tool` emits
`url('{{or .hero_url .background_image}}')` but declares **no image-typed field**, and
`plan_sections_action.go:2846` gates the whole per-page hero path on `sectionHasImageField`. So the
aliases are never written and the template falls through to the site-wide default — for ever. The
code comment above the gate predicts exactly this. Fleet-wide: **`hero-tool` 69 instances, 0 with a
per-page image; `hero` (which declares the field) 632 instances, 72 with one.** **54 of the 69 have
their own asset going unshowable, across 21 sites.** Four-line fix, cannot regress anything —
**contributed to `bugs_open/114`** (owned, actively worked) rather than applied, because it is shared
across 21 sites.

**⚠ Do not "fix" (b) per page by writing `background_image` into `content_data`.** It works — one
leopardessconsulting page does it — and it is hand-setting a resolver-owned URL (MEMORY
`the-framework-writes-the-content-not-you`).

## 4. Written to other lanes today — chase these, they are not ours to land

- **`bugs_open/114`** — the `hero-tool` schema gap above, with the 0-of-69/72-of-632 control and the
  exact field to declare. Their fix, fleet-wide.
- **`bugs_open/357`** — 701 moves the ladder's subject key from `<slug>` to `tool-<slug>`, **orphaning
  all 8 of this site's fences including `tool-simple`, its own pilot's**. One `UPDATE` alongside 701;
  two cautions in the CONTRIB (the unique index, and that `doc_plans` keys are fleet-wide, so
  `stamp-duty`/`equity-release` may be shared with our two sibling domains). **We offered to do the
  enumeration for our three domains — pick that up if they ask.**

## 5. Unchanged from the morning handoff, still open

§4 (state), §5 (three unread CONTRIBs — the copy-quality one still **needs an answer**) and §6
(scorecard-simulator dead link; 13 `fact_drift_review`; contact-page email; the "Contact" title
question) all stand. The 12 `needs_imagery` items deferred since 08-02 are still deferred and
re-arming them is still real spend and still the owner's call.

**One item closed today:** `a7c5d5ab` (btl-investor 404 script) — **cancelled**, condition gone, 0
references live and 0 stored. ⚠ My first control for that was worthless (see `WRONG_CALLS.md`
2026-09-02) — `snippets.js` returns the same 0 because the assembler injects it. Use
`mortgage-lender-directory-listing.js`, which returns 1. RUNBOOK **§17**.

## 5a. ⚠ BLOCKED — kubeconfig token expired, and this is the resume list

`kubectl` returns `Unauthorized` fleet-wide (the known 3-day expiry; the owner refreshes). Resume in
this order:

1. **Confirm `tool-simple`'s last acceptance verdict** — the one tool that should pass today. If it
   did not, §2a's reasoning needs revisiting before anything else.
2. **Run `bugs_open/448` §5's two queries** and replace its `[UNMEASURED]` markers with real
   numbers. Do not quote a size for 448 before that.
3. **Re-emit fences for the 8 scoped tools** — `toolgolden.py --emit-criteria` drives the deployed
   page, so it picks up the `c-tool-…` ids by itself; then `verify_criteria.py` to 0 MISMATCH **and
   the mutation test exiting 1**, `install_fences.py --apply`, fire Tier-4, read the verdicts.
   ⚠ **Lane workaround, not `441`'s fix** — it re-breaks at the next conversion and cannot satisfy
   a split function. Check none of the 8 is split first (query in NOTES `## 2026-09-02 (c)`).

## 6. Files of record

`NOTES_mortgagecalculator_couk.md` `## 2026-09-02 (b)` (full working) ·
`README_where_we_are.md` 2026-09-02 later (owner's read-out, has **one decision** in it) ·
`RUNBOOK` **§15** (eligibility), **§16** (two extractor traps), **§17** (demand control) — and §14's
due-sweep paragraph is now **corrected**, it was stale and said the opposite of the truth ·
`bugs_open/441` (new) · CONTRIBs appended to `bugs_open/114` and `bugs_open/357` ·
`016b` §9 (the rename/describer pattern) · `WRONG_CALLS.md` 2026-09-02 ×3.
