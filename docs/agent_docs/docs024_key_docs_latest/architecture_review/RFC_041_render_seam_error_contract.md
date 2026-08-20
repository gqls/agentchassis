# RFC 041 — The component render seam now has an error contract, and it is consumed by four pipelines

**Raised:** 2026-08-19, by the `bugs_open/260` renderer-half lane.
**Trigger:** the council gate's `architecture` seat, round 1 of trail
`a44d9eb8-2e53-4447-8e02-0d36fb8889f4`, verdict REVISE:

> ARCHITECTURE_SIGNAL: needs_rfc … `RenderTemplateReportingMissing`/`RenderTemplate` changes
> from `string` to `(string, error)`, a new state-machine contract ("executed or errored, no
> third state") consumed across build, rerender, chrome-inject, chrome-store, section-editor and
> a second independent executor — 15+ call sites spanning `platform/orchestration/actions`,
> `platform/orchestration/datahelpers`, and `cmd/component-render-check`. … Recommend the RFC be
> written **in parallel, not as a gate** on this urgent fix.

**Status:** OPEN — written in parallel as the seat recommended. The change it describes is
committed and inert until a chassis roll. This document exists so the contract is decided
deliberately rather than inherited from a bug fix, which is the `bugs_closed/124` lesson.

---

## 1. What actually changed, in one paragraph a non-specialist can hold

Our page sections are HTML templates with instructions in them — "if there is a heading, print
it", "for each step in this list, print one of these". The instructions are meant to be carried
out and then vanish. When the template engine hit a problem, the code did not stop: it handed
the job to an older, simpler renderer that speaks a **different dialect**, which filled in the
individual words it recognised and left every instruction it did not understand sitting in the
page. Well-formed HTML, values resolved, directives intact, one line at Warn. The change deletes
that older renderer and gives the seam an error to return, so a render either happened or
visibly did not.

## 2. Why this is an architecture question and not a bug fix

Because the seam is shared, and because what changed is a **guarantee**, which is the 2026-07-29
§1 test: an addition to a shared mechanism needs architecture review when it changes what the
mechanism GUARANTEES, not merely because the mechanism is shared.

Before: *"`RenderTemplate` returns a string. It may be output that no template engine produced."*
After: *"`RenderTemplate` returns a string and an error. A non-nil error means nothing was
rendered; a nil error means `text/template` executed it."*

Four pipelines consume that guarantee — page build, page/section rerender, chrome (both the
inject path and the cached store path), and the section editor — plus an offline audit binary
and an acceptance gate for a component-conversion programme. The old guarantee was weak enough
that no caller could act on it; the new one is strong enough that **every caller must**, which
is exactly why the signature break was chosen over a `""` return. A `""` return would have been
a second silent shape: `assembleComponents` would stitch a page with a section missing, and
`GateConvertedTemplate` would gate an empty render.

## 3. The decisions this RFC records, so they are not re-litigated per call site

1. **The seam refuses; the caller decides.** The seam has no policy beyond "do not emit what you
   did not execute". Fail-the-step, carry-the-stored-bytes, refuse-the-store, refuse-the-edit and
   fall-back-to-plain-chrome are all correct, in different places, and the per-site table is in
   the concept register (STY-057) rather than here.
2. **The repair vehicle may not refuse.** `rerender_page_sections` carries the stored HTML and
   does not fail the page, because a re-render dispatched to fix a page must not refuse on the
   state it was dispatched to fix. Its escalation is the one the same action already uses for a
   missing required field — same helper, same `needs_page` key, same remedy.
3. **Chrome escalates only when there is nothing to serve.** An existing stored row keeps serving;
   a site with no stored chrome fails the build. This is a **deliberate behaviour change for
   greenfield provisioning** and is flagged as such.
4. **New authority ships opt-in with the unsafe default OFF** (owner ruling 2026-08-02 §2). The
   pre-render type gate can refuse content that renders fine today, so it is keyed
   `refuse_mistyped_llm_fields`, default OFF, zero live consumers at ship time. The *enricher*
   form of the same checker is unconditional, because it fires only on a render that already
   failed and can therefore refuse nothing.

## 4. The prior-art correction this RFC also records

The lane claimed `RenderTemplateWithMap` (`rerender_pages_actions.go`) was "a thirteenth seam
nobody had named". **That is false**, and the council's `prior_art_librarian` seat caught it:

- the concept register already names it — `page-build-pipeline.md`: *"`RenderTemplateWithMap`
  … is deliberately EXEMPT, named rather than silently skipped: one caller … its blast radius is
  a contact line, not a section artefact"*;
- the `bugs_open/238` lane enumerated it as one of eight unguarded call sites in its own council
  round;
- the `idea_uk_vm_site` lane's `bug_historian` seat FOUND it as a sibling silent-drop path and
  routed it through the same `<no value>` detector.

What **is** new is narrower and worth keeping: (a) its caller does
`html = contactInfoRe.ReplaceAllString(html, rendered)`, so an error there **deletes** the live
contact block rather than mangling it — which revises 238's stated exemption reasoning, since
"a contact line" understates deleting the block; and (b) its language DIVERGES from the other
seam's — no FuncMap, no `missingkey=zero`, so `{{safe}}`, `{{default}}` and `{{isset}}`, ordinary
everywhere else in this library, are PARSE errors there.

And a second correction, made by this lane against itself: the seam is **currently unreachable**.
`RenderTemplateWithMap ← rerenderInjectContactInfo ← rerenderSinglePage ← RerenderSitePagesAction`,
and `RerenderSitePagesAction` appears in **no entry of `GlobalActionRegistry`** (320 handlers,
checked 2026-08-19) — which is why the `idea_uk_vm_site` lane measured the function as absent
from the binary. The fix there is a trap disarmed before the path is revived, not damage being
stopped today. Anyone reviving that action inherits both (a) and (b).

## 5. The open question this RFC exists to have decided

**Should `RenderTemplate` — the one-line wrapper — exist at all?**

It now returns `(string, error)` and is used by nine call sites that mostly want "render this,
and if it cannot be rendered, tell me". The reporting form returns four values and is used by the
three call sites that also want the dead-control report. The 238 lane's council round already
recorded the adjacent question — *"making `RenderTemplate` itself the reporting form fleet-wide
changes the primitive every render flows through — the RFC-shaped move, not a rider on a bug
fix"* — and declined to answer it inside a bug fix. It is still unanswered, and this change makes
it cheaper to answer, because every call site now handles an error and the diff would be a
mechanical widening rather than a behaviour change.

**Secondary questions, each already stated as a landmine in STY-057:**

- the **absent-field sibling** (`missingkey=zero` renders an absent field as empty, silently) is
  untouched by this change and is covered at only two of fifteen call sites, by
  `missingRequiredLLMFields`, and only for fields marked both `source:"llm"` and `required`.
  **UNOWNED, and stated as unowned rather than left as an open-ended note** — the council's
  `bug_historian` and `architecture` seats both asked for an owner and a date in round 2, and this
  lane cannot assign either. What closing it would take, so the next reader is choosing rather
  than scoping: (a) the honest fix is at the SEAM, not at more call sites — have the render report
  which declared-required `source:"llm"` fields were absent, which it can already do
  (`missingBareFields` walks the template and tests the data map) and which would make the
  coverage question disappear rather than move; (b) the cheap fix is to call the existing presence
  gate at the other thirteen call sites, which is thirteen edits and leaves optional fields and
  schema-less components uncovered exactly as they are now; (c) doing nothing has a known cost —
  it is the mechanism behind `bugs_closed/004/005`' fleet-wide blanking, and this change's own
  measurement says **75 of 253 active components declare no schema at all**, so for those neither
  gate can ever fire. **A reader who takes this on should file it as its own bug and claim it**;
  it is deliberately not folded into 260, whose renderer half is now closed;
- **75 of 253 active components declare no schema at all**, so the type gate is silent for them
  by construction — the seam's error, not the gate, is the complete detector;
- `carried_render_failed` and `chrome_render_failed` are surfaced in an action result and a work
  item; neither feeds a discovery check yet.

## 6. What would have to be true for this to have been wrong

Stated so the decision is falsifiable rather than merely argued. The change is unsafe if any of
these turns out to be false; each was measured on 2026-08-19 with a control that could have come
out otherwise, and each is re-runnable:

| claim | how it was measured | what would falsify it |
|---|---|---|
| No live component template relies on the deleted dialect | 0 of 253 active components contain `{{#`, `{{nav_items_html}}`, `{{quick_links_html}}` | one that does — it would now hard-fail to parse |
| The fallback is never entered via a parse error | 0 of 251 active templates fail to Parse; controls: an unclosed `{{if}}` must fail, a valid nested one must pass | a template that parses only under the fallback |
| Nothing currently working depends on the fallback | 0 of 1,778 stored sections fail to Execute against their own `content_data`; controls: the bug's own A/B pair | a stored section that renders only via the fallback |
| Brace-bearing tool-page copy is unaffected | the design contains no output brace-scan; a positive-control test renders `{{ variable }}` in copy and must pass | any brace heuristic added later |

**Sources:** `bugs_open/260`; `docs/agent_docs/docs024_key_docs_latest/bugfix_260_render_fallback/`
(PLAN, NOTES, probes, submission); concept register `STY-057`; council trail
`a44d9eb8-2e53-4447-8e02-0d36fb8889f4`.
