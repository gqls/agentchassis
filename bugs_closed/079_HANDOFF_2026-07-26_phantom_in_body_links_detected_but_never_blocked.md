# Handoff — the platform detects a dead in-body link and deploys the page anyway

> ## CLOSED 2026-07-29 — the repair now PERSISTS, proven on a natural production run
>
> Fixed by moving the repair to the persistence point: `repairSectionsBeforePersist` inside
> `SavePageSectionsAction` (`save_sections_link_repair.go`, commits `5083124e3` + `fd87d6194`).
> Council `7c24776e-07f8-4c2e-b1b6-ad3e73c6023c`: round 1 REVISE, round 2 **APPROVED** (11
> reviewers, 0 unreadable). LIVE on chassis **v1.0.1196**, both replicas pod-grepped for the
> compiled marker (`marker=1`, positive control `1`, v1.0.1194 baseline `0`).
>
> **The proof is a NATURAL run, not an induced one — vetcomparison.uk `index`, 2026-07-29
> 01:58:04Z.** It is the exact inverse of this bug's own evidence, measured the same way:
>
> | | the bug (fundamentallyai, 07-28) | the fix (vetcomparison, 07-29) |
> |---|---|---|
> | repair logged | 10:45:01.347Z, 10 repairs | 01:58:04.382Z, 5 unlinks |
> | components saved | 10:45:01.768–.807Z (+400ms) | 01:58:04.427–.451Z (+45ms) |
> | hrefs in the saved rows | **still there, all 9** | **gone, all 5** |
> | served page | all 9 × 404 | all 5 absent |
>
> The five were `/search`, `/about-pricing`, `/about-ownership-disclosure`,
> `/guides/pet-owner-rights`, `/claim-listing` — re-probed live and all genuinely 404, so they
> were real phantoms, not a false positive. **Attribution is exact:** the row carries
> `action='save_page_sections'`, and that value was written **0 times before the roll and 1
> after**, while `rerender_page` (20 before / 1 after) is the positive control proving the
> query itself works. Only the new call site can have written it.
>
> **WHAT WAS NOT PROVEN — the induced repro FAILED, and it is worth knowing why.** The
> pre-registered zero-LLM test (gamesdesign `bayesian-ranking`, stored `href=""` since 07-21)
> did **not** exercise the fix. `rerender_sections` re-renders each section from
> `content_data` through the CURRENT template, and that page's `content_data` has
> `cta_primary_label`/`cta_secondary_label` but **no url fields at all** — so the template's
> skip gate (LNK-006) dropped the buttons entirely and left an empty `brht-cta-row`. There was
> nothing for the repair to act on; no `save_page_sections` repair row was written for that
> run. The 2 unlinks logged against that page at 07:31:46Z came from the **outbound** rerender
> seam (`action='rerender_page'`, LNK-023) acting on the assembled page's chrome — a different
> call site. **A repro that is re-rendered from `content_data` can be destroyed by the render
> itself; check the template gate before trusting one.** (Side effect, recorded: that rerender
> removed two dead CTA buttons from the live gamesdesign page. Correct per the
> correct-or-absent principle, but it was a change to a live page.)
>
> **STILL OPEN, and deliberately not closed here:**
> - **`bugs_open/136_…section_editor_and_three_siblings…`** — `save_page_sections` is only 1 of
>   **10** Go writers of `page_components.rendered_html`; three others persist LLM prose with no
>   repair at all. Found by this fix's own council review. The claim "no build path can persist
>   an unrepaired section" is FALSE; the true claim is "no `save_page_sections` invocation".
>   (NB a different `bugs_open/136` exists — resolve by slug, per CLAUDE.md.)
> - **`bugs_open/092`** — the upstream cause. vetcomparison's five phantoms and gamesdesign's
>   missing url fields are both 092: the writer never receives its link constraints, so it
>   invents targets or omits them. This fix makes the symptom un-shippable; it does not stop
>   the writer producing it.
>
> Registered **LNK-024**. First `doc_notes` row for `subject_type='action'`,
> `subject_key='save_page_sections'`.

