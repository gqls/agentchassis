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
