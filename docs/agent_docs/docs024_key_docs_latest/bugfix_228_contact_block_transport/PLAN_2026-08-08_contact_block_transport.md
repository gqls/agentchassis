# PLAN 2026-08-08 — contact-block transport (bugs_open/228)

Workstream: `bugfix_228_contact_block_transport`. Bug:
`bugs_open/228_HANDOFF_2026-08-08_contact_block_tells_the_visitor_the_message_was_sent_and_sends_nothing.md`.
`scripts/who-owns.py` shows no owning workstream for the bug; the component itself
was built by the `staged_component_build` lane, which filed the bug while writing
the component's acceptance fence — coordination notes go into the bug file, and
the fence pointer in §f below is addressed to that lane. Re-verified live today
(fresh curl of `robot-hands.com/contact.html` and of the served
`/tools/assets/contact-block.js`; `git log --since="3 hours ago" --all` shows no
other session has touched contact-block since filing).

## a) What is broken, and the cause one level below the bug file

`contact-block` (one shared row in `content_components`, `function =
'contact-block'`, added 2026-08-08) renders a contact form with genuinely good
client-side validation — and then fabricates the send. On a valid submit the JS
disables the button, waits a 1200 ms `setTimeout` whose only purpose is to look
like a network round-trip, prints "Your message has been sent", and
`form.reset()`s the visitor's text away. There is no `fetch`, no
`XMLHttpRequest`, no `sendBeacon`, no `form.submit()` anywhere in the ~2,100-byte
asset, and the `<form>` tag has no `action` and no `method` at all. The message
is discarded, and the visitor is told otherwise. The good validation is what
makes it dangerous: a mistyped email gets a correct, specific error, so the form
reads as wired-up from the outside.

**Live blast radius: TWO pages, not the three the placement rows claim.** The bug
file corrected itself same-day: `finetuning.uk/case-studies.html` has an active
`page_components` row for contact-block but the served HTML contains no
contact-block markup at all — a drift row, one of five that lane has now found.
The live pages are `robot-hands.com/contact.html` (the site's actual contact
page, reached from nav; `sites.email` = robot-hands@contactforsales.com) and
`leopardessconsulting.co.uk/ai-readiness-quiz.html` (leopardess@contactforsales.com).
Both sites have real configured addresses, which matters for the fix below.

The bug file states the defect; the root cause sits one level deeper, and it is
what makes this a framework fix rather than a component patch. The chassis
already has a chosen, proven mechanism for "deliver a contact form on a platform
with no form backend": `sanitiseFormAction` /
`deliverableFormAction` in
`platform/orchestration/actions/component_library.go` (~lines 1236–1319) rewrite
a known "goes-nowhere" `form_action` value (`""`, `"#"`, `"#contact"`,
`"#contact-form"`, `"/contact"`) into
`mailto:<sites.email>?subject=<domain> enquiry`, refusing when the site has no
real address (including the synthesised `info@<own domain>` display fallback —
see the guard at ~line 1310). The sibling component `contact-form` uses it on 13
live pages via `action="{{.form_action}}" method="POST"` and carries no JS at
all.