> ## REOPENED 2026-07-28 — the repair runs, and its output is DISCARDED before persistence
>
> Moved back from `bugs_closed/` by the brochure_component_library thread. The FIXED banner
> below is preserved unedited; everything in it about the *repair action* remains true. What
> is false is one clause: *"`clean_html` — the string `save_sections` persists"*. It is not.
>
> **Mechanism, cited.** `validate_page_content` applies `RepairPageLinks` to `cleanHTML`
> and returns it as `clean_html` (`validate_page_content.go:357`, comment at :352–354).
> The next step, `save_sections` (`save_page_sections_action.go`), tries the **structured
> metadata path first**: if `sections_metadata_field` resolves to a non-empty array it
> persists those per-section `rendered_html` strings and never reads `html_field`
> (`save_page_sections_action.go:166–188`); `html_field` — configured as
> `validation_result.clean_html` — is read **only when metadata yields zero sections**
> (:192 onward). The live page-build plan (read from orchestration
> `bcdc6455-c279-4646-85d8-bfa8cbea3f2e`, page-build-handler, 2026-07-28) sets BOTH
> `sections_metadata_field: page_content.response.sections_metadata` AND
> `require_sections_metadata: true` on the validate step — metadata is therefore always
> present, the fallback is unreachable, and **the repaired string is structurally discarded
> on every page build**. Not a race; a dead branch.
>
> **Production evidence, same day, two sites:**
>
> - `fundamentallyai.com/capabilities.html` (work item `8f366ce5`, build orchestration
>   `bcdc6455`): `CONTENT_LINK_REPAIR_DETAIL` written **10:45:01.347Z** — 10 repairs listed,
>   naming exactly the 9 dead targets (8 unlinked, `/contact` → `/contact.html` rewritten
>   twice). The page's six `page_components` rows were saved **10:45:01.768–.807Z** — and
>   `hero-card-carousel` + `info-card-grid` carry `href="/contact"`,
>   `/capabilities/review-council` and the rest, unrepaired. Live crawl ~14:20Z: **all 9
>   serve 404 from the deployed page.**
> - `vonc.com/about.html`: repair row **14:28:53Z** ("unlink `/how-it-works`" ×2); the
>   `platform-comparison` `page_components` rows saved in the same run window still contain
>   `href="/how-it-works"`.
>
> **Why the closure's proof did not catch this.** The end-to-end induction ran through
> `content-reviewer` with `html_field` repointed at `input_data.page_html` — a route that
> has **no `save_sections` step**. It proved the action transforms and logs correctly
> (which it does); nothing on that route persisted anything. The unit tests proved the
> helper; the pod-greps proved deployment. No check ever read back what a real build
> **saved**. Writes-the-field ≠ reads-the-field, on our own fix.
>
> **Blast radius beyond links.** Any transformation applied only to `clean_html` cannot
> reach persistence on the structured path — the gate's comment-stripping is equally
> unreachable **[INFERRED from code; no observed comment damage — saved components checked
> 07-28 carry none, but the writer emitted none either]**. Also note the same build
> invented four `<img src="/assets/illustrations/*.svg">` paths that shipped: image srcs
> were never in `RepairPageLinks`' remit at all (see `071`'s 07-28 entry).
>
> **Fix candidates, ordered by what closes the door:**
> 1. **Repair where persistence happens** — `save_page_sections` applies `RepairPageLinks`
>    to each section's `rendered_html` before insert (it already holds `params.DB` and
>    `site_id` to build the `PageURLIndex`). No build path can then save an unrepaired
>    section, whatever the workflow config says.
> 2. `validate_content` repairs `sections_metadata` in place alongside `clean_html`, so
>    both representations leave the gate consistent.
> 3. Make the structured path derive from repaired `clean_html` (re-split by section
>    boundaries) — highest fidelity risk, listed for completeness.
>
> **How to verify the real fix**: after a natural build with a dead in-body link, SELECT
> the saved `page_components.rendered_html` for the repaired href — it must be absent —
> then crawl the deployed page. The persisted row is the artefact; the action's return map
> is not.
>
> **Diagnosis-loop note:** a 090 verification run should be fired once the Anthropic lanes
> return (spend cap exhausted until 08-01; the queue currently fails LLM work). The claim
> above is cited from code + two same-day production runs, but the loop's re-check is owed
> on a reopen of this size.
>
> **Verification run DONE 2026-07-28 evening** (the cap was raised ~14:50 BST, not 08-01;
> corr `954d8da9-789a-4515-be07-1b15b9511f4b`, 5 iterations, verdict in the child
> orchestration's `collected_data->'verdict'` — NOT in `diagnosis_artifacts`, which got
> bundles only). **Outcome: UNVERIFIABLE — no refutation.** The loop confirmed the static
> mechanism chain (validate_content `next_step: save_sections`; the repair-log format
> string) but stated it could not observe "a page-build-handler/validate_content run with
> valid=true" inside its search scope, so the discard remained "unobserved" from where it
> stood — it did not find the 10:45Z fundamentallyai timeline this file cites. Chasing its
> one NEW citation produced a **third corroborating site**: gamesdesign.co.uk
> `/tools/bayesian-ranking.html` — repair computed TODAY 15:47:03Z (2 empty-href unlinks,
> `CONTENT_LINK_REPAIR_DETAIL`) while every `page_components` row for that page was last
> written **2026-07-21** and `bayesian-ranking-hero-tool` still carries `href=""`. The gate
> recomputes the same repair on every pass; the store never changes. No counter-example
> found anywhere: no run exists whose computed repair reached its saved row.
>
> **IMPLEMENTATION HANDOFF for candidate 1 (designed 2026-07-28 evening, NOT yet built):**
> `docs/agent_docs/docs024_key_docs_latest/bugfix_079_phantom_link_gate/HANDOFF_2026-07-28_platform_fix_candidate1.md`
> — insertion point, the existing helpers to reuse (`loadValidPagePaths` is in the SAME
> package as the save action), fail-open + reversal-lever design, test plan, and a
> zero-LLM live verification via a gamesdesign rerender. Start there, not from scratch.
>
> **Live-reproduction note for the fixing thread:** fundamentallyai `capabilities.html` is
> NO LONGER a live reproduction — the brochure thread reverted its two corrupted components
> to pre-regression `content_data` (corrupted state archived in `page_component_history`,
> source `operator_restore_pre_regression_2026-07-28`) and re-rendered; the served page is
> phantom-free. Use vonc `/about.html` and gamesdesign `/tools/bayesian-ranking.html`
> (the stored `href=""`, above) as the standing reproductions.

