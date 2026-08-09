# 228 — `contact-block` tells the visitor "Your message has been sent" and there is no transport in the component at all

**Filed** 2026-08-08 by lane `staged_component_build`, found while writing this
component's acceptance contract under D10 (production-line batch 6, the interactive
stock). Class: **a component that reports success for work it never did** — the same
family as `bugs_open/012` (a status of `complete` over a truncated artefact) and
`a-complete-work-item-is-not-a-repaired-artefact`, but pointed at a *visitor* rather
than at us.

**Severity is about trust, not volume.** Two live pages on two client sites invite a
visitor to type their name, email and message, and then tell them it was sent. Nothing
is sent, nowhere, ever. The enquiry is dropped and the visitor believes it was received,
so they do not follow up — which is the worst possible failure mode for a contact form,
strictly worse than an error message.

## Symptom

On any page carrying `contact-block`, fill the form validly and press the submit
button. After ~1.2s the button re-enables and a green status message reads:

> Your message has been sent. We'll be in touch shortly.

No request leaves the browser.

## Root cause `[VERIFIED — read in the live DB row AND in the SERVED asset]`

Three independent facts, each checked at the artefact rather than inferred:

1. **The served form has no destination.** `curl https://robot-hands.com/contact.html`:
   ```html
   <form class="cb-form" id="cb-contact-form" novalidate aria-label="Submit a Technical Enquiry">
   ```
   No `action`, no `method`. (Its sibling `contact-form` has
   `action="{{.form_action}}" method="POST"`, so the absence here is not a
   platform-wide convention.)

2. **The served script has no transport.** `curl
   https://robot-hands.com/tools/assets/contact-block.js` → 2,100 bytes, and
   `grep -cE 'fetch\(|XMLHttpRequest|sendBeacon|form\.submit\(|action *='` over it
   returns **0**. Same result against the live `content_components.js_content` row, so
   the deployed asset and the source agree — this is not a build-stripping artefact.

3. **The success message is a timer.** The submit handler calls
   `e.preventDefault()`, runs four client-side validations, and then:
   ```js
   setTimeout(function() {
     submitBtn.disabled = false;
     if (btnText) btnText.textContent = originalText;
     setStatus('success', 'Your message has been sent. We\'ll be in touch shortly.');
     form.reset();
   }, 1200);
   ```
   The 1,200 ms delay exists only to *look* like a network round-trip. `form.reset()`
   then destroys the visitor's typed text, so they cannot even retry from the page.

**The validation is real and the send is not**, which is precisely why this looks
healthy from the outside: a visitor who mistypes their email gets a correct, specific
error, so the form appears to be wired up.

## Who is affected — measured, not estimated

```sql
-- every ACTIVE component whose template carries a <form>, asked three questions:
-- does the form name a destination, does anything in it transport, does it claim
-- the message was sent?
SELECT cc.function,
       (cc.html_template ~* '<form[^>]*action=')                       AS form_has_action,
       (coalesce(cc.js_content,'') ~ 'fetch\(|XMLHttpRequest|sendBeacon') AS js_has_transport,
       (cc.html_template ~ 'fetch\(|XMLHttpRequest|sendBeacon')         AS tpl_has_transport,
       count(DISTINCT pc.page_id) AS pages
FROM content_components cc
LEFT JOIN page_components pc ON pc.component_id = cc.id
LEFT JOIN pages p ON p.id = pc.page_id AND p.status = 'active'
WHERE cc.is_active AND cc.html_template ~* '<form'
GROUP BY 1,2,3,4 ORDER BY 5 DESC;
```

30 rows, 2026-08-08. **`contact-block` is the only one that has no `action`, no
transport of any kind, AND tells the visitor the message was sent.** Every other
form-bearing component either names an action (`contact-form` 13 pages,
`intent-probe`, `report-request-form`, the two header search forms) or fetches
(`audience-check-form`, which posts to `form.getAttribute('action')`), or is a
calculator whose `<form>` is a layout wrapper that never claims to send anything.

The live pages:

| site | page | placement row | component in the SERVED html |
|---|---|---|---|
| robot-hands.com | `/contact.html` | active | **yes** — form live |
| leopardessconsulting.co.uk | `/ai-readiness-quiz.html` | active | **yes** — form live |
| ~~finetuning.uk~~ | ~~`/case-studies.html`~~ | active | **no** — see correction |

`robot-hands.com/contact.html` is the site's **contact page** — the one a buyer
reaches from the nav.