`contact-block` never reaches this mechanism because its content-generation
schema never asks the LLM to author a `form_action` field, so
`ctx.ContentData["form_action"]` is not empty but **absent** — and
`sanitiseFormAction`'s presence gate (`raw, present := data["form_action"]; if
!present { return }`, ~line 1259) makes the sanitiser silently decline to act.
The gate itself is right in general — its comment reads "Absent is only a defect
if this component actually has a form to point somewhere; templates without one
must not gain the field" — but as implemented, "this component actually has a
form" is proxied by "the content-generation schema happened to ask for the
field". That is a fragile, invisible precondition: forget it (as contact-block's
schema did) and the safety mechanism never even attempts to help, with no error
anywhere.

## b) The design, and why this shape

### Rejected: a real POST backend (the `audience-check-form` shape)

`audience-check-form` does `fetch(form.getAttribute('action'))` with a real
await, and the bug file's preferred candidate 1 says "the success message must be
downstream of a response, not a timer". Taken literally that argues for a real
endpoint. It is the wrong call here. The chassis is a static-site generator with
no general form backend — `component_library.go`'s own comment on `"/contact"`
reads "no such backend has ever existed in the chassis" — so this path means
building new shared server-side infrastructure: an endpoint, per-site routing of
messages to client inboxes, spam handling, a new deployment surface. That is an
architecture-scope piece of work in its own right, not a bug fix, and the house
style is reuse existing machinery before building new. The owner already made
this exact call on 2026-07-17 (`idea_uk_vm_site/RUNNING_NOTES §Q`): the pattern
for this platform is mailto built from the site's real configured address, and it
is proven on `contact-form`'s 13 pages. A mailto is a client-side navigation,
not an awaitable request, so "success downstream of a response" is adapted to the
closest honest analogue: the status message describes the mail-client handoff at
the moment it is initiated, and never claims delivery.

### Rejected: patching only contact-block's schema, or hardcoding mailto in its template

Adding `form_action` to contact-block's content-generation schema would make the
presence gate pass — for this one component, until the next component forgets the
same way. It also puts an infrastructure decision (where does this form deliver)
into LLM-authored content, which is exactly the failure `sanitiseFormAction`
exists to repair: the LLM demonstrably writes `"#contact"` garbage. Hardcoding
`action="mailto:{{.email}}"` in the template would bypass the sanitiser's
refusal logic (the synthesised-`info@` guard) and duplicate the rule in a second
place. One rule, one place; the value is derived from site config, not authored.

### Chosen: make the presence-gate's own stated condition mechanical

**Framework change**, in `component_library.go`,
`RenderTemplateReportingMissing` (~line 965): immediately after the
`if templateStr == "" { return "", nil, nil }` guard and before
`data := contextToInterfaceMap(ctx)`, add — if
`strings.Contains(templateStr, "form_action")` (the template actually references
the field) and `ctx.ContentData` does not already carry the key, seed
`ctx.ContentData["form_action"] = ""` (initialising the map first if nil), with a
Debug log naming the seeding (the log line doubles as the pod-grep target in §c).
`strings` is already imported.

Why this placement and this value:

- **It implements the comment's rule instead of approximating it.** "This
  component actually has a form to point somewhere" becomes "the template
  references `.form_action`" — true for `contact-form` today, for `contact-block`
  after the data change, and for any future component built the same way, with
  zero dependency on a content schema remembering anything. Templates without the
  reference gain nothing, exactly as the gate's comment demands.
- **`""` is already in `nonDeliveringFormActions`** (~line 1241), so the seeded
  value is the exact input shape the sanitiser already handles for `contact-form`
  pages where the LLM wrote nothing usable. No new vocabulary, no new branch in
  the sanitiser.
- **Seeding `ctx.ContentData` — not the base map — respects the ordering rule
  already written at ~line 1227**: "form_action MUST be sanitised after the
  ContentData merge, not defaulted in the base map above", because a base-map
  default is overwritten by the merge. A seeded ContentData key *participates in*
  the merge, and the sanitiser (called at the end of `contextToInterfaceMap`,
  ~line 1231, and of `contextToMap`, ~line 1152) runs after it. Seeding upstream
  in `RenderTemplateReportingMissing` also covers **both** render paths — the Go
  template path and the regex fallback — with one edit, since both build their
  map from the same context.
- **The mutation cannot leak between components or pages.** Each component render
  gets a freshly built per-render `*RenderContext`
  (`buildRenderContextFromDB(...)` is called per component instance —
  `section_editor_actions.go:776, 867` — before `RenderTemplate`). Worst case on
  an unrelated template that mentions "form_action" in prose is an inert extra
  map key.
- **Consumer impact is a strict improvement, but it IS a behaviour change on the
  sibling**, and per the 2026-07-29 owner ruling (#3) the consumers must be told,
  not merely measured. For a `contact-form` page whose ContentData wholly lacks
  `form_action`, the old pipeline rendered `<no value>`, which the chokepoint
  strips (~lines 1019–1021), shipping `action=""` — a native submit to the page's
  own URL, a 405/404 on static hosting. After this change that case gets the same
  mailto repair the empty-string case already gets. Before submitting to the
  council, **enumerate the consumers with a query, not an assertion**:
  `SELECT function FROM content_components WHERE is_active AND html_template LIKE '%form_action%';`
  and name every hit in the submission. ("No collision is possible" is a query,
  not an argument — the 2026-07-28 ruling.)

This is a `platform/` Go change: council-gate submission (advisory), and inert
until the image is rebuilt, rolled, and pod-verified. Because it widens the
activation condition of a shared mechanism, the concept-register entry for the
form-action repair mechanism is updated (or created — a grep of
`docs026_concept_register/register/` today finds no `form_action` entry) **in the
same commit**, per the ordering-exemption's surviving condition (2).

**Data change**, in the `content_components` row for `function =
'contact-block'` — live config, applied only per the ordering in §c:

1. `html_template`: the form tag gains the sibling's exact convention —
   `<form class="cb-form" id="cb-contact-form" action="{{.form_action}}" method="POST" novalidate aria-label="{{.form_heading}}">`
   — so it is driven by the same, already-proven sanitiser. (`method="POST"` on a
   mailto action is the shape `contact-form` already ships on 13 pages; browsers
   open the mail client regardless.)
2. `js_content`: delete the `setTimeout` fabrication and the unconditional
   `e.preventDefault()`. Keep the validation exactly as-is — it is better than
   `contact-form`'s native-only validation and worth keeping; on invalid input,
   keep `preventDefault()` and the specific inline error, unchanged. On **valid**
   input, do not `preventDefault()`: let the browser's native submission proceed
   to the now-real mailto action, so "sent" can never be claimed without a
   mail-client handoff actually being triggered. Show an honest transitional
   status before falling through — `setStatus('success', 'Opening your email
   client to send this message…')` — never "sent" (nothing client-side can
   confirm delivery), and do **not** `form.reset()` on this path, so the typed
   text survives if the mail client fails to open.

