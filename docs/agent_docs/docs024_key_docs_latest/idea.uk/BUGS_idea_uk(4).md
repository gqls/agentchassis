# idea.uk — known bugs

A running list of bugs to track and fix (mobile and otherwise). Newest first.
Started 2026-06-11.

---

## BUG: mobile — body text too close to the screen edges (margin too small)

- **Status:** fix applied 2026-06-11; awaiting confirmation on-device after rebuild + redeploy.
- **Severity:** low (cosmetic / readability).
- **Where:** idea.uk on a phone. Two places share the cause — the styled wrapper pages
  (thank-you, Terms, Refunds, Privacy, taster error pages — the `a.page()` wrapper in
  `service.go`) and the landing page (`page.html`) on small screens.
- **Symptom:** body text sits hard against the left/right edges of the phone, with too little
  margin.
- **Likely cause:** flat side padding (landing `.wrap` was 22px on mobile; wrapper `main` 24px)
  with no allowance for notches / rounded corners, so on some phones the text crowds the edge.
- **Fix applied:** side padding raised to 24px and made safe-area aware —
  `padding-left/right: max(24px, env(safe-area-inset-left/right))` — in both `page.html` (mobile
  media query) and the `a.page()` wrapper in `service.go`. Both are compiled/embedded, so this
  needs a rebuild + redeploy, then a check on the phone.
- **If still tight afterwards:** note the exact page (URL) and we'll bump further or fix the
  specific element.

---

## 2026-06-11 — report email feedback, logged as bugs (for this and future builds)

Categorised from review of the report email. Most are addressed in this build; all are kept as
standing checks so future report/output builds start better.

### Copy — too technical / LLM-speak / dense / abbreviated
- **BUG:** report wording was jargon-heavy and hard to follow — e.g. "tuned vision reading with
  schema-checked output", and abbreviations like "ML staff", "PII", "T&Cs". *Fixed this build:* a
  global plain-language rule in the prompt (no jargon/acronyms/buzzwords, spell terms out, describe
  concretely); the abbreviations in `riskNote` rewritten in plain words.