> ## FIXED — live in chassis `v1.0.1171`, 2026-07-26 — **superseded by the REOPENED banner above; preserved unedited below**
>
> **The fix.** The gate no longer only reports a dead in-body link, it repairs it, in
> `clean_html` — the string `save_sections` persists, which is why the gate is the only step on
> the build path where this can work. A target that resolves at its `.html` form has its href
> rewritten to the **stored** `pages.url` (never a constructed one — `bugs_closed/029`'s rule);
> anything else has its `<a>` dropped and its inner markup kept. Page ships, prose survives, 404
> dies. New file `platform/orchestration/datahelpers/link_repair.go`; gate changes in
> `validate_page_content.go`. Commits `43f254be5` + `31d8ac7dc`.
>
> **Candidate chosen: 3 (strip), with a rewrite arm in front of it.** Candidates 1 and 2
> (promote severity) were ruled out **by the measurement this file demanded** — see "The census"
> below. Candidate 4 (work item) was ruled out by `bugs_open/083` (this item type detected 22
> times, fixed **zero** times ever) and `bugs_open/077`.
>
> **PROVEN:**
> - Unit: 13 cases in `link_repair_test.go`, including four asserting the pass changes *nothing*.
>   Proven non-vacuous by inducing a fault — 8 of 8 repair tests fail with `RepairPageLinks`
>   stubbed, the four no-change tests correctly still pass.
> - Deployed: pod-grep on the running `v1.0.1171` chassis, against a baseline taken before the
>   roll. `CONTENT_LINK_REPAIR_DETAIL` 0 → **1**, `repair_internal_links` 0 → **1**,
>   `link check and repair SKIPPED` 0 → **3**; positive control `CONTENT_VALIDATION_BLOCKER_DETAIL`
>   **2 both times**, so the grep discriminates.
>
> **PROVEN END TO END on the deployed binary, 2026-07-27 12:23Z** (owner-approved induction).
> Crafted markup fed through the live gate on `dartsonline.com`, exercising both arms plus the
> two must-not-touch cases:
>
> ```
> in:  <p>Read our <a href="/definitely-not-a-page.html">pricing guide</a> for details, then
>      <a href="/contact">talk to us</a> or browse <a href="/new-arrivals.html">new arrivals</a>.
>      See <a href="https://example.com/x">an external ref</a>.</p>
>
> out: <p>Read our pricing guide for details, then
>      <a href="/contact.html">talk to us</a> or browse <a href="/new-arrivals.html">new arrivals</a>.
>      See <a href="https://example.com/x">an external ref</a>.</p>
> ```
>
> `checked_links 3 · links_rewritten 1 · links_unlinked 1`, and the repair list carried
> `{"href":"/definitely-not-a-page.html","action":"unlink"}` and
> `{"href":"/contact","action":"rewrite","new_href":"/contact.html"}`. The invented target lost
> its link and **kept its text**; the extension-less target was rewritten to the **stored**
> `pages.url`; the already-valid link and the external link were untouched.
>
> The durable record landed too — `agent_error_log`, `error_code=CONTENT_LINK_REPAIR_DETAIL`:
> *"Repaired 2 dead internal link(s) before save: 1 href(s) rewritten, 1 link(s) removed"*, with
> both hrefs in `context.repairs`. That is `071` gap 3 answered for this class: the finding now
> outlives `collected_data` pruning.
>
> **How it was induced, and why that took four attempts.** No natural induction was available —
> the gate ran on 5 real builds under the new binary and every one had `checked_links: 0` (the
> writer is emitting no anchors at all, which is `bugs_open/092`). Crafted routes fail unless the
> target agent type has a **live pod**: an inline `config.workflow` and a freshly-seeded agent
> type both vanished without an orchestration row. What worked was `content-reviewer` (live pod)
> with its `validate_content` step's `html_field` pointed at `input_data.page_html` for ~2
> minutes — an owner-approved change to a live agent definition, restored immediately and
> verified byte-identical against the pre-patch snapshot. Risk was ~nil: content-reviewer had run
> **once in 24h** and that run was also mine.
>
> **No `Council-Reviewed:` trailer.** Submission `97904892-5c09-4782-aeda-37dd944abdfc` never got
> an orchestration row (1h40m; the lane had a run stuck at `council_decide` for 239 min). The
> commits now pre-date any possible verdict, so a trailer is permanently impossible here and the
> `098` report will list them as unreviewed — by design, not by oversight.
>
> **The upstream cause is `bugs_open/092`** — the writer is never handed the site's real page
> list (`page_count: 0` on 20 of 20 runs), which is why all 15 phantoms in the census were pure
> inventions rather than typos. This fix is the backstop; 092 is the prevention.

**Filed 2026-07-26** while closing `bugs_open/029` (tool-suggester phantom tool links).
Filed because the council gate raised it as a **high-severity gating objection** against 029's
fix, and it is right that it is a separate defect: 029 closed one *source* of phantom links; this
is the reason a phantom link *ships* whatever the source.

> **bug_historian, council round 1 on submission `745f9dfd`:** *"Fixing tool-suggester's URL
> fabrication closes one source, but any OTHER future source of a phantom href (manual edit, a
> different agent, a different cross-link mechanism, a regression in the new emitter itself) will
> still deploy silently with only a non-blocking warning — the platform's own root guard against
> shipping dead links stays generic and exploitable."*

## Mechanism

`platform/orchestration/actions/validate_page_content.go`:

- `validateInternalLinks` (`:540-582`) **does** the hard part already: it extracts every in-body
  `href` from the generated content and resolves it against the site's real `pages.url` set.
- A miss is filed as `Severity: "warning"` (`:571`).
- The overall verdict is `valid := blockerCount == 0 && errorCount == 0` (`:257`) — **warnings do
  not count**. The page proceeds to `save_sections` → `update_status` → `deploy_page`.

So the platform sees the dead link, writes down that it saw it, and publishes it. The comment
justifying warning-severity is that "the improvement loop resolves it" — the same premise that
`023`, `033` and `049` each falsified in their own area. Nothing drains that finding.

## Why this is worth a bug of its own rather than a note on 029

029's emitter is now incapable of fabricating a URL, and its items carry a real `pages.url`. But:

1. **The writer is an LLM.** `page-content-writer` is free to emit any `<a href>` it likes in
   prose; 029 only fixed what the *instruction* says. A model that invents `/pricing.html` on a
   site with no pricing page produces exactly the same 404, from a path this fix does not touch.
2. **Other emitters exist and will be written.** Any future feature that suggests a link goes
   through this same guard.
3. **A regression in 029's own fix would be invisible.** The gate that should catch it is the one
   described here.

## Evidence

- Code as above (`validate_page_content.go:257`, `:540-582`, `:571`).
- Live damage that got past it: 9 dead `/tools/*.html` references on 8 **deployed** pages across 3
  sites — the census in `bugs_open/029` §"The existing damage, measured 2026-07-26". Every one of
  those pages was validated before deploy.
- `resolve_internal_links_action.go:98-105` resolves only structured **CTA fields**, so it is not
  a second line of defence for prose hrefs — it never sees them.

## The census this file demanded, run before choosing (2026-07-26)

`orchestration_states` retains **13 days**, not the ~24h assumed here and in `071` — so the
measurement was possible after all.

| | |
|---|---|
| builds carrying a `validation_result` | 16 |
| builds carrying phantom-link findings | **3 (19%)** |
| phantom instances / unique targets | 17 / 15 |
| of those 15, targets existing in ANY form (incl. `+.html`) | **0 — all pure inventions** |
| builds that deployed anyway | 3 of 3 (`valid:true`) |
| pages | oufe.com `/index.html`, webdesign.co.uk `/index.html` — **both homepages** |

This is what killed candidate 1. On `!valid` the action returns `(nil, error)`, routing to
`mark_needs_review` → `fail_work_item`: the page never saves and never deploys. Promoting the
severity would have stopped **two homepages** from shipping — "no page at all" instead of "a page
with one bad link", exactly the outcome this file warned was worse.

## Fix candidates (none implemented — this is a filing, not a fix)

1. **Promote in-body phantom links to `error` severity.** One-line change, fleet-wide behaviour
   change: pages with a dead in-body link stop deploying. **Do not do this without measuring
   first** — run the census across all pending content, because if the current rate is
   non-trivial this converts "a page with one bad link" into "no page at all", which is worse.
2. **Promote to `error` only when the href resolves to nothing AND the page is on the `.html`
   fleet** (extension-less targets are `049`'s known false-positive class).
3. **Strip the dead href instead of blocking** — deploy the page with the anchor text intact and
   the link removed, and file the finding. Keeps the content, kills the 404. Probably the right
   answer, and it needs a place to put the finding that somebody drains.
4. **Leave severity alone and make the finding actionable** — a work item with a handler that can
   repair or strip the link. Note `bugs_open/077`'s trap: do not file items whose handler has no
   remit to fix them.

Read `023` (CTA label/url pairing), `033` (the review queue nobody drains) and `049` (312 live
broken links) before choosing — this is the same family, and 049 is **owned by the
`cta_link_integrity` workstream**, so coordinate rather than compete (`scripts/who-owns.py 049`).

## How to verify a fix

1. Induce it: give a page's content a `<a href="/definitely-not-a-page.html">` and run the build.
   The page must not deploy with that link live — whichever candidate is chosen.
2. Confirm the finding is durable and countable afterwards (a log line is not a record).
3. Re-run the 029 census query: no new dead `/tools/*.html` refs on deployed pages.