> **CORRECTED 2026-08-08, same day, before anyone acted on it: this bug first said
> "three live pages". It is TWO.** The third placement row,
> `finetuning.uk/case-studies.html`, is **drift**: the row exists, the page is active,
> and the served HTML contains no `contact-block` markup at all (`grep -c
> 'data-component="contact-block"'` → 0; that page serves `hero-case-studies`,
> `case-studies-list`, `testimonials`). **What caught it:** a later sweep in the same
> session probed all 38 active placements of JS-bearing section components against
> their SERVED pages, and this row came back component-absent. The original figure came
> from `page_components` alone — a placement row is a claim about a page, not a
> measurement of one, and this lane has now found five such rows. The defect itself is
> unchanged and still live on the two pages above, one of them a contact page.

## Fix candidates, ordered by what closes the door

1. **Give the form a real destination and delete the timer.** Reuse the sibling's
   mechanism rather than invent one: `contact-form` already carries
   `action="{{.form_action}}" method="POST"` and is placed on 13 pages, and
   `audience-check-form` already does the `fetch(form.getAttribute('action'))` version
   with a real await. Make the success message conditional on the response, and keep the
   visitor's text on failure instead of `form.reset()`-ing it away.
   **Prefer this.** It is the only candidate where the success message can no longer be
   printed without a send having happened — the bad state becomes unrepresentable rather
   than merely unlikely.
2. **Make the absence honest**: no `setStatus('success', …)` at all — the button becomes
   an ordinary `mailto:` link, or the block renders its `cb-details` (email, phone,
   location, all already in the template and already correct) without the form. Smaller
   and safe, but it removes a capability rather than fixing it.
3. **Leave the timer and add a "demo form" note.** Rejected: a disclaimer next to a
   green "your message has been sent" is not read, and the component would still be
   collecting personal data it immediately discards.

Whichever is taken, **the success message must be downstream of a response, not of a
timer.** A component that can print "sent" with the network unplugged will drift back
to this state the first time an endpoint changes.

## How to verify a fix

1. On `https://robot-hands.com/contact.html`, submit a valid message with DevTools'
   Network panel open (or drive it headless) and confirm **a request is made** and that
   the success message appears only after it resolves 2xx. Confirm a forced failure
   (offline, or a 500) shows an error and **leaves the typed text in place**.
2. Re-run the census query above and confirm `contact-block` now reports
   `form_has_action` or `js_has_transport` true.
3. Re-run this component's acceptance fence — the contract written today
   (`doc_plans`, `subject_type='component'`, `subject_key='contact-block'`) asserts the
   VALIDATION path only and **deliberately does not assert the success message**, so a
   correct fix does not have to fight it. If the fix lands, the fence should gain a
   check that the success state requires a response; that is the point at which
   asserting it stops ratifying the defect.

## A note on why this bug file exists at all, rather than a 090 run

CLAUDE.md's owner ruling (2026-07-31) requires a `bugs_open/` file asserting a
cross-cutting or structural root cause to go through the diagnosis loop, or to say
plainly why first-hand verification substituted. **Substituting, and here is the why:**
the claimed cause is not cross-cutting and is not remote from the symptom — it is the
absence of a call inside the one component's own 2,100-byte script, read in the deployed
artefact and in the DB row, with the blast radius measured by a census over every
form-bearing component rather than argued. There is no candidate mechanism elsewhere for
the loop to find. What the loop *could* still add is a judgement on candidate 1's
destination, which is a design question, not a diagnosis.

## Adjacent observation, NOT part of this bug's claim

`contact-form` (13 pages) resolves `{{.form_action}}` to a **`mailto:` URL with
`method="POST"`** — e.g. on `leopardessconsulting.co.uk/contact.html`:
`action="mailto:leopardess@contactforsales.com?subject=leopardessconsulting.co.uk enquiry"
method="POST"`. `[UNVERIFIED]` what a current browser actually does with that: a POST to
a `mailto:` target is not something Chrome or Safari reliably deliver, and it may open
an empty compose window or nothing at all. That would be a *different* defect on 13
pages and it deserves its own check before anyone asserts it — drive one of those forms
in a real browser and watch the result. It is recorded here only so the next reader does
not assume the sibling component is known-good simply because it names an action.

## Cross-links

- `bugs_open/012` — the same family: a status reported over work that did not happen.
- `docs/agent_docs/docs024_key_docs_latest/staged_component_build/` — the lane, batch 6;
  the component's fence and mutants are `scripts/fence_component_contact_block.json`
  and `scripts/mutants_component_contact_block.json` (8/8 mutants caught).
- `LANDMINES.md` — entry added 2026-08-08, footprinted on `contact-block`.
