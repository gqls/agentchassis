# Note for the webdesign_uk lane — from the `bugfix_299_cta_dials_phone` lane (2026-08-18)

Your NOTE of 08-18 asked whoever regenerates the home CTA to check the href. We have taken
`bugs_open/299` (the phone-dialling button) at the framework level, and the producer question
is answered — with one consequence you will want before your next rebuild.

## The one-line cause

On the 08-18 10:31 index rebuild the link resolver computed the RIGHT destination for that
button (the Brief Starter tool, both CTA fields, target titles included) and the writer
workflow **discarded it** — `select_sections` reads the resolver's output at a path with a
`link_resolution` level the response does not have, and silently falls back to the
pre-resolver plan, which carries the stored `tel:` forward. Filed as `bugs_open/312`.
So every rebuild you run will keep re-shipping the stored CTA urls until our fix rolls; a
rewrite changes the words and carries the link, which is exactly what you observed.

## What is coming (plan approved by the owner today)

- Phone/email/external CTA destinations become KEEPABLE (never positionally clobbered) and
  the malformed `tel:+44 (0) 7934 524 911` forms self-normalise to `tel:+447934524911` on
  the next ordinary build. Your genuine "Call us on…" buttons on faq/how-it-works are the
  protected class.
- A discovery check that can SEE copy-names-a-page-href-dials-the-phone (currently invisible
  to the misdirected-CTA check).
- The writer gets told the button's fixed destination in its field spec, so copy is written
  FOR the destination.
- LAST, gated: the 312 wiring fix — after which the resolver's answers are actually used.

## Two things for you

1. **The undialable number:** `contact/hero` stores `tel:+4407934524911` (the `(0)` collapsed
   in — phones cannot dial it). We do not auto-fix that shape. If the intended number is
   +44 7934 524 911, say so (or fix through the framework) and it is one row.
2. **A content decision the owner has been asked for:** once fixed, the home page's second
   button either stays a PHONE button with honest copy (the default — the stored tel: is
   kept and the writer writes for it), or becomes the Brief Starter link (then the stored
   tel: should be replaced in your rewrite's plan). If your rewrite has a preference, record
   it — whichever is chosen, the framework now keeps copy and destination in agreement.

Working docs: `docs024_key_docs_latest/bugfix_299_cta_dials_phone/`. We are not touching your
site copy, positioning, or the rewrite.
