# PLAN — bugs_open/084, asset reference resolution

**Date** 2026-08-05 · **Bug** `bugs_open/084_HANDOFF_2026-07-26_owned_pages_get_no_browser_verification_and_script_urls_are_never_fetched.md`

> **Note on how this plan was produced.** The brief asked for the plan to be
> drafted by `fable`. That agent terminated on a model session limit before
> returning, so the plan below is mine, written against the same brief and the
> same research. Recording it because "who wrote the plan" is the sort of detail
> a reader will otherwise assume.

## What the bug is, restated after measuring it

Behavioural verification is opt-in and narrowly scoped. **Nothing anywhere asserts
that a subresource a deployed page references actually resolves.** A `<script src>`
pointing at a 404 renders a page that looks finished and does nothing, and every
status in the database says `complete`.

The bug lists five fix candidates. After research, three are closed to me:

| candidate | disposition |
|---|---|
| 1 — make `asset_loads` actually load | **OUT OF SCOPE — already ruled RFC-scope, twice.** See NOTES. Touching it would also flip `asset_loads` from `experienceStaticConfirming` to `experienceStaticRefuting` (`static_attribute_checks.go:95-112`), which is verbatim the RFC_002 trigger, and would fail `TestEveryStaticCheckTypeIsClassified`. |
| 2 — a reference-resolution discovery check | **THIS PLAN.** Unbuilt, unnamed anywhere else, and the bug's own "single highest-value item". |
| 3 — widen Tier 4 past `component_level='tool'` | **ALREADY DONE** — `tool_eligibility.go`, registered TL-033. 084 §1 is stale and gets a correction. |
| 4 — build T5.1 post-hydration dead-control assertion | out of scope; a browser-tier build, not a bug patch. |
| 5 — generalise `checkScriptParity` | out of scope; a distinct class (a WRITER losing content) already worked by other threads under `bugs_open/178`/`198`. |

## The honest size of it, stated before the design

**The current population of broken references fleet-wide is ZERO** — 96 of 96
distinct referenced assets return 200 (measurement in NOTES). So this is a
**regression guard**, not a repair. The class has bitten: `bugs_closed/041`
(chrome `js_content` published to a path nothing loaded), and `cmd/webdesignport`
carries `checkScriptParity` because ~60 tools nearly shipped as dead markup —
caught, per `WRONG_CALLS.md`, by luck.

That has one hard consequence for the plan: **there is no live positive to prove
the check bites, so the proof has to be induced.** Step 7 is the part of this
plan I care most about.

## Design

### The one file

`platform/orchestration/actions/discovery_checks/check_asset_reference_404.go`
Check name and `item_type`: **`asset_reference_404`**.

Named as a sibling of `image_url_404` and deliberately not `script_*`: it owns
`<script src>` **and** `<link rel="stylesheet">`, which fail the same way for the
same reason and cost one probe each. The header states the boundary explicitly —
`<img>` belongs to `check_image_url_404`, brand-head `favicon`/`og_card` to
`check_undeployed_assets` — because four checks already share the asset space and
the landmine keyed to that file warns that widening one silently competes with
another.

### Why it earns its place beside the browser tier