No new deploy plumbing is needed. `RerenderSinglePageAction`'s `collectJSAssets`
(`rerender_single_page_action.go`, ~lines 222–238 and 360–386) reads
`content_components.js_content` for the page's components and writes
`tools/assets/contact-block.js` into the same git commit as the re-rendered HTML
— the same page-rerender handler that `check_contact_form_undeliverable.go`'s
auto-remediation branch already dispatches (reason `section_data_resolved`) for
the analogous `contact-form` case.

## c) Ordered steps, and why the ordering is hard

**The constraint: the Go change must be verified live on the pod before the DB
row changes.** Go changes are inert until rolled; DB config is live immediately.
If the row changes first, the next render of either page under the **old** binary
takes `{{.form_action}}` against a map with no such key: Go's text/template
yields `<no value>`, which `RenderTemplateReportingMissing` strips (~lines
1019–1021), shipping `action=""`. Combined with the new JS — which deliberately
no longer intercepts valid submits — that is a native POST to the page's own URL:
a 405/404, the message lost, on the site's actual contact page. Today's
fabricated success would become a navigation to an error page — a live regression
on client-facing pages. (Correction to this plan's own brief: the regression is
`action=""` after the `<no value>` strip, not a literal `<no value>` in the
attribute — the mechanism was verified against the code today; the constraint is
identical either way.)

**Why the 2026-07-29 owner ruling does not touch this.** That ruling retired
"ordering constraint" as a licence to hold a *commit* back from the shared tree —
because HEAD is shared, nobody can hold one back, and any session's roll ships
it. Nothing here holds a commit. The Go change is committed immediately and is
safe at any point: absent the data change, it is additive and inert (an extra
`""` key that the sanitiser repairs or leaves, on templates that reference the
field). The ordering is between two of **this session's own actions** — commit +
build + pod-verify, then apply the SQL — and the second is an imperative DB write
this session controls, not a commit anyone else can ship early. It is the same
direction as the standing rule "image first, then seeds". If another session
rolls the fleet with the commit first, so much the better; the pod-grep in step 4
is the gate either way, not the roll.

Steps:

