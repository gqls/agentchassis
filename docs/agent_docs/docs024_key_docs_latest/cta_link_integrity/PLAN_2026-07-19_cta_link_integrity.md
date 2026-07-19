# PLAN — CTA / link integrity: every button has a destination, or it isn't a button

**Date:** 2026-07-19 · **Branch:** `085_debug_and_feature_loops` · **Status:** planned, not started
**Bug:** `bugs_open/023_HANDOFF_2026-07-19_cta_label_url_pairing_unchecked.md`
**Evidence:** `NOTES_cta_link_integrity.md` (read this first — every claim below is grounded there)
**Owner ask:** a build-time check for all sites, plus a checker *and handler* for improvement-loop time.

---

## 1. The problem in one paragraph

A component can declare a button's **label** and its **URL** as two unrelated schema
fields. The label is typically `source:static`, which re-applies its fallback on every
render and bypasses `required`/`on_missing` entirely. The URL may be absent from the
schema, unresolvable, empty, or authored by an LLM with nothing to look it up against.
Nothing in the platform expresses "a label implies a destination" as a constraint, and the
template renders the anchor ungated. The result is a button that always has text and
sometimes has nowhere to go. Four such buttons shipped on leopardess; **51 dead or suspect
controls exist across 7 of 11 sites.** Separately, when a check *does* catch one, the
finding is filed at `needs_human_review`, which nothing consumes — one of these four was
correctly diagnosed on 2026-07-17 and sat unread until the owner clicked it two days later.

## 2. Defect classes

Each is independently sufficient to ship a broken button. A fix that addresses only one
leaves the others live.

| # | Class | Evidence | Blast radius |
|---|---|---|---|
| **A** | **Label/URL pairing is never a constraint.** `*_label` static + `*_url` absent/empty. `check_required_fields_missing.go:189-192` skips any field whose `source != "llm"` — and CTA url fields are `renderer`/`site_specs` by design, so they are *categorically exempt* from the only content_data required-field check. | *See How It Works*, *Visit the Tool* | fleet |
| **B** | **Three link scopes are exempt from every check.** `href=""` → warning only (`validate_page_content.go:551-560`, `:257`). `href="#frag"` → `LinkScopeAnchor`, skipped by phantom/misdirected/validate. External → skipped by all; **zero HTTP checks exist anywhere in `platform/`**. | all four | fleet |
| **C** | **Templates render CTA anchors ungated.** `<a href="{{.cta_secondary_url}}">` with no `{{if}}` turns an empty value into `href=""`. Violates the platform's own stated invariant LNK-005 ("an unresolvable destination renders nothing"). | *See How It Works* | fleet |
| **D** | **Hardcoded template fragment with no target.** `tool-guide-intro` hardcodes `href="#guide-start"`; no such id exists on any page using it. | *Start the Guide* | 4 pages / 3 sites |
| **E** | **`source:llm` + `required:true` on a URL manufactures fabrication.** The model has no page set to resolve against and must return something. It invented two different hostnames on adjacent pages. | *Visit the Tool* | fleet |
| **F** | **Static fallbacks carry another tool's vocabulary.** `bayesian-ranking-hero-tool_pre_037` freezes "Start Ranking Free" / "Calculate Rankings" / "Try the Bayesian Ranker" onto whatever page adopts it. Content_data cannot override a static source. Compounded by the live row being an **`is_active` backup snapshot**. | *Start Ranking Free* | 3 pages / 2 sites |
| **G** | **Detection exists; nothing consumes it.** `unresolved_cta`, `cta_names_unknown_destination`, `dead_control` all land `needs_human_review` → never triaged, no handler, excluded from re-open queries. Grep of `platform/`: emission sites only, **zero consumers**. 34 such items open on leopardess. | the detected one | fleet |
| **H** | **Repair scope < detection scope.** `ctaFieldNames` (`resolve_internal_links_action.go:91-98`) is a hardcoded 6-component map. Everything outside it is, in the code's own words, *"detectable but not repairable"*. Both offending components are outside it. | both components | fleet |

## 3. The single highest-leverage change

**Derive CTA field pairs from `input_schema` instead of the hardcoded `ctaFieldNames` map.**

Any field named `*_url` with a sibling `*_label` / `*_text` is a CTA pair, discoverable
generically. This one change:

- kills class **H** permanently (no component is ever outside the map again — there is no map);
- makes class **A** *checkable*, because the pairing finally exists as data;
- lets the existing repair machinery (`applyCTARecompute`) reach every component;
- removes a file that has been hand-edited by migrations 091, 096, 097b and 098 — four
  migrations of the same lesson, which is the drift class the council reviews for.

Everything else in this plan is cheaper and less valuable than this. Do it first.

## 4. Phasing

### Phase 1 — Make the pairing exist and be checked (build time)

**P1.1** Replace `ctaFieldNames` with schema-derived pairing (§3). Keep the map as an
override for the handful of components whose fields don't follow the convention.