- **BUG:** vague references the reader can't picture — "customer's drawings". *Fixed this build:* the
  generate prompt now asks for concrete descriptions with an everyday example (e.g. "the scanned floor
  plans and wiring diagrams your customers email you").
- **BUG:** unclear who did what — "Checks out: Checked that…" (checked by whom?). *Fixed this build:*
  the label is now "What we found", and findings are written as "We searched/checked and found…".
- *For future builds:* default every customer-facing string to plain English written for a
  non-technical owner; treat jargon/acronyms/buzzwords as defects.

### Orientation — recipient doesn't know what the email is
- **BUG:** a busy inbox won't recognise where the email came from or what it's for. *Fixed this build:*
  a plain summary at the top (what idea.uk is, what they asked for, what the report contains) and a
  short footer explaining idea.uk and how to ask questions.
- *For future builds:* any standalone deliverable should open with a one-paragraph plain summary of
  what it is and what it delivers.

### Completeness — rejected sections too terse
- **BUG:** "Didn't make the cut" and "Set aside on risk" showed only a title + scores, so the reader
  didn't know what the concept was. *Fixed this build:* each rejected idea now gets a plain
  description plus a plain reason it was set aside (derived from its scores / risk).
- *For future builds:* when showing rejected options, always say what the thing was and why it was
  rejected, not just a headline.

### Design — should feel like a paid-for, professional document
- **BUG:** the report email reused the landing-page palette and typefaces; it should feel like a
  considered document worth £29, distinct from the marketing page. *Fixed this build:* the HTML email
  was redesigned with its own professional look — deep navy headings, a restrained gold accent, a
  serif for headings over a clean sans body, white "sheet" on a soft grey page, generous spacing.
- *For future builds:* deliverables get a deliberate, professional design separate from marketing
  surfaces; the reader should feel the spend was worth it.

### Still open
- Real-report wording can only be confirmed on a live engine run (the sample uses placeholder text).
  After the next paid/real run, re-read the actual model output and tighten the prompts again if any
  jargon slips through.

- **BUG (copy clarity):** "refunds make customers whole" (a riskNote string, ours not the model) was idiomatic/unclear — reworded to "a mistake would be minor, and a refund would put it right".

### Consistency — off-domain content
- **Observation:** the sample's "set aside on risk" item was a *medical* idea in a report about a
  *legal* firm. Cause: placeholder text in the test stub (mine), not the model. *Fixed:* swapped the
  stub for an on-theme legal example.
- **Prevention (real runs):** "medical/financial advice" appears in the prompts only as scoring
  examples (score step) and one capability-menu example ("medical images"), not as idea seeds, and the
  generate step is anchored on the audience — so off-domain ideas are unlikely. Added a line to the
  generate prompt making this explicit (every candidate must serve THIS audience/domain; cross-sector
  mentions are illustration only). *For future builds:* keep illustrative examples from leaking into
  generated output; check the first real report of any new vertical for off-domain drift.

## 2026-06-11 — OPEN: confirmed order's draft email didn't arrive

- **Status:** OPEN, investigating. Click-through link + page worked end to end up to the "report is
  being generated" page (status: running); the draft never arrived by email.
- **Flow reminder:** Confirm → `fulfil` runs in the background → engine (minutes) → store draft →
  email it to `OPERATOR_EMAIL`. A missing draft means one of: engine still running, engine failed, or
  engine finished but the email send failed.
- **Found while diagnosing:** the engine logged *nothing* and `fulfil` only logged panics, so the run
  was invisible in `journalctl`. *Fix applied:* added progress logging to `fulfil` (engine start,
  engine error, "engine done (N chars)", "draft ready, emailing review to <addr>") and a success log to
  the mailer ("email to <addr> sent" / "... failed: <err>"). A re-test will now show exactly where it
  stops.
- **Leading hypotheses (to confirm, not assume):** (a) engine still running / hung, or the service was
  restarted mid-run (kills the in-flight goroutine); (b) engine finished but the *multipart HTML* draft
  email failed at SES — that path is new this session, whereas the plain request email that *did*
  arrive uses the older text path; (c) engine errored (then a RUN FAILED email should have gone to
  OPERATOR_EMAIL — check spam).
- **Note:** the draft is stored in `orders.json` (`report`/`report_html`) even if the email failed, so
  it isn't lost. The draft goes to OPERATOR_EMAIL (same inbox the request email arrived in), not the
  requester.
- **Next:** check the order's status + stored report; redeploy with the new logging; re-confirm a test
  order while tailing the logs to pinpoint the stop.

## 2026-06-11 — FIXED: corrupted HTML in the report email (line folding)

- **Symptom:** the report email showed a literal `< p style=…>` (a space inside a tag) in one section.
- **Cause:** the HTML report is built as one very long line; when it exceeds the SMTP 998-octet line
  limit a mail server folds it, and the break can land inside a tag. It showed in the review email but
  not the delivered one because the review email's extra lead text shifts where the 998-char boundary
  falls. (Not the model, not the renderer logic — a transport encoding issue.)
- **Fix:** the multipart email parts (text + HTML) are now **base64-encoded and wrapped at 76 chars**
  (`b64Body` in service.go), so no line can approach 998 octets. Verified the wrapping is ≤76 and
  round-trips. The single-part plain emails were already fine (short, newline-broken lines).
- *For future builds:* any HTML email must be transfer-encoded (base64 or quoted-printable), never sent
  as raw long lines.

## 2026-06-11 — DONE: pay-link email rewritten (copy + HTML)

- **Was:** plain text opening "We can do a useful job on this" — too terse, no context for a buyer who
  may not recognise the sender.
- **Now:** a simple HTML email (with plain-text fallback) that opens by saying what it is ("This is
  idea.uk. You asked us to find AI product ideas for <business>…"), a clear "Pay £29 and start my
  report" button, and the delivery + 14-day refund promise. Helpers `payLinkText` / `payLinkHTML`.