1. **Pre-flight.** Grep `LANDMINES.md` for the symbol footprints being touched
   (`content_components`, `RenderTemplateReportingMissing`; the SessionStart hook
   only matches dirty paths). Check the queue for open work on the target
   (`site_work_items` where status not terminal, matching contact-block / the two
   pages). Run the consumer-enumeration query from §b and keep the output for the
   council submission. `\d content_components` before writing any SQL.
2. **Go change + register entry, one commit.** The seeding edit in
   `RenderTemplateReportingMissing` with its Debug log line; the concept-register
   update in the same commit (ordering-exemption condition 2). Commit by explicit
   pathspec. Submit to the council gate before or alongside; if committing before
   the verdict lands, use the `Council-Submitted: <corr>` trailer. The submission
   names `contact-form` (and any other consumer the query finds) and states what
   changed about their guarantee: wholly-absent `form_action` now gets the same
   repair as empty-string.
3. **Build and roll.** Bump `IMAGE_TAG` (makefile ~line 16 — a same-tag rebuild
   ships the node's stale cached binary), `make build-<chassis service>` (builds
   from committed HEAD), push, deploy.
4. **Pod-verify at the artefact.** `kubectl exec` + `strings /app/agent-chassis |
   grep -c` for the new Debug log string, on **every** replica — this is why the
   log line exists. The change removes no string, so pair the positive grep with
   a deliberate-misspelling grep (expect 0) as the control that the grep itself
   can fail. A roll is not evidence the fix shipped; the grep is.
5. **Apply the data change — after step 4 only.** First `SELECT` and save the
   current `html_template` and `js_content` for the row into this workstream's
   directory: the row is live config with no git history, and this saved copy is
   the only rollback. Then one `UPDATE content_components ... WHERE function =
   'contact-block'`, exact SQL recorded in this workstream's RUNBOOK.
6. **Rerender the two live pages** — `robot-hands.com/contact.html` and
   `leopardessconsulting.co.uk/ai-readiness-quiz.html` — via the page-rerender
   handler (the same dispatch the undeliverable check's auto-remediation emits).
   Respect the ~300 s no-dispatch window after any chassis pod restart from step
   3's roll. **Do not rerender `finetuning.uk/case-studies.html`**: its placement
   row is drift (component absent from the served page), and a rerender
   regenerates from DB state — it could *materialise* a contact-block onto a page
   that does not currently serve one, a behaviour change well outside this fix.
   The drift row is noted in the bug file for whoever owns placement-drift
   cleanup.
7. **Verify per §d**, then record the outcome in the bug file. Per the owner's
   2026-08-06 ruling, a finished bug **stays in `bugs_open/`** — update the file
   in place with the fixed-and-live evidence rather than moving it.

## d) Verification

Mirrors the bug file's §"How to verify a fix", adapted where that section assumed
an awaitable request (a mailto is a navigation; there is no 2xx to await).

1. On `https://robot-hands.com/contact.html` (served page, not DB): a valid
   submit triggers a real mail-client handoff — the form's submit is not
   prevented and the browser navigates to the mailto action; the status text says
   the client is being opened and never claims "sent". A forced failure (no mail
   client configured) leaves the typed text in place — no `form.reset()` fired.
   An invalid submit behaves exactly as today: specific inline error, no
   navigation.
2. In the served HTML of both pages, the form carries
   `action="mailto:<the site's configured address>?subject=<domain> enquiry"` —
   the sanitiser's output, confirming the whole chain (seed → merge → sanitise →
   render) ran, not just the template edit.
3. Curl the served `/tools/assets/contact-block.js`: zero matches for
   `setTimeout` and for "has been sent" (the negative greps — the fabrication is
   gone), a match for the new transitional status string (the positive grep, and
   the control that the fetch got the new asset).
4. Re-run the bug file's census query over all form-bearing components and
   confirm `contact-block` now reports `form_has_action = true`. Expect no other
   row to have changed — if one has, another session moved something, look before
   proceeding.
