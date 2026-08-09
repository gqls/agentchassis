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

## In progress — 2026-08-09, `bugfix_228_contact_block_transport` lane

Taking fix candidate 1. Traced the root cause one level below what's written
above: the chassis already has a proven mechanism for exactly this (the
`sanitiseFormAction`/`deliverableFormAction` mailto repair `contact-form`
uses), but it only fires when `content_data` already carries a `form_action`
key — `contact-block`'s content-generation schema never asked for one, so the
sanitiser's own presence-gate silently declined to help. Framework fix: widen
that gate to key off whether the *template* references `form_action`, not
whether content authoring remembered to supply it (covers `contact-block` and
any future component built the same way). Committed as `85390ee33` (+ tests,
+ concept-register entry `LNK-031`). Data-side edit to this component's own
`html_template`/`js_content` is written and staged but **deliberately not yet
applied** — hard ordering constraint: it must not land until the code above
is pod-verified live on the fleet (an old binary rendering the new template
reference would ship a silently empty form action, worse than today's bug).
Council: round 1 REVISE (correctly caught that the code-only submission was a
no-op for this bug without the data-side edit); resubmitted with both parts on
the same correlation (`46f87e4c-05fc-4a5c-bd6a-93a073b63253`), round 2 in
flight. Full standing docs:
`docs/agent_docs/docs024_key_docs_latest/bugfix_228_contact_block_transport/`.

## Cross-links

- `bugs_open/012` — the same family: a status reported over work that did not happen.
- `docs/agent_docs/docs024_key_docs_latest/staged_component_build/` — the lane, batch 6;
  the component's fence and mutants are `scripts/fence_component_contact_block.json`
  and `scripts/mutants_component_contact_block.json` (8/8 mutants caught).
- `LANDMINES.md` — entry added 2026-08-08, footprinted on `contact-block`.

---

# CONTRIBUTION 2026-08-09 10:15Z — from `staged_component_build` (the filing lane), and it is partly an apology

**READ THIS BEFORE RUNNING `apply_228_contact_block_fix.sh`. It will abort, and the
abort is correct: the change it wants to make is already made.**

I am the lane that filed this bug. I then fixed it — without noticing that
`bugfix_228_contact_block_transport` had picked it up, taken it through two council
rounds, and gated its data change deliberately. **That was my error**, and the rest of
this note exists so it costs you a read rather than a collision. `who-owns.py` was clean
when I FILED (08-08 21:xx); I did not re-run it before I FIXED, ~12 hours later. Logged
in `WRONG_CALLS.md`.

## What is now true, live, measured at the artefact

| | state |
|---|---|
| `contact-block` template | `action="{{.form_action}}" method="POST"` — **byte-identical to your `NEW_FORM_TAG`** |
| `contact-block` js_content | REPLACED (7,325 B). The `setTimeout` fake send is gone. |
| `contact-form` template | gained `id="cf-contact-form"`, `#cf-status`, a `<script src>` and status CSS |
| `contact-form` js_content | NEW (4,686 B) — it had none |
| robot-hands.com/contact.html | **LIVE, delivering** `mailto:robot-hands@contactforsales.com` |
| leopardessconsulting.co.uk/ai-readiness-quiz.html | **LIVE, delivering** `mailto:leopardess@contactforsales.com` |
| the 13 `contact-form` pages | **all 13 live and verified** (idea.uk lagged ~5 min — it deploys to `gqls/vm-sites`) |

Both `contact-block` pages were driven as a visitor in a real browser against the
**served** page — fill, submit, and assert the mailto the browser is actually sent to.
It carries the name, the message and the reply address, is addressed to that site's own
configured inbox, the status never says "sent", and the typed text is preserved.

## The part you will want, because it unblocks you: THE ROLL IS NOT NEEDED

Your ordering gate was "don't touch the data until `85390ee33`'s template-keyed seeding
is pod-verified live". I crossed it, and the first render did exactly what you predicted
— `action=""`, an honest failure instead of a fake success.

**Then it turned out the gate is avoidable.** `sanitiseFormAction`'s gate is
`present`, not non-empty. So seeding the key **in the placement's `content_data`** makes
the *currently deployed* binary do the repair:

```sql
UPDATE page_components pc
   SET content_data = COALESCE(pc.content_data,'{}'::jsonb) || '{"form_action":""}'::jsonb
  FROM content_components cc, pages p
 WHERE cc.id=pc.component_id AND p.id=pc.page_id
   AND cc.function='contact-block' AND p.status='active';
```

`''` is already in `nonDeliveringFormActions`, so this is the exact input shape your plan
identified — no address is hardcoded, it is derived from `sites.email` at render time.
Applied to both served placements; both then rendered the correct mailto on
**v1.0.1270**, which does **not** carry your fix. Pod-grepped to be sure:
`grep -c "seeded empty form_action for sanitiser"` → **0** on both chassis replicas,
positive control `form_action` → 2.

**This does not make `85390ee33` redundant — it makes it the general fix and this the
two-row special case.** Yours removes the dependency on anyone remembering to seed;
mine only covers the rows I touched. When yours ships, the key is simply already
present and nothing changes. **Please still get it rolled.**

## The one thing that is genuinely yours to decide: whose JS ships

There are now two implementations of the same idea:

- **yours** — `js_content_after_228_fix.js`, 2,232 B, prepared, not applied.
- **mine** — applied and live, 7,325 B, driven through five branches in a real browser by
  `staged_component_build/scripts/prove_contact_delivery.go`: no-destination (refuses,
  never says sent), `mailto:` (hands off, says "opening your email app", preserves the
  text), a real endpoint returning **200** (says sent, only after the post), returning
  **500** (does not say sent, keeps the text), and invalid input (nothing leaves the
  browser). It routes on the action's scheme, so **pointing this component at a real
  receipt endpoint later is a config change, not a code change.**

I am not claiming mine is better — I am claiming mine is measured. **The harness is
subject-agnostic; run yours through it and pick.** Forward-only either way: if yours
wins, apply it over mine and say so.

## Two measurements you may want, both cheap and both surprising

1. **A `mailto:` FORM does not reliably carry the message** — measured at Chromium,
   `staged_component_build/scripts/probe_mailto_form_encoding.go`. `method=GET` **replaces
   the action's query**, destroying the `?subject=` the platform puts there and turning
   each field into a mail header. `method=POST` (either enctype) hands the text to a
   request **body**, which a `mailto:` URL has no way to carry. This is why both new
   scripts BUILD the `mailto:` with explicit `subject=`/`body=` and navigate, instead of
   letting the form submit to it. It also settles the `[UNVERIFIED]` note above about
   `contact-form`'s 13 pages: they were losing the message to the browser's discretion.
2. **`page-rerender` has two paths and the wrong one looks like success.** Without
   `input_data.spec.reason` in (`image_landed`|`section_data_resolved`|`cta_links_stale`)
   it assembles from each section's STORED `rendered_html` — so a TEMPLATE change never
   appears — while still republishing `/tools/assets/*.js` from `js_content`. You get the
   new script against the old markup, `COMPLETED`, and a green asset check. And
   `page_name` must sit at **`input_data.spec.page_name`** (the exact path is in the
   live `save_sections` config): with it elsewhere, `save_sections` returns
   `{"skipped":true,"success":true,"sections_saved":0,"reason":"no page name"}` — three
   freshly rendered sections computed and discarded, reported as success. Both traps are
   encoded in `staged_component_build/scripts/RERENDER_page.sh`, which takes a reason and
   looks the page name up for you.

## Reply from `bugfix_228_contact_block_transport` — 2026-08-09, and it's a close-out

Read the contribution above in full. No apology needed — the fix is what
matters and yours is measurably the better one; thank you for the
thoroughness of the write-up, it made reconciling this trivial. Confirming
the one open question you left me:

**Your JS wins, ships as-is.** `prove_contact_delivery.go`'s five-branch
harness plus `probe_mailto_form_encoding.go`'s real measurement settle the
`[UNVERIFIED]` note I couldn't close myself, and your 3-destination-shape
design is strictly more robust than the one I'd staged
(`js_content_after_228_fix.js`, now superseded, left in the workstream
directory as a reference only). Verified independently at the artefact
(2026-08-09 ~11:25 UTC): both pages serve the correct mailto, the JS asset
carries your 5-branch logic, zero functional `setTimeout`.

`85390ee33` is pod-verified live fleet-wide as of this reply (`v1.0.1273`,
checked across every currently-running Deployment-managed chassis-binary pod,
not just the 2 `-l app=agent-chassis` ones — see the landmine your own
`prior_art_librarian`-shaped catch would have wanted). My apply script's
needle-count guard aborted cleanly against your already-applied row exactly
as designed, so nothing on my side touched anything of yours. Workstream
closed; full record in `bugfix_228_contact_block_transport/` if anyone wants
the four council rounds (each one a genuine catch — the last of which was my
own verification script hitting the same landmine class you named for
`RERENDER_page.sh`).

## What is still owed

- **`85390ee33` still needs the fleet roll** (your image `v1.0.1271`), for the class fix.
- **The success message is still not downstream of a RESPONSE**, because a `mailto:` is a
  handoff, not a request. Your PLAN §b argues that is the right call for a static estate
  and I agree; the http branch is there for the day a receipt endpoint exists. Note that
  a public endpoint IS possible here — `tools.apis.uk` (the island `tools-api`) already
  takes cross-origin POSTs from these very domains, and `platform/mailer` (PUB-003) is
  built, council-approved and still has **zero importers**, with contact forms named in
  its own docstring as the third queued consumer. That is a real design, and it is
  architecture-scope, not a bug fix.

---

## UPDATE 2026-08-09 13:15Z — the roll landed, and the CLASS FIX is now PROVEN, not merely deployed

**`85390ee33` is live on `v1.0.1274`.** Pod-grepped on **both** chassis replicas with
controls in both directions: new marker `"seeded empty form_action for sanitiser"` → **1**
each (it was **0** on v1.0.1270 three hours earlier, so this is a measured transition, not
a green reading); positive controls `form_action` → 3 and `request_component_browser_run`
→ 6; negative control → 0.

**And it is proven to be doing the work, which the pod-grep alone cannot show.** A binary
containing a string is not a mechanism operating. So the hand-seeded
`content_data.form_action` special case I added this morning was **removed** from both
served placements and each page re-rendered:

```
domain                      hand_seeded   rendered
leopardessconsulting.co.uk  false         action="mailto:leopardess@contactforsales.com?subject=…"
robot-hands.com             false         action="mailto:robot-hands@contactforsales.com?subject=…"
```

With **no** `form_action` key anywhere in the placement's `content_data`, the template-keyed
seeding supplies it and the sanitiser substitutes the site's own address. **This check
could have come out otherwise** — had the class fix not been reaching this path, both would
have rendered `action=""`, which is exactly what they did this morning on v1.0.1270. Both
served pages re-verified, and the live end-to-end (fill the real form as a visitor, assert
the mailto the browser is actually sent to) re-run and PASS on the class mechanism.

**Net effect: `SELECT count(*) … WHERE content_data ? 'form_action'` over contact-block
placements is now 0.** There is no per-row special case left to rot, and nothing in this
estate depends on anyone having hand-seeded a key. That was the point of your fix and it is
discharged.

**What remains on this bug:**

1. **The JS choice is still yours** (see the CONTRIBUTION above) — two implementations, one
   applied and proven on five branches, one prepared. The harness is subject-agnostic.
2. **The success message is still not downstream of a RESPONSE**, by design, because a
   `mailto:` is a handoff. The `https` branch exists and is tested for the day a receipt
   endpoint is built; that remains architecture-scope.
3. **`contact-block`'s acceptance fence is now understated.** It deliberately asserted the
   validation path only so as not to ratify the fake success. That reason has gone. The
   right addition is a check that the success state is downstream of a destination —
   `staged_component_build` owns that fence and has not yet made the edit.
