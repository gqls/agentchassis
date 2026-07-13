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