5. Pointer, not a step this session performs: the component's acceptance fence
   (the contract in `doc_plans`, `subject_type='component'`,
   `subject_key='contact-block'`, owned by `staged_component_build`) can now gain
   the check the bug file describes — that the success state requires a handoff —
   without ratifying a defect. See §f.

## e) Risks

- **A site with no real configured email** (or the synthesised
  `info@<own domain>` fallback, which `deliverableFormAction` correctly refuses,
  ~line 1310). There the seeded `""` survives sanitisation and ships
  `action=""`: a native submit to the page's own URL, 405/404, message lost —
  honest failure rather than fabricated success, but still lost. Neither
  affected site is in this state today (both have real `contactforsales.com`
  addresses), and this is precisely the case `check_contact_form_undeliverable.go`
  exists to raise for a human — but that check is scoped to
  `data-component="contact-form"` only, so a **future** contact-block placement
  on an address-less site fails undetected until the check is widened (§f). The
  render chokepoint's own `missingBareFields` Warn on the empty field is the only
  automatic signal in the interim. This residual is accepted, named, and has a
  designated follow-on.
- **The old-binary/new-row window** (§c). Closed by the hard ordering; the
  dangerous state is only reachable if the SQL is applied before the pod-grep
  passes.
- **mailto's UX limits**: a visitor with no configured mail client (webmail-only
  desktop users) gets a failed or blank handoff. The typed text survives by
  design (no reset on the submit path), and the component's `cb-details` block
  already displays the address itself. This trade-off is the owner's chosen
  platform pattern (2026-07-17) and has been live on `contact-form`'s 13 pages
  since; this fix does not widen it.
- **Sibling behaviour change on `contact-form`** — the wholly-absent-field case
  now gets repaired. A strict improvement of the mechanism's own stated intent,
  but it is a change to a shared mechanism's behaviour, so it is measured (the
  consumer query) *and told* (named in the council submission), per the
  2026-07-29 ruling.
- **Future garbage the sanitiser does not recognise**: if a later schema change
  makes the LLM author a novel non-delivering value (not in
  `nonDeliveringFormActions`), it passes through untouched — the same exposure
  `contact-form` has, backstopped there by the discovery check and not yet here.
  Same follow-on as above.
- **No repo-side source for the component row.** `[MEASURED 2026-08-08:
  `grep -rl contact-block scripts/` → no hits]` — the DB row appears to be the
  sole source of the template and JS, so nothing in-repo will re-seed the old
  timer back. The `staged_component_build` lane is told of the row change via the
  bug file regardless, in case their staging keeps a copy this grep cannot see.
- **Concurrent sessions.** who-owns shows no owner; the filing lane's fence work
  is adjacent, not conflicting. The bug file records this plan's existence and
  this directory, so a session picking the bug up finds the work in flight.

## f) Explicitly out of scope — deliberate decisions, not oversights

- **Widening `check_contact_form_undeliverable.go` to also match
  `data-component="contact-block"`.** Legitimate and desirable — it is what closes
  the address-less-site residual in §e — but it only becomes *safe* after this
  fix ships: the check reads `action="..."` out of stored `rendered_html`, and
  its own header comment explains it deliberately refuses to judge any form whose
  submit can be overridden client-side (its first draft wrongly caught 6
  tool-calculator pages exactly this way). That exclusion was true of
  contact-block while its JS unconditionally intercepted submit, and stops being
  true once the fix lands. It is a second, separable change to a shared discovery
  check with its own consumers, and folding it into this fix unannounced is the
  scope failure the guardian seat vetoes. Follow-on, filed against the bug.
- **Extending the component's acceptance fence** so the success state requires a
  handoff ("8/8 mutants caught" today, validation path only — the bug file notes
  the fence deliberately does not assert the success message, so this fix does
  not fight it). That contract belongs to the `staged_component_build` lane; the
  bug file's verification section already names the extension point. Pointer
  left there, nothing touched here.
- **The `finetuning.uk/case-studies.html` drift row** (placement row without
  served markup). Rerendering it as part of this fix could materialise the
  component on a page that does not serve it; the drift class (five rows found by
  the filing lane) needs its own owner and its own sweep, not a side effect here.
