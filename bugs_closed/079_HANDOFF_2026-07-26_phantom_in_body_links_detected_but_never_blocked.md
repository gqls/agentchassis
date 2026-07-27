# Handoff — the platform detects a dead in-body link and deploys the page anyway

> ## FIXED — live in chassis `v1.0.1171`, 2026-07-26
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