TL-032 (`register/tool-lifecycle.md:288`) rules that the external-script class
"belongs to the BROWSER tier". The answer is the one
`check_orphan_element_refs.go:22-23` already makes for itself: **the browser tier
is criteria-gated.** Tier 4 runs only where a PLAN carries a ` ```criteria `
fence — 1 of 71 newly-eligible components at TL-033 time — and it never reaches
`site_components` chrome, ordinary sections or `js_snippets` at all. A check that
needs no criteria, no browser and no PLAN covers the population Tier 4 cannot,
which is most of what the fleet serves.

### Surfaces (mirroring `check_image_url_404`, plus the page URL it must have)

1. **`page_components`** — `build_status='deployed'`, `locked_at IS NULL`, joined
   to `pages` for `p.url`, and `p.deployed_at IS NOT NULL`. The join is not
   cosmetic: a relative `src` is meaningless without the page URL it resolves
   against, and 17 of webdesign.co.uk's references are relative.
2. **`site_components`** — `locked_at IS NULL`. Chrome has no page URL, so only
   root-relative and absolute references can be resolved there; a page-relative
   reference in shared chrome is **skipped and logged**, never guessed.

### The probe, and the rules that make it safe

- **GET, not HEAD.** A static origin need not implement HEAD, and a 405 would be
  indistinguishable from a policy refusal. Body discarded through a 1 KB
  `LimitReader` — the status is the whole answer.
- **ONLY 404 and 410 are findings.** 2xx, 3xx, 401/403/429, every 5xx, timeouts,
  DNS and TLS errors all **SKIP**, counted and logged. This is the single most
  important property of the design, and it is what makes the Cloudflare landmine
  harmless: a bare non-browser request was refused for all 63 webdesign.co.uk
  tools in a prior Python sweep, which under a "non-200 is a finding" rule would
  have filed 63 false 404s **on the very site this bug is about**. Under this rule
  a refusal is a 403 and files nothing.
- **A 404 is confirmed by a second request before it is filed.** One extra GET per
  candidate finding, and candidates are rare by construction.
- **`curl`-style `000` does not exist here** — a transport error is an error, not
  a status, and lands in the skip tally.
- **An empty or `#` reference is never probed.** Per the HTML spec an empty `src`
  resolves against the current document, so the probe would score a broken
  reference **200** — `bugfix_128`'s recorded landmine. It is reported
  structurally instead, as `kind: "empty_src"`.
- **URLs are resolved, never constructed** — `net/url`'s `ResolveReference`
  against `https://<domain><pages.url>`. TL-032's own false positive came from a
  built URL: "a verdict from a URL you built is a verdict about a page you
  invented".
- **Cost bounds, logged not silent** (copying `check_backend_entry_orphaned`'s
  block): identical resolved URLs deduped per site so each is probed once;
  `maxProbeURLs = 40` per site with the dropped remainder **logged**; 10 s per
  request; 4 workers.

### Parse the DOM, never regex the HTML

`goquery` (already a dependency, used by this same package). This is not a style
preference: my own first measurement regexed `<script[^>]+src="..."` over raw
`rendered_html`, matched **a JS comment describing a regex**, and I nearly filed a
phantom production 404. Tool pages — this bug's whole population — are the pages
most likely to talk about HTML inside JavaScript.

### Work item

