# SUMMARY 2026-08-09 — contact-block transport (bugs_open/228)

Milestone read-out. Current state only — for the chronology and the wrong
turns, see `NOTES_contact_block_transport.md` and `README_where_we_are.md`.

## What we were trying to do

Pick up an unowned bug from `bugs_open/` and fix it properly: not just patch
the one symptom, but find and fix the structural reason it was possible, so
the same defect class can't recur on the next component built the same way.
The bug: `contact-block`, a shared section component live on client contact
pages, validated a visitor's input for real and then faked a "message sent"
confirmation from a `setTimeout` — no request ever left the browser. Every
enquiry through it was silently dropped while the visitor believed it had
gone through.

## Where we came from

The chassis already had a working, owner-chosen mechanism for exactly this
problem class (a static site with no server backend, needing to deliver a
form anyway): `sanitiseFormAction` rewrites a dead form target into a
`mailto:` built from the site's real address, and the sibling `contact-form`
component already used it on 13 pages. `contact-block` never reached that
mechanism because the one function that runs it only acts when the page's
content already carries a `form_action` key — and `contact-block`'s content
was never authored with one, so the repair silently never even tried.

## What we did

Fixed the mechanism, not just the component: `RenderTemplateReportingMissing`
(the one chokepoint every component template render passes through) now
supplies that key whenever a template references it, regardless of whether
content-authoring remembered to. This is the general fix — it covers
`contact-block` and any future component built the same way. It shipped with
tests (mutation-checked by hand: revert the fix, watch the right test fail),
a concept-register entry (`LNK-031`, since the underlying mechanism had never
been documented), and went through four rounds of the platform's council
review — each round a genuine, correct catch (a submission that was a no-op
without its second half; a live-config mutation with no backup/rollback
discipline; a verification script that would itself have produced false
readings) — approved on the fourth, by 15 reviewers.

While that review was in flight, the bug's original filer independently
re-attempted the same fix without checking whether it was already owned,
converged on an identical template edit, hit the exact regression this
lane's ordering discipline existed to prevent, found a sharper workaround
for it, and shipped a more thoroughly tested JavaScript implementation than
the one staged here — one that resolved a browser-behaviour question this
lane had flagged but not settled. That version now ships.

## Where we are now

**Fixed, live, and verified at the artefact**, not just at a status. The
framework fix is pod-verified live fleet-wide. The component fix — html and
JS both — is confirmed live on both real client pages, serving a working
`mailto:` action with an honest status message, from JS whose destination
routing (real endpoint / `mailto:` handoff / honest refusal) has been driven
through a browser five ways and passed. Nothing from this lane's own staged
apply/dispatch scripts was executed — they weren't needed, and running them
would have overwritten already-good work, which is exactly what their own
safety guard correctly refused to do when tried. The bug file, the concept
register, `LANDMINES.md`, and `WRONG_CALLS.md` are all current.

## Where we're going

Nowhere, on this bug — it's closed in substance. Three small, explicitly
named residuals were scoped OUT rather than folded in, each left as a pointer
for whoever picks it up: widening `check_contact_form_undeliverable.go`'s
discovery scope to also cover `contact-block`-shaped components (currently
scoped to `contact-form` only); extending the component's own acceptance
fence to assert the success path now that it's honest; and the broader
architectural point one council reviewer raised — that the generic
`<no value>`-stripped-to-empty-string render behaviour this fix builds on is
the same shape as the platform's worst-documented recurring bug class,
narrowed here for one field rather than closed generally. None of these are
urgent; none block anything live.