**P1.2** New build-time validation in `validate_page_content.go`: for every derived CTA
pair, if the **label resolves non-empty** and the **URL resolves empty/absent** → finding
`cta_without_destination`. This is class **A**, caught before render.

**P1.3** Promote `empty_internal_href` from warning to blocker — **but staged** (see §5).

**P1.4** New check: fragment hrefs (`#foo`) resolved against the ids present in the
**assembled** page. Note this cannot live in per-component validation — a component cannot
see whether a sibling component defines the id. It belongs at assemble time. This is class
**D**, and it is the one place the plan needs new plumbing rather than a new rule.

**P1.5 — the email→hostname transform (deterministic, cheap, do it early).**
`leopardess.contactforsales.com` was not a wild guess: the site's identity spec holds the
real contact email `leopardess@contactforsales.com`, and the model turned it into a hostname
by swapping `@` for `.`. **A hostname equal to a known contact address with `@`→`.` is
fabricated by construction** — that is a string identity against data the platform already
holds, needing no network call and no heuristic. Check every external URL's host against the
site's own contact emails (and, cheaply, the reverse: any host that is a `<local>.<domain>`
recomposition of any current `site_specs` identity email, fleet-wide).

> **Exposure: 6 sites share `contactforsales.com` as their contact domain**
> (`agents@`, `finetuning@`, `gas@`, `idea.uk@`, `leopardess@`), each in its *current*
> identity spec. Any of them can produce this exact fabrication. See NOTES for the table.

Also worth a cheap sibling rule: **an external URL whose host is a different-TLD variant of
the site's own registrable domain** (`example.com` on an `example.co.uk` site) is
near-always either a mistake or an unintended off-site exit. Today
`checkDomainContamination` (`validate_page_content.go:481-534`) cannot see this — it
substring-matches a hardcoded 5-domain list and never compares against `expectedDomain`.

### Phase 2 — Remove the class structurally (templates + schemas)

**P2.1** Gate every CTA anchor fleet-wide: `{{if .x_url}}<a href="{{.x_url}}">…{{end}}`.
The precedent already exists — `info-card-grid` was gated exactly this way *on leopardess*
and never generalised. This makes LNK-005 true by construction and retires class **C**.
Mechanical, high-volume, ideal for a workflow fan-out.

> **Sizing (measured 2026-07-19, RUNBOOK R9): 75 ungated CTA anchors across 38 active
> components, versus 14 gated across 12.** About **84% of URL-bound CTA anchors in the
> library violate LNK-005**. The invariant is documented, agreed, and almost universally
> unenforced — which is why this keeps recurring and why P1 alone (detection) would just
> generate a large backlog. Figure is from a lookback heuristic; re-derive with a real
> template parse before mass-editing.

**P2.2** `tool-guide-intro`: remove the hardcoded `#guide-start` (or emit the id). Class **D**.

**P2.3** `tool-guide-intro.cta_secondary_url`: change `source:llm, required:true` →
`source:renderer` resolved from the real page set, or optional + gated. **No URL field
should ever be `source:llm` + `required` with no fallback.** Add that as a schema-lint rule
so it cannot recur. Class **E**.

**P2.4** Audit `source:static` label fallbacks for cross-tool vocabulary. Class **F**.

> **Measured 2026-07-19 (RUNBOOK R10): 16 active `_pre_037` rows**, including
> `blog-listing`, four `header-*` variants, `footer-with-disclaimer` and five `tool-*`
> components. **None of them collides with a canonical sibling** — each is the *sole*
> active row for its function. So they are not stale duplicates shadowing a good row; they
> *are* the live component, named as though they were a backup. Migration 037's
> replacements never landed, and that is how the frozen Bayesian labels survived.
> **Do not delete them — that deletes the live component.** Rename/supersede deliberately,
> and treat the content of all 16 as un-reviewed pre-migration material.

### Phase 3 — Improvement-loop checker **and handler**

**P3.1 (checker)** Extend `check_dead_controls.go` to cover `href=""` (currently
explicitly ceded to the phantom check, so it falls between two stools) and unresolvable
fragments.

**P3.2 (checker)** New post-hoc check `external_link_unreachable`: HTTP HEAD with caching,
backoff and an allowlist. **Post-hoc only — never at build time** (network flakiness must
not block a deploy). This is the only mechanism that catches class **E** after the fact;
it would have caught both fabricated hostnames.

**P3.3 (handler) — the actual gap.** Findings must stop dying at `needs_human_review`.
Two lanes:
- **Auto-repairable** — a real, non-excluded, non-circular destination exists → repair via
  the (now schema-derived) recompute path, same keep-rule as today.
- **Not auto-repairable** — no honest destination exists. **The correct fix is to drop the
  button, not to point it at `/contact.html`.** Pointing a promise-shaped label at contact
  is what produced *Start Ranking Free → /contact.html* in the first place. Dropping it
  satisfies LNK-005 and is safe to automate.