| field | value |
|---|---|
| `ItemType` | `asset_reference_404` |
| `ItemKey` | `asset_reference_404:<kind>:<resolved URL>` — the full URL, because `app.js` under two tool directories are two files with two HTTP results, and a key that cannot tell them apart lets `idx_swi_dedup` silently drop the second (`bugs_open/091`'s failure mode) |
| `Severity` | `high` for a confirmed 404 on either surface; `medium` for `empty_src`. Severity deliberately does **not** vary by surface — a dead script breaks the page it is on just as completely as one in chrome breaks every page — and the surface travels in the spec instead |
| `HandlerAgent` | `""` — flag-only |
| `Priority` | 40, matching `image_url_404` |
| spec | `check`, `kind`, `url`, `reference`, `element` (`script`/`stylesheet`), `surface`, `page_url`, `http_status` |

**Flag-only is a decision, not a default.** The repair for a 404ing reference is to
remove it, repoint it, or republish the file — a judgement, not a transform, and
no generator can make it. That is the same reasoning
`check_orphan_element_refs.go:45-50` and `check_image_url_404.go:274` record. It
is also explicitly **not** RFC-scope: CLAUDE.md's owner ruling of 2026-08-02 §1
says a work-item type with no automated consumer is not the kind of shared
vocabulary whose guarantees change. The cost is real and named in
`bugs_open/083:308` — "a detector whose output nobody drains is not neutral, it is
actively misleading" — so the finding's audience is stated in the header: the
`needs_human_review`-shaped queue plus the run's `Findings`, which is where
`image_url_404` and `dead_control` already land.

### Retraction

`CheckResult.Resolved` with `ItemKey`, on a **positive observation only**: a URL
still referenced that now returns 200. Never `AllOfType`, and never inferred from
an empty result — a check that was blinded returns exactly that. A reference that
was *deleted* rather than repaired leaves its item open, because the check can no
longer observe the URL at all; that gap is stated in the header rather than closed
by guessing.

## Steps

1. Write `check_asset_reference_404.go` — header first, in the shape of
   `check_orphan_element_refs.go`: what it owns, what it does not, why it sits
   beside the browser tier, the probe rules, the landmines it is written against.
2. `init()` → `Register(&AssetReference404Check{})`.
3. Collect: two SQL surfaces above; `goquery` extraction of `script[src]` and
   `link[rel~=stylesheet][href]`; resolve each against its page URL; tally
   empty/`#` separately.
4. Probe: worker pool, skip taxonomy, 404 confirmation, caps logged.
5. Emit work items + findings + `Resolved`.
6. Classify `asset_reference_404` in `verifier_coverage_test.go` (`catMechanical`,
   with the completion-path reasoning spelled out in the same shape as
   `image_url_404`'s and `backend_entry_orphaned`'s entries).
7. **Tests, including the proof it can fail.** A pure-function seam
   (`probeStatus` as a package var, as `fetchDeployedPage` already is in
   `check_tool_acceptance.go`) so tests inject statuses without a network. Cases:
   404 files exactly one item · 403/429/500/timeout file **nothing** · 200
   retracts · empty `src` files `empty_src` and is never probed · a relative src
   resolves against the page URL and not against the site root · two identical
   basenames under different directories produce two distinct `ItemKey`s · a
   `<script src>` mentioned inside a JS comment produces **no** reference
   (the regression test for my own misstep) · the cap logs what it dropped.
   **Then mutate the code to prove each guard is load-bearing** — delete the
   404-confirmation, invert the skip taxonomy, drop the URL from the ItemKey — and
   watch a distinct test fail for each. A guard no test can be made to fail
   against is not verified.
8. `gofmt`, `go build ./...`, `go test` the package, `scripts/pattern-check.py`.
9. Council gate submission (§below), then commit with `Council-Submitted:`.
10. Concept-register entry in the same commit that ships the seam (CLAUDE.md
    condition (2), still binding after the 2026-07-29 ruling retired condition
    (1)). Category: `docs026_concept_register/register/` — the discovery/check
    family file.
11. **Enablement is a SEPARATE, LATER step, and its ordering is not negotiable.**
    A check name the binary does not register **fails the step**. So: image first
    → pod-grep a symbol the change ADDS plus a negative control it does not →
    only then the SQL adding `asset_reference_404` to `design-discovery-agent`'s
    `checks` array. Until then the check is inert and the bug stays OPEN.
12. Docs: 016b §9 pattern, `WRONG_CALLS.md` row for the regex misstep, the two
    stale-citation corrections owed to `bugs_open/084` (it cites
    `check_tool_acceptance.go:375`; the real site is `:433-440`; and its §1 Tier-4
    claim is superseded by TL-033), `README_where_we_are.md`, and a `SUMMARY_`.

## Council submission sketch

**Rationale** — a subresource reference is the one asset class the platform checks
by presence only; the check that is named for loading does not load, and changing
it is RFC-scope, so the coverage has to arrive beside it rather than inside it.
The population Tier 4 cannot reach is most of what the fleet serves. Current live
findings: zero — this is a guard, and the submission says so rather than implying
a repair.

**Edits** (≤8): the new check file · its test file · the
`verifier_coverage_test.go` classification · the concept-register entry · the
enablement SQL (held, not applied) · the corrections to `bugs_open/084`.

**Consumers to TELL, not merely measure** (2026-07-29 ruling item 3): the
`experience_register` lane, because they were forced to re-type `feed_loads` to
Tier 4 for want of exactly this capability, and `design-discovery-agent`'s owners,
because their check array grows.

## Risks, and what I will not do

- **A guard with no live positive can rot unexercised.** Mitigated by the
  mutation testing in step 7 and by the induced fault, not by hoping.
- **Probe cost against 14 live domains.** Bounded to ~7 URLs per site today, hard
  capped at 40, deduped, and logged.
- **I will not touch `asset_loads`, `evaluateStaticCriteria`, or the
  confirming/refuting classification.** That is an RFC, and a veto on SCOPE is not
  answered by better measurements.
- **I will not enable the check in the same commit as the code.** That inverts the
  documented ordering and fails the step.
- **I will not claim the bug is closed while the check is inert.** `bugs_closed`'s
  bar is fixed AND live.