- Anything genuinely undecidable (a label promising a product that doesn't exist) is a
  product decision — but it must arrive somewhere a human actually reads, not a queue with
  34 items and no reader.

#### Phase 4 — leopardess remediation

Fix the four buttons. **Do it at component/schema level (Phase 2), not in
`page_components`** — `bugs_open/001` means anything written to `page_components` has an
undefined shelf life on this site, while template and schema fixes survive a re-plan.

**P4.1 — owner action, approved 2026-07-19: redirect `leopardessconsulting.com` →
`leopardessconsulting.co.uk`.** The `.com` is owner-owned and currently serves a 114-byte
blank page for every path, and a live button points at it. A 301 at the apex (with path
preserved, so `/tools/llm-cost-calculator` lands on the real page) turns *Visit the Tool* on
the LLM-cost page into a working button immediately, independently of every code fix above.

Do it as a Cloudflare redirect rule alongside the existing `.co.uk` setup, not as a new
origin. **This is worth doing regardless of the button** — an owned `.com` next to a live
`.co.uk` brand should not serve a blank page to anyone who guesses it, and search engines
and typed traffic reach it too.

> Note the ordering interaction: P4.1 makes one fabricated URL *resolve*. That is a genuine
> improvement for visitors, but it must **not** be mistaken for the defect being fixed — the
> field is still `source:llm, required:true` and will invent a different hostname on the next
> build. **Do P2.3 as well**, and do not let a green link lull the fix. (The other page's
> button, `leopardess.contactforsales.com`, is unaffected by the redirect and stays dead.)

## 5. Staging — do not turn the blocker on cold

30 empty hrefs exist across 7 sites **today**. Flipping `empty_internal_href` to blocker in
one step fails the next rebuild of most of the fleet. The repo already has the right
pattern for this (LNK-009, phantom_internal_links): **ship as warning + work item, measure
the true count, fix the backlog, then flip.** Sequence: P1.2 as warning → drain → flip to
blocker. A gate that fires on everything gets disabled, and then nothing is checked at all.

## 6. On invoking the experience loop

The owner asked whether the experience loop is needed. **Partly — but not as the fix, and
not yet.**

- The experience loop is the right *home* for P3.1/P3.2: `dead_controls` was built there
  (`f2824a713`, migration 165, live in v1.0.1134) and interaction-level checking is its
  remit.
- But it is a **detection** loop, and detection is not what is missing here. Running it
  against leopardess today would re-discover buttons that are *already correctly
  described* in 34 open work items and add more rows to a queue nothing drains.
- **Build the handler (P3.3) first.** Then the experience loop's output becomes actionable
  rather than accumulating. Running it before then converts a visible problem into a
  larger invisible one.

## 7. What this plan deliberately does not do

- **No HTTP checking at build time.** Network calls in a deploy gate trade one flake class
  for another.
- **No rewriting of labels by an LLM.** Class F is fixed by retiring a bad component and
  un-freezing static fields, not by asking a model to re-author copy — that reintroduces
  the fabrication surface (090/096 landmine).
- **No auto-retargeting to `/contact.html`.** That heuristic is the origin of the bug.

## 8. Appendix — citation map

| Claim | Source |
|---|---|
| static/renderer bypass `required`/`on_missing`, re-apply every render | `plan_sections_action.go:1210-1218`; `sourceResolver.resolve:400-402,414-416` |
| warnings never block; `empty_internal_href` is a warning | `validate_page_content.go:257`, `:551-560`, `:562-566` |
| anchor + external scopes skipped | `validate_page_content.go:577-578`; `check_phantom_internal_links.go:279`; `check_misdirected_cta.go:198-201` |
| `dead_controls` cedes `href=""`; matches only 6 literals | `check_dead_controls.go:58-91`; `datahelpers/links.go:107,109-116` |
| domain contamination = hardcoded 5-site substring list | `validate_page_content.go:481-534` |
| required-fields check skips non-llm sources | `check_required_fields_missing.go:189-192` |
| `ctaFieldNames` 6-component map; outside = not repairable | `resolve_internal_links_action.go:83-86, 91-98` |
| recompute keep-rule, gated on `cta_links_stale` | `rerender_page_sections_action.go:264, 371-390` |
| `needs_human_review` never triaged / excluded from re-open | `triage_detect_items_action.go:91-104`; `load_work_item_actions.go:804` |
| checks are per-agent config, no `site_discovery_checks` table | `discovery_checks.go:72-78`; `092_enable_link_checks_and_cta_rerender.sql:36-38`; `165_enable_dead_controls_check.sql:43` |
| LNK-005 invariant; link-integrity loop scope + residue | `docs024_key_docs_latest/register/link-management.md:36-39`; `social001_vonc_tiktok_social/minilobby_task/HANDOFF_link_integrity_arena_2026-07-16.md` |
